package redvsblack

import (
	"encoding/json"
	"fmt"
	"games.yol.com/win88/common"
	. "games.yol.com/win88/gamerule/redvsblack"
	"games.yol.com/win88/gamesrv/base"
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/player"
	"games.yol.com/win88/protocol/redvsblack"
	"games.yol.com/win88/protocol/server"
	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/timer"
	"time"
)

/*
pos:1-6表示座位的6个位置
	7表示我的位置
    8表示在线的位置（所有其它玩家）
*/

var ScenePolicyRedVsBlackSington = &ScenePolicyRedVsBlack{}

type ScenePolicyRedVsBlack struct {
	base.BaseScenePolicy
	states [RedVsBlackSceneStateMax]base.SceneState
}

//创建场景扩展数据
func (this *ScenePolicyRedVsBlack) CreateSceneExData(s *base.Scene) interface{} {
	sceneEx := NewRedVsBlackSceneData(s)
	if sceneEx != nil {
		if sceneEx.init() {
			s.ExtraData = sceneEx
		}
	}
	return sceneEx
}

//创建玩家扩展数据
func (this *ScenePolicyRedVsBlack) CreatePlayerExData(s *base.Scene, p *base.Player) interface{} {
	playerEx := &RedVsBlackPlayerData{Player: p}
	if playerEx != nil {
		p.ExtraData = playerEx
	}
	return playerEx
}

func RedVsBlackBatchSendBet(sceneEx *RedVsBlackSceneData, force bool) {
	needSend := false
	pack := &redvsblack.SCRedVsBlackSendBet{}
	var olBetChips [RVSB_ZONE_MAX]int64
	olBetInfo := &redvsblack.RedVsBlackBetInfo{
		SnId: proto.Int(0),
	}
	for _, playerEx := range sceneEx.seats {
		if playerEx.Pos == RVSB_OLPOS {
			for i := 0; i < RVSB_ZONE_MAX; i++ {
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
		for i := 0; i < RVSB_ZONE_MAX; i++ {
			pack.TotalChips = append(pack.TotalChips, int64(sceneEx.betInfo[i]))
		}
		proto.SetDefaults(pack)
		logger.Logger.Trace("RedVsBlackBatchSendBet", pack)
		sceneEx.Broadcast(int(redvsblack.RedVsBlackPacketID_PACKET_SC_RVSB_SENDBET), pack, 0)
	}
}

//场景开启事件
func (this *ScenePolicyRedVsBlack) OnStart(s *base.Scene) {
	logger.Logger.Trace("(this *ScenePolicyRedVsBlack) OnStart, sceneId=", s.SceneId)
	sceneEx := NewRedVsBlackSceneData(s)
	if sceneEx != nil {
		if sceneEx.init() {
			if sceneEx.hBatchSend != timer.TimerHandle(0) {
				timer.StopTimer(sceneEx.hBatchSend)
				sceneEx.hBatchSend = timer.TimerHandle(0)
			}
			//批量发送筹码
			if hNext, ok := common.DelayInvake(func() {
				if sceneEx.SceneState.GetState() != RedVsBlackSceneStateStake {
					return
				}
				RedVsBlackBatchSendBet(sceneEx, false)

			}, nil, RedVsBlackBatchSendBetTimeout, -1); ok {
				sceneEx.hBatchSend = hNext
			}

			s.ExtraData = sceneEx
			s.ChangeSceneState(RedVsBlackSceneStateStakeAnt)
		}
	}
}

//场景关闭事件
func (this *ScenePolicyRedVsBlack) OnStop(s *base.Scene) {
	logger.Logger.Trace("(this *ScenePolicyRedVsBlack) OnStop , sceneId=", s.SceneId)
	if sceneEx, ok := s.ExtraData.(*RedVsBlackSceneData); ok {
		if sceneEx.hRunRecord != timer.TimerHandle(0) {
			timer.StopTimer(sceneEx.hRunRecord)
			sceneEx.hRunRecord = timer.TimerHandle(0)
		}
		sceneEx.SaveData()
	}
}

//场景心跳事件
func (this *ScenePolicyRedVsBlack) OnTick(s *base.Scene) {
	if s == nil {
		return
	}
	if s.SceneState != nil {
		s.SceneState.OnTick(s)
	}
}

func (this *ScenePolicyRedVsBlack) GetPlayerNum(s *base.Scene) int32 {
	if s == nil {
		return 0
	}
	if sceneEx, ok := s.ExtraData.(*RedVsBlackSceneData); ok {
		return int32(len(sceneEx.seats))
	}
	return 0
}

//玩家进入事件
func (this *ScenePolicyRedVsBlack) OnPlayerEnter(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyRedVsBlack) OnPlayerEnter, sceneId=", s.SceneId, " player=", p.Name)
	if sceneEx, ok := s.ExtraData.(*RedVsBlackSceneData); ok {
		playerEx := &RedVsBlackPlayerData{Player: p}
		if playerEx != nil {
			playerEx.Clean()

			playerEx.Pos = RVSB_OLPOS

			sceneEx.seats = append(sceneEx.seats, playerEx)

			sceneEx.players[p.SnId] = playerEx

			p.ExtraData = playerEx

			//进房时金币低于下限,状态切换到观众
			if p.Coin < int64(sceneEx.DbGameFree.GetBetLimit()) || p.Coin < int64(sceneEx.DbGameFree.GetOtherIntParams()[0]) {
				p.MarkFlag(base.PlayerState_GameBreak)
			}

			//给自己发送房间信息
			RedVsBlackSendRoomInfo(s, p, sceneEx, playerEx)

			s.FirePlayerEvent(p, base.PlayerEventEnter, nil)
		}
	}
}

//玩家离开事件
func (this *ScenePolicyRedVsBlack) OnPlayerLeave(s *base.Scene, p *base.Player, reason int) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyRedVsBlack) OnPlayerLeave, sceneId=", s.SceneId, " player=", p.Name)
	if sceneEx, ok := s.ExtraData.(*RedVsBlackSceneData); ok {
		if this.CanChangeCoinScene(s, p) {
			sceneEx.OnPlayerLeave(p, reason)
			s.FirePlayerEvent(p, base.PlayerEventLeave, []int64{int64(reason)})
		}
	}
}

//玩家掉线
func (this *ScenePolicyRedVsBlack) OnPlayerDropLine(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyRedVsBlack) OnPlayerDropLine, sceneId=", s.SceneId, " player=", p.Name)
	s.FirePlayerEvent(p, base.PlayerEventDropLine, nil)
	if sceneEx, ok := s.ExtraData.(*RedVsBlackSceneData); ok {
		if sceneEx.Gaming {
			return
		}
	}
}

//玩家重连
func (this *ScenePolicyRedVsBlack) OnPlayerRehold(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyRedVsBlack) OnPlayerRehold, sceneId=", s.SceneId, " player=", p.Name)
	if sceneEx, ok := s.ExtraData.(*RedVsBlackSceneData); ok {
		if playerEx, ok := p.ExtraData.(*RedVsBlackPlayerData); ok {
			//发送房间信息给自己
			RedVsBlackSendRoomInfo(s, p, sceneEx, playerEx)

			s.FirePlayerEvent(p, base.PlayerEventRehold, nil)
		}
	}
}

//玩家返回房间
func (this *ScenePolicyRedVsBlack) OnPlayerReturn(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyRedVsBlack) OnPlayerReturn, sceneId=", s.SceneId, " player=", p.Name)
	if sceneEx, ok := s.ExtraData.(*RedVsBlackSceneData); ok {
		if playerEx, ok := p.ExtraData.(*RedVsBlackPlayerData); ok {
			//发送房间信息给自己
			RedVsBlackSendRoomInfo(s, p, sceneEx, playerEx)

			s.FirePlayerEvent(p, base.PlayerEventReturn, nil)
		}
	}
}

//玩家操作
func (this *ScenePolicyRedVsBlack) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if s == nil || p == nil {
		return false
	}
	if s.SceneState != nil {
		return s.SceneState.OnPlayerOp(s, p, opcode, params)
	}
	return true
}

//玩家事件
func (this *ScenePolicyRedVsBlack) OnPlayerEvent(s *base.Scene, p *base.Player, evtcode int, params []int64) {
	if s == nil || p == nil {
		return
	}
	if s.SceneState != nil {
		s.SceneState.OnPlayerEvent(s, p, evtcode, params)
	}
}

//是否完成了整个牌局
func (this *ScenePolicyRedVsBlack) IsCompleted(s *base.Scene) bool {
	return false
}

//是否可以强制开始
func (this *ScenePolicyRedVsBlack) IsCanForceStart(s *base.Scene) bool {
	return true
}

//当前状态能否换桌
func (this *ScenePolicyRedVsBlack) CanChangeCoinScene(s *base.Scene, p *base.Player) bool {
	if s == nil || p == nil {
		return true
	}
	if s.SceneState != nil {
		return s.SceneState.CanChangeCoinScene(s, p)
	}
	return true
}
func (this *ScenePolicyRedVsBlack) PacketGameData(s *base.Scene) interface{} {
	if s == nil {
		return nil
	}
	if s.SceneState != nil {
		if sceneEx, ok := s.ExtraData.(*RedVsBlackSceneData); ok {
			switch s.SceneState.GetState() {
			case RedVsBlackSceneStateStakeAnt:
				fallthrough
			case RedVsBlackSceneStateStake:
				lt := int32((RedVsBlackStakeTimeout - time.Now().Sub(sceneEx.StateStartTime)) / time.Second)
				pack := &server.GWDTRoomInfo{
					DCoin:      proto.Int(sceneEx.betInfo[RVSB_ZONE_BLACK] - sceneEx.betInfoRob[RVSB_ZONE_BLACK]),
					TCoin:      proto.Int(sceneEx.betInfo[RVSB_ZONE_RED] - sceneEx.betInfoRob[RVSB_ZONE_RED]),
					NCoin:      proto.Int(sceneEx.betInfo[RVSB_ZONE_LUCKY] - sceneEx.betInfoRob[RVSB_ZONE_LUCKY]),
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
					pack.Players = append(pack.Players, &server.PlayerDTCoin{
						NickName: proto.String(value.Name),
						Snid:     proto.Int32(value.SnId),
						DCoin:    proto.Int(value.betInfo[RVSB_ZONE_BLACK]),
						TCoin:    proto.Int(value.betInfo[RVSB_ZONE_RED]),
						NCoin:    proto.Int(value.betInfo[RVSB_ZONE_LUCKY]),
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
func (this *ScenePolicyRedVsBlack) InterventionGame(s *base.Scene, data interface{}) interface{} {
	if s == nil {
		return nil
	}
	if s.SceneState != nil {
		if sceneEx, ok := s.ExtraData.(*RedVsBlackSceneData); ok {
			if v, ok := data.(base.InterventionResults); ok {
				return sceneEx.ParserResults1(v.Results, v.Webuser)
			}
			switch s.SceneState.GetState() {
			case RedVsBlackSceneStateStakeAnt:
				fallthrough
			case RedVsBlackSceneStateStake:
				if d, ok := data.(base.InterventionData); ok {
					if sceneEx.NumOfGames == int(d.NumOfGames) {
						winFlag, _ := sceneEx.CalcuResult()
						if d.Flag != int32(winFlag) {
							sceneEx.cards[0], sceneEx.cards[1] = sceneEx.cards[1], sceneEx.cards[0]
						}
						sceneEx.bIntervention = true
						sceneEx.webUser = d.Webuser
					}
				}
			default:
			}
		}
	}
	return nil
}
func (this *ScenePolicyRedVsBlack) NotifyGameState(s *base.Scene) {
	if s.GetState() > 2 {
		return
	}
	sec := int((RedVsBlackStakeAntTimeout + RedVsBlackStakeTimeout).Seconds())
	s.SyncGameState(sec, 0)
}

//座位数据
func RedVsBlackCreateSeats(sceneEx *RedVsBlackSceneData) []*redvsblack.RedVsBlackPlayerData {
	var datas []*redvsblack.RedVsBlackPlayerData
	cnt := 0
	const N = RVSB_RICHTOP5 + 1
	var seats [N]*RedVsBlackPlayerData
	if sceneEx.winTop1 != nil { //神算子
		seats[cnt] = sceneEx.winTop1
		cnt++
	}
	for i := 0; i < RVSB_RICHTOP5; i++ {
		if sceneEx.betTop5[i] != nil {
			seats[cnt] = sceneEx.betTop5[i]
			cnt++
		}
	}
	for i := 0; i < N; i++ {
		if seats[i] != nil {
			pd := &redvsblack.RedVsBlackPlayerData{
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
	return datas
}

func RedVsBlackCreateRoomInfoPacket(s *base.Scene, sceneEx *RedVsBlackSceneData, playerEx *RedVsBlackPlayerData) proto.Message {
	pack := &redvsblack.SCRedVsBlackRoomInfo{
		RoomId:    proto.Int(s.SceneId),
		Creator:   proto.Int32(s.GetCreator()),
		GameId:    proto.Int(s.GameId),
		RoomMode:  proto.Int(s.SceneMode),
		AgentId:   proto.Int32(s.GetAgentor()),
		SceneType: proto.Int(s.SceneType),
		Params: []int32{sceneEx.DbGameFree.GetLimitCoin(), sceneEx.DbGameFree.GetMaxCoinLimit(),
			sceneEx.DbGameFree.GetServiceFee(), sceneEx.DbGameFree.GetLowerThanKick(), sceneEx.DbGameFree.GetBaseScore(),
			0, sceneEx.DbGameFree.GetBetLimit()},
		NumOfGames:      proto.Int(sceneEx.NumOfGames),
		State:           proto.Int(s.SceneState.GetState()),
		TimeOut:         proto.Int(s.SceneState.GetTimeout(s)),
		Players:         RedVsBlackCreateSeats(sceneEx),
		Trend100Cur:     sceneEx.trend100Cur,
		Trend20Lately:   sceneEx.trend20Lately,
		Trend20CardKind: sceneEx.trend20CardKindLately,
		OLNum:           proto.Int(len(sceneEx.seats)),
		DisbandGen:      proto.Int(sceneEx.GetDisbandGen()),
		ParamsEx:        s.GetParamsEx(),
		OtherIntParams:  s.DbGameFree.GetOtherIntParams(),
		LoopNum:         proto.Int(sceneEx.LoopNum),
	}
	for _, value := range sceneEx.DbGameFree.GetMaxBetCoin() {
		pack.Params = append(pack.Params, value)
	}
	if playerEx != nil { //自己的数据
		pd := &redvsblack.RedVsBlackPlayerData{
			SnId: proto.Int32(playerEx.SnId),
			Name: proto.String(playerEx.Name),
			Head: proto.Int32(playerEx.Head),
			Sex:  proto.Int32(playerEx.Sex),
			Coin: proto.Int64(playerEx.Coin),
			Pos:  proto.Int(RVSB_SELFPOS),
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
	for i := 0; i < RVSB_ZONE_MAX; i++ {
		if len(sceneEx.betDetailInfo[i]) != 0 {
			total := 0
			for k, v := range sceneEx.betDetailInfo[i] {
				total += k * v
			}
			pack.TotalChips = append(pack.TotalChips, int32(total))
		} else {
			pack.TotalChips = append(pack.TotalChips, 0)
		}
	}
	if playerEx != nil { //自己的数据
		for i := 0; i < RVSB_ZONE_MAX; i++ {
			if len(playerEx.betDetailInfo[i]) != 0 {
				chip := &redvsblack.RedVsBlackZoneChips{
					Zone: proto.Int(i),
				}
				for k, v := range playerEx.betDetailInfo[i] {
					chip.Detail = append(chip.Detail, &redvsblack.RedVsBlackChips{
						Chip:  proto.Int(k),
						Count: proto.Int(v),
					})
				}
				pack.MyChips = append(pack.MyChips, chip)
			}
		}
	}
	if sceneEx.SceneState.GetState() == RedVsBlackSceneStateOpenCard {
		for i := 0; i < 2; i++ {
			if sceneEx.kindOfcards[i] != nil {
				for _, c := range sceneEx.kindOfcards[i].OrderCards {
					pack.Cards = append(pack.Cards, int32(c))
				}
			}
		}
		for i := 0; i < 2; i++ {
			if sceneEx.kindOfcards[i] != nil {
				pack.Cards = append(pack.Cards, int32(sceneEx.kindOfcards[i].Kind))
			}
		}
		pack.Cards = append(pack.Cards, int32(sceneEx.pkResult))
		if sceneEx.luckyKind > CardsKind_Double {
			pack.Cards = append(pack.Cards, 1)
		} else {
			pack.Cards = append(pack.Cards, 0)
		}
	}
	proto.SetDefaults(pack)
	return pack
}

func RedVsBlackSendRoomInfo(s *base.Scene, p *base.Player, sceneEx *RedVsBlackSceneData, playerEx *RedVsBlackPlayerData) {
	pack := RedVsBlackCreateRoomInfoPacket(s, sceneEx, playerEx)
	logger.Logger.Trace("RedVsBlackSendRoomInfo pack:", pack)
	p.SendToClient(int(redvsblack.RedVsBlackPacketID_PACKET_SC_RVSB_ROOMINFO), pack)
}

func RedVsBlackSendSeatInfo(s *base.Scene, sceneEx *RedVsBlackSceneData) {
	pack := &redvsblack.SCRedVsBlackSeats{
		PlayerNum: proto.Int(len(sceneEx.seats)),
		Data:      RedVsBlackCreateSeats(sceneEx),
	}
	proto.SetDefaults(pack)
	logger.Logger.Trace("RedVsBlackSendSeatInfo pack:", pack)
	s.Broadcast(int(redvsblack.RedVsBlackPacketID_PACKET_SC_RVSB_SEATS), pack, 0)
}

//////////////////////////////////////////////////////////////
//状态基类
//////////////////////////////////////////////////////////////
type SceneBaseStateRedVsBlack struct {
}

func (this *SceneBaseStateRedVsBlack) GetTimeout(s *base.Scene) int {
	if sceneEx, ok := s.ExtraData.(*RedVsBlackSceneData); ok {
		return int(time.Now().Sub(sceneEx.StateStartTime) / time.Second)
	}
	return 0
}

func (this *SceneBaseStateRedVsBlack) CanChangeTo(s base.SceneState) bool {
	return true
}

//当前状态能否换桌
func (this *SceneBaseStateRedVsBlack) CanChangeCoinScene(s *base.Scene, p *base.Player) bool {
	if playerEx, ok := p.ExtraData.(*RedVsBlackPlayerData); ok {
		if playerEx.betTotal != 0 {
			playerEx.OpCode = player.OpResultCode_OPRC_Hundred_YouHadBetCannotLeave
			return false
		}
	}
	return true
}

func (this *SceneBaseStateRedVsBlack) OnEnter(s *base.Scene) {
	if sceneEx, ok := s.ExtraData.(*RedVsBlackSceneData); ok {
		sceneEx.StateStartTime = time.Now()
	}
}

func (this *SceneBaseStateRedVsBlack) OnLeave(s *base.Scene) {
}

func (this *SceneBaseStateRedVsBlack) OnTick(s *base.Scene) {
}

//发送玩家操作情况
func (this *SceneBaseStateRedVsBlack) SendSCPlayerOp(s *base.Scene, p *base.Player, pos int, opcode int, opRetCode redvsblack.OpResultCode, params []int64, broadcastall bool) {
	pack := &redvsblack.SCRedVsBlackOp{
		OpCode:    proto.Int(opcode),
		SnId:      proto.Int32(p.SnId),
		OpRetCode: opRetCode,
		Params:    params,
	}
	proto.SetDefaults(pack)
	if broadcastall {
		s.Broadcast(int(redvsblack.RedVsBlackPacketID_PACKET_SC_RVSB_PLAYEROP), pack, 0)
	} else {
		p.SendToClient(int(redvsblack.RedVsBlackPacketID_PACKET_SC_RVSB_PLAYEROP), pack)
	}
}

func (this *SceneBaseStateRedVsBlack) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if sceneEx, ok := s.ExtraData.(*RedVsBlackSceneData); ok {
		switch opcode {
		case RedVsBlackPlayerOpGetOLList: //在线玩家列表
			seats := make([]*RedVsBlackPlayerData, 0, RVSB_OLTOP20+1)
			if sceneEx.winTop1 != nil { //神算子
				seats = append(seats, sceneEx.winTop1)
			}
			count := len(sceneEx.seats)
			topCnt := 0
			for i := 0; i < count && topCnt < RVSB_OLTOP20; i++ { //top20
				if sceneEx.seats[i] != sceneEx.winTop1 {
					seats = append(seats, sceneEx.seats[i])
					topCnt++
				}
			}
			pack := &redvsblack.SCRedVsBlackPlayerList{}
			for i := 0; i < len(seats); i++ {
				pack.Data = append(pack.Data, &redvsblack.RedVsBlackPlayerData{
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
			pack.OLNum = int32(len(s.Players))
			proto.SetDefaults(pack)
			p.SendToClient(int(redvsblack.RedVsBlackPacketID_PACKET_SC_RVSB_PLAYERLIST), pack)
			return true
		}
	}
	return false
}

func (this *SceneBaseStateRedVsBlack) OnPlayerEvent(s *base.Scene, p *base.Player, evtcode int, params []int64) {
	if sceneEx, ok := s.ExtraData.(*RedVsBlackSceneData); ok {
		if playerEx, ok := p.ExtraData.(*RedVsBlackPlayerData); ok {
			needResort := false
			switch evtcode {
			case base.PlayerEventEnter:
				if sceneEx.winTop1 == nil {
					needResort = true
				}
				if !needResort {
					for i := 0; i < RVSB_RICHTOP5; i++ {
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
					for i := 0; i < RVSB_RICHTOP5; i++ {
						if sceneEx.betTop5[i] == playerEx {
							needResort = true
							break
						}
					}
				}
			case base.PlayerEventRecharge:
				//oldflag := p.MarkBroadcast(playerEx.Pos < RVSB_SELFPOS)
				p.AddCoin(params[0], common.GainWay_Pay, base.SyncFlag_ToClient, "system", p.GetScene().GetSceneName())
				//p.MarkBroadcast(oldflag)
				if p.Coin >= int64(sceneEx.DbGameFree.GetBetLimit()) && p.Coin >= int64(sceneEx.DbGameFree.GetOtherIntParams()[0]) {
					if p.IsMarkFlag(base.PlayerState_GameBreak) {
						p.UnmarkFlag(base.PlayerState_GameBreak)
						p.SyncFlag(true)
					}
				}
			}
			if needResort {
				seatKey := sceneEx.Resort()
				if seatKey != sceneEx.constSeatKey {
					RedVsBlackSendSeatInfo(s, sceneEx)
					sceneEx.constSeatKey = seatKey
				}
			}
		}
	}
}

//////////////////////////////////////////////////////////////
//押注动画状态
//////////////////////////////////////////////////////////////
type SceneStakeAntStateRedVsBlack struct {
	SceneBaseStateRedVsBlack
}

//获取当前场景状态
func (this *SceneStakeAntStateRedVsBlack) GetState() int {
	return RedVsBlackSceneStateStakeAnt
}

//是否可以改变到其它场景状态
func (this *SceneStakeAntStateRedVsBlack) CanChangeTo(s base.SceneState) bool {
	if s.GetState() == RedVsBlackSceneStateStake {
		return true
	}
	return false
}

//玩家操作
func (this *SceneStakeAntStateRedVsBlack) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if this.SceneBaseStateRedVsBlack.OnPlayerOp(s, p, opcode, params) {
		return true
	}
	return true
}

//场景状态进入
func (this *SceneStakeAntStateRedVsBlack) OnEnter(s *base.Scene) {
	this.SceneBaseStateRedVsBlack.OnEnter(s)
	if sceneEx, ok := s.ExtraData.(*RedVsBlackSceneData); ok {
		logger.Logger.Tracef("(this *Scene) [%v] 场景状态进入 %v", s.SceneId, len(sceneEx.seats))
		sceneEx.Clean()
		sceneEx.poker.Shuffle()
		sceneEx.NumOfGames++
		sceneEx.GameNowTime = time.Now()
		//发牌
		for i := 0; i < 2; i++ {
			for j := 0; j < Hand_CardNum; j++ {
				sceneEx.cards[i][j] = int(sceneEx.poker.Next())
			}
		}
		sceneEx.CalcuResult()
		sceneEx.CheckResults()
		pack := &redvsblack.SCRedVsBlackRoomState{
			State: proto.Int(this.GetState()),
		}
		proto.SetDefaults(pack)
		s.Broadcast(int(redvsblack.RedVsBlackPacketID_PACKET_SC_RVSB_ROOMSTATE), pack, 0)
	}
}

//场景状态离开
func (this *SceneStakeAntStateRedVsBlack) OnLeave(s *base.Scene) {
}

func (this *SceneStakeAntStateRedVsBlack) OnTick(s *base.Scene) {
	this.SceneBaseStateRedVsBlack.OnTick(s)
	if sceneEx, ok := s.ExtraData.(*RedVsBlackSceneData); ok {
		if time.Now().Sub(sceneEx.StateStartTime) > RedVsBlackStakeAntTimeout {
			s.ChangeSceneState(RedVsBlackSceneStateStake)
		}
	}
}

//////////////////////////////////////////////////////////////
//押注状态
//////////////////////////////////////////////////////////////
type SceneStakeStateRedVsBlack struct {
	SceneBaseStateRedVsBlack
}

func (this *SceneStakeStateRedVsBlack) GetState() int {
	return RedVsBlackSceneStateStake
}

func (this *SceneStakeStateRedVsBlack) CanChangeTo(s base.SceneState) bool {
	switch s.GetState() {
	case RedVsBlackSceneStateOpenCardAnt:
		return true
	}
	return false
}

func (this *SceneStakeStateRedVsBlack) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if this.SceneBaseStateRedVsBlack.OnPlayerOp(s, p, opcode, params) {
		return true
	}
	if playerEx, ok := p.ExtraData.(*RedVsBlackPlayerData); ok {
		if sceneEx, ok := s.ExtraData.(*RedVsBlackSceneData); ok {
			//玩家下注
			switch opcode {
			case RedVsBlackPlayerOpBet:
				if len(params) >= 2 {
					betPos := int(params[0]) //下注位置
					if betPos < 0 || betPos >= RVSB_ZONE_MAX {
						return false
					}
					betCoin := int(params[1]) //下流金额
					chips := sceneEx.DbGameFree.GetOtherIntParams()
					if !common.InSliceInt32(chips, int32(params[1])) {
						return false
					}
					//最小面额的筹码
					minChip := int(chips[0])
					//期望下注金额
					expectBetCoin := betCoin
					if p.Coin < int64(sceneEx.DbGameFree.GetBetLimit()) && playerEx.betTotal == 0 {
						logger.Logger.Trace("======提示低于多少不能下注======")
						this.SendSCPlayerOp(s, p, playerEx.Pos, opcode, redvsblack.OpResultCode_OPRC_CoinMustReachTheValue, params, false)
						return false
					}
					if playerEx.betZone != -1 && betPos != RVSB_ZONE_LUCKY && betPos != playerEx.betZone {
						//你只能支持一方势力
						this.SendSCPlayerOp(s, p, playerEx.Pos, opcode, redvsblack.OpResultCode_OPRC_OnlySupportOne, params, false)
						return false
					}
					if betPos != RVSB_ZONE_LUCKY {
						playerEx.betZone = betPos
					}
					total := int64(betCoin + playerEx.betTotal)
					if total <= playerEx.Coin {
						//闲家单门下注总额是否到达上限
						maxBetCoin := sceneEx.DbGameFree.GetMaxBetCoin()
						if ok, coinLimit := playerEx.MaxChipCheck(betPos, betCoin, maxBetCoin); !ok {
							betCoin = int(coinLimit) - playerEx.betInfo[betPos]
							//对齐到最小面额的筹码
							betCoin /= minChip
							betCoin *= minChip
							if betCoin <= 0 {
								//this.SendSCPlayerOp(s, p, playerEx.Pos, opcode, redvsblack.OpResultCode_OPRC_Hundred_EachBetsLimit, params, false)
								msgPack := common.CreateSrvMsg(common.SRVMSG_CODE_DEFAULT, fmt.Sprintf("该门押注金额上限%.2f", float64(coinLimit)/100))
								p.SendToClient(int(player.PlayerPacketID_PACKET_SC_SRVMSG), msgPack)
								return false
							}
						}
					} else {
						this.SendSCPlayerOp(s, p, playerEx.Pos, opcode, redvsblack.OpResultCode_OPRC_CoinIsNotEnough, params, false)
						return false
					}
					playerEx.Trusteeship = 0
					//那一个闲家的下注总额
					//累积总投注额
					playerEx.TotalBet += int64(betCoin)
					playerEx.betTotal += betCoin
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
								sceneEx.betDetailInfo[betPos][chip] = sceneEx.betDetailInfo[betPos][chip] + cntChip
								playerEx.betDetailInfo[betPos][chip] = playerEx.betDetailInfo[betPos][chip] + cntChip
								playerEx.betCacheInfo[betPos][chip] = playerEx.betCacheInfo[betPos][chip] + cntChip
								val -= cntChip * chip
							}
						}
					}
					restCoin := playerEx.Coin - int64(playerEx.betTotal)
					params[1] = int64(betCoin) //修正下当前的实际下注额
					params = append(params, (restCoin))
					//params = append(params, int32(sceneEx.betInfo[betPos]))
					this.SendSCPlayerOp(s, p, playerEx.Pos, opcode, redvsblack.OpResultCode_OPRC_Sucess, params, playerEx.Pos < RVSB_SELFPOS)

					playerEx.betDetailOrderInfo[betPos] = append(playerEx.betDetailOrderInfo[betPos], int64(betCoin))

					//没钱了，要转到观战模式
					if (restCoin) < int64(chips[0]) || (restCoin) < int64(sceneEx.DbGameFree.GetBetLimit()) {
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

func (this *SceneStakeStateRedVsBlack) OnEnter(s *base.Scene) {
	this.SceneBaseStateRedVsBlack.OnEnter(s)
	if sceneEx, ok := s.ExtraData.(*RedVsBlackSceneData); ok {
		sceneEx.CheckResults()
		pack := &redvsblack.SCRedVsBlackRoomState{
			State: proto.Int(this.GetState()),
		}
		proto.SetDefaults(pack)
		s.Broadcast(int(redvsblack.RedVsBlackPacketID_PACKET_SC_RVSB_ROOMSTATE), pack, 0)
	}
}

func (this *SceneStakeStateRedVsBlack) OnLeave(s *base.Scene) {
	this.SceneBaseStateRedVsBlack.OnLeave(s)
	if sceneEx, ok := s.ExtraData.(*RedVsBlackSceneData); ok {
		logger.Logger.Trace("SceneStakeStateRedVsBlack.OnLeave RedVsBlackBatchSendBet")
		RedVsBlackBatchSendBet(sceneEx, true)
		/*var allPlayerBet int
		for i := 0; i < RVSB_ZONE_MAX; i++ {
			allPlayerBet += sceneEx.betInfo[i] - sceneEx.betInfoRob[i]
		}
		base.CoinPoolMgr.PushCoin(sceneEx.gamefreeId, sceneEx.platform, int64(allPlayerBet))*/
		for _, v := range sceneEx.players {
			if v != nil && !v.IsRob && v.betTotal <= 0 {
				v.Trusteeship++
				if v.Trusteeship >= model.GameParamData.NotifyPlayerWatchNum {
					v.SendTrusteeshipTips()
				}
			}
		}
	}
}

func (this *SceneStakeStateRedVsBlack) OnTick(s *base.Scene) {
	this.SceneBaseStateRedVsBlack.OnTick(s)
	if sceneEx, ok := s.ExtraData.(*RedVsBlackSceneData); ok {
		if time.Now().Sub(sceneEx.StateStartTime) > RedVsBlackStakeTimeout {
			s.ChangeSceneState(RedVsBlackSceneStateOpenCardAnt)
		}
	}
}

//////////////////////////////////////////////////////////////
//开牌动画状态
//////////////////////////////////////////////////////////////
type SceneOpenCardAntStateRedVsBlack struct {
	SceneBaseStateRedVsBlack
}

func (this *SceneOpenCardAntStateRedVsBlack) GetState() int {
	return RedVsBlackSceneStateOpenCardAnt
}

func (this *SceneOpenCardAntStateRedVsBlack) CanChangeTo(s base.SceneState) bool {
	switch s.GetState() {
	case RedVsBlackSceneStateOpenCard:
		return true
	}
	return false
}

func (this *SceneOpenCardAntStateRedVsBlack) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if this.SceneBaseStateRedVsBlack.OnPlayerOp(s, p, opcode, params) {
		return true
	}
	return true
}

func (this *SceneOpenCardAntStateRedVsBlack) OnEnter(s *base.Scene) {
	this.SceneBaseStateRedVsBlack.OnEnter(s)
	if sceneEx, ok := s.ExtraData.(*RedVsBlackSceneData); ok {
		if sceneEx.bIntervention == false {
			ok, singlePlayer := sceneEx.IsSingleRegulatePlayer()
			if ok {
				if sceneEx.IsNeedSingleRegulate(singlePlayer) { //本轮结果是否需要单控
					//走闲家调控
					if sceneEx.RegulationXianCard(singlePlayer) {
						singlePlayer.MarkFlag(base.PlayerState_SAdjust)
						singlePlayer.AddAdjustCount(sceneEx.GetGameFreeId())
					} else {
						singlePlayer.result = 0
					}
				} else {
					if singlePlayer.betTotal != 0 {
						singlePlayer.MarkFlag(base.PlayerState_SAdjust)
						singlePlayer.AddAdjustCount(sceneEx.GetGameFreeId())
					} else {
						singlePlayer.result = 0
					}
				}
			} else if !sceneEx.AutoBalance() {
				sceneEx.ChangeCard()
			}
		}
		/*
			const MAXTYR = 100
			sysOutput, luckyOutput := sceneEx.GetAllPlayerWinScore()
			totalOut := sysOutput + luckyOutput
			minOut := totalOut
			minCards := sceneEx.cards
			isOk := base.CoinPoolMgr.IsCoinEnough(sceneEx.gamefreeId, sceneEx.groupId, sceneEx.platform, totalOut)
			for i := 0; i < MAXTYR && !isOk; i++ {
				//先互换
				sceneEx.cards[0], sceneEx.cards[1] = sceneEx.cards[1], sceneEx.cards[0]
				//计算新牌的赔付结果
				sysOutput, luckyOutput = sceneEx.GetAllPlayerWinScore()
				totalOut = sysOutput + luckyOutput
				if totalOut < minOut {
					minOut = totalOut
					minCards = sceneEx.cards
				}

				isOk = base.CoinPoolMgr.IsCoinEnough(sceneEx.gamefreeId, sceneEx.groupId, sceneEx.platform, totalOut)
				if !isOk {
					for sceneEx.poker.Count() < 6 {
						sceneEx.poker.Shuffle()
					}
					//尝试另一组牌
					for i := 0; i < 2; i++ {
						for j := 0; j < redvsblack.Hand_CardNum; j++ {
							sceneEx.cards[i][j] = int(sceneEx.poker.Next())
						}
					}
					//计算新牌的赔付结果
					sysOutput, luckyOutput = sceneEx.GetAllPlayerWinScore()
					totalOut = sysOutput + luckyOutput
					if totalOut < minOut {
						minOut = totalOut
						minCards = sceneEx.cards
					}
				} else {
					break
				}
			}

			if !isOk { //尽可能将系统亏损降到最低
				sceneEx.cards = minCards
				sysOutput, luckyOutput = sceneEx.GetAllPlayerWinScore()
			}
			//更新金币池
			base.CoinPoolMgr.PopCoin(sceneEx.gamefreeId, sceneEx.platform, sysOutput+luckyOutput)*/

		pack := &redvsblack.SCRedVsBlackRoomState{
			State: proto.Int(this.GetState()),
		}
		proto.SetDefaults(pack)
		s.Broadcast(int(redvsblack.RedVsBlackPacketID_PACKET_SC_RVSB_ROOMSTATE), pack, 0)
	}
}

func (this *SceneOpenCardAntStateRedVsBlack) OnLeave(s *base.Scene) {}

func (this *SceneOpenCardAntStateRedVsBlack) OnTick(s *base.Scene) {
	this.SceneBaseStateRedVsBlack.OnTick(s)
	if sceneEx, ok := s.ExtraData.(*RedVsBlackSceneData); ok {
		if time.Now().Sub(sceneEx.StateStartTime) > RedVsBlackOpenCardAntTimeout {
			s.ChangeSceneState(RedVsBlackSceneStateOpenCard)
		}
	}
}

//////////////////////////////////////////////////////////////
//开牌状态
//////////////////////////////////////////////////////////////

type SceneOpenCardStateRedVsBlack struct {
	SceneBaseStateRedVsBlack
}

func (this *SceneOpenCardStateRedVsBlack) GetState() int {
	return RedVsBlackSceneStateOpenCard
}

func (this *SceneOpenCardStateRedVsBlack) CanChangeTo(s base.SceneState) bool {
	switch s.GetState() {
	case RedVsBlackSceneStateBilled:
		return true
	}
	return false
}

func (this *SceneOpenCardStateRedVsBlack) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if this.SceneBaseStateRedVsBlack.OnPlayerOp(s, p, opcode, params) {
		return true
	}
	return true
}

func (this *SceneOpenCardStateRedVsBlack) OnEnter(s *base.Scene) {
	this.SceneBaseStateRedVsBlack.OnEnter(s)
	if sceneEx, ok := s.ExtraData.(*RedVsBlackSceneData); ok {
		result, luckyKind := sceneEx.CalcuResult()
		sceneEx.PushTrend(int32(result), int32(luckyKind))
		logger.Logger.Tracef("3-> sceneEx.CalcuResult() result=%v luckykind=%v", result, luckyKind)
		pack := &redvsblack.SCRedVsBlackRoomState{
			State: proto.Int(this.GetState()),
		}
		for i := 0; i < 2; i++ {
			if sceneEx.kindOfcards[i] != nil {
				for _, c := range sceneEx.kindOfcards[i].OrderCards {
					pack.Params = append(pack.Params, int32(c))
				}
			}
		}
		for i := 0; i < 2; i++ {
			if sceneEx.kindOfcards[i] != nil {
				pack.Params = append(pack.Params, int32(sceneEx.kindOfcards[i].Kind))
			}
		}
		pack.Params = append(pack.Params, int32(result))
		if sceneEx.luckyKind > CardsKind_Double {
			pack.Params = append(pack.Params, 1)
		} else {
			pack.Params = append(pack.Params, 0)
		}
		proto.SetDefaults(pack)
		s.Broadcast(int(redvsblack.RedVsBlackPacketID_PACKET_SC_RVSB_ROOMSTATE), pack, 0)
		logger.Logger.Trace("SceneOpenCardStateRedVsBlack ->", pack)
	}
}
func (this *SceneOpenCardStateRedVsBlack) OnLeave(s *base.Scene) {
	this.SceneBaseStateRedVsBlack.OnLeave(s)
	if sceneEx, ok := s.ExtraData.(*RedVsBlackSceneData); ok {
		//同步开奖结果
		result, luckyKind := sceneEx.CalcuResult()
		pack := &server.GWGameStateLog{
			SceneId: proto.Int(s.SceneId),
			GameLog: proto.Int(result | (luckyKind << 8)),
			LogCnt:  proto.Int(len(sceneEx.trend100Cur)),
		}
		s.SendToWorld(int(server.SSPacketID_PACKET_GW_GAMESTATELOG), pack)
	}
}

func (this *SceneOpenCardStateRedVsBlack) OnTick(s *base.Scene) {
	this.SceneBaseStateRedVsBlack.OnTick(s)
	if sceneEx, ok := s.ExtraData.(*RedVsBlackSceneData); ok {
		if time.Now().Sub(sceneEx.StateStartTime) > RedVsBlackOpenCardTimeout {
			s.ChangeSceneState(RedVsBlackSceneStateBilled)
		}
	}
}

//////////////////////////////////////////////////////////////
//结算状态
//////////////////////////////////////////////////////////////
type SceneBilledStateRedVsBlack struct {
	SceneBaseStateRedVsBlack
}

func (this *SceneBilledStateRedVsBlack) GetState() int {
	return RedVsBlackSceneStateBilled
}

func (this *SceneBilledStateRedVsBlack) CanChangeTo(s base.SceneState) bool {
	switch s.GetState() {
	case RedVsBlackSceneStateStakeAnt:
		return true
	}
	return false
}

func (this *SceneBilledStateRedVsBlack) CanChangeCoinScene(s *base.Scene, p *base.Player) bool {
	return true
}

func (this *SceneBilledStateRedVsBlack) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if this.SceneBaseStateRedVsBlack.OnPlayerOp(s, p, opcode, params) {
		return true
	}
	return true
}

func (this *SceneBilledStateRedVsBlack) OnEnter(s *base.Scene) {
	this.SceneBaseStateRedVsBlack.OnEnter(s)
	if sceneEx, ok := s.ExtraData.(*RedVsBlackSceneData); ok {
		sceneEx.AddLoopNum()
		pack := &redvsblack.SCRedVsBlackRoomState{
			State:  proto.Int(this.GetState()),
			Params: []int32{int32(sceneEx.pkResult), int32(sceneEx.luckyKind)},
		}
		proto.SetDefaults(pack)
		s.Broadcast(int(redvsblack.RedVsBlackPacketID_PACKET_SC_RVSB_ROOMSTATE), pack, 0)

		// 水池上下文环境
		sceneEx.CpCtx = base.CoinPoolMgr.GetCoinPoolCtx(sceneEx.Platform, sceneEx.GetGameFreeId(), sceneEx.GroupId)
		sceneEx.CpCtx.Controlled = sceneEx.CpControlled

		playerTotalGain := int64(0)
		sysGain := int64(0)
		olTotalGain := int64(0)
		olTotalBet := int64(0)
		result, luckyKind := sceneEx.CalcuResult()

		var bigWinner *RedVsBlackPlayerData
		//计算出所有位置的输赢
		count := len(sceneEx.seats)
		for i := 0; i < count; i++ {
			playerEx := sceneEx.seats[i]
			if playerEx != nil {
				var tax int
				var winCoin int32
				var lostCoin int32
				playerEx.gainCoin = 0
				if result >= 0 && result < RVSB_ZONE_MAX {
					if playerEx.betInfo[result] > 0 {
						playerEx.winCoin[result] = (RVSB_TheOdds[result]*100 + 100) * playerEx.betInfo[result] / 100
						playerEx.gainCoin += playerEx.winCoin[result]
						sceneEx.isWin = [3]int{-1, -1, -1} //输赢初始化
						if result == 0 {
							sceneEx.isWin[0] = 1 //黑方赢
						} else {
							sceneEx.isWin[1] = 1 //红方赢
						}
						//税收只算赢钱的
						tax += int(int64(RVSB_TheOdds[result]*playerEx.betInfo[result]) * int64(s.DbGameFree.GetTaxRate()) / 10000)
						winCoin += int32(RVSB_TheOdds[result] * playerEx.betInfo[result])
					}
					switch result {
					case RVSB_ZONE_BLACK:
						lostCoin += int32(playerEx.betInfo[RVSB_ZONE_RED])
					case RVSB_ZONE_RED:
						lostCoin += int32(playerEx.betInfo[RVSB_ZONE_BLACK])
					}
				}
				if luckyKind > CardsKind_Double {
					if playerEx.betInfo[RVSB_ZONE_LUCKY] > 0 {
						playerEx.winCoin[RVSB_ZONE_LUCKY] = (RVSB_LuckyKindOdds[luckyKind]*100 + 100) * playerEx.betInfo[RVSB_ZONE_LUCKY] / 100
						playerEx.gainCoin += playerEx.winCoin[RVSB_ZONE_LUCKY]
						sceneEx.isWin[2] = 1 //Luck赢
						//税收只算赢钱的
						tax += int(int64(RVSB_LuckyKindOdds[luckyKind]*playerEx.betInfo[RVSB_ZONE_LUCKY]) * int64(s.DbGameFree.GetTaxRate()) / 10000)
						winCoin += int32(RVSB_LuckyKindOdds[luckyKind] * playerEx.betInfo[RVSB_ZONE_LUCKY])
					}
				} else {
					lostCoin += int32(playerEx.betInfo[RVSB_ZONE_LUCKY])
				}

				playerEx.taxCoin = int64(tax)
				playerEx.winorloseCoin = int64(playerEx.gainCoin)
				playerEx.gainWinLost = playerEx.gainCoin - tax - playerEx.betTotal

				if playerEx.gainWinLost > 0 {
					playerEx.winRecord = append(playerEx.winRecord, 1)
				} else {
					playerEx.winRecord = append(playerEx.winRecord, 0)
				}
				sysGain += int64(playerEx.gainWinLost)
				if playerEx.Pos == RVSB_OLPOS {
					olTotalGain += int64(playerEx.gainWinLost)
				}
				playerEx.betBigRecord = append(playerEx.betBigRecord, playerEx.betTotal)
				playerEx.RecalcuLatestBet20()
				playerEx.RecalcuLatestWin20()
				if playerEx.gainWinLost > 0 {
					//oldflag := playerEx.MarkBroadcast(playerEx.Pos < RVSB_SELFPOS)
					playerEx.AddCoin(int64(playerEx.gainWinLost), common.GainWay_HundredSceneWin, base.SyncFlag_ToClient, "system", s.GetSceneName())
					//playerEx.MarkBroadcast(oldflag)

					//累积税收
					playerEx.AddServiceFee(int64(tax))
					chips := sceneEx.DbGameFree.GetOtherIntParams()
					if (playerEx.Coin) >= int64(chips[0]) && playerEx.Coin >= int64(sceneEx.DbGameFree.GetBetLimit()) {
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
				} else {
					//oldflag := playerEx.MarkBroadcast(playerEx.Pos < RVSB_SELFPOS)
					playerEx.AddCoin(int64(playerEx.gainWinLost), common.GainWay_HundredSceneLost, base.SyncFlag_ToClient, "system", s.GetSceneName())
					//playerEx.MarkBroadcast(oldflag)
				}
				if playerEx.betTotal != 0 {
					//上报游戏事件
					//playerEx.ReportGameEvent(int64(playerEx.gainWinLost), int64(tax), int64(playerEx.betTotal))
				}
				//统计游戏局数
				if playerEx.betTotal > 0 {
					playerEx.GameTimes++
					if playerEx.gainWinLost > 0 {
						playerEx.WinTimes++
					} else {
						playerEx.FailTimes++
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
				//统计金币变动
				if playerEx.GetCurrentCoin() == 0 {
					playerEx.SetCurrentCoin(playerEx.GetTakeCoin())
				}
				if playerEx.betTotal > 0 {
					playerEx.SaveSceneCoinLog(playerEx.GetCurrentCoin(), playerEx.Coin-playerEx.GetCurrentCoin(),
						playerEx.Coin, int64(playerEx.betTotal), playerEx.taxCoin, playerEx.winorloseCoin, 0, 0)
				}
				//playerEx.currentCoin = playerEx.Coin

				//统计投入产出
				if playerEx.gainWinLost != 0 {
					//赔率统计
					playerEx.Statics(sceneEx.KeyGameId, sceneEx.KeyGamefreeId, int64(playerEx.gainWinLost), true)
				}
			}
		}

		var constBills []*redvsblack.RedVsBlackBill
		if sceneEx.winTop1 != nil && sceneEx.winTop1.betTotal != 0 { //神算子
			constBills = append(constBills, &redvsblack.RedVsBlackBill{
				SnId:     proto.Int32(sceneEx.winTop1.SnId),
				Coin:     proto.Int64(sceneEx.winTop1.Coin),
				GainCoin: proto.Int64(int64(sceneEx.winTop1.gainWinLost)),
			})
		}
		if olTotalBet != 0 { //在线玩家
			constBills = append(constBills, &redvsblack.RedVsBlackBill{
				SnId:     proto.Int(0), //在线玩家约定snid=0
				Coin:     proto.Int64(olTotalGain),
				GainCoin: proto.Int64(olTotalGain),
			})
		} else {
			if sysGain < 0 {
				constBills = append(constBills, &redvsblack.RedVsBlackBill{
					SnId:     proto.Int(0), //在线玩家约定snid=0
					Coin:     proto.Int64(sysGain * -1),
					GainCoin: proto.Int64(sysGain * -1),
				})
			}
		}
		for i := 0; i < RVSB_RICHTOP5; i++ { //富豪前5名
			if sceneEx.betTop5[i] != nil && sceneEx.betTop5[i].betTotal != 0 {
				constBills = append(constBills, &redvsblack.RedVsBlackBill{
					SnId:     proto.Int32(sceneEx.betTop5[i].SnId),
					Coin:     proto.Int64(sceneEx.betTop5[i].Coin),
					GainCoin: proto.Int64(int64(sceneEx.betTop5[i].gainWinLost)),
				})
			}
		}
		for i := 0; i < count; i++ {
			playerEx := sceneEx.seats[i]
			if playerEx != nil && playerEx.IsOnLine() && !playerEx.IsMarkFlag(base.PlayerState_Leave) {
				pack := &redvsblack.SCRedVsBlackBilled{
					LoopNum: proto.Int(sceneEx.LoopNum),
				}
				pack.BillData = append(pack.BillData, constBills...)
				if sceneEx.seats[i].betTotal != 0 {
					pack.BillData = append(pack.BillData, &redvsblack.RedVsBlackBill{
						SnId:     proto.Int32(sceneEx.seats[i].SnId),
						Coin:     proto.Int64(sceneEx.seats[i].Coin),
						GainCoin: proto.Int64(int64(sceneEx.seats[i].gainWinLost)),
					})
				}
				if bigWinner != nil {
					pack.BigWinner = &redvsblack.RedVsBlackBigWinner{
						SnId:        proto.Int32(bigWinner.SnId),
						Name:        proto.String(bigWinner.Name),
						Head:        proto.Int32(bigWinner.Head),
						HeadOutLine: proto.Int32(bigWinner.HeadOutLine),
						VIP:         proto.Int32(bigWinner.VIP),
						Sex:         proto.Int32(bigWinner.Sex),
						Coin:        proto.Int64(bigWinner.Coin),
						GainCoin:    proto.Int(bigWinner.gainWinLost),
						City:        proto.String(bigWinner.GetCity()),
					}
				}
				proto.SetDefaults(pack)
				playerEx.SendToClient(int(redvsblack.RedVsBlackPacketID_PACKET_SC_RVSB_GAMEBILLED), pack)
			}
		}

		//统计参与游戏次数
		if !sceneEx.Testing {
			//var playerCtxs []*server.PlayerCtx
			//for _, p := range sceneEx.seats {
			//	if p == nil {
			//		continue
			//	}
			//	logger.Logger.Trace("snid:", p.SnId, p.IsRob, p.betTotal)
			//	if p.IsRob || p.betTotal == 0 {
			//		continue
			//	}
			//	playerCtxs = append(playerCtxs, &server.PlayerCtx{SnId: proto.Int32(p.SnId), Coin: proto.Int64(p.Coin)})
			//}
			//if len(playerCtxs) > 0 {
			//	pack := &server.GWSceneEnd{
			//		GameFreeId: proto.Int32(sceneEx.DbGameFree.GetId()),
			//		Players:    playerCtxs,
			//	}
			//	proto.SetDefaults(pack)
			//	sceneEx.SendToWorld(int(server.MmoPacketID_PACKET_GW_SCENEEND), pack)
			//}
		}

		//result, luckyKind := sceneEx.CalcuResult()

		s.SystemCoinOut = -int64(sceneEx.GetAllPlayerWinScore(result, luckyKind, true))

		////////////////////////水池更新提前////////////////
		//更新金币池
		result, luckyKind = sceneEx.CalcuResult()

		sysOut := sceneEx.GetAllPlayerWinScore(result, luckyKind, true)
		//if sysOut > 0 { //机器人输钱，用户赢钱
		base.CoinPoolMgr.PopCoin(sceneEx.GetGameFreeId(), sceneEx.GroupId, sceneEx.Platform, int64(sysOut))
		//}
		////////////////////////////////////////

		//统计下注数
		if !sceneEx.Testing {
			gwPlayerBet := &server.GWPlayerBet{
				GameFreeId: proto.Int32(sceneEx.DbGameFree.GetId()),
				RobotGain:  proto.Int64(s.SystemCoinOut),
			}
			for _, p := range sceneEx.seats {
				if p == nil || p.IsRob || p.betTotal == 0 {
					continue
				}
				playerBet := &server.PlayerBet{
					SnId: proto.Int32(p.SnId),
					Bet:  proto.Int64(int64(p.betTotal)),
					Gain: proto.Int64(int64(p.gainWinLost) + p.taxCoin),
					Tax:  proto.Int64(p.taxCoin),
				}
				gwPlayerBet.PlayerBets = append(gwPlayerBet.PlayerBets, playerBet)
			}
			if len(gwPlayerBet.PlayerBets) > 0 {
				proto.SetDefaults(gwPlayerBet)
				sceneEx.SendToWorld(int(server.SSPacketID_PACKET_GW_PLAYERBET), gwPlayerBet)
			}
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
		/////////////////////////////////////统计牌局详细记录
		hundredType := make([]model.HundredType, RVSB_ZONE_MAX, RVSB_ZONE_MAX)
		var isSave bool //有真人参与保存牌局
		for i := 0; i < RVSB_ZONE_MAX; i++ {
			hundredType[i] = model.HundredType{
				RegionId:  int32(i),
				IsWin:     sceneEx.isWin[i],
				CardsInfo: []int32{-1, -1, -1},
			}
			if i == 0 || i == 1 {
				hundredType[i].CardsKind = int32(sceneEx.kindOfcards[i].Kind)
				hundredType[i].CardsInfo = common.ThreeTonullArray(sceneEx.cards[i])
			} else {
				hundredType[i].CardsKind = int32(sceneEx.luckyKind)
			}
			playNum, index := 0, 0
			for _, o_player := range sceneEx.players {
				if o_player != nil && o_player.betInfo[i] > 0 && !o_player.IsRob {
					playNum++
					isSave = true
				}
			}
			if playNum == 0 {
				hundredType[i].PlayerData = nil
				continue
			}
			hundredPerson := make([]model.HundredPerson, playNum, playNum)
			for _, o_player := range sceneEx.players {
				if o_player != nil && !o_player.IsRob && o_player.betInfo[i] > 0 {
					if o_player.winCoin[i] <= 0 {
						o_player.winCoin[i] = -o_player.betInfo[i]
					} else if o_player.winCoin[i] > 0 {
						o_player.winCoin[i] -= o_player.betInfo[i]
					}
					pe := model.HundredPerson{
						UserId:       o_player.SnId,
						UserBetTotal: int64(o_player.betInfo[i]),
						ChangeCoin:   int64(o_player.winCoin[i]),
						BeforeCoin:   o_player.GetCurrentCoin(),
						AfterCoin:    o_player.Coin,
						IsRob:        o_player.IsRob,
						IsFirst:      sceneEx.IsPlayerFirst(o_player.Player),
						WBLevel:      o_player.WBLevel,
						Result:       o_player.result,
					}
					betDetail, ok := o_player.betDetailOrderInfo[i]
					if ok {
						pe.UserBetTotalDetail = betDetail
					}
					hundredPerson[index] = pe
					index++
				}
			}
			hundredType[i].PlayerData = hundredPerson
		}
		if isSave {
			info, err := model.MarshalGameNoteByHUNDRED(&hundredType)
			if err == nil {
				logid, _ := model.AutoIncGameLogId()
				for _, o_player := range sceneEx.players {
					if o_player != nil {
						if !o_player.IsRob {
							if o_player.betTotal > 0 {
								totalin, totalout := int64(0), int64(0)
								wincoin := 0

								for _, v := range o_player.winCoin {
									wincoin += v
								}

								if wincoin > 0 {
									totalout = int64(wincoin)

								} else {
									totalin = int64(-wincoin)
								}

								sceneEx.SaveGamePlayerListLog(o_player.SnId,
									&base.SaveGamePlayerListLogParam{
										Platform:          o_player.Platform,
										Channel:           o_player.Channel,
										Promoter:          o_player.BeUnderAgentCode,
										PackageTag:        o_player.PackageID,
										InviterId:         o_player.InviterId,
										LogId:             logid,
										TotalIn:           totalin,
										TotalOut:          totalout,
										TaxCoin:           o_player.taxCoin,
										ClubPumpCoin:      0,
										BetAmount:         int64(o_player.betTotal),
										WinAmountNoAnyTax: int64(o_player.gainWinLost + o_player.betTotal),
										IsFirstGame:       sceneEx.IsPlayerFirst(o_player.Player),
									})
							}
						}
						o_player.SetCurrentCoin(o_player.Coin)
						//清空玩家缓存数据
						o_player.Clean()
					}
				}
				//trend20Lately, _ := json.Marshal(sceneEx.trend20Lately)
				trends := []string{}
				if len(sceneEx.trend20Lately) > 0 {
					for _, v := range sceneEx.trend20Lately {
						switch v {
						case 1:
							trends = append(trends, "红")
						case 0:
							trends = append(trends, "黑")
						}
					}
				}
				trend20Lately, _ := json.Marshal(trends)
				sceneEx.SaveGameDetailedLog(logid, info, &base.GameDetailedParam{
					Trend20Lately: string(trend20Lately),
				})
				hundredType = nil
			}
		}
		sceneEx.DealyTime = int64(common.RandFromRange(0, 2000)) * int64(time.Millisecond)
	}
}

func (this *SceneBilledStateRedVsBlack) OnLeave(s *base.Scene) {
	if sceneEx, ok := s.ExtraData.(*RedVsBlackSceneData); ok {
		//清理掉线或者暂离的玩家
		for _, p := range sceneEx.players {
			if !p.IsOnLine() {
				s.PlayerLeave(p.Player, common.PlayerLeaveReason_DropLine, true)
			} else if p.IsRob {
				if s.CoinOverMaxLimit(p.Coin, p.Player) {
					s.PlayerLeave(p.Player, common.PlayerLeaveReason_Normal, true)
				} else if p.Trusteeship >= model.GameParamData.PlayerWatchNum {
					s.PlayerLeave(p.Player, common.PlayerLeaveReason_LongTimeNoOp, true)
				}
			} else {
				if !s.CoinInLimit(p.Coin) {
					s.PlayerLeave(p.Player, common.PlayerLeaveReason_Bekickout, true)
				} else if p.Trusteeship >= model.GameParamData.PlayerWatchNum {
					s.PlayerLeave(p.Player, common.PlayerLeaveReason_LongTimeNoOp, true)
				}
				todayGamefreeIDSceneData, _ := p.GetDaliyGameData(int(sceneEx.DbGameFree.GetId()))
				if todayGamefreeIDSceneData != nil &&
					sceneEx.DbGameFree.GetPlayNumLimit() != 0 &&
					todayGamefreeIDSceneData.GameTimes >= int64(sceneEx.DbGameFree.GetPlayNumLimit()) {
					s.PlayerLeave(p.Player, common.PlayerLeaveReason_GameTimes, true)
				}
			}
		}
		seatKey := sceneEx.Resort()
		if seatKey != sceneEx.constSeatKey {
			RedVsBlackSendSeatInfo(s, sceneEx)
			sceneEx.constSeatKey = seatKey
		}

		if s.CheckNeedDestroy() {
			sceneEx.SceneDestroy(true)
		}
	}
}

func (this *SceneBilledStateRedVsBlack) OnTick(s *base.Scene) {
	this.SceneBaseStateRedVsBlack.OnTick(s)
	if sceneEx, ok := s.ExtraData.(*RedVsBlackSceneData); ok {
		if time.Now().Sub(sceneEx.StateStartTime) > RedVsBlackBilledTimeout {
			s.ChangeSceneState(RedVsBlackSceneStateStakeAnt)
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
func (this *ScenePolicyRedVsBlack) RegisteSceneState(state base.SceneState) {
	if state == nil {
		return
	}
	stateid := state.GetState()
	if stateid < 0 || stateid >= RedVsBlackSceneStateMax {
		return
	}
	this.states[stateid] = state
}

//
func (this *ScenePolicyRedVsBlack) GetSceneState(s *base.Scene, stateid int) base.SceneState {
	if stateid >= 0 && stateid < RedVsBlackSceneStateMax {
		return ScenePolicyRedVsBlackSington.states[stateid]
	}
	return nil
}

func init() {
	ScenePolicyRedVsBlackSington.RegisteSceneState(&SceneStakeAntStateRedVsBlack{})
	ScenePolicyRedVsBlackSington.RegisteSceneState(&SceneStakeStateRedVsBlack{})
	ScenePolicyRedVsBlackSington.RegisteSceneState(&SceneOpenCardAntStateRedVsBlack{})
	ScenePolicyRedVsBlackSington.RegisteSceneState(&SceneOpenCardStateRedVsBlack{})
	ScenePolicyRedVsBlackSington.RegisteSceneState(&SceneBilledStateRedVsBlack{})
	core.RegisteHook(core.HOOK_BEFORE_START, func() error {
		base.RegisteScenePolicy(common.GameId_RedVsBlack, 0, ScenePolicyRedVsBlackSington)
		base.RegisteScenePolicy(common.GameId_RedVsBlack, 1, ScenePolicyRedVsBlackSington)
		return nil
	})
}

////////////////////////////////////////////////////////////////////////////////
