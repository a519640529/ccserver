package roulette

import "time"

//场景状态
const (
	RouletteSceneStateWait      int = iota //等待状态
	RouletteSceneStateBet                  //下注
	RouletteSceneStateOpenPrize            //开奖
	RouletteSceneStateBilled               //结算
	RouletteSceneStateStart                //开始倒计时
	RouletteSceneStateMax
)

//玩家操作
const (
	RoulettePlayerOpReady       int = iota //准备
	RoulettePlayerOpCancelReady            //取消准备
	RoulettePlayerOpKickout                //踢人
	RoulettePlayerOpBet                    //下注
	RoulettePlayerOpRecall                 //撤销
	RoulettePlayerOpPlayerList             //玩家列表
	RoulettePlayerOpProceedBet             //续投
)
const (
	RouletteSceneWaitTimeout = time.Second * 2  //等待倒计时
	RouletteStartTimeout     = time.Second * 6  //开始倒计时
	RouletteBetTimeout       = time.Second * 15 //下注
	RouletteOpenPrizeTimeout = time.Second * 7  //开奖
	RouletteBilledTimeout    = time.Second * 3  //结算
)
const (
	Roulette_RICHTOP1   = 1 //富豪no.1
	Roulette_RICHTOP2   = 3 //富豪no.2
	Roulette_BESTWINPOS = 6 //神算子位置
	Roulette_SELFPOS    = 7 //自己的位置
	Roulette_OLPOS      = 8 //其他在线玩家的位置
)
const (
	RoulettePlayerOpSuccess            int32 = iota //成功
	RoulettePlayerOpBetCoinThanLimit                //押注超过上限
	RoulettePlayerOpNotEnoughCoin                   //金币不足
	RoulettePlayerOpAlreadyProceedBet               //已经续压
	RoulettePlayerOpNoBetCoinNotRecall              //没有下注不能撤销
	RoulettePlayerOpError                           //失败
)
const (
	BetTypeStraight  int = iota //0.直接注(0-36个号码)
	BetTypeSplit                //1.分注
	BetTypeStreet               //2.街注(1列3个数)/三个号码
	BetTypeCorner               //3.角注(4个数位方框)/四个号码
	BetTypeLine                 //4.线注(2列6个数)
	BetTypeThreeSide            //5.三面 2to1 --- 1st12 2nd12 3rd12
	BetTypeTwoSide              //6.双面 红 黑 单 双 高 低

//BetTypeStraight int = iota //直接注(0-36个号码)
//BetTypeSplit               //分注(非0的相邻2个数)
//BetTypeStreet              //街注(1列3个数)
//BetTypeThree               //三数({0,1,2} {0,2,3})
//BetTypeCorner              //角注(4个数位方框)
//BetTypeFour                //四个号码(0-3)
//BetTypeLine                //线注(2列6个数)
//BetTypeColumn              //直行注(第一行:3.+3.36 第二行:2.+3.35 第三行:1.+3.34)
//BetTypeDozen               //打注(第一打:1-12 第二打:13-24 第三打:25-36)
//BetTypeBlack               //黑注(0通杀)
//BetTypeRed                 //红注(0通杀)
//BetTypeOdd                 //奇数注(0通杀)
//BetTypeEven                //偶数注(0通杀)
//BetTypeLow                 //低注(1-18 0通杀)
//BetTypeHi                  //高注(19-36 0通杀)
)
const (
	PointLow int = iota + 43
	PointHi
	PointDouble
	PointSingle
	PointRed
	PointBlack
)

type PointType struct {
	RateMap      []int                //倍率
	pointMap     map[int][]int        //位置详细值  0-157  0-36    每一个位置对应什么号码
	PointTypeMap map[int]map[int]bool //类型对应的位置 0-6类型  0-157位置 每一个类型对应什么位置
	PointMapNums map[int]map[int]bool //数字对应的位置号码 0-36   0-157  每一个号码对应什么位置
}

//四个下注区域初始化
func InitBet() []int {
	return make([]int, 4)
}
func (this *PointType) Init() {
	this.RateMap = []int{35, 17, 11, 8, 5, 2, 1}
	this.setPointMap()
	this.setPointSlice()
	this.setWinPointMap()
}

//计算当前下注位置号码 是什么类型  key 为0-157 位置号码
func (this *PointType) GetBetType(key int) (betType int) {
	betType = -1
	//for k, v := range this.PointTypeMap {
	//	for _, n := range v {
	//		if key == n {
	//			betType = k
	//			return
	//		}
	//	}
	//}
	for k, v := range this.PointTypeMap {
		if _, ok := v[key]; ok {
			betType = k
			break
		}
	}
	return
}

//设置中奖的位置号码集合 0-157
func (this *PointType) setWinPointMap() {
	this.PointMapNums = make(map[int]map[int]bool)
	for n := 0; n <= 36; n++ {
		pointMapNums := make(map[int]bool)
		if n == 0 {
			//pointMapNums = []int{0, 72, 73, 74, 97, 98, 99}
			pointMapNums[0] = true
			pointMapNums[72] = true
			pointMapNums[73] = true
			pointMapNums[74] = true
			pointMapNums[97] = true
			pointMapNums[98] = true
			pointMapNums[99] = true
		} else {
			for pm, v := range this.pointMap {
				pushBool := false
				switch pm {
				case 37:
					//列
					if n%3 == 0 {
						pushBool = true
					}
				case 38:
					if n%3 == 2 {
						pushBool = true
					}
				case 39:
					if n%3 == 1 {
						pushBool = true
					}
				case 40:
					//打
					if n >= 1 && n <= 12 {
						pushBool = true
					}
				case 41:
					if n >= 13 && n <= 24 {
						pushBool = true
					}
				case 42:
					if n >= 25 && n <= 36 {
						pushBool = true
					}
				case PointLow: //小
					if n >= 1 && n <= 18 {
						pushBool = true
					}
				case PointHi: //大
					if n >= 19 && n <= 36 {
						pushBool = true
					}
				case PointDouble: //双
					if n%2 == 0 {
						pushBool = true
					}
				case PointSingle: //单
					if n%2 == 1 {
						pushBool = true
					}
				}
				if pm < 37 || pm > 46 {
					for _, m := range v {
						if m == n {
							pushBool = true
							break
						}
					}
				}
				if pushBool {
					//pointMapNums = append(pointMapNums, pm)
					pointMapNums[pm] = true
				}
			}
		}
		this.PointMapNums[n] = pointMapNums
	}

}
func (this *PointType) setPointSlice() {
	this.PointTypeMap = make(map[int]map[int]bool)
	for i := 0; i < 157; i++ {
		betType := BetTypeStraight
		if i >= 0 && i <= 36 {
			//直接注 35
			betType = BetTypeStraight
		} else if i >= 37 && i <= 42 {
			//三面 2
			betType = BetTypeThreeSide
		} else if i >= 43 && i <= 48 {
			//双面 1
			betType = BetTypeTwoSide
		} else if i >= 49 && i <= 59 {
			//线注 5
			betType = BetTypeLine
		} else if i >= 60 && i <= 73 {
			//街注/3个号码 11
			betType = BetTypeStreet
		} else if i >= 74 && i <= 96 {
			//角注/4个号码 8
			betType = BetTypeCorner
		} else if i >= 97 {
			//分注 2个号码 17
			betType = BetTypeSplit
		}
		//this.PointTypeMap[betType] = append(this.PointTypeMap[betType], i)
		if data, ok := this.PointTypeMap[betType]; ok {
			data[i] = true
		} else {
			this.PointTypeMap[betType] = make(map[int]bool)
			this.PointTypeMap[betType][i] = true
		}
	}
}
func (this *PointType) setPointMap() {
	this.pointMap = make(map[int][]int)
	//0-36
	for i := 0; i <= 36; i++ {
		this.pointMap[i] = []int{i}
	}
	// 列
	this.pointMap[37] = []int{3 /*6, 9, 12, 15, 18, 21, 24, 27, 30, 33,*/, 36} //3
	this.pointMap[38] = []int{2 /*5, 8, 11, 14, 17, 20, 23, 26, 29, 32,*/, 35} //2
	this.pointMap[39] = []int{1 /*4, 7, 10, 13, 16, 19, 22, 25, 28, 31,*/, 34} //1
	// 打
	this.pointMap[40] = []int{1 /*2, 3, 4, 5, 6, 7, 8, 9, 10, 11,*/, 12}          //1
	this.pointMap[41] = []int{13 /*14, 15, 16, 17, 18, 20, 21, 22, 23,*/, 24}     //2
	this.pointMap[42] = []int{25 /*26, 27, 28, 29, 30, 31, 32, 33, 34, 35,*/, 36} //3

	//<=18小
	this.pointMap[43] = []int{1 /*2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17,*/, 18}

	//>=19大
	this.pointMap[44] = []int{19 /*20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35,*/, 36}

	//双数
	this.pointMap[45] = []int{2 /*4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32, 34,*/, 36}
	//单数
	this.pointMap[46] = []int{1 /* 3, 5, 7, 9, 11, 13, 15, 17, 19, 21, 23, 25, 27, 29, 31, 33,*/, 35}

	//红
	this.pointMap[47] = []int{1, 3, 5, 7, 9, 12, 14, 16, 18, 19, 21, 23, 25, 27, 30, 32, 34, 36}
	//黑
	this.pointMap[48] = []int{2, 4, 6, 8, 10, 11, 13, 15, 17, 20, 22, 24, 26, 28, 29, 31, 33, 35}

	//1 2 3 4 5 6
	for i := 1; i <= 11; i++ {
		this.pointMap[48+i] = []int{1 + 3*(i-1), 2 + 3*(i-1), 3 + 3*(i-1), 4 + 3*(i-1), 5 + 3*(i-1), 6 + 3*(i-1)}
	}
	// 1 2 3
	for i := 1; i <= 12; i++ {
		this.pointMap[59+i] = []int{1 + 3*(i-1), 2 + 3*(i-1), 3 + 3*(i-1)}
	}
	this.pointMap[72] = []int{0, 1, 2}
	this.pointMap[73] = []int{0, 2, 3}

	//四
	//0 1 2 3
	this.pointMap[74] = []int{0, 1, 2, 3}

	//1 2 4 5
	for i := 1; i <= 11; i++ {
		this.pointMap[74+i] = []int{1 + 3*(i-1), 2 + 3*(i-1), 4 + 3*(i-1), 5 + 3*(i-1)}
	}

	for i := 1; i <= 11; i++ {
		this.pointMap[85+i] = []int{2 + 3*(i-1), 3 + 3*(i-1), 5 + 3*(i-1), 6 + 3*(i-1)}
	}

	//二
	this.pointMap[97] = []int{0, 1}
	this.pointMap[98] = []int{0, 2}
	this.pointMap[99] = []int{0, 3}

	for i := 1; i <= 12; i++ {
		this.pointMap[99+i] = []int{1 + 3*(i-1), 2 + 3*(i-1)}
	}
	for i := 1; i <= 12; i++ {
		this.pointMap[111+i] = []int{2 + 3*(i-1), 3 * i}
	}
	for i := 1; i <= 11; i++ {
		this.pointMap[123+i] = []int{1 + 3*(i-1), 4 + 3*(i-1)}
	}
	for i := 1; i <= 11; i++ {
		this.pointMap[134+i] = []int{2 + 3*(i-1), 5 + 3*(i-1)}
	}
	for i := 1; i <= 11; i++ {
		this.pointMap[145+i] = []int{3 + 3*(i-1), 6 + 3*(i-1)}
	}
}
