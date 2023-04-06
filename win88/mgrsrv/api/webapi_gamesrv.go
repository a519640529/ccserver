package api

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"games.yol.com/win88/common"
	"games.yol.com/win88/model"
	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/core/transact"
	"github.com/idealeak/goserver/core/utils"
	"github.com/idealeak/goserver/srvlib"
	"io"
	"io/ioutil"
	"net/http"
	"sync/atomic"
	"time"
)

const (
	GAMESRVAPI_TRANSACTE_EVENT    = "GAMESRVAPI_TRANSACTE_EVENT"
	GAMESRVAPI_TRANSACTE_RESPONSE = "GAMESRVAPI_TRANSACTE_RESPONSE"
)

//
//type ResponseData struct {
//	State int
//	Data  string
//}
//
func GameSrvWebAPI(rw http.ResponseWriter, req *http.Request) {
	defer utils.DumpStackIfPanic("api.GameSrvApi")
	logger.Logger.Info("GameSrvApi receive:", req.URL.Path, req.URL.RawQuery)

	if common.RequestCheck(req, model.GameParamData.WhiteHttpAddr) == false {
		logger.Logger.Info("RemoteAddr [%v] require api.", req.RemoteAddr)
		return
	}
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		webApiResponse(rw, nil)
		return
	}
	m := req.URL.Query()
	timestamp := m.Get("nano")
	if timestamp == "" {
		logger.Logger.Info(req.RemoteAddr, " GameSrvApi param error: nano not allow null")
		return
	}
	sign := m.Get("sign")
	if sign == "" {
		logger.Logger.Info(req.RemoteAddr, " GameSrvApi param error: sign not allow null")
		return
	}
	startTime := time.Now().UnixNano()
	args := fmt.Sprintf("%v;%v;%v;%v", common.Config.AppId, req.URL.Path, string(data), timestamp)
	h := md5.New()
	io.WriteString(h, args)
	realSign := hex.EncodeToString(h.Sum(nil))
	if realSign != sign && !common.Config.IsDevMode {
		logger.Logger.Info(req.RemoteAddr, " srvCtrlMain sign error: expect ", realSign, " ; but get ", sign)
		webApiResponse(rw, nil)
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
		logger.Logger.Trace("GameSrvApi start transcate")
		tnp := &transact.TransNodeParam{
			Tt:     common.TransType_GameSrvWebApi,
			Ot:     transact.TransOwnerType(common.GetSelfSrvType()),
			Oid:    common.GetSelfSrvId(),
			AreaID: common.GetSelfAreaId(),
		}
		tNode := transact.DTCModule.StartTrans(tnp, event, transact.DefaultTransactTimeout)
		if tNode != nil {
			tNode.TransEnv.SetField(GAMESRVAPI_TRANSACTE_EVENT, event)
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

////--------------------------------------------------------------------------------------
func init() {
	transact.RegisteHandler(common.TransType_GameSrvWebApi, &transact.TransHanderWrapper{
		OnExecuteWrapper: transact.OnExecuteWrapper(func(tNode *transact.TransNode, ud interface{}) transact.TransExeResult {
			logger.Logger.Trace("GameSrvApi start TransType_GameSrvWebApi OnExecuteWrapper")
			gameSrvIds := common.GetGameSrvIds()
			logger.Logger.Trace("Current game id:", gameSrvIds)
			for _, value := range gameSrvIds {
				tnp := &transact.TransNodeParam{
					Tt:     common.TransType_GameSrvWebApi,
					Ot:     transact.TransOwnerType(srvlib.GameServerType),
					Oid:    value,
					AreaID: common.GetSelfAreaId(),
					Tct:    transact.TransactCommitPolicy_TwoPhase,
				}
				if event, ok := ud.(*WebApiEvent); ok {
					userData := &common.M2GWebApiRequest{Path: event.path, RawQuery: event.rawQuery, Body: event.body, ReqIp: event.req.RemoteAddr}
					ter := tNode.StartChildTrans(tnp, userData, transact.DefaultTransactTimeout)
					if ter != transact.TransExeResult_Success {
						logger.Logger.Tracef("StartChildTrans %v game server failed.", value)
					}
				} else {
					logger.Logger.Trace("Conver ud to WebApiEvent failed OnExecuteWrapper")
				}
			}
			return transact.TransExeResult_Success
		}),
		OnCommitWrapper: transact.OnCommitWrapper(func(tNode *transact.TransNode) transact.TransExeResult {
			logger.Logger.Trace("GameSrvApi start TransType_GameSrvWebApi OnCommitWrapper")
			event := tNode.TransEnv.GetField(GAMESRVAPI_TRANSACTE_EVENT).(*WebApiEvent)
			resp := tNode.TransEnv.GetField(GAMESRVAPI_TRANSACTE_RESPONSE)
			if ud, ok := resp.([]byte); ok {
				event.Response(netlib.SkipHeaderGetRaw(ud))
				return transact.TransExeResult_Success
			}
			event.Response(nil /*map[string]interface{}{webapi.RESPONSE_STATE: webapi.STATE_ERR, webapi.RESPONSE_ERRMSG: "execute failed!"}*/)
			return transact.TransExeResult_Success
		}),
		OnRollBackWrapper: transact.OnRollBackWrapper(func(tNode *transact.TransNode) transact.TransExeResult {
			logger.Logger.Trace("GameSrvApi start TransType_GameSrvWebApi OnRollBackWrapper")
			event := tNode.TransEnv.GetField(GAMESRVAPI_TRANSACTE_EVENT).(*WebApiEvent)
			resp := tNode.TransEnv.GetField(GAMESRVAPI_TRANSACTE_RESPONSE)
			if ud, ok := resp.([]byte); ok {
				event.Response(netlib.SkipHeaderGetRaw(ud))
				return transact.TransExeResult_Success
			}
			event.Response(nil)
			return transact.TransExeResult_Success
		}),
		OnChildRespWrapper: transact.OnChildRespWrapper(func(tNode *transact.TransNode, hChild transact.TransNodeID, retCode int, ud interface{}) transact.TransExeResult {
			logger.Logger.Tracef("GameSrvApi OnChildRespWrapper %v:%v", hChild, ud)
			tNode.TransEnv.SetField(GAMESRVAPI_TRANSACTE_RESPONSE, ud)
			return transact.TransExeResult(retCode)
		}),
	})
	//	//参数设置
	//	admin.MyAdminApp.Route("/api/Param/CommonTax", GameSrvWebAPI)
	//	//捕鱼金币池查询
	//	admin.MyAdminApp.Route("/api/CoinPool/FishingPool", GameSrvWebAPI)
	//	//通用金币池查询
	//	admin.MyAdminApp.Route("/api/CoinPool/GamePool", GameSrvWebAPI)
	//	//捕鱼渔场保留金币
	//	admin.MyAdminApp.Route("/api/CoinPool/GameFishsAllCoin", GameSrvWebAPI)
	//	//单控数据
	//admin.MyAdminApp.Route("/api/Game/SinglePlayerAdjust", GameSrvWebAPI)
}
