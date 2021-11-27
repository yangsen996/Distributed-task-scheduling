package master

import (
	"context"
	"crontab/common"
	"encoding/json"
	"log"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

//任务管理器
type JobMgr struct {
	client *clientv3.Client
	kv     clientv3.KV
	lease  clientv3.Lease
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

	G_jobMgr = &JobMgr{
		client: client,
		kv:     kv,
		lease:  lease,
	}
	return
}

// 保存任务
func (j *JobMgr) SaveJob(job *common.Job) error {
	jobKey := common.JOB_SAVE_DIR + job.Name
	jobVal, err := json.Marshal(job)
	if err != nil {
		log.Println(err)
	}
	//put
	_, err = j.kv.Put(context.TODO(), jobKey, string(jobVal))
	if err != nil {
		return err
	}
	return err
	//判断是否更新，更新则返回oldjob
	//if putRes.PrevKv != nil {
	//	//反序列化
	//	var oldJobObj *common.Job
	//	err := json.Unmarshal(putRes.PrevKv.Value, oldJob)
	//	if err != nil {
	//		err = nil
	//	}
	//	oldJob = oldJobObj
	//	return
	//}
}

// 删除job
func (j *JobMgr) DeleteJob(jobName string) (err error) {
	jobKey := common.JOB_SAVE_DIR + jobName
	if _, err = j.kv.Delete(context.TODO(), jobKey); err != nil {
		log.Println("删除错误", err)
		return err
	}
	return
}

// 获取job列表
func (j *JobMgr) JobList() (jobList []*common.Job, err error) {
	dirName := common.JOB_SAVE_DIR
	putResp, err := j.kv.Get(context.TODO(), dirName, clientv3.WithPrefix())
	if err != nil {
		return
	}
	//遍历
	jobList = make([]*common.Job, 0)
	for _, kvPair := range putResp.Kvs {
		var job = &common.Job{}
		_ = json.Unmarshal(kvPair.Value, job)
		jobList = append(jobList, job)
	}
	return
}

// 杀死job
func (j *JobMgr) killJob(name string) (err error) {
	KillKey := common.KILL_JOB_DIR + name
	//创建1s租约
	leaseRes, err := j.lease.Grant(context.TODO(), 1)
	if err != nil {
		return
	}
	//获取id
	leaseID := leaseRes.ID
	_, err = j.kv.Put(context.TODO(), KillKey, "", clientv3.WithLease(leaseID))
	if err != nil {
		return
	}
	return
}
