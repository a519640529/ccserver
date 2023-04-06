package main

import (
	server_proto "games.yol.com/win88/protocol/server"
	"games.yol.com/win88/protocol/webapi"
	"math/rand"
	"time"

	"games.yol.com/win88/common"
	"games.yol.com/win88/proto"
	hall_proto "games.yol.com/win88/protocol/gamehall"
	"games.yol.com/win88/srvdata"
	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/module"
	"github.com/idealeak/goserver/core/transact"
	"github.com/idealeak/goserver/srvlib"
)

const (
	CoinSceneType_Primary    int = iota //初级
	CoinSceneType_Mid                   //中级
	CoinSceneType_Senior                //高级
	CoinSceneType_Experience            //体验场
	CoinSceneType_Professor             //专家
	CoinSceneType_Max
)

type CreateRoomCache struct {
	//cfgId        int32
	gameFreeId   int32
	platformName string
}

var CoinSceneMgrSington = &CoinSceneMgr{
	//pool:           make(map[int32]*CoinScenePool),
	playerIning:    make(map[int32]int32),
	playerChanging: make(map[int32]int32),
	//按平台管理
	scenesOfPlatform: make(map[string]map[int32]*CoinScenePool),
	platformOfScene:  make(map[int32]string),
	//按组管理
	scenesOfGroup: make(map[int32]map[int32]*CoinScenePool),
	groupOfScene:  make(map[int32]int32),
	//
	sceneOfcsp: make(map[int]*CoinScenePool),
}

type CoinSceneMgr struct {
	//pool           map[int32]*CoinScenePool
	playerIning    map[int32]int32
	playerChanging map[int32]int32
	//按平台管理
	scenesOfPlatform map[string]map[int32]*CoinScenePool
	platformOfScene  map[int32]string
	//按组管理
	scenesOfGroup map[int32]map[int32]*CoinScenePool
	groupOfScene  map[int32]int32
	//延迟创建房间列表
	delayCache []CreateRoomCache
	//scene到csp的映射
	sceneOfcsp map[int]*CoinScenePool
	//
	CreateRoomTick int64
}

func (csm *CoinSceneMgr) GetCoinSceneMgr(p *Player, dbGameFree *server_proto.DB_GameFree) *CoinScenePool {
	if ss, ok := csm.scenesOfPlatform[p.Platform]; ok {
		if csp, ok := ss[dbGameFree.GetId()]; ok && csp != nil {
			return csp
		}
	}

	csp := &CoinScenePool{
		id:            dbGameFree.GetId(),
		dbGameFree:    dbGameFree,
		scenes:        make(map[int]*Scene),
		players:       make(map[int32]int),
		queue:         make(map[int32]*Player),
		robRequireNum: 1,
	}

	return csp
}

func (csm *CoinSceneMgr) RebindPlayerSnId(oldSnId, newSnId int32) {
	if id, exist := csm.playerIning[oldSnId]; exist {
		delete(csm.playerIning, oldSnId)
		csm.playerIning[newSnId] = id
	}
	for _, ss := range csm.scenesOfPlatform {
		for _, pool := range ss {
			pool.RebindPlayerSnId(oldSnId, newSnId)
		}
	}
	for _, ss := range csm.scenesOfGroup {
		for _, pool := range ss {
			pool.RebindPlayerSnId(oldSnId, newSnId)
		}
	}
}

func (csm *CoinSceneMgr) PlayerEnter(p *Player, id int32, roomId int32, exclude []int32, ischangeroom bool) hall_proto.OpResultCode {
	logger.Logger.Tracef("(csm *CoinSceneMgr) PlayerEnter snid:%v id:%v roomid:%v exclude:%v", p.SnId, id, roomId, exclude)

	if p.isDelete { //删档用户不让进游戏
		return hall_proto.OpResultCode_OPRC_RoomHadClosed
	}
	//多平台支持
	platformName := Default_Platform
	platform := PlatformMgrSington.GetPlatform(p.Platform)
	if platform != nil && platform.Isolated {
		platformName = platform.IdStr
	} else {
		platform = PlatformMgrSington.GetPlatform(Default_Platform)
	}
	if platform == nil {
		return hall_proto.OpResultCode_OPRC_RoomHadClosed
	}

	gps := PlatformMgrSington.GetGameConfig(platform.IdStr, id)
	if gps == nil {
		return hall_proto.OpResultCode_OPRC_RoomHadClosed
	}

	if gps.GroupId != 0 { //按分组进入场景游戏
		pgg := PlatformGameGroupMgrSington.GetGameGroup(gps.GroupId)
		if pgg != nil {
			if _, ok := csm.scenesOfGroup[gps.GroupId]; !ok {
				csm.scenesOfGroup[gps.GroupId] = make(map[int32]*CoinScenePool)
			}
			if ss, ok := csm.scenesOfGroup[gps.GroupId]; ok {
				if csp, ok := ss[id]; ok && csp != nil {
					ret := csp.PlayerEnter(p, roomId, exclude, ischangeroom)
					if ret == hall_proto.OpResultCode_OPRC_Sucess {
						csm.OnPlayerEnter(p, id)
						return hall_proto.OpResultCode_OPRC_Sucess
					}
					if ret == hall_proto.OpResultCode_OPRC_CoinSceneEnterQueueSucc {
						return ret
					}
					logger.Logger.Warnf("(csm *CoinSceneMgr) PlayerEnter snid:%v find in id:%v exclude:%v return false", p.SnId, id, exclude)
					return ret
				}

				csp := NewCoinScenePool(id, gps.GroupId, pgg.DbGameFree)
				if csp == nil {
					logger.Logger.Warnf("(csm *CoinSceneMgr) PlayerEnter snid:%v find in id:%v exclude:%v NewCoinScenePool failed", p.SnId, id, exclude)
					return hall_proto.OpResultCode_OPRC_Error
				}
				ss[id] = csp
				if in, ok := csm.playerIning[p.SnId]; ok {
					if in != id {
						logger.Logger.Warnf("(csm *CoinSceneMgr) PlayerEnter snid:%v find in old:%v new:%v", p.SnId, in, id)
						return hall_proto.OpResultCode_OPRC_Error
					}
				}
				ret := csp.PlayerEnter(p, roomId, exclude, ischangeroom)
				if ret == hall_proto.OpResultCode_OPRC_Sucess {
					csm.OnPlayerEnter(p, id)
					return hall_proto.OpResultCode_OPRC_Sucess
				}
				if ret == hall_proto.OpResultCode_OPRC_CoinSceneEnterQueueSucc {
					return ret
				}
				logger.Logger.Warnf("(csm *CoinSceneMgr) PlayerEnter snid:%v find in id:%v exclude:%v csp.PlayerEnter return false", p.SnId, id, exclude)
			}
			return hall_proto.OpResultCode_OPRC_Error
		}
	}
	//没有场景，尝试创建
	if _, ok := csm.scenesOfPlatform[platformName]; !ok {
		csm.scenesOfPlatform[platformName] = make(map[int32]*CoinScenePool)
	}
	if ss, ok := csm.scenesOfPlatform[platformName]; ok {
		if csp, ok := ss[id]; ok && csp != nil {
			ret := csp.PlayerEnter(p, roomId, exclude, ischangeroom)
			if ret == hall_proto.OpResultCode_OPRC_Sucess {
				csm.OnPlayerEnter(p, id)
				return hall_proto.OpResultCode_OPRC_Sucess
			}
			if ret == hall_proto.OpResultCode_OPRC_CoinSceneEnterQueueSucc {
				return ret
			}
			logger.Logger.Warnf("(csm *CoinSceneMgr) PlayerEnter snid:%v find in id:%v exclude:%v return false", p.SnId, id, exclude)
			return ret
		}

		csp := NewCoinScenePool(id, 0, gps.DbGameFree)
		if csp == nil {
			logger.Logger.Warnf("(csm *CoinSceneMgr) PlayerEnter snid:%v find in id:%v exclude:%v NewCoinScenePool failed", p.SnId, id, exclude)
			return hall_proto.OpResultCode_OPRC_Error
		}
		ss[id] = csp
		if in, ok := csm.playerIning[p.SnId]; ok {
			if in != id {
				logger.Logger.Warnf("(csm *CoinSceneMgr) PlayerEnter snid:%v find in old:%v new:%v", p.SnId, in, id)
				return hall_proto.OpResultCode_OPRC_Error
			}
		}
		ret := csp.PlayerEnter(p, roomId, exclude, ischangeroom)
		if ret == hall_proto.OpResultCode_OPRC_Sucess {
			csm.OnPlayerEnter(p, id)
			return hall_proto.OpResultCode_OPRC_Sucess
		}
		if ret == hall_proto.OpResultCode_OPRC_CoinSceneEnterQueueSucc {
			return ret
		}
		logger.Logger.Warnf("(csm *CoinSceneMgr) PlayerEnter snid:%v find in id:%v exclude:%v csp.PlayerEnter return false", p.SnId, id, exclude)
	}
	return hall_proto.OpResultCode_OPRC_Error
}

func (csm *CoinSceneMgr) PlayerEnterLocalGame(p *Player, id int32, roomId int32, exclude []int32, ischangeroom bool) hall_proto.OpResultCode {
	logger.Logger.Tracef("(csm *CoinSceneMgr) PlayerEnterLocalGame snid:%v id:%v roomid:%v exclude:%v", p.SnId, id, roomId, exclude)

	if p.isDelete { //删档用户不让进游戏
		return hall_proto.OpResultCode_OPRC_RoomHadClosed
	}
	//多平台支持
	platformName := Default_Platform
	platform := PlatformMgrSington.GetPlatform(p.Platform)
	if platform != nil && platform.Isolated {
		platformName = platform.IdStr
	} else {
		platform = PlatformMgrSington.GetPlatform(Default_Platform)
	}
	if platform == nil {
		return hall_proto.OpResultCode_OPRC_RoomHadClosed
	}

	gps := PlatformMgrSington.GetGameConfig(platform.IdStr, id)
	if gps == nil {
		return hall_proto.OpResultCode_OPRC_RoomHadClosed
	}

	//没有场景，尝试创建
	if _, ok := csm.scenesOfPlatform[platformName]; !ok {
		csm.scenesOfPlatform[platformName] = make(map[int32]*CoinScenePool)
	}
	if ss, ok := csm.scenesOfPlatform[platformName]; ok {
		if csp, ok := ss[id]; ok && csp != nil {
			ret := csp.PlayerEnterLocalGame(p, roomId, exclude, ischangeroom)
			if ret == hall_proto.OpResultCode_OPRC_Sucess {
				csm.OnPlayerEnter(p, id)
				return hall_proto.OpResultCode_OPRC_Sucess
			}
			if ret == hall_proto.OpResultCode_OPRC_CoinSceneEnterQueueSucc {
				return ret
			}
			logger.Logger.Warnf("(csm *CoinSceneMgr) PlayerEnter snid:%v find in id:%v exclude:%v return false", p.SnId, id, exclude)
			return ret
		}

		csp := NewCoinScenePool(id, 0, gps.DbGameFree)
		if csp == nil {
			logger.Logger.Warnf("(csm *CoinSceneMgr) PlayerEnter snid:%v find in id:%v exclude:%v NewCoinScenePool failed", p.SnId, id, exclude)
			return hall_proto.OpResultCode_OPRC_Error
		}
		ss[id] = csp
		if in, ok := csm.playerIning[p.SnId]; ok {
			if in != id {
				logger.Logger.Warnf("(csm *CoinSceneMgr) PlayerEnter snid:%v find in old:%v new:%v", p.SnId, in, id)
				return hall_proto.OpResultCode_OPRC_Error
			}
		}
		ret := csp.PlayerEnterLocalGame(p, roomId, exclude, ischangeroom)
		if ret == hall_proto.OpResultCode_OPRC_Sucess {
			csm.OnPlayerEnter(p, id)
			return hall_proto.OpResultCode_OPRC_Sucess
		}
		if ret == hall_proto.OpResultCode_OPRC_CoinSceneEnterQueueSucc {
			return ret
		}
		logger.Logger.Warnf("(csm *CoinSceneMgr) PlayerEnter snid:%v find in id:%v exclude:%v csp.PlayerEnter return false", p.SnId, id, exclude)
	}
	return hall_proto.OpResultCode_OPRC_Error
}

func (csm *CoinSceneMgr) OnPlayerEnter(p *Player, id int32) {
	csm.playerIning[p.SnId] = id
}

func (csm *CoinSceneMgr) PlayerLeave(p *Player, reason int) bool {
	if p == nil {
		return false
	}

	//多平台支持
	platformName := Default_Platform
	platform := PlatformMgrSington.GetPlatform(p.Platform)
	if platform != nil && platform.Isolated {
		platformName = platform.IdStr
	}

	if csm.PlayerLeaveByPlatformName(platformName, p, reason) {
		return true
	}

	if p.scene != nil && p.scene.limitPlatform != nil {
		oldName := platformName
		platformName = p.scene.limitPlatform.IdStr
		logger.Logger.Warnf("(this *CoinSceneMgr) PlayerLeave(%v) exception scene=%v gameid=%v before platform=%v try use platform=%v", p.SnId, p.scene.sceneId, p.scene.gameId, oldName, platformName)
		if csm.PlayerLeaveByPlatformName(platformName, p, reason) {
			return true
		}
	}

	return false
}

func (csm *CoinSceneMgr) PlayerLeaveByPlatformName(platformName string, p *Player, reason int) bool {
	if id, ok := csm.playerIning[p.SnId]; ok {
		//没有场景，尝试创建
		if p.scene != nil {
			s := p.scene
			if s.groupId != 0 {
				if ss, ok := csm.scenesOfGroup[s.groupId]; ok {
					if csp, ok := ss[id]; ok && csp != nil {
						if csp.PlayerLeave(p, reason) {
							csm.OnPlayerLeave(s, p)
							return true
						}
					}
				}
			} else {
				if ss, ok := csm.scenesOfPlatform[platformName]; ok {
					if csp, ok := ss[id]; ok && csp != nil {
						if csp.PlayerLeave(p, reason) {
							csm.OnPlayerLeave(s, p)
							return true
						}
					}
				}
			}
		}
	}

	return false
}

func (csm *CoinSceneMgr) OnPlayerLeave(s *Scene, p *Player) {
	delete(csm.playerIning, p.SnId)
}

func (csm *CoinSceneMgr) AudienceEnter(p *Player, id int32, roomId int32, exclude []int32, ischangeroom bool) hall_proto.OpResultCode {
	//多平台支持
	platformName := Default_Platform
	platform := PlatformMgrSington.GetPlatform(p.Platform)
	if platform != nil && platform.Isolated {
		platformName = platform.IdStr
	}

	gps := PlatformMgrSington.GetGameConfig(platform.IdStr, id)
	if gps == nil {
		return hall_proto.OpResultCode_OPRC_RoomHadClosed
	}

	if gps.GroupId != 0 { //按分组进入场景游戏
		pgg := PlatformGameGroupMgrSington.GetGameGroup(gps.GroupId)
		if pgg != nil {
			if _, ok := csm.scenesOfGroup[gps.GroupId]; !ok {
				return hall_proto.OpResultCode_OPRC_NoFindDownTiceRoom
			}
			if ss, ok := csm.scenesOfGroup[gps.GroupId]; ok {
				if csp, ok := ss[id]; ok && csp != nil {
					ret := csp.AudienceEnter(p, roomId, exclude, ischangeroom)
					if ret == hall_proto.OpResultCode_OPRC_Sucess {
						csm.OnPlayerEnter(p, id)
						return hall_proto.OpResultCode_OPRC_Sucess
					}
					logger.Logger.Warnf("(csm *CoinSceneMgr) AudienceEnter snid:%v find in id:%v exclude:%v return false", p.SnId, id, exclude)
					return ret
				}

				logger.Logger.Warnf("(csm *CoinSceneMgr) AudienceEnter snid:%v find in id:%v exclude:%v csp.PlayerEnter return false", p.SnId, id, exclude)
			}
			return hall_proto.OpResultCode_OPRC_Error
		}
	}

	//没有场景，尝试创建
	if pool, exist := csm.scenesOfPlatform[platformName]; !exist {
		return hall_proto.OpResultCode_OPRC_NoFindDownTiceRoom
	} else {
		if csp, ok := pool[id]; ok && csp != nil {
			ret := csp.AudienceEnter(p, roomId, exclude, ischangeroom)
			if ret == hall_proto.OpResultCode_OPRC_Sucess {
				csm.OnPlayerEnter(p, id)
				return hall_proto.OpResultCode_OPRC_Sucess
			}
			logger.Logger.Warnf("(csm *CoinSceneMgr) AudienceEnter snid:%v find in id:%v exclude:%v return false", p.SnId, id, exclude)
			return ret
		} else {

		}
	}

	return hall_proto.OpResultCode_OPRC_NoFindDownTiceRoom
}

func (csm *CoinSceneMgr) AudienceLeave(p *Player, reason int) bool {
	if p == nil {
		return false
	}
	//多平台支持
	platformName := Default_Platform
	platform := PlatformMgrSington.GetPlatform(p.Platform)
	if platform != nil && platform.Isolated {
		platformName = platform.IdStr
	}

	//没有场景，尝试创建
	if p.scene != nil {
		s := p.scene
		if id, ok := csm.playerIning[p.SnId]; ok {
			if s.groupId != 0 {
				if ss, ok := csm.scenesOfGroup[s.groupId]; ok {
					if csp, ok := ss[id]; ok && csp != nil {
						if csp.PlayerLeave(p, reason) {
							csm.OnPlayerLeave(s, p)
							return true
						}
					}
				}
			} else {
				//没有场景，尝试创建
				if ss, ok := csm.scenesOfPlatform[platformName]; ok {
					if csp, ok := ss[id]; ok && csp != nil {
						if csp.AudienceLeave(p, reason) {
							csm.OnPlayerLeave(s, p)
							return true
						}
					}
				}
			}
		}
	}
	return false
}

func (csm *CoinSceneMgr) OnDestroyScene(sceneid int) {
	if platformName, ok := csm.platformOfScene[int32(sceneid)]; ok {
		if ss, ok := csm.scenesOfPlatform[platformName]; ok {
			for _, csp := range ss {
				csp.OnDestroyScene(sceneid)
			}
		}
		delete(csm.platformOfScene, int32(sceneid))
	}

	if groupId, ok := csm.groupOfScene[int32(sceneid)]; ok {
		if ss, ok := csm.scenesOfGroup[groupId]; ok {
			for _, csp := range ss {
				csp.OnDestroyScene(sceneid)
			}
		}
		delete(csm.groupOfScene, int32(sceneid))
	}

	delete(csm.sceneOfcsp, sceneid)
}

func (csm *CoinSceneMgr) GetPlatformBySceneId(sceneid int) string {
	if platformName, ok := csm.platformOfScene[int32(sceneid)]; ok {
		return platformName
	}
	s := SceneMgrSington.GetScene(sceneid)
	if s != nil && s.limitPlatform != nil {
		return s.limitPlatform.IdStr
	}
	return Default_Platform
}

func (csm *CoinSceneMgr) GetPlayerNums(p *Player, gameId, gameMode int32) []int32 {
	//多平台支持
	platformName := Default_Platform
	platform := PlatformMgrSington.GetPlatform(p.Platform)
	if platform != nil && platform.Isolated {
		platformName = platform.IdStr
	} else if p.Platform != Default_Platform {
		platform = PlatformMgrSington.GetPlatform(Default_Platform)
	}
	var nums [CoinSceneType_Max]int32
	wantNum := []int32{80, 50, 30, 20, 10}
	for i := 0; i < CoinSceneType_Max; i++ {
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
				if ss, exist := csm.scenesOfGroup[gps.GroupId]; exist {
					if csp, exist := ss[id]; exist {
						sceneType := csp.GetSceneType() - 1
						//体验场 todo 这个地方为了兼容客户端的问题，客户端的体验场是4，至尊场是5，但是百人又没有处理百人又是4,抢庄又是9
						if sceneType == CoinSceneType_Experience {
							sceneType = CoinSceneType_Professor
						} else if sceneType == -2 {
							sceneType = CoinSceneType_Experience
						}
						if sceneType >= CoinSceneType_Max {
							sceneType = CoinSceneType_Experience
						}
						if sceneType >= 0 && sceneType < CoinSceneType_Max {
							nums[sceneType] += csp.GetPlayerNum() + csp.GetFakePlayerNum()
						}
					}
				}
			} else {
				if ss, ok := csm.scenesOfPlatform[platformName]; ok {
					if csp, exist := ss[id]; exist {
						sceneType := csp.GetSceneType() - 1
						//体验场 todo 这个地方为了兼容客户端的问题，客户端的体验场是4，至尊场是5，但是百人又没有处理百人又是4,抢庄又是9
						if sceneType == CoinSceneType_Experience {
							sceneType = CoinSceneType_Professor
						} else if sceneType == -2 {
							sceneType = CoinSceneType_Experience
						}
						if sceneType >= CoinSceneType_Max {
							sceneType = CoinSceneType_Experience
						}
						if sceneType >= 0 && sceneType < CoinSceneType_Max {
							nums[sceneType] += csp.GetPlayerNum() + csp.GetFakePlayerNum()
						}
					}
				}
			}
		}
	}

	return nums[:]
}

func (csm *CoinSceneMgr) InCoinScene(p *Player) bool {
	if p == nil {
		logger.Logger.Tracef("(csm *CoinSceneMgr) InCoinScene p == nil snid:%v ", p.SnId)
		return false
	}
	if _, ok := csm.playerIning[p.SnId]; ok {
		return true
	}
	logger.Logger.Tracef("(csm *CoinSceneMgr) InCoinScene false snid:%v ", p.SnId)
	return false
}

func (csm *CoinSceneMgr) PlayerTryLeave(p *Player, id int32, isAudience bool) hall_proto.OpResultCode {
	if !csm.InCoinScene(p) {
		logger.Logger.Tracef("(csm *CoinSceneMgr) PlayerTryLeave !csm.InCoinScene(p) snid:%v ", p.SnId)
		csm.PlayerTryLeaveQueue(p)
		return hall_proto.OpResultCode_OPRC_Sucess
	}
	if p.scene == nil || p.scene.gameSess == nil {
		logger.Logger.Tracef("(csm *CoinSceneMgr) PlayerTryLeave p.scene == nil || p.scene.gameSess == nil snid:%v ", p.SnId)
		csm.PlayerTryLeaveQueue(p)
		return hall_proto.OpResultCode_OPRC_Sucess
	}
	//通知gamesrv托管
	if id, ok := csm.playerIning[p.SnId]; ok {
		op := common.CoinSceneOp_Leave
		if isAudience {
			op = common.CoinSceneOp_AudienceLeave
		}
		pack := &hall_proto.CSCoinSceneOp{
			Id:     proto.Int32(id),
			OpType: proto.Int32(op),
		}
		proto.SetDefaults(pack)
		common.TransmitToServer(p.sid, int(hall_proto.CoinSceneGamePacketID_PACKET_CS_COINSCENE_OP), pack, p.scene.gameSess.Session)
	}
	return hall_proto.OpResultCode_OPRC_OpYield
}
func (csm *CoinSceneMgr) PlayerTryLeaveQueue(p *Player) hall_proto.OpResultCode {
	if p.CoinSceneQueue != nil {
		p.CoinSceneQueue.QuitQueue(p.SnId)
		return hall_proto.OpResultCode_OPRC_Sucess
	}
	return hall_proto.OpResultCode_OPRC_Error
}
func (csm *CoinSceneMgr) PlayerQueueState(p *Player) {
	if p.CoinSceneQueue == nil {
		return
	}
	if p.CoinSceneQueue.InQueue(p) {
		pack := &hall_proto.SCCoinSceneQueueState{
			GameFreeId: proto.Int32(p.CoinSceneQueue.dbGameFree.GetId()),
			Count:      proto.Int32(p.CoinSceneQueueRound),
			Ts:         proto.Int64(p.EnterCoinSceneQueueTs - time.Now().Unix()),
		}
		p.SendToClient(int(hall_proto.CoinSceneGamePacketID_PACKET_SC_COINSCENE_QUEUESTATE), pack)
	}
	return
}
func (csm *CoinSceneMgr) PlayerInChanging(p *Player) bool {
	_, exist := csm.playerChanging[p.SnId]
	return exist
}

func (csm *CoinSceneMgr) ClearPlayerChanging(p *Player) {
	delete(csm.playerChanging, p.SnId)
}

func (csm *CoinSceneMgr) PlayerTryChange(p *Player, id int32, exclude []int32, isAudience bool) hall_proto.OpResultCode {
	if csm.InCoinScene(p) {
		return csm.StartChangeCoinSceneTransact(p, id, exclude, isAudience)
	}

	//多平台支持
	if !isAudience {
		return csm.PlayerEnter(p, id, 0, exclude, true)
	} else {
		return csm.AudienceEnter(p, id, 0, exclude, true)
	}
}
func (csm *CoinSceneMgr) StartChangeCoinSceneTransact(p *Player, id int32, exclude []int32, isAudience bool) hall_proto.OpResultCode {
	if p == nil || p.scene == nil {
		logger.Logger.Warnf("(csm *CoinSceneMgr) StartChangeCoinSceneTransact p == nil || p.scene == nil snid:%v id:%v", p.SnId, id)
		return hall_proto.OpResultCode_OPRC_Error
	}

	tNow := time.Now()
	if !p.lastChangeScene.IsZero() && tNow.Sub(p.lastChangeScene) < time.Second {
		logger.Logger.Warnf("(csm *CoinSceneMgr) StartChangeCoinSceneTransact !p.lastChangeScene.IsZero() && tNow.Sub(p.lastChangeScene) < time.Second snid:%v id:%v", p.SnId, id)
		return hall_proto.OpResultCode_OPRC_ChangeRoomTooOften
	}

	tnp := &transact.TransNodeParam{
		Tt:     common.TransType_CoinSceneChange,
		Ot:     transact.TransOwnerType(common.GetSelfSrvType()),
		Oid:    common.GetSelfSrvId(),
		AreaID: common.GetSelfAreaId(),
	}
	ctx := &CoinSceneChangeCtx{
		id:         id,
		isClub:     false,
		snid:       p.SnId,
		sceneId:    int32(p.scene.sceneId),
		exclude:    exclude,
		isAudience: isAudience,
	}
	tNode := transact.DTCModule.StartTrans(tnp, ctx, CoinSceneChangeTimeOut)
	if tNode != nil {
		tNode.Go(core.CoreObject())
		csm.playerChanging[p.SnId] = id
		p.lastChangeScene = tNow
		return hall_proto.OpResultCode_OPRC_Sucess
	}
	logger.Logger.Warnf("(csm *CoinSceneMgr) StartChangeCoinSceneTransact tNode == nil snid:%v id:%v", p.SnId, id)
	return hall_proto.OpResultCode_OPRC_Error
}

func (csm *CoinSceneMgr) ModuleName() string {
	return "CoinSceneMgr"
}

func (csm *CoinSceneMgr) Init() {
	var errorCoinPool []int32
	for platformName, platform := range PlatformMgrSington.Platforms {
		if platform.Isolated || platformName == "" {
			arr := srvdata.PBDB_GameFreeMgr.Datas.GetArr()
			for _, dbGame := range arr {
				gps := PlatformMgrSington.GetGameConfig(platform.IdStr, dbGame.GetId())
				if gps != nil {
					dbGameFree := gps.DbGameFree
					if gps.GroupId != 0 {
						pgg := PlatformGameGroupMgrSington.GetGameGroup(gps.GroupId)
						if pgg != nil {
							dbGameFree = pgg.DbGameFree
						}
					}
					csp := NewCoinScenePool(dbGame.GetId(), gps.GroupId, dbGameFree)
					if csp != nil {
						if gps.GroupId != 0 {
							if ss, exist := csm.scenesOfGroup[gps.GroupId]; exist {
								ss[dbGame.GetId()] = csp
							} else {
								ss = make(map[int32]*CoinScenePool)
								ss[dbGame.GetId()] = csp
								csm.scenesOfGroup[gps.GroupId] = ss
							}
						} else {
							if ss, exist := csm.scenesOfPlatform[platformName]; exist {
								ss[dbGame.GetId()] = csp
							} else {
								ss = make(map[int32]*CoinScenePool)
								ss[dbGame.GetId()] = csp
								csm.scenesOfPlatform[platformName] = ss
							}
						}
					} else {
						errorCoinPool = append(errorCoinPool, dbGame.GetId())
					}
				}
			}
		}
	}
	//pre creat coinscene queue
	for pltId, platform := range PlatformMgrSington.Platforms {
		if pltId == Default_Platform {
			continue
		}
		arr := srvdata.PBDB_GameFreeMgr.Datas.GetArr()
		for _, dbGame := range arr {
			if dbGame.GetGameId() <= 0 {
				continue
			}
			gps := PlatformMgrSington.GetGameConfig(platform.IdStr, dbGame.GetId())
			if gps != nil {
				dbGameFree := gps.DbGameFree
				if gps.GroupId != 0 {
					pgg := PlatformGameGroupMgrSington.GetGameGroup(gps.GroupId)
					if pgg != nil {
						dbGameFree = pgg.DbGameFree
					}
				}
				if dbGameFree.GetMatchMode() == 1 {
					csp := NewCoinScenePool(dbGame.GetId(), gps.GroupId, dbGameFree)
					if csp != nil {
						if gps.GroupId != 0 {
							if ss, exist := csm.scenesOfGroup[gps.GroupId]; exist {
								ss[dbGame.GetId()] = csp
							} else {
								ss = make(map[int32]*CoinScenePool)
								ss[dbGame.GetId()] = csp
								csm.scenesOfGroup[gps.GroupId] = ss
							}
						} else {
							if ss, exist := csm.scenesOfPlatform[pltId]; exist {
								ss[dbGame.GetId()] = csp
							} else {
								ss = make(map[int32]*CoinScenePool)
								ss[dbGame.GetId()] = csp
								csm.scenesOfPlatform[pltId] = ss
							}
						}
					} else {
						errorCoinPool = append(errorCoinPool, dbGame.GetId())
					}
				}
			} else {
				logger.Logger.Errorf("PlatformName[%v] hasn't config[%v].", pltId, dbGame.GetId())
			}
		}
	}
	if len(errorCoinPool) > 0 {
		if len(errorCoinPool) > 10 {
			logger.Logger.Errorf("More than %v coin scene pool init failed.", errorCoinPool[:15])
		} else {
			logger.Logger.Errorf("%v coin scene pool init failed.", errorCoinPool)
		}
	}
}

// 撮合
func (csm *CoinSceneMgr) Update() {
	//预创建房间
	if time.Now().Unix() > csm.CreateRoomTick {
		csm.CreateRoomByCache()
		csm.CreateRoomTick = time.Now().Add(time.Second * 2).Unix()
	}

	for platform, v := range csm.scenesOfPlatform {
		if platform == "0" {
			continue
		}
		for _, csp := range v {
			csp.TryRefreshDeviation()
			csp.ProcessQueue(-1)
			csp.UpdateAndCleanQueue()
			csp.InviteRob(platform)
			csp.DismissRob(platform)
		}
	}
	for _, v := range csm.scenesOfGroup {
		for _, csp := range v {
			csp.ProcessQueue(-1)
			csp.UpdateAndCleanQueue()
			platform := PlatformMgrSington.GetPlatformByGroup(csp.groupId)
			if len(platform) != 0 {
				csp.InviteRob(platform)
				csp.DismissRob(platform)
			}
		}
	}
}
func (csm *CoinSceneMgr) Shutdown() {
	module.UnregisteModule(csm)
}

// 金豆在使用中
func (this *Player) CoinInUse() bool {
	return CoinSceneMgrSington.InCoinScene(this) || HundredSceneMgrSington.InHundredScene(this)
}

func (this *CoinSceneMgr) OnPlatformCreate(p *Platform) {
	if p.IdStr == "0" {
		return
	}
	//获取配置
	gps := PlatformMgrSington.GetPlatformGameConfig(p.IdStr)
	for _, v := range gps {
		if v.Status && v.DbGameFree.GetCreateRoomNum() > 0 {
			this.delayCache = append(this.delayCache, CreateRoomCache{gameFreeId: v.DbGameFree.Id, platformName: p.IdStr})
		}
	}
}

func (this *CoinSceneMgr) OnPlatformDestroy(p *Platform) {
	if p == nil {
		return
	}
	if csps, ok := this.scenesOfPlatform[p.IdStr]; ok {
		for _, csp := range csps {
			pack := &server_proto.WGGraceDestroyScene{}
			for _, scene := range csp.scenes {
				pack.Ids = append(pack.Ids, int32(scene.sceneId))
			}
			srvlib.ServerSessionMgrSington.Broadcast(int(server_proto.SSPacketID_PACKET_WG_GRACE_DESTROYSCENE), pack, common.GetSelfAreaId(), srvlib.GameServerType)
		}
	}
}

func (this *CoinSceneMgr) OnPlatformChangeIsolated(p *Platform, isolated bool) {
	if p == nil {
		return
	}
	if !isolated {
		if csps, ok := this.scenesOfPlatform[p.IdStr]; ok {
			for _, csp := range csps {
				pack := &server_proto.WGGraceDestroyScene{}
				for _, scene := range csp.scenes {
					pack.Ids = append(pack.Ids, int32(scene.sceneId))
				}
				srvlib.ServerSessionMgrSington.Broadcast(int(server_proto.SSPacketID_PACKET_WG_GRACE_DESTROYSCENE), pack, common.GetSelfAreaId(), srvlib.GameServerType)
			}
		}
	}
}

func (this *CoinSceneMgr) OnPlatformChangeDisabled(p *Platform, disabled bool) {
	if p == nil {
		return
	}
	if disabled {
		if csps, ok := this.scenesOfPlatform[p.IdStr]; ok {
			for _, csp := range csps {
				pack := &server_proto.WGGraceDestroyScene{}
				for _, scene := range csp.scenes {
					pack.Ids = append(pack.Ids, int32(scene.sceneId))
				}
				srvlib.ServerSessionMgrSington.Broadcast(int(server_proto.SSPacketID_PACKET_WG_GRACE_DESTROYSCENE), pack, common.GetSelfAreaId(), srvlib.GameServerType)
			}
		}
	}
}

func (this *CoinSceneMgr) OnPlatformConfigUpdate(p *Platform, oldCfg, newCfg *webapi.GameFree) {
	if p == nil || newCfg == nil {
		return
	}

	if scenes, exist := this.scenesOfPlatform[p.IdStr]; exist {

		if cps, ok := scenes[newCfg.DbGameFree.Id]; ok {
			cps.dbGameFree = newCfg.DbGameFree
			pack := &server_proto.WGGraceDestroyScene{}
			for _, scene := range cps.scenes {
				pack.Ids = append(pack.Ids, int32(scene.sceneId))
			}
			srvlib.ServerSessionMgrSington.Broadcast(int(server_proto.SSPacketID_PACKET_WG_GRACE_DESTROYSCENE),
				pack, common.GetSelfAreaId(), srvlib.GameServerType)

			//预创建房间配置更新
			if newCfg.DbGameFree.GetCreateRoomNum() != 0 && p.Name != "0" {
				logger.Logger.Tracef(">>>预创建房间 platform:%v %v_%v gamefreeid:%v CreateRoomNum:%v", p.Name,
					newCfg.DbGameFree.GetName(), newCfg.DbGameFree.GetTitle(), newCfg.DbGameFree.GetId(), newCfg.DbGameFree.GetCreateRoomNum())
				this.delayCache = append(this.delayCache, CreateRoomCache{gameFreeId: newCfg.DbGameFree.GetId(), platformName: p.IdStr})
			}
		}
	}
}

func (this *CoinSceneMgr) OnGameGroupUpdate(oldCfg, newCfg *webapi.GameConfigGroup) {
	if newCfg == nil {
		return
	}
	if scenes, exist := this.scenesOfGroup[newCfg.Id]; exist {
		if cps, ok := scenes[newCfg.DbGameFree.Id]; ok {
			needDestroy := false
			if cps.dbGameFree.GetBot() != newCfg.DbGameFree.GetBot() ||
				cps.dbGameFree.GetBaseScore() != newCfg.DbGameFree.GetBaseScore() ||
				cps.dbGameFree.GetLimitCoin() != newCfg.DbGameFree.GetLimitCoin() ||
				cps.dbGameFree.GetMaxCoinLimit() != newCfg.DbGameFree.GetMaxCoinLimit() ||
				cps.dbGameFree.GetTaxRate() != newCfg.DbGameFree.GetTaxRate() ||
				!common.SliceInt32Equal(cps.dbGameFree.GetOtherIntParams(), newCfg.DbGameFree.GetOtherIntParams()) ||
				!common.SliceInt32Equal(cps.dbGameFree.GetRobotTakeCoin(), newCfg.DbGameFree.GetRobotTakeCoin()) ||
				!common.SliceInt32Equal(cps.dbGameFree.GetRobotLimitCoin(), newCfg.DbGameFree.GetRobotLimitCoin()) {
				needDestroy = true
			}
			//TODO 预创建房间配置更新,unsupport group model
			cps.dbGameFree = newCfg.DbGameFree
			if needDestroy {
				pack := &server_proto.WGGraceDestroyScene{}
				for _, scene := range cps.scenes {
					pack.Ids = append(pack.Ids, int32(scene.sceneId))
				}
				srvlib.ServerSessionMgrSington.Broadcast(int(server_proto.SSPacketID_PACKET_WG_GRACE_DESTROYSCENE), pack, common.GetSelfAreaId(), srvlib.GameServerType)
			}
		}
	}
}

// 列出所有房间
func (csm *CoinSceneMgr) ListRooms(p *Player, id int32) bool {
	//多平台支持
	platformName := Default_Platform
	platform := PlatformMgrSington.GetPlatform(p.Platform)
	if platform != nil && platform.Isolated {
		platformName = platform.IdStr
	} else {
		platform = PlatformMgrSington.GetPlatform(Default_Platform)
	}
	if platform == nil {
		return false
	}

	gps := PlatformMgrSington.GetGameConfig(platform.IdStr, id)
	if gps == nil {
		return false
	}

	//分组模式
	if gps.GroupId != 0 {
		pgg := PlatformGameGroupMgrSington.GetGameGroup(gps.GroupId)
		if pgg != nil {
			if _, ok := csm.scenesOfGroup[gps.GroupId]; !ok {
				csm.scenesOfGroup[gps.GroupId] = make(map[int32]*CoinScenePool)
			}
			if ss, ok := csm.scenesOfGroup[gps.GroupId]; ok {
				if csp, ok := ss[id]; ok && csp != nil {
					return csp.PlayerListRoom(p)
				}
				csp := NewCoinScenePool(id, gps.GroupId, pgg.DbGameFree)
				if csp == nil {
					return false
				}
				ss[id] = csp
				return csp.PlayerListRoom(p)
			}
			return false
		}
	}
	//独立模式
	if _, ok := csm.scenesOfPlatform[platformName]; !ok {
		csm.scenesOfPlatform[platformName] = make(map[int32]*CoinScenePool)
	}
	if ss, ok := csm.scenesOfPlatform[platformName]; ok {
		if csp, ok := ss[id]; ok && csp != nil {
			return csp.PlayerListRoom(p)
		}

		csp := NewCoinScenePool(id, 0, gps.DbGameFree)
		if csp == nil {
			return false
		}
		ss[id] = csp
		return csp.PlayerListRoom(p)
	}
	return false
}

func (csm *CoinSceneMgr) CreateRoomByCache() {
	cnt := len(csm.delayCache)
	if cnt > 0 {
		data := csm.delayCache[cnt-1]
		csm.delayCache = csm.delayCache[:cnt-1]
		pdd := PlatformMgrSington.GetGameConfig(data.platformName, data.gameFreeId)
		if pdd != nil {
			//分组模式
			if pdd.GroupId != 0 {
				pgg := PlatformGameGroupMgrSington.GetGameGroup(pdd.GroupId)
				if pgg != nil {
					if _, ok := csm.scenesOfGroup[pdd.GroupId]; !ok {
						csm.scenesOfGroup[pdd.GroupId] = make(map[int32]*CoinScenePool)
					}
					if ss, ok := csm.scenesOfGroup[pdd.GroupId]; ok {
						if csp, ok := ss[pdd.DbGameFree.Id]; ok && csp != nil {
							csp.EnsurePreCreateRoom(data.platformName)
							return
						} else {
							csp := NewCoinScenePool(pdd.DbGameFree.Id, pdd.GroupId, pgg.DbGameFree)
							if csp != nil {
								ss[pdd.DbGameFree.Id] = csp
								csp.EnsurePreCreateRoom(data.platformName)
								return
							}
						}
					}
				}
			}
			//独立模式
			if _, ok := csm.scenesOfPlatform[data.platformName]; !ok {
				csm.scenesOfPlatform[data.platformName] = make(map[int32]*CoinScenePool)
			}
			if ss, ok := csm.scenesOfPlatform[data.platformName]; ok {
				if csp, ok := ss[pdd.DbGameFree.Id]; ok && csp != nil {
					csp.EnsurePreCreateRoom(data.platformName)
					return
				} else {
					csp := NewCoinScenePool(pdd.DbGameFree.Id, 0, pdd.DbGameFree)
					if csp != nil {
						ss[pdd.DbGameFree.Id] = csp
						csp.EnsurePreCreateRoom(data.platformName)
						return
					}
				}
			}
		}
	}
}

func (csm *CoinSceneMgr) OnGameSessionRegiste(gs *GameSession) {
	wildGs := len(gs.gameIds) == 0 || common.InSliceInt32(gs.gameIds, 0)
	for _, platform := range PlatformMgrSington.Platforms {
		if platform.IdStr == "0" {
			continue
		}
		//获取配置
		gps := PlatformMgrSington.GetPlatformGameConfig(platform.IdStr)
		for _, v := range gps {
			if v.Status && v.DbGameFree.GetCreateRoomNum() > 0 && (wildGs || common.InSliceInt32(gs.gameIds, v.DbGameFree.GetGameId())) {
				csm.delayCache = append(csm.delayCache, CreateRoomCache{gameFreeId: v.DbGameFree.Id, platformName: platform.IdStr})
			}
		}
	}
}

func (this *CoinSceneMgr) OnGameSessionUnregiste(gs *GameSession) {

}

func (this *CoinSceneMgr) OnPlatformDestroyByGameFreeId(p *Platform, gameFreeId int32) {
	if p == nil {
		return
	}
	if csps, ok := this.scenesOfPlatform[p.IdStr]; ok {
		for _, csp := range csps {
			pack := &server_proto.WGGraceDestroyScene{}
			for _, scene := range csp.scenes {
				if scene.dbGameFree.Id == gameFreeId {
					pack.Ids = append(pack.Ids, int32(scene.sceneId))
				}
			}
			srvlib.ServerSessionMgrSington.Broadcast(int(server_proto.SSPacketID_PACKET_WG_GRACE_DESTROYSCENE), pack, common.GetSelfAreaId(), srvlib.GameServerType)
		}
	}
}

func init() {
	//定期撮合一次
	module.RegisteModule(CoinSceneMgrSington, time.Second, 0)
	PlatformMgrSington.RegisteObserver(CoinSceneMgrSington)
	PlatformGameGroupMgrSington.RegisteObserver(CoinSceneMgrSington)
	RegisteGameSessionListener(CoinSceneMgrSington)
}
