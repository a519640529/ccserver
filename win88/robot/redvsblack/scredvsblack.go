package redvsblack

import (
	rule "games.yol.com/win88/gamerule/redvsblack"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/player"
	"games.yol.com/win88/protocol/redvsblack"
	"games.yol.com/win88/robot/base"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
)

type SCRedVsBlackRoomInfoPacketFactory struct {
}

type SCRedVsBlackRoomInfoHandler struct {
}

func (this *SCRedVsBlackRoomInfoPacketFactory) CreatePacket() interface{} {
	pack := &redvsblack.SCRedVsBlackRoomInfo{}
	return pack
}

func (this *SCRedVsBlackRoomInfoHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	logger.Logger.Tracef("(this *SCRedVsBlackRoomInfoHandler) Process [%v].", s.GetSessionConfig().Id)
	if msg, ok := pack.(*redvsblack.SCRedVsBlackRoomInfo); ok {
		var scene *RedVsBlackScene
		if s, exist := base.SceneMgrSington.Scenes[msg.GetRoomId()]; exist {
			scene = s.(*RedVsBlackScene)
		}
		if scene == nil {
			scene := NewRedVsBlackScene(msg)
			base.SceneMgrSington.AddScene(scene)
		}

		if scene != nil {
			var selfSnid int32
			if user, ok := s.GetAttribute(base.SessionAttributeUser).(*player.SCPlayerData); ok {
				selfSnid = user.GetData().GetSnId()
			}

			for _, pd := range msg.GetPlayers() {
				if pd.GetSnId() == selfSnid {
					if scene.GetPlayerBySnid(pd.GetSnId()) == nil {
						p := NewRedVsBlackPlayer(pd)
						if p != nil {
							scene.AddPlayer(p)
							p.Scene = scene
							p.TreeID = scene.RandPlayerType()
							p.TakeCoin = p.GetCoin()
							p.Trend20 = msg.GetTrend20Lately()[:]
						}
					} else {
						p := scene.GetPlayerBySnid(pd.GetSnId())
						if mp, ok := p.(*RedVsBlackPlayer); ok {
							mp.TakeCoin = p.GetCoin()
							mp.Trend20 = msg.GetTrend20Lately()[:]
						}
					}
				}
			}
			//logger.Logger.Trace(msg)
			s.SetAttribute(base.SessionAttributeScene, scene)
			switch int(msg.GetState()) {
			case rule.RedVsBlackSceneStateStake:
				player := scene.GetMe(s)
				if player != base.NilPlayer {
					if me, ok := player.(*RedVsBlackPlayer); ok {
						scene.Action(s, me)
					}
				}
			case rule.RedVsBlackSceneStateBilled:
				scene.Clear()
				player := scene.GetMe(s)
				if player != base.NilPlayer {
					if me, ok := player.(*RedVsBlackPlayer); ok {
						if me != RedVsBlackNilPlayer {
							me.Clear()

						}
					}
				}
			}
		}
	} else {
		logger.Logger.Error("SCRedVsBlackRoomInfo package data error.")
	}
	return nil
}

type SCRedVsBlackOpPacketFactory struct {
}

type SCRedVsBlackOpHandler struct {
}

func (this *SCRedVsBlackOpPacketFactory) CreatePacket() interface{} {
	pack := &redvsblack.SCRedVsBlackOp{}
	return pack
}

func (this *SCRedVsBlackOpHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	//logger.Logger.Tracef("(this *SCRedVsBlackOpHandler) Process [%v].", s.GetSessionConfig().Id)
	if scRedVsBlackOp, ok := pack.(*redvsblack.SCRedVsBlackOp); ok {
		if scene, ok := s.GetAttribute(base.SessionAttributeScene).(*RedVsBlackScene); ok && scene != nil {
			var me *RedVsBlackPlayer
			player := scene.GetMe(s)
			if player != RedVsBlackNilPlayer && player != base.NilPlayer {
				me = player.(*RedVsBlackPlayer)
			}
			params := scRedVsBlackOp.GetParams()
			if me != RedVsBlackNilPlayer {
				if int(scRedVsBlackOp.GetOpRetCode()) == 0 {
					switch int(scRedVsBlackOp.GetOpCode()) {
					case rule.RedVsBlackPlayerOpBet:
						if len(params) >= 4 && params[0] >= 0 && int(params[0]) < RVSB_ZONE_MAX {
							scene.totalBet[params[0]] = params[2]
						}
					}
					if scRedVsBlackOp.GetSnId() == me.GetSnId() {
						if len(params) >= 4 {
							me.Coin = proto.Int64(int64(params[3]))
							me.bets[params[0]] += params[1]
							me.totalBet += params[1]
						}
					}
				}
			}
		}
	} else {
		logger.Logger.Error("SCRedVsBlackOp package data error.")
	}
	return nil
}

type SCRedVsBlackRoomStatePacketFactory struct {
}

type SCRedVsBlackRoomStateHandler struct {
}

func (this *SCRedVsBlackRoomStatePacketFactory) CreatePacket() interface{} {
	pack := &redvsblack.SCRedVsBlackRoomState{}
	return pack
}

func (this *SCRedVsBlackRoomStateHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	logger.Logger.Tracef("(this *SCRedVsBlackRoomStateHandler) Process [%v].", s.GetSessionConfig().Id)
	if scRedVsBlackRoomState, ok := pack.(*redvsblack.SCRedVsBlackRoomState); ok {
		if scene, ok := s.GetAttribute(base.SessionAttributeScene).(*RedVsBlackScene); ok {
			scene.State = scRedVsBlackRoomState.State
			p := scene.GetMe(s)
			if me, ok := p.(*RedVsBlackPlayer); ok && me != RedVsBlackNilPlayer {
				//logger.Logger.Trace(scRedVsBlackRoomState)
				switch int(scRedVsBlackRoomState.GetState()) {
				case rule.RedVsBlackSceneStateOpenCard:
					me.Trend20 = append(me.Trend20, scRedVsBlackRoomState.GetParams()[8])
					if len(me.Trend20) > 20 {
						me.Trend20 = me.Trend20[len(me.Trend20)-20:]
					}

				case rule.RedVsBlackSceneStateStake:
					scene.Action(s, me)
				case rule.RedVsBlackSceneStateBilled:
					scene.Clear()
					me.Clear()
					me.GameCount += 1
				}
			}
		}
	} else {
		logger.Logger.Error("SCRedVsBlackRoomState package data error.")
	}
	return nil
}

type SCRedVsBlackGameBilledPacketFactory struct {
}

type SCRedVsBlackGameBilledHandler struct {
}

func (this *SCRedVsBlackGameBilledPacketFactory) CreatePacket() interface{} {
	pack := &redvsblack.SCRedVsBlackBilled{}
	return pack
}

func (this *SCRedVsBlackGameBilledHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	logger.Logger.Tracef("(this *SCRedVsBlackGameBilledHandler) Process [%v].", s.GetSessionConfig().Id)
	if scRedVsBlackBilled, ok := pack.(*redvsblack.SCRedVsBlackBilled); ok {
		if scene, ok := s.GetAttribute(base.SessionAttributeScene).(*RedVsBlackScene); ok {
			//logger.Logger.Trace(scRedVsBlackBilled)
			billData := scRedVsBlackBilled.GetBillData()
			for _, data := range billData {
				p := scene.GetMe(s)
				if me, ok := p.(*RedVsBlackPlayer); ok && me != RedVsBlackNilPlayer {
					if data.GetSnId() == me.GetSnId() { //自己的数据
						me.Coin = proto.Int64(data.GetCoin())
						if data.GetGainCoin() > 0 {
							me.LastWinOrLoss = 1
						} else if data.GetGainCoin() < 0 {
							me.LastWinOrLoss = -1
						} else {
							me.LastWinOrLoss = 0
						}
					}
				}
			}
		}
	} else {
		logger.Logger.Error("SCRedVsBlackGameBilled package data error.")
	}
	return nil
}

func init() {
	//SCRedVsBlackRoomInfo
	netlib.RegisterHandler(int(redvsblack.RedVsBlackPacketID_PACKET_SC_RVSB_ROOMINFO), &SCRedVsBlackRoomInfoHandler{})
	netlib.RegisterFactory(int(redvsblack.RedVsBlackPacketID_PACKET_SC_RVSB_ROOMINFO), &SCRedVsBlackRoomInfoPacketFactory{})
	//SCRedVsBlackOp
	netlib.RegisterHandler(int(redvsblack.RedVsBlackPacketID_PACKET_SC_RVSB_PLAYEROP), &SCRedVsBlackOpHandler{})
	netlib.RegisterFactory(int(redvsblack.RedVsBlackPacketID_PACKET_SC_RVSB_PLAYEROP), &SCRedVsBlackOpPacketFactory{})
	//SCRedVsBlackRoomState
	netlib.RegisterHandler(int(redvsblack.RedVsBlackPacketID_PACKET_SC_RVSB_ROOMSTATE), &SCRedVsBlackRoomStateHandler{})
	netlib.RegisterFactory(int(redvsblack.RedVsBlackPacketID_PACKET_SC_RVSB_ROOMSTATE), &SCRedVsBlackRoomStatePacketFactory{})
	//SCRedVsBlackFinalBilled
	netlib.RegisterHandler(int(redvsblack.RedVsBlackPacketID_PACKET_SC_RVSB_GAMEBILLED), &SCRedVsBlackGameBilledHandler{})
	netlib.RegisterFactory(int(redvsblack.RedVsBlackPacketID_PACKET_SC_RVSB_GAMEBILLED), &SCRedVsBlackGameBilledPacketFactory{})
}
