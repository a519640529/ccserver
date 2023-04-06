package model

import (
	"github.com/globalsign/mgo/bson"
	"time"
)

const (
	THDPLATFORM_DG = "dg"
)

var (
	ThdPlatformDBName   = "user"
	ThdPlatformCollName = "user_thdplatform"
)

type ThirdPlatform struct {
	Coin     int64 //当前额度
	NextCoin int64 //下月额度
}

type PlatformOfThirdPlatform struct {
	Id          bson.ObjectId `bson:"_id"`
	Platform    string
	ThdPlatform map[string]*ThirdPlatform
	CreateTime  time.Time
	LastTime    time.Time
}

func NewThirdPlatform(platform string) *PlatformOfThirdPlatform {
	tNow := time.Now()
	log := &PlatformOfThirdPlatform{
		Id:          bson.NewObjectId(),
		Platform:    platform,
		ThdPlatform: make(map[string]*ThirdPlatform),
		CreateTime:  tNow,
		LastTime:    tNow,
	}
	return log
}

func InsertThirdPlatform(platforms ...*PlatformOfThirdPlatform) (err error) {
	if rpcCli == nil {
		return ErrRPClientNoConn
	}
	var ret bool
	err = rpcCli.CallWithTimeout("PlatformOfThirdPlatformSvc.InsertThirdPlatform", platforms, &ret, time.Second*30)
	return
}

func UpdateThirdPlatform(platform *PlatformOfThirdPlatform) (err error) {
	if rpcCli == nil {
		return ErrRPClientNoConn
	}
	var ret bool
	err = rpcCli.CallWithTimeout("PlatformOfThirdPlatformSvc.UpdateThirdPlatform", platform, &ret, time.Second*30)
	return
}

func GetAllThirdPlatform() (ret []PlatformOfThirdPlatform, err error) {
	if rpcCli == nil {
		return nil, ErrRPClientNoConn
	}

	err = rpcCli.CallWithTimeout("PlatformOfThirdPlatformSvc.GetAllThirdPlatform", struct{}{}, &ret, time.Second*30)
	return
}
