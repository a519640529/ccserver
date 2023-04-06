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
	HundredjackpotLogErr = errors.New("log hundredjack log open failed.")
)

const HundredjackpotLogMaxLimitPerQuery = 99

// HundredjackpotLogsCollection mgo连接
func HundredjackpotLogsCollection(plt string) *mongo.Collection {
	s := mongo.MgoSessionMgrSington.GetPltMgoSession(plt, model.HundredjackpotLogDBName)
	if s != nil {
		c, first := s.DB().C(model.HundredjackpotLogCollName)
		if first {
			c.EnsureIndex(mgo.Index{Key: []string{"snid"}, Background: true, Sparse: true})
			c.EnsureIndex(mgo.Index{Key: []string{"channel"}, Background: true, Sparse: true})
			c.EnsureIndex(mgo.Index{Key: []string{"promoter"}, Background: true, Sparse: true})
			c.EnsureIndex(mgo.Index{Key: []string{"ts"}, Background: true, Sparse: true})
			c.EnsureIndex(mgo.Index{Key: []string{"time"}, Background: true, Sparse: true})
			c.EnsureIndex(mgo.Index{Key: []string{"roomid"}, Background: true, Sparse: true})
			c.EnsureIndex(mgo.Index{Key: []string{"ingame"}, Background: true, Sparse: true})
		}
		return c
	}
	return nil
}

// GetHundredjackpotLogTsByPlatformAndGameID 时间排名
func GetHundredjackpotLogTsByPlatformAndGameID(platform string, gameid int32) (ret []model.HundredjackpotLog, err error) {
	err = HundredjackpotLogsCollection(platform).Find(bson.M{"ingame": gameid}).Sort("-ts").Limit(HundredjackpotLogMaxLimitPerQuery).All(&ret)
	return
}

// GetHundredjackpotLogCoinByPlatformAndGameID 中奖金币排名
func GetHundredjackpotLogCoinByPlatformAndGameID(platform string, gameid int32) (ret []model.HundredjackpotLog, err error) {
	err = HundredjackpotLogsCollection(platform).Find(bson.M{"ingame": gameid}).Sort("-coin", "ts").Limit(HundredjackpotLogMaxLimitPerQuery).All(&ret)
	return
}

// GetHundredjackpotLogTsByPlatformAndGameFreeID 时间排名
func GetHundredjackpotLogTsByPlatformAndGameFreeID(platform string, gamefreeid int32) (ret []model.HundredjackpotLog, err error) {
	err = HundredjackpotLogsCollection(platform).Find(bson.M{"roomid": gamefreeid}).Sort("-ts").Limit(HundredjackpotLogMaxLimitPerQuery).All(&ret)
	return
}

// GetHundredjackpotLogCoinByPlatformAndGameFreeID 中奖金币排名
func GetHundredjackpotLogCoinByPlatformAndGameFreeID(platform string, gamefreeid int32) (ret []model.HundredjackpotLog, err error) {
	err = HundredjackpotLogsCollection(platform).Find(bson.M{"roomid": gamefreeid}).Sort("-coin", "ts").Limit(HundredjackpotLogMaxLimitPerQuery).All(&ret)
	return
}

// GetLastHundredjackpotLogBySnidAndGameID .
func GetLastHundredjackpotLogBySnidAndGameID(plt string, id int32, gamefreeid int32) (log *model.HundredjackpotLog, err error) {
	var data model.HundredjackpotLog
	err = HundredjackpotLogsCollection(plt).Find(bson.M{"snid": id, "roomid": gamefreeid}).Sort("-ts").Limit(1).One(&data)
	if err == nil {
		log = &data
	}
	return
}

// GetHundredjackpotLogByID .
func GetHundredjackpotLogByID(plt string, id int32) ([]string, error) {
	var data model.HundredjackpotLog
	err := HundredjackpotLogsCollection(plt).Find(bson.M{"_id": id}).One(&data)
	if err == nil {
		return data.GameData, err
	}
	return nil, err
}

// InsertHundredjackpotLog .
func InsertHundredjackpotLog(log *model.HundredjackpotLog) (err error) {
	clog := HundredjackpotLogsCollection(log.Platform)
	if clog == nil {
		return
	}
	err = clog.Insert(log)
	if err != nil {
		logger.Logger.Warn("InsertHundredjackpotLog error:", err)
		return
	}
	return
}

// UpdateLikeNum 点赞次数
func UpdateLikeNum(plt string, id bson.ObjectId, likeNum int32, linkeSnids string) error {
	hlog := HundredjackpotLogsCollection(plt)
	if hlog == nil {
		return HundredjackpotLogErr
	}

	return hlog.Update(bson.M{"_id": id}, bson.D{{"$set", bson.D{{"likenum", likeNum}, {"linkesnids", linkeSnids}}}})
}

// UpdatePlayBlackNum 回放次数
func UpdatePlayBlackNum(plt string, id bson.ObjectId, playblackNum int32) ([]string, error) {
	hlog := HundredjackpotLogsCollection(plt)
	if hlog == nil {
		return nil, HundredjackpotLogErr
	}
	err := hlog.Update(bson.M{"_id": id}, bson.D{{"$set", bson.D{{"playblacknum", playblackNum}}}})
	var data model.HundredjackpotLog
	err = hlog.Find(bson.M{"_id": id}).One(&data)
	if err == nil {
		return data.GameData, err
	}
	return nil, err
}

// RemoveHundredjackpotLogOne 移除
func RemoveHundredjackpotLogOne(plt string, id bson.ObjectId) error {
	cpay := HundredjackpotLogsCollection(plt)
	if cpay == nil {
		return HundredjackpotLogErr
	}
	return cpay.RemoveId(id)
}

type HundredjackpotLogSvc struct {
}

func (svc *HundredjackpotLogSvc) GetHundredjackpotLogTsByPlatformAndGameID(args *model.GetHundredjackpotLogArgs, ret *[]model.HundredjackpotLog) (err error) {
	*ret, err = GetHundredjackpotLogTsByPlatformAndGameID(args.Plt, args.Id1)
	return err
}

func (svc *RebateLogSvc) GetHundredjackpotLogCoinByPlatformAndGameID(args *model.GetHundredjackpotLogArgs, ret *[]model.HundredjackpotLog) (err error) {
	*ret, err = GetHundredjackpotLogCoinByPlatformAndGameID(args.Plt, args.Id1)
	return err
}

func (svc *RebateLogSvc) GetHundredjackpotLogTsByPlatformAndGameFreeID(args *model.GetHundredjackpotLogArgs, ret *[]model.HundredjackpotLog) (err error) {
	*ret, err = GetHundredjackpotLogTsByPlatformAndGameFreeID(args.Plt, args.Id1)
	return err
}

func (svc *RebateLogSvc) GetHundredjackpotLogCoinByPlatformAndGameFreeID(args *model.GetHundredjackpotLogArgs, ret *[]model.HundredjackpotLog) (err error) {
	*ret, err = GetHundredjackpotLogCoinByPlatformAndGameFreeID(args.Plt, args.Id1)
	return err
}

func (svc *RebateLogSvc) GetLastHundredjackpotLogBySnidAndGameID(args *model.GetHundredjackpotLogArgs, ret **model.HundredjackpotLog) (err error) {
	*ret, err = GetLastHundredjackpotLogBySnidAndGameID(args.Plt, args.Id1, args.Id2)
	return err
}

func (svc *RebateLogSvc) InsertHundredjackpotLog(args *model.HundredjackpotLog, ret *bool) (err error) {
	err = InsertHundredjackpotLog(args)
	if err == nil {
		*ret = true
	}
	return err
}

func (svc *RebateLogSvc) UpdateLikeNum(args *model.UpdateLikeNumArgs, ret *bool) (err error) {
	err = UpdateLikeNum(args.Plt, args.Id, args.LikeNum, args.LikeSnIds)
	if err == nil {
		*ret = true
	}
	return err
}

func (svc *RebateLogSvc) UpdatePlayBlackNum(args *model.UpdatePlayBlackNumArgs, ret *[]string) (err error) {
	*ret, err = UpdatePlayBlackNum(args.Plt, args.Id, args.PlayblackNum)
	return err
}

func (svc *RebateLogSvc) RemoveHundredjackpotLogOne(args *model.RemoveHundredjackpotLogOneArgs, ret *bool) (err error) {
	err = RemoveHundredjackpotLogOne(args.Plt, args.Id)
	if err == nil {
		*ret = true
	}
	return err
}

func init() {
	rpc.Register(new(HundredjackpotLogSvc))
}
