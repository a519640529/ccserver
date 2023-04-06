package svc

import (
	"errors"
	"games.yol.com/win88/dbproxy/mongo"
	"games.yol.com/win88/model"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/idealeak/goserver/core/logger"
	"math"
	"net/rpc"
	"time"
)

var (
	RebateLogDBErr = errors.New("log_rebatetask db open failed.")
)

func RebateCollection(plt string) *mongo.Collection {
	s := mongo.MgoSessionMgrSington.GetPltMgoSession(plt, model.RebateLogDBName)
	if s != nil {
		c, first := s.DB().C(model.RebateLogCollName)
		if first {
			c.EnsureIndex(mgo.Index{Key: []string{"id"}, Unique: true, Background: true, Sparse: true})
			c.EnsureIndex(mgo.Index{Key: []string{"snid"}, Background: true, Sparse: true})
		}
		return c
	}
	return nil
}

func InsertRebateLog(plt string, data *model.Rebate) error {
	rebatec := RebateCollection(plt)
	if rebatec == nil {
		return nil
	}
	data.Id = bson.NewObjectId()
	data.Ts = time.Now().Unix()
	err := rebatec.Insert(data)
	if err != nil {
		logger.Logger.Error("Insert rebate data eror.", err)
		return err
	}
	return nil
}

func GetRebateLog(plt string, pageNo, pageSize int, snId int32) (r *model.RebateLog, err error) {
	r = new(model.RebateLog)
	rebatec := RebateCollection(plt)
	if rebatec == nil {
		return
	}
	total, err := rebatec.Find(bson.M{"snid": snId}).Count()
	if err != nil || total == 0 {
		return nil, err
	}
	r.PageSum = int(math.Ceil(float64(total) / float64(pageSize)))
	if pageNo > r.PageSum {
		pageNo = r.PageSum
	}
	if pageNo <= 0 {
		pageNo = 1
	}
	limitNum := (pageNo - 1) * pageSize
	err = rebatec.Find(bson.M{"snid": snId}).Skip(limitNum).Limit(pageSize).Sort("-ts").All(&r.Rebates)
	if err != nil {
		logger.Logger.Error("Find rebate data eror.", err)
		return nil, err
	}
	r.PageNo = pageNo
	r.PageSize = pageSize
	return
}

type RebateLogSvc struct {
}

func (svc *RebateLogSvc) InsertRebateLog(log *model.BankBindLog, ret *bool) error {
	err := InsertBankBindLog(log)
	if err == nil {
		*ret = true
	}
	return err
}

func (svc *RebateLogSvc) GetRebateLog(args *model.GetRebateLogArgs, ret **model.RebateLog) (err error) {
	*ret, err = GetRebateLog(args.Plt, args.PageNo, args.PageSize, args.SnId)
	return err
}

func init() {
	rpc.Register(new(RebateLogSvc))
}
