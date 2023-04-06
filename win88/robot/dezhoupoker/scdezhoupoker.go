package dezhoupoker

import (
	"fmt"
	"games.yol.com/win88/common"
	rule "games.yol.com/win88/gamerule/dezhoupoker"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/dezhoupoker"
	"games.yol.com/win88/robot/base"
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/core/task"
)

//DZPKPacketID_PACKET_SC_DEZHOUPOKER_ROOMINFO
type SCDezhouPokerRoomInfoPacketFactory struct {
}

type SCDezhouPokerRoomInfoHandler struct {
}

func (this *SCDezhouPokerRoomInfoPacketFactory) CreatePacket() interface{} {
	pack := &dezhoupoker.SCDezhouPokerRoomInfo{}
	return pack
}

func (this *SCDezhouPokerRoomInfoHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	logger.Logger.Trace("(this *SCDezhouPokerRoomInfoHandler) Process ", s.GetSessionConfig().Id, pack)
	if msg, ok := pack.(*dezhoupoker.SCDezhouPokerRoomInfo); ok {
		scene := base.SceneMgrSington.GetScene(msg.GetRoomId())
		if scene == nil {
			scene = NewDezhouPokerScene(msg)
			base.SceneMgrSington.AddScene(scene)
		}
		if scene != nil {
			for _, pd := range msg.GetPlayers() {
				if scene.GetPlayerBySnid(pd.GetSnId()) == nil {
					p := NewDezhouPokerPlayer(pd)
					if p != nil {
						scene.AddPlayer(p)
					}
				}
			}
			//logger.Logger.Trace(msg)
			s.SetAttribute(base.SessionAttributeSceneId, scene.GetRoomId())
		}
	} else {
		logger.Logger.Error("SCDezhouPokerRoomInfo package data error.")
	}
	return nil
}

//DZPKPacketID_PACKET_SC_DEZHOUPOKER_ROOMSTATE
type SCDezhouPokerRoomStatePacketFactory struct {
}

type SCDezhouPokerRoomStateHandler struct {
}

func (this *SCDezhouPokerRoomStatePacketFactory) CreatePacket() interface{} {
	pack := &dezhoupoker.SCDezhouPokerRoomState{}
	return pack
}

func (this *SCDezhouPokerRoomStateHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	logger.Logger.Trace("(this *SCDezhouPokerRoomStateHandler) Process ", s.GetSessionConfig().Id, pack)
	if scDezhouPokerRoomState, ok := pack.(*dezhoupoker.SCDezhouPokerRoomState); ok {
		if scene, ok := base.GetScene(s).(*DezhouPokerScene); ok {
			switch int(scDezhouPokerRoomState.GetState()) {
			case rule.DezhouPokerSceneStateWaitPlayer:
				scene.Clear()
			case rule.DezhouPokerSceneStateSelectBankerAndBlinds:
				scene.OnNewGame()
			case rule.DezhouPokerSceneStateGameEnd:
			case rule.DezhouPokerSceneStateSelectCard:
				scene.SelectCard(s)
			default:
			}
			scene.State = proto.Int32(scDezhouPokerRoomState.GetState())
		}
	} else {
		logger.Logger.Error("SCDezhouPokerRoomStateHandler package data error.")
	}
	return nil
}

//DZPKPacketID_PACKET_SC_DEZHOUPOKER_OP
type SCDezhouPokerPlayerOpPacketFactory struct {
}

type SCDezhouPokerPlayerOpHandler struct {
}

func (this *SCDezhouPokerPlayerOpPacketFactory) CreatePacket() interface{} {
	pack := &dezhoupoker.SCDezhouPokerPlayerOp{}
	return pack
}

func (this *SCDezhouPokerPlayerOpHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	logger.Logger.Trace("(this *SCDezhouPokerPlayerOpHandler) Process ", s.GetSessionConfig().Id, pack)
	if scDezhouPokerOp, ok := pack.(*dezhoupoker.SCDezhouPokerPlayerOp); ok {
		if scDezhouPokerOp.GetOpRetCode() != dezhoupoker.OpResultCode_OPRC_Sucess {
			return nil
		}

		//logger.Logger.Trace(scDezhouPokerOp)
		if scene, ok := base.GetScene(s).(*DezhouPokerScene); ok {
			player := scene.GetPlayerByPos(scDezhouPokerOp.GetPos()).(*DezhouPokerPlayer)
			if player != DezhouPokerNilPlayer {
				player.SetLastOp(scDezhouPokerOp.GetOpCode())
			}
			//var me *DezhouPokerPlayer
			//if scene.GetMe(s) != nil {
			//	me = scene.GetMe(s).(*DezhouPokerPlayer)
			//} else {
			//	return nil
			//}
			//pos := scDezhouPokerOp.GetPos()
			//if me.GetPos() != pos {
			//	return nil
			//}
			////bet, curRoundTotalbet, gameCoin, totalbet
			//player := scene.GetPlayerByPos(pos).(*DezhouPokerPlayer)
			//if player != DezhouPokerNilPlayer && me != DezhouPokerNilPlayer {
			//	player.SetLastOp(scDezhouPokerOp.GetOpCode())
			//}
		}
	} else {
		logger.Logger.Error("SCDezhouPokerPlayerOp package data error.")
	}
	return nil
}

// DZPKPacketID_PACKET_SC_DEZHOUPOKER_PLAYERENTER
type SCDezhouPokerPlayerEnterPacketFactory struct {
}

type SCDezhouPokerPlayerEnterHandler struct {
}

func (this *SCDezhouPokerPlayerEnterPacketFactory) CreatePacket() interface{} {
	pack := &dezhoupoker.SCDezhouPokerPlayerEnter{}
	return pack
}

func (this *SCDezhouPokerPlayerEnterHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	logger.Logger.Trace("(this *SCDezhouPokerPlayerEnterHandler) Process ", s.GetSessionConfig().Id, pack)
	if msg, ok := pack.(*dezhoupoker.SCDezhouPokerPlayerEnter); ok {
		if scene := base.GetScene(s); scene != nil {
			if oldPlayer, ok := scene.GetPlayerBySnid(msg.GetData().GetSnId()).(*DezhouPokerPlayer); ok {
				p := NewDezhouPokerPlayer(msg.GetData())
				p.OpDelayTimes = oldPlayer.OpDelayTimes
				if p != nil {
					scene.AddPlayer(p)
				}
			} else {
				p := NewDezhouPokerPlayer(msg.GetData())
				if p != nil {
					scene.AddPlayer(p)
				}
			}
		}
	} else {
		logger.Logger.Error("SCDezhouPokerPlayerEnter package data error.")
	}
	return nil
}

//DZPKPacketID_PACKET_SC_DEZHOUPOKER_PLAYERLEAVE
type SCDezhouPokerPlayerLeavePacketFactory struct {
}

type SCDezhouPokerPlayerLeaveHandler struct {
}

func (this *SCDezhouPokerPlayerLeavePacketFactory) CreatePacket() interface{} {
	pack := &dezhoupoker.SCDezhouPokerPlayerLeave{}
	return pack
}

func (this *SCDezhouPokerPlayerLeaveHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	logger.Logger.Trace("(this *SCDezhouPokerPlayerLeaveHandler) Process ", s.GetSessionConfig().Id, pack)

	if scDezhouPokerPlayerLeave, ok := pack.(*dezhoupoker.SCDezhouPokerPlayerLeave); ok {
		if scene, ok := base.GetScene(s).(*DezhouPokerScene); ok && scene != nil {
			p := scene.GetPlayerByPos(scDezhouPokerPlayerLeave.GetPos())
			if p != nil {
				if player, ok := p.(*DezhouPokerPlayer); ok && player != DezhouPokerNilPlayer {
					scene.DelPlayer(player.GetSnId())
				}
			}
		}
	} else {
		logger.Logger.Error("scDezhouPokerPlayerLeave package data error.")
	}
	return nil
}

//DZPKPacketID_PACKET_SC_DEZHOUPOKER_OP
type SCDezhouPokerNextOperNotifyPacketFactory struct {
}

type SCDezhouPokerNextOperNotifyHandler struct {
}

func (this *SCDezhouPokerNextOperNotifyPacketFactory) CreatePacket() interface{} {
	pack := &dezhoupoker.SCDezhouPokerOperNotify{}
	return pack
}

func (this *SCDezhouPokerNextOperNotifyHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	logger.Logger.Trace("(this *SCDezhouPokerNextOperNotifyHandler) Process ", s.GetSessionConfig().Id, pack)
	if scNextOperNotify, ok := pack.(*dezhoupoker.SCDezhouPokerOperNotify); ok {
		if scene, ok := base.GetScene(s).(*DezhouPokerScene); ok {
			scene.SetDezhouPokerOperNotify(scNextOperNotify)

			p := scene.GetMe(s)
			if p == nil {
				return nil
			}
			if p.GetPos() != scNextOperNotify.GetNextOperPos() {
				return nil
			}
			if me, ok := p.(*DezhouPokerPlayer); ok && me != DezhouPokerNilPlayer {
				//选择一个操作。
				minBet := scNextOperNotify.GetRoundMaxBet() - scNextOperNotify.GetNextOperCurRoundTotalBet()
				minRaise := scNextOperNotify.GetMinRaise()
				//opCode, opValue := this.CalOpAndValue(scene, me, minRaise, minBet)
				////发送请求
				//logger.Logger.Trace("Pos:", p.GetPos(), " Snid:", p.GetSnId(), " opCode:", opCode, " opValue:", opValue)
				//
				//roundNum := this.CalOpStateRoundNum(int(scene.GetState()))
				//delaySecondMin, delaySecondMax := this.CalOpDelaySeconds(roundNum)
				//
				//pack := &protocol.CSDezhouPokerPlayerOp{
				//	OpCode: proto.Int32(opCode),
				//}
				//pack.OpParam = append(pack.OpParam, opValue)
				//proto.SetDefaults(pack)
				//DelaySend(s, int(protocol.DZPKPacketID_PACKET_CS_DEZHOUPOKER_OP), pack, delaySecondMin, delaySecondMax)
				DezhouCalOpAndValueByEV(s, scene, me, minRaise, minBet, scNextOperNotify.GetBetChips(), scNextOperNotify.GetPot(), scNextOperNotify.GetRaiseOption(), scNextOperNotify.GetRemainChips(), scNextOperNotify.GetNextOperCurRoundTotalBet(), scNextOperNotify.GetCurrRoundPerBet(), scNextOperNotify.GetTotalPlayers(), scNextOperNotify.GetRestPlayerCnt(), scNextOperNotify.GetRolePos())
			} else {
				logger.Logger.Error("DezhouPokerPlayer data error.")
			}
		}
	} else {
		logger.Logger.Error("scNextOperNotify package data error.")
	}
	return nil
}

func (this *SCDezhouPokerNextOperNotifyHandler) CalOpAndValue(sceneEx *DezhouPokerScene, playerEx *DezhouPokerPlayer, minRaise, minBet int64) (int32, int64) {
	if sceneEx != nil && playerEx != nil {
		if len(playerEx.GetCards()) == int(rule.HandCardNum) && len(sceneEx.GetCards()) == int(rule.CommunityCardNum) {
			var handCards [rule.HandCardNum]int32
			for i := 0; i < len(playerEx.GetCards()); i++ {
				handCards[i] = playerEx.GetCards()[i]
			}
			var commonCard [rule.CommunityCardNum]int32
			for i := 0; i < len(sceneEx.GetCards()); i++ {
				commonCard[i] = sceneEx.GetCards()[i]
			}
			cardsInfo := rule.KindOfCardFigureUpSington.FigureUpByCard(handCards, commonCard)
			if cardsInfo != nil {
				cardsInfo.ValueScore = rule.CalCardsKindScore(cardsInfo)
				curRoomState := sceneEx.GetState()
				curRoundNum := DezhouCalOpStateRoundNum(int(curRoomState))
				logger.Logger.Trace("playerEx:", playerEx.GetSnId())
				return CalOpKindAndOpValue(sceneEx.AIMode, curRoundNum, cardsInfo.ValueScore, minRaise, minBet, sceneEx.SCDezhouPokerRoomInfo.GetSmallBlind()*2)
			}
		}
	}
	logger.Logger.Trace("DezhouPokerPlayerOpFold, 0")
	return rule.DezhouPokerPlayerOpFold, 0
}

//DZPKPacketID_PACKET_SC_DEZHOUPOKER_BANKANDBLINDPOS
type SCDezhouPokerBankerAndBlindPosPacketFactory struct {
}

type SCDezhouPokerBankerAndBlindPosHandler struct {
}

func (this *SCDezhouPokerBankerAndBlindPosPacketFactory) CreatePacket() interface{} {
	pack := &dezhoupoker.SCDezhouPokerBankerAndBlindPos{}
	return pack
}

func (this *SCDezhouPokerBankerAndBlindPosHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	logger.Logger.Trace("(this *SCDezhouPokerBankerAndBlindPosHandler) Process ", s.GetSessionConfig().Id, pack)
	if scBankerAndBlindPos, ok := pack.(*dezhoupoker.SCDezhouPokerBankerAndBlindPos); ok {
		logger.Logger.Trace(scBankerAndBlindPos)
		if scene, ok := base.GetScene(s).(*DezhouPokerScene); ok {
			scene.SetBankerAndBlindPos(scBankerAndBlindPos)
		}
	} else {
		logger.Logger.Error("scBankerAndBlindPos package data error.")
	}
	return nil
}

//DZPKPacketID_PACKET_SC_DEZHOUPOKER_CARD
type SCDezhouPokerCardPacketFactory struct {
}

type SCDezhouPokerCardHandler struct {
}

func (this *SCDezhouPokerCardPacketFactory) CreatePacket() interface{} {
	pack := &dezhoupoker.SCDezhouPokerCard{}
	return pack
}

func (this *SCDezhouPokerCardHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	logger.Logger.Trace("(this *SCDezhouPokerCardHandler) Process ", pack)
	if scCard, ok := pack.(*dezhoupoker.SCDezhouPokerCard); ok {

		if scene, ok := base.GetScene(s).(*DezhouPokerScene); ok {
			cardSort := scCard.GetSort()
			switch cardSort {
			case rule.CardType_HandCard:
				p := scene.GetMe(s)
				if p != nil {
					if me, ok := p.(*DezhouPokerPlayer); ok && me != DezhouPokerNilPlayer {
						me.SetCards(scCard.Cards)
					}
				}

			case rule.CardType_FlopCard:
				scene.AddNewCard(rule.CardType_FlopCard, scCard.Cards)
			case rule.CardType_TrunCard:
				scene.AddNewCard(rule.CardType_TrunCard, scCard.Cards)
			case rule.CardType_RiverCard:
				scene.AddNewCard(rule.CardType_RiverCard, scCard.Cards)
			default:
			}
			//test use
			if common.CustomConfig.GetBool("UseDZPWPDebug") {
				if scene.GetState() >= int32(rule.DezhouPokerSceneStateHandCard) {
					if scene.stateFlag&(1<<scene.GetState()) == 0 {
						scene.stateFlag |= (1 << scene.GetState())
						gctx := scene.CreateGameCtx()
						task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
							rule.CalWinningProbability(gctx)
							return gctx
						}), task.CompleteNotifyWrapper(func(data interface{}, tt task.Task) {
							pack := &dezhoupoker.CSCDezhouPokerWPUpdate{}
							for _, pc := range gctx.PlayerCards {
								pack.Datas = append(pack.Datas, &dezhoupoker.DezhouPokerCardsWP{
									SnId:               proto.Int32(pc.UserData.(int32)),
									Cards:              pc.HandCard,
									WinningProbability: proto.Int32(pc.WinningProbability),
								})
							}
							s.Send(int(dezhoupoker.DZPKPacketID_PACKET_CSC_DEZHOUPOKER_WPUPDATE), pack)
						}), "CalWinningProbability").StartByFixExecutor(fmt.Sprintf("dezhouscene_%d", scene.GetRoomId()))
					}
				}
			}
		}

	} else {
		logger.Logger.Error("scBankerAndBlindPos package data error.")
	}
	return nil
}

//SCDezhouPokerPlayerGameCoin
type SCDezhouPokerGameCoinPacketFactory struct {
}

type SCDezhouPokerGameCoinHandler struct {
}

func (this *SCDezhouPokerGameCoinPacketFactory) CreatePacket() interface{} {
	pack := &dezhoupoker.SCDezhouPokerPlayerGameCoin{}
	return pack
}

func (this *SCDezhouPokerGameCoinHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	logger.Logger.Trace("(this *SCDezhouPokerGameCoinHandler) Process ", pack)
	if scPlayerGameCoin, ok := pack.(*dezhoupoker.SCDezhouPokerPlayerGameCoin); ok {
		if scene, ok := base.GetScene(s).(*DezhouPokerScene); ok {
			p := scene.GetMe(s)
			if p != nil && p.GetPos() == scPlayerGameCoin.GetPos() {
				if me, ok := p.(*DezhouPokerPlayer); ok && me != DezhouPokerNilPlayer {
					logger.Logger.Trace("Update GameCoin, OldGameCoin:", me.GetGameCoin(), "new GameCoin", scPlayerGameCoin.GetGameCoin())
					me.SetGameCoin(scPlayerGameCoin.GetGameCoin())
				}
			}
		}
	} else {
		logger.Logger.Error("scPlayerGameCoin package data error.")
	}
	return nil
}

//SCDezhouPokerGameBilled
type SCDezhouPokerGameBilledPacketFactory struct {
}

type SCDezhouPokerGameBilledHandler struct {
}

func (this *SCDezhouPokerGameBilledPacketFactory) CreatePacket() interface{} {
	pack := &dezhoupoker.SCDezhouPokerGameBilled{}
	return pack
}

func (this *SCDezhouPokerGameBilledHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	logger.Logger.Trace("(this *SCDezhouPokerGameBilledHandler) Process ", pack)
	if scGameBilled, ok := pack.(*dezhoupoker.SCDezhouPokerGameBilled); ok {
		if sceneEx, ok := base.GetScene(s).(*DezhouPokerScene); ok {
			sceneEx.OnDezhouPokerNormalBilled(scGameBilled)
		}
	} else {
		logger.Logger.Error("SCDezhouPokerGameBilledHandler package data error.")
	}
	return nil
}

//SCDezhouPokerGameBilledMiddle
type SCDezhouPokerGameBilledMiddlePacketFactory struct {
}

type SCDezhouPokerGameBilledMiddleHandler struct {
}

func (this *SCDezhouPokerGameBilledMiddlePacketFactory) CreatePacket() interface{} {
	pack := &dezhoupoker.SCDezhouPokerGameBilledMiddle{}
	return pack
}

func (this *SCDezhouPokerGameBilledMiddleHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	logger.Logger.Trace("(this *SCDezhouPokerGameBilledMiddleHandler) Process ", pack)
	if scGameBilled, ok := pack.(*dezhoupoker.SCDezhouPokerGameBilledMiddle); ok {
		if sceneEx, ok := base.GetScene(s).(*DezhouPokerScene); ok {
			sceneEx.OnDezhouPokerMiddleBilled(scGameBilled)
		}
	} else {
		logger.Logger.Error("SCDezhouPokerGameBilledMiddleHandler package data error.")
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
func init() {
	//SCDezhouPokerRoomInfo
	netlib.RegisterHandler(int(dezhoupoker.DZPKPacketID_PACKET_SC_DEZHOUPOKER_ROOMINFO), &SCDezhouPokerRoomInfoHandler{})
	netlib.RegisterFactory(int(dezhoupoker.DZPKPacketID_PACKET_SC_DEZHOUPOKER_ROOMINFO), &SCDezhouPokerRoomInfoPacketFactory{})
	//SCDezhouPokerRoomState
	netlib.RegisterHandler(int(dezhoupoker.DZPKPacketID_PACKET_SC_DEZHOUPOKER_ROOMSTATE), &SCDezhouPokerRoomStateHandler{})
	netlib.RegisterFactory(int(dezhoupoker.DZPKPacketID_PACKET_SC_DEZHOUPOKER_ROOMSTATE), &SCDezhouPokerRoomStatePacketFactory{})
	//SCDezhouPokerPlayerOp
	netlib.RegisterHandler(int(dezhoupoker.DZPKPacketID_PACKET_SC_DEZHOUPOKER_OP), &SCDezhouPokerPlayerOpHandler{})
	netlib.RegisterFactory(int(dezhoupoker.DZPKPacketID_PACKET_SC_DEZHOUPOKER_OP), &SCDezhouPokerPlayerOpPacketFactory{})
	//SCDezhouPokerPlayerEnter
	netlib.RegisterHandler(int(dezhoupoker.DZPKPacketID_PACKET_SC_DEZHOUPOKER_PLAYERENTER), &SCDezhouPokerPlayerEnterHandler{})
	netlib.RegisterFactory(int(dezhoupoker.DZPKPacketID_PACKET_SC_DEZHOUPOKER_PLAYERENTER), &SCDezhouPokerPlayerEnterPacketFactory{})
	//SCDezhouPokerPlayerLeave
	netlib.RegisterHandler(int(dezhoupoker.DZPKPacketID_PACKET_SC_DEZHOUPOKER_PLAYERLEAVE), &SCDezhouPokerPlayerLeaveHandler{})
	netlib.RegisterFactory(int(dezhoupoker.DZPKPacketID_PACKET_SC_DEZHOUPOKER_PLAYERLEAVE), &SCDezhouPokerPlayerLeavePacketFactory{})
	//SCDezhouPokerNextOperNotify
	netlib.RegisterHandler(int(dezhoupoker.DZPKPacketID_PACKET_SC_DEZHOUPOKER_NEXTOPERNOTIFY), &SCDezhouPokerNextOperNotifyHandler{})
	netlib.RegisterFactory(int(dezhoupoker.DZPKPacketID_PACKET_SC_DEZHOUPOKER_NEXTOPERNOTIFY), &SCDezhouPokerNextOperNotifyPacketFactory{})
	//SCDezhouPokerBankerAndBlindPos
	netlib.RegisterHandler(int(dezhoupoker.DZPKPacketID_PACKET_SC_DEZHOUPOKER_BANKANDBLINDPOS), &SCDezhouPokerBankerAndBlindPosHandler{})
	netlib.RegisterFactory(int(dezhoupoker.DZPKPacketID_PACKET_SC_DEZHOUPOKER_BANKANDBLINDPOS), &SCDezhouPokerBankerAndBlindPosPacketFactory{})
	//SCDezhouPokerBankerAndBlindPos
	netlib.RegisterHandler(int(dezhoupoker.DZPKPacketID_PACKET_SC_DEZHOUPOKER_CARD), &SCDezhouPokerCardHandler{})
	netlib.RegisterFactory(int(dezhoupoker.DZPKPacketID_PACKET_SC_DEZHOUPOKER_CARD), &SCDezhouPokerCardPacketFactory{})
	//SCDezhouPokerPlayerGameCoin
	netlib.RegisterHandler(int(dezhoupoker.DZPKPacketID_PACKET_SC_DEZHOUPOKER_GAMECOIN), &SCDezhouPokerGameCoinHandler{})
	netlib.RegisterFactory(int(dezhoupoker.DZPKPacketID_PACKET_SC_DEZHOUPOKER_GAMECOIN), &SCDezhouPokerGameCoinPacketFactory{})
	//SCDezhouPokerGameBilled
	netlib.RegisterHandler(int(dezhoupoker.DZPKPacketID_PACKET_SC_DEZHOUPOKER_GAMEBILLED), &SCDezhouPokerGameBilledHandler{})
	netlib.RegisterFactory(int(dezhoupoker.DZPKPacketID_PACKET_SC_DEZHOUPOKER_GAMEBILLED), &SCDezhouPokerGameBilledPacketFactory{})
	//SCDezhouPokerGameBilledMiddle
	netlib.RegisterHandler(int(dezhoupoker.DZPKPacketID_PACKET_SC_DEZHOUPOKER_GAMEBILLED_MIDDLE), &SCDezhouPokerGameBilledMiddleHandler{})
	netlib.RegisterFactory(int(dezhoupoker.DZPKPacketID_PACKET_SC_DEZHOUPOKER_GAMEBILLED_MIDDLE), &SCDezhouPokerGameBilledMiddlePacketFactory{})

}
