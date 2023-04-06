package rollpoint

import (
	"encoding/json"
	"fmt"
	"games.yol.com/win88/gamesrv/base"
	"games.yol.com/win88/protocol/player"
	"games.yol.com/win88/protocol/server"
	"sort"
	"time"

	"games.yol.com/win88/common"
	rule "games.yol.com/win88/gamerule/rollpoint"
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/rollpoint"
	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/logger"
)

const (
	RollPointWaiteTimeout  = time.Second * 0  //等待
	RollPointStartTimeout  = time.Second * 5  //开始
	RollPointBetTimeout    = time.Second * 15 //押注
	RollPointOpenTimeout   = time.Second * 5  //开奖
	RollPointBilledTimeout = time.Second * 10 //结算
)

//场景状态
const (
	RollPointSceneStateWait   int = iota //等待状态
	RollPointSceneStateStart             //开始
	RollPointSceneStateBet               //押注
	RollPointSceneStateOpen              //开奖
	RollPointSceneStateBilled            //结算
	RollPointSceneStateMax
)

//玩家操作
const (
	RollPointPlayerOpPushCoin int = iota //押注
	RollPointPlayerList                  //玩家列表
)

var ScenePolicyRollPointSington = &ScenePolicyRollPoint{}

type ScenePolicyRollPoint struct {
	base.BaseScenePolicy
	states [RollPointSceneStateMax]base.SceneState
}

//创建场景扩展数据
func (this *ScenePolicyRollPoint) CreateSceneExData(s *base.Scene) interface{} {
	logger.Logger.Trace("(this *ScenePolicyRollPoint) CreateSceneExData, sceneId=", s.GetSceneId())
	sceneEx := NewRollPointSceneData(s)
	if sceneEx != nil {
		if sceneEx.init() {
			s.SetExtraData(sceneEx)
		}
	}
	return sceneEx
}

//创建玩家扩展数据
func (this *ScenePolicyRollPoint) CreatePlayerExData(s *base.Scene, p *base.Player) interface{} {
	playerEx := NewRollPointPlayerData(p)
	if playerEx != nil {
		p.SetExtraData(playerEx)
	}
	return playerEx
}

//场景开启事件
func (this *ScenePolicyRollPoint) OnStart(s *base.Scene) {
	logger.Logger.Trace("(this *ScenePolicyRollPoint) OnStart, sceneId=", s.GetSceneId())
	sceneEx := NewRollPointSceneData(s)
	if sceneEx != nil {
		if sceneEx.init() {
			s.SetExtraData(sceneEx)
			s.ChangeSceneState(RollPointSceneStateWait)
		}
	}
}

//场景心跳事件
func (this *ScenePolicyRollPoint) OnTick(s *base.Scene) {
	if s == nil {
		return
	}
	if s.GetSceneState() != nil {
		s.GetSceneState().OnTick(s)
	}
}

//玩家进入事件
func (this *ScenePolicyRollPoint) OnPlayerEnter(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyRollPoint) OnPlayerEnter, sceneId=", s.GetSceneId(), " player=", p.Name)
	if sceneEx, ok := s.GetExtraData().(*RollPointSceneData); ok {
		s.FirePlayerEvent(p, base.PlayerEventEnter, nil)
		playerEx := NewRollPointPlayerData(p)
		playerEx.init()
		sceneEx.players[p.SnId] = playerEx
		p.SetExtraData(playerEx)
		if playerEx.GetCoin() < int64(sceneEx.GetDBGameFree().GetBetLimit()) || playerEx.GetCoin() < int64(sceneEx.GetDBGameFree().GetOtherIntParams()[0]) { //进入房间，金币少于规定值，只能观看，不能下注
			playerEx.MarkFlag(base.PlayerState_GameBreak)
			playerEx.SyncFlag(true)
		}
		//向其他人广播玩家进入信息
		pack := &rollpoint.SCRollPointPlayerNum{
			PlayerNum: proto.Int(sceneEx.CountPlayer()),
		}
		proto.SetDefaults(pack)
		s.Broadcast(int(rollpoint.RPPACKETID_ROLLPOINT_SC_PLAYERNUM), pack, 0)
		logger.Logger.Trace("SCRollPointPlayerNum:", pack)
		//给自己发送房间信息
		this.SendRoomInfo(s, p, sceneEx, playerEx)
	}
}

//玩家离开事件
func (this *ScenePolicyRollPoint) OnPlayerLeave(s *base.Scene, p *base.Player, reason int) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyRollPoint) OnPlayerLeave, sceneId=", s.GetSceneId(), " player=", p.Name)
	if sceneEx, ok := s.GetExtraData().(*RollPointSceneData); ok {
		s.FirePlayerEvent(p, base.PlayerEventLeave, nil)
		sceneEx.OnPlayerLeave(p, reason)
	}
}

//玩家掉线
func (this *ScenePolicyRollPoint) OnPlayerDropLine(s *base.Scene, p *base.Player) {
	logger.Logger.Trace("(this *ScenePolicyRollPoint) OnPlayerDropLine, sceneId=", s.GetSceneId(), " player=", p.Name)
	if s == nil || p == nil {
		return
	}
	s.FirePlayerEvent(p, base.PlayerEventDropLine, nil)
}

//玩家重连
func (this *ScenePolicyRollPoint) OnPlayerRehold(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyRollPoint) OnPlayerRehold, sceneId=", s.GetSceneId(), " player=", p.Name)
	if sceneEx, ok := s.GetExtraData().(*RollPointSceneData); ok {
		if playerEx, ok := p.GetExtraData().(*RollPointPlayerData); ok {
			pack := &rollpoint.SCRollPointPlayerNum{
				PlayerNum: proto.Int(sceneEx.CountPlayer()),
			}
			proto.SetDefaults(pack)
			playerEx.SendToClient(int(rollpoint.RPPACKETID_ROLLPOINT_SC_PLAYERNUM), pack)
			//发送房间信息给自己
			this.SendRoomInfo(s, p, sceneEx, playerEx)
			s.FirePlayerEvent(p, base.PlayerEventRehold, nil)
		}
	}
}

//玩家重连
func (this *ScenePolicyRollPoint) OnPlayerReturn(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyRollPoint) OnPlayerReturn, sceneId=", s.GetSceneId(), " player=", p.Name)
	if sceneEx, ok := s.GetExtraData().(*RollPointSceneData); ok {
		if playerEx, ok := p.GetExtraData().(*RollPointPlayerData); ok {
			pack := &rollpoint.SCRollPointPlayerNum{
				PlayerNum: proto.Int(sceneEx.CountPlayer()),
			}
			proto.SetDefaults(pack)
			playerEx.SendToClient(int(rollpoint.RPPACKETID_ROLLPOINT_SC_PLAYERNUM), pack)
			//发送房间信息给自己
			this.SendRoomInfo(s, p, sceneEx, playerEx)
			s.FirePlayerEvent(p, base.PlayerEventReturn, nil)
		}
	}
}

//玩家操作
func (this *ScenePolicyRollPoint) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if s == nil || p == nil {
		return false
	}
	if s.GetSceneState() != nil {
		return s.GetSceneState().OnPlayerOp(s, p, opcode, params)
	}
	return true
}

//玩家事件
func (this *ScenePolicyRollPoint) OnPlayerEvent(s *base.Scene, p *base.Player, evtcode int, params []int64) {
	if s == nil || p == nil {
		return
	}
	if s.GetSceneState() != nil {
		s.GetSceneState().OnPlayerEvent(s, p, evtcode, params)
	}
}

//观众进入
func (this *ScenePolicyRollPoint) OnAudienceEnter(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyRollPoint) OnAudienceEnter, sceneId=", s.GetSceneId(), " player=", p.Name)
	if sceneEx, ok := s.GetExtraData().(*RollPointSceneData); ok {
		this.SendRoomInfo(s, p, sceneEx, nil)
	}
}

//观众掉线
func (this *ScenePolicyRollPoint) OnAudienceDropLine(s *base.Scene, p *base.Player) {
	if s == nil || p == nil {
		return
	}
	logger.Logger.Trace("(this *ScenePolicyRollPoint) OnAudienceDropLine, sceneId=", s.GetSceneId(), " player=", p.Name)
	s.AudienceLeave(p, common.PlayerLeaveReason_DropLine)
}

//是否完成了整个牌局
func (this *ScenePolicyRollPoint) IsCompleted(s *base.Scene) bool {
	if s == nil {
		return false
	}
	return s.GetState() == RollPointSceneStateBilled
}

//是否可以强制开始
func (this *ScenePolicyRollPoint) IsCanForceStart(s *base.Scene) bool {
	if sceneEx, ok := s.GetExtraData().(*RollPointSceneData); ok {
		return len(sceneEx.players) > 0
	}
	return false
}

//强制开始
func (this *ScenePolicyRollPoint) ForceStart(s *base.Scene) {
	s.ChangeSceneState(RollPointSceneStateWait)
}

//当前状态能否换桌
func (this *ScenePolicyRollPoint) CanChangeCoinScene(s *base.Scene, p *base.Player) bool {
	if s == nil || p == nil {
		return false
	}
	if playerEx, ok := p.GetExtraData().(*RollPointPlayerData); ok {
		if playerEx.GetTotlePushCoin() > 0 { //没押注可以离开
			p.OpCode = player.OpResultCode_OPRC_Hundred_YouHadBetCannotLeave
			return false
		}
	}
	if s.GetSceneState() != nil {
		return s.GetSceneState().CanChangeCoinScene(s, p)
	}
	return true
}

func (this *ScenePolicyRollPoint) SendRoomInfo(s *base.Scene, p *base.Player, sceneEx *RollPointSceneData, playerEx *RollPointPlayerData) {
	pack := &rollpoint.SCRollPointRoomInfo{
		RollLog:   sceneEx.WinLog,
		SceneType: proto.Int(s.SceneType),
		State:     proto.Int(s.GetSceneState().GetState()),
		TimeOut:   proto.Int(s.GetSceneState().GetTimeout(s)),
		GameId:    proto.Int(common.GameId_RollPoint),
		RoomMode:  proto.Int(s.GetSceneMode()),
		SceneId:   proto.Int(sceneEx.GetSceneId()),
		ParamsEx:  s.GetParamsEx(),
		Params: []int32{s.GetDBGameFree().GetLimitCoin(), s.GetDBGameFree().GetMaxCoinLimit(), s.GetDBGameFree().GetServiceFee(),
			s.GetDBGameFree().GetBanker(), s.GetDBGameFree().GetBaseScore(), int32(s.SceneType),
			s.GetDBGameFree().GetLowerThanKick(), s.GetDBGameFree().GetBetLimit()},
	}
	for _, value := range s.GetDBGameFree().GetMaxBetCoin() {
		pack.Params = append(pack.Params, value)
	}
	pack.Player = &rollpoint.RollPointPlayer{
		SnId: proto.Int32(playerEx.SnId),
		Name: proto.String(playerEx.Name),
		Sex:  proto.Int32(playerEx.Sex),
		Head: proto.Int32(playerEx.Head),
		Coin: proto.Int64(playerEx.GetCoin() - playerEx.allBetCoin),
		//Params:      proto.String(playerEx.Params),
		Flag:        proto.Int(playerEx.GetFlag()),
		HeadOutLine: proto.Int32(playerEx.HeadOutLine),
		VIP:         proto.Int32(playerEx.VIP),
		City:        proto.String(playerEx.GetCity()),
	}
	for i := int32(0); i < rule.BetArea_Max; i++ {
		coinLog := &rollpoint.RollPointCoinLog{Index: proto.Int32(i)}
		if pool, ok := sceneEx.TotalBet[i]; ok {
			for _, coin := range sceneEx.RollPoint {
				coinLog.Coins = append(coinLog.Coins, coin, int32(pool[coin]))
			}
		}
		pack.CoinPool = append(pack.CoinPool, coinLog)
	}
	player := sceneEx.players[p.SnId]
	for i := int32(0); i < rule.BetArea_Max; i++ {
		coinLog := &rollpoint.RollPointCoinLog{Index: proto.Int32(i)}
		if pool, ok := player.PushCoinLog[i]; ok {
			for _, coin := range sceneEx.RollPoint {
				coinLog.Coins = append(coinLog.Coins, coin, int32(pool[coin]))
			}
		}
		pack.PushLog = append(pack.PushLog, coinLog)
	}
	proto.SetDefaults(pack)
	p.SendToClient(int(rollpoint.RPPACKETID_ROLLPOINT_SC_ROOMINFO), pack)
	logger.Logger.Trace("SCRollPointRoomInfo:", pack)
}
func (this *ScenePolicyRollPoint) NotifyGameState(s *base.Scene) {
	if s.GetState() > 2 {
		return
	}
	sec := int((RollPointStartTimeout + RollPointBetTimeout).Seconds())
	s.SyncGameState(sec, 0)
}

//////////////////////////////////////////////////////////////
//状态基类
//////////////////////////////////////////////////////////////
type SceneBaseStateRollPoint struct {
}

func (this *SceneBaseStateRollPoint) GetTimeout(s *base.Scene) int {
	if sceneEx, ok := s.GetExtraData().(*RollPointSceneData); ok {
		return int(time.Now().Sub(sceneEx.StateStartTime) / time.Second)
	}
	return 0
}
func (this *SceneBaseStateRollPoint) CanChangeTo(s base.SceneState) bool {
	return true
}
func (this *SceneBaseStateRollPoint) OnEnter(s *base.Scene) {
	if sceneEx, ok := s.GetExtraData().(*RollPointSceneData); ok {
		sceneEx.StateStartTime = time.Now()
	}
}
func (this *SceneBaseStateRollPoint) OnLeave(s *base.Scene) {}
func (this *SceneBaseStateRollPoint) OnTick(s *base.Scene)  {}
func (this *SceneBaseStateRollPoint) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if sceneEx, ok := s.GetExtraData().(*RollPointSceneData); ok {
		switch opcode {
		case RollPointPlayerList:
			//清空
			for i := len(sceneEx.seats) - 1; i >= 0; i-- {
				sceneEx.seats = append(sceneEx.seats[:i], sceneEx.seats[i+1:]...)
			}
			for _, sp := range sceneEx.players {
				sceneEx.seats = append(sceneEx.seats, sp)
			}

			packList := &rollpoint.SCRollPointPlayerList{
				PlayerNum: proto.Int32(int32(sceneEx.CountPlayer())),
			}
			sort.Slice(sceneEx.seats, func(i, j int) bool {
				return sceneEx.seats[i].cGetBetGig20 > sceneEx.seats[j].cGetBetGig20
			})
			//近20局押注胜率最高的玩家
			var winTimesBig *RollPointPlayerData
			winTimesBig = nil
			for _, v := range sceneEx.seats {
				if v.IsMarkFlag(base.PlayerState_GameBreak) { //观众
					continue
				}
				if v.cGetWin20 <= 0 { //近20局胜率不能为0
					continue
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
				curcoin := winTimesBig.GetCoin()
				pd := &rollpoint.RollPointPlayer{
					SnId:        proto.Int32(winTimesBig.SnId),
					Name:        proto.String(winTimesBig.Name),
					Head:        proto.Int32(winTimesBig.Head),
					Sex:         proto.Int32(winTimesBig.Sex),
					Coin:        proto.Int64(curcoin - winTimesBig.allBetCoin),
					Flag:        proto.Int(winTimesBig.GetFlag()),
					Lately20Bet: proto.Int32(int32(winTimesBig.cGetBetGig20)),
					Lately20Win: proto.Int32(int32(winTimesBig.cGetWin20)),
					HeadOutLine: proto.Int32(winTimesBig.HeadOutLine),
					VIP:         proto.Int32(winTimesBig.VIP),
					City:        proto.String(winTimesBig.GetCity()),
				}
				packList.Data = append(packList.Data, pd)
			}

			i20 := 0 //只显示20个玩家信息
			for _, pp := range sceneEx.seats {
				if pp.IsMarkFlag(base.PlayerState_GameBreak) { //观众
					continue
				}
				if !pp.IsMarkFlag(base.PlayerState_Check) {
					i20++
					if i20 > 20 {
						break
					}
					curcoin := pp.GetCoin()
					pd := &rollpoint.RollPointPlayer{
						SnId:        proto.Int32(pp.SnId),
						Name:        proto.String(pp.Name),
						Head:        proto.Int32(pp.Head),
						Sex:         proto.Int32(pp.Sex),
						Coin:        proto.Int64(curcoin - pp.allBetCoin),
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
			p.SendToClient(int(rollpoint.RPPACKETID_ROLLPOINT_SC_PLAYERLIST), packList)
			return true
		}
	}
	return false
}

func (this *SceneBaseStateRollPoint) OnPlayerEvent(s *base.Scene, p *base.Player, evtcode int, params []int64) {
	if sceneEx, ok := s.GetExtraData().(*RollPointSceneData); ok {
		if playerEx, ok := p.GetExtraData().(*RollPointPlayerData); ok {
			switch evtcode {
			case base.PlayerEventRecharge:
				//oldflag := p.MarkBroadcast(false)
				p.AddCoin(params[0], common.GainWay_Pay, base.SyncFlag_ToClient, "system", p.GetScene().GetSceneName())
				//p.MarkBroadcast(oldflag)
				if p.GetCoin() >= int64(sceneEx.GetDBGameFree().GetBetLimit()) && p.GetCoin() >= int64(sceneEx.GetDBGameFree().GetOtherIntParams()[0]) {
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
type SceneRollPointStateWait struct {
	SceneBaseStateRollPoint
}

func (this *SceneRollPointStateWait) GetState() int { return RollPointSceneStateWait }
func (this *SceneRollPointStateWait) CanChangeCoinScene(s *base.Scene, p *base.Player) bool {
	return true
}
func (this *SceneRollPointStateWait) CanChangeTo(s base.SceneState) bool {
	switch s.GetState() {
	case RollPointSceneStateStart:
		return true
	default:
		return false
	}
}
func (this *SceneRollPointStateWait) OnEnter(s *base.Scene) {
	this.SceneBaseStateRollPoint.OnEnter(s)
	if sceneEx, ok := s.GetExtraData().(*RollPointSceneData); ok {
		sceneEx.GameStartTime = time.Now()
		pack := &rollpoint.SCRollPointRoomState{
			State: proto.Int(this.GetState()),
		}
		proto.SetDefaults(pack)
		s.Broadcast(int(rollpoint.RPPACKETID_ROLLPOINT_SC_ROOMSTATE), pack, 0)
	}
}
func (this *SceneRollPointStateWait) OnTick(s *base.Scene) {
	this.SceneBaseStateRollPoint.OnTick(s)
	if sceneEx, ok := s.GetExtraData().(*RollPointSceneData); ok {
		if time.Now().Sub(sceneEx.StateStartTime) > RollPointWaiteTimeout {
			s.ChangeSceneState(RollPointSceneStateStart)
		}
	}
}
func (this *SceneRollPointStateWait) OnLeave(s *base.Scene) {
	this.SceneBaseStateRollPoint.OnLeave(s)
}
func (this *SceneRollPointStateWait) OnPlayerEvent(s *base.Scene, p *base.Player, evtcode int, params []int64) {
	this.SceneBaseStateRollPoint.OnPlayerEvent(s, p, evtcode, params)
}

//////////////////////////////////////////////////////////////
//开始状态
//////////////////////////////////////////////////////////////
type SceneRollPointStateStart struct {
	SceneBaseStateRollPoint
}

func (this *SceneRollPointStateStart) GetState() int { return RollPointSceneStateStart }
func (this *SceneRollPointStateStart) CanChangeCoinScene(s *base.Scene, p *base.Player) bool {
	return true
}
func (this *SceneRollPointStateStart) CanChangeTo(s base.SceneState) bool {
	switch s.GetState() {
	case RollPointSceneStateBet:
		return true
	}
	return false
}

func (this *SceneRollPointStateStart) OnEnter(s *base.Scene) {
	this.SceneBaseStateRollPoint.OnEnter(s)
	if sceneEx, ok := s.GetExtraData().(*RollPointSceneData); ok {
		sceneEx.Clean()
		sceneEx.NumOfGames++
		sceneEx.GameNowTime = time.Now()
		sceneEx.blackBox.Roll()
		pack := &rollpoint.SCRollPointRoomState{
			State: proto.Int(this.GetState()),
		}
		proto.SetDefaults(pack)
		s.Broadcast(int(rollpoint.RPPACKETID_ROLLPOINT_SC_ROOMSTATE), pack, 0)
	}
}
func (this *SceneRollPointStateStart) OnPlayerEvent(s *base.Scene, p *base.Player, evtcode int, params []int64) {
	this.SceneBaseStateRollPoint.OnPlayerEvent(s, p, evtcode, params)
}
func (this *SceneRollPointStateStart) OnTick(s *base.Scene) {
	this.SceneBaseStateRollPoint.OnTick(s)
	if sceneEx, ok := s.GetExtraData().(*RollPointSceneData); ok {
		if time.Now().Sub(sceneEx.StateStartTime) > RollPointStartTimeout {
			s.ChangeSceneState(RollPointSceneStateBet)
		}
	}
}
func (this *SceneRollPointStateStart) OnLeave(s *base.Scene) {
	logger.Logger.Trace("SceneRollPointStateStart::OnLeave")
	if sceneEx, ok := s.GetExtraData().(*RollPointSceneData); ok {
		sceneEx.CalcLastWiner()
	}
}

//////////////////////////////////////////////////////////////
//押注
//////////////////////////////////////////////////////////////
type SceneRollPointStateRoll struct {
	SceneBaseStateRollPoint
}

func (this *SceneRollPointStateRoll) GetState() int { return RollPointSceneStateBet }
func (this *SceneRollPointStateRoll) CanChangeCoinScene(s *base.Scene, p *base.Player) bool {
	return true
}
func (this *SceneRollPointStateRoll) CanChangeTo(s base.SceneState) bool {
	switch s.GetState() {
	case RollPointSceneStateOpen:
		return true
	}
	return false
}
func (this *SceneRollPointStateRoll) OnEnter(s *base.Scene) {
	this.SceneBaseStateRollPoint.OnEnter(s)
	if sceneEx, ok := s.GetExtraData().(*RollPointSceneData); ok {
		pack := &rollpoint.SCRollPointRoomState{
			State: proto.Int(this.GetState()),
		}
		proto.SetDefaults(pack)
		s.Broadcast(int(rollpoint.RPPACKETID_ROLLPOINT_SC_ROOMSTATE), pack, 0)
		sceneEx.SyncTime = time.Now()
		for key, pool := range sceneEx.SyncData {
			for coin, _ := range pool {
				sceneEx.SyncData[key][coin] = 0
			}
		}
		for _, player := range sceneEx.players {
			if player == nil {
				continue
			}
		}
	}
}
func (this *SceneRollPointStateRoll) OnTick(s *base.Scene) {
	this.SceneBaseStateRollPoint.OnTick(s)
	if sceneEx, ok := s.GetExtraData().(*RollPointSceneData); ok {
		if time.Now().Sub(sceneEx.StateStartTime) > RollPointBetTimeout {
			s.ChangeSceneState(RollPointSceneStateOpen)
		}
		if time.Now().Sub(sceneEx.SyncTime) > time.Second*2 {
			sceneEx.SyncCoinLog()
			sceneEx.SyncTime = time.Now()
		}
	}
}
func (this *SceneRollPointStateRoll) OnPlayerOp(s *base.Scene, p *base.Player, opcode int, params []int64) bool {
	if this.SceneBaseStateRollPoint.OnPlayerOp(s, p, opcode, params) {
		return true
	}
	if sceneEx, ok := s.GetExtraData().(*RollPointSceneData); ok {
		if playerEx, ok := p.GetExtraData().(*RollPointPlayerData); ok {
			switch opcode {
			case RollPointPlayerOpPushCoin:
				if len(params) < 2 {
					logger.Logger.Error("Client send error param in RollPointPlayerData.")
					return false
				}
				if playerEx.IsMarkFlag(base.PlayerState_GameBreak) {
					return false
				}
				flag := params[0]
				coin := params[1]
				expectCoin := coin
				if int32(flag) < 0 || int32(flag) >= rule.BetArea_Max {
					logger.Logger.Info("Recive client error flag:", params, "in push coin op.")
					return false
				}
				if !common.InSliceInt32(sceneEx.RollPoint, int32(coin)) {
					logger.Logger.Info("Recive client error coin:", params, "in push coin op.")
					return false
				}
				minChip := int64(sceneEx.RollPoint[0])
				//押注的金币不能超过拥有的金币
				if coin > (playerEx.GetCoin() - playerEx.allBetCoin) {
					return true
				}
				//押注限制(低于该值不能押注)
				if p.GetCoin() < int64(sceneEx.GetDBGameFree().GetBetLimit()) && playerEx.allBetCoin == 0 {
					msgPack := common.CreateSrvMsg(common.SRVMSG_CODE_DEFAULT,
						fmt.Sprintf("抱歉，%.2f金币以上才能下注", float64(sceneEx.GetDBGameFree().GetBetLimit())/100))
					p.SendToClient(int(player.PlayerPacketID_PACKET_SC_SRVMSG), msgPack)
					return true
				}
				//单门押注上限
				maxBetCoin := sceneEx.GetDBGameFree().GetMaxBetCoin()
				if ok, coinLimit := playerEx.MaxChipCheck(flag, coin, maxBetCoin); !ok {
					coin = int64(coinLimit - playerEx.CoinPool[int32(flag)])
					//对齐到最小面额的筹码
					coin /= minChip
					coin *= minChip
					if coin <= 0 {
						msgPack := common.CreateSrvMsg(common.SRVMSG_CODE_DEFAULT,
							fmt.Sprintf("该门押注金额上限%.2f", float64(coinLimit)/100))
						p.SendToClient(int(player.PlayerPacketID_PACKET_SC_SRVMSG), msgPack)
						return true
					}
				}
				playerEx.Trusteeship = 0
				playerEx.gameing = true
				if coin == expectCoin {
					playerEx.PushCoin(int32(flag), int32(coin))
					sceneEx.CacheCoinLog(int32(flag), int32(coin), playerEx.IsRob)
				} else {
					val := coin
					for i := len(sceneEx.RollPoint) - 1; i >= 0 && val > 0; i-- {
						chip := int64(sceneEx.RollPoint[i])
						cntChip := val / chip
						for j := int64(0); j < cntChip; j++ {
							playerEx.PushCoin(int32(flag), int32(chip))
							sceneEx.CacheCoinLog(int32(flag), int32(chip), playerEx.IsRob)
						}
						val -= cntChip * chip
					}
				}
				playerEx.TotalBet += int64(coin)
				playerEx.allBetCoin += int64(coin)
				//oldflag := playerEx.MarkBroadcast(false)
				//playerEx.MarkBroadcast(oldflag)
				pack := &rollpoint.SCRollPointPushCoin{
					OpRetCode: rollpoint.OpResultCode_OPRC_Sucess,
					Coin:      proto.Int32(int32(coin)),
					Flag:      proto.Int32(int32(flag)),
					MeTotle:   proto.Int64(int64(playerEx.GetTotlePushCoin())),
					SnId:      proto.Int32(playerEx.SnId),
				}
				pack.PlayerType = proto.Int32(0)
				proto.SetDefaults(pack)
				playerEx.SendToClient(int(rollpoint.RPPACKETID_ROLLPOINT_SC_PUSHCOIN), pack)
				if sceneEx.LastWinerSnid == playerEx.SnId {
					sceneEx.LastWinerBetPos = append(sceneEx.LastWinerBetPos, int32(flag))
					pack := &rollpoint.SCRollPointCoinLog{
						Pos: sceneEx.LastWinerBetPos,
					}
					sceneEx.Broadcast(int(rollpoint.RPPACKETID_ROLLPOINT_SC_COINLOG), pack, 0)
				}
			}
		}
	}
	return true
}

func (this *SceneRollPointStateRoll) OnPlayerEvent(s *base.Scene, p *base.Player, evtcode int, params []int64) {
	this.SceneBaseStateRollPoint.OnPlayerEvent(s, p, evtcode, params)
}
func (this *SceneRollPointStateRoll) OnLeave(s *base.Scene) {
	logger.Logger.Trace("SceneCoinStateRollPoint::OnLeave")
	if sceneEx, ok := s.GetExtraData().(*RollPointSceneData); ok {
		sceneEx.SyncCoinLog()
		sceneEx.ChangeCard()
	}
}

//////////////////////////////////////////////////////////////
//开奖状态
//////////////////////////////////////////////////////////////
type SceneRollPointStateOpen struct {
	SceneBaseStateRollPoint
}

func (this *SceneRollPointStateOpen) GetState() int { return RollPointSceneStateOpen }
func (this *SceneRollPointStateOpen) CanChangeCoinScene(s *base.Scene, p *base.Player) bool {
	return true
}
func (this *SceneRollPointStateOpen) CanChangeTo(s base.SceneState) bool {
	switch s.GetState() {
	case RollPointSceneStateBilled:
		return true
	}
	return false
}

func (this *SceneRollPointStateOpen) OnEnter(s *base.Scene) {
	this.SceneBaseStateRollPoint.OnEnter(s)
	if sceneEx, ok := s.GetExtraData().(*RollPointSceneData); ok {
		pack := &rollpoint.SCRollPointRoomState{
			State: proto.Int(this.GetState()),
		}
		for _, value := range sceneEx.blackBox.Point {
			pack.Params = append(pack.Params, value)
		}
		for _, value := range sceneEx.blackBox.Score {
			pack.Params = append(pack.Params, value)
		}
		proto.SetDefaults(pack)
		s.Broadcast(int(rollpoint.RPPACKETID_ROLLPOINT_SC_ROOMSTATE), pack, 0)
		for _, v := range sceneEx.players {
			if v != nil && !v.IsRob {
				if v.allBetCoin <= 0 {
					v.Trusteeship++
					if v.Trusteeship >= model.GameParamData.NotifyPlayerWatchNum {
						v.SendTrusteeshipTips()
					}
				}
			}
		}
	}
}
func (this *SceneRollPointStateOpen) OnPlayerEvent(s *base.Scene, p *base.Player, evtcode int, params []int64) {
	this.SceneBaseStateRollPoint.OnPlayerEvent(s, p, evtcode, params)
}
func (this *SceneRollPointStateOpen) OnTick(s *base.Scene) {
	this.SceneBaseStateRollPoint.OnTick(s)
	if sceneEx, ok := s.GetExtraData().(*RollPointSceneData); ok {
		if time.Now().Sub(sceneEx.StateStartTime) > RollPointOpenTimeout {
			s.ChangeSceneState(RollPointSceneStateBilled)
		}
	}
}
func (this *SceneRollPointStateOpen) OnLeave(s *base.Scene) {
	if sceneEx, ok := s.GetExtraData().(*RollPointSceneData); ok {
		point := sceneEx.blackBox.Point
		pack := &server.GWGameStateLog{
			SceneId: proto.Int(s.GetSceneId()),
			GameLog: proto.Int32(point[0] | point[1]<<8 | point[2]<<16),
		}
		s.SendToWorld(int(server.SSPacketID_PACKET_GW_GAMESTATELOG), pack)
	}

}

//////////////////////////////////////////////////////////////
//结算状态
//////////////////////////////////////////////////////////////
type SceneRollPointStateBilled struct {
	SceneBaseStateRollPoint
}

func (this *SceneRollPointStateBilled) GetState() int {
	return RollPointSceneStateBilled
}

func (this *SceneRollPointStateBilled) CanChangeTo(s base.SceneState) bool {
	switch s.GetState() {
	case RollPointSceneStateStart, RollPointSceneStateWait:
		return true
	}
	return false
}

//当前状态能否换桌
func (this *SceneRollPointStateBilled) CanChangeCoinScene(s *base.Scene, p *base.Player) bool {
	return true
}

func (this *SceneRollPointStateBilled) OnEnter(s *base.Scene) {
	logger.Logger.Trace("SceneBilledStateRollPoint::OnEnter")
	this.SceneBaseStateRollPoint.OnEnter(s)
	if sceneEx, ok := s.GetExtraData().(*RollPointSceneData); ok {
		//水池上下文环境
		cpCtx := base.CoinPoolMgr.GetCoinPoolCtx(sceneEx.GetPlatform(), sceneEx.GetGameFreeId(), sceneEx.GetGroupId())
		cpCtx.Controlled = sceneEx.GetCpControlled()
		sceneEx.SetCpCtx(cpCtx)
		//计算赢家位置
		sceneEx.LogWinFlag()
		//发送状态消息
		pack := &rollpoint.SCRollPointRoomState{
			State: proto.Int(this.GetState()),
		}
		proto.SetDefaults(pack)
		s.Broadcast(int(rollpoint.RPPACKETID_ROLLPOINT_SC_ROOMSTATE), pack, 0)
		//计算积分，结算金币
		bankerWinCoin := int64(0)  //统计庄家赢取的金币
		bankerLoseCoin := int64(0) //统计庄家输的金币
		////////////////////////////////////////////////////
		rollhundredtype := model.RollPointType{
			RoomId: int32(s.GetSceneId()),
		}
		for _, value := range sceneEx.blackBox.Point {
			rollhundredtype.Point = append(rollhundredtype.Point, value)
		}
		for _, value := range sceneEx.blackBox.Score {
			rollhundredtype.Score = append(rollhundredtype.Score, value)
		}
		//系统坐庄
		for _, player := range sceneEx.players {
			if player == nil {
				continue
			}
			loseCoin := int64(0) //统计玩家输的金币
			winCoin := int64(0)  //统计玩家赢取的金币
			returnCoin := int32(0)
			rp := model.RollPointPerson{
				UserId:       player.SnId,
				IsRob:        player.IsRob,
				WBLevel:      player.WBLevel,
				BeforeCoin:   player.GetTakeCoin(),
				IsFirst:      sceneEx.IsPlayerFirst(player.Player),
				UserBetTotal: int64(player.GetTotlePushCoin()),
			}
			rp.BetCoin = make([]int32, rule.BetArea_Max, rule.BetArea_Max)
			for flag, pushCoin := range player.CoinPool {
				if pushCoin == 0 {
					continue
				}
				if sceneEx.blackBox.Score[flag] > 0 { //计算压中的位置，赢取的金币为押注金额*倍率
					winCoin += int64(pushCoin * sceneEx.blackBox.Score[flag])
					returnCoin += pushCoin //押注返还
					rp.ChangeCoin += (int64(winCoin) + int64(pushCoin))
				} else { //未压中的位置，为输掉的金币，输给庄家
					loseCoin += int64(pushCoin)
					rp.ChangeCoin = int64(-pushCoin)
				}
				rp.BetCoin[flag] = pushCoin
			}
			player.gainCoin = int32(winCoin) + returnCoin
			tax := int64(0)
			if player.gainCoin > 0 {
				player.AddCoin(-(player.allBetCoin), common.GainWay_HundredSceneLost, 0, "system", s.GetSceneName())
				tax = int64(float64(winCoin) * (float64(s.GetDBGameFree().GetTaxRate()) / 10000))
				player.winCoin = int64(player.gainCoin)
				player.taxCoin = tax
				player.gainCoin -= int32(tax)
				player.AddServiceFee(tax)
				//oldflag := player.MarkBroadcast(false)
				player.AddCoin(int64(player.gainCoin), common.GainWay_HundredSceneWin, base.SyncFlag_ToClient, "system", s.GetSceneName())
				//player.MarkBroadcast(oldflag)
			} else {
				player.AddCoin(-(player.allBetCoin), common.GainWay_HundredSceneLost, base.SyncFlag_ToClient, "system", s.GetSceneName())
			}
			rp.ChangeCoin = (winCoin - player.taxCoin) + int64(returnCoin)
			//betTotal := int64(player.allBetCoin)
			//if betTotal != 0 {
			//	player.ReportGameEvent(int64(player.gainCoin)-betTotal, tax, betTotal)
			//}
			rp.AfterCoin = player.GetCoin()
			if !rp.IsRob {
				bankerLoseCoin += winCoin //庄家输掉的金币为玩家压中赢取的金币
				bankerWinCoin += loseCoin //庄家赢取的金币为玩家输掉的金币
				rollhundredtype.BetCoin += int64(player.allBetCoin)
				rollhundredtype.WinCoin += (winCoin - player.taxCoin) - loseCoin
				rollhundredtype.Person = append(rollhundredtype.Person, rp)
			}
			//统计游戏局数
			if winCoin != 0 || loseCoin != 0 {
				if winCoin-loseCoin > 0 {
					player.SetWinTimes(player.GetWinTimes() + 1)
				} else {
					player.SetLostTimes(player.GetLostTimes() + 1)
				}
			}
		}
		//统计系统产出的金币,系统输钱,更新奖池
		base.CoinPoolMgr.PushCoin(sceneEx.GetGameFreeId(), sceneEx.GetGroupId(), sceneEx.GetPlatform(), bankerWinCoin)
		base.CoinPoolMgr.PopCoin(sceneEx.GetGameFreeId(), sceneEx.GetGroupId(), sceneEx.GetPlatform(), bankerLoseCoin)
		//构建结算消息
		var bSaveDetailLog bool
		logid, _ := model.AutoIncGameLogId()
		for _, player := range sceneEx.players {
			betPlayerTotal := int64(0)
			for _, v := range player.CoinPool {
				betPlayerTotal += int64(v)
			}
			//统计投入产出
			inout := int64(player.winCoin) - betPlayerTotal
			if inout != 0 {
				//赔率统计
				player.Statics(sceneEx.GetKeyGameId(), sceneEx.GetKeyGameId(), inout, true)
			}
			if player.gainCoin > 0 {
				player.WinRecord = append(player.WinRecord, 1)
			} else {
				player.WinRecord = append(player.WinRecord, 0)
			}
			player.BetBigRecord = append(player.BetBigRecord, int(player.GetTotlePushCoin()))

			player.cGetWin20 = player.GetWin20()
			player.cGetBetGig20 = player.GetBetGig20()

			packBilled := &rollpoint.SCRollPointBill{
				Coin:   proto.Int64(int64(player.gainCoin)),
				Banker: proto.Int64(int64(bankerWinCoin - bankerLoseCoin)),
			}
			for _, value := range sceneEx.blackBox.Point {
				packBilled.Point = append(packBilled.Point, value)
			}
			for _, value := range sceneEx.blackBox.Score {
				packBilled.Flag = append(packBilled.Flag, value)
			}
			for key, value := range player.PushCoinLog {
				rpcl := &rollpoint.RollPointCoinLog{
					Index: proto.Int32(key),
				}
				for k, v := range value {
					rpcl.Coins = append(rpcl.Coins, k, v)
				}
				packBilled.Coins = append(packBilled.Coins, rpcl)
			}
			player.SendToClient(int(rollpoint.RPPACKETID_ROLLPOINT_SC_BILL), packBilled)
			logger.Logger.Trace("SCRollPointBill:", packBilled)
			betTotal := player.allBetCoin
			//统计金币变动
			if betTotal > 0 {
				if !player.IsRob {
					player.SaveSceneCoinLog(player.GetTakeCoin(), int64(player.gainCoin),
						player.GetCoin(), betTotal, player.taxCoin, player.winCoin, 0, 0)
					totalin, totalout := int64(0), int64(0)
					if player.GetCoin()-player.GetTakeCoin() > 0 {
						totalout = player.GetCoin() - player.GetTakeCoin() + player.taxCoin
					} else {
						totalin = -(player.GetCoin() - player.GetTakeCoin() + player.taxCoin)
					}
					validFlow := totalin + totalout
					validBet := common.AbsI64(totalin - totalout)
					sceneEx.SaveGamePlayerListLog(player.GetSnId(),
						base.GetSaveGamePlayerListLogParam(player.Platform, player.Channel, player.BeUnderAgentCode, player.PackageID, logid, player.InviterId, totalin, totalout,
							player.taxCoin, 0, player.allBetCoin, int64(player.gainCoin), validBet, validFlow, sceneEx.IsPlayerFirst(player.Player), false))
					bSaveDetailLog = true
				}
			}
		}

		//统计参与游戏次数
		if !sceneEx.GetTesting() {
			//var playerCtxs []*server.PlayerCtx
			//for _, p := range sceneEx.players {
			//	if p == nil || p.IsRob || p.GetTotlePushCoin() == 0 {
			//		continue
			//	}
			//	playerCtxs = append(playerCtxs, &server.PlayerCtx{SnId: proto.Int32(p.SnId), Coin: proto.Int64(p.GetCoin())})
			//}
			//if len(playerCtxs) > 0 {
			//	pack := &server.GWSceneEnd{
			//		GameFreeId: proto.Int32(sceneEx.GetDBGameFree().GetId()),
			//		Players:    playerCtxs,
			//	}
			//	proto.SetDefaults(pack)
			//	sceneEx.SendToWorld(int(server.MmoPacketID_PACKET_GW_SCENEEND), pack)
			//}
		}
		//统计下注数
		if !sceneEx.GetTesting() {
			gwPlayerBet := &server.GWPlayerBet{
				GameFreeId: proto.Int32(sceneEx.GetDBGameFree().GetId()),
				RobotGain:  proto.Int64(bankerWinCoin - bankerLoseCoin),
			}
			for _, p := range sceneEx.players {
				if p == nil || p.IsRob || p.allBetCoin == 0 {
					continue
				}
				playerBet := &server.PlayerBet{
					SnId: proto.Int32(p.SnId),
					Bet:  proto.Int64(int64(p.allBetCoin)),
					Gain: proto.Int64(p.winCoin - p.allBetCoin),
					Tax:  proto.Int64(p.taxCoin),
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
			info, err := model.MarshalGameNoteByHUNDRED(&rollhundredtype)
			if err == nil {
				//WinLog, _ := json.Marshal(sceneEx.WinLog)
				trends := [][3]int32{}
				if len(sceneEx.WinLog) > 0 {
					for _, value := range sceneEx.WinLog {
						trend := [3]int32{}
						for i := 0; i < 3; i++ { //0 1 2
							switch value >> uint(i*8) & 255 {
							case 0:
								trend[i] = 1
							case 1:
								trend[i] = 2
							case 2:
								trend[i] = 3
							case 3:
								trend[i] = 4
							case 4:
								trend[i] = 5
							case 5:
								trend[i] = 6
							}
						}
						trends = append(trends, trend)
					}
				}
				trend20Lately, _ := json.Marshal(trends)
				sceneEx.SaveGameDetailedLog(logid, info, &base.GameDetailedParam{
					Trend20Lately: string(trend20Lately),
				})
			}
		}
		for _, player := range sceneEx.players {
			if player != nil {
				player.allBetCoin = 0
			}
		}
	}
}

func (this *SceneRollPointStateBilled) OnTick(s *base.Scene) {
	this.SceneBaseStateRollPoint.OnTick(s)
	if sceneEx, ok := s.GetExtraData().(*RollPointSceneData); ok {
		if time.Now().Sub(sceneEx.StateStartTime) > RollPointBilledTimeout {
			for _, value := range sceneEx.players {
				value.GameTimes++
				if value.GetCoin() < int64(sceneEx.GetDBGameFree().GetBetLimit()) || value.GetCoin() < int64(sceneEx.GetDBGameFree().GetOtherIntParams()[0]) {
					value.MarkFlag(base.PlayerState_GameBreak)
					value.SyncFlag(true)
				} else {
					if value.IsMarkFlag(base.PlayerState_GameBreak) {
						value.UnmarkFlag(base.PlayerState_GameBreak)
						value.SyncFlag(true)
					}
				}
			}
			s.ChangeSceneState(RollPointSceneStateStart)
		}
	}
}
func (this *SceneRollPointStateBilled) OnLeave(s *base.Scene) {
	if sceneEx, ok := s.GetExtraData().(*RollPointSceneData); ok {
		sceneEx.Clean()
		for _, value := range sceneEx.players {
			value.UnmarkFlag(base.PlayerState_WaitNext)
			value.SyncFlag(true)
			if !value.IsOnLine() {
				sceneEx.PlayerLeave(value.Player, common.PlayerLeaveReason_DropLine, true)
			} else if value.IsRob {
				if s.CoinOverMaxLimit(value.GetCoin(), value.Player) {
					s.PlayerLeave(value.Player, common.PlayerLeaveReason_Normal, true)
				} else if value.Trusteeship >= 5 {
					sceneEx.PlayerLeave(value.Player, common.PlayerLeaveReason_LongTimeNoOp, true)
				}
			} else {
				if !s.CoinInLimit(value.GetCoin()) {
					sceneEx.PlayerLeave(value.Player, common.PlayerLeaveReason_Bekickout, true)
				} else if value.Trusteeship >= model.GameParamData.PlayerWatchNum {
					sceneEx.PlayerLeave(value.Player, common.PlayerLeaveReason_LongTimeNoOp, true)
				}
			}
		}
		if s.CheckNeedDestroy() {
			sceneEx.SceneDestroy(true)
		}
	}
}
func (this *SceneRollPointStateBilled) OnPlayerEvent(s *base.Scene, p *base.Player, evtcode int, params []int64) {
	this.SceneBaseStateRollPoint.OnPlayerEvent(s, p, evtcode, params)
	switch evtcode {
	case base.PlayerEventEnter:
		fallthrough
	case base.PlayerEventRehold:
		if playerEx, ok := p.GetExtraData().(*RollPointPlayerData); ok {
			playerEx.MarkFlag(base.PlayerState_WaitNext)
			playerEx.SyncFlag(true)
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
func (this *ScenePolicyRollPoint) RegisteSceneState(state base.SceneState) {
	if state == nil {
		return
	}
	stateid := state.GetState()
	if stateid < 0 || stateid >= RollPointSceneStateMax {
		return
	}
	this.states[stateid] = state
}
func (this *ScenePolicyRollPoint) GetSceneState(s *base.Scene, stateid int) base.SceneState {
	if stateid >= 0 && stateid < RollPointSceneStateMax {
		return ScenePolicyRollPointSington.states[stateid]
	}
	return nil
}
func init() {
	ScenePolicyRollPointSington.RegisteSceneState(&SceneRollPointStateWait{})
	ScenePolicyRollPointSington.RegisteSceneState(&SceneRollPointStateStart{})
	ScenePolicyRollPointSington.RegisteSceneState(&SceneRollPointStateRoll{})
	ScenePolicyRollPointSington.RegisteSceneState(&SceneRollPointStateOpen{})
	ScenePolicyRollPointSington.RegisteSceneState(&SceneRollPointStateBilled{})
	core.RegisteHook(core.HOOK_BEFORE_START, func() error {
		base.RegisteScenePolicy(common.GameId_RollPoint, 0, ScenePolicyRollPointSington)
		return nil
	})
}
