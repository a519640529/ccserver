package easterisland

import (
	"games.yol.com/win88/common"
	"games.yol.com/win88/gamesrv/base"
	"games.yol.com/win88/protocol/easterisland"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
)

// 复活岛的操作
type CSEasterIslandOpPacketFactory struct {
}
type CSEasterIslandOpHandler struct {
}

func (this *CSEasterIslandOpPacketFactory) CreatePacket() interface{} {
	pack := &easterisland.CSEasterIslandOp{}
	return pack
}
func (this *CSEasterIslandOpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSEasterIslandOpHandler Process recv ", data)
	if csEasterIslandOp, ok := data.(*easterisland.CSEasterIslandOp); ok {
		p := base.PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSEasterIslandOpHandler p == nil")
			return nil
		}
		scene := p.GetScene()
		if scene == nil {
			logger.Logger.Warn("CSEasterIslandOpHandler p.scene == nil")
			return nil
		}
		if scene.GameId != common.GameId_EasterIsland {
			logger.Logger.Error("CSEasterIslandOpHandler gameId Error ", scene.GameId)
			return nil
		}
		if !scene.HasPlayer(p) {
			return nil
		}
		sp := scene.GetScenePolicy()
		if sp != nil {
			sp.OnPlayerOp(scene, p, int(csEasterIslandOp.GetOpCode()), csEasterIslandOp.GetParams())
		}
		return nil
	}
	return nil
}

func init() {
	// 复活岛的操作
	common.RegisterHandler(int(easterisland.EasterIslandPacketID_PACKET_CS_EASTERISLAND_PLAYEROP), &CSEasterIslandOpHandler{})
	netlib.RegisterFactory(int(easterisland.EasterIslandPacketID_PACKET_CS_EASTERISLAND_PLAYEROP), &CSEasterIslandOpPacketFactory{})
}
