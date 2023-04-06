package svc

import (
	"errors"
	"games.yol.com/win88/dbproxy/mongo"
	"games.yol.com/win88/model"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/idealeak/goserver/core/logger"
	"net/rpc"
)

var (
	FriendUnreadDBName   = "log"
	FriendUnreadCollName = "log_friendunread"
	FriendUnreadColError = errors.New("friendunread collection open failed")
)

func FriendUnreadCollection(plt string) *mongo.Collection {
	s := mongo.MgoSessionMgrSington.GetPltMgoSession(plt, FriendUnreadDBName)
	if s != nil {
		c, first := s.DB().C(FriendUnreadCollName)
		if first {
			c.EnsureIndex(mgo.Index{Key: []string{"snid"}, Background: true, Sparse: true})
		}
		return c
	}
	return nil
}

type FriendUnreadSvc struct {
}

func (svc *FriendUnreadSvc) UpsertFriendUnread(args *model.FriendUnreadByKey, ret *model.FriendUnreadRet) error {
	cc := FriendUnreadCollection(args.Platform)
	if cc == nil {
		return FriendUnreadColError
	}
	_, err := cc.Upsert(bson.M{"snid": args.SnId}, args.FU)
	if err != nil && err != mgo.ErrNotFound {
		logger.Logger.Error("UpsertFriendUnread is err: ", err)
		return err
	}
	ret.FU = args.FU
	return nil
}

func (svc *FriendUnreadSvc) QueryFriendUnreadByKey(args *model.FriendUnreadByKey, ret *model.FriendUnreadRet) error {
	fc := FriendUnreadCollection(args.Platform)
	if fc == nil {
		return FriendUnreadColError
	}
	err := fc.Find(bson.M{"snid": args.SnId}).One(&ret.FU)
	if err != nil && err != mgo.ErrNotFound {
		logger.Logger.Error("QueryFriendUnreadByKey is err: ", err)
		return err
	}
	return nil
}

func (svc *FriendUnreadSvc) DelFriendUnread(args *model.FriendUnreadByKey, ret *bool) error {
	cc := FriendUnreadCollection(args.Platform)
	if cc == nil {
		return FriendUnreadColError
	}
	err := cc.Remove(bson.M{"snid": args.SnId})
	if err != nil && err != mgo.ErrNotFound {
		logger.Logger.Error("DelFriendUnread is err: ", err)
		return err
	}
	return nil
}

var _FriendUnreadSvc = &FriendUnreadSvc{}

func init() {
	rpc.Register(_FriendUnreadSvc)
}
