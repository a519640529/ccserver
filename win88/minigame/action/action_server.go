package action

import (
	"games.yol.com/win88/common"
	"games.yol.com/win88/minigame/base"
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/server"
	"games.yol.com/win88/protocol/webapi"
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
			base.SceneMgrSington.CreateScene(s, sceneId, gameMode, sceneMode, gameId, msg.GetPlatform(), msg.GetParams(),
				msg.GetAgentor(), msg.GetCreator(), msg.GetReplayCode(), hallId, groupId, totalOfGames, dbGameFree,
				bEnterAfterStart, paramsEx...)
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
					if s.Gaming {
						s.SetGraceDestroy(true)
					} else {
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
	netlib.RegisterFactory(int(server.SSPacketID_PACKET_WG_PLAYERENTER_MINIGAME), netlib.PacketFactoryWrapper(func() interface{} {
		return &server.WGPlayerEnterMiniGame{}
	}))
	netlib.RegisterHandler(int(server.SSPacketID_PACKET_WG_PLAYERENTER_MINIGAME), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
		logger.Logger.Trace("receive WGPlayerEnterMiniGame")
		if msg, ok := pack.(*server.WGPlayerEnterMiniGame); ok {
			sceneId := int(msg.GetSceneId())
			sid := msg.GetSid()
			data := msg.GetPlayerData()
			gateSid := msg.GetGateSid()
			IsQM := msg.GetIsQM()

			sendLeave := func(reason int) {
				pack := &server.GWPlayerLeaveMiniGame{
					SceneId: msg.SceneId,
					SnId:    msg.SnId,
					Reason:  proto.Int(reason),
				}
				proto.SetDefaults(pack)
				s.Send(int(server.SSPacketID_PACKET_GW_PLAYERLEAVE_MINIGAME), pack)
			}
			p := base.NewPlayer(sid, data, nil, nil)
			if p == nil {
				sendLeave(common.PlayerLeaveReason_OnDestroy)
				return nil
			}
			p.Coin = msg.GetTakeCoin()
			p.SetTakeCoin(msg.GetTakeCoin())
			//p.StartCoin = msg.GetTakeCoin()
			//机器人用
			p.ExpectGameTime = msg.GetExpectGameTimes()
			p.ExpectLeaveCoin = msg.GetExpectLeaveCoin()
			//当局游戏结束后剩余金额 起始设置
			p.SetCurrentCoin(msg.GetTakeCoin())

			p.IsQM = IsQM
			p.UnMarshalSingleAdjustData(msg.SingleAdjust)
			scene := base.SceneMgrSington.GetScene(sceneId)
			if sid == 0 || scene == nil {
				if scene == nil {
					logger.Logger.Warn("when WGPlayerEnterMiniGame (scene == nil)")
				}

				if sid == 0 {
					logger.Logger.Warnf("when WGPlayerEnterMiniGame (sid == 0)")
				}

				//进入房间失败
				sendLeave(common.PlayerLeaveReason_OnDestroy)
				return nil
			}

			var sessionId srvlib.SessionId
			sessionId.Set(gateSid)
			gateSess := srvlib.ServerSessionMgrSington.GetSession(int(sessionId.AreaId()), int(sessionId.SrvType()), int(sessionId.SrvId()))
			logger.Logger.Tracef("WGPlayerEnterMiniGame, AreaId=%v, SrvType=%v, SrvId=%v, GateSess=%v", int(sessionId.AreaId()), int(sessionId.SrvType()), int(sessionId.SrvId()), gateSess)
			if gateSess == nil {
				logger.Logger.Warnf("WGPlayerEnterMiniGame, AreaId=%v, SrvType=%v, SrvId=%v, GateSess=<nil>", int(sessionId.AreaId()), int(sessionId.SrvType()), int(sessionId.SrvId()))
				//进入房间失败
				sendLeave(common.PlayerLeaveReason_OnDestroy)
				return nil
			}

			p.SetGateSess(gateSess)
			p.SetWorldSess(s)

			if scene.Testing {
				p.Coin = int64(scene.DbGameFree.GetTestTakeCoin())
			}
			scene.PlayerEnter(p, true)
			//进场失败
			if p.IsMarkFlag(base.PlayerState_EnterSceneFailed) {
				scene.PlayerLeave(p, common.PlayerLeaveReason_Normal, true)
			} else {
				// 进入成功
				if !p.IsRobot() && !scene.Testing && !scene.IsMatchScene() {
					//base.LogChannelSington.WriteMQData(model.GenerateEnterEvent(scene.GetRecordId(), p.SnId, p.Platform,
					//	p.DeviceOS, scene.GameId, scene.GameMode, scene.GetGameFreeId()))
				}
			}
		}
		return nil
	}))
	//玩家离开
	netlib.RegisterFactory(int(server.SSPacketID_PACKET_WG_PLAYERLEAVE_MINIGAME), netlib.PacketFactoryWrapper(func() interface{} {
		return &server.WGPlayerLeaveMiniGame{}
	}))
	netlib.RegisterHandler(int(server.SSPacketID_PACKET_WG_PLAYERLEAVE_MINIGAME), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
		logger.Logger.Trace("receive WGPlayerLeaveMiniGame")
		if msg, ok := pack.(*server.WGPlayerLeaveMiniGame); ok {
			sceneId := int(msg.GetSceneId())
			snId := msg.GetSnId()
			scene := base.SceneMgrSington.GetScene(sceneId)
			p := scene.Players[snId]
			if scene != nil && p != nil {
				scene.PlayerLeave(p, 0, true)
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
				scene.PlayerRehold(msg.GetId(), msg.GetSid(), gateSess)
			}
		}
		return nil
	}))

	//玩家返回房间 gs 添加  rehold 和 return 消息感觉是不严谨的。因为这个时候，房间有可能已经不存在了。这个没有处理。可以参考gamesrv中的处理
	netlib.RegisterFactory(int(server.SSPacketID_PACKET_WG_PLAYERRETURN), netlib.PacketFactoryWrapper(func() interface{} {
		return &server.WGPlayerReturn{}
	}))
	netlib.RegisterHandler(int(server.SSPacketID_PACKET_WG_PLAYERRETURN), netlib.HandlerWrapper(func(s *netlib.Session, packetId int, pack interface{}) error {
		logger.Logger.Trace("receive WGPlayerReturn")
		if msg, ok := pack.(*server.WGPlayerReturn); ok {
			sceneId := int(msg.GetRoomId())
			scene := base.SceneMgrSington.GetScene(sceneId)
			playerSnId := msg.GetPlayerId()
			if scene != nil {
				scene.PlayerReturn(scene.Players[playerSnId], true)
			}
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
		if msg, ok := pack.(*server.WGSetPlayerBlackLevel); ok {
			scene := base.SceneMgrSington.GetScene(int(msg.GetSceneId()))
			if scene != nil {
				p := scene.GetPlayer(msg.GetSnId())
				p.WBLevel = msg.GetWBLevel()
				if p.WBLevel > 0 {
					p.WhiteLevel = p.WBLevel
				} else if p.WBLevel < 0 {
					p.BlackLevel = -p.WBLevel
				}
				p.WBCoinLimit = msg.GetWBCoinLimit()
				p.WBMaxNum = msg.GetMaxNum()
				if msg.GetResetTotalCoin() {
					p.WBCoinTotalIn = 0
					p.WBCoinTotalOut = 0
				}
			}
			return nil
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
			scene := base.SceneMgrSington.GetScene(int(msg.GetSceneId()))
			if scene != nil {
				p := scene.GetPlayer(msg.GetSnId())
				if p == nil {
					logger.Logger.Trace("SSRedirectToPlayer Process recv p == nil ", msg.GetSnId())
					return nil
				}
				p.SendToClient(int(msg.GetPacketId()), msg.GetData())
			}

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
				logger.Logger.Warn("mini WGSingleAdjust sa == nil")
				return nil
			}
			scene := base.SceneMgrSington.GetScene(int(msg.SceneId))
			if scene == nil {
				logger.Logger.Warn("mini WGSingleAdjust scene == nil")
				return nil
			}
			p := scene.GetPlayer(sa.SnId)
			if p == nil {
				logger.Logger.Warn("mini WGSingleAdjust p == nil")
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
}
