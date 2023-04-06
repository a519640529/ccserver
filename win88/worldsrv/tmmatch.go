package main

import (
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/server"
	"games.yol.com/win88/protocol/tournament"
	webapi_proto "games.yol.com/win88/protocol/webapi"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/core/timer"
	srvproto "github.com/idealeak/goserver/srvlib/protocol"
	"time"
)

type TmPlayer struct {
	SnId  int32
	IsRob bool
	seq   int
}
type TmMatch struct {
	SortId     int32 //比赛顺序 key:GetTmId
	TMId       int32 //比赛配置Id
	TmPlayer   map[int32]*TmPlayer
	Platform   string
	gmd        *webapi_proto.GameMatchDate
	dbGameFree *server.DB_GameFree //游戏配置
}

func (tm *TmMatch) Start() {
	logger.Logger.Trace("(this *TmMatch) Start()")
	//通知客户端比赛开始
	pack := &tournament.SCTMStart{
		MatchId: proto.Int32(tm.TMId),
	}
	proto.SetDefaults(pack)
	logger.Logger.Trace("SCTMStart:", pack)
	tm.BroadcastMessage(int(tournament.TOURNAMENTID_PACKET_TM_SCTMStart), pack)

	//创建房间
	timer.StartTimer(timer.TimerActionWrapper(func(h timer.TimerHandle, ud interface{}) bool {
		MatchSceneMgrSington.MatchStart(tm)
		return true
	}), nil, time.Millisecond*3500, 1)
}

func (tm *TmMatch) Stop() {
	//销毁房间
	MatchSceneMgrSington.MatchStop(tm)
	logger.Logger.Trace("(this *TmMatch) Stop()")
}

func (tm *TmMatch) BroadcastMessage(packetId int, rawPack interface{}) {
	mgs := make(map[*netlib.Session][]*srvproto.MCSessionUnion)
	for _, tmp := range tm.TmPlayer {
		p := PlayerMgrSington.GetPlayerBySnId(tmp.SnId)
		if p != nil && p.gateSess != nil && p.IsOnLine() && p.scene == nil {
			mgs[p.gateSess] = append(mgs[p.gateSess], &srvproto.MCSessionUnion{
				Mccs: &srvproto.MCClientSession{
					SId: proto.Int64(p.sid),
				},
			})
		}
	}
	for gateSess, v := range mgs {
		if gateSess != nil && len(v) != 0 {
			pack, err := MulticastMaker.CreateMulticastPacket(packetId, rawPack, v...)
			if err == nil {
				proto.SetDefaults(pack)
				gateSess.Send(int(srvproto.SrvlibPacketID_PACKET_SS_MULTICAST), pack)
			}
		}
	}
}
func (s *TmMatch) CopyMap(b map[int32]*TmPlayer) {
	s.TmPlayer = make(map[int32]*TmPlayer)
	for _, v := range b {
		var tmp TmPlayer
		tmp = *v
		s.TmPlayer[v.SnId] = &tmp
	}
}
