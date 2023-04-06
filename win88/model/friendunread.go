package model

import (
	"github.com/globalsign/mgo/bson"
	"github.com/idealeak/goserver/core/logger"
	"time"
)

type FriendUnread struct {
	Id          bson.ObjectId `bson:"_id"`
	Snid        int32
	UnreadSnids []*FriendUnreadSnid
}

type FriendUnreadSnid struct {
	SnId      int32
	UnreadNum int32
}

type FriendUnreadRet struct {
	FU *FriendUnread
}

type FriendUnreadByKey struct {
	Platform string
	SnId     int32
	FU       *FriendUnread
}

func NewFriendUnread(snid int32) *FriendUnread {
	f := &FriendUnread{Id: bson.NewObjectId()}
	f.Snid = snid
	f.UnreadSnids = []*FriendUnreadSnid{}
	return f
}

func UpsertFriendUnread(platform string, snid int32, fu *FriendUnread) *FriendUnread {
	if rpcCli == nil {
		logger.Logger.Error("model.UpsertFriendUnread rpcCli == nil")
		return nil
	}
	if fu.UnreadSnids == nil || len(fu.UnreadSnids) == 0 {
		DelFriendUnread(platform, snid)
		return nil
	}
	args := &FriendUnreadByKey{
		Platform: platform,
		SnId:     snid,
		FU:       fu,
	}
	var ret *FriendUnreadRet
	err := rpcCli.CallWithTimeout("FriendUnreadSvc.UpsertFriendUnread", args, ret, time.Second*30)
	if err != nil {
		logger.Logger.Warn("UpsertFriendUnread error:", err)
		return nil
	}
	return ret.FU
}

func QueryFriendUnreadBySnid(platform string, snid int32) (fu *FriendUnread, err error) {
	if rpcCli == nil {
		logger.Logger.Error("model.QueryFriendUnreadBySnid rpcCli == nil")
		return
	}
	args := &FriendUnreadByKey{
		Platform: platform,
		SnId:     snid,
	}
	var ret *FriendUnreadRet
	err = rpcCli.CallWithTimeout("FriendUnreadSvc.QueryFriendUnreadByKey", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Warn("QueryFriendUnreadBySnid error:", err)
	}
	if ret != nil {
		fu = ret.FU
	}
	return
}

func DelFriendUnread(platform string, snid int32) {
	if rpcCli == nil {
		logger.Logger.Error("model.UpsertFriendUnread rpcCli == nil")
		return
	}
	args := &FriendUnreadByKey{
		Platform: platform,
		SnId:     snid,
	}
	err := rpcCli.CallWithTimeout("FriendUnreadSvc.DelFriendUnread", args, nil, time.Second*30)
	if err != nil {
		logger.Logger.Warn("DelFriendUnread error:", err)
	}
	return
}
