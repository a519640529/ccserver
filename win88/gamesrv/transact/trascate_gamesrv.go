package transact

//
//import (
//	"errors"
//	"fmt"
//	"games.yol.com/win88/common"
//	"games.yol.com/win88/gamesrv/base"
//	"games.yol.com/win88/model"
//	"games.yol.com/win88/proto"
//	webapi_proto "games.yol.com/win88/protocol/webapi"
//	"github.com/idealeak/goserver/core/basic"
//	"github.com/idealeak/goserver/core/logger"
//	"github.com/idealeak/goserver/core/netlib"
//	"github.com/idealeak/goserver/core/task"
//	"github.com/idealeak/goserver/core/transact"
//	"sync"
//)
//
//const __REQIP__ = "__REQIP__"
//
//var (
//	WebAPIErrParam    = errors.New("param err")
//	WebAPIErrNoPlayer = errors.New("player no find")
//)
//
//func init() {
//	transact.RegisteHandler(common.TransType_GameSrvWebApi, &WebAPITranscateHandler{})
//}
//
//var WebAPIHandlerMgrSingleton = &WebAPIHandlerMgr{wshMap: make(map[string]WebAPIHandler)}
//
//type WebAPITranscateHandler struct {
//}
//
//func (this *WebAPITranscateHandler) OnExcute(tNode *transact.TransNode, ud interface{}) transact.TransExeResult {
//	logger.Logger.Trace("WebAPITranscateHandler.OnExcute ")
//	req := &common.M2GWebApiRequest{}
//	err := netlib.UnmarshalPacketNoPackId(ud.([]byte), req)
//	if err == nil {
//		wsh := WebAPIHandlerMgrSingleton.GetWebAPIHandler(req.Path)
//		if wsh == nil {
//			logger.Logger.Error("WebAPITranscateHandler no registe WebAPIHandler ", req.Path)
//			return transact.TransExeResult_Failed
//		}
//		tag, msg := wsh.Handler(tNode, req.Body)
//		tNode.TransRep.RetFiels = msg
//		switch tag {
//		case common.ResponseTag_Ok:
//			return transact.TransExeResult_Success
//		case common.ResponseTag_TransactYield:
//			return transact.TransExeResult_Yield
//		}
//	}
//	logger.Logger.Error("WebAPITranscateHandler.OnExcute err:", err.Error())
//	return transact.TransExeResult_Failed
//}
//
//func (this *WebAPITranscateHandler) OnCommit(tNode *transact.TransNode) transact.TransExeResult {
//	logger.Logger.Trace("WebAPITranscateHandler.OnCommit ")
//	return transact.TransExeResult_Success
//}
//
//func (this *WebAPITranscateHandler) OnRollBack(tNode *transact.TransNode) transact.TransExeResult {
//	logger.Logger.Trace("WebAPITranscateHandler.OnRollBack ")
//	return transact.TransExeResult_Success
//}
//
//func (this *WebAPITranscateHandler) OnChildTransRep(tNode *transact.TransNode, hChild transact.TransNodeID, retCode int,
//	ud interface{}) transact.TransExeResult {
//	logger.Logger.Trace("WebAPITranscateHandler.OnChildTransRep ")
//	return transact.TransExeResult_Success
//}
//
//type WebAPIHandler interface {
//	Handler(*transact.TransNode, []byte) (int, proto.Message)
//}
//
//type WebAPIHandlerWrapper func(*transact.TransNode, []byte) (int, proto.Message)
//
//func (wshw WebAPIHandlerWrapper) Handler(tNode *transact.TransNode, params []byte) (int, proto.Message) {
//	return wshw(tNode, params)
//}
//
//type WebAPIHandlerMgr struct {
//	wshMap       map[string]WebAPIHandler
//	DataWaitList sync.Map
//}
//
//func (this *WebAPIHandlerMgr) RegisteWebAPIHandler(name string, wsh WebAPIHandler) {
//	this.wshMap[name] = wsh
//}
//
//func (this *WebAPIHandlerMgr) GetWebAPIHandler(name string) WebAPIHandler {
//	if wsh, exist := this.wshMap[name]; exist {
//		return wsh
//	}
//	return nil
//}
//
//func init() {
//	//单控
//	WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/api/Game/SinglePlayerAdjust", WebAPIHandlerWrapper(
//		func(tNode *transact.TransNode, params []byte) (int, proto.Message) {
//			pack := &webapi_proto.SASinglePlayerAdjust{}
//			msg := &webapi_proto.ASSinglePlayerAdjust{}
//			err := proto.Unmarshal(params, msg)
//			if err != nil {
//				fmt.Printf("err:%v", err)
//				pack.Tag = webapi_proto.TagCode_FAILED
//				pack.Msg = "数据序列化失败"
//				return common.ResponseTag_ParamError, pack
//			}
//			pack.Tag = webapi_proto.TagCode_SUCCESS
//			switch msg.GetOpration() {
//			case 1:
//				psa := base.PlayerSingleAdjustMgr.AddNewSingleAdjust(msg.GetPlayerSingleAdjust())
//				if psa != nil {
//					task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
//						return model.AddNewSingleAdjust(psa)
//					}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
//						if data != nil {
//							pack.Tag = webapi_proto.TagCode_FAILED
//							pack.Msg = "insert err" + data.(error).Error()
//						}
//						tNode.TransRep.RetFiels = pack
//						tNode.Resume()
//					}), "AddNewSingleAdjust").Start()
//					return common.ResponseTag_TransactYield, pack
//				}
//			case 2:
//				base.PlayerSingleAdjustMgr.EditSingleAdjust(msg.GetPlayerSingleAdjust())
//			case 3:
//				psa := msg.PlayerSingleAdjust
//				if psa != nil {
//					base.PlayerSingleAdjustMgr.DeleteSingleAdjust(psa.Platform, psa.SnId, psa.GameFreeId)
//				}
//			case 4:
//				ps := msg.PlayerSingleAdjust
//				webp := base.PlayerSingleAdjustMgr.GetSingleAdjust(ps.Platform, ps.SnId, ps.GameFreeId)
//				if webp == nil {
//					pack.Tag = webapi_proto.TagCode_FAILED
//					pack.Msg = fmt.Sprintf("webp == nil %v %v %v", ps.Platform, ps.SnId, ps.GameFreeId)
//				}
//				pack.PlayerSingleAdjust = webp
//			default:
//				pack.Tag = webapi_proto.TagCode_FAILED
//				pack.Msg = "Opration param is error!"
//				return common.ResponseTag_ParamError, pack
//			}
//			return common.ResponseTag_Ok, pack
//		}))
//}
