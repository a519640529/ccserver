package main

//
//import (
//	"encoding/json"
//	"errors"
//	"games.yol.com/win88/common"
//	"games.yol.com/win88/model"
//	"games.yol.com/win88/webapi"
//	"github.com/idealeak/goserver/core/basic"
//	"github.com/idealeak/goserver/core/i18n"
//	"github.com/idealeak/goserver/core/logger"
//	"github.com/idealeak/goserver/core/task"
//	"github.com/idealeak/goserver/core/transact"
//	"time"
//)
//
////比赛相关提供给后台的API
//func init() {
//	//-------------------------------------------------------------------------------------------------------
//	//更新比赛配置
//	WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/api/Match/UpdateMatchConfig", WebAPIHandlerWrapper(
//		func(tNode *transact.TransNode, params webapi.RequestBody) (int, webapi.ResponseBody) {
//			resp := webapi.NewResponseBody()
//			params_data, _ := params.GetRequestBody("Param")
//			if data, ok := params_data.GetRequestBody("Data"); ok {
//				buf, err := json.Marshal(data)
//				if err != nil {
//					resp[webapi.RESPONSE_STATE] = webapi.STATE_ERR
//					resp[webapi.RESPONSE_ERRMSG] = "params error!"
//					return common.ResponseTag_ParamError, resp
//				}
//
//				var cfg MatchConfig
//				err = json.Unmarshal(buf, &cfg)
//				if err != nil {
//					resp[webapi.RESPONSE_STATE] = webapi.STATE_ERR
//					resp[webapi.RESPONSE_ERRMSG] = "params error!"
//					return common.ResponseTag_ParamError, resp
//				}
//
//				MatchMgrSington.UpdateConfig(&cfg, false)
//				resp[webapi.RESPONSE_STATE] = webapi.STATE_OK
//				return common.ResponseTag_Ok, resp
//			}
//
//			resp[webapi.RESPONSE_STATE] = webapi.STATE_ERR
//			resp[webapi.RESPONSE_ERRMSG] = "need Data!"
//			return common.ResponseTag_ParamError, resp
//		}))
//
//	//-------------------------------------------------------------------------------------------------------
//	//删除比赛配置
//	WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/api/Match/DeleteMatchConfig", WebAPIHandlerWrapper(
//		func(tNode *transact.TransNode, params webapi.RequestBody) (int, webapi.ResponseBody) {
//			resp := webapi.NewResponseBody()
//			params_data, _ := params.GetRequestBody("Param")
//			platform, _ := params_data.GetStr("Platform")
//			matchId, _ := params_data.GetInt("MatchId")
//
//			m := MatchMgrSington.GetMatch(int32(matchId))
//			if m != nil {
//				if platform != "" && m.dbMatch.Platform != platform {
//					resp[webapi.RESPONSE_STATE] = webapi.STATE_ERR
//					resp[webapi.RESPONSE_ERRMSG] = "limits of authority!"
//					return common.ResponseTag_OpFailed, resp
//				}
//			}
//
//			MatchMgrSington.DeleteMatch(int32(matchId))
//			resp[webapi.RESPONSE_STATE] = webapi.STATE_OK
//			return common.ResponseTag_Ok, resp
//		}))
//
//	//-------------------------------------------------------------------------------------------------------
//	//获取当前未开始的比赛
//	WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/api/Match/QueryNotStartedMatch", WebAPIHandlerWrapper(
//		func(tNode *transact.TransNode, params webapi.RequestBody) (int, webapi.ResponseBody) {
//			resp := webapi.NewResponseBody()
//			params_data, _ := params.GetRequestBody("Param")
//			platform, _ := params_data.GetStr("Platform")
//			data := MatchMgrSington.MarshalAllNotStartedMatch(platform)
//			resp[webapi.RESPONSE_STATE] = webapi.STATE_OK
//			resp[webapi.RESPONSE_DATA] = data
//			return common.ResponseTag_Ok, resp
//		}))
//
//	//-------------------------------------------------------------------------------------------------------
//	//强制解散正在进行中的比赛
//	WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/api/Match/CancelSignupNotStartedMatch", WebAPIHandlerWrapper(
//		func(tNode *transact.TransNode, params webapi.RequestBody) (int, webapi.ResponseBody) {
//			resp := webapi.NewResponseBody()
//			params_data, _ := params.GetRequestBody("Param")
//			platform, _ := params_data.GetStr("Platform")
//			matchId, _ := params_data.GetInt("MatchId")
//
//			m := MatchMgrSington.GetMatch(int32(matchId))
//			if m != nil {
//				if platform != "" && m.dbMatch.Platform != platform {
//					resp[webapi.RESPONSE_STATE] = webapi.STATE_ERR
//					resp[webapi.RESPONSE_ERRMSG] = "limits of authority!"
//					return common.ResponseTag_OpFailed, resp
//				}
//
//				m.ForceCancelSignupAll()
//			}
//
//			resp[webapi.RESPONSE_STATE] = webapi.STATE_OK
//			return common.ResponseTag_Ok, resp
//		}))
//
//	//-------------------------------------------------------------------------------------------------------
//	//获取当前正在进行中的比赛
//	WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/api/Match/QueryRunningMatch", WebAPIHandlerWrapper(
//		func(tNode *transact.TransNode, params webapi.RequestBody) (int, webapi.ResponseBody) {
//			resp := webapi.NewResponseBody()
//			params_data, _ := params.GetRequestBody("Param")
//			platform, _ := params_data.GetStr("Platform")
//			data := MatchMgrSington.MarshalAllRunningMatch(platform)
//			resp[webapi.RESPONSE_STATE] = webapi.STATE_OK
//			resp[webapi.RESPONSE_DATA] = data
//			return common.ResponseTag_Ok, resp
//		}))
//
//	//-------------------------------------------------------------------------------------------------------
//	//强制解散正在进行中的比赛
//	WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/api/Match/DestroyRunningMatch", WebAPIHandlerWrapper(
//		func(tNode *transact.TransNode, params webapi.RequestBody) (int, webapi.ResponseBody) {
//			resp := webapi.NewResponseBody()
//			params_data, _ := params.GetRequestBody("Param")
//			platform, _ := params_data.GetStr("Platform")
//			matchId, _ := params_data.GetInt("MatchCopyId")
//
//			m := MatchMgrSington.GetCopyMatch(int32(matchId))
//			if m != nil {
//				if platform != "" && m.dbMatch.Platform != platform {
//					resp[webapi.RESPONSE_STATE] = webapi.STATE_ERR
//					resp[webapi.RESPONSE_ERRMSG] = "limits of authority!"
//					return common.ResponseTag_OpFailed, resp
//				}
//
//				m.ForceDestroy()
//			}
//
//			resp[webapi.RESPONSE_STATE] = webapi.STATE_OK
//			return common.ResponseTag_Ok, resp
//		}))
//
//	//-------------------------------------------------------------------------------------------------------
//	//更新比赛券配置
//	WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/api/act/UpdateTicketConfig", WebAPIHandlerWrapper(
//		func(tNode *transact.TransNode, params webapi.RequestBody) (int, webapi.ResponseBody) {
//			resp := webapi.NewResponseBody()
//			params_data, _ := params.GetRequestBody("Param")
//			if data, ok := params_data.GetRequestBody("Data"); ok {
//				buf, err := json.Marshal(data)
//				if err != nil {
//					resp[webapi.RESPONSE_STATE] = webapi.STATE_ERR
//					resp[webapi.RESPONSE_ERRMSG] = "params error!"
//					return common.ResponseTag_ParamError, resp
//				}
//
//				var cfg ActTicketConfig
//				err = json.Unmarshal(buf, &cfg)
//				if err != nil {
//					resp[webapi.RESPONSE_STATE] = webapi.STATE_ERR
//					resp[webapi.RESPONSE_ERRMSG] = "params error!"
//					return common.ResponseTag_ParamError, resp
//				}
//
//				ActTicketMgrSington.UpdateConfig(&cfg, false)
//				resp[webapi.RESPONSE_STATE] = webapi.STATE_OK
//				return common.ResponseTag_Ok, resp
//			}
//
//			resp[webapi.RESPONSE_STATE] = webapi.STATE_ERR
//			resp[webapi.RESPONSE_ERRMSG] = "need Data!"
//			return common.ResponseTag_ParamError, resp
//		}))
//
//	//-------------------------------------------------------------------------------------------------------
//	//钱包操作接口
//	WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/api/Match/AddTicket", WebAPIHandlerWrapper(
//		func(tNode *transact.TransNode, params webapi.RequestBody) (int, webapi.ResponseBody) {
//			params_data, _ := params.GetRequestBody("Param")
//			resp := webapi.NewResponseBody()
//
//			member_snid, _ := params_data.GetInt("ID")
//			count, _ := params_data.GetInt64("Ticket")
//			oper, _ := params_data.GetStr("Oper")
//			desc, _ := params_data.GetStr("Desc")
//			billNo, _ := params_data.GetInt("BillNo")
//			platform, _ := params_data.GetStr("Platform")
//
//			if CacheDataMgr.CacheBillCheck(billNo) {
//				resp[webapi.RESPONSE_STATE] = webapi.STATE_ERR
//				resp[webapi.RESPONSE_ERRMSG] = "Bill number repeated!"
//				resp[webapi.RESPONSE_DATA] = make(map[string]interface{})
//				return common.ResponseTag_ParamError, resp
//			}
//			CacheDataMgr.CacheBillNumber(billNo) //防止手抖点两下
//
//			var err error
//			var pd *model.PlayerData
//			oldTicket := int64(0)
//			var timeStamp = time.Now().UnixNano()
//			type PlayerTicketData struct {
//				ID     int32
//				Ticket int64
//			}
//			player := PlayerMgrSington.GetPlayerBySnId(int32(member_snid))
//			if player != nil { //在线玩家处理
//				pd = player.PlayerData
//				if len(platform) > 0 && player.Platform != platform {
//					CacheDataMgr.ClearCacheBill(billNo)
//					resp[webapi.RESPONSE_STATE] = webapi.STATE_ERR
//					resp[webapi.RESPONSE_ERRMSG] = "player platform forbit!"
//					resp[webapi.RESPONSE_DATA] = make(map[string]interface{})
//					return common.ResponseTag_ParamError, resp
//				}
//
//				if count < 0 {
//					if player.Ticket+count < 0 {
//						CacheDataMgr.ClearCacheBill(billNo)
//						resp[webapi.RESPONSE_STATE] = webapi.STATE_ERR
//						resp[webapi.RESPONSE_ERRMSG] = "ticket not enough!"
//						resp[webapi.RESPONSE_DATA] = make(map[string]interface{})
//						return common.ResponseTag_ParamError, resp
//					}
//				}
//
//				gainWay := int32(common.GainWay_API_AddTicket)
//
//				oldTicket = player.Ticket
//				ticketLog := model.NewPayCoinLog(int64(billNo), int32(member_snid), count, gainWay,
//					desc, model.PayCoinLogType_Ticket, 0)
//				timeStamp = ticketLog.TimeStamp
//				//增加日志记录
//				log := model.NewTicketLogEx(int32(member_snid), count, oldTicket+count, player.Ver, gainWay, oper, desc, player.Platform, player.Channel, player.BeUnderAgentCode, player.PackageID, player.InviterId)
//				task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
//					err = model.InsertPayCoinLogs(ticketLog)
//					if err != nil {
//						logger.Logger.Errorf("model.InsertPayCoinLogs err:%v log:%v", err, ticketLog)
//						return err
//					}
//					err = model.InsertTicketLogs(log)
//					if err != nil {
//						//回滚到对账日志
//						model.RemovePayCoinLog(ticketLog.LogId)
//						logger.Logger.Errorf("model.InsertTicketLogs err:%v log:%v", err, ticketLog)
//						return err
//					}
//					return err
//				}), task.CompleteNotifyWrapper(func(data interface{}, t *task.Task) {
//					CacheDataMgr.ClearCacheBill(billNo)
//					if data != nil {
//						resp[webapi.RESPONSE_STATE] = webapi.STATE_ERR
//						resp[webapi.RESPONSE_ERRMSG] = data.(error).Error()
//						resp[webapi.RESPONSE_DATA] = make(map[string]interface{})
//					} else {
//						player.Ticket += count
//						player.TicketTotal += count
//						player.SetTicketPayTs(timeStamp)
//						player.dirty = true
//						if player.scene == nil { //如果在大厅,那么同步下数据
//							player.SendDiffData()
//						}
//						pcd := &PlayerTicketData{
//							ID:     int32(member_snid),
//							Ticket: player.Ticket,
//						}
//						resp[webapi.RESPONSE_STATE] = webapi.STATE_OK
//						resp[webapi.RESPONSE_ERRMSG] = ""
//						resp[webapi.RESPONSE_DATA] = pcd
//					}
//					dataResp := &common.M2GWebApiResponse{}
//					dataResp.Body, err = resp.Marshal()
//					tNode.TransRep.RetFiels = dataResp
//					tNode.Resume()
//					if err != nil {
//						logger.Logger.Error("AddTicket task marshal data error:", err)
//					}
//				}), "AddTicket").Start()
//				return common.ResponseTag_TransactYield, resp
//			} else {
//				op := OfflinePlayerMgrSington.GetPlayer(int32(member_snid))
//				if op != nil {
//					if count < 0 {
//						if op.Ticket+count < 0 {
//							CacheDataMgr.ClearCacheBill(billNo)
//							resp[webapi.RESPONSE_STATE] = webapi.STATE_ERR
//							resp[webapi.RESPONSE_ERRMSG] = "ticket not enough!"
//							resp[webapi.RESPONSE_DATA] = make(map[string]interface{})
//							return common.ResponseTag_ParamError, resp
//						}
//					}
//				}
//				task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
//					pd = model.GetPlayerDataBySnId(int32(member_snid), true)
//					if pd == nil {
//						return errors.New("Player not find.")
//					}
//					if len(platform) > 0 && pd.Platform != platform {
//						return errors.New("player platform forbit.")
//					}
//					oldTicket = pd.Ticket
//					if count < 0 {
//						if oldTicket+count < 0 {
//							return errors.New("ticket not enough!")
//						}
//					}
//
//					gainWay := int32(common.GainWay_API_AddTicket)
//					ticketLog := model.NewPayCoinLog(int64(billNo), int32(member_snid), count, gainWay,
//						desc, model.PayCoinLogType_Ticket, 0)
//					timeStamp = ticketLog.TimeStamp
//					err = model.InsertPayCoinLogs(ticketLog)
//					if err != nil {
//						logger.Logger.Errorf("model.InsertPayCoinLogs err:%v log:%v", err, ticketLog)
//						return err
//					}
//					//增加帐变记录
//					//增加日志记录
//					log := model.NewTicketLogEx(int32(member_snid), count, oldTicket+count, pd.Ver, gainWay, oper, desc, pd.Platform, pd.Channel, pd.BeUnderAgentCode, pd.PackageID, pd.InviterId)
//					err = model.InsertTicketLogs(log)
//					if err != nil {
//						//回滚到对账日志
//						model.RemovePayCoinLog(ticketLog.LogId)
//						logger.Logger.Errorf("model.InsertTicketLogs err:%v log:%v", err, log)
//						return err
//					}
//					return err
//				}), task.CompleteNotifyWrapper(func(data interface{}, t *task.Task) {
//					CacheDataMgr.ClearCacheBill(billNo)
//					if data != nil {
//						resp[webapi.RESPONSE_STATE] = webapi.STATE_ERR
//						resp[webapi.RESPONSE_ERRMSG] = data.(error).Error()
//						resp[webapi.RESPONSE_DATA] = make(map[string]interface{})
//					} else {
//						pcd := &PlayerTicketData{
//							ID:     int32(member_snid),
//							Ticket: pd.Ticket + count,
//						}
//						resp[webapi.RESPONSE_STATE] = webapi.STATE_OK
//						resp[webapi.RESPONSE_ERRMSG] = ""
//						resp[webapi.RESPONSE_DATA] = pcd
//						//更新线下缓存数据
//						op := OfflinePlayerMgrSington.GetPlayer(int32(member_snid))
//						if op != nil {
//							op.Ticket += count
//							op.TicketTotal += count
//							op.SetTicketPayTs(timeStamp)
//						}
//					}
//
//					dataResp := &common.M2GWebApiResponse{}
//					dataResp.Body, err = resp.Marshal()
//					tNode.TransRep.RetFiels = dataResp
//					tNode.Resume()
//					if err != nil {
//						logger.Logger.Error("AddTicket task marshal data error:", err)
//					}
//				}), "AddTicket").Start()
//				return common.ResponseTag_TransactYield, resp
//			}
//		}))
//
//	//-------------------------------------------------------------------------------------------------------
//	//更新比赛券配置
//	WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/api/act/ManualDistTicket", WebAPIHandlerWrapper(
//		func(tNode *transact.TransNode, params webapi.RequestBody) (int, webapi.ResponseBody) {
//			resp := webapi.NewResponseBody()
//			params_data, _ := params.GetRequestBody("Param")
//			orderId, _ := params_data.GetStr("OrderId")
//			platform, _ := params_data.GetStr("Platform")
//			task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
//				order, err := model.DistAwaitIssuedTicketOrder(orderId, platform)
//				if err != nil {
//					return err
//				}
//				title := i18n.Tr("cn", "ActTicketTitle")
//				content := i18n.Tr("cn", "ActTicketContent", order.SnId, order.MatchName, order.Rank, order.Count)
//				msg := model.NewMessage("", int32(0), order.InviterId, model.MSGTYPE_MATCH_TICKETREWARD, title, content, 0,
//					model.MSGSTATE_UNREAD, time.Now().Unix(), model.MSGATTACHSTATE_DEFAULT, "", nil, order.Platform)
//				msg.Ticket = order.Count
//				err = model.InsertMessage(msg)
//				if err != nil {
//					return err
//				}
//				return msg
//			}), task.CompleteNotifyWrapper(func(data interface{}, t *task.Task) {
//				if data != nil {
//					if msg, ok := data.(*model.Message); ok && msg != nil {
//						p := PlayerMgrSington.GetPlayerBySnId(msg.SnId)
//						if p != nil {
//							p.AddMessage(msg)
//						}
//					}
//					resp[webapi.RESPONSE_STATE] = webapi.STATE_OK
//				} else {
//					resp[webapi.RESPONSE_STATE] = webapi.STATE_ERR
//				}
//				dataResp := &common.M2GWebApiResponse{}
//				dataResp.Body, _ = resp.Marshal()
//				tNode.TransRep.RetFiels = dataResp
//				tNode.Resume()
//			}), "DistAwaitIssuedTicketOrder").Start()
//			return common.ResponseTag_TransactYield, resp
//		}))
//}
