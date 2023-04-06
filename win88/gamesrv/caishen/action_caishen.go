package caishen

import (
	"games.yol.com/win88/common"
	"games.yol.com/win88/gamesrv/base"
	"games.yol.com/win88/protocol/caishen"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
)

// 财神的操作
type CSCaiShenOpPacketFactory struct {
}
type CSCaiShenOpHandler struct {
}

func (this *CSCaiShenOpPacketFactory) CreatePacket() interface{} {
	pack := &caishen.CSCaiShenOp{}
	return pack
}
func (this *CSCaiShenOpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSCaiShenOpHandler Process recv ", data)
	if csCaiShenOp, ok := data.(*caishen.CSCaiShenOp); ok {
		p := base.PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSCaiShenOpHandler p == nil")
			return nil
		}
		scene := p.GetScene()
		if scene == nil {
			logger.Logger.Warn("CSCaiShenOpHandler p.scene == nil")
			return nil
		}
		if scene.GameId != common.GameId_CaiShen {
			logger.Logger.Error("CSCaiShenOpHandler gameId Error ", scene.GameId)
			return nil
		}
		if !scene.HasPlayer(p) {
			return nil
		}
		sp := scene.GetScenePolicy()
		if sp != nil {
			sp.OnPlayerOp(scene, p, int(csCaiShenOp.GetOpCode()), csCaiShenOp.GetParams())
		}
		return nil
	}
	return nil
}

func init() {
	// 财神的操作
	common.RegisterHandler(int(caishen.CaiShenPacketID_PACKET_CS_CAISHEN_PLAYEROP), &CSCaiShenOpHandler{})
	netlib.RegisterFactory(int(caishen.CaiShenPacketID_PACKET_CS_CAISHEN_PLAYEROP), &CSCaiShenOpPacketFactory{})
}
