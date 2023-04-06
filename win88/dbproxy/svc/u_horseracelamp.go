package svc

import (
	"errors"
	"games.yol.com/win88/dbproxy/mongo"
	"games.yol.com/win88/model"
	"github.com/globalsign/mgo/bson"
	"github.com/idealeak/goserver/core/logger"
	"net/rpc"
)

var ErrHorseRaceLampDBNotOpen = model.NewDBError(model.HorseRaceLampDBName, model.HorseRaceLampCollName, model.NOT_OPEN)

func HorseRaceLampCollection(plt string) *mongo.Collection {
	s := mongo.MgoSessionMgrSington.GetPltMgoSession(plt, model.HorseRaceLampDBName)
	if s != nil {
		c, first := s.DB().C(model.HorseRaceLampCollName)
		if first {
			//	c.EnsureIndex(mgo.Index{Key: []string{"username", "tagkey"}, Background: true, Sparse: true})
			//	c.EnsureIndex(mgo.Index{Key: []string{"tel", "tagkey"}, Background: true, Sparse: true})
		}
		return c
	}
	return nil
}

type HorseRaceLampSvc struct {
}

var horseRaceLampSvc = &HorseRaceLampSvc{}

func (svc *HorseRaceLampSvc) GetAllHorseRaceLamp(args string, ret *[]model.HorseRaceLamp) error {
	logger.Logger.Info("HorseRaceLampSvc GetAllHorseRaceLamp")

	c := HorseRaceLampCollection(args)
	if c != nil {
		err := c.Find(bson.M{}).All(ret)
		if err != nil {
			return err
		}
	}
	return nil
}
func (svc *HorseRaceLampSvc) GetHorseRaceLamp(args *model.GetHorseRaceLampArgs, ret *model.HorseRaceLamp) (err error) {
	c := HorseRaceLampCollection(args.Platform)
	if c == nil {
		return ErrHorseRaceLampDBNotOpen
	}
	err = c.Find(bson.M{"_id": args.Id}).One(ret)
	return
}
func (svc *HorseRaceLampSvc) RemoveHorseRaceLamp(args *model.RemoveHorseRaceLampArg, ret *bool) error {
	logger.Logger.Info("HorseRaceLampSvc RemoveHorseRaceLamp")
	c := HorseRaceLampCollection(args.Platform)
	if c != nil {
		_, err := c.RemoveAll(bson.M{"_id": bson.ObjectIdHex(args.Key)})
		if err != nil {
			return err
		}
		*ret = true
	}
	return nil
}
func (svc *HorseRaceLampSvc) InsertHorseRaceLamp(args *model.InsertHorseRaceLampArgs, ret *bool) (err error) {
	c := HorseRaceLampCollection(args.Platform)
	if c == nil {
		return ErrHorseRaceLampDBNotOpen
	}
	switch len(args.HorseRaceLamps) {
	case 1:
		err = c.Insert(args.HorseRaceLamps[0])
	default:
		docs := make([]interface{}, 0, len(args.HorseRaceLamps))
		for _, notice := range args.HorseRaceLamps {
			docs = append(docs, notice)
		}
		err = c.Insert(docs...)
	}
	if err != nil {
		logger.Logger.Error("InsertHorseRaceLamps error:", err)
		return
	}
	*ret = true
	return
}
func (svc *HorseRaceLampSvc) EditHorseRaceLamp(args *model.EditHorseRaceLampArg, ret *bool) error {
	logger.Logger.Info("HorseRaceLampSvc EditHorseRaceLamp")
	c := HorseRaceLampCollection(args.Platform)
	if c != nil {
		if !bson.IsObjectIdHex(args.Key) {
			return errors.New("key is invalid bson.ObjectId")
		}
		Id := bson.ObjectIdHex(args.Key)
		_, err := c.Upsert(bson.M{"_id": Id}, bson.D{{"$set", bson.D{{"channel", args.Channel},
			{"title", args.Title}, {"content", args.Content}, {"footer", args.Footer},
			{"starttime", args.StartTime}, {"interval", args.Interval}, {"count", args.Count},
			{"priority", args.Priority}, {"msgtype", args.MsgType}, {"platform", args.Platform}, {"state", args.State},
			{"target", args.Target}, {"standsec", args.StandSec}}}})
		if err != nil {
			logger.Logger.Warn("EditHorseRaceLamp to db failed.", err)
			return err
		}
		*ret = true
	}
	return nil
}
func (svc *HorseRaceLampSvc) GetHorseRaceLampInRangeTsLimitByRange(args *model.QueryHorseRaceLampArg, ret *model.QueryHorseRaceLampRet) (err error) {
	logger.Logger.Info("HorseRaceLampSvc EditHorseRaceLamp")
	c := HorseRaceLampCollection(args.Platform)
	if c == nil {
		return ErrHorseRaceLampDBNotOpen
	}
	if len(args.Platform) == 0 {
		ret.Count, _ = c.Find(bson.M{"state": args.State, "msgtype": args.MsgType}).Count()
	} else {
		ret.Count, _ = c.Find(bson.M{"platform": args.Platform, "state": args.State, "msgtype": args.MsgType}).Count()
	}
	if len(args.Platform) == 0 {
		err = c.Find(bson.M{"state": args.State, "msgtype": args.MsgType}).Sort("createtime").Skip(args.FromIndex).Limit(args.LimitDataNum).All(&ret.Data)
	} else {
		selector := bson.M{"platform": args.Platform, "state": args.State, "msgtype": args.MsgType}
		err = c.Find(selector).Sort("createtime").Skip(args.FromIndex).Limit(args.LimitDataNum).All(&ret.Data)
	}
	return
}
func init() {
	rpc.Register(horseRaceLampSvc)
}
