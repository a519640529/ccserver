package minipoker

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"games.yol.com/win88/common"
	"games.yol.com/win88/gamerule/minipoker"
	"games.yol.com/win88/minigame/base"
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	proto_minipoker "games.yol.com/win88/protocol/minipoker"
	proto_server "games.yol.com/win88/protocol/server"
	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/task"
	"github.com/idealeak/goserver/core/timer"
)

////////////////////////////////////////////////////////////////////////////////
// MiniPoker
////////////////////////////////////////////////////////////////////////////////

// 房间内主要逻辑
var ScenePolicyMiniPokerSington = &ScenePolicyMiniPoker{}

type ScenePolicyMiniPoker struct {
	base.BaseScenePolicy
	states [MiniPokerSceneStateMax]base.SceneState
}

//创建场景扩展数据
func (this *ScenePolicyMiniPoker) CreateSceneExData(s *base.Scene) interface{} {
	sceneEx := NewMiniPokerSceneData(s)
	if sceneEx != nil {
		if sceneEx.init() {
			s.ExtraData = sceneEx
		}
	}
	return sceneEx
}

//创建玩家扩展数据
func (this *ScenePolicyMiniPoker) CreatePlayerExData(s *base.Scene, p *base.Player) interface{} {
	playerEx := &MiniPokerPlayerData{Player: p}
	if playerEx != nil {
		p.ExtraData = playerEx
	}
	return playerEx
}

//场景开启事件
func (this *ScenePolicyMiniPoker) OnStart(s *base.Scene) {
	logger.Trace("(this *ScenePolicyMiniPoker) OnStart, sceneId=", s.SceneId)
	sceneEx := NewMiniPokerSceneData(s)
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
			s.ChangeSceneState(MiniPokerSceneStateStart) //改变当前的玩家状态
		}
	}
}

//场景关闭事件
func (this *ScenePolicyMiniPoker) OnStop(s *base.Scene) {
	logger.Trace("(this *ScenePolicyMiniPoker) OnStop , sceneId=", s.SceneId)
	if sceneEx, ok := s.ExtraData.(*MiniPokerSceneData); ok {
		if sceneEx.jackpotNoticeHandle != timer.TimerHandle(0) {
			timer.StopTimer(sceneEx.jackpotNoticeHandle)
		}
		sceneEx.SaveData(true)
	}
}

//场景心跳事件
func (this *ScenePolicyMiniPoker) OnTick(s *base.Scene) {
	if s == nil {
		return
	}
	if s.SceneState != nil {
		s.SceneState.OnTick(s)
	}
}

//玩家进入事件
func (this *ScenePolicyMiniPoker) OnPlayerEnter(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Trace("(this *ScenePolicyMiniPoker) OnPlayerEnter, sceneId=", s.SceneId, " player=", p.SnId)
	if sceneEx, ok := s.ExtraData.(*MiniPokerSceneData); ok {
		playerEx := &MiniPokerPlayerData{Player: p}
		playerEx.init(s) // 玩家当前信息初始化
		sceneEx.players[p.SnId] = playerEx
		p.ExtraData = playerEx
		MiniPokerSendRoomInfo(s, p, sceneEx, playerEx, nil)
		s.FirePlayerEvent(p, base.PlayerEventEnter, nil) //回调会调取 onPlayerEvent事件
	}
}

//玩家离开事件
func (this *ScenePolicyMiniPoker) OnPlayerLeave(s *base.Scene, p *base.Player, reason int) {
	if s == nil || p == nil {
		return
	}
	logger.Trace("(this *ScenePolicyMiniPoker) OnPlayerLeave, sceneId=", s.SceneId, " player=", p.SnId)
	if sceneEx, ok := s.ExtraData.(*MiniPokerSceneData); ok {
		if this.CanChangeCoinScene(s, p) {
			sceneEx.OnPlayerLeave(p, reason)
			s.FirePlayerEvent(p, base.PlayerEventLeave, []int64{int64(reason)})
		}
	}
}

//玩家掉线
func (this *ScenePolicyMiniPoker) OnPlayerDropLine(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Trace("(this *ScenePolicyMiniPoker) OnPlayerDropLine, sceneId=", s.SceneId, " player=", p.SnId)
	s.FirePlayerEvent(p, base.PlayerEventDropLine, nil)
	if sceneEx, ok := s.ExtraData.(*MiniPokerSceneData); ok {
		if sceneEx.Gaming {
			return
		}
	}
}

//玩家重连
func (this *ScenePolicyMiniPoker) OnPlayerRehold(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Trace("(this *ScenePolicyMiniPoker) OnPlayerRehold, sceneId=", s.SceneId, " player=", p.Name)
	if _, ok := s.ExtraData.(*MiniPokerSceneData); ok {
		if _, ok := p.ExtraData.(*MiniPokerPlayerData); ok {
			//发送房间信息给自己
			//MiniPokerSendRoomInfo(s, p, sceneEx, playerEx)
			s.FirePlayerEvent(p, base.PlayerEventRehold, nil)
		}
	}
}

//玩家返回房间
func (this *ScenePolicyMiniPoker) OnPlayerReturn(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Trace("(this *ScenePolicyMiniPoker) OnPlayerReturn,sceneId=", s.SceneId, " player= ", p.Name)
	if sceneEx, ok := s.ExtraData.(*MiniPokerSceneData); ok {
		if playerEx, ok := p.ExtraData.(*MiniPokerPlayerData); ok {
			MiniPokerSendRoomInfo(s, p, sceneEx, playerEx, playerEx.billedData)
			s.FirePlayerEvent(p, base.PlayerReturn, nil)
		}
	}
}

//玩家操作
func (this *ScenePolicyMiniPoker) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if s == nil || p == nil {
		return false
	}
	if s.SceneState != nil {
		return s.SceneState.OnPlayerOp(s, p, opcode, params)
	}
	return true
}

//玩家事件
func (this *ScenePolicyMiniPoker) OnPlayerEvent(s *base.Scene, p *base.Player, evtcode int, params []int64) {
	if s == nil || p == nil {
		return
	}
	if s.SceneState != nil {
		s.SceneState.OnPlayerEvent(s, p, evtcode, params)
	}
}

//是否完成了整个牌局
func (this *ScenePolicyMiniPoker) IsCompleted(s *base.Scene) bool { return false }

//是否可以强制开始
func (this *ScenePolicyMiniPoker) IsCanForceStart(s *base.Scene) bool { return true }

//当前状态能否换桌
func (this *ScenePolicyMiniPoker) CanChangeCoinScene(s *base.Scene, p *base.Player) bool {
	if s == nil || p == nil {
		return true
	}
	if s.SceneState != nil {
		return s.SceneState.CanChangeScene(s, p)
	}
	return true
}

func (this *ScenePolicyMiniPoker) RegisteSceneState(state base.SceneState) {
	if state == nil {
		return
	}
	stateid := state.GetState()
	if stateid < 0 || stateid >= MiniPokerSceneStateMax {
		return
	}
	this.states[stateid] = state
}

func (this *ScenePolicyMiniPoker) GetSceneState(s *base.Scene, stateid int) base.SceneState {
	if stateid >= 0 && stateid < MiniPokerSceneStateMax {
		return ScenePolicyMiniPokerSington.states[stateid]
	}
	return nil
}

func MiniPokerSendRoomInfo(s *base.Scene, p *base.Player, sceneEx *MiniPokerSceneData, playerEx *MiniPokerPlayerData, data *proto_minipoker.GameBilledData) {
	logger.Trace("-------------------发送房间消息 ", s.SceneId, p.SnId)
	pack := &proto_minipoker.SCMiniPokerRoomInfo{
		RoomId:     proto.Int(s.SceneId),
		GameId:     proto.Int(s.GameId),
		RoomMode:   proto.Int(s.GameMode),
		Params:     s.GetParams(),
		State:      proto.Int(s.SceneState.GetState()),
		Jackpot:    proto.Int64(sceneEx.jackpot.JackpotFund[playerEx.betIdx]),
		GameFreeId: proto.Int32(s.GetDBGameFree().GetId()),
		BilledData: data,
	}
	if playerEx != nil {
		pd := &proto_minipoker.MiniPokerPlayerData{
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
	proto.SetDefaults(pack)
	p.SendToClient(int(proto_minipoker.MiniPokerPacketID_PACKET_SC_MINIPOKER_ROOMINFO), pack)
}

type SceneStateMiniPokerStart struct {
}

//获取当前场景状态
func (this *SceneStateMiniPokerStart) GetState() int { return MiniPokerSceneStateStart }

//是否可以切换状态到
func (this *SceneStateMiniPokerStart) CanChangeTo(s base.SceneState) bool { return true }

//当前状态能否换桌
func (this *SceneStateMiniPokerStart) CanChangeScene(s *base.Scene, p *base.Player) bool {
	if _, ok := p.ExtraData.(*MiniPokerPlayerData); ok {
		return true
	}
	return true
}

func (this *SceneStateMiniPokerStart) GetTimeout(s *base.Scene) int { return 0 }

func (this *SceneStateMiniPokerStart) OnEnter(s *base.Scene) {
	if sceneEx, ok := s.ExtraData.(*MiniPokerSceneData); ok {
		logger.Tracef("(this *Scene) [%v] 场景状态进入 %v", s.SceneId, len(sceneEx.players))
		sceneEx.StateStartTime = time.Now()
		pack := &proto_minipoker.SCMiniPokerRoomState{
			State: proto.Int(this.GetState()),
		}
		proto.SetDefaults(pack)
		s.Broadcast(int(proto_minipoker.MiniPokerPacketID_PACKET_SC_MINIPOKER_ROOMSTATE), pack, 0)
	}
}

func (this *SceneStateMiniPokerStart) OnLeave(s *base.Scene) {}

func (this *SceneStateMiniPokerStart) OnTick(s *base.Scene) {
	if sceneEx, ok := s.ExtraData.(*MiniPokerSceneData); ok {
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

func (this *SceneStateMiniPokerStart) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	playerEx, ok := p.ExtraData.(*MiniPokerPlayerData)
	if !ok {
		return false
	}
	sceneEx, ok := s.ExtraData.(*MiniPokerSceneData)
	if !ok {
		return false
	}
	if sceneEx.CheckNeedDestroy() {
		//离开有统计
		sceneEx.PlayerLeave(playerEx.Player, common.PlayerLeaveReason_OnDestroy, true)
		return false
	}
	switch opcode {
	case MiniPokerPlayerOpStart: //开始
		// 洗牌
		// playerEx.cardsData = minipoker.CardInit()

		//if !minipokerBenchTest {
		//	minipokerBenchTest = true
		//	//for i := 0; i < 10; i++ {
		//	//	this.BenchTest(s, p)
		//	//	//this.WinTargetBenchTest(s, p)
		//	//}
		//	this.MultiplayerBenchTest(s)
		//	return true
		//}

		//先做底注校验
		if !common.InSliceInt32(sceneEx.DbGameFree.GetOtherIntParams(), int32(params[0])) {
			this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, proto_minipoker.OpResultCode_OPRC_Error, params)
			return false
		}
		playerEx.score = int32(params[0]) // 压注数
		playerEx.betIdx = int32(common.InSliceInt32Index(sceneEx.DbGameFree.GetOtherIntParams(), playerEx.score))
		betValue := int64(playerEx.score)
		if betValue > playerEx.Coin {
			this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, proto_minipoker.OpResultCode_OPRC_CoinNotEnough, params)
			return false
		} else if int64(sceneEx.DbGameFree.GetBetLimit()) > playerEx.Coin { //押注限制
			this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, proto_minipoker.OpResultCode_OPRC_CoinNotEnough, params)
			return false
		}
		logger.Logger.Infof("minipoker 开始转动 当前金币%v 总下注:%v", playerEx.Coin, betValue)
		//扣除投注金币
		p.AddCoin(-betValue, common.GainWay_HundredSceneLost, true, false, "system", s.GetSceneName(), base.PlayerAddCoinDefRetryCnt, func(context *base.AddCoinContext, coin int64, ok bool) {
			if !ok {
				this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, proto_minipoker.OpResultCode_OPRC_CoinNotEnough, params)
				return
			}
			logger.Logger.Tracef("minipoker 扣钱事务成功 当前玩家金币:%v context:%v 回传金币:%v", p.Coin, context.Player.Coin, context.Coin)
			//p.Coin += context.Coin
			p.Coin = coin
			p.LastOPTimer = time.Now()
			sceneEx.GameNowTime = time.Now()
			sceneEx.NumOfGames++
			p.GameTimes++
			//playerEx.StartCoin = playerEx.Coin

			//获取当前水池的上下文环境
			sceneEx.CpCtx = base.CoinPoolMgr.GetCoinPoolCtx(sceneEx.Platform, sceneEx.GetGameFreeId(), sceneEx.GetGroupId())

			//税收比例
			taxRate := sceneEx.DbGameFree.GetTaxRate()
			if taxRate < 0 || taxRate > 10000 {
				logger.Tracef("MiniPokerErrorTaxRate [%v][%v][%v]", sceneEx.GetGameFreeId(), playerEx.SnId, taxRate)
				taxRate = 500
			}
			//水池设置
			coinPoolSetting := base.CoinPoolMgr.GetCoinPoolSetting(sceneEx.Platform, sceneEx.GetGameFreeId(), sceneEx.GetGroupId())
			baseRate := coinPoolSetting.GetBaseRate() //基础赔率
			ctroRate := coinPoolSetting.GetCtroRate() //调节赔率 暗税系数
			//if baseRate >= 10000 || baseRate <= 0 || ctroRate < 0 || ctroRate >= 1000 || baseRate+ctroRate > 9900 {
			//	logger.Warnf("MiniPokerErrorBaseRate [%v][%v][%v][%v]", sceneEx.GetGameFreeId(), playerEx.SnId, baseRate, ctroRate)
			//	baseRate = 9700
			//	ctroRate = 200
			//}
			//jackpotRate := 10000 - (baseRate + ctroRate) //奖池系数
			jackpotRate := ctroRate //奖池系数
			logger.Tracef("MiniPokerRates [%v][%v][%v][%v][%v]", sceneEx.GetGameFreeId(), playerEx.SnId, taxRate, baseRate, ctroRate)

			gamePoolCoin := base.CoinPoolMgr.LoadCoin(sceneEx.GetGameFreeId(), sceneEx.Platform, sceneEx.GetGroupId()) // 当前水池金额
			prizeFund := gamePoolCoin - sceneEx.jackpot.JackpotFund[playerEx.betIdx]                                   // 除去奖池的水池剩余金额

			// 奖池参数
			var jackpotParam = sceneEx.DbGameFree.GetJackpot()
			var jackpotInit = int64(jackpotParam[minipoker.MINIPOKER_JACKPOT_InitJackpot] * playerEx.score) //奖池初始值

			var jackpotFundAdd, prizeFundAdd int64                                       // 选线记录
			jackpotFundAdd = int64(float64(betValue) * (float64(jackpotRate) / 10000.0)) //奖池要增加的金额
			prizeFundAdd = int64(float64(betValue) * (float64(baseRate) / 10000.0))
			sceneEx.jackpot.JackpotFund[playerEx.betIdx] += jackpotFundAdd //奖池增加
			playerEx.TotalBet += betValue                                  //总下注额（从进房间开始,包含多局游戏的下注）

			p.Statics(sceneEx.KeyGameId, sceneEx.KeyGamefreeId, -betValue, true)
			if !p.IsRob && !sceneEx.Testing {
				// 推送金币
				base.CoinPoolMgr.PushCoin(sceneEx.GetGameFreeId(), sceneEx.GetGroupId(), sceneEx.Platform, int64(float64(betValue)*(float64(10000-ctroRate)/10000.0)))
			}
			writeBlackTryTimes := 0
		WriteBlack:
			var spinRes MiniPokerSpinResult
			var slotDataIsOk bool
			var cardsType int
			for i := 0; i < 3; i++ {
				// playerEx.cards = playerEx.cardsData[playerEx.cardPos : playerEx.cardPos+minipoker.CARDNUM]
				playerEx.cards = minipoker.GetCards()
				playerEx.cardPos += minipoker.CARDNUM
				cardsType = minipoker.CalcCardsType(playerEx.cards)
				if !minipoker.CheckCardsType(cardsType) {
					continue
				}
				spinRes = sceneEx.CalcCardsPrize(params[0], cardsType)

				// 水池不足以支付玩家
				spinCondition := prizeFund + prizeFundAdd - spinRes.PrizeValue
				if spinRes.IsJackpot {
					spinCondition += sceneEx.jackpot.JackpotFund[playerEx.betIdx] - jackpotInit*2
				}
				if spinCondition <= 0 {
					if !spinRes.IsJackpot {
						writeBlackTryTimes = 999
						break
					}
					continue
				}
				slotDataIsOk = true
				break
			}

			if !slotDataIsOk {
				for i := 0; i < 13; i++ {
					// playerEx.cards = playerEx.cardsData[playerEx.cardPos : playerEx.cardPos+minipoker.CARDNUM]
					playerEx.cards = minipoker.GetCards()
					playerEx.cardPos += minipoker.CARDNUM
					// 重新洗牌
					/* if playerEx.cardPos > minipoker.CARDDATANUM-minipoker.CARDNUM {
						playerEx.cardPos = 0
						playerEx.cardsData = minipoker.CardInit()
					} */
					cardsType = minipoker.CalcCardsType(playerEx.cards)
					if !minipoker.CheckCardsType(cardsType) {
						continue
					}
					if cardsType != minipoker.CARDTYPE_ONEPAIR_J && cardsType != minipoker.CARDTYPE_HIGHCARD {
						continue
					} else {
						break
					}
				}
				spinRes = sceneEx.CalcCardsPrize(params[0], cardsType)
			}

			// 黑白名单调控 防止异常循环，添加上限次数
			if writeBlackTryTimes < 100 && playerEx.CheckBlackWriteList(spinRes.PrizeValue > betValue) {
				writeBlackTryTimes++
				goto WriteBlack
			} else if writeBlackTryTimes >= 100 && writeBlackTryTimes != 999 {
				logger.Warnf("MiniPokerWriteBlackTryTimesOver [%v][%v][%v][%v][%v]", sceneEx.GetGameFreeId(), playerEx.SnId, gamePoolCoin, playerEx.BlackLevel, playerEx.WhiteLevel)
			}

			//if playerEx.debugJackpot {
			//	playerEx.cards = []int32{5, 6, 7, 8, 9}
			//	cardsType = minipoker.CalcCardsType(playerEx.cards)
			//	spinRes = sceneEx.CalcCardsPrize(params[0], cardsType)
			//	playerEx.debugJackpot = false
			//}
			//if playerEx.DebugGame && playerEx.betIdx == 0 {
			//	if playerEx.TestNum >= len(DebugData) {
			//		playerEx.TestNum = 0
			//	}
			//	playerEx.cards = DebugData[playerEx.TestNum]
			//	cardsType = minipoker.CalcCardsType(playerEx.cards)
			//	spinRes = sceneEx.CalcCardsPrize(params[0], cardsType)
			//	playerEx.TestNum++
			//}
			// 奖池水池处理
			if spinRes.IsJackpot {
				sceneEx.jackpot.JackpotFund[playerEx.betIdx] = jackpotInit
			}

			if spinRes.PrizeValue >= 0 {
				p.AddCoin(spinRes.PrizeValue, common.GainWay_HundredSceneWin, false, false, "system", s.GetSceneName(), base.PlayerAddCoinDefRetryCnt, func(context *base.AddCoinContext, coin int64, b bool) {
					//if b {
					//	p.Coin = coin
					//}
					//p.Coin += context.Coin
					p.Coin = coin
					p.Statics(sceneEx.KeyGameId, sceneEx.KeyGamefreeId, spinRes.PrizeValue+spinRes.TotalTaxScore, true)
					if !p.IsRob && !sceneEx.Testing {
						base.CoinPoolMgr.PopCoin(sceneEx.GetGameFreeId(), sceneEx.GetGroupId(), sceneEx.Platform, spinRes.PrizeValue+spinRes.TotalTaxScore)
					}
					playerEx.taxCoin = spinRes.TotalTaxScore
					playerEx.AddServiceFee(playerEx.taxCoin)
					this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, proto_minipoker.OpResultCode_OPRC_Sucess, params)

					// minipoker.SpinID++
					// playerEx.spinID = minipoker.SpinID
					// playerEx.cardPos = 0
					playerEx.winCoin = spinRes.PrizeValue + playerEx.taxCoin
					//playerEx.linesWinCoin = spinRes.CardTypeValue
					playerEx.jackpotWinCoin = spinRes.JackpotValue
					playerEx.CurrentBet = betValue
					playerEx.CurrentTax = playerEx.taxCoin

					playerEx.billedData = &proto_minipoker.GameBilledData{
						//SpinID:       proto.Int64(playerEx.spinID),
						BetValue:     proto.Int64(betValue),
						Cards:        playerEx.cards,
						IsJackpot:    proto.Bool(spinRes.IsJackpot),
						PrizeValue:   proto.Int64(spinRes.PrizeValue),   //总赢分
						JackpotValue: proto.Int64(spinRes.JackpotValue), //奖池得分
						Balance:      proto.Int64(playerEx.Coin),
						Jackpot:      proto.Int64(sceneEx.jackpot.JackpotFund[playerEx.betIdx]),
					}
					pack := &proto_minipoker.SCMiniPokerGameBilled{
						BilledData: playerEx.billedData,
					}
					proto.SetDefaults(pack)
					logger.Logger.Infof("MiniPokerPlayerOpStart %v", pack)
					p.SendToClient(int(proto_minipoker.MiniPokerPacketID_PACKET_SC_MINIPOKER_GAMEBILLED), pack)
					// 记录本次操作this

					playerEx.RollGameType.BaseResult.Cards = pack.BilledData.GetCards()
					playerEx.RollGameType.BaseResult.WinRate = spinRes.WinRate
					playerEx.RollGameType.CardsType = spinRes.CardsType
					playerEx.RollGameType.BaseResult.WinLineScore = pack.BilledData.GetPrizeValue() - pack.BilledData.GetJackpotValue()
					playerEx.RollGameType.BaseResult.WinJackpot = pack.BilledData.GetJackpotValue()
					playerEx.RollGameType.BaseResult.WinTotal = pack.BilledData.GetPrizeValue()

					MiniPokerCheckAndSaveLog(sceneEx, playerEx)

					// 广播奖池
					if betValue == 0 && !spinRes.IsJackpot { // 没改变奖池
						return
					}
					// 添加进开奖记录里面
					if spinRes.IsJackpot {
						playerEx.lastJackpotTime = time.Now()
						spinid := strconv.FormatInt(int64(playerEx.SnId), 10)
						// 推送最新开奖记录到world
						msg := &proto_server.GWGameNewBigWinHistory{
							SceneId: proto.Int32(int32(sceneEx.SceneId)),
							BigWinHistory: &proto_server.BigWinHistoryInfo{
								SpinID:      proto.String(spinid),
								CreatedTime: proto.Int64(playerEx.lastJackpotTime.Unix()),
								BaseBet:     proto.Int64(int64(playerEx.score)),
								TotalBet:    proto.Int64(int64(playerEx.score)),
								PriceValue:  proto.Int64(int64(pack.BilledData.GetJackpotValue())),
								UserName:    proto.String(playerEx.Name),
								Cards:       playerEx.cards,
							},
						}
						proto.SetDefaults(msg)
						logger.Logger.Infof("GWGameNewBigWinHistory gameFreeID %v %v", sceneEx.DbGameFree.GetId(), msg)
						sceneEx.SendToWorld(int(proto_server.SSPacketID_PACKET_GW_GAMENEWBIGWINHISTORY), msg)
					} else {
						sceneEx.PushVirtualDataToWorld() // 推送虚拟数据
					}
					sceneEx.BroadcastJackpot()
				})
			}
		})
	case MiniPokerPlayerHistory:
		task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
			spinid := strconv.FormatInt(int64(playerEx.SnId), 10)
			gpl := model.GetPlayerListByHallEx(p.SnId, p.Platform, 0, 50, 0, 0, s.DbGameFree.SceneType, s.DbGameFree.GetGameClass(), s.GameId)
			pack := &proto_minipoker.SCMiniPokerPlayerHistory{}
			for _, v := range gpl.Data {
				if v.GameDetailedLogId == "" {
					logger.Error("CandyPlayerHistory GameDetailedLogId is nil")
					break
				}
				gdl := model.GetPlayerHistory(p.Platform, v.GameDetailedLogId)
				if gdl == nil {
					logger.Logger.Error("CandyPlayerHistory gdl is nil")
					continue
				}
				data, err := UnMarshalMiniPokerGameNote(gdl.GameDetailedNote)
				if err != nil {
					logger.Errorf("UnMarshalCandyGameNote error:%v", err)
				}
				if gnd, ok := data.(*GameResultLog); ok {
					if gnd.BaseResult != nil {
						player := &proto_minipoker.MiniPokerPlayerHistoryInfo{
							SpinID:          proto.String(spinid),
							CreatedTime:     proto.Int64(int64(v.Ts)),
							TotalBetValue:   proto.Int64(int64(gnd.BaseResult.TotalBet)),
							TotalPriceValue: proto.Int64(gnd.BaseResult.WinTotal),
							Cards:           gnd.BaseResult.Cards,
						}
						pack.PlayerHistory = append(pack.PlayerHistory, player)
					}
				}
			}
			proto.SetDefaults(pack)
			logger.Info("MiniPokerPlayerHistory: ", pack)
			return pack

		}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
			if data == nil {
				logger.Error("MiniPokerPlayerHistory data is nil")
				return
			}
			p.SendToClient(int(proto_minipoker.MiniPokerPacketID_PACKET_SC_MINIPOKER_PLAYERHISTORY), data)
		}), "CSGetCandyPlayerHistoryHandler").Start()
	case MiniPokerPlayerSelBet:
		//参数是否合法
		if !common.InSliceInt32(sceneEx.DbGameFree.GetOtherIntParams(), int32(params[0])) {
			this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, proto_minipoker.OpResultCode_OPRC_Error, params)
			return false
		}
		playerEx.score = int32(params[0])
		playerEx.betIdx = int32(common.InSliceInt32Index(sceneEx.DbGameFree.GetOtherIntParams(), playerEx.score))
		jackpotVal := sceneEx.jackpot.JackpotFund[playerEx.betIdx]
		params = append(params, jackpotVal)
		this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, proto_minipoker.OpResultCode_OPRC_Sucess, params)
	}
	return true
}

func (this *SceneStateMiniPokerStart) OnPlayerEvent(s *base.Scene, p *base.Player, evtcode int, params []int64) {
}

//发送玩家操作情况
func (this *SceneStateMiniPokerStart) OnPlayerSToCOp(s *base.Scene, p *base.Player, pos int, opcode int,
	opRetCode proto_minipoker.OpResultCode, params []int64) {
	pack := &proto_minipoker.SCMiniPokerOp{
		SnId:      proto.Int32(p.SnId),
		OpCode:    proto.Int(opcode),
		OpRetCode: opRetCode,
		Params:    params,
	}
	proto.SetDefaults(pack)
	p.SendToClient(int(proto_minipoker.MiniPokerPacketID_PACKET_SC_MINIPOKER_PLAYEROP), pack)
}

var minipokerBenchTest bool
var minipokerBenchTestTimes int

func (this *SceneStateMiniPokerStart) BenchTest(s *base.Scene, p *base.Player) {
	const BENCH_CNT = 10000
	setting := base.CoinPoolMgr.GetCoinPoolSetting(s.Platform, s.GetGameFreeId(), s.GetGroupId())
	oldPoolCoin := base.CoinPoolMgr.LoadCoin(s.GetGameFreeId(), s.Platform, s.GetGroupId())
	if minipokerBenchTestTimes == 0 {
		defaultVal := int64(setting.GetLowerLimit())
		if oldPoolCoin != defaultVal {
			base.CoinPoolMgr.PushCoin(s.GetGameFreeId(), s.GetGroupId(), s.Platform, defaultVal-oldPoolCoin)
		}
	}
	minipokerBenchTestTimes++

	fileName := fmt.Sprintf("minipoker-%v-%d.csv", s.DbGameFree.GetSceneType(), minipokerBenchTestTimes)
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm)
	defer file.Close()
	if err != nil {
		file, err = os.Create(fileName)
		if err != nil {
			return
		}
	}
	file.WriteString("玩家id,当前水位,之前余额,之后余额,投入,产出,税收,小游戏,中线倍数,中线数,剩余免费次数\r\n")

	oldCoin := p.Coin
	p.Coin = int64(5000 * s.DbGameFree.GetBaseScore())
	if playerEx, ok := p.ExtraData.(*MiniPokerPlayerData); ok {
		for i := 0; i < BENCH_CNT; i++ {
			startCoin := p.Coin
			// freeTimes := playerEx.freeTimes
			poolCoin := base.CoinPoolMgr.LoadCoin(s.GetGameFreeId(), s.Platform, s.GetGroupId())
			playerEx.UnmarkFlag(base.PlayerState_GameBreak)
			suc := this.OnPlayerOp(s, p, MiniPokerPlayerOpStart, []int64{int64(playerEx.score)})
			inCoin := int64(playerEx.RollGameType.BaseResult.TotalBet)
			outCoin := playerEx.RollGameType.BaseResult.ChangeCoin + inCoin
			taxCoin := playerEx.RollGameType.BaseResult.Tax

			str := fmt.Sprintf("%v,%v,%v,%v,%v,%v,%v,%v\r\n", p.SnId, poolCoin, startCoin, p.Coin, inCoin, outCoin, taxCoin,
				playerEx.RollGameType.BaseResult.WinRate)
			file.WriteString(str)
			if !suc {
				break
			}
		}
	}
	p.Coin = oldCoin
}
func (this *SceneStateMiniPokerStart) WinTargetBenchTest(s *base.Scene, p *base.Player) {
	const BENCH_CNT = 10000
	var once = sync.Once{}
	once.Do(func() {
		setting := base.CoinPoolMgr.GetCoinPoolSetting(s.Platform, s.GetGameFreeId(), s.GetGroupId())
		oldPoolCoin := base.CoinPoolMgr.LoadCoin(s.GetGameFreeId(), s.Platform, s.GetGroupId())
		if minipokerBenchTestTimes == 0 {
			defaultVal := int64(setting.GetLowerLimit())
			if oldPoolCoin != defaultVal {
				base.CoinPoolMgr.PushCoin(s.GetGameFreeId(), s.GetGroupId(), s.Platform, defaultVal-oldPoolCoin)
			}
		}
	})
	minipokerBenchTestTimes++

	fileName := fmt.Sprintf("minipoker-win-%v-%d.csv", s.DbGameFree.GetSceneType(), minipokerBenchTestTimes)
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm)
	defer file.Close()
	if err != nil {
		file, err = os.Create(fileName)
		if err != nil {
			return
		}
	}
	file.WriteString("玩家id,当前水位,之前余额,之后余额,投入,产出,税收,小游戏,中线倍数,中线数,剩余免费次数\r\n")
	oldCoin := p.Coin
	switch s.DbGameFree.GetSceneType() {
	case 1:
		p.Coin = 100000
	case 2:
		p.Coin = 500000
	case 3:
		p.Coin = 1000000
	default:
		p.Coin = 100000
	}
	var targetCoin = p.Coin + p.Coin/10
	if playerEx, ok := p.ExtraData.(*MiniPokerPlayerData); ok {
		for i := 0; p.Coin < targetCoin; i++ {
			startCoin := p.Coin
			poolCoin := base.CoinPoolMgr.LoadCoin(s.GetGameFreeId(), s.Platform, s.GetGroupId())
			playerEx.UnmarkFlag(base.PlayerState_GameBreak)
			suc := this.OnPlayerOp(s, p, MiniPokerPlayerOpStart, []int64{int64(playerEx.score)})
			inCoin := int64(playerEx.RollGameType.BaseResult.TotalBet)
			outCoin := playerEx.RollGameType.BaseResult.ChangeCoin + inCoin
			taxCoin := playerEx.RollGameType.BaseResult.Tax

			str := fmt.Sprintf("%v,%v,%v,%v,%v,%v,%v,%v\r\n", p.SnId, poolCoin, startCoin, p.Coin, inCoin, outCoin, taxCoin,
				playerEx.RollGameType.BaseResult.WinRate)
			file.WriteString(str)
			if !suc {
				break
			}
			if i > BENCH_CNT {
				break
			}
		}
	}
	p.Coin = oldCoin
}

// MultiplayerBenchTest 多人同时测试 模拟正常环境
func (this *SceneStateMiniPokerStart) MultiplayerBenchTest(s *base.Scene) {
	const BENCH_CNT = 10000
	var once = sync.Once{}
	once.Do(func() {
		setting := base.CoinPoolMgr.GetCoinPoolSetting(s.Platform, s.GetGameFreeId(), s.GetGroupId())
		oldPoolCoin := base.CoinPoolMgr.LoadCoin(s.GetGameFreeId(), s.Platform, s.GetGroupId())
		if minipokerBenchTestTimes == 0 {
			defaultVal := int64(setting.GetLowerLimit())
			if oldPoolCoin != defaultVal {
				base.CoinPoolMgr.PushCoin(s.GetGameFreeId(), s.GetGroupId(), s.Platform, defaultVal-oldPoolCoin)
			}
		}
	})
	minipokerBenchTestTimes++

	fileName := fmt.Sprintf("minipoker-total-%v-%d.csv", s.DbGameFree.GetSceneType(), minipokerBenchTestTimes)
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm)
	defer file.Close()
	if err != nil {
		file, err = os.Create(fileName)
		if err != nil {
			return
		}
	}
	file.WriteString("玩家id,当前水位,之前余额,之后余额,投入,产出,税收,小游戏,爆奖,中线倍数,中线数,剩余免费次数\r\n")

	playersFile := make(map[int32]*os.File)
	oldCoins := make(map[int32]int64)
	hasCoin := int64(1000 * s.DbGameFree.GetBaseScore())
	robots := make(map[int32]bool)
	testPlayers := make(map[int32]*base.Player)
	for _, p := range s.Players {
		if p.IsRob {
			p.IsRob = false
			robots[p.SnId] = true
		}
		fileName := fmt.Sprintf("minipoker-player%v-%v-%d.csv", p.SnId, s.DbGameFree.GetSceneType(), minipokerBenchTestTimes)
		file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm)
		if err != nil {
			file, err = os.Create(fileName)
			if err != nil {
				return
			}
		}
		file.WriteString("玩家id,当前水位,之前余额,之后余额,投入,产出,税收,小游戏,爆奖,中线倍数,中线数,剩余免费次数\r\n")
		playersFile[p.SnId] = file
		oldCoins[p.SnId] = p.Coin
		p.Coin = hasCoin
		hasCoin = int64(float64(hasCoin) * 1.6)
		testPlayers[p.SnId] = p
	}
	defer func() {
		for _, file := range playersFile {
			file.Close()
		}
		for snid, coin := range oldCoins {
			if player := s.GetPlayer(snid); player != nil {
				player.Coin = coin
				if robots[player.SnId] {
					player.IsRob = true
				}
			}
		}
	}()

	totalBet := s.DbGameFree.GetBaseScore()
	for i := 0; i < BENCH_CNT; i++ {
		for snid, p := range testPlayers {
			if playerEx, ok := p.ExtraData.(*MiniPokerPlayerData); ok {
				startCoin := p.Coin
				if startCoin < int64(totalBet) {
					continue
				}
				// freeTimes := playerEx.freeTimes
				poolCoin := base.CoinPoolMgr.LoadCoin(s.GetGameFreeId(), s.Platform, s.GetGroupId())
				playerEx.UnmarkFlag(base.PlayerState_GameBreak)
				suc := this.OnPlayerOp(s, p, MiniPokerPlayerOpStart, []int64{int64(playerEx.score)})
				inCoin := int64(playerEx.RollGameType.BaseResult.TotalBet)
				outCoin := playerEx.RollGameType.BaseResult.ChangeCoin + inCoin
				taxCoin := playerEx.RollGameType.BaseResult.Tax
				lineScore := float64(playerEx.RollGameType.BaseResult.WinRate*s.DbGameFree.GetBaseScore()) * float64(10000.0-s.DbGameFree.GetTaxRate()) / 10000.0
				jackpotScore := outCoin - int64(lineScore+0.00001)

				str := fmt.Sprintf("%v,%v,%v,%v,%v,%v,%v,%v,%v\r\n", p.SnId, poolCoin, startCoin, p.Coin, inCoin, outCoin, taxCoin,
					jackpotScore, playerEx.RollGameType.BaseResult.WinRate)
				file.WriteString(str)
				if pFile := playersFile[snid]; pFile != nil {
					pFile.WriteString(str)
				}

				if !suc {
					continue
				}

				/* if playerEx.totalPriceBonus > 0 {
					this.OnPlayerOp(s, p, MiniPokerBonusGame, []int64{playerEx.spinID})
				} */
			}
		}
	}
}

func MiniPokerCheckAndSaveLog(sceneEx *MiniPokerSceneData, playerEx *MiniPokerPlayerData) {
	//统计金币变动
	//log1
	logger.Trace("MiniPokerCheckAndSaveLog Save ", playerEx.SnId)
	//changeCoin := playerEx.Coin - playerEx.StartCoin
	changeCoin := playerEx.winCoin - playerEx.taxCoin - playerEx.CurrentBet
	startCoin := playerEx.Coin - changeCoin
	playerEx.SaveSceneCoinLog(startCoin, changeCoin,
		playerEx.Coin, playerEx.CurrentBet, playerEx.taxCoin, playerEx.winCoin, playerEx.jackpotWinCoin, 0)

	//log2
	playerEx.RollGameType.BaseResult.BeforeCoin = startCoin
	playerEx.RollGameType.BaseResult.AfterCoin = playerEx.Coin
	playerEx.RollGameType.BaseResult.ChangeCoin = changeCoin
	playerEx.RollGameType.BaseResult.IsFirst = sceneEx.IsPlayerFirst(playerEx.Player)
	playerEx.RollGameType.BaseResult.Tax = playerEx.taxCoin
	playerEx.RollGameType.BaseResult.WBLevel = sceneEx.players[playerEx.SnId].WBLevel
	playerEx.RollGameType.BaseResult.BasicBet = playerEx.score
	playerEx.RollGameType.BaseResult.TotalBet = int32(playerEx.CurrentBet)
	playerEx.RollGameType.BaseResult.RoomId = int32(sceneEx.SceneId)
	playerEx.RollGameType.BaseResult.PlayerSnid = playerEx.SnId
	playerEx.RollGameType.UserName = playerEx.Name

	if playerEx.score > 0 {
		if !playerEx.IsRob {
			info, err := model.MarshalGameNoteByMini(playerEx.RollGameType)
			if err == nil {
				logid, _ := model.AutoIncGameLogId()
				sceneEx.SaveGameDetailedLog(logid, info, &base.GameDetailedParam{})
				totalin := int64(playerEx.RollGameType.BaseResult.TotalBet)
				totalout := playerEx.RollGameType.BaseResult.ChangeCoin + playerEx.taxCoin + totalin
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
						BetAmount:         int64(playerEx.RollGameType.BaseResult.TotalBet),
						WinAmountNoAnyTax: playerEx.RollGameType.BaseResult.ChangeCoin,
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
			Gain: proto.Int64(playerEx.RollGameType.BaseResult.ChangeCoin),
			Tax:  proto.Int64(playerEx.taxCoin),
		}
		gwPlayerBet := &proto_server.GWPlayerBet{
			GameFreeId: proto.Int32(sceneEx.DbGameFree.GetId()),
			RobotGain:  proto.Int64(-playerEx.RollGameType.BaseResult.ChangeCoin),
		}
		gwPlayerBet.PlayerBets = append(gwPlayerBet.PlayerBets, playerBet)
		proto.SetDefaults(gwPlayerBet)
		sceneEx.SendToWorld(int(proto_server.SSPacketID_PACKET_GW_PLAYERBET), gwPlayerBet)
		logger.Trace("Send msg gwPlayerBet ===>", gwPlayerBet)
	}

	playerEx.taxCoin = 0
	playerEx.winCoin = 0
	//playerEx.linesWinCoin = 0
	playerEx.jackpotWinCoin = 0

	if sceneEx.CheckNeedDestroy() {
		sceneEx.PlayerLeave(playerEx.Player, common.PlayerLeaveReason_OnDestroy, true)
	}
}

func init() {
	ScenePolicyMiniPokerSington.RegisteSceneState(&SceneStateMiniPokerStart{})
	core.RegisteHook(core.HOOK_BEFORE_START, func() error {
		base.RegisteScenePolicy(common.GameId_MiniPoker, 0, ScenePolicyMiniPokerSington)
		return nil
	})
}
