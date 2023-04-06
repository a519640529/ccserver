package hundredyxx

import (
	"games.yol.com/win88/common"
	rule "games.yol.com/win88/gamerule/hundredyxx"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/hundredyxx"
	"games.yol.com/win88/robot/base"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/core/timer"
	"math/rand"
	"time"
)

var HundredYXXChipWeight = []int64{50, 30, 10, 8, 2}
var HundredYXXBigZoneWeight = []int64{7, 3} //大类概率，单图案、双图案
var HundredYXXSmallZoneWeight = [][]int64{{1, 1, 1, 1, 1, 1}, {1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}}

type HundredYXXPlayer struct {
	base.BasePlayers
	*hundredyxx.HundredYXXPlayerData
	oldCoin  int64 //当前金币
	tNextBet time.Time
	betPos   int32 //本局选的下注区域
	betChip  int32 //选择的筹码
	betTotal int32
	betLimit int32
}

func NewHundredYXXPlayer(data *hundredyxx.HundredYXXPlayerData) *HundredYXXPlayer {
	p := &HundredYXXPlayer{HundredYXXPlayerData: data}
	p.Init()
	return p
}

func (p *HundredYXXPlayer) Init() {
	p.oldCoin = p.GetCoin()
}

func (p *HundredYXXPlayer) Clear() {
	p.UnmarkFlag(base.PlayerState_Ready)
	s := base.PlayerMgrSington.GetPlayerSession(p.GetSnId())
	if s != nil {
		base.StopSessionGameTimer(s)
	}
	p.betChip = 0
	p.betPos = -1
	p.betLimit = 0
	p.betTotal = 0
}

func (p *HundredYXXPlayer) MarkFlag(flag int32) {
	myFlag := p.GetFlag()
	myFlag |= flag
	p.Flag = proto.Int32(myFlag)
}

func (p *HundredYXXPlayer) UnmarkFlag(flag int32) {
	myFlag := p.GetFlag()
	myFlag &= ^flag
	p.Flag = proto.Int32(myFlag)
}

func (p *HundredYXXPlayer) IsMarkFlag(flag int32) bool {
	if (p.GetFlag() & flag) != 0 {
		return true
	}
	return false
}

func (p *HundredYXXPlayer) IsOnLine() bool {
	return p.IsMarkFlag(base.PlayerState_Online)
}

func (p *HundredYXXPlayer) IsReady() bool {
	return p.IsMarkFlag(base.PlayerState_Ready)
}

func (p *HundredYXXPlayer) IsSceneOwner() bool {
	return p.IsMarkFlag(base.PlayerState_SceneOwner)
}

func (p *HundredYXXPlayer) IsRobot() bool {
	player := base.PlayerMgrSington.GetPlayer(p.GetSnId())
	return player != nil
}

func (p *HundredYXXPlayer) SetFlag(flag int32) {
	p.Flag = proto.Int32(flag)
}

func (p *HundredYXXPlayer) GetLastOp() int32 {
	return 0
}

func (p *HundredYXXPlayer) SetLastOp(op int32) {
}

func (p *HundredYXXPlayer) UpdateCards(cards []int32) {
}

func (p *HundredYXXPlayer) ChoseBetPos() {
	//先随机下注大区域（单图案、双图案）
	bigZoneIdx := common.RandSliceIndexByWight(HundredYXXBigZoneWeight)
	p.betPos = int32(bigZoneIdx*6) + int32(common.RandSliceIndexByWight(HundredYXXSmallZoneWeight[bigZoneIdx]))
	return
}

func (p *HundredYXXPlayer) ChoseBetChip(scene *HundredYXXScene) {
	p.betLimit = int32(p.GetCoin())
	params := scene.GetParamsEx()
	if len(params) != 0 {
		dbGameFree := base.SceneMgrSington.GetSceneDBGameFree(scene.GetRoomId(), params[0])
		if dbGameFree != nil {
			otherParams := dbGameFree.GetOtherIntParams()
			chipIdx := common.RandSliceIndexByWight(HundredYXXChipWeight)
			//选一个最适合自己的筹码
			p.betChip = otherParams[chipIdx]
			for ; p.betChip > p.betLimit && chipIdx > 0; chipIdx-- {
				p.betChip = otherParams[chipIdx]
			}
			maxBetCoin := dbGameFree.GetMaxBetCoin()
			if len(maxBetCoin) != 0 {
				if p.betLimit > maxBetCoin[0] {
					p.betLimit = maxBetCoin[0]/2 + rand.Int31n(maxBetCoin[0])
				}
			}
		}
	}
}

func (p *HundredYXXPlayer) TryChoseSmallerChip(scene *HundredYXXScene) bool {
	params := scene.GetParamsEx()
	if len(params) != 0 {
		dbGameFree := base.SceneMgrSington.GetSceneDBGameFree(scene.GetRoomId(), params[0])
		if dbGameFree != nil {
			otherParams := dbGameFree.GetOtherIntParams()
			for i := len(otherParams) - 1; i >= 0; i-- {
				if otherParams[i] < p.betChip && p.betTotal+otherParams[i] <= p.betLimit {
					p.betChip = otherParams[i]
					return true
				}
			}
		}
	}
	return false
}

func (p *HundredYXXPlayer) Action(s *netlib.Session, scene *HundredYXXScene) {
	if base.StartSessionGameTimer(s, timer.TimerActionWrapper(func(h timer.TimerHandle, ud interface{}) bool {
		if scene.GetState() != int32(rule.HundredYXXSceneStateStake) {
			return true
		}
		if time.Now().Before(p.tNextBet) {
			return true
		}

		p.ChoseBetPos()
		p.ChoseBetChip(scene)

		if p.betTotal+p.betChip > p.betLimit {
			//chose smaller chip try
			if !p.TryChoseSmallerChip(scene) {
				return false
			} else { //本次押注，跳过，下次再压
				return true
			}
		}

		//金币低于最低下注限制
		betLimit := int32(0)
		params := scene.GetParamsEx()
		if len(params) != 0 {
			dbGameFree := base.SceneMgrSington.GetSceneDBGameFree(scene.GetRoomId(), params[0])
			if dbGameFree != nil {
				betLimit = dbGameFree.GetBetLimit()
			}
		}
		ownerCoin := p.GetCoin() - int64(p.betTotal)
		if betLimit != 0 && ownerCoin < int64(betLimit) && p.betTotal == 0 {
			return true
		}

		p.betTotal += p.betChip
		//pos := p.betPos[rand.Intn(len(p.betPos))]
		pos := p.betPos
		pack := &hundredyxx.CSHundredYXXOp{
			OpCode: proto.Int32(int32(rule.HundredYXXPlayerOpBet)),
			Params: []int64{int64(pos), int64(p.betChip)},
		}
		proto.SetDefaults(pack)
		s.Send(int(hundredyxx.HundredYXXPacketID_PACKET_CS_HYXX_PLAYEROP), pack)
		logger.Logger.Tracef(">>>>>>snid:%d name:%s coin:%d local totalbet:%d remote totalbet:%d try bet pos:%d use chip:%d ", p.GetSnId(), p.GetName(), p.GetCoin(), p.betTotal, p.GetBetTotal(), pos, p.betChip)
		randInterv := int32(1000)
		if p.GetPos() == HYXX_OLPOS { //在座玩家
			randInterv = 2000
		}

		nextInterv := time.Duration(rand.Int31n(randInterv)+300) * time.Millisecond
		p.tNextBet = time.Now().Add(nextInterv)
		logger.Logger.Tracef(">>>>>>NNNNNNNNNNNNNN snid:%d name:%s next to bet interv:%v ", p.GetSnId(), p.GetName(), nextInterv)
		return true
	}), nil, time.Millisecond*200, -1) {
		randInterv := int32(1000)
		if p.GetPos() == HYXX_OLPOS { //在座玩家
			randInterv = 2000
		}
		p.tNextBet = time.Now().Add(time.Duration(rand.Int31n(randInterv)+300) * time.Millisecond)
	}
}

func (p *HundredYXXPlayer) TryUpBanker(s *netlib.Session, scene *HundredYXXScene) {
	params := scene.GetParamsEx()
	if len(params) != 0 {
		dbGameFree := base.SceneMgrSington.GetSceneDBGameFree(scene.GetRoomId(), params[0])
		if dbGameFree != nil {
			bankerLimit := dbGameFree.GetBanker()
			if p.GetCoin() > int64(bankerLimit) {
				if len(scene.tryUpBanker) < 3 && scene.waitingBankerNum < 3 {
					pack := &hundredyxx.CSHundredYXXOp{
						OpCode: proto.Int32(int32(rule.HundredYXXPlayerOpUpBanker)),
					}
					proto.SetDefaults(pack)
					s.Send(int(hundredyxx.HundredYXXPacketID_PACKET_CS_HYXX_PLAYEROP), pack)
					scene.tryUpBanker[p.GetSnId()] = p
				}
			}
		}
	}
}
