package main

import (
	"time"

	"github.com/idealeak/goserver/core/module"
)

var ClockMgrSington = &ClockMgr{
	LastHour:  -1,
	LastDay:   -1,
	Notifying: false,
}

const (
	CLOCK_EVENT_SECOND int = iota
	CLOCK_EVENT_MINUTE
	CLOCK_EVENT_HOUR
	CLOCK_EVENT_DAY
	CLOCK_EVENT_WEEK
	CLOCK_EVENT_MONTH
	CLOCK_EVENT_SHUTDOWN
	CLOCK_EVENT_MAX
)

type ClockSinker interface {
	InterestClockEvent() int
	OnSecTimer()
	OnMiniTimer()
	OnHourTimer()
	OnDayTimer()
	OnWeekTimer()
	OnMonthTimer()
	OnShutdown()
}

type BaseClockSinker struct {
}

func (s *BaseClockSinker) InterestClockEvent() int { return 0 }
func (s *BaseClockSinker) OnSecTimer()             {}
func (s *BaseClockSinker) OnMiniTimer()            {}
func (s *BaseClockSinker) OnHourTimer()            {}
func (s *BaseClockSinker) OnDayTimer()             {}
func (s *BaseClockSinker) OnWeekTimer()            {}
func (s *BaseClockSinker) OnMonthTimer()           {}
func (s *BaseClockSinker) OnShutdown()             {}

type ClockMgr struct {
	sinkers     [CLOCK_EVENT_MAX][]ClockSinker
	LastTime    time.Time
	LastMonth   time.Month
	LastWeek    int
	LastDay     int
	LastHour    int
	LastMini    int
	LastSec     int
	Notifying   bool
	LastFiveMin int
}

func (this *ClockMgr) RegisteSinker(sinker ClockSinker) {
	interest := sinker.InterestClockEvent()
	for i := 0; i < CLOCK_EVENT_MAX; i++ {
		if (1<<i)&interest != 0 {
			found := false
			ss := this.sinkers[i]
			for _, s := range ss {
				if s == sinker {
					found = true
					break
				}
			}
			if !found {
				this.sinkers[i] = append(this.sinkers[i], sinker)
			}
		}
	}
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
	this.LastFiveMin = -1
}

func (this *ClockMgr) Update() {
	tNow := time.Now().Local()
	sec := tNow.Second()
	if sec != this.LastSec {
		this.LastSec = sec
		this.fireSecondEvent()

		min := tNow.Minute()
		if min != this.LastMini {
			this.LastMini = min
			this.fireMinuteEvent()

			hour := tNow.Hour()
			if hour != this.LastHour {
				this.LastHour = hour
				this.fireHourEvent()

				day := tNow.Day()
				if day != this.LastDay {
					this.LastDay = day
					this.fireDayEvent()

					_, week := tNow.ISOWeek()
					if week != this.LastWeek {
						this.LastWeek = week
						this.fireWeekEvent()
					}

					month := tNow.Month()
					if month != this.LastMonth {
						this.LastMonth = month
						this.fireMonthEvent()
					}
				}
			}
		}
		if tNow.Sub(this.LastTime) >= time.Minute*30 {
			this.LastTime = tNow
		}
	}
}

func (this *ClockMgr) Shutdown() {
	this.fireShutdownEvent()
	module.UnregisteModule(this)
}

func (this *ClockMgr) fireSecondEvent() {
	for _, s := range this.sinkers[CLOCK_EVENT_SECOND] {
		s.OnSecTimer()
	}
}

func (this *ClockMgr) fireMinuteEvent() {
	for _, s := range this.sinkers[CLOCK_EVENT_MINUTE] {
		s.OnMiniTimer()
	}
}

func (this *ClockMgr) fireHourEvent() {
	for _, s := range this.sinkers[CLOCK_EVENT_HOUR] {
		s.OnHourTimer()
	}
}

func (this *ClockMgr) fireDayEvent() {
	for _, s := range this.sinkers[CLOCK_EVENT_DAY] {
		s.OnDayTimer()
	}
}

func (this *ClockMgr) fireWeekEvent() {
	for _, s := range this.sinkers[CLOCK_EVENT_WEEK] {
		s.OnWeekTimer()
	}
}

func (this *ClockMgr) fireMonthEvent() {
	for _, s := range this.sinkers[CLOCK_EVENT_MONTH] {
		s.OnMonthTimer()
	}
}

func (this *ClockMgr) fireShutdownEvent() {
	for _, s := range this.sinkers[CLOCK_EVENT_SHUTDOWN] {
		s.OnShutdown()
	}
}

func (this *ClockMgr) GetLast() (int, int, int, int, int, int) {
	return this.LastSec, this.LastMini, this.LastHour, this.LastDay, this.LastWeek, int(this.LastMonth)
}

func init() {
	module.RegisteModule(ClockMgrSington, time.Millisecond*500, 0)
}
