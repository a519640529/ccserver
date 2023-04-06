package main

import (
	"games.yol.com/win88/protocol/server"
	"games.yol.com/win88/srvdata"
	"games.yol.com/win88/webapi"
)

var (
	ThirdPltGameMappingConfig = &ThirdPlatformGameMappingConfiguration{
		DB_ThirdPlatformGameMappingMgr: srvdata.PBDB_ThirdPlatformGameMappingMgr,
		GamefreeIdMappingMap:           make(map[int32]*server.DB_ThirdPlatformGameMapping),
	}
)

type ThirdPlatformGameMappingConfiguration struct {
	*srvdata.DB_ThirdPlatformGameMappingMgr
	GamefreeIdMappingMap map[int32]*server.DB_ThirdPlatformGameMapping
}

func (this *ThirdPlatformGameMappingConfiguration) Init() {
	// this.Test()
	var rawMappingInfo = make(map[int32]*webapi.WebAPI_ThirdPlatformGameMapping)
	for _, v := range this.Datas.Arr {
		this.GamefreeIdMappingMap[v.GetSystemGameID()] = v
		rawMappingInfo[v.GetSystemGameID()] = &webapi.WebAPI_ThirdPlatformGameMapping{
			GameFreeID:            v.GetSystemGameID(),
			ThirdPlatformName:     v.GetThirdPlatformName(),
			ThirdGameID:           v.GetThirdGameID(),
			Desc:                  v.GetDesc(),
			ScreenOrientationType: v.GetScreenOrientationType(),
			ThirdID:               v.GetThirdID(),
		}
	}
	webapi.ThridPlatformMgrSington.ThridPlatformMap.Range(func(key, value interface{}) bool {
		value.(webapi.IThirdPlatform).InitMappingRelation(rawMappingInfo)
		return true
	})
}

func (this *ThirdPlatformGameMappingConfiguration) Reload() {
	//todo 缓存数据加快查找
	//logger.Logger.Info("=== 缓存三方平台游戏id映射关系数据加快查找===")
	this.GamefreeIdMappingMap = make(map[int32]*server.DB_ThirdPlatformGameMapping)
	var rawMappingInfo = make(map[int32]*webapi.WebAPI_ThirdPlatformGameMapping)
	for _, v := range this.Datas.Arr {
		this.GamefreeIdMappingMap[v.GetSystemGameID()] = v
		rawMappingInfo[v.GetSystemGameID()] = &webapi.WebAPI_ThirdPlatformGameMapping{
			GameFreeID:            v.GetSystemGameID(),
			ThirdPlatformName:     v.GetThirdPlatformName(),
			ThirdGameID:           v.GetThirdGameID(),
			Desc:                  v.GetDesc(),
			ScreenOrientationType: v.GetScreenOrientationType(),
			ThirdID:               v.GetThirdID(),
		}
	}
	webapi.ThridPlatformMgrSington.ThridPlatformMap.Range(func(key, value interface{}) bool {
		value.(webapi.IThirdPlatform).InitMappingRelation(rawMappingInfo)
		return true
	})
}

func (this *ThirdPlatformGameMappingConfiguration) Test() {
	var rawMappingInfo = make(map[int32]*webapi.WebAPI_ThirdPlatformGameMapping)
	v := &server.DB_ThirdPlatformGameMapping{
		Id:                    1,
		SystemGameID:          9010001,
		ThirdPlatformName:     "测试平台",
		ThirdGameID:           "901",
		Desc:                  "",
		ScreenOrientationType: 0,
		ThirdID:               901,
	}
	this.GamefreeIdMappingMap[v.GetSystemGameID()] = v
	rawMappingInfo[v.GetSystemGameID()] = &webapi.WebAPI_ThirdPlatformGameMapping{
		GameFreeID:            v.GetSystemGameID(),
		ThirdPlatformName:     v.GetThirdPlatformName(),
		ThirdGameID:           v.GetThirdGameID(),
		Desc:                  v.GetDesc(),
		ScreenOrientationType: v.GetScreenOrientationType(),
		ThirdID:               v.GetThirdID(),
	}

	webapi.ThridPlatformMgrSington.ThridPlatformMap.Range(func(key, value interface{}) bool {
		value.(webapi.IThirdPlatform).InitMappingRelation(rawMappingInfo)
		return true
	})
}

func (this *ThirdPlatformGameMappingConfiguration) FindByGameID(gamefreeId int32) (ok bool, item *server.DB_ThirdPlatformGameMapping) {
	item, ok = this.GamefreeIdMappingMap[gamefreeId]
	return
}

// 包含dg的查询
func (this *ThirdPlatformGameMappingConfiguration) FindSystemGamefreeidByThirdGameInfo(thirdPlt string, inThirdGameId, inThirdGameName string) (gamefreeid int32) {
	if v, exist := webapi.ThridPlatformMgrSington.ThridPlatformMap.Load(thirdPlt); exist {
		return v.(webapi.IThirdPlatform).ThirdGameInfo2GamefreeId(&webapi.WebAPI_ThirdPlatformGameMapping{
			ThirdPlatformName: thirdPlt,
			ThirdGameID:       inThirdGameId,
			Desc:              inThirdGameName,
		})
	}
	return 0
}
func (this *ThirdPlatformGameMappingConfiguration) FindThirdIdByThird(thirdName string) (thirdId int32) {
	if v, exist := webapi.ThridPlatformMgrSington.ThridPlatformMap.Load(thirdName); exist {
		if plt, ok := v.(webapi.IThirdPlatform); ok {
			return int32(plt.GetPlatformBase().BaseGameID)
		}
	}
	return 0
}

// gamefreeid与DG视讯gameid映射关系
var g_GamefreeidMappingDGgameid = map[int32]int32{
	280010001: 1,  //视讯百家乐
	280020001: 3,  //视讯龙虎斗
	280030001: 7,  //视讯牛牛
	280040001: 4,  //视讯轮盘
	280050001: 5,  //视讯骰宝
	280060001: 11, //视讯炸金花
}
