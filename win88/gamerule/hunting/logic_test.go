package hunting

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"testing"
	"time"
)

var rd *rand.Rand

func init() {
	rd = rand.New(rand.NewSource(time.Now().UnixNano()))
}

func CreateTestData() {
	for i := 0; i < 1; i++ {
		level := rd.Intn(3) + 1
		data, b := NewGems(rd, level)
		s1, _ := json.Marshal(data)
		fmt.Printf("%d,%d\n", level, b)

		result := make([][]int32, len(data))
		for i := 0; i < len(result); i++ {
			result[i] = make([]int32, len(data[i]))
			copy(result[i], data[i])
		}

		index := Check(level, data)
		for _, v := range index {
			for j := 0; j < len(v); j++ {
				x, y := GetXY(level, v[j])
				result[y][x] = -22
			}
		}
		s2, _ := json.Marshal(result)

		for j := len(data) - 1; j >= 0; j-- {
			fmt.Printf("%v\t%v\n", toStringLine(data[j]), toStringLine(result[j]))
		}
		fmt.Println("测试数据：")
		fmt.Printf("{\"%s\",\"%s\",\"%s\"},\n", s1, s2, fmt.Sprint(level))
	}
}

func TestCheck(t *testing.T) {
	//CreateTestData()
	testData := [][]string{
		{"[[2,1,1,1],[1,2,4,3],[1,3,3,1],[3,1,2,4],[1,2,5,1]]", "[[2,1,1,1],[1,2,4,3],[1,3,3,1],[3,1,2,4],[1,2,5,1]]", "1"},
		{"[[3,2,16,1],[1,1,1,1],[1,3,1,2],[1,2,1,2],[2,1,3,2]]", "[[3,2,-22,1],[1,1,1,1],[1,3,1,2],[1,2,1,2],[2,1,3,2]]", "1"},
		{"[[6,6,9,8,8],[7,6,8,7,6],[9,7,6,7,9],[7,7,9,8,7],[8,6,6,6,8],[8,6,7,6,7]]", "[[6,6,9,8,8],[7,6,8,7,6],[9,7,6,7,9],[7,7,9,8,7],[8,6,6,6,8],[8,6,7,6,7]]", "2"},
		{"[[1,2,3,2],[3,2,1,1],[1,2,1,3],[1,2,1,1],[1,2,3,3]]", "[[1,-22,3,2],[3,-22,-22,-22],[1,-22,-22,3],[1,-22,-22,-22],[1,2,3,3]]", "1"},
		{"[[4,1,1,2],[1,1,3,3],[1,3,1,1],[1,3,4,1],[1,3,4,1]]", "[[4,-22,-22,2],[-22,-22,3,3],[-22,3,1,1],[-22,3,4,1],[1,3,4,1]]", "1"},
		{"[[4,1,1,3],[2,1,3,2],[3,3,1,3],[1,5,1,2],[4,3,2,1]]", "[[4,1,1,3],[2,1,3,2],[3,3,1,3],[1,5,1,2],[4,3,2,1]]", "1"},
		{"[[11,12,11,12,11,12],[13,13,11,11,12,11],[15,11,12,13,13,12],[12,12,12,12,13,15],[15,14,13,12,11,15],[12,13,12,11,11,11],[12,11,12,13,12,14]]", "[[11,12,11,12,11,12],[13,13,11,11,12,11],[15,11,-22,13,13,12],[-22,-22,-22,-22,13,15],[15,14,13,-22,11,15],[12,13,12,11,11,11],[12,11,12,13,12,14]]", "3"},
		{"[[11,12,13,11,11,11],[11,11,11,11,11,14],[13,15,11,11,11,11],[12,12,13,11,12,13],[11,11,13,11,12,13],[12,13,13,14,11,12],[11,11,12,15,12,11]]", "[[-22,12,13,-22,-22,-22],[-22,-22,-22,-22,-22,14],[13,15,-22,-22,-22,-22],[12,12,13,-22,12,13],[11,11,13,-22,12,13],[12,13,13,14,11,12],[11,11,12,15,12,11]]", "3"},
		{"[[11,11,11,12,15,13],[11,12,15,11,12,12],[13,12,15,13,11,11],[13,11,12,15,13,12],[13,13,12,12,11,11],[12,11,12,12,12,11],[11,13,12,12,12,11]]", "[[11,11,11,12,15,13],[11,12,15,11,12,12],[13,12,15,13,11,11],[13,11,-22,15,13,12],[13,13,-22,-22,11,11],[12,11,-22,-22,-22,11],[11,13,12,12,12,11]]", "3"},
		{"[[13,12,12,12,12,13],[11,13,12,13,13,13],[11,14,11,12,12,12],[13,11,14,12,11,12],[11,14,11,15,14,12],[11,14,11,11,12,11],[11,13,13,11,13,13]]", "[[13,12,12,12,12,13],[11,13,12,13,13,13],[11,14,11,-22,-22,-22],[13,11,14,-22,11,-22],[11,14,11,15,14,-22],[11,14,11,11,12,11],[11,13,13,11,13,13]]", "3"},
		{"[[7,8,7,9,10],[7,6,6,6,7],[8,6,8,6,7],[7,6,7,7,7],[7,6,7,6,6],[7,7,6,6,7]]", "[[7,8,7,9,10],[7,-22,-22,-22,-22],[8,-22,8,-22,-22],[7,-22,-22,-22,-22],[7,-22,-22,6,6],[7,7,6,6,7]]", "2"},
		{"[[16,6,7,7,6],[8,6,6,7,6],[8,7,9,6,8],[7,8,8,6,6],[7,6,9,6,7],[8,8,6,6,10]]", "[[-22,6,7,7,6],[8,6,6,7,6],[8,7,9,6,8],[7,8,8,6,6],[7,6,9,6,7],[8,8,6,6,10]]", "2"},
	}
	for _, v := range testData {
		var gems, result [][]int32
		json.Unmarshal([]byte(v[0]), &gems)
		level, _ := strconv.Atoi(v[2])
		result = make([][]int32, len(gems))
		for i := 0; i < len(result); i++ {
			result[i] = make([]int32, len(gems[i]))
			copy(result[i], gems[i])
		}
		index := Check(level, gems)
		for _, v := range index {
			for j := 0; j < len(v); j++ {
				x, y := GetXY(level, v[j])
				result[y][x] = -22
			}
		}
		s, _ := json.Marshal(result)
		s2 := string(s)
		if v[1] != string(s) {
			var su [][]int32
			json.Unmarshal([]byte(v[1]), &su)
			fmt.Printf("%-12s\t%-12s\t%-12s\n", "Gems:", "want:", "result:")
			for i := len(gems) - 1; i >= 0; i-- {
				fmt.Printf("%v\t%v\t%v\n", toStringLine(gems[i]), toStringLine(su[i]), toStringLine(result[i]))
			}
			fmt.Println("正确数据：")
			fmt.Printf("{\"%s\",\"%s\",\"%s\"},\n", v[0], s2, fmt.Sprint(level))
			t.Error("1")
		}
	}
}

func testDrop(data [][]int32) []int {
	var ret []int
	if len(data) == 0 {
		return ret
	}
	colNum := len(data[0])
	rowNum := len(data)

	result := make([][]int32, len(data))
	for i := 0; i < len(result); i++ {
		result[i] = make([]int32, len(data[i]))
		copy(result[i], data[i])
	}

	var y int
	for i := 0; i < colNum; i++ {
		y = 0
		for j := 0; j < rowNum; j++ {
			if data[j][i] != NoGem {
				result[y][i] = data[j][i]
				y++
			}
		}
	}
	for i := 0; i < colNum; i++ {
		for j := 0; j < rowNum; j++ {
			if data[j][i] == NoGem {
				ret = append(ret, int(xyToIndex(colNum, i, j)))
			}
		}
	}
	sort.Slice(ret, func(i, j int) bool {
		return ret[i] < ret[j]
	})
	return ret
}

func TestDrop(t *testing.T) {
	for i := 0; i < 1000; i++ {
		// 随机创建宝石
		level := rd.Intn(3) + 1
		data, _ := NewGems(rd, level)
		//fmt.Printf("%d,%d\n", level, b)
		//fmt.Println(toString(data))
		// 随机删除宝石
		n := LevelGemsNum(level)
		delNum := rd.Intn(n)
		for i := 0; i < delNum; i++ {
			x, y := GetXY(level, rd.Intn(n))
			data[y][x] = NoGem
		}
		//fmt.Println(toString(data))
		// 对比测试结果
		index := Drop(data)
		//fmt.Println(toString(data))
		res := testDrop(data)
		//fmt.Println(index)
		//fmt.Println(res)
		if len(index) != len(res) {
			t.Error("1")
		}
		for j := 0; j < len(index); j++ {
			if index[j] != res[j] {
				t.Error("2")
			}
		}
	}
}

func TestNoLink(t *testing.T) {
	level := 3
	n := 10000
	sum := 0
	for j := 0; j < n; j++ {
		i := 0
		for {
			i++
			data, b := NewGems(rd, level)
			if b > 0 {
				continue
			}
			del := Check(level, data)
			var rate int64
			for j := 0; j < len(del); j++ {
				x, y := GetXY(level, del[j][0])
				rate += GemsRate(level, len(del[j]), data[y][x])
			}
			if rate == 0 {
				sum += i
				break
			}
		}
	}
	fmt.Println("平均搜索次数:", sum/n) // 平均搜索次数: 4
}

func TestLess(t *testing.T) {
	level := 3
	n := 10000
	sum := 0
	for j := 0; j < n; j++ {
		i := 0
		for {
			i++
			data, b := NewGems(rd, level)
			if b > 0 {
				continue
			}
			del := Check(level, data)
			var rate int64
			for j := 0; j < len(del); j++ {
				x, y := GetXY(level, del[j][0])
				rate += GemsRate(level, len(del[j]), data[y][x])
			}
			if rate < 10 {
				sum += i
				break
			}
		}
	}
	fmt.Println("平均搜索次数:", sum/n) // 平均搜索次数: 3
}

func TestAddGemsLess(t *testing.T) {
	var data [][]int32
	var b int32 = 1
	// 随机创建宝石
	level := rd.Intn(3) + 1
	for b > 0 {
		data, b = NewGems(rd, level)
	}
	// 随机删除宝石
	n := LevelGemsNum(level)
	delNum := rd.Intn(n)
	for i := 0; i < delNum; i++ {
		x, y := GetXY(level, rd.Intn(n))
		data[y][x] = NoGem
	}
	fmt.Println(ToString(data))
	// 对比测试结果
	index := Drop(data)
	fmt.Println(ToString(data))
	gems := AddGemsLess(rd, level, data, index)
	for i := 0; i < len(index); i++ {
		x, y := GetXY(level, index[i])
		data[y][x] = gems[i]
	}
	fmt.Println(ToString(data))
}

func TestNewGemsLessBomb(t *testing.T) {
	level := rd.Intn(3) + 1
	data := NewGemsLessBomb(rd, level)
	fmt.Println(ToString(data))
}

func TestNewGemsNoLink(t *testing.T) {
	var i int
	var arr [][]int
	var data [][]int32
	for len(arr) == 0 && i < 10000 {
		data = NewGemsNoLink(rd, 3)
		arr = Check(3, data)
		i++
	}
	if len(arr) > 0 {
		fmt.Println(ToString(data))
		t.Error()
	}
}

func TestNewGemsLess(t *testing.T) {
	data := NewGemsNoLink(rd, 1)
	fmt.Println(ToString(data))
}

func TestGemsC(t *testing.T) {
	data := GemsC(rd, 2)
	fmt.Println(ToString(data))
}

func TestAddGemsLess2(t *testing.T) {
	n := 1
	for i := 0; i < n; {
		level := 3
		data := GemsC(rd, level)
		AddGemsLess2(level, data, nil)
		arr := Check(level, data)
		if len(arr) == 1 {
			i++
			fmt.Println(ToString(data))
			for _, v := range arr[0] {
				x, y := GetXY(level, v)
				data[y][x] = 0
			}
			fmt.Println(ToString(data))
			index := Drop(data)
			fmt.Println(ToString(data))

			gems := AddGemsLess2(level, data, index)
			for i := 0; i < len(index); i++ {
				x, y := GetXY(level, index[i])
				data[y][x] = gems[i]
			}

			if len(Check(level, data)) > 0 {
				t.Error()
			}
		}
	}
}

func TestRateGems(t *testing.T) {
	//for k, v := range RateGems(1, 80) {
	//	fmt.Println(k, v)
	//}
	data := NewGemsRate(rd, 1, 6000)
	fmt.Println(ToString(data))
}

func TestNewGemsRate(t *testing.T) {
	for i := 1; i <= 3; i++ {
		for _, v := range rate {
			data := NewGemsRate(rd, i, v)
			del := Check(i, data)
			if len(del) == 0 {
				continue
			}
			var rate int64
			for j := 0; j < len(del); j++ {
				x, y := GetXY(i, del[j][0])
				rate += GemsRate(i, len(del[j]), data[y][x])
			}
			//fmt.Println(i, v, rate)
			//fmt.Println(ToString(data))
			if rate != int64(v) {
				t.Error("error")
			}
		}
	}
}
