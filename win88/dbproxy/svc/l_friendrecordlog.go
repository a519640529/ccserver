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
	FriendRecordLogErr = errors.New("friend record log open failed.")
)

func FriendRecordLogsCollection(plt string) *mongo.Collection {
	s := mongo.MgoSessionMgrSington.GetPltMgoSession(plt, model.FriendRecordLogDBName)
	if s != nil {
		c_friendrecordlog, first := s.DB().C(model.FriendRecordLogCollName)
		if first {
			c_friendrecordlog.EnsureIndex(mgo.Index{Key: []string{"snid"}, Background: true, Sparse: true})
			c_friendrecordlog.EnsureIndex(mgo.Index{Key: []string{"platform"}, Background: true, Sparse: true})
			c_friendrecordlog.EnsureIndex(mgo.Index{Key: []string{"gameid"}, Background: true, Sparse: true})
			c_friendrecordlog.EnsureIndex(mgo.Index{Key: []string{"ts"}, Background: true, Sparse: true})
		}
		return c_friendrecordlog
	}
	return nil
}
func InsertFriendRecordLog(log model.FriendRecord) error {
	frlog := FriendRecordLogsCollection(log.Platform)
	if frlog == nil {
		return FriendRecordLogErr
	}
	count, err := frlog.Find(bson.M{"snid": log.SnId, "platform": log.Platform, "gameid": log.GameId}).Count()
	if err != nil {
		return err
	}
	logger.Logger.Warn("InsertFriendRecordLog: count ", count)
	if count >= 50 {
		var oldLogs []model.FriendRecord
		err = frlog.Find(bson.M{"snid": log.SnId, "platform": log.Platform, "gameid": log.GameId}).Sort("-ts").Skip(19).All(&oldLogs)
		if err != nil {
			return err
		}
		logger.Logger.Warn("InsertFriendRecordLog: oldLogs ", oldLogs)
		for _, oldLog := range oldLogs {
			err = frlog.RemoveId(oldLog.LogId)
			if err != nil {
				return err
			}
		}
	}
	err = frlog.Insert(log)

	return err
}

func (svc *FriendRecordLogSvc) GetFriendRecordLogBySnid(args *model.FriendRecordSnidArg, ret *model.FriendRecordSnidRet) error {
	frlog := FriendRecordLogsCollection(args.Platform)
	if frlog == nil {
		return FriendRecordLogErr
	}
	var sql []bson.M
	if args.SnId != 0 {
		sql = append(sql, bson.M{"snid": args.SnId})
	}
	if args.Platform != "" {
		sql = append(sql, bson.M{"platform": args.Platform})
	}
	if args.GameId != 0 {
		sql = append(sql, bson.M{"gameid": args.GameId})
	}
	if args.Size == 0 {
		return nil
	}
	var datas []*model.FriendRecord
	err := frlog.Find(bson.M{"$and": sql}).Sort("-ts").Limit(20).All(&datas)
	if err != nil {
		logger.Logger.Warn("FriendRecordLogSvc error: ", err)
		return err
	}
	if datas == nil {
		return nil
	}
	frs := []*model.FriendRecord{}
	for _, data := range datas {
		fr := &model.FriendRecord{
			SnId:      data.SnId,
			GameId:    data.GameId,
			BaseScore: data.BaseScore,
			IsWin:     data.IsWin,
			Platform:  data.Platform,
			Ts:        data.Ts,
		}
		frs = append(frs, fr)
	}
	ret.FR = frs
	return nil
}

type FriendRecordLogSvc struct {
}

func init() {
	rpc.Register(new(FriendRecordLogSvc))
}
