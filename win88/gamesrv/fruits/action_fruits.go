package fruits

import (
	"games.yol.com/win88/common"
	"games.yol.com/win88/gamesrv/base"
	"games.yol.com/win88/protocol/fruits"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
)

type CSFruitsOpPacketFactory struct {
}
type CSFruitsOpHandler struct {
}

func (this *CSFruitsOpPacketFactory) CreatePacket() interface{} {
	pack := &fruits.CSFruitsOp{}
	return pack
}

func (this *CSFruitsOpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	if op, ok := data.(*fruits.CSFruitsOp); ok {
		p := base.PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSFruitsOpHandler p == nil")
			return nil
		}
		scene := p.GetScene()
		if scene == nil {
			logger.Logger.Warn("CSFruitsOpHandler p.scene == nil")
			return nil
		}
		if !scene.HasPlayer(p) {
			return nil
		}
		if scene.GetScenePolicy() != nil {
			scene.GetScenePolicy().OnPlayerOp(scene, p, int(op.GetOpCode()), op.GetParams())
		}
		return nil
	}
	return nil
}
func init() {
	//水果拉霸
	common.RegisterHandler(int(fruits.FruitsPID_PACKET_FRUITS_CSFruitsOp), &CSFruitsOpHandler{})
	netlib.RegisterFactory(int(fruits.FruitsPID_PACKET_FRUITS_CSFruitsOp), &CSFruitsOpPacketFactory{})
}
