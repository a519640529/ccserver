package luckydice

//场景状态
const (
	LuckyDiceSceneStateBetting         int = iota // 投注
	LuckyDiceSceneStateShowResult                 // 展示结果
	LuckyDiceSceneStatePrepareNewRound            // 准备新牌局
	LuckyDiceSceneStateEndBetting                 // 投注结束
	LuckyDiceSceneStateMax
)

//场景超时 单位秒
//var LuckyDiceSceneStateTimeout = []int32{50, 20, 3, 1}
var LuckyDiceSceneStateTimeout = []int32{10, 10, 3, 1}

//玩家操作
const (
	LuckyDicePlayerOpBet   int = iota //投注
	LuckyDicePlayerHistory            //玩家历史投注
	LuckyDiceRoundBetData             //单局投注记录
	LuckyDiceDiceHistory              //近100局投注骰子记录
)
