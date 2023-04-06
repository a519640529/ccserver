package model

import (
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/idealeak/goserver/core/logger"
)

var (
	MonitorDBName     = "monitor"
	MonitorPrefixName = "m"
)

type MonitorData struct {
	LogId   bson.ObjectId `bson:"_id"`
	SrvId   int32         //服务器id
	SrvType int32         //服务器类型
	Key     string        //自定义key
	Time    time.Time     //时间戳
	Data    interface{}   //数据体
}

func NewMonitorData(srvid, srvtype int32, key string, data interface{}) *MonitorData {
	log := &MonitorData{
		LogId:   bson.NewObjectId(),
		SrvId:   srvid,
		SrvType: srvtype,
		Key:     key,
		Data:    data,
		Time:    time.Now(),
	}
	return log
}

type MonitorDataArg struct {
	Name string
	Log  *MonitorData
}

func InsertMonitorData(name string, log *MonitorData) (err error) {
	if rpcCli == nil {
		logger.Logger.Error("model.InsertMonitorData rpcCli == nil")
		return
	}
	args := &MonitorDataArg{
		Name: name,
		Log:  log,
	}
	var ret bool
	err = rpcCli.CallWithTimeout("MonitorDataSvc.InsertMonitorData", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Warn("InsertMonitorData error:", err)
	}
	return
}

func UpsertMonitorData(name string, log *MonitorData) (err error) {
	if rpcCli == nil {
		logger.Logger.Error("model.UpsertMonitorData rpcCli == nil")
		return
	}
	args := &MonitorDataArg{
		Name: name,
		Log:  log,
	}
	var ret bool
	err = rpcCli.CallWithTimeout("MonitorDataSvc.UpsertMonitorData", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Warn("UpsertMonitorData error:", err)
	}
	return
}

func RemoveMonitorData(t time.Time) (chged []*mgo.ChangeInfo, err error) {
	if rpcCli == nil {
		logger.Logger.Error("model.RemoveMonitorData rpcCli == nil")
		return
	}
	err = rpcCli.CallWithTimeout("MonitorDataSvc.RemoveMonitorData", t, &chged, time.Second*30)
	if err != nil {
		logger.Logger.Warn("RemoveMonitorData error:", err)
	}
	return
}

type PlayerOLStats struct {
	PlatformStats map[string]*PlayerStats
	RobotStats    PlayerStats
}

type PlayerStats struct {
	InGameCnt map[int32]map[int32]int32
	InHallCnt int32
}

type APITransactStats struct {
	RunTimes        int64 //执行次数
	TotalRuningTime int64 //总执行时间
	MaxRuningTime   int64 //最长执行时间
	TotalBodyLen    int64 //返回体总长度
	MaxBodyLen      int   //最长返回体长度
	SuccessTimes    int64 //执行成功次数
	FailedTimes     int64 //执行失败次数
}
