package redvsblack

import (
	rule "games.yol.com/win88/gamerule/redvsblack"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/redvsblack"
	"games.yol.com/win88/robot/base"
	b3core "github.com/magicsea/behavior3go/core"
	"time"
)

var RedVsBlackNilPlayer *RedVsBlackPlayer = nil

type RedVsBlackPlayer struct {
	base.BasePlayers
	base.HBasePlayers
	*redvsblack.RedVsBlackPlayerData
	bets     [RVSB_ZONE_MAX]int64
	choose   int32
	totalBet int64
	tNextBet time.Time
}

func NewRedVsBlackPlayer(data *redvsblack.RedVsBlackPlayerData) *RedVsBlackPlayer {
	p := &RedVsBlackPlayer{RedVsBlackPlayerData: data}
	p.Init()
	return p
}

func (p *RedVsBlackPlayer) Init() {
	p.BlackData = b3core.NewBlackboard()
	p.Clear()
}

func (p *RedVsBlackPlayer) Clear() {
	for i := 0; i < RVSB_ZONE_MAX; i++ {
		p.bets[i] = 0
	}
	p.totalBet = 0
	s := base.PlayerMgrSington.GetPlayerSession(p.GetSnId())
	if s != nil {
		base.StopSessionGameTimer(s)
	}
	p.choose = -1
}

func (p *RedVsBlackPlayer) MarkFlag(flag int32) {
	myFlag := p.GetFlag()
	myFlag |= flag
	p.Flag = proto.Int32(myFlag)
}

func (p *RedVsBlackPlayer) UnmarkFlag(flag int32) {
	myFlag := p.GetFlag()
	myFlag &= ^flag
	p.Flag = proto.Int32(myFlag)
}

func (p *RedVsBlackPlayer) IsMarkFlag(flag int32) bool {
	if (p.GetFlag() & flag) != 0 {
		return true
	}
	return false
}

func (p *RedVsBlackPlayer) IsOnLine() bool {
	return p.IsMarkFlag(base.PlayerState_Online)
}

func (p *RedVsBlackPlayer) IsReady() bool {
	return true
}

func (p *RedVsBlackPlayer) IsSceneOwner() bool {
	return false
}

func (p *RedVsBlackPlayer) CanOp() bool {
	return true
}

func (p *RedVsBlackPlayer) IsRobot() bool {
	if base.PlayerMgrSington.GetPlayerSession(p.GetSnId()) != nil {
		return true
	}
	return false
}

func (p *RedVsBlackPlayer) SetFlag(flag int32) {
	p.Flag = proto.Int32(flag)
}

func (p *RedVsBlackPlayer) GetLastOp() int32 {
	return 0
}

func (p *RedVsBlackPlayer) SetLastOp(op int32) {
}

func (p *RedVsBlackPlayer) UpdateCards(cards []int32) {
}

func (p *RedVsBlackPlayer) BetCoin(coin int64, index int) bool {
	if p.Scene != nil {

		pack := &redvsblack.CSRedVsBlackOp{
			OpCode: proto.Int(rule.RedVsBlackPlayerOpBet),
			Params: []int64{int64(index), coin},
		}
		proto.SetDefaults(pack)
		return p.SendMsg(p.GetSnId(), int(redvsblack.RedVsBlackPacketID_PACKET_CS_RVSB_PLAYEROP), pack)
	}
	return true
}

func (p *RedVsBlackPlayer) GetOutLimitCoin() int {
	if p.Scene != nil {
		sc, _ := p.Scene.(*RedVsBlackScene)
		params := sc.GetOtherIntParams()
		if len(params) == 0 {
			return -1
		}

		return int(params[0])
	}
	return -1
}

func (p *RedVsBlackPlayer) UpdateAction(ts int64) {
	if p.Scene != nil {
		sc, _ := p.Scene.(*RedVsBlackScene)
		tree := base.GetBevTree(sc.RobotTypeAIName[p.TreeID])
		if tree != nil && p.IsRobot() {
			tree.Tick(p, p.BlackData)
		}
	}
}
