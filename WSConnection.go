package go_push

import (
	"github.com/gorilla/websocket"
	"sync"
)

type WSConnection struct {
	mutex sync.Mutex
	connId uint64
	wsSocket *websocket.Conn
	inChan chan*WSMessage
	outChan chan*WSMessage
	closeChan chan byte
	isClosed bool
}

// 读websocket
func (wsConnection *WSConnection) readLoop() {
	var (
		msgType int
		msgData []byte
		message *WSMessage
		err error
	)
	for {
		if msgType, msgData, err = wsConnection.wsSocket.ReadMessage(); err != nil {
			goto ERR
		}

		message = BuildWSMessage(msgType, msgData)

		select {
		case wsConnection.inChan <- message:
		case <- wsConnection.closeChan:
			goto CLOSED
		}
	}

ERR:
	wsConnection.Close()
CLOSED:
}

// 写websocket
func (wsConnection *WSConnection) writeLoop() {
	var (
		message *WSMessage
		err error
	)
	for {
		select {
		case message = <- wsConnection.outChan:
			if err = wsConnection.wsSocket.WriteMessage(message.msgType, message.msgData); err != nil {
				goto ERR
			}
		case <- wsConnection.closeChan:
			goto CLOSED
		}
	}
ERR:
	wsConnection.Close()
CLOSED:
}

/**
		以下是API
 */

func InitWSConnection(connId uint64, wsSocket *websocket.Conn) (wsConnection *WSConnection) {
	wsConnection = &WSConnection{
		wsSocket: wsSocket,
		connId: connId,
		inChan: make(chan *WSMessage, 1000),
		outChan: make(chan *WSMessage, 1000),
		closeChan: make(chan byte),
	}

	go wsConnection.readLoop()
	go wsConnection.writeLoop()

	return
}

// 发送消息
func (wsConnection *WSConnection) SendMessage(message *WSMessage) (err error) {
	select {
	case wsConnection.outChan <- message:
	case <- wsConnection.closeChan:
		err = ERR_CONNECTION_LOSS
	default:	// 写操作不会阻塞, 因为channel已经预留给websocket一定的缓冲空间
		err = ERR_SEND_MESSAGE_FULL
	}
	return
}

// 读取消息
func (wsConnection *WSConnection) ReadMessage() (message *WSMessage, err error) {
	select {
	case message = <- wsConnection.inChan:
	case <- wsConnection.closeChan:
		err = ERR_CONNECTION_LOSS
	}
	return
}

// 关闭连接
func (wsConnection *WSConnection) Close() {
	wsConnection.wsSocket.Close()

	wsConnection.mutex.Lock()
	defer wsConnection.mutex.Unlock()

	if !wsConnection.isClosed {
		wsConnection.isClosed = true
		close(wsConnection.closeChan)
	}
}