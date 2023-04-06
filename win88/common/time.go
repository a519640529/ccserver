package common

import (
	"time"

	"github.com/idealeak/goserver/core/timer"
)

type WGDayTimeChange struct {
	LastMin   int
	LastHour  int
	LastDay   int
	LastWeek  int
	LastMonth int
}

func InSameDay(tNow, tPre time.Time) bool {
	if tPre.IsZero() {
		return true
	}

	if tNow.Day() != tPre.Day() {
		return false
	}

	if tNow.Sub(tPre) < time.Hour*24 {
		return true
	}
	return false
}
func InSameDayNoZero(tNow, tPre time.Time) bool {

	if tNow.Day() != tPre.Day() {
		return false
	}

	if tNow.Sub(tPre) < time.Hour*24 {
		return true
	}
	return false
}

func TsInSameDay(tsNow, tsPre int64) bool {
	tNow := time.Unix(tsNow, 0)
	tPre := time.Unix(tsPre, 0)
	return InSameDay(tNow, tPre)
}

func IsContinueDay(tNow, tPre time.Time) bool {
	if tPre.IsZero() {
		return true
	}
	tNext := tPre.AddDate(0, 0, 1)
	if InSameDay(tNow, tNext) {
		return true
	}
	return false
}

func InSameMonth(tNow, tPre time.Time) bool {
	if tPre.IsZero() {
		return true
	}

	if tNow.Month() != tPre.Month() {
		return false
	}
	if tNow.Year() == tPre.Year() {
		return true
	}
	return false
}

func InSameWeek(tNow, tPre time.Time) bool {
	if tPre.IsZero() {
		return true
	}

	preYear, preWeek := tPre.ISOWeek()
	nowYear, nowWeek := tNow.ISOWeek()
	if preYear == nowYear && preWeek == nowWeek {
		return true
	}
	return false
}

func DiffDay(tNow, tPre time.Time) int {
	y, m, d := tPre.Date()
	tStart := time.Date(y, m, d, 0, 0, 0, 0, tPre.Location())
	return int(tNow.Sub(tStart) / (time.Hour * 24))
}

func DiffMonth(tNow, tPre time.Time) int {
	y1, m1, _ := tNow.Date()
	y2, m2, _ := tPre.Date()
	diffMonth := (y1-y2)*12 + (int(m1) - int(m2))
	return int(diffMonth)
}

func DelayInvake(method func(), ud interface{}, delay time.Duration, times int) (timer.TimerHandle, bool) {
	return timer.StartTimer(timer.TimerActionWrapper(func(h timer.TimerHandle, ud interface{}) bool {
		if method != nil {
			method()
		}
		return true
	}), ud, delay, times)
}

func StrTimeToTs(strTime string) int64 {
	rTime, err := time.ParseInLocation("2006-01-02 15:04:05", strTime, time.Local)
	if err != nil {
		return 0
	}
	return rTime.Unix()
}
func TsToStrTime(tc int64) string {
	//return time.Now().Format("2018-07-02 19:14:00")
	return time.Unix(tc, 0).Format("2006-01-02 15:04:05")
}
func TsToStrDateTime(tc int64) string {
	//return time.Now().Format("2018-07-02 19:14:00")
	return time.Unix(tc, 0).Format("2006-01-02")
}

func InTimeRange(beginHour, beginMinute, endHour, endMinute, checkHour, checkMinute int32) bool {
	beginTime := beginHour*100 + beginMinute
	endTime := endHour*100 + endMinute
	checkTime := checkHour*100 + checkMinute
	return beginTime <= checkTime && checkTime <= endTime
}
