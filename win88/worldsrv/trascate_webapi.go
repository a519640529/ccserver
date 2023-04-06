package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io"
	"sort"

	"games.yol.com/win88/gamerule/crash"
	"games.yol.com/win88/protocol/bag"
	"games.yol.com/win88/protocol/qpapi"
	"games.yol.com/win88/protocol/server"
	webapi "games.yol.com/win88/protocol/telegramapi"
	"games.yol.com/win88/srvdata"
	"github.com/globalsign/mgo/bson"

	//"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"sync"
	"time"

	login_proto "games.yol.com/win88/protocol/login"
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/task"

	"games.yol.com/win88/common"
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	webapi_proto "games.yol.com/win88/protocol/webapi"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/core/transact"
)

const __REQIP__ = "__REQIP__"

const (
	WebAPITransactParam_Path = iota
	WebAPITransactParam_CreateTime
)

var (
	WebAPIErrParam    = errors.New("param err")
	WebAPIErrNoPlayer = errors.New("player no find")
)

func init() {
	transact.RegisteHandler(common.TransType_WebApi, &WebAPITranscateHandler{})
}

var WebAPIHandlerMgrSingleton = &WebAPIHandlerMgr{wshMap: make(map[string]WebAPIHandler)}

type WebAPITranscateHandler struct {
}

var WebAPIStats = make(map[string]*model.APITransactStats)

// 返利配置
type RebateInfos struct {
	Platform           string                //平台名称
	RebateSwitch       bool                  //返利开关
	RebateManState     int                   //返利是开启个人返利  0 关闭  1 开启
	ReceiveMode        int                   //领取方式  0实时领取  1次日领取
	NotGiveOverdue     int                   //0不过期   1过期不给  2过期邮件给
	RebateGameCfg      []*RebateGameCfg      //key为"gameid"+"gamemode"
	RebateGameThirdCfg []*RebateGameThirdCfg //第三方key
	Version            int                   //活动版本 后台控制
}

// 后台请求玩家的基础信息数据
type PlayerDataForWebSimple struct {
	SnId             int32     //数字唯一id
	Name             string    //名字
	Platform         string    //平台
	Channel          string    //渠道信息
	DeviceOS         string    //设备操作系统
	Ip               string    //最后登录ip地址
	LastLoginTime    time.Time //最后登陆时间
	CreateTime       time.Time //创建时间
	Tel              string    //电话号码
	Coin             int64     //金豆
	SafeBoxCoin      int64     //保险箱金币
	AlipayAccount    string    //支付宝账号
	AlipayAccName    string    //支付宝实名
	Bank             string    //绑定的银行名称
	BankAccount      string    //绑定的银行账号
	BankAccName      string    //绑定的银行账号
	BeUnderAgentCode string    //隶属经销商（推广人）
	PromoterTree     int32     //推广树信息
}

func SimplifyPlayerDataForWeb(pd *model.PlayerData) *PlayerDataForWebSimple {
	if pd == nil {
		return nil
	}

	return &PlayerDataForWebSimple{
		SnId:             pd.SnId,             //数字唯一id
		Name:             pd.Name,             //名字
		Platform:         pd.Platform,         //平台
		Channel:          pd.Channel,          //渠道信息
		DeviceOS:         pd.DeviceOS,         //设备操作系统
		Ip:               pd.Ip,               //最后登录ip地址
		LastLoginTime:    pd.LastLoginTime,    //最后登陆时间
		CreateTime:       pd.CreateTime,       //创建时间
		Tel:              pd.Tel,              //电话号码
		Coin:             pd.Coin,             //金豆
		SafeBoxCoin:      pd.SafeBoxCoin,      //保险箱金币
		AlipayAccount:    pd.AlipayAccount,    //支付宝账号
		AlipayAccName:    pd.AlipayAccName,    //支付宝实名
		Bank:             pd.Bank,             //绑定的银行名称
		BankAccount:      pd.BankAccount,      //绑定的银行账号
		BankAccName:      pd.BankAccName,      //绑定的银行账号
		BeUnderAgentCode: pd.BeUnderAgentCode, //隶属经销商（推广人）
		PromoterTree:     pd.PromoterTree,     //推广树信息
	}
}

func (this *WebAPITranscateHandler) OnExcute(tNode *transact.TransNode, ud interface{}) transact.TransExeResult {
	logger.Logger.Trace("WebAPITranscateHandler.OnExcute ")
	req := &common.M2GWebApiRequest{}
	err := netlib.UnmarshalPacketNoPackId(ud.([]byte), req)
	if err == nil {
		wsh := WebAPIHandlerMgrSingleton.GetWebAPIHandler(req.Path)
		if wsh == nil {
			logger.Logger.Trace("WebAPITranscateHandler no registe WebAPIHandler ", req.Path)
			return transact.TransExeResult_Failed
		}

		tNode.TransEnv.SetField(WebAPITransactParam_Path, req.Path)
		tNode.TransEnv.SetField(WebAPITransactParam_CreateTime, time.Now())

		tag, msg := wsh.Handler(tNode, req.Body)
		tNode.TransRep.RetFiels = msg
		switch tag {
		case common.ResponseTag_Ok:
			return transact.TransExeResult_Success
		case common.ResponseTag_TransactYield:
			return transact.TransExeResult_Yield
		}
	}
	logger.Logger.Trace("WebAPITranscateHandler.OnExcute err:", err)
	return transact.TransExeResult_Failed
}

func (this *WebAPITranscateHandler) OnCommit(tNode *transact.TransNode) transact.TransExeResult {
	logger.Logger.Trace("WebAPITranscateHandler.OnCommit ")
	paramPath := tNode.TransEnv.GetField(WebAPITransactParam_Path)
	paramCreateTime := tNode.TransEnv.GetField(WebAPITransactParam_CreateTime)
	if path, ok := paramPath.(string); ok {
		var stats *model.APITransactStats
		var exist bool
		if stats, exist = WebAPIStats[path]; !exist {
			stats = &model.APITransactStats{}
			WebAPIStats[path] = stats
		}

		if stats != nil {
			stats.RunTimes++
			runingTime := int64(time.Now().Sub(paramCreateTime.(time.Time)) / time.Millisecond)
			if runingTime > stats.MaxRuningTime {
				stats.MaxRuningTime = runingTime
			}
			stats.TotalRuningTime += runingTime
			stats.SuccessTimes++
		}
	}
	return transact.TransExeResult_Success
}

func (this *WebAPITranscateHandler) OnRollBack(tNode *transact.TransNode) transact.TransExeResult {
	logger.Logger.Trace("WebAPITranscateHandler.OnRollBack ")
	paramPath := tNode.TransEnv.GetField(WebAPITransactParam_Path)
	paramCreateTime := tNode.TransEnv.GetField(WebAPITransactParam_CreateTime)
	if path, ok := paramPath.(string); ok {
		var stats *model.APITransactStats
		var exist bool
		if stats, exist = WebAPIStats[path]; !exist {
			stats = &model.APITransactStats{}
			WebAPIStats[path] = stats
		}

		if stats != nil {
			stats.RunTimes++
			runingTime := int64(time.Now().Sub(paramCreateTime.(time.Time)) / time.Millisecond)
			if runingTime > stats.MaxRuningTime {
				stats.MaxRuningTime = runingTime
			}
			stats.TotalRuningTime += runingTime
			stats.FailedTimes++
		}
	}
	return transact.TransExeResult_Success
}

func (this *WebAPITranscateHandler) OnChildTransRep(tNode *transact.TransNode, hChild transact.TransNodeID, retCode int,
	ud interface{}) transact.TransExeResult {
	logger.Logger.Trace("WebAPITranscateHandler.OnChildTransRep ")
	return transact.TransExeResult_Success
}

type WebAPIHandler interface {
	Handler(*transact.TransNode, []byte) (int, proto.Message)
}

type WebAPIHandlerWrapper func(*transact.TransNode, []byte) (int, proto.Message)

func (wshw WebAPIHandlerWrapper) Handler(tNode *transact.TransNode, params []byte) (int, proto.Message) {
	return wshw(tNode, params)
}

type WebAPIHandlerMgr struct {
	wshMap       map[string]WebAPIHandler
	DataWaitList sync.Map
}

func (this *WebAPIHandlerMgr) RegisteWebAPIHandler(name string, wsh WebAPIHandler) {
	this.wshMap[name] = wsh
}

func (this *WebAPIHandlerMgr) GetWebAPIHandler(name string) WebAPIHandler {
	if wsh, exist := this.wshMap[name]; exist {
		return wsh
	}
	return nil
}

func init() {
	//API用户登录
	//WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/api/Member/APIMemberRegisterOrLogin", WebAPIHandlerWrapper(
	//	func(tNode *transact.TransNode, params []byte) (int, proto.Message) {
	//		logger.Logger.Trace("WebAPIHandler:/api/Member/APIMemberRegisterOrLogin", params)
	//		pack := &webapi.SALogin{}
	//		msg := &webapi.ASLogin{}
	//		err1 := proto.Unmarshal(params, msg)
	//		if err1 != nil {
	//			pack.Tag = webapi.TagCode_FAILED
	//			pack.Msg = "数据序列化失败" + err1.Error()
	//			return common.ResponseTag_ParamError, pack
	//		}
	//
	//		platformID, channel, _, _, tagkey := PlatformMgrSington.GetPlatformByPackageTag(msg.GetPlatformTag())
	//		platform := PlatformMgrSington.GetPlatform(strconv.Itoa(int(platformID)))
	//		if platform == nil {
	//			pack.Tag = webapi.TagCode_FAILED
	//			pack.Msg = "没有对应的包标识"
	//			return common.ResponseTag_ParamError, pack
	//		}
	//		merchantkey := platform.MerchantKey
	//
	//		sign := msg.GetSign()
	//
	//		raw := fmt.Sprintf("%v%v%v%v%v", msg.GetTelegramId(), msg.GetPlatformTag(), msg.GetUsername(), merchantkey, msg.GetTs())
	//		h := md5.New()
	//		io.WriteString(h, raw)
	//		newsign := hex.EncodeToString(h.Sum(nil))
	//
	//		if newsign != sign {
	//			pack.Tag = webapi.TagCode_FAILED
	//			pack.Msg = "商户验签失败"
	//			return common.ResponseTag_ParamError, pack
	//		}
	//
	//		var acc *model.Account
	//		var errcode int
	//
	//		rawpwd := fmt.Sprintf("%v%v%v", msg.GetSign(), common.GetAppId(), time.Now().Unix())
	//		hpwd := md5.New()
	//		io.WriteString(hpwd, rawpwd)
	//		pwd := hex.EncodeToString(h.Sum(nil))
	//
	//		task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
	//			acc, errcode = model.AccountIsExist(msg.GetTelegramId(), msg.GetTelegramId(), "", platform.IdStr, time.Now().Unix(), 5, tagkey, false)
	//			if errcode == 2 {
	//				var err int
	//				acc, err = model.InsertAccount(msg.GetTelegramId(), pwd, platform.IdStr, strconv.Itoa(int(channel)), "", "",
	//					"telegramapi", 0, 0, msg.GetPlatformTag(), "", "", "", tagkey)
	//				if acc != nil { //需要预先创建玩家数据
	//					_, _ = model.GetPlayerData(acc.PackegeTag,acc.AccountId.Hex())
	//				}
	//				if err == 5 {
	//					return 1
	//				} else {
	//					//t.flag = login_proto.OpResultCode_OPRC_Login_CreateAccError
	//				}
	//			}
	//			return errcode
	//		}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
	//			switch data.(int) {
	//			case 1:
	//				{
	//					if err := model.UpdatePlayerTokenPassword(acc.Platform, acc.AccountId.Hex(), pwd); err != nil {
	//						pack.Tag = webapi.TagCode_FAILED
	//						pack.Msg = err.Error()
	//					} else {
	//						tokenuserdata := &common.TokenUserData{
	//							TelegramId: msg.GetTelegramId(),
	//							Password:   pwd,
	//							Packagetag: msg.GetPlatformTag(),
	//							Expired:    time.Now().Add(time.Minute * 15).Unix(),
	//						}
	//
	//						token, err := common.CreateTokenAes(tokenuserdata)
	//						if err != nil {
	//							pack.Tag = webapi.TagCode_FAILED
	//							pack.Msg = err.Error()
	//						} else {
	//							pack.Tag = webapi.TagCode_SUCCESS
	//							pack.Msg = ""
	//							pack.Snid = proto.Int32(acc.SnId)
	//							pack.Token = fmt.Sprintf("%v?token=%v&snid=%v", model.GameParamData.CSURL, token, acc.SnId)
	//						}
	//					}
	//				}
	//			case 2:
	//				{
	//					pack.Tag = webapi.TagCode_FAILED
	//					pack.Msg = "创建帐号失败"
	//				}
	//			case 3:
	//				{
	//					//t.flag = login_proto.OpResultCode_OPRC_LoginPassError
	//					pack.Tag = webapi.TagCode_FAILED
	//					pack.Msg = "账号密码错误"
	//				}
	//			case 4:
	//				{
	//					//t.flag = login_proto.OpResultCode_OPRC_AccountBeFreeze
	//					pack.Tag = webapi.TagCode_FAILED
	//					pack.Msg = "帐号冻结"
	//				}
	//			}
	//			tNode.TransRep.RetFiels = pack
	//			tNode.Resume()
	//		}), "APIMemberRegisterOrLogin").Start()
	//		return common.ResponseTag_TransactYield, pack
	//	}))

	//API用户加减币
	//WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/api/Game/APIAddSubCoinById", WebAPIHandlerWrapper(
	//	func(tNode *transact.TransNode, params []byte) (int, proto.Message) {
	//		pack := &webapi.SAAddCoinById{}
	//		msg := &webapi.ASAddCoinById{}
	//		err1 := proto.Unmarshal(params, msg)
	//		if err1 != nil {
	//			pack.Tag = webapi.TagCode_FAILED
	//			pack.Msg = "数据序列化失败" + err1.Error()
	//			return common.ResponseTag_ParamError, pack
	//		}
	//
	//		member_snid := msg.GetID()
	//		coin := msg.GetGold()
	//		coinEx := msg.GetGoldEx()
	//		oper := msg.GetOper()
	//		gold_desc := msg.GetDesc()
	//		billNo := int(msg.GetBillNo())
	//		platform := msg.GetPlatform()
	//		//logType := msg.GetLogType()
	//		isAccTodayRecharge := msg.GetIsAccTodayRecharge()
	//		needFlowRate := msg.GetNeedFlowRate()
	//		needGiveFlowRate := msg.GetNeedGiveFlowRate()
	//
	//		if CacheDataMgr.CacheBillCheck(billNo, platform) {
	//			pack.Tag = webapi.TagCode_FAILED
	//			pack.Msg = "Bill number repeated!"
	//			return common.ResponseTag_ParamError, pack
	//		}
	//		CacheDataMgr.CacheBillNumber(billNo, platform) //防止手抖点两下
	//
	//		var err error
	//		var pd *model.PlayerData
	//		oldGold := int64(0)
	//		oldSafeBoxGold := int64(0)
	//		var timeStamp = time.Now().UnixNano()
	//		player := PlayerMgrSington.GetPlayerBySnId(int32(member_snid))
	//		if player != nil { //在线玩家处理
	//			if player.scene != nil {
	//				CacheDataMgr.ClearCacheBill(billNo, platform)
	//				pack.Tag = webapi.TagCode_FAILED
	//				pack.Msg = "Unsupported!!! because player in scene!"
	//				return common.ResponseTag_ParamError, pack
	//			}
	//			pd = player.PlayerData
	//			if len(platform) > 0 && player.Platform != platform {
	//				CacheDataMgr.ClearCacheBill(billNo, platform)
	//				pack.Tag = webapi.TagCode_FAILED
	//				pack.Msg = "player platform forbit!"
	//				return common.ResponseTag_ParamError, pack
	//			}
	//
	//			opcode := int32(common.GainWay_Api_In)
	//
	//			if coin < 0 {
	//				opcode = int32(common.GainWay_Api_Out)
	//				if player.Coin+coin < 0 {
	//					CacheDataMgr.ClearCacheBill(billNo, platform)
	//					pack.Tag = webapi.TagCode_FAILED
	//					pack.Msg = "coin not enough!"
	//					return common.ResponseTag_ParamError, pack
	//				}
	//			}
	//
	//			//if logType != 0 {
	//			//	opcode = logType
	//			//}
	//
	//			oldGold = player.Coin
	//			oldSafeBoxGold = player.SafeBoxCoin
	//			coinLog := model.NewPayCoinLog(int64(billNo), int32(member_snid), coin, opcode,
	//				gold_desc, model.PayCoinLogType_Coin, coinEx)
	//			timeStamp = coinLog.TimeStamp
	//			//增加帐变记录
	//			coinlogex := model.NewCoinLogEx(int32(member_snid), coin+coinEx, oldGold+coin+coinEx,
	//				oldSafeBoxGold, 0, opcode, 0, oper, gold_desc, pd.Platform, pd.Channel,
	//				pd.BeUnderAgentCode, 0, pd.PackageID, 0)
	//
	//			task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
	//				err = model.InsertPayCoinLogs(platform, coinLog)
	//				if err != nil {
	//					logger.Logger.Errorf("model.InsertPayCoinLogs err:%v log:%v", err, coinLog)
	//					return err
	//				}
	//				err = model.InsertCoinLog(coinlogex)
	//				if err != nil {
	//					//回滚到对账日志
	//					model.RemovePayCoinLog(platform, coinLog.LogId)
	//					logger.Logger.Errorf("model.InsertCoinLogs err:%v log:%v", err, coinlogex)
	//					return err
	//				}
	//				return err
	//			}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
	//				CacheDataMgr.ClearCacheBill(billNo, platform)
	//				if data != nil {
	//					pack.Tag = webapi.TagCode_FAILED
	//					pack.Msg = data.(error).Error()
	//				} else {
	//					//player.Coin += coin + coinEx
	//					player.AddCoinAsync(coin+coinEx, common.GainWay_Api_In, oper, gold_desc, true, 0, false)
	//					//增加相应的泥码量
	//					player.AddDirtyCoin(coin, coinEx)
	//					player.SetPayTs(timeStamp)
	//
	//					if player.TodayGameData == nil {
	//						player.TodayGameData = model.NewPlayerGameCtrlData()
	//					}
	//					//actRandCoinMgr.OnPlayerRecharge(player, coin)
	//					if isAccTodayRecharge {
	//
	//						player.AddCoinPayTotal(coin)
	//						player.TodayGameData.RechargeCoin += coin //累加当天充值金额
	//						if coin >= 0 && coinEx >= 0 {
	//							plt := PlatformMgrSington.GetPlatform(pd.Platform)
	//							curVer := int32(0)
	//							if plt != nil {
	//								curVer = plt.ExchangeVer
	//							}
	//							log := model.NewCoinGiveLogEx(pd.SnId, pd.Name, coin, coinEx, 0, opcode, pd.PromoterTree,
	//								model.COINGIVETYPE_PAY, curVer, pd.Platform, pd.Channel, pd.BeUnderAgentCode,
	//								"", "system", pd.PackageID, int32(needFlowRate), int32(needGiveFlowRate))
	//							if log != nil {
	//								err := model.InsertGiveCoinLog(log)
	//								if err == nil {
	//									if pd.LastExchangeOrder != "" && pd.TotalConvertibleFlow > 0 {
	//										err = model.UpdateGiveCoinLastFlow(platform, pd.LastExchangeOrder, pd.TotalConvertibleFlow)
	//									}
	//								}
	//								//清空流水，更新id
	//								pd.TotalConvertibleFlow = 0
	//								pd.LastExchangeOrder = log.LogId.Hex()
	//								if player == nil {
	//									//需要回写数据库
	//									task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
	//										model.UpdatePlayerExchageFlowAndOrder(platform, member_snid, 0, pd.LastExchangeOrder)
	//										return nil
	//									}), nil, "UpdateGiveCoinLogs").StartByExecutor(pd.AccountId)
	//								}
	//							}
	//						}
	//					}
	//
	//					player.dirty = true
	//					player.Time2Save()
	//					if player.scene == nil { //如果在大厅,那么同步下金币
	//						player.SendDiffData()
	//					}
	//					player.SendPlayerRechargeAnswer(coin)
	//					pack.Tag = webapi.TagCode_SUCCESS
	//					pack.Msg = ""
	//				}
	//				tNode.TransRep.RetFiels = pack
	//				tNode.Resume()
	//				if err != nil {
	//					logger.Logger.Error("AddSubCoinById task marshal data error:", err)
	//				}
	//			}), "APIAddCoinById").Start()
	//			return common.ResponseTag_TransactYield, pack
	//		} else {
	//			task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
	//				pd, _ = model.GetPlayerDataBySnId(platform, int32(member_snid), false, true)
	//				if pd == nil {
	//					return errors.New("Player not find.")
	//				}
	//				if len(platform) > 0 && pd.Platform != platform {
	//					return errors.New("player platform forbit.")
	//				}
	//				oldGold = pd.Coin
	//				oldSafeBoxGold = pd.SafeBoxCoin
	//
	//				opcode := int32(common.GainWay_Api_In)
	//				if coin < 0 {
	//					opcode = int32(common.GainWay_Api_Out)
	//					if pd.Coin+coin < 0 {
	//						return errors.New("coin not enough!")
	//					}
	//				}
	//
	//				//if logType != 0 {
	//				//	opcode = logType
	//				//}
	//				coinLog := model.NewPayCoinLog(int64(billNo), int32(member_snid), coin, opcode,
	//					gold_desc, model.PayCoinLogType_Coin, coinEx)
	//				timeStamp = coinLog.TimeStamp
	//				err = model.InsertPayCoinLogs(platform, coinLog)
	//				if err != nil {
	//					logger.Logger.Errorf("model.InsertPayCoinLogs err:%v log:%v", err, coinLog)
	//					return err
	//				}
	//				//增加帐变记录
	//				coinlogex := model.NewCoinLogEx(int32(member_snid), coin+coinEx, oldGold+coin+coinEx,
	//					oldSafeBoxGold, 0, opcode, 0, oper, gold_desc, pd.Platform,
	//					pd.Channel, pd.BeUnderAgentCode, 0, pd.PackageID, 0)
	//				err = model.InsertCoinLog(coinlogex)
	//				if err != nil {
	//					//回滚到对账日志
	//					model.RemovePayCoinLog(platform, coinLog.LogId)
	//					logger.Logger.Errorf("model.InsertCoinLogs err:%v log:%v", err, coinlogex)
	//					return err
	//				}
	//				return err
	//			}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
	//				CacheDataMgr.ClearCacheBill(billNo, platform)
	//				if data != nil {
	//					pack.Tag = webapi.TagCode_FAILED
	//					pack.Msg = data.(error).Error()
	//				} else {
	//					pack.Tag = webapi.TagCode_SUCCESS
	//					pack.Msg = ""
	//					if isAccTodayRecharge && coin >= 0 {
	//						OnPlayerPay(pd, coin)
	//					}
	//					if isAccTodayRecharge && coin >= 0 && coinEx >= 0 {
	//
	//						plt := PlatformMgrSington.GetPlatform(pd.Platform)
	//						curVer := int32(0)
	//						if plt != nil {
	//							curVer = plt.ExchangeVer
	//						}
	//						log := model.NewCoinGiveLogEx(pd.SnId, pd.Name, coin, coinEx, 0, common.GainWay_Api_In, pd.PromoterTree,
	//							model.COINGIVETYPE_PAY, curVer, pd.Platform, pd.Channel, pd.BeUnderAgentCode,
	//							"", "system", pd.PackageID, int32(needFlowRate), int32(needGiveFlowRate))
	//						if log != nil {
	//							err := model.InsertGiveCoinLog(log)
	//							if err == nil {
	//								if pd.LastExchangeOrder != "" && pd.TotalConvertibleFlow > 0 {
	//									err = model.UpdateGiveCoinLastFlow(platform, pd.LastExchangeOrder, pd.TotalConvertibleFlow)
	//								}
	//							}
	//							//清空流水，更新id
	//							pd.TotalConvertibleFlow = 0
	//							pd.LastExchangeOrder = log.LogId.Hex()
	//							if player == nil {
	//								//需要回写数据库
	//								task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
	//									model.UpdatePlayerExchageFlowAndOrder(platform, int32(member_snid), 0, pd.LastExchangeOrder)
	//									return nil
	//								}), nil, "UpdateGiveCoinLogs").StartByExecutor(pd.AccountId)
	//							}
	//						}
	//
	//					}
	//				}
	//
	//				tNode.TransRep.RetFiels = pack
	//				tNode.Resume()
	//				if err != nil {
	//					logger.Logger.Error("AddSubCoinById task marshal data error:", err)
	//				}
	//			}), "APIAddSubCoinById").Start()
	//			return common.ResponseTag_TransactYield, pack
	//		}
	//	}))

	//获取用户金币数量
	//WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/api/Member/GetMemberGoldById", WebAPIHandlerWrapper(
	//	func(tNode *transact.TransNode, params []byte) (int, proto.Message) {
	//		pack := &webapi.SAMemberGold{}
	//		msg := &webapi.ASMemberGold{}
	//
	//		err1 := proto.Unmarshal(params, msg)
	//		if err1 != nil {
	//			pack.Tag = webapi.TagCode_FAILED
	//			pack.Msg = "数据序列化失败" + err1.Error()
	//			return common.ResponseTag_ParamError, pack
	//		}
	//
	//		platform := PlatformMgrSington.GetPlatform(msg.GetPlatform())
	//		if platform == nil {
	//			pack.Tag = webapi.TagCode_FAILED
	//			pack.Msg = "没有对应的包标识"
	//			return common.ResponseTag_ParamError, pack
	//		}
	//		platform_param := msg.GetPlatform()
	//		member_snid := msg.GetSnid()
	//		var err error
	//		gold := int64(0)
	//		bank := int64(0)
	//		p := PlayerMgrSington.GetPlayerBySnId(member_snid)
	//		task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
	//			if p != nil {
	//				if len(platform_param) > 0 && p.Platform != platform_param {
	//					return errors.New("Platform error.")
	//				}
	//				bank = p.SafeBoxCoin
	//				gold = p.GetCoin()
	//				return nil
	//			} else {
	//				pbi, _ := model.GetPlayerDataBySnId(platform_param, member_snid, true, true)
	//				if pbi == nil {
	//					return errors.New("snid error")
	//				}
	//				if len(platform_param) > 0 && pbi.Platform != platform_param {
	//					return errors.New("Platform error.")
	//				}
	//				bank = pbi.SafeBoxCoin
	//				gold = pbi.Coin
	//				return nil
	//			}
	//		}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
	//			if data != nil {
	//				pack.Tag = webapi.TagCode_FAILED
	//				pack.Msg = data.(error).Error()
	//			} else {
	//
	//				pd := &webapi.PlayerCoinData{
	//					Id:   member_snid,
	//					Gold: gold,
	//					Bank: bank,
	//				}
	//				pack.Tag = webapi.TagCode_SUCCESS
	//				//pack.Msg = data.(error).Error()
	//				pack.Data = pd
	//			}
	//			tNode.TransRep.RetFiels = pack
	//			tNode.Resume()
	//			if err != nil {
	//				logger.Logger.Error("AddSubCoinById task marshal data error:", err)
	//			}
	//		}), "GetMemberGoldById").Start()
	//
	//		return common.ResponseTag_TransactYield, pack
	//	}))

	//获取用户注单记录游戏记录
	//WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/api/Member/GetGameHistory", WebAPIHandlerWrapper(
	//	func(tNode *transact.TransNode, params []byte) (int, proto.Message) {
	//		pack := &webapi.SAPlayerHistory{}
	//		msg := &webapi.ASPlayerHistory{}
	//
	//		err1 := proto.Unmarshal(params, msg)
	//		if err1 != nil {
	//			pack.Tag = webapi.TagCode_FAILED
	//			pack.Msg = "数据序列化失败" + err1.Error()
	//			return common.ResponseTag_ParamError, pack
	//		}
	//
	//		platform := PlatformMgrSington.GetPlatform(msg.GetPlatform())
	//		if platform == nil {
	//			pack.Tag = webapi.TagCode_FAILED
	//			pack.Msg = "没有对应的包标识"
	//			return common.ResponseTag_ParamError, pack
	//		}
	//		platform_param := msg.GetPlatform()
	//		member_snid := msg.GetSnid()
	//
	//		gameid := int(msg.GetGameId())
	//		historyModel := msg.GetGameHistoryModel()
	//		p := PlayerMgrSington.GetPlayerBySnId(member_snid)
	//		if p == nil{
	//			pi,_:=model.GetPlayerDataBySnId(platform_param, member_snid, true, true)
	//			p = &Player{PlayerData:pi}
	//		}
	//
	//		if p != nil {
	//			switch historyModel {
	//			case PLAYER_HISTORY_MODEL: // 历史记录
	//				task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
	//					var genPlayerHistoryInfo = func(spinID string, isFree bool, createdTime, totalBetValue, totalPriceValue, totalBonusValue, multiple int64, player *gamehall.PlayerHistoryInfo) {
	//						player.SpinID = proto.String(spinID)
	//						player.CreatedTime = proto.Int64(createdTime)
	//						player.TotalBetValue = proto.Int64(totalBetValue)
	//						player.TotalPriceValue = proto.Int64(totalPriceValue)
	//						player.IsFree = proto.Bool(isFree)
	//						player.TotalBonusValue = proto.Int64(totalBonusValue)
	//						player.Multiple = proto.Int64(multiple)
	//					}
	//
	//					var genPlayerHistoryInfoMsg = func(spinid string, v *model.NeedGameRecord, gdl *model.GameDetailedLog, player *gamehall.PlayerHistoryInfo) {
	//						switch gameid {
	//						case common.GameId_Crash:
	//							data, err := model.UnMarshalGameNoteByHUNDRED(gdl.GameDetailedNote)
	//							if err != nil {
	//								logger.Logger.Errorf("World UnMarshalAvengersGameNote error:%v", err)
	//							}
	//							jsonString, _ := json.Marshal(data)
	//
	//							// convert json to struct
	//							gnd := model.CrashType{}
	//							json.Unmarshal(jsonString, &gnd)
	//
	//							//gnd := data.(*model.CrashType)
	//							for _, curplayer := range gnd.PlayerData {
	//								if curplayer.UserId == p.SnId {
	//									genPlayerHistoryInfo(spinid, false, int64(v.Ts), int64(curplayer.UserBetTotal), curplayer.ChangeCoin, 0, int64(curplayer.UserMultiple), player)
	//									break
	//								}
	//							}
	//						case common.GameId_Avengers:
	//							data, err := model.UnMarshalAvengersGameNote(gdl.GameDetailedNote)
	//							if err != nil {
	//								logger.Logger.Errorf("World UnMarshalAvengersGameNote error:%v", err)
	//							}
	//							gnd := data.(*model.GameResultLog)
	//							genPlayerHistoryInfo(spinid, gnd.BaseResult.IsFree, int64(v.Ts), int64(gnd.BaseResult.TotalBet), gnd.BaseResult.WinTotal, gnd.BaseResult.WinSmallGame, 0, player)
	//						//case common.GameId_EasterIsland:
	//						//	data, err := model.UnMarshalEasterIslandGameNote(gdl.GameDetailedNote)
	//						//	if err != nil {
	//						//		logger.Logger.Errorf("World UnMarshalEasterIslandGameNote error:%v", err)
	//						//	}
	//						//	gnd := data.(*model.EasterIslandType)
	//						//	genPlayerHistoryInfo(spinid, gnd.IsFree, int64(v.Ts), int64(gnd.Score), gnd.TotalPriceValue, gnd.TotalBonusValue, player)
	//						default:
	//							logger.Logger.Errorf("World CSHundredSceneGetGameHistoryInfoHandler receive gameid(%v) error", gameid)
	//						}
	//					}
	//
	//					gameclass := int32(2)
	//					spinid := strconv.FormatInt(int64(p.SnId), 10)
	//					dbGameFrees := srvdata.PBDB_GameFreeMgr.Datas.Arr //.GetData(data.DbGameFree.Id)
	//					roomtype := int32(0)
	//					for _, v := range dbGameFrees {
	//						if int32(gameid) == v.GetGameId() {
	//							gameclass = v.GetGameClass()
	//							roomtype = v.GetSceneType()
	//							break
	//						}
	//					}
	//
	//					gpl := model.GetPlayerListByHallEx(p.SnId, p.Platform, 0, 50, 0, 0, roomtype, gameclass, gameid)
	//					pack := &gamehall.SCPlayerHistory{}
	//					for _, v := range gpl.Data {
	//						if v.GameDetailedLogId == "" {
	//							logger.Logger.Error("World PlayerHistory GameDetailedLogId is nil")
	//							break
	//						}
	//						gdl := model.GetPlayerHistory(p.Platform, v.GameDetailedLogId)
	//						player := &gamehall.PlayerHistoryInfo{}
	//						genPlayerHistoryInfoMsg(spinid, v, gdl, player)
	//						pack.PlayerHistory = append(pack.PlayerHistory, player)
	//					}
	//					proto.SetDefaults(pack)
	//					logger.Logger.Infof("World gameid:%v PlayerHistory:%v ", gameid, pack)
	//					return pack
	//				}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
	//					if data == nil {
	//						tNode.TransRep.RetFiels = data
	//						tNode.Resume()
	//					}
	//				}), "CSGetPlayerHistoryHandlerWorld").Start()
	//			case BIGWIN_HISTORY_MODEL: // 爆奖记录
	//				jackpotList := JackpotListMgrSington.GetJackpotList(gameid)
	//				//if len(jackpotList) < 1 {
	//				//	JackpotListMgrSington.GenJackpot(gameid) // 初始化爆奖记录
	//				//	JackpotListMgrSington.after(gameid)      // 开启定时器
	//				//	jackpotList = JackpotListMgrSington.GetJackpotList(gameid)
	//				//}
	//				pack := JackpotListMgrSington.GetStoCMsg(jackpotList)
	//				pack.GameId = msg.GetGameId()
	//				logger.Logger.Infof("World BigWinHistory: %v %v", gameid, pack)
	//				p.SendToClient(int(gamehall.HundredScenePacketID_PACKET_SC_GAMEBIGWINHISTORY), pack)
	//			case GAME_HISTORY_MODEL:
	//				task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
	//					var genGameHistoryInfo = func(gameNumber string, createdTime, multiple int64, hash string, gamehistory *gamehall.GameHistoryInfo) {
	//						gamehistory.GameNumber = proto.String(gameNumber)
	//						gamehistory.CreatedTime = proto.Int64(createdTime)
	//						gamehistory.Hash = proto.String(hash)
	//						gamehistory.Multiple = proto.Int64(multiple)
	//					}
	//
	//					gls := model.GetAllGameDetailedLogsByGameIdAndTs(p.Platform, gameid, 20)
	//
	//					pack := &gamehall.SCPlayerHistory{}
	//					for _, v := range gls {
	//
	//						gamehistory := &gamehall.GameHistoryInfo{}
	//
	//						data, err := model.UnMarshalGameNoteByHUNDRED(v.GameDetailedNote)
	//						if err != nil {
	//							logger.Logger.Errorf("World UnMarshalAvengersGameNote error:%v", err)
	//						}
	//						jsonString, _ := json.Marshal(data)
	//
	//						// convert json to struct
	//						gnd := model.CrashType{}
	//						json.Unmarshal(jsonString, &gnd)
	//
	//						genGameHistoryInfo(v.LogId, int64(v.Ts), int64(gnd.Rate), gnd.Hash, gamehistory)
	//						pack.GameHistory = append(pack.GameHistory, gamehistory)
	//					}
	//					proto.SetDefaults(pack)
	//					logger.Logger.Infof("World gameid:%v History:%v ", gameid, pack)
	//					return pack
	//				}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
	//					if data == nil {
	//						tNode.TransRep.RetFiels = data
	//						tNode.Resume()
	//					}
	//				}), "CSGetGameHistoryHandlerWorld").Start()
	//			default:
	//				logger.Logger.Errorf("World CSHundredSceneGetGameHistoryInfoHandler receive historyModel(%v) error", historyModel)
	//			}
	//		}
	//
	//		return common.ResponseTag_TransactYield, pack
	//	}))

	//API用户登录
	WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/api/Member/QPAPIRegisterOrLogin", WebAPIHandlerWrapper(
		func(tNode *transact.TransNode, params []byte) (int, proto.Message) {
			logger.Logger.Trace("WebAPIHandler:/api/Member/APIMemberRegisterOrLogin", params)
			pack := &qpapi.SALogin{}
			msg := &qpapi.ASLogin{}
			err1 := proto.Unmarshal(params, msg)
			if err1 != nil {
				pack.Tag = qpapi.TagCode_FAILED
				pack.Msg = "数据序列化失败" + err1.Error()
				return common.ResponseTag_ParamError, pack
			}

			//platformID, channel, _, _, tagkey := PlatformMgrSington.GetPlatformByPackageTag(msg.GetPlatformTag())
			platform := PlatformMgrSington.GetPlatform(msg.GetMerchantTag())
			//platform := PlatformMgrSington.GetPlatform(strconv.Itoa(int(platformID)))
			if platform == nil {
				pack.Tag = qpapi.TagCode_FAILED
				pack.Msg = "没有对应的平台"
				return common.ResponseTag_ParamError, pack
			}
			merchantkey := platform.MerchantKey

			sign := msg.GetSign()

			raw := fmt.Sprintf("%v%v%v%v", msg.GetMerchantTag(), msg.GetUserName(), merchantkey, msg.GetTs())
			h := md5.New()
			io.WriteString(h, raw)
			newsign := hex.EncodeToString(h.Sum(nil))

			if newsign != sign {
				pack.Tag = qpapi.TagCode_FAILED
				pack.Msg = "商户验签失败"
				return common.ResponseTag_ParamError, pack
			}

			var acc *model.Account
			var errcode int

			rawpwd := fmt.Sprintf("%v%v%v", msg.GetSign(), common.GetAppId(), time.Now().Unix())
			hpwd := md5.New()
			io.WriteString(hpwd, rawpwd)
			pwd := hex.EncodeToString(h.Sum(nil))

			task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
				acc, errcode = model.AccountIsExist(msg.GetUserName(), msg.GetUserName(), "", platform.IdStr, time.Now().Unix(), 5, 0, false)
				if errcode == 2 {
					var err int
					acc, err = model.InsertAccount(msg.GetUserName(), pwd, platform.IdStr, "", "", "",
						"qpapi", 0, 0, "", "", "", "", 0)
					if acc != nil { //需要预先创建玩家数据
						_, _ = model.GetPlayerData(acc.Platform, acc.AccountId.Hex())
					}
					if err == 5 {
						return 1
					} else {
						//t.flag = login_proto.OpResultCode_OPRC_Login_CreateAccError
					}
				}
				return errcode
			}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
				switch data.(int) {
				case 1:
					{
						if err := model.UpdatePlayerTokenPassword(acc.Platform, acc.AccountId.Hex(), pwd); err != nil {
							pack.Tag = qpapi.TagCode_FAILED
							pack.Msg = err.Error()
						} else {
							tokenuserdata := &common.TokenUserData{
								TelegramId: msg.GetUserName(),
								Password:   pwd,
								Packagetag: msg.GetMerchantTag(),
								Expired:    time.Now().Add(time.Minute * 15).Unix(),
							}

							token, err := common.CreateTokenAes(tokenuserdata)
							if err != nil {
								pack.Tag = qpapi.TagCode_FAILED
								pack.Msg = err.Error()
							} else {
								pack.Tag = qpapi.TagCode_SUCCESS
								pack.Msg = ""
								pack.Snid = proto.Int32(acc.SnId)
								pack.Token = fmt.Sprintf("%v?token=%v&snid=%v", model.GameParamData.CSURL, token, acc.SnId)
							}
						}
					}
				case 2:
					{
						pack.Tag = qpapi.TagCode_FAILED
						pack.Msg = "创建帐号失败"
					}
				case 3:
					{
						//t.flag = login_proto.OpResultCode_OPRC_LoginPassError
						pack.Tag = qpapi.TagCode_FAILED
						pack.Msg = "账号密码错误"
					}
				case 4:
					{
						//t.flag = login_proto.OpResultCode_OPRC_AccountBeFreeze
						pack.Tag = qpapi.TagCode_FAILED
						pack.Msg = "帐号冻结"
					}
				}
				tNode.TransRep.RetFiels = pack
				tNode.Resume()
			}), "APIMemberRegisterOrLogin").Start()
			return common.ResponseTag_TransactYield, pack
		}))

	//API用户加减币
	WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/api/Game/QPAPIAddSubCoinById", WebAPIHandlerWrapper(
		func(tNode *transact.TransNode, params []byte) (int, proto.Message) {
			pack := &qpapi.SAAddCoinById{}
			msg := &qpapi.ASAddCoinById{}
			err1 := proto.Unmarshal(params, msg)
			if err1 != nil {
				pack.Tag = qpapi.TagCode_FAILED
				pack.Msg = "数据序列化失败" + err1.Error()
				return common.ResponseTag_ParamError, pack
			}

			username := msg.GetUsername()
			coin := msg.GetGold()
			billNo := int(msg.GetBillNo())
			platform := msg.GetMerchantTag()
			curplatform := PlatformMgrSington.GetPlatform(platform)

			if curplatform == nil {
				pack.Tag = qpapi.TagCode_FAILED
				pack.Msg = "没有对应的平台"
				return common.ResponseTag_ParamError, pack
			}
			merchantkey := curplatform.MerchantKey

			sign := msg.GetSign()

			raw := fmt.Sprintf("%v%v%v%v%v%v", username, coin, billNo, platform, merchantkey, msg.GetTs())
			h := md5.New()
			io.WriteString(h, raw)
			newsign := hex.EncodeToString(h.Sum(nil))

			if newsign != sign {
				pack.Tag = qpapi.TagCode_FAILED
				pack.Msg = "商户验签失败"
				return common.ResponseTag_ParamError, pack
			}

			if CacheDataMgr.CacheBillCheck(billNo, platform) {
				pack.Tag = qpapi.TagCode_FAILED
				pack.Msg = "Bill number repeated!"
				return common.ResponseTag_ParamError, pack
			}
			CacheDataMgr.CacheBillNumber(billNo, platform) //防止手抖点两下

			var err error
			var pd *model.PlayerData
			oldGold := int64(0)
			oldSafeBoxGold := int64(0)
			var timeStamp = time.Now().UnixNano()
			acc, accerr := model.GetAccountByName(platform, username)
			if accerr != nil {
				CacheDataMgr.ClearCacheBill(billNo, platform)
				pack.Tag = qpapi.TagCode_FAILED
				pack.Msg = accerr.Error()
				return common.ResponseTag_ParamError, pack
			}
			member_snid := acc.SnId
			player := PlayerMgrSington.GetPlayerBySnId(int32(member_snid))
			if player != nil { //在线玩家处理
				task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
					if player.scene != nil && coin < 0 {
						//player.Kickout(common.KickReason_CheckCodeErr)

						leavemsg := &server.WGPlayerLeave{
							SnId: proto.Int32(player.SnId),
						}
						proto.SetDefaults(leavemsg)
						player.SendToGame(int(server.SSPacketID_PACKET_WG_PlayerLEAVE), leavemsg)

						select {
						case <-player.leavechan:
						case <-time.After(time.Second * 1):
						}

					}
					//player = PlayerMgrSington.GetPlayerBySnId(int32(member_snid))
					if player.scene != nil {
						CacheDataMgr.ClearCacheBill(billNo, platform)
						//pack.Tag = qpapi.TagCode_FAILED
						//pack.Msg = "Unsupported!!! because player in scene!"
						//return common.ResponseTag_ParamError, pack
						return errors.New("Unsupported!!! because player in scene!")
					}
					pd = player.PlayerData
					if len(platform) > 0 && player.Platform != platform {
						CacheDataMgr.ClearCacheBill(billNo, platform)
						//pack.Tag = qpapi.TagCode_FAILED
						//pack.Msg = "player platform forbit!"
						//return common.ResponseTag_ParamError, pack
						return errors.New("player platform forbit!")
					}

					opcode := int32(common.GainWay_Api_In)

					if coin < 0 {
						opcode = int32(common.GainWay_Api_Out)
						if player.Coin+coin < 0 {
							CacheDataMgr.ClearCacheBill(billNo, platform)
							//pack.Tag = qpapi.TagCode_FAILED
							//pack.Msg = "coin not enough!"
							//return common.ResponseTag_ParamError, pack
							return errors.New("coin not enough!")
						}
					}

					//if logType != 0 {
					//	opcode = logType
					//}

					oldGold = player.Coin
					oldSafeBoxGold = player.SafeBoxCoin
					coinLog := model.NewPayCoinLog(int64(billNo), int32(member_snid), coin, opcode,
						"qpsystem", model.PayCoinLogType_Coin, 0)
					timeStamp = coinLog.TimeStamp
					//增加帐变记录
					coinlogex := model.NewCoinLogEx(int32(member_snid), coin, oldGold+coin,
						oldSafeBoxGold, 0, opcode, 0, "oper", "online", pd.Platform, pd.Channel,
						pd.BeUnderAgentCode, 0, pd.PackageID, 0)

					err = model.InsertPayCoinLogs(platform, coinLog)
					if err != nil {
						logger.Logger.Errorf("model.InsertPayCoinLogs err:%v log:%v", err, coinLog)
						return err
					}
					err = model.InsertCoinLog(coinlogex)
					if err != nil {
						//回滚到对账日志
						model.RemovePayCoinLog(platform, coinLog.LogId)
						logger.Logger.Errorf("model.InsertCoinLogs err:%v log:%v", err, coinlogex)
						return err
					}
					return err
				}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
					CacheDataMgr.ClearCacheBill(billNo, platform)
					if data != nil {
						pack.Tag = qpapi.TagCode_FAILED
						pack.Msg = data.(error).Error()
					} else {
						//player.Coin += coin + coinEx
						player.AddCoinAsync(coin, common.GainWay_Api_In, "oper", "Async", true, 0, false)
						//增加相应的泥码量
						player.AddDirtyCoin(coin, 0)
						player.SetPayTs(timeStamp)

						if player.TodayGameData == nil {
							player.TodayGameData = model.NewPlayerGameCtrlData()
						}
						//actRandCoinMgr.OnPlayerRecharge(player, coin)
						/*
							if isAccTodayRecharge {

								player.AddCoinPayTotal(coin)
								player.TodayGameData.RechargeCoin += coin //累加当天充值金额
								if coin >= 0 {
									plt := PlatformMgrSington.GetPlatform(pd.Platform)
									curVer := int32(0)
									if plt != nil {
										curVer = plt.ExchangeVer
									}
									log := model.NewCoinGiveLogEx(pd.SnId, pd.Name, coin, 0, 0, opcode, pd.PromoterTree,
										model.COINGIVETYPE_PAY, curVer, pd.Platform, pd.Channel, pd.BeUnderAgentCode,
										"", "system", pd.PackageID, int32(needFlowRate), int32(needGiveFlowRate))
									if log != nil {
										err := model.InsertGiveCoinLog(log)
										if err == nil {
											if pd.LastExchangeOrder != "" && pd.TotalConvertibleFlow > 0 {
												err = model.UpdateGiveCoinLastFlow(platform, pd.LastExchangeOrder, pd.TotalConvertibleFlow)
											}
										}
										//清空流水，更新id
										pd.TotalConvertibleFlow = 0
										pd.LastExchangeOrder = log.LogId.Hex()
										if player == nil {
											//需要回写数据库
											task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
												model.UpdatePlayerExchageFlowAndOrder(platform, member_snid, 0, pd.LastExchangeOrder)
												return nil
											}), nil, "UpdateGiveCoinLogs").StartByExecutor(pd.AccountId)
										}
									}
								}
							}
						*/
						player.dirty = true
						player.Time2Save()
						if player.scene == nil { //如果在大厅,那么同步下金币
							player.SendDiffData()
						}
						player.SendPlayerRechargeAnswer(coin)
						pack.Tag = qpapi.TagCode_SUCCESS
						pack.Msg = ""
					}
					tNode.TransRep.RetFiels = pack
					tNode.Resume()
					if err != nil {
						logger.Logger.Error("AddSubCoinById task marshal data error:", err)
					}
				}), "APIAddSubCoinById").Start()
				return common.ResponseTag_TransactYield, pack
			} else {
				task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
					pd, _ = model.GetPlayerDataBySnId(platform, int32(member_snid), false, true)
					if pd == nil {
						return errors.New("Player not find.")
					}
					if len(platform) > 0 && pd.Platform != platform {
						return errors.New("player platform forbit.")
					}
					oldGold = pd.Coin
					oldSafeBoxGold = pd.SafeBoxCoin

					opcode := int32(common.GainWay_Api_In)
					if coin < 0 {
						opcode = int32(common.GainWay_Api_Out)
						if pd.Coin+coin < 0 {
							return errors.New("coin not enough!")
						}
					}

					//if logType != 0 {
					//	opcode = logType
					//}
					coinLog := model.NewPayCoinLog(int64(billNo), int32(member_snid), coin, opcode,
						"not online", model.PayCoinLogType_Coin, 0)
					timeStamp = coinLog.TimeStamp
					err = model.InsertPayCoinLogs(platform, coinLog)
					if err != nil {
						logger.Logger.Errorf("model.InsertPayCoinLogs err:%v log:%v", err, coinLog)
						return err
					}
					//增加帐变记录
					coinlogex := model.NewCoinLogEx(int32(member_snid), coin, oldGold+coin,
						oldSafeBoxGold, 0, opcode, 0, "oper", "not online", pd.Platform,
						pd.Channel, pd.BeUnderAgentCode, 0, pd.PackageID, 0)
					err = model.InsertCoinLog(coinlogex)
					if err != nil {
						//回滚到对账日志
						model.RemovePayCoinLog(platform, coinLog.LogId)
						logger.Logger.Errorf("model.InsertCoinLogs err:%v log:%v", err, coinlogex)
						return err
					}
					return err
				}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
					CacheDataMgr.ClearCacheBill(billNo, platform)
					if data != nil {
						pack.Tag = qpapi.TagCode_FAILED
						pack.Msg = data.(error).Error()
					} else {
						pack.Tag = qpapi.TagCode_SUCCESS
						pack.Msg = ""
						/*
							if isAccTodayRecharge && coin >= 0 {
								OnPlayerPay(pd, coin)
							}
							if isAccTodayRecharge && coin >= 0 && coinEx >= 0 {

								plt := PlatformMgrSington.GetPlatform(pd.Platform)
								curVer := int32(0)
								if plt != nil {
									curVer = plt.ExchangeVer
								}
								log := model.NewCoinGiveLogEx(pd.SnId, pd.Name, coin, coinEx, 0, common.GainWay_Api_In, pd.PromoterTree,
									model.COINGIVETYPE_PAY, curVer, pd.Platform, pd.Channel, pd.BeUnderAgentCode,
									"", "system", pd.PackageID, int32(needFlowRate), int32(needGiveFlowRate))
								if log != nil {
									err := model.InsertGiveCoinLog(log)
									if err == nil {
										if pd.LastExchangeOrder != "" && pd.TotalConvertibleFlow > 0 {
											err = model.UpdateGiveCoinLastFlow(platform, pd.LastExchangeOrder, pd.TotalConvertibleFlow)
										}
									}
									//清空流水，更新id
									pd.TotalConvertibleFlow = 0
									pd.LastExchangeOrder = log.LogId.Hex()
									if player == nil {
										//需要回写数据库
										task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
											model.UpdatePlayerExchageFlowAndOrder(platform, int32(member_snid), 0, pd.LastExchangeOrder)
											return nil
										}), nil, "UpdateGiveCoinLogs").StartByExecutor(pd.AccountId)
									}
								}

							}
						*/
					}

					tNode.TransRep.RetFiels = pack
					tNode.Resume()
					if err != nil {
						logger.Logger.Error("AddSubCoinById task marshal data error:", err)
					}
				}), "APIAddSubCoinById").Start()
				return common.ResponseTag_TransactYield, pack
			}
		}))

	//获取用户金币数量
	WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/api/Member/QPGetMemberGoldById", WebAPIHandlerWrapper(
		func(tNode *transact.TransNode, params []byte) (int, proto.Message) {
			pack := &qpapi.SAMemberGold{}
			msg := &qpapi.ASMemberGold{}

			err1 := proto.Unmarshal(params, msg)
			if err1 != nil {
				pack.Tag = qpapi.TagCode_FAILED
				pack.Msg = "数据序列化失败" + err1.Error()
				return common.ResponseTag_ParamError, pack
			}
			username := msg.GetUsername()
			platform := msg.GetMerchantTag()
			curplatform := PlatformMgrSington.GetPlatform(platform)
			//platform := PlatformMgrSington.GetPlatform(strconv.Itoa(int(platformID)))
			if curplatform == nil {
				pack.Tag = qpapi.TagCode_FAILED
				pack.Msg = "没有对应的平台"
				return common.ResponseTag_ParamError, pack
			}
			merchantkey := curplatform.MerchantKey

			sign := msg.GetSign()

			raw := fmt.Sprintf("%v%v%v%v", username, platform, merchantkey, msg.GetTs())
			h := md5.New()
			io.WriteString(h, raw)
			newsign := hex.EncodeToString(h.Sum(nil))

			if newsign != sign {
				pack.Tag = qpapi.TagCode_FAILED
				pack.Msg = "商户验签失败"
				return common.ResponseTag_ParamError, pack
			}

			//platform := PlatformMgrSington.GetPlatform(msg.GetPlatform())
			//if platform == nil {
			//	pack.Tag = qpapi.TagCode_FAILED
			//	pack.Msg = "没有对应的包标识"
			//	return common.ResponseTag_ParamError, pack
			//}
			//platform_param := msg.GetPlatform()
			//member_snid := msg.GetSnid()

			var err error
			gold := int64(0)
			bank := int64(0)

			task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
				acc, accerr := model.GetAccountByName(platform, msg.GetUsername())
				if accerr != nil {
					//pack.Tag = qpapi.TagCode_FAILED
					//pack.Msg = accerr.Error()
					//return common.ResponseTag_ParamError, pack
					return accerr
				}
				member_snid := acc.SnId
				p := PlayerMgrSington.GetPlayerBySnId(member_snid)
				if p != nil {
					if len(platform) > 0 && p.Platform != platform {
						return errors.New("Platform error.")
					}
					bank = p.SafeBoxCoin
					gold = p.GetCoin()
					return nil
				} else {
					pbi, _ := model.GetPlayerDataBySnId(platform, member_snid, true, true)
					if pbi == nil {
						return errors.New("snid error")
					}
					if len(platform) > 0 && pbi.Platform != platform {
						return errors.New("Platform error.")
					}
					bank = pbi.SafeBoxCoin
					gold = pbi.Coin
					return nil
				}
			}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
				if data != nil {
					pack.Tag = qpapi.TagCode_FAILED
					pack.Msg = data.(error).Error()
				} else {

					pd := &qpapi.PlayerCoinData{
						Username: msg.GetUsername(),
						Gold:     gold,
						Bank:     bank,
					}
					pack.Tag = qpapi.TagCode_SUCCESS
					//pack.Msg = data.(error).Error()
					pack.Data = pd
				}
				tNode.TransRep.RetFiels = pack
				tNode.Resume()
				if err != nil {
					logger.Logger.Error("AddSubCoinById task marshal data error:", err)
				}
			}), "GetMemberGoldById").Start()

			return common.ResponseTag_TransactYield, pack
		}))

	//获取用户注单记录游戏记录
	WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/api/Member/QPGetGameHistory", WebAPIHandlerWrapper(
		func(tNode *transact.TransNode, params []byte) (int, proto.Message) {
			pack := &qpapi.SAPlayerHistory{}
			msg := &qpapi.ASPlayerHistory{}

			err1 := proto.Unmarshal(params, msg)
			if err1 != nil {
				pack.Tag = qpapi.TagCode_FAILED
				pack.Msg = "数据序列化失败" + err1.Error()
				return common.ResponseTag_ParamError, pack
			}

			historyModel := msg.GetGameHistoryModel()
			username := msg.GetUsername()
			platform := msg.GetMerchantTag()
			curplatform := PlatformMgrSington.GetPlatform(platform)
			starttime := msg.GetStartTime()
			endtime := msg.GetEndTime()
			pageno := msg.GetPageNo()
			pagesize := msg.GetPageSize()
			if pagesize == 0 {
				pagesize = 50
			}
			//platform := PlatformMgrSington.GetPlatform(strconv.Itoa(int(platformID)))
			if curplatform == nil {
				pack.Tag = qpapi.TagCode_FAILED
				pack.Msg = "没有对应的平台"
				return common.ResponseTag_ParamError, pack
			}
			merchantkey := curplatform.MerchantKey

			sign := msg.GetSign()

			raw := fmt.Sprintf("%v%v%v%v%v%v%v%v%v", username, platform, historyModel, starttime, endtime, pageno, pagesize, merchantkey, msg.GetTs())
			h := md5.New()
			io.WriteString(h, raw)
			newsign := hex.EncodeToString(h.Sum(nil))

			if newsign != sign {
				pack.Tag = qpapi.TagCode_FAILED
				pack.Msg = "商户验签失败"
				return common.ResponseTag_ParamError, pack
			}

			searchsnid := int32(0)
			if username != "" {
				acc, accerr := model.GetAccountByName(platform, username)
				if accerr != nil {
					pack.Tag = qpapi.TagCode_FAILED
					pack.Msg = accerr.Error()
					return common.ResponseTag_ParamError, pack
				}
				member_snid := acc.SnId

				p := PlayerMgrSington.GetPlayerBySnId(member_snid)
				if p == nil {
					pi, _ := model.GetPlayerDataBySnId(platform, member_snid, true, true)
					p = &Player{PlayerData: pi}
				}
				searchsnid = p.SnId
			}

			//if p != nil {

			//gameid := 112
			switch int(historyModel) {
			case PLAYER_HISTORY_MODEL: // 历史记录
				task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
					var genPlayerHistoryInfo = func(logid string, gameid int32, spinID, username string, isFree bool, createdTime, totalBetValue, totalPriceValue, totalBonusValue, multiple int64, player *qpapi.PlayerHistoryInfo) {
						player.SpinID = proto.String(spinID)
						player.CreatedTime = proto.Int64(createdTime)
						player.TotalBetValue = proto.Int64(totalBetValue)
						player.TotalPriceValue = proto.Int64(totalPriceValue)
						player.IsFree = proto.Bool(isFree)
						player.TotalBonusValue = proto.Int64(totalBonusValue)
						player.Multiple = proto.Int64(multiple)
						player.Gameid = proto.Int32(gameid)
						player.Logid = proto.String(logid)
						player.UserName = proto.String(username)
					}

					var genPlayerHistoryInfoMsg = func(v *model.NeedGameRecord, gdl *model.GameDetailedLog, player *qpapi.PlayerHistoryInfo) {
						switch gdl.GameId {
						case common.GameId_Crash:
							data, err := model.UnMarshalGameNoteByHUNDRED(gdl.GameDetailedNote)
							if err != nil {
								logger.Logger.Errorf("World UnMarshalAvengersGameNote error:%v", err)
							}
							jsonString, _ := json.Marshal(data)

							// convert json to struct
							gnd := model.CrashType{}
							json.Unmarshal(jsonString, &gnd)

							//gnd := data.(*model.CrashType)
							for _, curplayer := range gnd.PlayerData {
								if curplayer.UserId == v.SnId {
									genPlayerHistoryInfo(gdl.LogId, gdl.GameId, strconv.FormatInt(int64(curplayer.UserId), 10), v.Username, false, int64(v.Ts), int64(curplayer.UserBetTotal), curplayer.ChangeCoin, 0, int64(curplayer.UserMultiple), player)
									break
								}
							}
						default:
							logger.Logger.Errorf("World CSHundredSceneGetGameHistoryInfoHandler receive gameid(%v) error", gdl.GameId)
						}
					}

					gpl := model.GetPlayerListByHallExAPI(searchsnid, curplatform.IdStr, starttime, endtime, int(pageno), int(pagesize))
					//pack := &gamehall.SCPlayerHistory{}
					for _, v := range gpl.Data {
						if v.GameDetailedLogId == "" {
							logger.Logger.Error("World PlayerHistory GameDetailedLogId is nil")
							break
						}
						gdl := model.GetPlayerHistory(curplatform.IdStr, v.GameDetailedLogId)
						player := &qpapi.PlayerHistoryInfo{}
						genPlayerHistoryInfoMsg(v, gdl, player)
						pack.PlayerHistory = append(pack.PlayerHistory, player)
					}
					pack.PageNo = proto.Int32(int32(gpl.PageNo))
					pack.PageSize = proto.Int32(int32(gpl.PageSize))
					pack.PageSum = proto.Int32(int32(gpl.PageSum))
					proto.SetDefaults(pack)
					//logger.Logger.Infof("World gameid:%v PlayerHistory:%v ", gdl.GameId, pack)
					return pack
				}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
					if data != nil {
						pack.Tag = qpapi.TagCode_SUCCESS
						tNode.TransRep.RetFiels = data
						tNode.Resume()
					}
				}), "CSGetPlayerHistoryHandlerWorld").Start()
				return common.ResponseTag_TransactYield, pack
			case GAME_HISTORY_MODEL:

				task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
					var genGameHistoryInfo = func(gameNumber string, createdTime, multiple int64, hash string, gamehistory *qpapi.GameHistoryInfo) {
						gamehistory.GameNumber = proto.String(gameNumber)
						gamehistory.CreatedTime = proto.Int64(createdTime)
						gamehistory.Hash = proto.String(hash)
						gamehistory.Multiple = proto.Int64(multiple)
					}

					gls := model.GetPlayerHistoryAPI(searchsnid, curplatform.IdStr, starttime, endtime, int(pageno), int(pagesize))

					//pack := &gamehall.SCPlayerHistory{}
					for _, v := range gls.Data {

						gamehistory := &qpapi.GameHistoryInfo{}

						data, err := model.UnMarshalGameNoteByHUNDRED(v.GameDetailedNote)
						if err != nil {
							logger.Logger.Errorf("World UnMarshalAvengersGameNote error:%v", err)
						}
						jsonString, _ := json.Marshal(data)

						// convert json to struct
						gnd := model.CrashType{}
						json.Unmarshal(jsonString, &gnd)

						genGameHistoryInfo(v.LogId, int64(v.Ts), int64(gnd.Rate), gnd.Hash, gamehistory)
						pack.GameHistory = append(pack.GameHistory, gamehistory)
					}
					proto.SetDefaults(pack)
					//logger.Logger.Infof("World gameid:%v History:%v ", gameid, pack)
					return pack
				}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
					if data != nil {
						pack.Tag = qpapi.TagCode_SUCCESS
						tNode.TransRep.RetFiels = data
						tNode.Resume()
					}
				}), "CSGetGameHistoryHandlerWorld").Start()
				return common.ResponseTag_TransactYield, pack

			default:
				pack.Tag = qpapi.TagCode_FAILED
				pack.Msg = "GameHistoryModel 有误"
				return common.ResponseTag_ParamError, pack
				//logger.Logger.Errorf("World CSHundredSceneGetGameHistoryInfoHandler receive historyModel(%v) error", historyModel)
			}
			//}
		}))

	//Cresh游戏Hash校验
	WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/api/Game/CrashVerifier", WebAPIHandlerWrapper(
		func(tNode *transact.TransNode, params []byte) (int, proto.Message) {
			logger.Logger.Trace("WebAPIHandler:/api/Game/CrashVerifier", params)
			resp := &webapi.SACrachHash{}
			msg := &webapi.ASCrachHash{}
			err := proto.Unmarshal(params, msg)
			if err != nil {
				resp.Tag = webapi.TagCode_FAILED
				resp.Msg = "数据序列化失败"
				return common.ResponseTag_ParamError, resp
			}

			hash := msg.GetHash()
			wheel := msg.GetWheel()
			if hash != "" {
				result := crash.HashToMultiple(hash, int(wheel))
				resp.Tag = webapi.TagCode_SUCCESS
				resp.Msg = ""
				resp.Multiple = result
				return common.ResponseTag_Ok, resp
			} else {
				resp.Tag = webapi.TagCode_FAILED
				resp.Msg = "Hash错误"
				return common.ResponseTag_ParamError, resp
			}
		}))

	//----------------------------------传输数据改成protobuf结构---------------------------------------------------
	//后台修改平台数据后推送给游服
	WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/game_srv/platform_config", WebAPIHandlerWrapper(
		func(tNode *transact.TransNode, params []byte) (int, proto.Message) {
			resp := &webapi_proto.SAUpdatePlatform{}
			msg := &webapi_proto.ASUpdatePlatform{}
			err := proto.Unmarshal(params, msg)
			if err != nil {
				resp.Tag = webapi_proto.TagCode_FAILED
				resp.Msg = "数据序列化失败"
				return common.ResponseTag_ParamError, resp
			}

			platforms := msg.GetPlatforms()
			if platforms == nil {
				logger.Logger.Error("Get platform data error")
				resp.Tag = webapi_proto.TagCode_FAILED
				resp.Msg = "Get platform data error!"
				return common.ResponseTag_ParamError, resp
			} else {
				for _, ptf := range platforms {
					platformName := ptf.GetPlatformName()
					isolated := ptf.GetIsolated()
					disable := ptf.GetDisabled()
					id := ptf.GetId()
					url := ptf.GetCustomService()
					bindOption := ptf.GetBindOption()
					serviceFlag := ptf.GetServiceFlag()
					upgradeAccountGiveCoin := ptf.GetUpgradeAccountGiveCoin()
					newAccountGiveCoin := ptf.GetNewAccountGiveCoin()
					perBankNoLimitAccount := ptf.GetPerBankNoLimitAccount()
					exchangeMin := ptf.GetExchangeMin()
					exchangeLimit := ptf.GetExchangeLimit()
					exchangeTax := ptf.GetExchangeTax()
					exchangeFlow := ptf.GetExchangeFlow()
					exchangeFlag := ptf.GetExchangeFlag()
					spreadConfig := ptf.GetSpreadConfig()
					vipRange := ptf.GetVipRange()
					otherParams := ""    //未传递
					ccf := &ClubConfig{} //未传递
					verifyCodeType := ptf.GetVerifyCodeType()
					thirdGameMerchant := ptf.GetThirdGameMerchant()
					ths := make(map[int32]int32)
					for _, v := range thirdGameMerchant {
						ths[v.Id] = v.Merchant
					}
					customType := ptf.GetCustomType()
					needDeviceInfo := false //未传递
					needSameName := ptf.GetNeedSameName()
					exchangeForceTax := ptf.GetExchangeForceTax()
					exchangeGiveFlow := ptf.GetExchangeGiveFlow()
					exchangeVer := ptf.GetExchangeVer()
					exchangeBankMax := ptf.GetExchangeBankMax()
					exchangeAlipayMax := ptf.GetExchangeAlipayMax()
					dbHboCfg := 0 //未传递
					PerBankNoLimitName := ptf.GetPerBankNoLimitName()
					IsCanUserBindPromoter := ptf.GetIsCanUserBindPromoter()
					UserBindPromoterPrize := ptf.GetUserBindPromoterPrize()
					SpreadWinLose := false //未传递
					exchangeMultiple := ptf.GetExchangeMultiple()
					registerVerifyCodeSwitch := false //未传递
					merchantKey := ptf.GetMerchantKey()

					platform := PlatformMgrSington.UpsertPlatform(platformName, isolated, disable, id, url,
						int32(bindOption), serviceFlag, int32(upgradeAccountGiveCoin), int32(newAccountGiveCoin),
						int32(perBankNoLimitAccount), int32(exchangeMin), int32(exchangeLimit), int32(exchangeTax),
						int32(exchangeFlow), int32(exchangeFlag), int32(spreadConfig), vipRange, otherParams, ccf,
						int32(verifyCodeType), ths, int32(customType), needDeviceInfo, needSameName, int32(exchangeForceTax),
						int32(exchangeGiveFlow), int32(exchangeVer), int32(exchangeBankMax), int32(exchangeAlipayMax),
						int32(dbHboCfg), int32(PerBankNoLimitName), IsCanUserBindPromoter, int32(UserBindPromoterPrize),
						SpreadWinLose, int32(exchangeMultiple), registerVerifyCodeSwitch, merchantKey)

					if platform != nil {
						//通知客户端
						scPlatForm := &login_proto.SCPlatFormConfig{
							Platform:               platform.IdStr,
							OpRetCode:              login_proto.OpResultCode_OPRC_Sucess,
							UpgradeAccountGiveCoin: upgradeAccountGiveCoin,
							ExchangeMin:            exchangeMin,
							ExchangeLimit:          exchangeLimit,
							VipRange:               vipRange,
							OtherParams:            otherParams,
							SpreadConfig:           spreadConfig,
							ExchangeTax:            platform.ExchangeTax,
							ExchangeFlow:           platform.ExchangeFlow,
							ExchangeBankMax:        platform.ExchangeBankMax,
							ExchangeAlipayMax:      platform.ExchangeAlipayMax,
							ExchangeMultiple:       platform.ExchangeMultiple,
						}

						proto.SetDefaults(scPlatForm)
						PlayerMgrSington.BroadcastMessageToPlatform(platform.IdStr, int(login_proto.LoginPacketID_PACKET_SC_PLATFORMCFG), scPlatForm)
					}
				}

				resp.Tag = webapi_proto.TagCode_SUCCESS
				resp.Msg = ""
				return common.ResponseTag_Ok, resp
			}
			//return common.ResponseTag_ParamError, resp
		}))
	//======= 全局游戏开关 =======
	WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/game_srv/update_global_game_status", WebAPIHandlerWrapper(
		func(tNode *transact.TransNode, params []byte) (int, proto.Message) {
			pack := &webapi_proto.SAUpdateGameConfigGlobal{}
			msg := &webapi_proto.ASUpdateGameConfigGlobal{}
			err := proto.Unmarshal(params, msg)
			if err != nil {
				pack.Tag = webapi_proto.TagCode_FAILED
				pack.Msg = fmt.Sprintf("err:%v", err.Error())
				return common.ResponseTag_ParamError, pack
			}

			gcGlobal := msg.GetGameStatus()
			gameStatus := gcGlobal.GetGameStatus()
			if len(gameStatus) == 0 {
				pack.Tag = webapi_proto.TagCode_FAILED
				pack.Msg = "Get game status data error"
				return common.ResponseTag_ParamError, pack
			} else {
				//更改所有游戏平台下，游戏的状态
				for _, s := range gameStatus {
					gameId := s.GetGameId()
					status := s.GetStatus()
					PlatformMgrSington.GameStatus[gameId] = status
				}
				pack.Tag = webapi_proto.TagCode_SUCCESS
				pack.Msg = ""
				return common.ResponseTag_Ok, pack
			}
			//pack.Tag = webapi_proto.TagCode_FAILED
			//pack.Msg = "error"
			//return common.ResponseTag_ParamError, pack
		}))
	//======= 更新游戏配置 =======
	WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/game_srv/update_game_configs", WebAPIHandlerWrapper(
		func(tNode *transact.TransNode, params []byte) (int, proto.Message) {
			pack := &webapi_proto.SAUpdateGameConfig{}
			msg := &webapi_proto.ASUpdateGameConfig{}
			err := proto.Unmarshal(params, msg)
			if err != nil {
				pack.Tag = webapi_proto.TagCode_FAILED
				pack.Msg = fmt.Sprintf("err:%v", err.Error())
				return common.ResponseTag_ParamError, pack
			}

			platformConfig := msg.GetConfig()
			platformId := int(platformConfig.GetPlatformId())
			platform := PlatformMgrSington.GetPlatform(strconv.Itoa(platformId))
			dbGameFrees := platformConfig.GetDbGameFrees()
			if len(dbGameFrees) == 0 {
				pack.Tag = webapi_proto.TagCode_FAILED
				pack.Msg = "Not have DbGameFrees"
				return common.ResponseTag_ParamError, pack
			} else {
				//更改所有游戏平台下，游戏的状态
				for _, v := range dbGameFrees {
					dbGameFree := v.GetDbGameFree()
					logicId := dbGameFree.GetId()
					platform.PltGameCfg.games[logicId] = v
				}
				platform.PltGameCfg.RecreateCache()
				pack.Tag = webapi_proto.TagCode_SUCCESS
				pack.Msg = ""
				return common.ResponseTag_Ok, pack
			}
			//pack.Tag = webapi_proto.TagCode_FAILED
			//pack.Msg = "error"
			//return common.ResponseTag_ParamError, pack
		}))
	//======= 更新游戏分组配置 =======
	WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/game_srv/game_config_group", WebAPIHandlerWrapper(
		func(tNode *transact.TransNode, params []byte) (int, proto.Message) {
			pack := &webapi_proto.SAUpdateGameConfigGroup{}
			msg := &webapi_proto.ASUpdateGameConfigGroup{}
			err := proto.Unmarshal(params, msg)
			if err != nil {
				pack.Tag = webapi_proto.TagCode_FAILED
				pack.Msg = fmt.Sprintf("err:%v", err.Error())
				return common.ResponseTag_ParamError, pack
			}

			value := msg.GetGameConfigGroup()
			PlatformGameGroupMgrSington.UpsertGameGroup(value)
			logger.Logger.Trace("PlatformGameGroup data:", value)
			pack.Tag = webapi_proto.TagCode_SUCCESS
			pack.Msg = ""
			return common.ResponseTag_Ok, pack
		}))

	//----------------------------------------------------------------------------------------------------------
	//加币加钻石
	WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/api/Game/AddCoinByIdAndPT", WebAPIHandlerWrapper(
		func(tNode *transact.TransNode, params []byte) (int, proto.Message) {
			pack := &webapi_proto.SAAddCoinByIdAndPT{}
			msg := &webapi_proto.ASAddCoinByIdAndPT{}
			err1 := proto.Unmarshal(params, msg)
			if err1 != nil {
				pack.Tag = webapi_proto.TagCode_FAILED
				pack.Msg = "数据序列化失败" + err1.Error()
				return common.ResponseTag_ParamError, pack
			}
			member_snid := msg.GetID()
			billNo := int(msg.GetBillNo())
			platform := msg.GetPlatform()

			if CacheDataMgr.CacheBillCheck(billNo, platform) {
				pack.Tag = webapi_proto.TagCode_FAILED
				pack.Msg = "Bill number repeated!"
				return common.ResponseTag_ParamError, pack
			}
			CacheDataMgr.CacheBillNumber(billNo, platform) //防止手抖点两下
			player := PlayerMgrSington.GetPlayerBySnId(member_snid)

			var addcoin, diamond int64 = msg.GetGold(), 0
			var logtype = int32(common.GainWay_API_AddCoin)
			if msg.GetLogType() == 1 {
				addcoin = 0
				diamond = msg.GetGold()
			}
			//玩家在线
			if player != nil {
				//玩家在游戏内
				if player.scene != nil {
					CacheDataMgr.ClearCacheBill(billNo, platform)
					pack.Tag = webapi_proto.TagCode_FAILED
					pack.Msg = "Unsupported!!! because player in scene!"
					return common.ResponseTag_ParamError, pack
				}
				if len(platform) > 0 && player.Platform != platform {
					CacheDataMgr.ClearCacheBill(billNo, platform)
					pack.Tag = webapi_proto.TagCode_FAILED
					pack.Msg = "player platform forbit!"
					return common.ResponseTag_ParamError, pack
				}
				if player.Coin+addcoin < 0 || player.Diamond+diamond < 0 {
					CacheDataMgr.ClearCacheBill(billNo, platform)
					pack.Tag = webapi_proto.TagCode_FAILED
					pack.Msg = "coin not enough!"
					return common.ResponseTag_ParamError, pack
				}
				//增加帐变记录
				coinlogex := model.NewCoinLogDiamondEx(member_snid, msg.GetGold(), player.Coin+addcoin, player.SafeBoxCoin,
					diamond+player.Diamond, 0, logtype, 0, msg.GetOper(), msg.GetDesc(), player.Platform, player.Channel,
					player.BeUnderAgentCode, msg.GetLogType(), player.PackageID, 0)
				task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
					err := model.InsertCoinLog(coinlogex)
					if err != nil {
						//回滚到对账日志
						model.RemoveCoinLogOne(platform, coinlogex.LogId)
						logger.Logger.Errorf("model.InsertCoinLogs err:%v log:%v", err, coinlogex)
						return err
					}
					return nil
				}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
					CacheDataMgr.ClearCacheBill(billNo, platform)
					pack.Tag = webapi_proto.TagCode_SUCCESS
					pack.Msg = "success."
					if data != nil {
						pack.Tag = webapi_proto.TagCode_FAILED
						pack.Msg = data.(error).Error()
					} else {
						player.Coin += addcoin
						player.Diamond += diamond
						player.SendDiffData()
					}
					tNode.TransRep.RetFiels = pack
					tNode.Resume()
				}), "AddCoinByIdAndPT").Start()
			} else {
				//玩家不在线
				task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
					findPlayer, _ := model.GetPlayerDataBySnId(platform, int32(member_snid), false, true)
					if findPlayer != nil {
						if len(platform) > 0 && findPlayer.Platform != platform {
							CacheDataMgr.ClearCacheBill(billNo, platform)
							pack.Msg = "player platform forbit!"
							return nil
						}
						if findPlayer.Coin+addcoin < 0 || findPlayer.Diamond+diamond < 0 {
							CacheDataMgr.ClearCacheBill(billNo, platform)
							pack.Msg = "coin not enough!"
							return nil
						}
						//增加帐变记录
						coinlogex := model.NewCoinLogDiamondEx(member_snid, msg.GetGold(), findPlayer.Coin+addcoin, findPlayer.SafeBoxCoin,
							diamond+findPlayer.Diamond, 0, logtype, 0, msg.GetOper(), msg.GetDesc(), findPlayer.Platform, findPlayer.Channel,
							findPlayer.BeUnderAgentCode, msg.GetLogType(), findPlayer.PackageID, 0)

						err := model.UpdatePlayerCoin(findPlayer.Platform, findPlayer.SnId, findPlayer.Coin+addcoin,
							findPlayer.Diamond+diamond, findPlayer.SafeBoxCoin, findPlayer.CoinPayTs, findPlayer.SafeBoxCoinTs)
						if err != nil {
							logger.Logger.Errorf("model.UpdatePlayerCoin err:%v.", err)
							return nil
						}
						//账变记录
						err = model.InsertCoinLog(coinlogex)
						if err != nil {
							//回滚到对账日志
							model.RemoveCoinLogOne(platform, coinlogex.LogId)
							logger.Logger.Errorf("model.InsertCoinLogs err:%v log:%v", err, coinlogex)
							return err
						}
					}
					return findPlayer
				}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
					if data == nil {
						pack.Tag = webapi_proto.TagCode_FAILED
						if pack.Msg == "" {
							pack.Msg = "Player not find."
						}
					} else {
						pack.Tag = webapi_proto.TagCode_SUCCESS
						pack.Msg = "success."
					}
					tNode.TransRep.RetFiels = pack
					tNode.Resume()
				}), "AddCoinByIdAndPT").Start()

			}

			return common.ResponseTag_TransactYield, pack
		}))

	// //-------------------------------------------------------------------------------------------------------
	// //钱包操作接口
	WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/api/Game/AddCoinById", WebAPIHandlerWrapper(
		func(tNode *transact.TransNode, params []byte) (int, proto.Message) {
			pack := &webapi_proto.SAAddCoinById{}
			msg := &webapi_proto.ASAddCoinById{}
			err1 := proto.Unmarshal(params, msg)
			if err1 != nil {
				pack.Tag = webapi_proto.TagCode_FAILED
				pack.Msg = "数据序列化失败" + err1.Error()
				return common.ResponseTag_ParamError, pack
			}

			member_snid := msg.GetID()
			coin := msg.GetGold()
			coinEx := msg.GetGoldEx()
			oper := msg.GetOper()
			gold_desc := msg.GetDesc()
			billNo := int(msg.GetBillNo())
			platform := msg.GetPlatform()
			logType := msg.GetLogType()
			isAccTodayRecharge := msg.GetIsAccTodayRecharge()
			needFlowRate := msg.GetNeedFlowRate()
			needGiveFlowRate := msg.GetNeedGiveFlowRate()

			if CacheDataMgr.CacheBillCheck(billNo, platform) {
				pack.Tag = webapi_proto.TagCode_FAILED
				pack.Msg = "Bill number repeated!"
				return common.ResponseTag_ParamError, pack
			}
			CacheDataMgr.CacheBillNumber(billNo, platform) //防止手抖点两下

			var err error
			var pd *model.PlayerData
			oldGold := int64(0)
			oldSafeBoxGold := int64(0)
			var timeStamp = time.Now().UnixNano()
			player := PlayerMgrSington.GetPlayerBySnId(int32(member_snid))
			if player != nil { //在线玩家处理
				if player.scene != nil {
					CacheDataMgr.ClearCacheBill(billNo, platform)
					pack.Tag = webapi_proto.TagCode_FAILED
					pack.Msg = "Unsupported!!! because player in scene!"
					return common.ResponseTag_ParamError, pack
				}
				pd = player.PlayerData
				if len(platform) > 0 && player.Platform != platform {
					CacheDataMgr.ClearCacheBill(billNo, platform)
					pack.Tag = webapi_proto.TagCode_FAILED
					pack.Msg = "player platform forbit!"
					return common.ResponseTag_ParamError, pack
				}

				if coin < 0 {
					if player.Coin+coin < 0 {
						CacheDataMgr.ClearCacheBill(billNo, platform)
						pack.Tag = webapi_proto.TagCode_FAILED
						pack.Msg = "coin not enough!"
						return common.ResponseTag_ParamError, pack
					}
				}

				opcode := int32(common.GainWay_API_AddCoin)
				if logType != 0 {
					opcode = logType
				}

				oldGold = player.Coin
				oldSafeBoxGold = player.SafeBoxCoin
				coinLog := model.NewPayCoinLog(int64(billNo), int32(member_snid), coin, opcode,
					gold_desc, model.PayCoinLogType_Coin, coinEx)
				timeStamp = coinLog.TimeStamp
				//增加帐变记录
				coinlogex := model.NewCoinLogEx(int32(member_snid), coin+coinEx, oldGold+coin+coinEx,
					oldSafeBoxGold, 0, opcode, 0, oper, gold_desc, pd.Platform, pd.Channel,
					pd.BeUnderAgentCode, 0, pd.PackageID, 0)

				task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
					err = model.InsertPayCoinLogs(platform, coinLog)
					if err != nil {
						logger.Logger.Errorf("model.InsertPayCoinLogs err:%v log:%v", err, coinLog)
						return err
					}
					err = model.InsertCoinLog(coinlogex)
					if err != nil {
						//回滚到对账日志
						model.RemovePayCoinLog(platform, coinLog.LogId)
						logger.Logger.Errorf("model.InsertCoinLogs err:%v log:%v", err, coinlogex)
						return err
					}
					return err
				}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
					CacheDataMgr.ClearCacheBill(billNo, platform)
					if data != nil {
						pack.Tag = webapi_proto.TagCode_FAILED
						pack.Msg = data.(error).Error()
					} else {
						//player.Coin += coin + coinEx
						player.AddCoinAsync(coin+coinEx, common.GainWay_API_AddCoin, oper, gold_desc, true, 0, false)
						//增加相应的泥码量
						player.AddDirtyCoin(coin, coinEx)
						player.SetPayTs(timeStamp)

						if player.TodayGameData == nil {
							player.TodayGameData = model.NewPlayerGameCtrlData()
						}
						//actRandCoinMgr.OnPlayerRecharge(player, coin)
						if isAccTodayRecharge {

							player.AddCoinPayTotal(coin)
							player.TodayGameData.RechargeCoin += coin //累加当天充值金额
							if coin >= 0 && coinEx >= 0 {
								plt := PlatformMgrSington.GetPlatform(pd.Platform)
								curVer := int32(0)
								if plt != nil {
									curVer = plt.ExchangeVer
								}
								log := model.NewCoinGiveLogEx(pd.SnId, pd.Name, coin, coinEx, 0, opcode, pd.PromoterTree,
									model.COINGIVETYPE_PAY, curVer, pd.Platform, pd.Channel, pd.BeUnderAgentCode,
									"", "system", pd.PackageID, int32(needFlowRate), int32(needGiveFlowRate))
								if log != nil {
									err := model.InsertGiveCoinLog(log)
									if err == nil {
										if pd.LastExchangeOrder != "" && pd.TotalConvertibleFlow > 0 {
											err = model.UpdateGiveCoinLastFlow(platform, pd.LastExchangeOrder, pd.TotalConvertibleFlow)
										}
									}
									//清空流水，更新id
									pd.TotalConvertibleFlow = 0
									pd.LastExchangeOrder = log.LogId.Hex()
									if player == nil {
										//需要回写数据库
										task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
											model.UpdatePlayerExchageFlowAndOrder(platform, member_snid, 0, pd.LastExchangeOrder)
											return nil
										}), nil, "UpdateGiveCoinLogs").StartByExecutor(pd.AccountId)
									}
								}
							}
						}

						player.dirty = true
						player.Time2Save()
						if player.scene == nil { //如果在大厅,那么同步下金币
							player.SendDiffData()
						}
						player.SendPlayerRechargeAnswer(coin)
						pack.Tag = webapi_proto.TagCode_SUCCESS
						pack.Msg = ""
					}
					tNode.TransRep.RetFiels = pack
					tNode.Resume()
					if err != nil {
						logger.Logger.Error("AddCoinById task marshal data error:", err)
					}
				}), "AddCoinById").Start()
				return common.ResponseTag_TransactYield, pack
			} else {
				task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
					pd, _ = model.GetPlayerDataBySnId(platform, int32(member_snid), false, true)
					if pd == nil {
						return errors.New("Player not find.")
					}
					if len(platform) > 0 && pd.Platform != platform {
						return errors.New("player platform forbit.")
					}
					oldGold = pd.Coin
					oldSafeBoxGold = pd.SafeBoxCoin

					if coin < 0 {
						if pd.Coin+coin < 0 {
							return errors.New("coin not enough!")
						}
					}

					opcode := int32(common.GainWay_API_AddCoin)
					if logType != 0 {
						opcode = logType
					}
					coinLog := model.NewPayCoinLog(int64(billNo), int32(member_snid), coin, opcode,
						gold_desc, model.PayCoinLogType_Coin, coinEx)
					timeStamp = coinLog.TimeStamp
					err = model.InsertPayCoinLogs(platform, coinLog)
					if err != nil {
						logger.Logger.Errorf("model.InsertPayCoinLogs err:%v log:%v", err, coinLog)
						return err
					}
					//增加帐变记录
					coinlogex := model.NewCoinLogEx(int32(member_snid), coin+coinEx, oldGold+coin+coinEx,
						oldSafeBoxGold, 0, opcode, 0, oper, gold_desc, pd.Platform,
						pd.Channel, pd.BeUnderAgentCode, 0, pd.PackageID, 0)
					err = model.InsertCoinLog(coinlogex)
					if err != nil {
						//回滚到对账日志
						model.RemovePayCoinLog(platform, coinLog.LogId)
						logger.Logger.Errorf("model.InsertCoinLogs err:%v log:%v", err, coinlogex)
						return err
					}
					return err
				}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
					CacheDataMgr.ClearCacheBill(billNo, platform)
					if data != nil {
						pack.Tag = webapi_proto.TagCode_FAILED
						pack.Msg = data.(error).Error()
					} else {
						pack.Tag = webapi_proto.TagCode_SUCCESS
						pack.Msg = ""
						if isAccTodayRecharge && coin >= 0 {
							OnPlayerPay(pd, coin)
						}
						if isAccTodayRecharge && coin >= 0 && coinEx >= 0 {

							plt := PlatformMgrSington.GetPlatform(pd.Platform)
							curVer := int32(0)
							if plt != nil {
								curVer = plt.ExchangeVer
							}
							log := model.NewCoinGiveLogEx(pd.SnId, pd.Name, coin, coinEx, 0, common.GainWay_API_AddCoin, pd.PromoterTree,
								model.COINGIVETYPE_PAY, curVer, pd.Platform, pd.Channel, pd.BeUnderAgentCode,
								"", "system", pd.PackageID, int32(needFlowRate), int32(needGiveFlowRate))
							if log != nil {
								err := model.InsertGiveCoinLog(log)
								if err == nil {
									if pd.LastExchangeOrder != "" && pd.TotalConvertibleFlow > 0 {
										err = model.UpdateGiveCoinLastFlow(platform, pd.LastExchangeOrder, pd.TotalConvertibleFlow)
									}
								}
								//清空流水，更新id
								pd.TotalConvertibleFlow = 0
								pd.LastExchangeOrder = log.LogId.Hex()
								if player == nil {
									//需要回写数据库
									task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
										model.UpdatePlayerExchageFlowAndOrder(platform, int32(member_snid), 0, pd.LastExchangeOrder)
										return nil
									}), nil, "UpdateGiveCoinLogs").StartByExecutor(pd.AccountId)
								}
							}

						}
					}

					tNode.TransRep.RetFiels = pack
					tNode.Resume()
					if err != nil {
						logger.Logger.Error("AddCoinById task marshal data error:", err)
					}
				}), "AddCoinById").Start()
				return common.ResponseTag_TransactYield, pack
			}
		}))
	//重置水池
	WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/api/Game/ResetGamePool", WebAPIHandlerWrapper(
		func(tNode *transact.TransNode, params []byte) (int, proto.Message) {
			pack := &webapi_proto.SAResetGamePool{}
			msg := &webapi_proto.ASResetGamePool{}
			err := proto.Unmarshal(params, msg)
			if err != nil {
				pack.Tag = webapi_proto.TagCode_FAILED
				pack.Msg = "数据序列化失败" + err.Error()
				return common.ResponseTag_ParamError, pack
			}
			if msg.GameFreeId == 0 {
				pack.Tag = webapi_proto.TagCode_FAILED
				pack.Msg = "GameFreeId id zero"
				return common.ResponseTag_ParamError, pack
			}
			gs := GameSessMgrSington.GetGameSess(int(msg.ServerId))
			if gs != nil {
				msg := &server.WGResetCoinPool{
					Platform:   proto.String(msg.GetPlatform()),
					GameFreeId: proto.Int32(msg.GameFreeId),
					ServerId:   proto.Int32(msg.ServerId),
					GroupId:    proto.Int32(msg.GroupId),
					PoolType:   proto.Int32(msg.PoolType),
					Value:      proto.Int32(msg.Value),
				}
				proto.SetDefaults(msg)
				gs.Send(int(server.SSPacketID_PACKET_WG_RESETCOINPOOL), msg)
			} else {
				pack.Tag = webapi_proto.TagCode_FAILED
				pack.Msg = "no find srvId"
				return common.ResponseTag_ParamError, pack
			}
			pack.Tag = webapi_proto.TagCode_SUCCESS
			pack.Msg = "reset game pool success"
			return common.ResponseTag_Ok, pack
		}))

	//更新水池
	WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/api/Game/UpdateGamePool", WebAPIHandlerWrapper(
		func(tNode *transact.TransNode, params []byte) (int, proto.Message) {
			pack := &webapi_proto.SAUpdateGamePool{}
			msg := &webapi_proto.ASUpdateGamePool{}
			err := proto.Unmarshal(params, msg)
			if err != nil {
				pack.Tag = webapi_proto.TagCode_FAILED
				pack.Msg = "数据序列化失败" + err.Error()
				return common.ResponseTag_ParamError, pack
			}
			coinPoolSetting := msg.GetCoinPoolSetting()
			if coinPoolSetting == nil {
				pack.Tag = webapi_proto.TagCode_FAILED
				pack.Msg = "GetCoinPoolSetting is nil"
				return common.ResponseTag_ParamError, pack
			}
			var old *model.CoinPoolSetting
			cps := model.GetCoinPoolSetting(msg.GetGameFreeId(), msg.GetServerId(), msg.GetGroupId(), msg.GetPlatform())
			if cps == nil {
				pack.Tag = webapi_proto.TagCode_FAILED
				pack.Msg = "CoinPoolSettingDatas is nil"
				return common.ResponseTag_NoData, pack
			}
			temp := *cps
			old = &temp
			cps.InitValue = coinPoolSetting.GetInitValue()
			cps.LowerLimit = coinPoolSetting.GetLowerLimit()
			cps.UpperLimit = coinPoolSetting.GetUpperLimit()
			cps.UpperOffsetLimit = coinPoolSetting.GetUpperOffsetLimit()
			cps.MaxOutValue = coinPoolSetting.GetMaxOutValue()
			cps.ChangeRate = coinPoolSetting.GetChangeRate()
			cps.MinOutPlayerNum = coinPoolSetting.GetMinOutPlayerNum()
			cps.BaseRate = coinPoolSetting.GetBaseRate()
			cps.CtroRate = coinPoolSetting.GetCtroRate()
			cps.HardTimeMin = coinPoolSetting.GetHardTimeMin()
			cps.HardTimeMax = coinPoolSetting.GetHardTimeMax()
			cps.NormalTimeMin = coinPoolSetting.GetNormalTimeMin()
			cps.NormalTimeMax = coinPoolSetting.GetNormalTimeMax()
			cps.EasyTimeMin = coinPoolSetting.GetEasyTimeMin()
			cps.EasyTimeMax = coinPoolSetting.GetEasyTimeMax()
			cps.EasrierTimeMin = coinPoolSetting.GetEasrierTimeMin()
			cps.EasrierTimeMax = coinPoolSetting.GetEasrierTimeMax()
			cps.CpCangeType = coinPoolSetting.GetCpCangeType()
			cps.CpChangeInterval = coinPoolSetting.GetCpChangeInterval()
			cps.CpChangeTotle = coinPoolSetting.GetCpChangeTotle()
			cps.CpChangeLower = coinPoolSetting.GetCpChangeLower()
			cps.CpChangeUpper = coinPoolSetting.GetCpChangeUpper()
			cps.ProfitRate = coinPoolSetting.GetProfitRate()
			cps.CoinPoolMode = coinPoolSetting.GetCoinPoolMode()
			cps.ResetTime = coinPoolSetting.GetResetTime()
			task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
				return model.UpsertCoinPoolSetting(cps, old)
			}), task.CompleteNotifyWrapper(func(data interface{}, task task.Task) {
				if data != nil {
					logger.Logger.Errorf("model.UpsertCoinPoolSetting cps:%v err:%v", cps, data)
				}
			}), "UpsertCoinPoolSetting").Start()
			gs := GameSessMgrSington.GetGameSess(int(cps.ServerId))
			if gs != nil {
				var msg = &webapi_proto.CoinPoolSetting{
					Platform:         proto.String(cps.Platform),
					GameFreeId:       proto.Int32(cps.GameFreeId),
					ServerId:         proto.Int32(cps.ServerId),
					GroupId:          proto.Int32(cps.GroupId),
					InitValue:        proto.Int32(cps.InitValue),
					LowerLimit:       proto.Int32(cps.LowerLimit),
					UpperLimit:       proto.Int32(cps.UpperLimit),
					UpperOffsetLimit: proto.Int32(cps.UpperOffsetLimit),
					MaxOutValue:      proto.Int32(cps.MaxOutValue),
					ChangeRate:       proto.Int32(cps.ChangeRate),
					MinOutPlayerNum:  proto.Int32(cps.MinOutPlayerNum),
					BaseRate:         proto.Int32(cps.BaseRate),
					CtroRate:         proto.Int32(cps.CtroRate),
					HardTimeMin:      proto.Int32(cps.HardTimeMin),
					HardTimeMax:      proto.Int32(cps.HardTimeMax),
					NormalTimeMin:    proto.Int32(cps.NormalTimeMin),
					NormalTimeMax:    proto.Int32(cps.NormalTimeMax),
					EasyTimeMin:      proto.Int32(cps.EasyTimeMin),
					EasyTimeMax:      proto.Int32(cps.EasyTimeMax),
					EasrierTimeMin:   proto.Int32(cps.EasrierTimeMin),
					EasrierTimeMax:   proto.Int32(cps.EasrierTimeMax),
					CpCangeType:      proto.Int32(cps.CpCangeType),
					CpChangeInterval: proto.Int32(cps.CpChangeInterval),
					CpChangeTotle:    proto.Int32(cps.CpChangeTotle),
					CpChangeLower:    proto.Int32(cps.CpChangeLower),
					CpChangeUpper:    proto.Int32(cps.CpChangeUpper),
					ProfitRate:       proto.Int32(cps.ProfitRate),
					CoinPoolMode:     proto.Int32(cps.CoinPoolMode),
					ResetTime:        proto.Int32(cps.ResetTime),
				} //ProfitControlMgrSington.FillCoinPoolSetting(msg)
				proto.SetDefaults(msg)
				gs.Send(int(server.SSPacketID_PACKET_WG_COINPOOLSETTING), msg)
			}
			pack.Tag = webapi_proto.TagCode_SUCCESS
			pack.Msg = "update game pool success"
			return common.ResponseTag_Ok, pack
		}))

	//查询水池gameid
	WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/api/Game/QueryGamePoolByGameId", WebAPIHandlerWrapper(
		func(tNode *transact.TransNode, params []byte) (int, proto.Message) {
			pack := &webapi_proto.SAQueryGamePoolByGameId{}
			msg := &webapi_proto.ASQueryGamePoolByGameId{}
			err := proto.Unmarshal(params, msg)
			if err != nil {
				pack.Tag = webapi_proto.TagCode_FAILED
				pack.Msg = "数据序列化失败" + err.Error()
				return common.ResponseTag_ParamError, pack
			}
			StartQueryCoinPoolTransact(tNode, msg.GetGameId(), msg.GetGameMode(), msg.GetPlatform(), msg.GetGroupId())
			return common.ResponseTag_TransactYield, nil
		}))

	//查询水池
	//WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/api/Game/QueryAllGamePool", WebAPIHandlerWrapper(
	//	func(tNode *transact.TransNode, params []byte) (int, proto.Message) {
	//		pack := &webapi_proto.SAQueryAllGamePool{}
	//		msg := &webapi_proto.ASQueryAllGamePool{}
	//		err := proto.Unmarshal(params, msg)
	//		if err != nil {
	//			pack.Tag = webapi_proto.TagCode_FAILED
	//			pack.Msg = "数据序列化失败" + err.Error()
	//			return common.ResponseTag_ParamError, pack
	//		}
	//		pageNo := msg.GetPageNo()
	//		pageSize := msg.GetPageSize()
	//		StartQueryCoinPoolStatesTransact(tNode, int32(pageNo), int32(pageSize))
	//		return common.ResponseTag_TransactYield, nil
	//	}))
	//////////////////////////房间//////////////////////////
	WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/api/Cache/GetRoom", WebAPIHandlerWrapper(
		func(tNode *transact.TransNode, params []byte) (int, proto.Message) {
			pack := &webapi_proto.SAGetRoom{}
			msg := &webapi_proto.ASGetRoom{}
			err := proto.Unmarshal(params, msg)
			if err != nil {
				pack.Tag = webapi_proto.TagCode_FAILED
				pack.Msg = "数据序列化失败" + err.Error()
				return common.ResponseTag_ParamError, pack
			}
			roomId := int(msg.GetSceneId())
			snid := msg.GetSnId()
			pack.Tag = webapi_proto.TagCode_SUCCESS
			if roomId == 0 {
				if snid != 0 {
					p := PlayerMgrSington.GetPlayerBySnId(snid)
					if p != nil && p.scene != nil {
						roomId = p.scene.sceneId
					}
				}
			}
			s := SceneMgrSington.GetScene(roomId)
			if s == nil {
				pack.Msg = "no found"
				return common.ResponseTag_NoData, pack
			} else {
				si := &webapi_proto.RoomInfo{
					Platform:   s.limitPlatform.Name,
					SceneId:    int32(s.sceneId),
					GameId:     int32(s.gameId),
					GameMode:   int32(s.gameMode),
					SceneMode:  int32(s.sceneMode),
					GroupId:    s.groupId,
					GameFreeId: s.paramsEx[0],
					Creator:    s.creator,
					Agentor:    s.agentor,
					ReplayCode: s.replayCode,
					Params:     s.params,
					PlayerCnt:  int32(len(s.players) - s.robotNum),
					RobotCnt:   int32(s.robotNum),
					CreateTime: s.createTime.Unix(),
					ClubId:     s.ClubId,
				}
				if s.starting {
					si.Start = 1
				} else {
					si.Start = 0
				}
				if s.gameSess != nil {
					si.SrvId = s.gameSess.GetSrvId()
				}
				cnt := 0
				total := len(s.players)
				robots := []int32{}
				//优先显示玩家
				for id, p := range s.players {
					if !p.IsRob || total < 10 {
						si.PlayerIds = append(si.PlayerIds, id)
						cnt++
					} else {
						robots = append(robots, id)
					}
					if cnt > 10 {
						break
					}
				}
				//不够再显示机器人
				if total > cnt && cnt < 10 && len(robots) != 0 {
					for i := 0; cnt < 10 && i < len(robots); i++ {
						si.PlayerIds = append(si.PlayerIds, robots[i])
						cnt++
						if cnt > 10 {
							break
						}
					}
				}
				pack.RoomInfo = si
			}
			return common.ResponseTag_Ok, pack
		}))
	WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/api/Cache/ListRoom", WebAPIHandlerWrapper(
		func(tNode *transact.TransNode, params []byte) (int, proto.Message) {
			pack := &webapi_proto.SAListRoom{}
			msg := &webapi_proto.ASListRoom{}
			err := proto.Unmarshal(params, msg)
			if err != nil {
				pack.Tag = webapi_proto.TagCode_FAILED
				pack.Msg = "数据序列化失败" + err.Error()
				return common.ResponseTag_ParamError, pack
			}

			pageNo := msg.GetPageNo()
			pageSize := msg.GetPageSize()
			//数据校验
			if pageNo == 0 {
				pageNo = 1
			}
			if pageSize == 0 {
				pageSize = 20
			}
			pack.Tag = webapi_proto.TagCode_SUCCESS
			start := (pageNo - 1) * pageSize
			end := pageNo * pageSize
			roomList, count, roomSum := SceneMgrSington.MarshalAllRoom(msg.GetPlatform(), int(msg.GetGroupId()), int(msg.GetGameId()),
				int(msg.GetGameMode()), int(msg.GetClubId()), int(msg.GetRoomType()), int(msg.GetSceneId()), msg.GamefreeId, msg.GetSnId(), start, end, pageSize)
			if count < pageNo {
				pageNo = 1
			}
			pack.RoomInfo = roomList
			pack.PageCount = count
			pack.PageNo = pageNo
			pack.TotalList = roomSum
			return common.ResponseTag_Ok, pack
		}))
	WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/api/Cache/DestroyRoom", WebAPIHandlerWrapper(
		func(tNode *transact.TransNode, params []byte) (int, proto.Message) {
			pack := &webapi_proto.SADestroyRoom{}
			msg := &webapi_proto.ASDestroyRoom{}
			err := proto.Unmarshal(params, msg)
			if err != nil {
				pack.Tag = webapi_proto.TagCode_FAILED
				pack.Msg = "数据序列化失败" + err.Error()
				return common.ResponseTag_ParamError, pack
			}
			pack.Tag = webapi_proto.TagCode_SUCCESS
			platform := msg.GetPlatform()
			if len(platform) == 0 {
				pack.Tag = webapi_proto.TagCode_FAILED
				pack.Msg = "the platform is nil"
				return common.ResponseTag_ParamError, pack
			}
			switch msg.DestroyType {
			case 1: //删除所有空房间
				for _, s := range SceneMgrSington.scenes {
					if !s.isPlatform(platform) {
						continue
					}
					if s != nil && !s.deleting && len(s.players) == 0 {
						logger.Logger.Warnf("WebService SpecailEmptySceneId destroyroom scene:%v", s.sceneId)
						s.ForceDelete(false)
					}
				}
			case 2: //删除所有未开始的房间
				for _, s := range SceneMgrSington.scenes {
					if !s.isPlatform(platform) {
						continue
					}
					if s != nil && !s.deleting && !s.starting && !s.IsHundredScene() {
						logger.Logger.Warnf("WebService SpecailUnstartSceneId destroyroom scene:%v", s.sceneId)
						s.ForceDelete(false)
					}
				}
			default: //删除指定房间
				if len(msg.SceneIds) == 0 {
					pack.Tag = webapi_proto.TagCode_FAILED
					pack.Msg = "the sceneid is nil"
					return common.ResponseTag_NoFindRoom, pack
				}
				for _, sceneId := range msg.GetSceneIds() {
					s := SceneMgrSington.GetScene(int(sceneId))
					if s == nil {
						pack.Tag = webapi_proto.TagCode_FAILED
						pack.Msg = "the sceneid is nil"
						return common.ResponseTag_NoFindRoom, pack
					}
					if !s.isPlatform(platform) {
						pack.Tag = webapi_proto.TagCode_FAILED
						pack.Msg = "the sceneid is not ower platform"
						return common.ResponseTag_NoFindRoom, pack
					}
					logger.Logger.Warnf("WebService destroyroom scene:%v", s.sceneId)
					s.ForceDelete(false)
				}
			}
			return common.ResponseTag_Ok, pack
		}))
	///////////////////////////玩家信息///////////////////////////
	WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/api/Player/PlayerData", WebAPIHandlerWrapper(
		func(tNode *transact.TransNode, params []byte) (int, proto.Message) {
			pack := &webapi_proto.SAGetPlayerData{}
			msg := &webapi_proto.ASGetPlayerData{}
			err := proto.Unmarshal(params, msg)
			if err != nil {
				pack.Tag = webapi_proto.TagCode_FAILED
				pack.Msg = "数据序列化失败" + err.Error()
				return common.ResponseTag_ParamError, pack
			}
			id := msg.GetID()
			platform := msg.GetPlatform()
			player := PlayerMgrSington.GetPlayerBySnId(id)
			if player == nil {
				task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
					data, _ := model.GetPlayerDataBySnId(platform, id, true, false)
					return data
				}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
					playerData, ok := data.(*model.PlayerData)
					if !ok || playerData == nil {
						pack.Msg = "no find player"
						pack.Tag = webapi_proto.TagCode_FAILED
					} else {
						pack.Tag = webapi_proto.TagCode_SUCCESS
						pdfw := model.ConvertPlayerDataToWebData(playerData)
						pdfw.Online = false
						pack.PlayerData = pdfw
					}
					tNode.TransRep.RetFiels = pack
					tNode.Resume()
				}), "PlayerData").Start()
				return common.ResponseTag_TransactYield, pack
			} else {
				if len(platform) > 0 && player.Platform == platform {
					pack.Tag = webapi_proto.TagCode_SUCCESS
					pwdf := model.ConvertPlayerDataToWebData(player.PlayerData)
					pwdf.Online = player.IsOnLine()
					if pwdf != nil {
						if player.scene != nil && player.scene.sceneId != common.DgSceneId && !player.scene.IsTestScene() {
							pwdf.Coin = player.sceneCoin
						}
					}
					pack.PlayerData = pwdf
				} else {
					pack.Tag = webapi_proto.TagCode_FAILED
					pack.Msg = "no find player"
				}
				return common.ResponseTag_Ok, pack
			}
		}))
	//多个玩家数据
	WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/api/Player/MorePlayerData", WebAPIHandlerWrapper(
		func(tNode *transact.TransNode, params []byte) (int, proto.Message) {
			pack := &webapi_proto.SAMorePlayerData{}
			msg := &webapi_proto.ASMorePlayerData{}
			err := proto.Unmarshal(params, msg)
			pack.Tag = webapi_proto.TagCode_FAILED
			if err != nil {
				pack.Msg = "数据序列化失败" + err.Error()
				return common.ResponseTag_ParamError, pack
			}
			if len(msg.GetSnIds()) == 0 {
				pack.Msg = "IDs value error!"
				return common.ResponseTag_ParamError, pack
			}
			if len(msg.GetSnIds()) > model.GameParamData.MorePlayerLimit {
				pack.Msg = "IDs too more error!"
				return common.ResponseTag_ParamError, pack
			}
			pack.Tag = webapi_proto.TagCode_SUCCESS
			var leavePlayer []int32
			for _, snid := range msg.GetSnIds() {
				player := PlayerMgrSington.GetPlayerBySnId(snid)
				if player != nil {
					pwdf := model.ConvertPlayerDataToWebData(player.PlayerData)
					if pwdf != nil {
						pwdf.Online = true
						if player.scene != nil && player.scene.sceneId != common.DgSceneId {
							pwdf.Coin = player.sceneCoin
						}
						pack.PlayerData = append(pack.PlayerData, pwdf)
					}
				} else {
					leavePlayer = append(leavePlayer, snid)
				}
			}
			if len(leavePlayer) == 0 {
				return common.ResponseTag_Ok, pack
			}
			task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
				return model.GetPlayerDatasBySnIds(msg.Platform, leavePlayer, false)
			}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
				if data != nil {
					playerDatas, _ := data.([]*model.PlayerData)
					for _, v := range playerDatas {
						pdfw := model.ConvertPlayerDataToWebData(v)
						if pdfw != nil {
							pdfw.Online = false
							pack.PlayerData = append(pack.PlayerData, pdfw)
						}
					}
				}
				tNode.TransRep.RetFiels = pack
				tNode.Resume()
			}), "PlayerData").Start()
			return common.ResponseTag_TransactYield, pack
		}))
	//更新玩家数据
	WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/api/Cache/UpdatePlayerElement", WebAPIHandlerWrapper(
		func(tNode *transact.TransNode, params []byte) (int, proto.Message) {
			pack := &webapi_proto.SAUpdatePlayerElement{}
			msg := &webapi_proto.ASUpdatePlayerElement{}
			uerr := proto.Unmarshal(params, msg)
			pack.Tag = webapi_proto.TagCode_FAILED
			if uerr != nil {
				pack.Msg = "数据序列化失败" + uerr.Error()
				return common.ResponseTag_ParamError, pack
			}
			//简单校验接收到的数据
			if msg.SnId == 0 || len(msg.GetPlayerEleArgs()) == 0 {
				pack.Msg = "error: ID or ElementList is nil"
				return common.ResponseTag_Ok, pack
			}

			playerMap := make(map[string]interface{})
			for _, v := range msg.GetPlayerEleArgs() {
				playerMap[v.Key] = v.Val
			}
			listBan := [...]string{"Id", "SnId", "AccountId"}
			for i := 0; i < len(listBan); i++ {
				if _, exist := playerMap[listBan[i]]; exist {
					delete(playerMap, listBan[i])
				}
			}
			if len(playerMap) == 0 {
				pack.Msg = "no any data"
				return common.ResponseTag_Ok, pack
			}
			pack.Tag = webapi_proto.TagCode_SUCCESS
			player := PlayerMgrSington.GetPlayerBySnId(msg.SnId)
			if player != nil {
				if len(msg.Platform) > 0 && player.Platform != msg.Platform {
					pack.Tag = webapi_proto.TagCode_FAILED
					pack.Msg = "Platform error."
					return common.ResponseTag_NoFindUser, pack
				} else {
					//玩家在线
					//获取玩家数据
					var pd *model.PlayerData
					pd = player.PlayerData
					pd = SetInfo(pd, *pd, playerMap)
					player.dirty = true
					player.SendDiffData()
					var alipayAcc, alipayAccName, bankAccount, bankAccName string
					if val, ok := playerMap["AlipayAccount"]; ok {
						if str, ok := val.(string); ok {
							alipayAcc = str
						}
					}
					if val, ok := playerMap["AlipayAccName"]; ok {
						if str, ok := val.(string); ok {
							alipayAccName = str
						}
					}
					if val, ok := playerMap["BankAccount"]; ok {
						if str, ok := val.(string); ok {
							bankAccount = str
						}
					}
					if val, ok := playerMap["BankAccName"]; ok {
						if str, ok := val.(string); ok {
							bankAccName = str
						}
					}
					if alipayAcc != "" || alipayAccName != "" {
						task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
							model.NewBankBindLog(msg.SnId, msg.Platform, model.BankBindLogType_Ali,
								alipayAccName, alipayAcc, 2)
							return nil
						}), nil, "NewBankBindLog").Start()
					}
					if bankAccount != "" || bankAccName != "" {
						task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
							model.NewBankBindLog(msg.SnId, msg.Platform, model.BankBindLogType_Bank,
								bankAccName, bankAccount, 2)
							return nil
						}), nil, "NewBankBindLog").Start()
					}
					pack.Msg = "success"
					return common.ResponseTag_Ok, pack
				}
			} else {
				task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
					//玩家不在线
					//先检测是否有玩家
					if len(msg.Platform) > 0 {
						pbi := model.GetPlayerBaseInfo(msg.Platform, msg.SnId)
						if pbi == nil {
							return errors.New("no player")
						}
					}
					//直接操作数据库
					err := model.UpdatePlayerElement(msg.Platform, msg.SnId, playerMap)
					if err == nil {
						var alipayAcc, alipayAccName, bankAccount, bankAccName string
						if val, ok := playerMap["AlipayAccount"]; ok {
							if str, ok := val.(string); ok {
								alipayAcc = str
							}
						}
						if val, ok := playerMap["AlipayAccName"]; ok {
							if str, ok := val.(string); ok {
								alipayAccName = str
							}
						}
						if val, ok := playerMap["BankAccount"]; ok {
							if str, ok := val.(string); ok {
								bankAccount = str
							}
						}
						if val, ok := playerMap["BankAccName"]; ok {
							if str, ok := val.(string); ok {
								bankAccName = str
							}
						}
						if alipayAcc != "" || alipayAccName != "" {
							model.NewBankBindLog(msg.SnId, msg.Platform, model.BankBindLogType_Ali,
								alipayAccName, alipayAcc, 2)
						}
						if bankAccount != "" || bankAccName != "" {
							model.NewBankBindLog(msg.SnId, msg.Platform, model.BankBindLogType_Bank,
								bankAccName, bankAccount, 2)
						}
					} else if err != nil {
						logger.Logger.Error("UpdatePlayerElement task marshal data error:", err)
					}
					return err
				}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
					if data == nil {
						pack.Msg = "success"
					} else {
						pack.Tag = webapi_proto.TagCode_FAILED
						pack.Msg = fmt.Sprintf("%v", data)
					}
					tNode.TransRep.RetFiels = pack
					tNode.Resume()
				}), "UpdatePlayerElement").Start()
				return common.ResponseTag_TransactYield, pack
			}
		}))
	WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/api/Player/KickPlayer", WebAPIHandlerWrapper(
		func(tNode *transact.TransNode, params []byte) (int, proto.Message) {
			pack := &webapi_proto.SAKickPlayer{}
			msg := &webapi_proto.ASKickPlayer{}
			err := proto.Unmarshal(params, msg)
			pack.Tag = webapi_proto.TagCode_FAILED
			if err != nil {
				pack.Msg = "数据序列化失败" + err.Error()
				return common.ResponseTag_ParamError, pack
			}
			if msg.SnId == 0 {
				pack.Msg = "snid is zero"
				return common.ResponseTag_ParamError, pack
			}
			player := PlayerMgrSington.GetPlayerBySnId(msg.SnId)
			if player != nil {
				if msg.Platform == "" || msg.Platform != player.Platform {
					pack.Msg = "platform is dif"
					return common.ResponseTag_ParamError, pack
				}
				sid := player.sid
				endTime := time.Now().Unix() + int64(time.Minute.Seconds())*int64(msg.Minute)
				ls := LoginStateMgrSington.GetLoginStateOfSid(sid)
				if ls != nil {
					if ls.als != nil && ls.als.acc != nil {
						ls.als.acc.State = endTime
					}
				}
				task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
					return model.FreezeAccount(msg.Platform, player.AccountId, int(msg.Minute))
				}), nil, "FreezeAccount").Start()
				player.Kickout(common.KickReason_Freeze)
				pack.Tag = webapi_proto.TagCode_SUCCESS
				pack.Msg = "success"
			} else {
				pack.Msg = "the player not online or no player."
			}
			return common.ResponseTag_Ok, pack
		}))
	WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/api/Report/QueryOnlineReportList", WebAPIHandlerWrapper(
		func(tNode *transact.TransNode, params []byte) (int, proto.Message) {
			pack := &webapi_proto.SAQueryOnlineReportList{}
			msg := &webapi_proto.ASQueryOnlineReportList{}
			err := proto.Unmarshal(params, msg)
			if err != nil {
				pack.Tag = webapi_proto.TagCode_FAILED
				pack.Msg = "数据序列化失败" + err.Error()
				return common.ResponseTag_ParamError, pack
			}
			platform := msg.Platform
			pageNo := msg.PageNo
			pageSize := msg.PageSize
			orderColumn := msg.OrderColumn
			orderType := msg.OrderType

			start := (pageNo - 1) * pageSize
			end := pageNo * pageSize
			count := len(PlayerMgrSington.players)
			if count < int(start) || count == 0 {
				pack.Tag = webapi_proto.TagCode_SUCCESS
				return common.ResponseTag_Ok, pack
			}
			var players []*Player
			for _, p := range PlayerMgrSington.players {
				if !p.IsOnLine() {
					continue
				}
				if msg.GameFreeId == 0 {
					if msg.GameId != 0 && (p.scene == nil || p.scene.gameId != int(msg.GameId)) {
						continue
					}
				} else {
					if p.scene == nil || p.scene.dbGameFree.GetId() != msg.GameFreeId {
						continue
					}
				}
				if len(platform) != 0 && p.Platform != platform {
					continue
				}
				players = append(players, p)
			}
			// 排序
			var sortFunc func(i, j int) bool
			switch orderColumn {
			case common.OrderColumnCoinPayTotal:
				sortFunc = func(i, j int) bool {
					if players[i].CoinPayTotal == players[j].CoinPayTotal {
						return false
					}
					return players[i].CoinPayTotal > players[j].CoinPayTotal
				}
			case common.OrderColumnCoinExchangeTotal:
				sortFunc = func(i, j int) bool {
					if players[i].CoinExchangeTotal == players[j].CoinExchangeTotal {
						return false
					}
					return players[i].CoinExchangeTotal > players[j].CoinExchangeTotal
				}
			case common.OrderColumnTaxTotal:
				sortFunc = func(i, j int) bool {
					if players[i].GameTax == players[j].GameTax {
						return false
					}
					return players[i].GameTax > players[j].GameTax
				}
			case common.OrderColumnRegisterTime:
				sortFunc = func(i, j int) bool {
					if players[i].CreateTime.Equal(players[j].CreateTime) {
						return false
					}
					return players[i].CreateTime.After(players[j].CreateTime)
				}
			case common.OrderColumnRoomNumber:
				sortFunc = func(i, j int) bool {
					a, b := players[i], players[j]
					if a.scene == nil && b.scene == nil {
						return false
					}
					if a.scene != nil && b.scene == nil {
						return true
					}
					if a.scene == nil && b.scene != nil {
						return false
					}
					if a.scene.sceneId == b.scene.sceneId {
						return false
					}
					return a.scene.sceneId > b.scene.sceneId
				}
			case common.OrderColumnLose:
				sortFunc = func(i, j int) bool {
					if players[i].FailTimes == players[j].FailTimes {
						return false
					}
					return players[i].FailTimes > players[j].FailTimes
				}
			case common.OrderColumnWin:
				sortFunc = func(i, j int) bool {
					if players[i].WinTimes == players[j].WinTimes {
						return false
					}
					return players[i].WinTimes > players[j].WinTimes
				}
			case common.OrderColumnDraw:
				sortFunc = func(i, j int) bool {
					if players[i].DrawTimes == players[j].DrawTimes {
						return false
					}
					return players[i].DrawTimes > players[j].DrawTimes
				}
			case common.OrderColumnWinCoin:
				sortFunc = func(i, j int) bool {
					if players[i].WinCoin == players[j].WinCoin {
						return false
					}
					return players[i].WinCoin > players[j].WinCoin
				}
			case common.OrderColumnLoseCoin:
				sortFunc = func(i, j int) bool {
					if players[i].FailCoin == players[j].FailCoin {
						return false
					}
					return players[i].FailCoin > players[j].FailCoin
				}
			default:
			}
			if sortFunc != nil {
				sort.Slice(players, func(i, j int) bool {
					if orderType == 0 {
						return sortFunc(i, j)
					}
					return !sortFunc(i, j)
				})
			}
			count = len(players)
			if int(end) > count {
				end = int32(count)
			}
			for i := start; i < end; i++ {
				p := players[i]
				if p != nil {
					pb := model.ConvertPlayerDataToWebData(p.PlayerData)
					if p.scene != nil {
						pb.GameFreeId = p.scene.dbGameFree.Id
						pb.SceneId = int32(p.scene.sceneId)
					}
					pack.PlayerData = append(pack.PlayerData, pb)
				}
			}
			pack.Tag = webapi_proto.TagCode_SUCCESS
			pack.PageCount = int32(count)
			pack.PageNo = pageNo
			pack.PageSize = pageSize
			return common.ResponseTag_Ok, pack
		}))
	//黑白名单
	WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/api/Player/WhiteBlackControl", WebAPIHandlerWrapper(
		func(tNode *transact.TransNode, params []byte) (int, proto.Message) {
			pack := &webapi_proto.SAWhiteBlackControl{}
			msg := &webapi_proto.ASWhiteBlackControl{}
			err := proto.Unmarshal(params, msg)
			pack.Tag = webapi_proto.TagCode_FAILED
			if err != nil {
				pack.Msg = "数据序列化失败" + err.Error()
				return common.ResponseTag_ParamError, pack
			}
			if msg.SnId == 0 {
				pack.Msg = "snid is zero"
				return common.ResponseTag_ParamError, pack
			}
			idMap := make(map[int32]struct{})
			if msg.SnId > 0 {
				idMap[msg.SnId] = struct{}{}
			}
			for _, id := range msg.SnIds {
				idMap[id] = struct{}{}
			}
			wbCoinLimit := msg.WBCoinLimit
			maxNum := msg.MaxNum
			resetTotalCoin := msg.ResetTotalCoin
			tNow := time.Now()
			var errMsg interface{}
			wg := new(sync.WaitGroup)
			wg.Add(len(idMap))
			for id := range idMap {
				snid := id
				player := PlayerMgrSington.GetPlayerBySnId(snid)
				if player != nil {
					if len(msg.Platform) == 0 || player.Platform == msg.Platform {
						player.WBLevel = msg.WBLevel
						player.WBCoinLimit = wbCoinLimit
						player.WBMaxNum = maxNum
						if resetTotalCoin == 1 {
							player.WBCoinTotalIn = 0
							player.WBCoinTotalOut = 0
						}
						player.WBTime = tNow
						//出去比赛场
						if player.scene != nil && !player.scene.IsMatchScene() {
							//同步刷到游服
							wgpack := &server.WGSetPlayerBlackLevel{
								SnId:           proto.Int(int(snid)),
								WBLevel:        proto.Int32(msg.WBLevel),
								WBCoinLimit:    proto.Int64(wbCoinLimit),
								MaxNum:         proto.Int32(maxNum),
								ResetTotalCoin: proto.Bool(resetTotalCoin != 0),
							}
							player.SendToGame(int(server.SSPacketID_PACKET_WG_SETPLAYERBLACKLEVEL), wgpack)
						}
					}
				}
				task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
					if resetTotalCoin == 1 {
						//后台sql有统计这里屏蔽掉
						//log := model.NewBlackWhiteCoinLog(snid, msg.WBLevel, 0, int32(resetTotalCoin), msg.Platform)
						//if log != nil {
						//	model.InsertBlackWhiteCoinLog(log)
						//}
						return model.SetBlackWhiteLevel(snid, msg.WBLevel, int32(maxNum), 0, 0, wbCoinLimit, msg.Platform, tNow)
					} else {
						//后台sql有统计这里屏蔽掉
						//log := model.NewBlackWhiteCoinLog(snid, msg.WBLevel, wbCoinLimit, int32(resetTotalCoin), msg.Platform)
						//if log != nil {
						//	model.InsertBlackWhiteCoinLog(log)
						//}
						return model.SetBlackWhiteLevelUnReset(snid, msg.WBLevel, int32(maxNum), wbCoinLimit, msg.Platform, tNow)
					}
				}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
					defer wg.Done()
					if data == nil {
						delete(idMap, snid)
					} else {
						errMsg = data
					}
				}), "SetBlackWhiteLevel").Start()
			}
			task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
				wg.Wait()
				if errMsg != nil || len(idMap) > 0 {
					pack.Tag = webapi_proto.TagCode_FAILED
					pack.Msg = fmt.Sprintf("数据库写入错误：%v , 玩家id：%v", errMsg, idMap)
				} else {
					pack.Tag = webapi_proto.TagCode_SUCCESS
				}
				tNode.TransRep.RetFiels = pack
				tNode.Resume()
				return nil
			}), nil).Start()
			return common.ResponseTag_TransactYield, pack
		}))
	WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/api/Ctrl/ListServerStates", WebAPIHandlerWrapper(
		func(tNode *transact.TransNode, params []byte) (int, proto.Message) {
			pack := &webapi_proto.SAListServerStates{}
			msg := &webapi_proto.ASListServerStates{}
			err := proto.Unmarshal(params, msg)
			if err != nil {
				pack.Tag = webapi_proto.TagCode_FAILED
				pack.Msg = "数据序列化失败" + err.Error()
				return common.ResponseTag_ParamError, pack
			}
			pack.Tag = webapi_proto.TagCode_SUCCESS
			pack.ServerInfo = GameSessMgrSington.ListServerState(int(msg.SrvId), int(msg.SrvType))
			return common.ResponseTag_Ok, pack
		}))
	//邮件
	WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/api/Game/CreateShortMessage", WebAPIHandlerWrapper(
		func(tNode *transact.TransNode, params []byte) (int, proto.Message) {
			pack := &webapi_proto.SACreateShortMessage{}
			msg := &webapi_proto.ASCreateShortMessage{}
			err1 := proto.Unmarshal(params, msg)
			if err1 != nil {
				pack.Tag = webapi_proto.TagCode_FAILED
				pack.Msg = "数据序列化失败" + err1.Error()
				return common.ResponseTag_ParamError, pack
			}

			title := msg.GetNoticeTitle()
			content := msg.GetNoticeContent()
			platform := msg.GetPlatform()
			srcSnid := msg.GetSrcSnid()
			destSnid := msg.GetDestSnid()
			messageType := msg.GetMessageType()
			coin := msg.GetCoin()
			diamond := msg.GetDiamond()
			otherParams := msg.GetParams()
			showId := msg.GetShowId()

			if messageType == model.MSGTYPE_ITEM && len(otherParams) != 0 && len(otherParams)%2 != 0 {
				pack.Tag = webapi_proto.TagCode_FAILED
				pack.Msg = "otherParams is err"
				return common.ResponseTag_ParamError, pack
			}

			//由于多线程问题,次数组只能在此处和task.new 内使用,在其它地方使用会引起多线程崩溃的问题
			var onlinePlayerSnid []int32
			if destSnid == 0 {
				for _, p := range PlayerMgrSington.playerSnMap {
					if p.IsRob == true { //排除掉机器人
						continue
					}
					if platform == "" || p.Platform == platform {
						onlinePlayerSnid = append(onlinePlayerSnid, p.SnId)
					}
				}
			}

			var newMsg *model.Message
			var dbMsgs []*model.Message
			var err error
			task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
				code := login_proto.OpResultCode_OPRC_Sucess
				//如果不是群发则检查一下该用户snid是不是存在，不存在就返回，在这里不关心玩家帐变是否正确，不做校正
				if destSnid != 0 {
					playerData, _ := model.GetPlayerDataBySnId(platform, destSnid, true, false)
					if playerData == nil {
						code = login_proto.OpResultCode_OPRC_Error
						return code
					}
				}
				// var otherParams []int32
				newMsg = model.NewMessage("", int32(srcSnid), "系统", int32(destSnid), int32(messageType), title, content, coin, diamond,
					model.MSGSTATE_UNREAD, time.Now().Unix(), model.MSGATTACHSTATE_DEFAULT, "", otherParams, platform, showId)
				if newMsg != nil {
					err := model.InsertMessage(platform, newMsg)
					if err != nil {
						code = login_proto.OpResultCode_OPRC_Error
						logger.Logger.Errorf("/api/Game/CreateShortMessage,InsertMessage err:%v title:%v content:%v platform:%v srcSnid:%v destSnid:%v messageType:%v ", err, title, content, platform, srcSnid, destSnid, messageType)
						return code
					}
				}

				if destSnid == 0 {
					for _, psnid := range onlinePlayerSnid {
						newMsg := model.NewMessage(newMsg.Id.Hex(), newMsg.SrcId, "系统", psnid, newMsg.MType, newMsg.Title, newMsg.Content, newMsg.Coin, newMsg.Diamond,
							newMsg.State, newMsg.CreatTs, newMsg.AttachState, newMsg.GiftId, otherParams, platform, newMsg.ShowId)
						if newMsg != nil {
							dbMsgs = append(dbMsgs, newMsg)
						}
					}
					if len(dbMsgs) > 0 {
						err := model.InsertMessage(platform, dbMsgs...)
						if err != nil {
							logger.Logger.Errorf("/api/Game/CreateShortMessage,InsertMessage err:%v title:%v content:%v platform:%v srcSnid:%v destSnid:%v messageType:%v ", err, title, content, platform, srcSnid, destSnid, messageType)
						}
					}
				}
				return code
			}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
				if code, ok := data.(login_proto.OpResultCode); ok {
					if code != login_proto.OpResultCode_OPRC_Sucess {
						pack.Tag = webapi_proto.TagCode_SUCCESS
						pack.Msg = ""
					} else {
						pp := PlayerMgrSington.GetPlayerBySnId(int32(destSnid))
						if pp != nil {
							//单独消息
							pp.AddMessage(newMsg)
						} else {
							if destSnid == 0 { //群发消息
								MsgMgrSington.AddMsg(newMsg)

								if len(platform) > 0 { //特定渠道消息
									for _, m := range dbMsgs {
										pp := PlayerMgrSington.GetPlayerBySnId(m.SnId)
										if pp != nil && pp.Platform == platform {
											pp.AddMessage(m)
										}
									}
								} else { //全渠道消息
									for _, m := range dbMsgs {
										pp := PlayerMgrSington.GetPlayerBySnId(m.SnId)
										if pp != nil {
											pp.AddMessage(m)
										}
									}
								}
							}
						}
						pack.Tag = webapi_proto.TagCode_SUCCESS
						pack.Msg = ""
					}
					tNode.TransRep.RetFiels = pack
					tNode.Resume()
					if err != nil {
						logger.Logger.Error("Marshal CreateShortMessage response data error:", err)
					}
				} else {
					logger.Logger.Error("CreateShortMessage task result data covert failed.")
				}
			}), "CreateShortMessage").Start()

			return common.ResponseTag_TransactYield, pack
		}))
	//查询邮件
	WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/api/Game/QueryShortMessageList", WebAPIHandlerWrapper(
		func(tNode *transact.TransNode, params []byte) (int, proto.Message) {
			pack := &webapi_proto.SAQueryShortMessageList{}
			msg := &webapi_proto.ASQueryShortMessageList{}
			err := proto.Unmarshal(params, msg)
			if err != nil {
				pack.Tag = webapi_proto.TagCode_FAILED
				pack.Msg = "数据序列化失败" + err.Error()
				return common.ResponseTag_ParamError, pack
			}
			startTime := msg.GetStartTime()
			endTime := msg.GetEndTime()
			pageNo := msg.GetPageNo()
			pageSize := msg.GetPageSize()
			messageType := msg.GetMessageType()
			platform := msg.GetPlatform()
			destSnid := msg.GetDestSnid()
			if destSnid == 0 {
				destSnid = -1
			}
			start := (pageNo - 1) * pageSize
			end := pageNo * pageSize

			var msgs []model.Message
			var count int
			task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
				mas := &model.MsgArgs{
					Platform:  platform,
					StartTs:   startTime,
					EndTs:     endTime,
					ToIndex:   int(end),
					DestSnId:  int(destSnid),
					MsgType:   int(messageType),
					FromIndex: int(start),
				}
				retMsgs, err := model.GetMessageInRangeTsLimitByRange(mas)
				count = retMsgs.Count
				msgs = retMsgs.Msg
				return err
			}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
				if data != nil {
					pack.Tag = webapi_proto.TagCode_FAILED
					pack.Msg = ""
				} else {
					pack.Tag = webapi_proto.TagCode_SUCCESS
					pack.Msg = ""
					pack.Count = int32(count)
					for i := 0; i < len(msgs); i++ {
						giftState := int32(0)
						mi := &webapi_proto.MessageInfo{
							Id:         msgs[i].Id.Hex(),
							MType:      msgs[i].MType,
							Title:      msgs[i].Title,
							Content:    msgs[i].Content,
							State:      msgs[i].State,
							CreateTime: msgs[i].CreatTs,
							SrcSnid:    msgs[i].SrcId,
							DestSnid:   msgs[i].SnId,
							Coin:       msgs[i].Coin,
							GiftId:     msgs[i].GiftId,
							GiftState:  giftState,
							Platform:   msgs[i].Platform,
						}
						pack.MessageInfo = append(pack.MessageInfo, mi)
					}
				}
				tNode.TransRep.RetFiels = pack
				tNode.Resume()
			}), "QueryShortMessageList").Start()
			return common.ResponseTag_TransactYield, pack
		}))
	//删除邮件
	WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/api/Game/DeleteShortMessage", WebAPIHandlerWrapper(
		func(tNode *transact.TransNode, params []byte) (int, proto.Message) {
			pack := &webapi_proto.SADeleteShortMessage{}
			msg := &webapi_proto.ASDeleteShortMessage{}
			err1 := proto.Unmarshal(params, msg)
			if err1 != nil {
				pack.Tag = webapi_proto.TagCode_FAILED
				pack.Msg = "数据序列化失败" + err1.Error()
				return common.ResponseTag_ParamError, pack
			}
			msgId := msg.GetId()
			platform := msg.GetPlatform()
			if len(msgId) == 0 {
				pack.Tag = webapi_proto.TagCode_FAILED
				pack.Msg = "not find msg id"
				return common.ResponseTag_ParamError, pack
			}
			var delMsg *model.Message
			var err error
			task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
				delMsg, err = model.GetMessageById(msgId, platform)
				if err == nil {
					if len(platform) > 0 && msg.Platform != platform {
						return errors.New("Platform error.")
					}
					args := &model.DelMsgArgs{
						Platform: platform,
						Id:       delMsg.Id,
					}
					err = model.DelMessage(args) // 真删除
				}
				return err
			}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
				if data != nil || delMsg == nil {
					pack.Tag = webapi_proto.TagCode_FAILED
					pack.Msg = "not find msg id"
					tNode.TransRep.RetFiels = pack
					tNode.Resume()
					return
				}
				/*if delMsg.State == model.MSGSTATE_REMOVEED {
					pack.Tag = webapi_proto.TagCode_FAILED
					pack.Msg = "repeate delete"
					tNode.TransRep.RetFiels = pack
					tNode.Resume()
					return
				}*/
				pp := PlayerMgrSington.GetPlayerBySnId(delMsg.SnId)
				if pp != nil {
					pp.WebDelMessage(delMsg.Id.Hex())
				} else {
					MsgMgrSington.RemoveMsg(delMsg)
				}

				pack.Tag = webapi_proto.TagCode_SUCCESS
				pack.Msg = ""
				tNode.TransRep.RetFiels = pack
				tNode.Resume()
			}), "DeleteShortMessage").Start()
			return common.ResponseTag_TransactYield, pack
		}))
	//删除玩家已经删除的邮件
	WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/api/Game/DeleteAllShortMessage", WebAPIHandlerWrapper(
		func(tNode *transact.TransNode, params []byte) (int, proto.Message) {
			pack := &webapi_proto.SADeleteShortMessage{}
			msg := &webapi_proto.ASDeleteShortMessage{}
			err1 := proto.Unmarshal(params, msg)
			if err1 != nil {
				pack.Tag = webapi_proto.TagCode_FAILED
				pack.Msg = "数据序列化失败" + err1.Error()
				return common.ResponseTag_ParamError, pack
			}
			platform := msg.GetPlatform()
			if len(platform) == 0 {
				pack.Tag = webapi_proto.TagCode_FAILED
				pack.Msg = "platform is nil"
				return common.ResponseTag_ParamError, pack
			}
			task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
				return model.DelAllMessage(&model.DelAllMsgArgs{Platform: platform})
			}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
				if data != nil {
					pack.Tag = webapi_proto.TagCode_FAILED
					pack.Msg = "del err:" + data.(error).Error()
					tNode.TransRep.RetFiels = pack
					tNode.Resume()
					return
				}
				pack.Tag = webapi_proto.TagCode_SUCCESS
				pack.Msg = "del success."
				tNode.TransRep.RetFiels = pack
				tNode.Resume()
			}), "DeleteShortMessage").Start()
			return common.ResponseTag_TransactYield, pack
		}))
	WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/api/Player/BlackBySnId", WebAPIHandlerWrapper(
		func(tNode *transact.TransNode, params []byte) (int, proto.Message) {
			pack := &webapi_proto.SABlackBySnId{}
			msg := &webapi_proto.ASBlackBySnId{}
			err := proto.Unmarshal(params, msg)
			if err != nil {
				pack.Tag = webapi_proto.TagCode_FAILED
				pack.Msg = "数据序列化失败" + err.Error()
				return common.ResponseTag_ParamError, pack
			}
			player := PlayerMgrSington.GetPlayerBySnId(msg.SnId)
			if player != nil {
				if player.Platform != msg.Platform {
					pack.Tag = webapi_proto.TagCode_FAILED
					pack.Msg = "Platform is dif"
					return common.ResponseTag_ParamError, pack
				}
				pack.Tag = webapi_proto.TagCode_SUCCESS
				player.BlacklistType = msg.BlacklistType
				player.dirty = true
				player.Time2Save()
				//在线需要踢掉玩家
				if uint(msg.BlacklistType)&(BlackState_Login+1) == uint(BlackState_Login+1) {
					logger.Logger.Infof("found platform:%v player:%d snid in blacklist", msg.Platform, player.SnId)
					player.Kickout(int32(login_proto.SSDisconnectTypeCode_SSDTC_BlackList))
				}
				return common.ResponseTag_Ok, pack
			} else {
				task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
					data, _ := model.GetPlayerDataBySnId(msg.Platform, msg.SnId, true, false)
					if data != nil {
						model.UpdatePlayerBlacklistType(msg.Platform, msg.SnId, msg.BlacklistType)
					}
					return data
				}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
					playerData, ok := data.(*model.PlayerData)
					if !ok || playerData == nil {
						pack.Msg = "no find player"
						pack.Tag = webapi_proto.TagCode_FAILED
					} else {
						pack.Tag = webapi_proto.TagCode_SUCCESS
						pack.Msg = "success"
					}
					tNode.TransRep.RetFiels = pack
					tNode.Resume()
				}), "PlayerData").Start()
				return common.ResponseTag_TransactYield, pack
			}
		}))
	WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/api/Message/QueryHorseRaceLampList", WebAPIHandlerWrapper(
		func(tNode *transact.TransNode, params []byte) (int, proto.Message) {
			pack := &webapi_proto.SAQueryHorseRaceLampList{}
			msg := &webapi_proto.ASQueryHorseRaceLampList{}
			err := proto.Unmarshal(params, msg)
			if err != nil {
				pack.Tag = webapi_proto.TagCode_FAILED
				pack.Msg = "数据序列化失败" + err.Error()
				return common.ResponseTag_ParamError, pack
			}
			start := (msg.PageNo - 1) * msg.PageSize
			end := msg.PageNo * msg.PageSize
			task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
				ret, erro := model.GetHorseRaceLampInRangeTsLimitByRange(msg.Platform, msg.MsgType, msg.State, start, end)
				if erro != nil {
					logger.Logger.Error("api GetNoticeInRangeTsLimitByRange is error", erro)
					return nil
				}
				return ret
			}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
				if data == nil {
					pack.Tag = webapi_proto.TagCode_SUCCESS
					pack.Msg = "no these msg"
				} else {
					noticesRet := data.(*model.QueryHorseRaceLampRet)
					if noticesRet != nil {
						var apiNoticMsgList []*webapi_proto.HorseRaceLamp
						for _, notice := range noticesRet.Data {
							apiNoticeMsg := &webapi_proto.HorseRaceLamp{
								Id:         notice.Id.Hex(),
								Platform:   notice.Platform,
								Title:      notice.Title,
								Content:    notice.Content,
								Footer:     notice.Footer,
								StartTime:  notice.StartTime,
								Frequency:  notice.Interval,
								State:      notice.State,
								CreateTime: notice.CreateTime,
								Count:      notice.Count,
								Priority:   notice.Priority,
								MsgType:    notice.MsgType,
								Target:     notice.Target,
								StandSec:   notice.StandSec,
							}
							apiNoticMsgList = append(apiNoticMsgList, apiNoticeMsg)
						}
						pack.PageCount = int32(noticesRet.Count)
						pack.HorseRaceLamp = apiNoticMsgList
					}
					pack.Tag = webapi_proto.TagCode_SUCCESS
				}
				tNode.TransRep.RetFiels = pack
				tNode.Resume()
				if err != nil {
					logger.Logger.Error("Marshal notice msg list error:", err)
				}
			}), "QueryHorseRaceLampList").Start()
			return common.ResponseTag_TransactYield, pack
		}))
	WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/api/Message/CreateHorseRaceLamp", WebAPIHandlerWrapper(
		func(tNode *transact.TransNode, params []byte) (int, proto.Message) {
			pack := &webapi_proto.SACreateHorseRaceLamp{}
			msg := &webapi_proto.ASCreateHorseRaceLamp{}
			err := proto.Unmarshal(params, msg)
			if err != nil {
				pack.Tag = webapi_proto.TagCode_FAILED
				pack.Msg = "数据序列化失败" + err.Error()
				return common.ResponseTag_ParamError, pack
			}
			platform := msg.Platform
			title := msg.Title
			content := msg.Content
			footer := msg.Footer
			count := msg.Count
			state := msg.State
			startTime := msg.StartTime
			priority := msg.Priority
			msgType := msg.MsgType
			standSec := msg.StandSec
			target := msg.Target
			var horseRaceLamp *model.HorseRaceLamp
			task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
				horseRaceLamp = model.NewHorseRaceLamp("", platform, title, content, footer, startTime, standSec, count,
					priority, state, msgType, target, standSec)
				return model.InsertHorseRaceLamp(platform, horseRaceLamp)
			}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
				if data != nil {
					pack.Tag = webapi_proto.TagCode_FAILED
					pack.Msg = data.(error).Error()
					tNode.TransRep.RetFiels = pack
					tNode.Resume()
					return
				}
				HorseRaceLampMgrSington.AddHorseRaceLampMsg(horseRaceLamp.Id.Hex(), "", platform, title, content, footer, startTime, standSec,
					count, msgType, state, priority, horseRaceLamp.CreateTime, target, standSec)
				pack.Tag = webapi_proto.TagCode_SUCCESS
				tNode.TransRep.RetFiels = pack
				tNode.Resume()
			}), "CreateHorseRaceLamp").Start()
			return common.ResponseTag_TransactYield, pack
		}))
	WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/api/Message/GetHorseRaceLampById", WebAPIHandlerWrapper(
		func(tNode *transact.TransNode, params []byte) (int, proto.Message) {
			pack := &webapi_proto.SAGetHorseRaceLampById{}
			msg := &webapi_proto.ASGetHorseRaceLampById{}
			err := proto.Unmarshal(params, msg)
			if err != nil {
				pack.Tag = webapi_proto.TagCode_FAILED
				pack.Msg = "数据序列化失败" + err.Error()
				return common.ResponseTag_ParamError, pack
			}
			noticeKey := msg.NoticeId
			platform := msg.Platform
			horseRaceLamp := HorseRaceLampMgrSington.HorseRaceLampMsgList[noticeKey]
			if horseRaceLamp == nil || (len(platform) > 0 && horseRaceLamp.Platform != platform) {
				pack.Tag = webapi_proto.TagCode_FAILED
				pack.Msg = "not find data"
			} else {
				pack.HorseRaceLamp = &webapi_proto.HorseRaceLamp{
					Id:         noticeKey,
					Title:      horseRaceLamp.Title,
					Content:    horseRaceLamp.Content,
					Footer:     horseRaceLamp.Footer,
					StartTime:  horseRaceLamp.StartTime,
					Frequency:  horseRaceLamp.Interval,
					Count:      horseRaceLamp.Count,
					State:      horseRaceLamp.State,
					CreateTime: horseRaceLamp.CreateTime,
					Priority:   horseRaceLamp.Priority,
					MsgType:    horseRaceLamp.MsgType,
				}
				pack.Tag = webapi_proto.TagCode_SUCCESS
			}
			return common.ResponseTag_Ok, pack
		}))
	WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/api/Message/EditHorseRaceLamp", WebAPIHandlerWrapper(
		func(tNode *transact.TransNode, params []byte) (int, proto.Message) {
			pack := &webapi_proto.SAEditHorseRaceLamp{}
			msg := &webapi_proto.ASEditHorseRaceLamp{}
			err := proto.Unmarshal(params, msg)
			if err != nil {
				pack.Tag = webapi_proto.TagCode_FAILED
				pack.Msg = "数据序列化失败" + err.Error()
				return common.ResponseTag_ParamError, pack
			}
			noticeKey := msg.HorseRaceLamp.Id
			if len(noticeKey) == 0 {
				pack.Tag = webapi_proto.TagCode_FAILED
				pack.Msg = "NoticeMsg id is nil"
				return common.ResponseTag_ParamError, pack
			}
			hrl := msg.HorseRaceLamp
			task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
				notice, err2 := model.GetHorseRaceLamp(hrl.Platform, bson.ObjectIdHex(noticeKey))
				if err2 != nil {
					logger.Logger.Error("api GetNotice is error", err2)
					return nil
				}
				model.EditHorseRaceLamp(&model.HorseRaceLamp{
					Id:         bson.ObjectIdHex(noticeKey),
					Channel:    "",
					Title:      hrl.Title,
					Content:    hrl.Content,
					Footer:     hrl.Footer,
					StartTime:  hrl.StartTime,
					Interval:   hrl.Frequency,
					Count:      hrl.Count,
					CreateTime: hrl.CreateTime,
					Priority:   hrl.Priority,
					MsgType:    hrl.MsgType,
					Platform:   hrl.Platform,
					State:      hrl.State,
					Target:     hrl.Target,
					StandSec:   hrl.StandSec,
				})
				return notice
			}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
				if data == nil {
					pack.Tag = webapi_proto.TagCode_FAILED
					pack.Msg = "api GetNotice is error"
					tNode.TransRep.RetFiels = pack
					tNode.Resume()
					if err != nil {
						logger.Logger.Error("Marshal EditHorseRaceLamp response data error1:", err)
					}
					return
				} else {
					cache := HorseRaceLampMgrSington.EditHorseRaceLampMsg(&HorseRaceLamp{
						Key:        noticeKey,
						Channel:    "",
						Title:      hrl.Title,
						Content:    hrl.Content,
						Footer:     hrl.Footer,
						StartTime:  hrl.StartTime,
						Interval:   hrl.Frequency,
						Count:      hrl.Count,
						CreateTime: hrl.CreateTime,
						Priority:   hrl.Priority,
						MsgType:    hrl.MsgType,
						Platform:   hrl.Platform,
						State:      hrl.State,
						Target:     hrl.Target,
						StandSec:   hrl.StandSec,
					})
					if !cache {
						pack.Tag = webapi_proto.TagCode_FAILED
						pack.Msg = "api EditNoticeMsg is error"
						tNode.TransRep.RetFiels = pack
						tNode.Resume()
						return
					}
				}
				pack.Tag = webapi_proto.TagCode_SUCCESS
				tNode.TransRep.RetFiels = pack
				tNode.Resume()
				if err != nil {
					logger.Logger.Error("Marshal EditHorseRaceLamp response data error3:", err)
				}
			}), "EditHorseRaceLamp").Start()

			return common.ResponseTag_TransactYield, pack
		}))
	WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/api/Message/RemoveHorseRaceLampById", WebAPIHandlerWrapper(
		func(tNode *transact.TransNode, params []byte) (int, proto.Message) {
			pack := &webapi_proto.SARemoveHorseRaceLampById{}
			msg := &webapi_proto.ASRemoveHorseRaceLampById{}
			err := proto.Unmarshal(params, msg)
			if err != nil {
				pack.Tag = webapi_proto.TagCode_FAILED
				pack.Msg = "数据序列化失败" + err.Error()
				return common.ResponseTag_ParamError, pack
			}
			noticeKey := msg.GetHorseRaceId()
			platform := msg.GetPlatform()
			notice := HorseRaceLampMgrSington.HorseRaceLampMsgList[noticeKey]
			if notice == nil {
				pack.Tag = webapi_proto.TagCode_FAILED
				pack.Msg = "not find data"
				return common.ResponseTag_Ok, pack
			}
			if len(platform) > 0 && notice.Platform != notice.Platform {
				pack.Tag = webapi_proto.TagCode_FAILED
				pack.Msg = "not find data"
				return common.ResponseTag_Ok, pack
			}
			task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
				return model.RemoveHorseRaceLamp(platform, noticeKey)
			}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
				if data != nil {
					pack.Tag = webapi_proto.TagCode_FAILED
					pack.Msg = "RemoveNotice is error" + data.(error).Error()
					tNode.TransRep.RetFiels = pack
					tNode.Resume()
					return
				}
				HorseRaceLampMgrSington.DelHorseRaceLampMsg(noticeKey)
				pack.Tag = webapi_proto.TagCode_SUCCESS
				tNode.TransRep.RetFiels = pack
				tNode.Resume()
			}), "ResponseTag_TransactYield").Start()
			return common.ResponseTag_TransactYield, pack
		}))
	WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/api/Ctrl/ResetEtcdData", WebAPIHandlerWrapper(
		func(tNode *transact.TransNode, params []byte) (int, proto.Message) {
			pack := &webapi_proto.SAResetEtcdData{}
			msg := &webapi_proto.ASResetEtcdData{}
			err := proto.Unmarshal(params, msg)
			if err != nil {
				pack.Tag = webapi_proto.TagCode_FAILED
				pack.Msg = "数据序列化失败" + err.Error()
				return common.ResponseTag_ParamError, pack
			}
			EtcdMgrSington.Reset()
			pack.Tag = webapi_proto.TagCode_SUCCESS
			pack.Msg = "Etcd Reset success"
			return common.ResponseTag_Ok, nil
		}))
	WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/api/Game/SinglePlayerAdjust", WebAPIHandlerWrapper(
		func(tNode *transact.TransNode, params []byte) (int, proto.Message) {
			pack := &webapi_proto.SASinglePlayerAdjust{}
			msg := &webapi_proto.ASSinglePlayerAdjust{}
			err2 := proto.Unmarshal(params, msg)
			if err2 != nil {
				fmt.Printf("err:%v", err2)
				pack.Tag = webapi_proto.TagCode_FAILED
				pack.Msg = "数据序列化失败"
				return common.ResponseTag_ParamError, pack
			}
			////////////////////验证玩家///////////////////
			snid := msg.PlayerSingleAdjust.SnId
			platform := msg.PlayerSingleAdjust.Platform
			p := PlayerMgrSington.GetPlayerBySnId(snid)
			if p == nil {
				task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
					pb, _ := model.GetPlayerDataBySnId(platform, snid, false, false)
					return pb
				}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
					if data == nil || data.(*model.PlayerData) == nil {
						pack.Tag = webapi_proto.TagCode_FAILED
						pack.Msg = "not find player."
					} else {
						pack.Tag = webapi_proto.TagCode_SUCCESS
						pack.Msg = "success"
						sa := PlayerSingleAdjustMgr.WebData(msg, &Player{PlayerData: data.(*model.PlayerData)})
						if sa != nil {
							pack.PlayerSingleAdjust = sa
						}
					}
					tNode.TransRep.RetFiels = pack
					tNode.Resume()
				}), "SinglePlayerAdjust").Start()
				return common.ResponseTag_TransactYield, pack
			} else {
				pack.Tag = webapi_proto.TagCode_SUCCESS
				pack.Msg = "success"
				sa := PlayerSingleAdjustMgr.WebData(msg, p)
				if sa != nil {
					pack.PlayerSingleAdjust = sa
				}
			}
			return common.ResponseTag_Ok, pack
		}))
	//获取在线统计
	WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/api/Report/OnlineReportTotal", WebAPIHandlerWrapper(
		func(tNode *transact.TransNode, params []byte) (int, proto.Message) {
			pack := &webapi_proto.SAOnlineReportTotal{}
			msg := &webapi_proto.ASOnlineReportTotal{}
			err := proto.Unmarshal(params, msg)
			if err != nil {
				pack.Tag = webapi_proto.TagCode_FAILED
				pack.Msg = "数据序列化失败" + err.Error()
				return common.ResponseTag_ParamError, pack
			}
			inGameCnt := make(map[int]int32)
			pack.OnlineReport = &webapi_proto.OnlineReport{}
			tNow := time.Now()
			for _, p := range PlayerMgrSington.players {
				if !p.IsOnLine() {
					continue
				}
				if len(msg.Platform) != 0 && msg.Platform != "0" && p.Platform != msg.Platform {
					continue
				}
				if len(msg.Channel) != 0 && p.Channel != msg.Channel {
					continue
				}
				if len(msg.Promoter) != 0 && p.BeUnderAgentCode != msg.Promoter {
					continue
				}
				pack.OnlineReport.TotalCnt++
				if p.DeviceOS == common.DeviceOS_Android {
					pack.OnlineReport.AndroidOnlineCnt++
				} else if p.DeviceOS == common.DeviceOS_IOS {
					pack.OnlineReport.IosOnlineCnt++
				}
				if p.scene == nil {
					pack.OnlineReport.DatingPlayers++
				} else {
					pack.OnlineReport.OnRoomPlayers++
					inGameCnt[p.scene.gameId]++
				}
				if common.InSameDay(p.CreateTime, tNow) {
					pack.OnlineReport.TodayRegisterOnline++
				} else {
					if common.DiffDay(tNow, p.CreateTime) <= 7 {
						pack.OnlineReport.SevenDayRegisterOnline++
					}
				}
			}
			var onlineGameCnt []*webapi_proto.OnlineGameCnt
			for gameId, cnt := range inGameCnt {
				onlineGameCnt = append(onlineGameCnt, &webapi_proto.OnlineGameCnt{
					GameId: int32(gameId),
					Cnt:    cnt},
				)
			}
			pack.Tag = webapi_proto.TagCode_SUCCESS
			pack.OnlineReport.GameCount = onlineGameCnt
			return common.ResponseTag_Ok, pack
		}))
	//   CreateJYB 创建礼包码
	WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/api/Game/CreateJYB", WebAPIHandlerWrapper(
		func(tNode *transact.TransNode, params []byte) (int, proto.Message) {
			pack := &webapi_proto.SACreateJYB{}
			msg := &webapi_proto.ASCreateJYB{}

			err := proto.Unmarshal(params, msg)
			if err != nil {
				pack.Tag = webapi_proto.TagCode_FAILED
				pack.Msg = "数据序列化失败" + err.Error()
				return common.ResponseTag_ParamError, pack
			}
			if msg.CodeType == 1 && (msg.Code == "" || len(msg.Code) > 11) {
				pack.Tag = webapi_proto.TagCode_JYB_DATA_ERROR
				pack.Msg = "CreateJYB Code failed"
				return common.ResponseTag_Ok, pack
			}
			// (plt, name, content string, startTs, endTs int64, max int32, award *JybInfoAward)
			award := &model.JybInfoAward{}
			if msg.GetAward() != nil {
				award.Coin = msg.GetAward().Coin
				award.Diamond = msg.GetAward().Diamond
				if msg.GetAward().GetItemId() != nil {
					for _, item := range msg.GetAward().ItemId {
						if v := srvdata.PBDB_GameItemMgr.GetData(item.ItemId); item != nil {
							if item.ItemNum == 0 {
								pack.Tag = webapi_proto.TagCode_JYB_DATA_ERROR
								pack.Msg = "ItemNum failed"
								return common.ResponseTag_Ok, pack
							}
							award.Item = append(award.Item, &model.Item{
								ItemId:     v.Id,
								ItemNum:    item.ItemNum,
								ObtainTime: time.Now().Unix(),
							})
						} else {
							pack.Tag = webapi_proto.TagCode_JYB_DATA_ERROR
							pack.Msg = "ItemId failed"
							return common.ResponseTag_Ok, pack
						}
					}
				}
			}
			jyb := model.NewJybInfo(msg.Platform, msg.Name, msg.Content, msg.Code, msg.StartTime, msg.EndTime, msg.Max, msg.CodeType, award)
			args := &model.CreateJyb{
				JybInfo: jyb,
				Codelen: msg.CodeLen,
				Num:     uint64(msg.Max),
			}
			task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
				return model.CreateJybInfo(args)
			}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
				if data == nil {
					pack.Tag = webapi_proto.TagCode_SUCCESS
					pack.Msg = "CreateJYB success"
				} else {
					logger.Logger.Trace("err: ", data.(error))
					if data.(error) == model.ErrJYBCode {
						pack.Tag = webapi_proto.TagCode_JYB_CODE_EXIST
					} else {
						pack.Tag = webapi_proto.TagCode_FAILED
					}
					pack.Msg = "CreateJYB failed"
				}
				tNode.TransRep.RetFiels = pack
				tNode.Resume()
			}), "webCreateJybInfo").Start()
			return common.ResponseTag_TransactYield, pack
		}))
	WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/api/Game/UpdateJYB", WebAPIHandlerWrapper(
		func(tNode *transact.TransNode, params []byte) (int, proto.Message) {
			pack := &webapi_proto.SAUpdateJYB{}
			msg := &webapi_proto.ASUpdateJYB{}
			err := proto.Unmarshal(params, msg)
			if err != nil {
				pack.Tag = webapi_proto.TagCode_FAILED
				pack.Msg = "数据序列化失败" + err.Error()
				return common.ResponseTag_ParamError, pack
			}

			if msg.Opration == 1 {
				args := &model.GetJybInfoArgs{
					Id:  msg.JYBID,
					Plt: msg.Platform,
				}
				task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
					return model.DelJybInfo(args)
				}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
					if data == nil {
						pack.Tag = webapi_proto.TagCode_SUCCESS
						pack.Msg = "UpdateJYB success"

					} else {
						pack.Tag = webapi_proto.TagCode_FAILED
						pack.Msg = "UpdateJYB failed"
					}
					tNode.TransRep.RetFiels = pack
					tNode.Resume()
				}), "webDelJybInfo").Start()
				return common.ResponseTag_TransactYield, pack
			}

			return common.ResponseTag_Ok, pack
		}))
	WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/api/Customer/UpExchangeStatus", WebAPIHandlerWrapper(
		func(tNode *transact.TransNode, params []byte) (int, proto.Message) {
			pack := &webapi_proto.SAUpExchangeStatus{}
			msg := &webapi_proto.ASUpExchangeStatus{}
			err := proto.Unmarshal(params, msg)
			if err != nil {
				pack.Tag = webapi_proto.TagCode_FAILED
				pack.Msg = "数据序列化失败" + err.Error()
				return common.ResponseTag_ParamError, pack
			}
			snid := msg.Snid
			platform := msg.Platform
			player := PlayerMgrSington.GetPlayerBySnId(snid)
			if msg.Status != Shop_Status_Revoke {
				pack.Tag = webapi_proto.TagCode_FAILED
				pack.Msg = "Status is Revoke"
				return common.ResponseTag_ParamError, pack
			}
			//cdata := ShopMgrSington.GetExchangeData(msg.GoodsId)
			item := srvdata.PBDB_GameItemMgr.GetData(VCard)
			if item == nil {
				pack.Tag = webapi_proto.TagCode_FAILED
				pack.Msg = "item is nil"
				return common.ResponseTag_ParamError, pack
			}
			addvcoin := msg.NeedNum
			remark := fmt.Sprintf("兑换撤单 %v-%v", msg.GoodsId, msg.Name)
			if player != nil {
				// 在线
				items := []*Item{{
					ItemId:  VCard,    // 物品id
					ItemNum: addvcoin, // 数量
				}}
				if _, code := BagMgrSington.AddJybBagInfo(player, items); code != bag.OpResultCode_OPRC_Sucess { // 领取失败
					logger.Logger.Errorf("UpExchangeStatus AddJybBagInfo err", code)
					pack.Msg = "AddJybBagInfo err"
					return common.ResponseTag_ParamError, pack
				}
				pack.Tag = webapi_proto.TagCode_SUCCESS
				pack.Msg = "UpExchange success"
				BagMgrSington.RecordItemLog(player.Platform, player.SnId, ItemObtain, VCard, item.Name, addvcoin, remark)
				return common.ResponseTag_Ok, pack
			} else {
				var findPlayer *model.PlayerBaseInfo
				task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
					findPlayer = model.GetPlayerBaseInfo(platform, snid)
					if findPlayer == nil {
						pack.Tag = webapi_proto.TagCode_Play_NotEXIST
						pack.Msg = fmt.Sprintf("player is not exist %v", snid)
						return errors.New("player is not exist")
					}
					newBagInfo := &model.BagInfo{
						SnId:     findPlayer.SnId,
						Platform: findPlayer.Platform,
						BagItem:  make(map[int32]*model.Item),
					}
					newBagInfo.BagItem[VCard] = &model.Item{ItemId: VCard, ItemNum: addvcoin}
					return model.SaveDBBagItem(newBagInfo)
				}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
					if data == nil && findPlayer != nil {
						pack.Tag = webapi_proto.TagCode_SUCCESS
						pack.Msg = "UpExchange success"
						BagMgrSington.RecordItemLog(findPlayer.Platform, findPlayer.SnId, ItemObtain, VCard, item.Name, addvcoin, remark)
					} else {
						pack.Tag = webapi_proto.TagCode_FAILED
						pack.Msg = "UpExchange failed:" + data.(error).Error()
					}
					tNode.TransRep.RetFiels = pack
					tNode.Resume()
				}), "UpExchange").Start()
				return common.ResponseTag_TransactYield, pack
			}
		}))
}

func SetInfo(u *model.PlayerData, o interface{}, playerMap map[string]interface{}) *model.PlayerData {
	t := reflect.TypeOf(o)
	v := reflect.ValueOf(o)
	s := reflect.ValueOf(u).Elem()
	for k, n := range playerMap {
		if f, ok := t.FieldByName(k); ok {
			if v.FieldByName(k).CanInterface() {
				switch f.Type.Kind() {
				case reflect.Int64, reflect.Int32:
					a, _ := strconv.ParseInt((fmt.Sprintf("%v", n)), 10, 64)
					s.FieldByName(k).SetInt(a)
				case reflect.Bool:
					s.FieldByName(k).SetBool(n.(bool))
				case reflect.String:
					s.FieldByName(k).SetString(n.(string))
				default:
					logger.Logger.Warn("type not in the player !!")
				}
			}
		}
	}
	return u
}
