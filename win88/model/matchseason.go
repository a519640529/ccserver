package model

import (
	"github.com/globalsign/mgo/bson"
	"github.com/idealeak/goserver/core/logger"
	"time"
)

// 段位信息
type MatchSeason struct {
	Id       bson.ObjectId `bson:"_id"`
	Platform string
	SnId     int32
	Name     string
	SeasonId int32 //赛季id
	Lv       int32 //段位
	LastLv   int32 //上赛季段位
	IsAward  bool  //上赛季是否领奖
	AwardTs  int64 //领奖时间
	UpdateTs int64
}

type MatchSeasonRet struct {
	Ms *MatchSeason
}

type MatchSeasonRets struct {
	Mss []*MatchSeason
}

type MatchSeasonByKey struct {
	Platform string
	SnId     int32
}

func NewMatchSeason(platform string, snid int32, name string, sid, lv int32) *MatchSeason {
	ms := &MatchSeason{Id: bson.NewObjectId()}
	ms.Platform = platform
	ms.SnId = snid
	ms.Name = name
	ms.SeasonId = sid
	ms.Lv = lv
	ms.LastLv = 0
	ms.IsAward = false
	ms.AwardTs = 0
	ms.UpdateTs = time.Now().Unix()
	return ms
}

func UpsertMatchSeason(MatchSeason *MatchSeason) *MatchSeason {
	if rpcCli == nil {
		logger.Logger.Error("model.UpsertMatchSeason rpcCli == nil")
		return nil
	}

	ret := &MatchSeasonRet{}
	err := rpcCli.CallWithTimeout("MatchSeasonSvc.UpsertMatchSeason", MatchSeason, ret, time.Second*30)
	if err != nil {
		logger.Logger.Error("UpsertMatchSeason error:", err)
		return nil
	}
	return ret.Ms
}

func QueryMatchSeasonBySnid(platform string, snid int32) (MatchSeason *MatchSeason, err error) {
	if rpcCli == nil {
		logger.Logger.Error("model.QueryMatchSeasonBySnid rpcCli == nil")
		return
	}
	args := &MatchSeasonByKey{
		Platform: platform,
		SnId:     snid,
	}
	var ret *MatchSeasonRet
	err = rpcCli.CallWithTimeout("MatchSeasonSvc.QueryMatchSeasonByKey", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Error("QueryMatchSeasonBySnid error:", err)
	}
	if ret != nil {
		MatchSeason = ret.Ms
	}
	return
}

func QueryMatchSeason(platform string) (MatchSeasons []*MatchSeason, err error) {
	if rpcCli == nil {
		logger.Logger.Error("model.QueryMatchSeason rpcCli == nil")
		return
	}
	var ret *MatchSeasonRets
	err = rpcCli.CallWithTimeout("MatchSeasonSvc.QueryMatchSeason", platform, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Error("QueryMatchSeason error:", err)
	}
	if ret != nil {
		MatchSeasons = ret.Mss
	}
	return
}

// 赛季信息
type MatchSeasonId struct {
	Id         bson.ObjectId `bson:"_id"`
	Platform   string
	SeasonId   int32 //赛季id
	StartStamp int64 //开始时间戳
	EndStamp   int64 //结束时间戳
	UpdateTs   int64 //更新时间戳
}

type MatchSeasonIdRet struct {
	MsId *MatchSeasonId
}

type MatchSeasonIdByKey struct {
	Platform string
}

func NewMatchSeasonId(platform string, sid int32, sstamp, estamp int64) *MatchSeasonId {
	ms := &MatchSeasonId{Id: bson.NewObjectId()}
	ms.Platform = platform
	ms.SeasonId = sid
	ms.StartStamp = sstamp
	ms.EndStamp = estamp
	ms.UpdateTs = time.Now().Unix()
	return ms
}

func UpsertMatchSeasonId(MatchSeasonId *MatchSeasonId) *MatchSeasonId {
	if rpcCli == nil {
		logger.Logger.Error("model.UpsertMatchSeasonId rpcCli == nil")
		return nil
	}

	ret := &MatchSeasonIdRet{}
	err := rpcCli.CallWithTimeout("MatchSeasonIdSvc.UpsertMatchSeasonId", MatchSeasonId, ret, time.Second*30)
	if err != nil {
		logger.Logger.Error("UpsertMatchSeasonId error:", err)
		return nil
	}
	return ret.MsId
}

func QueryMatchSeasonId(platform string) (MatchSeasonId *MatchSeasonId, err error) {
	if rpcCli == nil {
		logger.Logger.Error("model.QueryMatchSeasonId rpcCli == nil")
		return
	}
	args := &MatchSeasonIdByKey{
		Platform: platform,
	}
	var ret *MatchSeasonIdRet
	err = rpcCli.CallWithTimeout("MatchSeasonIdSvc.QueryMatchSeasonId", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Error("QueryMatchSeasonId error:", err)
	}
	if ret != nil {
		MatchSeasonId = ret.MsId
	}
	return
}

// 排行榜
type MatchSeasonRank struct {
	Id       bson.ObjectId `bson:"_id"`
	Platform string
	SnId     int32
	Name     string
	Lv       int32 //段位
	UpdateTs int64
}

type MatchSeasonRankRet struct {
	MsRanks []*MatchSeasonRank
}

type MatchSeasonRankByKey struct {
	Platform string
	MsRanks  []*MatchSeasonRank
}

func UpsertMatchSeasonRank(platform string, MatchSeasonRanks []*MatchSeasonRank) {
	if rpcCli == nil {
		logger.Logger.Error("model.UpsertMatchSeasonRank rpcCli == nil")
		return
	}
	args := &MatchSeasonRankByKey{
		Platform: platform,
		MsRanks:  MatchSeasonRanks,
	}
	err := rpcCli.CallWithTimeout("MatchSeasonRankSvc.UpsertMatchSeasonRank", args, nil, time.Second*30)
	if err != nil {
		logger.Logger.Error("UpsertMatchSeasonRank error:", err)
		return
	}
}

func QueryMatchSeasonRank(platform string) (MatchSeasonRanks []*MatchSeasonRank, err error) {
	if rpcCli == nil {
		logger.Logger.Error("model.QueryMatchSeasonRank rpcCli == nil")
		return
	}
	args := &MatchSeasonRankByKey{
		Platform: platform,
	}
	var ret *MatchSeasonRankRet
	err = rpcCli.CallWithTimeout("MatchSeasonRankSvc.QueryMatchSeasonRank", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Error("QueryMatchSeasonRank error:", err)
	}
	if ret != nil {
		MatchSeasonRanks = ret.MsRanks
	}
	return
}

func DropMatchSeasonRank(platform string) {
	if rpcCli == nil {
		logger.Logger.Error("model.QueryMatchSeasonRank rpcCli == nil")
		return
	}
	args := &MatchSeasonRankByKey{
		Platform: platform,
	}
	err := rpcCli.CallWithTimeout("MatchSeasonRankSvc.DropMatchSeasonRank", args, nil, time.Second*30)
	if err != nil {
		logger.Logger.Error("DropMatchSeasonRank error:", err)
	}
	return
}
