package model

import (
	"github.com/globalsign/mgo/bson"
	"github.com/idealeak/goserver/core/logger"
	"time"
)

type CrashHashAtom struct {
	CrashHashAtomId bson.ObjectId `bson:"_id"`
	Wheel           int           //第几轮
	Hash            string        //服务器Hash
}

func NewCrashHashAtom() *CrashHashAtom {
	crashhash := &CrashHashAtom{CrashHashAtomId: bson.NewObjectId()}
	return crashhash
}

type HashAtomIsExistArg struct {
	Wheel int //第几轮
}
type HashAtomRet struct {
	Hash []*CrashHashAtom
	Tag  int
}

func InsertCrashHashAtom(wheel int, hashstr string) (*CrashHashAtom, int) {
	if rpcCli == nil {
		return nil, 4
	}

	hash := NewCrashHashAtom()
	if hash == nil {
		return nil, 4
	}

	hash.Wheel = wheel
	hash.Hash = hashstr

	ret := &HashAtomRet{}
	err := rpcCli.CallWithTimeout("CrashHashAtomSvc.InsertCrashHashAtom", hash, ret, time.Second*30)
	if err != nil {
		return nil, 0
	}
	return nil, ret.Tag
}

type HashAtomIdArg struct {
	Wheel int
}

//func GetCrashHashAtom(wheel int) (*HashRet, error) {
//	if rpcCli == nil {
//		return nil, ErrRPClientNoConn
//	}
//	args := &HashIdArg{
//		Wheel: wheel,
//	}
//	var ret *HashRet
//	err := rpcCli.CallWithTimeout("CrashHashAtomSvc.GetCrashHashAtom", args, ret, time.Second*30)
//	if err != nil {
//		return nil, err
//	}
//	return ret, err
//}

func GetCrashHashAtom(wheel int) (data []*CrashHashAtom) {
	if rpcCli == nil {
		return
	}
	args := &HashAtomIdArg{
		Wheel: wheel,
	}
	ret := &HashAtomRet{}
	err := rpcCli.CallWithTimeout("CrashHashAtomSvc.GetCrashHashAtom", args, ret, time.Second*30)
	if err != nil {
		logger.Logger.Error("Get GetCrashHashAtom data eror.", err)
	}
	return ret.Hash
}
