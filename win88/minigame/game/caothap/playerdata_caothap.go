package caothap

import (
	proto_caothap "games.yol.com/win88/protocol/caothap"
	"math/rand"
	"time"

	"games.yol.com/win88/gamerule/caothap"
	"games.yol.com/win88/minigame/base"
	"games.yol.com/win88/model"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/timer"
)

type CaoThapPlayerData struct {
	*base.Player
	score            int32                         //押注数
	cards            []int32                       //已翻牌数据
	RollGameType     *model.CaoThapType            //记录信息
	enterGameCoin    int64                         //玩家进入初始金币
	taxCoin          int64                         //本局税收
	winCoin          int64                         //本局收税前赢的钱
	leavetime        int32                         //用户离开时间
	cardsData        []int32                       // 牌库
	cardPos          int                           // 当前牌索引
	lastJackpotTime  time.Time                     // 最后一次爆奖时间
	step             int                           // 游戏阶段
	cardCount        int32                         // 牌的数量
	currentTurnID    int32                         // 本局游戏标志
	betValue         int64                         // 下注金额
	prizeValue       int64                         // 赢取金额
	CaoThapCardAData                               // A牌数据
	betRateUp        int64                         // 压大赢的概率
	betRateDown      int64                         // 压小赢的概率
	bigWinScore      int64                         //压大赢分
	littleWinScore   int64                         //压小赢分
	remainTime       int64                         // 剩余时间，单位 s
	isEven           bool                          // 牌值是否相等
	isJackpot        bool                          // 是否爆奖池
	jackpotValue     int64                         // 爆奖分数
	playerOpHandler  timer.TimerHandle             // 操作定时器
	betIdx           int32                         //下注筹码索引
	billedData       *proto_caothap.GameBilledData //上一局结算信息

	//测试
	debugJackpot bool
}

type CaoThapCardAData struct {
	currentAces         []int32   // 牌A数据
	currentAcesQuantity int       // 牌A数量
	createdTimePlay     time.Time // 创建时间
}

//玩家初始化
func (this *CaoThapPlayerData) init(s *base.Scene) {
	this.score = s.DbGameFree.OtherIntParams[0] // 底注
	this.betIdx = 0
	this.RollGameType = &model.CaoThapType{}
	this.enterGameCoin = this.Coin
	this.cards = make([]int32, 0)
	this.cardsData = make([]int32, caothap.CARDDATANUM)
	this.cardPos = 0
	this.lastJackpotTime = time.Now().Add(-caothap.JACKPOTTIMEINTERVAL * time.Hour)
	this.step = 0
	this.cardCount = 0
	this.currentAces = make([]int32, 0)
	this.currentAcesQuantity = 0
	this.betValue = 0
	this.prizeValue = 0
	this.createdTimePlay = time.Now()
	this.betRateUp = 0
	this.betRateDown = 0
	this.remainTime = 0
	this.isEven = false
	this.isJackpot = false
	this.jackpotValue = 0
	this.playerOpHandler = timer.TimerHandle(0)
	this.LastOPTimer = time.Now()
	this.billedData = &proto_caothap.GameBilledData{}
	this.debugJackpot = true
}

//黑白名单的限制是否生效
func (this *CaoThapPlayerData) CheckBlackWriteList(isWin bool) bool {
	if isWin && this.BlackLevel > 0 && this.BlackLevel <= 10 {
		rand.Seed(time.Now().UnixNano())
		if rand.Int31n(100) < this.BlackLevel*10 {
			return true
		}
	} else if !isWin && this.WhiteLevel > 0 && this.WhiteLevel <= 10 {
		rand.Seed(time.Now().UnixNano())
		if rand.Int31n(100) < this.WhiteLevel*10 {
			return true
		}
	}
	return false
}

// ROUND(@_Rate*(1 + 1.000 * @_CardDownValue/ @_CardUpValue), 2)
// 计算上下赢分概率, 返回的值是原值的100倍
func (this *CaoThapPlayerData) CalcBetRate(baseRate int32) (betRateUp, betRateDown int64) {
	var totalCardUp, totalCardDown, totalCurrentUp, totalCurrentDown int32
	currentCardValue := caothap.GetCardValue(this.cardsData[this.cardPos])

	for i := this.cardPos; i < len(this.cardsData); i++ {
		cardValue := caothap.GetCardValue(this.cardsData[i])
		if cardValue > currentCardValue {
			totalCardUp++
		}
		if cardValue < currentCardValue {
			totalCardDown++
		}
	}

	for _, v := range this.cards {
		cardValue := caothap.GetCardValue(v)
		if cardValue > currentCardValue {
			totalCurrentUp++
		}
		if cardValue < currentCardValue {
			totalCurrentDown++
		}
	}
	cardUpValue := float64(totalCardUp - totalCurrentUp)
	cardDownValue := float64(totalCardDown - totalCurrentDown)
	if cardUpValue == 0 {
		betRateUp = 0
		betRateDown = 1
	} else if cardDownValue == 0 {
		betRateUp = 1
		betRateDown = 0
	} else {
		up := float64(baseRate) * (cardDownValue / cardUpValue) / 10000
		betRateUp = int64(up * 100)
		down := float64(baseRate) * (cardUpValue / cardDownValue) / 10000
		betRateDown = int64(down * 100)
	}
	//if cardUpValue == 0 {
	//	betRateUp = 0
	//} else {
	//	//up := float64(baseRate) *  (1.0 + cardDownValue/cardUpValue) / 10000
	//	//betRateUp = int64(up*100 + 0.5)
	//	up := float64(baseRate) * (cardDownValue / cardUpValue) / 10000
	//	betRateUp = int64(up * 100)
	//}
	//if cardDownValue == 0 {
	//	betRateDown = 0
	//} else {
	//	//down := float64(baseRate) * (1.0 + cardUpValue/cardDownValue) / 10000
	//	//betRateDown = int64(down*100 + 0.5)
	//	down := float64(baseRate) * (cardUpValue / cardDownValue) / 10000
	//	betRateDown = int64(down * 100)
	//}
	//if betRateUp > 0 && betRateUp < 100 {
	//	betRateUp = 100
	//}
	//if betRateDown > 0 && betRateDown < 100 {
	//	betRateDown = 100
	//}
	//if betRateUp > 0 && betRateUp < 1 {
	//	betRateUp = 1
	//}
	//if betRateDown > 0 && betRateDown < 1 {
	//	betRateDown = 1
	//}
	this.cardPos++
	return
}

func (this *CaoThapPlayerData) CalcPrizeValue(betValue, spinCondition, locationID int64) (prizeValue int64) {
	// 水池不足以支付玩家
	currentCard := this.cardsData[this.cardPos]
	// 本次下注为上次赢分
	lastCard := this.cardsData[this.cardPos-1]
	lastCardValue := caothap.GetCardValue(lastCard)
	// checkMin := int64(jackpotParam[caothap.CAOTHAP_JACKPOT_LIMITWIN_PRIZELOW])
	// checkMax := int64(jackpotParam[caothap.CAOTHAP_JACKPOT_LIMITWIN_PRIZEHIGH])
	//for i := this.cardPos; i < caothap.CARDDATANUM-this.cardPos; i++ {
	for i := this.cardPos; i < caothap.CARDDATANUM; i++ {
		curretCardValue := caothap.GetCardValue(this.cardsData[i])
		switch {
		case lastCardValue == curretCardValue:
			prizeValue = betValue * 9 / 10
		case lastCardValue > curretCardValue && locationID == CaoThapLitte:
			//prizeValue = betValue * this.betRateDown / 100
			prizeValue = this.littleWinScore
		case lastCardValue < curretCardValue && locationID == CaoThapBig:
			//prizeValue = betValue * this.betRateUp / 100
			prizeValue = this.bigWinScore
		default:
			prizeValue = 0
		}
		if currentCard < 0 {
			newPos := this.GetCardPos(locationID)
			this.SwitchCardPos(this.cardPos, newPos)
			break
		}
		// 校验水池浮动值
		// if prizeValue > checkMax || prizeValue < checkMin {
		// 	continue
		// }
		// 校验A牌
		if caothap.IsCardA(this.cardsData[i]) && this.currentAcesQuantity == 2 && this.cardCount < 7 {
			//if this.debugJackpot {
			//	this.debugJackpot = false
			//	break
			//}
			continue
		}
		// A牌校验水池
		if caothap.IsCardA(this.cardsData[i]) && this.currentAcesQuantity == 2 && spinCondition-prizeValue < 0 {
			//if this.debugJackpot {
			//	this.debugJackpot = false
			//	break
			//}
			continue
		}
		// 校验水池
		if spinCondition-prizeValue < 0 {
			newPos := this.GetCardPos(locationID)
			this.SwitchCardPos(this.cardPos, newPos)
		} else {
			this.SwitchCardPos(this.cardPos, i)
		}
		break
	}
	currentCard = this.cardsData[this.cardPos] // 更新当前牌数据
	if lastCardValue == caothap.GetCardValue(currentCard) {
		this.isEven = true
	} else {
		this.isEven = false
	}
	this.cards = append(this.cards, currentCard)
	this.cardCount = int32(len(this.cards))
	return
}

func (this *CaoThapPlayerData) SwitchCardPos(lastPos, newPos int) {
	if lastPos < 0 || newPos < 0 || lastPos > caothap.CARDSTYPE_NUM-1 || newPos > caothap.CARDSTYPE_NUM-1 {
		logger.Logger.Errorf("SwitchCardPos err: input invalid parameter")
		return
	}
	this.cardsData[lastPos], this.cardsData[newPos] = this.cardsData[newPos], this.cardsData[lastPos]
}

func (this *CaoThapPlayerData) GetCardPos(locationID int64) int {
	var cardValue int
	lastCard := this.cardsData[this.cardPos-1]
	lastCardValue := caothap.GetCardValue(lastCard)
	if locationID == CaoThapLitte {
		cardValue = lastCardValue + rand.Intn(int(caothap.CARDSTYPE_NUM)-lastCardValue)
	} else if locationID == CaoThapBig {
		cardValue = lastCardValue - rand.Intn(int(caothap.CARDSTYPE_NUM)-lastCardValue)
	}
	for i := this.cardPos; i < len(this.cardsData); i++ {
		if cardValue == caothap.GetCardValue(this.cardsData[i]) {
			return i
		}
	}
	return -1
}

func (this *CaoThapPlayerData) CleanPlayerData() {
	this.score = 0
	this.RollGameType = &model.CaoThapType{}
	this.enterGameCoin = this.Coin
	this.taxCoin = 0
	this.winCoin = 0
	this.cards = make([]int32, 0)
	this.cardsData = make([]int32, caothap.CARDDATANUM)
	this.cardPos = 0
	this.lastJackpotTime = time.Now().Add(-caothap.JACKPOTTIMEINTERVAL * time.Hour)
	this.step = 0
	this.cardCount = 0
	this.currentAces = make([]int32, 0)
	this.currentAcesQuantity = 0
	this.betValue = 0
	this.prizeValue = 0
	this.createdTimePlay = time.Now()
	this.betRateUp = 0
	this.betRateDown = 0
	this.remainTime = 0
	this.isEven = false
	this.isJackpot = false
	this.jackpotValue = 0
	if this.playerOpHandler != timer.TimerHandle(0) {
		timer.StopTimer(this.playerOpHandler)
		this.playerOpHandler = timer.TimerHandle(0)
	}
}
