package main

import (
	"encoding/gob"
	"games.yol.com/win88/common"
	"games.yol.com/win88/model"
	"games.yol.com/win88/webapi"
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
	//player online stats
	olStats := PlayerMgrSington.StatsOnline()
	log := model.NewMonitorData(int32(common.GetSelfSrvId()), int32(common.GetSelfSrvType()), "", olStats)
	if log != nil {
		task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
			return model.InsertMonitorData("online", log)
		}), nil, "InsertMonitorData").StartByFixExecutor("monitor")
	}

	//api stats
	if len(WebAPIStats) > 0 {
		log := model.NewMonitorData(int32(common.GetSelfSrvId()), int32(common.GetSelfSrvType()), "", WebAPIStats)
		if log != nil {
			task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
				return model.UpsertMonitorData("webapi", log)
			}), nil, "InsertMonitorData").StartByFixExecutor("monitor")
		}
	}
	apiStats := webapi.Stats()
	if len(apiStats) > 0 {
		log := model.NewMonitorData(int32(common.GetSelfSrvId()), int32(common.GetSelfSrvType()), "api", apiStats)
		if log != nil {
			task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
				return model.UpsertMonitorData("webapi", log)
			}), nil, "InsertMonitorData").StartByFixExecutor("monitor")
		}
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

	//gob registe
	gob.Register(model.PlayerOLStats{})
	gob.Register(map[string]*model.APITransactStats{})
	gob.Register(webapi.ApiStats{})
	gob.Register(map[string]webapi.ApiStats{})
	gob.Register(profile.TimeElement{})
	gob.Register(map[string]profile.TimeElement{})
	gob.Register(netlib.ServiceStats{})
	gob.Register(map[int]netlib.ServiceStats{})
	gob.Register(schedule.TaskStats{})
	gob.Register(map[string]schedule.TaskStats{})
	gob.Register(transact.TransStats{})
	gob.Register(map[int]transact.TransStats{})
	gob.Register(utils.PanicStackInfo{})
	gob.Register(map[string]utils.PanicStackInfo{})
	gob.Register(basic.CmdStats{})
	gob.Register(map[string]basic.CmdStats{})
	gob.Register(utils.RuntimeStats{})
	//gob registe

	module.RegisteModule(MonitorMgrSington, time.Minute*5, 0)

}
