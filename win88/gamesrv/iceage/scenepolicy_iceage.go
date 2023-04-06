package iceage

import (
	"fmt"
	"games.yol.com/win88/protocol/server"
	"os"
	"strconv"
	"sync"
	"time"

	"games.yol.com/win88/common"
	rule "games.yol.com/win88/gamerule/iceage"
	"games.yol.com/win88/gamesrv/base"
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/iceage"
	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/task"
	"github.com/idealeak/goserver/core/timer"
)

////////////////////////////////////////////////////////////////////////////////
//冰河世纪
////////////////////////////////////////////////////////////////////////////////

// 房间内主要逻辑
var ScenePolicyIceAgeSington = &ScenePolicyIceAge{}

type ScenePolicyIceAge struct {
	base.BaseScenePolicy
	states [IceAgeSceneStateMax]base.SceneState
}

//创建场景扩展数据
func (this *ScenePolicyIceAge) CreateSceneExData(s *base.Scene) interface{} {
	sceneEx := NewIceAgeSceneData(s)
	if sceneEx != nil {
		if sceneEx.init() {
			s.SetExtraData(sceneEx)
		}
	}
	return sceneEx
}

//创建玩家扩展数据
func (this *ScenePolicyIceAge) CreatePlayerExData(s *base.Scene, p *base.Player) interface{} {
	playerEx := &IceAgePlayerData{Player: p}
	if playerEx != nil {
		p.SetExtraData(playerEx)
	}
	return playerEx
}

//场景开启事件
func (this *ScenePolicyIceAge) OnStart(s *base.Scene) {
	logger.Trace("(this *ScenePolicyIceAge) OnStart, sceneId=", s.GetSceneId())
	sceneEx := NewIceAgeSceneData(s)
	if sceneEx != nil {
		if sceneEx.init() {
			s.SetExtraData(sceneEx)
			sceneEx.BroadcastJackpot()
			s.ChangeSceneState(IceAgeSceneStateStart) //改变当前的玩家状态
		}
	}
}

//场景关闭事件
func (this *ScenePolicyIceAge) OnStop(s *base.Scene) {
	logger.Trace("(this *ScenePolicyIceAge) OnStop , sceneId=", s.GetSceneId())
	if sceneEx, ok := s.GetExtraData().(*IceAgeSceneData); ok {
		sceneEx.SaveData(true)
	}
}

//场景心跳事件
func (this *ScenePolicyIceAge) OnTick(s *base.Scene) {
	if s == nil {
		return
	}
	if s.GetSceneState() != nil {
		s.GetSceneState().OnTick(s)
	}
}

//玩家进入事件
func (this *ScenePolicyIceAge) OnPlayerEnter(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Trace("(this *ScenePolicyIceAge) OnPlayerEnter, sceneId=", s.GetSceneId(), " player=", p.SnId, s.GetGameId())
	if sceneEx, ok := s.GetExtraData().(*IceAgeSceneData); ok {
		playerEx := &IceAgePlayerData{Player: p}
		playerEx.init(s)                                        // 玩家当前信息初始化
		playerEx.score = sceneEx.GetDBGameFree().GetBaseScore() // 底注初始化
		sceneEx.players[p.SnId] = playerEx
		p.SetExtraData(playerEx)
		IceAgeSendRoomInfo(s, p, sceneEx, playerEx, nil)
		s.FirePlayerEvent(p, base.PlayerEventEnter, nil) //回调会调取 onPlayerEvent事件
	}
}

//玩家离开事件
func (this *ScenePolicyIceAge) OnPlayerLeave(s *base.Scene, p *base.Player, reason int) {
	if s == nil || p == nil {
		return
	}
	logger.Trace("(this *ScenePolicyIceAge) OnPlayerLeave, sceneId=", s.GetSceneId(), " player=", p.SnId)
	if sceneEx, ok := s.GetExtraData().(*IceAgeSceneData); ok {
		if this.CanChangeCoinScene(s, p) {
			if playerEx, ok := p.GetExtraData().(*IceAgePlayerData); ok {
				playerEx.SavePlayerGameData(strconv.Itoa(int(s.GetGameFreeId())))
			}
			sceneEx.OnPlayerLeave(p, reason)
			s.FirePlayerEvent(p, base.PlayerEventLeave, []int64{int64(reason)})
		}
	}
}

//玩家掉线
func (this *ScenePolicyIceAge) OnPlayerDropLine(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Trace("(this *ScenePolicyIceAge) OnPlayerDropLine, sceneId=", s.GetSceneId(), " player=", p.SnId)
	s.FirePlayerEvent(p, base.PlayerEventDropLine, nil)
	if sceneEx, ok := s.GetExtraData().(*IceAgeSceneData); ok {
		if sceneEx.GetGaming() {
			return
		}
	}
}

//玩家重连
func (this *ScenePolicyIceAge) OnPlayerRehold(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Trace("(this *ScenePolicyIceAge) OnPlayerRehold, sceneId=", s.GetSceneId(), " player=", p.Name)
	//gs 玩家rehold的时候不再发送 RoomInfo消息
	if _, ok := s.GetExtraData().(*IceAgeSceneData); ok {
		if _, ok := p.GetExtraData().(*IceAgePlayerData); ok {
			//发送房间信息给自己
			//IceAgeSendRoomInfo(s, p, sceneEx, playerEx)
			s.FirePlayerEvent(p, base.PlayerEventRehold, nil)
		}
	}
}

//玩家返回房间 gs添加
func (this *ScenePolicyIceAge) OnPlayerReturn(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Trace("(this *ScenePolicyIceAge) OnPlayerReturn,sceneId =", s.GetSceneId(), " player= ", p.Name)
	if sceneEx, ok := s.GetExtraData().(*IceAgeSceneData); ok {
		if playerEx, ok := p.GetExtraData().(*IceAgePlayerData); ok {
			IceAgeSendRoomInfo(s, p, sceneEx, playerEx, playerEx.billedData)
			s.FirePlayerEvent(p, base.PlayerEventReturn, nil)
		}
	}
}

//玩家操作
func (this *ScenePolicyIceAge) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if s == nil || p == nil {
		return false
	}
	if s.GetSceneState() != nil {
		return s.GetSceneState().OnPlayerOp(s, p, opcode, params)
	}
	return true
}

//玩家事件
func (this *ScenePolicyIceAge) OnPlayerEvent(s *base.Scene, p *base.Player, evtcode int, params []int64) {
	if s == nil || p == nil {
		return
	}
	if s.GetSceneState() != nil {
		s.GetSceneState().OnPlayerEvent(s, p, evtcode, params)
	}
}

//是否完成了整个牌局
func (this *ScenePolicyIceAge) IsCompleted(s *base.Scene) bool { return false }

//是否可以强制开始
func (this *ScenePolicyIceAge) IsCanForceStart(s *base.Scene) bool { return true }

//当前状态能否换桌
func (this *ScenePolicyIceAge) CanChangeCoinScene(s *base.Scene, p *base.Player) bool {
	if s == nil || p == nil {
		return true
	}
	if s.GetSceneState() != nil {
		return s.GetSceneState().CanChangeCoinScene(s, p)
	}
	return true
}

func (this *ScenePolicyIceAge) RegisteSceneState(state base.SceneState) {
	if state == nil {
		return
	}
	stateid := state.GetState()
	if stateid < 0 || stateid >= IceAgeSceneStateMax {
		return
	}
	this.states[stateid] = state
}

func (this *ScenePolicyIceAge) GetSceneState(s *base.Scene, stateid int) base.SceneState {
	if stateid >= 0 && stateid < IceAgeSceneStateMax {
		return ScenePolicyIceAgeSington.states[stateid]
	}
	return nil
}

func IceAgeSendRoomInfo(s *base.Scene, p *base.Player, sceneEx *IceAgeSceneData, playerEx *IceAgePlayerData, data *iceage.GameBilledData) {
	logger.Trace("-------------------发送房间消息 ", s.RoomId, p.SnId)
	pack := &iceage.SCIceAgeRoomInfo{
		RoomId:     proto.Int(s.GetSceneId()),
		Creator:    proto.Int32(s.GetCreator()),
		GameId:     proto.Int(s.GetGameId()),
		RoomMode:   proto.Int(s.GetSceneMode()),
		Params:     s.GetParams(),
		State:      proto.Int(s.GetSceneState().GetState()),
		Jackpot:    proto.Int64(sceneEx.jackpot.JackpotFund),
		GameFreeId: proto.Int32(s.GetDBGameFree().GetId()),
		BilledData: data,
	}
	if playerEx != nil {
		pd := &iceage.IceAgePlayerData{
			SnId:        proto.Int32(playerEx.SnId),
			Name:        proto.String(playerEx.Name),
			Head:        proto.Int32(playerEx.Head),
			Sex:         proto.Int32(playerEx.Sex),
			Coin:        proto.Int64(playerEx.GetCoin()),
			HeadOutLine: proto.Int32(playerEx.HeadOutLine),
			VIP:         proto.Int32(playerEx.VIP),
		}
		pack.Players = append(pack.Players, pd)
		pack.BetLines = playerEx.betLines
		pack.FreeTimes = proto.Int32(playerEx.freeTimes)
		pack.Chip = proto.Int32(s.DbGameFree.BaseScore)
		pack.TotalPriceBonus = proto.Int64(playerEx.totalPriceBonus)
		pack.SpinID = proto.Int64(playerEx.spinID)
	}
	proto.SetDefaults(pack)
	p.SendToClient(int(iceage.IceAgePacketID_PACKET_SC_ICEAGE_ROOMINFO), pack)
}

type SceneStateIceAgeStart struct {
}

//获取当前场景状态
func (this *SceneStateIceAgeStart) GetState() int { return IceAgeSceneStateStart }

//是否可以切换状态到
func (this *SceneStateIceAgeStart) CanChangeTo(s base.SceneState) bool { return true }

//当前状态能否换桌
func (this *SceneStateIceAgeStart) CanChangeCoinScene(s *base.Scene, p *base.Player) bool {
	return true
}

func (this *SceneStateIceAgeStart) GetTimeout(s *base.Scene) int { return 0 }

func (this *SceneStateIceAgeStart) OnEnter(s *base.Scene) {
	if sceneEx, ok := s.GetExtraData().(*IceAgeSceneData); ok {
		logger.Tracef("(this *base.Scene) [%v] 场景状态进入 %v", s.GetSceneId(), len(sceneEx.players))
		//sceneEx.stateStartTime = time.Now()
		sceneEx.SetStateStartTime(time.Now())
		pack := &iceage.SCIceAgeRoomState{
			State: proto.Int(this.GetState()),
		}
		proto.SetDefaults(pack)
		s.Broadcast(int(iceage.IceAgePacketID_PACKET_SC_ICEAGE_ROOMSTATE), pack, 0)
	}
}

func (this *SceneStateIceAgeStart) OnLeave(s *base.Scene) {}

func (this *SceneStateIceAgeStart) OnTick(s *base.Scene) {
	if time.Now().Sub(s.GetGameStartTime()) > time.Second*3 {
		if sceneEx, ok := s.GetExtraData().(*IceAgeSceneData); ok {
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
		//s.gameStartTime = time.Now()
		s.SetGameStartTime(time.Now())
	}

	if sceneEx, ok := s.GetExtraData().(*IceAgeSceneData); ok {
		for _, p := range sceneEx.players {
			//游戏次数达到目标值
			todayGamefreeIDSceneData, _ := p.GetDaliyGameData(int(sceneEx.GetDBGameFree().GetId()))
			if !p.IsRob &&
				todayGamefreeIDSceneData != nil &&
				sceneEx.GetDBGameFree().GetPlayNumLimit() != 0 &&
				todayGamefreeIDSceneData.GameTimes >= int64(sceneEx.GetDBGameFree().GetPlayNumLimit()) {
				s.PlayerLeave(p.Player, common.PlayerLeaveReason_GameTimes, true)
			}
		}
		if sceneEx.CheckNeedDestroy() {
			for _, player := range sceneEx.players {
				if !player.IsRob {
					if time.Now().Sub(player.GetLastOPTimer()) > 10*time.Second {
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

func (this *SceneStateIceAgeStart) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	playerEx, ok := p.GetExtraData().(*IceAgePlayerData)
	if !ok {
		return false
	}
	sceneEx, ok := s.GetExtraData().(*IceAgeSceneData)
	if !ok {
		return false
	}
	if sceneEx.CheckNeedDestroy() && playerEx.freeTimes <= 0 {
		//离开有统计
		sceneEx.PlayerLeave(playerEx.Player, common.PlayerLeaveReason_OnDestroy, true)
		return false
	}
	switch opcode {
	case IceAgePlayerOpStart: //开始
		//if !iceAgeBenchTest {
		//	iceAgeBenchTest = true
		//	for i := 0; i < 10; i++ {
		//		//this.BenchTest(s, p)
		//		this.WinTargetBenchTest(s, p)
		//	}
		//	return true
		//}

		//params 参数0底注，后面跟客户端选择的线n条线(1<=n<=20)，客户端线是从1开始算起1~20条线
		//参数是否合法
		if len(params) < 2 || len(params) > rule.LINENUM+1 {
			this.OnPlayerSToCOp(s, p, playerEx.GetPos(), opcode, iceage.OpResultCode_OPRC_Error, params)
			return false
		}

		//// 小游戏未结束 不能进行下一次旋转
		//if playerEx.totalPriceBonus > 0 {
		//	this.OnPlayerSToCOp(s, p, playerEx.GetPos(), opcode, iceage.OpResultCode_OPRC_Error, params)
		//	return false
		//}

		//先做底注校验，看是否在指定参数内
		if sceneEx.GetDBGameFree().GetBaseScore() != int32(params[0]) {
			this.OnPlayerSToCOp(s, p, playerEx.GetPos(), opcode, iceage.OpResultCode_OPRC_Error, params)
			return false
		}
		playerEx.score = int32(params[0])

		//判断线条是否重复，是否合法
		lineFlag := make(map[int64]bool)
		lineParams := make([]int64, 0)
		for i := 1; i < len(params); i++ {
			lineNum := params[i]
			if lineNum >= 1 && lineNum <= int64(rule.LINENUM) && !lineFlag[lineNum] {
				lineParams = append(lineParams, lineNum)
				lineFlag[lineNum] = true
			} else {
				this.OnPlayerSToCOp(s, p, playerEx.GetPos(), opcode, iceage.OpResultCode_OPRC_Error, params)
				return false
			}
		}
		//没有选线参数
		if len(lineParams) == 0 {
			this.OnPlayerSToCOp(s, p, playerEx.GetPos(), opcode, iceage.OpResultCode_OPRC_Error, params)
			return false
		}

		//获取总投注金额（所有线的总投注） |  校验玩家余额是否足够
		totalBetValue := (int64(len(lineParams))) * params[0]
		if playerEx.freeTimes <= 0 && totalBetValue > playerEx.GetCoin() {
			this.OnPlayerSToCOp(s, p, playerEx.GetPos(), opcode, iceage.OpResultCode_OPRC_CoinNotEnough, params)
			return false
		} else if playerEx.freeTimes <= 0 && int64(sceneEx.GetDBGameFree().GetBetLimit()) > playerEx.GetCoin() { //押注限制
			this.OnPlayerSToCOp(s, p, playerEx.GetPos(), opcode, iceage.OpResultCode_OPRC_CoinNotEnough, params)
			return false
		}

		//p.lastOPTimer = time.Now()
		//sceneEx.gameNowTime = time.Now()
		//sceneEx.numOfGames++
		//p.gameTimes++
		//playerEx.startCoin = playerEx.GetCoin()
		p.SetLastOPTimer(time.Now())
		sceneEx.SetGameNowTime(time.Now())
		p.SetGameTimes(p.GetGameTimes() + 1)
		sceneEx.SetNumOfGames(sceneEx.GetNumOfGames() + 1)
		//playerEx.SetStartCoin(playerEx.GetCoin())

		//获取当前水池的上下文环境
		cpCtx := base.GetCoinPoolMgr().GetCoinPoolCtx(sceneEx.GetPlatform(), sceneEx.GetGameFreeId(), sceneEx.GetGroupId())
		sceneEx.SetCpCtx(cpCtx)
		//税收比例
		taxRate := sceneEx.GetDBGameFree().GetTaxRate()
		if taxRate < 0 || taxRate > 10000 {
			logger.Warnf("IceAgeErrorTaxRate [%v][%v][%v][%v]", sceneEx.GetGameFreeId(), playerEx.SnId, playerEx.spinID, taxRate)
			taxRate = 500
		}
		//水池设置
		coinPoolSetting := base.GetCoinPoolMgr().GetCoinPoolSetting(sceneEx.GetPlatform(), sceneEx.GetGameFreeId(), sceneEx.GetGroupId())
		baseRate := coinPoolSetting.GetBaseRate() //基础赔率
		ctroRate := coinPoolSetting.GetCtroRate() //调节赔率 暗税系数
		//if baseRate >= 10000 || baseRate <= 0 || ctroRate < 0 || ctroRate >= 1000 || baseRate+ctroRate > 9900 {
		//	logger.Warnf("IceAgeErrorBaseRate [%v][%v][%v][%v][%v]", sceneEx.GetGameFreeId(), playerEx.SnId, playerEx.spinID, baseRate, ctroRate)
		//	baseRate = 9700
		//	ctroRate = 200
		//}
		//jackpotRate := 10000 - (baseRate + ctroRate) //奖池系数
		jackpotRate := ctroRate //奖池系数
		logger.Tracef("IceAgeRates [%v][%v][%v][%v][%v][%v]", sceneEx.GetGameFreeId(), playerEx.SnId, playerEx.spinID, taxRate, baseRate, ctroRate)

		gamePoolCoin := base.GetCoinPoolMgr().LoadCoin(sceneEx.GetGameFreeId(), sceneEx.GetPlatform(), sceneEx.GetGroupId()) // 当前水池金额
		prizeFund := gamePoolCoin - sceneEx.jackpot.JackpotFund                                                              // 除去奖池的水池剩余金额

		var jackpotParam = sceneEx.GetDBGameFree().GetJackpot() // 奖池参数
		var jackpotInit = int64(jackpotParam[rule.ICEAGE_JACKPOT_InitJackpot] * sceneEx.GetDBGameFree().GetBaseScore())
		var jackpotFundAdd int64     //奖池/水池增量
		if playerEx.freeTimes <= 0 { //正常模式才能记录用户的押注变化，免费模式不能改变押注
			playerEx.betLines = lineParams // 选线记录
			//prizeFundAdd := int64(float64(totalBetValue) * (float64(baseRate) / 10000.0))
			jackpotFundAdd = int64(float64(totalBetValue) * (float64(jackpotRate) / 10000.0))
			sceneEx.jackpot.JackpotFund += jackpotFundAdd //奖池增加
			//playerEx.totalBet += totalBetValue            //总下注额（从进房间开始,包含多局游戏的下注）
			playerEx.SetTotalBet(playerEx.GetTotalBet() + totalBetValue)

			p.AddCoin(-totalBetValue, common.GainWay_HundredSceneLost, base.SyncFlag_ToClient, "system", s.GetSceneName())
			p.Statics(sceneEx.GetKeyGameId(), sceneEx.KeyGamefreeId, -totalBetValue, true)
			if !p.IsRob && !sceneEx.GetTesting() {
				// 推送金币
				base.GetCoinPoolMgr().PushCoin(sceneEx.GetGameFreeId(), sceneEx.GetGroupId(), sceneEx.GetPlatform(), int64(float64(totalBetValue)*float64(10000-ctroRate)/10000.0))
			}

			////统计参与游戏次数
			//if !sceneEx.GetTesting() && !playerEx.IsRob {
			//	pack := &server.GWSceneEnd{
			//		GameFreeId: proto.Int32(sceneEx.GetDBGameFree().GetId()),
			//		Players:    []*server.PlayerCtx{&server.PlayerCtx{SnId: proto.Int32(playerEx.SnId), Coin: proto.Int64(playerEx.GetCoin())}},
			//	}
			//	proto.SetDefaults(pack)
			//	sceneEx.SendToWorld(int(server.SSPacketID_PACKET_GW_SCENEEND), pack)
			//}
		} else { //免费次数时，不能改线改底注
			totalBetValue = 0
		}

		var distributionParams, bonusDistributionParams []int32
		//奖池模式
		var slotMode = 0
		//if prizeFund <= int64(jackpotParam[iceage.ICEAGE_JACKPOT_RoomPrizeFundLimit]) { //奖池不足
		if prizeFund <= int64(coinPoolSetting.GetLowerLimit()) { //奖池不足
			distributionParams = getElementDistributionByName(SlotsData).GetParams()
			bonusDistributionParams = getElementDistributionByName(BonusData).GetParams()
		} else {
			slotMode = 1
			distributionParams = getElementDistributionByName(SlotsData_V2).GetParams()
			bonusDistributionParams = getElementDistributionByName(BonusData_V2).GetParams()
		}
		logger.Tracef("IceAgePlayerOpStart prizeFund:%v distributionParams:%v bonusDistributionParams:%v", prizeFund, distributionParams, bonusDistributionParams)

		writeBlackTryTimes := 0
	WriteBlack:
		slotData := make([]int, 0)
		var spinRes SpinResult
		loopCount := 0

		for {
			loopCount++
			if loopCount > 3 {
				break
			}
			slotData = getSlotsDataByElementDistribution(distributionParams)

			spinRes = sceneEx.CalcSpinsPrize(slotData, playerEx.betLines, getElementDistributionByName(SlotsData).GetParams(), bonusDistributionParams, params[0], taxRate)
			logger.Tracef("CalcSpinsPrize:%v", spinRes)

			//不允许有免费旋转同时还出大奖
			if playerEx.freeTimes > 0 && spinRes.IsJackpot {
				continue
			}
			//不允许出现两次小游戏
			if spinRes.BonusGameCnt > 1 {
				continue
			}

			if spinRes.TotalPrizeLine > 0 {
				if spinRes.IsJackpot {
					//是否超过一天内的返奖最大金额
					//if jackpotParam[iceage.ICEAGE_JACKPOT_QuantityInDay]
					//todo
					//IF(@SetJackPot = 0 AND @_RoomID IN (2,3,4))
					//CONTINUE;

					// 现金池不够补充 奖池初值 和 玩家奖金 时，不允许出现爆奖
					if prizeFund-jackpotInit-(spinRes.TotalPrizeLine-sceneEx.jackpot.JackpotFund)-spinRes.TotalPrizeBonus <= 0 {
						continue
					}
				}

				var limitWin = int32(0)
				switch slotMode {
				case 0:
					limitWin = jackpotParam[rule.ICEAGE_JACKPOT_LIMITWIN_PRIZELOW]
				case 1:
					limitWin = jackpotParam[rule.ICEAGE_JACKPOT_LIMITWIN_PRIZEHIGH]
				}

				if totalBetValue > 0 && !spinRes.IsJackpot && spinRes.TotalPrizeLine > totalBetValue*int64(limitWin) {
					continue
				}
				if !spinRes.IsJackpot && spinRes.TotalPrizeLine+spinRes.TotalPrizeBonus > prizeFund {
					// 水池不足 不再进行黑白名单调控
					writeBlackTryTimes = 999
					continue
				}
			}
			break
		}

		limitWinScore := int64(coinPoolSetting.GetLowerLimit())
		// 多次尝试后 没有选好牌
		if loopCount > 3 {
			logger.Warnf("CalcSpinsPrize at loopCount>3 :%v", spinRes)
			if sceneEx.GetDBGameFree().GetSceneType() > 1 { // IF(@_RoomID > 1)
				slotData = getSlotsDataByGroupName(DefaultData)
			} else if prizeFund < limitWinScore || playerEx.freeTimes > 0 {
				slotData = getSlotsDataByGroupName(DefaultData)
			} else {
				slotData = getSlotsDataByGroupName(DefaultData_v1)
			}
			spinRes = sceneEx.CalcSpinsPrize(slotData, playerEx.betLines, getElementDistributionByName(SlotsData).GetParams(), bonusDistributionParams, params[0], taxRate)
		}

		// 黑白名单调控 防止异常循环，添加上限次数
		if writeBlackTryTimes < 100 && playerEx.CheckBlackWriteList(spinRes.TotalPrizeLine+spinRes.TotalPrizeBonus > totalBetValue) {
			writeBlackTryTimes++
			goto WriteBlack
		} else if writeBlackTryTimes >= 100 && writeBlackTryTimes != 999 {
			logger.Warnf("IceAgeWriteBlackTryTimesOver [%v][%v][%v][%v][%v]", sceneEx.GetGameFreeId(), playerEx.SnId, gamePoolCoin, playerEx.BlackLevel, playerEx.WhiteLevel)
		}

		///////////测试游戏数据 开始//////////
		//if playerEx.DebugBonus && sceneEx.SceneType == 1 {
		//	if playerEx.TestNum >= len(DebugData) {
		//		playerEx.TestNum = 0
		//	}
		//	slotData = DebugData[playerEx.TestNum]
		//	spinRes = sceneEx.CalcSpinsPrize(slotData, playerEx.betLines, getElementDistributionByName(SlotsData).GetParams(), bonusDistributionParams, params[0], taxRate)
		//	//playerEx.DebugBonus = false
		//	playerEx.TestNum++
		//}
		///////////测试游戏数据 结束//////////

		if spinRes.IsJackpot {
			sceneEx.jackpot.JackpotFund = jackpotInit
		}

		if spinRes.TotalPrizeLine > 0 || spinRes.TotalPrizeBonus > 0 {
			p.AddCoin(spinRes.TotalPrizeLine+spinRes.TotalPrizeBonus, common.GainWay_HundredSceneWin, 0, "system", s.GetSceneName())
			p.Statics(sceneEx.GetKeyGameId(), sceneEx.KeyGamefreeId, spinRes.TotalPrizeLine+spinRes.TotalPrizeBonus+spinRes.TotalTaxScore, true)
			if !p.IsRob && !sceneEx.GetTesting() {
				base.GetCoinPoolMgr().PopCoin(sceneEx.GetGameFreeId(), sceneEx.GetGroupId(), sceneEx.GetPlatform(), spinRes.TotalPrizeLine+spinRes.TotalPrizeBonus+spinRes.TotalTaxScore)
			}
			playerEx.taxCoin = spinRes.TotalTaxScore
			playerEx.AddServiceFee(playerEx.taxCoin)
		}

		var isFreeFlag bool
		//免费次数
		if playerEx.freeTimes > 0 {
			playerEx.freeTimes--
			isFreeFlag = true
		}
		playerEx.freeTimes += spinRes.AddFreeTimes

		rule.SpinID++
		playerEx.spinID = rule.SpinID
		if spinRes.TotalPrizeBonus > 0 { // 小游戏开始
			logger.Tracef("BonusGame Start [%v][%v][%v][%v]", sceneEx.GetGameFreeId(), playerEx.SnId, playerEx.spinID, spinRes.TotalPrizeBonus)
			playerEx.totalPriceBonus = spinRes.TotalPrizeBonus
			playerEx.BonusLineIdx = 0
			//// 小游戏超时处理
			//playerEx.bonusTimerHandle, _ = timer.StartTimer(timer.TimerActionWrapper(func(h timer.TimerHandle, ud interface{}) bool {
			//	this.OnPlayerOp(s, p, IceAgeBonusGame, []int64{playerEx.spinID})
			//	return true
			//}), nil, IceAgeBonusGameTimeout, 1)
		}
		playerEx.cards = spinRes.SlotsData
		playerEx.winCoin = spinRes.TotalPrizeLine + spinRes.TotalPrizeJackpot + spinRes.TotalPrizeBonus + playerEx.taxCoin
		playerEx.linesWinCoin = spinRes.TotalPrizeLine
		playerEx.jackpotWinCoin = spinRes.TotalPrizeJackpot
		playerEx.smallGameWinCoin = spinRes.TotalPrizeBonus
		//playerEx.currentBet = totalBetValue
		//playerEx.currentTax = playerEx.taxCoin
		playerEx.SetCurrentBet(totalBetValue)
		playerEx.SetCurrentTax(playerEx.taxCoin)

		this.OnPlayerSToCOp(s, p, playerEx.GetPos(), opcode, iceage.OpResultCode_OPRC_Sucess, append(params[:1], playerEx.betLines...))
		playerEx.billedData = &iceage.GameBilledData{
			SpinID:          proto.Int64(playerEx.spinID),
			TotalBetValue:   proto.Int64(totalBetValue),
			TotalPriceValue: proto.Int64(spinRes.TotalPrizeLine),
			IsJackpot:       proto.Bool(spinRes.IsJackpot),
			Jackpot:         proto.Int64(sceneEx.jackpot.JackpotFund),
			Balance:         proto.Int64(playerEx.GetCoin()),
			TotalFreeSpin:   proto.Int32(playerEx.freeTimes),
			TotalPriceBonus: proto.Int64(spinRes.TotalPrizeBonus),
			TotalJackpot:    proto.Int64(spinRes.TotalPrizeJackpot),
			PrizesData:      spinRes.LinesInfo,
			SlotsData:       spinRes.SlotsData,
		}
		pack := &iceage.SCIceAgeGameBilled{
			BilledData: playerEx.billedData,
		}
		proto.SetDefaults(pack)
		logger.Infof("IceAgePlayerOpStart %v", pack)
		p.SendToClient(int(iceage.IceAgePacketID_PACKET_SC_ICEAGE_GAMEBILLED), pack)

		// 记录本次操作
		var allCards = make([][]int32, 0)
		for _, v := range pack.BilledData.SlotsData {
			allCards = append(allCards, v.Card)
		}
		playerEx.RollGameType.BaseResult.WinTotal = pack.BilledData.GetTotalPriceValue() + pack.BilledData.GetTotalPriceBonus()
		playerEx.RollGameType.BaseResult.IsFree = isFreeFlag
		playerEx.RollGameType.BaseResult.WinSmallGame = pack.BilledData.GetTotalPriceBonus()
		playerEx.RollGameType.BaseResult.AllWinNum = int32(len(pack.BilledData.PrizesData))
		playerEx.RollGameType.BaseResult.WinRate = spinRes.TotalWinRate
		playerEx.RollGameType.Cards = allCards
		playerEx.RollGameType.WinLines = spinRes.WinLines
		playerEx.RollGameType.BaseResult.WinJackpot = pack.BilledData.GetTotalJackpot()
		playerEx.RollGameType.BaseResult.WinLineScore = pack.BilledData.TotalPriceValue
		IceAgeCheckAndSaveLog(sceneEx, playerEx)

		// 广播奖池
		if totalBetValue == 0 && !spinRes.IsJackpot { // 没改变奖池
			return true
		}
		// 添加进开奖记录里面
		if spinRes.IsJackpot {
			spinid := strconv.FormatInt(int64(playerEx.SnId), 10)
			// 推送最新开奖记录到world
			msg := &server.GWGameNewBigWinHistory{
				SceneId: proto.Int32(int32(sceneEx.GetSceneId())),
				BigWinHistory: &server.BigWinHistoryInfo{
					SpinID:      proto.String(spinid),
					CreatedTime: proto.Int64(time.Now().Unix()),
					BaseBet:     proto.Int64(int64(playerEx.score)),
					TotalBet:    proto.Int64(int64(playerEx.CurrentBet)),
					PriceValue:  proto.Int64(pack.BilledData.GetTotalJackpot()),
					UserName:    proto.String(playerEx.Name),
				},
			}
			proto.SetDefaults(msg)
			logger.Logger.Infof("GWGameNewBigWinHistory gameFreeID %v %v", sceneEx.GetDBGameFree().GetId(), msg)
			sceneEx.SendToWorld(int(server.SSPacketID_PACKET_GW_GAMENEWBIGWINHISTORY), msg)
		} else {
			sceneEx.PushVirtualDataToWorld() // 推送虚拟数据
		}
		sceneEx.BroadcastJackpot()
	case IceAgePlayerHistory:
		task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
			spinid := strconv.FormatInt(int64(playerEx.SnId), 10)
			gpl := model.GetPlayerListByHallEx(p.SnId, p.Platform, 0, 50, 0, 0, 0, s.DbGameFree.GetGameClass(), s.GetGameId())
			pack := &iceage.SCIceAgePlayerHistory{}
			for _, v := range gpl.Data {
				if v.GameDetailedLogId == "" {
					logger.Error("IceAgePlayerHistory GameDetailedLogId is nil")
					break
				}
				gdl := model.GetPlayerHistory(p.Platform, v.GameDetailedLogId)
				if gdl == nil {
					logger.Error("IceAgePlayerHistory gdl is nil")
					continue
				}
				data, err := UnMarshalIceAgeGameNote(gdl.GameDetailedNote)
				if err != nil {
					logger.Errorf("UnMarshalIceAgeGameNote error:%v", err)
				}
				gnd := data.(*GameResultLog)
				player := &iceage.IceAgePlayerHistoryInfo{
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
			logger.Info("IceAgePlayerHistory: ", pack)
			return pack

		}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
			if data == nil {
				logger.Error("IceAgePlayerHistory data is nil")
				return
			}
			p.SendToClient(int(iceage.IceAgePacketID_PACKET_SC_ICEAGE_PLAYERHISTORY), data)
		}), "CSGetIceAgePlayerHistoryHandler").Start()
	case IceAgeBonusGame:
		//params 参数0 spinID
		//参数是否合法
		if len(params) < 2 {
			this.OnPlayerSToCOp(s, p, playerEx.GetPos(), opcode, iceage.OpResultCode_OPRC_Error, params)
			return false
		}
		if playerEx.spinID != params[0] || playerEx.totalPriceBonus <= 0 || params[1] != playerEx.BonusLineIdx {
			this.OnPlayerSToCOp(s, p, playerEx.GetPos(), opcode, iceage.OpResultCode_OPRC_Error, params)
			return false
		}
		idx := params[1]
		logger.Tracef("BonusGame End [%v][%v][%v][%v][%v]", sceneEx.GetGameFreeId(), playerEx.SnId, playerEx.spinID, playerEx.totalPriceBonus, idx)

		if playerEx.bonusTimerHandle != timer.TimerHandle(0) {
			timer.StopTimer(playerEx.bonusTimerHandle)
			playerEx.bonusTimerHandle = timer.TimerHandle(0)
		}

		//p.AddCoin(playerEx.totalPriceBonus, common.GainWay_HundredSceneWin, false, "system", s.GetSceneName())
		//if !p.IsRob && !sceneEx.GetTesting() {
		//	GetCoinPoolMgr().PopCoin(sceneEx.GetGameFreeId(), sceneEx.GetGroupId(), sceneEx.GetPlatform(), playerEx.totalPriceBonus)
		//	p.Statics(sceneEx.GetKeyGameId(), sceneEx.GetGameFreeId(), playerEx.totalPriceBonus, true)
		//}
		this.OnPlayerSToCOp(s, p, playerEx.GetPos(), opcode, iceage.OpResultCode_OPRC_Sucess, params)
		//playerEx.totalPriceBonus = 0
	case IceAgeBonusGameStart:
		logger.Logger.Tracef("params:%v", params)
		if len(params) < 2 || params[1] < 0 || int(params[1]) > len(playerEx.billedData.PrizesData) {
			this.OnPlayerSToCOp(s, p, playerEx.GetPos(), opcode, iceage.OpResultCode_OPRC_Error, params)
			return false
		}
		idx := params[1]
		if idx > int64(playerEx.BonusLineIdx) || playerEx.BonusLineIdx == 0 {
			// 小游戏超时处理
			playerEx.BonusLineIdx = idx
			if playerEx.bonusTimerHandle != timer.TimerHandle(0) {
				timer.StopTimer(playerEx.bonusTimerHandle)
				playerEx.bonusTimerHandle = timer.TimerHandle(0)
			}
			playerEx.bonusTimerHandle, _ = timer.StartTimer(timer.TimerActionWrapper(func(h timer.TimerHandle, ud interface{}) bool {
				this.OnPlayerOp(s, p, IceAgeBonusGame, []int64{playerEx.spinID, playerEx.BonusLineIdx})
				return true
			}), nil, IceAgeBonusGameTimeout, 1)
			this.OnPlayerSToCOp(s, p, playerEx.GetPos(), opcode, iceage.OpResultCode_OPRC_Sucess, params)
		}

	}
	return true
}

func (this *SceneStateIceAgeStart) OnPlayerEvent(s *base.Scene, p *base.Player, evtcode int, params []int64) {
}

//发送玩家操作情况
func (this *SceneStateIceAgeStart) OnPlayerSToCOp(s *base.Scene, p *base.Player, pos int, opcode int,
	opRetCode iceage.OpResultCode, params []int64) {
	pack := &iceage.SCIceAgeOp{
		SnId:      proto.Int32(p.SnId),
		OpCode:    proto.Int(opcode),
		OpRetCode: opRetCode,
		Params:    params,
	}
	proto.SetDefaults(pack)
	p.SendToClient(int(iceage.IceAgePacketID_PACKET_SC_ICEAGE_PLAYEROP), pack)
}

var iceAgeBenchTest bool
var iceAgeBenchTestTimes int

func (this *SceneStateIceAgeStart) BenchTest(s *base.Scene, p *base.Player) {
	const BENCH_CNT = 10000
	setting := base.GetCoinPoolMgr().GetCoinPoolSetting(s.GetPlatform(), s.GetGameFreeId(), s.GetGroupId())
	oldPoolCoin := base.GetCoinPoolMgr().LoadCoin(s.GetGameFreeId(), s.GetPlatform(), s.GetGroupId())
	if iceAgeBenchTestTimes == 0 {
		defaultVal := int64(setting.GetLowerLimit())
		if oldPoolCoin != defaultVal {
			base.GetCoinPoolMgr().PushCoin(s.GetGameFreeId(), s.GetGroupId(), s.GetPlatform(), defaultVal-oldPoolCoin)
		}
	}
	iceAgeBenchTestTimes++

	fileName := fmt.Sprintf("iceage-%v-%d.csv", p.SnId, iceAgeBenchTestTimes)
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm)
	defer file.Close()
	if err != nil {
		file, err = os.Create(fileName)
		if err != nil {
			return
		}
	}
	file.WriteString("玩家id,当前水位,之前余额,之后余额,投入,产出,税收,小游戏,爆奖,中线倍数,中线数,剩余免费次数\r\n")

	oldCoin := p.GetCoin()
	//p.coin = int64(5000 * s.GetDBGameFree().GetBaseScore())
	p.SetCoin(int64(5000 * s.GetDBGameFree().GetBaseScore()))
	if playerEx, ok := p.GetExtraData().(*IceAgePlayerData); ok {
		for i := 0; i < BENCH_CNT; i++ {
			startCoin := p.GetCoin()
			freeTimes := playerEx.freeTimes
			poolCoin := base.GetCoinPoolMgr().LoadCoin(s.GetGameFreeId(), s.GetPlatform(), s.GetGroupId())
			playerEx.UnmarkFlag(base.PlayerState_GameBreak)
			suc := this.OnPlayerOp(s, p, IceAgePlayerOpStart, append([]int64{int64(playerEx.score)}, rule.AllBetLines...))
			inCoin := int64(playerEx.RollGameType.BaseResult.TotalBet)
			outCoin := playerEx.RollGameType.BaseResult.ChangeCoin + inCoin
			taxCoin := playerEx.RollGameType.BaseResult.Tax
			lineScore := float64(playerEx.RollGameType.BaseResult.WinRate*s.GetDBGameFree().GetBaseScore()) * float64(10000.0-s.GetDBGameFree().GetTaxRate()) / 10000.0
			jackpotScore := outCoin - playerEx.RollGameType.BaseResult.WinSmallGame - int64(lineScore+0.00001)

			str := fmt.Sprintf("%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v\r\n", p.SnId, poolCoin, startCoin, p.GetCoin(), inCoin, outCoin, taxCoin,
				playerEx.RollGameType.BaseResult.WinSmallGame, jackpotScore, playerEx.RollGameType.BaseResult.WinRate, playerEx.RollGameType.BaseResult.AllWinNum, freeTimes)
			file.WriteString(str)
			if !suc {
				break
			}

			if playerEx.totalPriceBonus > 0 {
				this.OnPlayerOp(s, p, IceAgeBonusGame, []int64{playerEx.spinID})
			}
		}
	}
	p.SetCoin(oldCoin)
}
func (this *SceneStateIceAgeStart) WinTargetBenchTest(s *base.Scene, p *base.Player) {
	const BENCH_CNT = 10000
	var once = sync.Once{}
	once.Do(func() {
		setting := base.GetCoinPoolMgr().GetCoinPoolSetting(s.GetPlatform(), s.GetGameFreeId(), s.GetGroupId())
		oldPoolCoin := base.GetCoinPoolMgr().LoadCoin(s.GetGameFreeId(), s.GetPlatform(), s.GetGroupId())
		if iceAgeBenchTestTimes == 0 {
			defaultVal := int64(setting.GetLowerLimit())
			if oldPoolCoin != defaultVal {
				base.GetCoinPoolMgr().PushCoin(s.GetGameFreeId(), s.GetGroupId(), s.GetPlatform(), defaultVal-oldPoolCoin)
			}
		}
	})
	iceAgeBenchTestTimes++

	fileName := fmt.Sprintf("iceage-%v-%d.csv", p.SnId, iceAgeBenchTestTimes)
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm)
	defer file.Close()
	if err != nil {
		file, err = os.Create(fileName)
		if err != nil {
			return
		}
	}
	file.WriteString("玩家id,当前水位,之前余额,之后余额,投入,产出,税收,小游戏,爆奖,中线倍数,中线数,剩余免费次数\r\n")
	oldCoin := p.GetCoin()
	switch s.GetDBGameFree().GetSceneType() {
	case 1:
		p.SetCoin(100000)
	case 2:
		p.SetCoin(500000)
	case 3:
		p.SetCoin(1000000)
	default:
		p.SetCoin(100000)
	}
	var targetCoin = p.GetCoin() + p.GetCoin()/10
	if playerEx, ok := p.GetExtraData().(*IceAgePlayerData); ok {
		for i := 0; p.GetCoin() < targetCoin; i++ {
			startCoin := p.GetCoin()
			freeTimes := playerEx.freeTimes
			poolCoin := base.GetCoinPoolMgr().LoadCoin(s.GetGameFreeId(), s.GetPlatform(), s.GetGroupId())
			playerEx.UnmarkFlag(base.PlayerState_GameBreak)
			suc := this.OnPlayerOp(s, p, IceAgePlayerOpStart, append([]int64{int64(playerEx.score)}, rule.AllBetLines...))
			inCoin := int64(playerEx.RollGameType.BaseResult.TotalBet)
			outCoin := playerEx.RollGameType.BaseResult.ChangeCoin + inCoin
			taxCoin := playerEx.RollGameType.BaseResult.Tax
			lineScore := float64(playerEx.RollGameType.BaseResult.WinRate*s.GetDBGameFree().GetBaseScore()) * float64(10000.0-s.GetDBGameFree().GetTaxRate()) / 10000.0
			jackpotScore := outCoin - playerEx.RollGameType.BaseResult.WinSmallGame - int64(lineScore+0.00001)

			str := fmt.Sprintf("%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v\r\n", p.SnId, poolCoin, startCoin, p.GetCoin(), inCoin, outCoin, taxCoin,
				playerEx.RollGameType.BaseResult.WinSmallGame, jackpotScore, playerEx.RollGameType.BaseResult.WinRate, playerEx.RollGameType.BaseResult.AllWinNum, freeTimes)
			file.WriteString(str)
			if !suc {
				break
			}

			if playerEx.totalPriceBonus > 0 {
				this.OnPlayerOp(s, p, IceAgeBonusGame, []int64{playerEx.spinID})
			}
			if i > BENCH_CNT {
				break
			}
		}
	}
	p.SetCoin(oldCoin)
}
func IceAgeCheckAndSaveLog(sceneEx *IceAgeSceneData, playerEx *IceAgePlayerData) {
	//统计金币变动
	//log1
	logger.Trace("IceAgeCheckAndSaveLog Save ", playerEx.SnId)
	//changeCoin := playerEx.GetCoin() - playerEx.GetStartCoin()
	changeCoin := playerEx.winCoin - playerEx.taxCoin - playerEx.CurrentBet
	startCoin := playerEx.Coin - changeCoin
	playerEx.SaveSceneCoinLog(startCoin, changeCoin,
		playerEx.GetCoin(), playerEx.GetCurrentBet(), playerEx.taxCoin, playerEx.winCoin, playerEx.jackpotWinCoin, playerEx.smallGameWinCoin)

	//log2
	playerEx.RollGameType.BaseResult.ChangeCoin = changeCoin
	playerEx.RollGameType.BaseResult.BasicBet = sceneEx.GetDBGameFree().GetBaseScore()
	playerEx.RollGameType.BaseResult.RoomId = int32(sceneEx.GetSceneId())
	playerEx.RollGameType.BaseResult.AfterCoin = playerEx.GetCoin()
	playerEx.RollGameType.BaseResult.BeforeCoin = startCoin
	playerEx.RollGameType.BaseResult.IsFirst = sceneEx.IsPlayerFirst(playerEx.Player)
	playerEx.RollGameType.BaseResult.PlayerSnid = playerEx.SnId
	playerEx.RollGameType.BaseResult.TotalBet = int32(playerEx.GetCurrentBet())
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
	if !sceneEx.GetTesting() && !playerEx.IsRob {
		playerBet := &server.PlayerBet{
			SnId:       proto.Int32(playerEx.SnId),
			Bet:        proto.Int64(playerEx.GetCurrentBet()),
			Gain:       proto.Int64(playerEx.RollGameType.BaseResult.ChangeCoin),
			Tax:        proto.Int64(playerEx.taxCoin),
			Coin:       proto.Int64(playerEx.GetCoin()),
			GameCoinTs: proto.Int64(playerEx.GameCoinTs),
		}
		gwPlayerBet := &server.GWPlayerBet{
			SceneId:    proto.Int(sceneEx.SceneId),
			GameFreeId: proto.Int32(sceneEx.GetDBGameFree().GetId()),
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
	ScenePolicyIceAgeSington.RegisteSceneState(&SceneStateIceAgeStart{})
	core.RegisteHook(core.HOOK_BEFORE_START, func() error {
		base.RegisteScenePolicy(common.GameId_IceAge, 0, ScenePolicyIceAgeSington)
		return nil
	})
}
