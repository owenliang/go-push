package gateway

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
		msg string
		pushMsg json.RawMessage
	)
	if err = req.ParseForm(); err != nil {
		return
	}

	msg = req.PostForm.Get("msg")

	pushMsg = json.RawMessage(msg)
	pushMsg = pushMsg
}

// 房间推送POST room=xxx&msg
func handlePushRoom(resp http.ResponseWriter, req *http.Request) {
	var (
		err error
		room string
		msg string
		pushMsg json.RawMessage
	)
	if err = req.ParseForm(); err != nil {
		return
	}

	room = req.PostForm.Get("room")
	msg = req.PostForm.Get("msg")

	pushMsg = json.RawMessage(msg)

	room = room
	pushMsg = pushMsg
}

// 统计
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

	// HTTP服务
	server = &http.Server{
		ReadTimeout: time.Duration(2000) * time.Millisecond,
		WriteTimeout: time.Duration(2000) * time.Millisecond,
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