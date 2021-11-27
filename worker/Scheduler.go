package worker

import (
	"crontab/common"
	"fmt"
	"log"
	"time"
)

// Scheduler 任务调度
type Scheduler struct {
	jobEventChan      chan *common.JobEvent             //etcd任务事件队列
	jobPlanTable      map[string]*common.SchedulerPlan  //任务调度计划表
	jobExecutingTable map[string]*common.JobExecuteInfo //任务执行计划表
	jobResultChan     chan *common.JobExecuteResult     //任务结果队列
}

var G_scheduler *Scheduler

// 处理job事件
func (s *Scheduler) handleJobEvent(jobEvent *common.JobEvent) {
	switch jobEvent.JobEventType {
	case common.JOB_EVENT_SAVE: //保存事件
		jobSchedulePlan, err := common.NewBuildSchedulePlan(jobEvent.Job)
		if err != nil {
			return
		}
		//保存到任务调度计划表
		s.jobPlanTable[jobEvent.Job.Name] = jobSchedulePlan
	case common.JOB_EVENT_DELETE: //删除事件
		//查看计划表是否有这个任务，有直接删除
		_, ok := s.jobPlanTable[jobEvent.Job.Name]
		if ok {
			delete(s.jobPlanTable, jobEvent.Job.Name)
		}
	case common.JOB_EVENT_KILL: // 强杀任务
		//判断任务是否在执行中
		job, ok := s.jobExecutingTable[jobEvent.Job.Name]
		if ok {
			// 杀死任务
			job.ExecuteinfoCancelFunc()
		}
	}
}

// 处理执行结果
func (s *Scheduler) handleJobResult(result *common.JobExecuteResult) {
	//再执行计划表删除任务
	delete(s.jobExecutingTable, result.ExecuteInfo.Job.Name)
	//生成日志
	if result.Err != common.ERR_LOCK_ALREADY_REQUIRED {
		jobLog := &common.JobLog{
			JobName:      result.ExecuteInfo.Job.Name,
			Command:      result.ExecuteInfo.Job.Command,
			OutPut:       string(result.OutPut),
			PlanTime:     result.ExecuteInfo.PlanTime.UnixNano() / 1000 / 1000,
			ScheduleTime: result.ExecuteInfo.PlanTime.UnixNano() / 1000 / 1000,
			StartTime:    result.StartTime.UnixNano() / 1000 / 1000,
			EndTime:      result.EndTime.UnixNano() / 1000 / 1000,
		}
		if result.Err != nil {
			jobLog.Err = string(result.Err.Error())
		} else {
			jobLog.Err = ""
		}
		//存储到mongodb
		G_log.Append(jobLog)
	}
	log.Println("任务执行完成", result.ExecuteInfo.Job.Name, "输出", string(result.OutPut), "错误", result.Err)
}

// TryStartJob 尝试执行任务
func (s *Scheduler) TryStartJob(jobPlan *common.SchedulerPlan) {
	//任务正在执行跳过任务
	if _, ok := s.jobExecutingTable[jobPlan.Job.Name]; ok {
		// fmt.Println("尚未退出，跳过执行", jobPlan.Job.Name)
		return
	}
	//构建执行状态
	jobExecuteInfo := common.NewBuildJobExecuteInfo(jobPlan)
	//保存
	s.jobExecutingTable[jobPlan.Job.Name] = jobExecuteInfo
	//执行任务
	G_executor.ExecutorJob(jobExecuteInfo)
	fmt.Println("执行任务", jobExecuteInfo.Job.Name, "计划时间", jobExecuteInfo.PlanTime, "真实时间", jobExecuteInfo.RealTime)
}

//TryScheduler 计算时间
func (s *Scheduler) TryScheduler() (schedulerAfter time.Duration) {

	if len(s.jobPlanTable) == 0 {
		schedulerAfter = 1 * time.Second
	}
	//遍历所有任务
	var nearTime = new(time.Time)
	now := time.Now()
	for _, jobPlan := range s.jobPlanTable {
		if jobPlan.NextTime.Before(now) || jobPlan.NextTime.Equal(now) {
			//过期立即执行
			//执行任务（注意如果下次执行时间到了，任务还没有结束，就不能执行，特殊处理）
			s.TryStartJob(jobPlan)
			//重新计算
			jobPlan.NextTime = jobPlan.Expr.Next(now)
		}
		//统计最近要过期的任务时间
		if nearTime == nil || jobPlan.NextTime.Before(*nearTime) {
			nearTime = &jobPlan.NextTime
		}
	}
	schedulerAfter = (*nearTime).Sub(now)
	return
}

//调度协程
func (s *Scheduler) scheduleLoop() {
	//初始化
	schedulerAfter := s.TryScheduler()
	//延迟调度器
	scheduleTimer := time.NewTimer(schedulerAfter)
	//读取job
	for {
		select {
		case jobEvent := <-s.jobEventChan: //监听任务事件变化
			//进行增删改查
			s.handleJobEvent(jobEvent)
		case <-scheduleTimer.C: //最近的任务到期了
		case jobResult := <-s.jobResultChan:
			s.handleJobResult(jobResult)
		}
		//调度任务
		schedulerAfter = s.TryScheduler()
		//重置定时器
		scheduleTimer.Reset(schedulerAfter)
	}
}

// PushJobEvent 同步jobEvent
func (s *Scheduler) PushJobEvent(jobEvent *common.JobEvent) {
	s.jobEventChan <- jobEvent
}
func InitScheduler() (err error) {
	G_scheduler = &Scheduler{
		jobEventChan:      make(chan *common.JobEvent, 1000),
		jobPlanTable:      make(map[string]*common.SchedulerPlan),
		jobExecutingTable: make(map[string]*common.JobExecuteInfo),
		jobResultChan:     make(chan *common.JobExecuteResult, 1000),
	}
	//启动协程
	go G_scheduler.scheduleLoop()
	return
}

func (s *Scheduler) PushResult(jobResult *common.JobExecuteResult) {
	s.jobResultChan <- jobResult
}
