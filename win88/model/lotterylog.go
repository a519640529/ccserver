package model

import (
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

var (
	LotteryLogDBName   = "log"
	LotteryLogCollName = "log_lottery"
)

type LotteryLog struct {
	LogId      bson.ObjectId `bson:"_id"`
	Platform   string        //平台名称
	GameFreeId int32
	GameId     int32
	SnId       int32
	NickName   string
	IsRob      bool
	Cards      []int32
	Kind       int32
	Coin       int32 //变化金额
	Pool       int64 //变化前彩金池数量
	RecId      string
	Time       time.Time
}

func NewLotteryLog() *LotteryLog {
	log := &LotteryLog{LogId: bson.NewObjectId()}
	return log
}

func NewLotteryLogEx(platform string, gamefreeid, gameid, snid int32, nick string, isRob bool, cards []int32, kind, coin int32, pool int64, recId string) *LotteryLog {
	cl := NewLotteryLog()
	cl.SnId = snid
	cl.Platform = platform
	cl.NickName = nick
	cl.IsRob = isRob
	cl.GameFreeId = gamefreeid
	cl.GameId = gameid
	cl.Cards = cards
	cl.Kind = kind
	cl.Coin = coin
	cl.RecId = recId
	cl.Pool = pool
	cl.Time = time.Now()
	return cl
}
func InsertLotteryLog(platform string, gamefreeid, gameid, snid int32, nick string, isRob bool, cards []int32, kind, coin int32, pool int64, recId string) error {
	if rpcCli == nil {
		return ErrRPClientNoConn
	}
	var ret bool
	cl := NewLotteryLogEx(platform, gamefreeid, gameid, snid, nick, isRob, cards, kind, coin, pool, recId)
	return rpcCli.CallWithTimeout("LotteryLogSvc.InsertLotteryLog", cl, &ret, time.Second*30)
}

func InsertLotteryLogs(logs ...*LotteryLog) (err error) {
	if rpcCli == nil {
		return ErrRPClientNoConn
	}
	var ret bool
	return rpcCli.CallWithTimeout("LotteryLogSvc.InsertLotteryLogs", logs, &ret, time.Second*30)
}

type RemoveLotteryLogArgs struct {
	Plt string
	Ts  time.Time
}

func RemoveLotteryLog(plt string, ts time.Time) (ret *mgo.ChangeInfo, err error) {
	if rpcCli == nil {
		return nil, ErrRPClientNoConn
	}
	args := &RemoveLotteryLogArgs{
		Plt: plt,
		Ts:  ts,
	}
	err = rpcCli.CallWithTimeout("LotteryLogSvc.RemoveLotteryLog", args, &ret, time.Second*30)
	return
}

type GetLotteryLogBySnidAndLessTsArgs struct {
	Plt string
	Id  int32
	Ts  time.Time
}

func GetLotteryLogBySnidAndLessTs(plt string, id int32, ts time.Time) (ret []LotteryLog, err error) {
	if rpcCli == nil {
		return nil, ErrRPClientNoConn
	}
	args := &GetLotteryLogBySnidAndLessTsArgs{
		Plt: plt,
		Id:  id,
		Ts:  ts,
	}
	err = rpcCli.CallWithTimeout("LotteryLogSvc.GetLotteryLogBySnidAndLessTs", args, &ret, time.Second*30)
	return
}
