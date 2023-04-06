package common

import (
	"games.yol.com/win88/proto"
	protocol_game "games.yol.com/win88/protocol/server"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/srvlib"
	"github.com/idealeak/goserver/srvlib/protocol"
	rawproto "google.golang.org/protobuf/proto"
)

type WGCoinSceneChange struct {
	SnId    int32
	SceneId int32
}

type WGAddCoin struct {
	SnId      int32
	Coin      int64
	GainWay   int32
	Oper      string
	Remark    string
	Broadcast bool
	WriteLog  bool
}

func createMulticastPacket(packetid int, data interface{}, sis ...*protocol.MCSessionUnion) (rawproto.Message, error) {
	pack := &protocol.SSPacketMulticast{
		Sessions: sis,
		PacketId: proto.Int(packetid),
	}
	if byteData, ok := data.([]byte); ok {
		pack.Data = byteData
	} else {
		byteData, err := netlib.MarshalPacket(packetid, data)
		if err == nil {
			pack.Data = byteData
		} else {
			logger.Logger.Info("MulticastPacketFactory.CreateMulticastPacket err:", err)
			return nil, err
		}
	}
	proto.SetDefaults(pack)
	return pack, nil
}

func SendToGate(sid int64, packetid int, rawpack interface{}, s *netlib.Session) bool {
	if s == nil || rawpack == nil || sid == 0 {
		return false
	}
	pack, err := createMulticastPacket(packetid, rawpack,
		&protocol.MCSessionUnion{
			Mccs: &protocol.MCClientSession{
				SId: proto.Int64(sid)}})
	if err == nil && pack != nil {
		if d, err := netlib.MarshalPacket(int(protocol.SrvlibPacketID_PACKET_SS_MULTICAST), pack); err == nil {
			return s.Send(int(protocol.SrvlibPacketID_PACKET_SS_MULTICAST), d, true)
		} else {
			logger.Logger.Warn("SendToGate err:", err)
		}
	}
	return false
}

func SendToActThrSrv(packetid int, rawpack interface{}) bool {
	if rawpack == nil {
		return false
	}

	replaySess := srvlib.ServerSessionMgrSington.GetSession(GetSelfAreaId(), ActThrServerType, ActThrServerID)
	if replaySess != nil {
		return replaySess.Send(int(packetid), rawpack)
	}
	return false
}

func TransmitToServer(sid int64, packetid int, rawpack interface{}, s *netlib.Session) bool {
	if d, err := netlib.MarshalPacket(packetid, rawpack); err == nil {
		pack := &protocol_game.SSTransmit{
			PacketData: d,
			SessionId:  sid,
		}
		proto.SetDefaults(pack)
		return s.Send(int(protocol_game.TransmitPacketID_PACKET_SS_PACKET_TRANSMIT), pack, true)
	} else {
		logger.Logger.Warn("TransmitToServer err:", err)
	}
	return false
}
