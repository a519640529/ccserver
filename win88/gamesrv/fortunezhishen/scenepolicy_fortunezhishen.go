package fortunezhishen

import (
	"fmt"
	"games.yol.com/win88/common"
	gamerule "games.yol.com/win88/gamerule/fortunezhishen"
	"games.yol.com/win88/gamesrv/base"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/fortunezhishen"
	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/logger"
	"os"
	"time"
)

//////////////////////////////////////////////////////////////
var ScenePolicyFortuneZhiShenSington = &ScenePolicyFortuneZhiShen{}

type ScenePolicyFortuneZhiShen struct {
	base.BaseScenePolicy
	states [gamerule.FortuneZhiShenStateMax]base.SceneState
}

//创建场景扩展数据
func (this *ScenePolicyFortuneZhiShen) CreateSceneExData(s *base.Scene) interface{} {
	sceneEx := NewFortuneZhiShenSceneData(s)
	if sceneEx != nil {
		if sceneEx.GetInit() {
			s.SetExtraData(sceneEx)
		}
	}
	return sceneEx
}

//创建玩家扩展数据
func (this *ScenePolicyFortuneZhiShen) CreatePlayerExData(s *base.Scene, p *base.Player) interface{} {
	playerEx := &FortuneZhiShenPlayerData{Player: p}
	p.SetExtraData(playerEx)
	return playerEx
}

//场景开启事件
func (this *ScenePolicyFortuneZhiShen) OnStart(s *base.Scene) {
	logger.Logger.Trace("(this *ScenePolicyFortuneZhiShen) OnStart, sceneId=", s.GetSceneId())
	sceneEx := NewFortuneZhiShenSceneData(s)
	if sceneEx != nil {
		if sceneEx.GetInit() {
			s.SetExtraData(sceneEx)
			s.ChangeSceneState(gamerule.FortuneZhiShenStateStart)
		}
	}
}

//场景关闭事件
func (this *ScenePolicyFortuneZhiShen) OnStop(s *base.Scene) {
	logger.Logger.Trace("(this *ScenePolicyFortuneZhiShen) OnStop , sceneId=", s.GetSceneId())
}

//场景心跳事件
func (this *ScenePolicyFortuneZhiShen) OnTick(s *base.Scene) {
	if s == nil {
		return
	}
	if s.GetSceneState() != nil {
		s.GetSceneState().OnTick(s)
	}
}

//玩家进入事件
func (this *ScenePolicyFortuneZhiShen) OnPlayerEnter(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyFortuneZhiShen) OnPlayerEnter, sceneId=", s.GetSceneId(), " player=", p.Name)
	if sceneEx, ok := s.GetExtraData().(*FortuneZhiShenSceneData); ok {
		playerEx := &FortuneZhiShenPlayerData{Player: p}
		playerEx.init()

		sceneEx.players[p.SnId] = playerEx

		p.SetExtraData(playerEx)
		FortuneZhiShenSendRoomInfo(s, sceneEx, playerEx)

		s.FirePlayerEvent(p, base.PlayerEventEnter, nil)
	}
}

//玩家离开事件
func (this *ScenePolicyFortuneZhiShen) OnPlayerLeave(s *base.Scene, p *base.Player, reason int) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyFortuneZhiShen) OnPlayerLeave, sceneId=", s.GetSceneId(), " player=", p.SnId)
	if sceneEx, ok := s.GetExtraData().(*FortuneZhiShenSceneData); ok {
		s.FirePlayerEvent(p, base.PlayerEventLeave, nil)
		sceneEx.OnPlayerLeave(p, reason)
	}
}

//玩家掉线
func (this *ScenePolicyFortuneZhiShen) OnPlayerDropLine(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyFortuneZhiShen) OnPlayerDropLine, sceneId=", s.GetSceneId(), " player=", p.SnId)
	s.FirePlayerEvent(p, base.PlayerEventDropLine, nil)
}

//玩家重连
func (this *ScenePolicyFortuneZhiShen) OnPlayerRehold(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyFortuneZhiShen) OnPlayerRehold, sceneId=", s.GetSceneId(), " player=", p.SnId)
	if sceneEx, ok := s.GetExtraData().(*FortuneZhiShenSceneData); ok {
		if playerEx, ok := p.GetExtraData().(*FortuneZhiShenPlayerData); ok {
			FortuneZhiShenSendRoomInfo(s, sceneEx, playerEx)
		}
	}
}
func FortuneZhiShenSendRoomInfo(s *base.Scene, sceneEx *FortuneZhiShenSceneData, playerEx *FortuneZhiShenPlayerData) {
	pack := FortuneZhiShenCreateRoomInfoPacket(s, sceneEx, playerEx)
	logger.Logger.Trace("RoomInfo: ", pack)
	playerEx.SendToClient(int(fortunezhishen.FortuneZSPacketID_PACKET_SC_FORTUNEZHISHEN_ROOMINFO), pack)

	sceneEx.SendPrize(playerEx)
}
func FortuneZhiShenCreateRoomInfoPacket(s *base.Scene, sceneEx *FortuneZhiShenSceneData, playerEx *FortuneZhiShenPlayerData) interface{} {
	//房间信息
	pack := &fortunezhishen.SCFortuneZhiShenRoomInfo{
		RoomId:     proto.Int(s.GetSceneId()),
		Creator:    proto.Int32(s.GetCreator()),
		GameId:     proto.Int(s.GetGameId()),
		RoomMode:   proto.Int(s.GetSceneMode()),
		AgentId:    proto.Int32(s.GetAgentor()),
		SceneType:  proto.Int(int(s.DbGameFree.GetSceneType())),
		Params:     s.GetParams(),
		NumOfGames: proto.Int(sceneEx.GetNumOfGames()),
		State:      proto.Int(s.GetSceneState().GetState()),
		DisbandGen: proto.Int(sceneEx.GetDisbandGen()),
		ParamsEx:   s.GetParamsEx(),
		BetLimit:   s.GetDBGameFree().BetLimit,
		GameFreeId: proto.Int32(s.GetDBGameFree().GetId()),
	}
	//自己的信息
	if playerEx != nil {
		//if !playerEx.IsHaveFreeTimes() {
		//	playerEx.Clear()
		//}
		pd := &fortunezhishen.FortuneZhiShenPlayerData{
			SnId:        proto.Int32(playerEx.SnId),
			Name:        proto.String(playerEx.Name),
			Head:        proto.Int32(playerEx.Head),
			Sex:         proto.Int32(playerEx.Sex),
			Coin:        proto.Int64(playerEx.GetCoin()),
			Pos:         proto.Int(playerEx.GetPos()),
			Flag:        proto.Int(playerEx.GetFlag()),
			City:        proto.String(playerEx.GetCity()),
			HeadOutLine: proto.Int32(playerEx.HeadOutLine),
			VIP:         proto.Int32(playerEx.VIP),
		}
		pack.Players = append(pack.Players, pd)
		pack.TotalChipIdx = proto.Int(playerEx.betIdx)
		pack.FirstFreeTimes = proto.Int(playerEx.firstFreeTimes)
		pack.SecondFreeTimes = proto.Int(playerEx.secondFreeTimes)
		pack.UiShow = playerEx.result.Cards
		pack.GemstoneRateCoin = playerEx.gemstoneRateCoin
		pack.NowGameState = proto.Int(playerEx.gameState)

		pack.WinCoin = proto.Int64(playerEx.winCoin)
		pack.FirstWinCoin = proto.Int64(playerEx.firstWinCoin)
		pack.SecondWinCOin = proto.Int64(playerEx.secondWinCoin)
		//if !playerEx.IsHaveFreeTimes() {
		//	pack.UiShow = playerEx.normalCards
		//}
	}
	proto.SetDefaults(pack)
	return pack
}
func (this *ScenePolicyFortuneZhiShen) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if s == nil || p == nil {
		return false
	}
	logger.Logger.Trace("(this *ScenePolicyFortuneZhiShen) OnPlayerOp, sceneId=", s.GetSceneId(), " player=", p.SnId, " opcode=", opcode, " params=", params)
	if s.GetSceneState() != nil {
		if s.GetSceneState().OnPlayerOp(s, p, opcode, params) {
			p.SetLastOPTimer(time.Now())
			return true
		}
		return false
	}
	return true
}

func (this *ScenePolicyFortuneZhiShen) OnPlayerEvent(s *base.Scene, p *base.Player, evtcode int, params []int64) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyFortuneZhiShen) OnPlayerEvent, sceneId=", s.GetSceneId(), " player=", p.SnId, " eventcode=", evtcode, " params=", params)
	if s.GetSceneState() != nil {
		s.GetSceneState().OnPlayerEvent(s, p, evtcode, params)
	}
}

//当前状态能否换桌
func (this *ScenePolicyFortuneZhiShen) CanChangeCoinScene(s *base.Scene, p *base.Player) bool {
	if s == nil || p == nil {
		return false
	}
	if s.GetSceneState() != nil {
		return s.GetSceneState().CanChangeCoinScene(s, p)
	}
	return false
}

//状态基类
type SceneBaseStateFortuneZhiShen struct {
}

func (this *SceneBaseStateFortuneZhiShen) GetTimeout(s *base.Scene) int {
	if sceneEx, ok := s.GetExtraData().(*FortuneZhiShenSceneData); ok {
		return int(time.Now().Sub(sceneEx.GetStateStartTime()) / time.Second)
	}
	return 0
}

func (this *SceneBaseStateFortuneZhiShen) CanChangeTo(s base.SceneState) bool {
	return true
}

//当前状态能否换桌
func (this *SceneBaseStateFortuneZhiShen) CanChangeCoinScene(s *base.Scene, p *base.Player) bool {
	return true
}
func (this *SceneBaseStateFortuneZhiShen) OnEnter(s *base.Scene) {
	if sceneEx, ok := s.GetExtraData().(*FortuneZhiShenSceneData); ok {
		sceneEx.SetStateStartTime(time.Now())
	}
}

func (this *SceneBaseStateFortuneZhiShen) OnLeave(s *base.Scene) {}
func (this *SceneBaseStateFortuneZhiShen) OnTick(s *base.Scene) {
	if time.Now().Sub(s.GetGameStartTime()) > time.Second*3 {
		if sceneEx, ok := s.GetExtraData().(*FortuneZhiShenSceneData); ok {
			var grandPrize, bigPrize, midPrize, smallPrize []int64
			for i := 0; i < len(sceneEx.GetDBGameFree().GetOtherIntParams()); i++ {
				grandPrize = append(grandPrize, sceneEx.jackpot.HugePools[i]/gamerule.NowByte)
				bigPrize = append(bigPrize, sceneEx.jackpot.BigPools[i]/gamerule.NowByte)
				midPrize = append(midPrize, sceneEx.jackpot.MiddlePools[i]/gamerule.NowByte)
				smallPrize = append(smallPrize, sceneEx.jackpot.SmallPools[i]/gamerule.NowByte)
			}
			pack := &fortunezhishen.SCFortuneZhiShenPrize{
				GrandPrize: grandPrize,
				BigPrize:   bigPrize,
				MidPrize:   midPrize,
				SmallPrize: smallPrize,
			}
			proto.SetDefaults(pack)
			//logger.Logger.Trace("SCFortuneZhiShenPrize: ", pack)
			s.Broadcast(int(fortunezhishen.FortuneZSPacketID_PACKET_SC_FORTUNEZHISHEN_PRIZE), pack, 0)
			for _, p := range sceneEx.players {
				if p.IsOnLine() {
					p.leaveTime = 0
					continue
				}
				p.leaveTime++
				if p.leaveTime < 60/3*5 {
					continue
				}
				if p.IsHaveFreeTimes() {
					if p.secondFreeTimes > 0 {
						p.gameState = gamerule.StopAndRotate
						if p.result.MidIcon != -1 {
							p.gameState = gamerule.StopAndRotate2
						}
					} else if p.firstFreeTimes > 0 {
						p.gameState = gamerule.FreeGame
					}
					p.firstFreeTimes = 0
					p.secondFreeTimes = 0
					p.result.MakeAFortuneNum = 0
					p.result.NewAddGemstone = 0
					p.cpCtx = base.GetCoinPoolMgr().GetCoinPoolCtx(sceneEx.GetPlatform(), sceneEx.GetGameFreeId(), sceneEx.GetGroupId())
					sceneEx.Win(p)
					sceneEx.SaveLog(p, 1)
				} else {
					//踢出玩家
					sceneEx.PlayerLeave(p.Player, common.PlayerLeaveReason_LongTimeNoOp, true)
				}
			}
		}
		s.SetGameStartTime(time.Now())
	}
	if sceneEx, ok := s.GetExtraData().(*FortuneZhiShenSceneData); ok {
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
					if player.firstFreeTimes == 0 && player.secondFreeTimes == 0 {
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
func (this *SceneBaseStateFortuneZhiShen) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	return false
}
func (this *SceneBaseStateFortuneZhiShen) OnPlayerEvent(s *base.Scene, p *base.Player, evtcode int, params []int64) {
}

//////////////////////////////////////////////////////////////
//开始状态
//////////////////////////////////////////////////////////////
type SceneStateStartFortuneZhiShen struct {
	SceneBaseStateFortuneZhiShen
}

func (this *SceneStateStartFortuneZhiShen) GetState() int {
	return gamerule.FortuneZhiShenStateStart
}

func (this *SceneStateStartFortuneZhiShen) CanChangeTo(s base.SceneState) bool {
	return false
}

//当前状态能否换桌
func (this *SceneStateStartFortuneZhiShen) CanChangeCoinScene(s *base.Scene, p *base.Player) bool {
	if playerEx, ok := p.GetExtraData().(*FortuneZhiShenPlayerData); ok {
		if playerEx.IsOnLine() && playerEx.IsHaveFreeTimes() {
			return false
		}
	}
	return true
}

func (this *SceneStateStartFortuneZhiShen) GetTimeout(s *base.Scene) int {
	return 0
}

func (this *SceneStateStartFortuneZhiShen) OnEnter(s *base.Scene) {
	this.SceneBaseStateFortuneZhiShen.OnEnter(s)
	if sceneEx, ok := s.GetExtraData().(*FortuneZhiShenSceneData); ok {
		sceneEx.SetGameNowTime(time.Now())
	}
}

//状态离开时
func (this *SceneStateStartFortuneZhiShen) OnLeave(s *base.Scene) {
	this.SceneBaseStateFortuneZhiShen.OnLeave(s)
	logger.Logger.Tracef("(this *SceneStateStartFortuneZhiShen) OnLeave, sceneid=%v", s.GetSceneId())
}

var fortunezhishenbenchtesttimes int

func (sp *SceneStateStartFortuneZhiShen) BenchTest(s *base.Scene, p *base.Player) {
	type LogItem struct {
		time          string
		snid          int32
		currentCoin   int64
		in            int64
		out           int64
		lineNum       int32
		score         int32
		leftFreeTimes int32
		pool          int64
		gameMode      int
	}

	oldPoolCoin := base.GetCoinPoolMgr().LoadCoin(s.GetGameFreeId(), s.GetPlatform(), s.GetGroupId())
	if fortunezhishenbenchtesttimes == 0 {
		if oldPoolCoin != 20000 {
			base.GetCoinPoolMgr().PushCoin(s.GetGameFreeId(), s.GetGroupId(), s.GetPlatform(), 20000-oldPoolCoin)
		}
	}
	old := p.GetCoin()
	p.SetCoin(20000)
	//pgi := p.GetGameFreeIdData(strconv.Itoa(int(s.GetGameFreeId())))
	fortunezhishenbenchtesttimes++
	fileName := fmt.Sprintf("fortunezhishen-%v-%d.csv", p.SnId, fortunezhishenbenchtesttimes)
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm)
	defer file.Close()
	if err != nil {
		file, err = os.Create(fileName)
		if err != nil {
			return
		}
	}

	file.WriteString("玩家id,当前余额,投入,产出,中奖倍数,中奖线数,剩余免费次数,当前水位,游戏模式\r\n")
	if playerEx, ok := p.GetExtraData().(*FortuneZhiShenPlayerData); ok {
		for i := 0; i < BENCH_CNT; i++ {
			log := &LogItem{
				snid:          p.SnId,
				currentCoin:   p.GetCoin(),
				leftFreeTimes: int32(playerEx.firstFreeTimes),
				pool:          base.GetCoinPoolMgr().LoadCoin(s.GetGameFreeId(), s.GetPlatform(), s.GetGroupId()),
			}
			playerEx.UnmarkFlag(base.PlayerState_GameBreak)
			suc := sp.OnPlayerOp(s, p, gamerule.FortuneZhiShenPlayerOpStart, []int64{0})
			log.lineNum = int32(len(playerEx.result.WinLine))
			log.score = int32(playerEx.winTotalRate)
			changeCoin := playerEx.nowGetCoin - playerEx.betCoin
			if changeCoin > 0 {
				log.out = changeCoin
			} else {
				log.in = -(changeCoin)
			}
			log.gameMode = playerEx.gameState
			str := fmt.Sprintf("%v,%v,%v,%v,%v,%v,%v,%v,%v\r\n", p.SnId, log.currentCoin, log.in, log.out, log.score,
				log.lineNum, log.leftFreeTimes, log.pool, log.gameMode)
			file.WriteString(str)
			if !suc {
				break
			}
		}
	}

	p.SetCoin(old)
}

//玩家操作
func (this *SceneStateStartFortuneZhiShen) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	logger.Logger.Tracef("(this *SceneStateStartFortuneZhiShen) OnPlayerOp, sceneid=%v params=%v", s.GetSceneId(), params)
	if this.SceneBaseStateFortuneZhiShen.OnPlayerOp(s, p, opcode, params) {
		return true
	}
	if sceneEx, ok := s.GetExtraData().(*FortuneZhiShenSceneData); ok {
		if playerEx, ok := p.GetExtraData().(*FortuneZhiShenPlayerData); ok {
			switch opcode {
			case gamerule.FortuneZhiShenPlayerOpStart:
				playerEx.Clear()
				if len(params) == 0 {
					return true
				}
				//test code
				//if !playerEx.isF {
				//	playerEx.isF = true
				//	this.BenchTest(s, p)
				//	return true
				//}

				//只有开始算操作
				p.SetLastOPTimer(time.Now())

				idx := int(params[0])
				if len(sceneEx.GetDBGameFree().GetOtherIntParams()) <= idx {
					pack := &fortunezhishen.SCFortuneZhiShenOp{
						OpCode:    proto.Int(opcode),
						OpRetCode: proto.Int(3),
					}
					playerEx.SendToClient(int(fortunezhishen.FortuneZSPacketID_PACKET_SC_FORTUNEZHISHEN_PLAYEROP), pack)
					return false
				}

				//先减去免费次数 便于后边结算
				if playerEx.secondFreeTimes > 0 {
					playerEx.gameState = gamerule.StopAndRotate
					playerEx.secondFreeTimes--
					playerEx.nowSecondTimes++
					if playerEx.result.MidIcon != -1 {
						playerEx.gameState = gamerule.StopAndRotate2
					}
				} else if playerEx.firstFreeTimes > 0 {
					playerEx.gameState = gamerule.FreeGame
					playerEx.firstFreeTimes--
					playerEx.nowFirstTimes++
				}

				//水池上下文环境
				playerEx.cpCtx = base.GetCoinPoolMgr().GetCoinPoolCtx(sceneEx.GetPlatform(), sceneEx.GetGameFreeId(), sceneEx.GetGroupId())

				if playerEx.gameState == gamerule.Normal {
					playerEx.betIdx = idx
					playerEx.betCoin = int64(sceneEx.GetDBGameFree().GetOtherIntParams()[idx])
					playerEx.oneBetCoin = playerEx.betCoin / 50
					playerEx.noWinTimes++
					if playerEx.GetCoin() < int64(s.GetDBGameFree().GetBetLimit()) {
						//押注限制(低于该值不能押注)
						pack := &fortunezhishen.SCFortuneZhiShenOp{
							OpCode:    proto.Int(opcode),
							OpRetCode: proto.Int(2),
						}
						playerEx.SendToClient(int(fortunezhishen.FortuneZSPacketID_PACKET_SC_FORTUNEZHISHEN_PLAYEROP), pack)
						return false
					}
					if playerEx.betCoin > playerEx.GetCoin() {
						//金币不足
						pack := &fortunezhishen.SCFortuneZhiShenOp{
							OpCode:    proto.Int(opcode),
							OpRetCode: proto.Int(1),
						}
						playerEx.SendToClient(int(fortunezhishen.FortuneZSPacketID_PACKET_SC_FORTUNEZHISHEN_PLAYEROP), pack)
						return false
					}
					playerEx.SetCurrentBet(playerEx.betCoin)
					playerEx.SetCurrentTax(0)
					//没有免费次数 扣钱
					//SysProfitCoinMgr.Add(sceneEx.sysProfitCoinKey, playerEx.betCoin, -playerEx.betCoin)
					p.Statics(s.GetKeyGameId(), s.KeyGamefreeId, -playerEx.betCoin, false)
					playerEx.AddCoin(-playerEx.betCoin, common.GainWay_HundredSceneLost, 0, "system", s.GetSceneName())
					sceneEx.AddPrizeCoin(playerEx.betCoin, idx)
					if !sceneEx.GetTesting() {
						base.GetCoinPoolMgr().PushCoin(sceneEx.GetGameFreeId(), sceneEx.GetGroupId(), sceneEx.GetPlatform(), playerEx.betCoin)
					}
				}
				if !sceneEx.GetTesting() {
					sceneEx.Regulation(playerEx, 0)
				}
				sceneEx.CreateResult(playerEx)

				sceneEx.Win(playerEx)

				sceneEx.SaveLog(playerEx, 0)

				sceneEx.Billed(playerEx)

				sceneEx.SendPrize(playerEx)

				playerEx.wl = nil
			case gamerule.FortuneZhiShenPlayerOpSwitch:
				if len(params) > 0 && playerEx.firstFreeTimes == 0 && playerEx.secondFreeTimes == 0 {
					idx := int(params[0])
					if len(sceneEx.GetDBGameFree().GetOtherIntParams()) > idx {
						playerEx.betIdx = idx
						playerEx.betCoin = int64(sceneEx.GetDBGameFree().GetOtherIntParams()[idx])
						playerEx.oneBetCoin = playerEx.betCoin / 50
					}
				}
			}
		}
	}
	return true
}

//玩家事件
func (this *SceneStateStartFortuneZhiShen) OnPlayerEvent(s *base.Scene, p *base.Player, evtcode int, params []int64) {
	logger.Logger.Trace("(this *SceneStateStartFortuneZhiShen) OnPlayerEvent, sceneId=", s.GetSceneId(), " player=", p.SnId, " evtcode=", evtcode)
	this.SceneBaseStateFortuneZhiShen.OnPlayerEvent(s, p, evtcode, params)
	if playerEx, ok := p.GetExtraData().(*FortuneZhiShenPlayerData); ok {
		if sceneEx, ok := s.GetExtraData().(*FortuneZhiShenSceneData); ok {
			switch evtcode {
			case base.PlayerEventEnter:
				sceneEx.SendRoomState(playerEx)
			}
		}
	}
}

func (this *SceneStateStartFortuneZhiShen) OnTick(s *base.Scene) {
	this.SceneBaseStateFortuneZhiShen.OnTick(s)
}

////////////////////////////////////////////////////////////////////////////////
func (this *ScenePolicyFortuneZhiShen) RegisteSceneState(state base.SceneState) {
	if state == nil {
		return
	}
	stateid := state.GetState()
	if stateid < 0 || stateid >= gamerule.FortuneZhiShenStateMax {
		return
	}
	this.states[stateid] = state
}

//
func (this *ScenePolicyFortuneZhiShen) GetSceneState(s *base.Scene, stateid int) base.SceneState {
	if stateid >= 0 && stateid < gamerule.FortuneZhiShenStateMax {
		return this.states[stateid]
	}
	return nil
}

////
//func (this *ScenePolicyFortuneZhiShen) GetGameSubState(s *base.Scene, stateid int) base.SceneGamingSubState {
//	return nil
//}
func init() {
	//主状态
	ScenePolicyFortuneZhiShenSington.RegisteSceneState(&SceneStateStartFortuneZhiShen{})
	core.RegisteHook(core.HOOK_BEFORE_START, func() error {
		//base.RegisteScenePolicy(common.GameId_FortuneZhiShen, gamerule.RoomMode_Classic, ScenePolicyFortuneZhiShenSington)
		return nil
	})
}
