package baccarat

import (
	//"encoding/json"
	"fmt"
	//bjl "games.yol.com/win88/api3th/smart/baccarat"
	"games.yol.com/win88/common"
	"games.yol.com/win88/gamerule/baccarat"
	"games.yol.com/win88/gamesrv/base"
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	proto_baccarat "games.yol.com/win88/protocol/baccarat"
	//"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/logger"
	//"github.com/idealeak/goserver/core/task"
	"github.com/idealeak/goserver/core/timer"
	"math/rand"
	"sort"
	"strconv"
)

var BACCARAT_TheOdds map[int]float64
var BaccaratZoneMap map[int]bool

//百家乐场景数据
type BaccaratSceneData struct {
	*base.Scene                                         //房间信息
	poker           *baccarat.Poker                     //牌
	cards           [6]int32                            //记录庄闲两个位置的牌的信息
	cardsAndPoint   []int32                             //有牌信息和庄闲的点数和最后的输赢区域
	players         map[int32]*BaccaratPlayerData       //玩家信息
	winZone         int                                 //赢的区域，可能是多个，使用位运算
	seats           []*BaccaratPlayerData               //座位信息(富豪倒序)
	winTop1         *BaccaratPlayerData                 //神算子
	betTop5         [5]*BaccaratPlayerData              //押注最多的5位玩家
	betInfo         map[int]int                         //记录下注位置及筹码
	betInfoRob      map[int]int                         //记录机器人下注位置及筹码
	betDetailInfo   map[int]map[int]int                 //记录下注位置的详细筹码数据
	trend100Cur     []int32                             //最近全部走势
	trend20Lately   []int32                             //最近6局龙虎走势
	by              func(p, q *BaccaratPlayerData) bool //排序函数
	hRunRecord      timer.TimerHandle                   //每十分钟记录一次数据
	hBatchSend      timer.TimerHandle                   //批量发送筹码数据
	constSeatKey    string                              //固定座位的唯一码
	bankerList      []*BaccaratPlayerData               // 上庄玩家列表
	bankerTimes     int32                               // 当前上庄玩家已在庄次数
	bankerSnId      int32                               // 当前庄家位置
	resultAggregate map[int]int                         //开牌结果集
	logicId         string
	bIntervention   bool
	webUser         string
	BJLSmart
}

type BJLSmart struct {
	isSmartOperation bool
	smartSuccess     bool
	stop             bool
	//dealRes          *bjl.DealResponse
}

func (this *BaccaratSceneData) GetNextPoker() int32 {
	c, flag := this.poker.Next()
	if flag {
		this.trend100Cur = make([]int32, 0)
		this.trend20Lately = make([]int32, 0)
	}
	return c
}

//Len()
func (s *BaccaratSceneData) Len() int {
	return len(s.seats)
}

//Less():输赢记录将有高到底排序
func (s *BaccaratSceneData) Less(i, j int) bool {
	return s.by(s.seats[i], s.seats[j])
}

//Swap()
func (s *BaccaratSceneData) Swap(i, j int) {
	s.seats[i], s.seats[j] = s.seats[j], s.seats[i]
}

func NewBaccaratSceneData(s *base.Scene) *BaccaratSceneData {
	return &BaccaratSceneData{
		Scene:   s,
		players: make(map[int32]*BaccaratPlayerData),
		cards:   [6]int32{-1, -1, -1, -1, -1, -1},
	}
}

func (this *BaccaratSceneData) RebindPlayerSnId(oldSnId, newSnId int32) {
	if p, exist := this.players[oldSnId]; exist {
		delete(this.players, oldSnId)
		this.players[newSnId] = p
	}
}

func (this *BaccaratSceneData) init() bool {
	this.WithLocalAI = true
	this.BindAIMgr(&BaccaratSceneAIMgr{})
	//存储投注区域
	BaccaratZoneMap = make(map[int]bool)
	BaccaratZoneMap[baccarat.BACCARAT_ZONE_TIE] = true
	BaccaratZoneMap[baccarat.BACCARAT_ZONE_BANKER] = true
	BaccaratZoneMap[baccarat.BACCARAT_ZONE_PLAYER] = true
	BaccaratZoneMap[baccarat.BACCARAT_ZONE_BANKER_DOUBLE] = true
	BaccaratZoneMap[baccarat.BACCARAT_ZONE_PLAYER_DOUBLE] = true

	//存储倍率
	BACCARAT_TheOdds = make(map[int]float64)
	BACCARAT_TheOdds[baccarat.BACCARAT_ZONE_TIE] = 9.0
	BACCARAT_TheOdds[baccarat.BACCARAT_ZONE_BANKER] = 2.0
	BACCARAT_TheOdds[baccarat.BACCARAT_ZONE_PLAYER] = 2.0
	BACCARAT_TheOdds[baccarat.BACCARAT_ZONE_BANKER_DOUBLE] = 12.0
	BACCARAT_TheOdds[baccarat.BACCARAT_ZONE_PLAYER_DOUBLE] = 12.0

	this.bankerSnId = -1
	if this.DbGameFree == nil {
		return false
	}
	this.poker = baccarat.NewPoker()
	//将路单数据清空
	this.trend100Cur = make([]int32, 0)
	this.trend20Lately = make([]int32, 0)
	this.logicId, _ = model.AutoIncGameLogId()
	return true
}
func (this *BaccaratSceneData) CleanAll() {
	this.Clean()
	//投注信息的清理
	this.betInfo = make(map[int]int)
	this.betInfoRob = make(map[int]int)
	this.betDetailInfo = make(map[int]map[int]int)
	for k, _ := range BaccaratZoneMap {
		this.betDetailInfo[k] = make(map[int]int)
	}
	for _, p := range this.players {
		p.Clean()
	}
	this.bIntervention = false
	this.webUser = ""
	this.ResetSmart()
	//重置水池调控标记
	this.SetCpControlled(false)
	this.SetSystemCoinOut(0)
}
func (this *BaccaratSceneData) Clean() {
	//发的牌清理
	for i := 0; i < len(this.cards); i++ {
		this.cards[i] = -1
	}
	this.cardsAndPoint = make([]int32, 0)
	//赢的位置清理
	this.winZone = 0
	//重洗牌
	if this.poker.Count() < 6 {
		this.poker = baccarat.NewPoker()
		//将路单数据清空
		this.trend100Cur = make([]int32, 0)
		this.trend20Lately = make([]int32, 0)
	}
	this.SetSystemCoinOut(0)
}

func (this *BaccaratSceneData) CanStart() bool {
	cnt := len(this.players)
	if cnt > 0 {
		return true
	}
	return false
}

func (this *BaccaratSceneData) delPlayer(p *base.Player) {
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

func (this *BaccaratSceneData) OnPlayerEnter(p *base.Player) {
	if p.Coin < int64(this.DbGameFree.GetBetLimit()) || p.Coin < int64(this.DbGameFree.GetOtherIntParams()[0]) {
		p.MarkFlag(base.PlayerState_GameBreak)
		p.SyncFlag()
	}
}

func (this *BaccaratSceneData) OnPlayerLeave(p *base.Player, reason int) {
	this.delPlayer(p)
	this.BroadcastPlayerLeave(p, reason)
}

func (this *BaccaratSceneData) BroadcastPlayerLeave(p *base.Player, reason int) {

}

func (this *BaccaratSceneData) SceneDestroy(force bool) {
	//销毁房间
	this.Scene.Destroy(force)
}

//最大赢取倒序
func BACCARAT_OrderByLatest20Win_P(l, r *BaccaratPlayerData) bool {
	return l.lately20Win > r.lately20Win
}

//最大押注倒序
func BACCARAT_OrderByLatest20Bet_P(l, r *BaccaratPlayerData) bool {
	return l.lately20Bet > r.lately20Bet
}

func (this *BaccaratSceneData) Resort() string {
	constSeatKey := fmt.Sprintf("%v", this.bankerSnId)
	best := int64(0)
	count := len(this.seats)
	this.winTop1 = nil
	for i := 0; i < count; i++ {
		playerEx := this.seats[i]
		if playerEx != nil && playerEx.SnId != this.bankerSnId {
			playerEx.Pos = BACCARAT_OLPOS
			if playerEx.lately20Win > best {
				best = playerEx.lately20Win
				this.winTop1 = playerEx
			}
		}
	}
	if this.winTop1 != nil {
		this.winTop1.Pos = BACCARAT_BESTWINPOS
		constSeatKey = fmt.Sprintf("%d_%d", this.winTop1.SnId, this.winTop1.Pos)
	}

	this.by = BACCARAT_OrderByLatest20Bet_P
	sort.Sort(this)
	cnt := 0
	for i := 0; i < BACCARAT_RICHTOP5; i++ {
		this.betTop5[i] = nil
	}
	for i := 0; i < count && cnt < BACCARAT_RICHTOP5; i++ {
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

func (this *BaccaratSceneData) PushTrend(result int32) {
	this.trend100Cur = append(this.trend100Cur, result)
	cnt := len(this.trend100Cur)
	if cnt > 100 {
		this.trend100Cur = this.trend100Cur[cnt-100:]
	}
	this.trend20Lately = append(this.trend20Lately, result)
	cnt = len(this.trend20Lately)
	if cnt > 20 {
		this.trend20Lately = this.trend20Lately[cnt-20:]
	}
}

func (this *BaccaratSceneData) GetRecoveryCoe() float64 {
	id := this.GetGameFreeId()
	if coe, ok := model.GameParamData.HundredSceneRecoveryCoe[strconv.Itoa(int(id))]; ok {
		return float64(coe) / 100
	}
	return 0.05
}

//单控计算真人输赢
func (this *BaccaratSceneData) SingleRealPlayerCoinOut() int64 {
	if this.bankerSnId != -1 {
		banker := this.players[this.bankerSnId]
		if banker != nil && !banker.IsRob {
			return this.SingleRealPlayerCoinOutBanker()
		}
	}

	realOutCoin := 0
	systemOutCoin := make(map[int]int)

	//计算每个区域玩家下注数，不算机器人
	for key, value := range this.betInfo {
		systemOutCoin[key] += value - this.betInfoRob[key]
	}

	//计算赢位置的赢钱数
	for k := range systemOutCoin {
		if this.winZone&k == k {
			wincoin := int(float64(systemOutCoin[k]) * (BACCARAT_TheOdds[k] - 1))
			tax := int(float64(wincoin) * float64(this.DbGameFree.GetTaxRate()) / 10000)
			realOutCoin += wincoin - tax
		} else {
			//开和
			if baccarat.BACCARAT_ZONE_TIE&this.winZone == baccarat.BACCARAT_ZONE_TIE &&
				(k == baccarat.BACCARAT_ZONE_PLAYER || k == baccarat.BACCARAT_ZONE_BANKER) {
				//开和 玩家输赢不算 庄 闲区输赢
				continue
			} else {
				realOutCoin -= systemOutCoin[k]
			}
		}
	}
	return int64(realOutCoin)
}

//真人庄家单控计算输赢
func (this *BaccaratSceneData) SingleRealPlayerCoinOutBanker() int64 {
	if this.bankerSnId == -1 {
		return 0
	}
	banker := this.players[this.bankerSnId]
	if banker == nil {
		return 0
	}

	realOutCoin := 0
	systemOutCoin := map[int]int{}
	for key, value := range this.betInfoRob {
		systemOutCoin[key] += value
	}
	//计算赢位置的赢钱数
	for k := range systemOutCoin {
		if this.winZone&k == k {
			realOutCoin -= int(float64(systemOutCoin[k]) * (BACCARAT_TheOdds[k] - 1))
		} else {
			//开和
			if baccarat.BACCARAT_ZONE_TIE&this.winZone == baccarat.BACCARAT_ZONE_TIE &&
				(k == baccarat.BACCARAT_ZONE_PLAYER || k == baccarat.BACCARAT_ZONE_BANKER) {
				//开和 玩家输赢不算 庄 闲区输赢
				continue
			} else {
				tax := int(float64(systemOutCoin[k]) * float64(this.DbGameFree.GetTaxRate()) / 10000)
				realOutCoin += systemOutCoin[k] - tax
			}
		}
	}
	//庄家输得钱超过自身拥有的
	if realOutCoin < 0 && int64(-realOutCoin) > banker.Coin {
		realOutCoin = int(banker.Coin)
	}
	return int64(realOutCoin)
}

//系统输赢计算 > 0 系统赢钱 <0 系统亏钱
func (this *BaccaratSceneData) SystemCoinOutByWinZone() int64 {
	if this.bankerSnId != -1 {
		banker := this.players[this.bankerSnId]
		if banker != nil && !banker.IsRob {
			return this.SystemCoinOutByPlayerBanker()
		}
	}

	sysOutCoin := 0
	systemOutCoin := make(map[int]int)

	//计算每个区域玩家下注数，不算机器人
	for key, value := range this.betInfo {
		systemOutCoin[key] += value - this.betInfoRob[key]
	}

	//计算赢位置的赢钱数
	for k := range systemOutCoin {
		if this.winZone&k == k {
			sysOutCoin += int(float64(systemOutCoin[k]) * (BACCARAT_TheOdds[k]))
			sysOutCoin -= systemOutCoin[k]
		} else {
			//开和
			if baccarat.BACCARAT_ZONE_TIE&this.winZone == baccarat.BACCARAT_ZONE_TIE &&
				(k == baccarat.BACCARAT_ZONE_PLAYER || k == baccarat.BACCARAT_ZONE_BANKER) {
				//开和 系统输赢不算 庄 闲区输赢
				continue
			} else {
				sysOutCoin -= systemOutCoin[k]
			}

		}
	}
	return int64(-sysOutCoin)
}

func (this *BaccaratSceneData) SystemCoinOutByPlayerBanker() int64 {
	if this.bankerSnId == -1 {
		return 0
	}
	banker := this.players[this.bankerSnId]
	if banker == nil {
		return 0
	}
	sysOutCoin := int64(0)
	systemOutCoin := map[int]int64{}
	for key, value := range this.betInfoRob {
		systemOutCoin[key] += int64(value)
	}
	//计算赢位置的赢钱数
	for k := range systemOutCoin {
		if this.winZone&k == k {
			sysOutCoin += int64(float64(systemOutCoin[k]) * (BACCARAT_TheOdds[k] - 1))
		} else {
			//开和
			if baccarat.BACCARAT_ZONE_TIE&this.winZone == baccarat.BACCARAT_ZONE_TIE &&
				(k == baccarat.BACCARAT_ZONE_PLAYER || k == baccarat.BACCARAT_ZONE_BANKER) {
				//开和 玩家输赢不算 庄 闲区输赢
				continue
			} else {
				sysOutCoin -= systemOutCoin[k]
			}
		}
	}
	//玩家输钱只输一个押注的筹码，故不需要考虑玩家不够赔付，只需考虑庄家不够赔付的情况
	//庄家输得钱超过自身拥有的
	if sysOutCoin > 0 && sysOutCoin > banker.Coin {
		sysOutCoin = banker.Coin
	}
	return sysOutCoin
}

//根据投注信息选择相应的牌
func (this *BaccaratSceneData) tryTest_1() bool {
	out := this.SystemCoinOutByWinZone()
	logger.Trace("out:", out)

	//赔付<投注不干预
	if out <= 0 {
		return true
	}

	//赔付的金币超过赔付的最大干预
	setting := base.CoinPoolMgr.GetCoinPoolSetting(this.Platform, this.GetGameFreeId(), this.GroupId)
	logger.Trace("setting.GetMaxOutValue()=", setting.GetMaxOutValue())
	if out > int64(setting.GetMaxOutValue()) {
		return false
	}
	_, changeRate := base.CoinPoolMgr.GetCoinPoolStatus(this.Platform, this.GetGameFreeId(), this.GroupId)
	curPoolCoin := base.CoinPoolMgr.LoadCoin(this.GetGameFreeId(), this.Platform, this.GroupId)
	afterPayPoolCoin := curPoolCoin - out
	logger.Trace("curPoolCoin:", curPoolCoin)

	if afterPayPoolCoin > int64(setting.GetLowerLimit()) {
		return true
	}
	//再给你一次机会
	if common.RandInt(10000) < changeRate && this.CoinPoolCanOut() {
		return true
	}
	return false
}

//尝试计算是否满足要求
func (this *BaccaratSceneData) tryTest_2(out int64) bool {
	if out == 0 {
		//系统不输不赢不处理 有可能没有真人
		return true
	}
	//setting := base.CoinPoolMgr.GetCoinPoolSetting(this.Platform, this.GetGameFreeId(), this.GroupId)

	if out < 0 {
		if !base.CoinPoolMgr.IsMaxOutHaveEnough(this.Platform, this.GetGameFreeId(), this.GroupId, out) {
			return false
		}
	}

	status, changeRate := base.CoinPoolMgr.GetCoinPoolStatus2(this.Platform, this.GetGameFreeId(), this.GroupId, out)

	switch status {
	case base.CoinPoolStatus_Normal:
		return true
	case base.CoinPoolStatus_Low: //库存值 < 库存下限
		if out < 0 {
			if common.RandInt(10000) < changeRate {
				return false
			} else {
				return true
			}
		} else {
			return true
		}
	case base.CoinPoolStatus_High: //库存上限 < 库存值 < 库存上限+偏移量
		if out < 0 {
			return true
		} else {
			if common.RandInt(10000) < changeRate && this.CoinPoolCanOut() {
				return true
			} else {
				return false
			}
		}
	case base.CoinPoolStatus_TooHigh: //库存>库存上限+偏移量
		if out < 0 {
			return true
		} else {
			if common.RandInt(10000) < changeRate &&
				this.CoinPoolCanOut() {
				return true
			} else {
				return false
			}
		}
	default: //base.CoinPoolStatus_Normal//库存下限 < 库存值 < 库存上限
		//不处理
		return true
	}

}

//发牌
func (this *BaccaratSceneData) preAnalysis() {
	ischange := false //是否换过牌
	systemOut := make(map[int64][6]int32)
	for i := 0; i < 10; i++ {
		this.tryShuffle()         //尝试发牌
		this.calculationWinZone() //计算输赢区域
		out := this.SystemCoinOutByWinZone()
		systemOut[out] = this.cards
		ok := this.tryTest_2(out) //尝试计算是否满足要求
		if !ok {
			this.SetCpControlled(true)
			if !ischange {
				ischange = true
				//换牌
				bank := make([]int32, 3)
				copy(bank, this.cards[3:])
				copy(this.cards[3:], this.cards[:3]) //闲给庄
				copy(this.cards[:3], bank)           //庄给闲
				i--
			} else {
				ischange = false
				this.poker.PutIn(this.cards[:])
				this.Clean()
				if i == 9 {
					//十次随玩 发现还是系统还是亏 并且当前水池亏钱 选择系统赢钱最多的牌
					logger.Logger.Info("十次随玩 发现还是系统还是亏 并且当前水池亏钱 选择系统赢钱最多的牌")
					max := int64(0)
					for k, v := range systemOut {
						if max == 0 {
							max = k
							this.cards = v
							continue
						}
						if k > max {
							max = k
							this.cards = v
						}
					}
					//从牌堆把需要的牌拿出来
					this.poker.TakeOut(this.cards[:])
					this.calculationWinZone() //计算输赢区域
				}
			}
		} else {
			break
		}
	}
}

//尝试发牌
func (this *BaccaratSceneData) tryShuffle() {
	//在该状态时就提前计算出来要发的6张牌
	//开牌状态中Params中为10个字节，依次为闲家牌3张（没有补牌用-1表示）,闲家点,庄家牌（3张）,庄家点,比牌结果,剩余牌数
	//发牌顺序为：
	//------> 闲家第一张=cards[0]
	//------> 庄家第三张=cards[2]
	//------> 闲家第五张=cards[4]
	//------> 庄家第二张=cards[1]
	//------> 闲家第四张=cards[3]
	//------> 庄家第六张=cards[5]
	//在数值中表现为闲家牌[0][1][2]+庄家牌[3][4][5]
	this.cards = [6]int32{-1, -1, -1, -1, -1, -1}
	//if this.dealRes == nil {
	//	this.smartSuccess = false
	//}
	if this.smartSuccess {
		//for i := 0; i < 2; i++ {
		//	this.cards[i] = this.dealRes.XianCards[i]
		//	this.cards[i+3] = this.dealRes.BankerCards[i]
		//}
	} else {
		for i := 0; i < 2; i++ {
			this.cards[i] = this.GetNextPoker()
			this.cards[i+3] = this.GetNextPoker()
		}
	}

	//两张牌的点数和最终的点数
	playerLastCardPoint := baccarat.GetPointNum(this.cards[:], 0, 1)
	bankerLastCardPoint := baccarat.GetPointNum(this.cards[:], 3, 4)
	//闲家尝试补一张
	if this.poker.IsNeedPlayerAndOne(this.cards[:]) {
		if this.smartSuccess {
			//this.cards[2] = this.dealRes.XianCards[2]
		} else {
			if this.poker.Count() == 0 {
				this.Clean()
				this.preAnalysis()
				return
			}
			this.cards[2] = this.GetNextPoker()
		}
		playerLastCardPoint = baccarat.GetPointNum(this.cards[:], 0, 1, 2)
	}
	//庄家尝试补一张
	if this.poker.IsNeedBankerAndOne(this.cards[:]) {
		if this.smartSuccess {
			//this.cards[5] = this.dealRes.BankerCards[2]
		} else {
			if this.poker.Count() == 0 {
				this.Clean()
				this.preAnalysis()
				return
			}
			this.cards[5] = this.GetNextPoker()
		}
		bankerLastCardPoint = baccarat.GetPointNum(this.cards[:], 3, 4, 5)
	}
	this.winZone = 0
	if this.cards[0]%13 == this.cards[1]%13 {
		this.winZone |= baccarat.BACCARAT_ZONE_PLAYER_DOUBLE
	}
	if this.cards[3]%13 == this.cards[4]%13 {
		this.winZone |= baccarat.BACCARAT_ZONE_BANKER_DOUBLE
	}
	if playerLastCardPoint == bankerLastCardPoint {
		this.winZone |= baccarat.BACCARAT_ZONE_TIE
	} else if playerLastCardPoint > bankerLastCardPoint {
		this.winZone |= baccarat.BACCARAT_ZONE_PLAYER
	} else {
		this.winZone |= baccarat.BACCARAT_ZONE_BANKER
	}
}

//计算输赢区域
func (this *BaccaratSceneData) calculationWinZone() {
	playerLastCardPoint := baccarat.GetPointNum(this.cards[:], 0, 1, 2)
	bankerLastCardPoint := baccarat.GetPointNum(this.cards[:], 3, 4, 5)
	this.winZone = 0
	if this.cards[0]%13 == this.cards[1]%13 {
		this.winZone |= baccarat.BACCARAT_ZONE_PLAYER_DOUBLE
	}
	if this.cards[3]%13 == this.cards[4]%13 {
		this.winZone |= baccarat.BACCARAT_ZONE_BANKER_DOUBLE
	}
	if playerLastCardPoint == bankerLastCardPoint {
		this.winZone |= baccarat.BACCARAT_ZONE_TIE
	} else if playerLastCardPoint > bankerLastCardPoint {
		this.winZone |= baccarat.BACCARAT_ZONE_PLAYER
	} else {
		this.winZone |= baccarat.BACCARAT_ZONE_BANKER
	}
}

func (this *BaccaratSceneData) TryChangeBanker() {
	if len(this.bankerList) > 0 {
		this.ClearBankerList()
	}

	//让当前钱不够的庄家下庄
	if this.bankerSnId != -1 {
		banker := this.players[this.bankerSnId]
		if banker != nil {
			if banker.Coin < int64(this.DbGameFree.GetBanker()) {
				this.bankerSnId = -1
				banker.Pos = BACCARAT_OLPOS
				banker.winRecord = []int64{}
				banker.betBigRecord = []int64{}
			}
		} else {
			logger.Logger.Warnf("why have bank,but no player:%v", this.bankerSnId)
			this.bankerSnId = -1
		}
	}

	if this.bankerTimes >= BACCARAT_BANKERNUMBERS || this.bankerSnId == -1 {
		if len(this.bankerList) > 0 {
			this.bankerTimes = 0
			this.bankerSnId = this.bankerList[0].SnId
			this.bankerList = this.bankerList[1:]
			banker := this.players[this.bankerSnId]
			if banker != nil {
				banker.Pos = BACCARAT_BANKERPOS
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

	BaccaratSendSeatInfo(this.Scene, this)
}

func (this *BaccaratSceneData) ClearBankerList() {
	if len(this.bankerList) > 0 {
		filter := []*BaccaratPlayerData{}
		for _, pp := range this.bankerList {
			if (pp.Coin) < int64(this.DbGameFree.GetBanker()) {
				continue
			}
			if this.players[pp.SnId] == nil {
				continue
			}
			filter = append(filter, pp)
		}
		this.bankerList = filter
	}
}

func (this *BaccaratSceneData) InBankerList(snId int32) bool {
	for _, pp := range this.bankerList {
		if pp.SnId == snId {
			return true
		}
	}
	return false
}

func (this *BaccaratSceneData) GetBanker() (*BaccaratPlayerData, bool) {
	var isRobotBanker bool
	var banker *BaccaratPlayerData
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

func BaccaratMaxRate() int64 {
	var maxRate int64 = -1
	for k, v := range BaccaratZoneMap {
		if v {
			rate := int64(BACCARAT_TheOdds[k])
			if rate > maxRate {
				maxRate = rate
			}
		}
	}
	return maxRate - 1
}

func (this *BaccaratSceneData) BankerList() *proto_baccarat.SCBaccaratBankerList {
	pack := &proto_baccarat.SCBaccaratBankerList{
		Count: proto.Int(len(this.bankerList)),
	}
	for _, p := range this.bankerList {
		pd := &proto_baccarat.BaccaratPlayerData{
			SnId:        proto.Int32(p.SnId),
			Name:        proto.String(p.Name),
			Head:        proto.Int32(p.Head),
			Sex:         proto.Int32(p.Sex),
			Coin:        proto.Int64(p.Coin),
			Pos:         proto.Int(p.Pos),
			Flag:        proto.Int(p.GetFlag()),
			City:        proto.String(p.GetCity()),
			HeadOutLine: proto.Int32(p.HeadOutLine),
			VIP:         proto.Int32(p.VIP),
		}
		pack.Data = append(pack.Data, pd)
	}
	proto.SetDefaults(pack)
	return pack
}

//判断是否有满足单控条件的玩家
func (this *BaccaratSceneData) IsSingleRegulatePlayer() (bool, *BaccaratPlayerData) {
	// 多个单控玩家随机选一个
	var singlePlayers []*BaccaratPlayerData
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

	if this.bankerSnId != singlePlayer.SnId && singlePlayer.betTotal <= 0 {
		return false, nil
	}

	//闲家，下注额不在调控范围内
	if this.bankerSnId != singlePlayer.SnId {
		if singlePlayer.betTotal > data.BetMax || singlePlayer.betTotal < data.BetMin { //总下注下限≤玩家总下注额度≤总下注上限
			return false, nil
		}
	}

	return true, singlePlayer
}

//本局结果是否需要单控
func (this *BaccaratSceneData) IsNeedSingleRegulate(p *BaccaratPlayerData) bool {
	data, ok := p.IsSingleAdjustPlayer()
	if data == nil || !ok {
		logger.Logger.Tracef("玩家%v 在%v场 不需要单控", p.SnId, this.GetGameFreeId())
		return false
	}
	p.result = data.Mode
	if this.bankerSnId != p.SnId && p.betTotal == 0 {
		return false
	}
	return true
}

//单控
func (this *BaccaratSceneData) RegulationCard(p *BaccaratPlayerData) bool {
	data, ok := p.IsSingleAdjustPlayer()
	if data == nil || !ok {
		return false
	}
	realOut := this.SingleRealPlayerCoinOut()
	//真人输 并且调控设置为输
	if realOut < 0 && data.Mode == common.SingleAdjustModeLose {
		if this.bankerSnId == p.SnId {
			//真人坐庄 输的金额>=被输下限  并且 输的金额<=真人身上的钱(可以输)
			if -realOut >= data.BankerLoseMin && -realOut <= p.Coin {
				return true
			}
		} else if -realOut <= p.Coin {
			//真人为闲家 调控为输 直接输 并调控生效
			return true
		}
	} else if realOut > 0 && data.Mode == common.SingleAdjustModeWin {
		if this.bankerSnId == p.SnId {
			if realOut >= data.BankerWinMin {
				return true
			}
		} else {
			return true
		}
	}
	if this.poker.Count() < 6 {
		return false
	}
	resultSclice := this.BankerChoiceWinZONE(data, p)
	this.calculationWinZone() //计算输赢区域
	copyCards := this.cards   //复制一份牌
	for i := 0; i < len(resultSclice); i++ {
		isSingle := false
		singleResult := -1
		this.cards = [6]int32{-1, -1, -1, -1, -1, -1}
		if this.bankerSnId == p.SnId {
			singleResult = resultSclice[0]
		}
		if singleResult == -1 {
			//首先判断是否有符合结果的
			for _, v := range resultSclice {
				wz := this.resultAggregate[v]
				if wz == this.winZone {
					singleResult = v
					return true
				}
			}
			singleResult = this.SingleZoneRate(resultSclice)
		}
		switch singleResult {
		case BACCARAT_TIE:
			for k, v := range this.poker.GetTIE() {
				isSingle = true
				this.cards[k] = v
			}
		case BACCARAT_BANKER:
			for k, v := range this.poker.GetBankerWin() {
				isSingle = true
				this.cards[k] = v
			}
		case BACCARAT_PLAYER:
			for k, v := range this.poker.GetXianWin() {
				isSingle = true
				this.cards[k] = v
			}
		case BACCARAT_BANKER_BANKER_DOUBLE:
			for k, v := range this.poker.GetBankerAndBankerPair() {
				isSingle = true
				this.cards[k] = v
			}
		case BACCARAT_BANKER_PLAYER_DOUBLE:
			for k, v := range this.poker.GetBankerAndXianPair() {
				isSingle = true
				this.cards[k] = v
			}
		case BACCARAT_PLAYER_BANKER_DOUBLE:
			for k, v := range this.poker.GetXianAndBankerPair() {
				isSingle = true
				this.cards[k] = v
			}
		case BACCARAT_PLAYER_PLAYER_DOUBLE:
			for k, v := range this.poker.GetXianAndXianPair() {
				isSingle = true
				this.cards[k] = v
			}
		case BACCARAT_TIE_BANKER_DOUBLE:
			for k, v := range this.poker.GetTieAndBankerPair() {
				isSingle = true
				this.cards[k] = v
			}
		case BACCARAT_TIE_PLAYER_DOUBLE:
			for k, v := range this.poker.GetTieAndXianPair() {
				isSingle = true
				this.cards[k] = v
			}
		case BACCARAT_BANKER_BANKER_DOUBLE_PLAYER_DOUBLE:
			for k, v := range this.poker.GetBankerAndBankerXianPair() {
				isSingle = true
				this.cards[k] = v
			}
		case BACCARAT_PLAYER_BANKER_DOUBLE_PLAYER_DOUBLE:
			for k, v := range this.poker.GetXianAndBankerXianPair() {
				isSingle = true
				this.cards[k] = v
			}
		case BACCARAT_TIE_BANKER_DOUBLE_PLAYER_DOUBLE:
			for k, v := range this.poker.GetTieAndBankerXianPair() {
				isSingle = true
				this.cards[k] = v
			}
		}
		if isSingle {
			r, cs := this.poker.SingleRepairCard(this.cards[:])
			copy(this.cards[:], cs)
			if r {
				this.poker.PutIn(copyCards[:]) //把牌放回去
				this.calculationWinZone()      //计算输赢区域
				return true
			} else {
				this.poker.PutIn(this.cards[:]) //把牌放回去
			}
		}
		for k, v := range resultSclice {
			if singleResult == v {
				resultSclice = append(resultSclice[:k], resultSclice[k+1:]...)
				break
			}
		}
	}
	//this.poker.PutIn(copyCards[:]) //把牌放回去
	//this.preAnalysis()
	this.cards = copyCards
	this.calculationWinZone() //计算输赢区域
	return false
}

/////////////////单控结果
const (
	BACCARAT_BANKER                             int = iota //庄
	BACCARAT_TIE                                           //和
	BACCARAT_PLAYER                                        //闲
	BACCARAT_BANKER_BANKER_DOUBLE                          //庄 庄对
	BACCARAT_BANKER_PLAYER_DOUBLE                          //庄 闲对
	BACCARAT_PLAYER_BANKER_DOUBLE                          //闲 庄对
	BACCARAT_PLAYER_PLAYER_DOUBLE                          //闲 闲对
	BACCARAT_TIE_BANKER_DOUBLE                             //和 庄对
	BACCARAT_TIE_PLAYER_DOUBLE                             //和 闲对
	BACCARAT_BANKER_BANKER_DOUBLE_PLAYER_DOUBLE            //庄 庄对 闲对
	BACCARAT_PLAYER_BANKER_DOUBLE_PLAYER_DOUBLE            //闲 庄对 闲对
	BACCARAT_TIE_BANKER_DOUBLE_PLAYER_DOUBLE               //和 庄对 闲对
)

func (this *BaccaratSceneData) CountWinZONE() {
	this.resultAggregate = make(map[int]int)
	winZone := 0
	winZone |= baccarat.BACCARAT_ZONE_TIE
	this.resultAggregate[BACCARAT_TIE] = winZone
	winZone = 0
	winZone |= baccarat.BACCARAT_ZONE_BANKER
	this.resultAggregate[BACCARAT_BANKER] = winZone
	winZone = 0
	winZone |= baccarat.BACCARAT_ZONE_PLAYER
	this.resultAggregate[BACCARAT_PLAYER] = winZone
	winZone = 0
	winZone |= baccarat.BACCARAT_ZONE_BANKER
	winZone |= baccarat.BACCARAT_ZONE_BANKER_DOUBLE
	this.resultAggregate[BACCARAT_BANKER_BANKER_DOUBLE] = winZone
	winZone = 0
	winZone |= baccarat.BACCARAT_ZONE_BANKER
	winZone |= baccarat.BACCARAT_ZONE_PLAYER_DOUBLE
	this.resultAggregate[BACCARAT_BANKER_PLAYER_DOUBLE] = winZone
	winZone = 0
	winZone |= baccarat.BACCARAT_ZONE_PLAYER
	winZone |= baccarat.BACCARAT_ZONE_BANKER_DOUBLE
	this.resultAggregate[BACCARAT_PLAYER_BANKER_DOUBLE] = winZone
	winZone = 0
	winZone |= baccarat.BACCARAT_ZONE_PLAYER
	winZone |= baccarat.BACCARAT_ZONE_PLAYER_DOUBLE
	this.resultAggregate[BACCARAT_PLAYER_PLAYER_DOUBLE] = winZone
	winZone = 0
	winZone |= baccarat.BACCARAT_ZONE_TIE
	winZone |= baccarat.BACCARAT_ZONE_BANKER_DOUBLE
	this.resultAggregate[BACCARAT_TIE_BANKER_DOUBLE] = winZone
	winZone = 0
	winZone |= baccarat.BACCARAT_ZONE_TIE
	winZone |= baccarat.BACCARAT_ZONE_PLAYER_DOUBLE
	this.resultAggregate[BACCARAT_TIE_PLAYER_DOUBLE] = winZone
	winZone = 0
	winZone |= baccarat.BACCARAT_ZONE_BANKER
	winZone |= baccarat.BACCARAT_ZONE_BANKER_DOUBLE
	winZone |= baccarat.BACCARAT_ZONE_PLAYER_DOUBLE
	this.resultAggregate[BACCARAT_BANKER_BANKER_DOUBLE_PLAYER_DOUBLE] = winZone
	winZone = 0
	winZone |= baccarat.BACCARAT_ZONE_PLAYER
	winZone |= baccarat.BACCARAT_ZONE_BANKER_DOUBLE
	winZone |= baccarat.BACCARAT_ZONE_PLAYER_DOUBLE
	this.resultAggregate[BACCARAT_PLAYER_BANKER_DOUBLE_PLAYER_DOUBLE] = winZone
	winZone = 0
	winZone |= baccarat.BACCARAT_ZONE_TIE
	winZone |= baccarat.BACCARAT_ZONE_BANKER_DOUBLE
	winZone |= baccarat.BACCARAT_ZONE_PLAYER_DOUBLE
	this.resultAggregate[BACCARAT_TIE_BANKER_DOUBLE_PLAYER_DOUBLE] = winZone
}
func (this *BaccaratSceneData) BankerChoiceWinZONE(data *model.PlayerSingleAdjust, p *BaccaratPlayerData) (resultSclice []int) {
	this.CountWinZONE()
	type winOrLose struct {
		id      int
		winCoin int64
	}
	winorloseZoneSclice := make([]winOrLose, 0)
	for k, wz := range this.resultAggregate {
		this.winZone = wz
		realOut := this.SingleRealPlayerCoinOut()
		//真人输 并且调控设置为输
		if realOut < 0 && data.Mode == common.SingleAdjustModeLose {
			if this.bankerSnId == p.SnId {
				//真人坐庄 输的金额>=被输下限  并且 输的金额<=真人身上的钱(可以输)
				if -realOut >= data.BankerLoseMin && -realOut <= p.Coin {
					winorloseZoneSclice = append(winorloseZoneSclice, winOrLose{k, -realOut})
				}
			} else if -realOut <= p.Coin {
				//真人为闲家 调控为输 直接输 并调控生效
				winorloseZoneSclice = append(winorloseZoneSclice, winOrLose{k, -realOut})
			}
		} else if realOut > 0 && data.Mode == common.SingleAdjustModeWin {
			if this.bankerSnId == p.SnId {
				if realOut >= data.BankerWinMin {
					winorloseZoneSclice = append(winorloseZoneSclice, winOrLose{k, realOut})
				}
			} else {
				winorloseZoneSclice = append(winorloseZoneSclice, winOrLose{k, realOut})
			}
		}
	}
	resultSclice = make([]int, 0)
	if len(winorloseZoneSclice) > 0 {
		if this.bankerSnId == p.SnId {
			//真人是庄家 根据输赢额 从小到大排序 优先选输赢最小值
			sort.Slice(winorloseZoneSclice, func(i, j int) bool {
				if winorloseZoneSclice[i].winCoin < winorloseZoneSclice[j].winCoin {
					return true
				}
				return false
			})
		}
		for _, v := range winorloseZoneSclice {
			resultSclice = append(resultSclice, v.id)
		}
	}
	return
}
func (this *BaccaratSceneData) SingleZoneRate(resultSclice []int) int {
	zoneRate := []int{2000, 500, 2000, 1050, 1050, 1050, 1050, 500, 500, 100, 100, 100}
	res := make([]int, 0)
	for _, v := range resultSclice {
		res = append(res, zoneRate[v])
	}
	idx := common.RandInSliceIndex(res)
	if idx < len(resultSclice) {
		return resultSclice[idx]
	}
	return resultSclice[rand.Intn(len(resultSclice))]
}

//发送庄家列表数据给机器人
func (this *BaccaratSceneData) SendRobotUpBankerList() {
	pack := &proto_baccarat.SCBaccaratBankerList{
		Count: proto.Int(len(this.bankerList)),
	}
	this.RobotBroadcast(int(proto_baccarat.BaccaratPacketID_PACKET_SC_BACCARAT_BANKERLIST), pack)
}

func (this *BaccaratSceneData) ResetSmart() {
	this.isSmartOperation = false
	this.smartSuccess = false
	this.stop = false
}

func (this *BaccaratSceneData) IsSmartOperation() bool {
	if this.Testing || this.IsPrivateScene() {
		return false
	}
	// 都是机器人不启用
	//if this.GetRealPlayerCnt() <= 0 {
	//	return false
	//}
	//// 接口开关
	//if !bjl.ConfigV1.Switch() {
	//	return false
	//}
	//// 平台开关
	//if !model.GameParamData.SmartBJL.SwitchHundred(this.Platform) {
	//	return false
	//}
	// 非正常水池不走智能化运营
	state, _ := base.CoinPoolMgr.GetCoinPoolStatus(this.Platform, this.GetGameFreeId(), this.GroupId)
	if state != base.CoinPoolStatus_Normal {
		return false
	}
	this.isSmartOperation = true
	this.smartSuccess = true
	logger.Logger.Trace("--> BJL SmartOperation")
	return true
}

func BJLPlace(n int) int {
	switch n {
	case baccarat.BACCARAT_ZONE_BANKER:
		return 0
	case baccarat.BACCARAT_ZONE_PLAYER:
		return 1
	case baccarat.BACCARAT_ZONE_BANKER_DOUBLE:
		return 2
	case baccarat.BACCARAT_ZONE_PLAYER_DOUBLE:
		return 3
	case baccarat.BACCARAT_ZONE_TIE:
		return 4
	default:
		return -1
	}
}

func (this *BaccaratSceneData) SmartDeal() {
	//this.dealRes = nil
	//setting := base.CoinPoolMgr.GetCoinPoolSetting(this.Platform, this.GetGameFreeId(), this.GroupId)
	//chips := base.CoinPoolMgr.LoadCoin(this.GetGameFreeId(), this.Platform, this.GroupId)

	bets := make([]int, 5)
	blacklist := make([]int, 5)
	for _, v := range this.players {
		if v == nil || v.IsRobot() {
			continue
		}
		for i, bet := range v.betInfo {
			j := BJLPlace(i)
			if j < 0 {
				this.smartSuccess = false
				return
			}
			// bets
			bets[j] += bet
			// blacklist
			if blacklist[j] == 0 && bet > 0 && v.BlackLevel > 0 {
				blacklist[j] = 1
			}
		}
	}

	history := []int{}
	records := []int32{}
	if len(this.trend100Cur) > 100 {
		records = this.trend100Cur[len(this.trend100Cur)-100:]
	} else {
		records = this.trend100Cur
	}
	for _, v := range records {
		if int(v)&baccarat.BACCARAT_ZONE_BANKER > 0 {
			history = append(history, 1)
		} else if int(v)&baccarat.BACCARAT_ZONE_PLAYER > 0 {
			history = append(history, -1)
		} else {
			history = append(history, 0)
		}
	}

	//var err error
	//var res []byte
	//task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
	//	res, err = bjl.ConfigV1.Do(bjl.Deal, &bjl.DealRequest{
	//		LogicID:   this.logicId,
	//		Chips:     chips,
	//		History:   history,
	//		BaseScore: this.DbGameFree.GetBaseScore(),
	//		Level:     []int32{setting.GetUpperLimit(), setting.GetLowerLimit()},
	//		BlackList: blacklist,
	//		Bets:      bets,
	//	})
	//	return nil
	//}), task.CompleteNotifyWrapper(func(data interface{}, tt *task.Task) {
	//	defer func() {
	//		this.stop = false
	//	}()
	//	if err != nil || res == nil {
	//		bjl.ConfigV1.Log().Errorf("BJL Smart error: err != nil || res == nil %v", err)
	//		this.smartSuccess = false
	//		return
	//	}
	//	this.dealRes = new(bjl.DealResponse)
	//	if err = json.Unmarshal(res, this.dealRes); err != nil {
	//		bjl.ConfigV1.Log().Errorf("BJL Smart error: json.Unmarshal error: %v", err)
	//		this.smartSuccess = false
	//		return
	//	}
	//	this.tryShuffle()
	//	this.calculationWinZone()
	//}), "SmartDeal_BJL_action").Start()
}

// 生成符合结果的牌
// n: 1闲赢 2庄赢 3和
func (this *BaccaratSceneData) TryResult(n int) bool {
	for i := 0; i < 10; i++ {
		this.tryShuffle()
		this.calculationWinZone()
		switch n {
		case 1: // 闲赢
			if this.winZone == baccarat.BACCARAT_ZONE_PLAYER {
				return true
			}
			bank := make([]int32, 3)
			copy(bank, this.cards[3:])
			copy(this.cards[3:], this.cards[:3]) //闲给庄
			copy(this.cards[:3], bank)           //庄给闲
			this.calculationWinZone()
			if this.winZone == baccarat.BACCARAT_ZONE_PLAYER {
				return true
			}

		case 2: // 庄赢
			if this.winZone == baccarat.BACCARAT_ZONE_BANKER {
				return true
			}
			bank := make([]int32, 3)
			copy(bank, this.cards[3:])
			copy(this.cards[3:], this.cards[:3]) //闲给庄
			copy(this.cards[:3], bank)           //庄给闲
			this.calculationWinZone()
			if this.winZone == baccarat.BACCARAT_ZONE_BANKER {
				return true
			}

		default:
			logger.Logger.Errorf("Baccarat SetRoomResult error: no implement %v", n)
			return false
		}
	}
	logger.Logger.Error("Baccarat SetRoomResult error: fall through")
	return false
}

func (this *BaccaratSceneData) CheckResults() {
	if this.bIntervention {
		return
	}
	if this.GetResult() == common.DefaultResult {
		this.bIntervention = false
		return
	}
	if this.TryResult(this.GetResult()) {
		this.bIntervention = true
		this.webUser = this.Scene.WebUser
	}
}

// Bet 有玩家上庄时检查下注
// b 庄家
// p 下注的玩家
// pos 下注位置
// betCoin 下注金额
func (this *BaccaratSceneData) Bet(b, p *BaccaratPlayerData, pos, betCoin int) int {
	// 计算如果当前位置赢，是否庄家够赔
	var maxBet int
	switch pos {
	case baccarat.BACCARAT_ZONE_TIE:
		// 最大赔付金额 = 庄家余额 + 庄对下注 + 闲对下注 - 和位置的赢分
		maxBet = int(b.Coin) +
			this.betInfo[baccarat.BACCARAT_ZONE_PLAYER_DOUBLE] +
			this.betInfo[baccarat.BACCARAT_ZONE_BANKER_DOUBLE] - this.betInfo[pos]*
			int(BACCARAT_TheOdds[baccarat.BACCARAT_ZONE_TIE]-1)
		maxBet /= int(BACCARAT_TheOdds[baccarat.BACCARAT_ZONE_TIE] - 1)

	case baccarat.BACCARAT_ZONE_PLAYER, baccarat.BACCARAT_ZONE_PLAYER_DOUBLE:
		// 最大赔付金额 = 庄家余额 + 和下注 + 庄下注 + 庄对下注 - 闲位置赢分 - 闲对位置赢分
		maxBet = int(b.Coin) + this.betInfo[baccarat.BACCARAT_ZONE_TIE] + this.betInfo[baccarat.BACCARAT_ZONE_BANKER] +
			this.betInfo[baccarat.BACCARAT_ZONE_BANKER_DOUBLE] -
			this.betInfo[baccarat.BACCARAT_ZONE_PLAYER]*int(BACCARAT_TheOdds[baccarat.BACCARAT_ZONE_PLAYER]-1) -
			this.betInfo[baccarat.BACCARAT_ZONE_PLAYER_DOUBLE]*int(BACCARAT_TheOdds[baccarat.BACCARAT_ZONE_PLAYER_DOUBLE]-1)
		if pos == baccarat.BACCARAT_ZONE_PLAYER {
			maxBet /= int(BACCARAT_TheOdds[baccarat.BACCARAT_ZONE_PLAYER] - 1)
		} else {
			maxBet /= int(BACCARAT_TheOdds[baccarat.BACCARAT_ZONE_PLAYER_DOUBLE] - 1)
		}

	case baccarat.BACCARAT_ZONE_BANKER, baccarat.BACCARAT_ZONE_BANKER_DOUBLE:
		// 最大赔付金额 = 庄家余额 + 和下注 + 闲下注 + 闲对下注 - 庄位置赢分 - 庄对位置赢分
		maxBet = int(b.Coin) + this.betInfo[baccarat.BACCARAT_ZONE_TIE] + this.betInfo[baccarat.BACCARAT_ZONE_PLAYER] +
			this.betInfo[baccarat.BACCARAT_ZONE_PLAYER_DOUBLE] -
			this.betInfo[baccarat.BACCARAT_ZONE_BANKER]*int(BACCARAT_TheOdds[baccarat.BACCARAT_ZONE_BANKER]-1) -
			this.betInfo[baccarat.BACCARAT_ZONE_BANKER_DOUBLE]*int(BACCARAT_TheOdds[baccarat.BACCARAT_ZONE_BANKER_DOUBLE]-1)
		if pos == baccarat.BACCARAT_ZONE_BANKER {
			maxBet /= int(BACCARAT_TheOdds[baccarat.BACCARAT_ZONE_BANKER] - 1)
		} else {
			maxBet /= int(BACCARAT_TheOdds[baccarat.BACCARAT_ZONE_BANKER_DOUBLE] - 1)
		}
	}

	if betCoin <= maxBet {
		return betCoin
	}

	minChip := int(this.DbGameFree.GetOtherIntParams()[0])
	return maxBet / minChip * minChip
}

func (this *BaccaratSceneData) BetStop() bool {
	if this.bankerSnId == -1 {
		return false
	}
	minChip := int(this.DbGameFree.GetOtherIntParams()[0])
	b := this.players[this.bankerSnId]
	var maxBet int

	// 判断是否还有可以加注的位置

	// 和
	// 最大赔付金额 = 庄家余额 + 庄对下注 + 闲对下注 - 和位置的赢分
	maxBet = int(b.Coin) +
		this.betInfo[baccarat.BACCARAT_ZONE_PLAYER_DOUBLE] +
		this.betInfo[baccarat.BACCARAT_ZONE_BANKER_DOUBLE] - this.betInfo[baccarat.BACCARAT_ZONE_TIE]*
		int(BACCARAT_TheOdds[baccarat.BACCARAT_ZONE_TIE]-1)
	maxBet /= int(BACCARAT_TheOdds[baccarat.BACCARAT_ZONE_TIE] - 1) // 最大可押注金额
	if maxBet >= minChip {
		return false
	}

	// 最大赔付金额 = 庄家余额 + 和下注 + 庄下注 + 庄对下注 - 闲位置赢分 - 闲对位置赢分
	maxBet = int(b.Coin) + this.betInfo[baccarat.BACCARAT_ZONE_TIE] + this.betInfo[baccarat.BACCARAT_ZONE_BANKER] +
		this.betInfo[baccarat.BACCARAT_ZONE_BANKER_DOUBLE] -
		this.betInfo[baccarat.BACCARAT_ZONE_PLAYER]*int(BACCARAT_TheOdds[baccarat.BACCARAT_ZONE_PLAYER]-1) -
		this.betInfo[baccarat.BACCARAT_ZONE_PLAYER_DOUBLE]*int(BACCARAT_TheOdds[baccarat.BACCARAT_ZONE_PLAYER_DOUBLE]-1)
	// 闲
	m := maxBet / int(BACCARAT_TheOdds[baccarat.BACCARAT_ZONE_PLAYER]-1)
	if m >= minChip {
		return false
	}
	// 闲对
	m = maxBet / int(BACCARAT_TheOdds[baccarat.BACCARAT_ZONE_PLAYER_DOUBLE]-1)
	if m >= minChip {
		return false
	}

	// 最大赔付金额 = 庄家余额 + 和下注 + 闲下注 + 闲对下注 - 庄位置赢分 - 庄对位置赢分
	maxBet = int(b.Coin) + this.betInfo[baccarat.BACCARAT_ZONE_TIE] + this.betInfo[baccarat.BACCARAT_ZONE_PLAYER] +
		this.betInfo[baccarat.BACCARAT_ZONE_PLAYER_DOUBLE] -
		this.betInfo[baccarat.BACCARAT_ZONE_BANKER]*int(BACCARAT_TheOdds[baccarat.BACCARAT_ZONE_BANKER]-1) -
		this.betInfo[baccarat.BACCARAT_ZONE_BANKER_DOUBLE]*int(BACCARAT_TheOdds[baccarat.BACCARAT_ZONE_BANKER_DOUBLE]-1)
	// 庄
	m = maxBet / int(BACCARAT_TheOdds[baccarat.BACCARAT_ZONE_BANKER]-1)
	if m > minChip {
		return false
	}
	// 庄对
	m = maxBet / int(BACCARAT_TheOdds[baccarat.BACCARAT_ZONE_BANKER_DOUBLE]-1)
	if m >= minChip {
		return false
	}
	return true
}
