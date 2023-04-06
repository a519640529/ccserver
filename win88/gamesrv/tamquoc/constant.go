package tamquoc

import (
	"encoding/json"
	"games.yol.com/win88/model"
	"time"
)

//玩家游戏数据索引
const (
	TamQuocFreeTimes int = iota //0 当前剩余免费次数
	TamQuocBonusTime            //1 上一次小游戏的时间
	TamQuocIndexMax
)

// 通知奖池变化的时间间隔
const jackpotNoticeInterval = time.Second

//场景状态
const (
	TamQuocSceneStateStart int = iota //开始游戏
	TamQuocSceneStateMax
)

//玩家操作
const (
	TamQuocPlayerOpStart   int = iota //游戏
	TamQuocPlayerHistory              // 游戏记录
	TamQuocBigWinHistory              // 大奖记录 (已废弃)
	TamQuocBonusGame                  // 小游戏
	TamQuocBonusGameRecord            // 小游戏重连
)

//小游戏超时时间
const TamQuocBonusGameTimeout = time.Second * 60

// 小游戏操作时间
const TamQuocBonusGamePickTime = time.Second * 15

type GameResultLog struct {
	BaseResult *model.SlotBaseResultType
	AllLine    int32   //线路数
	UserName   string  //昵称
	BetLines   []int64 //下注的线
	//Cards      [][]int32 // 消除前后的牌（消除前15张，消除后15张...）
	WinLines []int // 赢分的线
}

// 百战成神解析的数据
type TamQuocGameNoteData struct {
	Source int32
	Data   *GameResultLog
}

// 百战成神游戏记录
func UnMarshalTamQuocGameNote(data string) (roll interface{}, err error) {
	gnd := &TamQuocGameNoteData{}
	if err := json.Unmarshal([]byte(data), gnd); err != nil {
		return nil, err
	}
	roll = gnd.Data
	return
}

var DebugData = [][]int{
	{
		1, 1, 1, 5, 5,
		1, 1, 1, 3, 4,
		6, 7, 2, 3, 5,
	},
	{
		1, 1, 1, 1, 5,
		1, 1, 1, 1, 4,
		6, 7, 2, 3, 5,
	},
	{
		1, 1, 1, 1, 1,
		1, 1, 1, 1, 1,
		6, 7, 2, 3, 5,
	},
	{
		2, 2, 2, 2, 2,
		2, 2, 2, 2, 2,
		2, 2, 2, 2, 2,
	},
	{
		3, 3, 3, 3, 3,
		3, 3, 3, 3, 3,
		3, 3, 3, 3, 3,
	},
	{
		4, 4, 4, 4, 4,
		4, 4, 4, 4, 4,
		4, 4, 4, 4, 4,
	},
	{
		5, 5, 5, 5, 5,
		5, 5, 5, 5, 5,
		5, 5, 5, 5, 5,
	},
	{
		6, 6, 6, 6, 6,
		6, 6, 6, 6, 6,
		6, 6, 6, 6, 6,
	},
	{
		7, 7, 7, 7, 7,
		7, 7, 7, 7, 7,
		7, 7, 7, 7, 7,
	},
}
