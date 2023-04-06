package main

import (
	"games.yol.com/win88/model"
	"github.com/idealeak/goserver/core/logger"
	"math"
	"sort"
)

const (
	ActState_Login    int32 = 1 << iota //登录.1
	ActState_Exchange                   //兑换.2
	ActState_Game                       //游戏.3
	ActState_Max
)

var ActMonitorMgrSington = &ActMonitorMgr{
	ActMonitorList: make(map[int64]*ActMonitorInfo),
}

type ActMonitorInfo struct {
	SeqNo       int64
	SnId        int32
	Platform    string //平台
	MonitorType int32  //二进制 1.登录 2.兑换 3.游戏
	CreateTime  int64  //创建时间
	Creator     string //创建者
	ReMark      string //备注
	GameName    string //当前所在游戏名字
	State       int    //玩家状态 0.全部 1.不在线 2.在线 3.游戏中
}
type ActMonitorMgr struct {
	ActMonitorList map[int64]*ActMonitorInfo
	NowActSeqNo    int64
}

// monitorType 自己的类型  flag 当前触发的类型
func (u *ActMonitorMgr) IsMarkFlag(monitorType, flag int32) bool {
	if (monitorType & flag) != 0 {
		return true
	}
	return false
}
func (u *ActMonitorMgr) Init() {
	actMonitorData := model.GetAllActMonitorData()
	for _, info := range actMonitorData {
		ami := &ActMonitorInfo{
			SeqNo:       info.SeqNo,
			SnId:        info.SnId,
			Platform:    info.Platform,
			MonitorType: info.MonitorType,
			CreateTime:  info.CreateTime,
			Creator:     info.Creator,
			ReMark:      info.ReMark,
		}
		if u.NowActSeqNo < info.SeqNo {
			u.NowActSeqNo = info.SeqNo
		}
		u.ActMonitorList[info.SeqNo] = ami
	}
}

type ActMonitorList struct {
	PageNo   int
	PageSize int
	PageSum  int
	TotalSum int
	Data     []*ActMonitorInfo
}

func (u *ActMonitorMgr) QueryAMIList(pageNo, pageSize int, platform string, snid, startTs, endTs, state int) *ActMonitorList {
	if len(u.ActMonitorList) == 0 {
		return nil
	}
	var amiList = make([]*ActMonitorInfo, 0)
	for _, v := range u.ActMonitorList {
		if len(platform) != 0 && v.Platform != platform {
			continue
		}
		if snid != 0 && v.SnId != int32(snid) {
			continue
		}
		if startTs != 0 && endTs != 0 && (v.CreateTime < int64(startTs) || v.CreateTime > int64(endTs)) {
			continue
		}
		if state != 0 && v.State != state {
			continue
		}
		amiList = append(amiList, v)
	}
	sort.Slice(amiList, func(i, j int) bool {
		if amiList[i].SeqNo > amiList[j].SeqNo {
			return true
		}
		return false
	})
	totalNum := len(amiList)                                         //总条目
	pageSum := int(math.Ceil(float64(totalNum) / float64(pageSize))) //总页数
	if pageNo <= 0 || pageNo > pageSum {
		pageNo = 1 //当前页
	}
	start := (pageNo - 1) * pageSize
	end := start + pageSize
	if totalNum > start {
		if totalNum < end {
			end = totalNum
		}
		amiList = amiList[start:end]
	}
	for k, v := range amiList {
		actPlayer := amiList[k]
		actPlayer.GameName = ""
		p := PlayerMgrSington.GetPlayerBySnId(v.SnId)
		if p != nil {
			if p.IsOnLine() {
				actPlayer.State = 2
			} else {
				actPlayer.State = 1
			}
			if p.scene != nil {
				actPlayer.State = 3
				actPlayer.GameName = p.scene.dbGameFree.GetName() + p.scene.dbGameFree.GetTitle()
			}
		} else {
			actPlayer.State = 1
		}
	}
	return &ActMonitorList{pageNo, pageSize, pageSum, totalNum, amiList}
}
func (u *ActMonitorMgr) Edit(amt *ActMonitorInfo) {
	u.ActMonitorList[amt.SeqNo] = amt
}
func (u *ActMonitorMgr) Del(seqNo int64) {
	delete(u.ActMonitorList, seqNo)
}
func (u *ActMonitorMgr) AddSeqNo() int64 {
	u.NowActSeqNo++
	return u.NowActSeqNo
}
func (u *ActMonitorMgr) GetSeqNo(snid int32, platform string) int64 {
	for _, v := range u.ActMonitorList {
		if v.SnId == snid && v.Platform == platform {
			return v.SeqNo
		}
	}
	return -1
}
func (u *ActMonitorMgr) SendActMonitorEvent(eventType, snid int32, name, platform string, billNo, exchangeCoin int64,
	gameSceneName string, state int32) {
	logger.Logger.Tracef("SendActMonitorEvent eventType:%v snid:%v name:%v platform:%v billNo:%v exchangeCoin:%v "+
		"gameSceneName:%v state:%v", eventType, snid, name, platform, billNo, exchangeCoin, gameSceneName, state)
	//seqNo := u.GetSeqNo(snid, platform)
	//if data, ok := u.ActMonitorList[seqNo]; ok {
	//	if u.IsMarkFlag(eventType, data.MonitorType) {
	//		var flag int32
	//		if eventType == ActState_Login {
	//			flag = 1
	//		} else if eventType == ActState_Exchange {
	//			flag = 2
	//		} else if eventType == ActState_Game {
	//			flag = 3
	//		}
	//		logger.Logger.Tracef("GenerateActMonitorEvent "+
	//			"flag:%v eventType:%v snid:%v name:%v platform:%v billNo:%v exchangeCoin:%v "+
	//			"gameSceneName:%v state:%v reMark:%v",
	//			flag, eventType, snid, name, platform, billNo, exchangeCoin, gameSceneName, state, data.ReMark)
	//		LogChannelSington.WriteMQData(model.GenerateActMonitorEvent(flag, snid, name, platform,
	//			time.Now().Unix(), billNo, exchangeCoin, gameSceneName, state, data.ReMark))
	//	}
	//}
}
func init() {
	//RegisteParallelLoadFunc("用户行为监控列表", func() error {
	//	ActMonitorMgrSington.Init()
	//	return nil
	//})
}
