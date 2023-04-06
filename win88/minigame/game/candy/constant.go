package candy

import (
	"encoding/json"
	"games.yol.com/win88/model"
)

//场景状态
const (
	CandySceneStateStart int = iota //开始游戏
	CandySceneStateMax
)

//玩家操作
const (
	CandyPlayerOpStart        int = iota //游戏
	CandyPlayerHistory                   //玩家记录信息
	CandyPlayerJackpotHistory            //爆奖池记录信息
	CandyPlayerSelBet                    //玩家修改下注筹码
)

type GameResultLog struct {
	BaseResult *model.SlotBaseResultType
	AllLine    int32   //线路数
	UserName   string  //昵称
	WinLines   []int   //赢分的线
	BetLines   []int64 //下注的线
}

// Candy解析的数据
type CandyGameNoteData struct {
	Source int32
	Data   *GameResultLog
}

// Candy游戏记录
func UnMarshalCandyGameNote(data string) (roll interface{}, err error) {
	gnd := &CandyGameNoteData{}
	if err := json.Unmarshal([]byte(data), gnd); err != nil {
		return nil, err
	}
	roll = gnd.Data
	return
}
