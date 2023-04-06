package baccarat

import (
	"games.yol.com/win88/common"
	"games.yol.com/win88/gamesrv/base"
	proto_baccarat "games.yol.com/win88/protocol/baccarat"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
)

//百家乐
type CSBaccaratOpPacketFactory struct {
}
type CSBaccaratOpHandler struct {
}

func (this *CSBaccaratOpPacketFactory) CreatePacket() interface{} {
	pack := &proto_baccarat.CSBaccaratOp{}
	return pack
}

func (this *CSBaccaratOpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	//logger.Logger.Trace("CSDragonVsTigerOpHandler Process recv ", data)
	if csBaccaratOp, ok := data.(*proto_baccarat.CSBaccaratOp); ok {
		p := base.PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSBaccaratOpHandler p == nil")
			return nil
		}
		scene := p.GetScene()
		if scene == nil {
			logger.Logger.Warn("CSBaccaratOpHandler p.scene == nil")
			return nil
		}
		if scene.GameId != common.GameId_Baccarat {
			logger.Logger.Error("CSBaccaratOpHandler gameId Error ", scene.GameId)
			return nil
		}
		if !scene.HasPlayer(p) {
			return nil
		}
		if scene.GetScenePolicy() != nil {
			scene.GetScenePolicy().OnPlayerOp(scene, p, int(csBaccaratOp.GetOpCode()), csBaccaratOp.GetParams())
		}
		return nil
	}
	return nil
}
func init() {
	//百家乐
	common.RegisterHandler(int(proto_baccarat.BaccaratPacketID_PACKET_CS_BACCARAT_PLAYEROP), &CSBaccaratOpHandler{})
	netlib.RegisterFactory(int(proto_baccarat.BaccaratPacketID_PACKET_CS_BACCARAT_PLAYEROP), &CSBaccaratOpPacketFactory{})
}
