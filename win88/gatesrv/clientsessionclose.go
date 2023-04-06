package main

import (
	"games.yol.com/win88/common"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/login"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/srvlib"
)

type ClientSessionCloseHandler struct {
}

func (sfcl ClientSessionCloseHandler) GetName() string {
	return "handler-client-close"
}

func (this *ClientSessionCloseHandler) GetInterestOps() uint {
	return 1<<netlib.InterestOps_Closed | 1<<netlib.InterestOps_Received
}

func (this *ClientSessionCloseHandler) OnSessionClosed(s *netlib.Session) {
	logger.Logger.Trace("Client handler close recv ", s.Id)

	var sid int64
	attr := s.GetAttribute(srvlib.SessionAttributeClientSession)
	if attr != nil {
		if sessId, ok := attr.(srvlib.SessionId); ok {
			sid = sessId.Get()
		}
	}

	//清理组标记
	attr = s.GetAttribute(common.ClientSessionAttribute_GroupTag)
	if tags, ok := attr.([]string); ok {
		CustomGroupMgrSington.DelFromGroup(tags, sid)
	}

	state := s.GetAttribute(common.ClientSessionAttribute_State)
	if state == common.ClientState_Logouted || state == common.ClientState_WaitLogout {
		return
	}

	//Get worldserver session
	ss := srvlib.ServerSessionMgrSington.GetSession(common.GetSelfAreaId(), srvlib.WorldServerType, common.GetWorldSrvId())
	if ss == nil {
		return
	}

	if sid != 0 {
		pack := &login.SSDisconnect{
			Type:      proto.Int32(common.KickReason_Disconnection),
			SessionId: proto.Int64(sid),
		}
		proto.SetDefaults(pack)
		ss.Send(int(login.GatePacketID_PACKET_SS_DICONNECT), pack)
		logger.Logger.Trace("client close ", sid)
	} else {
		logger.Logger.Tracef("client session attibute get error.")
	}
	s.SetAttribute(common.ClientSessionAttribute_State, common.ClientState_WaitLogout)
	return
}
func (this *ClientSessionCloseHandler) OnSessionOpened(s *netlib.Session) {
}
func (this *ClientSessionCloseHandler) OnSessionIdle(s *netlib.Session) {
}
func (this *ClientSessionCloseHandler) OnPacketReceived(s *netlib.Session, packetid int, logicNo uint32, packet interface{}) {
	logger.Logger.Tracef("SessionHandlerBase.OnPacketReceived")
	//	pack := &protocol.SCPong{}
	//	proto.SetDefaults(pack)
	//	s.Send(pack)
}

func (this *ClientSessionCloseHandler) OnPacketSent(s *netlib.Session, packetId int, logicNo uint32, data []byte) {
}

func init() {
	netlib.RegisteSessionHandlerCreator("handler-client-close", func() netlib.SessionHandler {
		return &ClientSessionCloseHandler{}
	})
}
