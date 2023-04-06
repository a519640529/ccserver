package crash

import (
	"encoding/json"
	"fmt"
	"games.yol.com/win88/common"
	. "games.yol.com/win88/gamerule/crash"
	"games.yol.com/win88/gamesrv/base"
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/crash"
	"games.yol.com/win88/protocol/player"
	"games.yol.com/win88/protocol/server"
	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/task"
	"github.com/idealeak/goserver/core/timer"
	"math"
	"sort"
	"time"
)

/*
    pos:1-6表示座位的6个位置
	7表示我的位置
    8表示在线的位置（所有其它玩家）
*/

var ScenePolicyCrashSington = &ScenePolicyCrash{}

type ScenePolicyCrash struct {
	base.BaseScenePolicy
	states [CrashSceneStateMax]base.SceneState
}

//创建场景扩展数据
func (this *ScenePolicyCrash) CreateSceneExData(s *base.Scene) interface{} {
	sceneEx := NewCrashSceneData(s)
	if sceneEx != nil {
		if sceneEx.init() {
			s.ExtraData = sceneEx
		}
	}
	return sceneEx
}

//创建玩家扩展数据
func (this *ScenePolicyCrash) CreatePlayerExData(s *base.Scene, p *base.Player) interface{} {
	playerEx := &CrashPlayerData{Player: p}
	if playerEx != nil {
		p.ExtraData = playerEx
	}
	return playerEx
}

//批量发送筹码
func CrashBatchSendBet(sceneEx *CrashSceneData, force bool) {
	sceneEx.by = CRASH_OrderByMultipleAndBetTotal
	sort.Sort(sceneEx)
	needSend := false
	pack := &crash.SCCrashSendBet{}
	//playerBetCoin := int64(0)
	for _, playerEx := range sceneEx.seats {
		//logger.Logger.Info("批量：",playerEx.SnId,playerEx.betTotal*int64(playerEx.multiple))
		if playerEx.multiple == 0 || playerEx.betTotal == 0 {
			continue
		}
		//for m, c := range playerEx.betInfo {
		chip := &crash.CrashChips{
			SnId:     proto.Int32(playerEx.SnId),
			Chip:     proto.Int64(playerEx.betTotal),
			Multiple: proto.Int32(playerEx.multiple),
		}
		//playerBetCoin += playerEx.betTotal
		if len(pack.Data) < CRASH_TOP20 {
			pack.Data = append(pack.Data, chip)
			//break
		}
		needSend = true
		//}
	}
	if needSend || force {
		pack.AllBetCoin = proto.Int64(sceneEx.allBetCoin)
		pack.AllOnlinePlayerNum = proto.Int32(int32(len(sceneEx.seats)))
		pack.AllBetPlayerNum = proto.Int32(sceneEx.allBetPlayerNum)
		proto.SetDefaults(pack)
		logger.Logger.Trace("CrashBatchSendBet", pack)
		sceneEx.Broadcast(int(crash.CrashPacketID_PACKET_SC_CRASH_SENDBET), pack, 0)
	}
}

//场景开启事件
func (this *ScenePolicyCrash) OnStart(s *base.Scene) {
	logger.Logger.Trace("(this *ScenePolicyCrash) OnStart, sceneId=", s.SceneId)
	sceneEx := NewCrashSceneData(s)
	if sceneEx != nil {
		if sceneEx.init() {
			if sceneEx.hBatchSend != timer.TimerHandle(0) {
				timer.StopTimer(sceneEx.hBatchSend)
				sceneEx.hBatchSend = timer.TimerHandle(0)
			}
			//批量发送筹码
			if hNext, ok := common.DelayInvake(func() {
				if sceneEx.SceneState.GetState() != CrashSceneStateStake {
					return
				}
				CrashBatchSendBet(sceneEx, false)

			}, nil, CrashBatchSendBetTimeout, -1); ok {
				sceneEx.hBatchSend = hNext
			}

			s.ExtraData = sceneEx
			s.ChangeSceneState(CrashSceneStateStakeAnt)
		}
	}
}

//场景关闭事件
func (this *ScenePolicyCrash) OnStop(s *base.Scene) {
	logger.Logger.Trace("(this *ScenePolicyCrash) OnStop , sceneId=", s.SceneId)
	if sceneEx, ok := s.ExtraData.(*CrashSceneData); ok {
		if sceneEx.hRunRecord != timer.TimerHandle(0) {
			timer.StopTimer(sceneEx.hRunRecord)
			sceneEx.hRunRecord = timer.TimerHandle(0)
		}
		sceneEx.SaveData()
	}
}

//场景心跳事件
func (this *ScenePolicyCrash) OnTick(s *base.Scene) {
	if s == nil {
		return
	}
	if s.SceneState != nil {
		s.SceneState.OnTick(s)
	}
}

func (this *ScenePolicyCrash) GetPlayerNum(s *base.Scene) int32 {
	if s == nil {
		return 0
	}
	if sceneEx, ok := s.ExtraData.(*CrashSceneData); ok {
		return int32(len(sceneEx.seats))
	}
	return 0
}

//玩家进入事件
func (this *ScenePolicyCrash) OnPlayerEnter(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyCrash) OnPlayerEnter, sceneId=", s.SceneId, " player=", p.Name)
	if sceneEx, ok := s.ExtraData.(*CrashSceneData); ok {
		playerEx := &CrashPlayerData{Player: p}
		if playerEx != nil {
			playerEx.Clean()

			playerEx.Pos = CRASH_OLPOS

			sceneEx.seats = append(sceneEx.seats, playerEx)

			sceneEx.players[p.SnId] = playerEx

			p.ExtraData = playerEx

			//进房时金币低于下限,状态切换到观众
			if p.Coin < int64(sceneEx.DbGameFree.GetBetLimit()) || p.Coin < int64(sceneEx.DbGameFree.GetOtherIntParams()[0]) {
				p.MarkFlag(base.PlayerState_GameBreak)
			}

			//给自己发送房间信息
			CrashSendRoomInfo(s, p, sceneEx, playerEx)

			s.FirePlayerEvent(p, base.PlayerEventEnter, nil)
		}
	}
}

//玩家离开事件
func (this *ScenePolicyCrash) OnPlayerLeave(s *base.Scene, p *base.Player, reason int) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyCrash) OnPlayerLeave, sceneId=", s.SceneId, " player=", p.Name)
	if sceneEx, ok := s.ExtraData.(*CrashSceneData); ok {
		if this.CanChangeCoinScene(s, p) {
			sceneEx.OnPlayerLeave(p, reason)
			s.FirePlayerEvent(p, base.PlayerEventLeave, []int64{int64(reason)})
		}
	}
}

//玩家掉线
func (this *ScenePolicyCrash) OnPlayerDropLine(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyCrash) OnPlayerDropLine, sceneId=", s.SceneId, " player=", p.Name)
	s.FirePlayerEvent(p, base.PlayerEventDropLine, nil)
	if sceneEx, ok := s.ExtraData.(*CrashSceneData); ok {
		if sceneEx.Gaming {
			return
		}
	}
}

//玩家重连
func (this *ScenePolicyCrash) OnPlayerRehold(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyCrash) OnPlayerRehold, sceneId=", s.SceneId, " player=", p.Name)
	if sceneEx, ok := s.ExtraData.(*CrashSceneData); ok {
		if playerEx, ok := p.ExtraData.(*CrashPlayerData); ok {
			//发送房间信息给自己
			CrashSendRoomInfo(s, p, sceneEx, playerEx)

			s.FirePlayerEvent(p, base.PlayerEventRehold, nil)
		}
	}
}

//玩家返回房间
func (this *ScenePolicyCrash) OnPlayerReturn(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyCrash) OnPlayerReturn, sceneId=", s.SceneId, " player=", p.Name)
	if sceneEx, ok := s.ExtraData.(*CrashSceneData); ok {
		if playerEx, ok := p.ExtraData.(*CrashPlayerData); ok {
			//发送房间信息给自己
			CrashSendRoomInfo(s, p, sceneEx, playerEx)

			s.FirePlayerEvent(p, base.PlayerEventReturn, nil)
		}
	}
}

//玩家操作
func (this *ScenePolicyCrash) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if s == nil || p == nil {
		return false
	}
	if s.SceneState != nil {
		return s.SceneState.OnPlayerOp(s, p, opcode, params)
	}
	return true
}

//玩家事件
func (this *ScenePolicyCrash) OnPlayerEvent(s *base.Scene, p *base.Player, evtcode int, params []int64) {
	if s == nil || p == nil {
		return
	}
	if s.SceneState != nil {
		s.SceneState.OnPlayerEvent(s, p, evtcode, params)
	}
}

//是否完成了整个牌局
func (this *ScenePolicyCrash) IsCompleted(s *base.Scene) bool {
	return false
}

//是否可以强制开始
func (this *ScenePolicyCrash) IsCanForceStart(s *base.Scene) bool {
	return true
}

//当前状态能否换桌
func (this *ScenePolicyCrash) CanChangeCoinScene(s *base.Scene, p *base.Player) bool {
	if s == nil || p == nil {
		return true
	}
	if s.SceneState != nil {
		return s.SceneState.CanChangeCoinScene(s, p)
	}
	return true
}
func (this *ScenePolicyCrash) PacketGameData(s *base.Scene) interface{} {
	if s == nil {
		return nil
	}
	//if s.SceneState != nil {
	//	if sceneEx, ok := s.ExtraData.(*CrashSceneData); ok {
	//		switch s.SceneState.GetState() {
	//		case CrashSceneStateStakeAnt:
	//			fallthrough
	//		case CrashSceneStateStake:
	//			lt := int32((CrashStakeTimeout - time.Now().Sub(sceneEx.StateStartTime)) / time.Second)
	//			pack := &server.GWDTRoomInfo{
	//				DCoin:      proto.Int(sceneEx.betInfo[CRASH_ZONE_BLACK] - sceneEx.betInfoRob[CRASH_ZONE_BLACK]),
	//				TCoin:      proto.Int(sceneEx.betInfo[CRASH_ZONE_RED] - sceneEx.betInfoRob[CRASH_ZONE_RED]),
	//				NCoin:      proto.Int(sceneEx.betInfo[CRASH_ZONE_LUCKY] - sceneEx.betInfoRob[CRASH_ZONE_LUCKY]),
	//				Onlines:    proto.Int(sceneEx.GetRealPlayerCnt()),
	//				LeftTimes:  proto.Int32(lt),
	//				CoinPool:   proto.Int64(base.CoinPoolMgr.LoadCoin(sceneEx.GetGameFreeId(), sceneEx.Platform, sceneEx.GroupId)),
	//				NumOfGames: proto.Int(sceneEx.NumOfGames),
	//				LoopNum:    proto.Int(sceneEx.LoopNum),
	//				Results:    sceneEx.ProtoResults(),
	//			}
	//			for _, value := range sceneEx.players {
	//				if value.IsRob {
	//					continue
	//				}
	//				win, lost := value.GetStaticsData(sceneEx.KeyGameId)
	//				pack.Players = append(pack.Players, &server.PlayerDTCoin{
	//					NickName: proto.String(value.Name),
	//					Snid:     proto.Int32(value.SnId),
	//					DCoin:    proto.Int(value.betInfo[CRASH_ZONE_BLACK]),
	//					TCoin:    proto.Int(value.betInfo[CRASH_ZONE_RED]),
	//					NCoin:    proto.Int(value.betInfo[CRASH_ZONE_LUCKY]),
	//					Totle:    proto.Int64(win - lost),
	//				})
	//			}
	//			return pack
	//		default:
	//			return &server.GWDTRoomInfo{}
	//		}
	//	}
	//}
	return nil
}
func (this *ScenePolicyCrash) InterventionGame(s *base.Scene, data interface{}) interface{} {
	if s == nil {
		return nil
	}
	//if s.SceneState != nil {
	//	if sceneEx, ok := s.ExtraData.(*CrashSceneData); ok {
	//		if v, ok := data.(base.InterventionResults); ok {
	//			return sceneEx.ParserResults1(v.Results, v.Webuser)
	//		}
	//		switch s.SceneState.GetState() {
	//		case CrashSceneStateStakeAnt:
	//			fallthrough
	//		case CrashSceneStateStake:
	//			if d, ok := data.(base.InterventionData); ok {
	//				if sceneEx.NumOfGames == int(d.NumOfGames) {
	//					winFlag, _ := sceneEx.CalcuResult()
	//					if d.Flag != int32(winFlag) {
	//						sceneEx.cards[0], sceneEx.cards[1] = sceneEx.cards[1], sceneEx.cards[0]
	//					}
	//					sceneEx.bIntervention = true
	//					sceneEx.webUser = d.Webuser
	//				}
	//			}
	//		default:
	//		}
	//	}
	//}
	return nil
}
func (this *ScenePolicyCrash) NotifyGameState(s *base.Scene) {
	if s.GetState() > 2 {
		return
	}
	sec := int((CrashStakeAntTimeout + CrashStakeTimeout).Seconds())
	s.SyncGameState(sec, 0)
}

//座位数据
func CrashCreateSeats(sceneEx *CrashSceneData) []*crash.CrashPlayerData {
	var datas []*crash.CrashPlayerData
	cnt := 0
	const N = CRASH_RICHTOP5 + 1
	var seats [N]*CrashPlayerData
	if sceneEx.winTop1 != nil { //神算子
		seats[cnt] = sceneEx.winTop1
		cnt++
	}
	for i := 0; i < CRASH_RICHTOP5; i++ {
		if sceneEx.betTop5[i] != nil {
			seats[cnt] = sceneEx.betTop5[i]
			cnt++
		}
	}
	for i := 0; i < N; i++ {
		if seats[i] != nil {
			pd := &crash.CrashPlayerData{
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

func CrashCreateRoomInfoPacket(s *base.Scene, sceneEx *CrashSceneData, playerEx *CrashPlayerData) proto.Message {
	pack := &crash.SCCrashRoomInfo{
		RoomId:    proto.Int(s.SceneId),
		Creator:   proto.Int32(s.GetCreator()),
		GameId:    proto.Int(s.GameId),
		RoomMode:  proto.Int(s.SceneMode),
		AgentId:   proto.Int32(s.GetAgentor()),
		SceneType: proto.Int(s.SceneType),
		Params: []int32{sceneEx.DbGameFree.GetLimitCoin(), sceneEx.DbGameFree.GetMaxCoinLimit(),
			sceneEx.DbGameFree.GetServiceFee(), sceneEx.DbGameFree.GetLowerThanKick(), sceneEx.DbGameFree.GetBaseScore(),
			0, sceneEx.DbGameFree.GetBetLimit()},
		NumOfGames:          proto.Int(sceneEx.NumOfGames),
		State:               proto.Int(s.SceneState.GetState()),
		TimeOut:             proto.Int(s.SceneState.GetTimeout(s)),
		Players:             CrashCreateSeats(sceneEx),
		Trend100Cur:         sceneEx.trend100Cur,
		Trend20Lately:       sceneEx.trend20Lately,
		Trend20CardKind:     sceneEx.trend20CardKindLately,
		AllOnlinePlayerNum:  proto.Int(len(sceneEx.seats)),
		DisbandGen:          proto.Int(sceneEx.GetDisbandGen()),
		ParamsEx:            s.GetParamsEx(),
		OtherIntParams:      s.DbGameFree.GetOtherIntParams(),
		LoopNum:             proto.Int(sceneEx.LoopNum),
		AllBetPlayerNum:     proto.Int32(sceneEx.allBetPlayerNum),
		AllBetCoin:          proto.Int64(sceneEx.allBetCoin),
		ParachutePlayerNum:  proto.Int32(sceneEx.parachutePlayerNum),
		ParachutePlayerCoin: proto.Int64(sceneEx.parachutePlayerCoin),
	}
	for _, value := range sceneEx.DbGameFree.GetMaxBetCoin() {
		pack.Params = append(pack.Params, value)
	}
	if playerEx != nil { //自己的数据
		pd := &crash.CrashPlayerData{
			SnId: proto.Int32(playerEx.SnId),
			Name: proto.String(playerEx.Name),
			Head: proto.Int32(playerEx.Head),
			Sex:  proto.Int32(playerEx.Sex),
			Coin: proto.Int64(playerEx.Coin),
			Pos:  proto.Int(CRASH_SELFPOS),
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
	if playerEx != nil { //自己的数据
		chip := &crash.CrashZoneChips{}
		//for m, c := range playerEx.betInfo {
		c := &crash.CrashChips{
			Chip:     proto.Int64(playerEx.betTotal),
			Multiple: proto.Int32(playerEx.multiple),
			SnId:     proto.Int32(playerEx.SnId),
		}
		chip.Data = append(chip.Data, c)
		//}
		pack.MyChips = append(pack.MyChips, chip)
	}
	if sceneEx.SceneState.GetState() == CrashSceneStateOpenCard || sceneEx.SceneState.GetState() == CrashSceneStateOpenCardAnt {
		sceneEx.by = CRASH_OrderByMultipleAndBetTotal
		sort.Sort(sceneEx)
		for _, v := range sceneEx.seats {
			if v.multiple == 0 || v.betTotal == 0 {
				continue
			}
			chip := &crash.CrashZoneChips{}
			//for m, c := range v.betInfo {
			if v.SnId == playerEx.SnId {
				continue
			}
			c := &crash.CrashChips{
				Chip:     proto.Int64(v.betTotal),
				Multiple: proto.Int32(v.multiple),
				SnId:     proto.Int32(v.SnId),
			}
			chip.Data = append(chip.Data, c)
			//}
			if len(pack.OtherChips) >= CRASH_TOP20 {
				break
			}
			pack.OtherChips = append(pack.OtherChips, chip)
		}
	}
	proto.SetDefaults(pack)
	return pack
}

func CrashSendRoomInfo(s *base.Scene, p *base.Player, sceneEx *CrashSceneData, playerEx *CrashPlayerData) {
	pack := CrashCreateRoomInfoPacket(s, sceneEx, playerEx)
	logger.Logger.Trace("CrashSendRoomInfo pack:", pack)
	p.SendToClient(int(crash.CrashPacketID_PACKET_SC_CRASH_ROOMINFO), pack)
}

func CrashSendSeatInfo(s *base.Scene, sceneEx *CrashSceneData) {
	pack := &crash.SCCrashSeats{
		PlayerNum: proto.Int(len(sceneEx.seats)),
		Data:      CrashCreateSeats(sceneEx),
	}
	proto.SetDefaults(pack)
	logger.Logger.Trace("CrashSendSeatInfo pack:", pack)
	s.Broadcast(int(crash.CrashPacketID_PACKET_SC_CRASH_SEATS), pack, 0)
}

//////////////////////////////////////////////////////////////
//状态基类
//////////////////////////////////////////////////////////////
type SceneBaseStateCrash struct {
}

func (this *SceneBaseStateCrash) GetTimeout(s *base.Scene) int {
	if sceneEx, ok := s.ExtraData.(*CrashSceneData); ok {
		return int(time.Now().Sub(sceneEx.StateStartTime) / time.Second)
	}
	return 0
}

func (this *SceneBaseStateCrash) CanChangeTo(s base.SceneState) bool {
	return true
}

//当前状态能否换桌
func (this *SceneBaseStateCrash) CanChangeCoinScene(s *base.Scene, p *base.Player) bool {
	if playerEx, ok := p.ExtraData.(*CrashPlayerData); ok {
		if playerEx.betTotal != 0 {
			playerEx.OpCode = player.OpResultCode_OPRC_Hundred_YouHadBetCannotLeave
			return false
		}
	}
	return true
}

func (this *SceneBaseStateCrash) OnEnter(s *base.Scene) {
	if sceneEx, ok := s.ExtraData.(*CrashSceneData); ok {
		sceneEx.StateStartTime = time.Now()
	}
}

func (this *SceneBaseStateCrash) OnLeave(s *base.Scene) {
}

func (this *SceneBaseStateCrash) OnTick(s *base.Scene) {
}

//发送玩家操作情况
func (this *SceneBaseStateCrash) SendSCPlayerOp(s *base.Scene, p *base.Player, pos int, opcode int, opRetCode crash.OpResultCode, params []int64, broadcastall bool) {
	pack := &crash.SCCrashOp{
		OpCode:    proto.Int(opcode),
		SnId:      proto.Int32(p.SnId),
		OpRetCode: opRetCode,
		Params:    params,
	}
	proto.SetDefaults(pack)
	if broadcastall {
		s.Broadcast(int(crash.CrashPacketID_PACKET_SC_CRASH_PLAYEROP), pack, 0)
	} else {
		p.SendToClient(int(crash.CrashPacketID_PACKET_SC_CRASH_PLAYEROP), pack)
	}
}

func (this *SceneBaseStateCrash) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if sceneEx, ok := s.ExtraData.(*CrashSceneData); ok {
		switch opcode {
		case CrashPlayerOpGetOLList: //在线玩家列表
			seats := make([]*CrashPlayerData, 0, CRASH_OLTOP20+1)
			if sceneEx.winTop1 != nil { //神算子
				seats = append(seats, sceneEx.winTop1)
			}
			count := len(sceneEx.seats)
			topCnt := 0
			for i := 0; i < count && topCnt < CRASH_OLTOP20; i++ { //top20
				if sceneEx.seats[i] != sceneEx.winTop1 {
					seats = append(seats, sceneEx.seats[i])
					topCnt++
				}
			}
			pack := &crash.SCCrashPlayerList{}
			for i := 0; i < len(seats); i++ {
				pack.Data = append(pack.Data, &crash.CrashPlayerData{
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

			pack.OLNum = int32(len(s.Players))
			proto.SetDefaults(pack)
			p.SendToClient(int(crash.CrashPacketID_PACKET_SC_CRASH_PLAYERLIST), pack)
			return true
		}
	}
	return false
}

func (this *SceneBaseStateCrash) OnPlayerEvent(s *base.Scene, p *base.Player, evtcode int, params []int64) {
	if sceneEx, ok := s.ExtraData.(*CrashSceneData); ok {
		if playerEx, ok := p.ExtraData.(*CrashPlayerData); ok {
			needResort := false
			switch evtcode {
			case base.PlayerEventEnter:
				if sceneEx.winTop1 == nil {
					needResort = true
				}
				if !needResort {
					for i := 0; i < CRASH_RICHTOP5; i++ {
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
					for i := 0; i < CRASH_RICHTOP5; i++ {
						if sceneEx.betTop5[i] == playerEx {
							needResort = true
							break
						}
					}
				}
			case base.PlayerEventRecharge:
				//oldflag := p.MarkBroadcast(playerEx.Pos < CRASH_SELFPOS)
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
					CrashSendSeatInfo(s, sceneEx)
					sceneEx.constSeatKey = seatKey
				}
			}
		}
	}
}

//////////////////////////////////////////////////////////////
//押注动画状态
//////////////////////////////////////////////////////////////
type SceneStakeAntStateCrash struct {
	SceneBaseStateCrash
}

//获取当前场景状态
func (this *SceneStakeAntStateCrash) GetState() int {
	return CrashSceneStateStakeAnt
}

//是否可以改变到其它场景状态
func (this *SceneStakeAntStateCrash) CanChangeTo(s base.SceneState) bool {
	if s.GetState() == CrashSceneStateStake {
		return true
	}
	return false
}

//玩家操作
func (this *SceneStakeAntStateCrash) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if this.SceneBaseStateCrash.OnPlayerOp(s, p, opcode, params) {
		return true
	}
	return true
}

//场景状态进入
func (this *SceneStakeAntStateCrash) OnEnter(s *base.Scene) {
	this.SceneBaseStateCrash.OnEnter(s)
	if sceneEx, ok := s.ExtraData.(*CrashSceneData); ok {
		logger.Logger.Tracef("(this *Scene) [%v] 场景状态进入 当前用户数%v", s.SceneId, len(sceneEx.seats))
		sceneEx.Clean()
		//if sceneEx.explode != nil {
		//	sceneEx.poker = NewPoker()
		//}

		sceneEx.NumOfGames++
		sceneEx.GameNowTime = time.Now()
		//发牌
		sceneEx.explode, sceneEx.period, sceneEx.wheel = sceneEx.poker.Next()

		logger.Logger.Infof("(this *SceneStakeAntStateCrash) OnEnter 当前轮数：%v 当前期数：%v", sceneEx.wheel, sceneEx.period)
		logger.Logger.Infof("(this *SceneStakeAntStateCrash) OnEnter 当前哈希数据：%v %v", sceneEx.explode.Hashstr, sceneEx.explode.Explode)

		task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
			return model.UptIntKVGameData("CrashPeriod", int64(sceneEx.period))
		}), nil, "UptCrashPeriodKVGameData").Start()

		pack := &crash.SCCrashRoomState{
			State: proto.Int(this.GetState()),
		}
		proto.SetDefaults(pack)
		s.Broadcast(int(crash.CrashPacketID_PACKET_SC_CRASH_ROOMSTATE), pack, 0)
	}
}

//场景状态离开
func (this *SceneStakeAntStateCrash) OnLeave(s *base.Scene) {
}

func (this *SceneStakeAntStateCrash) OnTick(s *base.Scene) {
	this.SceneBaseStateCrash.OnTick(s)
	if sceneEx, ok := s.ExtraData.(*CrashSceneData); ok {
		if time.Now().Sub(sceneEx.StateStartTime) > CrashStakeAntTimeout {
			s.ChangeSceneState(CrashSceneStateStake)
		}
	}
}

//////////////////////////////////////////////////////////////
//押注状态
//////////////////////////////////////////////////////////////
type SceneStakeStateCrash struct {
	SceneBaseStateCrash
}

func (this *SceneStakeStateCrash) GetState() int {
	return CrashSceneStateStake
}

func (this *SceneStakeStateCrash) CanChangeTo(s base.SceneState) bool {
	switch s.GetState() {
	case CrashSceneStateOpenCardAnt:
		return true
	}
	return false
}

func (this *SceneStakeStateCrash) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if this.SceneBaseStateCrash.OnPlayerOp(s, p, opcode, params) {
		return true
	}
	if playerEx, ok := p.ExtraData.(*CrashPlayerData); ok {
		if sceneEx, ok := s.ExtraData.(*CrashSceneData); ok {
			//玩家下注
			switch opcode {
			case CrashPlayerOpBet:
				if len(params) >= 2 {
					//if len(playerEx.betInfo) > 0 {
					if playerEx.multiple != 0 {
						logger.Logger.Trace("======已下注，不能重复下注======")
						this.SendSCPlayerOp(s, p, playerEx.Pos, opcode, crash.OpResultCode_OPRC_OnlyBet, params, false)
						return false
					}
					betmultiple := int32(params[0]) //下注倍数
					if betmultiple <= 0 || betmultiple > MaxMultiple {
						logger.Logger.Trace("======下注倍数错误======")
						return false
					}
					betCoin := params[1] //下注金额
					if betCoin <= 0 {
						logger.Logger.Trace("======下注金额错误======")
						return false
					}

					if p.Coin < int64(sceneEx.DbGameFree.GetBetLimit()) && playerEx.betTotal == 0 {
						logger.Logger.Trace("======提示低于多少不能下注======")
						this.SendSCPlayerOp(s, p, playerEx.Pos, opcode, crash.OpResultCode_OPRC_CoinMustReachTheValue, params, false)
						return false
					}

					total := betCoin + playerEx.betTotal
					if total <= playerEx.Coin {
						//闲家单门下注总额是否到达上限
						//maxBetCoin := sceneEx.DbGameFree.GetMaxBetCoin()
						//if int64(maxBetCoin[0]) > total {
						//	msgPack := common.CreateSrvMsg(common.SRVMSG_CODE_DEFAULT, fmt.Sprintf("该门押注金额上限%.2f", float64(maxBetCoin[0])/100))
						//	p.SendToClient(int(player.PlayerPacketID_PACKET_SC_SRVMSG), msgPack)
						//	return false
						//}
					} else {
						this.SendSCPlayerOp(s, p, playerEx.Pos, opcode, crash.OpResultCode_OPRC_CoinIsNotEnough, params, false)
						return false
					}
					playerEx.Trusteeship = 0
					//那一个闲家的下注总额
					//累积总投注额
					playerEx.TotalBet += betCoin
					playerEx.betTotal += betCoin
					playerEx.multiple = betmultiple
					sceneEx.allBetCoin += betCoin
					sceneEx.allBetPlayerNum++
					//playerEx.betInfo[betmultiple] += betCoin

					restCoin := playerEx.Coin - playerEx.betTotal
					params[1] = betCoin //修正下当前的实际下注额
					params = append(params, restCoin)
					//params = append(params, int32(sceneEx.betInfo[betPos]))
					this.SendSCPlayerOp(s, p, playerEx.Pos, opcode, crash.OpResultCode_OPRC_Sucess, params, false)

					//没钱了，要转到观战模式
					if restCoin < int64(sceneEx.DbGameFree.GetBetLimit()) {
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

func (this *SceneStakeStateCrash) OnEnter(s *base.Scene) {
	this.SceneBaseStateCrash.OnEnter(s)
	if _, ok := s.ExtraData.(*CrashSceneData); ok {
		pack := &crash.SCCrashRoomState{
			State: proto.Int(this.GetState()),
		}
		proto.SetDefaults(pack)
		s.Broadcast(int(crash.CrashPacketID_PACKET_SC_CRASH_ROOMSTATE), pack, 0)
	}
}

func (this *SceneStakeStateCrash) OnLeave(s *base.Scene) {
	this.SceneBaseStateCrash.OnLeave(s)
	if sceneEx, ok := s.ExtraData.(*CrashSceneData); ok {
		logger.Logger.Trace("SceneStakeStateCrash.OnLeave CrashBatchSendBet")
		CrashBatchSendBet(sceneEx, true)
		/*var allPlayerBet int
		for i := 0; i < CRASH_ZONE_MAX; i++ {
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

func (this *SceneStakeStateCrash) OnTick(s *base.Scene) {
	this.SceneBaseStateCrash.OnTick(s)
	if sceneEx, ok := s.ExtraData.(*CrashSceneData); ok {
		if time.Now().Sub(sceneEx.StateStartTime) > CrashStakeTimeout {
			s.ChangeSceneState(CrashSceneStateOpenCardAnt)
		}
	}
}

//////////////////////////////////////////////////////////////
//开牌动画状态
//////////////////////////////////////////////////////////////
type SceneOpenCardAntStateCrash struct {
	SceneBaseStateCrash
}

func (this *SceneOpenCardAntStateCrash) GetState() int {
	return CrashSceneStateOpenCardAnt
}

func (this *SceneOpenCardAntStateCrash) CanChangeTo(s base.SceneState) bool {
	switch s.GetState() {
	case CrashSceneStateOpenCard:
		return true
	}
	return false
}

func (this *SceneOpenCardAntStateCrash) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if this.SceneBaseStateCrash.OnPlayerOp(s, p, opcode, params) {
		return true
	}
	return true
}

func (this *SceneOpenCardAntStateCrash) OnEnter(s *base.Scene) {
	this.SceneBaseStateCrash.OnEnter(s)
	if _, ok := s.ExtraData.(*CrashSceneData); ok {
		pack := &crash.SCCrashRoomState{
			State: proto.Int(this.GetState()),
		}
		proto.SetDefaults(pack)
		s.Broadcast(int(crash.CrashPacketID_PACKET_SC_CRASH_ROOMSTATE), pack, 0)
	}
}

func (this *SceneOpenCardAntStateCrash) OnLeave(s *base.Scene) {}

func (this *SceneOpenCardAntStateCrash) OnTick(s *base.Scene) {
	this.SceneBaseStateCrash.OnTick(s)
	if sceneEx, ok := s.ExtraData.(*CrashSceneData); ok {
		if time.Now().Sub(sceneEx.StateStartTime) > CrashOpenCardAntTimeout {
			s.ChangeSceneState(CrashSceneStateOpenCard)
		}
	}
}

//////////////////////////////////////////////////////////////
//开始状态
//////////////////////////////////////////////////////////////

type SceneOpenCardStateCrash struct {
	SceneBaseStateCrash
}

func (this *SceneOpenCardStateCrash) GetState() int {
	return CrashSceneStateOpenCard
}

func (this *SceneOpenCardStateCrash) CanChangeTo(s base.SceneState) bool {
	switch s.GetState() {
	case CrashSceneStateBilled:
		return true
	}
	return false
}

func (this *SceneOpenCardStateCrash) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if this.SceneBaseStateCrash.OnPlayerOp(s, p, opcode, params) {
		return true
	}
	if sceneEx, ok := s.ExtraData.(*CrashSceneData); ok {
		if playerEx, ok := p.ExtraData.(*CrashPlayerData); ok {
			if _, ok := s.ExtraData.(*CrashSceneData); ok {
				//玩家跳伞
				switch opcode {
				case CrashPlayerOpParachute:
					betmultiple := playerEx.multiple        //原下注倍数
					nextbetmultiple := sceneEx.takeoffcurve //跳伞下注倍数
					if betmultiple <= MinMultiple || betmultiple > MaxMultiple {
						this.SendSCPlayerOp(s, p, playerEx.Pos, opcode, crash.OpResultCode_OPRC_OnlyMultiple, params, false)
						return false
					}
					if nextbetmultiple <= 0 || nextbetmultiple > betmultiple {
						this.SendSCPlayerOp(s, p, playerEx.Pos, opcode, crash.OpResultCode_OPRC_OnlyMultiple, params, false)
						return false
					}

					playerEx.multiple = nextbetmultiple
					sceneEx.parachutePlayerNum++
					sceneEx.parachutePlayerCoin += playerEx.betTotal
					pack := &crash.CrashChips{
						Chip:     proto.Int64(playerEx.betTotal),
						Multiple: proto.Int32(playerEx.multiple),
						SnId:     proto.Int32(playerEx.SnId),
					}
					s.Broadcast(int(crash.CrashPacketID_PACKET_SC_CRASH_Parachute), pack, 0)
					playerEx.parachute = true

					if len(params) == 0 {
						params = append(params, int64(nextbetmultiple))
					} else {
						params[0] = int64(nextbetmultiple) //修正下当前的实际下注倍率
					}

					this.SendSCPlayerOp(s, p, playerEx.Pos, opcode, crash.OpResultCode_OPRC_Sucess, params, false)
					return true
				}
			}
		}
	}
	return true
}

func (this *SceneOpenCardStateCrash) OnEnter(s *base.Scene) {
	this.SceneBaseStateCrash.OnEnter(s)
	if _, ok := s.ExtraData.(*CrashSceneData); ok {
		pack := &crash.SCCrashRoomState{
			State: proto.Int(this.GetState()),
		}
		proto.SetDefaults(pack)
		s.Broadcast(int(crash.CrashPacketID_PACKET_SC_CRASH_ROOMSTATE), pack, 0)
		logger.Logger.Trace("SceneOpenCardStateCrash ->", pack)
	}
}
func (this *SceneOpenCardStateCrash) OnLeave(s *base.Scene) {
	this.SceneBaseStateCrash.OnLeave(s)
	if sceneEx, ok := s.ExtraData.(*CrashSceneData); ok {
		//同步开奖结果
		pack := &server.GWGameStateLog{
			SceneId: proto.Int(s.SceneId),
			GameLog: proto.Int(int(sceneEx.explode.Explode)),
			LogCnt:  proto.Int(len(sceneEx.trend100Cur)),
		}
		s.SendToWorld(int(server.SSPacketID_PACKET_GW_GAMESTATELOG), pack)
	}
}

func (this *SceneOpenCardStateCrash) OnTick(s *base.Scene) {
	this.SceneBaseStateCrash.OnTick(s)
	if sceneEx, ok := s.ExtraData.(*CrashSceneData); ok {
		t := int64(time.Now().Sub(sceneEx.StateStartTime) / time.Millisecond)
		curTakeoffCurve := int32(math.Pow(float64(t)/600, 2)) + 100

		if curTakeoffCurve > sceneEx.explode.Explode {
			sceneEx.takeoffcurve = sceneEx.explode.Explode
		} else {
			sceneEx.takeoffcurve = curTakeoffCurve
		}

		for _, v := range sceneEx.seats {
			if v.parachute || v.multiple == 0 {
				continue
			}

			//for m,c := range v.betInfo {
			if v.multiple < sceneEx.takeoffcurve {
				pack := &crash.CrashChips{
					Chip:     proto.Int64(v.betTotal),
					Multiple: proto.Int32(v.multiple),
					SnId:     proto.Int32(v.SnId),
				}
				sceneEx.parachutePlayerNum++
				sceneEx.parachutePlayerCoin += v.betTotal
				s.Broadcast(int(crash.CrashPacketID_PACKET_SC_CRASH_Parachute), pack, 0)
				v.parachute = true
				logger.Logger.Infof("玩家跳伞：%v", pack)
			}
			//}
		}

		pack := &crash.CrashTime{
			Millisecond: proto.Int64(t),
			Multiple:    proto.Int32(sceneEx.takeoffcurve),
		}
		s.Broadcast(int(crash.CrashPacketID_PACKET_SC_CRASH_SENDTime), pack, 0)
		if pack.Multiple >= sceneEx.explode.Explode {
			s.ChangeSceneState(CrashSceneStateBilled)
		}
		if time.Now().Sub(sceneEx.StateStartTime) > CrashOpenCardTimeout {
			s.ChangeSceneState(CrashSceneStateBilled)
		}
	}
}

//////////////////////////////////////////////////////////////
//结算状态
//////////////////////////////////////////////////////////////
type SceneBilledStateCrash struct {
	SceneBaseStateCrash
}

func (this *SceneBilledStateCrash) GetState() int {
	return CrashSceneStateBilled
}

func (this *SceneBilledStateCrash) CanChangeTo(s base.SceneState) bool {
	switch s.GetState() {
	case CrashSceneStateStakeAnt:
		return true
	}
	return false
}

func (this *SceneBilledStateCrash) CanChangeCoinScene(s *base.Scene, p *base.Player) bool {
	return true
}

func (this *SceneBilledStateCrash) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if this.SceneBaseStateCrash.OnPlayerOp(s, p, opcode, params) {
		return true
	}
	return true
}

func (this *SceneBilledStateCrash) OnEnter(s *base.Scene) {
	this.SceneBaseStateCrash.OnEnter(s)
	if sceneEx, ok := s.ExtraData.(*CrashSceneData); ok {
		sceneEx.AddLoopNum()
		sceneEx.PushTrend(int32(sceneEx.explode.Explode))
		pack := &crash.SCCrashRoomState{
			State:  proto.Int(this.GetState()),
			Params: []int32{int32(sceneEx.explode.Explode)},
		}
		proto.SetDefaults(pack)
		s.Broadcast(int(crash.CrashPacketID_PACKET_SC_CRASH_ROOMSTATE), pack, 0)

		// 水池上下文环境
		sceneEx.CpCtx = base.CoinPoolMgr.GetCoinPoolCtx(sceneEx.Platform, sceneEx.GetGameFreeId(), sceneEx.GroupId)
		sceneEx.CpCtx.Controlled = sceneEx.CpControlled

		playerTotalGain := int64(0)
		sysGain := int64(0)
		olTotalGain := int64(0)
		olTotalBet := int64(0)

		var bigWinner *CrashPlayerData
		//计算出所有位置的输赢
		count := len(sceneEx.seats)
		for i := 0; i < count; i++ {
			playerEx := sceneEx.seats[i]
			if playerEx != nil {
				//var tax int64
				playerEx.gainCoin = 0
				//for m, c := range playerEx.betInfo {
				if playerEx.multiple < int32(sceneEx.explode.Explode) && playerEx.multiple >= MinMultiple {
					playerEx.winCoin[playerEx.multiple] = int64(playerEx.multiple) * playerEx.betTotal / 100
					playerEx.gainCoin += playerEx.winCoin[playerEx.multiple]

					playerEx.taxCoin = (playerEx.gainCoin - playerEx.betTotal) * int64(s.DbGameFree.GetTaxRate()) / 10000
				} else {
					playerEx.winCoin[playerEx.multiple] = -playerEx.betTotal
				}
				//}

				playerEx.winorloseCoin = playerEx.gainCoin
				playerEx.gainWinLost = playerEx.gainCoin - playerEx.taxCoin - playerEx.betTotal

				if playerEx.gainWinLost > 0 {
					playerEx.winRecord = append(playerEx.winRecord, 1)
				} else {
					playerEx.winRecord = append(playerEx.winRecord, 0)
				}
				sysGain += playerEx.gainWinLost
				if playerEx.Pos == CRASH_OLPOS {
					olTotalGain += playerEx.gainWinLost
				}
				playerEx.betBigRecord = append(playerEx.betBigRecord, playerEx.betTotal)
				playerEx.RecalcuLatestBet20()
				playerEx.RecalcuLatestWin20()
				if playerEx.gainWinLost > 0 {
					//oldflag := playerEx.MarkBroadcast(playerEx.Pos < CRASH_SELFPOS)
					playerEx.AddCoin(int64(playerEx.gainWinLost), common.GainWay_HundredSceneWin, base.SyncFlag_ToClient, "system", s.GetSceneName())
					//playerEx.MarkBroadcast(oldflag)

					//累积税收
					playerEx.AddServiceFee(int64(playerEx.taxCoin))
				} else {
					//oldflag := playerEx.MarkBroadcast(playerEx.Pos < CRASH_SELFPOS)
					playerEx.AddCoin(int64(playerEx.gainWinLost), common.GainWay_HundredSceneLost, base.SyncFlag_ToClient, "system", s.GetSceneName())
					//playerEx.MarkBroadcast(oldflag)
				}
				if playerEx.Coin >= int64(sceneEx.DbGameFree.GetBetLimit()) {
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

		var constBills []*crash.CrashBill
		if sceneEx.winTop1 != nil && sceneEx.winTop1.betTotal != 0 { //神算子
			constBills = append(constBills, &crash.CrashBill{
				SnId:     proto.Int32(sceneEx.winTop1.SnId),
				Coin:     proto.Int64(sceneEx.winTop1.Coin),
				GainCoin: proto.Int64(int64(sceneEx.winTop1.gainWinLost)),
			})
		}
		if olTotalBet != 0 { //在线玩家
			constBills = append(constBills, &crash.CrashBill{
				SnId:     proto.Int(0), //在线玩家约定snid=0
				Coin:     proto.Int64(olTotalGain),
				GainCoin: proto.Int64(olTotalGain),
			})
		} else {
			if sysGain < 0 {
				constBills = append(constBills, &crash.CrashBill{
					SnId:     proto.Int(0), //在线玩家约定snid=0
					Coin:     proto.Int64(sysGain * -1),
					GainCoin: proto.Int64(sysGain * -1),
				})
			}
		}
		for i := 0; i < CRASH_RICHTOP5; i++ { //富豪前5名
			if sceneEx.betTop5[i] != nil && sceneEx.betTop5[i].betTotal != 0 {
				constBills = append(constBills, &crash.CrashBill{
					SnId:     proto.Int32(sceneEx.betTop5[i].SnId),
					Coin:     proto.Int64(sceneEx.betTop5[i].Coin),
					GainCoin: proto.Int64(int64(sceneEx.betTop5[i].gainWinLost)),
				})
			}
		}
		for i := 0; i < count; i++ {
			playerEx := sceneEx.seats[i]
			if playerEx != nil && playerEx.IsOnLine() && !playerEx.IsMarkFlag(base.PlayerState_Leave) {
				pack := &crash.SCCrashBilled{
					LoopNum: proto.Int(sceneEx.LoopNum),
				}
				pack.BillData = append(pack.BillData, constBills...)
				pack.BillData = append(pack.BillData, &crash.CrashBill{
					SnId:     proto.Int32(playerEx.SnId),
					Coin:     proto.Int64(playerEx.Coin),
					GainCoin: proto.Int64(int64(playerEx.gainWinLost)),
				})
				//if sceneEx.seats[i].betTotal != 0 {
				//	pack.BillData = append(pack.BillData, &crash.CrashBill{
				//		SnId:     proto.Int32(sceneEx.seats[i].SnId),
				//		Coin:     proto.Int64(sceneEx.seats[i].Coin),
				//		GainCoin: proto.Int64(int64(sceneEx.seats[i].gainWinLost)),
				//	})
				//}
				if bigWinner != nil {
					pack.BigWinner = &crash.CrashBigWinner{
						SnId:        proto.Int32(bigWinner.SnId),
						Name:        proto.String(bigWinner.Name),
						Head:        proto.Int32(bigWinner.Head),
						HeadOutLine: proto.Int32(bigWinner.HeadOutLine),
						VIP:         proto.Int32(bigWinner.VIP),
						Sex:         proto.Int32(bigWinner.Sex),
						Coin:        proto.Int64(bigWinner.Coin),
						GainCoin:    proto.Int(int(bigWinner.gainWinLost)),
						City:        proto.String(bigWinner.GetCity()),
					}
				}
				proto.SetDefaults(pack)
				playerEx.SendToClient(int(crash.CrashPacketID_PACKET_SC_CRASH_GAMEBILLED), pack)
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

		//s.SystemCoinOut = -int64(sceneEx.GetAllPlayerWinScore(result, luckyKind, true))

		////////////////////////水池更新提前////////////////
		//更新金币池
		result := sceneEx.explode.Explode

		sysOut := sceneEx.GetAllPlayerWinScore(result)
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
		//if !sceneEx.Testing && sceneEx.bIntervention {
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
		//}
		/////////////////////////////////////统计牌局详细记录
		hundredType := &model.CrashType{
			RegionId: int32(0),
			//IsWin:     sceneEx.isWin[i],
			CardsInfo: []int32{int32(sceneEx.explode.Explode)},
			Hash:      sceneEx.explode.Hashstr,
			Rate:      int(sceneEx.explode.Explode),
			Wheel:     sceneEx.wheel,
			Period:    sceneEx.period,
		}
		var isSave bool //有真人参与保存牌局

		//有没有真人都记录
		isSave = true
		playNum, index := 0, 0
		for _, o_player := range sceneEx.players {
			if o_player != nil && o_player.multiple != 0 && !o_player.IsRob {
				playNum++
				isSave = true
			}
		}

		hundredPerson := make([]model.CrashPerson, playNum, playNum)
		for _, o_player := range sceneEx.players {
			if o_player != nil && !o_player.IsRob && o_player.multiple != 0 {

				pe := model.CrashPerson{
					UserId:       o_player.SnId,
					UserBetTotal: int64(o_player.betTotal),
					UserMultiple: o_player.multiple,
					ChangeCoin:   int64(o_player.gainCoin),
					BeforeCoin:   o_player.GetCurrentCoin(),
					AfterCoin:    o_player.Coin,
					IsRob:        o_player.IsRob,
					IsFirst:      sceneEx.IsPlayerFirst(o_player.Player),
					WBLevel:      o_player.WBLevel,
					//Result:       o_player.result,
					Tax: o_player.taxCoin,
				}

				hundredPerson[index] = pe
				index++
			}
		}
		hundredType.PlayerData = hundredPerson

		if isSave {
			info, err := model.MarshalGameNoteByHUNDRED(hundredType)
			if err == nil {
				//logid, _ := model.AutoIncGameLogId()
				//logid := fmt.Sprintf("%v%v%05d", sceneEx.wheel, time.Now().Unix(), sceneEx.period) //strconv.Itoa(sceneEx.period)

				//logid = fmt.Sprintf("%v-%v-%v",logid,sceneEx.wheel,sceneEx.period)
				//logger.Logger.Info("当前期数:", logid)

				logid := fmt.Sprintf("%05d%05d", sceneEx.wheel, sceneEx.period)
				for _, o_player := range sceneEx.players {
					if o_player != nil {
						if !o_player.IsRob {
							if o_player.betTotal > 0 {
								totalin, totalout := int64(0), int64(0)
								wincoin := int64(0)

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
						trends = append(trends, fmt.Sprintf("%v", v))
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

func (this *SceneBilledStateCrash) OnLeave(s *base.Scene) {
	if sceneEx, ok := s.ExtraData.(*CrashSceneData); ok {
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
			CrashSendSeatInfo(s, sceneEx)
			sceneEx.constSeatKey = seatKey
		}

		if s.CheckNeedDestroy() {
			sceneEx.SceneDestroy(true)
		}
	}
}

func (this *SceneBilledStateCrash) OnTick(s *base.Scene) {
	this.SceneBaseStateCrash.OnTick(s)
	if sceneEx, ok := s.ExtraData.(*CrashSceneData); ok {
		if time.Now().Sub(sceneEx.StateStartTime) > CrashBilledTimeout {
			s.ChangeSceneState(CrashSceneStateStakeAnt)
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
func (this *ScenePolicyCrash) RegisteSceneState(state base.SceneState) {
	if state == nil {
		return
	}
	stateid := state.GetState()
	if stateid < 0 || stateid >= CrashSceneStateMax {
		return
	}
	this.states[stateid] = state
}

//
func (this *ScenePolicyCrash) GetSceneState(s *base.Scene, stateid int) base.SceneState {
	if stateid >= 0 && stateid < CrashSceneStateMax {
		return ScenePolicyCrashSington.states[stateid]
	}
	return nil
}

func init() {
	ScenePolicyCrashSington.RegisteSceneState(&SceneStakeAntStateCrash{})
	ScenePolicyCrashSington.RegisteSceneState(&SceneStakeStateCrash{})
	ScenePolicyCrashSington.RegisteSceneState(&SceneOpenCardAntStateCrash{})
	ScenePolicyCrashSington.RegisteSceneState(&SceneOpenCardStateCrash{})
	ScenePolicyCrashSington.RegisteSceneState(&SceneBilledStateCrash{})
	core.RegisteHook(core.HOOK_BEFORE_START, func() error {
		base.RegisteScenePolicy(common.GameId_Crash, 0, ScenePolicyCrashSington)
		base.RegisteScenePolicy(common.GameId_Crash, 1, ScenePolicyCrashSington)
		return nil
	})
}

////////////////////////////////////////////////////////////////////////////////
