package model

import (
	"sync/atomic"
	"time"

	"github.com/globalsign/mgo/bson"
	"github.com/idealeak/goserver/core/logger"
)

var (
	CoinLogDBName        = "log"
	CoinLogCollName      = "log_coinex"
	TopicProbeCoinLogAck = "ack_logcoin"
)

var COINEX_GLOBAL_SEQ = int64(0)

type CoinLog struct {
	LogId        bson.ObjectId `bson:"_id"`
	SnId         int32         //玩家id
	Platform     string        //平台名称
	Channel      string        //渠道名称
	Promoter     string        //推广员
	PackageTag   string        //推广包标识
	Count        int64         //帐变数量
	RestCount    int64         //钱包余额
	SafeBoxCount int64         //保险箱余额
	DiamondCount int64         //钻石余额
	Oper         string        //操作者
	Remark       string        //备注
	Time         time.Time     //时间戳
	InGame       int32         //0：其他 1~N：具体游戏id
	Ver          int32         //数据版本(暂时不用)
	CoinType     int32         //金币类型 0:钱包 1:钻石
	LogType      int32         //log类型
	RoomId       int32         //房间id
	SeqNo        int64         //流水号(隶属于进程)
	Ts           int64         //时间戳
}

func NewCoinLog() *CoinLog {
	log := &CoinLog{LogId: bson.NewObjectId()}
	return log
}

func NewCoinLogEx(snid int32, count, restCount, bankCoin int64, ver, logType, inGame int32, oper, remark, platform, channel, promoter string, cointype int, packageid string, roomid int32) *CoinLog {
	cl := NewCoinLog()
	cl.SnId = snid
	cl.Platform = platform
	cl.Channel = channel
	cl.Promoter = promoter
	cl.Count = count
	cl.RestCount = restCount
	cl.SafeBoxCount = bankCoin
	cl.Oper = oper
	cl.Remark = remark
	tNow := time.Now()
	cl.Time = tNow
	cl.InGame = inGame
	cl.Ver = ver
	cl.CoinType = int32(cointype)
	cl.LogType = logType
	cl.PackageTag = packageid
	cl.RoomId = roomid
	cl.SeqNo = atomic.AddInt64(&COINEX_GLOBAL_SEQ, 1)
	cl.Ts = tNow.Unix()
	return cl
}
func NewCoinLogDiamondEx(snid int32, count, restCount, bankCoin, diamondCount int64, ver, logType, inGame int32, oper, remark,
	platform, channel, promoter string, cointype int32, packageid string, roomid int32) *CoinLog {
	cl := NewCoinLog()
	cl.SnId = snid
	cl.Platform = platform
	cl.Channel = channel
	cl.Promoter = promoter
	cl.Count = count
	cl.RestCount = restCount
	cl.SafeBoxCount = bankCoin
	cl.DiamondCount = diamondCount
	cl.Oper = oper
	cl.Remark = remark
	tNow := time.Now()
	cl.Time = tNow
	cl.InGame = inGame
	cl.Ver = ver
	cl.CoinType = cointype
	cl.LogType = logType
	cl.PackageTag = packageid
	cl.RoomId = roomid
	cl.SeqNo = atomic.AddInt64(&COINEX_GLOBAL_SEQ, 1)
	cl.Ts = tNow.Unix()
	return cl
}

type GetCoinLogBySnidAndLessTsArg struct {
	Plt  string
	SnId int32
	Ts   int64
}

func GetCoinLogBySnidAndLessTs(plt string, id int32, ts int64) (ret []CoinLog, err error) {
	if rpcCli == nil {
		logger.Logger.Error("model.GetCoinLogBySnidAndLessTs rpcCli == nil")
		return
	}
	args := &GetCoinLogBySnidAndLessTsArg{
		Plt:  plt,
		SnId: id,
		Ts:   ts,
	}
	err = rpcCli.CallWithTimeout("CoinLogSvc.GetCoinLogBySnidAndLessTs", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Warn("GetCoinLogBySnidAndLessTs error:", err)
	}
	return
}

type GetCoinLogBySnidAndTypeAndInRangeTsLimitByRangeArg struct {
	Plt     string
	SnId    int32
	LogType int
	FromIdx int
	ToIdx   int
	StartTs int64
	EndTs   int64
}
type GetCoinLogBySnidAndTypeAndInRangeTsLimitByRangeRet struct {
	Logs  []CoinLog
	Count int
}

func GetCoinLogBySnidAndTypeAndInRangeTsLimitByRange(plt string, id int32, logType int, startts, endts int64, fromIndex, toIndex int) (logs []CoinLog, count int, err error) {
	if rpcCli == nil {
		logger.Logger.Error("model.GetCoinLogBySnidAndInGameAndGreaterTs rpcCli == nil")
		return
	}

	args := &GetCoinLogBySnidAndTypeAndInRangeTsLimitByRangeArg{
		Plt:     plt,
		SnId:    id,
		LogType: logType,
		FromIdx: fromIndex,
		ToIdx:   toIndex,
		StartTs: startts,
		EndTs:   endts,
	}
	var ret GetCoinLogBySnidAndTypeAndInRangeTsLimitByRangeRet
	err = rpcCli.CallWithTimeout("CoinLogSvc.GetCoinLogBySnidAndTypeAndInRangeTsLimitByRange", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Warn("GetCoinLogBySnidAndTypeAndInRangeTsLimitByRange error:", err)
		return
	}

	logs = ret.Logs
	count = ret.Count
	return
}

func InsertCoinLog(log *CoinLog) (err error) {
	if rpcCli == nil {
		logger.Logger.Error("model.InsertCoinLog rpcCli == nil")
		return
	}

	var ret bool
	err = rpcCli.CallWithTimeout("CoinLogSvc.InsertCoinLog", log, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Warn("InsertCoinLog error:", err)
	}
	return
}

type RemoveCoinLogOneArg struct {
	Plt string
	Id  bson.ObjectId
}

func RemoveCoinLogOne(plt string, id bson.ObjectId) error {
	if rpcCli == nil {
		logger.Logger.Error("model.RemoveCoinLogOne rpcCli == nil")
		return nil
	}

	args := &RemoveCoinLogOneArg{
		Plt: plt,
		Id:  id,
	}
	var ret bool
	err := rpcCli.CallWithTimeout("CoinLogSvc.RemoveCoinLogOne", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Warn("RemoveCoinLogOne error:", err)
	}
	return err
}

type UpdateCoinLogRemarkArg struct {
	Plt    string
	Id     bson.ObjectId
	Remark string
}

func UpdateCoinLogRemark(plt string, id bson.ObjectId, remark string) error {
	if rpcCli == nil {
		logger.Logger.Error("model.UpdateCoinLogRemark rpcCli == nil")
		return nil
	}

	args := &UpdateCoinLogRemarkArg{
		Plt:    plt,
		Id:     id,
		Remark: remark,
	}
	var ret bool
	err := rpcCli.CallWithTimeout("CoinLogSvc.UpdateCoinLogRemark", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Warn("UpdateCoinLogRemark error:", err)
	}
	return err
}
