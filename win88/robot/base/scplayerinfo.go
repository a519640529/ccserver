package base

import (
	"fmt"
	"games.yol.com/win88/common"
	hall_proto "games.yol.com/win88/protocol/gamehall"
	login_proto "games.yol.com/win88/protocol/login"
	msg_proto "games.yol.com/win88/protocol/message"
	player_proto "games.yol.com/win88/protocol/player"
	"math/rand"
	"time"

	"games.yol.com/win88/proto"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/core/timer"
)

type SCPlayerInfoPacketFactory struct {
}

type SCPlayerInfoHandler struct {
}

func (this *SCPlayerInfoPacketFactory) CreatePacket() interface{} {
	pack := &player_proto.SCPlayerData{}
	return pack
}

func (this *SCPlayerInfoHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	logger.Logger.Tracef("(this *SCPlayerInfoHandler) Process [%v].", s.GetSessionConfig().Id)
	if scPlayerInfo, ok := pack.(*player_proto.SCPlayerData); ok {
		logger.Logger.Trace(scPlayerInfo)
		if scPlayerInfo.GetOpRetCode() == player_proto.OpResultCode_OPRC_Sucess {
			s.SetAttribute(SessionAttributeUser, scPlayerInfo)
			PlayerMgrSington.AddPlayer(scPlayerInfo, s)
			data := scPlayerInfo.GetData()
			if data != nil {
				coin := data.GetCoin()
				if coin < 1000000 {
					ExePMCmd(s, fmt.Sprintf("%v%v%v", common.PMCmd_AddCoin, common.PMCmd_SplitToken, 1000000))
				}
			}

			strategy := StrategyMgrSington.PopStrategy()
			if scPlayerInfo.GetRoomId() != 0 {
				strategy := StrategyMgrSington.PopStrategy()
				s.SetAttribute(SessionAttributeSceneId, scPlayerInfo.GetRoomId())
				s.SetAttribute(SessionAttributeStrategy, strategy)
				scReturnRoom := &hall_proto.CSReturnRoom{}
				proto.SetDefaults(scReturnRoom)
				s.Send(int(hall_proto.GameHallPacketID_PACKET_CS_RETURNROOM), scReturnRoom)
			} else {
				s.SetAttribute(SessionAttributeStrategy, strategy)
				ChooiesStrategy(s, strategy)
			}
			StartSessionPingTimer(s, timer.TimerActionWrapper(func(h timer.TimerHandle, ud interface{}) bool {
				if !s.IsConned() {
					StopSessionPingTimer(s)
					return false
				}
				pack := &login_proto.CSPing{}
				s.Send(int(login_proto.GatePacketID_PACKET_CS_PING), pack)
				ChooiesStrategy(s, strategy)
				return true
			}), nil, time.Second*time.Duration(60+rand.Int31n(100)), -1)
			return nil
		} else {
			logger.Logger.Trace("Get player info failed.")
		}
	} else {
		logger.Logger.Error("SCPlayerInfo package data error.")
	}
	s.Close()
	return nil
}

type SCPlayerInfoUpdatePacketFactory struct {
}

type SCPlayerInfoUpdateHandler struct {
}

func (this *SCPlayerInfoUpdatePacketFactory) CreatePacket() interface{} {
	pack := &player_proto.SCPlayerDataUpdate{}
	return pack
}

func (this *SCPlayerInfoUpdateHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	logger.Logger.Tracef("(this *SCPlayerInfoUpdateHandler) Process [%v].", s.GetSessionConfig().Id)
	if SCPlayerDataUpdate, ok := pack.(*player_proto.SCPlayerDataUpdate); ok {
		logger.Logger.Trace(SCPlayerDataUpdate)
		if SCPlayerData, ok := s.GetAttribute(SessionAttributeUser).(*player_proto.SCPlayerData); ok {
			SCPlayerData.GetData().Coin = SCPlayerDataUpdate.Coin
		}
	} else {
		logger.Logger.Error("SCPlayerDataUpdate package data error.")
	}
	return nil
}

type SCPlayerFlagPacketFactory struct {
}

type SCPlayerFlagHandler struct {
}

func (this *SCPlayerFlagPacketFactory) CreatePacket() interface{} {
	pack := &player_proto.SCPlayerFlag{}
	return pack
}

func (this *SCPlayerFlagHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	logger.Logger.Tracef("(this *SCPlayerFlagHandler) Process [%v].", s.GetSessionConfig().Id)
	if SCPlayerFlag, ok := pack.(*player_proto.SCPlayerFlag); ok {
		logger.Logger.Trace(SCPlayerFlag)
		if scene, ok := GetScene(s).(Scene); ok && scene != nil {
			p := scene.GetPlayerBySnid(SCPlayerFlag.GetPlayerId())
			if p != nil {
				p.SetFlag(SCPlayerFlag.GetFlag())
			}
		}
	} else {
		logger.Logger.Error("SCPlayerFlag package data error.")
	}
	return nil
}

type SCMessageListPacketFactory struct {
}

type SCMessageListHandler struct {
}

func (this *SCMessageListPacketFactory) CreatePacket() interface{} {
	pack := &msg_proto.SCMessageList{}
	return pack
}

func (this *SCMessageListHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	logger.Logger.Tracef("(this *SCMessageListHandler) Process [%v].", s.GetSessionConfig().Id)
	if SCMessageList, ok := pack.(*msg_proto.SCMessageList); ok {
		logger.Logger.Trace(SCMessageList)
	} else {
		logger.Logger.Error("SCMessageList package data error.")
	}
	return nil
}

type SCUploadLocPacketFactory struct {
}

type SCUploadLocHandler struct {
}

func (this *SCUploadLocPacketFactory) CreatePacket() interface{} {
	pack := &hall_proto.SCUploadLoc{}
	return pack
}

func (this *SCUploadLocHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	logger.Logger.Tracef("(this *SCUploadLocHandler) Process [%v].", s.GetSessionConfig().Id)
	if SCUploadLoc, ok := pack.(*hall_proto.SCUploadLoc); ok {
		logger.Logger.Trace(SCUploadLoc)
	} else {
		logger.Logger.Error("SCUploadLoc package data error.")
	}
	return nil
}

func init() {
	//SCPlayerInfo
	netlib.RegisterHandler(int(player_proto.PlayerPacketID_PACKET_SC_PLAYERDATA), &SCPlayerInfoHandler{})
	netlib.RegisterFactory(int(player_proto.PlayerPacketID_PACKET_SC_PLAYERDATA), &SCPlayerInfoPacketFactory{})

	//SCPlayerDataUpdate
	netlib.RegisterHandler(int(player_proto.PlayerPacketID_PACKET_SC_PLAYERDATAUPDATE), &SCPlayerInfoUpdateHandler{})
	netlib.RegisterFactory(int(player_proto.PlayerPacketID_PACKET_SC_PLAYERDATAUPDATE), &SCPlayerInfoUpdatePacketFactory{})

	//SCPlayerFlag
	netlib.RegisterHandler(int(player_proto.PlayerPacketID_PACKET_SC_PLAYERFLAG), &SCPlayerFlagHandler{})
	netlib.RegisterFactory(int(player_proto.PlayerPacketID_PACKET_SC_PLAYERFLAG), &SCPlayerFlagPacketFactory{})

	//SCMessageList
	netlib.RegisterHandler(int(msg_proto.MSGPacketID_PACKET_SC_MESSAGELIST), &SCMessageListHandler{})
	netlib.RegisterFactory(int(msg_proto.MSGPacketID_PACKET_SC_MESSAGELIST), &SCMessageListPacketFactory{})

	//SCUploadLoc
	netlib.RegisterHandler(int(hall_proto.GameHallPacketID_PACKET_SC_UPLOADLOC), &SCUploadLocHandler{})
	netlib.RegisterFactory(int(hall_proto.GameHallPacketID_PACKET_SC_UPLOADLOC), &SCUploadLocPacketFactory{})
}
