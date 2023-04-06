package caothap

import (
	"games.yol.com/win88/common"
	"games.yol.com/win88/minigame/base"
	"games.yol.com/win88/protocol/caothap"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
)

type CSCaoThapOpPacketFactory struct {
}

type CSCaoThapOpHandler struct {
}

func (this *CSCaoThapOpPacketFactory) CreatePacket() interface{} {
	pack := &caothap.CSCaoThapOp{}
	return pack
}

func (this *CSCaoThapOpHandler) Process(s *netlib.Session, packetid int, data interface{}, scene *base.Scene, p *base.Player) error {
	logger.Logger.Trace("CSCaoThapOpHandler Process recv ", data)
	if csCandyOp, ok := data.(*caothap.CSCaoThapOp); ok {
		if scene.GameId != common.GameId_CaoThap {
			logger.Logger.Error("CSCaoThapOpHandler gameId Error ", scene.GameId)
			return nil
		}
		sp := scene.GetScenePolicy()
		if sp != nil {
			sp.OnPlayerOp(scene, p, int(csCandyOp.GetOpCode()), csCandyOp.GetParams())
		}
		return nil
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////
func init() {
	base.RegisterHandler(int(caothap.CaoThapPacketID_PACKET_CS_CAOTHAP_PLAYEROP), &CSCaoThapOpHandler{})
	netlib.RegisterFactory(int(caothap.CaoThapPacketID_PACKET_CS_CAOTHAP_PLAYEROP), &CSCaoThapOpPacketFactory{})
}
