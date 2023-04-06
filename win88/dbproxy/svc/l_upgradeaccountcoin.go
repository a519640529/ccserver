package svc

import (
	"errors"
	"net/rpc"

	"games.yol.com/win88/dbproxy/mongo"
	"games.yol.com/win88/model"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

var (
	UpgradeAccountCoinLogErr = errors.New("log_upgradeaccountcoin db open failed.")
)

func UpgradeAccountCoinCollection(plt string) *mongo.Collection {
	s := mongo.MgoSessionMgrSington.GetPltMgoSession(plt, model.UpgradeAccountCoinDBName)
	if s != nil {
		c, first := s.DB().C(model.UpgradeAccountCoinCollName)
		if first {
			c.EnsureIndex(mgo.Index{Key: []string{"ip", "date"}, Background: true, Sparse: true})
		}
		return c
	}
	return nil
}

func GetUpgradeAccountCoinLogsByIPAndDate(ip, platform string, date int32) (int32, error) {
	upgradeaccountcoins := UpgradeAccountCoinCollection(platform)
	if upgradeaccountcoins == nil {
		return 0, UpgradeAccountCoinLogErr
	}
	n, err := upgradeaccountcoins.Find(bson.M{"ip": ip, "date": date}).Count()
	return int32(n), err
}

func InsertUpgradeAccountCoinLog(log *model.UpGradeAccountCoin) error {
	upgradeaccountcoins := UpgradeAccountCoinCollection(log.Platform)
	if upgradeaccountcoins == nil {
		return UpgradeAccountCoinLogErr
	}

	return upgradeaccountcoins.Insert(log)
}

type UpGradeAccountCoinSvc struct {
}

func (svc *UpGradeAccountCoinSvc) GetUpgradeAccountCoinLogsByIPAndDate(args *model.UpGradeAccountCoinArgs, count *int32) (err error) {
	*count, err = GetUpgradeAccountCoinLogsByIPAndDate(args.Ip, args.Plt, args.Date)
	return
}

func (svc *UpGradeAccountCoinSvc) InsertUpgradeAccountCoinLog(args *model.UpGradeAccountCoin, ret *bool) (err error) {
	err = InsertUpgradeAccountCoinLog(args)
	if err == nil {
		*ret = true
	}
	return
}

func init() {
	rpc.Register(new(UpGradeAccountCoinSvc))
}
