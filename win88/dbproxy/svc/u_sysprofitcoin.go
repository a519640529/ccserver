package svc

import (
	"errors"
	"net/rpc"

	"games.yol.com/win88/dbproxy/mongo"
	"games.yol.com/win88/model"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/idealeak/goserver/core/logger"
)

var (
	SysProfitCoinDBErr = errors.New("user_sysprofitcoin db open failed.")
)

func SysProfitCoinCollection() *mongo.Collection {
	s := mongo.MgoSessionMgrSington.GetPltMgoSession(mongo.G_P, model.SysProfitCoinDBName)
	if s != nil {
		c, first := s.DB().C(model.SysProfitCoinCollName)
		if first {
			c.EnsureIndex(mgo.Index{Key: []string{"key"}, Background: true, Sparse: true})
		}
		return c
	}
	return nil
}

func InitSysProfitCoinData(key string) *model.SysProfitCoin {
	s := SysProfitCoinCollection()
	data := &model.SysProfitCoin{}
	if s != nil {
		err := s.Find(bson.M{"key": key}).One(data)
		if err != nil {
			if err.Error() == mgo.ErrNotFound.Error() {
				data.LogId = bson.NewObjectId()
				data.Key = key
				data.ProfitCoin = make(map[string]*model.SysCoin)
				s.Insert(data)
			} else {
				logger.Logger.Trace("InitSysProfitCoinData err:", err)
				return nil
			}
		}
		return data
	}
	return nil
}

//保存
func SaveSysProfitCoin(data *model.SysProfitCoin) error {
	s := SysProfitCoinCollection()
	if s == nil {
		return SysProfitCoinDBErr
	}

	_, err := s.Upsert(bson.M{"_id": data.LogId}, data)
	return err
}

type SysProfitCoinSvc struct {
}

func (svc *SysProfitCoinSvc) InitSysProfitCoinData(key string, ret **model.SysProfitCoin) (err error) {
	*ret = InitSysProfitCoinData(key)
	return nil
}

func (svc *SysProfitCoinSvc) SaveSysProfitCoin(args *model.SysProfitCoin, ret *bool) (err error) {
	err = SaveSysProfitCoin(args)
	if err == nil {
		*ret = true
	}
	return
}

func init() {
	rpc.Register(new(SysProfitCoinSvc))
}
