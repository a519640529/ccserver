package candy

import (
	"github.com/idealeak/goserver/core/timer"
	"strconv"
	"time"

	"games.yol.com/win88/common"
	"games.yol.com/win88/gamerule/candy"
	"games.yol.com/win88/minigame/base"
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	proto_candy "games.yol.com/win88/protocol/candy"
	proto_server "games.yol.com/win88/protocol/server"
	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/task"
)

////////////////////////////////////////////////////////////////////////////////
//糖果
////////////////////////////////////////////////////////////////////////////////

// 房间内主要逻辑
var ScenePolicyCandySington = &ScenePolicyCandy{}

type ScenePolicyCandy struct {
	base.BaseScenePolicy
	states [CandySceneStateMax]base.SceneState
}

//创建场景扩展数据
func (this *ScenePolicyCandy) CreateSceneExData(s *base.Scene) interface{} {
	sceneEx := NewCandySceneData(s)
	if sceneEx != nil {
		if sceneEx.init() {
			s.ExtraData = sceneEx
		}
	}
	return sceneEx
}

//创建玩家扩展数据
func (this *ScenePolicyCandy) CreatePlayerExData(s *base.Scene, p *base.Player) interface{} {
	playerEx := &CandyPlayerData{Player: p}
	if playerEx != nil {
		p.ExtraData = playerEx
	}
	return playerEx
}

//场景开启事件
func (this *ScenePolicyCandy) OnStart(s *base.Scene) {
	logger.Trace("(this *ScenePolicyCandy) OnStart, sceneId=", s.SceneId)
	sceneEx := NewCandySceneData(s)
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

			s.ExtraData = sceneEx
			s.ChangeSceneState(CandySceneStateStart) //改变当前的玩家状态
		}
	}
}

//场景关闭事件
func (this *ScenePolicyCandy) OnStop(s *base.Scene) {
	logger.Trace("(this *ScenePolicyCandy) OnStop , sceneId=", s.SceneId)
	if sceneEx, ok := s.ExtraData.(*CandySceneData); ok {
		if sceneEx.jackpotNoticeHandle != timer.TimerHandle(0) {
			timer.StopTimer(sceneEx.jackpotNoticeHandle)
		}
		sceneEx.SaveData(true)
	}
}

//场景心跳事件
func (this *ScenePolicyCandy) OnTick(s *base.Scene) {
	if s == nil {
		return
	}
	if s.SceneState != nil {
		s.SceneState.OnTick(s)
	}
}

//玩家进入事件
func (this *ScenePolicyCandy) OnPlayerEnter(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Trace("(this *ScenePolicyCandy) OnPlayerEnter, sceneId=", s.SceneId, " player=", p.SnId)
	if sceneEx, ok := s.ExtraData.(*CandySceneData); ok {
		playerEx := &CandyPlayerData{Player: p}
		playerEx.init(s) // 玩家当前信息初始化
		sceneEx.players[p.SnId] = playerEx
		p.ExtraData = playerEx
		CandySendRoomInfo(s, p, sceneEx, playerEx, nil)
		s.FirePlayerEvent(p, base.PlayerEventEnter, nil) //回调会调取 onPlayerEvent事件
	}
}

//玩家离开事件
func (this *ScenePolicyCandy) OnPlayerLeave(s *base.Scene, p *base.Player, reason int) {
	if s == nil || p == nil {
		return
	}
	logger.Trace("(this *ScenePolicyCandy) OnPlayerLeave, sceneId=", s.SceneId, " player=", p.SnId)
	if sceneEx, ok := s.ExtraData.(*CandySceneData); ok {
		if this.CanChangeCoinScene(s, p) {
			sceneEx.OnPlayerLeave(p, reason)
			s.FirePlayerEvent(p, base.PlayerEventLeave, []int64{int64(reason)})
		}
	}
}

//玩家掉线
func (this *ScenePolicyCandy) OnPlayerDropLine(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Trace("(this *ScenePolicyCandy) OnPlayerDropLine, sceneId=", s.SceneId, " player=", p.SnId)
	s.FirePlayerEvent(p, base.PlayerEventDropLine, nil)
	if sceneEx, ok := s.ExtraData.(*CandySceneData); ok {
		if sceneEx.Gaming {
			return
		}
	}
}

//玩家重连
func (this *ScenePolicyCandy) OnPlayerRehold(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Trace("(this *ScenePolicyCandy) OnPlayerRehold, sceneId=", s.SceneId, " player=", p.Name)
	// gs 修改 rehold 不再发送roomInfo 消息
	if _, ok := s.ExtraData.(*CandySceneData); ok {
		if _, ok := p.ExtraData.(*CandyPlayerData); ok {
			//发送房间信息给自己
			//CandySendRoomInfo(s, p, sceneEx, playerEx, sceneEx.billedData)
			s.FirePlayerEvent(p, base.PlayerEventRehold, nil)
		}
	}
}

//玩家重新返回房间
func (this *ScenePolicyCandy) OnPlayerReturn(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Trace("(this *ScenePolicyCandy) OnPlayerReeturn,sceneId=", s.SceneId, "player=", p.Name)
	if sceneEx, ok := s.ExtraData.(*CandySceneData); ok {
		if playerEx, ok := p.ExtraData.(*CandyPlayerData); ok {
			CandySendRoomInfo(s, p, sceneEx, playerEx, playerEx.billedData)
			s.FirePlayerEvent(p, base.PlayerReturn, nil)
		}
	}
}

//玩家操作
func (this *ScenePolicyCandy) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if s == nil || p == nil {
		return false
	}
	if s.SceneState != nil {
		return s.SceneState.OnPlayerOp(s, p, opcode, params)
	}
	return true
}

//玩家事件
func (this *ScenePolicyCandy) OnPlayerEvent(s *base.Scene, p *base.Player, evtcode int, params []int64) {
	if s == nil || p == nil {
		return
	}
	if s.SceneState != nil {
		s.SceneState.OnPlayerEvent(s, p, evtcode, params)
	}
}

//是否完成了整个牌局
func (this *ScenePolicyCandy) IsCompleted(s *base.Scene) bool { return false }

//是否可以强制开始
func (this *ScenePolicyCandy) IsCanForceStart(s *base.Scene) bool { return true }

//当前状态能否换桌
func (this *ScenePolicyCandy) CanChangeCoinScene(s *base.Scene, p *base.Player) bool {
	if s == nil || p == nil {
		return true
	}
	if s.SceneState != nil {
		return s.SceneState.CanChangeScene(s, p)
	}
	return true
}

func (this *ScenePolicyCandy) RegisteSceneState(state base.SceneState) {
	if state == nil {
		return
	}
	stateid := state.GetState()
	if stateid < 0 || stateid >= CandySceneStateMax {
		return
	}
	this.states[stateid] = state
}

func (this *ScenePolicyCandy) GetSceneState(s *base.Scene, stateid int) base.SceneState {
	if stateid >= 0 && stateid < CandySceneStateMax {
		return ScenePolicyCandySington.states[stateid]
	}
	return nil
}

func CandySendRoomInfo(s *base.Scene, p *base.Player, sceneEx *CandySceneData, playerEx *CandyPlayerData, billedData *proto_candy.GameBilledData) {
	logger.Trace("-------------------发送房间消息 ", s.SceneId, p.SnId)
	pack := &proto_candy.SCCandyRoomInfo{
		RoomId:     proto.Int(s.SceneId),
		Creator:    proto.Int32(0),
		GameId:     proto.Int(s.GetGameId()),
		RoomMode:   proto.Int(s.GetGameMode()),
		Params:     s.GetParams(),
		State:      proto.Int(s.SceneState.GetState()),
		Jackpot:    proto.Int64(sceneEx.jackpot.JackpotFund[playerEx.betIdx]),
		GameFreeId: proto.Int32(s.GetDBGameFree().GetId()),
	}
	if playerEx != nil {
		pd := &proto_candy.CandyPlayerData{
			SnId:        proto.Int32(playerEx.SnId),
			Name:        proto.String(playerEx.Name),
			Head:        proto.Int32(playerEx.Head),
			Sex:         proto.Int32(playerEx.Sex),
			Coin:        proto.Int64(playerEx.Coin),
			HeadOutLine: proto.Int32(playerEx.HeadOutLine),
			VIP:         proto.Int32(playerEx.VIP),
		}
		pack.Players = append(pack.Players, pd)
		for _, value := range playerEx.cards {
			pack.Cards = append(pack.Cards, int32(value))
		}
		pack.BetLines = playerEx.betLines
		pack.Chip = proto.Int32(playerEx.score)
		pack.ParamsEx = sceneEx.DbGameFree.GetOtherIntParams()
		pack.BilledData = billedData
		// pack.SpinID = proto.Int64(playerEx.spinID)
	}
	proto.SetDefaults(pack)
	p.SendToClient(int(proto_candy.CandyPacketID_PACKET_SC_CANDY_ROOMINFO), pack)
}

type SceneStateCandyStart struct {
}

//获取当前场景状态
func (this *SceneStateCandyStart) GetState() int { return CandySceneStateStart }

//是否可以切换状态到
func (this *SceneStateCandyStart) CanChangeTo(s base.SceneState) bool { return true }

//当前状态能否换桌
func (this *SceneStateCandyStart) CanChangeScene(s *base.Scene, p *base.Player) bool {
	if _, ok := p.ExtraData.(*CandyPlayerData); ok {
		return true
	}
	return true
}

func (this *SceneStateCandyStart) GetTimeout(s *base.Scene) int { return 0 }

func (this *SceneStateCandyStart) OnEnter(s *base.Scene) {
	if sceneEx, ok := s.ExtraData.(*CandySceneData); ok {
		logger.Tracef("(this *Scene) [%v] 场景状态进入 %v", s.SceneId, len(sceneEx.players))
		sceneEx.StateStartTime = time.Now()
		pack := &proto_candy.SCCandyRoomState{
			State: proto.Int(this.GetState()),
		}
		proto.SetDefaults(pack)
		s.Broadcast(int(proto_candy.CandyPacketID_PACKET_SC_CANDY_ROOMSTATE), pack, 0)
	}
}

func (this *SceneStateCandyStart) OnLeave(s *base.Scene) {}

func (this *SceneStateCandyStart) OnTick(s *base.Scene) {
	// 检查玩家是否满足被踢出条件，检查服务器是否需要关闭
	if sceneEx, ok := s.ExtraData.(*CandySceneData); ok {
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

func (this *SceneStateCandyStart) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	playerEx, ok := p.ExtraData.(*CandyPlayerData)
	if !ok {
		return false
	}
	sceneEx, ok := s.ExtraData.(*CandySceneData)
	if !ok {
		return false
	}
	if sceneEx.CheckNeedDestroy() {
		//离开有统计
		sceneEx.PlayerLeave(playerEx.Player, common.PlayerLeaveReason_OnDestroy, true)
		return false
	}

	p.LastOPTimer = time.Now()
	switch opcode {
	case CandyPlayerOpStart: //开始

		//if !candyBenchTest {
		//	candyBenchTest = true
		//	//for i := 0; i < 10; i++ {
		//	//	this.BenchTest(s, p)
		//	//	//this.WinTargetBenchTest(s, p)
		//	//}
		//	this.MultiplayerBenchTest(s)
		//	return true
		//}
		//参数是否合法
		//params 参数0底注，后面跟客户端选择的线n条线(1<=n<=20)，客户端线是从1开始算起1~20条线
		if len(params) < 2 || len(params) > candy.LINENUM+1 {
			this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, proto_candy.OpResultCode_OPRC_Error, params)
			return false
		}
		//先做底注校验
		if !common.InSliceInt32(sceneEx.DbGameFree.GetOtherIntParams(), int32(params[0])) {
			this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, proto_candy.OpResultCode_OPRC_Error, params)
			return false
		}
		//判断线条是否重复，是否合法
		lineFlag := make(map[int64]bool)
		lineParams := make([]int64, 0)
		for i := 1; i < len(params); i++ {
			lineNum := params[i]
			if lineNum >= 1 && lineNum <= int64(candy.LINENUM) && !lineFlag[lineNum] {
				lineParams = append(lineParams, lineNum)
				lineFlag[lineNum] = true
			} else {
				this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, proto_candy.OpResultCode_OPRC_Error, params)
				return false
			}
		}
		//没有选线参数
		if len(lineParams) == 0 {
			this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, proto_candy.OpResultCode_OPRC_Error, params)
			return false
		}

		playerEx.score = int32(params[0]) // 单线押注数
		playerEx.betIdx = int32(common.InSliceInt32Index(sceneEx.DbGameFree.GetOtherIntParams(), playerEx.score))
		//获取总投注金额（所有线的总投注） |  校验玩家余额是否足够
		totalBetValue := (int64(len(lineParams))) * int64(playerEx.score)
		logger.Logger.Infof("candy 开始转动 当前金币%v 总下注:%v", playerEx.Coin, totalBetValue)
		if totalBetValue > playerEx.Coin {
			this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, proto_candy.OpResultCode_OPRC_CoinNotEnough, params)
			return false
		} else if int64(sceneEx.DbGameFree.GetBetLimit()) > playerEx.Coin { //押注限制
			this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, proto_candy.OpResultCode_OPRC_CoinNotEnough, params)
			return false
		}

		sceneEx.GameNowTime = time.Now()
		sceneEx.NumOfGames++
		p.GameTimes++
		//playerEx.StartCoin = playerEx.Coin

		//获取当前水池的上下文环境
		sceneEx.CpCtx = base.CoinPoolMgr.GetCoinPoolCtx(sceneEx.GetPlatform(), sceneEx.GetGameFreeId(), sceneEx.GetGroupId())

		//税收比例
		taxRate := sceneEx.DbGameFree.GetTaxRate()
		if taxRate < 0 || taxRate > 10000 {
			logger.Tracef("CandyErrorTaxRate [%v][%v][%v]", sceneEx.GetGameFreeId(), playerEx.SnId, taxRate)
			taxRate = 500
		}
		//水池设置
		coinPoolSetting := base.CoinPoolMgr.GetCoinPoolSetting(sceneEx.GetPlatform(), sceneEx.GetGameFreeId(), sceneEx.GetGroupId())
		baseRate := coinPoolSetting.GetBaseRate() //基础赔率
		ctroRate := coinPoolSetting.GetCtroRate() //调节赔率 暗税系数
		//if baseRate >= 10000 || baseRate <= 0 || ctroRate < 0 || ctroRate >= 1000 || baseRate+ctroRate > 9900 {
		//	logger.Warnf("CandyErrorBaseRate [%v][%v][%v][%v]", sceneEx.GetGameFreeId(), playerEx.SnId, baseRate, ctroRate)
		//	baseRate = 9700
		//	ctroRate = 200
		//}
		//jackpotRate := 10000 - (baseRate + ctroRate) //奖池系数
		jackpotRate := ctroRate //奖池系数
		logger.Tracef("CandyRates [%v][%v][%v][%v][%v]", sceneEx.GetGameFreeId(), playerEx.SnId, taxRate, baseRate, ctroRate)

		gamePoolCoin := base.CoinPoolMgr.LoadCoin(sceneEx.GetGameFreeId(), sceneEx.GetPlatform(), sceneEx.GetGroupId()) // 当前水池金额
		prizeFund := gamePoolCoin - sceneEx.jackpot.JackpotFund[playerEx.betIdx]                                        // 除去奖池的水池剩余金额

		// 奖池参数
		var jackpotParam = sceneEx.DbGameFree.GetJackpot()
		var jackpotInit = int64(jackpotParam[candy.CANDY_JACKPOT_InitJackpot] * playerEx.score) //奖池初始值

		var jackpotFundAdd, prizeFundAdd int64

		//扣除投注金币
		p.AddCoin(-totalBetValue, common.GainWay_HundredSceneLost, true, false, "system", s.GetSceneName(), base.PlayerAddCoinDefRetryCnt, func(context *base.AddCoinContext, coin int64, ok bool) {
			if !ok {
				this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, proto_candy.OpResultCode_OPRC_CoinNotEnough, params)
				return
			}
			logger.Logger.Tracef("candy 扣钱事务成功 当前玩家金币:%v context:%v 回传金币:%v, %v", p.Coin, context.Player.Coin, context.Coin, coin)
			//p.Coin += context.Coin
			p.Coin = coin
			playerEx.betLines = lineParams                                                    // 选线记录
			jackpotFundAdd = int64(float64(totalBetValue) * (float64(jackpotRate) / 10000.0)) //奖池要增加的金额
			prizeFundAdd = int64(float64(totalBetValue) * (float64(baseRate) / 10000.0))
			sceneEx.jackpot.JackpotFund[playerEx.betIdx] += jackpotFundAdd //奖池增加
			playerEx.TotalBet += totalBetValue                             //总下注额（从进房间开始,包含多局游戏的下注）
			p.Statics(sceneEx.KeyGameId, sceneEx.KeyGamefreeId, -totalBetValue, true)
			if !p.IsRob && !sceneEx.Testing {
				// 推送金币
				base.CoinPoolMgr.PushCoin(sceneEx.GetGameFreeId(), sceneEx.GetGroupId(), sceneEx.GetPlatform(), int64(float64(totalBetValue)*float64(10000-ctroRate)/10000.0))
			}

			var symbolType candy.Symbol
			if prizeFund <= int64(coinPoolSetting.GetLowerLimit()) { // 水池不足
				symbolType = candy.SYMBOL1
			} else {
				symbolType = candy.SYMBOL2
			}
			writeBlackTryTimes := 0
		WriteBlack:
			slotData := make([]int, 0)
			var spinRes CandySpinResult
			var slotDataIsOk bool
			for i := 0; i < 3; i++ {
				slotData = candy.GenerateSlotsData_v2(symbolType)
				// if sceneEx.DbGameFree.GetSceneType() == 1 {
				// 	slotData = []int{5, 2, 6, 4, 4, 6, 5, 6, 6}
				// }
				spinRes = sceneEx.CalcLinePrize(slotData, playerEx.betLines, int64(playerEx.score))
				// if sceneEx.DbGameFree.GetSceneType() == 1 {
				// 	// slotDataIsOk = true
				// 	break
				// }
				// 水池不足以支付玩家
				spinCondition := prizeFund + prizeFundAdd - (spinRes.TotalPrizeJackpot + spinRes.TotalPrizeLine /* + spinRes.BonusGame.GetTotalPrizeValue() */)
				if spinRes.IsJackpot {
					spinCondition += sceneEx.jackpot.JackpotFund[playerEx.betIdx] - jackpotInit
				}
				if spinCondition <= 0 {
					if !spinRes.IsJackpot {
						writeBlackTryTimes = 999
						break
					}
					continue
				}

				// 非爆奖时 大奖限制
				//var limitBigWin int64
				//if symbolType == candy.SYMBOL1 {
				//	limitBigWin = int64(jackpotParam[candy.CANDY_JACKPOT_LIMITWIN_PRIZELOW])
				//} else {
				//	limitBigWin = int64(jackpotParam[candy.CANDY_JACKPOT_LIMITWIN_PRIZEHIGH_ROOM])
				//}
				//if totalBetValue > 0 && !spinRes.IsJackpot && spinRes.TotalPrizeLine > totalBetValue*limitBigWin {
				//	continue
				//}
				slotDataIsOk = true
				break
			}

			if !slotDataIsOk {
				slotData = candy.GenerateSlotsData_v3()
				spinRes = sceneEx.CalcLinePrize(slotData, playerEx.betLines, int64(playerEx.score))
			}

			// 黑白名单调控 防止异常循环，添加上限次数
			if writeBlackTryTimes < 100 && playerEx.CheckBlackWriteList(spinRes.TotalPrizeLine+spinRes.TotalPrizeJackpot > totalBetValue) {
				writeBlackTryTimes++
				goto WriteBlack
			} else if writeBlackTryTimes >= 100 && writeBlackTryTimes != 999 {
				logger.Warnf("CandyWriteBlackTryTimesOver [%v][%v][%v][%v][%v]", sceneEx.GetGameFreeId(), playerEx.SnId, gamePoolCoin, playerEx.BlackLevel, playerEx.WhiteLevel)
			}

			//if playerEx.debugJackpot {
			//	spinRes = sceneEx.CalcLinePrize(candy.DebugSlotData, playerEx.betLines, int64(playerEx.score))
			//	playerEx.debugJackpot = false
			//}

			if spinRes.IsJackpot {
				sceneEx.jackpot.JackpotFund[playerEx.betIdx] = jackpotInit // 如果爆奖池，奖池需要初始化
			}

			// 玩家赢钱
			totalWinScore := spinRes.TotalPrizeLine + spinRes.TotalPrizeJackpot
			if totalWinScore >= 0 {
				p.AddCoin(totalWinScore, common.GainWay_HundredSceneWin, false, false, "system", s.GetSceneName(), base.PlayerAddCoinDefRetryCnt, func(context *base.AddCoinContext, coin int64, b bool) {
					logger.Logger.Tracef("赢钱事务成功 当前玩家金币:%v context:%v 回传金币:%v b:%v", context, p.Coin, coin, b)
					//if b {
					//	p.Coin = coin
					//}
					//p.Coin += context.Coin
					p.Coin = coin
					p.Statics(sceneEx.KeyGameId, sceneEx.KeyGamefreeId, totalWinScore+spinRes.TotalTaxScore, true)
					if !p.IsRob && !sceneEx.Testing {
						base.CoinPoolMgr.PopCoin(sceneEx.GetGameFreeId(), sceneEx.GetGroupId(), sceneEx.GetPlatform(), totalWinScore+spinRes.TotalTaxScore)
					}
					playerEx.taxCoin = spinRes.TotalTaxScore
					playerEx.AddServiceFee(playerEx.taxCoin)
					this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, proto_candy.OpResultCode_OPRC_Sucess, append(params[:1], playerEx.betLines...))

					// candy.SpinID++
					// playerEx.spinID = candy.SpinID
					playerEx.cards = spinRes.SlotsData
					playerEx.winCoin = totalWinScore + spinRes.TotalTaxScore
					playerEx.jackpotWinCoin = spinRes.TotalPrizeJackpot
					playerEx.CurrentBet = totalBetValue
					playerEx.CurrentTax = playerEx.taxCoin

					billedData := &proto_candy.GameBilledData{
						SlotsData:              spinRes.SlotsData,
						IsJackpot:              proto.Bool(spinRes.IsJackpot),
						PrizeLines:             spinRes.LinesInfo,
						TotalPrizeValue:        proto.Int64(totalWinScore),                                //玩家总赢取
						TotalPaylinePrizeValue: proto.Int64(spinRes.TotalPrizeLine),                       //玩家线奖励
						TotalJackpotValue:      proto.Int64(spinRes.TotalPrizeJackpot),                    //玩家奖池奖励
						Balance:                proto.Int64(playerEx.Coin),                                //玩家最终金额
						Jackpot:                proto.Int64(sceneEx.jackpot.JackpotFund[playerEx.betIdx]), //奖池最终金额
						LuckyData:              spinRes.LuckyData,
					}
					playerEx.billedData = billedData
					pack := &proto_candy.SCCandyGameBilled{
						//SpinID:                 proto.Int64(playerEx.spinID),
						BilledData: billedData,
					}
					proto.SetDefaults(pack)
					p.SendToClient(int(proto_candy.CandyPacketID_PACKET_SC_CANDY_GAMEBILLED), pack)
					logger.Logger.Infof("发送结算信息 当前金币%v msg:%v", playerEx.Coin, pack)
					// 记录本次操作this
					playerEx.RollGameType.BaseResult.WinTotal = billedData.GetTotalPrizeValue()
					playerEx.RollGameType.BaseResult.AllWinNum = int32(len(billedData.PrizeLines))
					playerEx.RollGameType.BaseResult.WinRate = spinRes.TotalWinRate
					playerEx.RollGameType.BaseResult.Cards = billedData.GetSlotsData()
					playerEx.RollGameType.BaseResult.WinLineScore = billedData.GetTotalPaylinePrizeValue()
					playerEx.RollGameType.WinLines = spinRes.WinLines
					playerEx.RollGameType.BaseResult.WinJackpot = billedData.GetTotalJackpotValue()
					CandyCheckAndSaveLog(sceneEx, playerEx)

					// 广播奖池
					if totalBetValue == 0 && !spinRes.IsJackpot { // 没改变奖池
						return
					}
					// 添加进开奖记录里面
					if spinRes.IsJackpot {
						spinid := strconv.FormatInt(int64(playerEx.SnId), 10)
						// 推送最新开奖记录到world
						msg := &proto_server.GWGameNewBigWinHistory{
							SceneId: proto.Int32(int32(sceneEx.SceneId)),
							BigWinHistory: &proto_server.BigWinHistoryInfo{
								SpinID:      proto.String(spinid),
								CreatedTime: proto.Int64(time.Now().Unix()),
								BaseBet:     proto.Int64(int64(playerEx.score)),
								TotalBet:    proto.Int64(int64(playerEx.CurrentBet)),
								PriceValue:  proto.Int64(int64(billedData.GetTotalJackpotValue())),
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
				})
			}
		})
	case CandyPlayerHistory:
		task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
			spinid := strconv.FormatInt(int64(playerEx.SnId), 10)
			gpl := model.GetPlayerListByHallEx(p.SnId, p.Platform, 0, 50, 0, 0, s.DbGameFree.SceneType, s.DbGameFree.GetGameClass(), s.GameId)
			pack := &proto_candy.SCCandyPlayerHistory{}
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
				data, err := UnMarshalCandyGameNote(gdl.GameDetailedNote)
				if err != nil {
					logger.Errorf("UnMarshalCandyGameNote error:%v", err)
				}
				if gnd, ok := data.(*GameResultLog); ok {
					if gnd.BaseResult != nil {
						player := &proto_candy.CandyPlayerHistoryInfo{
							SpinID:          proto.String(spinid),
							CreatedTime:     proto.Int64(int64(v.Ts)),
							TotalBetValue:   proto.Int64(int64(gnd.BaseResult.TotalBet)),
							TotalPriceValue: proto.Int64(gnd.BaseResult.WinTotal),
						}
						pack.PlayerHistory = append(pack.PlayerHistory, player)
					}
				}
			}
			proto.SetDefaults(pack)
			logger.Info("CandyPlayerHistory: ", pack)
			return pack

		}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
			if data == nil {
				logger.Error("CandyPlayerHistory data is nil")
				return
			}
			p.SendToClient(int(proto_candy.CandyPacketID_PACKET_SC_CANDY_PLAYERHISTORY), data)
		}), "CSGetCandyPlayerHistoryHandler").Start()
	case CandyPlayerSelBet: //修改下注筹码
		//参数是否合法
		//先做底注校验
		if !common.InSliceInt32(sceneEx.DbGameFree.GetOtherIntParams(), int32(params[0])) {
			this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, proto_candy.OpResultCode_OPRC_Error, params)
			return false
		}
		playerEx.score = int32(params[0]) // 单线押注数
		playerEx.betIdx = int32(common.InSliceInt32Index(sceneEx.DbGameFree.GetOtherIntParams(), playerEx.score))
		jackpotVal := sceneEx.jackpot.JackpotFund[playerEx.betIdx]
		params = append(params, jackpotVal)
		this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, proto_candy.OpResultCode_OPRC_Sucess, params)
	}
	return true
}

func (this *SceneStateCandyStart) OnPlayerEvent(s *base.Scene, p *base.Player, evtcode int, params []int64) {
}

//发送玩家操作情况
func (this *SceneStateCandyStart) OnPlayerSToCOp(s *base.Scene, p *base.Player, pos int, opcode int,
	opRetCode proto_candy.OpResultCode, params []int64) {
	pack := &proto_candy.SCCandyOp{
		SnId:      proto.Int32(p.SnId),
		OpCode:    proto.Int(opcode),
		OpRetCode: opRetCode,
		Params:    params,
	}
	proto.SetDefaults(pack)
	p.SendToClient(int(proto_candy.CandyPacketID_PACKET_SC_CANDY_PLAYEROP), pack)
}

//var candyBenchTest bool
//var candyBenchTestTimes int
//
//func (this *SceneStateCandyStart) BenchTest(s *base.Scene, p *base.Player) {
//	const BENCH_CNT = 10000
//	setting := base.CoinPoolMgr.GetCoinPoolSetting(s.GetPlatform(), s.GetGameFreeId(), s.GetGroupId())
//	oldPoolCoin := base.CoinPoolMgr.LoadCoin(s.GetGameFreeId(), s.GetPlatform(), s.GetGroupId())
//	if candyBenchTestTimes == 0 {
//		defaultVal := int64(setting.GetLowerLimit())
//		if oldPoolCoin != defaultVal {
//			base.CoinPoolMgr.PushCoin(s.GetGameFreeId(), s.GetGroupId(), s.GetPlatform(), defaultVal-oldPoolCoin)
//		}
//	}
//	candyBenchTestTimes++
//
//	fileName := fmt.Sprintf("candy-%v-%d.csv", s.DbGameFree.GetSceneType(), candyBenchTestTimes)
//	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm)
//	defer file.Close()
//	if err != nil {
//		file, err = os.Create(fileName)
//		if err != nil {
//			return
//		}
//	}
//	file.WriteString("玩家id,当前水位,之前余额,之后余额,投入,产出,税收,小游戏,中线倍数,中线数,剩余免费次数\r\n")
//
//	oldCoin := p.Coin
//	p.Coin = int64(5000 * s.DbGameFree.GetBaseScore())
//	if playerEx, ok := p.ExtraData.(*CandyPlayerData); ok {
//		for i := 0; i < BENCH_CNT; i++ {
//			startCoin := p.Coin
//			// freeTimes := playerEx.freeTimes
//			poolCoin := base.CoinPoolMgr.LoadCoin(s.GetGameFreeId(), s.GetPlatform(), s.GetGroupId())
//			playerEx.UnmarkFlag(base.PlayerState_GameBreak)
//			suc := this.OnPlayerOp(s, p, CandyPlayerOpStart, append([]int64{int64(playerEx.score)}, candy.AllBetLines...))
//			inCoin := int64(playerEx.RollGameType.Score)
//			outCoin := playerEx.RollGameType.ChangeCoin + inCoin
//			taxCoin := playerEx.RollGameType.Tax
//
//			str := fmt.Sprintf("%v,%v,%v,%v,%v,%v,%v,%v,%v\r\n", p.SnId, poolCoin, startCoin, p.Coin, inCoin, outCoin, taxCoin,
//				playerEx.RollGameType.WinScore, playerEx.RollGameType.AllWinNum)
//			file.WriteString(str)
//			if !suc {
//				break
//			}
//		}
//	}
//	p.Coin = oldCoin
//}
//func (this *SceneStateCandyStart) WinTargetBenchTest(s *base.Scene, p *base.Player) {
//	const BENCH_CNT = 10000
//	var once = sync.Once{}
//	once.Do(func() {
//		setting := base.CoinPoolMgr.GetCoinPoolSetting(s.GetPlatform(), s.GetGameFreeId(), s.GetGroupId())
//		oldPoolCoin := base.CoinPoolMgr.LoadCoin(s.GetGameFreeId(), s.GetPlatform(), s.GetGroupId())
//		if candyBenchTestTimes == 0 {
//			defaultVal := int64(setting.GetLowerLimit())
//			if oldPoolCoin != defaultVal {
//				base.CoinPoolMgr.PushCoin(s.GetGameFreeId(), s.GetGroupId(), s.GetPlatform(), defaultVal-oldPoolCoin)
//			}
//		}
//	})
//	candyBenchTestTimes++
//
//	fileName := fmt.Sprintf("candy-win-%v-%d.csv", s.DbGameFree.GetSceneType(), candyBenchTestTimes)
//	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm)
//	defer file.Close()
//	if err != nil {
//		file, err = os.Create(fileName)
//		if err != nil {
//			return
//		}
//	}
//	file.WriteString("玩家id,当前水位,之前余额,之后余额,投入,产出,税收,小游戏,中线倍数,中线数,剩余免费次数\r\n")
//	oldCoin := p.Coin
//	switch s.DbGameFree.GetSceneType() {
//	case 1:
//		p.Coin = 100000
//	case 2:
//		p.Coin = 500000
//	case 3:
//		p.Coin = 1000000
//	default:
//		p.Coin = 100000
//	}
//	var targetCoin = p.Coin + p.Coin/10
//	if playerEx, ok := p.ExtraData.(*CandyPlayerData); ok {
//		for i := 0; p.Coin < targetCoin; i++ {
//			startCoin := p.Coin
//			poolCoin := base.CoinPoolMgr.LoadCoin(s.GetGameFreeId(), s.GetPlatform(), s.GetGroupId())
//			playerEx.UnmarkFlag(base.PlayerState_GameBreak)
//			suc := this.OnPlayerOp(s, p, CandyPlayerOpStart, append([]int64{int64(playerEx.score)}, candy.AllBetLines...))
//			inCoin := int64(playerEx.RollGameType.Score)
//			outCoin := playerEx.RollGameType.ChangeCoin + inCoin
//			taxCoin := playerEx.RollGameType.Tax
//
//			str := fmt.Sprintf("%v,%v,%v,%v,%v,%v,%v,%v,%v\r\n", p.SnId, poolCoin, startCoin, p.Coin, inCoin, outCoin, taxCoin,
//				playerEx.RollGameType.WinScore, playerEx.RollGameType.AllWinNum)
//			file.WriteString(str)
//			if !suc {
//				break
//			}
//			if i > BENCH_CNT {
//				break
//			}
//		}
//	}
//	p.Coin = oldCoin
//}

//// MultiplayerBenchTest 多人同时测试 模拟正常环境
//func (this *SceneStateCandyStart) MultiplayerBenchTest(s *base.Scene) {
//	const BENCH_CNT = 10000
//	var once = sync.Once{}
//	once.Do(func() {
//		setting := base.CoinPoolMgr.GetCoinPoolSetting(s.GetPlatform(), s.GetGameFreeId(), s.GetGroupId())
//		oldPoolCoin := base.CoinPoolMgr.LoadCoin(s.GetGameFreeId(), s.GetPlatform(), s.GetGroupId())
//		if candyBenchTestTimes == 0 {
//			defaultVal := int64(setting.GetLowerLimit())
//			if oldPoolCoin != defaultVal {
//				base.CoinPoolMgr.PushCoin(s.GetGameFreeId(), s.GetGroupId(), s.GetPlatform(), defaultVal-oldPoolCoin)
//			}
//		}
//	})
//	candyBenchTestTimes++
//
//	fileName := fmt.Sprintf("candy-total-%v-%d.csv", s.DbGameFree.GetSceneType(), candyBenchTestTimes)
//	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm)
//	defer file.Close()
//	if err != nil {
//		file, err = os.Create(fileName)
//		if err != nil {
//			return
//		}
//	}
//	file.WriteString("玩家id,当前水位,之前余额,之后余额,投入,产出,税收,小游戏,爆奖,中线倍数,中线数,剩余免费次数\r\n")
//
//	playersFile := make(map[int32]*os.File)
//	oldCoins := make(map[int32]int64)
//	hasCoin := int64(1000 * s.DbGameFree.GetBaseScore())
//	robots := make(map[int32]bool)
//	testPlayers := make(map[int32]*base.Player)
//	for _, p := range s.Players {
//		if p.IsRob {
//			p.IsRob = false
//			robots[p.SnId] = true
//		}
//		fileName := fmt.Sprintf("candy-player%v-%v-%d.csv", p.SnId, s.DbGameFree.GetSceneType(), candyBenchTestTimes)
//		file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm)
//		if err != nil {
//			file, err = os.Create(fileName)
//			if err != nil {
//				return
//			}
//		}
//		file.WriteString("玩家id,当前水位,之前余额,之后余额,投入,产出,税收,小游戏,爆奖,中线倍数,中线数,剩余免费次数\r\n")
//		playersFile[p.SnId] = file
//		oldCoins[p.SnId] = p.Coin
//		p.Coin = hasCoin
//		hasCoin = int64(float64(hasCoin) * 1.6)
//		testPlayers[p.SnId] = p
//	}
//	defer func() {
//		for _, file := range playersFile {
//			file.Close()
//		}
//		for snid, coin := range oldCoins {
//			if player := s.GetPlayer(snid); player != nil {
//				player.Coin = coin
//				if robots[player.SnId] {
//					player.IsRob = true
//				}
//			}
//		}
//	}()
//
//	totalBet := s.DbGameFree.GetBaseScore() * int32(len(candy.AllBetLines))
//	for i := 0; i < BENCH_CNT; i++ {
//		for snid, p := range testPlayers {
//			if playerEx, ok := p.ExtraData.(*CandyPlayerData); ok {
//				startCoin := p.Coin
//				if startCoin < int64(totalBet) {
//					continue
//				}
//				// freeTimes := playerEx.freeTimes
//				poolCoin := base.CoinPoolMgr.LoadCoin(s.GetGameFreeId(), s.GetPlatform(), s.GetGroupId())
//				playerEx.UnmarkFlag(base.PlayerState_GameBreak)
//				suc := this.OnPlayerOp(s, p, CandyPlayerOpStart, append([]int64{int64(playerEx.score)}, candy.AllBetLines...))
//				inCoin := int64(playerEx.RollGameType.Score)
//				outCoin := playerEx.RollGameType.ChangeCoin + inCoin
//				taxCoin := playerEx.RollGameType.Tax
//				lineScore := float64(playerEx.RollGameType.WinScore*s.DbGameFree.GetBaseScore()) * float64(10000.0-s.DbGameFree.GetTaxRate()) / 10000.0
//				jackpotScore := outCoin - int64(lineScore+0.00001)
//
//				str := fmt.Sprintf("%v,%v,%v,%v,%v,%v,%v,%v,%v,%v\r\n", p.SnId, poolCoin, startCoin, p.Coin, inCoin, outCoin, taxCoin,
//					jackpotScore, playerEx.RollGameType.WinScore, playerEx.RollGameType.AllWinNum)
//				file.WriteString(str)
//				if pFile := playersFile[snid]; pFile != nil {
//					pFile.WriteString(str)
//				}
//
//				if !suc {
//					continue
//				}
//
//				/* if playerEx.totalPriceBonus > 0 {
//					this.OnPlayerOp(s, p, CandyBonusGame, []int64{playerEx.spinID})
//				} */
//			}
//		}
//	}
//}

func CandyCheckAndSaveLog(sceneEx *CandySceneData, playerEx *CandyPlayerData) {
	//统计金币变动
	//log1
	logger.Trace("CandyCheckAndSaveLog Save ", playerEx.SnId)
	//changeCoin := playerEx.Coin - playerEx.StartCoin
	changeCoin := playerEx.winCoin - playerEx.taxCoin - playerEx.CurrentBet
	startCoin := playerEx.Coin - changeCoin
	playerEx.SaveSceneCoinLog(startCoin, changeCoin,
		playerEx.Coin, playerEx.CurrentBet, playerEx.taxCoin, playerEx.winCoin, playerEx.jackpotWinCoin, 0 /*  playerEx.smallGameWinCoin */)

	//log2
	playerEx.RollGameType.BaseResult.ChangeCoin = changeCoin
	playerEx.RollGameType.BaseResult.BasicBet = playerEx.score
	playerEx.RollGameType.BaseResult.RoomId = int32(sceneEx.SceneId)
	playerEx.RollGameType.BaseResult.AfterCoin = playerEx.Coin
	playerEx.RollGameType.BaseResult.BeforeCoin = startCoin
	playerEx.RollGameType.BaseResult.IsFirst = sceneEx.IsPlayerFirst(playerEx.Player)
	playerEx.RollGameType.BaseResult.PlayerSnid = playerEx.SnId
	playerEx.RollGameType.BaseResult.TotalBet = int32(playerEx.CurrentBet)
	playerEx.RollGameType.AllLine = int32(len(playerEx.betLines))
	// playerEx.RollGameType.FreeTimes = playerEx.freeTimes
	playerEx.RollGameType.UserName = playerEx.Name
	playerEx.RollGameType.BetLines = playerEx.betLines
	playerEx.RollGameType.BaseResult.Tax = playerEx.taxCoin
	playerEx.RollGameType.BaseResult.WBLevel = playerEx.WBLevel
	if playerEx.score > 0 {
		if !playerEx.IsRob {
			info, err := model.MarshalGameNoteByMini(playerEx.RollGameType)
			if err == nil {
				logid, _ := model.AutoIncGameLogId()
				// playerEx.currentLogId = logid
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
	playerEx.jackpotWinCoin = 0
	// playerEx.smallGameWinCoin = 0

	if sceneEx.CheckNeedDestroy() /* && playerEx.freeTimes <= 0 */ {
		sceneEx.PlayerLeave(playerEx.Player, common.PlayerLeaveReason_OnDestroy, true)
	}
}

func init() {
	ScenePolicyCandySington.RegisteSceneState(&SceneStateCandyStart{})
	core.RegisteHook(core.HOOK_BEFORE_START, func() error {
		base.RegisteScenePolicy(common.GameId_Candy, 0, ScenePolicyCandySington)
		return nil
	})
}
