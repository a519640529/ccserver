package fruits

// 房间类型
const (
	RoomMode_Classic int = iota //经典
	RoomMode_Max
)

// 场景状态
const (
	FruitsStateStart int = iota //默认状态
	FruitsStateMax
)

// 玩家操作
const (
	FruitsPlayerOpStart int = iota
	FruitsPlayerOpSwitch
)
const NowByte int64 = 10000
const (
	Normal   = iota //正常
	FreeGame        //免费游戏
	MaryGame        //玛丽游戏
)
const (
	Column = 3
	Row    = 5
)
const (
	Wild       int32 = iota //0
	Bonus                   //1
	Scatter                 //2
	Bar                     //3
	Cherry                  //4.樱桃
	Bell                    //5.铃铛
	Pineapple               //6.菠萝
	Grape                   //7.葡萄
	Lemon                   //8.柠檬
	Watermelon              //9.西瓜
	Banana                  //10.香蕉
	Apple                   //11.苹果
	Bomb                    //12.炸弹
	EleMax
)

//0|0|0|0|1|0|1|1|1|1|1|1|1

//炸弹 香蕉 樱桃 西瓜 葡萄 菠萝 柠檬 苹果

//炸弹 香蕉 樱桃 西瓜 葡萄 菠萝
//炸弹 西瓜 香蕉 苹果 樱桃 柠檬
//炸弹 香蕉 菠萝 苹果 柠檬 葡萄
//炸弹 西瓜 菠萝 樱桃 柠檬 葡萄

var MaryEleArray = []int32{
	Bomb, Banana, Cherry, Watermelon, Grape, Pineapple,
	Bomb, Watermelon, Banana, Apple, Cherry, Lemon,
	Bomb, Banana, Pineapple, Apple, Lemon, Grape,
	Bomb, Watermelon, Pineapple, Cherry, Lemon, Grape,
}

// 玛丽外围元素概率
var MaryEleRate = []int32{
	550, 500, 400, 100, 200, 700,
	550, 100, 500, 600, 400, 300,
	550, 500, 700, 600, 300, 200,
	550, 100, 700, 400, 300, 200,
}

// 玛丽中间元素概率
var MaryMidEleRate = []int32{0, 0, 0, 0, 400, 0, 700, 200, 300, 100, 500, 600, 0}

// 元素对应数量的倍率
var EleNumRate = make(map[int32]map[int]int64)

// 9条线
var LineWinNum = [][]int{
	{0, 1, 2, 3, 4},
	{5, 6, 7, 8, 9},
	{10, 11, 12, 13, 14},
	{0, 6, 12, 8, 4},
	{10, 6, 2, 8, 14},
	{0, 1, 7, 13, 14},
	{10, 11, 7, 3, 4},
	{5, 1, 7, 13, 9},
	{5, 11, 7, 3, 9},
}

type WinLine struct {
	Lines  []int32
	Poss   []int32
	LineId int
	Rate   int64
}
type WinResult struct {
	EleValue   []int32
	WinLine    []WinLine //赢的线数
	JackPotNum int       //JackPot数量  按数量 翻倍 给奖池
	//玛丽游戏
	MaryOutSide  int32   //外围索引
	MaryMidArray []int32 //中间数组
	MaryOutRate  int64
	MaryMidRate  int64
	MaryLianXu   int32 //中间n连续
}

func init() {
	EleNumRate[Bonus] = make(map[int]int64)
	EleNumRate[Bonus][3] = 25
	EleNumRate[Bonus][4] = 50
	EleNumRate[Bonus][5] = 400

	EleNumRate[Scatter] = make(map[int]int64)
	EleNumRate[Scatter][3] = 100
	EleNumRate[Scatter][4] = 200
	EleNumRate[Scatter][5] = 1750

	EleNumRate[Bar] = make(map[int]int64)
	EleNumRate[Bar][3] = 75
	EleNumRate[Bar][4] = 175
	EleNumRate[Bar][5] = 1250

	EleNumRate[Cherry] = make(map[int]int64)
	EleNumRate[Cherry][3] = 45
	EleNumRate[Cherry][4] = 100
	EleNumRate[Cherry][5] = 800

	EleNumRate[Bell] = make(map[int]int64)
	EleNumRate[Bell][3] = 35
	EleNumRate[Bell][4] = 80
	EleNumRate[Bell][5] = 650

	EleNumRate[Pineapple] = make(map[int]int64)
	EleNumRate[Pineapple][3] = 30
	EleNumRate[Pineapple][4] = 70
	EleNumRate[Pineapple][5] = 550

	EleNumRate[Grape] = make(map[int]int64)
	EleNumRate[Grape][3] = 25
	EleNumRate[Grape][4] = 50
	EleNumRate[Grape][5] = 400

	EleNumRate[Lemon] = make(map[int]int64)
	EleNumRate[Lemon][3] = 15
	EleNumRate[Lemon][4] = 40
	EleNumRate[Lemon][5] = 250

	EleNumRate[Watermelon] = make(map[int]int64)
	EleNumRate[Watermelon][3] = 3
	EleNumRate[Watermelon][4] = 10
	EleNumRate[Watermelon][5] = 85

	EleNumRate[Banana] = make(map[int]int64)
	EleNumRate[Banana][2] = 1
	EleNumRate[Banana][3] = 3
	EleNumRate[Banana][4] = 10
	EleNumRate[Banana][5] = 75
}
