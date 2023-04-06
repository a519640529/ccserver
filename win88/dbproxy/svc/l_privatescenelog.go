package svc

import (
	"errors"
	"games.yol.com/win88/dbproxy/mongo"
	"games.yol.com/win88/model"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"net/rpc"
	"time"
)

var (
	PrivateSceneLogDBErr = errors.New("log_privatescene db open failed.")
)

func PrivateSceneLogCollection(plt string) *mongo.Collection {
	s := mongo.MgoSessionMgrSington.GetPltMgoSession(plt, model.PrivateSceneLogDBName)
	if s != nil {
		c, first := s.DB().C(model.PrivateSceneLogCollName)
		if first {
			c.EnsureIndex(mgo.Index{Key: []string{"snid"}, Background: true, Sparse: true})
			c.EnsureIndex(mgo.Index{Key: []string{"channel"}, Background: true, Sparse: true})
			c.EnsureIndex(mgo.Index{Key: []string{"promoter"}, Background: true, Sparse: true})
			c.EnsureIndex(mgo.Index{Key: []string{"packagetag"}, Background: true, Sparse: true})
		}
		return c
	}
	return nil
}

func InsertPrivateSceneLogs(log *model.PrivateSceneLog) error {
	clog := PrivateSceneLogCollection(log.Platform)
	if clog == nil {
		return PrivateSceneLogDBErr
	}

	return clog.Insert(log)
}

func GetPrivateSceneLogBySnId(plt string, snid int32, limit int) ([]*model.PrivateSceneLog, error) {
	clog := PrivateSceneLogCollection(plt)
	if clog == nil {
		return nil, PrivateSceneLogDBErr
	}

	var logs []*model.PrivateSceneLog
	err := clog.Find(bson.M{"snid": snid}).Sort("-createtime").Limit(limit).All(&logs)
	return logs, err
}

func RemovePrivateSceneLogs(plt string, ts time.Time) (*mgo.ChangeInfo, error) {
	return PrivateSceneLogCollection(plt).RemoveAll(bson.M{"createtime": bson.M{"$lt": ts}})
}

type PrivateSceneLogSvc struct {
}

func (svc *PrivateSceneLogSvc) InsertPrivateSceneLogs(log *model.PrivateSceneLog, ret *bool) error {
	err := InsertPrivateSceneLogs(log)
	if err == nil {
		*ret = true
	}
	return err
}

func (svc *RebateLogSvc) GetPrivateSceneLogBySnId(args *model.GetPrivateSceneLogBySnIdArgs, ret *[]*model.PrivateSceneLog) (err error) {
	*ret, err = GetPrivateSceneLogBySnId(args.Plt, args.SnId, args.Limit)
	return err
}

func (svc *RebateLogSvc) RemovePrivateSceneLogs(args *model.RemovePrivateSceneLogsArgs, ret **mgo.ChangeInfo) (err error) {
	*ret, err = RemovePrivateSceneLogs(args.Plt, args.Ts)
	return err
}

func init() {
	rpc.Register(new(PrivateSceneLogSvc))
}
