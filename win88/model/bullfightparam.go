package model

import (
	"encoding/json"
	"io/ioutil"

	"github.com/idealeak/goserver/core/logger"
)

type BullFightParam struct {
	//
	N             int32   // = 10 //循环周期
	PayUserLost   float64 // = 30  //充值用户忍耐倍数
	NoPayUserLost float64 // = 50 //非充值用户忍耐倍数
	PayCoinBase   float64 // = 2000 //充值增加次数倍数
	BaseRate      float64 // = 4  //默认系数

	BankModifyRate    float64 // = 500 //输赢修正比例
	BankModifyPowRate float64 // = 3 //输赢修正比例
	BankModifyMaxRate float64 // = 5000 //最大修正比例

}

var BullFightParamPath = "../data/bullfightparam.json"
var BullFightParamData = &BullFightParam{}

func InitBullFightParam() {
	buf, err := ioutil.ReadFile(BullFightParamPath)
	if err != nil {
		logger.Logger.Warn("InitBullFightParam ioutil.ReadFile error ->", err)
	}

	err = json.Unmarshal(buf, BullFightParamData)
	if err != nil {
		logger.Logger.Warn("InitBullFightParam json.Unmarshal error ->", err)
	}
	if BullFightParamData.N == 0 {
		BullFightParamData.N = 10
	}

	if BullFightParamData.PayUserLost == 0 {
		BullFightParamData.PayUserLost = 30
	}

	if BullFightParamData.NoPayUserLost == 0 {
		BullFightParamData.NoPayUserLost = 50
	}

	if BullFightParamData.BaseRate == 0 {
		BullFightParamData.BaseRate = 4
	}

	if BullFightParamData.PayCoinBase == 0 {
		BullFightParamData.PayCoinBase = 4
	}

	if BullFightParamData.BankModifyRate == 0 {
		BullFightParamData.BankModifyRate = 500
	}

	if BullFightParamData.BankModifyPowRate == 0 {
		BullFightParamData.BankModifyPowRate = 3
	}

	if BullFightParamData.BankModifyMaxRate == 0 {
		BullFightParamData.BankModifyMaxRate = 5000
	}

}
