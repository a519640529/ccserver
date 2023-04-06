package model

import (
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/idealeak/goserver/core/logger"
)

var (
	APILogDBName   = "log"
	APILogCollName = "log_api"
)

const (
	APILogType_Login int32 = iota
	APILogType_Rehold
	APILogType_Logout
)

type APILog struct {
	LogId       bson.ObjectId `bson:"_id"`
	Path        string
	RawQuery    string
	Body        string
	Ip          string
	Result      string
	ProcessTime int64
	GetTime     int64
	Time        time.Time
}

func NewAPILog(path, rawQuery, body, ip, result string, getTime int64, processTime int64) *APILog {
	cl := &APILog{LogId: bson.NewObjectId()}
	cl.Path = path
	cl.RawQuery = rawQuery
	cl.Body = body
	cl.Ip = ip
	cl.GetTime = getTime
	cl.Result = result
	cl.ProcessTime = processTime
	cl.Time = time.Now()
	return cl
}

func InsertAPILog(log *APILog) (err error) {
	if rpcCli == nil {
		logger.Logger.Error("model.InsertAPILog rpcCli == nil")
		return
	}

	var ret bool
	err = rpcCli.CallWithTimeout("APILogSvc.InsertAPILog", log, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Warn("InsertAPILog error:", err)
	}
	return
}

func RemoveAPILog(ts int64) (chged *mgo.ChangeInfo, err error) {
	if rpcCli == nil {
		logger.Logger.Error("model.RemoveAPILog rpcCli == nil")
		return
	}

	err = rpcCli.CallWithTimeout("APILogSvc.RemoveAPILog", ts, &chged, time.Second*30)
	if err != nil {
		logger.Logger.Warn("RemoveAPILog error:", err)
	}
	return
}
