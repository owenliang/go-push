package logic

import (
	"net/http"
	"time"
	"net"
	"strconv"
	"encoding/json"
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
		err error
		items string
		msgArr []json.RawMessage
	)
	if err = req.ParseForm(); err != nil {
		return
	}

	items = req.PostForm.Get("items")
	if err = json.Unmarshal([]byte(items), &msgArr); err != nil {
		return
	}

	G_gateConnMgr.PushAll(msgArr)
}

// 房间推送POST room=xxx&msg
func handlePushRoom(resp http.ResponseWriter, req *http.Request) {
	var (
		err error
		room string
		items string
		msgArr []json.RawMessage
	)
	if err = req.ParseForm(); err != nil {
		return
	}

	room = req.PostForm.Get("room")
	items = req.PostForm.Get("items")

	if err = json.Unmarshal([]byte(items), &msgArr); err != nil {
		return
	}

	G_gateConnMgr.PushRoom(room, msgArr)
}

// 处理统计
func handleStats(resp http.ResponseWriter, req *http.Request) {
	var (
		data []byte
		err error
	)

	if data, err = G_stats.Dump(); err != nil {
		return
	}

	resp.Write(data)
}

func InitService() (err error) {
	var (
		mux *http.ServeMux
		server *http.Server
		listener net.Listener
	)

	// 路由
	mux = http.NewServeMux()
	mux.HandleFunc("/push/all", handlePushAll)
	mux.HandleFunc("/push/room", handlePushRoom)
	mux.HandleFunc("/stats", handleStats)

	// HTTP/1服务
	server = &http.Server{
		ReadTimeout: time.Duration(G_config.ServiceReadTimeout) * time.Millisecond,
		WriteTimeout: time.Duration(G_config.ServiceWriteTimeout) * time.Millisecond,
		Handler: mux,
	}

	// 监听端口
	if listener, err = net.Listen("tcp", ":" + strconv.Itoa(G_config.ServicePort)); err != nil {
		return
	}

	// 赋值全局变量
	G_service = &Service{
		server: server,
	}

	// 拉起服务
	go server.Serve(listener)

	return
}