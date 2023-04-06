package model

import (
	"github.com/globalsign/mgo/bson"
	"github.com/idealeak/goserver/core/logger"
	"time"
)

type Friend struct {
	Id         bson.ObjectId `bson:"_id"`
	Platform   string
	SnId       int32
	BindFriend []*BindFriend
	UpdateTime int64
	Name       string
	Head       int32
	Sex        int32
	Coin       int64
	Diamond    int64
	VCard      int64
	Roles      map[int32]int32 //人物
	Pets       map[int32]int32 //宠物
	Shield     []int32
	LogoutTime int64 //登出时间
}

type BindFriend struct {
	SnId       int32
	CreateTime int64 //建立时间
}

type FriendRet struct {
	Fri *Friend
}

type FriendByKey struct {
	Platform string
	SnId     int32
}

type FriendsRet struct {
	Fris []*Friend
}

type FriendsByKey struct {
	Platform string
	SnIds    []int32
}

func NewFriend(platform string, snid int32, name string, head, sex int32, coin, diamond, vCard int64, roles, pets map[int32]int32) *Friend {
	f := &Friend{Id: bson.NewObjectId()}
	f.Platform = platform
	f.SnId = snid
	f.Name = name
	f.Head = head
	f.Sex = sex
	f.Coin = coin
	f.Diamond = diamond
	f.VCard = vCard
	f.Roles = roles
	f.Pets = pets
	f.Shield = []int32{}
	f.UpdateTime = time.Now().Unix()
	f.BindFriend = []*BindFriend{}
	return f
}

func UpsertFriend(friend *Friend) *Friend {
	if rpcCli == nil {
		logger.Logger.Error("model.UpsertFriend rpcCli == nil")
		return nil
	}

	ret := &FriendRet{}
	err := rpcCli.CallWithTimeout("FriendSvc.UpsertFriend", friend, ret, time.Second*30)
	if err != nil {
		logger.Logger.Error("UpsertFriend error:", err)
		return nil
	}
	return ret.Fri
}

func QueryFriendBySnid(platform string, snid int32) (friend *Friend, err error) {
	if rpcCli == nil {
		logger.Logger.Error("model.QueryFriendBySnid rpcCli == nil")
		return
	}
	args := &FriendByKey{
		Platform: platform,
		SnId:     snid,
	}
	var ret *FriendRet
	err = rpcCli.CallWithTimeout("FriendSvc.QueryFriendByKey", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Error("QueryFriendBySnid error:", err)
	}
	if ret != nil {
		friend = ret.Fri
	}
	return
}

func QueryFriendsBySnids(platform string, snids []int32) (fris []*Friend, err error) {
	if rpcCli == nil {
		logger.Logger.Error("model.QueryFriendsBySnids rpcCli == nil")
		return
	}
	args := &FriendsByKey{
		Platform: platform,
		SnIds:    snids,
	}
	ret := &FriendsRet{}
	err = rpcCli.CallWithTimeout("FriendSvc.QueryFriendsByKey", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Error("QueryFriendsBySnids error:", err)
	}
	if ret != nil {
		fris = ret.Fris
	}
	return
}
