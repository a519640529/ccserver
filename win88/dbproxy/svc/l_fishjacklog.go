package svc

import (
	"errors"
	"games.yol.com/win88/dbproxy/mongo"
	"games.yol.com/win88/model"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"net/rpc"
)

var (
	FishJackpotLogErr              = errors.New("log fishjackpot log open failed.")
)

type FishJackLogSvc struct{}

func init() {
	rpc.Register(new(FishJackLogSvc))
}

func FishJackpotLogsCollection(plt string) *mongo.Collection {
	s := mongo.MgoSessionMgrSington.GetPltMgoSession(plt, model.FishJackpotLogDBName)
	if s != nil {
		c, first := s.DB().C(model.FishJackpotLogCollName)
		if first {
			c.EnsureIndex(mgo.Index{Key: []string{"snid"}, Background: true, Sparse: true})
			c.EnsureIndex(mgo.Index{Key: []string{"channel"}, Background: true, Sparse: true})
			c.EnsureIndex(mgo.Index{Key: []string{"promoter"}, Background: true, Sparse: true})
			c.EnsureIndex(mgo.Index{Key: []string{"ts"}, Background: true, Sparse: true})
			c.EnsureIndex(mgo.Index{Key: []string{"time"}, Background: true, Sparse: true})
			c.EnsureIndex(mgo.Index{Key: []string{"roomid"}, Background: true, Sparse: true})
		}
		return c
	}
	return nil
}

func (svc *FishJackLogSvc) GetFishJackpotLogByPlatform(args *model.GetLogArgs, ret *model.GetLogRet) error {
	c := FishJackpotLogsCollection(args.Platform)
	if c == nil {
		return FishJackpotLogErr
	}
	err := c.Find(bson.M{"platform": args.Platform}).Sort("-ts").Limit(model.FishJackpotLogMaxLimitPerQuery).All(&ret.Logs)
	return err
}

func (svc *FishJackLogSvc) InsertSignleFishJackpotLog(args *model.InsertLogArgs, ret *model.InsertLogRet) error {
	c := FishJackpotLogsCollection(args.Log.Platform)
	if c == nil {
		return FishJackpotLogErr
	}
	err := c.Insert(args.Log)
	return err
}
