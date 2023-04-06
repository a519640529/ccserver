package main

import (
	"games.yol.com/win88/model"
	"games.yol.com/win88/protocol/server"
	"games.yol.com/win88/protocol/webapi"
	"github.com/globalsign/mgo/bson"
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/module"
	"github.com/idealeak/goserver/core/task"
	"time"
)

type PlayerSingleAdjustManager struct {
	AdjustData     map[uint64]*model.PlayerSingleAdjust
	dirtyList      map[uint64]bool
	cacheDirtyList map[uint64]bool //缓存待删除数据
}

var PlayerSingleAdjustMgr = &PlayerSingleAdjustManager{
	AdjustData:     make(map[uint64]*model.PlayerSingleAdjust),
	dirtyList:      make(map[uint64]bool),
	cacheDirtyList: make(map[uint64]bool),
}

func (this *PlayerSingleAdjustManager) WebData(msg *webapi.ASSinglePlayerAdjust, p *Player) (sa *webapi.PlayerSingleAdjust) {
	psa := model.WebSingleAdjustToModel(msg.PlayerSingleAdjust)
	switch msg.Opration {
	case 1:
		this.AddNewSingleAdjust(psa)
	case 2:
		this.EditSingleAdjust(psa)
	case 3:
		this.DeleteSingleAdjust(psa.Platform, psa.SnId, psa.GameFreeId)
	case 4:
		sa = this.WebGetSingleAdjust(psa.Platform, psa.SnId, psa.GameFreeId)
		return
	}
	//同步到游服
	if p != nil {
		if p.scene != nil && p.scene.dbGameFree.Id == psa.GameFreeId {
			gss := GameSessMgrSington.GetGameServerSess(int(psa.GameId))
			pack := &server.WGSingleAdjust{
				SceneId:            int32(p.scene.sceneId),
				Option:             msg.Opration,
				PlayerSingleAdjust: model.MarshalSingleAdjust(psa),
			}
			for _, gs := range gss {
				gs.Send(int(server.SSPacketID_PACKET_WG_SINGLEADJUST), pack)
			}
		}
		if p.miniScene != nil {
			for _, game := range p.miniScene {
				if game.dbGameFree.Id == psa.GameFreeId {
					gss := GameSessMgrSington.GetGameServerSess(int(psa.GameId))
					pack := &server.WGSingleAdjust{
						SceneId:            int32(game.sceneId),
						Option:             msg.Opration,
						PlayerSingleAdjust: model.MarshalSingleAdjust(psa),
					}
					for _, gs := range gss {
						gs.Send(int(server.SSPacketID_PACKET_WG_SINGLEADJUST), pack)
					}
					break
				}
			}
		}
	}
	return
}

func (this *PlayerSingleAdjustManager) IsSingleAdjustPlayer(snid int32, gameFreeId int32) (*model.PlayerSingleAdjust, bool) {
	key := uint64(snid)<<32 + uint64(gameFreeId)
	if data, ok := this.AdjustData[key]; ok {
		if data.CurTime < data.TotalTime {
			return data, true
		}
	}
	return nil, false
}
func (this *PlayerSingleAdjustManager) AddAdjustCount(snid int32, gameFreeId int32) {
	key := uint64(snid)<<32 + uint64(gameFreeId)
	if ad, ok := this.AdjustData[key]; ok {
		ad.CurTime++
		this.dirtyList[key] = true
	}
}
func (this *PlayerSingleAdjustManager) GetSingleAdjust(platform string, snid, gameFreeId int32) *model.PlayerSingleAdjust {
	key := uint64(snid)<<32 + uint64(gameFreeId)
	if psa, ok := this.AdjustData[key]; ok {
		return psa
	}
	return nil
}
func (this *PlayerSingleAdjustManager) WebGetSingleAdjust(platform string, snid, gameFreeId int32) *webapi.PlayerSingleAdjust {
	key := uint64(snid)<<32 + uint64(gameFreeId)
	if psa, ok := this.AdjustData[key]; ok {
		return &webapi.PlayerSingleAdjust{
			Id:            psa.Id.Hex(),
			Platform:      psa.Platform,
			GameFreeId:    psa.GameFreeId,
			SnId:          psa.SnId,
			Mode:          psa.Mode,
			TotalTime:     psa.TotalTime,
			CurTime:       psa.CurTime,
			BetMin:        psa.BetMin,
			BetMax:        psa.BetMax,
			BankerLoseMin: psa.BankerLoseMin,
			BankerWinMin:  psa.BankerWinMin,
			CardMin:       psa.CardMin,
			CardMax:       psa.CardMax,
			Priority:      psa.Priority,
			WinRate:       psa.WinRate,
			GameId:        psa.GameId,
			GameMode:      psa.GameMode,
			Operator:      psa.Operator,
			CreateTime:    psa.CreateTime,
			UpdateTime:    psa.UpdateTime,
		}
	}
	return nil
}
func (this *PlayerSingleAdjustManager) AddNewSingleAdjust(psa *model.PlayerSingleAdjust) *model.PlayerSingleAdjust {
	if psa != nil {
		key := uint64(psa.SnId)<<32 + uint64(psa.GameFreeId)
		psa.Id = bson.NewObjectId()
		psa.CreateTime = time.Now().Unix()
		psa.UpdateTime = time.Now().Unix()

		this.AdjustData[key] = psa
		logger.Logger.Trace("SinglePlayerAdjust new:", psa)
		task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
			return model.AddNewSingleAdjust(psa)
		}), nil, "AddNewSingleAdjust").StartByFixExecutor("AddNewSingleAdjust")
	}
	return psa
}
func (this *PlayerSingleAdjustManager) EditSingleAdjust(psa *model.PlayerSingleAdjust) {
	if psa != nil {
		var inGame bool
		psa.UpdateTime = time.Now().Unix()
		for key, value := range this.AdjustData {
			if value.Id == psa.Id {
				var tempKey = key
				if psa.GameFreeId != value.GameFreeId {
					delete(this.AdjustData, key)
					delete(this.dirtyList, key)
					tempKey = uint64(psa.SnId)<<32 + uint64(psa.GameFreeId)
				}
				this.AdjustData[tempKey] = psa
				this.dirtyList[tempKey] = true
				inGame = true
				break
			}
		}
		logger.Logger.Trace("SinglePlayerAdjust edit:", *psa)
		if !inGame {
			//不在游戏 直接更新库
			task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
				return model.EditSingleAdjust(psa)
			}), nil, "EditSingleAdjust").StartByFixExecutor("EditSingleAdjust")
		}
	}
}
func (this *PlayerSingleAdjustManager) DeleteSingleAdjust(platform string, snid, gameFreeId int32) {
	key := uint64(snid)<<32 + uint64(gameFreeId)
	if _, ok := this.AdjustData[key]; ok {
		delete(this.AdjustData, key)
		delete(this.dirtyList, key)
	}
	logger.Logger.Trace("SinglePlayerAdjust delete:", snid, gameFreeId)
	task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
		return model.DeleteSingleAdjust(&model.PlayerSingleAdjust{SnId: snid, GameFreeId: gameFreeId, Platform: platform})
	}), nil, "DeleteSingleAdjust").Start()
}

func (this *PlayerSingleAdjustManager) ModuleName() string {
	return "PlayerSingleAdjustManager"
}
func (this *PlayerSingleAdjustManager) Init() {
	//data, err := model.QueryAllSingleAdjust("1")
	//if err != nil {
	//	logger.Logger.Warn("QueryAllSingleAdjust is err:", err)
	//	return
	//}
	//if len(data) > 0 {
	//	for _, psa := range data {
	//		_, gameType := srvdata.DataMgr.GetGameFreeIds(psa.GameId, psa.GameMode)
	//		if gameType != common.GameType_Mini {
	//			key := uint64(psa.SnId)<<32 + uint64(psa.GameFreeId)
	//			this.AdjustData[key] = psa
	//		}
	//	}
	//}
}

// 登录加载
func (this *PlayerSingleAdjustManager) LoadSingleAdjustData(platform string, snid int32) {
	task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
		ret, err := model.QueryAllSingleAdjustByKey(platform, snid)
		if err != nil {
			return nil
		}
		return ret
	}), task.CompleteNotifyWrapper(func(data interface{}, tt task.Task) {
		if data != nil {
			ret := data.([]*model.PlayerSingleAdjust)
			for _, psa := range ret {
				key := uint64(psa.SnId)<<32 + uint64(psa.GameFreeId)
				this.AdjustData[key] = psa
			}
		}
	})).StartByFixExecutor("LoadPlayerSingleAdjust")
}

// 掉线删除
func (this *PlayerSingleAdjustManager) DelPlayerData(platform string, snid int32) {
	for _, psa := range this.AdjustData {
		if psa.Platform == platform && psa.SnId == snid {
			key := uint64(psa.SnId)<<32 + uint64(psa.GameFreeId)
			if this.dirtyList[key] {
				this.cacheDirtyList[key] = true
			} else {
				delete(this.AdjustData, key)
			}
		}
	}
}
func (this *PlayerSingleAdjustManager) Update() {
	if len(this.dirtyList) == 0 {
		return
	}
	var syncArr []*model.PlayerSingleAdjust
	for key, _ := range this.dirtyList {
		syncArr = append(syncArr, this.AdjustData[key])
		delete(this.dirtyList, key)
	}
	task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
		var saveArr [2][]uint64
		for _, value := range syncArr {
			err := model.EditSingleAdjust(value)
			if err != nil {
				logger.Logger.Error("PlayerSingleAdjustManager edit ", err)
				saveArr[0] = append(saveArr[0], uint64(value.SnId)<<32+uint64(value.GameFreeId))
			} else {
				saveArr[1] = append(saveArr[1], uint64(value.SnId)<<32+uint64(value.GameFreeId))
			}
		}
		return saveArr
	}), task.CompleteNotifyWrapper(func(data interface{}, tt task.Task) {
		if saveArr, ok := data.([2][]uint64); ok {
			//失败处理
			for _, key := range saveArr[0] {
				this.dirtyList[key] = true
			}
			//成功处理
			for _, key := range saveArr[1] {
				if this.cacheDirtyList[key] {
					delete(this.cacheDirtyList, key)
					delete(this.AdjustData, key)
				}
			}
		}
		return
	})).StartByFixExecutor("PlayerSingleAdjustManager")
}
func (this *PlayerSingleAdjustManager) Shutdown() {
	module.UnregisteModule(this)
}
func init() {
	module.RegisteModule(PlayerSingleAdjustMgr, time.Minute*5, 0)
}
