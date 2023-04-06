package fishing

import (
	"games.yol.com/win88/gamesrv/base"
	"math"
	"time"

	"games.yol.com/win88/common"
	"games.yol.com/win88/gamerule/fishing"
	"games.yol.com/win88/proto"
	fishing_proto "games.yol.com/win88/protocol/fishing"
	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/timer"
)

var ScenePolicyHPFishingSington = &ScenePolicyHPFishing{}

type ScenePolicyHPFishing struct {
	base.BaseScenePolicy
	states [FishingSceneStateMax]base.SceneState
}

func (this *ScenePolicyHPFishing) CreateSceneExData(s *base.Scene) interface{} {
	sceneEx := NewHPFishingSceneData(s)
	if sceneEx != nil {
		if sceneEx.init() {
			s.SetExtraData(sceneEx)
		}
	}
	return sceneEx
}
func (this *ScenePolicyHPFishing) CreatePlayerExData(s *base.Scene, p *base.Player) interface{} {
	playerEx := &FishingPlayerData{Player: p}
	if playerEx != nil {
		p.SetExtraData(playerEx)
	}
	return playerEx
}
func (this *ScenePolicyHPFishing) OnStart(s *base.Scene) {
	fishlogger.Trace("(this *ScenePolicyHPFishing) OnStart, sceneId=", s.GetSceneId())
	sceneEx := NewHPFishingSceneData(s)
	if sceneEx != nil {
		if sceneEx.init() {
			s.SetExtraData(sceneEx)
			s.ChangeSceneState(FishingSceneStateStart)
		}
	}
}
func (this *ScenePolicyHPFishing) OnTick(s *base.Scene) {
	if s == nil {
		return
	}
	if s.GetSceneState() != nil {
		s.GetSceneState().OnTick(s)
	}
}
func (this *ScenePolicyHPFishing) GetPlayerNum(s *base.Scene) int32 {
	if s == nil {
		return 0
	}
	if sceneEx, ok := s.GetExtraData().(*HPFishingSceneData); ok {
		return int32(len(sceneEx.players))
	}
	return 0
}
func (this *ScenePolicyHPFishing) OnPlayerEnter(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	fishlogger.Trace("(this *ScenePolicyHPFishing) OnPlayerEnter, sceneId=", s.GetSceneId(), " player=", p.SnId)
	if sceneEx, ok := s.GetExtraData().(*HPFishingSceneData); ok {
		sceneEx.SetGameNowTime(time.Now())
		playerEx := &FishingPlayerData{Player: p}
		playerEx.init(s)
		if playerEx.Prana > 0 {
			upperLimit := sceneEx.PranaUpperLimit()
			if int32(playerEx.Prana) > upperLimit {
				playerEx.PranaPercent = 100
			} else {
				playerEx.PranaPercent = int32(playerEx.Prana) * 100 / upperLimit
			}
		}
		playerEx.taxCoin = 0
		playerEx.sTaxCoin = 0
		playerEx.winCoin = 0
		playerEx.enterTime = time.Now()
		sceneEx.EnterPlayer(playerEx)
		s.FirePlayerEvent(p, base.PlayerEventEnter, nil)
	}
}

func (this *ScenePolicyHPFishing) OnPlayerLeave(s *base.Scene, p *base.Player, reason int) {
	if s == nil || p == nil {
		return
	}
	fishlogger.Trace("(this *ScenePolicyHPFishing) OnPlayerLeave, sceneId=", s.GetSceneId(), " player=", p.Name)
	if sceneEx, ok := s.GetExtraData().(*HPFishingSceneData); ok {
		if this.CanChangeCoinScene(s, p) {
			if playerEx, ok := p.GetExtraData().(*FishingPlayerData); ok {
				playerEx.SaveDetailedLog(s)
				playerEx.RetBulletCoin(s)
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
func (this *ScenePolicyHPFishing) OnPlayerDropLine(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	fishlogger.Trace("(this *ScenePolicyHPFishing) OnPlayerDropLine, sceneId=", s.GetSceneId(), " player=", p.Name)
	s.FirePlayerEvent(p, base.PlayerEventDropLine, nil)
	if _, ok := s.GetExtraData().(*HPFishingSceneData); ok {
		if _, ok := p.GetExtraData().(*FishingPlayerData); ok {
		}
	}
}
func (this *ScenePolicyHPFishing) OnPlayerRehold(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	fishlogger.Trace("(this *ScenePolicyHPFishing) OnPlayerRehold, sceneId=", s.GetSceneId(), " player=", p.SnId)
	if _, ok := s.GetExtraData().(*HPFishingSceneData); ok {
		if _, ok := p.GetExtraData().(*FishingPlayerData); ok {
			s.FirePlayerEvent(p, base.PlayerEventRehold, nil)
		}
	}
}
func (this *ScenePolicyHPFishing) OnPlayerReturn(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	fishlogger.Trace("(this *ScenePolicyHPFishing) OnPlayerReturn, sceneId=", s.GetSceneId(), " player=", p.SnId)
	if _, ok := s.GetExtraData().(*HPFishingSceneData); ok {
		if _, ok := p.GetExtraData().(*FishingPlayerData); ok {
			s.FirePlayerEvent(p, base.PlayerEventReturn, nil)
		}
	}
}
func (this *ScenePolicyHPFishing) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if s == nil || p == nil {
		return false
	}
	if s.GetSceneState() != nil {
		return s.GetSceneState().OnPlayerOp(s, p, opcode, params)
	}
	return true
}
func (this *ScenePolicyHPFishing) OnPlayerEvent(s *base.Scene, p *base.Player, evtcode int, params []int64) {
	if s == nil || p == nil {
		return
	}
	if s.GetSceneState() != nil {
		s.GetSceneState().OnPlayerEvent(s, p, evtcode, params)
	}
}
func (this *ScenePolicyHPFishing) IsCompleted(s *base.Scene) bool     { return false }
func (this *ScenePolicyHPFishing) IsCanForceStart(s *base.Scene) bool { return true }
func (this *ScenePolicyHPFishing) CanChangeCoinScene(s *base.Scene, p *base.Player) bool {
	if s == nil || p == nil {
		return true
	}
	if s.GetSceneState() != nil {
		return s.GetSceneState().CanChangeCoinScene(s, p)
	}
	return true
}
func HPFishingSendRoomInfo(p *base.Player, sceneEx *HPFishingSceneData) {
	pack := &fishing_proto.SCFishingRoomInfo{
		RoomId:     proto.Int(sceneEx.GetSceneId()),
		Creator:    proto.Int32(sceneEx.GetCreator()),
		GameId:     proto.Int(sceneEx.GameId),
		RoomMode:   proto.Int(sceneEx.GameMode),
		AgentId:    proto.Int32(sceneEx.GetAgentor()),
		SceneType:  proto.Int32(sceneEx.DbGameFree.GetSceneType()),
		Params:     sceneEx.GetParams(),
		NumOfGames: proto.Int(sceneEx.GetNumOfGames()),
		State:      proto.Int(sceneEx.GetSceneState().GetState()),
		TimeOut:    proto.Int(sceneEx.GetSceneState().GetTimeout(sceneEx.Scene)),
		DisbandGen: proto.Int(sceneEx.GetDisbandGen()),
		GameFreeId: proto.Int32(sceneEx.GetGameFreeId()),
		FrozenTick: proto.Int32(sceneEx.frozenTick),
	}
	if _, exist := FishJackpotCoinMgr.Jackpot[sceneEx.Platform]; exist {
		//pack.JackpotPool = proto.Int64(Jackpot.GetTotalBig())
		pack.JackpotPool = proto.Int64(sceneEx.GetJackpot(0))
	}
	for _, pp := range sceneEx.players {
		pd := &fishing_proto.FishingPlayerData{
			SnId:         proto.Int32(pp.SnId),
			Name:         proto.String(pp.Name),
			Head:         proto.Int32(pp.Head),
			Sex:          proto.Int32(pp.Sex),
			Coin:         proto.Int64(pp.CoinCache),
			Pos:          proto.Int(pp.GetPos()),
			Flag:         proto.Int(pp.GetFlag()),
			City:         proto.String(pp.GetCity()),
			Params:       proto.String(pp.Params),
			HeadOutLine:  proto.Int32(pp.HeadOutLine),
			VIP:          proto.Int32(pp.VIP),
			SelVip:       proto.Int32(pp.SelVip),
			Power:        proto.Int32(pp.power),
			AgentParam:   proto.Int32(pp.AgentParam),
			TargetSel:    proto.Int32(pp.SelTarget),
			AutoFishing:  proto.Int32(pp.AutoFishing),
			FireRate:     proto.Int32(pp.FireRate),
			IsRobot:      proto.Bool(pp.IsRob),
			PranaPercent: proto.Int32(pp.PranaPercent),
			NiceId:       proto.Int32(pp.NiceId),
		}
		for _, v := range pp.RobotSnIds {
			pd.RobotSnIds = append(pd.RobotSnIds, v)
		}
		pack.Players = append(pack.Players, pd)
	}
	proto.SetDefaults(pack)
	p.SendToClient(int(fishing_proto.FIPacketID_FISHING_SC_ROOMINFO), pack)
}

type SceneBaseStateHPFishing struct {
}

func (this *SceneBaseStateHPFishing) GetTimeout(s *base.Scene) int {
	if sceneEx, ok := s.GetExtraData().(*HPFishingSceneData); ok {
		return int(time.Now().Sub(sceneEx.GetStateStartTime()) / time.Second)
	}
	return 0
}
func (this *SceneBaseStateHPFishing) CanChangeTo(s base.SceneState) bool {
	return true
}
func (this *SceneBaseStateHPFishing) CanChangeCoinScene(s *base.Scene, p *base.Player) bool {
	if _, ok := p.GetExtraData().(*FishingPlayerData); ok {
		return true
	}
	return true
}
func (this *SceneBaseStateHPFishing) OnEnter(s *base.Scene) {
	if sceneEx, ok := s.GetExtraData().(*HPFishingSceneData); ok {
		sceneEx.SetStateStartTime(time.Now())
	}
}
func (this *SceneBaseStateHPFishing) OnLeave(s *base.Scene) {}
func (this *SceneBaseStateHPFishing) OnTick(s *base.Scene) {
	if time.Now().Sub(s.GetGameStartTime()) > time.Second*3 {
		if sceneEx, ok := s.GetExtraData().(*HPFishingSceneData); ok {
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
func (this *SceneBaseStateHPFishing) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if playerEx, ok := p.GetExtraData().(*FishingPlayerData); ok {
		if sceneEx, ok := s.GetExtraData().(*HPFishingSceneData); ok {
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
			switch opcode {
			case FishingPlayerOpFire:
				if len(params) >= 4 {
					//fishlogger.Infof("GameId_TFishing playerEx.SnId:%v OpFire buttleId:%v ", playerEx.SnId, int32(params[2]))
					opRetCode := sceneEx.PushBullet(s, playerEx.SnId, int32(params[0]), int32(params[1]),
						int32(params[2]), int32(params[3]))
					if opRetCode == fishing_proto.OpResultCode_OPRC_CoinNotEnough {
						this.OnPlayerSToCOp(s, p, opcode, opRetCode, params)
					}
				}
			case FishingPlayerOpRobotFire:
				if len(params) >= 5 {
					robotPlayer := sceneEx.GetPlayer(int32(params[0]))
					if robotPlayer != nil && robotPlayer.IsRob == true && p.IsRob == false {
						opRetCode := sceneEx.PushBullet(s, robotPlayer.SnId, int32(params[1]), int32(params[2]),
							int32(params[3]), int32(params[4]))
						if opRetCode == fishing_proto.OpResultCode_OPRC_CoinNotEnough {
							sceneEx.PlayerLeave(robotPlayer, common.PlayerLeaveReason_Bekickout, true)
						}
						if robotPlayerEx, ok := robotPlayer.GetExtraData().(*FishingPlayerData); ok {
							if robotPlayerEx.CoinCheck(robotPlayerEx.power) == false {
								sceneEx.PlayerLeave(robotPlayerEx.Player, common.PlayerLeaveReason_Bekickout, true)
							}
						}
					}
				}
			case FishingPlayerOpHitFish:
				if len(params) >= 3 {
					fishId := []int32{int32(params[2])}
					extfishis := []int32{}
					for i := 3; i < len(params); i++ {
						extfishis = append(extfishis, int32(params[i]))
					}
					//fishlogger.Infof("GameId_TFishing playerEx.SnId:%v OpHitFish buttleId:%v ", playerEx.SnId, int32(params[0]))
					sceneEx.PushBattle(playerEx, int32(params[0]), int32(params[1]), fishId, extfishis)
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
							sceneEx.PushBattle(robotPlayerEx, int32(params[1]), 0, fishId, extfishis)
						}
					}
				}
			case FishingPlayerOpSetPower:
				if len(params) >= 1 {
					fishlogger.Tracef("%v set power to %v.", playerEx.SnId, params[0])
					powers := sceneEx.GetDBGameFree().GetOtherIntParams()
					if common.InSliceInt32(powers, int32(params[0])) {
						playerEx.power = int32(params[0])
					} else {
						fishlogger.Errorf("Can't not switch power to %v.", params[0])
						fishlogger.Errorf("%v scene power list:%v.", sceneEx.GetGameFreeId(), powers)
					}
					pack := &fishing_proto.SCFirePower{
						Snid:  proto.Int32(playerEx.SnId),
						Power: proto.Int32(playerEx.power),
					}
					sceneEx.BroadCastMessage(int(fishing_proto.FIPacketID_FISHING_SC_FIREPOWER), pack, 0)
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
								fishlogger.Errorf("%v scene power list:%v.", sceneEx.GetGameFreeId(), powers)
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
			case FishingRobotOpAuto:
				{
					if len(params) > 1 {
						robotPlayer := sceneEx.GetPlayer(int32(params[0]))
						if robotPlayer != nil {
							if robotPlayerEx, ok := robotPlayer.GetExtraData().(*FishingPlayerData); ok {
								robotPlayerEx.AutoFishing = int32(params[1])
								pack := &fishing_proto.SCFishingOp{
									OpCode:    proto.Int(opcode),
									Params:    params,
									OpRetCode: fishing_proto.OpResultCode_OPRC_Sucess,
									SnId:      proto.Int32(playerEx.SnId),
								}
								sceneEx.BroadCastMessage(int(fishing_proto.FIPacketID_FISHING_SC_OP), pack, 0)
								//this.OnPlayerSToCOp(s, p, opcode, fishing_proto.OpResultCode_OPRC_Sucess, params)
							}
						}
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
			}
		}
	}
	return false
}
func SyncHPFishingTarget(sceneEx *HPFishingSceneData, playerEx *FishingPlayerData) {
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
func (this *SceneBaseStateHPFishing) OnPlayerEvent(s *base.Scene, p *base.Player, evtcode int, params []int64) {
	if sceneEx, ok := s.GetExtraData().(*HPFishingSceneData); ok {
		if playerEx, ok := p.GetExtraData().(*FishingPlayerData); ok {
			switch evtcode {
			case base.PlayerEventRehold:
				fishlogger.Tracef("用户 %v PlayerEventRehold", playerEx.SnId)
			case base.PlayerEventEnter:
				fallthrough
			case base.PlayerEventReturn:
				HPFishingSendRoomInfo(p, sceneEx)
				pack := &fishing_proto.SCfishingEnter{
					Data: &fishing_proto.FishingPlayerData{
						SnId:         proto.Int32(playerEx.SnId),
						Name:         proto.String(playerEx.Name),
						Head:         proto.Int32(playerEx.Head),
						Sex:          proto.Int32(playerEx.Sex),
						Coin:         proto.Int64(playerEx.CoinCache),
						Pos:          proto.Int(playerEx.GetPos()),
						Flag:         proto.Int(playerEx.GetFlag()),
						City:         proto.String(playerEx.GetCity()),
						Params:       proto.String(playerEx.Params),
						HeadOutLine:  proto.Int32(playerEx.HeadOutLine),
						VIP:          proto.Int32(playerEx.VIP),
						SelVip:       proto.Int32(playerEx.SelVip),
						Power:        proto.Int32(playerEx.power),
						AgentParam:   proto.Int32(playerEx.AgentParam),
						TargetSel:    proto.Int32(playerEx.SelTarget),
						AutoFishing:  proto.Int32(playerEx.AutoFishing),
						FireRate:     proto.Int32(playerEx.FireRate),
						IsRobot:      proto.Bool(playerEx.IsRob),
						PranaPercent: proto.Int32(playerEx.PranaPercent),
						NiceId:       proto.Int32(playerEx.NiceId),
					},
				}
				for _, v := range playerEx.RobotSnIds {
					pack.Data.RobotSnIds = append(pack.Data.RobotSnIds, v)
				}
				sceneEx.BroadCastMessage(int(fishing_proto.FIPacketID_FISHING_SC_ENTER), pack, int64(playerEx.GetSid()))
				SyncHPFishingTarget(sceneEx, playerEx)
				sceneEx.SyncFish(p)
			case base.PlayerEventLeave:
				pack := &fishing_proto.SCfishingLeave{
					SnId: proto.Int32(playerEx.SnId),
				}
				sceneEx.BroadCastMessage(int(fishing_proto.FIPacketID_FISHING_SC_LEAVE), pack, int64(playerEx.GetSid()))
			case base.PlayerEventRecharge:
				p.AddCoin(params[0], common.GainWay_Pay, base.SyncFlag_ToClient, "system", p.GetScene().GetSceneName())
				playerEx.CoinCache += params[0]
				p.SetTakeCoin(p.GetTakeCoin() + params[0]) // 为了保持游戏事件统计中的输赢分计算正确
				playerEx.SetMaxCoin()
			}
		}
	}
}
func (this *SceneBaseStateHPFishing) OnPlayerSToCOp(s *base.Scene, p *base.Player, opcode int, opRetCode fishing_proto.OpResultCode, params []int64) {
	pack := &fishing_proto.SCFishingOp{
		OpCode:    proto.Int(opcode),
		Params:    params,
		OpRetCode: opRetCode,
		SnId:      proto.Int32(p.SnId),
	}
	proto.SetDefaults(pack)
	p.SendToClient(int(fishing_proto.FIPacketID_FISHING_SC_OP), pack)
}

type SceneStateHPFishingStart struct {
	SceneBaseStateHPFishing
}

func (this *SceneStateHPFishingStart) GetState() int { return FishingSceneStateStart }
func (this *SceneStateHPFishingStart) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if this.SceneBaseStateHPFishing.OnPlayerOp(s, p, opcode, params) {
		return true
	}
	return true
}
func (this *SceneStateHPFishingStart) OnEnter(s *base.Scene) {
	this.SceneBaseStateHPFishing.OnEnter(s)
	if sceneEx, ok := s.GetExtraData().(*HPFishingSceneData); ok {
		fishlogger.Tracef("(this *base.Scene) [%v] 场景状态进入 %v", s.GetSceneId(), len(sceneEx.players))
		pack := &fishing_proto.SCFishingRoomState{
			State: proto.Int(this.GetState()),
		}
		proto.SetDefaults(pack)
		sceneEx.BroadCastMessage(int(fishing_proto.FIPacketID_FISHING_SC_ROOMSTATE), pack, 0)
		sceneEx.lastTick = 0
		sceneEx.remainder = 0
	}
}
func (this *SceneStateHPFishingStart) OnTick(s *base.Scene) {
	this.SceneBaseStateHPFishing.OnTick(s)
	if sceneEx, ok := s.GetExtraData().(*HPFishingSceneData); ok {
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
			//if sceneEx.Policy_Mode == Policy_Mode_Normal {
			//	sceneEx.ChangeSceneState(FishingSceneStateClear)
			//} else {
			sceneEx.ChangeFlushFish()
			//}
		}
	}
}

func (this *SceneStateHPFishingStart) OnLeave(s *base.Scene) {
	this.SceneBaseStateHPFishing.OnLeave(s)
	if sceneEx, ok := s.GetExtraData().(*HPFishingSceneData); ok {
		for _, p := range sceneEx.players {
			if p != nil && p.IsGameing() && !p.IsRob {
				p.SaveDetailedLog(s)
			}
		}
	}
}

type SceneStateHPFishingClear struct {
	SceneBaseStateHPFishing
}

func (this *SceneStateHPFishingClear) GetState() int                      { return FishingSceneStateClear }
func (this *SceneStateHPFishingClear) CanChangeTo(s base.SceneState) bool { return true }
func (this *SceneStateHPFishingClear) OnEnter(s *base.Scene) {
	this.SceneBaseStateHPFishing.OnEnter(s)
	if sceneEx, ok := s.GetExtraData().(*HPFishingSceneData); ok {
		fishlogger.Tracef("(this *base.Scene) [%v] 场景状态进入 %v", s.GetSceneId(), len(sceneEx.players))
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
func (this *SceneStateHPFishingClear) OnTick(s *base.Scene) {
	this.SceneBaseStateHPFishing.OnTick(s)
	if sceneEx, ok := s.GetExtraData().(*HPFishingSceneData); ok {
		if time.Now().Sub(sceneEx.GetStateStartTime()) > FishingSceneAniTimeout {
			sceneEx.OnTick()
			sceneEx.ChangeSceneState(FishingSceneStateStart)
		}
	}
}
func (this *SceneStateHPFishingClear) OnLeave(s *base.Scene) {
	this.SceneBaseStateHPFishing.OnLeave(s)
	if sceneEx, ok := s.GetExtraData().(*HPFishingSceneData); ok {
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
			if leave {
				s.PlayerLeave(p.Player, reason, true)
			}
		}
	}
}
func (this *ScenePolicyHPFishing) RegisteSceneState(state base.SceneState) {
	if state == nil {
		return
	}
	stateid := state.GetState()
	if stateid < 0 || stateid >= FishingSceneStateMax {
		return
	}
	this.states[stateid] = state
}
func (this *ScenePolicyHPFishing) GetSceneState(s *base.Scene, stateid int) base.SceneState {
	if stateid >= 0 && stateid < FishingSceneStateMax {
		return ScenePolicyHPFishingSington.states[stateid]
	}
	return nil
}

//	func (this *ScenePolicyHPFishing) GetGameSubState(s *base.Scene, stateid int) base.SceneGamingSubState {
//		return nil
//	}
func init() {
	ScenePolicyHPFishingSington.RegisteSceneState(&SceneStateHPFishingStart{})
	ScenePolicyHPFishingSington.RegisteSceneState(&SceneStateHPFishingClear{})
	core.RegisteHook(core.HOOK_BEFORE_START, func() error {
		base.RegisteScenePolicy(common.GameId_TFishing, fishing.RoomMode_TianTian, ScenePolicyHPFishingSington)
		return nil
	})
}
