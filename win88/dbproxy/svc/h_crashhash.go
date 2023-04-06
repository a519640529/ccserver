package svc

import (
	"games.yol.com/win88/dbproxy/mongo"
	"games.yol.com/win88/model"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/idealeak/goserver/core/logger"
	"net/rpc"
)

var (
	CrashHashDBName   = "user"
	CrashHashCollName = "user_crashhash"
	ErrHashDBNotOpen  = model.NewDBError(CrashHashDBName, CrashHashCollName, model.NOT_OPEN)
)

func CrashHashCollection() *mongo.Collection {
	s := mongo.MgoSessionMgrSington.GetPltMgoSession(mongo.G_P, CrashHashDBName)
	if s != nil {
		c, first := s.DB().C(CrashHashCollName)
		if first {
			c.EnsureIndex(mgo.Index{Key: []string{"wheel"}, Background: true, Sparse: true})
		}
		return c
	}
	return nil
}

type CrashHashSvc struct {
}

func (svc *CrashHashSvc) InsertCrashHash(args *model.CrashHash, ret *model.HashRet) error {
	ccrashhashs := CrashHashCollection()
	if ccrashhashs == nil {
		ret.Tag = 4
		return ErrHashDBNotOpen
	}

	err := ccrashhashs.Insert(args)
	if err != nil {
		logger.Logger.Info("InsertCrashHash error:", err)
		ret.Tag = 4
		return nil
	}

	ret.Hash = nil
	ret.Tag = 5
	return nil
}

func (svc *CrashHashSvc) GetCrashHash(args *model.HashIdArg, ret *model.HashRet) error {
	cat := CrashHashCollection()
	if cat == nil {
		return nil
	}
	//err := cat.Find(bson.M{"wheel": args.Wheel}).All(&ret.Hash)
	err := cat.Find(bson.M{}).All(&ret.Hash)
	if err != nil {
		logger.Logger.Error("Get model.GetCrashHash data eror.", err)
		ret.Tag = 4
		return err
	}
	return nil
}

var _CrashHashSvc = &CrashHashSvc{}

func init() {
	rpc.Register(_CrashHashSvc)
}
