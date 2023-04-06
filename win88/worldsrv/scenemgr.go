package main

import (
	"time"

	webapi2 "games.yol.com/win88/protocol/webapi"

	server_proto "games.yol.com/win88/protocol/server"
	"games.yol.com/win88/srvdata"

	"games.yol.com/win88/common"
	"games.yol.com/win88/model"

	"math"

	"games.yol.com/win88/webapi"
	"github.com/idealeak/goserver/core/logger"
)

type RoomInfo struct {
	Platform   string
	SceneId    int
	GameId     int
	GameMode   int
	SceneMode  int
	GroupId    int32
	GameFreeId int32
	SrvId      int32
	Creator    int32
	Agentor    int32
	ReplayCode string
	Params     []int32
	PlayerIds  []int32
	PlayerCnt  int
	RobotCnt   int
	Start      int
	CreateTime time.Time
	ClubId     int32
}

// 场景管理器
var SceneMgrSington = &SceneMgr{
	scenes:             make(map[int]*Scene),
	playerInScene:      make(map[int32]*Scene),
	autoId:             common.MatchSceneStartId,
	coinSceneAutoId:    common.CoinSceneStartId,
	hundredSceneAutoId: common.HundredSceneStartId,
	hallSceneAutoId:    common.HallSceneStartId,
	psIdGen:            common.NewRandDistinctId(common.PrivateSceneStartId, common.PrivateSceneMaxId),
}

type SceneMgr struct {
	BaseClockSinker
	scenes             map[int]*Scene
	playerInScene      map[int32]*Scene
	autoId             int
	coinSceneAutoId    int
	hundredSceneAutoId int
	hallSceneAutoId    int
	psIdGen            *common.RandDistinctId //private scene id生成器
}

func (this *SceneMgr) AllocReplayCode() string {
	code, _ := model.GetOneReplayId()
	return code
}

func (this *SceneMgr) CreateScene(agentor, creator int32, sceneId, gameId, gameMode, sceneMode int, clycleTimes int32,
	numOfGames int32, params []int32, gs *GameSession, limitPlatform *Platform, groupId int32, dbGameFree *server_proto.DB_GameFree,
	paramsEx ...int32) *Scene {
	logger.Logger.Trace("(this *SceneMgr) CreateScene ")
	s := NewScene(agentor, creator, sceneId, gameId, gameMode, sceneMode, clycleTimes, numOfGames, params, gs, limitPlatform, groupId,
		dbGameFree, paramsEx...)
	if s == nil {
		return nil
	}
	this.scenes[sceneId] = s

	if !s.IsMatchScene() && dbGameFree != nil {
		//平台水池设置
		gs.DetectCoinPoolSetting(limitPlatform.IdStr, dbGameFree.GetId(), s.groupId)
	}
	gs.AddScene(s)
	var platformName string
	if limitPlatform != nil {
		platformName = limitPlatform.IdStr
	}
	logger.Logger.Infof("(this *SceneMgr) CreateScene (gameId=%v, mode=%v), SceneId=%v groupid=%v platform=%v",
		gameId, gameMode, sceneId, groupId, platformName)
	return s
}

func (this *SceneMgr) CreateLocalGameScene(creator int32, sceneId, gameId, gameSite, sceneMode int, clycleTimes int32, params []int32, gs *GameSession, limitPlatform *Platform, playerNum int, dbGameFree *server_proto.DB_GameFree, baseScore int32,
	paramsEx ...int32) *Scene {
	logger.Logger.Trace("(this *SceneMgr) CreateLocalGameScene gameSite: ", gameSite, " sceneMode: ", sceneMode)
	s := NewLocalGameScene(creator, sceneId, gameId, gameSite, sceneMode, clycleTimes, params, gs, limitPlatform, playerNum, dbGameFree, baseScore, paramsEx...)
	if s == nil {
		return nil
	}
	this.scenes[sceneId] = s

	gs.AddScene(s)
	var platformName string
	if limitPlatform != nil {
		platformName = limitPlatform.IdStr
	}

	logger.Logger.Infof("(this *SceneMgr) CreateScene (gameId=%v), SceneId=%v platform=%v", gameId, sceneId, platformName)
	return s
}

func (this *SceneMgr) CreateClubScene(gameId int, gs *GameSession, limitPlatform *Platform, groupId int32, clubId int32,
	roomId string, roomPos int32, dbGameFree *server_proto.DB_GameFree, dbGameRule *server_proto.DB_GameRule) *Scene {
	gameMode := int(dbGameRule.GetGameMode())
	params := dbGameRule.GetParams()
	agentor := int32(0)
	creator := int32(0)
	sceneId := SceneMgrSington.RandGetSceneId()
	sceneMode := int(common.SceneMode_Club)
	clycleTimes := int32(1)
	numOfGames := int32(30) //model.ClubRoomGameTimes
	s := NewScene(agentor, creator, sceneId, gameId, gameMode, sceneMode, clycleTimes, numOfGames, params, gs, limitPlatform,
		groupId, dbGameFree, 0, dbGameFree.GetId())
	if s == nil {
		return nil
	}
	s.ClubId = clubId
	s.clubRoomID = roomId
	s.clubRoomPos = roomPos
	//club := clubManager.GetClub(clubId)
	//if club != nil {
	//	s.clubRoomTax = club.Setting.Taxes
	//}
	this.scenes[sceneId] = s
	gs.AddScene(s)
	var platformName string
	if limitPlatform != nil {
		platformName = limitPlatform.IdStr
	}
	logger.Logger.Infof("(this *SceneMgr) CreateClubScene (gameId=%v, mode=%v), SceneId=%v groupid=%v platform=%v",
		gameId, gameMode, sceneId, groupId, platformName)
	return s
}
func (this *SceneMgr) DestroyScene(sceneId int, isCompleted bool) {
	logger.Logger.Trace("(this *SceneMgr) DestroyScene ")
	if s, exist := this.scenes[sceneId]; exist {
		if s == nil {
			return
		}

		if s.IsCoinScene() {
			CoinSceneMgrSington.OnDestroyScene(s.sceneId)
		} else if s.IsHallScene() {
			PlatformMgrSington.OnDestroyScene(s)
		} else if s.IsHundredScene() {
			HundredSceneMgrSington.OnDestroyScene(s.sceneId)
		} else if s.IsPrivateScene() {
			//回收私有房间id
			this.psIdGen.Free(s.sceneId)
			if s.ClubId > 0 {
				//ClubSceneMgrSington.OnDestroyScene(s)
			} else {
				PrivateSceneMgrSington.OnDestroyScene(s)
			}
		} else if s.IsMatchScene() {
			//MatchMgrSington.DestroyScene(s.sceneId)
			CoinSceneMgrSington.OnDestroyScene(s.sceneId)
		}
		s.gameSess.DelScene(s)
		s.OnClose()
		delete(this.scenes, s.sceneId)

		logger.Logger.Infof("(this *SceneMgr) DestroyScene, SceneId=%v", sceneId)
	}
}

func (this *SceneMgr) DestroyMiniGameScene(sceneId int) {
	if s, exist := this.scenes[sceneId]; exist {
		if s == nil {
			return
		}

		if !s.IsMiniGameScene() {
			return
		}

		//MiniGameMgrSington.OnDestroyScene(s)

		s.gameSess.DelScene(s)
		delete(this.scenes, s.sceneId)
	}
}
func (this *SceneMgr) RandGetSceneId() int {
	return this.psIdGen.RandOne()
}

func (this *SceneMgr) GenOneMatchSceneId() int {
	this.autoId++
	if this.autoId > common.MatchSceneMaxId {
		this.autoId = common.MatchSceneStartId
	}
	return this.autoId
}

func (this *SceneMgr) GenOneCoinSceneId() int {
	this.coinSceneAutoId++
	if this.coinSceneAutoId > common.CoinSceneMaxId {
		this.coinSceneAutoId = common.CoinSceneStartId
	}
	return this.coinSceneAutoId
}

func (this *SceneMgr) GenOneHallSceneId() int {
	this.hallSceneAutoId++
	if this.hallSceneAutoId > common.HallSceneMaxId {
		this.hallSceneAutoId = common.HallSceneStartId
	}
	return this.hallSceneAutoId
}

func (this *SceneMgr) GenOneHundredSceneId() int {
	this.hundredSceneAutoId++
	if this.hundredSceneAutoId > common.HundredSceneMaxId {
		this.hundredSceneAutoId = common.HundredSceneStartId
	}
	return this.hundredSceneAutoId
}
func (this *SceneMgr) GetDgSceneId() int {
	return common.DgSceneId
}

func (this *SceneMgr) GetScene(sceneId int) *Scene {
	if s, exist := this.scenes[sceneId]; exist {
		return s
	}
	return nil
}

func (this *SceneMgr) GetSceneByPlayerId(snid int32) *Scene {
	if s, exist := this.playerInScene[snid]; exist {
		return s
	}
	return nil
}

func (this *SceneMgr) OnPlayerEnterScene(s *Scene, p *Player) {
	logger.Logger.Trace("(this *SceneMgr) OnPlayerEnterScene", p.SnId, s.sceneId)
	this.playerInScene[p.SnId] = s
}

func (this *SceneMgr) OnPlayerLeaveScene(s *Scene, p *Player) {
	logger.Logger.Trace("(this *SceneMgr) OnPlayerLeaveScene", p.SnId)
	delete(this.playerInScene, p.SnId)
	if !s.IsHundredScene() && !p.IsRob && !s.IsMatchScene() { //只记录对战场的
		const MINHOLD = 10
		const MAXHOLD = 20
		holdCnt := MINHOLD
		if csp, exist := CoinSceneMgrSington.sceneOfcsp[s.sceneId]; exist && csp != nil {
			holdCnt = csp.GetHasTruePlayerSceneCnt() + 2
			if holdCnt < MINHOLD {
				holdCnt = MINHOLD
			}
			if holdCnt > MAXHOLD {
				holdCnt = MAXHOLD
			}
		}
		if p.lastSceneId == nil {
			p.lastSceneId = make(map[int32][]int32)
		}
		id := s.dbGameFree.GetId()
		if sceneIds, exist := p.lastSceneId[id]; exist {
			if !common.InSliceInt32(sceneIds, int32(s.sceneId)) {
				sceneIds = append(sceneIds, int32(s.sceneId))
				cnt := len(sceneIds)
				if cnt > holdCnt {
					sceneIds = sceneIds[cnt-holdCnt:]
				}
				p.lastSceneId[id] = sceneIds
			}
		} else {
			p.lastSceneId[id] = []int32{int32(s.sceneId)}
		}
	}
}

func (this *SceneMgr) MarshalAllRoom(platform string, groupId, gameId int, gameMode, clubId, sceneMode, sceneId int,
	gameFreeId, snId int32, start, end, pageSize int32) ([]*webapi2.RoomInfo, int32, int32) {
	roomInfo := make([]*webapi2.RoomInfo, 0, len(this.scenes))
	var isNeedFindAll = false
	if model.GameParamData.IsFindRoomByGroup && platform != "" && snId != 0 && gameId == 0 &&
		gameMode == 0 && sceneId == -1 && groupId == 0 && clubId == 0 && sceneMode == 0 {
		p := PlayerMgrSington.GetPlayerBySnId(snId)
		if p != nil && p.Platform == platform {
			isNeedFindAll = true
		}
	}
	for _, s := range this.scenes {
		if (((s.limitPlatform != nil && s.limitPlatform.IdStr == platform) || platform == "") &&
			((s.gameId == gameId && s.gameMode == gameMode) || gameId == 0) &&
			(s.sceneId == sceneId || sceneId == 0) && (s.groupId == int32(groupId) || groupId == 0) &&
			(s.ClubId == int32(clubId) || clubId == 0) && (s.dbGameFree.GetId() == gameFreeId || gameFreeId == 0) &&
			(s.sceneMode == sceneMode || sceneMode == -1)) || isNeedFindAll {
			var platformName string
			if s.limitPlatform != nil {
				platformName = s.limitPlatform.IdStr
			}

			si := &webapi2.RoomInfo{
				Platform:   platformName,
				SceneId:    int32(s.sceneId),
				GameId:     int32(s.gameId),
				GameMode:   int32(s.gameMode),
				SceneMode:  int32(s.sceneMode),
				GroupId:    s.groupId,
				Creator:    s.creator,
				Agentor:    s.agentor,
				ReplayCode: s.replayCode,
				Params:     s.params,
				PlayerCnt:  int32(len(s.players) - s.robotNum),
				RobotCnt:   int32(s.robotNum),
				CreateTime: s.createTime.Unix(),
				ClubId:     s.ClubId,
			}
			if s.paramsEx != nil && len(s.paramsEx) > 0 {
				si.GameFreeId = s.paramsEx[0]
			}
			if s.starting {
				si.Start = 1
			} else {
				si.Start = 0
			}
			if s.IsHundredScene() {
				si.Start = 1
			}
			if s.gameSess != nil {
				si.SrvId = s.gameSess.GetSrvId()
			}
			cnt := 0
			total := len(s.players)
			robots := []int32{}

			isContinue := false
			if snId != 0 {
				for _, p := range s.players {
					if p.SnId == int32(snId) {
						isContinue = true
						break
					}
				}
			} else {
				isContinue = true
			}
			if !isContinue {
				continue
			}

			//优先显示玩家
			for id, p := range s.players {
				if !p.IsRob || total < 10 {
					si.PlayerIds = append(si.PlayerIds, id)
					cnt++
				} else {
					robots = append(robots, id)
				}
				if cnt > 10 {
					break
				}
			}
			//不够再显示机器人
			if total > cnt && cnt < 10 && len(robots) != 0 {
				for i := 0; cnt < 10 && i < len(robots); i++ {
					si.PlayerIds = append(si.PlayerIds, robots[i])
					cnt++
					if cnt > 10 {
						break
					}
				}
			}
			roomInfo = append(roomInfo, si)
		}
	}

	for i := 0; i < len(roomInfo); i++ {
		for k := 0; k < i; k++ {
			if roomInfo[i].CreateTime < roomInfo[k].CreateTime {
				roomInfo[i], roomInfo[k] = roomInfo[k], roomInfo[i]
			}
		}
	}

	//分页处理
	roomSum := float64(len(roomInfo))                          //房间总数
	pageCount := int32(math.Ceil(roomSum / float64(pageSize))) //总页数
	if roomSum <= float64(start) {
		start = 0
	}
	if roomSum < float64(end) {
		end = int32(roomSum)
	}
	needList := roomInfo[start:end] //需要的房间列表
	if len(needList) > 0 {
		return needList, pageCount, int32(roomSum)
	}
	return nil, 0, int32(roomSum)
}

func (this *SceneMgr) DeleteLongTimeInactive() {
	for _, s := range this.scenes {
		if webapi.ThridPlatformMgrSington.FindPlatformByPlatformBaseGameId(s.gameId) != nil {
			continue
		}
		if s.IsCoinScene() {
			if s.IsLongTimeInactive() && s.dbGameFree.GetCreateRoomNum() == 0 { //预创建的房间，暂不过期销毁，判定依据:CreateRoomNum>0
				logger.Logger.Warnf("SceneMgr.DeleteLongTimeInactive CoinScene ForceDelete scene:%v IsLongTimeInactive", s.sceneId)
				s.ForceDelete(false)
			}
		} else if s.IsPrivateScene() {
			if s.IsLongTimeInactive() {
				logger.Logger.Warnf("SceneMgr.DeleteLongTimeInactive PrivateScene ForceDelete scene:%v IsLongTimeInactive", s.sceneId)
				s.ForceDelete(false)
			}
		}
	}
}

func (this *SceneMgr) OnShutdown() {
	logger.Logger.Trace("(this *SceneMgr) Shutdown")
	for _, s := range this.scenes {
		s.Shutdown()
	}
}

func (this *SceneMgr) RebindPlayerSnId(oldSnId, newSnId int32) {
	if s, exist := this.playerInScene[oldSnId]; exist {
		delete(this.playerInScene, oldSnId)
		this.playerInScene[newSnId] = s
	}
	for _, s := range this.scenes {
		s.RebindPlayerSnId(oldSnId, newSnId)
	}
}
func (this *SceneMgr) GetDgScene() *Scene {
	sceneId := SceneMgrSington.GetDgSceneId()
	scene := this.GetScene(sceneId)
	if scene != nil {
		return scene
	}

	gs := GameSessMgrSington.GetMinLoadSess(common.GameId_Thr_Dg)
	if gs != nil {
		limitPlatform := PlatformMgrSington.GetPlatform(Default_Platform)
		var gameMode = 0
		dbGameFree := srvdata.PBDB_GameFreeMgr.GetData(280010001)
		scene := SceneMgrSington.CreateScene(0, 0, sceneId, common.GameId_Thr_Dg, gameMode, int(common.SceneMode_Thr), 1, -1,
			[]int32{}, gs, limitPlatform, 0, dbGameFree, 280010001)
		return scene
	} else {
		logger.Logger.Errorf("Get %v game min session failed.", common.GameId_Thr_Dg)
		return nil
	}
}
func (this *SceneMgr) GetThirdScene(i webapi.IThirdPlatform) *Scene {
	if i == nil {
		return nil
	}
	sceneId := i.GetPlatformBase().SceneId
	scene := this.GetScene(sceneId)
	if scene != nil {
		return scene
	}

	gs := GameSessMgrSington.GetMinLoadSess(i.GetPlatformBase().BaseGameID)
	if gs != nil {
		limitPlatform := PlatformMgrSington.GetPlatform(Default_Platform)
		var gameMode = common.SceneMode_Thr
		dbGameFree := srvdata.PBDB_GameFreeMgr.GetData(i.GetPlatformBase().VultGameID)
		scene := SceneMgrSington.CreateScene(0, 0, sceneId, i.GetPlatformBase().BaseGameID, gameMode, int(common.SceneMode_Thr), 1, -1,
			[]int32{}, gs, limitPlatform, 0, dbGameFree, i.GetPlatformBase().VultGameID)
		return scene
	} else {
		logger.Logger.Errorf("Get %v game min session failed.", i.GetPlatformBase().BaseGameID)
		return nil
	}
}

// 感兴趣所有clock event
func (this *SceneMgr) InterestClockEvent() int {
	return 1 << CLOCK_EVENT_MINUTE
}

func (this *SceneMgr) OnMiniTimer() {
	this.DeleteLongTimeInactive()
}

func (this *SceneMgr) SendToGame(gameid int, packetid int, pack interface{}) {
	gameServers := GameSessMgrSington.GetGameServerSess(gameid)
	if len(gameServers) == 0 {
		gameServers = GameSessMgrSington.GetGameServerSess(common.GameId_Unknow)
	}
	for _, value := range gameServers {
		value.Send(packetid, pack)
	}
}
func (this *SceneMgr) GetScenesByGame(gameid int) []*Scene {
	scenes := []*Scene{}
	for _, value := range this.scenes {
		if value.gameId == gameid {
			scenes = append(scenes, value)
		}
	}
	return scenes
}

func (this *SceneMgr) GetScenesByGameFreeId(gameFreeId int32) []*Scene {
	scenes := []*Scene{}
	for _, value := range this.scenes {
		if value.dbGameFree.GetId() == gameFreeId {
			scenes = append(scenes, value)
		}
	}
	return scenes
}

func init() {
	ClockMgrSington.RegisteSinker(SceneMgrSington)
}
