package rollpoint

import (
	"games.yol.com/win88/common"
	"games.yol.com/win88/gamesrv/base"
	"games.yol.com/win88/protocol/rollpoint"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
)

type CSRollPointPacketFactory struct {
}
type CSRollPointHandler struct {
}

func (this *CSRollPointPacketFactory) CreatePacket() interface{} {
	pack := &rollpoint.CSRollPointOp{}
	return pack
}

func (this *CSRollPointHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	if op, ok := data.(*rollpoint.CSRollPointOp); ok {
		p := base.PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSRollPointHandler p == nil")
			return nil
		}
		scene := p.GetScene()
		sp := scene.GetScenePolicy()
		if scene == nil {
			logger.Logger.Warn("CSRollPointHandler p.scene == nil")
			return nil
		}
		if scene.GameId != common.GameId_RollPoint {
			logger.Logger.Error("CSRollPointHandler gameId Error ", scene.GameId)
			return nil
		}
		if !scene.HasPlayer(p) {
			return nil
		}
		if sp != nil {
			sp.OnPlayerOp(scene, p, int(op.GetOpCode()), op.GetParams())
		}
		return nil
	}
	return nil
}

func init() {
	// 骰宝
	common.RegisterHandler(int(rollpoint.RPPACKETID_ROLLPOINT_CS_OP), &CSRollPointHandler{})
	netlib.RegisterFactory(int(rollpoint.RPPACKETID_ROLLPOINT_CS_OP), &CSRollPointPacketFactory{})
}
