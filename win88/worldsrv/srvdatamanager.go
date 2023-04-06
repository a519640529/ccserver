package main

import (
	"games.yol.com/win88/protocol/server"
	"games.yol.com/win88/srvdata"
	"math/rand"
)

var SrvDataMgrEx = &SrvDataManagerEx{
	DataReloader: make(map[string]SrvDataReloadInterface),
}

type SrvDataManagerEx struct {
	DataReloader map[string]SrvDataReloadInterface
}

func RegisterDataReloader(fileName string, sdri SrvDataReloadInterface) {
	SrvDataMgrEx.DataReloader[fileName] = sdri
}

type SrvDataReloadInterface interface {
	Reload()
}

var DgConfigMgrEx = &PBDB_DgConfigMgrEx{}

type PBDB_DgConfigMgrEx struct {
}

func (this *PBDB_DgConfigMgrEx) Reload() {
}

var GameFreeMgrEx = &PBDB_GameFreeMgrEx{
	DbGameFreeId:   make(map[string]int32),
	DbGameDif:      make(map[int32]map[int32]string),
	DbGameDifByGFI: make(map[int32]string),
	DbGameFreeMgr:  make(map[string]*server.DB_GameFree),
}

type PBDB_GameFreeMgrEx struct {
	DbGameFreeId   map[string]int32               //key为GameDif
	DbGameDif      map[int32]map[int32]string     //key为 GameId GameMode
	DbGameDifByGFI map[int32]string               //key 为 DBGameFreeId
	DbGameFreeMgr  map[string]*server.DB_GameFree //key为GameDif
}

func (this *PBDB_GameFreeMgrEx) Reload() {
	for _, gfm := range srvdata.PBDB_GameFreeMgr.Datas.Arr {
		if _, ok := this.DbGameFreeId[gfm.GetGameDif()]; !ok {
			this.DbGameFreeId[gfm.GetGameDif()] = gfm.GetId()
		}
		if _, ok := this.DbGameDif[gfm.GetGameId()]; !ok {
			this.DbGameDif[gfm.GetGameId()] = make(map[int32]string)
		}
		this.DbGameDif[gfm.GetGameId()][gfm.GetGameMode()] = gfm.GetGameDif()
		this.DbGameDifByGFI[gfm.GetId()] = gfm.GetGameDif()
		this.DbGameFreeMgr[gfm.GetGameDif()] = gfm
	}
}

// 查询当前游戏的DBGameFreeMgr
func (this *PBDB_GameFreeMgrEx) GetDBGameFreeMgrByGameDif(gameDif string) *server.DB_GameFree {
	return this.DbGameFreeMgr[gameDif]
}

// 查询当前游戏的GameDif
func (this *PBDB_GameFreeMgrEx) GetGameDifByGameFreeId(dbGameFreeId int32) string {
	if str, ok := this.DbGameDifByGFI[dbGameFreeId]; ok {
		return str
	}
	return ""
}

// 查询当前游戏的GameDif
func (this *PBDB_GameFreeMgrEx) GetGameDifByGameIdAndMode(gameId, gameMode int32) string {
	if data, ok := this.DbGameDif[gameId]; ok {
		if str, ok2 := data[gameMode]; ok2 {
			return str
		}
	}
	return ""
}

// 查询当前游戏的GameFreeId
func (this *PBDB_GameFreeMgrEx) GetGameFreeIdByGameDif(gameDif string) int32 {
	if d, ok := this.DbGameFreeId[gameDif]; ok {
		return d
	}
	return 0
}

var RobotCarryMgrEx = &PBDB_RobotGameMgrEx{
	pool: make(map[int32][]*server.DB_RobotGame),
}

type PBDB_RobotGameMgrEx struct {
	pool map[int32][]*server.DB_RobotGame
}

func (this *PBDB_RobotGameMgrEx) Reload() {
	for _, item := range srvdata.PBDB_RobotGameMgr.Datas.Arr {
		if pp, exist := this.pool[item.GetId()]; exist {
			pp = append(pp, item)
			this.pool[item.GetId()] = pp
		} else {
			this.pool[item.GetId()] = []*server.DB_RobotGame{item}
		}
	}
}

func (this *PBDB_RobotGameMgrEx) RandOneCarry(gamefreeId int32) (int32, int32, int32, bool) {
	if pp, exist := this.pool[gamefreeId]; exist {
		if len(pp) > 0 {
			item := pp[rand.Intn(len(pp))]
			enterCoin := item.GetEnterCoin()
			//尽量避免有重复的
			if enterCoin > 100000000 { //100W以上
				enterCoin = rand.Int31n(enterCoin%100000000+1) + enterCoin/100000000*100000000
			} else if enterCoin > 10000000 { //10W以上
				enterCoin = rand.Int31n(enterCoin%10000000+1) + enterCoin/10000000*10000000
			} else if enterCoin > 1000000 { //1W以上
				enterCoin = rand.Int31n(enterCoin%1000000+1) + enterCoin/1000000*1000000
			} else if enterCoin > 100000 { //1k以上
				enterCoin = rand.Int31n(enterCoin%100000+1) + enterCoin/100000*100000
			} else if enterCoin > 10000 { //1百以上
				enterCoin = rand.Int31n(enterCoin%10000+1) + enterCoin/10000*10000
			} else if enterCoin > 1000 { //十以上
				enterCoin = rand.Int31n(enterCoin%1000+1) + enterCoin/1000*1000
			} else if enterCoin > 100 { //1以上
				enterCoin = rand.Int31n(enterCoin%100+1) + enterCoin/100*100
			}
			return enterCoin, item.GetLeaveCoin(), item.GetGameTimes(), true
		}
	}
	return 0, 0, 0, false
}

func init() {
	srvdata.SrvDataModifyCB = func(fileName string, fullName string) {
		if dr, ok := SrvDataMgrEx.DataReloader[fileName]; ok {
			dr.Reload()
		}
	}
	//RegisterDataReloader("DB_SystemChance.dat", DgConfigMgrEx)
	RegisterDataReloader("DB_GameFree.dat", GameFreeMgrEx)
	RegisterDataReloader("DB_ThirdPlatformGameMapping.dat", ThirdPltGameMappingConfig)
	RegisterDataReloader("DB_RobotGame.dat", RobotCarryMgrEx)
	//RegisterDataReloader("DB_GameRule.dat", GameFreeMgrEx)
}
