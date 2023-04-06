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

var (
	GradeLogErr     = errors.New("log_grade open failed.")
	BillGradeLogErr = errors.New("User billgradelog open failed.")
)

const GradeLogMaxLimitPerQuery = 100

func GradeLogsCollection(plt string) *mongo.Collection {
	s := mongo.MgoSessionMgrSington.GetPltMgoSession(plt, model.GradeLogDBName)
	if s != nil {
		c, first := s.DB().C(model.GradeLogCollName)
		if first {
			c.EnsureIndex(mgo.Index{Key: []string{"channel"}, Background: true, Sparse: true})
			c.EnsureIndex(mgo.Index{Key: []string{"promoter"}, Background: true, Sparse: true})
			c.EnsureIndex(mgo.Index{Key: []string{"snid"}, Background: true, Sparse: true})
			c.EnsureIndex(mgo.Index{Key: []string{"inviterid"}, Background: true, Sparse: true})
			c.EnsureIndex(mgo.Index{Key: []string{"time"}, Background: true, Sparse: true})
		}
		return c
	}
	return nil
}

func InsertGradeLog(log *model.GradeLog) error {
	err := GradeLogsCollection(log.Platform).Insert(log)
	if err != nil {
		logger.Logger.Info("InsertGradeLog error:", err)
		return err
	}
	return nil
}

func InsertGradeLogs(logs ...*model.GradeLog) (err error) {
	if len(logs) == 0 {
		return nil
	}
	clog := GradeLogsCollection(logs[0].Platform)
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
		logger.Logger.Warn("InsertGradeLogs error:", err)
		return
	}
	return
}

func UpdateGradeLogStatus(plt string, id bson.ObjectId, status int32) error {
	glc := GradeLogsCollection(plt)
	if glc == nil {
		return GradeLogErr
	}
	return glc.UpdateId(id, bson.M{"$set": bson.M{"status": status}})
}

func GetGradeLogByPageAndSnId(plt string, snid, pageNo, pageSize int32) (ret *model.GradeLogLog) {
	clog := GradeLogsCollection(plt)
	if clog == nil {
		return nil
	}
	selecter := bson.M{"snid": snid}
	//selecter := bson.M{"snid": snid, "logtype": common.GainWay_GradeMatchGet}
	total, err := clog.Find(selecter).Count()
	if err != nil || total == 0 {
		return nil
	}
	ret = new(model.GradeLogLog)
	ret.PageNum = int32(math.Ceil(float64(total) / float64(pageSize)))
	if pageNo > ret.PageNum {
		pageNo = ret.PageNum
	}
	if pageNo <= 0 {
		pageNo = 1
	}
	limitNum := (pageNo - 1) * pageSize
	err = clog.Find(selecter).Skip(int(limitNum)).Limit(int(pageSize)).Sort("-time").All(&ret.Logs)
	if err != nil {
		logger.Logger.Error("Find gradelog data eror.", err)
		return nil
	}
	ret.PageNo = pageNo
	ret.PageSize = pageSize
	return
}

func RemoveGradeLog(plt string, id bson.ObjectId) error {
	glc := GradeLogsCollection(plt)
	if glc == nil {
		return GradeLogErr
	}
	return glc.RemoveId(id)
}

func GetGradeLogBySnidAndLessTs(plt string, id int32, ts time.Time) (ret []model.GradeLog, err error) {
	err = GradeLogsCollection(plt).Find(bson.M{"snid": id, "time": bson.M{"$lt": ts}}).Sort("-time", "-count").Limit(GradeLogMaxLimitPerQuery).All(&ret)
	return
}

func BillGradeLogsCollection(plt string) *mongo.Collection {
	s := mongo.MgoSessionMgrSington.GetPltMgoSession(plt, model.BillGradeLogDBName)
	if s != nil {
		c, first := s.DB().C(model.BillGradeLogCollName)
		if first {
			c.EnsureIndex(mgo.Index{Key: []string{"snid"}, Background: true, Sparse: true})
			c.EnsureIndex(mgo.Index{Key: []string{"logid"}, Background: true, Sparse: true})
			c.EnsureIndex(mgo.Index{Key: []string{"timestamp"}, Background: true, Sparse: true})
			c.EnsureIndex(mgo.Index{Key: []string{"time"}, Background: true, Sparse: true})
		}
		return c
	}
	return nil
}

func InsertBillGradeLogs(logs ...*model.BillGradeLog) (err error) {
	cpay := BillGradeLogsCollection(logs[0].Platform)
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
		logger.Logger.Warn("InsertBillGradeLogs error:", err)
		return
	}
	return
}

func InsertBillGradeLog(log *model.BillGradeLog) error {
	cpay := BillGradeLogsCollection(log.Platform)
	if cpay == nil {
		return BillGradeLogErr
	}
	return cpay.Insert(log)
}

func RemoveBillGradeLog(plt, logid string) error {
	bglc := BillGradeLogsCollection(plt)
	if bglc == nil {
		return BillGradeLogErr
	}
	return bglc.Remove(bson.M{"logid": logid})
}

func UpdateBillGradeLogStatus(plt, logId string, status int32) error {
	bglc := BillGradeLogsCollection(plt)
	if bglc == nil {
		return BillGradeLogErr
	}
	var conds []bson.M
	conds = append(conds, bson.M{"status": bson.M{"$ne": 3}})
	conds = append(conds, bson.M{"status": bson.M{"$ne": 9}})
	sql := bson.M{"logid": logId,
		"$and": conds,
	}
	return bglc.Update(sql, bson.M{"$set": bson.M{"status": status}})
}

func GetAllBillGradeLog(plt string, snid int32, ts int64) (ret []model.BillGradeLog, err error) {
	err = BillGradeLogsCollection(plt).Find(bson.M{"snid": snid, "timestamp": bson.M{"$gt": ts}}).All(&ret)
	return
}

func GetBillGradeLogByLogId(plt, logId string) (ret *model.BillGradeLog, err error) {
	var conds []bson.M
	conds = append(conds, bson.M{"status": bson.M{"$ne": 3}})
	conds = append(conds, bson.M{"status": bson.M{"$ne": 9}})
	sql := bson.M{"logid": logId,
		"$and": conds,
	}
	err = BillGradeLogsCollection(plt).Find(sql).One(&ret)
	return
}

func GetBillGradeLogByBillNo(plt string, snid int32, billNo int64) (*model.BillGradeLog, error) {
	var log model.BillGradeLog
	err := BillGradeLogsCollection(plt).Find(bson.M{"snid": snid, "billno": billNo}).One(&log)
	return &log, err
}

type GradeLogSvc struct {
}

func (svc *GradeLogSvc) InsertGradeLog(args *model.GradeLog, ret *bool) (err error) {
	err = InsertGradeLog(args)
	if err == nil {
		*ret = true
	}
	return
}

func (svc *GradeLogSvc) InsertGradeLogs(args []*model.GradeLog, ret *bool) (err error) {
	err = InsertGradeLogs(args...)
	if err == nil {
		*ret = true
	}
	return
}

func (svc *GradeLogSvc) UpdateGradeLogStatus(args *model.UpdateGradeLogStatusArgs, ret *bool) (err error) {
	err = UpdateGradeLogStatus(args.Plt, args.Id, args.Status)
	if err == nil {
		*ret = true
	}
	return
}

func (svc *GradeLogSvc) GetGradeLogByPageAndSnId(args *model.GetGradeLogByPageAndSnIdArgs, ret **model.GradeLogLog) (err error) {
	*ret = GetGradeLogByPageAndSnId(args.Plt, args.SnId, args.PageNo, args.PageSize)
	return
}

func (svc *GradeLogSvc) GetGradeLogBySnidAndLessTs(args *model.GetGradeLogBySnidAndLessTsArgs, ret *[]model.GradeLog) (err error) {
	*ret, err = GetGradeLogBySnidAndLessTs(args.Plt, args.SnId, args.Ts)
	return
}

func (svc *GradeLogSvc) RemoveGradeLog(args *model.RemoveGradeLogArgs, ret *bool) (err error) {
	err = RemoveGradeLog(args.Plt, args.Id)
	if err == nil {
		*ret = true
	}
	return
}

type BillGradeLogSvc struct {
}

func (svc *BillGradeLogSvc) InsertBillGradeLogs(args []*model.BillGradeLog, ret *bool) (err error) {
	err = InsertBillGradeLogs(args...)
	if err == nil {
		*ret = true
	}
	return
}

func (svc *BillGradeLogSvc) InsertBillGradeLog(args *model.BillGradeLog, ret *bool) (err error) {
	err = InsertBillGradeLog(args)
	if err == nil {
		*ret = true
	}
	return
}

func (svc *BillGradeLogSvc) RemoveBillGradeLog(args *model.RemoveBillGradeLogArgs, ret *bool) (err error) {
	err = RemoveBillGradeLog(args.Plt, args.LogId)
	if err == nil {
		*ret = true
	}
	return
}

func (svc *BillGradeLogSvc) UpdateBillGradeLogStatus(args *model.UpdateBillGradeLogStatusArgs, ret *bool) (err error) {
	err = UpdateBillGradeLogStatus(args.Plt, args.LogId, args.Status)
	if err == nil {
		*ret = true
	}
	return
}

func (svc *BillGradeLogSvc) GetAllBillGradeLog(args *model.GetAllBillGradeLogArgs, ret *[]model.BillGradeLog) (err error) {
	*ret, err = GetAllBillGradeLog(args.Plt, args.SnId, args.Ts)
	return
}

func (svc *BillGradeLogSvc) GetBillGradeLogByLogId(args *model.GetBillGradeLogByLogIdArgs, ret **model.BillGradeLog) (err error) {
	*ret, err = GetBillGradeLogByLogId(args.Plt, args.LogId)
	return
}

func (svc *BillGradeLogSvc) GetBillGradeLogByBillNo(args *model.GetBillGradeLogByBillNoArgs, ret **model.BillGradeLog) (err error) {
	*ret, err = GetBillGradeLogByBillNo(args.Plt, args.SnId, args.BillNo)
	return
}

func init() {
	rpc.Register(new(GradeLogSvc))
	rpc.Register(new(BillGradeLogSvc))
}
