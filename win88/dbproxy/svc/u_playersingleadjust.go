package svc

import (
	"errors"
	"games.yol.com/win88/dbproxy/mongo"
	"games.yol.com/win88/model"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/idealeak/goserver/core/logger"
	"net/rpc"
)

//单控数据表
var (
	SingleAdjustDBName   = "user"
	SingleAdjustCollName = "user_singleadjust"
	SingleAdjustColError = errors.New("SingleAdjust collection open failed")
)

func SingleAdjustCollection(plt string) *mongo.Collection {
	s := mongo.MgoSessionMgrSington.GetPltMgoSession(plt, SingleAdjustDBName)
	if s != nil {
		c_sj, first := s.DB().C(SingleAdjustCollName)
		if first {
			c_sj.EnsureIndex(mgo.Index{Key: []string{"platform", "snid", "gamefreeid"}, Unique: true, Background: true, Sparse: true})
			c_sj.EnsureIndex(mgo.Index{Key: []string{"platform"}, Background: true, Sparse: true})
			c_sj.EnsureIndex(mgo.Index{Key: []string{"snid"}, Background: true, Sparse: true})
			c_sj.EnsureIndex(mgo.Index{Key: []string{"gamefreeid"}, Background: true, Sparse: true})
		}
		return c_sj
	}

	return nil
}

type SingleAdjustSvc struct {
}

func (svc *SingleAdjustSvc) QueryAllSingleAdjust(platform string, ret *model.SingleAdjustRet) error {
	c_sj := SingleAdjustCollection(platform)
	if c_sj == nil {
		return SingleAdjustColError
	}
	err := c_sj.Find(nil).All(&ret.Ret)
	if err != nil {
		return err
	}
	return nil
}
func (svc *SingleAdjustSvc) QueryAllSingleAdjustByKey(args *model.SingleAdjustByKey, ret *model.SingleAdjustRet) error {
	c_sj := SingleAdjustCollection(args.Platform)
	if c_sj == nil {
		return SingleAdjustColError
	}
	err := c_sj.Find(bson.M{"platform": args.Platform, "snid": args.SnId}).All(&ret.Ret)
	if err != nil {
		logger.Logger.Warn("QueryAllSingleAdjustByKey is err: ", err)
		return err
	}
	return nil
}
func (svc *SingleAdjustSvc) AddNewSingleAdjust(args *model.PlayerSingleAdjust, ret *bool) error {
	c_sj := SingleAdjustCollection(args.Platform)
	if c_sj == nil {
		return SingleAdjustColError
	}
	err := c_sj.Insert(args)
	if err != nil {
		return err
	}
	*ret = true
	return nil
}
func (svc *SingleAdjustSvc) EditSingleAdjust(args *model.PlayerSingleAdjust, ret *bool) error {
	c_sj := SingleAdjustCollection(args.Platform)
	if c_sj == nil {
		return SingleAdjustColError
	}
	err := c_sj.Update(bson.M{"platform": args.Platform, "snid": args.SnId, "gamefreeid": args.GameFreeId}, args)
	if err != nil {
		return err
	}
	*ret = true
	return nil
}
func (svc *SingleAdjustSvc) DeleteSingleAdjust(args *model.PlayerSingleAdjust, ret *bool) error {
	c_sj := SingleAdjustCollection(args.Platform)
	if c_sj == nil {
		return SingleAdjustColError
	}
	err := c_sj.Remove(bson.M{"platform": args.Platform, "snid": args.SnId, "gamefreeid": args.GameFreeId})
	if err != nil {
		return err
	}
	*ret = true
	return nil
}
func init() {
	rpc.Register(&SingleAdjustSvc{})
}
