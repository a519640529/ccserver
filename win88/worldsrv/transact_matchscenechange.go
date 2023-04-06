package main

import (
	"time"

	"games.yol.com/win88/common"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/transact"
	"github.com/idealeak/goserver/srvlib"
)

var MatchSceneChangeTimeOut = time.Second * 10
var MatchSceneChangeTransactParam int

type MatchSceneChangeCtx struct {
	snid       int32
	sceneId    int32
	matchId    int32
	processIdx int
}

type MatchSceneChangeTransactHandler struct {
}

func (this *MatchSceneChangeTransactHandler) OnExcute(tNode *transact.TransNode, ud interface{}) transact.TransExeResult {
	logger.Logger.Trace("MatchSceneChangeTransactHandler.OnExcute")
	if ctx, ok := ud.(*MatchSceneChangeCtx); ok {
		player := PlayerMgrSington.GetPlayerBySnId(ctx.snid)
		if player != nil && player.scene != nil {
			tnp := &transact.TransNodeParam{
				Tt:     common.TransType_MatchSceneChange,
				Ot:     transact.TransOwnerType(srvlib.GameServerType),
				Oid:    int(player.scene.gameSess.GetSrvId()),
				AreaID: common.GetSelfAreaId(),
				Tct:    transact.TransactCommitPolicy_SelfDecide,
			}
			pack := &common.WGCoinSceneChange{
				SnId:    ctx.snid,
				SceneId: ctx.sceneId,
			}
			tNode.TransEnv.SetField(MatchSceneChangeTransactParam, ud)
			tNode.StartChildTrans(tnp, pack, MatchSceneChangeTimeOut)
		}
	}

	return transact.TransExeResult_Success
}

func (this *MatchSceneChangeTransactHandler) OnCommit(tNode *transact.TransNode) transact.TransExeResult {
	logger.Logger.Trace("MatchSceneChangeTransactHandler.OnCommit")
	ud := tNode.TransEnv.GetField(MatchSceneChangeTransactParam)
	if ctx, ok := ud.(*MatchSceneChangeCtx); ok {
		player := PlayerMgrSington.GetPlayerBySnId(ctx.snid)
		if player != nil {
			//m := MatchMgrSington.GetCopyMatch(ctx.matchId)
			//if m != nil {
			//	if m.process != nil && m.process.idx == ctx.processIdx {
			//		m.process.TryEnterMatch(player.matchCtx)
			//	}
			//}
		}
	}
	return transact.TransExeResult_Success
}

func (this *MatchSceneChangeTransactHandler) OnRollBack(tNode *transact.TransNode) transact.TransExeResult {
	logger.Logger.Trace("MatchSceneChangeTransactHandler.OnRollBack")
	ud := tNode.TransEnv.GetField(MatchSceneChangeTransactParam)
	if ctx, ok := ud.(*MatchSceneChangeCtx); ok {
		player := PlayerMgrSington.GetPlayerBySnId(ctx.snid)
		if player != nil {
			//m := MatchMgrSington.GetCopyMatch(ctx.matchId)
			//if m != nil {
			//	if m.process != nil && m.process.idx == ctx.processIdx {
			//		m.process.waitPlayer = append(m.process.waitPlayer, player.matchCtx)
			//	}
			//}
		}
	}
	return transact.TransExeResult_Success
}

func (this *MatchSceneChangeTransactHandler) OnChildTransRep(tNode *transact.TransNode, hChild transact.TransNodeID,
	retCode int, ud interface{}) transact.TransExeResult {
	logger.Logger.Trace("MatchSceneChangeTransactHandler.OnChildTransRep")
	return transact.TransExeResult_Success
}

func init() {
	transact.RegisteHandler(common.TransType_MatchSceneChange, &MatchSceneChangeTransactHandler{})
}
