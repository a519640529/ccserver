package rollcoin

import (
	"games.yol.com/win88/common"
	"games.yol.com/win88/gamerule/rollcoin"
	"games.yol.com/win88/gamesrv/base"
	"math/rand"
	"time"
)

type RollCoinPlayerAI struct {
	base.BaseAI
	choose   int32
	totalBet int32
	tNextBet time.Time
	selIndex []int
}

//房间状态变化事件
func (this *RollCoinPlayerAI) OnChangeSceneState(s *base.Scene, oldstate, newstate int) {
	this.BaseAI.OnChangeSceneState(s, oldstate, newstate)
	switch newstate {
	case rollcoin.RollCoinSceneStateStart:
		this.Choose(s)
	case rollcoin.RollCoinSceneStateBilled:
		this.Clear(s)
	}
}

//其他玩家操作事件
func (this *RollCoinPlayerAI) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	this.BaseAI.OnPlayerOp(s, p, opcode, params)
	return true
}

//心跳事件
func (this *RollCoinPlayerAI) OnTick(s *base.Scene) {
	this.BaseAI.OnTick(s)
	if this.choose != -1 {
		if !time.Now().Before(this.tNextBet) {
			this.Action(s)
		}
	}
}

func (this *RollCoinPlayerAI) Choose(s *base.Scene) {
	this.choose = rand.Int31n(2)
	this.tNextBet = time.Now().Add(time.Duration(rand.Int31n(1000)+500) * time.Millisecond)
}

func (this *RollCoinPlayerAI) Action(s *base.Scene) {
	var coin int32
	//params := s.dbGameFree.GetOtherIntParams()
	var params = []int32{100, 1000, 5000, 10000, 50000}
	if sceneEx, ok := s.ExtraData.(*RollCoinSceneData); ok {
		params = sceneEx.DbGameFree.GetOtherIntParams()
	}
	coin = params[int(common.RandItemByWight(RollCoinSelRate))]
	me := this.GetOwner()
	//金币不够
	if me.Coin <= int64(coin) {
		return
	}
	if len(this.selIndex) < 3 {
		this.selIndex = append(this.selIndex, int(common.RandItemByWight(RollCoinSelFlag)))
	}
	idx := this.selIndex[rand.Intn(len(this.selIndex))]
	this.totalBet += coin
	me.Coin -= int64(coin)
	s.GetScenePolicy().OnPlayerOp(s, me, rollcoin.RollCoinPlayerOpPushCoin, []int64{int64(idx), int64(coin)})
	this.tNextBet = time.Now().Add(time.Duration(rand.Int31n(3000)+1000) * time.Millisecond)
}

func (this *RollCoinPlayerAI) Clear(s *base.Scene) {
	this.choose = -1
	this.totalBet = 0
}
