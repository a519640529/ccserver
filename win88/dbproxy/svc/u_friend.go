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
	FriendDBName   = "user"
	FriendCollName = "user_friend"
	FriendColError = errors.New("friend collection open failed")
)

func FriendCollection(plt string) *mongo.Collection {
	s := mongo.MgoSessionMgrSington.GetPltMgoSession(plt, FriendDBName)
	if s != nil {
		c, first := s.DB().C(FriendCollName)
		if first {
			c.EnsureIndex(mgo.Index{Key: []string{"snid"}, Background: true, Sparse: true})
			c.EnsureIndex(mgo.Index{Key: []string{"platform"}, Background: true, Sparse: true})
		}
		return c
	}
	return nil
}

type FriendSvc struct {
}

func (svc *FriendSvc) UpsertFriend(args *model.Friend, ret *model.FriendRet) error {
	fc := FriendCollection(args.Platform)
	if fc == nil {
		return FriendColError
	}
	if args.Id == "" {
		args.Id = bson.NewObjectId()
	}
	_, err := fc.Upsert(bson.M{"platform": args.Platform, "snid": args.SnId}, args)
	if err != nil && err != mgo.ErrNotFound {
		logger.Logger.Error("UpsertFriend is err: ", err)
		return err
	}
	ret.Fri = args
	return nil
}

func (svc *FriendSvc) QueryFriendByKey(args *model.FriendByKey, ret *model.FriendRet) error {
	fc := FriendCollection(args.Platform)
	if fc == nil {
		return FriendColError
	}
	err := fc.Find(bson.M{"platform": args.Platform, "snid": args.SnId}).One(&ret.Fri)
	if err != nil && err != mgo.ErrNotFound {
		logger.Logger.Error("QueryFriendByKey is err: ", err)
		return err
	}
	return nil
}
func (svc *FriendSvc) QueryFriendsByKey(args *model.FriendsByKey, ret *model.FriendsRet) error {
	fc := FriendCollection(args.Platform)
	if fc == nil {
		return FriendColError
	}
	var sql []bson.M
	if len(args.SnIds) != 0 {
		for _, snid := range args.SnIds {
			sql = append(sql, bson.M{"snid": snid})
		}
	}
	err := fc.Find(bson.M{"$or": sql}).All(&ret.Fris)
	if err != nil && err != mgo.ErrNotFound {
		logger.Logger.Error("QueryFriendsByKey is err: ", err)
		return err
	}
	return nil
}

var _FriendSvc = &FriendSvc{}

func init() {
	rpc.Register(_FriendSvc)
}
