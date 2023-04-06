package transact

import (
	"games.yol.com/win88/common"
	"games.yol.com/win88/gamesrv/base"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/core/transact"
)

var CoinSceneChangePacket = &common.WGCoinSceneChange{}

type CoinSceneChangeTransactHandler struct {
}

func (this *CoinSceneChangeTransactHandler) OnExcute(tNode *transact.TransNode, ud interface{}) transact.TransExeResult {
	logger.Logger.Trace("CoinSceneChangeTransactHandler.OnExcute ")
	err := netlib.UnmarshalPacketNoPackId(ud.([]byte), CoinSceneChangePacket)
	if err == nil {
		player := base.PlayerMgrSington.GetPlayerBySnId(CoinSceneChangePacket.SnId)
		if player == nil {
			logger.Logger.Tracef("CoinSceneChangeTransactHandler.OnExcute player == nil snid=%v", CoinSceneChangePacket.SnId)
			return transact.TransExeResult_Success
		}

		if player.GetScene() == nil {
			logger.Logger.Tracef("CoinSceneChangeTransactHandler.OnExcute player.GetScene() == nil snid=%v expect=%v", CoinSceneChangePacket.SnId, CoinSceneChangePacket.SceneId)
			return transact.TransExeResult_Success
		}

		if !player.GetScene().CanChangeCoinScene(player) {
			logger.Logger.Tracef("CoinSceneChangeTransactHandler.OnExcute !GetScene().CanChangeCoinScene snid=%v sceneid=%v state=%v", CoinSceneChangePacket.SnId, player.GetScene().SceneId, player.GetScene().SceneState.GetState())
			return transact.TransExeResult_Failed
		}

		if player.GetScene().HasPlayer(player) {
			player.GetScene().PlayerLeave(player, common.PlayerLeaveReason_ChangeCoinScene, true)
		} else {
			player.GetScene().AudienceLeave(player, common.PlayerLeaveReason_ChangeCoinScene)
		}

		return transact.TransExeResult_Success
	}
	logger.Logger.Trace("CoinSceneChangeTransactHandler.OnExcute failed")
	return transact.TransExeResult_Failed
}

func (this *CoinSceneChangeTransactHandler) OnCommit(tNode *transact.TransNode) transact.TransExeResult {
	logger.Logger.Trace("CoinSceneChangeTransactHandler.OnCommit ")
	return transact.TransExeResult_Success
}

func (this *CoinSceneChangeTransactHandler) OnRollBack(tNode *transact.TransNode) transact.TransExeResult {
	logger.Logger.Trace("CoinSceneChangeTransactHandler.OnRollBack ")
	return transact.TransExeResult_Success
}

func (this *CoinSceneChangeTransactHandler) OnChildTransRep(tNode *transact.TransNode, hChild transact.TransNodeID,
	retCode int, ud interface{}) transact.TransExeResult {
	logger.Logger.Trace("CoinSceneChangeTransactHandler.OnChildTransRep ")
	return transact.TransExeResult_Success
}

func init() {
	transact.RegisteHandler(common.TransType_CoinSceneChange, &CoinSceneChangeTransactHandler{})
}
