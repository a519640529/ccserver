package base

import (
	"time"

	"games.yol.com/win88/common"
	"github.com/idealeak/goserver/core/module"
)

var ClockMgrSington = &ClockMgr{
	LastHour:  -1,
	LastDay:   -1,
	Notifying: false,
}

var LastDayTimeRec = common.WGDayTimeChange{}

type ClockMgr struct {
	LastTime  time.Time
	LastMonth time.Month
	LastWeek  int
	LastDay   int
	LastHour  int
	LastMini  int
	LastSec   int
	Notifying bool
}

func (this *ClockMgr) ModuleName() string {
	return "ClockMgr"
}

func (this *ClockMgr) Init() {
	tNow := time.Now().Local()
	this.LastTime = tNow
	_, this.LastMonth, this.LastDay = tNow.Date()
	this.LastHour, this.LastMini, this.LastSec = tNow.Hour(), tNow.Minute(), tNow.Second()
	_, this.LastWeek = tNow.ISOWeek()
}

func (this *ClockMgr) Update() {
	tNow := module.AppModule.GetCurrTime()
	sec := tNow.Second()
	if sec != this.LastSec {
		this.LastSec = sec
		PlayerMgrSington.OnSecondTimer()
		min := tNow.Minute()
		if min != this.LastMini {
			this.LastMini = min
			PlayerMgrSington.OnMiniTimer()
			hour := tNow.Hour()
			if hour != this.LastHour {
				ClientMgrSington.Update()
				this.LastHour = hour
				day := tNow.Day()
				if day != this.LastDay {
					ClientMgrSington.AccountValideCheck()
					this.LastDay = day
					_, week := tNow.ISOWeek()
					if week != this.LastWeek {
						this.LastWeek = week
					}
					month := tNow.Month()
					if month != this.LastMonth {
						this.LastMonth = month
					}
				}
			}
		}
	}
	PlayerMgrSington.OnHalfSecondTimer()
}

func (this *ClockMgr) Shutdown() {
	module.UnregisteModule(this)
}

func init() {
	module.RegisteModule(ClockMgrSington, time.Millisecond*500, 0)
}
