package caishen

import (
	"encoding/json"
	"games.yol.com/win88/model"
	"time"
)

//玩家游戏数据索引
const (
	CaiShenFreeTimes int = iota //0 当前剩余免费次数
	CaiShenIndexMax
)

// 通知奖池变化的时间间隔
const jackpotNoticeInterval = time.Second

//场景状态
const (
	CaiShenSceneStateStart int = iota //开始游戏
	CaiShenSceneStateMax
)

//玩家操作
const (
	CaiShenPlayerOpStart   int = iota //游戏
	CaiShenPlayerHistory              // 游戏记录
	CaiShenBigWinHistory              // 大奖记录 (已废弃)
	CaiShenBonusGame                  // 小游戏
	CaiShenBonusGameRecord            // 小游戏操作记录
)

//小游戏超时时间
const (
	CaiShenBonusGameTimeout      = time.Second * 60
	CaiShenBonusGameStageTimeout = 15 // 小游戏每阶段的超时时间 秒
)

type GameResultLog struct {
	BaseResult *model.SlotBaseResultType
	AllLine    int32   //线路数
	UserName   string  //昵称
	WinLines   []int   //赢分的线
	BetLines   []int64 //下注的线
}

// 财神解析的数据
type CaiShenGameNoteData struct {
	Source int32
	Data   *GameResultLog
}

// 财神游戏记录
func UnMarshalCaiShenGameNote(data string) (roll interface{}, err error) {
	gnd := &CaiShenGameNoteData{}
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
		6, 7, 8, 9, 10,
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
	{
		8, 8, 8, 8, 8,
		8, 8, 8, 8, 8,
		8, 8, 8, 8, 8,
	},
	{
		9, 9, 9, 9, 9,
		9, 9, 9, 9, 9,
		9, 9, 9, 9, 9,
	},
	{
		10, 10, 10, 10, 10,
		10, 10, 10, 10, 10,
		10, 10, 10, 10, 10,
	},
	{
		11, 11, 11, 11, 11,
		11, 11, 11, 11, 11,
		11, 11, 11, 11, 11,
	},
}
