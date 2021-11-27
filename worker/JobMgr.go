package worker

import (
	"context"
	"crontab/common"
	"time"

	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

//任务管理器
type JobMgr struct {
	client  *clientv3.Client
	kv      clientv3.KV
	lease   clientv3.Lease
	watcher clientv3.Watcher
}

//单例
var (
	G_jobMgr *JobMgr
)

func InitJobMgr() (err error) {
	config := clientv3.Config{
		Endpoints:   G_config.EtcdEndPoints,                                     //集群地址
		DialTimeout: time.Duration(G_config.EtcdDialTimeout) * time.Millisecond, //链接超时
	}
	client, err := clientv3.New(config)
	if err != nil {
		return err
	}
	//得到kv和lease的api
	kv := clientv3.NewKV(client)
	lease := clientv3.NewLease(client)
	watcher := clientv3.NewWatcher(client)

	G_jobMgr = &JobMgr{
		client:  client,
		kv:      kv,
		lease:   lease,
		watcher: watcher,
	}
	//启动监听
	if err := G_jobMgr.watchJobs(); err != nil {
		return err
	}
	G_jobMgr.watchKill()

	return
}

func (j *JobMgr) watchJobs() (err error) {
	//获取所有的job
	jobDir := common.JOB_SAVE_DIR
	getResp, err := j.kv.Get(context.TODO(), jobDir, clientv3.WithPrefix())
	if err != nil {
		return
	}
	for _, kvPair := range getResp.Kvs {
		//凡序列化
		jobs, err := common.UnpackJob(kvPair.Value)
		if err == nil {
			jobEvent := common.NewBuildJobEvent(common.JOB_EVENT_SAVE, jobs)
			//调度协程里面scheduler
			G_scheduler.PushJobEvent(jobEvent)
		}
	}
	//监听协程，监听版本变化
	go func() {
		//获取版本
		watchStartRevsion := getResp.Header.Revision + 1
		//监听后续变化
		watchChan := j.watcher.Watch(context.TODO(), common.JOB_SAVE_DIR, clientv3.WithRev(watchStartRevsion), clientv3.WithPrefix())
		for watchRes := range watchChan {
			for _, watchEvent := range watchRes.Events {
				switch watchEvent.Type {
				case mvccpb.PUT: //保存任务事件
					job, err := common.UnpackJob(watchEvent.Kv.Value)
					if err != nil {
						continue
					}
					//构建一个保存任务event
					jobEvent := common.NewBuildJobEvent(common.JOB_EVENT_SAVE, job)
					//同步到调度携程的channel里
					G_scheduler.PushJobEvent(jobEvent)
				case mvccpb.DELETE: //删除任务事件
					jobName := common.ExtraceName(string(watchEvent.Kv.Key))
					job := &common.Job{Name: jobName}
					//删除任务event
					jobEvent := common.NewBuildJobEvent(common.JOB_EVENT_DELETE, job)
					//同步到调度携程的channel里
					G_scheduler.PushJobEvent(jobEvent)
				}
			}
		}
	}()
	return
}

// 监听强杀目录
func (j *JobMgr) watchKill() {
	go func() {
		watchChan := j.watcher.Watch(context.TODO(), common.KILL_JOB_DIR, clientv3.WithPrefix())
		for watchRes := range watchChan {
			for _, event := range watchRes.Events {
				switch event.Type {
				case mvccpb.PUT: // kill任务
					//提取jobname
					jobName := common.ExtraceKillName(string(event.Kv.Key))
					job := &common.Job{Name: jobName}
					jobEvent := common.NewBuildJobEvent(common.JOB_EVENT_KILL, job)
					G_scheduler.PushJobEvent(jobEvent)
				case mvccpb.DELETE:
				}
			}
		}
	}()
}

// 创建锁
func (j *JobMgr) CreatJobLock(name string) *JobLock {
	return InitJobLock(name, j.kv, j.lease)
}
