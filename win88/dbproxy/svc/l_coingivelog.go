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
	CoinGiveLogErr = errors.New("log coingive log open failed.")
)

func CoinGiveLogCollection(plt string) *mongo.Collection {
	s := mongo.MgoSessionMgrSington.GetPltMgoSession(plt, model.CoinGiveLogDBName)
	if s != nil {
		c_coinGiveLogRec, first := s.DB().C(model.CoinGiveLogCollName)
		if first {
			c_coinGiveLogRec.EnsureIndex(mgo.Index{Key: []string{"snid"}, Background: true, Sparse: true})
			c_coinGiveLogRec.EnsureIndex(mgo.Index{Key: []string{"promoter"}, Background: true, Sparse: true})
			c_coinGiveLogRec.EnsureIndex(mgo.Index{Key: []string{"state"}, Background: true, Sparse: true})
			c_coinGiveLogRec.EnsureIndex(mgo.Index{Key: []string{"ts"}, Background: true, Sparse: true})
			c_coinGiveLogRec.EnsureIndex(mgo.Index{Key: []string{"time"}, Background: true, Sparse: true})
		}
		return c_coinGiveLogRec
	}
	return nil
}

type CoinGiveLogSvc struct {
}

func (svc *CoinGiveLogSvc) UpdateGiveCoinLastFlow(args *model.UpdateGiveCoinLastFlowArg, ret *bool) (err error) {
	clog := CoinGiveLogCollection(args.Plt)
	if clog == nil {
		return CoinGiveLogErr
	}
	err = clog.Update(bson.M{"_id": bson.ObjectIdHex(args.LogId)}, bson.M{"$set": bson.M{"flow": args.Flow}})
	if err != nil {
		logger.Logger.Warn("svc.InsertSignleLoginLog error:", err)
		return
	}
	*ret = true
	return
}

func (svc *CoinGiveLogSvc) GetGiveCoinLastFlow(args *model.GetGiveCoinLastFlowArg, log *model.CoinGiveLog) (err error) {
	clog := CoinGiveLogCollection(args.Plt)
	if clog == nil {
		return CoinGiveLogErr
	}

	err = clog.Find(bson.M{"_id": bson.ObjectIdHex(args.LogId)}).One(&log)
	return
}

func (svc *CoinGiveLogSvc) GetCoinGiveLogList(args *model.GetCoinGiveLogListArg, logs *[]*model.CoinGiveLog) (err error) {
	clog := CoinGiveLogCollection(args.Plt)
	if clog == nil {
		return CoinGiveLogErr
	}

	err = clog.Find(bson.M{"snid": args.SnId, "state": 0, "ts": bson.M{"$gt": args.TsStart, "$lte": args.TsEnd}, "ver": args.Ver}).All(&logs)
	return
}

func (svc *CoinGiveLogSvc) GetCoinGiveLogListByState(args *model.GetCoinGiveLogListByStateArg, logs *[]*model.CoinGiveLog) (err error) {
	clog := CoinGiveLogCollection(args.Plt)
	if clog == nil {
		return CoinGiveLogErr
	}

	err = clog.Find(bson.M{"snid": args.SnId, "state": args.State}).All(&logs)
	return
}

func (svc *CoinGiveLogSvc) ResetCoinGiveLogList(args *model.ResetCoinGiveLogListArg, ret *bool) (err error) {
	clog := CoinGiveLogCollection(args.Plt)
	if clog == nil {
		return CoinGiveLogErr
	}
	_, err = clog.UpdateAll(bson.M{"snid": args.SnId, "state": args.Ts}, bson.M{"$set": bson.M{"state": 0}})
	if err != nil {
		logger.Logger.Warn("svc.ResetCoinGiveLogList error:", err)
		return
	}
	*ret = true
	return
}

func (svc *CoinGiveLogSvc) SetCoinGiveLogList(args *model.SetCoinGiveLogListArg, ret *bool) (err error) {
	clog := CoinGiveLogCollection(args.Plt)
	if clog == nil {
		return CoinGiveLogErr
	}

	_, err = clog.UpdateAll(bson.M{"snid": args.SnId, "state": 0, "ts": bson.M{"$gt": args.TsStart, "$lte": args.TsEnd}}, bson.M{"$set": bson.M{"state": args.State}})
	if err != nil {
		logger.Logger.Warn("svc.SetCoinGiveLogList error:", err)
		return
	}
	*ret = true
	return
}

func (svc *CoinGiveLogSvc) UpdateCoinLog(args *model.UpdateCoinLogArg, ret *bool) (err error) {
	clog := CoinGiveLogCollection(args.Plt)
	if clog == nil {
		return CoinGiveLogErr
	}

	err = clog.Update(bson.M{"_id": bson.ObjectIdHex(args.LogId)}, bson.M{"$set": bson.M{"state": args.State}})
	if err != nil {
		logger.Logger.Warn("svc.UpdateCoinLog error:", err)
		return
	}
	*ret = true
	return
}

func (svc *CoinGiveLogSvc) InsertGiveCoinLog(log *model.CoinGiveLog, ret *bool) (err error) {
	clog := CoinGiveLogCollection(log.Platform)
	if clog == nil {
		return CoinGiveLogErr
	}
	err = clog.Insert(log)
	if err != nil {
		logger.Logger.Warn("InsertCoinGiveSLog error:", err)
		return
	}
	*ret = true
	return
}

func init() {
	rpc.Register(new(CoinGiveLogSvc))
}
