package base

import (
	"games.yol.com/win88/common"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/module"
	"github.com/idealeak/goserver/core/timer"
	"github.com/idealeak/goserver/core/transact"
	"time"
)

func init() {
	transact.RegisteHandler(common.TransType_StopServer, &transact.TransHanderWrapper{
		OnExecuteWrapper: transact.OnExecuteWrapper(func(tNode *transact.TransNode, ud interface{}) transact.TransExeResult {
			logger.Logger.Infof("StopApi start TransType_StopServer OnExecuteWrapper %x", tNode.MyTnp.TId)
			ClientMgrSington.Running = false
			timer.StartTimer(timer.TimerActionWrapper(func(h timer.TimerHandle, ud interface{}) bool {
				module.Stop()
				return true
			}), nil, time.Second*10, 1)
			return transact.TransExeResult_Success
		}),
		OnCommitWrapper: transact.OnCommitWrapper(func(tNode *transact.TransNode) transact.TransExeResult {
			logger.Logger.Info("StopApi start TransType_StopServer OnCommitWrapper")
			return transact.TransExeResult_Success
		}),
		OnRollBackWrapper: transact.OnRollBackWrapper(func(tNode *transact.TransNode) transact.TransExeResult {
			logger.Logger.Info("StopApi start TransType_StopServer OnRollBackWrapper")
			return transact.TransExeResult_Success
		}),
		OnChildRespWrapper: transact.OnChildRespWrapper(func(tNode *transact.TransNode, hChild transact.TransNodeID, retCode int, ud interface{}) transact.TransExeResult {
			logger.Logger.Infof("StopApi start TransType_StopServer OnChildRespWrapper ret:%v childid:%x", retCode, hChild)
			return transact.TransExeResult(retCode)
		}),
	})
}
