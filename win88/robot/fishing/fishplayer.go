package fishing

import (
	"fmt"
	"games.yol.com/win88/common"
	"games.yol.com/win88/proto"
	fish_proto "games.yol.com/win88/protocol/fishing"
	hall_proto "games.yol.com/win88/protocol/gamehall"
	server_proto "games.yol.com/win88/protocol/server"
	"games.yol.com/win88/robot/base"
	"github.com/idealeak/goserver/core/logger"
	"math/rand"
	"time"
)

type FishingPlayer struct {
	base.BasePlayers
	*fish_proto.FishingPlayerData
	dbGameFree      *server_proto.DB_GameFree //自由场数据
	scene           *FishingScene
	changePowerTs   int64
	changePower2Ts  int64 //倍率递归变化时间间隔
	lifeTimeTs      int64
	fireTimeTs      int64
	leaveTimeTs     int64
	changeVIPTimeTs int64
	changeVIPTime   int64
	curPower        int32 //当前倍率
	targetPower     int32 //目标倍率
	lockFish        int32 //当前锁定攻击的鱼
	logicTick       int32 //逻辑时钟
	bulletSeq       int32 //子弹增长因子
	bulletId        int32 //子弹编号
}

func NewFishingPlayer(data *fish_proto.FishingPlayerData, sceneEx *FishingScene, gameFreeId, roomId int32) *FishingPlayer {
	p := &FishingPlayer{
		FishingPlayerData: data,
		scene:             sceneEx,
		dbGameFree:        base.SceneMgrSington.GetSceneDBGameFree(roomId, gameFreeId),
		changePowerTs:     0,
		changePower2Ts:    0,
		lifeTimeTs:        0,
		fireTimeTs:        0,
		leaveTimeTs:       0,
		changeVIPTimeTs:   0,
		changeVIPTime:     0,
		curPower:          0,
		targetPower:       0,
		lockFish:          0,
	}
	p.Init()
	return p
}

func (this *FishingPlayer) RandomVipGun() {
	if this.IsRobot() == false {
		return
	}

	sceneType := this.dbGameFree.GetSceneType()

	minValue := int32(0)
	maxValue := int32(1)
	switch sceneType {
	case 1: //1，初级场 0至2中随机1个
		minValue = 0
		maxValue = 2
	case 2: //2，中级场 1至4中随机1个
		minValue = 1
		maxValue = 4
	case 3: //3，高级场 2至6中随机1个
		minValue = 2
		maxValue = 6
	default:
		minValue = 0
		maxValue = 1
	}
	vipLevel := this.GetVIP()
	if vipLevel < maxValue {
		maxValue = vipLevel
		if maxValue <= minValue {
			minValue = 0
			maxValue = 1
		}
	}

	s := base.PlayerMgrSington.GetPlayerSession(this.GetSnId())
	if s != nil {
		vipLevelSelected := rand.Int31n(maxValue-minValue) + minValue
		pack := fish_proto.CSFishingOp{
			OpCode: proto.Int32(FishingPlayerOpSelVip),
		}
		pack.Params = append(pack.Params, int64(vipLevelSelected))
		s.Send(int(fish_proto.FIPacketID_FISHING_CS_OP), pack)
	}

}

func (this *FishingPlayer) RandomPower() {
	if this.IsRobot() == false {
		logger.Logger.Trace("RandomPower not robot", this.GetSnId())
		return
	}

	s := base.PlayerMgrSington.GetPlayerSession(this.GetSnId())
	if s != nil {
		logger.Logger.Trace("RandomPower ", this.GetSnId())

		otherParams := this.dbGameFree.GetOtherIntParams()
		powerIndex := rand.Int31n(int32(len(otherParams)))
		powerValue := otherParams[powerIndex]
		if common.InSliceInt32(otherParams, powerValue) {
			this.targetPower = powerValue
		}
		logger.Logger.Trace("RandomPower ", this.GetSnId(), powerValue)

		this.changePower2Ts = int64(time.Duration(common.RandInt(500, 1000)) * time.Millisecond)
	}
}

func (this *FishingPlayer) CheckInitGameCoin() {
	if this.IsRobot() == false {
		logger.Logger.Trace("CheckInitGameCoin not robot", this.GetSnId())
		return
	}
	s := base.PlayerMgrSington.GetPlayerSession(this.GetSnId())
	if s != nil {
		minValue := int(this.dbGameFree.GetRobotTakeCoin()[0])
		maxValue := int(this.dbGameFree.GetRobotTakeCoin()[1])
		if maxValue == minValue {
			minValue = int(this.dbGameFree.GetBaseScore())
			maxValue = minValue * 100
		}
		if this.GetCoin() < int64(minValue) || this.GetCoin() > int64(maxValue) {
			coins := rand.Intn(maxValue-minValue) + minValue - int(this.GetCoin())
			base.ExePMCmd(s, fmt.Sprintf("%v%v%v", common.PMCmd_AddCoin, common.PMCmd_SplitToken, coins))
		}

		//ExePMCmd(s, fmt.Sprintf("%v%v%v", common.PMCmd_AddCoin, common.PMCmd_SplitToken, 100))
	}
}

func (this *FishingPlayer) Init() {
	//在场内时间
	this.leaveTimeTs = int64(time.Duration(common.RandInt(5, 10)) * time.Minute)
	//换炮间隔
	this.changeVIPTime = int64(time.Duration(common.RandInt(5, 10)) * time.Minute)

	this.RandomVipGun()
	this.RandomPower()
	this.CheckInitGameCoin()
}

func (this *FishingPlayer) Clear() {
	this.logicTick = 0
	this.bulletId = 0
}

func (this *FishingPlayer) UpdatePlayerData(data *fish_proto.FishingPlayerData) {
	this.FishingPlayerData = data
}

func (this *FishingPlayer) MarkFlag(flag int32) {
	myFlag := this.GetFlag()
	myFlag |= flag
	this.Flag = proto.Int32(myFlag)
}

func (this *FishingPlayer) UnmarkFlag(flag int32) {
	myFlag := this.GetFlag()
	myFlag &= ^flag
	this.Flag = proto.Int32(myFlag)
}

func (this *FishingPlayer) IsMarkFlag(flag int32) bool {
	if (this.GetFlag() & flag) != 0 {
		return true
	}
	return false
}

func (this *FishingPlayer) IsOnLine() bool {
	return this.IsMarkFlag(base.PlayerState_Online)
}

func (this *FishingPlayer) IsReady() bool {
	return this.IsMarkFlag(base.PlayerState_Ready)
}

func (this *FishingPlayer) IsSceneOwner() bool {
	return this.IsMarkFlag(base.PlayerState_SceneOwner)
}

func (this *FishingPlayer) IsRobot() bool {
	player := base.PlayerMgrSington.GetPlayer(this.GetSnId())
	return player != nil
}

func (this *FishingPlayer) SetFlag(flag int32) {
	this.Flag = proto.Int32(flag)
}

func (this *FishingPlayer) GetLastOp() int32 {
	return 0
}

func (this *FishingPlayer) SetLastOp(op int32) {
}

func (this *FishingPlayer) UpdateCards(cards []int32) {
}

func (this *FishingPlayer) Update(ts int64) {
	if !this.IsRobot() {
		return
	}
	//切换倍率
	if this.targetPower == this.curPower {
		this.changePowerTs += ts
		if this.changePowerTs >= int64(2*time.Minute) {
			this.changePowerTs = 0

			this.RandomPower()
		}
	} else {
		this.changePower2Ts = this.changePower2Ts - ts
		if this.changePower2Ts <= int64(0*time.Millisecond) {
			if this.targetPower < this.curPower {
				this.curPower = this.curPower - this.dbGameFree.GetBaseScore()
			} else if this.targetPower > this.curPower {
				this.curPower = this.curPower + this.dbGameFree.GetBaseScore()
			}

			s := base.PlayerMgrSington.GetPlayerSession(this.GetSnId())
			if s != nil {
				pack := &fish_proto.CSFishingOp{
					OpCode: proto.Int32(FishingPlayerOpSetPower),
				}
				pack.Params = append(pack.Params, int64(this.curPower), int64(this.targetPower))
				s.Send(int(fish_proto.FIPacketID_FISHING_CS_OP), pack)
				logger.Logger.Tracef("this.curPower =", this.curPower)
			}

			this.changePower2Ts = int64(time.Duration(common.RandInt(500, 1000)) * time.Millisecond)
		}
	}

	//换炮
	this.changeVIPTimeTs += ts
	if this.changeVIPTimeTs >= this.changeVIPTime {
		this.changeVIPTimeTs = 0
		this.RandomVipGun()
	}

	//离场
	this.lifeTimeTs += ts
	if this.lifeTimeTs >= this.leaveTimeTs {
		this.lifeTimeTs = 0

		s := base.PlayerMgrSington.GetPlayerSession(this.GetSnId())
		if s != nil {
			pack := &hall_proto.CSQuitGame{
				Id: proto.Int32(this.dbGameFree.Id),
			}
			proto.SetDefaults(pack)
			s.Send(int(hall_proto.GameHallPacketID_PACKET_CS_QUITGAME), pack)
		}
	}

	//一发子弹一发命中，交替执行
	if this.scene.logicTick%2 == 0 {
		this.Fire()
	} else {
		this.Hit()
	}
}

func (this *FishingPlayer) Fire() {
	s := base.PlayerMgrSington.GetPlayerSession(this.GetSnId())
	if s != nil && this.scene != nil {
		f := this.scene.RandGetOneFish()
		if f != nil {
			this.lockFish = f.id
			this.bulletSeq++
			this.bulletId = this.bulletSeq
			pack := &fish_proto.CSFishingOp{
				OpCode: proto.Int32(FishingPlayerOpFire),
				Params: []int64{rand.Int63n(1280), rand.Int63n(720), int64(this.bulletId), int64(this.curPower)},
			}
			proto.SetDefaults(pack)
			s.Send(int(fish_proto.FIPacketID_FISHING_CS_OP), pack)
			logger.Logger.Infof("FishingPlayer.Fire(snid:%v, bullet:%v)", this.GetSnId(), this.bulletId)
		}
	}
}

func (this *FishingPlayer) Hit() {
	if this.lockFish != 0 && this.bulletId != 0 {
		s := base.PlayerMgrSington.GetPlayerSession(this.GetSnId())
		if s != nil && this.scene != nil {
			f := this.scene.GetFish(this.lockFish)
			if f == nil {
				f = this.scene.RandGetOneFish()
			}
			if f == nil {
				return
			}

			pack := &fish_proto.CSFishingOp{
				OpCode: proto.Int32(FishingPlayerOpHitFish),
				Params: []int64{int64(this.bulletId), int64(f.id)},
			}
			proto.SetDefaults(pack)
			s.Send(int(fish_proto.FIPacketID_FISHING_CS_OP), pack)
			logger.Logger.Infof("FishingPlayer.Hit(snid:%v, bullet:%v)", this.GetSnId(), this.bulletId)
			this.bulletId = 0
		}
	}
}
