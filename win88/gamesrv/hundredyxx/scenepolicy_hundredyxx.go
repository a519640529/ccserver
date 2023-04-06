package hundredyxx

import (
	"fmt"
	rule "games.yol.com/win88/gamerule/hundredyxx"
	"games.yol.com/win88/gamesrv/base"
	"games.yol.com/win88/protocol/player"
	"games.yol.com/win88/protocol/server"
	"sort"
	"time"

	"games.yol.com/win88/common"
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/hundredyxx"
	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/timer"
)

/*
pos:1-5表示top5下注最多的人,1是大富豪
	6表示神算子
	7表示自己
    8表示庄
    9表示在线玩家
*/

var ScenePolicyHundredYXXSington = &ScenePolicyHundredYXX{}

type ScenePolicyHundredYXX struct {
	base.BaseScenePolicy
	states [rule.HundredYXXSceneStateMax]base.SceneState
}

//创建场景扩展数据
func (this *ScenePolicyHundredYXX) CreateSceneExData(s *base.Scene) interface{} {
	sceneEx := NewHundredYXXSceneData(s)
	if sceneEx != nil {
		if sceneEx.init() {
			s.ExtraData = sceneEx
		}
	}
	return sceneEx
}

//创建玩家扩展数据
func (this *ScenePolicyHundredYXX) CreatePlayerExData(s *base.Scene, p *base.Player) interface{} {
	playerEx := &HundredYXXPlayerData{Player: p}
	if playerEx != nil {
		p.ExtraData = playerEx
	}
	return playerEx
}

func HundredYXXSyncChip(sceneEx *HundredYXXSceneData, force bool) {
	pack := &hundredyxx.SCHundredYXXSyncChip{
		ChipTotal: sceneEx.betInfo[:],
	}
	sceneEx.Broadcast(int(hundredyxx.HundredYXXPacketID_PACKET_SC_HYXX_SYNCCHIP), pack, 0)
}

//场景开启事件
func (this *ScenePolicyHundredYXX) OnStart(s *base.Scene) {
	logger.Logger.Trace("(this *ScenePolicyHundredYXX) OnStart, SceneId=", s.SceneId)
	sceneEx := NewHundredYXXSceneData(s)
	if sceneEx != nil {
		if sceneEx.init() {
			s.ExtraData = sceneEx
			s.ChangeSceneState(rule.HundredYXXSceneStateSendCard)
			if sceneEx.hRunBetSend != timer.TimerHandle(0) {
				timer.StopTimer(sceneEx.hRunBetSend)
				sceneEx.hRunBetSend = timer.TimerHandle(0)
			}
			if hNext, ok := common.DelayInvake(func() {
				if sceneEx.SceneState.GetState() != rule.HundredYXXSceneStateStake {
					return
				}
				HundredYXXSyncChip(sceneEx, false)
			}, nil, rule.HundredYXXSendBetTime, -1); ok {
				sceneEx.hRunBetSend = hNext
			}
		}
	}
}

//场景关闭事件
func (this *ScenePolicyHundredYXX) OnStop(s *base.Scene) {
	logger.Logger.Trace("(this *ScenePolicyHundredYXX) OnStop , SceneId=", s.SceneId)
	if sceneEx, ok := s.ExtraData.(*HundredYXXSceneData); ok {
		if sceneEx.hRunBetSend != timer.TimerHandle(0) {
			timer.StopTimer(sceneEx.hRunBetSend)
			sceneEx.hRunBetSend = timer.TimerHandle(0)
		}
		sceneEx.SaveData(true)
	}
}

//场景心跳事件
func (this *ScenePolicyHundredYXX) OnTick(s *base.Scene) {
	if s == nil {
		return
	}
	if s.SceneState != nil {
		s.SceneState.OnTick(s)
	}
}

func (this *ScenePolicyHundredYXX) GetPlayerNum(s *base.Scene) int32 {
	if s == nil {
		return 0
	}
	if sceneEx, ok := s.ExtraData.(*HundredYXXSceneData); ok {
		return int32(len(sceneEx.players))
	}
	return 0
}

//玩家进入事件
func (this *ScenePolicyHundredYXX) OnPlayerEnter(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}

	logger.Logger.Trace("(this *ScenePolicyHundredYXX) OnPlayerEnter, SceneId=", s.SceneId, " player=", p.Name)
	if sceneEx, ok := s.ExtraData.(*HundredYXXSceneData); ok {

		pos := HYXX_SELFPOS

		playerEx := &HundredYXXPlayerData{Player: p}

		playerEx.Clean()
		playerEx.SetPos(pos)

		playerEx.UnmarkFlag(base.PlayerState_Check)
		//进房时金币低于下限,状态切换到观众
		if playerEx.GetCoin() < int64(sceneEx.DbGameFree.GetBetLimit()) || playerEx.GetCoin() < int64(sceneEx.DbGameFree.GetOtherIntParams()[0]) {
			if !playerEx.IsMarkFlag(base.PlayerState_GameBreak) {
				playerEx.MarkFlag(base.PlayerState_GameBreak)
				playerEx.SyncFlag(true)
			}
		}

		for i := 0; i < len(sceneEx.seats); i++ {
			seat := sceneEx.seats[i]
			if seat != nil {
				if seat.SnId == playerEx.SnId {
					sceneEx.seats = append(sceneEx.seats[:i], sceneEx.seats[i+1:]...)
					i--
				}
			}
		}

		sceneEx.seats = append(sceneEx.seats, playerEx)
		sceneEx.players[p.SnId] = playerEx

		p.ExtraData = playerEx

		//重新计算下人数
		sceneEx.CalcuFakePlayerNum()
		//给自己发送房间信息
		HundredYXXSendRoomInfo(s, p, sceneEx, playerEx)

		s.FirePlayerEvent(p, base.PlayerEventEnter, nil)
	}
}

//玩家离开事件
func (this *ScenePolicyHundredYXX) OnPlayerLeave(s *base.Scene, p *base.Player, reason int) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyHundredYXX) OnPlayerLeave, SceneId=", s.SceneId, " player=", p.Name)
	if sceneEx, ok := s.ExtraData.(*HundredYXXSceneData); ok {
		//维护座位列表,删除离开的玩家
		if this.CanChangeCoinScene(s, p) {
			sceneEx.OnPlayerLeave(p, reason)
			s.FirePlayerEvent(p, base.PlayerEventLeave, nil)
		}
	}
}

//玩家掉线
func (this *ScenePolicyHundredYXX) OnPlayerDropLine(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyHundredYXX) OnPlayerDropLine, SceneId=", s.SceneId, " player=", p.Name)
	s.FirePlayerEvent(p, base.PlayerEventDropLine, nil)
	if sceneEx, ok := s.ExtraData.(*HundredYXXSceneData); ok {
		if sceneEx.Gaming {
			return
		}
	}
}

//玩家重连
func (this *ScenePolicyHundredYXX) OnPlayerRehold(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyHundredYXX) OnPlayerRehold, SceneId=", s.SceneId, " player=", p.Name)
	if sceneEx, ok := s.ExtraData.(*HundredYXXSceneData); ok {
		if playerEx, ok := p.ExtraData.(*HundredYXXPlayerData); ok {
			playerEx.MarkFlag(base.PlayerState_Auto)
			//发送房间信息给自己
			HundredYXXSendRoomInfo(s, p, sceneEx, playerEx)

			s.FirePlayerEvent(p, base.PlayerEventRehold, nil)
		}
	}
}

//返回房间
func (this *ScenePolicyHundredYXX) OnPlayerReturn(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyHundredYXX) OnPlayerReturn, SceneId=", s.SceneId, " player=", p.Name)
	if sceneEx, ok := s.ExtraData.(*HundredYXXSceneData); ok {
		if playerEx, ok := p.ExtraData.(*HundredYXXPlayerData); ok {
			playerEx.MarkFlag(base.PlayerState_Auto)
			//发送房间信息给自己
			HundredYXXSendRoomInfo(s, p, sceneEx, playerEx)

			s.FirePlayerEvent(p, base.PlayerEventReturn, nil)
		}
	}
}

//玩家操作
func (this *ScenePolicyHundredYXX) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if s == nil || p == nil {
		return false
	}
	if s.SceneState != nil {
		return s.SceneState.OnPlayerOp(s, p, opcode, params)
	}
	return true
}

//玩家事件
func (this *ScenePolicyHundredYXX) OnPlayerEvent(s *base.Scene, p *base.Player, evtcode int, params []int64) {
	if s == nil || p == nil {
		return
	}
	if s.SceneState != nil {
		s.SceneState.OnPlayerEvent(s, p, evtcode, params)
	}
}

//是否完成了整个牌局
func (this *ScenePolicyHundredYXX) IsCompleted(s *base.Scene) bool {
	if s == nil {
		return false
	}
	return false
}

//是否可以强制开始
func (this *ScenePolicyHundredYXX) IsCanForceStart(s *base.Scene) bool {
	return true
}

//当前状态能否换桌
func (this *ScenePolicyHundredYXX) CanChangeCoinScene(s *base.Scene, p *base.Player) bool {
	if s == nil || p == nil {
		return true
	}
	if s.SceneState != nil {
		return s.SceneState.CanChangeCoinScene(s, p)
	}
	return true
}
func (this *ScenePolicyHundredYXX) NotifyGameState(s *base.Scene) {
	if s.GetState() > 2 {
		return
	}
	sec := int(rule.HundredYXXStakeTimeout.Seconds())
	s.SyncGameState(sec, 0)
}

func HundredYXXCreateSeats(s *base.Scene, sceneEx *HundredYXXSceneData) []*hundredyxx.HundredYXXPlayerData {
	seats := make([]*HundredYXXPlayerData, 0, HYXX_RICHTOP5+2)
	if sceneEx.winTop1 != nil { //神算子
		seats = append(seats, sceneEx.winTop1)
	}
	for i := 0; i < HYXX_RICHTOP5; i++ {
		if sceneEx.betTop5[i] != nil {
			seats = append(seats, sceneEx.betTop5[i])
		}
	}
	if sceneEx.bankerSnId != -1 {
		if banker, exist := sceneEx.players[sceneEx.bankerSnId]; exist {
			seats = append(seats, banker)
		}
	}
	var datas []*hundredyxx.HundredYXXPlayerData
	for i := 0; i < len(seats); i++ {
		pp := seats[i]
		if pp != nil {
			pd := &hundredyxx.HundredYXXPlayerData{
				SnId:        proto.Int32(pp.SnId),
				Name:        proto.String(pp.Name),
				Head:        proto.Int32(pp.Head),
				Sex:         proto.Int32(pp.Sex),
				Coin:        proto.Int64(pp.GetCoin()),
				Pos:         proto.Int(pp.GetPos()),
				Flag:        proto.Int(pp.GetFlag()),
				City:        proto.String(pp.GetCity()),
				HeadOutLine: proto.Int32(pp.HeadOutLine),
				VIP:         proto.Int32(pp.VIP),
				BetTotal:    proto.Int64(pp.betTotal),
				Lately20Win: proto.Int64(int64(pp.cGetWin20)),
				Lately20Bet: proto.Int64(pp.cGetBetGig20),
				NiceId:      proto.Int32(pp.NiceId),
			}
			datas = append(datas, pd)
		}
	}
	return datas
}

func HundredYXXCreateRoomInfoPacket(s *base.Scene, sceneEx *HundredYXXSceneData, playerEx *HundredYXXPlayerData) interface{} {
	//房间信息
	pack := &hundredyxx.SCHundredYXXRoomInfo{
		RoomId:        proto.Int(s.SceneId),
		GameFreeId:    proto.Int32(s.GetGameFreeId()),
		GameId:        proto.Int(s.GetGameId()),
		RoomMode:      proto.Int(s.GetSceneMode()),
		SceneType:     proto.Int(s.SceneType),
		NumOfGames:    proto.Int(sceneEx.NumOfGames),
		State:         proto.Int(s.SceneState.GetState()),
		TimeOut:       proto.Int(s.SceneState.GetTimeout(s)),
		BankerId:      proto.Int32(sceneEx.bankerSnId),
		LimitCoin:     proto.Int32(s.DbGameFree.GetLimitCoin()),     //进房金币下限
		MaxCoinLimit:  proto.Int32(s.DbGameFree.GetMaxCoinLimit()),  //进房金币上限
		BankerLimit:   proto.Int32(s.DbGameFree.GetBanker()),        //上庄条件
		BaseScore:     proto.Int32(s.DbGameFree.GetBaseScore()),     //底分
		LowerThanKick: proto.Int32(s.DbGameFree.GetLowerThanKick()), //低于多少踢出房间
		BetLimit:      proto.Int32(s.DbGameFree.GetBetLimit()),      //低于多少不能下注
		BetRateLimit:  proto.Int32(rule.MAX_RATE),                   //押注限制(区域押注总额不能超过自身携带的百分比)
		MaxBetCoin:    s.DbGameFree.GetMaxBetCoin(),                 //各区域的押注限制
		PlayerNum:     proto.Int(sceneEx.fakePlayerNum),             //在线人数，假的
		Params:        s.GetParams(),                                //其他参数
		ParamsEx:      s.GetParamsEx(),                              //其他扩展参数
		ChipData:      sceneEx.betInfo[:],                           //筹码信息
		MyChipData:    playerEx.betInfo[:],                          //玩家各个区域的筹码信息
		DicePoints:    sceneEx.blackBox.Point[:],                    //骰子数值
		MyPreChipData: playerEx.preBetInfo[:],                       //玩家上一局各个区域的筹码信息
		Players:       HundredYXXCreateSeats(s, sceneEx),
	}

	//检查下座位上有没有自己
	hasSelf := false
	for _, pd := range pack.Players {
		if pd.GetSnId() == playerEx.SnId {
			hasSelf = true
			break
		}
	}
	//自己的信息
	if !hasSelf && playerEx != nil {
		pd := &hundredyxx.HundredYXXPlayerData{
			SnId:        proto.Int32(playerEx.SnId),
			Name:        proto.String(playerEx.Name),
			Head:        proto.Int32(playerEx.Head),
			Sex:         proto.Int32(playerEx.Sex),
			Coin:        proto.Int64(playerEx.GetCoin()),
			Pos:         proto.Int(HYXX_SELFPOS),
			Flag:        proto.Int(playerEx.GetFlag()),
			City:        proto.String(playerEx.GetCity()),
			HeadOutLine: proto.Int32(playerEx.HeadOutLine),
			VIP:         proto.Int32(playerEx.VIP),
			BetTotal:    proto.Int64(playerEx.betTotal),
			Lately20Win: proto.Int64(playerEx.cGetWin20),
			Lately20Bet: proto.Int64(playerEx.cGetBetGig20),
		}
		pack.Players = append(pack.Players, pd)
	}
	proto.SetDefaults(pack)
	return pack
}

func HundredYXXSendRoomInfo(s *base.Scene, p *base.Player, sceneEx *HundredYXXSceneData, playerEx *HundredYXXPlayerData) {
	pack := HundredYXXCreateRoomInfoPacket(s, sceneEx, playerEx)
	if p.SendToClient(int(hundredyxx.HundredYXXPacketID_PACKET_SC_HYXX_ROOMINFO), pack) {
		logger.Logger.Tracef("HundredYXXSendRoomInfo failed sceneid:%v playerid:%v", s.SceneId, p.SnId)
	}
}

func HundredYXXSendSeatInfo(s *base.Scene, sceneEx *HundredYXXSceneData) {
	pack := &hundredyxx.SCHundredYXXSeats{
		BankerId:  proto.Int32(sceneEx.bankerSnId),
		PlayerNum: proto.Int(sceneEx.fakePlayerNum),
		Data:      HundredYXXCreateSeats(s, sceneEx),
	}
	proto.SetDefaults(pack)
	s.Broadcast(int(hundredyxx.HundredYXXPacketID_PACKET_SC_HYXX_SEATS), pack, 0)
}

//////////////////////////////////////////////////////////////
//状态基类
//////////////////////////////////////////////////////////////
type SceneBaseStateHundredYXX struct {
}

func (this *SceneBaseStateHundredYXX) GetTimeout(s *base.Scene) int {
	if sceneEx, ok := s.ExtraData.(*HundredYXXSceneData); ok {
		return int(time.Now().Sub(sceneEx.StateStartTime) / time.Second)
	}
	return 0
}

func (this *SceneBaseStateHundredYXX) CanChangeTo(s base.SceneState) bool {
	return true
}

//当前状态能否换桌
func (this *SceneBaseStateHundredYXX) CanChangeCoinScene(s *base.Scene, p *base.Player) bool {
	if playerEx, ok := p.ExtraData.(*HundredYXXPlayerData); ok {
		if sceneEx, ok := s.ExtraData.(*HundredYXXSceneData); ok {
			if playerEx.betTotal != 0 {
				p.OpCode = player.OpResultCode_OPRC_Hundred_YouHadBetCannotLeave
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

func (this *SceneBaseStateHundredYXX) OnEnter(s *base.Scene) {
	if sceneEx, ok := s.ExtraData.(*HundredYXXSceneData); ok {
		sceneEx.StateStartTime = time.Now()
	}
}

func (this *SceneBaseStateHundredYXX) OnLeave(s *base.Scene) {}
func (this *SceneBaseStateHundredYXX) OnTick(s *base.Scene) {
}

//发送玩家操作情况
func (this *SceneBaseStateHundredYXX) OnPlayerSToCOp(s *base.Scene, p *base.Player, opcode int, params []int64, opResultCode hundredyxx.OpResultCode) {
	pack := &hundredyxx.SCHundredYXXOp{
		OpRetCode: opResultCode,
		OpCode:    proto.Int(opcode),
		SnId:      proto.Int32(p.SnId),
		Params:    params,
	}
	proto.SetDefaults(pack)
	p.SendToClient(int(hundredyxx.HundredYXXPacketID_PACKET_SC_HYXX_PLAYEROP), pack)
}

func (this *SceneBaseStateHundredYXX) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if playerEx, ok := p.ExtraData.(*HundredYXXPlayerData); ok {
		if sceneEx, ok := s.ExtraData.(*HundredYXXSceneData); ok {
			//上庄、下庄、上庄列表、走势列表
			switch opcode {
			case rule.HundredYXXPlayerOpUpBanker: //上庄
				if (p.GetCoin()) < int64(sceneEx.DbGameFree.GetBanker()) {
					this.OnPlayerSToCOp(s, p, opcode, params, hundredyxx.OpResultCode_OPRC_Hundred_BankerLimit)
					return true
				}
				if len(sceneEx.upplayerlist) > 0 {
					for _, v := range sceneEx.upplayerlist {
						if v.SnId == playerEx.SnId {
							this.OnPlayerSToCOp(s, p, opcode, params, hundredyxx.OpResultCode_OPRC_Hundred_BankerWaiting)
							return true
						}
					}
				}
				sceneEx.upplayerlist = append(sceneEx.upplayerlist, playerEx)
				this.OnPlayerSToCOp(s, p, opcode, params, hundredyxx.OpResultCode_OPRC_Sucess)
				return false
			case rule.HundredYXXPlayerOpDwonBanker: //下庄
				down := false
				if len(sceneEx.upplayerlist) > 0 {
					index := 0
					for _, v := range sceneEx.upplayerlist {
						if v.SnId == playerEx.SnId {
							sceneEx.upplayerlist = append(sceneEx.upplayerlist[:index], sceneEx.upplayerlist[index+1:]...)
							down = true
						}
						index++
					}
				}
				if !down {
					this.OnPlayerSToCOp(s, p, opcode, params, hundredyxx.OpResultCode_OPRC_Hundred_NotBankerWaiting)
					return true
				}
				this.OnPlayerSToCOp(s, p, opcode, params, hundredyxx.OpResultCode_OPRC_Sucess)
				return false
			case rule.HundredYXXPlayerOpNowDwonBanker: //在庄的下庄
				if sceneEx.upplayerCount == 100 {
					//玩家已经下庄
					params = append(params, 2)
				} else {
					//下庄成功
					params = append(params, 1)
				}
				sceneEx.upplayerCount = 100
				this.OnPlayerSToCOp(s, p, opcode, params, hundredyxx.OpResultCode_OPRC_Sucess)
				return true
			case rule.HundredYXXPlayerOpUpList: //上庄列表
				down := 0
				if len(sceneEx.upplayerlist) > 0 {
					for _, v := range sceneEx.upplayerlist {
						if v.SnId == playerEx.SnId {
							down = 1
							break
						}
					}
				}

				pack := &hundredyxx.SCHundredYXXUpList{
					Count:   proto.Int(len(sceneEx.upplayerlist)),
					IsExist: proto.Int(down),
				}
				for _, p := range sceneEx.upplayerlist {
					pd := &hundredyxx.HundredYXXPlayerData{
						SnId:        proto.Int32(p.SnId),
						Name:        proto.String(p.Name),
						Head:        proto.Int32(p.Head),
						Sex:         proto.Int32(p.Sex),
						Coin:        proto.Int64(p.GetCoin()),
						Pos:         proto.Int(p.GetPos()),
						Flag:        proto.Int(p.GetFlag()),
						City:        proto.String(p.GetCity()),
						HeadOutLine: proto.Int32(p.HeadOutLine),
						VIP:         proto.Int32(p.VIP),
						BetTotal:    proto.Int64(p.betTotal),
					}
					pack.Data = append(pack.Data, pd)
				}
				proto.SetDefaults(pack)
				p.SendToClient(int(hundredyxx.HundredYXXPacketID_PACKET_SC_HYXX_UPLIST), pack)
			case rule.HundredYXXTrend: //走势列表
				pack := &hundredyxx.SCHundredYXXTrend{Data: sceneEx.trends}
				proto.SetDefaults(pack)
				p.SendToClient(int(hundredyxx.HundredYXXPacketID_PACKET_SC_HYXX_TREND), pack)
			case rule.HundredYXXPlayerList:
				pack := &hundredyxx.SCHundredYXXPlayerList{
					PlayerNum: proto.Int(sceneEx.fakePlayerNum),
				}
				seats := make([]*HundredYXXPlayerData, 0, HYXX_OLTOP20+1)
				if sceneEx.winTop1 != nil { //神算子
					seats = append(seats, sceneEx.winTop1)
				}
				count := len(sceneEx.seats)
				topCnt := 0
				for i := 0; i < count && topCnt < HYXX_OLTOP20; i++ { //top20
					if sceneEx.winTop1 == nil || sceneEx.seats[i].GetSnId() != sceneEx.winTop1.GetSnId() {
						seats = append(seats, sceneEx.seats[i])
						topCnt++
					}
				}
				for i := 0; i < len(seats); i++ {
					pp := seats[i]
					if pp != nil {
						pd := &hundredyxx.HundredYXXPlayerData{
							SnId:        proto.Int32(pp.SnId),
							Name:        proto.String(pp.Name),
							Head:        proto.Int32(pp.Head),
							Sex:         proto.Int32(pp.Sex),
							Coin:        proto.Int64(pp.GetCoin()),
							Pos:         proto.Int(pp.GetPos()),
							Flag:        proto.Int(pp.GetFlag()),
							City:        proto.String(pp.GetCity()),
							Lately20Bet: proto.Int64(pp.cGetBetGig20),
							Lately20Win: proto.Int64(pp.cGetWin20),
							HeadOutLine: proto.Int32(pp.HeadOutLine),
							VIP:         proto.Int32(pp.VIP),
							BetTotal:    proto.Int64(pp.betTotal),
							NiceId:      proto.Int32(pp.NiceId),
						}
						pack.Data = append(pack.Data, pd)
					}
				}
				proto.SetDefaults(pack)
				p.SendToClient(int(hundredyxx.HundredYXXPacketID_PACKET_SC_HYXX_PLAYERLIST), pack)
			}
		}
	}
	return false
}
func (this *SceneBaseStateHundredYXX) OnPlayerEvent(s *base.Scene, p *base.Player, evtcode int, params []int64) {
	if sceneEx, ok := s.ExtraData.(*HundredYXXSceneData); ok {
		if playerEx, ok := p.ExtraData.(*HundredYXXPlayerData); ok {
			needResort := false
			switch evtcode {
			case base.PlayerEventEnter:
				if sceneEx.winTop1 == nil {
					needResort = true
				}
				if !needResort {
					for i := 0; i < HYXX_RICHTOP5; i++ {
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
					for i := 0; i < HYXX_RICHTOP5; i++ {
						if sceneEx.betTop5[i] == playerEx {
							needResort = true
							break
						}
					}
				}

			case base.PlayerEventRecharge:
				//oldflag := p.MarkBroadcast(p.GetPos() < HYXX_SELFPOS)
				p.AddCoin(int64(params[0]), common.GainWay_Pay, base.SyncFlag_ToClient, "system", p.GetScene().GetSceneName())
				//p.MarkBroadcast(oldflag)
				if p.GetCoin() >= int64(sceneEx.DbGameFree.GetOtherIntParams()[0]) && p.GetCoin() >= int64(sceneEx.DbGameFree.GetBetLimit()) {
					p.UnmarkFlag(base.PlayerState_GameBreak)
					p.SyncFlag(true)
				}
			}

			if needResort {
				sceneEx.CalcuFakePlayerNum()
				seatKey := sceneEx.Resort()
				if seatKey != sceneEx.constSeatKey {
					HundredYXXSendSeatInfo(s, sceneEx)
					sceneEx.constSeatKey = seatKey
				}
			}
		}
	}
}

func (this *SceneBaseStateHundredYXX) BroadcastRoomState(s *base.Scene, state, subState int, params ...int64) {
	pack := &hundredyxx.SCHundredYXXRoomState{
		State:    proto.Int(state),
		SubState: proto.Int(subState),
		Params:   params,
	}
	proto.SetDefaults(pack)
	s.Broadcast(int(hundredyxx.HundredYXXPacketID_PACKET_SC_HYXX_ROOMSTATE), pack, 0)
}

//////////////////////////////////////////////////////////////
//摇骰子状态
//////////////////////////////////////////////////////////////
type SceneSendCardStateHundredYXX struct {
	SceneBaseStateHundredYXX
}

func (this *SceneSendCardStateHundredYXX) GetState() int {
	return rule.HundredYXXSceneStateSendCard
}

func (this *SceneSendCardStateHundredYXX) CanChangeTo(s base.SceneState) bool {
	switch s.GetState() {
	case rule.HundredYXXSceneStateStake:
		return true
	}
	return false
}

func (this *SceneSendCardStateHundredYXX) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if this.SceneBaseStateHundredYXX.OnPlayerOp(s, p, opcode, params) {
		return true
	}
	return true
}

func (this *SceneSendCardStateHundredYXX) OnEnter(s *base.Scene) {
	this.SceneBaseStateHundredYXX.OnEnter(s)
	if sceneEx, ok := s.ExtraData.(*HundredYXXSceneData); ok {
		this.BroadcastRoomState(s, this.GetState(), 0)
		sceneEx.Clean()
		sceneEx.GameNowTime = time.Now()
		sceneEx.SetNumOfGames(sceneEx.GetNumOfGames() + 1)
		//摇骰子
		sceneEx.blackBox.Roll()
	}
}

func (this *SceneSendCardStateHundredYXX) OnLeave(s *base.Scene) {
	this.SceneBaseStateHundredYXX.OnLeave(s)
}

func (this *SceneSendCardStateHundredYXX) OnTick(s *base.Scene) {
	this.SceneBaseStateHundredYXX.OnTick(s)
	if sceneEx, ok := s.ExtraData.(*HundredYXXSceneData); ok {
		if time.Now().Sub(sceneEx.StateStartTime) > rule.HundredYXXSendCardTimeout {
			s.ChangeSceneState(rule.HundredYXXSceneStateStake)
		}
	}
}

//////////////////////////////////////////////////////////////
//押注状态
//////////////////////////////////////////////////////////////
type SceneStakeStateHundredYXX struct {
	SceneBaseStateHundredYXX
}

func (this *SceneStakeStateHundredYXX) GetState() int {
	return rule.HundredYXXSceneStateStake
}
func (this *SceneStakeStateHundredYXX) CanChangeTo(s base.SceneState) bool {
	switch s.GetState() {
	case rule.HundredYXXSceneStateOpenCard:
		return true
	}
	return false
}

func (this *SceneStakeStateHundredYXX) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if this.SceneBaseStateHundredYXX.OnPlayerOp(s, p, opcode, params) {
		return true
	}
	if playerEx, ok := p.ExtraData.(*HundredYXXPlayerData); ok {
		if sceneEx, ok := s.ExtraData.(*HundredYXXSceneData); ok {
			//玩家下注
			switch opcode {
			case rule.HundredYXXPlayerOpBet:
				if len(params) >= 2 {
					if playerEx.SnId == sceneEx.bankerSnId { //庄家不能下注
						this.OnPlayerSToCOp(s, p, opcode, params, hundredyxx.OpResultCode_OPRC_Hundred_BankerCannotBet)
						return false
					}
					betPos := int(params[0]) //下注位置
					if betPos < 0 || betPos >= rule.BetField_MAX {
						return false
					}
					betCoin := params[1] //下流金额
					chips := sceneEx.DbGameFree.GetOtherIntParams()
					if !common.InSliceInt32(chips, int32(betCoin)) {
						//筹码不符合要求
						return false
					}

					//最小面额的筹码
					minChip := int64(chips[0])
					//ownerCoin := playerEx.coin - playerEx.betTotal
					ownerCoin := playerEx.GetCoin()
					if (ownerCoin < int64(sceneEx.DbGameFree.GetBetLimit()) && playerEx.betTotal == 0) ||
						ownerCoin < int64(sceneEx.DbGameFree.GetLowerThanKick()) {
						logger.Logger.Trace("======提示低于多少不能下注======")
						this.OnPlayerSToCOp(s, p, opcode, params, hundredyxx.OpResultCode_OPRC_Hundred_CoinMustReachTheValue)
						return false
					}

					triggerBill := false
					banker, _ := sceneEx.GetBanker()
					if banker != nil {
						totalLimitBetCoin := int64(float64(banker.GetCoin()) / float64(rule.MAX_RATE))
						allBet := int64(0)
						for _, value := range sceneEx.betInfo {
							allBet += value
						}
						if allBet+betCoin > totalLimitBetCoin {
							betCoin = totalLimitBetCoin - allBet
							//对齐到最小面额的筹码
							betCoin /= minChip
							betCoin *= minChip
							if betCoin <= 0 { //不能再继续下注了
								this.OnPlayerSToCOp(s, p, opcode, params, hundredyxx.OpResultCode_OPRC_Hundred_CoinMustReachTheValue)
								return false
							}
							triggerBill = true
						}
					}

					total := betCoin + playerEx.betTotal
					if total <= playerEx.GetCoin() {
						//闲家单门下注总额是否到达上限
						maxBetCoin := sceneEx.DbGameFree.GetMaxBetCoin()
						if ok, coinLimit := playerEx.MaxChipCheck(betPos, betCoin, maxBetCoin); !ok {
							betCoin = coinLimit - playerEx.betInfo[betPos]
							//对齐到最小面额的筹码
							betCoin /= minChip
							betCoin *= minChip
							if betCoin <= 0 {
								msgPack := common.CreateSrvMsg(common.SRVMSG_CODE_DEFAULT, fmt.Sprintf("单门押注金额上限%.2f", float64(coinLimit)/100))
								p.SendToClient(int(player.PlayerPacketID_PACKET_SC_SRVMSG), msgPack)
								return false
							}
						}
					} else {
						this.OnPlayerSToCOp(s, p, opcode, params, hundredyxx.OpResultCode_OPRC_Hundred_CoinIsNotEnough)
						return false
					}

					//selfLimitBetCoin := int64(float64(playerEx.coin) / float64(rule.MAX_RATE))
					selfLimitBetCoin := playerEx.GetCoin()
					if total > selfLimitBetCoin { //自身下注总额不能超过最高赔付倍数
						betCoin = selfLimitBetCoin - playerEx.betTotal
						//对齐到最小面额的筹码
						betCoin /= minChip
						betCoin *= minChip
						if betCoin <= 0 {
							this.OnPlayerSToCOp(s, p, opcode, params, hundredyxx.OpResultCode_OPRC_Hundred_SelfBetLimitRate)
							return false
						}
					}

					playerEx.Trusteeship = 0
					//那一个闲家的下注总额
					//累积总投注额
					playerEx.TotalBet += betCoin
					playerEx.betTotal += betCoin
					playerEx.betInfo[betPos] += betCoin
					sceneEx.betInfo[betPos] += betCoin
					if playerEx.IsRob { //机器人下注
						sceneEx.betInfoRob[betPos] += betCoin
					}
					pack := &hundredyxx.SCHundredYXXBet{
						SnId:           proto.Int32(p.SnId),
						BetPos:         proto.Int32(int32(betPos)),
						BetChip:        proto.Int64(betCoin),
						BetPosBetTotal: proto.Int64(playerEx.betInfo[betPos]),
						BetTotal:       proto.Int64(playerEx.betTotal),
						Coin:           proto.Int64(playerEx.GetCoin()),
					}
					proto.SetDefaults(pack)
					//广播消息需要处理
					if playerEx.GetPos() < HYXX_SELFPOS {
						p.Broadcast(int(hundredyxx.HundredYXXPacketID_PACKET_SC_HYXX_BET), pack, 0)
					} else {
						p.SendToClient(int(hundredyxx.HundredYXXPacketID_PACKET_SC_HYXX_BET), pack)
					}
					playerEx.betDetailOrderInfo[betPos] = append(playerEx.betDetailOrderInfo[betPos], int64(betCoin))
					//已经达到庄的押注上限，提前开牌
					if triggerBill {
						msgPack := common.CreateSrvMsg(common.SRVMSG_CODE_DEFAULT, "庄家可押注额度满，直接开牌")
						s.Broadcast(int(player.PlayerPacketID_PACKET_SC_SRVMSG), msgPack, -1)
						//切换到开牌阶段
						s.ChangeSceneState(rule.HundredYXXSceneStateOpenCard)
					}
				}
				return true
			}
		}
	}
	return false
}

func (this *SceneStakeStateHundredYXX) OnEnter(s *base.Scene) {
	this.SceneBaseStateHundredYXX.OnEnter(s)
	this.BroadcastRoomState(s, this.GetState(), 0)
}
func (this *SceneStakeStateHundredYXX) OnPlayerEvent(s *base.Scene, p *base.Player, evtcode int, params []int64) {

}
func (this *SceneStakeStateHundredYXX) OnLeave(s *base.Scene) {
	this.SceneBaseStateHundredYXX.OnLeave(s)
	if sceneEx, ok := s.ExtraData.(*HundredYXXSceneData); ok {
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
		HundredYXXSyncChip(sceneEx, true)
		//对骰子结果进行调控
		sceneEx.AdjustCard()
	}
}

func (this *SceneStakeStateHundredYXX) OnTick(s *base.Scene) {
	this.SceneBaseStateHundredYXX.OnTick(s)
	if sceneEx, ok := s.ExtraData.(*HundredYXXSceneData); ok {
		if time.Now().Sub(sceneEx.StateStartTime) > rule.HundredYXXStakeTimeout {
			s.ChangeSceneState(rule.HundredYXXSceneStateOpenCard)
		}
	}
}

//////////////////////////////////////////////////////////////
//开奖状态
//////////////////////////////////////////////////////////////
type SceneOpenCardStateHundredYXX struct {
	SceneBaseStateHundredYXX
}

func (this *SceneOpenCardStateHundredYXX) GetState() int {
	return rule.HundredYXXSceneStateOpenCard
}
func (this *SceneOpenCardStateHundredYXX) CanChangeTo(s base.SceneState) bool {
	switch s.GetState() {
	case rule.HundredYXXSceneStateBilled:
		return true
	}
	return false
}

func (this *SceneOpenCardStateHundredYXX) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if this.SceneBaseStateHundredYXX.OnPlayerOp(s, p, opcode, params) {
		return true
	}
	return true
}

func (this *SceneOpenCardStateHundredYXX) OnEnter(s *base.Scene) {
	this.SceneBaseStateHundredYXX.OnEnter(s)
	if sceneEx, ok := s.ExtraData.(*HundredYXXSceneData); ok {
		params := make([]int64, 0)
		for i := 0; i < len(sceneEx.blackBox.Point); i++ {
			params = append(params, int64(sceneEx.blackBox.Point[i]))
		}
		this.BroadcastRoomState(s, this.GetState(), 0, params[:]...)

		//计算输赢
		sceneEx.CalcWin()
		//发送牌到客户端
		pack := &hundredyxx.SCHundredYXXOpenDice{}
		for i := 0; i < rule.DICE_NUM; i++ {
			pack.DicePoints = append(pack.DicePoints, sceneEx.blackBox.Point[i])
		}
		proto.SetDefaults(pack)
		s.Broadcast(int(hundredyxx.HundredYXXPacketID_PACKET_SC_HYXX_OPENDICE), pack, 0)
	}
}
func (this *SceneOpenCardStateHundredYXX) OnLeave(s *base.Scene) {
	if sceneEx, ok := s.ExtraData.(*HundredYXXSceneData); ok {
		//记录输赢走势
		sceneEx.trends = append(sceneEx.trends, sceneEx.trend)
		if len(sceneEx.trends) > HYXX_TRENDNUM {
			sceneEx.trends = sceneEx.trends[1:]
		}
		pack := &server.GWGameStateLog{
			SceneId: proto.Int(s.SceneId),
			GameLog: proto.Int32(sceneEx.trend),
		}
		s.SendToWorld(int(server.SSPacketID_PACKET_GW_GAMESTATELOG), pack)
	}
}
func (this *SceneOpenCardStateHundredYXX) OnTick(s *base.Scene) {
	this.SceneBaseStateHundredYXX.OnTick(s)
	if sceneEx, ok := s.ExtraData.(*HundredYXXSceneData); ok {
		if time.Now().Sub(sceneEx.StateStartTime) > rule.HundredYXXOpenCardTimeout {
			s.ChangeSceneState(rule.HundredYXXSceneStateBilled)
		}
	}
}

//////////////////////////////////////////////////////////////
//结算状态
//////////////////////////////////////////////////////////////
type SceneBilledStateHundredYXX struct {
	SceneBaseStateHundredYXX
}

func (this *SceneBilledStateHundredYXX) GetState() int {
	return rule.HundredYXXSceneStateBilled
}
func (this *SceneBilledStateHundredYXX) CanChangeTo(s base.SceneState) bool {
	switch s.GetState() {
	case rule.HundredYXXSceneStateSendCard:
		return true
	}
	return false
}
func (this *SceneBilledStateHundredYXX) CanChangeCoinScene(s *base.Scene, p *base.Player) bool {
	return true
}
func (this *SceneBilledStateHundredYXX) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if this.SceneBaseStateHundredYXX.OnPlayerOp(s, p, opcode, params) {
		return true
	}
	return true
}
func (this *SceneBilledStateHundredYXX) OnEnter(s *base.Scene) {
	this.SceneBaseStateHundredYXX.OnEnter(s)
	if sceneEx, ok := s.ExtraData.(*HundredYXXSceneData); ok {
		this.BroadcastRoomState(s, this.GetState(), 0, int64(sceneEx.trend))

		// 水池上下文环境
		ctx := base.CoinPoolMgr.GetCoinPoolCtx(sceneEx.Platform, sceneEx.GetGameFreeId(), sceneEx.GetGroupId())
		ctx.Controlled = sceneEx.GetCpControlled()
		sceneEx.SetCpCtx(ctx)

		sceneEx.bankerWinCoin = 0
		var bigWinner *HundredYXXPlayerData
		//各个下注区域的总金额
		countCoin := [rule.BetField_MAX]int64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
		for _, p := range sceneEx.seats {
			if p == nil {
				continue
			}
			p.returnCoin = 0
			if sceneEx.bankerSnId == p.SnId {
				continue
			}
			for i := 0; i < rule.BetField_MAX; i++ {
				if p.betInfo[i] == 0 {
					continue
				}

				if sceneEx.betFiledInfo[i].winFlag > 0 {
					//玩家中奖 玩家金额=闲家倍率*玩家下注金额*玩家输赢
					p.winCoin[i] = int64(sceneEx.betFiledInfo[i].gainRate) * p.betInfo[i]
					p.taxCoin += int64(p.winCoin[i]) * int64(s.DbGameFree.GetTaxRate()) / 10000
					//直接返还下注金额
					p.returnCoin += int64(p.betInfo[i])
				} else {
					//玩家未中奖 玩家金额= -玩家下注金额
					p.winCoin[i] = -p.betInfo[i]
					//扣除下注金额
					p.returnCoin += int64(p.betInfo[i])
				}

				//黑红梅方的总额
				countCoin[i] += p.winCoin[i]
				sceneEx.bankerWinCoin += int64(p.winCoin[i])
			}
			p.betBigRecord = append(p.betBigRecord, p.betTotal)
			p.gainCoin = p.GetWinCoin()
			if p.gainCoin == 0 {
				p.winRecord = append(p.winRecord, 0)
			} else if p.gainCoin > 0 {
				p.winRecord = append(p.winRecord, 1)
				//if p.gainCoin > p.coin {
				//	p.gainCoin = p.coin
				//}
			} else if p.gainCoin < 0 {
				p.winRecord = append(p.winRecord, 0)
				//if (-p.gainCoin) > p.coin {
				//	p.gainCoin = -p.coin
				//	p.returnCoin = 0 // 当输的钱数比自身钱多时 不返回下注金
				//}
			}

			//找出大赢家
			if bigWinner == nil {
				bigWinner = p
			} else {
				if p.gainCoin > bigWinner.gainCoin {
					bigWinner = p
				}
			}
		}
		sceneEx.bankerWinCoin = -sceneEx.bankerWinCoin
		//系统统计
		for _, p := range sceneEx.seats {
			p.cGetBetGig20 = p.GetBetGig20()
			p.cGetWin20 = p.GetWin20()
			if p.IsRob {
				continue
			}
			betCount := p.GetBetCount()
			//统计玩家输赢局数
			if betCount > 0 {
				p.SetGameTimes(p.GetGameTimes() + 1)
				if p.gainCoin > 0 {
					p.WinTimes++
				} else {
					p.FailTimes++
				}
			}
		}

		playerWinCoin := int64(0)
		playerLosCoin := int64(0)
		robotWinCoin := int64(0)
		robotLosCoin := int64(0)

		for _, p := range sceneEx.seats {
			if p == nil {
				continue
			}
			if p.betTotal > 0 {
				p.AddCoin(-(p.betTotal), common.GainWay_HundredSceneLost, base.SyncFlag_ToClient, "system", s.GetSceneName()) //下注总金额
			}
			var amount int64
			if p.gainCoin > 0 {
				//20190426税收调整
				tax := p.taxCoin //(p.gainCoin + p.returnCoin) * int64(s.DbGameFree.GetTaxRate()) / 10000
				//p.taxCoin = tax
				p.winorloseCoin = p.gainCoin + p.returnCoin
				amount = p.gainCoin + p.returnCoin - tax
				//oldflag := p.MarkBroadcast(p.pos < 6)
				p.AddCoin(int64(amount), common.GainWay_HundredSceneWin, base.SyncFlag_ToClient, "system", s.GetSceneName())
				//p.MarkBroadcast(oldflag)
				p.AddServiceFee(int64(tax))
				if !p.IsRob {
					playerWinCoin += p.gainCoin //押注时的金币没有放进金币池，这里的returnCoin也不用统计
				} else {
					robotWinCoin += p.gainCoin
				}
			} else {
				amount = p.gainCoin + p.returnCoin - p.taxCoin
				//oldflag := p.MarkBroadcast(p.pos < 6)
				p.AddCoin(int64(amount), common.GainWay_HundredSceneLost, base.SyncFlag_ToClient, "system", s.GetSceneName())
				//p.MarkBroadcast(oldflag)
				p.AddServiceFee(p.taxCoin)
				if !p.IsRob {
					playerLosCoin += p.gainCoin
				} else {
					robotLosCoin += p.gainCoin
				}
			}
			p.changeScore = amount - p.betTotal
			if p.betTotal > 0 || p.SnId == sceneEx.bankerSnId {
				//p.ReportGameEvent(p.changeScore, p.taxCoin, p.betTotal)
			}

			//统计玩家输赢
			if p.gainCoin != 0 {
				//赔率统计
				p.Statics(sceneEx.GetKeyGameId(), sceneEx.GetKeyGameId(), int64(p.gainCoin+p.taxCoin), true)
			}
		}
		//更新金币池
		if sceneEx.IsRobotBanker() {
			//系统庄，或者是机器人坐庄，用户赢取的金币，要在金币池中检测
			systemCoinOut := playerWinCoin + playerLosCoin
			if systemCoinOut > 0 { //用户赢钱，系统输钱
				base.CoinPoolMgr.PopCoin(sceneEx.GetGameFreeId(), sceneEx.GetGroupId(), sceneEx.GetPlatform(), systemCoinOut)
			}
			if systemCoinOut < 0 { //用户输钱，系统赢钱
				base.CoinPoolMgr.PushCoin(sceneEx.GetGameFreeId(), sceneEx.GetGroupId(), sceneEx.GetPlatform(), -systemCoinOut)
			}
			s.SetSystemCoinOut(-systemCoinOut)
		} else {
			//普通用户坐庄，其他用户赢取的钱是赢取庄家的钱，不用计算，
			//庄赢取机器人的钱要在金币池中检测，庄输给机器人的钱，要添加到金币池中
			//这里统计机器人输的钱，机器人输的钱，就是系统的钱，输给了庄家，是要从金币池中扣除的
			systemCoinOut := robotWinCoin + robotLosCoin
			if systemCoinOut < 0 { //机器人输钱，用户赢钱
				base.CoinPoolMgr.PopCoin(sceneEx.GetGameFreeId(), sceneEx.GetGroupId(), sceneEx.GetPlatform(), -systemCoinOut)
			}
			if systemCoinOut > 0 { //机器人赢钱，用户输钱
				base.CoinPoolMgr.PushCoin(sceneEx.GetGameFreeId(), sceneEx.GetGroupId(), sceneEx.GetPlatform(), systemCoinOut)
			}
			s.SetSystemCoinOut(systemCoinOut)
		}
		//检测用户的金币是否可以继续游戏
		for _, value := range sceneEx.players {
			if value.GetCoin() < int64(sceneEx.DbGameFree.GetBetLimit()) || value.GetCoin() < int64(sceneEx.DbGameFree.GetOtherIntParams()[0]) {
				if !value.IsMarkFlag(base.PlayerState_GameBreak) {
					value.MarkFlag(base.PlayerState_GameBreak)
					value.SyncFlag(true)
				}
			} else {
				if value.IsMarkFlag(base.PlayerState_GameBreak) {
					value.UnmarkFlag(base.PlayerState_GameBreak)
					value.SyncFlag(true)
				}
			}
		}

		//赔率不需要上报,百人场不需要考虑配桌问题
		explosion := 0
		//所有玩家结算数据
		var constBills []*hundredyxx.HundredYXXPlayerFinalWinLost
		for _, ps := range sceneEx.seats {
			constBills = append(constBills, &hundredyxx.HundredYXXPlayerFinalWinLost{
				PlayerID: proto.Int32(ps.SnId),
				Coin:     proto.Int64(ps.GetCoin()),
				GainCoin: proto.Int64(int64(ps.gainCoin - ps.taxCoin)),
				Pos:      proto.Int32(int32(ps.GetPos())),
			})
		}

		var constBetPosData []*hundredyxx.HundredYXXBetField
		for i := 0; i < rule.BetField_MAX; i++ {
			wl := sceneEx.betFiledInfo[i].winFlag
			pd := &hundredyxx.HundredYXXBetField{
				IsWin: proto.Int32(int32(wl)),
				Coin:  proto.Int64(countCoin[i]),
			}
			for _, ps := range sceneEx.seats {
				if ps.betInfo[i] > 0 {
					pdd := &hundredyxx.HundredYXXWinLost{
						Pos:      proto.Int32(int32(ps.GetPos())),
						PlayerID: proto.Int32(ps.SnId),
						WinCoin:  proto.Int64(ps.winCoin[i]),
					}
					pd.PalyerData = append(pd.PalyerData, pdd)
				}
			}
			constBetPosData = append(constBetPosData, pd)
		}
		var constBigWinner *hundredyxx.HundredYXXBigWinner
		if bigWinner != nil {
			constBigWinner = &hundredyxx.HundredYXXBigWinner{
				SnId:        proto.Int32(bigWinner.SnId),
				Name:        proto.String(bigWinner.Name),
				Head:        proto.Int32(bigWinner.Head),
				Sex:         proto.Int32(bigWinner.Sex),
				Coin:        proto.Int64(bigWinner.GetCoin()),
				City:        proto.String(bigWinner.GetCity()),
				HeadOutLine: proto.Int32(bigWinner.HeadOutLine),
				VIP:         proto.Int32(bigWinner.VIP),
				GainCoin:    proto.Int64(bigWinner.gainCoin - bigWinner.taxCoin),
			}
		}
		//拼装结算数据发给对应的客户端
		for _, ps := range sceneEx.seats {
			packBill := &hundredyxx.SCHundredYXXGameBilled{
				NumOfGame: proto.Int(sceneEx.NumOfGames),
				Explosion: proto.Int32(int32(explosion)),
				NewTrend:  proto.Int32(sceneEx.trend),
			}
			packBill.PlayerData = append(packBill.PlayerData, constBills...)
			packBill.BetPosData = append(packBill.BetPosData, constBetPosData...)
			packBill.BigWinner = constBigWinner
			proto.SetDefaults(packBill)
			ps.SendToClient(int(hundredyxx.HundredYXXPacketID_PACKET_SC_HYXX_GAMEBILLED), packBill)
		}
		//结算排序
		sort.Sort(&HundredYXXSceneData{seats: sceneEx.seats, by: func(p, q *HundredYXXPlayerData) bool {
			return q.gainCoin > p.gainCoin && (q.SnId != sceneEx.bankerSnId || p.SnId != sceneEx.bankerSnId)
		}})

		//玩家大结算
		for _, p := range sceneEx.seats {
			if p.betTotal > 0 || (!p.IsRob && sceneEx.bankerSnId == p.SnId) {
				p.SaveSceneCoinLog(p.GetCurrentCoin(), p.GetCoin()-p.GetCurrentCoin(), p.GetCoin(), int64(p.betTotal), p.taxCoin, p.winorloseCoin, 0, 0)
			}
		}
		result := sceneEx.SingleAdjustResult()
		hundredType := make([]model.HundredType, rule.BetField_MAX, rule.BetField_MAX)
		var isSave bool //有真人参与保存牌局
		for i := 0; i < rule.BetField_MAX; i++ {
			betFieldInfo := sceneEx.betFiledInfo[i]
			hundredType[i] = model.HundredType{
				CardsInfo: common.CopySliceInt32(sceneEx.blackBox.Point[:]),
				RegionId:  int32(i),
				IsWin:     betFieldInfo.winFlag,
				Rate:      betFieldInfo.gainRate,
			}
			playNum, index := 0, 0
			//统计当前房间玩家数量
			for _, o_player := range sceneEx.players {
				if o_player != nil {
					if o_player.betInfo[i] > 0 && !o_player.IsRob {
						playNum++
						isSave = true
					}
				}
			}
			if playNum == 0 {
				hundredType[i].PlayerData = nil
				continue
			}
			hundredPersons := make([]model.HundredPerson, playNum, playNum)
			for _, o_player := range sceneEx.players {
				if o_player.betInfo[i] > 0 {
					taxCoin := int64(0)
					if o_player.winCoin[i] > 0 {
						taxCoin = o_player.winCoin[i] * int64(s.DbGameFree.GetTaxRate()) / 10000
					}
					//记录庄家之外的人
					pe := model.HundredPerson{
						UserId:       o_player.SnId,
						BeforeCoin:   o_player.GetCurrentCoin(),
						AfterCoin:    o_player.GetCoin(),
						UserBetTotal: o_player.betInfo[i],
						IsFirst:      sceneEx.IsPlayerFirst(o_player.Player),
						ChangeCoin:   o_player.winCoin[i] - taxCoin,
						IsRob:        o_player.IsRob,
						WBLevel:      o_player.WBLevel,
					}
					betDetail, ok := o_player.betDetailOrderInfo[i]
					if ok {
						pe.UserBetTotalDetail = betDetail
					}
					if !o_player.IsRob {
						pe.Result = result
					}
					if !o_player.IsRob {
						//如果庄家不是真人 只记录真人
						hundredPersons[index] = pe
						index++
					}
				}
			}
			hundredType[i].PlayerData = hundredPersons
		}
		if isSave {
			_, err := model.MarshalGameNoteByHUNDRED(&hundredType)
			if err == nil {
				for _, o_player := range sceneEx.players {
					if o_player != nil && !o_player.IsRob {
						//有真人下注了
						if o_player.betTotal > 0 {
							//if o_player.betTotal > 0 || (banker != nil && !banker.IsRob && banker.gainCoin != 0) {
							totalin, totalout := int64(0), int64(0)
							wincoin := o_player.GetWinCoin()
							if wincoin > 0 {
								totalout = wincoin
							} else {
								totalin = -wincoin
							}
							winNoAnyTax := int64(0)
							if o_player.changeScore > 0 {
								winNoAnyTax = o_player.changeScore
							}
							validFlow := totalin + totalout
							validBet := common.AbsI64(totalin - totalout)
							sceneEx.SaveGamePlayerListLog(o_player.SnId,
								base.GetSaveGamePlayerListLogParam(o_player.Platform, o_player.Channel, o_player.BeUnderAgentCode, o_player.PackageID, sceneEx.logId, o_player.InviterId, totalin, totalout,
									o_player.taxCoin, 0, o_player.betTotal, winNoAnyTax, validBet, validFlow, sceneEx.IsPlayerFirst(o_player.Player), false))
						}
						o_player.SetCurrentCoin(o_player.GetCoin())
					}
				}
				trends := [][3]string{}
				if len(sceneEx.trends) > 0 {
					for _, value := range sceneEx.trends {
						trend := [3]string{}
						for i := 0; i < 3; i++ { //0 1 2
							switch value >> uint(i*8) & 255 { //鱼鸡虾蟹葫鹿 -> 012345
							case 0:
								trend[i] = "鱼"
							case 1:
								trend[i] = "鸡"
							case 2:
								trend[i] = "虾"
							case 3:
								trend[i] = "蟹"
							case 4:
								trend[i] = "葫"
							case 5:
								trend[i] = "鹿"
							}
						}
						trends = append(trends, trend)
					}
				}
				hundredType = nil
			}
		}

		//统计参与游戏次数
		//if !sceneEx.testing {
		//var playerCtxs []*server.PlayerCtx
		//for _, p := range sceneEx.seats {
		//	if p == nil || p.IsRob || p.betTotal == 0 {
		//		continue
		//	}
		//	playerCtxs = append(playerCtxs, &server.PlayerCtx{SnId: proto.Int32(p.SnId), Coin: proto.Int64(p.GetCoin())})
		//}
		//if len(playerCtxs) > 0 {
		//pack := &server.GWSceneEnd{
		//	GameFreeId: proto.Int32(sceneEx.DbGameFree.GetId()),
		//	Players:    playerCtxs,
		//}
		//proto.SetDefaults(pack)
		//sceneEx.SendToWorld(int(server.MmoPacketID_PACKET_GW_SCENEEND), pack)
		//}
		//}

		//统计下注数
		if !sceneEx.testing {
			gwPlayerBet := &server.GWPlayerBet{
				SceneId:    proto.Int(sceneEx.SceneId),
				GameFreeId: proto.Int32(sceneEx.DbGameFree.GetId()),
				RobotGain:  proto.Int64(s.GetSystemCoinOut()),
			}
			for _, p := range sceneEx.seats {
				if p == nil || p.IsRob || p.betTotal == 0 {
					continue
				}
				playerBet := &server.PlayerBet{
					SnId:       proto.Int32(p.SnId),
					Bet:        proto.Int64(p.betTotal),
					Gain:       proto.Int64(p.gainCoin),
					Tax:        proto.Int64(p.taxCoin),
					Coin:       proto.Int64(p.GetCoin()),
					GameCoinTs: proto.Int64(p.GameCoinTs),
				}
				gwPlayerBet.PlayerBets = append(gwPlayerBet.PlayerBets, playerBet)
				p.betTotal = 0
			}
			if len(gwPlayerBet.PlayerBets) > 0 {
				proto.SetDefaults(gwPlayerBet)
				sceneEx.SendToWorld(int(server.SSPacketID_PACKET_GW_PLAYERBET), gwPlayerBet)
			}
		}
	}
}

func (this *SceneBilledStateHundredYXX) OnLeave(s *base.Scene) {
	if sceneEx, ok := s.ExtraData.(*HundredYXXSceneData); ok {
		if s.CheckNeedDestroy() {
			sceneEx.SceneDestroy(true)
			return
		}
		sceneEx.logId, _ = model.AutoIncGameLogId()

		for _, value := range sceneEx.players {
			if !value.IsOnLine() {
				if value.SnId == sceneEx.bankerSnId {
					sceneEx.bankerSnId = -1
				}
				//踢出玩家
				sceneEx.PlayerLeave(value.Player, common.PlayerLeaveReason_DropLine, true)
			} else if value.IsRob {
				//maxBetCoin := s.DbGameFree.GetMaxBetCoin()
				//if !s.CoinInLimit(value.coin) || value.coin < int64(maxBetCoin[0]) {
				if !s.CoinInLimit(value.GetCoin()) {
					sceneEx.PlayerLeave(value.Player, common.PlayerLeaveReason_Bekickout, true)
					//} else if s.CoinOverMaxLimit(value.coin, value.Player) && !sceneEx.InUpplayerlist(value.SnId) && value.SnId != sceneEx.bankerSnId {
				} else if s.CoinOverMaxLimit(value.GetCoin(), value.Player) {
					s.PlayerLeave(value.Player, common.PlayerLeaveReason_Normal, true)
				} else if value.SnId == sceneEx.bankerSnId {
					if sceneEx.upplayerCount >= int32(common.RandInt(HYXX_BANKERNUMBERS/2)) {
						sceneEx.bankerSnId = -1
					}
				} else if value.Trusteeship >= 2 {
					if value.SnId != sceneEx.bankerSnId {
						sceneEx.PlayerLeave(value.Player, common.PlayerLeaveReason_LongTimeNoOp, true)
					}
				} else if value.GetCoin() < int64(sceneEx.DbGameFree.GetBetLimit()*rule.MAX_RATE) {
					s.PlayerLeave(value.Player, common.PlayerLeaveReason_Normal, true)
				}
			} else {
				if !s.CoinInLimit(value.GetCoin()) {
					if value.SnId == sceneEx.bankerSnId {
						sceneEx.bankerSnId = -1
					}
					sceneEx.PlayerLeave(value.Player, common.PlayerLeaveReason_Bekickout, true)
				} else if value.Trusteeship >= model.GameParamData.PlayerWatchNum {
					if value.SnId != sceneEx.bankerSnId {
						sceneEx.PlayerLeave(value.Player, common.PlayerLeaveReason_LongTimeNoOp, true)
					}
				}
			}
		}

		//先处理下庄家
		sceneEx.TryChangeBanker()
		//重排下座位
		seatKey := sceneEx.Resort()
		if seatKey != sceneEx.constSeatKey {
			HundredYXXSendSeatInfo(s, sceneEx)
			sceneEx.constSeatKey = seatKey
		}
	}
}

func (this *SceneBilledStateHundredYXX) OnTick(s *base.Scene) {
	this.SceneBaseStateHundredYXX.OnTick(s)
	if sceneEx, ok := s.ExtraData.(*HundredYXXSceneData); ok {
		if time.Now().Sub(sceneEx.StateStartTime) > rule.HundredYXXBilledTimeout {
			s.ChangeSceneState(rule.HundredYXXSceneStateSendCard)
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
func (this *ScenePolicyHundredYXX) RegisteSceneState(state base.SceneState) {
	if state == nil {
		return
	}
	stateid := state.GetState()
	if stateid < 0 || stateid >= rule.HundredYXXSceneStateMax {
		return
	}
	this.states[stateid] = state
}

//
func (this *ScenePolicyHundredYXX) GetSceneState(s *base.Scene, stateid int) base.SceneState {
	if stateid >= 0 && stateid < rule.HundredYXXSceneStateMax {
		return ScenePolicyHundredYXXSington.states[stateid]
	}
	return nil
}

func init() {
	ScenePolicyHundredYXXSington.RegisteSceneState(&SceneSendCardStateHundredYXX{})
	ScenePolicyHundredYXXSington.RegisteSceneState(&SceneStakeStateHundredYXX{})
	ScenePolicyHundredYXXSington.RegisteSceneState(&SceneOpenCardStateHundredYXX{})
	ScenePolicyHundredYXXSington.RegisteSceneState(&SceneBilledStateHundredYXX{})
	core.RegisteHook(core.HOOK_BEFORE_START, func() error {
		base.RegisteScenePolicy(common.GameId_HundredYXX, 0, ScenePolicyHundredYXXSington)
		return nil
	})
}

////////////////////////////////////////////////////////////////////////////////
