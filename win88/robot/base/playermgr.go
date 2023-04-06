package base

import (
	"fmt"
	"games.yol.com/win88/common"
	"games.yol.com/win88/proto"
	hall_proto "games.yol.com/win88/protocol/gamehall"
	player_proto "games.yol.com/win88/protocol/player"
	"games.yol.com/win88/protocol/tournament"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"math/rand"
	"time"
)

var PlayerMgrSington = &PlayerMgr{
	playersMapSnId: make(map[int32]*player_proto.SCPlayerData),
	playersSession: make(map[int32]*netlib.Session),
}

type PlayerMgr struct {
	playersMapSnId map[int32]*player_proto.SCPlayerData
	playersSession map[int32]*netlib.Session
}

func (pm *PlayerMgr) AddPlayer(data *player_proto.SCPlayerData, s *netlib.Session) {
	snid := data.GetData().GetSnId()
	pm.playersMapSnId[snid] = data
	pm.playersSession[snid] = s
}

func (pm *PlayerMgr) DelPlayer(snid int32) {
	delete(pm.playersMapSnId, snid)
	if s, ok := pm.playersSession[snid]; ok {
		delete(pm.playersSession, snid)
		if s != nil {
			StopSessionTimer(s)
			StopSessionGameTimer(s)
		}
	}
}

func (pm *PlayerMgr) GetPlayer(snid int32) *player_proto.SCPlayerData {
	if data, ok := pm.playersMapSnId[snid]; ok {
		return data
	}
	return nil
}

func (pm *PlayerMgr) GetPlayerSession(snid int32) *netlib.Session {
	if data, ok := pm.playersSession[snid]; ok {
		return data
	}
	return nil
}

func (pm *PlayerMgr) ProcessInvite(roomId, id, cnt int32, platform string, isMatch, needAwait bool) {
	var freePlayers []*netlib.Session
	var inScene, inCoinSceneQueue, inMatch, waiting, doing int
	for _, s := range pm.playersSession {
		if HadScene(s) {
			inScene++
		} else if InCoinSceneQueue(s) {
			inCoinSceneQueue++
		} else if AwaitMatchOrMatchDoing(s) {
			inMatch++
			if matchid, ok := s.GetAttribute(SessionAttributeWaitingMatch).(int32); ok {
				if matchid > 0 {
					waiting++
				}
			}
			if s.GetAttribute(SessionAttributeMatchDoing) != nil {
				doing++
			}
		} else {
			freePlayers = append(freePlayers, s)
		}
	}

	total := len(pm.playersSession)
	freeCnt := len(freePlayers)
	logger.Logger.Infof(">>>>>>total:%v , current free count:%v , alloc:%v<<<<<<", total, freeCnt, cnt)
	logger.Logger.Infof(">>>>>>inScene:%v , inCoinSceneQueue:%v , inMatch:%v<<<<<<", inScene, inCoinSceneQueue, inMatch)
	logger.Logger.Infof(">>>>>>inMatch:%v , waiting:%v , doing:%v<<<<<<", inMatch, waiting, doing)

	if freeCnt <= 0 {
		return
	}
	if cnt > int32(freeCnt) {
		cnt = int32(freeCnt)
	}
	idies := rand.Perm(freeCnt)
	for i := 0; i < int(cnt) && i < len(idies); i++ {
		s := freePlayers[idies[i]]
		if roomId >= common.CoinSceneStartId && roomId < common.CoinSceneMaxId {
			dbGameFree := SceneMgrSington.GetSceneDBGameFree(roomId, id)
			if dbGameFree != nil {
				coin := dbGameFree.GetLimitCoin()*10 + 100000
				me := GetUser(s)
				if me != nil {
					if me.GetData().GetCoin() < int64(coin) {
						ExePMCmd(s, fmt.Sprintf("%v%v%v", common.PMCmd_AddCoin, common.PMCmd_SplitToken, coin))
					}
				}
			}
			CSCoinSceneOp := &hall_proto.CSCoinSceneOp{
				Id:       proto.Int32(id),
				OpType:   proto.Int32(0),
				OpParams: []int32{roomId},
			}
			proto.SetDefaults(CSCoinSceneOp)
			s.Send(int(hall_proto.CoinSceneGamePacketID_PACKET_CS_COINSCENE_OP), CSCoinSceneOp)
		} else if roomId >= common.HundredSceneStartId && roomId < common.HundredSceneMaxId {
			dbGameFree := SceneMgrSington.GetSceneDBGameFree(roomId, id)
			if dbGameFree != nil {
				takeCoinArr := dbGameFree.GetRobotTakeCoin()
				takeCoin := int64(0)
				if len(takeCoinArr) == 2 {
					takeCoin = int64(common.RandInt(int(takeCoinArr[0]), int(takeCoinArr[1])))
				}
				me := GetUser(s)
				if me != nil {
					if me.GetData().GetCoin() < int64(takeCoin) {
						ExePMCmd(s, fmt.Sprintf("%v%v%v", common.PMCmd_AddCoin, common.PMCmd_SplitToken, takeCoin))
					}
				}
			}
			time.AfterFunc(time.Second, func() {
				CSHundredSceneOp := &hall_proto.CSHundredSceneOp{
					Id:       proto.Int32(id),
					OpType:   proto.Int32(0),
					OpParams: []int32{roomId},
				}
				proto.SetDefaults(CSHundredSceneOp)
				s.Send(int(hall_proto.HundredScenePacketID_PACKET_CS_HUNDREDSCENE_OP), CSHundredSceneOp)
			})
		} else if roomId >= common.HallSceneStartId && roomId < common.HallSceneMaxId {
			CSEnterRoom := &hall_proto.CSEnterRoom{
				RoomId: proto.Int32(roomId),
				GameId: proto.Int(0),
			}
			proto.SetDefaults(CSEnterRoom)
			s.Send(int(hall_proto.GameHallPacketID_PACKET_CS_ENTERROOM), CSEnterRoom)
		} else {
			if isMatch { //比赛场
				pack := &tournament.CSSignRace{
					TMId: proto.Int32(id),
				}
				proto.SetDefaults(pack)
				s.Send(int(tournament.TOURNAMENTID_PACKET_TM_CSSignRace), pack)
				if needAwait {
					s.SetAttribute(SessionAttributeWaitingMatch, id)
				}
			} else { //班车系统
				if len(platform) > 0 {
					dbGameFree := SceneMgrSington.GetSceneDBGameFree(roomId, id)
					if dbGameFree != nil {
						coin := dbGameFree.GetLimitCoin()
						me := GetUser(s)
						if me != nil {
							if me.GetData().GetCoin() < int64(coin) {
								ExePMCmd(s, fmt.Sprintf("%v%v%v", common.PMCmd_AddCoin, common.PMCmd_SplitToken, coin))
							}
						}
					}
					CSCoinSceneOp := &hall_proto.CSCoinSceneOp{
						Id:       proto.Int32(id),
						OpType:   proto.Int32(0),
						Platform: proto.String(platform),
					}
					proto.SetDefaults(CSCoinSceneOp)
					s.Send(int(hall_proto.CoinSceneGamePacketID_PACKET_CS_COINSCENE_OP), CSCoinSceneOp)
					s.SetAttribute(SessionAttributeCoinSceneQueue, true)
				}
			}
		}
	}
}

func (pm *PlayerMgr) ProcessInviteCreateRoom(gamefreeId, cnt int32) {
	var freePlayers []*netlib.Session
	var inScene, inCoinSceneQueue int
	for _, s := range pm.playersSession {
		if HadScene(s) {
			inScene++
		} else if InCoinSceneQueue(s) {
			inCoinSceneQueue++
		} else {
			freePlayers = append(freePlayers, s)
		}
	}

	total := len(pm.playersSession)
	freeCnt := len(freePlayers)
	logger.Logger.Infof(">>>>>>total:%v , current free count:%v , alloc:%v<<<<<<", total, freeCnt, cnt)
	logger.Logger.Infof(">>>>>>inScene:%v , inCoinSceneQueue:%v<<<<<<", inScene, inCoinSceneQueue)

	if freeCnt <= 0 {
		return
	}
	if cnt > int32(freeCnt) {
		cnt = int32(freeCnt)
	}
	idies := rand.Perm(freeCnt)
	for i := 0; i < int(cnt) && i < len(idies); i++ {
		s := freePlayers[idies[i]]
		dbGameFree := SceneMgrSington.GetSceneDBGameFree(-1, gamefreeId)
		if dbGameFree == nil {
			return
		}
		if common.IsCoinSceneType(dbGameFree.GetGameType()) {
			if dbGameFree != nil {
				coin := dbGameFree.GetLimitCoin()
				me := GetUser(s)
				if me != nil {
					if me.GetData().GetCoin() < int64(coin) {
						ExePMCmd(s, fmt.Sprintf("%v%v%v", common.PMCmd_AddCoin, common.PMCmd_SplitToken, coin))
					}
				}
			}
			CSCoinSceneOp := &hall_proto.CSCoinSceneOp{
				Id:     proto.Int32(gamefreeId),
				OpType: proto.Int32(0),
			}
			proto.SetDefaults(CSCoinSceneOp)
			s.Send(int(hall_proto.CoinSceneGamePacketID_PACKET_CS_COINSCENE_OP), CSCoinSceneOp)
		} else if common.IsHundredType(dbGameFree.GetGameType()) {
			if dbGameFree != nil {
				coin := dbGameFree.GetLimitCoin()
				me := GetUser(s)
				if me != nil {
					if me.GetData().GetCoin() < int64(coin) {
						ExePMCmd(s, fmt.Sprintf("%v%v%v", common.PMCmd_AddCoin, common.PMCmd_SplitToken, coin))
					}
				}
			}
			CSHundredSceneOp := &hall_proto.CSHundredSceneOp{
				Id:     proto.Int32(gamefreeId),
				OpType: proto.Int32(0),
			}
			proto.SetDefaults(CSHundredSceneOp)
			s.Send(int(hall_proto.HundredScenePacketID_PACKET_CS_HUNDREDSCENE_OP), CSHundredSceneOp)
		} else {
			return
		}
	}
}

func (pm *PlayerMgr) OnHalfSecondTimer() {
}
func (pm *PlayerMgr) OnSecondTimer() {

}
func (pm *PlayerMgr) OnMiniTimer() {
	pm.ProcessInvite(0, 0, 0, "", false, false)
}
