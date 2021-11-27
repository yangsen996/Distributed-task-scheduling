package worker

import (
	"context"
	"crontab/common"
	"net"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type Register struct {
	client *clientv3.Client
	kv     clientv3.KV
	lease  clientv3.Lease

	localIP string
}

var (
	G_register *Register
)

func getLocalIP() (ipv4 string, err error) {
	addrs, err := net.InterfaceAddrs()
	for _, addr := range addrs {
		ipNet, isipNet := addr.(*net.IPNet)
		if isipNet && ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				ipv4 = ipNet.IP.String()
				return
			}
		}
	}
	err = common.ERR_LOCAL_IP_NOT_FOUND
	return
}

func (r *Register) keepOnline() {
	var (
		cancelCtx  context.Context
		cancelFunc context.CancelFunc
		keepChan   <-chan *clientv3.LeaseKeepAliveResponse
	)
	cancelFunc = nil
	for {
		reKey := common.JOB_REGISTER_DIR + r.localIP
		leaseResp, err := r.lease.Grant(context.TODO(), 10)
		if err != nil {
			goto RETRY
		}
		// 续租
		keepChan, err = r.lease.KeepAlive(context.TODO(), leaseResp.ID)
		if err != nil {
			goto RETRY
		}
		cancelCtx, cancelFunc = context.WithCancel(context.TODO())
		// 注册
		_, err = r.kv.Put(cancelCtx, reKey, "", clientv3.WithLease(leaseResp.ID))
		if err != nil {
			goto RETRY
		}
		//处理应答
		for {
			select {
			case keepResp := <-keepChan:
				if keepResp == nil {
					goto RETRY
				}
			}
		}
	RETRY:
		time.Sleep(1 * time.Second)
		if cancelFunc != nil {
			cancelFunc()
		}

	}
}

func InitRegister() (err error) {
	config := clientv3.Config{
		Endpoints:   G_config.EtcdEndPoints,                                     //集群地址
		DialTimeout: time.Duration(G_config.EtcdDialTimeout) * time.Millisecond, //链接超时
	}
	client, err := clientv3.New(config)
	if err != nil {
		return err
	}
	localIp, err := getLocalIP()
	if err != nil {
		return
	}
	//得到kv和lease的api
	kv := clientv3.NewKV(client)
	lease := clientv3.NewLease(client)
	G_register = &Register{
		client:  client,
		kv:      kv,
		lease:   lease,
		localIP: localIp,
	}
	go G_register.keepOnline()
	return
}
