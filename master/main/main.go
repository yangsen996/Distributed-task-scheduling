package main

import (
	"context"
	"crontab/master"
	"crontab/master/server"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"
	"time"
)

func intiEnv() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}
func main() {
	var err error
	if err = master.InitConfig(); err != nil {
		goto ERR
	}
	//初始化线程
	intiEnv()
	//启动管理器
	if err = master.InitJobMgr(); err != nil {
		goto ERR
	}
	if err = master.InitWorkerList(); err != nil {
		goto ERR
	}
	if err = master.InitLogMgr(); err != nil {
		goto ERR
	}
	// //启动服务
	// if err = master.InitApiServer(); err != nil {
	// 	goto ERR
	// }
ERR:
	fmt.Println(err)
	//注册路由
	r := server.SetUp()
	srv := &http.Server{
		Addr:    ":" + strconv.Itoa(master.G_config.ApiPort),
		Handler: r,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("listen:%s\n", err)
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-quit
	fmt.Println("shoutdown server....")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("server shuntdown err:", err)
	}
	log.Println("server exitting...")
}
