package master

import (
	"context"
	"crontab/common"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type LogMgr struct {
	client        *mongo.Client
	logCollection *mongo.Collection
}

var G_logMgr *LogMgr

func InitLogMgr() (err error) {
	client, err := mongo.Connect(context.TODO(),
		options.Client().ApplyURI(G_config.MongodbUri),
		options.Client().SetConnectTimeout(time.Duration(G_config.MongodbConnectTimeout)*time.Millisecond),
	)
	if err != nil {
		return
	}
	G_logMgr = &LogMgr{
		client:        client,
		logCollection: client.Database("cron").Collection("log"),
	}
	return
}

// 根据id查询日志
func (l *LogMgr) ListLog(name string, skip int, limit int) (logArr []*common.JobLog) {
	filer := &common.JobLogFilter{JobName: name}
	logSort := &common.SortByStartTime{SortOrder: -1}
	logArr = make([]*common.JobLog, 0)
	cursor, err := l.logCollection.Find(context.TODO(), filer, options.Find().SetSort(logSort), options.Find().SetLimit(int64(limit)), options.Find().SetSkip(int64(skip)))
	if err != nil {
		return
	}
	defer cursor.Close(context.TODO())
	for cursor.Next(context.TODO()) {
		jobLog := &common.JobLog{}
		err := cursor.Decode(jobLog)
		if err != nil {
			continue
		}
		logArr = append(logArr, jobLog)
	}
	return
}
