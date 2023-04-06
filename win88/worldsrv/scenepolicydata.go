package main

import (
	"time"

	"games.yol.com/win88/common"
	hall_proto "games.yol.com/win88/protocol/gamehall"
	"github.com/idealeak/goserver/core/logger"
)

const (
	SPDPCustomIndex_GamesOfCard  int = iota //局数选项
	SPDPCustomIndex_PlayerNum               //人数选项
	SPDPCustomIndex_RoomFeeMode             //房费模式选项
	SPDPCustomIndex_LimitOption             //新手限制选项
	SPDPCustomIndex_DoorOption              //中途不可进选项
	SPDPCustomIndex_SameIPForbid            //同IP不可进
)

// 创建房间选项
const (
	CreateRoomParam_NumOfGames   int = iota //局数选项
	CreateRoomParam_DoorOption              //中途允许加入选项
	CreateRoomParam_SameIPForbid            //同IP不可进
	CreateRoomParam_Max
)

type ScenePolicyDataParam struct {
	Name        string
	AliasName   string
	Desc        string
	Min         int32
	Max         int32
	Default     int32
	Value       []int32
	Value2      []int32
	CustomIndex int32
	index       int
}

type ScenePolicyData struct {
	GameName           string
	GameId             int32
	GameMode           []int32
	CanForceStart      bool
	MinPlayerCnt       int32
	DefaultPlayerCnt   int32
	MaxIndex           int32
	TimeFreeStart      int32
	TimeFreeEnd        int32
	DependentPlayerCnt bool
	EnterAfterStart    bool
	ConfigVer          int32
	PerGameTakeCard    int32
	ViewLogCnt         int32
	BetState           int32
	Params             []ScenePolicyDataParam
	nameMap            map[string]*ScenePolicyDataParam
	aliasNameMap       map[string]*ScenePolicyDataParam
	customIndexParams  []*ScenePolicyDataParam
}

func alignto(val, align int32) int32 {
	return (val + align - 1) / align
}

func (spd *ScenePolicyData) Init() bool {
	spd.nameMap = make(map[string]*ScenePolicyDataParam)
	spd.aliasNameMap = make(map[string]*ScenePolicyDataParam)
	if spd.MaxIndex > 0 {
		spd.customIndexParams = make([]*ScenePolicyDataParam, spd.MaxIndex, spd.MaxIndex)
	}
	for i := 0; i < len(spd.Params); i++ {
		spd.nameMap[spd.Params[i].Name] = &spd.Params[i]
		spd.Params[i].index = i
		spd.aliasNameMap[spd.Params[i].AliasName] = &spd.Params[i]
		if spd.Params[i].CustomIndex >= 0 && spd.MaxIndex > 0 {
			spd.customIndexParams[spd.Params[i].CustomIndex] = &spd.Params[i]
		}
	}
	return true
}

func (spd *ScenePolicyData) GetParam(idx int) *ScenePolicyDataParam {
	if idx >= 0 && idx < len(spd.Params) {
		return &spd.Params[idx]
	}
	return nil
}

func (spd *ScenePolicyData) GetParamByIndex(idx int) *ScenePolicyDataParam {
	if idx >= 0 && idx < len(spd.customIndexParams) {
		return spd.customIndexParams[idx]
	}
	return nil
}

func (spd *ScenePolicyData) GetParamByName(name string) *ScenePolicyDataParam {
	if spdp, exist := spd.nameMap[name]; exist {
		return spdp
	}
	return nil
}

func (spd *ScenePolicyData) GetParamByAliasName(name string) *ScenePolicyDataParam {
	if spdp, exist := spd.aliasNameMap[name]; exist {
		return spdp
	}
	return nil
}

func (spd *ScenePolicyData) IsTimeFree(mode int) bool {
	//是否限时免费
	ts := int32(time.Now().Unix())
	if ts >= spd.TimeFreeStart && ts < spd.TimeFreeEnd {
		return true
	}
	return false
}

func (spd *ScenePolicyData) IsEnoughRoomCardCnt(s *Scene, p *Player, roomFeeParam, needCardCnt, playerNum int32, isAgent bool) bool {
	return true
}

func (spd *ScenePolicyData) CostRoomCard(s *Scene, roomFeeParam, needCardCnt, playerNum int32, snid []int32) {
}

func (spd *ScenePolicyData) UpRoomCard(s *Scene, roomFeeParam, needCardCnt, playerNum int32) {
}

//ScenePolicy interface

// 能否进入
func (spd *ScenePolicyData) CanCreate(s *Scene, p *Player, mode, sceneType int, params []int32, isAgent bool) (bool, []int32) {
	//参数容错处理
	if len(params) < len(spd.Params) {
		logger.Logger.Infof("ScenePolicyData.CanCreate param count not enough, need:%v get:%v", len(spd.Params), len(params))
		for i := len(params); i < len(spd.Params); i++ {
			params = append(params, spd.Params[i].Default)
		}
	}

	//确保参数正确
	for i := 0; i < len(params); i++ {
		if params[i] < spd.Params[i].Min || params[i] > spd.Params[i].Max {
			params[i] = spd.Params[i].Default
		}
	}

	return true, params
}

func (spd *ScenePolicyData) OnStart(s *Scene) {

}

func (spd *ScenePolicyData) ReturnRoomCard(s *Scene, roomFeeParam, needCardCnt, playerNum int32) {
}

// 场景关闭事件
func (spd *ScenePolicyData) OnStop(s *Scene) {
}

// 场景心跳事件
func (spd *ScenePolicyData) OnTick(s *Scene) {

}

// 玩家进入事件
func (spd *ScenePolicyData) OnPlayerEnter(s *Scene, p *Player) {

}

// 玩家离开事件
func (spd *ScenePolicyData) OnPlayerLeave(s *Scene, p *Player) {

}

// 系统维护关闭事件
func (spd *ScenePolicyData) OnShutdown(s *Scene) {
	if s.IsPrivateScene() {
		PrivateSceneMgrSington.OnDestroyScene(s)
	}
}

// 获得场景的匹配因子(值越大越优先选择)
func (spd *ScenePolicyData) GetFitFactor(s *Scene, p *Player) int {
	return len(s.players)
}

// 是否满座了
func (spd *ScenePolicyData) IsFull(s *Scene, p *Player, num int32) bool {
	if s.HasPlayer(p) {
		return false
	}
	return s.GetPlayerCnt() >= int(num)
}

// 是否可以强制开始
func (spd *ScenePolicyData) IsCanForceStart(s *Scene) bool {
	return spd.CanForceStart && s.GetPlayerCnt() >= int(spd.MinPlayerCnt)
}

// 结算房卡
func (spd *ScenePolicyData) BilledRoomCard(s *Scene, snid []int32) {
	//spd.CostRoomCard(s, spd.GetRoomFeeMode(s), spd.GetNeedRoomCardCnt(s), int32(len(s.players)), snid)
}

func (spd *ScenePolicyData) getNeedRoomCardCnt(params []int32) int32 {
	return 0
}

// 需求房卡数量
func (spd *ScenePolicyData) GetNeedRoomCardCnt(s *Scene) int32 {
	return 0
}

func (spd *ScenePolicyData) getRoomFeeMode(params []int32) int32 {
	//if len(params) > 0 {
	//	param := spd.GetParamByIndex(SPDPCustomIndex_RoomFeeMode)
	//	if param != nil && param.index >= 0 && param.index < len(params) {
	//		return params[param.index]
	//	}
	//}
	return common.RoomFee_Owner
}

// 收费方式(AA|房主付费)
func (spd *ScenePolicyData) GetRoomFeeMode(s *Scene) int32 {
	//if s != nil {
	//	return spd.getRoomFeeMode(s.params)
	//}
	return common.RoomFee_Owner
}

func (spd *ScenePolicyData) GetNeedRoomCardCntDependentPlayerCnt(s *Scene) int32 {
	return 0
}

// 能否进入
func (spd *ScenePolicyData) CanEnter(s *Scene, p *Player) int {
	param := spd.GetParamByIndex(SPDPCustomIndex_SameIPForbid)
	if param != nil && len(s.params) <= param.index {
		logger.Logger.Errorf("game param len too long %v", s.gameId)
	}

	if param != nil && len(s.params) > param.index && s.params[param.index] != 0 {
		ip := p.GetIP()
		for i := 0; i < s.playerNum; i++ {
			pp := s.seats[i]
			if pp != nil && pp.GetIP() == ip {
				return int(hall_proto.OpResultCode_Game_OPRC_SameIpForbid_Game)
			}
		}
	}

	if !spd.EnterAfterStart {
		if s.starting {
			return int(hall_proto.OpResultCode_Game_OPRC_GameStarting_Game)
		}
	}

	if spd.EnterAfterStart {
		if s.starting {
			param := spd.GetParamByIndex(SPDPCustomIndex_DoorOption)
			if param != nil && s.params[param.index] != 0 {
				return int(hall_proto.OpResultCode_Game_OPRC_GameStarting_Game)
				//} else {
				//	return int(hall_proto.OpResultCode_OPRC_SceneEnterForWatcher)
			}
		}
	}
	return 0
}

// 人数
func (spd *ScenePolicyData) getPlayerNum(params []int32) int32 {
	if len(params) > 0 {
		param := spd.GetParamByIndex(SPDPCustomIndex_PlayerNum)
		if param != nil {
			idx := int(params[param.index])
			if idx >= 0 && idx < len(param.Value) {
				val := param.Value[idx]
				return val
			}
		}
	}
	return spd.DefaultPlayerCnt
}

func (spd *ScenePolicyData) GetPlayerNum(s *Scene) int32 {
	if s != nil {
		return spd.getPlayerNum(s.params)
	}
	return spd.DefaultPlayerCnt
}

// 局数
func (spd *ScenePolicyData) GetTotalOfGames(s *Scene) int32 {
	//if len(s.params) > 0 {
	//	param := spd.GetParamByIndex(SPDPCustomIndex_GamesOfCard)
	//	if param != nil {
	//		cardCostIdx := int(s.params[param.index])
	//		if cardCostIdx >= 0 && cardCostIdx < len(param.Value) {
	//			costCardNum := param.Value[cardCostIdx]
	//			return costCardNum
	//		} else if int32(cardCostIdx) >= common.CUSTOM_PER_GAME_INDEX_BEG { //自定义局数
	//			totalOfGames := int32(cardCostIdx) - common.CUSTOM_PER_GAME_INDEX_BEG + 1
	//			return totalOfGames
	//		}
	//	}
	//}
	//return 4
	return s.totalRound
}
func (spd *ScenePolicyData) GetBetState() int32 {
	return spd.BetState
}
func (spd *ScenePolicyData) GetViewLogLen() int32 {
	return spd.ViewLogCnt
}
