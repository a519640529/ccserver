package main

import (
	"games.yol.com/win88/common"
	"games.yol.com/win88/protocol/welfare"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
)

// 救济金
type CSGetReliefFundPacketFactory struct {
}

type CSGetReliefFundHandler struct {
}

func (this *CSGetReliefFundPacketFactory) CreatePacket() interface{} {
	pack := &welfare.CSGetReliefFund{}
	return pack
}

func (this *CSGetReliefFundHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSGetReliefFund Process recv ", data)
	if _, ok := data.(*welfare.CSGetReliefFund); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warnf("CSGetReliefFundHandler p == nil p.SnId = %v", p.SnId)
			return nil
		}
		WelfareMgrSington.GetReliefFund(p)
	}
	return nil
}

func init() {

	//领取救济金
	common.RegisterHandler(int(welfare.SPacketID_PACKET_CS_WELF_GETRELIEFFUND), &CSGetReliefFundHandler{})
	netlib.RegisterFactory(int(welfare.SPacketID_PACKET_CS_WELF_GETRELIEFFUND), &CSGetReliefFundPacketFactory{})
}
