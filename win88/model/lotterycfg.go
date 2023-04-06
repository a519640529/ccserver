package model

import "encoding/json"

type LotteryConfiger interface {
	HitRatio(card []int32, kind int32) int32
	GetTaxRatio() int32
}

type LotteryConfigParser func(cfg string) (LotteryConfiger, error)

type KindOfCard struct {
	Kind  int32 //牌型
	Ratio int32 //奖池比例,万分比
}

// 彩金配置
type LotteryConfig struct {
	Type  int32        //奖池类型 0:税收
	Ratio int32        //比例 百分比
	Koc   []KindOfCard //特殊牌型获得奖池占比
}

func (this *LotteryConfig) GetTaxRatio() int32 {
	return this.Ratio
}

func (this *LotteryConfig) HitRatio(card []int32, kind int32) int32 {
	//@testcode
	//return 500
	//@testcode
	for i := 0; i < len(this.Koc); i++ {
		if this.Koc[i].Kind == kind {
			return this.Koc[i].Ratio
		}
	}
	return 0
}

func CommonLotteryConfigParser(cfg string) (LotteryConfiger, error) {
	dncfg := &LotteryConfig{}
	err := json.Unmarshal([]byte(cfg), dncfg)
	return dncfg, err
}
