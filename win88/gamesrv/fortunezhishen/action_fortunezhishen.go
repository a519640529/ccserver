package fortunezhishen

import (
	"games.yol.com/win88/common"
	"games.yol.com/win88/gamesrv/base"
	"games.yol.com/win88/protocol/fortunezhishen"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
)

type CSFortuneZhiShenOpPacketFactory struct {
}
type CSFortuneZhiShenOpHandler struct {
}

func (this *CSFortuneZhiShenOpPacketFactory) CreatePacket() interface{} {
	pack := &fortunezhishen.CSFortuneZhiShenOp{}
	return pack
}

func (this *CSFortuneZhiShenOpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	if op, ok := data.(*fortunezhishen.CSFortuneZhiShenOp); ok {
		p := base.PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSFortuneZhiShenOpHandler p == nil")
			return nil
		}
		scene := p.GetScene()
		if scene == nil {
			logger.Logger.Warn("CSFortuneZhiShenOpHandler p.scene == nil")
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
	//财运之神
	common.RegisterHandler(int(fortunezhishen.FortuneZSPacketID_PACKET_CS_FORTUNEZHISHEN_PLAYEROP), &CSFortuneZhiShenOpHandler{})
	netlib.RegisterFactory(int(fortunezhishen.FortuneZSPacketID_PACKET_CS_FORTUNEZHISHEN_PLAYEROP), &CSFortuneZhiShenOpPacketFactory{})
}
