package base

import (
	"games.yol.com/win88/common"
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/server"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/module"
	"github.com/idealeak/goserver/srvlib"
	"time"
)

var NpcServerAgentSington = &NpcServerAgent{}

type NpcServerAgent struct {
}

func (nsa *NpcServerAgent) OnConnected() {
}

func (nsa *NpcServerAgent) OnDisconnected() {
	RobotSceneDBGameFreeSync = make(map[int]bool)
}

func (nsa *NpcServerAgent) Invite(roomId, cnt int, isAgent bool, p *Player, gamefreeid int32) bool {
	//logger.Logger.Trace("(nsa *NpcServerAgent) Invite", roomId, cnt, isAgent, gamefreeid)
	npcSess := srvlib.ServerSessionMgrSington.GetSession(common.GetSelfAreaId(), common.RobotServerType, common.RobotServerId)
	if npcSess != nil {
		if isAgent {
			cnt = 0
		}
		pack := &server.WRInviteRobot{
			RoomId:  proto.Int(roomId),
			MatchId: proto.Int32(gamefreeid),
			Cnt:     proto.Int(cnt),
		}
		proto.SetDefaults(pack)
		npcSess.Send(int(server.SSPacketID_PACKET_WR_INVITEROBOT), pack)
		return true
	} else {
		//logger.Logger.Error("Robot server not found.")
	}
	return false
}

func (nsa *NpcServerAgent) OnPlayerEnterScene(s *Scene, p *Player) {
	logger.Logger.Trace("(nsa *NpcServerAgent) OnPlayerEnterScene")
}

func (nsa *NpcServerAgent) OnPlayerLeaveScene(s *Scene, p *Player) {
	logger.Logger.Trace("(nsa *NpcServerAgent) OnPlayerLeaveScene")
}

func (nsa *NpcServerAgent) OnSceneClose(s *Scene) {
	logger.Logger.Trace("(nsa *NpcServerAgent) OnSceneClose")
}

func (nsa *NpcServerAgent) SendPacket(packetid int, pack interface{}) bool {
	npcSess := srvlib.ServerSessionMgrSington.GetSession(common.GetSelfAreaId(), common.RobotServerType, common.RobotServerId)
	if npcSess != nil {
		npcSess.Send(int(packetid), pack)
		return true
	}
	return false
}

var InviteRobMgrSington = &InviteRobMgr{}

type InviteRobMgr struct {
}

func (this *InviteRobMgr) ModuleName() string {
	return "InviteRobMgr"
}

func (this *InviteRobMgr) Init() {

}

func (this *InviteRobMgr) Update() {
	if !model.GameParamData.IsRobFightTest {
		return
	}
	//
	//gameIds := []int{}
	//data := netlib.Config.SrvInfo.Data
	//if data != "" {
	//	gameids := strings.Split(data, ",")
	//	for _, id := range gameids {
	//		if gameid, err := strconv.Atoi(id); err == nil {
	//			gameIds = append(gameIds, gameid)
	//		}
	//	}
	//}
	//
	//datas := srvdata.PBDB_GameFreeMgr.Datas.GetArr()
	//for _, data := range datas {
	//	if data.GetAi()[0] == 1 {
	//		if len(gameIds) > 0 {
	//			for index := 0; index < len(gameIds); index++ {
	//				if int(data.GetGameId()) == gameIds[index] {
	//					NpcServerAgentSington.InviteCreateRoom(100, false, data.GetId())
	//					break
	//				}
	//			}
	//		} else {
	//			NpcServerAgentSington.InviteCreateRoom(100, false, data.GetId())
	//		}
	//	}
	//}
}

func (this *InviteRobMgr) Shutdown() {
	module.UnregisteModule(this)
}

func init() {
	module.RegisteModule(InviteRobMgrSington, time.Minute, 0)
}
