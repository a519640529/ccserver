package redvsblack

import (
	"games.yol.com/win88/common"
	"games.yol.com/win88/gamesrv/base"
	"games.yol.com/win88/protocol/redvsblack"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
)

// 龙虎斗的操作

type CSRedVsBlackOpPacketFactory struct {
}
type CSRedVsBlackOpHandler struct {
}

func (this *CSRedVsBlackOpPacketFactory) CreatePacket() interface{} {
	pack := &redvsblack.CSRedVsBlackOp{}
	return pack
}
func (this *CSRedVsBlackOpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSRedVsBlackOpHandler Process recv ", data)
	if CSRedVsBlackOp, ok := data.(*redvsblack.CSRedVsBlackOp); ok {
		p := base.PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSRedVsBlackOpHandler p == nil")
			return nil
		}
		scene := p.GetScene()
		if scene == nil {
			logger.Logger.Warn("CSRedVsBlackOpHandler p.scene == nil")
			return nil
		}
		if scene.GameId != common.GameId_RedVsBlack {
			logger.Logger.Error("CSRedVsBlackOpHandler gameId Error ", scene.GameId)
			return nil
		}
		if !scene.HasPlayer(p) {
			return nil
		}
		sp := scene.GetScenePolicy()
		if sp != nil {
			sp.OnPlayerOp(scene, p, int(CSRedVsBlackOp.GetOpCode()), CSRedVsBlackOp.GetParams())
		}
		return nil
	}
	return nil
}

func init() {
	common.RegisterHandler(int(redvsblack.RedVsBlackPacketID_PACKET_CS_RVSB_PLAYEROP), &CSRedVsBlackOpHandler{})
	netlib.RegisterFactory(int(redvsblack.RedVsBlackPacketID_PACKET_CS_RVSB_PLAYEROP), &CSRedVsBlackOpPacketFactory{})
}
