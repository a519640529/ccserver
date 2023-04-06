package svc

import (
	"games.yol.com/win88/dbproxy/mongo"
	"games.yol.com/win88/model"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/idealeak/goserver/core/logger"
	"net/rpc"
)

var (
	GameKVDataDBName   = "user"
	GameKVDataCollName = "user_gamekvdata"
)

func GameKVDatasCollection() *mongo.Collection {
	s := mongo.MgoSessionMgrSington.GetPltMgoSession(mongo.G_P, GameKVDataDBName)
	if s != nil {
		c, first := s.DB().C(GameKVDataCollName)
		if first {
			c.EnsureIndex(mgo.Index{Key: []string{"key"}, Unique: true, Background: true, Sparse: true})
		}
		return c
	}
	return nil
}

type GameKVDataSvc struct {
}

func (svc *GameKVDataSvc) GetAllGameKVData(dummy struct{}, ret *[]*model.GameKVData) error {
	c := GameKVDatasCollection()
	if c != nil {
		err := c.Find(nil).All(ret)
		if err != nil {
			logger.Logger.Trace("GetAllGameKVData err:", err)
			return err
		}
	}
	return nil
}

func (svc *GameKVDataSvc) UptGameKVData(args *model.GameKVData, ret *bool) error {
	c := GameKVDatasCollection()
	if c != nil {
		info, err := c.Upsert(bson.M{"key": args.Key}, args)
		logger.Logger.Trace("UptGameKVData :", info, err)
		return err
	}

	*ret = true
	return nil
}

var _GameKVDataSvc = &GameKVDataSvc{}

func init() {
	rpc.Register(_GameKVDataSvc)
}
