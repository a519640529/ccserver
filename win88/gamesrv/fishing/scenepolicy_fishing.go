package fishing

import (
	"games.yol.com/win88/common"
	"games.yol.com/win88/gamerule/fishing"
	"games.yol.com/win88/gamesrv/base"
	"games.yol.com/win88/proto"
	fishing_proto "games.yol.com/win88/protocol/fishing"
	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/timer"
	"math"
	"time"
)

var ScenePolicyFishingSington = &ScenePolicyFishing{}

type ScenePolicyFishing struct {
	base.BaseScenePolicy
	states [FishingSceneStateMax]base.SceneState
}

func (this *ScenePolicyFishing) CreateSceneExData(s *base.Scene) interface{} {
	sceneEx := NewFishingSceneData(s)
	if sceneEx != nil {
		if sceneEx.init() {
			s.SetExtraData(sceneEx)
		}
	}
	return sceneEx
}
func (this *ScenePolicyFishing) CreatePlayerExData(s *base.Scene, p *base.Player) interface{} {
	playerEx := &FishingPlayerData{Player: p}
	if playerEx != nil {
		p.SetExtraData(playerEx)
	}
	return playerEx
}
func (this *ScenePolicyFishing) OnStart(s *base.Scene) {
	fishlogger.Trace("(this *ScenePolicyFishing) OnStart, sceneId=", s.GetSceneId())
	sceneEx := NewFishingSceneData(s) // 初始化当前场景的数据接口
	if sceneEx != nil {
		if sceneEx.init() {
			s.SetExtraData(sceneEx)
			s.ChangeSceneState(FishingSceneStateStart)
		}
	}
}
func (this *ScenePolicyFishing) OnTick(s *base.Scene) {
	if s == nil {
		return
	}
	if s.GetSceneState() != nil {
		s.GetSceneState().OnTick(s)
	}
}
func (this *ScenePolicyFishing) GetPlayerNum(s *base.Scene) int32 {
	if s == nil {
		return 0
	}
	if sceneEx, ok := s.GetExtraData().(*FishingSceneData); ok {
		return int32(len(sceneEx.players))
	}
	return 0
}
func (this *ScenePolicyFishing) OnPlayerEnter(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	fishlogger.Trace("(this *ScenePolicyFishing) OnPlayerEnter, sceneId=", s.GetSceneId(), " player=", p.SnId)
	if sceneEx, ok := s.GetExtraData().(*FishingSceneData); ok {
		sceneEx.SetGameNowTime(time.Now())
		playerEx := &FishingPlayerData{Player: p}
		playerEx.init(s)
		playerEx.taxCoin = 0
		playerEx.sTaxCoin = 0
		playerEx.winCoin = 0
		playerEx.enterTime = time.Now()
		sceneEx.EnterPlayer(playerEx)
		s.FirePlayerEvent(p, base.PlayerEventEnter, nil)
	}
}
func (this *ScenePolicyFishing) OnPlayerLeave(s *base.Scene, p *base.Player, reason int) {
	if s == nil || p == nil {
		return
	}
	fishlogger.Trace("(this *ScenePolicyFishing) OnPlayerLeave, sceneId=", s.GetSceneId(), " player=", p.Name)
	if sceneEx, ok := s.GetExtraData().(*FishingSceneData); ok {
		if this.CanChangeCoinScene(s, p) {
			if playerEx, ok := p.GetExtraData().(*FishingPlayerData); ok {
				playerEx.SaveDetailedLog(s)
				sceneEx.QuitPlayer(playerEx, reason)
				s.FirePlayerEvent(p, base.PlayerEventLeave, []int64{int64(reason)})
				if playerEx.GetCurrentCoin() == 0 {
					playerEx.SetCurrentCoin(playerEx.GetTakeCoin())
				}
				playerEx.SaveSceneCoinLog(playerEx.GetCurrentCoin(), int64(playerEx.CoinCache-playerEx.GetCurrentCoin()),
					playerEx.GetCoin(), 0, int64(math.Floor(playerEx.taxCoin+0.5)), playerEx.winCoin, 0, 0)
				playerEx.SetCurrentCoin(playerEx.GetCoin())
			}
		}
	}
}
func (this *ScenePolicyFishing) OnPlayerDropLine(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	fishlogger.Trace("(this *ScenePolicyFishing) OnPlayerDropLine, sceneId=", s.GetSceneId(), " player=", p.Name)
	s.FirePlayerEvent(p, base.PlayerEventDropLine, nil)
	if _, ok := s.GetExtraData().(*FishingSceneData); ok {
		if _, ok := p.GetExtraData().(*FishingPlayerData); ok {
		}
	}
}
func (this *ScenePolicyFishing) OnPlayerRehold(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	fishlogger.Trace("(this *ScenePolicyFishing) OnPlayerRehold, sceneId=", s.GetSceneId(), " player=", p.SnId)
	if _, ok := s.GetExtraData().(*FishingSceneData); ok {
		if _, ok := p.GetExtraData().(*FishingPlayerData); ok {
			s.FirePlayerEvent(p, base.PlayerEventRehold, nil)
		}
	}
}
func (this *ScenePolicyFishing) OnPlayerReturn(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	fishlogger.Trace("(this *ScenePolicyFishing) OnPlayerReturn, sceneId=", s.GetSceneId(), " player=", p.SnId)
	if _, ok := s.GetExtraData().(*FishingSceneData); ok {
		if _, ok := p.GetExtraData().(*FishingPlayerData); ok {
			s.FirePlayerEvent(p, base.PlayerEventReturn, nil)
		}
	}
}
func (this *ScenePolicyFishing) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if s == nil || p == nil {
		return false
	}
	if s.GetSceneState() != nil {
		return s.GetSceneState().OnPlayerOp(s, p, opcode, params)
	}
	return true
}
func (this *ScenePolicyFishing) OnPlayerEvent(s *base.Scene, p *base.Player, evtcode int, params []int64) {
	if s == nil || p == nil {
		return
	}
	if s.GetSceneState() != nil {
		s.GetSceneState().OnPlayerEvent(s, p, evtcode, params)
	}
}
func (this *ScenePolicyFishing) IsCompleted(s *base.Scene) bool     { return false }
func (this *ScenePolicyFishing) IsCanForceStart(s *base.Scene) bool { return true }
func (this *ScenePolicyFishing) CanChangeCoinScene(s *base.Scene, p *base.Player) bool {
	if s == nil || p == nil {
		return true
	}
	if s.GetSceneState() != nil {
		return s.GetSceneState().CanChangeCoinScene(s, p)
	}
	return true
}
func FishingSendRoomInfo(p *base.Player, sceneEx *FishingSceneData) {
	pack := &fishing_proto.SCFishingRoomInfo{
		RoomId:     proto.Int(sceneEx.GetSceneId()),
		Creator:    proto.Int32(sceneEx.GetCreator()),
		GameId:     proto.Int(sceneEx.gameId),
		RoomMode:   proto.Int(sceneEx.sceneMode),
		AgentId:    proto.Int32(sceneEx.agentor),
		SceneType:  proto.Int(sceneEx.sceneType),
		Params:     sceneEx.GetParams(),
		NumOfGames: proto.Int(sceneEx.GetNumOfGames()),
		State:      proto.Int(sceneEx.GetSceneState().GetState()),
		TimeOut:    proto.Int(sceneEx.GetSceneState().GetTimeout(sceneEx.Scene)),
		DisbandGen: proto.Int(sceneEx.GetDisbandGen()),
		GameFreeId: proto.Int32(sceneEx.gamefreeId),
		FrozenTick: proto.Int32(sceneEx.frozenTick),
	}
	for _, pp := range sceneEx.players {
		pd := &fishing_proto.FishingPlayerData{
			SnId:        proto.Int32(pp.SnId),
			Name:        proto.String(pp.Name),
			Head:        proto.Int32(pp.Head),
			Sex:         proto.Int32(pp.Sex),
			Coin:        proto.Int64(pp.CoinCache),
			Pos:         proto.Int(pp.GetPos()),
			Flag:        proto.Int(pp.GetFlag()),
			City:        proto.String(pp.GetCity()),
			Params:      proto.String(pp.Params),
			HeadOutLine: proto.Int32(pp.HeadOutLine),
			VIP:         proto.Int32(7),
			SelVip:      proto.Int32(pp.SelVip),
			Power:       proto.Int32(pp.power),
			AgentParam:  proto.Int32(pp.AgentParam),
			TargetSel:   proto.Int32(pp.SelTarget),
			AutoFishing: proto.Int32(pp.AutoFishing),
			FireRate:    proto.Int32(pp.FireRate),
			IsRobot:     proto.Bool(pp.IsRob),
			NiceId:      proto.Int32(pp.NiceId),
		}
		for _, v := range pp.RobotSnIds {
			pd.RobotSnIds = append(pd.RobotSnIds, v)
		}
		pack.Players = append(pack.Players, pd)
	}
	proto.SetDefaults(pack)
	p.SendToClient(int(fishing_proto.FIPacketID_FISHING_SC_ROOMINFO), pack)
}

type SceneBaseStateFishing struct {
	base.BaseScenePolicy
}

func (this *SceneBaseStateFishing) GetTimeout(s *base.Scene) int {
	if sceneEx, ok := s.GetExtraData().(*FishingSceneData); ok {
		return int(time.Now().Sub(sceneEx.GetStateStartTime()) / time.Second)
	}
	return 0
}
func (this *SceneBaseStateFishing) CanChangeTo(s base.SceneState) bool {
	return true
}
func (this *SceneBaseStateFishing) CanChangeCoinScene(s *base.Scene, p *base.Player) bool {
	if _, ok := p.GetExtraData().(*FishingPlayerData); ok {
		return true
	}
	return true
}
func (this *SceneBaseStateFishing) OnEnter(s *base.Scene) {
	if sceneEx, ok := s.GetExtraData().(*FishingSceneData); ok {
		sceneEx.SetStateStartTime(time.Now())
	}
}
func (this *SceneBaseStateFishing) OnLeave(s *base.Scene) {}
func (this *SceneBaseStateFishing) OnTick(s *base.Scene) {
	if time.Now().Sub(s.GetGameStartTime()) > time.Second*3 {
		if sceneEx, ok := s.GetExtraData().(*FishingSceneData); ok {
			for _, p := range sceneEx.players {
				if p.IsOnLine() {
					p.leavetime = 0
					continue
				}
				p.leavetime++
				if p.leavetime < 60/3*3 {
					continue
				}
				sceneEx.PlayerLeave(p.Player, common.PlayerLeaveReason_LongTimeNoOp, true)
			}
		}
		s.SetGameStartTime(time.Now())
	}
}
func (this *SceneBaseStateFishing) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if playerEx, ok := p.GetExtraData().(*FishingPlayerData); ok {
		if sceneEx, ok := s.GetExtraData().(*FishingSceneData); ok {
			// start  检测当前房间是否处于基本的关闭状态
			if s.CheckNeedDestroy() {
				if sceneEx.hDestroy == timer.InvalidTimerHandle {
					if hNext, ok := common.DelayInvake(func() {
						sceneEx.hDestroy = timer.InvalidTimerHandle
						sceneEx.SceneDestroy(true)
					}, nil, time.Second*10, 1); ok {
						sceneEx.hDestroy = hNext
					}
				}
				//
				if opcode == FishingPlayerOpFire || opcode == FishingPlayerOpRobotFire { //房间开始关闭，拦截掉发炮数据
					return true
				}
			}
			// end
			switch opcode {
			case FishingPlayerOpFire:
				// 真人玩家的开火倍数
				if len(params) >= 4 {
					opRetCode := sceneEx.PushBullet(s, playerEx.SnId, int32(params[0]), int32(params[1]),
						int32(params[2]), int32(params[3]))
					// start  异常处理的时候 发统一的 处理逻辑
					if opRetCode == fishing_proto.OpResultCode_OPRC_CoinNotEnough {
						this.OnPlayerSToCOp(s, p, opcode, opRetCode, params)
					}
					// end
				}
			case FishingPlayerOpRobotFire:
				if len(params) >= 5 {
					robotPlayer := sceneEx.GetPlayer(int32(params[0]))
					if robotPlayer != nil && robotPlayer.IsRob == true && p.IsRob == false {
						opRetCode := sceneEx.PushBullet(s, robotPlayer.SnId, int32(params[1]), int32(params[2]),
							int32(params[3]), int32(params[4]))
						if opRetCode == fishing_proto.OpResultCode_OPRC_CoinNotEnough {
							//sceneEx.PlayerLeave(robotPlayer, common.PlayerLeaveReason_Bekickout, true)
							if !robotPlayer.IsMarkFlag(base.PlayerState_GameBreak) {
								robotPlayer.MarkFlag(base.PlayerState_GameBreak)
								sceneEx.SCRobotBehavior(p.SnId, robotPlayer.SnId, FishingRobotBehaviorCode_StopFire)
								this.OnPlayerSToCOp(s, robotPlayer, opcode, opRetCode, params)
							}
						}
						if robotPlayerEx, ok := robotPlayer.GetExtraData().(*FishingPlayerData); ok {
							fishlogger.Tracef("playerId  snid %v  robotPlayer sid %v snid %v  coin %v  name %v leaveCoin %v ", p.SnId, robotPlayer.GetSid(), robotPlayer.SnId, robotPlayerEx.CoinCache, robotPlayer.Name, robotPlayer.ExpectLeaveCoin)
							if robotPlayerEx.CoinCheck(robotPlayerEx.power) == false {
								//sceneEx.PlayerLeave(robotPlayerEx.Player, common.PlayerLeaveReason_Bekickout, true)
								fishlogger.Tracef("robotId  sid %v snid %v CoinCheck ", robotPlayer.GetSid(), robotPlayerEx.SnId)
								if !robotPlayer.IsMarkFlag(base.PlayerState_GameBreak) {
									robotPlayer.MarkFlag(base.PlayerState_GameBreak)
									sceneEx.SCRobotBehavior(p.SnId, robotPlayer.SnId, FishingRobotBehaviorCode_StopFire)
									this.OnPlayerSToCOp(s, robotPlayer, opcode, opRetCode, params)
								}
							}
							if !s.CoinInLimit(robotPlayerEx.CoinCache) {
								//sceneEx.PlayerLeave(robotPlayerEx.Player, common.PlayerLeaveReason_Bekickout, true)
								fishlogger.Tracef("robotId  sid %v snid %v CoinInLimit ", robotPlayer.GetSid(), robotPlayerEx.SnId)
								if !robotPlayer.IsMarkFlag(base.PlayerState_GameBreak) {
									robotPlayer.MarkFlag(base.PlayerState_GameBreak)
									sceneEx.SCRobotBehavior(p.SnId, robotPlayer.SnId, FishingRobotBehaviorCode_StopFire)
									this.OnPlayerSToCOp(s, robotPlayer, opcode, opRetCode, params)
								}
							}
							if s.CoinOverMaxLimit(robotPlayerEx.CoinCache, robotPlayer) {
								//sceneEx.PlayerLeave(robotPlayerEx.Player, common.PlayerLeaveReason_Normal, true)
								fishlogger.Tracef("robotId  sid %v snid %v CoinOverMaxLimit ", robotPlayer.GetSid(), robotPlayerEx.SnId)
								if !robotPlayer.IsMarkFlag(base.PlayerState_GameBreak) {
									robotPlayer.MarkFlag(base.PlayerState_GameBreak)
									sceneEx.SCRobotBehavior(p.SnId, robotPlayer.SnId, FishingRobotBehaviorCode_StopFire)
									this.OnPlayerSToCOp(s, robotPlayer, opcode, opRetCode, params)
								}
							}
						}
					}
				}
			case FishingPlayerOpHitFish:
				if len(params) >= 2 {
					fishId := []int32{int32(params[1])}
					extfishis := []int32{}
					for i := 2; i < len(params); i++ {
						extfishis = append(extfishis, int32(params[i]))
					}
					sceneEx.PushBattle(playerEx, int32(params[0]), fishId, extfishis)
				}
			case FishingPlayerOpRobotHitFish:
				if len(params) >= 3 {
					robotPlayer := sceneEx.GetPlayer(int32(params[0]))
					if robotPlayer != nil && robotPlayer.IsRob == true && p.IsRob == false {
						fishId := []int32{int32(params[2])}
						extfishis := []int32{}
						for i := 3; i < len(params); i++ {
							extfishis = append(extfishis, int32(params[i]))
						}
						if robotPlayerEx, ok := robotPlayer.GetExtraData().(*FishingPlayerData); ok {
							sceneEx.PushBattle(robotPlayerEx, int32(params[1]), fishId, extfishis)
						}
					}
				}
			case FishingPlayerOpSetPower:
				if len(params) >= 1 {
					fishlogger.Tracef("%v set power to %v.", playerEx.SnId, params[0])
					powers := sceneEx.GetDBGameFree().GetOtherIntParams() // power的配置信息
					// 当前如果是免费炮时期，后端直接屏蔽协议
					if playerEx.powerType == FreePowerType {
						fishlogger.Tracef("player %v 正处于免费炮阶段.", playerEx.SnId)
						return true
					}
					if common.InSliceInt32(powers, int32(params[0])) {
						playerEx.power = int32(params[0])
					} else {
						fishlogger.Errorf("Can't not switch power to %v.", params[0])
						fishlogger.Errorf("%v scene power list:%v.", sceneEx.gamefreeId, powers)
					}
					pack := &fishing_proto.SCFirePower{
						Snid:  proto.Int32(playerEx.SnId),
						Power: proto.Int32(playerEx.power),
					}

					if playerEx.IsRobot() && len(params) >= 2 {
						pack.TargetPower = proto.Int32(int32(params[1]))
						pack.RobitFire = proto.Bool(int32(params[1]) == playerEx.power)
					}
					sceneEx.BroadCastMessage(int(fishing_proto.FIPacketID_FISHING_SC_FIREPOWER), pack, 0)
					if playerEx.IsRobot() {
						fishlogger.Tracef("机器人%v设置倍率：当前倍率%v 目标倍率%v 能否发泡%v:", playerEx.SnId, playerEx.power, params[1], pack.RobitFire)
					}
					fishlogger.Trace("SCFirePower:", pack)
					if playerEx.IsRob {
						if playerEx.CoinCheck(playerEx.power) == false {
							sceneEx.PlayerLeave(playerEx.Player, common.PlayerLeaveReason_Bekickout, true)
						}
					}
				}
			case FishingRobotOpSetPower:
				if len(params) >= 2 {
					robotPlayer := sceneEx.GetPlayer(int32(params[0]))
					if robotPlayer != nil && robotPlayer.IsRob == true && p.IsRob == false {
						if robotPlayerEx, ok := robotPlayer.GetExtraData().(*FishingPlayerData); ok {
							fishlogger.Tracef("RobotID %v set power to %v.", params[0], params[1])
							powers := sceneEx.GetDBGameFree().GetOtherIntParams() // power的配置信息
							if common.InSliceInt32(powers, int32(params[1])) {
								robotPlayerEx.power = int32(params[1])
							} else {
								fishlogger.Errorf("Can't not switch power to %v.", params[0])
								fishlogger.Errorf("%v scene power list:%v.", sceneEx.gamefreeId, powers)
							}
							pack := &fishing_proto.SCFirePower{
								Snid:  proto.Int32(robotPlayerEx.SnId),
								Power: proto.Int32(robotPlayerEx.power),
							}
							sceneEx.BroadCastMessage(int(fishing_proto.FIPacketID_FISHING_SC_FIREPOWER), pack, 0)
						}
					}
				}

			case FishingPlayerOpSelVip:
				if len(params) > 0 {
					//if p.VIP >= int32(params[0]) {
					playerEx.SelVip = int32(params[0])
					pack := &fishing_proto.SCSelVip{
						Snid: proto.Int32(playerEx.SnId),
						Vip:  proto.Int32(playerEx.SelVip),
					}
					sceneEx.BroadCastMessage(int(fishing_proto.FIPacketID_FISHING_SC_SELVIP), pack, 0)
					//}
				}
			case FishingPlayerOpLeave:
				if playerEx.AgentParam != 0 {
					robot := base.PlayerMgrSington.GetPlayerBySnId(playerEx.AgentParam)
					if robot != nil {
						sceneEx.PlayerLeave(robot, common.PlayerLeaveReason_Bekickout, true)
					}
				}
			case FishingPlayerOpAuto:
				{
					if len(params) > 0 {
						playerEx.AutoFishing = int32(params[0])
						this.OnPlayerSToCOp(s, p, opcode, fishing_proto.OpResultCode_OPRC_Sucess, params)
					}
				}
			case FishingPlayerOpSelTarget:
				{
					if len(params) > 0 {
						playerEx.SelTarget = int32(params[0])
						pack := &fishing_proto.SCFishingOp{
							OpCode:    proto.Int(opcode),
							Params:    params,
							OpRetCode: fishing_proto.OpResultCode_OPRC_Sucess,
							SnId:      proto.Int32(playerEx.SnId),
						}
						proto.SetDefaults(pack)
						sceneEx.BroadCastMessage(int(fishing_proto.FIPacketID_FISHING_SC_OP), pack, 0)
					}
				}
			case FishingPlayerOpFireRate:
				{
					if len(params) > 0 {
						playerEx.FireRate = int32(params[0])
						pack := &fishing_proto.SCFishingOp{
							OpCode:    proto.Int(opcode),
							Params:    params,
							OpRetCode: fishing_proto.OpResultCode_OPRC_Sucess,
							SnId:      proto.Int32(playerEx.SnId),
						}
						proto.SetDefaults(pack)
						sceneEx.BroadCastMessage(int(fishing_proto.FIPacketID_FISHING_SC_OP), pack, 0)
					}
				}
			case FishingPlayerHangup:
				{
					//客户端辅助,原路返回
					this.OnPlayerSToCOp(s, p, opcode, fishing_proto.OpResultCode_OPRC_Sucess, params)
				}
			case FishingRobotWantLeave:
				{
					//机器人想要离开
					if len(params) >= 1 {
						robotPlayer := sceneEx.GetPlayer(int32(params[0]))
						if robotPlayer != nil && robotPlayer.IsRob == true && p.IsRob == false {
							this.OnPlayerSToCOp(s, robotPlayer, opcode, fishing_proto.OpResultCode_OPRC_Sucess, params)
						}
					}
				}
			}
		}
	}
	return false
}
func SyncFishingTarget(sceneEx *FishingSceneData, playerEx *FishingPlayerData) {
	for _, value := range sceneEx.players {
		if value.SnId != playerEx.SnId {
			continue
		}
		if value.TargetFish == 0 || value.SelTarget == 0 {
			continue
		}
		pack := &fishing_proto.SCFishTarget{
			FishId: proto.Int32(value.TargetFish),
			SnId:   proto.Int32(value.SnId),
		}
		playerEx.SendToClient(int(fishing_proto.FIPacketID_FISHING_SC_FISHTARGET), pack)
	}
}

/*
同步当前玩家得炮台状态
*/
func SyncFishingPowerState(sceneEx *FishingSceneData, playerEx *FishingPlayerData) {
	playerPowerState := []*fishing_proto.PlayerPowerState{}
	for _, value := range sceneEx.players {
		pps := &fishing_proto.PlayerPowerState{
			Snid:  proto.Int32(value.SnId),
			State: proto.Int32(value.powerType),
			Num:   proto.Int32(value.FreePowerNum),
		}
		playerPowerState = append(playerPowerState, pps)
	}
	pack := &fishing_proto.SCPowerState{
		PowerState: playerPowerState,
	}
	//sceneEx.BroadCastMessage(int(fishing_proto.FIPacketID_FISHING_SC_POWERSTATE), pack,0)
	playerEx.SendToClient(int(fishing_proto.FIPacketID_FISHING_SC_POWERSTATE), pack)
}

func (this *SceneBaseStateFishing) OnPlayerEvent(s *base.Scene, p *base.Player, evtcode int, params []int64) {
	this.BaseScenePolicy.OnPlayerEvent(s, p, evtcode, params)
	if sceneEx, ok := s.GetExtraData().(*FishingSceneData); ok {
		if playerEx, ok := p.GetExtraData().(*FishingPlayerData); ok {
			switch evtcode {
			case base.PlayerEventRehold:
				fishlogger.Tracef("用户 %v PlayerEventRehold", playerEx.SnId)
			case base.PlayerEventEnter:
				fallthrough
			case base.PlayerEventReturn:
				FishingSendRoomInfo(p, sceneEx)
				pack := &fishing_proto.SCfishingEnter{
					Data: &fishing_proto.FishingPlayerData{
						SnId:        proto.Int32(playerEx.SnId),
						Name:        proto.String(playerEx.Name),
						Head:        proto.Int32(playerEx.Head),
						Sex:         proto.Int32(playerEx.Sex),
						Coin:        proto.Int64(playerEx.CoinCache),
						Pos:         proto.Int(playerEx.GetPos()),
						Flag:        proto.Int(playerEx.GetFlag()),
						City:        proto.String(playerEx.GetCity()),
						Params:      proto.String(playerEx.Params),
						HeadOutLine: proto.Int32(playerEx.HeadOutLine),
						VIP:         proto.Int32(playerEx.VIP),
						SelVip:      proto.Int32(playerEx.SelVip),
						Power:       proto.Int32(playerEx.power),
						AgentParam:  proto.Int32(playerEx.AgentParam),
						TargetSel:   proto.Int32(playerEx.SelTarget),
						AutoFishing: proto.Int32(playerEx.AutoFishing),
						FireRate:    proto.Int32(playerEx.FireRate),
						IsRobot:     proto.Bool(playerEx.IsRob),
						NiceId:      proto.Int32(playerEx.NiceId),
					},
				}
				for _, v := range playerEx.RobotSnIds {
					pack.Data.RobotSnIds = append(pack.Data.RobotSnIds, v)
				}
				fishlogger.Tracef("用户 %v 本身的VIP %v  selVip %v  playerEx.RobotSnIds %v", playerEx.SnId, playerEx.VIP, playerEx.SelVip, playerEx.RobotSnIds)
				sceneEx.BroadCastMessage(int(fishing_proto.FIPacketID_FISHING_SC_ENTER), pack, int64(playerEx.GetSid()))
				SyncFishingTarget(sceneEx, playerEx)
				sceneEx.SyncFish(p)
			case base.PlayerEventLeave:
				pack := &fishing_proto.SCfishingLeave{
					SnId: proto.Int32(playerEx.SnId),
				}
				sceneEx.BroadCastMessage(int(fishing_proto.FIPacketID_FISHING_SC_LEAVE), pack, int64(playerEx.GetSid()))
			case base.PlayerEventRecharge, base.PlayerEventAddCoin:
				playerEx.CoinCache += params[0]
			}
		}
	}
}

// Op协议包的统一发送
func (this *SceneBaseStateFishing) OnPlayerSToCOp(s *base.Scene, p *base.Player, opcode int, opRetCode fishing_proto.OpResultCode, params []int64) {
	pack := &fishing_proto.SCFishingOp{
		OpCode:    proto.Int(opcode),
		Params:    params,
		OpRetCode: opRetCode,
		SnId:      proto.Int32(p.SnId),
	}
	proto.SetDefaults(pack)
	p.SendToClient(int(fishing_proto.FIPacketID_FISHING_SC_OP), pack)
}

type SceneStateFishingStart struct {
	SceneBaseStateFishing
}

func (this *SceneStateFishingStart) GetState() int { return FishingSceneStateStart }
func (this *SceneStateFishingStart) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if this.SceneBaseStateFishing.OnPlayerOp(s, p, opcode, params) {
		return true
	}
	return true
}
func (this *SceneStateFishingStart) OnEnter(s *base.Scene) {
	this.SceneBaseStateFishing.OnEnter(s)
	if sceneEx, ok := s.GetExtraData().(*FishingSceneData); ok {
		fishlogger.Tracef("(this *Scene) [%v] 场景状态进入 %v", s.GetSceneId(), len(sceneEx.players))
		pack := &fishing_proto.SCFishingRoomState{
			State: proto.Int(this.GetState()),
		}
		proto.SetDefaults(pack)
		sceneEx.BroadCastMessage(int(fishing_proto.FIPacketID_FISHING_SC_ROOMSTATE), pack, 0)
		sceneEx.lastTick = 0
		sceneEx.remainder = 0
	}
}
func (this *SceneStateFishingStart) OnTick(s *base.Scene) {
	this.SceneBaseStateFishing.OnTick(s)
	if sceneEx, ok := s.GetExtraData().(*FishingSceneData); ok {
		if sceneEx.lastTick == 0 {
			sceneEx.lastTick = time.Now().UnixNano()
		} else {
			diff := time.Now().UnixNano() - sceneEx.lastTick
			diff = diff / time.Millisecond.Nanoseconds()
			if diff > 100 {
			}
			sceneEx.lastTick = time.Now().UnixNano()
			diff += sceneEx.remainder
			for i := int64(0); i < diff/100; i++ {
				sceneEx.fishFactory()
			}
			sceneEx.remainder = diff % 100
		}
		sceneEx.fishBattle()
		sceneEx.OnTick()
		if sceneEx.FlushFishOver() {
			if sceneEx.Policy_Mode == Policy_Mode_Normal {
				sceneEx.ChangeSceneState(FishingSceneStateClear)
			} else {
				sceneEx.ChangeFlushFish()
			}
		}
	}
}

func (this *SceneStateFishingStart) OnLeave(s *base.Scene) {
	this.SceneBaseStateFishing.OnLeave(s)
	if sceneEx, ok := s.GetExtraData().(*FishingSceneData); ok {
		for _, p := range sceneEx.players {
			if p != nil && p.IsGameing() && !p.IsRob {
				p.SaveDetailedLog(s)
			}
		}
	}
}

type SceneStateFishingClear struct {
	SceneBaseStateFishing
}

func (this *SceneStateFishingClear) GetState() int                      { return FishingSceneStateClear }
func (this *SceneStateFishingClear) CanChangeTo(s base.SceneState) bool { return true }
func (this *SceneStateFishingClear) OnEnter(s *base.Scene) {
	this.SceneBaseStateFishing.OnEnter(s)
	if sceneEx, ok := s.GetExtraData().(*FishingSceneData); ok {
		fishlogger.Tracef("(this *Scene) [%v] 场景状态进入 %v", s.GetSceneId(), len(sceneEx.players))
		sceneEx.Clean()
		sceneEx.ChangeFlushFish()
		pack := &fishing_proto.SCFishingRoomState{
			State:  proto.Int(this.GetState()),
			Params: []int32{int32(common.RandInt(5) + 1)},
		}
		proto.SetDefaults(pack)
		sceneEx.BroadCastMessage(int(fishing_proto.FIPacketID_FISHING_SC_ROOMSTATE), pack, 0)
	}
}
func (this *SceneStateFishingClear) OnTick(s *base.Scene) {
	this.SceneBaseStateFishing.OnTick(s)
	if sceneEx, ok := s.GetExtraData().(*FishingSceneData); ok {
		if time.Now().Sub(sceneEx.GetStateStartTime()) > FishingSceneAniTimeout {
			sceneEx.OnTick()
			sceneEx.ChangeSceneState(FishingSceneStateStart)
		}
	}
}
func (this *SceneStateFishingClear) OnLeave(s *base.Scene) {
	this.SceneBaseStateFishing.OnLeave(s)
	if sceneEx, ok := s.GetExtraData().(*FishingSceneData); ok {
		for _, p := range sceneEx.players {
			leave := false
			var reason int
			if p.IsRob {
				if s.CoinOverMaxLimit(p.CoinCache, p.Player) {
					leave = true
					reason = common.PlayerLeaveReason_Normal
				}
			} else {
				if !p.IsOnLine() {
					leave = true
					reason = common.PlayerLeaveReason_DropLine
				} else if !s.CoinInLimit(p.CoinCache) {
					leave = true
					reason = common.PlayerLeaveReason_Bekickout
				}
			}
			todayGamefreeIDSceneData, _ := p.GetDaliyGameData(int(sceneEx.GetDBGameFree().GetId()))
			if !p.IsRob &&
				todayGamefreeIDSceneData != nil &&
				sceneEx.GetDBGameFree().GetPlayNumLimit() != 0 &&
				todayGamefreeIDSceneData.GameTimes >= int64(sceneEx.GetDBGameFree().GetPlayNumLimit()) {
				leave = true
				reason = common.PlayerLeaveReason_GameTimes
			}
			if leave {
				s.PlayerLeave(p.Player, reason, true)
			}
		}
	}
}
func (this *ScenePolicyFishing) RegisteSceneState(state base.SceneState) {
	if state == nil {
		return
	}
	stateid := state.GetState()
	if stateid < 0 || stateid >= FishingSceneStateMax {
		return
	}
	this.states[stateid] = state
}
func (this *ScenePolicyFishing) GetSceneState(s *base.Scene, stateid int) base.SceneState {
	if stateid >= 0 && stateid < FishingSceneStateMax {
		return ScenePolicyFishingSington.states[stateid]
	}
	return nil
}

//	func (this *ScenePolicyFishing) GetGameSubState(s *base.Scene, stateid int) base.SceneGamingSubState {
//		return nil
//	}
func init() {
	ScenePolicyFishingSington.RegisteSceneState(&SceneStateFishingStart{})
	ScenePolicyFishingSington.RegisteSceneState(&SceneStateFishingClear{})
	core.RegisteHook(core.HOOK_BEFORE_START, func() error {
		base.RegisteScenePolicy(common.GameId_HFishing, fishing.RoomMode_HuanLe, ScenePolicyFishingSington)
		return nil
	})
}
