package dragonvstiger

import (
	. "games.yol.com/win88/gamerule/dragonvstiger"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/dragonvstiger"
	"games.yol.com/win88/protocol/player"
	"games.yol.com/win88/robot/base"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"math/rand"
)

type SCDragonVsTigerRoomInfoPacketFactory struct {
}

type SCDragonVsTigerRoomInfoHandler struct {
}

func (this *SCDragonVsTigerRoomInfoPacketFactory) CreatePacket() interface{} {
	pack := &dragonvstiger.SCDragonVsTigerRoomInfo{}
	return pack
}

func (this *SCDragonVsTigerRoomInfoHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	logger.Logger.Tracef("(this *SCDragonVsTigerRoomInfoHandler) Process [%v].", s.GetSessionConfig().Id)
	if msg, ok := pack.(*dragonvstiger.SCDragonVsTigerRoomInfo); ok {
		var scene *DragonVsTigerScene
		if s, exist := base.SceneMgrSington.Scenes[msg.GetRoomId()]; exist {
			scene = s.(*DragonVsTigerScene)
		}
		if scene == nil {
			scene := NewDragonVsTigerScene(msg)
			params := scene.GetParamsEx()
			if len(params) != 0 {
				scene.dbGameFree = base.SceneMgrSington.GetSceneDBGameFree(scene.GetRoomId(), params[0])
				scene.isRandNum[scene.dbGameFree.GetId()] = rand.Int31n(3) + 1
			}
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
						p := NewDragonVsTigerPlayer(pd)
						if p != nil {
							scene.AddPlayer(p)
							p.Scene = scene
							p.TreeID = scene.RandPlayerType()
							p.TakeCoin = p.GetCoin()
							p.Trend20 = msg.GetTrend20Lately()[:]
						}
					} else {
						p := scene.GetPlayerBySnid(pd.GetSnId())
						if mp, ok := p.(*DragonVsTigerPlayer); ok {
							mp.TakeCoin = p.GetCoin()
							mp.Trend20 = msg.GetTrend20Lately()[:]
						}
					}
				}
			}

			//logger.Logger.Trace(msg)
			s.SetAttribute(base.SessionAttributeScene, scene)
			switch int(msg.GetState()) {
			case DragonVsTigerSceneStateStake:
				me := scene.GetMe(s)
				if me != nil {
					scene.Action(s, me.(*DragonVsTigerPlayer))
				}
			case DragonVsTigerSceneStateBilled:
				scene.Clear()
				me := scene.GetMe(s).(*DragonVsTigerPlayer)
				if me != DragonVsTigerNilPlayer {
					me.Clear()
				}
			}
		}
	} else {
		logger.Logger.Error("SCDragonVsTigerRoomInfo package data error.")
	}
	return nil
}

type SCDragonVsTigerPlayerOpPacketFactory struct {
}

type SCDragonVsTigerPlayerOpHandler struct {
}

func (this *SCDragonVsTigerPlayerOpPacketFactory) CreatePacket() interface{} {
	pack := &dragonvstiger.SCDragonVsTiggerOp{}
	return pack
}

func (this *SCDragonVsTigerPlayerOpHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	//logger.Logger.Tracef("(this *SCDragonVsTigerPlayerOpHandler) Process [%v].", s.GetSessionConfig().Id)
	if scDragonVsTigerOp, ok := pack.(*dragonvstiger.SCDragonVsTiggerOp); ok {
		if scene, ok := s.GetAttribute(base.SessionAttributeScene).(*DragonVsTigerScene); ok {
			player := scene.GetMe(s)
			if player == base.NilPlayer {
				return nil
			}
			params := scDragonVsTigerOp.GetParams()
			if me, ok := player.(*DragonVsTigerPlayer); ok && me != DragonVsTigerNilPlayer {
				if int(scDragonVsTigerOp.GetOpRetCode()) == 0 {
					switch int(scDragonVsTigerOp.GetOpCode()) {
					case DragonVsTigerPlayerOpBet:
						if len(params) >= 4 && params[0] >= 0 && int(params[0]) < DVST_ZONE_MAX {
							scene.totalBet[params[0]] = int64(params[2])
						}
					}
					if scDragonVsTigerOp.GetSnId() == me.GetSnId() {
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
		logger.Logger.Error("SCDragonVsTigerPlayerOp package data error.")
	}
	return nil
}

type SCDragonVsTigerRoomStatePacketFactory struct {
}

type SCDragonVsTigerRoomStateHandler struct {
}

func (this *SCDragonVsTigerRoomStatePacketFactory) CreatePacket() interface{} {
	pack := &dragonvstiger.SCDragonVsTigerRoomState{}
	return pack
}

func (this *SCDragonVsTigerRoomStateHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	logger.Logger.Tracef("(this *SCDragonVsTigerRoomStateHandler) Process [%v].", s.GetSessionConfig().Id)
	if scDragonVsTigerRoomState, ok := pack.(*dragonvstiger.SCDragonVsTigerRoomState); ok {
		if scene, ok := s.GetAttribute(base.SessionAttributeScene).(*DragonVsTigerScene); ok {
			scene.State = scDragonVsTigerRoomState.State
			p := scene.GetMe(s)
			if p == base.NilPlayer {
				return nil
			}
			if me, ok := p.(*DragonVsTigerPlayer); ok && me != DragonVsTigerNilPlayer {

				//logger.Logger.Trace(scDragonVsTigerRoomState)
				switch int(scDragonVsTigerRoomState.GetState()) {
				case DragonVsTigerSceneStateOpenCard:
					me.Trend20 = append(me.Trend20, scDragonVsTigerRoomState.GetParams()[2])
					if len(me.Trend20) > 20 {
						me.Trend20 = me.Trend20[len(me.Trend20)-20:]
					}
				case DragonVsTigerSceneStateStake:
					scene.Action(s, me)
				case DragonVsTigerSceneStateBilled:
					scene.Clear()
					me.Clear()
					me.GameCount += 1
					scene.UpBanker(me)
				}
			}
		}
	} else {
		logger.Logger.Error("SCDragonVsTigerRoomState package data error.")
	}
	return nil
}

type SCDragonVsTigerGameBilledPacketFactory struct {
}

type SCDragonVsTigerGameBilledHandler struct {
}

func (this *SCDragonVsTigerGameBilledPacketFactory) CreatePacket() interface{} {
	pack := &dragonvstiger.SCDragonVsTigerBilled{}
	return pack
}

func (this *SCDragonVsTigerGameBilledHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	logger.Logger.Tracef("(this *SCDragonVsTigerGameBilledHandler) Process [%v].", s.GetSessionConfig().Id)
	if scDragonVsTigerBilled, ok := pack.(*dragonvstiger.SCDragonVsTigerBilled); ok {
		if scene, ok := s.GetAttribute(base.SessionAttributeScene).(*DragonVsTigerScene); ok {
			//logger.Logger.Trace(scDragonVsTigerBilled)
			billData := scDragonVsTigerBilled.GetBillData()
			for _, data := range billData {
				p := scene.GetMe(s)
				if p == base.NilPlayer {
					continue
				}
				if me, ok := p.(*DragonVsTigerPlayer); ok && me != DragonVsTigerNilPlayer {
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
		logger.Logger.Error("SCDragonVsTigerGameBilled package data error.")
	}
	return nil
}

type SCDragonVsTiggerUpListPacketFactory struct {
}

type SCDragonVsTiggerUpListHandler struct {
}

func (this *SCDragonVsTiggerUpListPacketFactory) CreatePacket() interface{} {
	pack := &dragonvstiger.SCDragonVsTiggerUpList{}
	return pack
}

func (this *SCDragonVsTiggerUpListHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	logger.Logger.Tracef("(this *SCDragonVsTiggerUpListHandler) Process [%v].", s.GetSessionConfig().Id)
	if msg, ok := pack.(*dragonvstiger.SCDragonVsTiggerUpList); ok {
		if scene, ok := s.GetAttribute(base.SessionAttributeScene).(*DragonVsTigerScene); ok {
			scene.upBankerListNum = msg.GetCount()
		}
	} else {
		logger.Logger.Error("SCDragonVsTiggerUpList package data error.")
	}
	return nil
}
func init() {
	//SCDragonVsTigerRoomInfo
	netlib.RegisterHandler(int(dragonvstiger.DragonVsTigerPacketID_PACKET_SC_DVST_ROOMINFO), &SCDragonVsTigerRoomInfoHandler{})
	netlib.RegisterFactory(int(dragonvstiger.DragonVsTigerPacketID_PACKET_SC_DVST_ROOMINFO), &SCDragonVsTigerRoomInfoPacketFactory{})
	//SCDragonVsTigerPlayerOp
	netlib.RegisterHandler(int(dragonvstiger.DragonVsTigerPacketID_PACKET_SC_DVST_PLAYEROP), &SCDragonVsTigerPlayerOpHandler{})
	netlib.RegisterFactory(int(dragonvstiger.DragonVsTigerPacketID_PACKET_SC_DVST_PLAYEROP), &SCDragonVsTigerPlayerOpPacketFactory{})
	//SCDragonVsTigerRoomState
	netlib.RegisterHandler(int(dragonvstiger.DragonVsTigerPacketID_PACKET_SC_DVST_ROOMSTATE), &SCDragonVsTigerRoomStateHandler{})
	netlib.RegisterFactory(int(dragonvstiger.DragonVsTigerPacketID_PACKET_SC_DVST_ROOMSTATE), &SCDragonVsTigerRoomStatePacketFactory{})
	//SCDragonVsTigerFinalBilled
	netlib.RegisterHandler(int(dragonvstiger.DragonVsTigerPacketID_PACKET_SC_DVST_GAMEBILLED), &SCDragonVsTigerGameBilledHandler{})
	netlib.RegisterFactory(int(dragonvstiger.DragonVsTigerPacketID_PACKET_SC_DVST_GAMEBILLED), &SCDragonVsTigerGameBilledPacketFactory{})
	//SCBaccaratBankerList
	netlib.RegisterHandler(int(dragonvstiger.DragonVsTigerPacketID_PACKET_SC_DVST_UPLIST), &SCDragonVsTiggerUpListHandler{})
	netlib.RegisterFactory(int(dragonvstiger.DragonVsTigerPacketID_PACKET_SC_DVST_UPLIST), &SCDragonVsTiggerUpListPacketFactory{})
}
