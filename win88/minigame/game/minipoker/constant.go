package minipoker

import (
	"encoding/json"
	"games.yol.com/win88/model"
	"time"
)

//场景状态
const (
	MiniPokerSceneStateStart int = iota //开始游戏
	MiniPokerSceneStateMax
)

//玩家操作
const (
	MiniPokerPlayerOpStart int = iota //游戏
	MiniPokerPlayerHistory            //玩家记录信息
	MiniPokerPlayerSelBet             //玩家修改下注筹码
)

var jackpotNoticeInterval = time.Second

type GameResultLog struct {
	BaseResult *model.SlotBaseResultType
	CardsType  int    //牌型
	UserName   string //昵称
}

// minipoker解析的数据
type MiniPokerGameNoteData struct {
	Source int32
	Data   *GameResultLog
}

// Candy游戏记录
func UnMarshalMiniPokerGameNote(data string) (roll interface{}, err error) {
	gnd := &MiniPokerGameNoteData{}
	if err := json.Unmarshal([]byte(data), gnd); err != nil {
		return nil, err
	}
	roll = gnd.Data
	return
}

var DebugData = [][]int32{
	{
		9, 0, 12, 11, 10,
	},
	{
		5, 6, 7, 8, 9,
	},
	{
		8, 9, 10, 11, 12,
	},
	{
		12, 0, 1, 2, 3,
	},
	{
		4, 5, 6, 7, 8,
	},
	{
		4, 2, 12, 6, 8,
	},
	{
		12, 25, 38, 51, 24,
	},
	{
		12, 25, 38, 37, 24,
	},
	{
		5, 6, 20, 8, 9,
	},
	{
		8, 9, 23, 11, 12,
	},
	{
		25, 0, 1, 2, 3,
	},
	{
		17, 5, 6, 7, 8,
	},
	{
		2, 15, 28, 8, 11,
	},
	{
		12, 25, 6, 19, 4,
	},
	{
		9, 22, 12, 3, 5,
	},
	{
		9, 22, 12, 3, 5,
	},
	{
		11, 24, 12, 3, 5,
	},
	{
		12, 25, 3, 6, 5,
	},
}
