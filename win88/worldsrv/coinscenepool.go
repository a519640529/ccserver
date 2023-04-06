package main

import (
	"sort"
	"time"

	"games.yol.com/win88/common"
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	gamehall_proto "games.yol.com/win88/protocol/gamehall"
	server_proto "games.yol.com/win88/protocol/server"
	"games.yol.com/win88/srvdata"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/module"
)

const (
	MatchTrueMan_Forbid    int32 = -1 //禁止匹配真人
	MatchTrueMan_Unlimited       = 0  //不限制
	MatchTrueMan_Priority        = 1  //优先匹配真人
)
const (
	CoinSceneMatchTime = 3 //撮合间隔
	CoinSceneRoundTime = 9 //每轮超时时间
	CoinSceneOverTime  = 1 //清退的轮数
)

type CoinSceneContext struct {
	coin int32
}

type CoinScenePool struct {
	id              int32
	groupId         int32
	dbGameFree      *server_proto.DB_GameFree
	dbGameRule      *server_proto.DB_GameRule
	scenes          map[int]*Scene
	players         map[int32]int
	nextDeviationTs int64
	numDeviation    int32
	queue           map[int32]*Player
	robRequireNum   int
	robDismissNum   int
	inQueueRob      int
	inQueueTrueMan  int
	lastInviteTs    int64
	lastQueueTs     int64
	lastDismissTs   int64
}

func NewCoinScenePool(id, groupId int32, dbGameFree *server_proto.DB_GameFree) *CoinScenePool {
	csp := &CoinScenePool{
		id:            id,
		groupId:       groupId,
		dbGameFree:    dbGameFree,
		scenes:        make(map[int]*Scene),
		players:       make(map[int32]int),
		queue:         make(map[int32]*Player),
		robRequireNum: 1,
	}
	if !csp.init() {
		return nil
	}
	return csp
}

func (csp *CoinScenePool) RebindPlayerSnId(oldSnId, newSnId int32) {
	if id, exist := csp.players[oldSnId]; exist {
		csp.players[newSnId] = id
	}
}

func (csp *CoinScenePool) init() bool {
	if csp.dbGameFree == nil {
		csp.dbGameFree = srvdata.PBDB_GameFreeMgr.GetData(csp.id)
	}
	if csp.dbGameFree == nil {
		logger.Logger.Errorf("Coin scene pool init failed,%v game free data not find.", csp.id)
		return false
	}
	csp.dbGameRule = srvdata.PBDB_GameRuleMgr.GetData(csp.dbGameFree.GetGameRule())
	if csp.dbGameRule == nil {
		if csp.dbGameFree.GetGameRule() != 0 {
			logger.Logger.Errorf("Coin scene pool init failed,%v game rule data not find.", csp.dbGameFree.GetGameRule())
		}
		return false
	}

	csp.TryRefreshDeviation()

	return true
}

func (csp *CoinScenePool) TryRefreshDeviation() {
	if csp.nextDeviationTs < module.AppModule.GetCurrTimeSec() {
		//csp.nextDeviationTs = module.AppModule.GetCurrTimeSec() + 3600
		//deviation := csp.dbGameFree.GetDeviation()
		//if deviation > 0 {
		//	val := rand.Int31n(deviation * 2)
		//	val -= deviation
		//	csp.numDeviation = val
		//} else {
		csp.numDeviation = 0
		//}
	}
}

func (csp *CoinScenePool) IsGame(gameId, gameMode int32) bool {
	if csp.dbGameRule != nil && csp.dbGameRule.GetGameId() == gameId && csp.dbGameRule.GetGameMode() == gameMode {
		return true
	}
	return false
}

func (csp *CoinScenePool) GetPlayerNum() int32 {
	return int32(len(csp.players))
}

func (csp *CoinScenePool) GetFakePlayerNum() int32 {
	//if csp.dbGameFree != nil {
	//	correctNum := csp.dbGameFree.GetCorrectNum()
	//	correctRate := csp.dbGameFree.GetCorrectRate()
	//	count := csp.GetPlayerNum()
	//	return correctNum + count*correctRate/100 + csp.numDeviation
	//}
	return 0
}

func (csp *CoinScenePool) GetSceneType() int {
	if csp.dbGameFree != nil {
		return int(csp.dbGameFree.GetSceneType())
	}
	return 0
}

func (csp *CoinScenePool) CanInviteRob() bool {
	if csp.dbGameFree != nil {
		return csp.dbGameFree.GetBot() != 0
	}
	return false
}

func (csp *CoinScenePool) CanEnter(p *Player) gamehall_proto.OpResultCode {
	if csp.dbGameFree == nil || p == nil {
		return gamehall_proto.OpResultCode_OPRC_Error
	}

	//检测房间状态是否开启
	gps := PlatformMgrSington.GetGameConfig(p.Platform, csp.id)
	if gps == nil {
		return gamehall_proto.OpResultCode_OPRC_RoomHadClosed
	}

	dbGameFree := csp.dbGameFree
	if dbGameFree == nil {
		return gamehall_proto.OpResultCode_OPRC_RoomHadClosed
	}
	if dbGameFree.GetLimitCoin() != 0 && int64(dbGameFree.GetLimitCoin()) > p.Coin {
		return gamehall_proto.OpResultCode_OPRC_CoinNotEnough
	}

	if dbGameFree.GetMaxCoinLimit() != 0 && int64(dbGameFree.GetMaxCoinLimit()) < p.Coin && !p.IsRob {
		return gamehall_proto.OpResultCode_OPRC_CoinTooMore
	}

	return gamehall_proto.OpResultCode_OPRC_Sucess
}

func (csp *CoinScenePool) CanAudienceEnter(p *Player) gamehall_proto.OpResultCode {
	if csp.dbGameFree == nil || p == nil {
		return gamehall_proto.OpResultCode_OPRC_Error
	}

	//检测房间状态是否开启
	gps := PlatformMgrSington.GetGameConfig(p.Platform, csp.id)
	if gps == nil {
		return gamehall_proto.OpResultCode_OPRC_RoomHadClosed
	}

	dbGameFree := csp.dbGameFree
	if dbGameFree == nil {
		return gamehall_proto.OpResultCode_OPRC_RoomHadClosed
	}

	return gamehall_proto.OpResultCode_OPRC_Sucess
}

func (csp *CoinScenePool) PlayerEnter(p *Player, roomId int32, exclude []int32, ischangeroom bool) gamehall_proto.OpResultCode {
	if ret := csp.CanEnter(p); ret != gamehall_proto.OpResultCode_OPRC_Sucess {
		logger.Logger.Warnf("(csp *CoinScenePool) PlayerEnter find snid:%v csp.CanEnter coin:%v ret:%v id:%v", p.SnId,
			p.Coin, ret, csp.dbGameFree.GetId())
		return ret
	}
	if sceneId, ok := csp.players[p.SnId]; ok && sceneId != 0 {
		logger.Logger.Warnf("(csp *CoinScenePool) PlayerEnter find snid:%v ining sceneid:%v", p.SnId, sceneId)
		return gamehall_proto.OpResultCode_OPRC_Error
	}

	if p.scene != nil {
		logger.Logger.Warnf("(csp *CoinScenePool) PlayerEnter[p.scene != nil] find snid:%v in scene:%v gameId:%v", p.SnId, p.scene.sceneId, p.scene.gameId)
		return gamehall_proto.OpResultCode_OPRC_Error
	}

	var scene *Scene
	if roomId != 0 && (p.IsRob || p.GMLevel > 0 || csp.dbGameFree.GetCreateRoomNum() != 0) {
		if s, ok := csp.scenes[int(roomId)]; ok {
			if s != nil && !s.deleting { //指定房间id进入，那么忽略掉排除id
				if s.IsFull() {
					return gamehall_proto.OpResultCode_OPRC_RoomIsFull
				}
				if sp, ok := s.sp.(*ScenePolicyData); ok {
					if !s.starting || sp.EnterAfterStart {
						scene = s
					} else {
						logger.Logger.Warnf("(csp *CoinScenePool) PlayerEnter[!s.starting || sp.EnterAfterStart] snid:%v sceneid:%v starting:%v EnterAfterStart:%v", p.SnId, s.sceneId, s.starting, sp.EnterAfterStart)
						return gamehall_proto.OpResultCode_OPRC_Error
					}
				}
			}
		} else {
			logger.Logger.Warnf("(csp *CoinScenePool) PlayerEnter(robot:%v,roomid:%v, exclude:%v, ischangeroom:%v) no found scene", p.SnId, roomId, exclude, ischangeroom)
			return gamehall_proto.OpResultCode_OPRC_Error
		}
	}
	if csp.dbGameFree.GetMatchMode() == 1 {
		if csp.EnterQueue(p) {
			return gamehall_proto.OpResultCode_OPRC_CoinSceneEnterQueueSucc
		} else {
			return gamehall_proto.OpResultCode_OPRC_CoinSceneEnterQueueFail
		}
	}
	totlePlayer := len(csp.players)
	// 配房规则
	// 如果全局为限制则看具体的游戏是否限制，如果全局为不限制则不考虑具体的游戏。
	sameIpLimit := !model.GameParamData.SameIpNoLimit
	if sameIpLimit && csp.dbGameFree.GetSameIpLimit() == 0 {
		sameIpLimit = false
	}
	//真人匹配

	matchTrueManRule := csp.dbGameFree.GetMatchTrueMan()
	//gameId := int(csp.dbGameFree.GetGameId())
	//先做下通用的过滤,尽量减少重复的计算
	scenes := make(map[int]*Scene)
	for sceneId, s := range csp.scenes {
		if s != nil && !s.IsFull() && !s.deleting && !common.InSliceInt32(exclude, int32(s.sceneId)) {
			//规避同ip的用户在一个房间内(GM除外)
			if sameIpLimit && p.GMLevel == 0 && s.HasSameIp(p.Ip) {
				continue
			}
			//多少局只能禁止再配对
			if s.dbGameFree.GetSamePlaceLimit() > 0 && sceneLimitMgr.LimitSamePlace(p, s) {
				continue
			}
			//牌局开始后禁止进入
			if sp, ok := s.sp.(*ScenePolicyData); ok {
				if s.starting && !sp.EnterAfterStart {
					continue
				}
			}
			//禁止真人匹配
			if matchTrueManRule == MatchTrueMan_Forbid && !p.IsRob && s.GetTruePlayerCnt() != 0 {
				continue
			}
			scenes[sceneId] = s
		}
	}
	//优先黑白名单
	if scene == nil && len(scenes) != 0 {
		gamefreeid := csp.dbGameFree.GetId()
		gameid := csp.dbGameFree.GetGameId()
		if p.BlackLevel > 0 { //黑名单玩家
			var cntWhite int
			var cntLose int
			var whiteScene *Scene
			var loseScene *Scene
			for _, s := range scenes {
				if s != nil {
					if sceneLimitMgr.LimitAvgPlayer(s, totlePlayer) {
						continue
					}
					cnt := s.GetWhitePlayerCnt()
					if cnt > cntWhite {
						cntWhite = cnt
						whiteScene = s
					}
					cnt = s.GetLostPlayerCnt()
					if cnt > cntLose {
						cntLose = cnt
						loseScene = s
					}
				}
			}
			if whiteScene != nil {
				scene = whiteScene
			} else if loseScene != nil {
				scene = loseScene
			}
		} else if p.WhiteLevel > 0 { //白名单玩家
			var cntBlack int
			var cntWin int
			var blackScene *Scene
			var winScene *Scene
			for _, s := range scenes {
				if s != nil {
					if sceneLimitMgr.LimitAvgPlayer(s, totlePlayer) {
						continue
					}
					cnt := s.GetBlackPlayerCnt()
					if cnt > cntBlack {
						cntBlack = cnt
						blackScene = s
					}
					cnt = s.GetWinPlayerCnt()
					if cnt > cntWin {
						cntWin = cnt
						winScene = s
					}

				}
			}
			if blackScene != nil {
				scene = blackScene
			} else if winScene != nil {
				scene = winScene
			}
		} else { //按类型匹配
			//优先真人
			if scene == nil && len(scenes) != 0 && matchTrueManRule == MatchTrueMan_Priority {
				selScene := []*Scene{}
				for _, value := range scenes {
					if value != nil {
						if value.GetTruePlayerCnt() > 0 && !value.IsFull() {
							selScene = append(selScene, value)
						}
					}
				}
				if len(selScene) > 0 {
					sort.Slice(selScene, func(i, j int) bool {
						return selScene[i].GetTruePlayerCnt() > selScene[j].GetTruePlayerCnt()
					})
					scene = selScene[0]
				}
			}

			if scene == nil && len(scenes) != 0 {
				//1.优先具体场次的配桌规则
				matchFunc := GetCoinSceneMatchFunc(int(gamefreeid))
				if matchFunc != nil {
					scene = matchFunc(csp, p, scenes, sameIpLimit, exclude)
				}
				//2.其次游戏的配桌规则
				if scene == nil {
					matchFunc = GetCoinSceneMatchFunc(int(gameid))
					if matchFunc != nil {
						scene = matchFunc(csp, p, scenes, sameIpLimit, exclude)
					}
				}
			}

			//3.最后通用的数据表格驱动匹配规则
			if scene == nil && len(scenes) != 0 {
				t := p.CheckType(gameid, gamefreeid)
				if t != nil {
					typesMap := make(map[int][]int32)
					for _, s := range scenes {
						if s != nil {
							if sceneLimitMgr.LimitAvgPlayer(s, totlePlayer) {
								continue
							}

							if t != nil {
								types := s.GetPlayerType(gameid, gamefreeid)
								typesMap[s.sceneId] = types
							}
						}
					}

					//优先排除掉可匹配的房间
					ep := t.GetExcludeMatch()
					if len(typesMap) != 0 {
						for _, tid := range ep {
							for sceneid, types := range typesMap {
								if common.InSliceInt32(types, tid) {
									delete(typesMap, sceneid)
								}
							}
						}
					}

					//根据匹配优先级找房间
					mp := t.GetMatchPriority()
					if len(typesMap) != 0 {
						for _, tid := range mp {
							for sceneid, types := range typesMap {
								if common.InSliceInt32(types, tid) {
									scene = SceneMgrSington.GetScene(sceneid)
									break
								}
							}
						}
						//没有针对类型的玩家，也可匹配
						for sceneid, types := range typesMap {
							if types == nil {
								scene = SceneMgrSington.GetScene(sceneid)
								break
							}
						}
					}
				} else {
					for _, s := range scenes {
						if s != nil {
							if sceneLimitMgr.LimitAvgPlayer(s, totlePlayer) {
								continue
							}
							scene = s
							break
						}
					}
				}
			}
		}
	}

	if scene == nil {
		scene = csp.CreateNewScene(p)
		if scene != nil {
			csp.scenes[scene.sceneId] = scene
		} else {
			logger.Logger.Errorf("Create %v scene failed.", csp.id)
		}
	}
	if scene != nil {
		if p.EnterScene(scene, ischangeroom, -1) {
			csp.OnPlayerEnter(p, scene)
			return gamehall_proto.OpResultCode_OPRC_Sucess
		}
	}
	logger.Logger.Warnf("(csp *CoinScenePool) PlayerEnter snid:%v not found scene", p.SnId)
	return gamehall_proto.OpResultCode_OPRC_SceneServerMaintain
}

func (csp *CoinScenePool) PlayerEnterLocalGame(p *Player, roomId int32, exclude []int32, ischangeroom bool) gamehall_proto.OpResultCode {
	if ret := csp.CanEnter(p); ret != gamehall_proto.OpResultCode_OPRC_Sucess {
		logger.Logger.Warnf("(csp *CoinScenePool) PlayerEnter find snid:%v csp.CanEnter coin:%v ret:%v id:%v", p.SnId,
			p.Coin, ret, csp.dbGameFree.GetId())
		return ret
	}
	if sceneId, ok := csp.players[p.SnId]; ok && sceneId != 0 {
		logger.Logger.Warnf("(csp *CoinScenePool) PlayerEnter find snid:%v ining sceneid:%v", p.SnId, sceneId)
		return gamehall_proto.OpResultCode_OPRC_Error
	}

	if p.scene != nil {
		logger.Logger.Warnf("(csp *CoinScenePool) PlayerEnter[p.scene != nil] find snid:%v in scene:%v gameId:%v", p.SnId, p.scene.sceneId, p.scene.gameId)
		return gamehall_proto.OpResultCode_OPRC_Error
	}

	var scene *Scene
	if roomId != 0 && (p.IsRob || p.GMLevel > 0 || csp.dbGameFree.GetCreateRoomNum() != 0) {
		if s, ok := csp.scenes[int(roomId)]; ok {
			if s != nil && !s.deleting { //指定房间id进入，那么忽略掉排除id
				if s.IsFull() {
					return gamehall_proto.OpResultCode_OPRC_RoomIsFull
				}
				if sp, ok := s.sp.(*ScenePolicyData); ok {
					if !s.starting || sp.EnterAfterStart {
						scene = s
					} else {
						logger.Logger.Warnf("(csp *CoinScenePool) PlayerEnter[!s.starting || sp.EnterAfterStart] snid:%v sceneid:%v starting:%v EnterAfterStart:%v", p.SnId, s.sceneId, s.starting, sp.EnterAfterStart)
						return gamehall_proto.OpResultCode_OPRC_Error
					}
				}
			}
		} else {
			logger.Logger.Warnf("(csp *CoinScenePool) PlayerEnter(robot:%v,roomid:%v, exclude:%v, ischangeroom:%v) no found scene", p.SnId, roomId, exclude, ischangeroom)
			return gamehall_proto.OpResultCode_OPRC_Error
		}
	}
	if csp.dbGameFree.GetMatchMode() == 1 {
		if csp.EnterQueue(p) {
			return gamehall_proto.OpResultCode_OPRC_CoinSceneEnterQueueSucc
		} else {
			return gamehall_proto.OpResultCode_OPRC_CoinSceneEnterQueueFail
		}
	}
	totlePlayer := len(csp.players)
	// 配房规则
	// 如果全局为限制则看具体的游戏是否限制，如果全局为不限制则不考虑具体的游戏。
	sameIpLimit := !model.GameParamData.SameIpNoLimit
	if sameIpLimit && csp.dbGameFree.GetSameIpLimit() == 0 {
		sameIpLimit = false
	}
	//真人匹配

	matchTrueManRule := csp.dbGameFree.GetMatchTrueMan()
	gameId := int(csp.dbGameFree.GetGameId())
	//根据携带金额取进房间底注 DB_Createroom
	playerTakeCoin := p.Coin
	var dbCreateRoom *server_proto.DB_Createroom
	arrs := srvdata.PBDB_CreateroomMgr.Datas.Arr
	for i := len(arrs) - 1; i >= 0; i-- {
		if arrs[i].GetGameId() == int32(gameId) {
			goldRange := arrs[i].GoldRange
			if len(goldRange) == 0 {
				continue
			}
			if playerTakeCoin >= int64(goldRange[0]) {
				dbCreateRoom = arrs[i]
				break
			}
		}
	}

	//先做下通用的过滤,尽量减少重复的计算
	scenes := make(map[int]*Scene)
	for sceneId, s := range csp.scenes {
		if s != nil && !s.IsFull() && !s.deleting && !common.InSliceInt32(exclude, int32(s.sceneId)) {
			//规避同ip的用户在一个房间内(GM除外)
			if sameIpLimit && p.GMLevel == 0 && s.HasSameIp(p.Ip) {
				continue
			}
			//多少局只能禁止再配对
			if s.dbGameFree.GetSamePlaceLimit() > 0 && sceneLimitMgr.LimitSamePlace(p, s) {
				continue
			}
			//牌局开始后禁止进入
			if sp, ok := s.sp.(*ScenePolicyData); ok {
				if s.starting && !sp.EnterAfterStart {
					continue
				}
			}
			//禁止真人匹配
			if matchTrueManRule == MatchTrueMan_Forbid && !p.IsRob && s.GetTruePlayerCnt() != 0 {
				continue
			}
			//根据携带金额取可进房间
			if dbCreateRoom != nil && len(dbCreateRoom.GetBetRange()) != 0 {
				betRange := dbCreateRoom.GetBetRange()
				for _, bet := range betRange {
					if s.BaseScore == bet {
						scenes[sceneId] = s
					}
				}
			}
		}
	}
	//优先黑白名单
	if scene == nil && len(scenes) != 0 {
		gamefreeid := csp.dbGameFree.GetId()
		gameid := csp.dbGameFree.GetGameId()
		if p.BlackLevel > 0 { //黑名单玩家
			var cntWhite int
			var cntLose int
			var whiteScene *Scene
			var loseScene *Scene
			for _, s := range scenes {
				if s != nil {
					if sceneLimitMgr.LimitAvgPlayer(s, totlePlayer) {
						continue
					}
					cnt := s.GetWhitePlayerCnt()
					if cnt > cntWhite {
						cntWhite = cnt
						whiteScene = s
					}
					cnt = s.GetLostPlayerCnt()
					if cnt > cntLose {
						cntLose = cnt
						loseScene = s
					}
				}
			}
			if whiteScene != nil {
				scene = whiteScene
			} else if loseScene != nil {
				scene = loseScene
			}
		} else if p.WhiteLevel > 0 { //白名单玩家
			var cntBlack int
			var cntWin int
			var blackScene *Scene
			var winScene *Scene
			for _, s := range scenes {
				if s != nil {
					if sceneLimitMgr.LimitAvgPlayer(s, totlePlayer) {
						continue
					}
					cnt := s.GetBlackPlayerCnt()
					if cnt > cntBlack {
						cntBlack = cnt
						blackScene = s
					}
					cnt = s.GetWinPlayerCnt()
					if cnt > cntWin {
						cntWin = cnt
						winScene = s
					}

				}
			}
			if blackScene != nil {
				scene = blackScene
			} else if winScene != nil {
				scene = winScene
			}
		} else { //按类型匹配
			//优先真人
			if scene == nil && len(scenes) != 0 && matchTrueManRule == MatchTrueMan_Priority {
				selScene := []*Scene{}
				for _, value := range scenes {
					if value != nil {
						if value.GetTruePlayerCnt() > 0 && !value.IsFull() {
							selScene = append(selScene, value)
						}
					}
				}
				if len(selScene) > 0 {
					sort.Slice(selScene, func(i, j int) bool {
						return selScene[i].GetTruePlayerCnt() > selScene[j].GetTruePlayerCnt()
					})
					scene = selScene[0]
				}
			}

			if scene == nil && len(scenes) != 0 {
				//1.优先具体场次的配桌规则
				matchFunc := GetCoinSceneMatchFunc(int(gamefreeid))
				if matchFunc != nil {
					scene = matchFunc(csp, p, scenes, sameIpLimit, exclude)
				}
				//2.其次游戏的配桌规则
				if scene == nil {
					matchFunc = GetCoinSceneMatchFunc(int(gameid))
					if matchFunc != nil {
						scene = matchFunc(csp, p, scenes, sameIpLimit, exclude)
					}
				}
			}

			//3.最后通用的数据表格驱动匹配规则
			if scene == nil && len(scenes) != 0 {
				t := p.CheckType(gameid, gamefreeid)
				if t != nil {
					typesMap := make(map[int][]int32)
					for _, s := range scenes {
						if s != nil {
							if sceneLimitMgr.LimitAvgPlayer(s, totlePlayer) {
								continue
							}

							if t != nil {
								types := s.GetPlayerType(gameid, gamefreeid)
								typesMap[s.sceneId] = types
							}
						}
					}

					//优先排除掉可匹配的房间
					ep := t.GetExcludeMatch()
					if len(typesMap) != 0 {
						for _, tid := range ep {
							for sceneid, types := range typesMap {
								if common.InSliceInt32(types, tid) {
									delete(typesMap, sceneid)
								}
							}
						}
					}

					//根据匹配优先级找房间
					mp := t.GetMatchPriority()
					if len(typesMap) != 0 {
						for _, tid := range mp {
							for sceneid, types := range typesMap {
								if common.InSliceInt32(types, tid) {
									scene = SceneMgrSington.GetScene(sceneid)
									break
								}
							}
						}
						//没有针对类型的玩家，也可匹配
						for sceneid, types := range typesMap {
							if types == nil {
								scene = SceneMgrSington.GetScene(sceneid)
								break
							}
						}
					}
				} else {
					for _, s := range scenes {
						if s != nil {
							if sceneLimitMgr.LimitAvgPlayer(s, totlePlayer) {
								continue
							}
							scene = s
							break
						}
					}
				}
			}
		}
	}

	if scene == nil {
		scene = csp.CreateLocalGameNewScene(p)
		if scene != nil {
			csp.scenes[scene.sceneId] = scene
		} else {
			logger.Logger.Errorf("Create %v scene failed.", csp.id)
		}
	}
	if scene != nil {
		if p.EnterScene(scene, ischangeroom, -1) {
			csp.OnPlayerEnter(p, scene)
			return gamehall_proto.OpResultCode_OPRC_Sucess
		}
	}
	logger.Logger.Warnf("(csp *CoinScenePool) PlayerEnter snid:%v not found scene", p.SnId)
	return gamehall_proto.OpResultCode_OPRC_SceneServerMaintain
}

func (csp *CoinScenePool) OnPlayerEnter(p *Player, scene *Scene) {
	csp.players[p.SnId] = scene.sceneId
}

func (csp *CoinScenePool) AudienceEnter(p *Player, roomId int32, exclude []int32, ischangeroom bool) gamehall_proto.OpResultCode {
	if ret := csp.CanAudienceEnter(p); ret != gamehall_proto.OpResultCode_OPRC_Sucess {
		logger.Logger.Warnf("(csp *CoinScenePool) AudienceEnter find snid:%v csp.CanEnter coin:%v ret:%v id:%v", p.SnId, p.Coin, ret, csp.dbGameFree.GetId())
		return ret
	}
	if sceneId, ok := csp.players[p.SnId]; ok && sceneId != 0 {
		logger.Logger.Warnf("(csp *CoinScenePool) AudienceEnter find snid:%v ining sceneid:%v", p.SnId, sceneId)
		return gamehall_proto.OpResultCode_OPRC_NoFindDownTiceRoom
	}

	if p.scene != nil {
		logger.Logger.Warnf("(csp *CoinScenePool) AudienceEnter[p.scene != nil] find snid:%v in scene:%v gameId:%v", p.SnId, p.scene.sceneId, p.scene.gameId)
		return gamehall_proto.OpResultCode_OPRC_Error
	}

	var scene *Scene
	if roomId != 0 {
		if s, ok := csp.scenes[int(roomId)]; ok {
			if s != nil && !s.deleting /*&& s.sceneId != int(exclude)*/ {
				scene = s
			} else {
				logger.Logger.Warnf("(csp *CoinScenePool) AudienceEnter[!s.starting || sp.EnterAfterStart] snid:%v sceneid:%v starting:%v EnterAfterStart:%v", p.SnId, s.sceneId, s.starting)
			}
		}
	}
	if scene == nil {
		for _, s := range csp.scenes {
			if s != nil && len(s.players) != 0 && !s.deleting && !common.InSliceInt32(exclude, int32(s.sceneId)) {
				scene = s
				break
			}
		}
	}

	if scene == nil {
		return gamehall_proto.OpResultCode_OPRC_NoFindDownTiceRoom
	}
	if scene != nil {
		// 预创建房间检查观众数量
		if scene.IsPreCreateScene() && scene.GetAudienceCnt() >= model.GameParamData.MaxAudienceNum {
			return gamehall_proto.OpResultCode_OPRC_RoomIsFull
		}
		if scene.AudienceEnter(p, ischangeroom) {
			csp.OnPlayerEnter(p, scene)
			return gamehall_proto.OpResultCode_OPRC_Sucess
		}
	}
	logger.Logger.Warnf("(csp *CoinScenePool) PlayerEnter snid:%v not found scene", p.SnId)
	return gamehall_proto.OpResultCode_OPRC_NoFindDownTiceRoom
}

func (csp *CoinScenePool) AudienceLeave(p *Player, reason int) bool {
	var s *Scene
	logger.Logger.Tracef("(csp *CoinScenePool) AudienceLeave")
	if sceneId, ok := csp.players[p.SnId]; ok {
		if sceneId != 0 {
			if scene, ok := csp.scenes[sceneId]; ok && scene != nil {
				if scene.HasAudience(p) {
					s = scene
					logger.Logger.Warnf("(csp *CoinScenePool) AudienceLeave snid:%v found had in scene:%v", p.SnId, sceneId)
					scene.AudienceLeave(p, reason)
				}
			}
		}
	}
	csp.OnPlayerLeave(s, p)
	return true
}

func (csp *CoinScenePool) PlayerLeave(p *Player, reason int) bool {
	var s *Scene
	if sceneId, ok := csp.players[p.SnId]; ok {
		if sceneId != 0 {
			if scene, ok := csp.scenes[sceneId]; ok && scene != nil {
				if scene.HasPlayer(p) {
					s = scene
					logger.Logger.Warnf("(csp *CoinScenePool) PlayerLeave snid:%v in scene:%v", p.SnId, sceneId)
					scene.PlayerLeave(p, reason)
				}
			}
		}
	} else {
		if p.scene != nil && p.scene.paramsEx[0] == csp.id {
			sceneId = p.scene.sceneId
			if scene, ok := csp.scenes[sceneId]; ok && scene != nil {
				if scene.HasPlayer(p) {
					s = scene
					logger.Logger.Warnf("(csp *CoinScenePool) PlayerLeave exception snid:%v in scene:%v", p.SnId, sceneId)
					scene.PlayerLeave(p, reason)
				}
			}
		}
	}
	csp.OnPlayerLeave(s, p)
	return true
}

func (csp *CoinScenePool) OnPlayerLeave(s *Scene, p *Player) {
	delete(csp.players, p.SnId)

	if s != nil {
		if s.IsPrivateScene() { //私有空房间直接删除
			if s.IsEmpty() {
				s.ForceDelete(false)
			}
		} else {
			if s.IsEmpty() && s.IsPreCreateScene() { //避免房间数量过度泛滥
				hasCnt := len(csp.scenes)
				if hasCnt > int(csp.dbGameFree.GetCreateRoomNum()) {
					s.ForceDelete(false)
				}
			}
		}
	}
}

func (csp *CoinScenePool) CreateNewScene(p *Player) *Scene {
	sceneId := SceneMgrSington.GenOneCoinSceneId()
	gameId := int(csp.dbGameRule.GetGameId())
	gs := GameSessMgrSington.GetMinLoadSess(gameId)
	if gs != nil {
		gameMode := csp.dbGameRule.GetGameMode()
		params := csp.dbGameRule.GetParams()
		var platformName string
		limitPlatform := PlatformMgrSington.GetPlatform(p.Platform)
		if limitPlatform == nil || !limitPlatform.Isolated {
			limitPlatform = PlatformMgrSington.GetPlatform(Default_Platform)
			platformName = Default_Platform
		} else {
			platformName = limitPlatform.IdStr
		}

		scene := SceneMgrSington.CreateScene(0, 0, sceneId, gameId, int(gameMode), int(common.SceneMode_Public), 1, -1, params,
			gs, limitPlatform, csp.groupId, csp.dbGameFree, int32(csp.id))
		if scene != nil {
			scene.hallId = csp.id
			if csp.groupId != 0 {
				CoinSceneMgrSington.groupOfScene[int32(sceneId)] = csp.groupId
			} else {
				CoinSceneMgrSington.platformOfScene[int32(sceneId)] = platformName
			}
			CoinSceneMgrSington.sceneOfcsp[sceneId] = csp
			//移动到SceneMgr中集中处理
			////比赛场没水池
			//if !scene.IsMatchScene() {
			//	//平台水池设置
			//	gs.DetectCoinPoolSetting(limitPlatform.Name, scene.hallId, scene.groupId)
			//}
			return scene
		}
	} else {
		logger.Logger.Errorf("Get %v game min session failed.", gameId)
	}
	return nil
}

func (csp *CoinScenePool) CreateLocalGameNewScene(p *Player) *Scene {
	sceneId := SceneMgrSington.GenOneCoinSceneId()
	gameId := int(csp.dbGameRule.GetGameId())
	gs := GameSessMgrSington.GetMinLoadSess(gameId)
	if gs != nil {
		params := csp.dbGameRule.GetParams()
		var platformName string
		limitPlatform := PlatformMgrSington.GetPlatform(p.Platform)
		if limitPlatform == nil || !limitPlatform.Isolated {
			limitPlatform = PlatformMgrSington.GetPlatform(Default_Platform)
			platformName = Default_Platform
		} else {
			platformName = limitPlatform.IdStr
		}

		//根据携带金额取可创房间 DB_Createroom
		baseScore := int32(0)
		gameSite := 0
		playerTakeCoin := p.Coin
		var dbCreateRoom *server_proto.DB_Createroom
		arrs := srvdata.PBDB_CreateroomMgr.Datas.Arr
		for i := len(arrs) - 1; i >= 0; i-- {
			if arrs[i].GetGameId() == int32(gameId) {
				goldRange := arrs[i].GoldRange
				if len(goldRange) == 0 {
					continue
				}
				if playerTakeCoin >= int64(goldRange[0]) {
					dbCreateRoom = arrs[i]
					break
				}
			}
		}
		if dbCreateRoom == nil {
			logger.Logger.Tracef("CoinScenePool CreateLocalGameNewScene failed! playerTakeCoin:%v ", playerTakeCoin)
			return nil
		}
		if len(dbCreateRoom.GetBetRange()) != 0 && dbCreateRoom.GetBetRange()[0] != 0 {
			baseScore = common.RandInt32Slice(dbCreateRoom.GetBetRange())
			gameSite = int(dbCreateRoom.GetGameSite())
		}
		if baseScore == 0 {
			logger.Logger.Tracef("CoinScenePool CreateLocalGameNewScene failed! baseScore==0")
			return nil
		}

		scene := SceneMgrSington.CreateLocalGameScene(p.SnId, sceneId, gameId, gameSite, common.SceneMode_Public, 1, params, gs, limitPlatform, 4, csp.dbGameFree, baseScore, int32(csp.id))
		if scene != nil {
			scene.hallId = csp.id
			CoinSceneMgrSington.platformOfScene[int32(sceneId)] = platformName
			CoinSceneMgrSington.sceneOfcsp[sceneId] = csp
			return scene
		}
	} else {
		logger.Logger.Errorf("Get %v game min session failed.", gameId)
	}
	return nil
}

func (csp *CoinScenePool) CreateNewSceneByPlatform(platform string) *Scene {
	sceneId := SceneMgrSington.GenOneCoinSceneId()
	gameId := int(csp.dbGameRule.GetGameId())
	gs := GameSessMgrSington.GetMinLoadSess(gameId)
	if gs != nil {
		gameMode := csp.dbGameRule.GetGameMode()
		params := csp.dbGameRule.GetParams()
		var platformName string
		limitPlatform := PlatformMgrSington.GetPlatform(platform)
		if limitPlatform == nil || !limitPlatform.Isolated {
			limitPlatform = PlatformMgrSington.GetPlatform(Default_Platform)
			platformName = Default_Platform
		} else {
			platformName = limitPlatform.IdStr
		}
		var scene *Scene
		if common.IsLocalGame(gameId) {
			tmpIdx := common.RandInt(100)
			playerNum := 4
			if tmpIdx < 20 { //20%创建两人房间
				playerNum = 2
			}
			//根据SceneType随机可创房间 DB_Createroom
			baseScore := int32(0)
			gameSite := 0
			var dbCreateRooms []*server_proto.DB_Createroom
			arrs := srvdata.PBDB_CreateroomMgr.Datas.Arr
			for i := len(arrs) - 1; i >= 0; i-- {
				if arrs[i].GetGameId() == int32(gameId) && arrs[i].GetGameSite() == csp.dbGameFree.GetSceneType() {
					goldRange := arrs[i].GoldRange
					if len(goldRange) == 0 {
						continue
					}
					if goldRange[0] == 0 {
						continue
					}
					dbCreateRooms = append(dbCreateRooms, arrs[i])
				}
			}
			if len(dbCreateRooms) != 0 {
				randIdx := common.RandInt(len(dbCreateRooms))
				dbCreateRoom := dbCreateRooms[randIdx]
				if len(dbCreateRoom.GetBetRange()) != 0 && dbCreateRoom.GetBetRange()[0] != 0 {
					baseScore = common.RandInt32Slice(dbCreateRoom.GetBetRange())
					gameSite = int(dbCreateRoom.GetGameSite())
				}
				if baseScore != 0 {
					scene = SceneMgrSington.CreateLocalGameScene(0, sceneId, gameId, gameSite, int(common.SceneMode_Public), 1, params,
						gs, limitPlatform, playerNum, csp.dbGameFree, baseScore, int32(csp.id))
					if scene != nil {
						logger.Logger.Tracef("CreateLocalGameScene success.gameId:%v gameSite:%v baseScore:%v randIdx:%v", scene.gameId, scene.gameSite, baseScore, randIdx)
						//if gameId == common.GameId_TienLen {
						//	scenes := SceneMgrSington.GetScenesByGame(int(gameId))
						//	for _, sss := range scenes {
						//		if sss != nil && sss.sceneMode == common.SceneMode_Public {
						//			logger.Logger.Tracef("CreateLocalGameScene")
						//		}
						//	}
						//}
					}
				}
			}
		} else {
			scene = SceneMgrSington.CreateScene(0, 0, sceneId, gameId, int(gameMode), int(common.SceneMode_Public), 1, -1, params,
				gs, limitPlatform, csp.groupId, csp.dbGameFree, int32(csp.id))
		}
		if scene != nil {
			scene.hallId = csp.id
			if csp.groupId != 0 {
				CoinSceneMgrSington.groupOfScene[int32(sceneId)] = csp.groupId
			} else {
				CoinSceneMgrSington.platformOfScene[int32(sceneId)] = platformName
			}
			CoinSceneMgrSington.sceneOfcsp[sceneId] = csp
			//移动到SceneMgr中集中处理
			//if !scene.IsMatchScene() {
			//	//平台水池设置
			//	gs.DetectCoinPoolSetting(limitPlatform.Name, scene.hallId, scene.groupId)
			//}
			return scene
		}
	} else {
		logger.Logger.Errorf("Get %v game min session failed.", gameId)
	}
	return nil
}

func (csp *CoinScenePool) OnDestroyScene(sceneid int) {
	if scene, exist := csp.scenes[sceneid]; exist {
		delete(csp.scenes, sceneid)
		var needRefreshCoin bool
		if !scene.IsMiniGameScene() {
			needRefreshCoin = true
		}
		for snid, id := range csp.players {
			if sceneid == id {
				delete(csp.players, snid)
				if needRefreshCoin {
					player := PlayerMgrSington.GetPlayerBySnId(snid)
					if player != nil {
						if !player.IsRob {
							ctx := scene.GetPlayerGameCtx(player.SnId)
							if ctx != nil {
								//发送一个探针,等待ack后同步金币
								player.TryRetrieveLostGameCoin(sceneid)
							}
						}
					}
				}
			}
		}
	}

	if len(csp.scenes) <= 0 {
		if csp.groupId != 0 {
			if groupId, ok := CoinSceneMgrSington.groupOfScene[int32(sceneid)]; ok {
				delete(CoinSceneMgrSington.scenesOfGroup[groupId], int32(sceneid))
			}
		} else {
			if platformName, ok := CoinSceneMgrSington.platformOfScene[int32(sceneid)]; ok {
				delete(CoinSceneMgrSington.scenesOfPlatform[platformName], int32(sceneid))
			}
		}
	}
}

func evaluateCoinSceneIncCount(curcount, minCnt, playernum, perscenemax int) int {
	if perscenemax == 0 {
		perscenemax = 1
	}
	expectcnt := (playernum/perscenemax + 1) * 5 / 4
	if expectcnt < minCnt {
		expectcnt = minCnt
	}
	return expectcnt - curcount
}

func (csp *CoinScenePool) EnsurePreCreateRoom(platform string) {
	if platform == "0" {
		return
	}
	preCreateNum := int(csp.dbGameFree.GetCreateRoomNum())
	if preCreateNum == 0 || model.GameParamData.ClosePreCreateRoom {
		return
	}
	if p := PlatformMgrSington.GetPlatform(platform); p == nil || p.Disable {
		return
	}

	if len(csp.scenes) < preCreateNum {
		inc := preCreateNum - len(csp.scenes)
		for i := 0; i < inc; i++ {
			scene := csp.CreateNewSceneByPlatform(platform)
			if scene != nil {
				csp.scenes[scene.sceneId] = scene
			}
		}
	}

	//maxPlayerNum := 0
	//maxLimit := preCreateNum * model.GameParamData.PreCreateRoomAllowMaxMultiple
	////评估下动态扩展创建
	//if len(csp.scenes) < maxLimit {
	//	playerTotalNum := 0
	//	for _, s := range csp.scenes {
	//		playerTotalNum += len(s.players)
	//		maxPlayerNum = s.playerNum
	//	}
	//	inc := evaluateCoinSceneIncCount(len(csp.scenes), preCreateNum, playerTotalNum, maxPlayerNum)
	//	if inc > 0 {
	//		if inc+len(csp.scenes) > maxLimit {
	//			inc = maxLimit - len(csp.scenes)
	//		}
	//	}
	//	if inc > 0 {
	//		for i := 0; i < inc; i++ {
	//			scene := csp.CreateNewSceneByPlatform(platform)
	//			if scene != nil {
	//				csp.scenes[scene.sceneId] = scene
	//			}
	//		}
	//	}
	//}
}
func (csp *CoinScenePool) PlayerListRoom(p *Player) bool {
	if p.scene != nil {
		logger.Logger.Warnf("(csp *CoinScenePool) PlayerListRoom[p.scene != nil] find snid:%v in scene:%v gameId:%v", p.SnId, p.scene.sceneId, p.scene.gameId)
		return false
	}

	csp.EnsurePreCreateRoom(p.Platform)

	if len(csp.scenes) == 0 {
		return false
	}

	pack := &gamehall_proto.SCCoinSceneListRoom{
		Id:             csp.dbGameFree.Id,
		LimitCoin:      csp.dbGameFree.LimitCoin,
		MaxCoinLimit:   csp.dbGameFree.MaxCoinLimit,
		BaseScore:      csp.dbGameFree.BaseScore,
		MaxScore:       csp.dbGameFree.MaxChip,
		OtherIntParams: csp.dbGameFree.OtherIntParams,
	}

	maxPlayerNum := 0
	for sceneId, s := range csp.scenes {
		data := &gamehall_proto.CoinSceneInfo{
			SceneId:   proto.Int(sceneId),
			PlayerNum: proto.Int(len(s.players)),
		}
		pack.Datas = append(pack.Datas, data)
		maxPlayerNum = s.playerNum
	}
	pack.MaxPlayerNum = proto.Int(maxPlayerNum)
	proto.SetDefaults(pack)
	p.SendToClient(int(gamehall_proto.CoinSceneGamePacketID_PACKET_SC_COINSCENE_LISTROOM), pack)
	return true
}

func (csp *CoinScenePool) GetHasTruePlayerSceneCnt() int {
	cnt := 0
	for _, s := range csp.scenes {
		if s.GetTruePlayerCnt() != 0 {
			cnt++
		}
	}
	return cnt
}
func (csp *CoinScenePool) InQueue(p *Player) bool {
	if _, ok := csp.queue[p.SnId]; ok {
		return true
	}
	return false
}
func (csp *CoinScenePool) InviteRob(platform string) bool {
	if time.Now().Unix() < csp.lastInviteTs {
		return false
	}
	csp.lastInviteTs = time.Now().Add(time.Second * 3).Unix()
	if csp.dbGameFree.GetBot() == 0 { //机器人不进的场
		csp.robRequireNum = 0
		return false
	}
	if csp.dbGameFree.GetMatchMode() == 0 {
		return false
	}
	if csp.robRequireNum <= 0 {
		return false
	}
	pack := &server_proto.WGInviteRobEnterCoinSceneQueue{
		Platform:   proto.String(platform),
		GameFreeId: proto.Int32(csp.dbGameFree.GetId()),
		RobNum:     proto.Int(csp.robRequireNum),
	}
	SceneMgrSington.SendToGame(int(csp.dbGameFree.GetGameId()),
		int(server_proto.SSPacketID_PACKET_WG_INVITEROBENTERCOINSCENEQUEUE), pack)
	csp.robRequireNum = 0
	return true
}
func (csp *CoinScenePool) DismissRob(platform string) bool {
	if time.Now().Unix() < csp.lastDismissTs {
		return false
	}
	csp.lastDismissTs = time.Now().Add(time.Second * 10).Unix()
	if csp.dbGameFree.GetMatchMode() == 0 {
		return false
	}
	if csp.dbGameFree.GetMatchTrueMan() == MatchTrueMan_Forbid {
		requiredRob := csp.inQueueTrueMan * 5
		if csp.inQueueRob > requiredRob {
			csp.robDismissNum += (csp.inQueueRob - requiredRob)
		}
	} else {
		requiredRob := csp.inQueueTrueMan * 4
		if csp.inQueueRob > requiredRob {
			csp.robDismissNum += (csp.inQueueRob - requiredRob)
		}
	}
	if csp.robDismissNum <= 0 {
		return false
	}
	for _, value := range csp.queue {
		if value.IsRob {
			csp.QuitQueue(value.SnId)
			pack := &gamehall_proto.SCCoinSceneOp{
				Id:       proto.Int32(csp.dbGameFree.GetId()),
				OpType:   proto.Int32(common.CoinSceneOp_Leave),
				OpParams: []int32{},
			}
			value.SendToClient(int(gamehall_proto.CoinSceneGamePacketID_PACKET_SC_COINSCENE_OP), pack)
			csp.robDismissNum--
			if csp.robDismissNum <= 0 {
				break
			}
		}
	}
	csp.robDismissNum = 0
	return true
}
func (csp *CoinScenePool) QuitQueue(snid int32) {
	if player, ok := csp.queue[snid]; ok {
		player.CoinSceneQueue = nil
		if player.IsRob {
			csp.inQueueRob--
		} else {
			csp.inQueueTrueMan--
		}
		delete(csp.queue, snid)
	}
}
func (csp *CoinScenePool) EnterQueue(player *Player) bool {
	if csp.InQueue(player) {
		logger.Logger.Info("Repeate ennter coinscene queue.")
		return true
	}
	csp.queue[player.SnId] = player
	if player.IsRob {
		csp.inQueueRob++
	} else {
		player.EnterQueueTime = time.Now()
		csp.inQueueTrueMan++
		num := GetGameSuiableNum(int(csp.dbGameFree.GetGameId()), csp.dbGameFree.GetMatchTrueMan())
		requiredRob := csp.inQueueTrueMan * num
		if csp.inQueueRob < requiredRob {
			csp.robRequireNum += (requiredRob - csp.inQueueRob)
		}
	}
	player.EnterCoinSceneQueueTs = time.Now().Add(time.Second * 15).Unix()
	player.CoinSceneQueueRound = 1
	player.CoinSceneQueue = csp
	pack := &gamehall_proto.SCCoinSceneQueueState{
		GameFreeId: proto.Int32(csp.dbGameFree.GetId()),
		Count:      proto.Int32(player.CoinSceneQueueRound),
		Ts:         proto.Int64(CoinSceneRoundTime),
	}
	player.SendToClient(int(gamehall_proto.CoinSceneGamePacketID_PACKET_SC_COINSCENE_QUEUESTATE), pack)
	return true
}
func (csp *CoinScenePool) ProcessQueue(index int) {
	if time.Now().Unix() < csp.lastQueueTs {
		return
	}
	csp.lastQueueTs = time.Now().Add(time.Second * CoinSceneMatchTime).Unix()
	if csp.dbGameFree.GetMatchMode() == 0 {
		return
	}
	if len(csp.queue) == 0 {
		return
	}
	for _, value := range csp.queue {
		if value.scene != nil {
			csp.QuitQueue(value.SnId)
			pack := &gamehall_proto.SCCoinSceneOp{
				OpCode: gamehall_proto.OpResultCode_OPRC_CoinSceneEnterQueueOverTime,
				Id:     proto.Int32(csp.dbGameFree.GetId()),
				OpType: proto.Int32(common.CoinSceneOp_Leave),
			}
			value.SendToClient(int(gamehall_proto.CoinSceneGamePacketID_PACKET_SC_COINSCENE_OP), pack)
		}
	}
	if csp.dbGameFree.GetMatchTrueMan() == MatchTrueMan_Forbid {
		maxRound := len(csp.queue)
		for i := 0; i < maxRound; i++ {
			if len(csp.queue) == 0 { //when the queue is empty,end the circulate
				break
			}
			var trueMan *Player
			for _, value := range csp.queue {
				if !value.IsRob {
					trueMan = value
					break
				}
			}
			if trueMan == nil { //no true man in queue,end the circulate
				break
			}
			var gameNum = GetGameSuiableNum(int(csp.dbGameFree.GetGameId()), MatchTrueMan_Forbid)
			var member []*Player
			member = append(member, trueMan)
			for _, value := range csp.queue {
				if value.IsRob {
					member = append(member, value)
					if len(member) == gameNum {
						break
					}
				}
			}
			gameStartMinNum := GetGameStartMinNum(int(csp.dbGameFree.GetGameId()))
			if len(member) >= gameStartMinNum && csp.IsWaitThreeSecond(member) {
				csp.startCoinScene(trueMan, member)
			} else {
				if csp.dbGameFree.GetBot() != 0 {
					truePlayerNum := 0
					for _, value := range csp.queue {
						if !value.IsRob {
							truePlayerNum++
						}
					}
					for i := 0; i < truePlayerNum; i++ {
						csp.robRequireNum += GetGameSuiableNum(int(csp.dbGameFree.GetGameId()), csp.dbGameFree.GetMatchTrueMan())
					}
				}
				break
			}
		}
		return
	}
	var gameNum = GetGameSuiableNum(int(csp.dbGameFree.GetGameId()), MatchTrueMan_Forbid)
	sameIpLimit := !model.GameParamData.SameIpNoLimit
	if sameIpLimit && csp.dbGameFree.GetSameIpLimit() == 0 {
		sameIpLimit = false
	}
	maxRound := len(csp.queue)
	for i := 0; i < maxRound; i++ {
		if len(csp.queue) == 0 {
			break
		}
		var trueMan *Player
		for _, value := range csp.queue {
			if !value.IsRob {
				trueMan = value
				break
			}
		}
		if trueMan == nil { //no true man in queue,end the circulate
			break
		}
		var member []*Player
		member = append(member, trueMan)
		for _, value := range csp.queue {
			if value.SnId == trueMan.SnId {
				continue
			}
			//规避同ip的用户在一个房间内(GM除外)
			if sameIpLimit && trueMan.GMLevel == 0 && trueMan.Ip == value.Ip {
				continue
			}
			if trueMan.WhiteLevel > 0 && value.BlackLevel > 0 {
				continue
			}
			if trueMan.BlackLevel > 0 && value.WhiteLevel > 0 {
				continue
			}
			//多少局只能禁止再配对
			if csp.dbGameFree.GetSamePlaceLimit() > 0 && sceneLimitMgr.LimitSamePlaceBySnid(member, value,
				csp.dbGameFree.GetGameId(), csp.dbGameFree.GetSamePlaceLimit()) {
				continue
			}
			member = append(member, value)
			if len(member) == gameNum {
				break
			}
		}
		if IsRegularNum(int(csp.dbGameFree.GetGameId())) {
			if len(member) == gameNum {
				csp.startCoinScene(trueMan, member)
			}
		} else {
			gameStartMinNum := GetGameStartMinNum(int(csp.dbGameFree.GetGameId()))
			if len(member) >= gameStartMinNum && csp.IsWaitThreeSecond(member) {
				csp.startCoinScene(trueMan, member)
			}
		}
	}
	if csp.dbGameFree.GetBot() != 0 {
		for _, value := range csp.queue { //invite rob
			if !value.IsRob {
				csp.robRequireNum += GetGameSuiableNum(int(csp.dbGameFree.GetGameId()), csp.dbGameFree.GetMatchTrueMan())
			}
		}
	}
}
func (csp *CoinScenePool) IsWaitThreeSecond(member []*Player) bool {
	if int(csp.dbGameFree.GetGameId()) == common.GameId_BlackJack && len(member) == 1 {
		//21点特殊处理
		//匹配到的人 有一个真人等待时间超过3秒 就开始游戏
		for _, qp := range member {
			if qp != nil && !qp.IsRob {
				if time.Now().Sub(qp.EnterQueueTime) >= time.Second*3 {
					return true
				}
			}
		}
	} else {
		return true
	}
	return false
}

func (csp *CoinScenePool) UpdateAndCleanQueue() {
	now := time.Now().Unix()
	for key, p := range csp.queue {
		if p == nil {
			csp.QuitQueue(key)
			continue
		}
		if p.IsRob {
			continue
		}
		if now < p.EnterCoinSceneQueueTs {
			continue
		}
		p.EnterCoinSceneQueueTs = time.Now().Add(time.Second * CoinSceneRoundTime).Unix()
		p.CoinSceneQueueRound++
		//sync state
		if p.CoinSceneQueueRound < CoinSceneOverTime {
			pack := &gamehall_proto.SCCoinSceneQueueState{
				GameFreeId: proto.Int32(csp.dbGameFree.GetId()),
				Count:      proto.Int32(p.CoinSceneQueueRound),
				Ts:         proto.Int64(CoinSceneRoundTime),
			}
			p.SendToClient(int(gamehall_proto.CoinSceneGamePacketID_PACKET_SC_COINSCENE_QUEUESTATE), pack)
		} else {
			csp.QuitQueue(p.SnId)
			pack := &gamehall_proto.SCCoinSceneOp{
				OpCode: gamehall_proto.OpResultCode_OPRC_CoinSceneEnterQueueOverTime,
				Id:     proto.Int32(csp.dbGameFree.GetId()),
				OpType: proto.Int32(common.CoinSceneOp_Leave),
			}
			p.SendToClient(int(gamehall_proto.CoinSceneGamePacketID_PACKET_SC_COINSCENE_OP), pack)
		}
	}
}
func (csp *CoinScenePool) startCoinScene(p *Player, queue []*Player) {
	scene := csp.CreateNewScene(p)
	if scene != nil {
		csp.scenes[scene.sceneId] = scene
		for _, value := range queue {
			if value.EnterScene(scene, false, -1) {
				csp.OnPlayerEnter(value, scene)
				CoinSceneMgrSington.OnPlayerEnter(value, csp.dbGameFree.GetId())
				csp.QuitQueue(value.SnId)
			} else {
				logger.Logger.Error("Queue member enter coin scene failed.")
			}
		}
		pack := &server_proto.WGGameForceStart{
			SceneId: proto.Int(scene.sceneId),
		}
		scene.SendToGame(int(server_proto.SSPacketID_PACKET_WG_GAMEFORCESTART), pack)
	} else {
		logger.Logger.Errorf("Create %v scene in coin scene queue failed.", csp.id)
	}
}
