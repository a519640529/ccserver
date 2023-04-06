package tamquoc

import (
	"games.yol.com/win88/common"
	"games.yol.com/win88/gamesrv/base"
	"games.yol.com/win88/protocol/tamquoc"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
)

// 百战成神的操作
type CSTamQuocOpPacketFactory struct {
}
type CSTamQuocOpHandler struct {
}

func (this *CSTamQuocOpPacketFactory) CreatePacket() interface{} {
	pack := &tamquoc.CSTamQuocOp{}
	return pack
}
func (this *CSTamQuocOpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSTamQuocOpHandler Process recv ", data)
	if csTamQuocOp, ok := data.(*tamquoc.CSTamQuocOp); ok {
		p := base.PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSTamQuocOpHandler p == nil")
			return nil
		}
		scene := p.GetScene()
		if scene == nil {
			logger.Logger.Warn("CSTamQuocOpHandler p.scene == nil")
			return nil
		}
		if scene.GameId != common.GameId_TamQuoc {
			logger.Logger.Error("CSTamQuocOpHandler gameId Error ", scene.GameId)
			return nil
		}
		if !scene.HasPlayer(p) {
			return nil
		}
		sp := scene.GetScenePolicy()
		if sp != nil {
			sp.OnPlayerOp(scene, p, int(csTamQuocOp.GetOpCode()), csTamQuocOp.GetParams())
		}
		return nil
	}
	return nil
}

func init() {
	// 百战成神的操作
	common.RegisterHandler(int(tamquoc.TamQuocPacketID_PACKET_CS_TAMQUOC_PLAYEROP), &CSTamQuocOpHandler{})
	netlib.RegisterFactory(int(tamquoc.TamQuocPacketID_PACKET_CS_TAMQUOC_PLAYEROP), &CSTamQuocOpPacketFactory{})
}
