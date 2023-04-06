package base

import (
	"games.yol.com/win88/common"
	"games.yol.com/win88/model"
	"github.com/globalsign/mgo"
	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/module"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/core/profile"
	"github.com/idealeak/goserver/core/schedule"
	"github.com/idealeak/goserver/core/task"
	"github.com/idealeak/goserver/core/transact"
	"github.com/idealeak/goserver/core/utils"
	"time"
)

var MonitorMgrSington = &MonitorMgr{}

type MonitorMgr struct {
}

func (this *MonitorMgr) ModuleName() string {
	return "MonitorMgr"
}

func (this *MonitorMgr) Init() {

}

func (this *MonitorMgr) Update() {
	//mongodb stats
	mgo.SetStats(true)
	mgoStats := mgo.GetStats()
	logMgo := model.NewMonitorData(int32(common.GetSelfSrvId()), int32(common.GetSelfSrvType()), "", mgoStats)
	if logMgo != nil {
		task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
			return model.InsertMonitorData("mgo", logMgo)
		}), nil, "InsertMonitorData").StartByFixExecutor("monitor")
	}

	//logic stats
	logicStats := profile.GetStats()
	if len(logicStats) > 0 {
		logLogic := model.NewMonitorData(int32(common.GetSelfSrvId()), int32(common.GetSelfSrvType()), "", logicStats)
		if logLogic != nil {
			task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
				return model.UpsertMonitorData("logic", logLogic)
			}), nil, "InsertMonitorData").StartByFixExecutor("monitor")
		}
	}

	//net session stats
	netStats := netlib.Stats()
	if len(netStats) > 0 {
		logNet := model.NewMonitorData(int32(common.GetSelfSrvId()), int32(common.GetSelfSrvType()), "", netStats)
		if logNet != nil {
			task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
				return model.UpsertMonitorData("net", logNet)
			}), nil, "InsertMonitorData").StartByFixExecutor("monitor")
		}
	}

	//schedule stats
	jobStats := schedule.Stats()
	if len(jobStats) > 0 {
		logJob := model.NewMonitorData(int32(common.GetSelfSrvId()), int32(common.GetSelfSrvType()), "", jobStats)
		if logJob != nil {
			task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
				return model.UpsertMonitorData("job", logJob)
			}), nil, "InsertMonitorData").StartByFixExecutor("monitor")
		}
	}

	//trans stats
	transStats := transact.Stats()
	if len(transStats) > 0 {
		logTrans := model.NewMonitorData(int32(common.GetSelfSrvId()), int32(common.GetSelfSrvType()), "", transStats)
		if logTrans != nil {
			task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
				return model.UpsertMonitorData("transact", logTrans)
			}), nil, "InsertMonitorData").StartByFixExecutor("monitor")
		}
	}

	//panic stats
	panicStats := utils.GetPanicStats()
	if len(panicStats) > 0 {
		for key, stats := range panicStats {
			logPanic := model.NewMonitorData(int32(common.GetSelfSrvId()), int32(common.GetSelfSrvType()), key, stats)
			if logPanic != nil {
				task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
					return model.UpsertMonitorData("panic", logPanic)
				}), nil, "InsertMonitorData").StartByFixExecutor("monitor")
			}
		}
	}

	//object command quene stats
	objStats := core.AppCtx.GetStats()
	if len(objStats) > 0 {
		logCmd := model.NewMonitorData(int32(common.GetSelfSrvId()), int32(common.GetSelfSrvType()), "obj", objStats)
		if logCmd != nil {
			task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
				return model.UpsertMonitorData("cmdque", logCmd)
			}), nil, "InsertMonitorData").StartByFixExecutor("monitor")
		}
	}

	//gorouting count, eg. system info
	runtimeStats := utils.StatsRuntime()
	logRuntime := model.NewMonitorData(int32(common.GetSelfSrvId()), int32(common.GetSelfSrvType()), "", runtimeStats)
	if logRuntime != nil {
		task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
			return model.InsertMonitorData("runtime", logRuntime)
		}), nil, "InsertMonitorData").StartByFixExecutor("monitor")
	}
}

func (this *MonitorMgr) Shutdown() {
	module.UnregisteModule(this)
}

func init() {
	module.RegisteModule(MonitorMgrSington, time.Minute*5, 0)

}
