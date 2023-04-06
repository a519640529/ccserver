package main

import (
	"games.yol.com/win88/common"
	"games.yol.com/win88/model"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/core/transact"
	"github.com/idealeak/goserver/srvlib"
	"time"
)

const (
	TRANSACT_MINIGAMEADDCOIN_CTX = iota
)

type MiniGameAddCoinTransactHandler struct {
}

func (this *MiniGameAddCoinTransactHandler) OnExcute(tNode *transact.TransNode, ud interface{}) transact.TransExeResult {
	logger.Logger.Trace("MiniGameAddCoinTransactHandler.OnExcute ")
	ctx := &common.WGAddCoin{}
	err := netlib.UnmarshalPacketNoPackId(ud.([]byte), ctx)
	if err != nil {
		logger.Logger.Trace("MiniGameAddCoinTransactHandler.OnExcute failed:", err)
		return transact.TransExeResult_Failed
	}
	p := PlayerMgrSington.GetPlayerBySnId(ctx.SnId)
	if p != nil {
		//玩家可能正在换房间
		if ctx.Coin != 0 && p.scene != nil && !p.scene.IsTestScene() && p.scene.sceneMode != common.SceneMode_Thr { //游戏场中加币,需要同步到gamesrv上
			tnp := &transact.TransNodeParam{
				Tt:     common.TransType_MiniGameAddCoin,
				Ot:     transact.TransOwnerType(srvlib.GameServerType),
				Oid:    int(p.scene.gameSess.GetSrvId()),
				AreaID: common.GetSelfAreaId(),
				Tct:    transact.TransactCommitPolicy_SelfDecide,
			}
			tNode.StartChildTrans(tnp, ctx, time.Duration(tNode.MyTnp.ExpiresTs-time.Now().UnixNano()))
			tNode.TransEnv.SetField(TRANSACT_MINIGAMEADDCOIN_CTX, ctx)
			return transact.TransExeResult_Success
		}

		//钱不够扣
		if ctx.Coin < 0 && -ctx.Coin > p.Coin {
			return transact.TransExeResult_Failed
		}

		logger.Logger.Tracef("MiniGameAddCoinTransactHandler.OnExcute 当前金币:%v 金币变化:%v 变化后金币:%v", p.Coin, ctx.Coin, p.Coin+ctx.Coin)
		p.Coin += ctx.Coin
		p.dirty = true
		p.SendDiffData()
		if !p.IsRob && ctx.WriteLog {
			log := model.NewCoinLogEx(p.SnId, ctx.Coin, p.Coin, p.SafeBoxCoin, p.Ver, ctx.GainWay, 0,
				ctx.Oper, ctx.Remark, p.Platform, p.Channel, p.BeUnderAgentCode, 0, p.PackageID, 0)
			if log != nil {
				LogChannelSington.WriteLog(log)
			}
		}
		tNode.TransRep.RetFiels = p.Coin
		return transact.TransExeResult_Success
	}

	return transact.TransExeResult_Failed
}

func (this *MiniGameAddCoinTransactHandler) OnCommit(tNode *transact.TransNode) transact.TransExeResult {
	logger.Logger.Trace("MiniGameAddCoinTransactHandler.OnCommit ")
	ud := tNode.TransEnv.GetField(TRANSACT_MINIGAMEADDCOIN_CTX)
	if ctx, ok := ud.(*common.WGAddCoin); ok {
		p := PlayerMgrSington.GetPlayerBySnId(ctx.SnId) //重新获得p
		if p != nil {
			p.Coin += ctx.Coin
			p.dirty = true
		}
	}
	return transact.TransExeResult_Success
}

func (this *MiniGameAddCoinTransactHandler) OnRollBack(tNode *transact.TransNode) transact.TransExeResult {
	logger.Logger.Trace("MiniGameAddCoinTransactHandler.OnRollBack ")
	return transact.TransExeResult_Success
}

func (this *MiniGameAddCoinTransactHandler) OnChildTransRep(tNode *transact.TransNode, hChild transact.TransNodeID,
	retCode int, ud interface{}) transact.TransExeResult {
	logger.Logger.Trace("MiniGameAddCoinTransactHandler.OnChildTransRep ")
	if retCode == transact.TransResult_Success {
		if data, ok := ud.([]byte); ok {
			var coin int64
			if netlib.UnmarshalPacketNoPackId(data, &coin) == nil {
				tNode.TransRep.RetFiels = coin
			}
		}
	}
	return transact.TransExeResult_Success
}

func init() {
	transact.RegisteHandler(common.TransType_MiniGameAddCoin, &MiniGameAddCoinTransactHandler{})
}
