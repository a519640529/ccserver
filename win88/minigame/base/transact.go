package base

import (
	"games.yol.com/win88/common"
	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/transact"
	"time"
)

var TransAddCoinTimeOut = time.Second * 30

type AddCoinContext struct {
	*Player
	Coin      int64
	GainWay   int32
	Oper      string
	Remark    string
	NotifyCli bool
	Broadcast bool
	WriteLog  bool
	RetryCnt  int
}

type AddCoinCb func(*AddCoinContext, int64, bool)
type AsynAddCoinTranscatCtx struct {
	Ctx  *AddCoinContext
	Cb   AddCoinCb
	Coin int64 //加币成功后的余额
}

func StartAsyncAddCoinTransact(ctx *AddCoinContext, cb AddCoinCb) bool {
	tnp := &transact.TransNodeParam{
		Tt:     common.TransType_MiniGameAddCoin,
		Ot:     transact.TransOwnerType(common.GetSelfSrvType()),
		Oid:    common.GetSelfSrvId(),
		AreaID: common.GetSelfAreaId(),
	}
	ctxTx := &AsynAddCoinTranscatCtx{
		Ctx: ctx,
		Cb:  cb,
	}
	tNode := transact.DTCModule.StartTrans(tnp, ctxTx, TransAddCoinTimeOut)
	if tNode != nil {
		tNode.Go(core.CoreObject())
		return true
	}

	return false
}
