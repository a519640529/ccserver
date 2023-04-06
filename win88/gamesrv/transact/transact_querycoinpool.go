package transact

import (
	"games.yol.com/win88/common"
	"games.yol.com/win88/gamesrv/base"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/core/transact"
	"time"
)

var QueryCoinPoolTimeOut = time.Second * 30

const (
	QueryCoinPoolTransactParam_ParentNode int = iota
	QueryCoinPoolTransactParam_Data
)

type QueryCoinPoolTransactHandler struct {
}

func (this *QueryCoinPoolTransactHandler) OnExcute(tNode *transact.TransNode, ud interface{}) transact.TransExeResult {
	logger.Logger.Trace("QueryCoinPoolTransactHandler.OnExcute ")
	pack := &common.W2GQueryCoinPool{}
	err := netlib.UnmarshalPacketNoPackId(ud.([]byte), pack)
	if err == nil {
		settings := base.CoinPoolMgr.GetCoinPoolSettingByGame(pack.Platform, pack.GameId, pack.GameMode, pack.GroupId)
		tNode.TransRep.RetFiels = settings
	}
	return transact.TransExeResult_Success
}

func (this *QueryCoinPoolTransactHandler) OnCommit(tNode *transact.TransNode) transact.TransExeResult {
	logger.Logger.Trace("QueryCoinPoolTransactHandler.OnCommit ")
	return transact.TransExeResult_Success
}

func (this *QueryCoinPoolTransactHandler) OnRollBack(tNode *transact.TransNode) transact.TransExeResult {
	logger.Logger.Trace("QueryCoinPoolTransactHandler.OnRollBack ")
	return transact.TransExeResult_Success
}

func (this *QueryCoinPoolTransactHandler) OnChildTransRep(tNode *transact.TransNode, hChild transact.TransNodeID,
	retCode int, ud interface{}) transact.TransExeResult {
	logger.Logger.Trace("QueryCoinPoolTransactHandler.OnChildTransRep ")
	return transact.TransExeResult_Success
}

func init() {
	transact.RegisteHandler(common.TransType_QueryCoinPool, &QueryCoinPoolTransactHandler{})
}
