package main

//
//import (
//	"math/rand"
//
//	"games.yol.com/win88/common"
//	"games.yol.com/win88/proto"
//	"games.yol.com/win88/protocol/mngame"
//	server_proto "games.yol.com/win88/protocol/server"
//	"github.com/idealeak/goserver/core/logger"
//	"github.com/idealeak/goserver/srvlib"
//	srvlibproto "github.com/idealeak/goserver/srvlib/protocol"
//)
//
//func (s *Scene) PlayerEnterMiniGame(p *Player) bool {
//	s.players[p.SnId] = p
//	s.gameSess.AddPlayer(p)
//	takeCoin := p.Coin
//	leaveCoin := int64(0)
//	gameTimes := rand.Int31n(100)
//	if p.IsRob {
//		takerng := s.dbGameFree.GetRobotTakeCoin()
//		if len(takerng) >= 2 && takerng[1] > takerng[0] {
//			if takerng[0] < s.dbGameFree.GetLimitCoin() {
//				takerng[0] = s.dbGameFree.GetLimitCoin()
//			}
//			takeCoin = int64(common.RandInt(int(takerng[0]), int(takerng[1])))
//		} else {
//			maxlimit := int64(s.dbGameFree.GetMaxCoinLimit())
//			if maxlimit != 0 && p.Coin > maxlimit {
//				logger.Logger.Trace("Player coin:", p.Coin)
//				//在下限和上限之间随机，并对其的100的整数倍
//				takeCoin = int64(common.RandInt(int(s.dbGameFree.GetLimitCoin()), int(maxlimit)))
//				logger.Logger.Trace("Take coin:", takeCoin)
//			}
//		}
//		takeCoin = takeCoin / 100 * 100
//		//离场金币
//		leaverng := s.dbGameFree.GetRobotLimitCoin()
//		if len(leaverng) >= 2 {
//			leaveCoin = int64(leaverng[0] + rand.Int31n(leaverng[1]-leaverng[0]))
//		}
//
//		if takeCoin > p.Coin {
//			takeCoin = p.Coin
//		}
//	}
//
//	if p.IsRob {
//		s.robotNum++
//		p.RobotRandName()
//		p.RobRandVipWhenEnterRoom(takeCoin)
//		name := s.GetSceneName()
//		logger.Logger.Tracef("(this *Scene) PlayerEnter(%v) robot(%v) robotlimit(%v)", name, s.robotNum, s.robotLimit)
//	}
//
//	data, err := p.MarshalData(s.gameId)
//	if err == nil {
//		var gateSid int64
//		if p.gateSess != nil {
//			if srvInfo, ok := p.gateSess.GetAttribute(srvlib.SessionAttributeServerInfo).(*srvlibproto.SSSrvRegiste); ok && srvInfo != nil {
//				sessionId := srvlib.NewSessionIdEx(srvInfo.GetAreaId(), srvInfo.GetType(), srvInfo.GetId(), 0)
//				gateSid = sessionId.Get()
//			}
//		}
//		isQuMin := false
//		//if !p.IsRob {
//		//	pt := PlatformMgrSington.GetPackageTag(p.PackageID)
//		//	if pt != nil && pt.SpreadTag == 1 {
//		//		isQuMin = true
//		//	}
//		//}
//		p.miniScene[s.dbGameFree.Id] = s
//		msg := &server_proto.WGPlayerEnterMiniGame{
//			Sid:        proto.Int64(p.sid),
//			SnId:       proto.Int32(p.SnId),
//			GateSid:    proto.Int64(gateSid),
//			SceneId:    proto.Int(s.sceneId),
//			PlayerData: data,
//			IsQM:       proto.Bool(isQuMin),
//		}
//		sa, err2 := p.MarshalSingleAdjustData(s.dbGameFree.Id)
//		if err2 == nil && sa != nil {
//			msg.SingleAdjust = sa
//		}
//		msg.TakeCoin = proto.Int64(takeCoin)
//		msg.ExpectLeaveCoin = proto.Int64(leaveCoin)
//		msg.ExpectGameTimes = proto.Int32(gameTimes)
//		proto.SetDefaults(msg)
//		s.SendToGame(int(server_proto.SSPacketID_PACKET_WG_PLAYERENTER_MINIGAME), msg)
//	}
//	return true
//}
//
//func (s *Scene) PlayerLeaveMiniGame(p *Player) bool {
//	var gateSid int64
//	if p.gateSess != nil {
//		if srvInfo, ok := p.gateSess.GetAttribute(srvlib.SessionAttributeServerInfo).(*srvlibproto.SSSrvRegiste); ok && srvInfo != nil {
//			sessionId := srvlib.NewSessionIdEx(srvInfo.GetAreaId(), srvInfo.GetType(), srvInfo.GetId(), 0)
//			gateSid = sessionId.Get()
//		}
//	}
//
//	delete(p.miniScene, s.dbGameFree.Id)
//
//	msg := &server_proto.WGPlayerLeaveMiniGame{
//		Sid:     proto.Int64(p.sid),
//		SnId:    proto.Int32(p.SnId),
//		GateSid: proto.Int64(gateSid),
//		SceneId: proto.Int(s.sceneId),
//	}
//	proto.SetDefaults(msg)
//	s.SendToGame(int(server_proto.SSPacketID_PACKET_WG_PLAYERLEAVE_MINIGAME), msg)
//
//	delete(s.players, p.SnId)
//	return true
//}
//
//func (s *Scene) RedirectMiniGameMsg(p *Player, msg *mngame.CSMNGameDispatcher) bool {
//	msg.Id = int32(s.sceneId) //换成真实房间id
//	msg.SnId = p.SnId
//	s.SendToGame(int(mngame.MNGamePacketID_PACKET_CS_MNGAME_DISPATCHER), msg)
//	return true
//}
