package main

import (
	"games.yol.com/win88/model"
	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/module"
	"github.com/idealeak/goserver/core/task"
	"strconv"
	"time"
)

const (
	INVALID_PLAYERID_CACHE_SEC  int64 = 60
	INVALID_PLAYERID_CACHE_MAX  int   = 100000
	PLAYERCACHESAVERSLICENUMBER int32 = 300
)

var PlayerCacheMgrSington = &PlayerCacheMgr{
	playerMap:        make(map[int32]*PlayerCacheItem),
	playerCbs:        make(map[int32][]func(*PlayerCacheItem, bool, bool)),
	playerInvalidIds: make(map[int32]int64),
	playerWaitClr:    make([]*PlayerCacheItem, 0, 128),
	DbSaver: &DbSaver{
		Tick:  PLAYERCACHESAVERSLICENUMBER,
		index: 0,
		init:  false,
		list:  make([]*SaverArray, PLAYERCACHESAVERSLICENUMBER),
		queue: make([]*BalanceQueue, 10),
		pool:  make(map[SaveTaskHandler]*SaverArray),
	},
}

type PlayerCacheItem struct {
	*model.PlayerData
	lastTs   int64
	isOnline bool
}

func (p *PlayerCacheItem) CanDel() bool {
	//return !p.isOnline && time.Now().Unix()-p.lastTs > int64(PLAYERCACHESAVERSLICENUMBER)
	return true
}

func (p *PlayerCacheItem) Time2Save() {
	if p.CanDel() {
		PlayerCacheMgrSington.playerWaitClr = append(PlayerCacheMgrSington.playerWaitClr, p)
	}
}

type PlayerCacheMgr struct {
	*DbSaver
	playerMap        map[int32]*PlayerCacheItem
	playerCbs        map[int32][]func(*PlayerCacheItem, bool, bool)
	playerInvalidIds map[int32]int64
	playerWaitClr    []*PlayerCacheItem
}

func (c *PlayerCacheMgr) GetFromCache(plt string, snid int32) *PlayerCacheItem {
	if p, exist := c.playerMap[snid]; exist {
		p.lastTs = time.Now().Unix()
		return p
	}
	return nil
}

func (c *PlayerCacheMgr) Get(plt string, snid int32, cb func(*PlayerCacheItem, bool, bool), createIfNotExist bool) {
	if p, exist := c.playerMap[snid]; exist {
		p.lastTs = time.Now().Unix()
		cb(p, false, false)
		return
	}

	//玩家不存在，避免打到db上
	if lastTs, exist := c.playerInvalidIds[snid]; exist {
		if time.Now().Unix()-lastTs < INVALID_PLAYERID_CACHE_SEC {
			cb(nil, false, false)
			return
		}
		delete(c.playerInvalidIds, snid)
	}

	if cbs, exist := c.playerCbs[snid]; exist {
		cbs = append(cbs, cb)
		c.playerCbs[snid] = cbs
		return
	}

	var isnew bool
	c.playerCbs[snid] = []func(*PlayerCacheItem, bool, bool){cb}
	task.New(core.CoreObject(), task.CallableWrapper(func(o *basic.Object) interface{} {
		pi, flag := model.GetPlayerDataBySnId(plt, snid, true, createIfNotExist)
		isnew = flag
		return pi
	}), task.CompleteNotifyWrapper(func(data interface{}, tt task.Task) {
		if pi, ok := data.(*model.PlayerData); ok && pi != nil {
			var item *PlayerCacheItem
			item, exist := c.playerMap[snid]
			if !exist {
				item = &PlayerCacheItem{PlayerData: pi, lastTs: time.Now().Unix(), isOnline: false}
				c.playerMap[snid] = item
				c.RegisterDbSaverTask(item)
			}
			if cbs, exist := c.playerCbs[snid]; exist {
				delete(c.playerCbs, snid)
				for _, cb := range cbs {
					cb(item, true, isnew)
				}
			}
		} else {
			if cbs, exist := c.playerCbs[snid]; exist {
				c.CacheInvalidPlayerId(snid)
				delete(c.playerCbs, snid)
				for _, cb := range cbs {
					cb(nil, true, false)
				}
			}
		}
		return
	}), "PlayerCacheMgr.Get").StartByExecutor(strconv.Itoa(int(snid)))
}

func (c *PlayerCacheMgr) GetMore(plt string, snid []int32, cb func([]*PlayerCacheItem, bool)) {
	isAsyn := false
	count := len(snid)
	result := make([]*PlayerCacheItem, 0, count)
	innerCb := func(item *PlayerCacheItem, asyn, isnew bool) {
		if item != nil {
			result = append(result, item)
		}
		if asyn {
			isAsyn = true
		}
		count--
		if count == 0 {
			cb(result, isAsyn)
		}
	}
	for _, id := range snid {
		c.Get(plt, id, innerCb, false)
	}
}

func (c *PlayerCacheMgr) Handler(p *Player, evt *GameEvent) {
	if evt == nil {
		return
	}

	if item, exist := c.playerMap[p.SnId]; exist {
		switch evt.eventType {
		case GAMEEVENT_LOGIN:
			item.isOnline = true
		case GAMEEVENT_LOGOUT:
			item.isOnline = false
		}
		item.lastTs = time.Now().Unix()
	}
}

func (c *PlayerCacheMgr) CacheInvalidPlayerId(snid int32) {
	if len(c.playerInvalidIds) >= INVALID_PLAYERID_CACHE_MAX {
		for id, _ := range c.playerInvalidIds {
			delete(c.playerInvalidIds, id)
			if len(c.playerInvalidIds) < INVALID_PLAYERID_CACHE_MAX {
				break
			}
		}
	}
	c.playerInvalidIds[snid] = time.Now().Unix()
}

func (c *PlayerCacheMgr) UncacheInvalidPlayerId(snid int32) {
	delete(c.playerInvalidIds, snid)
}

// ===module interface
func (c *PlayerCacheMgr) ModuleName() string {
	return "PlayerCacheMgr"
}

func (c *PlayerCacheMgr) Init() {
	c.DbSaver.Init()
}

func (c *PlayerCacheMgr) Update() {
	c.DbSaver.Update()
	for _, p := range c.playerWaitClr {
		delete(c.playerMap, p.SnId)
		c.UnregisteDbSaveTask(p)
	}
	c.playerWaitClr = c.playerWaitClr[0:0]
}

func (c *PlayerCacheMgr) Shutdown() {
	module.UnregisteModule(c)
}

func init() {
	module.RegisteModule(PlayerCacheMgrSington, time.Second, 0)
	GameEventHandlerMgrSingleton.RegisteGameEventHandler(GAMEEVENT_LOGIN, PlayerCacheMgrSington)
	GameEventHandlerMgrSingleton.RegisteGameEventHandler(GAMEEVENT_LOGOUT, PlayerCacheMgrSington)
}
