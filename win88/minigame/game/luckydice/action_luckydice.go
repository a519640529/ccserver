package luckydice

import (
	"games.yol.com/win88/common"
	"games.yol.com/win88/minigame/base"
	"games.yol.com/win88/protocol/luckydice"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
)

type CSLuckyDiceOpPacketFactory struct {
}

type CSLuckyDiceOpHandler struct {
}

func (this *CSLuckyDiceOpPacketFactory) CreatePacket() interface{} {
	pack := &luckydice.CSLuckyDiceOp{}
	return pack
}

func (this *CSLuckyDiceOpHandler) Process(s *netlib.Session, packetid int, data interface{}, scene *base.Scene, p *base.Player) error {
	logger.Logger.Trace("CSLuckyDiceOpHandler Process recv ", data)
	if csCandyOp, ok := data.(*luckydice.CSLuckyDiceOp); ok {
		if scene.GameId != common.GameId_LuckyDice {
			logger.Logger.Error("CSLuckyDiceOpHandler gameId Error ", scene.GameId)
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
	base.RegisterHandler(int(luckydice.LuckyDicePacketID_PACKET_CS_LUCKYDICE_PLAYEROP), &CSLuckyDiceOpHandler{})
	netlib.RegisterFactory(int(luckydice.LuckyDicePacketID_PACKET_CS_LUCKYDICE_PLAYEROP), &CSLuckyDiceOpPacketFactory{})
}
