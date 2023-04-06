package hunting

import (
	"bytes"
	"container/list"
	"fmt"
	"math"
	"math/rand"
	"sort"
)

// GetXY 获取宝石位置
// n 一维数组中宝石索引
// 返回对应的二维数组的索引
func GetXY(level int, n int) (x, y int) {
	return getXY(LevelColumn(level), n)
}

func getXY(col int, n int) (x, y int) {
	return int(n) % col, int(n) / col
}

// XYToIndex 二维坐标转一维索引
func XYToIndex(level int, x, y int) int {
	return xyToIndex(LevelColumn(level), x, y)
}

func xyToIndex(n int, x, y int) int {
	return x + n*y
}

// NoGemsData 没有任何宝石的初始状态
func NoGemsData(level int) [][]int32 {
	gems := make([][]int32, LevelColumn(level)+1)
	for i := 0; i < len(gems); i++ {
		gems[i] = make([]int32, LevelColumn(level))
	}
	return gems
}

func GemsDataCopy(a, b [][]int32) {
	for i := 0; i < len(a); i++ {
		copy(a[i], b[i])
	}
}

func createGems(r *rand.Rand, level int, f func(r *rand.Rand, num int) int) (data [][]int32, bombs int32) {
	if GetBomb(r) {
		// 有炸弹时减少宝石连接数量
		data = NewGemsLessBomb(r, level)
		bombs++
		return
	}
	var ret []int32
	gt := LevelGemType(level)
	n := LevelGemsNum(level) // 宝石数量
	for i := 0; i < n; i++ {
		ret = append(ret, gt[f(r, len(gt))])
	}
	r.Shuffle(len(ret), func(i, j int) {
		ret[i], ret[j] = ret[j], ret[i]
	})
	n = LevelColumn(level)
	for i := 0; i*n < len(ret); i++ {
		data = append(data, ret[i*n:i*n+n])
	}
	return
}

// NewGems 随机宝石
// 返回宝石和炸弹数
func NewGems(r *rand.Rand, level int) (data [][]int32, bombs int32) {
	return createGems(r, level, GetGemIndex)
}

// NewGemsNoLink 创建没有可消除的宝石
func NewGemsNoLink(r *rand.Rand, level int) (data [][]int32) {
	n := LevelGemsNum(level)
	data = NoGemsData(level)
	index := r.Perm(n)
	AddGemsLess2(level, data, index)
	return
}

func NewGemsBombsEnd(r *rand.Rand, level int) (data [][]int32) {
	data = NewGemsNoLink(r, level)
	x, y := GetXY(level, r.Intn(LevelGemsNum(level)-LevelColumn(level)))
	for j := len(data) - 1; j > y; j-- {
		data[j][x] = data[j-1][x]
	}
	data[y][x] = Bomb
	return
}

func GemsC(r *rand.Rand, level int) [][]int32 {
	data := NoGemsData(level)
	colNum := len(data[0])
	rowNum := len(data)

	var gt int32
	var n int
	m := GetsRateMin(level)
	for k, v := range m {
		gt = k
		n = v[r.Intn(len(v))]
		break
	}
	if n == 0 {
		return data
	}

	indexes := map[int]struct{}{
		r.Intn(LevelGemsNum(level)): {},
	}
	var x, y int
	for ; n > 0 && len(indexes) > 0; n-- {
		for k := range indexes {
			x, y = getXY(colNum, k)
			delete(indexes, k)
			break
		}
		data[y][x] = gt
		if y+1 < rowNum && data[y+1][x] == NoGem {
			indexes[xyToIndex(colNum, x, y+1)] = struct{}{}
		}
		if y-1 >= 0 && data[y-1][x] == NoGem {
			indexes[xyToIndex(colNum, x, y-1)] = struct{}{}
		}
		if x-1 >= 0 && data[y][x-1] == NoGem {
			indexes[xyToIndex(colNum, x-1, y)] = struct{}{}
		}
		if x+1 < colNum && data[y][x+1] == NoGem {
			indexes[xyToIndex(colNum, x+1, y)] = struct{}{}
		}
	}
	return data
}

// NewGemsLess 创建所有宝石使消除宝石倍率小于1，并且只能消除一次
func NewGemsLess(r *rand.Rand, level int) (data [][]int32) {
	if r.Intn(100) < 40 {
		// 不能消除
		return NewGemsNoLink(r, level)
	}
	var ret = NoGemsData(level)
	for {
		// 满足条件的情况很多，这里应该不会循环很多次，通常一次就可以搜索到
		data = GemsC(r, level)
		AddGemsLess2(level, data, nil)

		GemsDataCopy(ret, data)

		// 验证
		del := Check(level, data)
		if len(del) == 0 {
			return ret
		}
		if len(del) > 1 {
			continue
		}
		for _, v := range del[0] {
			x, y := GetXY(level, v)
			data[y][x] = NoGem
		}
		AddGemsLess2(level, data, Drop(data))
		if len(Check(level, data)) == 0 {
			break
		}
	}
	return ret
}

// NewGemsLessBomb 创建包含炸弹的所有宝石，并且炸弹消除后最多只能再消除一次不
func NewGemsLessBomb(r *rand.Rand, level int) (data [][]int32) {
	var ret = NoGemsData(level)
	for {
		data = NewGemsLess(r, level)
		// 炸弹不放最上方一排
		x, y := GetXY(level, r.Intn(LevelGemsNum(level)-LevelColumn(level)))
		for j := len(data) - 1; j > y; j-- {
			data[j][x] = data[j-1][x]
		}
		data[y][x] = Bomb

		GemsDataCopy(ret, data)

		// 验证，炸弹消除后最多再消除一次
		data[y][x] = NoGem
		AddGemsLess2(level, data, Drop(data))
		del := Check(level, data)
		if len(del) == 0 {
			break
		}
		if len(del) == 1 {
			for _, v := range del[0] {
				x, y := GetXY(level, v)
				data[y][x] = NoGem
			}
			AddGemsLess2(level, data, Drop(data))
			if len(Check(level, data)) == 0 {
				break
			}
		}
	}
	return ret
}

// Check 查询消除位置
// 有炸弹时只返回消除炸弹；没有炸弹的时候有多组需要消除就都查询出来
// 返回值，外层数组是有多少组可以消除，内层是可消除宝石所在的一维数组的索引
func Check(level int, data [][]int32) [][]int {
	return check(level, data, LevelColumn(level), len(data)-1)
}

func check(level int, data [][]int32, colNum, rowNum int) [][]int {
	var ret [][]int
	if colNum <= 0 || rowNum <= 0 {
		return ret
	}

	// tmp 用来记录已连接位置
	var tmp = make([][]bool, rowNum)
	for i := 0; i < len(tmp); i++ {
		tmp[i] = make([]bool, colNum)
	}

	var place [][2]int

	for y := 0; y < rowNum; y++ {
		for x := 0; x < colNum; x++ {
			if data[y][x] == NoGem {
				tmp[y][x] = true
				continue
			}
			if data[y][x] == Bomb {
				ret = [][]int{{xyToIndex(colNum, x, y)}}
				return ret
			}
			if tmp[y][x] {
				continue
			}
			tmp[y][x] = true
			// 搜索连接(广度优先搜索)
			place = place[:0]
			place = append(place, [2]int{x, y})
			var index []int
			for len(place) > 0 {
				e := place[0]
				place = place[1:]
				index = append(index, xyToIndex(colNum, e[0], e[1]))
				//fmt.Printf("(%d,%d) ",e[0],e[1])
				// 下
				if Y := e[1] - 1; Y >= 0 && data[Y][e[0]] == data[e[1]][e[0]] && !tmp[Y][e[0]] {
					place = append(place, [2]int{e[0], Y})
					tmp[Y][e[0]] = true
				}
				// 右
				if X := e[0] + 1; X < colNum && data[e[1]][X] == data[e[1]][e[0]] && !tmp[e[1]][X] {
					place = append(place, [2]int{X, e[1]})
					tmp[e[1]][X] = true
				}
				// 左
				if X := e[0] - 1; X >= 0 && data[e[1]][X] == data[e[1]][e[0]] && !tmp[e[1]][X] {
					place = append(place, [2]int{X, e[1]})
					tmp[e[1]][X] = true
				}
				// 上
				if Y := e[1] + 1; Y < rowNum && data[Y][e[0]] == data[e[1]][e[0]] && !tmp[Y][e[0]] {
					place = append(place, [2]int{e[0], Y})
					tmp[Y][e[0]] = true
				}
			}
			//fmt.Println()
			ret = append(ret, index)
		}
	}
	for i := 0; i < len(ret); {
		if len(ret[i]) < LevelDelete(level) {
			ret = append(ret[:i], ret[i+1:]...)
		} else {
			// 排序
			sort.Slice(ret[i], func(a, b int) bool {
				return ret[i][a] < ret[i][b]
			})
			i++
		}
	}
	return ret
}

func checkXY(level int, data [][]int32, colNum, rowNum, x, y int) []int {
	var ret []int
	if colNum <= 0 || rowNum <= 0 {
		return ret
	}

	// tmp 用来记录已查看位置
	var tmp = make([][]bool, rowNum)
	for i := 0; i < len(tmp); i++ {
		tmp[i] = make([]bool, colNum)
	}

	gt := data[y][x]
	tmp[y][x] = true

	queue := list.New()
	queue.PushBack(XYToIndex(level, x, y))
	for queue.Len() > 0 {
		i := queue.Remove(queue.Front()).(int)
		ret = append(ret, i)
		x, y := GetXY(level, i)

		// 下
		if y-1 >= 0 && !tmp[y-1][x] {
			tmp[y-1][x] = true
			if data[y-1][x] == gt {
				queue.PushBack(XYToIndex(level, x, y-1))
			}
		}

		// 右
		if x+1 < colNum && !tmp[y][x+1] {
			tmp[y][x+1] = true
			if data[y][x+1] == gt {
				queue.PushBack(XYToIndex(level, x+1, y))
			}
		}

		// 左
		if x-1 >= 0 && !tmp[y][x-1] {
			tmp[y][x-1] = true
			if data[y][x-1] == gt {
				queue.PushBack(XYToIndex(level, x-1, y))
			}
		}

		// 上
		if y+1 < rowNum && !tmp[y+1][x] {
			tmp[y+1][x] = true
			if data[y+1][x] == gt {
				queue.PushBack(XYToIndex(level, x, y+1))
			}
		}
	}
	return ret
}

// Drop 宝石下落
// 消除宝石后将上方的宝石向下移动
// 返回需要补充宝石的位置（一维数组的索引）
func Drop(data [][]int32) []int {
	var ret []int
	if len(data) == 0 {
		return ret
	}
	colNum := len(data[0])
	rowNum := len(data)
	a, b := 0, 0
	flag := false
	for i := 0; i < colNum; i++ {
		a, b = 0, 0
		flag = false
		for ; a < rowNum && b < rowNum; a++ {
			if data[a][i] != NoGem {
				continue
			}
			if !flag {
				flag = true
				b = a + 1
			} else {
				b++
			}
			for ; b < rowNum; b++ {
				if data[b][i] != NoGem {
					data[b][i], data[a][i] = data[a][i], data[b][i]
					break
				}
			}
		}
		j := a - 1
		if data[j][i] == NoGem {
			ret = append(ret, xyToIndex(colNum, i, j))
		}
		j++
		for ; j < rowNum; j++ {
			ret = append(ret, xyToIndex(colNum, i, j))
		}
	}
	sort.Slice(ret, func(i, j int) bool {
		return ret[i] < ret[j]
	})
	return ret
}

// AddGemsLess 填充空缺的宝石, 使连接宝石最少
// 返回添加的宝石
func AddGemsLess(r *rand.Rand, level int, data [][]int32, index []int) []int32 {
	//if len(index) > LevelDelete(level) {
	return AddGemsLess2(level, data, index)
	//}
	// 遍历所有情况，当空缺宝石数量太多时搜索速度太慢,所以只有填充宝石比较少的时候才这样搜
	//var ret = make([]int32, len(index))
	//var minRate int64 = math.MaxInt64
	//findGemsLose(r, level, data, LevelGemType(level), index, 0, ret, &minRate)
	//return ret
}

func findGemsLose(r *rand.Rand, level int, data [][]int32, gt []int32, index []int, i int, gems []int32, minRate *int64) bool {
	r.Shuffle(len(gt), func(i, j int) {
		gt[i], gt[j] = gt[j], gt[i]
	})
	x, y := getXY(len(data[0]), index[i])
	for _, v := range gt {
		data[y][x] = v
		if i == len(index)-1 {
			del := check(level, data, LevelColumn(level), len(data))
			var rate int64
			for j := 0; j < len(del); j++ {
				x, y := getXY(len(data[0]), del[j][0])
				rate += GemsRate(level, len(del[j]), data[y][x])
			}
			if rate < *minRate {
				for i := 0; i < len(index); i++ {
					x, y := getXY(len(data[0]), index[i])
					gems[i] = data[y][x]
				}
				*minRate = rate
			}
			if *minRate == 0 {
				// 停止搜索
				return true
			}
		} else {
			if findGemsLose(r, level, data, gt, index, i+1, gems, minRate) {
				return true
			}
		}
	}
	// 继续搜索
	return false
}

// AddGemsLess2 填充空缺的宝石,使连接宝石最少
// 返回添加的宝石
// 按空缺位置，从前往后依次填充新宝石，填充的宝石需要满足以下几个条件
// 相邻位置：上，下，左，右，左上，右上，左下，右下 8 个位置
// 1.和相邻位置宝石不相同，如果有多个类型可选，随机选择一种，
// 2.和相邻位置宝石连接最少
func AddGemsLess2(level int, data [][]int32, index []int) []int32 {
	var ret = make([]int32, 0, len(index))
	var m map[int32]int
	var n int
	colNum := len(data[0])
	rowNum := len(data)

	if len(index) == 0 {
		for j := 0; j < rowNum; j++ {
			for i := 0; i < colNum; i++ {
				if data[j][i] == NoGem {
					index = append(index, xyToIndex(colNum, i, j))
				}
			}
		}
	}

	gt := LevelGemType(level)
	var gts = make([]int32, 0, len(gt))
	gtMap := make(map[int32]struct{})
	for i := 0; i < len(gt); i++ {
		gtMap[gt[i]] = struct{}{}
	}
	for i := 0; i < len(index); i++ {
		x, y := getXY(colNum, index[i])
		m = make(map[int32]int)
		// 相邻8个位置
		if y+1 < rowNum && data[y+1][x] != NoGem {
			m[data[y+1][x]] += 1
		}
		if y-1 >= 0 && data[y-1][x] != NoGem {
			m[data[y-1][x]] += 1
		}
		if x-1 >= 0 && data[y][x-1] != NoGem {
			m[data[y][x-1]] += 1
		}
		if x+1 < colNum && data[y][x+1] != NoGem {
			m[data[y][x+1]] += 1
		}
		if x-1 >= 0 && y+1 < rowNum && data[y+1][x-1] != NoGem {
			m[data[y+1][x-1]] += 1
		}
		if x+1 < colNum && y+1 < rowNum && data[y+1][x+1] != NoGem {
			m[data[y+1][x+1]] += 1
		}
		if x-1 >= 0 && y-1 >= 0 && data[y-1][x-1] != NoGem {
			m[data[y-1][x-1]] += 1
		}
		if x+1 < colNum && y-1 >= 0 && data[y-1][x+1] != NoGem {
			m[data[y-1][x+1]] += 1
		}

		if len(m) != len(gt) {
			for k := range gtMap {
				if _, ok := m[k]; !ok {
					ret = append(ret, k)
					data[y][x] = k
					break
				}
			}
		} else {
			// 连接数最少
			n = math.MaxInt64
			gts = gts[:0]
			for k := range m {
				data[y][x] = k
				l := len(checkXY(level, data, colNum, rowNum, x, y))
				if l < n {
					n = l
					gts = gts[:0]
					gts = append(gts, k)
				} else if l == n {
					gts = append(gts, k)
				}
			}
			data[y][x] = gts[0]
			ret = append(ret, gts[0])
		}
	}
	return ret
}

// NewGemsRate 根据需要倍率随机生成宝石,并且只能消除一次
func NewGemsRate(r *rand.Rand, level int, rate int32) [][]int32 {
	data := NoGemsData(level)
	colNum := len(data[0])
	rowNum := len(data) - 1

	var gt int32
	var n int
	for k, v := range RateGems(level, rate) {
		gt = k
		n = v
		break
	}
	if n == 0 {
		return NewGemsNoLink(r, level)
	}
	indexes := map[int]struct{}{
		r.Intn(LevelGemsNum(level) - LevelColumn(level)): {},
	}
	var x, y int
	for ; n > 0 && len(indexes) > 0; n-- {
		for k := range indexes {
			x, y = getXY(colNum, k)
			delete(indexes, k)
			break
		}
		data[y][x] = gt
		if y+1 < rowNum && data[y+1][x] == NoGem {
			indexes[xyToIndex(colNum, x, y+1)] = struct{}{}
		}
		if y-1 >= 0 && data[y-1][x] == NoGem {
			indexes[xyToIndex(colNum, x, y-1)] = struct{}{}
		}
		if x-1 >= 0 && data[y][x-1] == NoGem {
			indexes[xyToIndex(colNum, x-1, y)] = struct{}{}
		}
		if x+1 < colNum && data[y][x+1] == NoGem {
			indexes[xyToIndex(colNum, x+1, y)] = struct{}{}
		}
	}
	AddGemsLess2(level, data, nil)

	var ret = NoGemsData(level)
	GemsDataCopy(ret, data)

	del := Check(level, data)
	if len(del) != 1 {
		return NewGemsRate(r, level, rate)
	}
	for _, v := range del[0] {
		x, y := GetXY(level, v)
		data[y][x] = NoGem
	}
	AddGemsLess2(level, data, Drop(data))
	if len(Check(level, data)) == 0 {
		return ret
	}
	return NewGemsRate(r, level, rate)
}

func b(n int32) int32 {
	if n == 80 {
		n = '.'
	}
	if n == 64 {
		n = ' '
	}
	return n
}

func toStringLine(data []int32) string {
	if len(data) == 0 {
		return ""
	}
	ret := bytes.Buffer{}
	ret.WriteString(fmt.Sprintf("[%s", string(b(data[0]+64))))
	for i := 1; i < len(data); i++ {
		ret.WriteString(fmt.Sprintf(" %s", string(b(data[i]+64))))
	}
	ret.WriteString("]")
	return ret.String()
}

// ToString 将宝石都打印出来，用于bug调试
// 不同宝石用 A,B,C... 表示；用 . 表示炸弹
func ToString(data [][]int32) string {
	ret := bytes.Buffer{}
	for i := len(data) - 1; i >= 0; i-- {
		ret.WriteString(toStringLine(data[i]))
		ret.WriteString("\n")
	}
	return ret.String()
}
