package svc

import (
	"games.yol.com/win88/dbproxy/mongo"
	"games.yol.com/win88/model"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/idealeak/goserver/core/logger"
	"math/rand"
	"net/rpc"
	"sync"
	"time"
)

const (
	PLAYERBUCKETID_MIN       = 10000000
	BATCH_PLAYERBUCKETID_CNT = 99999999
)

var (
	PlayerBucketIdsDBName       = "user"
	PlayerBucketIdsCollName     = "user_bucketids"
	PlayerBucketIdsAutoId       = bson.ObjectIdHex("60f69b4a09dbe3323c632f25")
	ErrPlayerBucketIdsDBNotOpen = model.NewDBError(PlayerBucketIdsDBName, PlayerBucketIdsCollName, model.NOT_OPEN)
	playerBucketId              = &model.PlayerBucketId{}
	playerBucketIdsCursor       = int32(0)
	playerBucketIdLock          = sync.Mutex{}
)

type PlayerBucketAutoId struct {
	Id     bson.ObjectId `bson:"_id"`
	AutoId int
}

func PlayerBucketIdCollection() *mongo.Collection {
	s := mongo.MgoSessionMgrSington.GetPltMgoSession(mongo.G_P, PlayerBucketIdsDBName)
	if s != nil {
		c, first := s.DB().C(PlayerBucketIdsCollName)
		if first {
			var id PlayerBucketAutoId
			err := c.Find(bson.M{"_id": PlayerBucketIdsAutoId}).One(&id)
			if err != nil && err == mgo.ErrNotFound {
				id.Id = PlayerBucketIdsAutoId
				id.AutoId = PLAYERBUCKETID_MIN
				c.Insert(id)
			}
		}
		return c
	}
	return nil
}

type PlayerBucketIdSvc struct {
}

func (svc *PlayerBucketIdSvc) GetPlayerOneBucketId(dummy struct{}, ret *model.PlayerBucketId) (err error) {
	c := PlayerBucketIdCollection()
	if c != nil {
		var flag bool
		change := mgo.Change{
			Update:    bson.M{"$set": bson.M{"used": true}},
			ReturnNew: true,
		}
	redo:
		_, err := c.Find(bson.M{"used": false}).Apply(change, ret)
		if err != nil {
			logger.Logger.Warnf("GetOnePlayerIdFromBucket Find failed:%v", err)
			if !flag && err == mgo.ErrNotFound {
				err = genABatchPlayerBucketId()
				flag = true
				if err == nil {
					goto redo
				}
			}
			return err
		}
		err = c.RemoveId(ret.Id)
		if err != nil {
			logger.Logger.Warnf("GetOnePlayerIdFromBucket RemoveId(%v) failed:%v ", ret.Id, err)
		}
		return err
	}
	return ErrPlayerBucketIdsDBNotOpen
}

func (svc *PlayerBucketIdSvc) GiveBackPlayerIdBucket(args *model.PlayerBucketId, ret *bool) error {
	c := PlayerBucketIdCollection()
	if c != nil {
		_, err := c.UpsertId(args.Id, args)
		if err != nil {
			logger.Logger.Warnf("GiveBackPlayerIdBucket UpsertId(%v) failed:%v", args, err)
			return err
		}
		*ret = true
		return nil
	}

	return ErrPlayerBucketIdsDBNotOpen
}

func genABatchPlayerBucketId() error {
	maxV, err := autoIncPlayerBucketId()
	if err != nil {
		return err
	}
	var (
		minV       = maxV - BATCH_PLAYERBUCKETID_CNT
		bucketSize = 100
		bucketArr  []*model.PlayerBucketId
		gSeedV     = time.Now().UnixNano()
	)
	cnt := ((maxV - minV) / bucketSize) + 1
	bucketArr = make([]*model.PlayerBucketId, 0, cnt)
	for i := minV; i < maxV; i = i + bucketSize {
		bucketArr = append(bucketArr, &model.PlayerBucketId{
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
	c := PlayerBucketIdCollection()
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
			err = c.Insert(docs[:cnt]...)
			docs = docs[cnt:]
		}
	}
	return err
}

func GetOnePlayerIdFromBucket() (pid int32, err error) {
	playerBucketIdLock.Lock()
	defer playerBucketIdLock.Unlock()

	if !playerBucketId.IsValid() {
		var flag bool
		c := PlayerBucketIdCollection()
		if c != nil {
			change := mgo.Change{
				Update:    bson.M{"$set": bson.M{"used": true}},
				ReturnNew: true,
			}
		redo:
			_, err := c.Find(bson.M{"used": false}).Apply(change, playerBucketId)
			if err != nil {
				logger.Logger.Warnf("GetOnePlayerIdFromBucket Find failed:%v", err)
				if !flag && err == mgo.ErrNotFound {
					err = genABatchPlayerBucketId()
					flag = true
					if err == nil {
						goto redo
					}
				}
				return 0, err
			}
			err = c.RemoveId(playerBucketId.Id)
			if err != nil {
				logger.Logger.Warnf("GetOnePlayerIdFromBucket RemoveId(%v) failed:%v ", playerBucketId.Id, err)
			}
			playerBucketIdsCursor = playerBucketId.StartPos
		} else {
			return 0, ErrPlayerBucketIdsDBNotOpen
		}
	}
	pid = playerBucketIdsCursor
	playerBucketIdsCursor++
	return pid, nil
}

func GiveBackPlayerIdBucket() {
	playerBucketIdLock.Lock()
	defer playerBucketIdLock.Unlock()

	if playerBucketId.IsValid() {
		c := PlayerBucketIdCollection()
		if c != nil {
			_, err := c.UpsertId(playerBucketId.Id, playerBucketId)
			if err != nil {
				logger.Logger.Warnf("GiveBackPlayerIdBucket UpsertId(%v) failed:%v", playerBucketId, err)
			}
		}
	}
}

func autoIncPlayerBucketId() (int, error) {
	c := PlayerBucketIdCollection()
	if c == nil {
		return 0, ErrPlayerBucketIdsDBNotOpen
	}
	change := mgo.Change{
		Update:    bson.M{"$inc": bson.M{"autoid": BATCH_PLAYERBUCKETID_CNT}},
		ReturnNew: true,
	}
	doc := PlayerBucketAutoId{}
	_, err := c.Find(bson.M{"_id": PlayerBucketIdsAutoId}).Apply(change, &doc)
	if err == nil {
		return doc.AutoId, nil
	}
	return 0, err
}

var _PlayerBucketIdSvc = &PlayerBucketIdSvc{}

func init() {
	rpc.Register(_PlayerBucketIdSvc)
}
