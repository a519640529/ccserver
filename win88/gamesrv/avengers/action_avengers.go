package avengers

import (
	"games.yol.com/win88/common"
	"games.yol.com/win88/gamesrv/base"
	"games.yol.com/win88/protocol/avengers"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
)

// 复仇者联盟的操作
type CSAvengersOpPacketFactory struct {
}
type CSAvengersOpHandler struct {
}

func (this *CSAvengersOpPacketFactory) CreatePacket() interface{} {
	pack := &avengers.CSAvengersOp{}
	return pack
}
func (this *CSAvengersOpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSAvengersOpHandler Process recv ", data)
	if csAvengersOp, ok := data.(*avengers.CSAvengersOp); ok {
		p := base.PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSAvengersOpHandler p == nil")
			return nil
		}
		scene := p.GetScene()
		if scene == nil {
			logger.Logger.Warn("CSAvengersOpHandler p.scene == nil")
			return nil
		}
		if scene.GameId != common.GameId_Avengers {
			logger.Logger.Error("CSAvengersOpHandler gameId Error ", scene.GameId)
			return nil
		}
		if !scene.HasPlayer(p) {
			return nil
		}
		sp := scene.GetScenePolicy()
		if sp != nil {
			sp.OnPlayerOp(scene, p, int(csAvengersOp.GetOpCode()), csAvengersOp.GetParams())
		}
		return nil
	}
	return nil
}
func init() {
	// 复仇者联盟的操作
	common.RegisterHandler(int(avengers.AvengersPacketID_PACKET_CS_AVENGERS_PLAYEROP), &CSAvengersOpHandler{})
	netlib.RegisterFactory(int(avengers.AvengersPacketID_PACKET_CS_AVENGERS_PLAYEROP), &CSAvengersOpPacketFactory{})
}
