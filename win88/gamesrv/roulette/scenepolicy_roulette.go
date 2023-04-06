package roulette

import (
	"games.yol.com/win88/common"
	"games.yol.com/win88/gamerule/roulette"
	. "games.yol.com/win88/gamerule/roulette"
	"games.yol.com/win88/gamesrv/base"
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	proto_roulette "games.yol.com/win88/protocol/roulette"
	proto_server "games.yol.com/win88/protocol/server"
	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/logger"
	"time"
)

var ScenePolicyRouletteSington = &ScenePolicyRoulette{}

type ScenePolicyRoulette struct {
	base.BaseScenePolicy
	states [RouletteSceneStateMax]base.SceneState
}

//创建场景扩展数据
func (this *ScenePolicyRoulette) CreateSceneExData(s *base.Scene) interface{} {
	sceneEx := NewRouletteSceneData(s)
	if sceneEx != nil {
		if sceneEx.init() {
			s.ExtraData = sceneEx
		}
	}
	return sceneEx
}

//创建玩家扩展数据
func (this *ScenePolicyRoulette) CreatePlayerExData(s *base.Scene, p *base.Player) interface{} {
	playerEx := &RoulettePlayerData{Player: p}
	p.ExtraData = playerEx
	return playerEx
}

//场景开启事件
func (this *ScenePolicyRoulette) OnStart(s *base.Scene) {
	logger.Logger.Trace("(this *ScenePolicyRoulette) OnStart, SceneId=", s.SceneId)
	sceneEx := NewRouletteSceneData(s)
	if sceneEx != nil {
		if sceneEx.init() {
			s.ExtraData = sceneEx
			s.ChangeSceneState(RouletteSceneStateWait)
		}
	}
}

//场景关闭事件
func (this *ScenePolicyRoulette) OnStop(s *base.Scene) {
	logger.Logger.Trace("(this *ScenePolicyRoulette) OnStop , SceneId=", s.SceneId)
	if _, ok := s.ExtraData.(*RouletteSceneData); ok {
	}
}

//场景心跳事件
func (this *ScenePolicyRoulette) OnTick(s *base.Scene) {
	if s == nil {
		return
	}
	if s.SceneState != nil {
		s.SceneState.OnTick(s)
	}
}

func (this *ScenePolicyRoulette) GetPlayerNum(s *base.Scene) int32 {
	if s == nil {
		return 0
	}
	if sceneEx, ok := s.ExtraData.(*RouletteSceneData); ok {
		return int32(len(sceneEx.players))
	}
	return 0
}
func RouletteBroadcastRoomState(s *base.Scene, params ...int32) {
	pack := &proto_roulette.SCRouletteRoomState{
		State:  proto.Int(s.SceneState.GetState()),
		Params: params,
	}
	s.Broadcast(int(proto_roulette.RouletteMmoPacketID_PACKET_SC_Roulette_RoomState), pack, 0)
}

//玩家进入事件
func (this *ScenePolicyRoulette) OnPlayerEnter(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}

	logger.Logger.Trace("(this *ScenePolicyRoulette) OnPlayerEnter, SceneId=", s.SceneId, " player=", p.Name)
	if sceneEx, ok := s.ExtraData.(*RouletteSceneData); ok {
		playerEx := &RoulettePlayerData{Player: p}
		playerEx.Clean()
		playerEx.Pos = Roulette_OLPOS
		sceneEx.players[p.SnId] = playerEx
		sceneEx.seats = append(sceneEx.seats, playerEx)
		p.ExtraData = playerEx
		//给自己发送房间信息
		RouletteSendRoomInfo(s, p, sceneEx, playerEx)
		s.FirePlayerEvent(p, base.PlayerEventEnter, nil)
	}
}

//玩家离开事件
func (this *ScenePolicyRoulette) OnPlayerLeave(s *base.Scene, p *base.Player, reason int) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyRoulette) OnPlayerLeave, SceneId=", s.SceneId, " player=", p.Name)
	if sceneEx, ok := s.ExtraData.(*RouletteSceneData); ok {
		if p.IsGameing() {
			//在游戏中不让离开
			return
		}
		//维护座位列表,删除离开的玩家
		if len(sceneEx.seats) > 0 {
			for k, v := range sceneEx.seats {
				if v != nil && v.SnId == p.SnId {
					sceneEx.seats = append(sceneEx.seats[:k], sceneEx.seats[k+1:]...)
					break
				}
			}
		}
		sceneEx.delPlayer(p)
		s.FirePlayerEvent(p, base.PlayerEventLeave, nil)
	}
}

//玩家掉线
func (this *ScenePolicyRoulette) OnPlayerDropLine(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyRoulette) OnPlayerDropLine, SceneId=", s.SceneId, " player=", p.Name)
	s.FirePlayerEvent(p, base.PlayerEventDropLine, nil)
	if sceneEx, ok := s.ExtraData.(*RouletteSceneData); ok {
		p.DropTime = time.Now()
		if sceneEx.Gaming {
			return
		}
	}
}

//玩家重连
func (this *ScenePolicyRoulette) OnPlayerRehold(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyRoulette) OnPlayerRehold, SceneId=", s.SceneId, " player=", p.Name)
	if sceneEx, ok := s.ExtraData.(*RouletteSceneData); ok {
		if playerEx, ok := p.ExtraData.(*RoulettePlayerData); ok {
			playerEx.MarkFlag(base.PlayerState_Auto)
			//发送房间信息给自己
			RouletteSendRoomInfo(s, p, sceneEx, playerEx)

			s.FirePlayerEvent(p, base.PlayerEventRehold, nil)
		}
	}
}

//玩家返回房间
func (this *ScenePolicyRoulette) OnPlayerReturn(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyRoulette) OnPlayerReturn, SceneId=", s.SceneId, " player=", p.Name)
	if sceneEx, ok := s.ExtraData.(*RouletteSceneData); ok {
		if playerEx, ok := p.ExtraData.(*RoulettePlayerData); ok {
			playerEx.MarkFlag(base.PlayerState_Auto)
			//发送房间信息给自己
			RouletteSendRoomInfo(s, p, sceneEx, playerEx)

			s.FirePlayerEvent(p, base.PlayerEventReturn, nil)
		}
	}
}

//玩家操作
func (this *ScenePolicyRoulette) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if s == nil || p == nil {
		return false
	}
	if s.SceneState != nil {
		return s.SceneState.OnPlayerOp(s, p, opcode, params)
	}
	return true
}

//玩家事件
func (this *ScenePolicyRoulette) OnPlayerEvent(s *base.Scene, p *base.Player, evtcode int, params []int64) {
	if s == nil || p == nil {
		return
	}
	if s.SceneState != nil {
		s.SceneState.OnPlayerEvent(s, p, evtcode, params)
	}
}

//观众进入事件
func (this *ScenePolicyRoulette) OnAudienceEnter(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyRoulette) OnAudienceEnter, SceneId=", s.SceneId, " player=", p.Name)
}

//观众掉线
func (this *ScenePolicyRoulette) OnAudienceDropLine(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyRoulette) OnAudienceDropLine, SceneId=", s.SceneId, " player=", p.Name)
	s.AudienceLeave(p, common.PlayerLeaveReason_DropLine)
}

//观众坐下
func (this *ScenePolicyRoulette) OnAudienceSit(s *base.Scene, p *base.Player) {
	if sceneEx, ok := s.ExtraData.(*RouletteSceneData); ok {
		pos := p.Pos
		if pos != -1 && sceneEx.seats[pos] == nil {
			playerEx := &RoulettePlayerData{Player: p}
			sceneEx.seats[pos] = playerEx
			sceneEx.players[p.SnId] = playerEx
			p.ExtraData = playerEx
			p.MarkFlag(base.PlayerState_WaitNext)
		}

		s.FirePlayerEvent(p, base.PlayerEventEnter, nil)
	}
}

//是否完成了整个牌局
func (this *ScenePolicyRoulette) IsCompleted(s *base.Scene) bool {
	if s == nil {
		return false
	}
	return false
}

//是否可以强制开始
func (this *ScenePolicyRoulette) IsCanForceStart(s *base.Scene) bool {
	return true
}

//当前状态能否换桌
func (this *ScenePolicyRoulette) CanChangeCoinScene(s *base.Scene, p *base.Player) bool {
	if s == nil || p == nil {
		return true
	}
	if s.SceneState != nil {
		return s.SceneState.CanChangeCoinScene(s, p)
	}
	return true
}
func (this *ScenePolicyRoulette) NotifyGameState(s *base.Scene) {
	if s.GetState() > 2 {
		return
	}
	sec := int(RouletteBetTimeout.Seconds())
	s.SyncGameState(sec, 0)
}
func RouletteCreateRoomInfoPacket(s *base.Scene, sceneEx *RouletteSceneData, playerEx *RoulettePlayerData) interface{} {
	//房间信息
	pack := &proto_roulette.SCRouletteRoomInfo{
		RoomId:     proto.Int(s.SceneId),
		Creator:    proto.Int32(s.Creator),
		GameId:     proto.Int(s.GameId),
		GameFreeId: proto.Int32(s.GetGameFreeId()),
		RoomMode:   proto.Int(s.SceneMode),
		AgentId:    proto.Int32(s.GetAgentor()),
		SceneType:  proto.Int(s.SceneType),
		Params: []int32{sceneEx.DbGameFree.GetLimitCoin(),
			sceneEx.DbGameFree.GetMaxCoinLimit(),
			sceneEx.DbGameFree.GetBetLimit()},
		//sceneEx.DbGameFree.GetServiceFee(), sceneEx.DbGameFree.GetLowerThanKick(),
		NumOfGames: proto.Int(sceneEx.NumOfGames),
		State:      proto.Int(s.SceneState.GetState()),
		TimeOut:    proto.Int(s.SceneState.GetTimeout(s)),
		DisbandGen: proto.Int(sceneEx.GetDisbandGen()),
		ParamsEx:   s.GetParamsEx(),
		WinRecord:  sceneEx.winRecord,
		MaxBetCoin: sceneEx.DbGameFree.GetMaxBetCoin(),
	}

	for _, v := range sceneEx.richTop {
		if v != nil {
			pd := &proto_roulette.RoulettePlayerData{
				SnId:        proto.Int32(v.SnId),
				Name:        proto.String(v.Name),
				Head:        proto.Int32(v.Head),
				Sex:         proto.Int32(v.Sex),
				Coin:        proto.Int64(v.Coin - v.totalBetCoin),
				Pos:         proto.Int(v.Pos),
				Flag:        proto.Int(v.GetFlag()),
				City:        proto.String(v.GetCity()),
				HeadOutLine: proto.Int32(v.HeadOutLine),
				VIP:         proto.Int32(v.VIP),
				NiceId:      proto.Int32(v.NiceId),
			}
			if s.SceneState.GetState() == RouletteSceneStateBilled {
				pd.Coin = proto.Int64(v.Coin)
			}
			pack.Players = append(pack.Players, pd)
		}
	}
	if playerEx != nil { //自己的数据
		pd := &proto_roulette.RoulettePlayerData{
			SnId:        proto.Int32(playerEx.SnId),
			Name:        proto.String(playerEx.Name),
			Head:        proto.Int32(playerEx.Head),
			Sex:         proto.Int32(playerEx.Sex),
			Coin:        proto.Int64(playerEx.Coin - playerEx.totalBetCoin),
			Pos:         proto.Int(Roulette_SELFPOS),
			Flag:        proto.Int(playerEx.GetFlag()),
			City:        proto.String(playerEx.GetCity()),
			HeadOutLine: proto.Int32(playerEx.HeadOutLine),
			VIP:         proto.Int32(playerEx.VIP),
			BetCoin:     proto.Int64(playerEx.totalBetCoin),
			NiceId:      proto.Int32(playerEx.NiceId),
		}
		if s.SceneState.GetState() == RouletteSceneStateBilled {
			pd.Coin = proto.Int64(playerEx.Coin)
		}
		pd.LastBetInfo = make([]*proto_roulette.RouletteBetInfo, 0)

		pd.BetInfo = make([]*proto_roulette.RouletteBetInfo, 0)

		//筹码数量
		for k, cntIdx := range playerEx.betCnt {
			pd.BetInfo = append(pd.BetInfo, &proto_roulette.RouletteBetInfo{
				BetPos: proto.Int32(int32(k)),
				Cnt:    common.IntSliceToInt32(cntIdx[:]),
			})
		}

		if playerEx.proceedBet {
			for k, cntIdx := range playerEx.lastBetRecord {
				pd.BetInfo = append(pd.BetInfo, &proto_roulette.RouletteBetInfo{
					BetPos: proto.Int32(int32(k)),
					Cnt:    common.IntSliceToInt32(cntIdx[:]),
				})
			}
		}

		//上一局下注筹码数量
		for k, cntIdx := range playerEx.lastBetRecord {
			pd.LastBetInfo = append(pd.LastBetInfo, &proto_roulette.RouletteBetInfo{
				BetPos: proto.Int32(int32(k)),
				Cnt:    common.IntSliceToInt32(cntIdx[:]),
			})
		}
		pack.Players = append(pack.Players, pd)
	}
	proto.SetDefaults(pack)
	return pack
}

func RouletteSendRoomInfo(s *base.Scene, p *base.Player, sceneEx *RouletteSceneData, playerEx *RoulettePlayerData) {
	pack := RouletteCreateRoomInfoPacket(s, sceneEx, playerEx)
	logger.Logger.Trace("-------------------发送房间消息: ", pack)
	p.SendToClient(int(proto_roulette.RouletteMmoPacketID_PACKET_SC_Roulette_RoomInfo), pack)
}

//状态基类
type SceneBaseStateRoulette struct {
}

func (this *SceneBaseStateRoulette) GetTimeout(s *base.Scene) int {
	if sceneEx, ok := s.ExtraData.(*RouletteSceneData); ok {
		return int(time.Now().Sub(sceneEx.StateStartTime) / time.Second)
	}
	return 0
}

func (this *SceneBaseStateRoulette) CanChangeTo(s base.SceneState) bool {
	return true
}

//当前状态能否换桌
func (this *SceneBaseStateRoulette) CanChangeCoinScene(s *base.Scene, p *base.Player) bool {
	return !p.IsGameing()
}
func (this *SceneBaseStateRoulette) OnEnter(s *base.Scene) {
	if sceneEx, ok := s.ExtraData.(*RouletteSceneData); ok {
		sceneEx.StateStartTime = time.Now()
	}
}
func (this *SceneBaseStateRoulette) OnLeave(s *base.Scene) {}
func (this *SceneBaseStateRoulette) OnTick(s *base.Scene)  {}
func (this *SceneBaseStateRoulette) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if sceneEx, ok := s.ExtraData.(*RouletteSceneData); ok {
		if opcode == RoulettePlayerOpPlayerList {
			p.Trusteeship = 0
			sceneEx.SendPlayerList(p)
			return true
		}
	}
	return false
}
func (this *SceneBaseStateRoulette) OnPlayerEvent(s *base.Scene, p *base.Player, evtcode int, params []int64) {
	if _, ok := s.ExtraData.(*RouletteSceneData); ok {
		switch evtcode {
		case base.PlayerEventEnter:
		case base.PlayerEventLeave:
		case base.PlayerEventRecharge:
		}
	}

}

//////////////////////////////////////////////////////////////
//等待状态
//////////////////////////////////////////////////////////////
type SceneWaitStateRoulette struct {
	SceneBaseStateRoulette
}

func (this *SceneWaitStateRoulette) GetState() int {
	return RouletteSceneStateWait
}

func (this *SceneWaitStateRoulette) CanChangeTo(s base.SceneState) bool {
	switch s.GetState() {
	case RouletteSceneStateBet:
		return true
	}
	return false
}

//当前状态能否换桌
func (this *SceneWaitStateRoulette) CanChangeCoinScene(s *base.Scene, p *base.Player) bool {
	return true
}

func (this *SceneWaitStateRoulette) GetTimeout(s *base.Scene) int {
	return 0
}

func (this *SceneWaitStateRoulette) OnEnter(s *base.Scene) {
	this.SceneBaseStateRoulette.OnEnter(s)
	RouletteBroadcastRoomState(s)
	if sceneEx, ok := s.ExtraData.(*RouletteSceneData); ok {
		s.Gaming = false
		sceneEx.Clean()
		if sceneEx.CanStart() {
			s.ChangeSceneState(RouletteSceneStateBet)
		}
	}
}

//状态离开时
func (this *SceneWaitStateRoulette) OnLeave(s *base.Scene) {
	this.SceneBaseStateRoulette.OnLeave(s)
	logger.Logger.Tracef("(this *SceneStateWaitRoulette) OnLeave, sceneid=%v", s.SceneId)
	if sceneEx, ok := s.ExtraData.(*RouletteSceneData); ok {
		sceneEx.GameStartTime = time.Now()
	}
}

//玩家操作
func (this *SceneWaitStateRoulette) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if this.SceneBaseStateRoulette.OnPlayerOp(s, p, opcode, params) {
		return true
	}
	if sceneEx, ok := s.ExtraData.(*RouletteSceneData); ok {
		if playerEx, ok := p.ExtraData.(*RoulettePlayerData); ok {
			switch opcode {
			case RoulettePlayerOpReady: //准备
				if !playerEx.IsReady() {
					playerEx.MarkFlag(base.PlayerState_Ready)
					playerEx.SyncFlag()
					if sceneEx.CanStart() {
						s.ChangeSceneState(RouletteSceneStateBet)
					}
				}
			case RoulettePlayerOpCancelReady: //取消准备
				if playerEx.IsReady() {
					playerEx.UnmarkFlag(base.PlayerState_Ready)
					playerEx.SyncFlag()
				}
			case RoulettePlayerOpKickout: //踢人
			}
		}
	}
	return true
}

//玩家事件
func (this *SceneWaitStateRoulette) OnPlayerEvent(s *base.Scene, p *base.Player, evtcode int, params []int64) {
	logger.Logger.Trace("(this *SceneStateWaitRoulette) OnPlayerEvent, SceneId=", s.SceneId, " player=", p.SnId, " evtcode=", evtcode)
	this.SceneBaseStateRoulette.OnPlayerEvent(s, p, evtcode, params)
}

func (this *SceneWaitStateRoulette) OnTick(s *base.Scene) {
	this.SceneBaseStateRoulette.OnTick(s)
	if sceneEx, ok := s.ExtraData.(*RouletteSceneData); ok {
		if s.CheckNeedDestroy() {
			sceneEx.SceneDestroy(true)
			return
		}
		if time.Now().Sub(sceneEx.StateStartTime) > RouletteSceneWaitTimeout {
			if sceneEx.CanStart() {
				s.ChangeSceneState(RouletteSceneStateBet)
			}
		}
	}
}

//////////////////////////////////////////////////////////////
//下注阶段
//////////////////////////////////////////////////////////////
type SceneBetStateRoulette struct {
	SceneBaseStateRoulette
}

//获取当前场景状态
func (this *SceneBetStateRoulette) GetState() int {
	return RouletteSceneStateBet
}

//是否可以改变到其它场景状态
func (this *SceneBetStateRoulette) CanChangeTo(s base.SceneState) bool {
	switch s.GetState() {
	case RouletteSceneStateOpenPrize:
		return true
	}
	return false
}

//场景状态进入
func (this *SceneBetStateRoulette) OnEnter(s *base.Scene) {
	this.SceneBaseStateRoulette.OnEnter(s)
	if sceneEx, ok := s.ExtraData.(*RouletteSceneData); ok {
		//sceneEx.Clean()
		sceneEx.NumOfGames++
		sceneEx.ShowRichTop()
		RouletteBroadcastRoomState(s)
		sceneEx.tempTime = time.Now()
		sceneEx.GameNowTime = time.Now()
		//for _, p := range sceneEx.players {
		//	if p != nil {
		//		p.Trusteeship++
		//	}
		//}
	}
}

//玩家操作
func (this *SceneBetStateRoulette) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	logger.Logger.Tracef("(this *SceneBetStateRoulette) OnPlayerOp opcode %v ,params %v", opcode, params)
	if this.SceneBaseStateRoulette.OnPlayerOp(s, p, opcode, params) {
		return true
	}
	if sceneEx, ok := s.ExtraData.(*RouletteSceneData); ok {
		if playerEx, ok := p.ExtraData.(*RoulettePlayerData); ok {
			p.Trusteeship = 0
			pack := &proto_roulette.SCRoulettePlayerOp{
				OpRCode: proto.Int32(RoulettePlayerOpError),
				OpCode:  proto.Int32(int32(opcode)),
				Pos:     proto.Int32(int32(p.Pos)),
			}
			switch opcode {
			case RoulettePlayerOpBet: //下注
				if !p.IsGameing() {
					p.UnmarkFlag(base.PlayerState_WaitNext)
				}
				isF, betCode := playerEx.Bet(params, sceneEx)
				pack.OpRCode = proto.Int32(betCode)
				if isF {
					br := playerEx.betRecord[len(playerEx.betRecord)-1]
					rbi := &proto_roulette.RouletteBetInfo{
						BetPos: proto.Int32(int32(br.BetPos)),
						Cnt:    common.IntSliceToInt32(br.CntIdx[:]),
					}
					pack.ProceedBetCoin = append(pack.ProceedBetCoin, rbi)
					//////////////////////////////
					if playerEx.Pos == Roulette_OLPOS {
						for _, tp := range sceneEx.players {
							if tp != nil && !tp.IsRob && tp.SnId != playerEx.SnId {
								if _, ok := tp.tempBetCnt[br.BetPos]; !ok {
									tp.tempBetCnt[br.BetPos] = roulette.InitBet()
								}
								for k, cnt := range br.CntIdx {
									tp.tempBetCnt[br.BetPos][k] += cnt
								}
							}
						}
					}
				}
			case RoulettePlayerOpRecall: //撤销
				pack.OpRCode = proto.Int32(RoulettePlayerOpSuccess)
				isType, betPos, betInt := playerEx.RecallLast(sceneEx)
				if isType == 1 {
					rbi := &proto_roulette.RouletteBetInfo{
						BetPos: proto.Int32(int32(betPos)),
						Cnt:    common.IntSliceToInt32(betInt[:]),
					}
					pack.ProceedBetCoin = append(pack.ProceedBetCoin, rbi)
				} else if isType == 2 {
					playerEx.proceedBet = false
					playerEx.totalBetCoin -= playerEx.lastBetCoin
					pack.ProceedBetCoin = make([]*proto_roulette.RouletteBetInfo, 0)
					for k, cnt := range playerEx.lastBetRecord {
						pack.ProceedBetCoin = append(pack.ProceedBetCoin, &proto_roulette.RouletteBetInfo{
							BetPos: proto.Int32(int32(k)),
							Cnt:    common.IntSliceToInt32(cnt[:]),
						})
					}
					p.MarkFlag(base.PlayerState_WaitNext)
				} else {
					pack.OpRCode = proto.Int32(RoulettePlayerOpNoBetCoinNotRecall)
				}
			case RoulettePlayerOpProceedBet: //续投
				pack.OpRCode = proto.Int32(RoulettePlayerOpSuccess)
				if !playerEx.proceedBet && playerEx.totalBetCoin == 0 {
					if playerEx.Coin-playerEx.lastBetCoin >= int64(sceneEx.DbGameFree.GetBetLimit()) {
						p.UnmarkFlag(base.PlayerState_WaitNext)
						playerEx.proceedBet = true
						playerEx.totalBetCoin += playerEx.lastBetCoin
						pack.ProceedBetCoin = make([]*proto_roulette.RouletteBetInfo, 0)
						for k, cnt := range playerEx.lastBetRecord {
							pack.ProceedBetCoin = append(pack.ProceedBetCoin, &proto_roulette.RouletteBetInfo{
								BetPos: proto.Int32(int32(k)),
								Cnt:    common.IntSliceToInt32(cnt[:]),
							})
						}
						if playerEx.Pos == Roulette_OLPOS {
							//只为客户端表现处理
							for _, v := range sceneEx.players {
								if v != nil && !v.IsRob && v.SnId != playerEx.SnId {
									for keyNum, br := range playerEx.lastBetRecord {
										for chipIdx, cnt := range br {
											if _, ok := v.tempBetCnt[keyNum]; !ok {
												v.tempBetCnt[keyNum] = roulette.InitBet()
											}
											tbc := v.tempBetCnt[keyNum]
											tbc[chipIdx] += cnt
										}
									}
								}
							}
						}
					}
				} else {
					pack.OpRCode = proto.Int32(RoulettePlayerOpAlreadyProceedBet)
				}
			}
			sceneEx.SendToClickBetInfo(pack, playerEx)
		}
	}
	return true
}

//场景状态离开
func (this *SceneBetStateRoulette) OnLeave(s *base.Scene) {
	if sceneEx, ok := s.ExtraData.(*RouletteSceneData); ok {
		//不管玩家下没下注 玩家都累计一局
		for _, p := range sceneEx.players {
			if p != nil {
				p.gameRouletteTimes++
			}
		}
		for _, v := range sceneEx.players {
			if v != nil && !v.IsRob {
				if v.totalBetCoin <= 0 {
					v.Trusteeship++
					if v.Trusteeship >= model.GameParamData.NotifyPlayerWatchNum {
						v.SendTrusteeshipTips()
					}
				}
			}
		}
	}
}

func (this *SceneBetStateRoulette) OnTick(s *base.Scene) {
	this.SceneBaseStateRoulette.OnTick(s)
	if sceneEx, ok := s.ExtraData.(*RouletteSceneData); ok {
		if time.Now().Sub(sceneEx.StateStartTime) > RouletteBetTimeout {
			s.ChangeSceneState(RouletteSceneStateOpenPrize)
		} else if time.Now().Sub(sceneEx.tempTime) > time.Second*2 {
			sceneEx.tempTime = time.Now()
			for _, p := range sceneEx.players {
				if p != nil && !p.IsRob && p.IsOnLine() {
					var rbc *proto_roulette.SCRouletteBetChange = nil
					if len(p.tempBetCnt) > 0 {
						rbc = new(proto_roulette.SCRouletteBetChange)
						for keyNum, br := range p.tempBetCnt {
							rbc.BetInfo = append(rbc.BetInfo, &proto_roulette.RouletteBetInfo{
								BetPos: proto.Int32(int32(keyNum)),
								Cnt:    common.IntSliceToInt32(br[:]),
							})
						}
					}
					if rbc != nil {
						proto.SetDefaults(rbc)
						logger.Logger.Trace("SCRouletteBetChange: ", rbc)
						p.SendToClient(int(proto_roulette.RouletteMmoPacketID_PACKET_SC_Roulette_BetChange), rbc)
						p.tempBetCnt = make(map[int][]int)
					}
					//s.Broadcast(int(proto_roulette.RouletteMmoPacketID_PACKET_SC_Roulette_BetChange), rbc, p.sid)
				}
			}
		}
	}
}

//////////////////////////////////////////////////////////////
//开奖阶段
//////////////////////////////////////////////////////////////
type SceneOpenPrizeStateRoulette struct {
	SceneBaseStateRoulette
}

//获取当前场景状态
func (this *SceneOpenPrizeStateRoulette) GetState() int {
	return RouletteSceneStateOpenPrize
}

//是否可以改变到其它场景状态
func (this *SceneOpenPrizeStateRoulette) CanChangeTo(s base.SceneState) bool {
	switch s.GetState() {
	case RouletteSceneStateBilled:
		return true
	}
	return false
}

//玩家操作
func (this *SceneOpenPrizeStateRoulette) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if this.SceneBaseStateRoulette.OnPlayerOp(s, p, opcode, params) {
		return true
	}
	return true
}

//场景状态进入
func (this *SceneOpenPrizeStateRoulette) OnEnter(s *base.Scene) {
	if sceneEx, ok := s.ExtraData.(*RouletteSceneData); ok {
		s.Gaming = true
		otherInts := sceneEx.DbGameFree.GetOtherIntParams()
		if len(otherInts) == 0 {
			return
		}
		for _, p := range sceneEx.players {
			if p != nil && p.IsGameing() {
				if p.proceedBet {
					for keyNum, br := range p.lastBetRecord {
						//下注记录
						if br != nil {
							p.betRecord = append(p.betRecord, &BetRecord{BetPos: keyNum, CntIdx: br})
						}
						//记录玩家每个区域下注筹码数量
						if _, ok := p.betCnt[keyNum]; !ok {
							p.betCnt[keyNum] = roulette.InitBet()
						}
						pbc := p.betCnt[keyNum]
						if p.Pos == Roulette_OLPOS {
							if _, ok := sceneEx.betCnt[keyNum]; !ok {
								sceneEx.betCnt[keyNum] = roulette.InitBet()
							}
						}
						sbc := sceneEx.betCnt[keyNum]
						for chipIdx, cnt := range br {
							if cnt > 0 {
								chipCoin := int64(otherInts[chipIdx])
								//记录玩家每个区域下注筹码数量
								pbc[chipIdx] += cnt
								//记录玩家当局每个区域下注额
								p.betCoin[keyNum] += chipCoin * int64(cnt)
								if p.Pos == Roulette_OLPOS {
									sbc[chipIdx] += cnt
								}
								if !p.IsRob {
									sceneEx.realBetCoin[keyNum] += chipCoin * int64(cnt)
								}
							}
						}
					}
				}
			}
		}

		//水池上下文环境
		sceneEx.CpCtx = base.CoinPoolMgr.GetCoinPoolCtx(sceneEx.Platform, sceneEx.GetGameFreeId(), sceneEx.GroupId)

		sceneEx.ComputePoolState()

		this.SceneBaseStateRoulette.OnEnter(s)
		RouletteBroadcastRoomState(s, int32(sceneEx.winDigit))
		sceneEx.CpCtx.Controlled = sceneEx.CpControlled

		sceneEx.BilledScore()
	}
}

//场景状态离开
func (this *SceneOpenPrizeStateRoulette) OnLeave(s *base.Scene) {
}

func (this *SceneOpenPrizeStateRoulette) OnTick(s *base.Scene) {
	this.SceneBaseStateRoulette.OnTick(s)
	if sceneEx, ok := s.ExtraData.(*RouletteSceneData); ok {
		if time.Now().Sub(sceneEx.StateStartTime) > RouletteOpenPrizeTimeout {
			s.ChangeSceneState(RouletteSceneStateBilled)
		}
	}
}

//////////////////////////////////////////////////////////////
//结算阶段
//////////////////////////////////////////////////////////////
type SceneBilledStateRoulette struct {
	SceneBaseStateRoulette
}

//获取当前场景状态
func (this *SceneBilledStateRoulette) GetState() int {
	return RouletteSceneStateBilled
}

//是否可以改变到其它场景状态
func (this *SceneBilledStateRoulette) CanChangeTo(s base.SceneState) bool {
	switch s.GetState() {
	case RouletteSceneStateStart:
		return true
	}
	return false
}

//玩家操作
func (this *SceneBilledStateRoulette) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if this.SceneBaseStateRoulette.OnPlayerOp(s, p, opcode, params) {
		return true
	}
	return true
}

//场景状态进入
func (this *SceneBilledStateRoulette) OnEnter(s *base.Scene) {
	this.SceneBaseStateRoulette.OnEnter(s)
	RouletteBroadcastRoomState(s)

	logid, _ := model.AutoIncGameLogId()
	bSaveDetailLog := false
	if sceneEx, ok := s.ExtraData.(*RouletteSceneData); ok {
		s.Gaming = false
		//结算阶段玩家金额变动
		for _, playerEx := range sceneEx.players {
			if playerEx != nil && playerEx.IsGameing() {
				way := common.GainWay_HundredSceneLost
				if playerEx.gainCoin > 0 {
					way = common.GainWay_HundredSceneWin
				}
				playerEx.AddCoin(playerEx.gainCoin, int32(way), base.SyncFlag_Broadcast, "system", sceneEx.GetSceneName())
				//游戏事件或者牌局事件
				//if playerEx.totalBetCoin > 0 {
				//playerEx.ReportGameEvent(playerEx.gainCoin, playerEx.taxCoin, playerEx.totalBetCoin)
				//}
				//玩家输赢计算
				if playerEx.gainCoin > 0 {
					playerEx.SetWinTimes(playerEx.GetWinTimes() + 1)
				} else {
					playerEx.SetLostTimes(playerEx.GetLostTimes() + 1)
				}

				//统计金币变动
				if !playerEx.IsRob {
					bSaveDetailLog = true
					playerEx.SaveSceneCoinLog(playerEx.Coin-playerEx.gainCoin, playerEx.gainCoin, playerEx.Coin, playerEx.totalBetCoin,
						playerEx.taxCoin, playerEx.gainCoin+playerEx.taxCoin, 0, 0)
					totalin, totalout := int64(0), int64(0)
					if playerEx.gainCoin+playerEx.taxCoin > 0 {
						totalout = playerEx.gainCoin + playerEx.taxCoin
					} else {
						totalin = -(playerEx.gainCoin + playerEx.taxCoin)
					}
					validFlow := totalin + totalout
					validBet := common.AbsI64(totalin - totalout)
					sceneEx.SaveGamePlayerListLog(playerEx.SnId,
						&base.SaveGamePlayerListLogParam{
							Platform:          playerEx.Platform,
							Channel:           playerEx.Channel,
							Promoter:          playerEx.BeUnderAgentCode,
							PackageTag:        playerEx.PackageID,
							InviterId:         playerEx.InviterId,
							LogId:             logid,
							TotalIn:           totalin,
							TotalOut:          totalout,
							TaxCoin:           playerEx.taxCoin,
							ClubPumpCoin:      0,
							BetAmount:         playerEx.totalBetCoin,
							WinAmountNoAnyTax: playerEx.gainCoin + playerEx.totalBetCoin,
							ValidBet:          validBet,
							ValidFlow:         validFlow,
							IsFirstGame:       sceneEx.IsPlayerFirst(playerEx.Player),
						})
				}

				//统计投入产出
				inout := playerEx.gainCoin + playerEx.taxCoin
				if inout != 0 {
					//赔率统计
					playerEx.Statics(sceneEx.KeyGameId, sceneEx.KeyGamefreeId, inout, true)
				}

			}
		}
		//统计下注数
		if !sceneEx.Testing {
			gwPlayerBet := &proto_server.GWPlayerBet{
				GameFreeId: proto.Int32(sceneEx.DbGameFree.GetId()),
			}
			for _, p := range sceneEx.players {
				if p == nil || p.IsRob || p.totalBetCoin == 0 {
					continue
				}
				playerBet := &proto_server.PlayerBet{
					SnId: proto.Int32(p.SnId),
					Bet:  proto.Int64(int64(p.totalBetCoin)),
				}
				gwPlayerBet.PlayerBets = append(gwPlayerBet.PlayerBets, playerBet)
			}
			if len(gwPlayerBet.PlayerBets) > 0 {
				proto.SetDefaults(gwPlayerBet)
				sceneEx.SendToWorld(int(proto_server.SSPacketID_PACKET_GW_PLAYERBET), gwPlayerBet)
			}
		}

		//给客户端发送结算结果
		for _, p := range sceneEx.players {
			if p.IsOnLine() {
				temp := make([]*proto_roulette.RouletteBilled, 0)

				for _, rp := range sceneEx.richTop {
					if rp != nil {
						if rp.Pos == p.Pos {
							continue
						}
						playerRb := &proto_roulette.RouletteBilled{
							Coin:     proto.Int64(rp.winCoin),
							GainCoin: proto.Int64(rp.gainCoin),
							LastCoin: proto.Int64(rp.Coin),
							Pos:      proto.Int(rp.Pos),
						}
						playerRb.BetInfo = make([]*proto_roulette.RouletteBetInfo, 0)
						for keyNum, br := range rp.betCnt {
							playerRb.BetInfo = append(playerRb.BetInfo, &proto_roulette.RouletteBetInfo{
								BetPos: proto.Int32(int32(keyNum)),
								Cnt:    common.IntSliceToInt32(br[:]),
							})
						}
						temp = append(temp, playerRb)
					}
				}
				pack := &proto_roulette.SCRouletteBilled{}
				if !p.IsRob {
					pack.Players = append(pack.Players, temp...) //上座玩家
				}
				tp := make(map[int][]int)
				for k, v := range sceneEx.betCnt {
					tp[k] = make([]int, len(v))
					copy(tp[k], v)
				}

				if p.IsGameing() {
					//发送结果
					playerRb := &proto_roulette.RouletteBilled{
						Coin:     proto.Int64(p.winCoin),
						GainCoin: proto.Int64(p.gainCoin),
						LastCoin: proto.Int64(p.Coin),
						Pos:      proto.Int(Roulette_SELFPOS),
					}
					playerRb.BetInfo = make([]*proto_roulette.RouletteBetInfo, 0)
					for keyNum, br := range p.betCnt {
						kn := tp[keyNum]
						if p.Pos == Roulette_OLPOS && kn != nil {
							for k, v := range br {
								kn[k] -= v
							}
						}
						playerRb.BetInfo = append(playerRb.BetInfo, &proto_roulette.RouletteBetInfo{
							BetPos: proto.Int32(int32(keyNum)),
							Cnt:    common.IntSliceToInt32(br[:]),
						})
					}
					pack.Players = append(pack.Players, playerRb) //自己
				}

				otherWin := sceneEx.winCoin
				if p.Pos == Roulette_OLPOS {
					otherWin = sceneEx.winCoin - p.winCoin
				}
				if otherWin > 0 {
					//其他所有玩家
					otherPlayer := &proto_roulette.RouletteBilled{
						Coin:     proto.Int64(otherWin),
						GainCoin: proto.Int64(0),
						LastCoin: proto.Int64(0),
						Pos:      proto.Int(Roulette_OLPOS),
					}
					otherPlayer.BetInfo = make([]*proto_roulette.RouletteBetInfo, 0)
					for keyNum, br := range tp {
						otherPlayer.BetInfo = append(otherPlayer.BetInfo, &proto_roulette.RouletteBetInfo{
							BetPos: proto.Int32(int32(keyNum)),
							Cnt:    common.IntSliceToInt32(br[:]),
						})
					}
					if !p.IsRob {
						pack.Players = append(pack.Players, otherPlayer) //其他玩家
					}
				}
				//proto.SetDefaults(pack)
				logger.Logger.Trace("SCRouletteBilled:", pack)
				p.SendToClient(int(proto_roulette.RouletteMmoPacketID_PACKET_SC_Roulette_Billed), pack)
			}
		}

		////////////////////////////////////////////////////////
		if bSaveDetailLog && !sceneEx.Testing {
			sceneEx.SaveLog(logid)
		}
	}
}

//场景状态离开
func (this *SceneBilledStateRoulette) OnLeave(s *base.Scene) {
	this.SceneBaseStateRoulette.OnLeave(s)
	if sceneEx, ok := s.ExtraData.(*RouletteSceneData); ok {
		//同步开奖结果
		pack := &proto_server.GWGameStateLog{
			SceneId: proto.Int(s.SceneId),
			GameLog: proto.Int(sceneEx.winDigit),
		}
		s.SendToWorld(int(proto_server.SSPacketID_PACKET_GW_GAMESTATELOG), pack)

		for _, value := range sceneEx.players {
			if !value.IsOnLine() {
				//踢出玩家
				sceneEx.PlayerLeave(value.Player, common.PlayerLeaveReason_DropLine, true)
			}
			if !value.IsRob {
				if !s.CoinInLimit(value.Coin) {
					sceneEx.PlayerLeave(value.Player, common.PlayerLeaveReason_Bekickout, true)
				}
				if value.Trusteeship >= model.GameParamData.PlayerWatchNum {
					sceneEx.PlayerLeave(value.Player, common.PlayerLeaveReason_LongTimeNoOp, true)
				}
			}
		}
		sceneEx.Clean()
		sceneEx.RobotLeaveHundred()
	}
}

func (this *SceneBilledStateRoulette) OnTick(s *base.Scene) {
	this.SceneBaseStateRoulette.OnTick(s)
	if sceneEx, ok := s.ExtraData.(*RouletteSceneData); ok {
		if time.Now().Sub(sceneEx.StateStartTime) > RouletteBilledTimeout {
			s.ChangeSceneState(RouletteSceneStateStart)
		}
	}
}

//////////////////////////////////////////////////////////////
//开始倒计时
//////////////////////////////////////////////////////////////
type SceneStartStateRoulette struct {
	SceneBaseStateRoulette
}

//获取当前场景状态
func (this *SceneStartStateRoulette) GetState() int {
	return RouletteSceneStateStart
}

//是否可以改变到其它场景状态
func (this *SceneStartStateRoulette) CanChangeTo(s base.SceneState) bool {
	switch s.GetState() {
	case RouletteSceneStateWait, RouletteSceneStateBet:
		return true
	}
	return false
}

//玩家操作
func (this *SceneStartStateRoulette) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if this.SceneBaseStateRoulette.OnPlayerOp(s, p, opcode, params) {
		return true
	}
	return true
}

//场景状态进入
func (this *SceneStartStateRoulette) OnEnter(s *base.Scene) {
	this.SceneBaseStateRoulette.OnEnter(s)
	RouletteBroadcastRoomState(s)
	//if sceneEx, ok := s.ExtraData.(*RouletteSceneData); ok {
	//	sceneEx.Clean()
	//}
}

//场景状态离开
func (this *SceneStartStateRoulette) OnLeave(s *base.Scene) {
}

func (this *SceneStartStateRoulette) OnTick(s *base.Scene) {
	this.SceneBaseStateRoulette.OnTick(s)
	if sceneEx, ok := s.ExtraData.(*RouletteSceneData); ok {
		if time.Now().Sub(sceneEx.StateStartTime) > RouletteStartTimeout {
			if sceneEx.CanStart() {
				s.ChangeSceneState(RouletteSceneStateBet)
			} else {
				s.ChangeSceneState(RouletteSceneStateWait)
			}
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
func (this *ScenePolicyRoulette) RegisteSceneState(state base.SceneState) {
	if state == nil {
		return
	}
	stateid := state.GetState()
	if stateid < 0 || stateid >= RouletteSceneStateMax {
		return
	}
	this.states[stateid] = state
}

//
func (this *ScenePolicyRoulette) GetSceneState(s *base.Scene, stateid int) base.SceneState {
	if stateid >= 0 && stateid < RouletteSceneStateMax {
		return this.states[stateid]
	}
	return nil
}

func init() {
	//主状态
	ScenePolicyRouletteSington.RegisteSceneState(&SceneWaitStateRoulette{})
	ScenePolicyRouletteSington.RegisteSceneState(&SceneBetStateRoulette{})
	ScenePolicyRouletteSington.RegisteSceneState(&SceneOpenPrizeStateRoulette{})
	ScenePolicyRouletteSington.RegisteSceneState(&SceneBilledStateRoulette{})
	ScenePolicyRouletteSington.RegisteSceneState(&SceneStartStateRoulette{})
	core.RegisteHook(core.HOOK_BEFORE_START, func() error {
		base.RegisteScenePolicy(common.GameId_Roulette, 0, ScenePolicyRouletteSington)
		return nil
	})
}
