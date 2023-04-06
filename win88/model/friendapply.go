package model

import (
	"github.com/globalsign/mgo/bson"
	"github.com/idealeak/goserver/core/logger"
	"time"
)

type FriendApply struct {
	Id         bson.ObjectId `bson:"_id"`
	Snid       int32
	ApplySnids []*FriendApplySnid
}

type FriendApplySnid struct {
	SnId     int32
	Name     string
	Head     int32
	CreateTs int64
}

type FriendApplyRet struct {
	FA *FriendApply
}

type FriendApplyByKey struct {
	Platform string
	SnId     int32
	FA       *FriendApply
}

func NewFriendApply(snid int32) *FriendApply {
	f := &FriendApply{Id: bson.NewObjectId()}
	f.Snid = snid
	f.ApplySnids = []*FriendApplySnid{}
	return f
}

func UpsertFriendApply(platform string, snid int32, fa *FriendApply) *FriendApply {
	if rpcCli == nil {
		logger.Logger.Error("model.UpsertFriendApply rpcCli == nil")
		return nil
	}
	if fa.ApplySnids == nil || len(fa.ApplySnids) == 0 {
		DelFriendApply(platform, snid)
		return nil
	}
	args := &FriendApplyByKey{
		Platform: platform,
		SnId:     snid,
		FA:       fa,
	}
	var ret *FriendApplyRet
	err := rpcCli.CallWithTimeout("FriendApplySvc.UpsertFriendApply", args, ret, time.Second*30)
	if err != nil {
		logger.Logger.Warn("UpsertFriendApply error:", err)
		return nil
	}
	return ret.FA
}

func QueryFriendApplyBySnid(platform string, snid int32) (fa *FriendApply, err error) {
	if rpcCli == nil {
		logger.Logger.Error("model.QueryFriendApplyBySnid rpcCli == nil")
		return
	}
	args := &FriendApplyByKey{
		Platform: platform,
		SnId:     snid,
	}
	var ret *FriendApplyRet
	err = rpcCli.CallWithTimeout("FriendApplySvc.QueryFriendApplyByKey", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Warn("QueryFriendApplyBySnid error:", err)
	}
	if ret != nil {
		fa = ret.FA
	}
	return
}

func DelFriendApply(platform string, snid int32) {
	if rpcCli == nil {
		logger.Logger.Error("model.UpsertFriendApply rpcCli == nil")
		return
	}
	args := &FriendApplyByKey{
		Platform: platform,
		SnId:     snid,
	}
	err := rpcCli.CallWithTimeout("FriendApplySvc.DelFriendApply", args, nil, time.Second*30)
	if err != nil {
		logger.Logger.Warn("DelFriendApply error:", err)
	}
	return
}
