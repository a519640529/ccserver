package model

import (
	"time"

	"github.com/globalsign/mgo/bson"
	"github.com/idealeak/goserver/core/logger"
)

var (
	CoinGiveLogDBName   = "log"
	CoinGiveLogCollName = "log_coingive"
)

const (
	COINGIVETYPE_PAY    = 0 //充值
	COINGIVETYPE_SYSTEM = 1 //系统赠送

)

const (
	COINGIVEVER = 0 //当前数据版本
)

type CoinGiveLog struct {
	LogId            bson.ObjectId `bson:"_id"`
	SnId             int32         //玩家id
	Platform         string        //平台名称
	Channel          string        //渠道名称
	Promoter         string        //推广员
	PromoterTree     int32         //全民树
	PayCoin          int64         //充值金额
	GiveCoin         int64         //赠送金额
	Remark           string        //备注
	Opter            string        //操作接口
	State            int64         //预留处理标记用
	FLow             int64         //完成流水
	Ts               int64         //时间戳 纳秒
	NeedFlowRate     int32         //需要流水比例
	NeedGiveFlowRate int32         //则送需要流水比例
	Time             time.Time     //时间戳
	CoinType         int32         //金币类型 0:钱包 1:保险箱
	RecType          int32         //记录类型 0:充值 1:系统赠送
	LogType          int32         //log类型
	Ver              int32         //数据版本
}

func NewCoinGiveLogEx(snid int32, username string, payCoin, giveCoin int64, coinType, logType, promoterTree, recType, ver int32, platform, channel,
	promoter, remark, op, packageid string, needFlowRate, needGiveFlowRate int32) *CoinGiveLog {
	cl := &CoinGiveLog{LogId: bson.NewObjectId()}
	cl.SnId = snid
	//cl.UserName = username
	cl.Platform = platform
	cl.Channel = channel
	cl.Promoter = promoter
	cl.PayCoin = payCoin
	cl.GiveCoin = giveCoin
	cl.RecType = recType
	cl.CoinType = coinType
	cl.LogType = logType
	cl.Ver = ver
	cl.PromoterTree = promoterTree
	cl.Remark = remark
	cl.Opter = op
	cl.Time = time.Now()
	cl.Ts = cl.Time.Local().UnixNano()
	cl.NeedFlowRate = needFlowRate
	cl.NeedGiveFlowRate = needGiveFlowRate
	//cl.PackageTag = packageid
	return cl
}

type UpdateGiveCoinLastFlowArg struct {
	Plt   string
	LogId string
	Flow  int64
}

func UpdateGiveCoinLastFlow(platform, logid string, flow int64) (err error) {
	if rpcCli == nil {
		logger.Logger.Error("model.UpdateGiveCoinLastFlow rpcCli == nil")
		return
	}
	args := &UpdateGiveCoinLastFlowArg{
		Plt:   platform,
		LogId: logid,
		Flow:  flow,
	}
	var ret bool
	err = rpcCli.CallWithTimeout("CoinGiveLogSvc.UpdateGiveCoinLastFlow", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Warn("UpdateGiveCoinLastFlow error:", err)
	}
	return
}

type GetGiveCoinLastFlowArg struct {
	Plt   string
	LogId string
}

func GetGiveCoinLastFlow(platform, logid string) (log *CoinGiveLog) {
	if rpcCli == nil {
		logger.Logger.Error("model.GetGiveCoinLastFlow rpcCli == nil")
		return nil
	}

	args := &GetGiveCoinLastFlowArg{
		Plt:   platform,
		LogId: logid,
	}
	err := rpcCli.CallWithTimeout("CoinGiveLogSvc.GetGiveCoinLastFlow", args, &log, time.Second*30)
	if err != nil {
		logger.Logger.Warn("GetGiveCoinLastFlow error:", err)
	}

	return
}

type GetCoinGiveLogListArg struct {
	Plt     string
	SnId    int32
	Ver     int32
	TsStart int64
	TsEnd   int64
}

func GetCoinGiveLogList(platform string, snid int32, ver int32, tsStart int64, tsEnd int64) (logs []*CoinGiveLog) {
	if rpcCli == nil {
		logger.Logger.Error("model.GetCoinGiveLogList rpcCli == nil")
		return nil
	}

	args := &GetCoinGiveLogListArg{
		Plt:     platform,
		SnId:    snid,
		Ver:     ver,
		TsStart: tsStart,
		TsEnd:   tsEnd,
	}
	err := rpcCli.CallWithTimeout("CoinGiveLogSvc.GetCoinGiveLogList", args, &logs, time.Second*30)
	if err != nil {
		logger.Logger.Warn("GetCoinGiveLogList error:", err)
	}

	return
}

type GetCoinGiveLogListByStateArg struct {
	Plt   string
	SnId  int32
	State int64
}

func GetCoinGiveLogListByState(platform string, snid int32, state int64) (logs []*CoinGiveLog) {
	if rpcCli == nil {
		logger.Logger.Error("model.GetCoinGiveLogListByState rpcCli == nil")
		return nil
	}

	args := &GetCoinGiveLogListByStateArg{
		Plt:   platform,
		SnId:  snid,
		State: state,
	}

	err := rpcCli.CallWithTimeout("CoinGiveLogSvc.GetCoinGiveLogListByState", args, &logs, time.Second*30)
	if err != nil {
		logger.Logger.Warn("GetCoinGiveLogListByState error:", err)
	}

	return
}

type ResetCoinGiveLogListArg struct {
	Plt  string
	SnId int32
	Ts   int64
}

func ResetCoinGiveLogList(platform string, snid int32, ts int64) error {
	if rpcCli == nil {
		logger.Logger.Error("model.ResetCoinGiveLogList rpcCli == nil")
		return nil
	}

	args := &ResetCoinGiveLogListArg{
		Plt:  platform,
		SnId: snid,
		Ts:   ts,
	}

	var ret bool
	err := rpcCli.CallWithTimeout("CoinGiveLogSvc.ResetCoinGiveLogList", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Warn("ResetCoinGiveLogList error:", err)
	}

	return err
}

type SetCoinGiveLogListArg struct {
	Plt     string
	SnId    int32
	State   int64
	TsStart int64
	TsEnd   int64
}

func SetCoinGiveLogList(platform string, snid int32, startTs int64, endTs int64, v int64) error {
	if rpcCli == nil {
		logger.Logger.Error("model.SetCoinGiveLogList rpcCli == nil")
		return nil
	}

	args := &SetCoinGiveLogListArg{
		Plt:     platform,
		SnId:    snid,
		State:   v,
		TsStart: startTs,
		TsEnd:   endTs,
	}

	var ret bool
	err := rpcCli.CallWithTimeout("CoinGiveLogSvc.SetCoinGiveLogList", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Warn("SetCoinGiveLogList error:", err)
	}
	return err
}

type UpdateCoinLogArg struct {
	Plt   string
	LogId string
	State int64
}

func UpdateCoinLog(platform, logid string, v int64) error {
	if rpcCli == nil {
		logger.Logger.Error("model.UpdateCoinLog rpcCli == nil")
		return nil
	}

	args := &UpdateCoinLogArg{
		Plt:   platform,
		LogId: logid,
		State: v,
	}

	var ret bool
	err := rpcCli.CallWithTimeout("CoinGiveLogSvc.UpdateCoinLog", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Warn("UpdateCoinLog error:", err)
	}

	return err
}

func InsertGiveCoinLog(log *CoinGiveLog) (err error) {
	if rpcCli == nil {
		logger.Logger.Error("model.InsertGiveCoinLog rpcCli == nil")
		return nil
	}

	var ret bool
	err = rpcCli.CallWithTimeout("CoinGiveLogSvc.InsertGiveCoinLog", log, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Warn("InsertGiveCoinLog error:", err)
	}

	return
}
