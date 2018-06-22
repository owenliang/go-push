package gateway

import "encoding/json"

// 推送任务
type PushJob struct {
	pushType int // 推送类型
	roomId string // 房间ID
	pushMsg *json.RawMessage	// 要推送的消息体
}

// 连接管理器
type ConnMgr struct {
	buckets []*Bucket
	jobChan []chan*PushJob	 // 每个Bucket对应一个Job Queue

	dispatchChan chan *PushJob	// 待分发消息队列
}

var (
	G_connMgr *ConnMgr
)

// 消息分发到Bucket
func (connMgr *ConnMgr)dispatchWorkerMain(dispatchWorkerIdx int) {
	var (
		bucketIdx int
		pushJob *PushJob
	)
	for {
		select {
		case pushJob = <- connMgr.dispatchChan:
			PushAllPending_DESC()
			// 分发给所有Bucket, 若Bucket拥塞则等待
			for bucketIdx, _ = range connMgr.buckets {
				PushJobPending_INCR()
				connMgr.jobChan[bucketIdx] <- pushJob
			}
		}
	}
}

// Job负责消息广播给客户端
func (connMgr *ConnMgr)jobWorkerMain(jobWorkerIdx int, bucketIdx int) {
	var (
		bucket = connMgr.buckets[bucketIdx]
		pushJob *PushJob
	)

	for {
		select {
		case pushJob = <-connMgr.jobChan[bucketIdx]:	// 从Bucket的job queue取出一个任务
			PushJobPending_DESC()
			if pushJob.pushType == PUSH_TYPE_ALL {
				bucket.PushAll(pushJob.pushMsg)
			} else if pushJob.pushType == PUSH_TYPE_ROOM {
				bucket.PushRoom(pushJob.roomId, pushJob.pushMsg)
			}
		}
	}
}

/**
	以下是API
 */

func InitConnMgr() (err error) {
	var (
		bucketIdx int
		jobWorkerIdx int
		dispatchWorkerIdx int
		connMgr *ConnMgr
	)

	connMgr = &ConnMgr{
		buckets: make([]*Bucket, G_config.BucketCount),
		jobChan: make([]chan*PushJob, G_config.BucketCount),
		dispatchChan: make(chan*PushJob, G_config.DispatchChannelSize),
	}
	for bucketIdx, _ = range connMgr.buckets {
		connMgr.buckets[bucketIdx] = InitBucket(bucketIdx)	// 初始化Bucket
		connMgr.jobChan[bucketIdx] = make(chan*PushJob, G_config.BucketJobChannelSize) // Bucket的Job队列
		// Bucket的Job worker
		for jobWorkerIdx = 0; jobWorkerIdx < G_config.BucketJobWorkerCount; jobWorkerIdx++ {
			go connMgr.jobWorkerMain(jobWorkerIdx, bucketIdx)
		}
	}
	// 初始化分发协程, 用于将消息扇出给各个Bucket
	for dispatchWorkerIdx = 0; dispatchWorkerIdx < G_config.DispatchWorkerCount; dispatchWorkerIdx++ {
		go connMgr.dispatchWorkerMain(dispatchWorkerIdx)
	}

	G_connMgr = connMgr
	return
}

func (connMgr *ConnMgr) GetBucket(wsConnection *WSConnection) (bucket *Bucket) {
	bucket = connMgr.buckets[wsConnection.connId % uint64(len(connMgr.buckets))]
	return
}

func (connMgr *ConnMgr) AddConn(wsConnection *WSConnection) {
	var (
		bucket *Bucket
	)

	bucket = connMgr.GetBucket(wsConnection)
	bucket.AddConn(wsConnection)
}

func (connMgr *ConnMgr) DelConn(wsConnection *WSConnection) {
	var (
		bucket *Bucket
	)

	bucket = connMgr.GetBucket(wsConnection)
	bucket.DelConn(wsConnection)
}

func (connMgr *ConnMgr) JoinRoom(roomId string, wsConn *WSConnection) (err error) {
	var (
		bucket *Bucket
	)

	bucket = connMgr.GetBucket(wsConn)
	err = bucket.JoinRoom(roomId, wsConn)
	return
}

func (connMgr *ConnMgr) LeaveRoom(roomId string, wsConn *WSConnection) (err error) {
	var (
		bucket *Bucket
	)

	bucket = connMgr.GetBucket(wsConn)
	err = bucket.LeaveRoom(roomId, wsConn)
	return
}

// 向所有在线用户推送
func (connMgr *ConnMgr) PushAll(pushMsg *json.RawMessage) (err error) {
	var (
		pushJob *PushJob
	)

	pushJob = &PushJob{
		pushType: PUSH_TYPE_ALL,
		pushMsg: pushMsg,
	}

	select {
	case 	connMgr.dispatchChan <- pushJob:
		PushAllPending_INCR()
	default:
		err = ERR_DISPATCH_CHANNEL_FULL
	}
	return
}

// 向指定房间推送
func (connMgr *ConnMgr) PushRoom(roomId string, pushMsg *json.RawMessage) (err error) {
	var (
		pushJob *PushJob
	)

	pushJob = &PushJob{
		pushType: PUSH_TYPE_ROOM,
		pushMsg: pushMsg,
		roomId: roomId,
	}

	select {
	case 	connMgr.dispatchChan <- pushJob:
	default:
		err = ERR_DISPATCH_CHANNEL_FULL
	}
	return
}