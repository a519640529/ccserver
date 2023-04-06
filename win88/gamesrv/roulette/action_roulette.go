package roulette

import (
	"games.yol.com/win88/common"
	"games.yol.com/win88/gamesrv/base"
	proto_roulette "games.yol.com/win88/protocol/roulette"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
)

//轮盘 Roulette
type CSRouletteOpPacketFactory struct {
}
type CSRouletteOpHandler struct {
}

func (this *CSRouletteOpPacketFactory) CreatePacket() interface{} {
	pack := &proto_roulette.CSRoulettePlayerOp{}
	return pack
}

func (this *CSRouletteOpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSRouletteOpHandler Process recv ", data)
	if op, ok := data.(*proto_roulette.CSRoulettePlayerOp); ok {
		p := base.PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSRouletteOpHandler p == nil")
			return nil
		}
		scene := p.GetScene()
		if scene == nil {
			logger.Logger.Warn("CSRouletteOpHandler p.scene == nil")
			return nil
		}
		if scene.GameId != common.GameId_Roulette {
			logger.Logger.Error("CSRouletteOpHandler GameId Error ", scene.GameId)
			return nil
		}
		if !scene.HasPlayer(p) {
			return nil
		}
		if scene.GetScenePolicy() != nil {
			scene.GetScenePolicy().OnPlayerOp(scene, p, int(op.GetOpCode()), op.GetOpParam())
		}
		return nil
	}
	return nil
}

func init() {
	// 轮盘
	common.RegisterHandler(int(proto_roulette.RouletteMmoPacketID_PACKET_CS_Roulette_PlayerOp), &CSRouletteOpHandler{})
	netlib.RegisterFactory(int(proto_roulette.RouletteMmoPacketID_PACKET_CS_Roulette_PlayerOp), &CSRouletteOpPacketFactory{})
}
