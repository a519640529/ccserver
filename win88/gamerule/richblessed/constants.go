package richblessed

// 房间类型
const (
	RoomMode_Classic int = iota //经典
	RoomMode_Max
)

// 场景状态
const (
	RichBlessedStateStart int = iota //默认状态
	RichBlessedStateMax
)

// 玩家操作
const (
	RichBlessedPlayerOpStart int = iota
	RichBlessedPlayerOpSwitch
	RichBlessedPlayerOpJack
)
const NowByte int64 = 10000
const (
	Normal   = iota //正常
	FreeGame        //免费游戏
	JackGame        //Jack
)
const (
	Column  = 3
	Row     = 5
	LineNum = 88
)
const (
	Scatter        int32 = iota //万能元素 福字
	Gongs                       //铜锣
	GoldenPhoenix               //金凤凰
	Sailboat                    //帆船
	GoldenTortoise              //金龟
	GoldIngot                   //金元宝
	Copper                      //金钱币
	A                           //A
	K                           //K
	Q                           //Q
	J                           //J
	Ten                         //10
	Nine                        //9
	EleMax
)

const (
	GoldBoy  int32 = iota //金色男孩
	GoldGirl              //金色女孩
	BlueBoy               //蓝色男孩
	BlueGirl              //蓝色女孩
	JackMax
)

// 元素对应数量的倍率
var EleNumRate = make(map[int32]map[int]int64)

var JkEleNumRate = [4]int64{1000000,
	75000,
	5000,
	1500,
}

var JKWeight = []int32{400, 80, 20, 10}

type WinLine struct {
	Lines  []int32
	Poss   []int32
	LineId int
	Rate   int64
}
type WinResult struct {
	EleValue []int32
	WinLine  []WinLine //赢的线数
	// JackPotNum int       //JackPot数量  按数量 翻倍 给奖池
	FreeNum int32
	AllRate int64

	IsHaveScatter bool

	//JACKPOT游戏
	JackpotEle  int32 //中奖元素
	JackpotRate int64
}

func init() {
	EleNumRate[Gongs] = make(map[int]int64)
	EleNumRate[Gongs][3] = 5 * LineNum
	EleNumRate[Gongs][4] = 10 * LineNum
	EleNumRate[Gongs][5] = 50 * LineNum

	EleNumRate[GoldenPhoenix] = make(map[int]int64)
	EleNumRate[GoldenPhoenix][3] = 100
	EleNumRate[GoldenPhoenix][4] = 200
	EleNumRate[GoldenPhoenix][5] = 1000

	EleNumRate[Sailboat] = make(map[int]int64)
	EleNumRate[Sailboat][3] = 50
	EleNumRate[Sailboat][4] = 100
	EleNumRate[Sailboat][5] = 500

	EleNumRate[GoldenTortoise] = make(map[int]int64)
	EleNumRate[GoldenTortoise][3] = 40
	EleNumRate[GoldenTortoise][4] = 80
	EleNumRate[GoldenTortoise][5] = 400

	EleNumRate[GoldIngot] = make(map[int]int64)
	EleNumRate[GoldIngot][3] = 25
	EleNumRate[GoldIngot][4] = 50
	EleNumRate[GoldIngot][5] = 250

	EleNumRate[Copper] = make(map[int]int64)
	EleNumRate[Copper][3] = 10
	EleNumRate[Copper][4] = 20
	EleNumRate[Copper][5] = 100

	EleNumRate[A] = make(map[int]int64)
	EleNumRate[A][3] = 5
	EleNumRate[A][4] = 10
	EleNumRate[A][5] = 50

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

	EleNumRate[Ten] = make(map[int]int64)
	EleNumRate[Ten][3] = 5
	EleNumRate[Ten][4] = 10
	EleNumRate[Ten][5] = 50

	EleNumRate[Nine] = make(map[int]int64)
	EleNumRate[Nine][3] = 5
	EleNumRate[Nine][4] = 10
	EleNumRate[Nine][5] = 50
}
