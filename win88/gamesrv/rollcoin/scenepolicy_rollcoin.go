package rollcoin

import (
	"fmt"
	"games.yol.com/win88/common"
	. "games.yol.com/win88/gamerule/rollcoin"
	"games.yol.com/win88/gamesrv/base"
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/player"
	"games.yol.com/win88/protocol/rollcoin"
	"games.yol.com/win88/protocol/server"
	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/logger"
	"math/rand"
	"sort"
	"time"
)

//var RollCoinTargetList = []int32{1, 2, 3, 0, 1, 6, 2, 3, 0, 1, 2, 3, 7, 0, 1, 2, 3, 0, 1, 5, 2, 3, 0, 1, 2, 3, 4, 0}
var ScenePolicyRollCoinSington = &ScenePolicyRollCoin{}

type ScenePolicyRollCoin struct {
	base.BaseScenePolicy
	states [RollCoinSceneStateMax]base.SceneState
}

//创建场景扩展数据
func (this *ScenePolicyRollCoin) CreateSceneExData(s *base.Scene) interface{} {
	logger.Logger.Trace("(this *ScenePolicyRollCoin) CreateSceneExData, sceneId=", s.SceneId)
	sceneEx := NewRollCoinSceneData(s)
	if sceneEx != nil {
		if sceneEx.init() {
			s.ExtraData = sceneEx
		}
	}
	return sceneEx
}

//创建玩家扩展数据
func (this *ScenePolicyRollCoin) CreatePlayerExData(s *base.Scene, p *base.Player) interface{} {
	playerEx := NewRollCoinPlayerData(p)
	if playerEx != nil {
		p.ExtraData = playerEx
	}
	return playerEx
}

//场景开启事件
func (this *ScenePolicyRollCoin) OnStart(s *base.Scene) {
	this.BaseScenePolicy.OnStart(s)
	logger.Logger.Trace("(this *ScenePolicyRollCoin) OnStart, sceneId=", s.SceneId)
	sceneEx := NewRollCoinSceneData(s)
	if sceneEx != nil {
		if sceneEx.init() {
			this.BaseScenePolicy.OnStart(s)
			s.ExtraData = sceneEx
			//临时保证系统不亏损的数值系统(场次开通前把池子清空),待正式数值调整好，这里要去掉
			/*poolCoin := base.CoinPoolMgr.LoadCoin(sceneEx.gamefreeId, sceneEx.platform, sceneEx.groupId)
			if poolCoin > 0 {
				base.CoinPoolMgr.PopCoin(sceneEx.gamefreeId, sceneEx.platform, poolCoin)
			}*/

			s.ChangeSceneState(RollCoinSceneStateWait)
		}
	}
}

//场景关闭事件
func (this *ScenePolicyRollCoin) OnStop(s *base.Scene) {
	logger.Logger.Trace("(this *ScenePolicyRollCoin) OnStop , SceneId=", s.SceneId)
	this.BaseScenePolicy.OnStop(s)
}

//场景心跳事件
func (this *ScenePolicyRollCoin) OnTick(s *base.Scene) {
	this.BaseScenePolicy.OnTick(s)
	if s == nil {
		return
	}
	this.BaseScenePolicy.OnTick(s)
	if s.SceneState != nil {
		s.SceneState.OnTick(s)
	}
}

//玩家进入事件
func (this *ScenePolicyRollCoin) OnPlayerEnter(s *base.Scene, p *base.Player) {
	this.BaseScenePolicy.OnPlayerEnter(s, p)
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyRollCoin) OnPlayerEnter, sceneId=", s.SceneId, " player=", p.Name)
	if sceneEx, ok := s.ExtraData.(*RollCoinSceneData); ok {
		s.FirePlayerEvent(p, base.PlayerEventEnter, nil)
		playerEx := NewRollCoinPlayerData(p)
		playerEx.init()
		this.BaseScenePolicy.OnPlayerEnter(s, p)
		sceneEx.players[p.SnId] = playerEx
		p.ExtraData = playerEx

		if playerEx.Coin < int64(sceneEx.CoinLimit) || playerEx.Coin < int64(sceneEx.Rollcoin[0]) { //进入房间，金币少于规定值，只能观看，不能下注
			playerEx.MarkFlag(base.PlayerState_GameBreak)
			playerEx.SyncFlag(true)
		}

		//向其他人广播玩家进入信息
		pack := &rollcoin.SCRollCoinPlayerNum{
			PlayerNum: proto.Int(sceneEx.CountPlayer()),
		}
		proto.SetDefaults(pack)
		s.Broadcast(int(rollcoin.PACKETID_ROLLCOIN_SC_PLAYERNUM), pack, 0)
		logger.Logger.Trace("SCRollCoinPlayerNum:", pack)
		//给自己发送房间信息
		this.SendRoomInfo(s, p, sceneEx, playerEx)
	}
}

//玩家离开事件
func (this *ScenePolicyRollCoin) OnPlayerLeave(s *base.Scene, p *base.Player, reason int) {
	this.BaseScenePolicy.OnPlayerLeave(s, p, reason)
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyRollCoin) OnPlayerLeave, sceneId=", s.SceneId, " player=", p.Name)
	this.BaseScenePolicy.OnPlayerLeave(s, p, reason)
	if sceneEx, ok := s.ExtraData.(*RollCoinSceneData); ok {
		s.FirePlayerEvent(p, base.PlayerEventLeave, nil)
		sceneEx.OnPlayerLeave(p, reason)
	}
}

//玩家掉线
func (this *ScenePolicyRollCoin) OnPlayerDropLine(s *base.Scene, p *base.Player) {
	logger.Logger.Trace("(this *ScenePolicyRollCoin) OnPlayerDropLine, sceneId=", s.SceneId, " player=", p.Name)
	this.BaseScenePolicy.OnPlayerDropLine(s, p)
	if s == nil || p == nil {
		return
	}
	this.BaseScenePolicy.OnPlayerDropLine(s, p)
	s.FirePlayerEvent(p, base.PlayerEventDropLine, nil)
}

//玩家重连
func (this *ScenePolicyRollCoin) OnPlayerRehold(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyRollCoin) OnPlayerRehold, sceneId=", s.SceneId, " player=", p.Name)
	this.BaseScenePolicy.OnPlayerRehold(s, p)
	if sceneEx, ok := s.ExtraData.(*RollCoinSceneData); ok {
		if playerEx, ok := p.ExtraData.(*RollCoinPlayerData); ok {
			pack := &rollcoin.SCRollCoinPlayerNum{
				PlayerNum: proto.Int(sceneEx.CountPlayer()),
			}
			proto.SetDefaults(pack)
			playerEx.SendToClient(int(rollcoin.PACKETID_ROLLCOIN_SC_PLAYERNUM), pack)
			//发送房间信息给自己
			this.SendRoomInfo(s, p, sceneEx, playerEx)
			s.FirePlayerEvent(p, base.PlayerEventRehold, nil)
		}
	}
}

//玩家返回房间
func (this *ScenePolicyRollCoin) OnPlayerReturn(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Trace("(this *ScenePolicyRollCoin) OnPlayerReturn,sceneId =", s.GetSceneId(), " player= ", p.Name)
	this.BaseScenePolicy.OnPlayerReturn(s, p)
	if sceneEx, ok := s.GetExtraData().(*RollCoinSceneData); ok {
		if playerEx, ok := p.GetExtraData().(*RollCoinPlayerData); ok {
			pack := &rollcoin.SCRollCoinPlayerNum{
				PlayerNum: proto.Int(sceneEx.CountPlayer()),
			}
			proto.SetDefaults(pack)
			playerEx.SendToClient(int(rollcoin.PACKETID_ROLLCOIN_SC_PLAYERNUM), pack)
			//发送房间信息给自己
			this.SendRoomInfo(s, p, sceneEx, playerEx)
			s.FirePlayerEvent(p, base.PlayerEventReturn, nil)
		}
	}
}

//玩家操作
func (this *ScenePolicyRollCoin) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if s == nil || p == nil {
		return false
	}
	this.BaseScenePolicy.OnPlayerOp(s, p, opcode, params)
	if s.SceneState != nil {
		return s.SceneState.OnPlayerOp(s, p, opcode, params)
	}
	return true
}

//玩家事件
func (this *ScenePolicyRollCoin) OnPlayerEvent(s *base.Scene, p *base.Player, evtcode int, params []int64) {
	if s == nil || p == nil {
		return
	}
	this.BaseScenePolicy.OnPlayerEvent(s, p, evtcode, params)
	if s.SceneState != nil {
		s.SceneState.OnPlayerEvent(s, p, evtcode, params)
	}
}

//观众进入事件
func (this *ScenePolicyRollCoin) OnAudienceEnter(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyRollCoin) OnAudienceEnter, sceneId=", s.SceneId, " player=", p.Name)
	this.BaseScenePolicy.OnAudienceEnter(s, p)
	if sceneEx, ok := s.ExtraData.(*RollCoinSceneData); ok {
		//给自己发送房间信息
		this.SendRoomInfo(s, p, sceneEx, nil)
	}
}

//观众掉线
func (this *ScenePolicyRollCoin) OnAudienceDropLine(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyRollCoin) OnAudienceDropLine, sceneId=", s.SceneId, " player=", p.Name)
	this.BaseScenePolicy.OnAudienceDropLine(s, p)
	s.AudienceLeave(p, common.PlayerLeaveReason_DropLine)
}

//是否完成了整个牌局
func (this *ScenePolicyRollCoin) IsCompleted(s *base.Scene) bool {
	if s == nil {
		return false
	}
	return false
}

//是否可以强制开始
func (this *ScenePolicyRollCoin) IsCanForceStart(s *base.Scene) bool {
	if sceneEx, ok := s.ExtraData.(*RollCoinSceneData); ok {
		return len(sceneEx.players) > 0
	}
	return false
}

//强制开始
func (this *ScenePolicyRollCoin) ForceStart(s *base.Scene) {
	s.ChangeSceneState(RollCoinSceneStateWait)
}

//当前状态能否换桌
func (this *ScenePolicyRollCoin) CanChangeCoinScene(s *base.Scene, p *base.Player) bool {
	if s == nil || p == nil {
		return false
	}
	if sceneEx, ok := s.ExtraData.(*RollCoinSceneData); ok {
		if sceneEx.Banker != nil { //有庄家
			if sceneEx.Banker.SnId == p.SnId { //庄家是自己
				p.OpCode = player.OpResultCode_OPRC_Hundred_YouHadBankerCannotLeave
				return false //不能离开
			}
		}
	}
	if playerEx, ok := p.ExtraData.(*RollCoinPlayerData); ok {
		if playerEx.GetTotlePushCoin() > 0 { //没押注可以离开
			p.OpCode = player.OpResultCode_OPRC_Hundred_YouHadBetCannotLeave
			return false
		}
	}
	if s.SceneState != nil {
		return s.SceneState.CanChangeCoinScene(s, p)
	}
	return true
}

func (this *ScenePolicyRollCoin) SendRoomInfo(s *base.Scene, p *base.Player, sceneEx *RollCoinSceneData, playerEx *RollCoinPlayerData) {
	pack := &rollcoin.SCRollCoinRoomInfo{
		RollLog:   sceneEx.WinLog,
		SceneType: proto.Int(s.SceneType),
		State:     proto.Int(s.SceneState.GetState()),
		TimeOut:   proto.Int(s.SceneState.GetTimeout(s)),
		GameId:    proto.Int(common.GameId_RollCoin),
		RoomMode:  proto.Int(s.SceneMode),
		SceneId:   proto.Int(sceneEx.SceneId),
		ParamsEx:  s.GetParamsEx(),
		Params: []int32{s.DbGameFree.GetLimitCoin(), s.DbGameFree.GetMaxCoinLimit(), s.DbGameFree.GetServiceFee(),
			s.DbGameFree.GetBanker(), s.DbGameFree.GetBaseScore(), s.DbGameFree.GetSceneType(),
			s.DbGameFree.GetLowerThanKick(), s.DbGameFree.GetBetLimit()},
	}
	for _, value := range s.DbGameFree.GetMaxBetCoin() {
		pack.Params = append(pack.Params, value)
	}
	if sceneEx.Banker != nil {
		pack.Banker = &rollcoin.RollCoinPlayer{
			SnId: proto.Int32(sceneEx.Banker.SnId),
			Name: proto.String(sceneEx.Banker.Name),
			Sex:  proto.Int32(sceneEx.Banker.Sex),
			Head: proto.Int32(sceneEx.Banker.Head),
			Coin: proto.Int64(sceneEx.Banker.Coin),
			//			Params:      sceneEx.Banker.Params,
			Flag:        proto.Int(sceneEx.Banker.GetFlag()),
			BankerTimes: proto.Int32(sceneEx.Banker.BankerTimes),
			HeadOutLine: proto.Int32(sceneEx.Banker.HeadOutLine),
			VIP:         proto.Int32(sceneEx.Banker.VIP),
			City:        proto.String(sceneEx.Banker.GetCity()),
		}
	}
	pack.Player = &rollcoin.RollCoinPlayer{
		SnId: proto.Int32(playerEx.SnId),
		Name: proto.String(playerEx.Name),
		Sex:  proto.Int32(playerEx.Sex),
		Head: proto.Int32(playerEx.Head),
		Coin: proto.Int64(playerEx.GetCoin() - playerEx.betTotal),
		//Params:      proto.String(playerEx.Params),
		Flag:        proto.Int(playerEx.GetFlag()),
		BankerTimes: proto.Int32(playerEx.BankerTimes),
		HeadOutLine: proto.Int32(playerEx.HeadOutLine),
		VIP:         proto.Int32(playerEx.VIP),
		City:        proto.String(playerEx.GetCity()),
	}
	for _, id := range base.SystemChanceMgrEx.RollCoinIds {
		coinLog := &rollcoin.PushCoinLog{}
		if pool, ok := sceneEx.TotalBet[id]; ok {
			for coinType, coinNum := range pool {
				coinLog.Coin = append(coinLog.Coin, coinType)
				coinLog.Num = append(coinLog.Num, int32(coinNum))
			}
		}
		pack.CoinPool = append(pack.CoinPool, coinLog)
	}
	player := sceneEx.players[p.SnId]
	for _, id := range base.SystemChanceMgrEx.RollCoinIds {
		coinLog := &rollcoin.PushCoinLog{}
		if pool, ok := player.PushCoinLog[id]; ok {
			for coinType, coinNum := range pool {
				coinLog.Coin = append(coinLog.Coin, coinType)
				coinLog.Num = append(coinLog.Num, int32(coinNum))
			}
		}
		pack.PushLog = append(pack.PushLog, coinLog)
	}
	proto.SetDefaults(pack)
	p.SendToClient(int(rollcoin.PACKETID_ROLLCOIN_SC_ROOMINFO), pack)
	logger.Logger.Trace("SCRollCoinRoomInfo:", pack)
}
func (this *ScenePolicyRollCoin) NotifyGameState(s *base.Scene) {
	if s.GetState() > 2 {
		return
	}
	sec := int((RollCoinStartTimeout + RollCoinRollTimeout + RollCoinCoinOverTimeout).Seconds())
	if sceneEx, ok := s.ExtraData.(*RollCoinSceneData); ok {
		s.SyncGameState(sec, len(sceneEx.BankerList))
	}
}

//////////////////////////////////////////////////////////////
//状态基类
//////////////////////////////////////////////////////////////
type SceneBaseStateRollCoin struct {
}

func (this *SceneBaseStateRollCoin) GetTimeout(s *base.Scene) int {
	if sceneEx, ok := s.ExtraData.(*RollCoinSceneData); ok {
		return int(time.Now().Sub(sceneEx.StateStartTime) / time.Second)
	}
	return 0
}
func (this *SceneBaseStateRollCoin) CanChangeTo(s base.SceneState) bool {
	return true
}
func (this *SceneBaseStateRollCoin) OnEnter(s *base.Scene) {
	if sceneEx, ok := s.ExtraData.(*RollCoinSceneData); ok {
		sceneEx.StateStartTime = time.Now()
	}
}
func (this *SceneBaseStateRollCoin) OnLeave(s *base.Scene) {}
func (this *SceneBaseStateRollCoin) OnTick(s *base.Scene)  {}
func (this *SceneBaseStateRollCoin) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if sceneEx, ok := s.ExtraData.(*RollCoinSceneData); ok {
		if playerEx, ok := p.ExtraData.(*RollCoinPlayerData); ok {
			switch opcode {
			case RollCoinPlayerOpCurDownBanker:
				pack := &rollcoin.SCRollCoinOp{
					OpCode: proto.Int(opcode),
					Params: params,
					SnId:   proto.Int32(p.SnId),
				}
				if sceneEx.downBanker {
					pack.OpRetCode = rollcoin.OpResultCode_OPRC_Error
				} else {
					sceneEx.downBanker = true
					pack.OpRetCode = rollcoin.OpResultCode_OPRC_Sucess
				}
				proto.SetDefaults(pack)
				//logger.Logger.Trace("-------------------发送玩家操作消息")
				p.SendToClient(int(rollcoin.PACKETID_ROLLCOIN_SC_OP), pack)
			case RollCoinPlayerOpBanker:
				if sceneEx.DbGameFree.GetBanker() == 0 {
					return true
				}
				if sceneEx.Banker != nil && sceneEx.Banker.SnId == p.SnId {
					return true
				}
				if playerEx.Coin < int64(sceneEx.BankCoinLimit) {
					logger.Logger.Tracef("Player %v coin is less banker coin limit.", playerEx.GetName())
					return true
				}
				if common.InSliceInt32(sceneEx.BankerList, playerEx.SnId) {
					logger.Logger.Trace("Player %v is already in banker list.", playerEx.GetName())
					return true
				}
				sceneEx.BankerList = append(sceneEx.BankerList, playerEx.SnId)
				logger.Logger.Trace("BankerList:", sceneEx.BankerList)
				pack := &rollcoin.SCRollCoinBankerList{
					Insert: proto.Bool(true),
				}
				for _, value := range sceneEx.BankerList {
					banker := sceneEx.players[value]
					if banker == nil {
						continue
					}
					pack.List = append(pack.List, &rollcoin.RollCoinPlayer{
						SnId: proto.Int32(banker.SnId),
						Name: proto.String(banker.Name),
						Sex:  proto.Int32(banker.Sex),
						Head: proto.Int32(banker.Head),
						Coin: proto.Int64(banker.Coin - banker.betTotal),
						//Params: proto.String(banker.Params),
						HeadOutLine: proto.Int32(banker.HeadOutLine),
						VIP:         proto.Int32(banker.VIP),
						City:        proto.String(banker.GetCity()),
					})
				}
				pack.Count = proto.Int(len(pack.List))
				proto.SetDefaults(pack)
				sceneEx.Broadcast(int(rollcoin.PACKETID_ROLLCOIN_SC_BANKERLIST), pack, 0)
				return true
			case RollCoinPlayerOpDownBanker:
				if !common.InSliceInt32(sceneEx.BankerList, playerEx.SnId) {
					logger.Logger.Trace("Player %v is not already in banker list.", playerEx.GetName())
					return true
				}

				if len(sceneEx.BankerList) > 0 {
					index := 0
					for _, v := range sceneEx.BankerList {
						if v == playerEx.SnId {
							sceneEx.BankerList = append(sceneEx.BankerList[:index], sceneEx.BankerList[index+1:]...)
							break
						}
						index++
					}
				}
				//sceneEx.BankerList = append(sceneEx.BankerList, playerEx.SnId)
				logger.Logger.Trace("Down BankerList:", sceneEx.BankerList)
				pack := &rollcoin.SCRollCoinBankerList{
					Insert: proto.Bool(false),
				}
				for _, value := range sceneEx.BankerList {
					banker := sceneEx.players[value]
					if banker == nil {
						continue
					}
					pack.List = append(pack.List, &rollcoin.RollCoinPlayer{
						SnId: proto.Int32(banker.SnId),
						Name: proto.String(banker.Name),
						Sex:  proto.Int32(banker.Sex),
						Head: proto.Int32(banker.Head),
						Coin: proto.Int64(banker.Coin - banker.betTotal),
						//Params: proto.String(banker.Params),
						HeadOutLine: proto.Int32(banker.HeadOutLine),
						VIP:         proto.Int32(banker.VIP),
						City:        proto.String(banker.GetCity()),
					})
				}
				pack.Count = proto.Int(len(pack.List))
				proto.SetDefaults(pack)
				sceneEx.Broadcast(int(rollcoin.PACKETID_ROLLCOIN_SC_BANKERLIST), pack, 0)
				return true
			case RollCoinPlayerOpBankerList:
				pack := &rollcoin.SCRollCoinBankerList{}

				if common.InSliceInt32(sceneEx.BankerList, playerEx.SnId) {
					pack.Insert = proto.Bool(true)
				} else {
					pack.Insert = proto.Bool(false)
				}
				for _, value := range sceneEx.BankerList {
					banker := sceneEx.players[value]
					if banker == nil {
						continue
					}
					pack.List = append(pack.List, &rollcoin.RollCoinPlayer{
						SnId: proto.Int32(banker.SnId),
						Name: proto.String(banker.Name),
						Sex:  proto.Int32(banker.Sex),
						Head: proto.Int32(banker.Head),
						Coin: proto.Int64(banker.Coin - banker.betTotal),
						//						Params: banker.Params,
						HeadOutLine: proto.Int32(banker.HeadOutLine),
						VIP:         proto.Int32(banker.VIP),
						City:        proto.String(banker.GetCity()),
					})
				}

				playerEx.SendToClient(int(rollcoin.PACKETID_ROLLCOIN_SC_BANKERLIST), pack)
				logger.Logger.Trace("SCRollCoinBankerList:", pack)
				return true
			case RollCoinPlayerList:
				//清空
				for i := len(sceneEx.seats) - 1; i >= 0; i-- {
					sceneEx.seats = append(sceneEx.seats[:i], sceneEx.seats[i+1:]...)
				}
				for _, sp := range sceneEx.players {
					sceneEx.seats = append(sceneEx.seats, sp)
				}

				packList := &rollcoin.SCRollCoinPlayerList{
					PlayerNum: proto.Int32(int32(sceneEx.CountPlayer())),
				}
				sort.Sort(&RollCoinSceneData{seats: sceneEx.seats, by: func(p, q *RollCoinPlayerData) bool {
					if q.cGetBetGig20 == p.cGetBetGig20 {
						return q.Coin-q.betTotal > p.Coin-p.betTotal
					} else {
						return q.cGetBetGig20 > p.cGetBetGig20
					}
				}})
				//近20局押注胜率最高的玩家
				var winTimesBig *RollCoinPlayerData
				winTimesBig = nil
				for _, v := range sceneEx.seats {
					if v.IsMarkFlag(base.PlayerState_GameBreak) { //观众
						continue
					}
					if v.cGetWin20 <= 0 { //近20局胜率不能为0
						continue
					}
					if sceneEx.Banker != nil {
						if sceneEx.Banker.SnId == v.SnId {
							continue
						}
					}

					if winTimesBig == nil {
						winTimesBig = v
					} else {
						if winTimesBig.cGetWin20 < v.cGetWin20 {
							winTimesBig = v
						}
					}
				}

				if winTimesBig != nil {
					winTimesBig.MarkFlag(base.PlayerState_Check) //标记为已经放到玩家列表，下面的不再加入
					curcoin := winTimesBig.Coin
					pd := &rollcoin.RollCoinPlayer{
						SnId:        proto.Int32(winTimesBig.SnId),
						Name:        proto.String(winTimesBig.Name),
						Head:        proto.Int32(winTimesBig.Head),
						Sex:         proto.Int32(winTimesBig.Sex),
						Coin:        proto.Int64(curcoin - winTimesBig.betTotal),
						Flag:        proto.Int(winTimesBig.GetFlag()),
						Lately20Bet: proto.Int32(int32(winTimesBig.cGetBetGig20)),
						Lately20Win: proto.Int32(int32(winTimesBig.cGetWin20)),
						HeadOutLine: proto.Int32(winTimesBig.HeadOutLine),
						VIP:         proto.Int32(winTimesBig.VIP),
						City:        proto.String(winTimesBig.GetCity()),
					}
					packList.Data = append(packList.Data, pd)
				} else {

				}

				i20 := 0 //只显示20个玩家信息
				for _, pp := range sceneEx.seats {
					if pp.IsMarkFlag(base.PlayerState_GameBreak) { //观众
						continue
					}
					if sceneEx.Banker != nil {
						if sceneEx.Banker.SnId == pp.SnId {
							continue
						}
					}
					if !pp.IsMarkFlag(base.PlayerState_Check) {
						i20++
						if i20 > 20 {
							break
						}
						curcoin := pp.Coin
						pd := &rollcoin.RollCoinPlayer{
							SnId:        proto.Int32(pp.SnId),
							Name:        proto.String(pp.Name),
							Head:        proto.Int32(pp.Head),
							Sex:         proto.Int32(pp.Sex),
							Coin:        proto.Int64(curcoin - pp.betTotal),
							Flag:        proto.Int(pp.GetFlag()),
							Lately20Bet: proto.Int32(int32(pp.cGetBetGig20)),
							Lately20Win: proto.Int32(int32(pp.cGetWin20)),
							HeadOutLine: proto.Int32(pp.HeadOutLine),
							VIP:         proto.Int32(pp.VIP),
							City:        proto.String(pp.GetCity()),
						}
						packList.Data = append(packList.Data, pd)
					} else {
						pp.UnmarkFlag(base.PlayerState_Check)
					}
				}
				proto.SetDefaults(packList)
				//logger.Logger.Trace("-------------------发送玩家列表消息")
				p.SendToClient(int(rollcoin.PACKETID_ROLLCOIN_SC_PLAYERLIST), packList)
				return true
			}
		}
	}
	return false
}

func (this *SceneBaseStateRollCoin) OnPlayerEvent(s *base.Scene, p *base.Player, evtcode int, params []int64) {
	if sceneEx, ok := s.ExtraData.(*RollCoinSceneData); ok {
		if playerEx, ok := p.ExtraData.(*RollCoinPlayerData); ok {
			switch evtcode {
			case base.PlayerEventRecharge:
				//oldflag := p.MarkBroadcast(false)
				p.AddCoin(params[0], common.GainWay_Pay, base.SyncFlag_ToClient, "system", p.GetScene().GetSceneName())
				//p.MarkBroadcast(oldflag)
				if p.Coin >= int64(sceneEx.CoinLimit) && p.Coin >= int64(sceneEx.Rollcoin[0]) {
					playerEx.UnmarkFlag(base.PlayerState_GameBreak)
					playerEx.SyncFlag(true)
				}
			}
		}
	}
}

//////////////////////////////////////////////////////////////
//等待状态
//////////////////////////////////////////////////////////////
type SceneRollCoinStateWait struct {
	SceneBaseStateRollCoin
}

func (this *SceneRollCoinStateWait) GetState() int { return RollCoinSceneStateWait }
func (this *SceneRollCoinStateWait) CanChangeCoinScene(s *base.Scene, p *base.Player) bool {
	return true
}
func (this *SceneRollCoinStateWait) CanChangeTo(s base.SceneState) bool {
	switch s.GetState() {
	case RollCoinSceneStateStart:
		return true
	default:
		return false
	}
}
func (this *SceneRollCoinStateWait) OnEnter(s *base.Scene) {
	logger.Logger.Trace("SceneWaitStateRollCoin::OnEnter")
	this.SceneBaseStateRollCoin.OnEnter(s)
	if sceneEx, ok := s.ExtraData.(*RollCoinSceneData); ok {
		//上庄列表中的玩家金币数量不足上庄条件，会自动取消上庄
		sceneEx.BankerListDelCheck()
		if sceneEx.Banker != nil { //连庄5轮的玩家，将会在结算时自动下庄
			if sceneEx.Banker.IsRob {
				if sceneEx.Banker.BankerTimes > rand.Int31n(5) {
					sceneEx.downBanker = true
				}
			}
			if sceneEx.Banker.BankerTimes > 5 || sceneEx.Banker.Coin < int64(sceneEx.BankCoinLimit) || sceneEx.downBanker {
				sceneEx.TryChangeBanker()
			}
			if sceneEx.downBanker {
				sceneEx.downBanker = false
			}
		} else { //没人坐庄，就在列表里找找
			sceneEx.TryChangeBanker()
		}
		sceneEx.SendRobotUpBankerList()
		pack := &rollcoin.SCRollCoinRoomState{
			State: proto.Int(this.GetState()),
		}
		if sceneEx.Banker != nil {
			pack.Params = append(pack.Params, sceneEx.Banker.BankerTimes)
		}
		proto.SetDefaults(pack)
		s.Broadcast(int(rollcoin.PACKETID_ROLLCOIN_SC_ROOMSTATE), pack, 0)
	}
}
func (this *SceneRollCoinStateWait) OnTick(s *base.Scene) {
	this.SceneBaseStateRollCoin.OnTick(s)
	if sceneEx, ok := s.ExtraData.(*RollCoinSceneData); ok {
		if time.Now().Sub(sceneEx.StateStartTime) > RollCoinWaiteTimeout {
			s.ChangeSceneState(RollCoinSceneStateStart)
		}
	}
}
func (this *SceneRollCoinStateWait) OnLeave(s *base.Scene) {
	logger.Logger.Trace("SceneWaitStateRollCoin::OnLeave")
	if sceneEx, ok := s.ExtraData.(*RollCoinSceneData); ok {
		sceneEx.GameStartTime = time.Now()
	}
}
func (this *SceneRollCoinStateWait) OnPlayerEvent(s *base.Scene, p *base.Player, evtcode int, params []int64) {
	this.SceneBaseStateRollCoin.OnPlayerEvent(s, p, evtcode, params)
}

//////////////////////////////////////////////////////////////
//开始状态
//////////////////////////////////////////////////////////////
type SceneRollCoinStateStart struct {
	SceneBaseStateRollCoin
}

func (this *SceneRollCoinStateStart) GetState() int { return RollCoinSceneStateStart }
func (this *SceneRollCoinStateStart) CanChangeCoinScene(s *base.Scene, p *base.Player) bool {
	return true
}
func (this *SceneRollCoinStateStart) CanChangeTo(s base.SceneState) bool {
	switch s.GetState() {
	case RollCoinSceneStateRoll:
		return true
	}
	return false
}

func (this *SceneRollCoinStateStart) OnEnter(s *base.Scene) {
	logger.Logger.Trace("SceneStartStateRollCoin::OnEnter")
	this.SceneBaseStateRollCoin.OnEnter(s)
	if sceneEx, ok := s.ExtraData.(*RollCoinSceneData); ok {
		sceneEx.Clean()
		sceneEx.NumOfGames++
		sceneEx.GameNowTime = time.Now()
		//通知world开始第几局
		//		if !s.IsMatchScene() && !s.IsCoinScene() {
		//			if sceneEx.numOfGames > 1 {
		//				s.NotifySceneRoundStart(sceneEx.numOfGames)
		//			}
		//		}
		pack := &rollcoin.SCRollCoinRoomState{
			State: proto.Int(this.GetState()),
		}
		proto.SetDefaults(pack)
		s.Broadcast(int(rollcoin.PACKETID_ROLLCOIN_SC_ROOMSTATE), pack, 0)
	}
}
func (this *SceneRollCoinStateStart) OnPlayerEvent(s *base.Scene, p *base.Player, evtcode int, params []int64) {
	this.SceneBaseStateRollCoin.OnPlayerEvent(s, p, evtcode, params)
}
func (this *SceneRollCoinStateStart) OnTick(s *base.Scene) {
	this.SceneBaseStateRollCoin.OnTick(s)
	if sceneEx, ok := s.ExtraData.(*RollCoinSceneData); ok {
		if time.Now().Sub(sceneEx.StateStartTime) > RollCoinStartTimeout {
			s.ChangeSceneState(RollCoinSceneStateRoll)
		}
	}
}

//////////////////////////////////////////////////////////////
//押注
//////////////////////////////////////////////////////////////
type SceneRollCoinStateRoll struct {
	SceneBaseStateRollCoin
}

func (this *SceneRollCoinStateRoll) GetState() int { return RollCoinSceneStateRoll }
func (this *SceneRollCoinStateRoll) CanChangeCoinScene(s *base.Scene, p *base.Player) bool {
	return true
}
func (this *SceneRollCoinStateRoll) CanChangeTo(s base.SceneState) bool {
	switch s.GetState() {
	case RollCoinSceneStateCoinOver:
		return true
	}
	return false
}
func (this *SceneRollCoinStateRoll) OnEnter(s *base.Scene) {
	logger.Logger.Trace("SceneRollStateRollCoin::OnEnter")
	this.SceneBaseStateRollCoin.OnEnter(s)
	if sceneEx, ok := s.ExtraData.(*RollCoinSceneData); ok {
		pack := &rollcoin.SCRollCoinRoomState{
			State: proto.Int(this.GetState()),
		}
		proto.SetDefaults(pack)
		s.Broadcast(int(rollcoin.PACKETID_ROLLCOIN_SC_ROOMSTATE), pack, 0)
		sceneEx.SyncTime = time.Now()
		for key, pool := range sceneEx.SyncData {
			for coin, _ := range pool {
				sceneEx.SyncData[key][coin] = 0
			}
		}
	}
}
func (this *SceneRollCoinStateRoll) OnTick(s *base.Scene) {
	this.SceneBaseStateRollCoin.OnTick(s)
	if sceneEx, ok := s.ExtraData.(*RollCoinSceneData); ok {
		if time.Now().Sub(sceneEx.StateStartTime) > RollCoinRollTimeout {
			s.ChangeSceneState(RollCoinSceneStateCoinOver)
		}
		if time.Now().Sub(sceneEx.SyncTime) >= time.Second*1 {
			sceneEx.SyncCoinLog()
		}
	}
}
func (this *SceneRollCoinStateRoll) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if this.SceneBaseStateRollCoin.OnPlayerOp(s, p, opcode, params) {
		return true
	}
	if sceneEx, ok := s.ExtraData.(*RollCoinSceneData); ok {
		if playerEx, ok := p.ExtraData.(*RollCoinPlayerData); ok {
			switch opcode {
			case RollCoinPlayerOpPushCoin:
				if len(params) < 2 {
					logger.Logger.Errorf("Player %v push coin param error:%v", playerEx.SnId, params)
					return false
				}
				if playerEx.IsMarkFlag(base.PlayerState_GameBreak) {
					return false
				}
				if playerEx == sceneEx.Banker {
					return false
				}
				flag := int32(params[0])
				coin := params[1]
				if coin < 0 {
					return false
				}
				if !common.InSliceInt32(base.SystemChanceMgrEx.RollCoinIds, flag) {
					logger.Logger.Error("Recive client error flag:", params, "in push coin op.")
					return false
				}
				if !common.InSliceInt32(sceneEx.Rollcoin, int32(coin)) {
					logger.Logger.Error("Recive client error coin:", params, "in push coin op.")
					return false
				}
				//最小面额的筹码
				minChip := int64(sceneEx.Rollcoin[0])
				//期望下注金额
				expectBetCoin := coin
				//最终下注金额
				lastBetCoin := int64(0)
				logger.Logger.Tracef("Player %v push %v coin in %v pool.", playerEx.GetName(), coin, flag)
				//庄家为玩家时，单个投注区域的投注上限=庄家剩余金币数/投注区倍率
				if sceneEx.Banker != nil {
					//检查当前下注区域是否还能下注
					if sceneEx.ZoneBetIsFull[flag] {
						pack := &rollcoin.SCRollCoinPushCoin{
							OpRetCode: rollcoin.OpResultCode_OPRC_RollCoin_PushCoinTooMuch,
						}
						proto.SetDefaults(pack)
						playerEx.SendToClient(int(rollcoin.PACKETID_ROLLCOIN_SC_PUSHCOIN), pack)
						return true
					}
					bankerCoinLimit := (sceneEx.Banker.Coin) / int64(base.SystemChanceMgrEx.RollCoinRate[flag])
					zoneBetTotal := sceneEx.ZoneTotal[flag]
					if zoneBetTotal+coin > bankerCoinLimit {
						//计算还能下注多少
						lastCanBetCoin := bankerCoinLimit - zoneBetTotal
						lastCanBetCoin /= 100
						lastCanBetCoin *= 100
						if lastCanBetCoin <= 0 {
							logger.Logger.Trace("Player %v push much coin to %v.", playerEx.SnId, flag)
							pack := &rollcoin.SCRollCoinPushCoin{
								OpRetCode: rollcoin.OpResultCode_OPRC_RollCoin_PushCoinTooMuch,
							}
							proto.SetDefaults(pack)
							playerEx.SendToClient(int(rollcoin.PACKETID_ROLLCOIN_SC_PUSHCOIN), pack)
							return true
						} else {
							lastBetCoin = lastCanBetCoin
						}
						sceneEx.ZoneBetIsFull[flag] = true
					} else {
						lastBetCoin = expectBetCoin
					}
				} else {
					lastBetCoin = expectBetCoin
				}
				//押注的金币不能超过拥有的金币
				if lastBetCoin > (playerEx.Coin - playerEx.betTotal) {
					return true
				}
				//押注限制(低于该值不能押注)
				if p.Coin < int64(sceneEx.DbGameFree.GetBetLimit()) && playerEx.betTotal == 0 {
					logger.Logger.Tracef("======提示低于多少不能下注====== snid:%v coin:%v betTotal:%v BetLimit:%v", p.SnId, p.Coin, playerEx.betTotal, sceneEx.DbGameFree.GetBetLimit())
					msgPack := common.CreateSrvMsg(common.SRVMSG_CODE_DEFAULT, fmt.Sprintf("抱歉，%.2f金币以上才能下注", float64(sceneEx.DbGameFree.GetBetLimit())/100))
					p.SendToClient(int(player.PlayerPacketID_PACKET_SC_SRVMSG), msgPack)
					return true
				}
				//单门押注上限
				maxBetCoin := sceneEx.DbGameFree.GetMaxBetCoin()
				if lastBetCoin == expectBetCoin {
					if ok, coinLimit := playerEx.MaxChipCheck(int64(flag), lastBetCoin, maxBetCoin); !ok {
						lastBetCoin = int64(coinLimit - playerEx.CoinPool[flag])
						//对齐到最小面额的筹码
						lastBetCoin /= minChip
						lastBetCoin *= minChip
						if lastBetCoin <= 0 {
							msgPack := common.CreateSrvMsg(common.SRVMSG_CODE_DEFAULT, fmt.Sprintf("该门押注金额上限%.2f", float64(coinLimit)/100))
							p.SendToClient(int(player.PlayerPacketID_PACKET_SC_SRVMSG), msgPack)
							return true
						}
					}
				}
				playerEx.Trusteeship = 0
				if !playerEx.gameing {
					playerEx.gameing = true
				}
				if playerEx.GetTotlePushCoin() == 0 {
					playerEx.SetCurrentCoin(playerEx.Coin)
				}
				if lastBetCoin == expectBetCoin {
					playerEx.PushCoin(flag, int32(lastBetCoin))
					sceneEx.CacheCoinLog(flag, int32(lastBetCoin), playerEx.IsRob)
				} else {
					val := lastBetCoin
					extChip := []int32{100, 200, 500, 1000, 5000}
					lastChips := make([]int32, 0)
					for _, chipValue := range extChip {
						if chipValue < sceneEx.Rollcoin[0] {
							lastChips = append(lastChips, chipValue)
						} else {
							break
						}
					}
					lastChips = append(lastChips, sceneEx.Rollcoin...)
					for i := len(lastChips) - 1; i >= 0 && val > 0; i-- {
						chip := int64(lastChips[i])
						cntChip := val / chip
						for j := int64(0); j < cntChip; j++ {
							playerEx.PushCoin(flag, int32(chip))
							sceneEx.CacheCoinLog(flag, int32(chip), playerEx.IsRob)
						}
						val -= cntChip * chip
					}
				}
				//累积总投注额
				playerEx.TotalBet += lastBetCoin
				playerEx.betTotal += lastBetCoin
				//oldflag := playerEx.MarkBroadcast(false)
				//playerEx.BetCoinChange(-lastBetCoin, common.GainWay_Game, false, "system", s.GetSceneName())
				//playerEx.MarkBroadcast(oldflag)

				pack := &rollcoin.SCRollCoinPushCoin{
					OpRetCode: rollcoin.OpResultCode_OPRC_Sucess,
					Coin:      proto.Int32(int32(lastBetCoin)),
					Flag:      proto.Int32(flag),
					MeTotle:   proto.Int64(int64(playerEx.GetTotlePushCoin())),
					SnId:      proto.Int32(playerEx.SnId),
				}
				if sceneEx.godid == playerEx.SnId {
					pack.PlayerType = proto.Int32(1)
					proto.SetDefaults(pack)
					sceneEx.Broadcast(int(rollcoin.PACKETID_ROLLCOIN_SC_PUSHCOIN), pack, 0) //广播给其他用户
				} else {
					pack.PlayerType = proto.Int32(0)
					proto.SetDefaults(pack)
					playerEx.SendToClient(int(rollcoin.PACKETID_ROLLCOIN_SC_PUSHCOIN), pack)
				}
				//每个下注区域提前达到庄的额度上限，直接开奖
				if sceneEx.CheckBetIsFull() {
					msgPack := common.CreateSrvMsg(common.SRVMSG_CODE_DEFAULT, "庄家可押注额度满，直接开牌")
					s.Broadcast(int(player.PlayerPacketID_PACKET_SC_SRVMSG), msgPack, -1)
					//切换到下一个阶段
					s.ChangeSceneState(RollCoinSceneStateCoinOver)
				}
				logger.Logger.Trace("SCRollCoinPushCoin:", pack)
			}
		}
	}
	return true
}

func (this *SceneRollCoinStateRoll) OnPlayerEvent(s *base.Scene, p *base.Player, evtcode int, params []int64) {
	this.SceneBaseStateRollCoin.OnPlayerEvent(s, p, evtcode, params)
}
func (this *SceneRollCoinStateRoll) OnLeave(s *base.Scene) {
	logger.Logger.Trace("SceneCoinStateRollCoin::OnLeave")
	if sceneEx, ok := s.ExtraData.(*RollCoinSceneData); ok {
		sceneEx.SyncCoinLog()
		/*betCoin := int64(0)
		for _, id := range base.SystemChanceMgrEx.RollCoinIds {
			betCoin += sceneEx.ZoneTotal[id] - sceneEx.ZoneRobTotal[id]
		}
		base.CoinPoolMgr.PushCoin(sceneEx.gamefreeId, sceneEx.platform, int64(betCoin))*/
	}
}

//////////////////////////////////////////////////////////////
//押注结束状态
//////////////////////////////////////////////////////////////
type SceneRollCoinStateOver struct {
	SceneBaseStateRollCoin
}

func (this *SceneRollCoinStateOver) GetState() int { return RollCoinSceneStateCoinOver }
func (this *SceneRollCoinStateOver) CanChangeCoinScene(s *base.Scene, p *base.Player) bool {
	return true
}
func (this *SceneRollCoinStateOver) CanChangeTo(s base.SceneState) bool {
	switch s.GetState() {
	case RollCoinSceneStateGame:
		return true
	default:
		return false
	}
}

func (this *SceneRollCoinStateOver) OnEnter(s *base.Scene) {
	this.SceneBaseStateRollCoin.OnEnter(s)
	if sceneEx, ok := s.ExtraData.(*RollCoinSceneData); ok {
		for _, v := range sceneEx.players {
			if v != nil && !v.IsRob {
				if sceneEx.Banker != nil && sceneEx.Banker == v {
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
	pack := &rollcoin.SCRollCoinRoomState{
		State: proto.Int(this.GetState()),
	}
	proto.SetDefaults(pack)
	s.Broadcast(int(rollcoin.PACKETID_ROLLCOIN_SC_ROOMSTATE), pack, 0)
}
func (this *SceneRollCoinStateOver) OnTick(s *base.Scene) {
	this.SceneBaseStateRollCoin.OnTick(s)
	if sceneEx, ok := s.ExtraData.(*RollCoinSceneData); ok {
		if time.Now().Sub(sceneEx.StateStartTime) > RollCoinCoinOverTimeout {
			s.ChangeSceneState(RollCoinSceneStateGame)
		}
	}
}
func (this *SceneRollCoinStateOver) OnPlayerEvent(s *base.Scene, p *base.Player, evtcode int, params []int64) {
	this.SceneBaseStateRollCoin.OnPlayerEvent(s, p, evtcode, params)
	switch evtcode {
	case base.PlayerEventEnter:
		fallthrough
	case base.PlayerEventRehold:
		if playerEx, ok := p.ExtraData.(*RollCoinPlayerData); ok {
			playerEx.MarkFlag(base.PlayerState_WaitNext)
			playerEx.SyncFlag(true)
		}
	}
}

//////////////////////////////////////////////////////////////
//游戏状态
//////////////////////////////////////////////////////////////
type SceneRollCoinStateGame struct {
	SceneBaseStateRollCoin
}

func (this *SceneRollCoinStateGame) GetState() int { return RollCoinSceneStateGame }
func (this *SceneRollCoinStateGame) CanChangeCoinScene(s *base.Scene, p *base.Player) bool {
	return true
}
func (this *SceneRollCoinStateGame) CanChangeTo(s base.SceneState) bool {
	switch s.GetState() {
	case RollCoinSceneStateBilled:
		return true
	}
	return false
}
func (this *SceneRollCoinStateGame) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if this.SceneBaseStateRollCoin.OnPlayerOp(s, p, opcode, params) {
		return true
	}
	return false
}
func (this *SceneRollCoinStateGame) OnEnter(s *base.Scene) {
	logger.Logger.Trace("SceneGameStateRollCoin::OnEnter")
	this.SceneBaseStateRollCoin.OnEnter(s)
	if sceneEx, ok := s.ExtraData.(*RollCoinSceneData); ok {
		//计算赢家位置
		winFlag := sceneEx.GetWinFlg()
		winpos := sceneEx.ChangeCard(int(winFlag))
		logger.Logger.Infof("start win pos%v,end pos %v", winFlag, winpos)
		if winpos != -1 {
			winFlag = int32(winpos)
		}
		//计算系统的金币产出
		/*sysOut := sceneEx.CalcuSystemCoinOut(winFlag)
		if !base.CoinPoolMgr.IsCoinEnough(sceneEx.gamefreeId, sceneEx.groupId, sceneEx.platform, int64(sysOut)) {
			//金币池中的金币不够扣除了
			var rollCoinIds = base.SystemChanceMgrEx.GetRollCoinIds() //避免1个以上的区域，金币产出为0时，固定到某个位置赢
			//尝试在所有的选择中，挑出来，金币产出最少的一个位置
			var minOut = int32(math.MaxInt32)
			var minId = int32(-1)
			for _, id := range rollCoinIds {
				sysOut = sceneEx.CalcuSystemCoinOut(id)
				if sysOut < minOut {
					minOut = sysOut
					minId = id
				}
			}
			winFlag = minId
		}*/
		if winFlag == -1 {
			winFlag = sceneEx.GetWinFlg()
			logger.Logger.Error("Roll coin system win flag is error.", winFlag)
		}
		winIndex := -1
		listLen := len(sceneEx.RollCoinTargetList)
		startIndex := rand.Intn(listLen)
		for i := startIndex; i < 100; i++ {
			if sceneEx.RollCoinTargetList[i%listLen] == winFlag {
				winIndex = i % listLen
				break
			}
		}
		sceneEx.WinFlg = winFlag
		sceneEx.WinIndex = int32(winIndex)
		logger.Logger.Tracef("RollCoinSceneData %v-%v pos is winer.", winFlag, winIndex)
		pack := &rollcoin.SCRollCoinRoomState{
			State:  proto.Int(this.GetState()),
			Params: []int32{int32(winFlag), int32(winIndex)},
		}
		proto.SetDefaults(pack)
		s.Broadcast(int(rollcoin.PACKETID_ROLLCOIN_SC_ROOMSTATE), pack, 0)
	}
}

func (this *SceneRollCoinStateGame) OnLeave(s *base.Scene) {
	this.SceneBaseStateRollCoin.OnLeave(s)
	if sceneEx, ok := s.ExtraData.(*RollCoinSceneData); ok {
		pack := &server.GWGameStateLog{
			SceneId: proto.Int(s.SceneId),
			GameLog: proto.Int32(sceneEx.WinFlg),
		}
		s.SendToWorld(int(server.SSPacketID_PACKET_GW_GAMESTATELOG), pack)
	}
}

func (this *SceneRollCoinStateGame) OnTick(s *base.Scene) {
	this.SceneBaseStateRollCoin.OnTick(s)
	if sceneEx, ok := s.ExtraData.(*RollCoinSceneData); ok {
		if time.Now().Sub(sceneEx.StateStartTime) > RollCoinGameTimeOut {
			s.ChangeSceneState(RollCoinSceneStateBilled)
		}
	}
}
func (this *SceneRollCoinStateGame) OnPlayerEvent(s *base.Scene, p *base.Player, evtcode int, params []int64) {
	this.SceneBaseStateRollCoin.OnPlayerEvent(s, p, evtcode, params)
	switch evtcode {
	case base.PlayerEventEnter:
		fallthrough
	case base.PlayerEventRehold:
		if playerEx, ok := p.ExtraData.(*RollCoinPlayerData); ok {
			playerEx.MarkFlag(base.PlayerState_WaitNext)
			playerEx.SyncFlag(true)
		}
	}
}

//////////////////////////////////////////////////////////////
//结算状态
//////////////////////////////////////////////////////////////
type SceneRollCoinStateBilled struct {
	SceneBaseStateRollCoin
}

func (this *SceneRollCoinStateBilled) GetState() int {
	return RollCoinSceneStateBilled
}

func (this *SceneRollCoinStateBilled) CanChangeTo(s base.SceneState) bool {
	switch s.GetState() {
	case RollCoinSceneStateWait:
		return true
	}
	return false
}

//当前状态能否换桌
func (this *SceneRollCoinStateBilled) CanChangeCoinScene(s *base.Scene, p *base.Player) bool {
	return true
}

func (this *SceneRollCoinStateBilled) OnEnter(s *base.Scene) {
	logger.Logger.Trace("SceneBilledStateRollCoin::OnEnter")
	this.SceneBaseStateRollCoin.OnEnter(s)
	if sceneEx, ok := s.ExtraData.(*RollCoinSceneData); ok {
		//计算赢家位置
		winFlag := sceneEx.WinFlg
		winIndex := sceneEx.WinIndex
		sceneEx.LogWinFlag()
		//发送状态消息
		pack := &rollcoin.SCRollCoinRoomState{
			State:  proto.Int(this.GetState()),
			Params: []int32{int32(winFlag), int32(winIndex)},
		}
		proto.SetDefaults(pack)
		s.Broadcast(int(rollcoin.PACKETID_ROLLCOIN_SC_ROOMSTATE), pack, 0)

		// 水池上下文环境
		sceneEx.CpCtx = base.CoinPoolMgr.GetCoinPoolCtx(sceneEx.Platform, sceneEx.GetGameFreeId(), sceneEx.GroupId)
		sceneEx.CpCtx.Controlled = sceneEx.CpControlled

		//计算积分，结算金币
		bankerWinCoin := int32(0)  //统计庄家赢取的金币
		bankerLoseCoin := int32(0) //统计庄家输的金币
		////////////////////////////////////////////////////
		rollhundredtype := make([]model.RollHundredType, 8, 8)
		for i := 0; i < 8; i++ {
			rollhundredtype[i].RegionId = int32(i)
			rollhundredtype[i].Rate = int(base.SystemChanceMgrEx.RollCoinRate[i])
			rollhundredtype[i].IsWin = -1
		}
		rollhundredtype[winFlag].IsWin = 1
		//TODO 这里的庄家判断可以不是真实玩家
		if sceneEx.Banker != nil { //玩家坐庄
			//临时增加一个处理，保存庄的当前金币
			sceneEx.Banker.SetCurrentCoin(sceneEx.Banker.Coin)

			logger.Logger.Tracef("Player %v is banker.", sceneEx.Banker.GetName())
			nRobotWinCoin := int32(0)  //统计用户赢取的金币
			nRobotLoseCoin := int32(0) //统计用户输的金币
			for _, player := range sceneEx.players {
				if player == nil {
					continue
				}
				if player.betTotal > 0 {
					player.AddCoin(-(player.betTotal), common.GainWay_HundredSceneLost, base.SyncFlag_ToClient, "system", s.GetSceneName())
				}
				if player.SnId == sceneEx.Banker.SnId {
					continue //不处理庄家的押注
				}
				loseCoin := int32(0) //统计玩家输的金币
				winCoin := int32(0)  //统计玩家赢取的金币
				for flag, pushCoin := range player.CoinPool {
					if pushCoin == 0 {
						continue
					}
					rp := model.RollPerson{}
					rp.UserId = player.SnId
					if flag == winFlag { //计算压中的位置，赢取的金币为押注金额*倍率
						winCoin = pushCoin * base.SystemChanceMgrEx.RollCoinRate[winFlag]
						rp.ChangeCoin = int64(winCoin) - int64(pushCoin)
						loseCoin += pushCoin
					} else { //未压中的位置，为输掉的金币，输给庄家
						loseCoin += pushCoin
						rp.ChangeCoin = int64(-pushCoin)
					}
					rp.IsRob = player.IsRob
					rp.WBLevel = player.WBLevel
					rp.BeforeCoin = player.GetCurrentCoin()
					rp.IsFirst = sceneEx.IsPlayerFirst(player.Player)
					rp.AfterCoin = player.Coin
					rp.UserBetTotal = int64(pushCoin)
					if !player.IsRob || !sceneEx.Banker.IsRob {
						rollhundredtype[flag].RollPerson = append(rollhundredtype[flag].RollPerson, rp)
					}
				}
				bankerWinCoin += loseCoin //庄家赢取的金币为玩家输掉的金币
				bankerLoseCoin += winCoin //庄家输掉的金币为玩家压中赢取的金币
				//统计
				if !player.IsRob {
					nRobotWinCoin += loseCoin
					nRobotLoseCoin += winCoin
				}
				player.gainCoin = winCoin // - loseCoin //玩家最终的金币
				logger.Logger.Tracef("%v win %v coin.", player.GetName(), winCoin)
				logger.Logger.Tracef("%v lose %v coin.", player.GetName(), loseCoin)
				var gainWay int32
				var tax int64
				if player.gainCoin > 0 {
					gainWay = common.GainWay_HundredSceneWin
					tax = int64(player.gainCoin) * int64(s.DbGameFree.GetTaxRate()) / 10000
					player.winCoin = int64(player.gainCoin)
					player.taxCoin = tax
					player.gainCoin -= int32(tax)
					player.AddServiceFee(tax)
				} else {
					gainWay = common.GainWay_HundredSceneLost
				}
				//oldflag := player.MarkBroadcast(false)
				player.AddCoin(int64(player.gainCoin), gainWay, 0, "system", s.GetSceneName())
				//player.MarkBroadcast(oldflag)
				betTotal := int64(player.betTotal)
				if betTotal != 0 {
					//上报游戏时间
					//player.ReportGameEvent(int64(player.gainCoin)-betTotal, tax, betTotal)
				}

				//统计游戏局数
				if winCoin != 0 || loseCoin != 0 {
					if winCoin-loseCoin > 0 {
						player.WinTimes++
					} else {
						player.FailTimes++
					}
				}
			}
			sceneEx.Banker.gainCoin = bankerWinCoin - bankerLoseCoin
			//			//统计
			//			if !sceneEx.Banker.IsRob {
			//				sceneEx.SummaryData.WinCoin += int64(bankerWinCoin - nRobotWinCoin)
			//				sceneEx.SummaryData.LoseCoin += int64(bankerLoseCoin - nRobotLoseCoin)
			//			} else {
			//				sceneEx.SummaryData.WinCoin += int64(nRobotWinCoin)
			//				sceneEx.SummaryData.LoseCoin += int64(nRobotLoseCoin)
			//			}
			logger.Logger.Tracef("Banker win %v coin.", bankerWinCoin)
			logger.Logger.Tracef("Banker lose %v coin.", bankerLoseCoin)
			sceneEx.Banker.BankerTimes++
			var gainWay int32
			tax := int64(0)
			if sceneEx.Banker.gainCoin > 0 {
				gainWay = common.GainWay_HundredSceneWin
				sceneEx.Banker.WinTimes++
				tax = int64(sceneEx.Banker.gainCoin) * int64(s.DbGameFree.GetTaxRate()) / 10000
				sceneEx.Banker.winCoin = int64(sceneEx.Banker.gainCoin)
				sceneEx.Banker.taxCoin = tax
				sceneEx.Banker.gainCoin -= int32(tax)
				sceneEx.Banker.AddServiceFee(tax)
			} else {
				gainWay = common.GainWay_HundredSceneLost
				sceneEx.Banker.FailTimes++
				sceneEx.Banker.winCoin = int64(sceneEx.Banker.gainCoin)
			}
			//oldflag := sceneEx.Banker.MarkBroadcast(true)
			sceneEx.Banker.AddCoin(int64(sceneEx.Banker.gainCoin), gainWay, 0, "system", s.GetSceneName())
			//sceneEx.Banker.MarkBroadcast(oldflag)
			//上报游戏事件
			//sceneEx.Banker.ReportGameEvent(int64(sceneEx.Banker.gainCoin), tax, 0)
		} else { //系统坐庄
			logger.Logger.Trace("System is banker.")
			for _, player := range sceneEx.players {
				if player == nil {
					continue
				}
				if player.betTotal > 0 {
					player.AddCoin(-(player.betTotal), common.GainWay_HundredSceneLost, base.SyncFlag_ToClient, "system", s.GetSceneName())
				}
				loseCoin := int32(0) //统计玩家输的金币
				winCoin := int32(0)  //统计玩家赢取的金币
				for flag, pushCoin := range player.CoinPool {
					if pushCoin == 0 {
						continue
					}
					rp := model.RollPerson{UserId: player.SnId}
					if flag == winFlag { //计算压中的位置，赢取的金币为押注金额*倍率
						winCoin = pushCoin * base.SystemChanceMgrEx.RollCoinRate[winFlag]
						loseCoin += pushCoin
						rp.ChangeCoin = int64(winCoin) - int64(pushCoin)
					} else { //未压中的位置，为输掉的金币，输给庄家
						loseCoin += pushCoin
						rp.ChangeCoin = int64(-pushCoin)
					}
					rp.IsRob = player.IsRob
					rp.WBLevel = player.WBLevel
					rp.BeforeCoin = player.GetCurrentCoin()
					rp.IsFirst = sceneEx.IsPlayerFirst(player.Player)
					rp.AfterCoin = player.Coin
					rp.UserBetTotal = int64(pushCoin)
					rollhundredtype[winFlag].IsWin = 1
					if !rp.IsRob {
						rollhundredtype[flag].RollPerson = append(rollhundredtype[flag].RollPerson, rp)
					}
				}
				bankerWinCoin += loseCoin //庄家赢取的金币为玩家输掉的金币
				bankerLoseCoin += winCoin //庄家输掉的金币为玩家压中赢取的金币
				//				//统计
				//				if !player.IsRob {
				//					sceneEx.SummaryData.WinCoin += int64(loseCoin)
				//					sceneEx.SummaryData.LoseCoin += int64(winCoin)
				//				}
				player.gainCoin = winCoin // - loseCoin //玩家最终的金币
				logger.Logger.Tracef("%v win %v coin.", player.GetName(), winCoin)
				logger.Logger.Tracef("%v lose %v coin.", player.GetName(), loseCoin)
				var gainWay int32
				tax := int64(0)
				if player.gainCoin > 0 {
					gainWay = common.GainWay_HundredSceneWin
					tax = int64(player.gainCoin) * int64(s.DbGameFree.GetTaxRate()) / 10000
					player.winCoin = int64(player.gainCoin)
					player.taxCoin = tax
					player.gainCoin -= int32(tax)
					player.AddServiceFee(tax)
				} else {
					gainWay = common.GainWay_HundredSceneLost
				}
				//oldflag := player.MarkBroadcast(false)
				player.AddCoin(int64(player.gainCoin), gainWay, 0, "system", s.GetSceneName())
				//player.MarkBroadcast(oldflag)
				betTotal := int64(player.betTotal)
				if betTotal != 0 {
					//上报游戏事件
					//player.ReportGameEvent(int64(player.gainCoin)-betTotal, tax, betTotal)
				}
				//统计游戏局数
				if winCoin != 0 || loseCoin != 0 {
					if winCoin-loseCoin > 0 {
						player.WinTimes++
					} else {
						player.FailTimes++
					}
				}
			}
		}

		//系统输钱,更新奖池
		//统计系统产出的金币
		systemOut := sceneEx.CalcuSystemCoinOut(int(winFlag), false)
		//if systemOut > 0 {
		base.CoinPoolMgr.PopCoin(sceneEx.GetGameFreeId(), sceneEx.GetGroupId(), sceneEx.GetPlatform(), systemOut)
		//}
		//构建结算消息
		logger.Logger.Trace("Current player count:", len(sceneEx.players))
		var bSaveDetailLog bool
		sceneEx.godid = 0
		wins := 0
		logid, _ := model.AutoIncGameLogId()
		var bankerGainCoin int32 //2020/5/13 修改玩家上庄，飘分不一致的情况，玩家为庄时的coin和banker字段应该一致
		if sceneEx.Banker != nil {
			bankerGainCoin = sceneEx.Banker.gainCoin
		} else {
			bankerGainCoin = bankerWinCoin - bankerLoseCoin
		}
		for _, player := range sceneEx.players {
			betPlayerTotal := int64(0)
			for _, v := range player.CoinPool {
				betPlayerTotal += int64(v)
			}
			//统计投入产出
			inout := int64(player.winCoin) - betPlayerTotal
			if inout != 0 {
				//logger.Logger.Trace("统计投入产出:", inout)
				//赔率统计
				player.Statics(sceneEx.KeyGameId, sceneEx.KeyGamefreeId, inout, true)

			}
			if player.gainCoin > 0 {
				player.WinRecord = append(player.WinRecord, 1)
			} else {
				player.WinRecord = append(player.WinRecord, 0)
			}
			player.BetBigRecord = append(player.BetBigRecord, int(player.GetTotlePushCoin()))

			player.cGetWin20 = player.GetWin20()
			player.cGetBetGig20 = player.GetBetGig20()

			if sceneEx.Banker != nil {
				if sceneEx.Banker.SnId != player.SnId {
					if player.cGetWin20 > wins {
						wins = player.cGetWin20
						sceneEx.godid = player.SnId
					}
				}
			} else {
				if player.cGetWin20 > wins {
					wins = player.cGetWin20
					sceneEx.godid = player.SnId
				}
			}

			packBilled := &rollcoin.SCRollCoinBill{
				Flag:   proto.Int32(winFlag),
				Coin:   proto.Int64(int64(player.gainCoin)),
				Banker: proto.Int64(int64(bankerGainCoin)),
			}
			for key, value := range player.PushCoinLog[winFlag] {
				for i := int32(0); i < value; i++ {
					packBilled.ChipList = append(packBilled.ChipList, key)
				}

			}
			player.SendToClient(int(rollcoin.PACKETID_ROLLCOIN_SC_BILL), packBilled)
			logger.Logger.Trace("SCRollCoinBill:", packBilled)

			betTotal := player.betTotal
			//统计金币变动
			if betTotal > 0 {

				//TODO 这个循环需要优化，太粗暴了
				//for k, v := range rollhundredtype {
				//	if v.RollPerson != nil {
				//		for m, n := range v.RollPerson {
				//			if n.UserId == player.SnId {
				//				n.IsRob = player.IsRob
				//				n.BeforeCoin = player.GetCurrentCoin()
				//				n.AfterCoin = player.Coin
				//				n.UserBetTotal = int64(player.CoinPool[int32(k)])
				//				rollhundredtype[k].RollPerson[m] = n
				//				break
				//			}
				//		}
				//	}
				//}
				if !player.IsRob {
					player.SaveSceneCoinLog(player.GetCurrentCoin(), player.Coin-player.GetCurrentCoin(),
						player.Coin, betTotal, player.taxCoin, player.winCoin, 0, 0)
					totalin, totalout := int64(0), int64(0)
					if player.Coin-player.GetCurrentCoin() > 0 {
						totalout = player.Coin - player.GetCurrentCoin() + player.taxCoin
					} else {
						totalin = -(player.Coin - player.GetCurrentCoin() + player.taxCoin)
					}

					sceneEx.SaveGamePlayerListLog(player.SnId,
						&base.SaveGamePlayerListLogParam{
							Platform:          player.Platform,
							Channel:           player.Channel,
							Promoter:          player.BeUnderAgentCode,
							PackageTag:        player.PackageID,
							InviterId:         player.InviterId,
							LogId:             logid,
							TotalIn:           totalin,
							TotalOut:          totalout,
							TaxCoin:           player.taxCoin,
							ClubPumpCoin:      0,
							BetAmount:         player.betTotal,
							WinAmountNoAnyTax: int64(player.gainCoin),
							IsFirstGame:       sceneEx.IsPlayerFirst(player.Player),
						})
					bSaveDetailLog = true
				}
			}

			//以前坐庄时，没有保存玩家的输赢日志
			if sceneEx.Banker != nil && sceneEx.Banker.SnId == player.SnId {
				if !player.IsRob {
					player.SaveSceneCoinLog(player.GetCurrentCoin(), player.Coin-player.GetCurrentCoin(),
						player.Coin, betTotal, player.taxCoin, player.winCoin, 0, 0)
					totalin, totalout := int64(0), int64(0)
					if player.Coin-player.GetCurrentCoin() > 0 {
						totalout = player.Coin - player.GetCurrentCoin() + player.taxCoin
					} else {
						totalin = -(player.Coin - player.GetCurrentCoin() + player.taxCoin)
					}

					sceneEx.SaveGamePlayerListLog(player.SnId,
						&base.SaveGamePlayerListLogParam{
							Platform:          player.Platform,
							Channel:           player.Channel,
							Promoter:          player.BeUnderAgentCode,
							PackageTag:        player.PackageID,
							InviterId:         player.InviterId,
							LogId:             logid,
							TotalIn:           totalin,
							TotalOut:          totalout,
							TaxCoin:           player.taxCoin,
							ClubPumpCoin:      0,
							BetAmount:         player.betTotal,
							WinAmountNoAnyTax: int64(player.gainCoin),
							IsFirstGame:       sceneEx.IsPlayerFirst(player.Player),
						})
					bSaveDetailLog = true
				}
			}

		}

		//统计参与游戏次数
		if !sceneEx.Testing {
			var playerCtxs []*server.PlayerCtx
			for _, p := range sceneEx.players {
				if p == nil || p.IsRob || p.GetTotlePushCoin() == 0 {
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
		if sceneEx.Banker != nil {
			for _, betCoin := range sceneEx.ZoneTotal {
				sceneEx.Banker.betTotal += betCoin
			}
		}
		//统计下注数
		if !sceneEx.Testing {
			gwPlayerBet := &server.GWPlayerBet{
				GameFreeId: proto.Int32(sceneEx.DbGameFree.GetId()),
				RobotGain:  proto.Int64(-systemOut),
			}
			for _, p := range sceneEx.players {
				if p == nil || p.IsRob || p.betTotal == 0 {
					continue
				}
				playerBet := &server.PlayerBet{
					SnId: proto.Int32(p.SnId),
					Bet:  proto.Int64(int64(p.betTotal)),
					Gain: proto.Int64(p.winCoin - p.betTotal),
					Tax:  proto.Int64(p.taxCoin),
				}
				//庄家下注额不是庄家输的钱
				if sceneEx.Banker != nil && sceneEx.Banker.SnId == p.SnId {
					playerBet.Gain = proto.Int64(p.winCoin)
				}
				gwPlayerBet.PlayerBets = append(gwPlayerBet.PlayerBets, playerBet)
			}
			if len(gwPlayerBet.PlayerBets) > 0 {
				proto.SetDefaults(gwPlayerBet)
				sceneEx.SendToWorld(int(server.SSPacketID_PACKET_GW_PLAYERBET), gwPlayerBet)
				logger.Logger.Trace("Send msg gwPlayerBet ===>", gwPlayerBet)
			}
		}

		//有真人下注再存
		if bSaveDetailLog {
			if sceneEx.Banker != nil {
				//增加莊家統計
				rollhundredtype = append(rollhundredtype, model.RollHundredType{
					RegionId: -1,
					RollPerson: []model.RollPerson{{
						UserId:       sceneEx.Banker.SnId,
						UserBetTotal: sceneEx.Banker.betTotal,
						BeforeCoin:   sceneEx.Banker.GetCurrentCoin(),
						IsFirst:      sceneEx.IsPlayerFirst(sceneEx.Banker.Player),
						AfterCoin:    sceneEx.Banker.Coin,
						ChangeCoin:   int64(sceneEx.Banker.gainCoin),
						IsRob:        sceneEx.Banker.IsRob,
						WBLevel:      sceneEx.Banker.WBLevel,
					}},
				})
			}

			info, err := model.MarshalGameNoteByHUNDRED(&rollhundredtype)
			if err == nil {
				sceneEx.SaveGameDetailedLog(logid, info, &base.GameDetailedParam{})
			}
		}

		for _, player := range sceneEx.players {
			if player != nil {
				player.betTotal = 0
			}
		}
		sceneEx.DealyTime = int64(common.RandFromRange(0, 2000)) * int64(time.Millisecond)
	}
}

func (this *SceneRollCoinStateBilled) OnTick(s *base.Scene) {
	this.SceneBaseStateRollCoin.OnTick(s)
	if sceneEx, ok := s.ExtraData.(*RollCoinSceneData); ok {
		if time.Now().Sub(sceneEx.StateStartTime) > RollCoinBilledTimeout+time.Duration(sceneEx.DealyTime) {
			for _, value := range sceneEx.players {
				value.GameTimes++
				if value.Coin < int64(sceneEx.CoinLimit) || value.Coin < int64(sceneEx.Rollcoin[0]) {
					value.MarkFlag(base.PlayerState_GameBreak)
					value.SyncFlag(true)
				} else {
					if value.IsMarkFlag(base.PlayerState_GameBreak) {
						value.UnmarkFlag(base.PlayerState_GameBreak)
						value.SyncFlag(true)
					}
				}
			}
			s.ChangeSceneState(RollCoinSceneStateWait)
		}
	}
}
func (this *SceneRollCoinStateBilled) OnLeave(s *base.Scene) {
	logger.Logger.Trace("SceneCoinStateRollCoin::OnLeave")
	if sceneEx, ok := s.ExtraData.(*RollCoinSceneData); ok {
		sceneEx.Clean()
		for _, value := range sceneEx.players {
			value.UnmarkFlag(base.PlayerState_WaitNext)
			value.SyncFlag(true)
			if !value.IsOnLine() {
				if value == sceneEx.Banker {
					sceneEx.Banker = nil
				}
				//踢出玩家
				sceneEx.PlayerLeave(value.Player, common.PlayerLeaveReason_DropLine, true)
			} else if value.IsRob {
				if !s.CoinInLimit(value.Coin) {
					if value == sceneEx.Banker {
						sceneEx.Banker = nil
					}
					sceneEx.PlayerLeave(value.Player, common.PlayerLeaveReason_Bekickout, true)
				} else if !common.InSliceInt32(sceneEx.BankerList, value.SnId) && s.CoinOverMaxLimit(value.Coin, value.Player) && value != sceneEx.Banker {
					s.PlayerLeave(value.Player, common.PlayerLeaveReason_Normal, true)
					if value == sceneEx.Banker {
						sceneEx.Banker = nil
					}
				} else if value.Trusteeship >= model.GameParamData.PlayerWatchNum {
					if value != sceneEx.Banker {
						sceneEx.PlayerLeave(value.Player, common.PlayerLeaveReason_LongTimeNoOp, true)
					}
				}
			} else {
				if !s.CoinInLimit(value.Coin) {
					if value == sceneEx.Banker {
						sceneEx.Banker = nil
					}
					sceneEx.PlayerLeave(value.Player, common.PlayerLeaveReason_Bekickout, true)
				} else if value.Trusteeship >= model.GameParamData.PlayerWatchNum {
					if value != sceneEx.Banker {
						sceneEx.PlayerLeave(value.Player, common.PlayerLeaveReason_LongTimeNoOp, true)
					}
				}
				todayGamefreeIDSceneData, _ := value.GetDaliyGameData(int(sceneEx.DbGameFree.GetId()))
				if todayGamefreeIDSceneData != nil &&
					sceneEx.DbGameFree.GetPlayNumLimit() != 0 &&
					todayGamefreeIDSceneData.GameTimes >= int64(sceneEx.DbGameFree.GetPlayNumLimit()) {
					s.PlayerLeave(value.Player, common.PlayerLeaveReason_GameTimes, true)
				}
			}
		}
		//		sceneEx.LogSummary()
		if s.CheckNeedDestroy() {
			sceneEx.SceneDestroy(true)
		}
	}
}
func (this *SceneRollCoinStateBilled) OnPlayerEvent(s *base.Scene, p *base.Player, evtcode int, params []int64) {
	this.SceneBaseStateRollCoin.OnPlayerEvent(s, p, evtcode, params)
	switch evtcode {
	case base.PlayerEventEnter:
		fallthrough
	case base.PlayerEventRehold:
		if playerEx, ok := p.ExtraData.(*RollCoinPlayerData); ok {
			playerEx.MarkFlag(base.PlayerState_WaitNext)
			playerEx.SyncFlag(true)
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
func (this *ScenePolicyRollCoin) RegisteSceneState(state base.SceneState) {
	if state == nil {
		return
	}
	stateid := state.GetState()
	if stateid < 0 || stateid >= RollCoinSceneStateMax {
		return
	}
	this.states[stateid] = state
}

//
func (this *ScenePolicyRollCoin) GetSceneState(s *base.Scene, stateid int) base.SceneState {
	if stateid >= 0 && stateid < RollCoinSceneStateMax {
		return ScenePolicyRollCoinSington.states[stateid]
	}
	return nil
}

//
//func (this *ScenePolicyRollCoin) GetGameSubState(s *base.Scene, stateid int) base.SceneGamingSubState {
//	return nil
//}

func init() {
	ScenePolicyRollCoinSington.RegisteSceneState(&SceneRollCoinStateWait{})
	ScenePolicyRollCoinSington.RegisteSceneState(&SceneRollCoinStateStart{})
	ScenePolicyRollCoinSington.RegisteSceneState(&SceneRollCoinStateRoll{})
	ScenePolicyRollCoinSington.RegisteSceneState(&SceneRollCoinStateOver{})
	ScenePolicyRollCoinSington.RegisteSceneState(&SceneRollCoinStateGame{})
	ScenePolicyRollCoinSington.RegisteSceneState(&SceneRollCoinStateBilled{})
	core.RegisteHook(core.HOOK_BEFORE_START, func() error {
		base.RegisteScenePolicy(common.GameId_RollCoin, 0, ScenePolicyRollCoinSington)
		base.RegisteScenePolicy(common.GameId_RollCoin, 1, ScenePolicyRollCoinSington)
		base.RegisteScenePolicy(common.GameId_RollCoin, 3, ScenePolicyRollCoinSington)
		return nil
	})
}

////////////////////////////////////////////////////////////////////////////////
