package main

import (
	"github.com/owenliang/go-push"
	"fmt"
	"os"
	"time"
	"flag"
	"runtime"
)

var (
	confFile string		// 配置文件路径
)

func initArgs() {
	flag.StringVar(&confFile, "config", "./go-push.json", "where go-push.json is.")
	flag.Parse()
}

func initEnv() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main()  {
	var (
		err error
	)

	// 初始化环境
	initArgs()
	initEnv()

	// 加载配置
	if err = go_push.InitConfig(confFile); err != nil {
		goto ERR
	}

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
