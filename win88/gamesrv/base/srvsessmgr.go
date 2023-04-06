package base

import (
	"games.yol.com/win88/common"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/srvlib"
	"github.com/idealeak/goserver/srvlib/protocol"
)

var SrvSessMgrSington = &SrvSessMgr{}

type SrvSessMgr struct {
}

// 注册事件
func (this *SrvSessMgr) OnRegiste(s *netlib.Session) {
	attr := s.GetAttribute(srvlib.SessionAttributeServerInfo)
	if attr != nil {
		if srvInfo, ok := attr.(*protocol.SSSrvRegiste); ok && srvInfo != nil {
			if srvInfo.GetType() == srvlib.WorldServiceType {
				logger.Logger.Warn("(this *SrvSessMgr) OnRegiste (WorldSrv):", s)
			} else if srvInfo.GetType() == srvlib.GateServiceType {
				logger.Logger.Warn("(this *SrvSessMgr) OnRegiste (GateSrv):", s)
			} else if srvInfo.GetType() == int32(common.RobotServerType) {
				logger.Logger.Warn("(this *SrvSessMgr) OnRegiste (RobotSrv):", s)
				NpcServerAgentSington.OnConnected()
			}
		}
	}
}

// 注销事件
func (this *SrvSessMgr) OnUnregiste(s *netlib.Session) {
	attr := s.GetAttribute(srvlib.SessionAttributeServerInfo)
	if attr != nil {
		if srvInfo, ok := attr.(*protocol.SSSrvRegiste); ok && srvInfo != nil {
			if srvInfo.GetType() == srvlib.WorldServiceType {
				logger.Logger.Warn("(this *SrvSessMgr) OnUnregiste (WorldSrv):", s)
			} else if srvInfo.GetType() == srvlib.GateServiceType {
				logger.Logger.Warn("(this *SrvSessMgr) OnUnregiste (GateSrv):", s)
			} else if srvInfo.GetType() == int32(common.RobotServerType) {
				logger.Logger.Warn("(this *SrvSessMgr) OnUnregiste (RobotSrv):", s)
				NpcServerAgentSington.OnDisconnected()
			}
		}
	}
}

func init() {
	srvlib.ServerSessionMgrSington.AddListener(SrvSessMgrSington)
}
