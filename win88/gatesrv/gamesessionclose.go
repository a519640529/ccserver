package main

import (
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
)

const (
	GameSessionCloseHandlerName = "handler-game-close"
)

type GameSessionCloseHandler struct {
	netlib.BasicSessionHandler
}

func (sfcl GameSessionCloseHandler) GetName() string {
	return GameSessionCloseHandlerName
}

func (this *GameSessionCloseHandler) GetInterestOps() uint {
	return 1 << netlib.InterestOps_Closed
}

func (this *GameSessionCloseHandler) OnSessionClosed(s *netlib.Session) {
	logger.Logger.Warn("GameSessionCloseHandler OnSessionClosed ", s.Id)
	//param := s.GetAttribute(srvlib.SessionAttributeServerInfo)
	//if registePacket, ok := param.(*protocol.SSSrvRegiste); ok {
	//	if registePacket.GetType() == srvlib.GameServiceType {
	//		sessions := srvlib.ClientSessionMgrSington.GetSessions()
	//		for _, ss := range sessions {
	//			gsid := ss.GetAttribute(common.ClientSessionAttribute_GameServer)
	//			if gsid != nil && gsid.(*netlib.Session) == s {
	//				ss.Close()
	//			}
	//		}
	//	}
	//}

	return
}

func init() {
	netlib.RegisteSessionHandlerCreator(GameSessionCloseHandlerName, func() netlib.SessionHandler {
		return &GameSessionCloseHandler{}
	})
}
