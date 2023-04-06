package avengers

import (
	"fmt"
	"math/rand"
	"sort"
	"testing"
	"time"
)

func TestIsLine(t *testing.T) {
	type TestData struct {
		data []int
		line int
	}
	testData := []TestData{
		{data: []int{0, 0, 0, 1, 1}, line: 3},
		{data: []int{0, 0, 0, 0, 1}, line: 4},
		{data: []int{0, 0, 0, 0, 0}, line: 5},
	}
	for _, value := range testData {
		if _, count := isLine(value.data); count != value.line {
			t.Error(isLine(value.data))
			t.Error("Error line data:", value)
			t.Fatal("TestIsLine")
		}
	}
	errorData := []TestData{
		{data: []int{1, 0, 0, 0, 1}, line: -1},
		{data: []int{1, 1, 0, 0, 0}, line: -1},
		{data: []int{1, 1, 0, 1, 1}, line: -1},
	}
	for _, value := range errorData {
		if _, count := isLine(value.data); count != value.line {
			t.Error(isLine(value.data))
			t.Error("Error data:", value)
			t.Fatal("TestIsLine")
		}
	}
}
func TestCalcLine(t *testing.T) {
	type TestData struct {
		data []int
		line int
	}
	testData := []TestData{
		{data: []int{2, 1, 3, 4, 1, 5, 6, 1, 7, 8, 1, 0, 8, 1, 7}, line: 1},
		{data: []int{6, 7, 5, 7, 6, 5, 6, 8, 8, 7, 7, 7, 6, 2, 9}, line: 1},
		{data: []int{10, 5, 9, 1, 9, 2, 7, 9, 9, 9, 5, 3, 2, 2, 9}, line: 4},
		{data: []int{3, 10, 6, 1, 3, 7, 3, 3, 3, 3, 7, 6, 10, 8, 0}, line: 6},
		{data: []int{9, 7, 10, 1, 5, 10, 4, 6, 9, 5, 7, 5, 1, 2, 4}, line: 2},
		{data: []int{9, 3, 0, 2, 0, 0, 3, 9, 0, 5, 9, 2, 2, 8, 2}, line: 2},
	}
	for _, value := range testData {
		lines := CalcLine(value.data, AllBetLines)
		if len(lines) != value.line {
			t.Log("lines:", lines)
			t.Log("Error line data:", value.data)
			t.Fatal("TestIsLine")
		}
	}
}
func TestCalcLineScore(t *testing.T) {
	lines := CalcLine([]int{9, 10, 7, 9, 7, 6, 7, 6, 10, 10, 1, 10, 8, 10, 4}, AllBetLines)
	t.Log(lines)
	line, allscore, _, _, _ := CaclScore([]int{9, 10, 7, 9, 7, 6, 7, 6, 10, 10, 1, 10, 8, 10, 4}, AllBetLines)
	PrintHuman([]int{9, 10, 7, 9, 7, 6, 7, 6, 10, 10, 1, 10, 8, 10, 4})
	t.Logf("lineNum:%v allScore:%v", line, allscore)
	t.Fatal("TestCalcLineScore")
}

func TestRandCalcLineScore(t *testing.T) {
	var cards, c = generateSlotsData()
	t.Logf("尝试次数：%v次，Data:%v", c, cards)
	PrintHuman(cards)

	lines := CalcLine(cards, AllBetLines)
	for _, line := range lines {
		t.Log(AllLineDraw[line.Index])
	}

	t.Log(lines)
	line, allscore, _, _, _ := CaclScore(cards, AllBetLines)
	t.Logf("lineNum:%v allScore:%v", line, allscore)
	t.Fatal("TestRandCalcLineScore")
}
func TestDelSliceRepEle(t *testing.T) {
	type TestData struct {
		data  []int
		rdata []int
	}
	testData := []TestData{
		{data: []int{1}, rdata: []int{1}},
		{data: []int{1, 2}, rdata: []int{1, 2}},
		{data: []int{1, 1, 2}, rdata: []int{1, 2}},
		{data: []int{1, 2, 2}, rdata: []int{1, 2}},
		{data: []int{1, 2, 2, 3}, rdata: []int{1, 2, 3}},
		{data: []int{1, 2, 2, 3, 3}, rdata: []int{1, 2, 3}},
		{data: []int{1, 1, 2, 2, 3, 3}, rdata: []int{1, 2, 3}},
		{data: []int{1, 2, 3, 3}, rdata: []int{1, 2, 3}},
		{data: []int{1, 1, 1, 1}, rdata: []int{1}},
	}
	for _, value := range testData {
		rdata := DelSliceRepEle(value.data)
		if !SliceEqual(rdata, value.rdata) {
			t.Error(value.data)
			t.Error(rdata)
			t.Fatal("TestDelSliceRepEle")
		}
	}
}

/*
 * 切片是否相等
 */
func SliceEqual(left []int, right []int) bool {
	if len(left) != len(right) {
		return false
	}
	sort.Ints(left)
	sort.Ints(right)
	for i := 0; i < len(left); i++ {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

//func TestRollNormalElement(t *testing.T) {
//	card := make([]int, ROLLLINEMAX, ROLLLINEMAX)
//	count := 100000
//	now := time.Now()
//	for i := 0; i < count; i++ {
//		card = RollNormalElement(card)
//	}
//	t.Logf("RollNormalElement rand %v data cost %v second.", count, time.Now().Sub(now).Seconds())
//}
//func TestRollFreeElement(t *testing.T) {
//	card := make([]int, ROLLLINEMAX, ROLLLINEMAX)
//	count := 100000
//	now := time.Now()
//	for i := 0; i < count; i++ {
//		card = RollFreeElement(card, rand.Intn(2))
//	}
//	t.Logf("RollFreeElement rand %v data cost %v second.", count, time.Now().Sub(now).Seconds())
//}

func generateSlotsData() ([]int, int) {
	var tryCount = 0
Next:
	var slotsData = make([]int, 0, ELEMENT_TOTAL)
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < ELEMENT_TOTAL; i++ {
		slotsData = append(slotsData, symbolSmall[rand.Intn(len(symbolSmall))])
	}
	tryCount++
	fmt.Print("tryCount:", tryCount)
	if CheckBigWin(slotsData) {
		goto Next
	}
	return slotsData, tryCount
}

var symbolSmall = []int{1, 2, 3, 3, 5, 5, 5, 5, 5, 5, 7, 7, 7, 7, 7, 6, 6, 6, 6, 6, 7, 4, 4, 4, 4, 4, 1, 1, 5, 5, 7, 7, 7, 8, 8, 8, 8, 9, 9, 9, 9, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10}

//随机取元素（从不同的分布里面取）

//该玩家的投入产出比来决定是否会出现两个通配同时还与水池有关
//水池过多的时候可以允许两个通配
//正常水池只允许一个通配
//水池太低则从牌库里面取不中奖的数据

func TestGenerateBonusGame(t *testing.T) {
	type args struct {
		totalBet   int
		startBonus int
	}
	tests := []struct {
		name string
		args args
		want BonusGameResult
	}{
		{
			name: "bet:100 bonus*3",
			args: args{
				totalBet:   100,
				startBonus: 1,
			},
			want: BonusGameResult{
				BonusData:       nil,
				DataMultiplier:  0,
				Mutiplier:       0,
				TotalPrizeValue: 0,
			},
		},
		{
			name: "bet:100 bonus*5",
			args: args{
				totalBet:   100,
				startBonus: 3,
			},
			want: BonusGameResult{
				BonusData:       nil,
				DataMultiplier:  0,
				Mutiplier:       0,
				TotalPrizeValue: 0,
			},
		},
		{
			name: "bet:1000 bonus*2",
			args: args{
				totalBet:   100,
				startBonus: 3,
			},
			want: BonusGameResult{
				BonusData:       nil,
				DataMultiplier:  0,
				Mutiplier:       0,
				TotalPrizeValue: 0,
			},
		},
		{
			name: "bet:1000 bonus*0",
			args: args{
				totalBet:   1000,
				startBonus: 0,
			},
			want: BonusGameResult{
				BonusData:       nil,
				DataMultiplier:  0,
				Mutiplier:       0,
				TotalPrizeValue: 0,
			},
		},
		{
			name: "bet:0 bonus*3",
			args: args{
				totalBet:   0,
				startBonus: 3,
			},
			want: BonusGameResult{
				BonusData:       nil,
				DataMultiplier:  0,
				Mutiplier:       0,
				TotalPrizeValue: 0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := GenerateBonusGame(tt.args.totalBet, tt.args.startBonus)
			t.Logf("BonusData:%v Len:%v", r.BonusData, len(r.BonusData))
			t.Logf("TotalPrizeValue:%v", r.TotalPrizeValue)
			t.Logf("Mutiplier:%v", r.Mutiplier)
			t.Logf("DataMultiplier:%v", r.DataMultiplier)
		})
	}
}

func TestGenerateSlotsData_v2(t *testing.T) {
	type args struct {
		s Symbol
	}
	tests := []struct {
		name  string
		args  args
		want  []int
		want1 int
	}{
		{
			name: "01",
			args: args{
				s: SYMBOL1,
			},
			want:  nil,
			want1: 0,
		},
		{
			name: "02",
			args: args{
				s: SYMBOL2,
			},
			want:  nil,
			want1: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, times := GenerateSlotsData_v2(tt.args.s)
			t.Logf("GenerateSlotsData_v2() data = %v, times %v", v, times)
			lines := CalcLine(v, AllBetLines)
			t.Logf("lines:%v", lines)
			PrintHuman(v)
			CaclScore(v, AllBetLines)
		})
	}
}
