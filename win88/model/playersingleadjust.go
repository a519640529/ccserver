package model

import (
	"games.yol.com/win88/protocol/webapi"
	"github.com/globalsign/mgo/bson"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"time"
)

type PlayerSingleAdjust struct {
	Platform      string
	Id            bson.ObjectId `bson:"_id"`
	GameFreeId    int32
	SnId          int32
	Mode          int32 //调控模式 1赢 2输 tinyint(1)
	TotalTime     int32 //调控总次数
	CurTime       int32 //当前调控次数
	BetMin        int64 //下注下限
	BetMax        int64 //下注上限
	BankerLoseMin int64 //坐庄被输下限
	BankerWinMin  int64 //坐庄被控赢下限
	CardMin       int32 //牌型下限
	CardMax       int32 //牌型上限
	Priority      int32 //优先级
	WinRate       int32 //万分比
	GameId        int32
	GameMode      int32
	Operator      string
	CreateTime    int64
	UpdateTime    int64
}
type SingleAdjustRet struct {
	Ret []*PlayerSingleAdjust
}

func QueryAllSingleAdjust(platfrom string) (pass []*PlayerSingleAdjust, err error) {
	if rpcCli == nil {
		logger.Logger.Error("model.QueryAllSingleAdjust rpcCli == nil")
		return
	}
	var ret *SingleAdjustRet
	err = rpcCli.CallWithTimeout("SingleAdjustSvc.QueryAllSingleAdjust", platfrom, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Warn("QueryAllSingleAdjust error:", err)
	}
	if ret != nil {
		pass = ret.Ret
	}
	return
}

type SingleAdjustByKey struct {
	Platform string
	SnId     int32
}

func QueryAllSingleAdjustByKey(platform string, snid int32) (pass []*PlayerSingleAdjust, err error) {
	if rpcCli == nil {
		logger.Logger.Error("model.QueryAllSingleAdjust rpcCli == nil")
		return
	}
	args := &SingleAdjustByKey{
		Platform: platform,
		SnId:     snid,
	}
	var ret *SingleAdjustRet
	err = rpcCli.CallWithTimeout("SingleAdjustSvc.QueryAllSingleAdjustByKey", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Warn("QueryAllSingleAdjust error:", err)
	}
	if ret != nil {
		pass = ret.Ret
	}
	return
}
func MarshalSingleAdjust(pas *PlayerSingleAdjust) []byte {
	if pas != nil {
		data, err := netlib.Gob.Marshal(pas)
		if err == nil {
			return data
		}
		logger.Logger.Warn("model.MarshalSingleAdjust err: ", err)
	}
	return nil
}
func UnmarshalSingleAdjust(data []byte) (psa *PlayerSingleAdjust) {
	if data != nil {
		err := netlib.Gob.Unmarshal(data, &psa)
		if err != nil {
			logger.Logger.Warn("model.UnmarshalSingleAdjust err: ", err)
			return nil
		}
	}
	return
}

func WebSingleAdjustToModel(psa *webapi.PlayerSingleAdjust) *PlayerSingleAdjust {
	psa_tmp := &PlayerSingleAdjust{
		Platform:      psa.Platform,
		GameFreeId:    psa.GameFreeId,
		SnId:          psa.SnId,
		Mode:          psa.Mode,
		TotalTime:     psa.TotalTime,
		CurTime:       psa.CurTime,
		BetMin:        psa.BetMin,
		BetMax:        psa.BetMax,
		BankerLoseMin: psa.BankerLoseMin,
		BankerWinMin:  psa.BankerWinMin,
		CardMin:       psa.CardMin,
		CardMax:       psa.CardMax,
		Priority:      psa.Priority,
		WinRate:       psa.WinRate,
		GameId:        psa.GameId,
		GameMode:      psa.GameMode,
		Operator:      psa.Operator,
		CreateTime:    psa.CreateTime,
		UpdateTime:    psa.UpdateTime,
	}
	if psa.Id != "" {
		psa_tmp.Id = bson.ObjectIdHex(psa.Id)
	}
	return psa_tmp
}
func AddNewSingleAdjust(args *PlayerSingleAdjust) (err error) {
	if rpcCli == nil {
		logger.Logger.Error("model.AddNewSingleAdjust rpcCli == nil")
		return
	}
	var ret bool
	err = rpcCli.CallWithTimeout("SingleAdjustSvc.AddNewSingleAdjust", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Warn("AddNewSingleAdjust error:", err)
	}
	return
}
func EditSingleAdjust(args *PlayerSingleAdjust) (err error) {
	if rpcCli == nil {
		logger.Logger.Error("model.EditSingleAdjust rpcCli == nil")
		return
	}
	var ret bool
	err = rpcCli.CallWithTimeout("SingleAdjustSvc.EditSingleAdjust", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Warn("EditSingleAdjust error:", err)
	}
	return
}
func DeleteSingleAdjust(args *PlayerSingleAdjust) (err error) {
	if rpcCli == nil {
		logger.Logger.Error("model.DeleteSingleAdjust rpcCli == nil")
		return
	}
	var ret bool
	err = rpcCli.CallWithTimeout("SingleAdjustSvc.DeleteSingleAdjust", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Warn("DeleteSingleAdjust error:", err)
	}
	return
}
