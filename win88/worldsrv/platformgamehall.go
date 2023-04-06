package main

import (
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/gamehall"
	"games.yol.com/win88/protocol/server"
	"games.yol.com/win88/srvdata"
	"github.com/idealeak/goserver/core/logger"
)

type PlatformGameHall struct {
	HallId     int32               //游戏厅id
	Scenes     map[int]*Scene      //游戏房间列表
	Players    map[int32]*Player   //大厅中的玩家
	p          *Platform           //所属平台
	dbGameFree *server.DB_GameFree //厅配置数据(这里都是模板值,非后台实例数据)
	dbGameRule *server.DB_GameRule //游戏配置数据(这里都是模板值,非后台实例数据)
}

func NewPlatformGameHall(p *Platform, hallId int32) *PlatformGameHall {
	pgh := &PlatformGameHall{
		p:       p,
		HallId:  hallId,
		Scenes:  make(map[int]*Scene),
		Players: make(map[int32]*Player),
	}
	if pgh.Init() {
		return pgh
	}
	return nil
}

func (pgh *PlatformGameHall) Init() bool {
	pgh.dbGameFree = srvdata.PBDB_GameFreeMgr.GetData(pgh.HallId)
	if pgh.dbGameFree == nil {
		return false
	}
	pgh.dbGameRule = srvdata.PBDB_GameRuleMgr.GetData(pgh.dbGameFree.GetGameRule())
	if pgh.dbGameRule == nil {
		return false
	}
	return true
}

func (pgh *PlatformGameHall) OpenSceneToPublic() {
	//开放所有房间
	var scenes []*Scene
	publics := PlatformMgrSington.GetOrCreateScenesByHall(pgh.HallId)
	if publics != nil {
		for sceneid, scene := range pgh.Scenes {
			publics[sceneid] = scene
			scene.limitPlatform = nil
			scenes = append(scenes, scene)
		}

		pgh.Scenes = make(map[int]*Scene)
		players := PlatformMgrSington.GetOrCreatePlayersByHall(pgh.HallId)
		if players != nil {
			PlatformMgrSington.BroadcastRoomList(pgh.HallId, pgh.dbGameRule, scenes, true, players, 0)
			for snid, player := range pgh.Players {
				players[snid] = player
			}
		}

		// 广播公共房间列表
		scenes = scenes[0:0]
		for _, s := range publics {
			scenes = append(scenes, s)
		}
		PlatformMgrSington.BroadcastRoomList(pgh.HallId, pgh.dbGameRule, scenes, false, pgh.Players, 0)
		pgh.Players = make(map[int32]*Player)
	}
}

func (pgh *PlatformGameHall) ConvertToIsolated() {
	players := PlatformMgrSington.GetOrCreatePlayersByHall(pgh.HallId)
	for snid, player := range players {
		if player.Platform == pgh.p.IdStr {
			pgh.Players[snid] = player
		}
	}

	//创建私有房间列表
	sp := GetScenePolicy(int(pgh.dbGameFree.GetGameId()), int(pgh.dbGameFree.GetGameMode()))
	if spd, ok := sp.(*ScenePolicyData); ok {
		playernum := spd.getPlayerNum(pgh.dbGameRule.GetParams())
		inc := evaluateSceneIncCount(len(pgh.Scenes), len(pgh.Players), int(playernum))
		if inc > 0 {
			var scenes []*Scene
			for i := 0; i < inc; i++ {
				scene := PlatformMgrSington.CreateNewScene(pgh.p, pgh.HallId)
				if scene != nil {
					pgh.Scenes[scene.sceneId] = scene
					scenes = append(scenes, scene)
				}
			}
			PlatformMgrSington.BroadcastRoomList(pgh.HallId, pgh.dbGameRule, scenes, true, pgh.Players, 0)
		}
	}
}

func (pgh *PlatformGameHall) PlayerEnter(p *Player) {
	pgh.Players[p.SnId] = p
	p.hallId = pgh.HallId
	sp := GetScenePolicy(int(pgh.dbGameFree.GetGameId()), int(pgh.dbGameFree.GetGameMode()))
	if spd, ok := sp.(*ScenePolicyData); ok {
		playernum := spd.getPlayerNum(pgh.dbGameRule.GetParams())
		inc := evaluateSceneIncCount(len(pgh.Scenes), len(pgh.Players), int(playernum))
		if inc > 0 {
			var scenes []*Scene
			for i := 0; i < inc; i++ {
				logger.Logger.Trace("PlayerEnter::CreateNewScene")
				scene := PlatformMgrSington.CreateNewScene(pgh.p, pgh.HallId)
				if scene != nil {
					pgh.Scenes[scene.sceneId] = scene
					scenes = append(scenes, scene)
				}
			}
			PlatformMgrSington.BroadcastRoomList(pgh.HallId, pgh.dbGameRule, scenes, true, pgh.Players, p.SnId)
		}
		//发送房间列表
		pack := &gamehall.SCHallRoomList{
			HallId:   proto.Int32(pgh.HallId),
			GameId:   proto.Int32(pgh.dbGameRule.GetGameId()),
			GameMode: proto.Int32(pgh.dbGameRule.GetGameMode()),
			IsAdd:    proto.Bool(false),
			Params:   pgh.dbGameRule.GetParams(),
		}
		for _, scene := range pgh.Scenes {
			ri := &gamehall.RoomInfo{
				RoomId:   proto.Int(scene.sceneId),
				Starting: proto.Bool(scene.starting),
			}
			for _, p := range scene.players {
				ri.Players = append(ri.Players, p.CreateRoomPlayerInfoProtocol())
			}
			pack.Rooms = append(pack.Rooms, ri)
		}
		proto.SetDefaults(pack)
		p.SendToClient(int(gamehall.GameHallPacketID_PACKET_SC_HALLROOMLIST), pack)
		//发送人数信息
		pgh.p.SendPlayerNum(pgh.dbGameFree.GetGameId(), p)
	}
}

func (pgh *PlatformGameHall) PlayerLeave(p *Player) {
	delete(pgh.Players, p.SnId)
	if p.hallId == pgh.HallId {
		p.hallId = 0
	}
}

func (p *Player) CreateRoomPlayerInfoProtocol() *gamehall.RoomPlayerInfo {
	pack := &gamehall.RoomPlayerInfo{
		SnId:        proto.Int32(p.SnId),
		Head:        proto.Int32(p.Head),
		Sex:         proto.Int32(p.Sex),
		Name:        proto.String(p.Name),
		Pos:         proto.Int(p.pos),
		Flag:        proto.Int32(p.flag),
		HeadOutLine: proto.Int32(p.HeadOutLine),
		VIP:         proto.Int32(p.VIP),
	}
	return pack
}

func (pgh *PlatformGameHall) OnPlayerEnterScene(scene *Scene, player *Player) {
	delete(pgh.Players, player.SnId)
	pack := &gamehall.SCRoomPlayerEnter{
		RoomId: proto.Int(scene.sceneId),
		Player: player.CreateRoomPlayerInfoProtocol(),
	}
	proto.SetDefaults(pack)
	pgh.Broadcast(int(gamehall.GameHallPacketID_PACKET_SC_ROOMPLAYERENTER), pack, player.SnId)
}

func (pgh *PlatformGameHall) OnPlayerLeaveScene(scene *Scene, player *Player) {
	pack := &gamehall.SCRoomPlayerLeave{
		RoomId: proto.Int(scene.sceneId),
		Pos:    proto.Int(player.pos),
	}
	proto.SetDefaults(pack)
	pgh.Broadcast(int(gamehall.GameHallPacketID_PACKET_SC_ROOMPLAYERLEAVE), pack, player.SnId)
}

func (pgh *PlatformGameHall) OnDestroyScene(scene *Scene) {
	delete(pgh.Scenes, scene.sceneId)
	pack := &gamehall.SCDestroyRoom{
		RoomId:    proto.Int(scene.sceneId),
		OpRetCode: gamehall.OpResultCode_Game_OPRC_Sucess_Game,
		IsForce:   proto.Int(1),
	}
	proto.SetDefaults(pack)
	pgh.Broadcast(int(gamehall.GameHallPacketID_PACKET_SC_DESTROYROOM), pack, 0)
}

func (pgh *PlatformGameHall) Broadcast(packetid int, packet interface{}, exclude int32) {
	PlatformMgrSington.Broadcast(packetid, packet, pgh.Players, exclude)
}
