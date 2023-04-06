package main

//
//import (
//	"fmt"
//	"games.yol.com/win88/common"
//	"games.yol.com/win88/webapi"
//	"github.com/idealeak/goserver/core/transact"
//)
//
//func init() {
//	WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/api/Cache/Get", WebAPIHandlerWrapper(
//		func(tNode *transact.TransNode, params webapi.RequestBody) (int, webapi.ResponseBody) {
//			params_data, _ := params.GetRequestBody("Param")
//			resp := webapi.NewResponseBody()
//			key, _ := params_data.GetStr("Key") //必填
//			resp_data := make(map[string]interface{})
//			resp_data["Key"] = key
//			if CacheMemory.IsExist(key) {
//				val := CacheMemory.Get(key)
//				resp_data["Val"] = val
//				resp_data["IsExist"] = true
//			} else {
//				resp_data["IsExist"] = false
//			}
//			resp[webapi.RESPONSE_STATE] = webapi.STATE_OK
//			resp[webapi.RESPONSE_DATA] = resp_data
//			return common.ResponseTag_Ok, resp
//		}))
//
//	WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/api/Cache/Put", WebAPIHandlerWrapper(
//		func(tNode *transact.TransNode, params webapi.RequestBody) (int, webapi.ResponseBody) {
//			params_data, _ := params.GetRequestBody("Param")
//			resp := webapi.NewResponseBody()
//			key, _ := params_data.GetStr("Key") //必填
//			val, _ := params_data.GetStr("Val")
//			timeout, _ := params_data.GetInt64("Timeout")
//			CacheMemory.Put(key, val, timeout)
//			resp[webapi.RESPONSE_STATE] = webapi.STATE_OK
//			resp[webapi.RESPONSE_DATA] = "ok"
//			return common.ResponseTag_Ok, resp
//		}))
//
//	WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/api/Cache/Del", WebAPIHandlerWrapper(
//		func(tNode *transact.TransNode, params webapi.RequestBody) (int, webapi.ResponseBody) {
//			params_data, _ := params.GetRequestBody("Param")
//			resp := webapi.NewResponseBody()
//			key, _ := params_data.GetStr("Key") //必填
//			if CacheMemory.IsExist(key) {
//				val := CacheMemory.Get(key)
//				resp_data := make(map[string]interface{})
//				resp_data["Key"] = key
//				resp_data["Val"] = val
//				resp[webapi.RESPONSE_DATA] = resp_data
//				CacheMemory.Delete(key)
//			} else {
//				resp[webapi.RESPONSE_STATE] = webapi.STATE_OK
//				resp[webapi.RESPONSE_DATA] = fmt.Sprintf("key:%v not exist", key)
//			}
//			return common.ResponseTag_Ok, resp
//		}))
//}
