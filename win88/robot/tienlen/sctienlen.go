package tienlen

import (
	"encoding/json"
	tienlenApi "games.yol.com/win88/api3th/smart/tienlen"
	"games.yol.com/win88/gamerule/tienlen"
	"games.yol.com/win88/proto"
	proto_tienlen "games.yol.com/win88/protocol/tienlen"
	"games.yol.com/win88/protocol/tournament"
	"games.yol.com/win88/robot/base"
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/core/task"
	"sort"
	"strconv"
	"strings"
)

type SCTienLenRoomInfoPacketFactory struct {
}

type SCTienLenRoomInfoHandler struct {
}

func (this *SCTienLenRoomInfoPacketFactory) CreatePacket() interface{} {
	pack := &proto_tienlen.SCTienLenRoomInfo{}
	return pack
}

func (this *SCTienLenRoomInfoHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	logger.Logger.Infof("(this *SCTienLenRoomInfoHandler) Process [%v].", s.GetSessionConfig().Id)
	if msg, ok := pack.(*proto_tienlen.SCTienLenRoomInfo); ok {
		scene := base.SceneMgrSington.GetScene(msg.GetRoomId())
		if scene == nil {
			scene = NewTienLenScene(msg)
			base.SceneMgrSington.AddScene(scene)
		}
		if scene != nil {
			for _, pd := range msg.GetPlayers() {
				if scene.GetPlayerBySnid(pd.GetSnId()) == nil {
					p := NewTienLenPlayer(pd)
					if p != nil {
						scene.AddPlayer(p)
					}
				}
			}
			s.SetAttribute(base.SessionAttributeScene, scene)
		}
	} else {
		logger.Logger.Error("SCTienLenRoomInfo package data error.")
	}
	return nil
}

type SCTienLenPlayerOpPacketFactory struct {
}

type SCTienLenPlayerOpHandler struct {
}

func (this *SCTienLenPlayerOpPacketFactory) CreatePacket() interface{} {
	pack := &proto_tienlen.SCTienLenPlayerOp{}
	return pack
}

func (this *SCTienLenPlayerOpHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	logger.Logger.Tracef("(this *SCTienLenPlayerOpHandler) Process [%v].", s.GetSessionConfig().Id)
	if scTienLenOp, ok := pack.(*proto_tienlen.SCTienLenPlayerOp); ok {
		if scene, ok := base.GetScene(s).(*TienLenScene); ok {
			player := scene.GetMe(s)

			if player == base.NilPlayer {
				return nil
			}
			if me, ok2 := player.(*TienLenPlayer); ok2 && me != TienLenNilPlayer {
				if me.GetSnId() == scTienLenOp.GetSnId() {
					if int(scTienLenOp.GetOpRetCode()) == 0 {
						switch scTienLenOp.GetOpCode() {
						case tienlen.TienLenPlayerOpPlay:
							delCards := scTienLenOp.GetOpParam()
							for _, delcard := range delCards {
								for i, card := range me.Cards {
									if card != tienlen.InvalideCard && card == int32(delcard) {
										me.Cards[i] = tienlen.InvalideCard
									}
								}
							}
						case tienlen.TienLenPlayerOpPass:
						}
					} else { //操作失败
						switch scTienLenOp.GetOpCode() {
						case tienlen.TienLenPlayerOpPlay: //出牌操作失败，改为过
							if scTienLenOp.GetSnId() == me.GetSnId() {
								packOp := &proto_tienlen.CSTienLenPlayerOp{
									OpCode: proto.Int32(2),
								}
								proto.SetDefaults(packOp)
								s.Send(int(proto_tienlen.TienLenPacketID_PACKET_CSTienLenPlayerOp), packOp)
							}
						case int32(tienlen.TienLenPlayerOpPass): //过牌操作失败，改为出
						}
					}
				}
			}
		}
	} else {
		logger.Logger.Error("SCTienLenPlayerOp package data error.")
	}
	return nil
}

type SCTienLenRoomStatePacketFactory struct {
}

type SCTienLenRoomStateHandler struct {
}

func (this *SCTienLenRoomStatePacketFactory) CreatePacket() interface{} {
	pack := &proto_tienlen.SCTienLenRoomState{}
	return pack
}

func (this *SCTienLenRoomStateHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	logger.Logger.Tracef("(this *SCTienLenRoomStateHandler) Process [%v].", s.GetSessionConfig().Id)
	if scTienLenRoomState, ok := pack.(*proto_tienlen.SCTienLenRoomState); ok {
		if scene, ok := s.GetAttribute(base.SessionAttributeScene).(*TienLenScene); ok {
			scene.State = scTienLenRoomState.State
			p := scene.GetMe(s)
			if me, ok2 := p.(*TienLenPlayer); ok2 && me != TienLenNilPlayer {
				switch scTienLenRoomState.GetState() {
				case int32(tienlen.TienLenSceneStateWaitPlayer):
					scene.Clear()
				case int32(tienlen.TienLenSceneStateWaitStart): //等待开始
					scene.Clear()
					if me.GetSnId() == scene.GetMasterSnid() {
						packOp := &proto_tienlen.CSTienLenPlayerOp{
							OpCode: proto.Int32(3),
						}
						proto.SetDefaults(packOp)
						//if scene.GetIsAllAi() {
						//	base.DelayAISend(s, int(proto_tienlen.TienLenPacketID_PACKET_CSTienLenPlayerOp), packOp)
						//} else {
						//	base.DelaySend(s, int(proto_tienlen.TienLenPacketID_PACKET_CSTienLenPlayerOp), packOp, []int{3, 7}...)
						//}
						base.DelaySend(s, int(proto_tienlen.TienLenPacketID_PACKET_CSTienLenPlayerOp), packOp, []int{3, 7}...)
					}
				case int32(tienlen.TienLenSceneStatePlayerOp):
				case int32(tienlen.TienLenSceneStateBilled):
					scene.Clear()
					me.Clear()
				}
			}
		}
	} else {
		logger.Logger.Error("SCTienLenRoomState package data error.")
	}
	return nil
}

type SCTienLenGameBilledPacketFactory struct {
}

type SCTienLenGameBilledHandler struct {
}

func (this *SCTienLenGameBilledPacketFactory) CreatePacket() interface{} {
	pack := &proto_tienlen.SCTienLenGameBilled{}
	return pack
}

func (this *SCTienLenGameBilledHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	logger.Logger.Tracef("(this *SCTienLenGameBilledHandler) Process [%v].", s.GetSessionConfig().Id)
	if scTienLenBilled, ok := pack.(*proto_tienlen.SCTienLenGameBilled); ok {
		if scene, ok := s.GetAttribute(base.SessionAttributeScene).(*TienLenScene); ok {
			//logger.Logger.Trace(scTienLenBilled)
			billData := scTienLenBilled.GetDatas()
			for _, data := range billData {
				p := scene.GetMe(s)
				if p == base.NilPlayer {
					continue
				}
				if me, ok2 := p.(*TienLenPlayer); ok2 && me != TienLenNilPlayer {
					if data.GetSnId() == me.GetSnId() { //自己的数据
						me.Coin = proto.Int64(data.GetGameCoin())
					}
				}
			}
		}
	} else {
		logger.Logger.Error("SCTienLenGameBilled package data error.")
	}
	return nil
}

type SCTienLenCurOpPosPacketFactory struct {
}

type SCTienLenCurOpPosHandler struct {
}

func (this *SCTienLenCurOpPosPacketFactory) CreatePacket() interface{} {
	pack := &proto_tienlen.SCTienLenCurOpPos{}
	return pack
}

func (this *SCTienLenCurOpPosHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	logger.Logger.Infof("(this *SCTienLenCurOpPosHandler) Process [%v].", s.GetSessionConfig().Id)
	if scTienLenCurOpPos, ok := pack.(*proto_tienlen.SCTienLenCurOpPos); ok {
		if scene, ok := s.GetAttribute(base.SessionAttributeScene).(*TienLenScene); ok {
			curPos := scTienLenCurOpPos.GetPos()
			p := scene.GetMe(s)
			if me, ok2 := p.(*TienLenPlayer); ok2 && me != TienLenNilPlayer {
				if scene.GetState() == int32(tienlen.TienLenSceneStatePlayerOp) {
					if me.GetPos() == curPos {
						packOp := &proto_tienlen.CSTienLenPlayerOp{
							OpCode: proto.Int32(1),
						}
						cpCards := []int32{}
						for _, card := range me.Cards {
							if card != tienlen.InvalideCard {
								cpCards = append(cpCards, card)
							}
						}
						recmCards := []int32{}

						exDelayTs := int(scTienLenCurOpPos.GetExDelay()) * 1000
						minS := 1500
						maxS := 3000
						if scene.IsMatchScene() && scene.GetIsAllAi() {
							exDelayTs = 1
							minS = 1
							maxS = 2
						}
						notExDelay := true //立即出牌
						notExDelayTs := 1
						notExDelayminS := 1
						notExDelaymaxS := 2
						if len(cpCards) > 0 {
							if tienlenApi.Config.Switch() && (!scene.IsMatchScene() || (scene.IsMatchScene() && !scene.GetIsAllAi())) { //开启 不是比赛场或者是比赛场但不是纯ai
								if me.data != nil && me.data.Num_cards_left_0 == 13 &&
									me.data.Num_cards_left_1 == 13 &&
									me.data.Num_cards_left_2 == 13 &&
									me.data.Num_cards_left_3 == 13 &&
									me.data.Card_play_action_seq == "" {
									//根据最小牌值去推荐出牌:顺子>2连对>三张>对子>单张
									recmCards = tienlen.RecommendCardsWithMinCard(cpCards)
									if len(recmCards) == 0 {
										logger.Logger.Error("RecommendCardsWithMinCard error1: ", me.GetSnId(), " recmCards:", recmCards)
									}
									if len(recmCards) > 0 {
										for _, card := range recmCards {
											packOp.OpParam = append(packOp.OpParam, int64(card))
										}
									} else {
										packOp.OpCode = proto.Int32(2)
									}
									proto.SetDefaults(packOp)
									if notExDelay && len(recmCards) == len(cpCards) { //最后一手牌立即打出
										exDelayTs = notExDelayTs
										minS = notExDelayminS
										maxS = notExDelaymaxS
									}
									base.DelaySendNewMillisecond(s, int(proto_tienlen.TienLenPacketID_PACKET_CSTienLenPlayerOp), packOp, []int{exDelayTs + minS, exDelayTs + maxS}...)
								} else {
									var err error
									var res []byte
									if me == nil {
										return nil
									}

									task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
										res, err = tienlenApi.Config.Do(tienlenApi.Predict, me.data)
										logger.Logger.Info("AI返回数据：", string(res), err)

										return nil
									}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
										if scene.GetState() != int32(tienlen.TienLenSceneStatePlayerOp) || me.GetPos() != curPos {
											//tienlenApi.Config.Log().Error("TienLen TienLenSmart return too later")
											return
										}

										if err != nil || res == nil {
											//tienlenApi.Config.Log().Errorf("TienLen TienLenSmart Err:%v", err)
											lastDelCards := scTienLenCurOpPos.GetCards()
											if len(lastDelCards) != 0 && !scTienLenCurOpPos.GetIsNew() { //根据上手牌出牌
												logger.Logger.Info("根据上手牌出牌: ", me.GetSnId(), " lastDelCards:", lastDelCards, " IsNew:", scTienLenCurOpPos.GetIsNew())
												recmCards = tienlen.RecommendCardsWithLastCards(lastDelCards, cpCards)
											} else { //自由出牌
												logger.Logger.Info("自由出牌: ", me.GetSnId(), " lastDelCards:", lastDelCards, " IsNew:", scTienLenCurOpPos.GetIsNew())
												if len(cpCards) > 5 {
													//根据最小牌值去推荐出牌:顺子>2连对>三张>对子>单张
													recmCards = tienlen.RecommendCardsWithMinCard(cpCards)
													if len(recmCards) == 0 {
														logger.Logger.Error("RecommendCardsWithMinCard error2: ", me.GetSnId(), " lastDelCards:", lastDelCards, " recmCards:", recmCards)
													}
												} else {
													//根据牌型牌数量最多去推荐出牌
													recmCards = tienlen.RecommendCardsWithCards(cpCards)
													if len(recmCards) == 0 {
														logger.Logger.Error("RecommendCardsWithCards error3: ", me.GetSnId(), " lastDelCards:", lastDelCards, " recmCards:", recmCards)
													}
												}
											}
											if len(recmCards) > 0 {
												for _, card := range recmCards {
													packOp.OpParam = append(packOp.OpParam, int64(card))
												}
											} else {
												packOp.OpCode = proto.Int32(2)
											}
											proto.SetDefaults(packOp)
											if notExDelay && len(recmCards) == len(cpCards) { //最后一手牌立即打出
												exDelayTs = notExDelayTs
												minS = notExDelayminS
												maxS = notExDelaymaxS
											}
											base.DelaySendNewMillisecond(s, int(proto_tienlen.TienLenPacketID_PACKET_CSTienLenPlayerOp), packOp, []int{exDelayTs + minS, exDelayTs + maxS}...)
											return
										}
										PredictResponse := new(tienlenApi.PredictResponse)
										if err = json.Unmarshal(res, PredictResponse); err != nil {
											//tienlenApi.Config.Log().Errorf("TienLen TienLenSmart json.Unmarshal() error: %v", err)
											return
										}
										//tienlenApi.Config.Log().Infof("PredictResponse：%v", *PredictResponse)
										cardstr := ""
										maxwinrate := 0.0
										for k, v := range PredictResponse.WinRates {
											v_rate, errrate := strconv.ParseFloat(v, 32)
											if errrate != nil {
												//tienlenApi.Config.Log().Errorf("TienLen TienLenSmart strconv.ParseFloat() error: %v", err)
												continue
											}
											if v_rate > maxwinrate {
												maxwinrate = v_rate
												cardstr = k
											}
										}
										recmCardss := strings.Split(cardstr, ",")
										for _, v := range recmCardss {
											card, errcard := strconv.Atoi(v)
											if errcard == nil {
												recmCards = append(recmCards, int32(card))
											}
										}
										tienlenApi.Config.Log().Infof("--> %v Start TienLen:%+v", me.GetSnId(), cardstr)
										if len(recmCards) > 0 {
											for _, card := range recmCards {
												packOp.OpParam = append(packOp.OpParam, int64(tienlenApi.AiCardToCard[card]))
											}
										} else {
											packOp.OpCode = proto.Int32(2)
										}
										tienlenApi.Config.Log().Infof("--> %v Start TienLen:%+v", me.GetSnId(), packOp.GetOpParam())
										proto.SetDefaults(packOp)
										if notExDelay && len(recmCards) == len(cpCards) { //最后一手牌立即打出
											exDelayTs = notExDelayTs
											minS = notExDelayminS
											maxS = notExDelaymaxS
										}
										base.DelaySendNewMillisecond(s, int(proto_tienlen.TienLenPacketID_PACKET_CSTienLenPlayerOp), packOp, []int{exDelayTs + minS, exDelayTs + maxS}...)
									}), "SmartStart_TienLen_action").Start()

								}
							} else {
								lastDelCards := scTienLenCurOpPos.GetCards()
								if len(lastDelCards) != 0 && !scTienLenCurOpPos.GetIsNew() { //根据上手牌出牌
									logger.Logger.Info("根据上手牌出牌: ", me.GetSnId(), " lastDelCards:", lastDelCards, " IsNew:", scTienLenCurOpPos.GetIsNew())
									recmCards = tienlen.RecommendCardsWithLastCards(lastDelCards, cpCards)
								} else { //自由出牌
									logger.Logger.Info("自由出牌: ", me.GetSnId(), " lastDelCards:", lastDelCards, " IsNew:", scTienLenCurOpPos.GetIsNew())
									if len(cpCards) > 5 {
										//根据最小牌值去推荐出牌:顺子>2连对>三张>对子>单张
										recmCards = tienlen.RecommendCardsWithMinCard(cpCards)
										if len(recmCards) == 0 {
											logger.Logger.Error("RecommendCardsWithMinCard error2: ", me.GetSnId(), " lastDelCards:", lastDelCards, " recmCards:", recmCards)
										}
									} else {
										//根据牌型牌数量最多去推荐出牌
										recmCards = tienlen.RecommendCardsWithCards(cpCards)
										if len(recmCards) == 0 {
											logger.Logger.Error("RecommendCardsWithCards error3: ", me.GetSnId(), " lastDelCards:", lastDelCards, " recmCards:", recmCards)
										}
									}
								}
								if len(recmCards) > 0 {
									for _, card := range recmCards {
										packOp.OpParam = append(packOp.OpParam, int64(card))
									}
								} else {
									packOp.OpCode = proto.Int32(2)
								}
								proto.SetDefaults(packOp)
								if notExDelay && len(recmCards) == len(cpCards) { //最后一手牌立即打出
									exDelayTs = notExDelayTs
									minS = notExDelayminS
									maxS = notExDelaymaxS
								}
								base.DelaySendNewMillisecond(s, int(proto_tienlen.TienLenPacketID_PACKET_CSTienLenPlayerOp), packOp, []int{exDelayTs + minS, exDelayTs + maxS}...)
							}
						}
					}
				}
			}
		}
	} else {
		logger.Logger.Error("SCTienLenCurOpPosHandler package data error.")
	}
	return nil
}

type SCTienLenCardPacketFactory struct {
}

type SCTienLenCardHandler struct {
}

func (this *SCTienLenCardPacketFactory) CreatePacket() interface{} {
	pack := &proto_tienlen.SCTienLenCard{}
	return pack
}

func (this *SCTienLenCardHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	logger.Logger.Tracef("(this *SCTienLenCardHandler) Process [%v].", s.GetSessionConfig().Id)
	if scTienLenCard, ok := pack.(*proto_tienlen.SCTienLenCard); ok {
		if scene, ok := s.GetAttribute(base.SessionAttributeScene).(*TienLenScene); ok {
			cards := scTienLenCard.GetCards()
			p := scene.GetMe(s)
			if me, ok2 := p.(*TienLenPlayer); ok2 && me != TienLenNilPlayer {
				if scene.GetState() == int32(tienlen.TienLenSceneStateHandCard) {
					me.Cards = []int32{}
					for _, card := range cards {
						me.Cards = append(me.Cards, card)
					}
					sort.Slice(me.Cards, func(i, j int) bool {
						if me.Cards[i] > me.Cards[j] {
							return false
						}
						return true
					})
				}
			}
		}
	} else {
		logger.Logger.Error("SCTienLenCardHandler package SCTienLenCard error.")
	}
	return nil
}

type SCTienLenUpdateMasterSnidPacketFactory struct {
}

type SCTienLenUpdateMasterSnidHandler struct {
}

func (this *SCTienLenUpdateMasterSnidPacketFactory) CreatePacket() interface{} {
	pack := &proto_tienlen.SCTienLenUpdateMasterSnid{}
	return pack
}

func (this *SCTienLenUpdateMasterSnidHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	logger.Logger.Tracef("(this *SCTienLenUpdateMasterSnidHandler) Process [%v].", s.GetSessionConfig().Id)
	if scTienLenUpdateMasterSnid, ok := pack.(*proto_tienlen.SCTienLenUpdateMasterSnid); ok {
		if scene, ok := s.GetAttribute(base.SessionAttributeScene).(*TienLenScene); ok {
			masterSnid := scTienLenUpdateMasterSnid.GetMasterSnid()
			scene.MasterSnid = masterSnid //更新房主
		}
	} else {
		logger.Logger.Error("SCTienLenUpdateMasterSnidHandler package SCTienLenUpdateMasterSnid error.")
	}
	return nil
}

type SCTienLenAIPacketFactory struct {
}

type SCTienLenAIHandler struct {
}

func (this *SCTienLenAIPacketFactory) CreatePacket() interface{} {
	pack := &proto_tienlen.SCTienLenAIData{}
	return pack
}

func (this *SCTienLenAIHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	logger.Logger.Tracef("(this *SCTienLenAIHandler) Process [%v].", s.GetSessionConfig().Id)
	if scTienLenAI, ok := pack.(*proto_tienlen.SCTienLenAIData); ok {
		if scene, ok := s.GetAttribute(base.SessionAttributeScene).(*TienLenScene); ok {
			//logger.Logger.Infof("(this *SCTienLenAIHandler) Process [%v] %v", scene.GetRoomId(), scTienLenAI)
			p := scene.GetMe(s)
			if me, ok2 := p.(*TienLenPlayer); ok2 && me != TienLenNilPlayer {
				if me.GetPos() == scTienLenAI.GetPlayerPosition() {
					me.data = &tienlenApi.PredictRequest{
						Bomb_num:             0,
						Card_play_action_seq: scTienLenAI.GetCardPlayActionSeq(),
						Last_move_0:          scTienLenAI.GetLastMove_0(),
						Last_move_1:          scTienLenAI.GetLastMove_1(),
						Last_move_2:          scTienLenAI.GetLastMove_2(),
						Last_move_3:          scTienLenAI.GetLastMove_3(),
						Num_cards_left_0:     int(scTienLenAI.GetNumCardsLeft_0()),
						Num_cards_left_1:     int(scTienLenAI.GetNumCardsLeft_1()),
						Num_cards_left_2:     int(scTienLenAI.GetNumCardsLeft_2()),
						Num_cards_left_3:     int(scTienLenAI.GetNumCardsLeft_3()),
						Other_hand_cards:     scTienLenAI.GetOtherHandCards(),
						Played_cards_0:       scTienLenAI.GetPlayedCards_0(),
						Played_cards_1:       scTienLenAI.GetPlayedCards_1(),
						Played_cards_2:       scTienLenAI.GetPlayedCards_2(),
						Played_cards_3:       scTienLenAI.GetPlayedCards_3(),
						Player_hand_cards:    scTienLenAI.GetPlayerHandCards(),
						Player_position:      int(scTienLenAI.GetPlayerPosition()),
					}
				}
			}
		}
	} else {
		logger.Logger.Error("SCTienLenAIHandler package SCTienLenAIData error.")
	}
	return nil
}

type SCSignRacePacketFactory struct {
}

type SCSignRaceHandler struct {
}

func (this *SCSignRacePacketFactory) CreatePacket() interface{} {
	pack := &tournament.SCSignRace{}
	return pack
}

func (this *SCSignRaceHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	logger.Logger.Tracef("(this *SCSignRaceHandler) Process [%v].", s.GetSessionConfig().Id)
	if msg, ok := pack.(*tournament.SCSignRace); ok {
		if msg.OpCode == 0 && msg.RetCode != 0 && msg.RetCode != 1 {
			logger.Logger.Info("SCSignRaceHandler 报名失败")
			s.RemoveAttribute(base.SessionAttributeWaitingMatch)
		}
		if msg.OpCode == 1 && msg.RetCode == 0 {
			logger.Logger.Info("SCSignRaceHandler 取消报名")
			s.RemoveAttribute(base.SessionAttributeWaitingMatch)
		}
	} else {
		logger.Logger.Error("SCSignRaceHandler package SCSignRace error.")
	}
	return nil
}

type SCTMStartPacketFactory struct {
}

type SCTMStartHandler struct {
}

func (this *SCTMStartPacketFactory) CreatePacket() interface{} {
	pack := &tournament.SCTMStart{}
	return pack
}

func (this *SCTMStartHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	logger.Logger.Tracef("(this *SCTMStartHandler) Process [%v].", s.GetSessionConfig().Id)
	if msg, ok := pack.(*tournament.SCTMStart); ok {
		s.RemoveAttribute(base.SessionAttributeWaitingMatch)
		s.SetAttribute(base.SessionAttributeMatchDoing, msg.MatchId)
	} else {
		logger.Logger.Error("SCTMStartHandler package SCTMStart error.")
	}
	return nil
}

type SCTMStopPacketFactory struct {
}

type SCTMStopHandler struct {
}

func (this *SCTMStopPacketFactory) CreatePacket() interface{} {
	pack := &tournament.SCTMStop{}
	return pack
}

func (this *SCTMStopHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	logger.Logger.Tracef("(this *SCTMStopHandler) Process [%v].", s.GetSessionConfig().Id)
	if _, ok := pack.(*tournament.SCTMStop); ok {
		s.RemoveAttribute(base.SessionAttributeMatchDoing)
		s.RemoveAttribute(base.SessionAttributeWaitingMatch)
	} else {
		logger.Logger.Error("SCTMStopHandler package SCTMStop error.")
	}
	return nil
}

func init() {
	//SCTienLenRoomInfo
	netlib.RegisterHandler(int(proto_tienlen.TienLenPacketID_PACKET_SCTienLenRoomInfo), &SCTienLenRoomInfoHandler{})
	netlib.RegisterFactory(int(proto_tienlen.TienLenPacketID_PACKET_SCTienLenRoomInfo), &SCTienLenRoomInfoPacketFactory{})
	//SCTienLenPlayerOp
	netlib.RegisterHandler(int(proto_tienlen.TienLenPacketID_PACKET_SCTienLenPlayerOp), &SCTienLenPlayerOpHandler{})
	netlib.RegisterFactory(int(proto_tienlen.TienLenPacketID_PACKET_SCTienLenPlayerOp), &SCTienLenPlayerOpPacketFactory{})
	//SCTienLenRoomState
	netlib.RegisterHandler(int(proto_tienlen.TienLenPacketID_PACKET_SCTienLenRoomState), &SCTienLenRoomStateHandler{})
	netlib.RegisterFactory(int(proto_tienlen.TienLenPacketID_PACKET_SCTienLenRoomState), &SCTienLenRoomStatePacketFactory{})
	//SCTienLenFinalBilled
	netlib.RegisterHandler(int(proto_tienlen.TienLenPacketID_PACKET_SCTienLenGameBilled), &SCTienLenGameBilledHandler{})
	netlib.RegisterFactory(int(proto_tienlen.TienLenPacketID_PACKET_SCTienLenGameBilled), &SCTienLenGameBilledPacketFactory{})
	//SCTienLenCurOpPos
	netlib.RegisterHandler(int(proto_tienlen.TienLenPacketID_PACKET_SCTienLenCurOpPos), &SCTienLenCurOpPosHandler{})
	netlib.RegisterFactory(int(proto_tienlen.TienLenPacketID_PACKET_SCTienLenCurOpPos), &SCTienLenCurOpPosPacketFactory{})
	//SCTienLenCard
	netlib.RegisterHandler(int(proto_tienlen.TienLenPacketID_PACKET_SCTienLenCard), &SCTienLenCardHandler{})
	netlib.RegisterFactory(int(proto_tienlen.TienLenPacketID_PACKET_SCTienLenCard), &SCTienLenCardPacketFactory{})
	//SCTienLenUpdateMasterSnid
	netlib.RegisterHandler(int(proto_tienlen.TienLenPacketID_PACKET_SCTienLenUpdateMasterSnid), &SCTienLenUpdateMasterSnidHandler{})
	netlib.RegisterFactory(int(proto_tienlen.TienLenPacketID_PACKET_SCTienLenUpdateMasterSnid), &SCTienLenUpdateMasterSnidPacketFactory{})
	//SCTienLenAIData
	netlib.RegisterHandler(int(proto_tienlen.TienLenPacketID_PACKET_SCTienLenAI), &SCTienLenAIHandler{})
	netlib.RegisterFactory(int(proto_tienlen.TienLenPacketID_PACKET_SCTienLenAI), &SCTienLenAIPacketFactory{})
	//报名 TOURNAMENTID_PACKET_TM_SCSignRace
	netlib.RegisterHandler(int(tournament.TOURNAMENTID_PACKET_TM_SCSignRace), &SCSignRaceHandler{})
	netlib.RegisterFactory(int(tournament.TOURNAMENTID_PACKET_TM_SCSignRace), &SCSignRacePacketFactory{})
	//比赛开始 TOURNAMENTID_PACKET_TM_SCTMStart
	netlib.RegisterHandler(int(tournament.TOURNAMENTID_PACKET_TM_SCTMStart), &SCTMStartHandler{})
	netlib.RegisterFactory(int(tournament.TOURNAMENTID_PACKET_TM_SCTMStart), &SCTMStartPacketFactory{})
	//比赛结束 TOURNAMENTID_PACKET_TM_SCTMStop
	netlib.RegisterHandler(int(tournament.TOURNAMENTID_PACKET_TM_SCTMStop), &SCTMStopHandler{})
	netlib.RegisterFactory(int(tournament.TOURNAMENTID_PACKET_TM_SCTMStop), &SCTMStopPacketFactory{})
}
