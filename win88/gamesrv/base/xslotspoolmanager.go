package base

import (
	"encoding/json"
	"fmt"
	"games.yol.com/win88/common"
	"games.yol.com/win88/model"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/module"
	"strings"
	"time"
)

var XSlotsPoolMgr = &XSlotsPoolManager{
	SlotsPoolData: make(map[string]interface{}),
	SlotsPoolStr:  make(map[string]string),
}

type XSlotsPoolManager struct {
	SlotsPoolData  map[string]interface{}
	SlotsPoolStr   map[string]string
	SlotsPoolDBKey string
}

func (this *XSlotsPoolManager) GetPool(gamefreeId int32, platform string) string {
	key := fmt.Sprintf("%v-%v", gamefreeId, platform)
	if str, exist := this.SlotsPoolStr[key]; exist {
		return str
	}
	return ""
}

func (this *XSlotsPoolManager) SetPool(gamefreeId int32, platform string, pool interface{}) {
	key := fmt.Sprintf("%v-%v", gamefreeId, platform)
	this.SlotsPoolData[key] = pool
}

func (this *XSlotsPoolManager) GetPoolByPlatform(platform string) map[string]interface{} {
	slots := make(map[string]interface{})
	for k, v := range this.SlotsPoolData {
		str := strings.Split(k, "-")
		if len(str) == 2 && str[1] == platform {
			idStr := str[0]
			slots[idStr] = v
		}
	}
	return slots
}

////////////////////////////////////////////////////////////////////
/// Module Implement [beg]
////////////////////////////////////////////////////////////////////
func (this *XSlotsPoolManager) ModuleName() string {
	return "XSlotsPoolManager"
}

func (this *XSlotsPoolManager) Init() {
	this.SlotsPoolDBKey = fmt.Sprintf("XSlotsPoolManager_Srv%v", common.GetSelfSrvId())
	data := model.GetStrKVGameData(this.SlotsPoolDBKey)
	err := json.Unmarshal([]byte(data), &this.SlotsPoolStr)
	if err != nil {
		logger.Logger.Error("Unmarshal slots pool error:", err)
	}
}

func (this *XSlotsPoolManager) Update() {
	this.Save()
}

func (this *XSlotsPoolManager) Shutdown() {
	this.Save()
	module.UnregisteModule(this)
}

func (this *XSlotsPoolManager) Save() {
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
		model.UptStrKVGameData(this.SlotsPoolDBKey, string(buff))
	} else {
		logger.Logger.Error("Marshal coin pool error:", err)
	}
}

func init() {
	module.RegisteModule(XSlotsPoolMgr, time.Minute, 0)
}
