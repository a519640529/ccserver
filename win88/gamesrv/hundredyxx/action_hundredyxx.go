package hundredyxx

import (
	"games.yol.com/win88/common"
	"games.yol.com/win88/gamesrv/base"
	"games.yol.com/win88/protocol/hundredyxx"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
)

/////////////////////////////////////////////////////
//鱼虾蟹
////////////////////////////////////////////////////
type CSHundredYXXOpPacketFactory struct {
}
type CSHundredYXXOpHandler struct {
}

func (this *CSHundredYXXOpPacketFactory) CreatePacket() interface{} {
	pack := &hundredyxx.CSHundredYXXOp{}
	return pack
}

func (this *CSHundredYXXOpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	if op, ok := data.(*hundredyxx.CSHundredYXXOp); ok {
		p := base.PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSHundredYXXOpHandler p == nil")
			return nil
		}
		scene := p.GetScene()
		sp := scene.GetScenePolicy()
		if scene == nil {
			logger.Logger.Warn("CSHundredYXXOpHandler p.scene == nil")
			return nil
		}
		if scene.GameId != common.GameId_HundredYXX {
			logger.Logger.Error("CSHundredYXXOpHandler gameId Error ", scene.GameId)
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
	// 鱼虾蟹
	common.RegisterHandler(int(hundredyxx.HundredYXXPacketID_PACKET_CS_HYXX_PLAYEROP), &CSHundredYXXOpHandler{})
	netlib.RegisterFactory(int(hundredyxx.HundredYXXPacketID_PACKET_CS_HYXX_PLAYEROP), &CSHundredYXXOpPacketFactory{})
}
