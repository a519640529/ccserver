package model

import (
	"encoding/json"
	"games.yol.com/win88/webapi"
	"github.com/globalsign/mgo"
	"github.com/idealeak/goserver/core/logger"
	"io/ioutil"
	"strconv"
)

const TimeFormat = "2006-01-02 15:04:05"

type GiftInfo struct {
	MType int32
	Price int64
	Name  string
}
type CtrlSwitch struct {
	IsCoinPoolCtrlSwitch   bool //水池控制开关
	IsPlayerFeelCtrlSwitch bool //玩家体验开关 细化调控等提升玩家体验
}
type JackpotRatio struct {
	BigExRatio    int32 //额外大奖池分摊比例：千分比
	MiddleExRatio int32 //额外中奖池分摊比例：千分比
	SmallExRatio  int32 //额外小奖池分摊比例：千分比
	NormalRatio   int32 //普通池分摊比例：千分比
}

func (this *JackpotRatio) CalTaxRatio() int32 {
	tax := 1000 - (this.BigExRatio + this.MiddleExRatio + this.SmallExRatio + this.NormalRatio)
	if tax < 0 {
		return 0
	}
	return tax
}

type PacketNumFliter struct {
	MaxNum int //最大次数
	AddNum int //增加次数
}

type CtrlSmart struct {
	Version   int   // 版本号，修改版本号清除玩家本场次的总输赢分
	Threshold int64 // 阈值
}

type GameParam struct {
	Observers                       []func()
	NewPlayerCoin                   int32            //新建角色初始金豆数量
	BullfightMostBackwardFind       int32            //最多向后找几张牌
	BullfightEachGameChgCardMax     int32            //一局最多换几次
	WinthreeMostBackwardFind        int32            //最多向后找几张牌
	VerifyClientVersion             bool             //验证客户端版本信息
	HundredBullBankNumbers          int32            //百人场庄家在庄局数最大值
	UseEtcd                         bool             //是否使用etcd
	SrvMaintain                     bool             //服务器维护切换
	HundredSceneRecoveryCoe         map[string]int32 //百人场回收系数
	HundredInviteRobCnt             int32            //百人场邀请机器人数量
	HundredScenePreCreate           bool             //百人场预创建
	HundredRecoveryParam            float32          //百人牛牛回收参数
	HundredCloseRecovery            bool             //百人牛牛回收功能开关
	HundredBullRealPlayerNum        int32            //百人牛牛真实人数
	HundredBullAllBetParam          []int32          //百人牛牛机器人总下注数量
	HundredBullPlayerBetLimitParam  []float32        //百人牛牛机器人每个机器人的下注极限倍数
	HundredSceneCoinCecycle         int32            //金币回收率，千分比
	RollColorEventChance            int32            //深林舞会事件几率
	WaterMarginReturnBonus          float32          //水浒传出现概率
	SendWorldJackpots               int64            //赢取大于多少金币推送到公告
	WriteEventLog                   bool             //是否写事件log
	UpgradeAccountGiveCoinLimit     int32            //升级账号奖励金币数量限制 日期和IP
	GameTax                         int64            //游戏税率，百分比
	GuestDefaultName                string           //游客默认昵称
	WaterMarginRatio                JackpotRatio     //水浒传奖池分摊比例
	GoddessRatio                    JackpotRatio     //女赌神池分摊比例
	FruitMachineRatio               JackpotRatio     //水果机池分摊比例
	RollLineRatio                   JackpotRatio     //老虎机池分摊比例
	RollTeamRatio                   JackpotRatio     //世界杯池分摊比例
	FootballHeroRatio               JackpotRatio     //足球英豪池分摊比例
	IslandSurvivalRatio             JackpotRatio     //绝地求生池分摊比例
	SpreadAccountQPT                int32            //流水上报 条数/每次
	FakeVerifyCode                  string           //假的验证码
	MorePlayerLimit                 int              //moreplayerdata单次最大查询用户量
	BirdPlayerFlag                  bool             //是否关闭新手玩家的处理
	SamePlaceLimit                  int32            //和同一个人游戏的局数限制
	BullFightUseOldBill             bool             //斗牛结算使用旧规则
	RobotInviteInitInterval         int64            //机器人邀请的初始间隔时间(s)
	RobotInviteIntervalMax          int64            //机器人邀请的最大间隔时间(s)
	SceneMaxIdle                    int64            //空房间最大空闲时间(超过指定时间自动销毁)(s)
	DonotGetPromoterByIp            bool             //不通过ip获取推广信息
	LoginStateCacheSec              int32            //登录态数据缓存时间
	KickoutDefaultFreezeMinute      int64            //踢下线默认封号几分钟
	GameConfigGroupUseRobot         bool             //游戏分组使用机器人
	WinthreePokerLimit              int              //发牌尝试次数
	SameIpNoLimit                   bool             //同ip不做限制
	PreLoadRobotCount               int              //预加载机器人数量
	AdjustBullCardPercent           int              //展示牛牌的概率
	CheckDGScene                    bool             //校验dg场次
	CheckHBOScene                   bool             //校验hbo场次
	ForceSaveWhenCoinDeltaGt        int32            //当金币变化超过该值时强制写库
	QMOptimization                  bool             //全民推广流水优化
	ActGoldMaxSavePerMinite         int              //财神任务每分钟最多保存数量
	CoinPoolChangeWeight            []int            //?
	ExchangeNoLimitBind             bool             //兑换只允许使用身上绑定的账号信息进行
	NoticeCoinMin                   int64
	NoticeCoinMax                   int64
	RollGameHitPoolRate             int32 //拉霸游戏爆奖池的概率(万分比)
	WarningLoseLimit                int64
	WarningBetMax                   int64
	WarningWinRate                  int64
	WarningWinMoney                 int64
	CtrlSwitch                      map[string]CtrlSwitch
	EnterAfterStartSwitch           bool     //中途加入开关
	MongoDebug                      bool     //mongodb调试开关
	AIApi3thTimeout                 int32    //3方AI api超时时间
	WhiteHttpAddr                   []string //白名单
	RbAutoBalance                   bool
	RbAutoBalanceRate               float64
	RbAutoBalanceCoin               int64
	PlatformCleanTip                int
	GoldComeReplaceRobotRate        int32 //财神降临对战场真实玩家人数>目标人数后，将GoldComeReplaceRobotRate%的真实玩家替换成机器人
	DGCheckPlatformQuota            bool
	WWGCheckPlatformQuota           bool
	FGCheckPlatformQuota            bool
	AMEBACheckPlatformQuota         bool
	VRCheckPlatformQuota            bool
	SPTCheckPlatformQuota           bool
	DWCheckPlatformQuota            bool
	LEGCheckPlatformQuota           bool
	BBINCheckPlatformQuota          bool
	KYCheckPlatformQuota            bool
	AGCheckPlatformQuota            bool
	ABNoCheckPlatformQuota          bool
	OBNoCheckPlatformQuota          bool
	OGCheckPlatformQuota            bool
	QLCheckPlatformQuota            bool
	UseNewNumericalLogic            bool                        //使用新的数值逻辑
	UseNewNumericalLogicPercent     int                         //新的数值逻辑占用的规则
	MonitorInterval                 int32                       //单位秒
	OpenPlatformRank                bool                        //开启平台排行榜
	ThirdPltTransferInterval        int32                       //查询第三方平台向系统能否转账时间间隔，单位秒
	ThirdPltTransferMaxTry          int32                       //查询第三方平台向系统能否转账,最大尝试次数。默认是1次
	UseOldBlackListCheck            bool                        //使用老的黑名单检测机制
	ThirdPltReqTimeout              int32                       //三方平台接口请求超时时间设置,默认是30s
	WinThreeTrueRandom              bool                        //扎金花不再进行控牌
	UseBlacklistBindPlayerinfo      bool                        //使用黑名单绑定玩家信息，不再通过API拉取黑名单
	PresidentCD                     int64                       //转让会长的cd时间
	IosStablePrizeMinRecharge       int32                       //IOS稳定版最低充值需求
	IosStableInstallPrize           int32                       //IOS稳定版安装奖励
	ValidDeviceInfo                 bool                        //是否验证设备信息
	InvalidRobotAccRate             int                         //每次更换过期机器人账号的比例，百分比
	InvalidRobotDay                 int                         //机器人过期的天数
	CreatePrivateSceneCnt           int                         //每人可以创建私有房间数量
	PrivateSceneLogLimit            int                         //私有房间日志上限
	PrivateSceneFreeDistroySec      int                         //私有房间免费解散时间，默认600秒
	PrivateSceneDestroyTax          int                         //私有房间提前解散税收,百分比
	NumOfGamesConfig                []int32                     //私人房间局数
	ClubGiveCoinLimit               int32                       //俱乐部赠送金币最低限额,默认最低10000分,负值不限
	ClubTaxMin                      int32                       //俱乐部抽水比例下限限制
	ClubTaxMax                      int32                       //俱乐部抽水比例上限限制
	BacklogGameHorseRaceLamp        int                         //游戏内公告储备多少条，超出丢弃
	IsRobFightTest                  bool                        //是否开启机器人自己对战功能
	BullFightCtrl0108               bool                        //牛牛是否使用新功能规则
	OpenPoolRec                     bool                        //是否打开水池数据记录
	CoinPoolMinOutRate              int32                       //水池最小出分
	CoinPoolMaxOutRate              int32                       //水池最大出分
	MaxRTP                          float64                     //最大rtp
	AddRTP                          float64                     //附加rtp
	LegendSmallExRatio              int32                       //传奇拉霸进入小奖池的比例（万分比）
	LegendMiddleExRatio             int32                       //传奇拉霸进入中奖池的比例（万分比）
	LegendBigExRatio                int32                       //传奇拉霸进入大奖池的比例（万分比）
	LegendNormalRatio               int32                       //传奇拉霸进入正常奖池的比例（万分比）
	OpenFiveCardStudBugMode         bool                        //开启梭哈bug模式,高概率大牌
	FiveCardStudTryChgCardMax       int32                       //梭哈
	PlayerWatchNum                  int32                       //百人游戏允许围观的局数
	NotifyPlayerWatchNum            int32                       //百人游戏围观多少局的时候开始提示
	CgAddr                          string                      //后台cg工程地址
	DgApiUrlBackUp                  string                      //DG请求地址
	HboApiUrlBackUp                 string                      //HBO请求地址
	IsUseAI_V2                      bool                        //是否使用AI2.0版本,默认是不使用
	HZJHTryAdjustCardTimes          int                         //百人炸金花调控时，尝试调牌次数
	HZJHTrySingleAdjustCardTimes    int                         //百人炸金花单控，尝试调牌次数
	ReplayDataUseJson               bool                        //回放数据使用json格式
	IsPriorityMatchTrueMan          bool                        //是否优先真人匹配
	PreCreateRoomAllowMaxMultiple   int                         //预创建房间，允许扩展数量的最大倍数
	NoticePolicy                    int                         //跑马灯策略 0：根据金额优先级 1：优先级随机(默认使用0)
	MaxAudienceNum                  int                         //最大观战人数
	InvalideSnidTime                int32                       //过期角色的日期限制
	InvalideSnidLimit               int32                       //过期角色的查询最大数量
	WebApiUrlBackUp                 string                      //备用url,当填写时，采用此API请求后台
	MatchTrueManUseWeb              bool                        //真人匹配规则使用后台配置
	IsFindRoomByGroup               bool                        //查询房间列表时是否使用互通查询，默认是不使用
	NoOpTimes                       int32                       //对战场允许托管的局数
	IsOpenGameGroupRec              bool                        //是否打开分组归属0平台
	IsOpenPlayerEnterGameStateCheck bool                        //是否开启玩家进入游戏的时候状态检查
	UseBevRobot                     bool                        //是否使用行为树机器人
	GameBillMilliSecondCfg          map[string]map[string]int   //对应关系是为"平台:gameid:millsecond"
	ClosePreCreateRoom              bool                        //关闭予创建房间
	FFTest                          bool                        //测试标记，正式运行关闭
	IsQMFlowNeedTax                 bool                        //全民流水，打码流水是否计算税收
	CloseQMThr                      bool                        //关闭全民三方流水计算
	ZJHMatchAntesMutiple            int32                       //扎金花比赛初始下注倍数
	ProfitControlAutoCorrectMax     int32                       //杀率调控自动修正最大值
	ProfitControlAutoCorrectMin     int32                       //杀率调控自动修正最小值
	ProfitControlAutoCorrectStep    int32                       //杀率调控自动修正步长
	ProfitControlEffectSmart        bool                        //杀率调控影响智能化运营
	ErrResetMongo                   bool                        //发生主从问题，是否重置连接
	InvitePromoterBind              bool                        //绑定全民判定是否推广员
	MaxProcSamePacketNum            int                         //最大处理包数量
	AutoAddPacketNum                int                         //自动恢复包次数
	SelectMaxProcPacketNum          map[int]*PacketNumFliter    //特殊包定义请求次数
	SelectMaxProcPacketTNum         map[string]*PacketNumFliter //特殊包定义请求次数
	DDZUrlBackUp                    string                      //ddz请求地址

	//德州牛仔
	HDZNZTryAdjustCardTimes       int //德州牛仔调控时，尝试调牌次数
	HDZNZTrySingleAdjustCardTimes int //德州牛仔单控，尝试调牌次数
	//鱼虾蟹
	HYXXTryAdjustCardTimes       int  //鱼虾蟹调控时，尝试调牌次数
	HYXXTrySingleAdjustCardTimes int  //鱼虾蟹单控，尝试调牌次数
	DezhouEstimateCardQuality    bool //德州评估手牌质量

	TenHalfRate int // 十点半水池正常状态时也走水池亏钱调控的概率; 30 代表有30%的概率

	CanCtlLoseWinRate        float64  //满足可控牌的输赢比例
	GamePlayerCheckNum       int      //玩家局数检测
	DDZHeightRate            int      // 斗地主玩家赔率高，水池正常时走收分模式的概率
	CSURL                    string   //客户端域名
	InitGameHash             []string //初始哈希
	AtomGameHash             []string //原子哈希
	IsRobotInMatchSeasonRank bool     //机器人在榜开关
	MatchSeasonRankMaxNum    int      //在榜最大人数
}

var MahjongCareDeficitValue int32
var MahjongTimeFreeStart []int64
var MahjongTimeFreeEnd []int64
var GameParamPath = "../data/gameparam.json"
var GameParamData = &GameParam{}

// type GameParamTriggerFunc func()
// var	GameParamFuncObersers []GameParamTriggerFunc
func InitGameParam() {
	buf, err := ioutil.ReadFile(GameParamPath)
	if err != nil {
		logger.Logger.Error("InitGameParam ioutil.ReadFile error ->", err)
	}

	err = json.Unmarshal(buf, GameParamData)
	if err != nil {
		logger.Logger.Error("InitGameParam json.Unmarshal error ->", err)
	}
	if GameParamData.SpreadAccountQPT == 0 {
		GameParamData.SpreadAccountQPT = 100
	}
	if GameParamData.PresidentCD == 0 {
		GameParamData.PresidentCD = 2592000
	}
	if GameParamData.MorePlayerLimit == 0 {
		GameParamData.MorePlayerLimit = 50
	}
	if GameParamData.SamePlaceLimit > 10 {
		GameParamData.SamePlaceLimit = 10
	}
	if GameParamData.RobotInviteInitInterval == 0 {
		GameParamData.RobotInviteInitInterval = 2
	}
	if GameParamData.RobotInviteIntervalMax == 0 {
		GameParamData.RobotInviteIntervalMax = 10
	}
	if GameParamData.SceneMaxIdle == 0 {
		GameParamData.SceneMaxIdle = 3600
	}
	if GameParamData.LoginStateCacheSec == 0 {
		GameParamData.LoginStateCacheSec = 300
	}
	if GameParamData.KickoutDefaultFreezeMinute == 0 {
		GameParamData.KickoutDefaultFreezeMinute = 5
	}
	if GameParamData.WinthreePokerLimit == 0 {
		GameParamData.WinthreePokerLimit = 1
	}
	if GameParamData.AdjustBullCardPercent == 0 {
		GameParamData.AdjustBullCardPercent = 100
	}

	if GameParamData.MaxProcSamePacketNum == 0 {
		GameParamData.MaxProcSamePacketNum = 5
	}

	if GameParamData.AutoAddPacketNum == 0 {
		GameParamData.AutoAddPacketNum = 2
	}

	if GameParamData.SelectMaxProcPacketTNum == nil {
		if GameParamData.SelectMaxProcPacketNum == nil {
			GameParamData.SelectMaxProcPacketNum = make(map[int]*PacketNumFliter)
		}
	} else {
		if GameParamData.SelectMaxProcPacketNum == nil {
			GameParamData.SelectMaxProcPacketNum = make(map[int]*PacketNumFliter)
		}

		for k, v := range GameParamData.SelectMaxProcPacketTNum {
			ikey, err := strconv.Atoi(k)
			if err == nil {
				GameParamData.SelectMaxProcPacketNum[ikey] = &PacketNumFliter{
					MaxNum: v.MaxNum,
					AddNum: v.AddNum,
				}
			}
		}

	}

	if GameParamData.ForceSaveWhenCoinDeltaGt == 0 {
		GameParamData.ForceSaveWhenCoinDeltaGt = 100000
	}
	if GameParamData.ActGoldMaxSavePerMinite == 0 {
		GameParamData.ActGoldMaxSavePerMinite = 128
	}
	if GameParamData.NoticeCoinMin == 0 {
		GameParamData.NoticeCoinMin = GameParamData.SendWorldJackpots
	}
	if GameParamData.NoticeCoinMax == 0 {
		GameParamData.NoticeCoinMax = GameParamData.SendWorldJackpots * 3
	}
	if GameParamData.RollGameHitPoolRate == 0 {
		GameParamData.RollGameHitPoolRate = 5
	}

	if GameParamData.MaxRTP <= 0.00001 {
		GameParamData.MaxRTP = 0.999999
	}

	if GameParamData.AddRTP <= 0.00001 {
		GameParamData.AddRTP = 0.1
	}
	if GameParamData.CanCtlLoseWinRate <= 0.00001 {
		GameParamData.CanCtlLoseWinRate = 1.0
	}
	if GameParamData.CtrlSwitch == nil {
		GameParamData.CtrlSwitch = make(map[string]CtrlSwitch)
	}
	if GameParamData.AIApi3thTimeout == 0 {
		GameParamData.AIApi3thTimeout = 10
	}
	mgo.SetDebug(GameParamData.MongoDebug)
	if GameParamData.RbAutoBalance {
		if GameParamData.RbAutoBalanceRate == 0 {
			GameParamData.RbAutoBalanceRate = 2
		}
		if GameParamData.RbAutoBalanceCoin == 0 {
			GameParamData.RbAutoBalanceCoin = 100000
		}
	}

	if GameParamData.PlatformCleanTip == 0 {
		GameParamData.PlatformCleanTip = 50
	}
	if GameParamData.UseNewNumericalLogicPercent == 0 {
		GameParamData.UseNewNumericalLogicPercent = 50
	}
	if GameParamData.MonitorInterval == 0 {
		GameParamData.MonitorInterval = 60
	}
	if GameParamData.ThirdPltTransferInterval == 0 {
		GameParamData.ThirdPltTransferInterval = 3
	}
	if GameParamData.ThirdPltTransferMaxTry == 0 {
		GameParamData.ThirdPltTransferMaxTry = 1
	}
	if GameParamData.ThirdPltReqTimeout == 0 {
		GameParamData.ThirdPltReqTimeout = 30
	}
	if GameParamData.IosStablePrizeMinRecharge == 0 {
		GameParamData.IosStablePrizeMinRecharge = 5000
	}
	if GameParamData.IosStableInstallPrize == 0 {
		GameParamData.IosStableInstallPrize = 100
	}
	if GameParamData.InvalidRobotDay == 0 {
		GameParamData.InvalidRobotDay = 90
	}
	if GameParamData.CreatePrivateSceneCnt == 0 {
		GameParamData.CreatePrivateSceneCnt = 20
	}
	if GameParamData.PrivateSceneLogLimit == 0 {
		GameParamData.PrivateSceneLogLimit = 7000
	}
	if GameParamData.PrivateSceneFreeDistroySec == 0 {
		GameParamData.PrivateSceneFreeDistroySec = 600
	}
	if GameParamData.PrivateSceneDestroyTax == 0 {
		GameParamData.PrivateSceneDestroyTax = 5
	}
	if len(GameParamData.NumOfGamesConfig) == 0 {
		GameParamData.NumOfGamesConfig = []int32{5, 10, 20, 50}
	}
	if GameParamData.ClubGiveCoinLimit == 0 {
		GameParamData.ClubGiveCoinLimit = 10000
	}
	if GameParamData.ClubTaxMax == 0 {
		GameParamData.ClubTaxMax = 1500
	}
	if GameParamData.CoinPoolMinOutRate == 0 {
		GameParamData.CoinPoolMinOutRate = 33
	}

	if GameParamData.CoinPoolMaxOutRate == 0 {
		GameParamData.CoinPoolMaxOutRate = 66
	}

	if GameParamData.PlayerWatchNum <= 2 {
		GameParamData.PlayerWatchNum = 20
	}
	if GameParamData.NotifyPlayerWatchNum <= 2 {
		GameParamData.NotifyPlayerWatchNum = 17
	}
	if GameParamData.NotifyPlayerWatchNum >= GameParamData.PlayerWatchNum {
		GameParamData.NotifyPlayerWatchNum = GameParamData.PlayerWatchNum - 3
	}
	if GameParamData.NoOpTimes <= 2 {
		GameParamData.NoOpTimes = 2
	}

	if GameParamData.BacklogGameHorseRaceLamp == 0 {
		GameParamData.BacklogGameHorseRaceLamp = 50
	}
	if GameParamData.HZJHTryAdjustCardTimes == 0 {
		GameParamData.HZJHTryAdjustCardTimes = 3
	}
	if GameParamData.HZJHTrySingleAdjustCardTimes == 0 {
		GameParamData.HZJHTrySingleAdjustCardTimes = 20
	}
	GameParamData.LegendSmallExRatio = 2000
	GameParamData.LegendMiddleExRatio = 2000
	GameParamData.LegendBigExRatio = 2000
	GameParamData.LegendNormalRatio = 4000
	if GameParamData.PreCreateRoomAllowMaxMultiple == 0 {
		GameParamData.PreCreateRoomAllowMaxMultiple = 10
	}
	if GameParamData.MaxAudienceNum == 0 {
		GameParamData.MaxAudienceNum = 20
	}

	if GameParamData.WebApiUrlBackUp != "" {
		webapi.Config.GameApiURL = GameParamData.WebApiUrlBackUp
	}
	for _, v := range GameParamData.Observers {
		v()
	}
	if GameParamData.InvalideSnidTime == 0 {
		GameParamData.InvalideSnidTime = 3
	}
	if GameParamData.ProfitControlAutoCorrectMax == 0 {
		GameParamData.ProfitControlAutoCorrectMax = 5
	}
	if GameParamData.ProfitControlAutoCorrectStep == 0 {
		GameParamData.ProfitControlAutoCorrectStep = 1
	}
	if GameParamData.ZJHMatchAntesMutiple == 0 {
		GameParamData.ZJHMatchAntesMutiple = 5
	}
	if GameParamData.HDZNZTryAdjustCardTimes == 0 {
		GameParamData.HDZNZTryAdjustCardTimes = 3
	}
	if GameParamData.HDZNZTrySingleAdjustCardTimes == 0 {
		GameParamData.HDZNZTrySingleAdjustCardTimes = 20
	}
	if GameParamData.HYXXTryAdjustCardTimes == 0 {
		GameParamData.HYXXTryAdjustCardTimes = 3
	}
	if GameParamData.HYXXTrySingleAdjustCardTimes == 0 {
		GameParamData.HYXXTrySingleAdjustCardTimes = 20
	}
	if GameParamData.TenHalfRate == 0 {
		GameParamData.TenHalfRate = 40
	}
	if GameParamData.GamePlayerCheckNum == 0 {
		GameParamData.GamePlayerCheckNum = 16
	}
	if GameParamData.DDZHeightRate == 0 {
		GameParamData.DDZHeightRate = 40
	}
	if len(GameParamData.InitGameHash) == 0 {
		GameParamData.InitGameHash = []string{"ff6c5b1daa1068897377f7a64a762eafda4d225f25bf8e3bb476a7c4f2d10468"}
	}
	if len(GameParamData.AtomGameHash) == 0 {
		GameParamData.AtomGameHash = []string{`0ead8d98e67a7c9197a6bb0e664bb84adbeb25e4e0db63d2158e48b98a50534d`}
	}
	if GameParamData.MatchSeasonRankMaxNum == 0 {
		GameParamData.MatchSeasonRankMaxNum = 50
	}
}
func IsUseCoinPoolControlGame(gameid string, gamefreeid int32) bool {
	if data, ok := GameParamData.CtrlSwitch[strconv.Itoa(int(gamefreeid))]; ok {
		return data.IsCoinPoolCtrlSwitch
	} else {
		if data, ok := GameParamData.CtrlSwitch[gameid]; ok {
			return data.IsCoinPoolCtrlSwitch
		} else {
			return true
		}
	}
}
func IsCheckPlayerRateControl(gameid string, gamefreeid int32) bool {
	if data, ok := GameParamData.CtrlSwitch[strconv.Itoa(int(gamefreeid))]; ok {
		return data.IsPlayerFeelCtrlSwitch
	} else {
		if data, ok := GameParamData.CtrlSwitch[gameid]; ok {
			return data.IsPlayerFeelCtrlSwitch
		} else {
			return false
		}
	}
}
func GetHundredBullAllBetParam(sceneType int32) int32 {
	if sceneType < 0 || int(sceneType) >= len(GameParamData.HundredBullAllBetParam) {
		return GameParamData.HundredBullAllBetParam[len(GameParamData.HundredBullAllBetParam)-1]
	}
	return GameParamData.HundredBullAllBetParam[sceneType]
}
func GetHundredBullPlayerBetLimitParam(sceneType int32) float32 {
	if sceneType < 0 || int(sceneType) >= len(GameParamData.HundredBullPlayerBetLimitParam) {
		return GameParamData.HundredBullPlayerBetLimitParam[len(GameParamData.HundredBullPlayerBetLimitParam)-1]
	}
	return GameParamData.HundredBullPlayerBetLimitParam[sceneType]
}

func GetSmartVersion(gameFreeId string) int {
	return 0
}
