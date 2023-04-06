package base

import (
	"fmt"
	"games.yol.com/win88/common"
	"time"

	"games.yol.com/win88/model"
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/module"
	"github.com/idealeak/goserver/core/task"
)

type SysProfitCoinManager struct {
	SysPfCoin *model.SysProfitCoin
}

var SysProfitCoinMgr = &SysProfitCoinManager{}

func (this *SysProfitCoinManager) ModuleName() string {
	return "SysProfitCoinManager"
}

func (this *SysProfitCoinManager) Init() {
	this.SysPfCoin = model.InitSysProfitCoinData(fmt.Sprintf("%d", common.GetSelfSrvId()))
}

func (this *SysProfitCoinManager) GetSysPfCoin(key string) (int64, int64) {
	if data, exist := this.SysPfCoin.ProfitCoin[key]; exist {
		return data.PlaysBet, data.SysPushCoin
	}
	return 1, 0
}

func (this *SysProfitCoinManager) GetSysCommonPool(key string) int64 {
	if data, exist := this.SysPfCoin.ProfitCoin[key]; exist {
		return data.CommonPool
	}
	return 0
}

func (this *SysProfitCoinManager) ChangeSysCommonPool(key string, coin int64) {
	if data, exist := this.SysPfCoin.ProfitCoin[key]; exist {
		data.CommonPool += coin
	}
}

func (this *SysProfitCoinManager) AddSysCommonPool(key string, coin int64) {
	data, ok := this.SysPfCoin.ProfitCoin[key]
	if !ok {
		data = new(model.SysCoin)
		this.SysPfCoin.ProfitCoin[key] = data
	}
	data.CommonPool += coin
}

func (this *SysProfitCoinManager) Update() {
	this.Save()
}

func (this *SysProfitCoinManager) Shutdown() {
	this.Save()
	module.UnregisteModule(this)
}

func (this *SysProfitCoinManager) Save() {
	data := &model.SysProfitCoin{ //
		ProfitCoin: make(map[string]*model.SysCoin),
		LogId:      this.SysPfCoin.LogId,
		Key:        this.SysPfCoin.Key,
	}
	for i, v := range this.SysPfCoin.ProfitCoin {
		v1 := &model.SysCoin{ //
			PlaysBet:    v.PlaysBet,
			SysPushCoin: v.SysPushCoin,
			CommonPool:  v.CommonPool,
			Version:     v.Version,
		}
		data.ProfitCoin[i] = v1
	}
	task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
		err := model.SaveSysProfitCoin(data /*this.SysPfCoin*/) //  fatal error: concurrent map iteration and map write
		if err != nil {
			logger.Logger.Errorf("SaveSysProfitCoin err:%v ", err)
		}
		return err
	}), nil, "SaveSysProfitCoin").Start()
}

//系统投入产出初始值
func (this *SysProfitCoinManager) InitData(gameid, scenetype int32, key string, ctroRate int64) {
	if this.SysPfCoin == nil {
		return
	}
	if _, ok := this.SysPfCoin.ProfitCoin[key]; ok {
		return
	}
	var syscoin = new(model.SysCoin)
	this.SysPfCoin.ProfitCoin[key] = syscoin
}

// Add 增加玩家总投入，系统总产出
func (this *SysProfitCoinManager) Add(key string, bet int64, systemOut int64) {
	data, ok := this.SysPfCoin.ProfitCoin[key]
	if !ok {
		data = new(model.SysCoin)
		this.SysPfCoin.ProfitCoin[key] = data
	}
	data.PlaysBet += bet
	data.SysPushCoin += systemOut
}

func (this *SysProfitCoinManager) Get(key string) *model.SysCoin {
	data, ok := this.SysPfCoin.ProfitCoin[key]
	if !ok {
		data = new(model.SysCoin)
		this.SysPfCoin.ProfitCoin[key] = data
	}
	return data
}

//del
func (this *SysProfitCoinManager) Del(key string) {
	if _, ok := this.SysPfCoin.ProfitCoin[key]; ok {
		delete(this.SysPfCoin.ProfitCoin, key)
	}
}
func init() {
	module.RegisteModule(SysProfitCoinMgr, time.Hour, 0)
}
