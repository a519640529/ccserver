package main

import (
	"games.yol.com/win88/protocol/webapi"
	"math/rand"
	"time"

	"games.yol.com/win88/common"
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	gamehall_proto "games.yol.com/win88/protocol/gamehall"
	server_proto "games.yol.com/win88/protocol/server"
	"games.yol.com/win88/srvdata"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/module"
	"github.com/idealeak/goserver/srvlib"
)

const (
	HundredSceneType_Primary    int = iota //初级
	HundredSceneType_Mid                   //中级
	HundredSceneType_Senior                //高级
	HundredSceneType_Professor             //专家
	HundredSceneType_Experience            //体验场
	HundredSceneType_Max
)

const (
	HundredSceneOp_Enter    int32 = iota //进入
	HundredSceneOp_Leave                 //离开
	HundredSceneOp_Change                //换桌
	HundredSceneOp_Audience              //观战
)

var HundredSceneMgrSington = &HundredSceneMgr{
	//分平台管理
	scenesOfPlatform: make(map[string]map[int32]*Scene),
	platformOfScene:  make(map[int32]string),
	//分组管理
	scenesOfGroup: make(map[int32]map[int32]*Scene),
	groupOfScene:  make(map[int32]int32),
	playerIning:   make(map[int32]int32),
}

type HundredSceneMgr struct {
	//分平台管理
	scenesOfPlatform map[string]map[int32]*Scene
	platformOfScene  map[int32]string
	//分组管理
	scenesOfGroup map[int32]map[int32]*Scene
	groupOfScene  map[int32]int32
	playerIning   map[int32]int32
	canInviteRob  int
}

func (this *HundredSceneMgr) GetPlatformNameBySceneId(sceneid int32) (string, bool) {
	if name, exist := this.platformOfScene[sceneid]; exist {
		return name, exist
	}
	if _, exist := this.groupOfScene[sceneid]; exist {
		s := SceneMgrSington.GetScene(int(sceneid))
		if s != nil && s.limitPlatform != nil {
			return s.limitPlatform.IdStr, true
		}
	}
	return Default_Platform, false
}

func (this *HundredSceneMgr) RebindPlayerSnId(oldSnId, newSnId int32) {
	if id, exist := this.playerIning[oldSnId]; exist {
		delete(this.playerIning, oldSnId)
		this.playerIning[newSnId] = id
	}
	for _, ss := range this.scenesOfPlatform {
		for _, s := range ss {
			s.RebindPlayerSnId(oldSnId, newSnId)
		}
	}
	for _, ss := range this.scenesOfGroup {
		for _, s := range ss {
			s.RebindPlayerSnId(oldSnId, newSnId)
		}
	}
}

func (this *HundredSceneMgr) PlayerEnter(p *Player, id int32) gamehall_proto.OpResultCode_Hundred {
	logger.Logger.Tracef("(this *HundredSceneMgr) PlayerEnter snid:%v id:%v", p.SnId, id)
	if oid, exist := this.playerIning[p.SnId]; exist {
		logger.Logger.Warnf("(this *HundredSceneMgr) PlayerEnter:%v snid:%v find in id:%v PlayerEnter return false", id, p.SnId, oid)
		return gamehall_proto.OpResultCode_Hundred_OPRC_Error_Hundred
	}

	if p.scene != nil {
		logger.Logger.Warnf("(this *HundredSceneMgr) PlayerEnter:%v snid:%v find in id:%v PlayerEnter return false", id, p.SnId, p.scene.sceneId)
		return gamehall_proto.OpResultCode_Hundred_OPRC_Error_Hundred
	}

	if p.isDelete { //删档用户不让进游戏
		return gamehall_proto.OpResultCode_Hundred_OPRC_RoomHadClosed_Hundred
	}

	//多平台支持
	var limitPlatform *Platform
	platformName := Default_Platform
	platform := PlatformMgrSington.GetPlatform(p.Platform)
	if platform != nil && platform.Isolated {
		platformName = platform.IdStr
		limitPlatform = platform
	} else {
		limitPlatform = PlatformMgrSington.GetPlatform(Default_Platform)
	}

	gps := PlatformMgrSington.GetGameConfig(limitPlatform.IdStr, id)
	if gps == nil {
		return gamehall_proto.OpResultCode_Hundred_OPRC_RoomHadClosed_Hundred
	}

	if gps.GroupId != 0 { //按分组进入场景游戏
		pgg := PlatformGameGroupMgrSington.GetGameGroup(gps.GroupId)
		if pgg != nil {
			if _, ok := this.scenesOfGroup[gps.GroupId]; !ok {
				this.scenesOfGroup[gps.GroupId] = make(map[int32]*Scene)

			}
			if ss, ok := this.scenesOfGroup[gps.GroupId]; ok {
				if s, ok := ss[id]; !ok {
					s = this.CreateNewScene(id, gps.GroupId, limitPlatform, pgg.DbGameFree)
					if s != nil {
						ss[id] = s
						this.groupOfScene[int32(s.sceneId)] = gps.GroupId
						logger.Logger.Tracef("(this *HundredSceneMgr) PlayerEnter(groupid=%v) Create %v scene success.", gps.GroupId, id)
					} else {
						logger.Logger.Tracef("(this *HundredSceneMgr) PlayerEnter(groupid=%v) Create %v scene failed.", gps.GroupId, id)
					}
				}
				//尝试进入
				if s, ok := ss[id]; ok && s != nil {
					if s.PlayerEnter(p, -1, true) {
						this.OnPlayerEnter(p, id)
						return gamehall_proto.OpResultCode_Hundred_OPRC_Sucess_Hundred
					} else {
						logger.Logger.Warnf("(this *HundredSceneMgr) PlayerEnter(groupid=%v) enter %v scene failed.", gps.GroupId, id)
					}
				} else {
					logger.Logger.Warnf("(this *HundredSceneMgr) PlayerEnter(groupid=%v) get %v scene failed.", gps.GroupId, id)
				}
			}
			logger.Logger.Warnf("(this *HundredSceneMgr) PlayerEnter(groupid=%v) snid:%v find in id:%v csp.PlayerEnter return false", gps.GroupId, p.SnId, id)
			return gamehall_proto.OpResultCode_Hundred_OPRC_Error_Hundred
		}
	}
	//没有场景，尝试创建
	if _, ok := this.scenesOfPlatform[platformName]; !ok {
		this.scenesOfPlatform[platformName] = make(map[int32]*Scene)
	}
	if ss, ok := this.scenesOfPlatform[platformName]; ok {
		if s, ok := ss[id]; !ok {
			s = this.CreateNewScene(id, 0, limitPlatform, gps.DbGameFree)
			if s != nil {
				ss[id] = s
				this.platformOfScene[int32(s.sceneId)] = platformName
				logger.Logger.Tracef("(this *HundredSceneMgr) PlayerEnter(platform=%v) Create %v scene success.", platformName, id)
			} else {
				logger.Logger.Tracef("(this *HundredSceneMgr) PlayerEnter(platform=%v) Create %v scene failed.", platformName, id)
			}
		}
		//尝试进入
		if s, ok := ss[id]; ok && s != nil {
			if s.PlayerEnter(p, -1, true) {
				this.OnPlayerEnter(p, id)
				return gamehall_proto.OpResultCode_Hundred_OPRC_Sucess_Hundred
			} else {
				logger.Logger.Warnf("(this *HundredSceneMgr) PlayerEnter(platform=%v) enter %v scene failed.", platformName, id)
			}
		} else {
			logger.Logger.Warnf("(this *HundredSceneMgr) PlayerEnter(platform=%v) get %v scene failed.", platformName, id)
		}
	}
	logger.Logger.Warnf("(this *HundredSceneMgr) PlayerEnter(platform=%v) snid:%v find in id:%v csp.PlayerEnter return false", platformName, p.SnId, id)
	return gamehall_proto.OpResultCode_Hundred_OPRC_SceneServerMaintain_Hundred
}

func (this *HundredSceneMgr) OnPlayerEnter(p *Player, id int32) {
	this.playerIning[p.SnId] = id
}

func (this *HundredSceneMgr) PlayerLeave(p *Player, reason int) bool {
	if p == nil {
		return false
	}
	if _, ok := this.playerIning[p.SnId]; ok {
		if p.scene != nil {
			p.scene.PlayerLeave(p, reason)
		} else {
			logger.Logger.Warnf("(this *HundredSceneMgr) PlayerLeave(%v) found scene=nil", p.SnId)
			delete(this.playerIning, p.SnId)
		}
		return true
	} else {
		if p.scene != nil && p.scene.IsHundredScene() {
			logger.Logger.Warnf("(this *HundredSceneMgr) PlayerLeave(%v) exception scene=%v gameid=%v", p.SnId, p.scene.sceneId, p.scene.gameId)
			p.scene.PlayerLeave(p, reason)
			return true
		}
	}
	logger.Logger.Warnf("(this *HundredSceneMgr) PlayerLeave(%v) not found in hundred scene", p.SnId)
	return false
}

func (this *HundredSceneMgr) PlayerTryLeave(p *Player) gamehall_proto.OpResultCode_Hundred {
	if p.scene == nil || p.scene.gameSess == nil {
		logger.Logger.Tracef("(csm *HundredSceneMgr) PlayerTryLeave p.scene == nil || p.scene.gameSess == nil snid:%v ", p.SnId)
		return 1
	}
	//通知gamesrv托管
	if _, ok := this.playerIning[p.SnId]; ok {
		pack := &gamehall_proto.CSLeaveRoom{Mode: proto.Int(0)}
		proto.SetDefaults(pack)
		common.TransmitToServer(p.sid, int(gamehall_proto.GameHallPacketID_PACKET_CS_LEAVEROOM), pack, p.scene.gameSess.Session)
	}
	return 0
}

func (this *HundredSceneMgr) OnPlayerLeave(p *Player) {
	delete(this.playerIning, p.SnId)
}

func (this *HundredSceneMgr) OnDestroyScene(sceneid int) {
	if platformName, ok := this.platformOfScene[int32(sceneid)]; ok {
		if ss, ok := this.scenesOfPlatform[platformName]; ok {
			for id, scene := range ss {
				if scene.sceneId == sceneid {
					//删除玩家
					for pid, hid := range this.playerIning {
						if hid == id {
							delete(this.playerIning, pid)
							//TODO 非正常删除房间时，尝试同步金币
							player := PlayerMgrSington.GetPlayerBySnId(pid)
							if player != nil {
								if !player.IsRob {
									ctx := scene.GetPlayerGameCtx(player.SnId)
									if ctx != nil {
										//发送一个探针,等待ack后同步金币
										player.TryRetrieveLostGameCoin(sceneid)

										logger.Logger.Warnf("(this *HundredSceneMgr) OnDestroyScene(sceneid:%v) snid:%v SyncGameCoin", sceneid, player.SnId)
									}
								}
							}
						}
					}
					delete(ss, id)
					break
				}
			}
		}
	}

	if groupId, ok := this.groupOfScene[int32(sceneid)]; ok {
		if ss, ok := this.scenesOfGroup[groupId]; ok {
			for id, scene := range ss {
				if scene.sceneId == sceneid {
					//删除玩家
					for pid, hid := range this.playerIning {
						if hid == id {
							delete(this.playerIning, pid)
							//TODO 非正常删除房间时，尝试同步金币
							player := PlayerMgrSington.GetPlayerBySnId(pid)
							if player != nil {
								if !player.IsRob {
									ctx := scene.GetPlayerGameCtx(player.SnId)
									if ctx != nil {
										//发送一个探针,等待ack后同步金币
										player.TryRetrieveLostGameCoin(sceneid)
										logger.Logger.Warnf("(this *HundredSceneMgr) OnDestroyScene(sceneid:%v) snid:%v SyncGameCoin", sceneid, player.SnId)
									}
								}
							}
						}
					}
					delete(ss, id)
					break
				}
			}
		}
	}
}

func (this *HundredSceneMgr) GetPlayerNums(p *Player, gameId, gameMode int32) []int32 {
	//多平台支持
	platformName := Default_Platform
	platform := PlatformMgrSington.GetPlatform(p.Platform)
	if platform != nil && platform.Isolated {
		platformName = platform.IdStr
	} else if p.Platform != Default_Platform {
		platform = PlatformMgrSington.GetPlatform(Default_Platform)
	}

	var nums [HundredSceneType_Max]int32
	wantNum := []int32{80, 50, 30, 20, 0}
	for i := 0; i < HundredSceneType_Max; i++ {
		if wantNum[i]/2 > 0 {
			nums[i] = rand.Int31n(wantNum[i]/2) + wantNum[i]
		}
	}

	if platform == nil {
		return nums[:]
	}

	ids, _ := srvdata.DataMgr.GetGameFreeIds(gameId, gameMode)
	for _, id := range ids {
		gps := PlatformMgrSington.GetGameConfig(platform.IdStr, id)
		if gps != nil {
			if gps.GroupId != 0 {
				if ss, exist := this.scenesOfGroup[gps.GroupId]; exist {
					for _, s := range ss {
						if s.paramsEx[0] == id {
							dbGame := srvdata.PBDB_GameFreeMgr.GetData(s.paramsEx[0])
							sceneType := int(dbGame.GetSceneType()) - 1
							if sceneType == -2 {
								//体验场
								sceneType = HundredSceneType_Experience
							}
							truePlayerCount := int32(s.GetPlayerCnt())

							//获取fake用户数量
							var fakePlayerCount int32
							//if truePlayerCount >= 21 {
							//	correctNum := dbGame.GetCorrectNum()
							//	correctRate := dbGame.GetCorrectRate()
							//	fakePlayerCount = correctNum + truePlayerCount*correctRate/100 + dbGame.GetDeviation()
							//}
							if sceneType >= 0 && sceneType < HundredSceneType_Max {
								nums[sceneType] += int32(truePlayerCount + fakePlayerCount)
							}
							break
						}
					}
				}
			} else {
				if ss, ok := this.scenesOfPlatform[platformName]; ok {
					for _, s := range ss {
						if s.paramsEx[0] == id {
							dbGame := srvdata.PBDB_GameFreeMgr.GetData(s.paramsEx[0])
							sceneType := int(dbGame.GetSceneType()) - 1
							if sceneType == -2 {
								//体验场
								sceneType = HundredSceneType_Experience
							}
							truePlayerCount := int32(s.GetPlayerCnt())

							//获取fake用户数量
							var fakePlayerCount int32
							//if truePlayerCount >= 21 {
							//	correctNum := dbGame.GetCorrectNum()
							//	correctRate := dbGame.GetCorrectRate()
							//	fakePlayerCount = correctNum + truePlayerCount*correctRate/100 + dbGame.GetDeviation()
							//}
							if sceneType >= 0 && sceneType < HundredSceneType_Max {
								nums[sceneType] += int32(truePlayerCount + fakePlayerCount)
							}
							break
						}
					}
				}
			}
		}
	}
	return nums[:]
}

func (this *HundredSceneMgr) InHundredScene(p *Player) bool {
	if p == nil {
		logger.Logger.Tracef("(this *HundredSceneMgr) InHundredScene p == nil snid:%v ", p.SnId)
		return false
	}
	if _, ok := this.playerIning[p.SnId]; ok {
		return true
	}
	logger.Logger.Tracef("(csm *HundredSceneMgr) InHundredScene false snid:%v ", p.SnId)
	return false
}

func (this *HundredSceneMgr) CreateNewScene(id, groupId int32, limitPlatform *Platform, dbGameFree *server_proto.DB_GameFree) *Scene {
	if dbGameFree != nil {
		dbGameRule := srvdata.PBDB_GameRuleMgr.GetData(dbGameFree.GetGameRule())
		if dbGameRule != nil {
			gameId := int(dbGameRule.GetGameId())
			gs := GameSessMgrSington.GetMinLoadSess(gameId)
			if gs != nil {
				sceneId := SceneMgrSington.GenOneHundredSceneId()
				gameMode := dbGameRule.GetGameMode()
				params := dbGameRule.GetParams()
				SceneType := dbGameFree.GetSceneType()

				scene := SceneMgrSington.CreateScene(0, 0, sceneId, gameId, int(gameMode), int(SceneType), 1, -1, params, gs, limitPlatform, groupId, dbGameFree, id)
				if scene != nil {
					scene.hallId = id
					//移动到SceneMgr中集中处理
					//if !scene.IsMatchScene() {
					//	//平台水池设置
					//	gs.DetectCoinPoolSetting(limitPlatform.Name, scene.hallId, scene.groupId)
					//}
					return scene
				} else {
					logger.Logger.Errorf("Create hundred scene %v-%v failed.", gameId, sceneId)
				}
			} else {
				logger.Logger.Errorf("Game %v server session no found.", gameId)
			}
		} else {
			logger.Logger.Errorf("Game rule data %v no found.", dbGameFree.GetGameRule())
		}
	} else {
		logger.Logger.Errorf("Game free data %v no found.", id)
	}

	return nil
}

func (this *HundredSceneMgr) TryCreateRoom() {
	if model.GameParamData.HundredScenePreCreate {
		arr := srvdata.PBDB_GameFreeMgr.Datas.GetArr()
		for _, dbGame := range arr {
			if dbGame.GetGameId() <= 0 {
				continue
			}
			if common.IsHundredType(dbGame.GetGameType()) { //百人场
				id := dbGame.GetId()
				for k, ss := range this.scenesOfPlatform {
					if _, exist := ss[id]; !exist {
						limitPlatform := PlatformMgrSington.GetPlatform(k)
						if limitPlatform == nil || !limitPlatform.Isolated {
							limitPlatform = PlatformMgrSington.GetPlatform(Default_Platform)
							k = Default_Platform
						}

						gps := PlatformMgrSington.GetGameConfig(limitPlatform.IdStr, id)
						if gps != nil && gps.GroupId == 0 {
							scene := this.CreateNewScene(id, gps.GroupId, limitPlatform, gps.DbGameFree)
							logger.Logger.Trace("(this *HundredSceneMgr) TryCreateRoom(platform) ", id, k, scene)
							if scene != nil {
								this.platformOfScene[int32(scene.sceneId)] = k
								ss[id] = scene
							}
						}
					}
				}
			}
		}
	}
}
func (this *HundredSceneMgr) PreCreateGame(platform string, createIds []int32) {
	limitPlatform := PlatformMgrSington.GetPlatform(platform)
	if limitPlatform == nil || !limitPlatform.Isolated {
		limitPlatform = PlatformMgrSington.GetPlatform(Default_Platform)
	}
	if this.scenesOfPlatform[platform] == nil {
		this.scenesOfPlatform[platform] = make(map[int32]*Scene)
	}
	//var platformName string
	platformData := PlatformMgrSington.GetPlatform(platform)
	if platformData != nil && platformData.Isolated {
		//platformName = platformData.Name
	} else if platform != Default_Platform {
		platformData = PlatformMgrSington.GetPlatform(Default_Platform)
	}

	//不创建已经存在的场景
	for _, id := range createIds {
		dbGame := srvdata.PBDB_GameFreeMgr.GetData(id)
		gps := PlatformMgrSington.GetGameConfig(platformData.IdStr, id)
		if gps != nil {
			if gps.GroupId != 0 {
				if this.scenesOfGroup[gps.GroupId] != nil && this.scenesOfGroup[gps.GroupId][id] != nil {
					continue
				} else {
					scene := this.CreateNewScene(dbGame.GetId(), gps.GroupId, limitPlatform, gps.DbGameFree)
					if scene != nil {
						this.scenesOfGroup[gps.GroupId][id] = scene
						this.groupOfScene[int32(scene.sceneId)] = gps.GroupId
					}
				}

			} else {
				if this.scenesOfPlatform[platform] != nil && this.scenesOfPlatform[platform][dbGame.GetId()] != nil {
					continue
				} else {
					scene := this.CreateNewScene(dbGame.GetId(), gps.GroupId, limitPlatform, gps.DbGameFree)
					if scene != nil {
						this.platformOfScene[int32(scene.sceneId)] = platform
						this.scenesOfPlatform[platform][dbGame.GetId()] = scene
					}
				}
			}
		}
	}
}
func (this *HundredSceneMgr) OnPlatformCreate(p *Platform) {
	if p != nil && p.Isolated {
		if _, exist := this.scenesOfPlatform[p.IdStr]; !exist {
			this.scenesOfPlatform[p.IdStr] = make(map[int32]*Scene)
			if model.GameParamData.HundredScenePreCreate {
				arr := srvdata.PBDB_GameFreeMgr.Datas.GetArr()
				for _, dbGame := range arr {
					if common.IsHundredType(dbGame.GetGameType()) { //百人场
						id := dbGame.GetId()
						gps := PlatformMgrSington.GetGameConfig(p.IdStr, id)
						if gps != nil {
							if gps.GroupId != 0 {
								if ss, ok := this.scenesOfGroup[gps.GroupId]; ok {
									if _, exist := ss[id]; !exist {
										pgg := PlatformGameGroupMgrSington.GetGameGroup(gps.GroupId)
										if pgg != nil {
											scene := this.CreateNewScene(id, gps.GroupId, p, pgg.DbGameFree)
											logger.Logger.Trace("(this *HundredSceneMgr) TryCreateRoom(group) ", id, gps.GroupId, scene)
											if scene != nil {
												ss[id] = scene
											}
										}
									}
								}
							} else {
								if ss, ok := this.scenesOfPlatform[p.IdStr]; ok {
									if _, exist := ss[id]; !exist {
										scene := this.CreateNewScene(id, 0, p, gps.DbGameFree)
										logger.Logger.Trace("(this *HundredSceneMgr) TryCreateRoom(platform) ", id, p.Name, scene)
										if scene != nil {
											ss[id] = scene
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
}

func (this *HundredSceneMgr) OnPlatformDestroy(p *Platform) {
	if p == nil {
		return
	}
	if ss, ok := this.scenesOfPlatform[p.IdStr]; ok {
		pack := &server_proto.WGGraceDestroyScene{}
		for _, scene := range ss {
			pack.Ids = append(pack.Ids, int32(scene.sceneId))
		}
		srvlib.ServerSessionMgrSington.Broadcast(int(server_proto.SSPacketID_PACKET_WG_GRACE_DESTROYSCENE), pack, common.GetSelfAreaId(), srvlib.GameServerType)
	}
}

func (this *HundredSceneMgr) OnPlatformChangeIsolated(p *Platform, isolated bool) {
	if p != nil {
		if isolated { //孤立
			this.OnPlatformCreate(p) //预创建场景
		} else {
			if ss, ok := this.scenesOfPlatform[p.IdStr]; ok {
				pack := &server_proto.WGGraceDestroyScene{}
				for _, scene := range ss {
					pack.Ids = append(pack.Ids, int32(scene.sceneId))
				}
				srvlib.ServerSessionMgrSington.Broadcast(int(server_proto.SSPacketID_PACKET_WG_GRACE_DESTROYSCENE), pack, common.GetSelfAreaId(), srvlib.GameServerType)
			}
		}
	}
}

func (this *HundredSceneMgr) OnPlatformChangeDisabled(p *Platform, disabled bool) {
	if p == nil {
		return
	}
	if disabled {
		if ss, ok := this.scenesOfPlatform[p.IdStr]; ok {
			pack := &server_proto.WGGraceDestroyScene{}
			for _, scene := range ss {
				pack.Ids = append(pack.Ids, int32(scene.sceneId))
			}
			srvlib.ServerSessionMgrSington.Broadcast(int(server_proto.SSPacketID_PACKET_WG_GRACE_DESTROYSCENE), pack, common.GetSelfAreaId(), srvlib.GameServerType)
		}
	}
}

func (this *HundredSceneMgr) OnPlatformConfigUpdate(p *Platform, oldCfg, newCfg *webapi.GameFree) {
	if p == nil || newCfg == nil {
		return
	}
	if scenes, exist := this.scenesOfPlatform[p.IdStr]; exist {
		pack := &server_proto.WGGraceDestroyScene{}
		if s, ok := scenes[newCfg.DbGameFree.Id]; ok {
			pack.Ids = append(pack.Ids, int32(s.sceneId))
		}
		srvlib.ServerSessionMgrSington.Broadcast(int(server_proto.SSPacketID_PACKET_WG_GRACE_DESTROYSCENE),
			pack, common.GetSelfAreaId(), srvlib.GameServerType)
	}
}

func (this *HundredSceneMgr) OnGameGroupUpdate(oldCfg, newCfg *webapi.GameConfigGroup) {
	if newCfg == nil {
		return
	}
	if scenes, exist := this.scenesOfGroup[newCfg.Id]; exist {
		if s, ok := scenes[newCfg.DbGameFree.Id]; ok {
			needDestroy := false
			if s.dbGameFree.GetBot() != newCfg.DbGameFree.GetBot() ||
				s.dbGameFree.GetBaseScore() != newCfg.DbGameFree.GetBaseScore() ||
				s.dbGameFree.GetLimitCoin() != newCfg.DbGameFree.GetLimitCoin() ||
				s.dbGameFree.GetMaxCoinLimit() != newCfg.DbGameFree.GetMaxCoinLimit() ||
				!common.SliceInt32Equal(s.dbGameFree.GetRobotTakeCoin(), newCfg.DbGameFree.GetRobotTakeCoin()) ||
				!common.SliceInt32Equal(s.dbGameFree.GetRobotLimitCoin(), newCfg.DbGameFree.GetRobotLimitCoin()) {
				needDestroy = true
			}
			if needDestroy {
				pack := &server_proto.WGGraceDestroyScene{}
				pack.Ids = append(pack.Ids, int32(s.sceneId))
				srvlib.ServerSessionMgrSington.Broadcast(int(server_proto.SSPacketID_PACKET_WG_GRACE_DESTROYSCENE), pack, common.GetSelfAreaId(), srvlib.GameServerType)
			}
		}
	}
}
func (this *HundredSceneMgr) GetPlatformSceneByGameFreeId(platform string, gameFreeIds []int32) []*Scene {
	platformName := Default_Platform
	platformData := PlatformMgrSington.GetPlatform(platform)
	if platformData != nil && platformData.Isolated {
		platformName = platformData.IdStr
	} else if platform != Default_Platform {
		platformData = PlatformMgrSington.GetPlatform(Default_Platform)
	}
	gameScenes := []*Scene{}
	for _, id := range gameFreeIds {
		gps := PlatformMgrSington.GetGameConfig(platformData.IdStr, id)
		if gps != nil {
			if gps.GroupId != 0 {
				if ss, exist := this.scenesOfGroup[gps.GroupId]; exist {
					if s, exist := ss[id]; exist && s != nil {
						gameScenes = append(gameScenes, s)
					}
				}
			} else {
				if ss, ok := this.scenesOfPlatform[platformName]; ok {
					if s, exist := ss[id]; exist && s != nil {
						gameScenes = append(gameScenes, s)
					}
				}
			}
		}
	}

	return gameScenes
}
func (this *HundredSceneMgr) GetPlatformScene(platform string, gameid int32) []*Scene {
	gameFreeIds := gameStateMgr.gameIds[gameid]
	gameScenes := this.GetPlatformSceneByGameFreeId(platform, gameFreeIds)
	if len(gameScenes) != len(gameFreeIds) {
		createIds := []int32{}
		for _, gfi := range gameFreeIds {
			bFind := false
			for _, s := range gameScenes {
				if s.dbGameFree.GetId() == gfi {
					bFind = false
					break
				}
			}
			if !bFind {
				createIds = append(createIds, gfi)
			}
		}
		if len(createIds) > 0 {
			this.PreCreateGame(platform, createIds)
			gameScenes = this.GetPlatformSceneByGameFreeId(platform, gameFreeIds)
		}
	}
	return gameScenes
}
func (this *HundredSceneMgr) ModuleName() string {
	return "HundredSceneMgr"
}

func (this *HundredSceneMgr) Init() {
	for platformName, platform := range PlatformMgrSington.Platforms {
		if platform.Isolated || platformName == Default_Platform {
			this.scenesOfPlatform[platformName] = make(map[int32]*Scene)
		}
	}
}

// 撮合
func (this *HundredSceneMgr) Update() {

}

func (this *HundredSceneMgr) Shutdown() {
	module.UnregisteModule(this)
}

func (this *HundredSceneMgr) OnPlatformDestroyByGameFreeId(p *Platform, gameFreeId int32) {
	if p == nil {
		return
	}
	if scenes, ok := this.scenesOfPlatform[p.IdStr]; ok {
		for _, scene := range scenes {
			pack := &server_proto.WGGraceDestroyScene{}
			if scene.dbGameFree.Id == gameFreeId {
				pack.Ids = append(pack.Ids, int32(scene.sceneId))
			}
			srvlib.ServerSessionMgrSington.Broadcast(int(server_proto.SSPacketID_PACKET_WG_GRACE_DESTROYSCENE), pack, common.GetSelfAreaId(), srvlib.GameServerType)
		}
	}
}
func init() {
	module.RegisteModule(HundredSceneMgrSington, time.Second*5, 0)
	PlatformMgrSington.RegisteObserver(HundredSceneMgrSington)
	PlatformGameGroupMgrSington.RegisteObserver(HundredSceneMgrSington)
}
