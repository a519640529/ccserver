package avengers

import (
	"encoding/json"
	"games.yol.com/win88/model"
	"time"
)

//玩家游戏数据索引
const (
	AvengersFreeTimes int = iota //0 当前剩余免费次数
	AvengersIndexMax
)
const jackpotNoticeInterval = time.Second

//场景状态
const (
	AvengersSceneStateStart int = iota //开始游戏
	AvengersSceneStateMax
)

//玩家操作
const (
	AvengersPlayerOpStart   int = iota //游戏
	AvengersPlayerHistory              // 游戏记录
	AvengersBigWinHistory              // 大奖记录 (已废弃)
	AvengersBonusGame                  // 小游戏
	AvengersBonusGameRecord            // 小游戏操作记录
)

const (
	AvengersBonusGameTimeout      = time.Second * 60 // 小游戏最大超时时间
	AvengersBonusGameStageTimeout = 15               // 小游戏每阶段的超时时间 秒
)

type GameResultLog struct {
	BaseResult *model.SlotBaseResultType
	AllLine    int32   //线路数
	UserName   string  //昵称
	WinLines   []int   //赢分的线
	BetLines   []int64 //下注的线
}

// 复仇者联盟解析的数据
type AvengersGameNoteData struct {
	Source int32
	Data   *GameResultLog
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

var DebugData = [][]int{
	{
		5, 5, 8, 4, 5,
		4, 4, 1, 4, 1,
		4, 4, 2, 4, 10,
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
