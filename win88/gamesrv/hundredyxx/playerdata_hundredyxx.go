package hundredyxx

import (
	rule "games.yol.com/win88/gamerule/hundredyxx"
	"games.yol.com/win88/gamesrv/base"
)

const (
	HYXX_TOP20         int = 20
	HYXX_OLTOP20           = 20 //在线玩家数量
	HYXX_TREND20           = 20 //近20局的趋势
	HYXX_TRENDNUM          = 40 //近20局的趋势
	HYXX_RICHTOP5          = 5  //富豪top5
	HYXX_BESTWINPOS        = 6  //神算子位置
	HYXX_SELFPOS           = 7  //自己的位置
	HYXX_BANKERPOS         = 8  //庄家位置
	HYXX_OLPOS             = 9  //其他在线玩家的位置
	HYXX_BANKERNUMBERS     = 10
)

//每一个下注区域输赢情况
type HundredYXXBetFieldData struct {
	gainRate  int    //本局倍率
	winFlag   int    //本局输赢 -1:输 1:赢
	gainCoin  int64  //本局输赢金币
	FieldName string //下注区域名称
}

func (this *HundredYXXBetFieldData) Clean() {
	this.gainRate = 0
	this.winFlag = 0
	this.gainCoin = 0
}

type HundredYXXPlayerData struct {
	*base.Player
	betInfo            [rule.BetField_MAX]int64 //记录下注位置及筹码
	winCoin            [rule.BetField_MAX]int64 //记录所得金币
	preBetInfo         [rule.BetField_MAX]int64 //记录上一局下注位置的筹码数据
	gainRate           int                      //本局倍率
	gainWinLost        int                      //本局输赢
	gainCoin           int64                    //本局输赢金币
	returnCoin         int64                    //本局返还金币
	winRecord          []int64                  //最近20局输赢记录
	betBigRecord       []int64                  //最近20局下注总额记录
	cGetWin20          int64                    //返回玩家最近20局的获胜次数
	cGetBetGig20       int64                    //返回玩家最近20局的下注总额
	betTotal           int64                    //当局总下注筹码
	taxCoin            int64                    //本局税收
	winorloseCoin      int64                    //本局收税前赢的钱
	changeScore        int64                    // 本局税后输赢分
	betDetailOrderInfo map[int][]int64          //记录下注位置的详细筹码顺序数据
}

//返回玩家获胜金币
func (this *HundredYXXPlayerData) GetWinCoin() int64 {
	total := int64(0)
	for _, v := range this.winCoin {
		total += v
	}
	return total
}

//返回玩家最近20局的获胜次数
func (this *HundredYXXPlayerData) GetWin20() int64 {
	total := int64(0)
	if len(this.winRecord) > 20 {
		this.winRecord = append(this.winRecord[:0], this.winRecord[1:]...)
	}
	if len(this.winRecord) > 0 {
		for _, v := range this.winRecord {
			total += v
		}
	}
	return total
}

//返回玩家最近20局的下注总额
func (this *HundredYXXPlayerData) GetBetGig20() int64 {
	total := int64(0)
	if len(this.betBigRecord) > 20 {
		this.betBigRecord = append(this.betBigRecord[:0], this.betBigRecord[1:]...)
	}
	if len(this.betBigRecord) > 0 {
		for _, v := range this.betBigRecord {
			total += int64(v)
		}
	}
	return total
}

//返回玩家下注额
func (this *HundredYXXPlayerData) GetBetCount() int64 {
	total := int64(0)
	for _, v := range this.betInfo {
		total += v
	}
	return total
}

func (this *HundredYXXPlayerData) Clean() {
	this.gainRate = 0
	this.gainCoin = 0
	this.gainWinLost = 0
	this.betTotal = 0
	this.taxCoin = 0
	this.winorloseCoin = 0
	this.changeScore = 0

	for i := 0; i < rule.BetField_MAX; i++ {
		this.winCoin[i] = 0
		this.preBetInfo[i] = this.betInfo[i]
		this.betInfo[i] = 0
	}
	if this.winRecord == nil {
		this.winRecord = append(this.winRecord, 0)
	}
	if this.betBigRecord == nil {
		this.betBigRecord = append(this.betBigRecord, 0)
	}
	this.betDetailOrderInfo = make(map[int][]int64)
}

func (this *HundredYXXPlayerData) MaxChipCheck(index int, coin int64, maxBetCoin []int32) (bool, int64) {
	//maxBetCoin 长度为2，分为单门下注上限、双门下注上限
	//if index < 0 || index >= len(maxBetCoin) {
	//	return false, -1
	//}
	maxBet := int32(0)
	if index >= rule.BetField_Single_1 && index <= rule.BetField_Single_6 {
		maxBet = maxBetCoin[0]
	} else if index < rule.BetField_MAX {
		maxBet = maxBetCoin[1]
	} else {
		return false, -1
	}

	if maxBet > 0 && this.betInfo[index]+coin > int64(maxBet) {
		return false, int64(maxBet)
	}
	return true, 0
}
