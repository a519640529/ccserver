package tamquoc

import (
	"strconv"
	"time"

	"games.yol.com/win88/common"
	rule "games.yol.com/win88/gamerule/tamquoc"
	"games.yol.com/win88/gamesrv/base"
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/server"
	"games.yol.com/win88/protocol/tamquoc"
	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/task"
	"github.com/idealeak/goserver/core/timer"
)

////////////////////////////////////////////////////////////////////////////////
//百战成神
////////////////////////////////////////////////////////////////////////////////

// 房间内主要逻辑
var ScenePolicyTamQuocSington = &ScenePolicyTamQuoc{}

type ScenePolicyTamQuoc struct {
	base.BaseScenePolicy
	states [TamQuocSceneStateMax]base.SceneState
}

//创建场景扩展数据
func (this *ScenePolicyTamQuoc) CreateSceneExData(s *base.Scene) interface{} {
	sceneEx := NewTamQuocSceneData(s)
	if sceneEx != nil {
		if sceneEx.init() {
			s.ExtraData = sceneEx
		}
	}
	return sceneEx
}

//创建玩家扩展数据
func (this *ScenePolicyTamQuoc) CreatePlayerExData(s *base.Scene, p *base.Player) interface{} {
	playerEx := &TamQuocPlayerData{Player: p}
	if playerEx != nil {
		p.ExtraData = playerEx
	}
	return playerEx
}

//场景开启事件
func (this *ScenePolicyTamQuoc) OnStart(s *base.Scene) {
	logger.Trace("(this *ScenePolicyTamQuoc) OnStart, sceneId=", s.SceneId)
	sceneEx := NewTamQuocSceneData(s)
	if sceneEx != nil {
		if sceneEx.init() {
			sceneEx.BroadcastJackpot()
			s.ExtraData = sceneEx
			s.ChangeSceneState(TamQuocSceneStateStart) //改变当前的玩家状态
		}
	}
}

//场景关闭事件
func (this *ScenePolicyTamQuoc) OnStop(s *base.Scene) {
	logger.Trace("(this *ScenePolicyTamQuoc) OnStop , sceneId=", s.SceneId)
	if sceneEx, ok := s.ExtraData.(*TamQuocSceneData); ok {
		sceneEx.SaveData(true)
	}
}

//场景心跳事件
func (this *ScenePolicyTamQuoc) OnTick(s *base.Scene) {
	if s == nil {
		return
	}
	if s.SceneState != nil {
		s.SceneState.OnTick(s)
	}
}

//玩家进入事件
func (this *ScenePolicyTamQuoc) OnPlayerEnter(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Trace("(this *ScenePolicyTamQuoc) OnPlayerEnter, sceneId=", s.SceneId, " player=", p.SnId)
	if sceneEx, ok := s.ExtraData.(*TamQuocSceneData); ok {
		playerEx := &TamQuocPlayerData{Player: p}
		playerEx.init(s)                                   // 玩家当前信息初始化
		playerEx.score = sceneEx.DbGameFree.GetBaseScore() // 底注
		sceneEx.players[p.SnId] = playerEx
		p.ExtraData = playerEx
		TamQuocSendRoomInfo(s, p, sceneEx, playerEx, nil)
		s.FirePlayerEvent(p, base.PlayerEventEnter, nil) //回调会调取 onPlayerEvent事件
	}
}

//玩家离开事件
func (this *ScenePolicyTamQuoc) OnPlayerLeave(s *base.Scene, p *base.Player, reason int) {
	if s == nil || p == nil {
		return
	}
	logger.Trace("(this *ScenePolicyTamQuoc) OnPlayerLeave, sceneId=", s.SceneId, " player=", p.SnId)
	if sceneEx, ok := s.ExtraData.(*TamQuocSceneData); ok {
		if this.CanChangeCoinScene(s, p) {
			if playerEx, ok := p.ExtraData.(*TamQuocPlayerData); ok {
				playerEx.SavePlayerGameData(strconv.Itoa(int(s.GetGameFreeId())))
			}
			sceneEx.OnPlayerLeave(p, reason)
			s.FirePlayerEvent(p, base.PlayerEventLeave, []int64{int64(reason)})
		}
	}
}

//玩家掉线
func (this *ScenePolicyTamQuoc) OnPlayerDropLine(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Trace("(this *ScenePolicyTamQuoc) OnPlayerDropLine, sceneId=", s.SceneId, " player=", p.SnId)
	s.FirePlayerEvent(p, base.PlayerEventDropLine, nil)
	if sceneEx, ok := s.ExtraData.(*TamQuocSceneData); ok {
		if sceneEx.Gaming {
			return
		}
	}
}

//玩家重连
func (this *ScenePolicyTamQuoc) OnPlayerRehold(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Trace("(this *ScenePolicyTamQuoc) OnPlayerRehold, sceneId=", s.SceneId, " player=", p.Name)
	//if sceneEx, ok := s.ExtraData.(*TamQuocSceneData); ok {
	//	if playerEx, ok := p.ExtraData.(*TamQuocPlayerData); ok {
	//		//发送房间信息给自己
	//		TamQuocSendRoomInfo(s, p, sceneEx, playerEx)
	s.FirePlayerEvent(p, base.PlayerEventRehold, nil)
	//	}
	//}
}

//玩家重连
func (this *ScenePolicyTamQuoc) OnPlayerReturn(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Trace("(this *ScenePolicyTamQuoc) OnPlayerReturn, sceneId=", s.SceneId, " player=", p.Name)
	if sceneEx, ok := s.ExtraData.(*TamQuocSceneData); ok {
		if playerEx, ok := p.ExtraData.(*TamQuocPlayerData); ok {
			//发送房间信息给自己
			TamQuocSendRoomInfo(s, p, sceneEx, playerEx, playerEx.billedData)
			s.FirePlayerEvent(p, base.PlayerEventReturn, nil)
		}
	}
}

//玩家操作
func (this *ScenePolicyTamQuoc) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if s == nil || p == nil {
		return false
	}
	if s.SceneState != nil {
		return s.SceneState.OnPlayerOp(s, p, opcode, params)
	}
	return true
}

//玩家事件
func (this *ScenePolicyTamQuoc) OnPlayerEvent(s *base.Scene, p *base.Player, evtcode int, params []int64) {
	if s == nil || p == nil {
		return
	}
	if s.SceneState != nil {
		s.SceneState.OnPlayerEvent(s, p, evtcode, params)
	}
}

//是否完成了整个牌局
func (this *ScenePolicyTamQuoc) IsCompleted(s *base.Scene) bool { return false }

//是否可以强制开始
func (this *ScenePolicyTamQuoc) IsCanForceStart(s *base.Scene) bool { return true }

//当前状态能否换桌
func (this *ScenePolicyTamQuoc) CanChangeCoinScene(s *base.Scene, p *base.Player) bool {
	if s == nil || p == nil {
		return true
	}
	if s.SceneState != nil {
		return s.SceneState.CanChangeCoinScene(s, p)
	}
	return true
}

func (this *ScenePolicyTamQuoc) RegisteSceneState(state base.SceneState) {
	if state == nil {
		return
	}
	stateid := state.GetState()
	if stateid < 0 || stateid >= TamQuocSceneStateMax {
		return
	}
	this.states[stateid] = state
}

func (this *ScenePolicyTamQuoc) GetSceneState(s *base.Scene, stateid int) base.SceneState {
	if stateid >= 0 && stateid < TamQuocSceneStateMax {
		return ScenePolicyTamQuocSington.states[stateid]
	}
	return nil
}

func TamQuocSendRoomInfo(s *base.Scene, p *base.Player, sceneEx *TamQuocSceneData, playerEx *TamQuocPlayerData, data *tamquoc.GameBilledData) {
	logger.Trace("-------------------发送房间消息 ", s.RoomId, p.SnId)
	pack := &tamquoc.SCTamQuocRoomInfo{
		RoomId:     proto.Int(s.SceneId),
		Creator:    proto.Int32(s.Creator),
		GameId:     proto.Int(s.GameId),
		RoomMode:   proto.Int(s.GameMode),
		Params:     s.Params,
		State:      proto.Int(s.SceneState.GetState()),
		Jackpot:    proto.Int64(sceneEx.jackpot.JackpotFund),
		GameFreeId: proto.Int32(s.GetDBGameFree().GetId()),
		BilledData: data,
	}
	if playerEx != nil {
		pd := &tamquoc.TamQuocPlayerData{
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
		pack.BetLines = playerEx.betLines
		pack.FreeTimes = proto.Int32(playerEx.freeTimes)
		pack.Chip = proto.Int32(s.DbGameFree.BaseScore)
		pack.SpinID = proto.Int64(playerEx.spinID)
		if playerEx.totalPriceBonus > 0 && playerEx.bonusGameStartTime.Add(TamQuocBonusGamePickTime).Before(time.Now()) {
			playerEx.totalPriceBonus = 0
			playerEx.bonusGamePickPos = make([]int32, 2)
			// 取消定时器
			if playerEx.bonusTimerHandle != timer.TimerHandle(0) {
				timer.StopTimer(playerEx.bonusTimerHandle)
				playerEx.bonusTimerHandle = timer.TimerHandle(0)
			}
		}
		if playerEx.totalPriceBonus > 0 {
			pack.TotalPriceBonus = proto.Int64(playerEx.totalPriceBonus)
			playerEx.bonusGamePickPos[1] = int32(playerEx.bonusGameStartTime.Add(TamQuocBonusGamePickTime).Unix() - time.Now().Unix())
			pack.ParamsEx = playerEx.bonusGamePickPos
			pack.BonusGame = &playerEx.bonusGame
		}
	}
	proto.SetDefaults(pack)
	p.SendToClient(int(tamquoc.TamQuocPacketID_PACKET_SC_TAMQUOC_ROOMINFO), pack)
}

type SceneStateTamQuocStart struct {
}

//获取当前场景状态
func (this *SceneStateTamQuocStart) GetState() int { return TamQuocSceneStateStart }

//是否可以切换状态到
func (this *SceneStateTamQuocStart) CanChangeTo(s base.SceneState) bool { return true }

//当前状态能否换桌
func (this *SceneStateTamQuocStart) CanChangeCoinScene(s *base.Scene, p *base.Player) bool {
	if _, ok := p.ExtraData.(*TamQuocPlayerData); ok {
		return true
	}
	return true
}

func (this *SceneStateTamQuocStart) GetTimeout(s *base.Scene) int { return 0 }

func (this *SceneStateTamQuocStart) OnEnter(s *base.Scene) {
	if sceneEx, ok := s.ExtraData.(*TamQuocSceneData); ok {
		logger.Tracef("(this *base.Scene) [%v] 场景状态进入 %v", s.SceneId, len(sceneEx.players))
		sceneEx.StateStartTime = time.Now()
		pack := &tamquoc.SCTamQuocRoomState{
			State: proto.Int(this.GetState()),
		}
		proto.SetDefaults(pack)
		s.Broadcast(int(tamquoc.TamQuocPacketID_PACKET_SC_TAMQUOC_ROOMSTATE), pack, 0)
	}
}

func (this *SceneStateTamQuocStart) OnLeave(s *base.Scene) {}

func (this *SceneStateTamQuocStart) OnTick(s *base.Scene) {
	if time.Now().Sub(s.GameStartTime) > time.Second*3 {
		if sceneEx, ok := s.ExtraData.(*TamQuocSceneData); ok {
			for _, p := range sceneEx.players {
				if p.IsOnLine() {
					p.leavetime = 0
					continue
				}
				p.leavetime++
				if p.leavetime < 60 {
					continue
				}
				//踢出玩家
				sceneEx.PlayerLeave(p.Player, common.PlayerLeaveReason_LongTimeNoOp, true)
			}
		}
		s.GameStartTime = time.Now()
	}

	if sceneEx, ok := s.ExtraData.(*TamQuocSceneData); ok {
		for _, p := range sceneEx.players {
			//游戏次数达到目标值
			todayGamefreeIDSceneData, _ := p.GetDaliyGameData(int(sceneEx.DbGameFree.GetId()))
			if !p.IsRob &&
				todayGamefreeIDSceneData != nil &&
				sceneEx.DbGameFree.GetPlayNumLimit() != 0 &&
				todayGamefreeIDSceneData.GameTimes >= int64(sceneEx.DbGameFree.GetPlayNumLimit()) {
				s.PlayerLeave(p.Player, common.PlayerLeaveReason_GameTimes, true)
			}
		}
		if sceneEx.CheckNeedDestroy() {
			for _, player := range sceneEx.players {
				if !player.IsRob {
					if time.Now().Sub(player.LastOPTimer) > 10*time.Second {
						//离开有统计
						sceneEx.PlayerLeave(player.Player, common.PlayerLeaveReason_OnDestroy, true)
					}
				}
			}
			if s.GetRealPlayerCnt() == 0 {
				sceneEx.SceneDestroy(true)
			}
		}
	}
}

func (this *SceneStateTamQuocStart) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	playerEx, ok := p.ExtraData.(*TamQuocPlayerData)
	if !ok {
		return false
	}
	sceneEx, ok := s.ExtraData.(*TamQuocSceneData)
	if !ok {
		return false
	}
	if sceneEx.CheckNeedDestroy() && playerEx.freeTimes <= 0 {
		//离开有统计
		sceneEx.PlayerLeave(playerEx.Player, common.PlayerLeaveReason_OnDestroy, true)
		return false
	}
	switch opcode {
	case TamQuocPlayerOpStart: //开始

		//参数是否合法
		//params 参数0底注，后面跟客户端选择的线n条线(1<=n<=20)，客户端线是从1开始算起1~20条线
		if len(params) < 2 || len(params) > rule.LINENUM+1 {
			this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, tamquoc.OpResultCode_OPRC_Error, params)
			return false
		}
		//先做底注校验
		if sceneEx.DbGameFree.GetBaseScore() != int32(params[0]) {
			this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, tamquoc.OpResultCode_OPRC_Error, params)
			return false
		}
		playerEx.score = int32(params[0]) // 单线押注数
		// 小游戏未结束 不能进行下一次旋转
		if playerEx.totalPriceBonus > 0 {
			this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, tamquoc.OpResultCode_OPRC_Error, params)
			return false
		}
		//判断线条是否重复，是否合法
		lineFlag := make(map[int64]bool)
		lineParams := make([]int64, 0)
		for i := 1; i < len(params); i++ {
			lineNum := params[i]
			if lineNum >= 1 && lineNum <= int64(rule.LINENUM) && !lineFlag[lineNum] {
				lineParams = append(lineParams, lineNum)
				lineFlag[lineNum] = true
			} else {
				this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, tamquoc.OpResultCode_OPRC_Error, params)
				return false
			}
		}
		//没有选线参数
		if len(lineParams) == 0 {
			this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, tamquoc.OpResultCode_OPRC_Error, params)
			return false
		}

		//获取总投注金额（所有线的总投注） |  校验玩家余额是否足够
		totalBetValue := (int64(len(lineParams))) * params[0]
		if playerEx.freeTimes <= 0 && totalBetValue > playerEx.Coin {
			this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, tamquoc.OpResultCode_OPRC_CoinNotEnough, params)
			return false
		} else if playerEx.freeTimes <= 0 && int64(sceneEx.DbGameFree.GetBetLimit()) > playerEx.Coin { //押注限制
			this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, tamquoc.OpResultCode_OPRC_CoinNotEnough, params)
			return false
		}

		p.LastOPTimer = time.Now()
		sceneEx.GameNowTime = time.Now()
		sceneEx.NumOfGames++
		p.GameTimes++
		//playerEx.StartCoin = playerEx.Coin

		//获取当前水池的上下文环境
		sceneEx.CpCtx = base.CoinPoolMgr.GetCoinPoolCtx(sceneEx.Platform, sceneEx.GetGameFreeId(), sceneEx.GroupId)

		//税收比例
		taxRate := sceneEx.DbGameFree.GetTaxRate()
		if taxRate < 0 || taxRate > 10000 {
			logger.Tracef("TamQuocErrorTaxRate [%v][%v][%v][%v]", sceneEx.GetGameFreeId(), playerEx.SnId, playerEx.spinID, taxRate)
			taxRate = 500
		}
		//水池设置
		coinPoolSetting := base.CoinPoolMgr.GetCoinPoolSetting(sceneEx.Platform, sceneEx.GetGameFreeId(), sceneEx.GroupId)
		baseRate := coinPoolSetting.GetBaseRate() //基础赔率
		ctroRate := coinPoolSetting.GetCtroRate() //调节赔率 暗税系数
		//if baseRate >= 10000 || baseRate <= 0 || ctroRate < 0 || ctroRate >= 1000 || baseRate+ctroRate > 9900 {
		//	logger.Warnf("TamQuocErrorBaseRate [%v][%v][%v][%v][%v]", sceneEx.GetGameFreeId(), playerEx.SnId, playerEx.spinID, baseRate, ctroRate)
		//	baseRate = 9700
		//	ctroRate = 200
		//}
		//jackpotRate := 10000 - (baseRate + ctroRate) //奖池系数
		jackpotRate := ctroRate //奖池系数
		logger.Tracef("TamQuocRates [%v][%v][%v][%v][%v][%v]", sceneEx.GetGameFreeId(), playerEx.SnId, playerEx.spinID, taxRate, baseRate, ctroRate)

		gamePoolCoin := base.CoinPoolMgr.LoadCoin(sceneEx.GetGameFreeId(), sceneEx.Platform, sceneEx.GroupId) // 当前水池金额
		prizeFund := gamePoolCoin - sceneEx.jackpot.JackpotFund                                               // 除去奖池的水池剩余金额

		// 奖池参数
		var jackpotParam = sceneEx.DbGameFree.GetJackpot()
		var jackpotInit = int64(jackpotParam[rule.TAMQUOC_JACKPOT_InitJackpot] * sceneEx.DbGameFree.GetBaseScore()) //奖池初始值

		var jackpotFundAdd, prizeFundAdd int64
		if playerEx.freeTimes <= 0 { //正常模式才能记录用户的押注变化，免费模式不能改变押注
			playerEx.betLines = lineParams                                                    // 选线记录
			jackpotFundAdd = int64(float64(totalBetValue) * (float64(jackpotRate) / 10000.0)) //奖池要增加的金额
			prizeFundAdd = int64(float64(totalBetValue) * (float64(baseRate) / 10000.0))
			sceneEx.jackpot.JackpotFund += jackpotFundAdd //奖池增加
			playerEx.TotalBet += totalBetValue            //总下注额（从进房间开始,包含多局游戏的下注）
			//扣除投注金币
			p.AddCoin(-totalBetValue, common.GainWay_HundredSceneLost, base.SyncFlag_ToClient, "system", s.GetSceneName())
			p.Statics(sceneEx.KeyGameId, sceneEx.KeyGamefreeId, -totalBetValue, true)
			if !p.IsRob && !sceneEx.Testing {
				// 推送金币
				base.CoinPoolMgr.PushCoin(sceneEx.GetGameFreeId(), sceneEx.GroupId, sceneEx.Platform, int64(float64(totalBetValue)*(float64(10000-ctroRate)/10000.0)))
			}

			//统计参与游戏次数
			//if !sceneEx.Testing && !playerEx.IsRob {
			//	pack := &server.GWSceneEnd{
			//		GameFreeId: proto.Int32(sceneEx.DbGameFree.GetId()),
			//		Players:    []*server.PlayerCtx{&server.PlayerCtx{SnId: proto.Int32(playerEx.SnId), Coin: proto.Int64(playerEx.Coin)}},
			//	}
			//	proto.SetDefaults(pack)
			//	sceneEx.SendToWorld(int(server.SSPacketID_PACKET_GW_SCENEEND), pack)
			//}
		} else { //免费次数时，不能改线改选线
			totalBetValue = 0
		}

		var symbolType rule.Symbol
		if prizeFund <= int64(coinPoolSetting.GetLowerLimit()) { // 水池不足
			symbolType = rule.SYMBOL1
		} else {
			symbolType = rule.SYMBOL2
		}
		writeBlackTryTimes := 0
	WriteBlack:
		slotData := make([]int, 0)
		var spinRes TamQuocSpinResult
		var slotDataIsOk bool
		for i := 0; i < 3; i++ {
			slotData = rule.GenerateSlotsData_v2(symbolType)
			//if sceneEx.DbGameFree.GetSceneType() == 1 {
			//	slotData = []int{1, 1, 1, 1, 1, 6, 6, 6, 6, 6, 7, 7, 7, 7, 7}
			//}
			spinRes = sceneEx.CalcLinePrize(slotData, playerEx.betLines, params[0])
			if sceneEx.DbGameFree.GetSceneType() == 1 {
				slotDataIsOk = true
				break
			}
			if spinRes.AddFreeTimes > 0 && len(playerEx.betLines) < 15 {
				continue
			}

			// 免费次数时 不允许爆奖
			if playerEx.freeTimes > 0 && spinRes.IsJackpot {
				continue
			}

			//if spinRes.IsJackpot && len(playerEx.betLines) < 20 {  // todo 限制指定用户爆奖
			//	continue
			//}

			// 水池不足以支付玩家
			spinCondition := prizeFund + prizeFundAdd - (spinRes.TotalPrizeJackpot + spinRes.TotalPrizeLine + spinRes.BonusGame.GetTotalPrizeValue())
			if spinRes.IsJackpot {
				spinCondition += sceneEx.jackpot.JackpotFund - jackpotInit
			}
			if spinCondition <= 0 {
				if !spinRes.IsJackpot {
					writeBlackTryTimes = 999
					break
				}
				continue
			}

			// 非爆奖时 大奖限制
			var limitBigWin int64
			if symbolType == rule.SYMBOL1 {
				limitBigWin = int64(jackpotParam[rule.TAMQUOC_JACKPOT_LIMITWIN_PRIZELOW])

			} else {
				limitBigWin = int64(jackpotParam[rule.TAMQUOC_JACKPOT_LIMITWIN_PRIZEHIGH])
			}
			if totalBetValue > 0 && !spinRes.IsJackpot && spinRes.TotalPrizeLine > totalBetValue*limitBigWin {
				continue
			}

			if spinRes.BonusGame.GetTotalPrizeValue() > 0 {
				if len(playerEx.betLines) < 20 {
					continue
				}

				// 小游戏最小时间间隔限制
				lastBonusGameTimeInterval := time.Now().Unix() - playerEx.bonusGameTime
				if symbolType == rule.SYMBOL1 && lastBonusGameTimeInterval < int64(jackpotParam[rule.TAMQUOC_JACKPOT_BonusMinTimeInterval_Low]) {
					continue
				} else if symbolType == rule.SYMBOL2 && lastBonusGameTimeInterval < int64(jackpotParam[rule.TAMQUOC_JACKPOT_BonusMinTimeInterval_High]) {
					continue
				}
			}

			slotDataIsOk = true
			break
		}

		if !slotDataIsOk {
			slotData = rule.GenerateSlotsData_v3(rule.DEFALUTROOMMODEL)
			spinRes = sceneEx.CalcLinePrize(slotData, playerEx.betLines, params[0])
		}

		// 黑白名单调控 防止异常循环，添加上限次数
		if writeBlackTryTimes < 100 && playerEx.CheckBlackWriteList(spinRes.TotalPrizeLine+spinRes.TotalPrizeJackpot+spinRes.BonusGame.GetTotalPrizeValue() > totalBetValue) {
			writeBlackTryTimes++
			goto WriteBlack
		} else if writeBlackTryTimes >= 100 && writeBlackTryTimes != 999 {
			logger.Warnf("TamquocWriteBlackTryTimesOver [%v][%v][%v][%v][%v]", sceneEx.GetGameFreeId(), playerEx.SnId, gamePoolCoin, playerEx.BlackLevel, playerEx.WhiteLevel)
		}
		//if playerEx.DebugGame && sceneEx.SceneType == 1 {
		//	if playerEx.TestNum >= len(DebugData) {
		//		playerEx.TestNum = 0
		//	}
		//	slotData = DebugData[playerEx.TestNum]
		//	spinRes = sceneEx.CalcLinePrize(slotData, playerEx.betLines, params[0])
		//	playerEx.TestNum++
		//}
		// 奖池水池处理
		if spinRes.IsJackpot {
			sceneEx.jackpot.JackpotFund = jackpotInit
		}

		// 玩家赢钱
		totalWinScore := spinRes.TotalPrizeLine + spinRes.TotalPrizeJackpot
		if totalWinScore > 0 || spinRes.BonusGame.GetTotalPrizeValue() > 0 {
			p.AddCoin(totalWinScore+spinRes.BonusGame.GetTotalPrizeValue(), common.GainWay_HundredSceneWin, 0, "system", s.GetSceneName())
			p.Statics(sceneEx.KeyGameId, sceneEx.KeyGamefreeId, totalWinScore+spinRes.BonusGame.GetTotalPrizeValue()+spinRes.TotalTaxScore, true)
			if !p.IsRob && !sceneEx.Testing {
				base.CoinPoolMgr.PopCoin(sceneEx.GetGameFreeId(), sceneEx.GroupId, sceneEx.Platform, totalWinScore+spinRes.BonusGame.GetTotalPrizeValue()+spinRes.TotalTaxScore)
			}
			playerEx.taxCoin = spinRes.TotalTaxScore
			playerEx.AddServiceFee(playerEx.taxCoin)
		}

		this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, tamquoc.OpResultCode_OPRC_Sucess, append(params[:1], playerEx.betLines...))

		//免费次数
		var isFreeFlag bool
		if playerEx.freeTimes > 0 {
			playerEx.freeTimes--
			isFreeFlag = true
		}
		playerEx.freeTimes += spinRes.AddFreeTimes

		rule.SpinID++
		playerEx.spinID = rule.SpinID
		playerEx.cards = spinRes.SlotsData
		playerEx.winCoin = spinRes.TotalPrizeLine + spinRes.TotalPrizeJackpot + spinRes.BonusGame.GetTotalPrizeValue() + playerEx.taxCoin
		playerEx.linesWinCoin = spinRes.TotalPrizeLine
		playerEx.jackpotWinCoin = spinRes.TotalPrizeJackpot
		playerEx.smallGameWinCoin = spinRes.BonusGame.GetTotalPrizeValue()
		playerEx.CurrentBet = totalBetValue
		playerEx.CurrentTax = playerEx.taxCoin

		// 小游戏超时处理
		if spinRes.IsBonusGame {
			playerEx.totalPriceBonus = spinRes.BonusGame.GetTotalPrizeValue()
			playerEx.bonusGameTime = time.Now().Unix()
			logger.Tracef("BonusGame Start [%v][%v][%v][%v]", sceneEx.GetGameFreeId(), playerEx.SnId, playerEx.spinID, playerEx.totalPriceBonus)
			playerEx.bonusTimerHandle, _ = timer.StartTimer(timer.TimerActionWrapper(func(h timer.TimerHandle, ud interface{}) bool {
				this.OnPlayerOp(s, p, TamQuocBonusGame, []int64{playerEx.spinID})
				return true
			}), nil, TamQuocBonusGameTimeout, 1)
			playerEx.bonusGameCanPickNum = len(spinRes.BonusGame.BonusData)
			playerEx.bonusGame = spinRes.BonusGame
		}

		playerEx.billedData = &tamquoc.GameBilledData{
			SpinID:                 proto.Int64(playerEx.spinID),
			SlotsData:              spinRes.SlotsData,
			AddFreeSpin:            proto.Int32(spinRes.AddFreeTimes),
			IsJackpot:              proto.Bool(spinRes.IsJackpot),
			PrizeLines:             spinRes.LinesInfo,
			TotalPrizeValue:        proto.Int64(totalWinScore),
			TotalPaylinePrizeValue: proto.Int64(spinRes.TotalPrizeLine),
			TotalJackpotValue:      proto.Int64(spinRes.TotalPrizeJackpot),
			Balance:                proto.Int64(playerEx.Coin - spinRes.BonusGame.GetTotalPrizeValue()),
			FreeSpins:              proto.Int32(playerEx.freeTimes),
			Jackpot:                proto.Int64(sceneEx.jackpot.JackpotFund),
			BonusGame:              &spinRes.BonusGame,
		}
		pack := &tamquoc.SCTamQuocGameBilled{
			BilledData: playerEx.billedData,
		}
		proto.SetDefaults(pack)
		logger.Logger.Infof("TamQuocPlayerOpStart %v", pack)
		p.SendToClient(int(tamquoc.TamQuocPacketID_PACKET_SC_TAMQUOC_GAMEBILLED), pack)

		// 记录本次操作this
		playerEx.RollGameType.BaseResult.WinTotal = pack.BilledData.GetTotalPrizeValue() + pack.BilledData.GetBonusGame().GetTotalPrizeValue()
		playerEx.RollGameType.BaseResult.IsFree = isFreeFlag
		playerEx.RollGameType.BaseResult.WinSmallGame = pack.BilledData.BonusGame.GetTotalPrizeValue()
		playerEx.RollGameType.BaseResult.AllWinNum = int32(len(pack.BilledData.PrizeLines))
		playerEx.RollGameType.BaseResult.WinRate = spinRes.TotalWinRate
		playerEx.RollGameType.BaseResult.Cards = pack.BilledData.GetSlotsData()
		playerEx.RollGameType.WinLines = spinRes.WinLines
		playerEx.RollGameType.BaseResult.WinJackpot = pack.BilledData.GetTotalJackpotValue()
		playerEx.RollGameType.BaseResult.WinLineScore = pack.BilledData.TotalPaylinePrizeValue
		TamQuocCheckAndSaveLog(sceneEx, playerEx)

		// 广播奖池
		if totalBetValue == 0 && !spinRes.IsJackpot { // 没改变奖池
			return true
		}
		// 添加进开奖记录里面
		if spinRes.IsJackpot {
			spinid := strconv.FormatInt(int64(playerEx.SnId), 10)
			// 推送最新开奖记录到world
			msg := &server.GWGameNewBigWinHistory{
				SceneId: proto.Int32(int32(sceneEx.SceneId)),
				BigWinHistory: &server.BigWinHistoryInfo{
					SpinID:      proto.String(spinid),
					CreatedTime: proto.Int64(time.Now().Unix()),
					BaseBet:     proto.Int64(int64(playerEx.score)),
					TotalBet:    proto.Int64(int64(playerEx.CurrentBet)),
					PriceValue:  proto.Int64(int64(pack.BilledData.GetTotalJackpotValue())),
					UserName:    proto.String(playerEx.Name),
				},
			}
			proto.SetDefaults(msg)
			logger.Logger.Infof("GWGameNewBigWinHistory gameFreeID %v %v", sceneEx.DbGameFree.GetId(), msg)
			sceneEx.SendToWorld(int(server.SSPacketID_PACKET_GW_GAMENEWBIGWINHISTORY), msg)
		} else {
			sceneEx.PushVirtualDataToWorld() // 推送虚拟数据
		}
		sceneEx.BroadcastJackpot()
	case TamQuocPlayerHistory:
		task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
			spinid := strconv.FormatInt(int64(playerEx.SnId), 10)
			gpl := model.GetPlayerListByHallEx(p.SnId, p.Platform, 0, 50, 0, 0, 0, s.DbGameFree.GetGameClass(), s.GameId)
			pack := &tamquoc.SCTamQuocPlayerHistory{}
			for _, v := range gpl.Data {
				if v.GameDetailedLogId == "" {
					logger.Logger.Error("TamQuocPlayerHistory GameDetailedLogId is nil")
					break
				}
				gdl := model.GetPlayerHistory(p.Platform, v.GameDetailedLogId)
				if gdl == nil {
					logger.Logger.Error("TamQuocPlayerHistory gdl is nil")
					continue
				}
				data, err := UnMarshalTamQuocGameNote(gdl.GameDetailedNote)
				if err != nil {
					logger.Logger.Errorf("UnMarshalTamQuoccGameNote error:%v", err)
				}
				gnd := data.(*GameResultLog)
				player := &tamquoc.TamQuocPlayerHistoryInfo{
					SpinID:          proto.String(spinid),
					CreatedTime:     proto.Int64(int64(v.Ts)),
					TotalBetValue:   proto.Int64(int64(gnd.BaseResult.TotalBet)),
					TotalPriceValue: proto.Int64(gnd.BaseResult.WinTotal),
					IsFree:          proto.Bool(gnd.BaseResult.IsFree),
					TotalBonusValue: proto.Int64(gnd.BaseResult.WinSmallGame),
				}
				pack.PlayerHistory = append(pack.PlayerHistory, player)
			}
			proto.SetDefaults(pack)
			logger.Logger.Info("TamQuocPlayerHistory: ", pack)
			return pack

		}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
			if data == nil {
				logger.Logger.Error("TamQuocPlayerHistory data is nil")
				return
			}
			p.SendToClient(int(tamquoc.TamQuocPacketID_PACKET_SC_TAMQUOC_PLAYERHISTORY), data)
		}), "CSGetTamQuocPlayerHistoryHandler").Start()
	case TamQuocBonusGame:
		//params 参数0 spinID
		//参数是否合法
		if len(params) < 1 {
			this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, tamquoc.OpResultCode_OPRC_Error, params)
			return false
		}
		if playerEx.spinID != params[0] || playerEx.totalPriceBonus <= 0 {
			this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, tamquoc.OpResultCode_OPRC_Error, params)
			return false
		}
		if playerEx.bonusTimerHandle != timer.TimerHandle(0) {
			timer.StopTimer(playerEx.bonusTimerHandle)
			playerEx.bonusTimerHandle = timer.TimerHandle(0)
		}
		logger.Tracef("BonusGame Start [%v][%v][%v][%v]", sceneEx.GetGameFreeId(), playerEx.SnId, playerEx.spinID, playerEx.totalPriceBonus)

		this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, tamquoc.OpResultCode_OPRC_Sucess, []int64{playerEx.totalPriceBonus, playerEx.Coin})
		playerEx.totalPriceBonus = 0
		playerEx.bonusGamePickPos = make([]int32, 2)
	case TamQuocBonusGameRecord:
		// params
		// 参数下标0  0表示进小游戏 1表示选图标界面
		// 参数下标1 玩家当前点击位置
		if (len(params) == 1 && params[0] != 0) ||
			(len(params) == 2 && (params[0] != 1 || params[1] < 1 || params[1] > 24)) {
			logger.Logger.Errorf("Invalid parameter")
			this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, tamquoc.OpResultCode_OPRC_Error, params)
			return true
		}
		switch params[0] {
		case 0:
			if playerEx.bonusGamePickPos[0] < 1 {
				// 记录进入小游戏时间
				playerEx.bonusGameStartTime = time.Now()
				playerEx.bonusGamePickPos[0] = 1
			}
		case 1:
			if len(playerEx.bonusGamePickPos) > playerEx.bonusGameCanPickNum {
				logger.Logger.Errorf("too many pickPos")
				this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, tamquoc.OpResultCode_OPRC_Error, params)
				return true
			}
			playerEx.bonusGamePickPos = append(playerEx.bonusGamePickPos, int32(params[1]))
		default:
			logger.Logger.Errorf("Invalid parameter, params[0]")
		}

	}
	return true
}

func (this *SceneStateTamQuocStart) OnPlayerEvent(s *base.Scene, p *base.Player, evtcode int, params []int64) {
}

//发送玩家操作情况
func (this *SceneStateTamQuocStart) OnPlayerSToCOp(s *base.Scene, p *base.Player, pos int, opcode int,
	opRetCode tamquoc.OpResultCode, params []int64) {
	pack := &tamquoc.SCTamQuocOp{
		SnId:      proto.Int32(p.SnId),
		OpCode:    proto.Int(opcode),
		OpRetCode: opRetCode,
		Params:    params,
	}
	proto.SetDefaults(pack)
	p.SendToClient(int(tamquoc.TamQuocPacketID_PACKET_SC_TAMQUOC_PLAYEROP), pack)
}

func TamQuocCheckAndSaveLog(sceneEx *TamQuocSceneData, playerEx *TamQuocPlayerData) {
	//统计金币变动
	//log1
	logger.Trace("TamQuocCheckAndSaveLog Save ", playerEx.SnId)
	//changeCoin := playerEx.Coin - playerEx.StartCoin
	changeCoin := playerEx.winCoin - playerEx.taxCoin - playerEx.CurrentBet
	startCoin := playerEx.Coin - changeCoin
	playerEx.SaveSceneCoinLog(startCoin, changeCoin,
		playerEx.Coin, playerEx.CurrentBet, playerEx.taxCoin, playerEx.winCoin, playerEx.jackpotWinCoin, playerEx.smallGameWinCoin)

	//log2
	playerEx.RollGameType.BaseResult.ChangeCoin = changeCoin
	playerEx.RollGameType.BaseResult.BasicBet = sceneEx.DbGameFree.GetBaseScore()
	playerEx.RollGameType.BaseResult.RoomId = int32(sceneEx.SceneId)
	playerEx.RollGameType.BaseResult.AfterCoin = playerEx.Coin
	playerEx.RollGameType.BaseResult.BeforeCoin = startCoin
	playerEx.RollGameType.BaseResult.IsFirst = sceneEx.IsPlayerFirst(playerEx.Player)
	playerEx.RollGameType.BaseResult.PlayerSnid = playerEx.SnId
	playerEx.RollGameType.BaseResult.TotalBet = int32(playerEx.CurrentBet)
	playerEx.RollGameType.AllLine = int32(len(playerEx.betLines))
	playerEx.RollGameType.BaseResult.FreeTimes = playerEx.freeTimes
	playerEx.RollGameType.UserName = playerEx.Name
	playerEx.RollGameType.BetLines = playerEx.betLines
	playerEx.RollGameType.BaseResult.Tax = playerEx.taxCoin
	playerEx.RollGameType.BaseResult.WBLevel = sceneEx.players[playerEx.SnId].WBLevel
	if playerEx.score > 0 {
		if !playerEx.IsRob {
			info, err := model.MarshalGameNoteByROLL(playerEx.RollGameType)
			if err == nil {
				logid, _ := model.AutoIncGameLogId()
				playerEx.currentLogId = logid
				sceneEx.SaveGameDetailedLog(logid, info, &base.GameDetailedParam{})
				totalin := int64(playerEx.RollGameType.BaseResult.TotalBet)
				totalout := playerEx.RollGameType.BaseResult.ChangeCoin + playerEx.taxCoin + totalin
				validFlow := totalin + totalout
				validBet := common.AbsI64(totalin - totalout)
				sceneEx.SaveGamePlayerListLog(playerEx.SnId,
					base.GetSaveGamePlayerListLogParam(playerEx.Platform, playerEx.Channel, playerEx.BeUnderAgentCode,
						playerEx.PackageID, logid, playerEx.InviterId, totalin, totalout, playerEx.taxCoin, 0,
						int64(playerEx.RollGameType.BaseResult.TotalBet), playerEx.RollGameType.BaseResult.ChangeCoin, validFlow, validBet, sceneEx.IsPlayerFirst(playerEx.Player),
						false))
			}
		}
	}

	//统计输下注金币数
	if !sceneEx.Testing && !playerEx.IsRob {
		playerBet := &server.PlayerBet{
			SnId:       proto.Int32(playerEx.SnId),
			Bet:        proto.Int64(playerEx.CurrentBet),
			Gain:       proto.Int64(playerEx.RollGameType.BaseResult.ChangeCoin),
			Tax:        proto.Int64(playerEx.taxCoin),
			Coin:       proto.Int64(playerEx.GetCoin()),
			GameCoinTs: proto.Int64(playerEx.GameCoinTs),
		}
		gwPlayerBet := &server.GWPlayerBet{
			SceneId:    proto.Int(sceneEx.SceneId),
			GameFreeId: proto.Int32(sceneEx.DbGameFree.GetId()),
			RobotGain:  proto.Int64(-playerEx.RollGameType.BaseResult.ChangeCoin),
		}
		gwPlayerBet.PlayerBets = append(gwPlayerBet.PlayerBets, playerBet)
		proto.SetDefaults(gwPlayerBet)
		sceneEx.SendToWorld(int(server.SSPacketID_PACKET_GW_PLAYERBET), gwPlayerBet)
		logger.Trace("Send msg gwPlayerBet ===>", gwPlayerBet)
	}

	playerEx.taxCoin = 0
	playerEx.winCoin = 0
	playerEx.linesWinCoin = 0
	playerEx.jackpotWinCoin = 0
	playerEx.smallGameWinCoin = 0

	if sceneEx.CheckNeedDestroy() && playerEx.freeTimes <= 0 {
		sceneEx.PlayerLeave(playerEx.Player, common.PlayerLeaveReason_OnDestroy, true)
	}
}

func init() {
	ScenePolicyTamQuocSington.RegisteSceneState(&SceneStateTamQuocStart{})
	core.RegisteHook(core.HOOK_BEFORE_START, func() error {
		base.RegisteScenePolicy(common.GameId_TamQuoc, 0, ScenePolicyTamQuocSington)
		return nil
	})
}
