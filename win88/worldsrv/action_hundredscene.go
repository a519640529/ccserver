package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"games.yol.com/win88/common"
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/gamehall"
	"games.yol.com/win88/srvdata"
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/core/task"
)

type CSHundredSceneGetPlayerNumPacketFactory struct {
}
type CSHundredSceneGetPlayerNumHandler struct {
}

func (this *CSHundredSceneGetPlayerNumPacketFactory) CreatePacket() interface{} {
	pack := &gamehall.CSHundredSceneGetPlayerNum{}
	return pack
}

func (this *CSHundredSceneGetPlayerNumHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSHundredSceneGetPlayerNumHandler Process recv ", data)
	if msg, ok := data.(*gamehall.CSHundredSceneGetPlayerNum); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p != nil {
			nums := HundredSceneMgrSington.GetPlayerNums(p, msg.GetGameId(), msg.GetGameModel())
			pack := &gamehall.SCHundredSceneGetPlayerNum{
				Nums: nums,
			}
			proto.SetDefaults(pack)
			p.SendToClient(int(gamehall.HundredScenePacketID_PACKET_SC_HUNDREDSCENE_GETPLAYERNUM), pack)
			logger.Logger.Trace("SCHundredSceneGetPlayerNum:", pack)
		}
	}

	return nil
}

type CSHundredSceneOpPacketFactory struct {
}
type CSHundredSceneOpHandler struct {
}

func (this *CSHundredSceneOpPacketFactory) CreatePacket() interface{} {
	pack := &gamehall.CSHundredSceneOp{}
	return pack
}

func (this *CSHundredSceneOpHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSHundredSceneOpHandler Process recv ", data)
	if msg, ok := data.(*gamehall.CSHundredSceneOp); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p != nil {
			var ret gamehall.OpResultCode_Hundred
			pack := &gamehall.SCHundredSceneOp{
				Id:     msg.Id,
				OpType: msg.OpType,
			}
			oldPlatform := p.Platform
			switch msg.GetOpType() {
			case HundredSceneOp_Enter:
				//pt := PlatformMgrSington.GetPackageTag(p.PackageID)
				//if pt != nil && pt.IsForceBind == 1 {
				//	if p.BeUnderAgentCode == "" || p.BeUnderAgentCode == "0" {
				//		ret = gamehall.OpResultCode_Hundred_OPRC_MustBindPromoter_Hundred
				//		goto done
				//	}
				//}
				if p.scene != nil {
					logger.Logger.Warnf("CSHundredSceneOpHandler CoinSceneOp_Enter found snid:%v had in scene:%v gameid:%v", p.SnId, p.scene.sceneId, p.scene.gameId)
					p.ReturnScene(false)
					return nil
				}
				dbGameFree := srvdata.PBDB_GameFreeMgr.GetData(msg.GetId())
				gameId := dbGameFree.GetGameId()
				var roomId int32
				params := msg.GetOpParams()
				if len(params) != 0 {
					roomId = params[0]
					name, ok := HundredSceneMgrSington.GetPlatformNameBySceneId(roomId)
					if p.IsRob {
						//机器人先伪装成对应平台的用户
						if ok {
							p.Platform = name
						}
					} else if p.GMLevel > 0 && p.Platform == name { //允许GM直接按房间ID进场
						roomId = params[0]
					}
				}
				//检测房间状态是否开启
				gps := PlatformMgrSington.GetGameConfig(p.Platform, msg.GetId())
				if gps == nil {
					ret = gamehall.OpResultCode_Hundred_OPRC_RoomHadClosed_Hundred
					goto done
				}

				dbGameFree = gps.DbGameFree
				if gps.GroupId != 0 {
					if len(params) != 0 && p.GMLevel > 0 { //允许GM直接按房间ID进场
						s := SceneMgrSington.GetScene(int(params[0]))
						if s != nil && s.groupId == gps.GroupId {
							roomId = params[0]
						}
					}
				}

				if dbGameFree == nil {
					ret = gamehall.OpResultCode_Hundred_OPRC_RoomHadClosed_Hundred
					goto done
				}

				if dbGameFree.GetLimitCoin() != 0 && int64(dbGameFree.GetLimitCoin()) > p.Coin {
					ret = gamehall.OpResultCode_Hundred_OPRC_CoinNotEnough_Hundred
					goto done
				}

				if dbGameFree.GetMaxCoinLimit() != 0 && int64(dbGameFree.GetMaxCoinLimit()) < p.Coin && !p.IsRob {
					ret = gamehall.OpResultCode_Hundred_OPRC_CoinTooMore_Hundred
					goto done
				}

				//检查游戏次数限制
				if !p.IsRob {
					//todayData, _ := p.GetDaliyGameData(int(dbGameFree.GetId()))
					//if dbGameFree.GetPlayNumLimit() != 0 &&
					//	todayData != nil &&
					//	todayData.GameTimes >= int64(dbGameFree.GetPlayNumLimit()) {
					//	ret = gamehall.OpResultCode_Hundred_OPRC_RoomGameTimes_Hundred
					//	goto done
					//}
				}

				gameVers := srvdata.GetGameVers(p.PackageID)
				if gameVers != nil {
					if ver, ok := gameVers[fmt.Sprintf("%v,%v", gameId, p.Channel)]; ok {
						pack.MinApkVer = proto.Int32(ver.MinApkVer)
						pack.MinResVer = proto.Int32(ver.MinResVer)
						pack.LatestApkVer = proto.Int32(ver.LatestApkVer)
						pack.LatestResVer = proto.Int32(ver.LatestResVer)

						if msg.GetApkVer() < ver.MinApkVer {
							ret = gamehall.OpResultCode_Hundred_OPRC_YourAppVerIsLow_Hundred
							goto done
						}

						if msg.GetResVer() < ver.MinResVer {
							ret = gamehall.OpResultCode_Hundred_OPRC_YourResVerIsLow_Hundred
							goto done
						}
					}
				}

				ret = HundredSceneMgrSington.PlayerEnter(p, msg.GetId())
				if p.scene != nil {
					pack.OpParams = append(pack.OpParams, msg.GetId())
					//TODO 有房间还进入失败，尝试returnroom
					if ret != gamehall.OpResultCode_Hundred_OPRC_Sucess_Hundred {
						p.ReturnScene(false)
						return nil
					}
				}
			case HundredSceneOp_Leave:
				ret = HundredSceneMgrSington.PlayerTryLeave(p)
			case HundredSceneOp_Change:
				/*var exclude int32
				if p.scene != nil {
					exclude = int32(p.scene.sceneId)
				}
				params := msg.GetOpParams()
				if len(params) != 0 {
					exclude = params[0]
				}
				if HundredSceneMgrSington.PlayerInChanging(p) { //换桌中
					return nil
				}
				ret = HundredSceneMgrSington.PlayerTryChange(p, msg.GetId(), exclude)*/
			}
		done:
			//机器人要避免身上的平台标记被污染
			if p.IsRob {
				if ret != gamehall.OpResultCode_Hundred_OPRC_Sucess_Hundred {
					p.Platform = oldPlatform
				}
			}
			pack.OpCode = ret
			proto.SetDefaults(pack)
			p.SendToClient(int(gamehall.HundredScenePacketID_PACKET_SC_HUNDREDSCENE_OP), pack)
			//}
			if msg.GetOpType() == common.CoinSceneOp_Enter && ret == gamehall.OpResultCode_Hundred_OPRC_Sucess_Hundred && p.scene != nil {
				gameName := p.scene.dbGameFree.GetName() + p.scene.dbGameFree.GetTitle()
				ActMonitorMgrSington.SendActMonitorEvent(ActState_Game, p.SnId, p.Name, p.Platform,
					0, 0, gameName, 0)
			}
		}
	}
	return nil
}

type CSGameObservePacketFactory struct {
}
type CSGameObserveHandler struct {
}

func (this *CSGameObservePacketFactory) CreatePacket() interface{} {
	pack := &gamehall.CSGameObserve{}
	return pack
}

func (this *CSGameObserveHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSGameObserveHandler Process recv ", data)
	if msg, ok := data.(*gamehall.CSGameObserve); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p != nil {
			if msg.GetStartOrEnd() {
				gameStateMgr.PlayerRegiste(p, msg.GetGameId(), msg.GetStartOrEnd())
				pack := &gamehall.SCGameSubList{}
				statePack := &gamehall.SCGameState{}
				scenes := HundredSceneMgrSington.GetPlatformScene(p.Platform, msg.GetGameId())
				for _, value := range scenes {
					pack.List = append(pack.List, &gamehall.GameSubRecord{
						GameFreeId: proto.Int32(value.dbGameFree.GetId()),
						NewLog:     proto.Int32(-1),
						LogCnt:     proto.Int(len(value.GameLog)),
						TotleLog:   value.GameLog,
					})
					leftTime := int64(value.StateSec) - (time.Now().Unix() - value.StateTs)
					if leftTime < 0 {
						leftTime = 0
					}
					statePack.List = append(statePack.List, &gamehall.GameState{
						GameFreeId: proto.Int32(value.dbGameFree.GetId()),
						Ts:         proto.Int64(leftTime),
						Sec:        proto.Int32(value.StateSec),
					})
				}
				p.SendToClient(int(gamehall.GameHallPacketID_PACKET_SC_GAMESUBLIST), pack)
				logger.Logger.Trace("SCGameSubList:", pack)
				p.SendToClient(int(gamehall.GameHallPacketID_PACKET_SC_GAMESTATE), statePack)
				logger.Logger.Trace("SCGameState:", statePack)
			} else {
				gameStateMgr.PlayerClear(p)
			}
		}
	}
	return nil
}

type CSHundredSceneGetGameJackpotPacketFactory struct {
}
type CSHundredSceneGetGameJackpotHandler struct {
}

func (this *CSHundredSceneGetGameJackpotPacketFactory) CreatePacket() interface{} {
	pack := &gamehall.CSHundredSceneGetPlayerNum{}
	return pack
}

func (this *CSHundredSceneGetGameJackpotHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSHundredSceneGetGameJackpotHandler Process recv ", data)
	if msg, ok := data.(*gamehall.CSHundredSceneGetPlayerNum); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p != nil {
			gameid := int(msg.GetGameId())
			// 冰河世纪, 百战成神, 财神, 复仇者联盟, 复活岛
			if gameid == common.GameId_IceAge || gameid == common.GameId_TamQuoc || gameid == common.GameId_CaiShen ||
				gameid == common.GameId_Avengers || gameid == common.GameId_EasterIsland {
				gameStateMgr.PlayerRegiste(p, msg.GetGameId(), true)
				pack := &gamehall.SCHundredSceneGetGameJackpot{}
				scenes := HundredSceneMgrSington.GetPlatformScene(p.Platform, msg.GetGameId())
				for _, v := range scenes {
					jpfi := &gamehall.GameJackpotFundInfo{
						GameFreeId:  proto.Int32(v.dbGameFree.GetId()),
						JackPotFund: proto.Int64(v.JackPotFund),
					}
					pack.GameJackpotFund = append(pack.GameJackpotFund, jpfi)
				}
				proto.SetDefaults(pack)
				p.SendToClient(int(gamehall.HundredScenePacketID_PACKET_SC_GAMEJACKPOT), pack)
				logger.Logger.Trace("SCHundredSceneGetGameJackpot:", pack)
			}
		}
	}
	return nil
}

type CSHundredSceneGetGameHistoryInfoPacketFactory struct {
}
type CSHundredSceneGetGameHistoryInfoHandler struct {
}

func (this *CSHundredSceneGetGameHistoryInfoPacketFactory) CreatePacket() interface{} {
	pack := &gamehall.CSHundredSceneGetHistoryInfo{}
	return pack
}

func (this *CSHundredSceneGetGameHistoryInfoHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("World CSHundredSceneGetGameHistoryInfoHandler Process recv ", data)
	if msg, ok := data.(*gamehall.CSHundredSceneGetHistoryInfo); ok {
		gameid := int(msg.GetGameId())
		historyModel := msg.GetGameHistoryModel()
		p := PlayerMgrSington.GetPlayer(sid)
		if p != nil {
			switch historyModel {
			case PLAYER_HISTORY_MODEL: // 历史记录
				task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
					var genPlayerHistoryInfo = func(spinID string, isFree bool, createdTime, totalBetValue, totalPriceValue, totalBonusValue, multiple int64, player *gamehall.PlayerHistoryInfo) {
						player.SpinID = proto.String(spinID)
						player.CreatedTime = proto.Int64(createdTime)
						player.TotalBetValue = proto.Int64(totalBetValue)
						player.TotalPriceValue = proto.Int64(totalPriceValue)
						player.IsFree = proto.Bool(isFree)
						player.TotalBonusValue = proto.Int64(totalBonusValue)
						player.Multiple = proto.Int64(multiple)
					}

					var genPlayerHistoryInfoMsg = func(spinid string, v *model.NeedGameRecord, gdl *model.GameDetailedLog, player *gamehall.PlayerHistoryInfo) {
						switch gameid {
						//case common.GameId_IceAge:
						//	data, err := model.UnMarshalIceAgeGameNote(gdl.GameDetailedNote)
						//	if err != nil {
						//		logger.Logger.Errorf("World UnMarshalIceAgeGameNote error:%v", err)
						//	}
						//	gnd := data.(*model.IceAgeType)
						//	genPlayerHistoryInfo(spinid, gnd.IsFree, int64(v.Ts), int64(gnd.Score), gnd.TotalPriceValue, gnd.TotalBonusValue, player)
						//case common.GameId_TamQuoc:
						//	data, err := model.UnMarshalTamQuocGameNote(gdl.GameDetailedNote)
						//	if err != nil {
						//		logger.Logger.Errorf("World UnMarshalTamQuocGameNote error:%v", err)
						//	}
						//	gnd := data.(*model.TamQuocType)
						//	genPlayerHistoryInfo(spinid, gnd.IsFree, int64(v.Ts), int64(gnd.Score), gnd.TotalPriceValue, gnd.TotalBonusValue, player)
						//case common.GameId_CaiShen:
						//	data, err := model.UnMarshalCaiShenGameNote(gdl.GameDetailedNote)
						//	if err != nil {
						//		logger.Logger.Errorf("World UnMarshalCaiShenGameNote error:%v", err)
						//	}
						//	gnd := data.(*model.CaiShenType)
						//	genPlayerHistoryInfo(spinid, gnd.IsFree, int64(v.Ts), int64(gnd.Score), gnd.TotalPriceValue, gnd.TotalBonusValue, player)
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
								if curplayer.UserId == p.SnId {
									genPlayerHistoryInfo(spinid, false, int64(v.Ts), int64(curplayer.UserBetTotal), curplayer.ChangeCoin, 0, int64(curplayer.UserMultiple), player)
									break
								}
							}
						case common.GameId_Avengers:
							data, err := model.UnMarshalAvengersGameNote(gdl.GameDetailedNote)
							if err != nil {
								logger.Logger.Errorf("World UnMarshalAvengersGameNote error:%v", err)
							}
							gnd := data.(*model.GameResultLog)
							genPlayerHistoryInfo(spinid, gnd.BaseResult.IsFree, int64(v.Ts), int64(gnd.BaseResult.TotalBet), gnd.BaseResult.WinTotal, gnd.BaseResult.WinSmallGame, 0, player)
						//case common.GameId_EasterIsland:
						//	data, err := model.UnMarshalEasterIslandGameNote(gdl.GameDetailedNote)
						//	if err != nil {
						//		logger.Logger.Errorf("World UnMarshalEasterIslandGameNote error:%v", err)
						//	}
						//	gnd := data.(*model.EasterIslandType)
						//	genPlayerHistoryInfo(spinid, gnd.IsFree, int64(v.Ts), int64(gnd.Score), gnd.TotalPriceValue, gnd.TotalBonusValue, player)
						default:
							logger.Logger.Errorf("World CSHundredSceneGetGameHistoryInfoHandler receive gameid(%v) error", gameid)
						}
					}

					gameclass := int32(2)
					spinid := strconv.FormatInt(int64(p.SnId), 10)
					dbGameFrees := srvdata.PBDB_GameFreeMgr.Datas.Arr //.GetData(data.DbGameFree.Id)
					roomtype := int32(0)
					for _, v := range dbGameFrees {
						if int32(gameid) == v.GetGameId() {
							gameclass = v.GetGameClass()
							roomtype = v.GetSceneType()
							break
						}
					}

					gpl := model.GetPlayerListByHallEx(p.SnId, p.Platform, 0, 50, 0, 0, roomtype, gameclass, gameid)
					pack := &gamehall.SCPlayerHistory{}
					for _, v := range gpl.Data {
						if v.GameDetailedLogId == "" {
							logger.Logger.Error("World PlayerHistory GameDetailedLogId is nil")
							break
						}
						gdl := model.GetPlayerHistory(p.Platform, v.GameDetailedLogId)
						player := &gamehall.PlayerHistoryInfo{}
						genPlayerHistoryInfoMsg(spinid, v, gdl, player)
						pack.PlayerHistory = append(pack.PlayerHistory, player)
					}
					proto.SetDefaults(pack)
					logger.Logger.Infof("World gameid:%v PlayerHistory:%v ", gameid, pack)
					return pack
				}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
					if data == nil {
						logger.Logger.Error("World PlayerHistory data is nil")
						return
					}
					p.SendToClient(int(gamehall.HundredScenePacketID_PACKET_SC_GAMEPLAYERHISTORY), data)
				}), "CSGetPlayerHistoryHandlerWorld").Start()
			case BIGWIN_HISTORY_MODEL: // 爆奖记录
				jackpotList := JackpotListMgrSington.GetJackpotList(gameid)
				//if len(jackpotList) < 1 {
				//	JackpotListMgrSington.GenJackpot(gameid) // 初始化爆奖记录
				//	JackpotListMgrSington.after(gameid)      // 开启定时器
				//	jackpotList = JackpotListMgrSington.GetJackpotList(gameid)
				//}
				pack := JackpotListMgrSington.GetStoCMsg(jackpotList)
				pack.GameId = msg.GetGameId()
				logger.Logger.Infof("World BigWinHistory: %v %v", gameid, pack)
				p.SendToClient(int(gamehall.HundredScenePacketID_PACKET_SC_GAMEBIGWINHISTORY), pack)
			case GAME_HISTORY_MODEL:
				task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
					var genGameHistoryInfo = func(gameNumber string, createdTime, multiple int64, hash string, gamehistory *gamehall.GameHistoryInfo) {
						gamehistory.GameNumber = proto.String(gameNumber)
						gamehistory.CreatedTime = proto.Int64(createdTime)
						gamehistory.Hash = proto.String(hash)
						gamehistory.Multiple = proto.Int64(multiple)
					}

					gls := model.GetAllGameDetailedLogsByGameIdAndTs(p.Platform, gameid, 20)

					pack := &gamehall.SCPlayerHistory{}
					for _, v := range gls {

						gamehistory := &gamehall.GameHistoryInfo{}

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
					logger.Logger.Infof("World gameid:%v History:%v ", gameid, pack)
					return pack
				}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
					if data == nil {
						logger.Logger.Error("World GameHistory data is nil")
						return
					}
					p.SendToClient(int(gamehall.HundredScenePacketID_PACKET_SC_GAMEPLAYERHISTORY), data)
				}), "CSGetGameHistoryHandlerWorld").Start()
			default:
				logger.Logger.Errorf("World CSHundredSceneGetGameHistoryInfoHandler receive historyModel(%v) error", historyModel)
			}
		}
	}
	return nil
}

func init() {
	common.RegisterHandler(int(gamehall.HundredScenePacketID_PACKET_CS_HUNDREDSCENE_GETPLAYERNUM), &CSHundredSceneGetPlayerNumHandler{})
	netlib.RegisterFactory(int(gamehall.HundredScenePacketID_PACKET_CS_HUNDREDSCENE_GETPLAYERNUM), &CSHundredSceneGetPlayerNumPacketFactory{})

	common.RegisterHandler(int(gamehall.HundredScenePacketID_PACKET_CS_HUNDREDSCENE_OP), &CSHundredSceneOpHandler{})
	netlib.RegisterFactory(int(gamehall.HundredScenePacketID_PACKET_CS_HUNDREDSCENE_OP), &CSHundredSceneOpPacketFactory{})
	//请求游戏列表
	common.RegisterHandler(int(gamehall.GameHallPacketID_PACKET_CS_GAMEOBSERVE), &CSGameObserveHandler{})
	netlib.RegisterFactory(int(gamehall.GameHallPacketID_PACKET_CS_GAMEOBSERVE), &CSGameObservePacketFactory{})

	//// 请求奖池信息
	//common.RegisterHandler(int(gamehall.HundredScenePacketID_PACKET_CS_GAMEJACKPOT), &CSHundredSceneGetGameJackpotHandler{})
	//netlib.RegisterFactory(int(gamehall.HundredScenePacketID_PACKET_CS_GAMEJACKPOT), &CSHundredSceneGetGameJackpotPacketFactory{})

	//// 请求历史记录和爆奖记录
	common.RegisterHandler(int(gamehall.HundredScenePacketID_PACKET_CS_GAMEHISTORYINFO), &CSHundredSceneGetGameHistoryInfoHandler{})
	netlib.RegisterFactory(int(gamehall.HundredScenePacketID_PACKET_CS_GAMEHISTORYINFO), &CSHundredSceneGetGameHistoryInfoPacketFactory{})
}
