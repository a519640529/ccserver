package rollpoint

import (
	"fmt"
	"testing"
)

func TestCreateBlackBox(t *testing.T) {
	box := CreateBlackBox()
	for i := 0; i < 50; i++ {
		box.Roll()
		//fmt.Println(box.Point)
		//fmt.Println(CalcPoint(box.Point))
	}
}

//var rate = [BetArea_Max]int32{
//	1,1,1,1,1,1,//Area1-6
//	60,30,20,12,8,6,6,6,6,8,12,20,30,60,//Sum4-17
//	1,1,//small,big
//	10,10,10,10,10,10,//Double1-6
//	30,//Boom
//	200,200,200,200,200,200,//Boom1-6
//}
func TestCalcPoint(t *testing.T) {
	type TestCase struct {
		point [RollNum]int32
		score [BetArea_Max]int32
	}
	testCase := []TestCase{
		{
			point: [RollNum]int32{0, 1, 2},
			score: [BetArea_Max]int32{
				1, 1, 1, 0, 0, 0,
				0, 0, 20, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				1, 0,
				0, 0, 0, 0, 0, 0,
				0,
				0, 0, 0, 0, 0, 0},
		}, {
			point: [RollNum]int32{1, 1, 2},
			score: [BetArea_Max]int32{
				0, 2, 1, 0, 0, 0,
				0, 0, 0, 12, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				1, 0,
				0, 10, 0, 0, 0, 0,
				0,
				0, 0, 0, 0, 0, 0},
		}, {
			point: [RollNum]int32{4, 3, 2},
			score: [BetArea_Max]int32{
				0, 0, 1, 1, 1, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 6, 0, 0, 0, 0, 0,
				0, 1,
				0, 0, 0, 0, 0, 0,
				0,
				0, 0, 0, 0, 0, 0},
		}, {
			point: [RollNum]int32{5, 5, 5},
			score: [BetArea_Max]int32{
				0, 0, 0, 0, 0, 3,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0,
				0, 0, 0, 0, 0, 10,
				30,
				0, 0, 0, 0, 0, 200},
		}, {
			point: [RollNum]int32{1, 1, 1},
			score: [BetArea_Max]int32{
				0, 3, 0, 0, 0, 0,
				0, 0, 20, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0,
				0, 10, 0, 0, 0, 0,
				30,
				0, 200, 0, 0, 0, 0},
		}, {
			point: [RollNum]int32{3, 4, 5},
			score: [BetArea_Max]int32{
				0, 0, 0, 1, 1, 1,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 20, 0, 0,
				0, 1,
				0, 0, 0, 0, 0, 0,
				0,
				0, 0, 0, 0, 0, 0},
		},
	}
	for _, value := range testCase {
		score := CalcPoint(value.point)
		if !Int32SliceEqual(score[:], value.score[:]) {
			fmt.Println(score)
			fmt.Println(value.score)
			t.Logf("Point %v calc score error:", value.point)
			t.Fatal("TestCalcPoint test fetal.")
		}
	}
}
func Int32SliceEqual(left []int32, right []int32) bool {
	if len(left) != len(right) {
		return false
	}
	for i := 0; i < len(left); i++ {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}
func TestGlobalData(t *testing.T) {
	fmt.Println(MaxChipColl)
}
