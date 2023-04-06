package baccarat

import (
	"games.yol.com/win88/common"
	"games.yol.com/win88/gamerule/baccarat"
	"games.yol.com/win88/gamesrv/base"
	"math/rand"
	"time"
)

var BaccaratChipWeight = []int64{60, 20, 10, 8, 2}
var BaccaratAreaWeight = []int64{8000 / 2, 8000 / 2, 600, 600, 800}

type BaccaratPlayerAI struct {
	base.BaseAI
	choose   int32
	totalBet int64
	tNextBet time.Time
}

//房间状态变化事件
func (this *BaccaratPlayerAI) OnChangeSceneState(s *base.Scene, oldstate, newstate int) {
	this.BaseAI.OnChangeSceneState(s, oldstate, newstate)
	switch newstate {
	case baccarat.BaccaratSceneStateStake:
		this.Choose(s)
	case baccarat.BaccaratSceneStateBilled:
		this.Clear(s)
	}
}

//其他玩家操作事件
func (this *BaccaratPlayerAI) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	this.BaseAI.OnPlayerOp(s, p, opcode, params)

	return true
}

//心跳事件
func (this *BaccaratPlayerAI) OnTick(s *base.Scene) {
	this.BaseAI.OnTick(s)
	if this.choose != -1 {
		if !time.Now().Before(this.tNextBet) {
			this.Action(s)
		}
	}
}

func (this *BaccaratPlayerAI) Choose(s *base.Scene) {
	this.choose = rand.Int31n(2) + 1
	this.tNextBet = time.Now().Add(time.Duration(rand.Int31n(1000)+500) * time.Millisecond)
}

func (this *BaccaratPlayerAI) Action(s *base.Scene) {
	var idx int32 = 1
	idx = idx << (1 + uint32(rand.Intn(2)))
	rn := common.RandSliceIndexByWight(BaccaratAreaWeight)
	switch rn {
	case 0, 1:
		idx = this.choose
	case 2:
		idx = int32(baccarat.BACCARAT_ZONE_BANKER_DOUBLE)
	case 3:
		idx = int32(baccarat.BACCARAT_ZONE_PLAYER_DOUBLE)
	case 4:
		idx = int32(baccarat.BACCARAT_ZONE_TIE)
	}
	chip := int32(0)
	otherParams := s.DbGameFree.GetOtherIntParams()
	o := common.RandSliceIndexByWight(BaccaratChipWeight)
	chip = otherParams[o]
	p := this.GetOwner()
	//金币不够
	if p.GetCoin() < int64(chip) {
		n := len(otherParams)
		coin := otherParams[n-1]
		p.AddCoin(int64(coin), common.GainWay_ByPMCmd, base.SyncFlag_Broadcast, "", "")
		this.tNextBet = time.Now().Add(time.Second * 10)
		return
	}

	s.GetScenePolicy().OnPlayerOp(s, p, baccarat.BaccaratPlayerOpBet, []int64{int64(idx), int64(chip)})
}

func (this *BaccaratPlayerAI) Clear(s *base.Scene) {
	this.totalBet = 0
	this.choose = -1
}
