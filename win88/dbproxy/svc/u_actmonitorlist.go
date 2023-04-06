package svc

import (
	"errors"
	"games.yol.com/win88/model"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/mongo"
	"net/rpc"
)

var (
	c_actmonitorrec    *mongo.Collection
	ActMonitorDBName   = "user"
	ActMonitorCollName = "user_actmonitorlist"
	ActMonitorErr      = errors.New("user_actmonitorlist log open failed.")
)

func ActMonitorCollection() *mongo.Collection {
	if c_actmonitorrec == nil || !c_actmonitorrec.IsValid() {
		c_actmonitorrec = mongo.DatabaseC(ActMonitorDBName, ActMonitorCollName)
		if c_actmonitorrec != nil {
			c_actmonitorrec.Hold()
			c_actmonitorrec.EnsureIndex(mgo.Index{Key: []string{"seqno"}, Background: true, Sparse: true})
			c_actmonitorrec.EnsureIndex(mgo.Index{Key: []string{"snid"}, Background: true, Sparse: true})
			c_actmonitorrec.EnsureIndex(mgo.Index{Key: []string{"createtime"}, Background: true, Sparse: true})
			c_actmonitorrec.EnsureIndex(mgo.Index{Key: []string{"platform"}, Background: true, Sparse: true})
		}
	}
	return c_actmonitorrec
}

type ActMonitorSvc struct {
}

func (svc *ActMonitorSvc) GetAllActMonitorData(id int, ret *model.ActMonitoRet) error {
	cat := ActMonitorCollection()
	if cat == nil {
		return nil
	}
	err := cat.Find(bson.M{}).All(&ret.Data)
	if err != nil {
		logger.Logger.Error("Get model.ActMonitor data eror.", err)
		ret.Err = err
		return err
	}
	return nil
}
func (svc *ActMonitorSvc) UpsertSignleActMonitor(am *model.ActMonitor, ret *model.ActMonitoRet) (err error) {
	cat := ActMonitorCollection()
	if cat == nil {
		return errors.New("svc.ActMonitor is nil")
	}
	_, err = cat.Upsert(bson.M{"seqno": am.SeqNo}, am)
	if err != nil {
		logger.Logger.Warn("svc.UpsertSignleActMonitor error:", err)
		ret.Err = err
		return
	}
	return
}

func (svc *ActMonitorSvc) UpdateSignleActMonitor(am *model.ActMonitor, ret *model.ActMonitoRet) (err error) {
	cat := ActMonitorCollection()
	if cat == nil {
		return errors.New("model.ActMonitor is nil")
	}
	old := new(model.ActMonitor)
	if err = cat.Find(bson.M{"seqno": am.SeqNo}).One(old); err != nil {
		logger.Logger.Error("FindSignleActMonitor error:", err)
		ret.Err = err
		return
	}
	am.CreateTime = old.CreateTime
	if err = cat.Update(bson.M{"seqno": am.SeqNo}, am); err != nil {
		logger.Logger.Error("UpdateSignleActMonitor error:", err)
		ret.Err = err
	}
	return
}

func (svc *ActMonitorSvc) RemoveActMonitorOne(seqno int, ret *model.ActMonitoRet) error {
	cat := ActMonitorCollection()
	if cat == nil {
		return ActMonitorErr
	}
	ret.Err = cat.Remove(bson.M{"seqno": seqno})
	if ret.Err != nil {
		logger.Logger.Error("srv.RemoveActMonitorOne error", ret.Err)
		return ret.Err
	}
	return nil
}
func init() {
	rpc.Register(&ActMonitorSvc{})
}
