package go_push

import (
	"fmt"
	"time"
)

// 每隔1秒, 检查一次连接是否健康
func (wsConnection *WSConnection) heartbeatChecker() {
	for {
		if !wsConnection.IsAlive() {
			break
		}
		time.Sleep(1 * time.Second)
	}

	// 确保连接被关闭
	wsConnection.Close()
}

// 处理websocket请求
func (wsConnection *WSConnection) WSHandle() {
	var (
		message *WSMessage
		err error
	)

	go wsConnection.heartbeatChecker()

	// 请求处理协程
	for {
		if message, err = wsConnection.ReadMessage(); err != nil {
			goto ERR
		}

		// TODO: 处理message
		message = message

		fmt.Println("消息", *message)

		// 客户端提交如下3种请求:
		// 1,收到PING则响应PONG
		// 2,收到JOIN则加入ROOM
		// 3,收到LEAVE则离开ROOM
	}
ERR:
	return
}
