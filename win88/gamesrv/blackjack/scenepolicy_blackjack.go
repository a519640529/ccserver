package blackjack

import (
	"games.yol.com/win88/common"
	rule "games.yol.com/win88/gamerule/blackjack"
	"games.yol.com/win88/gamesrv/base"
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/blackjack"
	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/timer"
	"math"
	"time"
)

var blackJackSingleton = &ScenePolicyBlackJack{}

type ScenePolicyBlackJack struct {
	base.BaseScenePolicy
	states [rule.StatusMax]base.SceneState
}

func (this *ScenePolicyBlackJack) OnStart(s *base.Scene) {
	logger.Logger.Trace("BlackJack OnStart, sceneId=", s.GetSceneId())
	sceneEx := NewBlackJackSceneData(s)
	if sceneEx.Init() {
		s.SetExtraData(sceneEx)
		s.ChangeSceneState(rule.StatusWait)
	}
}

//强制开始
func (this *ScenePolicyBlackJack) ForceStart(s *base.Scene) {
	if _, ok := s.GetExtraData().(*BlackJackSceneData); ok {
		if s.GetDBGameFree().GetMatchMode() == 1 {
			s.ChangeSceneState(rule.StatusBet)
		} else {
			if s.GetNumOfGames() == 0 && s.GetSceneState().GetState() == rule.StatusReady {
				s.ChangeSceneState(rule.StatusBet)
			}
		}
	}
}
func (this *ScenePolicyBlackJack) OnTick(s *base.Scene) {
	if s == nil {
		return
	}
	if s.GetSceneState() != nil {
		s.GetSceneState().OnTick(s)
	}
}

// 查询空座位号
func BlackJackGetSeat(s *base.Scene) int {
	sceneEx, ok := s.GetExtraData().(*BlackJackSceneData)
	if !ok {
		return 0
	}
	var seats []int
	for _, v := range sceneEx.chairIDs[1:] {
		if v.Player != nil {
			seats = append(seats, v.seat)
		}
	}
	for i := 1; i <= rule.MaxPlayer; i++ {
		var has bool
		for _, v := range seats {
			if i == v {
				has = true
				break
			}
		}
		if !has {
			return i
		}
	}
	return 0
}

// 根据座位号查找玩家数据
func BlackJackGetPlayerData(s *base.Scene, seat int) *BlackJackPlayerData {
	sceneEx, ok := s.GetExtraData().(*BlackJackSceneData)
	if !ok {
		return nil
	}
	for _, v := range sceneEx.chairIDs[1:] {
		if v.seat == seat {
			return v
		}
	}
	return nil
}

func (this *ScenePolicyBlackJack) OnPlayerEnter(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("BlackJack OnPlayerEnter, sceneId=", s.GetSceneId(), " player=", p.SnId)
	sceneEx, ok := s.GetExtraData().(*BlackJackSceneData)
	if !ok {
		return
	}
	if s.Gaming && !s.GetBEnterAfterStart() {
		return
	}
	seat := BlackJackGetSeat(s)
	if seat == 0 {
		return
	}
	pd := BlackJackGetPlayerData(s, seat)
	pd.Player = p
	p.SetPos(seat)
	p.SetExtraData(pd)
	sceneEx.players[p.SnId] = pd
	if s.Gaming {
		p.MarkFlag(base.PlayerState_WaitNext)
		p.SyncFlag()
	}
	if len(s.Players) > 1 {
		BlackJackPlayerEnter(s, p, seat, p.GetSid())
	}
	BlackJackSendRoomInfo(s, p, sceneEx, BlackJackGetPlayerData(s, seat))
	s.FirePlayerEvent(p, base.PlayerEventEnter, nil)
}

func BlackJackPlayerEnter(s *base.Scene, p *base.Player, seatId int, exclude int64) {
	pack := &blackjack.SCBlackJackPlayerEnter{
		Player: &blackjack.BlackJackPlayer{
			//SnId:        proto.Int32(p.SnId),
			SnId:        proto.Int32(p.SnId),
			Name:        proto.String(p.Name),
			Head:        proto.Int32(p.Head),
			Sex:         proto.Int32(p.Sex),
			Coin:        proto.Int64(p.GetCoin()),
			Pos:         proto.Int(p.GetPos()),
			Seat:        proto.Int32(int32(seatId)),
			Flag:        proto.Int(p.GetFlag()),
			Longitude:   proto.Int32(p.Longitude),
			Latitude:    proto.Int32(p.Latitude),
			City:        proto.String(p.GetCity()),
			AgentCode:   proto.String(p.AgentCode),
			HeadOutLine: proto.Int32(p.HeadOutLine),
			VIP:         proto.Int32(p.VIP),
			NiceId:      proto.Int32(p.NiceId),
		},
	}
	proto.SetDefaults(pack)
	s.Broadcast(int(blackjack.BlackJackPacketID_SC_PLAYER_ENTER), pack, exclude)
	logger.Logger.Trace("--> Broadcast SCBlackJackPlayerEnter ", pack.String())
}

func BlackJackSendRoomInfo(s *base.Scene, p *base.Player, sceneEx *BlackJackSceneData, playerEx *BlackJackPlayerData) {
	pack := BlackJackRoomInfo(s, sceneEx, playerEx, p)
	p.SendToClient(int(blackjack.BlackJackPacketID_SC_ROOM_INFO), pack)
	logger.Logger.Trace("--> SCBlackJackRoomInfo ", pack.String())
}

func BlackJackRoomInfo(s *base.Scene, sceneEx *BlackJackSceneData, playerEx *BlackJackPlayerData, p *base.Player) *blackjack.SCBlackJackRoomInfo {
	pack := &blackjack.SCBlackJackRoomInfo{
		RoomId:       proto.Int(s.SceneId),
		Creator:      proto.Int32(s.Creator),
		GameId:       proto.Int(s.GameId),
		RoomMode:     proto.Int(s.SceneMode),
		Params:       s.GetParams(),
		NumOfGames:   proto.Int(s.NumOfGames),
		TotalOfGames: proto.Int(int(s.TotalOfGames)),
		BankerPos:    proto.Int(0),
		State:        proto.Int(s.SceneState.GetState()),
		TimeOut:      proto.Int(s.SceneState.GetTimeout(s)),
		DisbandGen:   proto.Int(s.GetDisbandGen()),
		AgentId:      proto.Int32(s.GetAgentor()),
		ParamsEx:     s.GetParamsEx(),
		SceneType:    proto.Int(s.SceneType),
		BetScope:     sceneEx.BetMinMax(),
		Num:          proto.Int32(int32(len(sceneEx.pokers))),
		IsAudience:   proto.Bool(s.HasAudience(p)),
	}
	// 当前操作的玩家座位号
	if sceneEx.pos <= rule.MaxPlayer {
		pack.Pos = proto.Int(sceneEx.chairIDs[sceneEx.pos].seat)
	}
	// 场景状态
	state := s.SceneState.GetState()
	for _, v := range sceneEx.chairIDs {
		if v.Player == nil && !v.isBanker {
			continue
		}
		var playInfo *blackjack.BlackJackPlayer
		if v.isBanker {
			playInfo = &blackjack.BlackJackPlayer{}
		} else {
			playInfo = &blackjack.BlackJackPlayer{
				SnId:        proto.Int32(v.SnId),
				Name:        proto.String(v.Name),
				Head:        proto.Int32(v.Head),
				Sex:         proto.Int32(v.Sex),
				Coin:        proto.Int64(sceneEx.PlayerCoin(v)),
				Pos:         proto.Int(v.GetPos()),
				Seat:        proto.Int32(int32(v.seat)),
				BetCoin:     proto.Int64(v.BetCoin()),
				BaoCoin:     proto.Int64(v.baoCoin),
				Flag:        proto.Int(v.GetFlag()),
				Longitude:   proto.Int32(v.Longitude),
				Latitude:    proto.Int32(v.Latitude),
				City:        proto.String(v.City),
				AgentCode:   proto.String(v.AgentCode),
				HeadOutLine: proto.Int32(v.HeadOutLine),
				VIP:         proto.Int32(v.VIP),
				NiceId:      proto.Int32(v.NiceId),
			}
		}

		if state != rule.StatusWait && state != rule.StatusReady {
			// 玩家手牌信息
			for k, cs := range v.hands {
				if len(cs.handCards) == 0 {
					continue
				}
				hand := &blackjack.BlackJackCards{
					DCards:  proto.Int32(int32(cs.mulCards.Value())),
					Type:    proto.Int32(cs.tp),
					Point:   cs.point,
					State:   proto.Int32(cs.state),
					Id:      proto.Int32(int32(k)),
					BetCoin: proto.Int64(cs.betCoin),
				}
				for _, n := range cs.handCards {
					hand.Cards = append(hand.Cards, int32(n.Value()))
				}
				playInfo.Cards = append(playInfo.Cards, hand)
			}
		}
		var opFlagParams = []int32{v.opCodeFlag, v.opRightFlag}
		playInfo.OpFlagParams = opFlagParams
		pack.Players = append(pack.Players, playInfo)
	}
	proto.SetDefaults(pack)
	return pack
}

func (this *ScenePolicyBlackJack) OnPlayerLeave(s *base.Scene, p *base.Player, reason int) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("BlackJack OnPlayerLeave, sceneId=", s.SceneId, " player=", p.SnId)
	if sceneEx, ok := s.GetExtraData().(*BlackJackSceneData); ok {
		sceneEx.OnPlayerLeave(p)
		s.FirePlayerEvent(p, base.PlayerEventLeave, nil)
	}
}

func (this *ScenePolicyBlackJack) OnPlayerDropLine(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("BlackJack OnPlayerDropLine, sceneId=", s.SceneId, " player=", p.SnId)
	// 等待的玩家掉线，直接离开游戏
	// 游戏没有开始，直接离开游戏
	if p.IsMarkFlag(base.PlayerState_WaitNext) || !s.Gaming {
		s.PlayerLeave(p, common.PlayerLeaveReason_DropLine, false)
	}
}

//玩家重连
func (this *ScenePolicyBlackJack) OnPlayerRehold(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("BlackJack OnPlayerRehold, sceneId=", s.SceneId, " player=", p.SnId)
	if sceneEx, ok := s.GetExtraData().(*BlackJackSceneData); ok {
		if playerEx, ok := p.GetExtraData().(*BlackJackPlayerData); ok {
			BlackJackSendRoomInfo(s, p, sceneEx, playerEx)
		}
	}
	// 自动准备
	if p.IsGameing() || !s.Gaming {
		if !p.IsReady() {
			p.MarkFlag(base.PlayerState_Ready)
			p.SyncFlag()
		}
		if p.IsMarkFlag(base.PlayerState_WaitNext) {
			p.UnmarkFlag(base.PlayerState_WaitNext)
			p.SyncFlag()
		}
	}
	s.FirePlayerEvent(p, base.PlayerEventRehold, nil)
}

//返回房间
func (this *ScenePolicyBlackJack) OnPlayerReturn(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("BlackJack OnPlayerReturn, sceneId=", s.SceneId, " player=", p.SnId)
	if sceneEx, ok := s.GetExtraData().(*BlackJackSceneData); ok {
		if playerEx, ok := p.GetExtraData().(*BlackJackPlayerData); ok {
			BlackJackSendRoomInfo(s, p, sceneEx, playerEx)
		}
	}
	// 自动准备
	if p.IsGameing() || !s.Gaming {
		if !p.IsReady() {
			p.MarkFlag(base.PlayerState_Ready)
			p.SyncFlag()
		}
		if p.IsMarkFlag(base.PlayerState_WaitNext) {
			p.UnmarkFlag(base.PlayerState_WaitNext)
			p.SyncFlag()
		}
	}
	s.FirePlayerEvent(p, base.PlayerEventReturn, nil)
}

func (this *ScenePolicyBlackJack) OnPlayerOp(s *base.Scene, p *base.Player, code int, params []int64) bool {
	if s == nil || p == nil {
		return false
	}
	if s.SceneState != nil {
		logger.Logger.Trace("--> OnPlayerOp", p.SnId, code, params)
		return s.SceneState.OnPlayerOp(s, p, code, params)
	}
	return false
}

func (this *ScenePolicyBlackJack) OnPlayerEvent(s *base.Scene, p *base.Player, code int, params []int64) {
	if s == nil || p == nil {
		return
	}
	if s.SceneState != nil {
		s.SceneState.OnPlayerEvent(s, p, code, params)
	}
}

func (this *ScenePolicyBlackJack) OnAudienceEnter(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyBlackJack) OnAudienceEnter, sceneId=", s.SceneId, " player=", p.SnId)
	sceneEx, ok := s.GetExtraData().(*BlackJackSceneData)
	if !ok {
		return
	}
	BlackJackSendRoomInfo(s, p, sceneEx, nil)
}

func (this *ScenePolicyBlackJack) OnAudienceLeave(s *base.Scene, p *base.Player, reason int) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("ScenePolicyTenHalf OnAudienceLeave, sceneId=", s.SceneId, " player=", p.SnId)
	//this.OnPlayerLeave(s, p, common.PlayerLeaveReason_Normal)
}

func (this *ScenePolicyBlackJack) OnAudienceSit(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("BlackJack OnAudienceSit, sceneId=", s.SceneId, " player=", p.SnId)
	sceneEx, ok := s.GetExtraData().(*BlackJackSceneData)
	if !ok {
		return
	}
	if s.Gaming && !s.GetBEnterAfterStart() {
		return
	}
	seat := BlackJackGetSeat(s)
	if seat == 0 {
		return
	}
	pd := BlackJackGetPlayerData(s, seat)
	pd.Player = p
	p.SetPos(seat)
	p.SetExtraData(pd)
	sceneEx.players[p.SnId] = pd
	p.UnmarkFlag(base.PlayerState_Audience)
	if s.Gaming {
		p.MarkFlag(base.PlayerState_WaitNext)
		p.SyncFlag()
	}
	BlackJackPlayerEnter(s, p, seat, 0)
	s.FirePlayerEvent(p, base.PlayerEventEnter, nil)
}

func (this *ScenePolicyBlackJack) OnAudienceDropLine(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("ScenePolicyBlackJack OnAudienceDropLine, sceneId=", s.SceneId, " player=", p.SnId)
	s.AudienceLeave(p, common.PlayerLeaveReason_DropLine)
}

func (this *ScenePolicyBlackJack) GetSceneState(s *base.Scene, stateid int) base.SceneState {
	if stateid >= 0 && stateid < rule.StatusMax {
		return this.states[stateid]
	}
	return nil
}

func (this *ScenePolicyBlackJack) CanChangeCoinScene(s *base.Scene, p *base.Player) bool {
	if s == nil || p == nil {
		return false
	}
	if s.SceneState != nil {
		return s.SceneState.CanChangeCoinScene(s, p)
	}
	return false
}

func (this *ScenePolicyBlackJack) CreateSceneExData(s *base.Scene) interface{} {
	if s != nil {
		s.SetExtraData(NewBlackJackSceneData(s))
	}
	return nil
}

func (this *ScenePolicyBlackJack) CreatePlayerExData(s *base.Scene, p *base.Player) interface{} {
	if s != nil && p != nil {
		p.SetExtraData(NewBlackJackPlayerData(p))
	}
	return nil
}

func (this *ScenePolicyBlackJack) RegisterSceneState(state base.SceneState) {
	if state == nil {
		return
	}
	n := state.GetState()
	if n < 0 || n >= rule.StatusMax {
		return
	}
	this.states[n] = state
}

//=====================================
// 状态基类
//=====================================
type BaseBlackJackState struct {
}

func (this *BaseBlackJackState) GetState() int {
	return -1
}

func (this *BaseBlackJackState) CanChangeTo(s base.SceneState) bool {
	return true
}

func (this *BaseBlackJackState) CanChangeCoinScene(s *base.Scene, p *base.Player) bool {
	return !p.IsGameing() || p.IsMarkFlag(base.PlayerState_Audience) || !s.Gaming
}

func (this *BaseBlackJackState) GetTimeout(s *base.Scene) int {
	if _, ok := s.GetExtraData().(*BlackJackSceneData); ok {
		return int(time.Now().Sub(s.StateStartTime) / time.Second)
	}
	return 0
}

func (this *BaseBlackJackState) OnEnter(s *base.Scene) {
	if _, ok := s.GetExtraData().(*BlackJackSceneData); ok {
		s.StateStartTime = time.Now()
	}
}

func (this *BaseBlackJackState) OnLeave(s *base.Scene) {
}

func (this *BaseBlackJackState) OnTick(s *base.Scene) {
	//场景状态是所有房间公用，房间的私有属性不能放到场景状态上
	if time.Now().Unix() > s.GetTimerRandomRobot() {
		s.RandRobotCnt()
		s.SetTimerRandomRobot(s.GetRobotTime())
	}

	// 观战五分钟踢出房间
	for _, v := range s.GetAudiences() {
		if time.Now().Sub(v.LastOPTimer) > 5*time.Minute {
			s.AudienceLeave(v, common.PlayerLeaveReason_LongTimeNoOp)
		}
	}
}

func (this *BaseBlackJackState) OnPlayerOp(s *base.Scene, p *base.Player, code int, params []int64) bool {
	switch code {
	case rule.SubLeave:
		s.PlayerLeave(p, common.PlayerLeaveReason_Normal, false)
		return true
		//case rule.SubSit:
		//	// 座位已坐满
		//	if len(s.players) == rule.MaxPlayer {
		//		pack := &blackjack.SCBlackJackSit{Code: blackjack.SCBlackJackSit_ErrPos.Enum()}
		//		p.SendToClient(int(blackjack.BlackJackPacketID_SC_SIT), pack)
		//		return true
		//	}
		//	if s.GetPlayer(p.SnId) == nil {
		//		s.AudienceSit(p)
		//	}
		//	return true
	}
	return false
}

func (this *BaseBlackJackState) OnPlayerEvent(s *base.Scene, p *base.Player, code int, params []int64) {
	switch code {
	case base.PlayerEventEnter:
		if !s.Gaming {
			if !p.IsReady() {
				p.MarkFlag(base.PlayerState_Ready)
				p.SyncFlag()
			}
			if p.IsMarkFlag(base.PlayerState_WaitNext) {
				p.UnmarkFlag(base.PlayerState_WaitNext)
				p.SyncFlag()
			}
		}
	}
}

//=====================================
// 等待状态
//=====================================
type BlackJackWait struct {
	BaseBlackJackState
}

func (this *BlackJackWait) GetState() int {
	return rule.StatusWait
}

func (this *BlackJackWait) CanChangeTo(s base.SceneState) bool {
	if s.GetState() == rule.StatusReady {
		return true
	}
	return false
}

func (this *BlackJackWait) CanChangeCoinScene(s *base.Scene, p *base.Player) bool {
	return s.GetDBGameFree().GetMatchMode() == 0
}

func (this *BlackJackWait) GetTimeout(s *base.Scene) int {
	return 0
}

func (this *BlackJackWait) OnEnter(s *base.Scene) {
	this.BaseBlackJackState.OnEnter(s)
	logger.Logger.Trace("======== 等待状态 ======")
	// 通知场景停止
	if s.Gaming {
		s.NotifySceneRoundPause()
	}
	s.Gaming = false
	// 广播房间状态
	pack := &blackjack.SCBlackJackRoomStatus{
		Status: proto.Int(this.GetState()),
	}
	proto.SetDefaults(pack)
	s.Broadcast(int(blackjack.BlackJackPacketID_SC_ROOM_STATUS), pack, 0)
	logger.Logger.Trace("--> Status SCBlackJackRoomStatus ", pack.String())
}

func (this *BlackJackWait) OnTick(s *base.Scene) {
	this.BaseBlackJackState.OnTick(s)
	sceneEx, ok := s.GetExtraData().(*BlackJackSceneData)
	if !ok {
		return
	}
	if s.CheckNeedDestroy() {
		sceneEx.SceneDestroy(true)
		return
	}
	if s.GetRealPlayerCnt() == 0 {
		s.TryDismissRob()
	}
}

func (this *BlackJackWait) OnPlayerEvent(s *base.Scene, p *base.Player, code int, params []int64) {
	this.BaseBlackJackState.OnPlayerEvent(s, p, code, params)
	switch code {
	case base.PlayerEventEnter:
		if sceneEx, ok := s.GetExtraData().(*BlackJackSceneData); ok {
			if sceneEx.JudgeStart() {
				s.ChangeSceneState(rule.StatusReady)
			}
		}
	}
}

//=====================================
// 准备状态
//=====================================
type BlackJackReady struct {
	BaseBlackJackState
}

func (this *BlackJackReady) GetState() int {
	return rule.StatusReady
}

func (this *BlackJackReady) CanChangeTo(s base.SceneState) bool {
	switch s.GetState() {
	case rule.StatusWait, rule.StatusBet:
		return true
	}
	return false
}

func (this *BlackJackReady) CanChangeCoinScene(s *base.Scene, p *base.Player) bool {
	return s.GetDBGameFree().GetMatchMode() == 0
}

func (this *BlackJackReady) OnEnter(s *base.Scene) {
	this.BaseBlackJackState.OnEnter(s)
	logger.Logger.Trace("======== 准备状态 ======")

	hasLeave := false
	// 踢出机器人
	if s.GetRealPlayerCnt() >= 3 {
		if sceneEx, ok := s.GetExtraData().(*BlackJackSceneData); ok {
			for _, v := range sceneEx.chairIDs[1:] {
				if v.Player != nil && v.IsRobot() {
					s.PlayerLeave(v.Player, common.PlayerLeaveReason_DropLine, false)
					hasLeave = true
				}
			}
		}
	}
	//if s.gaming {
	//	s.NotifySceneRoundPause()
	//}
	s.Gaming = false
	sceneEx, ok := s.GetExtraData().(*BlackJackSceneData)
	if !ok {
		return
	}
	for _, v := range sceneEx.chairIDs[1:] {
		if v.Player == nil {
			continue
		}
		v.Billed = false
		// 掉线的玩家离开游戏
		if !v.IsMarkFlag(base.PlayerState_Online) {
			s.PlayerLeave(v.Player, common.PlayerLeaveReason_DropLine, false)
			hasLeave = true
			continue
		}
		// 钱不够踢出
		if !s.CoinInLimit(v.GetCoin()) {
			s.PlayerLeave(v.Player, common.PlayerLeaveReason_Bekickout, false)
			hasLeave = true
			continue
		}
		//// 机器人随机离开
		//if v.IsRobot() && s.RobotIsOverLimit() && s.RandInt()%3 == 1 {
		//	s.PlayerLeave(v.Player, common.PlayerLeaveReason_Normal, false)
		//	hasLeave = true
		//	continue
		//}
		//游戏次数达到目标值
		todayGamefreeIDSceneData, _ := v.GetDaliyGameData(int(s.GetDBGameFree().GetId()))
		if !v.IsRob &&
			todayGamefreeIDSceneData != nil &&
			s.GetDBGameFree().GetPlayNumLimit() != 0 &&
			todayGamefreeIDSceneData.GameTimes >= int64(s.GetDBGameFree().GetPlayNumLimit()) {
			s.PlayerLeave(v.Player, common.PlayerLeaveReason_GameTimes, false)
			hasLeave = true
			continue
		}
		// 多局未操作，踢出
		if v.noOptionTimes > int(model.GameParamData.NoOpTimes) && !v.IsRobot() {
			s.PlayerLeave(v.Player, common.PlayerLeaveReason_LongTimeNoOp, false)
			hasLeave = true
			continue
		}

		// 等待的玩家自动准备
		if v.IsMarkFlag(base.PlayerState_WaitNext) {
			v.UnmarkFlag(base.PlayerState_WaitNext)
			v.SyncFlag()
		}
		if !v.IsReady() {
			v.MarkFlag(base.PlayerState_Ready)
			v.SyncFlag()
		}
		//logger.Tracef("--> BlackJackReady Player: %+v", *v)
	}
	// 机器人离开规则
	hasLeave = hasLeave || sceneEx.RobotLeave(s)

	if !hasLeave && !s.IsRobFightGame() {
		//delay := time.Second * time.Duration(rand.Int63n(int64(rule.TimeoutReady/time.Second-1)))
		//if _, ok := common.DelayInvake(func() {
		s.TryDismissRob()
		//}, nil, delay, 1); ok {
		//}
	}

	if sceneEx.JudgeStart() {
		// 广播房间状态
		pack := &blackjack.SCBlackJackRoomStatus{
			Status: proto.Int(this.GetState()),
		}
		proto.SetDefaults(pack)
		s.Broadcast(int(blackjack.BlackJackPacketID_SC_ROOM_STATUS), pack, 0)
		logger.Logger.Trace("--> Status SCBlackJackRoomStatus ", pack.String())
	} else {
		s.ChangeSceneState(rule.StatusWait)
	}
}

func (this *BlackJackReady) OnTick(s *base.Scene) {
	this.BaseBlackJackState.OnTick(s)
	if sceneEx, ok := s.GetExtraData().(*BlackJackSceneData); ok {
		if time.Now().Sub(s.StateStartTime) > rule.TimeoutReady {
			if sceneEx.JudgeStart() {
				s.ChangeSceneState(rule.StatusBet)
			} else {
				s.ChangeSceneState(rule.StatusWait)
			}
		}
	}
}

//=====================================
// 下注状态
//=====================================
type BlackJackBet struct {
	BaseBlackJackState
}

func (this *BlackJackBet) GetState() int {
	return rule.StatusBet
}

func (this *BlackJackBet) CanChangeTo(s base.SceneState) bool {
	switch s.GetState() {
	case rule.StatusDeal:
		return true
	}
	return false
}

func (this *BlackJackBet) OnEnter(s *base.Scene) {
	this.BaseBlackJackState.OnEnter(s)
	logger.Logger.Trace("======== 下注状态 ======")
	if sceneEx, ok := s.GetExtraData().(*BlackJackSceneData); ok {

		// 剩余牌数小于 n*5*2+5时重新洗牌
		if s.GetGameingPlayerCnt()*10+5 > len(sceneEx.pokers) {
			rule.RandomShuffle(&sceneEx.pokers)
		}

		sceneEx.Gaming = true
		sceneEx.winResult = rule.ResultDefault

		// 通知玩家下注
		pack := &blackjack.SCBlackJackRoomStatus{
			Status: proto.Int(this.GetState()),
			Param:  sceneEx.BetMinMax(),
		}
		proto.SetDefaults(pack)

		// 通知所有玩家下注
		s.Broadcast(int(blackjack.BlackJackPacketID_SC_ROOM_STATUS), pack, 0)
		logger.Logger.Trace("--> Status SCBlackJackRoomStatus All Player Bet", pack.String())

		sceneEx.BetState()
	}
}

func (this *BlackJackBet) OnLeave(s *base.Scene) {
}

func (this *BlackJackBet) OnTick(s *base.Scene) {
	this.BaseBlackJackState.OnTick(s)
	if sceneEx, ok := s.GetExtraData().(*BlackJackSceneData); ok {
		if time.Now().Sub(s.StateStartTime) > rule.TimeoutBet {
			// 自动下最小注
			sceneEx.AutoBet(sceneEx.BetMinMax()[0])
			// 确定1号位
			sceneEx.SortSeat()
			s.ChangeSceneState(rule.StatusDeal)
		}
	}
}

func (this *BlackJackBet) OnPlayerOp(s *base.Scene, p *base.Player, code int, params []int64) bool {
	if this.BaseBlackJackState.OnPlayerOp(s, p, code, params) {
		return true
	}
	if !p.IsGameing() || len(params) != 2 || code != rule.SubBet {
		return false
	}
	playerEx, ok := p.GetExtraData().(*BlackJackPlayerData)
	if !ok {
		return false
	}
	sceneEx, ok := s.GetExtraData().(*BlackJackSceneData)
	if !ok {
		return false
	}
	pack := &blackjack.SCBlackJackPlayerBet{
		Code: blackjack.SCBlackJackPlayerBet_Success,
		Pos1: proto.Int32(int32(p.GetPos())),
		Pos2: proto.Int32(int32(params[0])),
	}

	// 取整
	if params[1]%10 != 0 {
		params[1] = (params[1] / 10) * 10
	}
	// 金额检查
	if params[1] < sceneEx.BetMinMax()[0] || params[1] > sceneEx.BetMinMax()[1] || params[1] > sceneEx.PlayerCoin(playerEx) {
		pack.Code = blackjack.SCBlackJackPlayerBet_ErrCoin
		p.SendToClient(int(blackjack.BlackJackPacketID_SC_PLAYER_BET), pack)
		return true
	}
	// 下注参数为两个，第一个是下注位置，第二个是下注金额
	if int(params[0]) != playerEx.GetPos() {
		pack.Code = blackjack.SCBlackJackPlayerBet_ErrPos
		p.SendToClient(int(blackjack.BlackJackPacketID_SC_PLAYER_BET), pack)
		return true
	}
	if playerEx.hands[0].betCoin > 0 {
		pack.Code = blackjack.SCBlackJackPlayerBet_ErrBet
		p.SendToClient(int(blackjack.BlackJackPacketID_SC_PLAYER_BET), pack)
		return true
	}
	playerEx.isBet = true
	playerEx.noOptionTimes = 0
	playerEx.hands[0].betCoin = params[1]
	// 广播下注结果
	pack.Coin = proto.Int64(params[1])
	pack.ReCoin = proto.Int64(sceneEx.PlayerCoin(playerEx))
	sceneEx.Broadcast(int(blackjack.BlackJackPacketID_SC_PLAYER_BET), pack, 0)
	logger.Logger.Trace("--> Bet SCBlackJackPlayerBet_BetCoin ", pack.String())
	// 记录下注花费时长
	playerEx.betTime = s.SceneState.GetTimeout(sceneEx.Scene)
	// 游戏中的玩家都下注后直接开始发牌
	allBet := true
	for _, v := range sceneEx.chairIDs[1:] {
		if v.Player != nil && v.IsGameing() && !v.isBet {
			allBet = false
			break
		}
	}
	if allBet {
		// 确定1号位
		sceneEx.SortSeat()
		s.ChangeSceneState(rule.StatusDeal)
	}
	return true
}

//=====================================
// 发牌状态
//=====================================
type BlackJackDeal struct {
	BaseBlackJackState
}

func (this *BlackJackDeal) GetState() int {
	return rule.StatusDeal
}

func (this *BlackJackDeal) CanChangeTo(s base.SceneState) bool {
	switch s.GetState() {
	case rule.StatusPlayer, rule.StatusBuy:
		return true
	}
	return false
}

func (this *BlackJackDeal) OnEnter(s *base.Scene) {
	this.BaseBlackJackState.OnEnter(s)
	logger.Logger.Trace("======== 发牌状态 ======")
	sceneEx, ok := s.GetExtraData().(*BlackJackSceneData)
	if !ok {
		return
	}
	pack := &blackjack.SCBlackJackRoomStatus{
		Status: proto.Int(this.GetState()),
	}
	proto.SetDefaults(pack)
	s.Broadcast(int(blackjack.BlackJackPacketID_SC_ROOM_STATUS), pack, 0)
	logger.Logger.Trace("--> Status SCBlackJackRoomStatus", pack.String())

	sceneEx.FaPai()
	// 通知发牌结果
	deal := &blackjack.SCBlackJackDeal{
		Num: proto.Int32(int32(len(sceneEx.pokers))),
	}
	for _, v := range sceneEx.chairIDs {
		if !v.isBet && !v.isBanker {
			continue
		}
		seat := &blackjack.BlackJackCards{
			Cards:   rule.CardsToInt32(v.hands[0].handCards),
			DCards:  proto.Int32(int32(v.hands[0].mulCards.Value())),
			Type:    proto.Int32(v.hands[0].tp),
			Point:   v.hands[0].point,
			State:   proto.Int32(v.hands[0].state),
			Seat:    proto.Int32(int32(v.seat)),
			BetCoin: proto.Int64(v.hands[0].betCoin),
		}
		if v.isBanker {
			seat.Cards[1] = rule.OneCard
			seat.Type = proto.Int32(0)
			seat.Point = []int32{}
			seat.State = proto.Int32(0)
		}
		proto.SetDefaults(seat)
		deal.Seats = append(deal.Seats, seat)
	}
	s.Broadcast(int(blackjack.BlackJackPacketID_SC_DEAL), deal, 0)
	logger.Logger.Trace("--> Deal SCBlackJackDeal ", deal.String())
}

func (this *BlackJackDeal) OnTick(s *base.Scene) {
	this.BaseBlackJackState.OnTick(s)
	if sceneEx, ok := s.GetExtraData().(*BlackJackSceneData); ok {
		a := time.Now().Sub(s.StateStartTime) * 1000
		b := sceneEx.timeOutDeal
		if a > b {
			if sceneEx.chairIDs[0].hands[0].handCards[0].Point() == 1 {
				s.ChangeSceneState(rule.StatusBuy)
			} else {
				s.ChangeSceneState(rule.StatusPlayer)
			}
		}
	}
}

//=====================================
// 买保险状态
//=====================================
type BlackJackBuy struct {
	BaseBlackJackState
}

func (this *BlackJackBuy) GetState() int {
	return rule.StatusBuy
}

func (this *BlackJackBuy) CanChangeTo(s base.SceneState) bool {
	switch s.GetState() {
	case rule.StatusPlayer, rule.StatusBuyEnd:
		return true
	}
	return false
}

func (this *BlackJackBuy) OnEnter(s *base.Scene) {
	this.BaseBlackJackState.OnEnter(s)
	logger.Logger.Trace("======== 买保险状态 ======")
	sceneEx, ok := s.GetExtraData().(*BlackJackSceneData)
	if !ok {
		return
	}
	sceneEx.pos = 0
	sceneEx.buy = false
	// 广播房间状态
	pack := &blackjack.SCBlackJackRoomStatus{
		Status: proto.Int(this.GetState()),
	}
	proto.SetDefaults(pack)
	s.Broadcast(int(blackjack.BlackJackPacketID_SC_ROOM_STATUS), pack, 0)
	logger.Logger.Trace("--> Status SCBlackJackRoomStatus", pack.String())
	// 通知买保险
	sceneEx.NotifyBuy()
}

func (this *BlackJackBuy) OnLeave(s *base.Scene) {
}

func (this *BlackJackBuy) OnTick(s *base.Scene) {
	this.BaseBlackJackState.OnTick(s)
	if sceneEx, ok := s.GetExtraData().(*BlackJackSceneData); ok {
		if time.Now().Sub(s.StateStartTime) > rule.TimeoutBuy {
			// 不买保险
			pack := &blackjack.SCBlackJackBuy{
				Code: blackjack.SCBlackJackBuy_UnBuy,
				Pos:  proto.Int32(int32(sceneEx.chairIDs[sceneEx.pos].seat)),
			}
			sceneEx.Broadcast(int(blackjack.BlackJackPacketID_SC_PLAYER_BUY), pack, 0)
			logger.Logger.Trace("--> unBuy SCBlackJackBuy ", pack.String())
			// 是否还有下一个玩家
			if !sceneEx.NotifyBuy() {
				// 是否进行保险结算
				if sceneEx.HasBuy() {
					s.ChangeSceneState(rule.StatusBuyEnd)
				} else {
					s.ChangeSceneState(rule.StatusPlayer)
				}
			}
		}
	}
}

func (this *BlackJackBuy) OnPlayerOp(s *base.Scene, p *base.Player, code int, params []int64) bool {
	if this.BaseBlackJackState.OnPlayerOp(s, p, code, params) {
		return true
	}
	if !p.IsGameing() || len(params) != 1 || code != rule.SubBuy {
		return false
	}
	// 参数是否正确
	if params[0] != 0 && params[0] != 1 {
		return false
	}
	playerEx, ok := p.GetExtraData().(*BlackJackPlayerData)
	if !ok {
		return false
	}
	sceneEx, ok := s.GetExtraData().(*BlackJackSceneData)
	if !ok {
		return false
	}
	if sceneEx.pos > rule.MaxPlayer {
		return false
	}
	seat := sceneEx.chairIDs[sceneEx.pos] // 买保险的座位
	pack := &blackjack.SCBlackJackBuy{
		Code: blackjack.SCBlackJackBuy_Success,
		Pos:  proto.Int32(int32(seat.seat)),
		Coin: proto.Int64(seat.hands[0].betCoin / 2),
	}
	// 是否该玩家操作
	if seat.SnId != p.SnId {
		pack.Code = blackjack.SCBlackJackBuy_ErrPos
		p.SendToClient(int(blackjack.BlackJackPacketID_SC_PLAYER_BUY), pack)
		logger.Logger.Trace("--> buy error SCBlackJackBuy ", pack.String())
		return true
	}

	if params[0] == 1 {
		// 买保险金额是否够用(下注金额+下注额的一半 < 玩家金额总额)
		if seat.GetCoin() < playerEx.hands[0].betCoin+playerEx.hands[0].betCoin/2 {
			pack.Code = blackjack.SCBlackJackBuy_ErrCoin
			p.SendToClient(int(blackjack.BlackJackPacketID_SC_PLAYER_BUY), pack)
			logger.Logger.Trace("--> buy error SCBlackJackBuy ", pack.String())
			return true
		}
		// 买保险
		sceneEx.Buy()
		seat.noOptionTimes = 0
	} else {
		// 不买保险
		pack.Code = blackjack.SCBlackJackBuy_UnBuy
		pack.Coin = proto.Int64(0)
		sceneEx.Broadcast(int(blackjack.BlackJackPacketID_SC_PLAYER_BUY), pack, 0)
		logger.Logger.Trace("--> unBuy SCBlackJackBuy ", pack.String())
		seat.noOptionTimes = 0
	}
	if !sceneEx.NotifyBuy() {
		// 是否进行保险结算
		if sceneEx.HasBuy() {
			s.ChangeSceneState(rule.StatusBuyEnd)
		} else {
			s.ChangeSceneState(rule.StatusPlayer)
		}
	}
	return true
}

//=====================================
// 保险结算
//=====================================
type BlackJackBuyEnd struct {
	BaseBlackJackState
}

func (this *BlackJackBuyEnd) GetState() int {
	return rule.StatusBuyEnd
}

func (this *BlackJackBuyEnd) CanChangeTo(s base.SceneState) bool {
	switch s.GetState() {
	case rule.StatusPlayer, rule.StatusEnd:
		return true
	}
	return false
}

func (this *BlackJackBuyEnd) OnEnter(s *base.Scene) {
	this.BaseBlackJackState.OnEnter(s)
	logger.Logger.Trace("======== 保险结算状态 ======")
	sceneEx, ok := s.GetExtraData().(*BlackJackSceneData)
	if !ok {
		return
	}
	// 庄家是黑杰克，翻牌，并结算
	// 否则，游戏继续
	if sceneEx.chairIDs[0].hands[0].tp == rule.CardTypeA10 {
		sceneEx.NotifyCards(nil)
	}
	sceneEx.n = -1
	s.StateStartTime = time.Now()
}

func (this *BlackJackBuyEnd) OnTick(s *base.Scene) {
	this.BaseBlackJackState.OnTick(s)
	if sceneEx, ok := s.GetExtraData().(*BlackJackSceneData); ok {
		if sceneEx.n == -1 {
			if time.Now().Sub(s.StateStartTime) > 2*time.Second {
				pack := &blackjack.SCBlackJackEnd{}
				if sceneEx.chairIDs[0].hands[0].tp == rule.CardTypeA10 {
					pack.IsBlackJack = proto.Bool(true)
				}
				// 结算
				sceneEx.n = 0
				for _, v := range sceneEx.chairIDs[1:] {
					if v.isBuy {
						sceneEx.n++
						if sceneEx.chairIDs[0].hands[0].tp == rule.CardTypeA10 {
							// 赢
							gain := v.baoCoin
							v.revenue = int64(math.Ceil(float64(gain) * float64(sceneEx.GetDBGameFree().GetTaxRate()) / 10000))
							v.baoChange = gain - v.revenue
						} else {
							// 输
							v.baoChange = -v.baoCoin
						}
						pack.Players = append(pack.Players, &blackjack.BlackJackPlayerEnd{
							Pos:       proto.Int32(int32(v.GetPos())),
							Gain:      proto.Int64(v.baoChange),
							LeftGain:  proto.Int64(v.LeftBetChange()),
							RightGain: proto.Int64(v.RightBetChange()),
							Coin:      proto.Int64(sceneEx.PlayerCoin(v)),
						})
					}
				}
				proto.SetDefaults(pack)
				sceneEx.Broadcast(int(blackjack.BlackJackPacketID_SC_BUY_END), pack, 0)
				logger.Logger.Trace("--> Broadcast SCBlackJackEnd", pack.String())
				s.StateStartTime = time.Now()
			}
		} else {
			state := rule.StatusPlayer
			if sceneEx.chairIDs[0].hands[0].tp == rule.CardTypeA10 { // 庄家是黑杰克
				state = rule.StatusEnd
			}
			if time.Now().Sub(s.StateStartTime) > rule.TimeoutBuyEnd {
				s.ChangeSceneState(state)
			}
			//if time.Now().Sub(sceneEx.stateStartTime) > rule.TimeoutBuyEnd*time.Duration(sceneEx.n) {
			//	// 庄家是黑杰克
			//	if sceneEx.chairIDs[0].hands[0].tp == rule.CardTypeA10 {
			//		s.ChangeSceneState(rule.StatusEnd)
			//	} else {
			//		s.ChangeSceneState(rule.StatusPlayer)
			//	}
			//}
		}
	}
}

//=====================================
// 闲家操作状态
//=====================================
type BlackJackPlayer struct {
	BaseBlackJackState
}

func (this *BlackJackPlayer) GetState() int {
	return rule.StatusPlayer
}

func (this *BlackJackPlayer) CanChangeTo(s base.SceneState) bool {
	switch s.GetState() {
	case rule.StatusBanker:
		return true
	}
	return false
}

func (this *BlackJackPlayer) OnEnter(s *base.Scene) {
	this.BaseBlackJackState.OnEnter(s)
	logger.Logger.Trace("======== 闲家操作状态 ======")
	sceneEx, ok := s.GetExtraData().(*BlackJackSceneData)
	if !ok {
		return
	}
	sceneEx.pos = 0
	sceneEx.lastOperaPos = -1
	// 广播房间状态
	pack := &blackjack.SCBlackJackRoomStatus{
		Status: proto.Int(this.GetState()),
	}
	proto.SetDefaults(pack)
	s.Broadcast(int(blackjack.BlackJackPacketID_SC_ROOM_STATUS), pack, 0)
	logger.Logger.Trace("--> Status SCBlackJackRoomStatus", pack.String())
	// 通知闲家操作
	if !sceneEx.NotifyPlayer() {
		s.ChangeSceneState(rule.StatusBanker)
	}
}

func (this *BlackJackPlayer) OnTick(s *base.Scene) {
	this.BaseBlackJackState.OnTick(s)
	if sceneEx, ok := s.GetExtraData().(*BlackJackSceneData); ok {
		//if sceneEx.n < 0 && time.Now().Sub(sceneEx.stateStartTime) > time.Second*3 {
		//	if !sceneEx.NotifyPlayer() {
		//		s.ChangeSceneState(rule.StatusBanker)
		//	}
		//	return
		//}
		if time.Now().Sub(s.StateStartTime) > rule.TimeoutPlayer {
			// 停牌
			if sceneEx.pos > rule.MaxPlayer {
				s.ChangeSceneState(rule.StatusBanker)
				return
			}
			seat := sceneEx.chairIDs[sceneEx.pos]

			sceneEx.AllSkip(seat)
			// 是否还有下一个玩家需要操作
			if !sceneEx.NotifyPlayer() {
				s.ChangeSceneState(rule.StatusBanker)
				return
			}
		}
	}
}

func (this *BlackJackPlayer) OnPlayerOp(s *base.Scene, p *base.Player, code int, params []int64) bool {
	if this.BaseBlackJackState.OnPlayerOp(s, p, code, params) {
		return true
	}
	if !p.IsGameing() {
		return false
	}
	pExtra, ok := p.GetExtraData().(*BlackJackPlayerData)
	if !ok {
		return false
	}
	pExtra.opCodeFlag = int32(code)
	sceneEx, ok := s.GetExtraData().(*BlackJackSceneData)
	if !ok {
		return false
	}
	if sceneEx.pos > rule.MaxPlayer {
		return false
	}
	seat := sceneEx.chairIDs[sceneEx.pos]
	// 是否该当前玩家操作
	if seat.SnId != p.SnId {
		pack := &blackjack.SCBlackJackPlayerOperate{
			Code:    blackjack.SCBlackJackPlayerOperate_ErrPos,
			Operate: proto.Int32(int32(code)),
			Pos:     proto.Int32(int32(seat.seat)),
		}
		proto.SetDefaults(pack)
		p.SendToClient(int(blackjack.BlackJackPacketID_SC_PLAYER_OPERATE), pack)
		logger.Logger.Trace("--> SCBlackJackPlayerOperate ", pack.String())
		return true
	}
	seat.noOptionTimes = 0

	var su bool
	switch code {
	case rule.SubFenPai: // 分牌
		su = sceneEx.FenPai(seat)
	case rule.SubDouble: // 双倍
		su = sceneEx.Double(seat)
	case rule.SubOuts: // 要牌
		su = sceneEx.Outs(seat)
	case rule.SubSkip: // 停牌
		su = sceneEx.Skip(seat)
	default:
		return false
	}
	if sceneEx.hTimerOpDelay != timer.TimerHandle(0) {
		timer.StopTimer(sceneEx.hTimerOpDelay)
		sceneEx.hTimerOpDelay = timer.TimerHandle(0)
	}
	// 提前结算
	if seat.hands[0].tp == rule.CardTypeBoom && (len(seat.hands[1].handCards) == 0 || seat.hands[1].tp == rule.CardTypeBoom) {
		sceneEx.calculate(sceneEx.chairIDs[0], seat)
		sceneEx.result(seat, sceneEx.pos)
		seat.Billed = true
		//seat.MarkFlag(PlayerState_WaitNext)
		seat.MarkFlag(base.PlayerState_Lose)
		seat.SyncFlag()
	}
	if su {
		if !sceneEx.NotifyPlayer() {
			s.ChangeSceneState(rule.StatusBanker)
		}
	}
	return true
}

//=====================================
// 庄家操作状态
//=====================================
type BlackJackBanker struct {
	BaseBlackJackState
}

func (this *BlackJackBanker) GetState() int {
	return rule.StatusBanker
}

func (this *BlackJackBanker) CanChangeTo(s base.SceneState) bool {
	switch s.GetState() {
	case rule.StatusEnd:
		return true
	}
	return false
}

func (this *BlackJackBanker) OnEnter(s *base.Scene) {
	this.BaseBlackJackState.OnEnter(s)
	logger.Logger.Trace("======== 庄家操作状态 ======")
	sceneEx, ok := s.GetExtraData().(*BlackJackSceneData)
	if !ok {
		return
	}
	// 广播房间状态
	pack := &blackjack.SCBlackJackRoomStatus{
		Status: proto.Int(this.GetState()),
	}
	proto.SetDefaults(pack)
	s.Broadcast(int(blackjack.BlackJackPacketID_SC_ROOM_STATUS), pack, 0)
	logger.Logger.Trace("--> Status SCBlackJackRoomStatus", pack.String())

	hand := &sceneEx.chairIDs[0].hands[0]
	sceneEx.NotifyCards(nil)
	if hand.tp == rule.CardTypeA10 || hand.point[len(hand.point)-1] >= 17 {
		sceneEx.chairIDs[0].hands[0].state = 1
	}
	sceneEx.lastTime = time.Now()
	//sceneEx.MaxHand()
}

func (this *BlackJackBanker) OnTick(s *base.Scene) {
	this.BaseBlackJackState.OnTick(s)
	// 系统庄要牌
	// 黑杰克：直接结算
	// 小于17点：要牌
	// 大于等于17点：停牌
	if sceneEx, ok := s.GetExtraData().(*BlackJackSceneData); ok {
		if time.Now().Sub(sceneEx.lastTime) > time.Second {
			banker := sceneEx.chairIDs[0]
			hand := &banker.hands[0]
			if len(hand.handCards) >= rule.MaxCardNum || hand.tp == rule.CardTypeInvalid ||
				hand.tp == rule.CardTypeA10 || hand.point[len(hand.point)-1] >= 17 || sceneEx.AllBilled() /*|| sceneEx.BankerStop()*/ {
				hand.state = 1
				s.ChangeSceneState(rule.StatusEnd)
				return
			}
			// 根据水池状态调控补牌
			if !sceneEx.Testing && hand.point[0] > 11 && model.IsUseCoinPoolControlGame(sceneEx.GetKeyGameId(), sceneEx.GetGameFreeId()) {
				status, _ := base.GetCoinPoolMgr().GetCoinPoolStatus(sceneEx.GetPlatform(), sceneEx.GetGameFreeId(), sceneEx.GetGroupId())
				switch status {
				case base.CoinPoolStatus_Low:
					// 100% 不爆牌
					sceneEx.le(21 - int(hand.point[0]))
				case base.CoinPoolStatus_High:
					// 80% 爆牌
					if sceneEx.RandInt(100) < 80 {
						sceneEx.gt(21 - int(hand.point[0]))
					}
				case base.CoinPoolStatus_TooHigh:
					// 100% 爆牌
					sceneEx.gt(21 - int(hand.point[0]))
				}
			}
			(*hand).handCards = append((*hand).handCards, sceneEx.GetCard(1)[0])
			hand.tp, hand.point = rule.GetCardsType((*hand).handCards)
			sceneEx.NotifyCards(nil)
			logger.Logger.Trace("--> banker SCBlackJackNotifyCards", hand)
			sceneEx.lastTime = time.Now()
		}
	}
}

//=====================================
// 结算状态
//=====================================
type BlackJackEnd struct {
	BaseBlackJackState
}

func (this *BlackJackEnd) GetState() int {
	return rule.StatusEnd
}

func (this *BlackJackEnd) CanChangeTo(s base.SceneState) bool {
	switch s.GetState() {
	case rule.StatusReady:
		return true
	}
	return false
}

func (this *BlackJackEnd) OnEnter(s *base.Scene) {
	this.BaseBlackJackState.OnEnter(s)
	logger.Logger.Trace("======== 结算状态 ======")
	sceneEx, ok := s.GetExtraData().(*BlackJackSceneData)
	if !ok {
		return
	}
	// 广播房间状态
	pack := &blackjack.SCBlackJackRoomStatus{
		Status: proto.Int(this.GetState()),
	}
	proto.SetDefaults(pack)
	s.Broadcast(int(blackjack.BlackJackPacketID_SC_ROOM_STATUS), pack, 0)
	logger.Logger.Trace("--> Status SCBlackJackRoomStatus", pack.String())
	// 结算
	sceneEx.CalculateResult()
	// 广播结算
	s.Broadcast(int(blackjack.BlackJackPacketID_SC_END), sceneEx.End(), 0)
	s.StateStartTime = time.Now()

	//////////////////////////结算之后
	s.NotifySceneRoundPause()
	s.Gaming = false

	//结束记录录像
	sceneEx.RecordReplayOver()

	// 重置座位数据
	for k, v := range sceneEx.chairIDs {
		//if v.Player != nil && v.IsGameing() && v.pos != v.seat {
		//	seatId = append(seatId, int32(v.seat))
		//}
		v.Release()
		if k < rule.MaxPlayer {
			sceneEx.backup[k] = nil
		}
	}

	//防伙牌机制
	//samePlaceLimit := sceneEx.GetDBGameFree().GetSamePlaceLimit()
	//timer.StartTimer(timer.TimerActionWrapper(func(h timer.TimerHandle, ud interface{}) bool {
	//	if sceneEx.GetDBGameFree().GetMatchMode() == 1 {
	//		return true
	//	}
	//	for _, value := range sceneEx.players {
	//		if value.IsRob {
	//			continue
	//		}
	//		if samePlaceLimit > 0 && value.gameTimes >= samePlaceLimit {
	//			if !value.IsRob && sceneEx.ClubId == 0 {
	//				//TODO 强制换桌
	//				opPack := &blackjack.CSCoinSceneOp{
	//					Id:     proto.Int32(sceneEx.GetDBGameFree().GetId()),
	//					OpType: proto.Int32(common.CoinSceneOP_Server),
	//				}
	//				common.TransmitToServer(value.sid, int(blackjack.CoinSceneGamePacketID_PACKET_CS_COINSCENE_OP), opPack, value.worldSess)
	//			}
	//		}
	//	}
	//	return true
	//}), nil, time.Second*3, 1)
	s.ChangeSceneEvent()
}

func (this *BlackJackEnd) OnLeave(s *base.Scene) {
}

func (this *BlackJackEnd) OnTick(s *base.Scene) {
	this.BaseBlackJackState.OnTick(s)
	if sceneEx, ok := s.GetExtraData().(*BlackJackSceneData); ok {
		//if time.Now().Sub(sceneEx.stateStartTime) > rule.TimeoutEnd {
		timeOut := rule.TimeoutEndLost
		if sceneEx.winResult == rule.ResultWin {
			timeOut = rule.TimeoutEndWin
		} else if sceneEx.winResult == rule.ResultWinAndLost {
			timeOut = rule.TimeoutEndWinAndLost
		}
		a := time.Now().Sub(s.StateStartTime)
		if a > timeOut {
			//if time.Now().Sub(sceneEx.stateStartTime) > rule.TimeoutEnd {
			//if time.Now().Sub(sceneEx.stateStartTime) > rule.TimeoutEnd*time.Duration(sceneEx.n) {
			if s.CheckNeedDestroy() || s.GetDBGameFree().GetMatchMode() == 1 {
				for _, p := range s.Players {
					s.PlayerLeave(p, common.PlayerLeaveReason_Normal, true)
				}
				sceneEx.SceneDestroy(true)
			} else {
				s.ChangeSceneState(rule.StatusReady)
			}
		}
	}
}

func (this *BlackJackEnd) OnPlayerEvent(s *base.Scene, p *base.Player, code int, params []int64) {
	switch code {
	case base.PlayerEventRehold:
		if sceneEx, ok := s.GetExtraData().(*BlackJackSceneData); ok {
			// 返回结算信息
			p.SendToClient(int(blackjack.BlackJackPacketID_SC_END), sceneEx.End())
		}
	}
}

func init() {
	blackJackSingleton.RegisterSceneState(&BlackJackWait{})
	blackJackSingleton.RegisterSceneState(&BlackJackReady{})
	blackJackSingleton.RegisterSceneState(&BlackJackBet{})
	blackJackSingleton.RegisterSceneState(&BlackJackDeal{})
	blackJackSingleton.RegisterSceneState(&BlackJackBuy{})
	blackJackSingleton.RegisterSceneState(&BlackJackBuyEnd{})
	blackJackSingleton.RegisterSceneState(&BlackJackPlayer{})
	blackJackSingleton.RegisterSceneState(&BlackJackBanker{})
	blackJackSingleton.RegisterSceneState(&BlackJackEnd{})
	core.RegisteHook(core.HOOK_BEFORE_START, func() error {
		base.RegisteScenePolicy(common.GameId_BlackJack, rule.RoomModeClassic, blackJackSingleton)
		return nil
	})
}
