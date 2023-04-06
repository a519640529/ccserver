package candy

import (
	proto_candy "games.yol.com/win88/protocol/candy"
	"math/rand"
	"time"

	"games.yol.com/win88/gamerule/candy"
	"games.yol.com/win88/minigame/base"
	"games.yol.com/win88/model"
)

type CandyPlayerData struct {
	*base.Player
	score      int32                       //单线押注数
	cards      []int32                     //9张牌
	billedData *proto_candy.GameBilledData //上一局结算信息
	//RollGameType   *model.CandyType            //记录信息
	RollGameType   *GameResultLog //记录信息
	enterGameCoin  int64          //玩家进入初始金币
	betLines       []int64        //下注的选线
	taxCoin        int64          //本局税收
	winCoin        int64          //本局收税前赢的钱
	jackpotWinCoin int64          //本局奖池赢的钱
	leavetime      int32          //用户离开时间
	betIdx         int32          //下注筹码索引

	//测试
	debugJackpot bool
}

//玩家初始化
func (this *CandyPlayerData) init(s *base.Scene) {
	this.betIdx = 0
	this.score = s.DbGameFree.GetOtherIntParams()[0]
	//this.RollGameType = &model.CandyType{}
	this.RollGameType = &GameResultLog{}
	this.RollGameType.BaseResult = &model.SlotBaseResultType{}
	this.enterGameCoin = this.Coin
	for i := 0; i < candy.LINE_CELL; i++ {
		this.cards = append(this.cards, rand.Int31n(int32(candy.Element_Max-1)+1))
	}
	//线条全选
	if len(this.betLines) == 0 {
		this.betLines = candy.AllBetLines
	}
	this.LastOPTimer = time.Now()
	this.billedData = &proto_candy.GameBilledData{}
	this.debugJackpot = true
}

//黑白名单的限制是否生效
func (this *CandyPlayerData) CheckBlackWriteList(isWin bool) bool {
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
