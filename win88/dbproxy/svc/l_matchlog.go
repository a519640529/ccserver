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
	MatchLogDBErr = errors.New("log_matchlog db open failed.")
)

func MatchLogCollection(plt string) *mongo.Collection {
	s := mongo.MgoSessionMgrSington.GetPltMgoSession(plt, model.MatchLogDBName)
	if s != nil {
		c, first := s.DB().C(model.MatchLogCollName)
		if first {
			c.EnsureIndex(mgo.Index{Key: []string{"gamefreeid"}, Background: true, Sparse: true})
		}
		return c
	}
	return nil
}

func InsertMatchLogs(logs ...*model.MatchLog) (err error) {
	clog := MatchLogCollection(logs[0].Platform)
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
		logger.Logger.Warn("InsertMatchLogs error:", err)
		return
	}
	return
}

func RemoveMatchLogs(plt string, ts time.Time) (*mgo.ChangeInfo, error) {
	return MatchLogCollection(plt).RemoveAll(bson.M{"endtime": bson.M{"$lt": ts}})
}

type MatchLogSvc struct {
}

func (svc *MatchLogSvc) InsertMatchLogs(args []*model.MatchLog, ret *bool) (err error) {
	err = InsertMatchLogs(args...)
	if err == nil {
		*ret = true
	}
	return
}

func (svc *MatchLogSvc) RemoveMatchLogs(args *model.RemoveMatchLogsArgs, ret **mgo.ChangeInfo) (err error) {
	*ret, err = RemoveMatchLogs(args.Plt, args.Ts)
	return
}

func init() {
	rpc.Register(new(MatchLogSvc))
}
