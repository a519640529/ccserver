package richblessed

import (
	"time"

	"games.yol.com/win88/common"
	"games.yol.com/win88/gamerule/richblessed"
	"games.yol.com/win88/gamesrv/base"
	"games.yol.com/win88/proto"
	protocol "games.yol.com/win88/protocol/richblessed"
	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/logger"
)

// ////////////////////////////////////////////////////////////
var ScenePolicyRichBlessedSington = &ScenePolicyRichBlessed{}

type ScenePolicyRichBlessed struct {
	base.BaseScenePolicy
	states [richblessed.RichBlessedStateMax]base.SceneState
}

// 创建场景扩展数据
func (this *ScenePolicyRichBlessed) CreateSceneExData(s *base.Scene) interface{} {
	sceneEx := NewRichBlessedSceneData(s)
	if sceneEx != nil {
		if sceneEx.GetInit() {
			s.SetExtraData(sceneEx)
		}
	}
	return sceneEx
}

// 创建玩家扩展数据
func (this *ScenePolicyRichBlessed) CreatePlayerExData(s *base.Scene, p *base.Player) interface{} {
	playerEx := &RichBlessedPlayerData{Player: p}
	p.SetExtraData(playerEx)
	return playerEx
}

// 场景开启事件
func (this *ScenePolicyRichBlessed) OnStart(s *base.Scene) {
	logger.Logger.Trace("(this *ScenePolicyRichBlessed) OnStart, sceneId=", s.GetSceneId())
	sceneEx := NewRichBlessedSceneData(s)
	if sceneEx != nil {
		if sceneEx.GetInit() {
			s.SetExtraData(sceneEx)
			s.ChangeSceneState(richblessed.RichBlessedStateStart)
		}
	}
}

// 场景关闭事件
func (this *ScenePolicyRichBlessed) OnStop(s *base.Scene) {
	logger.Logger.Trace("(this *ScenePolicyRichBlessed) OnStop , sceneId=", s.GetSceneId())
}

// 场景心跳事件
func (this *ScenePolicyRichBlessed) OnTick(s *base.Scene) {
	if s == nil {
		return
	}
	if s.GetSceneState() != nil {
		s.GetSceneState().OnTick(s)
	}
}

// 玩家进入事件
func (this *ScenePolicyRichBlessed) OnPlayerEnter(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyRichBlessed) OnPlayerEnter, sceneId=", s.GetSceneId(), " player=", p.Name, "bet:", s.DbGameFree.GetOtherIntParams())
	if sceneEx, ok := s.GetExtraData().(*RichBlessedSceneData); ok {
		playerEx := &RichBlessedPlayerData{Player: p}
		playerEx.init()
		playerEx.betIdx = 0
		playerEx.betCoin = int64(s.DbGameFree.GetOtherIntParams()[0]) // 初始化
		playerEx.oneBetCoin = playerEx.betCoin / richblessed.LineNum  // 单注

		sceneEx.players[p.SnId] = playerEx

		p.SetExtraData(playerEx)
		RichBlessedSendRoomInfo(s, sceneEx, playerEx)

		s.FirePlayerEvent(p, base.PlayerEventEnter, nil)
	}
}

// 玩家离开事件
func (this *ScenePolicyRichBlessed) OnPlayerLeave(s *base.Scene, p *base.Player, reason int) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyRichBlessed) OnPlayerLeave, sceneId=", s.GetSceneId(), " player=", p.SnId)
	if sceneEx, ok := s.ExtraData.(*RichBlessedSceneData); ok {
		s.FirePlayerEvent(p, base.PlayerEventLeave, nil)
		sceneEx.OnPlayerLeave(p, reason)
	}
}

// 玩家掉线
func (this *ScenePolicyRichBlessed) OnPlayerDropLine(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyRichBlessed) OnPlayerDropLine, sceneId=", s.GetSceneId(), " player=", p.SnId)
	s.FirePlayerEvent(p, base.PlayerEventDropLine, nil)
}

// 玩家重连
func (this *ScenePolicyRichBlessed) OnPlayerRehold(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyRichBlessed) OnPlayerRehold, sceneId=", s.GetSceneId(), " player=", p.SnId)
	if sceneEx, ok := s.GetExtraData().(*RichBlessedSceneData); ok {
		if playerEx, ok := p.GetExtraData().(*RichBlessedPlayerData); ok {
			RichBlessedSendRoomInfo(s, sceneEx, playerEx)
		}
	}
}

// 返回房间
func (this *ScenePolicyRichBlessed) OnPlayerReturn(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyRichBlessed) OnPlayerReturn, GetSceneId()=", s.GetSceneId(), " player=", p.Name)
	if sceneEx, ok := s.GetExtraData().(*RichBlessedSceneData); ok {
		if playerEx, ok := p.GetExtraData().(*RichBlessedPlayerData); ok {
			//if p.IsMarkFlag(base.PlayerState_Auto) {
			//	p.UnmarkFlag(base.PlayerState_Auto)
			//	p.SyncFlag()
			//}
			//发送房间信息给自己
			RichBlessedSendRoomInfo(s, sceneEx, playerEx)
			s.FirePlayerEvent(p, base.PlayerEventReturn, nil)
		}
	}
}

func RichBlessedSendRoomInfo(s *base.Scene, sceneEx *RichBlessedSceneData, playerEx *RichBlessedPlayerData) {
	pack := RichBlessedCreateRoomInfoPacket(s, sceneEx, playerEx)
	logger.Logger.Trace("RoomInfo: ", pack)
	playerEx.SendToClient(int(protocol.RBPID_PACKET_RICHBLESSED_SCRBRoomInfo), pack)
}
func RichBlessedCreateRoomInfoPacket(s *base.Scene, sceneEx *RichBlessedSceneData, playerEx *RichBlessedPlayerData) interface{} {
	//房间信息
	pack := &protocol.SCRBRoomInfo{
		RoomId:     proto.Int(s.SceneId),
		GameId:     proto.Int(s.GameId),
		RoomMode:   proto.Int(s.SceneMode),
		SceneType:  proto.Int(s.SceneType),
		Params:     s.Params,
		NumOfGames: proto.Int(sceneEx.NumOfGames),
		State:      proto.Int(s.SceneState.GetState()),
		ParamsEx:   s.DbGameFree.OtherIntParams, //s.GetParamsEx(),
		//BetLimit:   s.DbGameFree.BetLimit,

		NowGameState:  proto.Int(playerEx.gameState),
		BetIdx:        proto.Int(playerEx.betIdx),
		FreeAllWin:    proto.Int64(playerEx.freewinCoin),
		SmallJackpot:  proto.Int64(playerEx.oneBetCoin * richblessed.JkEleNumRate[richblessed.BlueGirl]),
		MiddleJackpot: proto.Int64(playerEx.oneBetCoin * richblessed.JkEleNumRate[richblessed.BlueBoy]),
		BigJackpot:    proto.Int64(playerEx.oneBetCoin * richblessed.JkEleNumRate[richblessed.GoldGirl]),
		GrandJackpot:  proto.Int64(playerEx.oneBetCoin * richblessed.JkEleNumRate[richblessed.GoldBoy]),
		WinEleCoin:    proto.Int64(playerEx.winCoin),
		WinRate:       proto.Int64(playerEx.winLineRate),
		FreeNum:       proto.Int64(int64(playerEx.freeTimes)),
		AddFreeNum:    proto.Int64(int64(playerEx.addfreeTimes)),
		JackpotEle:    proto.Int32(playerEx.JackpotEle),
		WinJackpot:    proto.Int64(playerEx.JackwinCoin),
		GameFreeId:    proto.Int32(s.DbGameFree.Id),
	}

	//自己的信息
	if playerEx != nil {
		pd := &protocol.RBPlayerData{
			SnId:        proto.Int32(playerEx.SnId),
			Name:        proto.String(playerEx.Name),
			Head:        proto.Int32(playerEx.Head),
			Sex:         proto.Int32(playerEx.Sex),
			Coin:        proto.Int64(playerEx.Coin),
			Pos:         proto.Int(playerEx.Pos),
			Flag:        proto.Int(playerEx.GetFlag()),
			City:        proto.String(playerEx.City),
			HeadOutLine: proto.Int32(playerEx.HeadOutLine),
			VIP:         proto.Int32(playerEx.VIP),
		}
		pack.Player = pd
	}
	if playerEx.result != nil {
		pack.Cards = playerEx.result.EleValue
		var wl []*protocol.RichWinLine
		for _, r := range playerEx.result.WinLine {
			wl = append(wl, &protocol.RichWinLine{
				Poss: r.Poss,
			})
		}
		pack.WinLines = wl
	}
	proto.SetDefaults(pack)
	return pack
}
func (this *ScenePolicyRichBlessed) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if s == nil || p == nil {
		return false
	}
	logger.Logger.Trace("(this *ScenePolicyRichBlessed) OnPlayerOp, sceneId=", s.GetSceneId(), " player=", p.SnId, " opcode=", opcode, " params=", params)
	if s.GetSceneState() != nil {
		if s.GetSceneState().OnPlayerOp(s, p, opcode, params) {
			p.SetLastOPTimer(time.Now())
			return true
		}
		return false
	}
	return true
}

func (this *ScenePolicyRichBlessed) OnPlayerEvent(s *base.Scene, p *base.Player, evtcode int, params []int64) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyRichBlessed) OnPlayerEvent, sceneId=", s.GetSceneId(), " player=", p.SnId, " eventcode=", evtcode, " params=", params)
	if s.GetSceneState() != nil {
		s.GetSceneState().OnPlayerEvent(s, p, evtcode, params)
	}
}

// 当前状态能否换桌
func (this *ScenePolicyRichBlessed) CanChangeCoinScene(s *base.Scene, p *base.Player) bool {
	if s == nil || p == nil {
		return false
	}
	if s.GetSceneState() != nil {
		return s.GetSceneState().CanChangeCoinScene(s, p)
	}
	return false
}

// 状态基类
type SceneBaseStateRichBlessed struct {
}

func (this *SceneBaseStateRichBlessed) GetTimeout(s *base.Scene) int {
	if sceneEx, ok := s.GetExtraData().(*RichBlessedSceneData); ok {
		return int(time.Now().Sub(sceneEx.GetStateStartTime()) / time.Second)
	}
	return 0
}

func (this *SceneBaseStateRichBlessed) CanChangeTo(s base.SceneState) bool {
	return true
}

// 当前状态能否换桌
func (this *SceneBaseStateRichBlessed) CanChangeCoinScene(s *base.Scene, p *base.Player) bool {
	return true
}
func (this *SceneBaseStateRichBlessed) OnEnter(s *base.Scene) {
	if sceneEx, ok := s.GetExtraData().(*RichBlessedSceneData); ok {
		sceneEx.SetStateStartTime(time.Now())
	}
}

func (this *SceneBaseStateRichBlessed) OnLeave(s *base.Scene) {}
func (this *SceneBaseStateRichBlessed) OnTick(s *base.Scene) {
	if sceneEx, ok := s.GetExtraData().(*RichBlessedSceneData); ok {
		if sceneEx.CheckNeedDestroy() {
			if s.GetRealPlayerCnt() == 0 {
				sceneEx.SceneDestroy(true)
			}
		}
	}
	if time.Now().Sub(s.GameStartTime) > time.Second*3 {
		if sceneEx, ok := s.ExtraData.(*RichBlessedSceneData); ok {
			pack := &protocol.SCRBPrize{
				PrizePool: proto.Int64(sceneEx.jackpot.GetTotalSmall() / 10000),
			}
			proto.SetDefaults(pack)
			//logger.Logger.Trace("SCRBPrize: ", pack)
			s.Broadcast(int(protocol.RBPID_PACKET_RICHBLESSED_SCRBPrize), pack, 0)
			for _, p := range sceneEx.players {
				if p.IsOnLine() {
					p.leaveTime = 0
					continue
				}

				p.leaveTime++
				if p.leaveTime < 60*2 {
					continue
				}
				//踢出玩家
				sceneEx.PlayerLeave(p.Player, common.PlayerLeaveReason_LongTimeNoOp, true)
			}
		}
		s.GameStartTime = time.Now()
	}
	if sceneEx, ok := s.ExtraData.(*RichBlessedSceneData); ok {
		//for _, p := range sceneEx.players {
		//	//游戏次数达到目标值
		//	todayGamefreeIDSceneData, _ := p.GetDaliyGameData(int(sceneEx.DbGameFree.GetId()))
		//	if !p.IsRob &&
		//		todayGamefreeIDSceneData != nil &&
		//		sceneEx.DbGameFree.GetPlayNumLimit() != 0 &&
		//		todayGamefreeIDSceneData.GameTimes >= int64(sceneEx.DbGameFree.GetPlayNumLimit()) {
		//		s.PlayerLeave(p.Player, common.PlayerLeaveReason_GameTimes, true)
		//	}
		//}
		if sceneEx.CheckNeedDestroy() {
			for _, player := range sceneEx.players {
				if !player.IsRob {
					if player.freeTimes == 0 {
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
func (this *SceneBaseStateRichBlessed) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	return false
}
func (this *SceneBaseStateRichBlessed) OnPlayerEvent(s *base.Scene, p *base.Player, evtcode int, params []int64) {
}

// ////////////////////////////////////////////////////////////
// 开始状态
// ////////////////////////////////////////////////////////////
type SceneStateStartRichBlessed struct {
	SceneBaseStateRichBlessed
}

func (this *SceneStateStartRichBlessed) GetState() int {
	return richblessed.RichBlessedStateStart
}

func (this *SceneStateStartRichBlessed) CanChangeTo(s base.SceneState) bool {
	return false
}

// 当前状态能否换桌
func (this *SceneStateStartRichBlessed) CanChangeCoinScene(s *base.Scene, p *base.Player) bool {
	if playerEx, ok := p.GetExtraData().(*RichBlessedPlayerData); ok {
		if playerEx.IsOnLine() && playerEx.freeTimes > 0 {
			return false
		}
	}
	return true
}

func (this *SceneStateStartRichBlessed) GetTimeout(s *base.Scene) int {
	return 0
}

func (this *SceneStateStartRichBlessed) OnEnter(s *base.Scene) {
	this.SceneBaseStateRichBlessed.OnEnter(s)
	if sceneEx, ok := s.GetExtraData().(*RichBlessedSceneData); ok {
		sceneEx.SetGameNowTime(time.Now())
	}
}

// 状态离开时
func (this *SceneStateStartRichBlessed) OnLeave(s *base.Scene) {
	this.SceneBaseStateRichBlessed.OnLeave(s)
	logger.Logger.Tracef("(this *SceneStateStartRichBlessed) OnLeave, sceneid=%v", s.GetSceneId())
}

// 玩家操作
func (this *SceneStateStartRichBlessed) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	logger.Logger.Tracef("(this *SceneStateStartRichBlessed) OnPlayerOp, sceneid=%v params=%v", s.GetSceneId(), params)
	if this.SceneBaseStateRichBlessed.OnPlayerOp(s, p, opcode, params) {
		return true
	}
	if sceneEx, ok := s.GetExtraData().(*RichBlessedSceneData); ok {
		if playerEx, ok := p.GetExtraData().(*RichBlessedPlayerData); ok {
			switch opcode {
			case richblessed.RichBlessedPlayerOpStart:
				playerEx.Clear()
				if len(params) > 0 {
					//只有开始算操作
					p.LastOPTimer = time.Now()
					idx := int(params[0])
					if len(sceneEx.DbGameFree.GetOtherIntParams()) <= idx {
						pack := &protocol.SCRichBlessedOp{
							OpCode:    proto.Int(opcode),
							OpRetCode: proto.Int(3),
						}
						playerEx.SendToClient(int(protocol.RBPID_PACKET_RICHBLESSED_SCRichBlessedOp), pack)
						return false
					}

					if playerEx.freeTimes > 0 {
						playerEx.gameState = richblessed.FreeGame
						playerEx.freeTimes--
						playerEx.nowFreeTimes++
					}
					//水池上下文环境
					playerEx.cpCtx = base.CoinPoolMgr.GetCoinPoolCtx(sceneEx.Platform, sceneEx.GetGameFreeId(), sceneEx.GroupId)
					if playerEx.gameState == richblessed.Normal {
						logger.Logger.Tracef("(this *SceneStateStartRichBlessed) OnPlayerOp, 下注 %v %v %v", playerEx.betCoin, playerEx.maxbetCoin, playerEx.oneBetCoin)
						playerEx.betIdx = idx
						playerEx.betCoin = int64(sceneEx.DbGameFree.GetOtherIntParams()[idx])
						maxidx := len(sceneEx.DbGameFree.GetOtherIntParams()) - 1
						playerEx.maxbetCoin = int64(sceneEx.DbGameFree.GetOtherIntParams()[maxidx])
						playerEx.oneBetCoin = playerEx.betCoin / richblessed.LineNum // 单注
						playerEx.noWinTimes++

						if playerEx.Coin < int64(s.DbGameFree.GetBetLimit()) {
							//押注限制(低于该值不能押注)
							pack := &protocol.SCRichBlessedOp{
								OpCode:    proto.Int(opcode),
								OpRetCode: proto.Int(2),
							}
							playerEx.SendToClient(int(protocol.RBPID_PACKET_RICHBLESSED_SCRichBlessedOp), pack)
							return false
						}
						if playerEx.betCoin > playerEx.Coin {
							//金币不足
							pack := &protocol.SCRichBlessedOp{
								OpCode:    proto.Int(opcode),
								OpRetCode: proto.Int(1),
							}
							playerEx.SendToClient(int(protocol.RBPID_PACKET_RICHBLESSED_SCRichBlessedOp), pack)
							return false
						}
						playerEx.CurrentBet = playerEx.betCoin
						playerEx.CurrentTax = 0
						//SysProfitCoinMgr.Add(sceneEx.sysProfitCoinKey, playerEx.betCoin, -playerEx.betCoin)
						//没有免费次数 扣钱
						p.Statics(s.KeyGameId, s.KeyGamefreeId, -playerEx.betCoin, false)
						playerEx.AddCoin(-playerEx.betCoin, common.GainWay_HundredSceneLost, 0, "system", s.GetSceneName())
						sceneEx.AddPrizeCoin(playerEx.IsRob, playerEx.betCoin) // 12%jack
						if !sceneEx.Testing {
							bet := playerEx.betCoin / 100 * 88 // 88%
							base.CoinPoolMgr.PushCoin(sceneEx.GetGameFreeId(), sceneEx.GroupId, sceneEx.Platform, bet)
						}
						playerEx.CreateResult(sceneEx.GetNormWeight())
						if !playerEx.test {
							playerEx.result.EleValue = []int32{1, 0, 1, 0, 5, 6, 2, 7, 7, 9, 4, 7, 4, 8, 7} //三锣免费
						}
					} else { // 免费游戏
						playerEx.CreateResult(sceneEx.GetFreeWeight())
					}

					sceneEx.Win(playerEx)

					if sceneEx.CanJACKPOT(playerEx, playerEx.betCoin*1000/playerEx.maxbetCoin, 1000, JACKPOTElementsParams) { // 中奖 包含CreateJACKPOT 下注不存在越界问题
						playerEx.gameState = richblessed.JackGame
						// sceneEx.JACKPOTWin(playerEx)
					}
					//发送结算
					sceneEx.SendBilled(playerEx)

					sceneEx.SaveLog(playerEx, 0)

				}
			case richblessed.RichBlessedPlayerOpSwitch:
				if len(params) > 0 && playerEx.freeTimes == 0 {
					idx := int(params[0])
					if len(sceneEx.DbGameFree.GetOtherIntParams()) > idx {
						playerEx.betIdx = idx
						playerEx.betCoin = int64(sceneEx.DbGameFree.GetOtherIntParams()[idx])
						playerEx.oneBetCoin = playerEx.betCoin / richblessed.LineNum
						pack := &protocol.SCRichBlessedOp{
							OpCode:    proto.Int(opcode),
							OpRetCode: proto.Int(4),
						}
						for i := int32(0); i != richblessed.JackMax; i++ {
							pack.Params = append(pack.Params, playerEx.oneBetCoin*richblessed.JkEleNumRate[i])
						}
						playerEx.SendToClient(int(protocol.RBPID_PACKET_RICHBLESSED_SCRichBlessedOp), pack)
						logger.Logger.Tracef("(this *SceneStateStartRichBlessed) OnPlayerOp, sceneid=%v pack=%v", s.GetSceneId(), pack)
					}
				}
			case richblessed.RichBlessedPlayerOpJack: // jack领奖 注意不要再次结算wincoin
				if playerEx.gameState != richblessed.JackGame || playerEx.JackpotEle == -1 {
					pack := &protocol.SCRichBlessedOp{
						OpCode:    proto.Int(opcode),
						OpRetCode: proto.Int(3),
					}
					playerEx.SendToClient(int(protocol.RBPID_PACKET_RICHBLESSED_SCRichBlessedOp), pack)
					return false
				}
				sceneEx.JACKPOTWin(playerEx)
				//发送结算
				sceneEx.SendJACKPOTBilled(playerEx)
				sceneEx.SaveLog(playerEx, 0)
				playerEx.Clear() // 小游戏模式结束
			}
		}
	}
	return true
}

// 玩家事件
func (this *SceneStateStartRichBlessed) OnPlayerEvent(s *base.Scene, p *base.Player, evtcode int, params []int64) {
	logger.Logger.Trace("(this *SceneStateStartRichBlessed) OnPlayerEvent, sceneId=", s.GetSceneId(), " player=", p.SnId, " evtcode=", evtcode)
	this.SceneBaseStateRichBlessed.OnPlayerEvent(s, p, evtcode, params)
	if sceneEx, ok := s.GetExtraData().(*RichBlessedSceneData); ok {
		switch evtcode {
		case base.PlayerEventEnter:
			pack := &protocol.SCRBPrize{
				PrizePool: proto.Int64(sceneEx.jackpot.GetTotalSmall() / 10000),
			}
			proto.SetDefaults(pack)
			//logger.Logger.Trace("SCRBPrize: ", pack)
			p.SendToClient(int(protocol.RBPID_PACKET_RICHBLESSED_SCRBPrize), pack)
		}
	}
}

func (this *SceneStateStartRichBlessed) OnTick(s *base.Scene) {
	this.SceneBaseStateRichBlessed.OnTick(s)
}

// //////////////////////////////////////////////////////////////////////////////
func (this *ScenePolicyRichBlessed) RegisteSceneState(state base.SceneState) {
	if state == nil {
		return
	}
	stateid := state.GetState()
	if stateid < 0 || stateid >= richblessed.RichBlessedStateMax {
		return
	}
	this.states[stateid] = state
}

func (this *ScenePolicyRichBlessed) GetSceneState(s *base.Scene, stateid int) base.SceneState {
	if stateid >= 0 && stateid < richblessed.RichBlessedStateMax {
		return this.states[stateid]
	}
	return nil
}

func init() {
	//主状态
	ScenePolicyRichBlessedSington.RegisteSceneState(&SceneStateStartRichBlessed{})
	core.RegisteHook(core.HOOK_BEFORE_START, func() error {
		base.RegisteScenePolicy(common.GameId_Richblessed, richblessed.RoomMode_Classic, ScenePolicyRichBlessedSington)
		return nil
	})
}
