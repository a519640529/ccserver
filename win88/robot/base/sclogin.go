package base

import (
	"encoding/json"
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	login_proto "games.yol.com/win88/protocol/login"
	player_proto "games.yol.com/win88/protocol/player"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
)

func init() {
	//得到登录信息以后，获取角色信息
	netlib.RegisterHandler(int(login_proto.LoginPacketID_PACKET_SC_LOGIN),
		netlib.HandlerWrapper(func(s *netlib.Session, packetid int, data interface{}) error {
			logger.Logger.Trace("(this *SCLoginHandler) Process", data)
			if scLogin, ok := data.(*login_proto.SCLogin); ok {
				if scLogin.GetOpRetCode() == login_proto.OpResultCode_OPRC_Sucess {
					_, _, _, ip := RandRobotInfo(scLogin.GetAccId())
					csPlayerData := &player_proto.CSPlayerData{
						AccId: scLogin.AccId,
					}
					pp := &model.PlayerParams{
						Platform:     2,
						Ip:           ip,
						City:         RandZone(),
						Logininmodel: "app",
					}
					data, err := json.Marshal(pp)
					if err == nil {
						csPlayerData.Params = proto.String(string(data))
					}
					proto.SetDefaults(csPlayerData)
					s.Send(int(player_proto.PlayerPacketID_PACKET_CS_PLAYERDATA), csPlayerData)
					//定时器停掉
					StopSessionLoginTimer(s)
				} else {
					logger.Logger.Trace("Login failed,client seccion close.")
					s.Close()
				}
			}
			return nil
		}))
	netlib.RegisterFactory(int(login_proto.LoginPacketID_PACKET_SC_LOGIN), netlib.PacketFactoryWrapper(func() interface{} {
		return &login_proto.SCLogin{}
	}))
	//心跳协议
	netlib.RegisterHandler(int(login_proto.GatePacketID_PACKET_SC_PONG), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
		if _, ok := pack.(*login_proto.SCPong); ok {
			//logger.Logger.Trace("(this *SCPongHandler) Process", *scPong)
		}
		return nil
	}))
	netlib.RegisterFactory(int(login_proto.GatePacketID_PACKET_SC_PONG), netlib.PacketFactoryWrapper(func() interface{} {
		return &login_proto.SCPong{}
	}))
}
