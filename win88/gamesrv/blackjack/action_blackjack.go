package blackjack

import (
	"games.yol.com/win88/common"
	"games.yol.com/win88/gamesrv/base"
	"games.yol.com/win88/protocol/blackjack"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
)

// 21点操作
type CSBlackJackOpPacketFactory struct {
}
type CSBlackJackOpHandler struct {
}

func (this *CSBlackJackOpPacketFactory) CreatePacket() interface{} {
	pack := &blackjack.CSBlackJackOP{}
	return pack
}

func (this *CSBlackJackOpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	////logger.Logger.Trace("CSBlackJackOpHandler Process recv ", data)
	if op, ok := data.(*blackjack.CSBlackJackOP); ok {
		p := base.PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSBlackJackOpHandler p == nil")
			return nil
		}
		scene := p.GetScene()
		sp := scene.GetScenePolicy()
		if scene == nil {
			logger.Logger.Warn("CSBlackJackOpHandler p.scene == nil")
			return nil
		}
		if scene.GameId != common.GameId_BlackJack {
			logger.Logger.Error("CSBlackJackOpHandler gameId Error ", scene.GameId)
			return nil
		}
		if !scene.HasPlayer(p) && !scene.HasAudience(p) {
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
	// 21点
	common.RegisterHandler(int(blackjack.BlackJackPacketID_CS_PLAYER_OPERATE), &CSBlackJackOpHandler{})
	netlib.RegisterFactory(int(blackjack.BlackJackPacketID_CS_PLAYER_OPERATE), &CSBlackJackOpPacketFactory{})
}
