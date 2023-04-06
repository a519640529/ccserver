package dragonvstiger

import (
	"encoding/json"
	"fmt"
	"games.yol.com/win88/common"
	. "games.yol.com/win88/gamerule/dragonvstiger"
	"games.yol.com/win88/gamesrv/base"
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/dragonvstiger"
	"games.yol.com/win88/protocol/player"
	"games.yol.com/win88/protocol/server"
	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/timer"
	"strconv"
	"time"
)

/*
pos:1-6表示座位的6个位置
	7表示我的位置
    8表示在线的位置（所有其它玩家）
	9表示庄家
*/
var ScenePolicyDragonVsTigerSington = &ScenePolicyDragonVsTiger{}

type ScenePolicyDragonVsTiger struct {
	base.BaseScenePolicy
	states [DragonVsTigerSceneStateMax]base.SceneState
}

func (this *ScenePolicyDragonVsTiger) CreateSceneExData(s *base.Scene) interface{} {
	sceneEx := NewDragonVsTigerSceneData(s)
	if sceneEx != nil {
		if sceneEx.init() {
			s.ExtraData = sceneEx
		}
	}
	return sceneEx
}
func (this *ScenePolicyDragonVsTiger) CreatePlayerExData(s *base.Scene, p *base.Player) interface{} {
	playerEx := &DragonVsTigerPlayerData{Player: p}
	if playerEx != nil {
		p.ExtraData = playerEx
	}
	return playerEx
}
func DragonVsTigerBatchSendBet(sceneEx *DragonVsTigerSceneData, force bool) {
	needSend := false
	pack := &dragonvstiger.SCDragonVsTigerSendBet{}
	var olBetChips [DVST_ZONE_MAX]int64
	olBetInfo := &dragonvstiger.DragonVsTigerBetInfo{
		SnId: proto.Int(0),
	}
	for _, playerEx := range sceneEx.seats {
		if playerEx.Pos == DVST_OLPOS {
			for i := 0; i < DVST_ZONE_MAX; i++ {
				if len(playerEx.betCacheInfo[i]) != 0 {
					for k, v := range playerEx.betCacheInfo[i] {
						olBetChips[i] += int64(k * v)
					}
					playerEx.betCacheInfo[i] = make(map[int]int)
					needSend = true
				}
			}
		}
	}
	if needSend || force {
		olBetInfo.TotalChips = olBetChips[:]
		pack.Data = append(pack.Data, olBetInfo)
		for i := 0; i < DVST_ZONE_MAX; i++ {
			pack.TotalChips = append(pack.TotalChips, int64(sceneEx.betInfo[i]))
		}
		proto.SetDefaults(pack)
		sceneEx.Broadcast(int(dragonvstiger.DragonVsTigerPacketID_PACKET_SC_DVST_SENDBET), pack, 0)
	}
}
func (this *ScenePolicyDragonVsTiger) OnStart(s *base.Scene) {
	logger.Logger.Trace("(this *ScenePolicyDragonVsTiger) OnStart, sceneId=", s.SceneId)
	sceneEx := NewDragonVsTigerSceneData(s)
	if sceneEx != nil {
		if sceneEx.init() {
			if sceneEx.hBatchSend != timer.TimerHandle(0) {
				timer.StopTimer(sceneEx.hBatchSend)
				sceneEx.hBatchSend = timer.TimerHandle(0)
			}

			if hNext, ok := common.DelayInvake(func() {
				if sceneEx.SceneState.GetState() != DragonVsTigerSceneStateStake {
					return
				}
				DragonVsTigerBatchSendBet(sceneEx, false)
			}, nil, DragonVsTigerBatchSendBetTimeout, -1); ok {
				sceneEx.hBatchSend = hNext
			}
			s.ExtraData = sceneEx
			s.ChangeSceneState(DragonVsTigerSceneStateStakeAnt)
		}
	}
}
func (this *ScenePolicyDragonVsTiger) OnStop(s *base.Scene) {
	logger.Logger.Trace("(this *ScenePolicyDragonVsTiger) OnStop , sceneId=", s.SceneId)
	if sceneEx, ok := s.ExtraData.(*DragonVsTigerSceneData); ok {
		if sceneEx.hRunRecord != timer.TimerHandle(0) {
			timer.StopTimer(sceneEx.hRunRecord)
			sceneEx.hRunRecord = timer.TimerHandle(0)
		}
		sceneEx.SaveData()
	}
}
func (this *ScenePolicyDragonVsTiger) OnTick(s *base.Scene) {
	if s == nil {
		return
	}
	if s.SceneState != nil {
		s.SceneState.OnTick(s)
	}
}
func (this *ScenePolicyDragonVsTiger) GetPlayerNum(s *base.Scene) int32 {
	if s == nil {
		return 0
	}
	if sceneEx, ok := s.ExtraData.(*DragonVsTigerSceneData); ok {
		return int32(len(sceneEx.seats))
	}
	return 0
}
func (this *ScenePolicyDragonVsTiger) OnPlayerEnter(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyDragonVsTiger) OnPlayerEnter, sceneId=", s.SceneId, " player=", p.Name)
	if sceneEx, ok := s.ExtraData.(*DragonVsTigerSceneData); ok {
		playerEx := &DragonVsTigerPlayerData{Player: p}
		if playerEx != nil {
			playerEx.Clean()
			playerEx.Pos = DVST_OLPOS
			sceneEx.seats = append(sceneEx.seats, playerEx)
			sceneEx.players[p.SnId] = playerEx
			p.ExtraData = playerEx

			if p.Coin < int64(sceneEx.DbGameFree.GetBetLimit()) || p.Coin < int64(sceneEx.DbGameFree.GetOtherIntParams()[0]) {
				p.MarkFlag(base.PlayerState_GameBreak)
			}

			dragonVsTigerSendRoomInfo(s, p, sceneEx, playerEx)
			s.FirePlayerEvent(p, base.PlayerEventEnter, nil)
		}
	}
}
func (this *ScenePolicyDragonVsTiger) OnPlayerLeave(s *base.Scene, p *base.Player, reason int) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyDragonVsTiger) OnPlayerLeave, sceneId=", s.SceneId, " player=", p.Name)
	if sceneEx, ok := s.ExtraData.(*DragonVsTigerSceneData); ok {
		if this.CanChangeCoinScene(s, p) {
			sceneEx.OnPlayerLeave(p, reason)
			s.FirePlayerEvent(p, base.PlayerEventLeave, []int64{int64(reason)})
		}
	}
}
func (this *ScenePolicyDragonVsTiger) OnPlayerDropLine(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyDragonVsTiger) OnPlayerDropLine, sceneId=", s.SceneId, " player=", p.Name)
	s.FirePlayerEvent(p, base.PlayerEventDropLine, nil)
	if sceneEx, ok := s.ExtraData.(*DragonVsTigerSceneData); ok {
		if sceneEx.Gaming {
			return
		}
	}
}
func (this *ScenePolicyDragonVsTiger) OnPlayerRehold(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyDragonVsTiger) OnPlayerRehold, sceneId=", s.SceneId, " player=", p.Name)
	if sceneEx, ok := s.ExtraData.(*DragonVsTigerSceneData); ok {
		if playerEx, ok := p.ExtraData.(*DragonVsTigerPlayerData); ok {

			dragonVsTigerSendRoomInfo(s, p, sceneEx, playerEx)
			s.FirePlayerEvent(p, base.PlayerEventRehold, nil)
		}
	}
}

//玩家返回房间
func (this *ScenePolicyDragonVsTiger) OnPlayerReturn(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyDragonVsTiger) OnPlayerReturn, sceneId=", s.SceneId, " player=", p.Name)
	if sceneEx, ok := s.ExtraData.(*DragonVsTigerSceneData); ok {
		if playerEx, ok := p.ExtraData.(*DragonVsTigerPlayerData); ok {

			dragonVsTigerSendRoomInfo(s, p, sceneEx, playerEx)
			s.FirePlayerEvent(p, base.PlayerEventReturn, nil)
		}
	}
}
func (this *ScenePolicyDragonVsTiger) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if s == nil || p == nil {
		return false
	}
	if s.SceneState != nil {
		return s.SceneState.OnPlayerOp(s, p, opcode, params)
	}
	return true
}
func (this *ScenePolicyDragonVsTiger) OnPlayerEvent(s *base.Scene, p *base.Player, evtcode int, params []int64) {
	if s == nil || p == nil {
		return
	}
	if s.SceneState != nil {
		s.SceneState.OnPlayerEvent(s, p, evtcode, params)
	}
}

func (this *ScenePolicyDragonVsTiger) IsCompleted(s *base.Scene) bool {
	return false
}
func (this *ScenePolicyDragonVsTiger) IsCanForceStart(s *base.Scene) bool {
	return true
}
func (this *ScenePolicyDragonVsTiger) CanChangeCoinScene(s *base.Scene, p *base.Player) bool {
	if s == nil || p == nil {
		return true
	}
	if s.SceneState != nil {
		return s.SceneState.CanChangeCoinScene(s, p)
	}
	return true
}
func (this *ScenePolicyDragonVsTiger) PacketGameData(s *base.Scene) interface{} {
	if s == nil {
		return nil
	}
	if s.SceneState != nil {
		if sceneEx, ok := s.ExtraData.(*DragonVsTigerSceneData); ok {
			switch s.SceneState.GetState() {
			case DragonVsTigerSceneStateStake:
				lt := int32((DragonVsTigerStakeTimeout - time.Now().Sub(sceneEx.StateStartTime)) / time.Second)
				pack := &server.GWDTRoomInfo{
					DCoin:      proto.Int(sceneEx.betInfo[DVST_ZONE_DRAGON] - sceneEx.betInfoRob[DVST_ZONE_DRAGON]),
					TCoin:      proto.Int(sceneEx.betInfo[DVST_ZONE_TIGER] - sceneEx.betInfoRob[DVST_ZONE_TIGER]),
					NCoin:      proto.Int(sceneEx.betInfo[DVST_ZONE_DRAW] - sceneEx.betInfoRob[DVST_ZONE_DRAW]),
					Onlines:    proto.Int(sceneEx.GetRealPlayerCnt()),
					LeftTimes:  proto.Int32(lt),
					CoinPool:   proto.Int64(base.CoinPoolMgr.LoadCoin(sceneEx.GetGameFreeId(), sceneEx.Platform, sceneEx.GroupId)),
					NumOfGames: proto.Int(sceneEx.NumOfGames),
					LoopNum:    proto.Int(sceneEx.LoopNum),
					Results:    sceneEx.ProtoResults(),
				}
				for _, value := range sceneEx.players {
					if value.IsRob {
						continue
					}
					win, lost := value.GetStaticsData(s.KeyGameId)
					pack.Players = append(pack.Players, &server.PlayerDTCoin{
						NickName: proto.String(value.Name),
						Snid:     proto.Int32(value.SnId),
						DCoin:    proto.Int(value.betInfo[DVST_ZONE_DRAGON]),
						TCoin:    proto.Int(value.betInfo[DVST_ZONE_TIGER]),
						NCoin:    proto.Int(value.betInfo[DVST_ZONE_DRAW]),
						Totle:    proto.Int64(win - lost),
					})
				}
				return pack
			default:
				return &server.GWDTRoomInfo{}
			}
		}
	}
	return nil
}
func (this *ScenePolicyDragonVsTiger) NotifyGameState(s *base.Scene) {
	if s.GetState() > 2 {
		return
	}
	sec := int((DragonVsTigerStakeAntTimeout + DragonVsTigerStakeTimeout).Seconds())
	if sceneEx, ok := s.ExtraData.(*DragonVsTigerSceneData); ok {
		s.SyncGameState(sec, len(sceneEx.upplayerlist))
	}
}
func DragonVsTigerCreateSeats(sceneEx *DragonVsTigerSceneData) []*dragonvstiger.DragonVsTigerPlayerData {

	var datas []*dragonvstiger.DragonVsTigerPlayerData
	cnt := 0
	const N = DVST_RICHTOP5 + 2
	var seats [N]*DragonVsTigerPlayerData
	if sceneEx.winTop1 != nil {
		seats[cnt] = sceneEx.winTop1
		cnt++
	}
	for i := 0; i < DVST_RICHTOP5; i++ {
		if sceneEx.betTop5[i] != nil && sceneEx.betTop5[i] != sceneEx.winTop1 {
			seats[cnt] = sceneEx.betTop5[i]
			cnt++
		}
	}
	for i := 0; i < N-1; i++ {
		if seats[i] != nil {
			pd := &dragonvstiger.DragonVsTigerPlayerData{
				SnId: proto.Int32(seats[i].SnId),
				Name: proto.String(seats[i].Name),
				Head: proto.Int32(seats[i].Head),
				Sex:  proto.Int32(seats[i].Sex),
				Coin: proto.Int64(seats[i].Coin),
				Pos:  proto.Int(seats[i].Pos),
				Flag: proto.Int(seats[i].GetFlag()),
				City: proto.String(seats[i].GetCity()),

				Lately20Win: proto.Int64(seats[i].lately20Win),
				Lately20Bet: proto.Int64(seats[i].lately20Bet),
				HeadOutLine: proto.Int32(seats[i].HeadOutLine),
				VIP:         proto.Int32(seats[i].VIP),
				NiceId:      proto.Int32(seats[i].NiceId),
			}
			datas = append(datas, pd)
		}
	}

	//把庄家信息扔进去
	if sceneEx.bankerSnId != -1 {
		if banker, exist := sceneEx.players[sceneEx.bankerSnId]; exist {
			seats[cnt] = banker
			cnt++

			pd := &dragonvstiger.DragonVsTigerPlayerData{
				SnId: proto.Int32(banker.SnId),
				Name: proto.String(banker.Name),
				Head: proto.Int32(banker.Head),
				Sex:  proto.Int32(banker.Sex),
				Coin: proto.Int64(banker.Coin),
				Pos:  proto.Int(DVST_BANKERPOS),
				Flag: proto.Int(banker.GetFlag()),
				City: proto.String(banker.GetCity()),

				Lately20Win: proto.Int64(banker.lately20Win),
				Lately20Bet: proto.Int64(banker.lately20Bet),
				HeadOutLine: proto.Int32(banker.HeadOutLine),
				VIP:         proto.Int32(banker.VIP),
				NiceId:      proto.Int32(banker.NiceId),
			}
			datas = append(datas, pd)
		}
	}

	return datas
}
func DragonVsTigerCreateRoomInfoPacket(s *base.Scene, sceneEx *DragonVsTigerSceneData, playerEx *DragonVsTigerPlayerData) proto.Message {
	pack := &dragonvstiger.SCDragonVsTigerRoomInfo{
		RoomId:    proto.Int(s.SceneId),
		Creator:   proto.Int32(s.GetCreator()),
		GameId:    proto.Int(s.GameId),
		RoomMode:  proto.Int(s.SceneMode),
		AgentId:   proto.Int32(s.GetCreator()),
		SceneType: proto.Int(s.SceneType),
		Cards:     common.CopySliceInt32(sceneEx.cards[:]),
		Params: []int32{sceneEx.DbGameFree.GetLimitCoin(), sceneEx.DbGameFree.GetMaxCoinLimit(), sceneEx.DbGameFree.GetServiceFee(),
			sceneEx.DbGameFree.GetLowerThanKick(), sceneEx.DbGameFree.GetBaseScore(), 0,
			sceneEx.DbGameFree.GetBetLimit(), sceneEx.DbGameFree.GetBanker()},
		NumOfGames:     proto.Int(sceneEx.NumOfGames),
		State:          proto.Int(s.SceneState.GetState()),
		TimeOut:        proto.Int(s.SceneState.GetTimeout(s)),
		BankerId:       proto.Int32(sceneEx.bankerSnId),
		Players:        DragonVsTigerCreateSeats(sceneEx),
		Trend100Cur:    sceneEx.trend100Cur,
		Trend20Lately:  sceneEx.trend20Lately,
		OLNum:          proto.Int(len(sceneEx.seats)),
		DisbandGen:     proto.Int(sceneEx.GetDisbandGen()),
		ParamsEx:       s.GetParamsEx(),
		OtherIntParams: s.DbGameFree.OtherIntParams,
		LoopNum:        proto.Int(sceneEx.LoopNum),
	}
	for _, value := range sceneEx.DbGameFree.GetMaxBetCoin() {
		pack.Params = append(pack.Params, value)
	}

	if playerEx != nil {
		pd := &dragonvstiger.DragonVsTigerPlayerData{
			SnId: proto.Int32(playerEx.SnId),
			Name: proto.String(playerEx.Name),
			Head: proto.Int32(playerEx.Head),
			Sex:  proto.Int32(playerEx.Sex),
			Coin: proto.Int64(playerEx.Coin),
			Pos:  proto.Int(DVST_SELFPOS),
			Flag: proto.Int(playerEx.GetFlag()),
			City: proto.String(playerEx.GetCity()),

			Lately20Win: proto.Int64(playerEx.lately20Win),
			Lately20Bet: proto.Int64(playerEx.lately20Bet),
			HeadOutLine: proto.Int32(playerEx.HeadOutLine),
			VIP:         proto.Int32(playerEx.VIP),
			NiceId:      proto.Int32(playerEx.NiceId),
		}
		pack.Players = append(pack.Players, pd)
	}
	for i := 0; i < DVST_ZONE_MAX; i++ {
		total := sceneEx.betInfo[i]
		if total != 0 {
			pack.TotalChips = append(pack.TotalChips, int64(total))
		} else {
			pack.TotalChips = append(pack.TotalChips, 0)
		}
	}
	if playerEx != nil {
		for i := 0; i < DVST_ZONE_MAX; i++ {
			if len(playerEx.betDetailInfo[i]) != 0 {
				chip := &dragonvstiger.DragonVsTigerZoneChips{
					Zone: proto.Int(i),
				}
				for k, v := range playerEx.betDetailInfo[i] {
					chip.Detail = append(chip.Detail, &dragonvstiger.DragonVsTigerChips{
						Chip:  proto.Int(k),
						Count: proto.Int(v),
					})
				}
				pack.MyChips = append(pack.MyChips, chip)
			}
		}
	}
	proto.SetDefaults(pack)
	return pack
}
func dragonVsTigerSendRoomInfo(s *base.Scene, p *base.Player, sceneEx *DragonVsTigerSceneData, playerEx *DragonVsTigerPlayerData) {
	pack := DragonVsTigerCreateRoomInfoPacket(s, sceneEx, playerEx)

	p.SendToClient(int(dragonvstiger.DragonVsTigerPacketID_PACKET_SC_DVST_ROOMINFO), pack)
}
func dragonVsTigerSendSeatInfo(s *base.Scene, sceneEx *DragonVsTigerSceneData) {
	pack := &dragonvstiger.SCDragonVsTigerSeats{
		PlayerNum: proto.Int(len(sceneEx.seats)),
		Data:      DragonVsTigerCreateSeats(sceneEx),
		BankerId:  proto.Int32(sceneEx.bankerSnId),
	}
	proto.SetDefaults(pack)
	s.Broadcast(int(dragonvstiger.DragonVsTigerPacketID_PACKET_SC_DVST_SEATS), pack, 0)
}

type SceneBaseStateDragonVsTiger struct {
}

func (this *SceneBaseStateDragonVsTiger) GetTimeout(s *base.Scene) int {
	if sceneEx, ok := s.ExtraData.(*DragonVsTigerSceneData); ok {
		return int(time.Now().Sub(sceneEx.StateStartTime) / time.Second)
	}
	return 0
}
func (this *SceneBaseStateDragonVsTiger) CanChangeTo(s base.SceneState) bool {
	return true
}
func (this *SceneBaseStateDragonVsTiger) CanChangeCoinScene(s *base.Scene, p *base.Player) bool {
	if playerEx, ok := p.ExtraData.(*DragonVsTigerPlayerData); ok {
		if sceneEx, ok := s.ExtraData.(*DragonVsTigerSceneData); ok {
			if playerEx.betTotal != 0 {
				playerEx.OpCode = player.OpResultCode_OPRC_Hundred_YouHadBetCannotLeave
				return false
			}
			if playerEx.SnId == sceneEx.bankerSnId {
				p.OpCode = player.OpResultCode_OPRC_Hundred_YouHadBankerCannotLeave
				return false
			}
		}
	}
	return true
}
func (this *SceneBaseStateDragonVsTiger) OnEnter(s *base.Scene) {
	if sceneEx, ok := s.ExtraData.(*DragonVsTigerSceneData); ok {
		sceneEx.StateStartTime = time.Now()
	}
}
func (this *SceneBaseStateDragonVsTiger) OnLeave(s *base.Scene) {
}
func (this *SceneBaseStateDragonVsTiger) OnTick(s *base.Scene) {
}
func (this *SceneBaseStateDragonVsTiger) SendSCPlayerOp(s *base.Scene, p *base.Player, pos int, opcode int, opRetCode dragonvstiger.OpResultCode, params []int64, broadcastall bool) {
	pack := &dragonvstiger.SCDragonVsTiggerOp{
		OpCode:    proto.Int(opcode),
		SnId:      proto.Int32(p.SnId),
		OpRetCode: opRetCode,
		Params:    params,
	}
	proto.SetDefaults(pack)
	if broadcastall {
		s.Broadcast(int(dragonvstiger.DragonVsTigerPacketID_PACKET_SC_DVST_PLAYEROP), pack, 0)
	} else {
		p.SendToClient(int(dragonvstiger.DragonVsTigerPacketID_PACKET_SC_DVST_PLAYEROP), pack)
	}
}
func (this *SceneBaseStateDragonVsTiger) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if playerEx, ok := p.ExtraData.(*DragonVsTigerPlayerData); ok {
		if sceneEx, ok := s.ExtraData.(*DragonVsTigerSceneData); ok {
			switch opcode {
			case DragonVsTigerPlayerOpGetOLList:
				seats := make([]*DragonVsTigerPlayerData, 0, DVST_OLTOP20+1)
				if sceneEx.winTop1 != nil {
					seats = append(seats, sceneEx.winTop1)
				}
				count := len(sceneEx.seats)
				topCnt := 0
				for i := 0; i < count && topCnt < DVST_OLTOP20; i++ {
					if sceneEx.seats[i] != sceneEx.winTop1 {
						seats = append(seats, sceneEx.seats[i])
						topCnt++
					}
				}
				pack := &dragonvstiger.SCDragonVsTigerPlayerList{}
				for i := 0; i < len(seats); i++ {
					pack.Data = append(pack.Data, &dragonvstiger.DragonVsTigerPlayerData{
						SnId: proto.Int32(seats[i].SnId),
						Name: proto.String(seats[i].Name),
						Head: proto.Int32(seats[i].Head),
						Sex:  proto.Int32(seats[i].Sex),
						Coin: proto.Int64(seats[i].Coin),
						Pos:  proto.Int(seats[i].Pos),
						Flag: proto.Int(seats[i].GetFlag()),
						City: proto.String(seats[i].GetCity()),

						Lately20Win: proto.Int64(seats[i].lately20Win),
						Lately20Bet: proto.Int64(seats[i].lately20Bet),
						HeadOutLine: proto.Int32(seats[i].HeadOutLine),
						VIP:         proto.Int32(seats[i].VIP),
						NiceId:      proto.Int32(seats[i].NiceId),
					})
				}
				pack.OLNum = int32(len(s.Players))
				//truePlayerCount := int32(len(s.Players))

				//correctNum := sceneEx.DbGameFree.GetCorrectNum()
				//correctRate := sceneEx.DbGameFree.GetCorrectRate()
				//fakePlayerCount := correctNum + truePlayerCount*correctRate/100 + sceneEx.DbGameFree.GetDeviation()
				//pack.OLNum = proto.Int32(truePlayerCount + fakePlayerCount + rand.Int31n(truePlayerCount))
				proto.SetDefaults(pack)
				p.SendToClient(int(dragonvstiger.DragonVsTigerPacketID_PACKET_SC_DVST_PLAYERLIST), pack)
				return true
			case DragonVsTigerPlayerOpUpBanker: //上庄
				if sceneEx.bankerSnId == playerEx.SnId {
					return true
				}
				if sceneEx.DbGameFree.GetBanker() == 0 {
					this.SendSCPlayerOp(s, p, playerEx.Pos, opcode, dragonvstiger.OpResultCode_OPRC_BankerLimit, params, false)
					return true
				}
				if (p.Coin) < int64(sceneEx.DbGameFree.GetBanker()) {
					this.SendSCPlayerOp(s, p, playerEx.Pos, opcode, dragonvstiger.OpResultCode_OPRC_BankerLimit, params, false)
					return true
				}
				if len(sceneEx.upplayerlist) > 0 {
					for _, v := range sceneEx.upplayerlist {
						if v.SnId == playerEx.SnId {
							this.SendSCPlayerOp(s, p, playerEx.Pos, opcode, dragonvstiger.OpResultCode_OPRC_BankerWaiting, params, false)
							return true
						}
					}
				}
				sceneEx.upplayerlist = append(sceneEx.upplayerlist, playerEx)
				this.SendSCPlayerOp(s, p, playerEx.Pos, opcode, dragonvstiger.OpResultCode_OPRC_Sucess, params, false)
				sceneEx.Broadcast(int(dragonvstiger.DragonVsTigerPacketID_PACKET_SC_DVST_UPLIST), &dragonvstiger.SCDragonVsTiggerUpList{
					Count:   proto.Int(len(sceneEx.upplayerlist)),
					IsExist: proto.Int32(-1),
				}, 0)
				return false
			case DragonVsTigerPlayerOpDwonBanker: //下庄
				down := false
				if len(sceneEx.upplayerlist) > 0 {
					for index, v := range sceneEx.upplayerlist {
						if v.SnId == playerEx.SnId {
							sceneEx.upplayerlist = append(sceneEx.upplayerlist[:index], sceneEx.upplayerlist[index+1:]...)
							down = true
						}
					}
				}
				if !down {
					this.SendSCPlayerOp(s, p, playerEx.Pos, opcode, dragonvstiger.OpResultCode_OPRC_NotBankerWaiting, params, false)
					return true
				}
				this.SendSCPlayerOp(s, p, playerEx.Pos, opcode, dragonvstiger.OpResultCode_OPRC_Sucess, params, false)
				sceneEx.Broadcast(int(dragonvstiger.DragonVsTigerPacketID_PACKET_SC_DVST_UPLIST), &dragonvstiger.SCDragonVsTiggerUpList{
					Count:   proto.Int(len(sceneEx.upplayerlist)),
					IsExist: proto.Int32(-1),
				}, 0)
				return false
			case DragonVsTigerPlayerOpNowDwonBanker: //在庄的下庄
				logger.Logger.Tracef("玩家是庄，现在申请下庄 sceneEx.upplayerCount = %v", sceneEx.upplayerCount)
				opRetCode := dragonvstiger.OpResultCode_OPRC_Sucess
				if sceneEx.upplayerCount == DVST_BANKERNUMBERS {
					//玩家已经下庄
					params = append(params, 2)
					opRetCode = dragonvstiger.OpResultCode_OPRC_Error
				} else {
					//下庄成功
					params = append(params, 1)
				}
				sceneEx.upplayerCount = DVST_BANKERNUMBERS
				this.SendSCPlayerOp(s, p, playerEx.Pos, opcode, opRetCode, params, false)
				return true
			case DragonVsTigerPlayerOpUpList: //上庄列表
				down := 0
				if len(sceneEx.upplayerlist) > 0 {
					for _, v := range sceneEx.upplayerlist {
						if v.SnId == playerEx.SnId {
							down = 1
							break
						}
					}
				}

				pack := &dragonvstiger.SCDragonVsTiggerUpList{
					Count:   proto.Int(len(sceneEx.upplayerlist)),
					IsExist: proto.Int(down),
				}
				for _, p := range sceneEx.upplayerlist {
					pd := &dragonvstiger.DragonVsTigerPlayerData{
						SnId:        proto.Int32(p.SnId),
						Name:        proto.String(p.Name),
						Head:        proto.Int32(p.Head),
						Sex:         proto.Int32(p.Sex),
						Coin:        proto.Int64(p.Coin),
						Pos:         proto.Int(p.Pos),
						Flag:        proto.Int(p.GetFlag()),
						City:        proto.String(p.GetCity()),
						HeadOutLine: proto.Int32(p.HeadOutLine),
						VIP:         proto.Int32(p.VIP),
					}
					pack.Data = append(pack.Data, pd)
				}
				proto.SetDefaults(pack)
				p.SendToClient(int(dragonvstiger.DragonVsTigerPacketID_PACKET_SC_DVST_UPLIST), pack)
			}
		}
	}
	return false
}
func (this *SceneBaseStateDragonVsTiger) OnPlayerEvent(s *base.Scene, p *base.Player, evtcode int, params []int64) {
	if sceneEx, ok := s.ExtraData.(*DragonVsTigerSceneData); ok {
		if playerEx, ok := p.ExtraData.(*DragonVsTigerPlayerData); ok {
			needResort := false
			switch evtcode {
			case base.PlayerEventEnter:
				if sceneEx.winTop1 == nil {
					needResort = true
				}
				if !needResort {
					for i := 0; i < DVST_RICHTOP5; i++ {
						if sceneEx.betTop5[i] == nil {
							needResort = true
							break
						}
					}
				}
			case base.PlayerEventLeave:
				if playerEx == sceneEx.winTop1 {
					needResort = true
				}
				if !needResort {
					for i := 0; i < DVST_RICHTOP5; i++ {
						if sceneEx.betTop5[i] == playerEx {
							needResort = true
							break
						}
					}
				}
				//正在做庄，将坐庄计数至为最大
				if sceneEx.bankerSnId != -1 && playerEx.SnId == sceneEx.bankerSnId {
					sceneEx.upplayerCount = MaxBankerNum
				}
				//在上庄列表中，从列表中剔除
				if len(sceneEx.upplayerlist) > 0 {
					for index, v := range sceneEx.upplayerlist {
						if v.SnId == playerEx.SnId {
							sceneEx.upplayerlist = append(sceneEx.upplayerlist[:index], sceneEx.upplayerlist[index+1:]...)
							sceneEx.Broadcast(int(dragonvstiger.DragonVsTigerPacketID_PACKET_SC_DVST_UPLIST), &dragonvstiger.SCDragonVsTiggerUpList{
								Count:   proto.Int(len(sceneEx.upplayerlist)),
								IsExist: proto.Int32(-1),
							}, 0)
							break
						}
					}
				}
			case base.PlayerEventRecharge:
				//oldflag := p.MarkBroadcast(playerEx.Pos < DVST_SELFPOS)
				p.AddCoin(params[0], common.GainWay_Pay, base.SyncFlag_ToClient, "system", p.GetScene().GetSceneName())
				//p.MarkBroadcast(oldflag)
				if p.Coin >= int64(sceneEx.DbGameFree.GetBetLimit()) && p.Coin >= int64(sceneEx.DbGameFree.GetOtherIntParams()[0]) {
					p.UnmarkFlag(base.PlayerState_GameBreak)
					p.SyncFlag(true)
				}
			}
			if needResort {
				seatKey := sceneEx.Resort()
				if seatKey != sceneEx.constSeatKey {
					dragonVsTigerSendSeatInfo(s, sceneEx)
					sceneEx.constSeatKey = seatKey
				}
			}
		}
	}
}

type SceneStakeAntStateDragonVsTiger struct {
	SceneBaseStateDragonVsTiger
}

func (this *SceneStakeAntStateDragonVsTiger) GetState() int {
	return DragonVsTigerSceneStateStakeAnt
}
func (this *SceneStakeAntStateDragonVsTiger) CanChangeTo(s base.SceneState) bool {
	if s.GetState() == DragonVsTigerSceneStateStake {
		return true
	}
	return false
}
func (this *SceneStakeAntStateDragonVsTiger) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if this.SceneBaseStateDragonVsTiger.OnPlayerOp(s, p, opcode, params) {
		return true
	}
	return true
}
func (this *SceneStakeAntStateDragonVsTiger) OnEnter(s *base.Scene) {
	this.SceneBaseStateDragonVsTiger.OnEnter(s)
	if sceneEx, ok := s.ExtraData.(*DragonVsTigerSceneData); ok {
		sceneEx.Clean()
		sceneEx.poker.Shuffle()
		sceneEx.NumOfGames++
		sceneEx.GameNowTime = time.Now()
		logger.Logger.Tracef("(this *Scene) [%v] 场景状态进入 %v", s.SceneId, len(sceneEx.seats))
		pack := &dragonvstiger.SCDragonVsTigerRoomState{
			State: proto.Int(this.GetState()),
		}
		proto.SetDefaults(pack)
		s.Broadcast(int(dragonvstiger.DragonVsTigerPacketID_PACKET_SC_DVST_ROOMSTATE), pack, 0)
	}
}
func (this *SceneStakeAntStateDragonVsTiger) OnLeave(s *base.Scene) {
}
func (this *SceneStakeAntStateDragonVsTiger) OnTick(s *base.Scene) {
	this.SceneBaseStateDragonVsTiger.OnTick(s)
	if sceneEx, ok := s.ExtraData.(*DragonVsTigerSceneData); ok {
		if time.Now().Sub(sceneEx.StateStartTime) > DragonVsTigerStakeAntTimeout {
			s.ChangeSceneState(DragonVsTigerSceneStateStake)
		}
	}
}

type SceneStakeStateDragonVsTiger struct {
	SceneBaseStateDragonVsTiger
}

func (this *SceneStakeStateDragonVsTiger) GetState() int {
	return DragonVsTigerSceneStateStake
}
func (this *SceneStakeStateDragonVsTiger) CanChangeTo(s base.SceneState) bool {
	switch s.GetState() {
	case DragonVsTigerSceneStateOpenCardAnt:
		return true
	}
	return false
}
func (this *SceneStakeStateDragonVsTiger) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if this.SceneBaseStateDragonVsTiger.OnPlayerOp(s, p, opcode, params) {
		return true
	}
	if playerEx, ok := p.ExtraData.(*DragonVsTigerPlayerData); ok {
		if sceneEx, ok := s.ExtraData.(*DragonVsTigerSceneData); ok {

			switch opcode {
			case DragonVsTigerPlayerOpBet:
				if len(params) >= 2 {
					if playerEx.SnId == sceneEx.bankerSnId { //庄家不能下注
						this.SendSCPlayerOp(s, p, playerEx.Pos, opcode, dragonvstiger.OpResultCode_OPRC_BankerCannotBet, params, false)
						return false
					}
					betPos := int(params[0])
					if sceneEx.winTop1 != nil && playerEx.IsRob {
						if sceneEx.winTop1.lastBetPos == -1 {
							sceneEx.winTop1.lastBetPos = betPos
						}
						if sceneEx.winTop1.IsRob && sceneEx.winTop1.lastBetPos != betPos && playerEx.lastBetPos != -1 {
							return false
						}
					}
					if betPos < 0 || betPos >= DVST_ZONE_MAX {
						return false
					}
					betCoin := int(params[1])
					chips := sceneEx.DbGameFree.GetOtherIntParams()
					if !common.InSliceInt32(chips, int32(params[1])) {
						return false
					}
					if p.Coin < int64(sceneEx.DbGameFree.GetBetLimit()) && playerEx.betTotal == 0 {
						logger.Logger.Trace("======提示低于多少不能下注======")
						this.SendSCPlayerOp(s, p, playerEx.Pos, opcode, dragonvstiger.OpResultCode_OPRC_CoinMustReachTheValue, params, false)
						return false
					}
					//最小面额的筹码
					minChip := int(chips[0])
					if sceneEx.bankerSnId != -1 {
						banker := sceneEx.players[sceneEx.bankerSnId]
						if banker != nil {
							betCoin = sceneEx.Bet(banker, playerEx, betPos, betCoin)
							if betCoin <= 0 { //不能再继续下注了
								params = append(params, int64(sceneEx.betInfo[betPos]))
								params = append(params, 6)
								params = append(params, playerEx.Coin-playerEx.betTotal)
								this.SendSCPlayerOp(s, p, playerEx.Pos, opcode, dragonvstiger.OpResultCode_OPRC_EachBetsLimit, params, false)
								//切换到开牌阶段
								if sceneEx.BetStop() {
									msgPack := common.CreateSrvMsg(common.SRVMSG_CODE_DEFAULT, "庄家可押注额度满，直接开牌")
									s.Broadcast(int(player.PlayerPacketID_PACKET_SC_SRVMSG), msgPack, p.GetSid())
									s.ChangeSceneState(DragonVsTigerSceneStateOpenCardAnt)
								} else {
									msgPack := common.CreateSrvMsg(common.SRVMSG_CODE_DEFAULT, "该区域押注额度满")
									p.SendToClient(int(player.PlayerPacketID_PACKET_SC_SRVMSG), msgPack)
								}
								return false
							}
						}
					}
					//期望下注金额
					expectBetCoin := betCoin
					total := int64(betCoin) + playerEx.betTotal
					if total <= playerEx.Coin {
						maxBetCoin := sceneEx.DbGameFree.GetMaxBetCoin()
						if ok, coinLimit := playerEx.MaxChipCheck(betPos, betCoin, maxBetCoin); !ok {
							betCoin = int(coinLimit) - playerEx.betInfo[betPos]
							//对齐到最小面额的筹码
							betCoin /= minChip
							betCoin *= minChip
							if betCoin <= 0 {
								//this.SendSCPlayerOp(s, p, playerEx.Pos, opcode, dragonvstiger.OpResultCode_OPRC_Hundred_EachBetsLimit, params, false)
								msgPack := common.CreateSrvMsg(common.SRVMSG_CODE_DEFAULT, fmt.Sprintf("该门押注金额上限%.2f",
									float64(coinLimit)/100))
								p.SendToClient(int(player.PlayerPacketID_PACKET_SC_SRVMSG), msgPack)
								return false
							}
						}
					} else {
						this.SendSCPlayerOp(s, p, playerEx.Pos, opcode, dragonvstiger.OpResultCode_OPRC_CoinIsNotEnough, params, false)
						return false
					}
					playerEx.Trusteeship = 0
					playerEx.TotalBet += int64(betCoin)

					playerEx.betTotal += int64(betCoin)
					playerEx.betInfo[betPos] += betCoin
					sceneEx.betInfo[betPos] += betCoin
					if playerEx.IsRob {
						sceneEx.betInfoRob[betPos] += betCoin
					}
					if betCoin == expectBetCoin {
						sceneEx.betDetailInfo[betPos][betCoin] = sceneEx.betDetailInfo[betPos][betCoin] + 1
						playerEx.betDetailInfo[betPos][betCoin] = playerEx.betDetailInfo[betPos][betCoin] + 1
						playerEx.betCacheInfo[betPos][betCoin] = playerEx.betCacheInfo[betPos][betCoin] + 1
					} else { //拆分筹码
						val := betCoin
						for i := len(chips) - 1; i >= 0 && val > 0; i-- {
							chip := int(chips[i])
							cntChip := val / chip
							if cntChip > 0 {
								sceneEx.betDetailInfo[betPos][betCoin] = sceneEx.betDetailInfo[betPos][chip] + cntChip
								playerEx.betDetailInfo[betPos][betCoin] = playerEx.betDetailInfo[betPos][chip] + cntChip
								playerEx.betCacheInfo[betPos][betCoin] = playerEx.betCacheInfo[betPos][chip] + cntChip
								val -= cntChip * chip
							}
						}
					}
					restCoin := playerEx.Coin - playerEx.betTotal
					params[1] = int64(betCoin) //修正下当前的实际下注额
					params = append(params, restCoin)

					this.SendSCPlayerOp(s, p, playerEx.Pos, opcode, dragonvstiger.OpResultCode_OPRC_Sucess, params, playerEx.Pos < DVST_SELFPOS)

					playerEx.betDetailOrderInfo[betPos] = append(playerEx.betDetailOrderInfo[betPos], int64(betCoin))

					if restCoin < int64(sceneEx.DbGameFree.GetBetLimit()) || restCoin < int64(chips[0]) {
						playerEx.MarkFlag(base.PlayerState_GameBreak)
						playerEx.SyncFlag(true)
					}
				}
				return true
			}
		}
	}
	return false
}
func (this *SceneStakeStateDragonVsTiger) OnEnter(s *base.Scene) {
	this.SceneBaseStateDragonVsTiger.OnEnter(s)
	if sceneEx, ok := s.ExtraData.(*DragonVsTigerSceneData); ok {
		sceneEx.CheckResults()
		pack := &dragonvstiger.SCDragonVsTigerRoomState{
			State: proto.Int(this.GetState()),
		}
		proto.SetDefaults(pack)
		s.Broadcast(int(dragonvstiger.DragonVsTigerPacketID_PACKET_SC_DVST_ROOMSTATE), pack, 0)
	}
}
func (this *SceneStakeStateDragonVsTiger) OnLeave(s *base.Scene) {
	this.SceneBaseStateDragonVsTiger.OnLeave(s)
	if sceneEx, ok := s.ExtraData.(*DragonVsTigerSceneData); ok {
		DragonVsTigerBatchSendBet(sceneEx, true)
		/*betCoin := 0
		for i := 0; i < DVST_ZONE_MAX; i++ {
			betCoin += sceneEx.betInfo[i] - sceneEx.betInfoRob[i]
		}
		base.CoinPoolMgr.PushCoin(sceneEx.GetGameFreeId(), sceneEx.platform, int64(betCoin))*/
		for _, v := range sceneEx.players {
			if v != nil /*&& !v.IsRob*/ {
				if v.SnId == sceneEx.bankerSnId {
					v.Trusteeship = 0
				} else if v.betTotal <= 0 {
					v.Trusteeship++
					if v.Trusteeship >= model.GameParamData.NotifyPlayerWatchNum && !v.IsRob {
						v.SendTrusteeshipTips()
					}
				}
			}
		}
	}
}
func (this *SceneStakeStateDragonVsTiger) OnTick(s *base.Scene) {
	this.SceneBaseStateDragonVsTiger.OnTick(s)
	if sceneEx, ok := s.ExtraData.(*DragonVsTigerSceneData); ok {
		if time.Now().Sub(sceneEx.StateStartTime) > DragonVsTigerStakeTimeout {
			s.ChangeSceneState(DragonVsTigerSceneStateOpenCardAnt)
		}
	}
}

type SceneOpenCardAntStateDragonVsTiger struct {
	SceneBaseStateDragonVsTiger
}

func (this *SceneOpenCardAntStateDragonVsTiger) GetState() int {
	return DragonVsTigerSceneStateOpenCardAnt
}
func (this *SceneOpenCardAntStateDragonVsTiger) CanChangeTo(s base.SceneState) bool {
	switch s.GetState() {
	case DragonVsTigerSceneStateOpenCard:
		return true
	}
	return false
}
func (this *SceneOpenCardAntStateDragonVsTiger) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if this.SceneBaseStateDragonVsTiger.OnPlayerOp(s, p, opcode, params) {
		return true
	}
	return true
}
func (this *SceneOpenCardAntStateDragonVsTiger) OnEnter(s *base.Scene) {
	this.SceneBaseStateDragonVsTiger.OnEnter(s)
	if _, ok := s.ExtraData.(*DragonVsTigerSceneData); ok {
		pack := &dragonvstiger.SCDragonVsTigerRoomState{
			State: proto.Int(this.GetState()),
		}
		proto.SetDefaults(pack)
		s.Broadcast(int(dragonvstiger.DragonVsTigerPacketID_PACKET_SC_DVST_ROOMSTATE), pack, 0)
	}
}
func (this *SceneOpenCardAntStateDragonVsTiger) OnLeave(s *base.Scene) {}
func (this *SceneOpenCardAntStateDragonVsTiger) OnTick(s *base.Scene) {
	this.SceneBaseStateDragonVsTiger.OnTick(s)
	if sceneEx, ok := s.ExtraData.(*DragonVsTigerSceneData); ok {
		if time.Now().Sub(sceneEx.StateStartTime) > DragonVsTigerOpenCardAntTimeout {
			s.ChangeSceneState(DragonVsTigerSceneStateOpenCard)
		}
	}
}

type SceneOpenCardStateDragonVsTiger struct {
	SceneBaseStateDragonVsTiger
}

func (this *SceneOpenCardStateDragonVsTiger) GetState() int {
	return DragonVsTigerSceneStateOpenCard
}
func (this *SceneOpenCardStateDragonVsTiger) CanChangeTo(s base.SceneState) bool {
	switch s.GetState() {
	case DragonVsTigerSceneStateBilled:
		return true
	}
	return false
}
func (this *SceneOpenCardStateDragonVsTiger) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if this.SceneBaseStateDragonVsTiger.OnPlayerOp(s, p, opcode, params) {
		return true
	}
	return true
}
func (this *SceneOpenCardStateDragonVsTiger) OnEnter(s *base.Scene) {
	this.SceneBaseStateDragonVsTiger.OnEnter(s)
	if sceneEx, ok := s.ExtraData.(*DragonVsTigerSceneData); ok {
		sceneEx.CalcAllAvgBet()
		if !sceneEx.bIntervention {
			//for i := 0; i < 2; i++ {
			//	sceneEx.cards[i] = int32(sceneEx.poker.Next())
			//}
			sceneEx.SendCards()

			ok, singlePlayer := sceneEx.IsSingleRegulatePlayer()
			if ok { //单控
				if sceneEx.IsNeedSingleRegulate(singlePlayer) { //本轮结果是否需要单控
					if sceneEx.bankerSnId == singlePlayer.SnId {
						//走庄家调控
						if sceneEx.RegulationBankerCard(singlePlayer) {
							//singlePlayer.单控计数++
							singlePlayer.MarkFlag(base.PlayerState_SAdjust)
							singlePlayer.AddAdjustCount(sceneEx.GetGameFreeId())
						} else {
							singlePlayer.result = 0
						}
					} else {
						//走闲家调控
						if sceneEx.RegulationXianCard(singlePlayer) {
							//singlePlayer.单控计数++
							singlePlayer.MarkFlag(base.PlayerState_SAdjust)
							singlePlayer.AddAdjustCount(sceneEx.GetGameFreeId())
						} else {
							singlePlayer.result = 0
						}
					}
				} else {
					//singlePlayer.单控计数++
					if singlePlayer.betTotal != 0 || singlePlayer.SnId == sceneEx.bankerSnId {
						singlePlayer.MarkFlag(base.PlayerState_SAdjust)
						singlePlayer.AddAdjustCount(sceneEx.GetGameFreeId())
					} else {
						singlePlayer.result = 0
					}
				}

			} else {
				var banker, isRobotBanker = sceneEx.GetBanker()
				if banker == nil || isRobotBanker == true { //非真人坐庄
					if !sceneEx.AutoBalance() {
						//关闭强制修正路单,被玩家抓住规律了,屏蔽测试后，发现效果不理想，不符合玩家预期，导致数据下滑，目前打开
						//result := sceneEx.CalcuResult()
						result := sceneEx.CalcuResult()
						//暂时屏蔽路单纠正，增加一个杀概率，保证水池会慢慢增长

						//sceneEx.ConsecutiveCheck(result)
						//sceneEx.AutoBalance2()
						//sceneEx.ChangeCardByRand(result)
						result = sceneEx.CalcuResult()
						sceneEx.ChangeCard(result)
					}
					//新算法
					sceneEx.AutoBalance3()
				} else {
					//真人坐庄，调控
					sceneEx.BankerBalance()
					//新算法
					sceneEx.AutoBalanceBank3()
				}
			}

			sceneEx.CalcuResult()
		}
		if sceneEx.cards[0] == -1 || sceneEx.cards[1] == -1 {
			logger.Logger.Errorf("DragonVsTiger roll error card data:%v-%v", sceneEx.cards[0], sceneEx.cards[0])
			//sceneEx.poker.Shuffle()
			//for i := 0; i < 2; i++ {
			//	sceneEx.cards[i] = int32(sceneEx.poker.Next())
			//}
			sceneEx.SendCards()
		}

		result := sceneEx.CalcuResult()
		sceneEx.PushTrend(int32(result))
		pack := &dragonvstiger.SCDragonVsTigerRoomState{
			State:  proto.Int(this.GetState()),
			Params: common.CopySliceInt32(sceneEx.cards[:]),
		}
		pack.Params = append(pack.Params, int32(result))
		proto.SetDefaults(pack)
		s.Broadcast(int(dragonvstiger.DragonVsTigerPacketID_PACKET_SC_DVST_ROOMSTATE), pack, 0)
	}
}
func (this *SceneOpenCardStateDragonVsTiger) OnLeave(s *base.Scene) {
	this.SceneBaseStateDragonVsTiger.OnLeave(s)
	if sceneEx, ok := s.ExtraData.(*DragonVsTigerSceneData); ok {
		//同步开奖结果
		//0:和 1:龙 2:虎
		result := sceneEx.CalcuResult()
		pack := &server.GWGameStateLog{
			SceneId: proto.Int(s.SceneId),
			GameLog: proto.Int(result),
			LogCnt:  proto.Int(len(sceneEx.trend100Cur)),
		}
		s.SendToWorld(int(server.SSPacketID_PACKET_GW_GAMESTATELOG), pack)
	}
}
func (this *SceneOpenCardStateDragonVsTiger) OnTick(s *base.Scene) {
	this.SceneBaseStateDragonVsTiger.OnTick(s)
	if sceneEx, ok := s.ExtraData.(*DragonVsTigerSceneData); ok {
		if time.Now().Sub(sceneEx.StateStartTime) > DragonVsTigerOpenCardTimeout {
			s.ChangeSceneState(DragonVsTigerSceneStateBilled)
		}
	}
}

type SceneBilledStateDragonVsTiger struct {
	SceneBaseStateDragonVsTiger
}

func (this *SceneBilledStateDragonVsTiger) GetState() int {
	return DragonVsTigerSceneStateBilled
}
func (this *SceneBilledStateDragonVsTiger) CanChangeTo(s base.SceneState) bool {
	switch s.GetState() {
	case DragonVsTigerSceneStateStakeAnt:
		return true
	}
	return false
}
func (this *SceneBilledStateDragonVsTiger) CanChangeCoinScene(s *base.Scene, p *base.Player) bool {
	return true
}
func (this *SceneBilledStateDragonVsTiger) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if this.SceneBaseStateDragonVsTiger.OnPlayerOp(s, p, opcode, params) {
		return true
	}
	return true
}
func (this *SceneBilledStateDragonVsTiger) OnEnter(s *base.Scene) {
	this.SceneBaseStateDragonVsTiger.OnEnter(s)
	if sceneEx, ok := s.ExtraData.(*DragonVsTigerSceneData); ok {
		// 校验后台调控是否正确
		if !sceneEx.Testing {
			//sceneEx.resultCheck()
			res := sceneEx.GetResult()
			history := sceneEx.GetResultHistoryResult(sceneEx.LoopNum)
			if res != 0 || history != -1 {
				logger.Logger.Tracef("局数:%d history:%d res:%d bIntervention:%v",
					sceneEx.LoopNum, history, res, sceneEx.bIntervention)
				// 调控结果是否和记录中的结果一样
				if res != history {
					logger.Logger.Errorf("调控历史与调控结果不同 history:%d res:%d bIntervention:%v",
						history, res, sceneEx.bIntervention)
				}
				// 调控是否成功
				switch sceneEx.CalcuResult() {
				case DVST_ZONE_DRAW:
					if res != 3 {
						logger.Logger.Errorf("控%d失败 %d,%v", res, DVST_ZONE_DRAW, sceneEx.bIntervention)
					}
				case DVST_ZONE_DRAGON:
					if res != 1 {
						logger.Logger.Errorf("控%d失败 %d,%v", res, DVST_ZONE_DRAGON, sceneEx.bIntervention)
					}
				case DVST_ZONE_TIGER:
					if res != 2 {
						logger.Logger.Errorf("控%d失败 %d,%v", res, DVST_ZONE_TIGER, sceneEx.bIntervention)
					}
				}
			}
		}
		sceneEx.AddLoopNum()
		pack := &dragonvstiger.SCDragonVsTigerRoomState{
			State: proto.Int(this.GetState()),
		}
		proto.SetDefaults(pack)
		s.Broadcast(int(dragonvstiger.DragonVsTigerPacketID_PACKET_SC_DVST_ROOMSTATE), pack, 0)
		playerTotalGain := int64(0)
		sysGain := int64(0) //
		olTotalGain := int64(0)
		olTotalBet := 0
		result := sceneEx.CalcuResult()

		// 水池上下文环境
		sceneEx.CpCtx = base.CoinPoolMgr.GetCoinPoolCtx(sceneEx.Platform, sceneEx.GetGameFreeId(), sceneEx.GroupId)
		sceneEx.CpCtx.Controlled = sceneEx.CpControlled

		var bigWinner *DragonVsTigerPlayerData
		countCoin := [DVST_ZONE_MAX]int64{0, 0, 0}
		count := len(sceneEx.seats)
		for i := 0; i < count; i++ {
			playerEx := sceneEx.seats[i]
			if playerEx != nil {
				if sceneEx.bankerSnId == playerEx.SnId {
					continue
				}

				var tax int64
				var lostCoin int32 // 玩家输分
				var winCoin int32  // 玩家赢分
				playerEx.gainCoin = 0
				sceneEx.isWin = [3]int{-1, -1, -1}
				switch result {
				case DVST_ZONE_TIGER, DVST_ZONE_DRAGON:
					for j := 0; j < DVST_ZONE_MAX; j++ {
						if playerEx.betInfo[j] > 0 {
							if j == result {
								playerEx.winCoin[j] = (DVST_TheOdds[j] + 1) * playerEx.betInfo[j]
								playerEx.gainCoin = int64(playerEx.winCoin[j])
								tax = int64(DVST_TheOdds[j]*playerEx.betInfo[j]) * int64(s.DbGameFree.GetTaxRate()) / 10000
								winCoin += int32(DVST_TheOdds[j] * playerEx.betInfo[j])
								countCoin[j] -= int64(DVST_TheOdds[j] * playerEx.betInfo[j])
							} else {
								lostCoin += int32(playerEx.betInfo[j])
								countCoin[j] += int64(playerEx.betInfo[j])
							}
						}
					}
					sceneEx.isWin[result] = 1
				case DVST_ZONE_DRAW:
					for j := 0; j < DVST_ZONE_MAX; j++ {
						if playerEx.betInfo[j] > 0 {
							if j != result {
								//20190426:开和不扣钱
								playerEx.winCoin[j] += playerEx.betInfo[j]
								playerEx.gainCoin += int64(playerEx.winCoin[j])
							} else {
								playerEx.winCoin[j] = (DVST_TheOdds[j] + 1) * playerEx.betInfo[j]
								playerEx.gainCoin = int64(playerEx.winCoin[j])
								tax = int64(DVST_TheOdds[j]*playerEx.betInfo[j]) * int64(s.DbGameFree.GetTaxRate()) / 10000
								winCoin += int32(DVST_TheOdds[j] * playerEx.betInfo[j])
								countCoin[j] -= int64(DVST_TheOdds[j] * playerEx.betInfo[j])
							}
						}
					}
					sceneEx.isWin = [3]int{1, 0, 0}
				}
				//计算庄家输赢
				sceneEx.bankerWinCoin -= int64(winCoin)
				sceneEx.bankerWinCoin += int64(lostCoin)

				playerEx.taxCoin = tax
				playerEx.winorloseCoin = playerEx.gainCoin
				playerEx.gainWinLost = playerEx.gainCoin - tax - playerEx.betTotal
				if playerEx.gainWinLost > 0 {

					playerEx.winRecord = append(playerEx.winRecord, 1)
				} else {
					playerEx.winRecord = append(playerEx.winRecord, 0)
				}
				sysGain += playerEx.gainWinLost
				if playerEx.Pos == DVST_OLPOS {
					olTotalGain += playerEx.gainWinLost
				}
				playerEx.betBigRecord = append(playerEx.betBigRecord, playerEx.betTotal)
				playerEx.RecalcuLatestBet20()
				if playerEx.betTotal > 0 {
					playerEx.GameTimes++
					if playerEx.gainWinLost > 0 {
						playerEx.WinTimes++
					} else {
						playerEx.FailTimes++
					}
				}
				playerEx.RecalcuLatestWin20()
				if playerEx.gainWinLost > 0 {
					//oldflag := playerEx.MarkBroadcast(playerEx.Pos < DVST_SELFPOS)
					playerEx.AddCoin(playerEx.gainWinLost, common.GainWay_HundredSceneWin, base.SyncFlag_ToClient, "system", s.GetSceneName())
					//playerEx.MarkBroadcast(oldflag)

				} else {
					//oldflag := playerEx.MarkBroadcast(playerEx.Pos < DVST_SELFPOS)
					playerEx.AddCoin(playerEx.gainWinLost, common.GainWay_HundredSceneLost, base.SyncFlag_ToClient, "system", s.GetSceneName())
					//playerEx.MarkBroadcast(oldflag)
				}

				if playerEx.betTotal != 0 {
					playerEx.AddServiceFee(tax)
					//playerEx.ReportGameEvent(playerEx.gainWinLost, tax, playerEx.betTotal)
				}
				if playerEx.Coin >= int64(sceneEx.DbGameFree.GetBetLimit()) && playerEx.Coin >= int64(sceneEx.DbGameFree.GetOtherIntParams()[0]) {
					if playerEx.IsMarkFlag(base.PlayerState_GameBreak) {
						playerEx.UnmarkFlag(base.PlayerState_GameBreak)
						playerEx.SyncFlag(true)
					}
				} else {
					if !playerEx.IsMarkFlag(base.PlayerState_GameBreak) {
						playerEx.MarkFlag(base.PlayerState_GameBreak)
						playerEx.SyncFlag(true)
					}
				}

				if playerEx.gainWinLost > 0 {
					if bigWinner == nil {
						bigWinner = playerEx
					} else {
						if bigWinner.gainWinLost < playerEx.gainWinLost {
							bigWinner = playerEx
						}
					}
				}

				if !playerEx.IsRob {

					playerTotalGain += int64(playerEx.gainWinLost)
				}
				if playerEx.GetCurrentCoin() == 0 {
					playerEx.SetCurrentCoin(playerEx.GetTakeCoin())
				}
				if playerEx.betTotal > 0 {
					playerEx.SaveSceneCoinLog(playerEx.GetCurrentCoin(), playerEx.Coin-playerEx.GetCurrentCoin(),
						playerEx.Coin, playerEx.betTotal, playerEx.taxCoin, playerEx.winorloseCoin, 0, 0)
				}

				if playerEx.gainWinLost != 0 {
					playerEx.Statics(sceneEx.KeyGameId, sceneEx.KeyGamefreeId, int64(playerEx.gainWinLost+playerEx.taxCoin), true)
				}
			}
		}

		//庄家结算
		//sceneEx.bankerWinCoin = -sceneEx.bankerWinCoin
		banker := sceneEx.players[sceneEx.bankerSnId]
		if sceneEx.bankerSnId != -1 && banker != nil {
			logger.Logger.Tracef("庄家收益 %v", sceneEx.bankerWinCoin)
			for i := 0; i < DVST_ZONE_MAX; i++ {
				banker.winCoin[i] = int(countCoin[i])
			}
			banker.gainCoin = sceneEx.bankerWinCoin

			//庄家输赢记录
			if sceneEx.bankerWinCoin > 0 {
				banker.winRecord = append(banker.winRecord, 1)
				if sceneEx.bankerWinCoin > banker.Coin {
					banker.gainCoin = banker.Coin
				} else {
					banker.gainCoin = sceneEx.bankerWinCoin
				}
			} else if sceneEx.bankerWinCoin <= 0 {
				banker.winRecord = append(banker.winRecord, 0)
				if -sceneEx.bankerWinCoin > banker.Coin {
					banker.gainCoin = -banker.Coin
				} else {
					banker.gainCoin = sceneEx.bankerWinCoin
				}
			}

			//庄家金币变化
			if banker.gainCoin > 0 {
				tax := banker.gainCoin * int64(s.DbGameFree.GetTaxRate()) / 10000
				banker.taxCoin = tax
				banker.winorloseCoin = banker.gainCoin
				banker.gainWinLost = banker.gainCoin - tax
				banker.AddCoin(banker.gainWinLost, common.GainWay_HundredSceneWin, base.SyncFlag_ToClient, "system", s.GetSceneName())
			} else if banker.gainCoin < 0 {
				banker.winorloseCoin = banker.gainCoin
				banker.gainWinLost = banker.gainCoin
				banker.AddCoin(banker.gainCoin, common.GainWay_HundredSceneLost, base.SyncFlag_ToClient, "system", s.GetSceneName())
			}
		}

		var constBills []*dragonvstiger.DragonVsTigerBill
		if sceneEx.winTop1 != nil && sceneEx.winTop1.betTotal != 0 {
			constBills = append(constBills, &dragonvstiger.DragonVsTigerBill{
				SnId:     proto.Int32(sceneEx.winTop1.SnId),
				Coin:     proto.Int64(sceneEx.winTop1.Coin),
				GainCoin: proto.Int64(sceneEx.winTop1.gainWinLost),
			})
		}
		if olTotalBet != 0 {
			constBills = append(constBills, &dragonvstiger.DragonVsTigerBill{
				SnId:     proto.Int(0),
				Coin:     proto.Int64(olTotalGain),
				GainCoin: proto.Int64(olTotalGain),
			})
		} else {
			if sysGain < 0 {
				constBills = append(constBills, &dragonvstiger.DragonVsTigerBill{
					SnId:     proto.Int(0),
					Coin:     proto.Int64(sysGain * -1),
					GainCoin: proto.Int64(sysGain * -1),
				})
			}
		}
		for i := 0; i < DVST_RICHTOP5; i++ {
			if sceneEx.betTop5[i] != nil && sceneEx.betTop5[i].betTotal != 0 {
				constBills = append(constBills, &dragonvstiger.DragonVsTigerBill{
					SnId:     proto.Int32(sceneEx.betTop5[i].SnId),
					Coin:     proto.Int64(sceneEx.betTop5[i].Coin),
					GainCoin: proto.Int64(sceneEx.betTop5[i].gainWinLost),
				})
			}
		}
		for i := 0; i < count; i++ {
			playerEx := sceneEx.seats[i]
			if playerEx.IsOnLine() && !playerEx.IsMarkFlag(base.PlayerState_Leave) {
				pack := &dragonvstiger.SCDragonVsTigerBilled{
					LoopNum: proto.Int(sceneEx.LoopNum),
				}
				pack.BillData = append(pack.BillData, constBills...)
				if playerEx.betTotal != 0 {
					pack.BillData = append(pack.BillData, &dragonvstiger.DragonVsTigerBill{
						SnId:     proto.Int32(playerEx.SnId),
						Coin:     proto.Int64(playerEx.Coin),
						GainCoin: proto.Int64(playerEx.gainWinLost),
					})
				}

				//庄家输赢数据
				if sceneEx.bankerSnId != -1 && banker != nil {
					pack.BillData = append(pack.BillData, &dragonvstiger.DragonVsTigerBill{
						SnId:     proto.Int32(banker.SnId),
						Coin:     proto.Int64(banker.Coin),
						GainCoin: proto.Int64(banker.gainWinLost),
					})
				}

				if bigWinner != nil {
					pack.BigWinner = &dragonvstiger.DragonVsTigerBigWinner{
						SnId:        proto.Int32(bigWinner.SnId),
						Name:        proto.String(bigWinner.Name),
						Head:        proto.Int32(bigWinner.Head),
						HeadOutLine: proto.Int32(bigWinner.HeadOutLine),
						VIP:         proto.Int32(bigWinner.VIP),
						Sex:         proto.Int32(bigWinner.Sex),
						Coin:        proto.Int64(bigWinner.Coin),
						GainCoin:    proto.Int32(int32(bigWinner.gainWinLost)),
						City:        proto.String(bigWinner.GetCity()),
					}
				}
				proto.SetDefaults(pack)
				playerEx.SendToClient(int(dragonvstiger.DragonVsTigerPacketID_PACKET_SC_DVST_GAMEBILLED), pack)
			}
		}

		if !sceneEx.Testing {
			var playerCtxs []*server.PlayerCtx
			for _, p := range sceneEx.seats {
				if p == nil || p.IsRob || p.betTotal == 0 {
					continue
				}
				playerCtxs = append(playerCtxs, &server.PlayerCtx{SnId: proto.Int32(p.SnId), Coin: proto.Int64(p.Coin)})
			}
			if len(playerCtxs) > 0 {
				//pack := &server.GWSceneEnd{
				//	GameFreeId: proto.Int32(sceneEx.DbGameFree.GetId()),
				//	Players:    playerCtxs,
				//}
				//proto.SetDefaults(pack)
				//sceneEx.SendToWorld(int(server.MmoPacketID_PACKET_GW_SCENEEND), pack)
			}
		}

		s.SystemCoinOut = sceneEx.GetSystemChangeCoin(result)

		keyPf := fmt.Sprintf("%v_%v", sceneEx.Platform, sceneEx.GetGameFreeId())
		if s.SystemCoinOut > 0 {
			base.CoinPoolMgr.PushCoin(sceneEx.GetGameFreeId(), sceneEx.GroupId, sceneEx.Platform, s.SystemCoinOut)
			base.SysProfitCoinMgr.Add(keyPf, s.SystemCoinOut, 0)
		} else if s.SystemCoinOut < 0 {
			base.CoinPoolMgr.PopCoin(sceneEx.GetGameFreeId(), sceneEx.GroupId, sceneEx.Platform, -s.SystemCoinOut)
			base.SysProfitCoinMgr.Add(keyPf, 0, -s.SystemCoinOut)
		}

		if !sceneEx.Testing && sceneEx.bIntervention {
			//curCoin := base.CoinPoolMgr.LoadCoin(sceneEx.GetGameFreeId(), sceneEx.Platform, sceneEx.GroupId)
			playerData := make(map[string]int64)
			for _, value := range sceneEx.players {
				if value.IsRob {
					continue
				}
				playerData[strconv.Itoa(int(value.SnId))] = value.gainWinLost
			}
			//task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
			//	model.NewDtInterventionLog(sceneEx.webUser, result, sysGain, playerData, result, curCoin)
			//	return nil
			//}), nil, "NewDtInterventionLog").Start()
		}

		if sceneEx.bankerSnId != -1 && banker != nil {
			banker.winRecord = []int64{}
			banker.betBigRecord = []int64{}
			banker.lately20Bet = 0
			banker.lately20Win = 0
			//连庄次数+1
			sceneEx.upplayerCount++
		}
		/////////////////////////////////////统计牌局详细记录///////////////////
		hundredType := make([]model.HundredType, DVST_ZONE_MAX, DVST_ZONE_MAX)
		var isSave bool //有真人参与保存牌局
		if sceneEx.bankerSnId != -1 && banker != nil {
			//目前只有玩家可以上庄
			sceneEx.bankerbetCoin = 0
			for _, n := range sceneEx.players {
				if n != nil && n.betTotal > 0 {
					sceneEx.bankerbetCoin += n.betTotal
				}
			}
			banker.betTotal = sceneEx.bankerbetCoin
			if banker.betTotal > 0 && !banker.IsRob {
				isSave = true
			}
		}

		for i := 0; i < DVST_ZONE_MAX; i++ {
			hundredType[i] = model.HundredType{
				RegionId:  int32(i),
				IsWin:     sceneEx.isWin[i],
				CardsInfo: []int32{-1},
			}
			if i == 1 || i == 2 {
				hundredType[i].CardsKind = sceneEx.cards[i-1]
				hundredType[i].CardsInfo[0] = sceneEx.cards[i-1]
			}
			playNum, index := 0, 0
			for _, oPlayer := range sceneEx.players {
				if oPlayer != nil && oPlayer.betInfo[i] > 0 {
					if banker != nil && !banker.IsRob {
						//真人玩家坐庄时 需要统计机器人
						playNum++
						isSave = true
					} else if !oPlayer.IsRob {
						//非真人玩家坐庄，不统计机器人
						playNum++
						isSave = true
					}
				}

				//if o_player != nil && o_player.betInfo[i] > 0 && !o_player.IsRob {
				//	playNum++
				//	isSave = true
				//}
			}
			if playNum == 0 {
				hundredType[i].PlayerData = nil
				continue
			}
			hundredPersons := make([]model.HundredPerson, playNum, playNum)
			for _, oPlayer := range sceneEx.players {
				if oPlayer != nil && oPlayer.betInfo[i] > 0 {
					if oPlayer.winCoin[i] <= 0 {
						oPlayer.winCoin[i] = -oPlayer.betInfo[i]
					} else if oPlayer.winCoin[i] > 0 {
						oPlayer.winCoin[i] -= oPlayer.betInfo[i]
					}
					pe := model.HundredPerson{
						UserId:       oPlayer.SnId,
						UserBetTotal: int64(oPlayer.betInfo[i]),
						ChangeCoin:   int64(oPlayer.winCoin[i]),
						BeforeCoin:   oPlayer.GetCurrentCoin(),
						AfterCoin:    oPlayer.Coin,
						IsRob:        oPlayer.IsRob,
						IsFirst:      sceneEx.IsPlayerFirst(oPlayer.Player),
						WBLevel:      oPlayer.WBLevel,
						Result:       oPlayer.result,
					}
					betDetail, ok := oPlayer.betDetailOrderInfo[i]
					if ok {
						pe.UserBetTotalDetail = betDetail
					}
					if banker != nil && !banker.IsRob {
						hundredPersons[index] = pe
						index++
					} else if !oPlayer.IsRob {
						hundredPersons[index] = pe
						index++
					}
				}

				//if o_player != nil && o_player.betInfo[i] > 0 && !o_player.IsRob {
				//	if o_player.winCoin[i] <= 0 {
				//		o_player.winCoin[i] = -o_player.betInfo[i]
				//	} else if o_player.winCoin[i] > 0 {
				//		o_player.winCoin[i] -= o_player.betInfo[i]
				//	}
				//	pe := model.HundredPerson{
				//		UserId:       o_player.SnId,
				//		UserBetTotal: int64(o_player.betInfo[i]),
				//		ChangeCoin:   int64(o_player.winCoin[i]),
				//		BeforeCoin:   o_player.currentCoin,
				//		AfterCoin:    o_player.Coin,
				//		IsRob:        o_player.IsRob,
				//		IsFirst:      sceneEx.IsPlayerFirst(o_player.Player),
				//		WhiteLevel:   o_player.WhiteLevel,
				//		BlackLevel:   o_player.BlackLevel,
				//		Result:       o_player.result,
				//	}
				//	hundredPersons[index] = pe
				//	index++
				//}
			}
			hundredType[i].PlayerData = hundredPersons
		}

		//最后单独添加庄记录
		if banker != nil && banker.betTotal > 0 {
			win := -1
			if banker.gainCoin > 0 {
				win = 1
			}
			bankerHundredType := model.HundredType{
				RegionId:  int32(-1),
				IsWin:     win,
				CardsInfo: []int32{-1},
			}

			hundredPersons := make([]model.HundredPerson, 1, 1)
			pe := model.HundredPerson{
				UserId:       banker.SnId,
				UserBetTotal: int64(banker.betTotal),
				ChangeCoin:   int64(banker.gainCoin),
				BeforeCoin:   banker.GetCurrentCoin(),
				AfterCoin:    banker.Coin,
				IsRob:        banker.IsRob,
				IsFirst:      sceneEx.IsPlayerFirst(banker.Player),
				WBLevel:      banker.WBLevel,
				Result:       banker.result,
			}
			hundredPersons[0] = pe
			bankerHundredType.PlayerData = hundredPersons
			hundredType = append(hundredType, bankerHundredType)
		}

		if !sceneEx.Testing {
			gwPlayerBet := &server.GWPlayerBet{
				GameFreeId: proto.Int32(sceneEx.DbGameFree.GetId()),
				RobotGain:  proto.Int64(sceneEx.SystemCoinOut),
			}
			for _, p := range sceneEx.players {
				if !p.IsRob && p.betTotal > 0 {
					playerBet := &server.PlayerBet{
						SnId: proto.Int32(p.SnId),
						Bet:  proto.Int64(p.betTotal),
						Gain: proto.Int64(p.gainWinLost + p.taxCoin),
						Tax:  proto.Int64(p.taxCoin),
					}
					gwPlayerBet.PlayerBets = append(gwPlayerBet.PlayerBets, playerBet)
				}
			}
			if len(gwPlayerBet.PlayerBets) > 0 {
				proto.SetDefaults(gwPlayerBet)
				sceneEx.SendToWorld(int(server.SSPacketID_PACKET_GW_PLAYERBET), gwPlayerBet)
			}
		}

		if isSave {
			info, err := model.MarshalGameNoteByHUNDRED(&hundredType)
			if err == nil {
				logid, _ := model.AutoIncGameLogId()
				for _, oPlayer := range sceneEx.players {
					if oPlayer != nil {
						if !oPlayer.IsRob {
							if oPlayer.betTotal > 0 {
								totalIn, totalOut := int64(0), int64(0)
								winMoney := 0
								for _, v := range oPlayer.winCoin {
									winMoney += v
								}
								if winMoney > 0 {
									totalOut = int64(winMoney)
								} else {
									totalIn = int64(-winMoney)
								}

								sceneEx.SaveGamePlayerListLog(oPlayer.SnId,
									&base.SaveGamePlayerListLogParam{
										Platform:          oPlayer.Platform,
										Channel:           oPlayer.Channel,
										Promoter:          oPlayer.BeUnderAgentCode,
										PackageTag:        oPlayer.PackageID,
										InviterId:         oPlayer.InviterId,
										LogId:             logid,
										TotalIn:           totalIn,
										TotalOut:          totalOut,
										TaxCoin:           oPlayer.taxCoin,
										ClubPumpCoin:      0,
										BetAmount:         oPlayer.betTotal,
										WinAmountNoAnyTax: oPlayer.gainWinLost + oPlayer.betTotal,
										IsFirstGame:       sceneEx.IsPlayerFirst(oPlayer.Player),
									})
							}
						}
						oPlayer.SetCurrentCoin(oPlayer.Coin)
						oPlayer.Clean()
					}
				}
				var trends []string
				if len(sceneEx.trend20Lately) > 0 {
					for _, v := range sceneEx.trend20Lately {
						switch v {
						case 0:
							trends = append(trends, "和")
						case 1:
							trends = append(trends, "龙")
						case 2:
							trends = append(trends, "虎")
						}
					}
				}
				trend20Lately, _ := json.Marshal(trends)
				sceneEx.SaveGameDetailedLog(logid, info,
					&base.GameDetailedParam{
						Trend20Lately: string(trend20Lately),
					})
				hundredType = nil
			}
		}
		sceneEx.DealyTime = int64(common.RandFromRange(0, 2000)) * int64(time.Millisecond)
	}
}
func (this *SceneBilledStateDragonVsTiger) OnLeave(s *base.Scene) {
	if sceneEx, ok := s.ExtraData.(*DragonVsTigerSceneData); ok {
		for _, p := range sceneEx.players {
			leave := false
			var reason int
			if !p.IsOnLine() {
				leave = true
				reason = common.PlayerLeaveReason_DropLine
			} else if p.IsRob {
				if s.CoinOverMaxLimit(p.Coin, p.Player) {
					leave = true
					reason = common.PlayerLeaveReason_Normal
				} else if p.Trusteeship >= model.GameParamData.PlayerWatchNum {
					leave = true
					reason = common.PlayerLeaveReason_LongTimeNoOp
				}
			} else {
				if !s.CoinInLimit(p.Coin) {
					leave = true
					reason = common.PlayerLeaveReason_Bekickout
				} else if p.Trusteeship >= model.GameParamData.PlayerWatchNum {
					leave = true
					reason = common.PlayerLeaveReason_LongTimeNoOp
				}
			}
			todayGamefreeIDSceneData, _ := p.GetDaliyGameData(int(sceneEx.DbGameFree.GetId()))
			if !p.IsRob &&
				todayGamefreeIDSceneData != nil &&
				sceneEx.DbGameFree.GetPlayNumLimit() != 0 &&
				todayGamefreeIDSceneData.GameTimes >= int64(sceneEx.DbGameFree.GetPlayNumLimit()) {
				leave = true
				reason = common.PlayerLeaveReason_GameTimes
			}
			if leave {
				s.PlayerLeave(p.Player, reason, true)
			}
		}
		//先处理下庄家
		sceneEx.TryChangeBanker()

		seatKey := sceneEx.Resort()
		if seatKey != sceneEx.constSeatKey {
			dragonVsTigerSendSeatInfo(s, sceneEx)
			sceneEx.constSeatKey = seatKey
		}

		sceneEx.SendRobotUpBankerList()

		if s.CheckNeedDestroy() {
			sceneEx.SceneDestroy(true)
		}
	}
}
func (this *SceneBilledStateDragonVsTiger) OnTick(s *base.Scene) {
	this.SceneBaseStateDragonVsTiger.OnTick(s)
	if sceneEx, ok := s.ExtraData.(*DragonVsTigerSceneData); ok {
		if time.Now().Sub(sceneEx.StateStartTime) > DragonVsTigerBilledTimeout {
			s.ChangeSceneState(DragonVsTigerSceneStateStakeAnt)
		}
	}
}
func (this *ScenePolicyDragonVsTiger) RegisteSceneState(state base.SceneState) {
	if state == nil {
		return
	}
	stateid := state.GetState()
	if stateid < 0 || stateid >= DragonVsTigerSceneStateMax {
		return
	}
	this.states[stateid] = state
}
func (this *ScenePolicyDragonVsTiger) GetSceneState(s *base.Scene, stateid int) base.SceneState {
	if stateid >= 0 && stateid < DragonVsTigerSceneStateMax {
		return ScenePolicyDragonVsTigerSington.states[stateid]
	}
	return nil
}
func init() {
	ScenePolicyDragonVsTigerSington.RegisteSceneState(&SceneStakeAntStateDragonVsTiger{})
	ScenePolicyDragonVsTigerSington.RegisteSceneState(&SceneStakeStateDragonVsTiger{})
	ScenePolicyDragonVsTigerSington.RegisteSceneState(&SceneOpenCardAntStateDragonVsTiger{})
	ScenePolicyDragonVsTigerSington.RegisteSceneState(&SceneOpenCardStateDragonVsTiger{})
	ScenePolicyDragonVsTigerSington.RegisteSceneState(&SceneBilledStateDragonVsTiger{})
	core.RegisteHook(core.HOOK_BEFORE_START, func() error {
		base.RegisteScenePolicy(common.GameId_DragonVsTiger, 0, ScenePolicyDragonVsTigerSington)
		return nil
	})
}
