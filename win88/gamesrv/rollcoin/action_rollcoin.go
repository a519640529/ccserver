package rollcoin

import (
	"games.yol.com/win88/common"
	"games.yol.com/win88/gamesrv/base"
	"games.yol.com/win88/protocol/rollcoin"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
)

//奔驰宝马

type CSRollCoinOpPacketFactory struct {
}
type CSRollCoinOpHandler struct {
}

func (this *CSRollCoinOpPacketFactory) CreatePacket() interface{} {
	pack := &rollcoin.CSRollCoinOp{}
	return pack
}

func (this *CSRollCoinOpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	if op, ok := data.(*rollcoin.CSRollCoinOp); ok {
		p := base.PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSRollCoinOpHandler p == nil")
			return nil
		}
		scene := p.GetScene()
		sp := scene.GetScenePolicy()
		if scene == nil {
			logger.Logger.Warn("CSRollCoinOpHandler p.scene == nil")
			return nil
		}
		if scene.GameId != common.GameId_RollCoin {
			logger.Logger.Error("CSRollCoinOpHandler gameId Error ", scene.GameId)
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
	common.RegisterHandler(int(rollcoin.PACKETID_ROLLCOIN_CS_OP), &CSRollCoinOpHandler{})
	netlib.RegisterFactory(int(rollcoin.PACKETID_ROLLCOIN_CS_OP), &CSRollCoinOpPacketFactory{})
}
