package main

import (
	"games.yol.com/win88/common"
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	webapi_proto "games.yol.com/win88/protocol/webapi"
	"games.yol.com/win88/srvdata"
	"games.yol.com/win88/webapi"
	"github.com/idealeak/goserver/core/logger"
)

type PlatformGameGroupObserver interface {
	OnGameGroupUpdate(oldCfg, newCfg *webapi_proto.GameConfigGroup)
}

var PlatformGameGroupMgrSington = &PlatformGameGroupMgr{
	groups: make(map[int32]*webapi_proto.GameConfigGroup),
}

type PlatformGameGroupMgr struct {
	groups    map[int32]*webapi_proto.GameConfigGroup
	observers []PlatformGameGroupObserver
}

func (this *PlatformGameGroupMgr) RegisteObserver(observer PlatformGameGroupObserver) {
	for _, ob := range this.observers {
		if ob == observer {
			return
		}
	}
	this.observers = append(this.observers, observer)
}

func (this *PlatformGameGroupMgr) UnregisteObserver(observer PlatformGameGroupObserver) {
	for i, ob := range this.observers {
		if ob == observer {
			count := len(this.observers)
			if i == 0 {
				this.observers = this.observers[1:]
			} else if i == count-1 {
				this.observers = this.observers[:count-1]
			} else {
				arr := this.observers[:i]
				arr = append(arr, this.observers[i+1:]...)
				this.observers = arr
			}
		}
	}
}

func (this *PlatformGameGroupMgr) GetGameGroup(groupId int32) *webapi_proto.GameConfigGroup {
	if g, exist := this.groups[groupId]; exist {
		return g
	}
	return nil
}

func (this *PlatformGameGroupMgr) LoadGameGroup() {
	//不使用etcd的情况下走api获取
	if model.GameParamData.UseEtcd {
		EtcdMgrSington.InitGameGroup()
	} else {
		//获取平台游戏组信息
		logger.Logger.Trace("API_GetGameGroupData")
		buf, err := webapi.API_GetGameGroupData(common.GetAppId())
		if err == nil {
			pcdr := &webapi_proto.ASGameConfigGroup{}
			err = proto.Unmarshal(buf, pcdr)
			if err == nil && pcdr.Tag == webapi_proto.TagCode_SUCCESS {
				gameCfgGroup := pcdr.GetGameConfigGroup()
				for _, value := range gameCfgGroup {
					groupId := value.GetId()
					vDbGameFree := value.GetDbGameFree()

					dbGameFree := srvdata.PBDB_GameFreeMgr.GetData(value.LogicId)
					if dbGameFree == nil {
						continue
					}
					if vDbGameFree == nil {
						vDbGameFree = dbGameFree
					} else {
						CopyDBGameFreeField(dbGameFree, vDbGameFree)
					}

					this.groups[groupId] = value
					logger.Logger.Trace("PlatformGameGroup data:", value)
				}
			} else {
				logger.Logger.Error("Unmarshal PlatformGameGroup data error:", err)
			}
		} else {
			logger.Logger.Error("Get PlatformGameGroup data error:", err)
		}
	}
}

func (this *PlatformGameGroupMgr) UpsertGameGroup(conf *webapi_proto.GameConfigGroup) {
	if conf == nil {
		return
	}
	dbGameFree := srvdata.PBDB_GameFreeMgr.GetData(conf.GetLogicId())
	if dbGameFree == nil {
		return
	}
	if conf.DbGameFree == nil {
		conf.DbGameFree = dbGameFree
	} else {
		CopyDBGameFreeField(dbGameFree, conf.DbGameFree)
	}
	old, ok := this.groups[conf.GetId()]

	//更新相关配置
	this.groups[conf.GetId()] = conf
	//发布更新事件
	if ok && old != nil {
		this.OnGameGroupUpdate(old, conf)
	}
}

func (this *PlatformGameGroupMgr) OnGameGroupUpdate(oldCfg, newCfg *webapi_proto.GameConfigGroup) {
	for _, observer := range this.observers {
		observer.OnGameGroupUpdate(oldCfg, newCfg)
	}
}
