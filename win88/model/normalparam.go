package model

import (
	"encoding/json"
	"io/ioutil"

	"github.com/idealeak/goserver/core/logger"
)

type NormalParam struct {
	//
	LHMin                  int       // = 6000 //龙虎最小几率
	LHMax                  int       // = 8500 //龙虎最大几率
	FishRate               int       // = 10 //捕鱼流水比例
	FishLuckRate           int       // = 10 //捕鱼积分流水比例
	PoolMaxOutRate         int       //=0 //最大出分参考比例
	MaxAddBase             int       //=10 //最大强化概率倍数
	IsCloseAddRtp          int       //=1  //是否关闭附加概率
	NormalAddBase          int       //=5 //正常强化概率倍数
	NewManRate             float64   //=0.5 //是否属于新人
	AttrWeight             []float64 //属性权重 押注权重 历史流水权重 今日流水权重 新手权重
	WinWeight              []float64 //胜率参考权重 历史权重，现在权重
	BetWeight              []float64 //押注参考权重 历史权重，现在权重
	WinMaxRate             float64   //最大胜率
	DLAddRateBase          float64   //= 1.5//龙虎基数
	BetMaxBase             float64   //平均押注额的倍率
	BetCutoffRate          float64   //押注折算比例
	VolatilityMaxRate      float64   //最大波动率调整
	VolatilityGameNum      float64   //波动率调整局数
	AddRateNum             float64   //增加倍数
	BullCardTypeWin        []int     //属性权重 出现赢的概率
	BullCardTypeWin2       []int     //属性权重 出现赢的概率
	RobotRandomTimeMin     int       //机器人随机时间
	RobotRandomTimeMax     int       //机器人随机时间
	RollCoinNeedMin        []int64   //老虎机类需要最小的流水值
	RollCoinMinLine        int32     //不满足流水，最小的线数比例
	IsRollCoinClose        int32     //不满足流水，是否关闭
	HBChangeBankCardRate   int       //百人牛牛庄换牌概率
	HBChangeBankCardLevel  int32     //百人牛牛庄换牌增减值
	HZJHChangeBankCardRate int       //百人金华庄换牌概率
	LHChangeCardRate       int       //龙虎杀牌换牌概率
	RobotVipChangeRate     int       //vip调整概率
	FishBWAddNum           int       //捕鱼黑白名单炮弹计数

}

var NormalParamPath = "../data/normalparam.json"
var NormalParamData = &NormalParam{}

func InitNormalParam() {
	buf, err := ioutil.ReadFile(NormalParamPath)
	if err != nil {
		logger.Logger.Warn("InitNormalParam ioutil.ReadFile error ->", err)
	}

	err = json.Unmarshal(buf, NormalParamData)
	if err != nil {
		logger.Logger.Warn("InitNormalParam json.Unmarshal error ->", err)
	}

	if NormalParamData.LHMin == 0 {
		NormalParamData.LHMin = 6000
	}

	if NormalParamData.LHMax == 0 {
		NormalParamData.LHMax = 8500
	}

	if NormalParamData.FishRate == 0 {
		NormalParamData.FishRate = 5
	}

	if NormalParamData.FishLuckRate == 0 {
		NormalParamData.FishLuckRate = 10
	}

	if NormalParamData.MaxAddBase == 0 {
		NormalParamData.MaxAddBase = 10
	}

	if NormalParamData.NormalAddBase == 0 {
		NormalParamData.NormalAddBase = 5
	}

	if len(NormalParamData.AttrWeight) == 0 {
		NormalParamData.AttrWeight = []float64{0.2, 0.2, 0.5, 0.1}
	}

	if len(NormalParamData.WinWeight) == 0 {
		NormalParamData.WinWeight = []float64{0.3, 0.7}
	}
	if NormalParamData.WinMaxRate <= 1 {
		NormalParamData.WinMaxRate = 4
	}

	if NormalParamData.DLAddRateBase <= 0.1 {
		NormalParamData.DLAddRateBase = 1.8
	}

	if NormalParamData.NewManRate <= 0.0001 {
		NormalParamData.NewManRate = 0.1
	}

	if len(NormalParamData.BetWeight) == 0 {
		NormalParamData.BetWeight = []float64{0.6, 0.4}
	}

	if NormalParamData.BetMaxBase <= 0.0001 {
		NormalParamData.BetMaxBase = 30.0
	}

	if NormalParamData.BetCutoffRate <= 0.0001 {
		NormalParamData.BetCutoffRate = 0.3
	}

	if NormalParamData.VolatilityMaxRate <= 0.0001 {
		NormalParamData.VolatilityMaxRate = 0.5
	}

	if NormalParamData.VolatilityGameNum <= 0.0001 {
		NormalParamData.VolatilityGameNum = 800.0
	}

	if NormalParamData.AddRateNum <= 0.0001 {
		NormalParamData.AddRateNum = 2.2
	}

	if NormalParamData.RobotRandomTimeMin == 0 {
		NormalParamData.RobotRandomTimeMin = 15
	}

	if NormalParamData.RobotRandomTimeMax == 0 {
		NormalParamData.RobotRandomTimeMax = 40
	}

	if NormalParamData.RollCoinMinLine == 0 {
		NormalParamData.RollCoinMinLine = 30
	}

	if NormalParamData.RollCoinMinLine > 100 {
		NormalParamData.RollCoinMinLine = 100
	}

	if len(NormalParamData.RollCoinNeedMin) == 0 {
		NormalParamData.RollCoinNeedMin = []int64{30000, 100000, 300000, 1000000}
	}

	if len(NormalParamData.BullCardTypeWin) == 0 {
		NormalParamData.BullCardTypeWin = []int{1300, 2500, 3000, 2500, 3000}
	}

	if len(NormalParamData.BullCardTypeWin2) == 0 {
		NormalParamData.BullCardTypeWin2 = []int{1300, 2500, 3000, 2500, 3000}
	}

	if NormalParamData.HBChangeBankCardRate == 0 {
		NormalParamData.HBChangeBankCardRate = 500
	}

	if NormalParamData.HBChangeBankCardLevel == 0 {
		NormalParamData.HBChangeBankCardLevel = 1
	}

	if NormalParamData.HZJHChangeBankCardRate == 0 {
		NormalParamData.HZJHChangeBankCardRate = 300
	}

	if NormalParamData.RobotVipChangeRate == 0 {
		NormalParamData.RobotVipChangeRate = 2000
	}

	if NormalParamData.LHChangeCardRate == 0 {
		NormalParamData.LHChangeCardRate = 300
	}

	if NormalParamData.FishBWAddNum == 0 {
		NormalParamData.FishBWAddNum = 20
	}

}
