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
	BlackWhiteCoinLogDBName   = "log"
	BlackWhiteCoinLogCollName = "log_blackwhitelist"
	BlackWhiteCoinLogErr      = errors.New("User blackwhite log open failed.")
)

func BlackWhiteCoinLogsCollection(plt string) *mongo.Collection {
	s := mongo.MgoSessionMgrSington.GetPltMgoSession(plt, BlackWhiteCoinLogDBName)
	if s != nil {
		c_blackwhiterec, first := s.DB().C(BlackWhiteCoinLogCollName)
		if first {
			c_blackwhiterec.EnsureIndex(mgo.Index{Key: []string{"snid"}, Background: true, Sparse: true})
			c_blackwhiterec.EnsureIndex(mgo.Index{Key: []string{"platform"}, Background: true, Sparse: true})
		}
		return c_blackwhiterec
	}
	return nil
}

type BlackWhiteCoinSvc struct {
}

func (svc *BlackWhiteCoinSvc) InsertBlackWhiteCoinLogs(logs []*model.BlackWhiteCoinLog, ret *model.BlackWhiteCoinRet) (err error) {
	if len(logs) == 0 {
		return errors.New("len(logs) == 0")
	}
	clog := BlackWhiteCoinLogsCollection(logs[0].Platform)
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
		logger.Logger.Warn("svc.InsertBlackWhiteCoinLog error:", err)
		ret.Err = err
		return
	}
	return
}
func (svc *BlackWhiteCoinSvc) InsertBlackWhiteCoinLog(log *model.BlackWhiteCoinLog, ret *model.BlackWhiteCoinRet) error {
	if log == nil {
		return errors.New("log == nil")
	}
	clog := BlackWhiteCoinLogsCollection(log.Platform)
	if clog == nil {
		return BlackWhiteCoinLogErr
	}
	err := clog.Insert(log)
	if err != nil {
		logger.Logger.Error("svc.InsertBlackWhiteCoinLog error", err)
		ret.Err = err
		return err
	}
	return nil
}

func (svc *BlackWhiteCoinSvc) RemoveBlackWhiteCoinLog(args *model.BlackWhiteCoinArg, ret *model.BlackWhiteCoinRet) error {
	clog := BlackWhiteCoinLogsCollection(args.Platform)
	if clog == nil {
		return BlackWhiteCoinLogErr
	}
	ret.Err = clog.RemoveId(args.Id)
	return ret.Err
}

func (svc *BlackWhiteCoinSvc) GetBlackWhiteCoinLogByBillNo(args *model.BlackWhiteCoinArg, ret *model.BlackWhiteCoinRet) error {
	clog := BlackWhiteCoinLogsCollection(args.Platform)
	if clog == nil {
		return BlackWhiteCoinLogErr
	}
	err := clog.Find(bson.M{"snid": args.SnId, "billno": args.BillNo}).One(ret.Data)
	if err != nil {
		logger.Logger.Error("svc.GetBlackWhiteCoinLogByBillNo error", err)
		return err
	}
	return nil
}
func init() {
	rpc.Register(&BlackWhiteCoinSvc{})
}
