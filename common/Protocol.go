package common

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/gorhill/cronexpr"
)

//任务
type Job struct {
	Name     string `json:"name"`     //任务名
	Command  string `json:"command"`  //shell
	CronExpr string `json:"cronExpr"` //时间
}

//任务事件
type JobEvent struct {
	JobEventType int
	Job          *Job
}

//调度计划
type SchedulerPlan struct {
	Job      *Job                 //任务
	Expr     *cronexpr.Expression //表达式
	NextTime time.Time
}

//任务执行状态
type JobExecuteInfo struct {
	Job                   *Job
	PlanTime              time.Time //计划调度时间
	RealTime              time.Time //实际调度时间
	ExecuteinfoCtx        context.Context
	ExecuteinfoCancelFunc context.CancelFunc
}

// 任务执行结构
type JobExecuteResult struct {
	ExecuteInfo *JobExecuteInfo //执行状态
	OutPut      []byte          //脚本输出
	Err         error           //脚本错误原因
	StartTime   time.Time       //开始时间
	EndTime     time.Time       //结束时间
}

// 任务日志
type JobLog struct {
	JobName      string `bson:"jobName"` // 任务名称
	Command      string `bson:"command"` // 脚本命令
	Err          string `bson:"err"`
	OutPut       string `bson:"outPut"`
	PlanTime     int64  `bson:"planTime"`
	ScheduleTime int64  `bson:"scheduleTime"`
	StartTime    int64  `bson:"startTime"`
	EndTime      int64  `bson:"endTime"`
}

// 日志批次
type LogBatch struct {
	Logs []interface{}
}

// 日志过滤条件
type JobLogFilter struct {
	JobName string `bson:"jobName"`
}
type SortByStartTime struct {
	SortOrder int `bson:"startTime`
}

// type Response struct {
// 	Code int         `json:"code"`
// 	Msg  string      `json:"msg"`
// 	Data interface{} `json:"data"`
// }

// func BuileResponse(code int, msg string, data interface{}) (res *Response) {
// 	return &Response{
// 		Code: code,
// 		Msg:  msg,
// 		Data: data,
// 	}
// }

func UnpackJob(value []byte) (ret *Job, err error) {
	job := &Job{}
	if err = json.Unmarshal(value, job); err != nil {
		return
	}
	ret = job
	return
}

func ExtraceName(key string) string {
	// "cron/jobs/job10" => "job10"
	return strings.TrimPrefix(key, JOB_SAVE_DIR)
}
func ExtraceKillName(key string) string {
	return strings.TrimPrefix(key, KILL_JOB_DIR)
}
func ExtraceIP(key string) string {
	return strings.TrimPrefix(key, JOB_REGISTER_DIR)
}
func NewBuildJobEvent(eventType int, job *Job) *JobEvent {
	return &JobEvent{
		JobEventType: eventType,
		Job:          job,
	}
}

//构造执行计划
func NewBuildSchedulePlan(job *Job) (schedulePlan *SchedulerPlan, err error) {
	expr, err := cronexpr.Parse(job.CronExpr)
	if err != nil {
		return
	}
	schedulePlan = &SchedulerPlan{
		Job:      job,
		Expr:     expr,
		NextTime: expr.Next(time.Now()),
	}
	return
}

// 构造执行状态信息
func NewBuildJobExecuteInfo(sp *SchedulerPlan) (jobExecuteInfo *JobExecuteInfo) {
	jobExecuteInfo = &JobExecuteInfo{
		Job:      sp.Job,
		PlanTime: sp.NextTime,
		RealTime: time.Now(),
	}
	jobExecuteInfo.ExecuteinfoCtx, jobExecuteInfo.ExecuteinfoCancelFunc = context.WithCancel(context.TODO())
	return
}
