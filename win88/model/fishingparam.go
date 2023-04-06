package model

import (
	"encoding/json"
	"io/ioutil"

	"github.com/idealeak/goserver/core/logger"
)

type FishingParam struct {
	//2.1 历史流水调整系数相关配置
	Wa    float64 //    = 100000 //历史流水的调整基准，高于此值后历史流水系数不再变化
	Ramax float64 // = 1.3  //历史流水调整系数最大值
	Ramin float64 // = 0.97 //历史流水调整系数最小值

	//2.2 历史输赢调整系数相关配置
	Hmax_chu  float64 //  = 20000
	Hmin_chu  float64 //  = -15000
	Beta_chu  float64 //  = 0.0001
	Alpha_chu float64 // = 0.00015

	RwMax float64 //  = 1.03

	Hmax_zho  float64 //  = 100000
	Hmin_zho  float64 //  = -100000
	Beta_zho  float64 //  = 0.0001
	Alpha_zho float64 // = 0.0001

	Hmax_gao  float64 //  = 200000
	Hmin_gao  float64 //  = -300000
	Beta_gao  float64 //  = 0.0002
	Alpha_gao float64 // = 0.00005

	R1    float64 //    = 1
	Theta float64 // = 0.5

	//2.3 充值行为调整系数相关配置
	P1  float64 //  = 10000   //充值流水分段1
	Rp1 float64 // = 1.01  //在充值后水流在[0，P1]之间的调整系数
	P2  float64 //  = 30000   //充值流水分段2
	Rp2 float64 // = 1.001 //在充值后水流在[P1，P2]之间的调整系数
	Rp3 float64 // = 1     //在充值后水流在[P2，∞]之间的调整系数

	//2.4 每日优惠系数相关配置
	D1  float64 //  = 2000    //日流水分段1
	Rd1 float64 // = 1.01  //日水流在[0，D1)之间的调整系数
	D2  float64 //  = 5000    //日流水分段2
	Rd2 float64 // = 1.001 //日水流在[D1，D2)之间的调整系数
	Rd3 float64 // = 1     //日水流在[D2，∞)之间的调整系数

	//2.5 捕鱼平台调整系数相关配置
	Tmin                            float64          //      = 0       //平台收益底线
	R2                              float64          //        = 1       //平台收益调整系数基准值
	Rtmin                           float64          //     = 0.5     //平台收益调整系数最小值
	Rtmin2                          float64          //     = 0      // 寻红夺宝特有字段
	Delta_chu                       float64          // = 0.00001 //初级场下压系数
	Delta_zho                       float64          // = 0.00001 //中级场下压系数
	Delta_gao                       float64          // = 0.00001 //高级场下压系数
	ReturnCoefficient_chu           float64          // 初级场上限返还系数
	ReturnCoefficient_zho           float64          // 中级场上限返还系数
	ReturnCoefficient_gao           float64          // 高级场上限返还系数
	UpperLimitReturnCoefficient     [][]float64      // 水池平台收益返还
	UpperFishMultiple               [][]ControlParam //鱼倍数百分比上限   (按照权重随机)
	LowerFishMultiple               [][]ControlParam //鱼倍数百分比下限  （按照权重随机）
	UpperFishMultipleFixedValue     [][]ControlParam //鱼倍数百分比上限   (指定固定值)
	LowerFishMultipleFixedValue     [][]ControlParam //鱼倍数百分比下限   (指定固定值)
	PersonalPoolFloatingCoefficient []int32          // 个人水池场此浮动系数
	PersonalPoolInitialValue        []int32          // 个人水池场此初始值
	JackpotRate                     float64          // 天天捕鱼jack奖池抽成
	JackpotOne                      int32            // 天天捕鱼jack奖池一等奖概率
	JackpotTwo                      int32            // 天天捕鱼jack奖池二等奖概率
	JackpotThreeRate                int32            // 天天捕鱼jack奖池三等奖概率
	JackpotFourRate                 int32            // 天天捕鱼jack奖池四等奖概率
	JackpotD1                       int32            // 天天捕鱼jack奖池爆奖系数
	JackpotD2                       int32            // 天天捕鱼jack奖池爆奖系数
	JackpotD3                       int32            // 天天捕鱼jack奖池爆奖系数
	JackpotD4                       int32            // 天天捕鱼jack奖池爆奖系数
	JackpotBate1                    int32            // 奖池分母
	JackpotBate2                    int32            // 奖池分母
	JackpotBate3                    int32            // 奖池分母
	JackpotCoin1                    int64            // 奖池奖励金币
	JackpotCoin2                    int64            // 奖池奖励金币
	JackpotCoin3                    int64            // 奖池奖励金币
	JackpotInitCoin                 int64            //奖池初始金额
	PranaRatio                      float64          // 能量炮抽成比例
	PranaECoin                      int32            // 初级场蓄能上限
	PranaMCoin                      int32            // 中级场蓄能上限
	PranaHCoin                      int32            // 高级场蓄能上限
	FishDealY1                      int32            //致命一击系数Y
	FishDealY2                      int32            //致命一击系数Y
	FishDealY3                      int32            //致命一击系数Y
	FishDealY4                      int32            //致命一击系数Y
	//CommonPoolRatio  float64 // 公共池比例
	//PersonPoolRatio  float64 // 私人池比例
	SLFishRate float64 //特殊鱼L概率
	SHFishRate float64 //特殊鱼H概率
	SignRate   int

	LimiteWinCoin1 int64 //天天捕鱼初级场单条鱼赢取上限
	LimiteWinCoin2 int64 //天天捕鱼中级场单条鱼赢取上限
	LimiteWinCoin3 int64 //天天捕鱼高级场单条鱼赢取上限
	FreeUserLimit  int64 //免费用户金币赢取上限
}

/*
控制百分比
*/
type ControlParam struct {
	Percentage int // 炮数占比
	Weight     int // 对应权重
}

var FishingParamPath = "../data/fishingparam.json"
var FishingParamData = &FishingParam{}

func InitFishingParam() {
	buf, err := ioutil.ReadFile(FishingParamPath)
	if err != nil {
		logger.Logger.Warn("InitFishingParam ioutil.ReadFile error ->", err)
	}

	err = json.Unmarshal(buf, FishingParamData)
	if err != nil {
		logger.Logger.Warn("InitFishingParam json.Unmarshal error ->", err)
	}
	if FishingParamData.Wa == 0 {
		FishingParamData.Wa = 100000 //历史流水的调整基准，高于此值后历史流水系数不再变化
	}
	if FishingParamData.Ramax == 0 {
		FishingParamData.Ramax = 1.3 //历史流水调整系数最大值
	}
	if FishingParamData.Ramin == 0 {
		FishingParamData.Ramin = 0.97 //历史流水调整系数最小值
	}
	//2.2 历史输赢调整系数相关配置
	if FishingParamData.Hmax_chu == 0 {
		FishingParamData.Hmax_chu = 20000
	}
	if FishingParamData.Hmin_chu == 0 {
		FishingParamData.Hmin_chu = -15000
	}
	if FishingParamData.Beta_chu == 0 {
		FishingParamData.Beta_chu = 0.0001
	}
	if FishingParamData.Alpha_chu == 0 {
		FishingParamData.Alpha_chu = 0.00015
	}

	if FishingParamData.RwMax == 0 {
		FishingParamData.RwMax = 1.03
	}

	if FishingParamData.Hmax_zho == 0 {
		FishingParamData.Hmax_zho = 100000
	}
	if FishingParamData.Hmin_zho == 0 {
		FishingParamData.Hmin_zho = -100000
	}
	if FishingParamData.Beta_zho == 0 {
		FishingParamData.Beta_zho = 0.0001
	}
	if FishingParamData.Alpha_zho == 0 {
		FishingParamData.Alpha_zho = 0.0001
	}

	if FishingParamData.Hmax_gao == 0 {
		FishingParamData.Hmax_gao = 200000
	}
	if FishingParamData.Hmin_gao == 0 {
		FishingParamData.Hmin_gao = -300000
	}
	if FishingParamData.Beta_gao == 0 {
		FishingParamData.Beta_gao = 0.0002
	}
	if FishingParamData.Alpha_gao == 0 {
		FishingParamData.Alpha_gao = 0.00005
	}
	if FishingParamData.R1 == 0 {
		FishingParamData.R1 = 1
	}
	if FishingParamData.Theta == 0 {
		FishingParamData.Theta = 0.5
	}
	//2.3 充值行为调整系数相关配置
	if FishingParamData.P1 == 0 {
		FishingParamData.P1 = 10000 //充值流水分段1
	}
	if FishingParamData.Rp1 == 0 {
		FishingParamData.Rp1 = 1.01 //在充值后水流在[0，P1]之间的调整系数
	}
	if FishingParamData.P2 == 0 {
		FishingParamData.P2 = 30000 //充值流水分段2
	}
	if FishingParamData.Rp2 == 0 {
		FishingParamData.Rp2 = 1.001 //在充值后水流在[P1，P2]之间的调整系数
	}
	if FishingParamData.Rp3 == 0 {
		FishingParamData.Rp3 = 1 //在充值后水流在[P2，∞]之间的调整系数
	}
	//2.4 每日优惠系数相关配置
	if FishingParamData.D1 == 0 {
		FishingParamData.D1 = 2000 //日流水分段1
	}
	if FishingParamData.Rd1 == 0 {
		FishingParamData.Rd1 = 1.01 //日水流在[0，D1)之间的调整系数
	}
	if FishingParamData.D2 == 0 {
		FishingParamData.D2 = 5000 //日流水分段2
	}
	if FishingParamData.Rd2 == 0 {
		FishingParamData.Rd2 = 1.001 //日水流在[D1，D2)之间的调整系数
	}
	if FishingParamData.Rd3 == 0 {
		FishingParamData.Rd3 = 1 //日水流在[D2，∞)之间的调整系数
	}
	//2.5 捕鱼平台调整系数相关配置
	if FishingParamData.R2 == 0 {
		FishingParamData.R2 = 1 //平台收益调整系数基准值
	}
	if FishingParamData.Rtmin == 0 {
		FishingParamData.Rtmin = 0.5 //平台收益调整系数最小值
	}
	if FishingParamData.Rtmin2 == 0 {
		FishingParamData.Rtmin2 = 0 //平台收益调整系数最小值
	}
	if FishingParamData.Delta_chu == 0 {
		FishingParamData.Delta_chu = 0.00001 //初级场下压系数
	}
	if FishingParamData.Delta_zho == 0 {
		FishingParamData.Delta_zho = 0.00001 //中级场下压系数
	}
	if FishingParamData.Delta_gao == 0 {
		FishingParamData.Delta_gao = 0.00001 //高级场下压系数
	}
	if FishingParamData.ReturnCoefficient_chu == 0 {
		FishingParamData.ReturnCoefficient_chu = 1 // 初级场返还系数
	}
	if FishingParamData.ReturnCoefficient_zho == 0 {
		FishingParamData.ReturnCoefficient_zho = 1 // 中级场返还系数
	}
	if FishingParamData.ReturnCoefficient_gao == 0 {
		FishingParamData.ReturnCoefficient_gao = 1 // 高级场返还系数
	}
	if len(FishingParamData.UpperLimitReturnCoefficient) == 0 {
		FishingParamData.UpperLimitReturnCoefficient = [][]float64{
			{1.5, 2, 1.5},
			{1.2, 1.5, 1.2},
			{1, 1.2, 1.05},
		}
	}

	if len(FishingParamData.UpperFishMultiple) == 0 {
		FishingParamData.UpperFishMultiple = [][]ControlParam{
			{{140, 8}, {150, 8}, {160, 7}, {170, 7}, {180, 6}, {190, 6}, {200, 4}},
			//{{140, 10}, {150, 8}, {160, 7}, {170, 6}, {180, 6}, {190, 6}, {200, 4}},
			{{140, 8}, {150, 8}, {160, 7}, {170, 7}, {180, 6}, {200, 2}},
		}

	}

	if len(FishingParamData.LowerFishMultiple) == 0 {
		FishingParamData.LowerFishMultiple = [][]ControlParam{
			{{10, 6}, {20, 6}, {30, 7}, {40, 7},
				{50, 8}, {60, 8}, {70, 9}, {80, 9},
				{90, 10}, {100, 10}, {110, 10}, {120, 9}, {130, 9}},
			//{{10, 6}, {20, 6}, {30, 6}, {40, 7},
			//	{50, 8}, {60, 10}, {70, 12}, {80, 15},
			//	{90, 20}, {100, 25}, {110, 20}, {120, 15}, {130, 12}},
			{{20, 6}, {30, 7}, {40, 7}, {50, 8}, {60, 8},
				{70, 9}, {80, 9}, {90, 10}, {100, 10}, {110, 10}, {120, 9}, {130, 9}},
		}
	}
	if len(FishingParamData.PersonalPoolFloatingCoefficient) == 0 {
		FishingParamData.PersonalPoolFloatingCoefficient = []int32{1, 1, 1}
	}

	if len(FishingParamData.PersonalPoolInitialValue) == 0 {
		FishingParamData.PersonalPoolInitialValue = []int32{100, 1000, 10000}
	}

	if FishingParamData.JackpotRate == 0 {
		FishingParamData.JackpotRate = 0.05
	}
	if FishingParamData.JackpotOne == 0 {
		FishingParamData.JackpotOne = 3
	}
	if FishingParamData.JackpotTwo == 0 {
		FishingParamData.JackpotTwo = 5
	}
	if FishingParamData.JackpotThreeRate == 0 {
		FishingParamData.JackpotThreeRate = 12
	}
	if FishingParamData.JackpotFourRate == 0 {
		FishingParamData.JackpotFourRate = 80
	}
	if FishingParamData.JackpotD1 == 0 {
		FishingParamData.JackpotD1 = 5
	}
	if FishingParamData.JackpotD2 == 0 {
		FishingParamData.JackpotD2 = 3
	}
	if FishingParamData.JackpotD3 == 0 {
		FishingParamData.JackpotD1 = 2
	}
	if FishingParamData.JackpotD4 == 0 {
		FishingParamData.JackpotD4 = 1
	}
	if FishingParamData.JackpotBate1 == 0 {
		FishingParamData.JackpotBate1 = 500000
	}
	if FishingParamData.JackpotBate2 == 0 {
		FishingParamData.JackpotBate2 = 5000000
	}
	if FishingParamData.JackpotBate3 == 0 {
		FishingParamData.JackpotBate3 = 50000000
	}
	if FishingParamData.JackpotCoin1 == 0 {
		FishingParamData.JackpotCoin1 = 5000
	}
	if FishingParamData.JackpotCoin2 == 0 {
		FishingParamData.JackpotCoin2 = 50000
	}
	if FishingParamData.JackpotCoin3 == 0 {
		FishingParamData.JackpotCoin3 = 500000
	}
	if FishingParamData.PranaRatio == 0 {
		FishingParamData.PranaRatio = 0.01
	}
	if FishingParamData.PranaECoin == 0 {
		FishingParamData.PranaECoin = 2000
	}
	if FishingParamData.PranaMCoin == 0 {
		FishingParamData.PranaMCoin = 20000
	}
	if FishingParamData.PranaHCoin == 0 {
		FishingParamData.PranaHCoin = 200000
	}
	if FishingParamData.FishDealY1 == 0 {
		FishingParamData.FishDealY1 = 15
	}
	if FishingParamData.FishDealY2 == 0 {
		FishingParamData.FishDealY2 = 13
	}
	if FishingParamData.FishDealY3 == 0 {
		FishingParamData.FishDealY3 = 10
	}

	if FishingParamData.SLFishRate == 0 {
		FishingParamData.SLFishRate = 0.60
	}

	if FishingParamData.SHFishRate == 0 {
		FishingParamData.SHFishRate = 0.80
	}
	if FishingParamData.SignRate == 0 {
		FishingParamData.SignRate = 70
	}
	if FishingParamData.JackpotInitCoin == 0 {
		FishingParamData.JackpotInitCoin = 234982125
		//FishingParamData.JackpotInitCoin = 0
	}
	if FishingParamData.LimiteWinCoin1 == 0 {
		FishingParamData.LimiteWinCoin1 = 20000
	}
	if FishingParamData.LimiteWinCoin2 == 0 {
		FishingParamData.LimiteWinCoin2 = 60000
	}
	if FishingParamData.LimiteWinCoin3 == 0 {
		FishingParamData.LimiteWinCoin3 = 300000
	}
	if FishingParamData.FreeUserLimit == 0 {
		FishingParamData.FreeUserLimit = 2000
	}
}
