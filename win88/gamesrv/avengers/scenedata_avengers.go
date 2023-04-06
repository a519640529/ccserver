package avengers

import (
	"encoding/json"
	"math/rand"
	"time"

	rule "games.yol.com/win88/gamerule/avengers"
	"games.yol.com/win88/gamesrv/base"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/avengers"
	"games.yol.com/win88/protocol/gamehall"
	"games.yol.com/win88/protocol/server"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/timer"
)

type AvengersJackpot struct {
	createdTime time.Time
	userName    string
	priceValue  int64
	roomID      int64
	spinID      string
}

type AvengersSceneData struct {
	*base.Scene                                       //房间信息
	players             map[int32]*AvengersPlayerData //玩家信息
	jackpot             *base.XSlotJackpotPool        //奖池
	jackpotNoticeHandle timer.TimerHandle             //奖池金额通知
	jackpotNoticeTime   time.Time                     //上一次通知奖池的时间
	jackpotTime         time.Time                     //上一次奖池变化的时间
	lastJackpotValue    int64                         //上一次奖池变化时的值
}

func NewAvengersSceneData(s *base.Scene) *AvengersSceneData {
	return &AvengersSceneData{
		Scene:   s,
		players: make(map[int32]*AvengersPlayerData),
	}
}

func (this *AvengersSceneData) SaveData(force bool) {
}

func (this *AvengersSceneData) OnPlayerLeave(p *base.Player, reason int) {
	if p, exist := this.players[p.SnId]; exist {
		delete(this.players, p.SnId)
	}
}

func (this *AvengersSceneData) SceneDestroy(force bool) {
	//销毁房间
	this.Scene.Destroy(force)
}

func (this *AvengersSceneData) init() bool {
	if this.DbGameFree == nil {
		return false
	}
	params := this.DbGameFree.GetJackpot()
	this.jackpot = &base.XSlotJackpotPool{}
	if this.jackpot.JackpotFund <= 0 {
		this.jackpot.JackpotFund = int64(params[rule.AVENGERS_JACKPOT_InitJackpot] * this.DbGameFree.GetBaseScore())
	}
	str := base.XSlotsPoolMgr.GetPool(this.GetGameFreeId(), this.Platform)
	if str != "" {
		jackpot := &base.XSlotJackpotPool{}
		err := json.Unmarshal([]byte(str), jackpot)
		if err == nil {
			this.jackpot = jackpot
		}
	}

	if this.jackpot != nil {
		base.XSlotsPoolMgr.SetPool(this.GetGameFreeId(), this.Platform, this.jackpot)
	}
	this.lastJackpotValue = this.jackpot.JackpotFund
	this.jackpotTime = time.Now()
	this.AfterTimer()
	return true
}

type AvengersSpinResult struct {
	LinesInfo         []*avengers.AvengersLinesInfo
	SlotsData         []int32
	TotalPrizeLine    int64 // 线条总金额
	TotalPrizeJackpot int64 // 爆奖总金额
	JackpotCnt        int   // 爆奖的次数
	AddFreeTimes      int32 // 新增免费次数
	IsJackpot         bool  // 是否爆奖
	BonusGame         avengers.AvengersBonusGameInfo
	BonusX            []int32
	TotalWinRate      int32 // 中奖总倍率
	TotalTaxScore     int64 // 税收
	WinLines          []int // 赢分的线
}

func (this *AvengersSceneData) CalcLinePrize(cards []int, betLines []int64, betValue int64) (spinRes AvengersSpinResult) {
	taxRate := this.DbGameFree.GetTaxRate()
	calcTaxScore := func(score int64, taxScore *int64) int64 {
		newScore := int64(float64(score) * float64(10000-taxRate) / 10000.0)
		if taxScore != nil {
			*taxScore += score - newScore
		}
		return newScore
	}

	lines := rule.CalcLine(cards, betLines)
	for _, line := range lines {
		if line.Element == rule.Element_JACKPOT && line.Count == rule.LINE_CELL {
			spinRes.IsJackpot = true
			spinRes.JackpotCnt++
		}

		curScore := betValue * int64(line.Score)
		curScore = calcTaxScore(curScore, &spinRes.TotalTaxScore)
		spinRes.TotalPrizeLine += curScore
		spinRes.TotalWinRate += int32(line.Score)

		lineInfo := &avengers.AvengersLinesInfo{
			LineId:     proto.Int32(int32(line.Index)),
			Position:   line.Position,
			PrizeValue: proto.Int64(curScore),
		}
		spinRes.LinesInfo = append(spinRes.LinesInfo, lineInfo)
		spinRes.WinLines = append(spinRes.WinLines, int(lineInfo.GetLineId()))
	}

	if spinRes.IsJackpot { // 爆奖只计一条线
		spinRes.TotalPrizeJackpot = calcTaxScore(this.jackpot.JackpotFund, &spinRes.TotalTaxScore)
	}

	var countBonus, countFree int
	for _, card := range cards {
		if card == rule.Element_BONUS {
			countBonus++
		}
		if card == rule.Element_FREESPIN {
			countFree++
		}
		spinRes.SlotsData = append(spinRes.SlotsData, int32(card))
	}

	// bonus game
	if countBonus >= 3 {
		if countBonus == 3 {
			spinRes.BonusX = []int32{1, 2, 3}
		} else if countBonus == 4 {
			spinRes.BonusX = []int32{2, 3, 4}
		} else {
			spinRes.BonusX = []int32{3, 4, 5}
			countBonus = 5
		}
		totalBet := int64(len(betLines)) * betValue
		bonusGame := rule.GenerateBonusGame(int(totalBet), countBonus-2)
		var totalBonusValue int64
		bonusData := make([]int64, 0)
		for _, value := range bonusGame.BonusData {
			value = calcTaxScore(value, nil)
			totalBonusValue += value
			bonusData = append(bonusData, value)
		}
		spinRes.BonusGame = avengers.AvengersBonusGameInfo{
			TotalPrizeValue: proto.Int64(totalBonusValue * int64(bonusGame.Mutiplier)),
			Mutiplier:       proto.Int32(int32(bonusGame.Mutiplier)),
			DataMultiplier:  proto.Int64(totalBonusValue),
			BonusData:       bonusData,
		}
		// 小游戏税收
		bonusTax := (bonusGame.DataMultiplier - totalBonusValue) * int64(bonusGame.Mutiplier)
		spinRes.TotalTaxScore += bonusTax
	}

	// add free
	if countFree >= 3 {
		spinRes.AddFreeTimes = int32(countFree-2) * 4
	}

	return
}

func (this *AvengersSceneData) broadcastJackpot() {
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
		//	SceneId:     proto.Int32(int32(this.SceneId)),
		//	JackpotFund: proto.Int64(this.jackpot.JackpotFund),
		//}
		//proto.SetDefaults(msg)
		//logger.Logger.Infof("GWGameJackpot GamefreeId %v %v", this.DbGameFree.GetId(), msg)
		//this.SendToWorld(int(server.SSPacketID_PACKET_GW_GAMEJACKPOT), msg)

		//以平台为标识向该平台内所有玩家广播奖池变动消息，游戏内外的玩家可监听该消息，减少由gamesrv向worldsrv转发这一步
		tags := []string{this.Platform}
		base.PlayerMgrSington.BroadcastMessageToGroup(int(gamehall.HundredScenePacketID_PACKET_SC_GAMEJACKPOT), pack, tags)
	}

}

func (this *AvengersSceneData) BroadcastJackpot() {
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
func (this *AvengersSceneData) AfterTimer() {
	var timeInterval = time.Second * 30 // 时间间隔
	timer.AfterTimer(func(h timer.TimerHandle, ud interface{}) bool {
		if this.jackpotTime.Add(timeInterval).Before(time.Now()) {
			this.PushVirtualDataToWorld()
		}
		this.AfterTimer()
		return true
	}, nil, timeInterval)
}

func (this *AvengersSceneData) PushVirtualDataToWorld() {
	var isVirtualData bool
	var jackpotParam = this.DbGameFree.GetJackpot() // 奖池参数
	var jackpotInit = int64(jackpotParam[rule.AVENGERS_JACKPOT_InitJackpot] * this.DbGameFree.GetBaseScore())
	if rand.Int31n(100000) < int32(this.jackpot.JackpotFund*10/jackpotInit) { // 保留一位小数位
		isVirtualData = true
	}
	if isVirtualData {
		// 推送最新开奖记录到world
		msg := &server.GWGameNewBigWinHistory{
			SceneId: proto.Int32(int32(this.SceneId)),
			BigWinHistory: &server.BigWinHistoryInfo{
				CreatedTime:   proto.Int64(time.Now().Unix()),
				BaseBet:       proto.Int64(int64(this.DbGameFree.GetBaseScore())),
				PriceValue:    proto.Int64(this.jackpot.JackpotFund),
				IsVirtualData: proto.Bool(isVirtualData),
				TotalBet:      proto.Int64(int64(this.DbGameFree.GetBaseScore())),
			},
		}
		this.jackpot.JackpotFund = jackpotInit // 仅初始化奖池金额
		proto.SetDefaults(msg)
		logger.Logger.Infof("GWGameNewBigWinHistory GamefreeId %v %v", this.DbGameFree.GetId(), msg)
		this.SendToWorld(int(server.SSPacketID_PACKET_GW_GAMENEWBIGWINHISTORY), msg)
	}
}
