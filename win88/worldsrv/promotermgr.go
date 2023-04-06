package main

import (
	"encoding/json"
	"games.yol.com/win88/common"
	"games.yol.com/win88/model"
	"games.yol.com/win88/webapi"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/module"
	"strconv"
	"time"
)

// 0 全民 1无限代理 2 渠道 3 推广员
const (
	PROMOTER_TYPE_ALL      = 0 //全民
	PROMOTER_TYPE_INFINITY = 1 //无限代理
	PROMOTER_TYPE_CHANNAL  = 2 //渠道
	PROMOTER_TYPE_PROMOTE  = 3 //推广员
)

var PromoterMgrSington = &PromoterMgr{
	PromoterConfigMap: make(map[string]*PromoterConfig),
	LastTicket:        0,
}

func GetPromoterKeyByInput(key1 string, pType int32) string {
	return key1 + "_" + strconv.Itoa(int(pType))
}

func GetPromoterKey(promoterTree int32, promoter string, channel string) (string, error) {
	var key string
	if promoterTree != 0 {
		key = GetPromoterKeyByInput(strconv.Itoa(int(promoterTree)), PROMOTER_TYPE_INFINITY)
		return key, nil
	} else if promoter != "" {
		key = GetPromoterKeyByInput(promoter, PROMOTER_TYPE_PROMOTE)
		return key, nil

	} else if channel != "" {
		key = GetPromoterKeyByInput(channel, PROMOTER_TYPE_CHANNAL)
		return key, nil

	}

	return key, &ErrorString{code: "no find key"}
}

type PromoterMgr struct {
	PromoterConfigMap map[string]*PromoterConfig
	LastTicket        int64
}

func (this *PromoterMgr) AddConfig(promoter *PromoterConfig) {
	//todo 可能牵涉修改事件之类的在此处理
	this.PromoterConfigMap[promoter.GetKey()] = promoter
}

func (this *PromoterMgr) RemoveConfig(promoter *PromoterConfig) {
	delete(this.PromoterConfigMap, promoter.GetKey())
}

func (this *PromoterMgr) RemoveConfigByKey(promoter string) {
	delete(this.PromoterConfigMap, promoter)
}

func (this *PromoterMgr) GetConfig(promoter string) *PromoterConfig {
	pc, ok := this.PromoterConfigMap[promoter]
	if ok {
		return pc
	}
	return nil
}

func (this *PromoterConfig) GetKey() string {
	key := GetPromoterKeyByInput(this.PromoterID, this.PromoterType)
	return key
}

// //////////////////////////////////////////////////////////////////
// / Module Implement [beg]
// //////////////////////////////////////////////////////////////////
func (this *PromoterMgr) ModuleName() string {
	return "PromoterMgr"
}

type PromoterConfig struct {
	PromoterID             string //代理ID
	Platform               string //平台
	PromoterType           int32  //代理类型 0 全民 1无限代理 2 渠道 3 推广员
	UpgradeAccountGiveCoin int32  //升级账号奖励金币
	NewAccountGiveCoin     int32  //新账号奖励金币
	ExchangeTax            int32  //兑换税收（万分比）
	ExchangeForceTax       int32  //强制兑换税收
	ExchangeFlow           int32  //兑换流水比例
	ExchangeGiveFlow       int32  //赠送兑换流水比例
	ExchangeFlag           int32  //兑换标记
	IsInviteRoot           int32  //是否绑定全民用户

}

func (this *PromoterMgr) Init() {
	this.LastTicket = time.Now().Unix()

	if this.PromoterConfigMap == nil {
		this.PromoterConfigMap = make(map[string]*PromoterConfig)
	}

	type ApiResult struct {
		Tag int
		Msg []PromoterConfig
	}

	//不使用etcd的情况下走api获取
	if !model.GameParamData.UseEtcd {
		buff, err := webapi.API_GetPromoterConfig(common.GetAppId())
		if err == nil {
			ar := ApiResult{}
			err = json.Unmarshal(buff, &ar)
			if err == nil && ar.Tag == 0 {
				logger.Logger.Trace("API_GetPromoterConfig response:", string(buff))
				for _, promoterConfig := range ar.Msg {
					temp := promoterConfig
					this.AddConfig(&temp)
				}
			} else {
				logger.Logger.Error("Unmarshal PromoterMgr data error:", err, " buff:", string(buff))
			}
		} else {
			logger.Logger.Error("Init PromoterMgr list failed.")
		}
	} else {
		EtcdMgrSington.InitPromoterConfig()
	}

}

func (this *PromoterMgr) Update() {

}

func (this *PromoterMgr) Shutdown() {
	module.UnregisteModule(this)
}

func init() {
	module.RegisteModule(PromoterMgrSington, time.Minute, 0)
}
