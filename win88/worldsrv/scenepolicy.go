package main

import "github.com/idealeak/goserver/core/logger"

// 根据不同的房间模式,选择不同的房间业务策略
var ScenePolicyPool map[int]map[int]ScenePolicy = make(map[int]map[int]ScenePolicy)

type ScenePolicy interface {
	//场景开启事件
	OnStart(s *Scene)
	//场景关闭事件
	OnStop(s *Scene)
	//场景心跳事件
	OnTick(s *Scene)
	//玩家进入事件
	OnPlayerEnter(s *Scene, p *Player)
	//玩家离开事件
	OnPlayerLeave(s *Scene, p *Player)
	//系统维护关闭事件
	OnShutdown(s *Scene)
	//获得场景的匹配因子(值越大越优先选择)
	GetFitFactor(s *Scene, p *Player) int
	//能否创建
	CanCreate(s *Scene, p *Player, mode, sceneType int, params []int32, isAgent bool) (bool, []int32)
	//能否进入
	CanEnter(s *Scene, p *Player) int
	//房间座位是否已满
	IsFull(s *Scene, p *Player, num int32) bool
	//是否可以强制开始
	IsCanForceStart(s *Scene) bool
	//结算房卡
	BilledRoomCard(s *Scene, snid []int32)
	//需要几张房卡
	GetNeedRoomCardCnt(s *Scene) int32
	//房费模式
	GetRoomFeeMode(s *Scene) int32
	//游戏人数
	GetPlayerNum(s *Scene) int32
	//游戏总局数
	GetTotalOfGames(s *Scene) int32
	//
	GetNeedRoomCardCntDependentPlayerCnt(s *Scene) int32
	//
	GetBetState() int32
	GetViewLogLen() int32
}

func GetScenePolicy(gameId, mode int) ScenePolicy {
	if g, exist := ScenePolicyPool[gameId]; exist {
		if p, exist := g[mode]; exist {
			return p
		}
	}
	return nil
}

func RegisteScenePolicy(gameId, mode int, sp ScenePolicy) {
	pool := ScenePolicyPool[gameId]
	if pool == nil {
		pool = make(map[int]ScenePolicy)
		ScenePolicyPool[gameId] = pool
	}
	if pool != nil {
		pool[mode] = sp
	}
}

func CheckGameConfigVer(ver int32, gameid int32, modeid int32) bool {
	sp := GetScenePolicy(int(gameid), int(modeid))
	if sp == nil {
		return true
	}
	spd, ok := sp.(*ScenePolicyData)
	if !ok {
		return true
	}
	logger.Logger.Info("Old game config ver:", spd.ConfigVer)
	if ver == spd.ConfigVer {
		return true
	}
	return false
}
func GetGameConfigVer(gameid int32, modeid int32) int32 {
	sp := GetScenePolicy(int(gameid), int(modeid))
	if sp == nil {
		return 0
	}
	spd, ok := sp.(*ScenePolicyData)
	if !ok {
		return 0
	}
	return spd.ConfigVer
}
