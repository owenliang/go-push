package main

import (
	"github.com/owenliang/go-push/gateway"
	"fmt"
	"os"
	"flag"
	"runtime"
	"time"
)

var (
	confFile string		// 配置文件路径
)

func initArgs() {
	flag.StringVar(&confFile, "config", "./gateway.json", "where gateway.json is.")
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
	if err = gateway.InitConfig(confFile); err != nil {
		goto ERR
	}

	// 统计
	if err = gateway.InitStats(); err != nil {
		goto ERR
	}

	// 初始化连接管理器
	if err = gateway.InitConnMgr(); err != nil {
		goto ERR
	}

	// 初始化websocket服务器
	if err = gateway.InitWSServer(); err != nil {
		goto ERR
	}

	// 初始化merger合并层
	if err = gateway.InitMerger(); err != nil {
		goto ERR
	}

	// 初始化service接口
	if err = gateway.InitService(); err != nil {
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
