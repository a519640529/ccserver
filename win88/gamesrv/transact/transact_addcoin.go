package transact

import (
	"games.yol.com/win88/common"
	"games.yol.com/win88/gamesrv/base"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/core/transact"
)

const (
	TRANSACT_ADDCOIN_CTX = iota
)

type AddCoinTransactHandler struct {
}

func (this *AddCoinTransactHandler) OnExcute(tNode *transact.TransNode, ud interface{}) transact.TransExeResult {
	logger.Logger.Trace("AddCoinTransactHandler.OnExcute ")
	ctx := &common.WGAddCoin{}
	err := netlib.UnmarshalPacketNoPackId(ud.([]byte), ctx)
	if err != nil {
		logger.Logger.Trace("AddCoinTransactHandler.OnExcute failed:", err)
		return transact.TransExeResult_Failed
	}
	p := base.PlayerMgrSington.GetPlayerBySnId(ctx.SnId)
	if p != nil {
		s := p.GetScene()
		if s != nil {
			sp := s.GetScenePolicy()
			if sp != nil && sp.CanAddCoin(s, p, ctx.Coin) {
				p.AddCoinAsync(ctx.Coin, ctx.GainWay, true, ctx.Broadcast, ctx.Oper, ctx.Remark, ctx.WriteLog)
				//触发下事件
				sp.OnPlayerEvent(s, p, base.PlayerEventRecharge, []int64{ctx.Coin})
				//
				tNode.TransEnv.SetField(TRANSACT_ADDCOIN_CTX, ctx)
				tNode.TransRep.RetFiels = p.Coin
				return transact.TransExeResult_Success
			}
		}
	}

	return transact.TransExeResult_Failed
}

func (this *AddCoinTransactHandler) OnCommit(tNode *transact.TransNode) transact.TransExeResult {
	logger.Logger.Trace("AddCoinTransactHandler.OnCommit ")
	return transact.TransExeResult_Success
}

func (this *AddCoinTransactHandler) OnRollBack(tNode *transact.TransNode) transact.TransExeResult {
	logger.Logger.Trace("AddCoinTransactHandler.OnRollBack ")
	ud := tNode.TransEnv.GetField(TRANSACT_ADDCOIN_CTX)
	if ctx, ok := ud.(*common.WGAddCoin); ok {
		p := base.PlayerMgrSington.GetPlayerBySnId(ctx.SnId)
		if p != nil {
			s := p.GetScene()
			if s != nil {
				sp := s.GetScenePolicy()
				if sp != nil && sp.CanAddCoin(s, p, -ctx.Coin) {
					p.AddCoinAsync(-ctx.Coin, ctx.GainWay, true, ctx.Broadcast, ctx.Oper, ctx.Remark, ctx.WriteLog)
					//触发下事件
					sp.OnPlayerEvent(s, p, base.PlayerEventRecharge, []int64{0})
					return transact.TransExeResult_Success
				}
			}
		}
	}
	return transact.TransExeResult_Success
}

func (this *AddCoinTransactHandler) OnChildTransRep(tNode *transact.TransNode, hChild transact.TransNodeID,
	retCode int, ud interface{}) transact.TransExeResult {
	logger.Logger.Trace("AddCoinTransactHandler.OnChildTransRep ")
	return transact.TransExeResult_Success
}

func init() {
	transact.RegisteHandler(common.TransType_AddCoin, &AddCoinTransactHandler{})
}
