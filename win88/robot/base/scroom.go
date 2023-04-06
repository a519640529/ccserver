package base

import (
	gamehall_proto "games.yol.com/win88/protocol/gamehall"
	msg_proto "games.yol.com/win88/protocol/message"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
)

type SCEnterRoomPacketFactory struct {
}

type SCEnterRoomHandler struct {
}

func (this *SCEnterRoomPacketFactory) CreatePacket() interface{} {
	pack := &gamehall_proto.SCEnterRoom{}
	return pack
}

func (this *SCEnterRoomHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	logger.Logger.Tracef("(this *SCEnterRoomHandler) Process [%v].", s.GetSessionConfig().Id)
	if SCEnterRoom, ok := pack.(*gamehall_proto.SCEnterRoom); ok {
		logger.Logger.Trace(SCEnterRoom)
		if SCEnterRoom.GetOpRetCode() == gamehall_proto.OpResultCode_Game_OPRC_Sucess_Game {
			//上传坐标
			//			longitude, latitude := RandLongitudeAndLatitude()
			//			zone := RandZone()
			//			csUploadLoc := &gamehall_proto.CSUploadLoc{
			//				Longitude: proto.Int32(longitude),
			//				Latitude:  proto.Int32(latitude),
			//				City:      proto.String(zone),
			//			}
			//			proto.SetDefaults(csUploadLoc)
			//			s.Send(int(gamehall_proto.MmoPacketID_PACKET_CS_UPLOADLOC), csUploadLoc)

			//			logger.Logger.Trace("(this *SCEnterRoomHandler) Process CSUploadLoc ", csUploadLoc)
			s.RemoveAttribute(SessionAttributeCoinSceneQueue)
			s.SetAttribute(SessionAttributeSceneId, SCEnterRoom.GetRoomId())
			return nil
		} else {
			logger.Logger.Trace("EnterRoom failed.")
			if strategy, ok := s.GetAttribute(SessionAttributeStrategy).(int); ok {
				ChooiesStrategy(s, strategy)
			}
		}
	} else {
		logger.Logger.Error("SCEnterRoom package data error.")
	}
	return nil
}

type SCCreateRoomPacketFactory struct {
}

type SCCreateRoomHandler struct {
}

func (this *SCCreateRoomPacketFactory) CreatePacket() interface{} {
	pack := &gamehall_proto.SCCreateRoom{}
	return pack
}

func (this *SCCreateRoomHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	logger.Logger.Tracef("(this *SCCreateRoomHandler) Process [%v].", s.GetSessionConfig().Id)
	if scCreateRoom, ok := pack.(*gamehall_proto.SCCreateRoom); ok {
		logger.Logger.Trace(scCreateRoom)
		if scCreateRoom.GetOpRetCode() == gamehall_proto.OpResultCode_Game_OPRC_Sucess_Game {
			if strategy, ok := s.GetAttribute(SessionAttributeStrategy).(int); ok {
				ChooiesStrategy(s, strategy)
			}
			return nil
		} else {
			logger.Logger.Trace("SCCreateRoomHandler failed.")
			if strategy, ok := s.GetAttribute(SessionAttributeStrategy).(int); ok {
				ChooiesStrategy(s, strategy)
			}
		}
	} else {
		logger.Logger.Error("SCCreateRoom package data error.")
	}
	return nil
}

type SCDestroyRoomPacketFactory struct {
}

type SCDestroyRoomHandler struct {
}

func (this *SCDestroyRoomPacketFactory) CreatePacket() interface{} {
	pack := &gamehall_proto.SCDestroyRoom{}
	return pack
}

func (this *SCDestroyRoomHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	logger.Logger.Tracef("(this *SCDestroyRoomHandler) Process [%v].", s.GetSessionConfig().Id)
	if scDestroyRoom, ok := pack.(*gamehall_proto.SCDestroyRoom); ok {
		logger.Logger.Trace(scDestroyRoom)
		if scDestroyRoom.GetOpRetCode() == gamehall_proto.OpResultCode_Game_OPRC_Sucess_Game {
			StopSessionGameTimer(s)
			s.RemoveAttribute(SessionAttributeScene)
			s.RemoveAttribute(SessionAttributeSceneId)
			if strategy, ok := s.GetAttribute(SessionAttributeStrategy).(int); ok {
				ChooiesStrategy(s, strategy)
			}
			return nil
		} else {
			logger.Logger.Trace("DestroyRoom failed.")
			if strategy, ok := s.GetAttribute(SessionAttributeStrategy).(int); ok {
				ChooiesStrategy(s, strategy)
			}
		}
	} else {
		logger.Logger.Error("SCDestroyRoom package data error.")
	}
	return nil
}

type SCLeaveRoomPacketFactory struct {
}

type SCLeaveRoomHandler struct {
}

func (this *SCLeaveRoomPacketFactory) CreatePacket() interface{} {
	pack := &gamehall_proto.SCLeaveRoom{}
	return pack
}

func (this *SCLeaveRoomHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	logger.Logger.Tracef("(this *SCLeaveRoomHandler) Process [%v].", s.GetSessionConfig().Id)
	if scLeaveRoom, ok := pack.(*gamehall_proto.SCLeaveRoom); ok {
		logger.Logger.Trace(scLeaveRoom)
		if scLeaveRoom.GetOpRetCode() == gamehall_proto.OpResultCode_Game_OPRC_Sucess_Game {
			if scene, ok := GetScene(s).(Scene); ok && scene != nil {
				p := scene.GetMe(s)

				if p != nil {
					logger.Logger.Trace("(this *SCLeaveRoomHandler) snid [%v].", p.GetSnId())
					scene.DelPlayer(p.GetSnId())
				}
			}
			StopSessionGameTimer(s)
			s.RemoveAttribute(SessionAttributeScene)
			s.RemoveAttribute(SessionAttributeSceneId)
			if !ClientMgrSington.isWaitCloseSession(s) {
				if strategy, ok := s.GetAttribute(SessionAttributeStrategy).(int); ok {
					ChooiesStrategy(s, strategy)
				}
			}
			return nil
		} else {
			logger.Logger.Trace("LeaveRoom failed.")
		}
	} else {
		logger.Logger.Error("SCLeaveRoom package data error.")
	}
	return nil
}

type SCReturnRoomPacketFactory struct {
}

type SCReturnRoomHandler struct {
}

func (this *SCReturnRoomPacketFactory) CreatePacket() interface{} {
	pack := &gamehall_proto.SCReturnRoom{}
	return pack
}

func (this *SCReturnRoomHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	logger.Logger.Tracef("(this *SCReturnRoomHandler) Process [%v].", s.GetSessionConfig().Id)
	if scReturnRoom, ok := pack.(*gamehall_proto.SCReturnRoom); ok {
		logger.Logger.Trace(scReturnRoom)
		if scReturnRoom.GetOpRetCode() == gamehall_proto.OpResultCode_Game_OPRC_Sucess_Game {
			return nil
		} else {
			logger.Logger.Trace("ReturnRoom failed.")
		}
	} else {
		logger.Logger.Error("SCReturnRoom package data error.")
	}
	return nil
}

type SCNoticePacketFactory struct {
}

type SCNoticeHandler struct {
}

func (this *SCNoticePacketFactory) CreatePacket() interface{} {
	pack := &msg_proto.SCNotice{}
	return pack
}

func (this *SCNoticeHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	logger.Logger.Tracef("(this *SCNoticeHandler) Process [%v].", s.GetSessionConfig().Id)
	if SCNotice, ok := pack.(*msg_proto.SCNotice); ok {
		logger.Logger.Trace(SCNotice)
	} else {
		logger.Logger.Error("SCNotice package data error.")
	}
	return nil
}

type SCQuitGamePacketFactory struct {
}

type SCQuitGameHandler struct {
}

func (this *SCQuitGamePacketFactory) CreatePacket() interface{} {
	pack := &gamehall_proto.SCQuitGame{}
	return pack
}

func (this *SCQuitGameHandler) Process(s *netlib.Session, packid int, pack interface{}) error {
	logger.Logger.Tracef("(this *SCQuitGameHandler) Process [%v].", s.GetSessionConfig().Id)
	if scQuitGame, ok := pack.(*gamehall_proto.SCQuitGame); ok {
		logger.Logger.Trace(scQuitGame)
		if scQuitGame.GetOpCode() == gamehall_proto.OpResultCode_Game_OPRC_Sucess_Game {
			if scene, ok := GetScene(s).(Scene); ok && scene != nil {
				p := scene.GetMe(s)

				if p != nil {
					logger.Logger.Trace("(this *SCQuitGameHandler) snid [%v].", p.GetSnId())
					scene.DelPlayer(p.GetSnId())
				}
			}
			StopSessionGameTimer(s)
			s.RemoveAttribute(SessionAttributeScene)
			s.RemoveAttribute(SessionAttributeSceneId)
			if !ClientMgrSington.isWaitCloseSession(s) {
				if strategy, ok := s.GetAttribute(SessionAttributeStrategy).(int); ok {
					ChooiesStrategy(s, strategy)
				}
			}
			return nil
		} else {
			logger.Logger.Trace("scQuitGame failed.")
		}
	} else {
		logger.Logger.Error("scQuitGame package data error.")
	}
	return nil
}

func init() {
	//SCCreateRoom
	netlib.RegisterHandler(int(gamehall_proto.GameHallPacketID_PACKET_SC_CREATEROOM), &SCCreateRoomHandler{})
	netlib.RegisterFactory(int(gamehall_proto.GameHallPacketID_PACKET_SC_CREATEROOM), &SCCreateRoomPacketFactory{})

	//SCEnterRoom
	netlib.RegisterHandler(int(gamehall_proto.GameHallPacketID_PACKET_SC_ENTERROOM), &SCEnterRoomHandler{})
	netlib.RegisterFactory(int(gamehall_proto.GameHallPacketID_PACKET_SC_ENTERROOM), &SCEnterRoomPacketFactory{})
	//SCDestroyRoom
	netlib.RegisterHandler(int(gamehall_proto.GameHallPacketID_PACKET_SC_DESTROYROOM), &SCDestroyRoomHandler{})
	netlib.RegisterFactory(int(gamehall_proto.GameHallPacketID_PACKET_SC_DESTROYROOM), &SCDestroyRoomPacketFactory{})
	//SCLeaveRoom
	netlib.RegisterHandler(int(gamehall_proto.GameHallPacketID_PACKET_SC_LEAVEROOM), &SCLeaveRoomHandler{})
	netlib.RegisterFactory(int(gamehall_proto.GameHallPacketID_PACKET_SC_LEAVEROOM), &SCLeaveRoomPacketFactory{})
	//SCReturnRoom
	netlib.RegisterHandler(int(gamehall_proto.GameHallPacketID_PACKET_SC_RETURNROOM), &SCReturnRoomHandler{})
	netlib.RegisterFactory(int(gamehall_proto.GameHallPacketID_PACKET_SC_RETURNROOM), &SCReturnRoomPacketFactory{})
	//SCNotice
	netlib.RegisterHandler(int(msg_proto.MSGPacketID_PACKET_SC_NOTICE), &SCNoticeHandler{})
	netlib.RegisterFactory(int(msg_proto.MSGPacketID_PACKET_SC_NOTICE), &SCNoticePacketFactory{})
	//SCQuitGame
	netlib.RegisterHandler(int(gamehall_proto.GameHallPacketID_PACKET_SC_QUITGAME), &SCQuitGameHandler{})
	netlib.RegisterFactory(int(gamehall_proto.GameHallPacketID_PACKET_SC_QUITGAME), &SCQuitGamePacketFactory{})
}
