package worker

import (
	"context"
	"crontab/common"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type JobLock struct {
	kv         clientv3.KV
	lease      clientv3.Lease
	JobName    string
	cancelFunc context.CancelFunc
	leaseID    clientv3.LeaseID
	IsJobLock  bool
}

func InitJobLock(name string, kv clientv3.KV, lease clientv3.Lease) *JobLock {
	return &JobLock{
		kv:      kv,
		lease:   lease,
		JobName: name,
	}
}

func (j *JobLock) TryLock() (err error) {
	//创建租约
	leaseResp, err := j.lease.Grant(context.TODO(), 5)
	if err != nil {
		return
	}
	leaseID := leaseResp.ID
	var (
		txn     clientv3.Txn
		lockKey string
		txnRes  *clientv3.TxnResponse
	)
	//自动续租
	ctx, cancelFunc := context.WithCancel(context.TODO())
	keepAliveChan, err := j.lease.KeepAlive(ctx, leaseID)
	if err != nil {
		goto FAIL
	}
	//应答
	go func() {
		for {
			select {
			case keepRes := <-keepAliveChan:
				if keepRes == nil { //租约过期失效
					goto END
				}
			}
		}
	END:
	}()
	//创建事务
	txn = j.kv.Txn(context.TODO())
	lockKey = common.JOB_LOCK_DIR + j.JobName
	//抢锁
	txn.If(clientv3.Compare(clientv3.CreateRevision(lockKey), "=", 0)).
		Then(clientv3.OpPut(lockKey, "", clientv3.WithLease(leaseID))).
		Else(clientv3.OpGet(lockKey))
	//提交事务
	txnRes, err = txn.Commit()
	if err != nil {
		goto FAIL
	}
	// 判断是否成功
	if !txnRes.Succeeded {
		err = common.ERR_LOCK_ALREADY_REQUIRED
		goto FAIL
	}
	//成功
	j.leaseID = leaseID
	j.cancelFunc = cancelFunc
	j.IsJobLock = true
FAIL:
	//取消续租
	cancelFunc()
	//释放租约
	_, _ = j.lease.Revoke(context.TODO(), leaseID)
	return
}

func (j *JobLock) UnLock() {
	if j.IsJobLock {
		j.cancelFunc()
		j.lease.Revoke(context.TODO(), j.leaseID)
	}
}
