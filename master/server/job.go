package server

import (
	"crontab/common"
	"crontab/master"
	"crontab/utils"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func SaveHandle(c *gin.Context) {
	var (
		job *common.Job
	)
	//参数效验
	err := c.ShouldBindJSON(&job)
	if err != nil {
		utils.ResponseErr(c, utils.CodeErr)
		return
	}
	j := &common.Job{
		Name:     job.Name,
		Command:  job.Command,
		CronExpr: job.CronExpr,
	}
	//保存
	err = master.G_jobMgr.SaveJob(j)
	if err != nil {
		return
	}
	//if err != nil {
	//	c.JSON(http.StatusOK,gin.H{"msg":"错误"})
	//}
	log.Println(j.Name, "保存成功", "执行命令", j.Command, "表达式", j.CronExpr)
	c.JSON(http.StatusOK, gin.H{"msg": "success"})
}

func DeleteHandle(c *gin.Context) {
	jobName := c.PostForm("name")
	err := master.G_jobMgr.DeleteJob(jobName)
	if err != nil {
		// c.JSON(http.StatusOK,gin.H{"msg":"删除失败err","err":err})
		utils.ResponseErr(c, utils.CodeErr)
	}
	log.Println(jobName, "删除成功")
	utils.ResponseSuccess(c, utils.CodeSuccess)
}

func JobListHandle(c *gin.Context) {

}

func killJobHandle(c *gin.Context) {

}

func JobLogHandle(c *gin.Context) {

}
