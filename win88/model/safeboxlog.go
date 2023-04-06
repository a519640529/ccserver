package model

import (
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

var (
	SafeBoxDBName   = "log"
	SafeBoxCollName = "log_safeboxrec"
)

const (
	SafeBoxLogType_Save int32 = iota
	SafeBoxLogType_TakeOut
	SafeBoxLogType_Exchange
	SafeBoxLogType_Op
)

type SafeBoxRec struct {
	Id            bson.ObjectId `bson:"_id"`
	Platform      string        //平台id
	Channel       string        //平台
	Promoter      string        //推广员
	UserId        int32         //用户id
	Count         int64         //存取金额
	BeforeSafeBox int64         //操作前保险箱金额
	AfterSafeBox  int64         //操作后保险箱金额
	BeforeCoin    int64         //操作前钱包金币
	AfterCoin     int64         //操作后钱包金币
	LogType       int32         //操作类型
	IP            string        //ip
	Oper          string        //操作者
	Time          time.Time     //操作时间
}

func InsertSafeBox(userid int32, opcount, beforesafebox, aftersafebox, beforecoin, aftercoin int64, logtype int32,
	ts time.Time, ip, oper, platform, channel, promoter string) (error, bson.ObjectId) {
	if rpcCli == nil {
		return ErrRPClientNoConn, bson.ObjectId("")
	}

	gr := &SafeBoxRec{}
	gr.Id = bson.NewObjectId()
	gr.UserId = userid
	gr.Time = ts
	gr.IP = ip
	gr.Oper = oper
	gr.Count = opcount
	gr.BeforeSafeBox = beforesafebox
	gr.AfterSafeBox = aftersafebox
	gr.BeforeCoin = beforecoin
	gr.AfterCoin = aftercoin
	gr.LogType = logtype
	gr.Platform = platform
	gr.Channel = channel
	gr.Promoter = promoter
	var ret bool
	err := rpcCli.CallWithTimeout("SafeBoxRecSvc.InsertSafeBox", gr, &ret, time.Second*30)
	if err != nil {
		return err, bson.ObjectId("")
	}
	return nil, gr.Id
}

type GetSafeBoxsArgs struct {
	Plt  string
	SnId int32
}

func GetSafeBoxs(plt string, userid int32) (recs []SafeBoxRec, err error) {
	if rpcCli == nil {
		return nil, ErrRPClientNoConn
	}
	args := &GetSafeBoxsArgs{
		Plt:  plt,
		SnId: userid,
	}

	err = rpcCli.CallWithTimeout("SafeBoxRecSvc.GetSafeBoxs", args, &recs, time.Second*30)
	if err != nil {
		return nil, err
	}
	return
}

type RemoveSafeBoxsArgs struct {
	Plt string
	Ts  time.Time
}

func RemoveSafeBoxs(plt string, ts time.Time) (info *mgo.ChangeInfo, err error) {
	if rpcCli == nil {
		return nil, ErrRPClientNoConn
	}
	args := &RemoveSafeBoxsArgs{
		Plt: plt,
		Ts:  ts,
	}

	err = rpcCli.CallWithTimeout("SafeBoxRecSvc.RemoveSafeBoxs", args, &info, time.Second*30)
	return
}

type RemoveSafeBoxCoinLogArgs struct {
	Plt string
	Id  bson.ObjectId
}

func RemoveSafeBoxCoinLog(plt string, id bson.ObjectId) (err error) {
	if rpcCli == nil {
		return ErrRPClientNoConn
	}
	args := &RemoveSafeBoxCoinLogArgs{
		Plt: plt,
		Id:  id,
	}

	var ret bool
	err = rpcCli.CallWithTimeout("SafeBoxRecSvc.RemoveSafeBoxCoinLog", args, &ret, time.Second*30)
	return
}

type GetSafeBoxCoinLogArgs struct {
	Plt string
	Ts  time.Time
}

func GetSafeBoxCoinLog(plt string, ts time.Time) (recs []SafeBoxRec, err error) {
	if rpcCli == nil {
		return nil, ErrRPClientNoConn
	}
	args := &GetSafeBoxCoinLogArgs{
		Plt: plt,
		Ts:  ts,
	}

	err = rpcCli.CallWithTimeout("SafeBoxRecSvc.GetSafeBoxCoinLog", args, &recs, time.Second*30)
	return
}
