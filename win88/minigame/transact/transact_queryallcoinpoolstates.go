package transact

import (
	"games.yol.com/win88/common"
	"games.yol.com/win88/minigame/base"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/core/transact"
	"time"
)

var QueryAllCoinPoolTimeOut = time.Second * 30

const (
	QueryAllCoinPoolTransactParam_ParentNode int = iota
	QueryAllCoinPoolTransactParam_Data
)

type QueryAllCoinPoolTransactHandler struct {
}

func (this *QueryAllCoinPoolTransactHandler) OnExcute(tNode *transact.TransNode, ud interface{}) transact.TransExeResult {
	//logger.Logger.Trace("QueryAllCoinPoolTransactHandler.OnExcute ")
	PlatformStates := make(map[string]*common.PlatformStates)
	pfs := &common.QueryGames{}
	err := netlib.UnmarshalPacketNoPackId(ud.([]byte), pfs)
	if err == nil {
		for pfId, game := range pfs.Index {
			// pfId 为paltform ID  p为=该平台下的所有开启的游戏
			pf := &common.PlatformStates{}
			pf.Platform = pfId
			settings := base.CoinPoolMgr.GetCoinPoolStatesByPlatform(pfId, game)
			pf.GamesVal = settings
			PlatformStates[pfId] = pf
		}
	}
	tNode.TransRep.RetFiels = PlatformStates
	return transact.TransExeResult_Success
}

func (this *QueryAllCoinPoolTransactHandler) OnCommit(tNode *transact.TransNode) transact.TransExeResult {
	//logger.Logger.Trace("QueryAllCoinPoolTransactHandler.OnCommit ")
	return transact.TransExeResult_Success
}

func (this *QueryAllCoinPoolTransactHandler) OnRollBack(tNode *transact.TransNode) transact.TransExeResult {
	//logger.Logger.Trace("QueryAllCoinPoolTransactHandler.OnRollBack ")
	return transact.TransExeResult_Success
}

func (this *QueryAllCoinPoolTransactHandler) OnChildTransRep(tNode *transact.TransNode, hChild transact.TransNodeID,
	retCode int, ud interface{}) transact.TransExeResult {
	//logger.Logger.Trace("QueryAllCoinPoolTransactHandler.OnChildTransRep ")
	return transact.TransExeResult_Success
}

func init() {
	transact.RegisteHandler(common.TransType_QueryAllCoinPool, &QueryAllCoinPoolTransactHandler{})
}
