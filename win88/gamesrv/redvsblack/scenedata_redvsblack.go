package redvsblack

import (
	"fmt"
	"games.yol.com/win88/common"
	"games.yol.com/win88/gamerule/redvsblack"
	"games.yol.com/win88/gamesrv/base"
	"games.yol.com/win88/model"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/timer"
	"sort"
)

const (
	RVSB_ZONE_BLACK int = iota
	RVSB_ZONE_RED
	RVSB_ZONE_LUCKY
	RVSB_ZONE_MAX
)

//赔率
var RVSB_TheOdds = [RVSB_ZONE_MAX]int{1, 1, 1}
var RVSB_LuckyKindOdds = [redvsblack.CardsKind_Max]int{0, 0, 1, 2, 2, 2, 3, 5, 5, 5, 10}

const (
	RVSB_PHASE_NORMAL   int = iota //常规阶段
	RVSB_PHASE_RECOVERY            //回收阶段
)

type RedVsBlackSaveData struct {
	SysAllOutput int64 //系统总产出额
}

type RedVsBlackSceneData struct {
	*base.Scene                                                 //房间信息
	poker                 *redvsblack.Poker                     //牌
	cards                 [2][redvsblack.Hand_CardNum]int       //记录黑红两个位置的牌的信息
	kindOfcards           [2]*redvsblack.KindOfCard             //记录黑红两个位置的牌型信息
	pkResult              int                                   //比牌结果
	isWin                 [3]int                                //本局输赢 1赢  0平 -1输
	luckyKind             int                                   //幸运一击的牌型
	players               map[int32]*RedVsBlackPlayerData       //玩家信息
	seats                 []*RedVsBlackPlayerData               //座位信息(富豪倒序)
	winTop1               *RedVsBlackPlayerData                 //神算子
	betTop5               [RVSB_RICHTOP5]*RedVsBlackPlayerData  //押注最多的5位玩家
	betInfo               [RVSB_ZONE_MAX]int                    //记录下注位置及筹码
	betInfoRob            [RVSB_ZONE_MAX]int                    //记录机器人下注位置及筹码
	betDetailInfo         [RVSB_ZONE_MAX]map[int]int            //记录下注位置的详细筹码数据
	trend100Cur           []int32                               //最近100局走势
	trend20Lately         []int32                               //最近20局走势
	trend20CardKindLately []int32                               //最近20局幸运一击走势
	by                    func(p, q *RedVsBlackPlayerData) bool //排序函数
	hRunRecord            timer.TimerHandle                     //每十分钟记录一次数据
	hBatchSend            timer.TimerHandle                     //批量发送筹码数据
	constSeatKey          string                                //固定座位的唯一码
	bIntervention         bool
	webUser               string
}
type RBTSortData struct {
	*RedVsBlackPlayerData
	lostCoin int
	winCoin  int64
	realRate float64
}

//Len()
func (s *RedVsBlackSceneData) Len() int {
	return len(s.seats)
}

//Less():输赢记录将有高到底排序
func (s *RedVsBlackSceneData) Less(i, j int) bool {
	return s.by(s.seats[i], s.seats[j])
}

//Swap()
func (s *RedVsBlackSceneData) Swap(i, j int) {
	s.seats[i], s.seats[j] = s.seats[j], s.seats[i]
}

func NewRedVsBlackSceneData(s *base.Scene) *RedVsBlackSceneData {
	return &RedVsBlackSceneData{
		Scene:   s,
		players: make(map[int32]*RedVsBlackPlayerData),
	}
}

func (this *RedVsBlackSceneData) RebindPlayerSnId(oldSnId, newSnId int32) {
	if p, exist := this.players[oldSnId]; exist {
		delete(this.players, oldSnId)
		this.players[newSnId] = p
	}
}

func (this *RedVsBlackSceneData) init() bool {
	if this.DbGameFree == nil {
		return false
	}

	this.poker = redvsblack.NewPoker()
	this.LoadData()

	return true
}

func (this *RedVsBlackSceneData) LoadData() {
}

func (this *RedVsBlackSceneData) SaveData() {
}

func (this *RedVsBlackSceneData) Clean() {
	for _, p := range this.players {
		p.Clean()
	}
	for i := 0; i < 2; i++ {
		for j := 0; j < redvsblack.Hand_CardNum; j++ {
			this.cards[i][j] = -1
		}
		this.kindOfcards[i] = nil
	}

	for i := 0; i < RVSB_ZONE_MAX; i++ {
		this.betInfo[i] = 0
		this.betInfoRob[i] = 0
		this.betDetailInfo[i] = make(map[int]int)
	}
	this.bIntervention = false
	this.webUser = ""
	//重置水池调控标记
	this.CpControlled = false
}

func (this *RedVsBlackSceneData) CanStart() bool {
	cnt := len(this.players)
	if cnt > 0 {
		return true
	}
	return false
}

func (this *RedVsBlackSceneData) delPlayer(p *base.Player) {
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

func (this *RedVsBlackSceneData) OnPlayerEnter(p *base.Player) {
	if (p.Coin) < int64(this.DbGameFree.GetLimitCoin()) {
		p.MarkFlag(base.PlayerState_GameBreak)
		p.SyncFlag()
	}
}

func (this *RedVsBlackSceneData) OnPlayerLeave(p *base.Player, reason int) {
	this.delPlayer(p)
	this.BroadcastPlayerLeave(p, reason)
}

func (this *RedVsBlackSceneData) BroadcastPlayerLeave(p *base.Player, reason int) {

}

func (this *RedVsBlackSceneData) SceneDestroy(force bool) {
	//销毁房间
	this.Scene.Destroy(force)
}

func (this *RedVsBlackSceneData) IsNeedCare(p *RedVsBlackPlayerData) bool {
	if p.IsNeedCare() {
		return true
	}
	return false
}

func (this *RedVsBlackSceneData) IsNeedWeaken(p *RedVsBlackPlayerData) bool {
	if p.IsNeedWeaken() {
		return true
	}
	return false
}

//最大赢取倒序
func RVSB_OrderByLatest20Win(l, r *RedVsBlackPlayerData) bool {
	return l.lately20Win > r.lately20Win
}

//最大押注倒序
func RVSB_OrderByLatest20Bet(l, r *RedVsBlackPlayerData) bool {
	return l.lately20Bet > r.lately20Bet
}

func (this *RedVsBlackSceneData) Resort() string {
	var constSeatKey string
	best := int64(0)
	count := len(this.seats)
	this.winTop1 = nil
	for i := 0; i < count; i++ {
		if this.seats[i] != nil {
			this.seats[i].Pos = RVSB_OLPOS
			if this.seats[i].lately20Win > best {
				best = this.seats[i].lately20Win
				this.winTop1 = this.seats[i]
			}
		}
	}
	if this.winTop1 != nil {
		this.winTop1.Pos = RVSB_BESTWINPOS
		constSeatKey = fmt.Sprintf("%d_%d", this.winTop1.SnId, this.winTop1.Pos)
	}

	this.by = RVSB_OrderByLatest20Bet
	sort.Sort(this)
	cnt := 0
	for i := 0; i < RVSB_RICHTOP5; i++ {
		this.betTop5[i] = nil
	}
	for i := 0; i < count && cnt < RVSB_RICHTOP5; i++ {
		if this.seats[i] != this.winTop1 {
			this.seats[i].Pos = cnt + 1
			this.betTop5[cnt] = this.seats[i]
			cnt++
			constSeatKey += fmt.Sprintf("%d_%d", this.seats[i].SnId, this.seats[i].Pos)
		}
	}
	return constSeatKey
}

func (this *RedVsBlackSceneData) CalcuResult() (int, int) {
	for i := 0; i < 2; i++ {
		this.kindOfcards[i] = redvsblack.CardsKindFigureUpSington.FigureUpByCard(this.cards[i][:])
		if this.kindOfcards[i] != nil {
			this.kindOfcards[i].TidyCards()
		}
	}
	n := redvsblack.CompareCards(this.kindOfcards[0], this.kindOfcards[1])
	switch n {
	case -1: //红方胜利
		this.pkResult = 1
	case 1: //黑方胜利
		this.pkResult = 0
	}
	this.luckyKind = this.kindOfcards[this.pkResult].Kind
	return this.pkResult, this.luckyKind
}

func (this *RedVsBlackSceneData) PushTrend(result, lucky int32) {
	this.trend20Lately = append(this.trend20Lately, result)
	cnt := len(this.trend20Lately)
	if cnt > 20 { //近20局的趋势
		this.trend20Lately = this.trend20Lately[cnt-20:]
	}

	this.trend20CardKindLately = append(this.trend20CardKindLately, lucky)
	cnt = len(this.trend20CardKindLately)
	if cnt > 20 { //近20局的幸运一击趋势
		this.trend20CardKindLately = this.trend20CardKindLately[cnt-20:]
	}

	this.trend100Cur = append(this.trend100Cur, result)
	cnt = len(this.trend100Cur)
	if cnt >= RVSB_TREND100 { //近100局的趋势,每100局清理
		this.trend100Cur = nil
		this.trend20CardKindLately = nil
	}
}

func (this *RedVsBlackSceneData) GetRecoveryCoe() float64 {
	if coe, ok := model.GameParamData.HundredSceneRecoveryCoe[this.KeyGamefreeId]; ok {
		return float64(coe) / 100
	}
	return 0.05
}

// GetAllPlayerWinScore 所有玩家总输赢分
func (this *RedVsBlackSceneData) GetAllPlayerWinScore(result, kind int, billed bool) int {
	isLucky := kind > redvsblack.CardsKind_Double

	systemOutCoin := [RVSB_ZONE_MAX]int{0, 0, 0}
	sysOutCoin := 0

	//计算每个区域玩家下注数
	for key, value := range this.betInfo {
		systemOutCoin[key] += (value - this.betInfoRob[key])
	}

	//不结算时把幸运一击下注清空
	if !billed {
		systemOutCoin[RVSB_ZONE_LUCKY] = 0
	}

	//计算赢位置的赢钱数
	for i := 0; i < RVSB_ZONE_MAX; i++ {
		if result == i {
			sysOutCoin += systemOutCoin[i] * RVSB_TheOdds[i]
		} else {
			sysOutCoin -= systemOutCoin[i]
		}
	}

	//换牌不需要幸运一击，也不影响幸运一击的输赢情况，反而会影响到红黑区域下注的金币结算,结算时加入幸运一击数值
	if isLucky && billed {
		sysOutCoin += (RVSB_LuckyKindOdds[kind] + 1) * systemOutCoin[RVSB_ZONE_LUCKY]
	}

	return sysOutCoin
}

func (this *RedVsBlackSceneData) GetPlayerWinScore(p *RedVsBlackPlayerData, result, kind int, billed bool) int64 {
	isLucky := kind > redvsblack.CardsKind_Double
	//计算赢位置的赢钱数
	winScore := int64(p.betInfo[result]*(RVSB_TheOdds[result]+1) - p.betTotal)
	if !billed {
		winScore += int64(p.betInfo[RVSB_ZONE_LUCKY])
	}

	//换牌不需要幸运一击，也不影响幸运一击的输赢情况，反而会影响到红黑区域下注的金币结算,结算时加入幸运一击数值
	if isLucky && billed {
		winScore += int64((RVSB_LuckyKindOdds[kind] + 1) * p.betInfo[RVSB_ZONE_LUCKY])
	}
	return winScore
}

func (this *RedVsBlackSceneData) ChangeCard() bool {
	re, kind := this.CalcuResult()
	status, changeRate := base.CoinPoolMgr.GetCoinPoolStatus(this.Platform, this.GetGameFreeId(), this.GroupId)
	switch status {
	case base.CoinPoolStatus_Normal:
		//如果“当前库存值-开奖区域奖励倍数X此区域押注额<库存下线”，那么本局游戏，系统从“区域奖励倍数X区域押注金额”最小的两个中随机选择一个开奖；
		if re != RVSB_ZONE_LUCKY {
			curCoin := base.CoinPoolMgr.LoadCoin(this.GetGameFreeId(), this.Platform, this.GroupId)

			systemOutCoin := [RVSB_ZONE_MAX]int{0, 0, 0}
			//计算每个区域玩家下注数
			for key, value := range this.betInfo {
				systemOutCoin[key] += (value - this.betInfoRob[key])
			}

			rate := RVSB_TheOdds[re]
			setting := base.CoinPoolMgr.GetCoinPoolSetting(this.Platform, this.GetGameFreeId(), this.GroupId)
			if (curCoin - int64(systemOutCoin[re])*int64(rate)) < int64(setting.GetLowerLimit()) {
				if this.GetAllPlayerWinScore(re, kind, false) > 0 {
					this.cards[0], this.cards[1] = this.cards[1], this.cards[0]
					this.kindOfcards[0], this.kindOfcards[1] = nil, nil
					this.CpControlled = true
					return true
				}
			}
		}
	case base.CoinPoolStatus_Low: //库存值 < 库存下限
		if common.RandInt(10000) < changeRate {
			if this.GetAllPlayerWinScore(re, kind, false) > 0 {
				this.cards[0], this.cards[1] = this.cards[1], this.cards[0]
				this.kindOfcards[0], this.kindOfcards[1] = nil, nil
				this.CpControlled = true
				return true
			}
		}
	case base.CoinPoolStatus_High: //库存上限 < 库存值 < 库存上限+偏移量
		if common.RandInt(10000) < changeRate && this.CoinPoolCanOut() {
			if this.GetAllPlayerWinScore(re, kind, false) < 0 {
				this.cards[0], this.cards[1] = this.cards[1], this.cards[0]
				this.kindOfcards[0], this.kindOfcards[1] = nil, nil
				this.CpControlled = true
				return true
			}
		}
	case base.CoinPoolStatus_TooHigh: //库存>库存上限+偏移量
		if common.RandInt(10000) < changeRate && this.CoinPoolCanOut() {
			systmCoinOut := this.GetAllPlayerWinScore(re, kind, false)
			setting := base.CoinPoolMgr.GetCoinPoolSetting(this.Platform, this.GetGameFreeId(), this.GroupId)
			if systmCoinOut < 0 && (-systmCoinOut) <= int(setting.GetMaxOutValue()) {
				this.cards[0], this.cards[1] = this.cards[1], this.cards[0]
				this.kindOfcards[0], this.kindOfcards[1] = nil, nil
				this.CpControlled = true
				return true
			}
		}
	default: //CoinPoolStatus_Normal//库存下限 < 库存值 < 库存上限
		//不处理
	}
	return false
}
func (this *RedVsBlackSceneData) AutoBalance() bool {
	if !model.GameParamData.RbAutoBalance {
		return false
	}
	this.CalcuResult()
	bKind, rKind := this.kindOfcards[0].Kind, this.kindOfcards[1].Kind
	wins := []*RBTSortData{}
	for _, value := range this.players {
		if value.betTotal == 0 || value.IsRob {
			continue
		}
		realRate := float64(value.WinCoin+1) / float64(value.FailCoin+1)
		realCoin := value.WinCoin - value.FailCoin
		if realRate >= model.GameParamData.RbAutoBalanceRate || realCoin > model.GameParamData.RbAutoBalanceCoin {
			loseCoin, _, _, _ := value.GetLoseMaxPos(rKind, bKind)
			wins = append(wins, &RBTSortData{
				RedVsBlackPlayerData: value,
				lostCoin:             loseCoin,
				winCoin:              value.WinCoin,
				realRate:             realRate,
			})
		}
	}
	var beKill *RBTSortData
	switch len(wins) {
	case 0:
		beKill = nil
		return false
	case 1:
		beKill = wins[0]
		loseCoin, loseMaxPos, _, lKind := beKill.GetLoseMaxPos(rKind, bKind)
		systemWinCoin := this.GetAllPlayerWinScore(loseMaxPos, lKind, true)
		realCoin := loseCoin + systemWinCoin
		if systemWinCoin == 0 {
			systemWinCoin = 1
		}
		globalRate := float64(loseCoin) / float64(systemWinCoin)
		playerRate := float64(loseCoin) / float64(beKill.WinCoin+1)
		if realCoin < 0 && globalRate <= playerRate {
			return false
		}
		if realCoin < 0 && globalRate <= 0.8 {
			return false
		}
		if loseMaxPos == RVSB_ZONE_BLACK {
			if bKind < rKind {
				this.cards[0], this.cards[1] = this.cards[1], this.cards[0]
			}
		}
		if loseMaxPos == RVSB_ZONE_RED {
			if bKind > rKind {
				this.cards[0], this.cards[1] = this.cards[1], this.cards[0]
			}
		}
		//betinfo := [3]int{beKill.betInfo[0], beKill.betInfo[1], beKill.betInfo[2]}
		//wincoin := beKill.WinCoin
		//failcoin := beKill.FailCoin
		//re, kind := this.CalcuResult()
		//swc := this.GetAllPlayerWinScore(re, kind, true)
		//card := [2][3]int{this.cards[0], this.cards[1]}
		//task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
		//	model.NewRbAutoBalanceLog(beKill.SnId, betinfo, wincoin, failcoin, swc, card)
		//	return nil
		//}), nil, "NewRbAutoBalanceLog").Start()
		return true
	default:
		sort.Slice(wins, func(i, j int) bool {
			if wins[i].lostCoin == wins[j].lostCoin {
				if wins[i].winCoin == wins[j].winCoin {
					return wins[i].realRate > wins[j].realRate
				} else {
					return wins[i].winCoin > wins[j].winCoin
				}
			} else {
				return wins[i].lostCoin > wins[j].lostCoin
			}
		})
		for _, value := range wins {
			loseCoin, loseMaxPos, _, lKind := value.GetLoseMaxPos(rKind, bKind)
			systemWinCoin := this.GetAllPlayerWinScore(loseMaxPos, lKind, true)
			realCoin := loseCoin + systemWinCoin
			if systemWinCoin == 0 {
				systemWinCoin = 1
			}
			globalRate := float64(loseCoin) / float64(systemWinCoin)
			playerRate := float64(loseCoin) / float64(value.WinCoin+1)
			if realCoin < 0 && globalRate <= playerRate {
				continue
			}
			if realCoin < 0 && globalRate <= 0.8 {
				continue
			}
			if loseMaxPos == RVSB_ZONE_BLACK {
				if bKind < rKind {
					this.cards[0], this.cards[1] = this.cards[1], this.cards[0]
					this.CpControlled = true
				}
			}
			if loseMaxPos == RVSB_ZONE_RED {
				if bKind > rKind {
					this.cards[0], this.cards[1] = this.cards[1], this.cards[0]
					this.CpControlled = true
				}
			}
			//betinfo := [3]int{value.betInfo[0], value.betInfo[1], value.betInfo[2]}
			//wincoin := value.WinCoin
			//failcoin := value.FailCoin
			//re, kind := this.CalcuResult()
			//swc := this.GetAllPlayerWinScore(re, kind, true)
			//card := [2][3]int{this.cards[0], this.cards[1]}
			//task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
			//	model.NewRbAutoBalanceLog(value.SnId, betinfo, wincoin, failcoin, swc, card)
			//	return nil
			//}), nil, "NewRbAutoBalanceLog").Start()
			return true
		}
		return false
	}
}

//判断玩家是否需要单控
func (this *RedVsBlackSceneData) IsSingleRegulatePlayer() (bool, *RedVsBlackPlayerData) {
	// 多个单控玩家随机选一个
	var singlePlayers []*RedVsBlackPlayerData
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

	//闲家
	if singlePlayer.betTotal == 0 || int64(singlePlayer.betTotal) > data.BetMax || int64(singlePlayer.betTotal) < data.BetMin { //总下注下限≤玩家总下注额度≤总下注上限
		return false, nil
	}

	return true, singlePlayer
}

//本局结果是否需要单控
func (this *RedVsBlackSceneData) IsNeedSingleRegulate(p *RedVsBlackPlayerData) bool {
	data, ok := p.IsSingleAdjustPlayer()
	if data == nil || !ok {
		logger.Logger.Tracef("玩家%v 在%v场 不需要单控", p.SnId, this.GetGameFreeId())
		return false
	}

	//没有玩家上庄功能，所以可以有这个条件
	if p.betTotal == 0 {
		return false
	}
	p.result = data.Mode
	//如果是闲家
	re, kind := this.CalcuResult()
	winScore := this.GetPlayerWinScore(p, re, kind, true)
	if data.Mode == common.SingleAdjustModeLose && winScore >= 0 { //单控输 && 本局赢
		return true
	} else if data.Mode == common.SingleAdjustModeWin && winScore <= 0 { //单控赢 && 本局输
		return true
	}

	return false
}

//闲家单控
func (this *RedVsBlackSceneData) RegulationXianCard(p *RedVsBlackPlayerData) bool {
	data, ok := p.IsSingleAdjustPlayer()
	if data == nil || !ok {
		return false
	}

	this.cards[0], this.cards[1] = this.cards[1], this.cards[0]

	//检查下结果是否满足条件
	re, kind := this.CalcuResult()
	winScore := this.GetPlayerWinScore(p, re, kind, true)
	if (data.Mode == common.SingleAdjustModeWin && winScore > 0) ||
		(data.Mode == common.SingleAdjustModeLose && winScore < 0) {
		return true
	}
	//判断是否下注幸运一击
	if (this.betInfo[RVSB_ZONE_LUCKY] - this.betInfoRob[RVSB_ZONE_LUCKY]) > 0 {
		if data.Mode == common.SingleAdjustModeWin { // 单控赢，创造金花
			//在对子、顺子、金花、豹子这三种结果随机选择一种
			var idxList = [4]int{redvsblack.CardsKind_BigDouble, redvsblack.CardsKind_Straight,
				redvsblack.CardsKind_Flush, redvsblack.CardsKind_ThreeSame}
			kind := idxList[common.RandInt(4)]
			temp := redvsblack.CreateCardByKind(this.cards, kind)

			for j := 0; j < 3; j++ {
				this.cards[0][j] = temp[j]
				this.cards[1][j] = temp[j+3]
			}

			re, kind = this.CalcuResult()
			winScore = this.GetPlayerWinScore(p, re, kind, true)
			if winScore <= 0 {
				this.cards[0], this.cards[1] = this.cards[1], this.cards[0]
				this.CalcuResult()
			}
		} else if data.Mode == common.SingleAdjustModeLose { // 单控输，拆散特殊牌型
			if kind == redvsblack.CardsKind_ThreeSame { //如果存在豹子，换任一一张牌拆散豹子
				for j, _ := range this.cards[1] {
					this.cards[0][0], this.cards[1][j] = this.cards[1][j], this.cards[0][0]
					_, k := this.CalcuResult()
					if k != redvsblack.CardsKind_ThreeSame {
						break
					}
				}
			}

			if kind > redvsblack.CardsKind_Double {
				for i, _ := range this.cards[0] {
					for j, _ := range this.cards[1] {
						this.cards[0][i], this.cards[1][j] = this.cards[1][j], this.cards[0][i]
						_, k := this.CalcuResult()
						if k <= redvsblack.CardsKind_Double {
							break
						}
					}

					re, kind = this.CalcuResult()
					winScore = this.GetPlayerWinScore(p, re, kind, true)
					if winScore < 0 {
						break
					} else {
						this.cards[0], this.cards[1] = this.cards[1], this.cards[0]
						re, kind = this.CalcuResult()
						winScore = this.GetPlayerWinScore(p, re, kind, true)
						if winScore < 0 {
							break
						}
					}
				}
			}
		}
	}
	logger.Logger.Trace("RegulationXianCard 已调控")
	return true
}

func (this *RedVsBlackSceneData) CheckResults() {
	if this.bIntervention {
		return
	}
	if this.GetResult() == common.DefaultResult {
		this.bIntervention = false
		return
	}
	// 这里和CalcuResult方法的返回值规则对应一下
	n := this.GetResult()
	switch n {
	case 1: // 红赢
	case 2: // 黑赢
		n = 0
	default:
		this.bIntervention = false
		return
	}
	winFlag, _ := this.CalcuResult()
	if n != winFlag {
		this.cards[0], this.cards[1] = this.cards[1], this.cards[0]
	}
	this.bIntervention = true
	//this.webUser = this.Scene.webUser
}
