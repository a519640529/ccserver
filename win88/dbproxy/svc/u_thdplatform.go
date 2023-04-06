package svc

import (
	"games.yol.com/win88/dbproxy/mongo"
	"games.yol.com/win88/model"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/idealeak/goserver/core/logger"
	"github.com/wendal/errors"
	"net/rpc"
	"time"
)

var (
	PlatformOfThirdPlatformDBErr = errors.New("user_thdplatform open failed.")
)

func ThdPlatformCollection() *mongo.Collection {
	s := mongo.MgoSessionMgrSington.GetPltMgoSession(mongo.G_P, model.ThdPlatformDBName)
	if s != nil {
		c, first := s.DB().C(model.ThdPlatformCollName)
		if first {
			c.EnsureIndex(mgo.Index{Key: []string{"platform"}, Background: true, Sparse: true})
		}
		return c
	}
	return nil
}

func InsertThirdPlatform(platforms ...*model.PlatformOfThirdPlatform) (err error) {
	switch len(platforms) {
	case 0:
		return errors.New("no data")
	case 1:
		err = ThdPlatformCollection().Insert(platforms[0])
	default:
		docs := make([]interface{}, 0, len(platforms))
		for _, p := range platforms {
			docs = append(docs, p)
		}
		err = ThdPlatformCollection().Insert(docs...)
	}
	if err != nil {
		logger.Logger.Warn("InsertThirdPlatform error:", err)
		return
	}
	return
}

func UpdateThirdPlatform(platform *model.PlatformOfThirdPlatform) (err error) {
	platform.LastTime = time.Now()
	err = ThdPlatformCollection().Update(bson.M{"_id": platform.Id}, platform)
	if err != nil {
		logger.Logger.Info("UpdateThirdPlatform to db failed.")
		return err
	}
	return nil
}

func GetAllThirdPlatform() (ret []model.PlatformOfThirdPlatform, err error) {
	err = ThdPlatformCollection().Find(bson.M{}).All(&ret)
	return
}

type PlatformOfThirdPlatformSvc struct {
}

func (svc *PlatformOfThirdPlatformSvc) InsertThirdPlatform(args []*model.PlatformOfThirdPlatform, ret *bool) (err error) {
	err = InsertThirdPlatform(args...)
	if err == nil {
		*ret = true
	}
	return
}

func (svc *PlatformOfThirdPlatformSvc) UpdateThirdPlatform(args *model.PlatformOfThirdPlatform, ret *bool) (err error) {
	err = UpdateThirdPlatform(args)
	if err == nil {
		*ret = true
	}
	return
}

func (svc *PlatformOfThirdPlatformSvc) GetAllThirdPlatform(args struct{}, ret *[]model.PlatformOfThirdPlatform) (err error) {
	*ret, err = GetAllThirdPlatform()
	return
}

func init() {
	rpc.Register(new(PlatformOfThirdPlatformSvc))
}
