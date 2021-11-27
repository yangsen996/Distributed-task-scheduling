package worker

import (
	"context"
	"crontab/common"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type LogSink struct {
	client         *mongo.Client
	logCollection  *mongo.Collection
	logChan        chan *common.JobLog
	autoCommitLogs chan *common.LogBatch
}

var (
	G_log *LogSink
)

func InitLog() (err error) {
	client, err := mongo.Connect(context.TODO(),
		options.Client().ApplyURI(G_config.MongodbUri),
		options.Client().SetConnectTimeout(time.Duration(G_config.MongodbConnectTimeout)*time.Millisecond),
	)
	if err != nil {
		return
	}
	G_log = &LogSink{
		client:         client,
		logCollection:  client.Database("cron").Collection("log"),
		logChan:        make(chan *common.JobLog, 1000),
		autoCommitLogs: make(chan *common.LogBatch, 1000),
	}
	G_log.writeLoop()
	return
}
func (l *LogSink) saveLogs(batch *common.LogBatch) {
	l.logCollection.InsertMany(context.TODO(), batch.Logs)
}

func (l *LogSink) writeLoop() {
	var (
		logBatch    *common.LogBatch
		commitTimer *time.Timer
	)
	for {
		select {
		case log := <-l.logChan:
			if logBatch == nil {
				logBatch = &common.LogBatch{}
				// 超时1s自动提交
				commitTimer = time.AfterFunc(1000*time.Millisecond, func(batch *common.LogBatch) func() {
					return func() {
						l.autoCommitLogs <- logBatch
					}
				}(logBatch),
				)
			}
			logBatch.Logs = append(logBatch.Logs, log)
			if len(logBatch.Logs) >= 100 {
				// 保存
				l.saveLogs(logBatch)
				logBatch = nil
				// 取消定时器
				commitTimer.Stop()
			}
		case timeoutBatch := <-l.autoCommitLogs:
			if timeoutBatch != logBatch {
				continue
			}
			l.saveLogs(timeoutBatch)
			logBatch = nil
		}
	}
}

func (l *LogSink) Append(joblog *common.JobLog) {

	select {
	case l.logChan <- joblog:
	default:
		// 满了就丢弃
	}
}
