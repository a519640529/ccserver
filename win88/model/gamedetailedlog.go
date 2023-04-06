package model

import (
	"errors"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/idealeak/goserver/core/logger"
)

var (
	GameDetailedLogDBName   = "log"
	GameDetailedLogCollName = "log_gamedetailed"
)

// 水池上下文
type CoinPoolCtx struct {
	LowerLimit int32 //水池下限
	UpperLimit int32 //水池上限
	CurrValue  int32 //当前水位
	CurrMode   int32 //当前水位模式
	Controlled bool  //被水池控制了
}

type GameDetailedLogRet struct {
	Gplt GameDetailedLogType
}
type GameDetailedLogType struct {
	PageNo   int                //当前页码
	PageSize int                //每页数量
	PageSum  int                //总页码
	Data     []*GameDetailedLog //当页数据
}

type GameDetailedLog struct {
	Id               bson.ObjectId `bson:"_id"` //记录ID
	LogId            string        //记录ID
	GameId           int32         //游戏id
	ClubId           int32         //俱乐部Id
	ClubRoom         string        //俱乐部包间
	Platform         string        //平台id
	Channel          string        //渠道
	Promoter         string        //推广员
	MatchId          int32         //比赛ID
	SceneId          int32         //场景ID
	GameMode         int32         //游戏类型
	GameFreeid       int32         //游戏类型房间号
	PlayerCount      int32         //玩家数量
	GameTiming       int32         //本局游戏用时(mm)
	GameBaseBet      int32         //游戏单位低分
	GameDetailedNote string        //游戏详情
	GameDetailVer    int32         //游戏详情版本
	CpCtx            CoinPoolCtx   //水池上下文信息
	Time             time.Time     //记录时间
	Trend20Lately    string        //最近游戏走势
	Ts               int64         //时间戳
}

func NewGameDetailedLog() *GameDetailedLog {
	log := &GameDetailedLog{Id: bson.NewObjectId()}
	return log
}

func NewGameDetailedLogEx(logid string, gameid, sceneid, gamemode, gamefreeid, playercount, gametiming, gamebasebet int32,
	gamedetailednote string, platform string, clubId int32, clubRoom string, cpCtx CoinPoolCtx, ver int32, trend20Lately string) *GameDetailedLog {
	cl := NewGameDetailedLog()
	cl.LogId = logid
	cl.GameId = gameid
	cl.ClubId = clubId
	cl.ClubRoom = clubRoom
	cl.SceneId = sceneid
	cl.GameMode = gamemode
	cl.GameFreeid = gamefreeid
	cl.PlayerCount = playercount
	cl.GameTiming = gametiming
	cl.GameBaseBet = gamebasebet
	cl.GameDetailedNote = gamedetailednote
	cl.Platform = platform
	tNow := time.Now()
	cl.Time = tNow
	cl.CpCtx = cpCtx
	cl.GameDetailVer = ver
	cl.Trend20Lately = trend20Lately
	cl.Ts = time.Now().Unix()
	return cl
}

func InsertGameDetailedLog(log *GameDetailedLog) (err error) {
	if rpcCli == nil {
		logger.Logger.Error("model.InsertGameDetailedLog rpcCli == nil")
		return
	}
	var ret bool
	err = rpcCli.CallWithTimeout("GameDetailedSvc.InsertGameDetailedLog", log, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Warn("model.InsertGameDetailedLog error:", err)
		return
	}
	return
}

type RemoveGameDetailedLogArg struct {
	Plt string
	Ts  time.Time
}

func RemoveGameDetailedLog(plt string, ts time.Time) (*mgo.ChangeInfo, error) {
	if rpcCli == nil {
		logger.Logger.Error("model.RemoveGameDetailedLog rpcCli == nil")
		return nil, errors.New("rpcCli is nil")
	}
	args := &RemoveGameDetailedLogArg{
		Plt: plt,
		Ts:  ts,
	}
	var ret mgo.ChangeInfo
	err := rpcCli.CallWithTimeout("GameDetailedSvc.RemoveGameDetailedLog", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Warn("model.RemoveGameDetailedLog error:", err)
		return nil, err
	}
	return &ret, nil
}

type GameDetailedArg struct {
	Plt                string
	StartTime, EndTime int64
}

func GetAllGameDetailedLogsByTs(plt string, startTime, endTime int64) (data []GameDetailedLog) {
	if rpcCli == nil {
		logger.Logger.Error("model.GetAllGameDetailedLogsByTs rpcCli == nil")
		return nil
	}
	args := &GameDetailedArg{
		Plt:       plt,
		StartTime: startTime,
		EndTime:   endTime,
	}

	err := rpcCli.CallWithTimeout("GameDetailedSvc.GetAllGameDetailedLogsByTs", args, &data, time.Second*30)
	if err != nil {
		logger.Logger.Error("model.GetAllGameDetailedLogsByTs is error", err)
		return nil
	}
	return
}

type GetPlayerHistoryArg struct {
	Plt   string
	LogId string
}

func GetPlayerHistory(plt, logid string) *GameDetailedLog {
	if rpcCli == nil {
		logger.Logger.Error("model.GetPlayerHistory rpcCli == nil")
		return nil
	}
	args := &GetPlayerHistoryArg{
		Plt:   plt,
		LogId: logid,
	}
	var ret GameDetailedLog
	err := rpcCli.CallWithTimeout("GameDetailedSvc.GetPlayerHistory", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Error("model.GetPlayerHistory is error", err)
		return nil
	}
	return &ret
}

type GameDetailedGameIdAndArg struct {
	Plt      string
	Gameid   int
	LimitNum int
}

func GetAllGameDetailedLogsByGameIdAndTs(plt string, gameid int, limitnum int) (data []GameDetailedLog) {
	if rpcCli == nil {
		logger.Logger.Error("model.GetAllGameDetailedLogsByGameIdAndTs rpcCli == nil")
		return nil
	}
	args := &GameDetailedGameIdAndArg{
		Plt:      plt,
		Gameid:   gameid,
		LimitNum: limitnum,
	}

	err := rpcCli.CallWithTimeout("GameDetailedSvc.GetAllGameDetailedLogsByGameIdAndTs", args, &data, time.Second*30)
	if err != nil {
		logger.Logger.Error("model.GetAllGameDetailedLogsByGameIdAndTs is error", err)
		return nil
	}
	return
}

type GetPlayerHistoryAPIArg struct {
	//SnId               int32
	Platform           string
	StartTime, EndTime int64
	PageNo, PageSize   int
}

func GetPlayerHistoryAPI(snId int32, platform string, startTime, endTime int64, pageno, pagesize int) (gdt GameDetailedLogType) {
	if rpcCli == nil {
		logger.Logger.Error("model.GetPlayerHistoryAPI rpcCli == nil")
		return
	}
	args := &GetPlayerHistoryAPIArg{
		//SnId:      snId,
		Platform:  platform,
		StartTime: startTime,
		EndTime:   endTime,
		PageNo:    pageno,
		PageSize:  pagesize,
	}
	ret := &GameDetailedLogRet{}
	err := rpcCli.CallWithTimeout("GameDetailedSvc.GetPlayerHistoryAPI", args, ret, time.Second*30)
	if err != nil {
		logger.Logger.Error("model.GetPlayerHistoryAPI is error", err)
		return
	}
	return ret.Gplt
}
