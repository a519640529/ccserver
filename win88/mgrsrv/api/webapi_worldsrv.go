package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/idealeak/goserver/core/netlib"

	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"sync"
	"sync/atomic"

	"games.yol.com/win88/common"
	"games.yol.com/win88/model"
	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/admin"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/transact"
	"github.com/idealeak/goserver/core/utils"
	"github.com/idealeak/goserver/srvlib"
)

// API
// http://127.0.0.1:9595/api/Report/QueryOnlineReportList?ts=20141024000000&sign=41cc8cee8dd93f7dc70b6426cfd1029d
const (
	WEBAPI_TRANSACTE_EVENT int = iota
	WEBAPI_TRANSACTE_RESPONSE
)

var WebApiStats = new(sync.Map)

type ApiStats struct {
	RunTimes        int64 //执行次数
	TotalRuningTime int64 //总执行时间
	MaxRuningTime   int64 //最长执行时间
	TimeoutTimes    int64 //执行超时次数
	UnreachTimes    int64 //不可达次数
}

func WorldSrvApi(rw http.ResponseWriter, req *http.Request) {
	defer utils.DumpStackIfPanic("api.WorldSrvApi")
	logger.Logger.Info("WorldSrvApi receive:", req.URL.Path, req.URL.RawQuery)

	if common.RequestCheck(req, model.GameParamData.WhiteHttpAddr) == false {
		logger.Logger.Info("RemoteAddr [%v] require api.", req.RemoteAddr)
		return
	}
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		logger.Logger.Info("Body err.", err)
		webApiResponse(rw, nil /*map[string]interface{}{webapi.RESPONSE_STATE: webapi.STATE_ERR, webapi.RESPONSE_ERRMSG: "Post data is null!"}*/)
		return
	}
	m := req.URL.Query()
	timestamp := m.Get("nano")
	if timestamp == "" {
		logger.Logger.Info(req.RemoteAddr, " WorldSrvApi param error: nano not allow null")
		return
	}
	sign := m.Get("sign")
	if sign == "" {
		logger.Logger.Info(req.RemoteAddr, " WorldSrvApi param error: sign not allow null")
		return
	}
	startTime := time.Now().UnixNano()
	args := fmt.Sprintf("%v;%v;%v;%v", common.Config.AppId, req.URL.Path, string(data), timestamp)
	h := md5.New()
	io.WriteString(h, args)
	realSign := hex.EncodeToString(h.Sum(nil))
	if realSign != sign && !common.Config.IsDevMode {
		logger.Logger.Info(req.RemoteAddr, " srvCtrlMain sign error: expect ", realSign, " ; but get ", sign, " raw=", args)
		webApiResponse(rw, nil /*map[string]interface{}{webapi.RESPONSE_STATE: webapi.STATE_ERR, webapi.RESPONSE_ERRMSG: "Sign error!"}*/)
		return
	}

	var stats *ApiStats
	if v, exist := WebApiStats.Load(req.URL.Path); exist {
		stats = v.(*ApiStats)
	} else {
		stats = &ApiStats{}
		WebApiStats.Store(req.URL.Path, stats)
	}
	var rep []byte
	start := time.Now()
	res := make(chan []byte, 1)
	suc := core.CoreObject().SendCommand(&WebApiEvent{req: req, path: req.URL.Path, h: HandlerWrapper(func(event *WebApiEvent, data []byte) bool {
		logger.Logger.Trace("WorldSrvApi start transcate")
		tnp := &transact.TransNodeParam{
			Tt:     common.TransType_WebApi,
			Ot:     transact.TransOwnerType(common.GetSelfSrvType()),
			Oid:    common.GetSelfSrvId(),
			AreaID: common.GetSelfAreaId(),
		}
		tNode := transact.DTCModule.StartTrans(tnp, event, transact.DefaultTransactTimeout) //超时时间30秒
		if tNode != nil {
			tNode.TransEnv.SetField(WEBAPI_TRANSACTE_EVENT, event)
			tNode.Go(core.CoreObject())
		}
		return true
	}), body: data, rawQuery: req.URL.RawQuery, res: res}, false)
	if suc {
		select {
		case rep = <-res:
			if rep != nil {
				webApiResponse(rw, rep)
			}
		case <-time.After(ApiDefaultTimeout):
			//rep = make(map[string]interface{})
			//rep[webapi.RESPONSE_STATE] = webapi.STATE_ERR
			//rep[webapi.RESPONSE_ERRMSG] = "proccess timeout!"
			webApiResponse(rw, rep)
			if stats != nil {
				atomic.AddInt64(&stats.TimeoutTimes, 1)
			}
		}
	} else {
		webApiResponse(rw, nil)
		if stats != nil {
			atomic.AddInt64(&stats.UnreachTimes, 1)
		}
	}
	ps := int64(time.Now().Sub(start) / time.Millisecond)
	if stats != nil {
		atomic.AddInt64(&stats.RunTimes, 1)
		atomic.AddInt64(&stats.TotalRuningTime, ps)
		if atomic.LoadInt64(&stats.MaxRuningTime) < ps {
			atomic.StoreInt64(&stats.MaxRuningTime, ps)
		}
	}
	result, err := json.Marshal(rep)
	if err == nil {
		log := model.NewAPILog(req.URL.Path, req.URL.RawQuery, string(data[:]), req.RemoteAddr, string(result[:]), startTime, ps)
		LogChannelSington.WriteLog(log)
	}
	return
}

//--------------------------------------------------------------------------------------
func init() {
	transact.RegisteHandler(common.TransType_WebApi, &transact.TransHanderWrapper{
		OnExecuteWrapper: transact.OnExecuteWrapper(func(tNode *transact.TransNode, ud interface{}) transact.TransExeResult {
			logger.Logger.Trace("WorldSrvApi start TransType_WebApi OnExecuteWrapper ")
			tnp := &transact.TransNodeParam{
				Tt:     common.TransType_WebApi,
				Ot:     transact.TransOwnerType(srvlib.WorldServerType),
				Oid:    common.GetWorldSrvId(),
				AreaID: common.GetSelfAreaId(),
				Tct:    transact.TransactCommitPolicy_TwoPhase,
			}
			if event, ok := ud.(*WebApiEvent); ok {
				userData := &common.M2GWebApiRequest{Path: event.path, RawQuery: event.rawQuery, Body: event.body, ReqIp: event.req.RemoteAddr}
				tNode.StartChildTrans(tnp, userData, transact.DefaultTransactTimeout)

				pid := tNode.MyTnp.TId
				cid := tnp.TId
				logger.Logger.Tracef("WorldSrvApi start TransType_WebApi OnExecuteWrapper tid:%x childid:%x", pid, cid)
				return transact.TransExeResult_Success
			}
			return transact.TransExeResult_Failed
		}),
		OnCommitWrapper: transact.OnCommitWrapper(func(tNode *transact.TransNode) transact.TransExeResult {
			logger.Logger.Trace("WorldSrvApi start TransType_WebApi OnCommitWrapper")
			event := tNode.TransEnv.GetField(WEBAPI_TRANSACTE_EVENT).(*WebApiEvent)
			resp := tNode.TransEnv.GetField(WEBAPI_TRANSACTE_RESPONSE)
			if ud, ok := resp.([]byte); ok {
				event.Response(netlib.SkipHeaderGetRaw(ud))
				return transact.TransExeResult_Success
			}
			event.Response(nil /*map[string]interface{}{webapi.RESPONSE_STATE: webapi.STATE_ERR, webapi.RESPONSE_ERRMSG: "execute failed!"}*/)
			return transact.TransExeResult_Success
		}),
		OnRollBackWrapper: transact.OnRollBackWrapper(func(tNode *transact.TransNode) transact.TransExeResult {
			logger.Logger.Trace("WorldSrvApi start TransType_WebApi OnRollBackWrapper")
			event := tNode.TransEnv.GetField(WEBAPI_TRANSACTE_EVENT).(*WebApiEvent)
			resp := tNode.TransEnv.GetField(WEBAPI_TRANSACTE_RESPONSE)
			if ud, ok := resp.([]byte); ok {
				event.Response(netlib.SkipHeaderGetRaw(ud))
				return transact.TransExeResult_Success
			}
			event.Response(nil /*map[string]interface{}{webapi.RESPONSE_STATE: webapi.STATE_ERR, webapi.RESPONSE_ERRMSG: "execute failed!"}*/)
			return transact.TransExeResult_Success
		}),
		OnChildRespWrapper: transact.OnChildRespWrapper(func(tNode *transact.TransNode, hChild transact.TransNodeID, retCode int, ud interface{}) transact.TransExeResult {
			logger.Logger.Tracef("WorldSrvApi start TransType_WebApi OnChildRespWrapper ret:%v childid:%x", retCode, hChild)
			tNode.TransEnv.SetField(WEBAPI_TRANSACTE_RESPONSE, ud)
			return transact.TransExeResult(retCode)
		}),
	})
	//api注册登录，获取token
	//admin.MyAdminApp.Route("/api/Member/APIMemberRegisterOrLogin", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Member/QPAPIRegisterOrLogin", WorldSrvApi)
	//api加减币
	//admin.MyAdminApp.Route("/api/Game/APIAddSubCoinById", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Game/QPAPIAddSubCoinById", WorldSrvApi)
	//校验Hash
	admin.MyAdminApp.Route("/api/Game/CrashVerifier", WorldSrvApi)
	//获取用户注单记录游戏记录
	//admin.MyAdminApp.Route("/api/Member/GetGameHistory", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Member/QPGetGameHistory", WorldSrvApi)
	//后台修改平台数据后推送给游服
	admin.MyAdminApp.Route("/game_srv/platform_config", WorldSrvApi)
	//全局游戏开关
	admin.MyAdminApp.Route("/game_srv/update_global_game_status", WorldSrvApi)
	//更新游戏配置
	admin.MyAdminApp.Route("/game_srv/update_game_configs", WorldSrvApi)
	//更新游戏分组配置
	admin.MyAdminApp.Route("/game_srv/game_config_group", WorldSrvApi)

	//测试
	admin.MyAdminApp.Route("/api/Test", WorldSrvApi)
	//修改用户信息
	admin.MyAdminApp.Route("/api/Member/AddMemberGoldById", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Member/EditMemberName", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Member/EditMemberPwd", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Member/RestoreMemberPwd", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Member/EditMemberSafePwd", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Member/EditAlipayName", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Member/EditAlipayAccount", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Member/EditAlipayInfo", WorldSrvApi)
	//admin.MyAdminApp.Route("/api/Member/GetMemberGoldById", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Member/QPGetMemberGoldById", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Member/EditPlatform", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Member/RestoreExchangeCoin", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Member/QueryGoldChangeList", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Member/UpdateMemberPlatform", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Member/SetMemberPlatform", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Member/GetSnidByUnionid", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Member/BindPromoterTree", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Member/DeleteUser", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Member/SyncTelUser", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Member/SyncTelPwd", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Member/GetFlowExchangeList", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Member/GetCurFlowList", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Member/TotalConvertibleFlow", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Member/UpdateFlowListState", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Member/SetPlayerCanRebate", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Member/SetPlayerIsCanRename", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Member/BindInviteID", WorldSrvApi)

	//黑名单
	admin.MyAdminApp.Route("/api/Member/EditBlackList", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Member/DeleteBlackList", WorldSrvApi)

	admin.MyAdminApp.Route("/api/Report/QueryOnlineReportList", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Report/OnlineReportTotal", WorldSrvApi)

	admin.MyAdminApp.Route("/api/Game/UpsertPlatform", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Game/UpsertPlatformConfig", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Game/UpsertPackageTag", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Game/UpsertGameConfigGroup", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Game/UpsertPromoterConfig", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Game/UpsertPlatformProfitControlConfig", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Game/QueryPlatformProfitControlConfig", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Game/ClearPlatformProfitData", WorldSrvApi)

	//更新返利
	admin.MyAdminApp.Route("/api/Game/UpsertGameRebateConfig", WorldSrvApi)

	admin.MyAdminApp.Route("/api/Game/RechargeById", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Game/TransformCoin", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Game/DGAddCoinById", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Game/AddCoinById", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Game/IpPlayer", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Game/BankPlayer", WorldSrvApi)

	admin.MyAdminApp.Route("/api/Game/AddCoinByIdAndPT", WorldSrvApi)

	//水池相关
	admin.MyAdminApp.Route("/api/Game/QueryGamePoolByGameId", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Game/QueryAllGamePool", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Game/RefreshGamePool", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Game/UpdateGamePool", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Game/ResetGamePool", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Game/DTRoomInfo", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Game/DTRoomFlag", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Game/SetRoomResults", WorldSrvApi)

	admin.MyAdminApp.Route("/api/Player/PlayerData", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Player/MorePlayerData", WorldSrvApi)
	//admin.MyAdminApp.Route("/api/Player/MorePlayerData2", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Player/KickPlayer", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Player/WhiteBlackControl", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Player/PlayerList", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Player/GetPlayerNames", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Player/SetGMLevel", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Player/SignData", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Player/SetVipLevel", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Player/SetLogicLevel", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Player/ClrLogicLevel", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Player/BlackBySnId", WorldSrvApi)

	//跑马灯
	admin.MyAdminApp.Route("/api/Message/CreateHorseRaceLamp", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Message/QueryHorseRaceLampList", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Message/GetHorseRaceLampById", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Message/EditHorseRaceLamp", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Message/RemoveHorseRaceLampById", WorldSrvApi)

	//邮箱
	admin.MyAdminApp.Route("/api/Game/CreateShortMessage", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Game/QueryShortMessageList", WorldSrvApi)
	//admin.MyAdminApp.Route("/api/Game/GetShortMessageById", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Game/DeleteShortMessage", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Game/DeleteAllShortMessage", WorldSrvApi)

	//公告
	admin.MyAdminApp.Route("/api/Game/UpdateBulletinById", WorldSrvApi)
	//招商
	admin.MyAdminApp.Route("/api/Game/UpdateCustomerById", WorldSrvApi)

	//添加代理
	admin.MyAdminApp.Route("/api/Member/AgentToMember", WorldSrvApi)

	//发送手机验证码
	admin.MyAdminApp.Route("/api/Send/SendCode", WorldSrvApi)

	//代理加金币
	admin.MyAdminApp.Route("/api/Agent/GiveGift", WorldSrvApi)
	//礼物
	//admin.MyAdminApp.Route("/api/Game/CreateGift", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Game/QueryGiftList", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Game/GetGiftById", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Game/EditGift", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Game/QueryGiftCfg", WorldSrvApi)

	//获取房间列表
	admin.MyAdminApp.Route("/api/Cache/ListRoom", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Cache/ClubList", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Cache/ClubListRoom", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Cache/GetRoom", WorldSrvApi)
	//销毁房间
	admin.MyAdminApp.Route("/api/Cache/DestroyRoom", WorldSrvApi)
	//admin.MyAdminApp.Route("/api/Cache/DestroyMoreRoom", WorldSrvApi)
	//修改内存属性
	admin.MyAdminApp.Route("/api/Cache/UpdatePlayerElement", WorldSrvApi)

	//MemoryCache
	admin.MyAdminApp.Route("/api/Cache/Get", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Cache/Put", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Cache/Delete", WorldSrvApi)
	//DG API
	//admin.MyAdminApp.Route("/dgapi/user/getBalance", WorldSrvApi)
	//admin.MyAdminApp.Route("/dgapi/account/transfer", WorldSrvApi)
	//admin.MyAdminApp.Route("/dgapi/account/inform", WorldSrvApi)
	//admin.MyAdminApp.Route("/dgapi/account/order", WorldSrvApi)
	//admin.MyAdminApp.Route("/dgapi/account/unsettle", WorldSrvApi)

	//Act Sign
	admin.MyAdminApp.Route("/api/act/UpdateSignConfig", WorldSrvApi)
	//Act Gold
	admin.MyAdminApp.Route("/api/act/UpdateGoldTaskConfig", WorldSrvApi)
	admin.MyAdminApp.Route("/api/act/UpdateGoldComeConfig", WorldSrvApi)

	admin.MyAdminApp.Route("/api/act/UpdateOnlineRewardConfig", WorldSrvApi)
	admin.MyAdminApp.Route("/api/act/UpdateLuckyTurntableConfig", WorldSrvApi)
	admin.MyAdminApp.Route("/api/act/UpdateShareConfig", WorldSrvApi)
	admin.MyAdminApp.Route("/api/act/UpdateYebConfig", WorldSrvApi)
	admin.MyAdminApp.Route("/api/act/GetLuckyTurntableData", WorldSrvApi)

	admin.MyAdminApp.Route("/api/act/UpdateCardConfig", WorldSrvApi)
	admin.MyAdminApp.Route("/api/act/UpdateStepRechargeConfig", WorldSrvApi)
	admin.MyAdminApp.Route("/api/act/CardBuy", WorldSrvApi)
	admin.MyAdminApp.Route("/api/act/StepRechargeBuy", WorldSrvApi)
	admin.MyAdminApp.Route("/api/act/WeiXinShare", WorldSrvApi)
	admin.MyAdminApp.Route("/api/act/UpdateRandCoinSetting", WorldSrvApi)
	admin.MyAdminApp.Route("/api/act/GetRandCoinLog", WorldSrvApi)
	admin.MyAdminApp.Route("/api/act/SwitchRandCoinOnOff", WorldSrvApi)

	//act vip
	admin.MyAdminApp.Route("/api/act/UpdateVipConfig", WorldSrvApi)
	admin.MyAdminApp.Route("/api/act/UpdatePayActIsPay", WorldSrvApi)

	//act fpay
	admin.MyAdminApp.Route("/api/act/UpdateFpayConfig", WorldSrvApi)
	admin.MyAdminApp.Route("/api/act/OpFpayUser", WorldSrvApi)

	//act give
	admin.MyAdminApp.Route("/api/act/UpdateActGiveConfig", WorldSrvApi)

	admin.MyAdminApp.Route("/api/act/UpdatePayActConfig", WorldSrvApi)
	//活跃任务
	admin.MyAdminApp.Route("/api/act/UpdateTaskConfig", WorldSrvApi)

	//客服系统
	admin.MyAdminApp.Route("/api/Customer/PushOfflineMsg", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Customer/GetPlayerByToken", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Customer/PlatformClean", WorldSrvApi)
	//兑换物品
	admin.MyAdminApp.Route("/api/Customer/UpExchangeStatus", WorldSrvApi)
	//系统接口
	admin.MyAdminApp.Route("/api/System/Ping", WorldSrvApi)
	//3方平台
	admin.MyAdminApp.Route("/api/thd/GetThridPlatform", WorldSrvApi)
	admin.MyAdminApp.Route("/api/thd/PlatformAddCoin", WorldSrvApi)
	admin.MyAdminApp.Route("/api/thd/GeeCheck", WorldSrvApi)
	admin.MyAdminApp.Route("/api/thd/UpdatePlayerCoin", WorldSrvApi)
	admin.MyAdminApp.Route("/api/thd/SetPlatformCoin", WorldSrvApi)

	//俱乐部相关
	admin.MyAdminApp.Route("/api/Club/ClubCreatorBaseInfo", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Club/ClubCreateCheck", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Club/ClubNoticeCheck", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Club/ClubRoomSet", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Club/ClubAreaSet", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Club/ClubMemberSet", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Club/ClubBaseInfo", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Club/ClubSet", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Club/ClubOutline", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Club/RefreshWaitCheckClubInfo", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Club/PlayerClubRelation", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Club/AddClubGoldById", WorldSrvApi)

	//修改玩家备注
	admin.MyAdminApp.Route("/api/Member/EditMemberMarkInfo", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Member/EditMemberTelephonePromoter", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Member/EditMemberTelephoneCallNum", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Member/EditWhiteFlag", WorldSrvApi)

	//比赛相关
	admin.MyAdminApp.Route("/api/Match/UpdateMatchConfig", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Match/DeleteMatchConfig", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Match/QueryRunningMatch", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Match/DestroyRunningMatch", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Match/QueryNotStartedMatch", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Match/CancelSignupNotStartedMatch", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Match/AddTicket", WorldSrvApi)
	admin.MyAdminApp.Route("/api/act/UpdateTicketConfig", WorldSrvApi)
	admin.MyAdminApp.Route("/api/act/ManualDistTicket", WorldSrvApi)

	admin.MyAdminApp.Route("/api/Game/UpsertGradeShopConfig", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Game/HandleOrderByGradeShop", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Game/UpsertLogicLevelConfig", WorldSrvApi)

	//用户行为监听
	admin.MyAdminApp.Route("/api/Player/EditActMonitorInfo", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Player/DelActMonitorInfo", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Player/QueryAMIList", WorldSrvApi)

	//用户分层
	admin.MyAdminApp.Route("/api/Player/UpdateLayered", WorldSrvApi)

	//单控
	admin.MyAdminApp.Route("/api/Game/SinglePlayerAdjust", WorldSrvApi)

	//兑换卷
	admin.MyAdminApp.Route("/api/Game/CreateJYB", WorldSrvApi)
	admin.MyAdminApp.Route("/api/Game/UpdateJYB", WorldSrvApi)
}

func Stats() map[string]ApiStats {
	stats := make(map[string]ApiStats)
	WebApiStats.Range(func(k, v interface{}) bool {
		if s, ok := v.(*ApiStats); ok {
			ss := *s //计数可能不精准
			stats[k.(string)] = ss
		}
		return true
	})
	return stats
}
