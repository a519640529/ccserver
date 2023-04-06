package model

import (
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

const (
	MATCHPLAYER_FLAG_ISROB            uint = iota //机器人
	MATCHPLAYER_FLAG_ISFIRST                      //首次参赛
	MATCHPLAYER_FLAG_ISNEWBIE                     //新人
	MATCHPLAYER_FLAG_ISQUIT                       //退赛
	MATCHPLAYER_FLAG_SIGNUP_USEFREE               //免费报名
	MATCHPLAYER_FLAG_SIGNUP_USETICKET             //使用报名券报名
	MATCHPLAYER_FLAG_SIGNUP_USECOIN               //使用金币报名
)

// 比赛各阶段时间消耗
type MatchProcessTimeSpend struct {
	Name   string //赛程名称
	MMType int32  //赛制类型
	Spend  int32  //用时,单位:秒
}

// 比赛参赛人员信息
type MatchPlayer struct {
	SnId      int32 //玩家id
	Flag      int32 //玩家类型，二进制 第1位:1是机器人,0玩家 第2位:1首次参加 0:非首次参加 第3位:1新人 0:老玩家 第4位:1退赛 0未退赛 第5位: 1免费报名 第6位: 1使用入场券报名 第7位: 1使用金币报名
	Spend     int32 //报名消耗
	Gain      int32 //获得奖励
	Rank      int32 //名次
	WaitTime  int32 //从报名到开始比赛的等待时间 单位:秒
	MatchTime int32 //比赛中用时 从开始比赛到比赛结束总的用时 单位:秒
}

// 比赛牌局记录
type MatchGameLog struct {
	GameLogId  string  //牌局id
	Name       string  //赛程名称
	ProcessIdx int32   //赛程索引
	NumOfGame  int32   //第几局
	SpendTime  int32   //花费时间
	SnIds      []int32 //参与玩家
}

// 比赛详情
type MatchLog struct {
	Id         bson.ObjectId            `bson:"_id"`
	Platform   string                   //平台编号
	MatchId    int32                    //比赛编号
	MatchName  string                   //比赛名称
	GameFreeId int32                    //游戏类型
	StartTime  time.Time                //开始时间
	EndTime    time.Time                //结束时间
	Players    []*MatchPlayer           //参赛人员数据
	TimeSpend  []*MatchProcessTimeSpend //赛程用时
	GameLogs   []*MatchGameLog          //牌局记录
}

func (ml *MatchLog) AppendMatchProcess(name string, mmtype int32, spend int32) {
	ml.TimeSpend = append(ml.TimeSpend, &MatchProcessTimeSpend{
		Name:   name,
		MMType: mmtype,
		Spend:  spend,
	})
}

func (ml *MatchLog) AppendGameLog(name string, gamelogId string, processIdx, numOfGame, spendTime int32, snids []int32) {
	ml.GameLogs = append(ml.GameLogs, &MatchGameLog{
		Name:       name,
		GameLogId:  gamelogId,
		ProcessIdx: processIdx,
		NumOfGame:  numOfGame,
		SpendTime:  spendTime,
		SnIds:      snids,
	})
}

var (
	MatchLogDBName   = "log"
	MatchLogCollName = "log_matchlog"
)

func NewMatchLog() *MatchLog {
	return &MatchLog{Id: bson.NewObjectId()}
}

func InsertMatchLogs(logs ...*MatchLog) (err error) {
	if rpcCli == nil {
		return ErrRPClientNoConn
	}
	var ret bool
	return rpcCli.CallWithTimeout("MatchLogSvc.InsertMatchLogs", logs, &ret, time.Second*30)
}

type RemoveMatchLogsArgs struct {
	Plt string
	Ts  time.Time
}

func RemoveMatchLogs(plt string, ts time.Time) (ret *mgo.ChangeInfo, err error) {
	if rpcCli == nil {
		return nil, ErrRPClientNoConn
	}
	args := &RemoveMatchLogsArgs{
		Plt: plt,
		Ts:  ts,
	}
	rpcCli.CallWithTimeout("MatchLogSvc.RemoveMatchLogs", args, &ret, time.Second*30)
	return
}
