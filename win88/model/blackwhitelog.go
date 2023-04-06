package model

import (
	"time"

	"github.com/globalsign/mgo/bson"
	"github.com/idealeak/goserver/core/logger"
)

type BlackWhiteCoinLog struct {
	LogId          bson.ObjectId `bson:"_id"`
	SnId           int32         //用户ID
	WBLevel        int32
	WBCoinLimit    int64
	ResetTotalCoin int32
	Platform       string    //平台id
	Time           time.Time //账单时间
}

func NewBlackWhiteCoinLog(snid, wbLevel int32, wbCoinLimit int64, resetTotalCoin int32, platform string) *BlackWhiteCoinLog {
	tNow := time.Now()
	log := &BlackWhiteCoinLog{
		LogId:          bson.NewObjectId(),
		SnId:           snid,
		WBLevel:        wbLevel,
		WBCoinLimit:    wbCoinLimit,
		ResetTotalCoin: resetTotalCoin,
		Platform:       platform,
		Time:           tNow,
	}
	return log
}

type BlackWhiteCoinArg struct {
	Id       bson.ObjectId
	SnId     int32
	BillNo   int64
	Platform string
}
type BlackWhiteCoinRet struct {
	Err  error
	Data *BlackWhiteCoinLog
}

func InsertBlackWhiteCoinLogs(logs ...*BlackWhiteCoinLog) error {
	if rpcCli == nil {
		return ErrRPClientNoConn
	}
	ret := &BlackWhiteCoinRet{}
	err := rpcCli.CallWithTimeout("BlackWhiteCoinSvc.InsertBlackWhiteCoinLogs", logs, ret, time.Second*30)
	if err != nil {
		logger.Logger.Error("model.InsertBlackWhiteCoinLogs error", err)
		return err
	}
	return ret.Err
}
func InsertBlackWhiteCoinLog(log *BlackWhiteCoinLog) error {
	if rpcCli == nil {
		return ErrRPClientNoConn
	}
	ret := &BlackWhiteCoinRet{}
	err := rpcCli.CallWithTimeout("BlackWhiteCoinSvc.InsertBlackWhiteCoinLog", log, ret, time.Second*30)
	if err != nil {
		logger.Logger.Error("model.InsertBlackWhiteCoinLog error", err)
		return err
	}
	return ret.Err
}

func RemoveBlackWhiteCoinLog(id bson.ObjectId, platform string) error {
	if rpcCli == nil {
		return ErrRPClientNoConn
	}
	args := &BlackWhiteCoinArg{
		Id:       id,
		Platform: platform,
	}
	ret := &BlackWhiteCoinRet{}
	err := rpcCli.CallWithTimeout("BlackWhiteCoinSvc.RemoveBlackWhiteCoinLog", args, ret, time.Second*30)
	if err != nil {
		logger.Logger.Error("model.RemoveBlackWhiteCoinLog error", err)
		return err
	}
	return ret.Err
}

func GetBlackWhiteCoinLogByBillNo(snid int32, billNo int64, platform string) (*BlackWhiteCoinLog, error) {
	if rpcCli == nil {
		return nil, ErrRPClientNoConn
	}
	args := &BlackWhiteCoinArg{
		SnId:     snid,
		BillNo:   billNo,
		Platform: platform,
	}
	ret := &BlackWhiteCoinRet{}
	err := rpcCli.CallWithTimeout("BlackWhiteCoinSvc.GetBlackWhiteCoinLogByBillNo", args, ret, time.Second*30)
	if err != nil {
		logger.Logger.Error("model.GetBlackWhiteCoinLogByBillNo error", err)
		return nil, err
	}
	return ret.Data, ret.Err
}
