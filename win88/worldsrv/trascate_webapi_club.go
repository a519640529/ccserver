package main

//
//import (
//	"games.yol.com/win88/proto"
//	"encoding/json"
//	"errors"
//	"fmt"
//	"games.yol.com/win88/common"
//	"games.yol.com/win88/model"
//	"games.yol.com/win88/protocol"
//	"games.yol.com/win88/srvdata"
//	"games.yol.com/win88/webapi"
//	"github.com/idealeak/goserver/core/basic"
//	"github.com/idealeak/goserver/core/logger"
//	"github.com/idealeak/goserver/core/task"
//	"github.com/idealeak/goserver/core/transact"
//	"sort"
//	"strconv"
//	"strings"
//	"time"
//)
//
////俱乐部相关提供给后台的API
//func init() {
//	//-------------------------------------------------------------------------------------------------------
//	//获取俱乐部创建者的基本信息(正式)
//	WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/api/Club/ClubCreatorBaseInfo", WebAPIHandlerWrapper(
//		func(tNode *transact.TransNode, params webapi.RequestBody) (int, webapi.ResponseBody) {
//			params_data, _ := params.GetRequestBody("Param")
//			resp := webapi.NewResponseBody()
//			plt_id, _ := params_data.GetStr("PltID") //必填
//			snid_int, _ := params_data.GetInt("Snid")
//			snid := int32(snid_int)
//			//判断参数是否合法
//			if snid <= 0 {
//				resp[webapi.RESPONSE_STATE] = webapi.STATE_ERR
//				resp[webapi.RESPONSE_ERRMSG] = "Snid value error!"
//				resp[webapi.RESPONSE_DATA] = make(map[string]interface{})
//				return common.ResponseTag_ParamError, resp
//			}
//			//检测平台参数
//			if len(plt_id) != 0 {
//				plt := PlatformMgrSington.GetPlatform(plt_id)
//				if plt == nil {
//					resp[webapi.RESPONSE_STATE] = webapi.STATE_ERR
//					resp[webapi.RESPONSE_ERRMSG] = "Can't find PltID info error!"
//					resp[webapi.RESPONSE_DATA] = make(map[string]interface{})
//					return common.ResponseTag_ParamError, resp
//				}
//			}
//
//			creatorClubMaxLevel := int32(0)
//			platformId := ""
//			for _, v := range clubManager.clubList {
//				if v.Owner == snid {
//					if v.Level > creatorClubMaxLevel {
//						platformId = v.Platform
//						creatorClubMaxLevel = v.Level
//					}
//				}
//			}
//			if creatorClubMaxLevel == 0 {
//				resp[webapi.RESPONSE_STATE] = webapi.STATE_OK
//				resp[webapi.RESPONSE_ERRMSG] = "Can't find player."
//				resp[webapi.RESPONSE_DATA] = ""
//			} else {
//				resp_data := make(map[string]interface{})
//				pf := PlatformMgrSington.GetPlatform(platformId)
//				if pf != nil {
//					p := PlayerMgrSington.GetPlayerBySnId(snid)
//					if p != nil {
//						resp_data["ClubCoin"] = p.ClubCoin
//						resp_data["ClubMaxLevel"] = creatorClubMaxLevel
//						resp_data["GiveRate"] = pf.ClubConfig.GiveCoinRate[creatorClubMaxLevel-1]
//						resp[webapi.RESPONSE_STATE] = webapi.STATE_OK
//						resp[webapi.RESPONSE_DATA] = resp_data
//						return common.ResponseTag_Ok, resp
//					} else {
//						async := true
//						OfflinePlayerMgrSington.GetOfflinePlayer(snid, func(op *OfflinePlayer, bAsync bool) {
//							async = bAsync
//							resp_data["ClubCoin"] = op.ClubCoin
//							resp_data["ClubMaxLevel"] = creatorClubMaxLevel
//							resp_data["GiveRate"] = pf.ClubConfig.GiveCoinRate[creatorClubMaxLevel-1]
//							resp[webapi.RESPONSE_STATE] = webapi.STATE_OK
//							resp[webapi.RESPONSE_DATA] = resp_data
//							dataResp := &common.M2GWebApiResponse{}
//							var err error
//							dataResp.Body, err = resp.Marshal()
//							if err == nil {
//								if bAsync {
//									tNode.TransRep.RetFiels = dataResp
//									tNode.Resume()
//								}
//							}
//						}, false)
//						if !async {
//							return common.ResponseTag_Ok, resp
//						} else {
//							return common.ResponseTag_TransactYield, resp
//						}
//					}
//				}
//			}
//			return common.ResponseTag_Ok, resp
//		}))
//
//	//-------------------------------------------------------------------------------------------------------
//	//获取俱乐部大纲信息(正式)
//	WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/api/Club/ClubOutline", WebAPIHandlerWrapper(
//		func(tNode *transact.TransNode, params webapi.RequestBody) (int, webapi.ResponseBody) {
//			params_data, _ := params.GetRequestBody("Param")
//			resp := webapi.NewResponseBody()
//			plt_id, _ := params_data.GetStr("PltID")
//			club_id, _ := params_data.GetInt("ClubID")
//
//			//检测平台参数
//			if len(plt_id) != 0 {
//				plt := PlatformMgrSington.GetPlatform(plt_id)
//				if plt == nil {
//					resp[webapi.RESPONSE_STATE] = webapi.STATE_ERR
//					resp[webapi.RESPONSE_ERRMSG] = "Can't find PltID info error!"
//					resp[webapi.RESPONSE_DATA] = make(map[string]interface{})
//					return common.ResponseTag_ParamError, resp
//				}
//			}
//
//			//判断俱乐部号码是否存在
//			club := clubManager.clubList[int32(club_id)]
//			data := model.ClubOutline{}
//			if club == nil {
//				//查一个平台下的数据
//				//通过俱乐部的平台找到这个平台下的所有场景
//				club_scenes := ClubSceneMgrSington.GetPlatformClub(plt_id)
//				if club_scenes == nil {
//					resp[webapi.RESPONSE_STATE] = webapi.STATE_ERR
//					resp[webapi.RESPONSE_ERRMSG] = "PltID have no scenes info error!"
//					resp[webapi.RESPONSE_DATA] = make(map[string]interface{})
//					return common.ResponseTag_ParamError, resp
//				}
//
//				for _, v := range club_scenes {
//					data.PlayerPlayingNum += int32(len(v.players))
//					data.RoomPlayingNum += int32(len(v.sceneList))
//				}
//			} else {
//				//看是否属于该平台
//				if club.Platform != plt_id {
//					resp[webapi.RESPONSE_STATE] = webapi.STATE_ERR
//					resp[webapi.RESPONSE_ERRMSG] = "PltID no have this ClubID error!"
//					resp[webapi.RESPONSE_DATA] = make(map[string]interface{})
//					return common.ResponseTag_ParamError, resp
//				}
//				//通过俱乐部的平台找到这个平台下的所有场景
//				club_scenes := ClubSceneMgrSington.GetPlatformClub(club.Platform)
//				if club_scenes == nil {
//					resp[webapi.RESPONSE_STATE] = webapi.STATE_ERR
//					resp[webapi.RESPONSE_ERRMSG] = "ClubID have no plt info error!"
//					resp[webapi.RESPONSE_DATA] = make(map[string]interface{})
//					return common.ResponseTag_ParamError, resp
//				}
//
//				//根据俱乐部ID找到该俱乐部下的ClubScenePool信息
//				csp := club_scenes[club.Id]
//				if csp == nil {
//					resp[webapi.RESPONSE_STATE] = webapi.STATE_ERR
//					resp[webapi.RESPONSE_ERRMSG] = "ClubID have no ClubScenePool info error!"
//					resp[webapi.RESPONSE_DATA] = make(map[string]interface{})
//					return common.ResponseTag_ParamError, resp
//				}
//
//				if len(plt_id) == 0 && club_id == 0 {
//					data.PlayerPlayingNum = int32(len(csp.players))
//					data.RoomPlayingNum = int32(len(csp.sceneList))
//				}
//			}
//			resp[webapi.RESPONSE_ERRMSG] = ""
//			resp[webapi.RESPONSE_STATE] = webapi.STATE_OK
//			resp[webapi.RESPONSE_DATA] = data
//			dataResp := &common.M2GWebApiResponse{}
//			dataResp.Body, _ = resp.Marshal()
//			tNode.TransRep.RetFiels = dataResp
//			tNode.Resume()
//			return common.ResponseTag_Ok, resp
//		}))
//
//	//-------------------------------------------------------------------------------------------------------
//	//获取多个俱乐部信息(正式)
//	WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/api/Club/ClubSet", WebAPIHandlerWrapper(
//		func(tNode *transact.TransNode, params webapi.RequestBody) (int, webapi.ResponseBody) {
//			params_data, _ := params.GetRequestBody("Param")
//			resp := webapi.NewResponseBody()
//			plt_id, _ := params_data.GetStr("PltID")
//			club_id, _ := params_data.GetInt("ClubID")
//			page_size, _ := params_data.GetInt("PageSize")
//			page_num, _ := params_data.GetInt("PageNum")
//
//			//判断参数
//			if page_size == 0 || page_num == 0 {
//				resp[webapi.RESPONSE_STATE] = webapi.STATE_ERR
//				resp[webapi.RESPONSE_ERRMSG] = "PageSize or PageNum arg must hava a value error!"
//				resp[webapi.RESPONSE_DATA] = make(map[string]interface{})
//				return common.ResponseTag_ParamError, resp
//			}
//			//检测平台参数
//			plt := PlatformMgrSington.GetPlatform(plt_id)
//			if plt == nil {
//				resp[webapi.RESPONSE_STATE] = webapi.STATE_ERR
//				resp[webapi.RESPONSE_ERRMSG] = "Can't find PltID info error!"
//				resp[webapi.RESPONSE_DATA] = make(map[string]interface{})
//				return common.ResponseTag_ParamError, resp
//			}
//			clubArr := make([]model.ClubItem, 0)
//			var club *ClubInst
//
//			//如果有俱乐部ID参数的就取一个
//			if club_id != 0 {
//				club = clubManager.clubList[int32(club_id)]
//				if club != nil && club.Platform == plt_id {
//					club_item := model.ClubItem{
//						ClubId:        club.Id,
//						PltID:         club.Platform,
//						ClubName:      club.Name,
//						Level:         club.Level,
//						CreatorSnid:   club.Owner,
//						MemberNum:     club.GetMemberSum(),
//						MaxMemberNum:  plt.ClubConfig.ClubInitPlayerNum,
//						OtherPumpRate: club.Setting.Taxes,
//						CreateTime:    club.CreateTs,
//					}
//					clubArr = append(clubArr, club_item)
//				}
//			} else {
//				//如果没有俱乐部ID参数，就取该平台下的全部俱乐部信息
//				for _, v := range clubManager.clubList {
//					if v.Platform == plt_id {
//						club_item := model.ClubItem{
//							ClubId:        v.Id,
//							PltID:         v.Platform,
//							ClubName:      v.Name,
//							Level:         v.Level,
//							CreatorSnid:   v.Owner,
//							MemberNum:     v.GetMemberSum(),
//							MaxMemberNum:  plt.ClubConfig.ClubInitPlayerNum,
//							OtherPumpRate: v.Setting.Taxes,
//							CreateTime:    v.CreateTs,
//						}
//						clubArr = append(clubArr, club_item)
//					}
//				}
//			}
//
//			//开始做分页
//			start := (page_num - 1) * page_size
//			end := page_num * page_size
//			count := len(clubArr)
//			if count < start {
//				return common.ResponseTag_Ok, resp
//			}
//			if end > count {
//				end = count
//			}
//			sort.Slice(clubArr, func(i, j int) bool {
//				return clubArr[i].ClubId < clubArr[j].ClubId
//			})
//			r := clubArr[start:end]
//			resp["PageCount"] = count
//			resp["PageNo"] = page_num
//			resp["PageSize"] = page_size
//			resp[webapi.RESPONSE_ERRMSG] = ""
//			resp[webapi.RESPONSE_STATE] = webapi.STATE_OK
//			resp[webapi.RESPONSE_DATA] = r
//			dataResp := &common.M2GWebApiResponse{}
//			dataResp.Body, _ = resp.Marshal()
//			tNode.TransRep.RetFiels = dataResp
//			tNode.Resume()
//			return common.ResponseTag_Ok, resp
//		}))
//
//	//-------------------------------------------------------------------------------------------------------
//	//刷新待审核的俱乐部信息
//	//接受后台俱乐部审核信息的刷新
//	//用途：游服给后台发送审核失败后，后台查不到该条俱乐部的待审核记录，这时后台可以通过该接口刷新一下。
//	//该接口会触发游服将未审核的俱乐部再次重新发送到后台。
//	WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/api/Club/RefreshWaitCheckClubInfo", WebAPIHandlerWrapper(
//		func(tNode *transact.TransNode, params webapi.RequestBody) (int, webapi.ResponseBody) {
//			resp := webapi.NewResponseBody()
//
//			waitCreateCheckClubArr := make([]*model.Club, 0)
//			waitNoticeCheckClubArr := make([]*model.Club, 0)
//			for _, v := range clubManager.clubList {
//				if !v.CreateCheckPosted {
//					waitCreateCheckClubArr = append(waitCreateCheckClubArr, v.Club)
//				}
//				if !v.NoticeCheckPosted {
//					waitNoticeCheckClubArr = append(waitCreateCheckClubArr, v.Club)
//				}
//			}
//			needPostCount := len(waitCreateCheckClubArr) + len(waitNoticeCheckClubArr)
//			retdata := make(map[string]interface{})
//			retdata["PushCount"] = needPostCount
//			if needPostCount != 0 {
//				task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
//					for _, v := range waitCreateCheckClubArr {
//						if err := webapi.API_ClubCreateWaitCheck(common.GetAppId(), v.Id, v.Owner, v.Platform, v.Name, v.Billboard); err == nil {
//							v.CreateCheckPosted = true
//						}
//					}
//					for _, v := range waitNoticeCheckClubArr {
//						if err := webapi.API_ClubNoticeWaitCheck(common.GetAppId(), v.Id, v.Owner, v.Platform, v.Name, v.Billboard); err == nil {
//							v.NoticeCheckPosted = true
//						}
//					}
//					return nil
//				}), task.CompleteNotifyWrapper(func(data interface{}, t *task.Task) {
//					resp[webapi.RESPONSE_STATE] = webapi.STATE_OK
//					resp[webapi.RESPONSE_DATA] = retdata
//					dataResp := &common.M2GWebApiResponse{}
//					dataResp.Body, _ = resp.Marshal()
//					tNode.TransRep.RetFiels = dataResp
//					tNode.Resume()
//				}), "RefreshWaitCheckClubInfo").Start()
//				return common.ResponseTag_TransactYield, resp
//			} else {
//				resp[webapi.RESPONSE_STATE] = webapi.STATE_OK
//				resp[webapi.RESPONSE_DATA] = retdata
//				dataResp := &common.M2GWebApiResponse{}
//				dataResp.Body, _ = resp.Marshal()
//				tNode.TransRep.RetFiels = dataResp
//				tNode.Resume()
//				return common.ResponseTag_Ok, resp
//			}
//		}))
//
//	//-------------------------------------------------------------------------------------------------------
//	//获取俱乐部下面的基础信息(正式)
//	WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/api/Club/ClubBaseInfo", WebAPIHandlerWrapper(
//		func(tNode *transact.TransNode, params webapi.RequestBody) (int, webapi.ResponseBody) {
//			params_data, _ := params.GetRequestBody("Param")
//			resp := webapi.NewResponseBody()
//			plt_id, _ := params_data.GetStr("PltID") //必填
//			club_id, _ := params_data.GetInt("ClubID")
//			//判断参数是否合法
//			if club_id <= 0 {
//				resp[webapi.RESPONSE_STATE] = webapi.STATE_ERR
//				resp[webapi.RESPONSE_ERRMSG] = "ClubID value error!"
//				resp[webapi.RESPONSE_DATA] = make(map[string]interface{})
//				return common.ResponseTag_ParamError, resp
//			}
//
//			//判断俱乐部号码是否存在
//			club := clubManager.clubList[int32(club_id)]
//			if club == nil {
//				resp[webapi.RESPONSE_STATE] = webapi.STATE_ERR
//				resp[webapi.RESPONSE_ERRMSG] = "ClubID is not exist error!"
//				resp[webapi.RESPONSE_DATA] = make(map[string]interface{})
//				return common.ResponseTag_ParamError, resp
//			}
//
//			//检查一下俱乐部ID和平台是否对应
//			if club.Platform != plt_id {
//				resp[webapi.RESPONSE_STATE] = webapi.STATE_ERR
//				resp[webapi.RESPONSE_ERRMSG] = "PltId ClubID is not match error!"
//				resp[webapi.RESPONSE_DATA] = make(map[string]interface{})
//				return common.ResponseTag_ParamError, resp
//			}
//
//			//通过俱乐部的平台找到这个平台下的所有场景
//			club_scenes := ClubSceneMgrSington.GetPlatformClub(club.Platform)
//			data := &model.ClubBaseInfo{
//				ClubId:          club.Id,
//				ClubName:        club.Name,
//				Level:           club.Level,
//				CreatorSnid:     club.Owner,
//				CreateTs:        club.CreateTs,
//				MemberNum:       club.GetMemberSum(),
//				PumpRate:        club.Setting.Taxes,
//				PlayingRoomsNum: 0,
//				JoinPlayerNum:   0,
//				Billboard:       club.Billboard,
//			}
//
//			//根据俱乐部ID找到该俱乐部下的ClubScenePool信息
//			if club_scenes == nil {
//				data.PlayingRoomsNum = 0
//				data.JoinPlayerNum = 0
//			} else {
//				data.PlayingRoomsNum = 0
//				data.JoinPlayerNum = 0
//				for _, v := range club_scenes {
//					if v.clubId == int32(club_id) {
//						data.PlayingRoomsNum = int32(len(v.players))
//						data.JoinPlayerNum = int32(len(v.sceneList))
//						break
//					}
//				}
//			}
//			resp[webapi.RESPONSE_STATE] = webapi.STATE_OK
//			resp[webapi.RESPONSE_DATA] = data
//			dataResp := &common.M2GWebApiResponse{}
//			dataResp.Body, _ = resp.Marshal()
//			tNode.TransRep.RetFiels = dataResp
//			tNode.Resume()
//			return common.ResponseTag_Ok, resp
//		}))
//
//	//-------------------------------------------------------------------------------------------------------
//	//获取俱乐部下面的成员信息(正式)
//	WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/api/Club/ClubMemberSet", WebAPIHandlerWrapper(
//		func(tNode *transact.TransNode, params webapi.RequestBody) (int, webapi.ResponseBody) {
//			params_data, _ := params.GetRequestBody("Param")
//			resp := webapi.NewResponseBody()
//			plt_id, _ := params_data.GetStr("PltID") //必填
//			club_id, _ := params_data.GetInt("ClubID")
//			page_size, _ := params_data.GetInt("PageSize")
//			page_num, _ := params_data.GetInt("PageNum")
//
//			//判断参数
//			if page_size == 0 || page_num == 0 {
//				resp[webapi.RESPONSE_STATE] = webapi.STATE_ERR
//				resp[webapi.RESPONSE_ERRMSG] = "PageSize or PageNum arg must hava a value error!"
//				resp[webapi.RESPONSE_DATA] = make(map[string]interface{})
//				return common.ResponseTag_ParamError, resp
//			}
//
//			//判断参数是否合法
//			if club_id <= 0 {
//				resp[webapi.RESPONSE_STATE] = webapi.STATE_ERR
//				resp[webapi.RESPONSE_ERRMSG] = "ClubID value error!"
//				resp[webapi.RESPONSE_DATA] = make(map[string]interface{})
//				return common.ResponseTag_ParamError, resp
//			}
//
//			//判断俱乐部号码是否存在
//			club := clubManager.clubList[int32(club_id)]
//			if club == nil {
//				resp[webapi.RESPONSE_STATE] = webapi.STATE_ERR
//				resp[webapi.RESPONSE_ERRMSG] = "ClubID is not exist error!"
//				resp[webapi.RESPONSE_DATA] = make(map[string]interface{})
//				return common.ResponseTag_ParamError, resp
//			}
//
//			//检查一下俱乐部ID和平台是否对应
//			if club.Platform != plt_id {
//				resp[webapi.RESPONSE_STATE] = webapi.STATE_ERR
//				resp[webapi.RESPONSE_ERRMSG] = "PltId and ClubID is not match error!"
//				resp[webapi.RESPONSE_DATA] = make(map[string]interface{})
//				return common.ResponseTag_ParamError, resp
//			}
//
//			clubMemberArr := make([]model.ClubMemberItem, 0)
//			for _, v := range club.memberList {
//				item := model.ClubMemberItem{
//					Snid:           v.SnId,
//					Position:       int32(v.Position),
//					IsBlack:        v.IsBlack,
//					JoinClubTimeTs: v.LastTime,
//				}
//				clubMemberArr = append(clubMemberArr, item)
//			}
//			//开始做分页
//			start := (page_num - 1) * page_size
//			end := page_num * page_size
//			count := len(clubMemberArr)
//			if count < start {
//				return common.ResponseTag_Ok, resp
//			}
//			if end > count {
//				end = count
//			}
//			//根据加入俱乐部的时间倒序
//			sort.Slice(clubMemberArr, func(i, j int) bool {
//				return clubMemberArr[i].JoinClubTimeTs > clubMemberArr[j].JoinClubTimeTs
//			})
//			r := clubMemberArr[start:end]
//			resp["PageCount"] = count
//			resp["PageNo"] = page_num
//			resp["PageSize"] = page_size
//			resp[webapi.RESPONSE_ERRMSG] = ""
//			resp[webapi.RESPONSE_STATE] = webapi.STATE_OK
//			resp[webapi.RESPONSE_DATA] = r
//			dataResp := &common.M2GWebApiResponse{}
//			dataResp.Body, _ = resp.Marshal()
//			tNode.TransRep.RetFiels = dataResp
//			tNode.Resume()
//			return common.ResponseTag_Ok, resp
//		}))
//
//	//-------------------------------------------------------------------------------------------------------
//	//获取玩家“名下的俱乐部信息”和“玩家加入的俱乐部信息”包含已经解散的俱乐部(正式)
//	WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/api/Club/PlayerClubRelation", WebAPIHandlerWrapper(
//		func(tNode *transact.TransNode, params webapi.RequestBody) (int, webapi.ResponseBody) {
//			params_data, _ := params.GetRequestBody("Param")
//			resp := webapi.NewResponseBody()
//			plt_id, _ := params_data.GetStr("PltID") //必填
//			snid, _ := params_data.GetInt("Snid")
//
//			//判断参数是否合法
//			if snid <= 0 {
//				resp[webapi.RESPONSE_STATE] = webapi.STATE_ERR
//				resp[webapi.RESPONSE_ERRMSG] = "Snid value error!"
//				resp[webapi.RESPONSE_DATA] = make(map[string]interface{})
//				return common.ResponseTag_ParamError, resp
//			}
//
//			//检测平台参数
//			plt := PlatformMgrSington.GetPlatform(plt_id)
//			if plt == nil {
//				resp[webapi.RESPONSE_STATE] = webapi.STATE_ERR
//				resp[webapi.RESPONSE_ERRMSG] = "Can't find PltID info error!"
//				resp[webapi.RESPONSE_DATA] = make(map[string]interface{})
//				return common.ResponseTag_ParamError, resp
//			}
//
//			type PlayerClub struct {
//				Snid   int32
//				ClubId int32
//				State  int32 //state=-1为解散，state=1为正常
//			}
//			ownerClubs := make([]PlayerClub, 0)
//			joinClubs := make([]PlayerClub, 0)
//
//			//获取内存中的数据
//			for _, v := range clubManager.clubList {
//				if v.Platform == plt_id {
//					if v.Owner == int32(snid) {
//						ownerClubs = append(ownerClubs, PlayerClub{
//							Snid:   v.Owner,
//							ClubId: v.Id,
//							State:  1,
//						})
//					}
//					for _, m := range v.memberList {
//						if m.SnId == int32(snid) {
//							joinClubs = append(joinClubs, PlayerClub{
//								Snid:   m.SnId,
//								ClubId: m.ClubId,
//								State:  1,
//							})
//							break
//						}
//					}
//				}
//			}
//
//			//获取日志中的数据
//			var clublog []*model.ClubLog
//			task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
//				clublog = model.GetAllClubLogDataBySnid(int32(snid))
//				return nil
//			}), task.CompleteNotifyWrapper(func(data interface{}, t *task.Task) {
//				for _, v := range clublog {
//					if v.PltId == plt_id {
//						if v.LogType == model.ClubLogType_Dismiss {
//							ownerClubs = append(ownerClubs, PlayerClub{
//								Snid:   v.SnId,
//								ClubId: v.ClubId,
//								State:  -1,
//							})
//						}
//						if v.LogType == model.ClubLogType_PassiveDismiss {
//							ownerClubs = append(ownerClubs, PlayerClub{
//								Snid:   v.SnId,
//								ClubId: v.ClubId,
//								State:  -1,
//							})
//						}
//					}
//				}
//				r := make(map[string]interface{})
//				r["ownerClubs"] = ownerClubs
//				r["joinClubs"] = joinClubs
//				resp[webapi.RESPONSE_STATE] = webapi.STATE_OK
//				resp[webapi.RESPONSE_DATA] = r
//				dataResp := &common.M2GWebApiResponse{}
//				dataResp.Body, _ = resp.Marshal()
//				tNode.TransRep.RetFiels = dataResp
//				tNode.Resume()
//			}), "PlayerClubRelation").Start()
//			return common.ResponseTag_TransactYield, resp
//		}))
//
//	//-------------------------------------------------------------------------------------------------------
//	//获取俱乐部下面的游戏包间信息(正式)
//	WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/api/Club/ClubAreaSet", WebAPIHandlerWrapper(
//		func(tNode *transact.TransNode, params webapi.RequestBody) (int, webapi.ResponseBody) {
//			params_data, _ := params.GetRequestBody("Param")
//			resp := webapi.NewResponseBody()
//			plt_id, _ := params_data.GetStr("PltID") //必填
//			club_id, _ := params_data.GetInt("ClubID")
//			page_size, _ := params_data.GetInt("PageSize")
//			page_num, _ := params_data.GetInt("PageNum")
//
//			//判断参数是否合法
//			if club_id <= 0 {
//				resp[webapi.RESPONSE_STATE] = webapi.STATE_ERR
//				resp[webapi.RESPONSE_ERRMSG] = "ClubID value error!"
//				resp[webapi.RESPONSE_DATA] = make(map[string]interface{})
//				return common.ResponseTag_ParamError, resp
//			}
//
//			//判断俱乐部号码是否存在
//			club := clubManager.clubList[int32(club_id)]
//			if club == nil {
//				resp[webapi.RESPONSE_STATE] = webapi.STATE_ERR
//				resp[webapi.RESPONSE_ERRMSG] = "ClubID is not exist error!"
//				resp[webapi.RESPONSE_DATA] = make(map[string]interface{})
//				return common.ResponseTag_ParamError, resp
//			}
//
//			//检查一下俱乐部ID和平台是否对应
//			if club.Platform != plt_id {
//				resp[webapi.RESPONSE_STATE] = webapi.STATE_ERR
//				resp[webapi.RESPONSE_ERRMSG] = "PltId and ClubID is not match error!"
//				resp[webapi.RESPONSE_DATA] = make(map[string]interface{})
//				return common.ResponseTag_ParamError, resp
//			}
//
//			clubAreaArr := make([]*model.ClubAreaItem, 0)
//			for _, r := range club.RoomGroup {
//				item := &model.ClubAreaItem{
//					AreaID:          r.ClubRoomId,
//					GameName:        srvdata.PBDB_GameFreeMgr.GetData(int32(r.GameFreeId)).GetName(),
//					BaseScore:       r.BaseScore,
//					PlayingRoomsNum: 0,
//					CurrentRound:    0,
//					JoinPlayerNum:   0,
//					Ts:              r.Ts,
//				}
//				clubAreaArr = append(clubAreaArr, item)
//			}
//
//			club_scene := ClubSceneMgrSington.GetPlatformClub(club.Platform)
//			var csp *ClubScenePool
//			if club_scene == nil {
//				goto Next
//			}
//			for _, value := range club_scene {
//				if value.clubId == club.Id {
//					csp = value
//					break
//				}
//			}
//			if csp == nil {
//				goto Next
//			}
//			for _, v := range clubAreaArr {
//				for _, s := range csp.sceneList {
//					if s.clubRoomID == v.AreaID {
//						v.JoinPlayerNum += int32(len(s.players))
//						v.PlayingRoomsNum++
//						v.CurrentRound = s.currRound
//					}
//				}
//			}
//		Next:
//			//开始做分页
//			start := (page_num - 1) * page_size
//			end := page_num * page_size
//			count := len(clubAreaArr)
//			if count < start {
//				return common.ResponseTag_Ok, resp
//			}
//			if end > count {
//				end = count
//			}
//			sort.Slice(clubAreaArr, func(i, j int) bool {
//				return clubAreaArr[i].Ts > clubAreaArr[j].Ts
//			})
//			r := clubAreaArr[start:end]
//			resp["PageCount"] = count
//			resp["PageNo"] = page_num
//			resp["PageSize"] = page_size
//			resp[webapi.RESPONSE_ERRMSG] = ""
//			resp[webapi.RESPONSE_STATE] = webapi.STATE_OK
//			resp[webapi.RESPONSE_DATA] = r
//			dataResp := &common.M2GWebApiResponse{}
//			dataResp.Body, _ = resp.Marshal()
//			tNode.TransRep.RetFiels = dataResp
//			tNode.Resume()
//			return common.ResponseTag_Ok, resp
//		}))
//
//	//-------------------------------------------------------------------------------------------------------
//	//获取俱乐部下面的游戏房间信息(正式)
//	WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/api/Club/ClubRoomSet", WebAPIHandlerWrapper(
//		func(tNode *transact.TransNode, params webapi.RequestBody) (int, webapi.ResponseBody) {
//			params_data, _ := params.GetRequestBody("Param")
//			resp := webapi.NewResponseBody()
//			plt_id, _ := params_data.GetStr("PltID")       //必填
//			club_id, _ := params_data.GetInt("ClubID")     //必填
//			page_size, _ := params_data.GetInt("PageSize") //必填
//			page_num, _ := params_data.GetInt("PageNum")   //必填
//
//			//判断参数是否合法
//			if club_id <= 0 {
//				resp[webapi.RESPONSE_STATE] = webapi.STATE_ERR
//				resp[webapi.RESPONSE_ERRMSG] = "ClubID value error!"
//				resp[webapi.RESPONSE_DATA] = make(map[string]interface{})
//				return common.ResponseTag_ParamError, resp
//			}
//
//			//判断俱乐部号码是否存在
//			club := clubManager.clubList[int32(club_id)]
//			if club == nil {
//				resp[webapi.RESPONSE_STATE] = webapi.STATE_ERR
//				resp[webapi.RESPONSE_ERRMSG] = "ClubID is not exist error!"
//				resp[webapi.RESPONSE_DATA] = make(map[string]interface{})
//				return common.ResponseTag_ParamError, resp
//			}
//
//			//检查一下俱乐部ID和平台是否对应
//			if club.Platform != plt_id {
//				resp[webapi.RESPONSE_STATE] = webapi.STATE_ERR
//				resp[webapi.RESPONSE_ERRMSG] = "PltId ClubID is not match error!"
//				resp[webapi.RESPONSE_DATA] = make(map[string]interface{})
//				return common.ResponseTag_ParamError, resp
//			}
//
//			clubRoomDetailArr := make([]model.ClubRoomDetailItem, 0)
//			var csp *ClubScenePool
//			//通过俱乐部的平台找到这个平台下的所有场景
//			club_scenes := ClubSceneMgrSington.GetPlatformClub(club.Platform)
//			if len(club_scenes) == 0 {
//				goto NEXT
//			}
//
//			//根据俱乐部ID找到该俱乐部下的ClubScenePool信息
//			for _, value := range club_scenes {
//				if value.clubId == club.Id {
//					csp = value
//					break
//				}
//			}
//			if csp == nil {
//				goto NEXT
//			}
//
//			//待优化数据结构
//			for _, s := range csp.sceneList {
//				players := make([]int32, 0)
//				for _, p := range s.players {
//					players = append(players, p.SnId)
//				}
//
//				item := model.ClubRoomDetailItem{
//					DeskID:      int32(s.sceneId),
//					GameName:    srvdata.PBDB_GameFreeMgr.GetData(int32(s.gameId)).GetName(),
//					BaseBet:     0,
//					AreaID:      s.clubRoomID,
//					PlayerSnids: players,
//				}
//				for _, r := range club.RoomGroup {
//					if r.ClubRoomId == s.clubRoomID {
//						item.BaseBet = r.BaseScore
//					}
//				}
//				clubRoomDetailArr = append(clubRoomDetailArr, item)
//			}
//		NEXT:
//			//开始做分页
//			start := (page_num - 1) * page_size
//			end := page_num * page_size
//			count := len(clubRoomDetailArr)
//			if count < start {
//				return common.ResponseTag_Ok, resp
//			}
//			if end > count {
//				end = count
//			}
//			sort.Slice(clubRoomDetailArr, func(i, j int) bool {
//				return clubRoomDetailArr[i].DeskID < clubRoomDetailArr[j].DeskID
//			})
//			r := clubRoomDetailArr[start:end]
//			resp["PageCount"] = count
//			resp["PageNo"] = page_num
//			resp["PageSize"] = page_size
//			resp[webapi.RESPONSE_STATE] = webapi.STATE_OK
//			resp[webapi.RESPONSE_DATA] = r
//			dataResp := &common.M2GWebApiResponse{}
//			dataResp.Body, _ = resp.Marshal()
//			tNode.TransRep.RetFiels = dataResp
//			tNode.Resume()
//			return common.ResponseTag_Ok, resp
//		}))
//
//	//-------------------------------------------------------------------------------------------------------
//	//后台发送俱乐部创建的审核结果，同意或者拒绝，并自动发送邮件给该俱乐部会长
//	WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/api/Club/ClubCreateCheck", WebAPIHandlerWrapper(
//		func(tNode *transact.TransNode, params webapi.RequestBody) (int, webapi.ResponseBody) {
//			params_data, _ := params.GetRequestBody("Param")
//			resp := webapi.NewResponseBody()
//			plt_id, _ := params_data.GetStr("PltID")
//			club_id_str, _ := params_data.GetStr("ClubID")
//			isAgree, _ := params_data.GetBool("IsAgree")
//			desc, _ := params_data.GetStr("Desc") //如果isAgree=true则desc可以为空
//
//			//检测平台参数
//			plt := PlatformMgrSington.GetPlatform(plt_id)
//			if plt == nil {
//				resp[webapi.RESPONSE_STATE] = webapi.STATE_ERR
//				resp[webapi.RESPONSE_ERRMSG] = "Can't find PltID info error!"
//				resp[webapi.RESPONSE_DATA] = make(map[string]interface{})
//				return common.ResponseTag_ParamError, resp
//			}
//
//			//判断逻辑是否合理
//			if !isAgree && len(desc) == 0 {
//				resp[webapi.RESPONSE_STATE] = webapi.STATE_ERR
//				resp[webapi.RESPONSE_ERRMSG] = "if IsAgree==false then Desc must be have value error!"
//				resp[webapi.RESPONSE_DATA] = make(map[string]interface{})
//				return common.ResponseTag_ParamError, resp
//			}
//			club_id_str_arr := strings.Split(club_id_str, ",")
//			var club_id_int_arr = make([]int32, 0)
//			for _, v := range club_id_str_arr {
//				i, err := strconv.Atoi(v)
//				if err == nil {
//					if i <= 0 {
//						resp[webapi.RESPONSE_STATE] = webapi.STATE_ERR
//						resp[webapi.RESPONSE_ERRMSG] = fmt.Sprintf("ClubID=%v value error!", i)
//						resp[webapi.RESPONSE_DATA] = make(map[string]interface{})
//						return common.ResponseTag_ParamError, resp
//					}
//					club_id_int_arr = append(club_id_int_arr, int32(i))
//				} else {
//					logger.Logger.Errorf("/api/Club/ClubCreateCheck arg err:%v", err)
//				}
//			}
//
//			//判断参数是否合法
//			for _, club_id := range club_id_int_arr {
//				//判断俱乐部号码是否存在
//				club := clubManager.clubList[int32(club_id)]
//				if club == nil {
//					logger.Logger.Errorf("ClubID=%v is not exist error!", club_id)
//					continue
//				}
//				//检查一下避免后台连点
//				if club.Activity {
//					continue
//				}
//				if isAgree {
//					club.Activity = true
//					//日志信息
//					club.SaveClubOpLog(&model.ClubLog{
//						SnId:    club.Owner,
//						Name:    club.OwnerName,
//						LogType: model.ClubLogType_Create,
//					})
//					task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
//						model.SaveClub(club.Club)
//						return nil
//					}), task.CompleteNotifyWrapper(func(data interface{}, t *task.Task) {
//					}), "SaveClub").StartByFixExecutor(club.Platform)
//					//邮件
//					sendClubMail_ClubCreateSucces(club.Owner, club.Id, club.OwnerName, club.Name, plt_id)
//					//给会长发送更新审核的俱乐部
//					if p, ok := club.memberList[club.Owner]; ok {
//						pack := &protocol.SCClubSyncList{
//							ClubData: clubManager.GetClubBaseInfo(club.Id),
//							ClubId:   proto.Int32(club.Id),
//						}
//						logger.Logger.Info("SCClubSyncList: ", pack)
//						p.SendToClient(protocol.ClubPacketID_PACKET_SC_CLUBSYNCLIST, pack)
//						pp := PlayerMgrSington.GetPlayerBySnId(p.SnId)
//						if pp != nil {
//							pp.CreateClubNum++
//						} else {
//							task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
//								return model.UpdateCreateCreateClubNum(p.SnId)
//							}), nil, "UpdateCreateCreateClubNum").StartByFixExecutor("UpdateCreateCreateClubNum")
//						}
//					}
//					ClubSceneMgrSington.OnClubCreate(club.Id)
//				} else {
//					clubManager.ClubDestory(int32(club_id))
//					sendClubMail_ClubCreateFail(club.Owner, club.Id, club.OwnerName, club.Name, plt_id, desc)
//				}
//			}
//
//			resp[webapi.RESPONSE_ERRMSG] = ""
//			resp[webapi.RESPONSE_STATE] = webapi.STATE_OK
//			dataResp := &common.M2GWebApiResponse{}
//			dataResp.Body, _ = resp.Marshal()
//			tNode.TransRep.RetFiels = dataResp
//			tNode.Resume()
//			return common.ResponseTag_Ok, resp
//		}))
//
//	//-------------------------------------------------------------------------------------------------------
//	//后台发送俱乐部公告修改的审核结果，同意或者拒绝，并自动发送邮件
//	WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/api/Club/ClubNoticeCheck", WebAPIHandlerWrapper(
//		func(tNode *transact.TransNode, params webapi.RequestBody) (int, webapi.ResponseBody) {
//			params_data, _ := params.GetRequestBody("Param")
//			resp := webapi.NewResponseBody()
//			plt_id, _ := params_data.GetStr("PltID")
//			op_snid, _ := params_data.GetInt("OpSnid")
//			club_id_str, _ := params_data.GetStr("ClubID")
//			isAgree, _ := params_data.GetBool("IsAgree")
//			desc, _ := params_data.GetStr("Desc") //如果isAgree=true则desc可以为空
//
//			//检测平台参数
//			plt := PlatformMgrSington.GetPlatform(plt_id)
//			if plt == nil {
//				resp[webapi.RESPONSE_STATE] = webapi.STATE_ERR
//				resp[webapi.RESPONSE_ERRMSG] = "Can't find PltID info error!"
//				resp[webapi.RESPONSE_DATA] = make(map[string]interface{})
//				return common.ResponseTag_ParamError, resp
//			}
//
//			//判断逻辑是否合理
//			if !isAgree && len(desc) == 0 {
//				resp[webapi.RESPONSE_STATE] = webapi.STATE_ERR
//				resp[webapi.RESPONSE_ERRMSG] = "if IsAgree==false then Desc must be has value error!"
//				resp[webapi.RESPONSE_DATA] = make(map[string]interface{})
//				return common.ResponseTag_ParamError, resp
//			}
//
//			club_id_str_arr := strings.Split(club_id_str, ",")
//			var club_id_int_arr = make([]int32, 0)
//			for _, v := range club_id_str_arr {
//				i, err := strconv.Atoi(v)
//				if err == nil {
//					if i <= 0 {
//						resp[webapi.RESPONSE_STATE] = webapi.STATE_ERR
//						resp[webapi.RESPONSE_ERRMSG] = fmt.Sprintf("ClubID=%v value error!", i)
//						resp[webapi.RESPONSE_DATA] = make(map[string]interface{})
//						return common.ResponseTag_ParamError, resp
//					}
//					club_id_int_arr = append(club_id_int_arr, int32(i))
//				} else {
//					logger.Logger.Errorf("/api/Club/ClubNoticeCheck arg err:%v", err)
//				}
//			}
//
//			for _, club_id := range club_id_int_arr {
//				//判断俱乐部号码是否存在
//				club := clubManager.clubList[int32(club_id)]
//				if club == nil {
//					logger.Logger.Warnf("ClubID=%v is not exist error!", club_id)
//					continue
//				}
//
//				//检查一下避免后台连点
//				if len(club.BillboardNew) != 0 {
//					if isAgree {
//						opp := club.memberList[int32(op_snid)]
//						if opp != nil {
//							club.SaveClubOpLog(&model.ClubLog{
//								SnId:    int32(op_snid),
//								Name:    opp.Name,
//								LogType: model.ClubLogType_NewBillboard,
//							})
//						}
//						club.Billboard = club.BillboardNew
//						club.BillboardNew = ""
//						club.BillboardTs = time.Now().Unix()
//						sendClubMail_ClubEditNoticeSucces(club.Owner, club.Id, club.OwnerName, club.Name, plt_id)
//					} else {
//						club.BillboardNew = ""
//						sendClubMail_ClubEditNoticeFail(club.Owner, club.Id, club.OwnerName, club.Name, plt_id, desc)
//					}
//					//修改完通知客户端
//					pack := &protocol.SCClubBillboard{
//						Billboard:      proto.String(club.Billboard),
//						BillboardState: protocol.SCClubBillboard_ClubNoticeUnLock,
//					}
//					proto.SetDefaults(pack)
//					club.Broadcast(protocol.ClubPacketID_PACKET_SC_CLUBBILLBOARD, pack, 0)
//				}
//			}
//
//			resp[webapi.RESPONSE_ERRMSG] = ""
//			resp[webapi.RESPONSE_STATE] = webapi.STATE_OK
//			dataResp := &common.M2GWebApiResponse{}
//			dataResp.Body, _ = resp.Marshal()
//			tNode.TransRep.RetFiels = dataResp
//			tNode.Resume()
//			return common.ResponseTag_Ok, resp
//		}))
//
//	//-------------------------------------------------------------------------------------------------------
//	//俱乐部充值
//	WebAPIHandlerMgrSingleton.RegisteWebAPIHandler("/api/Club/AddClubGoldById", WebAPIHandlerWrapper(
//		func(tNode *transact.TransNode, params webapi.RequestBody) (int, webapi.ResponseBody) {
//			params_data, _ := params.GetRequestBody("Param")
//			resp := webapi.NewResponseBody()
//
//			member_snid, _ := params_data.GetInt("ID")
//			coin, _ := params_data.GetInt64("Gold")
//			coinExt, _ := params_data.GetInt64("GoldExt")
//			gold_desc, _ := params_data.GetStr("Desc")
//			billNo, _ := params_data.GetInt("BillNo")
//			platform, _ := params_data.GetStr("Platform")
//			if CacheDataMgr.CacheBillCheck(billNo) {
//				resp[webapi.RESPONSE_STATE] = webapi.STATE_ERR
//				resp[webapi.RESPONSE_ERRMSG] = "Bill number repeated!"
//				resp[webapi.RESPONSE_DATA] = make(map[string]interface{})
//				return common.ResponseTag_ParamError, resp
//			}
//			CacheDataMgr.CacheBillNumber(billNo) //防止手抖点两下
//
//			var existBillNo bool
//			var err error
//			var pd *model.PlayerData
//			oldGold := int64(0)
//			oldSafeBoxGold := int64(0)
//			var timeStamp = time.Now().UnixNano()
//			type PlayerCoinData struct {
//				ID       int32
//				ClubCoin int64
//			}
//			player := PlayerMgrSington.GetPlayerBySnId(int32(member_snid))
//			task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
//				if player == nil {
//					pd = model.GetPlayerDataBySnId(int32(member_snid), true)
//					if pd == nil {
//						return errors.New("Player data not find.")
//					}
//				} else {
//					pd = player.PlayerData
//				}
//				if len(platform) > 0 && pd.Platform != platform {
//					return errors.New("Platform error.")
//				}
//
//				log, err := model.GetPayCoinLogByBillNo(int32(member_snid), int64(billNo))
//				if err == nil && log != nil && log.BillNo == int64(billNo) && log.SnId == int32(member_snid) {
//					existBillNo = true
//					return fmt.Errorf("paycoin billno(%v) exist!", billNo)
//				}
//				oldGold = pd.ClubCoin
//				oldSafeBoxGold = pd.SafeBoxCoin
//				coinLog := model.NewPayCoinLog(int64(billNo), int32(member_snid), coin, common.GainWay_ClubPay,
//					fmt.Sprintf("RechargeById-%v", gold_desc), model.PayCoinLogType_Club, coinExt)
//				//修正时间
//				timeStamp = coinLog.TimeStamp
//				err = model.InsertPayCoinLogs(coinLog)
//				if err != nil {
//					return err
//				}
//				//增加帐变记录
//				coinlogex := model.NewCoinLogEx(int32(member_snid), coin+coinExt, oldGold+coin+coinExt,
//					oldSafeBoxGold, 0, common.GainWay_ClubPay, 0, "超管加币",
//					gold_desc, pd.Platform, pd.Channel, pd.BeUnderAgentCode, 2, pd.PackageID, 0)
//				err = model.InsertCoinLogs(coinlogex)
//				if err != nil {
//					//回滚到对账日志
//					model.RemovePayCoinLog(coinLog.LogId)
//					return err
//				}
//				return err
//			}), task.CompleteNotifyWrapper(func(data interface{}, t *task.Task) {
//				CacheDataMgr.ClearCacheBill(billNo)
//				if data != nil {
//					if existBillNo {
//						resp[webapi.RESPONSE_STATE] = webapi.STATE_OK
//						resp[webapi.RESPONSE_ERRMSG] = data.(error).Error()
//						resp[webapi.RESPONSE_DATA] = make(map[string]interface{})
//					} else {
//						resp[webapi.RESPONSE_STATE] = webapi.STATE_ERR
//						resp[webapi.RESPONSE_ERRMSG] = data.(error).Error()
//						resp[webapi.RESPONSE_DATA] = make(map[string]interface{})
//					}
//				} else {
//					if player != nil {
//						pd = player.PlayerData
//
//						player.ClubCoin += (coin + coinExt)
//						player.SetClubPayTs(timeStamp)
//						player.dirty = true
//						player.ClubCoinPayTotal += coin
//						if !model.GameParamData.CloseOftenSavePlayerData {
//							player.Time2Save()
//						}
//						player.SendDiffData()
//						pcd := &PlayerCoinData{
//							ID:       int32(member_snid),
//							ClubCoin: player.ClubCoin,
//						}
//						resp[webapi.RESPONSE_STATE] = webapi.STATE_OK
//						resp[webapi.RESPONSE_ERRMSG] = ""
//						resp[webapi.RESPONSE_DATA] = pcd
//					} else {
//						if pd != nil {
//							pcd := &PlayerCoinData{
//								ID:       int32(member_snid),
//								ClubCoin: pd.ClubCoin + (coin + coinExt),
//							}
//							resp[webapi.RESPONSE_STATE] = webapi.STATE_OK
//							resp[webapi.RESPONSE_ERRMSG] = ""
//							resp[webapi.RESPONSE_DATA] = pcd
//						} else {
//							logger.Logger.Errorf("%v Recharge %v coin failed on %v,bill no %v", member_snid, coin, time.Now(), billNo)
//							resp[webapi.RESPONSE_STATE] = webapi.STATE_ERR
//							resp[webapi.RESPONSE_ERRMSG] = "Recharge error"
//							resp[webapi.RESPONSE_DATA] = make(map[string]interface{})
//						}
//						player = PlayerMgrSington.GetPlayerBySnId(int32(member_snid))
//						if player != nil {
//							if timeStamp > player.ClubCoinPayTs { //金币冲账
//								logger.Logger.Warnf("/api/Member/AddClubGoldById(%v) found ClubCoinPayTs not fit,try auto op(%v) curr(%v)",
//									player.SnId, timeStamp, player.ClubCoinPayTs)
//								coinlogs, err := model.GetAllPayClubCoinLog(player.SnId, player.ClubCoinPayTs)
//								if err == nil {
//									var dirty bool
//									var cnt int64
//									var cntExt int64
//									for i := 0; i < len(coinlogs); i++ {
//										cnt = coinlogs[i].Coin
//										cntExt = coinlogs[i].CoinEx
//
//										player.ClubCoin += cnt + cntExt
//										player.ClubCoinPayTotal += cnt
//										player.SetClubPayTs(coinlogs[i].TimeStamp)
//										dirty = true
//									}
//									if dirty {
//										player.dirty = true
//										if !model.GameParamData.CloseOftenSavePlayerData {
//											player.Time2Save()
//										}
//										player.SendDiffData()
//									}
//								}
//							}
//						} else {
//							//更新线下缓存数据
//							op := OfflinePlayerMgrSington.GetPlayer(int32(member_snid))
//							if op != nil {
//								op.ClubCoin += (coin + coinExt)
//								op.SetClubPayTs(timeStamp)
//								op.ClubCoinPayTotal += coin
//							}
//						}
//					}
//				}
//
//				dataResp := &common.M2GWebApiResponse{}
//				dataResp.Body, err = resp.Marshal()
//				tNode.TransRep.RetFiels = dataResp
//				tNode.Resume()
//				if err != nil {
//					logger.Logger.Error("/api/Member/AddClubGoldById task marshal data error:", err)
//				}
//			}), "/api/Member/AddClubGoldById").Start()
//			return common.ResponseTag_TransactYield, resp
//		}))
//}
//
////请求俱乐部的包间流水信息
////注意是阻塞协程
////DateTs参数为请求某一天的时间戳，只要该时间戳在这一天之内即可
//func reqClubTurnover(clubID int32, DateTs int64) []model.AreaTurnoverItem {
//	if clubID <= 0 {
//		return nil
//	}
//	buff, err := webapi.ReqClubTurnover(common.GetAppId(), clubID, DateTs)
//	type ApiResult struct {
//		Msg []model.AreaTurnoverItem `json:"Msg"`
//		Tag int                      `json:"Tag"`
//	}
//	result := ApiResult{}
//	err = json.Unmarshal(buff, &result)
//	if err != nil {
//		logger.Logger.Error("ReqClubTurnover:  ", string(buff))
//		logger.Logger.Errorf(fmt.Sprintf("reqClubTurnover json.Unmarshal failed._%v", err))
//		return nil
//	}
//	return result.Msg
//}
//
////请求俱乐部的包间流水信息
////注意是阻塞协程
////DateTs参数为请求某一天的时间戳，只要该时间戳在这一天之内即可
////PageNum为要请求那一页
//func reqClubRoomTurnover(clubID int32, clubRoomID string, PageSize, PageNum int32, DateTs int64) *model.RoundTurnoverDetail {
//	buff, err := webapi.ReqClubRoomPumpDetail(common.GetAppId(), clubID, clubRoomID, PageSize, PageNum, DateTs)
//	type ApiResult struct {
//		Msg *model.RoundTurnoverDetail `json:"Msg"`
//		Tag int                        `json:"Tag"`
//	}
//	result := ApiResult{}
//	err = json.Unmarshal(buff, &result)
//	if err != nil {
//		logger.Logger.Errorf(fmt.Sprintf("reqClubRoomTurnover json.Unmarshal failed._%v", err))
//		return nil
//	}
//	//fmt.Println(string(buff))
//	if result.Tag != 0 {
//		logger.Logger.Errorf(fmt.Sprintf("reqClubRoomTurnover result code failed._%v", result.Msg))
//		return nil
//	} else {
//		return result.Msg
//	}
//}
