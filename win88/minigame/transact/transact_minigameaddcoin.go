package transact

import (
	"games.yol.com/win88/common"
	"games.yol.com/win88/minigame/base"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/core/transact"
	"github.com/idealeak/goserver/srvlib"
)

const (
	TRANSACT_MINIGAMEADDCOIN_CTX = iota
)

type MiniGameAddCoinTransactHandler struct {
}

func (this *MiniGameAddCoinTransactHandler) OnExcute(tNode *transact.TransNode, ud interface{}) transact.TransExeResult {
	logger.Logger.Trace("MiniGameAddCoinTransactHandler.OnExcute ")
	if ctxTx, ok := ud.(*base.AsynAddCoinTranscatCtx); ok {
		if ctxTx.Ctx.Player != nil {
			pack := &common.WGAddCoin{
				SnId:      ctxTx.Ctx.Player.SnId,
				Coin:      ctxTx.Ctx.Coin,
				GainWay:   ctxTx.Ctx.GainWay,
				Oper:      ctxTx.Ctx.Oper,
				Remark:    ctxTx.Ctx.Remark,
				Broadcast: ctxTx.Ctx.Broadcast,
				WriteLog:  ctxTx.Ctx.WriteLog,
			}
			tnp := &transact.TransNodeParam{
				Tt:     common.TransType_MiniGameAddCoin,
				Ot:     transact.TransOwnerType(srvlib.WorldServerType),
				Oid:    int(common.GetWorldSrvId()),
				AreaID: common.GetSelfAreaId(),
				Tct:    transact.TransactCommitPolicy_TwoPhase,
			}

			tNode.TransEnv.SetField(TRANSACT_MINIGAMEADDCOIN_CTX, ud)
			tNode.StartChildTrans(tnp, pack, base.TransAddCoinTimeOut)
		}
	}

	return transact.TransExeResult_Success
}

func (this *MiniGameAddCoinTransactHandler) OnCommit(tNode *transact.TransNode) transact.TransExeResult {
	logger.Logger.Trace("MiniGameAddCoinTransactHandler.OnCommit ")
	ud := tNode.TransEnv.GetField(TRANSACT_MINIGAMEADDCOIN_CTX)
	if ctxTx, ok := ud.(*base.AsynAddCoinTranscatCtx); ok {
		if ctxTx.Ctx != nil && ctxTx.Cb != nil {
			logger.Logger.Tracef("MiniGameAddCoinTransactHandler.OnCommit  SnId=%v,ctxTx.Ctx.Player.Coin=%v,ctxTx.Ctx.Coin=%v,ctxTx.Coin=%v", ctxTx.Ctx.Player.SnId, ctxTx.Ctx.Player.Coin, ctxTx.Ctx.Coin, ctxTx.Coin)
			ctxTx.Cb(ctxTx.Ctx, ctxTx.Coin, true)
		} else {
			logger.Logger.Errorf("MiniGameAddCoinTransactHandler.OnCommit SnId=%v,Coin=%v,GainWay=%v,Oper=%v,Remark=%v,RetryCnt=%v", ctxTx.Ctx.Player.SnId, ctxTx.Ctx.Coin, ctxTx.Ctx.GainWay, ctxTx.Ctx.Oper, ctxTx.Ctx.Remark, ctxTx.Ctx.RetryCnt)
		}
	}
	return transact.TransExeResult_Success
}

func (this *MiniGameAddCoinTransactHandler) OnRollBack(tNode *transact.TransNode) transact.TransExeResult {
	logger.Logger.Trace("MiniGameAddCoinTransactHandler.OnRollBack ")
	ud := tNode.TransEnv.GetField(TRANSACT_MINIGAMEADDCOIN_CTX)
	if ctxTx, ok := ud.(*base.AsynAddCoinTranscatCtx); ok {
		ctxTx.Ctx.RetryCnt--
		if ctxTx.Ctx.RetryCnt > 0 && ctxTx.Ctx.Player.GetScene() != nil {
			ctxTx.Ctx.Player.AddCoin(ctxTx.Ctx.Coin, ctxTx.Ctx.GainWay, ctxTx.Ctx.NotifyCli, ctxTx.Ctx.Broadcast, ctxTx.Ctx.Oper, ctxTx.Ctx.Remark, ctxTx.Ctx.RetryCnt, ctxTx.Cb)
		} else {
			if ctxTx.Ctx != nil && ctxTx.Cb != nil {
				ctxTx.Cb(ctxTx.Ctx, 0, false)
			} else {
				logger.Logger.Errorf("MiniGameAddCoinTransactHandler.OnRollBack SnId=%v,Coin=%v,GainWay=%v,Oper=%v,Remark=%v,RetryCnt=%v", ctxTx.Ctx.Player.SnId, ctxTx.Ctx.Coin, ctxTx.Ctx.GainWay, ctxTx.Ctx.Oper, ctxTx.Ctx.Remark, ctxTx.Ctx.RetryCnt)
			}
		}
	}
	return transact.TransExeResult_Success
}

func (this *MiniGameAddCoinTransactHandler) OnChildTransRep(tNode *transact.TransNode, hChild transact.TransNodeID,
	retCode int, ud interface{}) transact.TransExeResult {
	logger.Logger.Trace("MiniGameAddCoinTransactHandler.OnChildTransRep ")
	if retCode == transact.TransResult_Success {
		if data, ok := ud.([]byte); ok {
			var coin int64
			if netlib.UnmarshalPacketNoPackId(data, &coin) == nil {
				ud := tNode.TransEnv.GetField(TRANSACT_MINIGAMEADDCOIN_CTX)
				if ctxTx, ok := ud.(*base.AsynAddCoinTranscatCtx); ok {
					ctxTx.Coin = coin
				}
			}
		}
	}
	return transact.TransExeResult_Success
}

func init() {
	transact.RegisteHandler(common.TransType_MiniGameAddCoin, &MiniGameAddCoinTransactHandler{})
}
