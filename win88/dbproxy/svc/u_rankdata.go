package svc

import (
	"errors"
	"games.yol.com/win88/dbproxy/mongo"
	"games.yol.com/win88/model"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	//"github.com/idealeak/goserver/core/logger"
	"net/rpc"
)

var (
	RankDBName    = "user"
	RankCollName  = "user_rank"
	RankDataDBErr = errors.New("user_rankdata db open failed.")
)

func RankDataCollection(plt string) *mongo.Collection {
	s := mongo.MgoSessionMgrSington.GetPltMgoSession(plt, RankDBName)
	if s != nil {
		c, first := s.DB().C(RankCollName)
		if first {
			c.EnsureIndex(mgo.Index{Key: []string{"key"}, Background: true, Sparse: true})
		}
		return c
	}
	return nil
}

type RankDataSvc struct {
}

func (svc *RankDataSvc) InitRankData(key string, ret *[]*model.RankData) (err error) {
	s := RankDataCollection(key)
	if s != nil {
		err := s.Find(bson.M{}).All(ret)
		if err != nil {
			return err
		}

		return nil
	}
	return nil
}

func (svc *RankDataSvc) SaveRankData(args *model.SaveRankDataArgs, ret *bool) (err error) {
	s := RankDataCollection(args.Plt)
	if s == nil {
		return RankDataDBErr
	}

	_, err = s.Upsert(bson.M{"key": args.Key}, args.Data)
	*ret = err == nil

	return err
}

func init() {
	rpc.Register(new(RankDataSvc))
}
