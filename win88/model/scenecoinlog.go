package model

import (
	"time"

	"github.com/globalsign/mgo/bson"
	"github.com/idealeak/goserver/core/logger"
)

// 每一局游戏记录log
var (
	SceneCoinLogDBName   = "log"
	SceneCoinLogCollName = "log_scenecoin"
)

type SceneCoinLog struct {
	LogId            bson.ObjectId `bson:"_id"`
	Platform         string        //平台id
	Channel          string        //渠道
	Promoter         string        //推广员
	PackageTag       string        //推广包标识
	SnId             int32         //玩家id
	ChangeCoin       int64         //变化金额
	OldCoin          int64         //该局游戏前金额
	NewCoin          int64         //当前游戏金币
	TotalBet         int64         //下注额
	BaseRate         int64         //底注
	GameId           int32         //游戏类型
	EventType        int64         //事件类型 输赢
	Ip               string        //IP
	GameIdex         int32         //房间id
	SceneId          int32         //场景ID
	GameMode         int32         //游戏类型
	GameFreeid       int32         //游戏类型房间号
	TaxCoin          int64         //税收
	WinCoin          int64         //赢的总钱数，扣税之前
	JackpotWinCoin   int64         //爆奖金额
	SmallGameWinCoin int64         //小游戏赢的钱
	SeatId           int           //座位号
	Time             time.Time
	Ts               int32
}

func NewSceneCoinLog() *SceneCoinLog {
	log := &SceneCoinLog{LogId: bson.NewObjectId()}
	return log
}

func NewSceneCoinLogEx(snid int32, changecoin, oldcoin, newcoin, eventtype, baserate, totalbet int64, gameid int32, ip string, gameidex int32, seatid int, platform, channel, promoter string, sceneid, gamemode, gamefreeid int32, taxcoin, wincoin, jackpotWinCoin, smallGameWinCoin int64, packageid string) *SceneCoinLog {
	cl := NewSceneCoinLog()
	cl.SnId = snid
	cl.ChangeCoin = changecoin
	cl.OldCoin = oldcoin
	cl.NewCoin = newcoin
	cl.TotalBet = totalbet
	cl.BaseRate = baserate
	cl.GameId = gameid
	cl.GameIdex = gameidex
	cl.SeatId = seatid
	cl.EventType = eventtype
	cl.Ip = ip
	tNow := time.Now()
	cl.Ts = int32(tNow.Unix())
	cl.Time = tNow
	cl.Platform = platform
	cl.Channel = channel
	cl.Promoter = promoter
	cl.SceneId = sceneid
	cl.GameFreeid = gamefreeid
	cl.GameMode = gamemode
	cl.TaxCoin = taxcoin
	cl.WinCoin = wincoin
	cl.JackpotWinCoin = jackpotWinCoin
	cl.SmallGameWinCoin = smallGameWinCoin
	cl.PackageTag = packageid
	return cl
}

func InsertSceneCoinLog(log *SceneCoinLog) (err error) {
	if rpcCli == nil {
		logger.Logger.Error("model.InsertSceneCoinLog rpcCli == nil")
		return
	}

	var ret bool
	err = rpcCli.CallWithTimeout("SceneCoinLogSvc.InsertSceneCoinLog", log, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Warn("InsertSceneCoinLog error:", err)
	}
	return
}
