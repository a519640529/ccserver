package transact

import (
	"games.yol.com/win88/common"
	"games.yol.com/win88/gamesrv/base"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/transact"
	"github.com/idealeak/goserver/srvlib"
	"time"
)

func init() {
	transact.RegisteHandler(common.TransType_StopServer, &transact.TransHanderWrapper{
		OnExecuteWrapper: transact.OnExecuteWrapper(func(tNode *transact.TransNode, ud interface{}) transact.TransExeResult {
			logger.Logger.Infof("StopApi start TransType_StopServer OnExecuteWrapper %x", tNode.MyTnp.TId)
			base.SceneMgrSington.DestoryAllScene()
			//通知机器人关闭
			npcSess := srvlib.ServerSessionMgrSington.GetSession(common.GetSelfAreaId(), common.RobotServerType, common.RobotServerId)
			if npcSess != nil {
				tnp := &transact.TransNodeParam{
					Tt:     common.TransType_StopServer,
					Ot:     transact.TransOwnerType(common.RobotServerType),
					Oid:    common.RobotServerId,
					AreaID: common.GetSelfAreaId(),
					Tct:    transact.TransactCommitPolicy_TwoPhase,
				}
				tNode.StartChildTrans(tnp, nil, time.Second*5)
				logger.Logger.Infof("StopApi start TransType_StopServer StartChildTrans srvid:%v srvtype:%v", common.RobotServerId, common.RobotServerType)
			}
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
