package iceage

import (
	"games.yol.com/win88/common"
	"games.yol.com/win88/gamesrv/base"
	"games.yol.com/win88/protocol/iceage"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
)

// 冰河世纪的操作
type CSIceAgeOpPacketFactory struct {
}
type CSIceAgeOpHandler struct {
}

func (this *CSIceAgeOpPacketFactory) CreatePacket() interface{} {
	pack := &iceage.CSIceAgeOp{}
	return pack
}
func (this *CSIceAgeOpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSIceAgeOpHandler Process recv ", data)
	if csIceAgeOp, ok := data.(*iceage.CSIceAgeOp); ok {
		p := base.PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSIceAgeOpHandler p == nil")
			return nil
		}
		scene := p.GetScene()
		sp := scene.GetScenePolicy()
		if scene == nil {
			logger.Logger.Warn("CSIceAgeOpHandler p.scene == nil")
			return nil
		}
		if scene.GameId != common.GameId_IceAge {
			logger.Logger.Error("CSIceAgeOpHandler gameId Error ", scene.GameId)
			return nil
		}
		if !scene.HasPlayer(p) {
			return nil
		}
		if sp != nil {
			sp.OnPlayerOp(scene, p, int(csIceAgeOp.GetOpCode()), csIceAgeOp.GetParams())
		}
		return nil
	}
	return nil
}

func init() {
	// 冰河世纪的操作
	common.RegisterHandler(int(iceage.IceAgePacketID_PACKET_CS_ICEAGE_PLAYEROP), &CSIceAgeOpHandler{})
	netlib.RegisterFactory(int(iceage.IceAgePacketID_PACKET_CS_ICEAGE_PLAYEROP), &CSIceAgeOpPacketFactory{})
}
