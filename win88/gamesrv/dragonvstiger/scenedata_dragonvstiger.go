package dragonvstiger

import (
	"fmt"
	"games.yol.com/win88/common"
	. "games.yol.com/win88/gamerule/dragonvstiger"
	"games.yol.com/win88/gamesrv/base"
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/dragonvstiger"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/timer"
	"math"
	"math/rand"
	"sort"
)

//龙虎斗,玩家游戏数据版本号
const DVST_PLAYERGAMEINFO_VERSION = 1

var DVST_TheOdds [DVST_ZONE_MAX]int = [DVST_ZONE_MAX]int{8, 1, 1}

type DragonVsTigerSceneData struct {
	*base.Scene
	poker         *Poker
	cards         [2]int32
	players       map[int32]*DragonVsTigerPlayerData
	upplayerlist  []*DragonVsTigerPlayerData //上庄玩家列表
	isWin         [3]int
	seats         []*DragonVsTigerPlayerData
	winTop1       *DragonVsTigerPlayerData
	betTop5       [DVST_RICHTOP5]*DragonVsTigerPlayerData
	bankerSnId    int32 //当前庄家位置
	betInfo       [DVST_ZONE_MAX]int
	betInfoRob    [DVST_ZONE_MAX]int
	betDetailInfo [DVST_ZONE_MAX]map[int]int
	trend100Cur   []int32
	trend20Lately []int32
	by            func(p, q *DragonVsTigerPlayerData) bool
	hRunRecord    timer.TimerHandle
	hBatchSend    timer.TimerHandle
	constSeatKey  string
	bIntervention bool
	webUser       string
	upplayerCount int32 //当前上庄玩家已在庄次数
	bankerWinCoin int64 //四个闲家的总金额
	bankerbetCoin int64 //四个闲家的下注
}
type DVTSortData struct {
	*DragonVsTigerPlayerData
	lostCoin int
	winCoin  int64
	realRate float64
}

func (s *DragonVsTigerSceneData) Len() int {
	return len(s.seats)
}
func (s *DragonVsTigerSceneData) Less(i, j int) bool {
	return s.by(s.seats[i], s.seats[j])
}
func (s *DragonVsTigerSceneData) Swap(i, j int) {
	s.seats[i], s.seats[j] = s.seats[j], s.seats[i]
}
func NewDragonVsTigerSceneData(s *base.Scene) *DragonVsTigerSceneData {
	return &DragonVsTigerSceneData{
		Scene:   s,
		players: make(map[int32]*DragonVsTigerPlayerData),
	}
}
func (this *DragonVsTigerSceneData) RebindPlayerSnId(oldSnId, newSnId int32) {
	if p, exist := this.players[oldSnId]; exist {
		delete(this.players, oldSnId)
		this.players[newSnId] = p
	}
}
func (this *DragonVsTigerSceneData) init() bool {
	this.bankerSnId = -1
	if this.DbGameFree == nil {
		return false
	}
	this.poker = NewPoker()
	this.LoadData()
	return true
}
func (this *DragonVsTigerSceneData) LoadData() {
}
func (this *DragonVsTigerSceneData) SaveData() {
}
func (this *DragonVsTigerSceneData) Clean() {
	for _, p := range this.players {
		p.Clean()
	}
	for i := 0; i < len(this.cards); i++ {
		this.cards[i] = -1
	}
	for i := 0; i < DVST_ZONE_MAX; i++ {
		this.betInfo[i] = 0
		this.betInfoRob[i] = 0
		this.betDetailInfo[i] = make(map[int]int)
	}
	this.bIntervention = false
	this.webUser = ""
	//重置水池调控标记
	this.CpControlled = false
	this.bankerWinCoin = 0
	this.bankerbetCoin = 0
	this.SystemCoinOut = 0
}
func (this *DragonVsTigerSceneData) CanStart() bool {
	cnt := len(this.players)
	if cnt > 0 {
		return true
	}
	return false
}
func (this *DragonVsTigerSceneData) delPlayer(p *base.Player) {
	if p, exist := this.players[p.SnId]; exist {
		delete(this.players, p.SnId)
		count := len(this.seats)
		idx := -1
		for i := 0; i < count; i++ {
			if this.seats[i].SnId == p.SnId {
				idx = i
				break
			}
		}
		if idx != -1 {
			temp := this.seats[:idx]
			temp = append(temp, this.seats[idx+1:]...)
			this.seats = temp
		}
	}
}
func (this *DragonVsTigerSceneData) OnPlayerEnter(p *base.Player) {
	if (p.Coin) < int64(this.DbGameFree.GetLimitCoin()) {
		p.MarkFlag(base.PlayerState_GameBreak)
		p.SyncFlag()
	}
}
func (this *DragonVsTigerSceneData) OnPlayerLeave(p *base.Player, reason int) {
	this.delPlayer(p)
	this.BroadcastPlayerLeave(p, reason)
}
func (this *DragonVsTigerSceneData) BroadcastPlayerLeave(p *base.Player, reason int) {
}
func (this *DragonVsTigerSceneData) SceneDestroy(force bool) {

	this.Scene.Destroy(force)
}
func (this *DragonVsTigerSceneData) IsNeedCare(p *DragonVsTigerPlayerData) bool {
	if p.IsNeedCare() {
		return true
	}
	return false
}
func (this *DragonVsTigerSceneData) IsNeedWeaken(p *DragonVsTigerPlayerData) bool {
	if p.IsNeedWeaken() {
		return true
	}
	return false
}
func DVST_OrderByLatest20Win(l, r *DragonVsTigerPlayerData) bool {
	return l.lately20Win > r.lately20Win
}
func DVST_OrderByLatest20Bet(l, r *DragonVsTigerPlayerData) bool {
	return l.lately20Bet > r.lately20Bet
}
func (this *DragonVsTigerSceneData) Resort() string {
	constSeatKey := fmt.Sprintf("%v", this.bankerSnId)
	best := int64(0)
	count := len(this.seats)
	this.winTop1 = nil
	for i := 0; i < count; i++ {
		playerEx := this.seats[i]
		if playerEx != nil && playerEx.SnId != this.bankerSnId {
			playerEx.Pos = DVST_OLPOS
			if playerEx.lately20Win > best {
				best = playerEx.lately20Win
				this.winTop1 = playerEx
			}
		}
	}
	if this.winTop1 != nil {
		this.winTop1.Pos = DVST_BESTWINPOS
		constSeatKey = fmt.Sprintf("%d_%d", this.winTop1.SnId, this.winTop1.Pos)
	}
	this.by = DVST_OrderByLatest20Bet
	sort.Sort(this)
	cnt := 0
	for i := 0; i < DVST_RICHTOP5; i++ {
		this.betTop5[i] = nil
	}
	for i := 0; i < count && cnt < DVST_RICHTOP5; i++ {
		playerEx := this.seats[i]
		if playerEx != this.winTop1 && playerEx.SnId != this.bankerSnId {
			playerEx.Pos = cnt + 1
			this.betTop5[cnt] = playerEx
			cnt++
			constSeatKey += fmt.Sprintf("%d_%d", playerEx.SnId, playerEx.Pos)
		}
	}
	return constSeatKey
}

func (this *DragonVsTigerSceneData) CalcAllAvgBet() {
	for _, p := range this.players {
		if !p.IsRob && p.IsGameing() {
			//today := p.GetTodayGameData(this.DbGameFree.GetId())
			//if today.AvgBetCoin == 0 {
			//	today.AvgBetCoin = p.betTotal
			//} else {
			//	today.AvgBetCoin = int64(float64(today.AvgBetCoin)*model.NormalParamData.BetWeight[0] +
			//		float64(p.betTotal)*model.NormalParamData.BetWeight[1])
			//}
		}

	}

}

func (this *DragonVsTigerSceneData) CalcuResult() int {
	dragonCard := this.cards[0] % PER_CARD_COLOR_MAX
	tigerCard := this.cards[1] % PER_CARD_COLOR_MAX
	if dragonCard == tigerCard {
		return DVST_ZONE_DRAW
	} else if dragonCard > tigerCard {
		return DVST_ZONE_DRAGON
	}
	return DVST_ZONE_TIGER
}

func (this *DragonVsTigerSceneData) PushTrend(result int32) {
	this.trend100Cur = append(this.trend100Cur, result)
	cnt := len(this.trend100Cur)
	if cnt >= DVST_TREND100 {
		this.trend100Cur = nil
	}
	this.trend20Lately = append(this.trend20Lately, result)
	cnt = len(this.trend20Lately)
	if cnt > 20 {
		this.trend20Lately = this.trend20Lately[cnt-20:]
	}
}
func (this *DragonVsTigerSceneData) ConsecutiveCheck(result int) {
	count := 0
	for i := len(this.trend20Lately) - 1; i > 0; i-- {
		if this.trend20Lately[i] == int32(result) {
			count++
		} else {
			break
		}
	}
	if count <= 2 {
		count = 0
	} else {
		count -= 2
	}
	repairRate := 1 + 0.1*(float64(common.RandInt(model.NormalParamData.LHMin, model.NormalParamData.LHMax))/float64(10000))
	repairRate = math.Pow(repairRate, float64(count-1)) - 1
	rate := int(repairRate * 10000)
	if common.RandInt(10000) < rate {
		if result == DVST_ZONE_DRAW {
			//for i := 0; i < 2; i++ {
			//	this.cards[i] = int32(this.poker.Next())
			//}
		} else {
			this.cards[1], this.cards[0] = this.cards[0], this.cards[1]
		}
	}
	this.CalcuResult()
}
func (this *DragonVsTigerSceneData) GetRecoveryCoe() float64 {
	if coe, ok := model.GameParamData.HundredSceneRecoveryCoe[this.KeyGamefreeId]; ok {
		return float64(coe) / 100
	}
	return 0.05
}

// SystemBankerPlayerOut 系统庄结算后玩家赢分
func (this *DragonVsTigerSceneData) SystemBankerPlayerOut(winIndex int) int {
	if !this.IsRobotBanker() {
		return 0
	}

	sysOutCoin := 0
	systemOutCoin := [DVST_ZONE_MAX]int{0, 0, 0}

	for key, value := range this.betInfo {
		systemOutCoin[key] += (value - this.betInfoRob[key])
	}

	for i := 0; i < DVST_ZONE_MAX; i++ {
		if winIndex == i {
			sysOutCoin += systemOutCoin[i] * DVST_TheOdds[i]
		} else {
			sysOutCoin -= systemOutCoin[i]
		}
	}

	if winIndex == DVST_ZONE_DRAW {
		if systemOutCoin[DVST_ZONE_DRAGON] != 0 {
			sysOutCoin += systemOutCoin[DVST_ZONE_DRAGON]
		}
		if systemOutCoin[DVST_ZONE_TIGER] != 0 {
			sysOutCoin += systemOutCoin[DVST_ZONE_TIGER]
		}
	}
	return sysOutCoin
}

func (this *DragonVsTigerSceneData) ChangeCardByRand(result int) bool {
	if result != DVST_ZONE_DRAW {
		resultCnt := 0
		totalCnt := len(this.trend20Lately)
		for i := 0; i < totalCnt; i++ {
			if this.trend20Lately[i] == int32(result) {
				resultCnt++
			}
		}
		if totalCnt > 2 {
			if resultCnt > totalCnt/2 {
				if this.Rand.Intn(2) == 1 {
					this.cards[0], this.cards[1] = this.cards[1], this.cards[0]
					log := common.GetLoggerInstanceByName("dtlogic")
					if log != nil {
						log.Tracef("DragonVsTigerSceneData.ChangeCardByRand(result=%v)", result)
					}
					return true
				}
			}
		}
	}
	return false
}

func (this *DragonVsTigerSceneData) ChangeCard(result int) bool {
	if model.GameParamData.UseNewNumericalLogic && this.Rand.Intn(100) < model.GameParamData.UseNewNumericalLogicPercent { //一半的概率走新规则，一半的概率走老规则
		if result != DVST_ZONE_DRAW {
			log := common.GetLoggerInstanceByName("dtlogic")
			setting := base.CoinPoolMgr.GetCoinPoolSetting(this.Platform, this.GetGameFreeId(), this.GroupId)
			if setting != nil {
				poolCap := (setting.GetUpperLimit() - setting.GetLowerLimit()) / 2
				middleBase := (setting.GetUpperLimit() + setting.GetLowerLimit()) / 2                   //中线
				curCoin := base.CoinPoolMgr.LoadCoin(this.GetGameFreeId(), this.Platform, this.GroupId) //当前水位线
				diff := ((curCoin - int64(middleBase)) * 100 / int64(poolCap))                          //偏移中线多少
				sysOut := this.SystemBankerPlayerOut(result)                                            //本局系统产出
				if diff > 0 {                                                                           //水位线偏上
					if sysOut > 0 { //输分
						val := 50 - diff
						if val <= 0 {
							val = rand.Int63n(10)
						}
						rnd := this.Rand.Int63n(100)
						if log != nil {
							log.Tracef("[1]DragonVsTigerSceneData.ChangeCard(scene=%v,result=%v) upper=%v lower=%v mid=%v diff=%v currCoin=%v sysOut=%v val=%v rnd=%v",
								this.SceneId,
								result,
								setting.GetUpperLimit(),
								setting.GetLowerLimit(),
								middleBase,
								diff,
								curCoin,
								sysOut,
								val,
								rnd)
						}
						if rnd < val { //小概率[赢]翻转
							log.Tracef("[1][triger]DragonVsTigerSceneData.ChangeCard(scene=%v,result=%v)val=%v rnd=%v", this.SceneId, result, val, rnd)
							this.cards[0], this.cards[1] = this.cards[1], this.cards[0]
							return true
						}
					} else if sysOut < 0 { //赢分
						val := 50 + diff
						if val >= 100 { //上下浮动下
							val = 100 - rand.Int63n(10)
						}
						rnd := this.Rand.Int63n(100)
						if log != nil {
							log.Tracef("[2]DragonVsTigerSceneData.ChangeCard(scene=%v,result=%v) upper=%v lower=%v mid=%v diff=%v currCoin=%v sysOut=%v val=%v rnd=%v",
								this.SceneId,
								result,
								setting.GetUpperLimit(),
								setting.GetLowerLimit(),
								middleBase,
								diff,
								curCoin,
								sysOut,
								val,
								rnd)
						}
						if rnd < val { //大概率[输]翻转
							log.Tracef("[2][triger]DragonVsTigerSceneData.ChangeCard(scene=%v,result=%v)val=%v rnd=%v", this.SceneId, result, val, rnd)
							this.cards[0], this.cards[1] = this.cards[1], this.cards[0]
							return true
						}
					}
				} else if diff < 0 { //水位线偏下
					if sysOut > 0 { //输分
						if sysOut > int(setting.GetMaxOutValue()) {
							this.cards[0], this.cards[1] = this.cards[1], this.cards[0]
							return true
						}
						val := 50 - diff
						if val >= 100 { //上下浮动下
							val = 100 - rand.Int63n(10)
						}
						rnd := this.Rand.Int63n(100)
						if log != nil {
							log.Tracef("[3]DragonVsTigerSceneData.ChangeCard(scene=%v,result=%v) upper=%v lower=%v mid=%v diff=%v currCoin=%v sysOut=%v val=%v rnd=%v",
								this.SceneId,
								result,
								setting.GetUpperLimit(),
								setting.GetLowerLimit(),
								middleBase,
								diff,
								curCoin,
								sysOut,
								val,
								rnd)
						}
						if rnd < val { //大概率[赢]翻转
							log.Tracef("[3][triger]DragonVsTigerSceneData.ChangeCard(scene=%v,result=%v)val=%v rnd=%v", this.SceneId, result, val, rnd)
							this.cards[0], this.cards[1] = this.cards[1], this.cards[0]
							return true
						}
					} else if sysOut < 0 { //赢分
						if -sysOut > int(setting.GetMaxOutValue()) { //如果翻转后，吐分量过大，不做调牌
							return false
						}
						val := 50 + diff
						if val <= 0 { //上下浮动下
							val = rand.Int63n(10)
						}
						rnd := this.Rand.Int63n(100)
						if log != nil {
							log.Tracef("[4]DragonVsTigerSceneData.ChangeCard(scene=%v,result=%v) upper=%v lower=%v mid=%v diff=%v currCoin=%v sysOut=%v val=%v rnd=%v",
								this.SceneId,
								result,
								setting.GetUpperLimit(),
								setting.GetLowerLimit(),
								middleBase,
								diff,
								curCoin,
								sysOut,
								val,
								rnd)
						}
						if rnd < val { //小概率[输]翻转
							log.Tracef("[4][triger]DragonVsTigerSceneData.ChangeCard(scene=%v,result=%v)val=%v rnd=%v", this.SceneId, result, val, rnd)
							this.cards[0], this.cards[1] = this.cards[1], this.cards[0]
							return true
						}
					}
				}
			}
		}
		return false

	} else {
		sysOut := int64(this.SystemBankerPlayerOut(result))
		if !base.CoinPoolMgr.IsMaxOutHaveEnough(this.Platform, this.GetGameFreeId(), this.GroupId, -sysOut) {
			this.cards[0], this.cards[1] = this.cards[1], this.cards[0]
			return true
		}
		status, changeRate := base.CoinPoolMgr.GetCoinPoolStatus2(this.Platform, this.GetGameFreeId(), this.GroupId, -sysOut)
		switch status {
		case base.CoinPoolStatus_Normal:
			//增加杀的概率
			if common.RandInt(10000) < model.NormalParamData.LHChangeCardRate {
				if this.SystemBankerPlayerOut(result) > 0 {
					this.cards[0], this.cards[1] = this.cards[1], this.cards[0]
					this.CpControlled = true
					return true
				}
			}
			/*
				if result != DVST_ZONE_DRAW {
					curCoin := base.CoinPoolMgr.LoadCoin(this.gamefreeId, this.platform, this.groupId)
					systemOutCoin := [DVST_ZONE_MAX]int{0, 0, 0}

					for key, value := range this.betInfo {
						systemOutCoin[key] += (value - this.betInfoRob[key])
					}
					rate := DVST_TheOdds[result]
					setting := base.CoinPoolMgr.GetCoinPoolSetting(this.platform, this.gamefreeId, this.groupId)
					if (curCoin - int64(systemOutCoin[result])*int64(rate)) < int64(setting.GetLowerLimit()) {
						if this.SystemBankerPlayerOut(result) > 0 {
							this.cards[0], this.cards[1] = this.cards[1], this.cards[0]
							return true
						}
					}
				}
			*/
		case base.CoinPoolStatus_Low:
			if common.RandInt(10000) < changeRate {
				if this.SystemBankerPlayerOut(result) > 0 {
					this.cards[0], this.cards[1] = this.cards[1], this.cards[0]
					this.CpControlled = true
					return true
				}
			}

		case base.CoinPoolStatus_High:
			fallthrough
		case base.CoinPoolStatus_TooHigh:
			if common.RandInt(10000) < changeRate && this.CoinPoolCanOut() {
				systmCoinOut := this.SystemBankerPlayerOut(result)
				if systmCoinOut < 0 {

					this.cards[0], this.cards[1] = this.cards[1], this.cards[0]
					this.CpControlled = true
					reIndex := this.CalcuResult()
					sysOut := int64(this.SystemBankerPlayerOut(reIndex))
					if !base.CoinPoolMgr.IsMaxOutHaveEnough(this.Platform, this.GetGameFreeId(), this.GroupId, -sysOut) {
						this.cards[0], this.cards[1] = this.cards[1], this.cards[0]
						return true
					}
					return true
				}
			}

		default:

		}
		return false
	}
}

//升级玩家游戏数据
func (this *DragonVsTigerSceneData) UpgradePlayerGameInfo(pgi *model.PlayerGameInfo) {
	if pgi == nil {
		return
	}

	if pgi.Version >= DVST_PLAYERGAMEINFO_VERSION {
		return
	}

	pgi.Statics.TotalIn = 0
	pgi.Statics.TotalOut = 0
	pgi.Version = DVST_PLAYERGAMEINFO_VERSION
}

func (this *DragonVsTigerSceneData) AutoBalance3() bool {
	//if !model.GameParamData.DtAutoBalance3 {
	//	return false
	//}
	//setting := base.CoinPoolMgr.GetCoinPoolSetting(this.Platform, this.GetGameFreeId(), this.GroupId) // 获取当前水池的配置表
	//if setting == nil {
	//	return false
	//}
	//
	////系统期望赔率
	//expectOdds := (float64(10000-setting.GetCtroRate()) / 10000)
	//keyPf := fmt.Sprintf("%v_%v", this.Platform, this.GetGameFreeId())
	//// 计算当前平台收益命中修正系数
	//in, out := base.SysProfitCoinMgr.GetSysPfCoin(keyPf)
	//if in < 100 {
	//	defPool := setting.GetInitValue() - setting.GetLowerLimit()
	//	if defPool > 0 {
	//		base.SysProfitCoinMgr.Add(keyPf, int64(defPool), 0)
	//		in += int64(defPool)
	//	}
	//}
	//
	//var initVal int32
	//for _, v := range this.DbGameFree.GetMaxBetCoin() {
	//	initVal += v
	//}
	//in += int64(initVal)
	//out += int64(float64(initVal) * expectOdds)
	//
	//var allowZones []int
	//var willOdds float64
	//for i := 0; i < DVST_ZONE_MAX; i++ {
	//	sysOut := this.GetSystemChangeCoin(i)
	//	if sysOut >= 0 {
	//		willOdds = float64(out) / float64(in+sysOut)
	//	} else {
	//		willOdds = float64(out-sysOut) / float64(in)
	//	}
	//	//当前区域开奖,赔率满足需求
	//	if willOdds <= expectOdds {
	//		allowZones = append(allowZones, i)
	//	}
	//}
	//
	//result := this.CalcuResult()
	//// 满足开奖结算，不换牌
	//if common.InSliceInt(allowZones, result) {
	//	return true
	//}
	//
	//// 到这里len(allowZones) <= 2, 并且当前开奖结果和想要的结果不一样
	//
	//if len(allowZones) > 1 {
	//	allowZones = common.DelSliceInt(allowZones, DVST_ZONE_DRAW)
	//}
	//if len(allowZones) == 1 {
	//	this.cards[0], this.cards[1] = this.poker.TryDTDraw(int32(allowZones[0]))
	//	return true
	//}
	//// 龙虎两个区域都满足开奖结果，或者都不满足开奖结果时
	//var avrInOut [DVST_ZONE_MAX]int64
	//var avrOdds [DVST_ZONE_MAX]float64
	//for i := 1; i < DVST_ZONE_MAX; i++ {
	//	count := 0
	//	for _, p := range this.players {
	//		if !p.IsRob && p.betInfo[i] > 0 {
	//			count++
	//			pgi := p.GetGameFreeIdData(this.KeyGamefreeId)
	//			if pgi != nil {
	//				this.UpgradePlayerGameInfo(pgi)
	//				avrInOut[i] += pgi.Statics.TotalOut - pgi.Statics.TotalIn
	//				avrOdds[i] += (float64(pgi.Statics.TotalOut) + float64(initVal)*expectOdds) / (float64(pgi.Statics.TotalIn) + float64(initVal))
	//			} else {
	//				avrOdds[i] += 1
	//			}
	//		}
	//	}
	//	//求总输赢和总赔率的均值
	//	if count > 0 {
	//		avrInOut[i] /= int64(count)
	//		avrOdds[i] /= float64(count)
	//	}
	//}
	////两个区域，没人下注，或者输赢额均值=0，不干涉结果
	//if avrInOut[DVST_ZONE_DRAGON] == 0 && avrInOut[DVST_ZONE_TIGER] == 0 {
	//	//尽可能让龙虎区域平衡
	//	this.ChangeCardByRand(result)
	//	return true
	//}
	//var expectResult int
	//if avrInOut[DVST_ZONE_DRAGON] >= 0 && avrInOut[DVST_ZONE_TIGER] <= 0 { //偏向于虎赢
	//	if rand.Float64() < 0.3 {
	//		expectResult = DVST_ZONE_DRAGON
	//	} else {
	//		expectResult = DVST_ZONE_TIGER
	//	}
	//} else if avrInOut[DVST_ZONE_DRAGON] <= 0 && avrInOut[DVST_ZONE_TIGER] >= 0 { //偏向于龙赢
	//	if rand.Float64() < 0.3 {
	//		expectResult = DVST_ZONE_TIGER
	//	} else {
	//		expectResult = DVST_ZONE_DRAGON
	//	}
	//} else if (avrInOut[DVST_ZONE_DRAGON] > 0 && avrInOut[DVST_ZONE_TIGER] > 0) || (avrInOut[DVST_ZONE_DRAGON] < 0 && avrInOut[DVST_ZONE_TIGER] < 0) {
	//	x1 := float64(avrInOut[DVST_ZONE_TIGER]) / float64(avrInOut[DVST_ZONE_TIGER]+avrInOut[DVST_ZONE_DRAGON])
	//	x2 := 1 - x1
	//	y1 := float64(avrOdds[DVST_ZONE_TIGER]) / float64(avrOdds[DVST_ZONE_TIGER]+avrOdds[DVST_ZONE_DRAGON])
	//	y2 := 1 - y1
	//	tRate := (x1 * y1) / (x1*y1 + x2*y2)
	//	if tRate < 0.3 {
	//		tRate = 0.3
	//	} else if tRate > 0.7 {
	//		tRate = 0.7
	//	}
	//	if rand.Float64() < tRate {
	//		expectResult = DVST_ZONE_TIGER
	//	} else {
	//		expectResult = DVST_ZONE_DRAGON
	//	}
	//}
	////结果不等于期望,换牌
	//if expectResult != result {
	//	this.cards[0], this.cards[1] = this.poker.TryDTDraw(int32(expectResult))
	//}
	return true
}

func (this *DragonVsTigerSceneData) AutoBalanceBank3() bool {
	//if !model.GameParamData.DtAutoBalance3 {
	//	return false
	//}
	//
	//setting := base.CoinPoolMgr.GetCoinPoolSetting(this.Platform, this.GetGameFreeId(), this.GroupId) // 获取当前水池的配置表
	//if setting == nil {
	//	return false
	//}
	//
	////系统期望赔率
	//expectOdds := (float64(10000-setting.GetCtroRate()) / 10000)
	//keyPf := fmt.Sprintf("%v_%v", this.Platform, this.GetGameFreeId())
	//// 计算当前平台收益命中修正系数
	//in, out := base.SysProfitCoinMgr.GetSysPfCoin(keyPf)
	//
	//var initVal int32
	//for _, v := range this.DbGameFree.GetMaxBetCoin() {
	//	initVal += v
	//}
	//in += int64(initVal)
	//out += int64(float64(initVal) * expectOdds)
	//
	//var allowZones []int
	//var willOdds float64
	//for i := 0; i < DVST_ZONE_MAX; i++ {
	//	sysOut := this.GetSystemChangeCoin(i)
	//	if sysOut >= 0 {
	//		willOdds = float64(out) / float64(in+sysOut)
	//	} else {
	//		willOdds = float64(out-sysOut) / float64(in)
	//	}
	//	//当前区域开奖,赔率满足需求
	//	if willOdds <= expectOdds {
	//		allowZones = append(allowZones, i)
	//	}
	//}
	//
	//result := this.CalcuResult()
	//// 满足开奖结算，不换牌
	//if common.InSliceInt(allowZones, result) {
	//	return true
	//}
	//
	//// 到这里len(allowZones) <= 2, 并且当前开奖结果和想要的结果不一样
	//
	//if len(allowZones) > 1 {
	//	allowZones = common.DelSliceInt(allowZones, DVST_ZONE_DRAW)
	//}
	//if len(allowZones) == 1 {
	//	this.cards[0], this.cards[1] = this.poker.TryDTDraw(int32(allowZones[0]))
	//	return true
	//}
	//// 龙虎两个区域都满足开奖结果，或者都不满足开奖结果时
	//var avrInOut [DVST_ZONE_MAX]int64
	//var avrOdds [DVST_ZONE_MAX]float64
	//key := strconv.Itoa(int(this.DbGameFree.GetId()))
	//if banker, ok := this.GetBanker(); banker != nil && !ok {
	//	var bankerInOut int64
	//	var bankerOdd float64
	//	pgi := banker.GetGameFreeIdData(key)
	//	if pgi != nil {
	//		this.UpgradePlayerGameInfo(pgi)
	//		bankerInOut = (pgi.Statics.TotalOut - pgi.Statics.TotalIn)
	//		bankerOdd = (float64(pgi.Statics.TotalOut) + float64(initVal)*expectOdds) / (float64(pgi.Statics.TotalIn) + float64(initVal))
	//	}
	//	for i := 1; i < DVST_ZONE_MAX; i++ {
	//		count := 1
	//		avrInOut[i] += bankerInOut
	//		avrOdds[i] += bankerOdd
	//		for _, p := range this.players {
	//			if !p.IsRob && p.betInfo[i] > 0 {
	//				count++
	//				pgi := p.GetGameFreeIdData(key)
	//				if pgi != nil {
	//					this.UpgradePlayerGameInfo(pgi)
	//					avrInOut[i] += (pgi.Statics.TotalOut - pgi.Statics.TotalIn)
	//					avrOdds[i] += (float64(pgi.Statics.TotalOut) + float64(initVal)*expectOdds) / (float64(pgi.Statics.TotalIn) + float64(initVal))
	//				} else {
	//					avrOdds[i] += 1
	//				}
	//			}
	//		}
	//		//求总输赢和总赔率的均值
	//		if count > 0 {
	//			avrInOut[i] /= int64(count)
	//			avrOdds[i] /= float64(count)
	//		}
	//	}
	//	//两个区域，没人下注，或者输赢额均值=0，不干涉结果
	//	if avrInOut[DVST_ZONE_DRAGON] == 0 && avrInOut[DVST_ZONE_TIGER] == 0 {
	//		return true
	//	}
	//	var expectResult int
	//	if avrInOut[DVST_ZONE_DRAGON] >= 0 && avrInOut[DVST_ZONE_TIGER] <= 0 { //偏向于虎赢
	//		if rand.Float64() < 0.3 {
	//			expectResult = DVST_ZONE_DRAGON
	//		} else {
	//			expectResult = DVST_ZONE_TIGER
	//		}
	//	} else if avrInOut[DVST_ZONE_DRAGON] <= 0 && avrInOut[DVST_ZONE_TIGER] >= 0 { //偏向于龙赢
	//		if rand.Float64() < 0.3 {
	//			expectResult = DVST_ZONE_TIGER
	//		} else {
	//			expectResult = DVST_ZONE_DRAGON
	//		}
	//	} else if (avrInOut[DVST_ZONE_DRAGON] > 0 && avrInOut[DVST_ZONE_TIGER] > 0) || (avrInOut[DVST_ZONE_DRAGON] < 0 && avrInOut[DVST_ZONE_TIGER] < 0) {
	//		x1 := float64(avrInOut[DVST_ZONE_TIGER]) / float64(avrInOut[DVST_ZONE_TIGER]+avrInOut[DVST_ZONE_DRAGON])
	//		x2 := 1 - x1
	//		y1 := float64(avrOdds[DVST_ZONE_TIGER]) / float64(avrOdds[DVST_ZONE_TIGER]+avrOdds[DVST_ZONE_DRAGON])
	//		y2 := 1 - y1
	//		tRate := (x1 * y1) / (x1*y1 + x2*y2)
	//		if tRate < 0.3 {
	//			tRate = 0.3
	//		} else if tRate > 0.7 {
	//			tRate = 0.7
	//		}
	//		if rand.Float64() < tRate {
	//			expectResult = DVST_ZONE_TIGER
	//		} else {
	//			expectResult = DVST_ZONE_DRAGON
	//		}
	//	}
	//	//结果不等于期望,换牌
	//	if expectResult != result {
	//		this.cards[0], this.cards[1] = this.poker.TryDTDraw(int32(expectResult))
	//	}
	//}
	return true
}

func (this *DragonVsTigerSceneData) AutoBalance() bool {
	//if !model.GameParamData.DtAutoBalance {
	//	return false
	//}
	//wins := []*DVTSortData{}
	//for _, value := range this.players {
	//	if value.betTotal == 0 || value.IsRob {
	//		continue
	//	}
	//	realRate := float64(value.WinCoin+1) / float64(value.FailCoin+1)
	//	realCoin := value.WinCoin - value.FailCoin
	//	if realRate >= model.GameParamData.DtAutoBalanceRate || realCoin > model.GameParamData.DtAutoBalanceCoin {
	//		loseCoin, _, _ := value.GetLoseMaxPos()
	//		wins = append(wins, &DVTSortData{
	//			DragonVsTigerPlayerData: value,
	//			lostCoin:                loseCoin,
	//			winCoin:                 value.WinCoin,
	//			realRate:                realRate,
	//		})
	//	}
	//}
	//var beKill *DVTSortData
	//switch len(wins) {
	//case 0:
	//	beKill = nil
	//	return false
	//case 1:
	//	beKill = wins[0]
	//	loseCoin, loseMaxPos, _ := beKill.GetLoseMaxPos()
	//	systemWinCoin := this.SystemLoseCoin(loseMaxPos)
	//	realCoin := loseCoin + systemWinCoin
	//	if systemWinCoin == 0 {
	//		systemWinCoin = 1
	//	}
	//	globalRate := float64(loseCoin) / float64(systemWinCoin)
	//	playerRate := float64(loseCoin) / float64(beKill.WinCoin+1)
	//	if realCoin < 0 && globalRate <= playerRate {
	//		return false
	//	}
	//	if realCoin < 0 && globalRate <= 0.8 {
	//		return false
	//	}
	//	this.cards[0], this.cards[1] = this.poker.TryDTDraw(int32(loseMaxPos))
	//	betinfo := [3]int{beKill.betInfo[0], beKill.betInfo[1], beKill.betInfo[2]}
	//	wincoin := beKill.WinCoin
	//	failcoin := beKill.FailCoin
	//	swc := systemWinCoin
	//	card := [2]int32{this.cards[0], this.cards[1]}
	//	task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
	//		model.NewDtAutoBalanceLog(beKill.SnId, betinfo, wincoin, failcoin, swc, card)
	//		return nil
	//	}), nil, "NewDtAutoBalanceLog").Start()
	//	return true
	//default:
	//	sort.Slice(wins, func(i, j int) bool {
	//		if wins[i].lostCoin == wins[j].lostCoin {
	//			if wins[i].winCoin == wins[j].winCoin {
	//				return wins[i].realRate > wins[j].realRate
	//			} else {
	//				return wins[i].winCoin > wins[j].winCoin
	//			}
	//		} else {
	//			return wins[i].lostCoin > wins[j].lostCoin
	//		}
	//	})
	//	for _, value := range wins {
	//		loseCoin, loseMaxPos, _ := value.GetLoseMaxPos()
	//		systemWinCoin := this.SystemLoseCoin(loseMaxPos)
	//		realCoin := loseCoin + systemWinCoin
	//		if systemWinCoin == 0 {
	//			systemWinCoin = 1
	//		}
	//		globalRate := float64(loseCoin) / float64(systemWinCoin)
	//		playerRate := float64(loseCoin) / float64(value.WinCoin+1)
	//		if realCoin < 0 && globalRate <= playerRate {
	//			continue
	//		}
	//		if realCoin < 0 && globalRate <= 0.8 {
	//			continue
	//		}
	//		this.cards[0], this.cards[1] = this.poker.TryDTDraw(int32(loseMaxPos))
	//		betinfo := [3]int{value.betInfo[0], value.betInfo[1], value.betInfo[2]}
	//		wincoin := value.WinCoin
	//		failcoin := value.FailCoin
	//		swc := systemWinCoin
	//		card := [2]int32{this.cards[0], this.cards[1]}
	//		task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
	//			model.NewDtAutoBalanceLog(value.SnId, betinfo, wincoin, failcoin, swc, card)
	//			return nil
	//		}), nil, "NewDtAutoBalanceLog").Start()
	//		return true
	//	}
	return false
	//}
}
func (this *DragonVsTigerSceneData) SystemLoseCoin(pos int) int {
	var noRobotBet [DVST_ZONE_MAX]int
	noRobotBet[DVST_ZONE_DRAGON] = this.betInfo[DVST_ZONE_DRAGON] - this.betInfoRob[DVST_ZONE_DRAGON]
	noRobotBet[DVST_ZONE_TIGER] = this.betInfo[DVST_ZONE_TIGER] - this.betInfoRob[DVST_ZONE_TIGER]
	noRobotBet[DVST_ZONE_DRAW] = this.betInfo[DVST_ZONE_DRAW] - this.betInfoRob[DVST_ZONE_DRAW]
	switch pos {
	case DVST_ZONE_DRAGON:
		return noRobotBet[DVST_ZONE_TIGER] + noRobotBet[DVST_ZONE_DRAW] - int(float64(noRobotBet[DVST_ZONE_DRAGON])*0.95)
	case DVST_ZONE_TIGER:
		return noRobotBet[DVST_ZONE_DRAGON] + noRobotBet[DVST_ZONE_DRAW] - int(float64(noRobotBet[DVST_ZONE_TIGER])*0.95)
	case DVST_ZONE_DRAW:
		return noRobotBet[DVST_ZONE_TIGER] + noRobotBet[DVST_ZONE_DRAGON] - int(float64(noRobotBet[DVST_ZONE_DRAW])*7.95)
	}
	return 0
}

func (this *DragonVsTigerSceneData) IsRobotBanker() bool {
	banker := this.players[this.bankerSnId]
	if banker == nil || banker.IsRob {
		return true
	}
	return false
}
func (this *DragonVsTigerSceneData) ClearUpplayerlist() {
	//清除掉所有金币不足的上庄列表
	if len(this.upplayerlist) > 0 {
		filter := []*DragonVsTigerPlayerData{}
		for _, pp := range this.upplayerlist {
			if (pp.Coin) < int64(this.DbGameFree.GetBanker()) {
				continue
			}
			if this.players[pp.SnId] == nil {
				continue
			}
			filter = append(filter, pp)
		}
		this.upplayerlist = filter
	}
}

func (this *DragonVsTigerSceneData) InUpplayerlist(snid int32) bool {
	for _, pp := range this.upplayerlist {
		if pp.SnId == snid {
			return true
		}
	}
	return false
}
func (this *DragonVsTigerSceneData) GetBanker() (*DragonVsTigerPlayerData, bool) {
	var isRobotBanker bool
	var banker *DragonVsTigerPlayerData
	if this.bankerSnId != -1 {
		banker = this.players[this.bankerSnId]
		if banker != nil && banker.IsRob {
			isRobotBanker = true
		}
	} else {
		banker = nil
		isRobotBanker = true
	}
	return banker, isRobotBanker
}
func (this *DragonVsTigerSceneData) TryChangeBanker() {
	//if this.bankerSnId != -1 {
	//	logger.Logger.Tracef("当前庄家snid = %v  已过%v局", this.bankerSnId, this.upplayerCount)
	//}
	if len(this.upplayerlist) > 0 {
		this.ClearUpplayerlist()
	}

	//让当前钱不够的庄家下庄
	if this.bankerSnId != -1 {
		banker := this.players[this.bankerSnId]
		if banker != nil {
			if banker.Coin < int64(this.DbGameFree.GetBanker()) {
				this.bankerSnId = -1
				banker.Pos = DVST_OLPOS
				banker.winRecord = []int64{}
				banker.betBigRecord = []int64{}
			}
		} else {
			logger.Logger.Warnf("why have bank,but no player:%v", this.bankerSnId)
			this.bankerSnId = -1
		}
	}

	if this.upplayerCount >= DVST_BANKERNUMBERS || this.bankerSnId == -1 {
		if len(this.upplayerlist) > 0 {
			this.upplayerCount = 0
			this.bankerSnId = this.upplayerlist[0].SnId
			this.upplayerlist = this.upplayerlist[1:]
			banker := this.players[this.bankerSnId]
			if banker != nil {
				banker.Pos = DVST_BANKERPOS
				banker.winRecord = []int64{}
				banker.betBigRecord = []int64{}
				logger.Logger.Tracef("切换庄家，庄家snid=%v", banker.SnId)
			} else {
				this.bankerSnId = -1
			}
		} else {
			this.bankerSnId = -1
		}
	}

	dragonVsTigerSendSeatInfo(this.Scene, this)
}

// GetSystemOut 玩家坐庄时，统计系统（机器人）的产出
func (this *DragonVsTigerSceneData) GetSystemOut() int64 {
	return this.PlayerBankerSystemOut(this.CalcuResult())
}

func (this *DragonVsTigerSceneData) PlayerBankerSystemOut(result int) int64 {
	banker, ok := this.players[this.bankerSnId]
	if !ok || banker.IsRobot() {
		return 0
	}

	sysOutCoin := 0
	systemOutCoin := [DVST_ZONE_MAX]int{0, 0, 0}

	for key, value := range this.betInfoRob {
		systemOutCoin[key] += value
	}

	for i := 0; i < DVST_ZONE_MAX; i++ {
		if result == i {
			sysOutCoin += systemOutCoin[i] * DVST_TheOdds[i]
		} else {
			sysOutCoin -= systemOutCoin[i]
		}
	}

	if result == DVST_ZONE_DRAW {
		if systemOutCoin[DVST_ZONE_DRAGON] != 0 {
			sysOutCoin += systemOutCoin[DVST_ZONE_DRAGON]
		}
		if systemOutCoin[DVST_ZONE_TIGER] != 0 {
			sysOutCoin += systemOutCoin[DVST_ZONE_TIGER]
		}
	}

	//玩家输钱只输一个押注的筹码，故不需要考虑玩家不够赔付，只需考虑庄家不够赔付的情况
	//庄家输得钱超过自身拥有的
	if sysOutCoin > 0 && int64(sysOutCoin) > banker.Coin {
		sysOutCoin = int(banker.Coin)
	}
	return int64(sysOutCoin)
}

// GetSystemChangeCoin 系统赢分
// result 开奖位置
func (this *DragonVsTigerSceneData) GetSystemChangeCoin(result int) int64 {
	if this.IsRobotBanker() {
		// 系统庄
		return int64(-this.SystemBankerPlayerOut(result))
	}
	// 真人庄
	return this.PlayerBankerSystemOut(result)
}

//真人坐庄时进行调控
func (this *DragonVsTigerSceneData) BankerBalance() {
	result := this.CalcuResult()
	systemCoinOut := this.GetSystemOut()
	status, changeRate := base.CoinPoolMgr.GetCoinPoolStatus2(this.Platform, this.GetGameFreeId(), this.GroupId, systemCoinOut)
	if status == base.CoinPoolStatus_Low {
		/**
			低水位时要实现系统收分，遍历龙、虎、和三种结果，若有一种结果可实现收分，则返回；若三种都不能实现收分，则返回系统吐分最少的结果
		**/
		n := this.Rand.Intn(10000)
		if n < changeRate && systemCoinOut < 0 {
			type resultInfo struct {
				flag  int
				cards [2]int32
				out   int64
			}
			infos := make([]resultInfo, 0)
			infos = append(infos, resultInfo{
				flag:  result,
				cards: this.cards,
				out:   systemCoinOut,
			})

			if result == DVST_ZONE_DRAW { //第一个结果为和
				//模拟do-while结构,找出不为和的结果
				for {
					this.cards[0] = int32(this.poker.Next())
					if this.cards[0] == -1 {
						this.poker.Shuffle()
						this.cards[0] = int32(this.poker.Next())
					}

					result = this.CalcuResult()
					if result != DVST_ZONE_DRAW {
						break
					}
				}
				//第二种结果
				systemCoinOut = this.GetSystemOut()
				if systemCoinOut > 0 {
					return
				} else {
					infos = append(infos, resultInfo{
						flag:  result,
						cards: this.cards,
						out:   systemCoinOut,
					})
				}
				//第三种结果
				this.cards[0], this.cards[1] = this.cards[1], this.cards[0]
				result = this.CalcuResult()
				systemCoinOut = this.GetSystemOut()
				if systemCoinOut > 0 {
					return
				} else {
					infos = append(infos, resultInfo{
						flag:  result,
						cards: this.cards,
						out:   systemCoinOut,
					})
				}
			} else { //第一种结果不为和
				//第二种结果，交换两张牌
				this.cards[0], this.cards[1] = this.cards[1], this.cards[0]
				result = this.CalcuResult()
				systemCoinOut = this.GetSystemOut()
				if systemCoinOut > 0 {
					return
				} else {
					infos = append(infos, resultInfo{
						flag:  result,
						cards: this.cards,
						out:   systemCoinOut,
					})
				}

				//第三种结果，创造和
				this.cards[1] = this.cards[0] + int32(common.RandInt(1, 4)*PER_CARD_COLOR_MAX)
				if this.cards[1] > int32(POKER_CART_CNT) {
					this.cards[1] -= int32(POKER_CART_CNT)
				}
				result = this.CalcuResult()
				systemCoinOut = this.GetSystemOut()
				if systemCoinOut > 0 {
					return
				} else {
					infos = append(infos, resultInfo{
						flag:  result,
						cards: this.cards,
						out:   systemCoinOut,
					})
				}
			}

			sort.Slice(infos, func(i, j int) bool {
				return infos[i].out < infos[j].out
			})
			this.cards[0], this.cards[1] = infos[0].cards[0], infos[0].cards[1]
		}
	}
}

//判断是否有满足单控条件的玩家
func (this *DragonVsTigerSceneData) IsSingleRegulatePlayer() (bool, *DragonVsTigerPlayerData) {
	// 多个单控玩家随机选一个
	var singlePlayers []*DragonVsTigerPlayerData
	for _, p := range this.players {
		if !p.IsRob {
			data, ok := p.IsSingleAdjustPlayer()
			if data == nil || !ok {
				continue
			}

			singlePlayers = append(singlePlayers, p)
		}
	}

	if len(singlePlayers) == 0 {
		return false, nil
	}

	singlePlayer := singlePlayers[this.RandInt(len(singlePlayers))]
	data, _ := singlePlayer.IsSingleAdjustPlayer()

	//闲家，下注额不在调控范围内
	if this.bankerSnId != singlePlayer.SnId {
		if singlePlayer.betTotal == 0 || singlePlayer.betTotal > data.BetMax || singlePlayer.betTotal < data.BetMin { //总下注下限≤玩家总下注额度≤总下注上限
			return false, nil
		}
	}

	return true, singlePlayer
}

//本局结果是否需要单控
func (this *DragonVsTigerSceneData) IsNeedSingleRegulate(p *DragonVsTigerPlayerData) bool {
	data, ok := p.IsSingleAdjustPlayer()
	if data == nil || !ok {
		logger.Logger.Tracef("玩家%v 在%v场 不需要单控", p.SnId, this.GetGameFreeId())
		return false
	}

	p.result = data.Mode
	winScore := this.GetPlayerScoreByResult(p, this.CalcuResult())
	if this.bankerSnId == p.SnId { //如果是庄家
		if data.Mode == common.SingleAdjustModeLose && (winScore >= 0 || -winScore < data.BankerLoseMin) { //单控输 && （ 本局赢 || 输钱<输钱下限）
			return true
		} else if data.Mode == common.SingleAdjustModeWin && (winScore <= 0 || winScore < data.BankerWinMin) { //单控赢 && （ 本局输 || 赢钱<赢钱下限）
			return true
		}
	} else { //如果是闲家
		if p.betTotal == 0 {
			return false
		}

		if p.betTotal >= data.BetMin && p.betTotal <= data.BetMax { //总下注下限≤玩家总下注额度≤总下注上限
			if data.Mode == common.SingleAdjustModeLose && winScore >= 0 { //单控输 && 本局赢
				return true
			} else if data.Mode == common.SingleAdjustModeWin && winScore <= 0 { //单控赢 && 本局输
				return true
			}

		}
	}

	return false
}

//庄家单控
func (this *DragonVsTigerSceneData) RegulationBankerCard(p *DragonVsTigerPlayerData) bool {
	data, ok := p.IsSingleAdjustPlayer()
	if data == nil || !ok {
		return false
	}

	type resultInfo struct {
		flag  int
		cards [2]int32
		out   int64
		first bool
	}
	infos := make([]resultInfo, 0)
	result := this.CalcuResult()
	winScore := this.GetPlayerScoreByResult(p, result)
	infos = append(infos, resultInfo{
		flag:  result,
		cards: this.cards,
		out:   winScore,
		first: true,
	})

	if result == DVST_ZONE_DRAW { //第一个结果为和
		//模拟do-while结构,找出不为和的结果
		for {
			this.cards[0] = int32(this.poker.Next())
			if this.cards[0] == -1 {
				this.poker.Shuffle()
				this.cards[0] = int32(this.poker.Next())
			}

			result = this.CalcuResult()
			if result != DVST_ZONE_DRAW {
				break
			}
		}

		//第二种结果
		winScore = this.GetPlayerScoreByResult(p, result)
		if data.Mode == common.SingleAdjustModeWin && winScore > 0 && winScore >= data.BankerWinMin { //单控赢
			return true
		} else if data.Mode == common.SingleAdjustModeLose && winScore < 0 && -winScore >= data.BankerLoseMin { //单控输
			return true
		} else {
			infos = append(infos, resultInfo{
				flag:  result,
				cards: this.cards,
				out:   winScore,
				first: false,
			})
		}

		//第三种结果
		this.cards[0], this.cards[1] = this.cards[1], this.cards[0]
		result = this.CalcuResult()
		winScore = this.GetPlayerScoreByResult(p, result)
		if data.Mode == common.SingleAdjustModeWin && winScore > 0 && winScore >= data.BankerWinMin { //单控赢
			return true
		} else if data.Mode == common.SingleAdjustModeLose && winScore < 0 && -winScore >= data.BankerLoseMin { //单控输
			return true
		} else {
			infos = append(infos, resultInfo{
				flag:  result,
				cards: this.cards,
				out:   winScore,
				first: false,
			})
		}
	} else { //第一种结果不为和
		//第二种结果，交换两张牌
		this.cards[0], this.cards[1] = this.cards[1], this.cards[0]
		result = this.CalcuResult()
		winScore = this.GetPlayerScoreByResult(p, result)
		if data.Mode == common.SingleAdjustModeWin && winScore > 0 && winScore >= data.BankerWinMin { //单控赢 && 庄家赢 && 庄赢金币>=控赢下限
			return true
		} else if data.Mode == common.SingleAdjustModeLose && winScore < 0 && -winScore >= data.BankerLoseMin { //单控输 && 庄家输 && 庄输金币>=控输下限
			return true
		} else {
			infos = append(infos, resultInfo{
				flag:  result,
				cards: this.cards,
				out:   winScore,
				first: false,
			})
		}

		//第三种结果，创造和
		this.cards[1] = this.cards[0] + int32(common.RandInt(1, 4)*PER_CARD_COLOR_MAX)
		if this.cards[1] > int32(POKER_CART_CNT) {
			this.cards[1] -= int32(POKER_CART_CNT)
		}
		result = this.CalcuResult()
		winScore = this.GetPlayerScoreByResult(p, result)
		if data.Mode == common.SingleAdjustModeWin && winScore > 0 && winScore >= data.BankerWinMin { //单控赢
			return true
		} else if data.Mode == common.SingleAdjustModeLose && winScore < 0 && -winScore >= data.BankerLoseMin { //单控输
			return true
		} else {
			infos = append(infos, resultInfo{
				flag:  result,
				cards: this.cards,
				out:   winScore,
				first: false,
			})
		}
	}
	if data.Mode == common.SingleAdjustModeWin {
		sort.Slice(infos, func(i, j int) bool {
			return infos[i].out > infos[j].out
		})

		if infos[0].out <= 0 || infos[0].out < data.BankerWinMin {
			for _, in := range infos {
				if in.first {
					this.cards[0], this.cards[1] = in.cards[0], in.cards[1]
					logger.Logger.Trace("RegulationBankerCard 已调控，玩家赢取的金币小于庄赢下限，还原初始结果并不增加调控计数")
					return false
				}
			}
		}
	} else if data.Mode == common.SingleAdjustModeLose {
		sort.Slice(infos, func(i, j int) bool {
			return infos[i].out < infos[j].out
		})

		if infos[0].out >= 0 || -infos[0].out < data.BankerLoseMin {
			for _, in := range infos {
				if in.first {
					this.cards[0], this.cards[1] = in.cards[0], in.cards[1]
					logger.Logger.Trace("RegulationBankerCard 已调控，玩家赢取的金币小于庄输下限，还原初始结果并不增加调控计数")
					return false
				}
			}
		}
	}

	logger.Logger.Trace("RegulationBankerCard 已调控，增加调控计数")
	return true
}

//闲家单控
func (this *DragonVsTigerSceneData) RegulationXianCard(p *DragonVsTigerPlayerData) bool {
	data, ok := p.IsSingleAdjustPlayer()
	if data == nil || !ok {
		return false
	}

	type resultInfo struct {
		flag  int
		cards [2]int32
		out   int64
	}
	infos := make([]resultInfo, 0)
	result := this.CalcuResult()
	winScore := this.GetPlayerScoreByResult(p, result)
	infos = append(infos, resultInfo{
		flag:  result,
		cards: this.cards,
		out:   winScore,
	})

	if result == DVST_ZONE_DRAW { //第一个结果为和
		//模拟do-while结构,找出不为和的结果
		for {
			this.cards[0] = int32(this.poker.Next())
			if this.cards[0] == -1 {
				this.poker.Shuffle()
				this.cards[0] = int32(this.poker.Next())
			}

			result = this.CalcuResult()
			if result != DVST_ZONE_DRAW {
				break
			}
		}

		//第二种结果
		winScore = this.GetPlayerScoreByResult(p, result)
		if data.Mode == common.SingleAdjustModeWin && winScore > 0 { //单控赢
			return true
		} else if data.Mode == common.SingleAdjustModeLose && winScore < 0 { //单控输
			return true
		} else {
			infos = append(infos, resultInfo{
				flag:  result,
				cards: this.cards,
				out:   winScore,
			})
		}

		//第三种结果
		this.cards[0], this.cards[1] = this.cards[1], this.cards[0]
		result = this.CalcuResult()
		winScore = this.GetPlayerScoreByResult(p, result)
		if data.Mode == common.SingleAdjustModeWin && winScore > 0 { //单控赢
			return true
		} else if data.Mode == common.SingleAdjustModeLose && winScore < 0 { //单控输
			return true
		} else {
			infos = append(infos, resultInfo{
				flag:  result,
				cards: this.cards,
				out:   winScore,
			})
		}
	} else { //第一种结果不为和
		//第二种结果，交换两张牌
		this.cards[0], this.cards[1] = this.cards[1], this.cards[0]
		result = this.CalcuResult()
		winScore = this.GetPlayerScoreByResult(p, result)
		if data.Mode == common.SingleAdjustModeWin && winScore > 0 { //单控赢
			return true
		} else if data.Mode == common.SingleAdjustModeLose && winScore < 0 { //单控输
			return true
		} else {
			infos = append(infos, resultInfo{
				flag:  result,
				cards: this.cards,
				out:   winScore,
			})
		}

		//第三种结果，创造和
		this.cards[1] = this.cards[0] + int32(common.RandInt(1, 4)*PER_CARD_COLOR_MAX)
		if this.cards[1] > int32(POKER_CART_CNT) {
			this.cards[1] -= int32(POKER_CART_CNT)
		}
		result = this.CalcuResult()
		winScore = this.GetPlayerScoreByResult(p, result)
		if data.Mode == common.SingleAdjustModeWin && winScore > 0 { //单控赢
			return true
		} else if data.Mode == common.SingleAdjustModeLose && winScore < 0 { //单控输
			return true
		} else {
			infos = append(infos, resultInfo{
				flag:  result,
				cards: this.cards,
				out:   winScore,
			})
		}
	}
	if data.Mode == common.SingleAdjustModeWin {
		sort.Slice(infos, func(i, j int) bool {
			return infos[i].out > infos[j].out
		})
		if infos[0].out <= 0 {
			return false
		}
	} else if data.Mode == common.SingleAdjustModeLose {
		sort.Slice(infos, func(i, j int) bool {
			return infos[i].out < infos[j].out
		})
		if infos[0].out >= 0 {
			return false
		}
	}

	this.cards[0], this.cards[1] = infos[0].cards[0], infos[0].cards[1]
	logger.Logger.Trace("RegulationXianCard 已调控，增加调控计数")
	return true
}

//发送庄家列表数据给机器人
func (this *DragonVsTigerSceneData) SendRobotUpBankerList() {
	pack := &dragonvstiger.SCDragonVsTiggerUpList{
		Count: proto.Int(len(this.upplayerlist)),
	}
	this.RobotBroadcast(int(dragonvstiger.DragonVsTigerPacketID_PACKET_SC_DVST_UPLIST), pack)
}

func (this *DragonVsTigerSceneData) CheckResults() {
	if this.bIntervention {
		return
	}
	if this.GetResult() == common.DefaultResult {
		this.bIntervention = false
		return
	}
	// 这里和TryDTDraw方法的参数规则对应一下
	n := this.GetResult()
	switch n {
	case 1, 2: // 龙赢，虎赢
	case 3: // 和
		n = 0
	default:
		this.bIntervention = false
		return
	}
	this.cards[0], this.cards[1] = this.poker.TryDTDraw(int32(n))
	this.bIntervention = true
	//this.webUser = this.Scene.webUser
}

// Bet 有玩家上庄时检查下注
// b 庄家
// p 下注的玩家
// pos 下注位置
// betCoin 下注金额
func (this *DragonVsTigerSceneData) Bet(b, p *DragonVsTigerPlayerData, pos, betCoin int) int {
	// 计算如果当前位置赢，是否庄家够赔
	var maxBet int
	if pos == DVST_ZONE_DRAW {
		maxBet = int(b.Coin) - this.betInfo[pos]*DVST_TheOdds[pos]
	} else {
		bet := this.betInfo[DVST_ZONE_TIGER]
		if pos == DVST_ZONE_TIGER {
			bet = this.betInfo[DVST_ZONE_DRAGON]
		}
		maxBet = int(b.Coin) + this.betInfo[DVST_ZONE_DRAW] + bet - this.betInfo[pos]*DVST_TheOdds[pos]
	}
	maxBet /= DVST_TheOdds[pos]
	if betCoin <= maxBet {
		return betCoin
	}
	minChip := int(this.DbGameFree.GetOtherIntParams()[0])
	return maxBet / minChip * minChip
}

// BetStop 是否所有位置都已经不能加注
func (this *DragonVsTigerSceneData) BetStop() bool {
	if this.bankerSnId == -1 {
		return false
	}
	minChip := int(this.DbGameFree.GetOtherIntParams()[0])
	b := this.players[this.bankerSnId]
	// 和还可以压
	if int(b.Coin)-this.betInfo[DVST_ZONE_DRAW]*DVST_TheOdds[DVST_ZONE_DRAW] >=
		minChip*DVST_TheOdds[DVST_ZONE_DRAW] {
		return false
	}
	// 龙还可以压
	if int(b.Coin)+this.betInfo[DVST_ZONE_TIGER]+this.betInfo[DVST_ZONE_DRAW] >=
		(this.betInfo[DVST_ZONE_DRAGON]+minChip)*DVST_TheOdds[DVST_ZONE_DRAGON] {
		return false
	}
	// 虎还可以压
	if int(b.Coin)+this.betInfo[DVST_ZONE_DRAGON]+this.betInfo[DVST_ZONE_DRAW] >=
		(this.betInfo[DVST_ZONE_TIGER]+minChip)*DVST_TheOdds[DVST_ZONE_TIGER] {
		return false
	}
	return true
}

// GetPlayerScoreByResult 开某个位置玩家赢分
func (this *DragonVsTigerSceneData) GetPlayerScoreByResult(p *DragonVsTigerPlayerData, result int) int64 {
	// 因为下注阶段已经做了判断，庄家是否够赔，这里就不用考虑了
	var winScore int64
	if !this.IsRobotBanker() && p.SnId == this.bankerSnId {
		for _, v := range this.seats {
			if v == nil || v.betTotal <= 0 {
				continue
			}
			winScore -= int64(v.betInfo[result] * DVST_TheOdds[result])
			if result != DVST_ZONE_DRAW {
				winScore += v.betTotal - int64(v.betInfo[result])
			}
		}
		if winScore > p.Coin {
			winScore = p.Coin
		}
		if winScore < 0 && -winScore > p.Coin {
			winScore = -p.Coin
		}
		return winScore
	}
	winScore = int64(p.betInfo[result] * DVST_TheOdds[result])
	if result != DVST_ZONE_DRAW {
		winScore -= p.betTotal - int64(p.betInfo[result])
	}
	return winScore
}
func (this *DragonVsTigerSceneData) SendCards() {
	this.poker.Shuffle()
	for i := 0; i < 2; i++ {
		this.cards[i] = int32(this.poker.Next())
	}
	//logger.Logger.Trace("初始结果......", this.cards)
	dif := math.Abs(float64(this.cards[0]%13 - this.cards[1]%13))
	if dif == 1 && rand.Intn(100) < 60 {
		if this.cards[0]%13 > this.cards[1]%13 {
			this.cards[0] = rand.Int31n(7) + 6 + rand.Int31n(4)*13
			this.cards[1] = rand.Int31n(5) + rand.Int31n(4)*13
		} else {
			this.cards[0] = rand.Int31n(5) + rand.Int31n(4)*13
			this.cards[1] = rand.Int31n(7) + 6 + rand.Int31n(4)*13
		}
		//logger.Logger.Trace("调控后结果结果......", this.cards)
	}
}
