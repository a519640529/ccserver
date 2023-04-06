package blackjack

import (
	"time"
)

const MaxCardNum = 5  // 最大手牌数量
const MaxPlayer = 5   // 玩家数量
const PokerNum = 8    // 8副牌
const OneCard = 100   // 牌背
const AudienceNum = 1 // 总人数

// 房间模式
const (
	RoomModeClassic = 0 // 经典模式
)

// 游戏状态
const (
	StatusWait   = 0 // 等待状态
	StatusReady  = 1 // 准备状态
	StatusBet    = 2 // 下注状态
	StatusDeal   = 3 // 发牌状态
	StatusBuy    = 4 // 买保险状态
	StatusBuyEnd = 5 // 保险结算状态
	StatusPlayer = 6 // 闲家操作
	StatusBanker = 7 // 庄家操作
	StatusEnd    = 8 // 结算状态
	StatusMax    = 9 // 游戏状态数量
)

// 玩家操作
const (
	SubBet      = 0 // 下注
	SubBuy      = 1 // 买保险 0不买 1买
	SubFenPai   = 2 // 分牌
	SubDouble   = 3 // 双倍
	SubSkip     = 4 // 停牌
	SubOuts     = 5 // 要牌
	SubLeave    = 6 // 离开
	SubSit      = 7 // 坐下
	SubSkipLeft = 40
	SubSkipBomb = 41
)

// 正在操作左右牌
const (
	OpDefault = 0 // 默认操作
	OpRight   = 1 // 操作右牌
)

// 超时时间
const (
	TimeoutReady         = time.Second * 3        // 准备时间
	TimeoutBet           = time.Second * 15       // 下注时间
	TimeoutDeal          = time.Second * 3        // 发牌时间
	TimeoutBuy           = time.Second * 10       // 买保险时间
	TimeoutBuyEnd        = time.Second * 3        // 保险结算时间(每人)
	TimeoutPlayer        = time.Second * 10       // 闲家操作时间
	TimeoutBanker        = time.Second * 5        // 庄家操作时间
	TimeoutEnd           = time.Second * 5        // 结算时间()
	TimeoutEndWin        = time.Second * 5        // 结算时间(全赢)5s
	TimeoutEndLost       = time.Second * 6        // 结算时间(全输)6s
	TimeoutEndWinAndLost = time.Second * 8        // 结算时间(有输有赢)8s
	TimeoutDelayOp       = time.Millisecond * 800 //玩家等待操作时间 0.8s
)

//结算结果
const (
	ResultDefault    = iota //默认结果
	ResultWin               //赢
	ResultLost              //输
	ResultWinAndLost        //有输有赢
)

// 赔率
const (
	//BaoRate  = 2   // 保险金赔率1:2
	A10Rate  = 1.5 // 黑杰克赔率2:3
	FiveRate = 1.5 // 五小龙赔率2:3
	//C21Rate   = 1   // 普通21赔率1:1
	OtherRate = 1 // 其他点赔率1:1
)

func CardsToInt32(cards []*Card) []int32 {
	var ret = make([]int32, 0, len(cards))
	for _, v := range cards {
		ret = append(ret, int32(v.Value()))
	}
	return ret
}

var Action = map[string]int{
	"split":  SubFenPai,
	"double": SubDouble,
	"supply": SubOuts,
	"pass":   SubSkip,
}

const (
	CharacterA = 1 // 激进性格
	CharacterB = 2 // 正常性格
	CharacterC = 3 // 保守性格
)
