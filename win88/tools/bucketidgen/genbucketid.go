package main

import (
	"github.com/globalsign/mgo/bson"
	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/mongo"
	"math/rand"
	"time"
)

var (
	c_playerids       *mongo.Collection
	PlayerIdsDBName   = "user"
	PlayerIdsCollName = "user_bucketids"
)

type PlayerId struct {
	Id       bson.ObjectId `bson:"_id"`
	StartPos int32
	EndPos   int32
	Used     bool
}

func PlayerIdCollection() *mongo.Collection {
	if c_playerids == nil || !c_playerids.IsValid() {
		c_playerids = mongo.DatabaseC(PlayerIdsDBName, PlayerIdsCollName)
		if c_playerids != nil {
			c_playerids.Hold()
		}
	}
	return c_playerids
}
func init() {
	core.RegisteHook(core.HOOK_BEFORE_START, func() error {
		var (
			minV       = 10000000
			maxV       = 99999999
			bucketSize = 100
			bucketArr  []*PlayerId
			gSeedV     = time.Now().UnixNano()
		)
		cnt := ((maxV - minV) / bucketSize) + 1
		bucketArr = make([]*PlayerId, 0, cnt)
		for i := minV; i < maxV; i = i + bucketSize {
			bucketArr = append(bucketArr, &PlayerId{
				Id:       bson.NewObjectId(),
				StartPos: int32(i),
				EndPos:   int32(i + bucketSize - 1),
				Used:     false,
			})
		}
		gSeedV++
		rand.Seed(gSeedV)
		for i, _ := range bucketArr {
			rnd := rand.Intn(len(bucketArr))
			bucketArr[i], bucketArr[rnd] = bucketArr[rnd], bucketArr[i]
		}
		c := PlayerIdCollection()
		docs := make([]interface{}, 0, len(bucketArr))
		for _, log := range bucketArr {
			docs = append(docs, log)
		}
		if c != nil {
			for len(docs) > 0 {
				cnt := len(docs)
				if cnt > 1000 {
					cnt = 1000
				}
				c.Insert(docs[:cnt]...)
				docs = docs[cnt:]
			}
		}
		return nil
	})
}
