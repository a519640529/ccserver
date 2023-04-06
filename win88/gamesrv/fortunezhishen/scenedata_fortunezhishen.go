package fortunezhishen

import (
	"encoding/json"
	"fmt"
	"games.yol.com/win88/common"
	gamerule "games.yol.com/win88/gamerule/fortunezhishen"
	"games.yol.com/win88/gamesrv/base"
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/fortunezhishen"
	server_proto "games.yol.com/win88/protocol/server"
	//"games.yol.com/win88/srvdata"
	"github.com/idealeak/goserver/core/logger"
	"math"
	"math/rand"
	"time"
)

//FortuneZhiShen
type FortuneZhiShenSceneData struct {
	*base.Scene                                          //场景
	players          map[int32]*FortuneZhiShenPlayerData //玩家信息
	jackpot          *base.SlotJackpotPool               //奖池
	levelRate        [][]int32                           //调控概率区间
	slotRateWeight   []int32                             //调控权重
	sysProfitCoinKey string
}

func NewFortuneZhiShenSceneData(s *base.Scene) *FortuneZhiShenSceneData {
	sceneEx := &FortuneZhiShenSceneData{
		Scene:   s,
		players: make(map[int32]*FortuneZhiShenPlayerData),
	}
	sceneEx.Init()
	return sceneEx
}

func (s *FortuneZhiShenSceneData) Init() {
	//s.EleLineAppearRate = []int32{337, 562, 562, 562, 562, 562, 899, 1124, 1124, 1124, 1124, 1124, 337}
	s.LoadJackPotData()
	//for _, data := range srvdata.PBDB_SlotRateMgr.Datas.Arr {
	//if int(data.GetGameId()) == common.GameId_FortuneZhiShen {
	//	s.levelRate = append(s.levelRate, data.RateSection)
	//	s.slotRateWeight = append(s.slotRateWeight, data.GetWeight())
	//}
	//}
	s.sysProfitCoinKey = fmt.Sprintf("%v_%v", s.GetPlatform(), s.GetGameFreeId())
}
func (s *FortuneZhiShenSceneData) CreateResult(p *FortuneZhiShenPlayerData) {
	//dbGameFree := s.GetDBGameFree()
	ele := make([]int32, 0)
	//ele := make([]int32, len(dbGameFree.GetElementsParams()))
	//copy(ele, dbGameFree.GetElementsParams())
	if s.GetTesting() {
		ele[gamerule.MakeAFortune] = 1000
		ele[gamerule.Gemstone] = 1000
	} else {
		if p.gameState == gamerule.StopAndRotate || p.gameState == gamerule.StopAndRotate2 {
			ele[gamerule.Gemstone] = 0
		}
	}
	p.CreateResult(ele)
	p.result.Win(p.betIdx, p.gameState)
	//调控
	if p.gameState == gamerule.Normal {
		if p.result.GemstoneNum >= 6 {
			var total = s.GetGemstoneTotalRate(p, false)
			r := p.preWinRate / 50 / int64(p.result.GemstoneNum)
			if total > p.preWinRate && r > 0 {
				for k := range p.result.GemstoneRate {
					if p.result.Cards[k] == gamerule.Gemstone {
						p.result.GemstoneRate[k] = rand.Int63n(r) + 1
					}
				}
			}
		} else if p.result.GemstoneNum > 0 {
			var total = s.GetGemstoneTotalRate(p, true)
			state, _ := base.GetCoinPoolMgr().GetCoinPoolStatus2(s.GetPlatform(), s.GetGameFreeId(), s.GetGroupId(), -total*p.oneBetCoin)
			if state == base.CoinPoolStatus_Low {
				for k, v := range p.result.GemstoneRate {
					if v == gamerule.GrandPrize || v == gamerule.BigPrize {
						p.result.GemstoneRate[k] = rand.Int63n(19) + 1
					}
				}
			}
		}
	} else if p.gameState == gamerule.FreeGame {
		if p.result.MidIcon == gamerule.Gemstone {
			var total = s.GetGemstoneTotalRate(p, false)
			if p.preWinRate < total {
				n := p.result.GemstoneNum - 9
				rCount := p.preWinRate / 50 / int64(n+1)
				//宝石数值随机
				midGe := p.result.GetGemstoneRate(p.betIdx, true)
				if midGe > rCount {
					midGe = rCount
				}
				for k := range p.result.GemstoneRate {
					if p.result.Cards[k] == gamerule.Gemstone {
						var randGem int64
						if k != 1 && k != 2 && k != 3 &&
							k != 6 && k != 7 && k != 8 &&
							k != 11 && k != 12 && k != 13 {
							randGem = p.result.GetGemstoneRate(p.betIdx, false)
							if randGem > rCount {
								randGem = rCount
							}
						} else {
							randGem = midGe
						}
						if randGem <= 0 {
							randGem = 1
						}
						p.result.GemstoneRate[k] = randGem
					}
				}
			}
		}
	}
	if p.gameState == gamerule.Normal {
		p.normalCards = make([]int32, 15)
		copy(p.normalCards, p.result.Cards)
	}
}
func (s *FortuneZhiShenSceneData) GetGemstoneTotalRate(p *FortuneZhiShenPlayerData, isLT6 bool) int64 {
	var total int64
	for k, v := range p.result.GemstoneRate {
		if p.result.MidIcon == gamerule.Gemstone &&
			(k == 2 || k == 3 || k == 6 || k == 7 || k == 8 || k == 11 || k == 12 || k == 13) {
			continue
		}
		if v == gamerule.SmallPrize {
			if !isLT6 {
				total += 20 * 50
			}
		} else if v == gamerule.MidPrize {
			if !isLT6 {
				total += 100 * 50
			}
		} else if v == gamerule.BigPrize {
			bigPrize := s.jackpot.BigPools[p.betIdx] / gamerule.NowByte
			win := bigPrize * 10 / 100
			total += win / p.oneBetCoin
		} else if v == gamerule.GrandPrize {
			grandPrize := s.jackpot.HugePools[p.betIdx] / gamerule.NowByte
			win := grandPrize * 10 / 100
			total += win / p.oneBetCoin
		} else if !isLT6 {
			total += v * 50
		}
	}
	return total
}
func (s *FortuneZhiShenSceneData) SendPrize(p *FortuneZhiShenPlayerData) {
	var grandPrize, bigPrize, midPrize, smallPrize []int64
	for i := 0; i < len(s.GetDBGameFree().GetOtherIntParams()); i++ {
		grandPrize = append(grandPrize, s.jackpot.HugePools[i]/gamerule.NowByte)
		bigPrize = append(bigPrize, s.jackpot.BigPools[i]/gamerule.NowByte)
		midPrize = append(midPrize, s.jackpot.MiddlePools[i]/gamerule.NowByte)
		smallPrize = append(smallPrize, s.jackpot.SmallPools[i]/gamerule.NowByte)
	}
	pack := &fortunezhishen.SCFortuneZhiShenPrize{
		GrandPrize: grandPrize,
		BigPrize:   bigPrize,
		MidPrize:   midPrize,
		SmallPrize: smallPrize,
	}
	proto.SetDefaults(pack)
	//logger.Logger.Trace("SCFortuneZhiShenPrize: ", pack)
	p.SendToClient(int(fortunezhishen.FortuneZSPacketID_PACKET_SC_FORTUNEZHISHEN_PRIZE), pack)
}
func (s *FortuneZhiShenSceneData) InFreeGame(p *FortuneZhiShenPlayerData) {
	if p.gameState == gamerule.FreeGame {
		var nr int64
		if p.preWinRate > 200 {
			if p.preWinRate >= 600 {
				nr = rand.Int63n(600-200) + 200
			} else if rand.Intn(60) <= p.nowFirstTimes*10 {
				nr = rand.Int63n(p.preWinRate-200) + 200
			}
		} else {
			nr = p.preWinRate
		}
		//p.ShowGameState(nr, s.GetDBGameFree().GetElementsParams())
		p.ShowGameState(nr, []int32{})
	} else if p.gameState == gamerule.StopAndRotate || p.gameState == gamerule.StopAndRotate2 {
		if 15-p.result.GemstoneNum > 0 {
			num := rand.Intn(3) + 1
			for i := 0; i < num; i++ {
				var total = s.GetGemstoneTotalRate(p, false)
				dif := p.preWinRate - total
				var r int64
				for si := 0; si < 10; si++ {
					rz := p.result.GetGemstoneRate(p.betIdx, false)
					r = rz
					if rz == gamerule.SmallPrize {
						rz = 20 * 50
					} else if rz == gamerule.MidPrize {
						rz = 100 * 50
					} else if rz == gamerule.BigPrize {
						bigPrize := s.jackpot.BigPools[p.betIdx] / gamerule.NowByte
						win := bigPrize * 10 / 100
						rz = win / p.oneBetCoin
					} else if rz == gamerule.GrandPrize {
						grandPrize := s.jackpot.HugePools[p.betIdx] / gamerule.NowByte
						win := grandPrize * 10 / 100
						rz = win / p.oneBetCoin
					} else {
						rz *= 50
					}
					var prize int64
					if p.result.GemstoneNum == 14 {
						//15个会额外给巨奖
						grandPrize := s.jackpot.HugePools[p.betIdx] / gamerule.NowByte
						win := grandPrize * 10 / 100
						prize = win / p.oneBetCoin
					}
					if dif-rz-prize > 0 {
						dif -= rz
						break
					} else {
						r = 0
					}
				}
				if dif > 0 && r > 0 {
					var noGePos []int
					for k, v := range p.result.Cards {
						if v != gamerule.Gemstone {
							noGePos = append(noGePos, k)
						}
					}
					if len(noGePos) > 0 {
						pos := noGePos[rand.Intn(len(noGePos))]
						p.result.Cards[pos] = gamerule.Gemstone
						p.result.GemstoneRate[pos] = r
						p.newAddGemstone++
						p.result.GemstoneNum++
					}
				}
			}
		}
	}
}

func (s *FortuneZhiShenSceneData) Regulation(p *FortuneZhiShenPlayerData, n int) {
	if len(s.levelRate) == 0 || len(s.slotRateWeight) == 0 {
		return
	}
	if p.gameState != gamerule.Normal {
		state, _ := base.GetCoinPoolMgr().GetCoinPoolStatus2(s.GetPlatform(), s.GetGameFreeId(), s.GetGroupId(), -p.preWinRate*p.oneBetCoin)
		if state == base.CoinPoolStatus_Low && p.preWinRate > 50 {
			p.preWinRate = rand.Int63n(50)
		}
		s.InFreeGame(p)
		return
	}
	setting := base.GetCoinPoolMgr().GetCoinPoolSetting(s.GetPlatform(), s.GetGameFreeId(), s.GetGroupId())
	state, _ := base.GetCoinPoolMgr().GetCoinPoolStatus(s.GetPlatform(), s.GetGameFreeId(), s.GetGroupId())
	isLow := false
	if state == base.CoinPoolStatus_Low {
		isLow = true
	}
	var levelMaxOut = -1
	var level = -1
	if setting != nil {
		for k, v := range s.levelRate {
			if setting.GetMaxOutValue() != 0 && int64(v[1]) <= int64(setting.GetMaxOutValue())/p.oneBetCoin {
				levelMaxOut = k
			}
			if isLow {
				if int64(v[1]) <= 50 {
					level = k
				}
			}
		}
	}
	if !isLow && levelMaxOut > level {
		level = levelMaxOut
	}
	//调控Y值
	var regY int32 = -1
	if p.WhiteLevel > 0 || p.BlackLevel > 0 {
		if p.WhiteLevel > 0 {
			regY = s.slotRateWeight[0] * (10 - p.WhiteLevel)
		} else if p.BlackLevel > 0 {
			regY = s.slotRateWeight[0] * (10 + p.BlackLevel*5) / 10
		}
	} else {
		if d, exist := p.GDatas[s.GetKeyGameId()]; exist {
			if d.Statics.TotalIn < 20000 {
				regY = s.slotRateWeight[0] * 50 / 100
			}
			if d.Statics.TotalIn != 0 && d.Statics.TotalOut*100/d.Statics.TotalIn >= 150 && regY == -1 {
				regY = s.slotRateWeight[0] * 2
			}
		}
		if regY == -1 {
			if setting != nil {
				poolValue := base.GetCoinPoolMgr().LoadCoin(s.GetGameFreeId(), s.GetPlatform(), s.GetGroupId())
				w := float64(s.slotRateWeight[0]) * (1 - float64(int32(poolValue)-setting.GetLowerLimit())/float64(setting.GetUpperLimit()-setting.GetLowerLimit()+1))
				if w < 0 {
					regY = 0
				} else {
					regY = int32(w)
				}
			}
		}
		if p.noWinTimes > 0 && regY == -1 {
			if rand.Intn(100) < 10*p.noWinTimes {
				//要中奖
				regY = s.slotRateWeight[0] * (10 - int32(p.noWinTimes)) * 10 / 100
			}
		}
	}
	if regY == -1 {
		regY = s.slotRateWeight[0]
	}
	slotRateWeight := make([]int32, level+1)
	copy(slotRateWeight, s.slotRateWeight[:level+1])
	slotRateWeight[0] = regY
	rIdx := gamerule.RandSliceInt32IndexByWightN(slotRateWeight)
	parsec := s.levelRate[rIdx]
	var needRate int32
	if rIdx != 0 {
		needRate = rand.Int31n(parsec[1]-parsec[0]) + parsec[0]
	}
	state, _ = base.GetCoinPoolMgr().GetCoinPoolStatus2(s.GetPlatform(), s.GetGameFreeId(), s.GetGroupId(), -int64(needRate)*p.oneBetCoin)
	if n >= 5 {
		needRate = rand.Int31n(50)
	}
	//test code
	if state == base.CoinPoolStatus_Low && needRate > 50 {
		s.Regulation(p, n+1)
		return
	}
	p.preWinRate = int64(needRate)
	p.preNeedRate = int64(needRate)
	if p.preNeedRate > 3000 {
		p.pre70WinRate = int64(needRate) * 70 / 100
	}
	//p.ShowGameState(int64(needRate), s.GetDBGameFree().GetElementsParams())
	p.ShowGameState(int64(needRate), []int32{})
}
func (s *FortuneZhiShenSceneData) SendRoomState(p *FortuneZhiShenPlayerData) {
	pack := &fortunezhishen.SCFortuneZhiShenRoomState{
		State:    proto.Int(s.GetSceneState().GetState()),
		SubState: proto.Int(p.gameState),
	}
	proto.SetDefaults(pack)
	p.SendToClient(int(fortunezhishen.FortuneZSPacketID_PACKET_SC_FORTUNEZHISHEN_ROOMSTATE), pack)
}
func (s *FortuneZhiShenSceneData) SceneDestroy(force bool) {
	//销毁房间
	s.Scene.Destroy(force)
}
func (s *FortuneZhiShenSceneData) AddPrizeCoin(totalBet int64, idx int) {
	addPrizeCoin := totalBet * gamerule.NowByte * 1 / 100
	jackpot := s.jackpot
	if jackpot != nil {
		jackpot.BigPools[idx] += addPrizeCoin
		jackpot.HugePools[idx] += addPrizeCoin
	}
	logger.Logger.Tracef("奖池增加...AddPrizeCoin... %f", float64(addPrizeCoin)/float64(gamerule.NowByte))
	base.SlotsPoolMgr.SetPool(s.GetGameFreeId(), s.GetPlatform(), s.jackpot)
}
func (s *FortuneZhiShenSceneData) DelPrizeCoin(win int64, idx int, idxPool int) {
	jackpot := s.jackpot
	if jackpot != nil && win > 0 {
		if idxPool == gamerule.GrandPrize {
			jackpot.HugePools[idx] -= win * gamerule.NowByte
		} else if idxPool == gamerule.BigPrize {
			jackpot.BigPools[idx] -= win * gamerule.NowByte
		}
		base.SlotsPoolMgr.SetPool(s.GetGameFreeId(), s.GetPlatform(), s.jackpot)
	}
	logger.Logger.Tracef("奖池减少...AddPrizeCoin... %d", win)
}
func (s *FortuneZhiShenSceneData) delPlayer(SnId int32) {
	if _, exist := s.players[SnId]; exist {
		delete(s.players, SnId)
	}
}
func (s *FortuneZhiShenSceneData) OnPlayerLeave(p *base.Player, reason int) {
	//5分钟之后 帮玩家选择并保存log 踢出玩家
	if playerEx, ok := p.GetExtraData().(*FortuneZhiShenPlayerData); ok {
		if playerEx.firstFreeTimes > 0 || playerEx.secondFreeTimes > 0 {
			playerEx.nowGetCoin = playerEx.winCoin + playerEx.firstWinCoin + playerEx.secondWinCoin
			if playerEx.nowGetCoin != 0 {
				s.BilledCoin(playerEx)
				s.SaveLog(playerEx, 1)
			}
		}
	}
	s.delPlayer(p.SnId)
}
func (s *FortuneZhiShenSceneData) Win(p *FortuneZhiShenPlayerData) {
	p.Win()
	var winCoin int64
	var gemstoneBill bool
	var needAward bool
	if p.result.GemstoneNum >= 6 {
		needAward = true
	} else {
		gemstoneBill = true
	}
	if needAward {
		if p.gameState == gamerule.Normal {
			//宝石金额
			for k, v := range p.gemstoneRateCoin {
				if p.result.Cards[k] == gamerule.Gemstone {
					if v != 0 && v < gamerule.GrandPrize*1000000 {
						winCoin += v
					}
				}
			}
		} else {
			//免费和旋转
			for k, coin := range p.gemstoneRateCoin {
				if p.result.Cards[k] == gamerule.Gemstone {
					if coin != 0 && coin < gamerule.GrandPrize*1000000 {
						if p.result.MidIcon != -1 {
							var mid = []int{2, 3, 6, 7, 8, 11, 12, 13}
							if gamerule.FindInxInArray(k, mid) {
								continue
							}
						}
						winCoin += coin
					}
				}
			}
			if p.secondFreeTimes == 0 || p.result.GemstoneNum == 15 {
				gemstoneBill = true
				p.secondFreeTimes = 0
			}
		}
	}
	//派奖  有就给
	if p.grandPrizeNum+p.rewardGrandPrizeNum > 0 {
		grandPrize := s.jackpot.HugePools[p.betIdx] / gamerule.NowByte
		win := grandPrize * (p.grandPrizeNum + p.rewardGrandPrizeNum) * 10 / 100
		winCoin += win
		p.hitPrizePool[3] = win
		s.DelPrizeCoin(win, p.betIdx, gamerule.GrandPrize)
	}
	if p.bigPrizeNum > 0 {
		bigPrize := s.jackpot.BigPools[p.betIdx] / gamerule.NowByte
		win := bigPrize * p.bigPrizeNum * 10 / 100
		winCoin += win
		p.hitPrizePool[2] = win
		s.DelPrizeCoin(win, p.betIdx, gamerule.BigPrize)
	}
	if needAward {
		//派奖 10%
		if p.midPrizeNum > 0 {
			midPrize := s.jackpot.MiddlePools[p.betIdx] / gamerule.NowByte
			win := midPrize * p.midPrizeNum
			winCoin += win
			p.hitPrizePool[1] = win
		}
		if p.smallPrizeNum > 0 {
			smallPrize := s.jackpot.SmallPools[p.betIdx] / gamerule.NowByte
			win := smallPrize * p.smallPrizeNum
			winCoin += win
			p.hitPrizePool[0] = win
		}
	}
	p.gemstoneWinCoin = winCoin
	//派奖 10%
	if !gemstoneBill {
		winCoin = 0
		p.hitPrizePool = make([]int64, 4)
	}

	//赢的线数(包含发财元素倍率)
	winCoin += p.winTotalRate * p.oneBetCoin
	p.nowGetCoin = 0
	if p.gameState == gamerule.Normal {
		if winCoin > 0 {
			p.winCoin = winCoin
			p.nowGetCoin = winCoin
		}
	} else if p.gameState == gamerule.FreeGame {
		p.firstWinCoin += winCoin
		if p.firstFreeTimes == 0 && p.firstWinCoin > 0 {
			p.nowGetCoin = p.firstWinCoin
		}
	} else {
		if p.secondFreeTimes == 0 && winCoin > 0 {
			p.secondWinCoin = winCoin
			p.nowGetCoin = winCoin
		}
	}
	if p.nowGetCoin != 0 {
		s.BilledCoin(p)
	}
}
func (s *FortuneZhiShenSceneData) BilledCoin(p *FortuneZhiShenPlayerData) {
	//SysProfitCoinMgr.Add(s.sysProfitCoinKey, 0, p.nowGetCoin)
	p.Statics(s.GetKeyGameId(), s.KeyGamefreeId, p.nowGetCoin, false)
	tax := int64(math.Ceil(float64(p.nowGetCoin) * float64(s.GetDBGameFree().GetTaxRate()) / 10000))
	p.taxCoin = tax
	p.nowGetCoin -= tax
	p.AddServiceFee(tax)
	p.AddCoin(p.nowGetCoin, common.GainWay_HundredSceneWin, 0, "system", s.GetSceneName())

	p.noWinTimes = 0
	if !s.GetTesting() {
		base.GetCoinPoolMgr().PushCoin(s.GetGameFreeId(), s.GetGroupId(), s.GetPlatform(), -(p.nowGetCoin + tax))
	}
}

func (s *FortuneZhiShenSceneData) LoadJackPotData() {
	str := base.SlotsPoolMgr.GetPool(s.GetGameFreeId(), s.GetPlatform())
	if str != "" {
		jackpot := &base.SlotJackpotPool{}
		err := json.Unmarshal([]byte(str), jackpot)
		if err == nil {
			s.jackpot = jackpot
		}
	}
	if s.jackpot != nil {
		base.SlotsPoolMgr.SetPool(s.GetGameFreeId(), s.GetPlatform(), s.jackpot)
	} else {
		s.jackpot = &base.SlotJackpotPool{}
		jp := s.GetDBGameFree().GetJackpot()
		for k, v := range jp {
			switch k % 4 {
			case 0:
				s.jackpot.SmallPools = append(s.jackpot.SmallPools, int64(v)*gamerule.NowByte)
			case 1:
				s.jackpot.MiddlePools = append(s.jackpot.MiddlePools, int64(v)*gamerule.NowByte)
			case 2:
				s.jackpot.BigPools = append(s.jackpot.BigPools, int64(v)*gamerule.NowByte)
			case 3:
				s.jackpot.HugePools = append(s.jackpot.HugePools, int64(v)*gamerule.NowByte)
			}
		}
	}
}
func (s *FortuneZhiShenSceneData) Billed(p *FortuneZhiShenPlayerData) {
	var wl []*fortunezhishen.FortuneZhiShenWinLine
	for _, r := range p.result.WinLine {
		wl = append(wl, &fortunezhishen.FortuneZhiShenWinLine{
			Poss:     r.Poss,
			WinScore: proto.Int64(int64(r.LineId)),
		})
	}
	var hitPrize = make([]int64, 5)
	var oneGrandPrize int64
	for k, v := range p.hitPrizePool {
		if k == 3 && (p.grandPrizeNum+p.rewardGrandPrizeNum) > 0 {
			oneGrandPrize = v / (p.grandPrizeNum + p.rewardGrandPrizeNum)
			hitPrize[3] = oneGrandPrize * p.grandPrizeNum
		} else {
			hitPrize[k] = v
		}
	}
	hitPrize[4] = oneGrandPrize * p.rewardGrandPrizeNum
	var show1, show2, show3 = p.winCoin, p.firstWinCoin, p.secondWinCoin
	if s.GetDBGameFree().GetTaxRate() > 0 {
		show1 = int64(math.Ceil(float64(p.winCoin) * float64(s.GetDBGameFree().GetTaxRate()) / 10000))
		show1 = p.winCoin - show1
		show2 = int64(math.Ceil(float64(p.firstWinCoin) * float64(s.GetDBGameFree().GetTaxRate()) / 10000))
		show2 = p.firstWinCoin - show2
		show3 = int64(math.Ceil(float64(p.secondWinCoin) * float64(s.GetDBGameFree().GetTaxRate()) / 10000))
		show3 = p.secondWinCoin - show3
	}
	pack := &fortunezhishen.SCFortuneZhiShenBilled{
		Coin:             proto.Int64(p.GetCoin()),
		UiShow:           p.result.Cards,
		FirstFreeTimes:   proto.Int(p.firstFreeTimes),
		SecondFreeTimes:  proto.Int(p.secondFreeTimes),
		WinLines:         wl,
		GemstoneRateCoin: p.gemstoneRateCoin,
		NowGameState:     proto.Int(p.gameState),
		WinRate:          proto.Int64(p.winTotalRate),
		WinCoin:          proto.Int64(show1),
		FirstWinCoin:     proto.Int64(show2),
		SecondWinCOin:    proto.Int64(show3),
		HitPrize:         hitPrize,
	}
	proto.SetDefaults(pack)
	logger.Logger.Trace("SCFortuneZhiShenBilled: ", pack)
	p.SendToClient(int(fortunezhishen.FortuneZSPacketID_PACKET_SC_FORTUNEZHISHEN_BILLED), pack)
}
func (s *FortuneZhiShenSceneData) SaveLog(p *FortuneZhiShenPlayerData, isOffline int) {
	if s.GetTesting() {
		return
	}
	s.SendPlayerBet(p)
	var betCoin int64
	var nowNRound int
	var nowGetCoin int64
	var isF bool
	if p.gameState == gamerule.Normal {
		betCoin = p.betCoin
		nowGetCoin = p.nowGetCoin
		isF = true
	} else if p.gameState == gamerule.FreeGame {
		nowNRound = p.nowFirstTimes
		nowGetCoin = p.firstWinCoin
		if p.firstFreeTimes != 0 {
			nowGetCoin += p.gemstoneWinCoin
		} else {
			isF = true
		}
	} else {
		nowNRound = p.nowSecondTimes
		nowGetCoin = p.secondWinCoin
		if p.secondFreeTimes != 0 {
			nowGetCoin += p.gemstoneWinCoin
		} else {
			isF = true
		}
	}
	fortuneZhiShenType := model.FortuneZhiShenType{
		RoomId:          s.GetSceneId(),
		BasicScore:      int32(p.oneBetCoin),
		PlayerSnId:      p.SnId,
		BeforeCoin:      p.GetCoin() - (p.nowGetCoin - betCoin),
		AfterCoin:       p.GetCoin(),
		ChangeCoin:      p.nowGetCoin - betCoin,
		TotalBetCoin:    betCoin,
		TotalLine:       50,
		TotalWinCoin:    nowGetCoin,
		NowGameState:    p.gameState,
		NowNRound:       nowNRound,
		IsOffline:       isOffline,
		FirstFreeTimes:  p.firstFreeTimes,
		SecondFreeTimes: p.secondFreeTimes,
		////////////////////
		HitPrizePool:    p.hitPrizePool,
		WinLineNum:      len(p.result.WinLine),
		WinLineRate:     p.winTotalRate,
		WinLineCoin:     p.winTotalRate * p.oneBetCoin,
		GemstoneNum:     p.result.GemstoneNum,
		GemstoneWinCoin: p.gemstoneWinCoin,
		/////
		Cards:            p.result.Cards,
		GemstoneRateCoin: p.gemstoneRateCoin,
		WhiteLevel:       p.WhiteLevel,
		BlackLevel:       p.BlackLevel,
		TaxCoin:          p.taxCoin,
	}
	var winLine []model.FortuneZhiShenWinLine
	for _, line := range p.result.WinLine {
		fw := model.FortuneZhiShenWinLine{
			Id:          line.LineId,
			EleValue:    line.Lines[0],
			Num:         len(line.Lines),
			Rate:        line.Rate,
			WinCoin:     line.Rate * p.oneBetCoin,
			WinFreeGame: -1,
		}
		winLine = append(winLine, fw)
	}
	winLine = append(winLine, p.wl...)
	fortuneZhiShenType.WinLine = winLine
	info, err := model.MarshalGameNoteByROLL(&fortuneZhiShenType)
	if err == nil {
		logId, _ := model.AutoIncGameLogId()
		s.SaveGameDetailedLog(logId, info, &base.GameDetailedParam{})

		//水池上下文环境s
		s.SetCpCtx(p.cpCtx)
		var totalIn, totalOut int64
		if betCoin > 0 {
			totalIn = betCoin
		}
		if nowGetCoin > 0 && isF {
			//totalOut = p.GetCoin() - p.GetStartCoin() + betCoin + p.taxCoin
			totalOut = p.nowGetCoin + p.taxCoin
		}
		validFlow := totalIn + totalOut
		validBet := common.AbsI64(totalIn - totalOut)
		param := base.GetSaveGamePlayerListLogParam(p.Platform, p.Channel, p.BeUnderAgentCode, p.PackageID, logId,
			p.InviterId, totalIn, totalOut, p.taxCoin, 0, totalIn, p.GetCoin()-(p.nowGetCoin+p.taxCoin), validFlow, validBet,
			s.IsPlayerFirst(p.Player), false)
		s.SaveGamePlayerListLog(p.SnId, param)
	}
	s.SetGameNowTime(time.Now())
	if s.CheckNeedDestroy() && p.firstFreeTimes == 0 && p.secondFreeTimes == 0 {
		s.PlayerLeave(p.Player, common.PlayerLeaveReason_OnDestroy, true)
	}
}
func (s *FortuneZhiShenSceneData) SendPlayerBet(p *FortuneZhiShenPlayerData) {
	//统计输下注金币数
	if !p.IsRob && !s.GetTesting() {
		betCoin := p.betCoin
		if p.gameState != gamerule.Normal {
			betCoin = 0
		}
		playerBet := &server_proto.PlayerBet{
			SnId:       proto.Int32(p.SnId),
			Bet:        proto.Int64(betCoin),
			Gain:       proto.Int64(p.nowGetCoin + p.taxCoin),
			Tax:        proto.Int64(p.taxCoin),
			Coin:       proto.Int64(p.GetCoin()),
			GameCoinTs: proto.Int64(p.GameCoinTs),
		}
		gwPlayerBet := &server_proto.GWPlayerBet{
			SceneId:    proto.Int(s.SceneId),
			GameFreeId: proto.Int32(s.GetDBGameFree().GetId()),
			RobotGain:  proto.Int64(-(p.GetCoin() - (p.nowGetCoin + p.taxCoin) + p.taxCoin)),
		}
		gwPlayerBet.PlayerBets = append(gwPlayerBet.PlayerBets, playerBet)
		proto.SetDefaults(gwPlayerBet)
		s.SendToWorld(int(server_proto.SSPacketID_PACKET_GW_PLAYERBET), gwPlayerBet)
		logger.Logger.Trace("Send msg gwPlayerBet ===>", gwPlayerBet)
	}
}
