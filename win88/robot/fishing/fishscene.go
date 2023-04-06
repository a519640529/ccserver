package fishing

import (
	"games.yol.com/win88/common"
	"games.yol.com/win88/gamerule/fishing"
	fish_proto "games.yol.com/win88/protocol/fishing"
	player_proto "games.yol.com/win88/protocol/player"
	server_proto "games.yol.com/win88/protocol/server"
	"games.yol.com/win88/robot/base"
	"games.yol.com/win88/srvdata"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
)

type FishingScene struct {
	base.BaseScene
	*fish_proto.SCFishingRoomInfo
	s              *netlib.Session
	dbGameFree     *server_proto.DB_GameFree //自由场数据
	seats          [fishing.MaxPlayer]*FishingPlayer
	players        map[int32]*FishingPlayer
	fishes         map[int32]*Fish
	syncedPolicy   map[int64]struct{}
	logicTick      int32                     //当前刷鱼逻辑时钟
	willExpiredMap map[int32]map[int32]*Fish //即将超期的鱼
}

func NewFishingScene(info *fish_proto.SCFishingRoomInfo) *FishingScene {
	s := &FishingScene{
		SCFishingRoomInfo: info,
		players:           make(map[int32]*FishingPlayer),
		fishes:            make(map[int32]*Fish),
		syncedPolicy:      make(map[int64]struct{}),
		willExpiredMap:    make(map[int32]map[int32]*Fish), //即将超期的鱼
		logicTick:         0,
	}
	s.Init()

	logger.Logger.Trace("NewFishingScene ")

	return s
}

func (this *FishingScene) Init() {
	this.dbGameFree = base.SceneMgrSington.GetSceneDBGameFree(this.GetRoomId(), this.GetGameFreeId())
}

func (this *FishingScene) Clear() {
	for _, player := range this.players {
		player.Clear()
	}

	for _, f := range this.fishes {
		this.DelFish(f)
	}
	this.syncedPolicy = make(map[int64]struct{})
	this.fishes = make(map[int32]*Fish)
	this.willExpiredMap = make(map[int32]map[int32]*Fish)
}

func (this *FishingScene) AddPlayer(p base.Player) {
	if mp, ok := p.(*FishingPlayer); ok {
		this.players[p.GetSnId()] = mp
		this.seats[p.GetPos()] = mp

		logger.Logger.Trace("playernum ", len(this.players))
	}
}

func (this *FishingScene) DelPlayer(snid int32) {
	if p, exist := this.players[snid]; exist && p != nil {
		delete(this.players, snid)
		this.seats[p.GetPos()] = nil
	}
}

func (this *FishingScene) GetPlayerByPos(pos int32) base.Player {
	if pos >= 0 && pos < fishing.MaxPlayer {
		return this.seats[pos]
	}
	return nil
}

func (this *FishingScene) GetPlayerBySnid(snid int32) base.Player {
	if p, exist := this.players[snid]; exist {
		return p
	}
	return nil
}

func (this *FishingScene) GetMe(s *netlib.Session) base.Player {
	if user, ok := s.GetAttribute(base.SessionAttributeUser).(*player_proto.SCPlayerData); ok {
		player := this.GetPlayerBySnid(user.GetData().GetSnId())
		return player
	}
	return nil
}

func (this *FishingScene) GetPlayerCount() int32 {
	return int32(len(this.players))
}

func (this *FishingScene) IsFull() bool {
	return len(this.players) >= int(fishing.MaxPlayer)
}

func (this *FishingScene) IsMatchScene() bool {
	return this.GetRoomId() >= common.MatchSceneStartId && this.GetRoomId() <= common.MatchSceneMaxId
}

func (this *FishingScene) IsCoinScene() bool {
	return this.GetRoomId() >= common.CoinSceneStartId && this.GetRoomId() <= common.CoinSceneMaxId
}

func (this *FishingScene) InitPlayer(s *netlib.Session, p *FishingPlayer) {

}

// 玩家操作
const (
	FishingPlayerOpFire     int32 = iota //开炮
	FishingPlayerOpHitFish               //命中
	FishingPlayerOpSetPower              //切换倍率
	FishingPlayerOpSelVip                //切换VIP鱼炮
	FishingPlayerOpRobotFire
	FishingPlayerOpRobotHitFish
	FishingPlayerOpLeave
	FishingPlayerOpEnter
	FishingPlayerOpAuto
	FishingPlayerOpSelTarget
	FishingPlayerOpFireRate
	FishingRobotOpAuto
	FishingRobotOpSetPower
	FishingPlayerHangup
	FishingRobotWantLeave
)

func (this *FishingScene) UpdateOnlinePlayers(onlinePlayers []int32) {
	for iPos := 0; iPos < fishing.MaxPlayer; iPos++ {
		seat := this.seats[iPos]
		if seat != nil {
			if common.InSliceInt32(onlinePlayers, seat.GetSnId()) == false {
				this.DelPlayer(seat.GetSnId())
			}
		}
	}
}

func (this *FishingScene) AllIsRobot() bool {
	bAllIsRobot := true
	for _, player := range this.players {
		if base.PlayerMgrSington.GetPlayer(player.GetSnId()) == nil {
			bAllIsRobot = false
		}
	}
	return bAllIsRobot
}

func (this *FishingScene) Update(ts int64) {
	//logger.Logger.Trace("FishingScene Update GetPlayerCount ", this.GetPlayerCount())

	//清理过期的鱼
	//this.DelExpiredFish(this.logicTick)
	//this.logicTick++
	//
	//for _, player := range this.players {
	//	player.Update(ts)
	//}
}

func (this *FishingScene) flushFish(policyId, logicTick int32) {
	if _, exist := this.syncedPolicy[common.MakeI64(policyId, logicTick)]; exist {
		//这波鱼已经刷过了
		return
	}

	this.logicTick = logicTick
	policyData := FishPolicyMgrSington.GetFishByTime(policyId, logicTick)
	if len(policyData) <= 0 {
		return
	}

	for _, value := range policyData {
		fishRateData := srvdata.PBDB_FishRateMgr.GetData(value.GetFishId())
		var coin int32
		gold := fishRateData.GetGold()
		if len(gold) > 0 {
			coin = gold[0]
			if len(gold) > 1 {
				coin = int32(common.RandInt(int(gold[0]), int(gold[1])))
			}
		}
		count := value.GetCount() // 获取当前的数量
		for index := int32(0); index < int32(count); index++ {
			id := policyId*1000000 + int32(value.GetId())*100 + int32(index+1) // 作为Fish的唯一标识
			var birthTick = logicTick + index*value.GetRefreshInterval()       // 这个条鱼出生的时间
			var liveTick = birthTick + value.GetTimeToLive()*10                // 这条鱼存货的时间
			fish := NewFish(id, value.GetFishId(), coin, birthTick, liveTick)  // 根据参数生成Fish对象
			if fish != nil {
				this.AddFish(fish)
			}
		}
	}
}

func (this FishingScene) GetFish(id int32) *Fish {
	return this.fishes[id]
}

func (this *FishingScene) AddFish(f *Fish) {
	this.fishes[f.id] = f
	willExpired := this.willExpiredMap[f.dieTick]
	if willExpired == nil {
		willExpired = make(map[int32]*Fish)
		this.willExpiredMap[f.dieTick] = willExpired
	}
	willExpired[f.id] = f
}

func (this *FishingScene) DelFish(f *Fish) {
	delete(this.fishes, f.id)
	if willExpired, exist := this.willExpiredMap[f.dieTick]; exist {
		delete(willExpired, f.id)
	}
}

func (this *FishingScene) DelExpiredFish(logicTick int32) {
	if expired, exist := this.willExpiredMap[logicTick]; exist {
		for _, f := range expired {
			this.DelFish(f)
		}
	}
}

func (this *FishingScene) RandGetOneFish() *Fish {
	for _, f := range this.fishes {
		if f.birthTick > this.logicTick || f.dieTick < this.logicTick { //还未出生|已经死亡
			continue
		}

		return f
	}
	return nil
}
