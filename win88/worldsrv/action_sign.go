package main

import (
	"games.yol.com/win88/common"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/activity"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
)

// ------------------------------------------------
type CSSignPacketFactory struct {
}

type CSSignHandler struct {
}

func (this *CSSignPacketFactory) CreatePacket() interface{} {
	pack := &activity.CSSign{}
	return pack
}

func (this *CSSignHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSSignHandler Process recv ", data)
	if msg, ok := data.(*activity.CSSign); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSSignHandler p == nil")
			return nil
		}

		pack := &activity.SCSign{}
		pack.SignIndex = proto.Int32(msg.GetSignIndex())
		pack.SignType = proto.Int32(msg.GetSignType())

		retCode := ActSignMgrSington.CanSign(p, int(msg.GetSignIndex()))
		if retCode != activity.OpResultCode_ActSign_OPRC_Activity_Sign_Sucess {
			pack.OpRetCode = retCode
			proto.SetDefaults(pack)
			p.SendToClient(int(activity.ActSignPacketID_PACKET_SCSign), pack)
			return nil
		}
		retCode = ActSignMgrSington.Sign(p, int(msg.GetSignIndex()), msg.GetSignType())
		if retCode != activity.OpResultCode_ActSign_OPRC_Activity_Sign_Sucess {
			pack.OpRetCode = retCode
			proto.SetDefaults(pack)
			p.SendToClient(int(activity.ActSignPacketID_PACKET_SCSign), pack)
			return nil
		}

		pack.OpRetCode = activity.OpResultCode_ActSign_OPRC_Activity_Sign_Sucess
		proto.SetDefaults(pack)
		p.SendToClient(int(activity.ActSignPacketID_PACKET_SCSign), pack)
	}
	return nil
}

// ------------------------------------------------
type CSSignDataPacketFactory struct {
}

type CSSignDataHandler struct {
}

func (this *CSSignDataPacketFactory) CreatePacket() interface{} {
	pack := &activity.CSSignData{}
	return pack
}

func (this *CSSignDataHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSSignDataHandler Process recv ", data)
	if _, ok := data.(*activity.CSSignData); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSSignDataHandler p == nil")
			return nil
		}
		ActSignMgrSington.SendSignDataToPlayer(p)
	}
	return nil
}

func init() {
	//签到
	common.RegisterHandler(int(activity.ActSignPacketID_PACKET_CSSign), &CSSignHandler{})
	netlib.RegisterFactory(int(activity.ActSignPacketID_PACKET_CSSign), &CSSignPacketFactory{})
	//签到数据
	common.RegisterHandler(int(activity.ActSignPacketID_PACKET_CSSignData), &CSSignDataHandler{})
	netlib.RegisterFactory(int(activity.ActSignPacketID_PACKET_CSSignData), &CSSignDataPacketFactory{})
}
