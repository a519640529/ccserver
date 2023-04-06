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
	MatchSeasonDBName   = "log"
	MatchSeasonCollName = "log_matchseason"
	MatchSeasonColError = errors.New("MatchSeason collection open failed")
)

func MatchSeasonCollection(plt string) *mongo.Collection {
	s := mongo.MgoSessionMgrSington.GetPltMgoSession(plt, MatchSeasonDBName)
	if s != nil {
		c, first := s.DB().C(MatchSeasonCollName)
		if first {
			c.EnsureIndex(mgo.Index{Key: []string{"snid"}, Background: true, Sparse: true})
		}
		return c
	}
	return nil
}

type MatchSeasonSvc struct {
}

func (svc *MatchSeasonSvc) UpsertMatchSeason(args *model.MatchSeason, ret *model.MatchSeasonRet) error {
	cc := MatchSeasonCollection(args.Platform)
	if cc == nil {
		return MatchSeasonColError
	}
	_, err := cc.Upsert(bson.M{"snid": args.SnId}, args)
	if err != nil && err != mgo.ErrNotFound {
		logger.Logger.Warn("UpsertMatchSeason is err: ", err)
		return err
	}
	return nil
}

func (svc *MatchSeasonSvc) QueryMatchSeasonByKey(args *model.MatchSeasonByKey, ret *model.MatchSeasonRet) error {
	fc := MatchSeasonCollection(args.Platform)
	if fc == nil {
		return MatchSeasonColError
	}
	err := fc.Find(bson.M{"snid": args.SnId}).One(&ret.Ms)
	if err != nil && err != mgo.ErrNotFound {
		logger.Logger.Warn("QueryMatchSeasonByKey is err: ", err)
		return err
	}
	return nil
}

func (svc *MatchSeasonSvc) QueryMatchSeason(platform string, ret *model.MatchSeasonRets) error {
	fc := MatchSeasonCollection(platform)
	if fc == nil {
		return MatchSeasonColError
	}

	err := fc.Find(bson.M{"platform": platform}).Sort("-lv").Limit(model.GameParamData.MatchSeasonRankMaxNum).All(&ret.Mss)
	if err != nil && err != mgo.ErrNotFound {
		logger.Logger.Warn("QueryMatchSeason is err: ", err)
		return err
	}
	return nil
}

func (svc *MatchSeasonSvc) DelMatchSeason(args *model.MatchSeasonByKey, ret *bool) error {
	cc := MatchSeasonCollection(args.Platform)
	if cc == nil {
		return MatchSeasonColError
	}
	err := cc.Remove(bson.M{"snid": args.SnId})
	if err != nil {
		logger.Logger.Warn("DelMatchSeason is err: ", err)
		return err
	}
	return nil
}

var _MatchSeasonSvc = &MatchSeasonSvc{}

func init() {
	rpc.Register(_MatchSeasonSvc)
}
