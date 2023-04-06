package main

import (
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/module"
	"github.com/idealeak/goserver/core/signal"
	"os"
	"time"

	"github.com/idealeak/goserver/core/timer"

	"games.yol.com/win88/common"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/server"
	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/transact"
	"github.com/idealeak/goserver/srvlib"
)

const (
	STOPAPI_TRANSACTE_UD int = iota
)

type StopAPIUserData struct {
	wait    chan struct{}
	srvtype int
	timeout time.Duration
}

type InterruptSignalHandler struct {
}

func (ish *InterruptSignalHandler) Process(s os.Signal, ud interface{}) error {
	logger.Logger.Warn("Receive Interrupt signal, process be close!!!")
	wait := make(chan struct{}, 1)

	//wait all world server close
	StopServer(wait, srvlib.WorldServerType, time.Minute*5)

	//shutdown all server
	ShutdownAllServer(wait, time.Minute*10)

	//close self
	core.CoreObject().SendCommand(basic.CommandWrapper(func(o *basic.Object) error {
		module.Stop()
		return nil
	}), false)
	return nil
}

func StopServer(wait chan struct{}, srvtype int, timeout time.Duration) {
	core.CoreObject().SendCommand(basic.CommandWrapper(func(o *basic.Object) error {
		logger.Logger.Infof("StopApi start transcate srvtype(%v) timeout(%v)", srvtype, timeout)
		tnp := &transact.TransNodeParam{
			Tt:     common.TransType_StopServer,
			Ot:     transact.TransOwnerType(common.GetSelfSrvType()),
			Oid:    common.GetSelfSrvId(),
			AreaID: common.GetSelfAreaId(),
		}
		tNode := transact.DTCModule.StartTrans(tnp, &StopAPIUserData{wait: wait, srvtype: srvtype, timeout: timeout}, timeout)
		if tNode != nil {
			tNode.Go(core.CoreObject())
		}
		return nil
	}), false)
	select {
	case _ = <-wait:
		logger.Logger.Infof("StopApi transcate srvtype(%v) all done!!!", srvtype)
	case <-time.After(timeout):
		logger.Logger.Infof("StopApi transcate srvtype(%v) timeout force stop!!!", srvtype)
	}
}

func ShutdownAllServer(wait chan struct{}, timeout time.Duration) {
	core.CoreObject().SendCommand(basic.CommandWrapper(func(o *basic.Object) error {
		logger.Logger.Infof("StopApi start shutdown all server")
		ctrlPacket := &server.ServerCtrl{
			CtrlCode: proto.Int32(common.SrvCtrlCloseCode),
		}
		proto.SetDefaults(ctrlPacket)
		srvlib.ServerSessionMgrSington.Broadcast(int(server.SSPacketID_PACKET_MS_SRVCTRL), ctrlPacket, common.GetSelfAreaId(), srvlib.GameServerType)
		srvlib.ServerSessionMgrSington.Broadcast(int(server.SSPacketID_PACKET_MS_SRVCTRL), ctrlPacket, common.GetSelfAreaId(), srvlib.GateServerType)
		srvlib.ServerSessionMgrSington.Broadcast(int(server.SSPacketID_PACKET_MS_SRVCTRL), ctrlPacket, common.GetSelfAreaId(), srvlib.WorldServerType)

		//启动定时器检测
		timer.StartTimer(timer.TimerActionWrapper(func(h timer.TimerHandle, ud interface{}) bool {
			gameId := srvlib.ServerSessionMgrSington.GetServerId(common.GetSelfAreaId(), srvlib.GameServerType)
			if gameId != -1 {
				logger.Logger.Infof("StopApi start shutdown all server gameId:%v", gameId)
				return true
			}
			gateId := srvlib.ServerSessionMgrSington.GetServerId(common.GetSelfAreaId(), srvlib.GateServerType)
			if gateId != -1 {
				logger.Logger.Infof("StopApi start shutdown all server gateId:%v", gateId)
				return true
			}
			worldId := srvlib.ServerSessionMgrSington.GetServerId(common.GetSelfAreaId(), srvlib.WorldServerType)
			if worldId != -1 {
				logger.Logger.Infof("StopApi start shutdown all server worldId:%v", worldId)
				return true
			}
			wait <- struct{}{}
			timer.StopTimer(h)
			return true
		}), nil, time.Second, -1)
		return nil
	}), false)

	select {
	case _ = <-wait:
		logger.Logger.Info("StopApi ShutdownAllServer all done!!!")
	case <-time.After(timeout):
		logger.Logger.Info("StopApi ShutdownAllServer timeout force stop!!!")
	}
}
func init() {
	core.RegisteHook(core.HOOK_BEFORE_START, func() error {
		signal.SignalHandlerModule.ClearHandler(os.Interrupt)
		signal.SignalHandlerModule.RegisteHandler(os.Interrupt, &InterruptSignalHandler{}, nil)
		return nil
	})

	transact.RegisteHandler(common.TransType_StopServer, &transact.TransHanderWrapper{
		OnExecuteWrapper: transact.OnExecuteWrapper(func(tNode *transact.TransNode, ud interface{}) transact.TransExeResult {
			logger.Logger.Info("StopApi start TransType_StopServer OnExecuteWrapper ")
			if stopUD, ok := ud.(*StopAPIUserData); ok {
				tNode.TransEnv.SetField(STOPAPI_TRANSACTE_UD, ud)
				ids := srvlib.ServerSessionMgrSington.GetServerIds(common.GetSelfAreaId(), stopUD.srvtype)
				for _, id := range ids {
					tnp := &transact.TransNodeParam{
						Tt:     common.TransType_StopServer,
						Ot:     transact.TransOwnerType(stopUD.srvtype),
						Oid:    id,
						AreaID: common.GetSelfAreaId(),
						Tct:    transact.TransactCommitPolicy_TwoPhase,
					}
					tNode.StartChildTrans(tnp, nil, stopUD.timeout)
				}
				return transact.TransExeResult_Success
			}
			return transact.TransExeResult_Failed
		}),
		OnCommitWrapper: transact.OnCommitWrapper(func(tNode *transact.TransNode) transact.TransExeResult {
			logger.Logger.Info("StopApi start TransType_StopServer OnCommitWrapper")
			field := tNode.TransEnv.GetField(STOPAPI_TRANSACTE_UD)
			if field != nil {
				if ud, ok := field.(*StopAPIUserData); ok {
					ud.wait <- struct{}{}
				}
			}
			return transact.TransExeResult_Success
		}),
		OnRollBackWrapper: transact.OnRollBackWrapper(func(tNode *transact.TransNode) transact.TransExeResult {
			logger.Logger.Info("StopApi start TransType_StopServer OnRollBackWrapper")
			field := tNode.TransEnv.GetField(STOPAPI_TRANSACTE_UD)
			if field != nil {
				if ud, ok := field.(*StopAPIUserData); ok {
					ud.wait <- struct{}{}
				}
			}
			return transact.TransExeResult_Success
		}),
		OnChildRespWrapper: transact.OnChildRespWrapper(func(tNode *transact.TransNode, hChild transact.TransNodeID, retCode int, ud interface{}) transact.TransExeResult {
			logger.Logger.Infof("StopApi start TransType_StopServer OnChildRespWrapper ret:%v childid:%x", retCode, hChild)
			return transact.TransExeResult(retCode)
		}),
	})
}
