package main

import (
	"time"

	"games.yol.com/win88/common"
	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/transact"
	"github.com/idealeak/goserver/srvlib"
)

var wgDayTimeChangePack = &common.WGDayTimeChange{}
var DayTimeChangeTimeOut = time.Second * 10

type DayTimeChangeTransactHandler struct {
}

func (this *DayTimeChangeTransactHandler) OnExcute(tNode *transact.TransNode, ud interface{}) transact.TransExeResult {
	//logger.Logger.Trace("DayTimeChangeTransactHandler.OnExcute ")
	ClockMgrSington.Notifying = true
	for sid, _ := range GameSessMgrSington.servers {
		tnp := &transact.TransNodeParam{
			Tt:     common.TransType_DayTimeChange,
			Ot:     transact.TransOwnerType(srvlib.GameServerType),
			Oid:    sid,
			AreaID: common.GetSelfAreaId(),
			Tct:    transact.TransactCommitPolicy_SelfDecide,
		}
		//logger.Logger.Tracef("TransNode=%v", *tnp)
		_, wgDayTimeChangePack.LastMin, wgDayTimeChangePack.LastHour, wgDayTimeChangePack.LastDay, wgDayTimeChangePack.LastWeek, wgDayTimeChangePack.LastMonth = ClockMgrSington.GetLast()
		tNode.StartChildTrans(tnp, wgDayTimeChangePack, DayTimeChangeTimeOut)
	}

	return transact.TransExeResult_Success
}

func (this *DayTimeChangeTransactHandler) OnCommit(tNode *transact.TransNode) transact.TransExeResult {
	//logger.Logger.Trace("DayTimeChangeTransactHandler.OnCommit ")
	ClockMgrSington.Notifying = false
	return transact.TransExeResult_Success
}

func (this *DayTimeChangeTransactHandler) OnRollBack(tNode *transact.TransNode) transact.TransExeResult {
	//logger.Logger.Trace("DayTimeChangeTransactHandler.OnRollBack ")
	ClockMgrSington.Notifying = false
	return transact.TransExeResult_Success
}

func (this *DayTimeChangeTransactHandler) OnChildTransRep(tNode *transact.TransNode, hChild transact.TransNodeID,
	retCode int, ud interface{}) transact.TransExeResult {
	//logger.Logger.Trace("DayTimeChangeTransactHandler.OnChildTransRep ")
	return transact.TransExeResult_Success
}

type DayTimeChangeTransactSinker struct {
	BaseClockSinker
}

func (this *DayTimeChangeTransactSinker) OnMiniTimer() {
	tnp := &transact.TransNodeParam{
		Tt:     common.TransType_DayTimeChange,
		Ot:     transact.TransOwnerType(common.GetSelfSrvType()),
		Oid:    common.GetSelfSrvId(),
		AreaID: common.GetSelfAreaId(),
	}
	tNode := transact.DTCModule.StartTrans(tnp, nil, DayTimeChangeTimeOut)
	if tNode != nil {
		tNode.Go(core.CoreObject())
	}
}

func init() {
	ClockMgrSington.RegisteSinker(&DayTimeChangeTransactSinker{})
	transact.RegisteHandler(common.TransType_DayTimeChange, &DayTimeChangeTransactHandler{})
}
