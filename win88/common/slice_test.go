package common

import (
	"math/rand"
	"sort"
	"testing"
)

func TestInsertValueToSlice(t *testing.T) {
	type InsertValueToSliceTestCase struct {
		arr   []int
		index int
		value int
	}
	testData := []InsertValueToSliceTestCase{
		{arr: []int{100, 90, 80, 70, 60, 50, 40, 30, 20, 10}, index: 0, value: 110},
		{arr: []int{100, 90, 80, 70, 60, 50, 40, 30, 20, 10}, index: 0, value: 1},
		{arr: []int{100, 90, 80, 70, 60, 50, 40, 30, 20, 10}, index: 0, value: 55},
		{arr: []int{100, 90, 80, 70, 60, 50, 40, 30, 20, 10}, index: 5, value: 110},
		{arr: []int{100, 90, 80, 70, 60, 50, 40, 30, 20, 10}, index: 5, value: 1},
		{arr: []int{100, 90, 80, 70, 60, 50, 40, 30, 20, 10}, index: 5, value: 55},
		{arr: []int{100, 90, 80, 70, 60, 50, 40, 30, 20, 10}, index: 9, value: 110},
		{arr: []int{100, 90, 80, 70, 60, 50, 40, 30, 20, 10}, index: 9, value: 1},
		{arr: []int{100, 90, 80, 70, 60, 50, 40, 30, 20, 10}, index: 9, value: 55},
	}
	for _, value := range testData {
		newArr := InsertValueToSlice(value.index, value.value, value.arr)
		if !IsLadderSlice(newArr) || len(newArr) != len(value.arr) {
			t.Log(value)
			t.Failed()
		}
	}
}
func TestSliceNoRepeate(t *testing.T) {
	type TestCase struct {
		Src  []int
		Dest []int
	}
	var testData = []TestCase{
		TestCase{Src: []int{}, Dest: []int{}},
		TestCase{Src: []int{1}, Dest: []int{1}},
		TestCase{Src: []int{1, 1}, Dest: []int{1}},
		TestCase{Src: []int{1, 1, 1}, Dest: []int{1}},
		TestCase{Src: []int{1, 2, 1}, Dest: []int{1, 2}},
		TestCase{Src: []int{1, 2, 3}, Dest: []int{1, 2, 3}},
		TestCase{Src: []int{1, 2, 3, 3}, Dest: []int{1, 2, 3}},
		TestCase{Src: []int{1, 2, 3, 3, 5}, Dest: []int{1, 2, 3, 5}},
		TestCase{Src: []int{1, 2, 3, 3, 5, 5}, Dest: []int{1, 2, 3, 5}},
		TestCase{Src: []int{1, 2, 3, 3, 5, 5, 6}, Dest: []int{1, 2, 3, 5, 6}},
		TestCase{Src: []int{1, 2, 3, 3, 5, 6, 6}, Dest: []int{1, 2, 3, 5, 6}},
		TestCase{Src: []int{1, 2, 3, 3, 5, 6, 7}, Dest: []int{1, 2, 3, 5, 6, 7}},
	}
	for _, value := range testData {
		dest := SliceNoRepeate(value.Src)
		equal := SliceIntEqual(dest, value.Dest)
		if !equal {
			t.Error(value)
			t.Error(dest)
			t.Fatal()
		}
	}
	for i := 0; i < 1000; i++ {
		length := rand.Intn(100) + 1
		src := rand.Perm(length)
		dest := make([]int, length)
		copy(dest, src)
		e := rand.Intn(50)
		for s := 0; s < e; s++ {
			src = append(src, rand.Intn(length))
		}
		norepeate := SliceNoRepeate(src)
		equal := SliceIntEqual(norepeate, dest)
		if !equal {
			sort.Ints(src)
			sort.Ints(dest)
			sort.Ints(norepeate)
			t.Error(src)
			t.Error(dest)
			t.Error(norepeate)
			t.Fatal()
		}
	}
}
func TestSliceDelValue(t *testing.T) {
	type DelTestData struct {
		Arr []int32
		Del int32
	}
	var TestCase = []DelTestData{
		{Arr: []int32{1, 2, 3, 4, 5}, Del: 1},
		{Arr: []int32{1, 2, 3, 4, 5}, Del: 3},
		{Arr: []int32{1, 2, 3, 4, 5}, Del: 5},
		{Arr: []int32{1, 2, 3, 4, 5}, Del: 6},
		{Arr: []int32{1, 2, 3, 4, 5}, Del: 0},
	}
	for i := 0; i < len(TestCase); i++ {
		randArr := DelSliceInt32(TestCase[i].Arr, TestCase[i].Del)
		if InSliceInt32(randArr, TestCase[i].Del) {
			t.Log(randArr)
			t.Log(TestCase[i].Del)
			t.Fatal("Slice del value failed.")
		}
	}
	for i := 0; i < 1000; i++ {
		randArr := rand.Perm(rand.Intn(1000))
		delValue := rand.Intn(1000)
		randArr = DelSliceInt(randArr, delValue)
		if InSliceInt(randArr, delValue) {
			t.Log(randArr)
			t.Log(delValue)
			t.Fatal("Slice del value failed.")
		}
	}
}
