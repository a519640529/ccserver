package tienlen

import (
	"fmt"
	"games.yol.com/win88/common"
	rule "games.yol.com/win88/gamerule/tienlen"
	"games.yol.com/win88/gamesrv/base"
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/tienlen"
	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/logger"
	"math/rand"
	"sort"
	"time"
)

var ScenePolicyTienLenSington = &ScenePolicyTienLen{}

type ScenePolicyTienLen struct {
	base.BaseScenePolicy
	states [rule.TienLenSceneStateMax]base.SceneState
}

// 创建场景扩展数据
func (this *ScenePolicyTienLen) CreateSceneExData(s *base.Scene) interface{} {
	sceneEx := NewTienLenSceneData(s)
	if sceneEx != nil {
		sceneEx.Clear()
		if sceneEx.init() {
			s.SetExtraData(sceneEx)
		}
	}
	return sceneEx
}

// 创建玩家扩展数据
func (this *ScenePolicyTienLen) CreatePlayerExData(s *base.Scene, p *base.Player) interface{} {
	playerEx := &TienLenPlayerData{Player: p}
	if playerEx != nil {
		playerEx.init()
		p.SetExtraData(playerEx)
	}
	return playerEx
}

// 场景开启事件
func (this *ScenePolicyTienLen) OnStart(s *base.Scene) {
	logger.Logger.Trace("(this *ScenePolicyTienLen) OnStart, GetSceneId()=", s.GetSceneId())

	sceneEx := NewTienLenSceneData(s)
	if sceneEx != nil {
		sceneEx.Clear()
		if sceneEx.init() {
			s.SetExtraData(sceneEx)
			s.ChangeSceneState(rule.TienLenSceneStateWaitPlayer)
		}
	}
}

// 场景心跳事件
func (this *ScenePolicyTienLen) OnTick(s *base.Scene) {
	if s == nil {
		return
	}
	if s.GetSceneState() != nil {
		s.GetSceneState().OnTick(s)
	}
}

// 玩家进入事件
func (this *ScenePolicyTienLen) OnPlayerEnter(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyTienLen) OnPlayerEnter, GetSceneId()=", s.GetSceneId(), " player=", p.GetName())
	if sceneEx, ok := s.GetExtraData().(*TienLenSceneData); ok {

		//自动带入金币
		pos := sceneEx.FindOnePos()
		if p.Pos != -1 {
			pos = p.Pos
		}
		if pos < 0 || pos > 3 {
			p.MarkFlag(base.PlayerState_EnterSceneFailed)
			cnt := len(sceneEx.players)
			logger.Logger.Warnf("ScenePolicyTienLen.OnPlayerEnter(scene:%v, player:%v) no found fit GetPos(), current player count:%v NumOfGames:%v", s.GetSceneId(), p.SnId, cnt, sceneEx.NumOfGames)
			return
		}

		playerEx := &TienLenPlayerData{Player: p}
		playerEx.init()
		playerEx.SetPos(pos)
		if sceneEx.GetGaming() {
			playerEx.MarkFlag(base.PlayerState_WaitNext)
		}
		p.SetExtraData(playerEx)
		sceneEx.seats[pos] = playerEx
		sceneEx.players[p.SnId] = playerEx
		logger.Logger.Trace("广播个人信息,curSeatsNum: ", sceneEx.GetSeatPlayerCnt())
		if sceneEx.GetSeatPlayerCnt() == 1 {
			sceneEx.masterSnid = playerEx.SnId
			sceneEx.BroadcastUpdateMasterSnid(false)
		}

		//广播个人信息
		playerData := TienLenCreatePlayerData(p)
		pack := &tienlen.SCTienLenPlayerEnter{
			Data: playerData,
		}
		proto.SetDefaults(pack)
		s.Broadcast(int(tienlen.TienLenPacketID_PACKET_SCTienLenPlayerEnter), pack, p.GetSid())

		//给自己发送房间信息
		TienLenSendRoomInfo(s, p, sceneEx, playerEx)

		s.FirePlayerEvent(p, base.PlayerEventEnter, nil)
	}
}

// 玩家离开事件
func (this *ScenePolicyTienLen) OnPlayerLeave(s *base.Scene, p *base.Player, reason int) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyTienLen) OnPlayerLeave, GetSceneId()=", s.GetSceneId(), " player=", p.SnId)
	if sceneEx, ok := s.GetExtraData().(*TienLenSceneData); ok {
		if playerEx, ok := p.GetExtraData().(*TienLenPlayerData); ok {
			if this.CanChangeCoinScene(s, p) {
				LeavePlayerSnid := playerEx.SnId

				sceneEx.OnPlayerLeave(p, reason)
				s.FirePlayerEvent(p, base.PlayerEventLeave, []int64{int64(reason)})

				if LeavePlayerSnid == sceneEx.masterSnid {
					findOne := false
					for i := 0; i < sceneEx.GetPlayerNum(); i++ {
						playerExSeat := sceneEx.seats[i]
						if playerExSeat != nil && playerExSeat.SnId != sceneEx.masterSnid {
							sceneEx.masterSnid = playerExSeat.SnId
							sceneEx.BroadcastUpdateMasterSnid(true)
							findOne = true
							break
						}
					}
					if !findOne {
						sceneEx.masterSnid = 0
						sceneEx.BroadcastUpdateMasterSnid(false)
					}
				}
				if sceneEx.lastWinSnid == LeavePlayerSnid {
					sceneEx.lastWinSnid = 0
				}

			}
		}

	}
}

// 玩家掉线
func (this *ScenePolicyTienLen) OnPlayerDropLine(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyTienLen) OnPlayerDropLine, GetSceneId()=", s.GetSceneId(), " player=", p.Name)
	s.FirePlayerEvent(p, base.PlayerEventDropLine, nil)
	if sceneEx, ok := s.GetExtraData().(*TienLenSceneData); ok {
		if sceneEx.GetGaming() {
			if s.IsMatchScene() || s.IsCoinScene() {
				p.MarkFlag(base.PlayerState_Auto)
				p.SyncFlag()
			}
			return
		} else {
			s.PlayerLeave(p, common.PlayerLeaveReason_DropLine, true)
		}
	}
}

// 玩家重连
func (this *ScenePolicyTienLen) OnPlayerRehold(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyTienLen) OnPlayerRehold, GetSceneId()=", s.GetSceneId(), " player=", p.Name)
	if sceneEx, ok := s.GetExtraData().(*TienLenSceneData); ok {
		if playerEx, ok := p.GetExtraData().(*TienLenPlayerData); ok {
			if p.IsMarkFlag(base.PlayerState_Auto) {
				p.UnmarkFlag(base.PlayerState_Auto)
				p.SyncFlag()
			}

			//发送房间信息给自己
			TienLenSendRoomInfo(s, p, sceneEx, playerEx)
			s.FirePlayerEvent(p, base.PlayerEventRehold, nil)
		}
	}
}

// 返回房间
func (this *ScenePolicyTienLen) OnPlayerReturn(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyTienLen) OnPlayerReturn, GetSceneId()=", s.GetSceneId(), " player=", p.Name)
	if sceneEx, ok := s.GetExtraData().(*TienLenSceneData); ok {
		if playerEx, ok := p.GetExtraData().(*TienLenPlayerData); ok {
			if p.IsMarkFlag(base.PlayerState_Auto) {
				p.UnmarkFlag(base.PlayerState_Auto)
				p.SyncFlag()
			}

			//发送房间信息给自己
			TienLenSendRoomInfo(s, p, sceneEx, playerEx)
			s.FirePlayerEvent(p, base.PlayerEventReturn, nil)
		}
	}
}

func (this *ScenePolicyTienLen) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if s == nil || p == nil {
		return false
	}
	logger.Logger.Trace("(this *ScenePolicyTienLen) OnPlayerOp, GetSceneId()=", s.GetSceneId(), " player=", p.GetName(), " opcode=", opcode, " params=", params)
	if s.GetSceneState() != nil {
		if s.GetSceneState().OnPlayerOp(s, p, opcode, params) {
			p.LastOPTimer = time.Now()
			return true
		}
		return false
	}
	return true
}

func (this *ScenePolicyTienLen) OnPlayerEvent(s *base.Scene, p *base.Player, evtcode int, params []int64) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyTienLen) OnPlayerEvent, GetSceneId()=", s.GetSceneId(), " player=", p.GetName(), " eventcode=", evtcode, " params=", params)
	if s.GetSceneState() != nil {
		s.GetSceneState().OnPlayerEvent(s, p, evtcode, params)
	}
}

func (this *ScenePolicyTienLen) OnAudienceEnter(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("ScenePolicyTienLen OnAudienceEnter, sceneId=", s.SceneId, " player=", p.SnId)
	this.BaseScenePolicy.OnAudienceEnter(s, p)
	if sceneEx, ok := s.ExtraData.(*TienLenSceneData); ok {
		//给自己发送房间信息
		p.UnmarkFlag(base.PlayerState_Leave)
		TienLenSendRoomInfo(s, p, sceneEx, nil)
		sceneEx.BroadcastAudienceNum(p)
	}
}

func (this *ScenePolicyTienLen) OnAudienceLeave(s *base.Scene, p *base.Player, reason int) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("ScenePolicyTienLen OnAudienceLeave, sceneId=", s.SceneId, " player=", p.SnId, " reason=", reason)
	if sceneEx, ok := s.ExtraData.(*TienLenSceneData); ok {
		sceneEx.BroadcastAudienceNum(p)
	}
}

func (this *ScenePolicyTienLen) OnAudienceSit(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("ScenePolicyTienLen OnAudienceSit, sceneId=", s.SceneId, " player=", p.SnId)
	sceneEx, ok := s.GetExtraData().(*TienLenSceneData)
	if !ok {
		return
	}
	if s.Gaming && !s.GetBEnterAfterStart() {
		return
	}

	//自动带入金币
	pos := sceneEx.FindOnePos()
	if pos < 0 || pos > 3 {
		p.MarkFlag(base.PlayerState_EnterSceneFailed)
		cnt := len(sceneEx.players)
		logger.Logger.Warnf("ScenePolicyTienLen.OnAudienceSit(scene:%v, player:%v) no found fit GetPos(), current player count:%v NumOfGames:%v", s.GetSceneId(), p.SnId, cnt, sceneEx.NumOfGames)
		return
	}

	playerEx := &TienLenPlayerData{Player: p}
	playerEx.init()
	playerEx.SetPos(pos)
	p.UnmarkFlag(base.PlayerState_Audience)
	p.UnmarkFlag(base.PlayerState_Leave)
	if sceneEx.GetGaming() {
		playerEx.MarkFlag(base.PlayerState_WaitNext)
		p.SyncFlag()
	}

	p.SetExtraData(playerEx)
	sceneEx.seats[pos] = playerEx
	sceneEx.players[p.SnId] = playerEx
	logger.Logger.Trace("广播个人信息,curSeatsNum: ", sceneEx.GetSeatPlayerCnt())
	if sceneEx.GetSeatPlayerCnt() == 1 {
		sceneEx.masterSnid = playerEx.SnId
		sceneEx.BroadcastUpdateMasterSnid(false)
	}

	//广播个人信息
	playerData := TienLenCreatePlayerData(p)
	pack := &tienlen.SCTienLenPlayerEnter{
		Data: playerData,
	}
	proto.SetDefaults(pack)
	s.Broadcast(int(tienlen.TienLenPacketID_PACKET_SCTienLenPlayerEnter), pack, p.GetSid())

	//给自己发送房间信息
	TienLenSendRoomInfo(s, p, sceneEx, playerEx)
	sceneEx.BroadcastAudienceNum(p)

	s.FirePlayerEvent(p, base.PlayerEventEnter, nil)

}

func (this *ScenePolicyTienLen) OnAudienceDropLine(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("ScenePolicyTienLen OnAudienceDropLine, sceneId=", s.SceneId, " player=", p.SnId)
	s.AudienceLeave(p, common.PlayerLeaveReason_DropLine)
	if sceneEx, ok := s.ExtraData.(*TienLenSceneData); ok {
		sceneEx.BroadcastAudienceNum(p)
	}
}

// 是否完成了整个牌局
func (this *ScenePolicyTienLen) IsCompleted(s *base.Scene) bool {
	if s == nil {
		return false
	}
	return false
}

// 是否可以强制开始
func (this *ScenePolicyTienLen) IsCanForceStart(s *base.Scene) bool {
	return false
}

// 当前状态能否换桌
func (this *ScenePolicyTienLen) CanChangeCoinScene(s *base.Scene, p *base.Player) bool {
	if s == nil || p == nil {
		return false
	}
	if s.GetSceneState() != nil {
		return s.GetSceneState().CanChangeCoinScene(s, p)
	}
	return true
}
func (this *ScenePolicyTienLen) ForceStart(s *base.Scene) {
	s.ChangeSceneState(rule.TienLenSceneStateWaitStart)
}
func TienLenCreatePlayerData(p *base.Player) *tienlen.TienLenPlayerData {
	pd := &tienlen.TienLenPlayerData{
		Name:        proto.String(p.Name),
		SnId:        proto.Int32(p.SnId),
		Head:        proto.Int32(p.Head),
		Sex:         proto.Int32(p.Sex),
		Coin:        proto.Int64(p.GetCoin()),
		Flag:        proto.Int(p.GetFlag()),
		Longitude:   proto.Int32(p.Longitude),
		Latitude:    proto.Int32(p.Latitude),
		City:        proto.String(p.City),
		VIP:         proto.Int32(p.VIP),
		HeadOutLine: proto.Int32(p.HeadOutLine),
		NiceId:      proto.Int32(p.NiceId),
		Pos:         proto.Int(p.GetPos()),
	}
	if p.Roles != nil {
		pd.RoleId = proto.Int32(p.Roles.ModId)
	}
	if p.Items != nil {
		pd.Items = make(map[int32]int32)
		for id, num := range p.Items {
			pd.Items[id] = proto.Int32(num)
		}
	}
	if len(p.MatchParams) > 0 {
		pd.MatchRankId = p.MatchParams[0]
	}
	if len(p.MatchParams) > 1 {
		pd.Lv = p.MatchParams[1]
	}
	logger.Logger.Trace("TienLenCreatePlayerData pd : ", pd)

	return pd
}
func TienLenCreateRoomInfoPacket(s *base.Scene, p *base.Player, sceneEx *TienLenSceneData, playerEx *TienLenPlayerData) *tienlen.SCTienLenRoomInfo {
	pack := &tienlen.SCTienLenRoomInfo{
		RoomId:       proto.Int(s.GetSceneId()),
		Creator:      proto.Int32(s.GetCreator()),
		GameId:       proto.Int(s.GetGameId()),
		RoomMode:     proto.Int(s.GetSceneMode()),
		Params:       s.GetParams(),
		State:        proto.Int32(int32(s.GetSceneState().GetState())),
		TimeOut:      proto.Int(s.GetSceneState().GetTimeout(s)),
		NumOfGames:   proto.Int(sceneEx.NumOfGames),
		TotalOfGames: proto.Int(sceneEx.TotalOfGames),
		CurOpIdx:     proto.Int(-1),
		MasterSnid:   proto.Int32(sceneEx.masterSnid),
		AudienceNum:  proto.Int(s.GetAudiencesNum()),
		BaseScore:    proto.Int32(s.BaseScore),
		MaxPlayerNum: proto.Int(s.GetPlayerNum()),
		// 比赛场相关
		Round:        proto.Int32(s.MatchRound),
		CurPlayerNum: proto.Int32(s.MatchCurPlayerNum),
		NextNeed:     proto.Int32(s.MatchNextNeed),
	}
	pack.IsMatch = int32(0)
	if s.IsMatchScene() {
		pack.IsMatch = s.MatchType
	}
	pack.MatchFinals = 0
	if s.MatchFinals {
		pack.MatchFinals = 1
		if s.NumOfGames >= 2 {
			pack.MatchFinals = 2
		}
	}
	for _, snid := range sceneEx.winSnids {
		pack.WinSnids = append(pack.WinSnids, snid)
	}
	if s.GetSceneState().GetState() == rule.TienLenSceneStateBilled {
		sceneEx.currOpPos = -1
		pack.CurOpIdx = proto.Int32(-1)
	}
	if s.GetSceneState().GetState() >= rule.TienLenSceneStatePlayerOp {
		pack.CurOpIdx = proto.Int32(sceneEx.currOpPos)
	}
	//玩家信息.第一个必然是自己
	pd := TienLenCreatePlayerData(p)
	if playerEx != nil {
		//手牌
		for i := int32(0); i < rule.HandCardNum; i++ {
			if playerEx.cards[i] != rule.InvalideCard {
				pd.Cards = append(pd.Cards, playerEx.cards[i])
			}
		}
	}
	pack.Players = append(pack.Players, pd)

	//剩下的按座位排序
	for i := 0; i < sceneEx.GetPlayerNum(); i++ {
		nowPlayer := sceneEx.seats[i]
		if nowPlayer != nil {
			if nowPlayer.SnId != p.SnId {
				pd1 := TienLenCreatePlayerData(nowPlayer.Player)
				//手牌
				for j := int32(0); j < rule.HandCardNum; j++ {
					if nowPlayer.cards[j] != rule.InvalideCard {
						if s.GetSceneState().GetState() == rule.TienLenSceneStateBilled { //结算状态显示用
							pd1.Cards = append(pd1.Cards, nowPlayer.cards[j])
						} else {
							pd1.Cards = append(pd1.Cards, rule.InvalideCard)
						}
					}
				}
				pack.Players = append(pack.Players, pd1)
			}
		}
	}
	//上两手玩家出的牌（有序）
	insertNum := 0
	for i := len(sceneEx.delOrders) - 1; i >= 0 && insertNum < 2; i-- {
		lastDelCard := &tienlen.LastDelCard{}
		delCards := sceneEx.delCards[i]
		for j := 0; j < len(delCards); j++ {
			lastDelCard.Cards = append(lastDelCard.Cards, delCards[j])
		}
		pack.LastDelCards = append(pack.LastDelCards, lastDelCard)
		insertNum++
	}
	proto.SetDefaults(pack)
	return pack
}

func TienLenSendRoomInfo(s *base.Scene, p *base.Player, sceneEx *TienLenSceneData, playerEx *TienLenPlayerData) {
	pack := TienLenCreateRoomInfoPacket(s, p, sceneEx, playerEx)
	ok := p.SendToClient(int(tienlen.TienLenPacketID_PACKET_SCTienLenRoomInfo), pack)
	logger.Logger.Trace("SCTienLenSendRoomInfo isok : ", ok, ",pack", pack)
}

// /////////////////////////////////////////////////////////////////////////////////////////////////////////////
// BaseState
// /////////////////////////////////////////////////////////////////////////////////////////////////////////////
type SceneBaseStateTienLen struct {
}

func (this *SceneBaseStateTienLen) GetTimeout(s *base.Scene) int {
	if sceneEx, ok := s.GetExtraData().(*TienLenSceneData); ok {
		return int(time.Now().Sub(sceneEx.StateStartTime) / time.Second)
	}
	return 0
}

func (this *SceneBaseStateTienLen) CanChangeTo(s base.SceneState) bool {
	return true
}

// 当前状态能否换桌
func (this *SceneBaseStateTienLen) CanChangeCoinScene(s *base.Scene, p *base.Player) bool {
	return !p.IsGameing() || s.GetDestroyed()
}

func (this *SceneBaseStateTienLen) OnEnter(s *base.Scene) {
	if sceneEx, ok := s.GetExtraData().(*TienLenSceneData); ok {
		sceneEx.StateStartTime = time.Now()
	}
}

func (this *SceneBaseStateTienLen) OnLeave(s *base.Scene) {}

func (this *SceneBaseStateTienLen) OnTick(s *base.Scene) {
	//场景状态是所有房间公用，房间的私有属性不能放到场景状态上
	if time.Now().Unix() > s.GetTimerRandomRobot() {
		s.RandRobotCnt()
		s.SetTimerRandomRobot(s.GetRobotTime())
	}

	if !s.GetGaming() {
		tNow := time.Now()
		if len(s.Players) < 2 {
			for _, p := range s.Players {
				if p.IsRob && tNow.Sub(p.GetLastOPTimer()) > time.Second*time.Duration(30+rand.Int63n(60)) {
					s.PlayerLeave(p, common.PlayerLeaveReason_Normal, true)
				}
			}
		}
		return
	}
}

// 发送玩家操作情况
func (this *SceneBaseStateTienLen) OnPlayerSToCOp(s *base.Scene, p *base.Player, pos int, opcode int, opRetCode tienlen.OpResultCode, params []int64) {
	pack := &tienlen.SCTienLenPlayerOp{
		OpCode:    proto.Int(opcode),
		SnId:      proto.Int32(p.SnId),
		OpRetCode: opRetCode,
		OpParam:   params,
	}
	proto.SetDefaults(pack)
	p.SendToClient(int(tienlen.TienLenPacketID_PACKET_SCTienLenPlayerOp), pack)
	logger.Logger.Trace("发送玩家操作情况 ", pack)
}

// 广播发送玩家操作情况
func (this *SceneBaseStateTienLen) BroadcastPlayerSToCOp(s *base.Scene, snid int32, pos int, opcode int, opRetCode tienlen.OpResultCode, params []int64) {
	pack := &tienlen.SCTienLenPlayerOp{
		OpCode:    proto.Int(opcode),
		SnId:      proto.Int32(snid),
		OpRetCode: opRetCode,
		OpParam:   params,
	}
	proto.SetDefaults(pack)
	s.Broadcast(int(tienlen.TienLenPacketID_PACKET_SCTienLenPlayerOp), pack, 0)
	if opRetCode == tienlen.OpResultCode_OPRC_Sucess {
		cards := []int{}
		for _, param := range params {
			cards = append(cards, rule.ValueStr(int32(param)))
		}
		logger.Logger.Trace("广播发送玩家操作情况 ", pack, "  snid:", snid, "  cards:", cards)
	}
}

func (this *SceneBaseStateTienLen) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	logger.Logger.Trace("SceneBaseStateTienLen.", " s.GetSceneId() : ", s.GetSceneId(), " p.SnId : ", p.SnId, " opcode : ", opcode, " params ", params)

	sceneEx, _ := s.GetExtraData().(*TienLenSceneData)
	if sceneEx == nil {
		return true
	}
	playerEx, _ := p.GetExtraData().(*TienLenPlayerData)
	if playerEx == nil {
		return true
	}
	return false
}

func (this *SceneBaseStateTienLen) OnPlayerEvent(s *base.Scene, p *base.Player, evtcode int, params []int64) {
}

func (this *SceneBaseStateTienLen) BroadcastRoomState(s *base.Scene, state int, params ...int64) {
	pack := &tienlen.SCTienLenRoomState{
		State:  proto.Int(state),
		Params: params,
	}
	proto.SetDefaults(pack)
	s.Broadcast(int(tienlen.TienLenPacketID_PACKET_SCTienLenRoomState), pack, 0)
}

// /////////////////////////////////////////////////////////////////////////////////////////////////////////////
// TienLenSceneStateWaitPlayer
// /////////////////////////////////////////////////////////////////////////////////////////////////////////////
type SceneWaitPlayerStateTienLen struct {
	SceneBaseStateTienLen
}

func (this *SceneWaitPlayerStateTienLen) GetState() int {
	return rule.TienLenSceneStateWaitPlayer
}

func (this *SceneWaitPlayerStateTienLen) CanChangeTo(s base.SceneState) bool {
	if s.GetState() == rule.TienLenSceneStateWaitStart || s.GetState() == rule.TienLenSceneStateHandCard {
		return true
	}
	return false
}

// 当前状态能否换桌
func (this *SceneWaitPlayerStateTienLen) CanChangeCoinScene(s *base.Scene, p *base.Player) bool {
	if s.IsMatchScene() {
		return false
	}
	return true
}

func (this *SceneWaitPlayerStateTienLen) OnEnter(s *base.Scene) {
	this.SceneBaseStateTienLen.OnEnter(s)

	if sceneEx, ok := s.GetExtraData().(*TienLenSceneData); ok {
		sceneEx.Clear()
		sceneEx.SetGaming(false)

		this.BroadcastRoomState(s, this.GetState())

		hasLeave := false
		//剔除下线玩家
		for i := 0; i < sceneEx.GetPlayerNum(); i++ {
			player_data := sceneEx.seats[i]
			if player_data == nil {
				continue
			}
			player_data.Clear()
			if sceneEx.IsMatchScene() {
				continue
			}
			if !player_data.IsOnLine() {
				sceneEx.PlayerLeave(player_data.Player, common.PlayerLeaveReason_DropLine, true)
				hasLeave = true
				continue
			}
			if player_data.IsRob {
				if player_data.robotGameTimes <= 0 {
					sceneEx.PlayerLeave(player_data.Player, common.PlayerLeaveReason_Normal, true)
					hasLeave = true
					continue
				}
				if s.CoinOverMaxLimit(player_data.GetCoin(), player_data.Player) {
					sceneEx.PlayerLeave(player_data.Player, common.PlayerLeaveReason_Normal, true)
					hasLeave = true
					continue
				}
			}
			if !s.CoinInLimit(player_data.GetCoin()) {
				sceneEx.PlayerLeave(player_data.Player, common.PlayerLeaveReason_Bekickout, true)
				hasLeave = true
				continue
			}
		}

		if !hasLeave && !sceneEx.IsRobFightGame() && !sceneEx.IsMatchScene() {
			s.TryDismissRob()
		}
	}
}

// 状态离开时
func (this *SceneWaitPlayerStateTienLen) OnLeave(s *base.Scene) {
	this.SceneBaseStateTienLen.OnLeave(s)
	logger.Logger.Tracef("(this *SceneWaitPlayerStateTienLen) OnLeave, sceneid=%v", s.GetSceneId())
}

// 玩家操作
func (this *SceneWaitPlayerStateTienLen) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if this.SceneBaseStateTienLen.OnPlayerOp(s, p, opcode, params) {
		return true
	}
	return true
}

// 玩家事件
func (this *SceneWaitPlayerStateTienLen) OnPlayerEvent(s *base.Scene, p *base.Player, evtcode int, params []int64) {
	logger.Logger.Trace("(this *SceneWaitPlayerStateTienLen) OnPlayerEvent, GetSceneId()=", s.GetSceneId(), " player=", p.Name, " evtcode=", evtcode)
	this.SceneBaseStateTienLen.OnPlayerEvent(s, p, evtcode, params)

	if _, ok := p.GetExtraData().(*TienLenPlayerData); ok {
		if sceneEx, ok := s.GetExtraData().(*TienLenSceneData); ok {
			switch evtcode {
			case base.PlayerEventEnter:
				//如果有人进入, 检查在线人是否能够开启游戏,够的话，切换到延迟开启状态
				if sceneEx.CanStart() {
					logger.Logger.Tracef("(this *SceneWaitPlayerStateTienLen) OnPlayerEvent s.ChangeSceneState(TienLenSceneStateWaitStart) %v", s.GetSceneId())
					s.ChangeSceneState(rule.TienLenSceneStateWaitStart)
				}
			}
		}
	}
}

func (this *SceneWaitPlayerStateTienLen) OnTick(s *base.Scene) {
	this.SceneBaseStateTienLen.OnTick(s)
	if sceneEx, ok := s.GetExtraData().(*TienLenSceneData); ok {
		if s.CheckNeedDestroy() {
			for _, p := range sceneEx.players {
				if p != nil {
					s.PlayerLeave(p.Player, common.PlayerLeaveReason_Normal, true)
				}
			}
			sceneEx.SceneDestroy(true)
			return
		}
		if sceneEx.CanStart() {
			logger.Logger.Tracef("(this *SceneWaitPlayerStateTienLen) OnTick s.ChangeSceneState(TienLenSceneStateWaitStart) %v", s.GetSceneId())
			s.ChangeSceneState(rule.TienLenSceneStateWaitStart)
		}
	}
}

// /////////////////////////////////////////////////////////////////////////////////////////////////////////////
// TienLenSceneStateWaitStart
// /////////////////////////////////////////////////////////////////////////////////////////////////////////////
type SceneWaitStartStateTienLen struct {
	SceneBaseStateTienLen
}

func (this *SceneWaitStartStateTienLen) GetState() int {
	return rule.TienLenSceneStateWaitStart
}

func (this *SceneWaitStartStateTienLen) CanChangeTo(s base.SceneState) bool {
	if s.GetState() == rule.TienLenSceneStateWaitPlayer || s.GetState() == rule.TienLenSceneStateWaitStart || s.GetState() == rule.TienLenSceneStateHandCard {
		return true
	}
	return false
}

// 当前状态能否换桌
func (this *SceneWaitStartStateTienLen) CanChangeCoinScene(s *base.Scene, p *base.Player) bool {
	if s.IsMatchScene() {
		return false
	}
	return true
}

func (this *SceneWaitStartStateTienLen) OnEnter(s *base.Scene) {
	this.SceneBaseStateTienLen.OnEnter(s)

	if sceneEx, ok := s.GetExtraData().(*TienLenSceneData); ok {
		sceneEx.Clear()
		sceneEx.SetGaming(false)
		this.BroadcastRoomState(s, this.GetState())
		logger.Logger.Trace("(this *SceneWaitStartStateTienLen) OnEnter", this.GetState())

		hasLeave := false
		//剔除下线玩家
		for i := 0; i < sceneEx.GetPlayerNum(); i++ {
			player_data := sceneEx.seats[i]
			if player_data == nil {
				continue
			}
			player_data.Clear()
			if sceneEx.IsMatchScene() {
				continue
			}
			if !player_data.IsOnLine() {
				sceneEx.PlayerLeave(player_data.Player, common.PlayerLeaveReason_DropLine, true)
				hasLeave = true
				continue
			}
			if player_data.IsRob {
				if player_data.robotGameTimes <= 0 {
					sceneEx.PlayerLeave(player_data.Player, common.PlayerLeaveReason_Normal, true)
					hasLeave = true
					continue
				}
				if s.CoinOverMaxLimit(player_data.GetCoin(), player_data.Player) {
					sceneEx.PlayerLeave(player_data.Player, common.PlayerLeaveReason_Normal, true)
					hasLeave = true
					continue
				}
			}
			if !s.CoinInLimit(player_data.GetCoin()) {
				sceneEx.PlayerLeave(player_data.Player, common.PlayerLeaveReason_Bekickout, true)
				hasLeave = true
				continue
			}
		}

		if !hasLeave && !sceneEx.IsRobFightGame() && !sceneEx.IsMatchScene() {
			s.TryDismissRob()
		}
	}
}

// 状态离开时
func (this *SceneWaitStartStateTienLen) OnLeave(s *base.Scene) {
	this.SceneBaseStateTienLen.OnLeave(s)
	logger.Logger.Tracef("(this *SceneWaitStartStateTienLen) OnLeave", this.GetState())
}

// 玩家操作
func (this *SceneWaitStartStateTienLen) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if this.SceneBaseStateTienLen.OnPlayerOp(s, p, opcode, params) {
		return true
	}
	sceneEx, _ := s.GetExtraData().(*TienLenSceneData)
	if sceneEx != nil {
		playerEx, _ := p.GetExtraData().(*TienLenPlayerData)
		if playerEx != nil {
			opRetCode := tienlen.OpResultCode_OPRC_Error
			if playerEx.SnId == sceneEx.masterSnid && this.GetState() == rule.TienLenSceneStateWaitStart {
				if sceneEx.IsMatchScene() {
					return false
				}
				switch int32(opcode) {
				case rule.TienLenPlayerOpStart: //房主开始游戏
					if sceneEx.CanStart() == false {
						s.ChangeSceneState(rule.TienLenSceneStateWaitPlayer)
					} else {
						opRetCode = tienlen.OpResultCode_OPRC_Sucess
						s.ChangeSceneState(rule.TienLenSceneStateHandCard)
					}
				}
			}
			if opRetCode == tienlen.OpResultCode_OPRC_Sucess {
				this.OnPlayerSToCOp(s, p, playerEx.GetPos(), opcode, opRetCode, params)
			} else {
				logger.Logger.Tracef("SceneWaitStartStateTienLen OnPlayerOp snid:%v, masterSnid:%v", playerEx.SnId, sceneEx.masterSnid)
			}
		}
	}
	return true
}

// 玩家事件
func (this *SceneWaitStartStateTienLen) OnPlayerEvent(s *base.Scene, p *base.Player, evtcode int, params []int64) {
	logger.Logger.Trace("(this *SceneWaitStartStateTienLen) OnPlayerEvent, GetSceneId()=", s.GetSceneId(), " player=", p.Name, " evtcode=", evtcode)
	this.SceneBaseStateTienLen.OnPlayerEvent(s, p, evtcode, params)

	if sceneEx, ok := s.GetExtraData().(*TienLenSceneData); ok {
		switch evtcode {
		case base.PlayerEventLeave:
			//如果有人退出, 检查在线人是否能够开启游戏,不够的话，切换到等待状态
			if sceneEx.CanStart() == false {
				s.ChangeSceneState(rule.TienLenSceneStateWaitPlayer)
			}
		}
	}
}

func (this *SceneWaitStartStateTienLen) OnTick(s *base.Scene) {
	this.SceneBaseStateTienLen.OnTick(s)
	if sceneEx, ok := s.GetExtraData().(*TienLenSceneData); ok {
		if s.CheckNeedDestroy() {
			for _, p := range sceneEx.players {
				if p != nil {
					s.PlayerLeave(p.Player, common.PlayerLeaveReason_Normal, true)
				}
			}
			sceneEx.SceneDestroy(true)
			return
		}
		if sceneEx.IsMatchScene() {
			delayT := time.Second * 2
			if sceneEx.MatchRound != 1 { //第一轮延迟2s，其他延迟3s 配合客户端播放动画
				delayT = time.Second * 4
			}
			if time.Now().Sub(sceneEx.StateStartTime) > delayT {
				s.ChangeSceneState(rule.TienLenSceneStateHandCard) // 比赛场直接发牌
				return
			}
		}
		if sceneEx.SceneMode == common.SceneMode_Public {
			if time.Now().Sub(sceneEx.StateStartTime) > rule.TienLenWaitStartTimeout {
				//开始前再次检查开始条件
				if sceneEx.CanStart() == true {
					s.ChangeSceneState(rule.TienLenSceneStateHandCard)
				} else {
					s.ChangeSceneState(rule.TienLenSceneStateWaitPlayer)
				}
			}
		}
	}
}

// /////////////////////////////////////////////////////////////////////////////////////////////////////////////
// TienLenSceneStateHandCard 发牌
// /////////////////////////////////////////////////////////////////////////////////////////////////////////////
type SceneHandCardStateTienLen struct {
	SceneBaseStateTienLen
}

func (this *SceneHandCardStateTienLen) GetState() int {
	return rule.TienLenSceneStateHandCard
}

func (this *SceneHandCardStateTienLen) CanChangeTo(s base.SceneState) bool {
	if s.GetState() == rule.TienLenSceneStatePlayerOp || s.GetState() == rule.TienLenSceneStateBilled {
		return true
	}
	return false
}

// 当前状态能否换桌
func (this *SceneHandCardStateTienLen) CanChangeCoinScene(s *base.Scene, p *base.Player) bool {
	return this.SceneBaseStateTienLen.CanChangeCoinScene(s, p)
}

func (this *SceneHandCardStateTienLen) OnEnter(s *base.Scene) {
	this.SceneBaseStateTienLen.OnEnter(s)

	if sceneEx, ok := s.GetExtraData().(*TienLenSceneData); ok {
		sceneEx.Clear()
		sceneEx.SetGaming(true)
		sceneEx.GameNowTime = time.Now()
		sceneEx.AllPlayerEnterGame()

		//参与游戏次数
		for i := 0; i < sceneEx.GetPlayerNum(); i++ {
			playerEx := sceneEx.seats[i]
			if playerEx != nil {
				if playerEx.IsGameing() {
					playerEx.GameTimes++
					sceneEx.curGamingPlayerNum++
					if playerEx.IsRob {
						playerEx.robotGameTimes--
					}
				}
			}
		}

		s.NumOfGames++
		s.NotifySceneRoundStart(s.NumOfGames)
		this.BroadcastRoomState(s, this.GetState(), int64(s.NumOfGames))

		//同步防伙牌数据
		sceneEx.SyncScenePlayer()
		//发牌
		sceneEx.SendHandCard()

		for _, seat := range sceneEx.seats {
			if seat != nil {
				tmpCards := seat.cards[:]
				if rule.Have2FourBomb(tmpCards) || rule.Have6StraightTwin(tmpCards) || rule.Have12Straight(tmpCards) {
					sceneEx.tianHuSnids = append(sceneEx.tianHuSnids, seat.SnId)
					sceneEx.winSnids = append(sceneEx.winSnids, seat.SnId)
				}
			}
		}
		if len(sceneEx.tianHuSnids) == 0 { //没有天胡玩家
			//有赢家，赢家先出；无赢家手持最小牌先
			pos := int32(sceneEx.FindWinPos())
			if pos == -1 {
				pos = sceneEx.startOpPos
			}
			pack := &tienlen.SCTienLenFirstOpPos{
				Pos: proto.Int32(pos),
			}
			proto.SetDefaults(pack)
			sceneEx.Broadcast(int(tienlen.TienLenPacketID_PACKET_SCTienLenFirstOpPos), pack, 0)
			logger.Logger.Trace("SCTienLenFirstOpPos: ", pack)
		}
	}
}

// 状态离开时
func (this *SceneHandCardStateTienLen) OnLeave(s *base.Scene) {
	this.SceneBaseStateTienLen.OnLeave(s)

	logger.Logger.Tracef("(this *SceneHandCardStateTienLen) OnLeave, sceneid=%v", s.GetSceneId())
}

// 玩家操作
func (this *SceneHandCardStateTienLen) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if this.SceneBaseStateTienLen.OnPlayerOp(s, p, opcode, params) {
		return true
	}

	return true
}

// 玩家事件
func (this *SceneHandCardStateTienLen) OnPlayerEvent(s *base.Scene, p *base.Player, evtcode int, params []int64) {
	logger.Logger.Trace("(this *SceneHandCardStateTienLen) OnPlayerEvent, GetSceneId()=", s.GetSceneId(), " player=", p.Name, " evtcode=", evtcode)
	this.SceneBaseStateTienLen.OnPlayerEvent(s, p, evtcode, params)
}

func (this *SceneHandCardStateTienLen) OnTick(s *base.Scene) {
	this.SceneBaseStateTienLen.OnTick(s)
	if sceneEx, ok := s.GetExtraData().(*TienLenSceneData); ok {
		newTime := rule.TienLenHandCardTimeout
		if sceneEx.isAllRob {
			newTime = time.Second * 1
		}
		if time.Now().Sub(sceneEx.StateStartTime) > newTime {
			if len(sceneEx.tianHuSnids) != 0 {
				//天胡牌型直接结算
				logger.Logger.Trace("天胡牌型直接结算:", sceneEx.tianHuSnids)
				s.ChangeSceneState(rule.TienLenSceneStateBilled)
			} else {
				//正常出牌:有赢家，赢家先出；无赢家手持最小牌先出（最小牌必出）
				winPos := sceneEx.FindWinPos()
				if winPos != -1 { //有赢家
					sceneEx.SetCurOpPos(int32(winPos))
				} else {
					sceneEx.SetCurOpPos(sceneEx.startOpPos)
				}
				logger.Logger.Trace("有赢家: ", winPos, ";谁先出：", sceneEx.GetCurOpPos())
				s.ChangeSceneState(rule.TienLenSceneStatePlayerOp)
			}
		}
	}
}

// /////////////////////////////////////////////////////////////////////////////////////////////////////////////
// TienLenSceneStatePlayerOp 出牌（玩家操作阶段）
// /////////////////////////////////////////////////////////////////////////////////////////////////////////////
type ScenePlayerOpStateTienLen struct {
	SceneBaseStateTienLen
}

func (this *ScenePlayerOpStateTienLen) GetState() int {
	return rule.TienLenSceneStatePlayerOp
}

func (this *ScenePlayerOpStateTienLen) CanChangeTo(s base.SceneState) bool {
	if s.GetState() == rule.TienLenSceneStateBilled {
		return true
	}
	return false
}

// 当前状态能否换桌
func (this *ScenePlayerOpStateTienLen) CanChangeCoinScene(s *base.Scene, p *base.Player) bool {
	return this.SceneBaseStateTienLen.CanChangeCoinScene(s, p)
}

func (this *ScenePlayerOpStateTienLen) OnEnter(s *base.Scene) {
	this.SceneBaseStateTienLen.OnEnter(s)
	this.BroadcastRoomState(s, this.GetState())

	sceneEx, _ := s.GetExtraData().(*TienLenSceneData)
	if sceneEx != nil {
		sceneEx.BroadcastOpPos()
	}
}

// 状态离开时
func (this *ScenePlayerOpStateTienLen) OnLeave(s *base.Scene) {
	this.SceneBaseStateTienLen.OnLeave(s)

	logger.Logger.Tracef("(this *SceneHandCardStateTienLen) OnLeave, sceneid=%v", s.GetSceneId())
}

// 玩家操作
func (this *ScenePlayerOpStateTienLen) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if this.SceneBaseStateTienLen.OnPlayerOp(s, p, opcode, params) {
		return true
	}
	sceneEx, _ := s.GetExtraData().(*TienLenSceneData)
	if sceneEx != nil {
		playerEx, _ := p.GetExtraData().(*TienLenPlayerData)
		if playerEx != nil {
			if sceneEx.GetCurOpPos() != int32(playerEx.GetPos()) {
				return false
			}
			opRetCode := tienlen.OpResultCode_OPRC_Error
			switch int32(opcode) {
			case rule.TienLenPlayerOpPlay: //出牌
				delCards := []int32{}
				for _, card := range params {
					isHave := false
					for _, hcard := range playerEx.cards { //去手牌里找找看有没有
						if int32(card) == hcard && hcard != rule.InvalideCard {
							isHave = true
						}
					}
					if isHave {
						delCards = append(delCards, int32(card))
					} else {
						logger.Logger.Trace("ScenePlayerOpStateTienLen,1params:", params)
						opRetCode = tienlen.OpResultCode_OPRC_Error
						break
					}
				}
				if len(delCards) == len(params) && len(delCards) > 0 {
					isRule, _ := rule.RulePopEnable(delCards)
					if sceneEx.IsTienLenYule() {
						ruleType := rule.Tienlen_Pass
						isRule, ruleType = rule.RulePopEnable_yl(delCards)
						if ruleType == rule.Plane_Single { //飞机带单只能最后一手出
							haveCardNum := 0
							for _, hcard := range playerEx.cards {
								if hcard != rule.InvalideCard {
									haveCardNum++
								}
							}
							if len(delCards) != haveCardNum {
								isRule = false
								opRetCode = tienlen.OpResultCode_OPRC_Error
								logger.Logger.Trace("飞机带单只能最后一手出, delCards:", delCards, " haveCardNum:", haveCardNum)
							}
						}
					}
					logger.Logger.Trace("ScenePlayerOpStateTienLen,2params:", params, " isRule:", isRule)
					if isRule { //符合出牌规则
						if int32(playerEx.GetPos()) == sceneEx.lastOpPos || sceneEx.lastOpPos == rule.InvalidePos { //当前操作者和上一个操作者是同一个人，必出牌
							if sceneEx.lastOpPos == rule.InvalidePos { //首出玩家
								//有赢家，赢家先出，出牌不受限制
								//无赢家，手持最小牌先出，最小牌必先出
								winPos := sceneEx.FindWinPos()
								logger.Logger.Trace("ScenePlayerOpStateTienLen,8params:", params, " winPos:", winPos)
								if winPos == -1 { //无赢家
									haveMinCard := false
									for _, card := range delCards {
										if card == sceneEx.curMinCard { //最小牌必先出
											haveMinCard = true
										}
									}
									logger.Logger.Trace("ScenePlayerOpStateTienLen,9params:", params, " curMinCard:", sceneEx.curMinCard, " haveMinCard", haveMinCard)
									if haveMinCard {
										isDel := sceneEx.DelCards(playerEx, delCards)
										logger.Logger.Trace("ScenePlayerOpStateTienLen,3params:", params, " isDel:", isDel)
										if isDel {
											sceneEx.DoNext(int32(playerEx.GetPos()))
											opRetCode = tienlen.OpResultCode_OPRC_Sucess
											sceneEx.SetLastOpPos(int32(playerEx.GetPos()))
										}
									}
								} else {
									isDel := sceneEx.DelCards(playerEx, delCards)
									logger.Logger.Trace("ScenePlayerOpStateTienLen,10params:", params, " isDel:", isDel)
									if isDel {
										sceneEx.DoNext(int32(playerEx.GetPos()))
										opRetCode = tienlen.OpResultCode_OPRC_Sucess
										sceneEx.SetLastOpPos(int32(playerEx.GetPos()))
									}
								}
							} else {
								isDel := sceneEx.DelCards(playerEx, delCards)
								logger.Logger.Trace("ScenePlayerOpStateTienLen,4params:", params, " isDel:", isDel)
								if isDel {
									nextPos := sceneEx.DoNext(int32(playerEx.GetPos()))
									logger.Logger.Trace("ScenePlayerOpStateTienLen,4paramssss:", params, " nextPos:", nextPos)
									if sceneEx.IsTienLenToEnd() && nextPos == rule.InvalidePos {
										sceneEx.UnmarkPass()
										nextPos = sceneEx.DoNext(int32(playerEx.GetPos()))
										logger.Logger.Trace("ScenePlayerOpStateTienLen,4paramssss:", params, " nextPos:", nextPos)
									}
									opRetCode = tienlen.OpResultCode_OPRC_Sucess
									sceneEx.SetLastOpPos(int32(playerEx.GetPos()))
								}
							}
							if opRetCode == tienlen.OpResultCode_OPRC_Sucess {
								isBomb := rule.IsFourBomb(delCards)
								if isBomb {
									sceneEx.isKongBomb = true
								}
							}
							sceneEx.UnmarkPass()
						} else { //当前操作者和上一个操作者不是同一个人，必压制
							if !playerEx.isPass {
								lastOpPlayer := sceneEx.GetLastOpPlayer()
								logger.Logger.Trace("ScenePlayerOpStateTienLen,5params:", params, " lastOpPlayer:", lastOpPlayer)
								if lastOpPlayer != nil && len(lastOpPlayer.delCards) != 0 {
									lastDelCards := lastOpPlayer.delCards[len(lastOpPlayer.delCards)-1]
									canDel, isBomb, bombScore := rule.CanDel(lastDelCards, delCards, sceneEx.IsTienLenToEnd())
									if sceneEx.IsTienLenYule() {
										canDel, isBomb, bombScore = rule.CanDel_yl(lastDelCards, delCards, sceneEx.IsTienLenToEnd())
									}
									logger.Logger.Trace("ScenePlayerOpStateTienLen,6params:", params, " canDel:", canDel, " lastDelCards:", lastDelCards)
									if canDel {
										if isBomb {
											sceneEx.curBombPos = int32(playerEx.GetPos())
											sceneEx.lastBombPos = sceneEx.lastOpPos
											sceneEx.roundScore += bombScore
										} else {
											sceneEx.curBombPos = rule.InvalidePos
											sceneEx.lastBombPos = rule.InvalidePos
											sceneEx.roundScore = 0
										}
										isDel := sceneEx.DelCards(playerEx, delCards)
										logger.Logger.Trace("ScenePlayerOpStateTienLen,7params:", params, " isDel:", isDel)
										if isDel {
											nextPos := sceneEx.DoNext(int32(playerEx.GetPos()))
											if sceneEx.IsTienLenToEnd() && nextPos == rule.InvalidePos {
												sceneEx.DoNext(int32(sceneEx.lastOpPos))
											}
											sceneEx.SetLastOpPos(int32(playerEx.GetPos()))
											opRetCode = tienlen.OpResultCode_OPRC_Sucess
											sceneEx.isKongBomb = false
										}
									}
								}
							}
						}
					}
				}
			case rule.TienLenPlayerOpPass: //过牌
				if int32(playerEx.GetPos()) == sceneEx.lastOpPos { //当前操作者和上一个操作者是同一个人，必出牌，不能过牌
					opRetCode = tienlen.OpResultCode_OPRC_Error
				} else {
					if sceneEx.lastOpPos != rule.InvalidePos {
						if sceneEx.lastOpPos != int32(playerEx.GetPos()) {
							logger.Logger.Info("***************sceneEx.lastOpPos != playerEx.GetPos()", sceneEx.lastOpPos, playerEx.GetPos())
						}
						sceneEx.card_play_action_seq = append(sceneEx.card_play_action_seq, fmt.Sprintf("%v-过", playerEx.GetPos()))
						sceneEx.card_play_action_seq_int32 = append(sceneEx.card_play_action_seq_int32, []int32{-1})
						nextPos := sceneEx.DoNext(int32(playerEx.GetPos()))

						if sceneEx.IsTienLenToEnd() && nextPos == rule.InvalidePos {
							sceneEx.UnmarkPass()
							nextPos = sceneEx.DoNext(int32(sceneEx.lastOpPos))
							sceneEx.SetLastOpPos(int32(nextPos))
						}
						opRetCode = tienlen.OpResultCode_OPRC_Sucess
						playerEx.isPass = true
						if nextPos == sceneEx.lastOpPos { //一轮都不出牌
							sceneEx.TrySmallGameBilled()
						}
					}
				}
			default:
				opRetCode = tienlen.OpResultCode_OPRC_Error
			}
			//next
			if opRetCode == tienlen.OpResultCode_OPRC_Sucess {
				this.BroadcastPlayerSToCOp(s, playerEx.SnId, playerEx.GetPos(), opcode, opRetCode, params)

				delCardNum := 0
				for _, hcard := range playerEx.cards {
					if hcard == rule.InvalideCard {
						delCardNum++
					}
				}
				if delCardNum == rule.Hand_CardNum { //牌出完了
					sceneEx.TrySmallGameBilled()
					sceneEx.winSnids = append(sceneEx.winSnids, playerEx.SnId)
					if sceneEx.IsTienLenToEnd() { //打到底
						if sceneEx.GetGameingPlayerCnt()-1 == len(sceneEx.winSnids) {
							sceneEx.ChangeSceneState(rule.TienLenSceneStateBilled)
						} else {
							playerEx.isDelAll = true
							sceneEx.BroadcastOpPos()
						}
					} else {
						sceneEx.ChangeSceneState(rule.TienLenSceneStateBilled)
					}
				} else {
					sceneEx.BroadcastOpPos()
				}
			} else {
				this.OnPlayerSToCOp(s, p, playerEx.GetPos(), opcode, opRetCode, params)
			}
		}
	}

	return true
}

// 玩家事件
func (this *ScenePlayerOpStateTienLen) OnPlayerEvent(s *base.Scene, p *base.Player, evtcode int, params []int64) {
	logger.Logger.Trace("(this *SceneHandCardStateTienLen) OnPlayerEvent, GetSceneId()=", s.GetSceneId(), " player=", p.Name, " evtcode=", evtcode)
	this.SceneBaseStateTienLen.OnPlayerEvent(s, p, evtcode, params)
}

func (this *ScenePlayerOpStateTienLen) OnTick(s *base.Scene) {
	this.SceneBaseStateTienLen.OnTick(s)
	if sceneEx, ok := s.GetExtraData().(*TienLenSceneData); ok {
		if time.Now().Sub(sceneEx.StateStartTime) > rule.TienLenPlayerOpTimeout {
			//超时当前玩家
			curOpPos := sceneEx.GetCurOpPos()
			if curOpPos != rule.InvalidePos {
				playerEx := sceneEx.seats[curOpPos]
				if playerEx != nil {
					//超时后单轮首出玩家出最小的牌，非单轮首出玩家不出牌
					if sceneEx.lastOpPos == curOpPos || sceneEx.lastOpPos == rule.InvalidePos { //单轮首出玩家
						delCards := []int32{}
						//排序
						cpCards := playerEx.cards[:]
						sort.Slice(cpCards, func(i, j int) bool {
							v_i := rule.Value(cpCards[i])
							v_j := rule.Value(cpCards[j])
							c_i := rule.Color(cpCards[i])
							c_j := rule.Color(cpCards[j])
							if v_i > v_j {
								return false
							} else if v_i == v_j {
								return c_i < c_j
							}
							return true
						})
						for _, card := range cpCards {
							if card != rule.InvalideCard {
								delCards = append(delCards, card)
								break
							}
						}
						isDel := sceneEx.DelCards(playerEx, delCards)
						if isDel {
							sceneEx.DoNext(int32(playerEx.GetPos()))
							params := []int64{}
							for _, card := range delCards {
								params = append(params, int64(card))
							}
							sceneEx.UnmarkPass()
							this.BroadcastPlayerSToCOp(s, playerEx.SnId, playerEx.GetPos(), int(rule.TienLenPlayerOpPlay), tienlen.OpResultCode_OPRC_Sucess, params)
							sceneEx.SetLastOpPos(int32(playerEx.GetPos()))
							sceneEx.BroadcastOpPos()
						} else {
							sceneEx.DoNext(int32(playerEx.GetPos()))
							playerEx.isPass = true
							this.BroadcastPlayerSToCOp(s, playerEx.SnId, playerEx.GetPos(), int(rule.TienLenPlayerOpPass), tienlen.OpResultCode_OPRC_Sucess, []int64{})
							sceneEx.BroadcastOpPos()
						}
						sceneEx.isKongBomb = false
						delAll := 0
						for _, card := range playerEx.cards {
							if card == rule.InvalideCard {
								delAll++
							}
						}
						if delAll == rule.Hand_CardNum {
							sceneEx.TrySmallGameBilled()
							sceneEx.winSnids = append(sceneEx.winSnids, playerEx.SnId)
							if sceneEx.IsTienLenToEnd() { //打到底
								if sceneEx.GetGameingPlayerCnt()-1 == len(sceneEx.winSnids) {
									sceneEx.ChangeSceneState(rule.TienLenSceneStateBilled)
								} else {
									playerEx.isDelAll = true
									sceneEx.BroadcastOpPos()
								}
							} else {
								sceneEx.ChangeSceneState(rule.TienLenSceneStateBilled)
							}
						}
					} else {
						sceneEx.card_play_action_seq = append(sceneEx.card_play_action_seq, fmt.Sprintf("%v-过", playerEx.GetPos()))
						sceneEx.card_play_action_seq_int32 = append(sceneEx.card_play_action_seq_int32, []int32{-1})
						nextPos := sceneEx.DoNext(int32(playerEx.GetPos()))
						if sceneEx.IsTienLenToEnd() && nextPos == rule.InvalidePos {
							sceneEx.UnmarkPass()
							nextPos = sceneEx.DoNext(int32(sceneEx.lastOpPos))
							sceneEx.SetLastOpPos(int32(nextPos))
						}
						playerEx.isPass = true
						if nextPos == sceneEx.lastOpPos { //一轮都不出牌
							sceneEx.TrySmallGameBilled()
						}
						this.BroadcastPlayerSToCOp(s, playerEx.SnId, playerEx.GetPos(), int(rule.TienLenPlayerOpPass), tienlen.OpResultCode_OPRC_Sucess, []int64{})
						sceneEx.BroadcastOpPos()
					}
				}
			}
			sceneEx.StateStartTime = time.Now()
		}
	}
}

// /////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Billed
// /////////////////////////////////////////////////////////////////////////////////////////////////////////////
type SceneBilledStateTienLen struct {
	SceneBaseStateTienLen
}

func (this *SceneBilledStateTienLen) GetState() int {
	return rule.TienLenSceneStateBilled
}

func (this *SceneBilledStateTienLen) CanChangeTo(s base.SceneState) bool {
	if s.GetState() == rule.TienLenSceneStateWaitPlayer || s.GetState() == rule.TienLenSceneStateWaitStart {
		return true
	}
	return false
}

// 当前状态能否换桌
func (this *SceneBilledStateTienLen) CanChangeCoinScene(s *base.Scene, p *base.Player) bool {
	return this.SceneBaseStateTienLen.CanChangeCoinScene(s, p)
}

func (this *SceneBilledStateTienLen) OnEnter(s *base.Scene) {
	this.SceneBaseStateTienLen.OnEnter(s)
	this.BroadcastRoomState(s, this.GetState())

	//在这里执行结算
	if sceneEx, ok := s.GetExtraData().(*TienLenSceneData); ok {
		winScore := int64(0)
		pack := &tienlen.SCTienLenGameBilled{}
		tienlenType := model.TienLenType{
			GameId:      sceneEx.GameId,
			RoomId:      int32(sceneEx.GetSceneId()),
			RoomType:    sceneEx.GetFreeGameSceneType(),
			NumOfGames:  int32(sceneEx.Scene.NumOfGames),
			BankId:      sceneEx.masterSnid,
			PlayerCount: sceneEx.curGamingPlayerNum,
			BaseScore:   s.BaseScore,
			TaxRate:     s.DbGameFree.GetTaxRate(),
			RoomMode:    s.GetSceneMode(),
		}

		nGamingPlayerCount := sceneEx.GetGameingPlayerCnt()
		if len(sceneEx.tianHuSnids) == nGamingPlayerCount { //和
			for i := 0; i < sceneEx.GetPlayerNum(); i++ {
				playerEx := sceneEx.seats[i]
				if playerEx == nil {
					continue
				}
				if !playerEx.IsGameing() {
					continue
				}
				billData := &tienlen.TienLenPlayerGameBilled{
					SnId:     proto.Int32(playerEx.SnId),
					IsWin:    proto.Int32(0),
					WinCoin:  proto.Int64(0),
					GameCoin: proto.Int64(playerEx.GetCoin()),
				}
				playerEx.CurIsWin = int64(0)
				tienlenPerson := model.TienLenPerson{
					UserId:      playerEx.SnId,
					UserIcon:    playerEx.Head,
					Platform:    playerEx.Platform,
					Channel:     playerEx.Channel,
					Promoter:    playerEx.BeUnderAgentCode,
					PackageTag:  playerEx.PackageID,
					InviterId:   playerEx.InviterId,
					WBLevel:     playerEx.WBLevel,
					IsRob:       playerEx.IsRob,
					IsFirst:     sceneEx.IsPlayerFirst(sceneEx.GetPlayer(playerEx.SnId)),
					IsLeave:     false,
					IsWin:       0,
					GainCoin:    0,
					BombCoin:    0,
					BillCoin:    0,
					GainTaxCoin: 0,
					BombTaxCoin: 0,
					BillTaxCoin: 0,
					Seat:        playerEx.GetPos(),
					IsTianHu:    sceneEx.IsTianhuPlayer(playerEx.SnId),
				}
				for _, card := range playerEx.cards {
					if card != rule.InvalideCard {
						billData.Cards = append(billData.Cards, card)
						tienlenPerson.CardInfoEnd = append(tienlenPerson.CardInfoEnd, card)
					}
				}
				//排下序，正常应该客户端排序
				sort.Slice(billData.Cards, func(i, j int) bool {
					v_i := rule.Value(int32(billData.Cards[i]))
					v_j := rule.Value(int32(billData.Cards[j]))
					c_i := rule.Color(int32(billData.Cards[i]))
					c_j := rule.Color(int32(billData.Cards[j]))
					if v_i > v_j {
						return false
					} else if v_i == v_j {
						return c_i < c_j
					}
					return true
				})
				pack.Datas = append(pack.Datas, billData)
				playerEx.Statics(sceneEx.KeyGameId, sceneEx.KeyGamefreeId, 0, true)
				sceneEx.TryBillExGameDrop(playerEx.Player) //和
				tienlenType.PlayerData = append(tienlenType.PlayerData, tienlenPerson)
			}
		} else { //输赢局
			if sceneEx.IsTienLenToEnd() && len(sceneEx.tianHuSnids) == 0 {
				lastSnid := int32(0)
				playerNum := 0
				for _, p := range sceneEx.players {
					if p != nil && p.IsGameing() {
						playerNum++
						if !common.InSliceInt32(sceneEx.winSnids, p.SnId) {
							lastSnid = p.SnId
						}
					}
				}
				if lastSnid == 0 {
					logger.Logger.Error("TienLenToEndGameBilled Error: lastSnid == 0")
					return
				}
				if playerNum-1 != len(sceneEx.winSnids) {
					logger.Logger.Error("TienLenToEndGameBilled Error: playerNum: ", playerNum, " this.winSnids： ", sceneEx.winSnids)
					return
				}
				// 没出完牌的输家
				losePlayerScore := int64(0)
				losePlayer := sceneEx.players[lastSnid]
				if losePlayer != nil {
					playerLoseScore := rule.GetLoseScore(losePlayer.cards, sceneEx.IsTienLenToEnd())
					score := int64(s.BaseScore) * (int64(playerLoseScore) + 100) / 100 //手牌输分和基础
					gainScore := score
					if sceneEx.IsTienLenYule() && sceneEx.bombToEnd > 0 { //娱乐版空放炸弹底分翻倍
						logger.Logger.Trace("娱乐版空放炸弹底分翻倍,bombToEnd: ", sceneEx.bombToEnd, " SnId: ", losePlayer.SnId)
						for bomb := 0; bomb < sceneEx.bombToEnd; bomb++ {
							gainScore *= 2
						}
					}
					losePlayerCoin := losePlayer.GetCoin()
					if !sceneEx.IsMatchScene() && losePlayerCoin < gainScore {
						gainScore = losePlayerCoin
					}
					losePlayerScore = gainScore
					if sceneEx.IsMatchScene() { //比赛场是积分，不应该增加账变
						losePlayer.AddCoinNoLog(int64(-gainScore), 0)
					} else {
						losePlayer.AddCoin(int64(-gainScore), common.GainWay_CoinSceneLost, 0, "system", s.GetSceneName())
					}
					losePlayer.winCoin -= gainScore
					billData := &tienlen.TienLenPlayerGameBilled{
						SnId:     proto.Int32(losePlayer.SnId),
						IsWin:    proto.Int32(2),
						WinCoin:  proto.Int64(gainScore),
						GameCoin: proto.Int64(losePlayer.GetCoin()),
					}
					isWin := int32(0)
					billCoin := losePlayer.bombScore - gainScore
					if billCoin > 0 {
						isWin = 1
					} else if billCoin < 0 {
						isWin = -1
					}
					losePlayer.CurIsWin = int64(isWin)
					tienlenPerson := model.TienLenPerson{
						UserId:      losePlayer.SnId,
						UserIcon:    losePlayer.Head,
						Platform:    losePlayer.Platform,
						Channel:     losePlayer.Channel,
						Promoter:    losePlayer.BeUnderAgentCode,
						PackageTag:  losePlayer.PackageID,
						InviterId:   losePlayer.InviterId,
						WBLevel:     losePlayer.WBLevel,
						IsRob:       losePlayer.IsRob,
						IsFirst:     sceneEx.IsPlayerFirst(sceneEx.GetPlayer(losePlayer.SnId)),
						IsLeave:     false,
						IsWin:       isWin,
						GainCoin:    -gainScore,
						BombCoin:    losePlayer.bombScore,
						BillCoin:    billCoin,
						GainTaxCoin: 0,
						BombTaxCoin: losePlayer.bombTaxScore,
						BillTaxCoin: losePlayer.bombTaxScore,
						Seat:        losePlayer.GetPos(),
						IsTianHu:    false,
					}
					tienlenPerson.DelOrderCards = make(map[int][]int32, len(sceneEx.delOrders))
					for i2, orderSnid := range sceneEx.delOrders {
						if orderSnid == losePlayer.SnId {
							tienlenPerson.DelOrderCards[i2] = sceneEx.delCards[i2]
						}
					}
					for _, card := range losePlayer.cards {
						if card != rule.InvalideCard {
							billData.Cards = append(billData.Cards, card)
							tienlenPerson.CardInfoEnd = append(tienlenPerson.CardInfoEnd, card)
						}
					}
					//排下序，正常应该客户端排序
					sort.Slice(billData.Cards, func(i, j int) bool {
						v_i := rule.Value(int32(billData.Cards[i]))
						v_j := rule.Value(int32(billData.Cards[j]))
						c_i := rule.Color(int32(billData.Cards[i]))
						c_j := rule.Color(int32(billData.Cards[j]))
						if v_i > v_j {
							return false
						} else if v_i == v_j {
							return c_i < c_j
						}
						return true
					})
					pack.Datas = append(pack.Datas, billData)
					losePlayer.Statics(sceneEx.KeyGameId, sceneEx.KeyGamefreeId, billCoin, true)
					sceneEx.TryBillExGameDrop(losePlayer.Player) //输家
					tienlenType.PlayerData = append(tienlenType.PlayerData, tienlenPerson)

					lastWinPlayerScore := int64(0)
					if playerNum == 3 || playerNum == 4 {
						// 最后一位出完牌的输家
						lastWinPlayer := sceneEx.players[sceneEx.winSnids[len(sceneEx.winSnids)-1]]
						if lastWinPlayer != nil {
							lastWinScore := rule.GetLoseScore(lastWinPlayer.cards, sceneEx.IsTienLenToEnd())
							lastWinscore := int64(s.BaseScore) * (int64(lastWinScore) + 50) / 100 //手牌输分和基础
							astWinGainScore := lastWinscore
							if sceneEx.IsTienLenYule() && sceneEx.bombToEnd > 0 { //娱乐版空放炸弹底分翻倍
								logger.Logger.Trace("娱乐版空放炸弹底分翻倍,bombToEnd: ", sceneEx.bombToEnd, " SnId: ", lastWinPlayer.SnId)
								for bomb := 0; bomb < sceneEx.bombToEnd; bomb++ {
									astWinGainScore *= 2
								}
							}
							lastWinPlayerCoin := lastWinPlayer.GetCoin()
							if !sceneEx.IsMatchScene() && lastWinPlayerCoin < astWinGainScore {
								astWinGainScore = lastWinPlayerCoin
							}
							lastWinPlayerScore = astWinGainScore
							if sceneEx.IsMatchScene() {
								lastWinPlayer.AddCoinNoLog(int64(-astWinGainScore), 0)
							} else {
								lastWinPlayer.AddCoin(int64(-astWinGainScore), common.GainWay_CoinSceneLost, 0, "system", s.GetSceneName())
							}
							lastWinPlayer.winCoin -= astWinGainScore
							billData := &tienlen.TienLenPlayerGameBilled{
								SnId:     proto.Int32(lastWinPlayer.SnId),
								IsWin:    proto.Int32(2),
								WinCoin:  proto.Int64(astWinGainScore),
								GameCoin: proto.Int64(lastWinPlayer.GetCoin()),
							}
							isWin := int32(0)
							billCoin := lastWinPlayer.bombScore - astWinGainScore
							if billCoin > 0 {
								isWin = 1
							} else if billCoin < 0 {
								isWin = -1
							}
							lastWinPlayer.CurIsWin = int64(isWin)
							tienlenPerson := model.TienLenPerson{
								UserId:      lastWinPlayer.SnId,
								UserIcon:    lastWinPlayer.Head,
								Platform:    lastWinPlayer.Platform,
								Channel:     lastWinPlayer.Channel,
								Promoter:    lastWinPlayer.BeUnderAgentCode,
								PackageTag:  lastWinPlayer.PackageID,
								InviterId:   lastWinPlayer.InviterId,
								WBLevel:     lastWinPlayer.WBLevel,
								IsRob:       lastWinPlayer.IsRob,
								IsFirst:     sceneEx.IsPlayerFirst(sceneEx.GetPlayer(lastWinPlayer.SnId)),
								IsLeave:     false,
								IsWin:       isWin,
								GainCoin:    -astWinGainScore,
								BombCoin:    lastWinPlayer.bombScore,
								BillCoin:    billCoin,
								GainTaxCoin: 0,
								BombTaxCoin: lastWinPlayer.bombTaxScore,
								BillTaxCoin: lastWinPlayer.bombTaxScore,
								Seat:        lastWinPlayer.GetPos(),
								IsTianHu:    false,
							}
							tienlenPerson.DelOrderCards = make(map[int][]int32, len(sceneEx.delOrders))
							for i2, orderSnid := range sceneEx.delOrders {
								if orderSnid == lastWinPlayer.SnId {
									tienlenPerson.DelOrderCards[i2] = sceneEx.delCards[i2]
								}
							}
							pack.Datas = append(pack.Datas, billData)
							lastWinPlayer.Statics(sceneEx.KeyGameId, sceneEx.KeyGamefreeId, billCoin, true)
							sceneEx.TryBillExGameDrop(lastWinPlayer.Player) //输家
							tienlenType.PlayerData = append(tienlenType.PlayerData, tienlenPerson)
						}
					}

					//first
					playerEx := sceneEx.players[sceneEx.winSnids[0]]
					if playerEx != nil {
						taxRate := sceneEx.DbGameFree.GetTaxRate()                                      //万分比
						gainScore := int64(float64(losePlayerScore) * float64(10000-taxRate) / 10000.0) //税后
						gainTaxScore := losePlayerScore - gainScore
						if playerNum == 3 {
							gainScore = int64(float64(losePlayerScore+lastWinPlayerScore) * float64(10000-taxRate) / 10000.0) //税后
							gainTaxScore = losePlayerScore + lastWinPlayerScore - gainScore
						}
						if sceneEx.IsMatchScene() {
							playerEx.AddCoinNoLog(int64(gainScore), 0)
						} else {
							playerEx.AddCoin(gainScore, common.GainWay_CoinSceneWin, 0, "system", s.GetSceneName())
						}
						playerEx.winCoin += gainScore
						billData := &tienlen.TienLenPlayerGameBilled{
							SnId:     proto.Int32(playerEx.SnId),
							IsWin:    proto.Int32(1),
							WinCoin:  proto.Int64(gainScore),
							GameCoin: proto.Int64(playerEx.GetCoin()),
						}
						isWin := int32(0)
						billCoin := playerEx.bombScore + gainScore
						if billCoin > 0 {
							isWin = 1
						} else if billCoin < 0 {
							isWin = -1
						}
						playerEx.CurIsWin = int64(isWin)
						tienlenPerson := model.TienLenPerson{
							UserId:      playerEx.SnId,
							UserIcon:    playerEx.Head,
							Platform:    playerEx.Platform,
							Channel:     playerEx.Channel,
							Promoter:    playerEx.BeUnderAgentCode,
							PackageTag:  playerEx.PackageID,
							InviterId:   playerEx.InviterId,
							WBLevel:     playerEx.WBLevel,
							IsRob:       playerEx.IsRob,
							IsFirst:     sceneEx.IsPlayerFirst(sceneEx.GetPlayer(playerEx.SnId)),
							IsLeave:     false,
							IsWin:       isWin,
							GainCoin:    gainScore,
							BombCoin:    playerEx.bombScore,
							BillCoin:    billCoin,
							GainTaxCoin: gainTaxScore,
							BombTaxCoin: playerEx.bombTaxScore,
							BillTaxCoin: playerEx.bombTaxScore + gainTaxScore,
							Seat:        playerEx.GetPos(),
							IsTianHu:    sceneEx.IsTianhuPlayer(playerEx.SnId),
						}
						tienlenPerson.DelOrderCards = make(map[int][]int32, len(sceneEx.delOrders))
						for i2, orderSnid := range sceneEx.delOrders {
							if orderSnid == playerEx.SnId {
								tienlenPerson.DelOrderCards[i2] = sceneEx.delCards[i2]
							}
						}
						pack.Datas = append(pack.Datas, billData)
						playerEx.Statics(sceneEx.KeyGameId, sceneEx.KeyGamefreeId, billCoin, true)
						sceneEx.TryBillExGameDrop(playerEx.Player) //赢家
						tienlenType.PlayerData = append(tienlenType.PlayerData, tienlenPerson)
					}
					//second
					if playerNum == 4 {
						playerEx = sceneEx.players[sceneEx.winSnids[1]]
						if playerEx != nil {
							taxRate := sceneEx.DbGameFree.GetTaxRate()                                         //万分比
							gainScore := int64(float64(lastWinPlayerScore) * float64(10000-taxRate) / 10000.0) //税后
							gainTaxScore := winScore - gainScore
							if sceneEx.IsMatchScene() {
								playerEx.AddCoinNoLog(int64(gainScore), 0)
							} else {
								playerEx.AddCoin(gainScore, common.GainWay_CoinSceneWin, 0, "system", s.GetSceneName())
							}
							playerEx.winCoin += gainScore
							billData := &tienlen.TienLenPlayerGameBilled{
								SnId:     proto.Int32(playerEx.SnId),
								IsWin:    proto.Int32(1),
								WinCoin:  proto.Int64(gainScore),
								GameCoin: proto.Int64(playerEx.GetCoin()),
							}
							isWin := int32(0)
							billCoin := playerEx.bombScore + gainScore
							if billCoin > 0 {
								isWin = 1
							} else if billCoin < 0 {
								isWin = -1
							}
							playerEx.CurIsWin = int64(isWin)
							tienlenPerson := model.TienLenPerson{
								UserId:      playerEx.SnId,
								UserIcon:    playerEx.Head,
								Platform:    playerEx.Platform,
								Channel:     playerEx.Channel,
								Promoter:    playerEx.BeUnderAgentCode,
								PackageTag:  playerEx.PackageID,
								InviterId:   playerEx.InviterId,
								WBLevel:     playerEx.WBLevel,
								IsRob:       playerEx.IsRob,
								IsFirst:     sceneEx.IsPlayerFirst(sceneEx.GetPlayer(playerEx.SnId)),
								IsLeave:     false,
								IsWin:       isWin,
								GainCoin:    gainScore,
								BombCoin:    playerEx.bombScore,
								BillCoin:    billCoin,
								GainTaxCoin: gainTaxScore,
								BombTaxCoin: playerEx.bombTaxScore,
								BillTaxCoin: playerEx.bombTaxScore + gainTaxScore,
								Seat:        playerEx.GetPos(),
								IsTianHu:    sceneEx.IsTianhuPlayer(playerEx.SnId),
							}
							tienlenPerson.DelOrderCards = make(map[int][]int32, len(sceneEx.delOrders))
							for i2, orderSnid := range sceneEx.delOrders {
								if orderSnid == playerEx.SnId {
									tienlenPerson.DelOrderCards[i2] = sceneEx.delCards[i2]
								}
							}
							pack.Datas = append(pack.Datas, billData)
							playerEx.Statics(sceneEx.KeyGameId, sceneEx.KeyGamefreeId, billCoin, true)
							sceneEx.TryBillExGameDrop(playerEx.Player) //赢家
							tienlenType.PlayerData = append(tienlenType.PlayerData, tienlenPerson)
						}
					}
				}
			} else {
				// 输家
				for i := 0; i < sceneEx.GetPlayerNum(); i++ {
					playerEx := sceneEx.seats[i]
					if playerEx == nil {
						continue
					}
					if !playerEx.IsGameing() {
						continue
					}
					if sceneEx.IsTianhuPlayer(playerEx.SnId) {
						continue
					}
					if sceneEx.IsWinPlayer(playerEx.SnId) && !sceneEx.IsTienLenToEnd() {
						continue
					}
					logger.Logger.Trace("SceneBilledStateTienLe,losePos: ", i, " SnId: ", playerEx.SnId, " BaseScore: ", s.BaseScore)
					playerLoseScore := rule.GetLoseScore(playerEx.cards, sceneEx.IsTienLenToEnd())
					score := int64(s.BaseScore) * int64(playerLoseScore)
					gainScore := int64(0)
					if len(sceneEx.tianHuSnids) != 0 { //天胡结算翻倍
						gainScore = int64(score) * int64(len(sceneEx.tianHuSnids)) * 2
						if sceneEx.IsTienLenToEnd() { //打到底天胡结算加上2倍底注（不是翻倍）
							gainScore = (int64(score) + int64(s.BaseScore)*int64(rule.Score2End10)*2) / 100 //特殊牌型加上2倍底注
						}
					} else { //正常结算
						gainScore = int64(score)
						if sceneEx.IsTienLenYule() && sceneEx.bombToEnd > 0 { //娱乐版空放炸弹底分翻倍
							logger.Logger.Trace("娱乐版空放炸弹底分翻倍,bombToEnd: ", sceneEx.bombToEnd, " SnId: ", playerEx.SnId)
							for bomb := 0; bomb < sceneEx.bombToEnd; bomb++ {
								gainScore *= 2
							}
						}
					}
					losePlayerCoin := playerEx.GetCoin()
					if !sceneEx.IsMatchScene() && losePlayerCoin < gainScore {
						gainScore = losePlayerCoin
					}
					winScore += gainScore
					if sceneEx.IsMatchScene() {
						playerEx.AddCoinNoLog(int64(-gainScore), 0)
					} else {
						playerEx.AddCoin(int64(-gainScore), common.GainWay_CoinSceneLost, 0, "system", s.GetSceneName())
					}
					playerEx.winCoin -= gainScore
					billData := &tienlen.TienLenPlayerGameBilled{
						SnId:     proto.Int32(playerEx.SnId),
						IsWin:    proto.Int32(2),
						WinCoin:  proto.Int64(gainScore),
						GameCoin: proto.Int64(playerEx.GetCoin()),
					}
					isWin := int32(0)
					billCoin := playerEx.bombScore - gainScore
					if billCoin > 0 {
						isWin = 1
					} else if billCoin < 0 {
						isWin = -1
					}
					playerEx.CurIsWin = int64(isWin)
					tienlenPerson := model.TienLenPerson{
						UserId:      playerEx.SnId,
						UserIcon:    playerEx.Head,
						Platform:    playerEx.Platform,
						Channel:     playerEx.Channel,
						Promoter:    playerEx.BeUnderAgentCode,
						PackageTag:  playerEx.PackageID,
						InviterId:   playerEx.InviterId,
						WBLevel:     playerEx.WBLevel,
						IsRob:       playerEx.IsRob,
						IsFirst:     sceneEx.IsPlayerFirst(sceneEx.GetPlayer(playerEx.SnId)),
						IsLeave:     false,
						IsWin:       isWin,
						GainCoin:    -gainScore,
						BombCoin:    playerEx.bombScore,
						BillCoin:    billCoin,
						GainTaxCoin: 0,
						BombTaxCoin: playerEx.bombTaxScore,
						BillTaxCoin: playerEx.bombTaxScore,
						Seat:        playerEx.GetPos(),
						IsTianHu:    false,
					}
					tienlenPerson.DelOrderCards = make(map[int][]int32, len(sceneEx.delOrders))
					for i2, orderSnid := range sceneEx.delOrders {
						if orderSnid == playerEx.SnId {
							tienlenPerson.DelOrderCards[i2] = sceneEx.delCards[i2]
						}
					}
					for _, card := range playerEx.cards {
						if card != rule.InvalideCard {
							billData.Cards = append(billData.Cards, card)
							tienlenPerson.CardInfoEnd = append(tienlenPerson.CardInfoEnd, card)
						}
					}
					//排下序，正常应该客户端排序
					sort.Slice(billData.Cards, func(i, j int) bool {
						v_i := rule.Value(int32(billData.Cards[i]))
						v_j := rule.Value(int32(billData.Cards[j]))
						c_i := rule.Color(int32(billData.Cards[i]))
						c_j := rule.Color(int32(billData.Cards[j]))
						if v_i > v_j {
							return false
						} else if v_i == v_j {
							return c_i < c_j
						}
						return true
					})
					pack.Datas = append(pack.Datas, billData)
					playerEx.Statics(sceneEx.KeyGameId, sceneEx.KeyGamefreeId, billCoin, true)
					sceneEx.TryBillExGameDrop(playerEx.Player) //输家
					tienlenType.PlayerData = append(tienlenType.PlayerData, tienlenPerson)
				}
				logger.Logger.Trace("SceneBilledStateTienLe, winSnids: ", sceneEx.winSnids, " winScore: ", winScore)

				//赢家
				if len(sceneEx.tianHuSnids) != 0 {
					winScore = winScore / int64(len(sceneEx.tianHuSnids))
				}
				for _, winSnid := range sceneEx.winSnids {
					playerEx := sceneEx.players[winSnid]
					if playerEx != nil {
						taxRate := sceneEx.DbGameFree.GetTaxRate()                               //万分比
						gainScore := int64(float64(winScore) * float64(10000-taxRate) / 10000.0) //税后
						gainTaxScore := winScore - gainScore
						if sceneEx.IsMatchScene() {
							playerEx.AddCoinNoLog(int64(gainScore), 0)
						} else {
							playerEx.AddCoin(gainScore, common.GainWay_CoinSceneWin, 0, "system", s.GetSceneName())
						}
						playerEx.winCoin += gainScore
						billData := &tienlen.TienLenPlayerGameBilled{
							SnId:     proto.Int32(playerEx.SnId),
							IsWin:    proto.Int32(1),
							WinCoin:  proto.Int64(gainScore),
							GameCoin: proto.Int64(playerEx.GetCoin()),
						}
						isWin := int32(0)
						billCoin := playerEx.bombScore + gainScore
						if billCoin > 0 {
							isWin = 1
						} else if billCoin < 0 {
							isWin = -1
						}
						playerEx.CurIsWin = int64(isWin)
						tienlenPerson := model.TienLenPerson{
							UserId:      playerEx.SnId,
							UserIcon:    playerEx.Head,
							Platform:    playerEx.Platform,
							Channel:     playerEx.Channel,
							Promoter:    playerEx.BeUnderAgentCode,
							PackageTag:  playerEx.PackageID,
							InviterId:   playerEx.InviterId,
							WBLevel:     playerEx.WBLevel,
							IsRob:       playerEx.IsRob,
							IsFirst:     sceneEx.IsPlayerFirst(sceneEx.GetPlayer(playerEx.SnId)),
							IsLeave:     false,
							IsWin:       isWin,
							GainCoin:    gainScore,
							BombCoin:    playerEx.bombScore,
							BillCoin:    billCoin,
							GainTaxCoin: gainTaxScore,
							BombTaxCoin: playerEx.bombTaxScore,
							BillTaxCoin: playerEx.bombTaxScore + gainTaxScore,
							Seat:        playerEx.GetPos(),
							IsTianHu:    sceneEx.IsTianhuPlayer(playerEx.SnId),
						}
						tienlenPerson.DelOrderCards = make(map[int][]int32, len(sceneEx.delOrders))
						for i2, orderSnid := range sceneEx.delOrders {
							if orderSnid == playerEx.SnId {
								tienlenPerson.DelOrderCards[i2] = sceneEx.delCards[i2]
							}
						}
						for _, card := range playerEx.cards {
							if card != rule.InvalideCard {
								billData.Cards = append(billData.Cards, card)
								tienlenPerson.CardInfoEnd = append(tienlenPerson.CardInfoEnd, card)
							}
						}
						//排下序，正常应该客户端排序
						sort.Slice(billData.Cards, func(i, j int) bool {
							v_i := rule.Value(int32(billData.Cards[i]))
							v_j := rule.Value(int32(billData.Cards[j]))
							c_i := rule.Color(int32(billData.Cards[i]))
							c_j := rule.Color(int32(billData.Cards[j]))
							if v_i > v_j {
								return false
							} else if v_i == v_j {
								return c_i < c_j
							}
							return true
						})
						pack.Datas = append(pack.Datas, billData)
						playerEx.Statics(sceneEx.KeyGameId, sceneEx.KeyGamefreeId, billCoin, true)
						sceneEx.TryBillExGameDrop(playerEx.Player) //赢家
						tienlenType.PlayerData = append(tienlenType.PlayerData, tienlenPerson)
					}
				}
			}
		}

		proto.SetDefaults(pack)
		s.Broadcast(int(tienlen.TienLenPacketID_PACKET_SCTienLenGameBilled), pack, 0)
		logger.Logger.Trace("TienLenPacketID_PACKET_SCTienLenGameBilled gameFreeId:", sceneEx.GetGameFreeId(), ";pack:", pack)

		info, err := model.MarshalGameNoteByFIGHT(&tienlenType)
		if err == nil {
			isSave := false
			var logid string
			for _, o_player := range tienlenType.PlayerData {
				if !sceneEx.Testing && !o_player.IsRob {
					if logid == "" {
						logid, _ = model.AutoIncGameLogId()
					}
					var totalin, totalout int64
					//if o_player.IsWin > 0 {
					//	totalout = o_player.BillCoin
					//} else {
					//	totalin = -o_player.BillCoin
					//}
					if o_player.GainCoin < 0 {
						totalin -= (o_player.GainCoin + o_player.GainTaxCoin)
					} else {
						totalout += (o_player.GainCoin + o_player.GainTaxCoin)
					}
					if o_player.BombCoin < 0 {
						totalin -= (o_player.BombCoin + o_player.BombTaxCoin)
					} else {
						totalout += (o_player.BombCoin + o_player.BombTaxCoin)
					}
					validFlow := totalin + totalout
					validBet := common.AbsI64(totalin - totalout)
					sceneEx.SaveGamePlayerListLog(o_player.UserId,
						base.GetSaveGamePlayerListLogParam(o_player.Platform, o_player.Channel, o_player.Promoter,
							o_player.PackageTag, logid, o_player.InviterId, totalin, totalout, o_player.BillTaxCoin,
							0, 0, 0, validBet, validFlow, o_player.IsFirst, o_player.IsLeave))
					isSave = true
					sceneEx.SaveFriendRecord(o_player.UserId, o_player.IsWin)
				}
			}
			if isSave {
				sceneEx.SaveGameDetailedLog(logid, info, &base.GameDetailedParam{})
				if !sceneEx.IsMatchScene() {
					sceneEx.SetSystemCoinOut(sceneEx.SystemCoinOut())
					base.CoinPoolMgr.PushCoin(sceneEx.GetCoinSceneTypeId(), sceneEx.GetGroupId(), sceneEx.GetPlatform(), sceneEx.GetSystemCoinOut())
				}
			}
		}
		sceneEx.SetGaming(false)
		sceneEx.NotifySceneRoundPause()
	}
}

// 状态离开时
func (this *SceneBilledStateTienLen) OnLeave(s *base.Scene) {
	this.SceneBaseStateTienLen.OnLeave(s)
	if sceneEx, ok := s.GetExtraData().(*TienLenSceneData); ok {
		sceneEx.lastGamingPlayerNum = sceneEx.curGamingPlayerNum
		sceneEx.curGamingPlayerNum = 0
		if len(sceneEx.tianHuSnids) != sceneEx.GetGameingPlayerCnt() { //非和局
			sceneEx.lastWinSnid = sceneEx.winSnids[0]
		}

		if s.CheckNeedDestroy() || (s.IsMatchScene() && (!s.MatchFinals || (s.MatchFinals && s.NumOfGames >= 2))) { // 非决赛打一场 决赛打两场
			//if s.CheckNeedDestroy() || s.IsMatchScene() {
			for _, p := range sceneEx.players {
				if p != nil {
					s.PlayerLeave(p.Player, common.PlayerLeaveReason_Normal, true)
				}
			}
			sceneEx.SceneDestroy(true)
		}
	}
}

// 玩家操作
func (this *SceneBilledStateTienLen) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if this.SceneBaseStateTienLen.OnPlayerOp(s, p, opcode, params) {
		return true
	}
	return true
}

// 玩家事件
func (this *SceneBilledStateTienLen) OnPlayerEvent(s *base.Scene, p *base.Player, evtcode int, params []int64) {
	this.SceneBaseStateTienLen.OnPlayerEvent(s, p, evtcode, params)
}

func (this *SceneBilledStateTienLen) OnTick(s *base.Scene) {
	this.SceneBaseStateTienLen.OnTick(s)

	if sceneEx, ok := s.GetExtraData().(*TienLenSceneData); ok {
		newTime := rule.TienLenBilledTimeout
		if sceneEx.isAllRob {
			newTime = time.Second * 1
		}
		if time.Now().Sub(sceneEx.StateStartTime) > newTime {
			//开始前再次检查开始条件
			if sceneEx.CanStart() == true {
				s.ChangeSceneState(rule.TienLenSceneStateWaitStart)
			} else {
				s.ChangeSceneState(rule.TienLenSceneStateWaitPlayer)
			}
		}
	}
}

// //////////////////////////////////////////////////////////////////////////////
func (this *ScenePolicyTienLen) RegisteSceneState(state base.SceneState) {
	if state == nil {
		return
	}
	stateid := state.GetState()
	if stateid < 0 || stateid >= rule.TienLenSceneStateMax {
		return
	}
	this.states[stateid] = state
}

func (this *ScenePolicyTienLen) GetSceneState(s *base.Scene, stateid int) base.SceneState {
	if stateid >= 0 && stateid < rule.TienLenSceneStateMax {
		return this.states[stateid]
	}
	return nil
}

func init() {
	ScenePolicyTienLenSington.RegisteSceneState(&SceneWaitPlayerStateTienLen{})
	ScenePolicyTienLenSington.RegisteSceneState(&SceneWaitStartStateTienLen{})
	ScenePolicyTienLenSington.RegisteSceneState(&SceneHandCardStateTienLen{})
	ScenePolicyTienLenSington.RegisteSceneState(&ScenePlayerOpStateTienLen{})
	ScenePolicyTienLenSington.RegisteSceneState(&SceneBilledStateTienLen{})

	core.RegisteHook(core.HOOK_BEFORE_START, func() error {
		base.RegisteScenePolicy(common.GameId_TienLen, 0, ScenePolicyTienLenSington)
		base.RegisteScenePolicy(common.GameId_TienLen_yl, 0, ScenePolicyTienLenSington)
		base.RegisteScenePolicy(common.GameId_TienLen_toend, 0, ScenePolicyTienLenSington)
		base.RegisteScenePolicy(common.GameId_TienLen_yl_toend, 0, ScenePolicyTienLenSington)
		base.RegisteScenePolicy(common.GameId_TienLen_m, 0, ScenePolicyTienLenSington)
		base.RegisteScenePolicy(common.GameId_TienLen_yl_toend_m, 0, ScenePolicyTienLenSington)
		return nil
	})
}
