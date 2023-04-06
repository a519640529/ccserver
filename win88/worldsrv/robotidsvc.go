package main

import (
	"errors"
	"math/rand"

	"games.yol.com/win88/model"
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/core/task"
)

const (
	ROBOTID_THRESHOLD = 1000
)

var ErrNotEnoughRobotIds = errors.New("not enough robot ids")
var LocalRobotIdMgrSington = &LocalRobotIdMgr{}

type LocalRobotIdMgr struct {
	freeIds []int32
}

func (mgr *LocalRobotIdMgr) Init() {
	ids, err := model.PrefetchRobotIds()
	if err == nil {
		cnt := 0
		for _, id := range ids {
			cnt += int(id.EndPos - id.StartPos + 1)
		}
		mgr.freeIds = make([]int32, 0, cnt)
		for _, id := range ids {
			mgr.freeIds = id.Fill(mgr.freeIds)
		}
		mgr.freeIds = ShuffleRobotIds(mgr.freeIds)
	}
}

func (mgr *LocalRobotIdMgr) checkThreshold() {
	if len(mgr.freeIds) < ROBOTID_THRESHOLD {
		t, done := task.NewMutexTask(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
			ids, err := model.AllocBatchRobotIds()
			var ret []int32
			if err == nil {
				cnt := 0
				for _, id := range ids {
					cnt += int(id.EndPos - id.StartPos + 1)
				}
				ret = make([]int32, 0, cnt)
				for _, id := range ids {
					ret = id.Fill(ret)
				}
			}
			return ShuffleRobotIds(ret)
		}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
			if ids, ok := data.([]int32); ok {
				mgr.freeIds = append(mgr.freeIds, ids...)
			}
		}), "AllocBatchRobotIds", "AllocBatchRobotIds")
		if !done {
			t.Start()
		}
	}
}

func (mgr *LocalRobotIdMgr) FreeId(snid int32) {
	mgr.freeIds = append(mgr.freeIds, snid)
	cnt := len(mgr.freeIds)
	idx := rand.Intn(cnt)
	mgr.freeIds[cnt-1], mgr.freeIds[idx] = mgr.freeIds[idx], mgr.freeIds[cnt-1]
}

func (mgr *LocalRobotIdMgr) GetIds(cnt int) ([]int32, error) {
	defer mgr.checkThreshold()
	n := len(mgr.freeIds)
	if n >= cnt {
		ids := make([]int32, 0, cnt)
		ids = append(ids, mgr.freeIds[n-cnt:]...)
		mgr.freeIds = mgr.freeIds[:n-cnt]
		return ids, nil
	}
	return nil, ErrNotEnoughRobotIds
}

func ShuffleRobotIds(ids []int32) []int32 {
	for i := 0; i < len(ids); i++ {
		j := rand.Intn(i + 1)
		ids[i], ids[j] = ids[j], ids[i]
	}
	return ids
}

type LocalRobotIdSvc struct {
}

func (svc *LocalRobotIdSvc) GetIds(cnt int, ids *[]int32) (err error) {
	*ids, err = LocalRobotIdMgrSington.GetIds(cnt)
	return
}

func init() {
	netlib.RegisterRpc(&LocalRobotIdSvc{})
}
