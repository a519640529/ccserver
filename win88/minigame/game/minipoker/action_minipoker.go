package minipoker

import (
	"games.yol.com/win88/common"
	"games.yol.com/win88/minigame/base"
	"games.yol.com/win88/protocol/minipoker"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
)

type CSMiniPokerOpPacketFactory struct {
}

type CSMiniPokerOpHandler struct {
}

func (this *CSMiniPokerOpPacketFactory) CreatePacket() interface{} {
	pack := &minipoker.CSMiniPokerOp{}
	return pack
}

func (this *CSMiniPokerOpHandler) Process(s *netlib.Session, packetid int, data interface{}, scene *base.Scene, p *base.Player) error {
	logger.Logger.Trace("CSMiniPokerOpHandler Process recv ", data)
	if csOp, ok := data.(*minipoker.CSMiniPokerOp); ok {
		if scene.GameId != common.GameId_MiniPoker {
			logger.Logger.Error("CSMiniPokerOpHandler gameId Error ", scene.GameId)
			return nil
		}
		sp := scene.GetScenePolicy()
		if sp != nil {
			sp.OnPlayerOp(scene, p, int(csOp.GetOpCode()), csOp.GetParams())
		}
		return nil
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////
func init() {
	base.RegisterHandler(int(minipoker.MiniPokerPacketID_PACKET_CS_MINIPOKER_PLAYEROP), &CSMiniPokerOpHandler{})
	netlib.RegisterFactory(int(minipoker.MiniPokerPacketID_PACKET_CS_MINIPOKER_PLAYEROP), &CSMiniPokerOpPacketFactory{})
}
