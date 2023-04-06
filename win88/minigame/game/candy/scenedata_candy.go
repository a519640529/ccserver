package candy

import (
	"encoding/json"
	proto_gamehall "games.yol.com/win88/protocol/gamehall"
	"math/rand"
	"time"

	"games.yol.com/win88/common"
	"games.yol.com/win88/gamerule/candy"
	"games.yol.com/win88/minigame/base"
	"games.yol.com/win88/proto"
	proto_candy "games.yol.com/win88/protocol/candy"
	proto_server "games.yol.com/win88/protocol/server"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/timer"
)

const jackpotNoticeInterval = time.Second

type CandyJackpot struct {
	createdTime time.Time
	userName    string
	priceValue  int64
	roomID      int64
	// spinID      string
}

type CandySceneData struct {
	*base.Scene                            //房间信息
	players     map[int32]*CandyPlayerData //玩家信息
	//billedData          *proto_candy.GameBilledData //上一局结算信息
	jackpot             *base.XSlotJackpotPool //奖池
	jackpotNoticeHandle timer.TimerHandle      //奖池金额通知
	jackpotNoticeTime   time.Time              //上一次通知奖池的时间
	jackpotTime         time.Time              //上一次奖池变化的时间
}

func NewCandySceneData(s *base.Scene) *CandySceneData {
	return &CandySceneData{
		Scene:   s,
		players: make(map[int32]*CandyPlayerData),
	}
}

func (this *CandySceneData) SaveData(force bool) {
}

func (this *CandySceneData) OnPlayerLeave(p *base.Player, reason int) {
	if p, exist := this.players[p.SnId]; exist {
		delete(this.players, p.SnId)
	}
}

func (this *CandySceneData) SceneDestroy(force bool) {
	//销毁房间
	this.Scene.Destroy(force)
}

func (this *CandySceneData) init() bool {
	if this.DbGameFree == nil {
		return false
	}
	//this.billedData = &proto_candy.GameBilledData{}

	//初始化奖池信息
	otherIntParams := this.DbGameFree.GetOtherIntParams()
	params := this.DbGameFree.GetJackpot()
	this.jackpot = &base.XSlotJackpotPool{}
	this.jackpot.InitJackPot(len(otherIntParams))
	for i := 0; i < len(otherIntParams); i++ {
		baseBet := otherIntParams[i]
		this.jackpot.JackpotFund[i] = int64(params[candy.CANDY_JACKPOT_InitJackpot] * baseBet)
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

	this.jackpotTime = time.Now()
	this.AfterTimer()
	return true
}

type CandySpinResult struct {
	LinesInfo         []*proto_candy.CandyLinesInfo
	SlotsData         []int32
	TotalPrizeLine    int64 // 线条总金额
	TotalPrizeJackpot int64 // 爆奖总金额
	JackpotCnt        int   // 爆奖的次数
	IsJackpot         bool  // 是否爆奖
	TotalWinRate      int32 // 中奖总倍率
	TotalTaxScore     int64 // 税收
	WinLines          []int // 赢分的线
	LuckyData         []*proto_candy.CandyLinesInfo
}

func (this *CandySceneData) CalcLinePrize(cards []int, betLines []int64, betValue int64) (spinRes CandySpinResult) {
	betIdx := int32(common.InSliceInt32Index(this.DbGameFree.GetOtherIntParams(), int32(betValue)))
	taxRate := this.DbGameFree.GetTaxRate()
	calcTaxScore := func(score int64, taxScore *int64) int64 {
		newScore := int64(float64(score) * float64(10000-taxRate) / 10000.0)
		if taxScore != nil {
			*taxScore += score - newScore
		}
		return newScore
	}
	lines := candy.CalcLine(cards, betLines)
	for _, line := range lines {
		if line.Element == candy.Element_VUANGMIEN && line.Count == candy.LINE_CELL {
			spinRes.IsJackpot = true
			spinRes.JackpotCnt++
			var prizeJackpot int64
			if spinRes.TotalPrizeJackpot == 0 { // 第一个爆奖 获取当前奖池所有
				prizeJackpot = this.jackpot.JackpotFund[betIdx]
			} else { // 之后的爆奖 奖励为奖池初值
				prizeJackpot = int64(this.DbGameFree.GetJackpot()[candy.CANDY_JACKPOT_InitJackpot] * int32(betValue))
			}
			prizeJackpot = calcTaxScore(prizeJackpot, &spinRes.TotalTaxScore)
			spinRes.TotalPrizeJackpot += prizeJackpot
		}

		curScore := betValue * int64(line.Score) / 10
		curScore = calcTaxScore(curScore, &spinRes.TotalTaxScore)
		spinRes.TotalPrizeLine += curScore
		spinRes.TotalWinRate += int32(line.Score)

		lineInfo := &proto_candy.CandyLinesInfo{
			LineId:     proto.Int32(int32(line.Index)),
			Position:   line.Position,
			PrizeValue: proto.Int64(curScore),
		}
		spinRes.LinesInfo = append(spinRes.LinesInfo, lineInfo)
		spinRes.WinLines = append(spinRes.WinLines, int(lineInfo.GetLineId()))
		// luckyData
		if line.Element == candy.Element_Min && line.Count == candy.LINE_CELL && line.Score > candy.LineScore[candy.Element_Min][candy.LINE_CELL-1] {
			spinRes.LuckyData = append(spinRes.LinesInfo, lineInfo)
		}
	}
	for _, card := range cards {
		spinRes.SlotsData = append(spinRes.SlotsData, int32(card))
	}
	return
}

func (this *CandySceneData) broadcastJackpot() {
	pack := &proto_gamehall.BroadcastGameJackpot{
		GameFreeId: proto.Int32(this.DbGameFree.Id),
	}
	pack.JackpotFund = append(pack.JackpotFund, this.jackpot.JackpotFund...)
	proto.SetDefaults(pack)
	this.Broadcast(int(proto_gamehall.HundredScenePacketID_PACKET_BD_GAMEJACKPOT), pack, -1)

	//if this.lastJackpotValue != this.jackpot.JackpotFund {
	//	this.lastJackpotValue = this.jackpot.JackpotFund
	//	pack := &proto_gamehall.BroadcastGameJackpot{
	//		JackpotFund: proto.Int64(this.jackpot.JackpotFund),
	//		GameFreeId:  proto.Int32(this.DbGameFree.Id),
	//	}
	//	proto.SetDefaults(pack)
	//	//this.Broadcast(int(proto_candy.CandyPacketID_PACKET_SC_CANDY_GAMEJACKPOT), pack, -1)
	//	this.Broadcast(int(proto_gamehall.HundredScenePacketID_PACKET_BD_GAMEJACKPOT), pack, -1)
	//	//
	//	//msg := &proto_server.GWGameJackpot{
	//	//	SceneId:     proto.Int32(int32(this.SceneId)),
	//	//	JackpotFund: proto.Int64(this.jackpot.JackpotFund),
	//	//}
	//	//proto.SetDefaults(msg)
	//	//logger.Logger.Infof("GWGameJackpot gameFreeID %v %v", this.DbGameFree.GetId(), msg)
	//	//this.SendToWorld(int(proto_server.SSPacketID_PACKET_GW_GAMEJACKPOT), msg)
	//	//this.BroadcastMessageToPlatform(int(proto_gamehall.HundredScenePacketID_PACKET_BD_GAMEJACKPOT), pack)
	//}
}

func (this *CandySceneData) BroadcastJackpot() {
	// 重置奖池变化的时间
	this.jackpotTime = time.Now()
	this.broadcastJackpot()
	////距上次通知时间较长时 直接发送新通知
	//if time.Now().After(this.jackpotNoticeTime.Add(jackpotNoticeInterval + time.Millisecond*100)) {
	//	this.jackpotNoticeTime = time.Now()
	//	this.broadcastJackpot()
	//	return
	//}
	////避免通知太频繁
	//if this.jackpotNoticeHandle != timer.TimerHandle(0) {
	//	return
	//}
	//this.jackpotNoticeHandle, _ = timer.StartTimer(timer.TimerActionWrapper(func(h timer.TimerHandle, ud interface{}) bool {
	//	this.jackpotNoticeHandle = timer.TimerHandle(0)
	//	this.jackpotNoticeTime = time.Now()
	//	this.broadcastJackpot()
	//	return true
	//}), nil, jackpotNoticeInterval, 1)
}
func (this *CandySceneData) AfterTimer() {
	var timeInterval = time.Second * 30 // 时间间隔
	timer.AfterTimer(func(h timer.TimerHandle, ud interface{}) bool {
		if this.jackpotTime.Add(timeInterval).Before(time.Now()) {
			this.PushVirtualDataToWorld()
		}
		this.AfterTimer()
		return true
	}, nil, timeInterval)
}

func (this *CandySceneData) PushVirtualDataToWorld() {
	var isVirtualData bool
	otherIntParams := this.DbGameFree.GetOtherIntParams()
	var jackpotParam = this.DbGameFree.GetJackpot() // 奖池参数
	var jackpotInit = int64(jackpotParam[candy.CANDY_JACKPOT_InitJackpot] * otherIntParams[0])
	if jackpotInit > 0 && rand.Int31n(100000) < int32(this.jackpot.JackpotFund[0]*10/jackpotInit) { // 保留一位小数位
		isVirtualData = true
	}
	if isVirtualData {
		// 推送最新开奖记录到world
		msg := &proto_server.GWGameNewBigWinHistory{
			SceneId: proto.Int32(int32(this.SceneId)),
			BigWinHistory: &proto_server.BigWinHistoryInfo{
				CreatedTime:   proto.Int64(time.Now().Unix()),
				BaseBet:       proto.Int64(int64(this.DbGameFree.GetBaseScore())),
				TotalBet:      proto.Int64(int64(this.DbGameFree.GetBaseScore())),
				PriceValue:    proto.Int64(this.jackpot.JackpotFund[0]),
				IsVirtualData: proto.Bool(isVirtualData),
			},
		}
		this.jackpot.JackpotFund[0] = jackpotInit // 仅初始化奖池金额
		proto.SetDefaults(msg)
		logger.Logger.Infof("GWGameNewBigWinHistory gameFreeID %v %v", this.DbGameFree.GetId(), msg)
		this.SendToWorld(int(proto_server.SSPacketID_PACKET_GW_GAMENEWBIGWINHISTORY), msg)
	}
}
