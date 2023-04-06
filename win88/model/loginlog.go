package model

import (
	"github.com/globalsign/mgo/bson"
	"github.com/idealeak/goserver/core/logger"
	"time"
)

const (
	LoginLogType_Login int32 = iota
	LoginLogType_Logout
	LoginLogType_Rehold
	LoginLogType_Drop
	LoginLogType_Max
)

var (
	LoginLogDBName   = "log"
	LoginLogCollName = "log_login"
)

type ClientLoginInfo struct {
	LoginType    int32
	ApkVer       int32
	ResVer       int32
	InviterId    int32
	PromoterTree int32
	Sid          int64
	UserName     string
	PlatformTag  string
	Promoter     string
}

type LoginLog struct {
	LogId        bson.ObjectId `bson:"_id"`
	Platform     string        //平台id
	Channel      string        //渠道
	Promoter     string        //推广员
	Package      string        //包名
	PackageTag   string        //推广包标识
	InviterId    int32         //邀请人id
	PromoterTree int32
	SnId         int32
	LogType      int32
	ApkVer       int32
	ResVer       int32
	Tel          string
	IP           string
	City         string
	IsBind       bool
	TotalCoin    int64 // 总余额
	GameId       int   // 玩家掉线时所在游戏
	Time         time.Time
}

func NewLoginLog(snid, logType int32, tel, ip, platform, channel, promoter, packageid, city string,
	clog *ClientLoginInfo, totalCoin int64, gameid int) *LoginLog {
	cl := &LoginLog{LogId: bson.NewObjectId()}
	cl.SnId = snid
	cl.LogType = logType
	cl.IP = ip
	cl.Tel = tel
	cl.City = city
	cl.Time = time.Now()
	cl.Platform = platform
	cl.Channel = channel
	cl.Promoter = promoter
	cl.PackageTag = packageid
	if clog != nil {
		cl.InviterId = clog.InviterId
		cl.ApkVer = clog.ApkVer
		cl.ResVer = clog.ResVer
		cl.PackageTag = clog.PlatformTag
	}
	cl.TotalCoin = totalCoin
	cl.GameId = gameid
	if tel != "" {
		cl.IsBind = true
	} else {
		cl.IsBind = false
	}

	return cl
}

func InsertLoginLogs(logs ...*LoginLog) (err error) {
	if rpcCli == nil {
		logger.Logger.Error("model.InsertLoginLogs rpcCli == nil")
		return
	}
	var ret bool
	err = rpcCli.CallWithTimeout("LoginsLogSvc.InsertLoginLogs", logs, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Warn("InsertLoginLogs error:", err)
	}
	return
}

func InsertSignleLoginLog(log *LoginLog) (err error) {
	if rpcCli == nil {
		logger.Logger.Error("model.InsertLoginLogs rpcCli == nil")
		return
	}
	var ret bool
	err = rpcCli.CallWithTimeout("LoginsLogSvc.InsertLoginLogs", log, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Warn("InsertLoginLogs error:", err)
	}
	return
}
