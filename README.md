## master
* 搭建框架
* 配置文件
* 命令行参数
* 线程配置
* httpApi（增删改查）gin框架搭建
## worker
* 从etcd中把任务同步到内存中
   - 监听任务，获取所有任务存到调度协程进行调度（jobEventChannel<-）
   - 监听后续的版本，判断每个版本event，放到调度协程进行处理(jobEventChannel<-)
* 实现调度模块，基于cron表达式调度多个任务
  - 创建任务调取以及任务调度计划表
  - 从(<-jobEventChannel)中获取event的type变化，save事件保存到计划调度表，delete事件删除
  - 计算任务的下一次执行时间，从调度计划表中获取下一次调度时间，和当前时间判断，
* 实现执行模块，并发执行多个job
* 对job的分布式锁，防止集群并发
* 执行日志保存到mongoDB
## 服务注册于发现
* 启动后获取本机网卡ip，作为节点的唯一标识
* 启动服务注册协程，首先创建lease并自动续租
* 带着lease注册到/cron/workers/{ip}ip下，供服务发现