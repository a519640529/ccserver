package base

import (
	server_proto "games.yol.com/win88/protocol/server"
	"games.yol.com/win88/srvdata"
	"github.com/idealeak/goserver/core/module"
	"time"
)

var SceneMgrTick = time.Millisecond * 100 //发手牌等待时间

var SceneMgrSington = &SceneMgr{
	Scenes:          make(map[int32]Scene),
	sceneDBGameFree: make(map[int32]*server_proto.DB_GameFree),
}

type SceneMgr struct {
	Scenes          map[int32]Scene
	sceneDBGameFree map[int32]*server_proto.DB_GameFree
}

func (sm *SceneMgr) AddScene(s Scene) {
	sm.Scenes[s.GetRoomId()] = s
}

func (sm *SceneMgr) DelScene(sceneId int32) {
	delete(sm.Scenes, sceneId)
	delete(sm.sceneDBGameFree, sceneId)
}

func (sm *SceneMgr) GetScene(sceneId int32) Scene {
	if s, exist := sm.Scenes[sceneId]; exist {
		return s
	}
	return nil
}

func (sm *SceneMgr) GetOneNoFullScene() Scene {
	for _, s := range sm.Scenes {
		if !s.IsFull() {
			return s
		}
	}
	return nil
}

func (sm *SceneMgr) GetOneFullScene() Scene {
	for _, s := range sm.Scenes {
		if s.IsFull() {
			return s
		}
	}
	return nil
}

func (sm *SceneMgr) UpdateSceneDBGameFree(sceneId int32, dbGameFree *server_proto.DB_GameFree) {
	sm.sceneDBGameFree[sceneId] = dbGameFree
}

func (sm *SceneMgr) GetSceneDBGameFree(sceneId, gamefreeId int32) *server_proto.DB_GameFree {
	if data, exist := sm.sceneDBGameFree[sceneId]; exist {
		return data
	}

	return srvdata.PBDB_GameFreeMgr.GetData(gamefreeId)
}

////////////////////////////////////////////////////////////////////
/// Module Implement [beg]
////////////////////////////////////////////////////////////////////
func (sm *SceneMgr) Update() {
	for _, s := range sm.Scenes {
		s.Update(int64(SceneMgrTick))
	}
	return
}

func (sm *SceneMgr) Init() {
}

func (sm *SceneMgr) ModuleName() string {
	return "SceneMgr-module"
}

func (sm *SceneMgr) Shutdown() {
	module.UnregisteModule(sm)
}

func init() {
	module.RegisteModule(SceneMgrSington, SceneMgrTick, 0)
}
