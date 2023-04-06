package svc

import (
	"errors"
	"games.yol.com/win88/dbproxy/mongo"
	"games.yol.com/win88/model"
	"github.com/globalsign/mgo"
	"github.com/idealeak/goserver/core/logger"
	"net/rpc"
)

var (
	LotteryDBErr = errors.New("user_coinlog db open failed.")
)

func LotteryCollection() *mongo.Collection {
	s := mongo.MgoSessionMgrSington.GetPltMgoSession(mongo.G_P, model.LotteryDBName)
	if s != nil {
		c, first := s.DB().C(model.LotteryCollName)
		if first {
			c.EnsureIndex(mgo.Index{Key: []string{"platform"}, Background: true, Sparse: true})
			c.EnsureIndex(mgo.Index{Key: []string{"gamefreeid"}, Background: true, Sparse: true})
		}
		return c
	}
	return nil
}

func GetAllLottery() (ret []model.Lottery, err error) {
	err = LotteryCollection().Find(nil).All(&ret)
	return
}

func UpsertLottery(item *model.Lottery) (err error) {
	c := LotteryCollection()
	if c == nil {
		return
	}
	_, err = c.UpsertId(item.Id, item)
	if err != nil {
		logger.Logger.Warn("UpsertLottery error:", err)
		return
	}
	return
}

type LotterySvc struct {
}

func (svc *LotterySvc) GetAllLottery(args struct{}, ret *[]model.Lottery) (err error) {
	*ret, err = GetAllLottery()
	return
}

func (svc *LotterySvc) UpsertLottery(args *model.Lottery, ret *bool) (err error) {
	err = UpsertLottery(args)
	if err == nil {
		*ret = true
	}
	return
}

func init() {
	rpc.Register(new(LotterySvc))
}
