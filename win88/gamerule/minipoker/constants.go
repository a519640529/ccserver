package minipoker

// MiniPoker
const (
	CARDTYPE_MIN              int = iota + 2 // 2 MIN
	CARDTYPE_FOURCARD                        // 3 四张
	CARDTYPE_THREETAKEPAIR                   // 4 三带对
	CARDTYPE_FLUSH                           // 5 同花
	CARDTYPE_STRAIGHT                        // 6 顺子
	CARDTYPE_THREETAKESINGLE                 // 7 三带单
	CARDTYPE_TWOPAIR                         // 8 两对
	CARDTYPE_ONEPAIR_J                       // 9 一对 J+ (对J 对Q 对K 对A)
	CARDTYPE_ONEPAIR                         // 10 一对 （2-10 的对子）
	CARDTYPE_HIGHCARD                        // 11 散牌
	CARDTYPE_STRAIGHT_FLUSH_J                // 12 同花顺 J (7、8、9、10开头的顺子) 爆奖
	CARDTYPE_STRAIGHT_FLUSH                  // 13 同花顺 (A、2、3、4、5、6开头的顺子)
	CARDTYPE_MAX                             // 14 MAX
)

const (
	CARDSTYPE_NUM       = 13 // 牌型数
	CARDDATANUM         = 52 // 牌库数
	CARDNUM             = 5  // 牌数
	JACKPOTTIMEINTERVAL = 24 // 单位小时
)

// jack params
const (
	MINIPOKER_JACKPOT_InitJackpot int = iota //初始化奖池倍率
)

// 牌型赔率 做相应 * 10 / 10 处理
var cardsTypeRate = [CARDSTYPE_NUM]int{0, 0, 1500, 500, 200, 130, 80, 50, 25, 0, 0, 0, 10000}

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

// CardID 下标索引 CardName 下标索引对应值
var cardName2 = [CARDDATANUM]string{
	"A♦", "2♦", "3♦", "4♦", "5♦", "6♦", "7♦", "8♦", "9♦", "10♦", "J♦", "Q♦", "K♦", // 26-38 方片
	"A♣", "2♣", "3♣", "4♣", "5♣", "6♣", "7♣", "8♣", "9♣", "10♣", "J♣", "Q♣", "K♣", // 13-25 梅花
	"A♥", "2♥", "3♥", "4♥", "5♥", "6♥", "7♥", "8♥", "9♥", "10♥", "J♥", "Q♥", "K♥", // 39-51 红桃
	"A♠", "2♠", "3♠", "4♠", "5♠", "6♠", "7♠", "8♠", "9♠", "10♠", "J♠", "Q♠", "K♠", // 0-12 黑桃
}

// 牌库数据
var cardData2 = [CARDDATANUM]int32{
	0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12,
	13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25,
	26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38,
	39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51,
}
