package server

import (
	"io"
	"os"

	"github.com/gin-gonic/gin"
)

func SetUp() *gin.Engine {
	// 禁用控制台颜色
	gin.DisableConsoleColor()
	// 日志写入文件和控制台
	f, _ := os.Create("crton.log")
	gin.DefaultWriter = io.MultiWriter(f, os.Stdout)
	r := gin.Default()
	v1 := r.Group("job")
	{
		v1.POST("/save", SaveHandle)
		v1.POST("/delete", DeleteHandle)
		v1.GET("/list", JobListHandle)
		v1.POST("/kill", killJobHandle)
		v1.GET("/job/log/:jobname/:skip/:limti", JobLogHandle)
		v1.GET("/healthy/worker")
	}
	return r
}
