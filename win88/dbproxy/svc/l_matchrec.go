package svc

import (
	"errors"
	"net/rpc"
	"time"

	"games.yol.com/win88/dbproxy/mongo"
	"games.yol.com/win88/model"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/idealeak/goserver/core/logger"
)

var (
	MatchRecDBErr = errors.New("log_matchrec db open failed.")
)

func MatchRecCollection(plt string) *mongo.Collection {
	s := mongo.MgoSessionMgrSington.GetPltMgoSession(plt, model.MatchRecDBName)
	if s != nil {
		c, first := s.DB().C(model.MatchRecCollName)
		if first {
			c.EnsureIndex(mgo.Index{Key: []string{"snid"}, Background: true, Sparse: true})
			c.EnsureIndex(mgo.Index{Key: []string{"matchid"}, Background: true, Sparse: true})
		}
		return c
	}
	return nil
}

func InsertMatchRecs(logs ...*model.MatchRec) (err error) {
	clog := MatchRecCollection(logs[0].Platform)
	if clog == nil {
		return
	}
	switch len(logs) {
	case 0:
		return errors.New("no data")
	case 1:
		err = clog.Insert(logs[0])
	default:
		docs := make([]interface{}, 0, len(logs))
		for _, log := range logs {
			docs = append(docs, log)
		}
		err = clog.Insert(docs...)
	}
	if err != nil {
		logger.Logger.Warn("InsertMatchRecs error:", err)
		return
	}
	return
}

func FetchMatchRecs(plt string, snid int32) (recs []model.MatchRec, err error) {
	start := time.Now().AddDate(0, 0, -7)
	err = MatchRecCollection(plt).Find(bson.M{"snid": snid, "ts": bson.M{"$gt": start}}).All(&recs)
	return
}

func RemoveMatchRecs(plt string, ts time.Time) (*mgo.ChangeInfo, error) {
	return MatchRecCollection(plt).RemoveAll(bson.M{"ts": bson.M{"$lt": ts}})
}

type MatchRecSvc struct {
}

func (svc *MatchRecSvc) InsertMatchRecs(args []*model.MatchRec, ret *bool) (err error) {
	err = InsertMatchRecs(args...)
	if err == nil {
		*ret = true
	}
	return
}

func (svc *MatchRecSvc) FetchMatchRecs(args *model.FetchMatchRecsArgs, ret *[]model.MatchRec) (err error) {
	*ret, err = FetchMatchRecs(args.Plt, args.SnId)
	return
}

func (svc *MatchRecSvc) RemoveMatchRecs(args *model.RemoveMatchRecsArgs, ret **mgo.ChangeInfo) (err error) {
	*ret, err = RemoveMatchRecs(args.Plt, args.Ts)
	return
}

func init() {
	rpc.Register(new(MatchRecSvc))
}