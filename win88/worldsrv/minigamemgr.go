package main

//
//import (
//	"games.yol.com/win88/common"
//	"games.yol.com/win88/proto"
//	"games.yol.com/win88/protocol/mngame"
//	server_proto "games.yol.com/win88/protocol/server"
//	webapi_proto "games.yol.com/win88/protocol/webapi"
//	"games.yol.com/win88/srvdata"
//	"github.com/idealeak/goserver/core/logger"
//	"github.com/idealeak/goserver/srvlib"
//	srvlibproto "github.com/idealeak/goserver/srvlib/protocol"
//)
//
//var MiniGameMgrSington = &MiniGameMgr{
//	//按平台管理
//	scenesOfPlatform: make(map[string]map[int32]*Scene),
//	//玩家当前打开的小游戏列表
//	playerGaming: make(map[int32]map[int32]*Scene),
//	autoId:       common.MiniGameSceneStartId,
//}
//
//type MiniGameMgr struct {
//	BasePlayerListener
//	//按平台管理
//	scenesOfPlatform map[string]map[int32]*Scene
//	//玩家当前打开的小游戏列表
//	playerGaming map[int32]map[int32]*Scene
//	autoId       int
//}
//
//func (this *MiniGameMgr) GenOneSceneId() int {
//	this.autoId++
//	if this.autoId > common.MiniGameSceneMaxId {
//		this.autoId = common.MiniGameSceneStartId
//	}
//	return this.autoId
//}
//
//func (this *MiniGameMgr) PlayerEnter(p *Player, id int32) mngame.MNGameOpResultCode {
//	plt := p.GetPlatform()
//	s := this.GetScene(plt, id)
//	if s == nil {
//		return mngame.MNGameOpResultCode_MNGAME_OPRC_Error
//	}
//
//	if !s.PlayerEnterMiniGame(p) {
//		return mngame.MNGameOpResultCode_MNGAME_OPRC_Error
//	}
//
//	gamings, ok := this.playerGaming[p.SnId]
//	if !ok {
//		gamings = make(map[int32]*Scene)
//		this.playerGaming[p.SnId] = gamings
//	}
//	gamings[id] = s
//
//	return mngame.MNGameOpResultCode_MNGAME_OPRC_Sucess
//}
//
//func (this *MiniGameMgr) PlayerLeave(p *Player, id int32) mngame.MNGameOpResultCode {
//	plt := p.GetPlatform()
//	s := this.GetScene(plt, id)
//	if s == nil {
//		return mngame.MNGameOpResultCode_MNGAME_OPRC_Error
//	}
//
//	if !s.PlayerLeaveMiniGame(p) {
//		return mngame.MNGameOpResultCode_MNGAME_OPRC_Error
//	}
//
//	gamings, ok := this.playerGaming[p.SnId]
//	if ok {
//		delete(gamings, id)
//	}
//
//	return mngame.MNGameOpResultCode_MNGAME_OPRC_Sucess
//}
//
//func (this *MiniGameMgr) PlayerMsgDispatcher(p *Player, msg *mngame.CSMNGameDispatcher) {
//	plt := p.GetPlatform()
//	s := this.GetScene(plt, msg.GetId())
//	if s == nil {
//		logger.Logger.Errorf("MiniGameMgr.PlayerMsgDispatcher Can't find scene! plt:%v gameId:%v", plt, msg.GetId())
//		return
//	}
//
//	//minigamesrv 重启容错
//	if !s.HasPlayer(p) {
//		this.PlayerEnter(p, msg.GetId())
//	}
//	s.RedirectMiniGameMsg(p, msg)
//}
//
//func (this *MiniGameMgr) GetScene(p *Platform, id int32) *Scene {
//	scenes, ok := this.scenesOfPlatform[p.IdStr]
//	if !ok {
//		scenes = make(map[int32]*Scene)
//		this.scenesOfPlatform[p.IdStr] = scenes
//	}
//
//	s, ok := scenes[id]
//	if !ok {
//		cfg := PlatformMgrSington.GetGameConfig(p.IdStr, id)
//		if cfg != nil && cfg.Status && cfg.DbGameFree.GetGameType() == common.GameType_Mini {
//			s = this.CreateSceneByPlatform(p, cfg)
//			if s != nil {
//				scenes[cfg.DbGameFree.Id] = s
//			} else {
//				return nil
//			}
//			return s
//		} else {
//			return nil
//		}
//	} else {
//		return s
//	}
//	//return nil
//}
//
//func (this *MiniGameMgr) CreateSceneByPlatform(p *Platform, cfg *webapi_proto.GameFree) *Scene {
//	sceneId := this.GenOneSceneId()
//	gameId := int(cfg.DbGameFree.GetGameId())
//	gs := GameSessMgrSington.GetMinLoadSess(gameId)
//	if gs == nil {
//		logger.Logger.Errorf("MiniGameMgr.CreateSceneByPlatform Get %v game min session failed.", gameId)
//		return nil
//	}
//	if gs != nil {
//		gameMode := cfg.DbGameFree.GetGameMode()
//		dbGameRule := srvdata.PBDB_GameRuleMgr.GetData(cfg.DbGameFree.GetGameRule())
//		params := dbGameRule.GetParams()
//		scene := SceneMgrSington.CreateScene(0, 0, sceneId, gameId, int(gameMode), common.SceneMode_Public, 1, -1, params,
//			gs, p, cfg.GroupId, cfg.DbGameFree, cfg.DbGameFree.Id)
//		if scene != nil {
//			scene.hallId = cfg.DbGameFree.Id
//			return scene
//		}
//	}
//	return nil
//}
//
//func (this *MiniGameMgr) OnPlatformCreate(p *Platform) {
//	if p == nil {
//		return
//	}
//	scenes := make(map[int32]*Scene)
//	this.scenesOfPlatform[p.IdStr] = scenes
//
//	gps := PlatformMgrSington.GetPlatformGameConfig(p.IdStr)
//	for _, v := range gps {
//		if v.Status && v.DbGameFree.GetGameType() == common.GameType_Mini {
//			s := this.CreateSceneByPlatform(p, v)
//			if s != nil {
//				scenes[v.DbGameFree.Id] = s
//			}
//		}
//	}
//}
//
//func (this *MiniGameMgr) OnPlatformDestroy(p *Platform) {
//	if p == nil {
//		return
//	}
//	if scenes, ok := this.scenesOfPlatform[p.IdStr]; ok {
//		for _, s := range scenes {
//			pack := &server_proto.WGGraceDestroyScene{}
//			pack.Ids = append(pack.Ids, int32(s.sceneId))
//			s.SendToGame(int(server_proto.SSPacketID_PACKET_WG_GRACE_DESTROYSCENE), pack)
//		}
//		delete(this.scenesOfPlatform, p.IdStr)
//	}
//}
//
//func (this *MiniGameMgr) OnPlatformChangeIsolated(p *Platform, isolated bool) {
//	if p == nil {
//		return
//	}
//	if !isolated {
//		this.OnPlatformDestroy(p)
//	}
//}
//
//func (this *MiniGameMgr) OnPlatformChangeDisabled(p *Platform, disabled bool) {
//	if p == nil {
//		return
//	}
//	if disabled {
//		this.OnPlatformDestroy(p)
//	} else {
//		this.OnPlatformCreate(p)
//	}
//}
//
//func (this *MiniGameMgr) OnPlatformConfigUpdate(p *Platform, oldCfg, newCfg *webapi_proto.GameFree) {
//	if p == nil {
//		return
//	}
//	if scenes, ok := this.scenesOfPlatform[p.IdStr]; ok {
//		if oldCfg != nil {
//			if s, ok := scenes[oldCfg.DbGameFree.Id]; ok {
//				pack := &server_proto.WGGraceDestroyScene{}
//				pack.Ids = append(pack.Ids, int32(s.sceneId))
//				s.SendToGame(int(server_proto.SSPacketID_PACKET_WG_GRACE_DESTROYSCENE), pack)
//				delete(scenes, oldCfg.DbGameFree.Id)
//			}
//		} else if newCfg != nil {
//			if newCfg.Status && newCfg.DbGameFree.GetGameType() == common.GameType_Mini {
//				s := this.CreateSceneByPlatform(p, newCfg)
//				if s != nil {
//					scenes[newCfg.DbGameFree.Id] = s
//				}
//			}
//		}
//	}
//}
//
//func (this *MiniGameMgr) OnGameGroupUpdate(oldCfg, newCfg *webapi_proto.GameConfigGroup) {
//	//donothing
//}
//
///*
//获取platform下面对应的 player SnId所在的scene
//*/
//func (this *MiniGameMgr) GetAllSceneByPlayer(p *Player) map[int32]*Scene {
//	if gameingScenes, ok := this.playerGaming[p.SnId]; ok {
//		return gameingScenes
//	}
//	return nil
//}
//
//func (this *MiniGameMgr) OnPlayerDropLine(p *Player) {
//	this.BasePlayerListener.OnPlayerDropLine(p)
//	if gamingScenes, ok := this.playerGaming[p.SnId]; ok {
//		for _, s := range gamingScenes {
//			pack := &server_proto.WGPlayerDropLine{
//				Id:      proto.Int32(p.SnId),
//				SceneId: proto.Int(s.sceneId),
//			}
//			proto.SetDefaults(pack)
//			s.SendToGame(int(server_proto.SSPacketID_PACKET_WG_PLAYERDROPLINE), pack)
//		}
//	}
//}
//
//func (this *MiniGameMgr) OnPlayerRehold(p *Player) {
//	this.BasePlayerListener.OnPlayerRehold(p)
//	var gateSid int64
//	if p.gateSess != nil {
//		if srvInfo, ok := p.gateSess.GetAttribute(srvlib.SessionAttributeServerInfo).(*srvlibproto.SSSrvRegiste); ok && srvInfo != nil {
//			sessionId := srvlib.NewSessionIdEx(srvInfo.GetAreaId(), srvInfo.GetType(), srvInfo.GetId(), 0)
//			gateSid = sessionId.Get()
//		}
//	}
//	if gamingScenes, ok := this.playerGaming[p.SnId]; ok {
//		for _, s := range gamingScenes {
//			pack := &server_proto.WGPlayerRehold{
//				Id:      proto.Int32(p.SnId),
//				Sid:     proto.Int64(p.sid),
//				SceneId: proto.Int(s.sceneId),
//				GateSid: proto.Int64(gateSid),
//			}
//			proto.SetDefaults(pack)
//			s.SendToGame(int(server_proto.SSPacketID_PACKET_WG_PLAYERREHOLD), pack)
//		}
//	}
//}
//func (this *MiniGameMgr) OnPlayerReturnScene(p *Player) {
//	this.BasePlayerListener.OnPlayerReturnScene(p)
//	if gameingScenes, ok := this.playerGaming[p.SnId]; ok {
//		for _, s := range gameingScenes {
//			pack := &server_proto.WGPlayerReturn{
//				PlayerId: p.SnId,
//				RoomId:   int32(s.sceneId),
//			}
//			proto.SetDefaults(pack)
//			s.SendToGame(int(server_proto.SSPacketID_PACKET_WG_PLAYERRETURN), pack)
//		}
//	}
//}
//
//func (this *MiniGameMgr) OnDestroyScene(s *Scene) {
//
//	if pltScenes, ok := this.scenesOfPlatform[s.limitPlatform.IdStr]; ok {
//		delete(pltScenes, s.dbGameFree.Id)
//	}
//
//	for snid, _ := range s.players {
//		if scenes, ok := this.playerGaming[snid]; ok {
//			delete(scenes, s.dbGameFree.Id)
//			if len(scenes) == 0 {
//				delete(this.playerGaming, snid)
//			}
//		}
//	}
//}
//
//func (this *MiniGameMgr) ClrPlayerWhiteBlackState(p *Player) {
//	if gamings, ok := this.playerGaming[p.SnId]; ok {
//		for _, s := range gamings {
//			pack := &server_proto.WGSetPlayerBlackLevel{
//				SnId:           proto.Int32(p.SnId),
//				SceneId:        proto.Int32(int32(s.sceneId)),
//				ResetTotalCoin: proto.Bool(true),
//			}
//			proto.SetDefaults(pack)
//			s.SendToGame(int(server_proto.SSPacketID_PACKET_GW_AUTORELIEVEWBLEVEL), pack)
//		}
//	}
//}
//
//func (this *MiniGameMgr) OnPlatformDestroyByGameFreeId(p *Platform, gameFreeId int32) {
//	if p == nil {
//		return
//	}
//	if scenes, ok := this.scenesOfPlatform[p.IdStr]; ok {
//		for _, s := range scenes {
//			if s.dbGameFree.Id == gameFreeId {
//				pack := &server_proto.WGGraceDestroyScene{}
//				pack.Ids = append(pack.Ids, int32(s.sceneId))
//				s.SendToGame(int(server_proto.SSPacketID_PACKET_WG_GRACE_DESTROYSCENE), pack)
//				delete(scenes, gameFreeId)
//			}
//		}
//	}
//}
//
//func init() {
//	RegistePlayerListener(MiniGameMgrSington)
//	PlatformMgrSington.RegisteObserver(MiniGameMgrSington)
//	PlatformGameGroupMgrSington.RegisteObserver(MiniGameMgrSington)
//}
