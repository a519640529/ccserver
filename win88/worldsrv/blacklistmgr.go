package main

import (
	"games.yol.com/win88/model"
	"github.com/globalsign/mgo"
	"github.com/idealeak/goserver/core/logger"
	"net"
	"strings"
)

const (
	BlackState_Login    uint = iota //登录
	BlackState_Exchange             //兑换
	BlackState_Recharge             //充值
	BlackState_Match                //比赛
	BlackState_Max
)

var BlackListMgrSington = &BlackListMgr{
	BlackList: make(map[int32]*BlackInfo),
}

//type BlackListObserver interface {
//	OnAddBlackInfo(blackinfo *BlackInfo)
//	OnEditBlackInfo(blackinfo *BlackInfo)
//	OnRemoveBlackInfo(blackinfo *BlackInfo)
//}

type BlackListMgr struct {
	BlackList            map[int32]*BlackInfo
	BlackListByPlatform  [BlackState_Max]map[string]map[int32]*BlackInfo
	AlipayAccByPlatform  [BlackState_Max]map[string]map[string]*BlackInfo
	AlipayNameByPlatform [BlackState_Max]map[string]map[string]*BlackInfo
	BankcardByPlatform   [BlackState_Max]map[string]map[string]*BlackInfo
	IpByPlatform         [BlackState_Max]map[string]map[string]*BlackInfo
	IpNetByPlatform      [BlackState_Max]map[string][]*BlackInfo
	PackageTagByPlatform [BlackState_Max]map[string]*BlackInfo
	DeviceByPlatform     [BlackState_Max]map[string]map[string]*BlackInfo
	//Observers            []BlackListObserver
}

type BlackInfo struct {
	Id             int32
	BlackType      int //1.游戏2.兑换3.充值4.比赛
	Alipay_account string
	Alipay_name    string
	Bankcard       string
	Ip             string //support like "192.0.2.0/24" or "2001:db8::/32", as defined in RFC 4632 and RFC 4291.
	Platform       string
	PackageTag     string
	DeviceId       string //设备ID
	ipNet          *net.IPNet
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
	DeviceId       string //设备ID
}

//func (this *BlackListMgr) RegisteObserver(observer BlackListObserver) {
//	for _, ob := range this.Observers {
//		if ob == observer {
//			return
//		}
//	}
//	this.Observers = append(this.Observers, observer)
//}
//
//func (this *BlackListMgr) UnregisteObserver(observer BlackListObserver) {
//	for i, ob := range this.Observers {
//		if ob == observer {
//			count := len(this.Observers)
//			if i == 0 {
//				this.Observers = this.Observers[1:]
//			} else if i == count-1 {
//				this.Observers = this.Observers[:count-1]
//			} else {
//				arr := this.Observers[:i]
//				arr = append(arr, this.Observers[i+1:]...)
//				this.Observers = arr
//			}
//		}
//	}
//}

func (this *BlackListMgr) Init() {
	if this.BlackList == nil {
		this.BlackList = make(map[int32]*BlackInfo)
	}
	for i := uint(0); i < BlackState_Max; i++ {
		this.BlackListByPlatform[i] = make(map[string]map[int32]*BlackInfo)
		this.AlipayAccByPlatform[i] = make(map[string]map[string]*BlackInfo)
		this.AlipayNameByPlatform[i] = make(map[string]map[string]*BlackInfo)
		this.BankcardByPlatform[i] = make(map[string]map[string]*BlackInfo)
		this.IpByPlatform[i] = make(map[string]map[string]*BlackInfo)
		this.IpNetByPlatform[i] = make(map[string][]*BlackInfo)
		this.PackageTagByPlatform[i] = make(map[string]*BlackInfo)
		this.DeviceByPlatform[i] = make(map[string]map[string]*BlackInfo)
	}
	if !model.GameParamData.UseEtcd {

	} else {
		EtcdMgrSington.InitBlackList()
	}

}

func (this *BlackListMgr) DivBlackInfo(blackInfo *BlackInfo) {
	for i := uint(0); i < BlackState_Max; i++ {
		if blackInfo.BlackType&(1<<i) != 0 {
			blbp := this.BlackListByPlatform[i]
			if blbp != nil {
				if pool, exist := blbp[blackInfo.Platform]; exist {
					pool[blackInfo.Id] = blackInfo
				} else {
					pool = make(map[int32]*BlackInfo)
					if pool != nil {
						pool[blackInfo.Id] = blackInfo
						blbp[blackInfo.Platform] = pool
					}
				}
			}

			//alipay account
			if len(blackInfo.Alipay_account) > 0 {
				aabp := this.AlipayAccByPlatform[i]
				if aabp != nil {
					if pool, exist := aabp[blackInfo.Platform]; exist {
						pool[blackInfo.Alipay_account] = blackInfo
					} else {
						pool = make(map[string]*BlackInfo)
						if pool != nil {
							pool[blackInfo.Alipay_account] = blackInfo
							aabp[blackInfo.Platform] = pool
						}
					}
				}
			}

			//alipay name
			if len(blackInfo.Alipay_name) > 0 {
				anbp := this.AlipayNameByPlatform[i]
				if anbp != nil {
					if pool, exist := anbp[blackInfo.Platform]; exist {
						pool[blackInfo.Alipay_name] = blackInfo
					} else {
						pool = make(map[string]*BlackInfo)
						if pool != nil {
							pool[blackInfo.Alipay_name] = blackInfo
							anbp[blackInfo.Platform] = pool
						}
					}
				}
			}

			//bank
			if len(blackInfo.Bankcard) > 0 {
				bankbp := this.BankcardByPlatform[i]
				if bankbp != nil {
					if pool, exist := bankbp[blackInfo.Platform]; exist {
						pool[blackInfo.Bankcard] = blackInfo
					} else {
						pool = make(map[string]*BlackInfo)
						if pool != nil {
							pool[blackInfo.Bankcard] = blackInfo
							bankbp[blackInfo.Platform] = pool
						}
					}
				}
			}

			//ip
			if len(blackInfo.Ip) > 0 {
				ipbp := this.IpByPlatform[i]
				if ipbp != nil {
					if pool, exist := ipbp[blackInfo.Platform]; exist {
						pool[blackInfo.Ip] = blackInfo
					} else {
						pool = make(map[string]*BlackInfo)
						if pool != nil {
							pool[blackInfo.Ip] = blackInfo
							ipbp[blackInfo.Platform] = pool
						}
					}
				}
			}

			//ipnet
			if blackInfo.ipNet != nil {
				ipbp := this.IpNetByPlatform[i]
				if ipbp != nil {
					if pool, exist := ipbp[blackInfo.Platform]; exist {
						pool = append(pool, blackInfo)
						ipbp[blackInfo.Platform] = pool
					} else {
						ipbp[blackInfo.Platform] = []*BlackInfo{blackInfo}
					}
				}
			}

			//packageid
			if len(blackInfo.PackageTag) > 0 {
				packbp := this.PackageTagByPlatform[i]
				if packbp != nil {
					packbp[blackInfo.PackageTag] = blackInfo
				}
			}

			//deviceinfo
			if len(blackInfo.DeviceId) > 0 {
				ipbp := this.DeviceByPlatform[i]
				if ipbp != nil {
					if pool, exist := ipbp[blackInfo.Platform]; exist {
						pool[blackInfo.DeviceId] = blackInfo
					} else {
						pool = make(map[string]*BlackInfo)
						if pool != nil {
							pool[blackInfo.DeviceId] = blackInfo
							ipbp[blackInfo.Platform] = pool
						}
					}
				}
			}
		}
	}
}

func (this *BlackListMgr) UnDivBlackInfo(blackInfo *BlackInfo) {
	for i := uint(0); i < BlackState_Max; i++ {
		if blackInfo.BlackType&(1<<i) != 0 {
			blbp := this.BlackListByPlatform[i]
			if blbp != nil {
				if pool, exist := blbp[blackInfo.Platform]; exist {
					delete(pool, blackInfo.Id)
				}
			}

			//alipay account
			if len(blackInfo.Alipay_account) > 0 {
				aabp := this.AlipayAccByPlatform[i]
				if aabp != nil {
					if pool, exist := aabp[blackInfo.Platform]; exist {
						delete(pool, blackInfo.Alipay_account)
					}
				}
			}

			//alipay name
			if len(blackInfo.Alipay_name) > 0 {
				anbp := this.AlipayNameByPlatform[i]
				if anbp != nil {
					if pool, exist := anbp[blackInfo.Platform]; exist {
						delete(pool, blackInfo.Alipay_name)
					}
				}
			}

			//bank
			if len(blackInfo.Bankcard) > 0 {
				bankbp := this.BankcardByPlatform[i]
				if bankbp != nil {
					if pool, exist := bankbp[blackInfo.Platform]; exist {
						delete(pool, blackInfo.Bankcard)
					}
				}
			}

			//ip
			if len(blackInfo.Ip) > 0 {
				ipbp := this.IpByPlatform[i]
				if ipbp != nil {
					if pool, exist := ipbp[blackInfo.Platform]; exist {
						delete(pool, blackInfo.Ip)
					}
				}
			}

			//ipnet
			if blackInfo.ipNet != nil {
				ipbp := this.IpNetByPlatform[i]
				if ipbp != nil {
					if pool, exist := ipbp[blackInfo.Platform]; exist {
						index := -1
						for i, bi := range pool {
							if bi.ipNet == blackInfo.ipNet {
								index = i
								break
							}
						}
						if index != -1 {
							pool = append(pool[:index], pool[index+1:]...)
						}
						ipbp[blackInfo.Platform] = pool
					}
				}
			}

			//packageid
			if len(blackInfo.PackageTag) > 0 {
				packbp := this.PackageTagByPlatform[i]
				if packbp != nil {
					delete(packbp, blackInfo.PackageTag)
				}
			}

			//deviceinfo
			if len(blackInfo.DeviceId) > 0 {
				ipbp := this.DeviceByPlatform[i]
				if ipbp != nil {
					if pool, exist := ipbp[blackInfo.Platform]; exist {
						delete(pool, blackInfo.DeviceId)
					}
				}
			}
		}
	}
}

// 初始化
func (this *BlackListMgr) InitBlackInfo(bia *BlackInfoApi) {
	blackInfo := &BlackInfo{
		Id:             bia.Id,
		BlackType:      int(bia.Space),
		Alipay_account: bia.Alipay_account,
		Alipay_name:    bia.Alipay_name,
		Bankcard:       bia.Bankcard,
		Ip:             bia.Ip,
		Platform:       bia.Platform,
		PackageTag:     bia.PackageTag,
		DeviceId:       bia.DeviceId,
	}
	if strings.Contains(blackInfo.Ip, "/") {
		_, blackInfo.ipNet, _ = net.ParseCIDR(blackInfo.Ip)
	}
	this.BlackList[blackInfo.Id] = blackInfo
	this.DivBlackInfo(blackInfo)
}
func (this *BlackListMgr) UpsertBlackInfo(bia *BlackInfoApi) {
	blackInfo := &BlackInfo{
		Id:             bia.Id,
		BlackType:      int(bia.Space),
		Alipay_account: bia.Alipay_account,
		Alipay_name:    bia.Alipay_name,
		Bankcard:       bia.Bankcard,
		Ip:             bia.Ip,
		Platform:       bia.Platform,
		PackageTag:     bia.PackageTag,
		DeviceId:       bia.DeviceId,
	}
	if strings.Contains(blackInfo.Ip, "/") {
		_, blackInfo.ipNet, _ = net.ParseCIDR(blackInfo.Ip)
	}
	if old, exist := this.BlackList[blackInfo.Id]; exist {
		this.UnDivBlackInfo(old)
	}
	this.BlackList[blackInfo.Id] = blackInfo
	this.DivBlackInfo(blackInfo)
}

func (this *BlackListMgr) RemoveBlackInfo(Id int32, platform string) {
	if org, exist := this.BlackList[Id]; exist {
		if len(platform) == 0 || org.Platform == platform {
			logger.Logger.Infof("(this *BlackListMgr) RemoveBlackInfo(Id:%v, platform:%v)", Id, platform)
			delete(this.BlackList, Id)
			this.UnDivBlackInfo(org)
		}
	}
}

//	func (this *BlackListMgr) RemoveBlackInfoByUser(blackinfo *BlackInfo) {
//		this.OnRemoveBlackInfo(blackinfo)
//	}
//
//	func (this *BlackListMgr) OnAddBlackInfo(blackinfo *BlackInfo) {
//		for _, ob := range this.Observers {
//			ob.OnAddBlackInfo(blackinfo)
//		}
//	}
//
//	func (this *BlackListMgr) OnEditBlackInfo(blackinfo *BlackInfo) {
//		for _, ob := range this.Observers {
//			ob.OnEditBlackInfo(blackinfo)
//		}
//	}
//
//	func (this *BlackListMgr) OnRemoveBlackInfo(blackinfo *BlackInfo) {
//		for _, ob := range this.Observers {
//			ob.OnRemoveBlackInfo(blackinfo)
//		}
//	}
func (this *BlackListMgr) CheckPlayerInvalid(blackinfo *BlackInfo, data *model.PlayerData) bool {
	if len(blackinfo.Platform) > 0 && (data.Platform != blackinfo.Platform && blackinfo.Platform != "0") {
		return false
	}
	switch {
	//case blackinfo.Snid == data.SnId: //根据用户的ID找到了对应的用户
	//	return true
	case len(blackinfo.Ip) > 0 && blackinfo.Ip == data.Ip: //根据用户的IP找到了对应的用户
		return true
	case len(blackinfo.Alipay_name) > 0 && blackinfo.Alipay_name == data.AlipayAccName: //根据用户的支付宝找到了对应的用户
		return true
	case len(blackinfo.Alipay_account) > 0 && blackinfo.Alipay_account == data.AlipayAccount: //根据用户的支付宝账户找到了对应的用户
		return true
	case len(blackinfo.Bankcard) > 0 && blackinfo.Bankcard == data.BankAccount: //根据用户的银行卡找到了对应的用户
		return true
	case len(blackinfo.PackageTag) > 0 && strings.HasPrefix(data.PackageID, blackinfo.PackageTag): //根据包标识找对应的用户
		return true
	}
	if blackinfo.ipNet != nil {
		ip := net.ParseIP(data.Ip)
		if ip != nil {
			if blackinfo.ipNet.Contains(ip) {
				return true
			}
		}
	}
	return false
}
func (this *BlackListMgr) CheckLogin(data *model.PlayerData) (*BlackInfo, bool) {
	if bi, ok := this.CheckPlayerInBlack(data, BlackState_Login); ok {
		return bi, false
	}
	return nil, true
}
func (this *BlackListMgr) CheckExchange(data *model.PlayerData) (*BlackInfo, bool) {
	if bi, ok := this.CheckPlayerInBlack(data, BlackState_Exchange); ok {
		return bi, false
	}
	return nil, true
}
func (this *BlackListMgr) CheckRecharge(data *model.PlayerData) (*BlackInfo, bool) {
	if bi, ok := this.CheckPlayerInBlack(data, BlackState_Recharge); ok {
		return bi, false
	}
	return nil, true
}
func (this *BlackListMgr) CheckMatch(data *model.PlayerData) (*BlackInfo, bool) {
	if bi, ok := this.CheckPlayerInBlack(data, BlackState_Match); ok {
		return bi, false
	}
	return nil, true
}
func (this *BlackListMgr) CheckPlayerInBlack(data *model.PlayerData, blackType uint) (bi *BlackInfo, ok bool) {
	if data.IsRob {
		return nil, false
	}
	if bi, ok := this.CheckPlayerInBlackByPlatfromUsePlayerBind(data, blackType, data.Platform); ok {
		return bi, true
	}
	return nil, false
}

func (this *BlackListMgr) CheckPlayerInBlackByPlatfromUsePlayerBind(data *model.PlayerData, blackType uint, platform string) (*BlackInfo, bool) {
	if blackType >= 0 && blackType < BlackState_Max {
		//先检查snid
		if uint(data.BlacklistType)&(blackType+1) == uint(blackType+1) {
			logger.Logger.Infof("found platform:%v player:%d snid in blacklist", platform, data.SnId)
			return nil, true
		}

		//支付宝账号
		if len(data.AlipayAccount) > 0 {
			aabp := this.AlipayAccByPlatform[blackType]
			if aabp != nil {
				if pool, exist := aabp[platform]; exist {
					if bi, exist := pool[data.AlipayAccount]; exist {
						logger.Logger.Infof("found platform:%v player:%d AlipayAccount:%v in blacklist", platform, data.SnId, data.AlipayAccount)
						return bi, true
					}
				}
			}
		}

		//支付宝用户名
		if len(data.AlipayAccName) > 0 {
			aabp := this.AlipayNameByPlatform[blackType]
			if aabp != nil {
				if pool, exist := aabp[platform]; exist {
					if bi, exist := pool[data.AlipayAccName]; exist {
						logger.Logger.Infof("found platform:%v player:%d AlipayAccName:%v in blacklist", platform, data.SnId, data.AlipayAccName)
						return bi, true
					}
				}
			}
		}

		//银行卡号
		if len(data.BankAccount) > 0 {
			bankbp := this.BankcardByPlatform[blackType]
			if bankbp != nil {
				if pool, exist := bankbp[platform]; exist {
					if bi, exist := pool[data.BankAccount]; exist {
						logger.Logger.Infof("found platform:%v player:%d BankAccount:%v in blacklist", platform, data.SnId, data.BankAccount)
						return bi, true
					}
				}
			}
		}

		//ip检查
		if len(data.Ip) > 0 {
			//精准ip
			ipbp := this.IpByPlatform[blackType]
			if ipbp != nil {
				if pool, exist := ipbp[platform]; exist {
					if bi, exist := pool[data.Ip]; exist {
						logger.Logger.Infof("found platform:%v player:%d Ip:%v in blacklist", platform, data.SnId, data.Ip)
						return bi, true
					}
				}
			}

			//ip段
			ipnetbp := this.IpNetByPlatform[blackType]
			if ipnetbp != nil {
				if pool, exist := ipnetbp[platform]; exist {
					for _, bi := range pool {
						ip := net.ParseIP(data.Ip)
						if ip != nil {
							if bi.ipNet.Contains(ip) {
								logger.Logger.Infof("found platform:%v player:%d Ip:%v in blacklist ipnet:%v", platform, data.SnId, data.Ip, bi.ipNet.String())
								return bi, true
							}
						}
					}
				}
			}
		}

		//包标识
		if len(data.PackageID) > 0 {
			packbp := this.PackageTagByPlatform[blackType]
			if packbp != nil {
				if bi, exist := packbp[data.PackageID]; exist {
					logger.Logger.Infof("found platform:%v player:%d PackageID:%v in blacklist", platform, data.SnId, data.PackageID)
					return bi, true
				}
			}
		}
		//设备号
		if len(data.DeviceId) > 0 {
			deviceByPlatform := this.DeviceByPlatform[blackType]
			if deviceByPlatform != nil {
				if devices, exist := deviceByPlatform[data.Platform]; exist {
					if bi, ok := devices[data.DeviceId]; ok {
						logger.Logger.Infof("found platform:%v player:%d PackageID:%v in blacklist", platform, data.SnId, data.PackageID)
						return bi, true
					}
				}
			}
		}
	}
	return nil, false
}

// 设备号验证
func (this *BlackListMgr) CheckDeviceInBlack(deviceId string, blackType uint, platform string) (*BlackInfo, bool) {
	if bi, ok := this.CheckDeviceInBlackByPlatfrom(deviceId, blackType, platform); ok {
		return bi, true
	}
	return nil, false
}

func (this *BlackListMgr) CheckDeviceInBlackByPlatfrom(deviceId string, blackType uint, platform string) (*BlackInfo, bool) {
	if blackType >= 0 && blackType < BlackState_Max {
		if len(deviceId) > 0 {
			devicebp := this.DeviceByPlatform[blackType]
			if devicebp != nil {
				if pool, exist := devicebp[platform]; exist {
					if bi, ok := pool[deviceId]; ok {
						logger.Logger.Infof("found platform:%v device:%s in blacklist", platform, deviceId)
						return bi, true
					}
				}
			}
		}
	}
	return nil, false
}

func init() {
	mgo.SetStats(true)
	RegisteParallelLoadFunc("平台黑名单", func() error {
		BlackListMgrSington.Init()
		return nil
	})
}
