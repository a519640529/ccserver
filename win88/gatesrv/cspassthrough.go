package main

import (
	"fmt"
	"games.yol.com/win88/common"
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/player"
	"games.yol.com/win88/webapi"
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/core/task"
)

func init() {
	netlib.RegisterFactory(int(player.PlayerPacketID_PACKET_CS_WEBAPI_PLAYERPASS), netlib.PacketFactoryWrapper(func() interface{} {
		return &player.CSWebAPIPlayerPass{}
	}))

	netlib.RegisterHandler(int(player.PlayerPacketID_PACKET_CS_WEBAPI_PLAYERPASS), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
		logger.Logger.Trace("receive CSWebAPIPlayerPass info==", pack)
		if msg, ok := pack.(*player.CSWebAPIPlayerPass); ok {
			var pdi = s.GetAttribute(common.ClientSessionAttribute_PlayerData)
			if pdi == nil {
				return nil
			}
			var playerData = pdi.(*model.PlayerData)
			opCode := player.OpResultCode_OPRC_Sucess
			errString := ""
			var err error
			t, done := task.NewMutexTask(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
				errString, err = webapi.API_PlayerPass(playerData.SnId, playerData.Platform, playerData.Channel, playerData.BeUnderAgentCode, msg.GetApiName(), msg.GetParams(), common.GetAppId(), playerData.LogicLevels)
				if err != nil {
					logger.Logger.Errorf("API_PlayerPass error:%v api:%v params:%v", err, msg.GetApiName(), msg.GetParams())
					opCode = player.OpResultCode_OPRC_Error
					return nil
				}
				return err
			}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
				pack := &player.SCWebAPIPlayerPass{
					OpRetCode: opCode,
					ApiName:   msg.ApiName,
					CBData:    msg.CBData,
					Response:  proto.String(errString),
				}
				s.Send(int(player.PlayerPacketID_PACKET_SC_WEBAPI_PLAYERPASS), pack)
				logger.Logger.Trace("CSWebAPIPlayerPass:", pack)
			}), fmt.Sprintf("%v?%v", msg.GetApiName(), msg.GetParams()), "API_PlayerPass")
			if !done {
				t.Start()
			}
		}
		return nil
	}))

	netlib.RegisterFactory(int(player.PlayerPacketID_PACKET_CS_WEBAPI_SYSTEMPASS), netlib.PacketFactoryWrapper(func() interface{} {
		return &player.CSWebAPISystemPass{}
	}))

	netlib.RegisterHandler(int(player.PlayerPacketID_PACKET_CS_WEBAPI_SYSTEMPASS), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
		logger.Logger.Trace("receive CSWebAPIPlayerPass info==", pack)
		if msg, ok := pack.(*player.CSWebAPISystemPass); ok {
			opCode := player.OpResultCode_OPRC_Sucess
			errString := ""
			var err error
			t, done := task.NewMutexTask(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
				errString, err = webapi.API_SystemPass(msg.GetApiName(), msg.GetParams(), common.GetAppId())
				if err != nil {
					logger.Logger.Errorf("API_SystemPass error:%v apiname=%v params=%v", err, msg.GetApiName(), msg.GetParams())
					opCode = player.OpResultCode_OPRC_Error
					return nil
				}
				return err
			}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
				pack := &player.SCWebAPISystemPass{
					OpRetCode: opCode,
					ApiName:   msg.ApiName,
					CBData:    msg.CBData,
					Response:  proto.String(errString),
				}
				s.Send(int(player.PlayerPacketID_PACKET_SC_WEBAPI_SYSTEMPASS), pack)
				logger.Logger.Trace("CSWebAPISystemPass:", pack)
			}), fmt.Sprintf("%v?%v", msg.GetApiName(), msg.GetParams()), "API_SystemPass")
			if !done {
				t.Start()
			}
		}
		return nil
	}))
}
