package baccarat

import (
	"games.yol.com/win88/gamerule/baccarat"
	"games.yol.com/win88/proto"
	proto_baccarat "games.yol.com/win88/protocol/baccarat"
	"games.yol.com/win88/robot/base"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"math/rand"
)

type SCBaccaratRoomInfoPacketFactory struct {
}

type SCBaccaratRoomInfoHandler struct {
}

func (this *SCBaccaratRoomInfoPacketFactory) CreatePacket() interface{} {
	pack := &proto_baccarat.SCBaccaratRoomInfo{}
	return pack
}

func (this *SCBaccaratRoomInfoHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	logger.Logger.Tracef("(this *SCBaccaratRoomInfoHandler) Process [%v].", s.GetSessionConfig().Id)
	if msg, ok := pack.(*proto_baccarat.SCBaccaratRoomInfo); ok {
		var scene *BaccaratScene
		ss := base.SceneMgrSington.GetScene(msg.GetRoomId())
		if ss != nil {
			scene = ss.(*BaccaratScene)
		}
		if scene == nil {
			scene := NewBaccaratScene(msg)
			params := scene.GetParamsEx()
			if len(params) != 0 {
				scene.dbGameFree = base.SceneMgrSington.GetSceneDBGameFree(scene.GetRoomId(), params[0])
				scene.isRandNum[scene.dbGameFree.GetId()] = rand.Int31n(3) + 1
			}
			for _, pd := range msg.GetPlayers() {
				if scene.GetPlayerBySnid(pd.GetSnId()) == nil {
					p := NewBaccaratPlayer(pd)
					if p != nil {
						scene.AddPlayer(p)
					}
				}
			}
			//logger.Logger.Trace(msg)
			s.SetAttribute(base.SessionAttributeScene, scene)
			switch msg.GetState() {
			case int32(baccarat.BaccaratSceneStateStake):
				me := scene.GetMe(s)
				if me != nil {
					scene.Action(s, me.(*BaccaratPlayer))
				}
			case int32(baccarat.BaccaratSceneStateBilled):
				scene.Clear()
				me := scene.GetMe(s)
				if me != nil {
					if bpm, ok := me.(*BaccaratPlayer); ok {
						if bpm != BaccaratNilPlayer {
							bpm.Clear()
						}
					}
				}
			}
			base.SceneMgrSington.AddScene(scene)
		} else {
			s.SetAttribute(base.SessionAttributeScene, scene)
			pd := msg.GetPlayers()
			for _, p := range pd {
				if p.GetPos() == 7 {
					scene.AddPlayer(NewBaccaratPlayer(p))
					break
				}
			}
		}
	} else {
		logger.Logger.Error("SCBaccaratRoomInfo package data error.")
	}
	return nil
}

type SCBaccaratPlayerOpPacketFactory struct {
}

type SCBaccaratPlayerOpHandler struct {
}

func (this *SCBaccaratPlayerOpPacketFactory) CreatePacket() interface{} {
	pack := &proto_baccarat.SCBaccaratOp{}
	return pack
}

func (this *SCBaccaratPlayerOpHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	//logger.Logger.Tracef("(this *SCBaccaratPlayerOpHandler) Process [%v].", s.GetSessionConfig().Id)
	if scBaccaratOp, ok := pack.(*proto_baccarat.SCBaccaratOp); ok {
		if scene, ok := s.GetAttribute(base.SessionAttributeScene).(*BaccaratScene); ok {
			player := scene.GetMe(s)
			if player == base.NilPlayer {
				return nil
			}
			params := scBaccaratOp.GetParams()
			if me, ok := player.(*BaccaratPlayer); ok && me != BaccaratNilPlayer {
				if int(scBaccaratOp.GetOpRetCode()) == 0 {
					switch scBaccaratOp.GetOpCode() {
					case int32(baccarat.BaccaratPlayerOpBet):
						if len(params) >= 4 && params[0] >= 0 && int(params[0]) < Baccarat_Zone_Max {
							scene.totalBet[params[0]] = int64(params[2])
						}
					}
					if scBaccaratOp.GetSnId() == me.GetSnId() {
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
		logger.Logger.Error("SCBaccaratPlayerOp package data error.")
	}
	return nil
}

type SCBaccaratRoomStatePacketFactory struct {
}

type SCBaccaratRoomStateHandler struct {
}

func (this *SCBaccaratRoomStatePacketFactory) CreatePacket() interface{} {
	pack := &proto_baccarat.SCBaccaratRoomState{}
	return pack
}

func (this *SCBaccaratRoomStateHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	logger.Logger.Tracef("(this *SCBaccaratRoomStateHandler) Process [%v].", s.GetSessionConfig().Id)
	if scBaccaratRoomState, ok := pack.(*proto_baccarat.SCBaccaratRoomState); ok {
		if scene, ok := s.GetAttribute(base.SessionAttributeScene).(*BaccaratScene); ok {
			scene.State = scBaccaratRoomState.State
			p := scene.GetMe(s)
			if p == base.NilPlayer {
				return nil
			}
			if me, ok := p.(*BaccaratPlayer); ok && me != BaccaratNilPlayer {
				//logger.Logger.Trace(scBaccaratRoomState)
				switch scBaccaratRoomState.GetState() {
				case int32(baccarat.BaccaratSceneStateStake):
					scene.Action(s, me)
				case int32(baccarat.BaccaratSceneStateBilled):
					scene.Clear()
					me.Clear()
				}
			}
		}
	} else {
		logger.Logger.Error("SCBaccaratRoomState package data error.")
	}
	return nil
}

type SCBaccaratGameBilledPacketFactory struct {
}

type SCBaccaratGameBilledHandler struct {
}

func (this *SCBaccaratGameBilledPacketFactory) CreatePacket() interface{} {
	pack := &proto_baccarat.SCBaccaratBilled{}
	return pack
}

func (this *SCBaccaratGameBilledHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	logger.Logger.Tracef("(this *SCBaccaratGameBilledHandler) Process [%v].", s.GetSessionConfig().Id)
	if scBaccaratBilled, ok := pack.(*proto_baccarat.SCBaccaratBilled); ok {
		if scene, ok := s.GetAttribute(base.SessionAttributeScene).(*BaccaratScene); ok {
			//logger.Logger.Trace(scBaccaratBilled)
			billData := scBaccaratBilled.GetBillData()
			for _, data := range billData {
				p := scene.GetMe(s)
				if p == base.NilPlayer {
					continue
				}
				if me, ok := p.(*BaccaratPlayer); ok && me != BaccaratNilPlayer {
					if data.GetSnId() == me.GetSnId() { //自己的数据
						me.Coin = proto.Int64(data.GetCoin())
					}
				}
			}
		}
	} else {
		logger.Logger.Error("SCBaccaratGameBilled package data error.")
	}
	return nil
}

type SCBaccaratBankerListPacketFactory struct {
}

type SCBaccaratBankerListHandler struct {
}

func (this *SCBaccaratBankerListPacketFactory) CreatePacket() interface{} {
	pack := &proto_baccarat.SCBaccaratBankerList{}
	return pack
}

func (this *SCBaccaratBankerListHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	logger.Logger.Tracef("(this *SCBaccaratBankerListHandler) Process [%v].", s.GetSessionConfig().Id)
	if msg, ok := pack.(*proto_baccarat.SCBaccaratBankerList); ok {
		if scene, ok := s.GetAttribute(base.SessionAttributeScene).(*BaccaratScene); ok {
			scene.upBankerListNum = msg.GetCount()
			p := scene.GetMe(s)
			if me, ok := p.(*BaccaratPlayer); ok && me != BaccaratNilPlayer {
				scene.UpBanker(me)
			}
		}
	} else {
		logger.Logger.Error("SCBaccaratBankerList package data error.")
	}
	return nil
}
func init() {
	//SCBaccaratRoomInfo
	netlib.RegisterHandler(int(proto_baccarat.BaccaratPacketID_PACKET_SC_BACCARAT_ROOMINFO), &SCBaccaratRoomInfoHandler{})
	netlib.RegisterFactory(int(proto_baccarat.BaccaratPacketID_PACKET_SC_BACCARAT_ROOMINFO), &SCBaccaratRoomInfoPacketFactory{})
	//SCBaccaratPlayerOp
	netlib.RegisterHandler(int(proto_baccarat.BaccaratPacketID_PACKET_SC_BACCARAT_PLAYEROP), &SCBaccaratPlayerOpHandler{})
	netlib.RegisterFactory(int(proto_baccarat.BaccaratPacketID_PACKET_SC_BACCARAT_PLAYEROP), &SCBaccaratPlayerOpPacketFactory{})
	//SCBaccaratRoomState
	netlib.RegisterHandler(int(proto_baccarat.BaccaratPacketID_PACKET_SC_BACCARAT_ROOMSTATE), &SCBaccaratRoomStateHandler{})
	netlib.RegisterFactory(int(proto_baccarat.BaccaratPacketID_PACKET_SC_BACCARAT_ROOMSTATE), &SCBaccaratRoomStatePacketFactory{})
	//SCBaccaratFinalBilled
	netlib.RegisterHandler(int(proto_baccarat.BaccaratPacketID_PACKET_SC_BACCARAT_GAMEBILLED), &SCBaccaratGameBilledHandler{})
	netlib.RegisterFactory(int(proto_baccarat.BaccaratPacketID_PACKET_SC_BACCARAT_GAMEBILLED), &SCBaccaratGameBilledPacketFactory{})
	//SCBaccaratBankerList
	netlib.RegisterHandler(int(proto_baccarat.BaccaratPacketID_PACKET_SC_BACCARAT_BANKERLIST), &SCBaccaratBankerListHandler{})
	netlib.RegisterFactory(int(proto_baccarat.BaccaratPacketID_PACKET_SC_BACCARAT_BANKERLIST), &SCBaccaratBankerListPacketFactory{})
}
