package caothap

import "time"

const jackpotNoticeInterval = time.Second

//场景状态
const (
	CaoThapSceneStateStart int = iota //开始游戏
	CaoThapSceneStateMax
)

//玩家操作
const (
	CaoThapPlayerOpInit   int = iota // 初始化首轮牌
	CaoThapPlayerOpSetBet            // 玩家下注
	CaoThapPlayerOpStop              // 提前结算
	CaoThapPlayerHistory             // 玩家操作记录
	CaoThapPlayerSelBet              // 玩家修改下注筹码
)

//大小
const (
	CaoThapLitte int64 = iota //小
	CaoThapBig                //大
)

//结算状态码
const (
	Status_Stop int64 = iota //主动结束
	Status_Win               //赢
	Status_Lose              //输
)
