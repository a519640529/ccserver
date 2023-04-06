package transact

import (
	"container/list"
	"fmt"
	"games.yol.com/win88/common"
	"games.yol.com/win88/gamesrv/base"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/core/transact"
	"github.com/idealeak/goserver/core/utils"
)

type DayTimeChangeListener interface {
	OnMiniTimer()
	OnHourTimer()
	OnDayTimer()
	OnWeekTimer()
	OnMonthTimer()
}

var DayTimeChangeListeners = list.New()
var WGDayTimeChangePack = &common.WGDayTimeChange{}
var LastDayTimeRec = common.WGDayTimeChange{}

func RegisteDayTimeChangeListener(lis DayTimeChangeListener) {
	for e := DayTimeChangeListeners.Front(); e != nil; e = e.Next() {
		if e.Value == lis {
			panic(fmt.Sprintf("RegisteDayTimeChangeListener repeated : %v", lis))
		}
	}
	DayTimeChangeListeners.PushBack(lis)
}

type DayTimeChangeTransactHandler struct {
}

func (this *DayTimeChangeTransactHandler) OnExcute(tNode *transact.TransNode, ud interface{}) transact.TransExeResult {
	err := netlib.UnmarshalPacketNoPackId(ud.([]byte), WGDayTimeChangePack)
	if err == nil {
		if LastDayTimeRec.LastMin != WGDayTimeChangePack.LastMin {
			LastDayTimeRec.LastMin = WGDayTimeChangePack.LastMin
			//SceneMgrSington.OnMiniTimer()
			for e := DayTimeChangeListeners.Front(); e != nil; e = e.Next() {
				if lis, ok := e.Value.(DayTimeChangeListener); ok {
					utils.CatchPanic(func() { lis.OnMiniTimer() })
				}
			}
		}
		if LastDayTimeRec.LastHour != WGDayTimeChangePack.LastHour {
			LastDayTimeRec.LastHour = WGDayTimeChangePack.LastHour
			//SceneMgrSington.OnHourTimer()
			for e := DayTimeChangeListeners.Front(); e != nil; e = e.Next() {
				if lis, ok := e.Value.(DayTimeChangeListener); ok {
					utils.CatchPanic(func() { lis.OnHourTimer() })
				}
			}
		}
		if LastDayTimeRec.LastDay != WGDayTimeChangePack.LastDay {
			LastDayTimeRec.LastDay = WGDayTimeChangePack.LastDay
			//SceneMgrSington.OnDayTimer()
			for e := DayTimeChangeListeners.Front(); e != nil; e = e.Next() {
				if lis, ok := e.Value.(DayTimeChangeListener); ok {
					utils.CatchPanic(func() { lis.OnDayTimer() })
				}
			}
		}
		if LastDayTimeRec.LastWeek != WGDayTimeChangePack.LastWeek {
			LastDayTimeRec.LastWeek = WGDayTimeChangePack.LastWeek
			//SceneMgrSington.OnWeekTimer()
			for e := DayTimeChangeListeners.Front(); e != nil; e = e.Next() {
				if lis, ok := e.Value.(DayTimeChangeListener); ok {
					utils.CatchPanic(func() { lis.OnWeekTimer() })
				}
			}
		}
		if LastDayTimeRec.LastMonth != WGDayTimeChangePack.LastMonth {
			LastDayTimeRec.LastMonth = WGDayTimeChangePack.LastMonth
			//SceneMgrSington.OnMonthTimer()
			for e := DayTimeChangeListeners.Front(); e != nil; e = e.Next() {
				if lis, ok := e.Value.(DayTimeChangeListener); ok {
					utils.CatchPanic(func() { lis.OnMonthTimer() })
				}
			}
		}
	}

	return transact.TransExeResult_Success
}

func (this *DayTimeChangeTransactHandler) OnCommit(tNode *transact.TransNode) transact.TransExeResult {
	//logger.Logger.Trace("DayTimeChangeTransactHandler.OnCommit ")
	return transact.TransExeResult_Success
}

func (this *DayTimeChangeTransactHandler) OnRollBack(tNode *transact.TransNode) transact.TransExeResult {
	//logger.Logger.Trace("DayTimeChangeTransactHandler.OnRollBack ")
	return transact.TransExeResult_Success
}

func (this *DayTimeChangeTransactHandler) OnChildTransRep(tNode *transact.TransNode, hChild transact.TransNodeID,
	retCode int, ud interface{}) transact.TransExeResult {
	//logger.Logger.Trace("DayTimeChangeTransactHandler.OnChildTransRep ")
	return transact.TransExeResult_Success
}

func init() {
	transact.RegisteHandler(common.TransType_DayTimeChange, &DayTimeChangeTransactHandler{})
	RegisteDayTimeChangeListener(base.CoinPoolMgr)
	RegisteDayTimeChangeListener(base.SceneMgrSington)
}
