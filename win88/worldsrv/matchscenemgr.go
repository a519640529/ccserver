package main

import (
	"games.yol.com/win88/common"
	"games.yol.com/win88/srvdata"
	"github.com/idealeak/goserver/core/logger"
)

var MatchSceneMgrSington = &MatchSceneMgr{
	scenes: make(map[int]*Scene),
}

type MatchSceneMgr struct {
	scenes map[int]*Scene
}

func (ms *MatchSceneMgr) MatchStart(tm *TmMatch) {

	scene := ms.FetchOneScene(tm, false, 1)
	if scene != nil {
		ms.scenes[scene.sceneId] = scene
	}
	for _, tmp := range tm.TmPlayer {
		if scene == nil || scene.IsFull() {
			scene = ms.FetchOneScene(tm, false, 1)
		}
		p := PlayerMgrSington.GetPlayerBySnId(tmp.SnId)
		if p != nil && p.scene == nil {
			mc := TournamentMgr.CreatePlayerMatchContext(p, tm, tmp.seq)
			if mc != nil {
				mc.gaming = true
				scene.PlayerEnter(mc.p, -1, true)
			}
		} else {
			if p != nil {
				logger.Logger.Error("MatchStart error: snid: ", p.SnId, " p.scene: ", p.scene)
			}
		}
	}
}

func (ms *MatchSceneMgr) MatchStop(tm *TmMatch) {
	if SceneMgrSington.scenes != nil && tm != nil {
		for _, scene := range SceneMgrSington.scenes {
			if scene.IsMatchScene() && scene.matchId == tm.SortId {
				scene.matchStop = true
				scene.ForceDelete(false)
			}
		}
	}
}

func (ms *MatchSceneMgr) NewRoundStart(tm *TmMatch, mct []*MatchContext, finals bool, round int32) {
	scene := ms.FetchOneScene(tm, finals, round)
	if scene != nil {
		ms.scenes[scene.sceneId] = scene
	}
	for _, tmp := range mct {
		if scene == nil || scene.IsFull() {
			scene = ms.FetchOneScene(tm, finals, round)
		}
		p := tmp.p
		if p != nil && p.scene == nil {
			if mc, ok := TournamentMgr.players[tm.SortId][p.SnId]; ok {
				mc.gaming = true
				mc.grade = mc.grade * 75 / 100 //积分衰减
				mc.rank = tmp.rank
				scene.PlayerEnter(mc.p, -1, true)
			}
		} else {
			if p != nil {
				logger.Logger.Error("NewRoundStart error: snid: ", p.SnId, " p.scene: ", p.scene)
			}
		}
	}
}

// 参数说明：是不是决赛，当前轮数
func (ms *MatchSceneMgr) FetchOneScene(tm *TmMatch, finals bool, round int32) *Scene {
	msp := NewMatchScenePool(tm)
	scene := msp.CreateTournamentScene(finals, round)
	return scene
}

type MatchScenePool struct {
	gameFreeId int32 //gamefreeid
	tm         *TmMatch
}

func NewMatchScenePool(tm *TmMatch) *MatchScenePool {
	csp := &MatchScenePool{
		gameFreeId: tm.dbGameFree.Id,
		tm:         tm,
	}
	if !csp.init() {
		return nil
	}
	return csp
}
func (msp *MatchScenePool) init() bool {
	if msp.tm.dbGameFree == nil {
		msp.tm.dbGameFree = srvdata.PBDB_GameFreeMgr.GetData(msp.gameFreeId)
	}
	if msp.tm.dbGameFree == nil {
		logger.Logger.Errorf("match Coin scene pool init failed,%v game free data not find.", msp.gameFreeId)
		return false
	}
	return true
}

// 创建锦标赛房间
func (msp *MatchScenePool) CreateTournamentScene(matchFinals bool, round int32) *Scene {
	sceneId := SceneMgrSington.GenOneMatchSceneId()
	gameId := int(msp.tm.dbGameFree.GameId)
	gs := GameSessMgrSington.GetMinLoadSess(gameId)
	if gs != nil {
		gameMode := msp.tm.dbGameFree.GetGameMode()
		limitPlatform := PlatformMgrSington.GetPlatform(msp.tm.Platform)
		if limitPlatform == nil || !limitPlatform.Isolated {
			limitPlatform = PlatformMgrSington.GetPlatform(Default_Platform)
		}
		finals := int32(0)
		if matchFinals {
			finals = 1
		}
		curPlayerNum := int32(1)
		nextNeed := int32(0)
		if msp.tm.gmd != nil && len(msp.tm.gmd.MatchPromotion) >= int(round) {
			curPlayerNum = msp.tm.gmd.MatchPromotion[round-1]
			if curPlayerNum == 1 { //最后一局特殊处理下，取倒数第二局人数
				curPlayerNum = msp.tm.gmd.MatchPromotion[round-2]
			}
			if len(msp.tm.gmd.MatchPromotion) > int(round) {
				nextNeed = msp.tm.gmd.MatchPromotion[round]
			}
		}
		scene := SceneMgrSington.CreateScene(0, 0, sceneId, gameId, int(gameMode), common.SceneMode_Match, 1,
			0, []int32{msp.tm.SortId, finals, round, curPlayerNum, nextNeed, msp.tm.gmd.MatchType},
			gs, limitPlatform, 0, msp.tm.dbGameFree)
		if scene != nil {
			scene.matchId = msp.tm.SortId
			return scene
		}
	}
	return nil
}
