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
	PayCoinLogErr = errors.New("user_coinlog db open failed.")
)

const (
	PayCoinLogType_Coin        int32 = iota //金币对账日志
	PayCoinLogType_SafeBoxCoin              //保险箱对账日志
	PayCoinLogType_Club                     //俱乐部对账日志
	PayCoinLogType_Ticket                   //比赛入场券
)

func PayCoinLogsCollection(plt string) *mongo.Collection {
	s := mongo.MgoSessionMgrSington.GetPltMgoSession(plt, model.PayCoinLogDBName)
	if s != nil {
		c, first := s.DB().C(model.PayCoinLogCollName)
		if first {
			c.EnsureIndex(mgo.Index{Key: []string{"snid", "logtype", "timestamp"}, Background: true, Sparse: true})
			c.EnsureIndex(mgo.Index{Key: []string{"snid", "billno"}, Unique: true, Background: true, Sparse: true})
			c.EnsureIndex(mgo.Index{Key: []string{"timestamp"}, Background: true, Sparse: true})
			c.EnsureIndex(mgo.Index{Key: []string{"time"}, Background: true, Sparse: true})
		}
		return c
	}
	return nil
}

func InsertPayCoinLogs(plt string, logs ...*model.PayCoinLog) (err error) {
	cpay := PayCoinLogsCollection(plt)
	if cpay == nil {
		return
	}
	switch len(logs) {
	case 0:
		return errors.New("no data")
	case 1:
		err = cpay.Insert(logs[0])
	default:
		docs := make([]interface{}, 0, len(logs))
		for _, log := range logs {
			docs = append(docs, log)
		}
		err = cpay.Insert(docs...)
	}
	if err != nil {
		logger.Logger.Warn("InsertPayCoinLog error:", err)
		return
	}

	if len(logs) == 1 {
		err = InsertCoinWAL(plt, model.NewCoinWAL(logs[0].SnId, logs[0].Coin+logs[0].CoinEx, logs[0].Inside, 0, logs[0].LogType, 0, logs[0].TimeStamp))
	} else {
		docs := make([]*model.CoinWAL, 0, len(logs))
		for _, log := range logs {
			docs = append(docs, model.NewCoinWAL(log.SnId, log.Coin+log.CoinEx, log.Inside, 0, log.LogType, 0, log.TimeStamp))
		}
		err = InsertCoinWAL(plt, docs...)
	}

	return
}

func InsertPayCoinLog(plt string, log *model.PayCoinLog) error {
	cpay := PayCoinLogsCollection(plt)
	if cpay == nil {
		return PayCoinLogErr
	}
	err := cpay.Insert(log)
	if err == nil {
		err = InsertCoinWAL(plt, model.NewCoinWAL(log.SnId, log.Coin+log.CoinEx, log.Inside, 0, log.LogType, 0, log.TimeStamp))
	}
	return err
}

func RemovePayCoinLog(plt string, id bson.ObjectId) error {
	cpay := PayCoinLogsCollection(plt)
	if cpay == nil {
		return PayCoinLogErr
	}
	return cpay.RemoveId(id)
}

func UpdatePayCoinLogStatus(plt string, id bson.ObjectId, status int32) error {
	cpay := PayCoinLogsCollection(plt)
	if cpay == nil {
		return PayCoinLogErr
	}
	return cpay.UpdateId(id, bson.M{"$set": bson.M{"status": status}})
}

func GetAllPayCoinLog(plt string, snid int32, ts int64) (ret []model.PayCoinLog, err error) {
	err = PayCoinLogsCollection(plt).Find(bson.M{"snid": snid, "logtype": PayCoinLogType_Coin, "timestamp": bson.M{"$gt": ts}}).All(&ret)
	return
}

func GetAllPaySafeBoxCoinLog(plt string, snid int32, ts int64) (ret []model.PayCoinLog, err error) {
	err = PayCoinLogsCollection(plt).Find(bson.M{"snid": snid, "logtype": PayCoinLogType_SafeBoxCoin, "timestamp": bson.M{"$gt": ts}}).All(&ret)
	return
}

func GetAllPayClubCoinLog(plt string, snid int32, ts int64) (ret []model.PayCoinLog, err error) {
	err = PayCoinLogsCollection(plt).Find(bson.M{"snid": snid, "logtype": PayCoinLogType_Club, "timestamp": bson.M{"$gt": ts}}).All(&ret)
	return
}

func GetAllPayTicketCoinLog(plt string, snid int32, ts int64) (ret []model.PayCoinLog, err error) {
	err = PayCoinLogsCollection(plt).Find(bson.M{"snid": snid, "logtype": PayCoinLogType_Ticket, "timestamp": bson.M{"$gt": ts}}).All(&ret)
	return
}

func GetPayCoinLogByBillNo(plt string, snid int32, billNo int64) (*model.PayCoinLog, error) {
	var log model.PayCoinLog
	err := PayCoinLogsCollection(plt).Find(bson.M{"snid": snid, "billno": billNo}).One(&log)
	return &log, err
}

type PayCoinLogSvc struct {
}

var _PayCoinLogSvc = &PayCoinLogSvc{}

func (svc *PayCoinLogSvc) InsertPayCoinLogs(args *model.InsertPayCoinLogArgs, ret *bool) (err error) {
	err = InsertPayCoinLogs(args.Plt, args.Logs...)
	if err != nil {
		return err
	}
	*ret = true
	return nil
}

func (svc *PayCoinLogSvc) InsertPayCoinLog(args *model.InsertPayCoinLogArgs, ret *bool) (err error) {
	err = InsertPayCoinLogs(args.Plt, args.Logs...)
	if err != nil {
		return err
	}
	*ret = true
	return nil
}

func (svc *PayCoinLogSvc) RemovePayCoinLog(args *model.RemovePayCoinLogArgs, ret *bool) (err error) {
	err = RemovePayCoinLog(args.Plt, args.Id)
	if err != nil {
		return err
	}
	*ret = true
	return nil
}

func (svc *PayCoinLogSvc) UpdatePayCoinLogStatus(args *model.UpdatePayCoinLogStatusArgs, ret *bool) (err error) {
	err = UpdatePayCoinLogStatus(args.Plt, args.Id, args.Status)
	if err != nil {
		return err
	}
	*ret = true
	return nil
}

func (svc *PayCoinLogSvc) GetAllPayCoinLog(args *model.GetPayCoinLogArgs, ret *[]model.PayCoinLog) (err error) {
	*ret, err = GetAllPayCoinLog(args.Plt, args.SnId, args.Cond)
	if err != nil {
		return err
	}
	return nil
}

func (svc *PayCoinLogSvc) GetAllPaySafeBoxCoinLog(args *model.GetPayCoinLogArgs, ret *[]model.PayCoinLog) (err error) {
	*ret, err = GetAllPaySafeBoxCoinLog(args.Plt, args.SnId, args.Cond)
	if err != nil {
		return err
	}
	return nil
}

func (svc *PayCoinLogSvc) GetAllPayClubCoinLog(args *model.GetPayCoinLogArgs, ret *[]model.PayCoinLog) (err error) {
	*ret, err = GetAllPayClubCoinLog(args.Plt, args.SnId, args.Cond)
	if err != nil {
		return err
	}
	return nil
}

func (svc *PayCoinLogSvc) GetAllPayTicketCoinLog(args *model.GetPayCoinLogArgs, ret *[]model.PayCoinLog) (err error) {
	*ret, err = GetAllPayTicketCoinLog(args.Plt, args.SnId, args.Cond)
	if err != nil {
		return err
	}
	return nil
}

func (svc *PayCoinLogSvc) GetPayCoinLogByBillNo(args *model.GetPayCoinLogArgs, ret **model.PayCoinLog) (err error) {
	*ret, err = GetPayCoinLogByBillNo(args.Plt, args.SnId, args.Cond)
	if err != nil {
		return err
	}
	return nil
}

func init() {
	rpc.Register(_PayCoinLogSvc)
}
