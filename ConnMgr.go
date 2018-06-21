package go_push

import (
	"sync"
	"fmt"
)

// 连接管理器
type ConnMgr struct {
	mutex sync.Mutex

	// 维护Buckets
}

var (
	G_connMgr *ConnMgr
)

/**
	以下是API
 */

func InitConnMgr() (err error) {
	G_connMgr = &ConnMgr{}
	return
}

func (connMgr *ConnMgr) AddConn(wsConnection *WSConnection) {
	fmt.Println("AddConn", *wsConnection)
}

func (connMgr *ConnMgr) DelConn(wsConnection *WSConnection) {
	fmt.Println("DelConn", *wsConnection)
}