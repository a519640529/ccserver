package crash

import (
	rule "games.yol.com/win88/gamerule/crash"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/player"
	"games.yol.com/win88/protocol/crash"
	"games.yol.com/win88/robot/base"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
)

type SCCrashRoomInfoPacketFactory struct {
}

type SCCrashRoomInfoHandler struct {
}

func (this *SCCrashRoomInfoPacketFactory) CreatePacket() interface{} {
	pack := &crash.SCCrashRoomInfo{}
	return pack
}

func (this *SCCrashRoomInfoHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	logger.Logger.Tracef("(this *SCCrashRoomInfoHandler) Process [%v].", s.GetSessionConfig().Id)
	if msg, ok := pack.(*crash.SCCrashRoomInfo); ok {
		var scene *CrashScene
		if s, exist := base.SceneMgrSington.Scenes[msg.GetRoomId()]; exist {
			scene = s.(*CrashScene)
		}
		if scene == nil {
			scene := NewCrashScene(msg)
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
						p := NewCrashPlayer(pd)
						if p != nil {
							scene.AddPlayer(p)
							p.Scene = scene
							p.TreeID = scene.RandPlayerType()
							p.TakeCoin = p.GetCoin()
							p.Trend20 = msg.GetTrend20Lately()[:]
						}
					} else {
						p := scene.GetPlayerBySnid(pd.GetSnId())
						if mp, ok := p.(*CrashPlayer); ok {
							mp.TakeCoin = p.GetCoin()
							mp.Trend20 = msg.GetTrend20Lately()[:]
						}
					}
				}
			}
			//logger.Logger.Trace(msg)
			s.SetAttribute(base.SessionAttributeScene, scene)
			switch int(msg.GetState()) {
			case rule.CrashSceneStateStake:
				player := scene.GetMe(s)
				if player != base.NilPlayer {
					if me, ok := player.(*CrashPlayer); ok {
						scene.Action(s, me)
					}
				}
			case rule.CrashSceneStateBilled:
				scene.Clear()
				player := scene.GetMe(s)
				if player != base.NilPlayer {
					if me, ok := player.(*CrashPlayer); ok {
						if me != CrashNilPlayer {
							me.Clear()

						}
					}
				}
			}
		}
	} else {
		logger.Logger.Error("SCCrashRoomInfo package data error.")
	}
	return nil
}

type SCCrashOpPacketFactory struct {
}

type SCCrashOpHandler struct {
}

func (this *SCCrashOpPacketFactory) CreatePacket() interface{} {
	pack := &crash.SCCrashOp{}
	return pack
}

func (this *SCCrashOpHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	//logger.Logger.Tracef("(this *SCCrashOpHandler) Process [%v].", s.GetSessionConfig().Id)
	if scCrashOp, ok := pack.(*crash.SCCrashOp); ok {
		if scene, ok := s.GetAttribute(base.SessionAttributeScene).(*CrashScene); ok && scene != nil {
			var me *CrashPlayer
			player := scene.GetMe(s)
			if player != CrashNilPlayer && player != base.NilPlayer {
				me = player.(*CrashPlayer)
			}
			params := scCrashOp.GetParams()
			if me != CrashNilPlayer {
				if int(scCrashOp.GetOpRetCode()) == 0 {
					switch int(scCrashOp.GetOpCode()) {
					case rule.CrashPlayerOpBet:
						//if len(params) >= 4 && params[0] >= 0 && int(params[0]) < CRASH_ZONE_MAX {
						//	scene.totalBet[params[0]] = params[2]
						//}
					}
					if scCrashOp.GetSnId() == me.GetSnId() {
						if len(params) >= 3 {
							me.Coin = proto.Int64(int64(params[2]))
							//me.bets[params[0]] += params[1]
							me.multiple = int32(params[0])
							me.betTotal = params[1]
							me.totalBet += params[1]
						}
					}
				}
			}
		}
	} else {
		logger.Logger.Error("SCCrashOp package data error.")
	}
	return nil
}

type SCCrashRoomStatePacketFactory struct {
}

type SCCrashRoomStateHandler struct {
}

func (this *SCCrashRoomStatePacketFactory) CreatePacket() interface{} {
	pack := &crash.SCCrashRoomState{}
	return pack
}

func (this *SCCrashRoomStateHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	logger.Logger.Tracef("(this *SCCrashRoomStateHandler) Process [%v].", s.GetSessionConfig().Id)
	if scCrashRoomState, ok := pack.(*crash.SCCrashRoomState); ok {
		if scene, ok := s.GetAttribute(base.SessionAttributeScene).(*CrashScene); ok {
			scene.State = scCrashRoomState.State
			p := scene.GetMe(s)
			if me, ok := p.(*CrashPlayer); ok && me != CrashNilPlayer {
				//logger.Logger.Trace(scCrashRoomState)
				switch int(scCrashRoomState.GetState()) {
				case rule.CrashSceneStateOpenCard:
					//me.Trend20 = append(me.Trend20, scCrashRoomState.GetParams()[8])
					//if len(me.Trend20) > 20 {
					//	me.Trend20 = me.Trend20[len(me.Trend20)-20:]
					//}

				case rule.CrashSceneStateStake:
					scene.Action(s, me)
				case rule.CrashSceneStateBilled:
					scene.Clear()
					me.Clear()
					me.GameCount += 1
				}
			}
		}
	} else {
		logger.Logger.Error("SCCrashRoomState package data error.")
	}
	return nil
}

type SCCrashGameBilledPacketFactory struct {
}

type SCCrashGameBilledHandler struct {
}

func (this *SCCrashGameBilledPacketFactory) CreatePacket() interface{} {
	pack := &crash.SCCrashBilled{}
	return pack
}

func (this *SCCrashGameBilledHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	logger.Logger.Tracef("(this *SCCrashGameBilledHandler) Process [%v].", s.GetSessionConfig().Id)
	if scCrashBilled, ok := pack.(*crash.SCCrashBilled); ok {
		if scene, ok := s.GetAttribute(base.SessionAttributeScene).(*CrashScene); ok {
			//logger.Logger.Trace(scCrashBilled)
			billData := scCrashBilled.GetBillData()
			for _, data := range billData {
				p := scene.GetMe(s)
				if me, ok := p.(*CrashPlayer); ok && me != CrashNilPlayer {
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
		logger.Logger.Error("SCCrashGameBilled package data error.")
	}
	return nil
}

func init() {
	//SCCrashRoomInfo
	netlib.RegisterHandler(int(crash.CrashPacketID_PACKET_SC_CRASH_ROOMINFO), &SCCrashRoomInfoHandler{})
	netlib.RegisterFactory(int(crash.CrashPacketID_PACKET_SC_CRASH_ROOMINFO), &SCCrashRoomInfoPacketFactory{})
	//SCCrashOp
	netlib.RegisterHandler(int(crash.CrashPacketID_PACKET_SC_CRASH_PLAYEROP), &SCCrashOpHandler{})
	netlib.RegisterFactory(int(crash.CrashPacketID_PACKET_SC_CRASH_PLAYEROP), &SCCrashOpPacketFactory{})
	//SCCrashRoomState
	netlib.RegisterHandler(int(crash.CrashPacketID_PACKET_SC_CRASH_ROOMSTATE), &SCCrashRoomStateHandler{})
	netlib.RegisterFactory(int(crash.CrashPacketID_PACKET_SC_CRASH_ROOMSTATE), &SCCrashRoomStatePacketFactory{})
	//SCCrashFinalBilled
	netlib.RegisterHandler(int(crash.CrashPacketID_PACKET_SC_CRASH_GAMEBILLED), &SCCrashGameBilledHandler{})
	netlib.RegisterFactory(int(crash.CrashPacketID_PACKET_SC_CRASH_GAMEBILLED), &SCCrashGameBilledPacketFactory{})
}
