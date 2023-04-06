package main

import (
	"fmt"
	"games.yol.com/win88/protocol/webapi"
	"sync"

	"games.yol.com/win88/common"
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	gamehall_proto "games.yol.com/win88/protocol/gamehall"
	server_proto "games.yol.com/win88/protocol/server"
	"games.yol.com/win88/srvdata"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/srvlib"
	libproto "github.com/idealeak/goserver/srvlib/protocol"
)

type GameSessionListener interface {
	OnGameSessionRegiste(*GameSession)
	OnGameSessionUnregiste(*GameSession)
}

var GameSessionListenerSet = new(sync.Map)

func RegisteGameSessionListener(listener GameSessionListener) {
	GameSessionListenerSet.Store(listener, listener)
}

func UnregisteGameSessionListener(listener GameSessionListener) {
	GameSessionListenerSet.Delete(listener)
}

type GameSession struct {
	*netlib.Session
	srvId   int
	srvType int
	state   common.GameSessState
	players map[int32]*Player
	scenes  map[int]*Scene
	cps     map[string]*model.CoinPoolSetting
	gameIds []int32
}

// 构造函数
func NewGameSession(srvId, srvType int, s *netlib.Session) *GameSession {
	gs := &GameSession{
		Session: s,
		srvId:   srvId,
		srvType: srvType,
		state:   common.GAME_SESS_STATE_ON,
		players: make(map[int32]*Player),
		scenes:  make(map[int]*Scene),
		cps:     make(map[string]*model.CoinPoolSetting),
	}
	return gs
}

func (this *GameSession) RebindPlayerSnId(oldSnId, newSnId int32) {
	if p, exist := this.players[oldSnId]; exist {
		delete(this.players, oldSnId)
		this.players[newSnId] = p
	}
}

func (this *GameSession) GetSrvId() int32 {
	if this.Session == nil {
		return 0
	}
	attr := this.GetAttribute(srvlib.SessionAttributeServerInfo)
	if attr != nil {
		if srvInfo, ok := attr.(*libproto.SSSrvRegiste); ok && srvInfo != nil {
			return srvInfo.GetId()
		}
	}
	return 0
}

// 关闭其上的所有场景
func (this *GameSession) CloseAllScene() {
	for sceneId, scene := range this.scenes {
		if scene.IsMiniGameScene() {
			SceneMgrSington.DestroyMiniGameScene(sceneId)
		} else {
			scDestroyRoom := &gamehall_proto.SCDestroyRoom{
				RoomId:    proto.Int(sceneId),
				OpRetCode: gamehall_proto.OpResultCode_Game_OPRC_Sucess_Game,
				IsForce:   proto.Int(1),
			}
			proto.SetDefaults(scDestroyRoom)
			scene.Broadcast(int(gamehall_proto.GameHallPacketID_PACKET_SC_DESTROYROOM), scDestroyRoom, 0)
			SceneMgrSington.DestroyScene(sceneId, true)
		}
	}
	this.scenes = nil
	this.players = nil
}

// 注册事件
func (this *GameSession) OnRegiste() {
	GameSessionListenerSet.Range(func(key, val interface{}) bool {
		if lis, ok := val.(GameSessionListener); ok {
			lis.OnGameSessionRegiste(this)
		}
		return true
	})
}

// 注销事件
func (this *GameSession) OnUnregiste() {
	//销毁比赛
	//MatchMgrSington.DestroyAllMatchByGameSession(this)
	//解散房间
	this.CloseAllScene()

	GameSessionListenerSet.Range(func(key, val interface{}) bool {
		if lis, ok := val.(GameSessionListener); ok {
			lis.OnGameSessionUnregiste(this)
		}
		return true
	})
}

// 负载因数
func (this *GameSession) GetLoadFactor() int {
	return len(this.scenes)*20 + len(this.players)
}

// 设置状态
func (this *GameSession) SwitchState(state common.GameSessState) {
	if state == this.state {
		return
	}
	this.state = state
	switch state {
	case common.GAME_SESS_STATE_ON:
		this.OnStateOn()
	case common.GAME_SESS_STATE_OFF:
		this.OnStateOff()
	}
}
func (this *GameSession) OnStateOn() {
	pack := &server_proto.ServerState{
		SrvState: proto.Int(int(this.state)),
	}
	proto.SetDefaults(pack)
	this.Send(int(server_proto.SSPacketID_PACKET_WG_SERVER_STATE), pack)
}
func (this *GameSession) OnStateOff() {
	pack := &server_proto.ServerState{
		SrvState: proto.Int(int(this.state)),
	}
	proto.SetDefaults(pack)
	this.Send(int(server_proto.SSPacketID_PACKET_WG_SERVER_STATE), pack)
}

func (this *GameSession) AddScene(s *Scene) {
	this.scenes[s.sceneId] = s
	//send msg
	msg := &server_proto.WGCreateScene{
		SceneId:      proto.Int(s.sceneId),
		GameId:       proto.Int(s.gameId),
		GameMode:     proto.Int(s.gameMode),
		SceneMode:    proto.Int(s.sceneMode),
		Params:       s.params,
		ParamsEx:     s.paramsEx,
		Creator:      proto.Int32(s.creator),
		Agentor:      proto.Int32(s.agentor),
		HallId:       proto.Int32(s.hallId),
		ReplayCode:   proto.String(s.replayCode),
		GroupId:      proto.Int32(s.groupId),
		TotalOfGames: proto.Int32(s.totalRound),
		BaseScore:    proto.Int32(s.BaseScore),
		PlayerNum:    proto.Int(s.playerNum),
	}
	var platform *Platform
	if s.limitPlatform != nil {
		msg.Platform = proto.String(s.limitPlatform.IdStr)
		platform = s.limitPlatform
	} else {
		msg.Platform = proto.String(Default_Platform)
		platform = PlatformMgrSington.GetPlatform(Default_Platform)
	}
	if s.dbGameFree != nil {
		msg.DBGameFree = s.dbGameFree
	} else if platform != nil {
		gps := PlatformMgrSington.GetGameConfig(platform.IdStr, s.paramsEx[0])
		if gps != nil {
			if gps.GroupId == 0 {
				msg.DBGameFree = gps.DbGameFree
			} else {
				pgg := PlatformGameGroupMgrSington.GetGameGroup(gps.GroupId)
				if pgg != nil {
					msg.DBGameFree = pgg.DbGameFree
				}
			}
		}
	}
	if s.IsCoinScene() {
		if sp, ok := s.sp.(*ScenePolicyData); ok {
			msg.EnterAfterStart = proto.Bool(sp.EnterAfterStart)
		}
	}
	if s.ClubId > 0 {
		msg.Club = proto.Int32(s.ClubId)
		msg.ClubRoomId = proto.String(s.clubRoomID)
		msg.ClubRoomPos = proto.Int32(s.clubRoomPos)
		msg.ClubRate = proto.Int32(s.clubRoomTax)
	}
	proto.SetDefaults(msg)
	this.Send(int(server_proto.SSPacketID_PACKET_WG_CREATESCENE), msg)
	logger.Logger.Trace("WGCreateScene:", msg)
}

func (this *GameSession) DelScene(s *Scene) {
	delete(this.scenes, s.sceneId)
	//from gameserver, so don't need send msg
}

func (this *GameSession) AddPlayer(p *Player) {
	this.players[p.SnId] = p
}

func (this *GameSession) DelPlayer(p *Player) {
	delete(this.players, p.SnId)
}

func (this *GameSession) GenCoinPoolSettingKey(platform string, groupId, gamefreeid, srvid int32) string {
	var key string
	if groupId != 0 {
		key = fmt.Sprintf("%v+%v_%v", gamefreeid, groupId, srvid)
	} else {
		key = fmt.Sprintf("%v_%v_%v", gamefreeid, platform, srvid)
	}
	return key
}

func (this *GameSession) DetectCoinPoolSetting(platform string, gamefreeid, groupId int32) bool {
	srvid := this.GetSrvId()
	key := this.GenCoinPoolSettingKey(platform, groupId, gamefreeid, srvid)
	if _, exist := this.cps[key]; !exist {
		data := model.GetCoinPoolSetting(gamefreeid, srvid, groupId, platform)
		if data == nil {
			dbGameCoinPool := srvdata.PBDB_GameCoinPoolMgr.GetData(gamefreeid)
			if dbGameCoinPool != nil {
				data = model.NewCoinPoolSetting(platform, groupId, gamefreeid, srvid, dbGameCoinPool.GetInitValue(), dbGameCoinPool.GetLowerLimit(), dbGameCoinPool.GetUpperLimit(), dbGameCoinPool.GetUpperOffsetLimit(), dbGameCoinPool.GetMaxOutValue(), dbGameCoinPool.GetChangeRate(), dbGameCoinPool.GetMinOutPlayerNum(), dbGameCoinPool.GetUpperLimitOfOdds(),
					dbGameCoinPool.GetBaseRate(), dbGameCoinPool.GetCtroRate(), dbGameCoinPool.GetHardTimeMin(), dbGameCoinPool.GetHardTimeMax(), dbGameCoinPool.GetNormalTimeMin(), dbGameCoinPool.GetNormalTimeMax(), dbGameCoinPool.GetEasyTimeMin(), dbGameCoinPool.GetEasyTimeMax(), dbGameCoinPool.GetEasrierTimeMin(), dbGameCoinPool.GetEasrierTimeMax(), dbGameCoinPool.GetProfitRate(), 0)
				if data != nil {
					err := model.UpsertCoinPoolSetting(data, nil)
					if err == nil {
						model.ManageCoinPoolSetting(data)
					}
				}
			}
		}
		if data != nil {
			this.cps[key] = data
			//send msg
			msg := &webapi.CoinPoolSetting{
				Platform:         proto.String(platform),
				GameFreeId:       proto.Int32(gamefreeid),
				ServerId:         proto.Int32(srvid),
				GroupId:          proto.Int32(groupId),
				InitValue:        proto.Int32(data.InitValue),
				LowerLimit:       proto.Int32(data.LowerLimit),
				UpperLimit:       proto.Int32(data.UpperLimit),
				UpperOffsetLimit: proto.Int32(data.UpperOffsetLimit),
				MaxOutValue:      proto.Int32(data.MaxOutValue),
				ChangeRate:       proto.Int32(data.ChangeRate),
				MinOutPlayerNum:  proto.Int32(data.MinOutPlayerNum),
				UpperLimitOfOdds: proto.Int32(data.UpperLimitOfOdds),
				BaseRate:         proto.Int32(data.BaseRate),
				CtroRate:         proto.Int32(data.CtroRate),
				HardTimeMin:      proto.Int32(data.HardTimeMin),
				HardTimeMax:      proto.Int32(data.HardTimeMax),
				NormalTimeMin:    proto.Int32(data.NormalTimeMin),
				NormalTimeMax:    proto.Int32(data.NormalTimeMax),
				EasyTimeMin:      proto.Int32(data.EasyTimeMin),
				EasyTimeMax:      proto.Int32(data.EasyTimeMax),
				EasrierTimeMin:   proto.Int32(data.EasrierTimeMin),
				EasrierTimeMax:   proto.Int32(data.EasrierTimeMax),
				CpCangeType:      proto.Int32(data.CpCangeType),
				CpChangeInterval: proto.Int32(data.CpChangeInterval),
				CpChangeTotle:    proto.Int32(data.CpChangeTotle),
				CpChangeLower:    proto.Int32(data.CpChangeLower),
				CpChangeUpper:    proto.Int32(data.CpChangeUpper),
				ProfitRate:       proto.Int32(data.ProfitRate),
				CoinPoolMode:     proto.Int32(data.CoinPoolMode),
				ResetTime:        proto.Int32(data.ResetTime),
			}
			//ProfitControlMgrSington.FillCoinPoolSetting(msg)
			proto.SetDefaults(msg)
			this.Send(int(server_proto.SSPacketID_PACKET_WG_COINPOOLSETTING), msg)
		}
	}
	return true
}
