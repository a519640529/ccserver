package main

//import (
//	"games.yol.com/win88/proto"
//	"games.yol.com/win88/common"
//	"games.yol.com/win88/protocol"
//	"github.com/idealeak/goserver/core/logger"
//	"github.com/idealeak/goserver/srvlib"
//)
//
//var NpcServerAgentSington = &NpcServerAgent{}
//
//type NpcServerAgent struct {
//}
//
//func (nsa *NpcServerAgent) Invite(roomId, cnt int, isAgent bool, p *Player, matchId int32) bool {
//	//logger.Logger.Trace("(nsa *NpcServerAgent) Invite", roomId, cnt, isAgent, matchId)
//	npcSess := srvlib.ServerSessionMgrSington.GetSession(common.GetSelfAreaId(), common.RobotServerType, common.RobotServerId)
//	if npcSess != nil {
//		if isAgent {
//			cnt = 0
//		}
//		pack := &protocol.WRInviteRobot{
//			RoomId:  proto.Int(roomId),
//			MatchId: proto.Int32(matchId),
//			Cnt:     proto.Int(cnt),
//		}
//		proto.SetDefaults(pack)
//		npcSess.Send(int(protocol.MmoPacketID_PACKET_WR_INVITEROBOT), pack)
//		return true
//	} else {
//		//logger.Logger.Error("Robot server not found.")
//	}
//	return false
//}
//
//func (nsa *NpcServerAgent) OnPlayerEnterScene(s *Scene, p *Player) {
//	logger.Logger.Trace("(nsa *NpcServerAgent) OnPlayerEnterScene")
//}
//
//func (nsa *NpcServerAgent) OnPlayerLeaveScene(s *Scene, p *Player) {
//	logger.Logger.Trace("(nsa *NpcServerAgent) OnPlayerLeaveScene")
//}
//
//func (nsa *NpcServerAgent) OnSceneClose(s *Scene) {
//	logger.Logger.Trace("(nsa *NpcServerAgent) OnSceneClose")
//}
