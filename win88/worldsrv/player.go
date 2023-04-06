package main

import (
	"bytes"
	"crypto/md5"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"math/rand"
	"sort"
	"strconv"
	"time"

	"games.yol.com/win88/common"
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/bag"
	hall_proto "games.yol.com/win88/protocol/gamehall"
	login_proto "games.yol.com/win88/protocol/login"
	msg_proto "games.yol.com/win88/protocol/message"
	player_proto "games.yol.com/win88/protocol/player"
	server_proto "games.yol.com/win88/protocol/server"
	shop_proto "games.yol.com/win88/protocol/shop"
	"games.yol.com/win88/srvdata"
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/core/task"
	"github.com/idealeak/goserver/srvlib"
	srvlibproto "github.com/idealeak/goserver/srvlib/protocol"
)

// 对应到客户端的一个玩家对象.

const (
	PlayerState_Online int = iota
	PlayerState_Offline
)

const (
	UpdateField_Coin int64 = 1 << iota
	UpdateField_SafeBoxCoin
	UpdateField_VIP
	UpdateField_CoinPayTotal
	UpdateField_TotalConvertibleFlow
	UpdateField_Ticket
	UpdateField_Grade
	UpdateField_Diamond
	UpdateField_VCoin
)

const (
	CurrencyType_Card int32 = iota
	CurrencyType_Coin
	CurrencyType_Money
	CurrencyType_RMB
)

var VIP_RandWeight = []int64{30, 30, 20, 10, 5, 5}

type ErrorString struct {
	code string
}

func (this *ErrorString) Error() string {
	return this.code
}

type Player struct {
	*model.PlayerData                                   //po 持久化对象
	diffData                  model.PlayerDiffData      //差异数据
	gateSess                  *netlib.Session           //所在GateServer的session
	scene                     *Scene                    //当前所在个Scene
	thrscene                  *Scene                    //当前所三方Scene
	sid                       int64                     //对应客户端的sessionId
	state                     int                       //玩家状态 PlayerState_Online|PlayerState_Offline
	lastSaved                 time.Time                 //最后保存的时间
	dirty                     bool                      //脏标记
	pos                       int                       //位置
	msgs                      map[string]*model.Message //
	beUnderAgent              bool                      //是否隶属于代理
	ge                        GameEvent                 //游戏事件
	lastSceneId               map[int32][]int32         //上一个房间id
	lastOllen                 time.Time                 //在线时长统计辅助
	ollenSecs                 int32                     //在线时长
	currClubId                int32                     //当前所在的俱乐部id
	takeCoin                  int64                     //携带金币
	sceneCoin                 int64                     //房间里当前的金币量
	isNewbie                  bool                      //是否是新用户
	applyPos                  int32                     //申请的场景位置
	flag                      int32                     //Game服务器的用户状态
	hallId                    int32                     //所在游戏大厅
	changeIconTime            time.Time                 //上次修改头像时间
	enterts                   time.Time                 //进入时间
	lastChangeScene           time.Time                 //上次换桌时间
	isAudience                bool                      //是否是观众
	customerToken             string                    //客服会话token
	isDelete                  bool                      //是否已删档用户
	thridBalanceRefreshReqing bool                      //玩家请求刷新三方金额进行中
	thridBalanceReqIsSucces   bool                      //三方请求刷新成功
	params                    *model.PlayerParams       //客户端登陆游戏带上来的临时参数
	exchangeState             bool                      //订单操作状态
	lastOnDayChange           time.Time                 //上次ondaychange
	lastOnWeekChange          time.Time
	lastOnMonthChange         time.Time
	//例如["WWG平台"][true:已经刷新过不用再次刷新，false:未知需要刷新]
	//注意：该数据不会落地到mgo,游服第一次启动的时候会请求全部三方刷新余额，之后就会根据标记来进行刷新。
	//确保thirdBalanceRefreshMark只在主协程里面操作
	thirdBalanceRefreshMark map[string]bool
	EnterCoinSceneQueueTs   int64
	CoinSceneQueueRound     int32
	CoinSceneQueue          *CoinScenePool
	EnterQueueTime          time.Time //进入队列的时间
	//比赛
	cparams map[string]string //平台登陆数据	"name", "head", "geo-info", "lang", "ip"
	Iparams map[int]int64     //整形参数
	sparams map[int]string    //字符参数
	//用户登录时获取 单用户活动开关控制
	layered map[int]bool //true 为关 false为开
	//用户分层ids
	layerlevels []int //缓存玩家身上所有分层数据
	WhiteLevel  int32
	BlackLevel  int32
	miniScene   map[int32]*Scene //当前所在个Scene
	leavechan   chan int
	matchCtx    *MatchContext //比赛环境
}

func NewPlayer(sid int64, pd *model.PlayerData, s *netlib.Session) *Player {
	p := &Player{
		PlayerData:              pd,
		sid:                     sid,
		gateSess:                s,
		state:                   PlayerState_Online,
		msgs:                    make(map[string]*model.Message),
		cparams:                 make(map[string]string),
		Iparams:                 make(map[int]int64),
		sparams:                 make(map[int]string),
		pos:                     -1,
		beUnderAgent:            pd.BeUnderAgentCode != "",
		thirdBalanceRefreshMark: make(map[string]bool),
		layered:                 make(map[int]bool),
		miniScene:               make(map[int32]*Scene),
	}

	if p.init() {
		return p
	}
	return nil
}

func (this *Player) init() bool {
	this.isNewbie = this.CreateTime == this.LastLoginTime
	this.applyPos = -1
	if this.WBLevel > 0 {
		this.WhiteLevel = this.WBLevel
	} else if this.WBLevel < 0 {
		this.BlackLevel = -this.WBLevel
	}
	return true
}

func (this *Player) GenCustomerToken() string {
	if this.customerToken != "" {
		return this.customerToken
	}

	raw := fmt.Sprintf("%v%v%v%v%v", this.SnId, this.AccountId, this.sid, common.GetAppId(), time.Now().UnixNano())
	h := md5.New()
	io.WriteString(h, raw)
	token := hex.EncodeToString(h.Sum(nil))
	return token
}
func (this *Player) SyncBagData(itemInfo []*bag.ItemInfo) {
	pack := &bag.SCSyncBagData{
		Infos: itemInfo,
	}
	proto.SetDefaults(pack)
	logger.Logger.Trace("SC_SYNCBAGDATA: ", pack)
	this.SendToClient(int(bag.SPacketID_PACKET_SC_SYNCBAGDATA), pack)
}
func (this *Player) SendToClient(packetid int, rawpack interface{}) bool {
	if this.gateSess == nil {
		logger.Logger.Tracef("[%v] sess == nil ", this.SnId, packetid)
		return false
	}
	if rawpack == nil {
		logger.Logger.Trace(" rawpack == nil ")
		return false
	}

	if this.state == PlayerState_Offline {
		logger.Logger.Trace("Player if offline.")
		return false
	}

	return common.SendToGate(this.sid, packetid, rawpack, this.gateSess)
}
func (this *Player) SendRawToClientIncOffLine(sid int64, gateSess *netlib.Session, packetid int, rawpack interface{}) bool {
	if gateSess == nil {
		logger.Logger.Tracef("[%v] sess == nil ", this.SnId, packetid)
		return false
	}
	if rawpack == nil {
		logger.Logger.Trace(" rawpack == nil ")
		return false
	}

	return common.SendToGate(sid, packetid, rawpack, gateSess)
}

func (this *Player) SendToGame(packetid int, rawpack interface{}) bool {
	if this.scene == nil || this.scene.gameSess == nil || this.scene.gameSess.Session == nil {
		logger.Logger.Tracef("[%v] sess == nil ", this.Name)
		return false
	}
	if rawpack == nil {
		logger.Logger.Trace(" rawpack == nil ")
		return false
	}

	return this.scene.SendToGame(packetid, rawpack)
}

func (this *Player) OnLogined() {
	logger.Logger.Tracef("(this *Player) OnLogined() %v", this.SnId)
	this.BindGroupTag([]string{this.Platform})

	this.lastSaved = time.Now()
	isFirstLogin := false
	if this.CreateTime.Unix() > this.LastLogoutTime.Unix() {
		isFirstLogin = true
	}

	//连续登录,跨天业务处理,要放在末尾处理
	tNow := time.Now().Local()
	tLastLogout := this.PlayerData.LastLogoutTime
	this.PlayerData.LastLoginTime = tNow
	this.PlayerData.LastLogoutTime = tNow
	this.lastOllen = tNow
	this.dirty = true
	//ActOnlineRewardMgrSington.CheckVersion(this)
	//ActStepRechargeMgrSington.CheckVersion(this)
	//ActYebMgrSington.CalcYebInterest(this, nil) // 计算余额宝利息
	//this.PlayerData.OnlineRewardData.Ts = tNow.Unix()
	inSameDay := common.InSameDay(tNow, tLastLogout)
	if !inSameDay { //跨天
		logger.Logger.Infof("(this *Player) OnLogined(%v) inSameDay LastLogoutTime(%v)", this.SnId, tLastLogout)
		isContineDay := common.IsContinueDay(tNow, tLastLogout)
		//计算跨了多少天
		var t int
		t1 := tLastLogout
		t2 := this.PlayerData.LastLoginTime
		out := time.Date(t1.Year(), t1.Month(), t1.Day(), 0, 0, 0, 0, t1.Location())
		in := time.Date(t2.Year(), t2.Month(), t2.Day(), 0, 0, 0, 0, t2.Location())
		t = int(math.Floor(in.Sub(out).Hours() / 24))
		this.OnDayTimer(true, isContineDay, t)
	}

	inSameMoney := common.InSameMonth(tNow, tLastLogout)
	if !inSameMoney { //跨月
		this.OnMonthTimer()
	}
	inSameWeek := common.InSameWeek(tNow, tLastLogout)
	if !inSameWeek { //跨周
		this.OnWeekTimer()
	}
	this.state = PlayerState_Online

	old := this.VIP
	this.VIP = this.GetVIPLevel(int64(this.CoinPayTotal))
	//测试用
	if !this.IsRob {
		//this.VIP = 6
	} else {
		this.VIP = rand.Int31n(6) + 1
		//机器人随机vip和头像
		this.RobRandVip()
	}
	//头像决定性别
	this.Sex = (this.Head%2 + 1) % 2
	if old != this.VIP {
		this.dirty = true
		//clubManager.UpdateVip(this)
	}
	//
	this.BackDiffData()
	//
	if !this.IsRob {

		for i := 0; i != 10; i++ {
			this.TestMail()
		}
		// this.TestSubMail()
		PlayerOnlineSington.Check = true
		this.LoadMessage(tLastLogout.Unix(), isFirstLogin, !inSameWeek)
	}

	if !this.IsRob {
		//登录次数
		this.LoginTimes++
		//登录事件
		this.ReportLoginEvent()
		//登录日志
		logState := LoginStateMgrSington.GetLoginStateOfSid(this.sid)
		var clog *model.ClientLoginInfo
		if logState != nil {
			clog = logState.clog
		}
		LogChannelSington.WriteLog(model.NewLoginLog(this.SnId, model.LoginLogType_Login, this.Tel, this.Ip,
			this.Platform, this.Channel, this.BeUnderAgentCode, this.PackageID, this.City, clog, this.GetTotalCoin(), 0))
	}
	//登录事件
	this.ge.eventType = GAMEEVENT_LOGIN
	this.ge.eventSrc = nil
	this.FireGameEvent()
	PlatformMgrSington.PlayerLogin(this)

	gameStateMgr.PlayerClear(this)
	// 用户登录保存下数据到player_logicleveldata
	if !this.IsRob {
		this.SendToRepSrv(this.PlayerData)
	}
	//用户分层数据初始化
	//this.InitLayered()

	//单控数据
	PlayerSingleAdjustMgr.LoadSingleAdjustData(this.Platform, this.SnId)

	//玩家登录事件
	FirePlayerLogined(this)

	if !this.IsRob {
		//玩家背包数据
		BagMgrSington.InitBagInfo(this, this.Platform)
		FriendMgrSington.LoadFriendData(this.Platform, this.SnId)
		FriendMgrSington.CheckSendFriendApplyData(this)
		//七日活动
		ActSignMgrSington.OnPlayerLogin(this)

		//红点检测
		this.CheckShowRed()

	} else {
		this.RandRobotExData()
	}
}
func (this *Player) CheckShowRed() {
	this.MessageShowRed()

	//商城红点
	ShopMgrSington.ShopCheckShowRed(this)

	//数据依赖于背包数据  放入加载背包数据之后执行
	//PetMgrSington.CheckShowRed(this)
}

// 为了worldsrv和gamesrv上机器人信息一致
func (this *Player) RandRobotExData() {
	if !this.IsRob {
		return
	}
	// 角色
	datas := srvdata.PBDB_Game_IntroductionMgr.Datas.GetArr()
	if datas != nil {
		//随机
		//var roles = []int32{}
		//for _, data := range datas {
		//    if data.Type == 1 { //角色
		//        roles = append(roles, data.Id)
		//    }
		//}
		//if roles != nil && len(roles) > 0 {
		//    randId := common.RandInt32Slice(roles)
		//    this.Roles = new(model.RolePetInfo)
		//    this.Roles.ModUnlock = make(map[int32]int32)
		//    this.Roles.ModUnlock[randId] = 1
		//    this.Roles.ModId = randId
		//}
		//女8男2
		rand := common.RandInt(100)
		randId := int32(2000001)
		if rand < 20 {
			randId = 2000002
		}
		this.Roles = new(model.RolePetInfo)
		this.Roles.ModUnlock = make(map[int32]int32)
		this.Roles.ModUnlock[randId] = 1
		this.Roles.ModId = randId
	}
	// 宠物
	//    datas := srvdata.PBDB_Game_IntroductionMgr.Datas.GetArr()
	//    if datas != nil {
	//        var pets = []int32{}
	//        for _, data := range datas {
	//            if data.Type == 2 { //宠物
	//                pets = append(pets, data.Id)
	//            }
	//        }
	//        if pets != nil && len(pets) > 0 {
	//            randId := common.RandInt32Slice(pets)
	//            this.Pets = new(model.RolePetInfo)
	//            this.Pets.ModUnlock = make(map[int32]int32)
	//            this.Pets.ModUnlock[randId] = 1
	//            this.Pets.ModId = randId
	//        }
	//    }
}

func (this *Player) OnRehold() {
	logger.Logger.Tracef("(this *Player) OnRehold() %v", this.SnId)
	this.BindGroupTag([]string{this.Platform})

	var gameid int
	if this.scene != nil && this.scene.gameSess != nil {
		if this.scene.sceneId == SceneMgrSington.GetDgSceneId() {
			//如果是之前进入的是DG游戏，就退出DG游戏
			this.DgGameLogout()
		} else {
			var gateSid int64
			if this.gateSess != nil {
				if srvInfo, ok := this.gateSess.GetAttribute(srvlib.SessionAttributeServerInfo).(*srvlibproto.SSSrvRegiste); ok && srvInfo != nil {
					sessionId := srvlib.NewSessionIdEx(srvInfo.GetAreaId(), srvInfo.GetType(), srvInfo.GetId(), 0)
					gateSid = sessionId.Get()
				}
			}
			pack := &server_proto.WGPlayerRehold{
				Id:      proto.Int32(this.SnId),
				Sid:     proto.Int64(this.sid),
				SceneId: proto.Int(this.scene.sceneId),
				GateSid: proto.Int64(gateSid),
			}
			proto.SetDefaults(pack)
			this.SendToGame(int(server_proto.SSPacketID_PACKET_WG_PLAYERREHOLD), pack)
		}
		gameid = this.scene.gameId
	}
	if !this.IsRob {
		PlayerOnlineSington.Check = true
	}
	this.state = PlayerState_Online
	tNow := time.Now().Local()
	inSameDay := common.InSameDay(tNow, this.LastLoginTime)
	if !inSameDay && !this.IsRob { //跨天
		//登录事件
		this.ReportLoginEvent()
	}
	logState := LoginStateMgrSington.GetLoginStateOfSid(this.sid)
	var clog *model.ClientLoginInfo
	if logState != nil {
		clog = logState.clog
	}
	//排除掉机器人
	if !this.IsRob {
		LogChannelSington.WriteLog(model.NewLoginLog(this.SnId, model.LoginLogType_Rehold, this.Tel, this.Ip,
			this.Platform, this.Channel, this.BeUnderAgentCode, this.PackageID, this.City, clog, this.GetTotalCoin(),
			gameid))
	}
	this.LastLoginTime = tNow

	//
	// 邮件统一由客户端拉取
	//  this.SendMessage()
	//	TeaHouseMgr.T2PRelationMgr.SetLoginTime(this.SnId, this.LastLoginTime.Unix())
	this.lastOllen = tNow
	PlatformMgrSington.PlayerLogin(this)

	gameStateMgr.PlayerClear(this)

	CoinSceneMgrSington.PlayerQueueState(this)

	//玩家重连事件
	FirePlayerRehold(this)

	//用户分层数据初始化
	//this.ActStateSend2Client

	FriendMgrSington.CheckSendFriendApplyData(this)
	FriendUnreadMgrSington.CheckSendFriendUnreadData(this.SnId)
	//七日活动.
	ActSignMgrSington.OnPlayerLogin(this)

	//玩家背包数据
	BagMgrSington.InitBagInfo(this, this.Platform)

	this.CheckShowRed()
}

// 玩家断线重连时，获取玩家所有游戏的配置信息
func (this *Player) SendGameConfig(gameId int32, plf, chl string) {
	pack := &hall_proto.SCGetGameConfig{}
	gps := PlatformMgrSington.GetPlatformGameConfig(this.Platform)
	for _, v := range gps {
		if v.Status && PlatformMgrSington.GameStatus[v.DbGameFree.Id] {
			if v.DbGameFree.GetGameRule() != 0 && v.DbGameFree.GetGameId() == gameId {
				lgc := &hall_proto.GameConfig1{
					LogicId:        proto.Int32(v.DbGameFree.Id),
					LimitCoin:      proto.Int32(v.DbGameFree.GetLimitCoin()),
					MaxCoinLimit:   proto.Int32(v.DbGameFree.GetMaxCoinLimit()),
					BaseScore:      proto.Int32(v.DbGameFree.GetBaseScore()),
					BetScore:       proto.Int32(v.DbGameFree.GetBetLimit()),
					OtherIntParams: v.DbGameFree.GetOtherIntParams(),
					MaxBetCoin:     v.DbGameFree.GetMaxBetCoin(),
					MatchMode:      proto.Int32(v.DbGameFree.GetMatchMode()),
					Status:         v.Status,
				}
				if v.DbGameFree.GetLottery() != 0 { //彩金池
					//lgc.LotteryCfg = v.DBGameFree.LotteryConfig
					//_, gl := LotteryMgrSington.FetchLottery(plf, v.DBGameFree.GetId(), v.DBGameFree.GetGameId())
					//if gl != nil {
					//	lgc.LotteryCoin = proto.Int64(gl.Value)
					//}
				}
				pack.GameCfg = append(pack.GameCfg, lgc)
			}
		}
	}
	this.SendToClient(int(hall_proto.GameHallPacketID_PACKET_SC_GETGAMECONFIG), pack)
}

// //////////////// 返利 ////////////////////////////
//
//	func (this *Player) ClearRebate() {
//		rebateTask := RebateInfoMgrSington.rebateTask[this.Platform]
//		if rebateTask != nil {
//			if this.LogonMarker == nil {
//				this.LogonMarker = make(map[string]int)
//			}
//			if rebateTask.Version != this.LogonMarker["Rebate"] {
//				this.LogonMarker["Rebate"] = rebateTask.Version
//				this.RebateData = make(map[string]*model.RebateData)
//				this.RebateRedIsShow(false)
//			} else {
//				this.CountAllRebate()
//			}
//		}
//	}
//
// //只是临时处理的函数，没别的用，只是清理一下版本的数据
//
//	func (this *Player) ClearRebate2() {
//		rebateTask := RebateInfoMgrSington.rebateTask[this.Platform]
//		if rebateTask != nil {
//			if this.LogonMarker == nil {
//				this.LogonMarker = make(map[string]int)
//			}
//			if rebateTask.Version != this.LogonMarker["Rebate"] {
//				this.LogonMarker["Rebate"] = rebateTask.Version
//				this.RebateData = make(map[string]*model.RebateData)
//				this.RebateRedIsShow(false)
//			}
//		}
//	}
//
// //计算当前返利
//
//	func (this *Player) CountRebate(key string, source int) {
//		if this.IsRob {
//			return
//		}
//		rebateTask := RebateInfoMgrSington.rebateTask[this.Platform]
//		if rebateTask != nil && rebateTask.RebateSwitch && this.RebateData != nil && (rebateTask.RebateManState == 0 ||
//			rebateTask.RebateManState == 1 && this.IsCanRebate == 1) {
//			nowrbd := this.RebateData[key]
//			if nowrbd == nil {
//				return
//			}
//			if source == 0 {
//				rgc := rebateTask.RebateGameCfg[key]
//				if rgc != nil {
//					var total int64 = 0
//					vbt := nowrbd.ValidBetTotal
//					a, b, c := int64(rgc.BaseCoin[0]), int64(rgc.BaseCoin[1]), int64(rgc.BaseCoin[2])
//					n := -1
//					if a > 0 && vbt >= a && (b == 0 || vbt < b) {
//						n = 0
//					} else if b > 0 && vbt >= b && (c == 0 || vbt < c) {
//						n = 1
//					} else if c > 0 && vbt >= c && b > 0 {
//						n = 2
//					}
//					if n != -1 {
//						coin := math.Floor(float64(vbt*int64(rgc.RebateRate[n])) / 10000)
//						if coin > 0 {
//							total += int64(coin)
//						}
//					}
//					if rgc.MaxRebateCoin != 0 {
//						if rebateTask.ReceiveMode == 0 {
//							if total+nowrbd.TotalRebateCoin > rgc.MaxRebateCoin {
//								total = rgc.MaxRebateCoin - nowrbd.TotalRebateCoin
//							}
//						} else {
//							if total > rgc.MaxRebateCoin {
//								total = rgc.MaxRebateCoin
//							}
//						}
//					}
//					if total < 0 {
//						total = 0
//					}
//					nowrbd.TodayRebateCoin = total
//				}
//			} else {
//				rgc := rebateTask.RebateGameThirdCfg[key]
//				if rgc != nil {
//					var total int64 = 0
//					vbt := nowrbd.ValidBetTotal
//					a, b, c := int64(rgc.BaseCoin[0]), int64(rgc.BaseCoin[1]), int64(rgc.BaseCoin[2])
//					n := -1
//					if a > 0 && vbt >= a && (b == 0 || vbt < b) {
//						n = 0
//					} else if b > 0 && vbt >= b && (c == 0 || vbt < c) {
//						n = 1
//					} else if c > 0 && vbt > c && b > 0 {
//						n = 2
//					}
//					if n != -1 {
//						coin := math.Floor(float64(vbt*int64(rgc.RebateRate[n])) / 10000)
//						if coin > 0 {
//							total += int64(coin)
//						}
//					}
//					if rgc.MaxRebateCoin != 0 {
//						if rebateTask.ReceiveMode == 0 {
//							if total+nowrbd.TotalRebateCoin > rgc.MaxRebateCoin {
//								total = rgc.MaxRebateCoin - nowrbd.TotalRebateCoin
//							}
//						} else {
//							if total > rgc.MaxRebateCoin {
//								total = rgc.MaxRebateCoin
//							}
//						}
//					}
//					if total < 0 {
//						total = 0
//					}
//					nowrbd.TodayRebateCoin = total
//				}
//			}
//			this.dirty = true
//		}
//		this.RebateRedIsShow(true)
//	}
//
// //重新计算所有返利值 防止后台修改后数据不同步
//
//	func (this *Player) CountAllRebate() {
//		rebateTask := RebateInfoMgrSington.rebateTask[this.Platform]
//		if rebateTask != nil && rebateTask.RebateSwitch && this.RebateData != nil && (rebateTask.RebateManState == 0 ||
//			rebateTask.RebateManState == 1 && this.IsCanRebate == 1) {
//			for k, rgc := range rebateTask.RebateGameCfg {
//				nowrbd := this.RebateData[k]
//				if nowrbd != nil {
//					vbt := nowrbd.ValidBetTotal
//					var total int64 = 0
//					a, b, c := int64(rgc.BaseCoin[0]), int64(rgc.BaseCoin[1]), int64(rgc.BaseCoin[2])
//					n := -1
//					if a > 0 && vbt >= a && (b == 0 || vbt < b) {
//						n = 0
//					} else if b > 0 && vbt >= b && (c == 0 || vbt < c) {
//						n = 1
//					} else if c > 0 && vbt >= c && b > 0 {
//						n = 2
//					}
//					if n != -1 {
//						coin := math.Floor(float64(vbt*int64(rgc.RebateRate[n])) / 10000)
//						if coin > 0 {
//							total += int64(coin)
//						}
//					}
//					if rgc.MaxRebateCoin != 0 {
//						if rebateTask.ReceiveMode == 0 {
//							if total+nowrbd.TotalRebateCoin > rgc.MaxRebateCoin {
//								total = rgc.MaxRebateCoin - nowrbd.TotalRebateCoin
//							}
//						} else {
//							if total > rgc.MaxRebateCoin {
//								total = rgc.MaxRebateCoin
//							}
//						}
//					}
//					if total < 0 {
//						total = 0
//					}
//					nowrbd.TodayRebateCoin = total
//				}
//			}
//			for k, rgc := range rebateTask.RebateGameThirdCfg {
//				nowrbd := this.RebateData[k]
//				if nowrbd != nil {
//					vbt := nowrbd.ValidBetTotal
//					var total int64 = 0
//					a, b, c := int64(rgc.BaseCoin[0]), int64(rgc.BaseCoin[1]), int64(rgc.BaseCoin[2])
//					n := -1
//					if a > 0 && vbt >= a && (b == 0 || vbt < b) {
//						n = 0
//					} else if b > 0 && vbt >= b && (c == 0 || vbt < c) {
//						n = 1
//					} else if c > 0 && vbt > c && b > 0 {
//						n = 2
//					}
//					if n != -1 {
//						coin := math.Floor(float64(vbt*int64(rgc.RebateRate[n])) / 10000)
//						if coin > 0 {
//							total += int64(coin)
//						}
//					}
//					if rgc.MaxRebateCoin != 0 {
//						if rebateTask.ReceiveMode == 0 {
//							if total+nowrbd.TotalRebateCoin > rgc.MaxRebateCoin {
//								total = rgc.MaxRebateCoin - nowrbd.TotalRebateCoin
//							}
//						} else {
//							if total > rgc.MaxRebateCoin {
//								total = rgc.MaxRebateCoin
//							}
//						}
//					}
//					if total < 0 {
//						total = 0
//					}
//					nowrbd.TodayRebateCoin = total
//				}
//			}
//		}
//		this.RebateRedIsShow(true)
//	}
//
//	func (this *Player) RebateInfoUpdate(continuous bool) {
//		this.ClearRebate()
//		var total int64 = 0
//		rebateTask := RebateInfoMgrSington.rebateTask[this.Platform]
//		if rebateTask != nil && rebateTask.RebateSwitch && (rebateTask.RebateManState == 0 ||
//			(rebateTask.RebateManState == 1 && this.IsCanRebate == 1)) {
//			if rebateTask.ReceiveMode == 0 {
//				if rebateTask.NotGiveOverdue == 0 {
//					for _, v := range this.RebateData {
//						v.TotalHaveRebateCoin += v.TodayRebateCoin
//						v.TodayRebateCoin = 0
//					}
//				} else if rebateTask.NotGiveOverdue == 1 {
//					this.RebateData = make(map[string]*model.RebateData)
//				} else if rebateTask.NotGiveOverdue == 2 {
//					for _, v := range this.RebateData {
//						total += v.TodayRebateCoin
//						v.TodayRebateCoin = 0
//					}
//				}
//			} else if rebateTask.ReceiveMode == 1 {
//				if continuous {
//					if rebateTask.NotGiveOverdue == 0 {
//						for _, v := range this.RebateData {
//							v.TotalHaveRebateCoin += v.YesterdayRebateCoin
//							v.YesterdayRebateCoin = v.TodayRebateCoin
//							v.TotalHaveValidBetTotal += v.YesterdayValidBetTotal
//							v.YesterdayValidBetTotal = v.ValidBetTotal
//							v.TodayRebateCoin = 0
//						}
//					} else if rebateTask.NotGiveOverdue == 1 {
//						for _, v := range this.RebateData {
//							v.YesterdayRebateCoin = v.TodayRebateCoin
//							v.YesterdayValidBetTotal = v.ValidBetTotal
//							v.TodayRebateCoin = 0
//						}
//					} else if rebateTask.NotGiveOverdue == 2 {
//						for _, v := range this.RebateData {
//							total += v.YesterdayRebateCoin
//							v.YesterdayRebateCoin = v.TodayRebateCoin
//							v.YesterdayValidBetTotal = v.ValidBetTotal
//							v.TodayRebateCoin = 0
//						}
//					}
//				} else {
//					if rebateTask.NotGiveOverdue == 0 {
//						for _, v := range this.RebateData {
//							v.TotalHaveRebateCoin += v.YesterdayRebateCoin + v.TodayRebateCoin
//							v.TotalHaveValidBetTotal += v.YesterdayValidBetTotal + v.ValidBetTotal
//							v.YesterdayRebateCoin = 0
//							v.YesterdayValidBetTotal = 0
//							v.TodayRebateCoin = 0
//						}
//					} else if rebateTask.NotGiveOverdue == 1 {
//						for _, v := range this.RebateData {
//							v.YesterdayRebateCoin = 0
//							v.YesterdayValidBetTotal = 0
//							v.TodayRebateCoin = 0
//						}
//					} else if rebateTask.NotGiveOverdue == 2 {
//						for _, v := range this.RebateData {
//							total += v.YesterdayRebateCoin + v.TodayRebateCoin
//							v.YesterdayValidBetTotal = 0
//							v.YesterdayRebateCoin = 0
//							v.TodayRebateCoin = 0
//						}
//					}
//				}
//			}
//
//			for _, v := range this.RebateData {
//				if rebateTask.NotGiveOverdue != 0 {
//					v.TotalHaveRebateCoin = 0
//				}
//				v.ValidBetTotal = 0
//				v.TotalRebateCoin = 0
//			}
//		} else {
//			this.RebateData = make(map[string]*model.RebateData)
//		}
//
//		if rebateTask != nil && rebateTask.NotGiveOverdue != 0 && total > 0 {
//			//过期邮件给
//			var newMsg *model.Message
//			task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
//				a := strconv.FormatFloat(float64(total)/100, 'f', 2, 64)
//				str := fmt.Sprintf("昨日您在游戏返利中，共获得 %v，请查收", a)
//				newMsg = model.NewMessage("", 0, this.SnId, model.MSGTYPE_REBATE, "系统通知", str,
//					total, 0, time.Now().Unix(), 0, "", nil, this.Platform)
//				return model.InsertMessage(newMsg)
//			}), task.CompleteNotifyWrapper(func(data interface{}, t *task.Task) {
//				if data == nil {
//					this.AddMessage(newMsg)
//				}
//			}), "SendMessage").Start()
//		}
//		this.ClearRebate()
//		SendRebateList(this)
//	}
//
//	func (this *Player) RebateRedIsShow(isF bool) {
//		pack := &hall_proto.SCRedCtrl{
//			OpCode:  proto.Int64(0),
//			IsFShow: proto.Bool(false),
//		}
//		if this.RebateData != nil && isF {
//			rebateTask := RebateInfoMgrSington.rebateTask[this.Platform]
//			if rebateTask != nil {
//				for k, _ := range rebateTask.RebateGameCfg {
//					rbd := this.RebateData[k]
//					if rbd != nil {
//						if rebateTask.ReceiveMode == 0 && rbd.TodayRebateCoin+rbd.TotalHaveRebateCoin > 0 {
//							pack.IsFShow = proto.Bool(true)
//							break
//						} else if rebateTask.ReceiveMode == 1 && rbd.YesterdayRebateCoin+rbd.TotalHaveRebateCoin > 0 {
//							pack.IsFShow = proto.Bool(true)
//							break
//						}
//					}
//				}
//				for k, _ := range rebateTask.RebateGameThirdCfg {
//					rbd := this.RebateData[k]
//					if rbd != nil {
//						if rebateTask.ReceiveMode == 0 && rbd.TodayRebateCoin+rbd.TotalHaveRebateCoin > 0 {
//							pack.IsFShow = proto.Bool(true)
//							break
//						} else if rebateTask.ReceiveMode == 1 && rbd.YesterdayRebateCoin+rbd.TotalHaveRebateCoin > 0 {
//							pack.IsFShow = proto.Bool(true)
//							break
//						}
//					}
//				}
//
//			}
//		}
//		proto.SetDefaults(pack)
//		logger.Logger.Trace("SCRedCtrl: ", pack)
//		this.SendToClient(int(hall_proto.HallPacketID_PACKET_SC_HALL_REDCTRL), pack)
//	}
func (this *Player) LoadMessage(lastLogoutTime int64, isFirstLogin, isSkipWeek bool) {
	task.New(nil,
		task.CallableWrapper(func(o *basic.Object) interface{} {
			//msgs, err := model.GetMessageByNotState(this.SnId, model.MSGSTATE_REMOVEED)
			//if err == nil {
			//	return msgs
			//} else {
			//	logger.Logger.Warnf("[%v] LoadMessage err:%v", this.Name, err)
			//}
			msgs, err := model.GetMessage(this.Platform, this.SnId) // model.GetNotDelMessage(this.Platform, this.SnId)
			if err == nil {
				return msgs
			} else {
				logger.Logger.Warnf("[%v] LoadMessage err:%v", this.Name, err)
			}
			return nil
		}),
		task.CompleteNotifyWrapper(func(data interface{}, tt task.Task) {
			if data != nil {
				if msgs, ok := data.([]model.Message); ok {
					for i := 0; i < len(msgs); i++ {
						this.msgs[msgs[i].Id.Hex()] = &msgs[i]
					}
				}
				//跨周 邮件>0
				if isSkipWeek && len(this.msgs) > 0 {
					//过期邮件处理
					for key, msg := range this.msgs {
						if msg.AttachState == model.MSGATTACHSTATE_DEFAULT && msg.Ticket > 0 {
							this.TicketTotalDel += msg.Ticket
							this.DelMessage(key, 1)
						}
					}
					this.dirty = true
				}

				var dbMsgs []*model.Message

				//第一次登录需要屏蔽订阅的邮件
				if !isFirstLogin {
					msgs := MsgMgrSington.GetSubscribeMsgs(this.Platform, lastLogoutTime)
					for _, msg := range msgs {
						bHasAddToPlayer := false
						for _, pMsg := range this.msgs {
							if pMsg.Pid == msg.Id.Hex() {
								bHasAddToPlayer = true
								break
							}
						}
						if bHasAddToPlayer == false {
							newMsg := model.NewMessage(msg.Id.Hex(), msg.SrcId, "系统", this.SnId, msg.MType, msg.Title,
								msg.Content, msg.Coin, msg.Diamond, model.MSGSTATE_UNREAD, msg.CreatTs, msg.AttachState,
								msg.GiftId, msg.Params, msg.Platform, msg.ShowId)
							dbMsgs = append(dbMsgs, newMsg)
						}
					}
				}

				if len(dbMsgs) != 0 {
					task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
						return model.InsertMessage(this.Platform, dbMsgs...)
					}), task.CompleteNotifyWrapper(func(data interface{}, tt task.Task) {
						if data == nil {
							for _, msg := range dbMsgs {
								bHasAddToPlayer := false
								for _, pMsg := range this.msgs {
									if pMsg.Pid == msg.Id.Hex() {
										bHasAddToPlayer = true
										break
									}
								}
								if bHasAddToPlayer == false {
									this.AddMessage(msg)
								}
							}
							// 邮件由客户端拉取
							//this.SendMessage()
						}
					}), "InsertMessage").StartByFixExecutor("logic_message")
				} else {
					// this.SendMessage()
				}
				if len(this.msgs) > model.MSG_MAX_COUNT {
					maxTs := int64(0)
					for _, m := range this.msgs {
						if m.CreatTs > maxTs {
							maxTs = m.CreatTs
						}
					}
					this.verifyMessage(maxTs)
				}
			}
		}), "GetMessage").StartByFixExecutor("logic_message")
}

func (this *Player) SendMessage(showId int64) {
	pack := &msg_proto.SCMessageList{}
	if len(this.msgs) != 0 {
		for _, msg := range this.msgs {

			if msg.State != model.MSGSTATE_REMOVEED && (msg.ShowId == model.HallAll || msg.ShowId&showId != 0) {

				giftState := int32(0)
				//if len(msg.GiftId) > 0 {
				//	gift := GiftMgrSington.GetFromRecvGift(this.SnId, msg.GiftId)
				//	if gift != nil {
				//		giftState = gift.State
				//	} else {
				//		logger.Logger.Error("player: ", this.SnId, " not find gift : ", msg.GiftId)
				//	}
				//}
				pack.Msgs = append(pack.Msgs, &msg_proto.MessageData{
					Id:          proto.String(msg.Id.Hex()),
					Title:       proto.String(msg.Title),
					Content:     proto.String(msg.Content),
					MType:       proto.Int32(msg.MType),
					SrcId:       proto.Int32(msg.SrcId),
					SrcName:     proto.String(msg.SrcName),
					Coin:        proto.Int64(msg.Coin),
					Diamond:     proto.Int64(msg.Diamond),
					Ticket:      proto.Int64(msg.Ticket),
					Grade:       proto.Int64(msg.Grade),
					State:       proto.Int32(msg.State),
					Ts:          proto.Int32(int32(msg.CreatTs)),
					Params:      msg.Params,
					AttachState: proto.Int32(msg.AttachState),
					GiftId:      proto.String(msg.GiftId),
					GiftState:   proto.Int32(giftState),
				})
			}
		}
		proto.SetDefaults(pack)
	}
	//nil的msg需要发给前端便于覆盖前一个玩家的信息
	this.SendToClient(int(msg_proto.MSGPacketID_PACKET_SC_MESSAGELIST), pack)
}

func (this *Player) ReadMessage(id string) {
	if msg, exist := this.msgs[id]; exist {
		if msg.State == model.MSGSTATE_UNREAD {
			msg.State = model.MSGSTATE_READED

			task.New(nil,
				task.CallableWrapper(func(o *basic.Object) interface{} {
					return model.ReadMessage(msg.Id, msg.Platform)
				}),
				task.CompleteNotifyWrapper(func(data interface{}, tt task.Task) {
					if data == nil {
						pack := &msg_proto.SCMessageRead{
							Id: proto.String(id),
						}
						proto.SetDefaults(pack)
						this.SendToClient(int(msg_proto.MSGPacketID_PACKET_SC_MESSAGEREAD), pack)
					}
				}), "ReadMessage").StartByFixExecutor("logic_message")
		}
	}
}

func (this *Player) WebDelMessage(id string) {
	if msg, exist := this.msgs[id]; exist {
		if msg.State != model.MSGSTATE_REMOVEED { // 未删除状态通知客户端
			pack := &msg_proto.SCMessageDel{
				Id: id,
			}

			proto.SetDefaults(pack)
			this.SendToClient(int(msg_proto.MSGPacketID_PACKET_SC_MESSAGEDEL), pack)
		}
		//删除此邮件
		delete(this.msgs, id)
	}
}

func (this *Player) DelMessage(id string, del int32) bool {
	if msg, exist := this.msgs[id]; exist {
		if msg.State != model.MSGSTATE_REMOVEED {

			task.New(nil,
				task.CallableWrapper(func(o *basic.Object) interface{} {
					args := &model.DelMsgArgs{
						Platform: msg.Platform,
						Id:       msg.Id,
						Del:      del,
					}
					return model.DelMessage(args)
				}),
				task.CompleteNotifyWrapper(func(data interface{}, tt task.Task) {
					if data == nil {
						pack := &msg_proto.SCMessageDel{
							Id: id,
						}

						proto.SetDefaults(pack)
						this.SendToClient(int(msg_proto.MSGPacketID_PACKET_SC_MESSAGEDEL), pack)

						msg.State = model.MSGSTATE_REMOVEED
						//删除此邮件
						delete(this.msgs, id)
					}
				}), "DelMessage").StartByFixExecutor("logic_message")

			return true
		}
	}
	return false
}
func (this *Player) MessageShowRed() {
	msgMap := make(map[int64]int)
	for _, msg := range this.msgs {
		if len(msgMap) == 3 {
			break
		}
		if msg.ShowId == model.HallAll {
			msgMap[model.HallMain] = 1
			msgMap[model.HallTienlen] = 1
			msgMap[model.HallFish] = 1
			break
		} else {
			if _, ok := msgMap[msg.ShowId]; !ok {
				if msg.State == model.MSGSTATE_UNREAD {
					msgMap[msg.ShowId] = 1
					continue
				}
				if msg.AttachState == model.MSGATTACHSTATE_DEFAULT {
					if msg.Coin > 0 || msg.Diamond > 0 || len(msg.Params) > 0 {
						msgMap[msg.ShowId] = 1
					}
				}
			}
		}
	}
	for showId := range msgMap {
		this.SendShowRed(hall_proto.ShowRedCode_Mail, int32(showId), 1)
	}
}

/*
func (this *Player) DelAllMessage() bool {

	var keys []string
	args := &model.DelAllMsgArgs{}
	pack := &msg_proto.SCMessageDel{}
	for key, _ := range this.msgs {
		if msg, exist := this.msgs[key]; exist {
			if msg.State == model.MSGSTATE_REMOVEED {
				break
			}
			msg.State = model.MSGSTATE_REMOVEED
			// model.DelMessage(msg.Id, msg.Platform)
			keys = append(keys, key)
			pack.Ids = append(pack.Ids, msg.Id.Hex())
			args.Ids = append(args.Ids, msg.Id)
		}
	}
	for _, key := range keys {
		delete(this.msgs, key)
	}

	task.New(nil,
		task.CallableWrapper(func(o *basic.Object) interface{} {
			return model.DelAllMessage(args)
		}),
		task.CompleteNotifyWrapper(func(data interface{}, tt task.Task) {
			if data == nil {
				pack := &msg_proto.SCMessageDel{}
				for _, id := range keys {
					pack.Ids = append(pack.Ids, id)
					delete(this.msgs, id)
				}
				proto.SetDefaults(pack)
				this.SendToClient(int(msg_proto.MSGPacketID_PACKET_SC_MESSAGEDEL), pack)
				for _, msg := range this.msgs {
					if msg.State == model.MSGSTATE_REMOVEED {
						break
					}
					msg.State = model.MSGSTATE_REMOVEED
				}

				//删除此邮件

			}
		}), "DelMessage").StartByFixExecutor("logic_message")

	proto.SetDefaults(pack)
	this.SendToClient(int(msg_proto.MSGPacketID_PACKET_SC_MESSAGEDEL), pack)

	return true

}*/

func (this *Player) AddMessage(msg *model.Message) {
	if msg == nil {
		return
	}

	if _, exist := this.msgs[msg.Id.Hex()]; !exist {
		this.msgs[msg.Id.Hex()] = msg

		//giftState := int32(0)
		//if len(msg.GiftId) > 0 {
		//	gift := GiftMgrSington.GetFromRecvGift(this.SnId, msg.GiftId)
		//	if gift != nil {
		//		giftState = gift.State
		//	} else {
		//		logger.Logger.Error("player: ", this.SnId, " not find gift : ", msg.GiftId)
		//	}
		//}

		//pack := &msg_proto.SCMessageAdd{
		//	Msg: &msg_proto.MessageData{
		//		Id:          proto.String(msg.Id.Hex()),
		//		Title:       proto.String(msg.Title),
		//		Content:     proto.String(msg.Content),
		//		MType:       proto.Int32(msg.MType),
		//		SrcId:       proto.Int32(msg.SrcId),
		//		SrcName:     proto.String(msg.SrcName),
		//		Coin:        proto.Int64(msg.Coin),
		//		Ticket:      proto.Int64(msg.Ticket),
		//		Grade:       proto.Int64(msg.Grade),
		//		State:       proto.Int32(msg.State),
		//		Ts:          proto.Int32(int32(msg.CreatTs)),
		//		Params:      msg.Params,
		//		AttachState: proto.Int32(msg.AttachState),
		//		GiftId:      proto.String(msg.GiftId),
		//		GiftState:   proto.Int32(giftState),
		//	},
		//}
		//
		//proto.SetDefaults(pack)
		//this.SendToClient(int(msg_proto.MSGPacketID_PACKET_SC_MESSAGEADD), pack)

		//if len(this.msgs) > model.MSG_MAX_COUNT {
		//如果邮件达到上限,删除最旧的一封邮件，调整一下，如果有附件没有领取，就不删除
		// OldestKeyId := msg.Id.Hex()
		this.verifyoneMessage(msg) // 最新的一封时间最大

		//}
		msgMap := make(map[int64]int)
		if msg.ShowId == model.HallAll {
			msgMap[model.HallMain] = 1
			msgMap[model.HallTienlen] = 1
			msgMap[model.HallFish] = 1
		} else {
			msgMap[msg.ShowId] = 1
		}
		for showId := range msgMap {
			this.SendShowRed(hall_proto.ShowRedCode_Mail, int32(showId), 1)
		}
	}
}

func (this *Player) verifyoneMessage(msg *model.Message) {
	var len int
	for _, m := range this.msgs {
		if m.ShowId == model.HallAll || m.ShowId&msg.ShowId != 0 {
			len++
		}
	}
	OldestKeyId := msg.Id.Hex()
	OldestCreatTs := msg.CreatTs

	if len > model.MSG_MAX_COUNT {
		for id, m := range this.msgs {
			if m.ShowId == model.HallAll || m.ShowId&msg.ShowId != 0 {
				if m.CreatTs < OldestCreatTs {
					OldestCreatTs = m.CreatTs
					OldestKeyId = id
				}
			}
		}
		this.DelMessage(OldestKeyId, 1)
	}
}

type msgsort struct {
	Id []string
	Ts []int64
}

func (p *msgsort) Len() int           { return len(p.Id) }
func (p *msgsort) Less(i, j int) bool { return p.Ts[i] > p.Ts[j] }
func (p *msgsort) Swap(i, j int) {
	p.Id[i], p.Id[j] = p.Id[j], p.Id[i]
	p.Ts[i], p.Ts[j] = p.Ts[j], p.Ts[i]
}
func (p *msgsort) Sort() { sort.Sort(p) }

// 初始化
func (this *Player) verifyMessage1() {
	var delId []string
	// delmap := make(map[string]bool)
	var mainsort, tiensort, fishsort msgsort

	for id, m := range this.msgs {

		if m.ShowId == model.HallAll { // 所有可见
			mainsort.Id = append(mainsort.Id, id)
			mainsort.Ts = append(mainsort.Ts, m.CreatTs)
			tiensort.Id = append(tiensort.Id, id)
			tiensort.Ts = append(tiensort.Ts, m.CreatTs)
			fishsort.Id = append(fishsort.Id, id)
			fishsort.Ts = append(fishsort.Ts, m.CreatTs)
		} else {
			if m.ShowId&model.HallMain != 0 {
				mainsort.Id = append(mainsort.Id, id)
				mainsort.Ts = append(mainsort.Ts, m.CreatTs)
			} else if m.ShowId&model.HallTienlen != 0 {
				tiensort.Id = append(tiensort.Id, id)
				tiensort.Ts = append(tiensort.Ts, m.CreatTs)
			} else if m.ShowId&model.HallFish != 0 {
				fishsort.Id = append(fishsort.Id, id)
				fishsort.Ts = append(fishsort.Ts, m.CreatTs)
			}
		}

	}

	if mainsort.Len() > model.MSG_MAX_COUNT {
		sort.Sort(&mainsort)
		delId = append(delId, mainsort.Id[model.MSG_MAX_COUNT:]...)
	}
	if tiensort.Len() > model.MSG_MAX_COUNT {
		sort.Sort(&tiensort)
		delId = append(delId, tiensort.Id[model.MSG_MAX_COUNT:]...)
	}
	if fishsort.Len() > model.MSG_MAX_COUNT {
		sort.Sort(&fishsort)
		delId = append(delId, fishsort.Id[model.MSG_MAX_COUNT:]...)
	}

	for _, id := range delId {
		this.DelMessage(id, 1)
	}
}

func (this *Player) verifyMessage(maxCreatTs int64) {
	var OldmainKeyId, OldsunKeyId []string
	var mainnum, sunnum int

	for id, m := range this.msgs {

		if m.CreatTs < maxCreatTs {
			if m.ShowId == model.HallAll || m.ShowId&model.HallMain != 0 {
				mainnum++
			} else {
				sunnum++
			}
			if mainnum > model.MSG_MAX_COUNT { // 主大厅
				OldmainKeyId = append(OldmainKeyId, id)
			} else if sunnum > model.MSG_MAX_COUNT { // 子大厅
				OldsunKeyId = append(OldsunKeyId, id)
			}
			/*if (m.Coin <= 0 && m.Ticket <= 0 && m.Grade <= 0) || m.AttachState == model.MSGATTACHSTATE_GOT {
				OldestCreatTs = m.CreatTs

			}*/
		}
	}

	for _, id := range OldmainKeyId {
		this.DelMessage(id, 1)
	}
	for _, id := range OldsunKeyId {
		this.DelMessage(id, 1)
	}
}

func (this *Player) EditMessage(msg *model.Message) {
	if msg == nil {
		return
	}

	if _, exist := this.msgs[msg.Id.Hex()]; exist {
		this.msgs[msg.Id.Hex()] = msg
	}
}

func (this *Player) SendIosInstallStableMail() {
	if this.layered[common.ActId_IOSINSTALLSTABLE] {
		logger.Logger.Trace("this.layered[common.ActId_IOSINSTALLSTABLE] is true")
		return
	}
	var newMsg *model.Message
	task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
		newMsg = model.NewMessage("", 0, "系统", this.SnId, model.MSGTYPE_IOSINSTALLSTABLE, "系统通知", fmt.Sprintf("感谢您下载稳定版本，额外奖励%d元，请查收", int(model.GameParamData.IosStableInstallPrize/100)),
			int64(model.GameParamData.IosStableInstallPrize), 0, 0, time.Now().Unix(), 0, "", nil, this.Platform, model.HallAll)
		return model.InsertMessage(this.Platform, newMsg)
	}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
		if data == nil {
			this.AddMessage(newMsg)
		}
	}), "SendMessage").Start()
}

func (this *Player) TestMail() {

	var newMsg *model.Message

	task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
		var otherParams []int32
		otherParams = append(otherParams, 10001, 3)
		otherParams = append(otherParams, 20001, 3)
		otherParams = append(otherParams, 20002, 3)
		newMsg = model.NewMessage("", 0, "系统", this.SnId, model.MSGTYPE_ITEM, "系统通知道具test", "测试",
			100000, 100, model.MSGSTATE_UNREAD, time.Now().Unix(), 0, "", otherParams, this.Platform, 0)
		return model.InsertMessage(this.Platform, newMsg)
	}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
		if data == nil {
			this.AddMessage(newMsg)

		}
	}), "TestSendMessage").Start()
}

func (this *Player) TestSubMail() {

	var newMsg *model.Message

	task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
		var otherParams []int32
		otherParams = append(otherParams, 10001, 3)
		otherParams = append(otherParams, 20001, 3)
		otherParams = append(otherParams, 20002, 3)
		newMsg = model.NewMessage("", 0, "系统", 0, model.MSGTYPE_ITEM, "系统", "测试",
			100000, 100, model.MSGSTATE_UNREAD, time.Now().Unix(), 0, "", otherParams, this.Platform, model.HallTienlen)
		return model.InsertMessage(this.Platform, newMsg)
	}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
		if data == nil {
			MsgMgrSington.AddMsg(newMsg)
		}
	}), "TestSubSendMessage").Start()
}

func (this *Player) ClubChangeCoin(gainWay int32, coin int64, remark string) {
	this.AddCoin(coin, gainWay, "", remark)
}
func (this *Player) GetMessageAttach(id string) {
	if msg, exist := this.msgs[id]; exist {
		if msg.AttachState == model.MSGATTACHSTATE_DEFAULT && (msg.Coin > 0 || msg.Ticket > 0 ||
			msg.Grade > 0 || len(msg.Params) > 0 || msg.Diamond > 0) {
			task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
				gift, err := model.GetMessageById(msg.Id.Hex(), msg.Platform)
				if err != nil || gift == nil {
					return nil
				}
				if gift.AttachState == model.MSGATTACHSTATE_GOT {
					return nil
				}
				err = model.GetMessageAttach(msg.Id, msg.Platform)
				if err != nil {
					return nil
				}
				return gift
			}), task.CompleteNotifyWrapper(func(data interface{}, tt task.Task) {
				attach_msg, ok := data.(*model.Message)
				dirtyCoin := int64(0)
				if ok && attach_msg != nil {
					msg.AttachState = model.MSGATTACHSTATE_GOT
					notifyClient := true
					var remark string
					var gainWay int32 = common.GainWay_MessageAttach
					switch msg.MType {
					case model.MSGTYPE_ITEM:
						remark = "领取道具"
						gainWay = common.GainWay_MAIL_MTEM
						dirtyCoin = msg.Coin
						items := make([]*Item, 0)
						if num := len(msg.Params); num > 0 && num%2 == 0 {
							for i := 0; i < num; i += 2 {
								items = append(items, &Item{
									ItemId:     msg.Params[i],   // 物品id
									ItemNum:    msg.Params[i+1], // 数量
									ObtainTime: time.Now().Unix(),
								})
							}
							if _, code := BagMgrSington.AddJybBagInfo(this, items); code != bag.OpResultCode_OPRC_Sucess { // 领取失败
								logger.Logger.Errorf("CSPlayerSettingHandler AddJybBagInfo err", code)
								pack := &msg_proto.SCGetMessageAttach{
									Id: proto.String(""),
								}
								proto.SetDefaults(pack)
								this.SendToClient(int(msg_proto.MSGPacketID_PACKET_SC_GETMESSAGEATTACH), pack)
							} else {
								var itemIds []int32
								for _, v := range items {
									itemIds = append(itemIds, v.ItemId)
									itemData := srvdata.PBDB_GameItemMgr.GetData(v.ItemId)
									if itemData != nil {
										BagMgrSington.RecordItemLog(this.Platform, this.SnId, ItemObtain, v.ItemId, itemData.Name, v.ItemNum, "邮件领取")
									}
								}
								PetMgrSington.CheckShowRed(this)
							}
							this.dirty = true
						}
					case model.MSGTYPE_IOSINSTALLSTABLE:
						remark = "IOS下载稳定版本"
						gainWay = common.GainWay_IOSINSTALLSTABLE
						dirtyCoin = msg.Coin
					case model.MSGTYPE_GIFT:
						remark = "礼物"
					case model.MSGTYPE_GOLDCOMERANK:
						remark = "财神降临奖励"
						gainWay = common.GainWay_GoldCome
						notifyClient = false
						dirtyCoin = msg.Coin
					case model.MSGTYPE_RANDCOIN:
						remark = "红包雨"
						gainWay = common.GainWay_OnlineRandCoin
						notifyClient = false
						dirtyCoin = msg.Coin
					case model.MSGTYPE_REBATE:
						remark = "流水返利"
						gainWay = common.GainWay_RebateTask
						notifyClient = false
						dirtyCoin = msg.Coin
						//邮件领取 添加日志
						task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
							return model.InsertRebateLog(this.Platform, &model.Rebate{
								SnId:        this.SnId,
								RebateCoin:  msg.Coin,
								ReceiveType: 1,
								CodeCoin:    0,
							})
						}), nil, "InsertRebateLog").StartByFixExecutor("ReceiveCodeCoin")
					case model.MSGTYPE_ClubGet:
						//if len(msg.Params) != 0 {
						//	//如果俱乐部解散 就存msg.Params[0]
						//	remark = fmt.Sprintf("%v", msg.Params[0])
						//}
						//gainWay = common.GainWay_ClubGetCoin
						//dirtyCoin = msg.Coin
					case model.MSGTYPE_ClubPump:
						//if len(msg.Params) != 0 {
						//	remark = fmt.Sprintf("%v", msg.Params[0])
						//}
						//gainWay = common.GainWay_ClubPumpCoin
						//notifyClient = false
						//dirtyCoin = msg.Coin
					case model.MSGTYPE_MATCH_SIGNUPFEE:
						gainWay = common.GainWay_MatchBreakBack
						notifyClient = false
					case model.MSGTYPE_MATCH_TICKETREWARD:
						gainWay = common.GainWay_MatchSystemSupply
						notifyClient = false
						this.TicketTotal += msg.Ticket
					case model.MSGTYPE_MATCH_SHOPEXCHANGE:
						remark = "积分商城兑换"
						gainWay = common.GainWay_Exchange
					case model.MSGTYPE_MATCH_SHOPERETURN:
						remark = "撤单返还"
						gainWay = common.GainWay_GradeShopReturn
					}
					if msg.Coin > 0 {
						this.AddCoin(msg.Coin, gainWay, msg.Id.Hex(), remark)
						//增加泥码
						this.AddDirtyCoin(0, dirtyCoin)
						//俱乐部获取不算系统赠送
						if msg.MType != model.MSGTYPE_ClubGet {
							this.ReportSystemGiveEvent(int32(msg.Coin), gainWay, notifyClient) //邮件附件算是系统赠送
						} else { //俱乐部获取算充值
							this.AddCoinGiveLog(msg.Coin, 0, 0, gainWay, model.COINGIVETYPE_PAY, "club", "club")
						}
						this.AddPayCoinLog(msg.Coin, model.PayCoinLogType_Coin, "mail")
					}
					if msg.Ticket > 0 {
						//增加报名券
						this.AddTicket(msg.Ticket, gainWay, msg.Id.Hex(), remark)
					}
					if msg.Grade > 0 {
						//增加积分
						this.AddGrade(msg.Grade, gainWay, msg.Id.Hex(), remark)
					}
					if msg.Diamond > 0 {
						this.AddDiamond(msg.Diamond, gainWay, msg.Id.Hex(), remark)
					}
					pack := &msg_proto.SCGetMessageAttach{
						Id: proto.String(id),
					}
					proto.SetDefaults(pack)
					this.SendToClient(int(msg_proto.MSGPacketID_PACKET_SC_GETMESSAGEATTACH), pack)
				}
			}), "GetMessageAttach").StartByFixExecutor("logic_message")
		}
	}
}

// 一键领取
func (this *Player) GetMessageAttachs(ids []string) {
	var msgs []*model.Message
	var Ids []string // 可以领取的邮件
	var platform string
	for _, id := range ids {
		if msg, exist := this.msgs[id]; exist {
			if msg.AttachState == model.MSGATTACHSTATE_DEFAULT && (msg.Coin > 0 || msg.Ticket > 0 ||
				msg.Grade > 0 || len(msg.Params) > 0 || msg.Diamond > 0) {
				Ids = append(Ids, id)
				platform = msg.Platform
			}
		}
	}

	task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {

		magids, err := model.GetMessageAttachs(Ids, platform)
		if err != nil {
			logger.Logger.Trace("GetMessageAttachs err ", err)
		}
		return magids
	}), task.CompleteNotifyWrapper(func(data interface{}, tt task.Task) {
		magids, ok := data.(*[]string)

		if ok && magids != nil {
			for _, id := range *magids {
				if msg, exist := this.msgs[id]; exist {
					if msg.AttachState == model.MSGATTACHSTATE_DEFAULT && (msg.Coin > 0 || msg.Ticket > 0 ||
						msg.Grade > 0 || len(msg.Params) > 0 || msg.Diamond > 0) {
						msgs = append(msgs, msg)
						platform = msg.Platform
					}
				}
			}

			pack := &msg_proto.SCGetMessageAttach{
				// Id: proto.String(id),
			}

			for _, msg := range msgs {
				pack.Ids = append(pack.Ids, msg.Id.Hex())
				dirtyCoin := int64(0)
				msg.AttachState = model.MSGATTACHSTATE_GOT
				notifyClient := true
				var remark string
				var gainWay int32 = common.GainWay_MessageAttach
				switch msg.MType {
				case model.MSGTYPE_ITEM:
					remark = "领取道具"
					gainWay = common.GainWay_MAIL_MTEM
					dirtyCoin = msg.Coin
					items := make([]*Item, 0)
					if num := len(msg.Params); num > 0 && num%2 == 0 {
						for i := 0; i < num; i += 2 {
							items = append(items, &Item{
								ItemId:     msg.Params[i],   // 物品id
								ItemNum:    msg.Params[i+1], // 数量
								ObtainTime: time.Now().Unix(),
							})
						}
						if _, code := BagMgrSington.AddJybBagInfo(this, items); code != bag.OpResultCode_OPRC_Sucess { // 领取失败
							logger.Logger.Errorf("CSPlayerSettingHandler AddJybBagInfo err", code)
							/*
								pack := &msg_proto.SCGetMessageAttach{
									Id: proto.String(""),
								}
								proto.SetDefaults(pack)
								this.SendToClient(int(msg_proto.MSGPacketID_PACKET_SC_GETMESSAGEATTACH), pack)
							*/
						} else {
							var itemIds []int32
							for _, v := range items {
								itemIds = append(itemIds, v.ItemId)
								itemData := srvdata.PBDB_GameItemMgr.GetData(v.ItemId)
								if itemData != nil {
									BagMgrSington.RecordItemLog(this.Platform, this.SnId, ItemObtain, v.ItemId, itemData.Name, v.ItemNum, "邮件领取")
								}
							}
							PetMgrSington.CheckShowRed(this)
						}
						this.dirty = true
					}
				case model.MSGTYPE_IOSINSTALLSTABLE:
					remark = "IOS下载稳定版本"
					gainWay = common.GainWay_IOSINSTALLSTABLE
					dirtyCoin = msg.Coin
				case model.MSGTYPE_GIFT:
					remark = "礼物"
				case model.MSGTYPE_GOLDCOMERANK:
					remark = "财神降临奖励"
					gainWay = common.GainWay_GoldCome
					notifyClient = false
					dirtyCoin = msg.Coin
				case model.MSGTYPE_RANDCOIN:
					remark = "红包雨"
					gainWay = common.GainWay_OnlineRandCoin
					notifyClient = false
					dirtyCoin = msg.Coin
				case model.MSGTYPE_REBATE:
					remark = "流水返利"
					gainWay = common.GainWay_RebateTask
					notifyClient = false
					dirtyCoin = msg.Coin
					//邮件领取 添加日志
					task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
						return model.InsertRebateLog(this.Platform, &model.Rebate{
							SnId:        this.SnId,
							RebateCoin:  msg.Coin,
							ReceiveType: 1,
							CodeCoin:    0,
						})
					}), nil, "InsertRebateLog").StartByFixExecutor("ReceiveCodeCoin")
				case model.MSGTYPE_ClubGet:
					//if len(msg.Params) != 0 {
					//	//如果俱乐部解散 就存msg.Params[0]
					//	remark = fmt.Sprintf("%v", msg.Params[0])
					//}
					//gainWay = common.GainWay_ClubGetCoin
					//dirtyCoin = msg.Coin
				case model.MSGTYPE_ClubPump:
					//if len(msg.Params) != 0 {
					//	remark = fmt.Sprintf("%v", msg.Params[0])
					//}
					//gainWay = common.GainWay_ClubPumpCoin
					//notifyClient = false
					//dirtyCoin = msg.Coin
				case model.MSGTYPE_MATCH_SIGNUPFEE:
					gainWay = common.GainWay_MatchBreakBack
					notifyClient = false
				case model.MSGTYPE_MATCH_TICKETREWARD:
					gainWay = common.GainWay_MatchSystemSupply
					notifyClient = false
					this.TicketTotal += msg.Ticket
				case model.MSGTYPE_MATCH_SHOPEXCHANGE:
					remark = "积分商城兑换"
					gainWay = common.GainWay_Exchange
				case model.MSGTYPE_MATCH_SHOPERETURN:
					remark = "撤单返还"
					gainWay = common.GainWay_GradeShopReturn
				}
				if msg.Coin > 0 {
					this.AddCoin(msg.Coin, gainWay, msg.Id.Hex(), remark)
					//增加泥码
					this.AddDirtyCoin(0, dirtyCoin)
					//俱乐部获取不算系统赠送
					if msg.MType != model.MSGTYPE_ClubGet {
						this.ReportSystemGiveEvent(int32(msg.Coin), gainWay, notifyClient) //邮件附件算是系统赠送
					} else { //俱乐部获取算充值
						this.AddCoinGiveLog(msg.Coin, 0, 0, gainWay, model.COINGIVETYPE_PAY, "club", "club")
					}
					this.AddPayCoinLog(msg.Coin, model.PayCoinLogType_Coin, "mail")
				}
				if msg.Ticket > 0 {
					//增加报名券
					this.AddTicket(msg.Ticket, gainWay, msg.Id.Hex(), remark)
				}
				if msg.Grade > 0 {
					//增加积分
					this.AddGrade(msg.Grade, gainWay, msg.Id.Hex(), remark)
				}
				if msg.Diamond > 0 {
					this.AddDiamond(msg.Diamond, gainWay, msg.Id.Hex(), remark)
				}

			}

			proto.SetDefaults(pack)
			this.SendToClient(int(msg_proto.MSGPacketID_PACKET_SC_GETMESSAGEATTACH), pack)
		}
	}), "GetMessageAttach").StartByFixExecutor("logic_message")

}

func (this *Player) GetMessageByGiftId(id string) *model.Message {
	for _, msg := range this.msgs {
		if msg.GiftId == id && msg.State != model.MSGSTATE_REMOVEED {
			return msg
		}
	}
	return nil
}

// 踢掉线
func (this *Player) Kickout(reason int32) {
	if this.IsOnLine() {
		logger.Logger.Trace("(this *Player) Kickout()", this.SnId)
		scDisconnect := &login_proto.SSDisconnect{
			SessionId: proto.Int64(this.sid),
			Type:      proto.Int32(reason),
		}
		proto.SetDefaults(scDisconnect)
		this.SendToClient(int(login_proto.GatePacketID_PACKET_SS_DICONNECT), scDisconnect)

		LoginStateMgrSington.LogoutBySid(this.sid)
		this.DropLine()
		this.DgGameLogout()
	}
}

// 掉线
func (this *Player) DropLine() {
	logger.Logger.Tracef("(this *Player) DropLine() %v", this.SnId)
	logState := LoginStateMgrSington.GetLoginStateOfSid(this.sid)
	var clog *model.ClientLoginInfo
	if logState != nil {
		clog = logState.clog
	}
	if clog == nil {
		//logger.Logger.Errorf("(this *Player[%v]) DropLine() Can't find logState[%v] clog", this.SnId, logState)
		return
	}
	//排除掉机器人
	if !this.IsRob {
		var gameid int
		if this.scene != nil && this.scene.gameSess != nil {
			gameid = this.scene.gameId
		}
		LogChannelSington.WriteLog(model.NewLoginLog(this.SnId, model.LoginLogType_Drop, this.Tel, this.Ip,
			this.Platform, this.Channel, this.BeUnderAgentCode, this.PackageID, this.City, clog, this.GetTotalCoin(),
			gameid))
	}
	this.state = PlayerState_Offline
	this.PlayerData.LastLogoutTime = time.Now().Local()
	FriendMgrSington.UpdateLogoutTime(this.SnId)
	if this.scene != nil && this.scene.gameSess != nil {
		pack := &server_proto.WGPlayerDropLine{
			Id:      proto.Int32(this.SnId),
			SceneId: proto.Int(this.scene.sceneId),
		}
		proto.SetDefaults(pack)
		this.SendToGame(int(server_proto.SSPacketID_PACKET_WG_PLAYERDROPLINE), pack)
	}
	PlayerMgrSington.DroplinePlayer(this)
	//clubManager.DropLinePlayer(this.SnId)
	PlatformMgrSington.PlayerLogout(this)
	if !this.IsRobot() {
		PlayerOnlineSington.Check = true
	}
	this.sid = 0
	this.gateSess = nil
	//统计在线时长日志
	this.StatisticsOllen(this.PlayerData.LastLogoutTime)

	gameStateMgr.PlayerClear(this)

	//玩家掉线事件
	FirePlayerDropLine(this)
}

// 退出
func (this *Player) Logout() {
	logger.Logger.Tracef("(this *Player) Logout() %v", this.SnId)
	//退出比赛
	//this.QuitMatch(false)
	// 在线奖励：累计在线时长
	//this.OnlineRewardAddUpOnlineDuration()

	scLogout := &login_proto.SCLogout{
		OpRetCode: login_proto.OpResultCode_OPRC_Sucess,
	}
	proto.SetDefaults(scLogout)
	this.SendToClient(int(login_proto.LoginPacketID_PACKET_SC_LOGOUT), scLogout)
	this.state = PlayerState_Offline
	this.LastLogoutTime = time.Now().Local()
	FriendMgrSington.UpdateLogoutTime(this.SnId)
	//clubManager.DropLinePlayer(this.SnId)
	PlayerMgrSington.DroplinePlayer(this)
	PlatformMgrSington.PlayerLogout(this)
	if !this.IsRobot() {
		PlayerOnlineSington.Check = true
	}
	this.sid = 0
	this.gateSess = nil
	this.DgGameLogout()
	gameStateMgr.PlayerClear(this)
	CoinSceneMgrSington.PlayerTryLeaveQueue(this)
}
func (this *Player) DgGameLogout() {
	//if this.scene != nil {
	//	if this.scene.sceneId == SceneMgrSington.GetDgSceneId() {
	//		var agentName, agentKey, thirdPlf string
	//		if len(this.BakDgHboName) > 0 {
	//			if strings.Contains(this.BakDgHboName, "dg") {
	//				agentName, agentKey, thirdPlf = model.OnlyGetDgConfigByPlatform(this.Platform)
	//			} else {
	//				agentName, agentKey, thirdPlf = model.OnlyGetHboConfigByPlatform(this.Platform)
	//			}
	//
	//			task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
	//				webapi.API_DgLogout(thirdPlf, common.GetAppId(), this.DgGame, this.DgPass, agentName, agentKey)
	//				return nil
	//			}), task.CompleteNotifyWrapper(func(data interface{}, t *task.Task) {
	//				this.scene = nil
	//			}), "DgGameLogout").Start()
	//		}
	//
	//	}
	//}
}
func (this *Player) ThirdGameLogout() {
	_StartTransferThird2SystemTask(this)
}
func (this *Player) IsOnLine() bool {
	return this.state != PlayerState_Offline
}

func (this *Player) OnLogouted() {
	logger.Logger.Tracef("(this *Player) OnLogouted() %v", this.SnId)

	//离线玩家数据
	//OfflinePlayerMgrSington.PlayerLogout(this)
	//在线时长日志
	this.WriteOllenLog()
	if !this.IsRob {
		FriendMgrSington.SaveFriendData(this.SnId, true)
		FriendUnreadMgrSington.SaveFriendUnreadData(this.Platform, this.SnId)
		MatchSeasonMgrSington.SaveMatchSeasonData(this.SnId, true)
	}
	//平台数据
	PlatformMgrSington.PlayerLogout(this)
	PlayerSingleAdjustMgr.DelPlayerData(this.Platform, this.SnId)

	//确保最后数据的保存
	this.dirty = true
	this.Force2Save()
	this.dirty = false
	//离线玩家清空俱乐部信息
	//delete(clubManager.theInClubId, this.SnId)

	//玩家登出事件
	FirePlayerLogouted(this)
	//退出事件
	this.ge.eventType = GAMEEVENT_LOGOUT
	this.ge.eventSrc = nil
	this.FireGameEvent()

	//登录日志
	logState := LoginStateMgrSington.GetLoginStateOfSid(this.sid)
	var clog *model.ClientLoginInfo
	if logState != nil {
		clog = logState.clog
	}
	//排除掉机器人
	if !this.IsRob {
		LogChannelSington.WriteLog(model.NewLoginLog(this.SnId, model.LoginLogType_Logout, this.Tel, this.Ip,
			this.Platform, this.Channel, this.BeUnderAgentCode, this.PackageID, this.City, clog, this.GetTotalCoin(),
			0))
	}

	//退出通知
	ActMonitorMgrSington.SendActMonitorEvent(ActState_Login, this.SnId, this.Name, this.Platform,
		0, 0, "", 1)

}

func (this *Player) MarshalData(gameid int) (d []byte, e error) {
	d, e = netlib.Gob.Marshal(this.PlayerData)
	return
}
func (this *Player) MarshalSingleAdjustData(gamefreeid int32) (d []byte, e error) {
	if this.IsRob {
		return
	}
	sa := PlayerSingleAdjustMgr.GetSingleAdjust(this.Platform, this.SnId, gamefreeid)
	if sa != nil {
		d, e = netlib.Gob.Marshal(sa)
	}
	return
}
func (this *Player) UnmarshalData(data []byte, scene *Scene) {
	pd := &model.PlayerData{}
	err := netlib.Gob.Unmarshal(data, pd)
	if err == nil {
		key := scene.dbGameFree.GetGameDif()
		if d, ok := pd.GDatas[key]; ok {
			this.GDatas[key] = d
		}
		key = strconv.Itoa(int(scene.dbGameFree.GetId()))
		if d, ok := pd.GDatas[key]; ok {
			this.GDatas[key] = d
		}
		this.LastRechargeWinCoin = pd.LastRechargeWinCoin
		oldRecharge := int64(0)
		oldExchange := int64(0)
		if this.PlayerData.TodayGameData != nil {
			oldRecharge = this.PlayerData.TodayGameData.RechargeCoin
			oldExchange = this.PlayerData.TodayGameData.ExchangeCoin
		}
		this.PlayerData.TodayGameData = pd.TodayGameData
		if this.PlayerData.TodayGameData != nil {
			this.PlayerData.TodayGameData.RechargeCoin = oldRecharge
			this.PlayerData.TodayGameData.ExchangeCoin = oldExchange
		}
		this.PlayerData.YesterdayGameData = pd.YesterdayGameData
		this.PlayerData.IsFoolPlayer = pd.IsFoolPlayer
		this.PlayerData.TotalGameData = pd.TotalGameData
		this.PlayerData.WinTimes = pd.WinTimes
		this.PlayerData.FailTimes = pd.FailTimes
		this.PlayerData.DrawTimes = pd.DrawTimes
		this.dirty = true
	} else {
		logger.Logger.Warn("Player.SyncData err:", err)
	}
}

func (this *Player) MarshalIParam() []*server_proto.PlayerIParam {
	var params []*server_proto.PlayerIParam
	for i, v := range this.Iparams {
		params = append(params, &server_proto.PlayerIParam{
			ParamId: proto.Int(i),
			IntVal:  proto.Int64(v),
		})
	}
	return params
}

func (this *Player) UnmarshalIParam(params []*server_proto.PlayerIParam) {
	for _, p := range params {
		this.Iparams[int(p.GetParamId())] = p.GetIntVal()
	}
}

func (this *Player) MarshalSParam() []*server_proto.PlayerSParam {
	var params []*server_proto.PlayerSParam
	for i, v := range this.sparams {
		params = append(params, &server_proto.PlayerSParam{
			ParamId: proto.Int(i),
			StrVal:  proto.String(v),
		})
	}
	return params
}

func (this *Player) UnmarshalSParam(params []*server_proto.PlayerSParam) {
	for _, p := range params {
		this.sparams[int(p.GetParamId())] = p.GetStrVal()
	}
}

func (this *Player) MarshalCParam() []*server_proto.PlayerCParam {
	var params []*server_proto.PlayerCParam
	for k, v := range this.cparams {
		params = append(params, &server_proto.PlayerCParam{
			StrKey: proto.String(k),
			StrVal: proto.String(v),
		})
	}
	return params
}

func (this *Player) Force2Save() {
	if this.CanDelete() { //防止循环调用产生的死循环
		return
	}
	this.Time2Save()
}

func (this *Player) SendToRepSrv(pd *model.PlayerData) {
	replaySess := srvlib.ServerSessionMgrSington.GetSession(common.GetSelfAreaId(), ReplayServerType, ReplayServerId)
	if replaySess != nil {
		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		err := enc.Encode(pd)
		if err != nil {
			logger.Logger.Info("(this *Player) SendToRepSrv json.Marshal error", err)
		} else {
			pack := &server_proto.WRPlayerData{
				PlayerData: buf.Bytes(),
			}
			proto.SetDefaults(pack)
			replaySess.Send(int(server_proto.SSPacketID_PACKET_WR_PlayerData), pack)
		}
	}
}

func (this *Player) Time2Save() {
	if this.isDelete { //删档用户不保存
		this.dirty = false
	}
	if this.dirty {
		this.dirty = false
		if !this.IsRob { //机器人数据不再保存

			logger.Logger.Infof("(this *Player) Time2Save() %v", this.SnId)
			pi := model.ClonePlayerData(this.PlayerData)
			if pi != nil {
				this.SendToRepSrv(pi)
				if pi.IsRob {
					pi.Platform = common.Platform_Rob
					pi.Channel = common.Channel_Rob
				}
				pi.LastLogoutTime = time.Now().Local()
				tl := &SaveTask{pi: pi, snid: this.SnId}
				t := task.New(nil, tl, tl, "SavePlayerTask")
				if b := t.StartByExecutor(strconv.Itoa(int(this.SnId))); b {
					this.lastSaved = time.Now()
				}
			}
			//背包数据存储
			BagMgrSington.SaveBagData(this.SnId, this.Platform)
		}
	} else {
		if this.CanDelete() {
			//this.QuitMatch(false)
			if !this.dirty {
				// 用户缓存数据清除前同步到player_logicleveldata
				if !this.IsRob {
					pi := model.ClonePlayerData(this.PlayerData)
					if pi != nil {
						this.SendToRepSrv(pi)
					}
				}
				PlayerMgrSington.DelPlayer(this.SnId)
			}
			return
		}
	}

	//在线时长日志
	if !this.IsOnLine() {
		this.WriteOllenLog()
	}

	//clubManager.UpdateClubPlayerData(this.SnId)
}

func (this *Player) CanDelete() bool {
	if this.isDelete {
		return true
	}
	return this.state == PlayerState_Offline &&
		!this.dirty &&
		time.Now().Sub(this.lastSaved) > time.Minute*5 &&
		SceneMgrSington.GetSceneByPlayerId(this.SnId) == nil
}

func (this *Player) GetCoin() int64 {
	return this.Coin
}
func (this *Player) TotalData(num int64, gainWay int32) {
	if this.IsRob {
		return
	}
	//num = int64(math.Abs(float64(num)))
	//sort := common.GetSortByGainWay(gainWay)
	//switch sort {
	//case common.GainWaySort_Act:
	//	//活动金额累加
	//	this.ActivityCoin += int32(num)
	//case common.GainWaySort_Club:
	//	switch gainWay {
	//	case common.GainWay_ClubGiveCoin: //出账
	//		//俱乐部出账
	//		this.ClubOutCoin += num
	//	case common.GainWay_ClubGetCoin:
	//		//俱乐部入账
	//		this.ClubInCoin += num
	//	}
	//case common.GainWaySort_Rebate:
	//	//返利获取 也叫 手动洗码
	//	this.TotalRebateCoin += num
	//}
}

func (this *Player) AddDiamond(num int64, gainWay int32, oper, remark string) {
	if num == 0 {
		return
	}
	logger.Logger.Tracef("snid(%v)  AddDiamond(%v)", this.SnId, num)
	//async := false
	//if num > 0 && this.scene != nil && !this.scene.IsTestScene() && this.scene.sceneMode != common.SceneMode_Thr { //游戏场中加币,需要同步到gamesrv上
	//	if StartAsyncAddCoinTransact(this, num, gainWay, oper, remark, true, 0, true) {
	//		async = true
	//	}
	//}

	if num != 0 /*&& !async*/ {
		this.dirty = true
		if num > 0 {
			this.Diamond += num
		} else {
			if -num > this.Diamond {
				logger.Logger.Errorf("Player.AddCoin exception!!! num(%v) oper(%v)", num, oper)
				num = -this.Diamond
				this.Diamond = 0
			} else {
				this.Diamond += num
			}
		}

		this.SendDiffData()
		if !this.IsRob {
			log := model.NewCoinLogDiamondEx(this.SnId, num, this.Coin, this.SafeBoxCoin,
				this.Diamond, this.Ver, gainWay, 0, oper, remark, this.Platform, this.Channel,
				this.BeUnderAgentCode, 1, this.PackageID, 0)
			if log != nil {
				LogChannelSington.WriteLog(log)
			}
		}
	}
}
func (this *Player) AddCoin(num int64, gainWay int32, oper, remark string) {
	if num == 0 {
		return
	}
	logger.Logger.Tracef("snid(%v)  AddCoin(%v)", this.SnId, num)

	//this.TotalData(num, gainWay)

	async := false
	if num > 0 && this.scene != nil && !this.scene.IsTestScene() && this.scene.sceneMode != common.SceneMode_Thr { //游戏场中加币,需要同步到gamesrv上
		if StartAsyncAddCoinTransact(this, num, gainWay, oper, remark, true, 0, true) {
			async = true
		}
	}

	if num != 0 && !async {
		this.dirty = true
		if num > 0 {
			this.Coin += num
		} else {
			if -num > this.Coin {
				logger.Logger.Errorf("Player.AddCoin exception!!! num(%v) oper(%v)", num, oper)
				num = -this.Coin
				this.Coin = 0
			} else {
				this.Coin += num
			}
		}

		this.SendDiffData()
		if !this.IsRob {
			log := model.NewCoinLogDiamondEx(this.SnId, num, this.Coin, this.SafeBoxCoin,
				this.Diamond, this.Ver, gainWay, 0, oper, remark, this.Platform, this.Channel,
				this.BeUnderAgentCode, 0, this.PackageID, 0)
			if log != nil {
				LogChannelSington.WriteLog(log)
			}
		}
	}
}

func (this *Player) AddCoinAsync(num int64, gainWay int32, oper, remark string, broadcast bool, retryCnt int, writeLog bool) bool {
	if num == 0 {
		return false
	}

	if retryCnt == 0 {
		this.TotalData(num, gainWay)
	}

	//玩家可能正在换房间
	async := false
	if num > 0 && retryCnt < 3 && this.scene != nil && !this.scene.IsTestScene() && this.scene.sceneMode != common.SceneMode_Thr { //游戏场中加币,需要同步到gamesrv上
		if StartAsyncAddCoinTransact(this, num, gainWay, oper, remark, broadcast, retryCnt, writeLog) {
			async = true
		}
	}

	if num != 0 && !async {
		this.dirty = true
		if num > 0 {
			this.Coin += num
		} else {
			if -num > this.Coin {
				logger.Logger.Errorf("Player.AddCoin exception!!! num(%v) oper(%v)", num, oper)
				num = -this.Coin
				this.Coin = 0
			} else {
				this.Coin += num
			}
		}

		this.SendDiffData()
		if !this.IsRob && writeLog {
			restCnt := this.Coin
			log := model.NewCoinLogEx(this.SnId, num, restCnt, this.SafeBoxCoin, this.Ver, gainWay, 0,
				oper, remark, this.Platform, this.Channel, this.BeUnderAgentCode, 0, this.PackageID, 0)
			if log != nil {
				LogChannelSington.WriteLog(log)
			}
		}
	}

	return async
}

//func (this *Player) AddClubCoin(num int64, gainWay int32, oper, remark string) {
//	if num == 0 {
//		return
//	}
//	this.TotalData(num, gainWay)
//	this.ClubCoin += num
//
//	if num != 0 {
//		this.dirty = true
//		this.SendDiffData()
//		restCnt := this.ClubCoin
//		log := model.NewCoinLogEx(this.SnId, num, restCnt, this.SafeBoxCoin, this.Ver, gainWay, 0,
//			oper, remark, this.Platform, this.Channel, this.BeUnderAgentCode, 0, this.PackageID, 0)
//		if log != nil {
//			CoinLogChannelSington.Write(log)
//		}
//	}
//}

// 增加泥码
func (this *Player) AddDirtyCoin(paycoin, givecoin int64) {
	if this.IsRob {
		return
	}

	//if cfg, ok := ProfitControlMgrSington.GetCfg(this.Platform); ok && cfg != nil && paycoin >= 0 {
	//	//洗码折算率=(玩家剩余泥码*洗码折算率+期望营收)/(充值额+赠送额+泥码余额)
	//	this.RecalcuWashingCoinConvRate(cfg.Rate, paycoin, givecoin)
	//}
	//
	//this.DirtyCoin += paycoin + givecoin
	//if this.DirtyCoin < 0 {
	//	this.DirtyCoin = 0
	//}
	this.dirty = true
}

// 洗码
func (this *Player) WashingCoin(coin int64) int64 {
	if this.IsRob {
		return 0
	}
	if coin <= 0 {
		return 0
	}

	//if this.DirtyCoin > coin {
	//	this.DirtyCoin -= coin
	//	this.dirty = true
	//	return coin
	//}
	//
	////剩余多少泥码，清洗多少
	//coin = this.DirtyCoin
	//this.DirtyCoin = 0
	return coin
}

func (this *Player) AddTicket(num int64, gainWay int32, oper, remark string) {
	if num == 0 {
		return
	}
	if num > 0 {
		this.Ticket += num
	} else {
		if -num > this.Ticket {
			logger.Logger.Errorf("Player.AddTicket exception!!! num(%v) oper(%v)", num, oper)
			num = -this.Ticket
			this.Ticket = 0
		} else {
			this.Ticket += num
		}
	}
	if num != 0 {
		this.dirty = true
		this.diffData.Ticket = -1
		this.SendDiffData()
		//restCnt := this.Ticket
		//log := model.NewTicketLogEx(this.SnId, num, restCnt, this.Ver, gainWay, oper, remark, this.Platform, this.Channel, this.BeUnderAgentCode, this.PackageID, this.InviterId)
		//if log != nil {
		//	TicketLogChannelSington.Write(log)
		//}
	}
}

func (this *Player) AddGrade(num int64, gainWay int32, oper, remark string) {
	if num == 0 {
		return
	}
	if num > 0 {
		this.Grade += num
	} else {
		if -num > this.Grade {
			logger.Logger.Errorf("Player.AddGrade exception!!! num(%v) oper(%v)", num, oper)
			num = -this.Grade
			this.Grade = 0
		} else {
			this.Grade += num
		}
	}
	if num != 0 {
		this.dirty = true
		this.diffData.Grade = -1
		this.SendDiffData()
		//restCnt := this.Grade
		//log := model.NewGradeLogEx(this.SnId, num, restCnt, this.Ver, gainWay, oper, remark, this.Platform, this.Channel, this.BeUnderAgentCode, this.PackageID, this.InviterId)
		//if log != nil {
		//	GradeLogChannelSington.Write(log)
		//}
	}
}

func (this *Player) OnSecTimer() {
	FirePlayerSecTimer(this)
}

func (this *Player) OnMiniTimer() {
	FirePlayerMiniTimer(this)
}

func (this *Player) OnHourTimer() {
	FirePlayerHourTimer(this)
}

func (this *Player) OnDayTimer(login, continuous bool, t int) {

	//增加此功能，0:0:0 上线的玩家，会更新两次
	if common.InSameDayNoZero(time.Now().Local(), this.lastOnDayChange) {
		return
	}

	logger.Logger.Infof("(this *Player) (%v) OnDayTimer(%v,%v) ", this.SnId, login, continuous)

	this.dirty = true

	//更新在线时间
	if !login {
		this.StatisticsOllen(time.Now().Local())
		this.WriteOllenLog()
		//PayActMgrSington.OnDayChange(this)

	}

	if login || this.scene == nil {
		//跨天登录 数据给昨天，今天置为空
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

	if !login {
		this.ge.eventType = GAMEEVENT_LOGIN
		this.ge.eventSrc = nil
		this.FireGameEvent()
	}

	this.OnTimeDayTotal(continuous, t)

	//标记已经更新,保持最后
	this.lastOnDayChange = time.Now().Local()

	FirePlayerDayTimer(this, login, continuous)

	//商城数据更新
	this.ShopTotal = nil
	this.ShopLastLookTime = nil
	// 福利活动更新
	this.WelfData = nil
	//七日活动
	ActSignMgrSington.OnDayChanged(this)

}
func (this *Player) OnTimeDayTotal(continuous bool, t int) {
	for k, tgd := range this.TotalGameData {
		for i := 0; i < t; i++ {
			tgd = append(tgd, new(model.PlayerGameTotal))
		}
		if len(tgd) > 30 {
			tgd = tgd[len(tgd)-30:]
		}
		this.TotalGameData[k] = tgd
	}
}

//func (this *Player) OnTimeDayFlow() {
//	this.TotalConvertibleFlow = 0
//	this.ExchangeTotal = 0
//	this.diffData.TotalConvertibleFlow = -1
//	this.SendDiffData()
//}

func (this *Player) OnMonthTimer() {
	//判断是否一天即可过滤0点多次切换
	if common.InSameDayNoZero(time.Now().Local(), this.lastOnMonthChange) {
		return
	}
	//保持最后
	this.lastOnMonthChange = time.Now().Local()

	FirePlayerMonthTimer(this)
}

func (this *Player) OnWeekTimer() {
	//判断是否一天即可过滤0点多次切换
	if common.InSameDayNoZero(time.Now().Local(), this.lastOnWeekChange) {
		return
	}

	//清理比赛券
	ticket := this.Ticket
	if ticket > 0 {
		this.AddTicket(-ticket, common.GainWay_Expire, "system", "过期清理")
		this.TicketTotalDel += ticket
		this.dirty = true
	}

	//保持最后
	this.lastOnWeekChange = time.Now().Local()

	if len(this.msgs) > 0 {
		//自然过渡执行
		//过期邮件处理
		var keysId []string
		for key, msg := range this.msgs {
			if msg.AttachState == model.MSGATTACHSTATE_DEFAULT && msg.Ticket > 0 {
				this.TicketTotalDel += msg.Ticket
				keysId = append(keysId, key)
			}
		}
		this.dirty = true
		for _, v := range keysId {
			this.DelMessage(v, 1)
		}
	}

	FirePlayerWeekTimer(this)
}

func (this *Player) StatisticsOllen(t time.Time) {
	sec := int32(t.Sub(this.lastOllen) / time.Second)
	this.ollenSecs += sec
	this.lastOllen = t
	logger.Logger.Tracef("(this *Player) StatisticsOllen snid:%v ollen:%v totallen:%v", this.SnId, sec, this.ollenSecs)
}

func (this *Player) WriteOllenLog() {
	if this.ollenSecs > 60 {
		//		sec := this.ollenSecs
		//		this.ollenSecs = 0
		//		log := model.NewOLLenLog(this.SnId, int32(sec))
		//		if log != nil {
		//			OLLenLogChannelSington.Write(log)
		//		}
	}
}

func (this *Player) GetName() string {
	return this.Name
}

func (this *Player) setName(newName string) string {
	this.Name = newName
	return this.Name
}

func (this *Player) GetIP() string {
	return this.Ip
}

func (this *Player) CreateScene(sceneId, gameId, gameMode, sceneMode int, numOfGames int32, params []int32, dbGameFree *server_proto.DB_GameFree) (*Scene, hall_proto.OpResultCode_Game) {
	gs := GameSessMgrSington.GetMinLoadSess(gameId)
	if gs == nil {
		logger.Logger.Warnf("(this *Player) EnterScene %v, %v GameSessMgrSington.GetMinLoadSess() = nil ", this.SnId, gameId)
		return nil, hall_proto.OpResultCode_Game_OPRC_SceneServerMaintain_Game
	}

	s := SceneMgrSington.CreateScene(0, this.SnId, sceneId, gameId, gameMode, sceneMode, 1, numOfGames, params, gs, this.GetPlatform(), 0, dbGameFree, dbGameFree.GetId())
	if s == nil {
		logger.Logger.Tracef("(this *Player) EnterScene %v, SceneMgrSington.CreateScene() = nil ", this.SnId)
		return nil, hall_proto.OpResultCode_Game_OPRC_Error_Game
	}
	return s, hall_proto.OpResultCode_Game_OPRC_Sucess_Game
}

func (this *Player) CreateLocalGameScene(sceneId, gameId, gameSite, sceneMode, playerNum int, params []int32, dbGameFree *server_proto.DB_GameFree, baseScore int32) (*Scene, hall_proto.OpResultCode_Game) {
	gs := GameSessMgrSington.GetMinLoadSess(gameId)
	if gs == nil {
		logger.Logger.Warnf("(this *Player) CreateLocalGameScene %v, %v GameSessMgrSington.GetMinLoadSess() = nil ", this.SnId, gameId)
		return nil, hall_proto.OpResultCode_Game_OPRC_SceneServerMaintain_Game
	}

	s := SceneMgrSington.CreateLocalGameScene(this.SnId, sceneId, gameId, gameSite, sceneMode, 1, params, gs, this.GetPlatform(), playerNum, dbGameFree, baseScore, dbGameFree.GetId())
	if s == nil {
		logger.Logger.Tracef("(this *Player) EnterScene %v, SceneMgrSington.CreateScene() = nil ", this.SnId)
		return nil, hall_proto.OpResultCode_Game_OPRC_Error_Game
	}
	return s, hall_proto.OpResultCode_Game_OPRC_Sucess_Game
}

func (this *Player) EnterScene(s *Scene, ischangeroom bool, pos int) bool {
	if s == nil {
		logger.Logger.Tracef("(this *Player) EnterScene, s == nil %v", this.SnId)
		return false
	}

	if s != nil {
		this.applyPos = -1
		if s.PlayerEnter(this, pos, ischangeroom) {
			FirePlayerEnterScene(this, s)
			return true
		}
	}

	return false
}

func (this *Player) LeaveScene() bool {
	logger.Logger.Tracef("(this *Player) LeaveScene %v", this.SnId)
	s := this.scene
	if s == nil {
		return false
	}
	this.applyPos = -1
	if s.DelPlayer(this) {
		FirePlayerLeaveScene(this, s)
		return true
	}

	return false
}

func (this *Player) ReturnScene(isLoaded bool) *Scene {
	logger.Logger.Tracef("(this *Player) ReturnScene %v", this.SnId)
	if this.scene == nil {
		logger.Logger.Warnf("(this *Player) ReturnScene this.scene == nil snid:%d", this.SnId)
		return nil
	}
	if !this.scene.HasPlayer(this) && !this.scene.HasAudience(this) {
		logger.Logger.Warnf("(this *Player) ReturnScene !this.scene.HasPlayer(this) && !this.scene.HasAudience(this) snid:%d", this.SnId)
		return nil
	}

	pack := &server_proto.WGPlayerReturn{
		PlayerId: proto.Int32(this.SnId),
		IsLoaded: proto.Bool(isLoaded),
		RoomId:   proto.Int(this.scene.sceneId),
	}
	ctx := this.scene.GetPlayerGameCtx(this.SnId)
	if ctx != nil {
		pack.EnterTs = proto.Int64(ctx.enterTs)
	}
	proto.SetDefaults(pack)
	if this.SendToGame(int(server_proto.SSPacketID_PACKET_WG_PLAYERRETURN), pack) {
		//比赛场返场检查
		//MatchMgrSington.OnPlayerReturnScene(this.scene, this)
		return this.scene
	}
	//不应该这这里处理，因为 miniGame中小游戏玩家 player.Scene是不存在的。
	//FirePlayerReturnScene(this)

	return nil
}

func (this *Player) BackDiffData() {
	this.diffData.Coin = this.Coin
	this.diffData.SafeBoxCoin = this.SafeBoxCoin
}
func (this *Player) UpdateVip() {
	this.VIP = this.GetVIPLevel(this.CoinPayTotal)
	//clubManager.UpdateVip(this)
}

//// V卡判断要查map 单独判断
//func (this *Player) SendVCoinDiffData() {
//	var dirty bool
//	//V卡
//	pack := &player_proto.SCPlayerDataUpdate{}
//	pack.UpdateField = 0
//	if item := BagMgrSington.GetBagItemById(this.SnId, VCard); item != nil && this.diffData.VCoin != int64(item.ItemNum) {
//
//		dirty = true
//		pack.VCoin = proto.Int64(int64(item.ItemNum))
//		this.diffData.VCoin = int64(item.ItemNum)
//		pack.UpdateField += UpdateField_VCoin
//
//	}
//
//	if dirty {
//		this.SendToClient(int(player_proto.PlayerPacketID_PACKET_SC_PLAYERDATAUPDATE), pack)
//		logger.Logger.Trace("(this *Player) SendVCoinDiffData() ", pack)
//	}
//}

func (this *Player) SendDiffData() {
	this.UpdateVip()
	var dirty bool
	pack := &player_proto.SCPlayerDataUpdate{}
	pack.UpdateField = 0
	//金币
	if this.diffData.Coin != this.Coin {
		dirty = true
		pack.Coin = proto.Int64(this.Coin)
		this.diffData.Coin = this.Coin
		pack.UpdateField += UpdateField_Coin
		FriendMgrSington.UpdateFriendCoin(this.SnId, this.Coin)
	}
	//钻石
	if this.diffData.Diamond != this.Diamond {
		dirty = true
		pack.Diamond = proto.Int64(this.Diamond)
		this.diffData.Diamond = this.Diamond
		pack.UpdateField += UpdateField_Diamond
		FriendMgrSington.UpdateFriendDiamond(this.SnId, this.Diamond)
	}
	//保险箱金币
	if this.diffData.SafeBoxCoin != this.SafeBoxCoin {
		dirty = true
		pack.SafeBoxCoin = proto.Int64(this.SafeBoxCoin)
		this.diffData.SafeBoxCoin = this.SafeBoxCoin
		pack.UpdateField += UpdateField_SafeBoxCoin
	}
	//VIP等级
	if this.diffData.VIP != this.VIP {
		dirty = true
		pack.Vip = proto.Int32(this.VIP)
		this.diffData.VIP = this.VIP
		pack.UpdateField += UpdateField_VIP
	}
	//总充值金额
	if this.diffData.CoinPayTotal != this.CoinPayTotal {
		dirty = true
		pack.CoinPayTotal = proto.Int64(this.CoinPayTotal)
		this.diffData.CoinPayTotal = this.CoinPayTotal
		pack.UpdateField += UpdateField_CoinPayTotal
	}
	//流水差异
	if this.diffData.TotalConvertibleFlow != this.TotalConvertibleFlow {
		dirty = true
		pack.TotalConvertibleFlow = proto.Int64(this.TotalConvertibleFlow)
		this.diffData.TotalConvertibleFlow = this.TotalConvertibleFlow
		pack.UpdateField += UpdateField_TotalConvertibleFlow
	}
	//比赛报名券
	if this.diffData.Ticket != this.Ticket {
		dirty = true
		pack.Ticket = proto.Int64(this.Ticket)
		this.diffData.Ticket = this.Ticket
		pack.UpdateField += UpdateField_Ticket
	}
	//积分
	if this.diffData.Grade != this.Grade {
		dirty = true
		pack.Grade = proto.Int64(this.Grade)
		this.diffData.Grade = this.Grade
		pack.UpdateField += UpdateField_Grade
	}
	if dirty {
		this.SendToClient(int(player_proto.PlayerPacketID_PACKET_SC_PLAYERDATAUPDATE), pack)
		logger.Logger.Trace("(this *Player) SendDiffData() ", pack)
	}
}

func GetExchangeFlow(pd *model.PlayerData) int32 {
	if pd == nil {
		return 0
	}

	platform := PlatformMgrSington.GetPlatform(pd.Platform)
	if platform != nil {
		if (platform.ExchangeFlag & ExchangeFlag_Flow) != 0 {
			key, err := GetPromoterKey(pd.PromoterTree, pd.BeUnderAgentCode, pd.Channel)
			if err == nil {
				cfg := PromoterMgrSington.GetConfig(key)
				if cfg != nil {
					if (cfg.ExchangeFlag & ExchangeFlag_Flow) != 0 {
						return cfg.ExchangeFlow
					}
				}
			}
			return platform.ExchangeFlow
		}
	}
	return 0
}

func GetExchangeGiveFlow(pd *model.PlayerData) int32 {
	if pd == nil {
		return 0
	}

	platform := PlatformMgrSington.GetPlatform(pd.Platform)
	if platform != nil {
		if (platform.ExchangeFlag & ExchangeFlag_Flow) != 0 {
			key, err := GetPromoterKey(pd.PromoterTree, pd.BeUnderAgentCode, pd.Channel)
			if err == nil {
				cfg := PromoterMgrSington.GetConfig(key)
				if cfg != nil {
					if (cfg.ExchangeFlag & ExchangeFlag_Flow) != 0 {
						return cfg.ExchangeGiveFlow
					}
				}
			}
			return platform.ExchangeGiveFlow
		}
	}
	return 0
}

func GetExchangeForceTax(pd *model.PlayerData) int32 {
	if pd == nil {
		return 0
	}

	platform := PlatformMgrSington.GetPlatform(pd.Platform)
	if platform != nil {
		if (platform.ExchangeFlag & ExchangeFlag_Force) != 0 {
			key, err := GetPromoterKey(pd.PromoterTree, pd.BeUnderAgentCode, pd.Channel)
			if err == nil {
				cfg := PromoterMgrSington.GetConfig(key)
				if cfg != nil {
					if (cfg.ExchangeFlag & ExchangeFlag_Force) != 0 {
						return cfg.ExchangeForceTax
					}
				}
			}
			return platform.ExchangeForceTax
		}
	}
	return 0
}

func GetExchangeTax(pd *model.PlayerData) int32 {
	if pd == nil {
		return 0
	}

	platform := PlatformMgrSington.GetPlatform(pd.Platform)
	if platform != nil {
		if (platform.ExchangeFlag & ExchangeFlag_Tax) != 0 {
			key, err := GetPromoterKey(pd.PromoterTree, pd.BeUnderAgentCode, pd.Channel)
			if err == nil {
				cfg := PromoterMgrSington.GetConfig(key)
				if cfg != nil {
					if (cfg.ExchangeFlag & ExchangeFlag_Tax) != 0 {
						return cfg.ExchangeTax
					}
				}
			}
			return platform.ExchangeTax
		}
	}
	return 0
}

func GetExchangeFlag(pd *model.PlayerData) int32 {
	if pd == nil {
		return 0
	}

	platform := PlatformMgrSington.GetPlatform(pd.Platform)
	if platform != nil {
		if (platform.ExchangeFlag & ExchangeFlag_Flow) != 0 {
			key, err := GetPromoterKey(pd.PromoterTree, pd.BeUnderAgentCode, pd.Channel)
			if err == nil {
				cfg := PromoterMgrSington.GetConfig(key)
				if cfg != nil {
					if (cfg.ExchangeFlag & ExchangeFlag_Flow) != 0 {
						return 2
					}
				}
			}
			return 1
		}
	}
	return 0
}

func (this *Player) GetRegisterPrize() int32 {
	platform := this.GetPlatform()
	key, err := this.GetPromoterKey()
	var cfg *PromoterConfig
	if err == nil {
		cfg = PromoterMgrSington.GetConfig(key)
	}

	return GGetRegisterPrize(platform, cfg)
}

func GGetRegisterPrize(platform *Platform, cfg *PromoterConfig) int32 {
	if platform != nil {
		if cfg != nil {
			if (cfg.ExchangeFlag & ExchangeFlag_UpAcc) != 0 {
				return cfg.NewAccountGiveCoin
			}
		}
		return platform.NewAccountGiveCoin
	}
	return 0
}

func (this *Player) GetUpdateAccPrize() int32 {
	platform := this.GetPlatform()
	key, err := this.GetPromoterKey()
	var cfg *PromoterConfig
	if err == nil {
		cfg = PromoterMgrSington.GetConfig(key)
	}

	return GGetUpdateAccPrize(platform, cfg)
}

func GGetUpdateAccPrize(platform *Platform, cfg *PromoterConfig) int32 {
	if platform != nil {
		if cfg != nil {
			if (cfg.ExchangeFlag & ExchangeFlag_UpAcc) != 0 {
				return cfg.UpgradeAccountGiveCoin
			}
		}
		return platform.UpgradeAccountGiveCoin
	}
	return 0
}

func (this *Player) GetPromoterKey() (string, error) {
	return GetPromoterKey(this.PromoterTree, this.BeUnderAgentCode, this.Channel)
}

//计算流水可以兑换的值  返回 还需要多少流水 赠送扣除 强制费用 合计流水
//func GetExchangeFlowTotal(pd *model.PlayerData, playerTotalFlow int64, givesInfo []*model.CoinGiveLog) (int64,
//	int64, int64, int64) {
//	//可兑换的流水
//	var flow int64
//
//	var giveLostCoin int64  //赠送扣除
//	var forceTax int64      //强制费用
//	var needTotalFlow int64 //需要的流水
//	var lockCoin int64      //多少金额无法兑换，如果兑换需要强制扣除行政费用
//
//	retIds := []string{}
//	retAllIds := []string{}
//
//	exchangeFlow := GetExchangeFlow(pd)
//	exchangeGiveFlow := GetExchangeGiveFlow(pd)
//	exchangeForceTax := GetExchangeForceTax(pd)
//
//	if GetExchangeFlag(pd) > 0 {
//
//		//逐笔计算兑换流水金额，从上到下
//		curTotalFlow := playerTotalFlow
//
//		//按照时间排序
//		sort.Slice(givesInfo, func(i, j int) bool { return givesInfo[i].Ts > givesInfo[j].Ts })
//
//		for i := 0; i < len(givesInfo); i++ {
//			info := givesInfo[i]
//			if info.Ts > pd.LastExchangeTime {
//				retAllIds = append(retAllIds, info.LogId.Hex())
//				//计算是否通过稽核
//				needFlow := int64(0)
//				curTotalFlow += info.FLow
//				//如果是系统赠送的，需要全部扣除
//				if info.RecType == model.COINGIVETYPE_SYSTEM {
//					exchangeGiveFlowS := exchangeGiveFlow
//					t := ActMgrSington.GetExchangeFlow(pd.Platform, info.LogType)
//					if t != 0 {
//						exchangeGiveFlowS = t
//					}
//
//					if info.NeedGiveFlowRate > 0 {
//						exchangeGiveFlowS = info.NeedGiveFlowRate
//					}
//
//					needFlow = int64(math.Floor(float64(info.GiveCoin)*float64(exchangeGiveFlowS)/100)) * 100
//
//					if curTotalFlow < needFlow {
//						//需要扣除
//						giveLostCoin += info.GiveCoin
//						lockCoin += info.GiveCoin
//						flow += needFlow - curTotalFlow
//						curTotalFlow = 0
//					} else {
//						curTotalFlow -= needFlow
//						retIds = append(retIds, info.LogId.Hex())
//					}
//				} else {
//					exchangeGiveFlowS := exchangeGiveFlow
//					t := ActMgrSington.GetExchangeFlow(pd.Platform, info.LogType)
//					if t != 0 {
//						exchangeGiveFlowS = t
//					}
//					if info.NeedGiveFlowRate > 0 {
//						exchangeGiveFlowS = info.NeedGiveFlowRate
//					}
//
//					exchangePayFlowS := exchangeFlow
//					if info.NeedFlowRate > 0 {
//						exchangePayFlowS = info.NeedFlowRate
//					}
//
//					//分两部分扣除
//					needFlow = int64(math.Floor(float64(info.GiveCoin)*float64(exchangeGiveFlowS)/100)) * 100
//					needFlow += int64(math.Floor(float64(info.PayCoin)*float64(exchangePayFlowS)/100)) * 100
//					if curTotalFlow < needFlow {
//						//需要扣除
//						giveLostCoin += info.GiveCoin
//						//强制费用
//						forceTax += info.PayCoin * int64(exchangeForceTax) / 10000
//						lockCoin += info.GiveCoin
//						lockCoin += info.PayCoin
//						flow += needFlow - curTotalFlow
//						curTotalFlow = 0
//					} else {
//						curTotalFlow -= needFlow
//						retIds = append(retIds, info.LogId.Hex())
//					}
//				}
//				needTotalFlow += needFlow
//			}
//		}
//	}
//
//	return flow, giveLostCoin, forceTax, needTotalFlow
//}
//
////pageNo 1开始
//func GetExchangeFlowTotalPacket(playerTotalFlow int64, givesInfo []*model.CoinGiveLog, pd *model.PlayerData,
//	pageNo, pageNum int32, isCheck bool) *shop_proto.SCGetPlayerPayFlowList {
//	pack := &shop_proto.SCGetPlayerPayFlowList{}
//	var giveLostCoin int64  //赠送扣除
//	var forceTax int64      //强制费用
//	var needTotalFlow int64 //需要的流水
//	startIndex := (pageNo - 1) * pageNum
//	endIndex := pageNo * pageNum
//	//platform := this.GetPlatform()
//	exchangeFlow := GetExchangeFlow(pd)
//	exchangeGiveFlow := GetExchangeGiveFlow(pd)
//	exchangeForceTax := GetExchangeForceTax(pd)
//
//	//按照时间排序
//	sort.Slice(givesInfo, func(i, j int) bool { return givesInfo[i].Ts > givesInfo[j].Ts })
//	//逐笔计算兑换流水金额，从上到下p
//	curTotalFlow := playerTotalFlow
//
//	index := int32(0)
//
//	if GetExchangeFlag(pd) > 0 {
//		for i := 0; i < len(givesInfo); i++ {
//			info := givesInfo[i]
//			if isCheck || info.Ts > pd.LastExchangeTime {
//				tInfo := &shop_proto.PlayerPayFlowLog{}
//				tInfo.Ts = proto.Int64(info.Ts)
//				tInfo.PayType = proto.Int32(info.RecType)
//				tInfo.PayCoin = proto.Int64(info.PayCoin)
//				tInfo.GiveCoin = proto.Int64(info.GiveCoin)
//				tInfo.FinishFlow = proto.Int64(info.FLow)
//				tInfo.OrderID = proto.String(info.LogId.Hex())
//				//计算是否通过稽核
//				needFlow := int64(0)
//				isPass := int32(0)
//				curTotalFlow += info.FLow
//				//如果是系统赠送的，需要全部扣除
//				if info.RecType == model.COINGIVETYPE_SYSTEM {
//					exchangeGiveFlowS := exchangeGiveFlow
//					t := ActMgrSington.GetExchangeFlow(pd.Platform, info.LogType)
//					if t != 0 {
//						exchangeGiveFlowS = t
//					}
//					if info.NeedGiveFlowRate > 0 {
//						exchangeGiveFlowS = info.NeedGiveFlowRate
//					}
//					needFlow = int64(math.Floor(float64(info.GiveCoin)*float64(exchangeGiveFlowS)/100)) * 100
//					tInfo.GiveNeedFlow = proto.Int64(needFlow)
//					if curTotalFlow < needFlow {
//						//需要扣除
//						giveLostCoin += info.GiveCoin
//						tInfo.ForceGiveCoin = proto.Int64(info.GiveCoin)
//
//						curTotalFlow = 0
//					} else {
//						curTotalFlow -= needFlow
//						isPass = 1
//					}
//				} else {
//					exchangeGiveFlowS := exchangeGiveFlow
//					t := ActMgrSington.GetExchangeFlow(pd.Platform, info.LogType)
//					if t != 0 {
//						exchangeGiveFlowS = t
//					}
//
//					if info.NeedGiveFlowRate > 0 {
//						exchangeGiveFlowS = info.NeedGiveFlowRate
//					}
//					exchangePayFlowS := exchangeFlow
//					if info.NeedFlowRate > 0 {
//						exchangePayFlowS = info.NeedFlowRate
//					}
//					//分两部分扣除
//					needFlow = int64(math.Floor(float64(info.GiveCoin)*float64(exchangeGiveFlowS)/100)) * 100
//					tInfo.GiveNeedFlow = proto.Int64(needFlow)
//
//					payNeedFlow := int64(math.Floor(float64(info.PayCoin)*float64(exchangePayFlowS)/100)) * 100
//					tInfo.PayNeedFlow = proto.Int64(payNeedFlow)
//					needFlow += payNeedFlow
//
//					if curTotalFlow < needFlow {
//						//需要扣除
//						giveLostCoin += info.GiveCoin
//						tInfo.ForceGiveCoin = proto.Int64(info.GiveCoin)
//
//						//强制费用
//						forceTax += info.PayCoin * int64(exchangeForceTax) / 10000
//						tInfo.ForceTax = proto.Int64(info.PayCoin * int64(exchangeForceTax) / 10000)
//						curTotalFlow = 0
//					} else {
//						curTotalFlow -= needFlow
//						isPass = 1
//					}
//				}
//
//				tInfo.IsPass = proto.Int32(isPass)
//				needTotalFlow += needFlow
//
//				if index >= startIndex && index < endIndex {
//					pack.Data = append(pack.Data, tInfo)
//				}
//
//				index += 1
//			}
//		}
//	}
//	pack.PageNo = proto.Int32(int32(pageNo))
//	pack.PageSum = proto.Int32(int32(math.Ceil(float64(index) / float64(pageNum))))
//	pack.PageSize = proto.Int32(pageNum)
//	pack.TotalNum = proto.Int32(index)
//	proto.SetDefaults(pack)
//	return pack
//}

func (this *Player) SendPlayerInfo() {
	this.UpdateVip()
	scPlayerData := &player_proto.SCPlayerData{
		OpRetCode: player_proto.OpResultCode_OPRC_Sucess,
		Data: &player_proto.PlayerData{
			AccId:         proto.String(this.AccountId),        //账号ID
			Platform:      proto.String(this.Platform),         //平台
			Channel:       proto.String(this.Channel),          //渠道
			Promoter:      proto.String(this.BeUnderAgentCode), //推广员
			Name:          proto.String(this.Name),             //名字
			SnId:          proto.Int32(this.SnId),              //数字账号
			Head:          proto.Int32(this.Head),              //头像
			Sex:           proto.Int32(this.Sex),               //性别
			GMLevel:       proto.Int32(this.GMLevel),           //GM等级
			Coin:          proto.Int64(this.Coin),              //金币
			SpecailFlag:   proto.Int32(int32(this.Flags)),      //特殊标记
			Tel:           proto.String(this.Tel),              //手机号码
			InviterId:     proto.Int32(this.InviterId),         //邀请人ID
			SafeBoxCoin:   proto.Int64(this.SafeBoxCoin),       //保险箱金币
			VIP:           proto.Int32(this.VIP),               //VIP帐号
			AlipayAccount: proto.String(this.AlipayAccount),    //支付宝账号
			AlipayAccName: proto.String(this.AlipayAccName),    //支付宝实名
			Bank:          proto.String(this.Bank),             //银行
			BankAccount:   proto.String(this.BankAccount),      //银行帐号
			BankAccName:   proto.String(this.BankAccName),      //银行开户名
			HeadOutLine:   proto.Int32(this.HeadOutLine),       //头像框
			CoinPayTotal:  proto.Int64(this.CoinPayTotal),      //总充值金额
			CreateTs:      proto.Int64(this.CreateTime.Unix()), //角色创建时间
			//ClubCoin:      proto.Int64(this.ClubCoin),          //俱乐部金币
			Ticket:  proto.Int64(this.Ticket),  //比赛入场券
			Grade:   proto.Int64(this.Grade),   //积分
			Diamond: proto.Int64(this.Diamond), //钻石
			HeadUrl: proto.String(this.HeadUrl),
		},
	}
	if this.Roles != nil {
		scPlayerData.Data.UseRoleId = this.Roles.ModId
	}
	if this.Pets != nil {
		scPlayerData.Data.UsePetId = this.Pets.ModId
	}
	if this.WelfData != nil {
		scPlayerData.Data.ReliefFundTimes = this.WelfData.ReliefFundTimes
	}
	if item := BagMgrSington.GetBagItemById(this.SnId, VCard); item != nil {
		scPlayerData.Data.VCoin = int64(item.ItemNum) //V卡

	}
	raw := fmt.Sprintf("%v%v", model.DEFAULT_PLAYER_SAFEBOX_PWD, common.GetAppId())
	h := md5.New()
	io.WriteString(h, raw)
	pwd := hex.EncodeToString(h.Sum(nil))

	if this.SafeBoxPassword != pwd {
		scPlayerData.Data.SafeBoxIsExist = proto.Int32(1)
	} else {
		scPlayerData.Data.SafeBoxIsExist = proto.Int32(0)
	}
	scene := SceneMgrSington.GetSceneByPlayerId(this.SnId)
	if scene != nil {
		scPlayerData.RoomId = proto.Int(scene.sceneId)
		scPlayerData.GameId = proto.Int(scene.gameId)
		//增加gameFreeId
		scPlayerData.LogicId = scene.dbGameFree.Id
	}
	platform := PlatformMgrSington.GetPlatform(this.Platform)
	if platform != nil {
		scPlayerData.BindOption = proto.Int32(platform.BindOption)
	} else {
		scPlayerData.BindOption = proto.Int32(7)
	}

	////gs添加 增加小游戏对应的 gameId roomId
	//gameingScenes := MiniGameMgrSington.GetAllSceneByPlayer(this)
	//			////scPlayerData.MiniGameArr = &player_proto.MiniGameInfo{}
	//for _, s := range gameingScenes {
	//	miniItem := &player_proto.MiniGameInfo{
	//		GameId:  int32(s.gameId),
	//		LogicId: s.dbGameFree.Id,
	//		RoomId:  int32(s.sceneId),
	//	}
	//	scPlayerData.MiniGameArr = append(scPlayerData.MiniGameArr, miniItem)
	//}
	//////////////////////////////////////////////

	//if platform != nil {
	//	if this.LogonMarker == nil {
	//		this.LogonMarker = make(map[string]int)
	//	}
	//	//新流水数值计算
	//	if m, ok := this.LogonMarker["flow"]; ok {
	//		//流水标记有数值
	//		if m == 2 && GetExchangeFlag(this.PlayerData) <= 0 {
	//			//兑换模式
	//			this.LogonMarker["flow"] = 1
	//			//this.TotalConvertibleFlow = 0
	//			this.ExchangeTotal = 0
	//			this.LastFlowTime = time.Now()
	//			//this.CanExchangeBeforeRecharge = 0
	//		} else if m == 1 && GetExchangeFlag(this.PlayerData) > 0 {
	//			//流水模式
	//			this.LogonMarker["flow"] = 2
	//			//this.TotalConvertibleFlow = 0
	//			this.ExchangeTotal = 0
	//			this.LastFlowTime = time.Now()
	//			//this.CanExchangeBeforeRecharge = 0
	//		}
	//	} else {
	//		//流水标记没有数值
	//		if GetExchangeFlag(this.PlayerData) <= 0 {
	//			//兑换模式
	//			this.LogonMarker["flow"] = 1
	//			//this.TotalConvertibleFlow = 0
	//			this.ExchangeTotal = 0
	//			this.LastFlowTime = time.Now()
	//			//this.CanExchangeBeforeRecharge = 0
	//		} else {
	//			//流水模式
	//			this.LogonMarker["flow"] = 2
	//			//this.TotalConvertibleFlow = 0
	//			this.ExchangeTotal = 0
	//			this.LastFlowTime = time.Now()
	//			//this.CanExchangeBeforeRecharge = 0
	//		}
	//	}
	//	//流水值
	//	scPlayerData.Data.TotalConvertibleFlow = proto.Int64(this.TotalConvertibleFlow)
	//
	//}
	proto.SetDefaults(scPlayerData)
	logger.Logger.Tracef("model.GameParamData.NewPlayerCoin %v", model.GameParamData.NewPlayerCoin)
	logger.Logger.Tracef("Send SCPlayerData %v", scPlayerData)
	this.SendToClient(int(player_proto.PlayerPacketID_PACKET_SC_PLAYERDATA), scPlayerData)
	if !this.IsRob {
		this.SyncPlayerDataToGateSrv(this.PlayerData)
	}
	//this.SendJackpotInfo()
}

func (this *Player) SendJackpotInfo() {
	//通知所有的gamesrv向玩家发送奖池信息
	if this.gateSess != nil {
		var gateSid int64
		if srvInfo, ok := this.gateSess.GetAttribute(srvlib.SessionAttributeServerInfo).(*srvlibproto.SSSrvRegiste); ok && srvInfo != nil {
			sessionId := srvlib.NewSessionIdEx(srvInfo.GetAreaId(), srvInfo.GetType(), srvInfo.GetId(), 0)
			gateSid = sessionId.Get()
		}

		//查找当前平台下所以开放的游戏id
		info := make([]*server_proto.GameInfo, 0)
		gps := PlatformMgrSington.GetPlatformGameConfig(this.Platform)
		for _, v := range gps {
			if v.Status {
				if v.DbGameFree.GetGameRule() != 0 {
					//lgi := &server_proto.GameInfo{
					//	GameId:     proto.Int32(v.DbGameFree.GetGameId()),
					//	GameFreeId: proto.Int32(v.DbGameFree.GetId()),
					//	GameType:   proto.Int32(v.DbGameFree.GetGameType()),
					//}
					info = append(info, &server_proto.GameInfo{
						GameId:     proto.Int32(v.DbGameFree.GetGameId()),
						GameFreeId: proto.Int32(v.DbGameFree.GetId()),
						GameType:   proto.Int32(v.DbGameFree.GetGameType()),
					})
				}
			}
		}

		servers := GameSessMgrSington.GetAllGameSess()
		for _, v := range servers {
			pack := &server_proto.WGGameJackpot{
				Sid:      this.sid,
				GateSid:  gateSid,
				Platform: this.Platform,
				Info:     info,
			}
			v.Send(int(server_proto.SSPacketID_PACKET_WG_GAMEJACKPOT), pack)
		}
	}
}

func (this *Player) IsGM() bool {
	if this.GMLevel > 0 {
		return true
	}
	return false
}

func (this *Player) IsAgentor() bool {
	return false
}

func (this *Player) IsPlayer() bool {
	return true
}

func (this *Player) HasAuthority(role int) bool {
	switch role {
	case common.ClientRole_Agentor:
		return this.IsAgentor()
	case common.ClientRole_GM:
		return this.IsGM()
	case common.ClientRole_Player:
		return true
	}
	return false
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

func (this *Player) ReportLoginEvent() {
	//用户登录
	if !this.IsRob {
		isBindPhone := int32(0)
		if this.Tel != "" {
			isBindPhone = 1
			if this.UpgradeTime.IsZero() {
				this.UpgradeTime = this.CreateTime
			}
		}
		LogChannelSington.WriteMQData(model.GenerateLogin(model.CreatePlayerLoginEvent(this.SnId,
			this.Channel, this.BeUnderAgentCode, this.Platform, this.City, this.DeviceOS, this.Ip,
			this.CreateTime, this.UpgradeTime, isBindPhone, this.TelephonePromoter, this.DeviceId)))
		//登录通知
		ActMonitorMgrSington.SendActMonitorEvent(ActState_Login, this.SnId, this.Name, this.Platform,
			0, 0, "", 0)
	}
}

func (this *Player) ReportBindPhoneEvent() {
	//升级账号事件
	if !this.IsRob {
		//LogChannelSington.WriteMQData(model.GenerateBindEvent(model.CreatePlayerBindPhoneEvent(
		//	this.SnId, this.Channel, this.BeUnderAgentCode, this.Platform, this.City, this.DeviceOS,
		//	this.CreateTime, this.TelephonePromoter)))
	}
}

func (this *Player) ReportBindAlipayEvent() {
	//绑定支付宝事件
	//if !this.IsRob {
	//	d, e := model.MarshalPlayerBindAlipayEvent(2, this.SnId, this.Channel, this.BeUnderAgentCode,
	//		this.Platform, this.City, this.DeviceOS, this.TelephonePromoter)
	//	if e == nil {
	//		rmd := model.NewInfluxDBData("hj.player_bind_alipay", d)
	//		if rmd != nil {
	//			InfluxDBDataChannelSington.Write(rmd)
	//		}
	//	}
	//}
}

func (this *Player) AddCoinGiveLog(payCoin, giveCoin int64, coinType, logType, recType int32, remark, oper string) {
	plt := PlatformMgrSington.GetPlatform(this.Platform)
	curVer := int32(0)
	if plt != nil {
		curVer = plt.ExchangeVer
	}

	log := model.NewCoinGiveLogEx(this.SnId, this.Name, payCoin, giveCoin, coinType, logType, this.PromoterTree,
		recType, curVer, this.Platform, this.Channel, this.BeUnderAgentCode, remark, oper, this.PackageID, 0, 0)
	if log != nil {
		task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
			err := model.InsertGiveCoinLog(log)
			if err == nil {
				if this.LastExchangeOrder != "" && this.TotalConvertibleFlow > 0 {
					err = model.UpdateGiveCoinLastFlow(log.Platform, this.LastExchangeOrder, this.TotalConvertibleFlow)
				}
			}
			return err
		}), task.CompleteNotifyWrapper(func(ud interface{}, t task.Task) {
			if ud == nil {
				//清空流水，更新id
				this.TotalConvertibleFlow = 0
				this.LastExchangeOrder = log.LogId.Hex()
			}
		}), "UpdateGiveCoinLastFlow").StartByGroupFixExecutor(log.Platform, "UpdateGiveCoinLastFlow")
	}
}

func ReportSystemGiveEvent(pd *model.PlayerData, amount, tag int32, notifyClient bool) {
	//系统赠送
	if !pd.IsRob {
		LogChannelSington.WriteMQData(model.GenerateSystemGive(pd.SnId, pd.Platform, pd.Channel, pd.BeUnderAgentCode, amount, tag, common.GetAppId()))

		//插入本地表
		if amount > 0 {
			if ActMgrSington.GetIsNeedGive(pd.Platform, tag) {
				plt := PlatformMgrSington.GetPlatform(pd.Platform)
				curVer := int32(0)
				if plt != nil {
					curVer = plt.ExchangeVer
				}

				log := model.NewCoinGiveLogEx(pd.SnId, pd.Name, 0, int64(amount), 0, tag, pd.PromoterTree,
					model.COINGIVETYPE_SYSTEM, curVer, pd.Platform, pd.Channel, pd.BeUnderAgentCode, "",
					"system", pd.PackageID, 0, 0)
				if log != nil {
					task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
						err := model.InsertGiveCoinLog(log)
						if err == nil {
							if pd.LastExchangeOrder != "" && pd.TotalConvertibleFlow > 0 {
								err = model.UpdateGiveCoinLastFlow(log.Platform, pd.LastExchangeOrder, pd.TotalConvertibleFlow)
							}
						}
						return err
					}), task.CompleteNotifyWrapper(func(ud interface{}, t task.Task) {
						if ud == nil {
							//清空流水，更新id
							pd.TotalConvertibleFlow = 0
							pd.LastExchangeOrder = log.LogId.Hex()
							player := PlayerMgrSington.GetPlayerBySnId(pd.SnId)
							if player == nil {
								//需要回写数据库
								task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
									model.UpdatePlayerExchageFlowAndOrder(pd.Platform, pd.SnId, 0, pd.LastExchangeOrder)
									return nil
								}), nil, "UpdatePlayerExchageFlowAndOrder").StartByGroupFixExecutor(log.Platform, pd.AccountId)
							}
						}
					}), "UpdateGiveCoinLastFlow").StartByGroupFixExecutor(log.Platform, "UpdateGiveCoinLastFlow")
				}
			}

			if notifyClient {
				player := PlayerMgrSington.GetPlayerBySnId(pd.SnId)
				if player != nil {
					//通知客户端
					sendPack := &shop_proto.SCNotifyGiveCoinInfo{
						GiveCoin: proto.Int64(int64(amount)),
						GiveTag:  proto.Int32(tag),
					}

					proto.SetDefaults(sendPack)
					player.SendToClient(int(shop_proto.SPacketID_SHOP_SC_GIVECOIN_INFO), sendPack)
				}
			}
		}
	}
}

func (this *Player) ReportSystemGiveEvent(amount, tag int32, notifyClient bool) {
	ReportSystemGiveEvent(this.PlayerData, amount, tag, notifyClient)
}

// 破产事件
func (this *Player) ReportBankRuptcy(gameid, gamemode, gamefreeid int32) {
	//if !this.IsRob {
	//	d, e := model.MarshalBankruptcyEvent(2, this.SnId, this.TelephonePromoter, this.Channel, this.BeUnderAgentCode,
	//		this.Platform, this.City, this.CreateTime, gameid, gamemode, gamefreeid)
	//	if e == nil {
	//		rmd := model.NewInfluxDBData("hj.player_bankruptcy", d)
	//		if rmd != nil {
	//			InfluxDBDataChannelSington.Write(rmd)
	//		}
	//	}
	//}
}

func (this *Player) CheckType(gameid, gamefreeId int32) *server_proto.DB_PlayerType {
	types := srvdata.PlayerTypeMgrSington.GetPlayerType(gamefreeId)
	cnt := len(types)
	if cnt > 0 {
		var pgs model.PlayerGameStatics
		if this.GDatas != nil {
			if d, exist := this.GDatas[strconv.Itoa(int(gameid))]; exist {
				pgs = d.Statics
			}
		}

		//赔率 产出/投入 万分比
		odds := int64(float64(float64(pgs.TotalOut+1)/float64(pgs.TotalIn+1)) * 10000)
		if odds > 10000000 {
			odds = 10000000
		}
		for i := 0; i < cnt; i++ {
			t := types[i]
			logger.Logger.Warn("Player CheckType 0  ", this.CoinPayTotal, t.GetPayLowerLimit(), t.GetPayUpperLimit(), pgs.GameTimes,
				t.GetGameTimeLowerLimit(), t.GetGameTimeUpperLimit(), pgs.TotalIn, t.GetTotalInLowerLimit(), t.GetTotalInUpperLimit(),
				odds, t.GetOddsLowerLimit(), t.GetOddsUpperLimit())
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

// 线程不安全，避免异步任务调用
func (this *Player) GetPlatform() *Platform {
	platform := PlatformMgrSington.GetPlatform(this.Platform)
	if platform != nil && platform.Isolated {
		return platform
	}
	return PlatformMgrSington.GetPlatform(Default_Platform)
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
		this.VIP = this.GetVIPLevel(int64(takeCoin))
		//todo 随机一个概率，提高vip等级
		if common.RandInt(10000) < model.NormalParamData.RobotVipChangeRate {
			this.VIP += int32(math.Ceil(rand.ExpFloat64() * 10000))
		}

		if this.VIP > this.GetMaxVIPLevel() {
			this.VIP = this.GetMaxVIPLevel()
		}

		if this.scene != nil {
			if !this.scene.IsTestScene() {
				if this.VIP <= 1 {
					this.VIP = 1 + rand.Int31n(2)
				}
			} else {
				if this.VIP <= 1 {
					this.VIP = rand.Int31n(2)
				}
			}
		}
		this.RobRandVip()
	}
}

func (this *Player) RobRandVipWhenEnterMatch(vipLimit int32) {
	var supportVip []int
	for i := 0; vipLimit != 0; i++ {
		if vipLimit&1 != 0 {
			supportVip = append(supportVip, i)
		}
		vipLimit = vipLimit >> 1
	}
	var vip int
	if len(supportVip) > 0 {
		vip = supportVip[rand.Intn(len(supportVip))]
	}
	this.VIP = int32(vip)
}

func (this *Player) PlayerMacAbnormal() bool {
	return false //s.HasSameIp(p.Ip) || s.HasSameMac() || s.HasSameTel() || s.HasSamePostion()
}

// 这个冲账记录不可随便写,需要加该日志时，请找lyk确认,暂定依据是金币直接加到身上，不依赖其他数据状态的可以写该日志
// 业务准则:先更新标记，再写冲账记录
func (this *Player) AddPayCoinLog(coin int64, coinType int32, oper string) {
	ts := time.Now().UnixNano()
	billNo := ts
	log := model.NewPayCoinLog(billNo, this.SnId, coin, 0, oper, coinType, 0)
	if log != nil {
		ts = log.TimeStamp
		task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
			err := model.InsertPayCoinLog(this.Platform, log)
			if err != nil {
				logger.Logger.Errorf("InsertPayCoinLog err:%v log:%v", err, log)
				return err
			}
			return err
		}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
			switch coinType {
			case model.PayCoinLogType_Coin:
				if this.CoinPayTs < ts {
					this.CoinPayTs = ts
					this.dirty = true
				}
			case model.PayCoinLogType_SafeBoxCoin:
				if this.SafeBoxCoinTs < ts {
					this.SafeBoxCoinTs = ts
					this.dirty = true
				}
			}
		}), "InsertPayCoinLog").StartByFixExecutor("InsertPayCoinLog")
	}
}

// 充值回调
func (this *Player) SendPlayerRechargeAnswer(coin int64) {
	if this.Tel == "" {
		pack := &player_proto.SCPlayerRechargeAnswer{
			OpParam:     proto.Int64(1),
			AddCoin:     proto.Int64(coin),
			Coin:        proto.Int64(this.Coin),
			SafeBoxCoin: proto.Int64(this.SafeBoxCoin),
		}
		proto.SetDefaults(pack)
		this.SendToClient(int(player_proto.PlayerPacketID_PACKET_SC_PLAYERRECHARGEANSWER), pack)
	}
}

//
//// 在线奖励: 重置, 清零在线时长及奖励领取信息
//func (this *Player) OnlineRewardReset() {
//	this.PlayerData.OnlineRewardData.OnlineDuration = 0
//	this.PlayerData.OnlineRewardData.RewardReceived = 0
//	this.PlayerData.OnlineRewardData.Ts = time.Now().Unix()
//	this.dirty = true
//}
//
//// 在线奖励: （Logout时）累计在线时长
//func (this *Player) OnlineRewardAddUpOnlineDuration() {
//	if this.state != PlayerState_Online {
//		return
//	}
//
//	tNow := time.Now()
//	inSameDay := common.InSameDay(tNow, time.Unix(this.PlayerData.OnlineRewardData.Ts, 0))
//	if inSameDay {
//		this.PlayerData.OnlineRewardData.OnlineDuration += uint32(tNow.Unix() - this.PlayerData.OnlineRewardData.Ts)
//	} else {
//		this.PlayerData.OnlineRewardData.OnlineDuration = uint32(tNow.Unix() - now.New(tNow).BeginningOfDay().Unix())
//	}
//
//	this.PlayerData.OnlineRewardData.Ts = tNow.Unix()
//	this.dirty = true
//}
//
//// 在线奖励: (实时)获取在线时长
//func (this *Player) OnlineRewardGetOnlineDuration() uint32 {
//	this.OnlineRewardAddUpOnlineDuration()
//	return this.PlayerData.OnlineRewardData.OnlineDuration
//}

// 幸运转盘
//func (this *Player) LuckyTurntableSwitchScore(continuous bool) {
//	if continuous {
//		this.PlayerData.LuckyTurnTableData.Score = this.PlayerData.LuckyTurnTableData.TomorrowScore +
//			int64(this.PlayerData.LuckyTurnTableData.TomorrowFloatScore/100)
//	} else {
//		this.PlayerData.LuckyTurnTableData.Score = 0
//	}
//
//	this.PlayerData.LuckyTurnTableData.TomorrowScore = 0
//	this.PlayerData.LuckyTurnTableData.TomorrowFloatScore = 0
//	this.dirty = true
//}

func (this *Player) SyncSafeBoxCoinToGame() {
	pack := &server_proto.WGSyncPlayerSafeBoxCoin{
		SnId:        proto.Int32(this.SnId),
		SafeBoxCoin: proto.Int64(this.SafeBoxCoin),
	}
	proto.SetDefaults(pack)
	this.SendToGame(int(server_proto.SSPacketID_PACKET_WG_SyncPlayerSafeBoxCoin), pack)
}

//func (this *Player) GetDgHboPlayerName(plt *Platform) (string, string) {
//	if plt == nil {
//		return "", ""
//	}
//	if plt.DgHboConfig == 0 {
//		return this.DgGame, this.DgPass
//	} else if plt.DgHboConfig == 1 {
//		return this.StoreDgGame, this.StoreDgPass
//	} else if plt.DgHboConfig == 2 {
//		return this.StoreHboGame, this.StoreHboPass
//	}
//	return "", ""
//}
//
//func (this *Player) SetDgHboPlayerName(plt *Platform, name, pass string) {
//	if plt == nil {
//		return
//	}
//
//	if plt.DgHboConfig == 0 {
//		this.DgGame = name
//		this.DgPass = pass
//		if strings.Contains(name, "dg") {
//			this.StoreDgGame = name
//			this.StoreDgPass = pass
//		} else {
//			this.StoreHboGame = name
//			this.StoreHboPass = pass
//		}
//	} else if plt.DgHboConfig == 1 {
//		this.StoreDgGame = name
//		this.StoreDgPass = pass
//	} else if plt.DgHboConfig == 2 {
//		this.StoreHboGame = name
//		this.StoreHboPass = pass
//	}
//
//}

func (this *Player) AddCoinPayTotal(coin int64) {
	this.CoinPayTotal += coin
}

// 当用户充值
func OnPlayerPay(pd *model.PlayerData, coin int64) {
	if pd == nil {
		return
	}

	buf, err := pd.GetPlayerDataEncoder()
	if err == nil {
		pack := &server_proto.WTPlayerPay{
			AddCoin:    proto.Int64(coin),
			PlayerData: buf.Bytes(),
		}
		proto.SetDefaults(pack)
		common.SendToActThrSrv(int(server_proto.SSPacketID_PACKET_WT_PLAYERPAY), pack)
	}

	//ActFPayMgrSington.OnPlayerPay(pd.SnId, pd.Platform, coin)
}

func (this *Player) SendPlatformCanUsePromoterBind() {
	state := int32(0)
	plt := PlatformMgrSington.GetPlatform(this.Platform)
	if plt != nil {
		if plt.IsCanUserBindPromoter {
			state = 1
			if this.BeUnderAgentCode != "" && this.BeUnderAgentCode != "0" {
				state = 2
			}
		}
	}

	pack := &player_proto.SCBindPromoterState{
		BindState: proto.Int32(state),
	}

	proto.SetDefaults(pack)
	this.SendToClient(int(player_proto.PlayerPacketID_PACKET_SC_BINDPROMOTERSTATE), pack)
}

func (this *Player) RedirectByGame(packetid int, rawpack interface{}) bool {
	if this.scene == nil || this.scene.gameSess == nil || this.scene.gameSess.Session == nil {
		logger.Logger.Tracef("[%v] sess == nil ", this.Name)
		return false
	}
	if rawpack == nil {
		logger.Logger.Trace(" rawpack == nil ")
		return false
	}

	data, err := netlib.MarshalPacket(packetid, rawpack)
	if err == nil {
		pack := &server_proto.SSRedirectToPlayer{
			SnId:     proto.Int32(this.SnId),
			PacketId: proto.Int(packetid),
			Data:     data,
		}
		proto.SetDefaults(pack)
		return this.SendToGame(int(server_proto.SSPacketID_PACKET_SS_REDIRECTTOPLAYER), pack)
	}
	return false
}

//	func (this *Player) GetMatchAr(matchId int32) *model.MatchAchievement {
//		if this.MatchAchievement == nil {
//			this.MatchAchievement = make(map[int32]*model.MatchAchievement)
//		}
//		if ar, exist := this.MatchAchievement[matchId]; exist {
//			return ar
//		}
//		ar := &model.MatchAchievement{MatchId: matchId, CreateTs: int32(time.Now().Unix())}
//		this.MatchAchievement[matchId] = ar
//		return ar
//	}
func (this *Player) InitLayered() {
	logger.Logger.Trace("(this *Player) InitLayered")
	if this.IsRob {
		return
	}
	this.layered = make(map[int]bool)
	task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
		return LogicLevelMgrSington.SendPostBySnIds(this.Platform, []int32{this.SnId})
	}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
		if data != nil {
			newData := data.([]NewMsg)
			for _, d := range newData {
				cnf := LogicLevelMgrSington.GetConfig(d.Platform)
				if cnf == nil {
					continue
				}
				this.layerlevels = d.Levels
				for _, v := range d.Levels {
					if td, ok := cnf.LogicLevelInfo[int32(v)]; ok {
						if td.StartAct == 1 {
							for _, id := range td.CheckActIds {
								this.layered[int(id)] = true
							}
						}
					}
				}
			}
		}
		//this.ActStateSend2Client()
	}), "SendPostBySnIds").Start()
	////
	for k, v := range this.layered {
		logger.Logger.Tracef("InitLayered....id=%v   %v", k, v)
	}

}

func (this *Player) ActStateSend2Client() {
	logger.Logger.Trace("(this *Player) ActStateSend2Client")
	actSwitchCfg := make([]int32, common.ActId_Max)
	for i := 0; i < common.ActId_Max; i++ {
		isOpen := true
		switch i {
		//case common.ActId_OnlineReward:
		//	cfg, ok := ActOnlineRewardMgrSington.Configs[this.Platform]
		//	if !ok || cfg.StartAct <= 0 {
		//		isOpen = false
		//	}
		//case common.ActId_GoldTask:
		//	plateFormConfig := ActGoldTaskMgrSington.GetPlateFormConfig(this.Platform)
		//	if plateFormConfig == nil || plateFormConfig.StartAct == 0 {
		//		isOpen = false
		//	}
		//case common.ActId_GoldCome:
		//	plateFormConfig := ActGoldComeMgrSington.GetPlateFormConfig(this.Platform)
		//	if plateFormConfig == nil || plateFormConfig.StartAct == 0 {
		//		isOpen = false
		//	}
		//case common.ActId_LuckyTurntable:
		//	cfg, ok := ActLuckyTurntableMgrSington.Configs[this.Platform]
		//	if !ok || cfg == nil || cfg.StartAct <= 0 {
		//		isOpen = false
		//	}
		//case common.ActId_Yeb:
		//	cfg, ok := ActYebMgrSington.Configs[this.Platform]
		//	if !ok || cfg == nil || cfg.StartAct <= 0 {
		//		isOpen = false
		//	}
		//case common.ActId_Card:
		//	cfg, ok := ActCardMgrSington.Configs[this.Platform]
		//	if !ok || cfg == nil || cfg.StartAct <= 0 {
		//		isOpen = false
		//	}
		case common.ActId_RebateTask:
			rebateTask := RebateInfoMgrSington.rebateTask[this.Platform]
			if rebateTask != nil && !rebateTask.RebateSwitch {
				isOpen = false
			}
		//case common.ActId_VipLevelBonus:
		//	config := ActVipMgrSington.GetConfig(this.Platform)
		//	if config == nil || config.StartAct == 0 {
		//		isOpen = false
		//	}
		//case common.ActId_Sign:
		//	config := ActSignMgrSington.GetConfig(this.Platform)
		//	if config == nil || config.StartAct == 0 {
		//		isOpen = false
		//	}
		//case common.ActId_Task:
		//	config := ActTaskMgrSington.GetPlatformConfig(this.Platform)
		//	if config == nil || config.StartAct == 0 {
		//		isOpen = false
		//	}
		case common.ExchangeId_Bank:
			platform := PlatformMgrSington.GetPlatform(this.Platform)
			if platform == nil || !platform.IsMarkFlag(int32(Bind_BackCard)) {
				isOpen = false
			}
		case common.ExchangeId_Alipay:
			platform := PlatformMgrSington.GetPlatform(this.Platform)
			if platform == nil || !platform.IsMarkFlag(int32(Bind_Alipay)) {
				isOpen = false
			}
		case common.ExchangeId_Wechat:
			platform := PlatformMgrSington.GetPlatform(this.Platform)
			if platform == nil || !platform.IsMarkFlag(int32(Bind_WeiXin)) {
				isOpen = false
			}
		}

		if !isOpen {
			actSwitchCfg[i] = 1
			continue
		}

		open, ok := this.layered[i]
		if ok && open {
			actSwitchCfg[i] = 1
		}
	}
	pack := &login_proto.SCActSwitchCfg{
		ActSwitchCfg: actSwitchCfg, // 0开 1关 默认开
	}
	proto.SetDefaults(pack)
	logger.Logger.Trace("ActStateSend2Client: ", this.SnId, pack)
	this.SendToClient(int(login_proto.LoginPacketID_PACKET_SC_ACTSWITCHCFG), pack)
}

func (this *Player) ModifyActSwitch() {
	logger.Logger.Trace("====================ModifyActSwitch==================")

	//this.ActStateSend2Client()
}

// 玩家信息同步到网关
func (this *Player) SyncPlayerDataToGateSrv(pd *model.PlayerData) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(pd)
	if err != nil {
		logger.Logger.Info("(this *Player) UpdateToGateSrv gob.Marshal error", err)
	} else {
		pack := &server_proto.WRPlayerData{
			Sid:        this.sid,
			PlayerData: buf.Bytes(),
		}
		proto.SetDefaults(pack)
		this.gateSess.Send(int(server_proto.SSPacketID_PACKET_WR_PlayerData), pack)
	}
}

func (this *Player) OnPlayerGameGain(gain int64) {
	clearWB := false
	if this.WhiteLevel != 0 || this.BlackLevel != 0 {
		if gain > 0 {
			this.WBCoinTotalOut += gain
		} else {
			this.WBCoinTotalIn += -gain
		}
		if this.WBCoinLimit != 0 {
			if this.WhiteLevel != 0 && this.WBCoinTotalOut-this.WBCoinTotalIn >= this.WBCoinLimit { //自动解除白名单
				clearWB = true
			} else if this.BlackLevel != 0 && this.WBCoinTotalIn-this.WBCoinTotalOut >= this.WBCoinLimit { //自动解除黑名单
				clearWB = true
			}
		}

		if this.WBMaxNum > 0 {
			if this.WhiteLevel > 0 && gain > 0 {
				this.WBMaxNum -= 100
				if this.WBMaxNum <= 0 {
					clearWB = true
				}
			} else if this.BlackLevel > 0 && gain < 0 {
				this.WBMaxNum -= 100
				if this.WBMaxNum <= 0 {
					clearWB = true
				}
			}
		}
	}

	if clearWB { //自动解除黑白名单
		this.WhiteLevel = 0
		this.BlackLevel = 0
		this.WBCoinTotalIn = 0
		this.WBCoinTotalOut = 0
		this.WBCoinLimit = 0
		this.WBMaxNum = 0
		pack := &server_proto.WGSetPlayerBlackLevel{
			SnId:           proto.Int32(this.SnId),
			ResetTotalCoin: proto.Bool(true),
		}
		proto.SetDefaults(pack)
		this.SendToGame(int(server_proto.SSPacketID_PACKET_GW_AUTORELIEVEWBLEVEL), pack)

		//同步小游戏的黑白名单状态
		//MiniGameMgrSington.ClrPlayerWhiteBlackState(this)
	}
}

func (this *Player) BindGroupTag(tags []string) {
	if this.gateSess == nil {
		return
	}
	pack := &server_proto.SGBindGroupTag{
		Sid:  proto.Int64(this.sid),
		Code: server_proto.SGBindGroupTag_OpCode_Add,
		Tags: tags,
	}
	proto.SetDefaults(pack)
	this.gateSess.Send(int(server_proto.SSPacketID_PACKET_SG_BINDGROUPTAG), pack)
}

func (this *Player) UnBindGroupTag(tags []string) {
	if this.gateSess == nil {
		return
	}
	pack := &server_proto.SGBindGroupTag{
		Sid:  proto.Int64(this.sid),
		Code: server_proto.SGBindGroupTag_OpCode_Del,
		Tags: tags,
	}
	proto.SetDefaults(pack)
	this.gateSess.Send(int(server_proto.SSPacketID_PACKET_SG_BINDGROUPTAG), pack)
}

func (this *Player) TryRetrieveLostGameCoin(sceneid int) {
	//发送一个探针,等待ack后同步金币
	logProbe := model.NewCoinLog()
	logProbe.SnId = this.SnId
	logProbe.Count = 0     //必须是0,探针标识
	logProbe.RestCount = 0 //this.Coin
	logProbe.SeqNo = this.GameCoinTs
	logProbe.RoomId = int32(sceneid)
	LogChannelSington.WriteLog(logProbe)
	//先把玩家身上的钱清掉
	//this.Coin = 0
	this.SendDiffData()
}

func (this *Player) SyncGameCoin(sceneid int, enterts int64) {
	//游服金币冲账
	endts := time.Now().UnixNano()
	var err error
	var gamecoinlogs []model.CoinWAL
	task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
		gamecoinlogs, err = model.GetCoinWALBySnidAndInGameAndGreaterTs(this.Platform, this.SnId, int32(sceneid), enterts)
		return err
	}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
		if err == nil && len(gamecoinlogs) != 0 {
			oldCoin := this.Coin
			var cnt int64
			for i := 0; i < len(gamecoinlogs); i++ {
				ts := gamecoinlogs[i].Ts
				if ts >= enterts && ts <= endts {
					cnt = gamecoinlogs[i].Count
					this.Coin += cnt
					if ts > this.GameCoinTs {
						this.GameCoinTs = ts
					}
				}
			}
			newCoin := this.Coin
			newTs := this.GameCoinTs
			logger.Logger.Warnf("PlayerData(%v) SyncGameCoin before:enterts=%v before:Coin=%v after:GameCoinTs=%v after:Coin=%v", this.SnId, enterts, oldCoin, newTs, newCoin)
		}
		this.SendDiffData()
	}), "GetCoinWALBySnidAndInGameAndGreaterTs").Start()
}
func (this *Player) SendShowRed(showType hall_proto.ShowRedCode, showChild, isShow int32) {
	pack := &hall_proto.SCShowRed{
		ShowRed: &hall_proto.ShowRed{
			ShowType:  showType,
			ShowChild: proto.Int32(showChild),
			IsShow:    proto.Int32(isShow),
		},
	}
	proto.SetDefaults(pack)
	logger.Logger.Trace("SCShowRed:", pack)
	this.SendToClient(int(hall_proto.HallPacketID_PACKET_SC_SHOWRED), pack)
}
