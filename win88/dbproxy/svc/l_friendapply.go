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
	FriendApplyDBName   = "log"
	FriendApplyCollName = "log_friendapply"
	FriendApplyColError = errors.New("friendapply collection open failed")
)

func FriendApplyCollection(plt string) *mongo.Collection {
	s := mongo.MgoSessionMgrSington.GetPltMgoSession(plt, FriendApplyDBName)
	if s != nil {
		c, first := s.DB().C(FriendApplyCollName)
		if first {
			c.EnsureIndex(mgo.Index{Key: []string{"snid"}, Background: true, Sparse: true})
		}
		return c
	}
	return nil
}

type FriendApplySvc struct {
}

func (svc *FriendApplySvc) UpsertFriendApply(args *model.FriendApplyByKey, ret *model.FriendApplyRet) error {
	cc := FriendApplyCollection(args.Platform)
	if cc == nil {
		return FriendApplyColError
	}
	_, err := cc.Upsert(bson.M{"snid": args.SnId}, args.FA)
	if err != nil && err != mgo.ErrNotFound {
		logger.Logger.Warn("UpsertFriendApply is err: ", err)
		return err
	}
	ret.FA = args.FA
	return nil
}

func (svc *FriendApplySvc) QueryFriendApplyByKey(args *model.FriendApplyByKey, ret *model.FriendApplyRet) error {
	fc := FriendApplyCollection(args.Platform)
	if fc == nil {
		return FriendApplyColError
	}
	err := fc.Find(bson.M{"snid": args.SnId}).One(&ret.FA)
	if err != nil && err != mgo.ErrNotFound {
		logger.Logger.Warn("QueryFriendApplyByKey is err: ", err)
		return err
	}
	return nil
}

func (svc *FriendApplySvc) DelFriendApply(args *model.FriendApplyByKey, ret *bool) error {
	cc := FriendApplyCollection(args.Platform)
	if cc == nil {
		return FriendApplyColError
	}
	err := cc.Remove(bson.M{"snid": args.SnId})
	if err != nil {
		logger.Logger.Warn("DelFriendApply is err: ", err)
		return err
	}
	return nil
}

var _FriendApplySvc = &FriendApplySvc{}

func init() {
	rpc.Register(_FriendApplySvc)
}
