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
	MatchSeasonIdDBName   = "log"
	MatchSeasonIdCollName = "log_matchseasonid"
	MatchSeasonIdColError = errors.New("MatchSeasonId collection open failed")
)

func MatchSeasonIdCollection(plt string) *mongo.Collection {
	s := mongo.MgoSessionMgrSington.GetPltMgoSession(plt, MatchSeasonIdDBName)
	if s != nil {
		c, first := s.DB().C(MatchSeasonIdCollName)
		if first {
			c.EnsureIndex(mgo.Index{Key: []string{"platform"}, Background: true, Sparse: true})
		}
		return c
	}
	return nil
}

type MatchSeasonIdSvc struct {
}

func (svc *MatchSeasonIdSvc) UpsertMatchSeasonId(args *model.MatchSeasonId, ret *model.MatchSeasonIdRet) error {
	cc := MatchSeasonIdCollection(args.Platform)
	if cc == nil {
		return MatchSeasonIdColError
	}
	_, err := cc.Upsert(bson.M{"platform": args.Platform}, args)
	if err != nil && err != mgo.ErrNotFound {
		logger.Logger.Warn("UpsertMatchSeasonId is err: ", err)
		return err
	}
	return nil
}

func (svc *MatchSeasonIdSvc) QueryMatchSeasonId(args *model.MatchSeasonId, ret *model.MatchSeasonIdRet) error {
	fc := MatchSeasonIdCollection(args.Platform)
	if fc == nil {
		return MatchSeasonIdColError
	}
	err := fc.Find(bson.M{"platform": args.Platform}).One(&ret.MsId)
	if err != nil && err != mgo.ErrNotFound {
		logger.Logger.Warn("QueryMatchSeasonId is err: ", err)
		return err
	}
	return nil
}

var _MatchSeasonIdSvc = &MatchSeasonIdSvc{}

func init() {
	rpc.Register(_MatchSeasonIdSvc)
}
