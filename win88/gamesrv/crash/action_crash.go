package crash

import (
	"games.yol.com/win88/common"
	"games.yol.com/win88/gamesrv/base"
	"games.yol.com/win88/protocol/crash"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
)

// 碰撞的操作

type CSCrashOpPacketFactory struct {
}
type CSCrashOpHandler struct {
}

func (this *CSCrashOpPacketFactory) CreatePacket() interface{} {
	pack := &crash.CSCrashOp{}
	return pack
}
func (this *CSCrashOpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSCrashOpHandler Process recv ", data)
	if CSCrashOp, ok := data.(*crash.CSCrashOp); ok {
		p := base.PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSCrashOpHandler p == nil")
			return nil
		}
		scene := p.GetScene()
		if scene == nil {
			logger.Logger.Warn("CSCrashOpHandler p.scene == nil")
			return nil
		}
		if scene.GameId != common.GameId_Crash {
			logger.Logger.Error("CSCrashOpHandler gameId Error ", scene.GameId)
			return nil
		}
		if !scene.HasPlayer(p) {
			return nil
		}
		sp := scene.GetScenePolicy()
		if sp != nil {
			sp.OnPlayerOp(scene, p, int(CSCrashOp.GetOpCode()), CSCrashOp.GetParams())
		}
		return nil
	}
	return nil
}

func init() {
	common.RegisterHandler(int(crash.CrashPacketID_PACKET_CS_CRASH_PLAYEROP), &CSCrashOpHandler{})
	netlib.RegisterFactory(int(crash.CrashPacketID_PACKET_CS_CRASH_PLAYEROP), &CSCrashOpPacketFactory{})
}
