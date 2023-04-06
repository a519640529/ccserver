package svc

import (
	"errors"
	"math"
	"net/rpc"
	"time"

	"games.yol.com/win88/dbproxy/mongo"
	"games.yol.com/win88/model"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/idealeak/goserver/core/logger"
)

func GameDetailedLogsCollection(plt string) *mongo.Collection {
	s := mongo.MgoSessionMgrSington.GetPltMgoSession(plt, model.GameDetailedLogDBName)
	if s != nil {
		c_gamedetailed, first := s.DB().C(model.GameDetailedLogCollName)
		if first {
			c_gamedetailed.EnsureIndex(mgo.Index{Key: []string{"gameid"}, Background: true, Sparse: true})
			c_gamedetailed.EnsureIndex(mgo.Index{Key: []string{"gamefreeid"}, Background: true, Sparse: true})
			c_gamedetailed.EnsureIndex(mgo.Index{Key: []string{"logid"}, Background: true, Sparse: true})
			c_gamedetailed.EnsureIndex(mgo.Index{Key: []string{"time"}, Background: true, Sparse: true})
			c_gamedetailed.EnsureIndex(mgo.Index{Key: []string{"ts"}, Background: true, Sparse: true})
		}
		return c_gamedetailed
	}
	return nil
}

type GameDetailedSvc struct {
}

func (svc *GameDetailedSvc) InsertGameDetailedLog(log *model.GameDetailedLog, ret *bool) (err error) {
	clog := GameDetailedLogsCollection(log.Platform)
	if clog == nil {
		logger.Logger.Error("svc.InsertGameDetailedLogs clog == nil")
		return
	}

	err = clog.Insert(log)
	if err != nil {
		logger.Logger.Warn("svc.InsertGameDetailedLogs error:", err)
		return
	}
	*ret = true
	return
}

func (svc *GameDetailedSvc) RemoveGameDetailedLog(args *model.RemoveGameDetailedLogArg, ret *mgo.ChangeInfo) (err error) {
	clog := GameDetailedLogsCollection(args.Plt)
	if clog == nil {
		logger.Logger.Error("svc.RemoveGameDetailedLog clog == nil")
		return errors.New("clog == nil")
	}
	ret, err = clog.RemoveAll(bson.M{"time": bson.M{"$lt": args.Ts}})
	if err != nil {
		logger.Logger.Error("svc.RemoveGameDetailedLog is error", err)
		return err
	}

	return nil
}

func (svc *GameDetailedSvc) GetAllGameDetailedLogsByTs(args *model.GameDetailedArg, ret *[]model.GameDetailedLog) error {
	gdlc := GameDetailedLogsCollection(args.Plt)
	if gdlc == nil {
		return nil
	}
	var sql = bson.M{"time": bson.M{"$gte": time.Unix(args.StartTime, 0), "$lte": time.Unix(args.EndTime, 0)}}
	err := gdlc.Find(sql).All(ret)
	if err != nil {
		logger.Logger.Error("svc.GetAllGameDetailedLogsByTs ")
		return err
	}
	return nil
}

func (svc *GameDetailedSvc) GetAllGameDetailedLogsByGameIdAndTs(args *model.GameDetailedGameIdAndArg, ret *[]model.GameDetailedLog) error {
	gdlc := GameDetailedLogsCollection(args.Plt)
	if gdlc == nil {
		return nil
	}
	var sql = bson.M{"gameid": args.Gameid}
	err := gdlc.Find(sql).Sort("-ts").Limit(args.LimitNum).All(ret)
	if err != nil {
		logger.Logger.Error("svc.GetAllGameDetailedLogsByGameIdAndTs ")
		return err
	}
	return nil
}

func (svc *GameDetailedSvc) GetPlayerHistory(args *model.GetPlayerHistoryArg, ret *model.GameDetailedLog) error {
	gdlc := GameDetailedLogsCollection(args.Plt)
	if gdlc == nil {
		return nil
	}
	err := gdlc.Find(bson.M{"logid": args.LogId}).One(ret)
	if err != nil {
		logger.Logger.Error("svc.GetPlayerHistory is error", err)
	}
	return nil
}

func (svc *GameDetailedSvc) GetPlayerHistoryAPI(args *model.GetPlayerHistoryAPIArg, ret *model.GameDetailedLogRet) error {
	logger.Logger.Tracef("GameDetailedSvc.GetPlayerHistoryAPI=====> args:%v", args)
	gdlc := GameDetailedLogsCollection(args.Platform)
	if gdlc == nil {
		return nil
	}
	var sql []bson.M
	//if args.SnId != 0 {
	//	sql = append(sql, bson.M{"snid": args.SnId})
	//}
	if args.Platform != "" {
		sql = append(sql, bson.M{"platform": args.Platform})
	}

	if args.StartTime != 0 {
		sql = append(sql, bson.M{"ts": bson.M{"$gte": args.StartTime, "$lte": args.EndTime}})
	}

	total, err := gdlc.Find(bson.M{"$and": sql}).Count()
	if err != nil {
		logger.Logger.Warn("svc.GetPlayerHistoryAPI Count error: ", err)
		return err
	}
	gdt := model.GameDetailedLogType{}
	if total == 0 {
		gdt.PageNo = args.PageNo
		gdt.PageSize = args.PageSize
		return nil
	}
	gdt.PageSum = int(math.Ceil(float64(total) / float64(args.PageSize)))
	if args.PageNo > gdt.PageSum {
		args.PageNo = gdt.PageSum
	}
	if args.PageNo <= 0 {
		args.PageNo = 1
	}
	limitNum := (args.PageNo - 1) * args.PageSize

	var data []*model.GameDetailedLog

	err = gdlc.Find(bson.M{"$and": sql}).Sort("-ts").Limit(args.PageSize).Skip(limitNum).All(&data)
	if err != nil {
		logger.Logger.Warn("svc.GetPlayerHistoryAPI error: ", err)
		return err
	}
	gdt.PageNo = args.PageNo
	gdt.PageSize = args.PageSize
	gdt.Data = data
	ret.Gplt = gdt
	logger.Logger.Tracef("GameDetailedSvc.GetPlayerHistoryAPI=====> ret:%v", ret)
	return nil
}

func init() {
	rpc.Register(new(GameDetailedSvc))
}
