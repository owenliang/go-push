package gateway

import (
	"encoding/json"
	"time"
	"github.com/owenliang/go-push/common"
)

type PushBatch struct {
	items []*json.RawMessage
	commitTimer *time.Timer

	// union {
	room string // 按room合并
	// }
}

type PushContext struct {
	msg *json.RawMessage

	// union {
	room string // 按room合并
	// }
}

type MergeWorker struct {
	mergeType int // 合并类型: 广播, room, uid...

	contextChan chan*PushContext
	timeoutChan chan*PushBatch

	// union {
	room2Batch map[string]*PushBatch	// room合并
	allBatch *PushBatch // 广播合并
	// }
}

// 广播消息、房间消息的合并
type Merger struct {
	roomWorkers  []*MergeWorker	// 房间合并
	broadcastWorker *MergeWorker	// 广播合并
}

var (
	G_merger *Merger
)

func (worker *MergeWorker) autoCommit(batch *PushBatch) func() {
	return func() {
		worker.timeoutChan <- batch
	}
}

func (worker *MergeWorker) commitBatch(batch *PushBatch) (err error) {
	var (
		bizPushData *common.BizPushData
		bizMessage *common.BizMessage
		buf []byte
	)

	bizPushData = &common.BizPushData{
		Items: batch.items,
	}
	if buf, err = json.Marshal(*bizPushData); err != nil {
		return
	}

	bizMessage = &common.BizMessage{
		Type: "PUSH",
		Data: json.RawMessage(buf),
	}

	// 打包发送
	if worker.mergeType == common.PUSH_TYPE_ROOM {
		delete(worker.room2Batch, batch.room)
		err = G_connMgr.PushRoom(batch.room, bizMessage)
	} else if worker.mergeType == common.PUSH_TYPE_ALL {
		worker.allBatch = nil
		err = G_connMgr.PushAll(bizMessage)
	}
	return
}

func (worker *MergeWorker) mergeWorkerMain() {
	var (
		context *PushContext
		batch *PushBatch
		timeoutBatch *PushBatch
		existed bool
		isCreated bool
		err error
	)
	for {
		select {
		case context = <- worker.contextChan:
			MergerPending_DESC()

			isCreated = false
			// 按房间合并
			if worker.mergeType == common.PUSH_TYPE_ROOM {
				if batch, existed = worker.room2Batch[context.room]; !existed {
					batch = &PushBatch{room: context.room}
					worker.room2Batch[context.room] = batch
					isCreated = true
				}
			} else if worker.mergeType == common.PUSH_TYPE_ALL {	// 广播合并
				batch = worker.allBatch
				if batch == nil {
					batch = &PushBatch{}
					worker.allBatch = batch
					isCreated = true
				}
			}

			// 合并消息
			batch.items = append(batch.items, context.msg)

			// 新建批次, 启动超时自动提交
			if isCreated {
				batch.commitTimer = time.AfterFunc(time.Duration(G_config.MaxMergerDelay) * time.Millisecond, worker.autoCommit(batch))
			}

			// 批次未满, 继续等待下次提交
			if len(batch.items) < G_config.MaxMergerBatchSize {
				continue
			}

			// 批次已满, 取消超时自动提交
			batch.commitTimer.Stop()
		case timeoutBatch = <- worker.timeoutChan:
			if worker.mergeType == common.PUSH_TYPE_ROOM {
				// 定时器触发时, 批次已被提交
				if batch, existed = worker.room2Batch[timeoutBatch.room]; !existed {
					continue
				}

				// 定时器触发时, 前一个批次已提交, 下一个批次已建立
				if batch != timeoutBatch {
					continue
				}
			} else if worker.mergeType == common.PUSH_TYPE_ALL {
				batch = worker.allBatch
				// 定时器触发时, 批次已被提交
				if timeoutBatch != batch {
					continue
				}
			}
		}
		// 提交批次
		err = worker.commitBatch(batch)

		// 打点统计
		if worker.mergeType == common.PUSH_TYPE_ALL {
			MergerAllTotal_INCR(int64(len(batch.items)))
			if err != nil {
				MergerAllFail_INCR(int64(len(batch.items)))
			}
		} else if worker.mergeType == common.PUSH_TYPE_ROOM {
			MergerRoomTotal_INCR(int64(len(batch.items)))
			if err != nil {
				MergerRoomFail_INCR(int64(len(batch.items)))
			}
		}
	}
}

func initMergeWorker(mergeType int) (worker *MergeWorker) {
	worker = &MergeWorker{
		mergeType: mergeType,
		room2Batch: make(map[string]*PushBatch),
		contextChan: make(chan*PushContext, G_config.MergerChannelSize),
		timeoutChan: make(chan*PushBatch, G_config.MergerChannelSize),
	}
	go worker.mergeWorkerMain()
	return
}

func (worker *MergeWorker) pushRoom(room string, msg *json.RawMessage) (err error) {
	var (
		context *PushContext
	)
	context = &PushContext{
		room: room,
		msg: msg,
	}
	select {
	case worker.contextChan <- context:
		MergerPending_INCR()
	default:
		err = common.ERR_MERGE_CHANNEL_FULL
	}
	return
}

func (worker *MergeWorker) pushAll(msg *json.RawMessage) (err error) {
	var (
		context *PushContext
	)
	context = &PushContext{
		msg: msg,
	}
	select {
	case worker.contextChan <- context:
		MergerPending_INCR()
	default:
		err = common.ERR_MERGE_CHANNEL_FULL
	}
	return
}

/**
	API
 */

func InitMerger() (err error) {
	var (
		workerIdx int
		merger *Merger
	)

	merger = &Merger{
		roomWorkers: make([]*MergeWorker, G_config.MergerWorkerCount),
	}
	for workerIdx = 0; workerIdx < G_config.MergerWorkerCount; workerIdx++ {
		merger.roomWorkers[workerIdx] = initMergeWorker(common.PUSH_TYPE_ROOM)
	}
	merger.broadcastWorker = initMergeWorker(common.PUSH_TYPE_ALL)

	G_merger = merger
	return
}

// 广播合并推送
func (merger *Merger) PushAll(msg *json.RawMessage) (err error) {
	return merger.broadcastWorker.pushAll(msg)
}

// 房间合并推送
func (merger *Merger) PushRoom(room string, msg *json.RawMessage) (err error) {
	// 计算room hash到某个worker
	var (
		workerIdx uint32= 0
		ch byte
	)
	for _, ch = range []byte(room) {
		workerIdx = (workerIdx + uint32(ch) * 33) % uint32(G_config.MergerWorkerCount)
	}
	return merger.roomWorkers[workerIdx].pushRoom(room, msg)
}