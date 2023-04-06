package svc

import (
	"errors"
	"net/rpc"
	"time"

	"games.yol.com/win88/dbproxy/mongo"
	"games.yol.com/win88/model"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/idealeak/goserver/core/logger"
)

var (
	CoinLogErr = errors.New("log coinex log open failed.")
)

const CoinLogMaxLimitPerQuery = 100

func CoinLogsCollection(plt string) *mongo.Collection {
	s := mongo.MgoSessionMgrSington.GetPltMgoSession(plt, model.CoinLogDBName)
	if s != nil {
		c_coinlogrec, first := s.DB().C(model.CoinLogCollName)
		if first {
			c_coinlogrec.EnsureIndex(mgo.Index{Key: []string{"snid"}, Background: true, Sparse: true})
			c_coinlogrec.EnsureIndex(mgo.Index{Key: []string{"logtype"}, Background: true, Sparse: true})
			c_coinlogrec.EnsureIndex(mgo.Index{Key: []string{"-time"}, Background: true, Sparse: true})
			c_coinlogrec.EnsureIndex(mgo.Index{Key: []string{"channel"}, Background: true, Sparse: true})
			c_coinlogrec.EnsureIndex(mgo.Index{Key: []string{"promoter"}, Background: true, Sparse: true})
			c_coinlogrec.EnsureIndex(mgo.Index{Key: []string{"oper"}, Background: true, Sparse: true})
			c_coinlogrec.EnsureIndex(mgo.Index{Key: []string{"roomid"}, Background: true, Sparse: true})
			c_coinlogrec.EnsureIndex(mgo.Index{Key: []string{"seqno"}, Background: true, Sparse: true})
		}
		return c_coinlogrec
	}
	return nil
}

type CoinLogSvc struct {
}

func GetCoinLogBySnidAndInGameAndGreaterTs(plt string, id int32, roomid int32, ts int64) (ret []model.CoinLog, err error) {
	if roomid != 0 {
		err = CoinLogsCollection(plt).Find(bson.M{"snid": id, "roomid": roomid, "ingame": bson.M{"$gt": 0}, "ts": bson.M{"$gt": ts}}).All(&ret)
	} else {
		err = CoinLogsCollection(plt).Find(bson.M{"snid": id, "ingame": bson.M{"$gt": 0}, "ts": bson.M{"$gt": ts}}).All(&ret)
	}
	return
}


func (svc *CoinLogSvc) GetCoinLogBySnidAndLessTs(args *model.GetCoinLogBySnidAndLessTsArg, ret *[]model.CoinLog) (err error) {
	err = CoinLogsCollection(args.Plt).Find(bson.M{"snid": args.SnId, "time": bson.M{"$lt": time.Unix(args.Ts, 0)}}).Sort("-ts").Limit(CoinLogMaxLimitPerQuery).All(ret)
	return
}

func (svc *CoinLogSvc) GetCoinLogBySnidAndTypeAndInRangeTsLimitByRange(args *model.GetCoinLogBySnidAndTypeAndInRangeTsLimitByRangeArg, ret *model.GetCoinLogBySnidAndTypeAndInRangeTsLimitByRangeRet) (err error) {
	limitDataNum := args.ToIdx - args.FromIdx
	if limitDataNum < 0 {
		limitDataNum = 0
	}
	if limitDataNum > CoinLogMaxLimitPerQuery {
		limitDataNum = CoinLogMaxLimitPerQuery
	}

	plt := args.Plt
	startts := args.StartTs
	endts := args.EndTs
	logType := args.LogType
	id := args.SnId
	fromIndex := args.FromIdx
	if (startts == 0 || endts == 0) && logType == 0 {
		ret.Count, _ = CoinLogsCollection(plt).Find(bson.M{"snid": id}).Count()
		err = CoinLogsCollection(plt).Find(bson.M{"snid": id}).Skip(fromIndex).Limit(limitDataNum).All(&ret.Logs)
	} else if startts == 0 || endts == 0 {
		ret.Count, _ = CoinLogsCollection(plt).Find(bson.M{"snid": id, "logtype": logType}).Count()
		err = CoinLogsCollection(plt).Find(bson.M{"snid": id, "logtype": logType}).Skip(fromIndex).Limit(limitDataNum).All(&ret.Logs)
	} else if logType == 0 {
		ret.Count, _ = CoinLogsCollection(plt).Find(bson.M{"snid": id, "ts": bson.M{"$gte": startts, "$lte": endts}}).Count()
		err = CoinLogsCollection(plt).Find(bson.M{"snid": id, "ts": bson.M{"$gte": startts, "$lte": endts}}).Skip(fromIndex).Limit(limitDataNum).All(&ret.Logs)
	} else {
		ret.Count, _ = CoinLogsCollection(plt).Find(bson.M{"snid": id, "logtype": logType, "ts": bson.M{"$gte": startts, "$lte": endts}}).Count()
		err = CoinLogsCollection(plt).Find(bson.M{"snid": id, "logtype": logType, "ts": bson.M{"$gte": startts, "$lte": endts}}).Skip(fromIndex).Limit(limitDataNum).All(&ret.Logs)
	}
	return
}

func (svc *CoinLogSvc) InsertCoinLog(log *model.CoinLog, ret *bool) (err error) {
	clog := CoinLogsCollection(log.Platform)
	if clog == nil {
		return
	}
	err = clog.Insert(log)
	if err != nil {
		logger.Logger.Warn("InsertCoinLog error:", err)
		return
	}

	//err = InsertCoinWAL(log.Platform, model.NewCoinWAL(log.SnId, log.Count, log.LogType, log.InGame, log.CoinType, log.RoomId, log.Time.UnixNano()))
	//if err != nil {
	//	logger.Logger.Warn("InsertCoinWAL error:", err)
	//	return
	//}
	*ret = true
	return
}

func (svc *CoinLogSvc) RemoveCoinLogOne(args *model.RemoveCoinLogOneArg, ret *bool) (err error) {
	clog := CoinLogsCollection(args.Plt)
	if clog == nil {
		return CoinLogErr
	}
	err = clog.RemoveId(args.Id)
	if err != nil {
		return
	}

	*ret = true
	return
}

func (svc *CoinLogSvc) UpdateCoinLogRemark(args *model.UpdateCoinLogRemarkArg, ret *bool) (err error) {
	clog := CoinLogsCollection(args.Plt)
	if clog == nil {
		return CoinLogErr
	}
	err = clog.UpdateId(args.Id, bson.M{"$set": bson.M{"remark": args.Remark}})
	if err != nil {
		return
	}

	*ret = true
	return
}

func init() {
	rpc.Register(new(CoinLogSvc))
}
