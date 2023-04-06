package api

//
//import (
//	"encoding/json"
//	"io/ioutil"
//	"net/http"
//	"time"
//
//	"crypto/md5"
//	"encoding/hex"
//	"fmt"
//	"games.yol.com/win88/common"
//	"games.yol.com/win88/model"
//	"games.yol.com/win88/webapi"
//	"github.com/idealeak/goserver/core"
//	"github.com/idealeak/goserver/core/admin"
//	"github.com/idealeak/goserver/core/logger"
//	"github.com/idealeak/goserver/core/netlib"
//	"github.com/idealeak/goserver/core/transact"
//	"github.com/idealeak/goserver/core/utils"
//	"io"
//	"sync/atomic"
//)
//
//// API
//// http://127.0.0.1:9595/api/Report/QueryOnlineReportList?ts=20141024000000&sign=41cc8cee8dd93f7dc70b6426cfd1029d
//
//func ActThrSrvApi(rw http.ResponseWriter, req *http.Request) {
//	defer utils.DumpStackIfPanic("api.ActThrSrvApi")
//	logger.Logger.Info("ActThrSrvApi receive:", req.URL.Path, req.URL.RawQuery)
//
//	if common.RequestCheck(req, model.GameParamData.WhiteHttpAddr) == false {
//		logger.Logger.Info("RemoteAddr [%v] require api.", req.RemoteAddr)
//		return
//	}
//	data, err := ioutil.ReadAll(req.Body)
//	if err != nil {
//		logger.Logger.Info("Body err.", err)
//		webApiResponse(rw, map[string]interface{}{webapi.RESPONSE_STATE: webapi.STATE_ERR, webapi.RESPONSE_ERRMSG: "Post data is null!"})
//		return
//	}
//	m := req.URL.Query()
//	timestamp := m.Get("nano")
//	if timestamp == "" {
//		logger.Logger.Info(req.RemoteAddr, " ActThrSrvApi param error: nano not allow null")
//		return
//	}
//	sign := m.Get("sign")
//	if sign == "" {
//		logger.Logger.Info(req.RemoteAddr, " ActThrSrvApi param error: sign not allow null")
//		return
//	}
//	startTime := time.Now().UnixNano()
//	args := fmt.Sprintf("%v;%v;%v;%v", common.Config.AppId, req.URL.Path, string(data), timestamp)
//	h := md5.New()
//	io.WriteString(h, args)
//	realSign := hex.EncodeToString(h.Sum(nil))
//	if realSign != sign && !common.Config.IsDevMode {
//		logger.Logger.Info(req.RemoteAddr, " srvCtrlMain sign error: expect ", realSign, " ; but get ", sign, " raw=", args)
//		webApiResponse(rw, map[string]interface{}{webapi.RESPONSE_STATE: webapi.STATE_ERR, webapi.RESPONSE_ERRMSG: "Sign error!"})
//		return
//	}
//
//	var stats *ApiStats
//	if v, exist := WebApiStats.Load(req.URL.Path); exist {
//		stats = v.(*ApiStats)
//	} else {
//		stats = &ApiStats{}
//		WebApiStats.Store(req.URL.Path, stats)
//	}
//	var rep map[string]interface{}
//	start := time.Now()
//	res := make(chan map[string]interface{}, 1)
//	core.CoreObject().SendCommand(&WebApiEvent{req: req, path: req.URL.Path, h: HandlerWrapper(func(event *WebApiEvent, data []byte) bool {
//		logger.Logger.Trace("ActThrSrvApi start transcate")
//		tnp := &transact.TransNodeParam{
//			Tt:     common.TransType_ActThrSrvWebApi,
//			Ot:     transact.TransOwnerType(common.GetSelfSrvType()),
//			Oid:    common.GetSelfSrvId(),
//			AreaID: common.GetSelfAreaId(),
//		}
//		tNode := transact.DTCModule.StartTrans(tnp, event, transact.DefaultTransactTimeout) //超时时间30秒
//		if tNode != nil {
//			tNode.TransEnv.SetField(WEBAPI_TRANSACTE_EVENT, event)
//			tNode.Go(core.CoreObject())
//		}
//		return true
//	}), body: data, rawQuery: req.URL.RawQuery, res: res}, false)
//	select {
//	case rep = <-res:
//		if rep != nil {
//			webApiResponse(rw, rep)
//		}
//	case <-time.After(ApiDefaultTimeout):
//		rep = make(map[string]interface{})
//		rep[webapi.RESPONSE_STATE] = webapi.STATE_ERR
//		rep[webapi.RESPONSE_ERRMSG] = "proccess timeout!"
//		webApiResponse(rw, rep)
//		if stats != nil {
//			atomic.AddInt64(&stats.TimeoutTimes, 1)
//		}
//	}
//	ps := int64(time.Now().Sub(start) / time.Millisecond)
//	if stats != nil {
//		atomic.AddInt64(&stats.RunTimes, 1)
//		atomic.AddInt64(&stats.TotalRuningTime, ps)
//		if atomic.LoadInt64(&stats.MaxRuningTime) < ps {
//			atomic.StoreInt64(&stats.MaxRuningTime, ps)
//		}
//	}
//	result, err := json.Marshal(rep)
//	if err == nil {
//		log := model.NewAPILog(req.URL.Path, req.URL.RawQuery, string(data[:]), req.RemoteAddr, string(result[:]), startTime, ps)
//		APILogChannelSington.Write(log)
//	}
//	return
//}
//
////--------------------------------------------------------------------------------------
//func init() {
//	transact.RegisteHandler(common.TransType_ActThrSrvWebApi, &transact.TransHanderWrapper{
//		OnExecuteWrapper: transact.OnExecuteWrapper(func(tNode *transact.TransNode, ud interface{}) transact.TransExeResult {
//			logger.Logger.Trace("ActThrSrvApi start TransType_WebApi OnExecuteWrapper ")
//			tnp := &transact.TransNodeParam{
//				Tt:     common.TransType_ActThrSrvWebApi,
//				Ot:     transact.TransOwnerType(common.ActThrServerType),
//				Oid:    common.ActThrServerID,
//				AreaID: common.GetSelfAreaId(),
//				Tct:    transact.TransactCommitPolicy_TwoPhase,
//			}
//			if event, ok := ud.(*WebApiEvent); ok {
//				userData := &common.M2GWebApiRequest{Path: event.path, RawQuery: event.rawQuery, Body: event.body, ReqIp: event.req.RemoteAddr}
//				tNode.StartChildTrans(tnp, userData, transact.DefaultTransactTimeout)
//
//				pid := tNode.MyTnp.TId
//				cid := tnp.TId
//				logger.Logger.Tracef("ActThrSrvApi start TransType_WebApi OnExecuteWrapper tid:%x childid:%x", pid, cid)
//				return transact.TransExeResult_Success
//			}
//			return transact.TransExeResult_Failed
//		}),
//		OnCommitWrapper: transact.OnCommitWrapper(func(tNode *transact.TransNode) transact.TransExeResult {
//			logger.Logger.Trace("ActThrSrvApi start TransType_WebApi OnCommitWrapper")
//			event := tNode.TransEnv.GetField(WEBAPI_TRANSACTE_EVENT).(*WebApiEvent)
//			resp := tNode.TransEnv.GetField(WEBAPI_TRANSACTE_RESPONSE)
//			if userData, ok := resp.(*common.M2GWebApiResponse); ok {
//				if len(userData.Body) > 0 {
//					m := make(map[string]interface{})
//					err := json.Unmarshal(userData.Body, &m)
//					if err == nil {
//						event.Response(m)
//						return transact.TransExeResult_Success
//					}
//				}
//			}
//			event.Response(map[string]interface{}{webapi.RESPONSE_STATE: webapi.STATE_ERR, webapi.RESPONSE_ERRMSG: "execute failed!"})
//			return transact.TransExeResult_Success
//		}),
//		OnRollBackWrapper: transact.OnRollBackWrapper(func(tNode *transact.TransNode) transact.TransExeResult {
//			logger.Logger.Trace("ActThrSrvApi start TransType_WebApi OnRollBackWrapper")
//			event := tNode.TransEnv.GetField(WEBAPI_TRANSACTE_EVENT).(*WebApiEvent)
//			resp := tNode.TransEnv.GetField(WEBAPI_TRANSACTE_RESPONSE)
//			if userData, ok := resp.(*common.M2GWebApiResponse); ok {
//				if len(userData.Body) > 0 {
//					m := make(map[string]interface{})
//					err := json.Unmarshal(userData.Body, &m)
//					if err == nil {
//						event.Response(m)
//						return transact.TransExeResult_Success
//					}
//				}
//				return transact.TransExeResult_Success
//			}
//			event.Response(map[string]interface{}{webapi.RESPONSE_STATE: webapi.STATE_ERR, webapi.RESPONSE_ERRMSG: "execute failed!"})
//			return transact.TransExeResult_Success
//		}),
//		OnChildRespWrapper: transact.OnChildRespWrapper(func(tNode *transact.TransNode, hChild transact.TransNodeID, retCode int, ud interface{}) transact.TransExeResult {
//			logger.Logger.Tracef("ActThrSrvApi start TransType_WebApi OnChildRespWrapper ret:%v childid:%x", retCode, hChild)
//			userData := &common.M2GWebApiResponse{}
//			err := netlib.UnmarshalPacketNoPackId(ud.([]byte), userData)
//			if err == nil {
//				tNode.TransEnv.SetField(WEBAPI_TRANSACTE_RESPONSE, userData)
//			} else {
//				logger.Logger.Trace("trascate.OnChildRespWrapper err:", err)
//			}
//			return transact.TransExeResult(retCode)
//		}),
//	})
//
//	//测试
//	admin.MyAdminApp.Route("/api/ActThr/Test", ActThrSrvApi)
//
//	//增加自动黑白名单控制
//	admin.MyAdminApp.Route("/api/ActThr/Upsert", ActThrSrvApi)
//
//	//删除自动黑白名单控制
//	admin.MyAdminApp.Route("/api/ActThr/Delete", ActThrSrvApi)
//
//	//删除所有添加的黑名单
//	admin.MyAdminApp.Route("/api/ActThr/ResetBW", ActThrSrvApi)
//
//}
