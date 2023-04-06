package webapi

import (
	"sync"
	"sync/atomic"
	"time"
)

var ThridPlatformMgrSington = &ThridPlatformMgr{
	billno: time.Now().UnixNano(),
}

type ThridPlatformMgr struct {
	ThridPlatformMap sync.Map
	billno           int64
}

func (this *ThridPlatformMgr) register(p IThirdPlatform) {
	this.ThridPlatformMap.Store(p.GetPlatformBase().Name, p)
}

// 为什么要这样做？
// 避免并发量大的话出现订单号重复的情况
// 但非三方游戏代码中有使用微秒时间戳作为billno，仅仅修改三方游戏的话必然存在碰撞问题，暂时先不要加了
func (this *ThridPlatformMgr) GenerateUniqueBillno() int64 {
	return atomic.AddInt64(&this.billno, 1)
}

// 根据平台ID获得平台map,id=-1则返回全部的平台
func (this *ThridPlatformMgr) FindPlatformByPlatformId(id int32) []IThirdPlatform {
	set := make([]IThirdPlatform, 0)
	if id == -1 {
		this.ThridPlatformMap.Range(func(key, value interface{}) bool {
			set = append(set, value.(IThirdPlatform))
			return true
		})
		return set
	}

	this.ThridPlatformMap.Range(func(key, value interface{}) bool {
		if value.(IThirdPlatform).GetPlatformBase().Id == id {
			set = append(set, value.(IThirdPlatform))
			return false
		}
		return true
	})
	return set
}

// 根据场景ID查找属于哪个平台
func (this *ThridPlatformMgr) FindPlatformByPlatformBaseGameId(baseGameId int) (plt IThirdPlatform) {
	this.ThridPlatformMap.Range(func(key, value interface{}) bool {
		if value.(IThirdPlatform).GetPlatformBase().BaseGameID == baseGameId {
			plt = value.(IThirdPlatform)
			return false
		}
		return true
	})
	return
}
func (this *ThridPlatformMgr) PlatformIsExist(id int32) (isExist bool) {
	isExist = false
	this.ThridPlatformMap.Range(func(key, value interface{}) bool {
		if value.(IThirdPlatform).GetPlatformBase().Id == id {
			isExist = true
			return false
		}
		return true
	})
	return isExist
}

// 获得平台数量
func (this *ThridPlatformMgr) AllPlatformCount() int32 {
	count := int32(0)
	this.ThridPlatformMap.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}
