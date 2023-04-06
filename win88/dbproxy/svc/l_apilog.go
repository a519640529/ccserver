package svc

import (
	"net/rpc"
	"time"

	"games.yol.com/win88/dbproxy/mongo"
	"games.yol.com/win88/model"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/idealeak/goserver/core/logger"
)

func APILogsCollection() *mongo.Collection {
	s := mongo.MgoSessionMgrSington.GetGlobal(model.APILogDBName)
	if s != nil {
		c_apilogrec, first := s.DB().C(model.APILogCollName)
		if first {
			c_apilogrec.EnsureIndex(mgo.Index{Key: []string{"path"}, Background: true, Sparse: true})
			c_apilogrec.EnsureIndex(mgo.Index{Key: []string{"time"}, Background: true, Sparse: true})
		}
		return c_apilogrec
	}
	return nil
}

type APILogSvc struct {
}

func (svc *APILogSvc) InsertAPILog(log *model.APILog, ret *bool) (err error) {
	clog := APILogsCollection()
	if clog == nil {
		return
	}
	err = clog.Insert(log)
	if err != nil {
		logger.Logger.Warn("InsertAPILog error:", err)
		return
	}
	*ret = true
	return
}

func (svc *APILogSvc) RemoveAPILog(ts int64, chged **mgo.ChangeInfo) (err error) {
	*chged, err = APILogsCollection().RemoveAll(bson.M{"time": bson.M{"$lte": time.Unix(ts, 0)}})
	return
}

func init() {
	rpc.Register(new(APILogSvc))
}
