package model

import (
	"fmt"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/idealeak/goserver/core/logger"
)

var (
	CoinPoolSettingDBName      = "user"
	CoinPoolSettingCollName    = "user_coinpoolsetting"
	CoinPoolSettingHisDBName   = "log"
	CoinPoolSettingHisCollName = "log_coinpoolsetting"
)

var CoinPoolSettingDatas = make(map[string]*CoinPoolSetting)

type CoinPoolSetting struct {
	Id               bson.ObjectId `bson:"_id"`
	Platform         string        //平台id
	GameFreeId       int32         //游戏id
	GroupId          int32         //组id
	ServerId         int32         //服务器id
	InitValue        int32         //初始库存值
	LowerLimit       int32         //库存下限
	UpperLimit       int32         //库存上限
	UpperOffsetLimit int32         //上限偏移值
	MaxOutValue      int32         //最大吐钱数
	ChangeRate       int32         //库存变化速度
	MinOutPlayerNum  int32         //最少吐钱人数
	UpperLimitOfOdds int32         //赔率上限(万分比)
	BaseRate         int32         //基础赔率
	CtroRate         int32         //调节赔率
	HardTimeMin      int32         //收分调节频率下限
	HardTimeMax      int32         //收分调节频率上限
	NormalTimeMin    int32         //正常调节频率下限
	NormalTimeMax    int32         //正常调节频率上限
	EasyTimeMin      int32         //放分调节频率下限
	EasyTimeMax      int32         //放分调节频率上限
	EasrierTimeMin   int32         //吐分调节频率下限
	EasrierTimeMax   int32         //吐分分调节频率上限
	CpCangeType      int32
	CpChangeInterval int32
	CpChangeTotle    int32
	CpChangeLower    int32
	CpChangeUpper    int32
	ProfitRate       int32 //收益抽取比例
	CoinPoolMode     int32 //当前池比例
	ResetTime        int32 //重置时间
	UptTime          time.Time
}

func NewCoinPoolSetting(platform string, groupId, gamefreeid, srverId, initValue, lowerLimit, upperLimit, upperOffsetLimit, maxOutValue, changeRate, minOutPlayerNum, upperLimitOfOdds, baseRate, ctroRate, hardTimeMin, hardTimeMax, normalTimeMin, normalTimeMax, easyTimeMin, easyTimeMax, easrierTimeMin, easrierTimeMax, profitRate int32, coinPoolMode int32) *CoinPoolSetting {
	cl := &CoinPoolSetting{Id: bson.NewObjectId()}
	cl.Platform = platform
	cl.GroupId = groupId
	cl.GameFreeId = gamefreeid
	cl.ServerId = srverId
	cl.InitValue = initValue
	cl.LowerLimit = lowerLimit
	cl.UpperLimit = upperLimit
	cl.UpperOffsetLimit = upperOffsetLimit
	cl.MaxOutValue = maxOutValue
	cl.ChangeRate = changeRate
	cl.MinOutPlayerNum = minOutPlayerNum
	cl.UpperLimitOfOdds = upperLimitOfOdds
	cl.BaseRate = baseRate
	cl.CtroRate = ctroRate
	cl.HardTimeMin = hardTimeMin
	cl.HardTimeMax = hardTimeMax
	cl.NormalTimeMin = normalTimeMin
	cl.NormalTimeMax = normalTimeMax
	cl.EasyTimeMin = easyTimeMin
	cl.EasyTimeMax = easyTimeMax
	cl.EasrierTimeMin = easrierTimeMin
	cl.EasrierTimeMax = easrierTimeMax
	cl.ProfitRate = profitRate
	cl.CoinPoolMode = coinPoolMode
	cl.UptTime = time.Now()
	return cl
}

func GetAllCoinPoolSettingData() error {
	if rpcCli == nil {
		return ErrRPClientNoConn
	}
	var datas []CoinPoolSetting
	err := rpcCli.CallWithTimeout("CoinPoolSettingSvc.GetAllCoinPoolSettingData", struct{}{}, &datas, time.Second*30)
	if err != nil {
		logger.Logger.Warn("GetAllCoinPoolSettingData error:", err)
		return err
	}
	for i := 0; i < len(datas); i++ {
		dbSetting := &datas[i]
		ManageCoinPoolSetting(dbSetting)
	}
	return nil
}

type UpsertCoinPoolSettingArgs struct {
	Cps *CoinPoolSetting
	Old *CoinPoolSetting
}

func UpsertCoinPoolSetting(cps, old *CoinPoolSetting) (err error) {
	if rpcCli == nil {
		return ErrRPClientNoConn
	}
	var ret bool
	args := &UpsertCoinPoolSettingArgs{
		Cps: cps,
		Old: old,
	}
	err = rpcCli.CallWithTimeout("CoinPoolSettingSvc.UpsertCoinPoolSetting", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Warn("UpsertCoinPoolSetting error:", err)
		return err
	}
	return
}

func GetCoinPoolSetting(gameFreeId, serverId, groupId int32, platform string) *CoinPoolSetting {
	var key string
	if groupId != 0 {
		key = fmt.Sprintf("%v+%v_%v", gameFreeId, groupId, serverId)
	} else {
		key = fmt.Sprintf("%v_%v_%v", gameFreeId, platform, serverId)
	}
	if data, exist := CoinPoolSettingDatas[key]; exist {
		return data
	}
	return nil
}

func ManageCoinPoolSetting(dbSetting *CoinPoolSetting) {
	var key string
	if dbSetting.GroupId != 0 {
		key = fmt.Sprintf("%v+%v_%v", dbSetting.GameFreeId, dbSetting.GroupId, dbSetting.ServerId)
	} else {
		key = fmt.Sprintf("%v_%v_%v", dbSetting.GameFreeId, dbSetting.Platform, dbSetting.ServerId)
	}
	CoinPoolSettingDatas[key] = dbSetting
}

// 删除水池历史调控记录
func RemoveCoinPoolSettingHis(ts time.Time) (*mgo.ChangeInfo, error) {
	if rpcCli == nil {
		return nil, ErrRPClientNoConn
	}
	var ret mgo.ChangeInfo
	err := rpcCli.CallWithTimeout("CoinPoolSettingSvc.RemoveCoinPoolSettingHis", ts, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Warn("RemoveCoinPoolSettingHis error:", err)
		return &ret, err
	}
	return &ret, err
}
