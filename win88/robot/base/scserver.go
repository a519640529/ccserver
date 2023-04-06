package base

import (
	server_proto "games.yol.com/win88/protocol/server"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"strconv"
)

type CoinSceneInvite struct {
	Id      int32
	SceneId int32
}

var WaitingRoomId = make(chan int32, 1024)
var WaitingMatchId = make(chan int32, 1024)
var WaitingCoinSceneId = make(chan *CoinSceneInvite, 8)
var WaitingHundredSceneId = make(chan *CoinSceneInvite, 8)

func init() {
	netlib.RegisterFactory(int(server_proto.SSPacketID_PACKET_WR_INVITEROBOT), netlib.PacketFactoryWrapper(func() interface{} {
		return &server_proto.WRInviteRobot{}
	}))
	netlib.RegisterHandler(int(server_proto.SSPacketID_PACKET_WR_INVITEROBOT), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
		logger.Logger.Trace("receive WRInviteRobot:", pack)
		if msg, ok := pack.(*server_proto.WRInviteRobot); ok {
			roomId := msg.GetRoomId()
			gamefreeId := msg.GetMatchId()
			cnt := msg.GetCnt()
			platform := msg.GetPlatform()
			isMatch := msg.GetIsMatch()
			needAwait := msg.GetNeedAwait()
			PlayerMgrSington.ProcessInvite(roomId, gamefreeId, cnt, platform, isMatch, needAwait)
		}
		return nil
	}))

	netlib.RegisterFactory(int(server_proto.SSPacketID_PACKET_WR_INVITECREATEROOM), netlib.PacketFactoryWrapper(func() interface{} {
		return &server_proto.WRInviteCreateRoom{}
	}))
	netlib.RegisterHandler(int(server_proto.SSPacketID_PACKET_WR_INVITECREATEROOM), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
		logger.Logger.Trace("receive WRInviteCreateRoom:", pack)
		if msg, ok := pack.(*server_proto.WRInviteCreateRoom); ok {
			gamefreeId := msg.GetMatchId()
			cnt := msg.GetCnt()
			PlayerMgrSington.ProcessInviteCreateRoom(gamefreeId, cnt)
		}
		return nil
	}))

	netlib.RegisterFactory(int(server_proto.SSPacketID_PACKET_GN_PLAYERCARDS), netlib.PacketFactoryWrapper(func() interface{} {
		return &server_proto.GNPlayerCards{}
	}))
	netlib.RegisterHandler(int(server_proto.SSPacketID_PACKET_GN_PLAYERCARDS), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
		logger.Logger.Trace("receive GNPlayerCards:", pack)
		if msg, ok := pack.(*server_proto.GNPlayerCards); ok {
			sceneId := msg.GetSceneId()
			scene := SceneMgrSington.GetScene(sceneId)
			if scene != nil {
				scene.SetAIMode(msg.GetNowRobotMode())
				playerCards := msg.GetPlayerCards()
				for _, pc := range playerCards {
					player := scene.GetPlayerByPos(pc.GetPos())
					if player != nil {
						player.UpdateCards(pc.GetCards())
					}
				}
			}
		}
		return nil
	}))

	netlib.RegisterFactory(int(server_proto.SSPacketID_PACKET_GN_PLAYERPARAM), netlib.PacketFactoryWrapper(func() interface{} {
		return &server_proto.GNPlayerParam{}
	}))
	netlib.RegisterHandler(int(server_proto.SSPacketID_PACKET_GN_PLAYERPARAM), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
		logger.Logger.Trace("receive GNPlayerParam:", pack)
		if msg, ok := pack.(*server_proto.GNPlayerParam); ok {
			sceneId := msg.GetSceneId()
			scene := SceneMgrSington.GetScene(sceneId)
			if scene != nil {
				for _, pc := range msg.GetPlayerdata() {
					if pc != nil {
						player := scene.GetPlayerByPos(pc.GetPos())
						////目前只有炸金花会走这个协议
						//if player != WinthreeNilPlayer {
						player.UpdateBasePlayers(strconv.Itoa(int(scene.GetGameId())), pc)
						//}
					}
				}
			}
		}
		return nil
	}))
	//GameFree数据同步
	netlib.RegisterFactory(int(server_proto.SSPacketID_PACKET_GR_GameFreeData), netlib.PacketFactoryWrapper(func() interface{} {
		return &server_proto.GRGameFreeData{}
	}))
	netlib.RegisterHandler(int(server_proto.SSPacketID_PACKET_GR_GameFreeData), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
		logger.Logger.Trace("receive GRGameFreeData:", pack)
		if msg, ok := pack.(*server_proto.GRGameFreeData); ok {
			roomId := msg.GetRoomId()
			dbGameFree := msg.GetDBGameFree()
			SceneMgrSington.UpdateSceneDBGameFree(roomId, dbGameFree)
		}
		return nil
	}))
	//房间销毁
	netlib.RegisterFactory(int(server_proto.SSPacketID_PACKET_GR_DESTROYSCENE), netlib.PacketFactoryWrapper(func() interface{} {
		return &server_proto.GRDestroyScene{}
	}))
	netlib.RegisterHandler(int(server_proto.SSPacketID_PACKET_GR_DESTROYSCENE), netlib.HandlerWrapper(func(s *netlib.Session, packetid int, pack interface{}) error {
		logger.Logger.Trace("receive GRGameFreeData:", pack)
		if msg, ok := pack.(*server_proto.GRDestroyScene); ok {
			sceneId := msg.GetSceneId()
			SceneMgrSington.DelScene(sceneId)
		}
		return nil
	}))

}
