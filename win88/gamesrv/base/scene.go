package base

import (
	"bytes"
	"games.yol.com/win88/protocol/gamehall"
	"games.yol.com/win88/protocol/player"
	"games.yol.com/win88/srvdata"
	rawproto "google.golang.org/protobuf/proto"
	"math/rand"
	"strings"
	"time"

	"fmt"

	"strconv"

	"games.yol.com/win88/common"
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/server"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/core/timer"
	"github.com/idealeak/goserver/core/utils"
	srvlibproto "github.com/idealeak/goserver/srvlib/protocol"
)

const (
	SCENE_STATE_INITED int = iota
	SCENE_STATE_RUNNING
	SCENE_STATE_OVER
)

const ReplayIdTf = "20060102150405"

var sceneRandSeed = time.Now().UnixNano()
var RobotSceneDBGameFreeSync = make(map[int]bool)

type GameScene interface {
	SceneDestroy(force bool)
}

type CanRebindSnId interface {
	RebindPlayerSnId(oldSnId, newSnId int32)
}

// 房间比赛数据变化
type SceneMatchChgData struct {
	NextBaseScore int32 //底分
	NextOutScore  int32 //淘汰分
}

type Scene struct {
	ws                 *netlib.Session
	Rand               *rand.Rand
	ExtraData          interface{}
	matchData          interface{}
	aiMgr              AIMgr
	WithLocalAI        bool
	SceneId            int
	GameId             int
	GameMode           int
	SceneMode          int
	SceneType          int
	Platform           string
	state              int
	Params             []int32
	paramsEx           []int32
	Creator            int32
	agentor            int32
	hallId             int32
	replayCode         string
	disbandGen         int               //第几次解散申请
	disbandParam       []int64           //解散参数
	disbandPos         int32             //发起解散的玩家位置
	disbandTs          int64             //解散发起时间戳
	playerNum          int               //游戏人数
	realPlayerNum      int               //真是玩家人数
	robotNum           int               //机器人数量
	robotLimit         int               //最大限制机器人数量
	robotNumLastInvite int               //上次邀请机器人时的数量
	TotalOfGames       int               //游戏总局数
	NumOfGames         int               //局数
	Players            map[int32]*Player //参与者
	audiences          map[int32]*Player //观众
	sp                 ScenePolicy       //场景游戏策略
	//mp                 MatchPolicy           //场景比赛策略
	rr               *ReplayRecorder     //回放记录器
	rrVer            int32               //录像的协议版本号
	DbGameFree       *server.DB_GameFree //自由场数据
	SceneState       SceneState          //场景状态
	hDisband         timer.TimerHandle   //解散handle
	StateStartTime   time.Time           //状态开始时间
	stateEndTime     time.Time           //状态结束时间
	GameStartTime    time.Time           //游戏开始计时时间
	GameNowTime      time.Time           //当局游戏开始时间
	nextInviteTime   time.Time           //下次邀请机器人时间
	inviteInterval   int64               //邀请间隔
	pause            bool
	Gaming           bool
	destroyed        bool
	completed        bool
	Testing          bool //是否为测试场
	graceDestroy     bool //等待销毁
	replayAddId      int32
	KeyGameId        string //游戏类型唯一ID
	KeyGamefreeId    string //游戏场次唯一id
	GroupId          int32  //分组id
	bEnterAfterStart bool   //是否允许中途加入
	ClubId           int32
	RoomId           string            //俱乐部那个包间
	RoomPos          int32             //房间桌号
	PumpCoin         int32             //抽水比例，同一个俱乐部下面的抽水比例是一定的,百分比
	DealyTime        int64             //结算延时时间
	CpCtx            model.CoinPoolCtx //水池环境
	CpControlled     bool              //被水池控制了
	timerRandomRobot int64
	nogDismiss       int //检查机器人离场时的局数(同一局只检查一次)
	//playerStatement    map[int32]*webapi.PlayerStatement //玩家流水记录
	SystemCoinOut int64              //本局游戏机器人营收 机器人赢：正值	机器人输：负值
	matchChgData  *SceneMatchChgData //比赛变化数据

	LoopNum           int     // 循环计数
	results           []int   // 本局游戏结果
	WebUser           string  // 操作人
	resultHistory     [][]int // 记录数 [控制结果,局数...]
	BaseScore         int32   //tienlen游戏底分
	MatchId           int32   //标记本次比赛的id，并不是后台id
	MatchFinals       bool    //比赛场决赛
	MatchRound        int32
	MatchCurPlayerNum int32
	MatchNextNeed     int32
	MatchType         int32 //锦标赛、冠军赛
	MatchStop         bool
}

func NewScene(ws *netlib.Session, sceneId, gameMode, sceneMode, gameId int, platform string, params []int32,
	agentor, creator int32, replayCode string, hallId, groupId, totalOfGames int32, dbGameFree *server.DB_GameFree, bEnterAfterStart bool, baseScore int32, playerNum int, paramsEx ...int32) *Scene {
	sp := GetScenePolicy(gameId, gameMode)
	if sp == nil {
		logger.Logger.Errorf("Game id %v not register in ScenePolicyPool.", gameId)
		return nil
	}
	tNow := time.Now()
	s := &Scene{
		ws:               ws,
		SceneId:          sceneId,
		GameId:           gameId,
		GameMode:         gameMode,
		SceneMode:        sceneMode,
		SceneType:        int(dbGameFree.GetSceneType()),
		Params:           params,
		paramsEx:         paramsEx,
		Creator:          creator,
		agentor:          agentor,
		replayCode:       replayCode,
		Players:          make(map[int32]*Player),
		audiences:        make(map[int32]*Player),
		sp:               sp,
		hDisband:         timer.TimerHandle(0),
		GameStartTime:    tNow,
		hallId:           hallId,
		Platform:         platform,
		DbGameFree:       dbGameFree,
		inviteInterval:   model.GameParamData.RobotInviteInitInterval,
		GroupId:          groupId,
		bEnterAfterStart: bEnterAfterStart,
		TotalOfGames:     int(totalOfGames),
		results:          make([]int, common.MaxLoopNum),
		BaseScore:        baseScore,
		playerNum:        playerNum,
	}
	if s != nil && s.init() {
		logger.Logger.Trace("NewScene init success.")
		if !s.Testing {
			s.rrVer = ReplayRecorderVer[gameId]
			s.RecordReplayStart()
		}
		return s
	} else {
		logger.Logger.Trace("NewScene init failed.")
		return nil
	}
}

func (this *Scene) BindAIMgr(aimgr AIMgr) {
	this.aiMgr = aimgr
}

// 根据gamedifid，转为gameid,然后返回所有的相同gameid的数据
func (this *Scene) GetTotalTodayDaliyGameData(keyGameId string, pd *Player) *model.PlayerGameStatics {
	todayData := &model.PlayerGameStatics{}

	if pd.TodayGameData == nil {
		return todayData
	}

	if pd.TodayGameData.CtrlData == nil {
		return todayData
	}

	if info, ok := pd.TodayGameData.CtrlData[keyGameId]; ok {
		todayData.TotalIn += info.TotalIn
		todayData.TotalOut += info.TotalOut
	}

	return todayData
}

func (this *Scene) RebindPlayerSnId(oldSnId, newSnId int32) {
	if p, exist := this.Players[oldSnId]; exist {
		delete(this.Players, oldSnId)
		this.Players[newSnId] = p
	}
	if p, exist := this.audiences[oldSnId]; exist {
		delete(this.audiences, oldSnId)
		this.audiences[newSnId] = p
	}
	if rebind, ok := this.ExtraData.(CanRebindSnId); ok {
		rebind.RebindPlayerSnId(oldSnId, newSnId)
	}
}
func (this *Scene) GetInit() bool {
	return this.init()
}
func (this *Scene) init() bool {
	tNow := time.Now()
	sceneRandSeed++
	this.Rand = rand.New(rand.NewSource(sceneRandSeed))
	this.nextInviteTime = tNow.Add(time.Second * time.Duration(this.Rand.Int63n(model.GameParamData.RobotInviteInitInterval)))
	this.RandRobotCnt()
	this.state = SCENE_STATE_INITED

	if len(this.paramsEx) != 0 {
		if this.IsMatchScene() {
			//this.mp = GetMatchPolicy(this.gameId)
			baseScore := this.GetParamEx(common.PARAMEX_MATCH_BASESCORE)
			this.DbGameFree.BaseScore = proto.Int32(baseScore)
		} else {
			if this.DbGameFree.GetSceneType() == -1 {
				this.Testing = true
			} else {
				this.Testing = false
			}
		}

		//this.keyGameId = strconv.Itoa(int(this.dbGameFree.GetGameId()))
		this.KeyGameId = this.DbGameFree.GetGameDif()
		this.KeyGamefreeId = strconv.Itoa(int(this.DbGameFree.GetId()))
	}
	// test
	//for i := 0; i < 100; i++ {
	//	n := this.rand.Intn(10)
	//	r := this.rand.Intn(3) + 1
	//	str := fmt.Sprint(this.rand.Intn(1000), ":", r)
	//	for j := 0; j < n; j++ {
	//		str += fmt.Sprint(",", this.rand.Intn(1000), ":", r)
	//	}
	//	logger.Logger.Trace("--> str ", str)
	//	this.ParserResults1(str, "")
	//}
	// test
	return true
}

func (this *Scene) GetParam(idx int) int32 {
	if idx < 0 || idx >= len(this.Params) {
		return -1
	}

	return this.Params[idx]
}

func (this *Scene) GetBetMap() []int32 {
	return this.DbGameFree.GetOtherIntParams()
}

func (this *Scene) IsDisbanding() bool {
	return this.hDisband != timer.TimerHandle(0)
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

func (this *Scene) GetMatchTotalOfGame() int {
	return int(this.GetParamEx(common.PARAMEX_MATCH_NUMOFGAME))
}

func (this *Scene) GetMatchBaseScore() int32 {
	return this.GetParamEx(common.PARAMEX_MATCH_BASESCORE)
}

func (this *Scene) GetGameFreeId() int32 {
	return this.DbGameFree.Id
}
func (this *Scene) GetDBGameFree() *server.DB_GameFree {
	return this.DbGameFree
}
func (this *Scene) GetPlatform() string {
	return this.Platform
}
func (this *Scene) GetKeyGameId() string {
	return this.KeyGameId
}
func (this *Scene) SetKeyGameId(keyGameId string) {
	this.KeyGameId = keyGameId
}
func (this *Scene) GetSceneId() int {
	return this.SceneId
}
func (this *Scene) SetSceneId(sceneId int) {
	this.SceneId = sceneId
}
func (this *Scene) GetGroupId() int32 {
	return this.GroupId
}
func (this *Scene) SetGroupId(groupId int32) {
	this.GroupId = groupId
}
func (this *Scene) GetExtraData() interface{} {
	return this.ExtraData
}
func (this *Scene) SetExtraData(data interface{}) {
	this.ExtraData = data
}
func (this *Scene) GetSceneState() SceneState {
	return this.SceneState
}
func (this *Scene) SetSceneState(state SceneState) {
	this.SceneState = state
}
func (this *Scene) GetGameId() int {
	return this.GameId
}
func (this *Scene) SetGameId(gameId int) {
	this.GameId = gameId
}
func (this *Scene) GetPlayerNum() int {
	return this.playerNum
}
func (this *Scene) SetPlayerNum(playerNum int) {
	this.playerNum = playerNum
}
func (this *Scene) GetGameMode() int {
	return this.GameMode
}
func (this *Scene) SetGameMode(gameMode int) {
	this.GameMode = gameMode
}
func (this *Scene) GetGaming() bool {
	return this.Gaming
}
func (this *Scene) SetGaming(gaming bool) {
	this.Gaming = gaming
}
func (this *Scene) GetTesting() bool {
	return this.Testing
}
func (this *Scene) SetTesting(testing bool) {
	this.Testing = testing
}
func (this *Scene) GetCreator() int32 {
	return this.Creator
}
func (this *Scene) SetCreator(creator int32) {
	this.Creator = creator
}
func (this *Scene) GetSceneMode() int {
	return this.SceneMode
}
func (this *Scene) SetSceneMode(sceneMode int) {
	this.SceneMode = sceneMode
}
func (this *Scene) GetParams() []int32 {
	return this.Params
}
func (this *Scene) SetParams(params []int32) {
	this.Params = params
}
func (this *Scene) GetParamsEx() []int32 {
	return this.paramsEx
}
func (this *Scene) SetParamsEx(paramsEx []int32) {
	this.paramsEx = paramsEx
}
func (this *Scene) GetStateStartTime() time.Time {
	return this.StateStartTime
}
func (this *Scene) SetStateStartTime(stateStartTime time.Time) {
	this.StateStartTime = stateStartTime
}
func (this *Scene) GetGameStartTime() time.Time {
	return this.GameStartTime
}
func (this *Scene) SetGameStartTime(gameStartTime time.Time) {
	this.GameStartTime = gameStartTime
}
func (this *Scene) GetGameNowTime() time.Time {
	return this.GameNowTime
}
func (this *Scene) SetGameNowTime(gameNowTime time.Time) {
	this.GameNowTime = gameNowTime
}
func (this *Scene) GetNumOfGames() int {
	return this.NumOfGames
}
func (this *Scene) SetNumOfGames(numOfGames int) {
	this.NumOfGames = numOfGames
}
func (this *Scene) GetCpCtx() model.CoinPoolCtx {
	return this.CpCtx
}
func (this *Scene) SetCpCtx(cpCtx model.CoinPoolCtx) {
	this.CpCtx = cpCtx
}
func (this *Scene) GetAudiences() map[int32]*Player {
	return this.audiences
}
func (this *Scene) GetAgentor() int32 {
	return this.agentor
}
func (this *Scene) SetAgentor(agentor int32) {
	this.agentor = agentor
}
func (this *Scene) GetDisbandGen() int {
	return this.disbandGen
}
func (this *Scene) SetDisbandGen(disbandGen int) {
	this.disbandGen = disbandGen
}
func (this *Scene) GetScenePolicy() ScenePolicy {
	return this.sp
}
func (this *Scene) SetScenePolicy(sp ScenePolicy) {
	this.sp = sp
}
func (this *Scene) GetGraceDestroy() bool {
	return this.graceDestroy
}
func (this *Scene) SetGraceDestroy(graceDestroy bool) {
	this.graceDestroy = graceDestroy
}
func (this *Scene) GetMatchChgData() *SceneMatchChgData {
	return this.matchChgData
}
func (this *Scene) SetMatchChgData(matchChgData *SceneMatchChgData) {
	this.matchChgData = matchChgData
}

func (this *Scene) GetCpControlled() bool {
	return this.CpControlled
}
func (this *Scene) SetCpControlled(cpControlled bool) {
	this.CpControlled = cpControlled
}

func (this *Scene) GetSystemCoinOut() int64 {
	return this.SystemCoinOut
}
func (this *Scene) SetSystemCoinOut(systemCoinOut int64) {
	this.SystemCoinOut = systemCoinOut
}

func (this *Scene) GetBEnterAfterStart() bool {
	return this.bEnterAfterStart
}
func (this *Scene) SetBEnterAfterStart(bEnterAfterStart bool) {
	this.bEnterAfterStart = bEnterAfterStart
}

func (this *Scene) GetTimerRandomRobot() int64 {
	return this.timerRandomRobot
}
func (this *Scene) SetTimerRandomRobot(timerRandomRobot int64) {
	this.timerRandomRobot = timerRandomRobot
}

func (this *Scene) GetDestroyed() bool {
	return this.destroyed
}
func (this *Scene) SetDestroyed(destroyed bool) {
	this.destroyed = destroyed
}

func (this *Scene) GetState() int {
	return this.state
}
func (this *Scene) SetState(state int) {
	this.state = state
}

// ////////////////////////////////////////////////
func (this *Scene) OnStart() {
	logger.Logger.Trace("Scene on start.")
	this.sp.OnStart(this)
	this.TryInviteRobot()
}

func (this *Scene) OnStop() {
	logger.Logger.Trace("Scene on stop.")
	this.sp.OnStop(this)
}

func (this *Scene) OnTick() {
	if !this.pause {
		this.TryInviteRobot()
		this.sp.OnTick(this)
	}
}
func (this *Scene) SendRoomType(p *Player) {
	//通知客户端 当前房间类型
	//RoomSign := &protocol.SCClubRoomSign{
	//	ClubId: proto.Int64(this.ClubId),
	//}
	//proto.SetDefaults(RoomSign)
	//logger.Logger.Trace("RoomSign: ", RoomSign)
	//p.SendToClient(int(protocol.MmoPacketID_PACKET_SC_CLUB_ROOMSIGN), RoomSign)
}

//////////////////////////////////////////////////

func (this *Scene) PlayerEnter(p *Player, isLoaded bool) {
	logger.Logger.Trace("(this *Scene) PlayerEnter:", isLoaded, this.SceneId, p.GetName())
	this.Players[p.SnId] = p
	p.scene = this

	pack := &gamehall.SCEnterRoom{
		GameId:    proto.Int(this.GameId),
		ModeType:  proto.Int(this.GameMode),
		RoomId:    proto.Int(this.SceneId),
		OpRetCode: gamehall.OpResultCode_Game_OPRC_Sucess_Game,
		Params:    []int32{},
		ClubId:    proto.Int32(this.ClubId),
	}
	proto.SetDefaults(pack)
	p.SendToClient(int(gamehall.GameHallPacketID_PACKET_SC_ENTERROOM), pack)

	if p.IsRob {
		this.robotNum++
		logger.Logger.Tracef("(this *Scene) PlayerEnter(%v) robot(%v) robotlimit(%v)", this.DbGameFree.GetName()+this.DbGameFree.GetTitle(), this.robotNum, this.robotLimit)
	} else {
		p.Trusteeship = 0
		p.ValidCacheBetTotal = 0
		this.realPlayerNum++
		this.RandRobotCnt()
	}

	p.OnEnter(this)
	p.SyncFlagToWorld()
	if !isLoaded && !p.IsRob { //等待玩家加载
		p.MarkFlag(PlayerState_Leave)
	}
	//避免游戏接口异常
	utils.RunPanicless(func() { this.sp.OnPlayerEnter(this, p) })

	if p.BlackLevel > 0 {
		WarningBlackPlayer(p.SnId, this.DbGameFree.Id)
	}

	this.ResetNextInviteTime()
}

func (this *Scene) PlayerLeave(p *Player, reason int, isBill bool) {
	logger.Logger.Trace("===(this *Scene) PlayerLeave ", p.SnId, reason, isBill)
	//logger.Logger.Trace(utils.GetCallStack())
	if _, exist := this.Players[p.SnId]; !exist {
		logger.Logger.Warnf("(this *Scene) PlayerLeave(%v) no found in scene(%v)", p.SnId, this.SceneId)
		return
	}

	//当前状态不能离场
	if !this.CanChangeCoinScene(p) && !this.destroyed {
		pack := &gamehall.SCLeaveRoom{
			//OpRetCode: p.opCode, //protocol.OpResultCode_OPRC_Hundred_YouHadBetCannotLeave,
			OpRetCode: gamehall.OpResultCode_Game(p.OpCode),
			RoomId:    proto.Int(this.SceneId),
		}
		if pack.GetOpRetCode() == gamehall.OpResultCode_Game_OPRC_Sucess_Game {
			//不能这么做，机器人有特殊判定
			//pack.OpRetCode = gamehall.OpResultCode_OPRC_Error
			pack.OpRetCode = gamehall.OpResultCode_Game_OPRC_Error_Game
		}
		proto.SetDefaults(pack)
		p.SendToClient(int(gamehall.GameHallPacketID_PACKET_SC_LEAVEROOM), pack)
		logger.Logger.Tracef("(this *Scene) Cant PlayerLeave(%v) no found in scene(%v)", p.SnId, this.SceneId)
		return
	}

	//避免游戏接口异常
	utils.RunPanicless(func() { this.sp.OnPlayerLeave(this, p, reason) })

	p.OnLeave(reason)
	delete(this.Players, p.SnId)
	isBill = true

	//send world离开房间
	pack := &server.GWPlayerLeave{
		RoomId:               proto.Int(this.SceneId),
		PlayerId:             proto.Int32(p.SnId),
		Reason:               proto.Int(reason),
		ServiceFee:           proto.Int64(p.serviceFee),
		GameTimes:            proto.Int32(p.GameTimes),
		BetCoin:              proto.Int64(p.TotalBet),
		WinTimes:             proto.Int(p.winTimes),
		LostTimes:            proto.Int(p.lostTimes),
		TotalConvertibleFlow: proto.Int64(p.TotalConvertibleFlow),
		ValidCacheBetTotal:   proto.Int64(p.ValidCacheBetTotal),
		MatchId:              proto.Int32(this.MatchId),
		CurIsWin:             proto.Int64(p.CurIsWin), // 负数：输  0：平局  正数：赢
		MatchStop:            proto.Bool(this.MatchStop),
	}

	pack.ReturnCoin = proto.Int64(p.Coin)
	if this.Testing {
		pack.ReturnCoin = proto.Int64(p.takeCoin)
	}
	pack.GameCoinTs = proto.Int64(p.GameCoinTs)
	if !p.IsLocal {
		data, err := p.MarshalData(this.GameId)
		if err == nil {
			pack.PlayerData = data
		}
	}
	items := p.Items
	if items != nil {
		pack.Items = make(map[int32]int32)
		for id, num := range items {
			pack.Items[id] = proto.Int32(num)
		}
	}
	proto.SetDefaults(pack)
	this.SendToWorld(int(server.SSPacketID_PACKET_GW_PLAYERLEAVE), pack)

	if p.IsRob {
		this.robotNum--
		logger.Logger.Tracef("(this *Scene) PlayerLeave(%v) robot(%v) robotlimit(%v)", this.DbGameFree.GetName()+this.DbGameFree.GetTitle(), this.robotNum, this.robotLimit)
	} else {
		this.realPlayerNum--
		this.RandRobotCnt()
	}

	this.ResetNextInviteTime()
}

func (this *Scene) AudienceEnter(p *Player, isload bool) {
	logger.Logger.Trace("(this *Scene) AudienceEnter")
	this.audiences[p.SnId] = p
	p.scene = this
	p.MarkFlag(PlayerState_Audience)
	pack := &gamehall.SCEnterRoom{
		GameId:    proto.Int(this.GameId),
		ModeType:  proto.Int(this.GameMode),
		RoomId:    proto.Int(this.SceneId),
		OpRetCode: gamehall.OpResultCode_Game_OPRC_Sucess_Game,
	}
	proto.SetDefaults(pack)
	p.SendToClient(int(gamehall.GameHallPacketID_PACKET_SC_ENTERROOM), pack)

	p.OnAudienceEnter(this)
	if !isload && !p.IsRob {
		p.MarkFlag(PlayerState_Leave)
	}
	//避免游戏接口异常
	utils.RunPanicless(func() { this.sp.OnAudienceEnter(this, p) })
}

func (this *Scene) AudienceLeave(p *Player, reason int) {
	logger.Logger.Trace("(this *Scene) AudienceLeave")
	//当前状态不能离场
	if !this.CanChangeCoinScene(p) {
		pack := &gamehall.SCLeaveRoom{
			OpRetCode: (gamehall.OpResultCode_Game(p.OpCode)), //protocol.OpResultCode_OPRC_Hundred_YouHadBetCannotLeave,
			RoomId:    proto.Int(this.SceneId),
		}
		proto.SetDefaults(pack)
		p.SendToClient(int(gamehall.GameHallPacketID_PACKET_SC_LEAVEROOM), pack)
		return
	}
	//避免游戏接口异常
	utils.RunPanicless(func() { this.sp.OnAudienceLeave(this, p, reason) })
	p.OnAudienceLeave(reason)
	delete(this.audiences, p.SnId)
	//send world离开房间
	pack := &server.GWPlayerLeave{
		RoomId:               proto.Int(this.SceneId),
		PlayerId:             proto.Int32(p.SnId),
		Reason:               proto.Int(reason),
		TotalConvertibleFlow: proto.Int64(p.TotalConvertibleFlow),
		ValidCacheBetTotal:   proto.Int64(p.ValidCacheBetTotal),
	}
	pack.ReturnCoin = proto.Int64(p.Coin)
	if this.Testing {
		pack.ReturnCoin = proto.Int64(p.takeCoin)
	}
	pack.GameCoinTs = proto.Int64(p.GameCoinTs)
	//	if p.dirty {
	//		data, err := p.MarshalData(this.gameId)
	//		if err == nil {
	//			pack.PlayerData = data
	//		}
	//	}
	proto.SetDefaults(pack)
	this.SendToWorld(int(server.SSPacketID_PACKET_GW_AUDIENCELEAVE), pack)
}

func (this *Scene) AudienceSit(p *Player) {
	logger.Logger.Trace("(this *Scene) AudienceSit")
	if _, exist := this.audiences[p.SnId]; exist {
		delete(this.audiences, p.SnId)

		this.Players[p.SnId] = p
		p.scene = this

		p.OnEnter(this)
		//避免游戏接口异常
		utils.RunPanicless(func() { this.sp.OnAudienceSit(this, p) })
	}
}

func (this *Scene) HasPlayer(p *Player) bool {
	if p == nil {
		return false
	}

	if pp, ok := this.Players[p.SnId]; ok && pp == p {
		return true
	}
	return false
}

func (this *Scene) HasAudience(p *Player) bool {
	if p == nil {
		return false
	}

	if pp, ok := this.audiences[p.SnId]; ok && pp == p {
		return true
	}
	return false
}

func (this *Scene) GetPlayer(id int32) *Player {
	if p, exist := this.Players[id]; exist {
		return p
	}
	return nil
}

func (this *Scene) GetPlayerByPos(pos int) *Player {
	for _, p := range this.Players {
		if p.Pos == pos {
			return p
		}
	}
	return nil
}

func (this *Scene) PlayerDropLine(snid int32) {
	logger.Logger.Trace("(this *Scene) PlayerDropLine")
	if p, exist := this.Players[snid]; exist {
		p.OnDropLine()
		//避免游戏接口异常
		utils.RunPanicless(func() { this.sp.OnPlayerDropLine(this, p) })
	} else if p, exist := this.audiences[snid]; exist {
		p.OnAudienceDropLine()
		//避免游戏接口异常
		utils.RunPanicless(func() { this.sp.OnAudienceDropLine(this, p) })
	}
}

func (this *Scene) PlayerRehold(snid int32, sid int64, gs *netlib.Session) {
	logger.Logger.Trace("(this *Scene) PlayerRehold")
	if p, exist := this.Players[snid]; exist {
		p.OnRehold(sid, gs)
		//if !p.IsRob {
		//	p.trusteeship = 0
		//}
		//避免游戏接口异常
		utils.RunPanicless(func() { this.sp.OnPlayerRehold(this, p) })
	} else if p, exist := this.audiences[snid]; exist {
		p.OnRehold(sid, gs)
		//if !p.IsRob {
		//	p.trusteeship = 0
		//}
		//避免游戏接口异常
		utils.RunPanicless(func() { this.sp.OnAudienceEnter(this, p) })
	}
}

func (this *Scene) PlayerReturn(p *Player, isLoaded bool) {
	logger.Logger.Trace("(this *Scene) PlayerReturn")
	pack := &gamehall.SCReturnRoom{
		RoomId:    proto.Int(this.SceneId),
		GameId:    proto.Int(this.GameId),
		ModeType:  proto.Int(this.GameMode),
		Params:    this.Params,
		HallId:    proto.Int32(this.hallId),
		IsLoaded:  proto.Bool(isLoaded),
		OpRetCode: gamehall.OpResultCode_Game_OPRC_Sucess_Game,
		ClubId:    proto.Int32(this.ClubId),
	}
	proto.SetDefaults(pack)
	p.SendToClient(int(gamehall.GameHallPacketID_PACKET_SC_RETURNROOM), pack)
	logger.Logger.Tracef("Scene.PlayerReturn %v", pack)
	//if !p.IsRob {
	//	p.trusteeship = 0
	//}
	if this.HasPlayer(p) {
		//避免游戏接口异常
		//utils.RunPanicless(func() { this.sp.OnPlayerRehold(this, p) })
		//这里应该调用 return消息 因为在上面rehold的消息已经处理过了
		utils.RunPanicless(func() { this.sp.OnPlayerReturn(this, p) })
	} else if this.HasAudience(p) {
		//避免游戏接口异常
		utils.RunPanicless(func() { this.sp.OnAudienceEnter(this, p) })
	}
	if !p.IsRob { //等待玩家加载
		if isLoaded {
			p.UnmarkFlag(PlayerState_Leave)
		} else {
			p.MarkFlag(PlayerState_Leave)
		}
	}
	if this.IsMatchScene() {
		p.SetIParam(common.PlayerIParam_IsQuit, 0)
		p.UnmarkFlag(PlayerState_MatchQuit)
	}
}

func (this *Scene) Broadcast(packetid int, msg rawproto.Message, excludeSid int64, includeOffline ...bool) {
	excludePos := -1
	mgs := make(map[*netlib.Session][]*srvlibproto.MCSessionUnion)
	for _, p := range this.Players {
		if p != nil {
			if p.sid != excludeSid {
				if (p.gateSess != nil && p.IsOnLine() && !p.IsMarkFlag(PlayerState_Leave)) || len(includeOffline) != 0 {
					mgs[p.gateSess] = append(mgs[p.gateSess], &srvlibproto.MCSessionUnion{
						Mccs: &srvlibproto.MCClientSession{
							SId: proto.Int64(p.sid),
						},
					})
				}
			}
			if p.sid == excludeSid {
				excludePos = p.Pos
			}
		}
	}
	for _, p := range this.audiences {
		if p != nil && p.sid != excludeSid {
			if (p.gateSess != nil && p.IsOnLine() && !p.IsMarkFlag(PlayerState_Leave)) || len(includeOffline) != 0 {
				mgs[p.gateSess] = append(mgs[p.gateSess], &srvlibproto.MCSessionUnion{
					Mccs: &srvlibproto.MCClientSession{
						SId: proto.Int64(p.sid),
					},
				})
			}
		}
	}
	if this.rr != nil && !this.Testing && this.Gaming && !this.IsHundredScene() && !this.IsMatchScene() {
		this.rr.Record(-1, excludePos, packetid, msg)
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
func (this *Scene) RobotBroadcast(packetid int, msg rawproto.Message) {
	mgs := make(map[*netlib.Session][]*srvlibproto.MCSessionUnion)
	for _, p := range this.Players {
		if p != nil && p.IsRob {
			if p.gateSess != nil && p.IsOnLine() && !p.IsMarkFlag(PlayerState_Leave) {
				mgs[p.gateSess] = append(mgs[p.gateSess], &srvlibproto.MCSessionUnion{
					Mccs: &srvlibproto.MCClientSession{
						SId: proto.Int64(p.sid),
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
func (this *Scene) BroadcastToAudience(packetid int, msg rawproto.Message) {
	if len(this.audiences) > 0 {
		mgs := make(map[*netlib.Session][]*srvlibproto.MCSessionUnion)
		for _, p := range this.audiences {
			if p != nil {
				if p.gateSess != nil && p.IsOnLine() {
					mgs[p.gateSess] = append(mgs[p.gateSess], &srvlibproto.MCSessionUnion{
						Mccs: &srvlibproto.MCClientSession{
							SId: proto.Int64(p.sid),
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
}

func (this *Scene) GetAudiencesNum() int {
	if this.audiences != nil {
		return len(this.audiences)
	}
	return 0
}

func (this *Scene) ChangeSceneState(stateid int) {
	if this.destroyed {
		return
	}
	state := this.sp.GetSceneState(this, stateid)
	if state == nil {
		return
	}
	oldState := -1
	if this.SceneState != nil {
		oldState = this.SceneState.GetState()
		if this.SceneState.CanChangeTo(state) {
			logger.Logger.Tracef("(this *Scene) [%v] ChangeSceneState %v -> %v", this.SceneId, this.SceneState.GetState(), state.GetState())

			this.SceneState.OnLeave(this)
			this.SceneState = state
			this.SceneState.OnEnter(this)
			this.sp.NotifyGameState(this)
		} else {
			logger.Logger.Tracef("(this *Scene) [%v] ChangeSceneState failed %v -> %v", this.SceneId, this.SceneState.GetState(), state.GetState())
		}
	} else {
		logger.Logger.Tracef("(this *Scene) [%v] ChangeSceneState -> %v", this.SceneId, state.GetState())
		this.SceneState = state
		this.SceneState.OnEnter(this)
		//this.NotifySceneState(stateid)
	}

	if this.aiMgr != nil {
		this.aiMgr.OnChangeState(this, oldState, stateid)
	}
}

func (this *Scene) SendToWorld(packetid int, pack interface{}) {
	if this.ws != nil {
		this.ws.Send(packetid, pack)
	}
}

func (this *Scene) FirePlayerEvent(p *Player, evtcode int, params []int64) {
	if this.SceneState != nil {
		this.SceneState.OnPlayerEvent(this, p, evtcode, params)
	}
	////比赛事件
	//if this.mp != nil {
	//	this.mp.OnPlayerEvent(this, p, evtcode, params)
	//}
}

func (this *Scene) Pause() {
	this.pause = true
}

func (this *Scene) Destroy(force bool) {

	this.destroyed = true
	this.pause = true

	if !this.IsMatchScene() {
		for _, p := range this.Players {
			this.PlayerLeave(p, common.PlayerLeaveReason_OnDestroy, true)
		}
		for _, p := range this.audiences {
			//this.PlayerLeave(p, common.PlayerLeaveReason_OnDestroy, true)
			this.AudienceLeave(p, common.PlayerLeaveReason_OnDestroy)
		}
	} else {
		for _, p := range this.Players {
			this.PlayerLeave(p, common.PlayerLeaveReason_OnBilled, true)
		}
	}

	for _, p := range this.Players {
		PlayerMgrSington.DelPlayerBySnId(p.SnId)
	}

	for _, p := range this.audiences {
		PlayerMgrSington.DelPlayerBySnId(p.SnId)
	}

	isCompleted := this.sp.IsCompleted(this) || this.completed
	SceneMgrSington.DestroyScene(this.SceneId)
	pack := &server.GWDestroyScene{
		SceneId:     proto.Int(this.SceneId),
		IsCompleted: proto.Bool(isCompleted),
	}
	proto.SetDefaults(pack)
	this.SendToWorld(int(server.SSPacketID_PACKET_GW_DESTROYSCENE), pack)
	logger.Logger.Trace("(this *Scene) Destroy(force bool) isCompleted", isCompleted)
}

// 是否公平竞争
func (this *Scene) IsFair() bool {
	//不公平竞争,关照过度亏损的玩家
	return false
}

func (this *Scene) IsPrivateScene() bool {
	return this.SceneId >= common.PrivateSceneStartId && this.SceneId <= common.PrivateSceneMaxId || this.SceneMode == common.SceneMode_Private
}

func (this *Scene) IsMatchScene() bool {
	return this.SceneId >= common.MatchSceneStartId && this.SceneId <= common.MatchSceneMaxId
}

func (this *Scene) IsFull() bool {
	return len(this.Players) >= this.playerNum
}

// 大厅场
func (this *Scene) IsHallScene() bool {
	return this.SceneId >= common.HallSceneStartId && this.SceneId <= common.HallSceneMaxId
}

// 金豆自由场
func (this *Scene) IsCoinScene() bool {
	return this.SceneId >= common.CoinSceneStartId && this.SceneId <= common.CoinSceneMaxId
}

// 百人场
func (this *Scene) IsHundredScene() bool {
	return this.SceneId >= common.HundredSceneStartId && this.SceneId <= common.HundredSceneMaxId
}

func (this *Scene) GetCoinSceneLowerThanKick() int32 {
	if this.DbGameFree != nil {
		return this.DbGameFree.GetLowerThanKick()
	}
	return 0
}

func (this *Scene) GetCoinSceneMaxCoinLimit() int32 {
	if this.DbGameFree != nil {
		return this.DbGameFree.GetMaxCoinLimit()
	}
	return 0
}

func (this *Scene) CoinInLimit(coin int64) bool {
	if common.IsLocalGame(this.GameId) {
		min := this.GetLimitCoin()
		if min != 0 && coin < min {
			return false
		}
		if coin <= 0 {
			return false
		}
		return true
	}
	min := int(this.GetCoinSceneLowerThanKick())
	max := int(this.GetCoinSceneMaxCoinLimit())
	if min != 0 && coin < int64(min) {
		return false
	}
	if max != 0 && coin > int64(max) {
		return false
	}
	if coin <= 0 {
		return false
	}
	return true
}

// 根据底注去取createroom表里面的最小携带金额
func (this *Scene) GetLimitCoin() int64 {
	limitCoin := int64(0)
	tmpIds := []int32{}
	for _, data := range srvdata.PBDB_CreateroomMgr.Datas.GetArr() {
		if int(data.GameId) == this.GameId && int(data.GameSite) == this.SceneType {
			betRange := data.GetBetRange()
			if len(betRange) == 0 {
				continue
			}
			for j := 0; j < len(betRange); j++ {
				if betRange[j] == this.BaseScore && len(data.GetGoldRange()) > 0 && data.GetGoldRange()[0] != 0 {
					tmpIds = append(tmpIds, data.GetId())
					break
				}
			}
		}
	}
	if len(tmpIds) > 0 {
		goldRange := srvdata.PBDB_CreateroomMgr.GetData(tmpIds[0]).GetGoldRange()
		if len(goldRange) != 0 && goldRange[0] != 0 {
			limitCoin = int64(goldRange[0])
		}
		if limitCoin != 0 {
			for _, id := range tmpIds {
				tmp := srvdata.PBDB_CreateroomMgr.GetData(id).GetGoldRange()
				if int64(tmp[0]) < limitCoin && tmp[0] != 0 {
					limitCoin = int64(tmp[0])
				}
			}
		}
	}
	return limitCoin
}

func (this *Scene) CoinOverMaxLimit(coin int64, p *Player) bool {
	if this.Testing {
		return false
	}
	if coin < 0 {
		return false
	}

	if p.ExpectLeaveCoin != 0 && this.IsCoinScene() { //暂只对对战场生效
		if p.ExpectLeaveCoin < p.takeCoin { //期望输的时候离场
			if coin <= p.ExpectLeaveCoin {
				return true
			}
		} else { //期望赢的时候离场
			if coin >= p.ExpectLeaveCoin {
				return true
			}
		}
	} else {
		if this.DbGameFree != nil {
			limit := this.DbGameFree.GetRobotLimitCoin()
			if len(limit) >= 2 {
				comp := common.RandInt(int(limit[0]), int(limit[1]))
				if coin > int64(comp) {
					return true
				}
			}
		}
	}

	return false
}

func (this *Scene) CorrectBillCoin(coin, limit1, limit2 int64) int64 {
	if coin > limit1 {
		coin = limit1
	}
	if coin > limit2 {
		coin = limit2
	}
	return coin
}

func (this *Scene) GetCoinSceneServiceFee() int32 {
	if this.DbGameFree != nil {
		return this.DbGameFree.GetServiceFee()
	}
	return 0
}

func (this *Scene) GetCoinSceneTypeId() int32 {
	if this.DbGameFree != nil {
		return this.DbGameFree.Id
	}
	return 0
}

func (this *Scene) GetCoinSceneName() string {
	if this.DbGameFree != nil {
		return this.DbGameFree.GetName() + this.DbGameFree.GetTitle()
	}
	return ""
}

func (this *Scene) GetHundredSceneName() string {
	if this.IsHundredScene() && this.DbGameFree != nil {
		if this.DbGameFree.GetName() == this.DbGameFree.GetTitle() {
			return this.DbGameFree.GetTitle()
		} else {
			return this.DbGameFree.GetName() + this.DbGameFree.GetTitle()
		}
	}
	return ""
}

func (this *Scene) GetSceneName() string {
	if this.IsCoinScene() {
		return this.GetCoinSceneName()
	} else if this.IsHundredScene() {
		return this.GetHundredSceneName()
	}
	return ""
}

func (this *Scene) CanChangeCoinScene(p *Player) bool {
	//if p.IsMarkFlag(PlayerState_Audience) {
	//	if this.drp != nil {
	//		return this.drp.CanChangeCoinScene(this, p)
	//	}
	//}
	//if this.mp != nil {
	//	if !this.mp.IsMatchEnd(this) {
	//		return false
	//	}
	//}
	if this.sp != nil {
		return this.sp.CanChangeCoinScene(this, p)
	}
	return false
}

func (this *Scene) SyncPlayerCoin() {
	//if this.Testing {
	//	return
	//}
	//pack := &server.GWSyncPlayerCoin{
	//	SceneId: proto.Int(this.SceneId),
	//}
	//switch this.GameId {
	//case common.GameId_HFishing, common.GameId_TFishing:
	//	for _, value := range this.Players {
	//		if value.IsRob {
	//			continue
	//		}
	//		//todo dev 捕鱼的逻辑暂时不用 开发的时候再增加
	//		//if exData, ok := value.extraData.(*FishingPlayerData); ok {
	//		//	if exData.CoinCache != value.LastSyncCoin {
	//		//		pack.PlayerCoins = append(pack.PlayerCoins, int64(value.SnId))
	//		//		pack.PlayerCoins = append(pack.PlayerCoins, exData.CoinCache)
	//		//		value.LastSyncCoin = exData.CoinCache
	//		//	}
	//		//}
	//	}
	//default:
	//	for _, value := range this.Players {
	//		if value.Coin != value.LastSyncCoin && !value.IsRob {
	//			pack.PlayerCoins = append(pack.PlayerCoins, int64(value.SnId))
	//			pack.PlayerCoins = append(pack.PlayerCoins, value.Coin)
	//			value.LastSyncCoin = value.Coin
	//		}
	//	}
	//}
	//if len(pack.PlayerCoins) > 0 {
	//	proto.SetDefaults(pack)
	//	this.SendToWorld(int(server.SSPacketID_PACKET_GW_SYNCPLAYERCOIN), pack)
	//}
}
func (this *Scene) NotifySceneStateFishing(state int) {
	pack := &server.GWSceneState{
		RoomId:  proto.Int(this.SceneId),
		Fishing: proto.Int32(int32(state)),
	}
	proto.SetDefaults(pack)
	this.SendToWorld(int(server.SSPacketID_PACKET_GW_SCENESTATE), pack)
}
func (this *Scene) NotifySceneRoundStart(round int) {
	pack := &server.GWSceneStart{
		RoomId:    proto.Int(this.SceneId),
		CurrRound: proto.Int(round),
		Start:     proto.Bool(true),
		MaxRound:  proto.Int(this.TotalOfGames),
	}
	proto.SetDefaults(pack)
	this.SendToWorld(int(server.SSPacketID_PACKET_GW_SCENESTART), pack)
}

func (this *Scene) NotifySceneRoundPause() {
	pack := &server.GWSceneStart{
		RoomId:    proto.Int(this.SceneId),
		Start:     proto.Bool(false),
		CurrRound: proto.Int(this.NumOfGames),
		MaxRound:  proto.Int(this.TotalOfGames),
	}
	proto.SetDefaults(pack)
	this.SendToWorld(int(server.SSPacketID_PACKET_GW_SCENESTART), pack)
}
func (this *Scene) SyncGameState(sec, bl int) {
	if this.SceneState != nil {
		pack := &server.GWGameState{
			SceneId:       proto.Int(this.SceneId),
			State:         proto.Int(this.SceneState.GetState()),
			Ts:            proto.Int64(time.Now().Unix()),
			Sec:           proto.Int(sec),
			BankerListNum: proto.Int(bl),
		}
		proto.SetDefaults(pack)
		this.SendToWorld(int(server.SSPacketID_PACKET_GW_GAMESTATE), pack)
	}
}

// 游戏开始的时候同步防伙牌数据
func (this *Scene) SyncScenePlayer() {
	pack := &server.GWScenePlayerLog{
		GameId:     proto.Int(this.GameId),
		GameFreeId: proto.Int32(this.DbGameFree.GetId()),
	}
	for _, value := range this.Players {
		if value.IsRob || !value.IsGameing() {
			continue
		}
		pack.Snids = append(pack.Snids, value.SnId)
		pack.IsGameing = append(pack.IsGameing, value.IsGameing())
	}
	this.SendToWorld(int(server.SSPacketID_PACKET_GW_SCENEPLAYERLOG), pack)
}

// 防伙牌换桌
func (this *Scene) ChangeSceneEvent() {
	timer.StartTimer(timer.TimerActionWrapper(func(h timer.TimerHandle, ud interface{}) bool {
		if this.DbGameFree.GetMatchMode() == 1 {
			return true
		}
		this.SendToWorld(int(server.SSPacketID_PACKET_GW_CHANGESCENEEVENT), &server.GWChangeSceneEvent{
			SceneId: proto.Int(this.SceneId),
		})
		return true
	}), nil, time.Second*3, 1)
}
func (this *Scene) RecordReplayStart() {
	if !this.IsHundredScene() && !this.IsMatchScene() {
		logger.Logger.Trace("RecordReplayStart-----", this.replayCode, this.NumOfGames, this.replayAddId)
		id := fmt.Sprintf("%d%d%v%d", this.GameId, this.SceneId, this.GameNowTime.Format(ReplayIdTf), this.replayAddId)
		this.rr = NewReplayRecorder(id)
	}
}

func (this *Scene) RecordReplayOver() {
	if !this.Testing && !this.IsHundredScene() && !this.IsMatchScene() {
		logger.Logger.Trace("RecordReplayOver-----", this.replayCode, this.NumOfGames, this.replayAddId)
		this.replayAddId++
		this.rr.Fini(this)

		this.RecordReplayStart()
	}
}
func (this *Scene) IsTienLen() bool {
	return this.GameId == common.GameId_TienLen || this.GameId == common.GameId_TienLen_yl ||
		this.GameId == common.GameId_TienLen_toend || this.GameId == common.GameId_TienLen_yl_toend ||
		this.GameId == common.GameId_TienLen_m || this.GameId == common.GameId_TienLen_yl_toend_m
}
func (this *Scene) TryDismissRob(params ...int) {
	if this.IsMatchScene() {
		return
	}
	if this.IsCoinScene() {
		allRobot := true
		for _, p := range this.Players {
			if !p.IsRob {
				allRobot = false
				break
			}
		}
		//一次离开一个
		hasLeave := false
		if allRobot && !this.IsPreCreateScene() {
			for _, p := range this.Players {
				if p.IsRob {
					this.PlayerLeave(p, common.PlayerLeaveReason_Normal, true)
					hasLeave = true
				}
			}
		}

		//当局已经检查过了
		if this.nogDismiss == this.NumOfGames {
			return
		}

		//如果是满桌并且是禁止匹配真人，那么保持满桌几局
		if this.DbGameFree.GetMatchTrueMan() == common.MatchTrueMan_Forbid && this.IsFull() && rand.Intn(4) == 1 {
			hasLeave = true
		}

		if !hasLeave && !this.Testing {
			for _, p := range this.Players {
				rands := this.Rand.Int63n(20) + 20
				a := float64(p.Coin) / float64(p.takeCoin)
				if p != nil && p.IsRob && a >= float64(rands)/10 {
					this.PlayerLeave(p, common.PlayerLeaveReason_Normal, true)
					hasLeave = true
					break
				}
			}
		}

		if !hasLeave && this.DbGameFree.GetMatchTrueMan() != common.MatchTrueMan_Forbid && len(params) > 0 &&
			params[0] == 1 && this.IsFull() && common.RandInt(10000) < 4000 {
			for _, r := range this.Players {
				if r.IsRob {
					this.PlayerLeave(r, common.PlayerLeaveReason_Normal, true)
					hasLeave = true
					break
				}
			}
		}

		if !hasLeave {
			for _, r := range this.Players {
				if r.IsRob {
					if !r.IsGameing() { //5%的概率,不玩游戏直接离场
						if rand.Intn(100) < 5 {
							this.PlayerLeave(r, common.PlayerLeaveReason_Normal, true)
							hasLeave = true
							break
						}
					} else { //玩游戏的,玩几局有概率离场
						expectTimes := (5 + rand.Int31n(20))
						if r.GameTimes >= expectTimes {
							this.PlayerLeave(r, common.PlayerLeaveReason_Normal, true)
							hasLeave = true
							break
						}
					}
				}
			}
		}
		this.nogDismiss = this.NumOfGames

		//如果当局有机器人离开,适当延长下下次邀请的时间
		if hasLeave {
			tNow := time.Now()
			if this.nextInviteTime.Sub(tNow) < time.Second {
				this.nextInviteTime = tNow.Add(time.Second * time.Duration(rand.Int31n(3)+1))
			}
		}
	}
}

func (this *Scene) CreateGameRecPacket() *server.GWGameRec {
	return &server.GWGameRec{
		RoomId:     proto.Int(this.SceneId),
		NumOfGames: proto.Int(this.NumOfGames),
		GameTime:   proto.Int(int(time.Now().Sub(this.GameStartTime) / time.Second)),
	}
}
func (this *Scene) IsAllReady() bool {
	for _, p := range this.Players {
		if !p.IsOnLine() || !p.IsReady() {
			return false
		}
	}
	return true
}

func (this *Scene) GetOnlineCnt() int {
	cnt := 0
	for _, p := range this.Players {
		if p.IsOnLine() && !p.IsMarkFlag(PlayerState_Leave) {
			cnt++
		}
	}
	return cnt
}

func (this *Scene) GetRealPlayerCnt() int {
	cnt := 0
	for _, p := range this.Players {
		if !p.IsRob {
			cnt++
		}
	}
	return cnt
}

func (this *Scene) GetGameingPlayerCnt() int {
	cnt := 0
	for _, p := range this.Players {
		if p != nil && p.IsGameing() {
			cnt += 1
		}
	}
	return cnt
}

func (this *Scene) GetGameingRealPlayerCnt() int {
	cnt := 0
	for _, p := range this.Players {
		if p != nil && p.IsGameing() && !p.IsRob {
			cnt += 1
		}
	}
	return cnt
}

func (this *Scene) GetRandomRobotPlayer() *Player {
	robotArray := []*Player{}
	for _, p := range this.Players {
		if p != nil && p.IsGameing() && p.IsRob {
			robotArray = append(robotArray, p)
		}
	}
	if len(robotArray) > 0 {
		return robotArray[common.RandInt(0, len(robotArray))]
	}

	return nil
}

func (this *Scene) CoinPoolCanOut() bool {
	return true
	/* 暂时屏蔽
	noRobotPlayerCount := this.GetRealPlayerCnt()
	setting := coinPoolMgr.GetCoinPoolSetting(this.platform, this.gamefreeId, this.groupId)
	if setting != nil {
		return int32(noRobotPlayerCount) >= setting.GetMinOutPlayerNum()
	}
	return int32(noRobotPlayerCount) >= this.dbGameFree.GetMinOutPlayerNum()
	*/
}
func (this *Scene) ClearAutoPlayer() {
	for _, p := range this.Players {
		if p.IsAuto() {
			p.UnmarkFlag(PlayerState_Auto)
			p.SyncFlag()
		}
	}
}

func (this *Scene) NewBigCoinNotice(player *Player, num int64, msgType int64) {
	if !this.Testing && !this.IsMatchScene() {
		if num < model.GameParamData.NoticeCoinMin || model.GameParamData.NoticeCoinMax < num {
			return
		}
		start := time.Now().Add(time.Second * 30).Unix()
		content := fmt.Sprintf("%v|%v|%v", player.GetName(), this.GetHundredSceneName(), num)
		pack := &server.GWNewNotice{
			Ch:       proto.String(""),
			Content:  proto.String(content),
			Start:    proto.Int64(start),
			Interval: proto.Int64(0),
			Count:    proto.Int64(1),
			Msgtype:  proto.Int64(msgType),
			Platform: proto.String(player.Platform),
			Isrob:    proto.Bool(player.IsRob),
			Priority: proto.Int32(int32(num)),
		}
		if common.HorseRaceLampPriority_Rand == model.GameParamData.NoticePolicy {
			pack.Priority = proto.Int32(rand.Int31n(100))
		}
		this.SendToWorld(int(server.SSPacketID_PACKET_GW_NEWNOTICE), pack)
	}
}

type GameDetailedParam struct {
	Trend20Lately string //最近20局开奖结果
}

// 保存详细游戏日志
func (this *Scene) SaveGameDetailedLog(logid string, gamedetailednote string, gameDetailedParam *GameDetailedParam) {
	if this != nil {
		if !this.Testing { //测试场屏蔽掉
			trend20Lately := gameDetailedParam.Trend20Lately
			baseScore := this.DbGameFree.GetBaseScore()
			if common.IsLocalGame(this.GameId) {
				baseScore = this.BaseScore
			}
			log := model.NewGameDetailedLogEx(logid, int32(this.GameId), int32(this.SceneId),
				this.DbGameFree.GetGameMode(), this.DbGameFree.Id, int32(len(this.Players)),
				int32(time.Now().Unix()-this.GameNowTime.Unix()), baseScore,
				gamedetailednote, this.Platform, this.ClubId, this.RoomId, this.CpCtx, GameDetailedVer[this.GameId], trend20Lately)
			if log != nil {
				if this.IsMatchScene() {
					log.MatchId = this.MatchId //this.GetParamEx(common.PARAMEX_MATCH_ID)
				}
				LogChannelSington.WriteLog(log)
			}
		}
	}
}

type SaveGamePlayerListLogParam struct {
	Platform          string //平台
	Channel           string //渠道
	Promoter          string //推广员
	PackageTag        string //包标识
	InviterId         int32  //邀请人
	LogId             string //日志id
	TotalIn           int64  //总投入
	TotalOut          int64  //总产出
	TaxCoin           int64  //总税收
	ClubPumpCoin      int64  //俱乐部抽水
	BetAmount         int64  //下注量
	WinAmountNoAnyTax int64  //税后赢取额(净利润)
	ValidBet          int64  //有效下注
	ValidFlow         int64  //有效流水
	IsFirstGame       bool   //是否第一次游戏
	IsLeave           bool   //是否中途离开，用于金花，德州可以中途离开游戏使用
}

func GetSaveGamePlayerListLogParam(platform, channel, promoter, packageTag, logid string,
	inviterId int32, totalin, totalout, taxCoin, clubPumpCoin, betAmount, winAmountNoAnyTax, validBet, validFlow int64,
	isFirstGame, isLeave bool) *SaveGamePlayerListLogParam {
	return &SaveGamePlayerListLogParam{
		Platform:          platform,
		Channel:           channel,
		Promoter:          promoter,
		PackageTag:        packageTag,
		InviterId:         inviterId,
		LogId:             logid,
		TotalIn:           totalin,
		TotalOut:          totalout,
		TaxCoin:           taxCoin,
		ClubPumpCoin:      clubPumpCoin,
		BetAmount:         betAmount,
		WinAmountNoAnyTax: winAmountNoAnyTax,
		ValidBet:          validBet,
		ValidFlow:         validFlow,
		IsFirstGame:       isFirstGame,
		IsLeave:           isLeave,
	}
}

func IsFishGame(gameId int) bool {
	if gameId == common.GameId_HFishing || gameId == common.GameId_TFishing {
		return true
	}

	return false
}
func (this *Scene) SaveFriendRecord(snid int32, isWin int32) {
	if this.SceneMode == common.SceneMode_Private {
		return
	}
	var baseScore = this.DbGameFree.GetBaseScore()
	if common.IsLocalGame(this.GameId) {
		baseScore = this.BaseScore
	}
	log := model.NewFriendRecordLogEx(this.Platform, snid, isWin, int32(this.GameId), baseScore)
	if log != nil {
		LogChannelSington.WriteLog(log)
	}
}

// 保存玩家和GameDetailedLog的映射表
func (this *Scene) SaveGamePlayerListLog(snid int32, param *SaveGamePlayerListLogParam) {
	if this != nil {
		if !this.Testing { //测试场屏蔽掉 龙虎两边都压,totalin和totalout都=0,这个条件去掉
			//统计流水值
			playerEx := this.GetPlayer(snid)
			//有些结算的时候，玩家已经退场，不要用是否在游戏，0709，修改为扣税后数值
			if playerEx != nil && (param.TotalIn != 0 || param.TotalOut != 0) && !param.IsLeave && !playerEx.IsRob {
				totalFlow := param.ValidFlow * int64(this.DbGameFree.GetBetWaterRate()) / 100
				playerEx.TotalConvertibleFlow += totalFlow
				playerEx.TotalFlow += totalFlow
				playerEx.ValidCacheBetTotal += param.ValidBet
				//报表统计
				playerEx.SaveReportForm(int(this.DbGameFree.GetGameClass()), this.SceneMode, this.KeyGameId,
					param.WinAmountNoAnyTax, totalFlow, param.ValidBet)
				//分配利润
				ProfitDistribution(playerEx, param.TaxCoin, param.ClubPumpCoin, totalFlow)
				//上报游戏事件
				playerEx.ReportGameEvent(param.TaxCoin, param.ClubPumpCoin, param.WinAmountNoAnyTax, param.ValidBet, totalFlow, param.TotalIn, param.TotalOut)
			}

			roomType := int32(this.SceneMode)
			if this.GameId == common.GameId_Avengers ||
				this.GameId == common.GameId_CaiShen ||
				this.GameId == common.GameId_EasterIsland ||
				this.GameId == common.GameId_IceAge ||
				this.GameId == common.GameId_TamQuoc { //复仇者联盟强制为0，所有场次操作记录放一起
				roomType = 0
			}
			baseScore := this.DbGameFree.GetBaseScore()
			if common.IsLocalGame(this.GameId) {
				baseScore = this.BaseScore
			}
			log := model.NewGamePlayerListLogEx(snid, param.LogId, param.Platform, param.Channel, param.Promoter, param.PackageTag,
				int32(this.GameId), baseScore, int32(this.SceneId), this.DbGameFree.GetGameMode(),
				this.GetGameFreeId(), param.TotalIn, param.TotalOut, this.ClubId, this.RoomId, param.TaxCoin, param.ClubPumpCoin, roomType,
				param.BetAmount, param.WinAmountNoAnyTax, this.KeyGameId, playerEx.Name, this.DbGameFree.GetGameClass(),
				param.IsFirstGame, this.MatchId)
			if log != nil {
				LogChannelSington.WriteLog(log)
			}
		}
	}
}

func (this *Scene) IsPlayerFirst(p *Player) bool {
	if p == nil {
		return false
	}
	if p.GDatas != nil {
		if data, ok := p.GDatas[this.KeyGameId]; ok {
			if data.Statics.GameTimes <= 1 {
				return true
			}
			return false
		}
		return true
	}
	return false
}

func (this *Scene) RobotIsLimit() bool {
	if this.robotLimit != 0 {
		if this.IsCoinScene() {
			if this.robotLimit <= this.robotNum {
				return true
			}
			// 房间需要给真人留一个空位
			if this.DbGameFree.GetMatchTrueMan() == common.MatchTrueMan_Priority && this.playerNum-this.realPlayerNum-1 <= this.robotNum {
				return true
			}
		} else if this.IsHundredScene() {
			if this.robotNum >= this.robotLimit {
				return true
			}
		}
	}
	return false
}

func (this *Scene) RobotIsOverLimit() bool {
	if this.robotLimit != 0 {
		if this.IsCoinScene() {
			if this.robotLimit+1-this.realPlayerNum-this.robotNum < 0 {
				return true
			}
		} else if this.IsHundredScene() {
			if this.robotNum > this.robotLimit {
				return true
			}
		}
	}
	return false
}

func (this *Scene) RandRobotCnt() {
	if this.DbGameFree != nil {
		if this.DbGameFree.GetMatchMode() == 1 {
			return
		}
		numrng := this.DbGameFree.GetRobotNumRng()
		if len(numrng) >= 2 {
			if numrng[1] == numrng[0] {
				this.robotLimit = int(numrng[0])
			} else {
				if numrng[1] < numrng[0] {
					numrng[1], numrng[0] = numrng[0], numrng[1]
				}
				this.robotLimit = int(numrng[0] + this.Rand.Int31n(numrng[1]-numrng[0]+1))
			}
		}
		logger.Logger.Tracef("===(this *Scene) RandRobotCnt() sceneid:%v gameid:%v mode:%v  robotLimit:%v robotNum:%v", this.SceneId, this.GameId, this.GameMode, this.robotLimit, this.robotNum)
	}
}
func (this *Scene) GetRobotTime() int64 {
	l := int64(common.RandInt(model.NormalParamData.RobotRandomTimeMin, model.NormalParamData.RobotRandomTimeMax))
	return l + time.Now().Unix()
}
func (this *Scene) IsPreCreateScene() bool {
	return this.DbGameFree.GetCreateRoomNum() > 0
}
func (this *Scene) TryInviteRobot() {
	if this.aiMgr != nil {
		return
	}
	if this.DbGameFree == nil {
		return
	}

	//私有房间不邀请机器人
	if this.IsPrivateScene() || this.IsMatchScene() {
		return
	}
	if this.DbGameFree.GetMatchMode() == 1 {
		return
	}
	//if this.ClubScene != nil && this.ClubId != 0 {
	//	//俱乐部不进机器人
	//	return
	//}
	bot := int(this.DbGameFree.GetBot())
	if bot == 0 { //机器人不进的场
		return
	}
	if this.DbGameFree.GetMatchTrueMan() != common.MatchTrueMan_Forbid &&
		this.playerNum > 3 && common.IsCoinSceneType(this.DbGameFree.GetGameType()) {
		if len(this.Players) >= this.playerNum-1 {
			return
		}
	}

	//分组模式下机器人是否使用
	if !model.GameParamData.GameConfigGroupUseRobot && this.GroupId != 0 {
		return
	}

	//对战场有真实玩家的情况才需要机器人匹配
	if !this.IsRobFightGame() && this.realPlayerNum <= 0 && !this.IsHundredScene() && !this.IsPreCreateScene() { //预创建房间的对战场可以优先进机器人，如:21点 判断依据:CreateRoomNum
		return
	}

	switch this.DbGameFree.GetGameType() {
	case common.GameType_Fishing:
		if this.robotNum >= this.robotLimit {
			return
		}
	}

	tNow := time.Now()
	if tNow.Before(this.nextInviteTime) {
		return
	}
	if model.GameParamData.EnterAfterStartSwitch && this.IsCoinScene() && this.Gaming && !this.bEnterAfterStart {
		return
	}

	if this.robotNumLastInvite == this.robotNum {
		this.inviteInterval = this.inviteInterval + 1
		if this.inviteInterval > model.GameParamData.RobotInviteIntervalMax {
			this.inviteInterval = model.GameParamData.RobotInviteIntervalMax
		}
	} else {
		this.inviteInterval = model.GameParamData.RobotInviteInitInterval
	}

	this.ResetNextInviteTime()
	this.robotNumLastInvite = this.robotNum

	if !this.RobotIsLimit() {
		var robCnt int
		if this.robotLimit != 0 {
			if this.IsCoinScene() {
				if this.robotNum >= this.robotLimit { //机器人数量已达上限
					return
				}
				hadCnt := len(this.Players)
				robCnt = this.robotLimit - this.robotNum
				if robCnt > this.playerNum-hadCnt {
					robCnt = this.playerNum - hadCnt
				}
			} else if this.IsHundredScene() {
				robCnt = this.robotLimit - this.robotNum
			}
		} else {
			if this.IsCoinScene() {
				if this.IsFull() {
					return
				}
				hadCnt := len(this.Players)
				robCnt = this.playerNum - hadCnt
				if this.realPlayerNum == 0 { //一个真人都没有，不让机器人坐满房间
					robCnt--
				}
			}
		}
		if robCnt > 0 {
			num := this.Rand.Int31n(int32(robCnt + 1))
			if num > 0 {
				if this.IsCoinScene() /* && this.gaming*/ { //如果牌局正在进行中,一个一个进
					num = 1
				}
				//logger.Logger.Tracef("(this *Scene)(groupid:%v sceneid:%v) TryInviteRobot(%v) current robot(%v+%v) robotlimit(%v)", this.groupId, this.sceneId, this.dbGameFree.GetName()+this.dbGameFree.GetTitle(), this.robotNum, num, this.robotLimit)
				//同步下房间里的参数'
				if !RobotSceneDBGameFreeSync[this.SceneId] {
					pack := &server.GRGameFreeData{
						RoomId:     proto.Int(this.SceneId),
						DBGameFree: this.DbGameFree,
					}
					proto.SetDefaults(pack)
					if NpcServerAgentSington.SendPacket(int(server.SSPacketID_PACKET_GR_GameFreeData), pack) {
						RobotSceneDBGameFreeSync[this.SceneId] = true
					}
				}
				//然后再邀请
				if NpcServerAgentSington.Invite(int(this.SceneId), int(num), false, nil, this.DbGameFree.Id) {
				}
			}
		}
	}
}

func (this *Scene) ResetNextInviteTime() {
	this.nextInviteTime = time.Now().Add(time.Second * (time.Duration(this.Rand.Int63n(2 + this.inviteInterval))))
}

// 是否有真人参与游戏
func (this *Scene) IsRealInGame() bool {
	for _, player := range this.Players {
		if player != nil && player.IsGameing() && !player.IsRob {
			return true
		}
	}
	return false
}

// 是否都是真人
func (this *Scene) IsAllRealInGame() bool {
	for _, player := range this.Players {
		if player != nil && player.IsGameing() && player.IsRob {
			return false
		}
	}
	return true
}

// 是否开启机器人对战游戏
func (this *Scene) IsRobFightGame() bool {
	if this.DbGameFree == nil {
		return false
	}
	if this.DbGameFree.GetAi()[0] == 1 && model.GameParamData.IsRobFightTest == true {
		return true
	}
	return false
}

// 百人场机器人离场规则
func (this *Scene) RobotLeaveHundred() {
	for _, p := range this.Players {
		if p != nil {
			leave := false
			var reason int
			//if p.trusteeship >= 5 {
			//	leave = true
			//	reason = common.PlayerLeaveReason_LongTimeNoOp
			//}
			if !leave && !p.IsOnLine() && time.Now().Sub(p.DropTime) > 30*time.Second {
				leave = true
				reason = common.PlayerLeaveReason_DropLine
			}
			if !leave && p.IsRob {
				if !this.CoinInLimit(p.Coin) {
					//钱少
					leave = true
					reason = common.PlayerLeaveReason_Bekickout
				} else if this.CoinOverMaxLimit(p.Coin, p) {
					//钱多
					leave = true
					reason = common.PlayerLeaveReason_Normal
				} else if p.Coin < int64(this.DbGameFree.GetBetLimit()) {
					//少于下注限额
					leave = true
					reason = common.PlayerLeaveReason_Normal
				}
			}
			if leave {
				this.PlayerLeave(p, reason, leave)
			}
		}
	}
}
func (this *Scene) RandInt(args ...int) int {
	switch len(args) {
	case 0:
		return this.Rand.Int()
	case 1:
		if args[0] != 0 {
			return this.Rand.Intn(args[0])
		} else {
			return 0
		}
	default:
		l := args[0]
		u := args[1]
		switch {
		case l == u:
			{
				return l
			}
		case l > u:
			{
				return u + this.Rand.Intn(l-u)
			}
		default:
			{
				return l + this.Rand.Intn(u-l)
			}
		}
	}
}

func (this *Scene) CheckNeedDestroy() bool {
	if common.IsLocalGame(this.GameId) {
		return (ServerStateMgr.GetState() == common.GAME_SESS_STATE_OFF || this.graceDestroy)
	} else {
		return (ServerStateMgr.GetState() == common.GAME_SESS_STATE_OFF || this.graceDestroy) || (this.IsPrivateScene() && this.NumOfGames >= this.TotalOfGames)
	}
}

func (this *Scene) GetRecordId() string {
	if this.rr != nil {
		return this.rr.id
	}
	return fmt.Sprintf("%d%d%v%d", this.GameId, this.SceneId, this.GameNowTime.Format(ReplayIdTf), this.NumOfGames)
}

//func (this *Scene) TryUseMatchNextBaseData() {
//	data := this.matchChgData
//	if data != nil {
//		this.matchChgData = nil
//		this.SetParamEx(common.PARAMEX_MATCH_BASESCORE, data.NextBaseScore)
//		this.SetParamEx(common.PARAMEX_MATCH_OUTSCORE, data.NextOutScore)
//		this.DbGameFree.BaseScore = proto.Int32(data.NextBaseScore)
//		//同步给玩家
//		pack := &match.SCMatchBaseScoreChange{
//			MatchId:   proto.Int32(this.GetParamEx(common.PARAMEX_MATCH_COPYID)),
//			BaseScore: proto.Int32(data.NextBaseScore),
//			OutScore:  proto.Int32(data.NextOutScore),
//		}
//		proto.SetDefaults(pack)
//		this.Broadcast(int(match.MatchPacketID_PACKET_SC_MATCH_BASESCORECHANGE), pack, 0)
//	}
//}

// ///////////////////////////////////////////////////////////////////
func (this *Scene) resultHistoryRemove(n int) {
	var i int
	for ; i < len(this.resultHistory); i++ {
		for j := 1; j < len(this.resultHistory[i]); j++ {
			if this.resultHistory[i][j] == n {
				this.resultHistory[i] = append(this.resultHistory[i][:j], this.resultHistory[i][j+1:]...)
				if len(this.resultHistory[i]) <= 1 {
					this.resultHistory = append(this.resultHistory[:i], this.resultHistory[i+1:]...)
				}
				return
			}
		}
	}
}

func (this *Scene) GetResultHistoryResult(n int) int {
	for i := 0; i < len(this.resultHistory); i++ {
		for j := 1; j < len(this.resultHistory[i]); j++ {
			if this.resultHistory[i][j] == n {
				return this.resultHistory[i][0]
			}
		}
	}
	return -1
}
func (this *Scene) GetResult() int {
	return this.results[this.LoopNum]
}
func (this *Scene) resultCheck() {
	logger.Logger.Tracef("历史记录: %v\n调控配置: %v", this.resultHistory, this.results)
	for i := 0; i < len(this.resultHistory); i++ {
		m := this.resultHistory[i][0]
		for j := 1; j < len(this.resultHistory[i]); j++ {
			if this.results[this.resultHistory[i][j]] != m {
				logger.Logger.Errorf("不匹配 局数:%d 配置结果:%d 历史记录的结果:%d",
					this.resultHistory[i][j], this.results[this.resultHistory[i][j]], m)
			}
		}
	}
}

func (this *Scene) AddLoopNum() {
	//defer this.resultCheck()
	// 维护 resultHistory
	this.resultHistoryRemove(this.LoopNum)

	this.results[this.LoopNum] = common.DefaultResult
	this.LoopNum++
	if this.LoopNum == common.MaxLoopNum {
		this.LoopNum = 0
	}
}
func (this *Scene) ProtoResults() []*server.EResult {
	ret := []*server.EResult{}
	for i := 0; i < len(this.resultHistory); i++ {
		if len(this.resultHistory[i]) <= 1 {
			continue
		}
		temp := new(server.EResult)
		temp.Result = proto.Int(this.resultHistory[i][0])
		buf := bytes.NewBufferString("")
		buf.WriteString(fmt.Sprint(this.resultHistory[i][1]))
		for j := 2; j < len(this.resultHistory[i]); j++ {
			buf.WriteString(fmt.Sprint(",", this.resultHistory[i][j]))
		}
		temp.Index = proto.String(buf.String())
		ret = append(ret, temp)
	}
	return ret
}

func (this *Scene) ParserResults(str, user string) {
	for _, v := range strings.Split(strings.TrimSpace(str), ",") {
		ns := strings.Split(v, ":")
		if len(ns) != 2 {
			continue
		}
		i, err := strconv.Atoi(ns[0])
		if err != nil {
			continue
		}
		n, err := strconv.Atoi(ns[1])
		if err != nil {
			continue
		}
		if n < 0 || n >= len(this.results) {
			continue
		}
		this.results[i] = n
	}
	this.WebUser = user
}

// ParserResults1 只能解析一种控制结果的设置
func (this *Scene) ParserResults1(str, user string) *server.GWRoomResults {
	// 维护 resultHistory
	ret := &server.GWRoomResults{
		Code: proto.Int(0),
	}

	arr := strings.Split(strings.TrimSpace(str), ",")
	if len(arr) == 0 {
		ret.Code = proto.Int(3)
		ret.Msg = proto.String("参数错误")
		return ret
	}
	ns := strings.Split(arr[0], ":")
	if len(ns) != 2 {
		ret.Code = proto.Int(3)
		ret.Msg = proto.String("参数错误")
		return ret
	}
	n, err := strconv.Atoi(ns[1])
	if err != nil {
		ret.Code = proto.Int(3)
		ret.Msg = proto.String("参数错误")
		return ret
	}

	//logger.Logger.Tracef("控牌设置历史 %s", str)
	//defer this.resultCheck()

	var is []int
	if n == common.DefaultResult {
		for _, v := range arr {
			ns := strings.Split(v, ":")
			if len(ns) != 2 {
				continue
			}
			// 局数
			i, err := strconv.Atoi(ns[0])
			if err != nil {
				continue
			}
			if i < 0 || i >= len(this.results) {
				ret.Code = proto.Int(2)
				ret.Msg = proto.String(fmt.Sprintf("局数错误 第%d局", i))
				return ret
			}
			is = append(is, i)
		}
		for _, v := range is {
			this.resultHistoryRemove(v)
			this.results[v] = common.DefaultResult
		}
		this.WebUser = user
		return ret
	}

	// 设置
	indexes := []int{n}
	for _, v := range arr {
		ns := strings.Split(v, ":")
		if len(ns) != 2 {
			continue
		}
		// 局数
		i, err := strconv.Atoi(ns[0])
		if err != nil {
			continue
		}
		if i < 0 || i >= len(this.results) {
			ret.Code = proto.Int(2)
			ret.Msg = proto.String(fmt.Sprintf("局数错误 第%d局", i))
			return ret
		}
		if this.results[i] != common.DefaultResult {
			ret.Code = proto.Int(1)
			ret.Msg = proto.String(fmt.Sprintf("重复设置第%d局", i))
			return ret
		}
		is = append(is, i)
	}

	for _, v := range is {
		this.results[v] = n
		indexes = append(indexes, v)
	}
	this.resultHistory = append(this.resultHistory, indexes)
	this.WebUser = user
	return ret
}

func (this *Scene) RandTakeCoin(p *Player) (takeCoin, leaveCoin, gameTimes int64) {
	if p.IsRob && p.IsLocal {
		dbGameFree := this.DbGameFree
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
	return
}

func (this *Scene) TryBillExGameDrop(p *Player) {
	if p.IsRob {
		return
	}
	baseScore := this.DbGameFree.BaseScore
	if common.IsLocalGame(this.GameId) {
		baseScore = this.BaseScore
	}
	if baseScore == 0 {
		return
	}
	dropInfo := srvdata.GameDropMgrSington.GetDropInfoByBaseScore(int32(this.GameId), baseScore)
	if dropInfo != nil && len(dropInfo) != 0 && p.Items != nil {
		realDrop := make(map[int32]int32)
		for _, drop := range dropInfo {
			if _, ok := p.Items[drop.ItemId]; ok {
				//概率
				randTmp := rand.Int31n(10000)
				if randTmp < drop.Rate {
					//个数
					num := drop.MinAmount
					if drop.MaxAmount > drop.MinAmount {
						num = rand.Int31n(drop.MaxAmount-drop.MinAmount+1) + drop.MinAmount
					}

					p.Items[drop.ItemId] += num
					realDrop[drop.ItemId] = num
				}
			} else {
				logger.Logger.Error("itemid not exist! ", drop.ItemId)
			}
		}
		if realDrop != nil && len(realDrop) != 0 {
			//通知客户端游戏内额外掉落
			pack := &player.SCGameExDropItems{}
			pack.Items = make(map[int32]int32)
			for id, num := range realDrop {
				pack.Items[id] = proto.Int32(num)
				itemData := srvdata.PBDB_GameItemMgr.GetData(id)
				if itemData != nil {
					//logType  0获得 1消耗
					log := model.NewItemLogEx(p.Platform, p.SnId, 0, itemData.Id, itemData.Name, num, "tienlen游戏掉落")
					if log != nil {
						logger.Logger.Trace("WriteLog: ", log)
						LogChannelSington.WriteLog(log)
					}
				}
			}
			if pack != nil && pack.Items != nil && len(pack.Items) != 0 {
				p.SendToClient(int(player.PlayerPacketID_PACKET_SCGAMEEXDROPITEMS), pack)
				logger.Logger.Trace("SCGAMEEXDROPITEMS ", pack)
			}
		}
	}
}
