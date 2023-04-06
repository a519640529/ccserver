package base

import (
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/srvlib"
	"github.com/idealeak/goserver/srvlib/protocol"
)

var WorldSessMgrSington = &WorldSessMgr{}

type WorldSessMgr struct {
}

// 注册事件
func (this *WorldSessMgr) OnRegiste(s *netlib.Session) {
	attr := s.GetAttribute(srvlib.SessionAttributeServerInfo)
	if attr != nil {
		if srvInfo, ok := attr.(*protocol.SSSrvRegiste); ok && srvInfo != nil {
			if srvInfo.GetType() == srvlib.WorldServiceType {
				logger.Logger.Warn("(this *WorldSessMgr) OnRegiste (Srv):", s)
			}
		}
	}
}

// 注销事件
func (this *WorldSessMgr) OnUnregiste(s *netlib.Session) {
	attr := s.GetAttribute(srvlib.SessionAttributeServerInfo)
	if attr != nil {
		if srvInfo, ok := attr.(*protocol.SSSrvRegiste); ok && srvInfo != nil {
			if srvInfo.GetType() == srvlib.WorldServiceType {
				logger.Logger.Warn("(this *WorldSessMgr) OnUnregiste (Srv):", s)
			}
		}
	}
}

func init() {
	srvlib.ServerSessionMgrSington.AddListener(WorldSessMgrSington)
}
