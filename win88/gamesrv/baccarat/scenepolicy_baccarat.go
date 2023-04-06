package baccarat

import (
	"encoding/json"
	"fmt"
	"games.yol.com/win88/common"
	rule "games.yol.com/win88/gamerule/baccarat"
	"games.yol.com/win88/gamesrv/base"
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	proto_baccarat "games.yol.com/win88/protocol/baccarat"
	proto_player "games.yol.com/win88/protocol/player"
	proto_server "games.yol.com/win88/protocol/server"
	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/timer"
	"math"
	"strconv"
	"time"
)

var ScenePolicyBaccaratSington = &ScenePolicyBaccarat{}

type ScenePolicyBaccarat struct {
	base.BaseScenePolicy
	states [rule.BaccaratSceneStateMax]base.SceneState
}

//创建场景扩展数据
func (this *ScenePolicyBaccarat) CreateSceneExData(s *base.Scene) interface{} {
	sceneEx := NewBaccaratSceneData(s)
	if sceneEx != nil {
		if sceneEx.init() {
			s.ExtraData = sceneEx
		}
	}
	return sceneEx
}

//创建玩家扩展数据
func (this *ScenePolicyBaccarat) CreatePlayerExData(s *base.Scene, p *base.Player) interface{} {
	playerEx := &BaccaratPlayerData{Player: p}
	if playerEx != nil {
		p.ExtraData = playerEx
	}
	return playerEx
}

func BaccaratBatchSendBet(sceneEx *BaccaratSceneData, force bool) {
	needSend := false
	pack := &proto_baccarat.SCBaccaratSendBet{}
	var olBetChips = make(map[int]int64)
	for _, playerEx := range sceneEx.seats {
		if playerEx.Pos == BACCARAT_OLPOS {
			for i := range BaccaratZoneMap {
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
		betinfo := &proto_baccarat.BaccaratBetInfo{}
		betinfo.SnId = proto.Int32(0)
		betinfo.TotalChips = append(betinfo.TotalChips, olBetChips[rule.BACCARAT_ZONE_TIE])
		betinfo.TotalChips = append(betinfo.TotalChips, olBetChips[rule.BACCARAT_ZONE_BANKER])
		betinfo.TotalChips = append(betinfo.TotalChips, olBetChips[rule.BACCARAT_ZONE_PLAYER])
		betinfo.TotalChips = append(betinfo.TotalChips, olBetChips[rule.BACCARAT_ZONE_BANKER_DOUBLE])
		betinfo.TotalChips = append(betinfo.TotalChips, olBetChips[rule.BACCARAT_ZONE_PLAYER_DOUBLE])
		pack.Data = append(pack.Data, betinfo)
		pack.TotalChips = append(pack.TotalChips, int64(sceneEx.betInfo[rule.BACCARAT_ZONE_TIE]))
		pack.TotalChips = append(pack.TotalChips, int64(sceneEx.betInfo[rule.BACCARAT_ZONE_BANKER]))
		pack.TotalChips = append(pack.TotalChips, int64(sceneEx.betInfo[rule.BACCARAT_ZONE_PLAYER]))
		pack.TotalChips = append(pack.TotalChips, int64(sceneEx.betInfo[rule.BACCARAT_ZONE_BANKER_DOUBLE]))
		pack.TotalChips = append(pack.TotalChips, int64(sceneEx.betInfo[rule.BACCARAT_ZONE_PLAYER_DOUBLE]))
		proto.SetDefaults(pack)
		sceneEx.Broadcast(int(proto_baccarat.BaccaratPacketID_PACKET_SC_BACCARAT_SENDBET), pack, 0)
	}
}

//场景开启事件
func (this *ScenePolicyBaccarat) OnStart(s *base.Scene) {
	logger.Logger.Trace("(this *ScenePolicyBaccarat) OnStart, SceneId=", s.SceneId)
	sceneEx := NewBaccaratSceneData(s)
	if sceneEx != nil {
		if sceneEx.init() {
			this.BaseScenePolicy.OnStart(s)
			if sceneEx.hBatchSend != timer.TimerHandle(0) {
				timer.StopTimer(sceneEx.hBatchSend)
				sceneEx.hBatchSend = timer.TimerHandle(0)
			}
			//批量广播发送筹码
			if hNext, ok := common.DelayInvake(func() {
				if sceneEx.SceneState.GetState() != rule.BaccaratSceneStateStake {
					return
				}
				BaccaratBatchSendBet(sceneEx, true)
			}, nil, rule.BaccaratBatchSendBetTimeout, -1); ok {
				sceneEx.hBatchSend = hNext
			}
			s.ExtraData = sceneEx
			s.ChangeSceneState(rule.BaccaratSceneStateStakeAnt)
		}
	}
}

//场景关闭事件
func (this *ScenePolicyBaccarat) OnStop(s *base.Scene) {
	logger.Logger.Trace("(this *ScenePolicyBaccaratSceneData) OnStop , SceneId=", s.SceneId)
	this.BaseScenePolicy.OnStop(s)
	if sceneEx, ok := s.ExtraData.(*BaccaratSceneData); ok {
		if sceneEx.hRunRecord != timer.TimerHandle(0) {
			timer.StopTimer(sceneEx.hRunRecord)
			sceneEx.hRunRecord = timer.TimerHandle(0)
		}
	}
}

//场景心跳事件
func (this *ScenePolicyBaccarat) OnTick(s *base.Scene) {
	this.BaseScenePolicy.OnTick(s)
	if s == nil {
		return
	}
	if s.SceneState != nil {
		s.SceneState.OnTick(s)
	}
}

func (this *ScenePolicyBaccarat) GetPlayerNum(s *base.Scene) int32 {
	if s == nil {
		return 0
	}
	if sceneEx, ok := s.ExtraData.(*BaccaratSceneData); ok {
		return int32(len(sceneEx.seats))
	}
	return 0
}

//玩家进入事件
func (this *ScenePolicyBaccarat) OnPlayerEnter(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyBaccarat) OnPlayerEnter, SceneId=", s.SceneId, " player=", p.Name)
	this.BaseScenePolicy.OnPlayerEnter(s, p)
	if sceneEx, ok := s.ExtraData.(*BaccaratSceneData); ok {
		playerEx := &BaccaratPlayerData{Player: p}
		if playerEx != nil {
			playerEx.init()
			playerEx.Clean()
			playerEx.Pos = BACCARAT_OLPOS
			sceneEx.seats = append(sceneEx.seats, playerEx)
			sceneEx.players[p.SnId] = playerEx
			p.ExtraData = playerEx
			//进房时金币低于下限,状态切换到观众
			if p.Coin < int64(sceneEx.DbGameFree.GetBetLimit()) || p.Coin < int64(sceneEx.DbGameFree.GetOtherIntParams()[0]) {
				p.MarkFlag(base.PlayerState_GameBreak)
			}
			//发送房间信息
			BaccaratSendRoomInfo(s, p, sceneEx, playerEx)
			s.FirePlayerEvent(p, base.PlayerEventEnter, nil)
		}
	}
}

//玩家离开事件
func (this *ScenePolicyBaccarat) OnPlayerLeave(s *base.Scene, p *base.Player, reason int) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyBaccarat) OnPlayerLeave, SceneId=", s.SceneId, " player=", p.Name)
	this.BaseScenePolicy.OnPlayerLeave(s, p, reason)
	if sceneEx, ok := s.ExtraData.(*BaccaratSceneData); ok {
		if this.CanChangeCoinScene(s, p) {
			sceneEx.OnPlayerLeave(p, reason)
			s.FirePlayerEvent(p, base.PlayerEventLeave, []int64{int64(reason)})
		}
	}
}

//玩家掉线
func (this *ScenePolicyBaccarat) OnPlayerDropLine(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyBaccarat) OnPlayerDropLine, SceneId=", s.SceneId, " player=", p.Name)
	this.BaseScenePolicy.OnPlayerDropLine(s, p)
	s.FirePlayerEvent(p, base.PlayerEventDropLine, nil)
	if sceneEx, ok := s.ExtraData.(*BaccaratSceneData); ok {
		if sceneEx.Gaming {
			return
		}
	}
}

//玩家重连
func (this *ScenePolicyBaccarat) OnPlayerRehold(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyBaccarat) OnPlayerRehold, SceneId=", s.SceneId, " player=", p.Name)
	this.BaseScenePolicy.OnPlayerRehold(s, p)
	if sceneEx, ok := s.ExtraData.(*BaccaratSceneData); ok {
		if playerEx, ok := p.ExtraData.(*BaccaratPlayerData); ok {
			BaccaratSendRoomInfo(s, p, sceneEx, playerEx)
			s.FirePlayerEvent(p, base.PlayerEventRehold, nil)
		}
	}
}

//玩家返回房间
func (this *ScenePolicyBaccarat) OnPlayerReturn(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyBaccarat) OnPlayerReturn, SceneId=", s.SceneId, " player=", p.Name)
	this.BaseScenePolicy.OnPlayerReturn(s, p)
	if sceneEx, ok := s.ExtraData.(*BaccaratSceneData); ok {
		if playerEx, ok := p.ExtraData.(*BaccaratPlayerData); ok {
			BaccaratSendRoomInfo(s, p, sceneEx, playerEx)
			s.FirePlayerEvent(p, base.PlayerEventReturn, nil)
		}
	}
}

//玩家操作，具体处理转到各个状态
func (this *ScenePolicyBaccarat) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	this.BaseScenePolicy.OnPlayerOp(s, p, opcode, params)
	if s == nil || p == nil {
		return false
	}
	if s.SceneState != nil {
		return s.SceneState.OnPlayerOp(s, p, opcode, params)
	}
	return true
}

//玩家事件
func (this *ScenePolicyBaccarat) OnPlayerEvent(s *base.Scene, p *base.Player, evtcode int, params []int64) {
	this.BaseScenePolicy.OnPlayerEvent(s, p, evtcode, params)
	if s == nil || p == nil {
		return
	}
	if s.SceneState != nil {
		s.SceneState.OnPlayerEvent(s, p, evtcode, params)
	}
}

//是否可以强制开始
func (this *ScenePolicyBaccarat) IsCanForceStart(s *base.Scene) bool {
	return true
}

////////////////////////////////////////////////////////////////////////////////
func (this *ScenePolicyBaccarat) RegisteSceneState(state base.SceneState) {
	if state == nil {
		return
	}
	stateid := state.GetState()
	if stateid < 0 || stateid >= rule.BaccaratSceneStateMax {
		return
	}
	this.states[stateid] = state
}

//
func (this *ScenePolicyBaccarat) GetSceneState(s *base.Scene, stateid int) base.SceneState {
	if stateid >= 0 && stateid < rule.BaccaratSceneStateMax {
		return ScenePolicyBaccaratSington.states[stateid]
	}
	return nil
}

//当前状态能否换桌
func (this *ScenePolicyBaccarat) CanChangeCoinScene(s *base.Scene, p *base.Player) bool {
	if s == nil || p == nil {
		return true
	}
	if s.SceneState != nil {
		return s.SceneState.CanChangeCoinScene(s, p)
	}
	return true
}
func (this *ScenePolicyBaccarat) NotifyGameState(s *base.Scene) {
	if s.GetState() > 2 {
		return
	}
	sec := int((rule.BaccaratStakeAntTimeout + rule.BaccaratStakeTimeout).Seconds())
	if sceneEx, ok := s.ExtraData.(*BaccaratSceneData); ok {
		s.SyncGameState(sec, len(sceneEx.bankerList))
	}
}

func (this *ScenePolicyBaccarat) PacketGameData(s *base.Scene) interface{} {
	if s == nil {
		return nil
	}
	if s.SceneState != nil {
		if sceneEx, ok := s.ExtraData.(*BaccaratSceneData); ok {
			switch s.SceneState.GetState() {
			case rule.BaccaratSceneStateStakeAnt, rule.BaccaratSceneStateStake:
				lt := int32((rule.BaccaratStakeTimeout - time.Now().Sub(sceneEx.StateStartTime)) / time.Second)
				pack := &proto_server.GWDTRoomInfo{
					DCoin:      proto.Int(sceneEx.betInfo[rule.BACCARAT_ZONE_PLAYER] - sceneEx.betInfoRob[rule.BACCARAT_ZONE_PLAYER]),
					TCoin:      proto.Int(sceneEx.betInfo[rule.BACCARAT_ZONE_BANKER] - sceneEx.betInfoRob[rule.BACCARAT_ZONE_BANKER]),
					NCoin:      proto.Int(sceneEx.betInfo[rule.BACCARAT_ZONE_TIE] - sceneEx.betInfoRob[rule.BACCARAT_ZONE_TIE]),
					DDCoin:     proto.Int(sceneEx.betInfo[rule.BACCARAT_ZONE_PLAYER_DOUBLE] - sceneEx.betInfoRob[rule.BACCARAT_ZONE_PLAYER_DOUBLE]),
					TDCoin:     proto.Int(sceneEx.betInfo[rule.BACCARAT_ZONE_BANKER_DOUBLE] - sceneEx.betInfoRob[rule.BACCARAT_ZONE_BANKER_DOUBLE]),
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
					win, lost := value.GetStaticsData(sceneEx.KeyGameId)
					pack.Players = append(pack.Players, &proto_server.PlayerDTCoin{
						NickName: proto.String(value.Name),
						Snid:     proto.Int32(value.SnId),
						DCoin:    proto.Int(value.betInfo[rule.BACCARAT_ZONE_PLAYER]),
						TCoin:    proto.Int(value.betInfo[rule.BACCARAT_ZONE_BANKER]),
						NCoin:    proto.Int(value.betInfo[rule.BACCARAT_ZONE_TIE]),
						DDCoin:   proto.Int(value.betInfo[rule.BACCARAT_ZONE_PLAYER_DOUBLE]),
						TDCoin:   proto.Int(value.betInfo[rule.BACCARAT_ZONE_BANKER_DOUBLE]),
						Totle:    proto.Int64(win - lost),
					})
				}
				return pack
			default:
				return &proto_server.GWDTRoomInfo{}
			}
		}
	}
	return nil
}

func (this *ScenePolicyBaccarat) InterventionGame(s *base.Scene, data interface{}) interface{} {
	if s == nil {
		return nil
	}
	if s.SceneState != nil {
		if sceneEx, ok := s.ExtraData.(*BaccaratSceneData); ok {
			if v, ok := data.(base.InterventionResults); ok {
				return sceneEx.ParserResults1(v.Results, v.Webuser)
			}
			switch s.SceneState.GetState() {
			case rule.BaccaratSceneStateStakeAnt, rule.BaccaratSceneStateStake:
				if d, ok := data.(base.InterventionData); ok {
					if sceneEx.NumOfGames == int(d.NumOfGames) {
						if sceneEx.TryResult(int(d.Flag)) {
							sceneEx.bIntervention = true
							sceneEx.webUser = d.Webuser
						}
					}
				}
			default:
			}
		}
	}
	return nil
}

//座位数据
func BaccaratCreateSeats(sceneEx *BaccaratSceneData) []*proto_baccarat.BaccaratPlayerData {
	var datas []*proto_baccarat.BaccaratPlayerData
	cnt := 0
	const N = BACCARAT_RICHTOP5 + 2
	var seats [N]*BaccaratPlayerData
	if sceneEx.winTop1 != nil { //神算子
		seats[cnt] = sceneEx.winTop1
		cnt++
	}
	for i := 0; i < BACCARAT_RICHTOP5; i++ {
		if sceneEx.betTop5[i] != nil && sceneEx.betTop5[i] != sceneEx.winTop1 {
			seats[cnt] = sceneEx.betTop5[i]
			cnt++
		}
	}
	for i := 0; i < N-1; i++ {
		if seats[i] != nil {
			pd := &proto_baccarat.BaccaratPlayerData{
				SnId: proto.Int32(seats[i].SnId),
				Name: proto.String(seats[i].Name),
				Head: proto.Int32(seats[i].Head),
				Sex:  proto.Int32(seats[i].Sex),
				Coin: proto.Int64(seats[i].Coin),
				Pos:  proto.Int(seats[i].Pos),
				Flag: proto.Int(seats[i].GetFlag()),
				City: proto.String(seats[i].GetCity()),
				//				Params:      seats[i].Params,
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

			pd := &proto_baccarat.BaccaratPlayerData{
				SnId: proto.Int32(banker.SnId),
				Name: proto.String(banker.Name),
				Head: proto.Int32(banker.Head),
				Sex:  proto.Int32(banker.Sex),
				Coin: proto.Int64(banker.Coin),
				Pos:  proto.Int(BACCARAT_BANKERPOS),
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

func BaccaratCreateRoomInfoPacket(s *base.Scene, sceneEx *BaccaratSceneData, playerEx *BaccaratPlayerData) proto.Message {
	pack := &proto_baccarat.SCBaccaratRoomInfo{
		RoomId:   proto.Int(s.SceneId),
		Creator:  proto.Int32(s.Creator),
		GameId:   proto.Int(s.GameId),
		RoomMode: proto.Int(s.SceneMode),
		AgentId:  proto.Int32(s.GetAgentor()),
		//SceneType: proto.Int(s.sceneType),
		//Cards:         common.CopySliceInt32(sceneEx.cards[:]),
		Params: []int32{sceneEx.DbGameFree.GetLimitCoin(), sceneEx.DbGameFree.GetMaxCoinLimit(),
			sceneEx.DbGameFree.GetServiceFee(), sceneEx.DbGameFree.GetLowerThanKick(), sceneEx.DbGameFree.GetBaseScore(),
			0, sceneEx.DbGameFree.GetBetLimit(), sceneEx.DbGameFree.GetBanker()},
		NumOfGames:    proto.Int(sceneEx.NumOfGames),
		State:         proto.Int(s.SceneState.GetState()),
		TimeOut:       proto.Int(s.SceneState.GetTimeout(s)),
		BankerId:      proto.Int32(sceneEx.bankerSnId),
		Players:       BaccaratCreateSeats(sceneEx),
		Trend100Cur:   sceneEx.trend100Cur,
		Trend20Lately: sceneEx.trend20Lately,
		TotalChips:    make([]int64, 5),
		OLNum:         proto.Int(len(sceneEx.seats)),
		DisbandGen:    proto.Int(sceneEx.GetDisbandGen()),
		ParamsEx:      s.GetParamsEx(),
		LastCards:     proto.Int(sceneEx.poker.Count()),
		LoopNum:       proto.Int(sceneEx.LoopNum),
	}
	for _, value := range sceneEx.DbGameFree.GetMaxBetCoin() {
		pack.Params = append(pack.Params, value)
	}
	if s.SceneState.GetState() == rule.BaccaratSceneStateOpenCard ||
		s.SceneState.GetState() == rule.BaccaratSceneStateBilled {
		pack.Cards = sceneEx.cardsAndPoint
	}

	if playerEx != nil { //自己的数据
		pd := &proto_baccarat.BaccaratPlayerData{
			SnId: proto.Int32(playerEx.SnId),
			Name: proto.String(playerEx.Name),
			Head: proto.Int32(playerEx.Head),
			Sex:  proto.Int32(playerEx.Sex),
			Coin: proto.Int64(playerEx.Coin),
			Pos:  proto.Int(BACCARAT_SELFPOS),
			Flag: proto.Int(playerEx.GetFlag()),
			City: proto.String(playerEx.GetCity()),
			//Params:      proto.String(playerEx.Params),
			Lately20Win: proto.Int64(playerEx.lately20Win),
			Lately20Bet: proto.Int64(playerEx.lately20Bet),
			HeadOutLine: proto.Int32(playerEx.HeadOutLine),
			VIP:         proto.Int32(playerEx.VIP),
			NiceId:      proto.Int32(playerEx.NiceId),
		}
		pack.Players = append(pack.Players, pd)
	}
	pack.TotalChips[0] = int64(sceneEx.betInfo[rule.BACCARAT_ZONE_TIE])
	pack.TotalChips[1] = int64(sceneEx.betInfo[rule.BACCARAT_ZONE_BANKER])
	pack.TotalChips[2] = int64(sceneEx.betInfo[rule.BACCARAT_ZONE_PLAYER])
	pack.TotalChips[3] = int64(sceneEx.betInfo[rule.BACCARAT_ZONE_BANKER_DOUBLE])
	pack.TotalChips[4] = int64(sceneEx.betInfo[rule.BACCARAT_ZONE_PLAYER_DOUBLE])
	if playerEx != nil {
		for i := range BaccaratZoneMap {
			if len(playerEx.betDetailInfo[i]) != 0 {
				chip := &proto_baccarat.BaccaratZoneChips{
					Zone: proto.Int(i),
				}
				for k, v := range playerEx.betDetailInfo[i] {
					chip.Detail = append(chip.Detail, &proto_baccarat.BaccaratChips{
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

func BaccaratSendRoomInfo(s *base.Scene, p *base.Player, sceneEx *BaccaratSceneData, playerEx *BaccaratPlayerData) {
	pack := BaccaratCreateRoomInfoPacket(s, sceneEx, playerEx)
	//logger.Logger.Trace("BaccaratSendRoomInfo pack:", pack)
	p.SendToClient(int(proto_baccarat.BaccaratPacketID_PACKET_SC_BACCARAT_ROOMINFO), pack)
}

func BaccaratSendSeatInfo(s *base.Scene, sceneEx *BaccaratSceneData) {
	pack := &proto_baccarat.SCBaccaratSeats{
		PlayerNum: proto.Int(len(sceneEx.seats)),
		Data:      BaccaratCreateSeats(sceneEx),
		BankerId:  proto.Int32(sceneEx.bankerSnId),
	}
	proto.SetDefaults(pack)
	s.Broadcast(int(proto_baccarat.BaccaratPacketID_PACKET_SC_BACCARAT_SEATS), pack, 0)
}

//////////////////////////////////////////////////////////////
//状态基类
//////////////////////////////////////////////////////////////
type SceneBaseStateBaccarat struct {
}

func (this *SceneBaseStateBaccarat) GetTimeout(s *base.Scene) int {
	if sceneEx, ok := s.ExtraData.(*BaccaratSceneData); ok {
		return int(time.Now().Sub(sceneEx.StateStartTime) / time.Second)
	}
	return 0
}
func (this *SceneBaseStateBaccarat) CanChangeTo(s base.SceneState) bool {
	return true
}

//当前状态能否换桌
func (this *SceneBaseStateBaccarat) CanChangeCoinScene(s *base.Scene, p *base.Player) bool {
	if playerEx, ok := p.ExtraData.(*BaccaratPlayerData); ok {
		if sceneEx, ok := s.ExtraData.(*BaccaratSceneData); ok {
			if playerEx.betTotal != 0 {
				playerEx.OpCode = proto_player.OpResultCode_OPRC_Hundred_YouHadBetCannotLeave
				return false
			}
			if playerEx.SnId == sceneEx.bankerSnId {
				p.OpCode = proto_player.OpResultCode_OPRC_Hundred_YouHadBankerCannotLeave
				return false
			}
		}
	}
	return true
}
func (this *SceneBaseStateBaccarat) OnEnter(s *base.Scene) {
	if sceneEx, ok := s.ExtraData.(*BaccaratSceneData); ok {
		sceneEx.StateStartTime = time.Now()
	}
}
func (this *SceneBaseStateBaccarat) OnLeave(s *base.Scene) {
}
func (this *SceneBaseStateBaccarat) OnTick(s *base.Scene) {
}

//发送玩家操作情况
func (this *SceneBaseStateBaccarat) SendSCPlayerOp(s *base.Scene, p *base.Player, pos int, opcode int, opRetCode proto_baccarat.OpResultCode, params []int64, broadcastall bool) {
	pack := &proto_baccarat.SCBaccaratOp{
		OpCode:    proto.Int(opcode),
		SnId:      proto.Int32(p.SnId),
		OpRetCode: opRetCode,
		Params:    params,
	}
	proto.SetDefaults(pack)
	if broadcastall {
		s.Broadcast(int(proto_baccarat.BaccaratPacketID_PACKET_SC_BACCARAT_PLAYEROP), pack, 0)
	} else {
		p.SendToClient(int(proto_baccarat.BaccaratPacketID_PACKET_SC_BACCARAT_PLAYEROP), pack)
	}
}
func (this *SceneBaseStateBaccarat) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	playerEx, ok := p.ExtraData.(*BaccaratPlayerData)
	if !ok {
		return false
	}
	if sceneEx, ok := s.ExtraData.(*BaccaratSceneData); ok {
		switch opcode {
		case rule.BaccaratPlayerOpGetOLList: //在线玩家列表
			seats := make([]*BaccaratPlayerData, 0, BACCARAT_OLTOP20+1)
			if sceneEx.winTop1 != nil { //神算子
				seats = append(seats, sceneEx.winTop1)
			}
			count := len(sceneEx.seats)
			topCnt := 0
			for i := 0; i < count && topCnt < BACCARAT_OLTOP20; i++ { //top20
				if sceneEx.seats[i] != sceneEx.winTop1 {
					seats = append(seats, sceneEx.seats[i])
					topCnt++
				}
			}
			pack := &proto_baccarat.SCBaccaratPlayerList{}
			for i := 0; i < len(seats); i++ {
				pack.Data = append(pack.Data, &proto_baccarat.BaccaratPlayerData{
					SnId: proto.Int32(seats[i].SnId),
					Name: proto.String(seats[i].Name),
					Head: proto.Int32(seats[i].Head),
					Sex:  proto.Int32(seats[i].Sex),
					Coin: proto.Int64(seats[i].Coin),
					Pos:  proto.Int(seats[i].Pos),
					Flag: proto.Int(seats[i].GetFlag()),
					City: proto.String(seats[i].GetCity()),
					//					Params:      seats[i].Params,
					Lately20Win: proto.Int64(seats[i].lately20Win),
					Lately20Bet: proto.Int64(seats[i].lately20Bet),
					HeadOutLine: proto.Int32(seats[i].HeadOutLine),
					VIP:         proto.Int32(seats[i].VIP),
					NiceId:      proto.Int32(seats[i].NiceId),
				})
			}
			//truePlayerCount := int32(len(s.Players))
			////获取fake用户数量
			//correctNum := sceneEx.DbGameFree.GetCorrectNum()
			//correctRate := sceneEx.DbGameFree.GetCorrectRate()
			//fakePlayerCount := correctNum + truePlayerCount*correctRate/100 + sceneEx.DbGameFree.GetDeviation()
			//pack.OLNum = proto.Int32(truePlayerCount + fakePlayerCount)
			proto.SetDefaults(pack)
			p.SendToClient(int(proto_baccarat.BaccaratPacketID_PACKET_SC_BACCARAT_PLAYERLIST), pack)
			return true
		case rule.BaccaratPlayerOpUpBanker: //上庄
			if sceneEx.bankerSnId == playerEx.SnId {
				return true
			}
			if sceneEx.DbGameFree.GetBanker() == 0 {
				this.SendSCPlayerOp(s, p, playerEx.Pos, opcode, proto_baccarat.OpResultCode_OPRC_Baccarat_BankerWaiting, params, false)
				return true
			}

			if (p.Coin) < int64(sceneEx.DbGameFree.GetBanker()) {
				this.SendSCPlayerOp(s, p, playerEx.Pos, opcode, proto_baccarat.OpResultCode_OPRC_Baccarat_BankerLimit, params, false)
				return true
			}
			if len(sceneEx.bankerList) > 0 {
				for _, v := range sceneEx.bankerList {
					if v.SnId == playerEx.SnId {
						this.SendSCPlayerOp(s, p, playerEx.Pos, opcode, proto_baccarat.OpResultCode_OPRC_Baccarat_BankerWaiting, params, false)
						return true
					}
				}
			}
			sceneEx.bankerList = append(sceneEx.bankerList, playerEx)
			this.SendSCPlayerOp(s, p, playerEx.Pos, opcode, proto_baccarat.OpResultCode_OPRC_Sucess, params, false)
			ret := sceneEx.BankerList()
			ret.IsExist = proto.Int(-1)
			sceneEx.Broadcast(int(proto_baccarat.BaccaratPacketID_PACKET_SC_BACCARAT_BANKERLIST), ret, 0)
			return true
		case rule.BaccaratPlayerOpNowDwonBanker: //在庄的下庄
			logger.Logger.Tracef("玩家是庄，现在申请下庄 sceneEx.bankerTimes = %v", sceneEx.bankerList)
			opRetCode := proto_baccarat.OpResultCode_OPRC_Sucess
			if sceneEx.bankerTimes == BACCARAT_BANKERNUMBERS {
				//玩家已经下庄
				params = append(params, 2)
				opRetCode = proto_baccarat.OpResultCode_OPRC_Error
			} else {
				//下庄成功
				params = append(params, 1)
			}
			sceneEx.bankerTimes = BACCARAT_BANKERNUMBERS
			this.SendSCPlayerOp(s, p, playerEx.Pos, opcode, opRetCode, params, false)
			return true
		case rule.BaccaratPlayerOpUpList: //上庄列表
			down := 0
			if len(sceneEx.bankerList) > 0 {
				for _, v := range sceneEx.bankerList {
					if v.SnId == playerEx.SnId {
						down = 1
						break
					}
				}
			}

			ret := sceneEx.BankerList()
			ret.IsExist = proto.Int(down)
			p.SendToClient(int(proto_baccarat.BaccaratPacketID_PACKET_SC_BACCARAT_BANKERLIST), ret)
			return true
		case rule.BaccaratPlayerOpDwonBanker: //下庄
			down := false
			if len(sceneEx.bankerList) > 0 {
				for index, v := range sceneEx.bankerList {
					if v.SnId == playerEx.SnId {
						sceneEx.bankerList = append(sceneEx.bankerList[:index], sceneEx.bankerList[index+1:]...)
						down = true
					}
				}
			}
			if !down {
				this.SendSCPlayerOp(s, p, playerEx.Pos, opcode, proto_baccarat.OpResultCode_OPRC_Baccarat_NotBankerWaiting, params, false)
				return true
			}
			this.SendSCPlayerOp(s, p, playerEx.Pos, opcode, proto_baccarat.OpResultCode_OPRC_Sucess, params, false)
			ret := sceneEx.BankerList()
			ret.IsExist = proto.Int(-1)
			sceneEx.Broadcast(int(proto_baccarat.BaccaratPacketID_PACKET_SC_BACCARAT_BANKERLIST), ret, 0)
			return true
		}
	}
	return false
}
func (this *SceneBaseStateBaccarat) OnPlayerEvent(s *base.Scene, p *base.Player, evtcode int, params []int64) {
	if sceneEx, ok := s.ExtraData.(*BaccaratSceneData); ok {
		if playerEx, ok := p.ExtraData.(*BaccaratPlayerData); ok {
			needResort := false
			switch evtcode {
			case base.PlayerEventEnter:
				if sceneEx.winTop1 == nil {
					needResort = true
				}
				if !needResort {
					for i := 0; i < BACCARAT_RICHTOP5; i++ {
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
					for i := 0; i < BACCARAT_RICHTOP5; i++ {
						if sceneEx.betTop5[i] == playerEx {
							needResort = true
							break
						}
					}
				}
				//正在做庄，将坐庄计数至为最大
				if sceneEx.bankerSnId != -1 && playerEx.SnId == sceneEx.bankerSnId {
					sceneEx.bankerTimes = BACCARAT_BANKERNUMBERS
				}
				//在上庄列表中，从列表中剔除
				if len(sceneEx.bankerList) > 0 {
					for index, v := range sceneEx.bankerList {
						if v.SnId == playerEx.SnId {
							sceneEx.bankerList = append(sceneEx.bankerList[:index], sceneEx.bankerList[index+1:]...)
							sceneEx.Broadcast(int(proto_baccarat.BaccaratPacketID_PACKET_SC_BACCARAT_BANKERLIST),
								&proto_baccarat.SCBaccaratBankerList{
									Count:   proto.Int(len(sceneEx.bankerList)),
									IsExist: proto.Int32(-1),
								}, 0)
							break
						}
					}
				}
			case base.PlayerEventRecharge:
				if p.Pos < BACCARAT_SELFPOS {
					p.AddCoin(params[0], common.GainWay_Pay, base.SyncFlag_Broadcast, "system", p.GetScene().GetSceneName())
				} else {
					p.AddCoin(params[0], common.GainWay_Pay, base.SyncFlag_ToClient, "system", p.GetScene().GetSceneName())
				}

				if p.Coin >= int64(sceneEx.DbGameFree.GetBetLimit()) && p.Coin >= int64(sceneEx.DbGameFree.GetOtherIntParams()[0]) {
					p.UnmarkFlag(base.PlayerState_GameBreak)
					p.SyncFlag(true)
				}
			}
			if needResort {
				seatKey := sceneEx.Resort()
				if seatKey != sceneEx.constSeatKey {
					BaccaratSendSeatInfo(s, sceneEx)
					sceneEx.constSeatKey = seatKey
				}
			}
		}
	}
}

//////////////////////////////////////////////////////////////
//准备押注动画状态
//////////////////////////////////////////////////////////////
type SceneStakeAntStateBaccarat struct {
	SceneBaseStateBaccarat
}

//获取当前场景状态
func (this *SceneStakeAntStateBaccarat) GetState() int {
	return rule.BaccaratSceneStateStakeAnt
}

//是否可以改变到其它场景状态
func (this *SceneStakeAntStateBaccarat) CanChangeTo(s base.SceneState) bool {
	if s.GetState() == rule.BaccaratSceneStateStake {
		return true
	}
	return false
}

//玩家操作
func (this *SceneStakeAntStateBaccarat) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if this.SceneBaseStateBaccarat.OnPlayerOp(s, p, opcode, params) {
		return true
	}
	return true
}

//场景状态进入广播所有客户端状态
func (this *SceneStakeAntStateBaccarat) OnEnter(s *base.Scene) {
	this.SceneBaseStateBaccarat.OnEnter(s)
	if sceneEx, ok := s.ExtraData.(*BaccaratSceneData); ok {
		sceneEx.CleanAll()
		sceneEx.NumOfGames++
		sceneEx.GameNowTime = time.Now()
		sceneEx.CheckResults()
		logger.Logger.Tracef("(this *base.Scene) [%v] 场景状态进入 %v", s.SceneId, len(sceneEx.seats))
		pack := &proto_baccarat.SCBaccaratRoomState{
			State:  proto.Int(this.GetState()),
			Params: []int32{int32(sceneEx.poker.Count())},
		}
		proto.SetDefaults(pack)
		s.Broadcast(int(proto_baccarat.BaccaratPacketID_PACKET_SC_BACCARAT_ROOMSTATE), pack, 0)
	}
}

//场景状态离开
func (this *SceneStakeAntStateBaccarat) OnLeave(s *base.Scene) {
}

//如果状态时间到则切换到押注状态
func (this *SceneStakeAntStateBaccarat) OnTick(s *base.Scene) {
	this.SceneBaseStateBaccarat.OnTick(s)
	if sceneEx, ok := s.ExtraData.(*BaccaratSceneData); ok {
		if time.Now().Sub(sceneEx.StateStartTime) > rule.BaccaratStakeAntTimeout {
			s.ChangeSceneState(rule.BaccaratSceneStateStake)
		}
	}
}

//////////////////////////////////////////////////////////////
//押注状态
//////////////////////////////////////////////////////////////
type SceneStakeStateBaccarat struct {
	SceneBaseStateBaccarat
}

func (this *SceneStakeStateBaccarat) GetState() int {
	return rule.BaccaratSceneStateStake
}
func (this *SceneStakeStateBaccarat) CanChangeTo(s base.SceneState) bool {
	switch s.GetState() {
	case rule.BaccaratSceneStateOpenCardAnt:
		return true
	}
	return false
}
func (this *SceneStakeStateBaccarat) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if this.SceneBaseStateBaccarat.OnPlayerOp(s, p, opcode, params) {
		return true
	}
	if playerEx, ok := p.ExtraData.(*BaccaratPlayerData); ok {
		if sceneEx, ok := s.ExtraData.(*BaccaratSceneData); ok {
			//玩家下注
			switch opcode {
			case rule.BaccaratPlayerOpBet:
				//logger.Trace(params)
				if len(params) >= 2 {
					if playerEx.SnId == sceneEx.bankerSnId { //庄家不能下注
						this.SendSCPlayerOp(s, p, playerEx.Pos, opcode, proto_baccarat.OpResultCode_OPRC_Baccarat_BankerCannotBet, params, false)
						return false
					}
					betPos := int(params[0]) //下注位置

					if sceneEx.winTop1 != nil && playerEx.IsRob {
						if sceneEx.winTop1.lastBetPos == -1 {
							sceneEx.winTop1.lastBetPos = betPos
						}
						if sceneEx.winTop1.IsRob && sceneEx.winTop1.lastBetPos != betPos && playerEx.lastBetPos != -1 {
							return false
						}
					}

					if _, exist := BaccaratZoneMap[betPos]; !exist {
						logger.Error("下注位置错误：", betPos)
						return false
					}
					betCoin := int(params[1]) //下注金额
					chips := sceneEx.DbGameFree.GetOtherIntParams()
					if !common.InSliceInt32(chips, int32(params[1])) {
						return false
					}
					if p.Coin < int64(sceneEx.DbGameFree.GetBetLimit()) && playerEx.betTotal == 0 {
						this.SendSCPlayerOp(s, p, playerEx.Pos, opcode, proto_baccarat.OpResultCode_OPRC_Baccarat_CoinMustReachTheValue, params, false)
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
								this.SendSCPlayerOp(s, p, playerEx.Pos, opcode, proto_baccarat.OpResultCode_OPRC_Baccarat_EachBetsLimit, params, false)

								//切换到开牌阶段
								if sceneEx.BetStop() {
									msgPack := common.CreateSrvMsg(common.SRVMSG_CODE_DEFAULT, "庄家可押注额度满，直接开牌")
									s.Broadcast(int(proto_player.PlayerPacketID_PACKET_SC_SRVMSG), msgPack, -1)
									s.ChangeSceneState(rule.BaccaratSceneStateOpenCardAnt)
								} else {
									msgPack := common.CreateSrvMsg(common.SRVMSG_CODE_DEFAULT, "该区域押注额度满")
									p.SendToClient(int(proto_player.PlayerPacketID_PACKET_SC_SRVMSG), msgPack)
								}
								return false
							}
						}
					}
					//期望下注金额
					expectBetCoin := betCoin
					total := int64(betCoin) + playerEx.betTotal
					if total <= playerEx.Coin {
						//闲家单门下注总额是否到达上限
						maxBetCoin := sceneEx.DbGameFree.GetMaxBetCoin()
						if ok, coinLimit := playerEx.MaxChipCheck(betPos, betCoin, maxBetCoin); !ok {
							betCoin = int(coinLimit) - playerEx.betInfo[betPos]
							//对齐到最小面额的筹码
							betCoin /= minChip
							betCoin *= minChip
							if betCoin <= 0 {
								//this.SendSCPlayerOp(s, p, playerEx.Pos, opcode, proto_baccarat.OpResultCode_OPRC_Hundred_EachBetsLimit, params, false)
								msgPack := common.CreateSrvMsg(common.SRVMSG_CODE_DEFAULT, fmt.Sprintf("该门押注金额上限%.2f", float64(coinLimit)/100))
								p.SendToClient(int(proto_player.PlayerPacketID_PACKET_SC_SRVMSG), msgPack)
								return false
							}
						}
					} else {
						this.SendSCPlayerOp(s, p, playerEx.Pos, opcode, proto_baccarat.OpResultCode_OPRC_Baccarat_CoinIsNotEnough, params, false)
						return false
					}
					playerEx.Trusteeship = 0
					//累积总投注额
					playerEx.TotalBet += int64(betCoin)
					//那一个闲家的下注总额
					playerEx.betTotal += int64(betCoin)
					playerEx.betInfo[betPos] += betCoin
					sceneEx.betInfo[betPos] += betCoin
					if playerEx.IsRob { //机器人下注
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
					params[1] = int64(betCoin) //修正下当前的实际下注额
					restCoin := playerEx.Coin - playerEx.betTotal
					params = append(params, restCoin)
					//params = append(params, int32(sceneEx.betInfo[betPos]))
					this.SendSCPlayerOp(s, p, playerEx.Pos, opcode, proto_baccarat.OpResultCode_OPRC_Sucess, params, playerEx.Pos < BACCARAT_SELFPOS)

					playerEx.betDetailOrderInfo[betPos] = append(playerEx.betDetailOrderInfo[betPos], int64(betCoin))

					//没钱了，要转到观战模式
					if restCoin < int64(chips[0]) || restCoin < int64(sceneEx.DbGameFree.GetBetLimit()) {
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
func (this *SceneStakeStateBaccarat) OnEnter(s *base.Scene) {
	this.SceneBaseStateBaccarat.OnEnter(s)
	if sceneEx, ok := s.ExtraData.(*BaccaratSceneData); ok {
		sceneEx.CheckResults()
		pack := &proto_baccarat.SCBaccaratRoomState{
			State: proto.Int(this.GetState()),
		}
		proto.SetDefaults(pack)
		s.Broadcast(int(proto_baccarat.BaccaratPacketID_PACKET_SC_BACCARAT_ROOMSTATE), pack, 0)
	}
}
func (this *SceneStakeStateBaccarat) OnLeave(s *base.Scene) {
	this.SceneBaseStateBaccarat.OnLeave(s)
	if sceneEx, ok := s.ExtraData.(*BaccaratSceneData); ok {
		BaccaratBatchSendBet(sceneEx, true)
		/*betCoin := 0
		for i := 0; i < DVST_ZONE_MAX; i++ {
			betCoin += sceneEx.betInfo[i] - sceneEx.betInfoRob[i]
		}
		base.CoinPoolMgr.PushCoin(sceneEx.GetGameFreeId(), sceneEx.Platform, int64(betCoin))*/
		for _, v := range sceneEx.players {
			if v != nil && !v.IsRob {
				if v.SnId == sceneEx.bankerSnId {
					v.Trusteeship = 0
				} else if v.betTotal <= 0 {
					v.Trusteeship++
					if v.Trusteeship >= model.GameParamData.NotifyPlayerWatchNum {
						v.SendTrusteeshipTips()
					}
				}
			}
		}
	}
}
func (this *SceneStakeStateBaccarat) OnTick(s *base.Scene) {
	this.SceneBaseStateBaccarat.OnTick(s)
	if sceneEx, ok := s.ExtraData.(*BaccaratSceneData); ok {
		if time.Now().Sub(sceneEx.StateStartTime) > rule.BaccaratStakeTimeout {
			s.ChangeSceneState(rule.BaccaratSceneStateOpenCardAnt)
		}
	}
}

//////////////////////////////////////////////////////////////
//准备开牌动画状态
//////////////////////////////////////////////////////////////
type SceneOpenCardAntStateBaccarat struct {
	SceneBaseStateBaccarat
}

func (this *SceneOpenCardAntStateBaccarat) GetState() int {
	return rule.BaccaratSceneStateOpenCardAnt
}
func (this *SceneOpenCardAntStateBaccarat) CanChangeTo(s base.SceneState) bool {
	switch s.GetState() {
	case rule.BaccaratSceneStateOpenCard:
		return true
	}
	return false
}
func (this *SceneOpenCardAntStateBaccarat) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if this.SceneBaseStateBaccarat.OnPlayerOp(s, p, opcode, params) {
		return true
	}
	return true
}
func (this *SceneOpenCardAntStateBaccarat) OnEnter(s *base.Scene) {
	this.SceneBaseStateBaccarat.OnEnter(s)
	if sceneEx, ok := s.ExtraData.(*BaccaratSceneData); ok {
		if !sceneEx.bIntervention {
			//发牌
			sceneEx.preAnalysis()
			//测试的代码
			//sceneEx.cards = [6]int32{3 + 13, 0, -1, 4, 0 + 13, -1}
			//sceneEx.calculationWinZone() //计算输赢区域
			openSingle, singlePlayer := sceneEx.IsSingleRegulatePlayer()
			if openSingle {
				if sceneEx.IsNeedSingleRegulate(singlePlayer) {
					if sceneEx.RegulationCard(singlePlayer) {
						//singlePlayer.单控计数++
						singlePlayer.MarkFlag(base.PlayerState_SAdjust)
						singlePlayer.AddAdjustCount(sceneEx.GetGameFreeId())
					} else {
						singlePlayer.result = 0
					}
				} else {
					if sceneEx.RegulationCard(singlePlayer) {
						//singlePlayer.单控计数++
						singlePlayer.MarkFlag(base.PlayerState_SAdjust)
						singlePlayer.AddAdjustCount(sceneEx.GetGameFreeId())
					} else {
						singlePlayer.result = 0
					}
				}
			}

			//if (!openSingle || (openSingle && singlePlayer.result == 0)) &&
			//	sceneEx.bankerSnId == -1 && sceneEx.IsSmartOperation() {
			//	sceneEx.stop = true // 等待接口返回后发牌，如果等待时间过长，修改接口超时时长
			//	sceneEx.SmartDeal()
			//}
		}

		//牌已经准备好，通知客户端状态
		pack := &proto_baccarat.SCBaccaratRoomState{
			State: proto.Int(this.GetState()),
		}
		proto.SetDefaults(pack)
		s.Broadcast(int(proto_baccarat.BaccaratPacketID_PACKET_SC_BACCARAT_ROOMSTATE), pack, 0)
	}
}
func (this *SceneOpenCardAntStateBaccarat) OnLeave(s *base.Scene) {}
func (this *SceneOpenCardAntStateBaccarat) OnTick(s *base.Scene) {
	this.SceneBaseStateBaccarat.OnTick(s)
	if sceneEx, ok := s.ExtraData.(*BaccaratSceneData); ok {
		if !sceneEx.stop && time.Now().Sub(sceneEx.StateStartTime) > rule.BaccaratOpenCardAntTimeout {
			s.ChangeSceneState(rule.BaccaratSceneStateOpenCard)
		}
	}
}

//////////////////////////////////////////////////////////////
//开牌状态
//////////////////////////////////////////////////////////////
type SceneOpenCardStateBaccarat struct {
	SceneBaseStateBaccarat
}

func (this *SceneOpenCardStateBaccarat) GetState() int {
	return rule.BaccaratSceneStateOpenCard
}
func (this *SceneOpenCardStateBaccarat) CanChangeTo(s base.SceneState) bool {
	switch s.GetState() {
	case rule.BaccaratSceneStateBilled:
		return true
	}
	return false
}
func (this *SceneOpenCardStateBaccarat) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if this.SceneBaseStateBaccarat.OnPlayerOp(s, p, opcode, params) {
		return true
	}
	return true
}
func (this *SceneOpenCardStateBaccarat) OnEnter(s *base.Scene) {
	this.SceneBaseStateBaccarat.OnEnter(s)
	if sceneEx, ok := s.ExtraData.(*BaccaratSceneData); ok {
		//将牌发给客户端
		sceneEx.cardsAndPoint = make([]int32, 3)
		copy(sceneEx.cardsAndPoint, sceneEx.cards[:3])
		sceneEx.cardsAndPoint = append(sceneEx.cardsAndPoint, int32(rule.GetPointNum(sceneEx.cards[:], 0, 1, 2)))
		sceneEx.cardsAndPoint = append(sceneEx.cardsAndPoint, sceneEx.cards[3:]...)
		sceneEx.cardsAndPoint = append(sceneEx.cardsAndPoint, int32(rule.GetPointNum(sceneEx.cards[:], 3, 4, 5)))
		sceneEx.cardsAndPoint = append(sceneEx.cardsAndPoint, int32(sceneEx.winZone))
		sceneEx.cardsAndPoint = append(sceneEx.cardsAndPoint, int32(sceneEx.poker.Count()))
		//logger.Trace("Baccarat params:", sceneEx.cardsAndPoint)
		pack := &proto_baccarat.SCBaccaratRoomState{
			State:  proto.Int(this.GetState()),
			Params: common.CopySliceInt32(sceneEx.cardsAndPoint[:]),
		}
		proto.SetDefaults(pack)
		s.Broadcast(int(proto_baccarat.BaccaratPacketID_PACKET_SC_BACCARAT_ROOMSTATE), pack, 0)
		sceneEx.PushTrend(int32(sceneEx.winZone))
	}
}
func (this *SceneOpenCardStateBaccarat) OnLeave(s *base.Scene) {
	this.SceneBaseStateBaccarat.OnLeave(s)
	if sceneEx, ok := s.ExtraData.(*BaccaratSceneData); ok {
		//同步开奖结果
		//winZone二进制标志 第1位:和 第2位:庄赢 第3位:闲赢 第4位:庄家对 第5位:庄家对
		pack := &proto_server.GWGameStateLog{
			SceneId: proto.Int(s.SceneId),
			GameLog: proto.Int(sceneEx.winZone),
			LogCnt:  proto.Int(len(sceneEx.trend100Cur)),
		}
		s.SendToWorld(int(proto_server.SSPacketID_PACKET_GW_GAMESTATELOG), pack)
	}
}
func (this *SceneOpenCardStateBaccarat) OnTick(s *base.Scene) {
	this.SceneBaseStateBaccarat.OnTick(s)
	if sceneEx, ok := s.ExtraData.(*BaccaratSceneData); ok {
		if time.Now().Sub(sceneEx.StateStartTime) > rule.BaccaratOpenCardTimeout {
			s.ChangeSceneState(rule.BaccaratSceneStateBilled)
		}
	}
}

//////////////////////////////////////////////////////////////
//结算状态
//////////////////////////////////////////////////////////////
type SceneBilledStateBaccarat struct {
	SceneBaseStateBaccarat
}

func (this *SceneBilledStateBaccarat) GetState() int {
	return rule.BaccaratSceneStateBilled
}
func (this *SceneBilledStateBaccarat) CanChangeTo(s base.SceneState) bool {
	switch s.GetState() {
	case rule.BaccaratSceneStateStakeAnt:
		return true
	}
	return false
}
func (this *SceneBilledStateBaccarat) CanChangeCoinScene(s *base.Scene, p *base.Player) bool {
	return true
}
func (this *SceneBilledStateBaccarat) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if this.SceneBaseStateBaccarat.OnPlayerOp(s, p, opcode, params) {
		return true
	}
	return true
}
func (this *SceneBilledStateBaccarat) OnEnter(s *base.Scene) {
	this.SceneBaseStateBaccarat.OnEnter(s)
	if sceneEx, ok := s.ExtraData.(*BaccaratSceneData); ok {
		//result := sceneEx.GetResult()
		sceneEx.AddLoopNum()
		//通知客户端状态
		pack := &proto_baccarat.SCBaccaratRoomState{
			State: proto.Int(this.GetState()),
		}
		proto.SetDefaults(pack)
		s.Broadcast(int(proto_baccarat.BaccaratPacketID_PACKET_SC_BACCARAT_ROOMSTATE), pack, 0)
		logger.Logger.Tracef("BJL Smart %v", sceneEx.smartSuccess)
		// 水池上下文环境
		sceneEx.CpCtx = base.CoinPoolMgr.GetCoinPoolCtx(sceneEx.Platform, sceneEx.GetGameFreeId(), sceneEx.GroupId)
		sceneEx.CpCtx.Controlled = sceneEx.GetCpControlled()

		playerTotalGain := int64(0)
		sysGain := int64(0)
		olTotalGain := int64(0)
		olTotalBet := 0
		//result := sceneEx.CalcuResult()
		var bigWinner *BaccaratPlayerData
		//计算出所有位置的输赢
		count := len(sceneEx.seats)

		bankerWinZone := make(map[int]int) //庄家每个区域输赢的钱 税前
		bankerWin := int64(0)              //庄家赢的钱 税前
		bankerTax := int64(0)              //庄家税收

		for i := 0; i < count; i++ {
			playerEx := sceneEx.seats[i]
			if playerEx != nil {
				if sceneEx.bankerSnId == playerEx.SnId {
					continue
				}
				playerEx.gainCoin = 0 //玩家金币变动 税后 不带下注额
				//计算玩家输赢及税
				for e := range playerEx.betInfo {
					if playerEx.betInfo[e] <= 0 {
						continue
					}
					if e&sceneEx.winZone == e {
						nowWinCoin := int64(float64(playerEx.betInfo[e]) * (BACCARAT_TheOdds[e] - 1)) //赢的金额 扣税扣这一部分
						tax := int64(float64(nowWinCoin) * float64(s.DbGameFree.GetTaxRate()) / 10000)
						playerEx.winCoin[e] = int(nowWinCoin) //实际赢的金额 税前
						playerEx.taxCoin += tax
						playerEx.gainCoin += nowWinCoin //实际赢的金额 税前

						bankerWinZone[e] -= int(nowWinCoin)
						bankerWin -= nowWinCoin
					} else {
						//开和 庄闲退压注额
						if rule.BACCARAT_ZONE_TIE&sceneEx.winZone == rule.BACCARAT_ZONE_TIE &&
							(e == rule.BACCARAT_ZONE_PLAYER || e == rule.BACCARAT_ZONE_BANKER) {
						} else {
							playerEx.gainCoin -= int64(playerEx.betInfo[e])
							playerEx.winCoin[e] -= playerEx.betInfo[e]

							tax := int64(float64(playerEx.betInfo[e]) * float64(s.DbGameFree.GetTaxRate()) / 10000)
							bankerWinZone[e] += playerEx.betInfo[e]
							bankerWin += int64(playerEx.betInfo[e])
							bankerTax += tax
						}
					}
				}

				playerEx.winorloseCoin = playerEx.gainCoin

				//玩家实际变化的金币
				playerEx.gainWinLost = playerEx.gainCoin - playerEx.taxCoin
				if playerEx.gainWinLost > 0 {
					playerEx.winRecord = append(playerEx.winRecord, 1)
				} else {
					playerEx.winRecord = append(playerEx.winRecord, 0)
				}
				sysGain += playerEx.gainWinLost
				if playerEx.Pos == BACCARAT_OLPOS {
					olTotalGain += playerEx.gainWinLost
				}
				playerEx.betBigRecord = append(playerEx.betBigRecord, playerEx.betTotal)

				//保留20条数据，使用切片，先进先出
				playerEx.RecalcuLatestBet20()

				//记录玩家玩的次数、输赢次数
				if playerEx.betTotal > 0 {
					playerEx.GameTimes++
					if playerEx.gainWinLost > 0 {
						playerEx.SetWinTimes(playerEx.GetWinTimes() + 1)
					} else {
						playerEx.SetLostTimes(playerEx.GetLostTimes() + 1)
					}
				}
				playerEx.RecalcuLatestWin20()
				if playerEx.taxCoin > 0 {
					//累积税收
					playerEx.AddServiceFee(playerEx.taxCoin)
				}
				if playerEx.gainWinLost > 0 {
					if playerEx.Pos < BACCARAT_SELFPOS {
						playerEx.AddCoin(playerEx.gainWinLost, common.GainWay_HundredSceneWin, base.SyncFlag_Broadcast, "system", s.GetSceneName())
					} else {
						playerEx.AddCoin(playerEx.gainWinLost, common.GainWay_HundredSceneWin, base.SyncFlag_ToClient, "system", s.GetSceneName())
					}
					//累积税收
					playerEx.AddServiceFee(playerEx.taxCoin)
				} else {
					if playerEx.Pos < BACCARAT_SELFPOS {
						playerEx.AddCoin(playerEx.gainWinLost, common.GainWay_HundredSceneLost, base.SyncFlag_Broadcast, "system", s.GetSceneName())
					} else {
						playerEx.AddCoin(playerEx.gainWinLost, common.GainWay_HundredSceneLost, base.SyncFlag_ToClient, "system", s.GetSceneName())
					}
				}

				////上报游戏事件
				//if playerEx.betTotal != 0 {
				//	playerEx.ReportGameEvent(playerEx.gainWinLost, playerEx.taxCoin, playerEx.betTotal)
				//}
				if (playerEx.Coin) >= int64(sceneEx.DbGameFree.GetBetLimit()) &&
					playerEx.Coin >= int64(sceneEx.DbGameFree.GetOtherIntParams()[0]) {
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
				//统计大赢家
				if playerEx.gainWinLost > 0 {
					if bigWinner == nil {
						bigWinner = playerEx
					} else {
						if bigWinner.gainWinLost < playerEx.gainWinLost {
							bigWinner = playerEx
						}
					}
				}
				//统计玩家的流水
				if !playerEx.IsRob {
					//本局玩家总产出
					playerTotalGain += int64(playerEx.gainWinLost)
				}
				if playerEx.betTotal > 0 {
					playerEx.SaveSceneCoinLog(playerEx.GetCurrentCoin(), playerEx.gainWinLost,
						playerEx.Coin, playerEx.betTotal, playerEx.taxCoin, playerEx.winorloseCoin, 0, 0)
				}

				//统计投入产出
				if playerEx.gainWinLost != 0 {
					//赔率统计
					playerEx.Statics(sceneEx.KeyGameId, strconv.Itoa(int(sceneEx.GetGameFreeId())), int64(playerEx.gainWinLost), true)
				}
			}
		}
		banker := sceneEx.players[sceneEx.bankerSnId]
		if sceneEx.bankerSnId != -1 && banker != nil {
			logger.Logger.Tracef("庄家收益 %v  税收 %v .", bankerWin, bankerTax)

			//庄家赋值
			for i := uint(1); int(i) < rule.BACCARAT_ZONE_MAX; {
				banker.winCoin[int(i)] = bankerWinZone[int(i)]
				i = i << 1
			}
			for k, v := range sceneEx.betInfo {
				banker.betInfo[k] = v
				banker.betTotal += int64(v)
				banker.TotalBet += int64(v)
			}
			banker.gainCoin = bankerWin
			banker.taxCoin = bankerTax
			banker.gainWinLost = bankerWin - bankerTax

			//庄家输赢记录
			if banker.gainWinLost > 0 {
				banker.winRecord = append(banker.winRecord, 1)
			} else if banker.gainWinLost <= 0 {
				banker.winRecord = append(banker.winRecord, 0)
				if -banker.gainWinLost > banker.Coin {
					banker.gainWinLost = -banker.Coin
				}
			}

			//庄家金币变化
			if banker.gainWinLost != 0 {
				banker.winorloseCoin = banker.gainCoin
				iswin := common.GainWay_HundredSceneWin
				if banker.gainWinLost < 0 {
					iswin = common.GainWay_HundredSceneLost
				}
				banker.AddCoin(banker.gainWinLost, int32(iswin), base.SyncFlag_Broadcast, "system", s.GetSceneName())
			}
		}

		var constBills []*proto_baccarat.BaccaratBill
		if sceneEx.winTop1 != nil && sceneEx.winTop1.betTotal != 0 { //神算子
			constBills = append(constBills, &proto_baccarat.BaccaratBill{
				SnId:     proto.Int32(sceneEx.winTop1.SnId),
				Coin:     proto.Int64(sceneEx.winTop1.Coin),
				GainCoin: proto.Int64(sceneEx.winTop1.gainWinLost),
			})
		}
		if olTotalBet != 0 { //在线玩家
			constBills = append(constBills, &proto_baccarat.BaccaratBill{
				SnId:     proto.Int(0), //在线玩家约定snid=0
				Coin:     proto.Int64(olTotalGain),
				GainCoin: proto.Int64(olTotalGain),
			})
		} else {
			if sysGain < 0 {
				constBills = append(constBills, &proto_baccarat.BaccaratBill{
					SnId:     proto.Int(0), //在线玩家约定snid=0
					Coin:     proto.Int64(sysGain * -1),
					GainCoin: proto.Int64(sysGain * -1),
				})
			}
		}
		for i := 0; i < BACCARAT_RICHTOP5; i++ { //富豪前5名
			if sceneEx.betTop5[i] != nil && sceneEx.betTop5[i].betTotal != 0 {
				constBills = append(constBills, &proto_baccarat.BaccaratBill{
					SnId:     proto.Int32(sceneEx.betTop5[i].SnId),
					Coin:     proto.Int64(sceneEx.betTop5[i].Coin),
					GainCoin: proto.Int64(sceneEx.betTop5[i].gainWinLost),
				})
			}
		}
		for i := 0; i < count; i++ {
			playerEx := sceneEx.seats[i]
			if playerEx.IsOnLine() && !playerEx.IsMarkFlag(base.PlayerState_Leave) {
				pack := &proto_baccarat.SCBaccaratBilled{
					LoopNum: proto.Int(sceneEx.LoopNum),
				}
				pack.BillData = append(pack.BillData, constBills...)
				if playerEx.betTotal != 0 {
					pack.BillData = append(pack.BillData, &proto_baccarat.BaccaratBill{
						SnId:     proto.Int32(playerEx.SnId),
						Coin:     proto.Int64(playerEx.Coin),
						GainCoin: proto.Int64(playerEx.gainWinLost),
					})
				}
				//庄家输赢数据
				if sceneEx.bankerSnId != -1 && banker != nil {
					pack.BillData = append(pack.BillData, &proto_baccarat.BaccaratBill{
						SnId:     proto.Int32(banker.SnId),
						Coin:     proto.Int64(banker.Coin),
						GainCoin: proto.Int64(banker.gainWinLost),
					})
				}
				if bigWinner != nil {
					pack.BigWinner = &proto_baccarat.BaccaratBigWinner{
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
				playerEx.SendToClient(int(proto_baccarat.BaccaratPacketID_PACKET_SC_BACCARAT_GAMEBILLED), pack)
			}
		}

		//统计参与游戏次数
		if !sceneEx.Testing {
			var playerCtxs []*proto_server.PlayerCtx
			for _, p := range sceneEx.seats {
				if p == nil || p.IsRob || p.betTotal == 0 {
					continue
				}
				playerCtxs = append(playerCtxs, &proto_server.PlayerCtx{SnId: proto.Int32(p.SnId), Coin: proto.Int64(p.Coin)})
			}
			//if len(playerCtxs) > 0 {
			//	pack := &proto_server.GWSceneEnd{
			//		GameFreeId: proto.Int32(sceneEx.DbGameFree.GetId()),
			//		Players:    playerCtxs,
			//	}
			//	proto.SetDefaults(pack)
			//	sceneEx.SendToWorld(int(proto_baccarat.MmoPacketID_PACKET_GW_SCENEEND), pack)
			//}
		}

		if sceneEx.bankerSnId != -1 && banker != nil {
			banker.winRecord = []int64{}
			banker.betBigRecord = []int64{}
			banker.lately20Bet = 0
			banker.lately20Win = 0
			//连庄次数+1
			sceneEx.bankerTimes++
		}

		if !sceneEx.Testing && sceneEx.bIntervention {
			//curCoin := base.CoinPoolMgr.LoadCoin(sceneEx.GetGameFreeId(), sceneEx.Platform, sceneEx.GroupId)
			//playerData := make(map[string]int64)
			//for _, value := range sceneEx.players {
			//	if value.IsRob {
			//		continue
			//	}
			//	playerData[strconv.Itoa(int(value.SnId))] = int64(value.gainWinLost)
			//}
			//task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
			//	model.NewRbInterventionLog(sceneEx.webUser, result, sysGain, playerData, result, curCoin)
			//	return nil
			//}), nil, "NewRbInterventionLog").Start()
		}

		//统计牌局详细记录，在mgo中写入
		hundredType := make([]model.HundredType, len(BaccaratZoneMap))
		var isSave bool //有真人参与保存牌局
		if sceneEx.bankerSnId != -1 && banker != nil {
			//目前只有玩家可以上庄
			if banker.betTotal > 0 && !banker.IsRob {
				isSave = true
			}
		}
		areaIndex := 0
		for k := range BaccaratZoneMap {
			b := model.HundredType{
				IsSmartOperation: sceneEx.smartSuccess,
			}
			switch k {
			case rule.BACCARAT_ZONE_TIE:
				b.RegionId = 0
				//和区域把庄闲牌放进去
				b.CardsInfo = []int32{-1}
			case rule.BACCARAT_ZONE_BANKER:
				b.RegionId = 1
				b.CardsInfo = sceneEx.cards[3:]
			case rule.BACCARAT_ZONE_PLAYER:
				b.RegionId = 2
				b.CardsInfo = sceneEx.cards[:3]
			case rule.BACCARAT_ZONE_BANKER_DOUBLE:
				b.RegionId = 3
				if sceneEx.cards[3]%13 == sceneEx.cards[4]%13 {
					b.CardsInfo = sceneEx.cards[3:]
				} else {
					b.CardsInfo = []int32{-1}
				}
			case rule.BACCARAT_ZONE_PLAYER_DOUBLE:
				b.RegionId = 4
				if sceneEx.cards[0]%13 == sceneEx.cards[1]%13 {
					b.CardsInfo = sceneEx.cards[:3]
				} else {
					b.CardsInfo = []int32{-1}
				}
			default:
				b.RegionId = 0
			}
			//计算区域输赢
			if sceneEx.winZone&k == k {
				b.IsWin = 1
			} else {
				if rule.BACCARAT_ZONE_TIE&sceneEx.winZone == rule.BACCARAT_ZONE_TIE &&
					(k == rule.BACCARAT_ZONE_PLAYER || k == rule.BACCARAT_ZONE_BANKER) {
					//开和 庄闲算和
					b.IsWin = 0
				} else {
					b.IsWin = -1
				}
			}
			if sceneEx.betInfo[k] < 0 {
				//没人压注
				hundredType[areaIndex] = b
				areaIndex++
				continue
			}
			playNum, index := 0, 0
			for _, o_player := range sceneEx.players {
				if o_player != nil && o_player.betInfo[k] > 0 && o_player.SnId != sceneEx.bankerSnId {
					if banker != nil && !banker.IsRob {
						//真人玩家坐庄时 需要统计机器人
						playNum++
						isSave = true
					} else if !o_player.IsRob {
						//非真人玩家坐庄，不统计机器人
						playNum++
						isSave = true
					}
				}
			}
			if playNum <= 0 {
				//没真人压注
				hundredType[areaIndex] = b
				areaIndex++
				continue
			}
			person := make([]model.HundredPerson, playNum, playNum)
			for _, player := range sceneEx.players {
				if player != nil && player.betInfo[k] > 0 && player.SnId != sceneEx.bankerSnId {
					isSave = true
					pe := model.HundredPerson{
						UserId:       player.SnId,
						UserBetTotal: int64(player.betInfo[k]),
						ChangeCoin:   int64(player.winCoin[k]),
						BeforeCoin:   player.GetCurrentCoin(),
						AfterCoin:    player.Coin,
						IsRob:        player.IsRob,
						IsFirst:      sceneEx.IsPlayerFirst(player.Player),
						WBLevel:      player.WBLevel,
						Result:       player.result,
					}
					betDetail, ok := player.betDetailOrderInfo[k]
					if ok {
						pe.UserBetTotalDetail = betDetail
					}
					if banker != nil && !banker.IsRob {
						person[index] = pe
						index++
					} else if !player.IsRob {
						person[index] = pe
						index++
					}
				}
			}
			b.PlayerData = person
			hundredType[areaIndex] = b
			areaIndex++
		}

		//最后单独添加庄记录
		if banker != nil && banker.betTotal > 0 {
			win := -1
			if banker.gainWinLost > 0 {
				win = 1
			}
			bankerHundredType := model.HundredType{
				RegionId:         int32(-1),
				IsWin:            win,
				CardsInfo:        []int32{-1},
				IsSmartOperation: sceneEx.smartSuccess,
			}
			hundredPersons := make([]model.HundredPerson, 1, 1)
			pe := model.HundredPerson{
				UserId:       banker.SnId,
				UserBetTotal: banker.betTotal,
				ChangeCoin:   banker.gainWinLost,
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

		if isSave {
			info, err := model.MarshalGameNoteByHUNDRED(hundredType)
			if err == nil {
				for _, o_player := range sceneEx.players {
					if o_player != nil && !o_player.IsRob && o_player.betTotal > 0 {
						totalin, totalout := int64(0), int64(0)
						if o_player.gainWinLost > 0 {
							totalout = o_player.gainWinLost + o_player.taxCoin
						} else {
							//输的人的税收包含在 玩家输的钱中
							totalin = -(o_player.gainWinLost + o_player.taxCoin)
						}

						validFlow := totalin + totalout
						validBet := common.AbsI64(totalin - totalout)
						winAmountNoAnyTax := o_player.gainWinLost + o_player.betTotal
						sceneEx.SaveGamePlayerListLog(o_player.SnId,
							base.GetSaveGamePlayerListLogParam(o_player.Platform, o_player.Channel, o_player.BeUnderAgentCode,
								o_player.PackageID, sceneEx.logicId, o_player.InviterId, totalin, totalout, o_player.taxCoin, 0,
								int64(o_player.betTotal), winAmountNoAnyTax, validFlow, validBet, sceneEx.IsPlayerFirst(o_player.Player),
								false))
					}
				}
				trends := []string{}
				if len(sceneEx.trend20Lately) > 0 {
					for _, value := range sceneEx.trend20Lately {
						if value&(1<<0) != 0 {
							trends = append(trends, "和")
						} else if value&(1<<1) != 0 {
							trends = append(trends, "庄")
						} else if value&(1<<2) != 0 {
							trends = append(trends, "闲")
						}
					}
				}
				trend20Lately, _ := json.Marshal(trends)
				sceneEx.SaveGameDetailedLog(sceneEx.logicId, string(info), &base.GameDetailedParam{
					Trend20Lately: string(trend20Lately),
				})
			}
		}

		systemCoinOut := sceneEx.SystemCoinOutByWinZone()
		s.SetSystemCoinOut(int64(systemCoinOut))
		//更新金币池
		if systemCoinOut > 0 {
			//系统赢钱  水池进钱
			base.CoinPoolMgr.PushCoin(sceneEx.GetGameFreeId(), sceneEx.GroupId, sceneEx.Platform, systemCoinOut)
		} else if systemCoinOut < 0 {
			//系统输钱  水池出钱
			base.CoinPoolMgr.PopCoin(sceneEx.GetGameFreeId(), sceneEx.GroupId, sceneEx.Platform, -systemCoinOut)
		}

		if !sceneEx.Testing {
			gwPlayerBet := &proto_server.GWPlayerBet{
				GameFreeId: proto.Int32(sceneEx.DbGameFree.GetId()),
				RobotGain:  proto.Int64(int64(sceneEx.GetSystemCoinOut())),
			}
			for _, p := range sceneEx.seats {
				//调整为和龙虎一样，庄闲对压，取绝对值
				if p == nil || p.IsRob {
					continue
				}

				t := p.betTotal
				t -= int64(p.betInfo[rule.BACCARAT_ZONE_BANKER])
				t -= int64(p.betInfo[rule.BACCARAT_ZONE_PLAYER])

				t += int64(math.Abs(float64(p.betInfo[rule.BACCARAT_ZONE_BANKER] -
					p.betInfo[rule.BACCARAT_ZONE_PLAYER])))

				if t == 0 && p.betTotal == 0 {
					continue
				}

				playerBet := &proto_server.PlayerBet{
					SnId: proto.Int32(p.SnId),
					Bet:  proto.Int64(t),
					Gain: proto.Int64(p.gainWinLost + p.taxCoin),
					Tax:  proto.Int64(p.taxCoin),
				}
				gwPlayerBet.PlayerBets = append(gwPlayerBet.PlayerBets, playerBet)
			}
			if len(gwPlayerBet.PlayerBets) > 0 {
				proto.SetDefaults(gwPlayerBet)
				sceneEx.SendToWorld(int(proto_server.SSPacketID_PACKET_GW_PLAYERBET), gwPlayerBet)
			}
		} //统计下注数，并将消息发到世界服

		sceneEx.DealyTime = int64(common.RandFromRange(0, 2000)) * int64(time.Millisecond)
	}
}
func (this *SceneBilledStateBaccarat) OnLeave(s *base.Scene) {
	if sceneEx, ok := s.ExtraData.(*BaccaratSceneData); ok {
		sceneEx.logicId, _ = model.AutoIncGameLogId()
		//清理掉线或者暂离的玩家
		for _, p := range sceneEx.players {
			if p == nil {
				continue
			}
			//离开结算阶段 设置当前剩余金额
			p.SetCurrentCoin(p.Coin)
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
			BaccaratSendSeatInfo(s, sceneEx)
			sceneEx.constSeatKey = seatKey
		}

		sceneEx.SendRobotUpBankerList()

		if s.CheckNeedDestroy() {
			sceneEx.SceneDestroy(true)
		}
	}
}
func (this *SceneBilledStateBaccarat) OnTick(s *base.Scene) {
	this.SceneBaseStateBaccarat.OnTick(s)
	if sceneEx, ok := s.ExtraData.(*BaccaratSceneData); ok {
		if time.Now().Sub(sceneEx.StateStartTime) > rule.BaccaratBilledTimeout+time.Duration(sceneEx.DealyTime) {
			s.ChangeSceneState(rule.BaccaratSceneStateStakeAnt)
		}
	}
}

func init() {
	ScenePolicyBaccaratSington.RegisteSceneState(&SceneStakeAntStateBaccarat{})
	ScenePolicyBaccaratSington.RegisteSceneState(&SceneStakeStateBaccarat{})
	ScenePolicyBaccaratSington.RegisteSceneState(&SceneOpenCardAntStateBaccarat{})
	ScenePolicyBaccaratSington.RegisteSceneState(&SceneOpenCardStateBaccarat{})
	ScenePolicyBaccaratSington.RegisteSceneState(&SceneBilledStateBaccarat{})
	core.RegisteHook(core.HOOK_BEFORE_START, func() error {
		base.RegisteScenePolicy(common.GameId_Baccarat, 0, ScenePolicyBaccaratSington)
		base.RegisteScenePolicy(common.GameId_Baccarat, 1, ScenePolicyBaccaratSington)
		return nil
	})
}
