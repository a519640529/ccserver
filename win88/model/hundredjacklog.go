package model

import (
	"github.com/globalsign/mgo/bson"
	"sync/atomic"
	"time"
)

var (
	HundredjackpotLogDBName   = "log"
	HundredjackpotLogCollName = "log_hundredjackpotlog" // 游戏爆奖
)

// HundredjackpotLog 百人场爆奖结构
type HundredjackpotLog struct {
	LogID        bson.ObjectId `bson:"_id"`
	SnID         int32         // 玩家id
	Platform     string        // 平台名称
	Channel      string        // 渠道名称
	Ts           int64         // 时间戳
	Time         time.Time     // 时间戳
	InGame       int32         // 0：其他 1~N：具体游戏id
	LogType      int32         // log类型 (具体见对应游戏)
	Coin         int64         // 中奖金额
	TurnCoin     int64         // 当前金额
	Name         string        // 名字
	Vip          int32         // vip等级
	RoomID       int32         // 房间id ->gamefreeId
	LikeNum      int32         // 点赞数
	LinkeSnids   string        // 点赞人
	PlayblackNum int32         // 回放次数
	GameData     []string      // 回放数据json
	SeqNo        int64         // 流水号(隶属于进程)
}

// HundredjackpotLogMaxLimitPerQuery 上榜人数
const HundredjackpotLogMaxLimitPerQuery = 99

// NewHundredjackpotLog 实例
func NewHundredjackpotLog() *HundredjackpotLog {
	log := &HundredjackpotLog{LogID: bson.NewObjectId()}
	return log
}

// NewHundredjackpotLogEx 赋值创建
func NewHundredjackpotLogEx(snid int32, coin, turncoin int64, roomid, logType, inGame, vip int32, platform, channel, name string, gamedata []string) *HundredjackpotLog {
	cl := NewHundredjackpotLog()
	cl.SnID = snid
	cl.Platform = platform
	cl.Channel = channel
	tNow := time.Now()
	cl.Ts = tNow.Unix()
	cl.Time = tNow
	cl.Coin = coin
	cl.InGame = inGame
	cl.Vip = vip
	cl.LogType = logType
	cl.Name = name
	cl.RoomID = roomid
	cl.GameData = gamedata
	cl.TurnCoin = turncoin
	cl.SeqNo = atomic.AddInt64(&COINEX_GLOBAL_SEQ, 1)
	return cl
}

// GetHundredjackpotLogTsByPlatformAndGameID 时间排名
type GetHundredjackpotLogArgs struct {
	Plt string
	Id1 int32
	Id2 int32
}

func GetHundredjackpotLogTsByPlatformAndGameID(platform string, gameid int32) (ret []HundredjackpotLog, err error) {
	if rpcCli == nil {
		return nil, ErrRPClientNoConn
	}

	args := &GetHundredjackpotLogArgs{
		Plt: platform,
		Id1: gameid,
	}
	err = rpcCli.CallWithTimeout("HundredjackpotLogSvc.GetHundredjackpotLogTsByPlatformAndGameID", args, &ret, time.Second*30)

	return
}

// GetHundredjackpotLogCoinByPlatformAndGameID 中奖金币排名
func GetHundredjackpotLogCoinByPlatformAndGameID(platform string, gameid int32) (ret []HundredjackpotLog, err error) {
	if rpcCli == nil {
		return nil, ErrRPClientNoConn
	}

	args := &GetHundredjackpotLogArgs{
		Plt: platform,
		Id1: gameid,
	}
	err = rpcCli.CallWithTimeout("HundredjackpotLogSvc.GetHundredjackpotLogCoinByPlatformAndGameID", args, &ret, time.Second*30)
	return
}

// GetHundredjackpotLogTsByPlatformAndGameFreeID 时间排名
func GetHundredjackpotLogTsByPlatformAndGameFreeID(platform string, gamefreeid int32) (ret []HundredjackpotLog, err error) {
	if rpcCli == nil {
		return nil, ErrRPClientNoConn
	}

	args := &GetHundredjackpotLogArgs{
		Plt: platform,
		Id1: gamefreeid,
	}
	err = rpcCli.CallWithTimeout("HundredjackpotLogSvc.GetHundredjackpotLogTsByPlatformAndGameFreeID", args, &ret, time.Second*30)
	return
}

// GetHundredjackpotLogCoinByPlatformAndGameFreeID 中奖金币排名
func GetHundredjackpotLogCoinByPlatformAndGameFreeID(platform string, gamefreeid int32) (ret []HundredjackpotLog, err error) {
	if rpcCli == nil {
		return nil, ErrRPClientNoConn
	}

	args := &GetHundredjackpotLogArgs{
		Plt: platform,
		Id1: gamefreeid,
	}
	err = rpcCli.CallWithTimeout("HundredjackpotLogSvc.GetHundredjackpotLogCoinByPlatformAndGameFreeID", args, &ret, time.Second*30)
	return
}

// GetLastHundredjackpotLogBySnidAndGameID .
func GetLastHundredjackpotLogBySnidAndGameID(plt string, id int32, gamefreeid int32) (log *HundredjackpotLog, err error) {
	if rpcCli == nil {
		return nil, ErrRPClientNoConn
	}
	args := &GetHundredjackpotLogArgs{
		Plt: plt,
		Id1: id,
		Id2: gamefreeid,
	}
	err = rpcCli.CallWithTimeout("HundredjackpotLogSvc.GetLastHundredjackpotLogBySnidAndGameID", args, &log, time.Second*30)
	return
}

// InsertHundredjackpotLog .
func InsertHundredjackpotLog(log *HundredjackpotLog) (err error) {
	if rpcCli == nil {
		return ErrRPClientNoConn
	}
	var ret bool
	err = rpcCli.CallWithTimeout("HundredjackpotLogSvc.InsertHundredjackpotLog", log, &ret, time.Second*30)
	return
}

// UpdateLikeNum 点赞次数
type UpdateLikeNumArgs struct {
	Plt       string
	Id        bson.ObjectId
	LikeNum   int32
	LikeSnIds string
}

func UpdateLikeNum(plt string, id bson.ObjectId, likeNum int32, linkeSnids string) error {
	if rpcCli == nil {
		return ErrRPClientNoConn
	}
	args := &UpdateLikeNumArgs{
		Plt:       plt,
		Id:        id,
		LikeNum:   likeNum,
		LikeSnIds: linkeSnids,
	}
	var ret bool
	return rpcCli.CallWithTimeout("HundredjackpotLogSvc.UpdateLikeNum", args, &ret, time.Second*30)
}

// UpdatePlayBlackNum 回放次数
type UpdatePlayBlackNumArgs struct {
	Plt          string
	Id           bson.ObjectId
	PlayblackNum int32
}

func UpdatePlayBlackNum(plt string, id bson.ObjectId, playblackNum int32) (ret []string, err error) {
	if rpcCli == nil {
		return nil, ErrRPClientNoConn
	}
	args := &UpdatePlayBlackNumArgs{
		Plt:          plt,
		Id:           id,
		PlayblackNum: playblackNum,
	}
	err = rpcCli.CallWithTimeout("HundredjackpotLogSvc.UpdatePlayBlackNum", args, &ret, time.Second*30)
	return
}

// RemoveHundredjackpotLogOne 移除
type RemoveHundredjackpotLogOneArgs struct {
	Plt string
	Id  bson.ObjectId
}

func RemoveHundredjackpotLogOne(plt string, id bson.ObjectId) error {
	if rpcCli == nil {
		return ErrRPClientNoConn
	}
	args := &RemoveHundredjackpotLogOneArgs{
		Plt: plt,
		Id:  id,
	}
	var ret bool
	return rpcCli.CallWithTimeout("HundredjackpotLogSvc.RemoveHundredjackpotLogOne", args, &ret, time.Second*30)
}
