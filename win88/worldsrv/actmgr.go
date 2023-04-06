package main

import (
	"games.yol.com/win88/common"
	"games.yol.com/win88/webapi"

	"encoding/json"
	"time"

	"games.yol.com/win88/model"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/module"
)

const ACT_LOGGER = "actmgr"

var ActMgrSington = &ActMgr{
	ConfigByPlateform: make(map[string]*ActGivePlateformConfig),
}

// config --------------------------------------------------------
type ActGiveConfig struct {
	Tag      int32 //赠与类型
	IsStop   int32 //是否阻止 1阻止 0放开
	NeedFlow int32 //需要流水的倍数 0 不计算 其他计算
}

type ActGivePlateformConfig struct {
	ActInfo  map[string]*ActGiveConfig //奖励信息
	findInfo map[int32]*ActGiveConfig  //奖励信息
	Platform string                    //平台
}

type ActMgr struct {
	ConfigByPlateform map[string]*ActGivePlateformConfig
	LastTicket        int64
}

func (this *ActMgr) AddGiveConfig(actComeConfig *ActGivePlateformConfig, plateFrom string) {
	actComeConfig.findInfo = make(map[int32]*ActGiveConfig)
	for _, v := range actComeConfig.ActInfo {
		actComeConfig.findInfo[v.Tag] = v
	}
	this.ConfigByPlateform[plateFrom] = actComeConfig

	this.OnConfigChanged(actComeConfig)
}

func (this *ActMgr) RemovePlateFormConfig(plateFrom string) {
	_, ok := this.ConfigByPlateform[plateFrom]
	if ok {
		delete(this.ConfigByPlateform, plateFrom)
	}
}

func (this *ActMgr) GetIsNeedGive(plateFrom string, tag int32) bool {
	info := this.GetGiveConfig(plateFrom)
	if info == nil {
		return true
	}

	if v, ok := info.findInfo[tag]; ok {
		if v.IsStop == 1 {
			return false
		}
	}
	return true
}

func (this *ActMgr) GetExchangeFlow(plateFrom string, tag int32) int32 {
	info := this.GetGiveConfig(plateFrom)
	if info == nil {
		return 0
	}

	if v, ok := info.findInfo[tag]; ok {
		return v.NeedFlow
	}
	return 0
}

func (this *ActMgr) GetGiveConfig(plateForm string) *ActGivePlateformConfig {
	plateFormConfig, ok := this.ConfigByPlateform[plateForm]
	if ok {
		return plateFormConfig
	}
	return nil
}

func (this *ActMgr) OnConfigChanged(actComeConfig *ActGivePlateformConfig) {

}

// //////////////////////////////////////////////////////////////////
// / Module Implement [beg]
// //////////////////////////////////////////////////////////////////
func (this *ActMgr) ModuleName() string {
	return "ActMgr"
}

func (this *ActMgr) Init() {
	this.LastTicket = time.Now().Unix()

	if this.ConfigByPlateform == nil {
		this.ConfigByPlateform = make(map[string]*ActGivePlateformConfig)
	}

	type ApiResult struct {
		Tag int
		Msg []ActGivePlateformConfig
	}

	//不使用etcd的情况下走api获取
	if !model.GameParamData.UseEtcd {
		buff, err := webapi.API_GetActConfig(common.GetAppId())
		if err == nil {
			ar := ApiResult{}
			err = json.Unmarshal(buff, &ar)
			if err == nil {
				for _, plateformConfig := range ar.Msg {
					t := plateformConfig
					this.AddGiveConfig(&t, plateformConfig.Platform)
				}
			} else {
				logger.Logger.Error("Unmarshal ActMgr data error:", err, " buff:", string(buff))
			}
		} else {
			logger.Logger.Error("Init ActMgr list failed.")
		}
	} else {
		EtcdMgrSington.InitPlatformAct()
	}

}

func (this *ActMgr) Update() {

}

func (this *ActMgr) Shutdown() {
	module.UnregisteModule(this)
}

func init() {
	module.RegisteModule(ActMgrSington, time.Minute, 0)
}
