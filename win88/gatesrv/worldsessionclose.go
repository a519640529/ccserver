package main

import (
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/srvlib"
	"github.com/idealeak/goserver/srvlib/protocol"
)

const (
	WorldSessionCloseHandlerName = "handler-world-close"
)

type WorldSessionCloseHandler struct {
	netlib.BasicSessionHandler
}

func (sfcl WorldSessionCloseHandler) GetName() string {
	return WorldSessionCloseHandlerName
}

func (this *WorldSessionCloseHandler) GetInterestOps() uint {
	return 1 << netlib.InterestOps_Closed
}

func (this *WorldSessionCloseHandler) OnSessionClosed(s *netlib.Session) {
	logger.Logger.Warn("WorldSessionCloseHandler OnSessionClosed ", s.Id)
	//close all client session
	param := s.GetAttribute(srvlib.SessionAttributeServerInfo)
	if registePacket, ok := param.(*protocol.SSSrvRegiste); ok {
		if registePacket.GetType() == srvlib.WorldServiceType {
			srvlib.ClientSessionMgrSington.CloseAll()
		}
	}

	return
}

func init() {
	netlib.RegisteSessionHandlerCreator(WorldSessionCloseHandlerName, func() netlib.SessionHandler {
		return &WorldSessionCloseHandler{}
	})
}
