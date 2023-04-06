package crash

import (
	"games.yol.com/win88/gamesrv/base"
)

const (
	CRASH_TOP20      int = 20
	CRASH_OLTOP20        = 20  //在线玩家数量
	CRASH_TREND100       = 100 //近100局的趋势
	CRASH_RICHTOP5       = 5   //富豪top5
	CRASH_BESTWINPOS     = 6   //神算子位置
	CRASH_SELFPOS        = 7   //自己的位置
	CRASH_OLPOS          = 8   //其他在线玩家的位置
)

type CrashPlayerData struct {
	*base.Player
	betTotal      int64           //当局总下注筹码
	multiple      int32           //下注倍率
	//betInfo       map[int32]int64 //记录下注倍数及筹码
	winCoin       map[int32]int64 //记录所得金币
	gainWinLost   int64           //本局输赢 最终加的钱数
	gainCoin      int64           //本局输赢金币  这个含税了
	winRecord     []int64         //最近20局输赢记录
	betBigRecord  []int64         //最近20局下注总额记录
	hadCost       bool            //是否扣过房费
	lately20Win   int64           //最近20局输赢总额
	lately20Bet   int64           //最近20局总下注额
	taxCoin       int64           //本局税收
	winorloseCoin int64           //本局收税前赢的钱
	parachute     bool            //是否跳过伞
}

func (this *CrashPlayerData) Clean() {
	this.betTotal = 0
	this.gainWinLost = 0
	this.gainCoin = 0
	this.multiple = 0
	this.hadCost = false
	this.parachute = false
	this.UnmarkFlag(base.PlayerState_SAdjust)

	//this.betInfo = make(map[int32]int64)
	this.winCoin = make(map[int32]int64)

	this.winorloseCoin = 0
	this.taxCoin = 0
}

//返回玩家最近20局的获胜次数
func (this *CrashPlayerData) RecalcuLatestWin20() {
	n := len(this.winRecord)
	if n > CRASH_TOP20 {
		this.winRecord = this.winRecord[n-CRASH_TOP20:]
	}
	this.lately20Win = 0
	if len(this.winRecord) > 0 {
		for _, v := range this.winRecord {
			this.lately20Win += int64(v)
		}
	}
}

//返回玩家最近20局的下注总额
func (this *CrashPlayerData) RecalcuLatestBet20() {
	n := len(this.betBigRecord)
	if n > CRASH_TOP20 {
		this.betBigRecord = this.betBigRecord[n-CRASH_TOP20:]
	}
	this.lately20Bet = 0
	if len(this.betBigRecord) > 0 {
		for _, v := range this.betBigRecord {
			this.lately20Bet += int64(v)
		}
	}
}
