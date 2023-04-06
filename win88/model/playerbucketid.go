package model

import (
	"sync"
	"time"

	"github.com/globalsign/mgo/bson"
	"github.com/idealeak/goserver/core/logger"
)

var (
	playerBucketId        = &PlayerBucketId{}
	playerBucketIdsCursor = int32(0)
	playerBucketIdLock    = sync.Mutex{}
)

type PlayerBucketId struct {
	Id       bson.ObjectId `bson:"_id"`
	StartPos int32
	EndPos   int32
	Used     bool
}

func (receiver *PlayerBucketId) IsValid() bool {
	return receiver.Id.Valid() && receiver.StartPos != 0 && receiver.EndPos != 0 && receiver.StartPos != receiver.EndPos && playerBucketIdsCursor >= receiver.StartPos && playerBucketIdsCursor <= receiver.EndPos && playerBucketIdsCursor != 0
}

func GetOnePlayerIdFromBucket() (pid int32, err error) {
	playerBucketIdLock.Lock()
	defer playerBucketIdLock.Unlock()

	if !playerBucketId.IsValid() {
		var args struct{}
		err := rpcCli.CallWithTimeout("PlayerBucketIdSvc.GetPlayerOneBucketId", args, &playerBucketId, time.Second*30)
		if err != nil {
			return 0, err
		}
		if !playerBucketId.IsValid() {
			return 0, err
		}
		playerBucketIdsCursor = playerBucketId.StartPos
	}
	pid = playerBucketIdsCursor
	playerBucketIdsCursor++
	return pid, nil
}

func GiveBackPlayerIdBucket() {
	playerBucketIdLock.Lock()
	defer playerBucketIdLock.Unlock()

	if playerBucketId.IsValid() {
		var ret bool
		err := rpcCli.CallWithTimeout("PlayerBucketIdSvc.GiveBackPlayerIdBucket", &playerBucketId, &ret, time.Second*30)
		if err != nil {
			logger.Logger.Error("GiveBackPlayerIdBucket err:", err)
		}
		playerBucketId.StartPos = 0
		playerBucketId.EndPos = 0
	}
}
