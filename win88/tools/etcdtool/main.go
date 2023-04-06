package main

import (
	"encoding/json"
	"fmt"
	_ "games.yol.com/win88"
	"games.yol.com/win88/common"
	"games.yol.com/win88/etcd"
	"games.yol.com/win88/protocol"
	"games.yol.com/win88/srvdata"
	"games.yol.com/win88/webapi"
	"github.com/idealeak/goserver/core"
	_ "github.com/idealeak/goserver/core/i18n"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/module"
	"go.etcd.io/etcd/clientv3"
	"net"
	"strconv"
	"time"
)

//获取平台数据 platform_list

// 排行榜开关
type RankSwitch struct {
	Asset    int32 // 财富榜
	Recharge int32 // 充值榜
	Exchange int32 // 兑换榜
	Profit   int32 // 盈利榜
}

//俱乐部配置
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

type PlatformInfoApi struct {
	PlatformName           string
	Isolated               bool
	Disabled               bool
	ConfigId               int32
	CustomService          string
	BindOption             int32
	ServiceFlag            int32
	UpgradeAccountGiveCoin int32   //升级账号奖励金币
	NewAccountGiveCoin     int32   //新账号奖励金币
	PerBankNoLimitAccount  int32   //同一银行卡号绑定用户数量限制
	ExchangeMin            int32   //最低兑换金额
	ExchangeLimit          int32   //兑换限制
	ExchangeTax            int32   //兑换税收（万分比）
	ExchangeForceTax       int32   //强制兑换税收
	ExchangeFlow           int32   //兑换流水比例
	ExchangeGiveFlow       int32   //赠送兑换流水比例
	ExchangeFlag           int32   //兑换标记 二进制 第一位:兑换税收 第二位:流水比例
	ExchangeVer            int32   //兑换版本
	ExchangeMultiple       int32   //兑换基数
	VipRange               []int32 //VIP充值区间
	OtherParams            string  //其他参数json串
	SpreadConfig           int32
	Leaderboard            RankSwitch
	ClubConfig             *ClubConfig     //俱乐部配置
	VerifyCodeType         int32           //验证码方式
	ThirdGameMerchant      map[int32]int32 //三方游戏平台状态
	CustomType             int32
	NeedDeviceInfo         bool
	NeedSameName           bool  //绑定的银行卡和支付宝用户名字需要相同
	ExchangeBankMax        int32 //银行卡最大兑换金额  0不限制
	ExchangeAlipayMax      int32 //支付宝最大兑换金额 0不限制
	DgHboConfig            int32 //dg hbo配置，默认0，dg 1 hbo 2
	PerBankNoLimitName     int32 //银行卡名字数量限制
	IsCanUserBindPromoter  bool  //是否允许用户手动绑定推广员
	UserBindPromoterPrize  int32 //手动绑定奖励
}

type PackageListApi struct {
	Tag            string //android包名或者ios标记
	Platform       int32  //所属平台
	ChannelId      int32  //渠道ID
	PromoterId     int32  //推广员ID
	PromoterTree   int32  //无级推广树
	SpreadTag      int32  //全民包标识 0:普通包 1:全民包
	OpenInstallTag int32  //是否是openinstall包 0:不是 1:是
	Status         int32  //状态
	AppStore       int32  //是否是苹果商店包 0:不是 1:是
	ExchangeFlag   int32  //兑换标记 0 关闭包返利 1打开包返利 受平台配置影响
	ExchangeFlow   int32  //兑换比例
	IsForceBind    int32
}

type PlatformCoinfig struct {
	Id        int32
	Params    []PlatConDataDetail
	Parent_id int32
	Platform  int32
}

//游戏平台配置
type PlatConDataDetail struct {
	LogicId    int32                 ////对应DB_GameFree.xlsx中的id
	Param      string                //参数
	State      int32                 //开关  0:关  1：开
	GroupId    int32                 //组id
	DBGameFree *protocol.DB_GameFree //游戏配置
}

//游戏平台状态
type PlatConfState struct {
	LogicId    int32                 //对应GameGlobalState中的LogicId
	Param      string                //参数
	State      int32                 //开关  0:关  1：开
	GroupId    int32                 //组id
	DBGameFree *protocol.DB_GameFree //配置参数
}

var PlatformList = make(map[string]*PlatformInfoApi)
var PackageList = make(map[string]*PackageListApi)
var PlatConList = make(map[int32]*PlatformCoinfig)
var etcdCli = &etcd.EtcdClient{}
var waitCnt int
var isClearEtcd = "n" //是否清理etcd数据，默认是不清理
var isPutToEtcd = "n" //是否拉取数据到etcd，默认是不拉取
var isWatchEtcd = "n" //是否监控etcd,默认是不监控
func main() {
	logger.Trace(`==================== Tips:  Submit Your Choice  =======================`)

	logger.Tracef(`Do you want delete etcd only key=%v data (y/n) ？`, etcd.ETCDKEY_ROOT_PREFIX)
	fmt.Scan(&isClearEtcd)

	logger.Tracef(`Do you want put data to etcd by WebAPI (y/n) ？`)
	fmt.Scan(&isPutToEtcd)

	logger.Tracef(`Do you want watch etcd (y/n) ？`)
	fmt.Scan(&isWatchEtcd)

	var err error
	if isClearEtcd != "y" && isPutToEtcd != "y" && isWatchEtcd != "y" {
		logger.Errorf(`you don't have select any action, bye! `)
		goto END
	}
	defer core.ClosePackages()
	core.LoadPackages("config.json")
	logger.Trace("etcdurl=", common.CustomConfig.GetStrings("etcdurl"))
	logger.Trace("etcduser=", common.CustomConfig.GetString("etcduser"))
	logger.Trace("etcdpwd=", common.CustomConfig.GetString("etcdpwd"))
	logger.Trace("正在连接ETCD服务...")
	err = etcdCli.Open(common.CustomConfig.GetStrings("etcdurl"), common.CustomConfig.GetString("etcduser"), common.CustomConfig.GetString("etcdpwd"), time.Minute)
	if err != nil {
		logger.Error("etcd connect fail because : ", err)
		return
	}
	logger.Trace("连接ETCD服务成功！")

	if isClearEtcd == "y" {
		logger.Trace("delete etcd data...")
		rep, err := etcdCli.DelValueWithPrefix(etcd.ETCDKEY_ROOT_PREFIX)
		if err != nil {
			logger.Logger.Errorf("delete etcd date err:%v", err)
		} else {
			logger.Tracef("delete %v item in key=%v", rep.Deleted, etcd.ETCDKEY_ROOT_PREFIX)
		}
	}

	if isPutToEtcd == "y" {
		logger.Trace("是否只获取公共黑名单数据=", common.CustomConfig.GetBool("UseBlacklistBindPlayerinfo"))
		//PutTestPlatform2Etcd()
		//return
		LoadActGoldCome()
		PutActGoldCome2Etcd()
		return
		LoadRebateData()
		PutRebateTaskEtcd()

		//获取平台列表
		GetPlatformList()
		PutPlatform2Etcd()

		//包
		LoadPlatformPackage()
		PutPlatformPackage2Etcd()

		//
		LoadPlatformConfig()
		PutPlatformConfig2Etcd()

		//
		LoadGameGroup()
		PutPlatformGroup2Etcd()

		//
		LoadBullet()
		PutBullet2Etcd()

		//
		LoadCustomer()
		PutCustomer2Etcd()

		LoadBlackList()
		PutBlackList2Etcd()

		LoadActSign()
		PutActSign2Etcd()

		//
		LoadActGoldTask()
		PutActGoldTask2Etcd()

		//

		LoadActOnlineReward()
		PutActOnlineReward2Etcd()

		LoadLuckyTurntableConfig()
		PutLuckyTurntableConfig2Etcd()

		LoadYebConfig()
		PutYeb2ConfigEtcd()

		LoadPromoterData()
		PutPromoterEtcd()

		LoadActVipData()
		PutActVipEtcd()

		LoadActWxShareData()
		PutActWxShareEtcd()

		LoadActData()
		PutAcEtcd()

		LoadRandomPrizeData()
		PutRandomPrizeEtcd()

		LoadActFPayData()
		PutAcFPayEtcd()

	}

	if isWatchEtcd == "y" {
		logger.Trace("Tip:watching etcd ,you can get/put keys....")
		go func() {
			defer func() {
				if err := recover(); err != nil {
					logger.Logger.Errorf("etcd watch WithPrefix(%v) panic:%v", etcd.ETCDKEY_ROOT_PREFIX, err)
				}
				logger.Logger.Warnf("etcd watch WithPrefix(%v) quit!!!", etcd.ETCDKEY_ROOT_PREFIX)
			}()
			var times int64
			for {
				times++
				logger.Logger.Warnf("etcd watch WithPrefix(%v) start[%v]!!!", etcd.ETCDKEY_ROOT_PREFIX, times)
				rch := etcdCli.WatchWithPrefix(etcd.ETCDKEY_ROOT_PREFIX)
				for wresp := range rch {
					if wresp.Canceled {
						logger.Logger.Warnf("etcd watch WithPrefix(%v) be closed, reason:%v", etcd.ETCDKEY_ROOT_PREFIX, wresp.Err())
						continue
					}
					for _, ev := range wresp.Events {
						switch ev.Type {
						case clientv3.EventTypeDelete:
							logger.Logger.Infof("etcd desc WithPrefix(%v) delete data:%v", string(ev.Kv.Key), string(ev.Kv.Value))
						case clientv3.EventTypePut:
							logger.Logger.Infof("etcd desc WithPrefix(%v) put data:%v", string(ev.Kv.Key), string(ev.Kv.Value))
						}
					}
				}
			}
		}()
	}
END:
	//启动业务模块
	waiter := module.Start()
	waiter.Wait("main()")
}

//平台列表
func GetPlatformList() {
	type ApiResult struct {
		Tag int
		Msg []PlatformInfoApi
	}
	logger.Logger.Trace("start API_GetPlatformList")
	platformBuff, err := webapi.API_GetPlatformData(common.GetAppId())
	if err == nil {
		//logger.Logger.Trace("API_GetPlatformData:", string(platformBuff))
		ar := ApiResult{}
		err = json.Unmarshal(platformBuff, &ar)
		if err == nil && ar.Tag == 0 {
			for i := 0; i < len(ar.Msg); i++ {
				PlatformList[ar.Msg[i].PlatformName] = &ar.Msg[i]
			}
		} else {
			logger.Logger.Error("Unmarshal platform data error:", err, string(platformBuff))
		}
	} else {
		logger.Logger.Error("Get platfrom data error:", err)
	}
}

func PutTestPlatform2Etcd() {

	for i := 1; i < 10000; i++ {
		s := "{\"PlatformName\":\"9\",\"Isolated\":true,\"Disabled\":true,\"ConfigId\":9,\"CustomService\":\"\",\"BindOption\":1,\"ServiceFlag\":1,\"UpgradeAccountGiveCoin\":300,\"NewAccountGiveCoin\":0,\"PerBankNoLimitAccount\":1,\"ExchangeMin\":10000,\"ExchangeLimit\":300,\"ExchangeTax\":100,\"ExchangeForceTax\":0,\"ExchangeMultiple\":0,\"ExchangeFlow\":0,\"ExchangeGiveFlow\":0,\"ExchangeFlag\":1,\"ExchangeVer\":0,\"VipRange\":[10000,50000,100000,300000,500000,1000000],\"OtherParams\":\"\",\"SpreadConfig\":1,\"Leaderboard\":{\"Asset\":0,\"Recharge\":0,\"Exchange\":0,\"Profit\":0},\"ClubConfig\":{\"CreationCoin\":10000,\"IncreaseCoin\":100000,\"ClubInitPlayerNum\":50,\"IncreasePlayerNum\":50,\"IsOpenClub\":false,\"CreateClubCheckByManual\":false,\"EditClubNoticeByManual\":false,\"CreateRoomAmount\":1,\"GiveCoinRate\":[1,2,3,4,5,6,7,8,9,10]},\"VerifyCodeType\":0,\"ThirdGameMerchant\":{\"38\":1,\"39\":1,\"41\":0,\"43\":1,\"46\":0,\"47\":0,\"48\":0},\"CustomType\":0,\"NeedDeviceInfo\":false,\"NeedSameName\":false,\"ExchangeBankMax\":0,\"ExchangeAlipayMax\":0,\"DgHboConfig\":0}"
		data, err := json.Marshal(s)
		if err == nil {
			_, err := etcdCli.PutValue(fmt.Sprintf("%v%v", etcd.ETCDKEY_PLATFORM_PREFIX, "7"), string(s))
			if err != nil {
				logger.Logger.Tracef("PutPlatform2Etcd err:%v data:%v", err, string(s))
			}
		} else {
			logger.Logger.Errorf("PutPlatform2Etcd err:%v data:%v", err, string(data))
		}
	}

}

func PutPlatform2Etcd() {
	etcdCli.DelValueWithPrefix(etcd.ETCDKEY_PLATFORM_PREFIX)

	logger.Logger.Trace("ETCD_PutPlatform2Etcd")
	for name, p := range PlatformList {
		data, err := json.Marshal(p)
		if err == nil {
			_, err := etcdCli.PutValue(fmt.Sprintf("%v%v", etcd.ETCDKEY_PLATFORM_PREFIX, name), string(data))
			if err != nil {
				logger.Logger.Tracef("PutPlatform2Etcd err:%v data:%v", err, string(data))
			}
		} else {
			logger.Logger.Errorf("PutPlatform2Etcd err:%v data:%v", err, string(data))
		}
	}
}

//包列表
func LoadPlatformPackage() {
	//获取包对应关系 package_list
	logger.Logger.Trace("start API_PackageList")
	packageBuff, err := webapi.API_PackageList(common.GetAppId())
	if err == nil {
		type ApiResult struct {
			Tag int
			Msg []PackageListApi
		}
		ar := ApiResult{}
		err = json.Unmarshal(packageBuff, &ar)
		if err == nil && ar.Tag == 0 {
			for i := 0; i < len(ar.Msg); i++ {
				if ar.Msg[i].Platform == 0 {
					ar.Msg[i].Platform = 1
				}
				PackageList[ar.Msg[i].Tag] = &ar.Msg[i]
				//logger.Logger.Trace("PlatformPackage data:", ar.Msg[i])
			}
		} else {
			logger.Logger.Error("Unmarshal package list data error:", err, string(packageBuff))
		}
	} else {
		logger.Logger.Error("Get package list data error:", err)
	}
}

func PutPlatformPackage2Etcd() {
	etcdCli.DelValueWithPrefix(etcd.ETCDKEY_PACKAGE_PREFIX)
	logger.Logger.Trace("ETCD_PutPlatformPackage2Etcd")
	//type ETCDPackageInfo struct {
	//	Tag          string //android包名或者ios标记
	//	Platform     int32  //所属平台
	//	Channel      int32  //渠道ID
	//	Promoter     int32  //推广员ID
	//	PromoterTree int32  //无级推广树
	//	SpreadTag    int32  //全民包标识 0:普通包 1:全民包
	//	Status       int32  //状态
	//}
	count := len(PackageList)
	cur := 0
	for name, p := range PackageList {
		//value := &ETCDPackageInfo{
		//	Tag:          p.Tag,
		//	Platform:     p.Platform,
		//	Channel:      p.ChannelId,
		//	Promoter:     p.PromoterId,
		//	PromoterTree: p.PromoterTree,
		//	SpreadTag:    p.SpreadTag,
		//	Status:       p.Status,
		//}
		data, err := json.Marshal(p)
		if err == nil {
			_, err := etcdCli.PutValue(fmt.Sprintf("%v%v", etcd.ETCDKEY_PACKAGE_PREFIX, name), string(data))
			if err != nil {
				logger.Logger.Tracef("PutPlatformPackage2Etcd err:%v data:%v", err, string(data))
			} else {
				//logger.Logger.Tracef("PutPlatformPackage2Etcd succes:%v data:%v", err, string(data))
			}
		} else {
			logger.Logger.Errorf("PutPlatformPackage2Etcd err:%v data:%v", err, string(data))
		}
		cur++
		fmt.Printf("\r")
		fmt.Printf("%d%%", int(float64(cur)/float64(count)*100))
	}
}

func LoadPlatformConfig() {
	//获取平台详细信息 game_config_list
	type PlatformConfigApi struct {
		Id        int32
		Params    []PlatConfState
		Parent_id int32
		Platform  int32
	}
	type PlatformConfigDataResult struct {
		Tag int
		Msg []PlatformConfigApi
	}
	//logger.Logger.Trace("LoadPlatformConfig")
	configBuff, err := webapi.API_GetPlatformConfigData(common.GetAppId())
	//logger.Trace(string(configBuff))
	if err == nil {
		pcdr := PlatformConfigDataResult{}
		err = json.Unmarshal(configBuff, &pcdr)
		if err == nil && pcdr.Tag == 0 {
			for _, config := range pcdr.Msg {
				pc := &PlatformCoinfig{
					Id:        config.Id,
					Parent_id: config.Parent_id,
					Platform:  config.Platform,
				}
				println(pc.Id)
				for _, config := range config.Params {
					dbGameFree := srvdata.PBDB_GameFreeMgr.GetData(config.LogicId)
					if dbGameFree == nil {
						logger.Logger.Error("Platform config data error logic id:", config.LogicId)
						continue
					}
					pcd := PlatConDataDetail{
						LogicId:    config.LogicId,
						Param:      config.Param,
						State:      config.State,
						GroupId:    config.GroupId,
						DBGameFree: config.DBGameFree,
					}
					if pcd.DBGameFree == nil { //数据容错
						pcd.DBGameFree = dbGameFree
					} else {

						CopyDBGameFreeField(dbGameFree, pcd.DBGameFree)
					}
					pc.Params = append(pc.Params, pcd)
				}
				//logger.Logger.Info("PlatformCoinfig data:", pc)
				PlatConList[config.Id] = pc
			}
		} else {
			logger.Logger.Error("Unmarshal platform config data error:", err, string(configBuff))
		}
	} else {
		logger.Logger.Error("Get platfrom config data error:", err)
	}
}

func PutPlatformConfig2Etcd() {
	etcdCli.DelValueWithPrefix(etcd.ETCDKEY_GAMECONFIG_PREFIX)
	logger.Logger.Trace("ETCD_PutPlatformConfig2Etcd")
	count := len(PlatConList)
	cur := 0
	for name, p := range PlatConList {
		data, err := json.Marshal(p)
		if err == nil {
			_, err := etcdCli.PutValue(fmt.Sprintf("%v%v", etcd.ETCDKEY_GAMECONFIG_PREFIX, name), string(data))
			if err != nil {
				logger.Logger.Tracef("PutPlatformConfig2Etcd err:%v data:%v", err, string(data))
			} else {
				//logger.Logger.Tracef("PutPlatformConfig2Etcd succes:%v data:%v", err, string(data))
			}
		} else {
			logger.Logger.Errorf("PutPlatformConfig2Etcd err:%v data:%v", err, string(data))
		}

		cur++
		fmt.Printf("\r")
		fmt.Printf("%d%%", int(float64(cur)/float64(count)*100))
	}
}

type PlatformGameGroup struct {
	GroupId    int32                 `json:"id"`
	LogicId    int32                 `json:"LogicId"`
	State      int32                 `json:"State"`
	DBGameFree *protocol.DB_GameFree `json:"DBGameFree"` //游戏配置
}

var GameGroups = make(map[int32]*PlatformGameGroup)

func LoadGameGroup() {
	//获取包对应关系 package_list
	logger.Logger.Trace("API_GetGameGroupData")
	packageBuff, err := webapi.API_GetGameGroupData(common.GetAppId())
	if err == nil {
		type ApiResult struct {
			Tag int
			Msg []*PlatformGameGroup
		}
		ar := ApiResult{}
		err = json.Unmarshal(packageBuff, &ar)
		if err == nil && ar.Tag == 0 {
			for _, value := range ar.Msg {
				dbGameFree := srvdata.PBDB_GameFreeMgr.GetData(value.LogicId)
				if dbGameFree == nil {
					continue
				}
				if value.DBGameFree == nil {
					value.DBGameFree = dbGameFree
				} else {
					CopyDBGameFreeField(dbGameFree, value.DBGameFree)
				}
				GameGroups[value.GroupId] = value
				//logger.Logger.Trace("PlatformGameGroup data:", value)
			}
		} else {
			logger.Logger.Error("Unmarshal PlatformGameGroup data error:", err, string(packageBuff))
		}
	} else {
		logger.Logger.Error("Get PlatformGameGroup data error:", err)
	}
}

func PutPlatformGroup2Etcd() {
	etcdCli.DelValueWithPrefix(etcd.ETCDKEY_GROUPCONFIG_PREFIX)
	logger.Logger.Trace("ETCD_PutPlatformGroup2Etcd")
	count := len(GameGroups)
	cur := 0
	for name, p := range GameGroups {
		data, err := json.Marshal(p)
		if err == nil {
			_, err := etcdCli.PutValue(fmt.Sprintf("%v%v", etcd.ETCDKEY_GROUPCONFIG_PREFIX, name), string(data))
			if err != nil {
				logger.Logger.Tracef("PutPlatformGroup2Etcd err:%v data:%v", err, string(data))
			}
		} else {
			logger.Logger.Errorf("PutPlatformGroup2Etcd err:%v data:%v", err, string(data))
		}

		cur++
		fmt.Printf("\r")
		fmt.Printf("%d%%", int(float64(cur)/float64(count)*100))
	}
}

type Bullet struct {
	Id            int32
	Sort          int32 //排序
	Platform      string
	NoticeTitle   string
	NoticeContent string
	UpdateTime    string
	State         int //0 关闭  1开启
} //声明与world中一样的结构体
type ApiBulletResult struct {
	Tag int
	Msg []Bullet
}

var BulletMsgList = make(map[int32]*Bullet)

func LoadBullet() {
	logger.Logger.Trace("LoadBullet")
	buff, err := webapi.API_GetBulletData(common.GetAppId())
	//logger.Logger.Warn("bulletin buff: ", string(buff))
	if err == nil {
		info := ApiBulletResult{}
		err = json.Unmarshal([]byte(buff), &info)
		if err == nil {
			for i := 0; i < len(info.Msg); i++ {
				BulletMsgList[info.Msg[i].Id] = &info.Msg[i]
			}
		} else {
			logger.Logger.Error("Unmarshal Bullet data error:", err, string(buff))
		}
	} else {
		logger.Logger.Error("Get Bullet data error:", err)
	}
}

func PutBullet2Etcd() {
	etcdCli.DelValueWithPrefix(etcd.ETCDKEY_BULLETIN_PREFIX)
	logger.Logger.Trace("ETCD_PutBullet2Etcd")
	count := len(BulletMsgList)
	cur := 0
	for name, p := range BulletMsgList {
		data, err := json.Marshal(p)
		if err == nil {
			_, err := etcdCli.PutValue(fmt.Sprintf("%v%v", etcd.ETCDKEY_BULLETIN_PREFIX, name), string(data))
			if err != nil {
				logger.Logger.Tracef("PutCustomer2Etcd err:%v data:%v", err, string(data))
			}
		} else {
			logger.Logger.Errorf("PutCustomer2Etcd err:%v data:%v", err, string(data))
		}
		cur++
		fmt.Printf("\r")
		fmt.Printf("%d%%", int(float64(cur)/float64(count)*100))
	}
}

func CompBullet() {
	logger.Logger.Trace("LoadBullet")
	buff, err := webapi.API_GetBulletData(common.GetAppId())
	//logger.Logger.Warn("bulletin buff: ", string(buff))
	if err == nil {
		info := ApiBulletResult{}
		err = json.Unmarshal([]byte(buff), &info)
		if err == nil {
			for i := 0; i < len(info.Msg); i++ {
				BulletMsgList[info.Msg[i].Id] = &info.Msg[i]
			}

			res, err := etcdCli.GetValueWithPrefix(etcd.ETCDKEY_BULLETIN_PREFIX)
			if err == nil {
				for i := int64(0); i < res.Count; i++ {
					var info Bullet
					err = json.Unmarshal(res.Kvs[i].Value, &info)
					if err == nil {
						if BulletMsgList[info.Id] == nil {
							etcdCli.DelValue(string(res.Kvs[i].Key))
							//logger.Logger.Errorf("etcd json pasre : ", string(res.Kvs[i].Key))
						}
					} else {
						logger.Logger.Errorf("etcd json pasre err: ", res.Kvs[i].Key, err)
					}

				}
			} else {
				logger.Logger.Errorf("etcd get WithPrefix(%v) panic:%v", etcd.ETCDKEY_BULLETIN_PREFIX, err)
				return
			}

		} else {
			logger.Logger.Error("Unmarshal Bullet data error:", err, string(buff))
			return
		}
	} else {
		logger.Logger.Error("Get Bullet data error:", err)
		return
	}

}

type Customer struct {
	Id             int32
	Platform       string
	Weixin_account string
	Qq_account     string
	Headurl        string
	Nickname       string
	Status         int
	Ext            string
} //声明与world中一样的结构体
type ApiCustomerResult struct {
	Tag int
	Msg []Customer
}

var CustomerMsgList = make(map[int32]*Customer)

func LoadCustomer() {
	logger.Logger.Trace("LoadCustomer")
	buff, err := webapi.API_GetCustomerData(common.GetAppId())
	//logger.Logger.Trace("customer buff:", string(buff))
	if err == nil {
		c_info := ApiCustomerResult{}
		err = json.Unmarshal([]byte(buff), &c_info)
		if err == nil {
			for i := 0; i < len(c_info.Msg); i++ {
				CustomerMsgList[c_info.Msg[i].Id] = &c_info.Msg[i]
			}
		} else {
			logger.Logger.Trace("CustomerMgr is Unmarshal error.", err, string(buff))
		}
	} else {
		logger.Logger.Trace("API_GetCustomerData is error. ", err)
	}
}

func PutCustomer2Etcd() {
	etcdCli.DelValueWithPrefix(etcd.ETCDKEY_AGENTCUSTOMER_PREFIX)
	logger.Logger.Trace("ETCD_PutCustomer2Etcd")
	count := len(CustomerMsgList)
	cur := 0
	for name, p := range CustomerMsgList {
		data, err := json.Marshal(p)
		if err == nil {
			_, err := etcdCli.PutValue(fmt.Sprintf("%v%v", etcd.ETCDKEY_AGENTCUSTOMER_PREFIX, name), string(data))
			if err != nil {
				logger.Logger.Tracef("PutCustomer2Etcd err:%v data:%v", err, string(data))
			}
		} else {
			logger.Logger.Errorf("PutCustomer2Etcd err:%v data:%v", err, string(data))
		}
		cur++
		fmt.Printf("\r")
		fmt.Printf("%d%%", int(float64(cur)/float64(count)*100))
	}
}

type BlackInfo struct {
	Id            int32
	BlackType     int   //1.游戏2.兑换3.充值
	MemberId      int32 //
	AlipayAccount string
	AlipayName    string
	Bankcard      string
	Ip            string //support like "192.0.2.0/24" or "2001:db8::/32", as defined in RFC 4632 and RFC 4291.
	Platform      string
	PackageTag    string
	ipNet         *net.IPNet
}
type BlackInfoApi struct {
	Id             int32
	Space          int32
	Snid           int32
	Alipay_account string
	Alipay_name    string
	Bankcard       string
	Ip             string
	Platform       string
	PackageTag     string
	Explain        string
	Creator        int32
	Create_time    string
	Update_time    string
} //声明与world一样结构体
var BlackList = make(map[int32]BlackInfoApi)

func LoadBlackList() {
	type ApiResult struct {
		Tag int
		Msg []BlackInfoApi
	}
	logger.Logger.Trace("LoadBlackList")
	page := 1
	for {
		logger.Logger.Trace("LoadBlackList req page:%v", page)
		var buff []byte
		var err error
		if common.CustomConfig.GetBool("UseBlacklistBindPlayerinfo") {
			buff, err = webapi.API_GetCommonBlackData(common.GetAppId(), page)
		} else {
			buff, err = webapi.API_GetBlackData(common.GetAppId(), page)
		}

		if err == nil {
			ar := ApiResult{}
			err = json.Unmarshal(buff, &ar)
			if err == nil {
				for _, value := range ar.Msg {
					BlackList[value.Id] = value
				}
				if len(ar.Msg) < 5000 {
					break
				} else {
					page++
				}
			} else {
				logger.Logger.Error("Unmarshal black data error:", err, string(buff))
				break
			}
		} else {
			logger.Logger.Error("Init black list failed.", err, string(buff))
			break
		}
	}
}

func PutBlackList2Etcd() {
	etcdCli.DelValueWithPrefix(etcd.ETCDKEY_BLACKLIST_PREFIX)
	logger.Logger.Trace("ETCD_PutBlackList2Etcd")
	count := len(BlackList)
	cur := 0
	for name, p := range BlackList {
		data, err := json.Marshal(&p)
		if err == nil {
			_, err := etcdCli.PutValue(fmt.Sprintf("%v%v", etcd.ETCDKEY_BLACKLIST_PREFIX, name), string(data))
			if err != nil {
				logger.Logger.Tracef("PutBlackList2Etcd err:%v data:%v", err, string(data))
			} else {
				//logger.Logger.Tracef("PutBlackList2Etcd succes:%v data:%v", err, string(data))
			}
			//time.Sleep(time.Millisecond * 200)
		} else {
			logger.Logger.Errorf("PutBlackList2Etcd err:%v data:%v", err, string(data))
		}
		cur++
		fmt.Printf("\r")
		fmt.Printf("%d%%", int(float64(cur)/float64(count)*100))
	}
}

type SignConfig struct {
	ConfigTickets int64 //tickets
	Platform      string
	StartAct      int32     //活动是否开启
	StartTickets  int64     //tickets
	Reward        [][]int64 // [vipLevel][dayIndex]
} //声明与world中一样的结构体
var SignConfigs = make(map[string]SignConfig)

func LoadActSign() {
	type SignConfigApi struct {
		Params []SignConfig
	}
	type ApiResult struct {
		Tag int
		Msg SignConfigApi
	}
	logger.Logger.Trace("LoadActSign")
	buff, err := webapi.API_GetPlatformSignConfig(common.GetAppId())
	if err == nil {
		ar := ApiResult{}
		err = json.Unmarshal(buff, &ar)
		if err == nil {
			for _, value := range ar.Msg.Params {
				if value.StartAct > 0 {
					SignConfigs[value.Platform] = value
					//logger.Logger.Trace("value:", value)
				}
			}
		} else {
			logger.Logger.Error("Unmarshal ActSign data error:", err, " buff:", string(buff))
		}
	} else {
		logger.Logger.Error("Init ActSign list failed.")
	}

}

func PutActSign2Etcd() {
	etcdCli.DelValueWithPrefix(etcd.ETCDKEY_ACT_SIGNIN_PREFIX)
	logger.Logger.Trace("ETCD_PutActSign2Etcd")
	count := len(SignConfigs)
	cur := 0
	for name, p := range SignConfigs {
		data, err := json.Marshal(p)
		if err == nil {
			_, err := etcdCli.PutValue(fmt.Sprintf("%v%v", etcd.ETCDKEY_ACT_SIGNIN_PREFIX, name), string(data))
			if err != nil {
				logger.Logger.Tracef("PutActSign2Etcd err:%v data:%v", err, string(data))
			}
		} else {
			logger.Logger.Errorf("PutActSign2Etcd err:%v data:%v", err, string(data))
		}
		cur++
		fmt.Printf("\r")
		fmt.Printf("%d%%", int(float64(cur)/float64(count)*100))
	}
}

type GoldTaskConfig struct {
	TaskId       int32 //
	TaskDsc      string
	StartAct     int32   //活动是否开启, 0:关闭 1:开启
	GameId       int32   //
	TaskSort     int32   //0,游戏对局; 1,打死某类鱼
	TaskParam    []int64 //活动参数 0;[count]; 1[fishid, count]
	LimitTimes   int32   //限制次数
	Reward       int64   //奖励
	DurationSort int32   //持续时间  0:永久有效, 1：每天清除; 2,...
}

type APIGoldTaskConfigs struct {
	Datas     []GoldTaskConfig //[taskid]
	Platform  string           //平台
	ConfigVer int64            //时间戳
	StartAct  int32            //开关
} //声明与world中一样的结构体

var ActGoldTaskMap = make(map[string]APIGoldTaskConfigs)

func LoadActGoldTask() {
	type ApiResult struct {
		Tag int
		Msg []APIGoldTaskConfigs
	}
	logger.Logger.Trace("LoadActGoldTask")
	buff, err := webapi.API_GetGoldTaskConfig(common.GetAppId())
	if err == nil {
		ar := ApiResult{}
		err = json.Unmarshal(buff, &ar)
		if err == nil {
			//logger.Logger.Trace("API_GetGoldTaskConfig response:", string(buff))
			for _, plateformConfig := range ar.Msg {
				//logger.Logger.Trace("Platform:", plateformConfig.Platform, " ConfigVer:", plateformConfig.ConfigVer)
				ActGoldTaskMap[plateformConfig.Platform] = plateformConfig
			}
		} else {
			logger.Logger.Error("Unmarshal ActGoldTaskMgr data error:", err, " buff:", string(buff))
		}
	} else {
		logger.Logger.Error("Init ActGoldTaskMgr list failed.")
	}
}

func PutActGoldTask2Etcd() {
	etcdCli.DelValueWithPrefix(etcd.ETCDKEY_ACT_GOLDTASK_PREFIX)
	logger.Logger.Trace("ETCD_PutActGoldTask2Etcd")
	count := len(ActGoldTaskMap)
	cur := 0
	for name, p := range ActGoldTaskMap {
		data, err := json.Marshal(p)
		if err == nil {
			_, err := etcdCli.PutValue(fmt.Sprintf("%v%v", etcd.ETCDKEY_ACT_GOLDTASK_PREFIX, name), string(data))
			if err != nil {
				logger.Logger.Tracef("PutActGoldTask2Etcd err:%v data:%v", err, string(data))
			}
		} else {
			logger.Logger.Errorf("PutActGoldTask2Etcd err:%v data:%v", err, string(data))
		}

		cur++
		fmt.Printf("\r")
		fmt.Printf("%d%%", int(float64(cur)/float64(count)*100))
	}
}

type GoldComeConfig struct {
	TaskId      int32 //
	TaskDsc     string
	StartAct    int32 //活动是否开启, 0:关闭 1:开启
	GameId      int32 //
	WinType     int32 //1,输赢分数; 2下注流水
	MinTimes    int64 //责任局数
	BeginHour   int32
	BeginMinute int32
	EndHour     int32
	EndMinute   int32
	Reward      []int64 //奖励1~20名奖励
} //声明与world中一样的结构体
type RobotType int32

const (
	ROBOTTYPE_ROBOT_CLOSE RobotType = iota //0:不使用机器人
	ROBOTTYPE_ROBOT_OPEN            = 1    //1:使用机器人
)

type APIGoldComeConfig struct {
	Datas     []GoldComeConfig //[taskid]
	Platform  string           //平台
	ConfigVer int64            //版本号
	StartAct  int32            //开启标记
	RobotType RobotType        //机器人类型
} //声明与world中一样的结构体

var ActGoldComeMap = make(map[string]APIGoldComeConfig)

func LoadActGoldCome() {
	type ApiResult struct {
		Tag int
		Msg []APIGoldComeConfig
	}
	logger.Logger.Trace("LoadActGoldCome")
	buff, err := webapi.API_GetGoldComeConfig(common.GetAppId())
	if err == nil {
		ar := ApiResult{}
		err = json.Unmarshal(buff, &ar)
		if err == nil {
			for _, plateformConfig := range ar.Msg {
				ActGoldComeMap[plateformConfig.Platform] = plateformConfig
			}
		} else {
			logger.Logger.Error("Unmarshal ActGoldComeMgr data error:", err, " buff:", string(buff))
		}
	} else {
		logger.Logger.Error("Init ActGoldComeMgr list failed.")
	}

}

func PutActGoldCome2Etcd() {
	etcdCli.DelValueWithPrefix(etcd.ETCDKEY_ACT_GOLDCOME_PREFIX)
	logger.Logger.Trace("ETCD_PutActGoldCome2Etcd")
	count := len(ActGoldComeMap)
	cur := 0
	for name, p := range ActGoldComeMap {
		data, err := json.Marshal(p)
		if err == nil {
			_, err := etcdCli.PutValue(fmt.Sprintf("%v%v", etcd.ETCDKEY_ACT_GOLDCOME_PREFIX, name), string(data))
			if err != nil {
				logger.Logger.Tracef("PutActGoldCome2Etcd err:%v data:%v", err, string(data))
			}
		} else {
			logger.Logger.Errorf("PutActGoldCome2Etcd err:%v data:%v", err, string(data))
		}
		cur++
		fmt.Printf("\r")
		fmt.Printf("%d%%", int(float64(cur)/float64(count)*100))
	}
}

type OnlineRewardConfig struct {
	Id             int32
	OnlineDuration int32
	RewardAmount   int64
}

type PlatformOnlineRewardConfig struct {
	Platform     string
	StartAct     int32                //活动是否开启
	StartTickets int64                //tickets
	Version      int32                // 活动版本
	PayNeed      int32                //充值需要的
	Reward       []OnlineRewardConfig `json:"Data,omitempty"`
} //声明与world中一样的结构体

type ApiOnlineRewardDatas struct {
	Params []PlatformOnlineRewardConfig
}

var ActOnlineRewardMap = make(map[string]PlatformOnlineRewardConfig)

type LuckyTurntableConfig struct {
	TurntableType  int32   `json:"Type"`     // 转盘类型
	PoolInitAmount int64   `json:"PoolCoin"` // 水池初始额度
	ScoreCost      int64   // 积分消耗
	Reward         []int64 // 奖励：依次为小奖1、小奖2、中奖1、中奖2、大奖1、大奖2、特大奖1、特大奖2
	PoolModify     int64   `json:"PoolChange"` // 水池变化额度
}

type PlatformLuckyTurntableConfig struct {
	Platform     string                  // 平台名
	StartAct     int32                   // 活动是否开启
	StartTickets int64                   // tickets
	Version      int32                   // 活动版本
	Turntables   []*LuckyTurntableConfig `json:"Data"`

	MapTurntables map[int32]*LuckyTurntableConfig
} //声明与world一样的结构体

type ApiResultDatas struct {
	Params []PlatformLuckyTurntableConfig
}

var LuckyTurntableMap = make(map[string]PlatformLuckyTurntableConfig)

//平台幸运转盘活动
func LoadLuckyTurntableConfig() {
	type ApiResult struct {
		Tag int
		Msg ApiResultDatas
	}
	logger.Logger.Trace("LoadLuckyTurntableConfig")
	buff, err := webapi.API_GetLuckyTurntableConfig(common.GetAppId())
	if err == nil {
		ar := ApiResult{}
		err = json.Unmarshal(buff, &ar)
		if err == nil {
			for _, value := range ar.Msg.Params {
				//if value.StartAct > 0 {
				LuckyTurntableMap[value.Platform] = value
				//logger.Logger.Trace("value:", value)
				//}
			}
		} else {
			logger.Logger.Error("LuckyTurntableConfig.Init Unmarshal LuckyTurntableConfig data error:", err, " buff:", string(buff))
		}
	} else {
		logger.Logger.Error("LuckyTurntableConfig.Init webapi.API_GetLuckyTurntableConfig failed.")
	}
}

func PutLuckyTurntableConfig2Etcd() {
	etcdCli.DelValueWithPrefix(etcd.ETCDKEY_ACT_LUCKLYTURNTABLE_PREFIX)
	logger.Logger.Trace("ETCD_PutLuckyTurntableConfig2Etcd")
	count := len(LuckyTurntableMap)
	cur := 0
	for name, p := range LuckyTurntableMap {
		data, err := json.Marshal(p)
		if err == nil {
			_, err := etcdCli.PutValue(fmt.Sprintf("%v%v", etcd.ETCDKEY_ACT_LUCKLYTURNTABLE_PREFIX, name), string(data))
			if err != nil {
				logger.Logger.Tracef("PutLuckyTurntableConfig2Etcd err:%v data:%v", err, string(data))
			}
		} else {
			logger.Logger.Errorf("PutLuckyTurntableConfig2Etcd err:%v data:%v", err, string(data))
		}
		cur++
		fmt.Printf("\r")
		fmt.Printf("%d%%", int(float64(cur)/float64(count)*100))
	}
}
func LoadActOnlineReward() {
	type ApiResult struct {
		Tag int
		Msg ApiOnlineRewardDatas
	}
	logger.Logger.Trace("LoadActOnlineReward")
	buff, err := webapi.API_GetOnlineRewardConfig(common.GetAppId())
	if err == nil {
		ar := ApiResult{}
		err = json.Unmarshal(buff, &ar)
		if err == nil {
			for _, value := range ar.Msg.Params {
				//if value.StartAct > 0 {
				ActOnlineRewardMap[value.Platform] = value
				//logger.Logger.Trace("value:", value)
				//}
			}
		} else {
			logger.Logger.Error("ActOnlineRewardMgr.Init Unmarshal ActOnlineReward data error:", err, " buff:", string(buff))
		}
	} else {
		logger.Logger.Error("ActOnlineRewardMgr.Init webapi.API_GetOnlineRewardConfig failed.")
	}
}

func PutActOnlineReward2Etcd() {
	etcdCli.DelValueWithPrefix(etcd.ETCDKEY_ACT_ONLINEREWARD_PREFIX)
	logger.Logger.Trace("ETCD_PutActOnlineReward2Etcd")
	count := len(ActOnlineRewardMap)
	cur := 0
	for name, p := range ActOnlineRewardMap {
		data, err := json.Marshal(p)
		if err == nil {
			_, err := etcdCli.PutValue(fmt.Sprintf("%v%v", etcd.ETCDKEY_ACT_ONLINEREWARD_PREFIX, name), string(data))
			if err != nil {
				logger.Logger.Tracef("PutActOnlineReward2Etcd err:%v data:%v", err, string(data))
			}
		} else {
			logger.Logger.Errorf("PutActOnlineReward2Etcd err:%v data:%v", err, string(data))
		}
		cur++
		fmt.Printf("\r")
		fmt.Printf("%d%%", int(float64(cur)/float64(count)*100))
	}
}

type YebConfig struct {
	Platform     string // 平台名
	StartAct     int32  // 活动是否开启
	StartTickets int64  // tickets
	Version      int32  // 活动版本

	Rate int64 // 配置利率（每万元收益，单位：分）
} //与world中一样

var YebConfigMap = make(map[string]YebConfig)

//平台余额宝活动
func LoadYebConfig() {
	type ApiResultDatas struct {
		Params []YebConfig
	}

	type ApiResult struct {
		Tag int
		Msg ApiResultDatas
	}
	logger.Logger.Trace("LoadYebConfig")
	buff, err := webapi.API_GetYebConfig(common.GetAppId())
	if err == nil {
		ar := ApiResult{}
		err = json.Unmarshal(buff, &ar)
		if err == nil {
			for _, value := range ar.Msg.Params {
				//if value.StartAct > 0 {
				YebConfigMap[value.Platform] = value
				//logger.Logger.Trace("value:", value)
				//}
			}
		} else {
			logger.Logger.Error("API_GetYebConfig.Init Unmarshal API_GetYebConfig data error:", err, " buff:", string(buff))
		}
	} else {
		logger.Logger.Error("API_GetYebConfig.Init webapi.API_GetYebConfig failed.")
	}
}

func PutYeb2ConfigEtcd() {
	etcdCli.DelValueWithPrefix(etcd.ETCDKEY_ACT_YEB_PREFIX)
	logger.Logger.Trace("ETCD_PutYeb2ConfigEtcd")
	count := len(YebConfigMap)
	cur := 0
	for name, p := range YebConfigMap {
		data, err := json.Marshal(p)
		if err == nil {
			_, err := etcdCli.PutValue(fmt.Sprintf("%v%v", etcd.ETCDKEY_ACT_YEB_PREFIX, name), string(data))
			if err != nil {
				logger.Logger.Tracef("PutYeb2ConfigEtcd err:%v data:%v", err, string(data))
			}
		} else {
			logger.Logger.Errorf("PutYeb2ConfigEtcd err:%v data:%v", err, string(data))
		}
		cur++
		fmt.Printf("\r")
		fmt.Printf("%d%%", int(float64(cur)/float64(count)*100))
	}
}

type RebateGameCfg struct {
	BaseCoin      [3]int32 //返利基准
	RebateRate    [3]int32 //返利比率
	GameId        int32    //游戏id
	GameMode      int32    //游戏类型
	MaxRebateCoin int64    //最高返利

}
type RebateGameThirdCfg struct {
	BaseCoin      [3]int32 //返利基准
	RebateRate    [3]int32 //返利比率
	ThirdName     string   //第三方key
	ThirdShowName string   //前端显示name
	MaxRebateCoin int64    //最高返利
	ThirdId       string   //三方游戏id
}
type RebateInfo struct {
	Platform           string                //平台名称
	RebateSwitch       bool                  //返利开关
	RebateManState     int                   //返利是开启个人返利  0 关闭  1 开启
	ReceiveMode        int                   //领取方式  0实时领取  1次日领取
	NotGiveOverdue     int                   //0不过期   1过期不给  2过期邮件给
	RebateGameCfg      []*RebateGameCfg      //key为"gameid"+"gamemode"
	RebateGameThirdCfg []*RebateGameThirdCfg //第三方key
	Version            int                   //活动版本 后台控制
}
type RebateTask struct {
	Platform           string                         //平台名称
	RebateSwitch       bool                           //返利开关
	RebateManState     int                            //返利是开启个人返利  0 关闭  1 开启
	ReceiveMode        int                            //领取方式  0实时领取  1次日领取
	NotGiveOverdue     int                            //0不过期   1过期不给  2过期邮件给
	RebateGameCfg      map[string]*RebateGameCfg      //key为"gameid"+"gamemode"
	RebateGameThirdCfg map[string]*RebateGameThirdCfg //第三方key
	Version            int                            //活动版本 后台控制
}

var RebateTaskMap = make(map[string]*RebateTask)

func LoadRebateData() {
	logger.Logger.Trace("LoadRebateData")

	//获取平台返利数据
	type ApiResult struct {
		Tag int
		Msg []RebateInfo
	}
	rebateBuff, err := webapi.API_GetGameRebateConfig(common.GetAppId())
	if err == nil {
		logger.Logger.Trace("API_GetGameRebateConfig:", string(rebateBuff))
		ar := ApiResult{}
		err = json.Unmarshal(rebateBuff, &ar)
		if err == nil && ar.Tag == 0 {
			for _, v := range ar.Msg {
				RebateTaskMap[v.Platform] = &RebateTask{
					Platform:       v.Platform,
					RebateSwitch:   v.RebateSwitch,
					ReceiveMode:    v.ReceiveMode,
					RebateManState: v.RebateManState,
					NotGiveOverdue: v.NotGiveOverdue,
					Version:        v.Version,
				}
				RebateTaskMap[v.Platform].RebateGameCfg = make(map[string]*RebateGameCfg)
				RebateTaskMap[v.Platform].RebateGameThirdCfg = make(map[string]*RebateGameThirdCfg)
				for _, cfg := range v.RebateGameCfg {
					for _, dfm := range srvdata.PBDB_GameFreeMgr.Datas.Arr {

						if dfm.GetGameId() == cfg.GameId && dfm.GetGameMode() == cfg.GameMode {
							RebateTaskMap[v.Platform].RebateGameCfg[dfm.GetGameDif()] = cfg
							break
						}
					}
				}

				for _, cfg := range v.RebateGameThirdCfg {
					RebateTaskMap[v.Platform].RebateGameThirdCfg[cfg.ThirdId] = cfg
				}
			}
		} else {
			logger.Logger.Error("Unmarshal RebateTask data error:", err, string(rebateBuff))
		}
	} else {
		logger.Logger.Error("Get RebateTask data error:", err)
	}
}
func PutRebateTaskEtcd() {
	etcdCli.DelValueWithPrefix(etcd.ETCDKEY_CONFIG_REBATE)
	logger.Logger.Trace("ETCD_PutRebateTaskEtcd")
	count := len(RebateTaskMap)
	cur := 0
	for name, p := range RebateTaskMap {
		data, err := json.Marshal(p)
		if err == nil {
			_, err := etcdCli.PutValue(fmt.Sprintf("%v%v", etcd.ETCDKEY_CONFIG_REBATE, name), string(data))
			if err != nil {
				logger.Logger.Tracef("PutRebateTaskEtcd err:%v data:%v", err, string(data))
			}
		} else {
			logger.Logger.Errorf("PutRebateTaskEtcd err:%v data:%v", err, string(data))
		}

		cur++
		fmt.Printf("\r")
		fmt.Printf("%d%%", int(float64(cur)/float64(count)*100))
	}
}

type PromoterConfig struct {
	PromoterID             string //代理ID
	Platform               string //平台
	PromoterType           int32  //代理类型 0 全民 1无限代理 2 渠道 3 推广员
	UpgradeAccountGiveCoin int32  //升级账号奖励金币
	NewAccountGiveCoin     int32  //新账号奖励金币
	ExchangeTax            int32  //兑换税收（万分比）
	ExchangeForceTax       int32  //强制兑换税收
	ExchangeFlow           int32  //兑换流水比例
	ExchangeGiveFlow       int32  //赠送兑换流水比例
	ExchangeFlag           int32  //兑换标记
	IsInviteRoot           int32  //是否绑定全民用户
}

var PromoterMap = make(map[string]*PromoterConfig)

func LoadPromoterData() {
	logger.Logger.Trace("LoadPromoterData")

	//获取平台返利数据
	type ApiResult struct {
		Tag int
		Msg []PromoterConfig
	}
	rebateBuff, err := webapi.API_GetPromoterConfig(common.GetAppId())
	if err == nil {
		logger.Logger.Trace("API_GetPromoterConfig:", string(rebateBuff))
		ar := ApiResult{}
		err = json.Unmarshal(rebateBuff, &ar)
		if err == nil && ar.Tag == 0 {
			for _, v := range ar.Msg {
				temp := v
				PromoterMap[v.PromoterID+"_"+strconv.Itoa(int(v.PromoterType))] = &temp
			}
		} else {
			logger.Logger.Error("Unmarshal Promoter data error:", err, string(rebateBuff))
		}
	} else {
		logger.Logger.Error("Get Promoter data error:", err)
	}
}
func PutPromoterEtcd() {
	etcdCli.DelValueWithPrefix(etcd.ETCDKEY_PROMOTER_PREFIX)
	logger.Logger.Trace("ETCD_PutPromoterEtcd")
	count := len(PromoterMap)
	cur := 0
	for name, p := range PromoterMap {
		data, err := json.Marshal(p)
		if err == nil {
			_, err := etcdCli.PutValue(fmt.Sprintf("%v%v", etcd.ETCDKEY_PROMOTER_PREFIX, name), string(data))
			if err != nil {
				logger.Logger.Tracef("ETCD_PutPromoterEtcd err:%v data:%v", err, string(data))
			}
		} else {
			logger.Logger.Errorf("ETCD_PutPromoterEtcd err:%v data:%v", err, string(data))
		}

		cur++
		fmt.Printf("\r")
		fmt.Printf("%d%%", int(float64(cur)/float64(count)*100))
	}
}

type ActVipRewardConfig struct {
	LevelCoin int32 //等级奖励金币
	DayCoin   int32 //每日奖励
	WeekCoin  int32 //每周奖励
	MonthCoin int32 //每月奖励
}

type ActVipPlateformConfig struct {
	VipBonusInfo []ActVipRewardConfig //奖励信息
	Platform     string               //平台
	StartAct     int32                //活动开启标记 0:关闭 1:开启
}

var ActVipMap = make(map[string]*ActVipPlateformConfig)

func LoadActVipData() {
	logger.Logger.Trace("LoadActVipData")

	//获取平台返利数据
	type ApiResult struct {
		Tag int
		Msg []ActVipPlateformConfig
	}

	rebateBuff, err := webapi.API_GetActVipConfig(common.GetAppId())
	if err == nil {
		logger.Logger.Trace("API_GetPromoterConfig:", string(rebateBuff))
		ar := ApiResult{}
		err = json.Unmarshal(rebateBuff, &ar)
		if err == nil && ar.Tag == 0 {
			for _, v := range ar.Msg {
				temp := v
				ActVipMap[temp.Platform] = &temp
			}
		} else {
			logger.Logger.Error("Unmarshal ActVip data error:", err, string(rebateBuff))
		}
	} else {
		logger.Logger.Error("Get ActVip data error:", err)
	}
}
func PutActVipEtcd() {
	etcdCli.DelValueWithPrefix(etcd.ETCDKEY_ACT_VIP_PREFIX)
	logger.Logger.Trace("ETCD_PutActVipEtcd")
	count := len(ActVipMap)
	cur := 0
	for name, p := range ActVipMap {
		data, err := json.Marshal(p)
		if err == nil {
			_, err := etcdCli.PutValue(fmt.Sprintf("%v%v", etcd.ETCDKEY_ACT_VIP_PREFIX, name), string(data))
			if err != nil {
				logger.Logger.Tracef("ETCD_ActVipMapEtcd err:%v data:%v", err, string(data))
			}
		} else {
			logger.Logger.Errorf("ETCD_ActVipMapEtcd err:%v data:%v", err, string(data))
		}

		cur++
		fmt.Printf("\r")
		fmt.Printf("%d%%", int(float64(cur)/float64(count)*100))
	}
}

type ShareConfig struct {
	Platform  string // 平台id
	Status    int32  // 开关
	Number    int64  // 分享彩金
	Times     int64  // 每日领取次数
	FriendUrl string
	SpaceUrl  string
	VipLevel  int32 // vip等级下限
	Recharge  int64 // 充值下限
}

var ActWxShareMap = make(map[string]*ShareConfig)

func LoadActWxShareData() {
	logger.Logger.Trace("LoadActWxShareData")

	//获取平台返利数据
	type ApiResult struct {
		Tag int
		Msg []ShareConfig
	}

	rebateBuff, err := webapi.API_GetWeiXinShareConfig(common.GetAppId())
	if err == nil {
		//logger.Logger.Trace("API_GetPromoterConfig:", string(rebateBuff))
		ar := ApiResult{}
		err = json.Unmarshal(rebateBuff, &ar)
		if err == nil && ar.Tag == 0 {
			for _, v := range ar.Msg {
				temp := v
				ActWxShareMap[temp.Platform] = &temp
			}
		} else {
			logger.Logger.Error("Unmarshal ActWxShare data error:", err, string(rebateBuff))
		}
	} else {
		logger.Logger.Error("Get ActWxShare data error:", err)
	}
}
func PutActWxShareEtcd() {
	etcdCli.DelValueWithPrefix(etcd.ETCDKEY_ACT_WEIXIN_SHARE_PREFIX)
	logger.Logger.Trace("ETCD_PutActEtcd")
	count := len(ActWxShareMap)
	cur := 0
	for name, p := range ActWxShareMap {
		data, err := json.Marshal(p)
		if err == nil {
			_, err := etcdCli.PutValue(fmt.Sprintf("%v%v", etcd.ETCDKEY_ACT_WEIXIN_SHARE_PREFIX, name), string(data))
			if err != nil {
				logger.Logger.Tracef("ETCD_PutActWxShareEtcd err:%v data:%v", err, string(data))
			}
		} else {
			logger.Logger.Errorf("ETCD_PutActWxShareEtcd err:%v data:%v", err, string(data))
		}

		cur++
		fmt.Printf("\r")
		fmt.Printf("%d%%", int(float64(cur)/float64(count)*100))
	}
}

type ActGiveConfig struct {
	Tag      int32 //赠与类型
	IsStop   int32 //是否阻止 1阻止 0放开
	NeedFlow int32
}

type ActGivePlateformConfig struct {
	ActInfo  map[string]*ActGiveConfig //奖励信息
	Platform string                    //平台
}

var ActMap = make(map[string]*ActGivePlateformConfig)

func LoadActData() {
	logger.Logger.Trace("LoadActData")

	//获取平台返利数据
	type ApiResult struct {
		Tag int
		Msg []ActGivePlateformConfig
	}

	rebateBuff, err := webapi.API_GetActConfig(common.GetAppId())
	if err == nil {
		//logger.Logger.Trace("API_GetPromoterConfig:", string(rebateBuff))
		ar := ApiResult{}
		err = json.Unmarshal(rebateBuff, &ar)
		if err == nil && ar.Tag == 0 {
			for _, v := range ar.Msg {
				temp := v
				ActMap[temp.Platform] = &temp
			}
		} else {
			logger.Logger.Error("Unmarshal Act data error:", err, string(rebateBuff))
		}
	} else {
		logger.Logger.Error("Get Act data error:", err)
	}
}
func PutAcEtcd() {
	etcdCli.DelValueWithPrefix(etcd.ETCDKEY_ACT_GIVE_PREFIX)
	logger.Logger.Trace("ETCD_PutActEtcd")
	count := len(ActMap)
	cur := 0
	for name, p := range ActMap {
		data, err := json.Marshal(p)
		if err == nil {
			_, err := etcdCli.PutValue(fmt.Sprintf("%v%v", etcd.ETCDKEY_ACT_GIVE_PREFIX, name), string(data))
			if err != nil {
				logger.Logger.Tracef("ETCD_PutActEtcd err:%v data:%v", err, string(data))
			}
		} else {
			logger.Logger.Errorf("ETCD_PutActEtcd err:%v data:%v", err, string(data))
		}

		cur++
		fmt.Printf("\r")
		fmt.Printf("%d%%", int(float64(cur)/float64(count)*100))
	}
}

type RandRange struct {
	Low  int32
	High int32
}

type RandCoinData struct {
	Id        int //红包活动编号
	Platform  string
	IsOpen    bool
	DType     int         //红包类型	0:全量红包 1：随机红包 2：五福红包
	StartTs   int64       //开始时间
	EndTs     int64       //结束时间
	CoinTs    int64       //可以领奖的时间（五福红包有特殊的时间，其他的和开始时间是一样的）
	Title     string      //红包标题
	Context   string      //红包文案
	Count     int32       //红包数量
	TotleCoin int64       //红包总额度(写错成了totlecoin，想修改成coin，但是这个需要后端配合修改，后端改起来麻烦，所以不要纠结这个变量名）
	VipSel    []int       //红包VIP选项	勾选的VIP等级
	ExtCoin   int64       //红包额外条件
	RandCoin  []RandRange //VIP随机范围（随机红包特有）
	NewYear   []int       //五福参数（五福红包特有）邀请好友数量，好友流水，每日充值，游戏局数，个人流水
}

type ActConfig struct {
	Platform    string
	IsOpen      bool
	ActivityArr []*RandCoinData
}

var RandomPrizeMap = make(map[string]*ActConfig)

func LoadRandomPrizeData() {
	logger.Logger.Trace("LoadRandomPrizeData")

	//获取平台返利数据
	type ApiResult struct {
		Tag int
		Msg []*ActConfig
	}

	rebateBuff, err := webapi.API_GetRandCoinData(common.GetAppId())
	if err == nil {
		//logger.Logger.Trace("API_GetPromoterConfig:", string(rebateBuff))
		ar := ApiResult{}
		err = json.Unmarshal(rebateBuff, &ar)
		if err == nil && ar.Tag == 0 {
			for _, v := range ar.Msg {
				RandomPrizeMap[v.Platform] = v
			}
		} else {
			logger.Logger.Error("Unmarshal RandomPrize data error:", err, string(rebateBuff))
		}
	} else {
		logger.Logger.Error("Get RandomPrize data error:", err)
	}
}

func PutRandomPrizeEtcd() {
	etcdCli.DelValueWithPrefix(etcd.ETCDKEY_ACT_RANDCOIN_PREFIX)
	logger.Logger.Trace("ETCD_PutActEtcd")
	count := len(RandomPrizeMap)
	cur := 0
	for name, p := range RandomPrizeMap {
		data, err := json.Marshal(p)
		if err == nil {
			_, err := etcdCli.PutValue(fmt.Sprintf("%v%v", etcd.ETCDKEY_ACT_RANDCOIN_PREFIX, name), string(data))
			if err != nil {
				logger.Logger.Tracef("ETCD_PutActEtcd err:%v data:%v", err, string(data))
			}
		} else {
			logger.Logger.Errorf("ETCD_PutActEtcd err:%v data:%v", err, string(data))
		}

		cur++
		fmt.Printf("\r")
		fmt.Printf("%d%%", int(float64(cur)/float64(count)*100))
	}
}

type ActFPayWinConfig struct {
	WinType int32 //类型
	WinRate int32 //赢取比例
}

type ActFPayConfig struct {
	StartAct     int32                        //活动开启标记 0:关闭 1:开启
	ConfigVer    int32                        //版本
	PlayerAddMax int32                        //玩家数量  //每分钟增长最大
	PlayerAddMin int32                        //玩家数量  //每分钟增长最小
	StartTime    int32                        //开启时间 //秒
	EndTime      int32                        //结束时间 //秒
	FPayCoin     int32                        //首充金额
	FPayGiveType int32                        //赠送类型 0 金额 1 比例
	FPayGiveCoin int32                        //赠送金额
	NeedWinCoin  int32                        //需要完成赢取金额
	Remark       string                       //备注
	WinConfig    map[string]*ActFPayWinConfig //赢取比例配置
}

type ActFPayPlateformConfig struct {
	FPayInfo *ActFPayConfig //奖励信息
	Platform string         //平台
}

var ActFPayMap = make(map[string]*ActFPayPlateformConfig)

func LoadActFPayData() {
	logger.Logger.Trace("LoadActData")

	//获取平台返利数据
	type ApiResult struct {
		Tag int
		Msg []ActFPayPlateformConfig
	}

	rebateBuff, err := webapi.API_GetActFPayConfig(common.GetAppId())
	if err == nil {
		logger.Logger.Trace("API_GetActFPayConfig:", string(rebateBuff))
		ar := ApiResult{}
		err = json.Unmarshal(rebateBuff, &ar)
		if err == nil && ar.Tag == 0 {
			for _, v := range ar.Msg {
				temp := v
				ActFPayMap[temp.Platform] = &temp
			}
		} else {
			logger.Logger.Error("Unmarshal ActFPay data error:", err, string(rebateBuff))
		}
	} else {
		logger.Logger.Error("Get ActFPay data error:", err)
	}
}
func PutAcFPayEtcd() {
	etcdCli.DelValueWithPrefix(etcd.ETCDKEY_ACT_FPAY_PREFIX)
	logger.Logger.Trace("PutAcFPayEtcd")
	count := len(ActFPayMap)
	cur := 0
	for name, p := range ActFPayMap {
		data, err := json.Marshal(p)
		if err == nil {
			_, err := etcdCli.PutValue(fmt.Sprintf("%v%v", etcd.ETCDKEY_ACT_FPAY_PREFIX, name), string(data))
			if err != nil {
				logger.Logger.Tracef("PutAcFPayEtcd err:%v data:%v", err, string(data))
			}
		} else {
			logger.Logger.Errorf("PutAcFPayEtcd err:%v data:%v", err, string(data))
		}

		cur++
		fmt.Printf("\r")
		fmt.Printf("%d%%", int(float64(cur)/float64(count)*100))
	}
}

func CopyDBGameFreeField(src, dst *protocol.DB_GameFree) {
	dst.Id = src.Id
	dst.Name = src.Name
	dst.Title = src.Title
	dst.ShowType = src.ShowType
	dst.SubShowType = src.SubShowType
	dst.Hot = src.Hot
	dst.GameRule = src.GameRule
	dst.TestTakeCoin = src.TestTakeCoin
	dst.SceneType = src.SceneType
	dst.IsHundred = src.IsHundred
	dst.GameId = src.GameId
	dst.GameMode = src.GameMode
	dst.ShowId = src.ShowId
	dst.ServiceFee = src.ServiceFee
	dst.Turn = src.Turn
	dst.BetDec = src.BetDec
	dst.CorrectNum = src.CorrectNum
	dst.CorrectRate = src.CorrectRate
	dst.Deviation = src.Deviation
	dst.Ready = src.Ready
	dst.Ai = src.Ai
	dst.Jackpot = src.Jackpot
	dst.ElementsParams = src.ElementsParams
	dst.OtherElementsParams = src.OtherElementsParams
	dst.DownRiceParams = src.DownRiceParams
	dst.InitValue = src.InitValue
	dst.LowerLimit = src.LowerLimit
	dst.UpperLimit = src.UpperLimit
	dst.UpperOffsetLimit = src.UpperOffsetLimit
	dst.MaxOutValue = src.MaxOutValue
	dst.ChangeRate = src.ChangeRate
	dst.MinOutPlayerNum = src.MinOutPlayerNum
	dst.UpperLimitOfOdds = src.UpperLimitOfOdds
	if len(dst.RobotNumRng) < 2 {
		dst.RobotNumRng = src.RobotNumRng
	}
	//dst.SameIpLimit = src.SameIpLimit
	dst.BaseRate = src.BaseRate
	dst.CtroRate = src.CtroRate
	dst.HardTimeMin = src.HardTimeMin
	dst.HardTimeMax = src.HardTimeMax
	dst.NormalTimeMin = src.NormalTimeMin
	dst.NormalTimeMax = src.NormalTimeMax
	dst.EasyTimeMin = src.EasyTimeMin
	dst.EasyTimeMax = src.EasyTimeMax
	dst.EasrierTimeMin = src.EasrierTimeMin
	dst.EasrierTimeMax = src.EasrierTimeMax
	dst.GameType = src.GameType
	dst.GameDif = src.GameDif
	dst.GameClass = src.GameClass
	dst.PlatformName = src.PlatformName
	dst.MaxBetCoin = src.MaxBetCoin
	if *dst.Id == 10000001 {
		println(dst.GetMatchTrueMan())
	}
}

//赛事奖励配置
type ActMatchRewardConfig struct {
	RankRange []int32 //名次区间
	Reward    int32   //奖励数量
}

//活动配置
type ActTicketConfig struct {
	Id         int32                   //配置编号
	Platform   string                  //平台编号
	MatchId    int32                   //比赛编号
	StartTime  int64                   //开始时间
	EndTime    int64                   //结束时间
	Enable     bool                    //是否开启
	AutoAgree  bool                    //自动同意
	PlayerType int32                   //玩家类型 0:受邀请的玩家
	AchievType int32                   //成绩类型 0:首日最好成绩
	Rewards    []*ActMatchRewardConfig //奖励配置
}

var ActTicketConfigData = make(map[string]*ActTicketConfig)

func LoadActTicketConfig() {
	type ActTicketConfigMsg struct {
		Tag int
		Msg []*ActTicketConfig
	}
	logger.Logger.Trace("API_GetActTicketConfigData")
	buff, err := webapi.API_GetActTicketConfigData(common.GetAppId())
	if err == nil {
		var data ActTicketConfigMsg
		err = json.Unmarshal(buff, &data)
		if err == nil && data.Tag == 0 {
			for _, cfg := range data.Msg {
				ActTicketConfigData[cfg.Platform] = cfg
			}
		} else {
			logger.Logger.Error("Unmarshal ActTicketConfig config data error:", err, string(buff))
		}
	} else {
		logger.Logger.Error("Get ActTicketConfig config data error:", err)
	}
}
func PutActTicketEtcd() {
	etcdCli.DelValueWithPrefix(etcd.ETCDKEY_ACT_TICKET_PROFIX)
	logger.Logger.Trace("PutActTicketEtcd")
	count := len(ActTicketConfigData)
	cur := 0
	for name, p := range ActTicketConfigData {
		data, err := json.Marshal(p)
		if err == nil {
			_, err := etcdCli.PutValue(fmt.Sprintf("%v%v", etcd.ETCDKEY_ACT_TICKET_PROFIX, name), string(data))
			if err != nil {
				logger.Logger.Tracef("PutAcFPayEtcd err:%v data:%v", err, string(data))
			}
		} else {
			logger.Logger.Errorf("PutAcFPayEtcd err:%v data:%v", err, string(data))
		}

		cur++
		fmt.Printf("\r")
		fmt.Printf("%d%%", int(float64(cur)/float64(count)*100))
	}
}

//比赛配置
type MatchConfig struct {
	Id             int32                 //比赛id
	Platform       string                //平台编号
	Ver            int32                 //数据版本号
	Name           string                //名称
	Desc           string                //比赛说明
	Turn           int32                 //显示排序
	Enable         bool                  //开启标记 false:未开启 true:开启
	StartMatchType int32                 //开赛类型 1:流水赛(人满即开) 2:定点赛(固定时间开赛)
	GameFreeId     int32                 //游戏id 对应gamefreeid
	Onumber        int32                 //最低开赛人数
	Gnumber        int32                 //分组人数
	NumberLimit    int32                 //最高报名数人数
	Rot            bool                  //是否使用机器人 false:不使用 true:使用
	RobCnt         []int32               //邀请机器人数量 5,8 => 5~8
	SameIpLimit    bool                  //同ip限制
	GuestLimit     bool                  //正式玩家限制 false:游客也可报名 true:只能正式玩家报名
	PrizeWashLimit bool                  //奖金是否有打码要求 false:无 true:需要打码
	VipLimit       int32                 //vip限制 二进制模式
	PrizeLimit     int32                 //单日奖金发放限制 0:不限制 >0:发完停赛
	ShowPrize      int32                 //给客户端展示的奖励限制
	RobOccupyPrize bool                  //机器人是否占用奖励
	SignupCost     MatchSignupConfig     //报名费
	StartTime      []MatchDateTimeConfig //开赛时间
	MatchProcess   []*MatchProcessConfig //赛程阶段
	Reward         []*MatchRewardConfig  //奖励配置
}

//赛程配置
type MatchProcessConfig struct {
	Name          string  //赛程名称
	Desc          string  //赛程描述
	InitGrade     int32   //初始积分
	GradeDiscount int32   //积分折算 0.使用初始积分 1:使用上个赛程的积分 2:上一轮积分进行折算
	MPPType       int32   //赛段赛制类型 1:定局积分 2:打立出局 3:瑞士移位 4:定时积分
	NumOfGames    int32   //局数
	BaseScore     int32   //初始底分
	Filter        int32   //晋级人数
	Params        []int32 //具体规则参数 打立出局：X,Y,Z 每X秒增加Y底分,低于Z分淘汰
	PhaseFlag     int32   //0:海选 1:预赛 2:淘汰赛 3:半决赛 4.决赛
}

//比赛奖励配置
type MatchRewardConfig struct {
	RankRange []int32 //名次区间
	Reward    int32   //奖励金币
}

//比赛报名配置
type MatchSignupConfig struct {
	SupportFreeTimes bool  //是否支持免费
	FreeTimes        int32 //免费次数
	SupportTicket    bool  //是否支持入场券
	TicketCnt        int32 //入场券数量
	SupportCoin      bool  //是否支持金币报名
	Coin             int32 //所需金币
}

//比赛时间配置
type MatchDateTimeConfig struct {
	StartDate int32 //开始日期 如:20200801 0表示不限开始日期
	EndDate   int32 //结束日期 如:20200808 0表示不限结束日期
	StartMini int32 //开始时间 如:1000 10点开赛 0表示不限开始时间
	EndMini   int32 //结束时间 如:1200 12点结束 0表示不限结束时间
}

var MatchConfigData = make(map[int32]*MatchConfig)

func LoadMatchConfig() {
	type ActTicketConfigMsg struct {
		Tag int
		Msg []*MatchConfig
	}
	logger.Logger.Trace("API_GetMatchConfigData")
	buff, err := webapi.API_GetMatchConfigData(common.GetAppId())
	if err == nil {
		var data ActTicketConfigMsg
		err = json.Unmarshal(buff, &data)
		if err == nil && data.Tag == 0 {
			for _, cfg := range data.Msg {
				MatchConfigData[cfg.Id] = cfg
			}
		} else {
			logger.Logger.Error("Unmarshal ActTicketConfig config data error:", err, string(buff))
		}
	} else {
		logger.Logger.Error("Get ActTicketConfig config data error:", err)
	}
}
func PutMatchConfigEtcd() {
	etcdCli.DelValueWithPrefix(etcd.ETCDKEY_MATCH_PROFIX)
	logger.Logger.Trace("PutMatchConfigEtcd")
	count := len(MatchConfigData)
	cur := 0
	for name, p := range MatchConfigData {
		data, err := json.Marshal(p)
		if err == nil {
			_, err := etcdCli.PutValue(fmt.Sprintf("%v%v", etcd.ETCDKEY_MATCH_PROFIX, name), string(data))
			if err != nil {
				logger.Logger.Tracef("PutAcFPayEtcd err:%v data:%v", err, string(data))
			}
		} else {
			logger.Logger.Errorf("PutAcFPayEtcd err:%v data:%v", err, string(data))
		}

		cur++
		fmt.Printf("\r")
		fmt.Printf("%d%%", int(float64(cur)/float64(count)*100))
	}
}
