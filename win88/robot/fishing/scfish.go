package fishing

import (
	"games.yol.com/win88/common"
	"games.yol.com/win88/proto"
	fish_proto "games.yol.com/win88/protocol/fishing"
	hall_proto "games.yol.com/win88/protocol/gamehall"
	"games.yol.com/win88/robot/base"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
)

//FIPacketID_FISHING_SC_ROOMINFO
type SCFishingRoomInfoPacketFactory struct {
}

type SCFishingRoomInfoHandler struct {
}

func (this *SCFishingRoomInfoPacketFactory) CreatePacket() interface{} {
	pack := &fish_proto.SCFishingRoomInfo{}
	return pack
}

func (this *SCFishingRoomInfoHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	logger.Logger.Trace("(this *SCFishingRoomInfoHandler) Process ", s.GetSessionConfig().Id, pack)
	if msg, ok := pack.(*fish_proto.SCFishingRoomInfo); ok {
		sceneEx, _ := base.SceneMgrSington.GetScene(msg.GetRoomId()).(*FishingScene)
		if sceneEx == nil {
			sceneEx = NewFishingScene(msg)
			base.SceneMgrSington.AddScene(sceneEx)
		}
		if sceneEx != nil {
			for _, pd := range msg.GetPlayers() {
				if oldPlayer, ok := sceneEx.GetPlayerBySnid(pd.GetSnId()).(*FishingPlayer); ok {
					oldPlayer.UpdatePlayerData(pd)
				} else {
					p := NewFishingPlayer(pd, sceneEx, sceneEx.GetGameFreeId(), sceneEx.GetRoomId())
					if p != nil {
						sceneEx.AddPlayer(p)
					}
				}
			}
			//logger.Logger.Trace(msg)
			s.SetAttribute(base.SessionAttributeSceneId, sceneEx.GetRoomId())

			var me *FishingPlayer
			if sceneEx.GetMe(s) != nil {
				me = sceneEx.GetMe(s).(*FishingPlayer)
				sceneEx.InitPlayer(s, me)
			}

		}
	} else {
		logger.Logger.Error("SCFishingRoomInfo package data error.")
	}
	return nil
}

type SCFishingRoomStatePacketFactory struct {
}

type SCFishingRoomStateHandler struct {
}

func (this *SCFishingRoomStatePacketFactory) CreatePacket() interface{} {
	pack := &fish_proto.SCFishingRoomState{}
	return pack
}

func (this *SCFishingRoomStateHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	if _, ok := pack.(*fish_proto.SCFishingRoomState); ok {
		if sceneEx, ok := base.GetScene(s).(*FishingScene); ok {
			sceneEx.Clear()
		}
	} else {
		logger.Logger.Error("SCFishingRoomStateHandler package data error.")
	}
	return nil
}

//FIPacketID_FISHING_SC_SEATS
type SCFishingSeatsPacketFactory struct {
}

type SCFishingSeatsHandler struct {
}

func (this *SCFishingSeatsPacketFactory) CreatePacket() interface{} {
	pack := &fish_proto.SCFishingSeats{}
	return pack
}

func (this *SCFishingSeatsHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	logger.Logger.Trace("(this *SCFishingSeatsHandler) Process ", s.GetSessionConfig().Id, pack)
	if scFishingSeats, ok := pack.(*fish_proto.SCFishingSeats); ok {
		if sceneEx, ok := base.GetScene(s).(*FishingScene); ok {
			var leftPlayers []int32
			for _, pd := range scFishingSeats.GetData() {
				if oldPlayer, ok := sceneEx.GetPlayerBySnid(pd.GetSnId()).(*FishingPlayer); ok {
					oldPlayer.UpdatePlayerData(pd)
				} else {
					p := NewFishingPlayer(pd, sceneEx, sceneEx.GetGameFreeId(), sceneEx.GetRoomId())
					if p != nil {
						sceneEx.AddPlayer(p)
					}
				}

				leftPlayers = append(leftPlayers, pd.GetSnId())
			}

			sceneEx.UpdateOnlinePlayers(leftPlayers)

			//if sceneEx.AllIsRobot() {
			//	sceneEx.Clear()
			//	base.SceneMgrSington.DelScene(sceneEx.GetRoomId())
			//}
		}
	} else {
		logger.Logger.Error("SCFishingSeatsHandler package data error.")
	}
	return nil
}

//FIPacketID_FISHING_SC_FIREPOWER
type SCFishingFirePowerPacketFactory struct {
}

type SCFishingFirePowerHandler struct {
}

func (this *SCFishingFirePowerPacketFactory) CreatePacket() interface{} {
	pack := &fish_proto.SCFirePower{}
	return pack
}

func (this *SCFishingFirePowerHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	//logger.Logger.Trace("(this *SCFishingFirePowerHandler) Process ", s.GetSessionConfig().Id, pack)
	if _, ok := pack.(*fish_proto.SCFirePower); ok {
		//logger.Logger.Trace(scPack)
	} else {
		logger.Logger.Error("SCFishingSeatsHandler package data error.")
	}
	return nil
}

//FIPacketID_FISHING_SC_FIREPOWER
type SCFishingSynFishPacketFactory struct {
}

type SCFishingSynFishHandler struct {
}

func (this *SCFishingSynFishPacketFactory) CreatePacket() interface{} {
	pack := &fish_proto.SCSyncRefreshFish{}
	return pack
}

func (this *SCFishingSynFishHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	//logger.Logger.Trace("(this *SCFishingSynFishHandler) Process ", s.GetSessionConfig().Id, pack)
	if msg, ok := pack.(*fish_proto.SCSyncRefreshFish); ok {
		//logger.Logger.Trace(scPack)
		if sceneEx, ok := base.GetScene(s).(*FishingScene); ok {
			sceneEx.flushFish(msg.GetPolicyId(), msg.GetTimePoint())
		}
	} else {
		logger.Logger.Error("SCFishingSynFishHandler package data error.")
	}
	return nil
}

//FIPacketID_FISHING_SC_FIREPOWER
type SCFishingOpPacketFactory struct {
}

type SCFishingOpHandler struct {
}

func (this *SCFishingOpPacketFactory) CreatePacket() interface{} {
	pack := &fish_proto.SCFishingOp{}
	return pack
}

func (this *SCFishingOpHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	if scPack, ok := pack.(*fish_proto.SCFishingOp); ok {
		if scPack.GetOpCode() == FishingPlayerOpRobotFire {
			logger.Logger.Infof("SCFishingSynFishHandler package data ==>  snid %v  opcode %v  opRetCode %v ", scPack.GetSnId(), scPack.GetOpCode(), scPack.GetOpRetCode())
			if scPack.GetOpRetCode() == fish_proto.OpResultCode_OPRC_CoinNotEnough || scPack.GetOpRetCode() == fish_proto.OpResultCode_OPRC_Sucess {
				//退出
				logger.Logger.Infof("SCFishingSynFishHandler package data ==>  snid %v  opcode %v  opRetCode %v ", scPack.GetSnId(), scPack.GetOpCode(), scPack.GetOpRetCode())
				pack := &hall_proto.CSCoinSceneOp{
					Id:     proto.Int32(0),
					OpType: proto.Int32(common.CoinSceneOp_Leave),
				}
				proto.SetDefaults(pack)
				//s.Send(int(fish_proto.CoinSceneGamePacketID_PACKET_CS_COINSCENE_OP), pack)
				base.DelaySend(s, int(hall_proto.CoinSceneGamePacketID_PACKET_CS_COINSCENE_OP), pack, 3, 10)
				//ExePMCmd(s, fmt.Sprintf("%v%v%v", common.PMCmd_AddCoin, common.PMCmd_SplitToken, 50000))
			}
		} else if scPack.GetOpCode() == FishingRobotWantLeave {
			pack := &hall_proto.CSCoinSceneOp{
				Id:     proto.Int32(0),
				OpType: proto.Int32(common.CoinSceneOp_Leave),
			}
			proto.SetDefaults(pack)
			base.DelaySend(s, int(hall_proto.CoinSceneGamePacketID_PACKET_CS_COINSCENE_OP), pack, 3, 10)
		}
	} else {
		logger.Logger.Error("SCFishingSynFishHandler package data error.")
	}
	return nil
}

//FIPacketID_FISHING_SC_FIREPOWER
type SCFishingFirePacketFactory struct {
}

type SCFishingFireHandler struct {
}

func (this *SCFishingFirePacketFactory) CreatePacket() interface{} {
	pack := &fish_proto.SCFire{}
	return pack
}

func (this *SCFishingFireHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	if _, ok := pack.(*fish_proto.SCFire); ok {

	} else {
		logger.Logger.Error("SCFishingFireHandler package data error.")
	}
	return nil
}

type SCFishingFireHitPacketFactory struct {
}

type SCFishingFireHitHandler struct {
}

func (this *SCFishingFireHitPacketFactory) CreatePacket() interface{} {
	pack := &fish_proto.SCFireHit{}
	return pack
}

func (this *SCFishingFireHitHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	if msg, ok := pack.(*fish_proto.SCFireHit); ok {
		if sceneEx, ok := base.GetScene(s).(*FishingScene); ok {
			ids := msg.GetFishId()
			for _, id := range ids {
				f := sceneEx.GetFish(id)
				if f != nil {
					sceneEx.DelFish(f)
				}
			}
		}
	} else {
		logger.Logger.Error("SCFishingFireHitHandler package data error.")
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
func init() {
	//SCFishingRoomInfo
	netlib.RegisterHandler(int(fish_proto.FIPacketID_FISHING_SC_ROOMINFO), &SCFishingRoomInfoHandler{})
	netlib.RegisterFactory(int(fish_proto.FIPacketID_FISHING_SC_ROOMINFO), &SCFishingRoomInfoPacketFactory{})
	//SCFishingSeats
	netlib.RegisterHandler(int(fish_proto.FIPacketID_FISHING_SC_SEATS), &SCFishingSeatsHandler{})
	netlib.RegisterFactory(int(fish_proto.FIPacketID_FISHING_SC_SEATS), &SCFishingSeatsPacketFactory{})

	//SCFishingSeats
	netlib.RegisterHandler(int(fish_proto.FIPacketID_FISHING_SC_FIREPOWER), &SCFishingFirePowerHandler{})
	netlib.RegisterFactory(int(fish_proto.FIPacketID_FISHING_SC_FIREPOWER), &SCFishingFirePowerPacketFactory{})

	//SCFishingSeats
	netlib.RegisterHandler(int(fish_proto.FIPacketID_FISHING_SC_SYNCFISH), &SCFishingSynFishHandler{})
	netlib.RegisterFactory(int(fish_proto.FIPacketID_FISHING_SC_SYNCFISH), &SCFishingSynFishPacketFactory{})

	//SCFishingOp
	netlib.RegisterHandler(int(fish_proto.FIPacketID_FISHING_SC_OP), &SCFishingOpHandler{})
	netlib.RegisterFactory(int(fish_proto.FIPacketID_FISHING_SC_OP), &SCFishingOpPacketFactory{})

	//SCFire
	netlib.RegisterHandler(int(fish_proto.FIPacketID_FISHING_SC_FIRE), &SCFishingFireHandler{})
	netlib.RegisterFactory(int(fish_proto.FIPacketID_FISHING_SC_FIRE), &SCFishingFirePacketFactory{})

	//SCFireHit
	netlib.RegisterHandler(int(fish_proto.FIPacketID_FISHING_SC_FIREHIT), &SCFishingFireHitHandler{})
	netlib.RegisterFactory(int(fish_proto.FIPacketID_FISHING_SC_FIREHIT), &SCFishingFireHitPacketFactory{})
}
