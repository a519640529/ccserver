package main

import (
	"errors"
	"fmt"
	"strconv"
	"sync/atomic"
	"time"

	webapi_proto "games.yol.com/win88/protocol/webapi"

	"games.yol.com/win88/common"
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/gamehall"
	"games.yol.com/win88/protocol/message"
	"games.yol.com/win88/protocol/player"
	"games.yol.com/win88/protocol/server"
	"games.yol.com/win88/srvdata"
	"games.yol.com/win88/webapi"
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/core/task"
)

type CSCreatePrivateRoomPacketFactory struct {
}
type CSCreatePrivateRoomHandler struct {
}

func (this *CSCreatePrivateRoomPacketFactory) CreatePacket() interface{} {
	pack := &gamehall.CSCreatePrivateRoom{}
	return pack
}

func (this *CSCreatePrivateRoomHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSCreatePrivateRoomHandler Process recv ", data)
	if msg, ok := data.(*gamehall.CSCreatePrivateRoom); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			return nil
		}

		params := msg.GetParams()
		paramCnt := len(params)
		if paramCnt > CreateRoomParam_Max {
			return nil
		}

		for i := paramCnt; i < CreateRoomParam_Max; i++ {
			params = append(params, 0)
		}

		idx := int(params[CreateRoomParam_NumOfGames] - 1)
		if idx < 0 || idx >= len(model.GameParamData.NumOfGamesConfig) {
			idx = 0
		}
		//局数
		numOfGames := model.GameParamData.NumOfGamesConfig[idx]
		doorOption := params[CreateRoomParam_DoorOption]
		sameIpLimit := params[CreateRoomParam_SameIPForbid]

		var roomId int
		var needCoin int64
		var scene *Scene
		var code gamehall.OpResultCode_Game
		var sp ScenePolicy
		var pps *PlayerPrivateScene
		var dbGameRule *server.DB_GameRule
		var platform *Platform
		gamefreeId := msg.GetGameFreeId()

		//if clubManager.IsClubNotOpen(p.Platform) {
		//	logger.Logger.Trace("CSCreateRoomHandler the platform is closed. platform id = ", p.Platform)
		//	return nil
		//}

		dbGameFree := srvdata.PBDB_GameFreeMgr.GetData(gamefreeId)
		if dbGameFree == nil {
			code = gamehall.OpResultCode_Game_OPRC_GameNotExist_Game
			logger.Logger.Tracef("CSCreateRoomHandler SnId:%v GameFreeId:%v not exist", p.SnId, gamefreeId)
			goto failed
		}

		dbGameRule = srvdata.PBDB_GameRuleMgr.GetData(dbGameFree.GetGameRule())
		if dbGameRule == nil {
			code = gamehall.OpResultCode_Game_OPRC_GameNotExist_Game
			logger.Logger.Tracef("CSCreateRoomHandler SnId:%v GameFreeId:%v gamerule not exist", p.SnId, gamefreeId)
			goto failed
		}

		params = common.CopySliceInt32(dbGameRule.GetParams())
		//var params []int32
		sp = GetScenePolicy(int(dbGameFree.GetGameId()), int(dbGameFree.GetGameMode()))
		if sp == nil {
			code = gamehall.OpResultCode_Game_OPRC_GameNotExist_Game
			logger.Logger.Tracef("CSCreateRoomHandler SnId:%v ScenePolicy(gameid:%v mode:%v) not registe", p.SnId, dbGameFree.GetGameId(), dbGameFree.GetGameMode())
			goto failed
		}

		if spd, ok := sp.(*ScenePolicyData); ok {
			//中途加入参数
			spdp := spd.GetParamByIndex(SPDPCustomIndex_DoorOption)
			if spdp != nil {
				if spdp.index >= 0 && spdp.index < len(params) {
					params[spdp.index] = doorOption
				}
			}

			//同ip限制
			spdp = spd.GetParamByIndex(SPDPCustomIndex_SameIPForbid)
			if spdp != nil {
				if spdp.index >= 0 && spdp.index < len(params) {
					params[spdp.index] = sameIpLimit
				}
			}
		}

		pps = PrivateSceneMgrSington.GetOrCreatePlayerPrivateScene(p)
		if pps == nil {
			code = gamehall.OpResultCode_Game_OPRC_Error_Game
			logger.Logger.Tracef("CSCreateRoomHandler SnId:%v GameFreeId:%v GetOrCreatePlayerPrivateScene", p.SnId, gamefreeId)
			goto failed
		}

		//数量上限
		if pps.GetCount() >= model.GameParamData.CreatePrivateSceneCnt {
			code = gamehall.OpResultCode_Game_OPRC_PrivateRoomCountLimit_Game
			logger.Logger.Tracef("CSCreateRoomHandler SnId:%v GameFreeId:%v GetOrCreatePlayerPrivateScene", p.SnId, gamefreeId)
			goto failed
		}

		//检测局数对应的消耗
		platform = p.GetPlatform()
		if platform != nil {
			needCoin = int64(numOfGames) * platform.ClubConfig.CreateRoomAmount
			if p.Coin < needCoin {
				code = gamehall.OpResultCode_Game_OPRC_CoinNotEnough_Game
				logger.Logger.Tracef("CSCreateRoomHandler SnId:%v GameFreeId:%v MoneyNotEnough", p.SnId, gamefreeId)
				goto failed
			}

			//没有开放俱乐部
			if platform.ClubConfig != nil && !platform.ClubConfig.IsOpenClub {
				code = gamehall.OpResultCode_Game_Oprc_Club_ClubIsClose_Game
				logger.Logger.Tracef("CSCreateRoomHandler SnId:%v GameFreeId:%v MoneyNotEnough", p.SnId, gamefreeId)
				goto failed
			}
		}

		roomId = SceneMgrSington.RandGetSceneId()
		if roomId == common.RANDID_INVALID {
			code = gamehall.OpResultCode_Game_OPRC_AllocRoomIdFailed_Game
			logger.Logger.Trace("CSCreateRoomHandler SnId:%v GameFreeId:%v sceneId == -1 ", p.SnId, gamefreeId)
			goto failed
		}
		scene, code = p.CreateScene(roomId, int(dbGameFree.GetGameId()), int(dbGameFree.GetGameMode()), int(common.SceneMode_Private), int32(numOfGames), params, dbGameFree)
		if scene != nil {
			roomId = scene.sceneId
			if code == gamehall.OpResultCode_Game_OPRC_Sucess_Game {
				if needCoin > 0 {
					scene.createFee = int32(needCoin)
					p.AddCoin(-needCoin, common.GainWay_CreatePrivateScene, "", "")
				}
				pps.OnCreateScene(p, scene)
			}
		}
	failed:
		resp := &gamehall.SCCreatePrivateRoom{
			GameFreeId: msg.GameFreeId,
			Params:     msg.GetParams(),
			RoomId:     proto.Int(roomId),
			OpRetCode:  code,
		}
		proto.SetDefaults(resp)
		p.SendToClient(int(gamehall.GameHallPacketID_PACKET_SC_CREATEPRIVATEROOM), resp)
	}
	return nil
}

type CSEnterRoomPacketFactory struct {
}
type CSEnterRoomHandler struct {
}

func (this *CSEnterRoomPacketFactory) CreatePacket() interface{} {
	pack := &gamehall.CSEnterRoom{}
	return pack
}

func (this *CSEnterRoomHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSEnterRoomHandler Process recv ", data)
	if msg, ok := data.(*gamehall.CSEnterRoom); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			return nil
		}
		var code gamehall.OpResultCode_Game
		var sp ScenePolicy
		var dbGameFree *server.DB_GameFree
		var roomId int
		var gameId int
		var gameMode int
		scene := SceneMgrSington.GetSceneByPlayerId(p.SnId)
		if scene != nil {
			code = gamehall.OpResultCode_Game_OPRC_RoomHadExist_Game
			roomId = scene.sceneId
			gameId = scene.gameId
			gameMode = scene.gameMode
			logger.Logger.Tracef("CSEnterRoomHandler had scene(%d)", scene.sceneId)
			goto failed
		}
		scene = SceneMgrSington.GetScene(int(msg.GetRoomId()))
		if scene == nil {
			code = gamehall.OpResultCode_Game_OPRC_RoomNotExist_Game
			logger.Logger.Trace("CSEnterRoomHandler scene == nil")
			goto failed
		}
		gameId = scene.gameId
		gameMode = scene.gameMode

		if scene.limitPlatform != nil {
			if scene.limitPlatform.Isolated && p.Platform != scene.limitPlatform.IdStr {
				code = gamehall.OpResultCode_Game_OPRC_RoomNotExist_Game
				logger.Logger.Tracef("CSEnterRoomHandler ScenePolicy(gameid:%v mode:%v) scene.limitPlatform.Isolated && p.Platform != scene.limitPlatform.Name", scene.gameId, scene.gameMode)
				goto failed
			}
		}
		if scene.deleting {
			code = gamehall.OpResultCode_Game_OPRC_RoomNotExist_Game
			logger.Logger.Trace("CSEnterRoomHandler scene is deleting")
			goto failed
		}

		if scene.closed {
			code = gamehall.OpResultCode_Game_OPRC_RoomHadClosed_Game
			logger.Logger.Trace("CSEnterRoomHandler scene is closed")
			goto failed
		}

		dbGameFree = scene.dbGameFree
		if dbGameFree != nil {
			limitCoin := srvdata.CreateRoomMgrSington.GetLimitCoinByBaseScore(int32(scene.gameId), int32(scene.gameSite), scene.BaseScore)
			if p.Coin < limitCoin {
				code = gamehall.OpResultCode_Game_OPRC_CoinNotEnough_Game
				logger.Logger.Trace("CSEnterRoomHandler scene is closed")
				goto failed
			}
		}

		sp = GetScenePolicy(scene.gameId, scene.gameMode)
		if sp == nil {
			code = gamehall.OpResultCode_Game_OPRC_GameNotExist_Game
			logger.Logger.Tracef("CSEnterRoomHandler ScenePolicy(gameid:%v mode:%v) not registe", scene.gameId, scene.gameMode)
			goto failed
		}
		//if reason := sp.CanEnter(scene, p); reason != 0 && reason != int(gamehall.OpResultCode_OPRC_SceneEnterForWatcher) {
		if reason := sp.CanEnter(scene, p); reason != 0 {
			code = gamehall.OpResultCode_Game(reason)
			logger.Logger.Trace("CSEnterRoomHandler sp.CanEnter(scene, p) reason ", reason)
			goto failed
		}

		if scene.IsFull() {
			code = gamehall.OpResultCode_Game_OPRC_RoomIsFull_Game
			logger.Logger.Trace("CSEnterRoomHandler ScenePolicy.IsFull = true")
			goto failed
		}

		if !p.EnterScene(scene, true, -1) {
			code = gamehall.OpResultCode_Game_OPRC_Error_Game
		} else {
			//成功进入房间的消息在gameserver上会发送
			CoinSceneMgrSington.OnPlayerEnter(p, dbGameFree.Id)
			return nil
		}
	failed:
		resp := &gamehall.SCEnterRoom{
			GameId:    proto.Int(gameId),
			ModeType:  proto.Int(gameMode),
			RoomId:    proto.Int(roomId),
			OpRetCode: code,
		}
		proto.SetDefaults(resp)
		p.SendToClient(int(gamehall.GameHallPacketID_PACKET_SC_ENTERROOM), resp)
		return nil
	}
	return nil
}

type CSAudienceEnterRoomHandler struct {
}

func (this *CSAudienceEnterRoomHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSAudienceEnterRoomHandler Process recv ", data)
	if msg, ok := data.(*gamehall.CSEnterRoom); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			return nil
		}
		var code gamehall.OpResultCode_Game
		var sp ScenePolicy
		var dbGameFree *server.DB_GameFree
		scene := SceneMgrSington.GetScene(int(msg.GetRoomId()))
		if scene == nil {
			code = gamehall.OpResultCode_Game_OPRC_RoomNotExist_Game
			logger.Logger.Trace("CSAudienceEnterRoomHandler scene == nil")
			goto failed
		}
		if scene.IsMatchScene() {
			code = gamehall.OpResultCode_Game_OPRC_RoomNotExist_Game
			logger.Logger.Tracef("CSAudienceEnterRoomHandler scene.IsMatchScene() %v", scene.sceneId)
			goto failed
		}
		if p.scene != nil {
			code = gamehall.OpResultCode_Game_OPRC_CannotWatchReasonInOther_Game
			logger.Logger.Trace("CSAudienceEnterRoomHandler p.scene != nil")
			goto failed
		}
		//if !scene.starting {
		//	code = gamehall.OpResultCode_Game_OPRC_CannotWatchReasonRoomNotStart_Game
		//	logger.Logger.Trace("CSAudienceEnterRoomHandler !scene.starting")
		//	goto failed
		//}
		if scene.deleting {
			code = gamehall.OpResultCode_Game_OPRC_RoomNotExist_Game
			logger.Logger.Trace("CSAudienceEnterRoomHandler scene is deleting")
			goto failed
		}
		if scene.closed {
			code = gamehall.OpResultCode_Game_OPRC_RoomHadClosed_Game
			logger.Logger.Trace("CSAudienceEnterRoomHandler scene is closed")
			goto failed
		}
		//if scene.IsCoinScene() || scene.IsHundredScene() {
		//	code = gamehall.OpResultCode_Game_OPRC_Error_Game
		//	logger.Logger.Trace("CSAudienceEnterRoomHandler scene is IsCoinScene IsHundredScene")
		//	goto failed
		//}

		sp = GetScenePolicy(scene.gameId, scene.gameMode)
		if sp == nil {
			code = gamehall.OpResultCode_Game_OPRC_GameNotExist_Game
			logger.Logger.Tracef("CSAudienceEnterRoomHandler ScenePolicy(gameid:%v mode:%v) not registe", scene.gameId, scene.gameMode)
			goto failed
		}
		dbGameFree = scene.dbGameFree
		code = gamehall.OpResultCode_Game(CoinSceneMgrSington.AudienceEnter(p, dbGameFree.GetId(), msg.GetRoomId(), nil, true))
		if code == gamehall.OpResultCode_Game_OPRC_Sucess_Game {
			//成功进入房间的消息在gameserver上会发送
			//CoinSceneMgrSington.OnPlayerEnter(p, dbGameFree.Id)
			return nil
		}
	failed:
		resp := &gamehall.SCEnterRoom{
			OpRetCode: code,
		}
		proto.SetDefaults(resp)
		p.SendToClient(int(gamehall.GameHallPacketID_PACKET_SC_ENTERROOM), resp)
		return nil
	}
	return nil
}

type CSReturnRoomPacketFactory struct {
}
type CSReturnRoomHandler struct {
}

func (this *CSReturnRoomPacketFactory) CreatePacket() interface{} {
	pack := &gamehall.CSReturnRoom{}
	return pack
}

func (this *CSReturnRoomHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSReturnRoomHandler Process recv ", data)
	if msg, ok := data.(*gamehall.CSReturnRoom); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			return nil
		}
		scene := p.scene
		pack := &gamehall.SCReturnRoom{}
		if scene == nil {
			//miniGameScene := MiniGameMgrSington.GetAllSceneByPlayer(p)
			isExist := false
			//for _, s := range miniGameScene {
			//	if s.sceneId == int(msg.GetRoomId()) {
			//		isExist = true
			//	}
			//}
			if isExist {
				//如果存在这尝试返回到房间内
				//MiniGameMgrSington.OnPlayerReturnScene(p)
			} else {
				pack.OpRetCode = gamehall.OpResultCode_Game_OPRC_RoomNotExist_Game
				//如果不存在则直接返回
				goto done
			}
		} else {
			pack.RoomId = proto.Int(scene.sceneId)
			pack.GameId = proto.Int(scene.gameId)
			pack.ModeType = proto.Int(scene.gameMode)
			pack.Params = scene.params
			pack.HallId = proto.Int32(scene.hallId)
			gameVers := srvdata.GetGameVers(p.PackageID)
			if ver, ok := gameVers[fmt.Sprintf("%v,%v", scene.gameId, p.Channel)]; ok {
				pack.MinApkVer = proto.Int32(ver.MinApkVer)
				pack.MinResVer = proto.Int32(ver.MinResVer)
				pack.LatestApkVer = proto.Int32(ver.LatestApkVer)
				pack.LatestResVer = proto.Int32(ver.LatestResVer)

				if msg.GetApkVer() < ver.MinApkVer {
					pack.OpRetCode = gamehall.OpResultCode_Game_OPRC_YourAppVerIsLow_Game
					goto done
				}

				if msg.GetResVer() < ver.MinResVer {
					pack.OpRetCode = gamehall.OpResultCode_Game_OPRC_YourResVerIsLow_Game
					goto done
				}
			}
			scene = p.ReturnScene(msg.GetIsLoaded())
			if scene == nil {
				pack.OpRetCode = gamehall.OpResultCode_Game_OPRC_RoomNotExist_Game
			} else {
				//成功返回房间的消息在gamesrv上发送，为了确保房间消息的顺序
				return nil
				//pack.OpRetCode = gamehall.OpResultCode_Game_OPRC_Sucess_Game
			}
		}
	done:
		proto.SetDefaults(pack)
		p.SendToClient(int(gamehall.GameHallPacketID_PACKET_SC_RETURNROOM), pack)
	}
	return nil
}

//type CSInviteRobotPacketFactory struct {
//}
//type CSInviteRobotHandler struct {
//}
//
//func (this *CSInviteRobotPacketFactory) CreatePacket() interface{} {
//	pack := &gamehall.CSInviteRobot{}
//	return pack
//}
//
//func (this *CSInviteRobotHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//	logger.Logger.Trace("CSInviteRobotHandler Process recv ", data)
//	if csInviteRobot, ok := data.(*gamehall.CSInviteRobot); ok {
//		p := PlayerMgrSington.GetPlayer(sid)
//		if p == nil {
//			logger.Logger.Warn("CSInviteRobotHandler p == nil")
//			return nil
//		}
//		if p.scene != nil && p.GMLevel >= model.GMACData.InviteRobot {
//			if csInviteRobot.GetIsAgent() && p.GMLevel < model.GMACData.InviteRobotAgent {
//				logger.Logger.Warnf("CSInviteRobotHandler snid:%v gmlevel:%v InviteRobotAgent:%v", p.SnId, p.GMLevel, model.GMACData.InviteRobotAgent)
//				return nil
//			}
//			hadCnt := p.scene.GetPlayerCnt()
//			needCnt := p.scene.playerNum - hadCnt
//			if needCnt > 0 || csInviteRobot.GetIsAgent() {
//				NpcServerAgentSington.Invite(p.scene.sceneId, needCnt, csInviteRobot.GetIsAgent(), p, 0)
//			}
//		}
//	}
//	return nil
//}

// 读取邮件
type CSReadMessagePacketFactory struct {
}
type CSReadMessageHandler struct {
}

func (this *CSReadMessagePacketFactory) CreatePacket() interface{} {
	pack := &message.CSMessageRead{}
	return pack
}

func (this *CSReadMessageHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSReadMessageHandler Process recv ", data)
	if csMessageRead, ok := data.(*message.CSMessageRead); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSReadMessageHandler p == nil")
			return nil
		}

		p.ReadMessage(csMessageRead.GetId())
	}
	return nil
}

// 删除邮件
type CSDelMessagePacketFactory struct {
}
type CSDelMessageHandler struct {
}

func (this *CSDelMessagePacketFactory) CreatePacket() interface{} {
	pack := &message.CSMessageDel{}
	return pack
}

func (this *CSDelMessageHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSDelMessageHandler Process recv ", data)
	if csMessageDel, ok := data.(*message.CSMessageDel); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSDelMessageHandler p == nil")
			return nil
		}

		p.DelMessage(csMessageDel.GetId(), 1)
	}
	return nil
}

// 提取邮件附件
type CSGetMessageAttachPacketFactory struct {
}
type CSGetMessageAttachHandler struct {
}

func (this *CSGetMessageAttachPacketFactory) CreatePacket() interface{} {
	pack := &message.CSGetMessageAttach{}
	return pack
}

func (this *CSGetMessageAttachHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSGetMessageAttachHandler Process recv ", data)
	if csGetMessageAttach, ok := data.(*message.CSGetMessageAttach); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSGetMessageAttachHandler p == nil")
			return nil
		}

		p.GetMessageAttach(csGetMessageAttach.GetId())
	}
	return nil
}

// 查看邮件
type SCMessageListPacketFactory struct {
}

type SCMessageListHandler struct {
}

func (this *SCMessageListPacketFactory) CreatePacket() interface{} {
	pack := &message.CSMessageList{}
	return pack
}

func (this *SCMessageListHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Tracef("(this *SCMessageListHandler) Process [%v].", s.GetSessionConfig().Id)
	if msg, ok := data.(*message.CSMessageList); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		p.SendMessage(msg.GetShowId())
	}
	return nil
}

// 中途入场
type CSJoinGamePacketFactory struct {
}
type CSJoinGameHandler struct {
}

func (this *CSJoinGamePacketFactory) CreatePacket() interface{} {
	pack := &gamehall.CSJoinGame{}
	return pack
}

//func (this *CSJoinGameHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
//	logger.Logger.Trace("CSJoinGameHandler Process recv ", data)
//	if msg, ok := data.(*gamehall.CSJoinGame); ok {
//		p := PlayerMgrSington.GetPlayer(sid)
//		if p == nil {
//			logger.Logger.Warn("CSJoinGameHandler p == nil")
//			return nil
//		}
//		scene := SceneMgrSington.GetSceneByPlayerId(p.SnId)
//		if scene == nil {
//			logger.Logger.Warn("CSJoinGameHandler scene == nil")
//			return nil
//		}
//
//		msgType := msg.GetMsgType()
//		//全部同意，可以加入
//		if msgType == 0 {
//			pos := msg.GetPos()
//			if pos != -1 { //-1表示自动匹配位置
//				if pos < 0 || pos > int32(scene.playerNum) {
//					logger.Logger.Warn("CSJoinGameHandler pos < 0 || pos > scene.playerNum ", pos, scene.playerNum)
//					return nil
//				}
//				if scene.seats[pos] != nil {
//					logger.Logger.Trace("CSJoinGameHandler scene.seats[pos] != nil ", pos)
//					return nil
//				}
//			}
//
//			if scene.HasPlayer(p) {
//				logger.Logger.Warn("CSJoinGameHandler scene.HasPlayer(p)", p.Name)
//				return nil
//			}
//
//			opCode := scene.EnterCheck(p)
//			if opCode != 0 {
//				logger.Logger.Tracef("====scene.EnterCheck(p) opCode:%v", opCode)
//				p.applyPos = -1
//				pack := &gamehall.SCJoinGame{
//					MsgType:   proto.Int32(msgType),
//					SnId:      proto.Int32(p.SnId),
//					OpRetCode: gamehall.OpResultCode(opCode),
//				}
//				proto.SetDefaults(pack)
//				p.SendToClient(int(gamehall.GameHallPacketID_PACKET_SC_JOINGAME), pack)
//				return nil
//			}
//			//TODO 先调整为直接加入
//			p.applyPos = msg.GetPos()
//			scene.AudienceSit(p, int(p.applyPos))
//			return nil
//		} else {
//			snid := msg.GetSnId()
//			agree := msg.GetAgree()
//			if info, ok := scene.joinList[snid]; ok { //正在等待确认处理
//				flagArr := info.applyStatus
//				if _, ok := flagArr[p.SnId]; !ok {
//					flagArr[p.SnId] = agree
//				}
//				if !agree || len(flagArr) == info.expectCnt {
//					//完成确认，检查确认标记
//					if info.applyTimer != timer.TimerHandle(0) {
//						timer.StopTimer(info.applyTimer)
//					}
//					delete(scene.joinList, snid)
//					for !agree {
//						//有人不同意，不能加入
//						pack := &gamehall.SCJoinGame{
//							MsgType:   proto.Int32(msgType),
//							SnId:      proto.Int32(snid),
//							OpRetCode: gamehall.OpResultCode_OPRC_SceneRefuse,
//						}
//						proto.SetDefaults(pack)
//						scene.SendToClient(int(gamehall.GameHallPacketID_PACKET_SC_JOINGAME), pack, p.SnId)
//						applyPlayer := PlayerMgrSington.GetPlayerBySnId(snid)
//						if applyPlayer != nil {
//							applyPlayer.applyPos = -1
//							applyPlayer.SendToClient(int(gamehall.GameHallPacketID_PACKET_SC_JOINGAME), pack)
//						}
//						return nil
//					}
//					//全部同意，可以加入
//					newPlayer := PlayerMgrSington.GetPlayerBySnId(snid)
//					if newPlayer != nil {
//						opCode := scene.EnterCheck(newPlayer)
//						if opCode == 0 {
//							scene.AudienceSit(newPlayer, int(newPlayer.applyPos))
//						} else {
//							logger.Logger.Tracef("====scene.EnterCheck(newPlayer) opCode:%v", opCode)
//						}
//						pack := &gamehall.SCJoinGame{
//							MsgType:   proto.Int32(msgType),
//							OpRetCode: gamehall.OpResultCode(opCode),
//						}
//						newPlayer.SendToClient(int(gamehall.GameHallPacketID_PACKET_SC_JOINGAME), pack)
//					}
//				}
//			}
//		}
//	}
//	return nil
//}

func (this *CSJoinGameHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSJoinGameHandler Process recv ", data)
	if msg, ok := data.(*gamehall.CSJoinGame); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSJoinGameHandler p == nil")
			return nil
		}
		scene := SceneMgrSington.GetSceneByPlayerId(p.SnId)
		if scene == nil {
			logger.Logger.Warn("CSJoinGameHandler scene == nil")
			return nil
		}

		newPlayer := PlayerMgrSington.GetPlayerBySnId(msg.GetSnId())
		if newPlayer != nil {
			pack := &gamehall.SCJoinGame{
				MsgType:   proto.Int32(msg.GetMsgType()),
				OpRetCode: gamehall.OpResultCode_Game(0),
			}
			if !scene.IsTestScene() {
				// 入场限额检查
				if scene.dbGameFree.GetLimitCoin() != 0 && int64(scene.dbGameFree.GetLimitCoin()) > p.Coin {
					pack.OpRetCode = gamehall.OpResultCode_Game_OPRC_MoneyNotEnough_Game
					newPlayer.SendToClient(int(gamehall.GameHallPacketID_PACKET_SC_JOINGAME), pack)
					return nil
				}
				// 携带金币检查
				if scene.dbGameFree.GetMaxCoinLimit() != 0 && int64(scene.dbGameFree.GetMaxCoinLimit()) < p.Coin && !p.IsRob {
					pack.OpRetCode = gamehall.OpResultCode_Game_OPRC_CoinTooMore_Game
					newPlayer.SendToClient(int(gamehall.GameHallPacketID_PACKET_SC_JOINGAME), pack)
					return nil
				}
			}
			// 是否还有空座位
			if scene.IsFull() {
				pack.OpRetCode = gamehall.OpResultCode_Game(1)
				newPlayer.SendToClient(int(gamehall.GameHallPacketID_PACKET_SC_JOINGAME), pack)
				return nil
			}
			scene.AudienceSit(newPlayer, -1)
			newPlayer.SendToClient(int(gamehall.GameHallPacketID_PACKET_SC_JOINGAME), pack)
		}
	}
	return nil
}

type CSGetDataLogPacketFactory struct {
}
type CSGetDataLogHandler struct {
}

func (this *CSGetDataLogPacketFactory) CreatePacket() interface{} {
	pack := &player.CSGetDataLog{}
	return pack
}

func (this *CSGetDataLogHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSGetDataLogHandler Process recv ", data)
	if msg, ok := data.(*player.CSGetDataLog); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSGetDataLogHandler p == nil")
			return nil
		}
		ts := int64(msg.GetVer())
		if ts == 0 {
			ts = time.Now().Unix()
		}
		type LogData struct {
			datas []model.CoinLog
			err   error
		}
		task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
			datas, err := model.GetCoinLogBySnidAndLessTs(p.Platform, p.SnId, ts)
			return &LogData{datas: datas, err: err}
		}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
			if d, ok := data.(*LogData); ok {
				if d != nil && d.err == nil {
					pack := &player.SCGetDataLog{
						DataType: msg.DataType,
					}
					ver := msg.GetVer()
					for i := 0; i < len(d.datas); i++ {
						ts := d.datas[i].Time.Unix()
						if int32(ts) < ver {
							ver = int32(ts)
						}
						pack.Datas = append(pack.Datas, &player.DataLog{
							LogType:     proto.Int32(d.datas[i].LogType),
							ChangeCount: proto.Int64(d.datas[i].Count),
							RestCount:   proto.Int64(d.datas[i].RestCount),
							Remark:      proto.String(d.datas[i].Remark),
							Ts:          proto.Int32(int32(ts)),
						})
					}
					pack.Ver = proto.Int32(ver)
					proto.SetDefaults(pack)
					p.SendToClient(int(player.PlayerPacketID_PACKET_SC_GETDATALOG), pack)
				}
			}
		}), "GetCoinLogBySnidAndLessTs").StartByFixExecutor("coinlog_r")
	}
	return nil
}

type CSEnterHallPacketFactory struct {
}
type CSEnterHallHandler struct {
}

func (this *CSEnterHallPacketFactory) CreatePacket() interface{} {
	pack := &gamehall.CSEnterHall{}
	return pack
}

func (this *CSEnterHallHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSEnterHallHandler Process recv ", data)
	if msg, ok := data.(*gamehall.CSEnterHall); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSEnterHallHandler p == nil")
			return nil
		}
		PlatformMgrSington.PlayerEnterHall(p, msg.GetHallId())
		pack := &gamehall.SCEnterHall{
			HallId:    msg.HallId,
			OpRetCode: gamehall.OpResultCode_Game_OPRC_Sucess_Game,
		}
		proto.SetDefaults(pack)
		p.SendToClient(int(gamehall.GameHallPacketID_PACKET_SC_ENTERHALL), pack)
	}
	return nil
}

type CSLeaveHallPacketFactory struct {
}
type CSLeaveHallHandler struct {
}

func (this *CSLeaveHallPacketFactory) CreatePacket() interface{} {
	pack := &gamehall.CSLeaveHall{}
	return pack
}

func (this *CSLeaveHallHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSLeaveHallHandler Process recv ", data)
	p := PlayerMgrSington.GetPlayer(sid)
	if p == nil {
		logger.Logger.Warn("CSLeaveHallHandler p == nil")
		return nil
	}
	hallId := PlatformMgrSington.PlayerLeaveHall(p)
	pack := &gamehall.SCLeaveHall{
		HallId: proto.Int32(hallId),
	}
	proto.SetDefaults(pack)
	p.SendToClient(int(gamehall.GameHallPacketID_PACKET_SC_LEAVEHALL), pack)
	return nil
}

type CSHallRoomListPacketFactory struct {
}
type CSHallRoomListHandler struct {
}

func (this *CSHallRoomListPacketFactory) CreatePacket() interface{} {
	pack := &gamehall.CSHallRoomList{}
	return pack
}

func (this *CSHallRoomListHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSHallRoomListHandler Process recv ", data)
	if msg, ok := data.(*gamehall.CSHallRoomList); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSHallRoomListHandler p == nil")
			return nil
		}

		PlatformMgrSington.PlayerEnterHall(p, msg.GetHallId())
	}
	return nil
}

type CSEnterDgGamePacketFactory struct {
}
type CSEnterDgGameHandler struct {
}

func (this *CSEnterDgGamePacketFactory) CreatePacket() interface{} {
	pack := &gamehall.CSEnterDgGame{}
	return pack
}

func (this *CSEnterDgGameHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSEnterDgGameHandler Process recv ", data)
	//if msg, ok := data.(*gamehall.CSEnterDgGame); ok {
	//	p := PlayerMgrSington.GetPlayer(sid)
	//	if p == nil {
	//		logger.Logger.Warn("CSEnterDgGameHandler p == nil")
	//		return nil
	//	}
	//	returnErrorCodeFunc := func(code gamehall.OpResultCode_Game) {
	//		pack := &gamehall.SCEnterDgGame{
	//			DgGameId: proto.Int32(msg.GetDgGameId()),
	//		}
	//		pack.OpRetCode = code
	//		p.SendToClient(int(gamehall.GameHallPacketID_PACKET_SC_ENTERDGGAME), pack)
	//	}
	//	if p.scene != nil {
	//		logger.Logger.Infof("Player %v in dg game.", p.SnId)
	//		returnErrorCodeFunc(gamehall.OpResultCode_Game_OPRC_Dg_LoginErr_Game)
	//		return nil
	//	}
	//
	//	if p.isDelete { //删档用户不让再进游戏
	//		returnErrorCodeFunc(gamehall.OpResultCode_Game_OPRC_RoomHadClosed_Game)
	//		return nil
	//	}
	//
	//	pt := PlatformMgrSington.GetPackageTag(p.PackageID)
	//	if pt != nil && pt.IsForceBind == 1 {
	//		if p.BeUnderAgentCode == "" || p.BeUnderAgentCode == "0" {
	//			returnErrorCodeFunc(gamehall.OpResultCode_Game_OPRC_MustBindPromoter_Game)
	//			return nil
	//		}
	//	}
	//
	//	var agentName, agentKey, thirdPlf = PlatformMgrSington.GetPlatformDgAgentConfig(p.Platform)
	//	if len(agentName) == 0 || len(agentKey) == 0 {
	//		logger.Logger.Infof("Player %v platfrom dg game not open.", p.SnId)
	//		returnErrorCodeFunc(gamehall.OpResultCode_Game_OPRC_Dg_PlatErr_Game)
	//		return nil
	//	}
	//
	//	var dbGameFreeId int32 = -1
	//	if thirdPlf == "HBO" {
	//		switch msg.GetDgGameId() {
	//		case 1: //百家乐 280010001
	//			dbGameFreeId = 280010001
	//		case 3: //龙虎 280020001
	//			dbGameFreeId = 280020001
	//		case 7: //斗牛 280030001
	//			dbGameFreeId = 280030001
	//		case 4: //HBO视讯轮盘
	//			dbGameFreeId = 280040001
	//		case 5: //HBO视讯骰宝
	//			dbGameFreeId = 280050001
	//		case 11: //HBO视讯炸金花
	//			dbGameFreeId = 280060001
	//		}
	//	} else if thirdPlf == "DG" {
	//		switch msg.GetDgGameId() {
	//		case 1: //百家乐 280010001
	//			dbGameFreeId = 280010001
	//		case 3: //龙虎 280020001
	//			dbGameFreeId = 280020001
	//		case 7: //斗牛 280030001
	//			dbGameFreeId = 280030001
	//		}
	//	}
	//	if dbGameFreeId == -1 {
	//		returnErrorCodeFunc(gamehall.OpResultCode_Game_OPRC_RoomHadClosed_Game)
	//		return nil
	//	}
	//
	//	domains := msg.GetDomains()
	//	if len(domains) > 20 {
	//		returnErrorCodeFunc(gamehall.OpResultCode_Game_OPRC_Dg_LoginErr_Game)
	//		return nil
	//	}
	//	//检测房间状态是否开启
	//	cfgid := PlatformMgrSington.GetPlatformConfigId(p.Platform, p.Channel)
	//	if !PlatformMgrSington.CheckGameState(int32(cfgid), dbGameFreeId, 0) {
	//		returnErrorCodeFunc(gamehall.OpResultCode_Game_OPRC_RoomHadClosed_Game)
	//		return nil
	//	}
	//	gps := PlatformMgrSington.GetGameConfig(int32(cfgid), dbGameFreeId)
	//	if gps == nil {
	//		returnErrorCodeFunc(gamehall.OpResultCode_Game_OPRC_RoomHadClosed_Game)
	//		return nil
	//	}
	//	dbGameFree := gps.DBGameFree
	//	if gps.GroupId != 0 {
	//		pgg := PlatformGameGroupMgrSington.GetGameGroup(gps.GroupId)
	//		if pgg != nil {
	//			dbGameFree = pgg.DBGameFree
	//		}
	//	}
	//	if dbGameFree == nil {
	//		returnErrorCodeFunc(gamehall.OpResultCode_Game_OPRC_RoomHadClosed_Game)
	//		return nil
	//	}
	//
	//	//客户要求先把限额去掉，可通过后台修改
	//	if dbGameFree.GetLimitCoin() != 0 && int64(dbGameFree.GetLimitCoin()) > p.Coin {
	//		returnErrorCodeFunc(gamehall.OpResultCode_Game_OPRC_MoneyNotEnough_Game)
	//		return nil
	//	}
	//
	//	//DG不检查游戏次数限制
	//	//todayData, _ := p.GetDaliyGameData(int(dbGameFreeId))
	//	//if dbGameFree.GetPlayNumLimit() != 0 &&
	//	//	todayData != nil &&
	//	//	todayData.GameTimes >= int64(dbGameFree.GetPlayNumLimit()) {
	//	//	logger.Logger.Infof("CSEnterDgGameHandler player snid:%v todayData.GameTimes:%v>PlayNumLimit:%v then kick", p.SnId, todayData.GameTimes, dbGameFree.GetPlayNumLimit())
	//	//	returnErrorCodeFunc(gamehall.OpResultCode_OPRC_RoomGameTimes)
	//	//	return nil
	//	//}
	//
	//	//检查平台DG配额是否足够
	//	if model.GameParamData.DGCheckPlatformQuota {
	//		if dbGameFree.GetLimitCoin() != 0 {
	//			dgQuota := ThirdPlatformMgrSington.GetThirdPlatformCoin(p.Platform, model.THDPLATFORM_DG)
	//			if dgQuota <= 0 || dgQuota <= int64(dbGameFree.GetLimitCoin()) {
	//				logger.Logger.Infof("Player %v platfrom Quota of dg game not enough.", p.SnId)
	//				returnErrorCodeFunc(gamehall.OpResultCode_Game_OPRC_Dg_QuotaNotEnough_Game)
	//				return nil
	//			}
	//		}
	//	}
	//
	//	scene := SceneMgrSington.GetDgScene()
	//	if scene != nil {
	//		p.scene = scene //预设数据，避免连点
	//	}
	//	var err error
	//	var codeId int
	//	var token, _, domain string
	//	var nickName = p.Name
	//	var sList []string
	//	var DgGame, DgPass string
	//	platformInfo := p.GetPlatform()
	//	task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
	//		if msg.GetLoginType() == 0 {
	//			/*
	//				err, codeId, token, _, _ = webapi.API_DgFree(thirdPlf, common.GetAppId(), p.DgGame, p.DgPass,
	//					agentName, agentKey)
	//				if err != nil {
	//					logger.Logger.Error("API_DgFree error:", err)
	//					return err
	//				}
	//				if codeId != 0 {
	//					logger.Logger.Info("API_DgFree code:", codeId)
	//					return errors.New("Dg free error.")
	//				}*/
	//			return nil
	//		} else {
	//			DgGame, DgPass = p.GetDgHboPlayerName(platformInfo)
	//			if len(DgGame) == 0 {
	//				userName := fmt.Sprintf("%v%v", "dg", p.SnId)
	//				if thirdPlf == "HBO" {
	//					userName = fmt.Sprintf("%v%v", "hbo", p.SnId)
	//				}
	//				pass := common.MakeMd5String(userName, time.Now().String())
	//				if len(pass) < 15 {
	//					pass = "ca9af0be12b2cd4ef55bb442a8784605"
	//				}
	//				err, codeId, token, _ = webapi.API_DgSignup(thirdPlf, common.GetAppId(), userName, pass[:15],
	//					nickName, agentName, agentKey)
	//				if err != nil {
	//					logger.Logger.Error("API_DgSignup error:", err)
	//					return err
	//				}
	//				if codeId != 0 && codeId != 116 {
	//					logger.Logger.Info("API_DgSignup code:", codeId)
	//					return errors.New("Dg sign up error.")
	//				}
	//				p.SetDgHboPlayerName(platformInfo, userName, pass[:15])
	//				DgGame = userName
	//				DgPass = pass[:15]
	//			}
	//			err, codeId, token, _, _, domain, sList = webapi.API_DgLogin(thirdPlf, common.GetAppId(), DgGame, DgPass,
	//				agentName, agentKey, p.Ip, domains)
	//			if err != nil {
	//				logger.Logger.Error("API_DgLogin error:", err)
	//				return err
	//			}
	//			if codeId != 0 {
	//				logger.Logger.Info("API_DgLogin code:", codeId)
	//				return errors.New("Dg login error.")
	//			}
	//			return nil
	//		}
	//	}), task.CompleteNotifyWrapper(func(data interface{}, t *task.Task) {
	//		pack := &gamehall.SCEnterDgGame{
	//			DgGameId: proto.Int32(msg.GetDgGameId()),
	//			CodeId:   proto.Int(codeId),
	//			Domains:  proto.String(domain),
	//			List:     sList,
	//		}
	//
	//		logger.Trace("pack.GetDomains()=", pack.GetDomains())
	//		if data == nil {
	//			scene := SceneMgrSington.GetDgScene()
	//			if scene != nil {
	//				p.BakDgHboName = DgGame
	//				p.scene = scene
	//				p.sceneCoin = p.Coin
	//				pack.Token = proto.String(token)
	//				pack.OpRetCode = gamehall.OpResultCode_Game_OPRC_Sucess_Game
	//				logger.Trace("data==nil pack=", pack)
	//				p.SendToClient(int(gamehall.GameHallPacketID_PACKET_SC_ENTERDGGAME), pack)
	//				return
	//			}
	//		}
	//		p.scene = nil //失败清空数据
	//		pack.OpRetCode = gamehall.OpResultCode_Game_OPRC_Dg_RegistErr_Game
	//		p.SendToClient(int(gamehall.GameHallPacketID_PACKET_SC_ENTERDGGAME), pack)
	//	}), "CSEnterDgGameHandler").Start()
	//}
	return nil
}

type CSLeaveDgGamePacketFactory struct {
}
type CSLeaveDgGameHandler struct {
}

func (this *CSLeaveDgGamePacketFactory) CreatePacket() interface{} {
	pack := &gamehall.CSLeaveDgGame{}
	return pack
}

func (this *CSLeaveDgGameHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSLeaveDgGameHandler Process recv ", data)
	if _, ok := data.(*gamehall.CSLeaveDgGame); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSLeaveDgGameHandler p == nil")
			return nil
		}
		if p.scene != nil {
			if p.scene.sceneId == SceneMgrSington.GetDgSceneId() {
				p.scene = nil
			}
		}
		pack := &gamehall.SCLeaveDgGame{}
		pack.OpRetCode = gamehall.OpResultCode_Game_OPRC_Sucess_Game
		p.SendToClient(int(gamehall.GameHallPacketID_PACKET_SC_LEAVEDGGAME), pack)
		p.diffData.Coin = -1 //强制更新金币
		p.SendDiffData()
	}
	return nil
}

// 第三方-->系统
func _StartTransferThird2SystemTask(p *Player) {

	if p.thrscene == nil {
		logger.Logger.Tracef("player snid=%v TransferThird2SystemTask p.scene == nil return", p.SnId)
		return
	}

	if p.thrscene.sceneMode != int(common.SceneMode_Thr) {
		logger.Logger.Infof("player snid=%v TransferThird2SystemTask p.scene == thrID return", p.SnId)
		return
	}

	if p.thridBalanceRefreshReqing {
		logger.Logger.Tracef("player snid=%v TransferThird2SystemTask p.thridBalanceRefreshReqing == true return", p.SnId)
		return
	}
	gainway := common.GainWay_Transfer_Thrid2System
	plt := webapi.ThridPlatformMgrSington.FindPlatformByPlatformBaseGameId(p.thrscene.gameId)
	if plt == nil {
		logger.Logger.Tracef("player snid=%v TransferThird2SystemTask plt == nil return", p.SnId)
		return
	}

	oper := plt.GetPlatformBase().Name + "2System"
	amount := int64(0)
	timeStamp := time.Now().UnixNano()
	p.thridBalanceRefreshReqing = true
	//timeout := false
	task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
		for i := int32(0); i < model.GameParamData.ThirdPltTransferMaxTry; {
			var err error
			var coinLog *model.PayCoinLog
			var coinlogex *model.CoinLog
			//var apiHasTransfer = false
			remark := plt.GetPlatformBase().Name + "转出到系统"
			//allow := plt.ReqIsAllowTransfer(p.SnId, p.Platform, p.Channel)
			//if !allow {
			//	goto Rollback
			//}
			err, amount = plt.ReqLeaveGame(p.SnId, "901", p.Ip, p.Platform, p.Channel)
			if err != nil {
				goto Rollback
			}
			if amount <= 0 {
				return nil
			}
			if plt.GetPlatformBase().TransferInteger {
				amount = (amount / 100) * 100
				if amount <= 0 {
					return nil
				}
			}
			timeStamp = time.Now().UnixNano()
			//如果請求超時的話資金就不能從三方出來，玩家的錢只會少
			//err, timeout = plt.ReqTransfer(p.SnId, -amount, strconv.FormatInt(timeStamp, 10), p.Platform, p.Channel, p.Ip)
			//if err != nil || timeout {
			//	goto Rollback
			//}
			//apiHasTransfer = true
			coinLog = model.NewPayCoinLog(timeStamp, int32(p.SnId), amount, int32(gainway), oper, model.PayCoinLogType_Coin, 0)
			timeStamp = coinLog.TimeStamp
			err = model.InsertPayCoinLogs(p.Platform, coinLog)
			if err != nil {
				goto Rollback
			}
			coinlogex = model.NewCoinLogEx(p.SnId, amount, p.Coin+amount, p.SafeBoxCoin, 0, int32(gainway),
				0, oper, remark, p.Platform, p.Channel, p.BeUnderAgentCode, 0, p.PackageID, int32(plt.GetPlatformBase().VultGameID))
			err = model.InsertCoinLog(coinlogex)
			if err != nil {
				goto Rollback
			}
			return nil
		Rollback:
			if coinLog != nil {
				model.RemovePayCoinLog(p.Platform, coinLog.LogId)
			}
			if coinlogex != nil {
				model.RemoveCoinLogOne(coinlogex.Platform, coinlogex.LogId)
			}
			//如果發現有任何一個超時，則就不在往下執行，因為不知道數據是否正確
			/*
				if timeout {
					logger.Logger.Errorf("player snid=%v third->system transfer %v timeout at try %v times,then stop try!", p.SnId, -amount, i+1)
					break
				}
				if apiHasTransfer {
					err, timeout = plt.ReqTransfer(p.SnId, amount, strconv.FormatInt(time.Now().UnixNano(), 10), p.Platform, p.Channel, p.Ip)
					if err != nil {
						logger.Logger.Errorf("player snid=%v third->system transfer rollback fail at try %v times", p.SnId, i+1)
					}
				}
				//如果发现有任何一個超时，則就不在往下执行，因为不知道数据是否在三方已经处理
				if timeout {
					logger.Logger.Errorf("player snid=%v third->system rollback transfer %v timeout at try %v times,then stop try!", p.SnId, amount, i+1)
					break
				}
			*/
			logger.Logger.Tracef("player snid=%v third->system transfer rollback at try %v times", p.SnId, i+1)
			i++
			if i < model.GameParamData.ThirdPltTransferMaxTry {
				time.Sleep(time.Duration(model.GameParamData.ThirdPltTransferInterval) * time.Duration(time.Second))
			}
		}
		return errors.New("third->system transfer error >max try times!")
	}), task.CompleteNotifyWrapper(func(data interface{}, tt task.Task) {
		statePack := &gamehall.SCThridGameBalanceUpdateState{}
		if data != nil {
			p.thirdBalanceRefreshMark[plt.GetPlatformBase().Name] = false
			p.thridBalanceReqIsSucces = false
			statePack.OpRetCode = gamehall.OpResultCode_Game_OPRC_Error_Game
			logger.Logger.Trace("SCThridAccountTransferHandler third->system transfer error:", data)
		} else {
			p.thridBalanceReqIsSucces = true
			statePack.OpRetCode = gamehall.OpResultCode_Game_OPRC_Sucess_Game
			// p.thrscene = nil
			p.Coin += amount
			p.SetPayTs(timeStamp)
			p.thirdBalanceRefreshMark[plt.GetPlatformBase().Name] = true
			//ThirdPlatformMgrSington.AddThirdPlatformCoin(p.Platform, plt.GetPlatformBase().Tag, thirdBalance)
		}
		p.SendToClient(int(gamehall.GameHallPacketID_PACKET_SC_THRIDGAMEBALANCEUPDATESTATE), statePack)
		p.dirty = true

		//这个地方虽然说拉取失败，但是为了不影响玩家玩其他的游戏，还是可以进入其它场景
		//后面玩家自己通过手动刷新余额
		if p.thrscene != nil {
			p.thrscene = nil
		}
		p.thridBalanceRefreshReqing = false
		p.diffData.Coin = -1 //强制更新金币
		p.SendDiffData()
		return
	}), "ThridAccountTransfer").Start()
}

// 进入三方
type CSEnterThridGamePacketFactory struct {
}
type CSEnterThridGameHandler struct {
}

func (this *CSEnterThridGamePacketFactory) CreatePacket() interface{} {
	pack := &gamehall.CSEnterThridGame{}
	return pack
}

func (this *CSEnterThridGameHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSEnterThridGameHandler Process recv ", data)
	p := PlayerMgrSington.GetPlayer(sid)
	if p == nil {
		logger.Logger.Warn("CSEnterThridGameHandler p == nil")
		return nil
	}

	if msg, ok := data.(*gamehall.CSEnterThridGame); ok {
		msg.ThridGameId = proto.Int32(9010001)
		returnErrorCodeFunc := func(code gamehall.OpResultCode_Game) {
			pack := &gamehall.SCEnterThridGame{}
			pack.OpRetCode = code
			pack.ThridGameId = msg.ThridGameId
			p.SendToClient(int(gamehall.GameHallPacketID_PACKET_SC_ENTERTHRIDGAME), pack)
			logger.Trace(pack)
		}
		//正在请求刷新余额中不能进入三方
		if p.thridBalanceRefreshReqing {
			logger.Logger.Warn("CSEnterThridGameHandler client req thridBalanceRefreshReqing", p.SnId)
			returnErrorCodeFunc(gamehall.OpResultCode_Game_OPRC_ThirdPltProcessing_Game)
			return nil
		}
		//msg.ThridGameId=proto.Int32(430010001)
		if p.thrscene != nil {
			logger.Logger.Infof("Player %v in game.", p.SnId)
			returnErrorCodeFunc(gamehall.OpResultCode_Game_OPRC_ThirdPltProcessing_Game)
			return nil
		}
		if p.isDelete {
			returnErrorCodeFunc(gamehall.OpResultCode_Game_OPRC_RoomHadClosed_Game)
			return nil
		}
		//pt := PlatformMgrSington.GetPackageTag(p.PackageID)
		//if pt != nil && pt.IsForceBind == 1 {
		//	if p.BeUnderAgentCode == "" || p.BeUnderAgentCode == "0" {
		//		returnErrorCodeFunc(gamehall.OpResultCode_Game_OPRC_MustBindPromoter_Game)
		//		return nil
		//	}
		//}

		ok, thridPltGameItem := ThirdPltGameMappingConfig.FindByGameID(msg.GetThridGameId())
		if !ok {
			logger.Logger.Infof("Player %v thridgame id err.", p.SnId, msg.GetThridGameId())
			returnErrorCodeFunc(gamehall.OpResultCode_Game_OPRC_Error_Game)
			return nil
		}
		thridGameId := thridPltGameItem.GetThirdGameID()

		//检测房间状态是否开启
		gps := PlatformMgrSington.GetGameConfig(p.Platform, msg.GetThridGameId())
		if gps == nil {
			logger.Logger.Infof("Player %v no cfg room close", p.SnId)

			returnErrorCodeFunc(gamehall.OpResultCode_Game_OPRC_RoomHadClosed_Game)
			return nil
		}
		dbGameFree := gps.DbGameFree
		if dbGameFree == nil {
			logger.Logger.Infof("Player %v no gamefree", p.SnId)
			returnErrorCodeFunc(gamehall.OpResultCode_Game_OPRC_RoomHadClosed_Game)
			return nil
		}

		//找到对应的平台
		v, ok := webapi.ThridPlatformMgrSington.ThridPlatformMap.Load(thridPltGameItem.GetThirdPlatformName())
		if !ok {
			logger.Logger.Infof("Player %v no platform", p.SnId)

			returnErrorCodeFunc(gamehall.OpResultCode_Game_OPRC_RoomHadClosed_Game)
			return nil
		}

		plt := v.(webapi.IThirdPlatform)
		pfConfig := PlatformMgrSington.GetPlatform(p.Platform)
		if pfConfig == nil || pfConfig.ThirdGameMerchant == nil || pfConfig.ThirdGameMerchant[int32(plt.GetPlatformBase().BaseGameID)] == 0 {
			logger.Logger.Infof("Player %v no pfcfg", p.SnId)

			//	returnErrorCodeFunc(gamehall.OpResultCode_Game_OPRC_RoomHadClosed_Game)
			//	return nil
		}

		//检查限额，金额不足
		if dbGameFree.GetLimitCoin() != 0 && p.GetCoin() < int64(dbGameFree.GetLimitCoin()) {
			returnErrorCodeFunc(gamehall.OpResultCode_Game_OPRC_CoinNotEnough_Game)
			return nil
		}

		//三方不检查游戏次数限制
		//todayData, _ := p.GetDaliyGameData(int(msg.GetThridGameId()))
		//if dbGameFree.GetPlayNumLimit() != 0 &&
		//	todayData != nil &&
		//	todayData.GameTimes >= int64(dbGameFree.GetPlayNumLimit()) {
		//	logger.Logger.Infof("CSEnterThridGameHandler player snid:%v todayData.GameTimes:%v>PlayNumLimit:%v then kick", p.SnId, todayData.GameTimes, dbGameFree.GetPlayNumLimit())
		//	returnErrorCodeFunc(gamehall.OpResultCode_OPRC_RoomGameTimes)
		//	return nil
		//}

		//检查平台配额是否足够
		/*if plt.GetPlatformBase().IsNeedCheckQuota {
			dgQuota := ThirdPlatformMgrSington.GetThirdPlatformCoin(p.Platform, plt.GetPlatformBase().Tag)
			if dgQuota <= 0 || dgQuota <= p.GetCoin() {
				logger.Logger.Infof("Player snid %v %v platfrom Quota of game not enough.", p.SnId, plt.GetPlatformBase().Name)
				returnErrorCodeFunc(gamehall.OpResultCode_Game_OPRC_Dg_QuotaNotEnough_Game)
				return nil
			}
		}*/

		//检查场景是否开放或者存在，预设数据
		scene := SceneMgrSington.GetThirdScene(plt)
		if scene != nil {
			p.thrscene = scene
		} else {
			logger.Logger.Infof("Player %v no scene", p.SnId)

			returnErrorCodeFunc(gamehall.OpResultCode_Game_OPRC_RoomHadClosed_Game)
			return nil
		}

		pack := &gamehall.SCEnterThridGame{}
		pack.ThridGameId = msg.ThridGameId
		amount := p.GetCoin()
		if plt.GetPlatformBase().TransferInteger {
			amount = (amount / 100) * 100
		}
		p.Coin = p.GetCoin() - amount
		gainway := common.GainWay_Transfer_System2Thrid
		oper := "System2" + plt.GetPlatformBase().Name
		timeStamp := time.Now().UnixNano()
		p.thridBalanceRefreshReqing = true
		transferTimeOut := false
		task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
			var err error
			var ok bool
			url := ""
			remark := "转入" + plt.GetPlatformBase().Name + thridPltGameItem.GetDesc()
			//thridPlatformCoin := int64(0)
			var coinLog *model.PayCoinLog
			var coinlogex *model.CoinLog
			/*err = ensureThridPltUserName(plt, p, amount, s.RemoteAddr())
			if err != nil && err != webapi.ErrNoCreated {
				goto Rollback
			}*/
			//ok = utils.RunPanicless(func() {
			err, url = enterThridPltUserName(plt, p, amount, thridGameId, s.RemoteAddr())
			//err, url = plt.ReqEnterGame(p.SnId, thridGameId, s.RemoteAddr(), p.Platform, p.Channel, p.Ip)
			//})
			ok = true
			if err != nil || !ok {
				logger.Logger.Errorf("plt.ReqEnterGame() snid:%v error: %v", p.SnId, err)
				if thrErr, ok := err.(webapi.ThirdError); ok {
					if thrErr.IsClose() {
						pack.OpRetCode = gamehall.OpResultCode_Game_OPRC_Thr_GameClose_Game
					} else {
						pack.OpRetCode = gamehall.OpResultCode_Game_OPRC_Error_Game
					}
				} else {
					pack.OpRetCode = gamehall.OpResultCode_Game_OPRC_Error_Game
				}
				goto Rollback
			}
			coinLog = model.NewPayCoinLog(timeStamp, int32(p.SnId), -amount, int32(gainway), oper, model.PayCoinLogType_Coin, 0)
			timeStamp = coinLog.TimeStamp
			err = model.InsertPayCoinLogs(p.Platform, coinLog)
			if err != nil {
				goto Rollback
			}
			coinlogex = model.NewCoinLogEx(p.SnId, -amount, p.Coin, p.SafeBoxCoin, 0, int32(gainway),
				0, oper, remark, p.Platform, p.Channel, p.BeUnderAgentCode, 0, p.PackageID, int32(plt.GetPlatformBase().VultGameID))
			err = model.InsertCoinLog(coinlogex)
			if err != nil {
				goto Rollback
			}
			//
			/*err, transferTimeOut = plt.ReqTransfer(p.SnId, amount, strconv.FormatInt(timeStamp, 10), p.Platform, p.Channel, p.Ip)
			if err != nil || transferTimeOut {
				logger.Logger.Errorf("plt.ReqTransfer() snid:%v error:%v", p.SnId, err)
				pack.OpRetCode = gamehall.OpResultCode_Game_OPRC_Error_Game
				goto Rollback
			}*/
			pack.OpRetCode = gamehall.OpResultCode_Game_OPRC_Sucess_Game
			pack.ScreenOrientationType = proto.Int32(thridPltGameItem.GetScreenOrientationType())
			pack.EnterUrl = proto.String(url)
			return nil
		Rollback:
			pack.OpRetCode = gamehall.OpResultCode_Game_OPRC_Error_Game
			if coinLog != nil {
				model.RemovePayCoinLog(p.Platform, coinLog.LogId)
			}
			if coinlogex != nil {
				if transferTimeOut {
					err2 := model.UpdateCoinLogRemark(coinlogex.Platform, coinlogex.LogId, plt.GetPlatformBase().Name+"需人工处理")
					if err2 != nil {
						logger.Logger.Errorf("thr UpdateCoinLogRemark(%v) error: %v", coinlogex.LogId, err2)
					}
				} else {
					model.RemoveCoinLogOne(coinlogex.Platform, coinlogex.LogId)
				}

			}
			return errors.New("system->third transfer rollback!")
		}), task.CompleteNotifyWrapper(func(data interface{}, tt task.Task) {
			if pack.GetOpRetCode() == gamehall.OpResultCode_Game_OPRC_Sucess_Game {
				// ThirdPlatformMgrSington.AddThirdPlatformCoin(p.Platform, plt.GetPlatformBase().Tag, -amount)
				p.SetPayTs(timeStamp)
			} else {
				//如帐变出现问题，就在日志里面查下面的输出信息！！！
				//如果转账超时，三方的转账是否成功就是未知的，这时不能将金币再加到玩家身上。
				//如果出现超时问题，就需要人工对账。
				//注意：这个地方说的超时已经包含CG工程Check订单后的超时
				if transferTimeOut {
					logger.Logger.Errorf("CSEnterThridGameHandler player snid:%v transfer %v to %v timeout:", p.SnId, amount, plt.GetPlatformBase().Name)
				} else {
					p.Coin += amount
				}
				p.thrscene = nil
				pack.OpRetCode = gamehall.OpResultCode_Game_OPRC_Error_Game
				logger.Logger.Trace("SCThridAccountTransferHandler system->third transfer error:", data)
			}
			p.dirty = true
			p.thirdBalanceRefreshMark[plt.GetPlatformBase().Name] = false

			p.SendDiffData()
			p.SendToClient(int(gamehall.GameHallPacketID_PACKET_SC_ENTERTHRIDGAME), pack)
			p.thridBalanceRefreshReqing = false
			logger.Logger.Trace("CSEnterThridGameHandler send client:", pack)
			return
		}), "CSEnterThridGameHandler").Start()
	}
	return nil
}
func ensureThridPltUserName(pltform webapi.IThirdPlatform, p *Player, amount int64, ip string) error {
	var err error
	err = pltform.ReqCreateAccount(p.SnId, p.Platform, p.Channel, p.GetIP())
	if err != nil {
		if err != webapi.ErrNoCreated {
			logger.Logger.Errorf("Snid=%v Plt=%v ReqCreateAccount error:%v", p.SnId, pltform.GetPlatformBase().Name, err)
		}
		return err
	}
	return nil
}

func enterThridPltUserName(pltform webapi.IThirdPlatform, p *Player, amount int64, gameId, ip string) (err error, url string) {
	// (snId int32, gameId string, clientIP string, platform, channel string, amount int64)
	err, url = pltform.ReqEnterGame(p.SnId, "", p.GetIP(), p.Platform, p.Channel, amount)
	if err != nil {
		if err != webapi.ErrNoCreated {
			logger.Logger.Errorf("Snid=%v Plt=%v ReqEnterGame error:%v", p.SnId, pltform.GetPlatformBase().Name, err)
		}
		return err, ""
	}
	return
}

// 离开三方
type CSLeaveThridGamePacketFactory struct {
}
type CSLeaveThridGameHandler struct {
}

func (this *CSLeaveThridGamePacketFactory) CreatePacket() interface{} {
	pack := &gamehall.CSLeaveThridGame{}
	return pack
}

func (this *CSLeaveThridGameHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSLeaveThridGameHandler Process recv ", data)
	if _, ok := data.(*gamehall.CSLeaveThridGame); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSLeaveThridGameHandler p == nil")
			return nil
		}
		_StartTransferThird2SystemTask(p)
		pack := &gamehall.SCLeaveThridGame{}
		pack.OpRetCode = gamehall.OpResultCode_Game_OPRC_Sucess_Game
		p.SendToClient(int(gamehall.GameHallPacketID_PACKET_SC_LEAVETHRIDGAME), pack)
	}
	return nil
}

type CSThridBalanceRefreshPacketFactory struct {
}
type CSThridBalanceRefreshHandler struct {
}

func (this *CSThridBalanceRefreshPacketFactory) CreatePacket() interface{} {
	pack := &gamehall.CSThridGameBalanceUpdate{}
	return pack
}
func (this *CSThridBalanceRefreshHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSThridBalanceRefreshHandler Process recv ", data)
	if _, ok := data.(*gamehall.CSThridGameBalanceUpdate); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSThridBalanceRefreshHandler p == nil")
			return nil
		}
		if p.scene != nil {
			logger.Logger.Tracef("player snid=%v CSThridBalanceRefreshHandler p.scene != nil return", p.SnId)
			return nil
		}

		//请求太快，不做处理，给API减轻一些压力
		if p.thridBalanceRefreshReqing {
			logger.Logger.Warn("CSThridBalanceRefreshHandler client req too fast")
			return nil
		}
		p.thridBalanceRefreshReqing = true
		gainway := common.GainWay_Transfer_Thrid2System
		waitgroup := webapi.ThridPlatformMgrSington.AllPlatformCount()
		isSucces := true
		timeout := false
		webapi.ThridPlatformMgrSington.ThridPlatformMap.Range(func(key, value interface{}) bool {
			plt := value.(webapi.IThirdPlatform)
			if stateOk, exist := p.thirdBalanceRefreshMark[plt.GetPlatformBase().Name]; exist && stateOk {
				waitgroup--
				if 0 == waitgroup {
					statePack := &gamehall.SCThridGameBalanceUpdateState{}
					pack := &gamehall.SCThridGameBalanceUpdate{}
					pack.OpRetCode = gamehall.OpResultCode_Game_OPRC_Sucess_Game
					p.thridBalanceReqIsSucces = true
					statePack.OpRetCode = gamehall.OpResultCode_Game_OPRC_Sucess_Game
					p.SendToClient(int(gamehall.GameHallPacketID_PACKET_SC_THRIDGAMEBALANCEUPDATESTATE), statePack)
					pack.Coin = proto.Int64(p.Coin)
					p.SendToClient(int(gamehall.GameHallPacketID_PACKET_SC_THRIDGAMEBALANCEUPDATE), pack)
					p.thridBalanceRefreshReqing = false
				}
				return true
			}
			var tsk task.Task
			tsk = task.New(nil,
				task.CallableWrapper(func(o *basic.Object) interface{} {
					plt := tsk.GetEnv("plt").(webapi.IThirdPlatform)
					tsk.PutEnv("timeStamp", time.Now().UnixNano())
					var err error
					var coinLog *model.PayCoinLog
					var coinlogex *model.CoinLog
					var apiHasTransfer = false
					remark := "刷新" + plt.GetPlatformBase().Name + "转出到系统"
					thirdBalance := int64(0)
					//allow := false
					if plt == nil {
						logger.Logger.Tracef("player snid=%v CSThridBalanceRefreshHandler plt == nil return", p.SnId)
						return int64(-1)
					}

					pfConfig := PlatformMgrSington.GetPlatform(p.Platform)
					if pfConfig == nil {
						return int64(0)
					}

					if pfConfig.ThirdGameMerchant == nil || pfConfig.ThirdGameMerchant[int32(plt.GetPlatformBase().BaseGameID)] == 0 {
						return int64(0)
					}

					oper := plt.GetPlatformBase().Name + "2System"
					err = ensureThridPltUserName(plt, p, thirdBalance, s.RemoteAddr())
					if err != nil {
						if err == webapi.ErrNoCreated {
							return int64(0)
						}
						logger.Logger.Tracef("player snid=%v at %v ensureThridPltUserName() err: %v", p.SnId, plt.GetPlatformBase().Name, err)
						goto Rollback
					}
					//allow = plt.ReqIsAllowTransfer(p.SnId, p.Platform, p.Channel)
					//if !allow {
					//	logger.Logger.Tracef("player snid=%v at %v is not allow Transfer", p.SnId, plt.GetPlatformBase().Name)
					//	goto Rollback
					//}
					err, thirdBalance = plt.ReqUserBalance(p.SnId, p.Platform, p.Channel, p.Ip)
					if err != nil {
						logger.Logger.Tracef("player snid=%v at %v plt.ReqUserBalance() err: %v", p.SnId, plt.GetPlatformBase().Name, err)
						goto Rollback
					}
					if thirdBalance <= 0 {
						return int64(0)
					}
					if plt.GetPlatformBase().TransferInteger {
						thirdBalance = (thirdBalance / 100) * 100
						if thirdBalance <= 0 {
							return int64(0)
						}
					}
					err, timeout = plt.ReqTransfer(p.SnId, -thirdBalance, strconv.FormatInt(time.Now().UnixNano(), 10), p.Platform, p.Channel, p.Ip)
					if err != nil || timeout {
						logger.Logger.Tracef("player snid=%v at %v plt.ReqTransfer() err: %v", p.SnId, plt.GetPlatformBase().Name, err)
						goto Rollback
					}
					apiHasTransfer = true
					coinLog = model.NewPayCoinLog(time.Now().UnixNano(), int32(p.SnId), thirdBalance, int32(gainway), oper, model.PayCoinLogType_Coin, 0)
					tsk.PutEnv("timeStamp", coinLog.TimeStamp)
					err = model.InsertPayCoinLogs(p.Platform, coinLog)
					if err != nil {
						logger.Logger.Tracef("player snid=%v at %v model.InsertPayCoinLogs() err: %v", p.SnId, plt.GetPlatformBase().Name, err)
						goto Rollback
					}
					coinlogex = model.NewCoinLogEx(p.SnId, thirdBalance, p.Coin+thirdBalance, p.SafeBoxCoin, 0, int32(gainway),
						0, oper, remark, p.Platform, p.Channel, p.BeUnderAgentCode, 0, p.PackageID, int32(plt.GetPlatformBase().VultGameID))
					err = model.InsertCoinLog(coinlogex)
					if err != nil {
						logger.Logger.Tracef("player snid=%v at %v model.InsertCoinLogs() err: %v", p.SnId, plt.GetPlatformBase().Name, err)
						goto Rollback
					}
					tsk.PutEnv("plt", plt)
					return thirdBalance
				Rollback:
					if coinLog != nil {
						model.RemovePayCoinLog(p.Platform, coinLog.LogId)
					}
					if coinlogex != nil {
						model.RemoveCoinLogOne(coinlogex.Platform, coinlogex.LogId)
					}
					if timeout {
						logger.Logger.Errorf("player snid=%v CSThridBalanceRefreshHandler transfer %v to %v timeout!", p.SnId, -thirdBalance, plt.GetPlatformBase().Name)
						return int64(-1)
					}
					if apiHasTransfer {
						err, timeout = plt.ReqTransfer(p.SnId, thirdBalance, strconv.FormatInt(time.Now().UnixNano(), 10), p.Platform, p.Channel, p.Ip)
						if timeout {
							logger.Logger.Errorf("player snid=%v CSThridBalanceRefreshHandler transfer rollback %v to %v timeout!", p.SnId, thirdBalance, plt.GetPlatformBase().Name)
						}
					}
					return int64(-1)
				}),
				task.CompleteNotifyWrapper(func(data interface{}, tt task.Task) {
					thirdBalance := data.(int64)
					plt := tsk.GetEnv("plt").(webapi.IThirdPlatform)
					timeStamp := tsk.GetEnv("timeStamp").(int64)
					if thirdBalance < 0 {
						isSucces = false
						logger.Logger.Tracef("player snid=%v at platform=%v CSThridBalanceRefreshHandler third->system transfer fail", p.SnId, plt.GetPlatformBase().Name)
					} else if thirdBalance > 0 {
						p.thirdBalanceRefreshMark[plt.GetPlatformBase().Name] = true
						p.Coin += thirdBalance
						p.SetPayTs(timeStamp)
						//ThirdPlatformMgrSington.AddThirdPlatformCoin(p.Platform, plt.GetPlatformBase().Tag, thirdBalance)
						p.dirty = true
						logger.Logger.Tracef("player snid=%v at platform=%v CSThridBalanceRefreshHandler third->system transfer succes", p.SnId, plt.GetPlatformBase().Name)
					} else {
						p.thirdBalanceRefreshMark[plt.GetPlatformBase().Name] = true
					}
					if atomic.AddInt32(&waitgroup, -1) == 0 {
						p.diffData.Coin = -1
						p.SendDiffData()
						statePack := &gamehall.SCThridGameBalanceUpdateState{}
						pack := &gamehall.SCThridGameBalanceUpdate{}
						if isSucces {
							pack.OpRetCode = gamehall.OpResultCode_Game_OPRC_Sucess_Game
							p.thridBalanceReqIsSucces = true
							statePack.OpRetCode = gamehall.OpResultCode_Game_OPRC_Sucess_Game
						} else {
							pack.OpRetCode = gamehall.OpResultCode_Game_OPRC_Error_Game
							p.thridBalanceReqIsSucces = false
							statePack.OpRetCode = gamehall.OpResultCode_Game_OPRC_Error_Game
						}
						p.SendToClient(int(gamehall.GameHallPacketID_PACKET_SC_THRIDGAMEBALANCEUPDATESTATE), statePack)
						pack.Coin = proto.Int64(p.Coin)
						p.SendToClient(int(gamehall.GameHallPacketID_PACKET_SC_THRIDGAMEBALANCEUPDATE), pack)
						p.thridBalanceRefreshReqing = false
						logger.Logger.Tracef("SendToClient() player snid=%v at CSThridBalanceRefreshHandler() pack:%v", p.SnId, pack.String())
					}
				}), "ThridAccount")
			tsk.PutEnv("plt", value)
			tsk.Start()
			return true
		})
	}
	return nil
}

type CSGetPrivateRoomListPacketFactory struct {
}
type CSGetPrivateRoomListHandler struct {
}

func (this *CSGetPrivateRoomListPacketFactory) CreatePacket() interface{} {
	pack := &gamehall.CSGetPrivateRoomList{}
	return pack
}

func (this *CSGetPrivateRoomListHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSGetPrivateRoomListHandler Process recv ", data)
	p := PlayerMgrSington.GetPlayer(sid)
	if p == nil {
		logger.Logger.Warn("CSGetPrivateRoomListHandler p == nil")
		return nil
	}

	pps := PrivateSceneMgrSington.GetOrCreatePlayerPrivateScene(p)
	if pps == nil {
		logger.Logger.Warnf("CSGetPrivateRoomListHandler PrivateSceneMgrSington.GetOrCreatePlayerPrivateScene(%v)", p.SnId)
		return nil
	}

	pps.SendPrivateScenes(p)
	return nil
}

type CSGetPrivateRoomHistoryPacketFactory struct {
}
type CSGetPrivateRoomHistoryHandler struct {
}

func (this *CSGetPrivateRoomHistoryPacketFactory) CreatePacket() interface{} {
	pack := &gamehall.CSGetPrivateRoomHistory{}
	return pack
}

func (this *CSGetPrivateRoomHistoryHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSGetPrivateRoomHistoryHandler Process recv ", data)
	if msg, ok := data.(*gamehall.CSGetPrivateRoomHistory); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSGetPrivateRoomHistoryHandler p == nil")
			return nil
		}

		pps := PrivateSceneMgrSington.GetOrCreatePlayerPrivateScene(p)
		if pps == nil {
			logger.Logger.Warnf("CSGetPrivateRoomHistoryHandler PrivateSceneMgrSington.GetOrCreatePlayerPrivateScene(%v)", p.SnId)
			return nil
		}

		pps.LoadLogs(p, msg.GetQueryTime())
	}
	return nil
}

type CSDestroyPrivateRoomPacketFactory struct {
}
type CSDestroyPrivateRoomHandler struct {
}

func (this *CSDestroyPrivateRoomPacketFactory) CreatePacket() interface{} {
	pack := &gamehall.CSDestroyPrivateRoom{}
	return pack
}

func (this *CSDestroyPrivateRoomHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSDestroyPrivateRoomHandler Process recv ", data)
	if msg, ok := data.(*gamehall.CSDestroyPrivateRoom); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSDestroyPrivateRoomHandler p == nil")
			return nil
		}

		pps := PrivateSceneMgrSington.GetOrCreatePlayerPrivateScene(p)
		if pps == nil {
			logger.Logger.Warnf("CSDestroyPrivateRoomHandler PrivateSceneMgrSington.GetOrCreatePlayerPrivateScene(%v)", p.SnId)
			return nil
		}

		state := PrivateSceneState_Deleting //删除中
		scene := pps.GetScene(int(msg.GetRoomId()))
		if scene != nil {
			if !scene.deleting {
				scene.ForceDelete(true)
			}
		} else {
			state = PrivateSceneState_Deleted //已删除
		}
		pack := &gamehall.SCDestroyPrivateRoom{
			OpRetCode: gamehall.OpResultCode_Game_OPRC_Sucess_Game,
			RoomId:    msg.RoomId,
			State:     proto.Int(state),
		}
		proto.SetDefaults(pack)
		p.SendToClient(int(gamehall.GameHallPacketID_PACKET_SC_DESTROYPRIVATEROOM), pack)
	}
	return nil
}

type CSQueryRoomInfoPacketFactory struct {
}
type CSQueryRoomInfoHandler struct {
}

func (this *CSQueryRoomInfoPacketFactory) CreatePacket() interface{} {
	pack := &gamehall.CSQueryRoomInfo{}
	return pack
}

func (this *CSQueryRoomInfoHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSQueryRoomInfoHandler Process recv ", data)
	if msg, ok := data.(*gamehall.CSQueryRoomInfo); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSQueryRoomInfoHandler p == nil")
			return nil
		}
		pack := &gamehall.SCQueryRoomInfo{
			GameSite: proto.Int32(msg.GetGameSite()),
		}
		for _, gameid := range msg.GetGameIds() {
			pack.GameIds = append(pack.GameIds, gameid)
			scenes := SceneMgrSington.GetScenesByGame(int(gameid))
			for _, scene := range scenes {
				if scene != nil && scene.sceneMode == common.SceneMode_Public && len(scene.players) != 0 {
					if scene.gameId == int(gameid) && scene.gameSite == int(msg.GetGameSite()) {
						roomInfo := &gamehall.QRoomInfo{
							GameFreeId: proto.Int32(scene.dbGameFree.GetId()),
							GameId:     proto.Int32(scene.dbGameFree.GetGameId()),
							RoomId:     proto.Int(scene.sceneId),
							BaseCoin:   proto.Int32(scene.BaseScore),
							LimitCoin:  proto.Int32(scene.dbGameFree.GetLimitCoin()),
							CurrNum:    proto.Int(scene.GetPlayerCnt()),
							MaxPlayer:  proto.Int(scene.playerNum),
							Creator:    proto.Int32(scene.creator),
							CreateTs:   proto.Int32(int32(scene.createTime.Unix())),
						}
						pack.RoomInfo = append(pack.RoomInfo, roomInfo)
					}
				}
			}
		}
		pack.OpRetCode = gamehall.OpResultCode_Game_OPRC_Sucess_Game
		proto.SetDefaults(pack)
		p.SendToClient(int(gamehall.GameHallPacketID_PACKET_SC_QUERYROOMINFO), pack)
		logger.Logger.Trace("SCQueryRoomInfo: ", pack)
	}
	return nil
}

type CSLotteryLogPacketFactory struct {
}
type CSLotteryLogHandler struct {
}

func (this *CSLotteryLogPacketFactory) CreatePacket() interface{} {
	pack := &gamehall.CSLotteryLog{}
	return pack
}

func (this *CSLotteryLogHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSLotteryLogHandler Process recv ", data)
	//if msg, ok := data.(*gamehall.CSLotteryLog); ok {
	//	p := PlayerMgrSington.GetPlayer(sid)
	//	if p == nil {
	//		logger.Logger.Warn("CSLotteryLogHandler p == nil")
	//		return nil
	//	}
	//
	//	LotteryMgrSington.FetchLotteryLog(p, msg.GetGameFreeId())
	//}
	return nil
}

// 获取指定游戏配置 包括分场信息
type CSGetGameConfigPacketFactory struct {
}
type CSGetGameConfigHandler struct {
}

func (this *CSGetGameConfigPacketFactory) CreatePacket() interface{} {
	pack := &gamehall.CSGetGameConfig{}
	return pack
}

func (this *CSGetGameConfigHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSGetGameConfigHandler Process recv ", data)
	if msg, ok := data.(*gamehall.CSGetGameConfig); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSGetGameConfigHandler p == nil")
			return nil
		}

		plf := msg.GetPlatform()
		if plf == "" {
			logger.Logger.Warn("CSGetGameConfigHandler plf == ''")
			return nil
		}

		chl := msg.GetChannel()
		if chl == "" {
			logger.Logger.Warn("CSGetGameConfigHandler chl == ''")
			return nil
		}

		gameId := msg.GetGameId()
		if gameId == 0 {
			logger.Logger.Warn("CSGetGameConfigHandler gameId == 0")
			return nil
		}
		p.SendGameConfig(gameId, plf, chl)
	}
	return nil
}

// 进入游戏操作
type CSEnterGamePacketFactory struct {
}
type CSEnterGameHandler struct {
}

func (this *CSEnterGamePacketFactory) CreatePacket() interface{} {
	pack := &gamehall.CSEnterGame{}
	return pack
}

func (this *CSEnterGameHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSEnterGameHandler Process recv ", data)
	if msg, ok := data.(*gamehall.CSEnterGame); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p != nil {
			var ret gamehall.OpResultCode_Game
			pack := &gamehall.SCEnterGame{
				Id: msg.Id,
			}
			oldPlatform := p.Platform

			//匿名函数
			sendEnterGame := func() {
				//机器人要避免身上的平台标记被污染
				if p.IsRob {
					if ret != gamehall.OpResultCode_Game_OPRC_Sucess_Game {
						p.Platform = oldPlatform
					}
				}
				pack.OpCode = ret
				proto.SetDefaults(pack)
				p.SendToClient(int(gamehall.GameHallPacketID_PACKET_SC_ENTERGAME), pack)

				if ret == gamehall.OpResultCode_Game_OPRC_Sucess_Game && p.scene != nil {
					gameName := p.scene.dbGameFree.GetName() + p.scene.dbGameFree.GetTitle()
					ActMonitorMgrSington.SendActMonitorEvent(ActState_Game, p.SnId, p.Name, p.Platform,
						0, 0, gameName, 0)
				}
			}

			//pt := PlatformMgrSington.GetPackageTag(p.PackageID)
			//if pt != nil && pt.IsForceBind == 1 {
			//	if p.BeUnderAgentCode == "" || p.BeUnderAgentCode == "0" {
			//		ret = gamehall.OpResultCode_Game_OPRC_MustBindPromoter_Game
			//		sendEnterGame()
			//		return nil
			//	}
			//}
			if p.scene != nil {
				logger.Logger.Warnf("CSEnterGameHandler  found snid:%v had in scene:%v gameid:%v", p.SnId, p.scene.sceneId, p.scene.gameId)
				p.ReturnScene(true)
				return nil
			}
			//检测房间状态是否开启
			gameId := msg.GetId() //gameid
			if gameId > 10000 {
				gameId = msg.GetId() / 10000 //gamefreeid
			}
			//tienlen游戏入场规则
			if common.IsLocalGame(int(gameId)) {
				playerTakeCoin := p.Coin
				gameSite := srvdata.CreateRoomMgrSington.GetGameSiteByGameId(gameId, playerTakeCoin)
				if gameSite == 0 {
					ret = gamehall.OpResultCode_Game_OPRC_CoinNotEnough_Game
					sendEnterGame()
					return nil
				}
				gamefreeid := gameId*10000 + gameSite
				gps := PlatformMgrSington.GetGameConfig(p.Platform, gamefreeid)
				if gps == nil {
					ret = gamehall.OpResultCode_Game_OPRC_RoomHadClosed_Game
					sendEnterGame()
					return nil
				}

				dbGameFree := gps.DbGameFree
				var roomId int32
				params := msg.GetOpParams()

				if dbGameFree == nil {
					ret = gamehall.OpResultCode_Game_OPRC_RoomHadClosed_Game
					sendEnterGame()
					return nil
				}

				if len(params) != 0 {
					roomId = params[0]
					platformName := CoinSceneMgrSington.GetPlatformBySceneId(int(roomId))
					if p.IsRob {
						p.Platform = platformName
					} else if p.GMLevel > 0 && p.Platform == platformName { //允许GM直接按房间ID进场
						roomId = params[0]
					}
				}
				if len(msg.GetPlatform()) > 0 && p.IsRob {
					p.Platform = msg.GetPlatform()
				}

				if len(params) != 0 && (p.GMLevel > 0 || dbGameFree.GetCreateRoomNum() != 0) { //允许GM|或者可选房间的游戏直接按房间ID进场
					s := SceneMgrSington.GetScene(int(params[0]))
					if s != nil {
						roomId = params[0]
					}
				}

				excludeSceneIds := p.lastSceneId[gamefreeid]
				ret = gamehall.OpResultCode_Game(CoinSceneMgrSington.PlayerEnterLocalGame(p, gamefreeid, roomId, excludeSceneIds, true))
				if p.scene != nil {
					pack.OpParams = append(pack.OpParams, int32(p.scene.sceneId))
					//TODO 有房间还进入失败，尝试returnroom
					if ret != gamehall.OpResultCode_Game_OPRC_Sucess_Game {
						p.ReturnScene(true)
					}
				}
				sendEnterGame()
				return nil
			}
			gps := PlatformMgrSington.GetGameConfig(p.Platform, msg.GetId())
			if gps == nil {
				ret = gamehall.OpResultCode_Game_OPRC_RoomHadClosed_Game
				sendEnterGame()
				return nil
			}

			dbGameFree := gps.DbGameFree
			gameType := dbGameFree.GetGameType()
			var roomId int32
			params := msg.GetOpParams()

			if dbGameFree == nil {
				ret = gamehall.OpResultCode_Game_OPRC_RoomHadClosed_Game
				sendEnterGame()
				return nil
			}

			if dbGameFree.GetLimitCoin() != 0 && int64(dbGameFree.GetLimitCoin()) > p.Coin {
				ret = gamehall.OpResultCode_Game_OPRC_CoinNotEnough_Game
				sendEnterGame()
				return nil
			}

			//if dbGameFree.GetMaxCoinLimit() != 0 && int64(dbGameFree.GetMaxCoinLimit()) < p.Coin && !p.IsRob {
			//	ret = gamehall.OpResultCode_Game_OPRC_CoinTooMore_Game
			//	sendEnterGame()
			//	return nil
			//}

			//检查游戏次数限制
			if !p.IsRob {
				todayData, _ := p.GetDaliyGameData(int(dbGameFree.GetId()))
				if dbGameFree.GetPlayNumLimit() != 0 &&
					todayData != nil &&
					todayData.GameTimes >= int64(dbGameFree.GetPlayNumLimit()) {
					ret = gamehall.OpResultCode_Game_OPRC_RoomGameTimes_Game
					sendEnterGame()
					return nil
				}
			}

			if common.IsHundredType(gameType) {
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

				if gps.GroupId != 0 {
					if len(params) != 0 && p.GMLevel > 0 { //允许GM直接按房间ID进场
						s := SceneMgrSington.GetScene(int(params[0]))
						if s != nil && s.groupId == gps.GroupId {
							roomId = params[0]
						}
					}
				}

				ret = gamehall.OpResultCode_Game(HundredSceneMgrSington.PlayerEnter(p, msg.GetId()))
				if p.scene != nil {
					pack.OpParams = append(pack.OpParams, msg.GetId())
					//TODO 有房间还进入失败，尝试returnroom
					if ret != gamehall.OpResultCode_Game_OPRC_Sucess_Game {
						p.ReturnScene(true)
						return nil
					}
				}
			} else if common.IsCoinSceneType(gameType) {
				if len(params) != 0 {
					roomId = params[0]
					platformName := CoinSceneMgrSington.GetPlatformBySceneId(int(roomId))
					if p.IsRob {
						p.Platform = platformName
					} else if p.GMLevel > 0 && p.Platform == platformName { //允许GM直接按房间ID进场
						roomId = params[0]
					}
				}
				if len(msg.GetPlatform()) > 0 && p.IsRob {
					p.Platform = msg.GetPlatform()
				}

				if len(params) != 0 && (p.GMLevel > 0 || dbGameFree.GetCreateRoomNum() != 0) { //允许GM|或者可选房间的游戏直接按房间ID进场
					s := SceneMgrSington.GetScene(int(params[0]))
					if s != nil && s.groupId == gps.GroupId {
						roomId = params[0]
					}
				}

				excludeSceneIds := p.lastSceneId[msg.GetId()]
				ret = gamehall.OpResultCode_Game(CoinSceneMgrSington.PlayerEnter(p, msg.GetId(), roomId, excludeSceneIds, true))
				if p.scene != nil {
					pack.OpParams = append(pack.OpParams, int32(p.scene.sceneId))
					//TODO 有房间还进入失败，尝试returnroom
					if ret != gamehall.OpResultCode_Game_OPRC_Sucess_Game {
						p.ReturnScene(true)
					}
				}
			}

			sendEnterGame()
		}
	}
	return nil
}

// 退出游戏操作
type CSQuitGamePacketFactory struct {
}
type CSQuitGameHandler struct {
}

func (this *CSQuitGamePacketFactory) CreatePacket() interface{} {
	pack := &gamehall.CSQuitGame{}
	return pack
}

func (this *CSQuitGameHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSQuitGameHandler Process recv ", data)
	if msg, ok := data.(*gamehall.CSQuitGame); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p != nil {
			var ret gamehall.OpResultCode_Game
			pack := &gamehall.SCQuitGame{
				Id: msg.Id,
			}

			dbGameFree := srvdata.PBDB_GameFreeMgr.GetData(msg.GetId())
			gameType := dbGameFree.GetGameType()
			if common.IsHundredType(gameType) {
				ret = gamehall.OpResultCode_Game(HundredSceneMgrSington.PlayerTryLeave(p))
			} else if common.IsCoinSceneType(gameType) {
				ret = gamehall.OpResultCode_Game(CoinSceneMgrSington.PlayerTryLeave(p, msg.GetId(), msg.IsAudience))
				if gamehall.OpResultCode_Game_OPRC_OpYield_Game == ret {
					return nil
				}
			}

			pack.OpCode = ret
			proto.SetDefaults(pack)
			p.SendToClient(int(gamehall.GameHallPacketID_PACKET_SC_QUITGAME), pack)

		}
	}
	return nil
}

type CSCreateRoomPacketFactory struct {
}
type CSCreateRoomHandler struct {
}

func (this *CSCreateRoomPacketFactory) CreatePacket() interface{} {
	pack := &gamehall.CSCreateRoom{}
	return pack
}

func (this *CSCreateRoomHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSCreateRoomHandler Process recv ", data)
	if msg, ok := data.(*gamehall.CSCreateRoom); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			return nil
		}

		var code gamehall.OpResultCode_Game
		var betRange []int32
		var inRange bool
		var dbCreateRoom *server.DB_Createroom
		var dbGameFree *server.DB_GameFree
		var dbGameRule *server.DB_GameRule
		var playerTakeCoin = p.GetCoin()
		var maxPlayerNum int
		var gameId = msg.GetGameId()
		var params = msg.GetParams()
		var roomId int
		var scene *Scene
		var sp ScenePolicy
		var gamefreeId int32
		var gameSite int32
		var csp *CoinScenePool
		var baseScore int32
		var gps *webapi_proto.GameFree

		//根据携带金额取可创房间 DB_Createroom
		arrs := srvdata.PBDB_CreateroomMgr.Datas.Arr
		for i := len(arrs) - 1; i >= 0; i-- {
			if arrs[i].GetGameId() == msg.GetGameId() {
				goldRange := arrs[i].GoldRange
				if len(goldRange) == 0 {
					continue
				}
				if playerTakeCoin >= int64(goldRange[0]) {
					dbCreateRoom = arrs[i]
					break
				}
			}
		}
		if dbCreateRoom == nil {
			code = gamehall.OpResultCode_Game_OPRC_GameNotExist_Game
			logger.Logger.Tracef("CSCreateRoomHandler SnId:%v GameFreeId:%v not exist", p.SnId, dbCreateRoom)
			goto failed
		}

		//校验下gameid
		if dbCreateRoom.GetGameId() != gameId {
			code = gamehall.OpResultCode_Game_OPRC_GameNotExist_Game
			logger.Logger.Tracef("CSCreateRoomHandler PBDB_Createroom Id:%v GameId:%v not exist", dbCreateRoom.GetGameId(), gameId)
			goto failed
		}

		gameSite = dbCreateRoom.GetGameSite()
		gamefreeId = gameId*10000 + gameSite
		gps = PlatformMgrSington.GetGameConfig(p.Platform, gamefreeId)
		if gps == nil {
			code = gamehall.OpResultCode_Game_OPRC_GameNotExist_Game
			logger.Logger.Tracef("CSCreateRoomHandler SnId:%v GameFreeId:%v not exist", p.SnId, gamefreeId)
			goto failed
		}

		dbGameFree = gps.DbGameFree
		//dbGameFree = srvdata.PBDB_GameFreeMgr.GetData(gamefreeId)
		if dbGameFree == nil {
			code = gamehall.OpResultCode_Game_OPRC_GameNotExist_Game
			logger.Logger.Tracef("CSCreateRoomHandler SnId:%v GameFreeId:%v not exist", p.SnId, gamefreeId)
			goto failed
		}

		dbGameRule = srvdata.PBDB_GameRuleMgr.GetData(dbGameFree.GetGameRule())
		if dbGameRule == nil {
			code = gamehall.OpResultCode_Game_OPRC_GameNotExist_Game
			logger.Logger.Tracef("CSCreateRoomHandler SnId:%v GameFreeId:%v gamerule not exist", p.SnId, gamefreeId)
			goto failed
		}

		params = common.CopySliceInt32(dbGameRule.GetParams())
		sp = GetScenePolicy(int(dbGameFree.GetGameId()), int(dbGameFree.GetGameMode()))
		if sp == nil {
			code = gamehall.OpResultCode_Game_OPRC_GameNotExist_Game
			logger.Logger.Tracef("CSCreateRoomHandler SnId:%v ScenePolicy(gameid:%v mode:%v) not registe", p.SnId, dbGameFree.GetGameId(), dbGameFree.GetGameMode())
			goto failed
		}

		if p.scene != nil {
			code = gamehall.OpResultCode_Game_OPRC_RoomHadExist_Game
			logger.Logger.Tracef("CSCreateRoomHandler had scene(%d)", p.scene.sceneId)
			goto failed
		}

		//校验底分
		betRange = dbCreateRoom.GetBetRange()
		inRange = false
		for _, bet := range betRange {
			if msg.GetBaseCoin() == bet {
				inRange = true
				break
			}
		}
		if !inRange || msg.GetBaseCoin() == 0 {
			code = gamehall.OpResultCode_Game_OPRC_Error_Game
			logger.Logger.Tracef("CSCreateRoomHandler BaseCoin:%v not in BetRange ", msg.GetBaseCoin())
			goto failed
		}
		//修正底注
		baseScore = msg.GetBaseCoin()

		//校验人数
		maxPlayerNum = int(msg.GetMaxPlayerNum())
		if maxPlayerNum != 4 && maxPlayerNum != 2 {
			code = gamehall.OpResultCode_Game_OPRC_Error_Game
			logger.Logger.Tracef("CSCreateRoomHandler GameId_TienLen_maxPlayerNum:%v ", maxPlayerNum)
			goto failed
		}

		//创建房间
		csp = CoinSceneMgrSington.GetCoinSceneMgr(p, dbGameFree)
		roomId = SceneMgrSington.GenOneCoinSceneId()
		if roomId == common.RANDID_INVALID {
			code = gamehall.OpResultCode_Game_OPRC_AllocRoomIdFailed_Game
			logger.Logger.Tracef("CSCreateRoomHandler SnId:%v GameId:%v sceneId == -1 ", p.SnId, gameId)
			goto failed
		}
		scene, code = p.CreateLocalGameScene(roomId, int(gameId), int(gameSite), int(msg.GetSceneMode()), maxPlayerNum, params, dbGameFree, baseScore)
		if scene != nil {
			if code == gamehall.OpResultCode_Game_OPRC_Sucess_Game {
				logger.Logger.Tracef("CSCreateRoomHandler SnId:%v Create Sucess GameId:%v", p.SnId, gameId)
				// try enter scene
				csp.scenes[scene.sceneId] = scene
				if p.EnterScene(scene, true, -1) {
					CoinSceneMgrSington.OnPlayerEnter(p, dbGameFree.Id)
				} else {
					code = gamehall.OpResultCode_Game_OPRC_Error_Game
				}
			}
		}

	failed:
		resp := &gamehall.SCCreateRoom{
			GameId:       msg.GetGameId(),
			BaseCoin:     msg.GetBaseCoin(),
			SceneMode:    msg.GetSceneMode(),
			MaxPlayerNum: msg.GetMaxPlayerNum(),
			Params:       msg.GetParams(),
			OpRetCode:    code,
		}
		proto.SetDefaults(resp)
		p.SendToClient(int(gamehall.GameHallPacketID_PACKET_SC_CREATEROOM), resp)
	}
	return nil
}

// 观众坐下
type CSAudienceSitPacketFactory struct {
}
type CSAudienceSitHandler struct {
}

func (this *CSAudienceSitPacketFactory) CreatePacket() interface{} {
	pack := &gamehall.CSAudienceSit{}
	return pack
}

func (this *CSAudienceSitHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSAudienceSitHandler Process recv ", data)
	if msg, ok := data.(*gamehall.CSAudienceSit); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSAudienceSitHandler p == nil")
			return nil
		}
		scene := SceneMgrSington.GetSceneByPlayerId(p.SnId)
		if scene == nil {
			logger.Logger.Warn("CSAudienceSitHandler scene == nil")
			return nil
		}
		if int32(scene.sceneId) != msg.GetRoomId() {
			logger.Logger.Warn("CSAudienceSitHandler sceneId != roomId")
			return nil
		}
		newPlayer := PlayerMgrSington.GetPlayerBySnId(p.SnId)
		if newPlayer != nil {
			pack := &gamehall.SCAudienceSit{
				RoomId: proto.Int32(msg.GetRoomId()),
			}
			if !scene.IsTestScene() {
				// 入场限额检查
				limitCoin := srvdata.CreateRoomMgrSington.GetLimitCoinByBaseScore(int32(scene.gameId), int32(scene.gameSite), scene.BaseScore)
				if p.Coin < limitCoin {
					pack.OpCode = gamehall.OpResultCode_Game_OPRC_MoneyNotEnough_Game
					newPlayer.SendToClient(int(gamehall.GameHallPacketID_PACKET_SC_AUDIENCESIT), pack)
					return nil
				}
			}
			// 是否还有空座位
			if scene.IsFull() {
				pack.OpCode = gamehall.OpResultCode_Game_OPRC_RoomIsFull_Game
				newPlayer.SendToClient(int(gamehall.GameHallPacketID_PACKET_SC_AUDIENCESIT), pack)
				return nil
			}
			scene.AudienceSit(newPlayer, -1)
			pack.OpCode = gamehall.OpResultCode_Game_OPRC_Sucess_Game
			newPlayer.SendToClient(int(gamehall.GameHallPacketID_PACKET_SC_AUDIENCESIT), pack)
		}
	}
	return nil
}

// 我的游戏信息及平台公告
type CSRecordAndNoticePacketFactory struct {
}
type CSRecordAndNoticeHandler struct {
}

func (this *CSRecordAndNoticePacketFactory) CreatePacket() interface{} {
	pack := &gamehall.CSRecordAndNotice{}
	return pack
}

func (this *CSRecordAndNoticeHandler) Process(s *netlib.Session, packetid int, data interface{}, sid int64) error {
	logger.Logger.Trace("CSRecordAndNoticeHandler Process recv ", data)
	if msg, ok := data.(*gamehall.CSRecordAndNotice); ok {
		p := PlayerMgrSington.GetPlayer(sid)
		if p == nil {
			logger.Logger.Warn("CSRecordAndNoticeHandler p == nil")
			return nil
		}

		pack := &gamehall.SCRecordAndNotice{
			OpCode: gamehall.OpResultCode_Game_OPRC_Sucess_Game,
		}
		if msg.Opt == 0 {
			list := PlatformMgrSington.GetCommonNotice(p.Platform)
			if list != nil {
				for _, v := range list.List {
					pack.List = append(pack.List, &gamehall.CommonNotice{
						Sort:      v.Sort,             // 排序 第几位
						Title:     v.Title,            // 标题
						Content:   v.Content,          // 内容
						TypeName:  v.TypeName,         // 大标题
						Type:      v.Type,             // 大标题类型
						StartTime: int32(v.StartTime), // 开始显示时间
						EndTime:   int32(v.EndTime),   // 结束显示时间
					})
				}
			}
			p.SendToClient(int(gamehall.GameHallPacketID_PACKET_SC_COMNOTICE), pack)
			return nil
		}
		if msg.Opt == 1 {

			task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
				gdi := model.GetPlayerExistListByTs(p.SnId, p.Platform, 7)
				return &gdi
			}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
				if data != nil && data.(*[]int64) != nil {
					gamets := data.(*[]int64)
					pack.GlistTs = append(pack.GlistTs, *gamets...)

				} else {
					pack.OpCode = gamehall.OpResultCode_Game_OPRC_Error_Game
				}
				p.SendToClient(int(gamehall.GameHallPacketID_PACKET_SC_COMNOTICE), pack)
			}), "SCRecordAndNotice").Start()
		} else if msg.Opt == 2 {

			starttime := msg.StartTime
			time := time.Unix(starttime, 0)
			endtime := time.AddDate(0, 0, 1).Unix() - 1 //23:59:59
			// starttime := time.AddDate(0, 0, 1).Unix()
			task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
				gdi := model.GetPlayerListByHallExAPI(p.SnId, p.Platform, starttime, endtime, int(msg.PageNo), int(msg.PageSize))
				return &gdi
			}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
				if data != nil && data.(*model.GamePlayerListType) != nil {
					gamedata := data.(*model.GamePlayerListType)
					for _, v := range gamedata.Data {
						pack.Glist = append(pack.Glist, &gamehall.PlayerRecord{
							GameFreeid:        v.GameFreeid,
							GameDetailedLogId: v.GameDetailedLogId,
							TotalIn:           v.TotalIn,
							TotalOut:          v.TotalOut,
							Ts:                v.Ts,
						})
					}
				} else {
					pack.OpCode = gamehall.OpResultCode_Game_OPRC_Error_Game
				}
				p.SendToClient(int(gamehall.GameHallPacketID_PACKET_SC_COMNOTICE), pack)
			}), "SCRecordAndNotice").Start()
		}
	}
	return nil
}

func init() {
	common.RegisterHandler(int(gamehall.GameHallPacketID_PACKET_CS_CREATEPRIVATEROOM), &CSCreatePrivateRoomHandler{})
	netlib.RegisterFactory(int(gamehall.GameHallPacketID_PACKET_CS_CREATEPRIVATEROOM), &CSCreatePrivateRoomPacketFactory{})
	common.RegisterHandler(int(gamehall.GameHallPacketID_PACKET_CS_DESTROYPRIVATEROOM), &CSDestroyPrivateRoomHandler{})
	netlib.RegisterFactory(int(gamehall.GameHallPacketID_PACKET_CS_DESTROYPRIVATEROOM), &CSDestroyPrivateRoomPacketFactory{})
	common.RegisterHandler(int(gamehall.GameHallPacketID_PACKET_CS_AUDIENCE_ENTERROOM), &CSAudienceEnterRoomHandler{})
	netlib.RegisterFactory(int(gamehall.GameHallPacketID_PACKET_CS_AUDIENCE_ENTERROOM), &CSEnterRoomPacketFactory{})
	common.RegisterHandler(int(gamehall.GameHallPacketID_PACKET_CS_RETURNROOM), &CSReturnRoomHandler{})
	netlib.RegisterFactory(int(gamehall.GameHallPacketID_PACKET_CS_RETURNROOM), &CSReturnRoomPacketFactory{})
	//common.RegisterHandler(int(gamehall.GameHallPacketID_PACKET_CS_INVITEROBOT), &CSInviteRobotHandler{})
	//netlib.RegisterFactory(int(gamehall.GameHallPacketID_PACKET_CS_INVITEROBOT), &CSInviteRobotPacketFactory{})

	common.RegisterHandler(int(message.MSGPacketID_PACKET_CS_MESSAGEREAD), &CSReadMessageHandler{})
	netlib.RegisterFactory(int(message.MSGPacketID_PACKET_CS_MESSAGEREAD), &CSReadMessagePacketFactory{})
	common.RegisterHandler(int(message.MSGPacketID_PACKET_CS_MESSAGEDEL), &CSDelMessageHandler{})
	netlib.RegisterFactory(int(message.MSGPacketID_PACKET_CS_MESSAGEDEL), &CSDelMessagePacketFactory{})
	common.RegisterHandler(int(message.MSGPacketID_PACKET_CS_GETMESSAGEATTACH), &CSGetMessageAttachHandler{})
	netlib.RegisterFactory(int(message.MSGPacketID_PACKET_CS_GETMESSAGEATTACH), &CSGetMessageAttachPacketFactory{})
	common.RegisterHandler(int(message.MSGPacketID_PACKET_CS_MESSAGELIST), &SCMessageListHandler{})
	netlib.RegisterFactory(int(message.MSGPacketID_PACKET_CS_MESSAGELIST), &SCMessageListPacketFactory{})

	common.RegisterHandler(int(gamehall.GameHallPacketID_PACKET_CS_JOINGAME), &CSJoinGameHandler{})
	netlib.RegisterFactory(int(gamehall.GameHallPacketID_PACKET_CS_JOINGAME), &CSJoinGamePacketFactory{})
	common.RegisterHandler(int(player.PlayerPacketID_PACKET_CS_GETDATALOG), &CSGetDataLogHandler{})
	netlib.RegisterFactory(int(player.PlayerPacketID_PACKET_CS_GETDATALOG), &CSGetDataLogPacketFactory{})

	common.RegisterHandler(int(gamehall.GameHallPacketID_PACKET_CS_ENTERHALL), &CSEnterHallHandler{})
	netlib.RegisterFactory(int(gamehall.GameHallPacketID_PACKET_CS_ENTERHALL), &CSEnterHallPacketFactory{})

	common.RegisterHandler(int(gamehall.GameHallPacketID_PACKET_CS_LEAVEHALL), &CSLeaveHallHandler{})
	netlib.RegisterFactory(int(gamehall.GameHallPacketID_PACKET_CS_LEAVEHALL), &CSLeaveHallPacketFactory{})

	common.RegisterHandler(int(gamehall.GameHallPacketID_PACKET_CS_HALLROOMLIST), &CSHallRoomListHandler{})
	netlib.RegisterFactory(int(gamehall.GameHallPacketID_PACKET_CS_HALLROOMLIST), &CSHallRoomListPacketFactory{})
	//DG
	common.RegisterHandler(int(gamehall.GameHallPacketID_PACKET_CS_ENTERDGGAME), &CSEnterDgGameHandler{})
	netlib.RegisterFactory(int(gamehall.GameHallPacketID_PACKET_CS_ENTERDGGAME), &CSEnterDgGamePacketFactory{})
	common.RegisterHandler(int(gamehall.GameHallPacketID_PACKET_CS_LEAVEDGGAME), &CSLeaveDgGameHandler{})
	netlib.RegisterFactory(int(gamehall.GameHallPacketID_PACKET_CS_LEAVEDGGAME), &CSLeaveDgGamePacketFactory{})

	common.RegisterHandler(int(gamehall.GameHallPacketID_PACKET_CS_ENTERTHRIDGAME), &CSEnterThridGameHandler{})
	netlib.RegisterFactory(int(gamehall.GameHallPacketID_PACKET_CS_ENTERTHRIDGAME), &CSEnterThridGamePacketFactory{})
	common.RegisterHandler(int(gamehall.GameHallPacketID_PACKET_CS_LEAVETHRIDGAME), &CSLeaveThridGameHandler{})
	netlib.RegisterFactory(int(gamehall.GameHallPacketID_PACKET_CS_LEAVETHRIDGAME), &CSLeaveThridGamePacketFactory{})
	common.RegisterHandler(int(gamehall.GameHallPacketID_PACKET_CS_THRIDGAMEBALANCEUPDATE), &CSThridBalanceRefreshHandler{})
	netlib.RegisterFactory(int(gamehall.GameHallPacketID_PACKET_CS_THRIDGAMEBALANCEUPDATE), &CSThridBalanceRefreshPacketFactory{})

	common.RegisterHandler(int(gamehall.GameHallPacketID_PACKET_CS_GETPRIVATEROOMLIST), &CSGetPrivateRoomListHandler{})
	netlib.RegisterFactory(int(gamehall.GameHallPacketID_PACKET_CS_GETPRIVATEROOMLIST), &CSGetPrivateRoomListPacketFactory{})
	common.RegisterHandler(int(gamehall.GameHallPacketID_PACKET_CS_GETPRIVATEROOMHISTORY), &CSGetPrivateRoomHistoryHandler{})
	netlib.RegisterFactory(int(gamehall.GameHallPacketID_PACKET_CS_GETPRIVATEROOMHISTORY), &CSGetPrivateRoomHistoryPacketFactory{})

	common.RegisterHandler(int(gamehall.GameHallPacketID_PACKET_CS_LOTTERYLOG), &CSLotteryLogHandler{})
	netlib.RegisterFactory(int(gamehall.GameHallPacketID_PACKET_CS_LOTTERYLOG), &CSLotteryLogPacketFactory{})
	//获取指定游戏配置 包括分场信息
	common.RegisterHandler(int(gamehall.GameHallPacketID_PACKET_CS_GETGAMECONFIG), &CSGetGameConfigHandler{})
	netlib.RegisterFactory(int(gamehall.GameHallPacketID_PACKET_CS_GETGAMECONFIG), &CSGetGameConfigPacketFactory{})
	//玩家进入游戏
	common.RegisterHandler(int(gamehall.GameHallPacketID_PACKET_CS_ENTERGAME), &CSEnterGameHandler{})
	netlib.RegisterFactory(int(gamehall.GameHallPacketID_PACKET_CS_ENTERGAME), &CSEnterGamePacketFactory{})
	//玩家进入房间
	common.RegisterHandler(int(gamehall.GameHallPacketID_PACKET_CS_ENTERROOM), &CSEnterRoomHandler{})
	netlib.RegisterFactory(int(gamehall.GameHallPacketID_PACKET_CS_ENTERROOM), &CSEnterRoomPacketFactory{})
	//玩家退出游戏
	common.RegisterHandler(int(gamehall.GameHallPacketID_PACKET_CS_QUITGAME), &CSQuitGameHandler{})
	netlib.RegisterFactory(int(gamehall.GameHallPacketID_PACKET_CS_QUITGAME), &CSQuitGamePacketFactory{})
	//创建房间
	common.RegisterHandler(int(gamehall.GameHallPacketID_PACKET_CS_CREATEROOM), &CSCreateRoomHandler{})
	netlib.RegisterFactory(int(gamehall.GameHallPacketID_PACKET_CS_CREATEROOM), &CSCreateRoomPacketFactory{})
	//查询公共房间列表
	common.RegisterHandler(int(gamehall.GameHallPacketID_PACKET_CS_QUERYROOMINFO), &CSQueryRoomInfoHandler{})
	netlib.RegisterFactory(int(gamehall.GameHallPacketID_PACKET_CS_QUERYROOMINFO), &CSQueryRoomInfoPacketFactory{})
	//观众坐下
	common.RegisterHandler(int(gamehall.GameHallPacketID_PACKET_CS_AUDIENCESIT), &CSAudienceSitHandler{})
	netlib.RegisterFactory(int(gamehall.GameHallPacketID_PACKET_CS_AUDIENCESIT), &CSAudienceSitPacketFactory{})
	//我的游戏信息及平台公告
	common.RegisterHandler(int(gamehall.GameHallPacketID_PACKET_CS_COMNOTICE), &CSRecordAndNoticeHandler{})
	netlib.RegisterFactory(int(gamehall.GameHallPacketID_PACKET_CS_COMNOTICE), &CSRecordAndNoticePacketFactory{})
}
