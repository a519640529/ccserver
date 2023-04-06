package svc

import (
	"net/rpc"

	"games.yol.com/win88/dbproxy/mongo"
	"games.yol.com/win88/model"
	"github.com/globalsign/mgo"
	"github.com/idealeak/goserver/core/logger"
)

func LoginLogsCollection(plt string) *mongo.Collection {
	s := mongo.MgoSessionMgrSington.GetPltMgoSession(plt, model.LoginLogDBName)
	if s != nil {
		c_loginlogrec, first := s.DB().C(model.LoginLogCollName)
		if first {
			c_loginlogrec.EnsureIndex(mgo.Index{Key: []string{"snid"}, Background: true, Sparse: true})
			c_loginlogrec.EnsureIndex(mgo.Index{Key: []string{"channel"}, Background: true, Sparse: true})
			c_loginlogrec.EnsureIndex(mgo.Index{Key: []string{"promoter"}, Background: true, Sparse: true})
			c_loginlogrec.EnsureIndex(mgo.Index{Key: []string{"logtype"}, Background: true, Sparse: true})
			c_loginlogrec.EnsureIndex(mgo.Index{Key: []string{"ip"}, Background: true, Sparse: true})
			c_loginlogrec.EnsureIndex(mgo.Index{Key: []string{"time"}, Background: true, Sparse: true})
		}
		return c_loginlogrec
	}
	return nil
}

type LoginsLogSvc struct {
}

func (svc *LoginsLogSvc) InsertLoginLogs(logs []*model.LoginLog, ret *bool) (err error) {
	for _, log := range logs {
		err = svc.InsertSignleLoginLog(log, ret)
		if err != nil {
			return
		}
	}
	*ret = true
	return
}

func (svc *LoginsLogSvc) InsertSignleLoginLog(log *model.LoginLog, ret *bool) (err error) {
	clog := LoginLogsCollection(log.Platform)
	if clog == nil {
		return
	}
	err = clog.Insert(log)
	if err != nil {
		logger.Logger.Warn("svc.InsertSignleLoginLog error:", err)
		return
	}
	*ret = true
	return
}

func init() {
	rpc.Register(new(LoginsLogSvc))
}
