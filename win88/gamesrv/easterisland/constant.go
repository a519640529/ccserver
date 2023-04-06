package easterisland

import (
	"encoding/json"
	"games.yol.com/win88/model"
	"time"
)

//玩家游戏数据索引
const (
	EasterIslandFreeTimes int = iota //0 当前剩余免费次数
	EasterIslandIndexMax
)

// 通知奖池变化的时间间隔
const jackpotNoticeInterval = time.Second

//场景状态
const (
	EasterIslandSceneStateStart int = iota //开始游戏
	EasterIslandSceneStateMax
)

//玩家操作
const (
	EasterIslandPlayerOpStart   int = iota //游戏
	EasterIslandPlayerHistory              //游戏记录
	EasterIslandBigWinHistory              //大奖记录 (已废弃)
	EasterIslandBonusGame                  //小游戏
	EasterIslandBonusGameRecord            // 小游戏操作记录
)

const (
	EasterIslandBonusGameTimeout      = time.Second * 60
	EasterIslandBonusGameStageTimeout = 15 // 小游戏每阶段的超时时间 秒
)

type GameResultLog struct {
	BaseResult *model.SlotBaseResultType
	AllLine    int32   //线路数
	UserName   string  //昵称
	WinLines   []int   //赢分的线
	BetLines   []int64 //下注的线
}

// 复活岛解析的数据
type EasterIslandGameNoteData struct {
	Source int32
	Data   *GameResultLog
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

//{0, 6, 12, 8, 4},     //线条4
//{10, 6, 2, 8, 14},    //线条5
//{5, 1, 12, 3, 9},     //线条22
//{5, 11, 2, 13, 9},    //线条23
var DebugData = [][]int{
	{
		5, 8, 6, 8, 5,
		9, 5, 8, 5, 4,
		6, 7, 5, 9, 6,
	},
	{
		6, 8, 5, 8, 6,
		9, 5, 8, 5, 4,
		5, 7, 6, 9, 5,
	},
	{
		5, 8, 6, 8, 5,
		6, 7, 8, 9, 6,
		5, 6, 5, 6, 5,
	},
	{
		5, 6, 5, 6, 5,
		6, 7, 8, 9, 6,
		5, 8, 6, 8, 5,
	},
	//{
	//	3, 3, 3, 3, 3,
	//	3, 3, 3, 3, 3,
	//	3, 3, 3, 3, 3,
	//},
	//{
	//	4, 4, 4, 4, 4,
	//	4, 4, 4, 4, 4,
	//	4, 4, 4, 4, 4,
	//},
	//{
	//	5, 5, 5, 5, 5,
	//	5, 5, 5, 5, 5,
	//	5, 5, 5, 5, 5,
	//},
	//{
	//	6, 6, 6, 6, 6,
	//	6, 6, 6, 6, 6,
	//	6, 6, 6, 6, 6,
	//},
	//{
	//	7, 7, 7, 7, 7,
	//	7, 7, 7, 7, 7,
	//	7, 7, 7, 7, 7,
	//},
	//{
	//	8, 8, 8, 8, 8,
	//	8, 8, 8, 8, 8,
	//	8, 8, 8, 8, 8,
	//},
	//{
	//	9, 9, 9, 9, 9,
	//	9, 9, 9, 9, 9,
	//	9, 9, 9, 9, 9,
	//},
	//{
	//	10, 10, 10, 10, 10,
	//	10, 10, 10, 10, 10,
	//	10, 10, 10, 10, 10,
	//},
	//{
	//	11, 11, 11, 11, 11,
	//	11, 11, 11, 11, 11,
	//	11, 11, 11, 11, 11,
	//},
}
