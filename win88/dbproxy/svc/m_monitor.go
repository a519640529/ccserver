package svc

import (
	"encoding/gob"
	"errors"
	"fmt"
	"net/rpc"
	"strings"
	"time"

	"games.yol.com/win88/dbproxy/mongo"
	"games.yol.com/win88/mgrsrv/api"
	"games.yol.com/win88/model"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/core/profile"
	"github.com/idealeak/goserver/core/schedule"
	"github.com/idealeak/goserver/core/transact"
	"github.com/idealeak/goserver/core/utils"
)

var (
	MonitorErr = errors.New("monitor_data open failed.")
)

func MonitorDataCollection(cname string) *mongo.Collection {
	s := mongo.MgoSessionMgrSington.GetGlobal(model.MonitorDBName)
	if s != nil {
		c, first := s.DB().C(fmt.Sprintf("%v_%v", model.MonitorPrefixName, strings.ToLower(cname)))
		if first {
			c.EnsureIndex(mgo.Index{Key: []string{"srvid"}, Background: true, Sparse: true})
			c.EnsureIndex(mgo.Index{Key: []string{"srvtype"}, Background: true, Sparse: true})
			c.EnsureIndex(mgo.Index{Key: []string{"key"}, Background: true, Sparse: true})
			c.EnsureIndex(mgo.Index{Key: []string{"time"}, Background: true, Sparse: true})
		}
		return c
	}
	return nil
}

type MonitorDataSvc struct {
}

func (svc *MonitorDataSvc) InsertMonitorData(args *model.MonitorDataArg, ret *bool) (err error) {
	clog := MonitorDataCollection(args.Name)
	if clog == nil {
		return
	}

	err = clog.Insert(args.Log)
	if err != nil {
		logger.Logger.Warnf("InsertMonitorData(%v) error:%v", args.Name, err)
		return
	}
	*ret = true
	return
}

func (svc *MonitorDataSvc) UpsertMonitorData(args *model.MonitorDataArg, ret *bool) (err error) {
	clog := MonitorDataCollection(args.Name)
	if clog == nil {
		return
	}

	var existLog model.MonitorData
	err = clog.Find(bson.M{"srvid": args.Log.SrvId, "srvtype": args.Log.SrvType, "key": args.Log.Key}).One(&existLog)
	if err == nil {
		args.Log.LogId = existLog.LogId
		err = clog.Update(bson.M{"_id": args.Log.LogId}, args.Log)
		if err != nil {
			logger.Logger.Warnf("UpsertMonitorData(%v) Update error:%v", args.Name, err)
			return
		}
	} else {
		if err == mgo.ErrNotFound {
			err = clog.Insert(args.Log)
			if err != nil {
				logger.Logger.Warnf("UpsertMonitorData(%v) Insert error:%v", args.Name, err)
				return
			}
		}
	}
	*ret = true
	return
}

func (svc *MonitorDataSvc) RemoveMonitorData(t time.Time, changes *[]*mgo.ChangeInfo) (err error) {
	s := mongo.MgoSessionMgrSington.GetGlobal(model.MonitorDBName)
	if s != nil {
		db := s.DB()
		if db != nil {
			cnames, err := db.CollectionNames()
			if err == nil {
				if len(cnames) != 0 {
					for _, name := range cnames {
						c, _ := db.C(name)
						if c != nil {
							chginfo, err := c.RemoveAll(bson.M{"time": bson.M{"$lt": t}})
							if err == nil {
								*changes = append(*changes, chginfo)
							}
						}
					}
				}
			}
		}
	}
	return
}

func init() {
	//gob registe
	gob.Register(model.PlayerOLStats{})
	gob.Register(map[string]*model.APITransactStats{})
	gob.Register(api.ApiStats{})
	gob.Register(map[string]api.ApiStats{})
	gob.Register(mgo.Stats{})
	gob.Register(profile.TimeElement{})
	gob.Register(map[string]profile.TimeElement{})
	gob.Register(netlib.ServiceStats{})
	gob.Register(map[int]netlib.ServiceStats{})
	gob.Register(schedule.TaskStats{})
	gob.Register(map[string]schedule.TaskStats{})
	gob.Register(transact.TransStats{})
	gob.Register(map[int]transact.TransStats{})
	gob.Register(utils.PanicStackInfo{})
	gob.Register(map[string]utils.PanicStackInfo{})
	gob.Register(basic.CmdStats{})
	gob.Register(map[string]basic.CmdStats{})
	gob.Register(utils.RuntimeStats{})
	//gob registe

	rpc.Register(new(MonitorDataSvc))
}
