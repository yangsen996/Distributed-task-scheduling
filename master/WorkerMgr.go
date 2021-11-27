package master

import (
	"context"
	"crontab/common"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type WorkerMgr struct {
	client *clientv3.Client
	kv     clientv3.KV
}

var G_workerList *WorkerMgr

func InitWorkerList() (err error) {
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
	G_workerList = &WorkerMgr{
		client: client,
		kv:     kv,
	}
	return
}

func (w *WorkerMgr) WorkerList() (listArr []string, err error) {
	key := common.JOB_REGISTER_DIR
	listArr = make([]string, 0)
	getResp, err := w.kv.Get(context.TODO(), key, clientv3.WithPrefix())
	if err != nil {
		return
	}
	for _, kv := range getResp.Kvs {
		workerIP := common.ExtraceIP(string(kv.Key))
		listArr = append(listArr, workerIP)
	}
	return
}
