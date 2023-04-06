package tienlen

import (
	rule "games.yol.com/win88/gamerule/tienlen"
	"games.yol.com/win88/gamesrv/base"
	"math/rand"
)

// 玩家身上的额外数据
type TienLenPlayerData struct {
	*base.Player                           //玩家信息
	cards          [rule.HandCardNum]int32 //手牌信息
	delCards       [][]int32               //出过的手牌
	bombScore      int64                   //炸弹得分
	bombTaxScore   int64                   //炸弹税收
	isPass         bool                    //小轮标记过牌
	isDelAll       bool                    //出完牌了
	robotGameTimes int32                   //机器人局数限制
	winCoin        int64                   //本局赢的钱
}

func (this *TienLenPlayerData) init() {
	for i := int32(0); i < rule.HandCardNum; i++ {
		this.cards[i] = rule.InvalideCard
	}
	this.delCards = [][]int32{}
	this.isPass = false
	this.isDelAll = false
	this.robotGameTimes = rule.RobotGameTimesMin + rand.Int31n(rule.RobotGameTimesMax)
	this.winCoin = 0
	this.bombScore = 0
	this.bombTaxScore = 0
}

func (this *TienLenPlayerData) Clear() {
	for i := int32(0); i < rule.HandCardNum; i++ {
		this.cards[i] = rule.InvalideCard
	}
	this.delCards = [][]int32{}
	this.isPass = false
	this.isDelAll = false
	this.winCoin = 0
	this.bombScore = 0
	this.bombTaxScore = 0
}

// 能否操作
func (this *TienLenPlayerData) CanOp() bool {
	if this.IsGameing() && !this.isPass && !this.isDelAll {
		return true
	}
	return false
}
