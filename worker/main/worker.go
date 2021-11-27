package main

import (
	"crontab/worker"
	"log"
	"runtime"
	"time"
)

func initEnv() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	var err error
	//初始化配置
	if err = worker.InitConfig(); err != nil {
		goto ERR
	}
	// 初始化线程
	initEnv()
	// 初始化日志
	if err = worker.InitLog(); err != nil {
		goto ERR
	}
	worker.InitScheduler()
	// 初始化执行器
	if err = worker.InitExecutor(); err != nil {
		goto ERR
	}
	// 初始化调度器
	if err = worker.InitScheduler(); err != nil {
		goto ERR
	}
	//加载任务管理器
	if err = worker.InitJobMgr(); err != nil {
		goto ERR
	}
ERR:
	log.Println(err)
	for {
		time.Sleep(1 * time.Second)
	}
}
