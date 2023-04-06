package dragonvstiger

import (
	. "games.yol.com/win88/gamerule/dragonvstiger"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/dragonvstiger"
	"games.yol.com/win88/robot/base"
	b3core "github.com/magicsea/behavior3go/core"
	"time"
)

var DragonVsTigerNilPlayer *DragonVsTigerPlayer = nil

type DragonVsTigerPlayer struct {
	base.BasePlayers
	base.HBasePlayers
	*dragonvstiger.DragonVsTigerPlayerData
	bets     [DVST_ZONE_MAX]int64
	choose   int32
	totalBet int64
	tNextBet time.Time
}

func NewDragonVsTigerPlayer(data *dragonvstiger.DragonVsTigerPlayerData) *DragonVsTigerPlayer {
	p := &DragonVsTigerPlayer{DragonVsTigerPlayerData: data}
	p.Init()
	return p
}

func (p *DragonVsTigerPlayer) Init() {
	p.BlackData = b3core.NewBlackboard()
	p.Clear()
}

func (p *DragonVsTigerPlayer) Clear() {
	for i := 0; i < DVST_ZONE_MAX; i++ {
		p.bets[i] = 0
	}
	p.totalBet = 0
	s := base.PlayerMgrSington.GetPlayerSession(p.GetSnId())
	if s != nil {
		base.StopSessionGameTimer(s)
	}
	p.choose = -1
}

func (p *DragonVsTigerPlayer) MarkFlag(flag int32) {
	myFlag := p.GetFlag()
	myFlag |= flag
	p.Flag = proto.Int32(myFlag)
}

func (p *DragonVsTigerPlayer) UnmarkFlag(flag int32) {
	myFlag := p.GetFlag()
	myFlag &= ^flag
	p.Flag = proto.Int32(myFlag)
}

func (p *DragonVsTigerPlayer) IsMarkFlag(flag int32) bool {
	if (p.GetFlag() & flag) != 0 {
		return true
	}
	return false
}

func (p *DragonVsTigerPlayer) IsOnLine() bool {
	return p.IsMarkFlag(base.PlayerState_Online)
}

func (p *DragonVsTigerPlayer) IsReady() bool {
	return true
}

func (p *DragonVsTigerPlayer) IsSceneOwner() bool {
	return false
}

func (p *DragonVsTigerPlayer) CanOp() bool {
	return true
}

func (p *DragonVsTigerPlayer) IsRobot() bool {
	player := base.PlayerMgrSington.GetPlayer(p.GetSnId())
	return player != nil
}

func (p *DragonVsTigerPlayer) SetFlag(flag int32) {
	p.Flag = proto.Int32(flag)
}

func (p *DragonVsTigerPlayer) GetLastOp() int32 {
	return 0
}

func (p *DragonVsTigerPlayer) SetLastOp(op int32) {
}

func (p *DragonVsTigerPlayer) UpdateCards(cards []int32) {
}

func (p *DragonVsTigerPlayer) BetCoin(coin int64, index int) bool {
	if p.Scene != nil {
		pack := &dragonvstiger.CSDragonVsTiggerOp{
			OpCode: proto.Int(DragonVsTigerPlayerOpBet),
			Params: []int64{int64(index), coin},
		}
		proto.SetDefaults(pack)
		return p.SendMsg(p.GetSnId(), int(dragonvstiger.DragonVsTigerPacketID_PACKET_CS_DVST_PLAYEROP), pack)
	}
	return true
}
func (p *DragonVsTigerPlayer) GetOutLimitCoin() int {
	if p.Scene != nil {
		sc, _ := p.Scene.(*DragonVsTigerScene)
		params := sc.GetOtherIntParams()
		if len(params) == 0 {
			return -1
		}

		return int(params[0])
	}
	return -1
}

func (p *DragonVsTigerPlayer) UpdateAction(ts int64) {
	if p.Scene != nil {
		sc, _ := p.Scene.(*DragonVsTigerScene)
		tree := base.GetBevTree(sc.RobotTypeAIName[p.TreeID])
		if tree != nil && p.IsRobot() {
			tree.Tick(p, p.BlackData)
		}
	}
}
func (p *DragonVsTigerPlayer) Banker() {
	session := base.PlayerMgrSington.GetPlayerSession(p.GetSnId())
	if session == nil {
		return
	}
	pack := &dragonvstiger.CSDragonVsTiggerOp{
		OpCode: proto.Int(DragonVsTigerPlayerOpUpBanker),
		Params: []int64{},
	}
	proto.SetDefaults(pack)
	session.Send(int(dragonvstiger.DragonVsTigerPacketID_PACKET_CS_DVST_PLAYEROP), pack)
}
