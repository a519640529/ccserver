package main

import (
	"games.yol.com/win88/protocol/tournament"
	"games.yol.com/win88/srvdata"
	"strconv"
	"strings"
	"time"

	"games.yol.com/win88/common"
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	gamehall_proto "games.yol.com/win88/protocol/gamehall"
	login_proto "games.yol.com/win88/protocol/login"
	player_proto "games.yol.com/win88/protocol/player"
	server_proto "games.yol.com/win88/protocol/server"
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/core/task"
)

func init() {
	//结算房卡
	netlib.RegisterFactory(int(server_proto.SSPacketID_PACKET_GW_BILLEDROOMCARD), netlib.PacketFactoryWrapper(func() interface{} {
		return &server_proto.GWBilledRoomCard{}
	}))
	netlib.RegisterHandler(int(server_proto.SSPacketID_PACKET_GW_BILLEDROOMCARD), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
		logger.Logger.Trace("receive GWBilledRoomCard:", pack)
		if msg, ok := pack.(*server_proto.GWBilledRoomCard); ok {
			scene := SceneMgrSington.GetScene(int(msg.GetRoomId()))
			if scene != nil {
				scene.BilledRoomCard(msg.GetSnId())
			}
		}
		return nil
	}))

	////游戏战绩
	//netlib.RegisterFactory(int(server_proto.SSPacketID_PACKET_GW_GAMEREC), netlib.PacketFactoryWrapper(func() interface{} {
	//	return &server_proto.GWGameRec{}
	//}))
	//netlib.RegisterHandler(int(server_proto.SSPacketID_PACKET_GW_GAMEREC), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
	//	logger.Logger.Trace("receive GWGameRec:", pack)
	//	if msg, ok := pack.(*server_proto.GWGameRec); ok {
	//		scene := SceneMgrSington.GetScene(int(msg.GetRoomId()))
	//		if scene != nil /*&& !scene.IsMatchScene() && !scene.IsCoinScene()*/ {
	//			var maxCoin int64
	//			var recDatas []model.PlayerGameRec
	//			gameTime := msg.GetGameTime()
	//			datas := msg.GetDatas()
	//			for _, d := range datas {
	//				p := scene.GetPlayer(d.GetId())
	//				if p != nil {
	//					recDatas = append(recDatas, model.PlayerGameRec{
	//						Id:          d.GetId(),
	//						Head:        d.GetHead(),
	//						Name:        d.GetName(),
	//						Coin:        d.GetCoin(),
	//						Pos:         d.GetPos(),
	//						OtherParams: d.GetOtherParams(),
	//					})
	//					if scene.replayCode != "" {
	//						GameLogChannelSington.Write(model.NewGameLog(p.SnId, int32(scene.gameId), int32(scene.mode), msg.GetReplayCode(), p.Platform, p.Channel, p.BeUnderAgentCode, int32(scene.sceneId)))
	//					}
	//					if d.GetCoin() > maxCoin {
	//						maxCoin = d.GetCoin()
	//					}
	//					p.dirty = true
	//					p.SendDiffData()
	//				}
	//			}
	//
	//			if scene.replayCode != "" {
	//				totalOfGames := scene.sp.GetTotalOfGames(scene)
	//				numOfGames := msg.GetNumOfGames()
	//				roomFeeMode := scene.sp.GetRoomFeeMode(scene)
	//				roomCardCnt := scene.sp.GetNeedRoomCardCntDependentPlayerCnt(scene)
	//				//agentor := scene.agentor
	//				//replayCode := scene.replayCode
	//				sceneId := int32(scene.sceneId)
	//				gameId := int32(scene.gameId)
	//				mode := int32(scene.mode)
	//				sceneType := int32(scene.sceneType)
	//				pos := int32(0) //scene.pos
	//				task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
	//					gr, err := model.InsertGameRec(msg.GetReplayCode(), sceneId, gameId, mode,
	//						sceneType, totalOfGames, numOfGames, roomFeeMode, roomCardCnt, recDatas, scene.params,
	//						gameTime, pos)
	//					if err == nil {
	//						//							if agentor != 0 && scene.replayCode != 0 {
	//						//								if !scene.inTeahourse {
	//						//									model.InsertAgentGameRec(agentor, msg.GetReplayCode())
	//						//								}
	//						//							}
	//						return gr
	//					}
	//					return nil
	//				}), task.CompleteNotifyWrapper(func(data interface{}, tt *task.Task) {
	//					return
	//				}), "InsertGameRec").StartByFixExecutor("logic_gamerec")
	//			}
	//		}
	//	}
	//	return nil
	//}))

	//销毁房间
	netlib.RegisterFactory(int(server_proto.SSPacketID_PACKET_GW_DESTROYSCENE), netlib.PacketFactoryWrapper(func() interface{} {
		return &server_proto.GWDestroyScene{}
	}))
	netlib.RegisterHandler(int(server_proto.SSPacketID_PACKET_GW_DESTROYSCENE), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
		logger.Logger.Trace("receive GWDestroyScene:", pack)
		if msg, ok := pack.(*server_proto.GWDestroyScene); ok {
			SceneMgrSington.DestroyScene(int(msg.GetSceneId()), msg.GetIsCompleted())
		}
		return nil
	}))

	//销毁小游戏房间
	netlib.RegisterFactory(int(server_proto.SSPacketID_PACKET_GW_DESTROYMINISCENE), netlib.PacketFactoryWrapper(func() interface{} {
		return &server_proto.GWDestroyMiniScene{}
	}))
	netlib.RegisterHandler(int(server_proto.SSPacketID_PACKET_GW_DESTROYMINISCENE), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
		logger.Logger.Trace("receive GWDestroyMiniScene:", pack)
		if msg, ok := pack.(*server_proto.GWDestroyMiniScene); ok {
			SceneMgrSington.DestroyMiniGameScene(int(msg.GetSceneId()))
		}
		return nil
	}))

	////玩家输赢统计
	//netlib.RegisterFactory(int(server_proto.SSPacketID_PACKET_GW_PLAYERSTATIC), netlib.PacketFactoryWrapper(func() interface{} {
	//	return &server_proto.GWPlayerStatics{}
	//}))
	//netlib.RegisterHandler(int(server_proto.SSPacketID_PACKET_GW_PLAYERSTATIC), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
	//	if msg, ok := pack.(*server_proto.GWPlayerStatics); ok {
	//		logger.Logger.Trace("receive GWPlayerStatics:", msg.GetRoomId())
	//		scene := SceneMgrSington.GetScene(int(msg.GetRoomId()))
	//		if scene != nil {
	//			datas := msg.GetDatas()
	//			key := scene.dbGameFree.GetGameDif()
	//			for _, data := range datas {
	//				snid := data.GetSnId()
	//				p := scene.GetPlayer(snid)
	//				if p != nil {
	//					if p.GameData == nil {
	//						p.GameData = make(map[string]*model.PlayerGameStatics)
	//					}
	//					if p.GameData != nil {
	//						if pgs, ok := p.GameData[key]; ok {
	//							pgs.GameTimes = data.GetGameTimes()
	//							pgs.WinGameTimes = data.GetWinGameTimes()
	//							pgs.LoseGameTimes = data.GetLoseGameTimes()
	//							pgs.TotalIn = data.GetTotalIn()
	//							pgs.TotalOut = data.GetTotalOut()
	//							pgs.TotalSysIn = data.GetTotalSysIn()
	//							pgs.TotalSysOut = data.GetTotalSysOut()
	//						} else {
	//							var pgs model.PlayerGameStatics
	//							pgs.GameTimes = data.GetGameTimes()
	//							pgs.WinGameTimes = data.GetWinGameTimes()
	//							pgs.LoseGameTimes = data.GetLoseGameTimes()
	//							pgs.TotalIn = data.GetTotalIn()
	//							pgs.TotalOut = data.GetTotalOut()
	//							pgs.TotalSysIn = data.GetTotalSysIn()
	//							pgs.TotalSysOut = data.GetTotalSysOut()
	//							p.GameData[key] = &pgs
	//						}
	//					}
	//					//新手状态同步
	//					if p.IsFoolPlayer == nil {
	//						p.IsFoolPlayer = make(map[string]bool)
	//					}
	//					if p.IsFoolPlayer[key] != data.GetIsFoolPlayer() {
	//						p.IsFoolPlayer[key] = data.GetIsFoolPlayer()
	//					}
	//				}
	//			}
	//			///////////////////////////////俱乐部抽水统计//////////////////////
	//			if msg.GetClubId() > 0 {
	//				//ClubInfoMgrSington.memoryClubTotal[msg.GetClubId()] += msg.GetPumpTotalCoin()
	//			}
	//		}
	//	}
	//	return nil
	//}))

	//离开房间
	netlib.RegisterFactory(int(server_proto.SSPacketID_PACKET_GW_PLAYERLEAVE), netlib.PacketFactoryWrapper(func() interface{} {
		return &server_proto.GWPlayerLeave{}
	}))
	netlib.RegisterHandler(int(server_proto.SSPacketID_PACKET_GW_PLAYERLEAVE), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
		if msg, ok := pack.(*server_proto.GWPlayerLeave); ok {
			logger.Logger.Trace("receive GWPlayerLeave:", msg.GetPlayerId())
			scene := SceneMgrSington.GetScene(int(msg.GetRoomId()))
			if scene != nil {
				p := PlayerMgrSington.GetPlayerBySnId(msg.GetPlayerId())
				if p != nil {
					data := msg.GetPlayerData()
					if len(data) != 0 {
						logger.Logger.Trace("GWPlayerLeave p.UnmarshalData(data)")
						p.UnmarshalData(data, scene)
					}
					if scene.IsCoinScene() { //金豆场
						if !CoinSceneMgrSington.PlayerLeave(p, int(msg.GetReason())) {
							logger.Logger.Warnf("GWPlayerLeave snid:%v sceneid:%v gameid:%v modeid:%v [coinscene]",
								p.SnId, scene.sceneId, scene.gameId, scene.gameMode)
						}
					} else if scene.IsHundredScene() { //百人场
						if !HundredSceneMgrSington.PlayerLeave(p, int(msg.GetReason())) {
							logger.Logger.Warnf("GWPlayerLeave snid:%v sceneid:%v gameid:%v modeid:%v [hundredcene]",
								p.SnId, scene.sceneId, scene.gameId, scene.gameMode)
						}
					} else if scene.IsMatchScene() {
						scene.PlayerLeave(p, int(msg.GetReason()))
						if msg.GetMatchStop() {
							////通知客户端比赛结束(主要是机器人取消标记)
							spack := &tournament.SCTMStop{}
							proto.SetDefaults(spack)
							logger.Logger.Trace("SCTMStop:", spack)
							p.SendToClient(int(tournament.TOURNAMENTID_PACKET_TM_SCTMStop), spack)
						}
						//结算积分
						TournamentMgr.UpdateMatchInfo(p, msg.MatchId, int32(msg.GetReturnCoin()), int32(msg.GetCurIsWin()))
					} else {
						if scene.ClubId > 0 {
							//if club, ok := clubManager.clubList[scene.ClubId]; ok {
							//	if cp, ok1 := club.memberList[p.SnId]; ok1 {
							//		cp.GameCount += msg.GetGameTimes()
							//		cp.DayCoin += msg.GetTotalConvertibleFlow() - p.TotalConvertibleFlow
							//	}
							//}
							//if !ClubSceneMgrSington.PlayerLeave(p, int(msg.GetReason())) {
							//	logger.Logger.Warnf("Club leave room msg snid:%v sceneid:%v gameid:%v modeid:%v [coinscene]",
							//		p.SnId, scene.sceneId, scene.gameId, scene.mode)
							//	scene.PlayerLeave(p, int(msg.GetReason()))
							//}
						} else {
							scene.PlayerLeave(p, int(msg.GetReason()))
						}
					}

					if p.scene != nil {
						logger.Logger.Warnf("after GWPlayerLeave found snid:%v sceneid:%v gameid:%v modeid:%v", p.SnId, p.scene.sceneId, p.scene.gameId, p.scene.gameMode)
					}

					//比赛场不处理下面的内容
					if !scene.IsMatchScene() {
						oldCoin := p.Coin
						//带回金币
						if p.Coin != msg.GetReturnCoin() {
							p.Coin = msg.GetReturnCoin()
							if p.Coin < 0 {
								p.Coin = 0
							}
							p.dirty = true
						}
						logger.Logger.Infof("SSPacketID_PACKET_GW_PLAYERLEAVE: snid:%v oldcoin:%v coin:%v", p.SnId, oldCoin, p.Coin)
						p.diffData.Coin = -1                 //强制更新金币
						p.diffData.TotalConvertibleFlow = -1 //强制更新流水
						p.SendDiffData()                     //只是把差异发给前端

						gameCoinTs := msg.GetGameCoinTs()
						if !p.IsRob && !scene.IsTestScene() {
							//同步背包数据
							diffItems := []*Item{}
							dbItemArr := srvdata.PBDB_GameItemMgr.Datas.Arr
							if dbItemArr != nil {
								items := msg.GetItems()
								if items != nil {
									for _, dbItem := range dbItemArr {
										if itemNum, exist := items[dbItem.Id]; exist {
											oldItem := BagMgrSington.GetBagItemById(p.SnId, dbItem.Id)
											diffNum := itemNum
											if oldItem != nil {
												diffNum = itemNum - oldItem.ItemNum
											}
											if diffNum != 0 {
												item := &Item{
													ItemId:  dbItem.Id,
													ItemNum: diffNum,
												}
												diffItems = append(diffItems, item)
											}
										}
									}
								}
							}
							if diffItems != nil && len(diffItems) != 0 {
								BagMgrSington.AddJybBagInfo(p, diffItems)
							}

							//对账点同步
							if p.GameCoinTs < gameCoinTs {
								p.GameCoinTs = gameCoinTs
								p.dirty = true
							}
							//破产统计
							if int(msg.GetReason()) == common.PlayerLeaveReason_Bekickout {
								if len(scene.paramsEx) > 0 {
									gameIdEx := scene.paramsEx[0]
									gps := PlatformMgrSington.GetGameConfig(scene.limitPlatform.IdStr, scene.paramsEx[0])
									if gps != nil {
										lowLimit := gps.DbGameFree.GetLowerThanKick()
										if lowLimit != 0 && p.Coin+p.SafeBoxCoin < int64(lowLimit) {
											p.ReportBankRuptcy(int32(scene.gameId), int32(scene.gameMode), gameIdEx)
										}
									}
								}
							}
						}
					}
				} else {
					LocalRobotIdMgrSington.FreeId(msg.GetPlayerId())
				}
			}
		}
		return nil
	}))

	//观众离开房间
	netlib.RegisterFactory(int(server_proto.SSPacketID_PACKET_GW_AUDIENCELEAVE), netlib.PacketFactoryWrapper(func() interface{} {
		return &server_proto.GWPlayerLeave{}
	}))
	netlib.RegisterHandler(int(server_proto.SSPacketID_PACKET_GW_AUDIENCELEAVE), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
		logger.Logger.Trace("receive PACKET_GW_AUDIENCELEAVE GWPlayerLeave:", pack)
		if msg, ok := pack.(*server_proto.GWPlayerLeave); ok {
			scene := SceneMgrSington.GetScene(int(msg.GetRoomId()))
			if scene != nil {
				p := PlayerMgrSington.GetPlayerBySnId(msg.GetPlayerId())
				if p != nil {
					gameCoinTs := msg.GetGameCoinTs()
					//带回的金币
					if p.Coin != msg.GetReturnCoin() && !scene.IsMatchScene() {
						p.Coin = msg.GetReturnCoin()
						if p.Coin < 0 {
							p.Coin = 0
						}
						p.dirty = true
					}

					//对账点同步
					if p.GameCoinTs != gameCoinTs {
						p.GameCoinTs = gameCoinTs
						p.dirty = true
					}

					if scene.IsCoinScene() { //金豆场
						if CoinSceneMgrSington.AudienceLeave(p, int(msg.GetReason())) {
						}
					} else {
						scene.AudienceLeave(p, int(msg.GetReason()))
					}

					//变化金币
					p.dirty = true
					p.diffData.Coin = -1                 //强制更新金币
					p.diffData.TotalConvertibleFlow = -1 //强制更新流水
					p.SendDiffData()
				} else {
					LocalRobotIdMgrSington.FreeId(msg.GetPlayerId())
				}
			}
		}
		return nil
	}))

	//房间游戏开始
	netlib.RegisterFactory(int(server_proto.SSPacketID_PACKET_GW_SCENESTART), netlib.PacketFactoryWrapper(func() interface{} {
		return &server_proto.GWSceneStart{}
	}))
	netlib.RegisterHandler(int(server_proto.SSPacketID_PACKET_GW_SCENESTART), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
		logger.Logger.Trace("receive SSPacketID_PACKET_GW_SCENESTART GWSceneStart:", pack)
		if msg, ok := pack.(*server_proto.GWSceneStart); ok {
			scene := SceneMgrSington.GetScene(int(msg.GetRoomId()))
			if scene != nil {
				scene.starting = msg.GetStart()
				scene.currRound = msg.GetCurrRound()
				scene.totalRound = msg.GetMaxRound()
				scene.lastTime = time.Now()
				if scene.starting {
					if scene.currRound == 1 {
						scene.startTime = time.Now()
						//p := PlayerMgrSington.GetPlayer(s.Sid)
					}
				}
				if scene.starting {
					logger.Logger.Trace("游戏开始------", scene.gameId, scene.sceneId, scene.replayCode, scene.currRound)
				} else {
					logger.Logger.Trace("游戏结束------", scene.gameId, scene.sceneId, scene.replayCode, scene.currRound)
				}
				PlatformMgrSington.OnChangeSceneState(scene, scene.starting)
			}
		}
		return nil
	}))

	//房间游戏状态
	netlib.RegisterFactory(int(server_proto.SSPacketID_PACKET_GW_SCENESTATE), netlib.PacketFactoryWrapper(func() interface{} {
		return &server_proto.GWSceneState{}
	}))
	netlib.RegisterHandler(int(server_proto.SSPacketID_PACKET_GW_SCENESTATE), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
		logger.Logger.Trace("receive SSPacketID_PACKET_GW_SCENESTATE GWSceneState:", pack)
		if msg, ok := pack.(*server_proto.GWSceneState); ok {
			scene := SceneMgrSington.GetScene(int(msg.GetRoomId()))
			if scene != nil {
				scene.state = msg.GetCurrState()
				scene.fishing = msg.GetFishing()
				PlatformMgrSington.OnChangeSceneState(scene, scene.starting)
			}
		}
		return nil
	}))

	//用户状态同步
	netlib.RegisterFactory(int(server_proto.SSPacketID_PACKET_GW_PLAYERSTATE), netlib.PacketFactoryWrapper(func() interface{} {
		return &server_proto.GWPlayerFlag{}
	}))
	netlib.RegisterHandler(int(server_proto.SSPacketID_PACKET_GW_PLAYERSTATE), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
		logger.Logger.Trace("receive GWPlayerFlag:", pack)
		if msg, ok := pack.(*server_proto.GWPlayerFlag); ok {
			player := PlayerMgrSington.GetPlayerBySnId(msg.GetSnId())
			if player != nil {
				player.flag = msg.GetFlag()
			}
		}
		return nil
	}))

	//房间服务器状态切换
	netlib.RegisterFactory(int(server_proto.SSPacketID_PACKET_GB_STATE_SWITCH), netlib.PacketFactoryWrapper(func() interface{} {
		return &server_proto.ServerStateSwitch{}
	}))

	netlib.RegisterHandler(int(server_proto.SSPacketID_PACKET_GB_STATE_SWITCH), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
		logger.Logger.Trace("receive SSPacketID_PACKET_GB_STATE_SWITCH ServerStateSwitch:", pack)
		if sr, ok := pack.(*server_proto.ServerStateSwitch); ok {
			srvid := int(sr.GetSrvId())
			gameSess := GameSessMgrSington.GetGameSess(srvid)
			if gameSess != nil {
				if gameSess.state == common.GAME_SESS_STATE_ON {
					gameSess.SwitchState(common.GAME_SESS_STATE_OFF)
				} else {
					gameSess.SwitchState(common.GAME_SESS_STATE_ON)
				}
			} else {
				gateSess := GameSessMgrSington.GetGateSess(srvid)
				if gateSess != nil {
					if gateSess.state == common.GAME_SESS_STATE_ON {
						gateSess.SwitchState(common.GAME_SESS_STATE_OFF)
					} else {
						gateSess.SwitchState(common.GAME_SESS_STATE_ON)
					}
				}
			}
		}
		return nil
	}))

	//游戏服务器的系统广播
	netlib.RegisterFactory(int(server_proto.SSPacketID_PACKET_GW_NEWNOTICE), netlib.PacketFactoryWrapper(func() interface{} {
		return &server_proto.GWNewNotice{}
	}))
	netlib.RegisterHandler(int(server_proto.SSPacketID_PACKET_GW_NEWNOTICE), netlib.HandlerWrapper(func(s *netlib.Session,
		packetid int, pack interface{}) error {
		logger.Logger.Trace("receive GWNewNotice:", pack)
		if msg, ok := pack.(*server_proto.GWNewNotice); ok {
			//立即发送改为定期发送，控制下广播包的频度
			HorseRaceLampMgrSington.PushGameHorseRaceLamp(msg.GetCh(), msg.GetPlatform(), msg.GetContent(), int32(msg.GetMsgtype()), msg.GetIsrob(), msg.GetPriority())
		}
		return nil
	}))

	//游戏服务器的解除黑白名单
	netlib.RegisterFactory(int(server_proto.SSPacketID_PACKET_GW_AUTORELIEVEWBLEVEL), netlib.PacketFactoryWrapper(func() interface{} {
		return &server_proto.GWAutoRelieveWBLevel{}
	}))
	netlib.RegisterHandler(int(server_proto.SSPacketID_PACKET_GW_AUTORELIEVEWBLEVEL), netlib.HandlerWrapper(func(s *netlib.Session,
		packetid int, pack interface{}) error {
		logger.Logger.Trace("receive GWAutoRelieveWBLevel:", pack)
		if msg, ok := pack.(*server_proto.GWAutoRelieveWBLevel); ok {
			player := PlayerMgrSington.GetPlayerBySnId(msg.GetSnId())
			if player != nil {
				player.WBLevel = 0
				player.WBCoinTotalIn = 0
				player.WBCoinTotalOut = 0
				player.WBCoinLimit = 0
				player.WBMaxNum = 0
				task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
					tNow := time.Now()
					return model.SetBlackWhiteLevel(player.SnId, 0, 0, 0, 0, 0, player.Platform, tNow)
				}), nil, "SetBlackWhiteLevel").StartByExecutor(strconv.Itoa(int(player.SnId)))
			}
		}
		return nil
	}))
	//来自Game的信息，房间里都是谁跟谁打牌的，配桌用的数据
	netlib.RegisterFactory(int(server_proto.SSPacketID_PACKET_GW_SCENEPLAYERLOG), netlib.PacketFactoryWrapper(func() interface{} {
		return &server_proto.GWScenePlayerLog{}
	}))
	netlib.RegisterHandler(int(server_proto.SSPacketID_PACKET_GW_SCENEPLAYERLOG), netlib.HandlerWrapper(func(s *netlib.Session,
		packetid int, pack interface{}) error {
		logger.Logger.Trace("receive GWScenePlayerLog:", pack)
		if msg, ok := pack.(*server_proto.GWScenePlayerLog); ok {
			sceneLimitMgr.ReciveData(msg.GetGameId(), msg.GetSnids())
		}
		return nil
	}))
	//强制离开房间
	netlib.RegisterFactory(int(server_proto.SSPacketID_PACKET_GW_PLAYERFORCELEAVE), netlib.PacketFactoryWrapper(func() interface{} {
		return &server_proto.GWPlayerForceLeave{}
	}))
	netlib.RegisterHandler(int(server_proto.SSPacketID_PACKET_GW_PLAYERFORCELEAVE), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
		if msg, ok := pack.(*server_proto.GWPlayerForceLeave); ok {
			logger.Logger.Warn("receive GWPlayerForceLeave:", msg)
			scene := SceneMgrSington.GetScene(int(msg.GetRoomId()))
			if scene != nil {
				p := PlayerMgrSington.GetPlayerBySnId(msg.GetPlayerId())
				if p != nil {
					ctx := scene.GetPlayerGameCtx(p.SnId)
					if ctx != nil && p.scene == scene && scene.HasPlayer(p) && ctx.enterTs == msg.GetEnterTs() {
						if scene.IsCoinScene() { //金豆场
							if !CoinSceneMgrSington.PlayerLeave(p, int(msg.GetReason())) {
								logger.Logger.Warnf("GWPlayerForceLeave snid:%v sceneid:%v gameid:%v modeid:%v [coinscene]", p.SnId, scene.sceneId, scene.gameId, scene.gameMode)
							}
						} else if scene.IsHundredScene() { //百人场
							if !HundredSceneMgrSington.PlayerLeave(p, int(msg.GetReason())) {
								logger.Logger.Warnf("GWPlayerForceLeave snid:%v sceneid:%v gameid:%v modeid:%v [hundredcene]", p.SnId, scene.sceneId, scene.gameId, scene.gameMode)
							}
						} else {
							scene.PlayerLeave(p, int(msg.GetReason()))
						}

						if p.scene != nil {
							logger.Logger.Warnf("after GWPlayerForceLeave found snid:%v sceneid:%v gameid:%v modeid:%v", p.SnId, p.scene.sceneId, p.scene.gameId, p.scene.gameMode)
							p.scene.PlayerLeave(p, int(msg.GetReason()))
						}
					} else {
						logger.Logger.Warnf("GWPlayerForceLeave snid:%v sceneid:%v gameid:%v modeid:%v fount not p.scene==scene && scene.HasPlayer(p)", p.SnId, scene.sceneId, scene.gameId, scene.gameMode)
					}
				}
			} else {
				logger.Logger.Warnf("GWPlayerForceLeave snid:%v scene:%v had closed", msg.GetPlayerId(), msg.GetRoomId())
			}
		}
		return nil
	}))

	//任务统计
	netlib.RegisterFactory(int(server_proto.SSPacketID_PACKET_GW_FISHRECORD), netlib.PacketFactoryWrapper(func() interface{} {
		return &server_proto.GWFishRecord{}
	}))
	netlib.RegisterHandler(int(server_proto.SSPacketID_PACKET_GW_FISHRECORD), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
		logger.Logger.Trace("receive GWFishRecord:", pack)
		//if msg, ok := pack.(*server_proto.GWFishRecord); ok {
		//	player := PlayerMgrSington.GetPlayerBySnId(msg.GetSnId())
		//	if player != nil && !player.IsRob {
		//		for _, v := range msg.FishRecords {
		//			for i := 0; i < 1; i++ {
		//				ActGoldTaskMgrSington.FireEvent(player, GoldTask_Sort_Fish, msg.GetGameFreeId(), []int64{int64(v.GetFishId()), int64(v.GetCount())})
		//			}
		//		}
		//	}
		//}
		return nil
	}))

	//一局游戏一结算
	//可处理以下业务
	//1.同步玩家游戏场内的实时金币
	//2.触发相关任务:统计有效下注金币数,赢取金币数,牌局次数等
	//3.黑名单处理
	netlib.RegisterFactory(int(server_proto.SSPacketID_PACKET_GW_PLAYERBET), netlib.PacketFactoryWrapper(func() interface{} {
		return &server_proto.GWPlayerBet{}
	}))
	netlib.RegisterHandler(int(server_proto.SSPacketID_PACKET_GW_PLAYERBET), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
		logger.Logger.Trace("receive GWPlayerBet:", pack)
		if msg, ok := pack.(*server_proto.GWPlayerBet); ok {
			scene := SceneMgrSington.GetScene(int(msg.GetSceneId()))
			for _, playerBet := range msg.PlayerBets {
				player := PlayerMgrSington.GetPlayerBySnId(playerBet.GetSnId())
				if player != nil && !player.IsRob {
					//当前金币
					if scene != nil && !scene.IsMiniGameScene() {
						player.Coin = playerBet.GetCoin()
						player.GameCoinTs = playerBet.GetGameCoinTs()
					}
					//流水
					flow := playerBet.GetFlowCoin()
					player.TotalConvertibleFlow += flow
					player.TotalFlow += flow
					player.TodayGameData.TodayConvertibleFlow += flow
					//税收
					player.GameTax += playerBet.GetTax()
					//输赢
					gain := playerBet.GetGain()
					//if gain > 0 {
					//	player.WinTimes++
					//} else if gain < 0 {
					//	player.FailTimes++
					//} else {
					//	player.DrawTimes++
					//}
					player.dirty = true
					player.OnPlayerGameGain(gain)

					//var washingCnt int64
					//if gain > 0 {
					//	totalCoinInOut += gain
					//	washingCnt = player.WashingCoin(gain)
					//} else {
					//	totalCoinInOut += (-gain)
					//	washingCnt = player.WashingCoin(-gain)
					//}
					//totalCoinWashing += washingCnt
					//totalProfit += washingCnt * player.WashingCoinConvRate / 10000
					//for i := 0; i < 1; i++ {
					//	ActGoldComeMgrSington.FireEvent(player, GoldCome_Sort_BetScore, msg.GetGameFreeId(), playerBet.GetBet(), 1)
					//	ActLuckyTurntableMgrSington.HandleBetScore(player, playerBet.GetBet(), msg.GetGameFreeId())
					//}
				}
			}

			//汇总游戏杀率相关数据
			//ProfitControlMgrSington.OnGameBill(platform, msg.GetGameFreeId(), totalCoinWashing, totalCoinInOut, totalTax, msg.GetRobotGain(), totalProfit)
		}
		return nil
	}))

	//房间游戏状态
	//netlib.RegisterFactory(int(server_proto.SSPacketID_PACKET_GW_SYNCPLAYERCOIN), netlib.PacketFactoryWrapper(func() interface{} {
	//	return &server_proto.GWSyncPlayerCoin{}
	//}))
	//netlib.RegisterHandler(int(server_proto.SSPacketID_PACKET_GW_SYNCPLAYERCOIN), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
	//	logger.Logger.Trace("receive SSPacketID_PACKET_GW_SYNCPLAYERCOIN GWSceneState:", pack)
	//	if msg, ok := pack.(*server_proto.GWSyncPlayerCoin); ok {
	//		scene := SceneMgrSington.GetScene(int(msg.GetSceneId()))
	//		if scene != nil && !scene.IsMatchScene() {
	//			coinsData := msg.GetPlayerCoins()
	//			if len(coinsData) != 0 && len(coinsData)%2 == 0 { //同步用户的实时金币量
	//				for i := 0; i < len(coinsData)/2; i = i + 2 {
	//					snid := int32(coinsData[i])
	//					coin := coinsData[i+1]
	//					if player, ok := scene.players[snid]; ok {
	//						player.sceneCoin = coin
	//					}
	//				}
	//			}
	//		}
	//	}
	//	return nil
	//}))

	//推送游戏的录单
	netlib.RegisterFactory(int(server_proto.SSPacketID_PACKET_GW_GAMESTATELOG), netlib.PacketFactoryWrapper(func() interface{} {
		return &server_proto.GWGameStateLog{}
	}))
	netlib.RegisterHandler(int(server_proto.SSPacketID_PACKET_GW_GAMESTATELOG), netlib.HandlerWrapper(func(s *netlib.Session,
		packetid int, pack interface{}) error {
		logger.Logger.Trace("receive SSPacketID_PACKET_GW_GAMESTATELOG GWDTRoomInfo:", pack)
		if msg, ok := pack.(*server_proto.GWGameStateLog); ok {
			scene := SceneMgrSington.GetScene(int(msg.GetSceneId()))
			if scene != nil {
				scene.GameLog = append(scene.GameLog, msg.GetGameLog())
				if msg.GetLogCnt() > 0 {
					remainCnt := int(msg.GetLogCnt())
					if len(scene.GameLog) > remainCnt {
						scene.GameLog = scene.GameLog[len(scene.GameLog)-remainCnt:]
					}
				} else {
					if len(scene.GameLog) > int(scene.sp.GetViewLogLen()) {
						scene.GameLog = scene.GameLog[1:]
					}
				}
				pack := &gamehall_proto.SCGameSubList{
					List: []*gamehall_proto.GameSubRecord{
						{
							GameFreeId: proto.Int32(scene.dbGameFree.GetId()),
							NewLog:     proto.Int32(msg.GetGameLog()),
							LogCnt:     proto.Int(len(scene.GameLog)),
						},
					},
				}
				gameStateMgr.BrodcastGameState(int32(scene.gameId), scene.limitPlatform.IdStr,
					int(gamehall_proto.GameHallPacketID_PACKET_SC_GAMESUBLIST), pack)
				logger.Logger.Trace("SCGameSubList:", pack)
			}

		}
		return nil
	}))
	//推送游戏的状态
	netlib.RegisterFactory(int(server_proto.SSPacketID_PACKET_GW_GAMESTATE), netlib.PacketFactoryWrapper(func() interface{} {
		return &server_proto.GWGameState{}
	}))
	netlib.RegisterHandler(int(server_proto.SSPacketID_PACKET_GW_GAMESTATE), netlib.HandlerWrapper(func(s *netlib.Session,
		packetid int, pack interface{}) error {
		logger.Logger.Trace("receive SSPacketID_PACKET_GW_GAMESTATE GWGameState:", pack)
		if msg, ok := pack.(*server_proto.GWGameState); ok {
			scene := SceneMgrSington.GetScene(int(msg.GetSceneId()))
			if scene != nil {
				scene.State = msg.GetState()
				scene.StateSec = msg.GetSec()
				scene.BankerListNum = msg.GetBankerListNum()
				if scene.State == scene.sp.GetBetState() {
					scene.StateTs = msg.GetTs()
					leftTime := int64(scene.StateSec) - (time.Now().Unix() - scene.StateTs)
					if leftTime < 0 {
						leftTime = 0
					}
					pack := &gamehall_proto.SCGameState{}
					pack.List = append(pack.List, &gamehall_proto.GameState{
						GameFreeId: proto.Int32(scene.dbGameFree.GetId()),
						Ts:         proto.Int64(leftTime),
						Sec:        proto.Int32(scene.StateSec),
					})
					gameStateMgr.BrodcastGameState(int32(scene.gameId), scene.limitPlatform.IdStr,
						int(gamehall_proto.GameHallPacketID_PACKET_SC_GAMESTATE), pack)
				}
			}
		}
		return nil
	}))
	netlib.RegisterFactory(int(server_proto.SSPacketID_PACKET_GW_JACKPOTLIST), netlib.PacketFactoryWrapper(func() interface{} {
		return &server_proto.GWGameJackList{}
	}))
	netlib.RegisterHandler(int(server_proto.SSPacketID_PACKET_GW_JACKPOTLIST), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
		logger.Logger.Trace("receive SSPacketID_PACKET_GW_JACKPOTLIST GWGameJackList:", pack)
		if msg, ok := pack.(*server_proto.GWGameJackList); ok {
			FishJackListMgr.Insert(msg.GetCoin(), msg.GetSnId(), msg.GetRoomId(), msg.GetJackType(), msg.GetGameId(),
				msg.GetPlatform(), msg.GetChannel(), msg.GetName())
		}
		return nil
	}))
	netlib.RegisterFactory(int(server_proto.SSPacketID_PACKET_GW_JACKPOTCOIN), netlib.PacketFactoryWrapper(func() interface{} {
		return &server_proto.GWGameJackCoin{}
	}))
	netlib.RegisterHandler(int(server_proto.SSPacketID_PACKET_GW_JACKPOTCOIN), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
		logger.Logger.Trace("receive SSPacketID_PACKET_GW_JACKPOTCOIN GWGameJackCoin:", pack)
		if msg, ok := pack.(*server_proto.GWGameJackCoin); ok {
			for i, pl := range msg.Platform {
				FishJackpotCoinMgr.Jackpot[pl] = msg.Coin[i]
			}
		}
		return nil
	}))

	//自动标签
	netlib.RegisterFactory(int(server_proto.SSPacketID_PACKET_GW_PLAYERAUTOMARKTAG), netlib.PacketFactoryWrapper(func() interface{} {
		return &server_proto.GWPlayerAutoMarkTag{}
	}))
	netlib.RegisterHandler(int(server_proto.SSPacketID_PACKET_GW_PLAYERAUTOMARKTAG), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
		logger.Logger.Trace("receive SSPacketID_PACKET_GW_PLAYERAUTOMARKTAG GWPlayerAutoMarkTag:", pack)
		if msg, ok := pack.(*server_proto.GWPlayerAutoMarkTag); ok {
			p := PlayerMgrSington.GetPlayerBySnId(msg.GetSnId())
			if p != nil {
				p.MarkAutoTag(msg.GetTag())
				p.dirty = true
			}
		}
		return nil
	}))
	//强制换桌
	netlib.RegisterFactory(int(server_proto.SSPacketID_PACKET_GW_CHANGESCENEEVENT), netlib.PacketFactoryWrapper(func() interface{} {
		return &server_proto.GWChangeSceneEvent{}
	}))
	netlib.RegisterHandler(int(server_proto.SSPacketID_PACKET_GW_CHANGESCENEEVENT), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
		logger.Logger.Trace("receive SSPacketID_PACKET_GW_CHANGESCENEEVENT GWChangeSceneEvent:", pack)
		if msg, ok := pack.(*server_proto.GWChangeSceneEvent); ok {
			scene := SceneMgrSington.GetScene(int(msg.GetSceneId()))
			if scene != nil {
				scene.PlayerTryChange()
			}
		}
		return nil
	}))

	//玩家比赛比分
	netlib.RegisterFactory(int(server_proto.SSPacketID_PACKET_GW_PLAYERMATCHGRADE), netlib.PacketFactoryWrapper(func() interface{} {
		return &server_proto.GWPlayerMatchGrade{}
	}))
	netlib.RegisterHandler(int(server_proto.SSPacketID_PACKET_GW_PLAYERMATCHGRADE), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
		logger.Logger.Trace("receive GWPlayerMatchGrade:", pack)
		//if msg, ok := pack.(*server_proto.GWPlayerMatchGrade); ok {
		//sceneId := msg.GetSceneId()
		//matchId := msg.GetMatchId()
		//numOfGame := msg.GetNumOfGame()
		//gamelogId := msg.GetGameLogId()
		//spendTime := msg.GetSpendTime()
		//players := msg.GetPlayers()
		//sort.Sort(MatchCoinSlice(players))
		//m := MatchMgrSington.GetCopyMatchByScene(int(sceneId))
		//if m != nil {
		//	for i, bill := range players {
		//		mc := m.GetPlayerMatchContect(bill.GetSnId())
		//		if mc != nil {
		//			mc.grade = bill.GetCoin()
		//			mc.sceneRank = int32(i + 1)
		//		}
		//	}
		//	MatchMgrSington.OnGameBilled(matchId, sceneId, gamelogId, numOfGame, spendTime)
		//}
		//}
		return nil
	}))

	//玩家比赛结算
	netlib.RegisterFactory(int(server_proto.SSPacketID_PACKET_GW_PLAYERMATCHBILLED), netlib.PacketFactoryWrapper(func() interface{} {
		return &server_proto.GWPlayerMatchBilled{}
	}))
	netlib.RegisterHandler(int(server_proto.SSPacketID_PACKET_GW_PLAYERMATCHBILLED), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
		//logger.Logger.Trace("receive GWPlayerMatchBilled:", pack)
		//if msg, ok := pack.(*server_proto.GWPlayerMatchBilled); ok {
		//	sceneId := msg.GetSceneId()
		//	matchId := msg.GetMatchId()
		//	winPos := msg.GetWinPos()
		//	players := msg.GetPlayers()
		//	sort.Sort(MatchCoinSlice(players))
		//	logger.Logger.Trace("receive GWPlayerMatchBilled:winpos=", winPos)
		//	m := MatchMgrSington.GetCopyMatchByScene(int(sceneId))
		//	if m != nil {
		//		for i, bill := range players {
		//			mc := m.GetPlayerMatchContect(bill.GetSnId())
		//			if mc != nil {
		//				mc.grade = bill.GetCoin()
		//				mc.sceneRank = int32(i + 1)
		//			}
		//		}
		//	}
		//	MatchMgrSington.OnSceneBilled(matchId, sceneId)
		//}
		return nil
	}))

	//玩家中转消息
	netlib.RegisterFactory(int(server_proto.SSPacketID_PACKET_SS_REDIRECTTOPLAYER), netlib.PacketFactoryWrapper(func() interface{} {
		return &server_proto.SSRedirectToPlayer{}
	}))
	netlib.RegisterHandler(int(server_proto.SSPacketID_PACKET_SS_REDIRECTTOPLAYER), netlib.HandlerWrapper(func(s *netlib.Session,
		packetid int, pack interface{}) error {
		logger.Logger.Trace("SSRedirectToPlayer Process recv ", pack)
		if msg, ok := pack.(*server_proto.SSRedirectToPlayer); ok {
			p := PlayerMgrSington.GetPlayerBySnId(msg.GetSnId())
			if p == nil {
				return nil
			}
			p.SendToClient(int(msg.GetPacketId()), msg.GetData())
		}
		return nil
	}))

	//// 奖池信息
	//netlib.RegisterFactory(int(server_proto.SSPacketID_PACKET_GW_GAMEJACKPOT), netlib.PacketFactoryWrapper(func() interface{} {
	//	return &server_proto.GWGameJackpot{}
	//}))
	//netlib.RegisterHandler(int(server_proto.SSPacketID_PACKET_GW_GAMEJACKPOT), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
	//	logger.Logger.Trace("GWGameJackpot Process recv ", pack)
	//	if msg, ok := pack.(*server_proto.GWGameJackpot); ok {
	//		scene := SceneMgrSington.GetScene(int(msg.GetSceneId()))
	//		if scene != nil {
	//			scene.JackPotFund = msg.GetJackpotFund()
	//			gameid := int(scene.dbGameFree.GetGameId())
	//			// 冰河世纪, 百战成神, 财神, 复仇者联盟, 复活岛 主动推送奖池变化信息
	//			if gameid == common.GameId_IceAge || gameid == common.GameId_TamQuoc || gameid == common.GameId_CaiShen ||
	//				gameid == common.GameId_Avengers || gameid == common.GameId_EasterIsland {
	//				jackpotMsg := &gamehall_proto.SCHundredSceneGetGameJackpot{}
	//				jackpotMsg.GameJackpotFund = append(jackpotMsg.GameJackpotFund, &gamehall_proto.GameJackpotFundInfo{
	//					GameFreeId:  proto.Int32(scene.dbGameFree.GetId()),
	//					JackPotFund: proto.Int64(scene.JackPotFund),
	//				})
	//				proto.SetDefaults(jackpotMsg)
	//				gameStateMgr.BrodcastGameState(int32(scene.gameId), scene.limitPlatform.IdStr,
	//					int(gamehall_proto.HundredScenePacketID_PACKET_SC_GAMEJACKPOT), jackpotMsg)
	//				logger.Logger.Trace("SCGameJackpot:", jackpotMsg)
	//			}
	//		}
	//	}
	//	return nil
	//}))

	// 爆奖信息
	netlib.RegisterFactory(int(server_proto.SSPacketID_PACKET_GW_GAMENEWBIGWINHISTORY), netlib.PacketFactoryWrapper(func() interface{} {
		return &server_proto.GWGameNewBigWinHistory{}
	}))
	netlib.RegisterHandler(int(server_proto.SSPacketID_PACKET_GW_GAMENEWBIGWINHISTORY), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
		logger.Logger.Trace("GWGameNewBigWinHistory Process recv ", pack)
		if msg, ok := pack.(*server_proto.GWGameNewBigWinHistory); ok {
			scene := SceneMgrSington.GetScene(int(msg.GetSceneId()))
			if scene != nil {
				gameMsg := &gamehall_proto.BigWinHistoryInfo{
					SpinID:      msg.GetBigWinHistory().SpinID,
					CreatedTime: msg.GetBigWinHistory().CreatedTime,
					BaseBet:     msg.GetBigWinHistory().BaseBet,
					TotalBet:    msg.GetBigWinHistory().TotalBet,
					PriceValue:  msg.GetBigWinHistory().PriceValue,
					UserName:    msg.GetBigWinHistory().UserName,
					Cards:       msg.GetBigWinHistory().Cards,
				}
				if msg.GetBigWinHistory().GetIsVirtualData() {
					JackpotListMgrSington.AddVirtualJackpot(scene.gameId, gameMsg)
				} else {
					JackpotListMgrSington.AddJackpotList(scene.gameId, gameMsg)
				}
				// 重置定时器
				//JackpotListMgrSington.ResetAfterTimer(scene.gameId)
			}
		}
		return nil
	}))
}

// 机器人服务器向worldsrv发送
type CSPMCmdPacketFactory struct {
}
type CSPMCmdHandler struct {
}

func (this *CSPMCmdPacketFactory) CreatePacket() interface{} {
	pack := &player_proto.CSPMCmd{}
	return pack
}

func (this *CSPMCmdHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSPMCmdHandler Process recv ", data)
	if msg, ok := data.(*player_proto.CSPMCmd); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Trace("CSPMCmdHandler p == nil")
			return nil
		}

		if !p.IsRob && p.GMLevel < 3 {
			logger.Logger.Trace("CSPMCmdHandler p.Channel != common.Channel_Rob && p.GMLevel < 3")
			return nil
		}

		cmd := msg.GetCmd()
		logger.Logger.Infof("CSPMCmdHandler %v %v", p.SnId, cmd)
		args := strings.Split(cmd, common.PMCmd_SplitToken)
		argsCnt := len(args)
		if argsCnt != 0 {
			switch args[0] {
			case common.PMCmd_AddCoin:
				if argsCnt > 1 {
					coin, err := strconv.ParseInt(args[1], 10, 64) //strconv.Atoi(args[1])
					if err != nil {
						logger.Logger.Warnf("CSPMCmdHandler %v parse %v err:%v", p.SnId, cmd, err)
						return nil
					}
					if coin != 0 {
						p.CoinPayTotal += int64(coin)
						p.dirty = true
						p.AddCoin(coin, common.GainWay_ByPMCmd, p.GetName(), cmd)
						p.ReportSystemGiveEvent(int32(coin), common.GainWay_ByPMCmd, true)
						p.SendDiffData()
					}
				}
			case common.PMCmd_Privilege:
				if p.GMLevel >= 3 {
					if p.Flags&model.PLAYER_FLAGS_PRIVILEGE == 0 {
						p.Flags |= model.PLAYER_FLAGS_PRIVILEGE
					} else {
						p.Flags &= ^model.PLAYER_FLAGS_PRIVILEGE
					}
				}
			}
		}
	}
	return nil
}

type CSRobotChgDataPacketFactory struct {
}
type CSRobotChgDataHandler struct {
}

func (this *CSRobotChgDataPacketFactory) CreatePacket() interface{} {
	pack := &player_proto.CSRobotChgData{}
	return pack
}

func (this *CSRobotChgDataHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSRobotChgDataHandler Process recv ", data)
	if _, ok := data.(*player_proto.CSRobotChgData); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Trace("CSRobotChgDataHandler p == nil")
			return nil
		}

		if !p.IsRob {
			logger.Logger.Trace("CSRobotChgDataHandler !p.IsRob")
			return nil
		}
	}
	return nil
}

type CSAccountInvalidPacketFactory struct {
}
type CSAccountInvalidHandler struct {
}

func (this *CSAccountInvalidPacketFactory) CreatePacket() interface{} {
	pack := &login_proto.CSAccountInvalid{}
	return pack
}

func (this *CSAccountInvalidHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSAccountInvalidHandler Process recv ", data)
	if _, ok := data.(*login_proto.CSAccountInvalid); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p != nil && p.IsRobot() {
			snid := p.SnId
			acc := p.AccountId
			task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
				err := model.RemoveAccount(p.Platform, acc)
				if err != nil {
					logger.Logger.Error("Remove robot account data error:", err)
				}
				err = model.RemovePlayerByAcc(p.Platform, acc)
				if err != nil {
					logger.Logger.Error("Remove robot player data error:", err)
				}
				logger.Logger.Trace("CSAccountInvalid message remove :", acc)
				return nil
			}), task.CompleteNotifyWrapper(func(data interface{}, tt task.Task) {
				PlayerMgrSington.DelPlayer(snid)
				LoginStateMgrSington.DelAccountByAccid(acc)
				return
			}), "RemoveAccount").StartByFixExecutor("RemoveAccount")
		}
	}
	return nil
}

type GWAddSingleAdjustPacketFactory struct {
}
type GWAddSingleAdjustHandler struct {
}

func (this *GWAddSingleAdjustPacketFactory) CreatePacket() interface{} {
	pack := &server_proto.GWAddSingleAdjust{}
	return pack
}

func (this *GWAddSingleAdjustHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("GWAddSingleAdjustHandler Process recv ", data)
	if msg, ok := data.(*server_proto.GWAddSingleAdjust); ok {
		PlayerSingleAdjustMgr.AddAdjustCount(msg.SnId, msg.GameFreeId)
	}
	return nil
}
func init() {
	common.RegisterHandler(int(player_proto.PlayerPacketID_PACKET_CS_PMCMD), &CSPMCmdHandler{})
	netlib.RegisterFactory(int(player_proto.PlayerPacketID_PACKET_CS_PMCMD), &CSPMCmdPacketFactory{})

	common.RegisterHandler(int(player_proto.PlayerPacketID_PACKET_CS_ROBOTCHGDATA), &CSRobotChgDataHandler{})
	netlib.RegisterFactory(int(player_proto.PlayerPacketID_PACKET_CS_ROBOTCHGDATA), &CSRobotChgDataPacketFactory{})

	common.RegisterHandler(int(login_proto.LoginPacketID_PACKET_CS_ACCOUNTINVALID), &CSAccountInvalidHandler{})
	netlib.RegisterFactory(int(login_proto.LoginPacketID_PACKET_CS_ACCOUNTINVALID), &CSAccountInvalidPacketFactory{})

	common.RegisterHandler(int(server_proto.SSPacketID_PACKET_GW_ADDSINGLEADJUST), &GWAddSingleAdjustHandler{})
	netlib.RegisterFactory(int(server_proto.SSPacketID_PACKET_GW_ADDSINGLEADJUST), &GWAddSingleAdjustPacketFactory{})
}
