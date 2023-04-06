package base

import (
	"time"
)

// 根据不同的房间模式,选择不同的房间业务逻辑
var ScenePolicyPool map[int]map[int]ScenePolicy = make(map[int]map[int]ScenePolicy)

type ScenePolicy interface {
	//心跳间隔
	GetHeartBeatInterval() time.Duration
	//场景开启事件
	OnStart(s *Scene)
	//场景关闭事件
	OnStop(s *Scene)
	//场景心跳事件
	OnTick(s *Scene)
	//玩家进入事件
	OnPlayerEnter(s *Scene, p *Player)
	//玩家离开事件
	OnPlayerLeave(s *Scene, p *Player, reason int)
	//玩家掉线
	OnPlayerDropLine(s *Scene, p *Player)
	//玩家重连
	OnPlayerRehold(s *Scene, p *Player)
	//玩家重新返回房间
	OnPlayerReturn(s *Scene,p *Player)
	//玩家操作
	OnPlayerOp(s *Scene, p *Player, opcode int, params []int64) bool
	//玩家操作
	OnPlayerOperate(s *Scene, p *Player, params interface{}) bool
	//玩家事件
	OnPlayerEvent(s *Scene, p *Player, evtcode int, params []int64)
	//
	GetSceneState(s *Scene, stateid int) SceneState
	//是否完成了整个牌局
	IsCompleted(s *Scene) bool
	//是否可以强制开始
	IsCanForceStart(s *Scene) bool
	//强制开始
	ForceStart(s *Scene)
	//当前状态能否换桌
	CanChangeCoinScene(s *Scene, p *Player) bool
	//创建场景扩展数据
	CreateSceneExData(s *Scene) interface{}
	//创建玩家扩展数据
	CreatePlayerExData(s *Scene, p *Player) interface{}
	//
	PacketGameData(s *Scene) interface{}
	//
	InterventionGame(s *Scene, data interface{}) interface{}
	//通知分场状态
	NotifyGameState(s *Scene)
}

//场景状态接口
type SceneState interface {
	GetState() int                                                   //状态ID
	CanChangeTo(s SceneState) bool                                   //切换到指定状态
	CanChangeScene(s *Scene, p *Player) bool                         //当前状态能否换桌
	GetTimeout(s *Scene) int                                         //超时时间
	OnEnter(s *Scene)                                                //状态进入时
	OnLeave(s *Scene)                                                //状态离开时
	OnTick(s *Scene)                                                 //状态tick
	OnPlayerOp(s *Scene, p *Player, opcode int, params []int64) bool //玩家操作
	OnPlayerEvent(s *Scene, p *Player, evtcode int, params []int64)  //玩家事件
}

type SceneStateOperate interface {
	OnPlayerOperate(s *Scene, p *Player, params interface{}) bool
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

type BaseScenePolicy struct {
}

func (bsp *BaseScenePolicy) GetHeartBeatInterval() time.Duration           { return time.Second }
func (bsp *BaseScenePolicy) OnStart(s *Scene)                              {}
func (bsp *BaseScenePolicy) OnStop(s *Scene)                               {}
func (bsp *BaseScenePolicy) OnTick(s *Scene)                               {}
func (bsp *BaseScenePolicy) OnPlayerEnter(s *Scene, p *Player)             {}
func (bsp *BaseScenePolicy) OnPlayerLeave(s *Scene, p *Player, reason int) {}
func (bsp *BaseScenePolicy) OnPlayerDropLine(s *Scene, p *Player)          {}
func (bsp *BaseScenePolicy) OnPlayerRehold(s *Scene, p *Player)            {}
func (bsp *BaseScenePolicy) OnPlayerReturn(s *Scene,p *Player)			   {}
func (bsp *BaseScenePolicy) OnPlayerOp(s *Scene, p *Player, op int, params []int64) bool {
	return false
}
func (bsp *BaseScenePolicy) OnPlayerOperate(s *Scene, p *Player, params interface{}) bool {
	return false
}
func (bsp *BaseScenePolicy) OnPlayerEvent(s *Scene, p *Player, evtcode int, params []int64) {}
func (bsp *BaseScenePolicy) GetSceneState(s *Scene, stateid int) SceneState                 { return G_BaseSceneState }
func (bsp *BaseScenePolicy) IsCompleted(s *Scene) bool                                      { return false }
func (bsp *BaseScenePolicy) IsCanForceStart(s *Scene) bool                                  { return false }
func (bsp *BaseScenePolicy) ForceStart(s *Scene)                                            {}
func (bsp *BaseScenePolicy) CanChangeCoinScene(s *Scene, p *Player) bool                    { return false }
func (bsp *BaseScenePolicy) CreateSceneExData(s *Scene) interface{}                         { return false }
func (bsp *BaseScenePolicy) CreatePlayerExData(s *Scene, p *Player) interface{}             { return false }
func (bsp *BaseScenePolicy) PacketGameData(s *Scene) interface{}                            { return nil }
func (bsp *BaseScenePolicy) InterventionGame(s *Scene, data interface{}) interface{}        { return nil }
func (bsp *BaseScenePolicy) NotifyGameState(s *Scene)                                       {}

var G_BaseSceneState = &BaseSceneState{}

type BaseSceneState struct {
}

func (bst *BaseSceneState) GetState() int                           { return -1 }
func (bst *BaseSceneState) CanChangeTo(s SceneState) bool           { return false }
func (bst *BaseSceneState) CanChangeScene(s *Scene, p *Player) bool { return false }
func (bst *BaseSceneState) GetTimeout(s *Scene) int                 { return 0 }
func (bst *BaseSceneState) OnEnter(s *Scene)                        {}
func (bst *BaseSceneState) OnLeave(s *Scene)                        {}
func (bst *BaseSceneState) OnTick(s *Scene)                         {}
func (bst *BaseSceneState) OnPlayerOp(s *Scene, p *Player, opcode int, params []int64) bool {
	return false
}
func (bst *BaseSceneState) OnPlayerEvent(s *Scene, p *Player, evtcode int, params []int64) {}
