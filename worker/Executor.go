package worker

import (
	"crontab/common"
	"os/exec"
	"time"
)

// 任务执行器
type Executor struct {
}

var G_executor *Executor

// 执行任务
/*
执行完成shell之后从执行计划表删除
*/
func (e *Executor) ExecutorJob(jbinfo *common.JobExecuteInfo) {
	result := &common.JobExecuteResult{
		ExecuteInfo: jbinfo,
		OutPut:      make([]byte, 0),
	}
	go func() {
		// 初始化锁
		jobLock := G_jobMgr.CreatJobLock(jbinfo.Job.Name)
		err := jobLock.TryLock()
		jobLock.TryLock() // 上锁
		defer jobLock.UnLock()
		if err != nil {
			result.Err = err
			result.EndTime = time.Now()
		} else {
			result.StartTime = time.Now()
			cmd := exec.CommandContext(jbinfo.ExecuteinfoCtx, "E://Git//bin//bash.exe", "-c", jbinfo.Job.Command)
			cmdout, err := cmd.CombinedOutput()
			result.EndTime = time.Now()
			result.OutPut = cmdout
			result.Err = err
		}
		//把输出结果返回给调度协程，把任务重执行计划表中删除
		G_scheduler.PushResult(result)
	}()
}

func InitExecutor() (err error) {
	G_executor = &Executor{}
	return
}
