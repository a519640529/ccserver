package baccarat

import (
	"games.yol.com/win88/gamerule/baccarat"
	"games.yol.com/win88/proto"
	proto_baccarat "games.yol.com/win88/protocol/baccarat"
	"games.yol.com/win88/robot/base"
	"time"
)

var BaccaratNilPlayer *BaccaratPlayer = nil

type BaccaratPlayer struct {
	base.BasePlayers
	*proto_baccarat.BaccaratPlayerData
	bets     [Baccarat_Zone_Max]int64
	choose   int32
	totalBet int64
	tNextBet time.Time
}

func NewBaccaratPlayer(data *proto_baccarat.BaccaratPlayerData) *BaccaratPlayer {
	p := &BaccaratPlayer{BaccaratPlayerData: data}
	p.Init()
	return p
}

func (p *BaccaratPlayer) Init() {
	p.Clear()
}

func (p *BaccaratPlayer) Clear() {
	for i := 0; i < Baccarat_Zone_Max; i++ {
		p.bets[i] = 0
	}
	p.totalBet = 0
	s := base.PlayerMgrSington.GetPlayerSession(p.GetSnId())
	if s != nil {
		base.StopSessionGameTimer(s)
	}
	p.choose = -1
}

func (p *BaccaratPlayer) MarkFlag(flag int32) {
	myFlag := p.GetFlag()
	myFlag |= flag
	p.Flag = proto.Int32(myFlag)
}

func (p *BaccaratPlayer) UnmarkFlag(flag int32) {
	myFlag := p.GetFlag()
	myFlag &= ^flag
	p.Flag = proto.Int32(myFlag)
}

func (p *BaccaratPlayer) IsMarkFlag(flag int32) bool {
	if (p.GetFlag() & flag) != 0 {
		return true
	}
	return false
}

func (p *BaccaratPlayer) IsOnLine() bool {
	return p.IsMarkFlag(base.PlayerState_Online)
}

func (p *BaccaratPlayer) IsReady() bool {
	return true
}

func (p *BaccaratPlayer) IsSceneOwner() bool {
	return false
}

func (p *BaccaratPlayer) CanOp() bool {
	return true
}

func (p *BaccaratPlayer) IsRobot() bool {
	player := base.PlayerMgrSington.GetPlayer(p.GetSnId())
	return player != nil
}

func (p *BaccaratPlayer) SetFlag(flag int32) {
	p.Flag = proto.Int32(flag)
}

func (p *BaccaratPlayer) GetLastOp() int32 {
	return 0
}

func (p *BaccaratPlayer) SetLastOp(op int32) {
}

func (p *BaccaratPlayer) UpdateCards(cards []int32) {
}
func (p *BaccaratPlayer) Banker() {
	session := base.PlayerMgrSington.GetPlayerSession(p.GetSnId())
	if session == nil {
		return
	}
	pack := &proto_baccarat.CSBaccaratOp{
		OpCode: proto.Int(baccarat.BaccaratPlayerOpUpBanker),
		Params: []int64{},
	}
	proto.SetDefaults(pack)
	session.Send(int(proto_baccarat.BaccaratPacketID_PACKET_CS_BACCARAT_PLAYEROP), pack)
}
