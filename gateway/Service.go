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

// 全量推送
func handlePushAll(resp http.ResponseWriter, req *http.Request) {
	var msg = json.RawMessage(`{"test": "这是一条测试消息"}`)
	G_connMgr.PushAll(&msg)
}

// 房间推送
func handlePushRoom(resp http.ResponseWriter, req *http.Request) {
	var msg = json.RawMessage(`{"test": "这是一条测试消息"}`)
	G_connMgr.PushRoom("默认房间", &msg)
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