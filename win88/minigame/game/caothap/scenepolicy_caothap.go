package caothap

import (
	"errors"
	"strconv"
	"time"

	"games.yol.com/win88/common"
	"games.yol.com/win88/gamerule/caothap"
	"games.yol.com/win88/minigame/base"
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	proto_caothap "games.yol.com/win88/protocol/caothap"
	proto_server "games.yol.com/win88/protocol/server"
	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/task"
	"github.com/idealeak/goserver/core/timer"
)

////////////////////////////////////////////////////////////////////////////////
// CaoThap
////////////////////////////////////////////////////////////////////////////////

// 房间内主要逻辑
var ScenePolicyCaoThapSington = &ScenePolicyCaoThap{}

type ScenePolicyCaoThap struct {
	base.BaseScenePolicy
	states [CaoThapSceneStateMax]base.SceneState
}

//创建场景扩展数据
func (this *ScenePolicyCaoThap) CreateSceneExData(s *base.Scene) interface{} {
	sceneEx := NewCaoThapSceneData(s)
	if sceneEx != nil {
		if sceneEx.init() {
			s.ExtraData = sceneEx
		}
	}
	return sceneEx
}

//创建玩家扩展数据
func (this *ScenePolicyCaoThap) CreatePlayerExData(s *base.Scene, p *base.Player) interface{} {
	playerEx := &CaoThapPlayerData{Player: p}
	if playerEx != nil {
		p.ExtraData = playerEx
	}
	return playerEx
}

//场景开启事件
func (this *ScenePolicyCaoThap) OnStart(s *base.Scene) {
	logger.Logger.Trace("(this *ScenePolicyCaoThap) OnStart, sceneId=", s.SceneId)
	sceneEx := NewCaoThapSceneData(s)
	if sceneEx != nil {
		if sceneEx.init() {
			//初始化奖池同步handler
			if sceneEx.jackpotNoticeHandle != timer.TimerHandle(0) {
				timer.StopTimer(sceneEx.jackpotNoticeHandle)
				sceneEx.jackpotNoticeHandle = timer.TimerHandle(0)
			}
			if hNext, ok := common.DelayInvake(func() {
				sceneEx.BroadcastJackpot()
			}, nil, 10*time.Second, -1); ok {
				sceneEx.jackpotNoticeHandle = hNext
			}
			//sceneEx.BroadcastJackpot()
			s.ExtraData = sceneEx
			s.ChangeSceneState(CaoThapSceneStateStart) //改变当前的玩家状态
		}
	}
}

//场景关闭事件
func (this *ScenePolicyCaoThap) OnStop(s *base.Scene) {
	logger.Logger.Trace("(this *ScenePolicyCaoThap) OnStop , sceneId=", s.SceneId)
	if sceneEx, ok := s.ExtraData.(*CaoThapSceneData); ok {
		if sceneEx.jackpotNoticeHandle != timer.TimerHandle(0) {
			timer.StopTimer(sceneEx.jackpotNoticeHandle)
		}
		sceneEx.SaveData(true)
	}
}

//场景心跳事件
func (this *ScenePolicyCaoThap) OnTick(s *base.Scene) {
	if s == nil {
		return
	}
	if s.SceneState != nil {
		s.SceneState.OnTick(s)
	}
}

//玩家进入事件
func (this *ScenePolicyCaoThap) OnPlayerEnter(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyCaoThap) OnPlayerEnter, sceneId=", s.SceneId, " player=", p.SnId)
	if sceneEx, ok := s.ExtraData.(*CaoThapSceneData); ok {
		playerEx := &CaoThapPlayerData{Player: p}
		playerEx.init(s) // 玩家当前信息初始化
		sceneEx.players[p.SnId] = playerEx
		p.ExtraData = playerEx
		CaoThapSendRoomInfo(s, p, sceneEx, playerEx, nil)
		s.FirePlayerEvent(p, base.PlayerEventEnter, nil) //回调会调取 onPlayerEvent事件
	}
}

//玩家离开事件
func (this *ScenePolicyCaoThap) OnPlayerLeave(s *base.Scene, p *base.Player, reason int) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyCaoThap) OnPlayerLeave, sceneId=", s.SceneId, " player=", p.SnId)
	if sceneEx, ok := s.ExtraData.(*CaoThapSceneData); ok {
		if this.CanChangeCoinScene(s, p) {
			sceneEx.OnPlayerLeave(p, reason)
			s.FirePlayerEvent(p, base.PlayerEventLeave, []int64{int64(reason)})
		}
	}
}

//玩家掉线
func (this *ScenePolicyCaoThap) OnPlayerDropLine(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyCaoThap) OnPlayerDropLine, sceneId=", s.SceneId, " player=", p.SnId)
	s.FirePlayerEvent(p, base.PlayerEventDropLine, nil)
	if sceneEx, ok := s.ExtraData.(*CaoThapSceneData); ok {
		if sceneEx.Gaming {
			return
		}
	}
}

//玩家重连
func (this *ScenePolicyCaoThap) OnPlayerRehold(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyCaoThap) OnPlayerRehold, sceneId=", s.SceneId, " player=", p.Name)
	if _, ok := s.ExtraData.(*CaoThapSceneData); ok {
		if _, ok := p.ExtraData.(*CaoThapPlayerData); ok {
			//发送房间信息给自己
			//CaoThapSendRoomInfo(s, p, sceneEx, playerEx)
			s.FirePlayerEvent(p, base.PlayerEventRehold, nil)
		}
	}
}

//玩家返回房间
func (this *ScenePolicyCaoThap) OnPlayerReturn(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyCaoThap) OnPlayerReturn,sceneId=", s.SceneId, " player=", p.Name)
	if sceneEx, ok := s.ExtraData.(*CaoThapSceneData); ok {
		if playerEx, ok := p.ExtraData.(*CaoThapPlayerData); ok {
			CaoThapSendRoomInfo(s, p, sceneEx, playerEx, playerEx.billedData)
			s.FirePlayerEvent(p, base.PlayerReturn, nil)
		}
	}
}

//玩家操作
func (this *ScenePolicyCaoThap) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if s == nil || p == nil {
		return false
	}
	if s.SceneState != nil {
		return s.SceneState.OnPlayerOp(s, p, opcode, params)
	}
	return true
}

//玩家事件
func (this *ScenePolicyCaoThap) OnPlayerEvent(s *base.Scene, p *base.Player, evtcode int, params []int64) {
	if s == nil || p == nil {
		return
	}
	if s.SceneState != nil {
		s.SceneState.OnPlayerEvent(s, p, evtcode, params)
	}
}

//是否完成了整个牌局
func (this *ScenePolicyCaoThap) IsCompleted(s *base.Scene) bool { return false }

//是否可以强制开始
func (this *ScenePolicyCaoThap) IsCanForceStart(s *base.Scene) bool { return true }

//当前状态能否换桌
func (this *ScenePolicyCaoThap) CanChangeCoinScene(s *base.Scene, p *base.Player) bool {
	if s == nil || p == nil {
		return true
	}
	if s.SceneState != nil {
		return s.SceneState.CanChangeScene(s, p)
	}
	return true
}

func (this *ScenePolicyCaoThap) RegisteSceneState(state base.SceneState) {
	if state == nil {
		return
	}
	stateid := state.GetState()
	if stateid < 0 || stateid >= CaoThapSceneStateMax {
		return
	}
	this.states[stateid] = state
}

func (this *ScenePolicyCaoThap) GetSceneState(s *base.Scene, stateid int) base.SceneState {
	if stateid >= 0 && stateid < CaoThapSceneStateMax {
		return ScenePolicyCaoThapSington.states[stateid]
	}
	return nil
}

func CaoThapSendRoomInfo(s *base.Scene, p *base.Player, sceneEx *CaoThapSceneData, playerEx *CaoThapPlayerData, data *proto_caothap.GameBilledData) {
	logger.Logger.Trace("-------------------发送房间消息 ", s.SceneId, p.SnId)
	pack := &proto_caothap.SCCaoThapRoomInfo{
		RoomId:     proto.Int(s.SceneId),
		Creator:    proto.Int32(0),
		GameId:     proto.Int(s.GameId),
		RoomMode:   proto.Int(s.GameMode),
		Params:     s.GetParams(),
		State:      proto.Int(s.SceneState.GetState()),
		Jackpot:    proto.Int64(sceneEx.jackpot.JackpotFund[playerEx.betIdx]),
		GameFreeId: proto.Int32(s.GetDBGameFree().GetId()),
		BilledData: data,
	}
	if playerEx != nil {
		pd := &proto_caothap.CaoThapPlayerData{
			SnId:        proto.Int32(playerEx.SnId),
			Name:        proto.String(playerEx.Name),
			Head:        proto.Int32(playerEx.Head),
			Sex:         proto.Int32(playerEx.Sex),
			Coin:        proto.Int64(playerEx.Coin),
			HeadOutLine: proto.Int32(playerEx.HeadOutLine),
			VIP:         proto.Int32(playerEx.VIP),
		}
		pack.Players = append(pack.Players, pd)
		//for _, value := range playerEx.cards {
		//	pack.Cards = append(pack.Cards, int32(value))
		//}
		pack.Chip = proto.Int32(playerEx.score)
	}
	playerEx.remainTime = time.Now().Unix() - playerEx.LastOPTimer.Unix() - caothap.TIMEINTERVAL
	pack.ParamsEx = append(append(pack.ParamsEx, int32(playerEx.remainTime)), playerEx.cards...)
	proto.SetDefaults(pack)
	p.SendToClient(int(proto_caothap.CaoThapPacketID_PACKET_SC_CAOTHAP_ROOMINFO), pack)
}

type SceneStateCaoThapStart struct {
}

//获取当前场景状态
func (this *SceneStateCaoThapStart) GetState() int { return CaoThapSceneStateStart }

//是否可以切换状态到
func (this *SceneStateCaoThapStart) CanChangeTo(s base.SceneState) bool { return true }

//当前状态能否换桌
func (this *SceneStateCaoThapStart) CanChangeScene(s *base.Scene, p *base.Player) bool {
	if _, ok := p.ExtraData.(*CaoThapPlayerData); ok {
		return true
	}
	return true
}

func (this *SceneStateCaoThapStart) GetTimeout(s *base.Scene) int { return 0 }

func (this *SceneStateCaoThapStart) OnEnter(s *base.Scene) {
	if sceneEx, ok := s.ExtraData.(*CaoThapSceneData); ok {
		logger.Logger.Tracef("(this *Scene) [%v] 场景状态进入 %v", s.SceneId, len(sceneEx.players))
		sceneEx.StateStartTime = time.Now()
		pack := &proto_caothap.SCCaoThapRoomState{
			State: proto.Int(this.GetState()),
		}
		proto.SetDefaults(pack)
		s.Broadcast(int(proto_caothap.CaoThapPacketID_PACKET_SC_CAOTHAP_ROOMSTATE), pack, 0)
	}
}

func (this *SceneStateCaoThapStart) OnLeave(s *base.Scene) {}

func (this *SceneStateCaoThapStart) OnTick(s *base.Scene) {
	// 检查玩家是否满足被踢出条件，检查服务器是否需要关闭
	if sceneEx, ok := s.ExtraData.(*CaoThapSceneData); ok {
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

func (this *SceneStateCaoThapStart) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	playerEx, ok := p.ExtraData.(*CaoThapPlayerData)
	if !ok {
		return false
	}
	sceneEx, ok := s.ExtraData.(*CaoThapSceneData)
	if !ok {
		return false
	}
	if sceneEx.CheckNeedDestroy() {
		//离开有统计
		sceneEx.PlayerLeave(playerEx.Player, common.PlayerLeaveReason_OnDestroy, true)
		return false
	}

	switch opcode {
	case CaoThapPlayerOpInit:
		playerEx.currentTurnID = int32(sceneEx.turnID)
		// 洗牌
		playerEx.cardsData = caothap.CardInit()
		//if playerEx.debugJackpot {
		//	copy(playerEx.cardsData, caothap.DebugCardData[:])
		//}
		card := playerEx.cardsData[playerEx.cardPos]
		// 校验step,TurnID
		if int64(playerEx.step) > caothap.STEP_FIRSTINIT || playerEx.currentTurnID > sceneEx.turnID {
			this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, proto_caothap.OpResultCode_OPRC_Error, params)
			return false
		}
		// 底注校验
		if !common.InSliceInt32(sceneEx.DbGameFree.GetOtherIntParams(), int32(params[0])) {
			this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, proto_caothap.OpResultCode_OPRC_Error, params)
			return false
		}
		playerEx.score = int32(params[0]) // 压注数
		playerEx.betIdx = int32(common.InSliceInt32Index(sceneEx.DbGameFree.GetOtherIntParams(), playerEx.score))
		playerEx.betValue = params[0] //起始下注金额
		betValue := int64(playerEx.score)
		if betValue > playerEx.Coin {
			this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, proto_caothap.OpResultCode_OPRC_CoinNotEnough, params)
			return false
		} else if int64(sceneEx.DbGameFree.GetBetLimit()) > playerEx.Coin { //押注限制
			this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, proto_caothap.OpResultCode_OPRC_CoinNotEnough, params)
			return false
		}

		// 是否是A牌
		if caothap.IsCardA(card) {
			playerEx.currentAces = append(playerEx.currentAces, card)
			playerEx.currentAcesQuantity = 1
			playerEx.createdTimePlay = time.Now()
		}

		p.LastOPTimer = time.Now()
		sceneEx.NumOfGames++
		p.GameTimes++
		//playerEx.StartCoin = playerEx.Coin

		//dbGamePool := srvdata.PBDB_GameCoinPoolMgr.GetData(sceneEx.GetGameFreeId())
		//baseRate := dbGamePool.GetBaseRate() // 基础倍率
		baseRate := int32(9200)
		playerEx.betRateUp, playerEx.betRateDown = playerEx.CalcBetRate(baseRate)

		playerEx.cards = append(playerEx.cards, card)
		playerEx.cardCount = 1

		playerEx.TotalBet += betValue //总下注额（从进房间开始,包含多局游戏的下注）
		//扣除投注金币
		p.AddCoin(-betValue, common.GainWay_HundredSceneLost, true, false, "system", s.GetSceneName(), base.PlayerAddCoinDefRetryCnt, func(context *base.AddCoinContext, coin int64, ok bool) {
			if !ok {
				this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, proto_caothap.OpResultCode_OPRC_CoinNotEnough, params)
				return
			}
			//p.Coin += context.Coin
			p.Coin = coin
			sceneEx.turnID++
			playerEx.currentTurnID = sceneEx.turnID
			//playerEx.cardPos++
			playerEx.prizeValue = betValue
			playerEx.winCoin = playerEx.prizeValue
			playerEx.taxCoin = 0
			playerEx.step++

			if playerEx.betRateUp == 0 {
				playerEx.bigWinScore = 0
			} else if playerEx.betRateUp == 1 {
				playerEx.bigWinScore = playerEx.prizeValue
			} else {
				playerEx.bigWinScore = playerEx.prizeValue + playerEx.betRateUp*betValue/100
			}
			if playerEx.betRateDown == 0 {
				playerEx.littleWinScore = 0
			} else if playerEx.betRateDown == 1 {
				playerEx.littleWinScore = playerEx.prizeValue
			} else {
				playerEx.littleWinScore = playerEx.prizeValue + playerEx.betRateDown*betValue/100
			}
			p.Statics(sceneEx.KeyGameId, sceneEx.KeyGamefreeId, -betValue, true)
			this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, proto_caothap.OpResultCode_OPRC_Sucess, params)

			playerEx.billedData = &proto_caothap.GameBilledData{
				TurnID:         proto.Int64(int64(playerEx.currentTurnID)),
				CardsData:      playerEx.cards,
				BetValue:       proto.Int64(betValue),
				PrizeValue:     proto.Int64(playerEx.prizeValue),
				JackpotValue:   proto.Int64(playerEx.jackpotValue),
				Balance:        proto.Int64(playerEx.Coin),
				Jackpot:        proto.Int64(sceneEx.jackpot.JackpotFund[playerEx.betIdx]),
				IsJackpot:      proto.Bool(playerEx.isJackpot),
				Step:           proto.Int32(int32(playerEx.step)),
				BigWinScore:    proto.Int64(playerEx.bigWinScore),
				LittleWinScore: proto.Int64(playerEx.littleWinScore),
				CardID:         proto.Int32(playerEx.cards[len(playerEx.cards)-1]),
				CurrentAces:    playerEx.currentAces,
				AcesCount:      proto.Int32(int32(playerEx.currentAcesQuantity)),
				ResponseStatus: proto.Int64(Status_Win),
			}
			pack := &proto_caothap.SCCaoThapGameBilled{
				BilledData: playerEx.billedData,
			}
			proto.SetDefaults(pack)
			logger.Logger.Infof("CaoThapPlayerOpInit %v", pack)
			p.SendToClient(int(proto_caothap.CaoThapPacketID_PACKET_SC_CAOTHAP_GAMEBILLED), pack)

			// 记录本次操作
			playerEx.RollGameType.BetInfo = append(playerEx.RollGameType.BetInfo, model.CaoThapBetInfo{
				TurnID:     playerEx.currentTurnID,
				TurnTime:   playerEx.LastOPTimer.Unix(),
				BetValue:   betValue,
				Card:       card,
				PrizeValue: betValue,
			})

			playerEx.remainTime = caothap.TIMEINTERVAL
			// 起定时器
			playerEx.playerOpHandler, _ = timer.StartTimer(timer.TimerActionWrapper(func(h timer.TimerHandle, ud interface{}) bool {
				playerEx.playerOpHandler = timer.TimerHandle(0)
				playerEx.remainTime = time.Now().Unix() - playerEx.LastOPTimer.Unix() - caothap.TIMEINTERVAL
				var locationID int64
				if playerEx.betRateDown == 0 {
					locationID = CaoThapBig
				}
				this.OnPlayerOp(s, p, CaoThapPlayerOpSetBet, []int64{locationID})
				return true
			}), nil, time.Second*(caothap.TIMEINTERVAL+3), 1)
		})

	case CaoThapPlayerOpSetBet:
		if len(params) < 1 {
			this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, proto_caothap.OpResultCode_OPRC_Error, params)
			return false
		}
		locationID := params[0]
		betValue := playerEx.prizeValue //上一局赢得分作为本局下注的分
		// 校验参数
		if locationID < CaoThapLitte || locationID > CaoThapBig {
			this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, proto_caothap.OpResultCode_OPRC_Error, params)
			return false
		}
		if playerEx.currentTurnID == 0 {
			this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, proto_caothap.OpResultCode_OPRC_Error, params)
			return false
		}
		if locationID == CaoThapLitte && playerEx.betRateDown == 0 || locationID == CaoThapBig && playerEx.betRateUp == 0 {
			this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, proto_caothap.OpResultCode_OPRC_Error, params)
			return false
		}
		//获取当前水池的上下文环境
		sceneEx.CpCtx = base.CoinPoolMgr.GetCoinPoolCtx(sceneEx.Platform, sceneEx.GetGameFreeId(), sceneEx.GetGroupId())

		//税收比例
		taxRate := sceneEx.DbGameFree.GetTaxRate()
		if taxRate < 0 || taxRate > 10000 {
			logger.Logger.Tracef("CaoThapErrorTaxRate [%v][%v][%v]", sceneEx.GetGameFreeId(), playerEx.SnId, taxRate)
			taxRate = 500
		}
		//水池设置
		coinPoolSetting := base.CoinPoolMgr.GetCoinPoolSetting(sceneEx.Platform, sceneEx.GetGameFreeId(), sceneEx.GetGroupId())
		baseRate := coinPoolSetting.GetBaseRate() //基础赔率
		ctroRate := coinPoolSetting.GetCtroRate() //调节赔率 暗税系数
		//if baseRate >= 10000 || baseRate <= 0 || ctroRate < 0 || ctroRate >= 1000 || baseRate+ctroRate > 9900 {
		//	logger.Logger.Warnf("CaoThapErrorBaseRate [%v][%v][%v][%v]", sceneEx.GetGameFreeId(), playerEx.SnId, baseRate, ctroRate)
		//	baseRate = 9700
		//	ctroRate = 200
		//}
		//jackpotRate := 10000 - (baseRate + ctroRate) //奖池系数
		jackpotRate := ctroRate //奖池系数
		logger.Logger.Tracef("CaoThapRates [%v][%v][%v][%v][%v]", sceneEx.GetGameFreeId(), playerEx.SnId, taxRate, baseRate, ctroRate)

		gamePoolCoin := base.CoinPoolMgr.LoadCoin(sceneEx.GetGameFreeId(), sceneEx.Platform, sceneEx.GetGroupId()) // 当前水池金额
		prizeFund := gamePoolCoin - sceneEx.jackpot.JackpotFund[playerEx.betIdx]                                   // 除去奖池的水池剩余金额

		var changeType int32
		var jackpotFundAdd, prizeFundAdd, changeCoin int64
		prizeFundAdd = int64(float64(betValue) * (float64(baseRate) / 10000.0))
		jackpotFundAdd = int64(float64(betValue) * (float64(jackpotRate) / 10000.0)) //奖池要增加的金额
		playerEx.isJackpot = false
		//// 超时结算金额
		//if playerEx.remainTime >= 0 && playerEx.playerOpHandler == timer.TimerHandle(0) {
		//	if playerEx.step > caothap.STEP_BALANCE {
		//		changeCoin = playerEx.prizeValue
		//	}
		//} else
		{
			spinCondition := prizeFund + prizeFundAdd
			// 计算赢分
			prizeValue := playerEx.CalcPrizeValue(betValue, spinCondition, locationID)
			//playerEx.prizeValue = playerEx.CalcPrizeValue(betValue, spinCondition, locationID)
			// 计算税收
			newScore := int64(float64(prizeValue) * float64(10000-taxRate) / 10000.0)
			playerEx.taxCoin = prizeValue - newScore
			prizeValue = newScore //本局赢得分，作为下一局押大、押小的基础分
			// A牌处理
			if playerEx.step < caothap.STEP_JACKPOT_LIMIT && caothap.IsCardA(playerEx.cards[len(playerEx.cards)-1]) && prizeValue > 0 {
				if playerEx.currentAcesQuantity == 2 {
					playerEx.isJackpot = true
					// if playerEx.isEven {
					// 	jackpotFundAdd = 0
					// } else {
					//jackpotFundAdd = int64(float64(betValue) * (float64(jackpotRate) / 10000.0)) //奖池要增加的金额
					// }
				} else {
					playerEx.currentAces = append(playerEx.currentAces, playerEx.cards[len(playerEx.cards)-1])
					playerEx.currentAcesQuantity++
					playerEx.createdTimePlay = time.Now()
				}
			}

			//dbGamePool := srvdata.PBDB_GameCoinPoolMgr.GetData(sceneEx.GetGameFreeId())
			//baseRate := dbGamePool.GetBaseRate() // 基础倍率
			baseRate := int32(9200)
			playerEx.betRateUp, playerEx.betRateDown = playerEx.CalcBetRate(baseRate)
			if playerEx.betRateUp == 0 {
				playerEx.bigWinScore = 0
			} else if playerEx.betRateUp == 1 {
				playerEx.bigWinScore = prizeValue
			} else {
				playerEx.bigWinScore = prizeValue + playerEx.betRateUp*prizeValue/100
			}
			if playerEx.betRateDown == 0 {
				playerEx.littleWinScore = 0
			} else if playerEx.betRateDown == 1 {
				playerEx.littleWinScore = prizeValue
			} else {
				playerEx.littleWinScore = prizeValue + playerEx.betRateDown*prizeValue/100
			}
			sceneEx.jackpot.JackpotFund[playerEx.betIdx] += jackpotFundAdd //奖池增加
			// 奖池水池处理
			if playerEx.isJackpot {
				playerEx.jackpotValue = sceneEx.jackpot.JackpotFund[playerEx.betIdx] / 2
				sceneEx.jackpot.JackpotFund[playerEx.betIdx] = playerEx.jackpotValue
				// 计算税收
				newScore := int64(float64(playerEx.jackpotValue) * float64(10000-taxRate) / 10000.0)
				playerEx.taxCoin += prizeValue - newScore
				playerEx.jackpotValue = newScore
				prizeValue += playerEx.jackpotValue
			}
			if playerEx.step > caothap.STEP_BALANCE {
				changeCoin = prizeValue - betValue
			} else {
				changeCoin = prizeValue
			}
			playerEx.prizeValue = prizeValue
		}
		if changeCoin >= 0 {
			changeType = common.GainWay_HundredSceneWin
			if !p.IsRob && !sceneEx.Testing {
				base.CoinPoolMgr.PushCoin(sceneEx.GetGameFreeId(), sceneEx.GetGroupId(), sceneEx.Platform, int64(float64(betValue)*(float64(10000-ctroRate)/10000.0)))
			}
			playerEx.AddServiceFee(playerEx.taxCoin)
		} else {
			changeType = common.GainWay_HundredSceneLost
		}

		// 直接计算分数
		p.AddCoin(changeCoin, changeType, true, false, "system", s.GetSceneName(), base.PlayerAddCoinDefRetryCnt, func(context *base.AddCoinContext, coin int64, ok bool) {
			//if ok {
			//	p.Coin = coin
			//}
			//p.Coin += context.Coin
			p.Coin = coin

			p.Statics(sceneEx.KeyGameId, sceneEx.KeyGamefreeId, changeCoin, true)
			this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, proto_caothap.OpResultCode_OPRC_Sucess, params)

			playerEx.step++
			playerEx.CurrentBet = betValue
			playerEx.CurrentTax = playerEx.taxCoin
			playerEx.billedData = &proto_caothap.GameBilledData{
				TurnID:         proto.Int64(int64(playerEx.currentTurnID)),
				CardsData:      playerEx.cards,
				BetValue:       proto.Int64(betValue),
				PrizeValue:     proto.Int64(playerEx.prizeValue),
				JackpotValue:   proto.Int64(playerEx.jackpotValue),
				Balance:        proto.Int64(playerEx.Coin),
				Jackpot:        proto.Int64(sceneEx.jackpot.JackpotFund[playerEx.betIdx]),
				IsJackpot:      proto.Bool(playerEx.isJackpot),
				Step:           proto.Int32(int32(playerEx.step)),
				BigWinScore:    proto.Int64(playerEx.bigWinScore),
				LittleWinScore: proto.Int64(playerEx.littleWinScore),
				CardID:         proto.Int32(playerEx.cards[len(playerEx.cards)-1]),
				CurrentAces:    playerEx.currentAces,
				AcesCount:      proto.Int32(int32(playerEx.currentAcesQuantity)),
			}
			if playerEx.prizeValue > 0 {
				playerEx.billedData.ResponseStatus = Status_Win
			} else {
				playerEx.billedData.ResponseStatus = Status_Lose
			}
			pack := &proto_caothap.SCCaoThapGameBilled{
				BilledData: playerEx.billedData,
			}
			proto.SetDefaults(pack)
			logger.Logger.Infof("SCCaoThapGameBilled %v", pack)
			p.SendToClient(int(proto_caothap.CaoThapPacketID_PACKET_SC_CAOTHAP_GAMEBILLED), pack)

			//if !(playerEx.remainTime <= 0 && playerEx.playerOpHandler == timer.TimerHandle(0)) {
			// 记录本次操作
			playerEx.RollGameType.BetInfo = append(playerEx.RollGameType.BetInfo, model.CaoThapBetInfo{
				TurnID:     playerEx.currentTurnID,
				TurnTime:   time.Now().Unix(),
				BetValue:   betValue,
				Card:       playerEx.cards[len(playerEx.cards)-1],
				PrizeValue: playerEx.prizeValue,
			})
			//}

			playerEx.isEven = false

			// 广播奖池
			if betValue != 0 || playerEx.isJackpot {
				// 添加进开奖记录里面
				if playerEx.isJackpot {
					playerEx.lastJackpotTime = time.Now()
					spinid := strconv.FormatInt(int64(playerEx.SnId), 10)
					// 推送最新开奖记录到world
					msg := &proto_server.GWGameNewBigWinHistory{
						SceneId: proto.Int32(int32(sceneEx.SceneId)),
						BigWinHistory: &proto_server.BigWinHistoryInfo{
							SpinID:      proto.String(spinid),
							CreatedTime: proto.Int64(playerEx.lastJackpotTime.Unix()),
							BaseBet:     proto.Int64(int64(betValue)),
							TotalBet:    proto.Int64(int64(betValue)),
							PriceValue:  proto.Int64(int64(pack.BilledData.GetJackpotValue())),
							UserName:    proto.String(playerEx.Name),
						},
					}
					proto.SetDefaults(msg)
					logger.Logger.Infof("GWGameNewBigWinHistory gameFreeID %v %v", sceneEx.DbGameFree.GetId(), msg)
					sceneEx.SendToWorld(int(proto_server.SSPacketID_PACKET_GW_GAMENEWBIGWINHISTORY), msg)
				} else {
					sceneEx.PushVirtualDataToWorld() // 推送虚拟数据
				}
				sceneEx.BroadcastJackpot()
			}

			//超时，模拟玩家执行结束操作
			if playerEx.remainTime >= 0 && playerEx.playerOpHandler == timer.TimerHandle(0) {
				this.OnPlayerOp(s, p, CaoThapPlayerOpStop, []int64{})
			}

			// 玩家输 或者 爆奖池
			if playerEx.billedData.ResponseStatus == Status_Lose || playerEx.billedData.IsJackpot {
				playerEx.RollGameType.TotalPriceValue = playerEx.prizeValue
				playerEx.RollGameType.Cards = playerEx.cards
				playerEx.RollGameType.WinJackpot = playerEx.jackpotValue
				CaoThapCheckAndSaveLog(sceneEx, playerEx)
				playerEx.CleanPlayerData()
			}
		})
	case CaoThapPlayerOpStop:
		if playerEx.currentTurnID == 0 {
			this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, proto_caothap.OpResultCode_OPRC_Error, params)
			return false
		}
		if playerEx.step < caothap.STEP_BALANCE {
			this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, proto_caothap.OpResultCode_OPRC_Error, params)
			return false
		}
		// 停掉定时器
		if playerEx.playerOpHandler != timer.TimerHandle(0) {
			if ok := timer.StopTimer(playerEx.playerOpHandler); !ok {
				logger.Logger.Infof("CaoThapPlayerOpStop StopTimer error: ", errors.New("TimerHandler is not found"))
			}
			playerEx.playerOpHandler = timer.TimerHandle(0)
		}
		playerEx.RollGameType.TotalPriceValue = playerEx.prizeValue
		playerEx.RollGameType.Cards = playerEx.cards
		playerEx.RollGameType.WinJackpot = playerEx.jackpotValue

		playerEx.billedData = &proto_caothap.GameBilledData{
			TurnID:         proto.Int64(int64(playerEx.currentTurnID)),
			BetValue:       proto.Int64(0),
			ResponseStatus: proto.Int64(Status_Stop),
		}
		pack := &proto_caothap.SCCaoThapGameBilled{
			BilledData: playerEx.billedData,
		}
		proto.SetDefaults(pack)
		logger.Logger.Infof("CaoThapPlayerOpStop %v", pack)
		p.SendToClient(int(proto_caothap.CaoThapPacketID_PACKET_SC_CAOTHAP_GAMEBILLED), pack)
		// 保存记录
		CaoThapCheckAndSaveLog(sceneEx, playerEx)
		// 重置用户数据
		playerEx.CleanPlayerData()
	case CaoThapPlayerHistory:
		task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
			gpl := model.GetPlayerListByHallEx(p.SnId, p.Platform, 0, 50, 0, 0, s.DbGameFree.SceneType, s.DbGameFree.GetGameClass(), s.GameId)
			pack := &proto_caothap.SCCaoThapPlayerHistory{}
			for _, v := range gpl.Data {
				if v.GameDetailedLogId == "" {
					logger.Logger.Error("CaoThapPlayerHistory GameDetailedLogId is nil")
					break
				}
				gdl := model.GetPlayerHistory(p.Platform, v.GameDetailedLogId)
				if gdl == nil {
					logger.Logger.Error("CaoThapPlayerHistory gdl is nil")
					continue
				}
				data, err := model.UnMarshalCaoThapGameNote(gdl.GameDetailedNote)
				if err != nil {
					logger.Logger.Errorf("UnMarshalCaoThapGameNote error:%v", err)
				}
				if gnd, ok := data.(*model.CaoThapType); ok {
					for _, v := range gnd.BetInfo {
						player := &proto_caothap.CaoThapPlayerHistoryInfo{
							SpinID:      proto.String(strconv.Itoa(int(v.TurnID))),
							CreatedTime: proto.Int64(int64(v.TurnTime)),
							BetValue:    proto.Int64(int64(v.BetValue)),
							PriceValue:  v.PrizeValue,
							CardID:      v.Card,
						}
						pack.PlayerHistory = append(pack.PlayerHistory, player)
					}
				}
			}
			proto.SetDefaults(pack)
			logger.Logger.Info("MiniPokerPlayerHistory: ", pack)
			return pack

		}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
			if data == nil {
				logger.Logger.Error("MiniPokerPlayerHistory data is nil")
				return
			}
			p.SendToClient(int(proto_caothap.CaoThapPacketID_PACKET_SC_CAOTHAP_PLAYERHISTORY), data)
		}), "CSGetCaoThapPlayerHistoryHandler").Start()
	case CaoThapPlayerSelBet:
		//参数是否合法
		//先做底注校验
		// 底注校验
		if !common.InSliceInt32(sceneEx.DbGameFree.GetOtherIntParams(), int32(params[0])) {
			this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, proto_caothap.OpResultCode_OPRC_Error, params)
			return false
		}
		playerEx.score = int32(params[0]) // 单线押注数
		playerEx.betIdx = int32(common.InSliceInt32Index(sceneEx.DbGameFree.GetOtherIntParams(), playerEx.score))
		jackpotVal := sceneEx.jackpot.JackpotFund[playerEx.betIdx]
		params = append(params, jackpotVal)
		this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, proto_caothap.OpResultCode_OPRC_Sucess, params)
	}
	return true
}

func (this *SceneStateCaoThapStart) OnPlayerEvent(s *base.Scene, p *base.Player, evtcode int, params []int64) {
}

//发送玩家操作情况
func (this *SceneStateCaoThapStart) OnPlayerSToCOp(s *base.Scene, p *base.Player, pos int, opcode int,
	opRetCode proto_caothap.OpResultCode, params []int64) {
	pack := &proto_caothap.SCCaoThapOp{
		SnId:      proto.Int32(p.SnId),
		OpCode:    proto.Int(opcode),
		OpRetCode: opRetCode,
		Params:    params,
	}
	proto.SetDefaults(pack)
	p.SendToClient(int(proto_caothap.CaoThapPacketID_PACKET_SC_CAOTHAP_PLAYEROP), pack)
}

func CaoThapCheckAndSaveLog(sceneEx *CaoThapSceneData, playerEx *CaoThapPlayerData) {
	//统计金币变动
	//log1
	logger.Logger.Trace("CaoThapCheckAndSaveLog Save ", playerEx.SnId)
	//changeCoin := playerEx.Coin - playerEx.StartCoin
	changeCoin := playerEx.winCoin - playerEx.taxCoin - playerEx.betValue
	startCoin := playerEx.Coin - changeCoin
	playerEx.SaveSceneCoinLog(startCoin, changeCoin,
		playerEx.Coin, playerEx.CurrentBet, playerEx.taxCoin, playerEx.winCoin, playerEx.jackpotValue, 0)

	//log2
	playerEx.RollGameType.ChangeCoin = changeCoin
	playerEx.RollGameType.BasicScore = sceneEx.DbGameFree.GetBaseScore()
	playerEx.RollGameType.RoomId = int32(sceneEx.SceneId)
	playerEx.RollGameType.AfterCoin = playerEx.Coin
	playerEx.RollGameType.BeforeCoin = startCoin
	playerEx.RollGameType.IsFirst = sceneEx.IsPlayerFirst(playerEx.Player)
	playerEx.RollGameType.PlayerSnid = playerEx.SnId
	playerEx.RollGameType.Score = int32(playerEx.CurrentBet)
	playerEx.RollGameType.UserName = playerEx.Name
	playerEx.RollGameType.Tax = playerEx.taxCoin
	playerEx.RollGameType.WBLevel = playerEx.WBLevel
	if playerEx.score > 0 {
		if !playerEx.IsRob {
			info, err := model.MarshalGameNoteByMini(playerEx.RollGameType)
			if err == nil {
				logid, _ := model.AutoIncGameLogId()
				sceneEx.SaveGameDetailedLog(logid, info, &base.GameDetailedParam{})
				totalin := int64(playerEx.RollGameType.Score)
				totalout := playerEx.RollGameType.ChangeCoin + playerEx.taxCoin + totalin
				sceneEx.SaveGamePlayerListLog(playerEx.SnId,
					&base.SaveGamePlayerListLogParam{
						Platform:          playerEx.Platform,
						Channel:           playerEx.Channel,
						Promoter:          playerEx.BeUnderAgentCode,
						PackageTag:        playerEx.PackageID,
						InviterId:         playerEx.InviterId,
						LogId:             logid,
						TotalIn:           totalin,
						TotalOut:          totalout,
						TaxCoin:           playerEx.taxCoin,
						ClubPumpCoin:      0,
						BetAmount:         int64(playerEx.RollGameType.Score),
						WinAmountNoAnyTax: playerEx.RollGameType.ChangeCoin,
						IsFirstGame:       sceneEx.IsPlayerFirst(playerEx.Player),
					})
			}
		}
	}

	//统计输下注金币数
	if !sceneEx.Testing && !playerEx.IsRob {
		playerBet := &proto_server.PlayerBet{
			SnId: proto.Int32(playerEx.SnId),
			Bet:  proto.Int64(int64(playerEx.CurrentBet)),
			Gain: proto.Int64(playerEx.RollGameType.ChangeCoin),
			Tax:  proto.Int64(playerEx.taxCoin),
		}
		gwPlayerBet := &proto_server.GWPlayerBet{
			GameFreeId: proto.Int32(sceneEx.DbGameFree.GetId()),
			RobotGain:  proto.Int64(-playerEx.RollGameType.ChangeCoin),
		}
		gwPlayerBet.PlayerBets = append(gwPlayerBet.PlayerBets, playerBet)
		proto.SetDefaults(gwPlayerBet)
		sceneEx.SendToWorld(int(proto_server.SSPacketID_PACKET_GW_PLAYERBET), gwPlayerBet)
		logger.Logger.Trace("Send msg gwPlayerBet ===>", gwPlayerBet)
	}

	playerEx.taxCoin = 0
	playerEx.winCoin = 0
	playerEx.jackpotValue = 0

	if sceneEx.CheckNeedDestroy() {
		sceneEx.PlayerLeave(playerEx.Player, common.PlayerLeaveReason_OnDestroy, true)
	}
}

func init() {
	ScenePolicyCaoThapSington.RegisteSceneState(&SceneStateCaoThapStart{})
	core.RegisteHook(core.HOOK_BEFORE_START, func() error {
		base.RegisteScenePolicy(common.GameId_CaoThap, 0, ScenePolicyCaoThapSington)
		return nil
	})
}
