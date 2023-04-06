package svc

import (
	"errors"
	"net/rpc"
	"time"

	"games.yol.com/win88/dbproxy/mongo"
	"games.yol.com/win88/model"
	"github.com/globalsign/mgo"
)

var RobotIdDBErr = errors.New("user_robotidpool db open failed.")

func RobotIdCollection() *mongo.Collection {
	s := mongo.MgoSessionMgrSington.GetPltMgoSession(mongo.G_P, model.RobotIdPoolDBName)
	if s != nil {
		c, first := s.DB().C(model.RobotIdPoolCollName)
		if first {
		}
		return c
	}
	return nil
}

func PrefetchRobotIds(expectCnt int) ([]*model.RobotIdPool, error) {
	c := RobotIdCollection()
	if c != nil {
		var data []*model.RobotIdPool
		err := c.Find(nil).All(&data)
		if err == nil || err == mgo.ErrNotFound {
			cnt := len(data)
			if cnt < expectCnt {
				pool, err := AllocBatchRobotIds(expectCnt-cnt)
				if err != nil {
					return data, err
				}
				data = append(data, pool...)
			}
			return data, nil
		}
		return nil, err
	}
	return nil, RobotIdDBErr
}

func AllocBatchRobotIds(cnt int) (ret []*model.RobotIdPool, err error) {
	c := RobotIdCollection()
	if c != nil {
		ret = make([]*model.RobotIdPool, 0, cnt)
		for i := 0; i < cnt; i++ {
			var id model.PlayerBucketId
			err = _PlayerBucketIdSvc.GetPlayerOneBucketId(struct{}{}, &id)
			if err != nil {
				return
			}
			pool := &model.RobotIdPool{
				Id:         id.Id,
				StartPos:   id.StartPos,
				EndPos:     id.EndPos,
				CreateTime: time.Now(),
			}
			err = c.Insert(pool)
			if err != nil {
				return
			}
			ret = append(ret, pool)
		}
		return
	}
	return nil, RobotIdDBErr
}

type RobotIdPoolSvc struct {
}

func (svc *RobotIdPoolSvc) PrefetchRobotIds(expectCnt int, ret *[]*model.RobotIdPool) (err error) {
	*ret, err = PrefetchRobotIds(expectCnt)
	return
}

func (svc *RobotIdPoolSvc) AllocBatchRobotIds(cnt int, ret *[]*model.RobotIdPool) (err error) {
	*ret, err = AllocBatchRobotIds(cnt)
	return
}

func init() {
	rpc.Register(new(RobotIdPoolSvc))
}
