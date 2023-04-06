package dragonvstiger

import (
	"games.yol.com/win88/common"
	"games.yol.com/win88/gamesrv/base"
	"games.yol.com/win88/protocol/dragonvstiger"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
)

// 龙虎斗的操作

type CSDragonVsTiggerOpPacketFactory struct {
}
type CSDragonVsTiggerOpHandler struct {
}

func (this *CSDragonVsTiggerOpPacketFactory) CreatePacket() interface{} {
	pack := &dragonvstiger.CSDragonVsTiggerOp{}
	return pack
}
func (this *CSDragonVsTiggerOpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSDragonVsTiggerOpHandler Process recv ", data)
	if CSDragonVsTiggerOp, ok := data.(*dragonvstiger.CSDragonVsTiggerOp); ok {
		p := base.PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSDragonVsTiggerOpHandler p == nil")
			return nil
		}
		scene := p.GetScene()
		if scene == nil {
			logger.Logger.Warn("CSDragonVsTiggerOpHandler p.scene == nil")
			return nil
		}
		if scene.GameId != common.GameId_DragonVsTiger {
			logger.Logger.Error("CSDragonVsTiggerOpHandler gameId Error ", scene.GameId)
			return nil
		}
		if !scene.HasPlayer(p) {
			return nil
		}
		sp := scene.GetScenePolicy()
		if sp != nil {
			sp.OnPlayerOp(scene, p, int(CSDragonVsTiggerOp.GetOpCode()), CSDragonVsTiggerOp.GetParams())
		}
		return nil
	}
	return nil
}

func init() {
	common.RegisterHandler(int(dragonvstiger.DragonVsTigerPacketID_PACKET_CS_DVST_PLAYEROP), &CSDragonVsTiggerOpHandler{})
	netlib.RegisterFactory(int(dragonvstiger.DragonVsTigerPacketID_PACKET_CS_DVST_PLAYEROP), &CSDragonVsTiggerOpPacketFactory{})
}
