package base

import (
	"games.yol.com/win88/protocol/server"
	"time"

	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/module"
	"github.com/idealeak/goserver/core/netlib"
)

var SceneMgrSington = &SceneMgr{
	scenes:           make(map[int]*Scene),
	scenesByGame:     make(map[int]map[int]*Scene),
	scenesByGameFree: make(map[int32]map[int]*Scene),
}

type SceneMgr struct {
	scenes           map[int]*Scene
	scenesByGame     map[int]map[int]*Scene
	scenesByGameFree map[int32]map[int]*Scene
}

func (this *SceneMgr) makeKey(gameid, gamemode int) int {
	return int(gameid*10000 + gamemode)
}

func (this *SceneMgr) CreateScene(s *netlib.Session, sceneId, gameMode, sceneMode, gameId int, platform string,
	params []int32, agentor, creator int32, replayCode string, hallId, groupId, totalOfGames int32,
	dbGameFree *server.DB_GameFree, bEnterAfterStart bool, baseScore int32, playerNum int, paramsEx ...int32) *Scene {
	scene := NewScene(s, sceneId, gameMode, sceneMode, gameId, platform, params, agentor, creator, replayCode,
		hallId, groupId, totalOfGames, dbGameFree, bEnterAfterStart, baseScore, playerNum, paramsEx...)
	if scene == nil {
		logger.Logger.Error("(this *SceneMgr) CreateScene, scene == nil")
		return nil
	}
	this.scenes[scene.SceneId] = scene
	//
	key := this.makeKey(gameId, gameMode)
	if ss, exist := this.scenesByGame[key]; exist {
		ss[scene.SceneId] = scene
	} else {
		ss = make(map[int]*Scene)
		ss[scene.SceneId] = scene
		this.scenesByGame[key] = ss
	}
	//
	if ss, exist := this.scenesByGameFree[dbGameFree.GetId()]; exist {
		ss[scene.SceneId] = scene
	} else {
		ss = make(map[int]*Scene)
		ss[scene.SceneId] = scene
		this.scenesByGameFree[dbGameFree.GetId()] = ss
	}
	scene.OnStart()
	logger.Logger.Infof("(this *SceneMgr) CreateScene,New scene,id:[%d] replaycode:[%v]", scene.SceneId, replayCode)
	return scene
}

func (this *SceneMgr) DestroyScene(sceneId int) {
	if scene, exist := this.scenes[sceneId]; exist {
		scene.OnStop()
		//
		key := this.makeKey(scene.GameId, scene.GameMode)
		if ss, exist := this.scenesByGame[key]; exist {
			delete(ss, scene.SceneId)
		}
		//
		if ss, exist := this.scenesByGameFree[scene.GetGameFreeId()]; exist {
			delete(ss, scene.SceneId)
		}
		delete(this.scenes, sceneId)
		logger.Logger.Infof("(this *SceneMgr) DestroyScene, sceneid = %v", sceneId)
	}
}

func (this *SceneMgr) GetPlayerNumByGameFree(platform string, gamefreeid, groupId int32) int32 {
	var num int32
	if ss, exist := SceneMgrSington.scenesByGameFree[gamefreeid]; exist {
		for _, scene := range ss {
			if groupId != 0 {
				if scene.GroupId == groupId {
					cnt := scene.GetRealPlayerCnt()
					num += int32(cnt)
				}
			} else {
				if scene.Platform == platform {
					cnt := scene.GetRealPlayerCnt()
					num += int32(cnt)
				}
			}
		}
	}
	return num
}

func (this *SceneMgr) GetPlayerNumByGame(platform string, gameid, gamemode, groupId int32) map[int32]int32 {
	nums := make(map[int32]int32)
	key := this.makeKey(int(gameid), int(gamemode))
	if ss, exist := SceneMgrSington.scenesByGame[key]; exist {
		for _, scene := range ss {
			if groupId != 0 {
				if scene.GroupId == groupId {
					cnt := scene.GetRealPlayerCnt()
					nums[scene.GetGameFreeId()] = nums[scene.GetGameFreeId()] + int32(cnt)
				}
			} else {
				if scene.Platform == platform {
					cnt := scene.GetRealPlayerCnt()
					nums[scene.GetGameFreeId()] = nums[scene.GetGameFreeId()] + int32(cnt)
				}
			}
		}
	}
	return nums
}

func (this *SceneMgr) GetPlayersByGameFree(platform string, gamefreeid int32) []*Player {
	players := make([]*Player, 0)
	if ss, exist := SceneMgrSington.scenesByGameFree[gamefreeid]; exist {
		for _, scene := range ss {
			if scene.Platform == platform {
				for _, p := range scene.Players {
					if p != nil {
						players = append(players, p)
					}
				}
			}
		}
	}
	return players
}

func (this *SceneMgr) GetScene(sceneId int) *Scene {
	if s, exist := this.scenes[sceneId]; exist {
		return s
	}
	return nil
}

func (this *SceneMgr) OnMiniTimer() {
	for _, scene := range this.scenes {
		scene.SyncPlayerCoin()
	}
}

func (this *SceneMgr) OnHourTimer() {
	//	for _, scene := range this.scenes {
	//		scene.OnHourTimer()
	//	}
}

func (this *SceneMgr) OnDayTimer() {
	logger.Logger.Info("(this *SceneMgr) OnDayTimer")
	//	for _, scene := range this.scenes {
	//		scene.OnDayTimer()
	//	}
	PlayerMgrSington.OnDayTimer()
}

func (this *SceneMgr) OnWeekTimer() {
	logger.Logger.Info("(this *SceneMgr) OnWeekTimer")
	//	for _, scene := range this.scenes {
	//		scene.OnWeekTimer()
	//	}
}

func (this *SceneMgr) OnMonthTimer() {
	logger.Logger.Info("(this *SceneMgr) OnMonthTimer")
	//	for _, scene := range this.scenes {
	//		scene.OnMonthTimer()
	//	}
}

func (this *SceneMgr) RebindPlayerSnId(oldSnId, newSnId int32) {
	for _, s := range this.scenes {
		s.RebindPlayerSnId(oldSnId, newSnId)
	}
}

func (this *SceneMgr) DestoryAllScene() {
	for _, s := range this.scenes {
		s.Destroy(true)
	}
}

// //////////////////////////////////////////////////////////////////
// / Module Implement [beg]
// //////////////////////////////////////////////////////////////////
func (this *SceneMgr) ModuleName() string {
	return "SceneMgr"
}

func (this *SceneMgr) Init() {
	//RestoreMDump(dumpFileName)
}

func (this *SceneMgr) Update() {
	for _, s := range this.scenes {
		s.OnTick()
	}
}

func (this *SceneMgr) Shutdown() {
	//WriteMDump(dumpFileName)
	for _, s := range this.scenes {
		s.Destroy(true)
	}
	module.UnregisteModule(this)
}

func init() {
	module.RegisteModule(SceneMgrSington, time.Millisecond*50, 0)
	//RegisteDayTimeChangeListener(SceneMgrSington)
}
