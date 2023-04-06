package svc

import (
	"encoding/json"
	"errors"
	"net/rpc"
	"time"

	"games.yol.com/win88/dbproxy/mongo"
	"games.yol.com/win88/model"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/idealeak/goserver/core/logger"
)

const SafeBoxMaxLimit = 50

var SafeBoxLogErr = errors.New("log_safeboxrec db open failed.")

func SafeBoxCollection(plt string) *mongo.Collection {
	s := mongo.MgoSessionMgrSington.GetPltMgoSession(plt, model.SafeBoxDBName)
	if s != nil {
		c, first := s.DB().C(model.SafeBoxCollName)
		if first {
			c.EnsureIndex(mgo.Index{Key: []string{"userid"}, Background: true, Sparse: true})
			c.EnsureIndex(mgo.Index{Key: []string{"channel"}, Background: true, Sparse: true})
			c.EnsureIndex(mgo.Index{Key: []string{"promoter"}, Background: true, Sparse: true})
		}
		return c
	}
	return nil
}

func InsertSafeBox(log *model.SafeBoxRec) error {
	c_safeboxrec := SafeBoxCollection(log.Platform)
	if c_safeboxrec == nil {
		return SafeBoxLogErr
	}
	err := c_safeboxrec.Insert(log)
	if err != nil {
		buff, _ := json.Marshal(log)
		logger.Logger.Error("InsertSafeBoxCoinLog param:", string(buff))
		logger.Logger.Error("InsertSafeBoxCoinLog error:", err)
		return err
	}
	return nil
}

func GetSafeBoxs(plt string, userid int32) (recs []model.SafeBoxRec, err error) {
	csafeboxrec := SafeBoxCollection(plt)
	if csafeboxrec == nil {
		return nil, SafeBoxLogErr
	}

	logger.Logger.Trace("GetSafeBoxs:", userid)
	err = csafeboxrec.Find(bson.M{"userid": userid}).Sort("-time").Limit(SafeBoxMaxLimit).All(&recs)
	return
}

func RemoveSafeBoxs(plt string, ts time.Time) (*mgo.ChangeInfo, error) {
	return SafeBoxCollection(plt).RemoveAll(bson.M{"time": bson.M{"$lt": ts}})
}

func RemoveSafeBoxCoinLog(plt string, id bson.ObjectId) error {
	sbc := SafeBoxCollection(plt)
	if sbc == nil {
		return SafeBoxLogErr
	}
	return sbc.RemoveId(id)
}

func GetSafeBoxCoinLog(plt string, ts time.Time) (recs []model.SafeBoxRec, err error) {
	csafeboxrec := SafeBoxCollection(plt)
	if csafeboxrec == nil {
		return nil, SafeBoxLogErr
	}

	err = csafeboxrec.Find(bson.M{"time": bson.M{"$gt": ts}}).All(&recs)
	return
}

type SafeBoxRecSvc struct {
}

func (svc *SafeBoxRecSvc) InsertSafeBox(log *model.SafeBoxRec, ret *bool) (err error) {
	err = InsertSafeBox(log)
	if err != nil {
		return err
	}
	*ret = true
	return nil
}

func (svc *SafeBoxRecSvc) GetSafeBoxs(args *model.GetSafeBoxsArgs, ret *[]model.SafeBoxRec) (err error) {
	*ret, err = GetSafeBoxs(args.Plt, args.SnId)
	if err != nil {
		return err
	}
	return nil
}

func (svc *SafeBoxRecSvc) RemoveSafeBoxs(args *model.RemoveSafeBoxsArgs, ret **mgo.ChangeInfo) (err error) {
	*ret, err = RemoveSafeBoxs(args.Plt, args.Ts)
	return
}

func (svc *SafeBoxRecSvc) RemoveSafeBoxCoinLog(args *model.RemoveSafeBoxCoinLogArgs, ret *bool) (err error) {
	err = RemoveSafeBoxCoinLog(args.Plt, args.Id)
	if err != nil {
		return err
	}
	*ret = true
	return
}

func (svc *SafeBoxRecSvc) GetSafeBoxCoinLog(args *model.GetSafeBoxCoinLogArgs, ret *[]model.SafeBoxRec) (err error) {
	*ret, err = GetSafeBoxCoinLog(args.Plt, args.Ts)
	if err != nil {
		return err
	}
	return
}

func init() {
	rpc.Register(new(SafeBoxRecSvc))
}
