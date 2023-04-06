package base

import (
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/task"
	"time"

	"encoding/json"
	"fmt"

	"games.yol.com/win88/common"
	"games.yol.com/win88/model"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/module"
)

var SlotsPoolMgr = &SlotsPoolManager{
	SlotsPoolData: make(map[string]interface{}),
	SlotsPoolStr:  make(map[string]string),
}

type SlotsPoolManager struct {
	SlotsPoolData  map[string]interface{}
	SlotsPoolStr   map[string]string
	SlotsPoolDBKey string
}

func (this *SlotsPoolManager) GetPool(gamefreeId int32, platform string) string {
	key := fmt.Sprintf("%v-%v", gamefreeId, platform)
	if str, exist := this.SlotsPoolStr[key]; exist {
		return str
	}
	return ""
}

func (this *SlotsPoolManager) SetPool(gamefreeId int32, platform string, pool interface{}) {
	key := fmt.Sprintf("%v-%v", gamefreeId, platform)
	this.SlotsPoolData[key] = pool
}
func (this *SlotsPoolManager) GetSlotsPoolData(gamefreeId int32, platform string) interface{} {
	key := fmt.Sprintf("%v-%v", gamefreeId, platform)
	return this.SlotsPoolData[key]
}

// //////////////////////////////////////////////////////////////////
// / Module Implement [beg]
// //////////////////////////////////////////////////////////////////
func (this *SlotsPoolManager) ModuleName() string {
	return "SlotsPoolManager"
}

func (this *SlotsPoolManager) Init() {
	this.SlotsPoolDBKey = fmt.Sprintf("SlotsPoolManager_Srv%v", common.GetSelfSrvId())
	data := model.GetStrKVGameData(this.SlotsPoolDBKey)
	err := json.Unmarshal([]byte(data), &this.SlotsPoolStr)
	if err != nil {
		logger.Logger.Error("Unmarshal slots pool error:", err)
	}
}

func (this *SlotsPoolManager) Update() {
	this.Save()
}

func (this *SlotsPoolManager) Shutdown() {
	this.Save()
	module.UnregisteModule(this)
}

func (this *SlotsPoolManager) Save() {
	//数据先整合
	for k, v := range this.SlotsPoolData {
		data, err := json.Marshal(v)
		if err == nil {
			this.SlotsPoolStr[k] = string(data)
		}
	}

	//再保存
	buff, err := json.Marshal(this.SlotsPoolStr)
	if err == nil {
		task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
			return model.UptStrKVGameData(this.SlotsPoolDBKey, string(buff))
		}), nil, "SaveActGoldCome").StartByFixExecutor("UptStrKVGameData")
	} else {
		logger.Logger.Error("Marshal coin pool error:", err)
	}
}

func init() {
	module.RegisteModule(SlotsPoolMgr, time.Minute, 0)
}
