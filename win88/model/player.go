package model

import (
	"bytes"
	"crypto/md5"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"games.yol.com/win88/protocol/webapi"

	"games.yol.com/win88/common"
	"github.com/globalsign/mgo/bson"
	"github.com/idealeak/goserver/core/logger"
)

const (
	VER_PLAYER_DEFAULT int32 = iota //初始版本
	VER_PLAYER_MAX
)

const (
	PLAYER_FLAGS_PRIVILEGE    int64 = 1 << iota
	PLAYER_FLAGS_FIRSTGAME          //首次游戏
	PLAYER_FLAGS_FIRSTBINDTEL       //首次绑定账号
	PLAYER_FLAGS_CANREBATE          //是否能够返利
)

const (
	DEFAULT_PLAYER_SAFEBOX_PWD = "" //保险箱默认密码
)

const (
	MAX_RANK_COUNT = 50
)

type PlayerParams struct {
	Platform     int32  `json:"platform"`     //0:windows   1:ios  2:android
	Ip           string `json:"ip"`           //ip地址
	City         string `json:"city"`         //所在城市
	Logininmodel string `json:"logininmodel"` //"app"   或者  "web"
	Name         string `json:"name"`         //微信昵称
	Head         string `json:"head"`         //微信头像
	UnionId      string `json:"unionid"`      //微信unionid
	IosIsStable  bool   `json:"iosIsStable"`  //是否是IOS稳定版
}

type LoginCallback func(pd *PlayerData)

var FirstLoginCB LoginCallback

type PlayerDiffData struct {
	Coin                 int64 //金币
	Diamond              int64 //钻石
	VCoin                int64 //V卡
	SafeBoxCoin          int64 //保险箱金币
	CoinPayTotal         int64 //总充值金额
	VIP                  int32 //vip等级
	TotalConvertibleFlow int64 //流水值
	ClubCoin             int64 //俱乐部金币
	Ticket               int64 //比赛入场券
	Grade                int64 //积分
}

type PlayerBaseData struct {
	SnId int32  //数字唯一id
	Name string //名字
	Sex  int32  //性别
	Coin int64  //金币
}

const (
	TASKSTATUS_RUNNING     int32 = iota //进行中
	TASKSTATUS_COMPLETE                 //已完成
	TASKSTATUS_PRIZEDRAWED              //奖励已领取
	TASKSTATUS_EXPIRE                   //已超期
)

const (
	WEBEVENT_LOGIN      = 1 //首次登陆
	WEBEVENT_UPGRADEACC = 2 //绑定升级账号
)

const (
	IOSSTABLESTATE_NIL  int32 = iota //IOS未标注稳定版
	IOSSTABLESTATE_MARK              //已标注
	IOSSTABLESTATE_DONE              //奖励已发放
)

// 任务数据
type TaskData struct {
	Id             int32 //任务id
	Ts             int32 //时间戳
	Data           int64 //任务数据
	CompletedTimes int32 //已完成次数
}

type TaskDatas struct {
	ConfigVer int64                //时间戳
	Tasks     map[string]*TaskData //taskid
}

type TaskCond struct {
	CondId int32
	Count  int32
}

// 通用活动数据
type ActData struct {
	Id    int32
	Ts    int32
	Datas []int32
}

// 通用限制数据
type LimitData struct {
	Id        int32
	Num       int32
	NextReset int32
}

// 游戏选项数据
type GameOption struct {
	Options []int64
}

type PlayerGameCtrlData struct {
	CtrlData             map[string]*PlayerGameStatics
	RechargeCoin         int64 //充值金额
	ExchangeCoin         int64 //兑换金额
	TodayConvertibleFlow int64 //今日流水
	Ts                   int64 //日期
}

func NewPlayerGameCtrlData() *PlayerGameCtrlData {
	t := &PlayerGameCtrlData{CtrlData: make(map[string]*PlayerGameStatics)}
	t.Ts = time.Now().Unix()
	return t
}

type PlayerGameStatics struct {
	GameTimes     int64  //总局数
	WinGameTimes  int64  //赢的局数
	LoseGameTimes int64  //输的局数
	DrawGameTimes int64  //和的局数
	TotalIn       int64  //总游戏投入(输掉的金币)
	TotalOut      int64  //总游戏产出(赢取的金币),税前
	MaxSysOut     int64  //单局最大盈利,税前
	OtherStatics  []byte //其他特殊统计信息(游戏自定义)
	Version       int32  //数据版本，游戏有时候数据计算错误，需要重置所有玩家的数据
}

type PlayerGameInfo struct {
	Statics   PlayerGameStatics //游戏统计类数据
	FirstTime time.Time         //首次参与游戏时间
	Data      []int64           //游戏场次数据
	DataEx    []byte            //其他扩展数据,可自定义json或者其他二进制数据
	Version   int32             //数据版本，游戏有时候数据计算错误，需要重置所有玩家的数据
}

type PlayerGameTotal struct {
	ProfitCoin int64 //盈利总额
	BetCoin    int64 //有效投注总额
	FlowCoin   int64 //流水总额
}
type ShopTotal struct {
	AdLookedNum  int32 //已经观看的次数
	AdReceiveNum int32 //已经领取的次数
}
type RolePetInfo struct {
	ModUnlock map[int32]int32 //已经解锁的id
	ModId     int32           //使用中的id
}

// 七日签到数据
type SignData struct {
	SignIndex       int   //签到次数
	LastSignTickets int64 //上一次签到时间 时间戳
}

//// 在线奖励数据
//type OnlineRewardData struct {
//	Version        int32
//	Ts             int64  // 上次累计时间戳
//	OnlineDuration uint32 // 累计在线时长
//	RewardReceived uint32 // 奖励获取情况
//}
//
//// 幸运转盘数据
//type LuckyTurnTableData struct {
//	Score              int64   // 当前积分
//	TomorrowScore      int64   // 明日积分
//	TomorrowFloatScore float64 // 明日积分64位的，防止精度丢失导致的数据不正常
//
//}

// 排行榜
type Rank struct {
	SnId      int32
	Name      string
	Head      int32
	VIP       int32
	TotalCoin int64
}

//
//// 余额宝数据
//type YebData struct {
//	TotalIncome int64 // 累计收益
//	PrevIncome  int64 // 上轮收益
//	Balance     int64 // 余额（总金额）
//	InterestTs  int64 // 起息时间（该时间一小时后产生收益）
//}
//
//// 卡（周卡月卡）数据
//type CardData struct {
//	BuyTs     int64 // 购买时间
//	ReceiveTs int64 // 领取时间
//}
//
//// 阶梯充值数据
//type StepRechargeData struct {
//	Version     int32 // 版本
//	CurrentStep int32 // 当前阶段
//	Ts          int64 // （上次）充值时间
//}

// 财神任务数据
type GoldTaskData struct {
	TaskId         int32 //任务ID
	Data           int64 //任务数据
	CompletedTimes int32 //已完成次数
}

type GoldTaskDatas struct {
	ConfigVer     int64                    //时间戳
	DataTaskIdMap map[string]*GoldTaskData //taskid
}
type RebateData struct {
	TodayRebateCoin        int64 //今日返利
	YesterdayRebateCoin    int64 //昨日返利
	ValidBetTotal          int64 //今日有效下注
	YesterdayValidBetTotal int64 //昨日有效下注
	TotalRebateCoin        int64 //今日累计已领取的返利
	TotalHaveRebateCoin    int64 //往日累计未领取的返利
	TotalHaveValidBetTotal int64 //往日累计有效下注
}

////vip数据
//type ActVipBonusData struct {
//	Level     int32 //等级礼包领取情况 位计算
//	Day       int32 //每日领取情况 0 未领取 1已领取
//	Week      int32 //每周领取情况	 0 未领取 1已领取
//	Month     int32 //每月领取情况	 0 未领取 1已领取
//	LastWeek  int64 //上次领取周时间
//	LastMonth int64 //上次领取月时间
//}

// 状态
type PayActState struct {
	Ts int64 //获得时间
}

func (this *GoldTaskDatas) GetData(taskid int32) *GoldTaskData {
	if this.DataTaskIdMap != nil {
		if td, ok := this.DataTaskIdMap[fmt.Sprintf("%v", taskid)]; ok {
			return td
		}
	}
	return nil
}

// 比赛免费报名记录
type MatchFreeSignupRec struct {
	LastSignupTs int64 //最后一次报名时间
	UseTimes     int32 //累计使用免费次数
}

type PlayerData struct {
	Id                        bson.ObjectId              `bson:"_id"`
	AccountId                 string                     //账号id
	SnId                      int32                      //数字唯一id
	NiceId                    int32                      //靓号
	Name                      string                     //名字
	Remark                    string                     //备注
	Platform                  string                     //平台
	Channel                   string                     //渠道信息
	DeviceOS                  string                     //设备操作系统
	DeviceId                  string                     //设备id
	PackageID                 string                     //推广包标识 对应客户端的packagetag
	Package                   string                     //包信息 android:包名 ios:bundleid
	IsRob                     bool                       //是否是机器人
	Head                      int32                      //头像
	HeadUrl                   string                     //头像
	Sex                       int32                      //性别
	HeadOutLine               int32                      //头像框
	VIP                       int32                      //VIP帐号 等级
	GMLevel                   int32                      //GM等级
	WinTimes                  int32                      //胜利次数
	FailTimes                 int32                      //失败次数
	DrawTimes                 int32                      //平局次数
	WinCoin                   int64                      //总赢钱数量
	FailCoin                  int64                      //总输钱数量
	Tel                       string                     //电话号码
	Ip                        string                     //最后登录ip地址
	RegIp                     string                     //注册ip地址
	City                      string                     //城市
	Params                    string                     //其他参数
	AlipayAccount             string                     //支付宝账号
	AlipayAccName             string                     //支付宝实名
	Bank                      string                     //绑定的银行名称
	BankAccount               string                     //绑定的银行账号
	BankAccName               string                     //绑定的银行账号
	Coin                      int64                      //金豆
	CoinPayTs                 int64                      //金豆冲账时间戳
	CoinPayTotal              int64                      //在线总充值金额
	CoinExchangeTotal         int64                      //总提现金额 兑换
	SafeBoxCoin               int64                      //保险箱金币
	SafeBoxPassword           string                     //保险箱密码
	Diamond                   int64                      //钻石
	InviterId                 int32                      //邀请人Id
	InviterName               string                     //邀请人名称
	InviterHead               int32                      //邀请人头像
	BeUnderAgentCode          string                     //隶属经销商（推广人）
	SubBeUnderAgentCode       string                     //经销商子id
	Flags                     int64                      //标记
	GameCoinTs                int64                      //游服金币对账时间戳
	Ver                       int32                      //数据版本号
	CheckSum                  uint32                     //校验码(预防暴库修改数据)
	UpgradeTime               time.Time                  //升级账号时间，绑定手机号时间
	CreateTime                time.Time                  //创建时间
	LastLoginTime             time.Time                  //最后登陆时间
	LastLogoutTime            time.Time                  //最后退出时间
	AllowSpeakTime            int64                      //允许下次发言的时间戳
	AgentType                 int32                      //代理类型 0:普通用户  其它为代理
	GameTax                   int64                      //总游戏税收
	SafeBoxCoinTs             int64                      //保险箱冲账时间戳
	WhiteFlag                 int32                      //特殊白名单标记
	WBCoinTotalOut            int64                      //加入黑白名单后玩家总产出
	WBCoinTotalIn             int64                      //加入黑白名单后玩家总投入
	WBCoinLimit               int64                      //黑白名单输赢额度,额度变为0时自动解除黑白名单
	WBMaxNum                  int32                      //黑白名单最大干预次数
	WBTime                    time.Time                  //黑白名单操作时间
	TotalCoin                 int64                      //总金币
	PromoterTree              int32                      //推广树信息
	TotalConvertibleFlow      int64                      //玩家流水总额 默认1:1 //流水统计使用
	TotalFlow                 int64                      //历史总流水
	CanExchangeBeforeRecharge int64                      //充值之前可兑换金额 //流水统计使用
	LastRechargeWinCoin       int64                      //充值后流水
	BlacklistType             int32                      //黑名单作用域和后台一样都是采用位标记的表示形式 // 0是不限制 第1位是游戏登录 第2位是兑换 第3位是充值，注意这个地方是黑名单管理的作用域+1 //主要是为了在mgo没有设置黑名单类型的时候，默认是不限制的
	ForceVip                  int32                      //强制VIP等级，通过后台设置，如果设置了当前值，就不再通过paytotal计算vip等级
	LastExchangeTime          int64                      //最后兑换时间
	LastExchangeOrder         string                     //最后的赠与订单
	LogicLevels               []int32                    //用户分层信息
	AutomaticTags             []int32                    //用户自动化标记
	TelephonePromoter         int32                      //电销推广员标识，用于电销标记
	TelephoneCallNum          int32                      //电销次数
	Ticket                    int64                      //比赛券
	TicketPayTs               int64                      //比赛券冲账时间点
	TicketTotal               int64                      //累计总获得比赛券数量
	TicketTotalDel            int64                      //累计清理掉的数量
	Grade                     int64                      //积分
	TagKey                    int32                      //包标识关键字
	LoginTimes                int                        //用户登录次数
	YesterdayGameData         *PlayerGameCtrlData        //昨日游戏统计数据
	TodayGameData             *PlayerGameCtrlData        //今日游戏统计数据
	IsFoolPlayer              map[string]bool            //每个游戏是否是新手玩家
	TotalGameData             map[int][]*PlayerGameTotal //统计数据 1.棋牌 2.电子 3.捕鱼 4.视讯 5.彩票 6.体育 7.个人房间 8.俱乐部房间 9.三方游戏
	GDatas                    map[string]*PlayerGameInfo //玩家游戏统计数据 key:gameFreeId, key:gameid
	MarkInfo                  string                     //用来备注玩家信息
	DeviceInfo                string                     //设备信息
	WBLevel                   int32                      //黑白名单 白:[1,10] 黑:[-1,-10]
	ShopTotal                 map[int32]*ShopTotal       //key为商品id
	ShopLastLookTime          map[int32]int64            //商品上一次的观看时间
	Roles                     *RolePetInfo               //人物
	Pets                      *RolePetInfo               //宠物
	WelfData                  *WelfareData               //活动数据
	*SignData
}

type WelfareData struct {
	ReliefFundTimes int32 //救济金领取次数
}

type GradeShop struct {
	ConsigneeName string          //收货人名字
	ConsigneeTel  string          //收货人电话
	ConsigneeAddr string          //收货人地址
	ExchangeDay   map[int32]int32 //今天兑换了多少
}

type PlayerBaseInfo struct {
	SnId             int32
	AccountId        string
	Name             string
	Platform         string
	Channel          string
	BeUnderAgentCode string
	DeviceOS         string
	PackageID        string
	Package          string
	Tel              string
	Head             int32
	Sex              int32
	GMLevel          int32
	Coin             int64
	SafeBoxCoin      int64
	PromoterTree     int32
	TelephoneCallNum int32
}
type PlayerDataForWeb struct {
	AlipayAccName     string
	AlipayAccount     string
	Bank              string
	BankAccName       string
	BankAccount       string
	BlackLevel        int32
	Coin              int64
	CoinExchangeTotal int64
	CoinPayTotal      int64
	CreateTime        time.Time
	DeviceId          string
	DeviceOS          string
	DrawTimes         int32
	FailCoin          int64
	FailTimes         int32
	GameTax           int64
	Ip                string
	IsRob             bool
	LastLoginTime     time.Time
	MarkInfo          string
	Name              string
	Online            bool
	Package           string
	PackageID         string
	Platform          string
	RegIp             string
	SafeBoxCoin       int64
	SnId              int32
	Tel               string
	VIP               int32
	WhiteLevel        int32
	WinCoin           int64
	WinTimes          int32
}

func ConvertPlayerDataToWebData(player *PlayerData) *webapi.PlayerData {
	if player == nil {
		return nil
	}
	pdfw := new(webapi.PlayerData)
	pdfw.Coin = player.Coin
	pdfw.Diamond = player.Diamond
	pdfw.CoinExchangeTotal = player.CoinExchangeTotal
	pdfw.CoinPayTotal = player.CoinPayTotal
	pdfw.CreateTime = player.CreateTime.Unix()
	pdfw.DeviceId = player.DeviceId
	pdfw.DeviceOS = player.DeviceOS
	pdfw.DrawTimes = player.DrawTimes
	pdfw.FailCoin = player.FailCoin
	pdfw.FailTimes = player.FailTimes
	pdfw.GameTax = player.GameTax
	pdfw.Ip = player.Ip
	pdfw.IsRob = player.IsRob
	pdfw.LastLoginTime = player.LastLoginTime.Unix()
	pdfw.MarkInfo = player.MarkInfo
	pdfw.Name = player.Name
	pdfw.Package = player.Package
	pdfw.PackageID = player.PackageID
	pdfw.Platform = player.Platform
	pdfw.RegIp = player.RegIp
	pdfw.SafeBoxCoin = player.SafeBoxCoin
	pdfw.SnId = player.SnId
	pdfw.Tel = player.Tel
	pdfw.VIP = player.VIP
	pdfw.WBLevel = player.WBLevel
	pdfw.WinCoin = player.WinCoin
	pdfw.WinTimes = player.WinTimes
	pdfw.BlacklistType = player.BlacklistType
	if player.Roles != nil {
		pdfw.RoleUnlock = int32(len(player.Roles.ModUnlock))
		for k := range player.Roles.ModUnlock {
			pdfw.RolesIds = append(pdfw.RolesIds, k)
		}
	}
	if player.Pets != nil {
		pdfw.PetUnlock = int32(len(player.Pets.ModUnlock))
		for k := range player.Pets.ModUnlock {
			pdfw.PetsIds = append(pdfw.PetsIds, k)
		}
	}
	return pdfw
}
func (this *PlayerData) IsMarkFlag(flag int) bool {
	return this.Flags&(1<<flag) != 0
}

func (this *PlayerData) MarkFlag(flag int) {
	this.Flags |= (1 << flag)
}

func (this *PlayerData) UnmarkFlag(flag int) {
	this.Flags &= ^(1 << flag)
}

func (this *PlayerData) GetSnId() int32 {
	return this.SnId
}

func (this *PlayerData) GetCoin() int64 {
	return this.Coin
}

func (this *PlayerData) GetTotalCoin() int64 {
	return this.Coin + this.SafeBoxCoin //+ this.YebData.Balance
}

func (this *PlayerData) IsRobot() bool {
	return this.IsRob
}

func (this *PlayerData) SetPayTs(payts int64) {
	if this.CoinPayTs < payts {
		this.CoinPayTs = payts
	}
}

func (this *PlayerData) SetSafeBoxPayTs(payts int64) {
	if this.SafeBoxCoinTs < payts {
		this.SafeBoxCoinTs = payts
	}
}

//func (pd *PlayerData) SetClubPayTs(payts int64) {
//	if pd.ClubCoinPayTs < payts {
//		pd.ClubCoinPayTs = payts
//	}
//}

func (this *PlayerData) SetTicketPayTs(payts int64) {
	if this.TicketPayTs < payts {
		this.TicketPayTs = payts
	}
}
func (this *PlayerData) GetTodayCoinFlowTotal() int64 {
	if this.LastLogoutTime.Day() != time.Now().Day() {
		return 0
	}
	if this.LastLogoutTime.Day() == time.Now().Day() && time.Now().Sub(this.LastLogoutTime) >= time.Hour*24 {
		return 0
	}
	if this.TodayGameData != nil {
		return this.TodayGameData.TodayConvertibleFlow
	}
	return 0
}

//	func (this *PlayerData) QueryFlowListLog(page int32) *FlowInfo {
//		fliwList := new(FlowInfo)
//		var err error
//		if page <= 0 {
//			//查库,兼容以前的协议
//			fliwList, err = GetFlowList(this.SnId, this.LastFlowTime)
//		} else {
//			fliwList, err = GetFlowListByPage(this.SnId, page, this.LastFlowTime)
//		}
//		if err != nil {
//			logger.Logger.Error("QueryFlowListLog is error: ", err)
//			return nil
//		}
//		if fliwList == nil {
//			return nil
//		}
//		return fliwList
//	}
func (this *PlayerData) UpdateParams(params string) *PlayerParams {
	if this.WhiteFlag > 0 {
		return nil
	}

	this.Params = params
	var pp PlayerParams
	err := json.Unmarshal([]byte(params), &pp)
	if err == nil {
		if common.IsValidIP(pp.Ip) {
			this.Ip = pp.Ip
		} else if common.IsValidIP(this.RegIp) {
			this.Ip = this.RegIp
		} else {
			this.Ip = ""
		}
		if pp.City != "" {
			this.City = pp.City
		}
		switch pp.Platform {
		case 0:
			this.DeviceOS = common.WebStr
		case 1:
			this.DeviceOS = common.IOSStr
		case 2:
			this.DeviceOS = common.AndroidStr
		}
		if pp.Name != "" {
			this.Name = pp.Name
		}
	}
	return &pp
}

func (this *PlayerData) GetPlayerDataEncoder() (bytes.Buffer, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(this)
	return buf, err
}

func NewPlayerData(acc string, name string, id int32, channel, platform, promoter string, inviterId,
	promoterTree int32, params, tel string, packTag, ip string, addCoin int64, unionid, deviceInfo string, subPromoter string, tagkey int32) *PlayerData {
	if len(name) == 0 {
		logger.Logger.Trace("New player name is empty.")
		return nil
	}
	//logger.Logger.Trace("New player information.")
	//safebox password default is '111111'
	raw := fmt.Sprintf("%v%v", DEFAULT_PLAYER_SAFEBOX_PWD, common.GetAppId())
	h := md5.New()
	io.WriteString(h, raw)
	pwd := hex.EncodeToString(h.Sum(nil))
	tNow := time.Now()
	isRobot := channel == common.Channel_Rob

	pd := &PlayerData{
		Id:           bson.NewObjectId(),
		AccountId:    acc,
		Name:         name,
		Channel:      channel,
		Platform:     platform,
		SnId:         id,
		InviterId:    inviterId,
		PromoterTree: promoterTree,
		Head:         rand.Int31n(6) + 1,
		Coin:         int64(GameParamData.NewPlayerCoin) + addCoin,
		//Coin:                int64(500000),
		SafeBoxPassword:     pwd,
		Ip:                  ip,
		RegIp:               ip,
		Params:              params,
		Tel:                 tel,
		SubBeUnderAgentCode: subPromoter,
		AgentType:           0,
		LastLoginTime:       tNow.Local(),
		LastLogoutTime:      tNow.AddDate(0, 0, -1).Local(),
		CreateTime:          tNow.Local(),
		Ver:                 VER_PLAYER_MAX - 1,
		HeadOutLine:         1,
		VIP:                 0,
		CoinPayTotal:        0,
		IsRob:               isRobot,
		BeUnderAgentCode:    promoter,
		PackageID:           packTag,
		WBLevel:             0,
		WBCoinTotalOut:      0,
		WBCoinTotalIn:       0,
		WBCoinLimit:         0,
		YesterdayGameData:   NewPlayerGameCtrlData(),
		TodayGameData:       NewPlayerGameCtrlData(),
		TotalGameData:       make(map[int][]*PlayerGameTotal),
		GDatas:              make(map[string]*PlayerGameInfo),
		TagKey:              tagkey,
		ShopTotal:           make(map[int32]*ShopTotal),
		ShopLastLookTime:    make(map[int32]int64),
	}

	if tel != "" {
		pd.UpgradeTime = time.Now()
	}

	pd.InitNewData(params)

	return pd
}
func NewPlayerDataThird(acc string, name, headUrl string, id int32, channel, platform, promoter string, inviterId,
	promoterTree int32, params, tel string, packTag, ip string, subPromoter string, tagkey int32) *PlayerData {
	if len(name) == 0 {
		logger.Logger.Trace("New player name is empty.")
		return nil
	}
	tNow := time.Now()
	pd := &PlayerData{
		Id:                  bson.NewObjectId(),
		AccountId:           acc,
		Name:                name,
		Channel:             channel,
		Platform:            platform,
		SnId:                id,
		InviterId:           inviterId,
		PromoterTree:        promoterTree,
		Head:                rand.Int31n(6) + 1,
		HeadUrl:             headUrl,
		Coin:                int64(GameParamData.NewPlayerCoin),
		Ip:                  ip,
		RegIp:               ip,
		Params:              params,
		Tel:                 tel,
		SubBeUnderAgentCode: subPromoter,
		LastLoginTime:       tNow.Local(),
		LastLogoutTime:      tNow.AddDate(0, 0, -1).Local(),
		CreateTime:          tNow.Local(),
		Ver:                 VER_PLAYER_MAX - 1,
		HeadOutLine:         1,
		IsRob:               false,
		BeUnderAgentCode:    promoter,
		PackageID:           packTag,
		YesterdayGameData:   NewPlayerGameCtrlData(),
		TodayGameData:       NewPlayerGameCtrlData(),
		TotalGameData:       make(map[int][]*PlayerGameTotal),
		GDatas:              make(map[string]*PlayerGameInfo),
		TagKey:              tagkey,
		ShopTotal:           make(map[int32]*ShopTotal),
		ShopLastLookTime:    make(map[int32]int64),
	}

	pd.InitNewData(params)

	return pd
}
func (this *PlayerData) InitNewData(params string) {
	this.TotalCoin = this.GetTotalCoin()
	//pd.ProfitCoin = pd.TotalCoin - pd.CoinPayTotal
	//0:男 1:女
	this.Sex = (this.Head%2 + 1) % 2
	//更新参数
	this.UpdateParams(params)
	//生成校验和
	RecalcuPlayerCheckSum(this)

	this.InitRolesAndPets()
}

func (this *PlayerData) InitRolesAndPets() {
	if this.Roles == nil || len(this.Roles.ModUnlock) == 0 {
		this.Roles = new(RolePetInfo)
		this.Roles.ModUnlock = make(map[int32]int32)
		this.Roles.ModUnlock[2000001] = 1
		this.Roles.ModId = 2000001
	}
	if this.Pets == nil {
		this.Pets = new(RolePetInfo)
		this.Pets.ModUnlock = make(map[int32]int32)
		this.Pets.ModUnlock[1000001] = 1
		this.Pets.ModId = 1000001
	}
	if _, ok := this.Pets.ModUnlock[1000001]; !ok {
		this.Pets.ModUnlock[1000001] = 1
		this.Pets.ModId = 1000001
	}
}

//func SavePlayerRebate(pd *PlayerData, thirdName string) error {
//	if pd == nil {
//		return nil
//	}
//	if rpcCli == nil {
//		return nil
//	}
//	var ret bool
//	err := rpcCli.CallWithTimeout("PlayerDataSvc.SavePlayerRebate", pd, &ret, time.Second*30)
//	if err != nil {
//		logger.Logger.Trace("SavePlayerRebate failed:", err)
//		return err
//	}
//	return nil
//}

type GetAgentInfoArgs struct {
	Plt string
	Tel string
}

// 获取代理信息
func GetAgentInfo(plt, tel string) *PlayerData {
	if rpcCli == nil {
		return nil
	}
	args := &GetAgentInfoArgs{
		Plt: plt,
		Tel: tel,
	}
	var pbi *PlayerData
	err := rpcCli.CallWithTimeout("PlayerDataSvc.GetAgentInfo", args, &pbi, time.Second*30)
	if err != nil {
		logger.Logger.Trace("GetPlayerBaseInfo failed:", err)
		return nil
	}
	return pbi
}

func ClonePlayerData(pd *PlayerData) *PlayerData {
	if pd == nil {
		return nil
	}

	pd.TotalCoin = pd.GetTotalCoin()
	//pd.ProfitCoin = pd.TotalCoin - pd.CoinPayTotal
	//增加可维护性，适度降低性能
	buf, err := json.Marshal(pd)
	if err != nil {
		logger.Logger.Warnf("ClonePlayerData %v json.Marshal fail:%v", pd.SnId, err)
		return nil
	}

	pdCopy := &PlayerData{}
	err = json.Unmarshal(buf, pdCopy)
	if err != nil {
		logger.Logger.Warnf("ClonePlayerData %v json.Unmarshal fail:%v", pd.SnId, err)
		return nil
	}
	return pdCopy
}

type GetPlayerDataArgs struct {
	Plt string
	Acc string
}
type PlayerDataRet struct {
	Pd    *PlayerData
	IsNew bool
}

func GetPlayerData(plt, acc string) (*PlayerData, bool) {
	if rpcCli == nil {
		return nil, false
	}
	args := &GetPlayerDataArgs{
		Plt: plt,
		Acc: acc,
	}
	var ret PlayerDataRet
	err := rpcCli.CallWithTimeout("PlayerDataSvc.GetPlayerData", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Trace("GetPlayerData failed:", err)
		return nil, false
	}
	return ret.Pd, ret.IsNew
}

type CreatePlayer struct {
	Plt      string
	AccId    string
	NickName string
	HeadUrl  string
}

func CreatePlayerDataByThird(plt, acc string, nickName, headUrl string) (*PlayerData, bool) {
	if rpcCli == nil {
		return nil, false
	}
	var args = &CreatePlayer{
		Plt:      plt,
		AccId:    acc,
		NickName: nickName,
		HeadUrl:  headUrl,
	}
	var ret = &PlayerDataRet{}
	err := rpcCli.CallWithTimeout("PlayerDataSvc.CreatePlayerDataByThird", args, ret, time.Second*30)
	if err != nil {
		logger.Logger.Trace("CreatePlayerDataByThird failed:", err)
		return nil, false
	}
	return ret.Pd, ret.IsNew
}

type PlayerDataArg struct {
	Plt     string
	AccId   string
	AddCoin int32
}

func CreatePlayerDataOnRegister(plt, acc string, addCoin int32) (*PlayerData, bool) {
	if rpcCli == nil {
		return nil, false
	}

	var args = &PlayerDataArg{
		Plt:     plt,
		AccId:   acc,
		AddCoin: addCoin,
	}
	var ret = &PlayerDataRet{}
	err := rpcCli.CallWithTimeout("PlayerDataSvc.CreatePlayerDataOnRegister", args, ret, time.Second*30)
	if err != nil {
		logger.Logger.Trace("CreatePlayerDataOnRegister failed:", err)
		return nil, false
	}
	return ret.Pd, true
}

type GetPlayerDataBySnIdArgs struct {
	Plt              string
	SnId             int32
	CorrectData      bool
	CreateIfNotExist bool
}

func GetPlayerDataBySnId(plt string, snid int32, correctData, createIfNotExist bool) (*PlayerData, bool) {
	if rpcCli == nil {
		return nil, false
	}
	args := &GetPlayerDataBySnIdArgs{
		Plt:              plt,
		SnId:             snid,
		CorrectData:      correctData,
		CreateIfNotExist: createIfNotExist,
	}
	var ret = &PlayerDataRet{}
	err := rpcCli.CallWithTimeout("PlayerDataSvc.GetPlayerDataBySnId", args, ret, time.Second*30)
	if err != nil {
		logger.Logger.Tracef("Get %v %v player data error:%v", plt, snid, err)
		return nil, false
	}
	return ret.Pd, ret.IsNew
}

//func UpdateCreateCreateClubNum(snId int32) error {
//	if rpcCli == nil {
//		return nil
//	}
//	var ret = &PlayerDataRet{}
//	err := rpcCli.CallWithTimeout("PlayerDataSvc.UpdateCreateCreateClubNum", snId, ret, time.Second*30)
//	if err != nil {
//		logger.Logger.Trace("Update player safeboxcoin error:", err)
//		return err
//	}
//	return nil
//}

type GetPlayerDatasBySnIdsArgs struct {
	Plt         string
	SnIds       []int32
	CorrectData bool
}

func GetPlayerDatasBySnIds(plt string, snIds []int32, correctData bool) []*PlayerData {
	if rpcCli == nil {
		return nil
	}
	args := &GetPlayerDatasBySnIdsArgs{
		Plt:         plt,
		SnIds:       snIds,
		CorrectData: correctData,
	}
	var ret = []*PlayerData{}
	err := rpcCli.CallWithTimeout("PlayerDataSvc.GetPlayerDatasBySnIds", args, ret, time.Second*30)
	if err != nil {
		logger.Logger.Tracef("GetPlayerDatasBySnIds(snids=%v) error:%v", args, err)
		return nil
	}

	return ret
}

type GetPlayerDataByUnionIdArgs struct {
	Plt         string
	UnionId     string
	CorrectData bool
}

func GetPlayerDataByUnionId(plt, unionid string, correctData bool) (*PlayerData, bool) {
	if rpcCli == nil {
		return nil, false
	}
	args := &GetPlayerDataByUnionIdArgs{
		Plt:         plt,
		UnionId:     unionid,
		CorrectData: correctData,
	}
	var ret = &PlayerDataRet{}
	err := rpcCli.CallWithTimeout("PlayerDataSvc.GetPlayerDataByUnionId", args, ret, time.Second*30)
	if err != nil {
		logger.Logger.Tracef("GetPlayerDataByUnionId % player data error:%v", args, err)
		return nil, false
	}
	return ret.Pd, true
}

type GetPlayerTelArgs struct {
	Plt    string
	Tel    string
	TagKey int32
}
type GetPlayerTelRet struct {
	SnId int32
	Err  error
}

func GetPlayerTel(tel, platform string, tagkey int32) (int32, string) {
	if rpcCli == nil {
		return 0, "no find account"
	}
	args := &GetPlayerTelArgs{
		Tel:    tel,
		Plt:    platform,
		TagKey: tagkey,
	}
	ret := &GetPlayerTelRet{}
	err := rpcCli.CallWithTimeout("PlayerDataSvc.GetPlayerTel", args, ret, time.Second*30)
	if err != nil {
		logger.Logger.Warn("playerdata.Find err:", err)
		return 0, err.Error()
	}
	return ret.SnId, ""
}

type GetPlayerCoinArgs struct {
	Plt  string
	SnId int32
}
type GetPlayerCoinRet struct {
	Coin        int64
	SafeBoxCoin int64
	Err         error
}

func GetPlayerCoin(plt string, snid int32) (int64, int64, error) {
	if rpcCli == nil {
		return 0, 0, fmt.Errorf("db may be closed")
	}

	args := GetPlayerCoinArgs{
		Plt:  plt,
		SnId: snid,
	}
	var ret = &GetPlayerCoinRet{}
	err := rpcCli.CallWithTimeout("PlayerDataSvc.GetPlayerCoin", args, ret, time.Second*30)
	if err != nil {
		logger.Logger.Trace("GetPlayerCoin error:", err)
		return 0, 0, err
	}
	if ret.Err != nil {
		return 0, 0, ret.Err
	}
	return ret.Coin, ret.SafeBoxCoin, nil
}

func ExtractPlayerPackageName(packageTag string) string {
	segs := strings.Split(packageTag, ".")
	if len(segs) > 3 {
		packageName := strings.Join(segs[:3], ".")
		return packageName
	}
	return packageTag
}

/*
 * 保存玩家的全部信息
 */
func SavePlayerData(pd *PlayerData) bool {
	if rpcCli == nil {
		return false
	}
	if pd != nil {
		var ret bool
		err := rpcCli.CallWithTimeout("PlayerDataSvc.SavePlayerData", pd, &ret, time.Second*30)
		if err != nil {
			logger.Logger.Errorf("SavePlayerData %v err:%v", pd.SnId, err)
			return false
		}
		return ret
	}
	return false
}

func BackupPlayerData(pd *PlayerData) {
	if pd == nil {
		return
	}

	buf, err := json.Marshal(pd)
	if err != nil {
		logger.Logger.Info("BackupPlayerInfo json.Marshal error", err)
		return
	}

	fileName := fmt.Sprintf("%v.json", pd.AccountId)
	err = ioutil.WriteFile(fileName, buf, os.ModePerm)
	if err != nil {
		logger.Logger.Info("BackupPlayerInfo ioutil.WriteFile error", err)
	}
}

func RestorePlayerData(fileName string) *PlayerData {
	pd := &PlayerData{}
	buf, err := ioutil.ReadFile(fileName)
	if err != nil {
		logger.Logger.Info("RestorePlayerInfo ioutil.ReadFile error", err)
		return nil
	}

	err = json.Unmarshal(buf, pd)
	if err != nil {
		logger.Logger.Info("RestorePlayerInfo json.Unmarshal error", err)
	}
	return pd
}

type RemovePlayerArgs struct {
	Plt  string
	SnId int32
}

func RemovePlayer(plt string, snid int32) error {
	if rpcCli == nil {
		return nil
	}
	args := &RemovePlayerArgs{
		Plt:  plt,
		SnId: snid,
	}
	var ret bool
	err := rpcCli.CallWithTimeout("PlayerDataSvc.RemovePlayer", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Info("Remove player failed.")
		return err
	}
	return nil
}

type RemovePlayerByAccArgs struct {
	Plt string
	Acc string
}

func RemovePlayerByAcc(plt, acc string) error {
	if rpcCli == nil {
		return nil
	}
	args := &RemovePlayerByAccArgs{
		Plt: plt,
		Acc: acc,
	}
	var ret bool
	err := rpcCli.CallWithTimeout("PlayerDataSvc.RemovePlayerByAcc", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Info("Remove player failed.")
		return err
	}
	return nil
}

/*
 * 检查手机号是否存在
 */
type GetPlayerSnidArgs struct {
	Plt string
	Acc string
}
type GetPlayerSnidRet struct {
	SnId int32
}

func GetPlayerSnid(plt, acc string) int32 {
	if rpcCli == nil {
		return 0
	}
	args := &GetPlayerSnidArgs{
		Plt: plt,
		Acc: acc,
	}
	var ret int32
	err := rpcCli.CallWithTimeout("PlayerDataSvc.GetPlayerSnid", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Error("model.GetPlayerSnid is error ", err)
		return 0
	}
	return ret
}

/*
 * 检查昵称是否存在
 */
type PlayerNickIsExistArgs struct {
	Plt  string
	Name string
}

func PlayerNickIsExist(plt, name string) bool {
	if rpcCli == nil {
		return false
	}
	args := &PlayerNickIsExistArgs{
		Plt:  plt,
		Name: name,
	}
	var ret bool
	err := rpcCli.CallWithTimeout("PlayerDataSvc.PlayerNickIsExist", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Error("model.PlayerNickIsExist is error ", err)
		return false
	}
	if err != nil {
		return false
	}

	return ret
}

/*
 * 检查玩家是否存在
 */
type PlayerIsExistBySnIdArgs struct {
	Plt  string
	SnId int32
}

func PlayerIsExistBySnId(plt string, snId int32) bool {
	if rpcCli == nil {
		return false
	}
	args := &PlayerIsExistBySnIdArgs{
		Plt:  plt,
		SnId: snId,
	}
	var ret bool
	err := rpcCli.CallWithTimeout("PlayerDataSvc.PlayerIsExistBySnId", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Error("model.PlayerIsExistBySnId is error", err)
		return false
	}

	return ret
}

type UpdatePackageId struct {
	AccId        string
	SnId         int32
	Tag          string
	Platform     int32
	Channel      int32
	Promoter     int32
	PlatformStr  string
	ChannelStr   string
	PromoterStr  string
	Inviterid    int32
	PromoterTree int32
	PackageTag   string
}

// 修改推广包标识
func UpdatePlayerPackageId(snid int32, tag string, platform, channel, promoter, inviterid, promoterTree int32) error {
	if rpcCli == nil {
		return errors.New("user_playerinfo not open")
	}
	var args = &UpdatePackageId{
		SnId:         snid,
		Tag:          tag,
		Platform:     platform,
		Channel:      channel,
		Promoter:     promoter,
		Inviterid:    inviterid,
		PromoterTree: promoterTree,
	}
	var ret bool
	err := rpcCli.CallWithTimeout("PlayerDataSvc.UpdatePlayerPackageId", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Trace("Update player packageid error:", err)
		return err
	}
	return nil
}
func UpdatePlayerPackageIdByStr(snid int32, tag string, platform, channel, promoter string, inviterid, promoterTree int32) error {
	if rpcCli == nil {
		return errors.New("user_playerinfo not open")
	}
	var args = &UpdatePackageId{
		SnId:         snid,
		Tag:          tag,
		PlatformStr:  platform,
		ChannelStr:   channel,
		PromoterStr:  promoter,
		Inviterid:    inviterid,
		PromoterTree: promoterTree,
	}
	var ret bool
	err := rpcCli.CallWithTimeout("PlayerDataSvc.UpdatePlayerPackageIdByStr", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Trace("UpdatePlayerPackageIdByStr error:", err)
		return err
	}
	return nil
}

func UpdatePlayerPackageIdByAcc(accid string, tag string, platform, channel, promoter string, inviterid, promoterTree int32) error {
	if rpcCli == nil {
		return errors.New("user_playerinfo not open")
	}
	var args = &UpdatePackageId{
		AccId:        accid,
		Tag:          tag,
		PlatformStr:  platform,
		ChannelStr:   channel,
		PromoterStr:  promoter,
		Inviterid:    inviterid,
		PromoterTree: promoterTree,
	}
	var ret bool
	err := rpcCli.CallWithTimeout("PlayerDataSvc.UpdatePlayerPackageIdByAcc", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Trace("UpdatePlayerPackageIdByAcc error:", err)
		return err
	}
	return nil
}

// 修改平台
func UpdatePlayerPlatform(snid int, platform, channel string) error {
	if rpcCli == nil {
		return errors.New("user_playerinfo not open")
	}
	var args = &UpdatePackageId{
		SnId:        int32(snid),
		PlatformStr: platform,
		ChannelStr:  channel,
	}
	var ret bool
	err := rpcCli.CallWithTimeout("PlayerDataSvc.UpdatePlayerPlatform", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Trace("Update player nick error:", err)
		return err
	}
	return nil
}

// 修改玩家无级代推广员id
func UpdatePlayerPromoterTree(plt string, snid int32, promoterTree int32) error {
	if rpcCli == nil {
		return errors.New("user_playerinfo not open")
	}
	var args = &UpdatePackageId{
		PlatformStr:  plt,
		SnId:         snid,
		PromoterTree: promoterTree,
	}
	var ret bool
	err := rpcCli.CallWithTimeout("PlayerDataSvc.UpdatePlayerPromoterTree", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Trace("UpdatePlayerPromoterTree error:", err)
		return err
	}
	return nil
}

// 修改玩家全民推广
func UpdatePlayerInviteID(plt string, snid int32, InviteID int32) error {
	if rpcCli == nil {
		return errors.New("user_playerinfo not open")
	}
	var args = &UpdatePackageId{
		PlatformStr: plt,
		SnId:        snid,
		Inviterid:   InviteID,
	}
	var ret bool
	err := rpcCli.CallWithTimeout("PlayerDataSvc.UpdatePlayerInviteID", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Trace("UpdatePlayerInviteID error:", err)
		return err
	}
	return nil
}

type UpdatePlayerInfo struct {
	Plt               string
	Acc               string
	SnId              int32
	Nick              string
	BlacklistType     int32
	MarkInfo          string
	WhiteFlag         int32
	PayActState       map[int32]*PayActState
	TelephonePromoter int32
	TelephoneCallNum  int32
	Head              int32
	Sex               int32
}

/*
 * 修改昵称
 */
func UpdatePlayerNick(plt, acc string, nick string) int {
	if rpcCli == nil {
		return 2
	}
	var args = &UpdatePlayerInfo{
		Plt:  plt,
		Acc:  acc,
		Nick: nick,
	}
	var ret bool
	err := rpcCli.CallWithTimeout("PlayerDataSvc.UpdatePlayerNick", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Trace("Update player nick error:", err)
		return 1
	}
	if !ret {
		return 1
	}
	return 0
}

/*
 * 修改黑名单类型
 */
type UpdatePlayerBlacklistTypeArgs struct {
	Plt           string
	SnId          int32
	BlackListType int32
}

func UpdatePlayerBlacklistType(plt string, snid int32, blacklisttype int32) {
	if rpcCli == nil {
		return
	}
	var args = &UpdatePlayerBlacklistTypeArgs{
		Plt:           plt,
		SnId:          snid,
		BlackListType: blacklisttype,
	}
	var ret bool
	err := rpcCli.CallWithTimeout("PlayerDataSvc.UpdatePlayerBlacklistType", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Trace("Update player blacklisttype error:", err)
		return
	}
}

/*
 * 修改玩家备注信息
 */
type UpdatePlayerMarkInfoArgs struct {
	Plt      string
	SnId     int32
	MarkInfo string
}

func UpdatePlayerMarkInfo(plt string, snid int32, mark string) error {
	if rpcCli == nil {
		return errors.New("user_playerinfo not open")
	}
	var args = &UpdatePlayerMarkInfoArgs{
		Plt:      plt,
		SnId:     snid,
		MarkInfo: mark,
	}
	var ret bool
	err := rpcCli.CallWithTimeout("PlayerDataSvc.UpdatePlayerMarkInfo", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Warnf("Update player mark error:", err)
		return err
	}
	return nil
}

/*
 * 修改玩家特殊白名单
 */
type UpdatePlayerWhiteFlagArgs struct {
	Plt       string
	SnId      int32
	WhiteFlag int32
}

func UpdatePlayerWhiteFlag(plt string, snid int32, whiteFlag int32) error {
	if rpcCli == nil {
		return errors.New("user_playerinfo not open")
	}
	var args = &UpdatePlayerWhiteFlagArgs{
		Plt:       plt,
		SnId:      snid,
		WhiteFlag: whiteFlag,
	}
	var ret bool
	err := rpcCli.CallWithTimeout("PlayerDataSvc.UpdatePlayerWhiteFlag", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Warnf("Update player whiteFlag error:", err)
		return err
	}
	return nil
}

/*
 * 修改玩家支付信息
 */

type UpdatePlayerPayActArgs struct {
	Plt         string
	SnId        int32
	PayActState map[int32]*PayActState
}

func UpdatePlayerPayAct(plt string, snid int32, player map[int32]*PayActState) error {
	if rpcCli == nil {
		return errors.New("user_playerinfo not open")
	}
	var args = &UpdatePlayerPayActArgs{
		Plt:         plt,
		SnId:        snid,
		PayActState: player,
	}
	var ret bool
	err := rpcCli.CallWithTimeout("PlayerDataSvc.UpdatePlayerPayAct", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Warnf("Update player payact error:", err)
		return err
	}
	return nil
}

/*
 * 修改玩家电销标记
 */
type UpdatePlayerTelephonePromoterArgs struct {
	Plt               string
	SnId              int32
	TelephonePromoter string
	TelephoneCallNum  int32
}

func UpdatePlayerTelephonePromoter(plt string, snid int32, mark string) error {
	if rpcCli == nil {
		return errors.New("user_playerinfo not open")
	}
	var args = &UpdatePlayerTelephonePromoterArgs{
		Plt:               plt,
		SnId:              snid,
		TelephonePromoter: mark,
	}
	var ret bool
	err := rpcCli.CallWithTimeout("PlayerDataSvc.UpdatePlayerTelephonePromoter", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Warnf("Update player telephonepromoter error:", err)
		return err
	}
	return nil
}

/*
 * 修改玩家电销标记
 */
func UpdatePlayerTelephoneCallNum(plt string, snid int32, mark int32) error {
	if rpcCli == nil {
		return errors.New("user_playerinfo not open")
	}
	var args = &UpdatePlayerTelephonePromoterArgs{
		Plt:              plt,
		SnId:             snid,
		TelephoneCallNum: mark,
	}
	var ret bool
	err := rpcCli.CallWithTimeout("PlayerDataSvc.UpdatePlayerTelephoneCallNum", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Warnf("Update player telephonecallnum error:", err)
		return err
	}
	return nil
}

/*
 * 修改头像
 */
type UpdatePlayeIconArgs struct {
	Plt  string
	Acc  string
	Head int32
}

func UpdatePlayeIcon(plt, acc string, icon int32) int {
	if rpcCli == nil {
		return 2
	}
	var args = &UpdatePlayeIconArgs{
		Plt:  plt,
		Acc:  acc,
		Head: icon,
	}
	var ret bool
	err := rpcCli.CallWithTimeout("PlayerDataSvc.UpdatePlayeIcon", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Trace("Update player icon error:", err)
		return 1
	}
	if !ret {
		return 1
	}
	return 0
}

/*
 * 修改性别
 */
type UpdatePlayeSexArgs struct {
	Plt string
	Acc string
	Sex int32
}

func UpdatePlayeSex(plt, acc string, sex int32) int {
	if rpcCli == nil {
		return 2
	}
	var args = &UpdatePlayeSexArgs{
		Plt: plt,
		Acc: acc,
		Sex: sex,
	}
	var ret bool
	err := rpcCli.CallWithTimeout("PlayerDataSvc.UpdatePlayeSex", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Trace("Update player sex error:", err)
		return 1
	}
	if !ret {
		return 1
	}
	return 0
}

/*
 * 检查手机号是否存在
 */
type PlayerTelIsExistArgs struct {
	Tel      string
	Platform string
	TagKey   int32
}

func PlayerTelIsExist(tel, platform string, tagkey int32) bool {
	if rpcCli == nil {
		return false
	}
	var args = &PlayerTelIsExistArgs{
		Tel:      tel,
		Platform: platform,
		TagKey:   tagkey,
	}
	var ret bool
	err := rpcCli.CallWithTimeout("PlayerDataSvc.PlayerTelIsExist", args, &ret, time.Second*30)
	if err != nil {
		return false
	}

	return ret
}

/*
 * 绑定手机号
 */
type UpdatePlayerTelArgs struct {
	Plt  string
	SnId int32
	Tel  string
}

func UpdatePlayerTel(plt string, snid int32, tel string) error {
	if rpcCli == nil {
		return errors.New("user_playerinfo not open")
	}
	var args = &UpdatePlayerTelArgs{
		Plt:  plt,
		SnId: snid,
		Tel:  tel,
	}
	var ret bool
	err := rpcCli.CallWithTimeout("PlayerDataSvc.UpdatePlayerTel", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Warn("UpdatePlayerTel error:", err)
		return err
	}
	return nil
}

type CountAlipayAccountCountArgs struct {
	Platform      string
	AliPayAccount string
}

func CountAlipayAccountCount(platform, alipayAccount string) int {
	if rpcCli == nil {
		return 0
	}
	var args = &CountAlipayAccountCountArgs{
		Platform:      platform,
		AliPayAccount: alipayAccount,
	}
	var count int
	err := rpcCli.CallWithTimeout("PlayerDataSvc.CountAlipayAccountCount", args, &count, time.Second*30)
	if err != nil {
		logger.Logger.Warn("CountAlipayAccountCount error:", err)
		return 0
	}
	return count
}

type CountBankAlipayNameCountArgs struct {
	SnId          int32
	Platform      string
	AliPayAccName string
	BankAccName   string
}

func CountBankAlipayNameCount(platform, name string, snid int32) int {
	if rpcCli == nil {
		return 0
	}
	var args = &CountBankAlipayNameCountArgs{
		SnId:          snid,
		Platform:      platform,
		AliPayAccName: name,
		BankAccName:   name,
	}
	var ret int
	err := rpcCli.CallWithTimeout("PlayerDataSvc.CountBankAlipayNameCount", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Warn("CountBankAlipayNameCount error:", err)
		return 0
	}
	return ret
}

/*
 * 绑定支付宝账号
 */
type UpdatePlayerAlipayArgs struct {
	SnId          int32
	Platform      string
	AliPayAccount string
	AliPayAccName string
}

func UpdatePlayerAlipay(plt string, snid int32, alipayAccount, alipayAccName string) error {
	if rpcCli == nil {
		return errors.New("user_playerinfo not open")
	}
	var args = &UpdatePlayerAlipayArgs{
		Platform:      plt,
		SnId:          snid,
		AliPayAccount: alipayAccount,
		AliPayAccName: alipayAccName,
	}
	var ret bool
	err := rpcCli.CallWithTimeout("PlayerDataSvc.UpdatePlayerAlipay", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Warn("UpdatePlayerAlipay error:", err)
		return err
	}
	return err
}
func UpdatePlayerAlipayAccount(plt string, snid int, alipayAccount string) error {
	if rpcCli == nil {
		return errors.New("user_playerinfo not open")
	}
	var args = &UpdatePlayerAlipayArgs{
		Platform:      plt,
		SnId:          int32(snid),
		AliPayAccount: alipayAccount,
	}
	var ret bool
	err := rpcCli.CallWithTimeout("PlayerDataSvc.UpdatePlayerAlipayAccount", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Warn("UpdatePlayerAlipay error:", err)
		return err
	}
	return nil
}
func UpdatePlayerAlipayName(plt string, snid int, alipayAccName string) error {
	if rpcCli == nil {
		return errors.New("user_playerinfo not open")
	}
	var args = &UpdatePlayerAlipayArgs{
		Platform:      plt,
		SnId:          int32(snid),
		AliPayAccName: alipayAccName,
	}
	var ret bool
	err := rpcCli.CallWithTimeout("PlayerDataSvc.UpdatePlayerAlipayName", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Warn("UpdatePlayerAlipay error:", err)
		return err
	}
	return nil
}

type CountBankAccountCountArgs struct {
	Platform    string
	BankAccount string
}

func CountBankAccountCount(platform, bankAccount string) int {
	if rpcCli == nil {
		return 0
	}
	var args = &CountBankAccountCountArgs{
		Platform:    platform,
		BankAccount: bankAccount,
	}
	var ret int
	err := rpcCli.CallWithTimeout("PlayerDataSvc.CountBankAccountCount", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Warn("CountBankAccountCount error:", err)
		return 0
	}
	return ret
}

/*
 * 绑定银行账号
 */
type UpdatePlayerBankArgs struct {
	Plt         string
	SnId        int32
	Bank        string
	BankAccount string
	BankAccName string
}

func UpdatePlayerBank(plt string, snid int32, bank, bankAccount, bankAccName string) error {
	if rpcCli == nil {
		return errors.New("user_playerinfo not open")
	}
	var args = &UpdatePlayerBankArgs{
		Plt:         plt,
		SnId:        snid,
		Bank:        bank,
		BankAccount: bankAccount,
		BankAccName: bankAccName,
	}
	var ret bool
	err := rpcCli.CallWithTimeout("PlayerDataSvc.UpdatePlayerBank", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Warn("UpdatePlayerBank error:", err)
		return err
	}
	return nil
}

/*
 * 修改玩家是否返利
 */
type UpdatePlayerIsRebateArgs struct {
	Plt         string
	SnId        int32
	IsCanRebate int32
}

func UpdatePlayerIsRebate(plt string, snid int32, isRebate int32) int {
	if rpcCli == nil {
		return 2
	}
	var args = &UpdatePlayerIsRebateArgs{
		Plt:         plt,
		SnId:        snid,
		IsCanRebate: isRebate,
	}
	var ret bool
	err := rpcCli.CallWithTimeout("PlayerDataSvc.UpdatePlayerIsRebate", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Trace("Update player isRebate error:", err)
		return 1
	}

	if !ret {
		return 1
	}
	return 0
}

/*
 * 修改玩家是否可以修改昵称
 */
type UpdatePlayerIsStopRenameArgs struct {
	Plt          string
	SnId         int32
	IsStopReName int32
}

func UpdatePlayerIsStopRename(plt string, snid int32, isStop int32) error {
	if rpcCli == nil {
		return errors.New("param err")
	}
	var args = &UpdatePlayerIsStopRenameArgs{
		Plt:          plt,
		SnId:         snid,
		IsStopReName: isStop,
	}
	var ret bool
	err := rpcCli.CallWithTimeout("PlayerDataSvc.UpdatePlayerIsStopRename", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Trace("Update player isRename error:", err)
		return err
	}
	return nil
}

type PlayerRebindSnIdArgs struct {
	Plt     string
	SnId    int32
	NewSnId int32
}

func PlayerRebindSnId(plt string, snid int32, newSnId int32) error {
	if snid == 0 || newSnId == 0 {
		return errors.New("param err")
	}
	if rpcCli == nil {
		return errors.New("rpcCli == nil")
	}
	var args = &PlayerRebindSnIdArgs{
		Plt:     plt,
		SnId:    snid,
		NewSnId: newSnId,
	}
	var ret bool
	err := rpcCli.CallWithTimeout("PlayerDataSvc.PlayerRebindSnId", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Trace("PlayerRebindSnId error:", err)
		return err
	}
	return nil
}

type PlayerSafeBoxArgs struct {
	Acc         string
	OldPassWord string
	PassWord    string
	Tel         string
	Platform    string
	TagKey      int32
	SnId        int32
}

/*
 * 修改保险箱密码
 */
func UpdateSafeBoxPassword(plt, acc string, oldpassword string, password string) int {
	if rpcCli == nil {
		return 2
	}
	var args = &PlayerSafeBoxArgs{
		Platform:    plt,
		Acc:         acc,
		OldPassWord: oldpassword,
		PassWord:    password,
	}
	var ret bool
	err := rpcCli.CallWithTimeout("PlayerDataSvc.UpdateSafeBoxPassword", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Trace("Update player safeboxpassword error:", err)
		return 1
	}
	if !ret {
		return 1
	}
	return 0
}

/*
 * 找回保险箱密码
 */
func GetBackSafeBoxPassword(tel string, password, platform string, tagkey int32) error {
	if rpcCli == nil {
		return errors.New("GetBackSafeBoxPassword error")
	}
	var args = &PlayerSafeBoxArgs{
		Tel:      tel,
		PassWord: password,
		Platform: platform,
		TagKey:   tagkey,
	}
	var ret bool
	err := rpcCli.CallWithTimeout("PlayerDataSvc.GetBackSafeBoxPassword", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Trace("Update player safeboxpassword error:", err)
	}
	return err
}

/*
 * 重置保险箱密码
 */
func ResetSafeBoxPassword(plt string, snid int, password string) error {
	if rpcCli == nil {
		return errors.New("ResetSafeBoxPassword error")
	}
	var args = &PlayerSafeBoxArgs{
		Platform: plt,
		SnId:     int32(snid),
		PassWord: password,
	}
	var ret bool
	err := rpcCli.CallWithTimeout("PlayerDataSvc.ResetSafeBoxPassword", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Trace("Reset player safeboxpassword error:", err)
		return err
	}
	return err
}

type PlayerSafeBoxCoin struct {
	Plt                                                  string
	SnId                                                 int32
	Coin, Diamond, SafeBoxCoin, CoinPayTs, SafeBoxCoinTs int64
	LastexChangeTime                                     int64
	TotalConvertibleFlow                                 int64
	TotalFlow                                            int64
	LastexChangeOrder                                    string
}

func UpdatePlayerCoin(plt string, snid int32, coin, diamond, safeboxcoin, coinpayts, safeboxcoints int64) error {
	if rpcCli == nil {
		return fmt.Errorf("db may be closed")
	}
	var args = &PlayerSafeBoxCoin{
		Plt:           plt,
		SnId:          snid,
		Coin:          coin,
		Diamond:       diamond,
		SafeBoxCoin:   safeboxcoin,
		CoinPayTs:     coinpayts,
		SafeBoxCoinTs: safeboxcoints,
	}
	var ret bool
	err := rpcCli.CallWithTimeout("PlayerDataSvc.UpdatePlayerCoin", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Errorf("UpdatePlayerCoin param:%v-%v-%v", snid, coin, safeboxcoin)
		logger.Logger.Error("UpdatePlayerCoin error:", err)
		return err
	}
	return nil
}

func UpdatePlayerSetCoin(plt string, snid int32, coin int64) error {
	if rpcCli == nil {
		return fmt.Errorf("db may be closed")
	}
	var args = &PlayerSafeBoxCoin{
		Plt:  plt,
		SnId: snid,
		Coin: coin,
	}
	var ret bool
	err := rpcCli.CallWithTimeout("PlayerDataSvc.UpdatePlayerSetCoin", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Error("model.UpdatePlayerSetCoin error:", err)
		return err
	}
	return nil
}

func UpdatePlayerLastExchangeTime(plt string, snid int32, ts int64) error {
	if rpcCli == nil {
		return fmt.Errorf("db may be closed")
	}
	var args = &PlayerSafeBoxCoin{
		Plt:              plt,
		SnId:             snid,
		LastexChangeTime: ts,
	}
	var ret bool
	err := rpcCli.CallWithTimeout("PlayerDataSvc.UpdatePlayerLastExchangeTime", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Error("model.UpdatePlayerLastExchangeTime error:", err)
		return err
	}
	return nil
}

func UpdatePlayerExchageFlow(plt string, snid int32, flow int64, flow2 int64) error {
	if rpcCli == nil {
		return fmt.Errorf("db may be closed")
	}
	var args = &PlayerSafeBoxCoin{
		Plt:                  plt,
		SnId:                 snid,
		TotalConvertibleFlow: flow,
		TotalFlow:            flow2,
	}
	var ret bool
	err := rpcCli.CallWithTimeout("PlayerDataSvc.UpdatePlayerExchageFlow", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Error("model.UpdatePlayerExchageFlow error:", err)
		return err
	}
	return nil
}
func UpdatePlayerExchageFlowAndOrder(plt string, snid int32, flow int64, logid string) error {
	if rpcCli == nil {
		return fmt.Errorf("db may be closed")
	}
	var args = &PlayerSafeBoxCoin{
		Plt:                  plt,
		SnId:                 snid,
		TotalConvertibleFlow: flow,
		LastexChangeOrder:    logid,
	}
	var ret bool
	err := rpcCli.CallWithTimeout("PlayerDataSvc.UpdatePlayerExchageFlowAndOrder", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Error("model.UpdatePlayerExchageFlowAndOrder error:", err)
		return err
	}
	return nil
}

// 重要数据要做校验和
func RecalcuPlayerCheckSum(pd *PlayerData) {
	h := bytes.NewBuffer(nil)
	io.WriteString(h, pd.AccountId)
	io.WriteString(h, fmt.Sprintf("%v", pd.SnId))
	io.WriteString(h, fmt.Sprintf("%v", pd.Tel))
	io.WriteString(h, fmt.Sprintf("%v", pd.Coin))
	io.WriteString(h, fmt.Sprintf("%v", pd.SafeBoxCoin))
	io.WriteString(h, fmt.Sprintf("%v", pd.CoinPayTs))
	io.WriteString(h, fmt.Sprintf("%v", pd.SafeBoxCoinTs))
	io.WriteString(h, fmt.Sprintf("%v", pd.GameCoinTs))
	io.WriteString(h, common.GetAppId())
	pd.CheckSum = crc32.ChecksumIEEE(h.Bytes())
}

// 校验数据,防止数据被破坏或者爆库修改
func VerifyPlayerCheckSum(pd *PlayerData) bool {
	h := bytes.NewBuffer(nil)
	io.WriteString(h, pd.AccountId)
	io.WriteString(h, fmt.Sprintf("%v", pd.SnId))
	io.WriteString(h, fmt.Sprintf("%v", pd.Tel))
	io.WriteString(h, fmt.Sprintf("%v", pd.Coin))
	io.WriteString(h, fmt.Sprintf("%v", pd.SafeBoxCoin))
	io.WriteString(h, fmt.Sprintf("%v", pd.CoinPayTs))
	io.WriteString(h, fmt.Sprintf("%v", pd.SafeBoxCoinTs))
	io.WriteString(h, fmt.Sprintf("%v", pd.GameCoinTs))
	io.WriteString(h, common.GetAppId())
	checkSum := crc32.ChecksumIEEE(h.Bytes())
	return checkSum == pd.CheckSum
}

/*
 * 查找多用户
 */
type PlayerSelect struct {
	Id        int
	Tel       string
	NickName  string
	Coinl     int
	Coinh     int
	Ip        string
	Alipay    string
	Registerl int
	Registerh int
	Platform  string
	Channel   string
}

func FindPlayerList(id int, tel string, nickname string, coinl int, coinh int, ip string,
	alipay string, registerl int, registerh int, platform, channel string) []*PlayerBaseInfo {
	var args = &PlayerSelect{
		Id:        id,
		Tel:       tel,
		NickName:  nickname,
		Coinl:     coinl,
		Coinh:     coinh,
		Ip:        ip,
		Alipay:    alipay,
		Registerl: registerl,
		Registerh: registerh,
		Platform:  platform,
		Channel:   channel,
	}
	var ret []*PlayerBaseInfo
	err := rpcCli.CallWithTimeout("PlayerDataSvc.FindPlayerList", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Error("model.FindPlayerList error:", err)
		return nil
	}
	return ret
}

type UpdateElement struct {
	Plt       string
	SnId      int32
	PlayerMap string
}

func UpdatePlayerElement(plt string, id int32, playerMap map[string]interface{}) error {
	if rpcCli == nil {
		return fmt.Errorf("db may be close")
	}
	pm, _ := json.Marshal(playerMap)
	var args = &UpdateElement{
		Plt:       plt,
		SnId:      id,
		PlayerMap: string(pm),
	}
	var ret bool
	err := rpcCli.CallWithTimeout("PlayerDataSvc.UpdatePlayerElement", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Warnf("UpdatePlayerElement error:%v", err)
		return err
	}
	return nil
}

type GetPlayerBaseInfoArgs struct {
	Plt  string
	SnId int32
}

func GetPlayerBaseInfo(plt string, snid int32) *PlayerBaseInfo {
	if rpcCli == nil {
		return nil
	}
	args := &GetPlayerBaseInfoArgs{
		Plt:  plt,
		SnId: snid,
	}
	var ret PlayerBaseInfo
	err := rpcCli.CallWithTimeout("PlayerDataSvc.GetPlayerBaseInfo", args, &ret, time.Second*30)
	if err != nil {
		return nil
	}
	return &ret
}

type BlackWhiteLevel struct {
	SnId, WBLevel, WbMaxNum                    int32
	WbCoinTotalIn, WbCoinTotalOut, WbCoinLimit int64
	Platform                                   string
	T                                          time.Time
}

func SetBlackWhiteLevel(snid, wbLevel, wbMaxNum int32, wbCoinTotalIn, wbCoinTotalOut, wbCoinLimit int64, platform string, t time.Time) error {
	if rpcCli == nil {
		return nil
	}
	var args = &BlackWhiteLevel{
		SnId:           snid,
		WBLevel:        wbLevel,
		WbMaxNum:       wbMaxNum,
		WbCoinTotalIn:  wbCoinTotalIn,
		WbCoinTotalOut: wbCoinTotalOut,
		WbCoinLimit:    wbCoinLimit,
		Platform:       platform,
		T:              t,
	}
	var ret bool
	err := rpcCli.CallWithTimeout("PlayerDataSvc.SetBlackWhiteLevel", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Error("SetBlackWhiteLevel error ", err)
		return err
	}
	return nil
}

func SetBlackWhiteLevelUnReset(snid, wbLevel, wbMaxNum int32, wbCoinLimit int64, platform string, t time.Time) error {
	if rpcCli == nil {
		return nil
	}
	var args = &BlackWhiteLevel{
		SnId:        snid,
		WBLevel:     wbLevel,
		WbMaxNum:    wbMaxNum,
		WbCoinLimit: wbCoinLimit,
		Platform:    platform,
		T:           t,
	}
	var ret bool
	err := rpcCli.CallWithTimeout("PlayerDataSvc.SetBlackWhiteLevelUnReset", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Info("SetBlackWhiteLevelUnReset error ", err)
		return err
	}
	return nil
}
func UpdateAllPlayerPackageTag(packageTag, platform, channel, promoter string, promoterTree, tagkey int32) error {
	if rpcCli == nil {
		return nil
	}
	args := &UpdatePackageId{
		PackageTag:   packageTag,
		PlatformStr:  platform,
		ChannelStr:   channel,
		PromoterStr:  promoter,
		PromoterTree: promoterTree,
	}
	var ret bool
	err := rpcCli.CallWithTimeout("PlayerDataSvc.UpdateAllPlayerPackageTag", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Info("UpdateAllPlayerPackageTag error ", err)
		return err
	}
	return nil
}

type SetPlayerAtt struct {
	SnId     int32
	VipLevel int32
	Platform string
	GmLevel  int32
}

func SetVipLevel(snid, vipLevel int32, platform string) error {
	if rpcCli == nil {
		return nil
	}
	args := &SetPlayerAtt{
		SnId:     snid,
		VipLevel: vipLevel,
		Platform: platform,
	}
	var ret bool
	err := rpcCli.CallWithTimeout("PlayerDataSvc.SetVipLevel", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Info("SetVipLevel error ", err)
		return err
	}
	return nil
}

func SetGMLevel(snid, gmLevel int32, platform string) error {
	if rpcCli == nil {
		return nil
	}
	args := &SetPlayerAtt{
		SnId:     snid,
		GmLevel:  gmLevel,
		Platform: platform,
	}
	var ret bool
	err := rpcCli.CallWithTimeout("PlayerDataSvc.SetGMLevel", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Info("SetGMLevel error ", err)
		return err
	}
	return nil
}

type GetSameParamPlayerArgs struct {
	Plt   string
	Param string
}

func GetSameIpPlayer(plt, ip string) []int32 {
	if rpcCli == nil {
		return nil
	}
	args := &GetSameParamPlayerArgs{
		Plt:   plt,
		Param: ip,
	}
	var ret []int32
	err := rpcCli.CallWithTimeout("PlayerDataSvc.GetSameIpPlayer", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Info("GetSameIpPlayer error ", err)
		return nil
	}
	return ret
}

func GetSameBankNamePlayer(plt, name string) []int32 {
	if rpcCli == nil {
		return nil
	}
	args := &GetSameParamPlayerArgs{
		Plt:   plt,
		Param: name,
	}
	var ret []int32
	err := rpcCli.CallWithTimeout("PlayerDataSvc.GetSameBankNamePlayer", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Info("GetSameIpPlayer error ", err)
		return nil
	}
	return ret
}
func GetSameBankCardPlayer(plt, card string) []int32 {
	if rpcCli == nil {
		return nil
	}
	args := &GetSameParamPlayerArgs{
		Plt:   plt,
		Param: card,
	}
	var ret []int32
	err := rpcCli.CallWithTimeout("PlayerDataSvc.GetSameBankCardPlayer", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Info("GetSameBankCardPlayer error ", err)
		return nil
	}
	return ret
}

func GetRobotPlayers(limit int) []*PlayerData {
	if rpcCli == nil {
		return nil
	}
	var ret []*PlayerData
	err := rpcCli.CallWithTimeout("PlayerDataSvc.GetRobotPlayers", limit, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Info("GetSameBankCardPlayer error ", err)
		return nil
	}
	return ret
}

type BackupPlayerRet struct {
	Err       error
	IsSuccess bool
	Pd        *PlayerData
}

/*
 * 保存玩家的删除备份全部信息
 */
func SaveDelBackupPlayerData(pd *PlayerData) bool {
	if rpcCli == nil {
		logger.Logger.Error("model.SaveDelBackupPlayerData rpcCli is nil")
		return false
	}
	if pd != nil {
		var ret bool
		err := rpcCli.CallWithTimeout("PlayerDelBackupDataSvc.SaveDelBackupPlayerData", pd, &ret, time.Second*30)
		if err != nil {
			logger.Logger.Error("model.SaveDelBackupPlayerData is error", err)
			return false
		}
		return ret
	}
	return false
}

type LogicInfoArg struct {
	SnIds      []int32
	LogicLevel int32
	Platform   string
}

func SetLogicLevel(snids []int32, logicLevel int32, platform string) error {
	if rpcCli == nil {
		return nil
	}
	args := &LogicInfoArg{
		SnIds:      snids,
		LogicLevel: logicLevel,
		Platform:   platform,
	}
	var ret bool
	err := rpcCli.CallWithTimeout("PlayerDataSvc.SetLogicLevel", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Error("model.SetLogicLevel error ", err)
		return err
	}
	return nil
}

func ClrLogicLevel(snids []int32, logicLevel int32, platform string) error {
	if rpcCli == nil {
		return nil
	}
	args := &LogicInfoArg{
		SnIds:      snids,
		LogicLevel: logicLevel,
		Platform:   platform,
	}
	var ret bool
	err := rpcCli.CallWithTimeout("PlayerDataSvc.ClrLogicLevel", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Error("model.ClrLogicLevel error ", err)
		return err
	}
	return nil
}

// 所有游戏都要加上当天统计数据
func (this *PlayerData) GetDaliyGameData(id int) (*PlayerGameStatics, *PlayerGameStatics) {
	gameId := strconv.Itoa(id)
	if this.TodayGameData == nil {
		this.TodayGameData = NewPlayerGameCtrlData()
	}
	if this.TodayGameData.CtrlData == nil {
		this.TodayGameData.CtrlData = make(map[string]*PlayerGameStatics)
	}
	if _, ok := this.TodayGameData.CtrlData[gameId]; !ok {
		this.TodayGameData.CtrlData[gameId] = &PlayerGameStatics{}
	}
	if this.YesterdayGameData == nil {
		this.YesterdayGameData = NewPlayerGameCtrlData()
	}
	if this.YesterdayGameData.CtrlData == nil {
		this.YesterdayGameData.CtrlData = make(map[string]*PlayerGameStatics)
	}
	if _, ok := this.YesterdayGameData.CtrlData[gameId]; !ok {
		this.YesterdayGameData.CtrlData[gameId] = &PlayerGameStatics{}
	}
	return this.TodayGameData.CtrlData[gameId], this.YesterdayGameData.CtrlData[gameId]
}

func (this *PlayerData) GetTodayGameData(gameFreeid int32) *PlayerGameStatics {
	gameFreeStr := strconv.Itoa(int(gameFreeid))
	if this.TodayGameData == nil {
		this.TodayGameData = NewPlayerGameCtrlData()
	}
	if this.TodayGameData.CtrlData == nil {
		this.TodayGameData.CtrlData = make(map[string]*PlayerGameStatics)
	}
	if _, ok := this.TodayGameData.CtrlData[gameFreeStr]; !ok {
		this.TodayGameData.CtrlData[gameFreeStr] = &PlayerGameStatics{}
	}
	return this.TodayGameData.CtrlData[gameFreeStr]
}

func (this *PlayerData) GetYestosdayGameData(gameFreeid string) *PlayerGameStatics {

	if this.YesterdayGameData == nil {
		this.YesterdayGameData = &PlayerGameCtrlData{}
	}
	if this.YesterdayGameData.CtrlData == nil {
		this.YesterdayGameData.CtrlData = make(map[string]*PlayerGameStatics)
	}
	if _, ok := this.YesterdayGameData.CtrlData[gameFreeid]; !ok {
		this.YesterdayGameData.CtrlData[gameFreeid] = &PlayerGameStatics{}
	}
	return this.YesterdayGameData.CtrlData[gameFreeid]
}

func (this *PlayerData) HasAutoTag(tag int32) bool {
	return common.InSliceInt32(this.AutomaticTags, tag)
}

func (this *PlayerData) MarkAutoTag(tag int32) bool {
	if this.HasAutoTag(tag) {
		return false
	}

	this.AutomaticTags = append(this.AutomaticTags, tag)
	return true
}

func (this *PlayerData) GetGameFreeIdData(id string) *PlayerGameInfo {
	data, ok := this.GDatas[id]
	if !ok {
		data = new(PlayerGameInfo)
		this.GDatas[id] = data
	}
	return data
}
