package blackjack

import (
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/blackjack"
	"games.yol.com/win88/robot/base"
)

type BlackJackPlayer struct {
	base.BasePlayers
	*blackjack.BlackJackPlayer
	character int   // 性格
	betCoin   int64 // 下注金额
	baoCoin   int64 // 保险金额
}

func NewBlackJackPlayer(info *blackjack.BlackJackPlayer) *BlackJackPlayer {
	p := &BlackJackPlayer{
		BlackJackPlayer: info,
	}
	return p
}

func (this *BlackJackPlayer) Release() {
	this.character = 0
	this.betCoin = 0
	this.baoCoin = 0
}

func (this *BlackJackPlayer) Clear() {

}

func (this *BlackJackPlayer) SetFlag(flag int32) {
	this.Flag = proto.Int32(flag)
}

func (this *BlackJackPlayer) GetLastOp() int32 {
	return 0
}

func (this *BlackJackPlayer) SetLastOp(op int32) {

}

func (this *BlackJackPlayer) MarkFlag(flag int32) {
	myFlag := this.GetFlag()
	myFlag |= flag
	this.Flag = proto.Int32(myFlag)
}

func (this *BlackJackPlayer) UnmarkFlag(flag int32) {
	myFlag := this.GetFlag()
	myFlag &= ^flag
	this.Flag = proto.Int32(myFlag)
}

func (this *BlackJackPlayer) IsMarkFlag(flag int32) bool {
	if (this.GetFlag() & flag) != 0 {
		return true
	}
	return false
}

func (this *BlackJackPlayer) IsOnLine() bool {
	return this.IsMarkFlag(base.PlayerState_Online)
}

func (this *BlackJackPlayer) IsReady() bool {
	return this.IsMarkFlag(base.PlayerState_Ready)
}

func (this *BlackJackPlayer) IsSceneOwner() bool {
	return this.IsMarkFlag(base.PlayerState_SceneOwner)
}

func (this *BlackJackPlayer) IsRobot() bool {
	player := base.PlayerMgrSington.GetPlayer(this.GetSnId())
	return player != nil
}

func (this *BlackJackPlayer) UpdateCards(cards []int32) {

}
