package fortunezhishen

//房间类型
const (
	RoomMode_Classic int = iota //经典
	RoomMode_Max
)

//场景状态
const (
	FortuneZhiShenStateStart int = iota //默认状态
	FortuneZhiShenStateMax
)

//玩家操作
const (
	FortuneZhiShenPlayerOpStart int = iota
	FortuneZhiShenPlayerOpSwitch
)
const (
	Normal         = iota //正常
	FreeGame              //免费游戏
	StopAndRotate         //停留并旋转
	StopAndRotate2        //停留并旋转(免费游戏触发的)
)
const (
	Column = 3
	Row    = 5
)
const (
	MakeAFortune int32 = iota //发财
	Wild                      //财神
	Firecrackers              //鞭炮
	Drum                      //鼓
	Jade                      //玉牌
	Copper                    //铜币
	A
	K
	Q
	J
	T
	Gemstone //宝石
)
const (
	BetGrandPrize = iota //巨奖
	BetBigPrize          //大奖
	BetMidPrize          //中奖
	BetSmallPrize        //小奖
	Bet1_4
	Bet5_8
	Bet9_12
	Bet13_16
	Bet17_19
)
const NowByte int64 = 10000
const (
	GrandPrize = 21 //巨奖
	BigPrize   = 22 //大奖
	MidPrize   = 23 //中奖
	SmallPrize = 24 //小奖
)

type WinLine struct {
	Lines  []int32
	Poss   []int32
	LineId int
	Rate   int64
}
type WinResult struct {
	Cards           []int32   //15 横着排列
	WinLine         []WinLine //赢的线数
	GemstoneRate    []int64   //宝石倍率
	GemstoneNum     int       //宝石数量
	MakeAFortuneNum int       //发财数量
	MidIcon         int32     //中间的大图标是什么
	LastRes         []int32   //上一局的结果
	NewAddGemstone  int       //新增宝石数量
}

var LineRateNum = [][]int{
	//发财 财神 鞭炮 鼓 玉牌 铜币 A K Q J T 宝石
	{1, 0, 5, 6, 6, 5, 10, 9, 8, 7, 6, 10}, //r1
	{1, 3, 4, 5, 4, 3, 11, 8, 6, 8, 7, 9},  //r2
	{1, 3, 4, 4, 5, 6, 12, 9, 8, 9, 8, 8},  //r3
	{1, 3, 4, 5, 4, 3, 11, 8, 6, 8, 9, 7},  //r4
	{1, 3, 4, 4, 5, 5, 10, 5, 4, 3, 3, 3},  //r5
}

//元素对应数量的倍率
var EleNumRate = make(map[int32]map[int]int64)

//小宝石概率
//累计巨奖 累计大奖 红利中奖 红利小奖 1-4 5-8 9-12 13-16 17-19
var SmallGemstoneRatePrize = []int{3, 7, 100, 390, 3500, 2700, 1800, 1000, 500}

//大宝石概率
//累计巨奖 累计大奖 红利中奖 红利小奖 1-4 5-8 9-12 13-16 17-19
var BigGemstoneRatePrize = []int{3, 7, 100, 390, 1800, 2700, 3500, 1000, 500}

//50种条线的结果
var LineWinNum = [][]int{
	{5, 6, 7, 8, 9},
	{0, 1, 2, 3, 4},
	{10, 11, 12, 13, 14},
	{0, 6, 12, 8, 4},
	{10, 6, 2, 8, 14},
	{0, 1, 7, 3, 4},
	{10, 11, 7, 13, 14},
	{5, 11, 12, 13, 9},
	{5, 1, 2, 3, 9},
	{0, 6, 7, 8, 4},
	///////
	{10, 6, 7, 8, 14},
	{0, 6, 2, 8, 4},
	{10, 6, 12, 8, 14},
	{5, 1, 7, 3, 9},
	{5, 11, 7, 13, 9},
	{5, 6, 2, 8, 9},
	{5, 6, 12, 8, 9},
	{0, 11, 2, 13, 4},
	{10, 1, 12, 3, 14},
	{5, 1, 12, 3, 9},
	///////
	{5, 11, 2, 13, 9},
	{0, 1, 12, 3, 4},
	{10, 11, 2, 13, 14},
	{0, 11, 12, 13, 4},
	{10, 1, 2, 3, 14},
	{0, 11, 7, 13, 4},
	{10, 1, 7, 3, 14},
	{5, 6, 7, 8, 14},
	{0, 1, 7, 13, 14},
	{10, 11, 7, 3, 4},
	///////
	{0, 6, 7, 8, 14},
	{10, 6, 7, 8, 4},
	{0, 6, 12, 8, 14},
	{10, 6, 2, 8, 4},
	{0, 1, 2, 3, 9},
	{10, 11, 12, 13, 9},
	{0, 6, 2, 8, 14},
	{10, 6, 12, 8, 4},
	{5, 1, 7, 13, 9},
	{5, 11, 7, 3, 9},
	///////
	{5, 6, 2, 3, 4},
	{5, 6, 12, 13, 14},
	{5, 1, 2, 8, 14},
	{5, 11, 12, 8, 4},
	{5, 1, 7, 13, 14},
	{5, 11, 7, 3, 4},
	{10, 6, 2, 3, 9},
	{0, 6, 12, 13, 9},
	{0, 1, 7, 13, 9},
	{10, 11, 7, 3, 9},
}

func init() {
	EleNumRate[MakeAFortune] = make(map[int]int64)
	EleNumRate[MakeAFortune][3] = 100
	EleNumRate[MakeAFortune][4] = 750
	EleNumRate[MakeAFortune][5] = 5000

	EleNumRate[Firecrackers] = make(map[int]int64)
	EleNumRate[Firecrackers][2] = 2
	EleNumRate[Firecrackers][3] = 10
	EleNumRate[Firecrackers][4] = 50
	EleNumRate[Firecrackers][5] = 150

	EleNumRate[Drum] = make(map[int]int64)
	EleNumRate[Drum][3] = 5
	EleNumRate[Drum][4] = 30
	EleNumRate[Drum][5] = 100

	EleNumRate[Jade] = make(map[int]int64)
	EleNumRate[Jade][3] = 5
	EleNumRate[Jade][4] = 25
	EleNumRate[Jade][5] = 100

	EleNumRate[Copper] = make(map[int]int64)
	EleNumRate[Copper][3] = 5
	EleNumRate[Copper][4] = 25
	EleNumRate[Copper][5] = 100

	EleNumRate[A] = make(map[int]int64)
	EleNumRate[A][3] = 5
	EleNumRate[A][4] = 20
	EleNumRate[A][5] = 75

	EleNumRate[K] = make(map[int]int64)
	EleNumRate[K][3] = 5
	EleNumRate[K][4] = 10
	EleNumRate[K][5] = 50

	EleNumRate[Q] = make(map[int]int64)
	EleNumRate[Q][3] = 5
	EleNumRate[Q][4] = 10
	EleNumRate[Q][5] = 50

	EleNumRate[J] = make(map[int]int64)
	EleNumRate[J][3] = 5
	EleNumRate[J][4] = 10
	EleNumRate[J][5] = 50

	EleNumRate[T] = make(map[int]int64)
	EleNumRate[T][3] = 5
	EleNumRate[T][4] = 10
	EleNumRate[T][5] = 50
}
