package rollpoint

import (
	rule "games.yol.com/win88/gamerule/rollpoint"
	"games.yol.com/win88/protocol/rollpoint"
	"games.yol.com/win88/robot/base"
	"math/rand"
	"time"
)

type RollPointPlayerData struct {
	base.BasePlayers
	*rollpoint.RollPointPlayer
	PushCoin int32 //押注的金币
	tNextBet time.Time
	tEndBet  time.Time
	selIndex []int32
	betMax   int32
}

func (this *RollPointPlayerData) Clear() {
	this.PushCoin = 0
	this.selIndex = nil
	if rand.Intn(100) < 50 {
		this.selIndex = append(this.selIndex, rule.BetArea_Small)
	} else {
		this.selIndex = append(this.selIndex, rule.BetArea_Big)
	}
	for i := 0; i < 5; i++ {
		arr := RollPointRateArea[rand.Intn(len(RollPointRateArea))]
		this.selIndex = append(this.selIndex, arr[rand.Intn(len(arr))])
	}
	s := base.PlayerMgrSington.GetPlayerSession(this.GetSnId())
	if s != nil {
		base.StopSessionGameTimer(s)
	}
}
func (this *RollPointPlayerData) MaxChipCheck(index int64, coin int64, maxBetCoin []int32) (bool, int32) {
	if index < 0 || int32(index) > rule.BetArea_Max {
		return false, -1
	}
	maxBetCoinIndex := rule.AreaIndex2MaxChipIndex[index]
	if maxBetCoin[maxBetCoinIndex] > 0 && this.PushCoin+int32(coin) > maxBetCoin[maxBetCoinIndex] {
		return false, maxBetCoin[maxBetCoinIndex]
	}
	return true, 0
}
func (this *RollPointPlayerData) GetSnId() int32             { return this.RollPointPlayer.GetSnId() }
func (this *RollPointPlayerData) GetPos() int32              { return 0 }
func (this *RollPointPlayerData) GetFlag() int32             { return this.RollPointPlayer.GetFlag() }
func (this *RollPointPlayerData) SetFlag(flag int32)         {}
func (this *RollPointPlayerData) GetLastOp() int32           { return 0 }
func (this *RollPointPlayerData) SetLastOp(op int32)         {}
func (this *RollPointPlayerData) MarkFlag(flag int32)        {}
func (this *RollPointPlayerData) UnmarkFlag(flag int32)      {}
func (this *RollPointPlayerData) IsMarkFlag(flag int32) bool { return true }
func (this *RollPointPlayerData) IsOnLine() bool             { return true }
func (this *RollPointPlayerData) IsReady() bool              { return true }
func (this *RollPointPlayerData) IsSceneOwner() bool         { return true }
func (this *RollPointPlayerData) IsRobot() bool              { return true }
func (this *RollPointPlayerData) UpdateCards(cards []int32)  {}
