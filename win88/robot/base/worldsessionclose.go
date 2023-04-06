package base

import (
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/module"
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
	//logger.Logger.Trace("WorldSessionCloseHandler OnSessionClosed ", s.Id)
	//close all client session
	param := s.GetAttribute(srvlib.SessionAttributeServerInfo)
	if registePacket, ok := param.(*protocol.SSSrvRegiste); ok {
		switch registePacket.GetType() {
		case srvlib.WorldServiceType:
			module.Stop()
		case srvlib.GameServiceType:

		}
	} else {
		logger.Logger.Info("Session attribute param error.")
	}

	return
}

func init() {
	netlib.RegisteSessionHandlerCreator(WorldSessionCloseHandlerName, func() netlib.SessionHandler {
		return &WorldSessionCloseHandler{}
	})
}
