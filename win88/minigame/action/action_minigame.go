package action

import (
	"fmt"
	//"games.yol.com/win88/common"
	"games.yol.com/win88/minigame/base"
	"games.yol.com/win88/protocol/mngame"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/core/profile"
	"github.com/idealeak/goserver/core/utils"
)

type CSMNGameDispatcherPacketFactory struct {
}
type CSMNGameDispatcherHandler struct {
}

func (this *CSMNGameDispatcherPacketFactory) CreatePacket() interface{} {
	pack := &mngame.CSMNGameDispatcher{}
	return pack
}

func (this *CSMNGameDispatcherHandler) Process(s *netlib.Session, packetid int, data interface{}) error {
	logger.Logger.Trace("CSMNGameDispatcherHandler Process recv ", data)
	if msg, ok := data.(*mngame.CSMNGameDispatcher); ok {
		scene := base.SceneMgrSington.GetScene(int(msg.Id))
		if scene != nil {
			p := scene.GetPlayer(msg.SnId)
			if p != nil {
				if !scene.HasPlayer(p) {
					return nil
				}
				packetid, packet, err := netlib.UnmarshalPacket(msg.Data)
				if err == nil {
					h := base.GetHandler(packetid)
					if h != nil {
						utils.DumpStackIfPanic(fmt.Sprintf("CSMNGameDispatcherHandler.Process error, packetid:%v", packetid))
						watch := profile.TimeStatisticMgr.WatchStart(fmt.Sprintf("/action/packet:%v", packetid), profile.TIME_ELEMENT_ACTION)
						err := h.Process(s, packetid, packet, scene, p)
						if watch != nil {
							watch.Stop()
						}
						if err != nil {
							logger.Logger.Tracef("Packet [%d] error:", packetid, err)
						}
						return err
					} else {
						logger.Logger.Tracef("Packet %v not find handler.", packetid)
					}
				} else {
					logger.Logger.Trace("CSMNGameDispatcherHandler process err:", err)
				}
			}
		}
	}

	return nil
}

func init() {
	netlib.RegisterHandler(int(mngame.MNGamePacketID_PACKET_CS_MNGAME_DISPATCHER), &CSMNGameDispatcherHandler{})
	netlib.RegisterFactory(int(mngame.MNGamePacketID_PACKET_CS_MNGAME_DISPATCHER), &CSMNGameDispatcherPacketFactory{})
}
