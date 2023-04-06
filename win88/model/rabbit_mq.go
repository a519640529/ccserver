package model

import (
	"fmt"
	"games.yol.com/win88/common"
	"github.com/idealeak/goserver/core/logger"
	"strconv"
	"time"
)

type mq_name string

const (
	MQ_NAME_PlayerEvent  mq_name = "evt_player"
	MQ_NAME_SystemGive   mq_name = "evt_systemgive"
	MQ_NAME_PlayerOnline mq_name = "evt_online"
	MQ_NAME_PlayerLogin  mq_name = "evt_login"
	MQ_NAME_GameEvent    mq_name = "evt_gamerec"
	MQ_NAME_EnterEvent   mq_name = "evt_gameentry"

	//MQ_NAME_BindEvent    mq_name = "evt_bind"
	//MQ_NAME_ActMonitorEvent mq_name = "evt_actmonitor" //用户行为事件(登录,兑换,游戏)
	//MQ_NAME_SpreadAccount mq_name = "evt_spreadaccount"
	//MQ_NAME_TaxDivide     mq_name = "evt_taxdivide"
	//MQ_NAME_MatchEvent       mq_name = "evt_match"       //比赛事件
	//MQ_NAME_MatchSignupEvent mq_name = "evt_matchsignup" //比赛报名事件
	//MQ_NAME_PlayerInfoEvent  mq_name = "evt_playerinfo"  //用户分层数据首次同步事件
)

// 首次登录通知
func GeneratePlayerEvent(event int, platform, packageTag string, snid int32, channel string, promoter string, promoterTree int32, isCreate int, isNew int, isBind int, appId string) *RabbitMQData {
	params := make(map[string]string)
	params["event"] = strconv.Itoa(event)
	params["snid"] = strconv.Itoa(int(snid))
	params["platform"] = platform
	params["channel"] = channel
	params["promoter"] = promoter
	params["promoter_tree"] = strconv.Itoa(int(promoterTree))
	params["packageTag"] = packageTag
	params["isCreate"] = strconv.Itoa(isCreate)
	params["isNew"] = strconv.Itoa(isNew)
	params["isBind"] = strconv.Itoa(isBind)
	params["create_time"] = strconv.Itoa(int(time.Now().Unix()))
	return NewRabbitMQData(string(MQ_NAME_PlayerEvent), params)
}

////流水推送
//func GenerateSpreadAccount(appId string, gamefreeId int32, data []*webapi.PlayerStatement) *RabbitMQData {
//	d, err := json.Marshal(data)
//	if err != nil {
//		logger.Error(err)
//		return nil
//	}
//	params := make(map[string]string)
//	params["gamefreeId"] = strconv.Itoa(int(gamefreeId))
//	params["data"] = string(d[:])
//	params["ts"] = strconv.Itoa(int(time.Now().Unix()))
//	return NewRabbitMQData(string(MQ_NAME_SpreadAccount), params)
//}

////税收分成
//func GenerateTaxDivide(snid int32, platform, channel, promoter, packageTag string, tax, taxex, validFlow int64, gameid, gamemode int, gamefreeid, promoterTree int32) *RabbitMQData {
//	params := make(map[string]string)
//	params["snid"] = strconv.Itoa(int(snid))
//	params["platform"] = platform
//	params["tax"] = fmt.Sprintf("%v", tax)
//	params["taxex"] = fmt.Sprintf("%v", taxex)
//	params["validflow"] = fmt.Sprintf("%v", validFlow)
//	params["channel"] = channel
//	params["promoter"] = promoter
//	params["promoter_tree"] = strconv.Itoa(int(promoterTree))
//	params["packageTag"] = packageTag
//	params["gameId"] = strconv.Itoa(gameid)
//	params["modeId"] = strconv.Itoa(gamemode)
//	params["gamefreeId"] = strconv.Itoa(int(gamefreeid))
//	params["ts"] = strconv.Itoa(int(time.Now().Unix()))
//	return NewRabbitMQData(string(MQ_NAME_TaxDivide), params)
//}

// 系统赠送
func GenerateSystemGive(snid int32, platform, channel, promoter string, ammount, tag int32, appId string) *RabbitMQData {
	params := make(map[string]string)
	params["snid"] = strconv.Itoa(int(snid))
	params["platform"] = platform
	params["channel"] = channel
	params["promoter"] = promoter
	params["amount"] = fmt.Sprintf("%v", ammount)
	params["tg"] = strconv.Itoa(int(tag))
	params["ts"] = strconv.Itoa(int(time.Now().Unix()))
	return NewRabbitMQData(string(MQ_NAME_SystemGive), params)
}

// 在线统计
func GenerateOnline(online map[string]int) *RabbitMQData {
	m := map[int]int{}
	for k, v := range online {
		i, _ := strconv.Atoi(k)
		m[i] = v
	}
	params := make(map[string]interface{})
	params["Online"] = m
	params["Time"] = time.Now().Unix()
	return NewRabbitMQData(string(MQ_NAME_PlayerOnline), params)
}

// 玩家登陆
func GenerateLogin(o *PlayerLoginEvent) *RabbitMQData {
	return NewRabbitMQData(string(MQ_NAME_PlayerLogin), o)
}

// 游戏事件
func GenerateGameEvent(o *PlayerGameRecEvent) *RabbitMQData {
	return NewRabbitMQData(string(MQ_NAME_GameEvent), o)
}

type PlayerBindEvent struct {
	SnId       int32     //用户ID
	Platform   int32     //平台
	OS         int       //0 Windows 1 Android 2 iOS
	CreateTime time.Time //创建时间 RFC3339 创建时间和登陆时间相同的话就认为是登陆当天创建的
	BindTime   time.Time //绑定日期 RFC3339 绑定时间和登陆时间相同的话就认为是登陆当天绑定的
}

// 绑定手机号
//func GenerateBindEvent(o *PlayerBindPhoneEvent) *RabbitMQData {
//	m := &PlayerBindEvent{
//		SnId: o.SnId,
//	}
//	platform, err := strconv.Atoi(o.Platform)
//	if err != nil {
//		logger.Error(err)
//		return nil
//	}
//	m.Platform = int32(platform)
//	m.OS = common.DeviceNum[o.OS]
//	if o.CreateTime > 0 {
//		m.CreateTime = time.Unix(0, o.CreateTime)
//	}
//	if o.BindTime > 0 {
//		m.BindTime = time.Unix(0, o.BindTime)
//	}
//	return NewRabbitMQData(string(MQ_NAME_BindEvent), m)
//}

type PlayerGameEntryEvent struct {
	RecordId string //游戏记录ID
	SnId     int32  //用户ID
	Platform int32  //平台
	OS       int    //0 Windows 1 Android 2 iOS
	GameId   int    //游戏id
	ModeId   int    //游戏模式
	Time     int64  //入场时间 RFC3339
	Id       int32  //游戏id
}

// 进入场次
func GenerateEnterEvent(recordId string, snId int32, platform, os string, gameId, modeId int, gameFreeId int32) *RabbitMQData {
	m := &PlayerGameEntryEvent{
		RecordId: recordId,
		SnId:     snId,
		GameId:   gameId,
		ModeId:   modeId,
		Time:     time.Now().Unix(),
		Id:       gameFreeId,
	}
	pf, err := strconv.Atoi(platform)
	if err != nil {
		logger.Error(err)
		return nil
	}
	m.Platform = int32(pf)
	m.OS = common.DeviceNum[os]
	return NewRabbitMQData(string(MQ_NAME_EnterEvent), m)
}

////比赛详情
//type MatchMQLog struct {
//	Platform     int32                    //平台编号
//	MatchId      int32                    //比赛编号
//	MatchName    string                   //比赛名称
//	GameFreeId   int32                    //游戏id
//	TimeSpend    int32                    //比赛用时
//	TotalOfGames int32                    //牌局数
//	EndTime      time.Time                //结束时间
//	Players      []*MatchPlayer           //参赛人员数据
//	MatchProcess []*MatchProcessTimeSpend //赛程
//	GameLogs     []*MatchGameLog          //牌局记录
//}

//// 比赛记录
//func GenerateMatchEvent(log *MatchLog) *RabbitMQData {
//	m := &MatchMQLog{
//		MatchId:      log.MatchId,
//		MatchName:    log.MatchName,
//		GameFreeId:   log.GameFreeId,
//		TimeSpend:    int32(log.EndTime.Sub(log.StartTime) / time.Second),
//		TotalOfGames: int32(len(log.GameLogs)),
//		EndTime:      log.EndTime,
//		Players:      log.Players,
//		MatchProcess: log.TimeSpend,
//		GameLogs:     log.GameLogs,
//	}
//	pf, err := strconv.Atoi(log.Platform)
//	if err != nil {
//		logger.Error(err)
//		return nil
//	}
//	m.Platform = int32(pf)
//	return NewRabbitMQData(string(MQ_NAME_MatchEvent), m)
//}

////比赛报名事件
//type MatchSignupLog struct {
//	Platform   int32     //平台编号
//	SnId       int32     //玩家id
//	MatchId    int32     //比赛编号
//	MatchName  string    //比赛名称
//	GameFreeId int32     //游戏id
//	WaitSec    int32     //报名等待用时 单位秒
//	Time       time.Time //退赛时间
//}
//
//// 比赛报名事件记录
//func GenerateMatchSignupEvent(snid int32, platform, matchName string, matchid, gamefreeid int32, signupTs int64) *RabbitMQData {
//	m := &MatchSignupLog{
//		MatchId:    matchid,
//		MatchName:  matchName,
//		GameFreeId: gamefreeid,
//		SnId:       snid,
//		Time:       time.Now(),
//	}
//	m.WaitSec = int32(m.Time.Unix() - signupTs)
//	pf, err := strconv.Atoi(platform)
//	if err != nil {
//		logger.Error(err)
//		return nil
//	}
//	m.Platform = int32(pf)
//	return NewRabbitMQData(string(MQ_NAME_MatchSignupEvent), m)
//}

//type ActMonitorEvent struct {
//	EventType int32  //1.登录 2.兑换 3.游戏
//	SnId      int32  //用户ID
//	Name      string //玩家名字
//	Platform  string //平台
//	Ts        int64  //事件时间
//	ReMark    string //备注
//	//兑换
//	ExchangeCoin int64 //兑换金额
//	BillNo       int64 //账单ID
//	//游戏
//	GameSceneName string //游戏场次名称(炸金花初级场)
//	State         int32  //0.进入 1.离开
//}
//
//// 用户行为事件
//func GenerateActMonitorEvent(eventType int32, snid int32, name, platform string, signupTs, billNo, exchangeCoin int64,
//	gameSceneName string, state int32, reMark string) *RabbitMQData {
//	m := &ActMonitorEvent{
//		EventType:     eventType,
//		SnId:          snid,
//		Name:          name,
//		Platform:      platform,
//		Ts:            signupTs,
//		ReMark:        reMark,
//		BillNo:        billNo,
//		ExchangeCoin:  exchangeCoin,
//		GameSceneName: gameSceneName,
//		State:         state,
//	}
//	return NewRabbitMQData(string(MQ_NAME_ActMonitorEvent), m)
//}

//type PlayerInfoEvent struct {
//	Platform    string `json:"platform,omitempty"`
//	Snid        int32  `json:"snid,omitempty"`
//	Is_pay      bool   `json:"is_pay,omitempty"`
//	Is_exchange bool   `json:"is_exchange,omitempty"`
//	Is_child    bool   `json:"is_child,omitempty"`
//	Is_group    bool   `json:"is_group,omitempty"`
//}
//
//// 用户分层数据首次同步事件
//func GeneratePlayerInfoEvent(snid int32, platform string) *RabbitMQData {
//	pie := &PlayerInfoEvent{
//		Platform:    platform,
//		Snid:        snid,
//		Is_pay:      true,
//		Is_exchange: true,
//		Is_child:    true,
//		Is_group:    true,
//	}
//	return NewRabbitMQData(string(MQ_NAME_PlayerInfoEvent), pie)
//}
