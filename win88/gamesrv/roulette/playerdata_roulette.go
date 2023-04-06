package roulette

import (
	"games.yol.com/win88/gamerule/roulette"
	rule "games.yol.com/win88/gamerule/roulette"
	"games.yol.com/win88/gamesrv/base"
	"github.com/idealeak/goserver/core/logger"
)

type RoulettePlayerData struct {
	*base.Player                    //玩家信息
	totalBetCoin      int64         //下注金额
	betCnt            map[int][]int //下注筹码 key为座位号码1-157 玩家每个区域下注筹码数量
	betCoin           map[int]int64 //下注筹码 key为座位号码1-157 玩家每个区域下注额
	lately20Win       int           //最近20局输赢总次数
	lately20Bet       int64         //最近20局总下注额
	totalWinRecord    []int         //最近20局输赢记录
	totalBetRecord    []int64       //最近20局下注总额记录
	gainCoin          int64         //玩家总输赢的钱
	taxCoin           int64         //税收
	winCoin           int64         //玩家赢的钱 --给前端显示的
	betRecord         []*BetRecord  //下注记录
	lastBetRecord     map[int][]int //上一局下注记录 Key为下注位置 []int 为下注筹码数量
	lastBetCoin       int64         //上一局下注金额
	proceedBet        bool          //玩家是否续投
	gameRouletteTimes int           //玩家参与游戏次数 只要进入下注阶段就算一局
	tempBetCnt        map[int][]int //当前下注筹码
	winCoinRecord     map[int]int64 //每个区域输赢的钱 --统计使用  //暂时没有用 不用统计每个玩家在每个区域的输赢
}
type BetRecord struct {
	CntIdx []int
	BetPos int
}

func (this *RoulettePlayerData) Clean() {
	this.MarkFlag(base.PlayerState_WaitNext)
	//this.betInfo = make(map[int]int64)
	this.totalBetCoin = 0
	this.gainCoin = 0
	this.taxCoin = 0
	this.winCoin = 0
	this.proceedBet = false
	this.tempBetCnt = make(map[int][]int)

	this.lastBetCoin = 0
	this.lastBetRecord = make(map[int][]int)

	this.lastBetRecord = this.betCnt
	//计算上一局玩家下注总额
	for _, coin := range this.betCoin {
		this.lastBetCoin += coin
	}
	this.betCnt = make(map[int][]int)
	this.betCoin = make(map[int]int64)
	this.winCoinRecord = make(map[int]int64)
	this.betRecord = make([]*BetRecord, 0)
	//计算玩家20局以内的输赢和下注总额
	this.TotalWinAndBet()
}

func (this *RoulettePlayerData) TotalWinAndBet() {
	twr := len(this.totalWinRecord)
	tbr := len(this.totalBetRecord)
	if twr > 20 {
		if this.totalWinRecord[0] > 0 {
			this.lately20Win--
		}
		this.totalWinRecord = this.totalWinRecord[1:]
	}
	if tbr > 20 {
		this.lately20Bet -= this.totalBetRecord[0]
		this.totalBetRecord = this.totalBetRecord[1:]
	}
	if len(this.totalWinRecord) != 0 && this.totalWinRecord[len(this.totalWinRecord)-1] > 0 {
		this.lately20Win++
	}
	if len(this.totalBetRecord) != 0 {
		this.lately20Bet += this.totalBetRecord[len(this.totalBetRecord)-1]
	}
}
func (this *RoulettePlayerData) BetCoinThanMaxLimit(needCoin int64, betCoinChips []int32) (chips [4]int) {
	for k := len(betCoinChips) - 1; k >= 0; k-- {
		if needCoin >= int64(betCoinChips[k]) {
			n := needCoin / int64(betCoinChips[k])
			needCoin -= int64(betCoinChips[k]) * n
			chips[k] += int(n)
		}
	}
	return
}

//下注
func (this *RoulettePlayerData) Bet(params []int64, sceneEx *RouletteSceneData) (bool, int32) {
	n := len(params)
	if n != 2 {
		return false, rule.RoulettePlayerOpError
	}
	otherInts := sceneEx.DbGameFree.GetOtherIntParams()
	if len(otherInts) == 0 {
		return false, rule.RoulettePlayerOpError
	}
	if this.Coin < int64(sceneEx.DbGameFree.GetBetLimit()) {
		logger.Logger.Trace("takeCoin is less. ", int64(sceneEx.DbGameFree.GetBetLimit())-this.Coin)
		return false, rule.RoulettePlayerOpError
	}
	keyNum := int(params[0])
	chipIdx := int(params[1])
	if keyNum < 0 || keyNum > 156 {
		logger.Logger.Trace("bet area is error. ", keyNum)
		return false, rule.RoulettePlayerOpError
	}
	//校验筹码是否正确
	if chipIdx >= len(otherInts) {
		logger.Logger.Tracef("bet chip is error: %v .", chipIdx)
		return false, rule.RoulettePlayerOpError
	}
	chipCoin := int64(otherInts[chipIdx])
	if this.Coin-this.totalBetCoin < chipCoin {
		logger.Logger.Trace("Coin is less. ", chipCoin-(this.Coin-this.totalBetCoin))
		return false, rule.RoulettePlayerOpNotEnoughCoin
	}
	betType := sceneEx.point.GetBetType(keyNum)
	maxBetCoin := sceneEx.DbGameFree.GetMaxBetCoin()[betType]
	pbCoin := int64(0)
	if this.proceedBet {
		if len(this.lastBetRecord) > 0 {
			if data, ok := this.lastBetRecord[keyNum]; ok {
				for k, ct := range data {
					if ct > 0 {
						pbCoin += int64(ct) * int64(otherInts[k])
					}
				}
			}
		}
	}
	chips := [4]int{}
	chips[chipIdx] = 1
	if this.betCoin[keyNum]+chipCoin+pbCoin > int64(maxBetCoin) && maxBetCoin != 0 {
		logger.Logger.Tracef("betCoin %v.MaxBetLimitCoin%v .params%v", chipCoin, maxBetCoin, params)
		chips = this.BetCoinThanMaxLimit(int64(maxBetCoin)-this.betCoin[keyNum]-pbCoin, otherInts)
	}
	this.recordBet(keyNum, chips, sceneEx)
	return true, rule.RoulettePlayerOpSuccess
}
func (this *RoulettePlayerData) recordBet(keyNum int, chips [4]int, sceneEx *RouletteSceneData) {
	//下注记录
	br := &BetRecord{BetPos: keyNum}
	br.CntIdx = chips[:]
	this.betRecord = append(this.betRecord, br)

	otherInts := sceneEx.DbGameFree.GetOtherIntParams()
	if len(otherInts) == 0 {
		return
	}
	if _, ok := this.betCnt[keyNum]; !ok {
		this.betCnt[keyNum] = roulette.InitBet()
	}
	bcs := this.betCnt[keyNum]
	if this.Pos == rule.Roulette_OLPOS {
		if _, ok := sceneEx.betCnt[keyNum]; !ok {
			sceneEx.betCnt[keyNum] = roulette.InitBet()
		}
	}
	scb := sceneEx.betCnt[keyNum]
	for chipIdx, num := range chips {
		if num > 0 {
			nowBetCoin := int64(otherInts[chipIdx]) * int64(num)
			//记录玩家每个区域下注筹码数量
			bcs[chipIdx] += num
			//记录玩家当局每个区域下注额
			this.betCoin[keyNum] += nowBetCoin
			//记录玩家当局总下注额
			this.totalBetCoin += nowBetCoin

			if this.Pos == rule.Roulette_OLPOS {
				scb[chipIdx] += num
			}
			if !this.IsRob {
				sceneEx.realBetCoin[keyNum] += nowBetCoin
			}
		}
	}
}

//撤销上次下注
func (this *RoulettePlayerData) RecallLast(sceneEx *RouletteSceneData) (isType int, betPos int, betInt [4]int) {
	n1 := len(this.betRecord)
	if n1 > 0 {
		otherInts := sceneEx.DbGameFree.GetOtherIntParams()
		if len(otherInts) == 0 {
			return
		}
		br := this.betRecord[n1-1]
		betPos = br.BetPos
		for k, v := range br.CntIdx {
			if v > 0 {
				betInt[k] += v
			}
		}
		this.betRecord = this.betRecord[:n1-1]

		lastBetCoin := int64(0)
		bcp := this.betCnt[betPos]
		for betChipIdx, betNum := range betInt {
			if bcp != nil {
				bcp[betChipIdx] -= betNum
			}
			chipCoin := int64(otherInts[betChipIdx])
			lastBetCoin += chipCoin * int64(betNum)
		}

		this.betCoin[betPos] -= lastBetCoin

		this.totalBetCoin -= lastBetCoin

		if !this.IsRob {
			sceneEx.realBetCoin[betPos] -= lastBetCoin
		}

		isType = 1
	} else if this.proceedBet {
		//撤销续压
		isType = 2
	}
	if this.totalBetCoin == 0 && !this.proceedBet {
		this.MarkFlag(base.PlayerState_WaitNext)
	}
	return
}
