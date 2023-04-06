package tienlen

import (
	tienlenApi "games.yol.com/win88/api3th/smart/tienlen"
	"games.yol.com/win88/proto"
	proto_tienlen "games.yol.com/win88/protocol/tienlen"
	"games.yol.com/win88/robot/base"
)

var TienLenNilPlayer *TienLenPlayer = nil

type TienLenPlayer struct {
	base.BasePlayers
	*proto_tienlen.TienLenPlayerData
	data *tienlenApi.PredictRequest
}

func NewTienLenPlayer(data *proto_tienlen.TienLenPlayerData) *TienLenPlayer {
	p := &TienLenPlayer{TienLenPlayerData: data}
	p.Init()
	return p
}

func (p *TienLenPlayer) Init() {
	p.Clear()
}

func (p *TienLenPlayer) Clear() {
	s := base.PlayerMgrSington.GetPlayerSession(p.GetSnId())
	if s != nil {
		base.StopSessionGameTimer(s)
	}
}

func (p *TienLenPlayer) MarkFlag(flag int32) {
	myFlag := p.GetFlag()
	myFlag |= flag
	p.Flag = proto.Int32(myFlag)
}

func (p *TienLenPlayer) UnmarkFlag(flag int32) {
	myFlag := p.GetFlag()
	myFlag &= ^flag
	p.Flag = proto.Int32(myFlag)
}

func (p *TienLenPlayer) IsMarkFlag(flag int32) bool {
	if (p.GetFlag() & flag) != 0 {
		return true
	}
	return false
}

func (p *TienLenPlayer) IsOnLine() bool {
	return p.IsMarkFlag(base.PlayerState_Online)
}

func (p *TienLenPlayer) IsReady() bool {
	return true
}

func (p *TienLenPlayer) IsSceneOwner() bool {
	return false
}

func (p *TienLenPlayer) CanOp() bool {
	return true
}

func (p *TienLenPlayer) IsRobot() bool {
	player := base.PlayerMgrSington.GetPlayer(p.GetSnId())
	return player != nil
}

func (p *TienLenPlayer) SetFlag(flag int32) {
	p.Flag = proto.Int32(flag)
}

func (p *TienLenPlayer) GetLastOp() int32 {
	return 0
}

func (p *TienLenPlayer) SetLastOp(op int32) {
}

func (p *TienLenPlayer) UpdateCards(cards []int32) {
}
