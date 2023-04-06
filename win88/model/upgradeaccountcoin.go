package model

import (
	"time"

	"github.com/globalsign/mgo/bson"
)

var (
	UpgradeAccountCoinDBName   = "log"
	UpgradeAccountCoinCollName = "log_upgradeaccountcoin"
)

type UpGradeAccountCoin struct {
	LogId      bson.ObjectId `bson:"_id"`
	IP         string        //IP
	Date       int32         //YYYYMMDD
	Coin       int32         //金币
	SnId       int32         //玩家id
	Platform   string        //平台名称
	Channel    string        //渠道名称
	Promoter   string        //推广员
	PackageTag string        //包标识
	InviterId  int32         //邀请人
	City       string        //城市
	CreateTime time.Time     //创建日期
}

type UpGradeAccountCoinArgs struct {
	Plt  string
	Ip   string
	Date int32
}

func GetUpgradeAccountCoinLogsByIPAndDate(ip, platform string, logTime time.Time) int32 {
	year, month, day := logTime.Date()
	date := year*10000 + int(month)*100 + day
	if rpcCli == nil {
		return 0
	}

	args := &UpGradeAccountCoinArgs{Plt: platform, Ip: ip, Date: int32(date)}
	var count int32
	err := rpcCli.CallWithTimeout("UpGradeAccountCoinSvc.GetUpgradeAccountCoinLogsByIPAndDate", args, &count, time.Second*30)
	if err != nil {
		return 0
	}
	return count
}

func InsertUpgradeAccountCoinLog(ip string, logTime time.Time, coin int32, snId int32, channel, platform, promoter, packageTag, city string, inviterId int32) error {
	if rpcCli == nil {
		return ErrRPClientNoConn
	}
	year, month, day := logTime.Date()
	date := year*10000 + int(month)*100 + day
	log := &UpGradeAccountCoin{LogId: bson.NewObjectId(), IP: ip, Date: int32(date), Coin: coin, SnId: snId, Platform: platform, Channel: channel, City: city, Promoter: promoter, PackageTag: packageTag, InviterId: inviterId, CreateTime: time.Now()}
	var ret bool
	err := rpcCli.CallWithTimeout("UpGradeAccountCoinSvc.InsertUpgradeAccountCoinLog", log, &ret, time.Second*30)
	if err != nil {
		return err
	}
	return err
}
