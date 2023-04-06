package fishing

import (
	"encoding/json"
	"fmt"
	"games.yol.com/win88/gamesrv/base"
	"math"
	"math/rand"
	"time"

	"github.com/idealeak/goserver/core/timer"

	"games.yol.com/win88/common"
	"games.yol.com/win88/gamerule/fishing"
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	fishing_proto "games.yol.com/win88/protocol/fishing"
	server_proto "games.yol.com/win88/protocol/server"
	"github.com/idealeak/goserver/core/netlib"
	srvlibproto "github.com/idealeak/goserver/srvlib/protocol"
)

//独立一个随机对象
var gHPFishingRand = rand.New(rand.NewSource(time.Now().UnixNano()))

type HPFishingSceneData struct {
	*base.Scene
	players          map[int32]*FishingPlayerData
	seats            [fishing.MaxPlayer]*FishingPlayerData
	BattleBuff       chan *FishBattle
	PolicyArr        []int32
	PolicyArrIndex   int
	PolicyId         int32
	TimePoint        int32
	NextTime         int64
	Policy_Mode      PolicyMode
	MaxTick          int32
	Randomer         *common.RandomGenerator
	fish_list        map[int32]*Fish
	oldfish_list     map[int32]*Fish
	jackpotFish_list map[int32]*Fish // 光圈鱼
	fish_Event       map[int32][]int32
	fish_surplusCoin map[int32][]int32 // 捕鱼鱼没死亡离开的金币放在新鱼身上
	fishLevel        int32
	frozenTick       int32
	lastTick         int64
	remainder        int64
	intervalTime     int64
	iTime            int64
	hDestroy         timer.TimerHandle
	enterTime        time.Time
}

func NewHPFishingSceneData(s *base.Scene) *HPFishingSceneData {
	return &HPFishingSceneData{
		Scene:            s,
		players:          make(map[int32]*FishingPlayerData),
		BattleBuff:       make(chan *FishBattle, 1000),
		fish_list:        make(map[int32]*Fish),
		fish_Event:       make(map[int32][]int32),
		jackpotFish_list: make(map[int32]*Fish),
		fish_surplusCoin: make(map[int32][]int32),
	}
}

func (this *HPFishingSceneData) ChangeFlushFish() {
	//if this.Policy_Mode == Policy_Mode_Normal { // 不刷鱼阵
	fishlogger.Info("ChangeFlushFish.", this.PolicyArr, this.PolicyArrIndex)
	this.fishTimeOut()
	this.NotifySceneStateFishing(common.SceneState_Fishing)
	this.Policy_Mode = Policy_Mode_Normal
	this.PolicyId = this.PolicyArr[this.PolicyArrIndex]
	for _, p := range this.players {
		p.BulletLimit = [BULLETLIMIT]int64{}
	}
	this.PolicyArrIndex++
	if this.PolicyArrIndex >= len(this.PolicyArr) {
		this.PolicyArrIndex = 0
	}

	this.MaxTick = fishMgr.Policy_Data[this.PolicyId].MaxTick
	this.NextTime = time.Now().Add(time.Millisecond * time.Duration(this.MaxTick*100)).UnixNano()
	fishlogger.Infof("Next policyid[%v],maxtick[%v],policy index[%v].", this.PolicyId, this.MaxTick, this.PolicyArrIndex)
}

func (this *HPFishingSceneData) FlushFishOver() bool {
	return time.Now().UnixNano() > this.NextTime
}

func (this *HPFishingSceneData) RebindPlayerSnId(oldSnId, newSnId int32) {
	if p, exist := this.players[oldSnId]; exist {
		delete(this.players, oldSnId)
		this.players[newSnId] = p
	}
}

func (this *HPFishingSceneData) init() bool {
	if this.GetDBGameFree() != nil {
		this.fishLevel = this.GetDBGameFree().GetSceneType()
	}
	this.SetPlayerNum(4)
	this.TimePoint = 0
	this.PolicyId = 0
	//this.Policy_Mode = Policy_Mode_Tide
	this.Policy_Mode = Policy_Mode_Normal
	this.MaxTick = 0
	this.Randomer = &common.RandomGenerator{}

	if FishJackpotCoinMgr.Jackpot[this.GetPlatform()] == nil {
		FishJackpotCoinMgr.Jackpot[this.Platform] = &base.SlotJackpotPool{}

		str := base.SlotsPoolMgr.GetPool(int32(this.GameId), this.Platform) // 三个场次公用一个
		if str != "" {
			jackpot := &base.SlotJackpotPool{}
			err := json.Unmarshal([]byte(str), jackpot)
			if err == nil {
				FishJackpotCoinMgr.Jackpot[this.Platform] = jackpot
			}
		}
		if FishJackpotCoinMgr.Jackpot[this.Platform] != nil {
			if FishJackpotCoinMgr.Jackpot[this.Platform].GetTotalSmall() < 1 { // 初始值
				FishJackpotCoinMgr.Jackpot[this.Platform].AddToSmall(true, model.FishingParamData.JackpotInitCoin)
			} else if FishJackpotCoinMgr.Jackpot[this.Platform].GetTotalMiddle() < 1 { // 初始值
				FishJackpotCoinMgr.Jackpot[this.Platform].AddToMiddle(true, model.FishingParamData.JackpotInitCoin)
			} else if FishJackpotCoinMgr.Jackpot[this.Platform].GetTotalBig() < 1 { // 初始值
				FishJackpotCoinMgr.Jackpot[this.Platform].AddToBig(true, model.FishingParamData.JackpotInitCoin)
			}
			base.SlotsPoolMgr.SetPool(int32(this.GameId), this.Platform, FishJackpotCoinMgr.Jackpot[this.Platform])
		}
		fishlogger.Info("HPFishingSceneData str ", str, FishJackpotCoinMgr.Jackpot[this.Platform])
	}

	this.Randomer.RandomSeed(int32(common.RandInt()))
	this.PolicyArr = InitScenePolicyMode(this.GameId, int(this.DbGameFree.GetSceneType()))
	this.PolicyArrIndex = 0
	this.ChangeFlushFish()
	//随机一个初始点
	start := rand.Int31n(this.MaxTick * 4 / 5)
	for i := int32(0); i < start; i++ {
		this.fishFactory()
	}
	this.NextTime -= int64(time.Millisecond * time.Duration(start*100))
	return true
}
func (this *HPFishingSceneData) Clean() {
	for _, p := range this.players {
		//
		fishID, coin, taxc := int32(0), int32(0), int64(0)
		for _, v := range p.bullet {
			coin += v.Power
			taxc += int64(float64(this.GetDBGameFree().GetTaxRate()) / 10000 * float64(v.Power))
		}
		this.RetBulletCoin(p, fishID, coin, taxc, false) // 合并后发送
		p.bullet = make(map[int32]*Bullet)
		p.BulletLimit = [BULLETLIMIT]int64{}
		p.fishEvent = make(map[string]*FishingPlayerEvent)
		p.bullet = make(map[int32]*Bullet)
		p.logFishCount = make(map[int64]*model.FishCoinNum)

	}
	this.fish_list = make(map[int32]*Fish)
	this.fish_Event = make(map[int32][]int32)
}
func (this *HPFishingSceneData) BroadcastPlayerLeave(p *base.Player, reason int) {
}
func (this *HPFishingSceneData) SceneDestroy(force bool) {
	for _, value := range this.fish_list {
		if value != nil && value.Hp != nil { // 销毁房间返回鱼表对应状态
			key := fmt.Sprintf("%v-%v", value.TemplateID, this.fishLevel)

			if _, exist := FishHPListEx.fishList[key]; exist {
				data := &FishRealHp{
					Id:     value.Hp.Id,
					CurrHp: value.Hp.CurrHp,
				}
				FishHPListEx.fishList[key].CurrTimeOut(data)
				//FishHPListEx.fishList[key].PutFirst(data) // 放在前
			}
		}
	}
	this.Scene.Destroy(force)
}

// RetBulletCoin 返还miss鱼分
func (this *HPFishingSceneData) RetBulletCoin(player *FishingPlayerData, fishID, coin int32, taxc int64, flag bool) {
	////返还能量池
	//player.Prana -= float64(coin) * model.FishingParamData.PranaRatio
	//if player.Prana < 0 {
	//	player.Prana = 0
	//}
	//upperLimit := this.PranaUpperLimit()
	//if int32(player.Prana) > upperLimit { // 满足能量炮发射条件
	//
	//	player.PranaPercent = 100
	//} else {
	//	player.PranaPercent = int32(player.Prana) * 100 / upperLimit
	//}

	if player.IsRob {
		return
	}
	pack := &fishing_proto.SCFishDel{
		FishId:            proto.Int32(int32(fishID)),
		Coin:              proto.Int32(coin),
		Snid:              proto.Int32(player.SnId),
		CurrentPlayerCoin: proto.Int64(player.CoinCache),
	}
	//player.NewStatics(int64(-coin), 0)
	player.LostCoin -= int64(coin)
	player.CoinCache += int64(coin)
	player.SetMaxCoin()
	totalBet := player.GetTotalBet() - int64(coin)
	player.SetTotalBet(totalBet)
	//player.Statics(this.keyGameId, this.gamefreeId, -int64(coin), false)
	player.LastRechargeWinCoin -= int64(coin)
	todata := player.GetTodayGameData(this.KeyGameId)
	todata.TotalIn -= int64(coin)
	if !this.Testing {
		player.taxCoin -= float64(taxc)
		player.sTaxCoin -= float64(taxc)
	}
	fishlogger.Trace("RetBulletCoin : ", fishID, coin, player.LostCoin, player.CoinCache, player.GetTotalBet())
	pack.CurrentPlayerCoin = proto.Int64(player.CoinCache)
	proto.SetDefaults(pack)
	this.BroadCastMessage(int(fishing_proto.FIPacketID_FISHING_SC_FISHDEL), pack, 0)
}

/*
 * 玩家相关
 */

// EnterPlayer 进入房间
func (this *HPFishingSceneData) EnterPlayer(player *FishingPlayerData) bool {
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
	player.MaxCoin = player.CoinCache
	player.MinCoin = player.CoinCache
	player.EnterCoin = player.CoinCache
	player.FishPoolKey = fmt.Sprintf("%v-%v", this.KeyGamefreeId, player.Platform)
	this.players[player.SnId] = player
	this.seats[pos] = player
	this.OnEnterPlayer(player)
	return true
}

// OnEnterPlayer .
func (this *HPFishingSceneData) OnEnterPlayer(player *FishingPlayerData) {
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

// QuitPlayer 离场 统计数据
func (this *HPFishingSceneData) QuitPlayer(player *FishingPlayerData, reason int) bool {
	if _, ok := this.players[player.SnId]; ok {
		player.SaveDetailedLog(this.Scene)
		delete(this.players, player.SnId)
		this.seats[player.GetPos()] = nil
		this.BroadcastPlayerLeave(player.Player, 0)
		diffCoin := player.CoinCache - player.GetCoin()
		player.AddCoin(diffCoin, common.GainWay_Fishing, base.SyncFlag_ToClient, "system", this.GetSceneName())
		if diffCoin != 0 || player.LostCoin != 0 {
			if !player.IsRob && !this.Testing {

				player.AddServiceFee(int64(player.taxCoin))
			}
			player.SetGameTimes(player.GetGameTimes() + 1)
			if diffCoin > 0 {
				player.SetWinTimes(player.GetWinTimes() + 1)
			} else {
				player.SetLostTimes(player.GetLostTimes() + 1)
			}
		}
		key := this.KeyGamefreeId
		var pgd *model.PlayerGameInfo
		if data, exist := player.GDatas[key]; !exist {
			pgd = new(model.PlayerGameInfo)
			player.GDatas[key] = pgd
		} else {
			pgd = data
		}
		if pgd != nil {
			//参数确保
			for i := len(pgd.Data); i < GDATAS_HPFISHING_MAX; i++ {
				pgd.Data = append(pgd.Data, 0)
			}
			pgd.Data[GDATAS_HPFISHING_PRANA] = int64(player.Prana)
		}

		this.OnQuitPlayer(player, reason)
		wpack := &server_proto.GWGameJackCoin{} // 玩家离开一次游戏 同步一次world
		for pl, data := range FishJackpotCoinMgr.Jackpot {
			wpack.Platform = append(wpack.Platform, pl)
			wpack.Coin = append(wpack.Coin, data.GetTotalBig())
			fishlogger.Info("FishJackpotCoin Init ", pl, data)
		}
		proto.SetDefaults(wpack)
		this.SendToWorld(int(server_proto.SSPacketID_PACKET_GW_JACKPOTCOIN), wpack)
		return true
	} else {
		return false
	}
}

// OnQuitPlayer 离场
func (this *HPFishingSceneData) OnQuitPlayer(player *FishingPlayerData, reason int) {
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

// OnTick 0.1s
func (this *HPFishingSceneData) OnTick() {
	if this.TimePoint%10 == 0 {
		this.fishTimeOut()
	}
}

// fishTimeOut 鱼超时离场
func (this *HPFishingSceneData) fishTimeOut() {
	for key, value := range this.fish_list {
		if value == nil {
			fishlogger.Trace("fishTimeOut ", key)
			delete(this.fish_list, key)
			delete(this.jackpotFish_list, key)
			continue
		}
		/*
			天天捕鱼没有状态切换，一张刷鱼表结束后，ChangeFlushFish函数中将TimePoint重置为0，
			但场中存活的鱼BirthTick>TimePoint，LiveTick>TimePoint,所以无法通过value.LiveTick <= this.TimePoint判断这些鱼是否离场，
			需修改判断条件 在上一张刷鱼表存活时间 + 本张表存活时间 >= 这条鱼存活时间
			this.MaxTick - value.BirthTick + this.TimePoint >= value.LiveTick-value.BirthTick
		*/
		if value.LiveTick <= this.TimePoint {
			/*if value.Hp != nil && value.Hp.CurrHp != 0 {
				this.fish_surplusCoin[value.TemplateID] = append(this.fish_surplusCoin[value.TemplateID], value.Hp.CurrHp)
				fishlogger.Trace("fishTimeOut fish_surplusCoin ", value.TemplateID, this.fish_surplusCoin[value.TemplateID])
			}*/

			delete(this.fish_list, key)
			delete(this.jackpotFish_list, key)
			if value.Hp != nil {
				fishlogger.Tracef("fishTimeOut fishId:%d templateId:%d hpId:%d birthTick:%d liveTick:%d curTick:%d", key, value.TemplateID, value.Hp.Id, value.BirthTick, value.LiveTick, this.TimePoint)
				key := fmt.Sprintf("%v-%v", value.TemplateID, this.fishLevel)
				if _, exist := FishHPListEx.fishList[key]; exist {
					data := &FishRealHp{
						Id:     value.Hp.Id,
						CurrHp: value.Hp.CurrHp,
					}
					FishHPListEx.fishList[key].CurrTimeOut(data)
					//FishHPListEx.fishList[key].PutFirst(data) // 放在前
				}
			}
			if value.Event != 0 {
				this.fish_Event[value.Event] = common.DelSliceInt32(this.fish_Event[value.Event], value.FishID)
			}
		}
	}
}

// AddFish 鱼表新增鱼
func (this *HPFishingSceneData) AddFish(id int32, fish *Fish) {
	this.fish_list[id] = fish
	if fish.IsJackpot /*|| common.RandInt(10) < 5*/ {
		if len(this.jackpotFish_list) >= 20 {
			this.jackpotFish_list = make(map[int32]*Fish)
		}
		fishlogger.Trace("AddFish ", len(this.fish_list), len(this.jackpotFish_list))
		this.jackpotFish_list[id] = fish
		this.SendToClientJackpotFish()
	}
}

// SendToClientJackpotFish 光圈鱼
func (this *HPFishingSceneData) SendToClientJackpotFish() {
	if len(this.jackpotFish_list) == 0 {
		return
	}
	pack := &fishing_proto.SCJackpotFish{
		//FishId: proto.Int32(id),
	}
	for id := range this.jackpotFish_list {
		pack.FishIds = append(pack.FishIds, id)
	}
	proto.SetDefaults(pack)
	// fishlogger.Trace("SendToClientJackpotFish ", pack)
	this.BroadCastMessage(int(fishing_proto.FIPacketID_FISHING_SC_JACKPOTFISHDEL), pack, 0)
}

// DelFish 删除列表鱼
func (this *HPFishingSceneData) DelFish(id int32, player *FishingPlayerData) {
	if value, exist := this.fish_list[id]; exist {
		delete(this.fish_list, id)
		delete(this.jackpotFish_list, id)
		if value.Hp != nil {
			fishlogger.Trace("DelFish ", id, value.Hp.Id)
			key := fmt.Sprintf("%v-%v", value.TemplateID, this.fishLevel)
			if _, exist := FishHPListEx.fishList[key]; exist {
				data := &FishRealHp{ // 机器人打死鱼不删除
					Id:     value.Hp.Id,
					CurrHp: value.Hp.CurrHp,
				}
				// 从渔场中清鱼,这里业务层走的是单线?
				if !player.IsRob {
					// 当 玩家是 真人的时候 , 将当前鱼从鱼的队列中删除
					FishHPListEx.fishList[key].CurrDel(data)
				} else {
					FishHPListEx.fishList[key].CurrTimeOut(data) // 将鱼从渔场中清控
				}
				//FishHPListEx.fishList[key].PutEnd(data) // 放在尾
			}
		}
	}
}

/*
 * 捕鱼相关
 */

// fishBattle 碰撞检测 0.1s
func (this *HPFishingSceneData) fishBattle() {
	for i := 0; i < 12; i++ {
		select {
		case data := <-this.BattleBuff:
			player := this.players[data.SnId]
			if player == nil {
				fishlogger.Tracef("Bullet %v owner %v offline.", data.Bullet, data.SnId, data.LockFish)
				continue
			}
			delete(player.bullet, data.Bullet)
			fishlogger.Trace("(this *HPFishingSceneData) fishBattle delete bullet")
			var count = len(data.FishsId)
			if count > 0 && data.Power > 0 {
				this.fishProcess(player, data.FishsId, data.Power, data.ExtFishis, data.LockFish)
			}
		default:
			break
		}
	}
}

// fishProcess 死鱼判断
func (this *HPFishingSceneData) fishProcess(player *FishingPlayerData, fishIds []int, power int32, extfishis []int32, lockFish int32) {
	if len(fishIds) == 0 {
		return
	}
	noRepeate := common.SliceNoRepeate(fishIds)
	var death bool
	var robot = player.IsRob
	var randRate = gHPFishingRand.Intn(FishDrop_Rate) //common.RandInt(FishDrop_Rate)
	var ts = time.Now().Unix()
	var fishes []*Fish
	hasAdd := false //是否增加投入、产出、税收
	var notExistFishes []int32
	player.TestHitNum++
	fishlogger.Tracef("============== TestHitNum:%v  power:%v", player.TestHitNum, power)
	for _, value := range noRepeate {
		var fish = this.fish_list[int32(value)]
		if fish == nil {
			player.TestHitNum--
			fishlogger.Tracef("Be hit fish [%v] is disappear. player.TestHitNum:%v cost:%v len(noRepeate):%v", value, player.TestHitNum, player.TestHitNum*int64(power), len(noRepeate))
			notExistFishes = append(notExistFishes, int32(value))
			//if true /*this.delFish_list[int32(value)] == 1*/ { // 是已经死亡的鱼
			//	taxc := int64(float64(this.GetDBGameFree().GetTaxRate()) / 10000 * float64(power))
			//	this.RetBulletCoin(player, int32(value), power, taxc, true)
			//}
			continue
		}
		fish.SwallowCoin[player.SnId] = fish.SwallowCoin[player.SnId] + power*fish.Ratio/10000
		fishlogger.Tracef("HPSwallowCoin ==> playerId:%v fish:%v SwallowCoin:%v hp:%v power:%v ratio:%v", player.SnId, value, fish.SwallowCoin[player.SnId], fish.Hp.Hp, power, fish.Ratio)
		//子弹命中鱼时统计玩家、系统的投入和产出（不需要考虑是否打死鱼）
		if !hasAdd {
			fishlogger.Tracef("============== len(noRepeate):%v", len(noRepeate))
			//player.NewStatics(int64(power), 0)
			//player.Statics(this.keyGameId, this.gamefreeId, -int64(power), false)
			hasAdd = true
		}

		//if !player.IsRob {
		//	var taxc int64
		//	taxc = int64(float64(this.GetDBGameFree().GetTaxRate()) / 10000 * float64(power))
		//	coinPoolMgr.PushCoin(this.gamefreeId, this.groupId, this.platform, int64(power)-taxc)
		//	coinPoolMgr.tax += float64(taxc)
		//
		//}
		death = false
		if robot /*|| this.testing*/ {
			fish.OnHit(power)
			if int(fish.BaseRate) >= randRate {
				death = true
			}
		} else if this.GetDBGameFree().GetGameId() == int32(common.GameId_TFishing) /*&& this.fishLevel == fishing.ROOM_LV_GAO*/ &&
			!this.Testing { //天天捕鱼
			logkey := player.MakeLogKey(fish.TemplateID, power)
			if v, ok := player.logFishCount[logkey]; ok {
				v.HitNum++
			} else {
				player.logFishCount[logkey] = &model.FishCoinNum{
					HitNum: 1,
					ID:     fish.TemplateID,
					Power:  power,
				}
			}
			player.logBulletHitNums++
			setting := base.GetCoinPoolMgr().GetCoinPoolSetting(this.Platform, this.GetGameFreeId(), this.GetGroupId())
			if setting == nil {
				fishlogger.Error("GetCoinPoolSetting is nil")
				death = false
			}
			fishlogger.Trace("NewFishRateHP ", player.SnId, power, value)
			fish.OnHit(power)
			//key := strconv.Itoa(int(this.GetDBGameFree().GetId()))
			//powers := this.GetDBGameFree().GetOtherIntParams()
			//death = fish.NewFishRateHP(setting.GetCtroRate(), power, this.fishLevel, key, pgs, powers)
			//death = fish.NewFishRateHP2(player, setting.GetCtroRate(), power, this.fishLevel, key, powers)
			ctroRate := setting.GetCtroRate()
			playerOdds, playerRatio := player.GetPlayerOdds(this.KeyGamefreeId, ctroRate, this.fishLevel)
			sysOdds := this.GetSysOdds()
			preCorrect := this.GetPreCorrect()
			rate := float64(0)
			if player.WhiteLevel == 0 && player.WhiteFlag == 0 && player.BlackLevel == 0 {
				rate = fish.NewFishRateHP3(playerOdds, sysOdds, preCorrect, playerRatio, ctroRate, power, player.SnId)
			} else if player.BlackLevel > 0 {
				rate = float64(power) / float64(fish.Hp.Hp) / float64(player.BlackLevel+1)
			} else if player.WhiteLevel+player.WhiteFlag > 0 {
				if player.WBMaxNum > 0 {
					rate = float64(power) / float64(fish.Hp.Hp) * (1 + float64(player.WhiteLevel+player.WhiteFlag)/100)
				} else {
					rate = float64(power)/float64(fish.Hp.Hp) + float64(player.WhiteLevel+player.WhiteFlag)/100
				}
			}

			if rate >= 1 {
				death = true
			} else {
				if int(rate*10000) > randRate {
					death = true
					fishlogger.Tracef("FISH DEATH!!! (%v*10000 > %v)", rate, randRate)
				} else {
					fishlogger.Tracef("FISH NOT DEATH!!! (%v*10000 <= %v)", rate, randRate)
				}
			}

			//预结算判断 若此时系统总投入-系统总产出+该鱼价值≥X，则鱼不会死亡
			if death {
				value := int32(0)
				if fish.Hp != nil {
					value = fish.Hp.Hp
					valueX := model.FishingParamData.LimiteWinCoin1
					if this.fishLevel == 2 {
						valueX = model.FishingParamData.LimiteWinCoin2
					} else if this.fishLevel == 3 {
						valueX = model.FishingParamData.LimiteWinCoin3
					}

					sysTotalIn, sysTotalOut := this.GetSysTotalInAndOut()
					if sysTotalOut-sysTotalIn+int64(value) >= int64(valueX) {
						death = false
						fishlogger.Tracef("SET FISH LIVE!!! (%v-%v+%v>=%v)", sysTotalOut, sysTotalIn, value, valueX)
					}
				}
			}
			//if common.Config.IsDevMode && player.GMLevel == 10 {
			if common.Config.IsDevMode {
				//test code
				death = true
				//test code
				fishlogger.Tracef("GM:% SET FISH:%v DEATH ", player.SnId, fish.FishID)
			}

			this.ShowTraceInfo(player, this.KeyGamefreeId, ctroRate, this.fishLevel, power, player.SnId, fish.FishID, fish.Hp.Hp, preCorrect, rate, playerRatio, death, player.WhiteLevel+player.WhiteFlag, player.BlackLevel, fish.GetPlayerSwallowCoin(player.SnId))
			//fishlogger.Tracef("GameId_TFishing playerSnid = %v fishId = %v death = %v", player.SnId, fish.FishID, death)
		}
		if !death {
			fishes = append(fishes, fish)
			continue
		}
		deathFishs := this.fishEvent(fish, player, power, ts, extfishis)
		if len(deathFishs) == 0 {
			continue
		}
		if this.RequireCoinPool(player) && len(deathFishs) > 1 {
			dropCoin := int32(0)
			for _, v := range deathFishs {
				dropCoin += v.DropCoin
			}
			if int(float64(player.realOdds+1)/float64(dropCoin+1)) < common.RandInt(10000) {
				continue
			}
		}
		if lockFish == 1 {
			for _, value := range deathFishs {
				player.lockFishCount[value.TemplateID]++
			}
		}
		this.fishSettlements(deathFishs, player, power, fish.Event, ts, 0)
	}

	if len(notExistFishes) > 0 && len(notExistFishes) == len(noRepeate) { //事件鱼中某一条不存在时，不返还金币
		taxc := int64(float64(this.GetDBGameFree().GetTaxRate()) / 10000 * float64(power))
		for _, value := range notExistFishes {
			this.RetBulletCoin(player, int32(value), power, taxc, true)
		}
	} else if !player.IsRob { //不需要返还金币时，增加税收、玩家投入、系统投入
		var taxc int64
		taxc = int64(float64(this.GetDBGameFree().GetTaxRate()) / 10000 * float64(power))
		coinPoolMgr := base.GetCoinPoolMgr()
		coinPoolMgr.PushCoin(this.GetGameFreeId(), this.GetGroupId(), this.Platform, int64(power)-taxc)
		coinPoolMgr.SetTax(coinPoolMgr.GetTax() + float64(taxc))

		player.NewStatics(int64(power), 0)
		//player.Statics(this.KeyGameId, this.KeyGamefreeId, -int64(power), false)

		this.AddJackpot(player, power)  //奖池变更
		this.CalcJackpot(player, power) //计算爆奖池
		this.CalcPrana(player, power)   //计算能量炮
	}

	if len(fishes) > 0 {
		this.SyncFishesHp(fishes)
	}
	//写条记录
	if player.logBulletHitNums >= CountSaveNums {
		player.SaveDetailedLog(this.Scene)
	}
}

// firePranaDel 激光炮打鱼
func (this *HPFishingSceneData) firePranaDel(player *FishingPlayerData, fishIds []int, allCoin int32) {
	// allCoin 指的是当前玩家所积蓄的能量
	if len(fishIds) == 0 {
		return
	}
	// 工具链 去重
	noRepeate := common.SliceNoRepeate(fishIds)
	var fishId []int32
	var fishCoin []int32
	var fishes []*Fish
	var ts = time.Now().Unix()
	dropCoin := int32(0)
	for _, value := range noRepeate {
		// 从所有鱼的队列中,获取对应的Fish对象
		var fish = this.fish_list[int32(value)]
		if fish == nil {
			fishlogger.Tracef("firePranaDel Be hit fish [%v] is disappear. %v", value)
			continue
		}
		fishes = append(fishes, fish)
		// start  记录鱼的次数
		key := player.MakeLogKey(fish.TemplateID, 0)
		if v, ok := player.logFishCount[key]; ok {
			// 增加对应的鱼的击中次数
			v.HitNum++
		} else {
			// 初始化
			player.logFishCount[key] = &model.FishCoinNum{
				HitNum: 1,
				ID:     fish.TemplateID,
				Power:  0, //特殊炮
			}
		}
		// end
		// 鱼的对应死亡事件
		deathFishs := this.fishEvent(fish, player, 0, ts, []int32{})
		if len(deathFishs) == 0 {
			continue
		}

		if len(deathFishs) >= 1 {
			dcoin := int32(0)
			for _, v := range deathFishs {
				if v.Hp != nil {
					dcoin += v.Hp.Hp //总生命值就是它所掉落的金币
					fishId = append(fishId, v.FishID)
					fishCoin = append(fishCoin, v.Hp.Hp)
				}
			}
			// 如果掉落的金币数值 > 能量炮的数值 不做处理
			value := allCoin
			setting := base.GetCoinPoolMgr().GetCoinPoolSetting(this.Platform, this.GetGameFreeId(), this.GetGroupId())
			if setting != nil {
				value = this.GetPranaValue(player, this.KeyGamefreeId, setting.GetCtroRate(), this.fishLevel, allCoin)
			}
			if dropCoin+dcoin > value {
				continue
			}
			dropCoin += dcoin
			// 会进行对应的元素设置
			this.fishSettlements(deathFishs, player, 0, fish.Event, ts, 0)
		}
	}
	fishlogger.Tracef("firePranaDel Be hit fishs:%v coins:%v allCoin:%v", fishId, fishCoin, allCoin)
	if dropCoin != 0 { // 通知客户端显示能量炮金币
		pack := &fishing_proto.SCRetPranaCoin{
			SnId: proto.Int32(player.SnId),
			Coin: proto.Int64(int64(dropCoin)),
		}
		proto.SetDefaults(pack)
		fishlogger.Trace("Send SCRetPranaCoin pb", pack)
		// 广播当前场景的所有玩家
		this.BroadCastMessage(int(fishing_proto.FIPacketID_FISHING_SC_RETPRANACOIN), pack, 0)
	}
	this.SyncFishesHp(fishes)
}

// fishSettlements .
func (this *HPFishingSceneData) fishSettlements(fishs []*Fish, player *FishingPlayerData, power int32, event int32,
	ts int64, eventFishId int32) {
	var coin int64
	pack := &fishing_proto.SCFireHit{
		Snid:              proto.Int32(player.SnId),
		Ts:                proto.Int64(ts),
		EventFish:         proto.Int32(eventFishId),
		CurrentPlayerCoin: proto.Int64(player.CoinCache),
		Power:             proto.Int32(power),
	}
	if event != 0 {
		pack.Event = proto.Int32(event)
	}
	for _, value := range fishs {
		// power  可以理解为  与本身掉落金币的权重
		dropCoin := value.DropCoin * power
		if value.Hp != nil {
			dropCoin = value.Hp.Hp
			if value.IsJackpot { // 奖金鱼 (特殊鱼种的计算)
				rd := int32(common.RandInt(11, 16))
				dropCoin = value.Hp.Hp * rd / 10
				fishlogger.Trace("fishSettlements IsJackpotFish ", rd, dropCoin, value.Hp.Hp)
			}
		}
		pack.FishId = append(pack.FishId, value.FishID)
		pack.Coin = append(pack.Coin, dropCoin)
		// start 统计某种鱼本身的击中次数,以及 掉落的总金币数
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
		// end
		coin = coin + int64(dropCoin)
		// 一些 特殊 模板鱼的处理 , 当命中该鱼的时候 推送事件广播
		if value.TemplateID == 21061 || value.TemplateID == 21062 ||
			value.TemplateID == 21063 || value.TemplateID == 21064 {
			fishTemplate := base.FishTemplateEx.FishPool[value.TemplateID]
			if fishTemplate != nil {
				this.NewGoldFishNotice(player.Player, dropCoin, fishTemplate.Name, 11)
			}
		}
		this.DelFish(value.FishID, player) // 鱼死了 从鱼池中 清鱼
		value.SetDeath()                   // 设置当前鱼 死亡
	}
	// TODO 暂时不清楚 整个 func是用来 做什么的 整个逻辑梳理 清晰后 回过头 再看 Player 中的逻辑
	player.NewStatics(0, coin)
	player.winCoin += coin
	player.CoinCache += coin // 设置金币缓存
	player.SetMaxCoin()      // 缓存更新
	pack.CurrentPlayerCoin = proto.Int64(player.CoinCache)
	proto.SetDefaults(pack)
	this.BroadCastMessage(int(fishing_proto.FIPacketID_FISHING_SC_FIREHIT), pack, 0)
	//player.Statics(this.KeyGameId, this.KeyGamefreeId, coin, false)
	todata := player.GetTodayGameData(this.KeyGameId)
	todata.TotalOut += coin
	if !player.IsRob /*&& !this.testing*/ {
		// 宏观调控整个水池设计
		base.GetCoinPoolMgr().PopCoin(this.GetGameFreeId(), this.GetGroupId(), this.Platform, int64(coin))
	}
}

// PushBattle 碰撞
func (this *HPFishingSceneData) PushBattle(player *FishingPlayerData, bulletid int32, lockFish int32, fishs []int32, extfishis []int32) {
	bullet := player.bullet[bulletid]
	if bullet == nil {
		fishlogger.Infof("%v bulletSaveGameDetailedLog not find in %v bullet buff. %v", bulletid, player.GetName(), len(player.bullet))
		return
	}
	//fishlogger.Info("PushBattle id ", bulletid, player.SnId)
	battleData := &FishBattle{
		SnId:      bullet.SrcId,
		Bullet:    bullet.Id,
		Power:     bullet.Power,
		ExtFishis: extfishis,
		LockFish:  lockFish,
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
			taxc := int64(float64(this.GetDBGameFree().GetTaxRate()) / 10000 * float64(battleData.Power))
			this.RetBulletCoin(player, int32(battleData.FishsId[0]), battleData.Power, taxc, true)
		}
	}
}

// PushBullet 发炮
func (this *HPFishingSceneData) PushBullet(s *base.Scene, snid, x, y, id, power int32) fishing_proto.OpResultCode {
	player := this.players[snid]
	if player == nil {
		fishlogger.Errorf("player %v is empty,bullet will be droped.", snid)
		return fishing_proto.OpResultCode_OPRC_Error
	}
	if power <= 0 {
		fishlogger.Tracef("[%v] power is zero.", player.GetName())
		return fishing_proto.OpResultCode_OPRC_Error
	}
	if !player.CoinCheck(power) {
		fishlogger.Tracef("%v no enough coin to fishing.", player.GetName())
		return fishing_proto.OpResultCode_OPRC_CoinNotEnough
	}
	if player.IsRob && (int64(power*5) > player.CoinCache) {
		fishlogger.Tracef("rob %v no enough coin to fishing.", player.GetName())
		return fishing_proto.OpResultCode_OPRC_CoinNotEnough
	}
	curTime := time.Now().Unix() % BULLETLIMIT
	sbl := player.BulletLimit[curTime]
	bulletCountLimit := int64(5)
	if player.FireRate > 0 {
		bulletCountLimit = 10
	}
	// fishlogger.Tracef("PushBullet id %v  %v   %v %v   %v %v", id, len(player.bullet), player.BulletLimit[curTime], curTime, snid, player.GetName())
	if sbl > bulletCountLimit {
		fishlogger.Infof("Player bullet too fast.")
		return fishing_proto.OpResultCode_OPRC_Error
	} else {
		player.BulletLimit[curTime] = sbl + 1
	}
	player.bullet[id] = &Bullet{
		Id:       id,
		Power:    power,
		SrcId:    player.SnId,
		LifeTime: 0,
	}
	//fishlogger.Info("GameId_TFishing PushBullet id ", id, player.SnId)
	//player.NewStatics(int64(power), 0)
	player.LostCoin += int64(power)
	player.CoinCache -= int64(power)
	player.SetMinCoin()
	player.SetTotalBet(player.GetTotalBet() + int64(power))
	//player.Statics(this.keyGameId, this.gamefreeId, -int64(power), false)
	player.LastRechargeWinCoin += int64(power)
	//player.jackpotCoin += float64(power) * model.FishingParamData.JackpotRate // 临时设置
	//jcoin := int64(player.jackpotCoin)
	//if jcoin > 0 { // 不够就累加
	//	if _, exist := FishJackpotCoinMgr.Jackpot[this.platform]; exist {
	//		//Jackpot.AddToBig(player.IsRob, jcoin)
	//		this.AddToJackpot(player.IsRob, jcoin, this.fishLevel)
	//		pack := &fishing_proto.SCJackpotPool{
	//			//Coin: proto.Int64(Jackpot.GetTotalBig()),
	//			Coin: proto.Int64(this.GetJackpot(0)),
	//		}
	//		proto.SetDefaults(pack)
	//		// 奖池
	//		this.BroadCastMessageAllPlayers(int(fishing_proto.FIPacketID_FISHING_SC_JACKPOTPOOLCHANGE), pack, 0)
	//		player.jackpotCoin -= float64(jcoin)
	//		// fishlogger.Info("AddToBig ", model.FishingParamData.JackpotRate, player.jackpotCoin, pack)
	//	}
	//}
	if !player.IsRob /*&& !this.testing*/ {
		var taxc int64
		taxc = int64(float64(this.GetDBGameFree().GetTaxRate()) / 10000 * float64(power))
		if !this.Testing {
			player.taxCoin += float64(taxc)
			player.sTaxCoin += float64(taxc)
		}
	}
	todata := player.GetTodayGameData(this.KeyGameId)
	todata.TotalIn += int64(power)

	//key := fmt.Sprintf("%v_%v", this.platform, this.GetDBGameFree().GetId())
	//if !player.IsRob && !this.testing {
	//	if Jackpot, exist := FishJackpotCoinMgr.Jackpot[this.platform]; exist {
	//		setting := coinPoolMgr.GetCoinPoolSetting(this.platform, this.gamefreeId, this.groupId)
	//		if setting != nil {
	//			ctroRate := setting.GetCtroRate()
	//			playerOdds, _ := player.GetPlayerOdds(this.keyGameId, ctroRate, this.fishLevel)
	//			jackCoin, jacktype := CalcuJackpotCoin(power, player.GMLevel, key, this.fishLevel, playerOdds)
	//			if jackCoin != 0 && this.GetJackpot(this.fishLevel)-jackCoin > 0 { // 爆奖前提条件：爆奖金额＜奖池当前金额-奖池初始金额
	//				// TODO 爆奖 -> insert dblog    updata coin jacklist
	//				player.CoinCache += jackCoin
	//				player.ExtraCoin += jackCoin
	//				player.SetMaxCoin()
	//				pack := &fishing_proto.SCJackpotCoin{
	//					SnId:         proto.Int32(player.SnId),
	//					Coin:         proto.Int32(int32(jackCoin)),
	//					JackpotLevel: proto.Int32(int32(jacktype)),
	//					Name:         proto.String(player.Name),
	//				}
	//				//Jackpot.AddToBig(player.IsRob, -jackCoin)
	//				this.AddToJackpot(player.IsRob, -jackCoin, this.fishLevel)
	//				player.NewStatics(0, jackCoin)                                   //系统产出
	//				player.Statics(this.keyGameId, this.gamefreeId, jackCoin, false) //个人产出
	//				todata := player.GetTodayGameData(this.keyGameId)
	//				todata.TotalOut += jackCoin
	//				proto.SetDefaults(pack)
	//				fishlogger.Info("JackpotCoin ", pack, Jackpot.GetTotalBig())
	//				//爆奖池
	//				this.BroadCastMessageAllPlayers(int(fishing_proto.FIPacketID_FISHING_SC_JACKPOTCOIN), pack, 0)
	//				wpack := &fishing_proto.GWGameJackList{
	//					SnId:     proto.Int32(player.SnId),
	//					Coin:     proto.Int64(jackCoin),
	//					RoomId:   proto.Int32(this.gamefreeId),
	//					GameId:   proto.Int32(this.GetDBGameFree().GetGameId()),
	//					JackType: proto.Int32(jacktype),
	//					Platform: proto.String(player.Platform),
	//					Channel:  proto.String(player.Channel),
	//					Name:     proto.String(player.Name),
	//				}
	//				proto.SetDefaults(wpack)
	//				this.SendToWorld(int(fishing_proto.SSPacketID_PACKET_GW_JACKPOTLIST), wpack)
	//				this.NewFishJackaptNotice(player.Player, jackCoin, jacktype, 10)
	//			}
	//		}
	//	}
	//}
	//this.CalcPrana(player, power) // 计算能量炮

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
	//if player.GameData[this.keyGameId].GameTimes%15 == 0 {
	//	curCoin := coinPoolMgr.LoadCoin(this.gamefreeId, this.platform, this.groupId)
	//	player.SaveFishingLog(curCoin, this.keyGameId)
	//}
	return fishing_proto.OpResultCode_OPRC_Sucess
}

// CalcPrana 计算能量炮
func (this *HPFishingSceneData) CalcPrana(player *FishingPlayerData, power int32) {
	opa := player.PranaPercent
	player.Prana += float64(power) * model.FishingParamData.PranaRatio
	if player.PranaPercent == 100 {
		return
	}
	upperLimit := this.PranaUpperLimit()
	if int32(player.Prana) > upperLimit { // 满足能量炮发射条件

		player.PranaPercent = 100
	} else {
		player.PranaPercent = int32(player.Prana) * 100 / upperLimit
	}

	if player.PranaPercent != opa {
		// 通知客户端能量炮变化
		pack := &fishing_proto.SCSendReadyPrana{
			SnId:    proto.Int32(player.SnId),
			Percent: proto.Int32(player.PranaPercent),
		}
		proto.SetDefaults(pack)
		this.BroadCastMessage(int(fishing_proto.FIPacketID_FISHING_SC_PRANACHANGE), pack, 0)
	}
}

// PranaUpperLimit .
func (this *HPFishingSceneData) PranaUpperLimit() int32 {
	var ret int32
	switch this.fishLevel {
	case fishing.ROOM_LV_CHU:
		ret = model.FishingParamData.PranaECoin
	case fishing.ROOM_LV_ZHO:
		ret = model.FishingParamData.PranaMCoin
	case fishing.ROOM_LV_GAO:
		ret = model.FishingParamData.PranaHCoin
	default:
		ret = model.FishingParamData.PranaECoin
	}
	return ret
}

// SyncFish 同步鱼状态
func (this *HPFishingSceneData) SyncFish(player *base.Player) {
	var cnt int
	var fishes []*fishing_proto.FishInfo
	for _, fish := range this.fish_list {
		if fish.IsDeath(this.TimePoint) {
			continue
		}

		if fish.BirthTick < this.TimePoint && fish.LiveTick < this.TimePoint {
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
			HpRatio:   proto.Int32(fish.Hp.CurrHp * 100 / fish.Hp.Hp),
		})
		cnt++
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
	this.SendToClientJackpotFish()
}

// flushFish 向client发送随机种子
func (this *HPFishingSceneData) flushFish() {
	pack := &fishing_proto.SCSyncRefreshFish{
		PolicyId:   proto.Int32(this.PolicyId),
		TimePoint:  proto.Int32(this.TimePoint % this.MaxTick),
		RandomSeed: proto.Int32(this.Randomer.GetRandomSeed()),
	}
	proto.SetDefaults(pack)
	this.BroadCastMessage(int(fishing_proto.FIPacketID_FISHING_SC_SYNCFISH), pack, 0)
}

// fishFactory 创建鱼
func (this *HPFishingSceneData) fishFactory() {
	tp := this.TimePoint % this.MaxTick
	if tp >= this.MaxTick {
		return
	}
	if this.frozenTick > 0 {
		this.frozenTick--
		return
	}
	this.TimePoint++
	tp++
	policyData := fishMgr.GetFishByTime(this.PolicyId, tp)
	if len(policyData) <= 0 {
		return
	}
	this.flushFish()
	for _, value := range policyData {
		paths := value.GetPaths()
		if len(paths) == 0 {
			fishlogger.Info("Policy path data error:", this.PolicyId, tp)
			continue
		}
		count := value.GetCount()
		for index := int32(0); index < count; index++ {
			var instance = this.PolicyId*1000000 + value.GetId()*100 + index + 1
			if value.GetTimeToLive() == 0 {
				fishlogger.Errorf("Policy data is error:[%v]-[%v].", this.PolicyId, value.GetId())
			}
			var event = value.GetEvent()
			if event != 0 {
				eventFishs := this.fish_Event[event]
				if eventFishs == nil {
					eventFishs = []int32{}
				}
				this.fish_Event[event] = append(eventFishs, instance)
			}
			var fishPath = paths[this.Randomer.Rand32(int32(len(paths)))]
			var birthTick = this.TimePoint + index*value.GetRefreshInterval()
			var liveTick = birthTick + value.GetTimeToLive()*15
			fish := NewFish(value.GetFishId(), this.fishLevel, instance, fishPath, liveTick, event, birthTick, value)
			/*if v, exist := this.fish_surplusCoin[fish.TemplateID]; exist && fish.Hp != nil {
				if len(v) != 0 {
					fish.Hp.CurrHp = v[0]
					this.fish_surplusCoin[fish.TemplateID] = common.DelSliceInt32(this.fish_surplusCoin[fish.TemplateID], v[0])
				}
				fishlogger.Info("fishFactory fish_surplusCoin: ", fish.FishID, fish.TemplateID, fish.Hp.CurrHp)
			}*/
			if fish != nil && fish.Hp != nil {
				key := fmt.Sprintf("%v-%v", fish.TemplateID, this.fishLevel)
				if _, exist := FishHPListEx.fishList[key]; exist {
					if v, ok := FishHPListEx.fishList[key].CurrPop(); /*Pop()*/ ok {
						//fish.Hp.CurrHp = v.CurrHp
						fish.Hp.RateHp = v.RateHp
						fish.Hp.Id = v.Id
						if fish.IsJackpot {
							fish.Hp.RateHp = fish.Hp.RateHp * 12 / 10
						}

						fishlogger.Info("New fish is coming ", fish.FishID, fish.TemplateID, fish.Hp.CurrHp, fish.Hp.RateHp, fish.Hp.Id, fish.IsJackpot, this.fishLevel)
						/*data := &FishRealHp{
							CurrHp: 0,
							RateHp: v.RateHp,
						}
						FishHPListEx.fishList[key].PutEnd(data) // 放在尾
						*/
					} else {
						fish.Hp.RateHp = -1
					}
				} else {
					fish.Hp.RateHp = -1
				}
			}
			if fish == nil {
				fishlogger.Warnf("PolicyId:[%v],Id:[%v],fish:[%v].", this.PolicyId, value.GetId(), value.GetFishId())
			} else {
				this.AddFish(instance, fish)
			}
		}
	}
}

// fishEvent .
func (this *HPFishingSceneData) fishEvent(fish *Fish, player *FishingPlayerData, power int32, ts int64, extfishs []int32) []*Fish {
	var deathFishs []*Fish
	if fish.Event == 0 {
		deathFishs = append(deathFishs, fish)
		return deathFishs
	}
	if fish.Event == fishing.Event_Rand {
		fish.Event = common.RandInt32Slice([]int32{int32(fishing.Event_Booms), int32(fishing.Event_Boom), int32(fishing.Event_Ring)})
	}
	// 初始化当前的DropCoin
	// start  将所有额外鱼的 dropCoin 进行累加
	dropcoin := int32(0)
	for _, v := range extfishs {
		var extfish = this.fish_list[int32(v)]
		if extfish == nil {
			continue
		}
		dropcoin += extfish.DropCoin
	}
	// end
	// 根据 该fish 所处的事件 , 进行分类划分
	switch {
	case fish.Event == fishing.Event_Ring:
		// 当前触发的是 冰冻事件
		{
			deathFishs = append(deathFishs, fish)
			// start  当前场景下的每一条fish LiveTick累加100 ,并且如果Fish的BirthTick大于场景的TimePoint,那么 Fish的 BirthTick 累加 100
			for _, value := range this.fish_list {
				value.LiveTick += 100
				if value.BirthTick >= this.TimePoint {
					value.BirthTick += 100
				}
			}
			// end
			// start  场景更新  NexTime  自增  10Nano ,并且 frozenTick 设置成为 100
			this.NextTime += time.Second.Nanoseconds() * 10
			this.frozenTick = 100
			// end
			pack := &fishing_proto.SCFreeze{SnId: proto.Int32(player.SnId), FishId: proto.Int32(fish.FishID)}
			// 将冻僵的鱼推广到，当前场景下的所有玩家
			this.BroadCastMessage(int(fishing_proto.FIPacketID_FISHING_SC_FREEZE), pack, 0)
		}
	case fish.Event == fishing.Event_Booms:
		// 当前出发的是全屏炸弹事件
		// 判断当前是否需要 投币池 (主要逻辑是看当前玩家是否是 Robot ，如果是 则  false  ，反之 true) ,核心处理逻辑是一样的
		if this.RequireCoinPool(player) {
			// 需要检测水池
			// +1 是数值强制纠错，避免分母为0
			if int(float64(player.realOdds+1)/float64((dropcoin+fish.DropCoin+1))) > common.RandInt(10000) {
				fishlogger.Tracef("Event fish %v-%v.", fish.TemplateID, fish.Event)
				fishlogger.Trace("Event ts:", ts)
				sign := fmt.Sprintf("%v;%v;%v;%v;%v;", fish.Event, fish.FishID, this.PolicyId, ts, player.SnId)
				sign = common.MakeMd5String(sign)
				// 将当前Fish触发事件进行封装,存储到 player.fishEvent中
				player.fishEvent[sign] = &FishingPlayerEvent{
					FishId: fish.FishID,
					Event:  fish.Event,
					Power:  power,
					Ts:     ts,
				}
				fishlogger.Trace("Event sign:", sign)
				deathFishs = append(deathFishs, fish)
			}
		} else {
			// 不需要检测水池,处理逻辑与需要检测水池的处理逻辑相同
			fishlogger.Tracef("Event fish %v-%v.", fish.TemplateID, fish.Event)
			fishlogger.Trace("Event ts:", ts)
			sign := fmt.Sprintf("%v;%v;%v;%v;%v;", fish.Event, fish.FishID, this.PolicyId, ts, player.SnId)
			sign = common.MakeMd5String(sign)
			player.fishEvent[sign] = &FishingPlayerEvent{
				FishId: fish.FishID,
				Event:  fish.Event,
				Power:  power,
				Ts:     ts,
			}
			fishlogger.Trace("Event sign:", sign)
			deathFishs = append(deathFishs, fish)
		}
	case fish.Event == fishing.Event_Boom:
		// 当前的事件是局部炸弹 (处理逻辑同上)
		// 判断当前是否需要 投币池 (主要逻辑是看当前玩家是否是 Robot ，如果是 则  false  ，反之 true) ,核心处理逻辑是一样的
		if this.RequireCoinPool(player) {
			if int(float64(player.realOdds+1)/float64((dropcoin+fish.DropCoin+1))) > common.RandInt(10000) {
				fishlogger.Tracef("Event fish %v-%v.", fish.TemplateID, fish.Event)
				fishlogger.Trace("Event ts:", ts)
				sign := fmt.Sprintf("%v;%v;%v;%v;%v;", fish.Event, fish.FishID, this.PolicyId, ts, player.SnId)
				sign = common.MakeMd5String(sign)
				player.fishEvent[sign] = &FishingPlayerEvent{
					FishId: fish.FishID,
					Event:  fish.Event,
					Power:  power,
					Ts:     ts,
				}
				fishlogger.Trace("Event sign:", sign)
				deathFishs = append(deathFishs, fish)
			}
		} else {
			fishlogger.Tracef("Event fish %v-%v.", fish.TemplateID, fish.Event)
			fishlogger.Trace("Event ts:", ts)
			sign := fmt.Sprintf("%v;%v;%v;%v;%v;", fish.Event, fish.FishID, this.PolicyId, ts, player.SnId)
			sign = common.MakeMd5String(sign)
			player.fishEvent[sign] = &FishingPlayerEvent{
				FishId: fish.FishID,
				Event:  fish.Event,
				Power:  power,
				Ts:     ts,
			}
			fishlogger.Trace("Event sign:", sign)
			deathFishs = append(deathFishs, fish)
		}
	case fish.Event == fishing.Event_Lightning:
		// 当前的事件是连锁闪电 (处理逻辑同上)
		// 判断当前是否需要 投币池 (主要逻辑是看当前玩家是否是 Robot ，如果是 则  false  ，反之 true) ,核心处理逻辑是一样的
		if this.RequireCoinPool(player) {
			if int(float64(player.realOdds+1)/float64((dropcoin+fish.DropCoin+1))) > common.RandInt(10000) {
				fishlogger.Tracef("Event fish %v-%v.", fish.TemplateID, fish.Event)
				fishlogger.Trace("Event ts:", ts)
				sign := fmt.Sprintf("%v;%v;%v;%v;%v;", fish.Event, fish.FishID, this.PolicyId, ts, player.SnId)
				sign = common.MakeMd5String(sign)
				player.fishEvent[sign] = &FishingPlayerEvent{
					FishId: fish.FishID,
					Event:  fish.Event,
					Power:  power,
					Ts:     ts,
				}
				fishlogger.Trace("Event sign:", sign)
				deathFishs = append(deathFishs, fish)
			}
		} else {
			fishlogger.Tracef("Event fish %v-%v.", fish.TemplateID, fish.Event)
			fishlogger.Trace("Event ts:", ts)
			sign := fmt.Sprintf("%v;%v;%v;%v;%v;", fish.Event, fish.FishID, this.PolicyId, ts, player.SnId)
			sign = common.MakeMd5String(sign)
			player.fishEvent[sign] = &FishingPlayerEvent{
				FishId: fish.FishID,
				Event:  fish.Event,
				Power:  power,
				Ts:     ts,
			}
			fishlogger.Trace("Event sign:", sign)
			deathFishs = append(deathFishs, fish)
		}
	case fish.Event >= fishing.Event_Same:
		//  玩家事件  是  托盘鱼 连锁闪电 和  随机
		{
			if this.RequireCoinPool(player) {
				if int(float64(player.realOdds+1)/float64((dropcoin+fish.DropCoin+1))) > common.RandInt(10000) {
					fishlogger.Tracef("Event fish %v-%v.", fish.TemplateID, fish.Event)
					fishlogger.Trace("Event ts:", ts)
					sign := fmt.Sprintf("%v;%v;%v;%v;%v;", fish.Event, fish.FishID, this.PolicyId, ts, player.SnId)
					sign = common.MakeMd5String(sign)
					player.fishEvent[sign] = &FishingPlayerEvent{
						FishId: fish.FishID,
						Event:  fish.Event,
						Power:  power,
						Ts:     ts,
					}
					fishlogger.Trace("Event sign:", sign)
					deathFishs = append(deathFishs, fish)
				}
			} else {
				fishlogger.Tracef("Event fish %v-%v.", fish.TemplateID, fish.Event)
				fishlogger.Trace("Event ts:", ts)
				sign := fmt.Sprintf("%v;%v;%v;%v;%v;", fish.Event, fish.FishID, this.PolicyId, ts, player.SnId)
				sign = common.MakeMd5String(sign)
				player.fishEvent[sign] = &FishingPlayerEvent{
					FishId: fish.FishID,
					Event:  fish.Event,
					Power:  power,
					Ts:     ts,
				}
				fishlogger.Trace("Event sign:", sign)
				deathFishs = append(deathFishs, fish)
			}
		}
	default:
		{ // start   默认场景 是从 场景中的Fish的事件队列中取出与当前Fish相同事件的所有鱼
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
				// end
			}
		}
	}
	return deathFishs
}

/*
func (this *HPFishingSceneData) PushEventFish(player *FishingPlayerData, sign string, fishs []int32, eventFish int32) bool {
	fishEvent := player.fishEvent[sign]
	if fishEvent == nil {
		fishlogger.Error("Recive event fish sign error.")
		fishlogger.Trace("Event sign:", sign)
		return false
	}
	if fishEvent.Ts < time.Now().Add(-time.Second*10).Unix() {
		fishlogger.Error("Recive event fish list to late.")
		fishlogger.Trace("Event ts:", fishEvent.Ts)
		fishlogger.Trace("Event event:", fishEvent.Event)
		return false
	}
	fishlogger.Trace("PushEventFish:", fishs)
	var selFishs []*Fish
	for _, id := range fishs {
		fish := this.fish_list[id]
		if fish == nil || fish.IsDeath(this.TimePoint) {
			continue
		}
		selFishs = append(selFishs, fish)
	}
	if len(selFishs) == 0 {
		fishlogger.Errorf("Event fish die all.")
		this.fishSettlements(selFishs, player, fishEvent.Power, fishEvent.Event, time.Now().UnixNano(), eventFish)
		return false
	}
	delete(player.fishEvent, sign)
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
	this.fishSettlements(selFishs, player, fishEvent.Power, fishEvent.Event, time.Now().UnixNano(), eventFish)
	return true
}
*/
func (this *HPFishingSceneData) RequireCoinPool(player *FishingPlayerData) bool {
	if player.IsRob /*|| this.testing*/ {
		return false
	}
	return true
}

func (this *HPFishingSceneData) BroadCastMessage(packetid int, msg proto.Message, excludeSid int64) {
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
	for _, p := range this.GetAudiences() {
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
func (this *HPFishingSceneData) BroadCastMessageAllPlayers(packetid int, msg proto.Message, excludeSid int64) {
	players := base.SceneMgrSington.GetPlayersByGameFree(this.Platform, this.GetGameFreeId())
	mgs := make(map[*netlib.Session][]*srvlibproto.MCSessionUnion)
	for _, p := range players {
		if p == nil || p.GetGateSess() == nil || p.IsRob {
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
	//for _, p := range this.audiences {
	//	if p == nil || p.gateSess == nil {
	//		continue
	//	}
	//	if !p.IsOnLine() || p.IsMarkFlag(PlayerState_Leave) {
	//		continue
	//	}
	//	if p.sid == excludeSid {
	//		continue
	//	}
	//	mgs[p.gateSess] = append(mgs[p.gateSess], &srvlibproto.MCSessionUnion{
	//		Mccs: &srvlibproto.MCClientSession{
	//			SId: proto.Int64(p.sid),
	//		},
	//	})
	//}
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

func (this *HPFishingSceneData) NewFishJackaptNotice(player *base.Player, num int64, jackType int32, msgType int64) {
	if !this.GetTesting() {
		if num <= 0 {
			return
		}
		start := time.Now().Add(time.Second * 3).Unix()
		content := fmt.Sprintf("%v|%v|%v|%v", player.GetName(), this.GetHundredSceneName(), jackType, num)
		pack := &server_proto.GWNewNotice{
			//PlayerName: proto.String(player.Name),
			Ch:       proto.String(""),
			Content:  proto.String(content),
			Start:    proto.Int64(start),
			Interval: proto.Int64(0),
			Count:    proto.Int64(num),
			Msgtype:  proto.Int64(msgType),
			Platform: proto.String(player.Platform),
			Isrob:    proto.Bool(player.IsRob),
			Priority: proto.Int32(int32(num)),
			//GameFreeid: proto.Int32(this.gamefreeId),
			//GameId:     proto.String(this.keyGameId),
		}
		this.SendToWorld(int(server_proto.SSPacketID_PACKET_GW_NEWNOTICE), pack)
		fishlogger.Trace("NewFishJackaptNotice ", player.GetName(), pack)
	}
}

func (this *HPFishingSceneData) NewGoldFishNotice(player *base.Player, num int32, fishname string, msgType int64) {
	if !this.GetTesting() {
		if num <= 0 {
			return
		}
		start := time.Now().Add(time.Second * 3).Unix()
		content := fmt.Sprintf("%v|%v|%v|%v|%v", player.GetName(), this.KeyGameId, this.KeyGamefreeId, fishname, num)
		pack := &server_proto.GWNewNotice{
			//PlayerName: proto.String(player.Name),
			Ch:       proto.String(""),
			Content:  proto.String(content),
			Start:    proto.Int64(start),
			Interval: proto.Int64(0),
			Count:    proto.Int64(int64(num)),
			Msgtype:  proto.Int64(msgType),
			Platform: proto.String(player.Platform),
			Isrob:    proto.Bool(player.IsRob),
			Priority: proto.Int32(int32(num)),
			//GameFreeid: proto.Int32(this.gamefreeId),
			//GameId:     proto.String(this.keyGameId),
		}
		this.SendToWorld(int(server_proto.SSPacketID_PACKET_GW_NEWNOTICE), pack)
		fishlogger.Trace("NewGoldFishNotice ", pack)
	}
}

//获取游戏场景总投入和总产出
func (this *HPFishingSceneData) GetSysTotalInAndOut() (int64, int64) {
	keyGlobal := fmt.Sprintf("%v_%v", this.Platform, this.KeyGamefreeId)
	if base.SysProfitCoinMgr.SysPfCoin == nil {
		base.SysProfitCoinMgr.SysPfCoin = model.InitSysProfitCoinData(fmt.Sprintf("%d", common.GetSelfSrvId()))
	}

	if base.SysProfitCoinMgr.SysPfCoin.ProfitCoin == nil {
		base.SysProfitCoinMgr.SysPfCoin.ProfitCoin = make(map[string]*model.SysCoin)
	}

	var syscoin *model.SysCoin
	if data, exist := base.SysProfitCoinMgr.SysPfCoin.ProfitCoin[keyGlobal]; !exist {
		syscoin = new(model.SysCoin)
		base.SysProfitCoinMgr.SysPfCoin.ProfitCoin[keyGlobal] = syscoin
	} else {
		syscoin = data
	}
	//总产出初始值 = 100*（1-调节频率）*100 （分）	初级场
	//总产出初始值 = 1000*（1-调节频率）*100 （分）	中级场
	//总产出初始值 = 10000*（1-调节频率）*100 （分）	高级场
	//总投入初始值 = 100*100 （分）	初级场
	//总投入初始值 = 1000*100 （分）	中级场
	//总投入初始值 = 10000*100 （分）	高级场
	initBaseValue := int64(10000) //1万分
	if this.fishLevel == 2 {
		initBaseValue = 100000
	} else if this.fishLevel == 3 {
		initBaseValue = 1000000
	}
	setting := base.GetCoinPoolMgr().GetCoinPoolSetting(this.Platform, this.GetGameFreeId(), this.GetGroupId())
	totalInInitValue := initBaseValue
	totalOutInitValue := initBaseValue * (10000 - int64(setting.GetCtroRate())) / 10000
	return totalInInitValue + syscoin.PlaysBet, totalOutInitValue + syscoin.SysPushCoin
}

//获取系统赔率和预警调控值
func (this *HPFishingSceneData) GetSysOdds() float64 {
	sysTotalIn, sysTotalOut := this.GetSysTotalInAndOut()
	//fishlogger.Tracef("GameId_TFishing sysTotalIn = %v sysTotalOut = %v", sysTotalIn, sysTotalOut)
	return float64(sysTotalOut) / float64(sysTotalIn)
}

//计算预警调控值
func (this *HPFishingSceneData) GetPreCorrect() float64 {
	//初始化预警调控值
	setting := base.GetCoinPoolMgr().GetCoinPoolSetting(this.Platform, this.GetGameFreeId(), this.GetGroupId())
	sysOdds := this.GetSysOdds()
	if sysOdds >= 1 {
		return 0.5
	} else if sysOdds <= float64(10000-setting.GetCtroRate())/10000 {
		return 1.1
	} else {
		return 0.7
	}
}

//打印日志信息
func (this *HPFishingSceneData) ShowTraceInfo(player *FishingPlayerData, gameid string, ctroRate, fishlevel, power, playerSnid, fishId, hp int32, preCorrect, rate, playerRatio float64, death bool, whitelevel, blacklevel, swallowCoin int32) {
	//计算个人赔率
	if data, ok := player.GDatas[gameid]; ok {
		//总产出初始值 = 100*（1-调节频率）*100 （分）	初级场
		//总产出初始值 = 1000*（1-调节频率）*100 （分）	中级场
		//总产出初始值 = 10000*（1-调节频率）*100 （分）	高级场
		//总投入初始值 = 100*100 （分）	初级场
		//总投入初始值 = 1000*100 （分）	中级场
		//总投入初始值 = 10000*100 （分）	高级场
		initBaseValue := int64(10000) //1万分
		if fishlevel == 2 {
			initBaseValue = 100000
		} else if fishlevel == 3 {
			initBaseValue = 1000000
		}
		totalInValue := initBaseValue + data.Statics.TotalIn
		totalOutValue := initBaseValue*(10000-int64(ctroRate))/10000 + data.Statics.TotalOut
		playerOdds := float64(totalOutValue) / float64(totalInValue)

		//计算系统赔率
		sysTotalIn, sysTotalOut := this.GetSysTotalInAndOut()
		sysOdds := float64(sysTotalOut) / float64(sysTotalIn)

		varB := math.Max(2-playerOdds-float64(ctroRate)/10000, 0.8)
		varC := math.Max(2-sysOdds-float64(ctroRate)/10000, 0.8)
		varD := 1 - float64(ctroRate)/10000

		oldHp := hp
		if swallowCoin > hp {
			hp = hp - (swallowCoin - hp)
		}
		//计算击杀概率
		fishlogger.Tracef("GameId_TFishing playerSnid = %v fishId = %v power = %v death = %v whitelevel = %v blacklevel = %v", playerSnid, fishId, power, death, whitelevel, blacklevel)
		fishlogger.Tracef("GameId_TFishing playerTotalIn = %v playerTotalOut = %v playerOdds = %v", totalInValue, totalOutValue, playerOdds)
		fishlogger.Tracef("GameId_TFishing sysTotalIn = %v sysTotalOut = %v sysOdds = %v", sysTotalIn, sysTotalOut, sysOdds)
		fishlogger.Tracef("GameId_TFishing ctroRate = %v preCorrect = %v", ctroRate, preCorrect)
		fishlogger.Tracef("GameId_TFishing hp = %v swallowCoin = %v hp2 = %v", oldHp, swallowCoin, hp)
		fishlogger.Tracef("GameId_TFishing killRate(%v/%v * %v * %v * %v * %v* %v) = %v", power, hp, varB, varC, varD, preCorrect, playerRatio, rate)
		fishlogger.Tracef("!!!!!!!!!!!!!!! TestHitNum:%v  power:%v ", player.TestHitNum, power)
	} else {
		fishlogger.Errorf("player.GDatas[%v] is %v", gameid, player.GDatas[gameid])
	}
}

//同步鱼的血量
func (this *HPFishingSceneData) SyncFishesHp(fishes []*Fish) {
	pack := &fishing_proto.SCSyncFishHp{}
	for _, fish := range fishes {
		info := &fishing_proto.FishHpInfo{
			FishID:  proto.Int32(fish.FishID),
			HpRatio: proto.Int32(fish.Hp.CurrHp * 100 / fish.Hp.Hp),
		}
		pack.HpInfo = append(pack.HpInfo, info)
	}
	this.BroadCastMessage(int(fishing_proto.FIPacketID_FISHING_SC_SYNCFISHHP), pack, 0)
}

//奖池变动
func (this *HPFishingSceneData) AddToJackpot(isRob bool, coin int64, sceneLevel int32) {
	if Jackpot, exist := FishJackpotCoinMgr.Jackpot[this.Platform]; exist {
		switch sceneLevel {
		case 1:
			Jackpot.AddToSmall(isRob, coin)
		case 2:
			Jackpot.AddToMiddle(isRob, coin)
		case 3:
			Jackpot.AddToBig(isRob, coin)
		}
	}
}

//获取奖池
func (this *HPFishingSceneData) GetJackpot(sceneLevel int32) int64 {
	if Jackpot, exist := FishJackpotCoinMgr.Jackpot[this.Platform]; exist {
		switch sceneLevel {
		case 1:
			return Jackpot.Small
		case 2:
			return Jackpot.Middle
		case 3:
			return Jackpot.Big
		default:
			return Jackpot.Small + Jackpot.Middle + Jackpot.Big + model.FishingParamData.JackpotInitCoin
		}
	}
	return 0
}

//获得能量炮的金额
func (this *HPFishingSceneData) GetPranaValue(player *FishingPlayerData, gameid string, ctroRate int32, fishlevel, allCoin int32) int32 {
	playerOdds, _ := player.GetPlayerOdds(gameid, ctroRate, fishlevel)
	sysOdds := this.GetSysOdds()
	a := 2 - playerOdds - float64(ctroRate)/10000
	b := 2 - sysOdds - float64(ctroRate)/10000
	value := float64(allCoin) * a * b
	fishlogger.Tracef("GetPranaValue %v * (2 - %v - %v) * (2 - %v - %v) = %v", allCoin, playerOdds, float64(ctroRate)/10000, sysOdds, float64(ctroRate)/10000, value)
	return int32(value)
}

//增加奖池
func (this *HPFishingSceneData) AddJackpot(player *FishingPlayerData, power int32) {
	player.jackpotCoin += float64(power) * model.FishingParamData.JackpotRate // 临时设置
	jcoin := int64(player.jackpotCoin)
	if jcoin > 0 { // 不够就累加
		if _, exist := FishJackpotCoinMgr.Jackpot[this.Platform]; exist {
			this.AddToJackpot(player.IsRob, jcoin, this.fishLevel)
			pack := &fishing_proto.SCJackpotPool{
				Coin: proto.Int64(this.GetJackpot(0)),
			}
			proto.SetDefaults(pack)
			// 奖池
			this.BroadCastMessageAllPlayers(int(fishing_proto.FIPacketID_FISHING_SC_JACKPOTPOOLCHANGE), pack, 0)
			player.jackpotCoin -= float64(jcoin)
		}
	}
}

//计算能否爆奖池
func (this *HPFishingSceneData) CalcJackpot(player *FishingPlayerData, power int32) {
	key := fmt.Sprintf("%v_%v", this.Platform, this.KeyGamefreeId)
	if !player.IsRob && !this.Testing {
		if _, exist := FishJackpotCoinMgr.Jackpot[this.Platform]; exist {
			setting := base.GetCoinPoolMgr().GetCoinPoolSetting(this.Platform, this.GetGameFreeId(), this.GetGroupId())
			if setting != nil {
				ctroRate := setting.GetCtroRate()
				playerOdds, _ := player.GetPlayerOdds(this.KeyGamefreeId, ctroRate, this.fishLevel)
				jackCoin, jacktype := CalcuJackpotCoin(power, player.GMLevel, ctroRate, key, this.fishLevel, playerOdds)
				if jackCoin != 0 && this.GetJackpot(this.fishLevel)-jackCoin > 0 { // 爆奖前提条件：爆奖金额＜奖池当前金额-奖池初始金额
					// TODO 爆奖 -> insert dblog    updata coin jacklist
					player.CoinCache += jackCoin
					player.ExtraCoin += jackCoin
					player.SetMaxCoin()
					pack := &fishing_proto.SCJackpotCoin{
						SnId:         proto.Int32(player.SnId),
						Coin:         proto.Int32(int32(jackCoin)),
						JackpotLevel: proto.Int32(int32(jacktype)),
						Name:         proto.String(player.Name),
					}
					//Jackpot.AddToBig(player.IsRob, -jackCoin)
					fishlogger.Info("JackpotCoin ", player.SnId, jackCoin, jacktype, this.GetJackpot(this.fishLevel))
					this.AddToJackpot(player.IsRob, -jackCoin, this.fishLevel)
					player.NewStatics(0, jackCoin) //系统产出
					//player.Statics(this.KeyGameId, this.KeyGamefreeId, jackCoin, false) //个人产出
					todata := player.GetTodayGameData(this.KeyGameId)
					todata.TotalOut += jackCoin
					proto.SetDefaults(pack)
					//爆奖池
					this.BroadCastMessageAllPlayers(int(fishing_proto.FIPacketID_FISHING_SC_JACKPOTCOIN), pack, 0)
					wpack := &server_proto.GWGameJackList{
						SnId:     proto.Int32(player.SnId),
						Coin:     proto.Int64(jackCoin),
						RoomId:   proto.Int32(this.GetGameFreeId()),
						GameId:   proto.Int(this.GetGameId()),
						JackType: proto.Int32(jacktype),
						Platform: proto.String(player.Platform),
						Channel:  proto.String(player.Channel),
						Name:     proto.String(player.Name),
					}
					proto.SetDefaults(wpack)
					this.SendToWorld(int(server_proto.SSPacketID_PACKET_GW_JACKPOTLIST), wpack)
					this.NewFishJackaptNotice(player.Player, jackCoin, jacktype, 10)
				}
			}
		}
	}
}
