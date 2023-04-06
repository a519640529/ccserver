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
	CrashHashAtomDBName   = "user"
	CrashHashAtomCollName = "user_crashhashatom"
	ErrHashAtomDBNotOpen  = model.NewDBError(CrashHashAtomDBName, CrashHashAtomCollName, model.NOT_OPEN)
)

func CrashHashAtomCollection() *mongo.Collection {
	s := mongo.MgoSessionMgrSington.GetPltMgoSession(mongo.G_P, CrashHashAtomDBName)
	if s != nil {
		c, first := s.DB().C(CrashHashAtomCollName)
		if first {
			c.EnsureIndex(mgo.Index{Key: []string{"wheel"}, Background: true, Sparse: true})
		}
		return c
	}
	return nil
}

type CrashHashAtomSvc struct {
}

func (svc *CrashHashAtomSvc) InsertCrashHashAtom(args *model.CrashHashAtom, ret *model.HashRet) error {
	ccrashhashatoms := CrashHashAtomCollection()
	if ccrashhashatoms == nil {
		ret.Tag = 4
		return ErrHashDBNotOpen
	}

	err := ccrashhashatoms.Insert(args)
	if err != nil {
		logger.Logger.Info("InsertCrashHashAtom error:", err)
		ret.Tag = 4
		return nil
	}

	ret.Hash = nil
	ret.Tag = 5
	return nil
}

func (svc *CrashHashAtomSvc) GetCrashHashAtom(args *model.HashIdArg, ret *model.HashRet) error {
	cat := CrashHashAtomCollection()
	if cat == nil {
		return nil
	}
	//err := cat.Find(bson.M{"wheel": args.Wheel}).All(&ret.Hash)
	err := cat.Find(bson.M{}).All(&ret.Hash)
	if err != nil {
		logger.Logger.Error("Get model.GetCrashHashAtom data eror.", err)
		ret.Tag = 4
		return err
	}
	return nil
}

var _CrashHashAtomSvc = &CrashHashAtomSvc{}

func init() {
	rpc.Register(_CrashHashAtomSvc)
}
