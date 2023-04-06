package fruits

import (
	"encoding/json"
	"fmt"
	"games.yol.com/win88/common"
	"games.yol.com/win88/gamerule/fruits"
	"games.yol.com/win88/gamesrv/base"
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	protocol "games.yol.com/win88/protocol/fruits"
	"games.yol.com/win88/srvdata"
	"github.com/idealeak/goserver/core/logger"
	"math"
	"math/rand"
	"time"
)

type FruitsSceneData struct {
	*base.Scene                             //场景
	players     map[int32]*FruitsPlayerData //玩家信息
	jackpot     *base.SlotJackpotPool       //奖池
	levelRate   [][]int32                   //调控概率区间
	//slotRateWeight      []int32                     //调控权重
	//slotRateWeightTotal [][]int32                   //总调控权重
	sysProfitCoinKey string
}

func NewFruitsSceneData(s *base.Scene) *FruitsSceneData {
	sceneEx := &FruitsSceneData{
		Scene:   s,
		players: make(map[int32]*FruitsPlayerData),
	}
	sceneEx.Init()
	return sceneEx
}
func (s *FruitsSceneData) Init() {
	s.LoadJackPotData()
	//for _, data := range srvdata.PBDB_SlotRateWeightMgr.Datas.Arr {
	//	if data.Id == s.DbGameFree.Id {
	//		//s.levelRate = append(s.levelRate, data.EleWeight1)
	//		//s.slotRateWeightTotal = append(s.slotRateWeightTotal, data.EleWeight1, data.EleWeight2, data.EleWeight3, data.EleWeight4, data.EleWeight5)
	//	}
	//}
	s.sysProfitCoinKey = fmt.Sprintf("%v_%v", s.Platform, s.GetGameFreeId())
}

func (s *FruitsSceneData) Clear() {
	//应该是水池变一次就判断修改一次
	//s.slotRateWeight = s.slotRateWeightTotal[0]
}
func (s *FruitsSceneData) SceneDestroy(force bool) {
	//销毁房间
	s.Scene.Destroy(force)
}
func (s *FruitsSceneData) AddPrizeCoin(isRob bool, totalBet int64) {
	addPrizeCoin := totalBet * fruits.NowByte * 1 / 100 //扩大10000倍
	s.jackpot.AddToSmall(isRob, addPrizeCoin)
	logger.Logger.Tracef("奖池增加...AddPrizeCoin... %f", float64(addPrizeCoin)/float64(10000))
	base.SlotsPoolMgr.SetPool(s.GetGameFreeId(), s.Platform, s.jackpot)
}
func (s *FruitsSceneData) DelPrizeCoin(isRob bool, win int64) {
	if win > 0 {
		s.jackpot.AddToSmall(isRob, -win*fruits.NowByte)
		logger.Logger.Tracef("奖池减少...AddPrizeCoin... %d", win)
		base.SlotsPoolMgr.SetPool(s.GetGameFreeId(), s.Platform, s.jackpot)
	}
}
func (s *FruitsSceneData) delPlayer(SnId int32) {
	if _, exist := s.players[SnId]; exist {
		delete(s.players, SnId)
	}
}
func (s *FruitsSceneData) OnPlayerLeave(p *base.Player, reason int) {
	if playerEx, ok := p.ExtraData.(*FruitsPlayerData); ok {
		if playerEx.freeTimes > 0 || playerEx.maryFreeTimes > 0 {
			//免费
			playerEx.winFreeCoin = playerEx.winLineCoin + playerEx.winJackPotCoin + playerEx.winEle777Coin
			//小玛丽
			playerEx.winCoin = playerEx.winMaryCoin + playerEx.winFreeCoin
			if playerEx.winCoin != 0 {
				//SysProfitCoinMgr.Add(s.sysProfitCoinKey, 0, playerEx.winCoin)
				p.Statics(s.KeyGameId, s.KeyGamefreeId, playerEx.winCoin, false)
				tax := int64(math.Ceil(float64(playerEx.winCoin) * float64(s.DbGameFree.GetTaxRate()) / 10000))
				playerEx.taxCoin = tax
				playerEx.winCoin -= tax
				p.AddServiceFee(tax)
				p.AddCoin(playerEx.winCoin, common.GainWay_HundredSceneWin, 0, "system", s.GetSceneName())
				if !s.Testing {
					base.CoinPoolMgr.PushCoin(s.GetGameFreeId(), s.GroupId, s.Platform, -(playerEx.winCoin + tax))
				}
				//p.isReportGameEvent = true
				//p.ReportGameEventParam()
				s.SaveLog(playerEx, 1)
			}
		}
	}
	s.delPlayer(p.SnId)
}
func (s *FruitsSceneData) Win(p *FruitsPlayerData) {
	if p.gameState != fruits.MaryGame && p.JackPotRate > 0 {
		//奖池
		win := s.jackpot.GetTotalSmall() / fruits.NowByte
		p.winNowJackPotCoin = win * p.JackPotRate / 100
		s.DelPrizeCoin(p.IsRob, p.winNowJackPotCoin)
		p.winJackPotCoin += p.winNowJackPotCoin
	}
	isBilled := false
	if p.gameState == fruits.Normal {
		isBilled = true
		p.winCoin = p.winLineCoin + p.winJackPotCoin + p.winEle777Coin
		p.winNowAllRate = p.winLineRate + p.JackPot7Rate
	} else if p.gameState == fruits.FreeGame {
		p.winFreeCoin = p.winLineCoin + p.winJackPotCoin + p.winEle777Coin
		p.winNowAllRate = p.winLineRate + p.JackPot7Rate
		if p.freeTimes == 0 {
			isBilled = true
			p.winCoin = p.winFreeCoin
		}
	} else {
		p.winNowAllRate = p.maryWinTotalRate
		if p.maryFreeTimes == 0 {
			isBilled = true
			p.winCoin = p.winMaryCoin
		}
	}
	if isBilled && p.winCoin != 0 {
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
	//p.ReportGameEventParam()
}
func (s *FruitsSceneData) SendBilled(p *FruitsPlayerData) {
	if p.gameState != fruits.MaryGame {
		//正常游戏 免费游戏
		var wl []*protocol.FruitsWinLine
		for _, r := range p.result.WinLine {
			wl = append(wl, &protocol.FruitsWinLine{
				Poss:   r.Poss,
				LineId: proto.Int(r.LineId),
			})
		}
		pack := &protocol.SCFruitsBilled{
			NowGameState:  proto.Int(p.gameState),
			BetIdx:        proto.Int(p.betIdx),
			Coin:          proto.Int64(p.Coin),
			FreeTotalWin:  proto.Int64(p.winFreeCoin),
			Jackpot:       proto.Int64(s.jackpot.GetTotalSmall() / 10000),
			WinJackpot:    proto.Int64(p.winNowJackPotCoin),
			WinEle777Coin: proto.Int64(p.oneBetCoin * p.JackPot7Rate),
			WinRate:       proto.Int64(p.winLineRate + p.JackPot7Rate),
			Cards:         p.result.EleValue,
			WinLines:      wl,
			FreeTimes:     proto.Int(p.freeTimes),
			MaryFreeTimes: proto.Int(p.maryFreeTimes),
			WinFreeTimes:  proto.Int(p.winFreeTimes),
			WinLineCoin:   proto.Int64(p.oneBetCoin * p.winLineRate),
		}
		logger.Logger.Trace("SCFruitsBilled：", pack)
		p.SendToClient(int(protocol.FruitsPID_PACKET_FRUITS_SCFruitsBilled), pack)
	} else {
		//玛丽游戏
		pack := &protocol.SCFruitsMaryBilled{
			Coin:          proto.Int64(p.Coin),
			FreeTotalWin:  proto.Int64(p.winFreeCoin),
			MaryTotalWin:  proto.Int64(p.winMaryCoin),
			MaryWinCoin:   proto.Int64(p.oneBetCoin * p.maryWinTotalRate),
			MaryWinId:     proto.Int32(p.result.MaryOutSide),
			MaryCards:     p.result.MaryMidArray,
			MaryFreeTimes: proto.Int(p.maryFreeTimes),
		}
		logger.Logger.Trace("SCFruitsMaryBilled：", pack)
		p.SendToClient(int(protocol.FruitsPID_PACKET_FRUITS_SCFruitsMaryBilled), pack)
	}
}
func (s *FruitsSceneData) LoadJackPotData() {
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
func (s *FruitsSceneData) SaveLog(p *FruitsPlayerData, isOffline int) {
	if s.Testing {
		return
	}
	s.SendPlayerBet(p)
	var betCoin int64
	var nowNRound int
	var nowGetCoin int64
	var isF bool

	if p.gameState == fruits.Normal {
		betCoin = p.betCoin
		nowGetCoin = p.winCoin
		isF = true
	} else if p.gameState == fruits.FreeGame {
		nowNRound = p.nowFreeTimes
		nowGetCoin = p.winFreeCoin
		if p.freeTimes == 0 {
			isF = true
		}
	} else {
		nowNRound = p.nowMaryTimes
		nowGetCoin = p.winMaryCoin
		if p.maryFreeTimes == 0 {
			isF = true
		}
	}
	Classic777Type := model.Classic777Type{
		RoomId:         s.SceneId,
		BasicScore:     int32(p.oneBetCoin),
		PlayerSnId:     p.SnId,
		BeforeCoin:     p.startCoin,
		AfterCoin:      p.Coin,
		ChangeCoin:     p.Coin - p.startCoin,
		TotalBetCoin:   betCoin,
		TotalLine:      9,
		TotalWinCoin:   nowGetCoin,
		NowGameState:   p.gameState,
		NowNRound:      nowNRound,
		IsOffline:      isOffline,
		FirstFreeTimes: p.freeTimes,
		MaryFreeTimes:  p.maryFreeTimes,
		//WhiteLevel:      p.WhiteLevel,
		//BlackLevel:      p.BlackLevel,
		TaxCoin: p.taxCoin,
	}
	if p.result.JackPotNum >= 3 {
		Classic777Type.JackPotNum = p.result.JackPotNum
		Classic777Type.JackPotWinCoin = p.oneBetCoin * p.JackPot7Rate
	}
	var winLine []model.Classic777WinLine
	if p.gameState == fruits.MaryGame {
		Classic777Type.MaryOutSide = fruits.MaryEleArray[p.result.MaryOutSide]
		Classic777Type.MaryMidCards = p.result.MaryMidArray
		winLine = append(winLine, model.Classic777WinLine{
			Id:          -1,
			EleValue:    Classic777Type.MaryOutSide,
			Num:         p.maryInNum,
			Rate:        p.maryWinTotalRate,
			WinCoin:     p.oneBetCoin * p.maryWinTotalRate,
			WinFreeGame: -1,
		})

	} else {
		Classic777Type.Cards = p.result.EleValue
		Classic777Type.HitPrizePool = p.winNowJackPotCoin
		n := 0
		for _, v := range p.result.WinLine {
			if v.Rate == 0 {
				n++
			}
		}
		Classic777Type.WinLineNum = len(p.result.WinLine) - n
		Classic777Type.WinLineRate = p.winLineRate
		Classic777Type.WinLineCoin = p.winLineRate * p.oneBetCoin
		for _, line := range p.result.WinLine {
			if line.LineId > 9 {
				continue
			}
			flag := line.Lines[0]
			if flag == fruits.Wild {
				for _, v := range line.Lines {
					if flag != v {
						flag = v
						break
					}
				}
			} else if flag == fruits.Bonus {
				continue
			}
			fw := model.Classic777WinLine{
				Id:          line.LineId,
				EleValue:    flag,
				Num:         len(line.Lines),
				Rate:        line.Rate,
				WinCoin:     line.Rate * p.oneBetCoin,
				WinFreeGame: -1,
			}
			winLine = append(winLine, fw)
		}
	}
	winLine = append(winLine, p.wl...)
	Classic777Type.WinLine = winLine
	info, err := model.MarshalGameNoteByROLL(&Classic777Type)
	if err == nil {
		logId, _ := model.AutoIncGameLogId()
		s.SaveGameDetailedLog(logId, info, &base.GameDetailedParam{})
		//水池上下文环境s
		s.CpCtx = p.cpCtx
		var totalIn, totalOut int64
		if betCoin > 0 {
			totalIn = betCoin
		}
		if nowGetCoin > 0 && isF {
			totalOut = p.Coin - p.startCoin + betCoin + p.taxCoin
		}
		s.SaveGamePlayerListLog(p.SnId,
			&base.SaveGamePlayerListLogParam{
				Platform:          p.Platform,
				Channel:           p.Channel,
				Promoter:          p.BeUnderAgentCode,
				PackageTag:        p.PackageID,
				InviterId:         p.InviterId,
				LogId:             logId,
				TotalIn:           totalIn,
				TotalOut:          totalOut,
				TaxCoin:           p.taxCoin,
				ClubPumpCoin:      0,
				BetAmount:         totalIn,
				WinAmountNoAnyTax: p.Coin - p.startCoin,
				IsFirstGame:       s.IsPlayerFirst(p.Player),
			})
	}
	s.GameNowTime = time.Now()
	if s.CheckNeedDestroy() && p.freeTimes == 0 && p.maryFreeTimes == 0 {
		s.PlayerLeave(p.Player, common.PlayerLeaveReason_OnDestroy, true)
	}
}
func (s *FruitsSceneData) SendPlayerBet(p *FruitsPlayerData) {
	//统计输下注金币数
	//if !p.IsRob && !s.Testing {
	//	betCoin := p.betCoin
	//	if p.gameState != fruits.Normal {
	//		betCoin = 0
	//	}
	//	playerBet := &protocol.PlayerBet{
	//		SnId: proto.Int32(p.SnId),
	//		Bet:  proto.Int64(betCoin),
	//		Gain: proto.Int64(p.coin - p.startCoin),
	//		Tax:  proto.Int64(p.taxCoin),
	//	}
	//	gwPlayerBet := &protocol.GWPlayerBet{
	//		GameFreeId: proto.Int32(s.dbGameFree.GetId()),
	//		RobotGain:  proto.Int64(-(p.coin - p.startCoin + p.taxCoin)),
	//	}
	//	gwPlayerBet.PlayerBets = append(gwPlayerBet.PlayerBets, playerBet)
	//	proto.SetDefaults(gwPlayerBet)
	//	s.SendToWorld(int(protocol.MmoPacketID_PACKET_GW_PLAYERBET), gwPlayerBet)
	//	logger.Logger.Trace("Send msg gwPlayerBet ===>", gwPlayerBet)
	//}
}
func (s *FruitsSceneData) Regulation(p *FruitsPlayerData, n int) {
	//if len(s.levelRate) == 0 || len(s.slotRateWeight) == 0 {
	//	return
	//}
	//if p.gameState != fruits.Normal {
	//	state, _ := base.CoinPoolMgr.GetCoinPoolStatus2(s.Platform, s.GetGameFreeId(), s.GroupId, -p.preWinRate*p.oneBetCoin)
	//	var isLow bool
	//	if state == base.CoinPoolStatus_Low {
	//		isLow = true
	//		if p.preWinRate > 9 {
	//			p.preWinRate = rand.Int63n(9) + 1
	//		}
	//	}
	//	s.InFreeGame(p, isLow)
	//	return
	//}
	//setting := base.CoinPoolMgr.GetCoinPoolSetting(s.Platform, s.GetGameFreeId(), s.GroupId)
	//state, _ := base.CoinPoolMgr.GetCoinPoolStatus(s.Platform, s.GetGameFreeId(), s.GroupId)
	//isLow := false
	//if state == base.CoinPoolStatus_Low {
	//	//test code
	//	isLow = true
	//}
	//var levelMaxOut = -1
	//var level = -1
	//if setting != nil {
	//	for k, v := range s.levelRate {
	//		if setting.GetMaxOutValue() != 0 && int64(v[1]) <= int64(setting.GetMaxOutValue())/p.oneBetCoin {
	//			levelMaxOut = k
	//		}
	//		if isLow {
	//			if int64(v[1]) <= 9 {
	//				level = k
	//			}
	//		}
	//	}
	//}
	//if !isLow && levelMaxOut > level {
	//	level = levelMaxOut
	//}
	////调控Y值
	//var regY int32 = -1
	//if p.WhiteLevel > 0 || p.BlackLevel > 0 {
	//	if p.WhiteLevel > 0 {
	//		regY = s.slotRateWeight[0] * (10 - p.WhiteLevel)
	//	} else if p.BlackLevel > 0 {
	//		regY = s.slotRateWeight[0] * (10 + p.BlackLevel*5) / 10
	//	}
	//} else {
	//	if d, exist := p.GDatas[s.KeyGameId]; exist {
	//		if d.Statics.TotalIn < 20000 {
	//			regY = s.slotRateWeight[0] * 50 / 100
	//		}
	//		if d.Statics.TotalIn != 0 && d.Statics.TotalOut*100/d.Statics.TotalIn >= 150 && regY == -1 {
	//			regY = s.slotRateWeight[0] * 4
	//		}
	//	}
	//	if regY == -1 {
	//		if setting != nil {
	//			poolValue := base.CoinPoolMgr.LoadCoin(s.GetGameFreeId(), s.Platform, s.GroupId)
	//			w := float64(s.slotRateWeight[0]) * (1 - float64(int32(poolValue)-setting.GetLowerLimit())/float64(setting.GetUpperLimit()-setting.GetLowerLimit()+1))
	//			if w < 0 {
	//				regY = 0
	//			} else {
	//				regY = int32(w)
	//			}
	//		}
	//	}
	//	if p.noWinTimes > 0 && regY == -1 {
	//		if rand.Intn(100) < 10*p.noWinTimes {
	//			//要中奖
	//			regY = s.slotRateWeight[0] * (10 - int32(p.noWinTimes)) * 10 / 100
	//		}
	//	}
	//}
	//if regY == -1 {
	//	regY = s.slotRateWeight[0]
	//}
	//slotRateWeight := make([]int32, level+1)
	//copy(slotRateWeight, s.slotRateWeight[:level+1])
	//slotRateWeight[0] = regY
	//rIdx := fruits.RandSliceInt32IndexByWightN(slotRateWeight)
	//parsec := s.levelRate[rIdx]
	//var needRate int32
	//if rIdx != 0 {
	//	needRate = rand.Int31n(parsec[1]-parsec[0]) + parsec[0]
	//}
	//state, _ = base.CoinPoolMgr.GetCoinPoolStatus2(s.Platform, s.GetGameFreeId(), s.GroupId, -int64(needRate)*p.oneBetCoin)
	//if n >= 5 {
	//	needRate = rand.Int31n(9)
	//}
	////test code
	////if state == base.CoinPoolStatus_Low && needRate > 9 {
	////	s.Regulation(p, n+1)
	////	return
	////}
	//p.preWinRate = int64(needRate)
	//p.preNeedRate = int64(needRate)
	//if p.preNeedRate > 1000 {
	//	p.pre30WinRate = int64(needRate) * 30 / 100
	//}
	//s.ShowGameState(int64(needRate), p, isLow)
}
func (s *FruitsSceneData) InFreeGame(p *FruitsPlayerData, isLow bool) {
	if p.gameState == fruits.FreeGame {
		var nr int64
		if p.preWinRate > 200 {
			nr = rand.Int63n(100-50) + 50
		} else {
			nr = p.preWinRate
		}
		s.ShowGameState(nr, p, isLow)
	} else if p.gameState == fruits.MaryGame {
		s.InMaryGame(p)
	}
}
func (s *FruitsSceneData) InMaryGame(p *FruitsPlayerData) {
	ele := []int32{fruits.Watermelon, fruits.Grape, fruits.Lemon, fruits.Cherry, fruits.Banana, fruits.Bonus, fruits.Pineapple}
	var eleVal int32 = -1
	var lianxuEle = 0
	if p.preWinRate > 550*9 {
		//四连
		rate := []int64{700 * 9, 600 * 9, 570 * 9, 550 * 9, 520 * 9, 510 * 9, 505 * 9}
		for k, v := range rate {
			if p.preWinRate > v {
				ele = ele[k:]
				break
			}
		}
		lianxuEle = 4
	} else if p.preWinRate > 80*9 {
		//三连
		rate := []int64{220 * 9, 120 * 9, 90 * 9, 70 * 9, 40 * 9, 30 * 9, 25 * 9}
		for k, v := range rate {
			if p.preWinRate > v {
				ele = ele[k:]
				break
			}
		}
		lianxuEle = 3
	} else if p.preWinRate > 5*9 {
		//单个
		rate := []int64{200 * 9, 100 * 9, 70 * 9, 50 * 9, 20 * 9, 10 * 9, 5 * 9}
		for k, v := range rate {
			if p.preWinRate > v {
				ele = ele[k:]
				break
			}
		}
		lianxuEle = 1
	}
	p.preMaryMidArray = make([]int32, 4)
	MaryMidEleRate := make([]int32, len(fruits.MaryMidEleRate))
	copy(MaryMidEleRate, fruits.MaryMidEleRate)
	for i := 0; i < 4; i++ {
		eler := fruits.RandSliceInt32IndexByWightN(MaryMidEleRate)
		p.preMaryMidArray[i] = eler
	}
	if lianxuEle > 0 {
		//需要中奖
		idx := rand.Intn(len(ele))
		eleVal = ele[idx]
		ele = append(ele[:idx], ele[idx+1:]...)
		for k, v := range p.preMaryMidArray {
			if v == eleVal {
				p.preMaryMidArray[k] = ele[rand.Intn(len(ele))]
			}
		}
		if lianxuEle > 1 {
			for i := 0; i < lianxuEle; i++ {
				p.preMaryMidArray[i] = eleVal
			}
		} else {
			for k := range p.preMaryMidArray {
				if rand.Intn(300) <= (k+1)*100 {
					p.preMaryMidArray[k] = eleVal
					break
				}
			}
		}
		for k, v := range fruits.MaryEleArray {
			if v == eleVal && rand.Intn(300) <= (k+1)*100 {
				p.preMaryOutSide = int32(k)
				break
			}
		}
		var rate int32
		switch eleVal {
		case fruits.Cherry:
			rate += 50
		case fruits.Pineapple:
			rate += 5
		case fruits.Grape:
			rate += 100
		case fruits.Lemon:
			rate += 70
		case fruits.Watermelon:
			rate += 200
		case fruits.Banana:
			rate += 20
		case fruits.Apple:
			rate += 10
		}
		if lianxuEle == 3 {
			rate += 20
		} else if lianxuEle == 4 {
			rate += 500
		}
		p.preWinRate -= int64(rate) * 9
	} else {
		//不需要中
		for _, v := range p.preMaryMidArray {
			for m, n := range ele {
				if n == v {
					ele = append(ele[:m], ele[m+1:]...)
					break
				}
			}
		}
		idxEle := ele[rand.Intn(len(ele))]
		for k, v := range fruits.MaryEleArray {
			if v == idxEle && rand.Intn(100) > 50 {
				p.preMaryOutSide = int32(k)
				break
			}
		}
	}
}
func (s *FruitsSceneData) ShowGameState(needRate int64, p *FruitsPlayerData, isLow bool) {
	var isCanFree bool
	var isCanMary bool
	//免费触发免费
	var isCanFree2 bool
	var bonusNum int
	var maryNum int
	if p.gameState == fruits.Normal {
		if needRate > 200 {
			if needRate < 2000 {
				//触发免费
				isCanFree = true
				needRate = rand.Int63n(100)
				bonusNum = 3
			} else {
				bonusNum = rand.Intn(2) + 4
				if rand.Intn(100) > 50 {
					//触发免费
					isCanFree = true
					needRate = rand.Int63n(100)
				} else {
					isCanMary = true
					needRate = 0
					maryNum = rand.Intn(3) + 3
				}
			}
		}
	} else if p.gameState == fruits.FreeGame {
		if p.nowFreeTimes == 6 {
			if p.preNeedRate > 1000 && p.preWinRate >= p.pre30WinRate {
				isCanFree2 = true
				needRate = 0
				bonusNum = 3
			}
		} else if p.nowFreeTimes > 6 && p.freeTimes == 0 {
			needRate = p.preWinRate
		}
	}
	if needRate < 0 {
		needRate = 0
	}
	//wl, noPoss, preInt := fruits.GetLineEleVal(p.gameState, needRate, s.DbGameFree.GetElementsParams(), isLow)
	wl, noPoss, preInt := fruits.GetLineEleVal(p.gameState, needRate, ElementsParams, isLow)
	if noPoss == nil && len(preInt) > 0 {
		wl.Init()
		wl.EleValue = preInt[rand.Intn(len(preInt))]
		wl.Win()
		var poss []int32
		for _, v := range wl.WinLine {
			poss = append(poss, v.Poss...)
		}
		for k := range wl.EleValue {
			for _, n := range poss {
				if k == int(n) {
					break
				}
			}
			noPoss = append(noPoss, k)
		}
	}
	p.preEleVal = wl.EleValue
	if p.gameState == fruits.Normal {
		var rate int64
		for _, v := range wl.WinLine {
			rate += v.Rate
		}
		//JackPot7 计算
		var JackPot7Rate int64
		var JackPotRate int64
		if wl.JackPotNum >= 3 {
			if wl.JackPotNum == 3 {
				JackPotRate = 5
				JackPot7Rate = 100
			} else if wl.JackPotNum == 4 {
				JackPotRate = 10
				JackPot7Rate = 200
			} else if wl.JackPotNum == 5 {
				JackPotRate = 15
				JackPot7Rate = 1750
			}
		}
		win := s.jackpot.GetTotalSmall() / fruits.NowByte
		winNowJackPotCoin := win * JackPotRate / 100
		jackPot := JackPot7Rate + winNowJackPotCoin/9
		if p.preWinRate-rate-jackPot < 0 {
			for i := 0; i < 100; i++ {
				//wl.CreateLine(s.DbGameFree.GetElementsParams())
				wl.CreateLine(ElementsParams)
				wl.Win()
				rate = 0
				for _, v := range wl.WinLine {
					rate += v.Rate
				}
				if rate <= 9 && wl.JackPotNum < 3 {
					break
				}
			}
		}
		copy(p.preEleVal, wl.EleValue)
		return
	}
	if isCanFree || isCanFree2 {
		lineId := rand.Intn(9) + 1
		if lineId > 0 {
			poss := fruits.GetLinePos(lineId)
			for k, pos := range poss {
				if k < bonusNum {
					wl.EleValue[pos] = fruits.Bonus
				}
			}
		}
		wl.WinLine = nil
		wl.JackPotNum = 0
		wl.Win()
		var rate int64
		for _, v := range wl.WinLine {
			if v.LineId != lineId {
				rate += v.Rate
			}
		}
		//JackPot7 计算
		var JackPot7Rate int64
		var JackPotRate int64
		if wl.JackPotNum >= 3 {
			if wl.JackPotNum == 3 {
				JackPotRate = 5
				JackPot7Rate = 100
			} else if wl.JackPotNum == 4 {
				JackPotRate = 10
				JackPot7Rate = 200
			} else if wl.JackPotNum == 5 {
				JackPotRate = 15
				JackPot7Rate = 1750
			}
		}
		win := s.jackpot.GetTotalSmall() / fruits.NowByte
		winNowJackPotCoin := win * JackPotRate / 100
		jackPot := JackPot7Rate + winNowJackPotCoin/9
		if p.preWinRate-rate-jackPot < 0 {
			jackPot = 0
			for i := 0; i < 100; i++ {
				//wl.CreateLine(s.DbGameFree.GetElementsParams())
				wl.CreateLine(ElementsParams)
				wl.Win()
				rate = 0
				for _, v := range wl.WinLine {
					rate += v.Rate
				}
				if rate <= 9 && wl.JackPotNum < 3 {
					break
				}
			}
		}
		copy(p.preEleVal, wl.EleValue)
		p.preWinRate -= rate + jackPot
	} else if isCanMary {
		lineId := rand.Intn(9) + 1
		if lineId > 0 {
			poss := fruits.GetLinePos(lineId)
			for k, pos := range poss {
				if k < maryNum {
					wl.EleValue[pos] = fruits.Wild
				}
			}
		}
		mary := fruits.WinResult{}
		//mary.CreateLine(s.DbGameFree.GetElementsParams())
		mary.CreateLine(ElementsParams)
		copy(mary.EleValue, wl.EleValue)
		mary.Win()
		//线计算
		var rate int64
		for _, v := range mary.WinLine {
			rate += v.Rate
		}
		//JackPot7 计算
		var JackPot7Rate int64
		var JackPotRate int64
		if mary.JackPotNum >= 3 {
			if mary.JackPotNum == 3 {
				JackPotRate = 5
				JackPot7Rate = 100
			} else if mary.JackPotNum == 4 {
				JackPotRate = 10
				JackPot7Rate = 200
			} else if mary.JackPotNum == 5 {
				JackPotRate = 15
				JackPot7Rate = 1750
			}
		}
		win := s.jackpot.GetTotalSmall() / fruits.NowByte
		winNowJackPotCoin := win * JackPotRate / 100
		jackPot := JackPot7Rate + winNowJackPotCoin/9
		if p.preWinRate-rate-jackPot < 0 {
			jackPot = 0
			for i := 0; i < 100; i++ {
				//mary.CreateLine(s.DbGameFree.GetElementsParams())
				mary.CreateLine(ElementsParams)
				mary.Win()
				rate = 0
				for _, v := range mary.WinLine {
					rate += v.Rate
				}
				if rate <= 9 && mary.JackPotNum < 3 {
					break
				}
			}
		}
		copy(p.preEleVal, mary.EleValue)
		p.preWinRate -= rate + jackPot
	}
}
func (s *FruitsSceneData) GetNormWeight() (norms [][]int32) {
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
func (s *FruitsSceneData) GetFreeWeight() (frees [][]int32) {
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
func (s *FruitsSceneData) GetMaryWeight() (marys [][]int32) {

	var key int
	curCoin := base.CoinPoolMgr.LoadCoin(s.GetGameFreeId(), s.Platform, s.GroupId)
	for i := len(s.DbGameFree.BalanceLine) - 1; i >= 0; i-- {
		balance := s.DbGameFree.BalanceLine[i]
		if curCoin >= int64(balance) {
			key = i
			break
		}
	}
	for _, mary := range srvdata.PBDB_SlotRateWeightMgr.Datas.GetArr() {
		if mary.GameFreeId == s.GetGameFreeId() && mary.Pos == int32(key) {
			marys = append(marys, mary.MaryOut)
			marys = append(marys, mary.MaryMid)
			break
		}
	}
	return
}
