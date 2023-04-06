package roulette

import (
	rule "games.yol.com/win88/gamerule/roulette"
	"games.yol.com/win88/proto"
	proto_roulette "games.yol.com/win88/protocol/roulette"
	"games.yol.com/win88/robot/base"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"math/rand"
	"time"
)

type SCRouletteRoomInfoPacketFactory struct {
}

type SCRouletteRoomInfoHandler struct {
}

func (this *SCRouletteRoomInfoPacketFactory) CreatePacket() interface{} {
	pack := &proto_roulette.SCRouletteRoomInfo{}
	return pack
}

func (this *SCRouletteRoomInfoHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	logger.Logger.Tracef("(this *SCRouletteRoomInfoHandler) Process [%v].", s.GetSessionConfig().Id)
	if msg, ok := pack.(*proto_roulette.SCRouletteRoomInfo); ok {
		scene := NewRouletteScene(msg)
		if scene != nil {
			for _, pd := range msg.GetPlayers() {
				if scene.GetPlayerBySnid(pd.GetSnId()) == nil {
					p := NewRoulettePlayer(pd)
					if p != nil {
						scene.AddPlayer(p)
					}
				}
			}
			s.SetAttribute(base.SessionAttributeScene, scene)
		}
	} else {
		logger.Logger.Error("SCRouletteRoomInfo package data error.")
	}
	return nil
}

type SCRouletteRoomStatePacketFactory struct {
}

type SCRouletteRoomStateHandler struct {
}

func (this *SCRouletteRoomStatePacketFactory) CreatePacket() interface{} {
	pack := &proto_roulette.SCRouletteRoomState{}
	return pack
}

func (this *SCRouletteRoomStateHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	logger.Logger.Tracef("(this *SCRouletteRoomStateHandler) Process [%v].", s.GetSessionConfig().Id)
	if msg, ok := pack.(*proto_roulette.SCRouletteRoomState); ok {
		if scene, ok := s.GetAttribute(base.SessionAttributeScene).(*RouletteScene); ok {
			scene.State = msg.State
			//scene.State = proto.Int32(msg.GetState())
			p := scene.GetMe(s)
			if p == base.NilPlayer {
				return nil
			}
			if me, ok := p.(*RoulettePlayer); ok && me != RouletteNilPlayer {
				if msg.GetState() == int32(rule.RouletteSceneStateBet) {
					//下注
					if me.GetPos() < rule.Roulette_SELFPOS {
						//上座玩家 能续压直接续压
						//if me.lastBetCoin > 0 && me.GetCoin() >= me.lastBetCoin {
						//	pack := &proto_roulette.CSRoulettePlayerOp{
						//		OpCode:  proto.Int32(int32(RoulettePlayerOpBet)),
						//		OpParam: []int64{int64(RoulettePlayerOpProceedBet)},
						//	}
						//	proto.SetDefaults(pack)
						//	s.Send(int(proto_roulette.RouletteMmoPacketID_PACKET_CS_Roulette_PlayerOp), pack)
						//} else {
						//不能续压
						me.betTime = time.Duration(rand.Intn(1000) + 1000)
						scene.Action(s, me, time.Now(), true)
						//}
					} else {
						me.betTime = time.Duration(rand.Intn(5000) + 8000) //确定该下注时间
						scene.Action(s, me, time.Now(), false)
					}
				} else if msg.GetState() == int32(rule.RouletteSceneStateBilled) {
					//结算
					scene.Clear()
				}
			}
		}
	} else {
		logger.Logger.Error("SCRouletteRoomState package data error.")
	}
	return nil
}

type SCRoulettePlayerOpPacketFactory struct {
}

type SCRoulettePlayerOpHandler struct {
}

func (this *SCRoulettePlayerOpPacketFactory) CreatePacket() interface{} {
	pack := &proto_roulette.SCRoulettePlayerOp{}
	return pack
}

func (this *SCRoulettePlayerOpHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	logger.Logger.Tracef("(this *SCRoulettePlayerOpHandler) Process [%v].", s.GetSessionConfig().Id)
	if msg, ok := pack.(*proto_roulette.SCRoulettePlayerOp); ok {
		if scene, ok := s.GetAttribute(base.SessionAttributeScene).(*RouletteScene); ok {
			player := scene.GetMe(s)
			if player == base.NilPlayer {
				return nil
			}
			if msg.GetOpRCode() == 0 {
				if msg.GetOpCode() == int32(rule.RoulettePlayerOpBet) {
					if me, ok := player.(*RoulettePlayer); ok && me != RouletteNilPlayer {
						betCoin := msg.GetProceedBetCoin()
						dbGameFree := base.SceneMgrSington.GetSceneDBGameFree(scene.GetRoomId(), scene.GetGameFreeId())
						if dbGameFree != nil {
							coin := int64(0)
							for _, cnts := range betCoin {
								for idx, cnt := range cnts.GetCnt() {
									coin += int64(dbGameFree.GetOtherIntParams()[idx] * cnt)
								}
							}
							me.BetCoin = proto.Int64(me.GetBetCoin() + coin)
						}
					}
				} else if msg.GetOpCode() == int32(rule.RoulettePlayerOpProceedBet) {
					if me, ok := player.(*RoulettePlayer); ok && me != RouletteNilPlayer {
						me.BetCoin = proto.Int64(me.lastBetCoin)
					}
				}

			}
		}
	} else {
		logger.Logger.Error("SCRoulettePlayerOpHandler package data error.")
	}
	return nil
}

type SCRouletteTopPlayerPacketFactory struct {
}

type SCRouletteTopPlayerHandler struct {
}

func (this *SCRouletteTopPlayerPacketFactory) CreatePacket() interface{} {
	pack := &proto_roulette.SCRouletteTopPlayer{}
	return pack
}

func (this *SCRouletteTopPlayerHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	logger.Logger.Tracef("(this *SCRouletteTopPlayerHandler) Process [%v].", s.GetSessionConfig().Id)
	if msg, ok := pack.(*proto_roulette.SCRouletteTopPlayer); ok {
		if scene, ok := s.GetAttribute(base.SessionAttributeScene).(*RouletteScene); ok {
			player := scene.GetMe(s)
			if player == base.NilPlayer {
				return nil
			}
			if me, ok := player.(*RoulettePlayer); ok && me != RouletteNilPlayer {
				for _, v := range msg.GetPlayers() {
					if v.GetSnId() == me.GetSnId() {
						me.Pos = proto.Int32(v.GetPos())
						break
					}
				}
			}
		}
	} else {
		logger.Logger.Error("SCRouletteTopPlayerHandler package data error.")
	}
	return nil
}

type SCRouletteBilledPacketFactory struct {
}

type SCRouletteBilledHandler struct {
}

func (this *SCRouletteBilledPacketFactory) CreatePacket() interface{} {
	pack := &proto_roulette.SCRouletteBilled{}
	return pack
}

func (this *SCRouletteBilledHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	logger.Logger.Tracef("(this *SCRouletteBilledHandler) Process [%v].", s.GetSessionConfig().Id)
	if msg, ok := pack.(*proto_roulette.SCRouletteBilled); ok {
		if scene, ok := s.GetAttribute(base.SessionAttributeScene).(*RouletteScene); ok {
			player := scene.GetMe(s)
			if player == base.NilPlayer {
				return nil
			}
			if me, ok := player.(*RoulettePlayer); ok && me != RouletteNilPlayer {
				for _, v := range msg.GetPlayers() {
					if v.GetPos() == rule.Roulette_SELFPOS {
						me.Coin = v.LastCoin
						break
					}
				}
			}
		}
	} else {
		logger.Logger.Error("SCRouletteBilledHandler package data error.")
	}
	return nil
}
func init() {
	//SCRouletteRoomInfo
	netlib.RegisterHandler(int(proto_roulette.RouletteMmoPacketID_PACKET_SC_Roulette_RoomInfo), &SCRouletteRoomInfoHandler{})
	netlib.RegisterFactory(int(proto_roulette.RouletteMmoPacketID_PACKET_SC_Roulette_RoomInfo), &SCRouletteRoomInfoPacketFactory{})
	//SCRouletteRoomState
	netlib.RegisterHandler(int(proto_roulette.RouletteMmoPacketID_PACKET_SC_Roulette_RoomState), &SCRouletteRoomStateHandler{})
	netlib.RegisterFactory(int(proto_roulette.RouletteMmoPacketID_PACKET_SC_Roulette_RoomState), &SCRouletteRoomStatePacketFactory{})
	//SCRoulettePlayerOp
	netlib.RegisterHandler(int(proto_roulette.RouletteMmoPacketID_PACKET_SC_Roulette_PlayerOp), &SCRoulettePlayerOpHandler{})
	netlib.RegisterFactory(int(proto_roulette.RouletteMmoPacketID_PACKET_SC_Roulette_PlayerOp), &SCRoulettePlayerOpPacketFactory{})
	//SCRouletteTopPlayer
	netlib.RegisterHandler(int(proto_roulette.RouletteMmoPacketID_PACKET_SC_Roulette_TopPlayer), &SCRouletteTopPlayerHandler{})
	netlib.RegisterFactory(int(proto_roulette.RouletteMmoPacketID_PACKET_SC_Roulette_TopPlayer), &SCRouletteTopPlayerPacketFactory{})
	//SCRouletteBilled
	netlib.RegisterHandler(int(proto_roulette.RouletteMmoPacketID_PACKET_SC_Roulette_Billed), &SCRouletteBilledHandler{})
	netlib.RegisterFactory(int(proto_roulette.RouletteMmoPacketID_PACKET_SC_Roulette_Billed), &SCRouletteBilledPacketFactory{})
}
