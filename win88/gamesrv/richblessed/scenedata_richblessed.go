package richblessed

import (
	"encoding/json"
	"fmt"
	"math"

	"games.yol.com/win88/common"
	"games.yol.com/win88/gamerule/richblessed"
	"games.yol.com/win88/gamesrv/base"
	"games.yol.com/win88/proto"
	protocol "games.yol.com/win88/protocol/richblessed"
	"games.yol.com/win88/srvdata"
	"github.com/idealeak/goserver/core/logger"
)

type RichBlessedSceneData struct {
	*base.Scene                                       //场景
	players          map[int32]*RichBlessedPlayerData //玩家信息
	jackpot          *base.SlotJackpotPool            //奖池
	levelRate        [][]int32                        //调控概率区间
	slotRateWeight   []int32                          //调控权重
	sysProfitCoinKey string
}

func NewRichBlessedSceneData(s *base.Scene) *RichBlessedSceneData {
	sceneEx := &RichBlessedSceneData{
		Scene:   s,
		players: make(map[int32]*RichBlessedPlayerData),
	}
	sceneEx.Init()
	return sceneEx
}
func (s *RichBlessedSceneData) Init() {
	s.LoadJackPotData()
	//for _, data := range srvdata.PBDB_SlotRateMgr.Datas.Arr {
	//	if int(data.GetGameId()) == common.GameId_Fruits {
	//		s.levelRate = append(s.levelRate, data.RateSection)
	//		s.slotRateWeight = append(s.slotRateWeight, data.GetWeight())
	//	}
	//}
	s.sysProfitCoinKey = fmt.Sprintf("%v_%v", s.Platform, s.GetGameFreeId())
}

func (s *RichBlessedSceneData) Clear() {

}
func (s *RichBlessedSceneData) SceneDestroy(force bool) {
	//销毁房间
	s.Scene.Destroy(force)
}
func (s *RichBlessedSceneData) AddPrizeCoin(isRob bool, totalBet int64) {
	addPrizeCoin := totalBet * richblessed.NowByte * 3 / 100 //扩大10000倍
	s.jackpot.AddToGrand(isRob, addPrizeCoin)
	s.jackpot.AddToBig(isRob, addPrizeCoin)
	s.jackpot.AddToMiddle(isRob, addPrizeCoin)
	s.jackpot.AddToSmall(isRob, addPrizeCoin)
	logger.Logger.Tracef("[Rich] 奖池增加...AddPrizeCoin... %f", float64(addPrizeCoin)/float64(10000))
	base.SlotsPoolMgr.SetPool(s.GetGameFreeId(), s.Platform, s.jackpot)
}
func (s *RichBlessedSceneData) DelPrizeCoin(isRob bool, win int64) {
	if win > 0 {
		s.jackpot.AddToSmall(isRob, -win*richblessed.NowByte)
		logger.Logger.Tracef("[Rich] 奖池减少...DelPrizeCoin... %d", win)
		base.SlotsPoolMgr.SetPool(s.GetGameFreeId(), s.Platform, s.jackpot)
	}
}
func (s *RichBlessedSceneData) delPlayer(SnId int32) {
	if _, exist := s.players[SnId]; exist {
		delete(s.players, SnId)
	}
}
func (s *RichBlessedSceneData) OnPlayerLeave(p *base.Player, reason int) {
	if playerEx, ok := p.ExtraData.(*RichBlessedPlayerData); ok {
		if playerEx.freeTimes > 0 || playerEx.JackpotEle != -1 {
			if playerEx.JackpotEle != -1 { // 小游戏
				s.JACKPOTWin(playerEx)
				playerEx.Clear()
			}
			freenum := playerEx.freeTimes
			for i := int32(0); i < freenum; i++ {
				playerEx.Clear()
				playerEx.CreateResult(s.GetFreeWeight()) // 没有铜锣概率
				s.Win(playerEx)

			}

		}

		s.delPlayer(p.SnId)
	}
}

// 是否中奖
func (s *RichBlessedSceneData) CanJACKPOT(p *RichBlessedPlayerData, bet int64, big int64, ele []int32) bool {
	re := p.result.CanJACKPOT(bet, big)
	logger.Logger.Tracef("RichBlessedSceneData CanJACKPOT %v %v", re, p.result.EleValue)
	if re || !p.test {
		ele := p.result.CreateJACKPOT(ele)
		JACKPOT := int64(0)
		// TODO 奖池判断
		// p.JackpotEle = ele
		// return true
		switch p.JackpotEle {
		case richblessed.GoldBoy:
			JACKPOT = s.jackpot.GetTotalGrand()
		case richblessed.GoldGirl:
			JACKPOT = s.jackpot.GetTotalBig()
		case richblessed.BlueBoy:
			JACKPOT = s.jackpot.GetTotalMiddle()
		default:
			JACKPOT = s.jackpot.GetTotalSmall()
		}
		if JACKPOT >= richblessed.JkEleNumRate[int((ele))] || !p.test {
			p.test = true // 中奖就取消
			p.JackpotEle = ele
			return true
		}
	}
	return false
}
func (s *RichBlessedSceneData) Win(p *RichBlessedPlayerData) {
	p.result.Win(p.betCoin, p.maxbetCoin)
	if p.result.FreeNum != 0 { //
		p.addfreeTimes = p.result.FreeNum
		p.freeTimes += p.addfreeTimes
	}
	p.winLineRate = p.result.AllRate
	p.winCoin += p.oneBetCoin * p.winLineRate
	if p.gameState == richblessed.FreeGame {
		p.freewinCoin += p.winCoin
	}
	if p.winCoin != 0 {
		p.noWinTimes = 0
		//SysProfitCoinMgr.Add(s.sysProfitCoinKey, 0, p.winCoin)
		p.Statics(s.KeyGameId, s.KeyGamefreeId, p.winCoin, false)
		tax := int64(math.Ceil(float64(p.winCoin) * float64(s.DbGameFree.GetTaxRate()) / 10000))
		p.taxCoin = tax
		p.winCoin -= tax
		p.AddServiceFee(tax)

		p.AddCoin(p.winCoin, common.GainWay_HundredSceneWin, 0, "system", s.GetSceneName())
		if !s.Testing && p.winCoin != 0 {
			base.CoinPoolMgr.PushCoin(s.GetGameFreeId(), s.GroupId, s.Platform, -(p.winCoin + tax))
		}
		//p.isReportGameEvent = true
	}
}

func (s *RichBlessedSceneData) JACKPOTWin(p *RichBlessedPlayerData) {
	p.result.JACKPOTWin()
	p.JackwinCoin = p.result.JackpotRate * p.oneBetCoin
	if p.JackwinCoin != 0 {
		s.DelPrizeCoin(p.IsRob, p.JackwinCoin)
		p.noWinTimes = 0
		//SysProfitCoinMgr.Add(s.sysProfitCoinKey, 0, p.JackwinCoin)
		p.Statics(s.KeyGameId, s.KeyGamefreeId, p.JackwinCoin, false)
		tax := int64(math.Ceil(float64(p.JackwinCoin) * float64(s.DbGameFree.GetTaxRate()) / 10000))
		p.taxCoin = tax
		p.JackwinCoin -= tax
		p.AddServiceFee(tax)

		p.AddCoin(p.JackwinCoin, common.GainWay_HundredSceneWin, 0, "system", s.GetSceneName())
		if !s.Testing && p.JackwinCoin != 0 {
			base.CoinPoolMgr.PushCoin(s.GetGameFreeId(), s.GroupId, s.Platform, -(p.JackwinCoin + tax))
		}
		p.JackpotEle = -1
		//p.isReportGameEvent = true
	}
	// TODO 广播
}

func (s *RichBlessedSceneData) SendBilled(p *RichBlessedPlayerData) {
	//正常游戏 免费游戏
	pack := &protocol.SCRBBilled{
		NowGameState: proto.Int(p.gameState),
		BetIdx:       proto.Int(p.betIdx),
		Coin:         proto.Int64(p.Coin),
		Cards:        p.result.EleValue,
		FreeAllWin:   proto.Int64(p.freewinCoin), //
		//SmallJackpot:  proto.Int64(s.jackpot.GetTotalSmall() / 10000),
		//MiddleJackpot: proto.Int64(s.jackpot.GetTotalMiddle() / 10000),
		//BigJackpot:    proto.Int64(s.jackpot.GetTotalBig() / 10000),
		//GrandJackpot:  proto.Int64(s.jackpot.GetTotalGrand() / 10000),
		WinEleCoin: proto.Int64(p.winCoin),
		WinRate:    proto.Int64(p.winLineRate),
		FreeNum:    proto.Int64(int64(p.freeTimes)),
		AddFreeNum: proto.Int64(int64(p.addfreeTimes)),
		JackpotEle: proto.Int32(p.JackpotEle),
	}
	var wl []*protocol.RichWinLine
	for _, r := range p.result.WinLine {
		wl = append(wl, &protocol.RichWinLine{
			Poss: r.Poss,
		})
	}
	pack.WinLines = wl
	logger.Logger.Trace("SCRBBilled：", pack)
	p.SendToClient(int(protocol.RBPID_PACKET_RICHBLESSED_SCRBBilled), pack)
}

func (s *RichBlessedSceneData) SendJACKPOTBilled(p *RichBlessedPlayerData) {

	pack := &protocol.SCRBBilled{
		NowGameState: proto.Int(p.gameState),
		BetIdx:       proto.Int(p.betIdx),
		Coin:         proto.Int64(p.Coin),
		//SmallJackpot:  proto.Int64(s.jackpot.GetTotalSmall() / 10000),
		//MiddleJackpot: proto.Int64(s.jackpot.GetTotalMiddle() / 10000),
		//BigJackpot:    proto.Int64(s.jackpot.GetTotalBig() / 10000),
		//GrandJackpot:  proto.Int64(s.jackpot.GetTotalGrand() / 10000),
		WinJackpot: proto.Int64(p.JackwinCoin),
	}
	logger.Logger.Trace("SendJACKPOTBilled:", pack)
	p.SendToClient(int(protocol.RBPID_PACKET_RICHBLESSED_SCRBJACKBilled), pack)
}

func (s *RichBlessedSceneData) LoadJackPotData() {
	str := base.SlotsPoolMgr.GetPool(s.GetGameFreeId(), s.Platform)
	if str != "" {
		jackpot := &base.SlotJackpotPool{}
		err := json.Unmarshal([]byte(str), jackpot)
		if err == nil {
			s.jackpot = jackpot
		}
	}
	if s.jackpot != nil {
		base.SlotsPoolMgr.SetPool(s.GetGameFreeId(), s.Platform, s.jackpot)
	} else {
		s.jackpot = &base.SlotJackpotPool{}
		jp := s.DbGameFree.GetJackpot()
		if len(jp) > 0 {
			s.jackpot.Small += int64(jp[0] * 10000)
		}
	}
}
func (s *RichBlessedSceneData) SaveLog(p *RichBlessedPlayerData, isOffline int) {
	if s.Testing {
		return
	}

}

func (s *RichBlessedSceneData) GetNormWeight() (norms [][]int32) {
	var key int
	curCoin := base.CoinPoolMgr.LoadCoin(s.GetGameFreeId(), s.Platform, s.GroupId)
	for i := len(s.DbGameFree.BalanceLine) - 1; i >= 0; i-- {
		balance := s.DbGameFree.BalanceLine[i]
		if curCoin >= int64(balance) {
			key = i
			break
		}
	}
	for _, norm := range srvdata.PBDB_SlotRateWeightMgr.Datas.GetArr() {
		if norm.GameFreeId == s.GetGameFreeId() && norm.Pos == int32(key) {
			norms = append(norms, norm.NormCol1)
			norms = append(norms, norm.NormCol2)
			norms = append(norms, norm.NormCol3)
			norms = append(norms, norm.NormCol4)
			norms = append(norms, norm.NormCol5)
			break
		}
	}
	return
}
func (s *RichBlessedSceneData) GetFreeWeight() (frees [][]int32) {
	var key int
	curCoin := base.CoinPoolMgr.LoadCoin(s.GetGameFreeId(), s.Platform, s.GroupId)
	for i := len(s.DbGameFree.BalanceLine) - 1; i >= 0; i-- {
		balance := s.DbGameFree.BalanceLine[i]
		if curCoin >= int64(balance) {
			key = i
			break
		}
	}
	for _, free := range srvdata.PBDB_SlotRateWeightMgr.Datas.GetArr() {
		if free.GameFreeId == s.GetGameFreeId() && free.Pos == int32(key) {
			frees = append(frees, free.FreeCol1)
			frees = append(frees, free.FreeCol2)
			frees = append(frees, free.FreeCol3)
			frees = append(frees, free.FreeCol4)
			frees = append(frees, free.FreeCol5)
			break
		}
	}
	return
}
