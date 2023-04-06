package dezhoupoker

import (
	"fmt"
	"games.yol.com/win88/common"
	rule "games.yol.com/win88/gamerule/dezhoupoker"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/dezhoupoker"
	"games.yol.com/win88/protocol/server"
	"games.yol.com/win88/robot/base"
	"games.yol.com/win88/srvdata"
	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/core/task"
	"math/rand"
	"time"
)

//通用结构
const (
	Act_Call int32 = iota //
	Act_Raise
	Act_AllIn
	Act_Fold
	Act_Check
	Act_Max
)

var DezhouPreflopRaiseRate = []struct {
	OrderRange []int32
	ChipRange  []int32
	Rate       []int32
}{
	{[]int32{1, 3}, []int32{3, 5}, []int32{100, 0, 0}},
	{[]int32{4, 11}, []int32{3, 5}, []int32{85, 15, 0}},
	{[]int32{12, 20}, []int32{3, 5}, []int32{70, 30, 0}},
	{[]int32{21, 35}, []int32{3, 5}, []int32{60, 40, 0}},
	{[]int32{36, 60}, []int32{3, 5}, []int32{40, 60, 0}},
	{[]int32{61, 100}, []int32{2, 4}, []int32{20, 80, 0}},
	{[]int32{101, 999}, []int32{2, 4}, []int32{0, 0, 100}},
}

var DezhouPreflopReraiseRate = []struct {
	OrderRange []int32
	ChipRange  []int32
	Rate       []int32
}{
	{[]int32{1, 3}, []int32{50, 200}, []int32{80, 20, 0}},
	{[]int32{4, 11}, []int32{50, 100}, []int32{40, 60, 0}},
	{[]int32{12, 20}, []int32{50, 66}, []int32{20, 80, 0}},
	{[]int32{21, 35}, []int32{50, 66}, []int32{5, 95, 0}},
	{[]int32{36, 60}, []int32{50, 66}, []int32{0, 60, 40}},
	{[]int32{61, 100}, []int32{50, 66}, []int32{0, 40, 60}},
	{[]int32{101, 999}, []int32{50, 66}, []int32{0, 0, 100}},
}

var DezhouPreflopPositionRate = []struct {
	ChipRange   []int32
	RateRaise   []int32
	RateReraise []int32
}{
	{[]int32{2, 4}, []int32{0, 10, 90}, []int32{0, 0, 100}}, //枪口
	{[]int32{2, 4}, []int32{0, 12, 88}, []int32{0, 0, 100}}, //枪口1
	{[]int32{2, 4}, []int32{0, 15, 85}, []int32{0, 0, 100}}, //枪口2
	{[]int32{2, 4}, []int32{0, 18, 82}, []int32{0, 0, 100}}, //关位
	{[]int32{2, 4}, []int32{0, 22, 78}, []int32{0, 0, 100}}, //关位1
	{[]int32{2, 4}, []int32{0, 26, 74}, []int32{0, 0, 100}}, //关位2
	{[]int32{2, 4}, []int32{0, 30, 70}, []int32{0, 0, 100}}, //庄家位
	{[]int32{2, 4}, []int32{0, 35, 65}, []int32{0, 0, 100}}, //小盲注
	{[]int32{2, 4}, []int32{60, 40, 0}, []int32{0, 0, 100}}, //大盲注
}

func DezhouCardKindRateFormat(cnt []int32, total int) string {
	var str string
	str = "{ "
	for i := 0; i < len(cnt) && i < int(rule.KindOfCard_Invalide); i++ {
		if cnt[i] > 0 {
			str += fmt.Sprintf("%s:%0.2f%%|%d/%d ", rule.KindOfCardStr[i], float32(cnt[i])/float32(total), cnt[i], total)
		}
	}
	str += "}"
	return str
}

var OpSet_Fold_Raise_Call_AllIn = []int32{Act_Call, Act_Raise, Act_AllIn, Act_Fold}
var OpSet_Fold_Raise_Call_AllIn_Check = []int32{Act_Call, Act_Raise, Act_AllIn, Act_Fold, Act_Check}

func CalOpKindAndOpValue(aiModel, curRoundNum, score int32, minRaise, minBet, bigBlind int64) (int32, int64) {
	datas := srvdata.PBDB_DZPKPlayerOPMgr.Datas.GetArr()
	for _, v := range datas {

		if v.GetModeType() == aiModel && score >= v.GetScore()[0] && score <= v.GetScore()[1] && curRoundNum >= v.GetRounds()[0] && curRoundNum <= v.GetRounds()[1] {
			opSetParams := v.GetParams()
			logger.Logger.Trace("v:", v.GetId(), v.GetParams())

			rstOpType := rule.DezhouPokerPlayerOpNull
			actOp, actOpValue := ChoiseOpSet(minRaise, minBet, opSetParams, bigBlind)
			switch actOp {
			case Act_Call:
				rstOpType = rule.DezhouPokerPlayerOpCall
			case Act_Raise:
				rstOpType = rule.DezhouPokerPlayerOpRaise
			case Act_AllIn:
				rstOpType = rule.DezhouPokerPlayerOpAllIn
			case Act_Fold:
				rstOpType = rule.DezhouPokerPlayerOpFold
			case Act_Check:
				rstOpType = rule.DezhouPokerPlayerOpCheck
			default:
				rstOpType = rule.DezhouPokerPlayerOpFold
			}
			logger.Logger.Trace("rstOpType:", rstOpType)
			return rstOpType, actOpValue
		}
	}
	logger.Logger.Trace("rule.DezhouPokerPlayerOpFold, 0", score)
	return rule.DezhouPokerPlayerOpFold, 0
}

func ChoiseOpSet(minRaise, minBet int64, operRateSet []int32, bigBlind int64) (int32, int64) {
	OpSelected := int32(0)

	var canOpSet []int32
	if minBet == 0 {
		canOpSet = common.CopySliceInt32(OpSet_Fold_Raise_Call_AllIn_Check)
	} else {
		canOpSet = common.CopySliceInt32(OpSet_Fold_Raise_Call_AllIn)
	}

	var rstOperRateSet []int64
	for i := 0; i < len(operRateSet); i++ {
		if common.InSliceInt32(canOpSet, int32(i)) {
			rstOperRateSet = append(rstOperRateSet, int64(operRateSet[i]))
		} else {
			rstOperRateSet = append(rstOperRateSet, 0)
		}
	}

	OpSelected = int32(common.RandSliceIndexByWight(rstOperRateSet))

	logger.Logger.Trace("rstOperRateSet ", rstOperRateSet)
	logger.Logger.Trace("OpSelected ", OpSelected)

	if OpSelected == Act_Call && minBet == 0 {
		OpSelected = Act_Raise

		logger.Logger.Trace("Act_Call ==> Act_Raise")
	}

	if OpSelected == Act_Raise {
		//60%概率最小加注值；40%概率2~5倍最小加注值
		var playerMinRaise int64
		if rand.Int31n(100) < 60 {
			playerMinRaise = minBet + minRaise

			logger.Logger.Trace("1~1 playerMinRaise", playerMinRaise)
		} else {
			randOdd := int64(common.RandFromRange(2, 5))
			playerMinRaise = (minBet + minRaise) * randOdd

			logger.Logger.Trace("2~5 playerMinRaise", playerMinRaise)
		}
		if playerMinRaise < bigBlind {
			playerMinRaise = bigBlind
		}
		return Act_Raise, playerMinRaise
	}
	return OpSelected, minBet

}

var DezhouHandCardWinRate = make(map[string]*server.DB_DZPKHandCardWin)

func HandCardIsEnoughNuts(cards []int32, num int32) bool {
	str := rule.HandCardShowStr(cards)
	str = fmt.Sprintf("%s_%d", str, num)
	if _, exist := DezhouHandCardWinRate[str]; exist {
		return true
	}
	return false
}

func HandCardWinRate(cards []int32, num int32) (int32, int32) {
	str := rule.HandCardShowStr(cards)
	str = fmt.Sprintf("%s_%d", str, num)
	if data, exist := DezhouHandCardWinRate[str]; exist {
		return data.GetWinRate(), data.GetOrder()
	}
	return -1, 101
}

func DezhouCalOpAndValueByEV(s *netlib.Session, sceneEx *DezhouPokerScene, playerEx *DezhouPokerPlayer, minRaise, minBet, myBet, totalBet int64, raiseOption []int64, remainChip, myCurRoundTotalBet int64, otherCurrRoundPerBet []int64, totalPlayerNum, restPlayerNum int32, rolePos int32) {
	if sceneEx != nil && playerEx != nil {

		filterRaiseOption := func(options []int64, rate int32, sBB, eBB int64) []int64 {
			if rand.Int31n(100) < rate {
				startIdx := 0
				for i := 0; i < len(options); i++ {
					if options[i] >= sBB {
						startIdx = i
						break
					}
				}
				lastIdx := 0
				for j := len(options) - 1; j > startIdx; j-- {
					if options[j] <= eBB {
						lastIdx = j
						break
					}
				}
				if lastIdx < startIdx {
					lastIdx = startIdx
				}
				options = options[startIdx:lastIdx]
				return options
			}
			return nil
		}

		//随机选择一个加注金额
		chooseRaiseValue := func(rate int, startBB, endBB int, bb int64) int64 {
			if rand.Intn(100) < rate {
				n := endBB - startBB
				if n == 0 {
					n = 1
				}
				return int64(rand.Intn(n)+startBB) * bb
			}
			return 0
		}

		start := time.Now()
		gctx := sceneEx.CreateGameCtx()

		var opCode int32
		var opValue int64
		bigBlind := sceneEx.SCDezhouPokerRoomInfo.GetSmallBlind() * 2
		commonCardCnt := 0
		state := int(sceneEx.GetState())
		if state >= rule.DezhouPokerSceneStateRiver {
			commonCardCnt = 5
		} else if state >= rule.DezhouPokerSceneStateTurn {
			commonCardCnt = 4
		} else if state >= rule.DezhouPokerSceneStateFlop {
			commonCardCnt = 3
		}
		cards := make([]int32, 0, 2+commonCardCnt)
		cards = append(cards, playerEx.GetCards()...)
		cards = append(cards, sceneEx.Cards[0:commonCardCnt]...)
		currCI := rule.KindOfCardFigureUpExSington.FigureUpByCard(cards)
		commCI := rule.KindOfCardFigureUpExSington.FigureUpByCard(sceneEx.Cards[0:commonCardCnt])

		//preflop
		if int(sceneEx.GetState()) < rule.DezhouPokerSceneStateFlop {
			//test use
			if common.CustomConfig.GetBool("UseDZPWPDebug") {
				pack := &dezhoupoker.CSCDezhouPokerWPUpdate{}
				for _, pc := range gctx.PlayerCards {
					winProbalility, _ := HandCardWinRate(pc.HandCard, restPlayerNum)
					if winProbalility < 0 {
						winProbalility = 0
					}
					pack.Datas = append(pack.Datas, &dezhoupoker.DezhouPokerCardsWP{
						SnId:               proto.Int32(pc.UserData.(int32)),
						Cards:              pc.HandCard,
						WinningProbability: proto.Int32(winProbalility),
					})
				}
				s.Send(int(dezhoupoker.DZPKPacketID_PACKET_CSC_DEZHOUPOKER_WPUPDATE), pack)
			}
			winRate, handSeq := HandCardWinRate(playerEx.Cards, restPlayerNum)
			logger.Logger.Infof("===[preflop] SnId:%v 手牌牌型:%s 参考胜率=%.2f 剩余人数=%v 手牌排序=%v", playerEx.GetSnId(), currCI.KindStr(), float32(winRate)/100, restPlayerNum, handSeq)
			opCode = rule.DezhouPokerPlayerOpCall
			opValue = minBet
			raise := false
			reraise := false
			var options []int64
			if myBet <= bigBlind*2 && minBet <= bigBlind { //raise
				if winRate > 0 {
					for i := 0; i < len(DezhouPreflopRaiseRate); i++ {
						if handSeq >= DezhouPreflopRaiseRate[i].OrderRange[0] && handSeq <= DezhouPreflopRaiseRate[i].OrderRange[1] {
							idx := common.RandSliceIndexByWight31N(DezhouPreflopRaiseRate[i].Rate)
							switch idx {
							case 0: //加注
								opValue = chooseRaiseValue(100, int(DezhouPreflopRaiseRate[i].ChipRange[0]), int(DezhouPreflopRaiseRate[i].ChipRange[1]), bigBlind)
								if opValue > minBet {
									raise = true
									if (opValue+myCurRoundTotalBet)%bigBlind != 0 { //对齐到大盲注
										opValue = ((opValue+myCurRoundTotalBet)/bigBlind + 1) * bigBlind
									}
									opCode = rule.DezhouPokerPlayerOpRaise
									raise = true
								}
							case 1: //跟注
								opCode = rule.DezhouPokerPlayerOpCall
								opValue = minBet
							case 2: //弃牌
								opCode = rule.DezhouPokerPlayerOpFold
								opValue = 0
							}
							logger.Logger.Infof("===[preflop] SnId:%v 手牌牌型:%s winRate:%v 手牌强度排序编号:%v ", playerEx.GetSnId(), currCI.KindStr(), winRate, handSeq)
							break
						}
					}
				} else {
					if rolePos >= 0 && rolePos < int32(len(DezhouPreflopPositionRate)) {
						idx := common.RandSliceIndexByWight31N(DezhouPreflopPositionRate[rolePos].RateRaise)
						switch idx {
						case 0: //跟注
							opCode = rule.DezhouPokerPlayerOpCall
							opValue = minBet
						case 1: //加注
							opValue = chooseRaiseValue(100, int(DezhouPreflopPositionRate[rolePos].ChipRange[0]), int(DezhouPreflopPositionRate[rolePos].ChipRange[1]), bigBlind)
							if opValue > minBet {
								raise = true
								if (opValue+myCurRoundTotalBet)%bigBlind != 0 { //对齐到大盲注
									opValue = ((opValue+myCurRoundTotalBet)/bigBlind + 1) * bigBlind
								}
								opCode = rule.DezhouPokerPlayerOpRaise
								raise = true
							}
						case 2: //弃牌
							opCode = rule.DezhouPokerPlayerOpFold
							opValue = 0
						}

						logger.Logger.Infof("===[preflop] SnId:%v 手牌牌型:%s 位置:%v opcode:%v opvalue:%v", playerEx.GetSnId(), currCI.KindStr(), rule.PosDesc[rolePos], opCode, opValue)
					}
				}
			} else { //reraise
				for i := 0; i < len(DezhouPreflopReraiseRate); i++ {
					if handSeq >= DezhouPreflopReraiseRate[i].OrderRange[0] && handSeq <= DezhouPreflopReraiseRate[i].OrderRange[1] {
						idx := common.RandSliceIndexByWight31N(DezhouPreflopReraiseRate[i].Rate)
						switch idx {
						case 0: //再加注
							//选一个加注金额

							for _, chip := range raiseOption {
								ev2 := int64(winRate)*(totalBet-myBet) - int64(10000-winRate)*chip
								if ev2 > 0 {
									options = append(options, chip)
								}
							}
							options = filterRaiseOption(options, 100, totalBet*int64(DezhouPreflopReraiseRate[i].ChipRange[0])/100, totalBet*int64(DezhouPreflopReraiseRate[i].ChipRange[1])/100)
							if len(options) > 0 {
								opValue = options[rand.Intn(len(options))]
							}
							if opValue > minBet {
								reraise = true
								opCode = rule.DezhouPokerPlayerOpRaise
							} else {
								opCode = rule.DezhouPokerPlayerOpFold
								opValue = 0
								if handSeq <= 20 && minBet <= 20*bigBlind {
									opCode = rule.DezhouPokerPlayerOpCall
									opValue = minBet
								} else {
									ev1 := int64(winRate)*(totalBet-myBet) - int64(10000-winRate)*minBet
									if ev1 > 0 {
										opCode = rule.DezhouPokerPlayerOpCall
										opValue = minBet
									}
								}
							}
						case 1: //跟注
							opCode = rule.DezhouPokerPlayerOpFold
							opValue = 0
							if handSeq <= 20 && minBet <= 20*bigBlind {
								opCode = rule.DezhouPokerPlayerOpCall
								opValue = minBet
							} else {
								ev1 := int64(winRate)*(totalBet-myBet) - int64(10000-winRate)*minBet
								if ev1 > 0 {
									opCode = rule.DezhouPokerPlayerOpCall
									opValue = minBet
								}
							}
						case 2: //弃牌
							opCode = rule.DezhouPokerPlayerOpFold
							opValue = 0
						}
						logger.Logger.Infof("===[preflop] SnId:%v 手牌牌型:%s winRate:%v 手牌强度排序编号:%v opcode:%v opvalue:%v", playerEx.GetSnId(), currCI.KindStr(), winRate, handSeq, opCode, opValue)
						break
					}
				}
			}
			logger.Logger.Infof("===[preflop] snid:%v [raise]:%v [reraise]=%v value=%v options:%v", playerEx.GetSnId(), raise, reraise, opValue, options)

			cost := int(time.Now().Sub(start).Seconds())
			roundNum := DezhouCalOpStateRoundNum(int(sceneEx.GetState()))
			delaySecondMin, delaySecondMax := DezhouCalOpDelaySeconds(roundNum)
			if cost <= delaySecondMin {
				delaySecondMin -= cost
				delaySecondMax -= cost
			} else if cost <= delaySecondMax {
				delaySecondMin = 0
				delaySecondMax -= cost
			}
			pack := &dezhoupoker.CSDezhouPokerPlayerOp{
				OpCode: proto.Int32(opCode),
			}
			pack.OpParam = append(pack.OpParam, opValue)
			proto.SetDefaults(pack)
			base.DelaySend(s, int(dezhoupoker.DZPKPacketID_PACKET_CS_DEZHOUPOKER_OP), pack, delaySecondMin, delaySecondMax)
			return
		}

		task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
			rule.CalWinningProbability(gctx)
			return gctx
		}), task.CompleteNotifyWrapper(func(data interface{}, tt task.Task) {
			//test use
			if common.CustomConfig.GetBool("UseDZPWPDebug") {
				pack := &dezhoupoker.CSCDezhouPokerWPUpdate{}
				for _, pc := range gctx.PlayerCards {
					pack.Datas = append(pack.Datas, &dezhoupoker.DezhouPokerCardsWP{
						SnId:               proto.Int32(pc.UserData.(int32)),
						Cards:              pc.HandCard,
						WinningProbability: proto.Int32(pc.WinningProbability),
					})
				}
				s.Send(int(dezhoupoker.DZPKPacketID_PACKET_CSC_DEZHOUPOKER_WPUPDATE), pack)
			}

			cost := int(time.Now().Sub(start).Seconds())
			var pc *rule.PlayerCard
			for _, p := range gctx.PlayerCards {
				if p.UserData.(int32) == playerEx.GetSnId() {
					pc = p
					break
				}
			}

			if minBet == 0 {
				opCode = rule.DezhouPokerPlayerOpCheck
			} else {
				opCode = rule.DezhouPokerPlayerOpFold
			}
			opValue = minBet

			if pc != nil {
				var options []int64
				//选一个加注金额
				for _, chip := range raiseOption {
					ev2 := int64(pc.WinningProbability)*(totalBet-myBet) - int64(10000-pc.WinningProbability)*chip
					if ev2 > 0 {
						options = append(options, chip)
					}
				}

				if remainChip < minBet { //如果自己allin，那么底池的真实收益要减去溢出值
					minBet = remainChip
					overVal := int64(0)
					myCurRoundTotalBetWill := myCurRoundTotalBet + minBet
					for _, v := range otherCurrRoundPerBet {
						if v > myCurRoundTotalBetWill {
							overVal += v - myCurRoundTotalBetWill
						}
					}
					totalBet -= overVal
				}

				switch gctx.CommonCardCnt {
				case 3, 4: //flop和turn 翻牌|转牌圈
					ev1 := int64(pc.WinningProbability)*(totalBet-myBet) - int64(10000-pc.WinningProbability)*minBet
					logger.Logger.Infof("===[flop|turn]SnId:%v 手牌牌型:%s WinningProbability=%.2f EV1=%v \n可能获胜牌型=%v \n", playerEx.GetSnId(), currCI.KindStr(), float32(pc.WinningProbability)/100, ev1, DezhouCardKindRateFormat(pc.WinCardKind[:], gctx.Possibilities))
					if ev1 > 0 {
						rate := int32(0)
						//增加被唬走的概率
						beBluff := false

						//若AI已下注金额≥剩余携带金额，不会被唬走，走EV值判断,否则可能被唬走
						if pc.WinningProbability >= 5000 && remainChip > myBet && minBet > 0 {
							switch currCI.Kind {
							case rule.KindOfCard_HighCard: //高牌
								if minBet <= 20*bigBlind { //<=20BB
									rate = int32(float32(10000-pc.WinningProbability) * 1.2)
								} else if minBet > 20*bigBlind && minBet < 50*bigBlind {
									rate = int32(float32(10000-pc.WinningProbability) * 1.5)
								} else {
									rate = int32(float32(10000-pc.WinningProbability) * 2.0)
								}

							case rule.KindOfCard_OnePair: //一对
								if rule.IsOverPair(currCI, playerEx.GetCards(), sceneEx.Cards[:commonCardCnt]) { //超对
									if minBet <= 20*bigBlind { //<=20BB
										rate = int32(float32(10000-pc.WinningProbability) / 2.5)
									} else if minBet > 20*bigBlind && minBet < 50*bigBlind {
										rate = int32(float32(10000-pc.WinningProbability) / 2)
									} else {
										rate = int32(float32(10000-pc.WinningProbability) / 1.5)
									}
								} else if rule.IsTopPair(currCI, playerEx.GetCards(), sceneEx.Cards[:commonCardCnt]) { //顶对
									if minBet <= 20*bigBlind { //<=20BB
										rate = int32(float32(10000-pc.WinningProbability) / 2)
									} else if minBet > 20*bigBlind && minBet < 50*bigBlind {
										rate = int32(float32(10000-pc.WinningProbability) / 1.5)
									} else {
										rate = int32(float32(10000-pc.WinningProbability) / 1.0)
									}
								} else if rule.IsMiddlePair(currCI, playerEx.GetCards(), sceneEx.Cards[:commonCardCnt]) { //中对
									if minBet <= 20*bigBlind { //<=20BB
										rate = int32(float32(10000-pc.WinningProbability) / 1.5)
									} else if minBet > 20*bigBlind && minBet < 50*bigBlind {
										rate = int32(float32(10000-pc.WinningProbability) * 1.0)
									} else {
										rate = int32(float32(10000-pc.WinningProbability) * 1.2)
									}
								} else { //底对|低对
									if minBet <= 20*bigBlind { //<=20BB
										rate = int32(float32(10000-pc.WinningProbability) * 1.0)
									} else if minBet > 20*bigBlind && minBet < 50*bigBlind {
										rate = int32(float32(10000-pc.WinningProbability) * 1.2)
									} else {
										rate = int32(float32(10000-pc.WinningProbability) * 1.5)
									}
								}
							case rule.KindOfCard_TwoPair: //两对
								if minBet <= 20*bigBlind { //<=20BB
									rate = int32(float32(10000-pc.WinningProbability) / 2.0)
								} else if minBet > 20*bigBlind && minBet < 50*bigBlind {
									rate = int32(float32(10000-pc.WinningProbability) / 1.5)
								} else {
									rate = int32(float32(10000-pc.WinningProbability) * 1.0)
								}
							}
							//若需要跟注金额≥已下注金额的/2
							if rand.Int31n(10000) < rate && minBet >= myBet/2 {
								opCode = rule.DezhouPokerPlayerOpFold
								opValue = 0
								beBluff = true
							}
							logger.Logger.Infof("===[flop|turn]检测是否被唬 snid:%v beBluff rate:%v beBluff=%v 手牌牌型=%v 公牌牌型=%v", playerEx.GetSnId(), rate, beBluff, currCI.KindStr(), commCI.KindStr())
						}

						//没被唬走,正常走逻辑
						if !beBluff {
							opCode = rule.DezhouPokerPlayerOpCall
							opValue = minBet

							//牌力值>30%,尝试下加注
							if (pc.WinningProbability > 3000 && len(raiseOption) > 0) || rand.Int31n(10000) < pc.WinningProbability || rule.KindOfCardIsBetter(playerEx.GetCards(), sceneEx.Cards[:commonCardCnt]) {
								options = nil
								//选一个加注金额
								for _, chip := range raiseOption {
									ev2 := int64(pc.WinningProbability)*(totalBet-myBet) - int64(10000-pc.WinningProbability)*chip
									if ev2 > 0 {
										options = append(options, chip)
									}
								}
								raise := false
								reraise := false
								//根据牌型选择加注值
								if currCI.Kind > commCI.Kind {
									switch currCI.Kind {
									case rule.KindOfCard_HighCard: //高牌
										//不可达路径
										if minBet == 0 { //raise
											raise = true
											options = filterRaiseOption(raiseOption, 20, 0, 5*bigBlind)
										} else { //reraise
											reraise = true
											options = filterRaiseOption(raiseOption, 1, 0, 5*bigBlind)
										}
									case rule.KindOfCard_OnePair: //一对
										if minBet == 0 { //raise
											raise = true
											if rule.IsOverPair(currCI, playerEx.GetCards(), sceneEx.Cards[:commonCardCnt]) || rule.IsTopPair(currCI, playerEx.GetCards(), sceneEx.Cards[:commonCardCnt]) {
												rate = 65
											} else if rule.IsMiddlePair(currCI, playerEx.GetCards(), sceneEx.Cards[:commonCardCnt]) {
												rate = 35
											} else {
												rate = 15
											}
											options = filterRaiseOption(raiseOption, rate, 2*bigBlind, 10*bigBlind)
										} else { //reraise
											reraise = true
											if rule.IsOverPair(currCI, playerEx.GetCards(), sceneEx.Cards[:commonCardCnt]) || rule.IsTopPair(currCI, playerEx.GetCards(), sceneEx.Cards[:commonCardCnt]) {
												rate = 20
											} else if rule.IsMiddlePair(currCI, playerEx.GetCards(), sceneEx.Cards[:commonCardCnt]) {
												rate = 10
											} else {
												rate = 5
											}
											options = filterRaiseOption(raiseOption, rate, 2*bigBlind, 10*bigBlind)
										}
									case rule.KindOfCard_TwoPair: //两对
										if minBet == 0 { //raise
											raise = true
											options = filterRaiseOption(raiseOption, 70, 5*bigBlind, 20*bigBlind)
										} else { //reraise
											reraise = true
											options = filterRaiseOption(raiseOption, 20, 5*bigBlind, 20*bigBlind)
										}
									case rule.KindOfCard_ThreeKind: //三条
										if minBet == 0 { //raise
											raise = true
											options = filterRaiseOption(raiseOption, 70, 7*bigBlind, 30*bigBlind)
										} else { //reraise
											reraise = true
											options = filterRaiseOption(raiseOption, 30, 7*bigBlind, 30*bigBlind)
										}
									case rule.KindOfCard_Straight: //顺子
										if minBet == 0 { //raise
											raise = true
											options = filterRaiseOption(raiseOption, pc.WinningProbability/100, 5*bigBlind, 30*bigBlind)
										} else { //reraise
											reraise = true
											options = filterRaiseOption(raiseOption, pc.WinningProbability/100, 10*bigBlind, 50*bigBlind)
										}
									case rule.KindOfCard_Flush: //同花
										if minBet == 0 { //raise
											raise = true
											options = filterRaiseOption(raiseOption, pc.WinningProbability/100, 5*bigBlind, 30*bigBlind)
										} else { //reraise
											reraise = true
											options = filterRaiseOption(raiseOption, pc.WinningProbability/100, 10*bigBlind, 50*bigBlind)
										}
									case rule.KindOfCard_Fullhouse: //葫芦
										if minBet == 0 { //raise
											raise = true
											options = filterRaiseOption(raiseOption, pc.WinningProbability/100, 5*bigBlind, 30*bigBlind)
										} else { //reraise
											reraise = true
											options = filterRaiseOption(raiseOption, pc.WinningProbability/100, 10*bigBlind, 50*bigBlind)
										}
									case rule.KindOfCard_FourKind: //金刚
										if minBet == 0 { //raise
											raise = true
											options = filterRaiseOption(raiseOption, pc.WinningProbability/100, 5*bigBlind, 30*bigBlind)
										} else { //reraise
											reraise = true
											options = filterRaiseOption(raiseOption, pc.WinningProbability/100, 10*bigBlind, 50*bigBlind)
										}
									case rule.KindOfCard_StraightFlush, rule.KindOfCard_RoyalFlush: //同花顺
										if minBet == 0 { //raise
											raise = true
											options = filterRaiseOption(raiseOption, pc.WinningProbability/100, 5*bigBlind, 30*bigBlind)
										} else { //reraise
											reraise = true
											options = filterRaiseOption(raiseOption, pc.WinningProbability/100, 10*bigBlind, 50*bigBlind)
										}
									}
								} else { //高牌的情况
									if rule.KindOfCardFigureUpExSington.IsTingKinds(cards, []int32{rule.KindOfCard_Straight, rule.KindOfCard_Flush, rule.KindOfCard_StraightFlush, rule.KindOfCard_RoyalFlush}) {
										if minBet == 0 { //raise
											raise = true
											options = filterRaiseOption(raiseOption, pc.WinningProbability/200, 0, 30*bigBlind)
										} else { //reraise
											reraise = true
											options = filterRaiseOption(raiseOption, pc.WinningProbability/300, 0, 30*bigBlind)
										}
										logger.Logger.Infof("===[flop|turn]snid:%v 听[同花或者顺子] 尝试加注%v", playerEx.GetSnId(), options)
									}
								}

								if (raise || reraise) && len(options) > 0 {
									opValue = options[rand.Intn(len(options))]
									opCode = rule.DezhouPokerPlayerOpRaise
								}
								logger.Logger.Infof("===[flop|turn]snid:%v raise:%v reraise=%v value=%v options:%v 手牌牌型=%v 公牌牌型=%v", playerEx.GetSnId(), raise, reraise, opValue, options, currCI.KindStr(), commCI.KindStr())
							}
						}
					} else { //弃牌时再权衡下是否要[搏一搏]
						logger.Logger.Infof("===[flop|turn]snid:%v try搏一搏", playerEx.GetSnId())
						byb := false
						if currCI.Kind > rule.KindOfCard_HighCard && currCI.Kind > commCI.Kind {
							byb = true
							logger.Logger.Infof("===[flop|turn]snid:%v try搏一搏 手牌牌型=%v 公牌牌型=%v", playerEx.GetSnId(), currCI.KindStr(), commCI.KindStr())
						}
						if !byb && rule.KindOfCardFigureUpExSington.IsTingKinds(cards, []int32{rule.KindOfCard_Straight, rule.KindOfCard_Flush, rule.KindOfCard_StraightFlush, rule.KindOfCard_RoyalFlush}) {
							byb = true
							logger.Logger.Infof("===[flop|turn]snid:%v try搏一搏 听[同花或者顺子]", playerEx.GetSnId())
						}

						if byb {
							rate := int32(0)
							if minBet <= 20*bigBlind { //<=20BB
								rate = int32(float32(pc.WinningProbability) * 1.5)
							} else if minBet > 20*bigBlind && minBet < 50*bigBlind {
								rate = pc.WinningProbability
							} else {
								rate = pc.WinningProbability / 2
							}
							flag := false
							if rand.Int31n(10000) < rate {
								opCode = rule.DezhouPokerPlayerOpCall
								opValue = minBet
								flag = true
							} else {
								if minBet > 0 {
									opCode = rule.DezhouPokerPlayerOpFold
									opValue = 0
								} else {
									opCode = rule.DezhouPokerPlayerOpCheck
									opValue = 0
								}
							}
							logger.Logger.Infof("===[flop|turn]snid:%v byb rate:%v succ:%v", playerEx.GetSnId(), rate, flag)
						}
					}
				case 5: //river 河牌圈
					if currCI.Kind != pc.CI.Kind || currCI.Value != pc.CI.Value {
						logger.Logger.Info("######bug currCI!=pc.CI", currCI, pc.CI)
					}
					maxCardInfo := gctx.GetMaxCardInfo()
					if pc.WinningProbability == 0 { //已知输
						mutilple := float32(1)
						if sceneEx.LeftAllRobot() {
							mutilple = 2
						}
						if pc.CI.Kind == commCI.Kind {
							mutilple /= 2
						}
						//假装自己比较蠢，概率跟注
						rate := int32(0)
						actDumb := false

						if minBet > 0 {
							if playerEx.IsMarkAIFlag(AI_DEZHOU_FLAG_FRIGHTEN) { //咋呼
								//尝试继续咋呼
								cnt, total := rule.PossibleGreaterKindOfCards(gctx, maxCardInfo)
								maybeRate := cnt * 100 / (total + 1)
								logger.Logger.Infof("===[river]已知输: 再咋呼 snid:%v kindOfCard=%v maxCardInfo=%v maybeRate=%v ", playerEx.GetSnId(), pc.CI.KindStr(), maxCardInfo.KindStr(), maybeRate)

								rate := 0
								startBB := 0
								endBB := 0
								if !actDumb {
									switch commCI.Kind {
									case rule.KindOfCard_HighCard: //高牌
										rate = 30
										startBB = 5
										endBB = 12
									case rule.KindOfCard_OnePair: //一对
										rate = 30
										startBB = 5
										endBB = 12
									case rule.KindOfCard_TwoPair: //两对
										rate = 30
										startBB = 10
										endBB = 20
									case rule.KindOfCard_ThreeKind: //三条
										rate = 30
										startBB = 10
										endBB = 20
									case rule.KindOfCard_Straight: //顺子
										rate = 30
										startBB = 10
										endBB = 20
									case rule.KindOfCard_Flush: //同花
										rate = 30
										startBB = 10
										endBB = 20
									case rule.KindOfCard_Fullhouse: //葫芦
										rate = 30
										startBB = 10
										endBB = 20
									case rule.KindOfCard_FourKind: //金刚
										rate = 30
										startBB = 10
										endBB = 20
									case rule.KindOfCard_StraightFlush: //同花顺
										rate = 30
										startBB = 10
										endBB = 20
									}
								}

								rate = maybeRate / 2
								opValue = chooseRaiseValue(rate, startBB, endBB, bigBlind)
								if opValue > 0 {
									if opValue < minBet*2 {
										opValue = minBet * 2
									}
									opCode = rule.DezhouPokerPlayerOpRaise
									playerEx.MarkAIFlag(AI_DEZHOU_FLAG_FRIGHTEN)
									actDumb = true
								}
								logger.Logger.Infof("===[river]已知输 [再咋呼] snid:%v KindOfCard=%v value=%v rate=%v maybeRate=%v", playerEx.GetSnId(), pc.CI.KindStr(), opValue, rate, maybeRate)
							} else {
								cnt, total := rule.PossibleGreaterKindOfCards(gctx, pc.CI)
								//ai胜率
								maybeRate := 100 - float32((cnt+len(gctx.PlayerCards))*100/(total+1))
								//底池赔率
								potAdds := float32(minBet) * 100 / float32(totalBet)
								//收益率=ai胜率/底池赔率
								rateOfReturn := float32(maybeRate) / float32(potAdds)

								logger.Logger.Infof("===[river]已知输: snid:%v kindOfCard=%v maybeRate=%v rateOfReturn=%v", playerEx.GetSnId(), pc.CI.KindStr(), maybeRate, rateOfReturn)
								if rateOfReturn > 1.0 && maybeRate > 50 { //赔率不够|手牌较弱,不适合跟
									switch pc.CI.Kind {
									case rule.KindOfCard_HighCard: //高牌
										if minBet <= 5*bigBlind {
											rate = int32(25 * mutilple)
										} else if minBet > 5*bigBlind && minBet <= 10*bigBlind {
											rate = int32(10 * mutilple)
										} else if minBet > 10*bigBlind && minBet <= 20*bigBlind {
											rate = int32(5 * mutilple)
										} else if minBet > 20*bigBlind && minBet <= 50*bigBlind {
											rate = 0
										} else if minBet > 50*bigBlind {
											rate = 0
										}
										maybeRate /= 2
									case rule.KindOfCard_OnePair: //一对
										if minBet <= 5*bigBlind {
											rate = int32(35 * mutilple)
										} else if minBet > 5*bigBlind && minBet <= 10*bigBlind {
											rate = int32(15 * mutilple)
										} else if minBet > 10*bigBlind && minBet <= 20*bigBlind {
											rate = int32(5 * mutilple)
										} else if minBet > 20*bigBlind && minBet <= 50*bigBlind {
											rate = 0
										} else if minBet > 50*bigBlind {
											rate = 0
										}
										maybeRate /= 1.5
									case rule.KindOfCard_TwoPair: //两对
										if minBet <= 5*bigBlind {
											rate = int32(45 * mutilple)
										} else if minBet > 5*bigBlind && minBet <= 10*bigBlind {
											rate = int32(20 * mutilple)
										} else if minBet > 10*bigBlind && minBet <= 20*bigBlind {
											rate = int32(10 * mutilple)
										} else if minBet > 20*bigBlind && minBet <= 50*bigBlind {
											rate = int32(5 * mutilple)
										} else if minBet > 50*bigBlind {
											rate = 0
										}
										maybeRate /= 1.2
									case rule.KindOfCard_ThreeKind: //三条
										if minBet <= 5*bigBlind {
											rate = int32(55 * mutilple)
										} else if minBet > 5*bigBlind && minBet <= 10*bigBlind {
											rate = int32(25 * mutilple)
										} else if minBet > 10*bigBlind && minBet <= 20*bigBlind {
											rate = int32(10 * mutilple)
										} else if minBet > 20*bigBlind && minBet <= 50*bigBlind {
											rate = int32(5 * mutilple)
										} else if minBet > 50*bigBlind {
											rate = 0
										}
										maybeRate /= 1.1
									case rule.KindOfCard_Straight: //顺子
										if minBet <= 5*bigBlind {
											rate = int32(65 * mutilple)
										} else if minBet > 5*bigBlind && minBet <= 10*bigBlind {
											rate = int32(30 * mutilple)
										} else if minBet > 10*bigBlind && minBet <= 20*bigBlind {
											rate = int32(10 * mutilple)
										} else if minBet > 20*bigBlind && minBet <= 50*bigBlind {
											rate = int32(5 * mutilple)
										} else if minBet > 50*bigBlind {
											rate = 0
										}
									case rule.KindOfCard_Flush: //同花
										if minBet <= 5*bigBlind {
											rate = int32(75 * mutilple)
										} else if minBet > 5*bigBlind && minBet <= 10*bigBlind {
											rate = int32(30 * mutilple)
										} else if minBet > 10*bigBlind && minBet <= 20*bigBlind {
											rate = int32(10 * mutilple)
										} else if minBet > 20*bigBlind && minBet <= 50*bigBlind {
											rate = int32(5 * mutilple)
										} else if minBet > 50*bigBlind {
											rate = 0
										}
									case rule.KindOfCard_Fullhouse: //葫芦
										if minBet <= 5*bigBlind {
											rate = int32(85 * mutilple)
										} else if minBet > 5*bigBlind && minBet <= 10*bigBlind {
											rate = int32(30 * mutilple)
										} else if minBet > 10*bigBlind && minBet <= 20*bigBlind {
											rate = int32(10 * mutilple)
										} else if minBet > 20*bigBlind && minBet <= 50*bigBlind {
											rate = int32(5 * mutilple)
										} else if minBet > 50*bigBlind {
											rate = 0
										}
									case rule.KindOfCard_FourKind: //金刚
										if minBet <= 5*bigBlind {
											rate = int32(90 * mutilple)
										} else if minBet > 5*bigBlind && minBet <= 10*bigBlind {
											rate = int32(30 * mutilple)
										} else if minBet > 10*bigBlind && minBet <= 20*bigBlind {
											rate = int32(10 * mutilple)
										} else if minBet > 20*bigBlind && minBet <= 50*bigBlind {
											rate = int32(5 * mutilple)
										} else if minBet > 50*bigBlind {
											rate = 0
										}
									case rule.KindOfCard_StraightFlush, rule.KindOfCard_RoyalFlush: //同花顺
										if minBet <= 5*bigBlind {
											rate = int32(100 * mutilple)
										} else if minBet > 5*bigBlind && minBet <= 10*bigBlind {
											rate = int32(30 * mutilple)
										} else if minBet > 10*bigBlind && minBet <= 20*bigBlind {
											rate = int32(10 * mutilple)
										} else if minBet > 20*bigBlind && minBet <= 50*bigBlind {
											rate = int32(5 * mutilple)
										} else if minBet > 50*bigBlind {
											rate = 0
										}
									}
									rate = int32(maybeRate * 2 / 3)
									if rand.Int31n(100) < rate {
										opCode = rule.DezhouPokerPlayerOpCall
										opValue = minBet
										actDumb = true
									}
									logger.Logger.Infof("===[river]已知输: snid:%v kindOfCard=%v rate=%v maybeRate=%v actDumb=%v minBet:%v rateOfReturn=%v", playerEx.GetSnId(), pc.CI.KindStr(), rate, maybeRate, actDumb, minBet, rateOfReturn)
								}
							}
						} else { //可能会主动加注
							actDumb = false
							if currCI.Kind > commCI.Kind { //假装打价值
								cnt, total := rule.PossibleGreaterKindOfCards(gctx, currCI)
								maybeRate := float32(100 - (cnt+len(gctx.PlayerCards))*100/(total+1))
								logger.Logger.Infof("===[river]已知输: snid:%v kindOfCard=%v maybeRate=%v ", playerEx.GetSnId(), pc.CI.KindStr(), maybeRate)
								if maybeRate > 50 { //底池赔率不够或者手牌较弱，适合摊牌
									rate := 0
									startBB := 0
									endBB := 0
									switch commCI.Kind {
									case rule.KindOfCard_HighCard: //[公牌]高牌
										switch currCI.Kind {
										case rule.KindOfCard_OnePair: //[组合]一对
											if rule.IsOverPair(currCI, playerEx.GetCards(), sceneEx.Cards[:commonCardCnt]) || rule.IsTopPair(currCI, playerEx.GetCards(), sceneEx.Cards[:commonCardCnt]) {
												rate = 45
											} else if rule.IsMiddlePair(currCI, playerEx.GetCards(), sceneEx.Cards[:commonCardCnt]) {
												rate = 20
											} else {
												rate = 5
											}
											startBB = 5
											endBB = 10
											//maybeRate /= 2
										case rule.KindOfCard_TwoPair: //[组合]两对
											rate = 30
											startBB = 9
											endBB = 18
											//maybeRate /= 1.5
										case rule.KindOfCard_ThreeKind: //[组合]三条
											rate = 30
											startBB = 6
											endBB = 12
											//maybeRate /= 1.2
										case rule.KindOfCard_Straight: //[组合]顺子
											rate = 40
											startBB = 7
											endBB = 14
										case rule.KindOfCard_Flush: //[组合]同花
											rate = 40
											startBB = 8
											endBB = 16
										case rule.KindOfCard_StraightFlush, rule.KindOfCard_RoyalFlush: //同花顺
											rate = 45
											startBB = 10
											endBB = 20
										}
									case rule.KindOfCard_OnePair: //[公牌]一对
										switch currCI.Kind {
										case rule.KindOfCard_TwoPair: //[组合]两对
											rate = 30
											startBB = 9
											endBB = 18
											//maybeRate /= 1.5
										case rule.KindOfCard_ThreeKind: //[组合]三条
											rate = 35
											startBB = 6
											endBB = 12
											//maybeRate /= 1.2
										case rule.KindOfCard_Fullhouse: //[组合]葫芦
											rate = 45
											startBB = 9
											endBB = 18
										}
									case rule.KindOfCard_TwoPair: //[公牌]两对
										rate = 30
										startBB = 9
										endBB = 18
									case rule.KindOfCard_ThreeKind: //[公牌]三条
										rate = 30
										startBB = 6
										endBB = 12
									case rule.KindOfCard_Straight: //[公牌]顺子
										rate = 30
										startBB = 7
										endBB = 14
									case rule.KindOfCard_Flush: //[公牌]同花
										rate = 30
										startBB = 8
										endBB = 16
									case rule.KindOfCard_Fullhouse: //[公牌]葫芦
										rate = 30
										startBB = 9
										endBB = 18
									case rule.KindOfCard_FourKind: //[公牌]金刚
										rate = 30
										startBB = 10
										endBB = 20
									case rule.KindOfCard_StraightFlush, rule.KindOfCard_RoyalFlush: //[公牌]同花顺
										rate = 0
										startBB = 10
										endBB = 20
									}
									rate = int(maybeRate * 2 / 3)
									opValue = chooseRaiseValue(rate, startBB, endBB, bigBlind)
									if opValue > 0 {
										actDumb = true
										opCode = rule.DezhouPokerPlayerOpRaise
									}
									logger.Logger.Infof("===[river]已知输: [装蠢打价值] snid:%v KindOfCard=%v value=%v actDumb=%v rate=%v maybeRate=%v", playerEx.GetSnId(), pc.CI.KindStr(), opValue, actDumb, rate, maybeRate)
								}
							}

							if !actDumb { //咋呼
								cnt, total := rule.PossibleGreaterKindOfCards(gctx, maxCardInfo)
								maybeRate := cnt * 100 / (total + 1)
								logger.Logger.Infof("===[river]已知输: 咋呼 snid:%v kindOfCard=%v maxCardInfo=%v maybeRate=%v ", playerEx.GetSnId(), pc.CI.KindStr(), maxCardInfo.KindStr(), maybeRate)

								rate := 0
								startBB := 0
								endBB := 0
								if !actDumb {
									switch commCI.Kind {
									case rule.KindOfCard_HighCard: //高牌
										rate = 30
										startBB = 5
										endBB = 12
									case rule.KindOfCard_OnePair: //一对
										rate = 30
										startBB = 5
										endBB = 12
									case rule.KindOfCard_TwoPair: //两对
										rate = 30
										startBB = 10
										endBB = 20
									case rule.KindOfCard_ThreeKind: //三条
										rate = 30
										startBB = 10
										endBB = 20
									case rule.KindOfCard_Straight: //顺子
										rate = 30
										startBB = 10
										endBB = 20
									case rule.KindOfCard_Flush: //同花
										rate = 30
										startBB = 10
										endBB = 20
									case rule.KindOfCard_Fullhouse: //葫芦
										rate = 30
										startBB = 10
										endBB = 20
									case rule.KindOfCard_FourKind: //金刚
										rate = 30
										startBB = 10
										endBB = 20
									case rule.KindOfCard_StraightFlush: //同花顺
										rate = 30
										startBB = 10
										endBB = 20
									}
								}

								rate = maybeRate * 2 / 3
								opValue = chooseRaiseValue(rate, startBB, endBB, bigBlind)
								if opValue > 0 {
									opCode = rule.DezhouPokerPlayerOpRaise
									playerEx.MarkAIFlag(AI_DEZHOU_FLAG_FRIGHTEN)
									actDumb = true
								}
								logger.Logger.Infof("===[river]已知输 [咋呼] snid:%v KindOfCard=%v value=%v rate=%v maybeRate=%v", playerEx.GetSnId(), pc.CI.KindStr(), opValue, rate, maybeRate)
							}
						}
						if !actDumb {
							if minBet > 0 {
								opCode = rule.DezhouPokerPlayerOpFold
								opValue = 0
							} else {
								opCode = rule.DezhouPokerPlayerOpCheck
								opValue = 0
							}
						}
					} else { //已知赢
						if pc.CI != nil {
							cnt, total := rule.PossibleGreaterKindOfCards(gctx, pc.CI)
							maybeRate := (cnt + len(gctx.PlayerCards)) * 100 / (total + 1)
							rate := int32(0)
							beBluff := false
							if minBet > 0 {
								//增加被唬走的概率(被咋呼)
								switch pc.CI.Kind {
								case rule.KindOfCard_HighCard: //高牌
									if minBet <= 4*bigBlind { //<=4BB
										rate = 5
									} else if minBet <= 20*bigBlind { //<=20BB
										rate = 20
									} else if minBet > 20*bigBlind && minBet < 50*bigBlind {
										rate = 40
									} else {
										rate = 60
									}
								case rule.KindOfCard_OnePair: //一对
									if minBet <= 4*bigBlind { //<=4BB
										rate = 0
									} else if minBet <= 20*bigBlind { //<=20BB
										rate = 15
									} else if minBet > 20*bigBlind && minBet < 50*bigBlind {
										rate = 30
									} else {
										rate = 50
									}
								case rule.KindOfCard_TwoPair: //两对
									if minBet <= 4*bigBlind { //<=4BB
										rate = 0
									} else if minBet <= 20*bigBlind { //<=20BB
										rate = 10
									} else if minBet > 20*bigBlind && minBet < 50*bigBlind {
										rate = 20
									} else {
										rate = 30
									}
								case rule.KindOfCard_ThreeKind: //三条
									if commCI.Kind == currCI.Kind {
										if minBet <= 20*bigBlind { //<=20BB
											rate = 10
										} else if minBet > 20*bigBlind && minBet < 50*bigBlind {
											rate = 20
										} else {
											rate = 30
										}
									}
								case rule.KindOfCard_Straight: //顺子
									if commCI.Value == currCI.Value && currCI.ValueToWeight(currCI.KindCards[0]) != 13 {
										if minBet <= 20*bigBlind { //<=20BB
											rate = 5
										} else if minBet > 20*bigBlind && minBet < 50*bigBlind {
											rate = 10
										} else {
											rate = 15
										}
									}
								case rule.KindOfCard_Flush: //同花
									if commCI.Value == currCI.Value {
										if minBet <= 20*bigBlind { //<=20BB
											rate = 5
										} else if minBet > 20*bigBlind && minBet < 50*bigBlind {
											rate = 8
										} else {
											rate = 10
										}
									}
								case rule.KindOfCard_Fullhouse: //葫芦
									if commCI.Value == currCI.Value && commCI.ValueToWeight(commCI.KindCards[3]) > commCI.ValueToWeight(commCI.KindCards[0]) {
										if minBet <= 20*bigBlind { //<=20BB
											rate = 5
										} else if minBet > 20*bigBlind && minBet < 50*bigBlind {
											rate = 10
										} else {
											rate = 15
										}
									}
								case rule.KindOfCard_FourKind: //四条
									if commCI.Kind == currCI.Kind {
										if (currCI.ValueToWeight(currCI.KindCards[0]) == 13 && currCI.ValueToWeight(currCI.KindCards[4]) != 12) || (currCI.ValueToWeight(currCI.KindCards[0]) != 13 && currCI.ValueToWeight(currCI.KindCards[4]) != 13) {
											if minBet <= 20*bigBlind { //<=20BB
												rate = 5
											} else if minBet > 20*bigBlind && minBet < 50*bigBlind {
												rate = 10
											} else {
												rate = 15
											}
										}
									}
								}
								rate = int32(maybeRate)
								if rate > 50 {
									rate = 50
								}
								if rand.Int31n(100) < rate {
									opCode = rule.DezhouPokerPlayerOpFold
									opValue = 0
									beBluff = true
								}
							}
							logger.Logger.Infof("===[river]已知赢 [被咋呼] SnId:%v beBluff rate:%v maybeRate:%v beBluff=%v 手牌牌型=%v 公牌牌型=%v", playerEx.GetSnId(), rate, maybeRate, beBluff, currCI.KindStr(), commCI.KindStr())

							if !beBluff {
								//再加注的概率
								maybeRate = 100 - maybeRate
								raise := false
								reraise := false
								rate := 0
								startBB := 0
								endBB := 0
								//根据牌型选择加注值
								if currCI.Kind > commCI.Kind {
									switch currCI.Kind {
									case rule.KindOfCard_HighCard: //高牌
										if minBet == 0 { //raise
											raise = true
											rate = 10
										} else { //reraise
											reraise = true
											rate = 10
										}
										startBB = 5
										endBB = 10
										maybeRate /= 2
									case rule.KindOfCard_OnePair: //一对
										if minBet == 0 { //raise
											raise = true
											if rule.IsOverPair(currCI, playerEx.GetCards(), sceneEx.Cards[:commonCardCnt]) || rule.IsTopPair(currCI, playerEx.GetCards(), sceneEx.Cards[:commonCardCnt]) {
												rate = 50
											} else if rule.IsMiddlePair(currCI, playerEx.GetCards(), sceneEx.Cards[:commonCardCnt]) {
												rate = 30
											} else {
												rate = 10
											}
											startBB = 5
											endBB = 10
											maybeRate /= 2
										} else { //reraise
											reraise = true
											if rule.IsOverPair(currCI, playerEx.GetCards(), sceneEx.Cards[:commonCardCnt]) || rule.IsTopPair(currCI, playerEx.GetCards(), sceneEx.Cards[:commonCardCnt]) {
												rate = 30
											} else if rule.IsMiddlePair(currCI, playerEx.GetCards(), sceneEx.Cards[:commonCardCnt]) {
												rate = 15
											} else {
												rate = 5
											}
											startBB = 5
											endBB = 10
											maybeRate /= 3
										}
									case rule.KindOfCard_TwoPair: //两对
										if minBet == 0 { //raise
											raise = true
											rate = 70
										} else { //reraise
											reraise = true
											rate = 35
										}
										startBB = 5
										endBB = 30
										maybeRate = maybeRate * 2 / 3
									case rule.KindOfCard_ThreeKind: //三条
										if minBet == 0 { //raise
											raise = true
											rate = 80
										} else { //reraise
											reraise = true
											rate = 40
										}
										startBB = 7
										endBB = 30
									case rule.KindOfCard_Straight: //顺子
										if minBet == 0 { //raise
											raise = true
											rate = 80
										} else { //reraise
											reraise = true
											rate = 40
										}
										startBB = 10
										endBB = 50
									case rule.KindOfCard_Flush: //同花
										if minBet == 0 { //raise
											raise = true
											rate = 80
										} else { //reraise
											reraise = true
											rate = 40
										}
										startBB = 10
										endBB = 100
									case rule.KindOfCard_Fullhouse: //葫芦
										if minBet == 0 { //raise
											raise = true
											rate = 90
										} else { //reraise
											reraise = true
											rate = 45
										}
										startBB = 15
										endBB = 200
									case rule.KindOfCard_FourKind: //金刚
										if minBet == 0 { //raise
											raise = true
											rate = 100
										} else { //reraise
											reraise = true
											rate = 60
										}
										startBB = 20
										endBB = 500
									case rule.KindOfCard_StraightFlush, rule.KindOfCard_RoyalFlush: //同花顺
										if minBet == 0 { //raise
											raise = true
											rate = 100
										} else { //reraise
											reraise = true
											rate = 80
										}
										startBB = 20
										endBB = 1000
									}
								}

								rate = maybeRate
								opValue = chooseRaiseValue(rate, startBB, endBB, bigBlind)
								logger.Logger.Infof("===[river]已知赢 加注|再加注 snid:%v raise:%v reraise=%v opValue:%v 手牌牌型=%v 公牌牌型=%v rate=%v maybeRate=%v", playerEx.GetSnId(), raise, reraise, opValue, currCI.KindStr(), commCI.KindStr(), rate, maybeRate)
								if (raise || reraise) && opValue > 0 && maybeRate > 50 { //手牌强度不够适合摊牌
									if minBet > 0 && opValue < minBet*2 {
										opValue = minBet * 2
									}
									opCode = rule.DezhouPokerPlayerOpRaise
								} else {
									if minBet > 0 { //跟注或者再加注
										opCode = rule.DezhouPokerPlayerOpCall
										opValue = minBet
									} else { //过牌或者加注
										opCode = rule.DezhouPokerPlayerOpCheck
										opValue = 0
									}
								}
							}
						}
					}
				}
			}

			roundNum := DezhouCalOpStateRoundNum(int(sceneEx.GetState()))
			delaySecondMin, delaySecondMax := DezhouCalOpDelaySeconds(roundNum)
			if cost <= delaySecondMin {
				delaySecondMin -= cost
				delaySecondMax -= cost
			} else if cost <= delaySecondMax {
				delaySecondMin = 0
				delaySecondMax -= cost
			}
			pack := &dezhoupoker.CSDezhouPokerPlayerOp{
				OpCode: proto.Int32(opCode),
			}
			pack.OpParam = append(pack.OpParam, opValue)
			proto.SetDefaults(pack)
			base.DelaySend(s, int(dezhoupoker.DZPKPacketID_PACKET_CS_DEZHOUPOKER_OP), pack, delaySecondMin, delaySecondMax)
			logger.Logger.Infof("===snid:%v ai cal take:%vs, opcode=%v oparam=%v", playerEx.GetSnId(), cost, opCode, opValue)
		})).StartByFixExecutor(fmt.Sprintf("dezhouscene_%d", sceneEx.GetRoomId()))
	}
}

func DezhouCalOpStateRoundNum(gameState int) int32 {
	switch gameState {
	case rule.DezhouPokerSceneStateHandCardBet:
		return 1
	case rule.DezhouPokerSceneStateFlopBet:
		return 2
	case rule.DezhouPokerSceneStateTurnBet:
		return 3
	case rule.DezhouPokerSceneStateRiverBet:
		return 4
	}
	return 1
}

/*
轮数=1时：
机器人有80%概率，操作时间=Rand（1-2）秒；
机器人有20%概率，操作时间=Rand（2-4）秒；
轮数=2时：
机器人有70%概率，操作时间=Rand（1-3）秒；
机器人有30%概率，操作时间=Rand（2-8）秒；
轮数=3时：
机器人有70%概率，操作时间=Rand（2-4）秒；
机器人有30%概率，操作时间=Rand（2-6）秒；
轮数=4时：
机器人有50%概率，操作时间=Rand（1-3）秒；
机器人有50%概率，操作时间=Rand（2-4）秒；
*/
func DezhouCalOpDelaySeconds(roundNum int32) (int, int) {
	randValue := rand.Int31n(100)
	switch roundNum {
	case 1:
		if randValue < 80 {
			return 2, 3
		} else {
			return 3, 5
		}
	case 2:
		if randValue < 70 {
			return 2, 4
		} else {
			return 3, 9
		}
	case 3:
		if randValue < 70 {
			return 3, 5
		} else {
			return 4, 6
		}
	case 4:
		if randValue < 50 {
			return 2, 4
		} else {
			return 3, 5
		}
	}
	return 2, 5
}

func init() {
	core.RegisteHook(core.HOOK_BEFORE_START, func() error {
		for _, v := range srvdata.PBDB_DZPKHandCardWinMgr.Datas.Arr {
			DezhouHandCardWinRate[fmt.Sprintf("%s_%d", v.GetHandCard(), v.GetNumber())] = v
		}
		return nil
	})
}
