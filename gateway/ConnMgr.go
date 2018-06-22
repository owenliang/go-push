package gateway

import (
	"sync"
)

// 连接管理器
type ConnMgr struct {
	mutex sync.Mutex

	buckets []*Bucket
}

var (
	G_connMgr *ConnMgr
)

/**
	以下是API
 */

func InitConnMgr() (err error) {
	var (
		connMgr *ConnMgr
	)

	connMgr = &ConnMgr{
		buckets: make([]*Bucket, G_config.BucketCount),
	}

	for idx, _ := range connMgr.buckets {
		connMgr.buckets[idx] = InitBucket(idx)
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