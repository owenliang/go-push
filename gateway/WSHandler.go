package gateway

import (
	"time"
	"github.com/gorilla/websocket"
	"encoding/json"
	"github.com/owenliang/go-push/common"
)

// 每隔1秒, 检查一次连接是否健康
func (wsConnection *WSConnection) heartbeatChecker() {
	var (
		timer *time.Timer
	)
	timer = time.NewTimer(time.Duration(G_config.WsHeartbeatInterval) * time.Second)
	for {
		select {
		case <- timer.C:
			if !wsConnection.IsAlive() {
				wsConnection.Close()
				goto EXIT
			}
			timer.Reset(time.Duration(G_config.WsHeartbeatInterval) * time.Second)
		case <- wsConnection.closeChan:
			timer.Stop()
			goto EXIT
		}
	}

EXIT:
	// 确保连接被关闭
}

// 处理PING请求
func (wsConnection *WSConnection) handlePing(bizReq *common.BizMessage) (bizResp *common.BizMessage, err error) {
	var (
		buf []byte
	)

	wsConnection.KeepAlive()

	if buf, err = json.Marshal(common.BizPongData{}); err != nil {
		return
	}
	bizResp = &common.BizMessage{
		Type: "PONG",
		Data: json.RawMessage(buf),
	}
	return
}

// 处理JOIN请求
func (wsConnection *WSConnection) handleJoin(bizReq *common.BizMessage) (bizResp *common.BizMessage, err error) {
	var (
		bizJoinData *common.BizJoinData
		existed bool
	)
	bizJoinData = &common.BizJoinData{}
	if err = json.Unmarshal(bizReq.Data, bizJoinData); err != nil {
		return
	}
	if len(bizJoinData.Room) == 0 {
		err = common.ERR_ROOM_ID_INVALID
		return
	}
	if len(wsConnection.rooms) >= G_config.MaxJoinRoom {
		// 超过了房间数量限制, 忽略这个请求
		return
	}
	// 已加入过
	if _, existed = wsConnection.rooms[bizJoinData.Room]; existed {
		// 忽略掉这个请求
		return
	}
	// 建立房间 -> 连接的关系
	if err = G_connMgr.JoinRoom(bizJoinData.Room, wsConnection); err != nil {
		return
	}
	// 建立连接 -> 房间的关系
	wsConnection.rooms[bizJoinData.Room] = true
	return
}

// 处理LEAVE请求
func (wsConnection *WSConnection) handleLeave(bizReq *common.BizMessage) (bizResp *common.BizMessage, err error) {
	var (
		bizLeaveData *common.BizLeaveData
		existed bool
	)
	bizLeaveData = &common.BizLeaveData{}
	if err = json.Unmarshal(bizReq.Data, bizLeaveData); err != nil {
		return
	}
	if len(bizLeaveData.Room) == 0 {
		err = common.ERR_ROOM_ID_INVALID
		return
	}
	// 未加入过
	if _, existed = wsConnection.rooms[bizLeaveData.Room]; !existed {
		// 忽略掉这个请求
		return
	}
	// 删除房间 -> 连接的关系
	if err = G_connMgr.LeaveRoom(bizLeaveData.Room, wsConnection); err != nil {
		return
	}
	// 删除连接 -> 房间的关系
	delete(wsConnection.rooms, bizLeaveData.Room)
	return
}

func (wsConnection *WSConnection) leaveAll() {
	var (
		roomId string
	)
	// 从所有房间中退出
	for roomId, _ = range wsConnection.rooms {
		G_connMgr.LeaveRoom(roomId, wsConnection)
		delete(wsConnection.rooms, roomId)
	}
}

// 处理websocket请求
func (wsConnection *WSConnection) WSHandle() {
	var (
		message *common.WSMessage
		bizReq *common.BizMessage
		bizResp *common.BizMessage
		err error
		buf []byte
	)

	// 连接加入管理器, 可以推送端查找到
	G_connMgr.AddConn(wsConnection)

	// 心跳检测线程
	go wsConnection.heartbeatChecker()

	// 请求处理协程
	for {
		if message, err = wsConnection.ReadMessage(); err != nil {
			goto ERR
		}

		// 只处理文本消息
		if message.MsgType != websocket.TextMessage {
			continue
		}

		// 解析消息体
		if bizReq, err = common.DecodeBizMessage(message.MsgData); err != nil {
			goto ERR
		}

		bizResp = nil

		// 1,收到PING则响应PONG: {"type": "PING"}, {"type": "PONG"}
		// 2,收到JOIN则加入ROOM: {"type": "JOIN", "data": {"room": "chrome-plugin"}}
		// 3,收到LEAVE则离开ROOM: {"type": "LEAVE", "data": {"room": "chrome-plugin"}}

		// 请求串行处理
		switch bizReq.Type {
		case "PING":
			if bizResp, err = wsConnection.handlePing(bizReq); err != nil {
				goto ERR
			}
		case "JOIN":
			if bizResp, err = wsConnection.handleJoin(bizReq); err != nil {
				goto ERR
			}
		case "LEAVE":
			if bizResp, err = wsConnection.handleLeave(bizReq); err != nil {
				goto ERR
			}
		}

		if bizResp != nil {
			if buf, err = json.Marshal(*bizResp); err != nil {
				goto ERR
			}
			// socket缓冲区写满不是致命错误
			if err = wsConnection.SendMessage(&common.WSMessage{websocket.TextMessage, buf}); err != nil {
				if err != common.ERR_SEND_MESSAGE_FULL {
					goto ERR
				} else {
					err = nil
				}
			}
		}
	}

ERR:
	// 确保连接关闭
	wsConnection.Close()

	// 离开所有房间
	wsConnection.leaveAll()

	// 从连接池中移除
	G_connMgr.DelConn(wsConnection)
	return
}