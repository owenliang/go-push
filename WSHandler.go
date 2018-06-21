package go_push

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
	timer = time.NewTimer(1 * time.Second)
	for {
		select {
		case <- timer.C:
			if !wsConnection.IsAlive() {
				wsConnection.Close()
				goto EXIT
			}
			timer.Reset(1 * time.Second)
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

// 处理websocket请求
func (wsConnection *WSConnection) WSHandle() {
	var (
		message *WSMessage
		bizReq *BizMessage
		bizResp *BizMessage
		bizJoinData *BizJoinData
		bizLeaveData *BizLeaveData
		err error
		buf []byte
	)

	go wsConnection.heartbeatChecker()
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

		// TODO: 客户端提交如下3种请求:
		// 1,收到PING则响应PONG: {"type": "PING"}, {"type": "PONG"}
		// 2,收到JOIN则加入ROOM: {"type": "JOIN", "data": {"room": "chrome-plugin"}}
		// 3,收到LEAVE则离开ROOM: {"type": "LEAVE", "data": {"room": "chrome-plugin"}}
		switch bizReq.Type {
		case "PING":
			wsConnection.KeepAlive()

			if buf, err = json.Marshal(BizPongData{}); err != nil {
				goto ERR
			}
			bizResp = &BizMessage{
				Type: "PONG",
				Data: json.RawMessage(buf),
			}
		case "JOIN":
			bizJoinData = &BizJoinData{}
			if err = json.Unmarshal(bizReq.Data, bizJoinData); err != nil {
				goto ERR
			}
			fmt.Println("JOIN:", *bizJoinData)
		case "LEAVE":
			bizLeaveData = &BizLeaveData{}
			if err = json.Unmarshal(bizReq.Data, bizLeaveData); err != nil {
				goto ERR
			}
			fmt.Println("LEAVE:", *bizLeaveData)
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
	return

ERR:
	wsConnection.Close()
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