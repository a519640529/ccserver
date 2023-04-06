package model

import (
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

var (
	MonthlyProfitPoolDBName   = "log"
	MonthlyProfitPoolCollName = "log_monthlyprofitpool"
)

type MonthlyProfitPool struct {
	LogId      bson.ObjectId `bson:"_id"` //记录ID
	ServerId   int32         //服务器id
	GroupId    int32         //组id
	Platform   string        //平台id
	GameFreeid int32         //游戏类型房间号
	Coin       int64         //水池金币
	Time       time.Time     //记录时间
}

func NewMonthlyProfitPool() *MonthlyProfitPool {
	log := &MonthlyProfitPool{LogId: bson.NewObjectId()}
	return log
}

func NewMonthlyProfitPoolEx(serverid, gamefreeid, GroupId int32, platform string, coin int64) *MonthlyProfitPool {
	cl := NewMonthlyProfitPool()
	cl.ServerId = serverid
	cl.GameFreeid = gamefreeid
	cl.GroupId = GroupId
	cl.Coin = coin
	cl.Platform = platform
	cl.Time = time.Now()
	return cl
}

func InsertMonthlyProfitPool(logs ...*MonthlyProfitPool) (err error) {
	if rpcCli == nil {
		return ErrRPClientNoConn
	}
	var ret bool
	err = rpcCli.CallWithTimeout("MonthlyProfitPoolSvc.InsertMonthlyProfitPool", logs, &ret, time.Second*30)
	return
}

func InsertSignleMonthlyProfitPool(log *MonthlyProfitPool) (err error) {
	if rpcCli == nil {
		return ErrRPClientNoConn
	}
	var ret bool
	err = rpcCli.CallWithTimeout("MonthlyProfitPoolSvc.InsertSignleMonthlyProfitPool", log, &ret, time.Second*30)
	return
}

func RemoveMonthlyProfitPool(ts time.Time) (ret *mgo.ChangeInfo, err error) {
	if rpcCli == nil {
		return nil, ErrRPClientNoConn
	}
	err = rpcCli.CallWithTimeout("MonthlyProfitPoolSvc.RemoveMonthlyProfitPool", ts, &ret, time.Second*30)
	return
}
