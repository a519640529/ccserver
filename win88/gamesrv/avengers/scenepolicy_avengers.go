package avengers

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"

	"games.yol.com/win88/common"
	rule "games.yol.com/win88/gamerule/avengers"
	"games.yol.com/win88/gamesrv/base"
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/avengers"
	"games.yol.com/win88/protocol/server"
	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/task"
	"github.com/idealeak/goserver/core/timer"
)

////////////////////////////////////////////////////////////////////////////////
//复仇者联盟
////////////////////////////////////////////////////////////////////////////////

// 房间内主要逻辑
var ScenePolicyAvengersSington = &ScenePolicyAvengers{}

type ScenePolicyAvengers struct {
	base.BaseScenePolicy
	states [AvengersSceneStateMax]base.SceneState
}

//创建场景扩展数据
func (this *ScenePolicyAvengers) CreateSceneExData(s *base.Scene) interface{} {
	sceneEx := NewAvengersSceneData(s)
	if sceneEx != nil {
		if sceneEx.init() {
			s.ExtraData = sceneEx
		}
	}
	return sceneEx
}

//创建玩家扩展数据
func (this *ScenePolicyAvengers) CreatePlayerExData(s *base.Scene, p *base.Player) interface{} {
	playerEx := &AvengersPlayerData{Player: p}
	if playerEx != nil {
		p.ExtraData = playerEx
	}
	return playerEx
}

//场景开启事件
func (this *ScenePolicyAvengers) OnStart(s *base.Scene) {
	logger.Trace("(this *ScenePolicyAvengers) OnStart, SceneId=", s.SceneId)
	sceneEx := NewAvengersSceneData(s)
	if sceneEx != nil {
		if sceneEx.init() {
			sceneEx.BroadcastJackpot()
			s.ExtraData = sceneEx
			s.ChangeSceneState(AvengersSceneStateStart) //改变当前的玩家状态
		}
	}
}

//场景关闭事件
func (this *ScenePolicyAvengers) OnStop(s *base.Scene) {
	logger.Trace("(this *ScenePolicyAvengers) OnStop , SceneId=", s.SceneId)
	if sceneEx, ok := s.ExtraData.(*AvengersSceneData); ok {
		sceneEx.SaveData(true)
	}
}

//场景心跳事件
func (this *ScenePolicyAvengers) OnTick(s *base.Scene) {
	if s == nil {
		return
	}
	if s.SceneState != nil {
		s.SceneState.OnTick(s)
	}
}

//玩家进入事件
func (this *ScenePolicyAvengers) OnPlayerEnter(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Trace("(this *ScenePolicyAvengers) OnPlayerEnter, SceneId=", s.SceneId, " player=", p.SnId)
	if sceneEx, ok := s.ExtraData.(*AvengersSceneData); ok {
		playerEx := &AvengersPlayerData{Player: p}
		playerEx.init(s)                                   // 玩家当前信息初始化
		playerEx.score = sceneEx.DbGameFree.GetBaseScore() // 底注
		sceneEx.players[p.SnId] = playerEx
		p.ExtraData = playerEx
		AvengersSendRoomInfo(s, p, sceneEx, playerEx, nil)
		s.FirePlayerEvent(p, base.PlayerEventEnter, nil) //回调会调取 onPlayerEvent事件
	}
}

//玩家离开事件
func (this *ScenePolicyAvengers) OnPlayerLeave(s *base.Scene, p *base.Player, reason int) {
	if s == nil || p == nil {
		return
	}
	logger.Trace("(this *ScenePolicyAvengers) OnPlayerLeave, SceneId=", s.SceneId, " player=", p.SnId)
	if sceneEx, ok := s.ExtraData.(*AvengersSceneData); ok {
		if this.CanChangeCoinScene(s, p) {
			if playerEx, ok := p.ExtraData.(*AvengersPlayerData); ok {
				playerEx.SavePlayerGameData(strconv.Itoa(int(s.GetGameFreeId())))
			}
			sceneEx.OnPlayerLeave(p, reason)
			s.FirePlayerEvent(p, base.PlayerEventLeave, []int64{int64(reason)})
		}
	}
}

//玩家掉线
func (this *ScenePolicyAvengers) OnPlayerDropLine(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Trace("(this *ScenePolicyAvengers) OnPlayerDropLine, SceneId=", s.SceneId, " player=", p.SnId)
	s.FirePlayerEvent(p, base.PlayerEventDropLine, nil)
	if sceneEx, ok := s.ExtraData.(*AvengersSceneData); ok {
		if sceneEx.Gaming {
			return
		}
	}
}

//玩家重连
func (this *ScenePolicyAvengers) OnPlayerRehold(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Trace("(this *ScenePolicyAvengers) OnPlayerRehold, SceneId=", s.SceneId, " player=", p.Name)
	//if sceneEx, ok := s.ExtraData.(*AvengersSceneData); ok {
	//	if playerEx, ok := p.ExtraData.(*AvengersPlayerData); ok {
	//发送房间信息给自己
	//AvengersSendRoomInfo(s, p, sceneEx, playerEx)
	s.FirePlayerEvent(p, base.PlayerEventRehold, nil)
	//	}
	//}
}

//玩家返回房间 gs添加
func (this *ScenePolicyAvengers) OnPlayerReturn(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Trace("(this *ScenePolicyIceAge) OnPlayerReturn,sceneId =", s.GetSceneId(), " player= ", p.Name)
	if sceneEx, ok := s.GetExtraData().(*AvengersSceneData); ok {
		if playerEx, ok := p.GetExtraData().(*AvengersPlayerData); ok {
			AvengersSendRoomInfo(s, p, sceneEx, playerEx, playerEx.billedData)
			s.FirePlayerEvent(p, base.PlayerEventReturn, nil)
		}
	}
}

//玩家操作
func (this *ScenePolicyAvengers) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if s == nil || p == nil {
		return false
	}
	if s.SceneState != nil {
		return s.SceneState.OnPlayerOp(s, p, opcode, params)
	}
	return true
}

//玩家事件
func (this *ScenePolicyAvengers) OnPlayerEvent(s *base.Scene, p *base.Player, evtcode int, params []int64) {
	if s == nil || p == nil {
		return
	}
	if s.SceneState != nil {
		s.SceneState.OnPlayerEvent(s, p, evtcode, params)
	}
}

//是否完成了整个牌局
func (this *ScenePolicyAvengers) IsCompleted(s *base.Scene) bool { return false }

//是否可以强制开始
func (this *ScenePolicyAvengers) IsCanForceStart(s *base.Scene) bool { return true }

//当前状态能否换桌
func (this *ScenePolicyAvengers) CanChangeCoinScene(s *base.Scene, p *base.Player) bool {
	if s == nil || p == nil {
		return true
	}
	if s.SceneState != nil {
		return s.SceneState.CanChangeCoinScene(s, p)
	}
	return true
}

func (this *ScenePolicyAvengers) RegisteSceneState(state base.SceneState) {
	if state == nil {
		return
	}
	stateid := state.GetState()
	if stateid < 0 || stateid >= AvengersSceneStateMax {
		return
	}
	this.states[stateid] = state
}

func (this *ScenePolicyAvengers) GetSceneState(s *base.Scene, stateid int) base.SceneState {
	if stateid >= 0 && stateid < AvengersSceneStateMax {
		return ScenePolicyAvengersSington.states[stateid]
	}
	return nil
}

func AvengersSendRoomInfo(s *base.Scene, p *base.Player, sceneEx *AvengersSceneData, playerEx *AvengersPlayerData, data *avengers.GameBilledData) {
	logger.Trace("-------------------发送房间消息 ", s.RoomId, p.SnId)
	pack := &avengers.SCAvengersRoomInfo{
		RoomId:     proto.Int(s.SceneId),
		Creator:    proto.Int32(s.Creator),
		GameId:     proto.Int(s.GameId),
		RoomMode:   proto.Int(s.GameMode),
		Params:     s.Params,
		State:      proto.Int(s.SceneState.GetState()),
		Jackpot:    proto.Int64(sceneEx.jackpot.JackpotFund),
		GameFreeId: proto.Int32(s.DbGameFree.Id),
		BilledData: data,
	}
	if playerEx != nil {
		pd := &avengers.AvengersPlayerData{
			SnId:        proto.Int32(playerEx.SnId),
			Name:        proto.String(playerEx.Name),
			Head:        proto.Int32(playerEx.Head),
			Sex:         proto.Int32(playerEx.Sex),
			Coin:        proto.Int64(playerEx.Coin),
			HeadOutLine: proto.Int32(playerEx.HeadOutLine),
			VIP:         proto.Int32(playerEx.VIP),
		}
		pack.Players = append(pack.Players, pd)
		pack.BetLines = playerEx.betLines
		pack.FreeTimes = proto.Int32(playerEx.freeTimes)
		pack.Chip = proto.Int32(s.DbGameFree.BaseScore)
		pack.SpinID = proto.Int64(playerEx.spinID)
		if playerEx.totalPriceBonus > 0 {
			switch playerEx.bonusStage {
			case 0:
				pack.ParamsEx = append(pack.ParamsEx, playerEx.bonusStage)
			case 1:
				if time.Now().Unix()-playerEx.bonusStartTime >= AvengersBonusGameStageTimeout*2 {
					playerEx.CleanBonus()
				} else if time.Now().Unix()-playerEx.bonusStartTime >= AvengersBonusGameStageTimeout {
					playerEx.bonusStage = 2
					playerEx.bonusStartTime += AvengersBonusGameStageTimeout
					pack.ParamsEx = append(pack.ParamsEx, playerEx.bonusStage)
					leftTime := playerEx.bonusStartTime + AvengersBonusGameStageTimeout + 2 - time.Now().Unix()
					pack.ParamsEx = append(pack.ParamsEx, int32(leftTime))
				} else {
					pack.ParamsEx = append(pack.ParamsEx, playerEx.bonusStage)
					leftTime := playerEx.bonusStartTime + AvengersBonusGameStageTimeout + 2 - time.Now().Unix()
					pack.ParamsEx = append(pack.ParamsEx, int32(leftTime))
					pack.ParamsEx = append(pack.ParamsEx, playerEx.bonusOpRecord...)
				}
			case 2:
				if time.Now().Unix()-playerEx.bonusStartTime >= AvengersBonusGameStageTimeout {
					playerEx.CleanBonus()
				} else {
					pack.ParamsEx = append(pack.ParamsEx, playerEx.bonusStage)
					leftTime := playerEx.bonusStartTime + AvengersBonusGameStageTimeout + 2 - time.Now().Unix()
					pack.ParamsEx = append(pack.ParamsEx, int32(leftTime))
				}
			}
		}
		if playerEx.totalPriceBonus > 0 {
			pack.TotalPriceBonus = proto.Int64(playerEx.totalPriceBonus)
			pack.BonusGame = playerEx.bonusGame
			pack.BonusX = playerEx.bonusX
		}
	}
	proto.SetDefaults(pack)
	p.SendToClient(int(avengers.AvengersPacketID_PACKET_SC_AVENGERS_ROOMINFO), pack)
}

type SceneStateAvengersStart struct {
}

//获取当前场景状态
func (this *SceneStateAvengersStart) GetState() int { return AvengersSceneStateStart }

//是否可以切换状态到
func (this *SceneStateAvengersStart) CanChangeTo(s base.SceneState) bool { return true }

//当前状态能否换桌
func (this *SceneStateAvengersStart) CanChangeCoinScene(s *base.Scene, p *base.Player) bool {
	if _, ok := p.ExtraData.(*AvengersPlayerData); ok {
		return true
	}
	return true
}

func (this *SceneStateAvengersStart) GetTimeout(s *base.Scene) int { return 0 }

func (this *SceneStateAvengersStart) OnEnter(s *base.Scene) {
	if sceneEx, ok := s.ExtraData.(*AvengersSceneData); ok {
		logger.Tracef("(this *Scene) [%v] 场景状态进入 %v", s.SceneId, len(sceneEx.players))
		sceneEx.StateStartTime = time.Now()
		pack := &avengers.SCAvengersRoomState{
			State: proto.Int(this.GetState()),
		}
		proto.SetDefaults(pack)
		s.Broadcast(int(avengers.AvengersPacketID_PACKET_SC_AVENGERS_ROOMSTATE), pack, 0)
	}
}

func (this *SceneStateAvengersStart) OnLeave(s *base.Scene) {}

func (this *SceneStateAvengersStart) OnTick(s *base.Scene) {
	if time.Now().Sub(s.GameStartTime) > time.Second*3 {
		if sceneEx, ok := s.ExtraData.(*AvengersSceneData); ok {
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

	if sceneEx, ok := s.ExtraData.(*AvengersSceneData); ok {
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

func (this *SceneStateAvengersStart) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	playerEx, ok := p.ExtraData.(*AvengersPlayerData)
	if !ok {
		return false
	}
	sceneEx, ok := s.ExtraData.(*AvengersSceneData)
	if !ok {
		return false
	}
	if sceneEx.CheckNeedDestroy() && playerEx.freeTimes <= 0 {
		//离开有统计
		sceneEx.PlayerLeave(playerEx.Player, common.PlayerLeaveReason_OnDestroy, true)
		return false
	}
	switch opcode {
	case AvengersPlayerOpStart: //开始
		//if !avengersBenchTest {
		//	avengersBenchTest = true
		//	for i := 0; i < 10; i++ {
		//		//this.BenchTest(s, p)
		//		this.WinTargetBenchTest(s, p)
		//	}
		//	return true
		//}

		//参数是否合法
		//params 参数0底注，后面跟客户端选择的线n条线(1<=n<=25)，客户端线是从1开始算起1~25条线
		if len(params) < 2 || len(params) > rule.LINENUM+1 {
			this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, avengers.OpResultCode_OPRC_Error, params)
			return false
		}
		//先做底注校验
		if sceneEx.DbGameFree.GetBaseScore() != int32(params[0]) {
			this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, avengers.OpResultCode_OPRC_Error, params)
			return false
		}
		playerEx.score = int32(params[0]) // 单线押注数
		// 小游戏未结束 不能进行下一次旋转
		if playerEx.totalPriceBonus > 0 {
			this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, avengers.OpResultCode_OPRC_Error, params)
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
				this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, avengers.OpResultCode_OPRC_Error, params)
				return false
			}
		}
		//没有选线参数
		if len(lineParams) == 0 {
			this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, avengers.OpResultCode_OPRC_Error, params)
			return false
		}

		//获取总投注金额（所有线的总投注） |  校验玩家余额是否足够
		totalBetValue := (int64(len(lineParams))) * params[0]
		if playerEx.freeTimes <= 0 && totalBetValue > playerEx.Coin {
			this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, avengers.OpResultCode_OPRC_CoinNotEnough, params)
			return false
		} else if playerEx.freeTimes <= 0 && int64(sceneEx.DbGameFree.GetBetLimit()) > playerEx.Coin { //押注限制
			this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, avengers.OpResultCode_OPRC_CoinNotEnough, params)
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
			logger.Warnf("AvengersErrorTaxRate [%v][%v][%v][%v]", sceneEx.GetGameFreeId(), playerEx.SnId, playerEx.spinID, taxRate)
			taxRate = 500
		}
		//水池设置
		coinPoolSetting := base.CoinPoolMgr.GetCoinPoolSetting(sceneEx.Platform, sceneEx.GetGameFreeId(), sceneEx.GroupId)
		baseRate := coinPoolSetting.GetBaseRate() //基础赔率
		ctroRate := coinPoolSetting.GetCtroRate() //调节赔率 暗税系数
		//if baseRate >= 10000 || baseRate <= 0 || ctroRate < 0 || ctroRate >= 1000 || baseRate+ctroRate > 9900 {
		//	logger.Warnf("AvengersErrorBaseRate [%v][%v][%v][%v][%v]", sceneEx.GetGameFreeId(), playerEx.SnId, playerEx.spinID, baseRate, ctroRate)
		//	baseRate = 9700
		//	ctroRate = 200
		//}
		//jackpotRate := 10000 - (baseRate + ctroRate) //奖池系数
		jackpotRate := ctroRate //奖池系数
		logger.Tracef("AvengersRates [%v][%v][%v][%v][%v][%v]", sceneEx.GetGameFreeId(), playerEx.SnId, playerEx.spinID, taxRate, baseRate, ctroRate)

		gamePoolCoin := base.CoinPoolMgr.LoadCoin(sceneEx.GetGameFreeId(), sceneEx.Platform, sceneEx.GroupId) // 当前水池金额
		prizeFund := gamePoolCoin - sceneEx.jackpot.JackpotFund                                               // 除去奖池的水池剩余金额

		// 奖池参数
		jackpotParams := sceneEx.DbGameFree.GetJackpot()
		var jackpotInit = int64(jackpotParams[rule.AVENGERS_JACKPOT_InitJackpot] * sceneEx.DbGameFree.GetBaseScore()) //奖池初始值

		var jackpotFundAdd, prizeFundAdd int64
		if playerEx.freeTimes <= 0 { //正常模式才能记录用户的押注变化，免费模式不能改变押注
			playerEx.betLines = lineParams                                                    // 选线记录
			jackpotFundAdd = int64(float64(totalBetValue) * (float64(jackpotRate) / 10000.0)) //奖池要增加的金额
			prizeFundAdd = int64(float64(totalBetValue) * (float64(baseRate) / 10000.0))
			playerEx.TotalBet += totalBetValue //总下注额（从进房间开始,包含多局游戏的下注）
			//扣除投注金币
			p.AddCoin(-totalBetValue, common.GainWay_HundredSceneLost, base.SyncFlag_ToClient, "system", s.GetSceneName())
			p.Statics(sceneEx.KeyGameId, sceneEx.KeyGamefreeId, -totalBetValue, true)
			if !p.IsRob && !sceneEx.Testing {
				// 推送金币
				base.CoinPoolMgr.PushCoin(sceneEx.GetGameFreeId(), sceneEx.GroupId, sceneEx.Platform, int64(float64(totalBetValue)*(float64(10000-ctroRate)/10000.0)))
			}

			////统计参与游戏次数
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

		writeBlackTryTimes := 0
	WriteBlack:
		slotData := make([]int, 0)
		var spinRes AvengersSpinResult
		var slotDataIsOk bool
		for {
			var symbolType rule.Symbol
			if prizeFund < int64(coinPoolSetting.GetLowerLimit()) { // userInfo.prizeFund < limit * roomId
				symbolType = rule.SYMBOL1
			} else {
				symbolType = rule.SYMBOL2
			}
			slotData, _ = rule.GenerateSlotsData_v2(symbolType)

			spinRes = sceneEx.CalcLinePrize(slotData, playerEx.betLines, params[0])

			// 免费次数时 不允许爆奖
			if playerEx.freeTimes > 0 && spinRes.IsJackpot {
				break
			}
			if spinRes.JackpotCnt > 1 {
				break
			}

			// 现金池不足时 重新发牌
			spinCondition := prizeFund + prizeFundAdd - (spinRes.TotalPrizeLine + spinRes.BonusGame.GetTotalPrizeValue() + spinRes.TotalPrizeJackpot)
			if spinRes.IsJackpot {
				spinCondition += jackpotFundAdd + sceneEx.jackpot.JackpotFund - jackpotInit
			}
			if spinCondition < 0 {
				if !spinRes.IsJackpot { // 非爆奖 水池不足 不再进行黑白名单调控
					writeBlackTryTimes = 999
				}
				break
			}

			// 非爆奖时 不允许赢取太大的奖励
			var limitBigWin int64 = 50
			if spinCondition < int64(coinPoolSetting.GetLowerLimit()) { //现金池不足时
				limitBigWin = int64(jackpotParams[rule.AVENGERS_JACKPOT_LIMITWIN_PRIZELOW])
			} else {
				limitBigWin = int64(jackpotParams[rule.AVENGERS_JACKPOT_LIMITWIN_PRIZEHIGH])
			}
			if totalBetValue > 0 && !spinRes.IsJackpot && spinRes.TotalPrizeLine+spinRes.BonusGame.GetTotalPrizeValue() > totalBetValue*limitBigWin {
				break
			}

			// 小游戏次数限制
			// 爆奖次数限制

			slotDataIsOk = true
			break
		}

		if !slotDataIsOk {
			rand.Seed(time.Now().Unix())
			slotData = rule.MissData[rand.Intn(len(rule.MissData))]
			spinRes = sceneEx.CalcLinePrize(slotData, playerEx.betLines, params[0])
		}

		// 黑白名单调控 防止异常循环，添加上限次数
		if writeBlackTryTimes < 100 && playerEx.CheckBlackWriteList(spinRes.TotalPrizeLine+spinRes.TotalPrizeJackpot+spinRes.BonusGame.GetTotalPrizeValue() > totalBetValue) {
			writeBlackTryTimes++
			goto WriteBlack
		} else if writeBlackTryTimes >= 100 && writeBlackTryTimes != 999 {
			logger.Warnf("AvengersWriteBlackTryTimesOver [%v][%v][%v][%v][%v]", sceneEx.GetGameFreeId(), playerEx.SnId, gamePoolCoin, playerEx.BlackLevel, playerEx.WhiteLevel)
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
		} else {
			sceneEx.jackpot.JackpotFund += jackpotFundAdd
		}

		// 玩家赢钱
		totalWinScore := spinRes.TotalPrizeLine + spinRes.TotalPrizeJackpot
		if totalWinScore > 0 || len(spinRes.BonusX) > 0 {
			p.AddCoin(totalWinScore+spinRes.BonusGame.GetTotalPrizeValue(), common.GainWay_HundredSceneWin, 0, "system", s.GetSceneName())
			p.Statics(sceneEx.KeyGameId, sceneEx.KeyGamefreeId, totalWinScore+spinRes.BonusGame.GetTotalPrizeValue()+spinRes.TotalTaxScore, true)
			if !p.IsRob && !sceneEx.Testing {
				base.CoinPoolMgr.PopCoin(sceneEx.GetGameFreeId(), sceneEx.GroupId, sceneEx.Platform, totalWinScore+spinRes.BonusGame.GetTotalPrizeValue()+spinRes.TotalTaxScore)
			}
			playerEx.taxCoin = spinRes.TotalTaxScore
			playerEx.AddServiceFee(playerEx.taxCoin)
		}

		this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, avengers.OpResultCode_OPRC_Sucess, append(params[:1], playerEx.betLines...))

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
		if len(spinRes.BonusX) > 0 {
			playerEx.CleanBonus()
			playerEx.totalPriceBonus = spinRes.BonusGame.GetTotalPrizeValue()
			playerEx.bonusGame = &spinRes.BonusGame
			playerEx.bonusX = spinRes.BonusX
			logger.Tracef("BonusGame Start [%v][%v][%v][%v]", sceneEx.GetGameFreeId(), playerEx.SnId, playerEx.spinID, playerEx.totalPriceBonus)
			//playerEx.bonusTimerHandle, _ = timer.StartTimer(timer.TimerActionWrapper(func(h timer.TimerHandle, ud interface{}) bool {
			//	this.OnPlayerOp(s, p, AvengersBonusGame, []int64{playerEx.spinID})
			//	return true
			//}), nil, AvengersBonusGameTimeout, 1)
		}
		playerEx.billedData = &avengers.GameBilledData{
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
			BonusX:                 spinRes.BonusX,
			BonusGame:              &spinRes.BonusGame,
		}
		pack := &avengers.SCAvengersGameBilled{
			BilledData: playerEx.billedData,
		}
		proto.SetDefaults(pack)
		logger.Infof("AvengersPlayerOpStart %v", pack)
		p.SendToClient(int(avengers.AvengersPacketID_PACKET_SC_AVENGERS_GAMEBILLED), pack)

		// 记录本次操作
		playerEx.RollGameType.BaseResult.WinTotal = pack.BilledData.GetTotalPrizeValue() + pack.BilledData.GetBonusGame().GetTotalPrizeValue() + pack.BilledData.GetTotalJackpotValue()
		playerEx.RollGameType.BaseResult.IsFree = isFreeFlag
		playerEx.RollGameType.BaseResult.WinJackpot = pack.BilledData.GetTotalJackpotValue()
		playerEx.RollGameType.BaseResult.WinLineScore = pack.BilledData.GetTotalPrizeValue()
		playerEx.RollGameType.BaseResult.WinSmallGame = pack.BilledData.BonusGame.GetTotalPrizeValue()
		playerEx.RollGameType.BaseResult.AllWinNum = int32(len(pack.BilledData.PrizeLines))
		playerEx.RollGameType.BaseResult.WinRate = spinRes.TotalWinRate
		playerEx.RollGameType.BaseResult.Cards = pack.BilledData.GetSlotsData()
		playerEx.RollGameType.WinLines = spinRes.WinLines
		AvengersCheckAndSaveLog(sceneEx, playerEx)

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
					PriceValue:  proto.Int64(int64(pack.BilledData.GetTotalJackpotValue())),
					UserName:    proto.String(playerEx.Name),
					TotalBet:    proto.Int64(int64(playerEx.CurrentBet)),
				},
			}
			proto.SetDefaults(msg)
			logger.Logger.Infof("GWGameNewBigWinHistory gameFreeID %v %v", sceneEx.DbGameFree.GetId(), msg)
			sceneEx.SendToWorld(int(server.SSPacketID_PACKET_GW_GAMENEWBIGWINHISTORY), msg)
		}
		sceneEx.BroadcastJackpot()
	case AvengersPlayerHistory:
		task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
			spinid := strconv.FormatInt(int64(playerEx.SnId), 10)
			gpl := model.GetPlayerListByHallEx(p.SnId, p.Platform, 0, 50, 0, 0, 0, s.DbGameFree.GetGameClass(), s.GameId) //复仇者联盟存储个人操作记录不分场次，因为选场界面也需要拉去个人操作记录
			pack := &avengers.SCAvengersPlayerHistory{}
			for _, v := range gpl.Data {
				if v.GameDetailedLogId == "" {
					logger.Error("AvengersPlayerHistory GameDetailedLogId is nil")
					break
				}
				gdl := model.GetPlayerHistory(p.Platform, v.GameDetailedLogId)
				if gdl == nil {
					logger.Logger.Error("AvengersPlayerHistory gdl is nil")
					continue
				}
				data, err := UnMarshalAvengersGameNote(gdl.GameDetailedNote)
				if err != nil {
					logger.Errorf("UnMarshalAvengersGameNote error:%v", err)
				}
				gnd := data.(*GameResultLog)
				player := &avengers.AvengersPlayerHistoryInfo{
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
			logger.Info("AvengersPlayerHistory: ", pack)
			return pack

		}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
			if data == nil {
				logger.Error("AvengersPlayerHistory data is nil")
				return
			}
			p.SendToClient(int(avengers.AvengersPacketID_PACKET_SC_AVENGERS_PLAYERHISTORY), data)
		}), "CSGetAvengersPlayerHistoryHandler").Start()
	case AvengersBonusGame:
		//params 参数0 spinID
		//参数是否合法
		if len(params) < 1 {
			this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, avengers.OpResultCode_OPRC_Error, params)
			return false
		}
		if playerEx.spinID != params[0] || playerEx.totalPriceBonus <= 0 {
			this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, avengers.OpResultCode_OPRC_Error, params)
			return false
		}
		if playerEx.bonusTimerHandle != timer.TimerHandle(0) {
			timer.StopTimer(playerEx.bonusTimerHandle)
			playerEx.bonusTimerHandle = timer.TimerHandle(0)
		}
		logger.Tracef("BonusGame Start [%v][%v][%v][%v]", sceneEx.GetGameFreeId(), playerEx.SnId, playerEx.spinID, playerEx.totalPriceBonus)

		this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, avengers.OpResultCode_OPRC_Sucess, []int64{playerEx.totalPriceBonus, playerEx.Coin})
		playerEx.CleanBonus()
	case AvengersBonusGameRecord:
		// params[0] 小游戏阶段： 0 小游戏动画开始 1 小游戏界面1 2 切换小游戏界面2
		// params[1] 小游戏界面1时，选择奖项的界面位置信息
		if len(params) < 1 || params[0] < 0 || params[0] > 2 || playerEx.totalPriceBonus <= 0 || (params[0] == 1 && len(params) < 2) {
			this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, avengers.OpResultCode_OPRC_Error, params)
			return false
		}
		if params[0] == 0 {
			if playerEx.bonusStage > 0 {
				this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, avengers.OpResultCode_OPRC_Error, params)
				return false
			}
			playerEx.bonusStage = 1
			playerEx.bonusStartTime = time.Now().Unix()
		} else if params[0] == 2 {
			if playerEx.bonusStage != 1 {
				this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, avengers.OpResultCode_OPRC_Error, params)
				return false
			}
			playerEx.bonusStage = 2
			playerEx.bonusStartTime = time.Now().Unix()
		} else if params[0] == 1 {
			if params[1] < 0 {
				this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, avengers.OpResultCode_OPRC_Error, params)
				return false
			} else if playerEx.bonusGame != nil && len(playerEx.bonusOpRecord) >= len(playerEx.bonusGame.BonusData) {
				this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, avengers.OpResultCode_OPRC_Error, params)
				return false
			}
			playerEx.bonusOpRecord = append(playerEx.bonusOpRecord, int32(params[1]))
		}
		this.OnPlayerSToCOp(s, p, playerEx.Pos, opcode, avengers.OpResultCode_OPRC_Sucess, params)
	}
	return true
}

func (this *SceneStateAvengersStart) OnPlayerEvent(s *base.Scene, p *base.Player, evtcode int, params []int64) {
}

//发送玩家操作情况
func (this *SceneStateAvengersStart) OnPlayerSToCOp(s *base.Scene, p *base.Player, pos int, opcode int,
	opRetCode avengers.OpResultCode, params []int64) {
	pack := &avengers.SCAvengersOp{
		SnId:      proto.Int32(p.SnId),
		OpCode:    proto.Int(opcode),
		OpRetCode: opRetCode,
		Params:    params,
	}
	proto.SetDefaults(pack)
	p.SendToClient(int(avengers.AvengersPacketID_PACKET_SC_AVENGERS_PLAYEROP), pack)
}

var avengersBenchTest bool
var avengersBenchTestTimes int

func (this *SceneStateAvengersStart) BenchTest(s *base.Scene, p *base.Player) {
	const BENCH_CNT = 10000
	setting := base.CoinPoolMgr.GetCoinPoolSetting(s.Platform, s.GetGameFreeId(), s.GroupId)
	oldPoolCoin := base.CoinPoolMgr.LoadCoin(s.GetGameFreeId(), s.Platform, s.GroupId)
	if avengersBenchTestTimes == 0 {
		defaultVal := int64(setting.GetLowerLimit())
		if oldPoolCoin != defaultVal {
			base.CoinPoolMgr.PushCoin(s.GetGameFreeId(), s.GroupId, s.Platform, defaultVal-oldPoolCoin)
		}
	}
	avengersBenchTestTimes++

	fileName := fmt.Sprintf("avengers-%v-%d.csv", p.SnId, avengersBenchTestTimes)
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm)
	defer file.Close()
	if err != nil {
		file, err = os.Create(fileName)
		if err != nil {
			return
		}
	}
	file.WriteString("玩家id,当前水位,之前余额,之后余额,投入,产出,税收,小游戏,爆奖,中线倍数,中线数,剩余免费次数\r\n")

	oldCoin := p.Coin
	p.Coin = int64(5000 * s.DbGameFree.GetBaseScore())
	if playerEx, ok := p.ExtraData.(*AvengersPlayerData); ok {
		for i := 0; i < BENCH_CNT; i++ {
			startCoin := p.Coin
			freeTimes := playerEx.freeTimes
			poolCoin := base.CoinPoolMgr.LoadCoin(s.GetGameFreeId(), s.Platform, s.GroupId)
			playerEx.UnmarkFlag(base.PlayerState_GameBreak)
			suc := this.OnPlayerOp(s, p, AvengersPlayerOpStart, append([]int64{int64(playerEx.score)}, rule.AllBetLines...))
			inCoin := int64(playerEx.RollGameType.BaseResult.TotalBet)
			outCoin := playerEx.RollGameType.BaseResult.ChangeCoin + inCoin
			taxCoin := playerEx.RollGameType.BaseResult.Tax
			lineScore := float64(playerEx.RollGameType.BaseResult.WinRate*s.DbGameFree.GetBaseScore()) * float64(10000.0-s.DbGameFree.GetTaxRate()) / 10000.0
			jackpotScore := outCoin - playerEx.RollGameType.BaseResult.ChangeCoin - int64(lineScore+0.00001)

			str := fmt.Sprintf("%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v\r\n", p.SnId, poolCoin, startCoin, p.Coin, inCoin, outCoin, taxCoin,
				playerEx.RollGameType.BaseResult.WinSmallGame, jackpotScore, playerEx.RollGameType.BaseResult.WinRate, playerEx.RollGameType.BaseResult.AllWinNum, freeTimes)
			file.WriteString(str)
			if !suc {
				break
			}

			if playerEx.totalPriceBonus > 0 {
				this.OnPlayerOp(s, p, AvengersBonusGame, []int64{playerEx.spinID})
			}
		}
	}
	p.Coin = oldCoin
}
func (this *SceneStateAvengersStart) WinTargetBenchTest(s *base.Scene, p *base.Player) {
	const BENCH_CNT = 10000
	var once = sync.Once{}
	once.Do(func() {
		setting := base.CoinPoolMgr.GetCoinPoolSetting(s.Platform, s.GetGameFreeId(), s.GroupId)
		oldPoolCoin := base.CoinPoolMgr.LoadCoin(s.GetGameFreeId(), s.Platform, s.GroupId)
		if avengersBenchTestTimes == 0 {
			defaultVal := int64(setting.GetLowerLimit())
			if oldPoolCoin != defaultVal {
				base.CoinPoolMgr.PushCoin(s.GetGameFreeId(), s.GroupId, s.Platform, defaultVal-oldPoolCoin)
			}
		}
	})
	avengersBenchTestTimes++

	fileName := fmt.Sprintf("avengers-%v-%d.csv", p.SnId, avengersBenchTestTimes)
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm)
	defer file.Close()
	if err != nil {
		file, err = os.Create(fileName)
		if err != nil {
			return
		}
	}
	file.WriteString("玩家id,当前水位,之前余额,之后余额,投入,产出,税收,小游戏,爆奖,中线倍数,中线数,剩余免费次数\r\n")
	oldCoin := p.Coin
	switch s.DbGameFree.GetSceneType() {
	case 1:
		p.Coin = 100000
	case 2:
		p.Coin = 500000
	case 3:
		p.Coin = 1000000
	case 4:
		p.Coin = 10000000
	default:
		p.Coin = 100000
	}
	var targetCoin = p.Coin + p.Coin/10
	if playerEx, ok := p.ExtraData.(*AvengersPlayerData); ok {
		for i := 0; p.Coin < targetCoin; i++ {
			startCoin := p.Coin
			freeTimes := playerEx.freeTimes
			poolCoin := base.CoinPoolMgr.LoadCoin(s.GetGameFreeId(), s.Platform, s.GroupId)
			playerEx.UnmarkFlag(base.PlayerState_GameBreak)
			suc := this.OnPlayerOp(s, p, AvengersPlayerOpStart, append([]int64{int64(playerEx.score)}, rule.AllBetLines...))
			inCoin := int64(playerEx.RollGameType.BaseResult.TotalBet)
			outCoin := playerEx.RollGameType.BaseResult.ChangeCoin + inCoin
			taxCoin := playerEx.RollGameType.BaseResult.Tax
			lineScore := float64(playerEx.RollGameType.BaseResult.WinRate*s.DbGameFree.GetBaseScore()) * float64(10000.0-s.DbGameFree.GetTaxRate()) / 10000.0
			jackpotScore := outCoin - playerEx.RollGameType.BaseResult.WinSmallGame - int64(lineScore+0.00001)

			str := fmt.Sprintf("%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v,%v\r\n", p.SnId, poolCoin, startCoin, p.Coin, inCoin, outCoin, taxCoin,
				playerEx.RollGameType.BaseResult.WinSmallGame, jackpotScore, playerEx.RollGameType.BaseResult.WinRate, playerEx.RollGameType.BaseResult.AllWinNum, freeTimes)
			file.WriteString(str)
			if !suc {
				break
			}

			if playerEx.totalPriceBonus > 0 {
				this.OnPlayerOp(s, p, AvengersBonusGame, []int64{playerEx.spinID})
			}
			if i > BENCH_CNT {
				break
			}
		}
	}
	p.Coin = oldCoin
}

func AvengersCheckAndSaveLog(sceneEx *AvengersSceneData, playerEx *AvengersPlayerData) {
	//统计金币变动
	//log1
	logger.Trace("AvengersCheckAndSaveLog Save ", playerEx.SnId)
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
	ScenePolicyAvengersSington.RegisteSceneState(&SceneStateAvengersStart{})
	core.RegisteHook(core.HOOK_BEFORE_START, func() error {
		base.RegisteScenePolicy(common.GameId_Avengers, 0, ScenePolicyAvengersSington)
		return nil
	})
}
