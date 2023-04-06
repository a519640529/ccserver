package model

import (
	"time"

	"sync/atomic"

	"github.com/globalsign/mgo/bson"
)

var (
	FishJackpotLogDBName   = "log"
	FishJackpotLogCollName = "log_fishjackpotlog"
)

const FishJackpotLogMaxLimitPerQuery = 10

type FishJackpotLog struct {
	LogId    bson.ObjectId `bson:"_id"`
	SnId     int32         //玩家id
	Platform string        //平台名称
	Channel  string        //渠道名称
	Ts       int64         //时间戳
	Time     time.Time     //时间戳
	InGame   int32         //0：其他 1~N：具体游戏id
	JackType int32         //log类型
	Coin     int64         // 中奖金额
	Name     string        // 名字
	RoomId   int32         //房间id
	SeqNo    int64         //流水号(隶属于进程)
}

func NewFishJackpotLog() *FishJackpotLog {
	log := &FishJackpotLog{LogId: bson.NewObjectId()}
	return log
}

func NewFishJackpotLogEx(snid int32, coin int64, roomid, jackType, inGame int32, platform, channel, name string) *FishJackpotLog {
	cl := NewFishJackpotLog()
	cl.SnId = snid
	cl.Platform = platform
	cl.Channel = channel
	tNow := time.Now()
	cl.Ts = tNow.Unix()
	cl.Time = tNow
	cl.Coin = coin
	cl.InGame = inGame
	cl.JackType = jackType
	cl.Name = name
	cl.RoomId = roomid
	cl.SeqNo = atomic.AddInt64(&COINEX_GLOBAL_SEQ, 1)
	return cl
}

type GetLogArgs struct {
	Platform string
}
type GetLogRet struct {
	Logs []FishJackpotLog
}

func GetFishJackpotLogByPlatform(platform string) (log []FishJackpotLog, err error) {
	if rpcCli == nil {
		return nil, ErrRPClientNoConn
	}
	args := &GetLogArgs{Platform: platform}
	ret := &GetLogRet{}
	err = rpcCli.CallWithTimeout("FishJackLogSvc.GetFishJackpotLogByPlatform", args, ret, time.Second*30)
	return ret.Logs, err
}

type InsertLogArgs struct {
	Log *FishJackpotLog
}
type InsertLogRet struct {
	Tag int
}

func InsertSignleFishJackpotLog(log *FishJackpotLog) (err error) {
	if rpcCli == nil {
		return ErrRPClientNoConn
	}
	args := &InsertLogArgs{Log: log}
	ret := &InsertLogRet{}
	err = rpcCli.CallWithTimeout("FishJackLogSvc.InsertSignleFishJackpotLog", args, ret, time.Second*30)
	return err
}
