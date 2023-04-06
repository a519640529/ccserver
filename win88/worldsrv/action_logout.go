package main

import (
	"games.yol.com/win88/common"
	login_proto "games.yol.com/win88/protocol/login"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"time"
)

func SessionLogout(sid int64, drop bool) bool {
	ls := LoginStateMgrSington.GetLoginStateOfSid(sid)
	if ls == nil {
		logger.Logger.Trace("SessionLogout ls == nil")
		return false
	}

	ls.state = LoginState_Logouting
	p := PlayerMgrSington.GetPlayer(sid)
	if p != nil {
		p.ThirdGameLogout()
		if drop {
			p.DropLine()
		} else {
			p.Logout()
		}
	} else {
		logger.Logger.Trace("SessionLogout p == nil")
	}

	if ls.als != nil && ls.als.acc != nil {
		ls.als.acc.LastLogoutTime = time.Now()
	}

	ls.state = LoginState_Logouted
	LoginStateMgrSington.Logout(ls)
	return true
}

type CSLogoutPacketFactory struct {
}

type CSLogoutHandler struct {
}

func (this *CSLogoutPacketFactory) CreatePacket() interface{} {
	pack := &login_proto.CSLogout{}
	return pack
}

func (this *CSLogoutHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {

	logger.Logger.Trace("CSLogoutHandler Process recv ", data)
	SessionLogout(sid, false)
	return nil
}

type SSDisconnectPacketFactory struct {
}

func (this *SSDisconnectPacketFactory) CreatePacket() interface{} {
	pack := &login_proto.SSDisconnect{}
	return pack
}

type SSDisconnectHandler struct {
}

func (this *SSDisconnectHandler) Process(s *netlib.Session, packetid int, data interface{}) error {
	logger.Logger.Trace("SSDisconnectHandler Process recv ", data)
	if ssd, ok := data.(*login_proto.SSDisconnect); ok {
		sid := ssd.GetSessionId()
		if sid != 0 {
			SessionLogout(sid, true)
		}
	}
	return nil
}
func init() {
	common.RegisterHandler(int(login_proto.LoginPacketID_PACKET_CS_LOGOUT), &CSLogoutHandler{})
	netlib.RegisterFactory(int(login_proto.LoginPacketID_PACKET_CS_LOGOUT), &CSLogoutPacketFactory{})
	netlib.RegisterHandler(int(login_proto.GatePacketID_PACKET_SS_DICONNECT), &SSDisconnectHandler{})
	netlib.RegisterFactory(int(login_proto.GatePacketID_PACKET_SS_DICONNECT), &SSDisconnectPacketFactory{})
}
