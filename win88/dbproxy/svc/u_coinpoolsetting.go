package svc

import (
	"net/rpc"
	"time"

	"games.yol.com/win88/dbproxy/mongo"
	"games.yol.com/win88/model"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/idealeak/goserver/core/logger"
)

func CoinPoolSettingCollection() *mongo.Collection {
	s := mongo.MgoSessionMgrSington.GetPltMgoSession(mongo.G_P, model.CoinPoolSettingDBName)
	if s != nil {
		c, first := s.DB().C(model.CoinPoolSettingCollName)
		if first {
			c.EnsureIndex(mgo.Index{Key: []string{"platform"}, Background: true, Sparse: true})
			c.EnsureIndex(mgo.Index{Key: []string{"gamefreeid"}, Background: true, Sparse: true})
			c.EnsureIndex(mgo.Index{Key: []string{"serverid"}, Background: true, Sparse: true})
		}
		return c
	}
	return nil
}

func CoinPoolSettingHisCollection() *mongo.Collection {
	s := mongo.MgoSessionMgrSington.GetPltMgoSession(mongo.G_P, model.CoinPoolSettingHisDBName)
	if s != nil {
		c, first := s.DB().C(model.CoinPoolSettingHisCollName)
		if first {
			c.EnsureIndex(mgo.Index{Key: []string{"platform"}, Background: true, Sparse: true})
			c.EnsureIndex(mgo.Index{Key: []string{"gamefreeid"}, Background: true, Sparse: true})
			c.EnsureIndex(mgo.Index{Key: []string{"serverid"}, Background: true, Sparse: true})
		}
		return c
	}
	return nil
}

func GetAllCoinPoolSettingData() (datas []model.CoinPoolSetting, err error) {
	c := CoinPoolSettingCollection()
	if c != nil {
		err = c.Find(nil).All(&datas)
		if err != nil {
			logger.Logger.Trace("GetAllCoinPoolSettingData err:", err)
			return
		}
	}
	return
}

func UpsertCoinPoolSetting(cps, old *model.CoinPoolSetting) (err error) {
	if cps != nil {
		c := CoinPoolSettingCollection()
		if c != nil {
			_, err = c.Upsert(bson.M{"_id": cps.Id}, cps)
			if err != nil {
				return err
			}
		}
	}
	if old != nil {
		c := CoinPoolSettingHisCollection()
		if c != nil {
			old.Id = bson.NewObjectId()
			old.UptTime = time.Now()
			err = c.Insert(old)
			if err != nil {
				return err
			}
		}
	}
	return
}

//删除水池历史调控记录
func RemoveCoinPoolSettingHis(ts time.Time) (*mgo.ChangeInfo, error) {
	return CoinPoolSettingHisCollection().RemoveAll(bson.M{"upttime": bson.M{"$lt": ts}})
}

type CoinPoolSettingSvc struct {
}

func (svc *CoinPoolSettingSvc) GetAllCoinPoolSettingData(args *struct{}, datas *[]model.CoinPoolSetting) (err error) {
	*datas, err = GetAllCoinPoolSettingData()
	return
}

func (svc *CoinPoolSettingSvc) UpsertCoinPoolSetting(args *model.UpsertCoinPoolSettingArgs, ret *bool) (err error) {
	err = UpsertCoinPoolSetting(args.Cps, args.Old)
	if err != nil {
		return err
	}
	*ret = true
	return
}

func (svc *CoinPoolSettingSvc) RemoveCoinPoolSettingHis(args *time.Time, ret *mgo.ChangeInfo) (err error) {
	ret, err = RemoveCoinPoolSettingHis(*args)
	if err != nil {
		return err
	}
	return
}

func init() {
	rpc.Register(new(CoinPoolSettingSvc))
}
