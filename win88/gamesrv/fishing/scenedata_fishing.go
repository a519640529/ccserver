package fishing

import (
	"fmt"
	"games.yol.com/win88/gamesrv/base"
	fishing_proto "games.yol.com/win88/protocol/fishing"
	"games.yol.com/win88/srvdata"
	"github.com/cihub/seelog"
	"github.com/idealeak/goserver/core"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/idealeak/goserver/core/timer"

	"games.yol.com/win88/common"
	"games.yol.com/win88/gamerule/fishing"
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	"github.com/idealeak/goserver/core/netlib"
	srvlibproto "github.com/idealeak/goserver/srvlib/protocol"
)

var fishlogger seelog.LoggerInterface

func init() {
	core.RegisteHook(core.HOOK_BEFORE_START, func() error {
		fishlogger = common.GetLoggerInstanceByName("FishLogger")
		return nil
	})
}

type FishBattle struct {
	SnId      int32
	Bullet    int32
	Power     int32
	FishsId   []int
	ExtFishis []int32
	LockFish  int32 // 用来提交捕捉到的指定选鱼
}
type Bullet struct {
	Id       int32
	Power    int32
	SrcId    int32
	LifeTime int32
}
type FishingSceneData struct {
	*base.Scene
	players            map[int32]*FishingPlayerData
	seats              [fishing.MaxPlayer]*FishingPlayerData
	BattleBuff         chan *FishBattle
	PolicyArr          []int32
	PolicyArrIndex     int
	PolicyId           int32
	TimePoint          int32
	NextTime           int64
	StartTime          int64 // 场景创建的
	Policy_Mode        PolicyMode
	MaxTick            int32
	Randomer           *common.RandomGenerator
	fish_list          map[int32]*Fish
	delFish_list       map[int32]int32
	fish_Event         map[int32][]int32
	fishLevel          int32
	frozenTick         int32
	lastTick           int64
	remainder          int64
	intervalTime       int64
	iTime              int64
	lastBossTime       int64 // boss 上一次出现的时间
	lastLittleBossTime int64 // 小boss 上一次出现的时间
	BossId             int32 // 当前场景的BOSS
	BossTag            int32 // 当前场景BOSS是否
	LastID             int32 // 当前场景鱼ID生成器
	hDestroy           timer.TimerHandle

	platform   string
	gameId     int
	sceneType  int
	sceneMode  int
	keyGameId  string //游戏ID
	testing    bool
	gamefreeId int32
	groupId    int32
	agentor    int32
}

func NewFishingSceneData(s *base.Scene) *FishingSceneData {
	return &FishingSceneData{
		Scene:        s,
		players:      make(map[int32]*FishingPlayerData),
		BattleBuff:   make(chan *FishBattle, 1000),
		fish_list:    make(map[int32]*Fish),
		fish_Event:   make(map[int32][]int32),
		delFish_list: make(map[int32]int32),
	}
}
func (this *FishingSceneData) RebindPlayerSnId(oldSnId, newSnId int32) {
	if p, exist := this.players[oldSnId]; exist {
		delete(this.players, oldSnId)
		this.players[newSnId] = p
	}
}
func (this *FishingSceneData) init() bool {
	if this.GetDBGameFree() != nil {
		this.fishLevel = this.GetDBGameFree().GetSceneType() // 更新当前的场此
	}
	this.SetPlayerNum(4)
	this.gameId = this.GetGameId()
	this.platform = this.GetPlatform()
	this.sceneType = int(this.DbGameFree.GetSceneType())
	this.keyGameId = this.GetKeyGameId()
	this.testing = this.GetTesting()
	this.gamefreeId = this.GetGameFreeId()
	this.groupId = this.GetGroupId()
	this.agentor = this.GetAgentor()
	this.sceneMode = this.GetSceneMode()

	this.TimePoint = 0 // 不清楚参数的具体作用 , 从代码逻辑上说 感觉是一个定时器
	this.PolicyId = 0
	this.Policy_Mode = Policy_Mode_Tide // 设置 策略模式 , 策略模式 是  鱼潮
	this.MaxTick = 0
	this.Randomer = &common.RandomGenerator{}
	this.Randomer.RandomSeed(int32(common.RandInt()))
	this.PolicyArr = InitScenePolicyMode(this.GetGameId(), this.sceneType) // 随机生成策略模式 乱序
	this.PolicyArrIndex = 0                                                // 初始化策略模式的  目标索引
	this.lastLittleBossTime = time.Now().Unix()
	this.lastBossTime = time.Now().Unix()
	this.ChangeFlushFish()
	//随机一个初始点
	start := rand.Int31n(this.MaxTick * 4 / 5)
	for i := int32(0); i < start; i++ {
		this.fishFactory()
	}
	this.NextTime -= int64(time.Millisecond * time.Duration(start*100))

	return true
}
func (this *FishingSceneData) Clean() {
	for _, p := range this.players {
		//5款捕鱼保持一致
		fishID, coin, taxc := int32(0), int32(0), int64(0)
		for _, v := range p.bullet {
			coin += v.Power
			taxc += int64(float64(this.GetDBGameFree().GetTaxRate()) / 10000 * float64(v.Power))
		}
		this.RetBulletCoin(p, fishID, coin, taxc, false) // 合并后发送
		p.bullet = make(map[int32]*Bullet)
		p.BulletLimit = [BULLETLIMIT]int64{}
		p.fishEvent = make(map[string]*FishingPlayerEvent)
		p.logFishCount = make(map[int64]*model.FishCoinNum)
	}
	//for debug
	//for key, _ := range this.fish_list {
	//	fishlogger.Tracef("[Clean] delete fish:[%v]", key)
	//}
	this.fish_list = make(map[int32]*Fish)
	this.delFish_list = make(map[int32]int32)
	this.fish_Event = make(map[int32][]int32)
}
func (this *FishingSceneData) BroadcastPlayerLeave(p *base.Player, reason int) {
}
func (this *FishingSceneData) SceneDestroy(force bool) {
	this.Scene.Destroy(force)
}

func (this *FishingSceneData) RetBulletCoin(player *FishingPlayerData, fishID, coin int32, taxc int64, flag bool) {
	if player.IsRob && !model.GameParamData.IsRobFightTest {
		return
	}
	pack := &fishing_proto.SCFishDel{
		FishId:            proto.Int32(int32(fishID)),
		Coin:              proto.Int32(coin),
		CurrentPlayerCoin: proto.Int64(player.CoinCache),
	}
	proto.SetDefaults(pack)
	if !this.GetTesting() && !player.IsRob { //所有捕鱼统一
		player.NewStatics(int64(-coin), 0)
	}
	//taxc := int64(float64(this.GetDBGameFree().GetTaxRate()) / 10000 * float64(coin))
	player.LostCoin -= int64(coin)
	player.CoinCache += int64(coin)
	player.SetTotalBet(player.GetTotalBet() - int64(coin))
	//player.Statics(this.KeyGameId, this.KeyGamefreeId, int64(coin), false)
	player.LastRechargeWinCoin -= int64(coin)

	if !this.GetTesting() {
		player.taxCoin -= float64(taxc)
		player.sTaxCoin -= float64(taxc)
	}
	fishlogger.Trace("RetBulletCoin : ", fishID, player.IsRob, coin, player.LostCoin, player.CoinCache, player.GetTotalBet())
	base.GetCoinPoolMgr().PopCoin(this.GetGameFreeId(), this.GetGroupId(), this.GetPlatform(), int64(coin)-taxc)
	tax := base.GetCoinPoolMgr().GetTax()
	base.GetCoinPoolMgr().SetTax(tax - float64(taxc))
	pack.CurrentPlayerCoin = proto.Int64(player.CoinCache)
	player.SendToClient(int(fishing_proto.FIPacketID_FISHING_SC_FISHDEL), pack)
}

/*
 * 玩家相关
 */
func (this *FishingSceneData) EnterPlayer(player *FishingPlayerData) bool {
	pos := -1
	emptyPos := []int{}
	for i := 0; i < fishing.MaxPlayer; i++ {
		if this.seats[i] == nil {
			emptyPos = append(emptyPos, i)
		}
	}
	if len(emptyPos) > 0 {
		pos = emptyPos[common.RandInt(len(emptyPos))]
	} else {
		fishlogger.Error("Fishing enter player find pos error.")
	}
	player.SetPos(pos)
	player.CoinCache = player.GetCoin()
	player.LostCoin = 0
	player.LostCoinCache = 0
	player.FishPoolKey = fmt.Sprintf("%v-%v", this.GetGameFreeId(), player.Platform)
	this.players[player.SnId] = player
	this.seats[pos] = player
	this.OnEnterPlayer(player)
	return true
}
func (this *FishingSceneData) OnEnterPlayer(player *FishingPlayerData) {
	/*
		如果进场的是机器人，寻找场内负载最小的玩家进行绑定
		如果进场的是玩家，寻找场内没有代理的机器人
	*/
	var tempPlayer *FishingPlayerData //新逻辑
	if player.IsRob {
		for i := 0; i < fishing.MaxPlayer; i++ {
			curSeatPlayer := this.seats[i]
			if curSeatPlayer != nil && !curSeatPlayer.IsRob {
				//新逻辑
				if tempPlayer == nil {
					tempPlayer = curSeatPlayer
				} else if len(tempPlayer.RobotSnIds) > len(curSeatPlayer.RobotSnIds) {
					tempPlayer = curSeatPlayer
				}

				//老逻辑
				if curSeatPlayer.AgentParam == 0 {
					curSeatPlayer.AgentParam = player.SnId
					player.AgentParam = curSeatPlayer.SnId
					curSeatPlayer.RobotSnIds = append(curSeatPlayer.RobotSnIds, player.SnId)
					break
				}
			}
		}

		//新逻辑 设置机器人的代理情况
		if tempPlayer != nil && player.AgentParam == 0 {
			player.AgentParam = tempPlayer.SnId
			tempPlayer.AgentParam = player.SnId
			tempPlayer.RobotSnIds = append(tempPlayer.RobotSnIds, player.SnId)
		}

	} else {
		for i := 0; i < fishing.MaxPlayer; i++ {
			curSeatRobot := this.seats[i]
			if curSeatRobot != nil && curSeatRobot.IsRob {
				if curSeatRobot.AgentParam == 0 { //无主机器人
					curSeatRobot.AgentParam = player.SnId
					player.AgentParam = curSeatRobot.SnId
					player.RobotSnIds = append(player.RobotSnIds, curSeatRobot.SnId) //新逻辑
				}
			}
		}
		if len(player.RobotSnIds) == 0 {
			for i := 0; i < fishing.MaxPlayer; i++ {
				curSeatRobot := this.seats[i]
				if curSeatRobot != nil && curSeatRobot.IsRob {
					if curSeatRobot.AgentParam != 0 {
						//需要重新平衡下机器人负载
						for j := 0; j < fishing.MaxPlayer; j++ {
							curSeatPlayer := this.seats[j]
							if curSeatPlayer != nil && !curSeatPlayer.IsRob && curSeatPlayer.SnId == curSeatRobot.AgentParam && len(curSeatPlayer.RobotSnIds) > 1 {
								//bind agent
								curSeatRobot.AgentParam = player.SnId
								player.AgentParam = curSeatRobot.SnId
								player.RobotSnIds = append(player.RobotSnIds, curSeatRobot.SnId)
								//unbind agent
								curSeatPlayer.RobotSnIds = common.DelSliceInt32(curSeatPlayer.RobotSnIds, curSeatRobot.SnId)
								curSeatPlayer.AgentParam = curSeatPlayer.RobotSnIds[0]
								break
							}
						}
					}
				}
				//分担一个就可以了
				if len(player.RobotSnIds) != 0 {
					break
				}
			}
		}
	}
	return
}
func (this *FishingSceneData) QuitPlayer(player *FishingPlayerData, reason int) bool {
	if _, ok := this.players[player.SnId]; ok {
		player.SetSelVip(this.KeyGameId)
		player.SaveDetailedLog(this.Scene)
		delete(this.players, player.SnId)
		this.seats[player.GetPos()] = nil
		this.BroadcastPlayerLeave(player.Player, 0)
		diffCoin := player.CoinCache - player.GetCoin()
		player.AddCoin(diffCoin, common.GainWay_Fishing, base.SyncFlag_ToClient, "system", this.GetSceneName())
		if diffCoin != 0 || player.LostCoin != 0 {
			if !player.IsRob && !this.GetTesting() {

				player.AddServiceFee(int64(player.taxCoin))
			}
			player.SetGameTimes(player.GetGameTimes() + 1)
			if diffCoin > 0 {
				//player.winTimes++
				player.SetWinTimes(player.GetWinTimes() + 1)
			} else {
				//player.lostTimes++
				player.SetLostTimes(player.GetLostTimes() + 1)
			}
		}
		this.OnQuitPlayer(player, reason)
		return true
	} else {
		return false
	}
}
func (this *FishingSceneData) OnQuitPlayer(player *FishingPlayerData, reason int) {
	/*
		如果离场的是机器人，从代理玩家身上将其删除
		如果离场的是玩家，为玩家身上代理的机器人寻找其他代理
	*/
	if player.IsRob { //机器人离场
		for i := 0; i < fishing.MaxPlayer; i++ {
			curSeatPlayer := this.seats[i]
			if curSeatPlayer != nil && !curSeatPlayer.IsRob {
				//老逻辑，不改动
				if curSeatPlayer.AgentParam == player.SnId {
					curSeatPlayer.AgentParam = 0
					player.AgentParam = 0
				}
				//新逻辑
				if common.InSliceInt32(curSeatPlayer.RobotSnIds, player.SnId) {
					curSeatPlayer.RobotSnIds = common.DelSliceInt32(curSeatPlayer.RobotSnIds, player.SnId)
					player.AgentParam = 0
					if len(curSeatPlayer.RobotSnIds) > 0 {
						curSeatPlayer.AgentParam = curSeatPlayer.RobotSnIds[0]
					} else {
						curSeatPlayer.AgentParam = 0
					}
				}
			}
		}
	} else { //玩家离场
		for i := 0; i < fishing.MaxPlayer; i++ {
			curSeatRobot := this.seats[i]
			if curSeatRobot != nil && curSeatRobot.IsRob {
				if curSeatRobot.AgentParam == player.SnId {
					curSeatRobot.AgentParam = 0
					player.AgentParam = 0

					var tempPlayer *FishingPlayerData //新逻辑 负载最小的玩家
					for j := 0; j < fishing.MaxPlayer; j++ {
						curSeatPlayer := this.seats[j]
						if curSeatPlayer != nil && !curSeatPlayer.IsRob && curSeatPlayer != player {
							if curSeatPlayer.AgentParam == 0 {
								curSeatPlayer.AgentParam = curSeatRobot.SnId
								curSeatRobot.AgentParam = curSeatPlayer.SnId
								curSeatPlayer.RobotSnIds = append(curSeatPlayer.RobotSnIds, curSeatRobot.SnId) //新逻辑
								pack := &fishing_proto.SCReBindAgent{
									PlayerSnid: proto.Int32(curSeatPlayer.SnId),
									RobSnid:    proto.Int32(curSeatRobot.SnId),
								}
								curSeatPlayer.SendToClient(int(fishing_proto.FIPacketID_FISHING_SC_REBINDAGENT), pack)
								break
							} else {
								if tempPlayer == nil {
									tempPlayer = curSeatPlayer
								} else if len(tempPlayer.RobotSnIds) > len(curSeatPlayer.RobotSnIds) {
									tempPlayer = curSeatPlayer
								}
							}
						}
					}

					//新逻辑
					if tempPlayer != nil && curSeatRobot.AgentParam == 0 {
						curSeatRobot.AgentParam = tempPlayer.SnId
						tempPlayer.AgentParam = curSeatRobot.SnId
						tempPlayer.RobotSnIds = append(tempPlayer.RobotSnIds, curSeatRobot.SnId)
						pack := &fishing_proto.SCReBindAgent{
							PlayerSnid: proto.Int32(tempPlayer.SnId),
							RobSnid:    proto.Int32(curSeatRobot.SnId),
						}
						tempPlayer.SendToClient(int(fishing_proto.FIPacketID_FISHING_SC_REBINDAGENT), pack)
					}
				}
			}
		}
		player.RobotSnIds = nil
	}
}
func (this *FishingSceneData) OnTick() {
	if this.TimePoint%10 == 0 {
		this.fishTimeOut()
	}
}
func (this *FishingSceneData) fishTimeOut() {
	for key, value := range this.fish_list {
		if value == nil {
			delete(this.fish_list, key)
			//fishlogger.Tracef("[fishTimeOut]1 delete fish:[%v]", key)
			continue
		}
		if value.LiveTick <= this.TimePoint {
			delete(this.fish_list, key)
			//fishlogger.Tracef("[fishTimeOut]2 delete fish:[%v]", key)
			if value.Event != 0 {
				this.fish_Event[value.Event] = common.DelSliceInt32(this.fish_Event[value.Event], value.FishID)
			}
		}
	}
}
func (this *FishingSceneData) AddFish(id int32, fish *Fish) {
	this.fish_list[id] = fish
}
func (this *FishingSceneData) DelFish(id int32) {
	if _, exist := this.fish_list[id]; exist {
		delete(this.fish_list, id)
		//fishlogger.Tracef("[DelFish] delete fish:[%v]", id)
		this.delFish_list[id] = 1
	}
}

/*
 * 捕鱼相关
 */
func (this *FishingSceneData) fishBattle() {
	for i := 0; i < 12; i++ {
		select {
		case data := <-this.BattleBuff:
			player := this.players[data.SnId]
			if player == nil {
				fishlogger.Tracef("Bullet %v owner %v offline.", data.Bullet, data.SnId)
				continue
			}
			delete(player.bullet, data.Bullet) // 二次清楚 防止没有删掉
			var count = len(data.FishsId)
			if count > 0 && data.Power > 0 {
				this.fishProcess(player, data.FishsId, data.Power, data.ExtFishis)
			}
		default:
			break
		}
	}
}

// 捕鱼击中的标准处理逻辑
func (this *FishingSceneData) fishProcess(player *FishingPlayerData, fishIds []int, power int32, extfishis []int32) {
	if len(fishIds) == 0 {
		return
	}

	//调试辅助
	sendMiss := func(fishid, rate int32) {
		if player.GMLevel > 0 {
			pack := &fishing_proto.SCFireMiss{
				FishId: proto.Int32(fishid),
				Rate:   proto.Int32(rate),
			}
			proto.SetDefaults(pack)
			player.SendToClient(int(fishing_proto.FIPacketID_FISHING_SC_FIREMISS), pack)
		}
	}

	extFishMap := make(map[int32]struct{})
	dropcoinext := int32(0)
	for _, v := range extfishis {
		if _, exist := extFishMap[v]; !exist {
			extFishMap[v] = struct{}{}
			var extfish = this.fish_list[v]
			if extfish == nil {
				continue
			}
			dropcoinext += extfish.DropCoin
		}
	}

	var killRate int32
	var death bool
	var robot = player.IsRob
	var ts = time.Now().Unix()
	hitFishMap := make(map[int32]struct{})
	for _, id := range fishIds {
		hitFishMap[int32(id)] = struct{}{}
	}
	for value, _ := range hitFishMap {
		var fish = this.fish_list[value] // 取出对应的鱼的概率
		// start  判断当前的鱼是否有效
		if fish == nil {
			fishlogger.Tracef("[fishProcess] Be hit fish [%v] is disappear.", value)
			taxc := int64(float64(this.GetDBGameFree().GetTaxRate()) / 10000 * float64(power))
			if player.powerType != FreePowerType {
				// 不是免费炮期间返还
				this.RetBulletCoin(player, int32(value), power, taxc, true)
			}
			//}
			//鱼不存在
			sendMiss(int32(value), -1)
			continue
		}
		//end
		//特殊鱼处理
		if fish.TemplateID == fishing.Fish_CaiShen && fish.DropCoin < fish.MaxDropCoin {
			fish.DropCoin++
			this.syncFishCoin(fish)
		}

		//同组的鱼(一网打尽)
		var groupcoinex int32
		if fish.Event > 0 && fish.Event <= fishing.Event_Group_Max {
			groupFishs := this.fish_Event[fish.Event]
			for _, fishId := range groupFishs {
				if fishId == value { //去重
					continue
				}
				if _, exist := extFishMap[value]; exist { //去重
					continue
				}
				fishg := this.fish_list[fishId]
				if fishg == nil {
					continue
				}
				if fishg.IsDeath(this.TimePoint) {
					continue
				}
				if !fishg.IsBirth(this.TimePoint) {
					continue
				}
				groupcoinex += fishg.DropCoin
			}
		}
		//同组的鱼(一网打尽)

		death = false
		if (robot && !model.GameParamData.IsRobFightTest) || this.GetTesting() {
			//体验场概率稍有提升
			killRate = 10000 / (fish.DropCoin + dropcoinext + groupcoinex)
			if this.GetTesting() {
				killRate *= 2
			}
			if rand.Int31n(10000) < killRate {
				death = true
			}
			// todo 强制设置机器人 鱼死不 (测试使用)
			//death = false
		} else { //欢乐捕鱼|李逵劈鱼...其他捕鱼都走这个算法
			// start 记录相关鱼的击中次数
			key := player.MakeLogKey(fish.TemplateID, power)
			if v, ok := player.logFishCount[key]; ok {
				v.HitNum++
			} else {
				player.logFishCount[key] = &model.FishCoinNum{
					HitNum: 1,
					ID:     fish.TemplateID,
					Power:  power,
				}
			}
			//击中计数
			player.logBulletHitNums++
			// end
			setting := base.GetCoinPoolMgr().GetCoinPoolSetting(this.GetPlatform(), this.GetGameFreeId(), this.GetGroupId()) // 获取当前水池的配置表
			if setting == nil {
				fishlogger.Error("GetCoinPoolSetting is nil")
				death = false
			} else {
				key := strconv.Itoa(int(this.GetDBGameFree().GetId()))
				rate := float64(0)
				// 判断鱼是否死亡
				ctroRate := setting.GetCtroRate()
				baseScore := this.GetDBGameFree().GetBaseScore()
				death, rate = fish.HappyFishRate(int(this.fishLevel), fish.DropCoin+dropcoinext+groupcoinex, player, this.GetKeyGameId(), baseScore, ctroRate, key, power)
				player.realOdds = int(rate * 10000)
				killRate = int32(rate * 10000)
			}
		}

		if common.Config.IsDevMode {
			//test code
			death = true
		}

		if !death {
			//鱼没死
			sendMiss(int32(value), killRate)
			continue
		}
		// 判断当前是否是 特殊鱼
		deathFishs := this.fishEvent(fish, player, power, ts, extfishis)
		if len(deathFishs) == 0 {
			//鱼没死
			sendMiss(int32(value), killRate)
			continue
		}
		//无效代码
		//if this.RequireCoinPool(player) && len(deathFishs) > 1 {
		//	dropCoin := int32(0)
		//	for _, v := range deathFishs {
		//		dropCoin += v.DropCoin
		//	}
		//	if int(float64(player.realOdds+1)/float64(dropCoin+1)) < common.RandInt(10000) {
		//		continue
		//	}
		//}
		this.fishSettlements(deathFishs, player, power, fish.Event, ts, 0, 0)
	}

	//写条记录
	if player.logBulletHitNums >= CountSaveNums {
		player.SaveDetailedLog(this.Scene)
	}
}

/*
event  对应得事件ID
*/
func (this *FishingSceneData) EventTreasureChestSettlements(power int32) (int32, []int32) {
	// 计算龙王多播相关得额外收益
	var totalWeight int32
	var eventCoin int32
	for _, weight := range fishing.TreasureChestWeight {
		totalWeight = totalWeight + weight
	}
	NowWeight := rand.Int31n(totalWeight)
	var cumulativeWeight int32
	for index, weight := range fishing.TreasureChestWeight {
		cumulativeWeight = cumulativeWeight + weight
		if NowWeight <= cumulativeWeight {
			treasureChestReward := fishing.TreasureChestReward[index]
			for _, value := range treasureChestReward {
				eventCoin = eventCoin + power*value
			}
			return eventCoin, treasureChestReward
		}
	}
	return 0, []int32{}
}

func (this *FishingSceneData) fishSettlements(fishs []*Fish, player *FishingPlayerData, power int32, event int32,
	ts int64, eventFishId int32, eventFishCoin int32) {
	var coin int64 // 鱼死亡本身得金币计算
	var treasureChestReward []int32
	pack := &fishing_proto.SCFireHit{
		Snid:      proto.Int32(player.SnId),
		Ts:        proto.Int64(ts),
		EventFish: proto.Int32(eventFishId),
		EventCoin: proto.Int32(eventFishCoin * power),
		Power:     proto.Int32(power),
	}

	if event != 0 {
		pack.Event = proto.Int32(event)
	}
	for _, value := range fishs {
		var dropCoin int32
		if value.Event == fishing.Event_Bit {
			dropCoin = 0
		} else if value.Event == fishing.Event_TreasureChest {
			dropCoin, treasureChestReward = this.EventTreasureChestSettlements(power)
			fishlogger.Infof("Event_TreasureChest  eventCoin %v  treasureChestReward %v", dropCoin, treasureChestReward)
		} else if value.Event == fishing.Event_NewBoom {
			dropCoin = 0
		} else if value.Event == fishing.Event_FreePower {
			dropCoin = 0
		} else {
			dropCoin = value.DropCoin * power
		}
		pack.FishId = append(pack.FishId, value.FishID)
		pack.Coin = append(pack.Coin, dropCoin)
		key := player.MakeLogKey(value.TemplateID, power)
		if v, ok := player.logFishCount[key]; ok {
			v.Coin += dropCoin
			v.Num++
		} else {
			player.logFishCount[key] = &model.FishCoinNum{
				Coin:  dropCoin,
				Num:   1,
				Power: power,
			}
		}
		fishlogger.Infof("logFishCount %v,%v,%v,%v", value.TemplateID, power, player.logFishCount[key].Coin, player.logFishCount[key].HitNum)
		coin = coin + int64(dropCoin)
		this.DelFish(value.FishID)
		value.SetDeath()
	}
	if event == fishing.Event_FreePower {
		player.UpdateFreePowerState()
		fishlogger.Infof("snid %v 更新为免费炮台", player.SnId)
	}
	if !this.GetTesting() && !player.IsRob { //5款捕鱼保持统一
		player.NewStatics(0, coin)
	}
	player.winCoin += coin
	player.CoinCache += coin
	fishlogger.Infof("fishSettlements player %v coin %v dropCoin %v", player.SnId, player.CoinCache, coin)
	pack.CurrentPlayerCoin = proto.Int64(player.CoinCache)
	proto.SetDefaults(pack)
	this.BroadCastMessage(int(fishing_proto.FIPacketID_FISHING_SC_FIREHIT), pack, 0)
	if _, ok := player.TodayGameData.CtrlData[this.GetKeyGameId()]; !ok {
		gs := &model.PlayerGameStatics{}
		player.TodayGameData.CtrlData[this.GetKeyGameId()] = gs
	}
	// 发送协议
	if event == fishing.Event_TreasureChest {
		eventTreasureChestPack := &fishing_proto.SCTreasureChestEvent{
			Snid:              proto.Int32(player.SnId),
			Reward:            treasureChestReward,
			CurrentPlayerCoin: proto.Int64(player.CoinCache),
		}
		proto.SetDefaults(eventTreasureChestPack)
		this.BroadCastMessage(int(fishing_proto.FIPacketID_FISHING_SC_TREASURECHESTEVENT), eventTreasureChestPack, 0)
		fishlogger.Infof("Event_TreasureChest  BroadCastMessage  %v", fishing_proto.FIPacketID_FISHING_SC_TREASURECHESTEVENT)

	}
	//player.Statics(this.KeyGameId, this.KeyGamefreeId, coin, true)
	//if coin > 0 {
	//	player.TodayGameData.CtrlData[this.GetKeyGameId()].TotalOut += coin
	//} else {
	//	player.TodayGameData.CtrlData[this.GetKeyGameId()].TotalIn += -coin
	//}

	if !player.IsRob || model.GameParamData.IsRobFightTest /*&& !this.GetTesting()*/ {
		base.GetCoinPoolMgr().PopCoin(this.GetGameFreeId(), this.GetGroupId(), this.GetPlatform(), int64(coin))
	}
}
func (this *FishingSceneData) PushBattle(player *FishingPlayerData, bulletid int32, fishs []int32, extfishis []int32) {
	bullet := player.bullet[bulletid]
	fishlogger.Infof("PushBattle player %v bullet %v fishs %v extfishis %v", player.SnId, bulletid, fishs, extfishis)
	if bullet == nil {
		fishlogger.Infof("%v not find in %v bullet buff. PushBattle player %v", bulletid, player.GetName(), player.SnId)
		return
	}

	battleData := &FishBattle{
		SnId:      bullet.SrcId,
		Bullet:    bullet.Id,
		Power:     bullet.Power,
		ExtFishis: extfishis,
	}
	if len(fishs) > 0 {
		battleData.FishsId = append(battleData.FishsId, int(fishs[0]))
	}

	select {
	case this.BattleBuff <- battleData:
		{
			delete(player.bullet, battleData.Bullet)
		}
	default:
		{
			delete(player.bullet, battleData.Bullet)
			fishlogger.Error("Player battle buff full.")
		}
	}
}

/*
向前端推送 绑定机器人对应的行为
behaviorCode  0  是 机器人 静默行为
*/
func (this *FishingSceneData) SCRobotBehavior(snid, robotId int32, behaviorCode int32) {
	player := this.players[snid]
	if player == nil {
		fishlogger.Errorf("SCRobotBehavior player %v is empty,bullet will be droped.", snid)
		return
	}
	pack := &fishing_proto.SCRobotBehavior{
		Code:    proto.Int32(behaviorCode),
		RobotId: proto.Int32(robotId),
	}
	proto.SetDefaults(pack)
	player.SendToClient(int(fishing_proto.FIPacketID_FISHING_SC_SCROBOTBEHAVIOR), pack)
}
func (this *FishingSceneData) PushBullet(s *base.Scene, snid, x, y, id, power int32) fishing_proto.OpResultCode {
	player := this.players[snid]
	if player == nil {
		fishlogger.Errorf("player %v is empty,bullet will be droped.", snid)
		return fishing_proto.OpResultCode_OPRC_Error
	}
	// end
	if power <= 0 || power != player.power {
		fishlogger.Tracef("[%v %v] power is invalid(%v) currpower(%v).", player.GetName(), player.SnId, power, player.power)
		return fishing_proto.OpResultCode_OPRC_Error
	}
	// start  检测当前玩家的的金币数够不够
	if !player.CoinCheck(power) {
		fishlogger.Tracef("%v no enough coin to fishing.", player.GetName())
		return fishing_proto.OpResultCode_OPRC_CoinNotEnough
	}
	// end
	// curTime 是 当前时间的时间戳 对 2048 取余
	curTime := time.Now().Unix() % BULLETLIMIT
	sbl := player.BulletLimit[curTime]
	// 子弹的数量限制,统一按最高倍速处理,配合10秒窗口期
	bulletCountLimit := int64(12)
	// start  玩家的开火率 》 0 的时候   子弹的数量限制 设置为 10
	//if player.FireRate > 0 {
	//	bulletCountLimit = 10
	//}

	// end
	// start  判断子弹的数量 是否过大
	if sbl > bulletCountLimit {
		//10秒的窗口期，避免堆包误判
		total := sbl
		for i := 1; i < WINDOW_SIZE; i++ {
			total += player.BulletLimit[(int(curTime)-i+BULLETLIMIT)%BULLETLIMIT]
		}
		if total/WINDOW_SIZE > bulletCountLimit {
			fishlogger.Infof("Player bullet too fast.")
			//子弹打飞机了~~~
			key := player.MakeLogKey(0, power)
			if v, ok := player.logFishCount[key]; ok {
				v.HitNum++
			} else {
				player.logFishCount[key] = &model.FishCoinNum{
					HitNum: 1,
					ID:     0, //飞机???
					Power:  power,
				}
			}
			return fishing_proto.OpResultCode_OPRC_Error
		}
	} else {
		player.BulletLimit[curTime] = sbl + 1
	}
	// end
	player.bullet[id] = &Bullet{
		Id:       id,
		Power:    power,
		SrcId:    player.SnId,
		LifeTime: 0,
	}

	if !this.GetTesting() && !player.IsRob { //5款捕鱼保持统一
		player.NewStatics(int64(power), 0)
	}
	// start 对玩家身上的金币变化进行变更
	if player.powerType != FreePowerType { // 只有当前炮台不是免费炮台的时候,才计算相关金额
		player.LostCoin += int64(power)
		player.CoinCache -= int64(power)
		//fishlogger.Infof("player %v coin %v", player.SnId, player.CoinCache)
		player.SetTotalBet(player.GetTotalBet() + int64(power))
		//if this.GetDBGameFree().GetGameId() != int32(common.GameId_NFishing) {
		//player.Statics(this.KeyGameId, this.KeyGamefreeId, -int64(power), true)
		player.LastRechargeWinCoin += int64(power)
		if _, ok := player.TodayGameData.CtrlData[this.GetKeyGameId()]; !ok {
			gs := &model.PlayerGameStatics{}
			player.TodayGameData.CtrlData[this.GetKeyGameId()] = gs
		}
		player.TodayGameData.CtrlData[this.GetKeyGameId()].TotalIn += int64(power) // 添加每日写入数据
		// end
		//}
		// start 根据税收调整对应的比例 , 并且调整整个水池的逻辑
		if !player.IsRob || model.GameParamData.IsRobFightTest /*&& this.GetDBGameFree().GetGameId() != int32(common.GameId_NFishing)*/ /*&& !this.GetTesting()*/ {
			var taxc int64
			taxc = int64(float64(this.GetDBGameFree().GetTaxRate()) / 10000 * float64(power))
			//if this.GetDBGameFree().GetGameId() != int32(common.GameId_NFishing) { // 寻龙夺宝库存读写后移
			coinPoolMgr := base.GetCoinPoolMgr()
			coinPoolMgr.PushCoin(this.GetGameFreeId(), this.GetGroupId(), this.GetPlatform(), int64(power)-taxc)
			coinPoolMgr.SetTax(coinPoolMgr.GetTax() + float64(taxc))
			//}

			if !this.GetTesting() {
				player.taxCoin += float64(taxc)
				player.sTaxCoin += float64(taxc)
			}
		}
	} else {
		player.FreePowerNum--
		if player.FreePowerNum == 0 {
			player.UpdateNormalPowerState() // 状态变更
			fishlogger.Infof("snid %v , 更新为普通炮台", player.SnId)
		}
	}

	// end
	pack := &fishing_proto.SCFire{
		Snid:              proto.Int32(player.SnId),
		X:                 proto.Int32(x),
		Y:                 proto.Int32(y),
		Bulletid:          proto.Int32(id),
		Power:             proto.Int32(power),
		CurrentPlayerCoin: proto.Int64(player.CoinCache),
	}
	proto.SetDefaults(pack)
	this.BroadCastMessage(int(fishing_proto.FIPacketID_FISHING_SC_FIRE), pack, 0)
	//if player.GameData[this.GetKeyGameId()].GameTimes%15 == 0 {
	//	curCoin := coinPoolMgr.LoadCoin(this.GetGameFreeId(), this.GetPlatform(), this.GetGroupId())
	//	player.SaveFishingLog(curCoin, this.GetKeyGameId())
	//}
	return fishing_proto.OpResultCode_OPRC_Sucess
}
func (this *FishingSceneData) SyncFish(player *base.Player) {
	var cnt int
	var fishes []*fishing_proto.FishInfo
	syncFish := false
	pack := &fishing_proto.SCSyncFishCoin{}
	for _, fish := range this.fish_list {
		if fish.IsDeath(this.TimePoint) {
			continue
		}
		if this.Policy_Mode == Policy_Mode_Normal && fish.BirthTick < this.TimePoint &&
			(this.TimePoint-fish.BirthTick > 200) && fish.IsBoss == 0 {
			continue
		}
		if fish.FishID/1000000 != this.PolicyId {
			continue
		}
		fishes = append(fishes, &fishing_proto.FishInfo{
			FishID:    proto.Int32(fish.FishID),
			Path:      proto.Int32(fish.Path),
			PolicyId:  proto.Int32(int32(fish.Policy.GetId())),
			BirthTime: proto.Int32(this.TimePoint - fish.BirthTick),
		})
		cnt++
		if fish.TemplateID == fishing.Fish_CaiShen {
			pack.FishId = proto.Int32(fish.FishID)
			pack.Coin = proto.Int64(int64(fish.DropCoin))
			syncFish = true
		}
	}
	fishlogger.Trace("Current fish list count:", cnt)
	fishlogger.Trace("Current timePoint :", this.TimePoint)
	packFishes := &fishing_proto.SCFishesEnter{
		PolicyId: proto.Int32(this.PolicyId),
		Fishes:   fishes,
		IceSec:   proto.Int32(int32(0)),
		TimeTick: proto.Int32(this.TimePoint),
	}
	proto.SetDefaults(packFishes)
	player.SendToClient(int(fishing_proto.FIPacketID_FISHING_SC_FISHERENTER), packFishes)

	if syncFish {
		player.SendToClient(int(fishing_proto.FIPacketID_FISHING_SC_SCSYNCFISHCOIN), pack)
	}
}

func (this *FishingSceneData) NewSyncFish(player *base.Player) {
	var cnt int
	var fishes []*fishing_proto.FishInfo
	for _, fish := range this.fish_list {
		if fish.IsDeath(this.TimePoint) {
			continue
		}
		if this.Policy_Mode == Policy_Mode_Normal && fish.BirthTick < this.TimePoint &&
			(this.TimePoint-fish.BirthTick > 200) && fish.IsBoss == 0 {
			continue
		}
		if this.Policy_Mode == Policy_Mode_Tide {
			// 只有在鱼潮的时候进行判断
			if fish.FishID/1000000 != this.PolicyId {
				continue
			}
		}
		fishes = append(fishes, &fishing_proto.FishInfo{
			FishID:    proto.Int32(fish.FishID),
			Path:      proto.Int32(fish.Path),
			PolicyId:  proto.Int32(int32(fish.Policy.GetId())),
			BirthTime: proto.Int32(this.TimePoint - fish.BirthTick),
		})
		cnt++
	}
	fishlogger.Trace("NewSyncFish Current fish list count:", cnt)
	fishlogger.Trace("NewSyncFish Current timePoint :", this.TimePoint)
	packFishes := &fishing_proto.SCFishesEnter{
		PolicyId: proto.Int32(this.PolicyId),
		Fishes:   fishes,
		IceSec:   proto.Int32(int32(0)),
		TimeTick: proto.Int32(this.TimePoint),
	}
	proto.SetDefaults(packFishes)
	player.SendToClient(int(fishing_proto.FIPacketID_FISHING_SC_FISHERENTER), packFishes)
}

func (this *FishingSceneData) flushFish() {
	pack := &fishing_proto.SCSyncRefreshFish{
		PolicyId:   proto.Int32(this.PolicyId),
		TimePoint:  proto.Int32(this.TimePoint),
		RandomSeed: proto.Int32(this.Randomer.GetRandomSeed()),
	}
	proto.SetDefaults(pack)
	this.BroadCastMessage(int(fishing_proto.FIPacketID_FISHING_SC_SYNCFISH), pack, 0)
}

func (this *FishingSceneData) NewFlushFish(info []*fishing_proto.SyncRefreshFishInfo) {
	pack := &fishing_proto.SCNewSyncRefreshFish{
		PolicyId:   proto.Int32(this.PolicyId),
		Info:       info,
		LatestID:   proto.Int32(this.LastID),
		RandomSeed: proto.Int32(this.Randomer.GetRandomSeed()),
	}
	proto.SetDefaults(pack)
	fishlogger.Infof("当前桌子ID %v ,NewFlushFish  pack  PolicyId %v , Info %v , latestID %v , RandomSeed %v", this.GetSceneId(), pack.GetPolicyId(), pack.GetInfo(), pack.GetLatestID(), pack.GetRandomSeed())
	this.BroadCastMessage(int(fishing_proto.FIPacketID_FISHING_SC_NEWSYNCFISH), pack, 0)
}

// 鱼类公场生成NEW (目前只有寻龙夺宝使用新的鱼类生成工厂)
func (this *FishingSceneData) NewFishFactory() {
	//  新机制 不需要 this.MaxTick
	/*	if this.TimePoint >= this.MaxTick {
		return
	}*/
	this.TimePoint++ // 存活时间得自增累加
	if this.frozenTick > 0 {
		this.frozenTick--
		return
	}
	if this.Policy_Mode == Policy_Mode_Tide {
		// 当前是鱼潮模式 (鱼潮模式依旧用老版的出鱼模式)
		// 获取当前所有鱼的策略信息
		//fishlogger.Infof("NewFishFactory 鱼潮模式 policyId %v , timePoint %v",this.PolicyId,this.TimePoint)
		policyData := fishMgr.GetFishByTime(this.PolicyId, this.TimePoint) //通过策略ID  以及  对应得 TimePoint 获取对应得策略数据
		if len(policyData) <= 0 {
			return //  表示该鱼没有对应得策略信息
		}
		this.flushFish() // 进行鱼类刷新
		for _, value := range policyData {
			// 便利当前得 策略数据
			paths := value.GetPaths() // 获取当前的路径信息
			if len(paths) == 0 {
				fishlogger.Info("Policy path data error:", this.PolicyId, this.TimePoint)
				continue
			}
			count := value.GetCount()                              // 获取当前的数量
			for index := int32(0); index < int32(count); index++ { // 进行当前 鱼类得生成
				var instance = this.PolicyId*1000000 + int32(value.GetId())*100 + int32(index+1) // 鱼
				if value.GetTimeToLive() == 0 {
					fishlogger.Errorf("Policy data is error:[%v]-[%v].", this.PolicyId, value.GetId())
				}
				var event = value.GetEvent() // 获取当前鱼类的时间ID
				// 如果是 事件鱼得话  讲事件鱼添加到 对应得map 中
				if event != 0 {
					// 代表当前得鱼是 事件鱼
					eventFishs := this.fish_Event[event]
					if eventFishs == nil {
						eventFishs = []int32{}
					}
					this.fish_Event[event] = append(eventFishs, instance)
				}
				var fishPath = paths[this.Randomer.Rand32(int32(len(paths)))]                                                 // 随机一条鱼的路径
				var birthTick = this.TimePoint + index*value.GetRefreshInterval()                                             // 这个条鱼出生的时间 (当前同类型得鱼的,第几条鱼)
				var liveTick = birthTick + value.GetTimeToLive()*15                                                           // 这条鱼的存活时间，同时也可以理解为这条鱼的死亡事件
				fish := NewFish2(value.GetFishId(), this.fishLevel, instance, fishPath, liveTick, event, birthTick, 0, value) // 根据参数生成Fish对象
				fishlogger.Infof("====NewFishFactory 鱼潮模式 : %v", fish.FishID, fish.TemplateID)
				if fish == nil {
					//fishlogger.Warnf("[NewFish] PolicyId:[%v],Id:[%v],fish:[%v].", this.PolicyId, value.GetId(), value.GetFishId())
				} else {
					//fishlogger.Tracef("[NewFish] PolicyId:[%v],Id:[%v],fish:[%v],fishid:[%v] birthTick:[%v] livetick:[%v].", this.PolicyId, value.GetId(), value.GetFishId(), fish.FishID, birthTick, liveTick)
					this.AddFish(instance, fish)
				}
			}
		}
	} else {
		addFish := []*fishing_proto.SyncRefreshFishInfo{}
		fishCoins := map[int32][]int32{}
		// 当前是普通模式的情况下
		if this.TimePoint == 1 {
			//初始化当前场景的鱼
			fishs := this.InitFishPool(fishCoins)
			for id, num := range fishs {
				info := &fishing_proto.SyncRefreshFishInfo{
					TemplateId: proto.Int32(id),
					Num:        proto.Int32(num),
				}
				addFish = append(addFish, info)
			}
		} else {
			//判断是否更新BOSS类别相关的鱼
			// 1. 判断当前是否更新小boss
			if time.Now().Unix()-this.lastLittleBossTime >= int64(this.GetFishCDTime(8)) {
				//fishlogger.Tracef("当前的桌子ID %v , NewFishFactory 更新 8 类 小boss", this.sceneId)
				this.lastLittleBossTime = time.Now().Unix()
				fishId := this.GetFishTemplateId(8)
				info := &fishing_proto.SyncRefreshFishInfo{
					TemplateId: proto.Int32(fishId),
					Num:        proto.Int32(1),
				}
				if fishCoins[fishId] == nil {
					fishCoins[fishId] = []int32{}
				}
				fishTemplate := base.FishTemplateEx.FishPool[fishId]
				dropCoin := GetFishDropCoin(fishTemplate.RandomCoin)
				fishCoins[fishId] = append(fishCoins[fishId], dropCoin)
				addFish = append(addFish, info)
			}
			// 判断是否更新大boss
			if time.Now().Unix()-this.lastBossTime >= int64(this.GetFishCDTime(7)) && this.BossTag == 0 {
				//fishlogger.Tracef("当前的桌子ID %v ,NewFishFactory 更新 7 类 大boss", this.sceneId)
				this.lastBossTime = time.Now().Unix()
				info := &fishing_proto.SyncRefreshFishInfo{
					TemplateId: proto.Int32(this.BossId),
					Num:        proto.Int32(1),
				}
				this.BossTag = 1
				fishTemplate := base.FishTemplateEx.FishPool[this.BossId]
				dropCoin := GetFishDropCoin(fishTemplate.RandomCoin)
				fishCoins[this.BossId] = append(fishCoins[this.BossId], dropCoin)
				addFish = append(addFish, info)
			}
			// 判断是否更新事件鱼  :  当前鱼池中没有事件鱼的时候 ,随机更新一条事件鱼
			isExistEventFish := 0
			for _, fish := range this.fish_list {
				if fish.FishType == 6 {
					//fishlogger.Tracef("当前的桌子ID %v 当前桌子里面的事件鱼 %v ", this.sceneId, fish.TemplateID)
					isExistEventFish++
				}
			}
			if isExistEventFish == 0 {
				//  更新当前事件鱼
				//fishlogger.Tracef("当前的桌子ID %v NewFishFactory 更新 一条事件鱼", this.sceneId)
				eventFishs := this.UpdateEventFish(1)
				for _, id := range eventFishs {
					info := &fishing_proto.SyncRefreshFishInfo{
						TemplateId: proto.Int32(id),
						Num:        proto.Int32(1),
					}
					fishTemplate := base.FishTemplateEx.FishPool[id]
					dropCoin := GetFishDropCoin(fishTemplate.RandomCoin)
					fishCoins[id] = append(fishCoins[id], dropCoin)
					addFish = append(addFish, info)
				}
			}
			// 判断是否更新普通鱼
			normalFish := this.UpdateNormalFish(fishCoins)
			if len(normalFish) > 0 {
				//fishlogger.Tracef("当前的桌子ID %v ,NewFishFactory 更新 1 ~ 5 类 普通鱼", this.sceneId)
				for fishId, num := range normalFish {
					info := &fishing_proto.SyncRefreshFishInfo{
						TemplateId: proto.Int32(fishId),
						Num:        proto.Int32(num),
					}
					addFish = append(addFish, info)
				}
			}
		}
		if len(addFish) > 0 {
			//fishlogger.Infof("当前的桌子ID %v ,NewFishFactory 有鱼类更新", this.sceneId)
			this.NewFlushFish(addFish)
		}
		// 生成对应鱼类的实例化对象
		for _, i := range addFish {
			policyData := fishMgr.GetFishByFishID(this.PolicyId, i.GetTemplateId())
			if len(policyData) <= 0 {
				continue //  表示该鱼没有对应得策略信息
			}
			for _, value := range policyData {
				for index := int32(0); index < i.GetNum(); index++ {
					dropCoin := fishCoins[i.GetTemplateId()][index]
					if value.GetTimeToLive() == 0 {
						fishlogger.Errorf("Policy data is error:[%v]-[%v].", this.PolicyId, value.GetId())
					}
					var event = value.GetEvent() // 获取当前鱼类的时间ID
					paths := value.GetPaths()    // 获取当前的路径信息
					this.LastID++
					//fishlogger.Infof("当前的桌子ID %v ,当前计数器 the LastID %v",this.sceneId,this.LastID)
					// 如果是 事件鱼得话  讲事件鱼添加到 对应得map 中
					if event != 0 {
						// 代表当前得鱼是 事件鱼
						eventFishs := this.fish_Event[event]
						if eventFishs == nil {
							eventFishs = []int32{}
						}
						this.fish_Event[event] = append(eventFishs, this.LastID)
					}
					var fishPath = paths[this.Randomer.Rand32(int32(len(paths)))]                                                           // 随机一条鱼的路径
					var birthTick = this.TimePoint + index*value.GetRefreshInterval()                                                       // 这个条鱼出生的时间 (当前同类型得鱼的,第几条鱼)
					var liveTick = birthTick + value.GetTimeToLive()*15                                                                     // 这条鱼的存活时间，同时也可以理解为这条鱼的死亡事件
					fish := NewFish2(value.GetFishId(), this.fishLevel, this.LastID, fishPath, liveTick, event, birthTick, dropCoin, value) // 根据参数生成Fish对象
					//fishlogger.Infof("当前的桌子ID %v ====NewFishFactory 普通模式 : %v_%v",this.sceneId,fish.FishID,fish.TemplateID)
					if fish == nil {
						//fishlogger.Warnf("[NewFish] PolicyId:[%v],Id:[%v],fish:[%v].", this.PolicyId, value.GetId(), value.GetFishId())
					} else {
						//fishlogger.Tracef("[NewFish] PolicyId:[%v],Id:[%v],fish:[%v],fishid:[%v] birthTick:[%v] livetick:[%v].", this.PolicyId, value.GetId(), value.GetFishId(), fish.FishID, birthTick, liveTick)
						this.AddFish(this.LastID, fish)
					}
				}
			}
		}
	}
}

// 鱼类工厂生成
func (this *FishingSceneData) fishFactory() {
	if this.TimePoint >= this.MaxTick {
		return
	}
	if this.frozenTick > 0 {
		this.frozenTick--
		return
	}
	this.TimePoint++
	// 获取当前所有鱼的策略信息
	policyData := fishMgr.GetFishByTime(this.PolicyId, this.TimePoint)
	if len(policyData) <= 0 {
		return
	}
	// 刷鱼信息通知
	this.flushFish()
	for _, value := range policyData {
		paths := value.GetPaths() // 获取当前的路径信息
		if len(paths) == 0 {
			fishlogger.Info("Policy path data error:", this.PolicyId, this.TimePoint)
			continue
		}
		count := value.GetCount() // 获取当前的数量
		for index := int32(0); index < int32(count); index++ {
			var instance = this.PolicyId*1000000 + int32(value.GetId())*100 + int32(index+1) // 作为Fish的唯一标识
			if value.GetTimeToLive() == 0 {
				fishlogger.Errorf("Policy data is error:[%v]-[%v].", this.PolicyId, value.GetId())
			}
			var event = value.GetEvent() // 获取当前鱼类的时间ID
			if event != 0 {
				eventFishs := this.fish_Event[event]
				if eventFishs == nil {
					eventFishs = []int32{}
				}
				this.fish_Event[event] = append(eventFishs, instance)
			}
			var fishPath = paths[this.Randomer.Rand32(int32(len(paths)))]                                             // 随机一条鱼的路径
			var birthTick = this.TimePoint + index*value.GetRefreshInterval()                                         // 这个条鱼出生的时间
			var liveTick = birthTick + value.GetTimeToLive()*15                                                       // 这条鱼存货的时间
			fish := NewFish(value.GetFishId(), this.fishLevel, instance, fishPath, liveTick, event, birthTick, value) // 根据参数生成Fish对象
			if fish == nil {
				//fishlogger.Warnf("[NewFish] PolicyId:[%v],Id:[%v],fish:[%v].", this.PolicyId, value.GetId(), value.GetFishId())
			} else {
				//fishlogger.Tracef("[NewFish] PolicyId:[%v],Id:[%v],fish:[%v],fishid:[%v] birthTick:[%v] livetick:[%v].", this.PolicyId, value.GetId(), value.GetFishId(), fish.FishID, birthTick, liveTick)
				this.AddFish(instance, fish)
				if fish.TemplateID == fishing.Fish_CaiShen {
					this.syncFishCoin(fish)
				}
			}
		}
	}
}
func (this *FishingSceneData) fishEvent(fish *Fish, player *FishingPlayerData, power int32, ts int64, extfishs []int32) []*Fish {
	var deathFishs []*Fish
	if fish.Event == 0 {
		deathFishs = append(deathFishs, fish)
		return deathFishs
	}
	if fish.Event == fishing.Event_Rand {
		fish.Event = common.RandInt32Slice([]int32{int32(fishing.Event_Booms), int32(fishing.Event_Boom), int32(fishing.Event_Ring)})
	}
	switch {
	case fish.Event == fishing.Event_Ring:
		{
			deathFishs = append(deathFishs, fish)
			for _, value := range this.fish_list {
				value.LiveTick += 100
				if value.BirthTick >= this.TimePoint {
					value.BirthTick += 100
				}
			}
			this.NextTime += time.Second.Nanoseconds() * 10
			this.frozenTick = 100
			pack := &fishing_proto.SCFreeze{SnId: proto.Int32(player.SnId), FishId: proto.Int32(fish.FishID)}
			this.BroadCastMessage(int(fishing_proto.FIPacketID_FISHING_SC_FREEZE), pack, 0)
		}
	case fish.Event == fishing.Event_Booms:
		{
			fishlogger.Tracef("Event fish %v-%v.", fish.TemplateID, fish.Event)
			fishlogger.Trace("Event ts:", ts)
			sign := fmt.Sprintf("%v;%v;%v;%v;%v;", fish.Event, fish.FishID, this.PolicyId, ts, player.SnId)
			sign = common.MakeMd5String(sign)
			player.fishEvent[sign] = &FishingPlayerEvent{
				FishId:    fish.FishID,
				Event:     fish.Event,
				Power:     power,
				DropCoin:  fish.DropCoin,
				Ts:        ts,
				ExtFishId: extfishs,
			}
			fishlogger.Trace("Event sign:", sign)
			deathFishs = append(deathFishs, fish)
		}
	case fish.Event == fishing.Event_Boom:
		{
			fishlogger.Tracef("Event fish %v-%v.", fish.TemplateID, fish.Event)
			fishlogger.Trace("Event ts:", ts)
			sign := fmt.Sprintf("%v;%v;%v;%v;%v;", fish.Event, fish.FishID, this.PolicyId, ts, player.SnId)
			sign = common.MakeMd5String(sign)
			player.fishEvent[sign] = &FishingPlayerEvent{
				FishId:    fish.FishID,
				Event:     fish.Event,
				Power:     power,
				DropCoin:  fish.DropCoin,
				Ts:        ts,
				ExtFishId: extfishs,
			}
			fishlogger.Trace("Event sign:", sign)
			deathFishs = append(deathFishs, fish)
		}
	case fish.Event == fishing.Event_NewBoom:
		{
			fishlogger.Tracef("Event fish %v-%v.", fish.TemplateID, fish.Event)
			fishlogger.Trace("Event ts:", ts)
			sign := fmt.Sprintf("%v;%v;%v;%v;%v;", fish.Event, fish.FishID, this.PolicyId, ts, player.SnId)
			sign = common.MakeMd5String(sign)
			player.fishEvent[sign] = &FishingPlayerEvent{
				FishId:    fish.FishID,
				Event:     fish.Event,
				Power:     power,
				Ts:        ts,
				ExtFishId: extfishs,
			}
			fishlogger.Trace("Event sign:", sign)
			deathFishs = append(deathFishs, fish)
		}
	case fish.Event == fishing.Event_Bit:
		// 处理钻头鱼相关代码
		{
			fishlogger.Tracef("Event fish %v-%v.", fish.TemplateID, fish.Event)
			fishlogger.Trace("Event ts:", ts)
			sign := fmt.Sprintf("%v;%v;%v;%v;%v;", fish.Event, fish.FishID, this.PolicyId, ts, player.SnId)
			sign = common.MakeMd5String(sign)
			fishlogger.Infof("playerID %v ,sign %v , fishId %v , eventId %v", player.SnId, sign, fish.FishID, fish.Event)
			player.fishEvent[sign] = &FishingPlayerEvent{
				FishId:    fish.FishID,
				Event:     fish.Event,
				Power:     power,
				Ts:        ts,
				ExtFishId: extfishs,
			}
			fishlogger.Trace("Event sign:", sign)
			deathFishs = append(deathFishs, fish)
		}
	case fish.Event == fishing.Event_Lightning:
		{
			fishlogger.Tracef("Event fish %v-%v.", fish.TemplateID, fish.Event)
			fishlogger.Trace("Event ts:", ts)
			sign := fmt.Sprintf("%v;%v;%v;%v;%v;", fish.Event, fish.FishID, this.PolicyId, ts, player.SnId)
			sign = common.MakeMd5String(sign)
			player.fishEvent[sign] = &FishingPlayerEvent{
				FishId:    fish.FishID,
				Event:     fish.Event,
				Power:     power,
				DropCoin:  fish.DropCoin,
				Ts:        ts,
				ExtFishId: extfishs,
			}
			fishlogger.Trace("Event sign:", sign)
			deathFishs = append(deathFishs, fish)
		}
	case fish.Event == fishing.Event_FreePower:
		// 处理免费炮相关逻辑
		{
			fishlogger.Tracef("Event fish %v-%v.", fish.TemplateID, fish.Event)
			fishlogger.Trace("Event ts:", ts)
			sign := fmt.Sprintf("%v;%v;%v;%v;%v;", fish.Event, fish.FishID, this.PolicyId, ts, player.SnId)
			sign = common.MakeMd5String(sign)
			player.fishEvent[sign] = &FishingPlayerEvent{
				FishId:    fish.FishID,
				Event:     fish.Event,
				Power:     power,
				Ts:        ts,
				ExtFishId: extfishs,
			}
			fishlogger.Trace("Event sign:", sign)
			deathFishs = append(deathFishs, fish)
		}
	case fish.Event >= fishing.Event_Same:
		{
			fishlogger.Tracef("Event fish %v-%v.", fish.TemplateID, fish.Event)
			fishlogger.Trace("Event ts:", ts)
			sign := fmt.Sprintf("%v;%v;%v;%v;%v;", fish.Event, fish.FishID, this.PolicyId, ts, player.SnId)
			sign = common.MakeMd5String(sign)
			player.fishEvent[sign] = &FishingPlayerEvent{
				FishId:    fish.FishID,
				Event:     fish.Event,
				Power:     power,
				DropCoin:  fish.DropCoin,
				Ts:        ts,
				ExtFishId: extfishs,
			}
			fishlogger.Trace("Event sign:", sign)
			deathFishs = append(deathFishs, fish)
		}
	default: //一网打尽(同组的鱼)
		{
			fishlogger.Trace("Event fish:", fish)
			eventFishs := this.fish_Event[fish.Event]
			fishlogger.Trace("eventFishs:", eventFishs)
			for _, fishId := range eventFishs {
				fishLink := this.fish_list[fishId]
				if fishLink == nil {
					fishlogger.Trace("Event link fish is null:", fishId)
					continue
				}
				if fishLink.IsDeath(this.TimePoint) {
					fishlogger.Tracef("Event link fish is death:%v-%v", fishLink.FishID, fishId)
					continue
				}
				if !fishLink.IsBirth(this.TimePoint) {
					fishlogger.Tracef("Event link fish is not birth:%v-%v", fishLink.FishID, fishId)
					continue
				}
				fishlogger.Trace("Event link fish:", fishLink)
				fishlogger.Trace("Drop coin:", fishLink.DropCoin*power)
				deathFishs = append(deathFishs, fishLink)
			}
		}
	}
	return deathFishs
}
func (this *FishingSceneData) PushEventFish(player *FishingPlayerData, sign string, fishs []int32, eventFish int32) bool {
	fishEvent := player.fishEvent[sign] // 根据 Sign 确定触发鱼本身的事件
	if fishEvent == nil {
		fishlogger.Error("Recive event fish sign error.")
		fishlogger.Trace("Event sign:", sign)
		return false
	} else {
		fishlogger.Infof("PushEventFish fishEvent %v, %v, sign %v", fishEvent.Event, fishEvent.Ts, sign)
	}
	//接收 Fish 的 相关事件 太晚了
	var timeout time.Duration
	if fishEvent.Event == fishing.Event_Bit {
		timeout = 15
	} else {
		timeout = 10
	}
	if fishEvent.Ts < time.Now().Add(-time.Second*timeout).Unix() {
		fishlogger.Error("Recive event fish list to late.")
		fishlogger.Infof("Event ts: %v", fishEvent.Ts)
		fishlogger.Infof("Event event:%v", fishEvent.Event)
		delete(player.fishEvent, sign) // 消费掉对应的sign
		return false
	}
	fishlogger.Infof("PushEventFish(client): %v", fishs)
	fishlogger.Infof("PushEventFish(srv): %v", fishEvent.ExtFishId)
	var selFishs []*Fish
	// start  获取当前场景 中的 Fish 对象
	for _, id := range fishEvent.ExtFishId { //用碰撞时带上来的id,防作弊
		fish := this.fish_list[id]
		if fish == nil || fish.IsDeath(this.TimePoint) {
			continue
		}
		if fishEvent.Event == fishing.Event_Bit && (fish.FishType == 7 || fish.FishType == 8 || fish.FishType == 6) {
			fishlogger.Infof("钻头贝事件,屏蔽的鱼 %v", fish.TemplateID)
			continue
		}
		selFishs = append(selFishs, fish)
	}
	if fishEvent.Event == fishing.Event_Bit {
		for _, id := range fishs { //用碰撞时带上来的id,防作弊
			fish := this.fish_list[id]
			if fish == nil || fish.IsDeath(this.TimePoint) {
				continue
			}
			if fish.FishType == 7 || fish.FishType == 8 || fish.FishType == 6 {
				fishlogger.Infof("钻头贝事件,屏蔽的鱼 %v", fish.TemplateID)
				continue
			}
			selFishs = append(selFishs, fish)
		}
	}
	// end
	if fishEvent.Event == fishing.Event_Bit {
		if len(fishs) == 0 {
			fishlogger.Errorf(" Event_Bit Event fish die all.")
			delete(player.fishEvent, sign) // 消费掉对应的sign
			this.fishSettlements(selFishs, player, fishEvent.Power, fishEvent.Event, time.Now().UnixNano(), eventFish, 0)
			return false
		}
	} else {
		if len(selFishs) == 0 {
			fishlogger.Errorf("Event fish die all.")
			delete(player.fishEvent, sign) // 消费掉对应的sign
			this.fishSettlements(selFishs, player, fishEvent.Power, fishEvent.Event, time.Now().UnixNano(), eventFish, fishEvent.DropCoin)
			return false
		}
	}

	if fishEvent.Event == fishing.Event_Bit && (len(selFishs) > 1) {
		delete(player.fishEvent, sign) // 消费掉对应的sign
	} else if fishEvent.Event != fishing.Event_Bit {
		delete(player.fishEvent, sign) // 消费掉对应的sign
	}
	// start  根据不同的场景截取 结算的 Fish
	if fishEvent.Event == fishing.Event_Boom && len(selFishs) > 15 {
		selFishs = selFishs[:15]
	}
	if fishEvent.Event == fishing.Event_Lightning && len(selFishs) > 35 {
		selFishs = selFishs[:35]
	}
	if fishEvent.Event == fishing.Event_Booms && len(selFishs) > 30 {
		selFishs = selFishs[:30]
	}
	if fishEvent.Event == fishing.Event_Same && len(selFishs) > 30 {
		selFishs = selFishs[:30]
	}
	// 新事件鱼代码处理相关
	if fishEvent.Event == fishing.Event_Bit && len(selFishs) > 50 {
		selFishs = selFishs[:50]
	}
	if fishEvent.Event == fishing.Event_NewBoom && len(selFishs) > 50 {
		selFishs = selFishs[:50]
	}
	// end
	// 进行 鱼类结算
	if fishEvent.Event == fishing.Event_Bit {
		if len(selFishs) == 1 {
		} else {
			this.fishSettlements(selFishs, player, fishEvent.Power, fishEvent.Event, time.Now().UnixNano(), eventFish, 0)
		}
	} else {
		this.fishSettlements(selFishs, player, fishEvent.Power, fishEvent.Event, time.Now().UnixNano(), eventFish, fishEvent.DropCoin)
	}
	return true
}
func (this *FishingSceneData) BroadCastMessage(packetid int, msg proto.Message, excludeSid int64) {
	mgs := make(map[*netlib.Session][]*srvlibproto.MCSessionUnion)
	for _, p := range this.players {
		if p == nil || p.GetGateSess() == nil {
			continue
		}
		if !p.IsOnLine() || p.IsMarkFlag(base.PlayerState_Leave) {
			continue
		}
		if p.GetSid() == excludeSid {
			continue
		}
		mgs[p.GetGateSess()] = append(mgs[p.GetGateSess()], &srvlibproto.MCSessionUnion{
			Mccs: &srvlibproto.MCClientSession{
				SId: proto.Int64(p.GetSid()),
			},
		})
	}
	audiences := this.GetAudiences()
	for _, p := range audiences {
		if p == nil || p.GetGateSess() == nil {
			continue
		}
		if !p.IsOnLine() || p.IsMarkFlag(base.PlayerState_Leave) {
			continue
		}
		if p.GetSid() == excludeSid {
			continue
		}
		mgs[p.GetGateSess()] = append(mgs[p.GetGateSess()], &srvlibproto.MCSessionUnion{
			Mccs: &srvlibproto.MCClientSession{
				SId: proto.Int64(p.GetSid()),
			},
		})
	}
	for gateSess, v := range mgs {
		if gateSess == nil || len(v) == 0 {
			continue
		}
		pack, err := base.MulticastMaker.CreateMulticastPacket(packetid, msg, v...)
		if err == nil {
			proto.SetDefaults(pack)
			gateSess.Send(int(srvlibproto.SrvlibPacketID_PACKET_SS_MULTICAST), pack)
		}
	}
}

func (this *FishingSceneData) syncFishCoin(fish *Fish) {
	pack := &fishing_proto.SCSyncFishCoin{
		FishId: proto.Int32(fish.FishID),
		Coin:   proto.Int64(int64(fish.DropCoin)),
	}
	proto.SetDefaults(pack)
	this.BroadCastMessage(int(fishing_proto.FIPacketID_FISHING_SC_SCSYNCFISHCOIN), pack, 0)
}

/*
获取当前场此某类鱼的总倍数
*/
func (this *FishingSceneData) GetFishTypeSumGold(fishType int32) (int32, int32) {
	for _, value := range srvdata.PBDB_FishRoomMgr.Datas.Arr {
		if value.GetRoomId() == this.fishLevel {
			if fishType == 1 {
				//1 类鱼
				return ParseFishSumGold(value.GetSumGold1())
			}
			if fishType == 2 {
				//2 类鱼
				return ParseFishSumGold(value.GetSumGold2())
			}
			if fishType == 3 {
				//3 类鱼
				return ParseFishSumGold(value.GetSumGold3())
			}
			if fishType == 4 {
				//4 类鱼
				return ParseFishSumGold(value.GetSumGold4())
			}
			if fishType == 5 {
				return ParseFishSumGold(value.GetSumGold5())
			}
		}
	}
	return -1, -1
}

/*
获取当前场此鱼的cd时间 (目前 支持查询 7类鱼  8类鱼)
*/
func (this *FishingSceneData) GetFishCDTime(fishType int) int32 {
	for _, value := range srvdata.PBDB_FishRoomMgr.Datas.Arr {
		if value.GetRoomId() == this.fishLevel {
			if fishType == 7 {
				//7 类鱼
				return value.GetBossCDTime()
			}
			if fishType == 8 {
				//8 类鱼
				return value.GetLittleBossCDTime()
			}
		}
	}
	return -1
}

/*
解析鱼的总倍数
*/
func ParseFishSumGold(sumGold string) (int32, int32) {
	s := strings.Split(sumGold, ",")
	low, _ := strconv.ParseInt(s[0], 10, 32)
	up, _ := strconv.ParseInt(s[1], 10, 32)
	return int32(low), int32(up)
}

/*
解析鱼的选择权重
*/
func ParseChoseFishWeight(fishWeight string) [][]int32 {
	result := [][]int32{}
	s := strings.Split(fishWeight, ",")
	for _, v := range s {
		r2 := []int32{}
		s2 := strings.Split(v, "_")
		fishID, _ := strconv.ParseInt(s2[0], 10, 32)
		weight, _ := strconv.ParseInt(s2[1], 10, 32)
		r2 = append(r2, int32(fishID))
		r2 = append(r2, int32(weight))
		result = append(result, r2)
	}
	return result
}

/*
解析鱼的金钱掉落选择权重
*/
func ParseChoseFishDropGoldWeight(fishWeight string) [][]int32 {
	result := [][]int32{}
	s := strings.Split(fishWeight, ",")
	for _, v := range s {
		r2 := []int32{}
		s2 := strings.Split(v, "_")
		gold, _ := strconv.ParseInt(s2[0], 10, 32)
		weight, _ := strconv.ParseInt(s2[1], 10, 32)
		r2 = append(r2, int32(gold))
		r2 = append(r2, int32(weight))
		result = append(result, r2)
	}
	return result
}

/*
获取当前鱼本身的权重
*/
func GetFishDropCoin(RandomCoin string) int32 {
	allCoin := ParseChoseFishDropGoldWeight(RandomCoin)
	var totalWeight int32
	for _, value := range allCoin {
		totalWeight = totalWeight + value[1]
	}
	if totalWeight == 1 {
		return allCoin[0][0]
	}
	randValue := common.RandInt(int(totalWeight))
	var targetWeight int32
	for _, param := range allCoin {
		targetWeight = targetWeight + param[1]
		if int32(randValue) < targetWeight {
			return param[0]
		}
	}
	return -1
}

/*
获取鱼的类型ID
*/
func (this *FishingSceneData) GetFishTemplateId(fishType int32) int32 {
	var fishIdWight string
	for _, value := range srvdata.PBDB_FishRoomMgr.Datas.Arr {
		if value.GetRoomId() == this.fishLevel {
			if fishType == 7 {
				//1 类鱼
				fishIdWight = value.GetEnableBoss()
			}
			if fishType == 8 {
				//2 类鱼
				fishIdWight = value.GetEnableLittleBoss()
			}
		}
	}
	fishInfo := ParseChoseFishWeight(fishIdWight)
	var totalWeight int32
	for _, value := range fishInfo {
		totalWeight = totalWeight + value[1]
	}
	randValue := common.RandInt(int(totalWeight))
	var targetWeight int32
	for _, param := range fishInfo {
		targetWeight = targetWeight + param[1]
		if int32(randValue) < targetWeight {
			return param[0]
		}
	}
	return -1
}

/*
	获取当前事件鱼
*/
/*func (this *FishingSceneData) UpdateEventFish(num int) []int32 {
	fishSlice := this.GetTheFishSlice(6)
	copySlice := make([]int32,len(fishSlice))
	copy(copySlice, fishSlice)
	result := []int32{}
	for i := 0; i < num; i++ {
		randValue := common.RandInt(len(copySlice) - 1)
		result = append(result, copySlice[randValue])
		copySlice = append(copySlice[:randValue], copySlice[randValue+1:]...)
	}
	return result
}*/
func (this *FishingSceneData) UpdateEventFish(num int) []int32 {
	fishSlice := this.GetTheFishSlice(6)
	//copySlice := make([]int32,len(fishSlice))
	//copy(copySlice, fishSlice)
	result := []int32{}
	if len(fishSlice) == 0 {
		return result
	}
	for i := 0; i < num; i++ {
		randValue := common.RandInt(len(fishSlice))
		result = append(result, fishSlice[randValue])
	}
	return result
}

/*
获取指定类型鱼的集合 (目前只支持404)
*/
func (this *FishingSceneData) GetTheFishSlice(fishType int32) []int32 {
	result := []int32{}
	policyData := fishMgr.GetPolicyData(this.PolicyId) // 通过策略ID 获取所有Policy信息
	for _, policy := range policyData.Data {
		for _, value := range policy {
			fishTemplate := base.FishTemplateEx.FishPool[value.GetFishId()]
			if fishTemplate != nil && fishTemplate.FishType == fishType {
				result = append(result, value.GetFishId())
			}
		}
	}
	return result
}

/*
更新普通鱼
*/
func (this *FishingSceneData) UpdateNormalFish(fishCoins map[int32][]int32) map[int32]int32 {
	result := make(map[int32]int32)
	// 1. 获取所有类型鱼的种类
	for i := 1; i <= 5; i++ {
		// 2. 获取该场此,该类鱼的上下限
		low, up := this.GetFishTypeSumGold(int32(i))
		// 3. 获取该类鱼目前的金额数目
		var nowSumGold int32
		for _, fish := range this.fish_list {
			if fish.FishType == int32(i) {
				nowSumGold = nowSumGold + fish.DropCoin
			}
		}
		if nowSumGold > up {
			fishlogger.Infof("异常 当前的桌子ID %v UpdateNormalFish fishType %v  low %v  up %v nowSumGold %v", this.GetSceneId(), i, low, up, nowSumGold)
		}
		if nowSumGold < low {
			// 增加该种类的鱼
			for nowSumGold <= up {
				fishSlice := this.GetTheFishSlice(int32(i))
				randValue := common.RandInt(len(fishSlice) - 1)
				fishTemplate := base.FishTemplateEx.FishPool[fishSlice[randValue]]
				dropCoin := GetFishDropCoin(fishTemplate.RandomCoin)
				if len(fishSlice) == 0 {
					break
				}
				if nowSumGold+dropCoin > up {
					break
				}
				if fishCoins[fishSlice[randValue]] == nil {
					fishCoins[fishSlice[randValue]] = []int32{}
				}
				if _, ok := result[fishSlice[randValue]]; ok {
					result[fishSlice[randValue]] = result[fishSlice[randValue]] + 1 // 数字累加
					fishCoins[fishSlice[randValue]] = append(fishCoins[fishSlice[randValue]], dropCoin)
					nowSumGold = nowSumGold + dropCoin
				} else {
					result[fishSlice[randValue]] = 1
					fishCoins[fishSlice[randValue]] = append(fishCoins[fishSlice[randValue]], dropCoin)
					nowSumGold = nowSumGold + dropCoin
				}
			}

		}
	}
	return result
}

/*
初始化当前鱼
@  按照当前鱼种类的最大倍数  只更新 1 ~ 6 种类的鱼  并且  6种类只更新两条
*/
func (this *FishingSceneData) InitFishPool(fishCoins map[int32][]int32) map[int32]int32 {

	result := make(map[int32]int32)
	// 1. 获取 1 ~ 5 类鱼的更新0
	for i := 1; i <= 5; i++ {
		// 2. 获取该场此,该类鱼的上下限
		_, up := this.GetFishTypeSumGold(int32(i))
		// 3. 获取该类鱼目前的金额数目
		fishSlice := this.GetTheFishSlice(int32(i))
		if len(fishSlice) == 0 {
			break
		}
		var nowSumGold int32
		// 增加该种类的鱼
		for nowSumGold <= up {
			randValue := common.RandInt(len(fishSlice) - 1)
			fishTemplate := base.FishTemplateEx.FishPool[fishSlice[randValue]]
			dropCoin := GetFishDropCoin(fishTemplate.RandomCoin)
			if nowSumGold+dropCoin > up {
				break
			}
			if fishCoins[fishSlice[randValue]] == nil {
				fishCoins[fishSlice[randValue]] = []int32{}
			}
			if _, ok := result[fishSlice[randValue]]; ok {
				result[fishSlice[randValue]] = result[fishSlice[randValue]] + 1 // 数字累加
				fishCoins[fishSlice[randValue]] = append(fishCoins[fishSlice[randValue]], dropCoin)
				nowSumGold = nowSumGold + dropCoin
			} else {
				nowSumGold = nowSumGold + dropCoin
				fishCoins[fishSlice[randValue]] = append(fishCoins[fishSlice[randValue]], dropCoin)
				result[fishSlice[randValue]] = 1
			}
		}
	}
	// 2. 获取 6  类鱼的信息
	eventFish := this.UpdateEventFish(2)
	for _, i := range eventFish {
		fishTemplate := base.FishTemplateEx.FishPool[i]
		dropCoin := GetFishDropCoin(fishTemplate.RandomCoin)
		if fishCoins[i] == nil {
			fishCoins[i] = []int32{}
		}
		if _, ok := result[i]; ok {
			fishCoins[i] = append(fishCoins[i], dropCoin)
			result[i]++
		} else {
			fishCoins[i] = append(fishCoins[i], dropCoin)
			result[i] = 1
		}
	}
	return result
}
