package main

import (
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/login"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
)

func init() {
	netlib.RegisterFactory(int(login.GatePacketID_PACKET_CS_PING), netlib.PacketFactoryWrapper(func() interface{} {
		return &login.CSPing{}
	}))

	netlib.RegisterHandler(int(login.GatePacketID_PACKET_CS_PING), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
		logger.Logger.Trace("receive gate load info==", pack)
		if msg, ok := pack.(*login.CSPing); ok {
			pong := &login.SCPong{
				TimeStamp: msg.TimeStamp,
			}
			proto.SetDefaults(pong)
			s.Send(int(login.GatePacketID_PACKET_SC_PONG), pong)
		}
		return nil
	}))
}
