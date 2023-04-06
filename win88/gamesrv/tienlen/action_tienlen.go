package tienlen

import (
	"games.yol.com/win88/common"
	"games.yol.com/win88/gamesrv/base"
	"games.yol.com/win88/protocol/tienlen"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
)

// tienlen
type CSTienLenPlayerOpPacketFactory struct {
}
type CSTienLenPlayerOpHandler struct {
}

func (this *CSTienLenPlayerOpPacketFactory) CreatePacket() interface{} {
	pack := &tienlen.CSTienLenPlayerOp{}
	return pack
}
func (this *CSTienLenPlayerOpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSTienLenPlayerOpHandler Process recv ", data)
	if msg, ok := data.(*tienlen.CSTienLenPlayerOp); ok {
		p := base.PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSTienLenPlayerOpHandler p == nil")
			return nil
		}
		scene := p.GetScene()
		sp := scene.GetScenePolicy()
		if scene == nil {
			logger.Logger.Warn("CSTienLenPlayerOpHandler p.scene == nil")
			return nil
		}
		if !scene.IsTienLen() {
			logger.Logger.Error("CSTienLenPlayerOpHandler gameId Error ", scene.GameId)
			return nil
		}
		if !scene.HasPlayer(p) {
			return nil
		}
		if sp != nil {
			sp.OnPlayerOp(scene, p, int(msg.GetOpCode()), msg.GetOpParam())
		}
		return nil
	}
	return nil
}

func init() {
	//tienlen
	common.RegisterHandler(int(tienlen.TienLenPacketID_PACKET_CSTienLenPlayerOp), &CSTienLenPlayerOpHandler{})
	netlib.RegisterFactory(int(tienlen.TienLenPacketID_PACKET_CSTienLenPlayerOp), &CSTienLenPlayerOpPacketFactory{})
}
