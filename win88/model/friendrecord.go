package model

import (
	"github.com/globalsign/mgo/bson"
	"github.com/idealeak/goserver/core/logger"
	"time"
)

var (
	FriendRecordLogDBName   = "log"
	FriendRecordLogCollName = "log_friendrecordlog"
)

// 好友聊天 战绩
type FriendRecord struct {
	LogId     bson.ObjectId `bson:"_id"` //记录ID
	Platform  string
	SnId      int32 //玩家id
	IsWin     int32 //1:赢 0:平 -1:输
	GameId    int32 //游戏场次
	BaseScore int32 //底分
	Ts        int64
}

func NewFriendRecordLog() *FriendRecord {
	return &FriendRecord{LogId: bson.NewObjectId()}
}
func NewFriendRecordLogEx(platform string, snid int32, isWin, gameId, baseScore int32) *FriendRecord {
	fri := NewFriendRecordLog()
	fri.Platform = platform
	fri.SnId = snid
	fri.IsWin = isWin
	fri.GameId = gameId
	fri.BaseScore = baseScore
	fri.Ts = time.Now().Unix()
	return fri
}

type FriendRecordSnidArg struct {
	Platform     string
	SnId, GameId int32
	Size         int
}

type FriendRecordSnidRet struct {
	FR []*FriendRecord
}

func GetFriendRecordLogBySnid(platform string, snid, gameid int32, size int) []*FriendRecord {
	if rpcCli == nil {
		logger.Logger.Error("model.GetFriendRecordLogBySnid rpcCli == nil")
		return nil
	}
	args := &FriendRecordSnidArg{
		SnId:     snid,
		Platform: platform,
		GameId:   gameid,
		Size:     size,
	}
	ret := &FriendRecordSnidRet{}
	err := rpcCli.CallWithTimeout("FriendRecordLogSvc.GetFriendRecordLogBySnid", args, ret, time.Second*30)
	if err != nil {
		logger.Logger.Error("model.GetFriendRecordLogBySnid is error", err)
		return nil
	}
	return ret.FR
}
