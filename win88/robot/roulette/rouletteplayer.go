package roulette

import (
	rule "games.yol.com/win88/gamerule/roulette"
	"games.yol.com/win88/proto"
	proto_roulette "games.yol.com/win88/protocol/roulette"
	"games.yol.com/win88/robot/base"
	"time"
)

var RouletteNilPlayer *RoulettePlayer = nil

type RoulettePlayer struct {
	base.BasePlayers
	*proto_roulette.RoulettePlayerData
	choose       int32
	tNextBet     time.Time
	betTime      time.Duration
	singleDouble int   //单双
	redBlack     int   //红黑
	lowHi        int   //高低
	lastBetCoin  int64 //上一局下注额
}

func NewRoulettePlayer(data *proto_roulette.RoulettePlayerData) *RoulettePlayer {
	p := &RoulettePlayer{RoulettePlayerData: data}
	p.Init()
	return p
}

func (p *RoulettePlayer) Init() {
	p.Clear()
}

func (p *RoulettePlayer) Clear() {
	p.betTime = 0
	p.singleDouble = 0
	p.redBlack = 0
	p.lowHi = 0
	p.lastBetCoin = p.GetBetCoin()
	p.BetCoin = proto.Int64(0)
	p.Pos = proto.Int32(rule.Roulette_OLPOS)
	s := base.PlayerMgrSington.GetPlayerSession(p.GetSnId())
	if s != nil {
		base.StopSessionGameTimer(s)
	}
	p.choose = -1
}

func (p *RoulettePlayer) MarkFlag(flag int32) {
	myFlag := p.GetFlag()
	myFlag |= flag
	p.Flag = proto.Int32(myFlag)
}

func (p *RoulettePlayer) UnmarkFlag(flag int32) {
	myFlag := p.GetFlag()
	myFlag &= ^flag
	p.Flag = proto.Int32(myFlag)
}

func (p *RoulettePlayer) IsMarkFlag(flag int32) bool {
	if (p.GetFlag() & flag) != 0 {
		return true
	}
	return false
}

func (p *RoulettePlayer) IsOnLine() bool {
	return p.IsMarkFlag(base.PlayerState_Online)
}

func (p *RoulettePlayer) IsReady() bool {
	return true
}

func (p *RoulettePlayer) IsSceneOwner() bool {
	return false
}

func (p *RoulettePlayer) CanOp() bool {
	return true
}

func (p *RoulettePlayer) IsRobot() bool {
	return true
}

func (p *RoulettePlayer) SetFlag(flag int32) {
	p.Flag = proto.Int32(flag)
}

func (p *RoulettePlayer) GetLastOp() int32 {
	return 0
}

func (p *RoulettePlayer) SetLastOp(op int32) {
}

func (p *RoulettePlayer) UpdateCards(cards []int32) {
}
