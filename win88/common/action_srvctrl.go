package common

import (
	"games.yol.com/win88/protocol/server"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/module"
	"github.com/idealeak/goserver/core/netlib"
)

type GameSessState int

const (
	GAME_SESS_STATE_OFF GameSessState = iota //关闭状态
	GAME_SESS_STATE_ON                       //开启状态
)

var ServerCtrlCallback func(int32)

func RegisteServerCtrlCallback(cb func(int32)) {
	ServerCtrlCallback = cb
}

type ServerCtrlPacketFactory struct {
}

type ServerCtrlHandler struct {
}

func (this *ServerCtrlPacketFactory) CreatePacket() interface{} {
	pack := &server.ServerCtrl{}
	return pack
}

func (this *ServerCtrlHandler) Process(s *netlib.Session, packetid int, data interface{}) error {

	if sc, ok := data.(*server.ServerCtrl); ok {
		logger.Logger.Trace("ServerCtrlHandler.Process== ", *sc)
		switch sc.GetCtrlCode() {
		case SrvCtrlCloseCode:
			module.Stop()
		}

		//回调
		if ServerCtrlCallback != nil {
			ServerCtrlCallback(sc.GetCtrlCode())
		}
	}
	return nil
}

func init() {
	netlib.RegisterHandler(int(server.SSPacketID_PACKET_MS_SRVCTRL), &ServerCtrlHandler{})
	netlib.RegisterFactory(int(server.SSPacketID_PACKET_MS_SRVCTRL), &ServerCtrlPacketFactory{})
}
