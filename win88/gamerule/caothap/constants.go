package caothap

// CaoThap
const (
	CARDSTYPE_NUM       = 13 // 牌型数
	CARDDATANUM         = 52 // 牌库数
	JACKPOTTIMEINTERVAL = 24 // 单位小时

	STEP_FIRSTINIT     = 0 // 初始化校验
	STEP_BALANCE       = 1 // 结算校验
	STEP_JACKPOT_LIMIT = 7 // 爆奖校验

	TURNID_FIRSTINIT = 10000 // 初始化 时用TurnID
	TIMEINTERVAL     = 120   // 超时时间
)

const (
	CAOTHAP_JACKPOT_InitJackpot        int = iota //初始化奖池倍率
	CAOTHAP_JACKPOT_LIMITWIN_PRIZELOW             //现金池不足时 最多赢分
	CAOTHAP_JACKPOT_LIMITWIN_PRIZEHIGH            //现金池充足时 最多赢分
)

// CardID 下标索引 CardName 下标索引对应值
var cardName = [CARDDATANUM]string{
	"2♠", "3♠", "4♠", "5♠", "6♠", "7♠", "8♠", "9♠", "10♠", "J♠", "Q♠", "K♠", "A♠", // 0-12 黑桃
	"2♣", "3♣", "4♣", "5♣", "6♣", "7♣", "8♣", "9♣", "10♣", "J♣", "Q♣", "K♣", "A♣", // 13-25 梅花
	"2♦", "3♦", "4♦", "5♦", "6♦", "7♦", "8♦", "9♦", "10♦", "J♦", "Q♦", "K♦", "A♦", // 26-38 方片
	"2♥", "3♥", "4♥", "5♥", "6♥", "7♥", "8♥", "9♥", "10♥", "J♥", "Q♥", "K♥", "A♥", // 39-51 红桃
}

// 牌库数据
var cardData = [CARDDATANUM]int32{
	0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12,
	13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25,
	26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38,
	39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51,
}

var DebugCardData = [CARDDATANUM]int32{
	12, 25, 38, 51, 4, 5, 6, 7, 8, 9, 10, 11, 0,
	13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 1,
	26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 2,
	39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 3,
}
