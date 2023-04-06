package main

import (
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/message"
	"games.yol.com/win88/srvdata"
	"github.com/globalsign/mgo/bson"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/module"
	"math"
	"strings"
	"time"
)

// 0] "<font color='#FFF200' size='24'>%s</font>"
// 5] "玩家 %s 在 %s 中获得了 %s元 ，进入游戏即可参与！"
// 6] "玩家 在 %s 中获得了爆池奖励 %s元 ，进入游戏即可参与！"
// 10] "恭喜 %s 在 %s 游戏内获得jackpot %s ,中奖金额【 %s】"
// 11] "恭喜 %s 击杀 %s %s 获得 %s奖励"
const (
	//消息类型 0:标示消息内容为服务端拼装好的字符串
	//1-99:根据客户端配置的编号格式化参数组成完整的消息内容(详见:HorseRaceLampMsgType)
	//100:弹幕跑马灯
	HorseRaceLampType_ServerStr  = 0
	HorseRaceLampType_CoinNormal = 5
	HorseRaceLampType_CoinReward = 6
	HorseRaceLampType_CustomMsg  = 100
)

const (
	MAX_GAMENOTICE_PER_PLATFORM = 100
)

var HorseRaceLampMgrSington = &HorseRaceLampMgr{
	HorseRaceLampMsgList:  make(map[string]*HorseRaceLamp),
	HorseRaceLampGameList: make(map[string][]*HorseRaceLamp),
	HorseRaceLampCastList: make(map[string]*HorseRaceLampCastInfo),
	NextGameHorseRaceLamp: make(map[string]int64),
}

type HorseRaceLampMgr struct {
	HorseRaceLampMsgList  map[string]*HorseRaceLamp
	HorseRaceLampGameList map[string][]*HorseRaceLamp
	HorseRaceLampCastList map[string]*HorseRaceLampCastInfo
	NextGameHorseRaceLamp map[string]int64
}

type HorseRaceLamp struct {
	Key           string
	Channel       string
	Title         string
	Content       string
	Footer        string
	StartTime     int64
	Interval      int32
	limitInterval int32
	Count         int32
	LastTime      int64
	Priority      int32
	CreateTime    int64
	MsgType       int32
	Platform      string
	State         int32
	isRob         bool
	Target        []int32
	StandSec      int32
}

type HorseRaceLampCastInfo struct {
	CurIndex  int
	CurTime   int64
	DealQueue []*HorseRaceLamp
	DealList  []*HorseRaceLamp
}

func (this *HorseRaceLampMgr) InitHorseRaceLamp() {
	for _, p := range PlatformMgrSington.Platforms {
		noticeList, err := model.GetAllHorseRaceLamp(p.IdStr)
		if err != nil {
			logger.Logger.Error("InitHorseRaceLamp count failed:", err, noticeList)
		}

		for _, value := range noticeList {
			msg := &HorseRaceLamp{
				Key:           value.Id.Hex(),
				Channel:       value.Channel,
				Content:       value.Content,
				StartTime:     value.StartTime,
				Interval:      value.Interval,
				Count:         value.Count,
				MsgType:       value.MsgType,
				Priority:      value.Priority,
				CreateTime:    value.CreateTime,
				Platform:      value.Platform,
				State:         value.State,
				Target:        value.Target,
				StandSec:      value.StandSec,
				limitInterval: int32(math.Floor(float64(len(value.Content))*0.3)) + 6,
			}

			this.HorseRaceLampMsgList[value.Id.Hex()] = msg
			this.insertToCastMsg(msg)
		}
	}
}

func (this *HorseRaceLampMgr) AddHorseRaceLampMsg(key, ch, p, title, content, footer string, startTime int64, interval, count,
	msgType, state, priority int32, createTime int64, target []int32, standSec int32) string {
	msg := &HorseRaceLamp{
		Key:        key,
		Channel:    ch,
		Title:      title,
		Content:    content,
		Footer:     footer,
		StartTime:  startTime,
		Interval:   interval,
		Count:      count,
		MsgType:    msgType,
		Priority:   priority,
		CreateTime: createTime,
		Platform:   p,
		State:      state,
		Target:     target,
		StandSec:   standSec,
	}
	this.HorseRaceLampMsgList[key] = msg
	this.insertToCastMsg(msg)

	return key
}

func (this *HorseRaceLampMgr) insertToCastMsg(msg *HorseRaceLamp) {
	pKey := msg.Platform
	if this.HorseRaceLampCastList[pKey] == nil {
		this.HorseRaceLampCastList[pKey] = &HorseRaceLampCastInfo{
			CurIndex:  0,
			CurTime:   0,
			DealQueue: []*HorseRaceLamp{},
			DealList:  []*HorseRaceLamp{},
		}
	} else {
		switch msg.MsgType {
		case HorseRaceLampType_CustomMsg:
			this.HorseRaceLampCastList[pKey].DealList = append(this.HorseRaceLampCastList[pKey].DealList, msg)
		default:
			this.HorseRaceLampCastList[pKey].DealQueue = append(this.HorseRaceLampCastList[pKey].DealQueue, msg)
		}
	}
}

func (this *HorseRaceLampMgr) EditHorseRaceLampMsg(hrl *HorseRaceLamp) bool {
	if _, ok := this.HorseRaceLampMsgList[hrl.Key]; !ok {
		return false
	}
	hrl.limitInterval = int32(math.Floor(float64(len(hrl.Content))*0.3)) + 6
	this.HorseRaceLampMsgList[hrl.Key] = hrl

	if pInfo, ok := this.HorseRaceLampCastList[hrl.Platform]; ok && hrl.MsgType != HorseRaceLampType_CustomMsg {
		pInfo.CurTime = 0
		pInfo.CurIndex = 0
	}
	return true
}

func (this *HorseRaceLampMgr) DelHorseRaceLampMsg(key string) {
	if needDel, ok := this.HorseRaceLampMsgList[key]; ok {
		if pInfo, ok := this.HorseRaceLampCastList[needDel.Platform]; ok {
			pInfo.CurTime = 0
			pInfo.CurIndex = 0
			for index := range pInfo.DealQueue {
				if pInfo.DealQueue[index] == needDel {
					pInfo.DealQueue = append(pInfo.DealQueue[:index], pInfo.DealQueue[index+1:]...)
					break
				}
			}
			for index := range pInfo.DealList {
				if pInfo.DealList[index] == needDel {
					pInfo.DealList = append(pInfo.DealList[:index], pInfo.DealList[index+1:]...)
					break
				}
			}
		}
	}
	delete(this.HorseRaceLampMsgList, key)
}

func (this *HorseRaceLampMgr) WordCheck(content string) string {
	has := srvdata.HasSensitiveWord([]rune(content))
	if has {
		content = string(srvdata.ReplaceSensitiveWord([]rune(content))[:])
	}
	return content
}

func (this *HorseRaceLampMgr) IsSensitiveWord(content string) bool {
	return srvdata.HasSensitiveWord([]rune(content))
}

func (this *HorseRaceLampMgr) PushGameHorseRaceLamp(ch, platform, content string, msgType int32, isRob bool, priority int32) bool {
	msg := &HorseRaceLamp{
		Channel:       ch,
		Content:       content,
		Count:         1,
		MsgType:       msgType,
		Platform:      platform,
		Priority:      priority,
		limitInterval: int32(math.Floor(float64(len(content))*0.3)) + 6,
		isRob:         isRob,
	}

	if pool, exist := this.HorseRaceLampGameList[platform]; exist {
		if len(pool) >= model.GameParamData.BacklogGameHorseRaceLamp && len(pool) > 0 {
			minRobPriority := priority
			minPlayerPriority := priority
			bestRobIdx := -1
			bestPlayerIdx := -1
			for k, v := range pool {
				if !isRob {
					if v.isRob { //优先替换机器人的跑马灯
						if v.Priority <= minRobPriority {
							bestRobIdx = k
							minRobPriority = v.Priority
						}
					} else { //其次替换真实玩家的跑马灯
						if v.Priority < minPlayerPriority {
							bestPlayerIdx = k
							minPlayerPriority = v.Priority
						}
					}
				} else {
					if v.isRob {
						if v.Priority < minRobPriority {
							bestRobIdx = k
							minRobPriority = v.Priority
						}
					}
				}
			}
			if bestRobIdx != -1 { //优先替换机器人的跑马灯
				pool = append(pool[:bestRobIdx], pool[bestRobIdx+1:]...)
			} else if bestPlayerIdx != -1 { //其次替换玩家的跑马灯
				pool = append(pool[:bestPlayerIdx], pool[bestPlayerIdx+1:]...)
			} else { //没找到要替换的，直接丢弃自己的
				return false
			}
		}

		if !isRob {
			isFindRob := false
			if len(pool) > 0 {
				for k, v := range pool {
					if v.isRob {
						pool[k] = msg
						isFindRob = true
						break
					}
				}
			}
			if isFindRob {
				this.HorseRaceLampGameList[platform] = pool
			} else {
				this.HorseRaceLampGameList[platform] = append(pool, msg)
			}
		} else {
			this.HorseRaceLampGameList[platform] = append(pool, msg)
		}
	} else {
		this.HorseRaceLampGameList[platform] = []*HorseRaceLamp{msg}
	}
	return true
}
func (this *HorseRaceLampMgr) DealHorseRaceLamp(uTime int64, value *HorseRaceLamp) {
	if value.Count > 0 {
		this.BroadcastHorseRaceLampMsg(value)
		value.Count = value.Count - 1
		value.LastTime = uTime
		value.StartTime = value.LastTime + int64(value.Interval)
		if value.Count <= 0 {
			this.DelHorseRaceLampMsg(value.Key)
			model.RemoveHorseRaceLamp(value.Key, value.Platform)
		}
	} else {
		this.BroadcastHorseRaceLampMsg(value)
		value.LastTime = uTime
		value.StartTime = value.LastTime + int64(value.Interval)
	}
}

// //////////////////////////////////////////////////////////////////
// / Module Implement [HorseRaceLampMgr]
// //////////////////////////////////////////////////////////////////
func (this *HorseRaceLampMgr) ModuleName() string {
	return "HorseRaceLampMgr"
}

func (this *HorseRaceLampMgr) Init() {
}

func (this *HorseRaceLampMgr) Update() {
	uTime := time.Now().Unix()
	//调整了跑马灯功能，需要排队发送
	for _, v := range this.HorseRaceLampCastList {
		if uTime > v.CurTime {
			if v.CurIndex < len(v.DealQueue) {
				value := v.DealQueue[v.CurIndex]
				//需要跳过
				if value.State != 0 {
					v.CurIndex += 1
					if v.CurIndex >= len(v.DealQueue) {
						v.CurIndex = 0
					}
				} else if uTime > value.StartTime && value.State == 0 {
					this.DealHorseRaceLamp(uTime, value)
					v.CurIndex += 1
					if v.CurIndex >= len(v.DealQueue) {
						v.CurIndex = 0
					}
					v.CurTime = uTime + int64(value.limitInterval)
				}
			}
		}
	}
	for _, nc := range this.HorseRaceLampCastList {
		for _, value := range nc.DealList {
			if uTime > value.StartTime && value.State == 0 {
				this.DealHorseRaceLamp(uTime, value)
			}
		}
	}
	for name, pool := range this.HorseRaceLampGameList {
		if len(pool) > 0 {
			msg := pool[0]
			nextTs := this.NextGameHorseRaceLamp[name]
			if uTime >= nextTs {
				this.HorseRaceLampGameList[name] = pool[1:]
				if msg != nil {
					this.BroadcastHorseRaceLampMsg(msg)
					this.NextGameHorseRaceLamp[name] = uTime + int64(msg.limitInterval)
				}
			}
		}
	}
}

func (this *HorseRaceLampMgr) SaveHorseRaceLamp() {
	for _, hrl := range this.HorseRaceLampMsgList {
		model.EditHorseRaceLamp(&model.HorseRaceLamp{
			Id:         bson.ObjectIdHex(hrl.Key),
			Channel:    "",
			Title:      hrl.Title,
			Content:    hrl.Content,
			Footer:     hrl.Footer,
			StartTime:  hrl.StartTime,
			Interval:   hrl.Interval,
			Count:      hrl.Count,
			CreateTime: hrl.CreateTime,
			Priority:   hrl.Priority,
			MsgType:    hrl.MsgType,
			Platform:   hrl.Platform,
			State:      hrl.State,
			Target:     hrl.Target,
			StandSec:   hrl.StandSec,
		})
	}
}

func (this *HorseRaceLampMgr) Shutdown() {
	this.SaveHorseRaceLamp()
	module.UnregisteModule(this)
}

func (this *HorseRaceLampMgr) BroadcastHorseRaceLampMsg(horseRaceLamp *HorseRaceLamp) {
	if horseRaceLamp.MsgType == HorseRaceLampType_CustomMsg {
		logger.Logger.Infof(">>>>>>>弹幕>>>>>>>>(this *HorseRaceLampMgr) BroadcastHorseRaceLampMsg content:%v msgType:%v "+
			"target:%v standSec:%v", horseRaceLamp.Content, horseRaceLamp.MsgType, horseRaceLamp.Target, horseRaceLamp.StandSec)
	}
	var rawpack = &message.SCNotice{
		Count:     proto.Int(1),
		MsgType:   proto.Int32(horseRaceLamp.MsgType),
		Ts:        proto.Int64(time.Now().Unix()), //发送时间
		ChannelId: proto.String(horseRaceLamp.Channel),
		Platform:  proto.String(horseRaceLamp.Platform),
		StandSec:  proto.Int32(horseRaceLamp.StandSec),
	}
	if horseRaceLamp.MsgType == 0 {
		rawpack.Params = append(rawpack.Params, &message.NoticeParam{StrParam: proto.String(horseRaceLamp.Content)})
	} else if horseRaceLamp.MsgType > 0 && horseRaceLamp.MsgType < 100 {
		strArr := strings.Split(horseRaceLamp.Content, "|")
		for _, value := range strArr {
			rawpack.Params = append(rawpack.Params, &message.NoticeParam{StrParam: proto.String(value)})
		}
	} else if horseRaceLamp.MsgType == 100 {
		rawpack.Params = append(rawpack.Params, &message.NoticeParam{StrParam: proto.String(horseRaceLamp.Title)})
		rawpack.Params = append(rawpack.Params, &message.NoticeParam{StrParam: proto.String(horseRaceLamp.Content)})
		rawpack.Params = append(rawpack.Params, &message.NoticeParam{StrParam: proto.String(horseRaceLamp.Footer)})
	}

	proto.SetDefaults(rawpack)
	if len(horseRaceLamp.Target) == 0 {
		PlayerMgrSington.BroadcastMessageToPlatform(horseRaceLamp.Platform, int(message.MSGPacketID_PACKET_SC_NOTICE), rawpack)
	} else {
		PlayerMgrSington.BroadcastMessageToTarget(horseRaceLamp.Platform, horseRaceLamp.Target, int(message.MSGPacketID_PACKET_SC_NOTICE), rawpack)
	}
}

func init() {
	module.RegisteModule(HorseRaceLampMgrSington, time.Second*3, 0)
	////使用并行加载
	//RegisteParallelLoadFunc("平台通知", func() error {
	//	HorseRaceLampMgrSington.InitHorseRaceLamp()
	//	return nil
	//})
}
