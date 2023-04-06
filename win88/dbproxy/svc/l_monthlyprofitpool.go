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
	MonthlyProfitPoolDBErr = errors.New("log_monthlyprofitpool open failed.")
)

func MonthlyProfitPoolCollection() *mongo.Collection {
	s := mongo.MgoSessionMgrSington.GetPltMgoSession(mongo.G_P, model.MonthlyProfitPoolDBName)
	if s != nil {
		c, first := s.DB().C(model.MonthlyProfitPoolCollName)
		if first {
			c.EnsureIndex(mgo.Index{Key: []string{"serverid"}, Background: true, Sparse: true})
			c.EnsureIndex(mgo.Index{Key: []string{"gamefreeid"}, Background: true, Sparse: true})
			c.EnsureIndex(mgo.Index{Key: []string{"platform"}, Background: true, Sparse: true})
			c.EnsureIndex(mgo.Index{Key: []string{"time"}, Background: true, Sparse: true})
		}
		return c
	}
	return nil
}

func InsertMonthlyProfitPool(logs ...*model.MonthlyProfitPool) (err error) {
	clog := MonthlyProfitPoolCollection()
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
		logger.Logger.Warn("InsertMonthlyProfitPool error:", err)
		return
	}
	return
}
func InsertSignleMonthlyProfitPool(log *model.MonthlyProfitPool) (err error) {
	clog := MonthlyProfitPoolCollection()
	if clog == nil {
		return
	}
	err = clog.Insert(log)
	if err != nil {
		logger.Logger.Warn("InsertSignleMonthlyProfitPool error:", err)
		return
	}
	return
}

func RemoveMonthlyProfitPool(ts time.Time) (*mgo.ChangeInfo, error) {
	return MonthlyProfitPoolCollection().RemoveAll(bson.M{"time": bson.M{"$lt": ts}})
}

type MonthlyProfitPoolSvc struct {
}

func (svc *MonthlyProfitPoolSvc) InsertMonthlyProfitPool(logs []*model.MonthlyProfitPool, ret *bool) (err error) {
	err = InsertMonthlyProfitPool(logs...)
	if err == nil {
		*ret = true
	}
	return
}

func (svc *MonthlyProfitPoolSvc) InsertSignleMonthlyProfitPool(log *model.MonthlyProfitPool, ret *bool) (err error) {
	err = InsertSignleMonthlyProfitPool(log)
	if err == nil {
		*ret = true
	}
	return
}

func (svc *MonthlyProfitPoolSvc) RemoveMonthlyProfitPool(ts time.Time, ret **mgo.ChangeInfo) (err error) {
	*ret, err = RemoveMonthlyProfitPool(ts)
	return
}

func init() {
	rpc.Register(new(MonthlyProfitPoolSvc))
}
