package model

import (
	"github.com/globalsign/mgo/bson"
	"github.com/idealeak/goserver/core/logger"
	"time"
)

type CrashHash struct {
	CrashHashId bson.ObjectId `bson:"_id"`
	Wheel       int           //第几轮
	Hash        string        //服务器Hash
}

func NewCrashHash() *CrashHash {
	crashhash := &CrashHash{CrashHashId: bson.NewObjectId()}
	return crashhash
}

type HashIsExistArg struct {
	Wheel int //第几轮
}
type HashRet struct {
	Hash []*CrashHash
	Tag  int
}

func InsertCrashHash(wheel int, hashstr string) (*CrashHash, int) {
	if rpcCli == nil {
		return nil, 4
	}

	hash := NewCrashHash()
	if hash == nil {
		return nil, 4
	}

	hash.Wheel = wheel
	hash.Hash = hashstr

	ret := &HashRet{}
	err := rpcCli.CallWithTimeout("CrashHashSvc.InsertCrashHash", hash, ret, time.Second*30)
	if err != nil {
		return nil, 0
	}
	return nil, ret.Tag
}

type HashIdArg struct {
	Wheel int
}

func GetCrashHash(wheel int) (data []*CrashHash) {
	if rpcCli == nil {
		return
	}
	args := &HashIdArg{
		Wheel: wheel,
	}
	ret := &HashRet{}
	err := rpcCli.CallWithTimeout("CrashHashSvc.GetCrashHash", args, ret, time.Second*30)
	if err != nil {
		logger.Logger.Error("Get GetCrashHash data eror.", err)
	}
	return ret.Hash
}
