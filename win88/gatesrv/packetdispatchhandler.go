package main

import (
	"bytes"
	"games.yol.com/win88/common"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/server"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/srvlib"
)

func init() {
	netlib.RegisteUnknowPacketHandlerCreator("packetdispatchhandler", func() netlib.UnknowPacketHandler {
		return netlib.UnknowPacketHandlerWrapper(func(s *netlib.Session, packetid int, logicNo uint32, data []byte) bool {
			if !s.Auth {
				logger.Logger.Trace("packetdispatchhandler session not auth! ")
				return false
			}
			var ss *netlib.Session
			if packetid >= 2000 && packetid < 5000 {
				worldSess := s.GetAttribute(common.ClientSessionAttribute_WorldServer)
				if worldSess == nil {
					ss = srvlib.ServerSessionMgrSington.GetSession(common.GetSelfAreaId(), srvlib.WorldServerType, common.GetWorldSrvId())
					s.SetAttribute(common.ClientSessionAttribute_WorldServer, ss)
				} else {
					ss = worldSess.(*netlib.Session)
					if ss == nil {
						ss = srvlib.ServerSessionMgrSington.GetSession(common.GetSelfAreaId(), srvlib.WorldServerType, common.GetWorldSrvId())
						s.SetAttribute(common.ClientSessionAttribute_WorldServer, ss)
					}
				}
			} else {
				gameSess := s.GetAttribute(common.ClientSessionAttribute_GameServer)
				if gameSess == nil {
					logger.Logger.Trace("packetdispatchhandler not found fit gamesession! ", packetid)
					return true
				}
				ss = gameSess.(*netlib.Session)
			}
			if ss == nil {
				logger.Logger.Trace("packetdispatchhandler redirect server session is nil ", packetid)
				return true
			}
			//must copy
			buf := bytes.NewBuffer(nil)
			buf.Write(data)
			pack := &server.SSTransmit{
				PacketData: buf.Bytes(),
			}
			param := s.GetAttribute(srvlib.SessionAttributeClientSession)
			if param != nil {
				if sid, ok := param.(srvlib.SessionId); ok {
					pack.SessionId = proto.Int64(sid.Get())
				}
			}
			proto.SetDefaults(pack)
			ss.Send(int(server.TransmitPacketID_PACKET_SS_PACKET_TRANSMIT), pack)
			return true
		})
	})
}
