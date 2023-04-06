package model

import (
	"github.com/globalsign/mgo/bson"
	"time"
)

// //////////////////////////////////////////////////俱乐部
var (
	RebateLogDBName   = "log"
	RebateLogCollName = "log_rebatetask"
)

type Rebate struct {
	Id          bson.ObjectId `bson:"_id"` //记录ID
	SnId        int32         //玩家Id
	CodeCoin    int64         //洗码量
	RebateCoin  int64         //洗码金额
	ReceiveType int32         //领取方式
	Ts          int64
}

type InsertRebateLogArgs struct {
	Log *Rebate
	Plt string
}

func InsertRebateLog(plt string, data *Rebate) error {
	if rpcCli == nil {
		return ErrRPClientNoConn
	}

	args := &InsertRebateLogArgs{
		Plt: plt,
		Log: data,
	}
	var ret bool
	return rpcCli.CallWithTimeout("RebateLogSvc.InsertRebateLog", args, &ret, time.Second*30)
}

type RebateLog struct {
	Rebates  []*Rebate
	PageNo   int
	PageSize int
	PageSum  int
}

type GetRebateLogArgs struct {
	Plt      string
	PageNo   int
	PageSize int
	SnId     int32
}

func GetRebateLog(plt string, pageNo, pageSize int, snId int32) (r *RebateLog, err error) {
	if rpcCli == nil {
		return nil, ErrRPClientNoConn
	}

	args := &GetRebateLogArgs{
		Plt:      plt,
		PageNo:   pageNo,
		PageSize: pageSize,
		SnId:     snId,
	}
	r = new(RebateLog)
	err = rpcCli.CallWithTimeout("RebateLogSvc.GetRebateLog", args, &r, time.Second*30)
	return
}
