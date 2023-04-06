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

var (
	CoinWALErr = errors.New("log_coinwal open failed.")
)

func CoinWALCollection(plt string) *mongo.Collection {
	s := mongo.MgoSessionMgrSington.GetPltMgoSession(plt, model.CoinWALDBName)
	if s != nil {
		c, first := s.DB().C(model.CoinWALCollName)
		if first {
			c.EnsureIndex(mgo.Index{Key: []string{"snid"}, Background: true, Sparse: true})
			c.EnsureIndex(mgo.Index{Key: []string{"ingame"}, Background: true, Sparse: true})
			c.EnsureIndex(mgo.Index{Key: []string{"sceneid"}, Background: true, Sparse: true})
			c.EnsureIndex(mgo.Index{Key: []string{"logtype"}, Background: true, Sparse: true})
			c.EnsureIndex(mgo.Index{Key: []string{"cointype"}, Background: true, Sparse: true})
			c.EnsureIndex(mgo.Index{Key: []string{"-ts"}, Background: true, Sparse: true})
		}
		return c
	}
	return nil
}

func GetCoinWALBySnidAndInGameAndGreaterTs(plt string, id, sceneid int32, ts int64) (ret []model.CoinWAL, err error) {
	cond := bson.M{"snid": id, "ingame": bson.M{"$gt": 0}, "ts": bson.M{"$gt": ts}}
	if sceneid > 0 {
		cond["sceneid"] = sceneid
	}
	err = CoinWALCollection(plt).Find(cond).All(&ret)
	return
}

func GetCoinWALBySnidAndCoinTypeAndGreaterTs(plt string, id, coinType int32, ts int64) (ret []model.CoinWAL, err error) {
	err = CoinWALCollection(plt).Find(bson.M{"snid": id, "ingame": 0, "cointype": coinType, "ts": bson.M{"$gt": ts}}).All(&ret)
	return
}

func InsertCoinWAL(plt string, logs ...*model.CoinWAL) (err error) {
	clog := CoinWALCollection(plt)
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
		logger.Logger.Warn("InsertCoinWAL error:", err)
		return
	}
	return
}

func RemoveCoinWALInGame(plt string, id, sceneid int32, ts int64) (err error) {
	clog := CoinWALCollection(plt)
	if clog == nil {
		return CoinLogErr
	}

	cond := bson.M{"snid": id, "ingame": bson.M{"$gt": 0}, "ts": bson.M{"$lte": ts}}
	if sceneid > 0 {
		cond["sceneid"] = sceneid
	}
	_, err = clog.RemoveAll(cond)
	if err != nil {
		return
	}

	return
}

func RemoveCoinWALByCoinType(plt string, id, cointype int32, ts int64) (err error) {
	clog := CoinWALCollection(plt)
	if clog == nil {
		return CoinLogErr
	}

	_, err = clog.RemoveAll(bson.M{"snid": id, "ingame": 0, "cointype": cointype, "ts": bson.M{"$lte": ts}})
	if err != nil {
		return
	}

	return
}

type CoinWALSvc struct {
}

func (svc *CoinWALSvc) GetCoinWALBySnidAndInGameAndGreaterTs(args *model.CoinWALWithSnid_InGame_GreaterTsArgs, ret *[]model.CoinWAL) (err error) {
	*ret, err = GetCoinWALBySnidAndInGameAndGreaterTs(args.Plt, args.SnId, args.RoomId, args.Ts)
	return
}

func (svc *CoinWALSvc) RemoveCoinWALBySnidAndInGameAndGreaterTs(args *model.CoinWALWithSnid_InGame_GreaterTsArgs, ret *bool) (err error) {
	err = RemoveCoinWALInGame(args.Plt, args.SnId, args.RoomId, args.Ts)
	*ret = err == nil
	return
}

func init() {
	rpc.Register(new(CoinWALSvc))
}
