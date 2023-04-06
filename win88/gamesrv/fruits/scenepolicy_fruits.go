package fruits

import (
	"time"

	"games.yol.com/win88/common"
	"games.yol.com/win88/gamerule/fruits"
	"games.yol.com/win88/gamesrv/base"
	"games.yol.com/win88/proto"
	protocol "games.yol.com/win88/protocol/fruits"
	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/logger"
)

// ////////////////////////////////////////////////////////////
var ScenePolicyFruitsSington = &ScenePolicyFruits{}

type ScenePolicyFruits struct {
	base.BaseScenePolicy
	states [fruits.FruitsStateMax]base.SceneState
}

// 创建场景扩展数据
func (this *ScenePolicyFruits) CreateSceneExData(s *base.Scene) interface{} {
	sceneEx := NewFruitsSceneData(s)
	if sceneEx != nil {
		if sceneEx.GetInit() {
			s.SetExtraData(sceneEx)
		}
	}
	return sceneEx
}

// 创建玩家扩展数据
func (this *ScenePolicyFruits) CreatePlayerExData(s *base.Scene, p *base.Player) interface{} {
	playerEx := &FruitsPlayerData{Player: p}
	p.SetExtraData(playerEx)
	return playerEx
}

// 场景开启事件
func (this *ScenePolicyFruits) OnStart(s *base.Scene) {
	logger.Logger.Trace("(this *ScenePolicyFruits) OnStart, sceneId=", s.GetSceneId())
	sceneEx := NewFruitsSceneData(s)
	if sceneEx != nil {
		if sceneEx.GetInit() {
			s.SetExtraData(sceneEx)
			s.ChangeSceneState(fruits.FruitsStateStart)
		}
	}
}

// 场景关闭事件
func (this *ScenePolicyFruits) OnStop(s *base.Scene) {
	logger.Logger.Trace("(this *ScenePolicyFruits) OnStop , sceneId=", s.GetSceneId())
}

// 场景心跳事件
func (this *ScenePolicyFruits) OnTick(s *base.Scene) {
	if s == nil {
		return
	}
	if s.GetSceneState() != nil {
		s.GetSceneState().OnTick(s)
	}
}

// 玩家进入事件
func (this *ScenePolicyFruits) OnPlayerEnter(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyFruits) OnPlayerEnter, sceneId=", s.GetSceneId(), " player=", p.Name)
	if sceneEx, ok := s.GetExtraData().(*FruitsSceneData); ok {
		playerEx := &FruitsPlayerData{Player: p}
		playerEx.init()

		sceneEx.players[p.SnId] = playerEx

		p.SetExtraData(playerEx)
		FruitsSendRoomInfo(s, sceneEx, playerEx)

		s.FirePlayerEvent(p, base.PlayerEventEnter, nil)
	}
}

// 玩家离开事件
func (this *ScenePolicyFruits) OnPlayerLeave(s *base.Scene, p *base.Player, reason int) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyFruits) OnPlayerLeave, sceneId=", s.GetSceneId(), " player=", p.SnId)
	if sceneEx, ok := s.ExtraData.(*FruitsSceneData); ok {
		s.FirePlayerEvent(p, base.PlayerEventLeave, nil)
		sceneEx.OnPlayerLeave(p, reason)
	}
}

// 玩家掉线
func (this *ScenePolicyFruits) OnPlayerDropLine(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyFruits) OnPlayerDropLine, sceneId=", s.GetSceneId(), " player=", p.SnId)
	s.FirePlayerEvent(p, base.PlayerEventDropLine, nil)
}

// 玩家重连
func (this *ScenePolicyFruits) OnPlayerRehold(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyFruits) OnPlayerRehold, sceneId=", s.GetSceneId(), " player=", p.SnId)
	if sceneEx, ok := s.GetExtraData().(*FruitsSceneData); ok {
		if playerEx, ok := p.GetExtraData().(*FruitsPlayerData); ok {
			FruitsSendRoomInfo(s, sceneEx, playerEx)
		}
	}
}

// 返回房间
func (this *ScenePolicyFruits) OnPlayerReturn(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyFruits) OnPlayerReturn, GetSceneId()=", s.GetSceneId(), " player=", p.Name)
	if sceneEx, ok := s.GetExtraData().(*FruitsSceneData); ok {
		if playerEx, ok := p.GetExtraData().(*FruitsPlayerData); ok {
			//if p.IsMarkFlag(base.PlayerState_Auto) {
			//	p.UnmarkFlag(base.PlayerState_Auto)
			//	p.SyncFlag()
			//}
			//发送房间信息给自己
			FruitsSendRoomInfo(s, sceneEx, playerEx)
			s.FirePlayerEvent(p, base.PlayerEventReturn, nil)
		}
	}
}

func FruitsSendRoomInfo(s *base.Scene, sceneEx *FruitsSceneData, playerEx *FruitsPlayerData) {
	pack := FruitsCreateRoomInfoPacket(s, sceneEx, playerEx)
	logger.Logger.Trace("RoomInfo: ", pack)
	playerEx.SendToClient(int(protocol.FruitsPID_PACKET_FRUITS_SCFruitsRoomInfo), pack)
}
func FruitsCreateRoomInfoPacket(s *base.Scene, sceneEx *FruitsSceneData, playerEx *FruitsPlayerData) interface{} {
	//房间信息
	pack := &protocol.SCFruitsRoomInfo{
		RoomId:     proto.Int(s.SceneId),
		GameId:     proto.Int(s.GameId),
		RoomMode:   proto.Int(s.SceneMode),
		SceneType:  proto.Int(s.SceneType),
		Params:     s.Params,
		NumOfGames: proto.Int(sceneEx.NumOfGames),
		State:      proto.Int(s.SceneState.GetState()),
		ParamsEx:   s.DbGameFree.OtherIntParams,
		GameFreeId: proto.Int32(s.DbGameFree.Id),
		//BetLimit:   s.DbGameFree.BetLimit,
	}

	//自己的信息
	if playerEx != nil {
		pd := &protocol.FruitsPlayerData{
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
		pack.NowGameState = proto.Int(playerEx.gameState)
		pack.BetIdx = proto.Int(playerEx.betIdx)
		pack.FreeTimes = proto.Int(playerEx.freeTimes)
		pack.MaryTimes = proto.Int(playerEx.maryFreeTimes)
		//玛丽
		pack.MaryTotalWin = proto.Int64(playerEx.winMaryCoin)
		pack.Cards = playerEx.result.EleValue
		if playerEx.maryFreeTimes > 0 && playerEx.result != nil {
			pack.MaryWinId = proto.Int32(playerEx.result.MaryOutSide)
			pack.MaryCards = playerEx.result.MaryMidArray
		}
		//免费 当局的
		if playerEx.freeTimes > 0 && playerEx.result != nil {
			pack.Cards = playerEx.result.EleValue
		}
		pack.WinLineCoin = proto.Int64(playerEx.winLineCoin)
		pack.WinJackpot = proto.Int64(playerEx.winNowJackPotCoin)
		pack.WinEle777Coin = proto.Int64(playerEx.winEle777Coin)
		pack.FreeTotalWin = proto.Int64(playerEx.winFreeCoin)
		pack.WinRate = proto.Int64(playerEx.winLineRate + playerEx.JackPot7Rate)
		pack.WinFreeTimes = proto.Int(playerEx.winFreeTimes)

		var wl []*protocol.FruitsWinLine
		for _, r := range playerEx.result.WinLine {
			wl = append(wl, &protocol.FruitsWinLine{
				Poss:   r.Poss,
				LineId: proto.Int(r.LineId),
			})
		}
		pack.WinLines = wl
	}
	proto.SetDefaults(pack)
	return pack
}
func (this *ScenePolicyFruits) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if s == nil || p == nil {
		return false
	}
	logger.Logger.Trace("(this *ScenePolicyFruits) OnPlayerOp, sceneId=", s.GetSceneId(), " player=", p.SnId, " opcode=", opcode, " params=", params)
	if s.GetSceneState() != nil {
		if s.GetSceneState().OnPlayerOp(s, p, opcode, params) {
			p.SetLastOPTimer(time.Now())
			return true
		}
		return false
	}
	return true
}

func (this *ScenePolicyFruits) OnPlayerEvent(s *base.Scene, p *base.Player, evtcode int, params []int64) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyFruits) OnPlayerEvent, sceneId=", s.GetSceneId(), " player=", p.SnId, " eventcode=", evtcode, " params=", params)
	if s.GetSceneState() != nil {
		s.GetSceneState().OnPlayerEvent(s, p, evtcode, params)
	}
}

// 当前状态能否换桌
func (this *ScenePolicyFruits) CanChangeCoinScene(s *base.Scene, p *base.Player) bool {
	if s == nil || p == nil {
		return false
	}
	if s.GetSceneState() != nil {
		return s.GetSceneState().CanChangeCoinScene(s, p)
	}
	return false
}

// 状态基类
type SceneBaseStateFruits struct {
}

func (this *SceneBaseStateFruits) GetTimeout(s *base.Scene) int {
	if sceneEx, ok := s.GetExtraData().(*FruitsSceneData); ok {
		return int(time.Now().Sub(sceneEx.GetStateStartTime()) / time.Second)
	}
	return 0
}

func (this *SceneBaseStateFruits) CanChangeTo(s base.SceneState) bool {
	return true
}

// 当前状态能否换桌
func (this *SceneBaseStateFruits) CanChangeCoinScene(s *base.Scene, p *base.Player) bool {
	return true
}
func (this *SceneBaseStateFruits) OnEnter(s *base.Scene) {
	if sceneEx, ok := s.GetExtraData().(*FruitsSceneData); ok {
		sceneEx.SetStateStartTime(time.Now())
	}
}

func (this *SceneBaseStateFruits) OnLeave(s *base.Scene) {}
func (this *SceneBaseStateFruits) OnTick(s *base.Scene) {
	if sceneEx, ok := s.GetExtraData().(*FruitsSceneData); ok {
		if sceneEx.CheckNeedDestroy() {
			if s.GetRealPlayerCnt() == 0 {
				sceneEx.SceneDestroy(true)
			}
		}
	}
	if time.Now().Sub(s.GameStartTime) > time.Second*3 {
		if sceneEx, ok := s.ExtraData.(*FruitsSceneData); ok {
			pack := &protocol.SCFruitsPrize{
				PrizePool: proto.Int64(sceneEx.jackpot.GetTotalSmall() / 10000),
			}
			proto.SetDefaults(pack)
			//logger.Logger.Trace("SCFruitsPrize: ", pack)
			s.Broadcast(int(protocol.FruitsPID_PACKET_FRUITS_SCFruitsPrize), pack, 0)
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
	if sceneEx, ok := s.ExtraData.(*FruitsSceneData); ok {
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
					if player.freeTimes == 0 && player.maryFreeTimes == 0 {
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
func (this *SceneBaseStateFruits) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	return false
}
func (this *SceneBaseStateFruits) OnPlayerEvent(s *base.Scene, p *base.Player, evtcode int, params []int64) {
}

// ////////////////////////////////////////////////////////////
// 开始状态
// ////////////////////////////////////////////////////////////
type SceneStateStartFruits struct {
	SceneBaseStateFruits
}

func (this *SceneStateStartFruits) GetState() int {
	return fruits.FruitsStateStart
}

func (this *SceneStateStartFruits) CanChangeTo(s base.SceneState) bool {
	return false
}

// 当前状态能否换桌
func (this *SceneStateStartFruits) CanChangeCoinScene(s *base.Scene, p *base.Player) bool {
	if playerEx, ok := p.GetExtraData().(*FruitsPlayerData); ok {
		if playerEx.IsOnLine() && (playerEx.freeTimes > 0 || playerEx.maryFreeTimes > 0) {
			return false
		}
	}
	return true
}

func (this *SceneStateStartFruits) GetTimeout(s *base.Scene) int {
	return 0
}

func (this *SceneStateStartFruits) OnEnter(s *base.Scene) {
	this.SceneBaseStateFruits.OnEnter(s)
	if sceneEx, ok := s.GetExtraData().(*FruitsSceneData); ok {
		sceneEx.SetGameNowTime(time.Now())
	}
}

// 状态离开时
func (this *SceneStateStartFruits) OnLeave(s *base.Scene) {
	this.SceneBaseStateFruits.OnLeave(s)
	logger.Logger.Tracef("(this *SceneStateStartFruits) OnLeave, sceneid=%v", s.GetSceneId())
}

// 玩家操作
func (this *SceneStateStartFruits) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	logger.Logger.Tracef("(this *SceneStateStartFruits) OnPlayerOp, sceneid=%v params=%v", s.GetSceneId(), params)
	if this.SceneBaseStateFruits.OnPlayerOp(s, p, opcode, params) {
		return true
	}
	if sceneEx, ok := s.GetExtraData().(*FruitsSceneData); ok {
		if playerEx, ok := p.GetExtraData().(*FruitsPlayerData); ok {
			switch opcode {
			case fruits.FruitsPlayerOpStart:
				playerEx.Clear()
				if len(params) > 0 {
					//只有开始算操作
					p.LastOPTimer = time.Now()
					idx := int(params[0])
					if len(sceneEx.DbGameFree.GetOtherIntParams()) <= idx {
						pack := &protocol.SCFruitsOp{
							OpCode:    proto.Int(opcode),
							OpRetCode: proto.Int(3),
						}
						playerEx.SendToClient(int(protocol.FruitsPID_PACKET_FRUITS_SCFruitsOp), pack)
						return false
					}
					if playerEx.maryFreeTimes > 0 {
						playerEx.gameState = fruits.MaryGame
						playerEx.maryFreeTimes--
						playerEx.nowMaryTimes++
					} else if playerEx.freeTimes > 0 {
						playerEx.gameState = fruits.FreeGame
						playerEx.freeTimes--
						playerEx.nowFreeTimes++
					}
					//水池上下文环境
					playerEx.cpCtx = base.CoinPoolMgr.GetCoinPoolCtx(sceneEx.Platform, sceneEx.GetGameFreeId(), sceneEx.GroupId)
					if playerEx.gameState == fruits.Normal {
						playerEx.betIdx = idx
						playerEx.betCoin = int64(sceneEx.DbGameFree.GetOtherIntParams()[idx])
						playerEx.oneBetCoin = playerEx.betCoin / 9
						//playerEx.isReportGameEvent = true
						playerEx.noWinTimes++

						if playerEx.Coin < int64(s.DbGameFree.GetBetLimit()) {
							//押注限制(低于该值不能押注)
							pack := &protocol.SCFruitsOp{
								OpCode:    proto.Int(opcode),
								OpRetCode: proto.Int(2),
							}
							playerEx.SendToClient(int(protocol.FruitsPID_PACKET_FRUITS_SCFruitsOp), pack)
							return false
						}
						if playerEx.betCoin > playerEx.Coin {
							//金币不足
							pack := &protocol.SCFruitsOp{
								OpCode:    proto.Int(opcode),
								OpRetCode: proto.Int(1),
							}
							playerEx.SendToClient(int(protocol.FruitsPID_PACKET_FRUITS_SCFruitsOp), pack)
							return false
						}
						playerEx.CurrentBet = playerEx.betCoin
						playerEx.CurrentTax = 0
						//SysProfitCoinMgr.Add(sceneEx.sysProfitCoinKey, playerEx.betCoin, -playerEx.betCoin)
						//没有免费次数 扣钱
						p.Statics(s.KeyGameId, s.KeyGamefreeId, -playerEx.betCoin, false)
						playerEx.AddCoin(-playerEx.betCoin, common.GainWay_HundredSceneLost, 0, "system", s.GetSceneName())
						sceneEx.AddPrizeCoin(playerEx.IsRob, playerEx.betCoin)
						if !sceneEx.Testing {
							base.CoinPoolMgr.PushCoin(sceneEx.GetGameFreeId(), sceneEx.GroupId, sceneEx.Platform, playerEx.betCoin)
						}
					}
					if !s.Testing {
						//sceneEx.Regulation(playerEx, 0)
					}
					if playerEx.gameState == fruits.MaryGame {
						playerEx.result.CreateMary(sceneEx.GetMaryWeight())
						playerEx.CreateMary()
					} else if playerEx.gameState == fruits.FreeGame {
						playerEx.CreateResult(sceneEx.GetFreeWeight())
					} else {
						playerEx.CreateResult(sceneEx.GetNormWeight())
					}

					sceneEx.Win(playerEx)
					//发送结算
					sceneEx.SendBilled(playerEx)

					//sceneEx.SaveLog(playerEx, 0)

					if playerEx.gameState == fruits.Normal {
						playerEx.winJackPotCoin = 0
						playerEx.winEle777Coin = 0
						playerEx.winLineCoin = 0
					}
					playerEx.wl = nil
				}
			case fruits.FruitsPlayerOpSwitch:
				if len(params) > 0 && playerEx.freeTimes == 0 && playerEx.maryFreeTimes == 0 {
					idx := int(params[0])
					if len(sceneEx.DbGameFree.GetOtherIntParams()) > idx {
						playerEx.betIdx = idx
						playerEx.betCoin = int64(sceneEx.DbGameFree.GetOtherIntParams()[idx])
						playerEx.oneBetCoin = playerEx.betCoin / 9
					}
				}
			}
		}
	}
	return true
}

// 玩家事件
func (this *SceneStateStartFruits) OnPlayerEvent(s *base.Scene, p *base.Player, evtcode int, params []int64) {
	logger.Logger.Trace("(this *SceneStateStartFruits) OnPlayerEvent, sceneId=", s.GetSceneId(), " player=", p.SnId, " evtcode=", evtcode)
	this.SceneBaseStateFruits.OnPlayerEvent(s, p, evtcode, params)
	if sceneEx, ok := s.GetExtraData().(*FruitsSceneData); ok {
		switch evtcode {
		case base.PlayerEventEnter:
			pack := &protocol.SCFruitsPrize{
				PrizePool: proto.Int64(sceneEx.jackpot.GetTotalSmall() / 10000),
			}
			proto.SetDefaults(pack)
			//logger.Logger.Trace("SCFruitsPrize: ", pack)
			p.SendToClient(int(protocol.FruitsPID_PACKET_FRUITS_SCFruitsPrize), pack)
		}
	}
}

func (this *SceneStateStartFruits) OnTick(s *base.Scene) {
	this.SceneBaseStateFruits.OnTick(s)
}

// //////////////////////////////////////////////////////////////////////////////
func (this *ScenePolicyFruits) RegisteSceneState(state base.SceneState) {
	if state == nil {
		return
	}
	stateid := state.GetState()
	if stateid < 0 || stateid >= fruits.FruitsStateMax {
		return
	}
	this.states[stateid] = state
}

func (this *ScenePolicyFruits) GetSceneState(s *base.Scene, stateid int) base.SceneState {
	if stateid >= 0 && stateid < fruits.FruitsStateMax {
		return this.states[stateid]
	}
	return nil
}

func init() {
	//主状态
	ScenePolicyFruitsSington.RegisteSceneState(&SceneStateStartFruits{})
	core.RegisteHook(core.HOOK_BEFORE_START, func() error {
		base.RegisteScenePolicy(common.GameId_Fruits, fruits.RoomMode_Classic, ScenePolicyFruitsSington)
		return nil
	})
}
