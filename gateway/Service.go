package gateway

import (
	"crypto/tls"
	"encoding/json"
	"github.com/owenliang/go-push/common"
	"net"
	"net/http"
	"strconv"
	"time"
)

type Service struct {
	server *http.Server
}

var (
	G_service *Service
)

// 全量推送POST msg={}
func handlePushAll(resp http.ResponseWriter, req *http.Request) {
	var (
		err    error
		items  string
		msgArr []json.RawMessage
		msgIdx int
	)
	if err = req.ParseForm(); err != nil {
		return
	}

	items = req.PostForm.Get("items")
	if err = json.Unmarshal([]byte(items), &msgArr); err != nil {
		return
	}

	for msgIdx, _ = range msgArr {
		G_merger.PushAll(&msgArr[msgIdx])
	}
}

// 房间推送POST room=xxx&msg
func handlePushRoom(resp http.ResponseWriter, req *http.Request) {
	var (
		err    error
		room   string
		items  string
		msgArr []json.RawMessage
		msgIdx int
	)
	if err = req.ParseForm(); err != nil {
		return
	}

	room = req.PostForm.Get("room")
	items = req.PostForm.Get("items")

	if err = json.Unmarshal([]byte(items), &msgArr); err != nil {
		return
	}

	for msgIdx, _ = range msgArr {
		G_merger.PushRoom(room, &msgArr[msgIdx])
	}
}

// 用户推送POST conn=xxx&msg
func handlePushConn(resp http.ResponseWriter, req *http.Request) {
	var (
		err    error
		conn   string
		connId uint64
		items  string
		msgArr []json.RawMessage
		wsMsg  *common.WSMessage
	)
	if err = req.ParseForm(); err != nil {
		return
	}

	conn = req.PostForm.Get("conn")
	items = req.PostForm.Get("items")

	connId, err = strconv.ParseUint(conn, 10, 64)
	if err != nil {
		return
	}

	if err = json.Unmarshal([]byte(items), &msgArr); err != nil {
		return
	}
	// 构造要推送的消息体
	wsMsg, err = common.EncodeWSMessage(&common.BizMessage{
		Type: "PUSH",
		Data: json.RawMessage(items),
	})
	if err != nil {
		return
	}

	// 1. 根据ConnId找出桶，向桶里的Conn 发送消息
	bucket := G_connMgr.GetBucketByConnId(connId)
	go bucket.PushConn(connId, wsMsg)

}

// 统计
func handleStats(resp http.ResponseWriter, req *http.Request) {
	var (
		data []byte
		err  error
	)

	if data, err = G_stats.Dump(); err != nil {
		return
	}

	resp.Write(data)
}

func InitService() (err error) {
	var (
		mux      *http.ServeMux
		server   *http.Server
		listener net.Listener
	)

	// 路由
	mux = http.NewServeMux()
	mux.HandleFunc("/push/all", handlePushAll)
	mux.HandleFunc("/push/room", handlePushRoom)
	mux.HandleFunc("/push/conn", handlePushConn)
	mux.HandleFunc("/stats", handleStats)

	// TLS证书解析验证
	if _, err = tls.LoadX509KeyPair(G_config.ServerPem, G_config.ServerKey); err != nil {
		return common.ERR_CERT_INVALID
	}

	// HTTP/2 TLS服务
	server = &http.Server{
		ReadTimeout:  time.Duration(G_config.ServiceReadTimeout) * time.Millisecond,
		WriteTimeout: time.Duration(G_config.ServiceWriteTimeout) * time.Millisecond,
		Handler:      mux,
	}

	// 监听端口
	if listener, err = net.Listen("tcp", ":"+strconv.Itoa(G_config.ServicePort)); err != nil {
		return
	}

	// 赋值全局变量
	G_service = &Service{
		server: server,
	}

	// 拉起服务
	go server.ServeTLS(listener, G_config.ServerPem, G_config.ServerKey)

	return
}
