package common

// 游戏ID
const (
	GameId_Unknow             int = iota
	__GameId_Hundred_Min__        = 100 //################百人类################
	GameId_HundredDZNZ            = 101 //德州牛仔
	GameId_HundredYXX             = 102 //鱼虾蟹
	GameId_Baccarat               = 103 //百家乐
	GameId_RollPoint              = 104 //骰宝
	GameId_Roulette               = 105 //轮盘
	GameId_DragonVsTiger          = 106 //龙虎斗
	GameId_RedVsBlack             = 107 //红黑大战
	GameId_RollCoin               = 108 //奔驰宝马
	GameId_RollAnimals            = 109 //飞禽走兽
	GameId_RollColor              = 110 //森林舞会
	GameId_RedPack                = 111 //红包扫雷
	GameId_Crash                  = 112 //碰撞游戏
	__GameId_VS_Min__             = 200 //################对战类################
	GameId_DezhouPoker            = 201 //德州扑克
	GameId_DezhouPoker_5To2       = 202 //德州扑克5选2
	GameId_OmahaPoker             = 203 //奥马哈
	GameId_TenHalf                = 204 //十点半
	GameId_FiveCardStud           = 205 //梭哈
	GameId_BlackJack              = 206 //21点
	GameId_TienLen                = 207 //tienlen经典版
	GameId_TienLen_yl             = 208 //tienlen娱乐版
	GameId_TienLen_toend          = 209 //tienlen经典版（打到底）
	GameId_TienLen_yl_toend       = 210 //tienlen娱乐版（打到底）
	GameId_TienLen_m              = 807 //tienlen经典比赛场
	GameId_TienLen_yl_toend_m     = 808 //tienlen打到底比赛场
	__GameId_Slot_Min__           = 300 //################拉霸类################
	GameId_CaiShen                = 301 //财神
	GameId_Avengers               = 302 //复仇者联盟
	GameId_EasterIsland           = 303 //复活岛
	GameId_IceAge                 = 304 //冰河世纪
	GameId_TamQuoc                = 305 //百战成神
	GameId_Fruits                 = 306 //水果拉霸
	GameId_Richblessed            = 307 //多福多财
	__GameId_Fishing_Min__        = 400 //################捕鱼类################
	GameId_HFishing               = 401 //欢乐捕鱼
	GameId_TFishing               = 402 //天天捕鱼
	__GameId_Casual_Min__         = 500 //################休闲类################
	__GameId_MiniGame_Min__       = 600 //################小游类################
	GameId_Candy                  = 601 //糖果小游戏
	GameId_MiniPoker              = 602 //MiniPoker
	GameId_BOOM                   = 603 //卡丁车
	GameId_LuckyDice              = 604 //幸运筛子
	GameId_CaoThap                = 605 //CaoThap
	__GameId_ThrGame_Min__        = 700 //################三方类################
	GameId_Thr_Dg                 = 701 //DG Game
	GameId_Thr_XHJ                = 901 //DG Game
)

// 特殊物品id
const (
	ItemId_Nil    int32 = iota
	ItemId_Coin         //1.金币
	ItemId_Money        //2.钻石
	ItemId_Card         //3.房卡
	ItemId_RMB          //4.红包
	ItemId_Ticket       //5.兑换券
	ItemId_Tool         //6.道具
)

// 房间编号区间
const (
	PrivateSceneStartId   = 10000000
	PrivateSceneMaxId     = 99999999
	MatchSceneStartId     = 100000000
	MatchSceneMaxId       = 199999999
	HundredSceneStartId   = 200000000
	HundredSceneMaxId     = 299999999
	HallSceneStartId      = 300000000
	HallSceneMaxId        = 399999999
	MiniGameSceneStartId  = 400000000
	MiniGameSceneMaxId    = 409999999
	SpecailEmptySceneId   = 410000000
	SpecailUnstartSceneId = 419999999
	CoinSceneStartId      = 1000000000 //区间预留大点,因为队列匹配比较耗id,假定一天100w牌局,那么这个id区间够用1000天
	CoinSceneMaxId        = 1999999999
	DgSceneId             = 99
	WwgSceneId            = 199
	FgSceneId             = 299
	AmebaSceneId          = 399
	VrSceneId             = 499
)

// 房间模式
const (
	SceneMode_Public  = iota //公共房间
	SceneMode_Club           //俱乐部房间
	SceneMode_Private        //私人房间
	SceneMode_Match          //赛事房间
	SceneMode_Thr            //三方房间
)

// 场景级别
const (
	SceneLvl_Test       = -1 // 试玩场(不要钱)
	SceneLvl_Experience = 0  // 体验场(花小钱)
	SceneLvl_Primary    = 1  // 初级场
	SceneLvl_Middle     = 2  // 中级场
	SceneLvl_Senior     = 3  // 高级场
	SceneLvl_Professor  = 4  // 专家场
)

// 房费选项
const (
	RoomFee_Owner int32 = iota //房主
	RoomFee_AA                 //AA
	RoomFee_Max
)

const (
	Platform_Rob     = "__$G_P$__"
	Platform_Sys     = "0"
	Channel_Rob      = "$${ROBOT}$$"
	PMCmd_SplitToken = "-"
	PMCmd_AddCoin    = "addcoin"
	PMCmd_Privilege  = "setprivilege"
	DeviceOS_Android = "android"
	DeviceOS_IOS     = "ios"
)

const (
	GainWay_NewPlayer             int32 = 0  //0.新建角色
	GainWay_Pay                         = 1  //1.后台增加(主要是充值)
	GainWay_ByPMCmd                     = 2  //2.pm命令
	GainWay_MatchBreakBack              = 3  //3.退赛退还
	GainWay_MatchSystemSupply           = 4  //4.比赛奖励
	GainWay_Exchange                    = 5  //5.兑换
	GainWay_ServiceFee                  = 6  //6.桌费
	GainWay_CoinSceneWin                = 7  //7.金豆场赢取
	GainWay_CoinSceneLost               = 8  //8.金豆场输
	GainWay_CoinSceneEnter              = 9  //9.进入金币场预扣
	GainWay_ShopBuy                     = 10 //10.商城购买或者兑换
	GainWay_CoinSceneLeave              = 11 //11.金豆场回兑
	GainWay_HundredSceneWin             = 12 //12.万人场赢取
	GainWay_HundredSceneLost            = 13 //13.万人场输
	GainWay_MessageAttach               = 14 //14.邮件
	GainWay_SafeBoxSave                 = 15 //15.保险箱存入
	GainWay_SafeBoxTakeOut              = 16 //16.保险箱取出
	GainWay_Fishing                     = 17 //17.捕鱼
	GainWay_CoinSceneExchange           = 18 //18.金豆场兑换
	GainWay_UpgradeAccount              = 19 //19.升级账号
	GainWay_API_AddCoin                 = 20 //20.API操作钱包
	GainWay_GoldCome                    = 21 //21.财神降临
	GainWay_Transfer_System2Thrid       = 22 //22.系统平台转入到第三方平台的金币
	GainWay_Transfer_Thrid2System       = 23 //23.第三方平台转入到系统平台的金币
	GainWay_RebateTask                  = 24 //24.返利获取
	GainWay_IOSINSTALLSTABLE            = 25 //25.ios安装奖励
	GainWay_VirtualChange               = 26 //26.德州虚拟账变
	GainWay_CreatePrivateScene          = 27 //27.创建私有房间
	GainWay_PrivateSceneReturn          = 28 //28.解散私有房间返还
	GainWay_OnlineRandCoin              = 29 //29.红包雨
	GainWay_Expire                      = 30 //30.到期清理
	GainWay_PromoterBind                = 31 //31.手动绑定推广员
	GainWay_GradeShopReturn             = 32 //32.积分商城撤单退还积分
	GainWay_Api_In                      = 33 //33.转移金币
	GainWay_Api_Out                     = 34 //34.转移金币
	GainWay_Shop_Buy                    = 35 //35.购买记录
	GainWay_MAIL_MTEM                   = 36 //36.邮件领取道具
	GainWay_Item_Sale                   = 37 //37.道具出售
	GainWay_ReliefFund                  = 38 //38.领取救济金
	GainWay_Shop_Revoke                 = 39 //39.撤单
	GainWay_ActSign                     = 40 //40.签到
	GainWay_MatchSignup                 = 41 //比赛报名费用
	GainWay_MatchSeason                 = 42 //比赛赛季奖励
)

// 后台选择 金币变化类型 的充值 类型id号起始
const GainWaySort_Api = 1000

// 自定义局数起始索引
const CUSTOM_PER_GAME_INDEX_BEG int32 = 10

const (
	ClientRole_Agentor int = iota
	ClientRole_Player
	ClientRole_GM
	ClientRole_Max
)

const (
	CoinSceneOp_Enter          int32 = iota //进入
	CoinSceneOp_Leave                       //离开
	CoinSceneOp_Change                      //换桌
	CoinSceneOp_AudienceEnter               //观众进入
	CoinSceneOp_AudienceLeave               //观众离开
	CoinSceneOp_AudienceChange              //观众换桌
	CoinSceneOP_Server
)

// 玩家离开原因
const (
	PlayerLeaveReason_Normal          int = iota //主动离开
	PlayerLeaveReason_Bekickout                  //钱不够被踢出
	PlayerLeaveReason_OnDestroy                  //房间销毁强制退出
	PlayerLeaveReason_CoinScene                  //玩家暂离金豆自由场
	PlayerLeaveReason_ChangeCoinScene            //玩家换桌
	PlayerLeaveReason_OnBilled                   //结算完毕
	PlayerLeaveReason_DropLine                   //掉线离开
	PlayerLeaveReason_LongTimeNoOp               //长时间未操作
	PlayerLeaveReason_GameTimes                  //超过游戏次数强制离开
	PlayerLeaveReason_RoomFull                   //房间人数已达上限
)

// 万分比
const RATE_BASE_VALUE int32 = 10000

const (
	SceneState_Normal  int = iota
	SceneState_Fishing     //鱼潮
)

const (
	PlayerType_Rob      int32 = 0
	PlayerType_Undefine       = 1
	PlayerType_Black          = -1
	PlayerType_White          = -2
)

const (
	CoinPoolAIModel_Default  int32 = iota //默认
	CoinPoolAIModel_Normal                //正常模式
	CoinPoolAIModel_ShouFen               //收分模式
	CoinPoolAIModel_ZheZhong              //折中模式
	CoinPoolAIModel_TuFen                 //吐分
	CoinPoolAIModel_Max                   //
)

const (
	RobotServerType  int = 9
	RobotServerId        = 901
	DataServerType   int = 10
	DataServerId         = 1001
	AIServerType         = 11
	AIServerId           = 1101
	ActThrServerType int = 12
	ActThrServerID       = 1201
)

// 踢号原因
const (
	KickReason_Unknow     int32 = iota
	KickReason_OtherLogin       //顶号
	KickReason_Freeze           //冻结
	KickReason_ResLow     = 5   //资源
	KickReason_AppLow     = 6   //App版本

	KickReason_CheckCodeErr  = 7  //校验错误
	KickReason_TaskErr       = 8  //任务错误
	KickReason_CantFindAcc   = 9  //查找sid acc错误
	KickReason_DBLoadAcc     = 10 //数据库登录错误
	KickReason_Logining           //登陆中
	KickReason_Disconnection = 99 //网络断开
)

// 游戏类型
const (
	GameType_Unknow  int32 = iota
	GameType_Hundred       // 百人类
	GameType_PVP           // 对战类
	GameType_Slot          // 拉霸类
	GameType_Fishing       // 捕鱼类
	GameType_Casual        // 休闲类
	GameType_Mini          // 小游戏类
	GameType_Thr           // 三方类
)

func IsHundredType(gt int32) bool {
	return gt == GameType_Hundred || gt == GameType_Slot || gt == GameType_Casual
}

func IsCoinSceneType(gt int32) bool {
	return gt == GameType_PVP || gt == GameType_Fishing
}

const PreCreateSceneAudienceNum = 20

const (
	HorseRaceLampPriority_Coin int = iota //根据玩家输赢金额大小设置优先级(在最大最小限额的基础上)
	HorseRaceLampPriority_Rand            //随机(在最大最小限额的基础上)
)

// 设备
const (
	Web        = 0
	Android    = 1
	IOS        = 2
	WebStr     = "h5"
	AndroidStr = "android"
	IOSStr     = "ios"
)

var DeviceName = map[int]string{
	Web:     WebStr,
	Android: AndroidStr,
	IOS:     IOSStr,
}

var DeviceNum = map[string]int{
	WebStr:     Web,
	AndroidStr: Android,
	IOSStr:     IOS,
}

const (
	MatchTrueMan_Forbid    int32 = -1 //禁止匹配真人
	MatchTrueMan_Unlimited       = 0  //不限制
	MatchTrueMan_Priority        = 1  //优先匹配真人
)

const (
	SingleAdjustModeNormal = 0
	SingleAdjustModeWin    = 1
	SingleAdjustModeLose   = 2
)

// 自动化标签(程序里产生的全部<0)
const (
	AutomaticTag_QZNN_Smart int32 = -1
)

const (
	SceneParamEx_DbGameFreeId = 0
	SceneParamEx_CanStartNum  = 1 //游戏开始的要求人数
)

// 比赛参数
const (
	PlayerIParam_MatchRank int = iota
	PlayerIParam_IsQuit
	PlayerIParam_MatchWeaken
	PlayerIParam_TotalOfGames
)

const (
	PARAMEX_GAMEFREEID          int = iota //游戏id
	PARAMEX_MATCH_COPYID                   //比赛副本id
	PARAMEX_MATCH_ID                       //比赛id
	PARAMEX_MATCH_PHASEIDX                 //赛程阶段idx
	PARAMEX_MATCH_MMTYPE                   //赛制
	PARAMEX_MATCH_BASESCORE                //底分
	PARAMEX_MATCH_NUMOFGAME                //局数
	PARAMEX_MATCH_OUTSCORE                 //出局分数
	PARAMEX_MATCH_RESTPLAYERNUM            //剩余人数
	PARAMEX_MATCH_STARTNUM                 //本轮开始人数
	//PARAMEX_MATCH_RANK                     //我的排名,构建房间消息时,额外补充,每个玩家不同
)

const (
	MaxLoopNum    = 1000
	DefaultResult = 0
)

const (
	CodeTypeSMS     = 0 // 短信验证码
	CodeTypeStr     = 1 // 字符串验证码
	CodeTypeHuaKuai = 2 // 滑块验证码
	CodeTypeNo      = 3 // 不使用验证码
)

const (
	ActId_Share            int = iota //0.微信分享
	ActId_OnlineReward                //1.在线奖励
	ActId_UpgradeAccount              //2.升级账号
	ActId_GoldTask                    //3.财神任务
	ActId_GoldCome                    //4.财神降临
	ActId_LuckyTurntable              //5.转盘活动
	ActId_Yeb                         //6.余额宝
	ActId_Card                        //7.周卡月卡
	ActId_RebateTask                  //8.返利获取
	ActId_IOSINSTALLSTABLE            //9.ios安装奖励
	ActId_VipLevelBonus               //10.vip日周月等级奖励
	ActId_LoginRandCoin               //11.登录红包
	ActId_OnlineRandCoin              //12.红包雨
	ActId_MatchSwitch                 //13.比赛开关
	ActId_PromoterBind                //14.手动绑定推广员
	ActId_Lottery                     //15.彩金池
	ActId_Task                        //16.活跃任务
	ActId_PROMOTER                    //17.全民推广
	ActId_Activity                    //18.活动界面
	ActId_NewYear                     //19.新年暗号红包活动
	ActId_Guess                       //20.猜灯谜活动
	ActId_Sign                        //21.七日签到
	ExchangeId_Alipay                 //22.兑换到支付宝
	ExchangeId_Bank                   //23.兑换到银行卡
	ExchangeId_Wechat                 //24.兑换到微信
	ActId_Max
)

// 匹配模式
const (
	MatchMode_Normal int32 = iota //普通匹配
	MatchMode_Quene               //队列匹配
)

const (
	SCENE_BIGWINHISTORY_MAXNUMBER    = 40 // 爆奖记录最大数量
	SCENE_BIGWINHISTORY_LIMITNUMBER  = 10 // 假数据生成临界值
	SCENE_BIGWINHISTORY_TIMEINTERVAL = 2  // 假数据生成定点时间间隔，单位：小时（实际时间 = 定点时间 + 随机时间）
)
const (
	OrderColumnInvalid           = 0  // 默认
	OrderColumnCoinPayTotal      = 1  // 充值
	OrderColumnCoinExchangeTotal = 2  // 提现
	OrderColumnTaxTotal          = 3  // 税收
	OrderColumnRegisterTime      = 4  // 注册时间
	OrderColumnRoomNumber        = 5  // 游戏房间号
	OrderColumnLose              = 6  // 输次数
	OrderColumnWin               = 7  // 赢次数
	OrderColumnDraw              = 8  // 平次数
	OrderColumnWinCoin           = 9  // 赢分
	OrderColumnLoseCoin          = 10 // 输分
)

func IsLocalGame(gameId int) bool { //本土玩法
	return gameId == GameId_TienLen || gameId == GameId_TienLen_yl || gameId == GameId_TienLen_toend || gameId == GameId_TienLen_yl_toend
}
