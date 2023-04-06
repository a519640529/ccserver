package model

import (
	"encoding/json"
)

const (
	DATASOURCE_NIL     = iota
	DATASOURCE_HUNDRED //1:百人场 百人牛牛、红黑、龙虎 奔驰宝马 森林舞会 红包 德州牛仔 鱼虾蟹
	DATASOURCE_FIGHT   //2.对战场 经典牛牛、抢庄牛牛、推饼、赢三张、德州、十三水 斗地主、跑得快 二人麻将 十点半
	DATASOURCE_ROLL    //3.拉霸 水浒装 水果机 足球英豪 女赌神 世界杯 绝地求生 皇家老虎机 财神到 冰河世纪 财神 百战成神 复仇者联盟 复活岛
	DATASOURCE_FISH    //4.捕鱼
	DATASOURCE_MINI    //5.小游戏 candy、caothap、minipoker、luckydice
)

// 百人场数据类型 ；百人牛牛、红黑、龙虎 奔驰宝马 森林舞会
func MarshalGameNoteByHUNDRED(hundRed interface{}) (data string, err error) {
	raw := &RabbitMQDataRaw{
		Source: DATASOURCE_HUNDRED,
		Data:   hundRed,
	}
	d, e := json.Marshal(raw)
	if e == nil {
		data = string(d[:])
	}
	err = e
	return
}

func UnMarshalGameNoteByHUNDRED(data string) (roll interface{}, err error) {
	gnd := &RabbitMQDataRaw{}
	if err := json.Unmarshal([]byte(data), gnd); err != nil {
		return nil, err
	}
	roll = gnd.Data
	return
}

// 对战场数据类型； 经典牛牛、抢庄牛牛、推饼、赢三张、德州、十三水 二人麻将、梭哈
func MarshalGameNoteByFIGHT(fight interface{}) (data string, err error) {
	raw := &RabbitMQDataRaw{
		Source: DATASOURCE_FIGHT,
		Data:   fight,
	}
	d, e := json.Marshal(raw)
	if e == nil {
		data = string(d[:])
	}
	err = e
	return
}

// 拉霸
func MarshalGameNoteByROLL(roll interface{}) (data string, err error) {
	raw := &RabbitMQDataRaw{
		Source: DATASOURCE_ROLL,
		Data:   roll,
	}
	d, e := json.Marshal(raw)
	if e == nil {
		data = string(d[:])
	}
	err = e
	return
}

// 小游戏
func MarshalGameNoteByMini(mini interface{}) (data string, err error) {
	raw := &RabbitMQDataRaw{
		Source: DATASOURCE_MINI,
		Data:   mini,
	}
	d, e := json.Marshal(raw)
	if e == nil {
		data = string(d[:])
	}
	err = e
	return
}

// 冰河世纪游戏记录
func UnMarshalIceAgeGameNote(data string) (roll interface{}, err error) {
	gnd := &IceAgeGameNoteData{}
	if err := json.Unmarshal([]byte(data), gnd); err != nil {
		return nil, err
	}
	roll = gnd.Data
	return
}

// 复仇者联盟游戏记录
func UnMarshalAvengersGameNote(data string) (roll interface{}, err error) {
	gnd := &AvengersGameNoteData{}
	if err := json.Unmarshal([]byte(data), gnd); err != nil {
		return nil, err
	}
	roll = gnd.Data
	return
}

//// 复仇者联盟游戏记录
//func UnMarshalAvengersGameNote(data string) (roll interface{}, err error) {
//	gnd := &AvengersGameNoteData{}
//	if err := json.Unmarshal([]byte(data), gnd); err != nil {
//		return nil, err
//	}
//	roll = gnd.Data
//	return
//}

// 财神游戏记录
func UnMarshalCaiShenGameNote(data string) (roll interface{}, err error) {
	gnd := &CaiShenGameNoteData{}
	if err := json.Unmarshal([]byte(data), gnd); err != nil {
		return nil, err
	}
	roll = gnd.Data
	return
}

// 财神游戏记录
func UnMarshalTamQuocGameNote(data string) (roll interface{}, err error) {
	gnd := &TamQuocGameNoteData{}
	if err := json.Unmarshal([]byte(data), gnd); err != nil {
		return nil, err
	}
	roll = gnd.Data
	return
}

// 复活岛游戏记录
func UnMarshalEasterIslandGameNote(data string) (roll interface{}, err error) {
	gnd := &EasterIslandGameNoteData{}
	if err := json.Unmarshal([]byte(data), gnd); err != nil {
		return nil, err
	}
	roll = gnd.Data
	return
}

// 糖果游戏记录
func UnMarshalCandyGameNote(data string) (roll interface{}, err error) {
	gnd := &CandyGameNoteData{}
	if err := json.Unmarshal([]byte(data), gnd); err != nil {
		return nil, err
	}
	roll = gnd.Data
	return
}

// MiniPoker游戏记录
func UnMarshalMiniPokerGameNote(data string) (roll interface{}, err error) {
	gnd := &MiniPokerGameNoteData{}
	if err := json.Unmarshal([]byte(data), gnd); err != nil {
		return nil, err
	}
	roll = gnd.Data
	return
}

// CaoThap游戏记录
func UnMarshalCaoThapGameNote(data string) (roll interface{}, err error) {
	gnd := &CaoThapGameNoteData{}
	if err := json.Unmarshal([]byte(data), gnd); err != nil {
		return nil, err
	}
	roll = gnd.Data
	return
}

// 幸运骰子游戏记录
func UnMarshalLuckyDiceGameNote(data string) (roll interface{}, err error) {
	gnd := &LuckyDiceGameNoteData{}
	if err := json.Unmarshal([]byte(data), gnd); err != nil {
		return nil, err
	}
	roll = gnd.Data
	return
}

// 捕鱼
func MarshalGameNoteByFISH(fish interface{}) (data string, err error) {
	raw := &RabbitMQDataRaw{
		Source: DATASOURCE_FISH,
		Data:   fish,
	}
	d, e := json.Marshal(raw)
	if e == nil {
		data = string(d[:])
	}
	err = e
	return
}
