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
	MatchSeasonRankDBName   = "log"
	MatchSeasonRankCollName = "log_matchseasonrank"
	MatchSeasonRankColError = errors.New("MatchSeasonRank collection open failed")
)

func MatchSeasonRankCollection(plt string) *mongo.Collection {
	s := mongo.MgoSessionMgrSington.GetPltMgoSession(plt, MatchSeasonRankDBName)
	if s != nil {
		c, first := s.DB().C(MatchSeasonRankCollName)
		if first {
			c.EnsureIndex(mgo.Index{Key: []string{"platform"}, Background: true, Sparse: true})
		}
		return c
	}
	return nil
}

type MatchSeasonRankSvc struct {
}

func (svc *MatchSeasonRankSvc) UpsertMatchSeasonRank(args *model.MatchSeasonRankByKey, ret *model.MatchSeasonRankRet) error {
	cc := MatchSeasonRankCollection(args.Platform)
	if cc == nil {
		return MatchSeasonRankColError
	}
	if args.MsRanks != nil && len(args.MsRanks) > 0 {
		for _, rank := range args.MsRanks {
			err := cc.Insert(rank)
			if err != nil && err != mgo.ErrNotFound {
				logger.Logger.Warn("UpsertMatchSeasonRank is err: ", err)
				return err
			}
		}
	}
	return nil
}

func (svc *MatchSeasonRankSvc) QueryMatchSeasonRank(args *model.MatchSeasonRankByKey, ret *model.MatchSeasonRankRet) error {
	fc := MatchSeasonRankCollection(args.Platform)
	if fc == nil {
		return MatchSeasonRankColError
	}
	err := fc.Find(bson.M{"platform": args.Platform}).All(&ret.MsRanks)
	if err != nil && err != mgo.ErrNotFound {
		logger.Logger.Warn("QueryMatchSeasonRank is err: ", err)
		return err
	}
	return nil
}

func (svc *MatchSeasonRankSvc) DropMatchSeasonRank(args *model.MatchSeasonRankByKey, ret *model.MatchSeasonRankRet) error {
	fc := MatchSeasonRankCollection(args.Platform)
	if fc == nil {
		return MatchSeasonRankColError
	}
	_, err := fc.RemoveAll(nil)
	if err != nil && err != mgo.ErrNotFound {
		logger.Logger.Warn("DropMatchSeasonRank is err: ", err)
		return err
	}
	return nil
}

var _MatchSeasonRankSvc = &MatchSeasonRankSvc{}

func init() {
	rpc.Register(_MatchSeasonRankSvc)
}
