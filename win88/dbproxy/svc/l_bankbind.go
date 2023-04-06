package svc

import (
	"errors"
	"net/rpc"

	"games.yol.com/win88/dbproxy/mongo"
	"games.yol.com/win88/model"
	"github.com/globalsign/mgo"
)

var (
	BankBindLogDBErr = errors.New("log_bankbind db open failed.")
)

func BankBindCollection(plt string) *mongo.Collection {
	s := mongo.MgoSessionMgrSington.GetPltMgoSession(plt, model.BankBindDBName)
	if s != nil {
		c, first := s.DB().C(model.BankBindCollName)
		if first {
			c.EnsureIndex(mgo.Index{Key: []string{"snid"}, Background: true, Sparse: true})
			c.EnsureIndex(mgo.Index{Key: []string{"bankname"}, Background: true, Sparse: true})
			c.EnsureIndex(mgo.Index{Key: []string{"bankcard"}, Background: true, Sparse: true})
		}
		return c
	}

	return nil
}

func InsertBankBindLog(log *model.BankBindLog) error {
	clog := BankBindCollection(log.Platform)
	if clog == nil {
		return BankBindLogDBErr
	}
	return clog.Insert(log)
}

type BankBindLogSvc struct {
}

func (svc *BankBindLogSvc) InsertBankBindLog(log *model.BankBindLog, ret *bool) error {
	err := InsertBankBindLog(log)
	if err == nil {
		*ret = true
	}
	return err
}

func init() {
	rpc.Register(new(BankBindLogSvc))
}
