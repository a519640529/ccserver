package common

import (
	"fmt"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/server"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/core/profile"
	"github.com/idealeak/goserver/core/utils"
	rawproto "google.golang.org/protobuf/proto"
)

var (
	TransmitMaker = &SSTransmitPacketFactory{}
)

type SSTransmitPacketFactory struct {
}

type SSTransmitHandler struct {
}

func (this *SSTransmitPacketFactory) CreatePacket() interface{} {
	pack := &server.SSTransmit{}
	return pack
}

func (this *SSTransmitPacketFactory) CreateTransmitPacket(packetid int, data interface{}, sid int64) (rawproto.Message, error) {
	pack := &server.SSTransmit{
		SessionId: sid,
	}
	if byteData, ok := data.([]byte); ok {
		pack.PacketData = byteData
		if byteData == nil || len(byteData) == 0 {
			logger.Logger.Info("SSTransmitPacketFactory.CreateTransmitPacket PacketData is empty")
		}
	} else {
		byteData, err := netlib.MarshalPacket(packetid, data)
		if err == nil {
			pack.PacketData = byteData
			if byteData == nil || len(byteData) == 0 {
				logger.Logger.Info("SSTransmitPacketFactory.CreateTransmitPacket PacketData is empty")
			}
		} else {
			logger.Logger.Info("SSTransmitPacketFactory.CreateTransmitPacket err:", err)
			return nil, err
		}
	}
	proto.SetDefaults(pack)
	return pack, nil
}

func (this *SSTransmitHandler) Process(s *netlib.Session, packetid int, data interface{}) error {
	//logger.Logger.Trace("SSTransmitHandler Process recv ", data)
	if transmitPack, ok := data.(*server.SSTransmit); ok {
		pd := transmitPack.GetPacketData()
		sid := transmitPack.GetSessionId()
		packetid, packet, err := netlib.UnmarshalPacket(pd)
		if err == nil {
			h := GetHandler(packetid)
			if h != nil {
				utils.DumpStackIfPanic(fmt.Sprintf("SSTransmitHandler.Process error, packetid:%v", packetid))
				watch := profile.TimeStatisticMgr.WatchStart(fmt.Sprintf("/action/packet:%v", packetid), profile.TIME_ELEMENT_ACTION)
				err := h.Process(s, packetid, packet, sid)
				if watch != nil {
					watch.Stop()
				}
				if err != nil {
					logger.Logger.Tracef("Packet [%d] error:", packetid, err)
				}
				return err
			} else {
				logger.Logger.Tracef("Packet %v not find handler.", packetid)
			}
		} else {
			logger.Logger.Trace("SSTransmitHandler process err:", err)
		}
	}
	return nil
}

func init() {
	netlib.RegisterHandler(int(server.TransmitPacketID_PACKET_SS_PACKET_TRANSMIT), &SSTransmitHandler{})
	netlib.RegisterFactory(int(server.TransmitPacketID_PACKET_SS_PACKET_TRANSMIT), TransmitMaker)
}
