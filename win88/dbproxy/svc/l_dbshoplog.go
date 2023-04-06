package svc

import (
	"games.yol.com/win88/dbproxy/mongo"
	"games.yol.com/win88/model"
	"github.com/globalsign/mgo"
	"github.com/idealeak/goserver/core/logger"
	"net/rpc"
)

func DbShopLogCollection(plt string) *mongo.Collection {
	s := mongo.MgoSessionMgrSington.GetPltMgoSession(plt, model.DbShopDBName)
	if s != nil {
		dbShopRec, first := s.DB().C(model.DbShopCollName)
		if first {
			dbShopRec.EnsureIndex(mgo.Index{Key: []string{"snid"}, Background: true, Sparse: true})
			dbShopRec.EnsureIndex(mgo.Index{Key: []string{"state"}, Background: true, Sparse: true})
			dbShopRec.EnsureIndex(mgo.Index{Key: []string{"shopid"}, Background: true, Sparse: true})
			dbShopRec.EnsureIndex(mgo.Index{Key: []string{"createts"}, Background: true, Sparse: true})
			dbShopRec.EnsureIndex(mgo.Index{Key: []string{"opts"}, Background: true, Sparse: true})
		}
		return dbShopRec
	}
	return nil
}

type DbShopLogSvc struct {
}

func (svc *DbShopLogSvc) InsertDbShopLog(args *model.DbShopLogArgs, ret *bool) (err error) {
	clog := DbShopLogCollection(args.Platform)
	if clog == nil {
		return
	}
	logger.Logger.Trace("DbShopLogSvc.InsertDbShopLog")
	err = clog.Insert(args.Log)
	if err != nil {
		logger.Logger.Warn("DbShopLogSvc.InsertDbShopLog error:", err)
		return
	}
	*ret = true
	return
}
func init() {
	rpc.Register(new(DbShopLogSvc))
}
