package gateway

import (
	"github.com/gorilla/websocket"
	"sync"
	"time"
	"encoding/json"
)

type WSConnection struct {
	mutex sync.Mutex
	connId uint64
	wsSocket *websocket.Conn
	inChan chan*WSMessage
	outChan chan*WSMessage
	closeChan chan byte
	isClosed bool

	lastHeartbeatTime time.Time // 最近一次心跳时间

	lastCommit time.Time // 上次提交batch时间
	pushBatch []*json.RawMessage	// 推送批次
	resetNotify chan byte // 提交通知定时器重置
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
		inChan: make(chan *WSMessage, G_config.WsInChannelSize),
		outChan: make(chan *WSMessage, G_config.WsOutChannelSize),
		closeChan: make(chan byte),
		lastHeartbeatTime: time.Now(),
		lastCommit: time.Now(),
		pushBatch: make([]*json.RawMessage, 0),
		resetNotify: make(chan byte, 1),
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

// 检查心跳（不需要太频繁）
func (wsConnection *WSConnection) IsAlive() bool {
	var (
		now = time.Now()
	)

	wsConnection.mutex.Lock()
	defer wsConnection.mutex.Unlock()

	// 连接已关闭 或者 太久没有心跳
	if wsConnection.isClosed || now.Sub(wsConnection.lastHeartbeatTime) > time.Duration(G_config.WsHeartbeatInterval) * time.Second {
		return false
	}
	return true
}

// 更新心跳
func (WSConnection *WSConnection) KeepAlive() {
	var (
		now = time.Now()
	)

	WSConnection.mutex.Lock()
	defer WSConnection.mutex.Unlock()

	WSConnection.lastHeartbeatTime = now
}