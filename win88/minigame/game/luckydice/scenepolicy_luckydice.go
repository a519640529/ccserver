package luckydice

import (
	proto_server "games.yol.com/win88/protocol/server"
	"time"

	"games.yol.com/win88/common"
	"games.yol.com/win88/gamerule/luckydice"
	"games.yol.com/win88/minigame/base"
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	proto_luckydice "games.yol.com/win88/protocol/luckydice"
	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/task"
)

////////////////////////////////////////////////////////////////////////////////
//幸运骰子
////////////////////////////////////////////////////////////////////////////////

// 房间内主要逻辑
var ScenePolicyLuckyDiceSington = &ScenePolicyLuckyDice{}

type ScenePolicyLuckyDice struct {
	base.BaseScenePolicy
	states [LuckyDiceSceneStateMax]base.SceneState
}

//创建场景扩展数据
func (this *ScenePolicyLuckyDice) CreateSceneExData(s *base.Scene) interface{} {
	sceneEx := NewLuckyDiceSceneData(s)
	if sceneEx != nil {
		if sceneEx.init() {
			s.ExtraData = sceneEx
		}
	}
	return sceneEx
}

//创建玩家扩展数据
func (this *ScenePolicyLuckyDice) CreatePlayerExData(s *base.Scene, p *base.Player) interface{} {
	playerEx := &LuckyDicePlayerData{Player: p}
	if playerEx != nil {
		p.ExtraData = playerEx
	}
	return playerEx
}

//场景开启事件
func (this *ScenePolicyLuckyDice) OnStart(s *base.Scene) {
	logger.Trace("(this *ScenePolicyLuckyDice) OnStart, sceneId=", s.SceneId)
	sceneEx := NewLuckyDiceSceneData(s)
	if sceneEx != nil {
		if sceneEx.init() {
			s.ExtraData = sceneEx
			s.ChangeSceneState(LuckyDiceSceneStatePrepareNewRound)
		}
	}
}

//场景关闭事件
func (this *ScenePolicyLuckyDice) OnStop(s *base.Scene) {
	logger.Trace("(this *ScenePolicyLuckyDice) OnStop , sceneId=", s.SceneId)
	if sceneEx, ok := s.ExtraData.(*LuckyDiceSceneData); ok {
		sceneEx.SaveData(true)
	}
}

//场景心跳事件
func (this *ScenePolicyLuckyDice) OnTick(s *base.Scene) {
	if s == nil {
		return
	}
	if s.SceneState != nil {
		s.SceneState.OnTick(s)
	}
}

//玩家进入事件
func (this *ScenePolicyLuckyDice) OnPlayerEnter(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Trace("(this *ScenePolicyLuckyDice) OnPlayerEnter, sceneId=", s.SceneId, " player=", p.SnId, s.GameId)
	if sceneEx, ok := s.ExtraData.(*LuckyDiceSceneData); ok {
		playerEx := sceneEx.players[p.SnId]
		if playerEx != nil {
			playerEx.isLeave = false
		} else {
			playerEx = &LuckyDicePlayerData{Player: p}
			playerEx.init(s) // 玩家当前信息初始化
			sceneEx.players[p.SnId] = playerEx
		}
		p.ExtraData = playerEx
		LuckyDiceSendRoomInfo(s, p, sceneEx, playerEx)
		s.FirePlayerEvent(p, base.PlayerEventEnter, nil) //回调会调取 onPlayerEvent事件
	}
}

//玩家离开事件
func (this *ScenePolicyLuckyDice) OnPlayerLeave(s *base.Scene, p *base.Player, reason int) {
	if s == nil || p == nil {
		return
	}
	logger.Trace("(this *ScenePolicyLuckyDice) OnPlayerLeave, sceneId=", s.SceneId, " player=", p.SnId)
	if sceneEx, ok := s.ExtraData.(*LuckyDiceSceneData); ok {
		if this.CanChangeCoinScene(s, p) {
			s.FirePlayerEvent(p, base.PlayerEventLeave, []int64{int64(reason)})
			sceneEx.OnPlayerLeave(p, reason)
		}
	}
}

//玩家掉线
func (this *ScenePolicyLuckyDice) OnPlayerDropLine(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Trace("(this *ScenePolicyLuckyDice) OnPlayerDropLine, sceneId=", s.SceneId, " player=", p.SnId)
	s.FirePlayerEvent(p, base.PlayerEventDropLine, nil)
}

//玩家重连
func (this *ScenePolicyLuckyDice) OnPlayerRehold(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Trace("(this *ScenePolicyLuckyDice) OnPlayerRehold, sceneId=", s.SceneId, " player=", p.Name)
	if _, ok := s.ExtraData.(*LuckyDiceSceneData); ok {
		if playerEx, ok := p.ExtraData.(*LuckyDicePlayerData); ok {
			playerEx.isLeave = false
			//发送房间信息给自己
			//gs 需要 不再发送 房间信息给客户端 return的时候发送
			//LuckyDiceSendRoomInfo(s, p, sceneEx, playerEx)
			s.FirePlayerEvent(p, base.PlayerEventRehold, nil)
		}
	}
}

//玩家返回房间
func (this *ScenePolicyLuckyDice) OnPlayerReturn(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Trace("(this *ScenePolicyLuckyDice) OnPlayerReturn,sceneId=", s.SceneId, " Player= ", p.Name)
	if sceneEx, ok := s.ExtraData.(*LuckyDiceSceneData); ok {
		if playerEx, ok := p.ExtraData.(*LuckyDicePlayerData); ok {
			LuckyDiceSendRoomInfo(s, p, sceneEx, playerEx)
			s.FirePlayerEvent(p, base.PlayerReturn, nil)
		}
	}
}

//玩家操作
func (this *ScenePolicyLuckyDice) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if s == nil || p == nil {
		return false
	}
	if s.SceneState != nil {
		return s.SceneState.OnPlayerOp(s, p, opcode, params)
	}
	return true
}

//玩家事件
func (this *ScenePolicyLuckyDice) OnPlayerEvent(s *base.Scene, p *base.Player, evtcode int, params []int64) {
	if s == nil || p == nil {
		return
	}
	if s.SceneState != nil {
		s.SceneState.OnPlayerEvent(s, p, evtcode, params)
	}
}

//是否完成了整个牌局
func (this *ScenePolicyLuckyDice) IsCompleted(s *base.Scene) bool { return false }

//是否可以强制开始
func (this *ScenePolicyLuckyDice) IsCanForceStart(s *base.Scene) bool { return true }

//当前状态能否换桌
func (this *ScenePolicyLuckyDice) CanChangeCoinScene(s *base.Scene, p *base.Player) bool {
	if s == nil || p == nil {
		return true
	}
	if s.SceneState != nil {
		return s.SceneState.CanChangeScene(s, p)
	}
	return true
}

func (this *ScenePolicyLuckyDice) RegisteSceneState(state base.SceneState) {
	if state == nil {
		return
	}
	stateid := state.GetState()
	if stateid < 0 || stateid >= LuckyDiceSceneStateMax {
		return
	}
	this.states[stateid] = state
}

func (this *ScenePolicyLuckyDice) GetSceneState(s *base.Scene, stateid int) base.SceneState {
	if stateid >= 0 && stateid < LuckyDiceSceneStateMax {
		return ScenePolicyLuckyDiceSington.states[stateid]
	}
	return nil
}

func LuckyDiceSendRoomInfo(s *base.Scene, p *base.Player, sceneEx *LuckyDiceSceneData, playerEx *LuckyDicePlayerData) {
	logger.Trace("-------------------发送房间消息 ", s.SceneId, p.SnId)
	pack := &proto_luckydice.SCLuckyDiceRoomInfo{
		RoomId:       proto.Int(s.SceneId),
		GameId:       proto.Int(s.GameId),
		RoomMode:     proto.Int(s.GameMode),
		Params:       s.GetParams(),
		State:        proto.Int(s.SceneState.GetState()),
		TimeOut:      proto.Int(s.SceneState.GetTimeout(s)),
		StateTimes:   LuckyDiceSceneStateTimeout,
		TotalBet:     []int64{sceneEx.totalBetBig, sceneEx.totalBetSmall},
		TotalPlayer:  []int32{int32(len(sceneEx.betBigPlayers)), int32(len(sceneEx.betSmlPlayers))},
		RoundHistory: sceneEx.roundHistory,
		RoundId:      proto.Int32(sceneEx.roundID),
		GameFreeId:   proto.Int32(s.GetDBGameFree().GetId()),
	}
	if sceneEx.dices == nil {
		pack.Dices = []int32{0, 0, 0}
	} else {
		pack.Dices = []int32{sceneEx.dices.Dice1, sceneEx.dices.Dice2, sceneEx.dices.Dice3}
	}
	if playerEx != nil {
		pd := &proto_luckydice.LuckyDicePlayerData{
			Name:        proto.String(playerEx.Name),
			SnId:        proto.Int32(playerEx.SnId),
			Head:        proto.Int32(playerEx.Head),
			Sex:         proto.Int32(playerEx.Sex),
			Coin:        proto.Int64(playerEx.Coin),
			HeadOutLine: proto.Int32(playerEx.HeadOutLine),
			VIP:         proto.Int32(playerEx.VIP),
			Bet:         proto.Int64(playerEx.bet),
			BetSide:     proto.Int32(playerEx.betSide),
			Award:       proto.Int64(playerEx.award),
		}
		pack.Players = append(pack.Players, pd)
	}
	proto.SetDefaults(pack)
	p.SendToClient(int(proto_luckydice.LuckyDicePacketID_PACKET_SC_LUCKYDICE_ROOMINFO), pack)
	logger.Tracef("-------------------发送房间消息 LuckyDice %v %v %v", s.SceneId, p.SnId, pack) //todo del
}
func SendRoomState(s *base.Scene, params []int32, state int) {
	pack := &proto_luckydice.SCLuckyDiceRoomState{
		State: proto.Int(state),
	}
	for _, v := range params {
		pack.Params = append(pack.Params, v)
	}
	proto.SetDefaults(pack)
	s.Broadcast(int(proto_luckydice.LuckyDicePacketID_PACKET_SC_LUCKYDICE_ROOMSTATE), pack, 0)
}

type SceneStateLuckyDiceBase struct {
}

func (this *SceneStateLuckyDiceBase) GetState() int { return -1 }

func (this *SceneStateLuckyDiceBase) CanChangeTo(s base.SceneState) bool {
	return true
}
func (this *SceneStateLuckyDiceBase) CanChangeCoinScene(s *base.Scene, p *base.Player) bool {
	return true
}
func (this *SceneStateLuckyDiceBase) OnEnter(s *base.Scene) {
	if sceneEx, ok := s.ExtraData.(*LuckyDiceSceneData); ok {
		sceneEx.StateStartTime = time.Now()
	}
}
func (this *SceneStateLuckyDiceBase) OnLeave(s *base.Scene) {}
func (this *SceneStateLuckyDiceBase) GetTimeout(s *base.Scene) int {
	if sceneEx, ok := s.ExtraData.(*LuckyDiceSceneData); ok {
		return int(time.Now().Sub(sceneEx.StateStartTime) / time.Second)
	}
	return 0
}
func (this *SceneStateLuckyDiceBase) OnPlayerEvent(s *base.Scene, p *base.Player, evtcode int, params []int64) {
}
func (this *SceneStateLuckyDiceBase) OnTick(s *base.Scene) {
	if sceneEx, ok := s.ExtraData.(*LuckyDiceSceneData); ok {
		// 检查玩家是否满足被踢出条件
		for _, p := range sceneEx.players {
			if time.Now().Sub(p.LastOPTimer) > time.Minute*3 { //3分钟内未操作提出
				sceneEx.PlayerLeave(p.Player, common.PlayerLeaveReason_LongTimeNoOp, true)
				continue
			}

			//游戏次数达到目标值踢出
			todayGamefreeIDSceneData, _ := p.GetDaliyGameData(int(sceneEx.DbGameFree.GetId()))
			if !p.IsRob &&
				todayGamefreeIDSceneData != nil &&
				sceneEx.DbGameFree.GetPlayNumLimit() != 0 &&
				todayGamefreeIDSceneData.GameTimes >= int64(sceneEx.DbGameFree.GetPlayNumLimit()) {
				s.PlayerLeave(p.Player, common.PlayerLeaveReason_GameTimes, true)
			}
		}

		// 检查服务器是否需要关闭
		if sceneEx.CheckNeedDestroy() {
			for _, player := range sceneEx.players {
				if !player.IsRob {
					sceneEx.PlayerLeave(player.Player, common.PlayerLeaveReason_OnDestroy, true)
				}
			}
			if s.GetRealPlayerCnt() == 0 {
				sceneEx.SceneDestroy(true)
			}
		}
	}
}
func (this *SceneStateLuckyDiceBase) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	playerEx, ok := p.ExtraData.(*LuckyDicePlayerData)
	if !ok {
		return false
	}
	sceneEx, ok := s.ExtraData.(*LuckyDiceSceneData)
	if !ok {
		return false
	}
	if sceneEx.CheckNeedDestroy() {
		//离开有统计
		sceneEx.PlayerLeave(playerEx.Player, common.PlayerLeaveReason_OnDestroy, true)
		return false
	}

	switch opcode {
	case LuckyDicePlayerHistory:
		task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
			gpl := model.GetPlayerListByHallEx(p.SnId, p.Platform, 0, 50, 0, 0, s.DbGameFree.SceneType, s.DbGameFree.GetGameClass(), s.GameId)
			pack := &proto_luckydice.SCLuckyDicePlayerHistory{}
			for _, v := range gpl.Data {
				if v.GameDetailedLogId == "" {
					logger.Error("LuckyDicePlayerHistory GameDetailedLogId is nil")
					break
				}
				gdl := model.GetPlayerHistory(p.Platform, v.GameDetailedLogId)
				if gdl == nil || gdl.GameDetailedNote == "" {
					logger.Logger.Errorf("LuckyDicePlayerHistory %v gdl is nil", v.GameDetailedLogId)
					continue
				}
				data, err := model.UnMarshalLuckyDiceGameNote(gdl.GameDetailedNote)
				if err != nil {
					logger.Errorf("UnMarshalLuckyDiceGameNote error:%v", err)
				}
				if gnd, ok := data.(*model.LuckyDiceType); ok {
					player := &proto_luckydice.LuckyDicePlayerHistoryInfo{
						RoundId:     proto.Int32(gnd.RoundId),
						CreatedTime: proto.Int64(int64(v.Ts)),
						Dices:       gnd.Dices,
						BetSide:     proto.Int32(gnd.BetSide),
						Bet:         proto.Int64(gnd.Bet),
						Refund:      proto.Int64(gnd.Refund),
						Award:       proto.Int64(gnd.Award),
					}
					pack.PlayerHistory = append(pack.PlayerHistory, player)
				}
			}
			proto.SetDefaults(pack)
			logger.Info("LuckyDicePlayerHistory: ", pack)
			return pack

		}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
			if data == nil {
				logger.Error("LuckyDicePlayerHistory data is nil")
				return
			}
			p.SendToClient(int(proto_luckydice.LuckyDicePacketID_PACKET_SC_LUCKYDICE_PLAYERHISTORY), data)
		}), "CSGetLuckyDicePlayerHistoryHandler").Start()

	case LuckyDiceRoundBetData:
		if len(params) != 1 {
			this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, proto_luckydice.OpResultCode_OPRC_Error, params)
			return true
		}
		pack := sceneEx.roundBetHistory[int32(params[0])]
		//if pack == nil {
		//	this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, proto_luckydice.OpResultCode_OPRC_Error, params)
		//	return true
		//}
		proto.SetDefaults(pack)
		p.SendToClient(int(proto_luckydice.LuckyDicePacketID_PACKET_SC_LUCKYDICE_ROUNDBETDATA), pack)

	case LuckyDiceDiceHistory:
		pack := &proto_luckydice.SCLuckyDiceDiceHistory{}
		for _, dices := range sceneEx.dicesHistory {
			pack.Dice1 = append(pack.Dice1, dices.Dice1)
			pack.Dice2 = append(pack.Dice2, dices.Dice2)
			pack.Dice3 = append(pack.Dice3, dices.Dice3)
		}
		proto.SetDefaults(pack)
		p.SendToClient(int(proto_luckydice.LuckyDicePacketID_PACKET_SC_LUCKYDICE_DICEHISTORY), pack)
	default:
		return false
	}

	this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, proto_luckydice.OpResultCode_OPRC_Sucess, params)
	return true
}

//发送玩家操作情况
func (this *SceneStateLuckyDiceBase) OnPlayerSToCOp(s *base.Scene, p *base.Player, pos int, opcode int,
	opRetCode proto_luckydice.OpResultCode, params []int64) {
	pack := &proto_luckydice.SCLuckyDiceOp{
		SnId:      proto.Int32(p.SnId),
		OpCode:    proto.Int(opcode),
		OpRetCode: opRetCode,
		Params:    params,
	}
	proto.SetDefaults(pack)
	p.SendToClient(int(proto_luckydice.LuckyDicePacketID_PACKET_SC_LUCKYDICE_PLAYEROP), pack)
}

type SceneStateLuckyDicePrepareNewRound struct {
	SceneStateLuckyDiceBase
}

//获取当前场景状态
func (this *SceneStateLuckyDicePrepareNewRound) GetState() int {
	return LuckyDiceSceneStatePrepareNewRound
}

//是否可以切换状态到
func (this *SceneStateLuckyDicePrepareNewRound) CanChangeTo(s base.SceneState) bool {
	if s.GetState() == LuckyDiceSceneStateBetting {
		return true
	}
	return false
}

//当前状态能否换桌
func (this *SceneStateLuckyDicePrepareNewRound) CanChangeScene(s *base.Scene, p *base.Player) bool {
	return true
}

func (this *SceneStateLuckyDicePrepareNewRound) OnEnter(s *base.Scene) {
	this.SceneStateLuckyDiceBase.OnEnter(s)
	if sceneEx, ok := s.ExtraData.(*LuckyDiceSceneData); ok {
		sceneEx.RefreshSceneData()
		s.TryDismissRob()
		for _, value := range sceneEx.players {
			if !value.IsOnLine() || value.isLeave {
				s.PlayerLeave(value.Player, common.PlayerLeaveReason_DropLine, true)
			}
		}
		SendRoomState(s, []int32{sceneEx.roundID}, this.GetState())
		//pack := &proto_luckydice.SCLuckyDiceRoomState{
		//	State:  proto.Int(this.GetState()),
		//	Params: []int32{sceneEx.roundID},
		//}
		//proto.SetDefaults(pack)
		//s.Broadcast(int(proto_luckydice.LuckyDicePacketID_PACKET_SC_LUCKYDICE_ROOMSTATE), pack, 0)
		logger.Trace("SceneStateLuckyDice OnEnter PrepareNewRound: ", sceneEx.roundID)
	}
}

func (this *SceneStateLuckyDicePrepareNewRound) OnTick(s *base.Scene) {
	this.SceneStateLuckyDiceBase.OnTick(s)
	if len(s.Players) == 0 { // todo
		return
	}
	if this.GetTimeout(s) > int(LuckyDiceSceneStateTimeout[this.GetState()]) {
		s.ChangeSceneState(LuckyDiceSceneStateBetting)
	}
}

func (this *SceneStateLuckyDicePrepareNewRound) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if this.SceneStateLuckyDiceBase.OnPlayerOp(s, p, opcode, params) {
		return true
	}
	this.OnPlayerSToCOp(s, p, 0, opcode, proto_luckydice.OpResultCode_OPRC_Error, params)
	return false
}

func (this *SceneStateLuckyDicePrepareNewRound) OnPlayerEvent(s *base.Scene, p *base.Player, evtcode int, params []int64) {
}

type SceneStateLuckyDiceBetting struct {
	SceneStateLuckyDiceBase
}

//获取当前场景状态
func (this *SceneStateLuckyDiceBetting) GetState() int { return LuckyDiceSceneStateBetting }

//是否可以切换状态到
func (this *SceneStateLuckyDiceBetting) CanChangeTo(s base.SceneState) bool {
	if s.GetState() == LuckyDiceSceneStateEndBetting {
		return true
	}
	return false
}

//当前状态能否换桌
func (this *SceneStateLuckyDiceBetting) CanChangeScene(s *base.Scene, p *base.Player) bool {
	return true
}

func (this *SceneStateLuckyDiceBetting) OnEnter(s *base.Scene) {
	this.SceneStateLuckyDiceBase.OnEnter(s)
	if sceneEx, ok := s.ExtraData.(*LuckyDiceSceneData); ok {
		sceneEx.GameNowTime = time.Now()
		sceneEx.NumOfGames++
		//pack := &proto_luckydice.SCLuckyDiceRoomState{
		//	State: proto.Int(this.GetState()),
		//}
		//proto.SetDefaults(pack)
		//s.Broadcast(int(proto_luckydice.LuckyDicePacketID_PACKET_SC_LUCKYDICE_ROOMSTATE), pack, 0)
		SendRoomState(s, []int32{}, this.GetState())
		logger.Trace("SceneStateLuckyDice OnEnter Betting: ")
		sceneEx.BroadcastBetChange()
	}
}

func (this *SceneStateLuckyDiceBetting) OnTick(s *base.Scene) {
	this.SceneStateLuckyDiceBase.OnTick(s)
	if this.GetTimeout(s) > int(LuckyDiceSceneStateTimeout[this.GetState()]) {
		s.ChangeSceneState(LuckyDiceSceneStateEndBetting)
	}
}

func (this *SceneStateLuckyDiceBetting) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if this.SceneStateLuckyDiceBase.OnPlayerOp(s, p, opcode, params) {
		return true
	}
	playerEx, ok := p.ExtraData.(*LuckyDicePlayerData)
	if !ok {
		return false
	}
	sceneEx, ok := s.ExtraData.(*LuckyDiceSceneData)
	if !ok {
		return false
	}

	switch opcode {
	case LuckyDicePlayerOpBet: //投注
		//params 参数0 投注金额; 参数1 押大押小
		//参数是否合法
		if !p.IsRob && (len(params) != 2 || (params[1] != int64(luckydice.Big) && params[1] != int64(luckydice.Small))) {
			//!common.InSliceInt32(sceneEx.DbGameFree.GetOtherIntParams(), int32(params[0])) ||||this.GetTimeout(s) < 5) { // 最后5秒不能再押注
			this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, proto_luckydice.OpResultCode_OPRC_Error, params)
			return false
		}
		//校验玩家余额是否足够
		if !p.IsRob && params[0] > playerEx.Coin {
			this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, proto_luckydice.OpResultCode_OPRC_CoinNotEnough, params)
			return false
		}
		if p.IsRob {
			params = make([]int64, 2)
			params[0] = luckydice.GetBotBetValue(int64(sceneEx.DbGameFree.GetBaseScore()))
			params[1] = int64(luckydice.GetBotBetSide())
		}
		// 不能同时押大又押小
		if playerEx.betSide >= 0 && playerEx.betSide != int32(params[1]) {
			this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, proto_luckydice.OpResultCode_OPRC_Error, params)
			return false
		}

		realBet := func() {
			p.LastOPTimer = time.Now()
			if playerEx.doOncePerRound {
				playerEx.doOncePerRound = false
				p.GameTimes++
				//playerEx.StartCoin = playerEx.Coin
				playerEx.betSide = int32(params[1])
			}
			playerEx.bet += params[0]
			sceneEx.betChanged = true
			betInfo := &LuckyDiceBetInfo{
				AccountID: playerEx.SnId,
				BetSide:   playerEx.betSide,
				Bet:       params[0],
				IsBotBet:  playerEx.IsRob,
				BetTime:   time.Now().Unix(),
			}
			sceneEx.allBets.PushBack(betInfo)
			if playerEx.betSide == luckydice.Big {
				sceneEx.betBigPlayers[playerEx.SnId] = playerEx
				sceneEx.totalBetBig += params[0]
				sceneEx.bigBets.PushBack(betInfo)
			} else {
				sceneEx.betSmlPlayers[playerEx.SnId] = playerEx
				sceneEx.totalBetSmall += params[0]
				sceneEx.smlBets.PushBack(betInfo)
			}
		}

		if !p.IsRob && !sceneEx.Testing {
			p.AddCoin(-params[0], common.GainWay_HundredSceneLost, false, false, "system", s.GetSceneName(), base.PlayerAddCoinDefRetryCnt, func(context *base.AddCoinContext, coin int64, ok bool) {
				if !ok {
					return
				}

				p.Coin = coin
				//p.Coin += context.Coin
				realBet()
			})
		} else {
			realBet()
		}

	default:
		this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, proto_luckydice.OpResultCode_OPRC_Error, params)
		return false
	}
	return true
}

func (this *SceneStateLuckyDiceBetting) OnPlayerEvent(s *base.Scene, p *base.Player, evtcode int, params []int64) {
}

type SceneStateLuckyDiceEndBetting struct {
	SceneStateLuckyDiceBase
}

//获取当前场景状态
func (this *SceneStateLuckyDiceEndBetting) GetState() int { return LuckyDiceSceneStateEndBetting }

//是否可以切换状态到
func (this *SceneStateLuckyDiceEndBetting) CanChangeTo(s base.SceneState) bool {
	if s.GetState() == LuckyDiceSceneStateShowResult {
		return true
	}
	return false
}

//当前状态能否换桌
func (this *SceneStateLuckyDiceEndBetting) CanChangeScene(s *base.Scene, p *base.Player) bool {
	return true
}

func (this *SceneStateLuckyDiceEndBetting) OnEnter(s *base.Scene) {
	this.SceneStateLuckyDiceBase.OnEnter(s)
	if sceneEx, ok := s.ExtraData.(*LuckyDiceSceneData); ok {
		//pack := &proto_luckydice.SCLuckyDiceRoomState{
		//	State: proto.Int(this.GetState()),
		//}
		//proto.SetDefaults(pack)
		//s.Broadcast(int(proto_luckydice.LuckyDicePacketID_PACKET_SC_LUCKYDICE_ROOMSTATE), pack, 0)
		SendRoomState(s, []int32{}, this.GetState())
		logger.Tracef("SceneStateLuckyDice OnEnter EndBetting: %v %v %v/%v %v/%v", sceneEx.roundID, len(sceneEx.players),
			sceneEx.totalBetBig, sceneEx.totalBetSmall, len(sceneEx.betBigPlayers), len(sceneEx.betSmlPlayers))
	}
}

func (this *SceneStateLuckyDiceEndBetting) OnTick(s *base.Scene) {
	this.SceneStateLuckyDiceBase.OnTick(s)
	if this.GetTimeout(s) > int(LuckyDiceSceneStateTimeout[this.GetState()]) {
		this.PrepareResult(s)
		s.ChangeSceneState(LuckyDiceSceneStateShowResult)
	}
}

func (this *SceneStateLuckyDiceEndBetting) PrepareResult(s *base.Scene) {
	sceneEx, ok := s.ExtraData.(*LuckyDiceSceneData)
	if !ok {
		return
	}

	if !sceneEx.bIntervention {
		sceneEx.dices = luckydice.GetDiceResult()
	}

	if len(sceneEx.players) == 0 {
		return
	}

	// 投注处理，双方投注总额须一致，多出部分返还
	diff := sceneEx.totalBetBig - sceneEx.totalBetSmall
	if diff < 0 {
		diff = -diff
	}
	var acceptBigBetBot, acceptBigBetPlayer int64
	for e := sceneEx.bigBets.Back(); e != nil; e = e.Prev() {
		betInfo, ok := e.Value.(*LuckyDiceBetInfo)
		if !ok || betInfo == nil {
			logger.Warn("LuckyDiceBetInfo Big List error")
			break
		}
		if sceneEx.totalBetBig > sceneEx.totalBetSmall && betInfo.Bet < diff {
			if sceneEx.betBigPlayers[betInfo.AccountID] != nil {
				sceneEx.betBigPlayers[betInfo.AccountID].refund += betInfo.Bet
			}
			diff -= betInfo.Bet
			betInfo.Bet = 0
		} else if sceneEx.totalBetBig > sceneEx.totalBetSmall && 0 < diff {
			if sceneEx.betBigPlayers[betInfo.AccountID] != nil {
				sceneEx.betBigPlayers[betInfo.AccountID].refund += diff
			}
			betInfo.Bet -= diff
			diff = 0
		}
		if betInfo.IsBotBet {
			acceptBigBetBot += betInfo.Bet
		} else {
			acceptBigBetPlayer += betInfo.Bet
		}
	}

	var acceptSmlBetBot, acceptSmlBetPlayer int64
	for e := sceneEx.smlBets.Back(); e != nil; e = e.Prev() {
		betInfo, ok := e.Value.(*LuckyDiceBetInfo)
		if !ok || betInfo == nil {
			logger.Warn("LuckyDiceBetInfo Sml List error")
			break
		}
		if sceneEx.totalBetBig < sceneEx.totalBetSmall && betInfo.Bet < diff {
			if sceneEx.betSmlPlayers[betInfo.AccountID] != nil {
				sceneEx.betSmlPlayers[betInfo.AccountID].refund += betInfo.Bet
			}
			diff -= betInfo.Bet
			betInfo.Bet = 0
		} else if sceneEx.totalBetBig < sceneEx.totalBetSmall && 0 < diff {
			if sceneEx.betSmlPlayers[betInfo.AccountID] != nil {
				sceneEx.betSmlPlayers[betInfo.AccountID].refund += diff
			}
			betInfo.Bet -= diff
			diff = 0
		}
		if betInfo.IsBotBet {
			acceptSmlBetBot += betInfo.Bet
		} else {
			acceptSmlBetPlayer += betInfo.Bet
		}
	}

	//根据水池调整骰子值
	coinPoolSetting := base.CoinPoolMgr.GetCoinPoolSetting(sceneEx.Platform, sceneEx.GetGameFreeId(), sceneEx.GroupId)
	gamePoolCoin := base.CoinPoolMgr.LoadCoin(sceneEx.GetGameFreeId(), sceneEx.Platform, sceneEx.GroupId) // 当前水池金额
	if !sceneEx.bIntervention && sceneEx.dices.BigSmall() == luckydice.Big && acceptBigBetPlayer > acceptSmlBetPlayer {
		if gamePoolCoin-int64(coinPoolSetting.GetLowerLimit())-(acceptSmlBetBot-acceptBigBetBot) < (acceptBigBetPlayer - acceptSmlBetPlayer) {
			for sceneEx.dices.BigSmall() == luckydice.Big {
				sceneEx.dices.Update()
			}
		}
	} else if !sceneEx.bIntervention && sceneEx.dices.BigSmall() == luckydice.Small && acceptBigBetPlayer < acceptSmlBetPlayer {
		if gamePoolCoin-int64(coinPoolSetting.GetLowerLimit())-(acceptBigBetBot-acceptSmlBetBot) < (acceptSmlBetPlayer - acceptBigBetPlayer) {
			for sceneEx.dices.BigSmall() == luckydice.Small {
				sceneEx.dices.Update()
			}
		}
	}
	if sceneEx.bIntervention {
		sceneEx.bIntervention = false
	}

	//获取当前水池的上下文环境
	sceneEx.CpCtx = base.CoinPoolMgr.GetCoinPoolCtx(sceneEx.Platform, sceneEx.GetGameFreeId(), sceneEx.GroupId)

	//水池更新
	ctroRate := coinPoolSetting.GetCtroRate() //调节赔率 暗税系数
	if ctroRate < 0 || ctroRate > 5000 {
		logger.Warnf("LuckyDiceErrorBaseRate [%v][%v]", sceneEx.GetGameFreeId(), ctroRate)
		ctroRate = 600
	}
	var playerMoneyChange int64
	if sceneEx.dices.BigSmall() == luckydice.Big {
		playerMoneyChange = acceptBigBetPlayer - acceptSmlBetPlayer
	} else {
		playerMoneyChange = acceptSmlBetPlayer - acceptBigBetPlayer
	}
	if playerMoneyChange > 0 {
		base.CoinPoolMgr.PushCoin(sceneEx.GetGameFreeId(), sceneEx.GroupId, sceneEx.Platform, int64(float64(playerMoneyChange)*float64(10000-ctroRate)/10000.0+0.000001))
	} else if playerMoneyChange < 0 {
		base.CoinPoolMgr.PopCoin(sceneEx.GetGameFreeId(), sceneEx.GroupId, sceneEx.Platform, -playerMoneyChange)
	}

	//赢家税收系数
	taxRate := sceneEx.DbGameFree.GetTaxRate()
	if taxRate < 0 || taxRate > 10000 {
		logger.Warnf("LuckyDiceErrorTaxRate [%v][%v]", sceneEx.GetGameFreeId(), taxRate)
		taxRate = 100
	}

	betFlag := make(map[int32]bool)
	roundBigBets := make([]*proto_luckydice.LuckyDiceRoundPlayerBet, 0)
	roundSmlBets := make([]*proto_luckydice.LuckyDiceRoundPlayerBet, 0)
	sceneEndPlayers := make([]*proto_server.PlayerCtx, 0)
	for e := sceneEx.allBets.Back(); e != nil; e = e.Prev() {
		betInfo, ok := e.Value.(*LuckyDiceBetInfo)
		if !ok || betInfo == nil {
			logger.Warn("LuckyDiceBetInfo All List error")
			break
		}
		if betFlag[betInfo.AccountID] {
			continue
		}
		betFlag[betInfo.AccountID] = true

		player := sceneEx.players[betInfo.AccountID]
		if player == nil {
			continue
		}
		roundPlayerBet := &proto_luckydice.LuckyDiceRoundPlayerBet{
			BetTime:  proto.Int64(player.betTime),
			UserName: proto.String(player.Name),
			Bet:      proto.Int64(player.bet),
			Refund:   proto.Int64(player.refund),
		}
		if player.betSide == luckydice.Big {
			roundBigBets = append(roundBigBets, roundPlayerBet)
		} else {
			roundSmlBets = append(roundSmlBets, roundPlayerBet)
		}

		if player.IsRob || sceneEx.Testing {
			continue
		}

		if sceneEx.dices.BigSmall() == betInfo.BetSide { // 赢家算分
			accept := player.bet - player.refund
			player.award = int64((float64(accept*2)*float64(10000-taxRate))/10000.0 + 0.000001)
			player.tax = accept - player.award
		}

		if player.bet-player.refund > 0 { // 玩家投入
			player.Statics(sceneEx.KeyGameId, sceneEx.KeyGamefreeId, -(player.bet - player.refund), true)
		}

		var totalGain int64
		if player.refund > 0 { // 返回金额 todo GainWay_HundredSceneWin
			totalGain += player.refund
		}
		if player.award > 0 { // 获奖金额
			totalGain += player.award
			player.Statics(sceneEx.KeyGameId, sceneEx.KeyGamefreeId, player.award, false)
		}

		//if totalGain != 0 {
		player.AddCoin(totalGain, common.GainWay_HundredSceneWin, false, false, "system", s.GetSceneName(), base.PlayerAddCoinDefRetryCnt, func(context *base.AddCoinContext, coin int64, ok bool) {
			if ok {
				player.Coin = coin
				sceneEndPlayers = append(sceneEndPlayers, &proto_server.PlayerCtx{SnId: proto.Int32(player.SnId), Coin: proto.Int64(player.Coin)})
			}
			//player.Coin += context.Coin
		})
		//}

		//sceneEndPlayers = append(sceneEndPlayers, &proto_server.PlayerCtx{SnId: proto.Int32(player.SnId), Coin: proto.Int64(player.Coin)})
	}

	sceneEx.roundBetHistory[sceneEx.roundID] = &proto_luckydice.SCLuckyDiceRoundBetHistory{
		RoundId:      proto.Int32(sceneEx.roundID),
		Dices:        []int32{sceneEx.dices.Dice1, sceneEx.dices.Dice2, sceneEx.dices.Dice3},
		TotalBet:     []int64{sceneEx.totalBetBig, sceneEx.totalBetSmall},
		BigBetters:   roundBigBets,
		SmallBetters: roundSmlBets,
	}
}

func (this *SceneStateLuckyDiceEndBetting) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if this.SceneStateLuckyDiceBase.OnPlayerOp(s, p, opcode, params) {
		return true
	}
	this.OnPlayerSToCOp(s, p, 0, opcode, proto_luckydice.OpResultCode_OPRC_Error, params)
	return false
}

func (this *SceneStateLuckyDiceEndBetting) OnPlayerEvent(s *base.Scene, p *base.Player, evtcode int, params []int64) {
}

type SceneStateLuckyDiceShowResult struct {
	SceneStateLuckyDiceBase
}

//获取当前场景状态
func (this *SceneStateLuckyDiceShowResult) GetState() int { return LuckyDiceSceneStateShowResult }

//是否可以切换状态到
func (this *SceneStateLuckyDiceShowResult) CanChangeTo(s base.SceneState) bool {
	if s.GetState() == LuckyDiceSceneStatePrepareNewRound {
		return true
	}
	return false
}

//当前状态能否换桌
func (this *SceneStateLuckyDiceShowResult) CanChangeScene(s *base.Scene, p *base.Player) bool {
	return true
}

func (this *SceneStateLuckyDiceShowResult) OnEnter(s *base.Scene) {
	this.SceneStateLuckyDiceBase.OnEnter(s)
	sceneEx, ok := s.ExtraData.(*LuckyDiceSceneData)
	if !ok {
		return
	}

	//pack := &proto_luckydice.SCLuckyDiceRoomState{
	//	State: proto.Int(this.GetState()),
	//}
	//proto.SetDefaults(pack)
	//s.Broadcast(int(proto_luckydice.LuckyDicePacketID_PACKET_SC_LUCKYDICE_ROOMSTATE), pack, 0)
	SendRoomState(s, []int32{}, this.GetState())
	logger.Tracef("SceneStateLuckyDice OnEnter ShowResult: %v %v", sceneEx.roundID, sceneEx.dices)

	for _, player := range sceneEx.players {
		if player.IsRob || sceneEx.Testing {
			logger.Info("player.IsRob || sceneEx.Testing ")
			continue
		}

		pack := &proto_luckydice.SCLuckyDiceGameBilled{
			RoundId: proto.Int32(sceneEx.roundID),
			Dices:   []int32{sceneEx.dices.Dice1, sceneEx.dices.Dice2, sceneEx.dices.Dice3},
			Bet:     proto.Int64(player.bet),
			Refund:  proto.Int64(player.refund),
			Award:   proto.Int64(player.award),
			Balance: proto.Int64(player.Coin),
		}
		proto.SetDefaults(pack)
		logger.Infof("LuckyDicePlayerResult [%v] %v", player.SnId, pack)
		player.SendToClient(int(proto_luckydice.LuckyDicePacketID_PACKET_SC_LUCKYDICE_GAMEBILLED), pack)

		player.CurrentBet = player.bet - player.refund
		player.CurrentTax = player.tax

		this.SavePlayerLog(sceneEx, player)
	}
}

func (this *SceneStateLuckyDiceShowResult) SavePlayerLog(sceneEx *LuckyDiceSceneData, playerEx *LuckyDicePlayerData) {
	//changeCoin := playerEx.Coin - playerEx.StartCoin
	changeCoin := playerEx.award + playerEx.refund - playerEx.tax - playerEx.bet
	startCoin := playerEx.Coin - changeCoin
	if changeCoin != 0 {
		playerEx.SaveSceneCoinLog(startCoin, changeCoin,
			playerEx.Coin, playerEx.CurrentBet, playerEx.tax, changeCoin+playerEx.tax, 0, 0)
	}

	luckyDiceType := &model.LuckyDiceType{
		RoomId:     int32(sceneEx.SceneId),
		RoundId:    sceneEx.roundID,
		BaseScore:  sceneEx.DbGameFree.GetBaseScore(),
		PlayerSnid: playerEx.SnId,
		UserName:   playerEx.Name,
		BeforeCoin: startCoin,
		AfterCoin:  playerEx.Coin,
		ChangeCoin: changeCoin,
		Bet:        playerEx.bet,
		Refund:     playerEx.refund,
		Award:      playerEx.award,
		BetSide:    playerEx.betSide,
		Dices:      []int32{sceneEx.dices.Dice1, sceneEx.dices.Dice2, sceneEx.dices.Dice3},
		Tax:        playerEx.tax,
		WBLevel:    sceneEx.players[playerEx.SnId].WBLevel,
	}

	info, err := model.MarshalGameNoteByMini(luckyDiceType)
	if err == nil {
		logid, _ := model.AutoIncGameLogId()
		sceneEx.SaveGameDetailedLog(logid, info, &base.GameDetailedParam{})
		totalin := playerEx.CurrentBet
		totalout := luckyDiceType.Award + playerEx.tax
		sceneEx.SaveGamePlayerListLog(playerEx.SnId, // todo
			&base.SaveGamePlayerListLogParam{
				Platform:          playerEx.Platform,
				Channel:           playerEx.Channel,
				Promoter:          playerEx.BeUnderAgentCode,
				PackageTag:        playerEx.PackageID,
				InviterId:         playerEx.InviterId,
				LogId:             logid,
				TotalIn:           totalin,
				TotalOut:          totalout,
				TaxCoin:           playerEx.tax,
				ClubPumpCoin:      0,
				BetAmount:         totalin,
				WinAmountNoAnyTax: luckyDiceType.Award,
				IsFirstGame:       sceneEx.IsPlayerFirst(playerEx.Player),
			})
	}

	//统计输下注金币数
	if !sceneEx.Testing && !playerEx.IsRob {
		playerBet := &proto_server.PlayerBet{
			SnId: proto.Int32(playerEx.SnId),
			Bet:  proto.Int64(playerEx.CurrentBet),
			Gain: proto.Int64(luckyDiceType.ChangeCoin),
			Tax:  proto.Int64(playerEx.tax),
		}
		gwPlayerBet := &proto_server.GWPlayerBet{
			GameFreeId: proto.Int32(sceneEx.DbGameFree.GetId()),
			RobotGain:  proto.Int64(-luckyDiceType.ChangeCoin),
		}
		gwPlayerBet.PlayerBets = append(gwPlayerBet.PlayerBets, playerBet)
		proto.SetDefaults(gwPlayerBet)
		sceneEx.SendToWorld(int(proto_server.SSPacketID_PACKET_GW_PLAYERBET), gwPlayerBet)
		logger.Trace("Send msg gwPlayerBet ===>", gwPlayerBet)
	}
}

func (this *SceneStateLuckyDiceShowResult) OnTick(s *base.Scene) {
	this.SceneStateLuckyDiceBase.OnTick(s)
	// 每日最后两分钟 不开始新游戏
	if time.Now().Hour() == 23 && time.Now().Minute() >= 58 {
		return
	}
	if this.GetTimeout(s) > int(LuckyDiceSceneStateTimeout[this.GetState()]) {
		if sceneEx, ok := s.ExtraData.(*LuckyDiceSceneData); ok {
			sceneEx.HandleHistory()
		}
		s.ChangeSceneState(LuckyDiceSceneStatePrepareNewRound)
	}
}

func (this *SceneStateLuckyDiceShowResult) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if this.SceneStateLuckyDiceBase.OnPlayerOp(s, p, opcode, params) {
		return true
	}
	this.OnPlayerSToCOp(s, p, 0, opcode, proto_luckydice.OpResultCode_OPRC_Error, params)
	return false
}

func (this *SceneStateLuckyDiceShowResult) OnPlayerEvent(s *base.Scene, p *base.Player, evtcode int, params []int64) {
}

func init() {
	ScenePolicyLuckyDiceSington.RegisteSceneState(&SceneStateLuckyDiceBetting{})
	ScenePolicyLuckyDiceSington.RegisteSceneState(&SceneStateLuckyDiceShowResult{})
	ScenePolicyLuckyDiceSington.RegisteSceneState(&SceneStateLuckyDicePrepareNewRound{})
	ScenePolicyLuckyDiceSington.RegisteSceneState(&SceneStateLuckyDiceEndBetting{})
	core.RegisteHook(core.HOOK_BEFORE_START, func() error {
		base.RegisteScenePolicy(common.GameId_LuckyDice, 0, ScenePolicyLuckyDiceSington)
		return nil
	})
}
