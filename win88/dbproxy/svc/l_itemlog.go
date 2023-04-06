package svc

import (
	"games.yol.com/win88/dbproxy/mongo"
	"games.yol.com/win88/model"
	"github.com/globalsign/mgo"
	"github.com/idealeak/goserver/core/logger"
	"net/rpc"
)

func ItemLogsCollection(plt string) *mongo.Collection {
	s := mongo.MgoSessionMgrSington.GetPltMgoSession(plt, model.ItemLogDBName)
	if s != nil {
		c_itemlog, first := s.DB().C(model.ItemLogCollName)
		if first {
			c_itemlog.EnsureIndex(mgo.Index{Key: []string{"snid"}, Background: true, Sparse: true})
			c_itemlog.EnsureIndex(mgo.Index{Key: []string{"platform"}, Background: true, Sparse: true})
		}
		return c_itemlog
	}
	return nil
}

type ItemLogSvc struct {
}

func (svc *ItemLogSvc) InsertItemLog(log *model.ItemLog, ret *bool) (err error) {
	clog := ItemLogsCollection(log.Platform)
	if clog == nil {
		return
	}
	err = clog.Insert(log)
	if err != nil {
		logger.Logger.Warn("InsertItemLog error:", err)
		return
	}
	*ret = true
	return
}

func init() {
	rpc.Register(new(ItemLogSvc))
}
