package base

import (
	hall_proto "games.yol.com/win88/protocol/gamehall"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
)

func init() {
	netlib.RegisterHandler(int(hall_proto.CoinSceneGamePacketID_PACKET_SC_COINSCENE_OP),
		netlib.HandlerWrapper(func(s *netlib.Session, packetid int, data interface{}) error {
			logger.Logger.Tracef("(this *SCCoinSceneOp) Process [%v].", s.GetSessionConfig().Id)
			if msg, ok := data.(*hall_proto.SCCoinSceneOp); ok {
				logger.Logger.Tracef("(this *SCCoinSceneOp) Data [%v].", msg)
				if msg.GetOpCode() == hall_proto.OpResultCode_OPRC_CoinSceneEnterQueueSucc {
					if !HadScene(s) {
						s.SetAttribute(SessionAttributeCoinSceneQueue, true)
					}
				} else {
					s.RemoveAttribute(SessionAttributeCoinSceneQueue)
				}
			} else {
				logger.Logger.Errorf("(this *SCCoinSceneOp) Data error [%v].", data)
			}
			return nil
		}))
	netlib.RegisterFactory(int(hall_proto.CoinSceneGamePacketID_PACKET_SC_COINSCENE_OP), netlib.PacketFactoryWrapper(func() interface{} {
		return &hall_proto.SCCoinSceneOp{}
	}))
}
