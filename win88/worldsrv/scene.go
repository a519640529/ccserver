package main

import (
	"math"
	"time"

	"github.com/idealeak/goserver/core/netlib"

	"math/rand"

	"games.yol.com/win88/common"
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	hall_proto "games.yol.com/win88/protocol/gamehall"
	player_proto "games.yol.com/win88/protocol/player"
	server_proto "games.yol.com/win88/protocol/server"
	"games.yol.com/win88/srvdata"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/timer"
	"github.com/idealeak/goserver/srvlib"
	srvlibproto "github.com/idealeak/goserver/srvlib/protocol"
	rawproto "google.golang.org/protobuf/proto"
)

const (
	MatchSceneState_Waiting = iota //等待状态
	MatchSceneState_Running        //进行状态
	MatchSceneState_Billed         //结算状态
)

const (
	PLAYER_HISTORY_MODEL = iota + 1
	BIGWIN_HISTORY_MODEL
	GAME_HISTORY_MODEL
)

type ApplyInfo struct {
	applyTimer  timer.TimerHandle
	applyStatus map[int32]bool
	expectCnt   int
}

type PlayerGameCtx struct {
	takeCoin             int64 //进房时携带的金币量
	enterTs              int64 //进入时间
	totalConvertibleFlow int64 //进房时玩家身上的总流水
}

// 场景
type Scene struct {
	sceneId         int                       //场景id
	gameId          int                       //游戏id
	gameMode        int                       //游戏模式
	sceneMode       int                       //房间模式,参考common.SceneMode_XXX
	params          []int32                   //场景参数
	paramsEx        []int32                   //其他扩展参数
	playerNum       int                       //人数
	robotNum        int                       //机器人数量
	robotLimit      int                       //最大限制机器人数量
	preInviteRobNum int                       //准备邀请机器人的数量
	creator         int32                     //创建者账号id
	agentor         int32                     //代理者id
	replayCode      string                    //回放码
	currRound       int32                     //当前第几轮
	totalRound      int32                     //总共几轮
	clycleTimes     int32                     //循环次数
	deleting        bool                      //正在删除
	starting        bool                      //正在开始
	closed          bool                      //房间已关闭
	force           bool                      //强制删除
	hadCost         bool                      //是否已经扣过房卡
	inTeahourse     bool                      //是否在棋牌馆
	players         map[int32]*Player         //玩家
	audiences       map[int32]*Player         //观众
	seats           [9]*Player                //座位
	gameSess        *GameSession              //所在gameserver
	sp              ScenePolicy               //场景上的一些业务策略
	createTime      time.Time                 //创建时间
	lastTime        time.Time                 //最后活跃时间
	startTime       time.Time                 //开始时间
	dirty           bool                      //脏标记
	joinList        map[int32]*ApplyInfo      //等待加入的用户列表
	applyTimes      map[int32]int32           //申请坐下次数
	limitPlatform   *Platform                 //限制平台
	groupId         int32                     //组id
	hallId          int32                     //厅id
	state           int32                     //场景当前状态
	fishing         int32                     //渔场的鱼潮状态
	gameCtx         map[int32]*PlayerGameCtx  //进入房间的环境
	dbGameFree      *server_proto.DB_GameFree //
	ClubId          int32
	clubRoomID      string  //俱乐部包间ID
	clubRoomPos     int32   //
	clubRoomTax     int32   //
	createFee       int32   //创建房间的费用
	manualDelete    bool    //是否手动解散
	GameLog         []int32 //游戏服务器同步的录单
	JackPotFund     int64   //游戏服务器同步的奖池
	State           int32   //当前游戏状态，后期放到ScenePolicy里去处理
	StateTs         int64   //切换到当前状态的时间
	StateSec        int32   //押注状态的秒数
	BankerListNum   int32   //庄家列表数量
	matchParams     []int32 //比赛参数
	matchState      int     //比赛状态
	quitMatchSnids  []int32 //退赛玩家id
	gameSite        int     //tienlen游戏场次区分 1.初级 2.中级 3.高级场
	BaseScore       int32   //tienlen游戏底分
	matchId         int32   //比赛场id
	matchStop       bool    //停止比赛
}

func NewScene(agentor, creator int32, id, gameId, gameMode, sceneMode int, clycleTimes, numOfGames int32, params []int32,
	gs *GameSession, limitPlatform *Platform, groupId int32, dbGameFree *server_proto.DB_GameFree, paramsEx ...int32) *Scene {
	sp := GetScenePolicy(gameId, gameMode)
	if sp == nil {
		logger.Logger.Errorf("NewScene sp == nil, gameId=%v gameMode=%v", gameId, gameMode)
		return nil
	}
	s := &Scene{
		sceneId:       id,
		hallId:        dbGameFree.Id,
		playerNum:     0,
		creator:       creator,
		agentor:       agentor,
		gameId:        gameId,
		gameMode:      gameMode,
		sceneMode:     sceneMode,
		params:        params,
		paramsEx:      paramsEx,
		clycleTimes:   clycleTimes,
		players:       make(map[int32]*Player),
		audiences:     make(map[int32]*Player),
		gameSess:      gs,
		sp:            sp,
		createTime:    time.Now(),
		joinList:      make(map[int32]*ApplyInfo),
		limitPlatform: limitPlatform,
		groupId:       groupId,
		gameCtx:       make(map[int32]*PlayerGameCtx), //进入房间的环境
		dbGameFree:    dbGameFree,
		currRound:     0,
		totalRound:    numOfGames,
	}
	s.playerNum = int(sp.GetPlayerNum(s))
	s.lastTime = s.createTime

	if s.IsHallScene() || s.IsCoinScene() {
		code := SceneMgrSington.AllocReplayCode()
		s.replayCode = code
	}
	if s.dbGameFree.GetMatchMode() == 0 {
		s.RandRobotCnt()
	}
	if s.IsMatchScene() {
		s.BaseScore = 10
	}
	s.sp.OnStart(s)
	return s
}

func NewLocalGameScene(creator int32, sceneId, gameId, gameSite, sceneMode int, clycleTimes int32, params []int32,
	gs *GameSession, limitPlatform *Platform, playerNum int, dbGameFree *server_proto.DB_GameFree, baseScore int32, paramsEx ...int32) *Scene {
	sp := GetScenePolicy(gameId, 0)
	if sp == nil {
		logger.Logger.Errorf("NewLocalGameScene sp == nil, gameId=%v ", gameId)
		return nil
	}
	s := &Scene{
		sceneId:       sceneId,
		hallId:        dbGameFree.Id,
		playerNum:     playerNum,
		creator:       creator,
		gameId:        gameId,
		sceneMode:     sceneMode,
		params:        params,
		paramsEx:      paramsEx,
		clycleTimes:   clycleTimes,
		players:       make(map[int32]*Player),
		audiences:     make(map[int32]*Player),
		gameSess:      gs,
		sp:            sp,
		createTime:    time.Now(),
		joinList:      make(map[int32]*ApplyInfo),
		limitPlatform: limitPlatform,
		gameCtx:       make(map[int32]*PlayerGameCtx), //进入房间的环境
		dbGameFree:    dbGameFree,
		gameSite:      gameSite,
		BaseScore:     baseScore,
	}
	s.lastTime = s.createTime

	code := SceneMgrSington.AllocReplayCode()
	s.replayCode = code
	s.sp.OnStart(s)
	return s
}

func (this *Scene) RebindPlayerSnId(oldSnId, newSnId int32) {
	if this.creator == oldSnId {
		this.creator = newSnId
	}
	if this.agentor == oldSnId {
		this.agentor = newSnId
	}
	if p, exist := this.players[oldSnId]; exist {
		delete(this.players, oldSnId)
		this.players[newSnId] = p
	}
	if p, exist := this.audiences[oldSnId]; exist {
		delete(this.audiences, oldSnId)
		this.audiences[newSnId] = p
	}
}

func (this *Scene) RobotIsLimit() bool {
	if this.robotLimit != 0 {
		if this.robotNum >= this.robotLimit {
			return true
		}
	}
	return false
}

func (this *Scene) PlayerEnter(p *Player, pos int, ischangeroom bool) bool {
	logger.Logger.Infof("(this *Scene:%v) PlayerEnter(%v, %v) ", this.sceneId, p.SnId, pos)
	//if !(this.IsMatchScene() && p.matchCtx != nil && p.matchCtx.isQuit) && p.scene != nil {
	//	logger.Logger.Warnf("(this *Scene:%v) PlayerEnter(%v, %v) found in sceneid:%v", this.sceneId, p.SnId, pos, p.scene.sceneId)
	//	return false
	//}

	if p.IsRob {
		if this.robotLimit != 0 {
			if !model.GameParamData.IsRobFightTest {
				//增加所有机器人对战场的
				if this.robotNum+1 > this.robotLimit {
					logger.Logger.Warnf("(this *Scene:%v) PlayerEnter(%v) robot num limit(%v)", this.sceneId, p.SnId, this.robotLimit)
					return false
				}
			}
		}
	}

	if !this.IsHundredScene() {
		if pos != -1 {
			if this.seats[pos] != nil {
				for i := 0; i < this.playerNum; i++ {
					if this.seats[i] == nil {
						p.pos = i
						this.seats[i] = p
						break
					}
				}
			} else {
				p.pos = pos
				this.seats[pos] = p
			}
			//delete(this.audiences, p.SnId)
		} else {
			for i := 0; i < this.playerNum; i++ {
				if this.seats[i] == nil {
					p.pos = i
					this.seats[i] = p
					break
				}
			}
		}
	}
	if p.CoinSceneQueue != nil {
		p.CoinSceneQueue.QuitQueue(p.SnId)
		//TODO send msg to rob
	}

	//if this.IsMatchScene() && p.matchCtx != nil && p.matchCtx.isQuit {
	//	//已退赛
	//	this.quitMatchSnids = append(this.quitMatchSnids, p.SnId)
	//} else {
	p.scene = this
	this.players[p.SnId] = p

	this.gameSess.AddPlayer(p)
	SceneMgrSington.OnPlayerEnterScene(this, p)
	//NpcServerAgentSington.OnPlayerEnterScene(this, p)
	//}
	if this.IsCoinScene() {
		CoinSceneMgrSington.OnPlayerEnter(p, int32(this.sceneId))
	} else if this.IsHundredScene() {
		HundredSceneMgrSington.OnPlayerEnter(p, this.paramsEx[0])
	} else if this.IsHallScene() {
		PlatformMgrSington.OnPlayerEnterScene(this, p)
	} else if this.IsPrivateScene() {
		//if this.ClubId > 0 {
		//	ClubSceneMgrSington.OnPlayerEnter(p, int32(this.sceneId))
		//}
	} else if this.IsMatchScene() {
		//MatchMgrSington.OnPlayerEnterScene(this, p)
		CoinSceneMgrSington.OnPlayerEnter(p, int32(this.sceneId))
	}

	if !this.IsMatchScene() {
		//如果正在等待比赛,退赛
		isWaiting, tmid := TournamentMgr.IsMatchWaiting(p.SnId)
		if isWaiting {
			TournamentMgr.CancelSignUp(tmid, p.SnId, true)
		}
	}

	takeCoin := p.Coin
	leaveCoin := int64(0)
	gameTimes := rand.Int31n(100)
	matchParams := []int32{}

	if this.IsMatchScene() && p.matchCtx != nil && p.matchCtx.tm != nil {
		takeCoin = int64(p.matchCtx.grade)
		matchParams = append(matchParams, p.matchCtx.rank) //排名
		ms := MatchSeasonMgrSington.GetMatchSeason(p.SnId)
		if ms != nil {
			matchParams = append(matchParams, ms.Lv) //段位
		} else {
			if p.IsRob {
				robotRandLv := MatchSeasonRankMgrSington.CreateRobotLv()
				matchParams = append(matchParams, robotRandLv) //机器人随机段位
			} else {
				matchParams = append(matchParams, 1) //段位默认值
			}
		}
	} else {
		if p.IsRob {
			if len(this.paramsEx) > 0 { //机器人携带金币动态调整
				gps := PlatformMgrSington.GetGameConfig(this.limitPlatform.IdStr, this.paramsEx[0])
				if gps != nil {
					dbGameFree := gps.DbGameFree
					if gps.GroupId != 0 {
						pgg := PlatformGameGroupMgrSington.GetGameGroup(gps.GroupId)
						if pgg != nil {
							dbGameFree = pgg.DbGameFree
						}
					}

					flag := false
					if common.IsLocalGame(this.gameId) {
						baseScore := this.BaseScore
						arrs := srvdata.PBDB_CreateroomMgr.Datas.Arr
						tmpIds := []int32{}
						for i := 0; i < len(arrs); i++ {
							arr := arrs[i]
							if int(arr.GameId) == this.gameId && int(arr.GameSite) == this.gameSite {
								betRange := arr.GetBetRange()
								if len(betRange) == 0 {
									continue
								}
								for j := 0; j < len(betRange); j++ {
									if betRange[j] == baseScore && len(arr.GetGoldRange()) > 0 && arr.GetGoldRange()[0] != 0 {
										tmpIds = append(tmpIds, arr.GetId())
										break
									}
								}
							}
						}
						if len(tmpIds) > 0 {
							randId := common.RandInt32Slice(tmpIds)
							crData := srvdata.PBDB_CreateroomMgr.GetData(randId)
							if crData != nil {
								goldRange := crData.GetGoldRange()
								if len(goldRange) == 2 {
									takeCoin = common.RandFromRangeInt64(int64(goldRange[0]), int64(goldRange[1]))
									flag = true
								} else if len(goldRange) == 1 {
									takeCoin = common.RandFromRangeInt64(int64(goldRange[0]), 2*int64(goldRange[0]))
									flag = true
								}
								leaveCoin = int64(goldRange[0])
								for _, id := range tmpIds {
									tmp := srvdata.PBDB_CreateroomMgr.GetData(id).GetGoldRange()
									if int64(tmp[0]) < leaveCoin && tmp[0] != 0 {
										leaveCoin = int64(tmp[0])
									}
								}
							}
						} else {
							logger.Logger.Warn("gameId: ", this.gameId, " gameSite: ", this.gameSite, " baseScore: ", baseScore)
						}
						if leaveCoin > takeCoin {
							logger.Logger.Warn("robotSnId: ", p.SnId, " baseScore: ", baseScore, " takeCoin: ", takeCoin, " leaveCoin: ", leaveCoin)
						}
						if takeCoin > p.Coin {
							p.Coin = takeCoin
						}
					}
					if !flag && this.IsCoinScene() && !this.IsTestScene() {
						if expectEnterCoin, expectLeaveCoin, ExpectGameTime, ok := RobotCarryMgrEx.RandOneCarry(dbGameFree.GetId()); ok && expectEnterCoin > dbGameFree.GetLimitCoin() && expectEnterCoin < dbGameFree.GetMaxCoinLimit() {
							takeCoin = int64(expectEnterCoin)
							leaveCoin = int64(expectLeaveCoin)
							//如果带入金币和离开金币比较接近，就调整离开金币值
							var delta = takeCoin - leaveCoin
							if math.Abs(float64(delta)) < float64(takeCoin/50) {
								if leaveCoin = takeCoin + delta*(10+rand.Int63n(50)); leaveCoin < 0 {
									leaveCoin = 0
								}
							}
							gameTimes = ExpectGameTime * 2
							flag = true
						}
					}

					if !flag {
						takerng := dbGameFree.GetRobotTakeCoin()
						if len(takerng) >= 2 && takerng[1] > takerng[0] {
							if takerng[0] < dbGameFree.GetLimitCoin() {
								takerng[0] = dbGameFree.GetLimitCoin()
							}
							takeCoin = int64(common.RandInt(int(takerng[0]), int(takerng[1])))
						} else {
							maxlimit := int64(dbGameFree.GetMaxCoinLimit())
							if maxlimit != 0 && p.Coin > maxlimit {
								logger.Logger.Trace("Player coin:", p.Coin)
								//在下限和上限之间随机，并对其的100的整数倍
								takeCoin = int64(common.RandInt(int(dbGameFree.GetLimitCoin()), int(maxlimit)))
								logger.Logger.Trace("Take coin:", takeCoin)
							}
							if maxlimit == 0 && this.IsCoinScene() {
								maxlimit = int64(common.RandInt(10, 50)) * int64(dbGameFree.GetLimitCoin())
								takeCoin = int64(common.RandInt(int(dbGameFree.GetLimitCoin()), int(maxlimit)))
								logger.Logger.Trace("Take coin:", takeCoin)
							}
						}
						takeCoin = takeCoin / 100 * 100
						//离场金币
						leaverng := dbGameFree.GetRobotLimitCoin()
						if len(leaverng) >= 2 {
							leaveCoin = int64(leaverng[0] + rand.Int31n(leaverng[1]-leaverng[0]))
						}
					}
					bankerLimit := this.dbGameFree.GetBanker()
					if bankerLimit != 0 {
						//上庄AI携带
						if /*this.gameId == common.GameId_HundredBull ||*/
						this.gameId == common.GameId_RollCoin ||
							this.gameId == common.GameId_RollAnimals ||
							this.gameId == common.GameId_DragonVsTiger ||
							this.gameId == common.GameId_Baccarat {
							if this.BankerListNum < 3 {
								if rand.Intn(100) < 5 {
									randCoin := int64(math.Floor(float64(bankerLimit) * 1 / float64(this.dbGameFree.GetSceneType())))
									takeCoin = rand.Int63n(randCoin) + int64(bankerLimit)
									if takeCoin > p.Coin {
										//加钱速度慢 暂时不用
										//ExePMCmd(p.gateSess, fmt.Sprintf("%v%v%v", common.PMCmd_AddCoin, common.PMCmd_SplitToken, takeCoin))
									}
								}
							}
						}
					}
					if takeCoin > p.Coin {
						takeCoin = p.Coin
					}
				}
			}
		}
	}

	if p.IsRob {
		this.robotNum++
		p.RobotRandName()
		if !this.IsMatchScene() {
			p.RobRandVipWhenEnterRoom(takeCoin)
		}
		//else if p.matchCtx != nil {
		//	p.RobRandVipWhenEnterMatch(p.matchCtx.m.dbMatch.VipLimit)
		//}
		name := this.GetSceneName()
		logger.Logger.Tracef("(this *Scene) PlayerEnter(%v) robot(%v) robotlimit(%v)", name, this.robotNum, this.robotLimit)
	}

	//todo:send add msg to gamesrv
	data, err := p.MarshalData(this.gameId)
	if err == nil {
		var gateSid int64
		if p.gateSess != nil {
			if srvInfo, ok := p.gateSess.GetAttribute(srvlib.SessionAttributeServerInfo).(*srvlibproto.SSSrvRegiste); ok && srvInfo != nil {
				sessionId := srvlib.NewSessionIdEx(srvInfo.GetAreaId(), srvInfo.GetType(), srvInfo.GetId(), 0)
				gateSid = sessionId.Get()
			}
		}
		isQuMin := false
		//if !p.IsRob {
		//	pt := PlatformMgrSington.GetPackageTag(p.PackageID)
		//	if pt != nil && pt.SpreadTag == 1 {
		//		isQuMin = true
		//	}
		//}
		msg := &server_proto.WGPlayerEnter{
			Sid:        proto.Int64(p.sid),
			SnId:       proto.Int32(p.SnId),
			GateSid:    proto.Int64(gateSid),
			SceneId:    proto.Int(this.sceneId),
			PlayerData: data,
			IsLoaded:   proto.Bool(ischangeroom),
			IsQM:       proto.Bool(isQuMin),
			IParams:    p.MarshalIParam(),
			SParams:    p.MarshalSParam(),
			CParams:    p.MarshalCParam(),
		}
		sa, err2 := p.MarshalSingleAdjustData(this.dbGameFree.Id)
		if err2 == nil && sa != nil {
			msg.SingleAdjust = sa
		}
		if this.ClubId != 0 {

		}
		p.takeCoin = takeCoin
		p.sceneCoin = takeCoin
		p.enterts = time.Now()

		if !p.IsRob { //保存下进入时的环境
			this.gameCtx[p.SnId] = &PlayerGameCtx{
				takeCoin:             p.takeCoin,
				enterTs:              p.enterts.Unix(),
				totalConvertibleFlow: p.TotalConvertibleFlow,
			}
		}
		msg.TakeCoin = proto.Int64(takeCoin)
		msg.ExpectLeaveCoin = proto.Int64(leaveCoin)
		msg.ExpectGameTimes = proto.Int32(gameTimes)
		msg.Pos = proto.Int(p.pos)
		if matchParams != nil {
			for _, param := range matchParams {
				msg.MatchParams = append(msg.MatchParams, param)
			}
		}
		snid := p.SnId
		dbItemArr := srvdata.PBDB_GameItemMgr.Datas.Arr
		if dbItemArr != nil {
			msg.Items = make(map[int32]int32)
			for _, dbItem := range dbItemArr {
				msg.Items[dbItem.Id] = proto.Int32(0)
				itemInfo := BagMgrSington.GetBagItemById(snid, dbItem.Id)
				if itemInfo != nil {
					msg.Items[dbItem.Id] = proto.Int32(itemInfo.ItemNum)
				}
			}
		}
		proto.SetDefaults(msg)
		this.SendToGame(int(server_proto.SSPacketID_PACKET_WG_PLAYERENTER), msg)
		logger.Logger.Tracef("SSPacketID_PACKET_WG_PLAYERENTER Scene:%v ;PlayerEnter(%v, %v)", this.sceneId, p.SnId, pos)
		this.lastTime = time.Now()
		return true
	} else {
		logger.Logger.Warnf("(this *Scene:%v) PlayerEnter(%v, %v) Marshal player data error %v", this.sceneId, p.SnId, pos, err)
		this.DelPlayer(p)
		return false
	}
}
func ExePMCmd(s *netlib.Session, cmd string) {
	CSPMCmd := &player_proto.CSPMCmd{
		Cmd: proto.String(cmd),
	}
	proto.SetDefaults(CSPMCmd)
	logger.Logger.Trace("CSPMCmd:", CSPMCmd)
	s.Send(int(player_proto.PlayerPacketID_PACKET_CS_PMCMD), CSPMCmd)
}
func (this *Scene) GetPlayerGameCtx(snid int32) *PlayerGameCtx {
	if ctx, exist := this.gameCtx[snid]; exist {
		return ctx
	}
	return nil
}

func (this *Scene) PlayerLeave(p *Player, reason int) {
	logger.Logger.Infof("(this *Scene:%v) PlayerLeave(%v, %v) ", this.sceneId, p.SnId, reason)
	//if !this.IsMatchScene() {
	//pack := &hall_proto.SCLeaveRoom{
	//	Reason:    proto.Int(reason),
	//	OpRetCode: hall_proto.OpResultCode_Game_OPRC_Sucess_Game,
	//	Mode:      proto.Int(0),
	//	RoomId:    proto.Int(this.sceneId),
	//}
	//proto.SetDefaults(pack)
	//p.SendToClient(int(hall_proto.GameHallPacketID_PACKET_SC_LEAVEROOM), pack)
	pack := &hall_proto.SCQuitGame{
		Id:     int32(this.dbGameFree.Id),
		Reason: proto.Int(reason),
	}
	pack.OpCode = hall_proto.OpResultCode_Game_OPRC_Sucess_Game
	proto.SetDefaults(pack)
	p.SendToClient(int(hall_proto.GameHallPacketID_PACKET_SC_QUITGAME), pack)
	//}

	//其他人直接从房间退出来
	this.DelPlayer(p)

	this.lastTime = time.Now()

}

func (this *Scene) DelPlayer(p *Player) bool {
	if p.scene != this {
		inroomid := 0
		if p.scene != nil {
			inroomid = p.scene.sceneId
		}
		logger.Logger.Warnf("(this *Scene) DelPlayer found player:%v in room:%v but room:%v", p.SnId, inroomid, this.sceneId)
	}
	if this.gameSess != nil {
		this.gameSess.DelPlayer(p)
	}
	delete(this.players, p.SnId)
	if !p.IsRob {
		delete(this.gameCtx, p.SnId)
	}

	p.scene = nil
	SceneMgrSington.OnPlayerLeaveScene(this, p)

	if !this.IsHundredScene() {
		if this.IsHallScene() {
			PlatformMgrSington.OnPlayerLeaveScene(this, p)
		} else if this.IsCoinScene() {
			CoinSceneMgrSington.OnPlayerLeave(this, p)
		}
		for i := 0; i < this.playerNum; i++ {
			if this.seats[i] == p {
				p.pos = -1
				this.seats[i] = nil
				break
			}
		}
	} else {
		HundredSceneMgrSington.OnPlayerLeave(p)
	}
	if p.IsRob {
		this.robotNum--
		name := this.GetSceneName()
		logger.Logger.Tracef("(this *Scene) PlayerLeave(%v) robot(%v) robotlimit(%v)", name, this.robotNum, this.robotLimit)
	}
	//from gameserver, so don't need send msg
	return true
}

func (this *Scene) AudienceEnter(p *Player, ischangeroom bool) bool {
	logger.Logger.Infof("(this *Scene:%v) AudienceEnter(%v) ", this.sceneId, p.SnId)
	p.scene = this
	p.applyPos = -1
	this.audiences[p.SnId] = p
	this.gameSess.AddPlayer(p)
	SceneMgrSington.OnPlayerEnterScene(this, p)
	if this.IsHundredScene() {
		HundredSceneMgrSington.OnPlayerEnter(p, this.paramsEx[0])
	}
	//todo:send add msg to gamesrv
	data, err := p.MarshalData(this.gameId)
	if err == nil {
		var gateSid int64
		if p.gateSess != nil {
			if srvInfo, ok := p.gateSess.GetAttribute(srvlib.SessionAttributeServerInfo).(*srvlibproto.SSSrvRegiste); ok && srvInfo != nil {
				sessionId := srvlib.NewSessionIdEx(srvInfo.GetAreaId(), srvInfo.GetType(), srvInfo.GetId(), 0)
				gateSid = sessionId.Get()
			}
		}
		isQuMin := false
		//if !p.IsRob {
		//	pt := PlatformMgrSington.GetPackageTag(p.PackageID)
		//	if pt != nil && pt.SpreadTag == 1 {
		//		isQuMin = true
		//	}
		//}
		msg := &server_proto.WGPlayerEnter{
			Sid:        proto.Int64(p.sid),
			SnId:       proto.Int32(p.SnId),
			GateSid:    proto.Int64(gateSid),
			SceneId:    proto.Int(this.sceneId),
			PlayerData: data,
			IsLoaded:   proto.Bool(ischangeroom),
			IsQM:       proto.Bool(isQuMin),
			IParams:    p.MarshalIParam(),
			SParams:    p.MarshalSParam(),
			CParams:    p.MarshalCParam(),
		}

		if !p.IsRob { //保存下进入时的环境
			this.gameCtx[p.SnId] = &PlayerGameCtx{
				takeCoin:             p.takeCoin,
				enterTs:              p.enterts.Unix(),
				totalConvertibleFlow: p.TotalConvertibleFlow,
			}
		}

		takeCoin := p.Coin
		p.takeCoin = takeCoin
		msg.TakeCoin = proto.Int64(takeCoin)
		proto.SetDefaults(msg)
		this.SendToGame(int(server_proto.SSPacketID_PACKET_WG_AUDIENCEENTER), msg)
		p.enterts = time.Now()
		this.lastTime = time.Now()
		return true
	}

	return false
}

func (this *Scene) AudienceLeave(p *Player, reason int) {
	logger.Logger.Infof("(this *Scene:%v) AudienceLeave(%v, %v) ", this.sceneId, p.SnId, reason)
	pack := &hall_proto.SCLeaveRoom{
		Reason:    proto.Int(reason),
		OpRetCode: hall_proto.OpResultCode_Game_OPRC_Sucess_Game,
		Mode:      proto.Int(0),
		RoomId:    proto.Int(this.sceneId),
	}
	proto.SetDefaults(pack)
	p.SendToClient(int(hall_proto.GameHallPacketID_PACKET_SC_LEAVEROOM), pack)
	//观众直接从房间退出来
	this.DelAudience(p)
	this.lastTime = time.Now()
}

func (this *Scene) DelAudience(p *Player) bool {
	logger.Logger.Infof("(this *Scene:%v) DelAudience(%v) ", this.sceneId, p.SnId)
	if p.scene != this {
		return false
	}
	if this.gameSess != nil {
		this.gameSess.DelPlayer(p)
	}
	delete(this.audiences, p.SnId)
	if !p.IsRob {
		delete(this.gameCtx, p.SnId)
	}
	p.scene = nil
	p.applyPos = -1
	SceneMgrSington.OnPlayerLeaveScene(this, p)
	if this.IsHundredScene() {
		HundredSceneMgrSington.OnPlayerLeave(p)
	}
	//from gameserver, so don't need send msg
	return true
}

//观众坐下
//func (this *Scene) AudienceSit(p *Player, pos int) bool {
//	logger.Logger.Infof("(this *Scene:%v) AudienceSit(%v, %v, %v) ", this.sceneId, p.SnId, pos)
//	if _, exist := this.audiences[p.SnId]; exist {
//		if pos == -1 && !this.IsHundredScene() { //自动匹配;百人场没座位概念
//			for i := 0; i < this.playerNum; i++ {
//				if this.seats[i] == nil {
//					pos = i
//					break
//				}
//			}
//		}
//		if pos != -1 || this.IsHundredScene() {
//			if !this.IsHundredScene() {
//				if this.seats[pos] != nil {
//					return false
//				}
//				p.pos = pos
//				p.applyPos = -1
//				this.seats[pos] = p
//			}
//			delete(this.audiences, p.SnId)
//		}
//
//		p.scene = this
//		this.players[p.SnId] = p
//
//		NpcServerAgentSington.OnPlayerEnterScene(this, p)
//		if this.IsCoinScene() {
//			CoinSceneMgrSington.OnPlayerEnter(p, int32(this.sceneId))
//		} else if this.IsHallScene() {
//			PlatformMgrSington.OnPlayerEnterScene(this, p)
//		}
//
//		msg := &protocol.WGAudienceSit{
//			SnId:    proto.Int32(p.SnId),
//			SceneId: proto.Int(this.sceneId),
//			Pos:     proto.Int(pos),
//		}
//		p.takeCoin = p.Coin
//		msg.TakeCoin = proto.Int64(p.Coin)
//		proto.SetDefaults(msg)
//		this.SendToGame(int(protocol.MmoPacketID_PACKET_WG_AUDIENCESIT), msg)
//		this.lastTime = time.Now()
//		return true
//	}
//	return false
//}

func (this *Scene) AudienceSit(p *Player, pos int) bool {
	logger.Logger.Infof("(this *Scene:%v) AudienceSit(%v, %v, %v) ", this.sceneId, p.SnId, pos, this.dbGameFree.GetId())
	if _, exist := this.audiences[p.SnId]; exist {
		delete(this.audiences, p.SnId)
		p.scene = this
		this.players[p.SnId] = p

		//NpcServerAgentSington.OnPlayerEnterScene(this, p)
		if this.IsCoinScene() {
			CoinSceneMgrSington.OnPlayerEnter(p, this.dbGameFree.GetId())
		} else if this.IsHallScene() {
			PlatformMgrSington.OnPlayerEnterScene(this, p)
		}
		msg := &server_proto.WGAudienceSit{
			SnId:    proto.Int32(p.SnId),
			SceneId: proto.Int(this.sceneId),
			Pos:     proto.Int(pos),
		}
		p.takeCoin = p.Coin
		msg.TakeCoin = proto.Int64(p.Coin)
		proto.SetDefaults(msg)
		this.SendToGame(int(server_proto.SSPacketID_PACKET_WG_AUDIENCESIT), msg)
		this.lastTime = time.Now()
		return true
	}
	return false
}

func (this *Scene) HasPlayer(p *Player) bool {
	if _, exist := this.players[p.SnId]; exist {
		return true
	}
	return false
}

func (this *Scene) HasAudience(p *Player) bool {
	if _, exist := this.audiences[p.SnId]; exist {
		return true
	}
	return false
}

func (this *Scene) GetPlayer(id int32) *Player {
	if p, exist := this.players[id]; exist {
		return p
	}
	return nil
}

func (this *Scene) GetAudience(id int32) *Player {
	if p, exist := this.audiences[id]; exist {
		return p
	}
	return nil
}

func (this *Scene) GetPlayerPos(snId int32) int {
	for index, value := range this.seats {
		if value == nil {
			continue
		}
		if value.SnId == snId {
			return index
		}
	}
	return -1
}

func (this *Scene) GetPlayerCnt() int {
	return len(this.players)
}

func (this *Scene) GetAudienceCnt() int {
	return len(this.audiences)
}

func (this *Scene) IsFull() bool {
	return this.GetPlayerCnt() >= this.playerNum
}

func (this *Scene) IsEmpty() bool {
	return this.GetPlayerCnt() == 0
}

func (this *Scene) AllIsRobot() bool {
	//for _, p := range this.players {
	//	if p != nil && !p.IsRob {
	//		return false
	//	}
	//}
	//return true
	return len(this.players) == this.robotNum
}

func (this *Scene) OnClose() {
	scDestroyRoom := &hall_proto.SCDestroyRoom{
		RoomId:    proto.Int(this.sceneId),
		OpRetCode: hall_proto.OpResultCode_Game_OPRC_Sucess_Game,
		IsForce:   proto.Int(1),
	}
	proto.SetDefaults(scDestroyRoom)
	this.Broadcast(int(hall_proto.GameHallPacketID_PACKET_SC_DESTROYROOM), scDestroyRoom, 0)

	this.closed = true
	PlatformMgrSington.OnChangeSceneState(this, this.closed)
	this.sp.OnStop(this)
	//NpcServerAgentSington.OnSceneClose(this)
	for _, p := range this.players {
		this.DelPlayer(p)
	}
	for _, p := range this.audiences {
		this.DelAudience(p)
	}
	this.players = nil
	this.audiences = nil
	this.gameSess = nil
	for _, info := range this.joinList {
		if info.applyTimer != timer.TimerHandle(0) {
			timer.StopTimer(info.applyTimer)
			info.applyTimer = timer.TimerHandle(0)
		}
	}
	this.joinList = nil
}

func (this *Scene) SendToGame(packetId int, pack interface{}) bool {
	if this.gameSess != nil {
		this.gameSess.Send(packetId, pack)
		return true
	}
	return false
}
func (this *Scene) SendToClient(packetid int, rawpack interface{}, excludeId int32) {
	for snid, value := range this.players {
		if snid == excludeId {
			continue
		}
		value.SendToClient(packetid, rawpack)
	}
}
func (this *Scene) BilledRoomCard(snid []int32) {
	if this.sp != nil {
		this.sp.BilledRoomCard(this, snid)
	}
}

func (this *Scene) GenAgentRoomCardBill(costCount int32) {
	//	var recs []*model.RoomCardRec
	//	ts := int32(time.Now().Unix())
	//	//房卡结算信息
	//	playerNum := int32(len(this.players))
	//	costCount /= playerNum
	//	costCount = (costCount + 4) / 5 * 5 //向0.5对齐
	//	for _, p := range this.players {
	//		rec := model.NewRoomCardRec()
	//		recs = append(recs, rec)
	//		rec.AgentId = this.agentor
	//		rec.SnId = p.SnId
	//		rec.RoomId = int32(this.sceneId)
	//		rec.ReplayCode = this.replayCode
	//		rec.Count = costCount
	//		rec.Ts = ts
	//	}
	//	task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
	//		err := model.InsertRoomCardRec(recs...)
	//		if err != nil {
	//			logger.Logger.Warn("InsertRoomCardRec err:", err)
	//		}
	//		return nil
	//	}), nil, "InsertRoomCardRec").StartByFixExecutor("logic_roomcardrec")
}

func (this *Scene) IsLongTimeInactive() bool {
	tNow := time.Now()
	//删除超过指定不活跃时间的房间
	if len(this.players) == 0 && tNow.Sub(this.lastTime) > time.Second*time.Duration(model.GameParamData.SceneMaxIdle) {
		return true
	}
	return false
}

func (this *Scene) ForceDelete(isManual bool) {
	this.manualDelete = isManual
	this.deleting = true
	this.force = true
	PlatformMgrSington.OnChangeSceneState(this, this.force)
	pack := &server_proto.WGDestroyScene{
		SceneId:   proto.Int(this.sceneId),
		MatchStop: proto.Bool(this.matchStop),
	}
	proto.SetDefaults(pack)
	this.SendToGame(int(server_proto.SSPacketID_PACKET_WG_DESTROYSCENE), pack)

	logger.Logger.Warnf("(this *Scene) ForceDelete() sceneid=%v", this.sceneId)
	if this.sceneId == SceneMgrSington.GetDgSceneId() {
		for _, value := range PlayerMgrSington.playerMap {
			if value.scene == nil {
				continue
			}
			if value.scene.sceneId == SceneMgrSington.GetDgSceneId() {
				value.DgGameLogout()
				pack := &hall_proto.SCLeaveDgGame{}
				pack.OpRetCode = hall_proto.OpResultCode_Game_OPRC_Sucess_Game
				value.SendToClient(int(hall_proto.GameHallPacketID_PACKET_SC_LEAVEDGGAME), pack)
			}
		}
	}
	if this.IsPreCreateScene() {
		CoinSceneMgrSington.delayCache = append(CoinSceneMgrSington.delayCache, CreateRoomCache{gameFreeId: this.dbGameFree.GetId(), platformName: this.limitPlatform.IdStr})
	}
}

func (this *Scene) Shutdown() {
	if this.hadCost && this.sp != nil {
		this.sp.OnShutdown(this)
	}
}

// 小游戏场
func (this *Scene) IsMiniGameScene() bool {
	return this.sceneId >= common.MiniGameSceneStartId && this.sceneId < common.MiniGameSceneMaxId
}

// 比赛场
func (this *Scene) IsMatchScene() bool {
	return this.sceneId >= common.MatchSceneStartId && this.sceneId < common.MatchSceneMaxId
}

// 大厅场
func (this *Scene) IsHallScene() bool {
	return this.sceneId >= common.HallSceneStartId && this.sceneId < common.HallSceneMaxId
}

// 金币场
func (this *Scene) IsCoinScene() bool {
	return this.sceneId >= common.CoinSceneStartId && this.sceneId < common.CoinSceneMaxId
}

// 百人场
func (this *Scene) IsHundredScene() bool {
	return this.sceneId >= common.HundredSceneStartId && this.sceneId < common.HundredSceneMaxId
}

// 私人房间
func (this *Scene) IsPrivateScene() bool {
	return this.sceneId >= common.PrivateSceneStartId && this.sceneId < common.PrivateSceneMaxId || this.sceneMode == common.SceneMode_Private
}

func (this *Scene) EnterCheck(p *Player) int {
	if this.limitPlatform != nil {
		if this.limitPlatform.Isolated && this.limitPlatform.IdStr != p.Platform {
			return int(hall_proto.OpResultCode_Game_OPRC_OnlyAllowClubMemberEnter_Game)
		}
	}
	sp := GetScenePolicy(this.gameId, this.gameMode)
	if sp == nil {
		return int(hall_proto.OpResultCode_Game_OPRC_GameNotExist_Game)
	}
	if reason := sp.CanEnter(this, p); reason != 0 && reason != int(hall_proto.OpResultCode_Game_OPRC_SceneEnterForWatcher_Game) {
		return int(hall_proto.OpResultCode_Game(reason))
	}
	if sp.IsFull(this, p, sp.GetPlayerNum(this)) {
		return int(hall_proto.OpResultCode_Game_OPRC_RoomIsFull_Game)
	}
	if p.applyPos != -1 {
		if this.seats[p.applyPos] != nil {
			logger.Logger.Tracef("(this *Scene) EnterCheck %v %v, seats:%v", p.SnId, p.applyPos, this.seats)
			p.applyPos = -1
			return int(hall_proto.OpResultCode_Game_OPRC_ScenePosFull_Game)
		}
	}
	return int(hall_proto.OpResultCode_Game_OPRC_Sucess_Game)
}

func (this *Scene) IsTestScene() bool {
	if this.dbGameFree != nil {
		return this.dbGameFree.GetSceneType() == -1
	}
	if len(this.paramsEx) > 0 {
		dbGameFree := srvdata.PBDB_GameFreeMgr.GetData(this.paramsEx[0])
		if dbGameFree != nil {
			return dbGameFree.GetSceneType() == -1
		}
	}
	return false
}

func (this *Scene) GetSceneName() string {
	if len(this.paramsEx) > 0 {
		dbGameFree := srvdata.PBDB_GameFreeMgr.GetData(this.paramsEx[0])
		if dbGameFree != nil {
			return dbGameFree.GetName() + dbGameFree.GetTitle()
		}
	}
	return "[unknow scene name]"
}
func (this *Scene) RandRobotCnt() {
	if len(this.paramsEx) > 0 {
		gps := PlatformMgrSington.GetGameConfig(this.limitPlatform.IdStr, this.paramsEx[0])
		if gps != nil {
			dbGameFree := gps.DbGameFree
			if gps.GroupId != 0 {
				pgg := PlatformGameGroupMgrSington.GetGameGroup(gps.GroupId)
				if pgg != nil {
					dbGameFree = pgg.DbGameFree
				}
			}
			numrng := dbGameFree.GetRobotNumRng()
			if len(numrng) >= 2 {
				if numrng[1] == numrng[0] {
					this.robotLimit = int(numrng[0])
				} else {
					if numrng[1] < numrng[0] {
						numrng[1], numrng[0] = numrng[0], numrng[1]
					}
					this.robotLimit = int(numrng[1]) //int(numrng[0] + rand.Int31n(numrng[1]-numrng[0]) + 1)
				}
			}
		}
	}
}

func (this *Scene) isPlatform(platform string) bool {
	if platform == "0" || platform == this.limitPlatform.IdStr {
		return true
	}
	return false
}

func (this *Scene) GetWhitePlayerCnt() int {
	var cnt int
	for _, p := range this.players {
		if p.WhiteLevel > 0 {
			cnt++
		}
	}
	return cnt
}

func (this *Scene) GetBlackPlayerCnt() int {
	var cnt int
	for _, p := range this.players {
		if p.BlackLevel > 0 {
			cnt++
		}
	}
	return cnt
}

func (this *Scene) GetLostPlayerCnt() int {
	var cnt int
	for _, p := range this.players {
		if p.GDatas != nil {
			if d, exist := p.GDatas[this.dbGameFree.GetGameDif()]; exist {
				if d.Statics.TotalIn > d.Statics.TotalOut {
					cnt++
				}
			}
		}
	}
	return cnt
}

func (this *Scene) GetWinPlayerCnt() int {
	var cnt int
	for _, p := range this.players {
		if p.GDatas != nil {
			if d, exist := p.GDatas[this.dbGameFree.GetGameDif()]; exist {
				if d.Statics.TotalIn < d.Statics.TotalOut {
					cnt++
				}
			}
		}
	}
	return cnt
}
func (this *Scene) GetTruePlayerCnt() int {
	//var cnt int
	//for _, p := range this.players {
	//	if !p.IsRob {
	//		cnt++
	//	}
	//}
	//return cnt
	return len(this.players) - this.robotNum
}

// 炸金花房间有几个新手
func (this *Scene) GetFoolPlayerCnt() int {
	var cnt int
	for _, p := range this.players {
		if p.IsFoolPlayer != nil && p.IsFoolPlayer[this.dbGameFree.GetGameDif()] {
			cnt++
		}
	}
	return cnt
}

func (this *Scene) GetPlayerType(gameid, gamefreeid int32) (types []int32) {
	for _, p := range this.players {
		t := int32(0)
		if p.IsRob {
			t = common.PlayerType_Rob
		} else if p.BlackLevel > 0 {
			t = common.PlayerType_Black
		} else if p.WhiteLevel > 0 {
			t = common.PlayerType_White
		} else {
			pt := p.CheckType(gameid, gamefreeid)
			if pt != nil {
				t = pt.GetId()
			} else {
				t = common.PlayerType_Undefine
			}
		}
		if !common.InSliceInt32(types, t) {
			types = append(types, t)
		}
	}
	return
}

func (this *Scene) CheckMatch(t *server_proto.DB_PlayerType) bool {
	return true
}

func (this *Scene) Broadcast(packetid int, msg rawproto.Message, excludeSid int64) {
	mgs := make(map[*netlib.Session][]*srvlibproto.MCSessionUnion)
	for _, p := range this.players {
		if p != nil {
			if p.sid != excludeSid {
				if p.gateSess != nil && p.IsOnLine() {
					mgs[p.gateSess] = append(mgs[p.gateSess], &srvlibproto.MCSessionUnion{
						Mccs: &srvlibproto.MCClientSession{
							SId: p.sid,
						},
					})
				}
			}
		}
	}
	for _, p := range this.audiences {
		if p != nil && p.sid != excludeSid {
			if p.gateSess != nil && p.IsOnLine() {
				mgs[p.gateSess] = append(mgs[p.gateSess], &srvlibproto.MCSessionUnion{
					Mccs: &srvlibproto.MCClientSession{
						SId: p.sid,
					},
				})
			}
		}
	}

	for gateSess, v := range mgs {
		if gateSess != nil && len(v) != 0 {
			pack, err := MulticastMaker.CreateMulticastPacket(packetid, msg, v...)
			if err == nil {
				proto.SetDefaults(pack)
				gateSess.Send(int(srvlibproto.SrvlibPacketID_PACKET_SS_MULTICAST), pack)
			}
		}
	}
}

func (this *Scene) HasSameIp(ip string) bool {
	for _, p := range this.players {
		if !p.IsRob {
			if p.GMLevel == 0 && p.Ip == ip {
				return true
			}
		}
	}

	return false
}

func (this *Scene) IsPreCreateScene() bool {
	return this.dbGameFree.GetCreateRoomNum() > 0
}
func (this *Scene) PlayerTryChange() {
	var member []*Player
	var player *Player
	for _, value := range this.players {
		if !value.IsRob {
			member = append(member, value)
			player = value
		}
	}
	if len(member) <= 1 {
		return
	}
	gameFreeId := this.dbGameFree.GetId()
	gameConfig := PlatformMgrSington.GetGameConfig(player.Platform, gameFreeId)
	if gameConfig != nil && gameConfig.DbGameFree.GetMatchMode() == 1 {
		return
	}
	for i := 0; i < len(member)-1; i++ {
		p := member[i]
		other := member[i+1:]
		if this.dbGameFree.GetSamePlaceLimit() > 0 && sceneLimitMgr.LimitSamePlaceBySnid(other, p,
			this.dbGameFree.GetGameId(), this.dbGameFree.GetSamePlaceLimit()) {
			if p.scene.IsPrivateScene() {
				//if ClubSceneMgrSington.PlayerInChanging(p) {
				//	continue
				//}
				//ClubSceneMgrSington.PlayerTryChange(p, gameFreeId, []int32{int32(this.sceneId)}, false)
			} else {
				if CoinSceneMgrSington.PlayerInChanging(p) {
					continue
				}
				excludeSceneIds := p.lastSceneId[gameFreeId]
				CoinSceneMgrSington.PlayerTryChange(p, gameFreeId, excludeSceneIds, false)
			}
		}

	}
}

func (this *Scene) GetParamEx(idx int) int32 {
	if idx < 0 || idx > len(this.paramsEx) {
		return -1
	}

	return this.paramsEx[idx]
}

func (this *Scene) SetParamEx(idx int, val int32) {
	cnt := len(this.paramsEx)
	if idx >= 0 && idx < cnt {
		this.paramsEx[idx] = val
	}
}
