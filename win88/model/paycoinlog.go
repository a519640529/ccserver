package model

import (
	"time"

	"github.com/globalsign/mgo/bson"
)

var (
	PayCoinLogDBName   = "user"
	PayCoinLogCollName = "user_coinlog"
)

const (
	PayCoinLogType_Coin        int32 = iota //金币对账日志
	PayCoinLogType_SafeBoxCoin              //保险箱对账日志
	PayCoinLogType_Club                     //俱乐部对账日志
	PayCoinLogType_Ticket                   //比赛入场券
)

type PayCoinLog struct {
	LogId     bson.ObjectId `bson:"_id"`
	BillNo    int64         //账单ID
	SnId      int32         //用户ID
	Coin      int64         //账单金币额
	CoinEx    int64         //额外赠送金币
	Inside    int32         //
	TimeStamp int64         //生效时间戳
	Time      time.Time     //账单时间
	Ver       int32         //账单版本号
	Oper      string        //操作人
	LogType   int32         //日志类型（金币或者保险箱）
	Status    int32         //状态 0 默认  1超时
}

func NewPayCoinLog(billNo int64, snid int32, coin int64, inside int32, oper string, logType int32, coinEx int64) *PayCoinLog {
	tNow := time.Now()
	log := &PayCoinLog{
		LogId:     bson.NewObjectId(),
		BillNo:    billNo,
		SnId:      snid,
		Coin:      coin,
		CoinEx:    coinEx,
		Inside:    inside,
		Ver:       VER_PLAYER_MAX - 1,
		TimeStamp: tNow.UnixNano(),
		Time:      tNow,
		Oper:      oper,
		LogType:   logType,
		Status:    0,
	}
	return log
}

type InsertPayCoinLogArgs struct {
	Plt  string
	Logs []*PayCoinLog
}

func InsertPayCoinLogs(plt string, logs ...*PayCoinLog) (err error) {
	if rpcCli == nil {
		return ErrRPClientNoConn
	}

	args := &InsertPayCoinLogArgs{Plt: plt, Logs: logs}
	var ret bool
	err = rpcCli.CallWithTimeout("PayCoinLogSvc.InsertPayCoinLogs", args, &ret, time.Second*30)
	if err != nil {
		return err
	}
	return nil
}

func InsertPayCoinLog(plt string, log *PayCoinLog) error {
	return InsertPayCoinLogs(plt, log)
}

type RemovePayCoinLogArgs struct {
	Plt string
	Id  bson.ObjectId
}

func RemovePayCoinLog(plt string, id bson.ObjectId) (err error) {
	if rpcCli == nil {
		return ErrRPClientNoConn
	}

	args := &RemovePayCoinLogArgs{Plt: plt, Id: id}
	var ret bool
	err = rpcCli.CallWithTimeout("PayCoinLogSvc.RemovePayCoinLog", args, &ret, time.Second*30)
	if err != nil {
		return err
	}
	return nil
}

type UpdatePayCoinLogStatusArgs struct {
	Plt    string
	Id     bson.ObjectId
	Status int32
}

func UpdatePayCoinLogStatus(plt string, id bson.ObjectId, status int32) (err error) {
	if rpcCli == nil {
		return ErrRPClientNoConn
	}

	args := &UpdatePayCoinLogStatusArgs{Plt: plt, Id: id, Status: status}
	var ret bool
	err = rpcCli.CallWithTimeout("PayCoinLogSvc.UpdatePayCoinLogStatus", args, &ret, time.Second*30)
	if err != nil {
		return err
	}
	return nil
}

type GetPayCoinLogArgs struct {
	Plt  string
	SnId int32
	Cond int64
}

func GetAllPayCoinLog(plt string, snid int32, ts int64) (ret []PayCoinLog, err error) {
	if rpcCli == nil {
		return nil, ErrRPClientNoConn
	}

	args := &GetPayCoinLogArgs{Plt: plt, SnId: snid, Cond: ts}
	err = rpcCli.CallWithTimeout("PayCoinLogSvc.GetAllPayCoinLog", args, &ret, time.Second*30)
	if err != nil {
		return
	}

	return
}

func GetAllPaySafeBoxCoinLog(plt string, snid int32, ts int64) (ret []PayCoinLog, err error) {
	if rpcCli == nil {
		return nil, ErrRPClientNoConn
	}

	args := &GetPayCoinLogArgs{Plt: plt, SnId: snid, Cond: ts}
	err = rpcCli.CallWithTimeout("PayCoinLogSvc.GetAllPaySafeBoxCoinLog", args, &ret, time.Second*30)
	if err != nil {
		return
	}
	return
}

func GetAllPayClubCoinLog(plt string, snid int32, ts int64) (ret []PayCoinLog, err error) {
	if rpcCli == nil {
		return nil, ErrRPClientNoConn
	}

	args := &GetPayCoinLogArgs{Plt: plt, SnId: snid, Cond: ts}
	err = rpcCli.CallWithTimeout("PayCoinLogSvc.GetAllPayClubCoinLog", args, &ret, time.Second*30)
	if err != nil {
		return
	}
	return
}

func GetAllPayTicketCoinLog(plt string, snid int32, ts int64) (ret []PayCoinLog, err error) {
	if rpcCli == nil {
		return nil, ErrRPClientNoConn
	}

	args := &GetPayCoinLogArgs{Plt: plt, SnId: snid, Cond: ts}
	err = rpcCli.CallWithTimeout("PayCoinLogSvc.GetAllPayTicketCoinLog", args, &ret, time.Second*30)
	if err != nil {
		return
	}
	return
}

type GetPayCoinLogByBillNoArgs struct {
	Plt    string
	SnId   int32
	BillNo int64
}

func GetPayCoinLogByBillNo(plt string, snid int32, billNo int64) (log *PayCoinLog, err error) {
	if rpcCli == nil {
		return nil, ErrRPClientNoConn
	}

	args := &GetPayCoinLogArgs{Plt: plt, SnId: snid, Cond: billNo}
	err = rpcCli.CallWithTimeout("PayCoinLogSvc.GetPayCoinLogByBillNo", args, &log, time.Second*30)
	if err != nil {
		return
	}

	return
}
