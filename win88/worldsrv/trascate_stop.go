package main

import (
	"time"

	"games.yol.com/win88/common"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/transact"
)

func init() {
	transact.RegisteHandler(common.TransType_StopServer, &transact.TransHanderWrapper{
		OnExecuteWrapper: transact.OnExecuteWrapper(func(tNode *transact.TransNode, ud interface{}) transact.TransExeResult {
			logger.Logger.Infof("StopApi start TransType_StopServer OnExecuteWrapper %x", tNode.MyTnp.TId)
			for _, s := range GameSessMgrSington.gates {
				tnp := &transact.TransNodeParam{
					Tt:     common.TransType_StopServer,
					Ot:     transact.TransOwnerType(s.srvType),
					Oid:    s.srvId,
					AreaID: common.GetSelfAreaId(),
					Tct:    transact.TransactCommitPolicy_TwoPhase,
				}
				tNode.StartChildTrans(tnp, nil, time.Minute*5)
				logger.Logger.Infof("StopApi start TransType_StopServer StartChildTrans srvid:%v srvtype:%v", s.srvId, s.srvType)
			}
			for _, s := range GameSessMgrSington.servers {
				tnp := &transact.TransNodeParam{
					Tt:     common.TransType_StopServer,
					Ot:     transact.TransOwnerType(s.srvType),
					Oid:    s.srvId,
					AreaID: common.GetSelfAreaId(),
					Tct:    transact.TransactCommitPolicy_TwoPhase,
				}
				tNode.StartChildTrans(tnp, nil, time.Minute*5)
				logger.Logger.Infof("StopApi start TransType_StopServer StartChildTrans srvid:%v srvtype:%v", s.srvId, s.srvType)
			}
			return transact.TransExeResult_Success
		}),
		OnCommitWrapper: transact.OnCommitWrapper(func(tNode *transact.TransNode) transact.TransExeResult {
			logger.Logger.Infof("StopApi start TransType_StopServer OnCommitWrapper ")
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
