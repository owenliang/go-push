package gateway

import (
	"sync"
	"fmt"
)

type Bucket struct {
	rwMutex sync.RWMutex
	index int // 我是第几个桶
	id2Conn map[uint64]*WSConnection	// 连接列表(key=连接唯一ID)
	rooms map[string]*Room // 房间列表
}

func InitBucket(bucketIdx int) (bucket *Bucket) {
	bucket = &Bucket{
		index: bucketIdx,
		id2Conn: make(map[uint64]*WSConnection),
		rooms: make(map[string]*Room),
	}
	return
}

func (bucket *Bucket) AddConn(wsConn *WSConnection) {
	bucket.rwMutex.Lock()
	defer bucket.rwMutex.Unlock()

	bucket.id2Conn[wsConn.connId] = wsConn

	fmt.Println("AddConn", bucket.index, *wsConn)
}

func (bucket *Bucket) DelConn(wsConn *WSConnection) {
	bucket.rwMutex.Lock()
	defer bucket.rwMutex.Unlock()

	delete(bucket.id2Conn, wsConn.connId)

	fmt.Println("DelConn", bucket.index, *wsConn)
}

func (bucket *Bucket) JoinRoom(roomId string, wsConn *WSConnection) (err error) {
	var (
		existed bool
		room *Room
	)
	bucket.rwMutex.Lock()
	defer bucket.rwMutex.Unlock()

	// 找到房间
	if room, existed = bucket.rooms[roomId]; !existed {
		room = InitRoom(roomId)
		bucket.rooms[roomId] = room
	}
	// 加入房间
	err = room.Join(wsConn)
	return
}

func (bucket *Bucket) LeaveRoom(roomId string, wsConn *WSConnection) (err error) {
	var (
		existed bool
		room *Room
	)
	bucket.rwMutex.Lock()
	defer bucket.rwMutex.Unlock()

	// 找到房间
	if room, existed = bucket.rooms[roomId]; !existed {
		err = ERR_NOT_IN_ROOM
		return
	}

	err = room.Leave(wsConn)

	// 房间为空, 则删除
	if room.Count() == 0 {
		delete(bucket.rooms, roomId)
	}
	return
}