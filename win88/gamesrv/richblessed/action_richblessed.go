package richblessed

import (
	"games.yol.com/win88/common"
	"games.yol.com/win88/gamesrv/base"
	"games.yol.com/win88/protocol/richblessed"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
)

type CSRichBlessedOpPacketFactory struct {
}
type CSRichBlessedOpHandler struct {
}

func (this *CSRichBlessedOpPacketFactory) CreatePacket() interface{} {
	pack := &richblessed.CSRichBlessedOp{}
	return pack
}

func (this *CSRichBlessedOpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	if op, ok := data.(*richblessed.CSRichBlessedOp); ok {
		p := base.PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSRichBlessedOpHandler p == nil")
			return nil
		}
		scene := p.GetScene()
		if scene == nil {
			logger.Logger.Warn("CSRichBlessedOpHandler p.scene == nil")
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
	//多财多福
	common.RegisterHandler(int(richblessed.RBPID_PACKET_RICHBLESSED_CSRichBlessedOp), &CSRichBlessedOpHandler{})
	netlib.RegisterFactory(int(richblessed.RBPID_PACKET_RICHBLESSED_CSRichBlessedOp), &CSRichBlessedOpPacketFactory{})
}
