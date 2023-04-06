package hunting

import (
	"games.yol.com/win88/common"
	"math/rand"
)

// 0 没有宝石

// level1
// 1 河心石
// 2 荒玉石
// 3 紫水晶
// 4 血榴石
// 5 锡黄晶

// level2
// 6 蓝锆晶
// 7 精灵榄
// 8 魇蛟眼
// 9 红玉髓
// 10 龙眼石

// level3
// 11 海洋之眼
// 12 梦境翡翠
// 13 影歌紫玉
// 14 地狱炎石
// 15 王者琥珀

// 16 炸弹

const (
	// 关卡数量
	LevelNum = 4
	// 掉线重连保持时长，单位秒
	DropLineTime = 30
	// 最短多少秒通知一次奖池金额
	CoinPoolTime = 1
	// 没有宝石
	NoGem = 0
	// 炸弹
	Bomb = 16
	// 多长时间未操作踢出房间
	TimeoutSecond = 180
)

// 游戏状态
const (
	GamingState  = 0
	GameStateMax = 1 // 状态数量
)

func levelNum(n int) int {
	if n-1 < 0 {
		return 0
	}
	if n-1 >= LevelNum {
		return LevelNum - 1
	}
	return n - 1
}

var levelPass = []int32{15, 15, 15, 0}

// 通关需要的炸弹数量
func LevelPassBombs(level int) int32 {
	return levelPass[levelNum(level)]
}

var levelGemsNum = []int{4*4 + 4, 5*5 + 5, 6*6 + 6, 0}

// 每关的初始宝石数量
func LevelGemsNum(level int) int {
	return levelGemsNum[levelNum(level)]
}

var levelColumn = []int{4, 5, 6, 0}

// 每关的宝石列数
func LevelColumn(level int) int {
	return levelColumn[levelNum(level)]
}

var levelGemType = []int32{1, 2, 3, 4, 5}

// 每关的宝石类型
func LevelGemType(level int) []int32 {
	l := levelNum(level)
	if l == 0 {
		return levelGemType
	}
	ret := make([]int32, len(levelGemType))
	for i := 0; i < len(ret); i++ {
		ret[i] = levelGemType[i] + int32(len(ret)*l)
	}
	return ret
}

var GemTypeName = map[int32]string{
	1:  "1河心石",
	2:  "2荒玉石",
	3:  "3紫水晶",
	4:  "4血榴石",
	5:  "5锡黄晶",
	6:  "1蓝锆晶",
	7:  "2精灵榄",
	8:  "3魇蛟眼",
	9:  "4红玉髓",
	10: "5龙眼石",
	11: "1海洋之眼",
	12: "2梦境翡翠",
	13: "3影歌紫玉",
	14: "4地狱炎石",
	15: "5王者琥珀",
	16: "炸弹",
}

var gemProbability = []int{4100, 2700, 2000, 800, 400}

func GetGemIndex(r *rand.Rand, num int) int {
	n := r.Intn(10000)
	sum := 0
	for i := 0; i < num; i++ {
		sum += gemProbability[i]
		if n < sum {
			return i
		}
	}
	return 0
}

var levelDelete = []int{4, 5, 6, 0}

// 每关宝石最少可消除的连接数量
func LevelDelete(level int) int {
	return levelDelete[levelNum(level)]
}

var rate = []int32{2, 3, 4, 10, 20, 40, 80, 160, 320, 640, 1000, 2000, 3000, 4000, 5000, 6000, 7000}
var level1GemRate = [][]int32{
	rate[:11],
	rate[1:12],
	rate[2:13],
	rate[3:14],
	rate[4:15],
}
var level2GemRate = [][]int32{
	level1GemRate[1],
	level1GemRate[2],
	level1GemRate[3],
	level1GemRate[4],
	rate[5:16],
}
var level3GemRate = [][]int32{
	level2GemRate[1],
	level2GemRate[2],
	level2GemRate[3],
	level2GemRate[4],
	rate[6:17],
}

// GemsRate 宝石倍率, 倍率比实际倍率扩大了10倍，使用后要除以10
// n 宝石连接数量
// tp 宝石类型
func GemsRate(level int, n int, tp int32) int64 {
	if tp == Bomb {
		return 0
	}
	switch levelNum(level) {
	case 0:
		if n < 4 || n > 14 {
			return 0
		}
		return int64(level1GemRate[(tp-1)%5][n-4])
	case 1:
		if n < 5 || n > 15 {
			return 0
		}
		return int64(level2GemRate[(tp-1)%5][n-5])
	case 2:
		if n < 6 || n > 16 {
			return 0
		}
		return int64(level3GemRate[(tp-1)%5][n-6])
	}
	return 0
}

func rateScope(level int, a, b int32, f func([][]int32) []int32) []int32 {
	if a > b {
		a, b = b, a
	}
	var ret []int32
	switch levelNum(level) {
	case 0:
		ret = f(level1GemRate)
	case 1:
		ret = f(level2GemRate)
	case 2:
		ret = f(level3GemRate)
	}
	tmp := map[int32]struct{}{}
	for i := 0; i < len(ret); i++ {
		tmp[ret[i]] = struct{}{}
	}
	ret = ret[:0]
	for k := range tmp {
		ret = append(ret, k)
	}
	return ret
}

// [a, b)
func RateScope(level int, a, b int32) []int32 {
	return rateScope(level, a, b, func(data [][]int32) []int32 {
		var ret []int32
		for _, v := range data {
			for i := 0; i < len(v); i++ {
				if v[i] >= b {
					break
				}
				if v[i] < a {
					continue
				}
				ret = append(ret, v[i])
			}
		}
		return ret
	})
}

// [a, b]
func RateScope1(level int, a, b int32) []int32 {
	return rateScope(level, a, b, func(data [][]int32) []int32 {
		var ret []int32
		for _, v := range data {
			for i := 0; i < len(v); i++ {
				if v[i] > b {
					break
				}
				if v[i] < a {
					continue
				}
				ret = append(ret, v[i])
			}
		}
		return ret
	})
}

func rateGems(level int, r int32, data [][]int32) map[int32]int {
	ret := map[int32]int{}
	for k, v := range data {
		var n = LevelDelete(level) - 1
		for i := 0; i < len(v); i++ {
			n++
			if v[i] < r {
				continue
			}
			if v[i] > r {
				break
			}
			if v[i] == r {
				ret[int32(k)+1+int32(5*levelNum(level))] = n
				break
			}
		}
	}
	return ret
}

// RateGems 根据需要的倍率，查询连接宝石和类型
// rate 倍率
func RateGems(level int, r int32) map[int32]int {
	switch levelNum(level) {
	case 0:
		return rateGems(level, r, level1GemRate)
	case 1:
		return rateGems(level, r, level2GemRate)
	case 2:
		return rateGems(level, r, level3GemRate)
	}
	return map[int32]int{}
}

func GetsRateMin(level int) map[int32][]int {
	ret := make(map[int32][]int)
	gt := LevelGemType(level)
	for i := 0; i < len(gt); i++ {
		num := 1
		var n = GemsRate(level, num, gt[i])
		for n < 10 {
			if n > 0 {
				ret[gt[i]] = append(ret[gt[i]], num)
			}
			num++
			n = GemsRate(level, num, gt[i])
		}
	}
	return ret
}

// CoinPoolRate 奖池倍率
// n 宝石连接数量
func CoinPoolRate(level int, n int) int64 {
	switch levelNum(level) {
	case 0:
		if n < 15 {
			return 0
		}
		if n <= 16 {
			return int64((n - 14) * 10)
		}
		return 20
	case 1:
		if n < 16 {
			return 0
		}
		if n <= 18 {
			return int64((n - 15) * 10)
		}
		return 30
	case 2:
		if n < 17 {
			return 0
		}
		if n <= 21 {
			return int64((n - 16) * 10)
		}
		return 50
	}
	return 0
}

// 每次下注，出现炸弹的概率是25%
func GetBomb(r *rand.Rand) bool {
	return r.Intn(100) < 25
}

// 翻牌概率
var cardRate = []int32{3500, 3000, 2000, 1000, 500}

var baseRate = [][]int64{
	{2, 5, 10, 15, 20, 30},
	{1, 3, 6, 10, 15, 20},
	{1, 2, 4, 6, 8, 10},
}

func CardScore(r *rand.Rand, baseScore int64, sceneType int32) int64 {
	var sum int32
	var i int
	n := r.Int31n(10000)
	for k, v := range cardRate {
		sum += v
		if n < sum {
			i = k
			break
		}
	}

	var arr []int64
	switch sceneType {
	case common.SceneType_Primary:
		arr = baseRate[0]
	case common.SceneType_Middle:
		arr = baseRate[1]
	default: // 高级场、试玩场
		arr = baseRate[2]
	}
	a := baseScore * arr[i]
	b := baseScore * arr[i+1]
	return a + r.Int63n(b-a+1)
}
