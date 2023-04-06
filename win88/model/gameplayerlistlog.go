package model

import (
	"errors"
	"time"

	"github.com/globalsign/mgo/bson"
	"github.com/idealeak/goserver/core/logger"
)

var (
	GamePlayerListLogDBName   = "log"
	GamePlayerListLogCollName = "log_gameplayerlistlog"
)

type GameTotalRecord struct {
	GameTotal     int32 //今日牌局数
	GameTotalCoin int32 //今日游戏流水
	GameWinTotal  int32 //今日游戏输赢总额
}

type GamePlayerListLog struct {
	LogId             bson.ObjectId `bson:"_id"` //记录ID
	SnId              int32         //用户Id
	Name              string        //名称
	GameId            int32         //游戏id
	BaseScore         int32         //游戏底注
	ClubId            int32         //俱乐部Id
	ClubRoom          string        //俱乐部包间
	TaxCoin           int64         //税收
	ClubPumpCoin      int64         //俱乐部额外抽水
	Platform          string        //平台id
	Channel           string        //渠道
	Promoter          string        //推广员
	PackageTag        string        //包标识
	SceneId           int32         //场景ID
	GameMode          int32         //游戏类型
	GameFreeid        int32         //游戏类型房间号
	GameDetailedLogId string        //游戏记录Id
	IsFirstGame       bool          //是否第一次游戏
	//对于拉霸类：BetAmount=100	WinAmountNoAnyTax=0	（表示投入多少、收益多少，值>=0）
	//拉霸类小游戏会是：BetAmount=0		WinAmountNoAnyTax=100	（投入0、收益多少，值>=0）
	//对战场：BetAmount=0		WinAmountNoAnyTax=100	（投入会有是0、收益有正负，WinAmountNoAnyTax=100则盈利，WinAmountNoAnyTax=-100则输100）
	BetAmount         int64     //下注金额
	WinAmountNoAnyTax int64     //盈利金额，不包含任何税
	TotalIn           int64     //本局投入
	TotalOut          int64     //本局产出
	Time              time.Time //记录时间
	RoomType          int32     //房间类型
	GameDif           string    //游戏标识
	GameClass         int32     //游戏类型 1棋牌	2电子	3百人	4捕鱼	5视讯	6彩票	7体育
	MatchId           int32
	Ts                int32
}

func NewGamePlayerListLog() *GamePlayerListLog {
	log := &GamePlayerListLog{LogId: bson.NewObjectId()}
	return log
}
func NewGamePlayerListLogEx(snid int32, gamedetailedlogid string, platform, channel, promoter, packageTag string, gameid, baseScore,
	sceneid, gamemode, gamefreeid int32, totalin, totalout int64, clubId int32, clubRoom string, taxCoin, pumpCoin int64, roomType int32,
	betAmount, winAmountNoAnyTax int64, key, name string, gameClass int32, isFirst bool, matchid int32) *GamePlayerListLog {
	cl := NewGamePlayerListLog()
	cl.SnId = snid
	cl.GameDetailedLogId = gamedetailedlogid
	cl.Platform = platform
	cl.Name = name
	cl.Channel = channel
	cl.Promoter = promoter
	cl.PackageTag = packageTag
	cl.GameFreeid = gamefreeid
	cl.GameId = gameid
	cl.BaseScore = baseScore
	cl.ClubId = clubId
	cl.GameMode = gamemode
	cl.SceneId = sceneid
	cl.TotalIn = totalin
	cl.TotalOut = totalout
	cl.ClubRoom = clubRoom
	cl.TaxCoin = taxCoin
	cl.IsFirstGame = isFirst
	cl.ClubPumpCoin = pumpCoin
	cl.RoomType = roomType
	cl.BetAmount = betAmount
	cl.WinAmountNoAnyTax = winAmountNoAnyTax
	cl.GameDif = key
	cl.GameClass = gameClass
	tNow := time.Now()
	cl.Ts = int32(tNow.Unix())
	cl.Time = tNow
	cl.MatchId = matchid
	return cl
}

type GamePlayerListRet struct {
	Gtr  *GameTotalRecord
	Gplt GamePlayerListType
}

func InsertGamePlayerListLog(log *GamePlayerListLog) error {
	if rpcCli == nil {
		logger.Logger.Error("model.InsertGamePlayerListLog rpcCli == nil")
		return errors.New("rpcCli == nil")
	}
	var ret bool
	err := rpcCli.CallWithTimeout("GamePlayerListSvc.InsertGamePlayerListLog", log, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Error("model.InsertGamePlayerListLog is error", err)
		return err
	}
	return nil
}

type NeedGameRecord struct {
	LogId             bson.ObjectId `bson:"_id"` //记录ID
	SnId              int32         //用户Id
	Username          string        //三方用户名
	ClubId            int32         //俱乐部Id
	ClubRoom          string        //俱乐部包间
	GameFreeid        int32         //游戏类型
	TaxCoin           int64         //税收
	ClubPumpCoin      int64         //俱乐部额外抽水
	TotalIn           int64         //本局投入
	TotalOut          int64         //本局产出
	BetAmount         int64         //投注金额（仅在个人中心表现使用，有原始数据加工得到）
	WinAmountNoAnyTax int64         //盈利金额，不包含任何税（仅在个人中心表现使用，有原始数据加工得到）
	SceneId           int32         //场景ID
	Ts                int32         //记录时间
	RoomType          int32
	GameDetailedLogId string // 游戏记录id
	ThirdOrderId      string // 三方游戏记录id
}

type GamePlayerListType struct {
	PageNo   int               //当前页码
	PageSize int               //每页数量
	PageSum  int               //总页码
	Data     []*NeedGameRecord //当页数据
	*GameTotalRecord
}
type TotalCoin struct {
	TotalIn      int64 //本局投入
	TotalOut     int64 //本局产出
	TaxCoin      int64 //税收
	ClubPumpCoin int64 //俱乐部额外抽水
}
type GamePlayerListArg struct {
	SnId                     int32
	Platform                 string
	StartTime, EndTime       int64
	ClubId                   int32
	PageNo, PageSize, GameId int
	RoomType, GameClass      int32
}
type GamePlayerListAPIArg struct {
	SnId               int32
	Platform           string
	StartTime, EndTime int64
	PageNo, PageSize   int
}

type GamePlayerExistListArg struct {
	SnId     int32
	Platform string
	DayNum   int
}

func GetPlayerCount(SnId int32, platform string, startTime, endTime int64, clubId int32) *GameTotalRecord {
	if rpcCli == nil {
		logger.Logger.Error("model.GetPlayerCount rpcCli == nil")
		return nil
	}
	if clubId == 0 {
		return nil
	}
	args := &GamePlayerListArg{
		SnId:      SnId,
		Platform:  platform,
		StartTime: startTime,
		EndTime:   endTime,
		ClubId:    clubId,
	}
	ret := &GamePlayerListRet{}
	err := rpcCli.CallWithTimeout("GamePlayerListSvc.GetPlayerCount", args, ret, time.Second*30)
	if err != nil {
		logger.Logger.Error("model.GetPlayerCount is error", err)
		return nil
	}
	return ret.Gtr
}

func GetPlayerListLog(snId int32, platform string, pageNo, pageSize int, startTime, endTime int64, clubId int32, n int) (gpt GamePlayerListType) {
	if rpcCli == nil {
		logger.Logger.Error("model.GetPlayerListLog rpcCli == nil")
		return
	}
	if n == 1 {
		gpt.GameTotalRecord = GetPlayerCount(snId, platform, startTime, endTime, clubId)
	}
	if clubId == 0 {
		return
	}
	args := &GamePlayerListArg{
		SnId:      snId,
		Platform:  platform,
		StartTime: startTime,
		EndTime:   endTime,
		ClubId:    clubId,
		PageNo:    pageNo,
		PageSize:  pageSize,
	}
	ret := &GamePlayerListRet{}
	err := rpcCli.CallWithTimeout("GamePlayerListSvc.GetPlayerListLog", args, ret, time.Second*30)
	if err != nil {
		logger.Logger.Error("model.GetPlayerListLog is error", err)
		return
	}
	return ret.Gplt
}

func GetPlayerListByHall(snId int32, platform string, pageNo, pageSize int, startTime, endTime int64, roomType, gameClass int32) (gdt GamePlayerListType) {
	if rpcCli == nil {
		logger.Logger.Error("model.GetPlayerListByHall rpcCli == nil")
		return
	}
	args := &GamePlayerListArg{
		SnId:      snId,
		Platform:  platform,
		StartTime: startTime,
		EndTime:   endTime,
		PageNo:    pageNo,
		PageSize:  pageSize,
		RoomType:  roomType,
		GameClass: gameClass,
	}
	ret := &GamePlayerListRet{}
	err := rpcCli.CallWithTimeout("GamePlayerListSvc.GetPlayerListByHall", args, ret, time.Second*30)
	if err != nil {
		logger.Logger.Error("model.GetPlayerListByHall is error", err)
		return
	}
	return ret.Gplt
}

func GetPlayerListByHallEx(snId int32, platform string, pageNo, pageSize int, startTime, endTime int64, roomType, gameClass int32, gameid int) (gdt GamePlayerListType) {
	if rpcCli == nil {
		logger.Logger.Error("model.GetPlayerListByHall rpcCli == nil")
		return
	}
	args := &GamePlayerListArg{
		SnId:      snId,
		Platform:  platform,
		StartTime: startTime,
		EndTime:   endTime,
		PageNo:    pageNo,
		PageSize:  pageSize,
		RoomType:  roomType,
		GameClass: gameClass,
		GameId:    gameid,
	}
	ret := &GamePlayerListRet{}
	err := rpcCli.CallWithTimeout("GamePlayerListSvc.GetPlayerListByHallEx", args, ret, time.Second*30)
	if err != nil {
		logger.Logger.Error("model.GetPlayerListByHallEx is error", err)
		return
	}
	return ret.Gplt
}

func GetPlayerListByHallExAPI(snId int32, platform string, startTime, endTime int64, pageno, pagesize int) (gdt GamePlayerListType) {
	if rpcCli == nil {
		logger.Logger.Error("model.GetPlayerListByHallExAPI rpcCli == nil")
		return
	}
	args := &GamePlayerListAPIArg{
		SnId:      snId,
		Platform:  platform,
		StartTime: startTime,
		EndTime:   endTime,
		PageNo:    pageno,
		PageSize:  pagesize,
	}
	ret := &GamePlayerListRet{}
	err := rpcCli.CallWithTimeout("GamePlayerListSvc.GetPlayerListByHallExAPI", args, ret, time.Second*30)
	if err != nil {
		logger.Logger.Error("model.GetPlayerListByHallExAPI is error", err)
		return
	}
	return ret.Gplt
}

func GetPlayerExistListByTs(snId int32, platform string, dayNum int) []int64 {
	if rpcCli == nil {
		logger.Logger.Error("model.GetPlayerExistListByTs rpcCli == nil")
		return nil
	}
	args := &GamePlayerExistListArg{
		SnId:     snId,
		Platform: platform,
		DayNum:   dayNum,
	}
	var ret []int64
	err := rpcCli.CallWithTimeout("GamePlayerListSvc.GetPlayerExistListByTs", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Error("model.GetPlayerExistListByTs is error", err)
		return nil
	}
	return ret
}
