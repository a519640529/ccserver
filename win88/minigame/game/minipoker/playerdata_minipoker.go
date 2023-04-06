package minipoker

import (
	proto_minipoker "games.yol.com/win88/protocol/minipoker"
	"math/rand"
	"time"

	"games.yol.com/win88/gamerule/minipoker"
	"games.yol.com/win88/minigame/base"
	"games.yol.com/win88/model"
)

type MiniPokerPlayerData struct {
	*base.Player
	score int32   //押注数
	cards []int32 //5张牌
	//RollGameType  *model.MiniPokerType //记录信息
	RollGameType  *GameResultLog //记录信息
	enterGameCoin int64          //玩家进入初始金币
	taxCoin       int64          //本局税收
	winCoin       int64          //本局收税前赢的钱
	//linesWinCoin    int64                //本局中奖牌型赢得钱
	jackpotWinCoin  int64                           //本局奖池赢的钱
	leavetime       int32                           //用户离开时间
	cardsData       []int32                         // 牌库
	cardPos         int                             // 当前牌索引
	lastJackpotTime time.Time                       // 最后一次爆奖时间
	billedData      *proto_minipoker.GameBilledData //上一局结算信息
	betIdx          int32                           //下注筹码索引
	//测试
	debugJackpot bool
	DebugGame    bool //测试
	TestNum      int
}

//玩家初始化
func (this *MiniPokerPlayerData) init(s *base.Scene) {
	otherIntParams := s.DbGameFree.GetOtherIntParams()
	this.score = otherIntParams[0] // 底注
	this.betIdx = 0
	//this.RollGameType = &model.MiniPokerType{}
	this.RollGameType = &GameResultLog{}
	this.RollGameType.BaseResult = &model.SlotBaseResultType{}
	this.enterGameCoin = this.Coin
	this.cards = make([]int32, minipoker.CARDNUM)
	this.cardsData = make([]int32, minipoker.CARDDATANUM)
	this.cardPos = 0
	this.lastJackpotTime = time.Now().Add(-minipoker.JACKPOTTIMEINTERVAL * time.Hour)
	this.winCoin = 0
	//this.linesWinCoin = 0
	this.jackpotWinCoin = 0
	this.LastOPTimer = time.Now()
	this.billedData = &proto_minipoker.GameBilledData{}
	this.debugJackpot = true
	this.DebugGame = true
	this.TestNum = 0
}

//黑白名单的限制是否生效
func (this *MiniPokerPlayerData) CheckBlackWriteList(isWin bool) bool {
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
