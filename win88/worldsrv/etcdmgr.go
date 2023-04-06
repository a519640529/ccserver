package main

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"games.yol.com/win88/common"
	"games.yol.com/win88/etcd"
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	hall_proto "games.yol.com/win88/protocol/gamehall"
	login_proto "games.yol.com/win88/protocol/login"
	webapi_proto "games.yol.com/win88/protocol/webapi"
	"github.com/idealeak/goserver/core/logger"
	"go.etcd.io/etcd/clientv3"
)

var EtcdMgrSington = &EtcdMgr{
	EtcdClient: &etcd.EtcdClient{},
}

type EtcdMgr struct {
	*etcd.EtcdClient
}

// 初始化平台数据
func (this *EtcdMgr) InitPlatform() {
	initFunc := func() int64 {
		if model.GameParamData.UseEtcd {
			logger.Logger.Info("ETCD 拉取数据:", etcd.ETCDKEY_PLATFORM_PREFIX)
			res, err := this.GetValueWithPrefix(etcd.ETCDKEY_PLATFORM_PREFIX)
			if err == nil {
				for i := int64(0); i < res.Count; i++ {
					var value webapi_proto.Platform
					err = proto.Unmarshal(res.Kvs[i].Value, &value)
					if err == nil {
						PlatformMgrSington.UpsertPlatform(value.PlatformName, value.Isolated, value.Disabled, value.Id,
							value.CustomService, value.BindOption, value.ServiceFlag, value.UpgradeAccountGiveCoin,
							value.NewAccountGiveCoin, value.PerBankNoLimitAccount, value.ExchangeMin, value.ExchangeLimit,
							value.ExchangeTax, value.ExchangeFlow, value.ExchangeFlag, value.SpreadConfig, value.VipRange, "",
							nil, value.VerifyCodeType, nil /*value.ThirdGameMerchant,*/, value.CustomType,
							false, value.NeedSameName, value.ExchangeForceTax, value.ExchangeGiveFlow, value.ExchangeVer,
							value.ExchangeBankMax, value.ExchangeAlipayMax, 0, value.PerBankNoLimitName, value.IsCanUserBindPromoter,
							value.UserBindPromoterPrize, false, value.ExchangeMultiple, false, value.MerchantKey)
					} else {
						logger.Logger.Errorf("etcd read(%v) proto.Unmarshal err:%v", string(res.Kvs[i].Key), err)
					}
				}
				if res.Header != nil {
					return res.Header.Revision
				}
			} else {
				logger.Logger.Errorf("etcd get WithPrefix(%v) panic:%v", etcd.ETCDKEY_PLATFORM_PREFIX, err)
			}
		}
		return -1
	}

	watchFunc := func(ctx context.Context, revision int64) {
		// 监控数据变动
		this.GoWatch(ctx, revision, etcd.ETCDKEY_PLATFORM_PREFIX, func(res clientv3.WatchResponse) error {
			for _, ev := range res.Events {
				switch ev.Type {
				case clientv3.EventTypeDelete:
				case clientv3.EventTypePut:
					var value webapi_proto.Platform
					err := proto.Unmarshal(ev.Kv.Value, &value)
					if err == nil {
						platform := PlatformMgrSington.UpsertPlatform(value.PlatformName, value.Isolated,
							value.Disabled, value.Id, value.CustomService, value.BindOption,
							value.ServiceFlag, value.UpgradeAccountGiveCoin, value.NewAccountGiveCoin,
							value.PerBankNoLimitAccount, value.ExchangeMin, value.ExchangeLimit,
							value.ExchangeTax, value.ExchangeFlow, value.ExchangeFlag, value.SpreadConfig,
							value.VipRange, "", nil, value.VerifyCodeType,
							nil, value.CustomType, true, value.NeedSameName,
							value.ExchangeForceTax, value.ExchangeGiveFlow, value.ExchangeVer,
							value.ExchangeBankMax, value.ExchangeAlipayMax, 0, value.PerBankNoLimitName,
							value.IsCanUserBindPromoter, value.UserBindPromoterPrize, false, value.ExchangeMultiple,
							false, value.MerchantKey)
						if platform != nil {
							//通知客户端平台配置发生改变
							scPlatForm := &login_proto.SCPlatFormConfig{
								Platform:               proto.String(platform.IdStr),
								OpRetCode:              login_proto.OpResultCode_OPRC_Sucess,
								UpgradeAccountGiveCoin: proto.Int32(platform.UpgradeAccountGiveCoin),
								ExchangeMin:            proto.Int32(platform.ExchangeMin),
								ExchangeLimit:          proto.Int32(platform.ExchangeLimit),
								VipRange:               platform.VipRange,
								OtherParams:            proto.String(platform.OtherParams),
								SpreadConfig:           proto.Int32(platform.SpreadConfig),
								ExchangeTax:            proto.Int32(platform.ExchangeTax),
								ExchangeFlow:           proto.Int32(platform.ExchangeFlow),
								ExchangeBankMax:        proto.Int32(platform.ExchangeBankMax),
								ExchangeAlipayMax:      proto.Int32(platform.ExchangeAlipayMax),
								ExchangeMultiple:       proto.Int32(platform.ExchangeMultiple),
							}
							rebateTask := RebateInfoMgrSington.rebateTask[platform.IdStr]
							if rebateTask != nil {
								scPlatForm.Rebate = &login_proto.RebateCfg{
									RebateSwitch:   proto.Bool(rebateTask.RebateSwitch),
									ReceiveMode:    proto.Int32(int32(rebateTask.ReceiveMode)),
									NotGiveOverdue: proto.Int32(int32(rebateTask.NotGiveOverdue)),
								}
							}
							if platform.ClubConfig != nil { //俱乐部配置
								scPlatForm.Club = &login_proto.ClubCfg{
									IsOpenClub:              proto.Bool(platform.ClubConfig.IsOpenClub),
									CreationCoin:            proto.Int64(platform.ClubConfig.CreationCoin),
									IncreaseCoin:            proto.Int64(platform.ClubConfig.IncreaseCoin),
									ClubInitPlayerNum:       proto.Int32(platform.ClubConfig.ClubInitPlayerNum),
									CreateClubCheckByManual: proto.Bool(platform.ClubConfig.CreateClubCheckByManual),
									EditClubNoticeByManual:  proto.Bool(platform.ClubConfig.EditClubNoticeByManual),
									CreateRoomAmount:        proto.Int64(platform.ClubConfig.CreateRoomAmount),
									GiveCoinRate:            platform.ClubConfig.GiveCoinRate,
								}
							}
							proto.SetDefaults(scPlatForm)
							PlayerMgrSington.BroadcastMessageToPlatform(platform.IdStr, int(login_proto.LoginPacketID_PACKET_SC_PLATFORMCFG), scPlatForm)
						}
					} else {
						logger.Logger.Errorf("etcd read(%v) proto.Unmarshal err:%v", string(ev.Kv.Key), err)
					}
				}
			}
			return nil
		})
	}

	this.InitAndWatch(initFunc, watchFunc)
}

// 初始化平台游戏配置
func (this *EtcdMgr) InitPlatformGameConfig() {
	initFunc := func() int64 {
		if model.GameParamData.UseEtcd {
			logger.Logger.Info("ETCD 拉取数据:", etcd.ETCDKEY_GAMECONFIG_PREFIX)
			res, err := this.GetValueWithPrefix(etcd.ETCDKEY_GAMECONFIG_PREFIX)
			if err == nil {
				for i := int64(0); i < res.Count; i++ {
					var config webapi_proto.GameFree
					err = proto.Unmarshal(res.Kvs[i].Value, &config)
					if err == nil {
						s := strings.TrimPrefix(string(res.Kvs[i].Key), etcd.ETCDKEY_GAMECONFIG_PREFIX)
						arr := strings.Split(s, "/")
						if len(arr) > 1 {
							pltId := arr[0]
							if err == nil {
								PlatformMgrSington.UpsertPlatformGameConfig(pltId, &config)
							}
						}
					} else {
						logger.Logger.Errorf("etcd read(%v) proto.Unmarshal err:%v", string(res.Kvs[i].Key), err)
					}
				}
				if res.Header != nil {
					return res.Header.Revision
				}
			} else {
				logger.Logger.Errorf("etcd get WithPrefix(%v) panic:%v", etcd.ETCDKEY_GAMECONFIG_PREFIX, err)
			}
		}
		return -1
	}

	watchFunc := func(ctx context.Context, revision int64) {
		// 监控数据变动
		this.GoWatch(ctx, revision, etcd.ETCDKEY_GAMECONFIG_PREFIX, func(res clientv3.WatchResponse) error {
			for _, ev := range res.Events {
				switch ev.Type {
				case clientv3.EventTypeDelete:
				case clientv3.EventTypePut:
					var config webapi_proto.GameFree
					err := proto.Unmarshal(ev.Kv.Value, &config)
					if err == nil {
						s := strings.TrimPrefix(string(ev.Kv.Key), etcd.ETCDKEY_GAMECONFIG_PREFIX)
						arr := strings.Split(s, "/")
						if len(arr) > 1 {
							pltId := arr[0]
							if err == nil {
								PlatformMgrSington.UpsertPlatformGameConfig(pltId, &config)
							}
						}
					} else {
						logger.Logger.Errorf("etcd read(%v) proto.Unmarshal err:%v", string(ev.Kv.Key), err)
					}
				}
			}
			return nil
		})
	}

	this.InitAndWatch(initFunc, watchFunc)
}

// 初始化平台包数据
func (this *EtcdMgr) InitPlatformPackage() {
	initFunc := func() int64 {
		if model.GameParamData.UseEtcd {
			logger.Logger.Info("ETCD 拉取数据:", etcd.ETCDKEY_PACKAGE_PREFIX)
			res, err := this.GetValueWithPrefix(etcd.ETCDKEY_PACKAGE_PREFIX)
			if err == nil {
				for i := int64(0); i < res.Count; i++ {
					var value webapi_proto.AppInfo
					err = proto.Unmarshal(res.Kvs[i].Value, &value)
					if err == nil {
						if value.PlatformId == 0 {
							value.PlatformId = int32(Default_PlatformInt)
						}

						PlatformMgrSington.PackageList[value.PackageName] = &value
						PlatformMgrSington.PackageList[value.BundleId] = &value
						//if _, ok := PlatformMgrSington.PromoterList[strconv.Itoa(int(value.PromoterId))]; !ok {
						//	PlatformMgrSington.PromoterList[strconv.Itoa(int(value.PromoterId))] = PlatformPromoter{
						//		Platform: strconv.Itoa(int(value.Platform)),
						//		Promoter: strconv.Itoa(int(value.PromoterId)),
						//		Tag:      value.Tag,
						//	}
						//}
					} else {
						logger.Logger.Errorf("etcd read(%v) proto.Unmarshal err:%v", string(res.Kvs[i].Key), err)
					}
				}
				if res.Header != nil {
					return res.Header.Revision
				}
			} else {
				logger.Logger.Errorf("etcd get WithPrefix(%v) panic:%v", etcd.ETCDKEY_PACKAGE_PREFIX, err)
			}
		}
		return -1
	}

	watchFunc := func(ctx context.Context, revision int64) {
		// 监控数据变动
		this.GoWatch(ctx, revision, etcd.ETCDKEY_PACKAGE_PREFIX, func(res clientv3.WatchResponse) error {
			for _, ev := range res.Events {
				switch ev.Type {
				case clientv3.EventTypeDelete:
				case clientv3.EventTypePut:
					var value webapi_proto.AppInfo
					err := proto.Unmarshal(ev.Kv.Value, &value)
					if err == nil {
						if value.PlatformId == 0 {
							value.PlatformId = int32(Default_PlatformInt)
						}

						PlatformMgrSington.PackageList[value.PackageName] = &value
						PlatformMgrSington.PackageList[value.BundleId] = &value
						//if _, ok := PlatformMgrSington.PromoterList[strconv.Itoa(int(value.PromoterId))]; !ok {
						//	PlatformMgrSington.PromoterList[strconv.Itoa(int(value.PromoterId))] = PlatformPromoter{
						//		Platform: strconv.Itoa(int(value.Platform)),
						//		Promoter: strconv.Itoa(int(value.PromoterId)),
						//		Tag:      value.Tag,
						//	}
						//}
						//
						//var strPlatorm string
						//var strChannel string
						//var strPromoter string
						//if value.Platform != 0 {
						//	strPlatorm = strconv.Itoa(int(value.Platform))
						//}
						//if value.ChannelId != 0 {
						//	strChannel = strconv.Itoa(int(value.ChannelId))
						//}
						//if value.PromoterId != 0 {
						//	strPromoter = strconv.Itoa(int(value.PromoterId))
						//}
						//if value.AppStore != 1 {
						//	uptCnt := PlayerMgrSington.UpdateAllPlayerPackageTag(value.Tag, strPlatorm, strChannel, strPromoter, int32(value.PromoterTree), value.TagKey)
						//	logger.Logger.Infof("/api/Game/UpsertPackageTag update(tag:%v, platform:%v, channel:%v, promoter:%v), updated cnt=%v", value.Tag, strPlatorm, strChannel, strPromoter, uptCnt)
						//	task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
						//		return model.UpdateAllPlayerPackageTag(value.Tag, strPlatorm, strChannel, strPromoter, int32(value.PromoterTree), value.TagKey)
						//	}), nil, "UpdateAllPlayerPackageTag").Start()
						//}
					} else {
						logger.Logger.Errorf("etcd desc WithPrefix(%v) panic:%v", etcd.ETCDKEY_PACKAGE_PREFIX, err)
					}
				}
			}
			return nil
		})
	}
	this.InitAndWatch(initFunc, watchFunc)
}

// 加载组配置
func (this *EtcdMgr) InitGameGroup() {
	initFunc := func() int64 {
		if model.GameParamData.UseEtcd {
			logger.Logger.Info("ETCD 拉取数据:", etcd.ETCDKEY_GROUPCONFIG_PREFIX)
			res, err := this.GetValueWithPrefix(etcd.ETCDKEY_GROUPCONFIG_PREFIX)
			if err == nil {
				for i := int64(0); i < res.Count; i++ {
					var value webapi_proto.GameConfigGroup
					err = proto.Unmarshal(res.Kvs[i].Value, &value)
					if err == nil {
						PlatformGameGroupMgrSington.UpsertGameGroup(&value)
					} else {
						logger.Logger.Errorf("etcd desc WithPrefix(%v) panic:%v", etcd.ETCDKEY_GROUPCONFIG_PREFIX, err)
					}
				}
				if res.Header != nil {
					return res.Header.Revision
				}
			} else {
				logger.Logger.Errorf("etcd get WithPrefix(%v) panic:%v", etcd.ETCDKEY_GROUPCONFIG_PREFIX, err)
			}
		}
		return -1
	}

	// 监控数据变动
	watchFunc := func(ctx context.Context, revision int64) {
		this.GoWatch(ctx, revision, etcd.ETCDKEY_GROUPCONFIG_PREFIX, func(res clientv3.WatchResponse) error {
			for _, ev := range res.Events {
				switch ev.Type {
				case clientv3.EventTypeDelete:
				case clientv3.EventTypePut:
					var value webapi_proto.GameConfigGroup
					err := proto.Unmarshal(ev.Kv.Value, &value)
					if err == nil {
						PlatformGameGroupMgrSington.UpsertGameGroup(&value)
					} else {
						logger.Logger.Errorf("etcd desc WithPrefix(%v) panic:%v", etcd.ETCDKEY_GROUPCONFIG_PREFIX, err)
					}
				}
			}
			return nil
		})
	}

	this.InitAndWatch(initFunc, watchFunc)
}

// 加载黑名单配置
func (this *EtcdMgr) InitBlackList() {
	initFunc := func() int64 {
		if model.GameParamData.UseEtcd {
			logger.Logger.Info("ETCD 拉取数据:", etcd.ETCDKEY_BLACKLIST_PREFIX)
			res, err := this.GetValueWithPrefix(etcd.ETCDKEY_BLACKLIST_PREFIX)
			if err == nil {
				for i := int64(0); i < res.Count; i++ {
					var value BlackInfoApi
					err = json.Unmarshal(res.Kvs[i].Value, &value)
					if err == nil {
						BlackListMgrSington.InitBlackInfo(&value)
					} else {
						logger.Logger.Errorf("etcd desc WithPrefix(%v) panic:%v", etcd.ETCDKEY_BLACKLIST_PREFIX, err)
					}
				}
				if res.Header != nil {
					return res.Header.Revision
				}
			} else {
				logger.Logger.Errorf("etcd get WithPrefix(%v) panic:%v", etcd.ETCDKEY_BLACKLIST_PREFIX, err)
			}
		}
		return -1
	}

	//@test code
	//go func() {
	//	for {
	//		i := int32(1)
	//		data := BlackInfoApi{
	//			Id:      i,
	//			Snid:    i,
	//			Creator: rand.Int31(),
	//		}
	//		buf, err := json.Marshal(data)
	//		if err == nil {
	//			key := fmt.Sprintf("%s%d", etcd.ETCDKEY_BLACKLIST_PREFIX, i)
	//			putResp, err := this.PutValue(key, string(buf))
	//			if err == nil {
	//				if putResp.PrevKv != nil {
	//					logger.Logger.Trace("@etcdtest put", string(putResp.PrevKv.Key), string(putResp.PrevKv.Value))
	//				}
	//				//delResp, err := this.DelValue(key)
	//				//if err == nil {
	//				//	logger.Logger.Trace("@etcdtest del", delResp.Deleted)
	//				//}
	//			}
	//		}
	//	}
	//}()
	//@test code

	//ETCD中现在只有公共黑名单信息
	//如果删除公共黑名单信息使用ETCD删除
	//如果删除个人玩家身上的黑名单信息使用API删除
	// 监控数据变动
	watchFunc := func(ctx context.Context, revision int64) {
		this.GoWatch(ctx, revision, etcd.ETCDKEY_BLACKLIST_PREFIX, func(res clientv3.WatchResponse) error {
			for _, ev := range res.Events {
				switch ev.Type {
				case clientv3.EventTypeDelete:
					dirs := strings.Split(string(ev.Kv.Key), "/")
					n := len(dirs)
					if n > 0 {
						last := dirs[n-1]
						id, err := strconv.Atoi(last)
						if err == nil {
							if value, exist := BlackListMgrSington.BlackList[int32(id)]; exist {
								BlackListMgrSington.RemoveBlackInfo(value.Id, value.Platform)
							}
						}
					}
				case clientv3.EventTypePut:
					var value BlackInfoApi
					err := json.Unmarshal(ev.Kv.Value, &value)
					if err == nil {
						BlackListMgrSington.UpsertBlackInfo(&value)
						if (value.Space & (1 << BlackState_Login)) != 0 {
							var targetPlayer []*Player //确定用户是否在线
							for _, value := range PlayerMgrSington.players {
								_, ok := BlackListMgrSington.CheckPlayerInBlack(value.PlayerData, BlackState_Login)
								if ok {
									targetPlayer = append(targetPlayer, value)
								}
							}
							for _, p := range targetPlayer {
								if p.sid != 0 {
									p.Kickout(int32(login_proto.SSDisconnectTypeCode_SSDTC_BlackList))
								} else {
									LoginStateMgrSington.LogoutByAccount(p.AccountId)
								}
							}
						}
					} else {
						logger.Logger.Errorf("etcd desc WithPrefix(%v) panic:%v", etcd.ETCDKEY_BLACKLIST_PREFIX, err)
					}
				}
			}
			return nil
		})
	}

	this.InitAndWatch(initFunc, watchFunc)
}

// 加载平台公告配置
func (this *EtcdMgr) InitPlatformBulletin() {
	initFunc := func() int64 {
		if model.GameParamData.UseEtcd {
			logger.Logger.Info("ETCD 拉取数据:", etcd.ETCDKEY_BULLETIN_PREFIX)
			res, err := this.GetValueWithPrefix(etcd.ETCDKEY_BULLETIN_PREFIX)
			if err == nil {
				for i := int64(0); i < res.Count; i++ {
					var info Bullet
					err = json.Unmarshal(res.Kvs[i].Value, &info)
					if err == nil {
						BulletMgrSington.BulletMsgList[info.Id] = &info
					} else {
						logger.Logger.Errorf("etcd desc WithPrefix(%v) panic:%v", etcd.ETCDKEY_BULLETIN_PREFIX, err)
					}
				}
				if res.Header != nil {
					return res.Header.Revision
				}
			} else {
				logger.Logger.Errorf("etcd get WithPrefix(%v) panic:%v", etcd.ETCDKEY_BULLETIN_PREFIX, err)
			}
		}
		return -1
	}

	// 监控数据变动
	watchFunc := func(ctx context.Context, revision int64) {
		this.GoWatch(ctx, revision, etcd.ETCDKEY_BULLETIN_PREFIX, func(res clientv3.WatchResponse) error {
			for _, ev := range res.Events {
				switch ev.Type {
				case clientv3.EventTypeDelete:
					dirs := strings.Split(string(ev.Kv.Key), "/")
					n := len(dirs)
					if n > 0 {
						last := dirs[n-1]
						id, err := strconv.Atoi(last)
						if err == nil {
							delete(BulletMgrSington.BulletMsgList, int32(id))
						}
					}
				case clientv3.EventTypePut:
					var value Bullet
					err := json.Unmarshal(ev.Kv.Value, &value)
					if err == nil {
						BulletMgrSington.BulletMsgList[value.Id] = &value
						var bulletList []*login_proto.Bulletion
						bulletList = append(bulletList, &login_proto.Bulletion{
							Id:            proto.Int32(value.Id),
							NoticeTitle:   proto.String(value.NoticeTitle),
							NoticeContent: proto.String(value.NoticeContent),
							UpdateTime:    proto.String(value.UpdateTime),
							Sort:          proto.Int32(value.Sort),
						})
						var rawpack = &login_proto.SCBulletionInfo{
							Id:            proto.Int32(value.Id),
							BulletionList: bulletList,
						}
						proto.SetDefaults(rawpack)
						PlayerMgrSington.BroadcastMessageToPlatform(value.Platform, int(login_proto.LoginPacketID_PACKET_SC_BULLETIONINFO), rawpack)
					} else {
						logger.Logger.Errorf("etcd desc WithPrefix(%v) panic:%v", etcd.ETCDKEY_BULLETIN_PREFIX, err)
					}
				}
			}
			return nil
		})
	}

	this.InitAndWatch(initFunc, watchFunc)
}

// 加载平台招商代理配置
func (this *EtcdMgr) InitPlatformAgent() {
	initFunc := func() int64 {
		if model.GameParamData.UseEtcd {
			logger.Logger.Info("ETCD 拉取数据:", etcd.ETCDKEY_AGENTCUSTOMER_PREFIX)
			res, err := this.GetValueWithPrefix(etcd.ETCDKEY_AGENTCUSTOMER_PREFIX)
			if err == nil {
				for i := int64(0); i < res.Count; i++ {
					var info Customer
					err = json.Unmarshal(res.Kvs[i].Value, &info)
					if err == nil {
						CustomerMgrSington.CustomerMsgList[info.Id] = &info
					} else {
						logger.Logger.Errorf("etcd desc WithPrefix(%v) panic:%v", etcd.ETCDKEY_AGENTCUSTOMER_PREFIX, err)
					}
				}
				if res.Header != nil {
					return res.Header.Revision
				}
			} else {
				logger.Logger.Errorf("etcd get WithPrefix(%v) panic:%v", etcd.ETCDKEY_AGENTCUSTOMER_PREFIX, err)
			}
		}
		return -1
	}

	// 监控数据变动
	watchFunc := func(ctx context.Context, revision int64) {
		this.GoWatch(ctx, revision, etcd.ETCDKEY_AGENTCUSTOMER_PREFIX, func(res clientv3.WatchResponse) error {
			for _, ev := range res.Events {
				switch ev.Type {
				case clientv3.EventTypeDelete:
					dirs := strings.Split(string(ev.Kv.Key), "/")
					n := len(dirs)
					if n > 0 {
						last := dirs[n-1]
						id, err := strconv.Atoi(last)
						if err == nil {
							delete(CustomerMgrSington.CustomerMsgList, int32(id))
						}
					}
				case clientv3.EventTypePut:
					var value Customer
					err := json.Unmarshal(ev.Kv.Value, &value)
					if err == nil {
						CustomerMgrSington.CustomerMsgList[value.Id] = &value
					} else {
						logger.Logger.Errorf("etcd desc WithPrefix(%v) panic:%v", etcd.ETCDKEY_AGENTCUSTOMER_PREFIX, err)
					}
				}
			}
			return nil
		})
	}

	this.InitAndWatch(initFunc, watchFunc)
}

// 加载平台签到活动配置
func (this *EtcdMgr) InitPlatformActSignin() {
	initFunc := func() int64 {
		if model.GameParamData.UseEtcd {
			logger.Logger.Info("ETCD 拉取数据:", etcd.ETCDKEY_ACT_SIGNIN_PREFIX)
			//res, err := this.GetValueWithPrefix(etcd.ETCDKEY_ACT_SIGNIN_PREFIX)
			//if err == nil {
			//	for i := int64(0); i < res.Count; i++ {
			//		var value SignConfig
			//		err = json.Unmarshal(res.Kvs[i].Value, &value)
			//		if err == nil {
			//			if value.StartAct > 0 {
			//				ActSignMgrSington.SignConfigs[value.Platform] = value
			//				//logger.Logger.Trace("value:", value)
			//			}
			//		} else {
			//			logger.Logger.Errorf("etcd desc WithPrefix(%v) panic:%v", etcd.ETCDKEY_ACT_SIGNIN_PREFIX, err)
			//		}
			//	}
			//} else {
			//	logger.Logger.Errorf("etcd get WithPrefix(%v) panic:%v", etcd.ETCDKEY_ACT_SIGNIN_PREFIX, err)
			//}
		}
		return -1
	}

	// 监控数据变动
	watchFunc := func(ctx context.Context, revision int64) {
		this.GoWatch(ctx, revision, etcd.ETCDKEY_ACT_SIGNIN_PREFIX, func(res clientv3.WatchResponse) error {
			for _, ev := range res.Events {
				switch ev.Type {
				case clientv3.EventTypeDelete:
					dirs := strings.Split(string(ev.Kv.Key), "/")
					n := len(dirs)
					if n > 0 {
						//platform := dirs[n-1]
						//delete(ActSignMgrSington.SignConfigs, platform)
					}
				case clientv3.EventTypePut:
					//var value SignConfig
					//err := json.Unmarshal(ev.Kv.Value, &value)
					//if err == nil {
					//	ActSignMgrSington.ModifyConfig(value)
					//} else {
					//	logger.Logger.Errorf("etcd desc WithPrefix(%v) panic:%v", etcd.ETCDKEY_ACT_SIGNIN_PREFIX, err)
					//}
				}
			}
			return nil
		})
	}
	this.InitAndWatch(initFunc, watchFunc)
}

// 加载平台任务活动配置
func (this *EtcdMgr) InitPlatformActTask() {
	initFunc := func() int64 {
		if model.GameParamData.UseEtcd {
			logger.Logger.Info("ETCD 拉取数据:", etcd.ETCDKEY_ACT_TASK_PREFIX)
			//res, err := this.GetValueWithPrefix(etcd.ETCDKEY_ACT_TASK_PREFIX)
			//if err == nil {
			//	for i := int64(0); i < res.Count; i++ {
			//		var taskConfig APITaskConfigs
			//		err = json.Unmarshal(res.Kvs[i].Value, &taskConfig)
			//		if err == nil {
			//			gtpf := &PlatformTaskConfig{
			//				ConfigByTaskId: make(map[int32]TaskConfig),
			//				StartAct:       taskConfig.StartAct,
			//				Platform:       taskConfig.Platform,
			//				ConfigVer:      taskConfig.ConfigVer,
			//			}
			//			ActTaskMgrSington.ConfigByPlatform[gtpf.Platform] = gtpf
			//			for _, config := range taskConfig.Datas {
			//				if config.StartAct > 0 {
			//					gtpf.ConfigByTaskId[config.Id] = config
			//				}
			//			}
			//		} else {
			//			logger.Logger.Errorf("etcd desc WithPrefix(%v) panic:%v", etcd.ETCDKEY_ACT_TASK_PREFIX, err)
			//		}
			//	}
			//} else {
			//	logger.Logger.Errorf("etcd get WithPrefix(%v) panic:%v", etcd.ETCDKEY_ACT_TASK_PREFIX, err)
			//}
		}
		return -1
	}

	// 监控数据变动
	watchFunc := func(ctx context.Context, revision int64) {
		this.GoWatch(ctx, revision, etcd.ETCDKEY_ACT_TASK_PREFIX, func(res clientv3.WatchResponse) error {
			for _, ev := range res.Events {
				switch ev.Type {
				case clientv3.EventTypeDelete:
				case clientv3.EventTypePut:
					//var taskConfig APITaskConfigs
					//err := json.Unmarshal(ev.Kv.Value, &taskConfig)
					//if err == nil {
					//	gtpf := PlatformTaskConfig{
					//		ConfigByTaskId: make(map[int32]TaskConfig),
					//		StartAct:       taskConfig.StartAct,
					//		Platform:       taskConfig.Platform,
					//		ConfigVer:      taskConfig.ConfigVer,
					//	}
					//	gtpf.ConfigByTaskId = make(map[int32]TaskConfig)
					//	for _, v := range taskConfig.Datas {
					//		gtpf.ConfigByTaskId[v.Id] = v
					//	}
					//	ActTaskMgrSington.ModifyConfig(gtpf)
					//} else {
					//	logger.Logger.Errorf("etcd desc WithPrefix(%v) panic:%v", etcd.ETCDKEY_ACT_TASK_PREFIX, err)
					//}
				}
			}
			return nil
		})
	}
	this.InitAndWatch(initFunc, watchFunc)
}

// 加载平台财神任务活动配置
func (this *EtcdMgr) InitPlatformActGoldTask() {
	initFunc := func() int64 {
		if model.GameParamData.UseEtcd {
			logger.Logger.Info("ETCD 拉取数据:", etcd.ETCDKEY_ACT_GOLDTASK_PREFIX)
			//res, err := this.GetValueWithPrefix(etcd.ETCDKEY_ACT_GOLDTASK_PREFIX)
			//if err == nil {
			//	for i := int64(0); i < res.Count; i++ {
			//		var taskConfig APIGoldTaskConfigs
			//		err = json.Unmarshal(res.Kvs[i].Value, &taskConfig)
			//		if err == nil {
			//			gtpf := &GoldTaskPlateformConfig{
			//				ConfigByTaskId: make(map[int32]GoldTaskConfig),
			//				StartAct:       taskConfig.StartAct,
			//				Platform:       taskConfig.Platform,
			//				ConfigVer:      taskConfig.ConfigVer,
			//			}
			//			ActGoldTaskMgrSington.ConfigByPlateform[gtpf.Platform] = gtpf
			//			for _, config := range taskConfig.Datas {
			//				if config.StartAct > 0 {
			//					ActGoldTaskMgrSington.AddConfig(config, taskConfig.Platform, false)
			//				}
			//			}
			//		} else {
			//			logger.Logger.Errorf("etcd desc WithPrefix(%v) panic:%v", etcd.ETCDKEY_ACT_GOLDTASK_PREFIX, err)
			//		}
			//	}
			//} else {
			//	logger.Logger.Errorf("etcd get WithPrefix(%v) panic:%v", etcd.ETCDKEY_ACT_GOLDTASK_PREFIX, err)
			//}
		}
		return -1
	}

	// 监控数据变动
	watchFunc := func(ctx context.Context, revision int64) {
		this.GoWatch(ctx, revision, etcd.ETCDKEY_ACT_GOLDTASK_PREFIX, func(res clientv3.WatchResponse) error {
			for _, ev := range res.Events {
				switch ev.Type {
				case clientv3.EventTypeDelete:
				case clientv3.EventTypePut:
					//var taskConfig APIGoldTaskConfigs
					//err := json.Unmarshal(ev.Kv.Value, &taskConfig)
					//if err == nil {
					//	gtpf := GoldTaskPlateformConfig{
					//		ConfigByTaskId: make(map[int32]GoldTaskConfig),
					//		StartAct:       taskConfig.StartAct,
					//		Platform:       taskConfig.Platform,
					//		ConfigVer:      taskConfig.ConfigVer,
					//	}
					//	gtpf.ConfigByTaskId = make(map[int32]GoldTaskConfig)
					//	for _, v := range taskConfig.Datas {
					//		gtpf.ConfigByTaskId[v.TaskId] = v
					//	}
					//	ActGoldTaskMgrSington.ModifyConfig(gtpf)
					//	if curPlateformConfig, exist := ActGoldTaskMgrSington.ConfigByPlateform[gtpf.Platform]; exist {
					//		//移除task
					//		for curTaskId, curTaskConfig := range curPlateformConfig.ConfigByTaskId {
					//			bFindTask := false
					//			if _, exist := gtpf.ConfigByTaskId[curTaskId]; exist {
					//				bFindTask = true
					//			}
					//			if !bFindTask {
					//				ActGoldTaskMgrSington.RemoveConfig(curTaskConfig.TaskId, gtpf.Platform, true)
					//			}
					//		}
					//	}
					//} else {
					//	logger.Logger.Errorf("etcd desc WithPrefix(%v) panic:%v", etcd.ETCDKEY_ACT_GOLDTASK_PREFIX, err)
					//}
				}
			}
			return nil
		})
	}
	this.InitAndWatch(initFunc, watchFunc)
}

// 加载平台财神降临活动配置
func (this *EtcdMgr) InitPlatformActGoldCome() {
	initFunc := func() int64 {
		if model.GameParamData.UseEtcd {
			logger.Logger.Info("ETCD 拉取数据:", etcd.ETCDKEY_ACT_GOLDCOME_PREFIX)
			//res, err := this.GetValueWithPrefix(etcd.ETCDKEY_ACT_GOLDCOME_PREFIX)
			//if err == nil {
			//	for i := int64(0); i < res.Count; i++ {
			//		var taskConfig APIGoldComeConfig
			//		err = json.Unmarshal(res.Kvs[i].Value, &taskConfig)
			//		if err == nil {
			//			gcpf := &GoldComePlateformConfig{
			//				ConfigByTaskId: make(map[int32]GoldComeConfig),
			//				StartAct:       taskConfig.StartAct,
			//				Platform:       taskConfig.Platform,
			//				ConfigVer:      taskConfig.ConfigVer,
			//				RobotType:      taskConfig.RobotType,
			//				MinPlayerCount: taskConfig.MinPlayerCount,
			//			}
			//			ActGoldComeMgrSington.ConfigByPlateform[gcpf.Platform] = gcpf
			//			for _, config := range taskConfig.Datas {
			//				if config.StartAct > 0 {
			//					ActGoldComeMgrSington.AddConfig(config, taskConfig.Platform, false)
			//				}
			//			}
			//		} else {
			//			logger.Logger.Errorf("etcd desc WithPrefix(%v) panic:%v", etcd.ETCDKEY_ACT_GOLDCOME_PREFIX, err)
			//		}
			//	}
			//} else {
			//	logger.Logger.Errorf("etcd get WithPrefix(%v) panic:%v", etcd.ETCDKEY_ACT_GOLDCOME_PREFIX, err)
			//}
		}
		return -1
	}

	// 监控数据变动
	watchFunc := func(ctx context.Context, revision int64) {
		this.GoWatch(ctx, revision, etcd.ETCDKEY_ACT_GOLDCOME_PREFIX, func(res clientv3.WatchResponse) error {
			for _, ev := range res.Events {
				switch ev.Type {
				case clientv3.EventTypeDelete:
				case clientv3.EventTypePut:
					//var taskConfig APIGoldComeConfig
					//err := json.Unmarshal(ev.Kv.Value, &taskConfig)
					//if err == nil {
					//	//1, 修改成最新配置
					//	var config GoldComePlateformConfig
					//	config.Platform = taskConfig.Platform
					//	config.ConfigVer = taskConfig.ConfigVer
					//	config.StartAct = taskConfig.StartAct
					//	config.RobotType = taskConfig.RobotType
					//	config.MinPlayerCount = taskConfig.MinPlayerCount
					//	config.ConfigByTaskId = make(map[int32]GoldComeConfig)
					//	for _, v := range taskConfig.Datas {
					//		config.ConfigByTaskId[v.TaskId] = v
					//	}
					//	ActGoldComeMgrSington.ModifyConfig(config)
					//
					//	if curPlateformConfig, exist := ActGoldTaskMgrSington.ConfigByPlateform[taskConfig.Platform]; exist {
					//		//移除task
					//		for curTaskId, curTaskConfig := range curPlateformConfig.ConfigByTaskId {
					//			bFindTask := false
					//			if _, exist := config.ConfigByTaskId[curTaskId]; exist {
					//				bFindTask = true
					//			}
					//			if !bFindTask {
					//				ActGoldComeMgrSington.RemoveConfig(curTaskConfig.TaskId, config.Platform, true)
					//			}
					//		}
					//	}
					//} else {
					//	logger.Logger.Errorf("etcd desc WithPrefix(%v) panic:%v", etcd.ETCDKEY_ACT_GOLDCOME_PREFIX, err)
					//}
				}
			}
			return nil
		})
	}
	this.InitAndWatch(initFunc, watchFunc)
}

// 加载平台在线奖励活动配置
func (this *EtcdMgr) InitPlatformActOnlineReward() {
	initFunc := func() int64 {
		if model.GameParamData.UseEtcd {
			logger.Logger.Info("ETCD 拉取数据:", etcd.ETCDKEY_ACT_ONLINEREWARD_PREFIX)
			//res, err := this.GetValueWithPrefix(etcd.ETCDKEY_ACT_ONLINEREWARD_PREFIX)
			//if err == nil {
			//	for i := int64(0); i < res.Count; i++ {
			//		var value PlatformOnlineRewardConfig
			//		err = json.Unmarshal(res.Kvs[i].Value, &value)
			//		if err == nil {
			//			ActOnlineRewardMgrSington.Configs[value.Platform] = value
			//		} else {
			//			logger.Logger.Errorf("etcd desc WithPrefix(%v) panic:%v", etcd.ETCDKEY_ACT_ONLINEREWARD_PREFIX, err)
			//		}
			//	}
			//} else {
			//	logger.Logger.Errorf("etcd get WithPrefix(%v) panic:%v", etcd.ETCDKEY_ACT_ONLINEREWARD_PREFIX, err)
			//}
		}
		return -1
	}

	// 监控数据变动
	watchFunc := func(ctx context.Context, revision int64) {
		this.GoWatch(ctx, revision, etcd.ETCDKEY_ACT_ONLINEREWARD_PREFIX, func(res clientv3.WatchResponse) error {
			for _, ev := range res.Events {
				switch ev.Type {
				case clientv3.EventTypeDelete:
				case clientv3.EventTypePut:
					//var value PlatformOnlineRewardConfig
					//err := json.Unmarshal(ev.Kv.Value, &value)
					//if err == nil {
					//	ActOnlineRewardMgrSington.ModifyConfig(value)
					//} else {
					//	logger.Logger.Errorf("etcd desc WithPrefix(%v) panic:%v", etcd.ETCDKEY_ACT_ONLINEREWARD_PREFIX, err)
					//}
				}
			}
			return nil
		})
	}
	this.InitAndWatch(initFunc, watchFunc)
}

// 加载平台转盘活动配置
func (this *EtcdMgr) InitPlatformActLuckyTurntable() {
	initFunc := func() int64 {
		if model.GameParamData.UseEtcd {
			logger.Logger.Info("ETCD 拉取数据:", etcd.ETCDKEY_ACT_LUCKLYTURNTABLE_PREFIX)
			//res, err := this.GetValueWithPrefix(etcd.ETCDKEY_ACT_LUCKLYTURNTABLE_PREFIX)
			//if err == nil {
			//	for i := int64(0); i < res.Count; i++ {
			//		var value PlatformLuckyTurntableConfig
			//		err = json.Unmarshal(res.Kvs[i].Value, &value)
			//		if err == nil {
			//			ActLuckyTurntableMgrSington.AdjustForBackgroundConfig(value)
			//		} else {
			//			logger.Logger.Errorf("etcd desc WithPrefix(%v) panic:%v", etcd.ETCDKEY_ACT_LUCKLYTURNTABLE_PREFIX, err)
			//		}
			//	}
			//} else {
			//	logger.Logger.Errorf("etcd get WithPrefix(%v) panic:%v", etcd.ETCDKEY_ACT_LUCKLYTURNTABLE_PREFIX, err)
			//}
		}
		return -1
	}

	// 监控数据变动
	watchFunc := func(ctx context.Context, revision int64) {
		this.GoWatch(ctx, revision, etcd.ETCDKEY_ACT_LUCKLYTURNTABLE_PREFIX, func(res clientv3.WatchResponse) error {
			for _, ev := range res.Events {
				switch ev.Type {
				case clientv3.EventTypeDelete:
				case clientv3.EventTypePut:
					//var value PlatformLuckyTurntableConfig
					//err := json.Unmarshal(ev.Kv.Value, &value)
					//if err == nil {
					//	ActLuckyTurntableMgrSington.ModifyConfig(value)
					//} else {
					//	logger.Logger.Errorf("etcd desc WithPrefix(%v) panic:%v", etcd.ETCDKEY_ACT_LUCKLYTURNTABLE_PREFIX, err)
					//}
				}
			}
			return nil
		})
	}

	this.InitAndWatch(initFunc, watchFunc)
}

// 加载余额宝配置
func (this *EtcdMgr) InitPlatformActYeb() {
	initFunc := func() int64 {
		if model.GameParamData.UseEtcd {
			logger.Logger.Info("ETCD 拉取数据:", etcd.ETCDKEY_ACT_YEB_PREFIX)
			//res, err := this.GetValueWithPrefix(etcd.ETCDKEY_ACT_YEB_PREFIX)
			//if err == nil {
			//	for i := int64(0); i < res.Count; i++ {
			//		var value YebConfig
			//		err = json.Unmarshal(res.Kvs[i].Value, &value)
			//		if err == nil {
			//			ActYebMgrSington.AdjustForBackgroundConfig(value, true, nil)
			//		} else {
			//			logger.Logger.Errorf("etcd desc WithPrefix(%v) panic:%v", etcd.ETCDKEY_ACT_YEB_PREFIX, err)
			//		}
			//	}
			//} else {
			//	logger.Logger.Errorf("etcd get WithPrefix(%v) panic:%v", etcd.ETCDKEY_ACT_YEB_PREFIX, err)
			//}
		}
		return -1
	}

	// 监控数据变动
	watchFunc := func(ctx context.Context, revision int64) {
		this.GoWatch(ctx, revision, etcd.ETCDKEY_ACT_YEB_PREFIX, func(res clientv3.WatchResponse) error {
			for _, ev := range res.Events {
				switch ev.Type {
				case clientv3.EventTypeDelete:
				case clientv3.EventTypePut:
					//var cfg YebConfig
					//err := json.Unmarshal(ev.Kv.Value, &cfg)
					//if err == nil {
					//	ActYebMgrSington.ModifyConfig(cfg)
					//	if cfg.StartAct <= 0 {
					//		type tempObject struct {
					//			newMsgs     []*model.Message
					//			playerDatas []model.PlayerData
					//		}
					//
					//		task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
					//			tmpObj := &tempObject{
					//				newMsgs:     make([]*model.Message, 0),
					//				playerDatas: make([]model.PlayerData, 0),
					//			}
					//
					//			players := model.GetYebPlayersDatas(cfg.Platform)
					//			for _, p := range players {
					//				pp := PlayerMgrSington.GetPlayerBySnId(p.SnId)
					//				if pp == nil {
					//					// 计算离线用户利息
					//					ActYebMgrSington.CalcYebInterest(nil, &p)
					//				}
					//
					//				var otherParams []int32
					//				strTitle := i18n.Tr("cn", "YebTurnOffTitle")
					//				strContent := i18n.Tr("cn", "YebTurnOffContent", float64(p.YebData.Balance)/100)
					//				newMsg := model.NewMessage("", int32(0), p.SnId, int32(model.MSGTYPE_YEB), strTitle, strContent, p.YebData.Balance,
					//					model.MSGSTATE_UNREAD, time.Now().Unix(), model.MSGATTACHSTATE_DEFAULT, "", otherParams, cfg.Platform)
					//				err := model.InsertMessage(newMsg)
					//				if err != nil {
					//					logger.Logger.Errorf("UpdateYebConfig model.InsertMessage snid %v err %v", p.SnId, err)
					//					continue
					//				}
					//
					//				tmpObj.newMsgs = append(tmpObj.newMsgs, newMsg)
					//				tmpObj.playerDatas = append(tmpObj.playerDatas, p)
					//
					//				err = model.ResetYeb(p.SnId, "")
					//				if err != nil {
					//					logger.Logger.Errorf("UpdateYebConfig err %v", err)
					//					continue
					//				}
					//
					//				if pp != nil {
					//					pp.YebData.TotalIncome = 0
					//					pp.YebData.PrevIncome = 0
					//					pp.YebData.Balance = 0
					//					pp.YebData.InterestTs = 0
					//				}
					//			}
					//
					//			model.RemoveActYebLog(cfg.Platform)
					//			return tmpObj
					//		}), task.CompleteNotifyWrapper(func(data interface{}, t *task.Task) {
					//			if obj, ok := data.(*tempObject); ok {
					//				for i, p := range obj.playerDatas {
					//					pp := PlayerMgrSington.GetPlayerBySnId(p.SnId)
					//					if pp != nil {
					//						pp.AddMessage(obj.newMsgs[i])
					//					}
					//				}
					//			} else {
					//				logger.Logger.Error("UpdateYebConfig task result data covert failed.")
					//			}
					//		}), "UpdateYebConfig").Start()
					//	}
					//} else {
					//	logger.Logger.Errorf("etcd desc WithPrefix(%v) panic:%v", etcd.ETCDKEY_ACT_YEB_PREFIX, err)
					//}
				}
			}
			return nil
		})
	}

	this.InitAndWatch(initFunc, watchFunc)
}

// 初始化返利数据
func (this *EtcdMgr) InitRebateConfig() {
	initFunc := func() int64 {
		if model.GameParamData.UseEtcd {
			logger.Logger.Info("ETCD 初始化返利数据 拉取数据:", etcd.ETCDKEY_CONFIG_REBATE)
			res, err := this.GetValueWithPrefix(etcd.ETCDKEY_CONFIG_REBATE)
			if err == nil {
				for i := int64(0); i < res.Count; i++ {
					var RebateTasks *RebateTask
					err = json.Unmarshal(res.Kvs[i].Value, &RebateTasks)
					if err == nil {
						RebateInfoMgrSington.rebateTask[RebateTasks.Platform] = RebateTasks
					} else {
						logger.Logger.Errorf("etcd desc WithPrefix(%v) panic:%v", etcd.ETCDKEY_CONFIG_REBATE, err)
					}
				}
				if res.Header != nil {
					return res.Header.Revision
				}
			} else {
				logger.Logger.Errorf("etcd get WithPrefix(%v) panic:%v", etcd.ETCDKEY_CONFIG_REBATE, err)
			}
		}
		return -1
	}

	// 监控数据变动
	watchFunc := func(ctx context.Context, revision int64) {
		this.GoWatch(ctx, revision, etcd.ETCDKEY_CONFIG_REBATE, func(res clientv3.WatchResponse) error {
			for _, ev := range res.Events {
				switch ev.Type {
				case clientv3.EventTypeDelete:
				case clientv3.EventTypePut:
					var RebateTasks *RebateTask
					err := json.Unmarshal(ev.Kv.Value, &RebateTasks)
					if err == nil {
						RebateInfoMgrSington.rebateTask[RebateTasks.Platform] = RebateTasks
					} else {
						logger.Logger.Errorf("etcd desc WithPrefix(%v) panic:%v", etcd.ETCDKEY_CONFIG_REBATE, err)
					}
				}
			}
			return nil
		})
	}
	this.InitAndWatch(initFunc, watchFunc)
}

// 初始化代理数据
func (this *EtcdMgr) InitPromoterConfig() {
	initFunc := func() int64 {
		if model.GameParamData.UseEtcd {
			logger.Logger.Info("ETCD 初始化代理数据 拉取数据:", etcd.ETCDKEY_PROMOTER_PREFIX)
			res, err := this.GetValueWithPrefix(etcd.ETCDKEY_PROMOTER_PREFIX)
			if err == nil {
				for i := int64(0); i < res.Count; i++ {
					var promoterConfig *PromoterConfig
					err = json.Unmarshal(res.Kvs[i].Value, &promoterConfig)
					if err == nil {
						PromoterMgrSington.AddConfig(promoterConfig)
					} else {
						logger.Logger.Errorf("etcd desc WithPrefix(%v) panic:%v", etcd.ETCDKEY_PROMOTER_PREFIX, err)
					}
				}
				if res.Header != nil {
					return res.Header.Revision
				}
			} else {
				logger.Logger.Errorf("etcd get WithPrefix(%v) panic:%v", etcd.ETCDKEY_PROMOTER_PREFIX, err)
			}
		}
		return -1
	}

	// 监控数据变动
	watchFunc := func(ctx context.Context, revision int64) {
		this.GoWatch(ctx, revision, etcd.ETCDKEY_PROMOTER_PREFIX, func(res clientv3.WatchResponse) error {
			for _, ev := range res.Events {
				switch ev.Type {
				case clientv3.EventTypeDelete:
					dirs := strings.Split(string(ev.Kv.Key), "/")
					n := len(dirs)
					if n > 0 {
						promoterConfig := dirs[n-1]
						PromoterMgrSington.RemoveConfigByKey(promoterConfig)
					}
					/*
						var promoterConfig *PromoterConfig
						err := json.Unmarshal(ev.Kv.Value, &promoterConfig)
						if err == nil {
							PromoterMgrSington.RemoveConfigByKey(promoterConfig)
						} else {
							logger.Logger.Errorf("etcd desc WithPrefix(%v) panic:%v", etcd.ETCDKEY_PROMOTER_PREFIX, err)
						}
					*/
				case clientv3.EventTypePut:
					var promoterConfig *PromoterConfig
					err := json.Unmarshal(ev.Kv.Value, &promoterConfig)
					if err == nil {
						PromoterMgrSington.AddConfig(promoterConfig)
					} else {
						logger.Logger.Errorf("etcd desc WithPrefix(%v) panic:%v", etcd.ETCDKEY_PROMOTER_PREFIX, err)
					}
				}
			}
			return nil
		})
	}

	this.InitAndWatch(initFunc, watchFunc)
}

// 初始化微信分享彩金配置
func (this *EtcdMgr) InitWeiXinShare() {
	initFunc := func() int64 {
		if !model.GameParamData.UseEtcd {
			return 0
		}
		logger.Logger.Info("ETCD 初始化分享彩金配置 拉取数据:", etcd.ETCDKEY_ACT_WEIXIN_SHARE_PREFIX)
		//res, err := this.GetValueWithPrefix(etcd.ETCDKEY_ACT_WEIXIN_SHARE_PREFIX)
		//if err != nil {
		//	logger.Logger.Errorf("etcd get WithPrefix(%v) panic:%v", etcd.ETCDKEY_ACT_WEIXIN_SHARE_PREFIX, err)
		//	return
		//}
		//for i := int64(0); i < res.Count; i++ {
		//	var cfg ShareConfig
		//	if err = json.Unmarshal(res.Kvs[i].Value, &cfg); err != nil {
		//		logger.Logger.Errorf("etcd desc WithPrefix(%v) panic:%v", etcd.ETCDKEY_ACT_WEIXIN_SHARE_PREFIX, err)
		//		continue
		//	}
		//	weiXinShareMgr.AddConfig(&cfg)
		//}
		return -1
	}

	// 监控数据变动
	watchFunc := func(ctx context.Context, revision int64) {
		this.GoWatch(ctx, revision, etcd.ETCDKEY_ACT_WEIXIN_SHARE_PREFIX, func(res clientv3.WatchResponse) error {
			for _, ev := range res.Events {
				switch ev.Type {
				case clientv3.EventTypeDelete:
					//dirs := strings.Split(string(ev.Kv.Key), "/")
					//if len(dirs) > 1 {
					//	weiXinShareMgr.RemoveConfig(dirs[len(dirs)-1])
					//}

				case clientv3.EventTypePut:
					//var cfg ShareConfig
					//err := json.Unmarshal(ev.Kv.Value, &cfg)
					//if err == nil {
					//	weiXinShareMgr.AddConfig(&cfg)
					//	weiXinShareMgr.BroadcastMessageToPlatform(cfg.Platform, cfg.Status)
					//} else {
					//	logger.Logger.Errorf("etcd desc WithPrefix(%v) panic:%v", etcd.ETCDKEY_ACT_WEIXIN_SHARE_PREFIX, err)
					//}
				}
			}
			return nil
		})
	}

	this.InitAndWatch(initFunc, watchFunc)
}

// 加载vip活动配置
func (this *EtcdMgr) InitPlatformActVip() {
	initFunc := func() int64 {
		if model.GameParamData.UseEtcd {
			logger.Logger.Info("ETCD 拉取数据:", etcd.ETCDKEY_ACT_VIP_PREFIX)
			//res, err := this.GetValueWithPrefix(etcd.ETCDKEY_ACT_VIP_PREFIX)
			//if err == nil {
			//	for i := int64(0); i < res.Count; i++ {
			//		var vipConfig ActVipPlateformConfig
			//		err = json.Unmarshal(res.Kvs[i].Value, &vipConfig)
			//		if err == nil {
			//			ActVipMgrSington.AddConfig(&vipConfig, vipConfig.Platform)
			//		} else {
			//			logger.Logger.Errorf("etcd desc WithPrefix(%v) panic:%v", etcd.ETCDKEY_ACT_VIP_PREFIX, err)
			//		}
			//	}
			//} else {
			//	logger.Logger.Errorf("etcd get WithPrefix(%v) panic:%v", etcd.ETCDKEY_ACT_VIP_PREFIX, err)
			//}
		}
		return -1
	}

	// 监控数据变动
	watchFunc := func(ctx context.Context, revision int64) {
		this.GoWatch(ctx, revision, etcd.ETCDKEY_ACT_VIP_PREFIX, func(res clientv3.WatchResponse) error {
			for _, ev := range res.Events {
				switch ev.Type {
				case clientv3.EventTypeDelete:
				case clientv3.EventTypePut:
					//var vipConfig ActVipPlateformConfig
					//err := json.Unmarshal(ev.Kv.Value, &vipConfig)
					//if err == nil {
					//	ActVipMgrSington.AddConfig(&vipConfig, vipConfig.Platform)
					//} else {
					//	logger.Logger.Errorf("etcd desc WithPrefix(%v) panic:%v", etcd.ETCDKEY_ACT_GOLDCOME_PREFIX, err)
					//}
				}
			}
			return nil
		})
	}

	this.InitAndWatch(initFunc, watchFunc)
}

// 加载活动give配置
func (this *EtcdMgr) InitPlatformAct() {
	initFunc := func() int64 {
		if model.GameParamData.UseEtcd {
			logger.Logger.Info("ETCD 拉取数据:", etcd.ETCDKEY_ACT_GIVE_PREFIX)
			res, err := this.GetValueWithPrefix(etcd.ETCDKEY_ACT_GIVE_PREFIX)
			if err == nil {
				for i := int64(0); i < res.Count; i++ {
					var vipConfig ActGivePlateformConfig
					err = json.Unmarshal(res.Kvs[i].Value, &vipConfig)
					if err == nil {
						ActMgrSington.AddGiveConfig(&vipConfig, vipConfig.Platform)
					} else {
						logger.Logger.Errorf("etcd desc WithPrefix(%v) panic:%v", etcd.ETCDKEY_ACT_GIVE_PREFIX, err)
					}
				}
				if res.Header != nil {
					return res.Header.Revision
				}
			} else {
				logger.Logger.Errorf("etcd get WithPrefix(%v) panic:%v", etcd.ETCDKEY_ACT_GIVE_PREFIX, err)
			}
		}
		return -1
	}

	// 监控数据变动
	watchFunc := func(ctx context.Context, revision int64) {
		this.GoWatch(ctx, revision, etcd.ETCDKEY_ACT_GIVE_PREFIX, func(res clientv3.WatchResponse) error {
			for _, ev := range res.Events {
				switch ev.Type {
				case clientv3.EventTypeDelete:
				case clientv3.EventTypePut:
					var vipConfig ActGivePlateformConfig
					err := json.Unmarshal(ev.Kv.Value, &vipConfig)
					if err == nil {
						ActMgrSington.AddGiveConfig(&vipConfig, vipConfig.Platform)
					} else {
						logger.Logger.Errorf("etcd desc WithPrefix(%v) panic:%v", etcd.ETCDKEY_ACT_GOLDCOME_PREFIX, err)
					}
				}
			}
			return nil
		})
	}

	this.InitAndWatch(initFunc, watchFunc)
}

//加载活动give配置
//func (this *EtcdMgr) InitPlatformPayAct() {
//	if model.GameParamData.UseEtcd {
//		logger.Logger.Info("ETCD 拉取数据:", etcd.ETCDKEY_ACT_PAY_PREFIX)
//		res, err := this.GetValueWithPrefix(etcd.ETCDKEY_ACT_PAY_PREFIX)
//		if err == nil {
//			for i := int64(0); i < res.Count; i++ {
//				var vipConfig PayActPlateformConfig
//				err = json.Unmarshal(res.Kvs[i].Value, &vipConfig)
//				if err == nil {
//					PayActMgrSington.AddPayConfig(&vipConfig, vipConfig.Platform)
//				} else {
//					logger.Logger.Errorf("etcd desc WithPrefix(%v) panic:%v", etcd.ETCDKEY_ACT_PAY_PREFIX, err)
//				}
//			}
//		} else {
//			logger.Logger.Errorf("etcd get WithPrefix(%v) panic:%v", etcd.ETCDKEY_ACT_PAY_PREFIX, err)
//		}
//	}
//
//	// 监控数据变动
//	this.GoWatch(etcd.ETCDKEY_ACT_PAY_PREFIX, func(res clientv3.WatchResponse) error {
//		for _, ev := range res.Events {
//			switch ev.Type {
//			case clientv3.EventTypeDelete:
//			case clientv3.EventTypePut:
//				var vipConfig PayActPlateformConfig
//				err := json.Unmarshal(ev.Kv.Value, &vipConfig)
//				if err == nil {
//					PayActMgrSington.AddPayConfig(&vipConfig, vipConfig.Platform)
//				} else {
//					logger.Logger.Errorf("etcd desc WithPrefix(%v) panic:%v", etcd.ETCDKEY_ACT_PAY_PREFIX, err)
//				}
//			}
//		}
//		return nil
//	})
//}

func (this *EtcdMgr) InitActLoginRandCoinConfig() {
	initFunc := func() int64 {
		if model.GameParamData.UseEtcd {
			logger.Logger.Info("ETCD 红包数据 拉取数据:", etcd.ETCDKEY_ACT_RANDCOIN_PREFIX)
			//res, err := this.GetValueWithPrefix(etcd.ETCDKEY_ACT_RANDCOIN_PREFIX)
			//if err == nil {
			//	for i := int64(0); i < res.Count; i++ {
			//		var actConfig = &ActConfig{}
			//		err = json.Unmarshal(res.Kvs[i].Value, actConfig)
			//		if err != nil {
			//			logger.Logger.Errorf("etcd desc WithPrefix(%v) panic:%v", etcd.ETCDKEY_ACT_RANDCOIN_PREFIX, err)
			//			continue
			//		}
			//		actRandCoinMgr.SetConfig(actConfig)
			//	}
			//} else {
			//	logger.Logger.Errorf("etcd get WithPrefix(%v) panic:%v", etcd.ETCDKEY_ACT_RANDCOIN_PREFIX, err)
			//}
		}
		return -1
	}

	// 监控数据变动
	watchFunc := func(ctx context.Context, revision int64) {
		this.GoWatch(ctx, revision, etcd.ETCDKEY_ACT_RANDCOIN_PREFIX, func(res clientv3.WatchResponse) error {
			for _, ev := range res.Events {
				switch ev.Type {
				case clientv3.EventTypePut:
					//var actConfig = &ActConfig{}
					//err := json.Unmarshal(ev.Kv.Value, &actConfig)
					//if err == nil {
					//	actRandCoinMgr.SetConfig(actConfig)
					//} else {
					//	logger.Logger.Errorf("etcd desc WithPrefix(%v) panic:%v", etcd.ETCDKEY_ACT_RANDCOIN_PREFIX, err)
					//}
				}
			}
			return nil
		})
	}

	this.InitAndWatch(initFunc, watchFunc)
}

// 加载活动fpay配置
func (this *EtcdMgr) InitPlatformActFPay() {
	initFunc := func() int64 {
		if model.GameParamData.UseEtcd {
			logger.Logger.Info("ETCD 拉取数据:", etcd.ETCDKEY_ACT_FPAY_PREFIX)
			//res, err := this.GetValueWithPrefix(etcd.ETCDKEY_ACT_FPAY_PREFIX)
			//if err == nil {
			//	for i := int64(0); i < res.Count; i++ {
			//		var vipConfig ActFPayPlateformConfig
			//		err = json.Unmarshal(res.Kvs[i].Value, &vipConfig)
			//		if err == nil {
			//			ActFPayMgrSington.AddConfig(&vipConfig, vipConfig.Platform)
			//		} else {
			//			logger.Logger.Errorf("etcd desc WithPrefix(%v) panic:%v", etcd.ETCDKEY_ACT_FPAY_PREFIX, err)
			//		}
			//	}
			//} else {
			//	logger.Logger.Errorf("etcd get WithPrefix(%v) panic:%v", etcd.ETCDKEY_ACT_FPAY_PREFIX, err)
			//}
		}
		return -1
	}

	// 监控数据变动
	watchFunc := func(ctx context.Context, revision int64) {
		this.GoWatch(ctx, revision, etcd.ETCDKEY_ACT_FPAY_PREFIX, func(res clientv3.WatchResponse) error {
			for _, ev := range res.Events {
				switch ev.Type {
				case clientv3.EventTypeDelete:
				case clientv3.EventTypePut:
					//var vipConfig ActFPayPlateformConfig
					//err := json.Unmarshal(ev.Kv.Value, &vipConfig)
					//if err == nil {
					//	ActFPayMgrSington.AddConfig(&vipConfig, vipConfig.Platform)
					//} else {
					//	logger.Logger.Errorf("etcd desc WithPrefix(%v) panic:%v", etcd.ETCDKEY_ACT_FPAY_PREFIX, err)
					//}
				}
			}
			return nil
		})
	}

	this.InitAndWatch(initFunc, watchFunc)
}

////加载杀率配置
//func (this *EtcdMgr) InitPlatformProfitControl() {
//	if model.GameParamData.UseEtcd {
//		logger.Logger.Info("ETCD 拉取数据:", etcd.ETCDKEY_PLATFORM_PROFITCONTROL)
//		res, err := this.GetValueWithPrefix(etcd.ETCDKEY_PLATFORM_PROFITCONTROL)
//		if err == nil {
//			for i := int64(0); i < res.Count; i++ {
//				var cfg ProfitControlConfig
//				err = json.Unmarshal(res.Kvs[i].Value, &cfg)
//				if err == nil {
//					ProfitControlMgrSington.UpdateConfig(&cfg, true)
//				} else {
//					logger.Logger.Errorf("etcd desc WithPrefix(%v) panic:%v", etcd.ETCDKEY_PLATFORM_PROFITCONTROL, err)
//				}
//			}
//		} else {
//			logger.Logger.Errorf("etcd get WithPrefix(%v) panic:%v", etcd.ETCDKEY_PLATFORM_PROFITCONTROL, err)
//		}
//	}
//
//	// 监控数据变动
//	this.GoWatch(etcd.ETCDKEY_PLATFORM_PROFITCONTROL, func(res clientv3.WatchResponse) error {
//		for _, ev := range res.Events {
//			switch ev.Type {
//			case clientv3.EventTypeDelete:
//			case clientv3.EventTypePut:
//				var cfg ProfitControlConfig
//				err := json.Unmarshal(ev.Kv.Value, &cfg)
//				if err == nil {
//					ProfitControlMgrSington.UpdateConfig(&cfg, false)
//				} else {
//					logger.Logger.Errorf("etcd desc WithPrefix(%v) panic:%v", etcd.ETCDKEY_PLATFORM_PROFITCONTROL, err)
//				}
//			}
//		}
//		return nil
//	})
//}

// 加载比赛配置
func (this *EtcdMgr) InitMatchConfig() {
	initFunc := func() int64 {
		if model.GameParamData.UseEtcd {
			logger.Logger.Info("ETCD 拉取数据:", etcd.ETCDKEY_MATCH_PROFIX)
			//res, err := this.GetValueWithPrefix(etcd.ETCDKEY_MATCH_PROFIX)
			//if err == nil {
			//	for i := int64(0); i < res.Count; i++ {
			//		var cfg MatchConfig
			//		err = json.Unmarshal(res.Kvs[i].Value, &cfg)
			//		if err == nil {
			//			MatchMgrSington.UpdateConfig(&cfg, true)
			//		} else {
			//			logger.Logger.Errorf("etcd desc WithPrefix(%v) panic:%v", etcd.ETCDKEY_MATCH_PROFIX, err)
			//		}
			//	}
			//} else {
			//	logger.Logger.Errorf("etcd get WithPrefix(%v) panic:%v", etcd.ETCDKEY_MATCH_PROFIX, err)
			//}
		}
		return -1
	}

	// 监控数据变动
	watchFunc := func(ctx context.Context, revision int64) {
		this.GoWatch(ctx, revision, etcd.ETCDKEY_MATCH_PROFIX, func(res clientv3.WatchResponse) error {
			for _, ev := range res.Events {
				switch ev.Type {
				case clientv3.EventTypeDelete:
				case clientv3.EventTypePut:
					//var cfg MatchConfig
					//err := json.Unmarshal(ev.Kv.Value, &cfg)
					//if err == nil {
					//	MatchMgrSington.UpdateConfig(&cfg, false)
					//} else {
					//	logger.Logger.Errorf("etcd desc WithPrefix(%v) panic:%v", etcd.ETCDKEY_MATCH_PROFIX, err)
					//}
				}
			}
			return nil
		})
	}

	this.InitAndWatch(initFunc, watchFunc)
}

// 加载比赛券活动配置
func (this *EtcdMgr) InitActTicketConfig() {
	initFunc := func() int64 {
		if model.GameParamData.UseEtcd {
			logger.Logger.Info("ETCD 拉取数据:", etcd.ETCDKEY_ACT_TICKET_PROFIX)
			//res, err := this.GetValueWithPrefix(etcd.ETCDKEY_ACT_TICKET_PROFIX)
			//if err == nil {
			//	for i := int64(0); i < res.Count; i++ {
			//		var cfg ActTicketConfig
			//		err = json.Unmarshal(res.Kvs[i].Value, &cfg)
			//		if err == nil {
			//			ActTicketMgrSington.UpdateConfig(&cfg, true)
			//		} else {
			//			logger.Logger.Errorf("etcd desc WithPrefix(%v) panic:%v", etcd.ETCDKEY_ACT_TICKET_PROFIX, err)
			//		}
			//	}
			//} else {
			//	logger.Logger.Errorf("etcd get WithPrefix(%v) panic:%v", etcd.ETCDKEY_ACT_TICKET_PROFIX, err)
			//}
		}
		return -1
	}

	// 监控数据变动
	watchFunc := func(ctx context.Context, revision int64) {
		this.GoWatch(ctx, revision, etcd.ETCDKEY_ACT_TICKET_PROFIX, func(res clientv3.WatchResponse) error {
			for _, ev := range res.Events {
				switch ev.Type {
				case clientv3.EventTypeDelete:
				case clientv3.EventTypePut:
					//var cfg ActTicketConfig
					//err := json.Unmarshal(ev.Kv.Value, &cfg)
					//if err == nil {
					//	ActTicketMgrSington.UpdateConfig(&cfg, false)
					//} else {
					//	logger.Logger.Errorf("etcd desc WithPrefix(%v) panic:%v", etcd.ETCDKEY_ACT_TICKET_PROFIX, err)
					//}
				}
			}
			return nil
		})
	}

	this.InitAndWatch(initFunc, watchFunc)
}

// 加载积分商城配置
func (this *EtcdMgr) InitGradeShopConfig() {
	initFunc := func() int64 {

		if model.GameParamData.UseEtcd {
			logger.Logger.Info("ETCD 拉取数据:", etcd.ETCDKEY_MATCH_GRADESHOP)
			//res, err := this.GetValueWithPrefix(etcd.ETCDKEY_MATCH_GRADESHOP)
			//if err == nil {
			//	for i := int64(0); i < res.Count; i++ {
			//		var cfg GradeShopConfig
			//		err = json.Unmarshal(res.Kvs[i].Value, &cfg)
			//		if err == nil {
			//			GradeShopMgrSington.UpdateConfig(&cfg, true)
			//		} else {
			//			logger.Logger.Errorf("etcd desc WithPrefix(%v) panic:%v", etcd.ETCDKEY_MATCH_PROFIX, err)
			//		}
			//	}
			//} else {
			//	logger.Logger.Errorf("etcd get WithPrefix(%v) panic:%v", etcd.ETCDKEY_MATCH_PROFIX, err)
			//}
		}
		return -1
	}

	// 监控数据变动
	watchFunc := func(ctx context.Context, revision int64) {
		this.GoWatch(ctx, revision, etcd.ETCDKEY_MATCH_GRADESHOP, func(res clientv3.WatchResponse) error {
			for _, ev := range res.Events {
				switch ev.Type {
				case clientv3.EventTypeDelete:
				case clientv3.EventTypePut:
					//var cfg GradeShopConfig
					//err := json.Unmarshal(ev.Kv.Value, &cfg)
					//if err == nil {
					//	GradeShopMgrSington.UpdateConfig(&cfg, false)
					//} else {
					//	logger.Logger.Errorf("etcd desc WithPrefix(%v) panic:%v", etcd.ETCDKEY_MATCH_PROFIX, err)
					//}
				}
			}
			return nil
		})
	}

	this.InitAndWatch(initFunc, watchFunc)
}

// 加载用户分层配置
func (this *EtcdMgr) InitLogicLevelConfig() {

	initFunc := func() int64 {
		if model.GameParamData.UseEtcd {
			logger.Logger.Info("ETCD 拉取数据:", etcd.ETCDKEY_CONFIG_LOGICLEVEL)
			res, err := this.GetValueWithPrefix(etcd.ETCDKEY_CONFIG_LOGICLEVEL)
			if err == nil {
				for i := int64(0); i < res.Count; i++ {
					var cfg LogicLevelConfig
					err = json.Unmarshal(res.Kvs[i].Value, &cfg)
					if err == nil {
						LogicLevelMgrSington.UpdateConfig(&cfg)
					} else {
						logger.Logger.Errorf("etcd desc WithPrefix(%v) panic:%v", etcd.ETCDKEY_CONFIG_LOGICLEVEL, err)
					}
				}
				if res.Header != nil {
					return res.Header.Revision
				}
			} else {
				logger.Logger.Errorf("etcd get WithPrefix(%v) panic:%v", etcd.ETCDKEY_CONFIG_LOGICLEVEL, err)
			}
		}
		return -1
	}

	// 监控数据变动
	watchFunc := func(ctx context.Context, revision int64) {
		this.GoWatch(ctx, revision, etcd.ETCDKEY_CONFIG_LOGICLEVEL, func(res clientv3.WatchResponse) error {
			for _, ev := range res.Events {
				switch ev.Type {
				case clientv3.EventTypeDelete:
				case clientv3.EventTypePut:
					var cfg LogicLevelConfig
					err := json.Unmarshal(ev.Kv.Value, &cfg)
					if err == nil {
						LogicLevelMgrSington.UpdateConfig(&cfg)
					} else {
						logger.Logger.Errorf("etcd desc WithPrefix(%v) panic:%v", etcd.ETCDKEY_CONFIG_LOGICLEVEL, err)
					}
				}
			}
			return nil
		})
	}

	this.InitAndWatch(initFunc, watchFunc)
}

// 加载全局游戏开关
func (this *EtcdMgr) InitGameGlobalStatus() {
	initFunc := func() int64 {
		if model.GameParamData.UseEtcd {
			logger.Logger.Info("ETCD 拉取数据:", etcd.ETCDKEY_GAME_CONFIG_GLOBAL)
			res, err := this.GetValueWithPrefix(etcd.ETCDKEY_GAME_CONFIG_GLOBAL)
			if err == nil {
				for i := int64(0); i < res.Count; i++ {
					var cfg webapi_proto.GameConfigGlobal
					err = proto.Unmarshal(res.Kvs[i].Value, &cfg)
					if err == nil {
						for _, v := range cfg.GetGameStatus() {
							gameId := v.GetGameId()
							status := v.GetStatus()
							PlatformMgrSington.GameStatus[gameId] = status
						}
					} else {
						logger.Logger.Errorf("etcd desc WithPrefix(%v) panic:%v", etcd.ETCDKEY_GAME_CONFIG_GLOBAL, err)
					}
				}
				if res.Header != nil {
					return res.Header.Revision
				}
			} else {
				logger.Logger.Errorf("etcd get WithPrefix(%v) panic:%v", etcd.ETCDKEY_GAME_CONFIG_GLOBAL, err)
			}
		}
		return -1
	}

	// 监控数据变动
	watchFunc := func(ctx context.Context, revision int64) {
		this.GoWatch(ctx, revision, etcd.ETCDKEY_GAME_CONFIG_GLOBAL, func(res clientv3.WatchResponse) error {
			for _, ev := range res.Events {
				switch ev.Type {
				case clientv3.EventTypeDelete:
				case clientv3.EventTypePut:
					var cfg webapi_proto.GameConfigGlobal
					err := proto.Unmarshal(ev.Kv.Value, &cfg)
					if err == nil {
						//closeGameId := make([]int32, 0)
						//openGameId := make([]int32, 0)
						cfgs := make([]*hall_proto.GameConfig1, 0)
						for _, v := range cfg.GetGameStatus() {
							gameId := v.GetGameId()
							status := v.GetStatus()

							if PlatformMgrSington.GameStatus[gameId] != status {
								cfgs = append(cfgs, &hall_proto.GameConfig1{
									LogicId: gameId,
									Status:  status,
								})
								//if status {
								//	openGameId = append(openGameId, gameId)
								//} else {
								//	closeGameId = append(closeGameId, gameId)
								//}
							}
							PlatformMgrSington.GameStatus[gameId] = status
						}
						//PlatformMgrSington.ChangeGameStatus(openGameId, closeGameId, cfgs)
						PlatformMgrSington.ChangeGameStatus(cfgs)
					} else {
						logger.Logger.Errorf("etcd desc WithPrefix(%v) panic:%v", etcd.ETCDKEY_GAME_CONFIG_GLOBAL, err)
					}
				}
			}
			return nil
		})
	}

	this.InitAndWatch(initFunc, watchFunc)
}

// 加载兑换
func (this *EtcdMgr) InitExchangeShop() {
	initFunc := func() int64 {
		if model.GameParamData.UseEtcd {
			logger.Logger.Info("ETCD 拉取数据:", etcd.ETCDKEY_SHOP_EXCHANGE)
			res, err := this.GetValueWithPrefix(etcd.ETCDKEY_SHOP_EXCHANGE)
			if err == nil {
				for i := int64(0); i < res.Count; i++ {
					cfg := &webapi_proto.ExchangeShopList{}
					//msg := &webapi.ASSrvCtrlClose{}
					err = proto.Unmarshal(res.Kvs[i].Value, cfg)
					if err == nil {
						ShopMgrSington.UpExShop(cfg)
					} else {
						logger.Logger.Errorf("etcd desc WithPrefix(%v) panic:%v", etcd.ETCDKEY_SHOP_EXCHANGE, err)
					}
				}
				if res.Header != nil {
					return res.Header.Revision
				}
			} else {
				logger.Logger.Errorf("etcd get WithPrefix(%v) panic:%v", etcd.ETCDKEY_SHOP_EXCHANGE, err)
			}
		}
		return -1
	}

	// 监控数据变动
	watchFunc := func(ctx context.Context, revision int64) {
		this.GoWatch(ctx, revision, etcd.ETCDKEY_SHOP_EXCHANGE, func(res clientv3.WatchResponse) error {
			for _, ev := range res.Events {
				switch ev.Type {
				case clientv3.EventTypeDelete:
				case clientv3.EventTypePut:
					cfg := &webapi_proto.ExchangeShopList{}
					err := proto.Unmarshal(ev.Kv.Value, cfg)
					if err == nil {
						ShopMgrSington.UpExShop(cfg)
					} else {
						logger.Logger.Errorf("etcd desc WithPrefix(%v) panic:%v", etcd.ETCDKEY_SHOP_EXCHANGE, err)
					}
				}
			}
			return nil
		})
	}

	this.InitAndWatch(initFunc, watchFunc)
}

// 加载道具商品
func (this *EtcdMgr) InitItemShop() {
	initFunc := func() int64 {
		if model.GameParamData.UseEtcd {
			logger.Logger.Info("ETCD 拉取数据:", etcd.ETCDKEY_SHOP_ITEM)
			res, err := this.GetValueWithPrefix(etcd.ETCDKEY_SHOP_ITEM)
			if err == nil {
				for i := int64(0); i < res.Count; i++ {
					cfg := &webapi_proto.ItemShopList{}
					//msg := &webapi.ASSrvCtrlClose{}
					err = proto.Unmarshal(res.Kvs[i].Value, cfg)
					if err == nil {
						ShopMgrSington.InitItemShop(cfg)
					} else {
						logger.Logger.Errorf("etcd desc WithPrefix(%v) panic:%v", etcd.ETCDKEY_SHOP_ITEM, err)
					}
				}
				if res.Header != nil {
					return res.Header.Revision
				}
			} else {
				logger.Logger.Errorf("etcd get WithPrefix(%v) panic:%v", etcd.ETCDKEY_SHOP_ITEM, err)
			}
		}
		return -1
	}

	// 监控数据变动
	watchFunc := func(ctx context.Context, revision int64) {
		this.GoWatch(ctx, revision, etcd.ETCDKEY_SHOP_ITEM, func(res clientv3.WatchResponse) error {
			for _, ev := range res.Events {
				switch ev.Type {
				case clientv3.EventTypeDelete:
				case clientv3.EventTypePut:
					cfg := &webapi_proto.ItemShopList{}
					err := proto.Unmarshal(ev.Kv.Value, cfg)
					if err == nil {
						ShopMgrSington.UpItemShop(cfg)
					} else {
						logger.Logger.Errorf("etcd desc WithPrefix(%v) panic:%v", etcd.ETCDKEY_SHOP_ITEM, err)
					}
				}
			}
			return nil
		})
	}

	this.InitAndWatch(initFunc, watchFunc)
}

// 加载公告
func (this *EtcdMgr) InitCommonNotice() {
	initFunc := func() int64 {
		if model.GameParamData.UseEtcd {
			logger.Logger.Info("ETCD 拉取数据:", etcd.ETCDKEY_GAME_NOTICE)
			res, err := this.GetValueWithPrefix(etcd.ETCDKEY_GAME_NOTICE)
			if err == nil {
				for i := int64(0); i < res.Count; i++ {

					var cfgs webapi_proto.CommonNoticeList
					err = proto.Unmarshal(res.Kvs[i].Value, &cfgs)

					if err == nil && len(cfgs.List) != 0 {
						now := time.Now().Unix()
						for i := 0; i < len(cfgs.List); {
							if cfgs.List[i].GetEndTime() < now || cfgs.List[i].GetStartTime() > now {

								cfgs.List = append(cfgs.List[:i], cfgs.List[i+1:]...)
							} else {
								i++
							}
						}

						PlatformMgrSington.CommonNotices[cfgs.Platform] = &cfgs

					}
				}
				if res.Header != nil {
					return res.Header.Revision
				}
			} else {
				logger.Logger.Errorf("etcd get WithPrefix(%v) panic:%v", etcd.ETCDKEY_GAME_NOTICE, err)
			}

		}
		return -1
	}

	// 监控数据变动
	watchFunc := func(ctx context.Context, revision int64) {
		this.GoWatch(ctx, revision, etcd.ETCDKEY_GAME_NOTICE, func(res clientv3.WatchResponse) error {
			for _, ev := range res.Events {
				switch ev.Type {
				case clientv3.EventTypeDelete:
				case clientv3.EventTypePut:
					var cfgs webapi_proto.CommonNoticeList
					err := proto.Unmarshal(ev.Kv.Value, &cfgs)

					if err == nil && len(cfgs.List) != 0 {
						now := time.Now().Unix()
						for i := 0; i < len(cfgs.List); {
							if cfgs.List[i].GetEndTime() < now || cfgs.List[i].GetStartTime() > now {

								cfgs.List = append(cfgs.List[:i], cfgs.List[i+1:]...)
							} else {
								i++
							}
						}

						PlatformMgrSington.CommonNotices[cfgs.Platform] = &cfgs

					} else {
						logger.Logger.Errorf("etcd desc WithPrefix(%v) panic:%v", etcd.ETCDKEY_GAME_NOTICE, err)
					}
				}
			}
			return nil
		})
	}

	this.InitAndWatch(initFunc, watchFunc)
}

// 加载比赛配置
func (this *EtcdMgr) InitGameMatchDate() {
	initFunc := func() int64 {
		if model.GameParamData.UseEtcd {
			logger.Logger.Info("ETCD 拉取数据:", etcd.ETCDKEY_GAME_MATCH)
			res, err := this.GetValueWithPrefix(etcd.ETCDKEY_GAME_MATCH)
			if err == nil {
				for i := int64(0); i < res.Count; i++ {
					var cfgs webapi_proto.GameMatchDateList
					err = proto.Unmarshal(res.Kvs[i].Value, &cfgs)
					if err == nil && len(cfgs.List) != 0 {
						TournamentMgr.UpdateData(true, &cfgs)
					}
				}
				if res.Header != nil {
					return res.Header.Revision
				}
			} else {
				logger.Logger.Errorf("etcd get WithPrefix(%v) panic:%v", etcd.ETCDKEY_GAME_MATCH, err)
			}

		}
		return -1
	}

	// 监控数据变动
	watchFunc := func(ctx context.Context, revision int64) {
		this.GoWatch(ctx, revision, etcd.ETCDKEY_GAME_MATCH, func(res clientv3.WatchResponse) error {
			for _, ev := range res.Events {
				switch ev.Type {
				case clientv3.EventTypeDelete:
				case clientv3.EventTypePut:
					var cfgs webapi_proto.GameMatchDateList
					err := proto.Unmarshal(ev.Kv.Value, &cfgs)
					if err == nil && len(cfgs.List) != 0 {
						TournamentMgr.UpdateData(false, &cfgs)
					} else {
						logger.Logger.Errorf("etcd desc WithPrefix(%v) panic:%v", etcd.ETCDKEY_GAME_MATCH, err)
					}
				}
			}
			return nil
		})
	}

	this.InitAndWatch(initFunc, watchFunc)
}

func (this *EtcdMgr) Init() {
	if model.GameParamData.UseEtcd {
		logger.Logger.Infof("EtcdClient开始连接url:%v;etcduser:%v;etcdpwd:%v", common.CustomConfig.GetStrings("etcdurl"), common.CustomConfig.GetString("etcduser"), common.CustomConfig.GetString("etcdpwd"))
		err := this.Open(common.CustomConfig.GetStrings("etcdurl"), common.CustomConfig.GetString("etcduser"), common.CustomConfig.GetString("etcdpwd"), time.Minute)
		if err != nil {
			logger.Logger.Tracef("EtcdMgr.Open err:%v", err)
			return
		}
	}
}

func (this *EtcdMgr) Shutdown() {
	this.Close()
}

func (this *EtcdMgr) Reset() {
	this.Close()
	this.Init()
	this.ReInitAndWatchAll()
}
