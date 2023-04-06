package baccarat

import (
	"games.yol.com/win88/gamesrv/base"
)

type BaccaratPlayerData struct {
	*base.Player
	betTotal           int64               //当局总下注筹码
	betInfo            map[int]int         //记录下注位置及筹码
	winCoin            map[int]int         //记录所得金币
	gainWinLost        int64               //本局输赢
	gainCoin           int64               //本局输赢金币
	winRecord          []int64             //最近20局输赢记录
	betBigRecord       []int64             //最近20局下注总额记录
	hadCost            bool                //是否扣过房费
	lately20Win        int64               //最近20局输赢总额
	lately20Bet        int64               //最近20局总下注额
	betDetailInfo      map[int]map[int]int //记录下注位置的详细筹码数据
	betCacheInfo       map[int]map[int]int //记录2秒内下注位置的详细筹码数据
	taxCoin            int64               //本局税收
	winorloseCoin      int64               //本局收税前赢的钱
	lastBetPos         int                 //神算子最后下注位置
	result             int32               //单控结果
	betDetailOrderInfo map[int][]int64     //记录下注位置的详细筹码顺序数据
}

//绑定ai
func (this *BaccaratPlayerData) init() {
	if this.IsLocal && this.IsRob {
		this.AttachAI(&BaccaratPlayerAI{})
	}
}

//返回玩家最近20局的获胜次数
func (this *BaccaratPlayerData) RecalcuLatestWin20() {
	n := len(this.winRecord)
	if n > BACCARAT_OLTOP20 {
		this.winRecord = this.winRecord[n-BACCARAT_OLTOP20:]
	}
	this.lately20Win = 0
	if len(this.winRecord) > 0 {
		for _, v := range this.winRecord {
			this.lately20Win += v
		}
	}
}

//返回玩家最近20局的下注总额
func (this *BaccaratPlayerData) RecalcuLatestBet20() {
	n := len(this.betBigRecord)
	if n > BACCARAT_OLTOP20 {
		this.betBigRecord = this.betBigRecord[n-BACCARAT_OLTOP20:]
	}
	this.lately20Bet = 0
	if len(this.betBigRecord) > 0 {
		for _, v := range this.betBigRecord {
			this.lately20Bet += v
		}
	}
}

func (this *BaccaratPlayerData) Clean() {
	this.betTotal = 0
	this.gainCoin = 0
	this.gainWinLost = 0
	this.hadCost = false
	this.taxCoin = 0
	this.result = 0
	this.winorloseCoin = 0
	this.lastBetPos = -1
	this.UnmarkFlag(base.PlayerState_SAdjust)

	if BaccaratZoneMap == nil {
		return
	}
	this.betInfo = make(map[int]int)
	this.winCoin = make(map[int]int)
	this.betDetailInfo = make(map[int]map[int]int)
	this.betCacheInfo = make(map[int]map[int]int)
	for k := range BaccaratZoneMap {
		this.betDetailInfo[k] = make(map[int]int)
		this.betCacheInfo[k] = make(map[int]int)
	}
	this.betDetailOrderInfo = make(map[int][]int64)
}
func (this *BaccaratPlayerData) MaxChipCheck(index int, coin int, maxBetCoin []int32) (bool, int32) {
	limitIdx := 0
	for n := index >> 1; n > 0; n = n >> 1 {
		limitIdx++
	}
	if limitIdx < 0 || limitIdx > len(maxBetCoin) {
		return false, -1
	}
	if maxBetCoin[limitIdx] > 0 && int32(this.betInfo[index])+int32(coin) > maxBetCoin[limitIdx] {
		return false, maxBetCoin[limitIdx]
	}
	return true, 0
}
