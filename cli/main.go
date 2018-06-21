package main

import (
	"github.com/owenliang/go-push"
	"fmt"
	"os"
	"time"
)

func main()  {
	var (
		err error
	)

	// 初始化连接管理器
	if err = go_push.InitConnMgr(); err != nil {
		goto ERR
	}

	// 初始化websocket服务器
	if err = go_push.InitWSServer(); err != nil {
		goto ERR
	}

	for {
		time.Sleep(1 * time.Second)
	}

	os.Exit(0)

ERR:
	fmt.Fprintln(os.Stderr, err)
	os.Exit(-1)
}
