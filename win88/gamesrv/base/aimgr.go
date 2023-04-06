package base

import (
	"fmt"
	"games.yol.com/win88/common"
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/server"
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/task"
	"time"
)

type AIMgr interface {
	//获取属性
	GetAttribute(key interface{}) (interface{}, bool)
	//设置属性
	SetAttribute(key, val interface{})
	//房间启动
	OnStart(s *Scene)
	//房间心跳
	OnTick(s *Scene)
	//房间停止
	OnStop(s *Scene)
	//房间状态变化
	OnChangeState(s *Scene, oldstate, newstate int)
}

type BaseAIMgr struct {
	attribute map[interface{}]interface{}
}

//获取属性
func (bm *BaseAIMgr) GetAttribute(key interface{}) (interface{}, bool) {
	if bm.attribute != nil {
		v, ok := bm.attribute[key]
		return v, ok
	}
	return nil, false
}

//设置属性
func (bm *BaseAIMgr) SetAttribute(key, val interface{}) {
	if bm.attribute != nil {
		bm.attribute[key] = val
	}
}

//房间启动
func (bm *BaseAIMgr) OnStart(s *Scene) {
}

//房间心跳
func (bm *BaseAIMgr) OnTick(s *Scene) {
	//先驱动ai
	if s.WithLocalAI {
		for _, p := range s.Players {
			if p.ai != nil {
				p.ai.OnTick(s)
			}
		}
	}
}

//房间停止
func (bm *BaseAIMgr) OnStop(s *Scene) {
}

//房间状态变化
func (bm *BaseAIMgr) OnChangeState(s *Scene, oldstate, newstate int) {
	if s.WithLocalAI {
		for _, pp := range s.Players {
			if pp.ai != nil {
				pp.ai.OnChangeSceneState(s, oldstate, newstate)
			}
		}
	}
}

//通用对战场机器人管理
type CoinSceneAIMgr struct {
	BaseAIMgr
}

func (m *CoinSceneAIMgr) OnTick(s *Scene) {
	m.BaseAIMgr.OnTick(s)

	//私有房间不邀请机器人
	if s.IsPrivateScene() || s.IsMatchScene() {
		return
	}

	if s.DbGameFree.GetMatchMode() == 1 {
		return
	}

	//然后看是否需要补充机器人
	bot := int(s.DbGameFree.GetBot())
	if bot == 0 { //机器人不进的场
		return
	}

	if s.DbGameFree.GetMatchTrueMan() != common.MatchTrueMan_Forbid {
		//给真人保留一个空位
		if len(s.Players) >= s.playerNum-1 {
			return
		}
	}

	//对战场有真实玩家的情况才需要机器人匹配
	if !s.IsRobFightGame() && s.realPlayerNum <= 0 && !s.IsPreCreateScene() { //预创建房间的对战场可以优先进机器人，如:21点 判断依据:CreateRoomNum
		return
	}

	tNow := time.Now()
	if tNow.Before(s.nextInviteTime) {
		return
	}
	if model.GameParamData.EnterAfterStartSwitch && s.Gaming && !s.bEnterAfterStart {
		return
	}
	if s.robotNumLastInvite == s.robotNum {
		s.inviteInterval = s.inviteInterval + 1
		if s.inviteInterval > model.GameParamData.RobotInviteIntervalMax {
			s.inviteInterval = model.GameParamData.RobotInviteIntervalMax
		}
	} else {
		s.inviteInterval = model.GameParamData.RobotInviteInitInterval
	}

	s.ResetNextInviteTime()
	s.robotNumLastInvite = s.robotNum
	if !s.RobotIsLimit() {
		var robCnt int
		if s.robotLimit != 0 {
			if s.robotNum >= s.robotLimit { //机器人数量已达上限
				return
			}
			hadCnt := len(s.Players)
			robCnt = s.robotLimit - s.robotNum
			if robCnt > s.playerNum-hadCnt {
				robCnt = s.playerNum - hadCnt
			}
		} else {
			if s.IsFull() {
				return
			}
			hadCnt := len(s.Players)
			robCnt = s.playerNum - hadCnt
			if s.realPlayerNum == 0 { //一个真人都没有，不让机器人坐满房间
				robCnt--
			}
		}
		if robCnt > 0 {
			num := s.Rand.Int31n(int32(robCnt + 1))
			if num > 0 {
				if s.Gaming { //如果牌局正在进行中,一个一个进
					num = 1
				}
				//同步下房间里的参数
				if !RobotSceneDBGameFreeSync[s.SceneId] {
					pack := &server.GRGameFreeData{
						RoomId:     proto.Int(s.SceneId),
						DBGameFree: s.DbGameFree,
					}
					proto.SetDefaults(pack)
					if NpcServerAgentSington.SendPacket(int(server.SSPacketID_PACKET_GR_GameFreeData), pack) {
						RobotSceneDBGameFreeSync[s.SceneId] = true
					}
				}
				//然后再邀请
				if NpcServerAgentSington.Invite(s.SceneId, int(num), false, nil, s.DbGameFree.Id) {
				}
			}
		}
	}
}

//通用百人场机器人管理
type HundredSceneAIMgr struct {
	BaseAIMgr
}

func (m *HundredSceneAIMgr) OnTick(s *Scene) {
	m.BaseAIMgr.OnTick(s)

	//然后看是否需要补充机器人
	bot := int(s.DbGameFree.GetBot())
	if bot == 0 { //机器人不进的场
		return
	}

	if s.DbGameFree.GetMatchMode() == 1 {
		return
	}

	tNow := time.Now()
	if tNow.Before(s.nextInviteTime) {
		return
	}

	if s.robotNumLastInvite == s.robotNum {
		s.inviteInterval = s.inviteInterval + 1
		if s.inviteInterval > model.GameParamData.RobotInviteIntervalMax {
			s.inviteInterval = model.GameParamData.RobotInviteIntervalMax
		}
	} else {
		s.inviteInterval = model.GameParamData.RobotInviteInitInterval
	}

	s.ResetNextInviteTime()
	s.robotNumLastInvite = s.robotNum

	if !s.RobotIsLimit() {
		var robCnt int
		if s.robotLimit != 0 {
			robCnt = s.robotLimit - s.robotNum
		}
		if robCnt > 0 {
			num := s.Rand.Int31n(int32(robCnt + 1))
			if num > 0 {
				//百人场机器人本地创建,不使用robotsrv
				t, done := task.NewMutexTask(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
					var ids []int32
					err := s.ws.CallRpc("LocalRobotIdSvc.GetIds", num, &ids)
					if err != nil {

					}
					return ids
				}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
					if ids, ok := data.([]int32); ok {
						for _, id := range ids {
							p := PlayerMgrSington.AddLocalPlayer(id)
							if p != nil {
								s.PlayerEnter(p, true)
							}
						}
					}
				}), fmt.Sprintf("s(%d)", s.SceneId), "AIMgrInviteRobot")
				if !done {
					t.Start()
				}
			}
		}
	}
}
