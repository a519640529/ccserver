package main

import (
	"games.yol.com/win88/common"
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/server"
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/core/task"
	"github.com/idealeak/goserver/srvlib"
	"strconv"
	"time"
)

type DWThirdRebateMessagePacketFactory struct {
}
type DWThirdRebateMessageHandler struct {
}

func (this *DWThirdRebateMessagePacketFactory) CreatePacket() interface{} {
	pack := &server.DWThirdRebateMessage{}
	return pack
}
func (this *DWThirdRebateMessageHandler) Process(s *netlib.Session, packetid int, data interface{}) error {
	logger.Logger.Trace("DWThirdRebateMessageHandler Process recv ", data)
	if msg, ok := data.(*server.DWThirdRebateMessage); ok {
		//TODO
		SendAckToDataSrv(msg.GetTag(), 2)
		if msg.GetAvailableBet() <= 0 {
			logger.Logger.Warn("DWThirdRebateMessageHandler is error: AvailableBet= ", msg.GetAvailableBet())
			return nil
		}

		//p := PlayerMgrSington.GetPlayerBySnId(msg.GetSnid())
		//if p != nil {
		//	p.dirty = true
		//	//actRandCoinMgr.OnPlayerLiuShui(p, msg.GetAvailableBet())
		//}
		//
		//thirdId := strconv.Itoa(int(ThirdPltGameMappingConfig.FindThirdIdByThird(msg.GetThird())))
		//rebateTask := RebateInfoMgrSington.rebateTask[strconv.Itoa(int(msg.GetPlt()))]
		//if rebateTask != nil {
		//	Third := rebateTask.RebateGameThirdCfg[thirdId]
		//	if Third != nil {
		//		p := PlayerMgrSington.GetPlayerBySnId(msg.GetSnid())
		//		if p == nil {
		//			logger.Logger.Trace("DWThirdRebateMessageHandler p == nil ", msg.GetSnid())
		//			OfflinePlayerMgrSington.GetOfflinePlayer(msg.GetSnid(), func(op *OfflinePlayer, asyn bool) {
		//				if op == nil {
		//					return
		//				}
		//				if op.IsRob {
		//					return
		//				}
		//
		//				if data, ok := op.RebateData[thirdId]; ok {
		//					data.ValidBetTotal += msg.GetAvailableBet()
		//				} else {
		//					op.RebateData[thirdId] = &model.RebateData{
		//						ValidBetTotal: msg.GetAvailableBet(),
		//					}
		//				}
		//				task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
		//					//注意流控，防止该任务过渡占用登陆队列，可以在datasrv上配置心跳和maxdone来控制
		//					return model.SavePlayerRebate(op.PlayerData, thirdId)
		//				}), task.CompleteNotifyWrapper(func(data interface{}, tt *task.Task) {
		//					if data != nil {
		//						logger.Logger.Errorf("SavePlayerRebate error:%v snid:%v platform:%v AvailableBet:%v", data, msg.GetSnid(), msg.GetThird(), msg.GetAvailableBet())
		//					} else {
		//						p = PlayerMgrSington.GetPlayerBySnId(msg.GetSnid()) //说明更新任务排在了玩家登陆的后面(造成了脏读，重新应用下该次下注)
		//						if p != nil {
		//							if data, ok := p.RebateData[thirdId]; ok {
		//								data.ValidBetTotal += msg.GetAvailableBet()
		//							} else {
		//								p.RebateData[thirdId] = &model.RebateData{
		//									ValidBetTotal: msg.GetAvailableBet(),
		//								}
		//							}
		//							p.dirty = true
		//						}
		//					}
		//				}), "SavePlayerRebate").StartByExecutor(op.AccountId) //保证和玩家存取在一条线程内(避免脏读或者脏写)
		//			}, false)
		//			return nil
		//		}
		//		if p.IsRob {
		//			logger.Logger.Trace("DWThirdRebateMessageHandler p is rob ", msg.GetSnid())
		//			return nil
		//		}
		//		if data, ok := p.RebateData[thirdId]; ok {
		//			data.ValidBetTotal += msg.GetAvailableBet()
		//		} else {
		//			p.RebateData[thirdId] = &model.RebateData{
		//				ValidBetTotal: msg.GetAvailableBet(),
		//			}
		//		}
		//
		//		p.dirty = true
		//		//p.CountRebate(thirdId, 1)
		//	} else {
		//		logger.Logger.Trace("DWThirdRebateMessageHandler Third is nil. ", msg.GetPlt(), msg.GetThird())
		//	}
		//} else {
		//	logger.Logger.Trace("DWThirdRebateMessageHandler rebateTask is nil. ", msg.GetPlt(), msg.GetThird())
		//}
	}
	return nil
}

type DWThirdRoundMessagePacketFactory struct {
}
type DWThirdRoundMessageHandler struct {
}

func (this *DWThirdRoundMessagePacketFactory) CreatePacket() interface{} {
	pack := &server.DWThirdRoundMessage{}
	return pack
}
func (this *DWThirdRoundMessageHandler) Process(s *netlib.Session, packetid int, data interface{}) error {
	logger.Logger.Trace("DWThirdRoundMessageHandler Process recv ", data)

	if msg, ok := data.(*server.DWThirdRoundMessage); ok {
		//todo
		//获取到对应的gamefreeid,三方的特殊处理了，只寻找对应大类的第一个gamefreeid,因为这个游戏都在不停的变，很多都不一致
		if thirdID := ThirdPltGameMappingConfig.FindThirdIdByThird(msg.GetThird()); thirdID != 0 {
			var dbGamefreeInfo *server.DB_GameFree
			platform := msg.GetPlatform()
			if platform != 0 {
				pltGameInfo := PlatformMgrSington.GetGameConfigByThird(strconv.Itoa(int(platform)), thirdID)
				if pltGameInfo != nil {
					dbGamefreeInfo = pltGameInfo.DbGameFree
				}
			}

			player := PlayerMgrSington.GetPlayerBySnId(msg.GetSnid())
			if player != nil {
				str := strconv.Itoa(int(thirdID))
				//处理三方全民流水问题
				totalOut := int32(0)
				totalIn := int32(0)
				if dbGamefreeInfo != nil {
					isBind := int32(0)
					if player.Tel != "" {
						isBind = 1
					}

					if msg.GetProfitCoinInTime() < 0 {
						totalIn = msg.GetBetCoinInTime()
					} else {
						totalOut = msg.GetBetCoinInTime()
					}
					isQuMin := false
					//pt := PlatformMgrSington.GetPackageTag(player.PackageID)
					//if pt != nil && pt.SpreadTag == 1 {
					//	isQuMin = true
					//}
					if isQuMin || !model.GameParamData.QMOptimization {
						QMFlowMgr.AddPlayerStatement(player.SnId, isBind, totalOut, totalIn, thirdID,
							player.Platform, player.PackageID, dbGamefreeInfo)
					}

					availableBet := int64(totalOut + totalIn)
					availableBet = availableBet * int64(dbGamefreeInfo.GetBetWaterRate()) / 100
					if availableBet > 0 {
						player.TotalConvertibleFlow += availableBet
						player.TotalFlow += availableBet
						player.dirty = true
						//今日流水增加
						player.TodayGameData.TodayConvertibleFlow += availableBet
					}
				}
				if gd, ok := player.GDatas[str]; ok {
					gd.Statics.GameTimes += int64(msg.GetAccRoundsInTime())
					gd.Statics.TotalOut += int64(totalOut)
					gd.Statics.TotalIn += int64(totalIn)
					if gd.Statics.MaxSysOut < int64(msg.GetOneroundMaxwin()) {
						gd.Statics.MaxSysOut = int64(msg.GetOneroundMaxwin())
					}
				} else {
					player.GDatas[str] = &model.PlayerGameInfo{
						FirstTime: time.Now(),
						Statics: model.PlayerGameStatics{
							GameTimes: int64(msg.GetAccRoundsInTime()),
							MaxSysOut: int64(msg.GetOneroundMaxwin()),
							TotalIn:   int64(totalIn),
							TotalOut:  int64(totalOut),
						},
					}
				}

				if player.TotalGameData == nil {
					player.TotalGameData = make(map[int][]*model.PlayerGameTotal)
				}
				showId := 9
				if len(player.TotalGameData[showId]) == 0 {
					player.TotalGameData[showId] = []*model.PlayerGameTotal{new(model.PlayerGameTotal)}
				}
				cnt := len(player.TotalGameData[showId])
				if cnt > 0 {
					td := player.TotalGameData[showId][cnt-1]
					if td == nil {
						td = &model.PlayerGameTotal{}
						player.TotalGameData[showId][cnt-1] = td
					}
					if td != nil {
						td.ProfitCoin += int64(msg.GetProfitCoinInTime())
						td.BetCoin += int64(msg.GetBetCoinInTime())
						td.FlowCoin += int64(msg.GetFlowCoinInTime())
					}
				}

				//洗码
				//三方游戏,通过进出场的营收差洗码
				washingCoin := msg.GetProfitCoinInTime()
				if washingCoin < 0 {
					washingCoin = -washingCoin
				}
				washedCoin := player.WashingCoin(int64(washingCoin))
				if washedCoin > 0 {
					logger.Logger.Tracef("三方游戏洗码:snid=%v,washingCoin=%v,gamefreeid=%v", player.SnId, washedCoin, thirdID)
				}
				//五福红包游戏局数检测
				//actRandCoinMgr.OnPlayerGameTimes(player, int64(msg.GetAccRoundsInTime()))
			} else {
				if dbGamefreeInfo != nil {
					totalOut := int32(0)
					totalIn := int32(0)
					if msg.GetProfitCoinInTime() < 0 {
						totalIn = msg.GetBetCoinInTime()
					} else {
						totalOut = msg.GetBetCoinInTime()
					}
					QMFlowMgr.AddOffPlayerStatement(msg.GetSnid(), totalOut, totalIn, thirdID, dbGamefreeInfo)
					availableBet := int64(totalOut + totalIn)
					availableBet = availableBet * int64(dbGamefreeInfo.GetBetWaterRate()) / 100
					if availableBet > 0 {
						PlayerCacheMgrSington.Get(strconv.Itoa(int(msg.GetPlatform())), msg.GetSnid(), func(op *PlayerCacheItem, asyn, isnew bool) {
							if op == nil {
								return
							}
							if op.IsRob {
								return
							}
							//总流水累加
							op.TotalConvertibleFlow += availableBet
							op.TotalFlow += availableBet
							//今日流水增加,todo 这个地方没有考虑彩票导致的跨天，或者注单延迟导致的今日流水问题
							op.TodayGameData.TodayConvertibleFlow += availableBet
							task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
								//注意流控，防止该任务过渡占用登陆队列，可以在datasrv上配置心跳和maxdone来控制
								return model.UpdatePlayerExchageFlow(op.Platform, op.SnId, op.TotalConvertibleFlow, op.TotalFlow)
							}), nil).StartByExecutor(strconv.Itoa(int(op.SnId)))
						}, false)
					}
				}
			}
		}
	}

	return nil
}
func init() {
	netlib.RegisterHandler(int(server.SSPacketID_PACKET_DW_ThirdRebateMessage), &DWThirdRebateMessageHandler{})
	netlib.RegisterFactory(int(server.SSPacketID_PACKET_DW_ThirdRebateMessage), &DWThirdRebateMessagePacketFactory{})
	netlib.RegisterHandler(int(server.SSPacketID_PACKET_DW_ThirdRoundMessage), &DWThirdRoundMessageHandler{})
	netlib.RegisterFactory(int(server.SSPacketID_PACKET_DW_ThirdRoundMessage), &DWThirdRoundMessagePacketFactory{})
}

// 暂时约定为：result=1正常，result=-1错误
func SendAckToDataSrv(snid uint64, result int32) bool {
	datasrvSess := srvlib.ServerSessionMgrSington.GetSession(common.GetSelfAreaId(), common.DataServerType, common.DataServerId)
	if datasrvSess != nil {
		pack := &server.WDACKThirdRebateMessage{
			Tag:    proto.Uint64(snid),
			Result: proto.Int32(result),
		}
		proto.SetDefaults(pack)
		datasrvSess.Send(int(server.SSPacketID_PACKET_WD_ACKThirdRebateMessage), pack)
		return true
	} else {
		logger.Logger.Error("datasrv server not found.")
	}
	return false
}

//func init() {
//	go func() {
//		for {
//			time.Sleep(time.Second)
//			SendAckToDataSrv(205656,1)
//		}
//	}()
//}
