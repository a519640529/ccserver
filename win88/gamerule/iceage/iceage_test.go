package iceage

import (
	"fmt"
	"math/rand"
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
		if _, count := IsLine(value.data); count != value.line {
			t.Error(IsLine(value.data))
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
		if _, count := IsLine(value.data); count != value.line {
			t.Error(IsLine(value.data))
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
	lines := CalcLine([]int{9, 10, 7, 9, 7, 6, 7, 6, 10, 10, 0, 10, 8, 10, 4}, AllBetLines)
	t.Log(lines)
	line, allscore, _, _, _ := CaclScore([]int{9, 10, 7, 9, 7, 6, 7, 6, 10, 10, 0, 10, 8, 10, 4}, AllBetLines)
	t.Logf("lineNum:%v allScore:%v", line, allscore)
	t.Fatal("TestCalcLineScore")
}
func TestRandCalcLineScore(t *testing.T) {
	var cards, c = generateSlotsData()
	t.Logf("尝试次数：%v次，Data:%v", c, cards)
	PrintHuman(cards)
	lines := CalcLine(cards, AllBetLines)
	t.Log(lines)
	line, allscore, _, _, _ := CaclScore(cards, AllBetLines)
	t.Logf("lineNum:%v allScore:%v", line, allscore)
	t.Fatal("TestRandCalcLineScore")
}
func TestMakePlan(t *testing.T) {
	type TestData struct {
		data []int
		prms []int32
	}
	testData := []TestData{
		{data: []int{6, 6, 5, 7, 7, 3, 4, 6, 7, 5, 5, 1, 3, 4, 5}, prms: []int32{1, 2, 3, 3, 1, 5, 5, 5, 5, 5, 7, 7, 7, 7, 7, 6, 6, 6, 6, 6, 6, 4, 4, 4, 4, 4, 4, 4}},
		{data: []int{4, 2, 3, 5, 1, 7, 4, 3, 4, 6, 5, 5, 4, 7, 3}, prms: []int32{1, 2, 3, 3, 1, 5, 5, 5, 5, 5, 7, 7, 7, 7, 7, 6, 6, 6, 6, 6, 6, 4, 4, 4, 4, 4, 4, 4}},
		{data: []int{6, 2, 2, 7, 5, 6, 6, 6, 6, 6, 4, 6, 4, 3, 2}, prms: []int32{1, 2, 3, 3, 1, 5, 5, 5, 5, 5, 7, 7, 7, 7, 7, 6, 6, 6, 6, 6, 6, 4, 4, 4, 4, 4, 4, 4}},
	}
	for _, value := range testData {
		t.Log("old:", value.data)
		line := MakePlan(value.data, value.prms, []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20})
		t.Log("new:", line)
	}
}
func generateSlotsData() ([]int, int) {
	var tryCount = 0
	var slotsData = make([]int, 0, ELEMENT_TOTAL)
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < ELEMENT_TOTAL; i++ {
		slotsData = append(slotsData, symbolSmall[rand.Intn(len(symbolSmall))])
	}
	tryCount++
	fmt.Print("tryCount:", tryCount)
	return slotsData, tryCount
}

var symbolSmall = []int{1, 2, 3, 3, 5, 5, 5, 5, 5, 5, 7, 7, 7, 7, 7, 6, 6, 6, 6, 6, 7, 4, 4, 4, 4, 4, 1, 1, 5, 5, 7, 7, 7}
