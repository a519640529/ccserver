package main

import (
	"games.yol.com/win88/common"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/gamehall"
	webapi_proto "games.yol.com/win88/protocol/webapi"
	"github.com/idealeak/goserver/core/logger"
	"strconv"
)

const (
	Bind_BackCard int = 1 << iota //银行卡
	Bind_Alipay                   //可以绑定阿里
	Bind_WeiXin                   //微信
)

const (
	VerifyCodeType_Nil    = iota //不使用
	VerifyCodeType_RndStr        //随机串
	VerifyCodeType_Slider        //滑块
)

const (
	ExchangeFlag_Tax   = 1 << iota //兑换税收
	ExchangeFlag_Flow              //兑换流水
	ExchangeFlag_Force             //强制兑换流水是否开启
	ExchangeFlag_UpAcc             //账号注册升级是否开启

)

// 排行榜开关
type RankSwitch struct {
	Asset    int32 // 财富榜
	Recharge int32 // 充值榜
	Exchange int32 // 兑换榜
	Profit   int32 // 盈利榜
	Flow     int32 // 流水榜
}

// 俱乐部配置
type ClubConfig struct {
	CreationCoin            int64   //创建俱乐部金额
	IncreaseCoin            int64   //升级俱乐部金额
	ClubInitPlayerNum       int32   //俱乐部初始人数
	IncreasePlayerNum       int32   //升级人数增加
	IsOpenClub              bool    //是否开放俱乐部
	CreateClubCheckByManual bool    //创建俱乐部人工审核，true=手动
	EditClubNoticeByManual  bool    //修改公告人工审核，true=手动
	CreateRoomAmount        int64   //创建房间金额（分/局）
	GiveCoinRate            []int64 //会长充值额外赠送比例
}
type Platform struct {
	Id                       int32  //平台ID
	IdStr                    string //字符id
	Name                     string //平台名称
	Isolated                 bool   //是否孤立(别的平台看不到)
	Disable                  bool
	Halls                    map[int32]*PlatformGameHall      //厅
	GamePlayerNum            map[int32]*PlatformGamePlayerNum //游戏人数
	dirty                    bool                             //
	ServiceUrl               string                           //客服地址
	BindOption               int32                            //绑定选项
	ServiceFlag              bool                             //客服标记 是否支持浏览器跳转  false否 true是
	UpgradeAccountGiveCoin   int32                            //升级账号奖励金币
	NewAccountGiveCoin       int32                            //新账号奖励金币
	PerBankNoLimitAccount    int32                            //同一银行卡号绑定用户数量限制
	ExchangeMin              int32                            //最低兑换金额
	ExchangeLimit            int32                            //兑换后身上保留最低余额
	ExchangeTax              int32                            //兑换税收（万分比）
	ExchangeFlow             int32                            //兑换流水比例
	ExchangeForceTax         int32                            //强制兑换税收
	ExchangeGiveFlow         int32                            //赠送兑换流水
	ExchangeFlag             int32                            //兑换标记 二进制 第一位:兑换税收 第二位:流水比例
	ExchangeVer              int32                            //兑换版本号
	ExchangeMultiple         int32                            //兑换基数（只能兑换此数的整数倍）
	VipRange                 []int32                          //VIP充值区间
	OtherParams              string                           //其他参数json串
	SpreadConfig             int32                            //0:等级返点 1:保底返佣
	RankSwitch               RankSwitch                       //排行榜开关
	ClubConfig               *ClubConfig                      //俱乐部配置
	VerifyCodeType           int32                            //注册账号使用验证码方式 0:短信验证码 1:随机字符串 2:滑块验证码 3:不使用
	RegisterVerifyCodeSwitch bool                             // 关闭注册验证码
	ThirdGameMerchant        map[int32]int32                  //三方游戏平台状态
	CustomType               int32                            //客服类型 0:live800 1:美洽 2:cc
	NeedDeviceInfo           bool                             //需要获取设备信息
	NeedSameName             bool                             //绑定的银行卡和支付宝用户名字需要相同
	ExchangeBankMax          int32                            //银行卡最大兑换金额  0不限制
	ExchangeAlipayMax        int32                            //支付宝最大兑换金额 0不限制
	DgHboConfig              int32                            //dg hbo配置，默认0，dg 1 hbo 2
	PerBankNoLimitName       int32                            //银行卡和支付宝 相同名字最大数量
	IsCanUserBindPromoter    bool                             //是否允许用户手动绑定推广员
	UserBindPromoterPrize    int32                            //手动绑定奖励
	SpreadWinLose            bool                             //是否打开客损开关
	PltGameCfg               *PlatformGameConfig              //平台游戏配置
	MerchantKey              string                           //商户秘钥
}

type PlatformGameConfig struct {
	games map[int32]*webapi_proto.GameFree   //以gamefreeid为key
	cache map[int32][]*webapi_proto.GameFree //以gameid为key
}

func (cfg *PlatformGameConfig) RecreateCache() {
	if cfg.cache == nil {
		cfg.cache = make(map[int32][]*webapi_proto.GameFree)
	}
	for _, val := range cfg.games {
		cfg.cache[val.DbGameFree.GetGameId()] = append(cfg.cache[val.DbGameFree.GetGameId()], val)
	}
}

func (cfg *PlatformGameConfig) GetGameCfg(gamefreeId int32) *webapi_proto.GameFree {
	if cfg.games != nil {
		if c, exist := cfg.games[gamefreeId]; exist {
			if c.GroupId == 0 {
				return c
			} else {
				groupCfg := PlatformGameGroupMgrSington.GetGameGroup(c.GroupId)
				temp := &webapi_proto.GameFree{
					GroupId:    groupCfg.GetId(),
					Status:     c.Status,
					DbGameFree: groupCfg.GetDbGameFree(),
				}
				return temp
			}
		} else {
			logger.Logger.Errorf("PlatformGameConfig GetGameCfg Can't Find GameCfg[%v]", gamefreeId)
		}
	}
	return nil
}

func CompareGameFreeConfigChged(oldCfg, newCfg *webapi_proto.GameFree) bool {
	if oldCfg.Status != newCfg.Status ||
		oldCfg.DbGameFree.GetBot() != newCfg.DbGameFree.GetBot() ||
		oldCfg.DbGameFree.GetBaseScore() != newCfg.DbGameFree.GetBaseScore() ||
		oldCfg.DbGameFree.GetLimitCoin() != newCfg.DbGameFree.GetLimitCoin() ||
		oldCfg.DbGameFree.GetMaxCoinLimit() != newCfg.DbGameFree.GetMaxCoinLimit() ||
		oldCfg.DbGameFree.GetSameIpLimit() != newCfg.DbGameFree.GetSameIpLimit() ||
		oldCfg.DbGameFree.GetSamePlaceLimit() != newCfg.DbGameFree.GetSamePlaceLimit() ||
		oldCfg.DbGameFree.GetLowerThanKick() != newCfg.DbGameFree.GetLowerThanKick() ||
		oldCfg.DbGameFree.GetBanker() != newCfg.DbGameFree.GetBanker() ||
		oldCfg.DbGameFree.GetMaxChip() != newCfg.DbGameFree.GetMaxChip() ||
		oldCfg.DbGameFree.GetBetLimit() != newCfg.DbGameFree.GetBetLimit() ||
		oldCfg.DbGameFree.GetLottery() != newCfg.DbGameFree.GetLottery() ||
		oldCfg.DbGameFree.GetLotteryConfig() != newCfg.DbGameFree.GetLotteryConfig() ||
		oldCfg.DbGameFree.GetTaxRate() != newCfg.DbGameFree.GetTaxRate() ||
		!common.SliceInt32Equal(oldCfg.DbGameFree.GetOtherIntParams(), newCfg.DbGameFree.GetOtherIntParams()) ||
		!common.SliceInt32Equal(oldCfg.DbGameFree.GetRobotTakeCoin(), newCfg.DbGameFree.GetRobotTakeCoin()) ||
		!common.SliceInt32Equal(oldCfg.DbGameFree.GetRobotLimitCoin(), newCfg.DbGameFree.GetRobotLimitCoin()) ||
		!common.SliceInt32Equal(oldCfg.DbGameFree.GetRobotNumRng(), newCfg.DbGameFree.GetRobotNumRng()) ||
		(len(newCfg.DbGameFree.GetMaxBetCoin()) > 0 &&
			!common.Int32SliceEqual(oldCfg.DbGameFree.GetMaxBetCoin(), newCfg.DbGameFree.GetMaxBetCoin())) ||
		oldCfg.DbGameFree.GetMatchMode() != newCfg.DbGameFree.GetMatchMode() ||
		oldCfg.DbGameFree.GetCreateRoomNum() != newCfg.DbGameFree.GetCreateRoomNum() ||
		oldCfg.DbGameFree.GetMatchTrueMan() != newCfg.DbGameFree.GetMatchTrueMan() {
		return false
	}
	return true
}

func NewPlatform(id int32, isolated bool) *Platform {
	p := &Platform{
		Id:            id,
		IdStr:         strconv.Itoa(int(id)),
		Isolated:      isolated,
		Halls:         make(map[int32]*PlatformGameHall),
		GamePlayerNum: make(map[int32]*PlatformGamePlayerNum),
		ClubConfig:    &ClubConfig{},
		PltGameCfg: &PlatformGameConfig{
			games: make(map[int32]*webapi_proto.GameFree),
			cache: make(map[int32][]*webapi_proto.GameFree),
		},
	}

	return p
}

// true 有流水标记 false 没有流水标记
func (p *Platform) IsFlowSwitch() bool {
	if (p.ExchangeFlag&ExchangeFlag_Flow) != 0 && p.ExchangeFlow > 0 {
		return true
	}
	return false
}

func (p *Platform) ChangeIsolated(isolated bool) bool {
	if p.Isolated == isolated {
		return false
	}
	p.Isolated = isolated
	if isolated { //开放平台转换成私有平台
		for _, hall := range p.Halls {
			hall.ConvertToIsolated()
		}
	} else { //私有平台转换为开放平台
		for _, hall := range p.Halls {
			hall.OpenSceneToPublic()
		}
	}
	return true
}

func (p *Platform) ChangeDisabled(disable bool) bool {
	if p.Disable == disable {
		return false
	}
	p.Disable = disable
	if disable { //关闭平台,踢掉平台上所有的人
		PlayerMgrSington.KickoutByPlatform(p.IdStr)
	}
	return true
}

func (p *Platform) PlayerEnter(player *Player, hallId int32) {
	if h, exist := p.Halls[hallId]; exist {
		h.PlayerEnter(player)
	} else {
		h := NewPlatformGameHall(p, hallId)
		if h != nil {
			p.Halls[hallId] = h
			h.PlayerEnter(player)
		}
	}
}

func (p *Platform) PlayerLeave(player *Player) {
	if h, exist := p.Halls[player.hallId]; exist {
		h.PlayerLeave(player)
	}
}

func (p *Platform) OnDestroyScene(scene *Scene) {
	if h, exist := p.Halls[scene.hallId]; exist {
		h.OnDestroyScene(scene)
	}
}

func (p *Platform) OnPlayerEnterScene(scene *Scene, player *Player) {
	if h, exist := p.Halls[scene.hallId]; exist {
		h.OnPlayerEnterScene(scene, player)
		gameid := h.dbGameFree.GetGameId()
		if _, exist := p.GamePlayerNum[gameid]; !exist {
			p.GamePlayerNum[gameid] = &PlatformGamePlayerNum{
				Nums:  make(map[int32]int),
				Dirty: true,
			}
		}
		if nums, exist := p.GamePlayerNum[gameid]; exist {
			sceneType := h.dbGameFree.GetSceneType()
			nums.Nums[sceneType] = nums.Nums[sceneType] + 1
			nums.Dirty = true
			p.dirty = true
		}
	}
}

func (p *Platform) OnPlayerLeaveScene(scene *Scene, player *Player) {
	if h, exist := p.Halls[scene.hallId]; exist {
		h.OnPlayerLeaveScene(scene, player)
		gameid := h.dbGameFree.GetGameId()
		if nums, exist := p.GamePlayerNum[gameid]; exist {
			sceneType := h.dbGameFree.GetSceneType()
			if n, exist := nums.Nums[sceneType]; exist && n > 0 {
				nums.Nums[sceneType] = n - 1
				nums.Dirty = true
				p.dirty = true
			}
		}
	}
}

func (p *Platform) PlayerLogin(player *Player) {
	//发送平台配置相关的数据
	player.SendPlatformCanUsePromoterBind()
}

func (p *Platform) PlayerLogout(player *Player) {
	if player.hallId != 0 {
		if h, exist := p.Halls[player.hallId]; exist {
			h.PlayerLeave(player)
		}
	}
}

func (p *Platform) SendPlayerNum(gameid int32, pp *Player) {
	if nums, exist := p.GamePlayerNum[gameid]; exist {
		pack := &gamehall.HallPlayerNum{}
		for k, v := range nums.Nums {
			pack.HallData = append(pack.HallData, &gamehall.HallInfo{
				SceneType: proto.Int32(k),
				PlayerNum: proto.Int(v),
			})
		}
		proto.SetDefaults(pack)
		pp.SendToClient(int(gamehall.GameHallPacketID_PACKET_SC_HALLPLAYERNUM), pack)
	}
}

func (p *Platform) BroadcastPlayerNum() {
	if p.dirty {
		p.dirty = false
		for gameid, nums := range p.GamePlayerNum {
			if nums.Dirty {
				nums.Dirty = false
				pack := &gamehall.HallPlayerNum{}
				for k, v := range nums.Nums {
					pack.HallData = append(pack.HallData, &gamehall.HallInfo{
						SceneType: proto.Int32(k),
						PlayerNum: proto.Int(v),
					})
				}
				proto.SetDefaults(pack)
				for _, h := range p.Halls {
					if h.dbGameFree.GetGameId() == gameid {
						h.Broadcast(int(gamehall.GameHallPacketID_PACKET_SC_HALLPLAYERNUM), pack, 0)
					}
				}
			}
		}
	}
}
func (p *Platform) IsMarkFlag(flag int32) bool {
	if (p.BindOption & flag) != 0 {
		return true
	}
	return false
}

func (p *Platform) OnDayTimer() {
	if p.IdStr == common.Platform_Rob {
		return
	}

	logger.Logger.Tracef("Platform.OnDayTimer platform %v", p.Name)
	//ActLuckyTurntableMgrSington.LuckyTurntableReset(p.Name)
	//ActYebMgrSington.SwitchRate(p.Name)
}

func (p *Platform) OnHourTimer() {
	if p.IdStr == common.Platform_Rob {
		return
	}

	//ActRankListMgrSington.RefreshRank(p.Name)
}
