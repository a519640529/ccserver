package action

import (
	"encoding/json"
	"games.yol.com/win88/common"
	"games.yol.com/win88/gamesrv/base"
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/gamehall"
	//"games.yol.com/win88/protocol/match"
	"games.yol.com/win88/protocol/server"
	"games.yol.com/win88/protocol/webapi"
	"games.yol.com/win88/srvdata"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/srvlib"
)

func init() {
	//创建场景
	netlib.RegisterFactory(int(server.SSPacketID_PACKET_WG_CREATESCENE), netlib.PacketFactoryWrapper(func() interface{} {
		return &server.WGCreateScene{}
	}))
	netlib.RegisterHandler(int(server.SSPacketID_PACKET_WG_CREATESCENE), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
		logger.Logger.Trace("receive WGCreateScene:", pack)
		if msg, ok := pack.(*server.WGCreateScene); ok {
			sceneId := int(msg.GetSceneId())
			gameMode := int(msg.GetGameMode())
			sceneMode := int(msg.GetSceneMode())
			gameId := int(msg.GetGameId())
			paramsEx := msg.GetParamsEx()
			hallId := msg.GetHallId()
			groupId := msg.GetGroupId()
			dbGameFree := msg.GetDBGameFree()
			bEnterAfterStart := msg.GetEnterAfterStart()
			totalOfGames := msg.GetTotalOfGames()
			baseScore := msg.GetBaseScore()
			playerNum := int(msg.GetPlayerNum())
			scene := base.SceneMgrSington.CreateScene(s, sceneId, gameMode, sceneMode, gameId, msg.GetPlatform(), msg.GetParams(),
				msg.GetAgentor(), msg.GetCreator(), msg.GetReplayCode(), hallId, groupId, totalOfGames, dbGameFree,
				bEnterAfterStart, baseScore, playerNum, paramsEx...)
			if scene != nil {
				if scene.IsMatchScene() {
					if len(scene.Params) > 0 {
						scene.MatchId = scene.Params[0]
					}
					if len(scene.Params) > 1 {
						scene.MatchFinals = scene.Params[1] == 1
					}
					if len(scene.Params) > 2 {
						scene.MatchRound = scene.Params[2]
					}
					if len(scene.Params) > 3 {
						scene.MatchCurPlayerNum = scene.Params[3]
					}
					if len(scene.Params) > 4 {
						scene.MatchNextNeed = scene.Params[4]
					}
					if len(scene.Params) > 5 {
						scene.MatchType = scene.Params[5]
					}
				}
				scene.ClubId = msg.GetClub()
				scene.RoomId = msg.GetClubRoomId()
				scene.RoomPos = msg.GetClubRoomPos()
				scene.PumpCoin = msg.GetClubRate()
			}
		}
		return nil
	}))
	//删除场景
	netlib.RegisterFactory(int(server.SSPacketID_PACKET_WG_DESTROYSCENE), netlib.PacketFactoryWrapper(func() interface{} {
		return &server.WGDestroyScene{}
	}))
	netlib.RegisterHandler(int(server.SSPacketID_PACKET_WG_DESTROYSCENE), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
		logger.Logger.Trace("receive WGDestroyScene:", pack)
		if msg, ok := pack.(*server.WGDestroyScene); ok {
			sceneId := int(msg.GetSceneId())
			s := base.SceneMgrSington.GetScene(sceneId)
			if s != nil {
				if gameScene, ok := s.ExtraData.(base.GameScene); ok {
					s.MatchStop = msg.GetMatchStop()
					gameScene.SceneDestroy(true)
				}
			}
		}
		return nil
	}))

	//删除场景
	netlib.RegisterFactory(int(server.SSPacketID_PACKET_WG_GRACE_DESTROYSCENE), netlib.PacketFactoryWrapper(func() interface{} {
		return &server.WGGraceDestroyScene{}
	}))
	netlib.RegisterHandler(int(server.SSPacketID_PACKET_WG_GRACE_DESTROYSCENE), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
		logger.Logger.Trace("receive WGGraceDestroyScene:", pack)
		if msg, ok := pack.(*server.WGGraceDestroyScene); ok {
			ids := msg.GetIds()
			for _, id := range ids {
				s := base.SceneMgrSington.GetScene(int(id))
				if s != nil {
					if s.IsHundredScene() || s.Gaming {
						s.SetGraceDestroy(true)
					} else {
						if s.IsMatchScene() {
							s.SetGraceDestroy(true)
						}
						if gameScene, ok := s.ExtraData.(base.GameScene); ok {
							gameScene.SceneDestroy(true)
						}
					}
				}
			}
		}
		return nil
	}))
	//玩家进入
	netlib.RegisterFactory(int(server.SSPacketID_PACKET_WG_PLAYERENTER), netlib.PacketFactoryWrapper(func() interface{} {
		return &server.WGPlayerEnter{}
	}))
	netlib.RegisterHandler(int(server.SSPacketID_PACKET_WG_PLAYERENTER), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
		logger.Logger.Trace("receive WGPlayerEnter")
		if msg, ok := pack.(*server.WGPlayerEnter); ok {
			sceneId := int(msg.GetSceneId())
			sid := msg.GetSid()
			data := msg.GetPlayerData()
			gateSid := msg.GetGateSid()
			isload := msg.GetIsLoaded()
			IsQM := msg.GetIsQM()
			sendLeave := func(reason int) {
				pack := &server.GWPlayerLeave{
					RoomId:     msg.SceneId,
					PlayerId:   msg.SnId,
					ReturnCoin: msg.TakeCoin,
					Reason:     proto.Int(reason),
				}
				proto.SetDefaults(pack)
				s.Send(int(server.SSPacketID_PACKET_GW_AUDIENCELEAVE), pack)
			}
			scene := base.SceneMgrSington.GetScene(sceneId)
			p := base.NewPlayer(sid, data, nil, nil)
			if p == nil && !scene.IsMatchScene() {
				sendLeave(common.PlayerLeaveReason_OnDestroy)
				return nil
			}
			p.UnmarshalIParam(msg.GetIParams())
			p.UnmarshalSParam(msg.GetSParams())
			p.UnmarshalCParam(msg.GetCParams())
			p.AgentCode = msg.GetAgentCode()
			p.Coin = msg.GetTakeCoin()
			p.Pos = int(msg.GetPos())
			p.MatchParams = msg.GetMatchParams()
			items := msg.GetItems()
			if items != nil {
				p.Items = make(map[int32]int32)
				for id, num := range items {
					p.Items[id] = num
				}
			}
			p.SetTakeCoin(msg.GetTakeCoin())
			//p.StartCoin = msg.GetTakeCoin()
			//机器人用
			p.ExpectGameTime = msg.GetExpectGameTimes()
			p.ExpectLeaveCoin = msg.GetExpectLeaveCoin()
			//当局游戏结束后剩余金额 起始设置
			p.SetCurrentCoin(msg.GetTakeCoin())

			p.LastSyncCoin = p.Coin
			p.IsQM = IsQM
			p.UnMarshalSingleAdjustData(msg.SingleAdjust)

			if (sid == 0 || scene == nil) && !scene.IsMatchScene() {
				if scene == nil {
					logger.Logger.Warn("when WGPlayerEnter (scene == nil)")
				}

				if sid == 0 {
					logger.Logger.Warnf("when WGPlayerEnter (sid == 0)")
				}

				//进入房间失败
				sendLeave(common.PlayerLeaveReason_OnDestroy)
				return nil
			}

			isQuit := p.GetIParam(common.PlayerIParam_IsQuit)
			logger.Logger.Tracef("WGPlayerEnter scene.IsMatchScene()=%v p.GetIParam(common.PlayerIParam_IsQuit)=%v", scene.IsMatchScene(), isQuit)
			if scene.IsMatchScene() && isQuit == 1 { //比赛场退赛
				p.MarkFlag(base.PlayerState_MatchQuit)
				p.MarkFlag(base.PlayerState_Auto)
				p.MarkFlag(base.PlayerState_Leave)
			}

			var sessionId srvlib.SessionId
			sessionId.Set(gateSid)
			gateSess := srvlib.ServerSessionMgrSington.GetSession(int(sessionId.AreaId()), int(sessionId.SrvType()), int(sessionId.SrvId()))
			logger.Logger.Tracef("WGPlayerEnter, AreaId=%v, SrvType=%v, SrvId=%v, GateSess=%v", int(sessionId.AreaId()), int(sessionId.SrvType()), int(sessionId.SrvId()), gateSess)
			if gateSess == nil && !scene.IsMatchScene() {
				logger.Logger.Warnf("WGPlayerEnter, AreaId=%v, SrvType=%v, SrvId=%v, GateSess=<nil>", int(sessionId.AreaId()), int(sessionId.SrvType()), int(sessionId.SrvId()))
				//进入房间失败
				sendLeave(common.PlayerLeaveReason_OnDestroy)
				return nil
			}

			p.SetGateSess(gateSess)
			p.SetWorldSess(s)

			if gateSess != nil {
				pack := &server.GGPlayerSessionBind{
					Sid: proto.Int64(sid),
				}
				if !p.IsRob {
					pack.SnId = proto.Int32(p.SnId)
					pack.Vip = proto.Int32(p.VIP)
					pack.CoinPayTotal = proto.Int64(p.CoinPayTotal)
					pack.Ip = proto.String(p.Ip)
					pack.Platform = proto.String(p.Platform)
				}
				proto.SetDefaults(pack)
				gateSess.Send(int(server.SSPacketID_PACKET_GG_PLAYERSESSIONBIND), pack)
			}

			if scene.Testing {
				p.Coin = int64(scene.DbGameFree.GetTestTakeCoin())
			}
			base.PlayerMgrSington.ManagePlayer(p)
			scene.PlayerEnter(p, isload)
			//进场失败
			if p.IsMarkFlag(base.PlayerState_EnterSceneFailed) {
				scene.PlayerLeave(p, common.PlayerLeaveReason_Normal, true)
			} else {
				// 进入成功
				if !p.IsRobot() && !scene.Testing && !scene.IsMatchScene() {
					base.LogChannelSington.WriteMQData(model.GenerateEnterEvent(scene.GetRecordId(), p.SnId, p.Platform,
						p.DeviceOS, scene.GameId, scene.GameMode, scene.GetGameFreeId()))
				}
			}
		}
		return nil
	}))

	//观众进入
	netlib.RegisterFactory(int(server.SSPacketID_PACKET_WG_AUDIENCEENTER), netlib.PacketFactoryWrapper(func() interface{} {
		return &server.WGPlayerEnter{}
	}))
	netlib.RegisterHandler(int(server.SSPacketID_PACKET_WG_AUDIENCEENTER), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
		logger.Logger.Trace("receive PACKET_WG_AUDIENCEENTER WGPlayerEnter")
		if msg, ok := pack.(*server.WGPlayerEnter); ok {
			sceneId := int(msg.GetSceneId())
			sid := msg.GetSid()
			data := msg.GetPlayerData()
			gateSid := msg.GetGateSid()
			isload := msg.GetIsLoaded()
			IsQM := msg.GetIsQM()
			var sessionId srvlib.SessionId
			sessionId.Set(gateSid)
			sendLeave := func(reason int) {
				pack := &server.GWPlayerLeave{
					RoomId:     msg.SceneId,
					PlayerId:   msg.SnId,
					ReturnCoin: msg.TakeCoin,
					Reason:     proto.Int(reason),
				}
				proto.SetDefaults(pack)
				s.Send(int(server.SSPacketID_PACKET_GW_AUDIENCELEAVE), pack)
			}
			scene := base.SceneMgrSington.GetScene(sceneId)
			if scene == nil || sid == 0 {
				if sid == 0 {
					logger.Logger.Warnf("when WGAUPlayerEnter (sid == 0)")
				}

				//进入房间失败
				sendLeave(common.PlayerLeaveReason_OnDestroy)
				return nil
			}

			gateSess := srvlib.ServerSessionMgrSington.GetSession(int(sessionId.AreaId()), int(sessionId.SrvType()), int(sessionId.SrvId()))
			logger.Logger.Tracef("PACKET_WG_AUDIENCEENTER WGPlayerEnter, AreaId=%v, SrvType=%v, SrvId=%v, GateSess=%v", int(sessionId.AreaId()), int(sessionId.SrvType()), int(sessionId.SrvId()), gateSess)
			if gateSess != nil {
				pack := &server.GGPlayerSessionBind{
					Sid: proto.Int64(sid),
				}
				proto.SetDefaults(pack)
				gateSess.Send(int(server.SSPacketID_PACKET_GG_PLAYERSESSIONBIND), pack)
			} else {
				//进入房间失败
				logger.Logger.Warnf("PACKET_WG_AUDIENCEENTER WGPlayerEnter, AreaId=%v, SrvType=%v, SrvId=%v, GateSess=<nil>", int(sessionId.AreaId()), int(sessionId.SrvType()), int(sessionId.SrvId()))
				sendLeave(common.PlayerLeaveReason_OnDestroy)
				return nil
			}

			// 自建房检查观众人数上限
			if scene.IsPreCreateScene() {
				if len(scene.GetAudiences()) >= model.GameParamData.MaxAudienceNum {
					sendLeave(common.PlayerLeaveReason_RoomFull)
					return nil
				}
			}

			p := base.PlayerMgrSington.AddPlayer(sid, data, s, gateSess)
			if p == nil {
				//进入房间失败
				sendLeave(common.PlayerLeaveReason_OnDestroy)
				return nil
			}
			p.UnmarshalIParam(msg.GetIParams())
			p.UnmarshalSParam(msg.GetSParams())
			p.UnmarshalCParam(msg.GetCParams())
			p.Coin = msg.GetTakeCoin()
			p.SetTakeCoin(msg.GetTakeCoin())
			p.LastSyncCoin = p.Coin
			p.IsQM = IsQM
			if scene != nil {
				scene.AudienceEnter(p, isload)
				if !p.IsRobot() && !scene.Testing {
					base.LogChannelSington.WriteMQData(model.GenerateEnterEvent(scene.GetRecordId(), p.SnId, p.Platform,
						p.DeviceOS, scene.GameId, scene.GameMode, scene.GetGameFreeId()))
				}
			}
		}
		return nil
	}))

	//观众坐下
	netlib.RegisterFactory(int(server.SSPacketID_PACKET_WG_AUDIENCESIT), netlib.PacketFactoryWrapper(func() interface{} {
		return &server.WGAudienceSit{}
	}))
	netlib.RegisterHandler(int(server.SSPacketID_PACKET_WG_AUDIENCESIT), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
		logger.Logger.Trace("receive PACKET_WG_AUDIENCESIT WGAudienceSit", pack)
		if msg, ok := pack.(*server.WGAudienceSit); ok {

			p := base.PlayerMgrSington.GetPlayerBySnId(msg.GetSnId())
			if p != nil {
				scene := p.GetScene()
				if scene != nil {
					p.Pos = int(msg.GetPos())
					//p.coin = msg.GetTakeCoin()
					//p.takeCoin = msg.GetTakeCoin()
					if scene.Testing {
						p.Coin = int64(scene.DbGameFree.GetTestTakeCoin())
					}
					p.LastSyncCoin = p.Coin
					scene.AudienceSit(p)
				}
			} else {
				leavePack := &server.GWPlayerLeave{
					RoomId:     msg.SceneId,
					PlayerId:   msg.SnId,
					Reason:     proto.Int(common.PlayerLeaveReason_Bekickout),
					ReturnCoin: msg.TakeCoin,
				}
				proto.SetDefaults(leavePack)
				s.Send(int(server.SSPacketID_PACKET_GW_AUDIENCELEAVE), leavePack)
			}
		}
		return nil
	}))

	//玩家返回房间
	netlib.RegisterFactory(int(server.SSPacketID_PACKET_WG_PLAYERRETURN), netlib.PacketFactoryWrapper(func() interface{} {
		return &server.WGPlayerReturn{}
	}))
	netlib.RegisterHandler(int(server.SSPacketID_PACKET_WG_PLAYERRETURN), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
		logger.Logger.Trace("receive WGPlayerReturn")
		if msg, ok := pack.(*server.WGPlayerReturn); ok {
			playerId := msg.GetPlayerId()
			p := base.PlayerMgrSington.GetPlayerBySnId(playerId)
			if p != nil {
				oldFlag := p.GetFlag()
				if !p.IsOnLine() {
					p.MarkFlag(base.PlayerState_Online)
				}
				if p.IsMarkFlag(base.PlayerState_Leave) {
					p.UnmarkFlag(base.PlayerState_Leave)
				}
				if p.GetFlag() != oldFlag {
					p.SyncFlag()
				}
				if p.GetScene() != nil {
					p.GetScene().PlayerReturn(p, msg.GetIsLoaded())
				} else {
					logger.Logger.Warnf("whern (%v) WGPlayerReturn p.scene == nil", playerId)
				}
			} else {
				logger.Logger.Warnf("WGPlayerReturn found player:%v not exist", playerId)
				scene := base.SceneMgrSington.GetScene(int(msg.GetRoomId()))
				if scene != nil {
					p := scene.GetPlayer(msg.GetPlayerId())
					if p != nil {
						logger.Logger.Warnf("WGPlayerReturn found player:%v not exist but in scene:%v gameid:%v", playerId, scene.SceneId, scene.GameId)
					}
				}
				//TODO try leave from room
				pack := &server.GWPlayerForceLeave{
					RoomId:   msg.RoomId,
					PlayerId: msg.PlayerId,
					Reason:   proto.Int(common.PlayerLeaveReason_Bekickout),
					EnterTs:  msg.EnterTs,
				}
				proto.SetDefaults(pack)
				s.Send(int(server.SSPacketID_PACKET_GW_PLAYERFORCELEAVE), pack)
			}
		}
		return nil
	}))

	//玩家掉线
	netlib.RegisterFactory(int(server.SSPacketID_PACKET_WG_PLAYERDROPLINE), netlib.PacketFactoryWrapper(func() interface{} {
		return &server.WGPlayerDropLine{}
	}))
	netlib.RegisterHandler(int(server.SSPacketID_PACKET_WG_PLAYERDROPLINE), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
		logger.Logger.Trace("receive WGPlayerDropLine:", pack)
		if msg, ok := pack.(*server.WGPlayerDropLine); ok {
			sceneId := int(msg.GetSceneId())
			scene := base.SceneMgrSington.GetScene(sceneId)
			if scene != nil {
				scene.PlayerDropLine(msg.GetId())
			}
		}
		return nil
	}))

	//玩家重连
	netlib.RegisterFactory(int(server.SSPacketID_PACKET_WG_PLAYERREHOLD), netlib.PacketFactoryWrapper(func() interface{} {
		return &server.WGPlayerRehold{}
	}))
	netlib.RegisterHandler(int(server.SSPacketID_PACKET_WG_PLAYERREHOLD), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
		logger.Logger.Trace("receive WGPlayerRehold")
		if msg, ok := pack.(*server.WGPlayerRehold); ok {
			sceneId := int(msg.GetSceneId())
			scene := base.SceneMgrSington.GetScene(sceneId)
			if scene != nil {
				var sessionId srvlib.SessionId
				sessionId.Set(msg.GetGateSid())
				gateSess := srvlib.ServerSessionMgrSington.GetSession(int(sessionId.AreaId()), int(sessionId.SrvType()), int(sessionId.SrvId()))
				logger.Logger.Tracef("WGPlayerRehold, AreaId=%v, SrvType=%v, SrvId=%v, SessionId=%v", int(sessionId.AreaId()), int(sessionId.SrvType()), int(sessionId.SrvId()), int64(sessionId))
				if gateSess != nil {
					pack := &server.GGPlayerSessionBind{
						Sid: msg.Sid,
					}
					proto.SetDefaults(pack)
					gateSess.Send(int(server.SSPacketID_PACKET_GG_PLAYERSESSIONBIND), pack)
				}
				p := base.PlayerMgrSington.GetPlayerBySnId(msg.GetId())
				if p != nil {
					base.PlayerMgrSington.ReholdPlayer(p.GetSid(), msg.GetSid(), gateSess)
					scene.PlayerRehold(msg.GetId(), msg.GetSid(), gateSess)
				}
			}
		}
		return nil
	}))

	//玩家换SNID
	netlib.RegisterFactory(int(server.SSPacketID_PACKET_WG_REBIND_SNID), netlib.PacketFactoryWrapper(func() interface{} {
		return &server.WGRebindPlayerSnId{}
	}))
	netlib.RegisterHandler(int(server.SSPacketID_PACKET_WG_REBIND_SNID), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
		logger.Logger.Trace("receive WGRebindPlayerSnId:", pack)
		if msg, ok := pack.(*server.WGRebindPlayerSnId); ok {
			oldSnId := msg.GetOldSnId()
			newSnId := msg.GetNewSnId()
			base.PlayerMgrSington.RebindPlayerSnId(oldSnId, newSnId)
			base.SceneMgrSington.RebindPlayerSnId(oldSnId, newSnId)
		}
		return nil
	}))

	//玩家充值
	netlib.RegisterFactory(int(server.SSPacketID_PACKET_WG_RECHARGE), netlib.PacketFactoryWrapper(func() interface{} {
		return &server.WGHundredOp{}
	}))

	netlib.RegisterHandler(int(server.SSPacketID_PACKET_WG_RECHARGE), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
		logger.Logger.Trace("WGHundredOp Process recv ", pack)
		if wgHundredOp, ok := pack.(*server.WGHundredOp); ok {
			if wgHundredOp.GetOpCode() == 1 {
				snid := wgHundredOp.GetSnid()
				param := wgHundredOp.GetParams()
				p := base.PlayerMgrSington.GetPlayerBySnId(snid)

				if p == nil {
					logger.Logger.Warn("WGHundredOp p == nil")
					return nil
				}

				scene := p.GetScene()
				if scene == nil {
					logger.Logger.Warn("WGHundredOp p.scene == nil")
					return nil
				}

				if !scene.HasPlayer(p) {
					return nil
				}
				//同步用户的充值累加额
				if len(param) > 0 {
					p.CoinPayTotal += param[0]
					if p.TodayGameData != nil {
						p.TodayGameData.RechargeCoin += param[0]
					}
				}
				//第2个参数是vip
				if len(param) > 1 && p.VIP < int32(param[1]) {
					p.VIP = int32(param[1])
				}
				scene.GetScenePolicy().OnPlayerEvent(scene, p, base.PlayerEventRecharge, param)
				return nil
			}
			return nil
		}
		return nil
	}))

	//同步水池设置
	netlib.RegisterFactory(int(server.SSPacketID_PACKET_WG_COINPOOLSETTING), netlib.PacketFactoryWrapper(func() interface{} {
		return &webapi.CoinPoolSetting{}
	}))

	netlib.RegisterHandler(int(server.SSPacketID_PACKET_WG_COINPOOLSETTING), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
		logger.Logger.Trace("CoinPoolSetting Process recv ", pack)
		if wgCoinPoolSetting, ok := pack.(*webapi.CoinPoolSetting); ok {
			base.CoinPoolMgr.UpdateCoinPoolSetting(wgCoinPoolSetting)
			return nil
		}
		return nil
	}))

	netlib.RegisterFactory(int(server.SSPacketID_PACKET_WG_RESETCOINPOOL), netlib.PacketFactoryWrapper(func() interface{} {
		return &server.WGResetCoinPool{}
	}))

	netlib.RegisterHandler(int(server.SSPacketID_PACKET_WG_RESETCOINPOOL), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
		logger.Logger.Trace("WGResetCoinPool Process recv ", pack)
		if wgResetCoinPool, ok := pack.(*server.WGResetCoinPool); ok {
			base.CoinPoolMgr.ResetCoinPool(wgResetCoinPool)
			return nil
		}
		return nil
	}))

	netlib.RegisterFactory(int(server.SSPacketID_PACKET_WG_PROFITCONTROL_CORRECT), netlib.PacketFactoryWrapper(func() interface{} {
		return &server.WGProfitControlCorrect{}
	}))

	netlib.RegisterHandler(int(server.SSPacketID_PACKET_WG_PROFITCONTROL_CORRECT), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
		logger.Logger.Trace("WGProfitControlCorrect Process recv ", pack)
		if wgProfitControlCorrect, ok := pack.(*server.WGProfitControlCorrect); ok {
			base.CoinPoolMgr.EffectCoinPool(wgProfitControlCorrect)
			return nil
		}
		return nil
	}))

	//设置玩家黑白名单
	netlib.RegisterFactory(int(server.SSPacketID_PACKET_WG_SETPLAYERBLACKLEVEL), netlib.PacketFactoryWrapper(func() interface{} {
		return &server.WGSetPlayerBlackLevel{}
	}))

	netlib.RegisterHandler(int(server.SSPacketID_PACKET_WG_SETPLAYERBLACKLEVEL), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
		logger.Logger.Trace("WGSetPlayerBlackLevel Process recv ", pack)
		if wgSetPlayerBlackLevel, ok := pack.(*server.WGSetPlayerBlackLevel); ok {
			p := base.PlayerMgrSington.GetPlayerBySnId(wgSetPlayerBlackLevel.GetSnId())
			if p != nil {
				p.WBLevel = wgSetPlayerBlackLevel.GetWBLevel()
				if p.WBLevel > 0 {
					p.WhiteLevel = p.WBLevel
				} else if p.WBLevel < 0 {
					p.BlackLevel = -p.WBLevel
				}
				p.WBCoinLimit = wgSetPlayerBlackLevel.GetWBCoinLimit()
				p.WBMaxNum = wgSetPlayerBlackLevel.GetMaxNum()
				if wgSetPlayerBlackLevel.GetResetTotalCoin() {
					p.WBCoinTotalIn = 0
					p.WBCoinTotalOut = 0
				}
			}
			return nil
		}
		return nil
	}))

	//同步游戏状态
	netlib.RegisterFactory(int(server.SSPacketID_PACKET_WG_SERVER_STATE), netlib.PacketFactoryWrapper(func() interface{} {
		return &server.ServerState{}
	}))

	netlib.RegisterHandler(int(server.SSPacketID_PACKET_WG_SERVER_STATE), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
		logger.Logger.Trace("PACKET_WG_SERVER_STATE Process recv ", pack)
		if srvState, ok := pack.(*server.ServerState); ok {
			base.ServerStateMgr.SetState(common.GameSessState(srvState.GetSrvState()))
			return nil
		}
		return nil
	}))

	//netlib.RegisterFactory(int(server.SSPacketID_PACKET_WG_DTRoomInfo), netlib.PacketFactoryWrapper(func() interface{} {
	//	return &server.WGDTRoomInfo{}
	//}))
	//netlib.RegisterHandler(int(server.SSPacketID_PACKET_WG_DTRoomInfo), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
	//	logger.Logger.Trace("SSPacketID_PACKET_WG_DTRoomInfo Process recv ", pack)
	//	if msg, ok := pack.(*server.WGDTRoomInfo); ok {
	//		scene := base.SceneMgrSington.GetScene(int(msg.GetRoomId()))
	//		if scene != nil {
	//			data := scene.GetScenePolicy().PacketGameData(scene)
	//			if pack, ok := data.(*server.GWDTRoomInfo); ok {
	//				pack.DataKey = proto.String(msg.GetDataKey())
	//				pack.RoomId = proto.Int32(msg.GetRoomId())
	//			} else {
	//				logger.Logger.Warn("Covert DT scene packet game data error.")
	//			}
	//			scene.SendToWorld(int(server.SSPacketID_PACKET_GW_DTRoomInfo), data)
	//		}
	//		return nil
	//	}
	//	return nil
	//}))
	//
	//netlib.RegisterFactory(int(server.SSPacketID_PACKET_WG_DTRoomFlag), netlib.PacketFactoryWrapper(func() interface{} {
	//	return &server.WGDTRoomFlag{}
	//}))
	//netlib.RegisterHandler(int(server.SSPacketID_PACKET_WG_DTRoomFlag), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
	//	logger.Logger.Trace("SSPacketID_PACKET_WG_DTRoomFlag Process recv ", pack)
	//	if msg, ok := pack.(*server.WGDTRoomFlag); ok {
	//		scene := base.SceneMgrSington.GetScene(int(msg.GetRoomId()))
	//		if scene != nil {
	//			data := base.InterventionData{
	//				Webuser:    msg.GetWebuser(),
	//				Flag:       msg.GetFlag(),
	//				NumOfGames: msg.GetNumGames(),
	//			}
	//			scene.GetScenePolicy().InterventionGame(scene, data)
	//		}
	//		return nil
	//	}
	//	return nil
	//}))
	//
	//netlib.RegisterFactory(int(server.SSPacketID_PACKET_WG_DTRoomResults), netlib.PacketFactoryWrapper(func() interface{} {
	//	return &server.WGRoomResults{}
	//}))
	//netlib.RegisterHandler(int(server.SSPacketID_PACKET_WG_DTRoomResults), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
	//	logger.Logger.Trace("SSPacketID_PACKET_WG_DTRoomResults Process recv:", pack)
	//	if msg, ok := pack.(*server.WGRoomResults); ok {
	//		scene := base.SceneMgrSington.GetScene(int(msg.GetRoomId()))
	//		if scene != nil {
	//			data := base.InterventionResults{
	//				Key:     msg.GetDataKey(),
	//				Webuser: msg.GetWebuser(),
	//				Results: msg.GetResults(),
	//			}
	//			ret := scene.GetScenePolicy().InterventionGame(scene, data)
	//			if pack, ok := ret.(*server.GWRoomResults); ok {
	//				pack.DataKey = proto.String(msg.GetDataKey())
	//			} else {
	//				logger.Logger.Warn("Covert DTRoomResults scene packet game data error.")
	//			}
	//			scene.SendToWorld(int(server.SSPacketID_PACKET_GW_DTRoomResults), ret)
	//		}
	//		return nil
	//	}
	//	return nil
	//}))

	netlib.RegisterFactory(int(server.SSPacketID_PACKET_WG_PlayerOnGameCount), netlib.PacketFactoryWrapper(func() interface{} {
		return &server.WGPayerOnGameCount{}
	}))
	netlib.RegisterHandler(int(server.SSPacketID_PACKET_WG_PlayerOnGameCount), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
		logger.Logger.Trace("SSPacketID_PACKET_WG_PlayerOnGameCount Process recv ", pack)
		if msg, ok := pack.(*server.WGPayerOnGameCount); ok {
			base.CoinPoolMgr.LastDayDtCount = nil
			for _, value := range msg.GetDTCount() {
				base.CoinPoolMgr.LastDayDtCount = append(base.CoinPoolMgr.LastDayDtCount, int(value))
			}
			return nil
		}
		return nil
	}))

	netlib.RegisterFactory(int(server.SSPacketID_PACKET_WG_SyncPlayerSafeBoxCoin), netlib.PacketFactoryWrapper(func() interface{} {
		return &server.WGSyncPlayerSafeBoxCoin{}
	}))
	netlib.RegisterHandler(int(server.SSPacketID_PACKET_WG_SyncPlayerSafeBoxCoin), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
		logger.Logger.Trace("WGSyncPlayerSafeBoxCoin Process recv ", pack)
		if msg, ok := pack.(*server.WGSyncPlayerSafeBoxCoin); ok {
			p := base.PlayerMgrSington.GetPlayerBySnId(msg.GetSnId())
			if p != nil {
				p.SafeBoxCoin = msg.GetSafeBoxCoin()
			}
			return nil
		}
		return nil
	}))

	//更新俱乐部房间配置
	netlib.RegisterFactory(int(server.SSPacketID_PACKET_WG_CLUB_MESSAGE), netlib.PacketFactoryWrapper(func() interface{} {
		return &server.WGClubMessage{}
	}))
	netlib.RegisterHandler(int(server.SSPacketID_PACKET_WG_CLUB_MESSAGE), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
		logger.Logger.Trace("receive WGClubMessage:", pack)
		if msg, ok := pack.(*server.WGClubMessage); ok {
			sceneIds := msg.GetSceneIds()
			for _, id := range sceneIds {
				s := base.SceneMgrSington.GetScene(int(id))
				if s != nil {
					if msg.GetPumpCoin() > 0 {
						s.PumpCoin = int32(msg.GetPumpCoin())
					}
					if msg.GetDBGameFree() != nil {
						s.DbGameFree = msg.GetDBGameFree()
					}
				}
			}
		}
		return nil
	}))
	//更新NiceId
	netlib.RegisterFactory(int(server.SSPacketID_PACKET_GW_NICEIDREBIND), netlib.PacketFactoryWrapper(func() interface{} {
		return &server.WGNiceIdRebind{}
	}))
	netlib.RegisterHandler(int(server.SSPacketID_PACKET_GW_NICEIDREBIND), netlib.HandlerWrapper(func(s *netlib.Session,
		packetid int, pack interface{}) error {
		logger.Logger.Trace("receive WGNiceIdRebind:", pack)
		if msg, ok := pack.(*server.WGNiceIdRebind); ok {
			player := base.PlayerMgrSington.GetPlayerBySnId(msg.GetUser())
			if player != nil {
				player.NiceId = msg.GetNewId()
			}
		}
		return nil
	}))
	//
	netlib.RegisterFactory(int(server.SSPacketID_PACKET_WG_INVITEROBENTERCOINSCENEQUEUE), netlib.PacketFactoryWrapper(func() interface{} {
		return &server.WGInviteRobEnterCoinSceneQueue{}
	}))
	netlib.RegisterHandler(int(server.SSPacketID_PACKET_WG_INVITEROBENTERCOINSCENEQUEUE), netlib.HandlerWrapper(func(s *netlib.Session,
		packetid int, pack interface{}) error {
		logger.Logger.Trace("receive WGInviteRobEnterCoinSceneQueue:", pack)
		if msg, ok := pack.(*server.WGInviteRobEnterCoinSceneQueue); ok {
			base.NpcServerAgentSington.QueueInvite(msg.GetGameFreeId(), msg.GetPlatform(), msg.GetRobNum())
		}
		return nil
	}))
	//
	netlib.RegisterFactory(int(server.SSPacketID_PACKET_WG_GAMEFORCESTART), netlib.PacketFactoryWrapper(func() interface{} {
		return &server.WGGameForceStart{}
	}))
	netlib.RegisterHandler(int(server.SSPacketID_PACKET_WG_GAMEFORCESTART), netlib.HandlerWrapper(func(s *netlib.Session,
		packetid int, pack interface{}) error {
		logger.Logger.Trace("receive WGGameForceStart:", pack)
		if msg, ok := pack.(*server.WGGameForceStart); ok {
			scene := base.SceneMgrSington.GetScene(int(msg.GetSceneId()))
			if scene != nil {
				scene.GetScenePolicy().ForceStart(scene)
				scene.NotifySceneRoundStart(1)
			}
		}
		return nil
	}))

	//邀请机器人进比赛
	netlib.RegisterFactory(int(server.SSPacketID_PACKET_WG_INVITEMATCHROB), netlib.PacketFactoryWrapper(func() interface{} {
		return &server.WGInviteMatchRob{}
	}))
	netlib.RegisterHandler(int(server.SSPacketID_PACKET_WG_INVITEMATCHROB), netlib.HandlerWrapper(func(s *netlib.Session,
		packetid int, pack interface{}) error {
		//logger.Logger.Trace("receive WGInviteMatchRob:", pack)
		if msg, ok := pack.(*server.WGInviteMatchRob); ok {
			base.NpcServerAgentSington.MatchInvite(msg.GetMatchId(), msg.GetPlatform(), msg.GetRobNum(), msg.GetNeedAwait())
		}
		return nil
	}))

	//比赛场底分变化
	netlib.RegisterFactory(int(server.SSPacketID_PACKET_WG_SCENEMATCHBASECHANGE), netlib.PacketFactoryWrapper(func() interface{} {
		return &server.WGSceneMatchBaseChange{}
	}))
	netlib.RegisterHandler(int(server.SSPacketID_PACKET_WG_SCENEMATCHBASECHANGE), netlib.HandlerWrapper(func(s *netlib.Session,
		packetid int, pack interface{}) error {
		logger.Logger.Trace("WGSceneMatchBaseChange Process recv ", pack)
		if msg, ok := pack.(*server.WGSceneMatchBaseChange); ok {
			ids := msg.GetSceneIds()
			for _, id := range ids {
				s := base.SceneMgrSington.GetScene(int(id))
				if s != nil {
					if s.GetMatchChgData() == nil {
						s.SetMatchChgData(&base.SceneMatchChgData{})
					}
					if s.GetMatchChgData() != nil {
						s.GetMatchChgData().NextBaseScore = msg.GetBaseScore()
						s.GetMatchChgData().NextOutScore = msg.GetOutScore()
					}
				}
			}
		}
		return nil
	}))

	//玩家退赛
	netlib.RegisterFactory(int(server.SSPacketID_PACKET_WG_PLAYERQUITMATCH), netlib.PacketFactoryWrapper(func() interface{} {
		return &server.WGPlayerQuitMatch{}
	}))
	netlib.RegisterHandler(int(server.SSPacketID_PACKET_WG_PLAYERQUITMATCH), netlib.HandlerWrapper(func(s *netlib.Session,
		packetid int, pack interface{}) error {
		logger.Logger.Trace("WGPlayerQuitMatch Process recv ", pack)
		if msg, ok := pack.(*server.WGPlayerQuitMatch); ok {
			p := base.PlayerMgrSington.GetPlayerBySnId(msg.GetSnId())
			if p == nil {
				return nil
			}
			scene := base.SceneMgrSington.GetScene(int(msg.GetSceneId()))
			if scene == nil {
				return nil
			}
			if scene.GetParamEx(common.PARAMEX_MATCH_COPYID) != msg.GetMatchId() {
				return nil
			}
			//if scene.mp != nil {
			//	if scene.mp.OnMatchBreak(scene, p.pos) {
			//		//base.PlayerMgrSington.DelPlayerBySnId(p.SnId)
			//		//p.gateSess = nil
			//		//p.worldSess = nil
			//		//p.gateSid = 0
			//		//p.sid = 0
			//		p.SetIParam(common.PlayerIParam_IsQuit, 1)
			//		p.MarkFlag(base.PlayerState_Leave)
			//		p.MarkFlag(PlayerState_Auto)
			//		p.MarkFlag(PlayerState_MatchQuit)
			//		p.SyncFlag()
			//	}
			//}
		}
		return nil
	}))

	//玩家中转消息
	netlib.RegisterFactory(int(server.SSPacketID_PACKET_SS_REDIRECTTOPLAYER), netlib.PacketFactoryWrapper(func() interface{} {
		return &server.SSRedirectToPlayer{}
	}))
	netlib.RegisterHandler(int(server.SSPacketID_PACKET_SS_REDIRECTTOPLAYER), netlib.HandlerWrapper(func(s *netlib.Session,
		packetid int, pack interface{}) error {
		logger.Logger.Trace("SSRedirectToPlayer Process recv ", pack)
		if msg, ok := pack.(*server.SSRedirectToPlayer); ok {
			p := base.PlayerMgrSington.GetPlayerBySnId(msg.GetSnId())
			if p == nil {
				logger.Logger.Trace("SSRedirectToPlayer Process recv p == nil ", msg.GetSnId())
				return nil
			}
			p.SendToClient(int(msg.GetPacketId()), msg.GetData())
		}
		return nil
	}))
	////同步玩家排名信息
	//netlib.RegisterFactory(int(match.MatchPacketID_PACKET_SS_MATCH_PLAYERDATA), netlib.PacketFactoryWrapper(func() interface{} {
	//	return &match.SSMatchPlayerData{}
	//}))
	//netlib.RegisterHandler(int(match.MatchPacketID_PACKET_SS_MATCH_PLAYERDATA), netlib.HandlerWrapper(func(s *netlib.Session,
	//	packetid int, pack interface{}) error {
	//	logger.Logger.Trace("SSMatchPlayerData Process recv ", pack)
	//	if msg, ok := pack.(*match.SSMatchPlayerData); ok {
	//		scene := base.SceneMgrSington.GetScene(int(msg.GetSceneId()))
	//		if scene == nil {
	//			return nil
	//		}
	//		if !scene.IsMatchScene() {
	//			return nil
	//		}
	//		for _, mp := range msg.GetMatchPlayerData() {
	//			if data, ok := scene.Players[mp.GetSnId()]; ok {
	//				data.Iparams[common.PlayerIParam_MatchRank] = int64(mp.GetRank())
	//			}
	//		}
	//	}
	//	return nil
	//}))
	//由worldsrv通知gamesrv向玩家发送奖池信息
	netlib.RegisterFactory(int(server.SSPacketID_PACKET_WG_GAMEJACKPOT), netlib.PacketFactoryWrapper(func() interface{} {
		return &server.WGGameJackpot{}
	}))
	netlib.RegisterHandler(int(server.SSPacketID_PACKET_WG_GAMEJACKPOT), netlib.HandlerWrapper(func(s *netlib.Session,
		packetid int, pack interface{}) error {
		logger.Logger.Trace("WGGameJackpot Process recv ", pack)
		if msg, ok := pack.(*server.WGGameJackpot); ok {
			sid := msg.GetSid()
			gateSid := msg.GetGateSid()
			platform := msg.GetPlatform()
			info := msg.GetInfo()

			var sessionId srvlib.SessionId
			sessionId.Set(gateSid)
			gateSess := srvlib.ServerSessionMgrSington.GetSession(int(sessionId.AreaId()), int(sessionId.SrvType()), int(sessionId.SrvId()))
			pack := &gamehall.SCHundredSceneGetGameJackpot{}
			for _, v := range info {
				if common.InSliceInt(base.BroadJackpotGame, int(v.GameId)) { //不是小游戏且需要广播游戏奖池
					jpfi := &gamehall.GameJackpotFundInfo{
						GameFreeId: proto.Int32(int32(v.GameFreeId)),
					}

					//
					str := base.XSlotsPoolMgr.GetPool(v.GetGameFreeId(), platform)
					if str != "" {
						jackpot := &base.XSlotJackpotPool{}
						err := json.Unmarshal([]byte(str), jackpot)
						if err == nil {
							jpfi.JackPotFund = jackpot.JackpotFund
						}
					}

					//初始化奖池金额
					if jpfi.JackPotFund == 0 {
						dbGameFree := srvdata.PBDB_GameFreeMgr.GetData(v.GameFreeId)
						if dbGameFree != nil {
							params := dbGameFree.GetJackpot()
							jpfi.JackPotFund = int64(params[0] * dbGameFree.GetBaseScore())
						}
					}
					pack.GameJackpotFund = append(pack.GameJackpotFund, jpfi)
				}
			}

			proto.SetDefaults(pack)
			common.SendToGate(sid, int(gamehall.HundredScenePacketID_PACKET_SC_GAMEJACKPOT), pack, gateSess)
			logger.Logger.Trace("SCHundredSceneGetGameJackpot:", pack)
		}
		return nil
	}))
	//单控
	netlib.RegisterFactory(int(server.SSPacketID_PACKET_WG_SINGLEADJUST), netlib.PacketFactoryWrapper(func() interface{} {
		return &server.WGSingleAdjust{}
	}))
	netlib.RegisterHandler(int(server.SSPacketID_PACKET_WG_SINGLEADJUST), netlib.HandlerWrapper(func(s *netlib.Session,
		packetid int, pack interface{}) error {
		logger.Logger.Trace("WGSingleAdjust Process recv ", pack)
		if msg, ok := pack.(*server.WGSingleAdjust); ok {
			//修改内存
			sa := model.UnmarshalSingleAdjust(msg.PlayerSingleAdjust)
			if sa == nil {
				logger.Logger.Warn("WGSingleAdjust sa == nil")
				return nil
			}
			p := base.PlayerMgrSington.GetPlayerBySnId(sa.SnId)
			if p == nil {
				logger.Logger.Warn("WGSingleAdjust p == nil")
				return nil
			}
			switch msg.Option {
			case 1, 2:
				p.UpsertSingleAdjust(sa)
			case 3:
				p.DeleteSingleAdjust(sa.Platform, sa.GameFreeId)
			}
		}
		return nil
	}))
	//玩家离开
	netlib.RegisterFactory(int(server.SSPacketID_PACKET_WG_PlayerLEAVE), netlib.PacketFactoryWrapper(func() interface{} {
		return &server.WGPlayerLeave{}
	}))
	netlib.RegisterHandler(int(server.SSPacketID_PACKET_WG_PlayerLEAVE), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
		logger.Logger.Trace("receive WGPlayerLeaveGame")
		if msg, ok := pack.(*server.WGPlayerLeave); ok {
			p := base.PlayerMgrSington.GetPlayerBySnId(msg.GetSnId())
			if p != nil {
				scene := p.GetScene()
				if scene != nil {
					scene.PlayerLeave(p, common.PlayerLeaveReason_DropLine, false)
				}
			}
		}
		return nil
	}))
}
