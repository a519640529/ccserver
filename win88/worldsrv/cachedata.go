package main

import (
	"fmt"
	"sync"
	"time"
)

const (
	AfterMini = 1 //数据缓存1分钟
	AfterHour = 2 //数据缓存1小时
	AfterDay  = 3 //数据缓存1天
	AfterWeek = 4 //数据缓存1周
	AfterOver = 5 //数据缓存到服务器关闭
)

var CacheDataMgr = &CacheDataManager{
	MiniCache: new(sync.Map),
	HourCache: new(sync.Map),
	DayCache:  new(sync.Map),
	WeekCache: new(sync.Map),
	OverCache: new(sync.Map),
}

type CacheDataManager struct {
	BaseClockSinker
	MiniCache *sync.Map
	HourCache *sync.Map
	DayCache  *sync.Map
	WeekCache *sync.Map
	OverCache *sync.Map
}
type CacheData struct {
	Data interface{}
	Ts   int64
}

func (this *CacheDataManager) addCacheData(timeRange int, key string, data interface{}) {
	switch timeRange {
	case AfterMini:
		this.MiniCache.Store(key, &CacheData{
			Data: data,
			Ts:   time.Now().Add(time.Minute).Unix(),
		})
	case AfterHour:
		this.HourCache.Store(key, &CacheData{
			Data: data,
			Ts:   time.Now().Add(time.Hour).Unix(),
		})
	case AfterDay:
		this.DayCache.Store(key, &CacheData{
			Data: data,
			Ts:   time.Now().Add(time.Hour * 24).Unix(),
		})
	case AfterWeek:
		this.DayCache.Store(key, &CacheData{
			Data: data,
			Ts:   time.Now().Add(time.Hour * 24 * 7).Unix(),
		})
	case AfterOver:
		this.DayCache.Store(key, &CacheData{
			Data: data,
			Ts:   time.Now().Add(time.Hour * 24 * 999).Unix(),
		})
	}
}

// 感兴趣所有clock event
func (this *CacheDataManager) InterestClockEvent() int {
	return 1 << CLOCK_EVENT_MINUTE
}

func (this *CacheDataManager) OnMiniTimer() {
	now := time.Now().Unix()
	this.MiniCache.Range(func(key, value interface{}) bool {
		if data, ok := value.(*CacheData); ok {
			if data.Ts < now {
				this.MiniCache.Delete(key)
			}
		}
		return true
	})
	this.HourCache.Range(func(key, value interface{}) bool {
		if data, ok := value.(*CacheData); ok {
			if data.Ts < now {
				this.HourCache.Delete(key)
			}
		}
		return true
	})
	this.DayCache.Range(func(key, value interface{}) bool {
		if data, ok := value.(*CacheData); ok {
			if data.Ts < now {
				this.DayCache.Delete(key)
			}
		}
		return true
	})
	this.WeekCache.Range(func(key, value interface{}) bool {
		if data, ok := value.(*CacheData); ok {
			if data.Ts < now {
				this.WeekCache.Delete(key)
			}
		}
		return true
	})
}

/*
 * 缓存账单号，避免有重复的账单号提交
 */
func (this *CacheDataManager) CacheBillNumber(billNo int, platform string) {
	key := fmt.Sprintf("BillNo-%v-%v", billNo, platform)
	this.addCacheData(AfterHour, key, key)
}
func (this *CacheDataManager) CacheBillCheck(billNo int, platform string) bool {
	key := fmt.Sprintf("BillNo-%v-%v", billNo, platform)
	if _, ok := this.HourCache.Load(key); ok {
		return true
	} else {
		return false
	}
}
func (this *CacheDataManager) ClearCacheBill(billNo int, platform string) {
	key := fmt.Sprintf("BillNo-%v-%v", billNo, platform)
	this.HourCache.Delete(key)
}

func init() {
	ClockMgrSington.RegisteSinker(CacheDataMgr)
}
