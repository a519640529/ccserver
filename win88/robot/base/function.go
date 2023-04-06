package base

import (
	gamehall_proto "games.yol.com/win88/protocol/gamehall"
	player_proto "games.yol.com/win88/protocol/player"
	"math/rand"
	"time"

	"games.yol.com/win88/common"
	"games.yol.com/win88/proto"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/core/timer"
)

func GetUser(s *netlib.Session) *player_proto.SCPlayerData {
	if user, ok := s.GetAttribute(SessionAttributeUser).(*player_proto.SCPlayerData); ok {
		return user
	}
	return nil
}

func GetScene(s *netlib.Session) Scene {
	if sceneId, ok := s.GetAttribute(SessionAttributeSceneId).(int32); ok {
		return SceneMgrSington.GetScene(sceneId)
	}
	ss := s.GetAttribute(SessionAttributeScene)
	if ss != nil {
		if scene, ok := ss.(Scene); ok {
			return scene
		}
	}
	return nil
}

func HadScene(s *netlib.Session) bool {
	if sceneId, ok := s.GetAttribute(SessionAttributeSceneId).(int32); ok {
		return sceneId > 0
	}
	if scene := s.GetAttribute(SessionAttributeScene); scene != nil {
		return true
	}
	return false
}
func InCoinSceneQueue(s *netlib.Session) bool {
	if _, ok := s.GetAttribute(SessionAttributeCoinSceneQueue).(bool); ok {
		return true
	} else {
		return false
	}
}
func AwaitMatchOrMatchDoing(s *netlib.Session) bool {
	if matchid, ok := s.GetAttribute(SessionAttributeWaitingMatch).(int32); ok {
		if matchid > 0 {
			return true
		}
	}
	if s.GetAttribute(SessionAttributeMatchDoing) != nil {
		return true
	}
	return false
}
func StartSessionTimer(s *netlib.Session, act timer.TimerAction, ud interface{}, interval time.Duration, times int) bool {
	StopSessionTimer(s)
	if hTimer, ok := timer.StartTimer(act, ud, interval, times); ok {
		s.SetAttribute(SessionAttributeTimer, hTimer)
		return true
	}
	return false
}

func StopSessionTimer(s *netlib.Session) {
	if h, ok := s.GetAttribute(SessionAttributeTimer).(timer.TimerHandle); ok {
		if h != timer.TimerHandle(0) {
			timer.StopTimer(h)
			s.RemoveAttribute(SessionAttributeTimer)
		}
	}
}

func StartSessionPingTimer(s *netlib.Session, act timer.TimerAction, ud interface{}, interval time.Duration, times int) bool {
	StopSessionPingTimer(s)
	if hTimer, ok := timer.StartTimer(act, ud, interval, times); ok {
		s.SetAttribute(SessionAttributePingTimer, hTimer)
		return true
	}
	return false
}

func StopSessionPingTimer(s *netlib.Session) {
	if h, ok := s.GetAttribute(SessionAttributePingTimer).(timer.TimerHandle); ok {
		if h != timer.TimerHandle(0) {
			timer.StopTimer(h)
			s.RemoveAttribute(SessionAttributePingTimer)
		}
	}
}

func StartSessionGameTimer(s *netlib.Session, act timer.TimerAction, ud interface{}, interval time.Duration, times int) bool {
	StopSessionGameTimer(s)
	if hTimer, ok := timer.StartTimer(act, ud, interval, times); ok {
		s.SetAttribute(SessionAttributeGameTimer, hTimer)
		return true
	}
	return false
}

func StopSessionGameTimer(s *netlib.Session) {
	if h, ok := s.GetAttribute(SessionAttributeGameTimer).(timer.TimerHandle); ok {
		if h != timer.TimerHandle(0) {
			timer.StopTimer(h)
			s.RemoveAttribute(SessionAttributeGameTimer)
		}
	}
}

func StartSessionLoginTimer(s *netlib.Session, act timer.TimerAction, ud interface{}, interval time.Duration, times int) bool {
	StopSessionLoginTimer(s)
	if hTimer, ok := timer.StartTimer(act, ud, interval, times); ok {
		s.SetAttribute(SessionAttributeLoginTimer, hTimer)
		return true
	}
	return false
}

func StopSessionLoginTimer(s *netlib.Session) {
	if h, ok := s.GetAttribute(SessionAttributeLoginTimer).(timer.TimerHandle); ok {
		if h != timer.TimerHandle(0) {
			timer.StopTimer(h)
			s.RemoveAttribute(SessionAttributeLoginTimer)
		}
	}
}

func TryEnterRoom(s *netlib.Session) {
	StartSessionTimer(s, timer.TimerActionWrapper(func(h timer.TimerHandle, ud interface{}) bool {
		if GetScene(s) == nil {
			scene := SceneMgrSington.GetOneNoFullScene()
			if scene == nil {
				StopSessionTimer(s)
				CSEnterRoom := &gamehall_proto.CSEnterRoom{
					RoomId: proto.Int(Config.RoomId),
					GameId: proto.Int(common.GameId_HundredDZNZ),
				}
				proto.SetDefaults(CSEnterRoom)
				s.Send(int(gamehall_proto.GameHallPacketID_PACKET_CS_ENTERROOM), CSEnterRoom)
				return true
			}
			StopSessionTimer(s)
			CSEnterRoom := &gamehall_proto.CSEnterRoom{
				RoomId: proto.Int32(scene.GetRoomId()),
				GameId: proto.Int(common.GameId_HundredDZNZ),
			}
			proto.SetDefaults(CSEnterRoom)
			s.Send(int(gamehall_proto.GameHallPacketID_PACKET_CS_ENTERROOM), CSEnterRoom)
			return true
		}
		return true
	}), nil, time.Second, -1)
}

func TryEnterRoomWatch(s *netlib.Session) {
	StartSessionTimer(s, timer.TimerActionWrapper(func(h timer.TimerHandle, ud interface{}) bool {
		if GetScene(s) == nil {
			scene := SceneMgrSington.GetOneFullScene()
			if scene == nil {
				StopSessionTimer(s)
				CSEnterRoom := &gamehall_proto.CSEnterRoom{
					RoomId: proto.Int(Config.RoomId),
					GameId: proto.Int(common.GameId_HundredDZNZ),
				}
				proto.SetDefaults(CSEnterRoom)
				s.Send(int(gamehall_proto.GameHallPacketID_PACKET_CS_AUDIENCE_ENTERROOM), CSEnterRoom)
				return true
			}
			StopSessionTimer(s)
			CSEnterRoom := &gamehall_proto.CSEnterRoom{
				RoomId: proto.Int32(scene.GetRoomId()),
				GameId: proto.Int(common.GameId_HundredDZNZ),
			}
			proto.SetDefaults(CSEnterRoom)
			s.Send(int(gamehall_proto.GameHallPacketID_PACKET_CS_AUDIENCE_ENTERROOM), CSEnterRoom)
			return true
		}
		return true
	}), nil, time.Second, -1)
}

func CreateRoom(s *netlib.Session) {
	//CSCreateRoom := &player_proto.CSCreateRoom{
	//	GameId:   proto.Int(common.GameId_HundredDZNZ),
	//	ModeType: proto.Int(mahjong.Mahjong_ErRen),
	//	Params:   []int32{0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0},
	//}
	//proto.SetDefaults(CSCreateRoom)
	//s.Send(int(player_proto.MmoPacketID_PACKET_CS_CREATEROOM), CSCreateRoom)
}

func DestroyRoom(s *netlib.Session) {
	pack := &gamehall_proto.CSDestroyRoom{}
	proto.SetDefaults(pack)
	s.Send(int(gamehall_proto.GameHallPacketID_PACKET_CS_DESTROYROOM), pack)
}

func ExePMCmd(s *netlib.Session, cmd string) {
	CSPMCmd := &player_proto.CSPMCmd{
		Cmd: proto.String(cmd),
	}
	proto.SetDefaults(CSPMCmd)
	s.Send(int(player_proto.PlayerPacketID_PACKET_CS_PMCMD), CSPMCmd)
}

func WaitInvite(s *netlib.Session) {
	//StartSessionTimer(s, timer.TimerActionWrapper(func(h timer.TimerHandle, ud interface{}) bool {
	//	scence := GetScene(s)
	//	if scence == nil {
	//		select {
	//		case roomId := <-WaitingRoomId:
	//			StopSessionTimer(s)
	//			CSEnterRoom := &player_proto.CSEnterRoom{
	//				RoomId: proto.Int32(roomId),
	//				GameId: proto.Int(common.GameId_HundredDZNZ),
	//			}
	//			proto.SetDefaults(CSEnterRoom)
	//			s.Send(int(player_proto.MmoPacketID_PACKET_CS_ENTERROOM), CSEnterRoom)
	//			return true
	//		case csi := <-WaitingCoinSceneId:
	//			StopSessionTimer(s)
	//			dbGameFree := srvdata.PBDB_GameFreeMgr.GetData(csi.Id)
	//			if dbGameFree != nil {
	//				coin := dbGameFree.GetLimitCoin()
	//				me := GetUser(s)
	//				if me != nil {
	//					if me.GetData().GetCoin() < int64(coin) {
	//						ExePMCmd(s, fmt.Sprintf("%v%v%v", common.PMCmd_AddCoin, common.PMCmd_SplitToken, coin))
	//					}
	//				}
	//			}
	//			CSCoinSceneOp := &player_proto.CSCoinSceneOp{
	//				Id:       proto.Int32(csi.Id),
	//				OpType:   proto.Int32(0),
	//				OpParams: []int32{csi.SceneId},
	//			}
	//			proto.SetDefaults(CSCoinSceneOp)
	//			s.Send(int(player_proto.CoinSceneGamePacketID_PACKET_CS_COINSCENE_OP), CSCoinSceneOp)
	//		case hsi := <-WaitingHundredSceneId:
	//			StopSessionTimer(s)
	//			dbGameFree := srvdata.PBDB_GameFreeMgr.GetData(hsi.Id)
	//			if dbGameFree != nil {
	//				coin := dbGameFree.GetLimitCoin()
	//				me := GetUser(s)
	//				if me != nil {
	//					if me.GetData().GetCoin() < int64(coin) {
	//						ExePMCmd(s, fmt.Sprintf("%v%v%v", common.PMCmd_AddCoin, common.PMCmd_SplitToken, coin))
	//					}
	//				}
	//			}
	//			CSHundredSceneOp := &player_proto.CSHundredSceneOp{
	//				Id:       proto.Int32(hsi.Id),
	//				OpType:   proto.Int32(0),
	//				OpParams: []int32{hsi.SceneId},
	//			}
	//			proto.SetDefaults(CSHundredSceneOp)
	//			s.Send(int(player_proto.MmoPacketID_PACKET_CS_HUNDREDSCENE_OP), CSHundredSceneOp)
	//		default:
	//		}
	//		return true
	//	}
	//	return true
	//}), nil, time.Second, -1)

}

func ChooiesStrategy(s *netlib.Session, strategy int) {
	switch strategy {
	case Strategy_CreateRoom:
		logger.Logger.Trace("ChooiesStrategy ", strategy)
		if GetScene(s) == nil {
			CreateRoom(s)
		}
	case Strategy_EnterRoom:
		logger.Logger.Trace("ChooiesStrategy ", strategy)
		TryEnterRoom(s)
	case Strategy_Watch:
		logger.Logger.Trace("ChooiesStrategy ", strategy)
		TryEnterRoomWatch(s)
	case Strategy_CreateAndDestroyRoom:
		logger.Logger.Trace("ChooiesStrategy ", strategy)
		if GetScene(s) == nil {
			CreateRoom(s)
		} else {
			DestroyRoom(s)
		}
	case Strategy_WaitInvite:
		WaitInvite(s)
	}
}

func isInClude(card int64, tingCards []int64) bool {
	for _, tingCard := range tingCards {
		if card == tingCard {
			return true
		}
	}
	return false
}

func DelaySend(s *netlib.Session, packetid int, pack interface{}, params ...int) {
	DelaySendDuration(s, packetid, pack, time.Second, params...)
}

// 纯AI快速
func DelayAISend(s *netlib.Session, packetid int, pack interface{}, params ...int) {
	//min := 1
	//if len(params) > 0 {
	//	min = params[0]
	//}
	//max := min + 1
	//if len(params) > 1 {
	//	max = params[1]
	//}
	//if max < min {
	//	min, max = max, min
	//} else if max == min {
	//	max = min + 1
	//}
	//thinkSec := time.Duration(min + rand.Intn(max-min))
	thinkSec := time.Duration(1)
	timer.StartTimer(timer.TimerActionWrapper(func(h timer.TimerHandle, ud interface{}) bool {
		s.Send(packetid, pack)
		return true
	}), nil, time.Second*thinkSec, 1)
}
func DelaySend_Jason(s *netlib.Session, packetid int, pack interface{}, delay time.Duration) {
	if delay <= 0 {
		s.Send(packetid, pack)
		return
	}
	timer.StartTimer(timer.TimerActionWrapper(func(h timer.TimerHandle, ud interface{}) bool {
		s.Send(packetid, pack)
		return true
	}), nil, delay, 1)
}

// 100毫秒触发一次
func DelaySendNew(s *netlib.Session, packetid int, pack interface{}, params ...int) {
	min := 1
	if len(params) > 0 {
		min = params[0]
	}
	max := min + 3
	if len(params) > 1 {
		max = params[1]
	}
	if max < min {
		min, max = max, min
	} else if max == min {
		max = min + 1
	}
	min *= 10
	max *= 10
	thinkSec := time.Duration(min + rand.Intn(max-min))

	timer.StartTimer(timer.TimerActionWrapper(func(h timer.TimerHandle, ud interface{}) bool {
		s.Send(packetid, pack)
		return true
	}), nil, time.Millisecond*thinkSec*100, 1)
}

func DelaySendNewMillisecond(s *netlib.Session, packetid int, pack interface{}, params ...int) {
	min := 1
	if len(params) > 0 {
		min = params[0]
	}
	max := min + 3
	if len(params) > 1 {
		max = params[1]
	}
	if max < min {
		min, max = max, min
	} else if max == min {
		max = min + 1
	}

	thinkSec := time.Duration(min + rand.Intn(max-min))

	timer.StartTimer(timer.TimerActionWrapper(func(h timer.TimerHandle, ud interface{}) bool {
		s.Send(packetid, pack)
		return true
	}), nil, time.Millisecond*thinkSec, 1)
}

// DelaySendDuration 在某个时间区间随机发送一次消息
// unit 时间单位
// params 延时区间
func DelaySendDuration(s *netlib.Session, packetid int, pack interface{}, unit time.Duration, params ...int) {
	min := 1
	if len(params) > 0 {
		min = params[0]
	}
	max := min + 3
	if len(params) > 1 {
		max = params[1]
	}
	if max < min {
		min, max = max, min
	} else if max == min {
		max = min + 1
	}
	thinkSec := time.Duration(min + rand.Intn(max-min))

	timer.StartTimer(timer.TimerActionWrapper(func(h timer.TimerHandle, ud interface{}) bool {
		s.Send(packetid, pack)
		return true
	}), nil, unit*thinkSec, 1)
}
