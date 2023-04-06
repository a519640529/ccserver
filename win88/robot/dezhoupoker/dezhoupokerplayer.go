package dezhoupoker

import (
	rule "games.yol.com/win88/gamerule/dezhoupoker"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/dezhoupoker"
	"games.yol.com/win88/robot/base"
)

var DezhouPokerNilPlayer *DezhouPokerPlayer = nil

const (
	AI_DEZHOU_FLAG_FRIGHTEN int32 = 1 << iota //吓唬
	AI_DEZHOU_FLAG_INCPOT                     //拉升底池，打价值
)

type DezhouPokerPlayer struct {
	base.BasePlayers
	*dezhoupoker.DezhouPokerPlayerData
	OpDelayTimes int32 //本局延迟操作次数
	aiFlag       int32 //ai辅助标识
}

func NewDezhouPokerPlayer(data *dezhoupoker.DezhouPokerPlayerData) *DezhouPokerPlayer {
	p := &DezhouPokerPlayer{DezhouPokerPlayerData: data}
	p.Init()
	return p
}

func (this *DezhouPokerPlayer) Init() {
}

func (this *DezhouPokerPlayer) Clear() {
	this.UnmarkFlag(base.PlayerState_Check)
	this.UnmarkFlag(base.PlayerState_Fold)
	this.UnmarkFlag(base.PlayerState_Lose)
	this.UnmarkFlag(base.PlayerState_Win)
	this.UnmarkFlag(base.PlayerState_WaitNext)
	this.UnmarkFlag(base.PlayerState_GameBreak)
	this.SetLastOp(rule.DezhouPokerPlayerOpNull)
	this.OpDelayTimes = 0
	this.aiFlag = 0
}

func (this *DezhouPokerPlayer) MarkAIFlag(flag int32) {
	this.aiFlag |= flag
}

func (this *DezhouPokerPlayer) UnmarkAIFlag(flag int32) {
	this.aiFlag &= ^flag
}

func (this *DezhouPokerPlayer) IsMarkAIFlag(flag int32) bool {
	return this.aiFlag&flag != 0
}

func (this *DezhouPokerPlayer) MarkFlag(flag int32) {
	myFlag := this.GetFlag()
	myFlag |= flag
	this.Flag = proto.Int32(myFlag)
}

func (this *DezhouPokerPlayer) UnmarkFlag(flag int32) {
	myFlag := this.GetFlag()
	myFlag &= ^flag
	this.Flag = proto.Int32(myFlag)
}

func (this *DezhouPokerPlayer) IsMarkFlag(flag int32) bool {
	if (this.GetFlag() & flag) != 0 {
		return true
	}
	return false
}

func (this *DezhouPokerPlayer) IsOnLine() bool {
	return this.IsMarkFlag(base.PlayerState_Online)
}

func (this *DezhouPokerPlayer) IsReady() bool {
	return this.IsMarkFlag(base.PlayerState_Ready)
}

func (this *DezhouPokerPlayer) IsSceneOwner() bool {
	return this.IsMarkFlag(base.PlayerState_SceneOwner)
}

func (this *DezhouPokerPlayer) IsCheck() bool {
	return this.IsMarkFlag(base.PlayerState_Check)
}

func (this *DezhouPokerPlayer) CanOp() bool {
	if this.IsMarkFlag(base.PlayerState_Fold) || this.IsMarkFlag(base.PlayerState_Lose) || this.IsMarkFlag(base.PlayerState_WaitNext) || this.IsMarkFlag(base.PlayerState_GameBreak) {
		return false
	}
	if this.GetLastOp() == rule.DezhouPokerPlayerOpFold {
		return false
	}
	return true
}
func (this *DezhouPokerPlayer) IsGameing() bool {
	return !this.IsMarkFlag(base.PlayerState_WaitNext) && !this.IsMarkFlag(base.PlayerState_GameBreak)
}

func (this *DezhouPokerPlayer) IsRobot() bool {
	player := base.PlayerMgrSington.GetPlayer(this.GetSnId())
	return player != nil
}

func (this *DezhouPokerPlayer) SetFlag(flag int32) {
	this.Flag = proto.Int32(flag)
}

func (this *DezhouPokerPlayer) SetLastOp(op int32) {
	this.LastOp = proto.Int32(op)
}

func (this *DezhouPokerPlayer) SetGameCoin(gamecoin int64) {
	this.GameCoin = proto.Int64(gamecoin)
}

func (this *DezhouPokerPlayer) SetCards(cards []int32) {
	this.Cards = cards
}

func (this *DezhouPokerPlayer) UpdateCards(cards []int32) {
	if this != nil {
		this.Cards = cards
	}
}
