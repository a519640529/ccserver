package main

import (
	"games.yol.com/win88/common"
	"games.yol.com/win88/protocol/login"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/srvlib"
)

func init() {
	netlib.RegisterFactory(int(login.GatePacketID_PACKET_SS_DICONNECT), netlib.PacketFactoryWrapper(func() interface{} {
		return &login.SSDisconnect{}
	}))

	netlib.RegisterHandler(int(login.GatePacketID_PACKET_SS_DICONNECT), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
		logger.Logger.Trace("receive SSDisconnect", pack)
		if ssdis, ok := pack.(*login.SSDisconnect); ok {
			client := srvlib.ClientSessionMgrSington.GetSession(ssdis.GetSessionId())
			if client != nil {
				client.SetAttribute(common.ClientSessionAttribute_State, common.ClientState_Logouted)
				if ssdis.GetType() != 2 { //非异常情况要告知客户端什么原因断线
					client.Send(int(login.GatePacketID_PACKET_SS_DICONNECT), ssdis)
				}
				client.Close()
			}
		}
		return nil
	}))
}
