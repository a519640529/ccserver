package transact

import (
	"games.yol.com/win88/common"
	"games.yol.com/win88/gamesrv/base"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/core/transact"
)

var MatchSceneChangePacket = &common.WGCoinSceneChange{}

type MatchSceneChangeTransactHandler struct {
}

func (this *MatchSceneChangeTransactHandler) OnExcute(tNode *transact.TransNode, ud interface{}) transact.TransExeResult {
	logger.Logger.Trace("MatchSceneChangeTransactHandler.OnExcute ")
	err := netlib.UnmarshalPacketNoPackId(ud.([]byte), MatchSceneChangePacket)
	if err == nil {
		player := base.PlayerMgrSington.GetPlayerBySnId(MatchSceneChangePacket.SnId)
		if player == nil {
			logger.Logger.Tracef("MatchSceneChangeTransactHandler.OnExcute player == nil snid=%v", MatchSceneChangePacket.SnId)
			return transact.TransExeResult_Success
		}

		if player.GetScene() == nil {
			logger.Logger.Tracef("MatchSceneChangeTransactHandler.OnExcute player.scene == nil snid=%v expect=%v", MatchSceneChangePacket.SnId, MatchSceneChangePacket.SceneId)
			return transact.TransExeResult_Success
		}

		if !player.GetScene().CanChangeCoinScene(player) {
			logger.Logger.Tracef("MatchSceneChangeTransactHandler.OnExcute !scene.CanChangeCoinScene snid=%v sceneid=%v state=%v", MatchSceneChangePacket.SnId, player.GetScene().SceneId, player.GetScene().SceneState.GetState())
			return transact.TransExeResult_Failed
		}

		if player.GetScene().HasPlayer(player) {
			player.GetScene().PlayerLeave(player, common.PlayerLeaveReason_OnBilled, true)
		} else {
			player.GetScene().AudienceLeave(player, common.PlayerLeaveReason_OnBilled)
		}

		return transact.TransExeResult_Success
	}
	logger.Logger.Trace("MatchSceneChangeTransactHandler.OnExcute failed")
	return transact.TransExeResult_Failed
}

func (this *MatchSceneChangeTransactHandler) OnCommit(tNode *transact.TransNode) transact.TransExeResult {
	logger.Logger.Trace("MatchSceneChangeTransactHandler.OnCommit ")
	return transact.TransExeResult_Success
}

func (this *MatchSceneChangeTransactHandler) OnRollBack(tNode *transact.TransNode) transact.TransExeResult {
	logger.Logger.Trace("MatchSceneChangeTransactHandler.OnRollBack ")
	return transact.TransExeResult_Success
}

func (this *MatchSceneChangeTransactHandler) OnChildTransRep(tNode *transact.TransNode, hChild transact.TransNodeID,
	retCode int, ud interface{}) transact.TransExeResult {
	logger.Logger.Trace("MatchSceneChangeTransactHandler.OnChildTransRep ")
	return transact.TransExeResult_Success
}

func init() {
	transact.RegisteHandler(common.TransType_MatchSceneChange, &MatchSceneChangeTransactHandler{})
}
