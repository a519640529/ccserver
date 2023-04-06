package redvsblack

import (
	"games.yol.com/win88/gamerule/redvsblack"
	"games.yol.com/win88/gamesrv/base"
)

const (
	RVSB_TOP20      int = 20
	RVSB_OLTOP20        = 20  //在线玩家数量
	RVSB_TREND100       = 100 //近100局的趋势
	RVSB_RICHTOP5       = 5   //富豪top5
	RVSB_BESTWINPOS     = 6   //神算子位置
	RVSB_SELFPOS        = 7   //自己的位置
	RVSB_OLPOS          = 8   //其他在线玩家的位置
)

type RedVsBlackPlayerData struct {
	*base.Player
	betTotal           int                        //当局总下注筹码
	betInfo            [RVSB_ZONE_MAX]int         //记录下注位置及筹码
	winCoin            [RVSB_ZONE_MAX]int         //记录所得金币
	gainWinLost        int                        //本局输赢
	gainCoin           int                        //本局输赢金币  这个含税了
	winRecord          []int                      //最近20局输赢记录
	betBigRecord       []int                      //最近20局下注总额记录
	hadCost            bool                       //是否扣过房费
	lately20Win        int64                      //最近20局输赢总额
	lately20Bet        int64                      //最近20局总下注额
	betDetailInfo      [RVSB_ZONE_MAX]map[int]int //记录下注位置的详细筹码数据
	betCacheInfo       [RVSB_ZONE_MAX]map[int]int //记录2秒内下注位置的详细筹码数据
	betZone            int                        //本局支持哪一方
	taxCoin            int64                      //本局税收
	winorloseCoin      int64                      //本局收税前赢的钱
	result             int32                      //单控结果
	betDetailOrderInfo map[int][]int64            //记录下注位置的详细筹码顺序数据
}

//返回玩家最近20局的获胜次数
func (this *RedVsBlackPlayerData) RecalcuLatestWin20() {
	n := len(this.winRecord)
	if n > RVSB_TOP20 {
		this.winRecord = this.winRecord[n-RVSB_TOP20:]
	}
	this.lately20Win = 0
	if len(this.winRecord) > 0 {
		for _, v := range this.winRecord {
			this.lately20Win += int64(v)
		}
	}
}

//返回玩家最近20局的下注总额
func (this *RedVsBlackPlayerData) RecalcuLatestBet20() {
	n := len(this.betBigRecord)
	if n > RVSB_TOP20 {
		this.betBigRecord = this.betBigRecord[n-RVSB_TOP20:]
	}
	this.lately20Bet = 0
	if len(this.betBigRecord) > 0 {
		for _, v := range this.betBigRecord {
			this.lately20Bet += int64(v)
		}
	}
}

func (this *RedVsBlackPlayerData) Clean() {
	this.betTotal = 0
	this.gainCoin = 0
	this.gainWinLost = 0
	this.hadCost = false
	this.betZone = -1
	this.result = 0
	this.UnmarkFlag(base.PlayerState_SAdjust)

	for i := 0; i < RVSB_ZONE_MAX; i++ {
		this.betInfo[i] = 0
		this.winCoin[i] = 0
		this.betDetailInfo[i] = make(map[int]int)
		this.betCacheInfo[i] = make(map[int]int)
	}
	this.winorloseCoin = 0
	this.taxCoin = 0
	this.betDetailOrderInfo = make(map[int][]int64)
}
func (this *RedVsBlackPlayerData) GetLoseMaxPos(rKind, bKind int) (int, int, int, int) {
	beRWinCoin := this.betInfo[RVSB_ZONE_RED]*2 - int(this.betTotal)
	if rKind > redvsblack.CardsKind_Double && this.betInfo[RVSB_ZONE_LUCKY] > 0 {
		beRWinCoin += (RVSB_LuckyKindOdds[rKind]*100 + 100) * this.betInfo[RVSB_ZONE_LUCKY] / 100
	}
	beBWinCoin := this.betInfo[RVSB_ZONE_BLACK]*2 - int(this.betTotal)
	if bKind > redvsblack.CardsKind_Double && this.betInfo[RVSB_ZONE_LUCKY] > 0 {
		beBWinCoin += (RVSB_LuckyKindOdds[rKind]*100 + 100) * this.betInfo[RVSB_ZONE_LUCKY] / 100
	}
	if beRWinCoin > beBWinCoin {
		return beBWinCoin, RVSB_ZONE_BLACK, this.betInfo[RVSB_ZONE_BLACK], bKind
	} else {
		return beRWinCoin, RVSB_ZONE_RED, this.betInfo[RVSB_ZONE_RED], rKind
	}
}

func (this *RedVsBlackPlayerData) MaxChipCheck(index int, coin int, maxBetCoin []int32) (bool, int32) {
	if index < 0 || index > len(maxBetCoin) {
		return false, -1
	}
	if maxBetCoin[index] > 0 && int32(this.betInfo[index])+int32(coin) > maxBetCoin[index] {
		return false, maxBetCoin[index]
	}
	return true, 0
}
