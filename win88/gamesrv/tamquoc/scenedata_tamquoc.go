package tamquoc

import (
	"encoding/json"
	"games.yol.com/win88/protocol/gamehall"
	"math/rand"
	"time"

	rule "games.yol.com/win88/gamerule/tamquoc"
	"games.yol.com/win88/gamesrv/base"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/server"
	"games.yol.com/win88/protocol/tamquoc"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/timer"
)

//type TamQuocJackpot struct {
//	createdTime time.Time
//	userName    string
//	priceValue  int64
//	roomID      int64
//	spinID      string
//}

type TamQuocSceneData struct {
	*base.Scene                                      //房间信息
	players             map[int32]*TamQuocPlayerData //玩家信息
	rooms               map[int][]*TamQuocPlayerData //房间
	jackpot             *base.XSlotJackpotPool       //奖池
	jackpotNoticeHandle timer.TimerHandle            //奖池金额通知
	jackpotNoticeTime   time.Time                    //上一次通知奖池的时间
	jackpotTime         time.Time                    //上一次奖池变化的时间
	lastJackpotValue    int64                        //上一次奖池变化时的值
}

func NewTamQuocSceneData(s *base.Scene) *TamQuocSceneData {
	return &TamQuocSceneData{
		Scene:   s,
		players: make(map[int32]*TamQuocPlayerData),
		rooms:   make(map[int][]*TamQuocPlayerData),
	}
}

func (this *TamQuocSceneData) SaveData(force bool) {
}

func (this *TamQuocSceneData) OnPlayerLeave(p *base.Player, reason int) {
	if p, exist := this.players[p.SnId]; exist {
		delete(this.players, p.SnId)
	}
}

func (this *TamQuocSceneData) SceneDestroy(force bool) {
	//销毁房间
	this.Scene.Destroy(force)
}

func (this *TamQuocSceneData) init() bool {
	if this.DbGameFree == nil {
		return false
	}
	params := this.DbGameFree.GetJackpot()
	this.jackpot = &base.XSlotJackpotPool{}
	if this.jackpot.JackpotFund <= 0 {
		this.jackpot.JackpotFund = int64(params[rule.TAMQUOC_JACKPOT_InitJackpot] * this.DbGameFree.GetBaseScore())
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

type TamQuocSpinResult struct {
	LinesInfo         []*tamquoc.TamQuocLinesInfo
	SlotsData         []int32
	TotalPrizeLine    int64 // 线条总金额
	TotalPrizeJackpot int64 // 爆奖总金额
	JackpotCnt        int   // 爆奖的次数
	AddFreeTimes      int32 // 新增免费次数
	IsJackpot         bool  // 是否爆奖
	BonusGame         tamquoc.TamQuocBonusGameInfo
	TotalWinRate      int32 // 中奖总倍率
	TotalTaxScore     int64 // 税收
	IsBonusGame       bool  // 是否进小游戏标志
	WinLines          []int // 赢分的线
}

func (this *TamQuocSceneData) CalcLinePrize(cards []int, betLines []int64, betValue int64) (spinRes TamQuocSpinResult) {
	taxRate := this.DbGameFree.GetTaxRate()
	calcTaxScore := func(score int64, taxScore *int64) int64 {
		newScore := int64(float64(score) * float64(10000-taxRate) / 10000.0)
		if taxScore != nil {
			*taxScore += score - newScore
		}
		return newScore
	}

	var startBonus int
	lines := rule.CalcLine(cards, betLines)
	for _, line := range lines {
		if line.Element == rule.Element_JACKPOT && line.Count == rule.LINE_CELL {
			spinRes.IsJackpot = true
			spinRes.JackpotCnt++
			var prizeJackpot int64
			if spinRes.TotalPrizeJackpot == 0 { // 第一个爆奖 获取当前奖池所有
				prizeJackpot = this.jackpot.JackpotFund
			} else { // 之后的爆奖 奖励为奖池初值
				prizeJackpot = int64(this.DbGameFree.GetJackpot()[rule.TAMQUOC_JACKPOT_InitJackpot] * this.DbGameFree.GetBaseScore())
			}
			prizeJackpot = calcTaxScore(prizeJackpot, &spinRes.TotalTaxScore)
			spinRes.TotalPrizeJackpot += prizeJackpot
		}

		curScore := betValue * int64(line.Score)
		curScore = calcTaxScore(curScore, &spinRes.TotalTaxScore)
		spinRes.TotalPrizeLine += curScore
		spinRes.TotalWinRate += int32(line.Score)

		lineInfo := &tamquoc.TamQuocLinesInfo{
			LineId:     proto.Int32(int32(line.Index)),
			Position:   line.Position,
			PrizeValue: proto.Int64(curScore),
		}
		spinRes.LinesInfo = append(spinRes.LinesInfo, lineInfo)
		spinRes.WinLines = append(spinRes.WinLines, int(lineInfo.GetLineId()))

		// bonus game
		if line.BonusElementCnt == 3 || line.BonusElementCnt == 4 {
			startBonus += rule.LineScore[rule.Element_BONUS][line.BonusElementCnt-1]
			spinRes.IsBonusGame = true
		}

		// add free
		if line.FreeElementCnt >= 3 {
			spinRes.AddFreeTimes += int32(rule.FreeSpinTimesRate[line.FreeElementCnt-1])
		}
	}

	if spinRes.IsBonusGame && startBonus > 0 {
		bonusGame := rule.GenerateBonusGame(int(betValue), startBonus)

		var totalBonusValue int64
		bonusData := make([]int64, 0)
		for _, value := range bonusGame.BonusData {
			value = calcTaxScore(value, nil)
			totalBonusValue += value
			bonusData = append(bonusData, value)
		}

		spinRes.BonusGame = tamquoc.TamQuocBonusGameInfo{
			TotalPrizeValue: proto.Int64(totalBonusValue * int64(bonusGame.Mutiplier)),
			Mutiplier:       proto.Int32(int32(bonusGame.Mutiplier)),
			DataMultiplier:  proto.Int64(totalBonusValue),
			BonusData:       bonusData,
		}
		// 小游戏税收
		bonusTax := (bonusGame.DataMultiplier - totalBonusValue) * int64(bonusGame.Mutiplier)
		spinRes.TotalTaxScore += bonusTax
	}

	for _, card := range cards {
		spinRes.SlotsData = append(spinRes.SlotsData, int32(card))
	}
	return
}

func (this *TamQuocSceneData) broadcastJackpot() {
	if this.lastJackpotValue != this.jackpot.JackpotFund {
		this.lastJackpotValue = this.jackpot.JackpotFund
		pack := &gamehall.SCHundredSceneGetGameJackpot{}
		jpfi := &gamehall.GameJackpotFundInfo{
			GameFreeId:  proto.Int32(this.DbGameFree.Id),
			JackPotFund: proto.Int64(this.jackpot.JackpotFund),
		}
		pack.GameJackpotFund = append(pack.GameJackpotFund, jpfi)
		proto.SetDefaults(pack)

		//以平台为标识向该平台内所有玩家广播奖池变动消息，游戏内外的玩家可监听该消息，减少由gamesrv向worldsrv转发这一步
		tags := []string{this.Platform}
		base.PlayerMgrSington.BroadcastMessageToGroup(int(gamehall.HundredScenePacketID_PACKET_SC_GAMEJACKPOT), pack, tags)
	}
}

func (this *TamQuocSceneData) BroadcastJackpot() {
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
func (this *TamQuocSceneData) AfterTimer() {
	var timeInterval = time.Second * 30 // 时间间隔
	timer.AfterTimer(func(h timer.TimerHandle, ud interface{}) bool {
		if this.jackpotTime.Add(timeInterval).Before(time.Now()) {
			this.PushVirtualDataToWorld()
		}
		this.AfterTimer()
		return true
	}, nil, timeInterval)
}

func (this *TamQuocSceneData) PushVirtualDataToWorld() {
	var isVirtualData bool
	var jackpotParam = this.DbGameFree.GetJackpot() // 奖池参数
	var jackpotInit = int64(jackpotParam[rule.TAMQUOC_JACKPOT_InitJackpot] * this.DbGameFree.GetBaseScore())
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
				TotalBet:      proto.Int64(int64(this.DbGameFree.GetBaseScore())),
				PriceValue:    proto.Int64(this.jackpot.JackpotFund),
				IsVirtualData: proto.Bool(isVirtualData),
			},
		}
		this.jackpot.JackpotFund = jackpotInit // 仅初始化奖池金额
		proto.SetDefaults(msg)
		logger.Logger.Infof("GWGameNewBigWinHistory gameFreeID %v %v", this.DbGameFree.GetId(), msg)
		this.SendToWorld(int(server.SSPacketID_PACKET_GW_GAMENEWBIGWINHISTORY), msg)
	}
}
