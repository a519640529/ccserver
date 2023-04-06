package model

import (
	"github.com/globalsign/mgo/bson"
	"time"
)

const (
	PREFETCH_ROBOT_ID  = 100
	NEXTBATCH_ROBOT_ID = 10
)

var (
	RobotIdPoolDBName   = "user"
	RobotIdPoolCollName = "user_robotidpool"
)

type RobotIdPool struct {
	Id         bson.ObjectId `bson:"_id"`
	StartPos   int32
	EndPos     int32
	CreateTime time.Time
}

func (id *RobotIdPool) Ids() []int32 {
	ids := make([]int32, 0, id.EndPos-id.StartPos+1)
	for i := id.StartPos; i <= id.EndPos; i++ {
		ids = append(ids, i)
	}
	return ids
}

func (id *RobotIdPool) Fill(buf []int32) []int32 {
	for i := id.StartPos; i <= id.EndPos; i++ {
		buf = append(buf, i)
	}
	return buf
}

func PrefetchRobotIds() (ids []*RobotIdPool, err error) {
	if rpcCli == nil {
		return nil, ErrRPClientNoConn
	}
	err = rpcCli.CallWithTimeout("RobotIdPoolSvc.PrefetchRobotIds", PREFETCH_ROBOT_ID, &ids, time.Second*30)
	return
}

func AllocBatchRobotIds() (ids []*RobotIdPool, err error) {
	if rpcCli == nil {
		return nil, ErrRPClientNoConn
	}
	err = rpcCli.CallWithTimeout("RobotIdPoolSvc.AllocBatchRobotIds", NEXTBATCH_ROBOT_ID, &ids, time.Second*30)
	return
}
