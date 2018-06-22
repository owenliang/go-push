package gateway

import (
	"fmt"
	"time"
	"github.com/gorilla/websocket"
	"encoding/json"
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
	fmt.Println("heartbeatChecker退出:", *wsConnection)
}

// 按秒粒度触发合并推送
func (wsConnection *WSConnection) batchCommitChecker() {
	var (
		now time.Time
		timer *time.Timer
		batch []*json.RawMessage
	)

	timer = time.NewTimer(time.Duration(G_config.MaxPushDelay) * time.Millisecond)

	for {
		select {
		case <- wsConnection.closeChan: // 需要感知socket被关闭
			goto EXIT
		case <- wsConnection.resetNotify: // 批次被提交, 重置定时器
		case <- timer.C:	// 定时器
			now = time.Now()
			// 换出一个batch
			batch = nil
			wsConnection.mutex.Lock()
			if len(wsConnection.pushBatch) != 0 {
				batch = wsConnection.pushBatch
				wsConnection.pushBatch = make([]*json.RawMessage, 0)
				wsConnection.lastCommit = now
			}
			wsConnection.mutex.Unlock()
			if batch != nil {
				fmt.Println("定时提交", *wsConnection)
				wsConnection.commitBatch(batch)
			}
		}
		// 重置定时器
		timer.Reset(time.Duration(G_config.MaxPushDelay) * time.Millisecond)
	}
EXIT:
	fmt.Println("batchCommitChecker退出:", *wsConnection)
	timer.Stop()
}

// 处理PING请求
func (wsConnection *WSConnection) handlePing(bizReq *BizMessage) (bizResp *BizMessage, err error) {
	var (
		buf []byte
	)

	wsConnection.KeepAlive()

	if buf, err = json.Marshal(BizPongData{}); err != nil {
		return
	}
	bizResp = &BizMessage{
		Type: "PONG",
		Data: json.RawMessage(buf),
	}
	return
}

// 处理JOIN请求
func (wsConnection *WSConnection) handleJoin(bizReq *BizMessage) (bizResp *BizMessage, err error) {
	var (
		bizJoinData *BizJoinData
		existed bool
	)
	bizJoinData = &BizJoinData{}
	if err = json.Unmarshal(bizReq.Data, bizJoinData); err != nil {
		return
	}
	if len(bizJoinData.Room) == 0 {
		err = ERR_ROOM_ID_INVALID
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
func (wsConnection *WSConnection) handleLeave(bizReq *BizMessage) (bizResp *BizMessage, err error) {
	var (
		bizLeaveData *BizLeaveData
		existed bool
	)
	bizLeaveData = &BizLeaveData{}
	if err = json.Unmarshal(bizReq.Data, bizLeaveData); err != nil {
		return
	}
	if len(bizLeaveData.Room) == 0 {
		err = ERR_ROOM_ID_INVALID
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
		message *WSMessage
		bizReq *BizMessage
		bizResp *BizMessage
		err error
		buf []byte
	)

	// 连接加入管理器, 可以推送端查找到
	G_connMgr.AddConn(wsConnection)

	// 心跳检测线程
	go wsConnection.heartbeatChecker()
	// 批量推送线程
	go wsConnection.batchCommitChecker()

	// 请求处理协程
	for {
		if message, err = wsConnection.ReadMessage(); err != nil {
			goto ERR
		}

		fmt.Println("消息", string(message.msgData))

		// 只处理文本消息
		if message.msgType != websocket.TextMessage {
			continue
		}

		// 解析消息体
		if bizReq, err = DecodeBizMessage(message.msgData); err != nil {
			fmt.Println("反序列化失败", *message)
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
			if err = wsConnection.SendMessage(&WSMessage{websocket.TextMessage, buf}); err != nil {
				if err != ERR_SEND_MESSAGE_FULL {
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

// 提交一批推送
func (wsConnection *WSConnection) commitBatch(batch []*json.RawMessage) (err error) {
	var (
		buf []byte
		bizMsg *BizMessage
		bizBuf []byte
	)

	// 打包多条推送
	if buf, err = json.Marshal(BizPushData{Items: batch}); err != nil {
		goto ERR
	}

	// 生成推送消息
	bizMsg = &BizMessage{
		Type: "PUSH",
		Data: json.RawMessage(buf),
	}
	// 序列化
	if bizBuf, err = json.Marshal(*bizMsg); err != nil {
		goto ERR
	}
	// socket缓冲区写满不是致命错误
	if err = wsConnection.SendMessage(&WSMessage{websocket.TextMessage, bizBuf}); err != nil {
		if err != ERR_SEND_MESSAGE_FULL {
			goto ERR
		} else {
			err = nil
		}
	}
	return

ERR:
	fmt.Println("批量提交失败", err)
	return
}

// 仅用于PUSH类型的消息
// 将推送消息做延迟发送, 在时间窗口内做消息打包合并
func (wsConnection *WSConnection) QueuePushForBatch(singlePush *json.RawMessage) (err error) {
	var (
		batch []*json.RawMessage
	)

	batch = nil

	wsConnection.mutex.Lock()

	wsConnection.pushBatch = append(wsConnection.pushBatch, singlePush)

	// 批次已满, 立即发送
	if len(wsConnection.pushBatch) >= G_config.MaxPushBatchSize {
		batch = wsConnection.pushBatch
		wsConnection.pushBatch = make([]*json.RawMessage, 0)
		wsConnection.lastCommit = time.Now()
		// 通知定时commit协程, 令其重置发送定时器
		select {
		case	wsConnection.resetNotify <- 1:
		default:
		}
	}

	wsConnection.mutex.Unlock()

	if batch != nil {
		fmt.Println("触发提交", *wsConnection)
		if err = wsConnection.commitBatch(batch); err != nil {
			goto ERR
		}
	}
	return

ERR:
	wsConnection.Close()
	return
}