package go_push

import "fmt"

// 处理websocket请求
func WSHandle(wsConnection *WSConnection) {
	var (
		message *WSMessage
		err error
	)

	for {
		if message, err = wsConnection.ReadMessage(); err != nil {
			goto ERR
		}

		// TODO: 处理message
		message = message

		fmt.Println("消息", *message)
	}
ERR:
	return
}
