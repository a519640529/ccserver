package base

import (
	"encoding/json"
	"sync"

	"games.yol.com/win88/model"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/module"
)

const (
	GameRuleCacheDBKey = "GameRuleCacheManager"
)

var gameRuleCacheMgr = &GameRuleCacheManager{
	DataPool: new(sync.Map),
}

type GameRuleCacheManager struct {
	DataPool *sync.Map
}

////////////////////////////////////////////////////////////////////
/// Module Implement [beg]
////////////////////////////////////////////////////////////////////
func (this *GameRuleCacheManager) ModuleName() string {
	return "GameRuleCacheManager"
}

func (this *GameRuleCacheManager) Init() {
	data := model.GetStrKVGameData(GameRuleCacheDBKey)
	coinPool := make(map[int32]int64)
	err := json.Unmarshal([]byte(data), &coinPool)
	if err == nil {
		for key, value := range coinPool {
			this.DataPool.Store(key, value)
		}
	} else {
		logger.Logger.Error("Unmarshal coin pool error:", err)
	}
}

func (this *GameRuleCacheManager) Update() {
}

func (this *GameRuleCacheManager) Shutdown() {
	coinPool := make(map[int32]int64)
	this.DataPool.Range(func(key, value interface{}) bool {
		coinPool[key.(int32)] = value.(int64)
		return true
	})
	buff, err := json.Marshal(coinPool)
	if err == nil {
		model.UptStrKVGameData(GameRuleCacheDBKey, string(buff))
	} else {
		logger.Logger.Error("Marshal coin pool error:", err)
	}
	module.UnregisteModule(this)
}

func init() {
	//module.RegisteModule(gameRuleCacheMgr, time.Hour, 0)
}
