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
	LotteryLogDBErr = errors.New("log_lottery db open failed.")
)

const LotteryLogMaxLimitPerQuery = 100

func LotteryLogsCollection(plt string) *mongo.Collection {
	s := mongo.MgoSessionMgrSington.GetPltMgoSession(plt, model.LotteryLogDBName)
	if s != nil {
		c, first := s.DB().C(model.LotteryLogCollName)
		if first {
			c.EnsureIndex(mgo.Index{Key: []string{"gamefreeid"}, Background: true, Sparse: true})
			c.EnsureIndex(mgo.Index{Key: []string{"gameid"}, Background: true, Sparse: true})
			c.EnsureIndex(mgo.Index{Key: []string{"snid"}, Background: true, Sparse: true})
			c.EnsureIndex(mgo.Index{Key: []string{"time"}, Background: true, Sparse: true})
		}
		return c
	}
	return nil
}

func InsertLotteryLog(log *model.LotteryLog) error {
	err := LotteryLogsCollection(log.Platform).Insert(log)
	if err != nil {
		logger.Logger.Info("InsertLotteryLog error:", err)
		return err
	}
	return nil
}

func InsertLotteryLogs(logs ...*model.LotteryLog) (err error) {
	clog := LotteryLogsCollection(logs[0].Platform)
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
		logger.Logger.Warn("InsertLotteryLogs error:", err)
		return
	}
	return
}

func RemoveLotteryLog(plt string, ts time.Time) (*mgo.ChangeInfo, error) {
	return LotteryLogsCollection(plt).RemoveAll(bson.M{"time": bson.M{"$lt": ts}})
}

func GetLotteryLogBySnidAndLessTs(plt string, id int32, ts time.Time) (ret []model.LotteryLog, err error) {
	err = LotteryLogsCollection(plt).Find(bson.M{"snid": id, "time": bson.M{"$lt": ts}}).Limit(LotteryLogMaxLimitPerQuery).All(&ret)
	return
}

type LotteryLogSvc struct {
}

func (svc *LotteryLogSvc) InsertLotteryLogs(args []*model.LotteryLog, ret *bool) (err error) {
	err = InsertLotteryLogs(args...)
	if err == nil {
		*ret = true
	}
	return
}

func (svc *LotteryLogSvc) InsertLotteryLog(args *model.LotteryLog, ret *bool) (err error) {
	err = InsertLotteryLog(args)
	if err == nil {
		*ret = true
	}
	return
}

func (svc *LotteryLogSvc) RemoveLotteryLog(args *model.RemoveLotteryLogArgs, ret **mgo.ChangeInfo) (err error) {
	*ret, err = RemoveLotteryLog(args.Plt, args.Ts)
	return
}

func (svc *LotteryLogSvc) GetLotteryLogBySnidAndLessTs(args *model.GetLotteryLogBySnidAndLessTsArgs, ret *[]model.LotteryLog) (err error) {
	*ret, err = GetLotteryLogBySnidAndLessTs(args.Plt, args.Id, args.Ts)
	return
}

func init() {
	rpc.Register(new(LotteryLogSvc))
}
