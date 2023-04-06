package base

import (
	"fmt"
	"games.yol.com/win88/protocol/gamehall"
	"games.yol.com/win88/protocol/mngame"
	"github.com/idealeak/goserver/srvlib"
	"math/rand"
	"strconv"
	"time"

	"games.yol.com/win88/common"
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/server"
	server_proto "games.yol.com/win88/protocol/server"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/core/utils"
	srvlibproto "github.com/idealeak/goserver/srvlib/protocol"
	rawproto "google.golang.org/protobuf/proto"
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

type Scene struct {
	ws                 *netlib.Session     //worldsrv session
	rand               *rand.Rand          //专属随机数
	ExtraData          interface{}         //具体游戏规则数据
	SceneId            int                 //房间id
	SceneMode          int                 //场次模式
	GameId             int                 //游戏id
	GameMode           int                 //游戏modeid
	Platform           string              //所属平台id
	state              int                 //房间状态
	Params             []int32             //游戏配置参数
	paramsEx           []int32             //其他扩展参数
	hallId             int32               //大厅id 同 gamefreeid
	playerNum          int                 //游戏人数
	realPlayerNum      int                 //真是玩家人数
	robotNum           int                 //机器人数量
	robotLimit         int                 //最大限制机器人数量
	robotNumLastInvite int                 //上次邀请机器人时的数量
	totalOfGames       int                 //游戏总局数
	NumOfGames         int                 //局数
	Players            map[int32]*Player   //房间内的玩家
	sp                 ScenePolicy         //场景游戏策略
	DbGameFree         *server.DB_GameFree //场次配置数据
	SceneState         SceneState          //场景状态
	StateStartTime     time.Time           //状态开始时间
	stateEndTime       time.Time           //状态结束时间
	GameStartTime      time.Time           //游戏开始计时时间
	GameNowTime        time.Time           //当局游戏开始时间
	nextInviteTime     time.Time           //下次邀请机器人时间
	inviteInterval     int64               //邀请间隔
	pause              bool                //是否暂停状态
	Gaming             bool                //游戏是否进行中
	destroyed          bool                //是否已销毁
	completed          bool                //牌局是否已完成
	Testing            bool                //是否为测试场
	graceDestroy       bool                //等待销毁
	replayAddId        int32               //录像自增id
	KeyGameId          string              //游戏ID
	KeyGamefreeId      string              //游戏分场次id
	GroupId            int32               //分组id
	bEnterAfterStart   bool                //是否允许中途加入
	CpCtx              model.CoinPoolCtx   //水池环境
	cpControlled       bool                //被水池控制了
	nogDismiss         int                 //检查机器人离场时的局数(同一局只检查一次)
	systemCoinOut      int64               //本局游戏机器人营收 机器人赢：正值	机器人输：负值
}

func NewScene(ws *netlib.Session, sceneId, gameMode, sceneMode, gameId int, platform string, params []int32,
	agentor, creator int32, replayCode string, hallId, groupId, totalOfGames int32, dbGameFree *server.DB_GameFree, bEnterAfterStart bool, paramsEx ...int32) *Scene {
	sp := GetScenePolicy(gameId, gameMode)
	if sp == nil {
		logger.Logger.Errorf("Game id %v not register in ScenePolicyPool.", gameId)
		return nil
	}
	tNow := time.Now()
	s := &Scene{
		ws:               ws,
		SceneId:          sceneId,
		SceneMode:        sceneMode,
		GameMode:         gameMode,
		GameId:           gameId,
		Params:           params,
		paramsEx:         paramsEx,
		Players:          make(map[int32]*Player),
		sp:               sp,
		GameStartTime:    tNow,
		hallId:           hallId,
		Platform:         platform,
		DbGameFree:       dbGameFree,
		inviteInterval:   model.GameParamData.RobotInviteInitInterval,
		GroupId:          groupId,
		bEnterAfterStart: bEnterAfterStart,
		totalOfGames:     int(totalOfGames),
	}
	if s != nil && s.init() {
		logger.Logger.Trace("NewScene init success.")
		return s
	} else {
		logger.Logger.Trace("NewScene init failed.")
		return nil
	}
}

//根据gamedifid，转为gameid,然后返回所有的相同gameid的数据
func (this *Scene) GetTotalTodayDaliyGameData(keyGameId string, pd *Player) *model.PlayerGameStatics {
	todayData := &model.PlayerGameStatics{}

	if pd.TodayGameData == nil {
		return todayData
	}

	if pd.TodayGameData.CtrlData == nil {
		return todayData
	}

	if info, ok := pd.TodayGameData.CtrlData[this.KeyGamefreeId]; ok {
		todayData.TotalIn += info.TotalIn
		todayData.TotalOut += info.TotalOut
	}

	return todayData
}

func (this *Scene) GetInit() bool {
	return this.init()
}
func (this *Scene) init() bool {
	tNow := time.Now()
	sceneRandSeed++
	this.rand = rand.New(rand.NewSource(sceneRandSeed))
	this.nextInviteTime = tNow.Add(time.Second * time.Duration(this.rand.Int63n(model.GameParamData.RobotInviteInitInterval)))
	this.RandRobotCnt()
	this.state = SCENE_STATE_INITED

	if len(this.paramsEx) != 0 {
		if this.DbGameFree.GetSceneType() == -1 {
			this.Testing = true
		} else {
			this.Testing = false
		}

		this.KeyGameId = this.DbGameFree.GetGameDif()
		this.KeyGamefreeId = strconv.Itoa(int(this.DbGameFree.GetId()))
	}

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

//////////////////////////////////////////////////
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

//////////////////////////////////////////////////

func (this *Scene) PlayerEnter(p *Player, isLoaded bool) {
	logger.Logger.Trace("(this *Scene) PlayerEnter:", isLoaded, this.SceneId, p.GetName())
	this.Players[p.SnId] = p
	p.scene = this

	if p.IsRob {
		this.robotNum++
		logger.Logger.Tracef("(this *Scene) PlayerEnter(%v) robot(%v) robotlimit(%v)", this.DbGameFree.GetName()+this.DbGameFree.GetTitle(), this.robotNum, this.robotLimit)
	} else {
		p.ValidCacheBetTotal = 0
		this.realPlayerNum++
		this.RandRobotCnt()
	}

	p.OnEnter(this)

	//避免游戏接口异常
	utils.RunPanicless(func() { this.sp.OnPlayerEnter(this, p) })

	if p.BlackLevel > 0 {
		//WarningBlackPlayer(p.SnId, this.GamefreeId)
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
		logger.Logger.Tracef("(this *Scene) Can't PlayerLeave(%v)  scene(%v)", p.SnId, this.SceneId)
		return
	}

	//避免游戏接口异常
	utils.RunPanicless(func() { this.sp.OnPlayerLeave(this, p, reason) })

	p.OnLeave(reason)
	delete(this.Players, p.SnId)
	isBill = true

	//send world离开房间
	pack := &server.GWPlayerLeaveMiniGame{
		SceneId:    proto.Int(this.SceneId),
		SnId:       proto.Int32(p.SnId),
		Reason:     proto.Int(reason),
		GameFreeId: proto.Int32(this.GetGameFreeId()),
	}
	proto.SetDefaults(pack)
	this.SendToWorld(int(server.SSPacketID_PACKET_GW_PLAYERLEAVE_MINIGAME), pack)

	if p.IsRob {
		this.robotNum--
		logger.Logger.Tracef("(this *Scene) PlayerLeave(%v) robot(%v) robotlimit(%v)", this.DbGameFree.GetName()+this.DbGameFree.GetTitle(), this.robotNum, this.robotLimit)
	} else {
		this.realPlayerNum--
		this.RandRobotCnt()
	}

	this.ResetNextInviteTime()
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
	}
}

func (this *Scene) PlayerRehold(snid int32, sid int64, gs *netlib.Session) {
	logger.Logger.Trace("(this *Scene) PlayerRehold")
	if p, exist := this.Players[snid]; exist {
		p.OnRehold(sid, gs)
		//避免游戏接口异常
		utils.RunPanicless(func() { this.sp.OnPlayerRehold(this, p) })
	}
}

//玩家返回房间的消息
func (this *Scene) PlayerReturn(p *Player, isLoaded bool) {
	logger.Logger.Trace("(this *Scene) PlayerReturn")

	//先处理return room 消息 然后 发送 返回房间消息
	pack := &gamehall.SCReturnRoom{
		RoomId:    proto.Int(this.SceneId),
		GameId:    proto.Int(this.GameId),
		ModeType:  proto.Int(this.GameMode),
		Params:    this.Params,
		HallId:    proto.Int32(this.hallId),
		IsLoaded:  proto.Bool(isLoaded),
		OpRetCode: gamehall.OpResultCode_Game_OPRC_Sucess_Game,
	}
	proto.SetDefaults(pack)
	p.SendToClient(int(gamehall.GameHallPacketID_PACKET_SC_RETURNROOM), pack)

	if this.HasPlayer(p) {
		//避免游戏接口异常
		//utils.RunPanicless(func() { this.sp.OnPlayerRehold(this, p) })
		//玩家返回房间的消息使用 return 和 rehold分开处理
		utils.RunPanicless(func() { this.sp.OnPlayerReturn(this, p) })
	}
}

func (this *Scene) Broadcast(packetid int, msg rawproto.Message, excludeSid int64, includeOffline ...bool) {
	//包装一下
	wrapPack := &mngame.SCMNGameDispatcher{
		Id: this.GetGameFreeId(),
	}
	if data, err := netlib.MarshalPacket(packetid, msg); err == nil {
		wrapPack.Data = data

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
			}
		}

		for gateSess, v := range mgs {
			if gateSess != nil && len(v) != 0 {
				pack, err := MulticastMaker.CreateMulticastPacket(int(mngame.MNGamePacketID_PACKET_SC_MNGAME_DISPATCHER), wrapPack, v...)
				if err == nil {
					proto.SetDefaults(pack)
					gateSess.Send(int(srvlibproto.SrvlibPacketID_PACKET_SS_MULTICAST), pack)
				}
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

func (this *Scene) BroadcastMessageToPlatform(packetid int, rawpack interface{}) bool {
	//暂时广播的奖池变动消息，是公共消息，所以不需要像处理小游戏协议一样再包一层
	pack := &server_proto.SSCustomTagMulticast{
		Tags: []string{this.Platform},
	}
	if byteData, ok := rawpack.([]byte); ok {
		pack.RawData = byteData
	} else {
		byteData, err := netlib.MarshalPacket(packetid, rawpack)
		if err == nil {
			pack.RawData = byteData
		} else {
			logger.Logger.Info("Scene.BroadcastMessageToPlatform err:", err)
			return false
		}
	}
	srvlib.ServerSessionMgrSington.Broadcast(int(server_proto.SSPacketID_PACKET_SS_CUSTOMTAG_MULTICAST), pack, common.GetSelfAreaId(), srvlib.GateServerType)
	return true
}

func (this *Scene) ChangeSceneState(stateid int) {
	if this.destroyed {
		return
	}
	state := this.sp.GetSceneState(this, stateid)
	if state == nil {
		return
	}
	if this.SceneState != nil {
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
}

func (this *Scene) Pause() {
	this.pause = true
}

func (this *Scene) Destroy(force bool) {

	this.destroyed = true
	this.pause = true

	for _, p := range this.Players {
		this.PlayerLeave(p, common.PlayerLeaveReason_OnDestroy, true)
	}

	isCompleted := this.sp.IsCompleted(this) || this.completed
	SceneMgrSington.DestroyScene(this.SceneId)
	pack := &server.GWDestroyMiniScene{
		SceneId: proto.Int(this.SceneId),
	}
	proto.SetDefaults(pack)
	this.SendToWorld(int(server.SSPacketID_PACKET_GW_DESTROYMINISCENE), pack)
	logger.Logger.Trace("(this *Scene) Destroy(force bool) isCompleted", isCompleted)
}

//是否公平竞争
func (this *Scene) IsFair() bool {
	//不公平竞争,关照过度亏损的玩家
	return false
}

//小游戏场
func (this *Scene) IsMiniGameScene() bool {
	return this.SceneId >= common.MiniGameSceneStartId && this.SceneId < common.MiniGameSceneMaxId
}

func (this *Scene) IsPrivateScene() bool {
	return this.SceneId >= common.PrivateSceneStartId && this.SceneId <= common.PrivateSceneMaxId
}

func (this *Scene) IsMatchScene() bool {
	return this.SceneId >= common.MatchSceneStartId && this.SceneId <= common.MatchSceneMaxId
}

func (this *Scene) IsFull() bool {
	return len(this.Players) >= this.playerNum
}

//大厅场
func (this *Scene) IsHallScene() bool {
	return this.SceneId >= common.HallSceneStartId && this.SceneId <= common.HallSceneMaxId
}

//金豆自由场
func (this *Scene) IsCoinScene() bool {
	return this.SceneId >= common.CoinSceneStartId && this.SceneId <= common.CoinSceneMaxId
}

//百人场
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
	min := int(this.GetCoinSceneLowerThanKick())
	max := int(this.GetCoinSceneMaxCoinLimit())
	if min != 0 && coin < int64(min) {
		return false
	}
	if max != 0 && coin > int64(max) {
		return false
	}
	if this.Testing && coin <= 0 {
		return false
	}
	return true
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

func (this *Scene) GetSceneName() string {
	return this.DbGameFree.GetName() + this.DbGameFree.GetTitle()
}

func (this *Scene) CanChangeCoinScene(p *Player) bool {
	if this.sp != nil {
		return this.sp.CanChangeCoinScene(this, p)
	}
	return false
}

func (this *Scene) SyncPlayerCoin() {
}

func (this *Scene) NotifySceneRoundStart(round int) {
	pack := &server.GWSceneStart{
		RoomId:    proto.Int(this.SceneId),
		CurrRound: proto.Int(round),
		Start:     proto.Bool(true),
		MaxRound:  proto.Int(this.totalOfGames),
	}
	proto.SetDefaults(pack)
	this.SendToWorld(int(server.SSPacketID_PACKET_GW_SCENESTART), pack)
}

func (this *Scene) NotifySceneRoundPause() {
	pack := &server.GWSceneStart{
		RoomId:    proto.Int(this.SceneId),
		Start:     proto.Bool(false),
		CurrRound: proto.Int(this.NumOfGames),
		MaxRound:  proto.Int(this.totalOfGames),
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

//防伙牌换桌
func (this *Scene) ChangeSceneEvent() {
}

func (this *Scene) TryDismissRob(params ...int) {

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
				rands := this.rand.Int63n(20) + 20
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
		content := fmt.Sprintf("%v|%v|%v", player.GetName(), this.GetSceneName(), num)
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

//保存详细游戏日志
func (this *Scene) SaveGameDetailedLog(logid string, gamedetailednote string, gameDetailedParam *GameDetailedParam) {
	if this != nil {
		if !this.Testing { //测试场屏蔽掉
			trend20Lately := gameDetailedParam.Trend20Lately
			log := model.NewGameDetailedLogEx(logid, int32(this.GameId), int32(this.SceneId),
				this.DbGameFree.GetGameMode(), this.GetGameFreeId(), int32(len(this.Players)),
				int32(time.Now().Unix()-this.GameNowTime.Unix()), this.DbGameFree.GetBaseScore(),
				gamedetailednote, this.Platform, 0, "", this.CpCtx, 0, trend20Lately)
			if log != nil {
				if this.IsMatchScene() {
					log.MatchId = this.GetParamEx(common.PARAMEX_MATCH_ID)
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
	WinAmountNoAnyTax int64  //税后赢取额
	ValidBet          int64  //有效下注
	ValidFlow         int64  //有效流水
	IsFirstGame       bool   //是否第一次游戏
	IsLeave           bool   //是否中途离开，用于金花，德州可以中途离开游戏使用
}

//保存玩家和GameDetailedLog的映射表
func (this *Scene) SaveGamePlayerListLog(snid int32, param *SaveGamePlayerListLogParam) {
	if !this.Testing { //测试场屏蔽掉 龙虎两边都压,totalin和totalout都=0,这个条件去掉
		//统计流水值
		playerEx := this.GetPlayer(snid)
		if playerEx != nil {
			totalFlow := param.ValidFlow * int64(this.DbGameFree.GetBetWaterRate()) / 100
			playerEx.TotalConvertibleFlow += totalFlow
			playerEx.TotalFlow += totalFlow
			playerEx.ValidCacheBetTotal += param.ValidBet

			playerEx.SaveReportForm(int(this.DbGameFree.GetGameClass()), 0, this.KeyGameId,
				param.TotalOut-param.TotalIn-param.TaxCoin-param.ClubPumpCoin, totalFlow, param.ValidBet)

			//分配利润
			ProfitDistribution(playerEx, param.TaxCoin, param.ClubPumpCoin, totalFlow)
			//上报游戏事件
			playerEx.ReportGameEvent(param.TaxCoin, param.ClubPumpCoin, param.WinAmountNoAnyTax, param.ValidBet, totalFlow, param.TotalIn, param.TotalOut)
		}

		log := model.NewGamePlayerListLogEx(snid, param.LogId, param.Platform, param.Channel, param.Promoter, param.PackageTag,
			int32(this.GameId), this.DbGameFree.GetBaseScore(), int32(this.SceneId), this.DbGameFree.GetGameMode(),
			this.GetGameFreeId(), param.TotalIn, param.TotalOut, 0, "", param.TaxCoin, param.ClubPumpCoin, this.DbGameFree.SceneType,
			param.BetAmount, param.WinAmountNoAnyTax, this.KeyGameId, playerEx.Name, this.DbGameFree.GetGameClass(),
			param.IsFirstGame)
		if log != nil {
			LogChannelSington.WriteLog(log)
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
		if this.robotNum >= this.robotLimit {
			return true
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
				this.robotLimit = int(numrng[0] + this.rand.Int31n(numrng[1]-numrng[0]+1))
			}
		}
		logger.Logger.Tracef("===(this *Scene) RandRobotCnt() sceneid:%v gameid:%v mode:%v  robotLimit:%v robotNum:%v", this.SceneId, this.GameId, this.SceneMode, this.robotLimit, this.robotNum)
	}
}

func (this *Scene) IsPreCreateScene() bool {
	return true
}

func (this *Scene) TryInviteRobot() {
	if this.DbGameFree == nil {
		return
	}

	bot := int(this.DbGameFree.GetBot())
	if bot == 0 { //机器人不进的场
		return
	}

	if this.DbGameFree.GetMatchTrueMan() != common.MatchTrueMan_Forbid &&
		common.IsCoinSceneType(this.DbGameFree.GetGameType()) {
		if len(this.Players) >= this.playerNum-1 {
			return
		}
	}

	//对战场有真实玩家的情况才需要机器人匹配
	if !this.IsRobFightGame() && this.realPlayerNum <= 0 && !this.IsHundredScene() && !this.IsPreCreateScene() { //预创建房间的对战场可以优先进机器人，如:21点 判断依据:CreateRoomNum
		return
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
			num := this.rand.Int31n(int32(robCnt + 1))
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
				if NpcServerAgentSington.Invite(int(this.SceneId), int(num), false, nil, this.GetGameFreeId()) {
				}
			}
		}
	}
}

func (this *Scene) ResetNextInviteTime() {
	this.nextInviteTime = time.Now().Add(time.Second * (time.Duration(this.rand.Int63n(2 + this.inviteInterval))))
}

//是否有真人参与游戏
func (this *Scene) IsRealInGame() bool {
	for _, player := range this.Players {
		if player != nil && player.IsGameing() && !player.IsRob {
			return true
		}
	}
	return false
}

//是否都是真人
func (this *Scene) IsAllRealInGame() bool {
	for _, player := range this.Players {
		if player != nil && player.IsGameing() && player.IsRob {
			return false
		}
	}
	return true
}

//是否开启机器人对战游戏
func (this *Scene) IsRobFightGame() bool {
	if this.DbGameFree == nil {
		return false
	}
	if this.DbGameFree.GetAi()[0] == 1 && model.GameParamData.IsRobFightTest == true {
		return true
	}
	return false
}

//百人场机器人离场规则
func (this *Scene) RobotLeaveHundred() {
	for _, p := range this.Players {
		if p != nil {
			leave := false
			var reason int
			if !leave && !p.IsOnLine() && time.Now().Sub(p.dropTime) > 30*time.Second {
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
		return this.rand.Int()
	case 1:
		if args[0] != 0 {
			return this.rand.Intn(args[0])
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
				return u + this.rand.Intn(l-u)
			}
		default:
			{
				return l + this.rand.Intn(u-l)
			}
		}
	}
}

func (this *Scene) CheckNeedDestroy() bool {
	return (ServerStateMgr.GetState() == common.GAME_SESS_STATE_OFF || this.graceDestroy)
}

func (this *Scene) GetRecordId() string {
	return fmt.Sprintf("%d%d%v%d", this.GameId, this.SceneId, this.GameNowTime.Format(ReplayIdTf), this.NumOfGames)
}
