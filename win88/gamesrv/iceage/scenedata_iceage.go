package iceage

import (
	"encoding/json"
	"games.yol.com/win88/protocol/gamehall"
	"games.yol.com/win88/protocol/iceage"
	"games.yol.com/win88/protocol/server"
	"math/rand"
	"time"

	rule "games.yol.com/win88/gamerule/iceage"
	"games.yol.com/win88/gamesrv/base"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/srvdata"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/timer"
)

type IceAgeJackpot struct {
	createdTime time.Time
	userName    string
	priceValue  int64
	roomID      int64
	spinID      string
}

type IceAgeSceneData struct {
	*base.Scene                                     //房间信息
	players             map[int32]*IceAgePlayerData //玩家信息
	jackpot             *base.XSlotJackpotPool      //奖池
	jackpotNoticeHandle timer.TimerHandle           //奖池金额通知
	jackpotNoticeTime   time.Time                   //上一次通知奖池的时间
	jackpotTime         time.Time                   //上一次奖池变化的时间
	lastJackpotValue    int64                       //上一次奖池变化时的值
}

func NewIceAgeSceneData(s *base.Scene) *IceAgeSceneData {
	return &IceAgeSceneData{
		Scene:   s,
		players: make(map[int32]*IceAgePlayerData),
	}
}

func (this *IceAgeSceneData) SaveData(force bool) {
}

func (this *IceAgeSceneData) OnPlayerLeave(p *base.Player, reason int) {
	if p, exist := this.players[p.SnId]; exist {
		delete(this.players, p.SnId)
	}
}

func (this *IceAgeSceneData) SceneDestroy(force bool) {
	//销毁房间
	this.Scene.Destroy(force)
}

func (this *IceAgeSceneData) init() bool {
	if this.GetDBGameFree() == nil {
		return false
	}
	params := this.GetDBGameFree().GetJackpot()
	this.jackpot = &base.XSlotJackpotPool{}
	if this.jackpot.JackpotFund <= 0 {
		this.jackpot.JackpotFund = int64(params[rule.ICEAGE_JACKPOT_InitJackpot] * this.GetDBGameFree().GetBaseScore())
	}

	str := base.XSlotsPoolMgr.GetPool(this.GetGameFreeId(), this.GetPlatform())
	if str != "" {
		jackpot := &base.XSlotJackpotPool{}
		err := json.Unmarshal([]byte(str), jackpot)
		if err == nil {
			this.jackpot = jackpot
		}
	}

	if this.jackpot != nil {
		base.XSlotsPoolMgr.SetPool(this.GetGameFreeId(), this.GetPlatform(), this.jackpot)
	}
	this.lastJackpotValue = this.jackpot.JackpotFund
	this.jackpotTime = time.Now()
	this.AfterTimer()
	return true
}

const (
	SlotsData      = "SlotsData"
	SlotsData_V2   = "SlotsData_V2"
	BonusData      = "BonusData"
	BonusData_V2   = "BonusData_V2"
	DefaultData    = "DefaultData"
	DefaultData_v1 = "DefaultData_v1"
)

func getElementDistributionByName(name string) *server.DB_IceAgeElementRate {
	for i := range srvdata.PBDB_IceAgeElementRateMgr.Datas.Arr {
		item := srvdata.PBDB_IceAgeElementRateMgr.GetData(int32(i + 1))
		if item.GetModeName() == name {
			return item
		}
	}
	return nil
}
func getSlotsDataByElementDistribution(data []int32) []int {
	var t = make([]int, rule.ELEMENT_TOTAL)
	rand.Seed(time.Now().UnixNano())
	for i, _ := range t {
		t[i] = int(data[rand.Intn(len(data))])
	}
	return t
}
func getSlotsDataByGroupName(name string) []int {
	var cardLib = make([]*server.DB_IceAgeCardLib, 0)
	for _, lib := range srvdata.PBDB_IceAgeCardLibMgr.Datas.Arr {
		if lib.GetModeName() == name {
			cardLib = append(cardLib, lib)
		}
	}
	rand.Seed(time.Now().UnixNano())
	lib := cardLib[rand.Intn(len(cardLib))].GetParams()
	var slotsData = make([]int, len(lib))
	for i, v := range lib {
		slotsData[i] = int(v)
	}
	return slotsData
}

type SpinResult struct {
	LinesInfo         []*iceage.IceAgeLinesInfo
	SlotsData         []*iceage.IceAgeCards
	TotalPrizeLine    int64   // 线条总金额+爆奖
	TotalPrizeBonus   int64   // 小游戏总金额
	BonusGameCnt      int32   // 小游戏次数
	TotalPrizeJackpot int64   // 爆奖总金额
	JackpotCnt        int     // 爆奖的次数
	AddFreeTimes      int32   // 新增免费次数
	IsJackpot         bool    // 是否爆奖
	TotalWinRate      int32   // 中奖总倍率
	TotalTaxScore     int64   // 税收
	WinLines          [][]int // 赢分的线
}

func (this *IceAgeSceneData) CalcSpinsPrize(cards []int, betLines []int64, distributionData []int32, bonusParams []int32, betValue int64, taxRate int32) (spinRes SpinResult) {

	tmpCards := make([]int, len(cards))
	copy(tmpCards, cards)

	calcTaxScore := func(score int64, taxScore *int64) int64 {
		tmpTaxScore := int64(float64(score) * float64(taxRate) / 10000.0)
		*taxScore += tmpTaxScore
		return score - tmpTaxScore
	}

	for loopIndex := 0; loopIndex < 99; loopIndex++ { // 避免死循环
		var winLines []int
		slotData := &iceage.IceAgeCards{}
		for _, card := range tmpCards {
			slotData.Card = append(slotData.Card, int32(card))
		}
		spinRes.SlotsData = append(spinRes.SlotsData, slotData)

		lineCount := 0
		for _, lineNum := range betLines {
			lineTemplate := rule.AllLineArray[int(lineNum)-1]
			edata := []int{}
			epos := []int32{}
			for _, pos := range lineTemplate {
				edata = append(edata, tmpCards[pos])
				epos = append(epos, int32(pos)+1)
			}
			head, count := rule.IsLine(edata)
			if head == 0 || count <= 0 {
				continue
			}

			var spinFree, prizeJackpot, bonus int64
			var prizesBonus = make([]int64, 0)
			if head == rule.Element_FREESPIN {
				spinFree = int64(rule.FreeSpinTimesRate[count-1])
			} else if head == rule.Element_JACKPOT && count == 5 {
				spinRes.IsJackpot = true
				spinRes.JackpotCnt++
				if spinRes.TotalPrizeJackpot == 0 { // 第一个爆奖 获取当前奖池所有
					prizeJackpot = this.jackpot.JackpotFund
				} else { // 之后的爆奖 奖励为奖池初值
					prizeJackpot = int64(this.GetDBGameFree().GetJackpot()[rule.ICEAGE_JACKPOT_InitJackpot] * this.GetDBGameFree().GetBaseScore())
				}
				prizeJackpot = calcTaxScore(prizeJackpot, &spinRes.TotalTaxScore)
				spinRes.TotalPrizeJackpot += prizeJackpot
			} else if head == rule.Element_BONUS && count >= 3 {
				rand.Seed(time.Now().UnixNano())
				bonus = betValue * int64(len(betLines)) * int64(bonusParams[rand.Intn(len(bonusParams))])
				bonus = calcTaxScore(bonus, &spinRes.TotalTaxScore)
				prizesBonus = []int64{bonus, bonus, bonus} // len大于0即可
				spinRes.BonusGameCnt++
			}

			curScore := int64(rule.LineScore[head][count-1]) * betValue
			curScore = calcTaxScore(curScore, &spinRes.TotalTaxScore)
			spinRes.TotalWinRate += int32(rule.LineScore[head][count-1])
			spinRes.TotalPrizeLine += curScore + prizeJackpot
			spinRes.TotalPrizeBonus += bonus
			spinRes.AddFreeTimes += int32(spinFree)

			line := &iceage.IceAgeLinesInfo{
				LineID:         proto.Int32(int32(lineNum)),
				Turn:           proto.Int32(int32(loopIndex + 1)),
				PrizeValue:     proto.Int64(curScore + prizeJackpot),
				PrizesFreespin: proto.Int64(spinFree),
				PrizesJackport: proto.Int64(prizeJackpot),
				PrizesBonus:    prizesBonus,
				Items:          epos[:count],
				RoleID:         proto.Int32(int32(head)),
			}
			spinRes.LinesInfo = append(spinRes.LinesInfo, line)
			winLines = append(winLines, int(line.GetLineID()))
			lineCount++
		}
		if winLines != nil {
			spinRes.WinLines = append(spinRes.WinLines, winLines)
		}

		if loopIndex == 98 {
			logger.Logger.Error("IceAgeSpinTimes99", cards, betLines, distributionData, bonusParams, betValue, taxRate)
		}

		// 没有匹配的线条 不用消除处理
		if lineCount == 0 {
			break
		}

		// 获取消除后的新线
		tmpCards = rule.MakePlan(tmpCards, distributionData, betLines)
		if len(tmpCards) == 0 {
			break
		}
	}
	return
}

func (this *IceAgeSceneData) broadcastJackpot() {
	if this.lastJackpotValue != this.jackpot.JackpotFund {
		this.lastJackpotValue = this.jackpot.JackpotFund
		pack := &gamehall.SCHundredSceneGetGameJackpot{}
		jpfi := &gamehall.GameJackpotFundInfo{
			GameFreeId:  proto.Int32(this.DbGameFree.Id),
			JackPotFund: proto.Int64(this.jackpot.JackpotFund),
		}
		pack.GameJackpotFund = append(pack.GameJackpotFund, jpfi)
		proto.SetDefaults(pack)
		//this.Broadcast(int(gamehall.HundredScenePacketID_PACKET_BD_GAMEJACKPOT), pack, -1)
		//
		//msg := &server.GWGameJackpot{
		//	SceneId:     proto.Int32(int32(this.GetSceneId())),
		//	JackpotFund: proto.Int64(this.jackpot.JackpotFund),
		//}
		//proto.SetDefaults(msg)
		//logger.Logger.Infof("GWGameJackpot gameFreeID %v %v", this.GetDBGameFree().GetId(), msg)
		//this.SendToWorld(int(server.SSPacketID_PACKET_GW_GAMEJACKPOT), msg)

		//以平台为标识向该平台内所有玩家广播奖池变动消息，游戏内外的玩家可监听该消息，减少由gamesrv向worldsrv转发这一步
		tags := []string{this.Platform}
		base.PlayerMgrSington.BroadcastMessageToGroup(int(gamehall.HundredScenePacketID_PACKET_SC_GAMEJACKPOT), pack, tags)
	}
}

func (this *IceAgeSceneData) BroadcastJackpot() {
	// 重置奖池变化的时间
	this.jackpotTime = time.Now()
	//距上次通知时间较长时 直接发送新通知
	if time.Now().After(this.jackpotNoticeTime.Add(jackpotNoticeInterval + time.Millisecond*100)) {
		this.jackpotNoticeTime = time.Now()
		this.broadcastJackpot()
		return
	}
	//避免通知太频繁
	if this.jackpotNoticeHandle != timer.TimerHandle(0) {
		return
	}
	this.jackpotNoticeHandle, _ = timer.StartTimer(timer.TimerActionWrapper(func(h timer.TimerHandle, ud interface{}) bool {
		this.jackpotNoticeHandle = timer.TimerHandle(0)
		this.jackpotNoticeTime = time.Now()
		this.broadcastJackpot()
		return true
	}), nil, jackpotNoticeInterval, 1)
}
func (this *IceAgeSceneData) AfterTimer() {
	var timeInterval = time.Second * 30 // 时间间隔
	timer.AfterTimer(func(h timer.TimerHandle, ud interface{}) bool {
		if this.jackpotTime.Add(timeInterval).Before(time.Now()) {
			this.PushVirtualDataToWorld()
		}
		this.AfterTimer()
		return true
	}, nil, timeInterval)
}

func (this *IceAgeSceneData) PushVirtualDataToWorld() {
	var isVirtualData bool
	var jackpotParam = this.GetDBGameFree().GetJackpot() // 奖池参数
	var jackpotInit = int64(jackpotParam[rule.ICEAGE_JACKPOT_InitJackpot] * this.GetDBGameFree().GetBaseScore())
	if rand.Int31n(100000) < int32(this.jackpot.JackpotFund*10/jackpotInit) { // 保留一位小数位
		isVirtualData = true
	}
	if isVirtualData {
		// 推送最新开奖记录到world
		msg := &server.GWGameNewBigWinHistory{
			SceneId: proto.Int32(int32(this.GetSceneId())),
			BigWinHistory: &server.BigWinHistoryInfo{
				CreatedTime:   proto.Int64(time.Now().Unix()),
				BaseBet:       proto.Int64(int64(this.GetDBGameFree().GetBaseScore())),
				TotalBet:      proto.Int64(int64(this.GetDBGameFree().GetBaseScore())),
				PriceValue:    proto.Int64(this.jackpot.JackpotFund),
				IsVirtualData: proto.Bool(isVirtualData),
			},
		}
		this.jackpot.JackpotFund = jackpotInit // 仅初始化奖池金额
		proto.SetDefaults(msg)
		logger.Logger.Infof("GWGameNewBigWinHistory gameFreeID %v %v", this.GetDBGameFree().GetId(), msg)
		this.SendToWorld(int(server.SSPacketID_PACKET_GW_GAMENEWBIGWINHISTORY), msg)
	}
}
