package model

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/jinzhu/now"
)

const (
	TimeFormat_DataEvent = "2006-01-02 15:04:05"
)

const (
	DataEventEx_PlayerLogin      int32 = iota //玩家登录事件
	DataEventEx_PlayerBindPhone               //玩家绑定手机事件
	DataEventEx_PlayerBindAlipay              //玩家绑定alipay事件
	DataEventEx_PlayerGameRec                 //玩家游戏事件
	DataEventEx_PlayerGameRecPay              //玩家充值游戏事件
	DataEventEx_Bankruptcy                    //破产统计事件
	DataEventEx_PlayerPay                     //玩家充值事件
	DataEventEx_SystemGive                    //系统赠送事件
)

// 写入队列的数据
type RabbitMQDataRaw struct {
	Source int32
	Data   interface{}
}

// 冰河世纪解析的数据
type IceAgeGameNoteData struct {
	Source int32
	Data   *IceAgeType
}

// 复仇者联盟解析的数据
type AvengersGameNoteData struct {
	Source int32
	Data   *GameResultLog
}

//// 复仇者联盟解析的数据
//type AvengersGameNoteData struct {
//	Source int32
//	Data   *AvengersType
//}

// 财神解析的数据
type CaiShenGameNoteData struct {
	Source int32
	Data   *CaiShenType
}

// 百战成神解析的数据
type TamQuocGameNoteData struct {
	Source int32
	Data   *TamQuocType
}

// 复活岛解析的数据
type EasterIslandGameNoteData struct {
	Source int32
	Data   *EasterIslandType
}

// 糖果解析的数据
type CandyGameNoteData struct {
	Source int32
	Data   *CandyType
}

// MiniPoker解析的数据
type MiniPokerGameNoteData struct {
	Source int32
	Data   *MiniPokerType
}

// CaoThap解析的数据
type CaoThapGameNoteData struct {
	Source int32
	Data   *CaoThapType
}

// 幸运骰子解析的数据
type LuckyDiceGameNoteData struct {
	Source int32
	Data   *LuckyDiceType
}

// 在线统计
type PlayerOnlineEvent struct {
	Online map[int]int
	Time   time.Time
}

func MarshalPlayerOnlineEvent(source int32, online map[string]int) (data string, err error) {
	m := map[int]int{}
	for k, v := range online {
		i, _ := strconv.Atoi(k)
		m[i] = v
	}
	raw := &RabbitMQDataRaw{
		Source: source,
		Data: &PlayerOnlineEvent{
			Online: m,
			Time:   time.Now(),
		},
	}
	d, err := json.Marshal(raw)
	if err != nil {
		return
	}
	return string(d), nil
}

// 玩家登录
type PlayerLoginEvent struct {
	SnId              int32  //用户ID
	Channel           string //渠道
	Promoter          string //推广
	Platform          string //平台
	City              string //城市
	OS                string //操作系统
	TelephonePromoter int32  //电销
	CreateTime        int64  //创建时间
	CreateDayTime     int64  //创建时间0点
	LoginTime         int64  //登录时间
	UpgradeTime       int64  //升级账号时间
	LastLoginIP       string //登录ip
	IsBindPhone       int32  //是否绑定过手机号
	IsNew             int32  //是否是新用户，1是 0否
	DeviceId          string //设备id
}

func CreatePlayerLoginEvent(snid int32, channel, promoter, platform, city, os, ip string, createTime,
	upgradeTime time.Time, isBindPhone int32, telephonePromoter int32, deviceId string) *PlayerLoginEvent {
	isNew := int32(0)
	if createTime.Local().YearDay() == time.Now().Local().YearDay() && createTime.Local().Year() == time.Now().Local().Year() {
		isNew = 1
	}
	return &PlayerLoginEvent{
		SnId:              snid,
		Channel:           channel,
		Promoter:          promoter,
		Platform:          platform,
		City:              city,
		OS:                os,
		TelephonePromoter: telephonePromoter,
		CreateTime:        createTime.Local().Unix(),
		CreateDayTime:     now.New(createTime).BeginningOfDay().Local().Unix(),
		LoginTime:         time.Now().Local().Unix(),
		UpgradeTime:       upgradeTime.Local().Unix(),
		LastLoginIP:       ip,
		IsBindPhone:       isBindPhone,
		IsNew:             isNew,
		DeviceId:          deviceId,
	}
}

func MarshalPlayerLoginEvent(source, snid int32, channel, promoter, platform, city, os, ip string,
	createTime, upgradeTime time.Time, isBindPhone int32, telephonePromoter int32, deviceId string) (data string, err error) {
	raw := &RabbitMQDataRaw{
		Source: source,
		Data: CreatePlayerLoginEvent(snid, channel, promoter, platform, city, os, ip, createTime,
			upgradeTime, isBindPhone, telephonePromoter, deviceId),
	}

	d, e := json.Marshal(raw)
	if e == nil {
		data = string(d[:])
	}
	err = e
	return
}

// 用户升级账号
type PlayerBindPhoneEvent struct {
	SnId              int32  //用户ID
	Channel           string //渠道
	Promoter          string //推广
	Platform          string //平台
	City              string //城市
	OS                string //操作系统
	Value             int32  //占位用
	TelephonePromoter int32  //电销
	CreateTime        int64  //创建日期
	BindTime          int64  //绑定日期
}

func CreatePlayerBindPhoneEvent(snid int32, channel, promoter, platform, city, os string,
	createTime time.Time, telephonePromoter int32) *PlayerBindPhoneEvent {
	return &PlayerBindPhoneEvent{
		SnId:              snid,
		Channel:           channel,
		Promoter:          promoter,
		TelephonePromoter: telephonePromoter,
		Platform:          platform,
		City:              city,
		OS:                os,
		Value:             1,
		CreateTime:        createTime.Unix(),
		BindTime:          time.Now().Unix(),
	}
}

//func MarshalPlayerBindPhoneEvent(source, snid int32, channel, promoter, platform, city, os string,
//	createTime time.Time, telephonePromoter int32) (data string, err error) {
//	raw := &RabbitMQDataRaw{
//		Source: source,
//		Data:   CreatePlayerBindPhoneEvent(snid, channel, promoter, platform, city, os, createTime, telephonePromoter),
//	}
//	d, e := json.Marshal(raw)
//	if e == nil {
//		data = string(d[:])
//	}
//	err = e
//	return
//}

// 用户升级账号
type PlayerBindAlipayEvent struct {
	SnId              int32  //用户ID
	Channel           string //渠道
	Promoter          string //推广
	TelephonePromoter int32  //电销
	Platform          string //平台
	City              string //城市
	OS                string //操作系统
	Value             int32  //占位用
	BindTime          int64  //绑定日期
}

func MarshalPlayerBindAlipayEvent(source, snid int32, channel, promoter, platform, city, os string, telephonePromoter int32) (data string, err error) {
	raw := &RabbitMQDataRaw{
		Source: source,
		Data: &PlayerBindAlipayEvent{
			SnId:              snid,
			Channel:           channel,
			Promoter:          promoter,
			Platform:          platform,
			TelephonePromoter: telephonePromoter,
			City:              city,
			OS:                os,
			Value:             1,
			BindTime:          time.Now().Local().Unix(),
		},
	}
	d, e := json.Marshal(raw)
	if e == nil {
		data = string(d[:])
	}
	err = e
	return
}

// 玩家游戏记录
type PlayerGameRecEvent struct {
	RecordId          string //游戏记录ID
	SnId              int32  //用户ID
	Channel           string //渠道
	Promoter          string //推广
	Platform          string //平台
	City              string //城市
	OS                string //操作系统
	TelephonePromoter int32  //电销标记
	GameId            int32  //游戏id
	ModeId            int32  //游戏模式
	Tax               int64  //税收
	//Taxex             int64     //税收2
	Amount        int64  //金币变化（正值为赢；负值为输）
	CreateTime    int64  //创建时间
	CreateDayTime int64  //账号创建时间0点
	ValidBet      int64  //有效下注数量
	ValidFlow     int64  //有效流水数量
	Out           int64  //产出
	In            int64  //投入
	IsNew         int32  //是否是新人
	GameFreeID    int32  //游戏freeid
	GamingTime    int32  //游戏开始到玩家结算的时长 单位：秒
	FirstTime     int64  //首次玩该场次游戏时间
	PlayTimes     int64  //该场次游戏次数
	FirstGameTime int64  //首次玩游戏时间
	PlayGameTimes int64  //该游戏总次数
	LastLoginTime int64  //最后登录时间
	DeviceId      string //设备id
}

func CreatePlayerGameRecEvent(snid int32, tax, taxex, amount, validbet, validflow, in, out int64, gameid, gameFreeId, modeid int32, recordId, channel, promoter,
	platform, city, os string, createDayTime time.Time, gamingTime int32, firstGameFreeTime, firstGameTime time.Time,
	playGameFreeTimes, playerGameTimes int64, lastLoginTime time.Time, teleponePromoter int32, deviceId string) *PlayerGameRecEvent {
	isNewbie := int32(0)
	tCreateDay := now.New(createDayTime).BeginningOfDay()
	if now.BeginningOfDay().Equal(tCreateDay) {
		isNewbie = 1
	}
	if gamingTime < 0 {
		gamingTime = 0
	}
	return &PlayerGameRecEvent{RecordId: recordId,
		SnId:              snid,
		Channel:           channel,
		Promoter:          promoter,
		TelephonePromoter: teleponePromoter,
		Platform:          platform,
		City:              city,
		OS:                os,
		GameId:            gameid,
		ModeId:            modeid,
		Tax:               tax,
		//Taxex:             taxex,
		Amount:        amount,
		ValidBet:      validbet,
		ValidFlow:     validflow,
		In:            in,
		Out:           out,
		CreateTime:    time.Now().Local().Unix(),
		CreateDayTime: tCreateDay.Local().Unix(),
		IsNew:         isNewbie,
		GameFreeID:    gameFreeId,
		GamingTime:    gamingTime,
		FirstTime:     firstGameFreeTime.Unix(),
		FirstGameTime: firstGameTime.Unix(),
		PlayTimes:     playGameFreeTimes,
		PlayGameTimes: playerGameTimes,
		LastLoginTime: lastLoginTime.Unix(),
		DeviceId:      deviceId}
}

func MarshalPlayerGameRecEvent(source, snid int32, tax, taxex, amount, validbet, validflow, in, out int64, gameid, gameFreeId, modeid int32, recordId, channel, promoter,
	platform, city, os string, createDayTime time.Time, gamingTime int32, firstGameFreeTime time.Time,
	playGameFreeTimes int64, lastLoginTime time.Time, telephonePromoter int32, firstGameTime time.Time,
	playGameTimes int64, deviceId string) (data string, err error) {
	raw := &RabbitMQDataRaw{
		Source: source,
		Data: CreatePlayerGameRecEvent(snid, tax, taxex, amount, validbet, validflow, in, out, gameid, gameFreeId, modeid, recordId, channel, promoter,
			platform, city, os, createDayTime, gamingTime, firstGameFreeTime, firstGameTime, playGameFreeTimes, playGameTimes, lastLoginTime, telephonePromoter, deviceId),
	}
	d, e := json.Marshal(raw)
	if e == nil {
		data = string(d[:])
	}
	err = e
	return
}

// 玩家游戏记录
type PlayerGameRecPayEvent struct {
	SnId              int32  //用户ID
	Channel           string //渠道
	Promoter          string //推广
	Platform          string //平台
	City              string //城市
	OS                string //操作系统
	TelephonePromoter int32  //电销标签
	IsNew             int32  //是否新人
	IsPay             int32  //是否付费
	IsGame            int32  //是否游戏
	CreateTime        int64  //记录创建时间
	CreateDayTime     int64  //记录创建时间0点
	Time              int64  //当前时间
	RegisterDayTime   int64  //玩家注册时间
}

func MarshalPlayerGameRecPayEvent(source, snid, isPay, isGame int32, channel, promoter, platform, city, os string,
	createDayTime time.Time, orderCreateTime int64, telephonePromoter int32) (data string, err error) {
	isNewbie := int32(0)
	if now.BeginningOfDay().Equal(now.New(createDayTime).BeginningOfDay()) {
		isNewbie = 1
	}
	tNow := time.Now()
	raw := &RabbitMQDataRaw{
		Source: source,
		Data: &PlayerGameRecPayEvent{
			SnId:              snid,
			Channel:           channel,
			Promoter:          promoter,
			Platform:          platform,
			City:              city,
			OS:                os,
			IsNew:             isNewbie,
			TelephonePromoter: telephonePromoter,
			IsPay:             isPay,
			IsGame:            isGame,
			RegisterDayTime:   createDayTime.Local().Unix(),
			CreateTime:        time.Unix(orderCreateTime, 0).Local().Unix(),
			CreateDayTime:     now.New(time.Unix(orderCreateTime, 0)).BeginningOfDay().Local().Unix(),
			Time:              tNow.Local().Unix(),
		},
	}
	d, e := json.Marshal(raw)
	if e == nil {
		data = string(d[:])
	}
	err = e
	return
}

// 破产统计
type BankruptcyEvent struct {
	SnId              int32  //用户id
	Channel           string //渠道
	Promoter          string //推广
	Platform          string //平台
	City              string //城市
	Value             int32  //值
	TelephonePromoter int32  //电销标签
	IsNew             int32  //是否新人
	Time              int64  //操作时间
	GameId            int32  //游戏id
	GameMode          int32  //游戏模式id
	GameFreeId        int32  //游戏场次id
}

func MarshalBankruptcyEvent(source, snid, telephonePromoter int32, channel, promoter, platform, city string, createDayTime time.Time, gameId, gameMode, gameFreeId int32) (data string, err error) {
	isNewbie := int32(0)
	if now.BeginningOfDay().Equal(now.New(createDayTime).BeginningOfDay()) {
		isNewbie = 1
	}
	raw := &RabbitMQDataRaw{
		Source: source,
		Data: &BankruptcyEvent{
			SnId:              snid,
			Channel:           channel,
			Promoter:          promoter,
			TelephonePromoter: telephonePromoter,
			Platform:          platform,
			City:              city,
			IsNew:             isNewbie,
			Value:             0,
			Time:              time.Now().Local().Unix(),
			GameId:            gameId,
			GameMode:          gameMode,
			GameFreeId:        gameFreeId,
		},
	}
	d, e := json.Marshal(raw)
	if e == nil {
		data = string(d[:])
	}
	err = e
	return
}

// 充值统计
type PlayerPayEvent struct {
	SnId              int32  //用户id
	Channel           string //渠道
	Promoter          string //推广
	Platform          string //平台
	City              string //城市
	TelephonePromoter int32  //电销标记
	Tag               int32  //#充值类型 0 API直接充值 1在线充值
	BeforeCoin        int32  //充值前钱包数量
	BeforeBank        int32  //充值前保险柜数量
	Amount            int32  //充值金额
	IsNew             int32  //是否是新人
	Time              int64  //操作时间
}

func MarshalPlayerPayEvent(source, snid, tag, beforeCoin, beforeBank, amount int32, channel,
	promoter, platform, city string, createDayTime time.Time, orderCreateTime int64,
	telephonePromoter int32) (data string, err error) {
	isNewbie := int32(0)
	if now.BeginningOfDay().Equal(now.New(createDayTime).BeginningOfDay()) {
		isNewbie = 1
	}
	raw := &RabbitMQDataRaw{
		Source: source,
		Data: &PlayerPayEvent{
			SnId:              snid,
			Channel:           channel,
			Promoter:          promoter,
			Platform:          platform,
			City:              city,
			Tag:               tag,
			TelephonePromoter: telephonePromoter,
			BeforeCoin:        beforeCoin,
			BeforeBank:        beforeBank,
			Amount:            amount,
			IsNew:             isNewbie,
			Time:              time.Unix(orderCreateTime, 0).Local().Unix(),
		},
	}
	d, e := json.Marshal(raw)
	if e == nil {
		data = string(d[:])
	}
	err = e
	return
}

// 系统赠送
type SystemGiveEvent struct {
	SnId              int32  //用户id
	Channel           string //渠道
	Promoter          string //推广
	Platform          string //平台
	City              string //城市
	TelephonePromoter int32  //电销
	Tag               int32  //#充值类型 0 API直接充值 1在线充值
	Amount            int32  //充值金额
	Time              int64  //操作时间
}

func MarshalSystemGiveEvent(source, snid, tag, amount int32, channel, promoter, platform, city string,
	telephonePromoter int32) (data string, err error) {
	raw := &RabbitMQDataRaw{
		Source: source,
		Data: &SystemGiveEvent{
			SnId:              snid,
			Channel:           channel,
			Promoter:          promoter,
			Platform:          platform,
			TelephonePromoter: telephonePromoter,
			City:              city,
			Tag:               tag,
			Amount:            amount,
			Time:              time.Now().Local().Unix(),
		},
	}
	d, e := json.Marshal(raw)
	if e == nil {
		data = string(d[:])
	}
	err = e
	return
}

// 水池变化记录
type GameCoinPoolEvent struct {
	Platform   string //平台
	GameId     int32  //游戏id
	GroupId    int32  //组id
	ChangeCoin int64  //变化金币
	CurCoin    int64  //变化后金币
	UpCoin     int64  //上限
	DownCoin   int64  //下限
	Time       int64  //操作时间
}

func MarshalGameCoinPoolEvent(source int32, platform string, gameid, groupId int32, changeCoin,
	curCoin, upCoin, downCoin int64) (data string, err error) {

	raw := &RabbitMQDataRaw{
		Source: source,
		Data: &GameCoinPoolEvent{
			Platform: platform,
			GameId:   gameid,

			GroupId:    groupId,
			ChangeCoin: changeCoin,
			CurCoin:    curCoin,
			UpCoin:     upCoin,
			DownCoin:   downCoin,
			Time:       time.Now().Local().Unix(),
		},
	}
	d, e := json.Marshal(raw)
	if e == nil {
		data = string(d[:])
	}
	err = e
	return
}
