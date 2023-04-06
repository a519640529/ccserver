package svc

import (
	"errors"
	"net/rpc"

	"games.yol.com/win88/dbproxy/mongo"
	"games.yol.com/win88/model"
	"github.com/globalsign/mgo"
	"github.com/idealeak/goserver/core/logger"
)

var (
	SceneCoinLogErr = errors.New("log_scenecoin db open failed.")
)

func SceneCoinLogsCollection(plt string) *mongo.Collection {
	s := mongo.MgoSessionMgrSington.GetPltMgoSession(plt, model.SceneCoinLogDBName)
	if s != nil {
		c_scenecoinlogrec, first := s.DB().C(model.SceneCoinLogCollName)
		if first {
			c_scenecoinlogrec.EnsureIndex(mgo.Index{Key: []string{"snid", "-time"}, Background: true, Sparse: true})
			c_scenecoinlogrec.EnsureIndex(mgo.Index{Key: []string{"gameidex"}, Background: true, Sparse: true})
			c_scenecoinlogrec.EnsureIndex(mgo.Index{Key: []string{"gameid"}, Background: true, Sparse: true})
			c_scenecoinlogrec.EnsureIndex(mgo.Index{Key: []string{"sceneid"}, Background: true, Sparse: true})
			c_scenecoinlogrec.EnsureIndex(mgo.Index{Key: []string{"gamemode"}, Background: true, Sparse: true})
			c_scenecoinlogrec.EnsureIndex(mgo.Index{Key: []string{"gamefreeid"}, Background: true, Sparse: true})
			c_scenecoinlogrec.EnsureIndex(mgo.Index{Key: []string{"ts"}, Background: true, Sparse: true})
			c_scenecoinlogrec.EnsureIndex(mgo.Index{Key: []string{"time"}, Background: true, Sparse: true})
		}
		return c_scenecoinlogrec
	}
	return nil
}

type SceneCoinLogSvc struct {
}

func (svc *SceneCoinLogSvc) InsertSceneCoinLog(log *model.SceneCoinLog, ret *bool) (err error) {
	clog := SceneCoinLogsCollection(log.Platform)
	if clog == nil {
		return
	}

	err = clog.Insert(log)
	if err != nil {
		logger.Logger.Warn("InsertSceneCoinLog error:", err)
		return
	}
	*ret = true
	return
}

func init() {
	rpc.Register(new(SceneCoinLogSvc))
}
