package main

import (
	"time"

	"games.yol.com/win88/common"
	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/transact"
	"github.com/idealeak/goserver/srvlib"
)

var TransAddCoinTimeOut = time.Second * 30

const (
	TRANSACT_ADDCOIN_CTX = iota
)

type AsyncAddCoinTransactContext struct {
	p         *Player
	coin      int64
	gainWay   int32
	oper      string
	remark    string
	broadcast bool
	writeLog  bool
	retryCnt  int
}
type AddCoinTransactHandler struct {
}

func (this *AddCoinTransactHandler) OnExcute(tNode *transact.TransNode, ud interface{}) transact.TransExeResult {
	logger.Logger.Trace("AddCoinTransactHandler.OnExcute ")
	if ctx, ok := ud.(*AsyncAddCoinTransactContext); ok {
		if ctx.p != nil && ctx.p.scene != nil {
			pack := &common.WGAddCoin{
				SnId:      ctx.p.SnId,
				Coin:      ctx.coin,
				GainWay:   ctx.gainWay,
				Oper:      ctx.oper,
				Remark:    ctx.remark,
				Broadcast: ctx.broadcast,
				WriteLog:  ctx.writeLog,
			}
			tnp := &transact.TransNodeParam{
				Tt:     common.TransType_AddCoin,
				Ot:     transact.TransOwnerType(srvlib.GameServerType),
				Oid:    int(ctx.p.scene.gameSess.GetSrvId()),
				AreaID: common.GetSelfAreaId(),
				Tct:    transact.TransactCommitPolicy_TwoPhase,
			}

			tNode.TransEnv.SetField(TRANSACT_ADDCOIN_CTX, ud)
			tNode.StartChildTrans(tnp, pack, TransAddCoinTimeOut)
		}
	}

	return transact.TransExeResult_Success
}

func (this *AddCoinTransactHandler) OnCommit(tNode *transact.TransNode) transact.TransExeResult {
	logger.Logger.Trace("AddCoinTransactHandler.OnCommit ")
	ud := tNode.TransEnv.GetField(TRANSACT_ADDCOIN_CTX)
	if ctx, ok := ud.(*AsyncAddCoinTransactContext); ok {
		p := PlayerMgrSington.GetPlayerBySnId(ctx.p.SnId) //重新获得p
		if p != nil {
			p.Coin += ctx.coin
			p.dirty = true
		}
	}
	return transact.TransExeResult_Success
}

func (this *AddCoinTransactHandler) OnRollBack(tNode *transact.TransNode) transact.TransExeResult {
	logger.Logger.Trace("AddCoinTransactHandler.OnRollBack ")
	ud := tNode.TransEnv.GetField(TRANSACT_ADDCOIN_CTX)
	if ctx, ok := ud.(*AsyncAddCoinTransactContext); ok {
		p := PlayerMgrSington.GetPlayerBySnId(ctx.p.SnId) //重新获得p
		if p != nil {
			p.AddCoinAsync(ctx.coin, ctx.gainWay, ctx.oper, ctx.remark, ctx.broadcast, ctx.retryCnt+1, ctx.writeLog)
		} else {
			logger.Logger.Errorf("AddCoinTransactHandler.OnRollBack SnId=%v,Coin=%v,GainWay=%v,Oper=%v,Remark=%v,RetryCnt=%v", ctx.p.SnId, ctx.coin, ctx.gainWay, ctx.oper, ctx.remark, ctx.retryCnt)
		}
	}
	return transact.TransExeResult_Success
}

func (this *AddCoinTransactHandler) OnChildTransRep(tNode *transact.TransNode, hChild transact.TransNodeID,
	retCode int, ud interface{}) transact.TransExeResult {
	logger.Logger.Trace("AddCoinTransactHandler.OnChildTransRep ")
	return transact.TransExeResult_Success
}

func StartAsyncAddCoinTransact(p *Player, num int64, gainWay int32, oper, remark string, broadcast bool, retryCnt int, writeLog bool) bool {
	tnp := &transact.TransNodeParam{
		Tt:     common.TransType_AddCoin,
		Ot:     transact.TransOwnerType(common.GetSelfSrvType()),
		Oid:    common.GetSelfSrvId(),
		AreaID: common.GetSelfAreaId(),
	}
	ctx := &AsyncAddCoinTransactContext{
		p:         p,
		coin:      num,
		gainWay:   gainWay,
		oper:      oper,
		remark:    remark,
		broadcast: broadcast,
		writeLog:  writeLog,
		retryCnt:  retryCnt,
	}
	tNode := transact.DTCModule.StartTrans(tnp, ctx, TransAddCoinTimeOut)
	if tNode != nil {
		tNode.Go(core.CoreObject())
		return true
	}

	return false
}

func init() {
	transact.RegisteHandler(common.TransType_AddCoin, &AddCoinTransactHandler{})
}
