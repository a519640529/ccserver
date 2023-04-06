package fishing

import (
	"fmt"
	"games.yol.com/win88/gamesrv/base"
	"math"
	"math/rand"
	"time"

	"games.yol.com/win88/common"
	"games.yol.com/win88/gamerule/fishing"
	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/logger"
)

type Fish struct {
	FishID      int32
	TemplateID  int32
	DropCoin    int32
	MaxDropCoin int32
	Path        int32
	BirthTick   int32
	LiveTick    int32
	Event       int32 // 事件标签，用于记录当前 鱼所触发的状态
	InitTime    int64
	Policy      IPolicy
	Death       bool
	BaseRate    int
	Speed       int32
	IsBoss      int32
	Hp          *HpInfo
	IsJackpot   bool
	DealRate    int32
	FishType    int32 // 当前鱼是属于 第几类鱼 目前一共有 8 类

	//天天捕鱼，每条鱼对应单个玩家投入分数
	SwallowCoin map[int32]int32 //吞吃金额  key 玩家id  value 累计投入的金额
	Ratio       int32           //底分投入系数
	Ratio1      []int32         //激光炮掉血比例
}

// FishHPLists 实例
var FishHPListEx = &FishHPList{
	fishList: make(map[string]*FishQueue),
}

// FishHPList .
type FishHPList struct {
	fishList map[string]*FishQueue // id +level
}

// FishQueue 刷鱼队列
type FishQueue struct {
	queue     []*FishRealHp // 渔池模板
	currQueue []*FishRealHp // 当前渔池
	num       int
	index     int
}

// FishRealHp .
type FishRealHp struct {
	Id       int64
	CurrHp   int32 // 当前血量
	RateHp   int32 // 死亡判断
	CurrType int32 // 0 空闲 1 在渔场
}

var SoleID int64 // 唯一id

// Pop 出队
func (qe *FishQueue) CurrPop() (data *FishRealHp, ok bool) {
	if qe.Len() == 0 {
		return nil, false
	}
	for i, v := range qe.currQueue {
		if v.CurrType == 0 {
			qe.currQueue[i].CurrType = 1
			return qe.currQueue[i], true
		}
	}
	// 没有空闲的
	if qe.index == qe.num {
		qe.index = 0
	}
	fr := &FishRealHp{
		CurrHp:   qe.queue[qe.index].CurrHp,
		RateHp:   qe.queue[qe.index].RateHp,
		CurrType: 1,
		Id:       SoleID,
	}
	SoleID++
	qe.index++
	qe.currQueue = append(qe.currQueue, fr)
	return fr, true
}

func (qe *FishQueue) CurrTimeOut(data *FishRealHp) {

	for _, v := range qe.currQueue {
		if v.Id == data.Id {
			v.CurrType = 0
			v.CurrHp = data.CurrHp
			return
		}
	}
	logger.Logger.Error("CurrTimeOut err ", data.Id)
}

func (qe *FishQueue) CurrDel(data *FishRealHp) {

	for index, v := range qe.currQueue {
		if v.Id == data.Id {
			qe.currQueue = append(qe.currQueue[:index], qe.currQueue[index+1:]...)
			return
		}
	}
	logger.Logger.Error("CurrDel err ", data.Id)
}

// PutFirst 入队(首) 用与超时未死亡的鱼的入场
func (qe *FishQueue) PutFirst(data *FishRealHp) {
	// qe.num++
	index := -1
	flag := false
	for i, v := range qe.queue {
		if v.RateHp == data.RateHp {
			index = i
			if v.CurrHp == 0 {
				qe.queue = append(qe.queue[:index], qe.queue[index+1:]...)
				flag = true
				break
			}
		}
	}
	if !flag {
		if index != -1 {
			data.CurrHp += qe.queue[index].CurrHp
			logger.Logger.Warnf("PutFirst Warnf %v", data.CurrHp, qe.queue[index].CurrHp)
			qe.queue = append(qe.queue[:index], qe.queue[index+1:]...)
		} else {
			logger.Logger.Error("PutFirst err ", data, qe.num)
		}
	}
	queue := append([]*FishRealHp{}, qe.queue[0:]...)
	qe.queue = append(qe.queue[:0], data)
	qe.queue = append(qe.queue, queue...)
	return
}

// PutEnd 入队(尾)
func (qe *FishQueue) PutEnd(data *FishRealHp) {
	qe.num++
	qe.queue = append(qe.queue, data)
	return
}

// Len 长度
func (qe *FishQueue) Len() int {
	return qe.num
}

// Pop 出队
func (qe *FishQueue) Pop() (data *FishRealHp, ok bool) {
	if qe.Len() == 0 {
		return nil, false
	}
	qe.num--
	ret := qe.queue[0]
	qe.queue = qe.queue[1:]
	return ret, true
}

type HpInfo struct {
	Id        int64
	CurrHp    int32 // 当前血量，仅用来给客户端同步，不参与击杀概率计算
	Hp        int32 // 总生命(金币)
	RateHp    int32 // 死亡判断
	RobCurrHp int32 // 机器人打鱼死亡判断
}

/*
	新的鱼类初始化目前只支持寻龙夺宝
*/
func NewFish2(templateId, lv, instance, fishPath, liveTick, event, birthTick int32, dropCoin int32, policy IPolicy) *Fish {
	fishTemplate := base.FishTemplateEx.FishPool[templateId] //从 鱼池里获取对应的 模板数据
	if fishTemplate == nil {
		logger.Logger.Warnf("Fish [%v] init failed,no find in FishTemplatePool.", templateId)
		return nil
	}
	now := time.Now()
	fish := &Fish{
		FishID:     instance,
		TemplateID: fishTemplate.ID,
		Path:       fishPath,
		LiveTick:   liveTick,
		Event:      int32(event),
		BirthTick:  birthTick,
		Policy:     policy,
		Speed:      fishTemplate.Speed,
		IsBoss:     fishTemplate.Boss,
		FishType:   fishTemplate.FishType,
	}
	if len(fishTemplate.BaseRateA) != 2 {
		fishlogger.Error("Fish template data error:", fishTemplate)
	}
	switch lv {
	case 1:
		fish.BaseRate = common.RandInt(fishTemplate.BaseRateA[0], fishTemplate.BaseRateA[1])
	case 2:
		fish.BaseRate = common.RandInt(fishTemplate.BaseRateB[0], fishTemplate.BaseRateB[1])
	case 3:
		fish.BaseRate = common.RandInt(fishTemplate.BaseRateC[0], fishTemplate.BaseRateC[1])
	case -1:
		fish.BaseRate = common.RandInt(fishTemplate.BaseRateA[0], fishTemplate.BaseRateA[1])
	default:
		fish.BaseRate = common.RandInt(fishTemplate.BaseRateC[0], fishTemplate.BaseRateC[1])
		fishlogger.Errorf("%v rate data init error.", templateId)
	}
	if len(fishTemplate.HP) == 3 {
		hp := int32(0)
		switch lv {
		case 1:
			hp = fishTemplate.HP[0]
		case 2:
			hp = fishTemplate.HP[1]
		case 3:
			hp = fishTemplate.HP[2]
		default:
			hp = fishTemplate.HP[0]
		}
		fish.Hp = &HpInfo{
			CurrHp: 0,
			Hp:     hp,
		}
		if fishTemplate.Jackpot != 0 {
			rf := int32(common.RandInt(100))
			if fishTemplate.Jackpot > rf { // 奖金鱼
				fish.IsJackpot = true
			}
		}
		fish.DealRate = fishTemplate.DealRate
		fish.Ratio = fishTemplate.Ratio
		fish.Ratio1 = fishTemplate.Ratio1
	}
	//fish.DropCoin = int32(common.RandInt(int(fishTemplate.DropCoin[0]), int(fishTemplate.DropCoin[1])))
	if dropCoin == 0 {
		fish.DropCoin = GetFishDropCoin(fishTemplate.RandomCoin) // 重新定义了金币
	} else {
		fish.DropCoin = dropCoin
	}

	fish.InitTime = now.Unix()
	if fish.BaseRate == 0 {
		fishlogger.Errorf("%v rate init zero data.", templateId)
	}
	fish.SwallowCoin = make(map[int32]int32)
	//pathData, ok := fishMgr.Path[fishPath]
	//if ok {
	//	fish.LiveTick = fish.BirthTick + pathData.Stay*10 + pathData.Length/fish.Speed
	//}
	return fish
}

func NewFish(templateId, lv, instance, fishPath, liveTick, event, birthTick int32, policy IPolicy) *Fish {
	fishTemplate := base.FishTemplateEx.FishPool[templateId]
	if fishTemplate == nil {
		logger.Logger.Warnf("Fish [%v] init failed,no find in FishTemplatePool.", templateId)
		return nil
	}
	now := time.Now()
	fish := &Fish{
		FishID:     instance,
		TemplateID: fishTemplate.ID,
		Path:       fishPath,
		LiveTick:   liveTick,
		Event:      int32(event),
		BirthTick:  birthTick,
		Policy:     policy,
		Speed:      fishTemplate.Speed,
		IsBoss:     fishTemplate.Boss,
	}
	if len(fishTemplate.BaseRateA) != 2 {
		fishlogger.Error("Fish template data error:", fishTemplate)
	}
	switch lv {
	case 1:
		fish.BaseRate = common.RandInt(fishTemplate.BaseRateA[0], fishTemplate.BaseRateA[1])
	case 2:
		fish.BaseRate = common.RandInt(fishTemplate.BaseRateB[0], fishTemplate.BaseRateB[1])
	case 3:
		fish.BaseRate = common.RandInt(fishTemplate.BaseRateC[0], fishTemplate.BaseRateC[1])
	case -1:
		fish.BaseRate = common.RandInt(fishTemplate.BaseRateA[0], fishTemplate.BaseRateA[1])
	default:
		fish.BaseRate = common.RandInt(fishTemplate.BaseRateC[0], fishTemplate.BaseRateC[1])
		//fishlogger.Errorf("%v rate data init error.", templateId)
	}
	if len(fishTemplate.HP) == 3 {
		hp := int32(0)
		switch lv {
		case 1:
			hp = fishTemplate.HP[0]
		case 2:
			hp = fishTemplate.HP[1]
		case 3:
			hp = fishTemplate.HP[2]
		default:
			hp = fishTemplate.HP[0]
		}
		fish.Hp = &HpInfo{
			CurrHp: hp,
			Hp:     hp,
		}
		if fishTemplate.Jackpot != 0 {
			rf := int32(common.RandInt(100))
			if fishTemplate.Jackpot > rf { // 奖金鱼
				fish.IsJackpot = true
			}
		}
		fish.DealRate = fishTemplate.DealRate
		fish.Ratio = fishTemplate.Ratio
	}
	if fishTemplate.ID == fishing.Fish_CaiShen { //财神鱼，掉落金币随被攻击次数累加，而不是随机
		fish.DropCoin = fishTemplate.DropCoin[0]
	} else {
		fish.DropCoin = int32(common.RandInt(int(fishTemplate.DropCoin[0]), int(fishTemplate.DropCoin[1])))
	}
	fish.MaxDropCoin = fishTemplate.DropCoin[1]
	fish.InitTime = now.Unix()
	if fish.BaseRate == 0 {
		fishlogger.Errorf("%v rate init zero data.", templateId)
	}
	fish.SwallowCoin = make(map[int32]int32)
	//pathData, ok := fishMgr.Path[fishPath]
	//if ok {
	//	fish.LiveTick = fish.BirthTick + pathData.Stay*10 + pathData.Length/fish.Speed
	//}
	return fish
}
func (this *Fish) IsBirth(timePoint int32) bool {
	return this.BirthTick <= timePoint && timePoint <= this.LiveTick
}
func (this *Fish) SetDeath() {
	this.Death = true
	this.LiveTick = 0
}
func (this *Fish) IsDeath(timePoint int32) bool {
	if this.Death {
		return true
	}
	//return this.BirthTick > timePoint && timePoint > this.LiveTick
	return this.BirthTick < timePoint && timePoint > this.LiveTick
}

/*
	计算当日玩家倍率
*/
func CalcTodayPlayerOdds(player *FishingPlayerData, gameId string) float64 {
	winCoin := player.GetTodayGameData(gameId).TotalOut
	lostCoin := player.GetTodayGameData(gameId).TotalIn
	//fishlogger.Infof("CalcTodayPlayerOdds  playerId %v  winCoin %v  lostCoin %v ", player.SnId, winCoin, lostCoin)
	return float64(winCoin) / (float64(lostCoin) + 1)
}

// 根据鱼自身带的鱼分池对应的
func (this *Fish) NewFishRateHP3(playerOdds, sysOdds, preCorrect, ratio float64, ctroRat, power, snid int32) float64 {
	//击杀公式 炮弹底分/初始血量 * max[(2-个人赔率),0.8] * max[(2-系统赔率)，0.8] * （1-调节赔率） * 预警调控 * 个人限制系数
	var varA float64
	if this.SwallowCoin != nil && this.SwallowCoin[snid] > this.Hp.Hp && playerOdds < 1-float64(ctroRat)/10000 {
		if this.Hp.Hp-(this.SwallowCoin[snid]-this.Hp.Hp) < power {
			varA = 1
		} else {
			varA = float64(power) / float64(this.Hp.Hp-(this.SwallowCoin[snid]-this.Hp.Hp))
		}
		fishlogger.Tracef("GameId_TFishing this.SwallowCoin[%v] = %v this.Hp.Hp = %v", snid, this.SwallowCoin[snid], this.Hp.Hp)
	} else {
		varA = float64(power) / float64(this.Hp.Hp)
	}
	varB := math.Max(2-playerOdds, 0.8)
	varC := math.Max(2-sysOdds, 0.8)
	varD := 1 - float64(ctroRat)/10000
	rate := varA * varB * varC * varD * preCorrect * ratio
	//fishlogger.Tracef("GameId_TFishing power = %v hp = %v playerOdds = %v sysOdds = %v ctroRat = %v preCorrect = %v killRate = %v", power, this.Hp.Hp, playerOdds, sysOdds, ctroRat, preCorrect, rate)
	return rate
}

func (this *Fish) GetPlayerSwallowCoin(snid int32) int32 {
	if this.SwallowCoin != nil {
		return this.SwallowCoin[snid]
	}
	return 0
}

//鱼被击中，设置当前血量
func (this *Fish) OnHit(power int32) {
	if power > 0 {
		if this.Hp.CurrHp-power <= this.Hp.Hp*10/100 {
			this.Hp.CurrHp = this.Hp.Hp * 10 / 100
		} else {
			this.Hp.CurrHp = this.Hp.CurrHp - power
		}
	} else {
		//按照比例掉血
		if len(this.Ratio1) == 2 {
			rw := common.RandInt(int(this.Ratio1[0]), int(this.Ratio1[1]))
			dHp := this.Hp.CurrHp * int32(rw) / 10000
			if this.Hp.CurrHp-dHp <= this.Hp.Hp*10/100 {
				this.Hp.CurrHp = this.Hp.Hp * 10 / 100
			} else {
				this.Hp.CurrHp = this.Hp.CurrHp - dHp
			}
		}
	}

}

func CalcuJackpotCoin(power, gmlevel, ctroRate int32, key string, level int32, odds float64) (int64, int32) {
	//首先计算个人系数
	playerParam := float64(1)
	if odds < 0.5 {
		playerParam = 5
	} else if odds < 0.7 {
		playerParam = 4
	} else if odds < 0.9 {
		playerParam = 3
	} else if odds < 1-float64(ctroRate/10000) {
		playerParam = 1.5
	} else if odds < 1 {
		playerParam = 0.5
	} else {
		playerParam = 0
	}

	//接着计算爆奖概率
	jackpotBase := int64(0)
	jackpotRate := float64(0)
	switch level {
	case fishing.ROOM_LV_CHU:
		jackpotRate = float64(power) / 100 / 1000 * playerParam
		jackpotBase = 50 * 100
	case fishing.ROOM_LV_ZHO:
		jackpotRate = float64(power) / 100 / 10000 * playerParam
		jackpotBase = 500 * 100
	case fishing.ROOM_LV_GAO:
		jackpotRate = float64(power) / 100 / 100000 * playerParam
		jackpotBase = 5000 * 100
	}
	//测试中奖池
	if common.Config.IsDevMode && gmlevel == 5 {
		jackpotRate = 1
	}
	//根据概率计算玩家中奖金额
	ret := int64(0)
	jType := int32(4)
	randValue := common.RandInt(100000)
	if float64(randValue) < jackpotRate*100000 {
		rd := int32(common.RandInt(100))
		min, max := 10, 10
		if rd < 3 {
			min = 30
			max = 40
			jType = 1
		} else if rd < 8 {
			min = 20
			max = 25
			jType = 2
		} else if rd < 20 {
			min = 13
			max = 16
			jType = 3
		}
		ret = jackpotBase * int64(common.RandInt(min, max)) / 10
	}
	if ret > 0 {
		fishlogger.Tracef("CalcuJackpotCoin %v %v %v  %v %v %v %v %v", power, playerParam, jackpotRate, jackpotBase, randValue, jackpotRate*100000, ret, jType)
	}

	return ret, jType
}

/*
	欢乐捕鱼
	@ levelType  当前场此ID
	@ dropCoin   鱼类死亡的时候所掉落的金币
	@ player     用户基础数据
	@ GameId     当前游戏的ID

*/
func (this *Fish) HappyFishRate(levelType int, dropCoin int32, player *FishingPlayerData, gameId string, baseScore, ctroRate int32, key string, power int32) (bool, float64) {

	minmax := func(v, min, max float64) float64 {
		if v < min {
			return min
		}
		if v > max {
			return max
		}
		return v
	}

	ret := false
	//白名单
	whiteLevel := player.WhiteLevel + player.WhiteFlag
	if whiteLevel > 0 { //白名单用户
		r := float64(0)
		if player.WBMaxNum > 0 {
			r = 1 / float64(dropCoin) * (1 + float64(whiteLevel)*0.01)
		} else {
			r = 1/float64(dropCoin) + float64(whiteLevel)*0.01
		}

		if r >= rand.Float64() {
			ret = true
		}
		return ret, r
	} else if player.BlackLevel > 0 { //黑名单用户
		r := 1 / float64(dropCoin) / float64(player.BlackLevel+1)
		if r >= rand.Float64() {
			ret = true
		}
		return ret, r
	} else { //正常用户
		r := 1 / float64(dropCoin)
		// 计算抽水的比例系数
		p := (float64(ctroRate) / 10000)
		// 计算新用户的命中修正
		x1 := 0.0
		//if levelType == fishing.ROOM_LV_CHU {
		//	if player.GetAllBet(key) < 1000*int64(baseScore) {
		//		x1 = 0.01
		//	}
		//}

		// 计算用户盈利命中修正
		var playerInOut int64
		var odd float64
		pgi := player.GetGameFreeIdData(key)
		if pgi != nil {
			var base int64
			switch levelType {
			case fishing.ROOM_LV_CHU:
				base = 10000
			case fishing.ROOM_LV_ZHO:
				base = 100000
			case fishing.ROOM_LV_GAO:
				base = 1000000
			case fishing.ROOM_LV_RICH:
				base = 10000000
			}
			baseIn := base
			baseOut := base * (10000 - int64(ctroRate)) / 10000
			odd = float64(pgi.Statics.TotalOut+baseOut) / float64(pgi.Statics.TotalIn+baseIn)
			playerInOut = pgi.Statics.TotalIn + pgi.Statics.TotalOut
		} else {
			odd = 1
		}
		x2 := 1 - odd
		switch levelType {
		case fishing.ROOM_LV_CHU:
			x2 = minmax(x2, -0.01, 0.01)
		case fishing.ROOM_LV_ZHO:
			x2 = minmax(x2, -0.01, 0.001)
		case fishing.ROOM_LV_GAO:
			x2 = minmax(x2, -0.01, 0.0005)
		case fishing.ROOM_LV_RICH:
			x2 = minmax(x2, -0.02, 0.0001)
		}

		keyPf := fmt.Sprintf("%v_%v", player.scene.GetPlatform(), player.scene.GetGameFreeId())
		// 计算当前平台收益命中修正系数
		in, out := base.SysProfitCoinMgr.GetSysPfCoin(keyPf)
		switch levelType {
		case fishing.ROOM_LV_CHU:
			if in < 10000 {
				in += 10000
				out = int64(float64(in) * (1 - p))
				base.SysProfitCoinMgr.Add(keyPf, in, out)
			}
		case fishing.ROOM_LV_ZHO:
			if in < 100000 {
				in += 100000
				out = int64(float64(in) * (1 - p))
				base.SysProfitCoinMgr.Add(keyPf, in, out)
			}
		case fishing.ROOM_LV_GAO:
			if in < 1000000 {
				in += 1000000
				out = int64(float64(in) * (1 - p))
				base.SysProfitCoinMgr.Add(keyPf, in, out)
			}
		case fishing.ROOM_LV_RICH:
			if in < 10000000 {
				in += 10000000
				out = int64(float64(in) * (1 - p))
				base.SysProfitCoinMgr.Add(keyPf, in, out)
			}
		}

		pf := float64(out) / float64(in)
		x3 := 1 - p - pf
		switch levelType {
		case fishing.ROOM_LV_CHU:
			x3 = minmax(x3, -0.02, 0.02)
		case fishing.ROOM_LV_ZHO:
			x3 = minmax(x3, -0.01, 0.005)
		case fishing.ROOM_LV_GAO:
			x3 = minmax(x3, -0.01, 0.002)
		case fishing.ROOM_LV_RICH:
			x3 = minmax(x3, -0.01, 0.0001)
		}
		// 后台修正系数
		x4 := 0.0
		// 综合修正
		//x := x1 + x2 + x3 + x4
		// 计算最终鱼的死亡概率
		dr := r //(r + x) * (1 - p)  7-14修改
		if dr < 0 {
			if 1-pf < 0 {
				dr = 1 / float64(dropCoin) * 0.65
			} else if 1-pf > 0 && 1-pf < p {
				dr = 1 / float64(dropCoin) * 0.85
			} else {
				dr = 1 / float64(dropCoin) * (1 - p)
			}
		}

		switch levelType {
		case fishing.ROOM_LV_CHU: //免费用户约束
			//if player.CoinPayTotal < 100 && player.CoinPayGiveTotal < 100 && player.GMLevel == 0 {
			if player.CoinPayTotal < 100 && player.GMLevel == 0 {
				//增加流水限制,防止免费用户薅羊毛
				val := playerInOut - 20000 //20000:因为体验场初始投入产出各设置了10000
				if val >= 10000 && val < 20000 {
					dr *= 0.9
				} else if val >= 20000 && val < 40000 {
					dr *= 0.7
				} else if val >= 40000 {
					dr *= 0.5
				}
			}
		case fishing.ROOM_LV_GAO, fishing.ROOM_LV_RICH: //大R做下保护
			if odd <= (1-p) && pf <= (1-p) && (pgi.Statics.TotalIn-pgi.Statics.TotalOut) >= int64(baseScore*2000) {
				dr *= 2
			}
		}

		if common.Config.IsDevMode && player.GMLevel >= 5 {
			//test code
			dr = 1.0
			//test code
		}
		if dr >= 1.0 {
			ret = true
		} else {
			if dr >= rand.Float64() {
				ret = true
			}
		}
		if ret {
			switch levelType {
			case fishing.ROOM_LV_RICH: //至尊场限制下
				if int64(dropCoin*power)+out-in > 300000 {
					ret = false
				}
			case fishing.ROOM_LV_CHU: //初级场(低进场门槛限制,低于10元就认为是低门槛)
				//if !player.IsRob && player.scene.dbGameFree.GetLimitCoin() <= 1000 && player.CoinPayTotal < 100 && player.CoinPayGiveTotal < 100 && player.GMLevel == 0 { //免费用户
				//暂时屏蔽
				//if !player.IsRob && player.scene.GetDBGameFree().GetLimitCoin() <= 1000 && player.CoinPayTotal < 100 && player.GMLevel == 0 { //免费用户
				//	if int64(dropCoin*power)+player.CoinCache+player.SafeBoxCoin >= model.FishingParamData.FreeUserLimit {
				//		ret = false
				//	}
				//}
			}
		}
		fishlogger.Tracef("HappyFishRate dropCoin=%v r=%v p=%v x1=%v x2=%v x3=%v x4=%v dr=%v dead=%v", dropCoin, r, p, x1, x2, x3, x4, dr, ret)
		return ret, dr
	}
}

func init() {
	core.RegisteHook(core.HOOK_BEFORE_START, func() error {
		base.FishTemplateEx.Reload()
		return nil
	})
}
