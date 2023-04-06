package candy

import (
	"games.yol.com/win88/common"
	"games.yol.com/win88/minigame/base"
	"games.yol.com/win88/protocol/candy"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
)

// 糖果的操作
type CSCandyOpPacketFactory struct {
}

type CSCandyOpHandler struct {
}

func (this *CSCandyOpPacketFactory) CreatePacket() interface{} {
	pack := &candy.CSCandyOp{}
	return pack
}

func (this *CSCandyOpHandler) Process(s *netlib.Session, packetid int, data interface{}, scene *base.Scene, p *base.Player) error {
	logger.Logger.Trace("CSCandyOpHandler Process recv ", data)
	if csCandyOp, ok := data.(*candy.CSCandyOp); ok {
		if scene.GameId != common.GameId_Candy {
			logger.Logger.Error("CSCandyOpHandler gameId Error ", scene.GameId)
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
	// 糖果的操作
	base.RegisterHandler(int(candy.CandyPacketID_PACKET_CS_CANDY_PLAYEROP), &CSCandyOpHandler{})
	netlib.RegisterFactory(int(candy.CandyPacketID_PACKET_CS_CANDY_PLAYEROP), &CSCandyOpPacketFactory{})
}
