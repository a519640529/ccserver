package main

import (
	"container/list"
	"fmt"
	"games.yol.com/win88/common"
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	hall_proto "games.yol.com/win88/protocol/gamehall"
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/task"
	"strconv"
	"time"
)

const (
	PrivateSceneState_Deleting = iota //删除中
	PrivateSceneState_Deleted         //已删除
)

var PrivateSceneMgrSington = &PrivateSceneMgr{
	pps: make(map[int32]*PlayerPrivateScene),
}

type PlayerPrivateScene struct {
	snid        int32  // 玩家id
	creatorName string //创建人昵称
	platform    string // 平台名称
	channel     string // 渠道名称
	promoter    string // 推广员
	packageTag  string // 推广包标识
	scenes      map[int]*Scene
	logsByDay   map[int]*list.List
	dupLog      map[string]struct{}
	loaded      bool
}

func (pps *PlayerPrivateScene) AddScene(s *Scene) {
	pps.scenes[s.sceneId] = s
}

func (pps *PlayerPrivateScene) GetScene(sceneId int) *Scene {
	if s, exist := pps.scenes[sceneId]; exist {
		return s
	}
	return nil
}

func (pps *PlayerPrivateScene) GetCount() int {
	return len(pps.scenes)
}

func (pps *PlayerPrivateScene) CanDelete() bool {
	return !pps.loaded && len(pps.scenes) == 0
}

func (pps *PlayerPrivateScene) OnPlayerLogin(p *Player) {

}

func (pps *PlayerPrivateScene) OnPlayerLogout(p *Player) {
	pps.logsByDay = nil
	pps.loaded = false
}

func (pps *PlayerPrivateScene) OnCreateScene(p *Player, s *Scene) {
	pps.scenes[s.sceneId] = s
}

func (pps *PlayerPrivateScene) LoadLogs(p *Player, yyyymmdd int32) {
	if !pps.loaded {
		var logs []*model.PrivateSceneLog
		var err error
		task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
			logs, err = model.GetPrivateSceneLogBySnId(p.Platform, p.SnId, model.GameParamData.PrivateSceneLogLimit)
			return nil
		}), task.CompleteNotifyWrapper(func(data interface{}, t task.Task) {
			if err == nil {
				pps.loaded = true
				pps.TidyLog(logs)
				pps.SendLogs(p, yyyymmdd)
			}
		}), "GetPrivateSceneLogBySnId").Start()
	} else {
		pps.SendLogs(p, yyyymmdd)
	}
}

func (pps *PlayerPrivateScene) TidyLog(logs []*model.PrivateSceneLog) {
	if pps.logsByDay == nil {
		pps.logsByDay = make(map[int]*list.List)
	}

	for _, log := range logs {
		if _, exist := pps.dupLog[log.LogId.Hex()]; exist {
			continue
		}
		y, m, d := log.CreateTime.Date()
		day := y*10000 + int(m)*100 + d
		if lst, exist := pps.logsByDay[day]; exist {
			lst.PushBack(log)
		} else {
			lst = list.New()
			pps.logsByDay[day] = lst
			lst.PushBack(log)
		}
	}
	pps.dupLog = nil
}

func (pps *PlayerPrivateScene) SendLogs(p *Player, yyyymmdd int32) {
	pack := &hall_proto.SCGetPrivateRoomHistory{
		QueryTime: proto.Int32(yyyymmdd),
	}
	if logs, exist := pps.logsByDay[int(yyyymmdd)]; exist {
		for e := logs.Front(); e != nil; e = e.Next() {
			if log, ok := e.Value.(*model.PrivateSceneLog); ok {
				data := &hall_proto.PrivateRoomHistory{
					GameFreeId:  proto.Int32(log.GameFreeId),
					RoomId:      proto.Int32(log.SceneId),
					CreateTime:  proto.Int32(int32(log.CreateTime.Unix())),
					DestroyTime: proto.Int32(int32(log.DestroyTime.Unix())),
					CreateFee:   proto.Int32(log.CreateFee),
				}
				pack.Datas = append(pack.Datas, data)
			}
		}
	}
	proto.SetDefaults(pack)
	p.SendToClient(int(hall_proto.GameHallPacketID_PACKET_SC_GETPRIVATEROOMHISTORY), pack)
}

func (pps *PlayerPrivateScene) PushLog(log *model.PrivateSceneLog) {
	if log == nil {
		return
	}

	y, m, d := log.CreateTime.Date()
	day := y*10000 + int(m)*100 + d
	if lst, exist := pps.logsByDay[day]; exist {
		lst.PushFront(log)
	} else {
		lst = list.New()
		pps.logsByDay[day] = lst
		lst.PushFront(log)
	}
	if !pps.loaded {
		pps.dupLog[log.LogId.Hex()] = struct{}{}
	}
}

func (pps *PlayerPrivateScene) SendPrivateScenes(p *Player) {
	pack := &hall_proto.SCGetPrivateRoomList{}
	for sceneid, s := range pps.scenes {
		data := &hall_proto.PrivateRoomInfo{
			GameFreeId: proto.Int32(s.dbGameFree.GetId()),
			RoomId:     proto.Int(sceneid),
			CurrRound:  proto.Int32(s.currRound),
			MaxRound:   proto.Int32(s.totalRound),
			CurrNum:    proto.Int(len(s.players)),
			MaxPlayer:  proto.Int(s.playerNum),
			CreateTs:   proto.Int32(int32(s.createTime.Unix())),
		}
		pack.Datas = append(pack.Datas, data)
	}
	proto.SetDefaults(pack)
	p.SendToClient(int(hall_proto.GameHallPacketID_PACKET_SC_GETPRIVATEROOMLIST), pack)
}

type PrivateSceneMgr struct {
	pps map[int32]*PlayerPrivateScene
}

func (psm *PrivateSceneMgr) GetOrCreatePlayerPrivateScene(p *Player) *PlayerPrivateScene {
	snid := p.SnId
	if pps, exist := psm.pps[snid]; exist {
		return pps
	}

	pps := &PlayerPrivateScene{
		snid:        snid,
		creatorName: p.Name,
		platform:    p.Platform,
		channel:     p.Channel,
		promoter:    strconv.Itoa(int(p.PromoterTree)),
		packageTag:  p.PackageID,
		scenes:      make(map[int]*Scene),
		logsByDay:   make(map[int]*list.List),
		dupLog:      make(map[string]struct{}),
	}

	psm.pps[snid] = pps
	return pps
}

func (psm *PrivateSceneMgr) GetPlayerPrivateScene(snid int32) *PlayerPrivateScene {
	if pps, exist := psm.pps[snid]; exist {
		return pps
	}
	return nil
}

func (psm *PrivateSceneMgr) OnDestroyScene(scene *Scene) {
	if scene == nil {
		return
	}

	if !scene.IsPrivateScene() {
		return
	}

	pps := psm.GetPlayerPrivateScene(scene.creator)
	if pps != nil {
		if pps.GetScene(scene.sceneId) == scene {
			delete(pps.scenes, scene.sceneId)
			var tax int32
			var returnCoin int32
			p := PlayerMgrSington.GetPlayerBySnId(scene.creator)

			if scene.currRound == 0 && !scene.starting && scene.createFee > 0 { //未开始
				if scene.manualDelete && time.Now().Sub(scene.createTime) < time.Second*time.Duration(model.GameParamData.PrivateSceneFreeDistroySec) { //低于指定时间，要扣除部分费用
					tax = scene.createFee * int32(model.GameParamData.PrivateSceneDestroyTax) / 100
					returnCoin = scene.createFee - tax
				} else {
					returnCoin = scene.createFee
				}
				if returnCoin > 0 {
					if p != nil {
						var remark string
						if tax > 0 {
							remark = fmt.Sprintf("提前解散扣除费用%.02f", float32(tax)/100.0)
						}
						p.AddCoin(int64(returnCoin), common.GainWay_PrivateSceneReturn, "", remark)
					} else {
						//TODO 发送邮件
						//sendClubMail_ClubCreateRoomRefund(scene.creator, scene.limitPlatform.Name, int32(scene.sceneId), int64(tax), int64(returnCoin))
					}
				}
				//if p != nil {
				//	//统计创建房间数量
				//	key := scene.dbGameFree.GetGameDif()
				//	if gd, ok := p.GameData[key]; ok {
				//		gd.CreateRoomTimes--
				//	} else {
				//		p.GameData[key] = &model.PlayerGameStatics{
				//			CreateRoomTimes: 0,
				//		}
				//	}
				//}
			}

			if p != nil {
				pack := &hall_proto.SCDestroyPrivateRoom{
					OpRetCode: hall_proto.OpResultCode_Game_OPRC_Sucess_Game,
					RoomId:    proto.Int(scene.sceneId),
					State:     proto.Int(PrivateSceneState_Deleted),
				}
				proto.SetDefaults(pack)
				p.SendToClient(int(hall_proto.GameHallPacketID_PACKET_SC_DESTROYPRIVATEROOM), pack)
			}

			//写log
			log := model.NewPrivateSceneLog()
			if log != nil {
				log.SnId = pps.snid
				log.Platform = pps.platform
				log.Channel = pps.channel
				log.Promoter = pps.promoter
				log.GameFreeId = scene.dbGameFree.GetId()
				log.SceneId = int32(scene.sceneId)
				log.CreateTime = scene.createTime
				log.DestroyTime = time.Now()
				if returnCoin > 0 {
					log.CreateFee = tax
				} else {
					log.CreateFee = scene.createFee
				}
				//PrivateSceneLogChannelSington.Write(log)
				pps.PushLog(log)
			}

			if pps.CanDelete() && p == nil {
				delete(psm.pps, scene.creator)
			}
		}
	}
}
