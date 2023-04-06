package crash

import (
	"fmt"
	"games.yol.com/win88/gamerule/crash"
	"games.yol.com/win88/gamesrv/base"
	"games.yol.com/win88/model"
	"github.com/idealeak/goserver/core/timer"
	"sort"
)

type CrashSceneData struct {
	*base.Scene                                            //房间信息
	poker                 *crash.Poker                     //牌
	players               map[int32]*CrashPlayerData       //玩家信息
	seats                 []*CrashPlayerData               //座位信息(富豪倒序)
	winTop1               *CrashPlayerData                 //神算子
	betTop5               [CRASH_RICHTOP5]*CrashPlayerData //押注最多的5位玩家
	trend100Cur           []int32                          //最近100局走势
	trend20Lately         []int32                          //最近20局走势
	trend20CardKindLately []int32                          //最近20局幸运一击走势
	by                    func(p, q *CrashPlayerData) bool //排序函数
	hRunRecord            timer.TimerHandle                //每十分钟记录一次数据
	hBatchSend            timer.TimerHandle                //批量发送筹码数据
	explode               *crash.Card                      //爆炸点
	period                int                              //当前多少期
	wheel                 int                              //第几轮
	constSeatKey          string                           //固定座位的唯一码
	takeoffcurve          int32                            //起飞曲线,当前倍率
	sinvalue              float64                          //起飞计算
	allBetCoin            int64                            //总下注金额
	allBetPlayerNum       int32                            //下注的玩家数量
	parachutePlayerNum    int32                            //已跳伞玩家数量
	parachutePlayerCoin   int64                            //已跳伞玩家下注金额
}
type RBTSortData struct {
	*CrashPlayerData
	lostCoin int
	winCoin  int64
	realRate float64
}

//Len()
func (s *CrashSceneData) Len() int {
	return len(s.seats)
}

//Less():输赢记录将有高到底排序
func (s *CrashSceneData) Less(i, j int) bool {
	return s.by(s.seats[i], s.seats[j])
}

//Swap()
func (s *CrashSceneData) Swap(i, j int) {
	s.seats[i], s.seats[j] = s.seats[j], s.seats[i]
}

func NewCrashSceneData(s *base.Scene) *CrashSceneData {
	return &CrashSceneData{
		Scene:   s,
		players: make(map[int32]*CrashPlayerData),
	}
}

func (this *CrashSceneData) RebindPlayerSnId(oldSnId, newSnId int32) {
	if p, exist := this.players[oldSnId]; exist {
		delete(this.players, oldSnId)
		this.players[newSnId] = p
	}
}

func (this *CrashSceneData) init() bool {
	if this.DbGameFree == nil {
		return false
	}

	this.period = int(model.GetIntKVGameData("CrashPeriod"))
	this.wheel = int(model.GetIntKVGameData("CrashWheel"))
	this.poker = crash.NewPoker(this.period, this.wheel)
	this.allBetCoin = 0
	this.allBetPlayerNum = 0
	this.parachutePlayerNum = 0
	this.parachutePlayerCoin = 0
	this.LoadData()

	this.takeoffcurve = 100
	this.sinvalue = 0.0
	return true
}

func (this *CrashSceneData) LoadData() {
}

func (this *CrashSceneData) SaveData() {
}

func (this *CrashSceneData) Clean() {
	for _, p := range this.players {
		p.Clean()
	}
	this.takeoffcurve = 100
	this.sinvalue = 0.0
	this.allBetCoin = 0
	this.allBetPlayerNum = 0
	this.parachutePlayerNum = 0
	this.parachutePlayerCoin = 0
	//重置水池调控标记
	this.CpControlled = false
}

func (this *CrashSceneData) CanStart() bool {
	cnt := len(this.players)
	if cnt > 0 {
		return true
	}
	return false
}

func (this *CrashSceneData) delPlayer(p *base.Player) {
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

func (this *CrashSceneData) OnPlayerEnter(p *base.Player) {
	if (p.Coin) < int64(this.DbGameFree.GetLimitCoin()) {
		p.MarkFlag(base.PlayerState_GameBreak)
		p.SyncFlag()
	}
}

func (this *CrashSceneData) OnPlayerLeave(p *base.Player, reason int) {
	this.delPlayer(p)
	this.BroadcastPlayerLeave(p, reason)
}

func (this *CrashSceneData) BroadcastPlayerLeave(p *base.Player, reason int) {

}

func (this *CrashSceneData) SceneDestroy(force bool) {
	//销毁房间
	this.Scene.Destroy(force)
}

func (this *CrashSceneData) IsNeedCare(p *CrashPlayerData) bool {
	if p.IsNeedCare() {
		return true
	}
	return false
}

func (this *CrashSceneData) IsNeedWeaken(p *CrashPlayerData) bool {
	if p.IsNeedWeaken() {
		return true
	}
	return false
}

//最大下注倍率排序
func CRASH_OrderByMultiple(l, r *CrashPlayerData) bool {
	return l.multiple < r.multiple
}

//最大下注排序
func CRASH_OrderByBetTotal(l, r *CrashPlayerData) bool {
	return l.betTotal < r.betTotal
}

//最大赢钱排序
func CRASH_OrderByMultipleAndBetTotal(l, r *CrashPlayerData) bool {
	return int64(l.multiple)*l.betTotal > int64(r.multiple)*r.betTotal
}

//最大赢取倒序
func CRASH_OrderByLatest20Win(l, r *CrashPlayerData) bool {
	return l.lately20Win > r.lately20Win
}

//最大押注倒序
func CRASH_OrderByLatest20Bet(l, r *CrashPlayerData) bool {
	return l.lately20Bet > r.lately20Bet
}

func (this *CrashSceneData) Resort() string {
	var constSeatKey string
	best := int64(0)
	count := len(this.seats)
	this.winTop1 = nil
	for i := 0; i < count; i++ {
		if this.seats[i] != nil {
			this.seats[i].Pos = CRASH_OLPOS
			if this.seats[i].lately20Win > best {
				best = this.seats[i].lately20Win
				this.winTop1 = this.seats[i]
			}
		}
	}
	if this.winTop1 != nil {
		this.winTop1.Pos = CRASH_BESTWINPOS
		constSeatKey = fmt.Sprintf("%d_%d", this.winTop1.SnId, this.winTop1.Pos)
	}

	this.by = CRASH_OrderByLatest20Bet
	sort.Sort(this)
	cnt := 0
	for i := 0; i < CRASH_RICHTOP5; i++ {
		this.betTop5[i] = nil
	}
	for i := 0; i < count && cnt < CRASH_RICHTOP5; i++ {
		if this.seats[i] != this.winTop1 {
			this.seats[i].Pos = cnt + 1
			this.betTop5[cnt] = this.seats[i]
			cnt++
			constSeatKey += fmt.Sprintf("%d_%d", this.seats[i].SnId, this.seats[i].Pos)
		}
	}
	return constSeatKey
}

func (this *CrashSceneData) GetRecoveryCoe() float64 {
	if coe, ok := model.GameParamData.HundredSceneRecoveryCoe[this.KeyGamefreeId]; ok {
		return float64(coe) / 100
	}
	return 0.05
}

func (this *CrashSceneData) PushTrend(result int32) {
	this.trend20Lately = append(this.trend20Lately, result)
	cnt := len(this.trend20Lately)
	if cnt > 20 { //近20局的趋势
		this.trend20Lately = this.trend20Lately[cnt-20:]
	}

	//this.trend20CardKindLately = append(this.trend20CardKindLately, result)
	//cnt = len(this.trend20CardKindLately)
	//if cnt > 20 { //近20局的幸运一击趋势
	//	this.trend20CardKindLately = this.trend20CardKindLately[cnt-20:]
	//}

	this.trend100Cur = append(this.trend100Cur, result)
	cnt = len(this.trend100Cur)
	if cnt >= CRASH_TREND100 { //近100局的趋势,每100局清理
		this.trend100Cur = nil
		this.trend20CardKindLately = nil
	}
}

// GetAllPlayerWinScore 所有玩家总输赢分
func (this *CrashSceneData) GetAllPlayerWinScore(result int32) int64 {
	sysOutCoin := int64(0)

	//计算赢位置的赢钱数
	for _, p := range this.seats {
		if p == nil || p.IsRob || p.betTotal == 0 {
			continue
		}
		if p.multiple < int32(this.explode.Explode) && p.multiple >= crash.MinMultiple {
			sysOutCoin = int64(p.multiple) * p.betTotal / 100
		} else {
			sysOutCoin = -p.betTotal
		}
	}

	return sysOutCoin
}
