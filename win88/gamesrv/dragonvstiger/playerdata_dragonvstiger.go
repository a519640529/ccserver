package dragonvstiger

import (
	"games.yol.com/win88/common"
	. "games.yol.com/win88/gamerule/dragonvstiger"
	"games.yol.com/win88/gamesrv/base"
	"sort"
)

const (
	DVST_TOP20         int = 20
	DVST_OLTOP20           = 20
	DVST_TREND100          = 100
	DVST_RICHTOP5          = 5
	DVST_BESTWINPOS        = 6
	DVST_SELFPOS           = 7
	DVST_OLPOS             = 8
	DVST_BANKERPOS         = 9
	DVST_BANKERNUMBERS     = 10
)

type DragonVsTigerPlayerData struct {
	*base.Player
	betTotal           int64
	betInfo            [DVST_ZONE_MAX]int
	winCoin            [DVST_ZONE_MAX]int
	gainWinLost        int64
	gainCoin           int64
	winRecord          []int64
	betBigRecord       []int64
	hadCost            bool
	lately20Win        int64
	lately20Bet        int64
	betDetailInfo      [DVST_ZONE_MAX]map[int]int
	betCacheInfo       [DVST_ZONE_MAX]map[int]int
	taxCoin            int64
	winorloseCoin      int64
	lastBetPos         int
	result             int32           //单控结果
	betDetailOrderInfo map[int][]int64 //记录下注位置的详细筹码顺序数据
}
type DVTPDSort struct {
	coin int
	flag int
}

func (this *DragonVsTigerPlayerData) RecalcuLatestWin20() {
	n := len(this.winRecord)
	if n > DVST_TOP20 {
		this.winRecord = this.winRecord[n-DVST_TOP20:]
	}
	this.lately20Win = 0
	if len(this.winRecord) > 0 {
		for _, v := range this.winRecord {
			this.lately20Win += v
		}
	}
}
func (this *DragonVsTigerPlayerData) RecalcuLatestBet20() {
	n := len(this.betBigRecord)
	if n > DVST_TOP20 {
		this.betBigRecord = this.betBigRecord[n-DVST_TOP20:]
	}
	this.lately20Bet = 0
	if len(this.betBigRecord) > 0 {
		for _, v := range this.betBigRecord {
			this.lately20Bet += v
		}
	}
}
func (this *DragonVsTigerPlayerData) Clean() {
	this.betTotal = 0
	this.gainCoin = 0
	this.gainWinLost = 0
	this.hadCost = false
	this.taxCoin = 0
	this.winorloseCoin = 0
	this.lastBetPos = -1
	this.result = 0
	this.UnmarkFlag(base.PlayerState_SAdjust)

	for i := 0; i < DVST_ZONE_MAX; i++ {
		this.betInfo[i] = 0
		this.winCoin[i] = 0
		this.betDetailInfo[i] = make(map[int]int)
		this.betCacheInfo[i] = make(map[int]int)
	}
	this.betDetailOrderInfo = make(map[int][]int64)
}
func (this *DragonVsTigerPlayerData) GetLoseMaxPos() (int, int, int) {
	sortData := []DVTPDSort{}
	beDWinCoin := this.betInfo[DVST_ZONE_DRAGON]*2 - int(this.betTotal)
	sortData = append(sortData, DVTPDSort{coin: beDWinCoin, flag: DVST_ZONE_DRAGON})
	beTWinCoin := this.betInfo[DVST_ZONE_TIGER]*2 - int(this.betTotal)
	sortData = append(sortData, DVTPDSort{coin: beTWinCoin, flag: DVST_ZONE_TIGER})
	beHWinCoin := this.betInfo[DVST_ZONE_DRAW]*8 - int(this.betTotal) +
		(this.betInfo[DVST_ZONE_TIGER] + this.betInfo[DVST_ZONE_DRAGON])
	sortData = append(sortData, DVTPDSort{coin: beHWinCoin, flag: DVST_ZONE_DRAW})
	randIndex := common.RandInt(DVST_ZONE_MAX)
	sortData[0], sortData[randIndex] = sortData[randIndex], sortData[0]
	sort.Slice(sortData, func(i, j int) bool {
		return sortData[i].coin < sortData[j].coin
	})
	if sortData[0].flag == DVST_ZONE_DRAW {
		if sortData[0].coin == sortData[1].coin {
			if common.RandInt(10000) > 700 {
				return sortData[1].coin, sortData[1].flag, this.betInfo[sortData[1].flag]
			}
		}
	}
	return sortData[0].coin, sortData[0].flag, this.betInfo[sortData[0].flag]
}

func (this *DragonVsTigerPlayerData) MaxChipCheck(index int, coin int, maxBetCoin []int32) (bool, int32) {
	if index < 0 || index > len(maxBetCoin) {
		return false, -1
	}
	if maxBetCoin[index] > 0 && int32(this.betInfo[index])+int32(coin) > maxBetCoin[index] {
		return false, maxBetCoin[index]
	}
	return true, 0
}
