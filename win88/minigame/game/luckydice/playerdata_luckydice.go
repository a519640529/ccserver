package luckydice

import (
	"games.yol.com/win88/minigame/base"
	"time"
)

type LuckyDicePlayerData struct {
	*base.Player
	leavetime      int32 //用户离开时间
	bet            int64 //本局总押注
	betSide        int32 //押大押小
	award          int64 //本局获胜金额 税后
	tax            int64 //赢家税收
	refund         int64 //返还押注数0
	betTime        int64 //最后押注时间
	doOncePerRound bool  //每局执行一次
	isLeave        bool  //离开
}

//玩家初始化
func (this *LuckyDicePlayerData) init(s *base.Scene) {
	this.Clean()
}

//玩家清理数据
func (this *LuckyDicePlayerData) Clean() {
	this.bet = 0
	this.betSide = -1
	this.award = 0
	this.tax = 0
	this.refund = 0
	this.betTime = 0
	this.doOncePerRound = true
	this.LastOPTimer = time.Now()
}
