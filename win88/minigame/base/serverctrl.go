package base

import (
	"games.yol.com/win88/common"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/server"
	"github.com/idealeak/goserver/srvlib"
)

func init() {
	common.RegisteServerCtrlCallback(func(code int32) {
		switch code {
		case common.SrvCtrlStateSwitchCode:
			pack := &server.ServerStateSwitch{
				SrvType: proto.Int(common.GetSelfSrvType()),
				SrvId:   proto.Int(common.GetSelfSrvId()),
			}
			proto.SetDefaults(pack)
			srvlib.ServerSessionMgrSington.Broadcast(int(server.SSPacketID_PACKET_GB_STATE_SWITCH), pack, common.GetSelfAreaId(), srvlib.WorldServerType)
		}
	})
}
