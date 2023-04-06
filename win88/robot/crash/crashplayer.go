package crash

import (
	rule "games.yol.com/win88/gamerule/crash"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/crash"
	"games.yol.com/win88/robot/base"
	b3core "github.com/magicsea/behavior3go/core"
	"time"
)

var CrashNilPlayer *CrashPlayer = nil

type CrashPlayer struct {
	base.BasePlayers
	base.HBasePlayers
	*crash.CrashPlayerData
	//bets     [CRASH_ZONE_MAX]int64
	betTotal      int64           //当局总下注筹码
	multiple      int32           //下注倍率
	choose   int32
	totalBet int64
	tNextBet time.Time
	down     bool //已下注
}

func NewCrashPlayer(data *crash.CrashPlayerData) *CrashPlayer {
	p := &CrashPlayer{CrashPlayerData: data}
	p.Init()
	return p
}

func (p *CrashPlayer) Init() {
	p.BlackData = b3core.NewBlackboard()
	p.Clear()
}

func (p *CrashPlayer) Clear() {
	//for i := 0; i < CRASH_ZONE_MAX; i++ {
	//	p.bets[i] = 0
	//}
	p.multiple = 0
	p.betTotal = 0
	p.down = false
	p.totalBet = 0
	s := base.PlayerMgrSington.GetPlayerSession(p.GetSnId())
	if s != nil {
		base.StopSessionGameTimer(s)
	}
	p.choose = 100
}

func (p *CrashPlayer) MarkFlag(flag int32) {
	myFlag := p.GetFlag()
	myFlag |= flag
	p.Flag = proto.Int32(myFlag)
}

func (p *CrashPlayer) UnmarkFlag(flag int32) {
	myFlag := p.GetFlag()
	myFlag &= ^flag
	p.Flag = proto.Int32(myFlag)
}

func (p *CrashPlayer) IsMarkFlag(flag int32) bool {
	if (p.GetFlag() & flag) != 0 {
		return true
	}
	return false
}

func (p *CrashPlayer) IsOnLine() bool {
	return p.IsMarkFlag(base.PlayerState_Online)
}

func (p *CrashPlayer) IsReady() bool {
	return true
}

func (p *CrashPlayer) IsSceneOwner() bool {
	return false
}

func (p *CrashPlayer) CanOp() bool {
	return true
}

func (p *CrashPlayer) IsRobot() bool {
	if base.PlayerMgrSington.GetPlayerSession(p.GetSnId()) != nil {
		return true
	}
	return false
}

func (p *CrashPlayer) SetFlag(flag int32) {
	p.Flag = proto.Int32(flag)
}

func (p *CrashPlayer) GetLastOp() int32 {
	return 0
}

func (p *CrashPlayer) SetLastOp(op int32) {
}

func (p *CrashPlayer) UpdateCards(cards []int32) {
}

func (p *CrashPlayer) BetCoin(coin int64, index int) bool {
	if p.Scene != nil {

		pack := &crash.CSCrashOp{
			OpCode: proto.Int(rule.CrashPlayerOpBet),
			Params: []int64{int64(index), coin},
		}
		proto.SetDefaults(pack)
		return p.SendMsg(p.GetSnId(), int(crash.CrashPacketID_PACKET_CS_CRASH_PLAYEROP), pack)
	}
	return true
}

func (p *CrashPlayer) GetOutLimitCoin() int {
	if p.Scene != nil {
		sc, _ := p.Scene.(*CrashScene)
		params := sc.GetOtherIntParams()
		if len(params) == 0 {
			return -1
		}

		return int(params[0])
	}
	return -1
}

func (p *CrashPlayer) UpdateAction(ts int64) {
	if p.Scene != nil {
		sc, _ := p.Scene.(*CrashScene)
		tree := base.GetBevTree(sc.RobotTypeAIName[p.TreeID])
		if tree != nil && p.IsRobot() {
			tree.Tick(p, p.BlackData)
		}
	}
}
