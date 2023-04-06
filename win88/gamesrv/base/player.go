package base

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"games.yol.com/win88/common"
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/player"
	"games.yol.com/win88/protocol/server"
	"games.yol.com/win88/srvdata"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/core/timer"
	rawproto "google.golang.org/protobuf/proto"
	"math/rand"
	"strconv"
	"time"
)

// 对应到客户端的一个玩家对象.

const (
	PlayerState_Online           int = 1 << iota //在线标记 1
	PlayerState_Ready                            //准备标记 2
	PlayerState_SceneOwner                       //房主标记 3
	PlayerState_Choke                            //呛标记 被复用于金花，是否被动弃牌 4
	PlayerState_Ting                             //听牌标记 5  金花复用，标记最后押注时，是否看牌
	PlayerState_NoisyBanker                      //闹庄标记 6  金花复用，标记allin时，是否看牌
	PlayerState_WaitOp                           //等待操作标记 7
	PlayerState_Auto                             //托管状态 8
	PlayerState_Check                            //已看牌状态 9
	PlayerState_Fold                             //弃牌状态 10
	PlayerState_Lose                             //输状态 11
	PlayerState_Win                              //赢状态 12
	PlayerState_WaitNext                         //等待下一局游戏 13
	PlayerState_GameBreak                        //不能继续游戏 14
	PlayerState_Leave                            //暂离状态 15
	PlayerState_Audience                         //观众标记 16
	PlayerState_AllIn                            //allin标记 17
	PlayerState_FinalAllIn                       //最后一圈，最后一个人allin标记 18
	PlayerState_Show                             //亮牌标记 19
	PlayerState_EnterSceneFailed                 //进场失败 20
	PlayerState_PKLost                           //发起Pk,失败 21
	PlayerState_IsChangeCard                     //牛牛标识是否换牌 22
	PlayerState_IsPayChangeCard                  //牛牛标识是否充值换牌 23
	PlayerState_Bankruptcy                       //玩家破产 24
	PlayerState_MatchQuit                        //退赛标记 25
	PlayerState_AllFollow                        //跟到底状态 26
	PlayerState_SAdjust                          //单控状态 27
	PlayerState_Max
)

// 玩家事件
const (
	PlayerEventEnter      int = iota //进入事件
	PlayerEventLeave                 //离开事件
	PlayerEventDropLine              //掉线
	PlayerEventRehold                //重连
	PlayerEventReturn                //返回房间 gs 添加
	PlayerEventRecharge              //冲值事件
	PlayerEventAddCoin               //其他加减币事件(例如:小游戏)
	AudienceEventEnter               //观众进入事件
	AudienceEventLeave               //观众离开事件
	AudienceEventDropLine            //观众掉线
	AudienceEventRehold              //观众重连
)

type Player struct {
	model.PlayerData                     //po 持久化对象
	ExtraData        interface{}         //扩展接口
	gateSess         *netlib.Session     //所在GateServer的session
	worldSess        *netlib.Session     //所在WorldServer的session
	scene            *Scene              //当前所在个Scene
	ai               AI                  //ai接口
	sid              int64               //对应客户端的sessionId
	gateSid          int64               //对应网关的sessionId
	Longitude        int32               //经纬度
	Latitude         int32               //经纬度
	city             string              //城市
	flag             int                 //状态标记
	Pos              int                 //当前位置
	dirty            bool                //脏标记
	Billed           bool                //是否已经结算过了
	AgentCode        string              //代理商编号
	Coin             int64               //金币
	serviceFee       int64               //服务费|税收
	TotalBet         int64               //总下注额（从进房间开始,包含多局游戏的下注）
	disbandGen       int                 //第几次解散申请
	hAuto            timer.TimerHandle   //托管handle
	GameTimes        int32               //游戏次数
	winTimes         int                 //胜利次数
	lostTimes        int                 //失败次数
	ActiveLeave      bool                //主动暂离
	OpCode           player.OpResultCode //错误码
	takeCoin         int64               //携带金币
	ExpectLeaveCoin  int64               //期望离场时的金币[机器人用]
	ExpectGameTime   int32               //期望进行的局数[机器人用]
	CurIsWin         int64               //当局输赢   负数：输  0：平局  正数：赢
	currentCoin      int64               //本局结束后剩余
	CurrentBet       int64               //本局下注额
	CurrentTax       int64               //本局税收
	//StartCoin          int64               //本局开始金币
	LastSyncCoin       int64             //
	IsQM               bool              //是否是全民推广用户
	LastOPTimer        time.Time         //玩家最后一次操作时间
	Trusteeship        int32             //玩家托管了几局
	ValidCacheBetTotal int64             //有效下注缓存
	isFightRobot       bool              //测试机器人，这种机器人可以用作记录水池数据，方便模拟用户输赢
	DropTime           time.Time         //掉线时间
	cparams            map[string]string //平台登陆数据
	Iparams            map[int]int64     //整形参数
	sparams            map[int]string    //字符参数
	WhiteLevel         int32
	BlackLevel         int32
	SingleAdjust       *model.PlayerSingleAdjust
	IsLocal            bool            //是否本地player
	Items              map[int32]int32 //背包数据
	MatchParams        []int32         //比赛参数
}

func NewPlayer(sid int64, data []byte, ws, gs *netlib.Session) *Player {
	p := &Player{
		sid:       sid,
		worldSess: ws,
		gateSess:  gs,
		flag:      PlayerState_Online,
		Pos:       -1,
		Longitude: -1,
		Latitude:  -1,
		cparams:   make(map[string]string), //平台登陆数据
		Iparams:   make(map[int]int64),     //整形参数
		sparams:   make(map[int]string),    //字符参数
	}
	if p.init(data) {
		return p
	}

	return nil
}

func NewLocalPlayer(snid int32) *Player {
	p := &Player{
		flag:      PlayerState_Online,
		Pos:       -1,
		Longitude: -1,
		Latitude:  -1,
		IsLocal:   true,
		cparams:   make(map[string]string), //平台登陆数据
		Iparams:   make(map[int]int64),     //整形参数
		sparams:   make(map[int]string),    //字符参数
	}
	pd := model.NewPlayerData(fmt.Sprintf("%d", snid), "贵宾", snid, common.Channel_Rob, common.Platform_Rob, "", 0,
		0, "", "", "", "", 0, fmt.Sprintf("%d", snid), "", "", 0)
	if pd == nil {
		return nil
	}

	p.PlayerData = *pd
	p.RobotRandName()
	if p.init(nil) {
		return p
	}

	return nil
}

func (this *Player) init(data []byte) bool {
	if !this.UnmarshalData(data) {
		return false
	}
	if this.GMLevel > 2 {
		this.Longitude = rand.Int31n(114103930-113216260) + 113216260
		this.Latitude = rand.Int31n(34963671-34592702) + 34592702
	}
	this.city = this.City
	this.LastOPTimer = time.Now()
	if this.GDatas == nil {
		this.GDatas = make(map[string]*model.PlayerGameInfo)
	}
	if this.WBLevel > 0 {
		this.WhiteLevel = this.WBLevel
	} else if this.WBLevel < 0 {
		this.BlackLevel = -this.WBLevel
	}
	return true
}
func (this *Player) MarkFlag(flag int) {
	this.flag |= flag
	switch flag {
	case PlayerState_Online, PlayerState_Ready, PlayerState_Leave:
		this.SyncFlagToWorld()
	}
}
func (this *Player) UnmarkFlag(flag int) {
	this.flag &= ^flag
	switch flag {
	case PlayerState_Online, PlayerState_Ready, PlayerState_Leave:
		this.SyncFlagToWorld()
	}
}
func (this *Player) IsMarkFlag(flag int) bool {
	if (this.flag & flag) != 0 {
		return true
	}
	return false
}
func (this *Player) IsOnLine() bool {
	return this.IsMarkFlag(PlayerState_Online)
}

func (this *Player) IsReady() bool {
	return this.IsMarkFlag(PlayerState_Ready)
}

func (this *Player) IsSceneOwner() bool {
	return this.IsMarkFlag(PlayerState_SceneOwner)
}

func (this *Player) IsAuto() bool {
	return this.IsMarkFlag(PlayerState_Auto)
}

func (this *Player) IsGameing() bool {
	return !this.IsMarkFlag(PlayerState_WaitNext) && !this.IsMarkFlag(PlayerState_GameBreak) && !this.IsMarkFlag(PlayerState_Bankruptcy) && !this.IsMarkFlag(PlayerState_Audience)
}

func (this *Player) IsAllFollow() bool {
	return this.IsMarkFlag(PlayerState_AllFollow)
}

func (this *Player) SyncFlag(onlysync2me ...bool) {
	if this.IsLocal {
		return
	}
	pack := &player.SCPlayerFlag{
		PlayerId: proto.Int32(this.SnId),
		Flag:     proto.Int(this.flag),
	}
	proto.SetDefaults(pack)
	if len(onlysync2me) != 0 {
		this.SendToClient(int(player.PlayerPacketID_PACKET_SC_PLAYERFLAG), pack)
	} else {
		this.Broadcast(int(player.PlayerPacketID_PACKET_SC_PLAYERFLAG), pack, 0)
	}
	//logger.Logger.Trace("SyncFlag:", pack)
}

func (this *Player) SyncFlagToWorld() {
	if this.IsLocal {
		return
	}
	if this.scene == nil || this.scene.IsCoinScene() || this.scene.IsMatchScene() || this.scene.IsHundredScene() {
		return
	}
	pack := &server.GWPlayerFlag{
		SnId:   proto.Int32(this.SnId),
		RoomId: proto.Int(this.scene.SceneId),
		Flag:   proto.Int(this.flag),
	}
	proto.SetDefaults(pack)
	this.SendToWorld(int(server.SSPacketID_PACKET_GW_PLAYERSTATE), pack)
	logger.Logger.Trace("SyncFlag to world:", pack)
}
func (this *Player) SendToClient(packetid int, rawpack interface{}, forceIgnore ...bool) bool {
	if this.IsLocal {
		return true
	}
	if !this.scene.Testing && this.scene.Gaming && this.scene.rr != nil && this.Pos != -1 && len(forceIgnore) == 0 && !this.scene.IsHundredScene() {
		this.scene.rr.Record(this.Pos, -1, packetid, rawpack)
	}
	if this.gateSess == nil {
		logger.Logger.Warnf("(this *Player) SendToClient [snid:%v packetid:%v] gatesess == nil ", this.SnId, packetid)
		return false
	}
	if rawpack == nil {
		logger.Logger.Tracef("(this *Player) SendToClient [snid:%v packetid:%v] rawpack == nil ", this.SnId, packetid)
		return false
	}
	if !this.IsOnLine() {
		logger.Logger.Warnf("(this *Player) SendToClient [snid:%v packetid:%v] Player if offline.", this.SnId, packetid)
		return false
	}
	if this.IsMarkFlag(PlayerState_Leave) {
		logger.Logger.Warnf("(this *Player) SendToClient [snid:%v packetid:%v] Player if leave.", this.SnId, packetid)
		return false
	}
	//logger.Logger.Trace("Send to player's packet:", packetid)
	return common.SendToGate(this.sid, packetid, rawpack, this.gateSess)
}

func (this *Player) Broadcast(packetid int, rawpack interface{}, excludeSid int64) bool {
	if this.scene != nil {
		this.scene.Broadcast(packetid, rawpack.(rawproto.Message), excludeSid)
		return true
	}
	return false
}

func (this *Player) SendToWorld(packetid int, rawpack interface{}) bool {
	if this.IsLocal {
		return true
	}
	if this.worldSess == nil {
		logger.Logger.Tracef("(this *Player) SendToWorld [%v] worldsess == nil ", this.Name)
		return false
	}
	if rawpack == nil {
		logger.Logger.Trace("(this *Player) SendToWorld rawpack == nil ")
		return false
	}

	this.scene.SendToWorld(packetid, rawpack)
	return true
}

func (this *Player) OnEnter(s *Scene) {
	this.scene = s
	//标记房主
	if this.SnId == s.Creator {
		this.MarkFlag(PlayerState_SceneOwner)
	}
}

func (this *Player) OnAudienceEnter(s *Scene) {
	this.scene = s
}

func (this *Player) OnRehold(newSid int64, newSess *netlib.Session) {
	this.sid = newSid
	this.gateSess = newSess
	this.MarkFlag(PlayerState_Online)
	// 2018-4-25
	// 这里先注释掉，暂离的状态在LeaveRoom和ReturnRoom的消息中设置
	// 如果这里清除暂离，那么在棋牌馆中，离开房间，用的状态更新为暂离状态，断线重连以后的话，在棋牌馆大厅界面看到的用户为非暂离状态，
	// 所以这里就先注释掉，让暂离的状态在离开和返回房间的消息中成对出现
	this.UnmarkFlag(PlayerState_Leave)
	this.SyncFlag()
}

func (this *Player) OnDropLine() {
	this.UnmarkFlag(PlayerState_Online)
	if !this.scene.Gaming && this.IsReady() && !this.scene.IsMatchScene() {
		this.UnmarkFlag(PlayerState_Ready)
	}
	this.SyncFlag()

	//存在假掉线的可能吗？
	//this.gateSess = nil
	//this.sid = 0
}

func (this *Player) OnAudienceDropLine() {
	this.gateSess = nil
}

func (this *Player) OnLeave(reason int) {
	unbindGateSess := true
	PlayerMgrSington.DelPlayerBySnId(this.SnId)
	//解绑gamesession
	if unbindGateSess && this.gateSess != nil {
		pack := &server.GGPlayerSessionUnBind{
			Sid: proto.Int64(this.sid),
		}
		proto.SetDefaults(pack)
		this.gateSess.Send(int(server.SSPacketID_PACKET_GG_PLAYERSESSIONBIND), pack)
	}
}

func (this *Player) OnAudienceLeave(reason int) {
	PlayerMgrSington.DelPlayerBySnId(this.SnId)
	//解绑gamesession
	if this.gateSess != nil {
		pack := &server.GGPlayerSessionUnBind{
			Sid: proto.Int64(this.sid),
		}
		proto.SetDefaults(pack)
		this.gateSess.Send(int(server.SSPacketID_PACKET_GG_PLAYERSESSIONBIND), pack)
	}
}

func (this *Player) MarshalData(gameid int) (d []byte, e error) {
	d, e = netlib.Gob.Marshal(&this.PlayerData)
	logger.Logger.Trace("(this *Player) MarshalData(gameid int)")
	return
}

func (this *Player) UnmarshalData(data []byte) bool {
	if len(data) == 0 {
		return true
	}
	err := netlib.Gob.Unmarshal(data, &this.PlayerData)
	if err == nil {
		this.dirty = true
		return true
	} else {
		logger.Logger.Warn("Player.SyncData err:", err)
	}
	return false
}
func (this *Player) UnMarshalSingleAdjustData(data []byte) bool {
	err := netlib.Gob.Unmarshal(data, &this.SingleAdjust)
	if err == nil {
		return true
	} else {
		logger.Logger.Warn("Player.SingleAdjust err:", err)
	}
	return false
}
func (this *Player) IsSingleAdjustPlayer() (*model.PlayerSingleAdjust, bool) {
	if this.SingleAdjust == nil {
		return nil, false
	}
	return this.SingleAdjust, true
}
func (this *Player) AddAdjustCount(gameFreeId int32) {
	sa := this.SingleAdjust
	if sa != nil && sa.GameFreeId == gameFreeId {
		this.SingleAdjust.CurTime++
		//通知world
		pack := &server.GWAddSingleAdjust{
			Platform:   this.Platform,
			SnId:       this.SnId,
			GameFreeId: gameFreeId,
		}
		this.SendToWorld(int(server.SSPacketID_PACKET_GW_ADDSINGLEADJUST), pack)
	}
}
func (this *Player) UpsertSingleAdjust(msg *model.PlayerSingleAdjust) {
	this.SingleAdjust = msg
}
func (this *Player) DeleteSingleAdjust(platform string, gamefreeid int32) {
	c := this.SingleAdjust
	if c != nil && c.Platform == platform && c.GameFreeId == gamefreeid {
		this.SingleAdjust = nil
	}
}
func (this *Player) OnSecTimer() {
}

func (this *Player) OnMiniTimer() {
}

func (this *Player) OnHourTimer() {
}

func (this *Player) OnDayTimer() {
	//在线跨天 数据给昨天，今天置为空
	this.YesterdayGameData = this.TodayGameData
	this.TodayGameData = model.NewPlayerGameCtrlData()
	/*
		for k, v := range this.YesterdayGameData.CtrlData {
			t := &model.PlayerGameStatics{}
			t.AvgBetCoin = v.AvgBetCoin
			this.TodayGameData.CtrlData[k] = t
		}
	*/
}

func (this *Player) OnMonthTimer() {
}

func (this *Player) OnWeekTimer() {
}

func (this *Player) GetName() string {
	return this.Name
}

func (this *Player) MarkDirty() {
	this.dirty = true
}

const (
	SyncFlag_ToClient  = 1 << iota //同步给客户端
	SyncFlag_ToWorld               //同步给服务端
	SyncFlag_Broadcast             //广播给房间内的用户
)

func (this *Player) AddCoin(num int64, gainWay int32, syncFlag int, oper, remark string) {
	if num == 0 {
		return
	}
	this.Coin += num
	if this.scene != nil {
		if !this.IsRob && !this.scene.Testing { //机器人log排除掉
			log := model.NewCoinLogEx(this.SnId, num, this.Coin, this.SafeBoxCoin,
				this.Ver, gainWay, int32(this.scene.GameId), oper, remark, this.Platform,
				this.Channel, this.BeUnderAgentCode, 0, this.PackageID, int32(this.scene.SceneId))
			if log != nil {
				this.GameCoinTs = log.Time.UnixNano()
				this.dirty = true
				LogChannelSington.WriteLog(log)
			}
		}
		//确保金币场金币数量不小于0
		if this.Coin < 0 {
			this.Coin = 0
		}
		if this.scene.IsHundredScene() {
			this.scene.NewBigCoinNotice(this, int64(num), 5)
		}
	}
	if (syncFlag & SyncFlag_ToClient) != 0 {
		pack := &player.SCPlayerCoinChange{
			SnId:     proto.Int32(this.SnId),
			AddCoin:  proto.Int64(num),
			RestCoin: proto.Int64(this.Coin),
		}
		proto.SetDefaults(pack)
		if (syncFlag & SyncFlag_Broadcast) != 0 {
			this.Broadcast(int(player.PlayerPacketID_PACKET_SC_PLAYERCOINCHANGE), pack, 0)
		} else {
			this.SendToClient(int(player.PlayerPacketID_PACKET_SC_PLAYERCOINCHANGE), pack)
		}
		logger.Logger.Trace("(this *Player) AddCoin SCPlayerCoinChange:", pack)
	}
}

func (this *Player) AddCoinNoLog(num int64, syncFlag int) {
	if num == 0 {
		return
	}
	this.Coin += num
	if this.scene != nil {
		if !this.IsRob && !this.scene.Testing { //机器人log排除掉
			this.dirty = true
		}
	}
	if (syncFlag & SyncFlag_ToClient) != 0 {
		pack := &player.SCPlayerCoinChange{
			SnId:     proto.Int32(this.SnId),
			AddCoin:  proto.Int64(num),
			RestCoin: proto.Int64(this.Coin),
		}
		proto.SetDefaults(pack)
		if (syncFlag & SyncFlag_Broadcast) != 0 {
			this.Broadcast(int(player.PlayerPacketID_PACKET_SC_PLAYERCOINCHANGE), pack, 0)
		} else {
			this.SendToClient(int(player.PlayerPacketID_PACKET_SC_PLAYERCOINCHANGE), pack)
		}
		logger.Logger.Trace("(this *Player) AddCoinNoLog SCPlayerCoinChange:", pack)
	}
}

func (this *Player) AddCoinAsync(num int64, gainWay int32, notifyC, broadcast bool, oper, remark string, writeLog bool) {
	if num == 0 {
		return
	}
	this.Coin += num
	if this.scene != nil {
		if !this.IsRob && !this.scene.Testing && writeLog { //机器人log排除掉
			log := model.NewCoinLogEx(this.SnId, int64(num), this.Coin, this.SafeBoxCoin,
				this.Ver, gainWay, int32(this.scene.GameId), oper, remark, this.Platform,
				this.Channel, this.BeUnderAgentCode, 0, this.PackageID, int32(this.scene.SceneId))
			if log != nil {
				this.GameCoinTs = log.Time.UnixNano()
				this.dirty = true
				LogChannelSington.WriteLog(log)
			}
		}
		//确保金币场金币数量不小于0
		if this.Coin < 0 {
			this.Coin = 0
		}
	}
	if notifyC {
		pack := &player.SCPlayerCoinChange{
			SnId:     proto.Int32(this.SnId),
			AddCoin:  proto.Int64(num),
			RestCoin: proto.Int64(this.Coin),
		}
		proto.SetDefaults(pack)
		if broadcast {
			this.Broadcast(int(player.PlayerPacketID_PACKET_SC_PLAYERCOINCHANGE), pack, 0)
		} else {
			this.SendToClient(int(player.PlayerPacketID_PACKET_SC_PLAYERCOINCHANGE), pack)
		}
	}
}

// 保存金币变动日志
// 数据用途: 个人房间内牌局账变记录，后台部分报表使用，确保数据计算无误，否则可能影响月底对账
// takeCoin: 牌局结算前玩家身上的金币
// changecoin:  本局玩家输赢的钱，注意是税后
// coin: 结算后玩家当前身上的金币余额
// totalbet: 总下注额
// taxcoin: 本局该玩家产生的税收，这里要包含俱乐部的税
// wincoin: 本局赢取的金币，含税 wincoin==changecoin+taxcoin
// jackpotWinCoin: 从奖池中赢取的金币(拉霸类游戏)
// smallGameWinCoin: 小游戏赢取的金币(拉霸类游戏)
func (this *Player) SaveSceneCoinLog(takeCoin, changecoin, coin, totalbet, taxcoin, wincoin int64, jackpotWinCoin int64, smallGameWinCoin int64) {
	if this.scene != nil {
		if !this.IsRob && !this.scene.Testing && !this.scene.IsMatchScene() { //机器人log排除掉
			var eventType int64 //输赢事件值 默认值为0
			if coin-takeCoin > 0 {
				eventType = 1
			} else if coin-takeCoin < 0 {
				eventType = -1
			}
			log := model.NewSceneCoinLogEx(this.SnId, changecoin, takeCoin, coin, eventType,
				int64(this.scene.DbGameFree.GetBaseScore()), totalbet, int32(this.scene.GameId), this.PlayerData.Ip,
				this.scene.paramsEx[0], this.Pos, this.Platform, this.Channel, this.BeUnderAgentCode, int32(this.scene.SceneId),
				this.scene.DbGameFree.GetGameMode(), this.scene.GetGameFreeId(), taxcoin, wincoin,
				jackpotWinCoin, smallGameWinCoin, this.PackageID)
			if log != nil {
				LogChannelSington.WriteLog(log)
			}
		}
	}
}

// 需要关照
func (this *Player) IsNeedCare() bool {
	return false
}

// 需要削弱
func (this *Player) IsNeedWeaken() bool {
	return false
}

func (this *Player) GetCoinOverPercent() int32 {
	return 0
}

func (this *Player) SyncCoin() {
	pack := &player.SCPlayerCoinChange{
		SnId:     proto.Int32(this.SnId),
		AddCoin:  proto.Int64(0),
		RestCoin: proto.Int64(this.Coin),
	}
	proto.SetDefaults(pack)
	this.SendToClient(int(player.PlayerPacketID_PACKET_SC_PLAYERCOINCHANGE), pack)
	logger.Logger.Trace("(this *Player) SyncCoin SCPlayerCoinChange:", pack)
}

func (this *Player) ReportGameEvent(tax, taxex, changeCoin, validbet, validFlow, in, out int64) {
	// 记录玩家 首次参与该场次的游戏时间 游戏次数
	data, ok := this.GDatas[this.scene.KeyGamefreeId]
	if !ok {
		data = &model.PlayerGameInfo{FirstTime: time.Now(), Statics: model.PlayerGameStatics{GameTimes: 1}}
		this.GDatas[this.scene.KeyGamefreeId] = data
	} else {
		if data.Statics.GameTimes <= 0 {
			data.FirstTime = time.Now()
		}
		data.Statics.GameTimes++
	}

	// 记录玩家 首次参与该游戏时间 游戏次数(不区分场次)
	dataGame, ok := this.GDatas[this.scene.KeyGameId]
	if !ok {
		dataGame = &model.PlayerGameInfo{FirstTime: time.Now(), Statics: model.PlayerGameStatics{GameTimes: 1}}
		this.GDatas[this.scene.KeyGameId] = data
	} else {
		if dataGame.Statics.GameTimes <= 0 {
			dataGame.FirstTime = time.Now()
		}
		dataGame.Statics.GameTimes++
	}

	gamingTime := int32(time.Now().Sub(this.scene.GameNowTime).Seconds())
	LogChannelSington.WriteMQData(model.GenerateGameEvent(model.CreatePlayerGameRecEvent(this.SnId, tax, taxex, changeCoin, validbet, validFlow, in, out,
		int32(this.scene.GameId), this.scene.DbGameFree.GetId(), int32(this.scene.GameMode),
		this.scene.GetRecordId(), this.Channel, this.BeUnderAgentCode, this.Platform, this.City, this.DeviceOS,
		this.CreateTime, gamingTime, data.FirstTime.Local(), dataGame.FirstTime.Local(), data.Statics.GameTimes, dataGame.Statics.GameTimes, this.LastLoginTime,
		this.TelephonePromoter, this.DeviceId)))
}

// 破产事件
func (this *Player) ReportBankRuptcy(gameId, gameMode, gameFreeId int32) {
	//if !this.IsRob {
	//	d, e := model.MarshalBankruptcyEvent(2, this.SnId, this.TelephonePromoter, this.Channel, this.BeUnderAgentCode, this.Platform, this.City, this.CreateTime, gameId, gameMode, gameFreeId)
	//	if e == nil {
	//		rmd := model.NewInfluxDBData("hj.player_bankruptcy", d)
	//		if rmd != nil {
	//			InfluxDBDataChannelSington.Write(rmd)
	//		}
	//	}
	//}
}

// 汇总玩家该次游戏总产生的税收
// 数据用途: 平台和推广间分账用，确保数据计算无误，
// 注意：该税收不包含俱乐部的抽水
// tax：游戏税收
func (this *Player) AddServiceFee(tax int64) {
	if this.scene == nil || this.scene.Testing || this.scene.IsMatchScene() { //测试场不统计
		return
	}
	if tax > 0 && !this.IsRob {
		this.serviceFee += tax
	}
}

func (this *Player) GetStaticsData(gameDiff string) (winCoin int64, lostCoin int64) {
	if this.PlayerData.GDatas != nil {
		if data, ok := this.PlayerData.GDatas[gameDiff]; ok {
			winCoin, lostCoin = data.Statics.TotalOut, data.Statics.TotalIn
			return
		}
	}
	return
}
func (this *Player) SaveReportForm(showId, sceneMode int, keyGameId string, profitCoin, flow int64, validBet int64) {
	//个人报表统计
	if this.TotalGameData == nil {
		this.TotalGameData = make(map[int][]*model.PlayerGameTotal)
	}
	if this.TotalGameData[showId] == nil {
		this.TotalGameData[showId] = []*model.PlayerGameTotal{new(model.PlayerGameTotal)}
	}
	td := this.TotalGameData[showId][len(this.TotalGameData[showId])-1]
	td.ProfitCoin += profitCoin
	td.BetCoin += validBet
	td.FlowCoin += flow
	///////////////最多盈利
	if pgs, exist := this.GDatas[keyGameId]; exist {
		if pgs.Statics.MaxSysOut < profitCoin {
			pgs.Statics.MaxSysOut = profitCoin
		}
	} else {
		this.GDatas[keyGameId] = &model.PlayerGameInfo{FirstTime: time.Now(), Statics: model.PlayerGameStatics{MaxSysOut: profitCoin}}
	}
}

// 个人投入产出汇总，以游戏id为key存储
// 数据用途：计算玩家赔率用，数据确保计算无误，否则可能影响玩家手牌的调控
// key: 游戏ID对应的字符串，牛牛目前用的是同一个ID，这块有待优化
// gain：输赢额，注意如果是[正值]这里一定要用税前数据，否则玩家会有数值调控优势
// 如果需要汇总gameid today的数据，可以使用game scene的GetTotalTodayDaliyGameData
func (this *Player) Statics(keyGameId string, keyGameFreeId string, gain int64, isAddNum bool) {
	if this.scene == nil || this.scene.Testing { //测试场|自建房和机器人不统计
		return
	}

	if this.IsRob && !this.scene.IsRobFightGame() {
		return
	}

	if this.TodayGameData == nil {
		this.TodayGameData = &model.PlayerGameCtrlData{}
	}
	if this.TodayGameData.CtrlData == nil {
		this.TodayGameData.CtrlData = make(map[string]*model.PlayerGameStatics)
	}

	var totalIn int64
	var totalOut int64
	if gain > 0 {
		totalOut = gain
	} else {
		totalIn = -gain
	}

	statics := make([]*model.PlayerGameStatics, 0, 4)
	//当天数据统计
	//按场次分
	if data, ok := this.TodayGameData.CtrlData[keyGameFreeId]; ok {
		statics = append(statics, data)
	} else {
		gs := &model.PlayerGameStatics{}
		this.TodayGameData.CtrlData[keyGameFreeId] = gs
		statics = append(statics, gs)
	}
	//按游戏分
	if data, ok := this.TodayGameData.CtrlData[keyGameId]; ok {
		statics = append(statics, data)
	} else {
		data = &model.PlayerGameStatics{}
		this.TodayGameData.CtrlData[keyGameId] = data
		statics = append(statics, data)
	}

	//按游戏场次进行的统计
	if data, ok := this.GDatas[keyGameFreeId]; ok {
		statics = append(statics, &data.Statics)
	} else {
		data = &model.PlayerGameInfo{FirstTime: time.Now(), Statics: model.PlayerGameStatics{}}
		this.GDatas[keyGameFreeId] = data
		statics = append(statics, &data.Statics)
	}
	if data, ok := this.GDatas[keyGameId]; ok {
		statics = append(statics, &data.Statics)
	} else {
		data = &model.PlayerGameInfo{FirstTime: time.Now(), Statics: model.PlayerGameStatics{}}
		this.GDatas[keyGameId] = data
		statics = append(statics, &data.Statics)
	}

	//if !this.scene.IsPrivateScene() {
	//	//增加黑白名单、GM过滤，因为黑白名单过后，会导致玩家体验急剧变化
	//	needStatic := this.WhiteLevel == 0 && this.WhiteFlag == 0 && this.BlackLevel == 0 && this.GMLevel == 0
	//	//增加黑白名单过滤，因为黑白名单后，会导致数据出现补偿
	//	if needStatic {
	for _, data := range statics {
		if data != nil {
			data.TotalIn += totalIn
			data.TotalOut += totalOut
			if isAddNum {
				data.GameTimes++
				if gain > 0 {
					data.WinGameTimes++
				} else if gain < 0 {
					data.LoseGameTimes++
				} else {
					data.DrawGameTimes++
				}
			}
		}
	}
	//玩家身上元数据
	this.GameTimes++
	if gain > 0 {
		this.winTimes++
		this.WinTimes++
	} else if gain < 0 {
		this.lostTimes++
		this.FailTimes++
	} else {
		this.DrawTimes++
	}
	//	}
	//}
}

func (this *Player) CheckType(gamefreeId, gameId int32) *server.DB_PlayerType {
	types := srvdata.PlayerTypeMgrSington.GetPlayerType(gamefreeId)
	cnt := len(types)
	if cnt > 0 {
		var pgs *model.PlayerGameStatics
		if this.GDatas != nil {
			if d, exist := this.GDatas[strconv.Itoa(int(gameId))]; exist {
				pgs = &d.Statics
			}
		}

		//赔率 产出/投入 万分比
		odds := int64(float64(float64(pgs.TotalOut+1)/float64(pgs.TotalIn+1)) * 10000)
		if odds > 10000000 {
			odds = 10000000
		}
		for i := 0; i < cnt; i++ {
			t := types[i]
			if t != nil {
				if this.CoinPayTotal >= int64(t.GetPayLowerLimit()) && this.CoinPayTotal <= int64(t.GetPayUpperLimit()) &&
					pgs.GameTimes >= int64(t.GetGameTimeLowerLimit()) && pgs.GameTimes <= int64(t.GetGameTimeUpperLimit()) &&
					pgs.TotalIn >= int64(t.GetTotalInLowerLimit()) && pgs.TotalIn <= int64(t.GetTotalInUpperLimit()) &&
					odds >= int64(t.GetOddsLowerLimit()) && odds <= int64(t.GetOddsUpperLimit()) {
					return t
				}
			}
		}
	}
	return nil
}

// 计算玩家赔率 产出/投入
func (this *Player) LoseRate(gamefreeId, gameId int32) (rate float64) {
	rate = -1
	if this.GDatas != nil {
		if d, exist := this.GDatas[strconv.Itoa(int(gameId))]; exist {
			rate = float64(float64(d.Statics.TotalOut+1) / float64(d.Statics.TotalIn+1))
			return rate
		}
	}

	return rate
}

// 计算玩家赔率 产出/投入
func (this *Player) LoseRateKeyGameid(gameKeyId string) (rate float64) {
	rate = -1
	var pgs *model.PlayerGameStatics
	if this.GDatas != nil {
		if d, exist := this.GDatas[gameKeyId]; exist {
			pgs = &d.Statics
		}
	}

	rate = float64(float64(pgs.TotalOut+1) / float64(pgs.TotalIn+1))
	return
}

// 是否是新手判定
func (this *Player) IsFoolPlayerBy(gameId string) {
	if this.IsRob || this.GDatas == nil {
		return
	}
	if this.GDatas[gameId] == nil {
		return
	}
	if this.IsFoolPlayer == nil {
		this.IsFoolPlayer = make(map[string]bool)
	}
	if model.GameParamData.BirdPlayerFlag == false {
		this.IsFoolPlayer[gameId] = false
		return
	}
	playerDate := this.GDatas[gameId]
	//金花游戏局数小于10局并且总产出<100000并且总产出/（总投入+10000）<=2的玩家定义为新手玩家
	if playerDate.Statics.GameTimes < 10 && playerDate.Statics.TotalOut < 100000 &&
		playerDate.Statics.TotalOut/(playerDate.Statics.TotalIn+10000) <= 2 {
		this.IsFoolPlayer[gameId] = true
	} else {
		this.IsFoolPlayer[gameId] = false
	}
}
func (this *Player) BirdPlayerCheck(gameDiff string) bool {
	if model.GameParamData.BirdPlayerFlag == false {
		return false
	}
	if this.IsRob {
		return false
	}
	if this.GDatas == nil {
		return true
	}
	gameId := gameDiff
	if this.GDatas[gameId] == nil {
		return true
	}
	playerDate := this.GDatas[gameId]
	if playerDate.Statics.GameTimes < 10 && playerDate.Statics.TotalOut < 100000 &&
		playerDate.Statics.TotalOut/(playerDate.Statics.TotalIn+10000) <= 2 {
		return true
	} else {
		return false
	}
}

func (this *Player) PlayerGameNewCheck(gameDiff string) bool {
	if this.IsRob {
		return false
	}
	if this.GDatas == nil {
		return true
	}
	gameId := gameDiff
	if this.GDatas[gameId] == nil {
		return true
	}
	playerDate := this.GDatas[gameId]
	if playerDate.Statics.GameTimes > int64(model.GameParamData.GamePlayerCheckNum) {
		return true
	} else {
		return false
	}
}

func (this *Player) SendTrusteeshipTips() {
	pack := &player.SCTrusteeshipTips{
		Trusteeship: proto.Int32(this.Trusteeship),
		TotalNum:    proto.Int32(model.GameParamData.PlayerWatchNum),
	}
	proto.SetDefaults(pack)
	logger.Logger.Trace("SCTrusteeshipTips: ", pack)
	this.SendToClient(int(player.PlayerPacketID_PACKET_SC_TRUSTEESHIPTIPS), pack)
}

func (this *Player) MarshalIParam() []*server.PlayerIParam {
	var params []*server.PlayerIParam
	for i, v := range this.Iparams {
		params = append(params, &server.PlayerIParam{
			ParamId: proto.Int(i),
			IntVal:  proto.Int64(v),
		})
	}
	return params
}

func (this *Player) UnmarshalIParam(params []*server.PlayerIParam) {
	for _, p := range params {
		this.Iparams[int(p.GetParamId())] = p.GetIntVal()
	}
}

func (this *Player) MarshalSParam() []*server.PlayerSParam {
	var params []*server.PlayerSParam
	for i, v := range this.sparams {
		params = append(params, &server.PlayerSParam{
			ParamId: proto.Int(i),
			StrVal:  proto.String(v),
		})
	}
	return params
}

func (this *Player) UnmarshalSParam(params []*server.PlayerSParam) {
	for _, p := range params {
		this.sparams[int(p.GetParamId())] = p.GetStrVal()
	}
}

func (this *Player) MarshalCParam() []*server.PlayerCParam {
	var params []*server.PlayerCParam
	for k, v := range this.cparams {
		params = append(params, &server.PlayerCParam{
			StrKey: proto.String(k),
			StrVal: proto.String(v),
		})
	}
	return params
}

func (this *Player) UnmarshalCParam(params []*server.PlayerCParam) {
	for _, p := range params {
		this.cparams[p.GetStrKey()] = p.GetStrVal()
	}
	logger.Logger.Trace("(this *Player) UnmarshalCParam ", this.cparams)
}

func (this *Player) GetIParam(k int) int64 {
	if v, exist := this.Iparams[k]; exist {
		return v
	}
	return 0
}

func (this *Player) SetIParam(k int, v int64) {
	this.Iparams[k] = v
}

func (this *Player) GetSParam(k int) string {
	if v, exist := this.sparams[k]; exist {
		return v
	}
	return ""
}

func (this *Player) SetSParam(k int, v string) {
	this.sparams[k] = v
}

// ////////////////////////////////////////////////
// 内存落地
type PlayerPO struct {
	model.PlayerData
	ExtraData  []byte
	Sid        int64
	Longitude  int32
	Latitude   int32
	City       string
	Flag       int
	Pos        int
	Dirty      bool
	Billed     bool
	AgentCode  string
	Coin       int64
	ServiceFee int64
	DisbandGen int
	GameTimes  int32
}

// 序列化
func (this *Player) Marshal() ([]byte, error) {
	po := PlayerPO{
		PlayerData: this.PlayerData,
		Sid:        this.sid,
		Longitude:  this.Longitude,
		Latitude:   this.Latitude,
		City:       this.city,
		Flag:       this.flag,
		Pos:        this.Pos,
		Dirty:      this.dirty,
		Billed:     this.Billed,
		AgentCode:  this.AgentCode,
		Coin:       this.Coin,
		ServiceFee: this.serviceFee,
		DisbandGen: this.disbandGen,
		GameTimes:  this.GameTimes,
	}
	if s, ok := this.ExtraData.(common.Serializable); ok {
		data, err := s.Marshal()
		if err != nil {
			logger.Logger.Warnf("(this *Player) Marshal() %v this.extraData.Marshal err:%v", this.SnId, err)
		}
		po.ExtraData = data
	}
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(&po)
	if err != nil {
		logger.Logger.Warnf("(this *Player) Marshal() %v gob.Encode err:%v", this.SnId, err)
		return nil, err
	}
	return buf.Bytes(), nil
}

// 反序列化
func (this *Player) Unmarshal(data []byte, ud interface{}) error {
	po := &PlayerPO{}
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(po)
	if err != nil {
		logger.Logger.Warnf("(this *Player) Unmarshal gob.Decode err:%v", err)
		return err
	}
	if scene, ok := ud.(*Scene); ok && scene != nil {
		this.PlayerData = po.PlayerData
		this.sid = po.Sid
		this.Longitude = po.Longitude
		this.Latitude = po.Latitude
		this.city = po.City
		this.flag = po.Flag
		this.Pos = po.Pos
		this.dirty = po.Dirty
		this.Billed = po.Billed
		this.AgentCode = po.AgentCode
		this.Coin = po.Coin
		this.serviceFee = po.ServiceFee
		this.disbandGen = po.DisbandGen
		this.GameTimes = po.GameTimes
		this.scene = scene
		this.UnmarkFlag(PlayerState_Online)
		if scene.sp != nil {
			this.ExtraData = scene.sp.CreatePlayerExData(scene, this)
			if s, ok := this.ExtraData.(common.Serializable); ok {
				err = s.Unmarshal(po.ExtraData, this)
				if err != nil {
					logger.Logger.Warnf("(this *Player) Unmarshal() %v this.extraData.Unmarshal err:%v", this.SnId, err)
					return err
				}
			}
		}
	}
	return nil
}
func (this *Player) GetCoin() int64 {
	return this.Coin
}
func (this *Player) SetCoin(coin int64) {
	this.Coin = coin
}

//	func (this *Player) GetStartCoin() int64 {
//		return this.StartCoin
//	}
//
//	func (this *Player) SetStartCoin(startCoin int64) {
//		this.StartCoin = startCoin
//	}
func (this *Player) GetExtraData() interface{} {
	return this.ExtraData
}
func (this *Player) SetExtraData(data interface{}) {
	this.ExtraData = data
}
func (this *Player) GetLastOPTimer() time.Time {
	return this.LastOPTimer
}
func (this *Player) SetLastOPTimer(lastOPTimer time.Time) {
	this.LastOPTimer = lastOPTimer
}
func (this *Player) GetPos() int {
	return this.Pos
}
func (this *Player) SetPos(pos int) {
	this.Pos = pos
}
func (this *Player) GetGameTimes() int32 {
	return this.GameTimes
}
func (this *Player) SetGameTimes(gameTimes int32) {
	this.GameTimes = gameTimes
}
func (this *Player) GetWinTimes() int {
	return this.winTimes
}
func (this *Player) SetWinTimes(winTimes int) {
	this.winTimes = winTimes
}
func (this *Player) GetLostTimes() int {
	return this.lostTimes
}
func (this *Player) SetLostTimes(lostTimes int) {
	this.lostTimes = lostTimes
}
func (this *Player) GetTotalBet() int64 {
	return this.TotalBet
}
func (this *Player) SetTotalBet(totalBet int64) {
	this.TotalBet = totalBet
}
func (this *Player) GetCurrentBet() int64 {
	return this.CurrentBet
}
func (this *Player) SetCurrentBet(currentBet int64) {
	this.CurrentBet = currentBet
}
func (this *Player) GetCurrentTax() int64 {
	return this.CurrentTax
}
func (this *Player) SetCurrentTax(currentTax int64) {
	this.CurrentTax = currentTax
}
func (this *Player) GetSid() int64 {
	return this.sid
}
func (this *Player) SetSid(sid int64) {
	this.sid = sid
}
func (this *Player) GetScene() *Scene {
	return this.scene
}
func (this *Player) GetGateSess() *netlib.Session {
	return this.gateSess
}
func (this *Player) SetGateSess(gateSess *netlib.Session) {
	this.gateSess = gateSess
}
func (this *Player) GetWorldSess() *netlib.Session {
	return this.worldSess
}
func (this *Player) SetWorldSess(worldSess *netlib.Session) {
	this.worldSess = worldSess
}
func (this *Player) GetCurrentCoin() int64 {
	return this.currentCoin
}
func (this *Player) SetCurrentCoin(currentCoin int64) {
	this.currentCoin = currentCoin
}
func (this *Player) GetTakeCoin() int64 {
	return this.takeCoin
}
func (this *Player) SetTakeCoin(takeCoin int64) {
	this.takeCoin = takeCoin
}
func (this *Player) GetFlag() int {
	return this.flag
}
func (this *Player) SetFlag(flag int) {
	this.flag = flag
}
func (this *Player) GetCity() string {
	return this.city
}
func (this *Player) SetCity(city string) {
	this.city = city
}

// 附加ai
func (this *Player) AttachAI(ai AI) {
	this.ai = ai
	ai.SetOwner(this)
	ai.OnStart()
}

// 解除ai
func (this *Player) UnattachAI() {
	ai := this.ai
	this.ai = nil
	ai.SetOwner(nil)
	ai.OnStop()
}

func (this *Player) RobotRandName() {
	if this.IsRob {
		if rand.Int31n(100) < 60 {
			pool := srvdata.PBDB_NameMgr.Datas.GetArr()
			cnt := int32(len(pool))
			if cnt > 0 {
				this.Name = pool[rand.Int31n(cnt)].GetName()
			}
		} else {
			this.Name = "贵宾"
		}
	}
	return
}

func (this *Player) RobRandVipWhenEnterRoom(takeCoin int64) {
	if this.IsRob {
		this.VIP = rand.Int31n(7)
		this.RobRandVip()
	}
}

func (this *Player) RobRandVip() {
	if this.IsRob {
		dbvip := srvdata.PBDB_VIPMgr.GetData(this.VIP)
		if dbvip != nil {
			outlines := dbvip.GetRewardOutlineID()
			n := len(outlines)
			this.HeadOutLine = outlines[rand.Intn(n)]
			logger.Logger.Tracef("(this *Player) RobRandVip() %d HeadOutLine=%d", this.SnId, this.HeadOutLine)
			this.dirty = true
		}
		this.Head = rand.Int31n(6) + 1
		//0:男 1:女
		this.Sex = (this.Head%2 + 1) % 2
	}
}

//////////////////////////////////////////////////
