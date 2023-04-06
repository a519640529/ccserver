package srvdata

import (
	"strings"
)

var DataMgr = &dataMgr{
	loaders:         make(map[string]DataLoader),
	cacheGameFreeId: make(map[int32]CacheGameType),
}

type dataMgr struct {
	loaders         map[string]DataLoader
	cacheGameFreeId map[int32]CacheGameType
}
type CacheGameType struct {
	Ids      []int32
	GameType int32
}

func (this *dataMgr) RegisteLoader(name string, loader DataLoader) {
	this.loaders[strings.ToLower(name)] = loader
}

func (this *dataMgr) GetLoader(name string) DataLoader {
	if loader, exist := this.loaders[strings.ToLower(name)]; exist {
		return loader
	}
	return nil
}

func (this *dataMgr) GetGameFreeIds(gameId, gameMode int32) (ids []int32, gameType int32) {
	key := gameId<<16 | gameMode
	if data, exist := this.cacheGameFreeId[key]; exist {
		return data.Ids, data.GameType
	} else {
		for _, dbGameFree := range PBDB_GameFreeMgr.Datas.Arr {
			if dbGameFree.GetGameId() == gameId && dbGameFree.GetGameMode() == gameMode {
				ids = append(ids, dbGameFree.GetId())
				gameType = dbGameFree.GetGameType()
			}
		}
		this.cacheGameFreeId[key] = CacheGameType{Ids: ids, GameType: gameType}
	}
	return
}
