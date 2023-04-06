package iceage

import (
	"encoding/json"
	"games.yol.com/win88/model"
	"time"
)

//玩家游戏数据索引
const (
	IAFreeTimes int = iota //0 当前剩余免费次数
	IAIndexMax
)

// 通知奖池变化的时间间隔
const jackpotNoticeInterval = time.Second

//场景状态
const (
	IceAgeSceneStateStart int = iota //开始游戏
	IceAgeSceneStateMax
)

//玩家操作
const (
	IceAgePlayerOpStart  int = iota //游戏
	IceAgePlayerHistory             // 游戏记录
	IceAgeBigWinHistory             // 大奖记录 (已废弃)
	IceAgeBonusGame                 // 小游戏结束
	IceAgeBonusGameStart            // 小游戏开始
)

//小游戏超时时间
const IceAgeBonusGameTimeout = time.Second * 15

type GameResultLog struct {
	BaseResult *model.SlotBaseResultType
	AllLine    int32     //线路数
	UserName   string    //昵称
	BetLines   []int64   //下注的线
	Cards      [][]int32 // 消除前后的牌（消除前15张，消除后15张...）
	WinLines   [][]int   // 赢分的线
}

// 冰河世纪解析的数据
type IceAgeGameNoteData struct {
	Source int32
	Data   *GameResultLog
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

//游戏测试数据 小游戏
var DebugBonusData = []int{
	7, 5, 7, 5, 4,
	3, 4, 4, 7, 6,
	1, 1, 1, 4, 6,
}

var DebugData = [][]int{
	{1, 1, 1, 4, 5,
		2, 3, 5, 4, 6,
		1, 1, 1, 4, 3},
	{2, 2, 2, 2, 2,
		2, 2, 2, 2, 2,
		2, 2, 2, 2, 2},
	{3, 3, 3, 3, 3,
		3, 3, 3, 3, 3,
		3, 3, 3, 3, 3},
	{4, 4, 4, 4, 4,
		4, 4, 4, 4, 4,
		4, 4, 4, 4, 4},
	{5, 5, 5, 5, 5,
		5, 5, 5, 5, 5,
		5, 5, 5, 5, 5},
	{6, 6, 6, 6, 6,
		6, 6, 6, 6, 6,
		6, 6, 6, 6, 6},
	{7, 7, 7, 7, 7,
		7, 7, 7, 7, 7,
		7, 7, 7, 7, 7},
}
