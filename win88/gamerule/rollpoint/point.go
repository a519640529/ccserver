package rollpoint

import (
	"math/rand"
	"time"
)

const (
	RollNum = 3
)
const (
	BetArea_1 int32 = iota
	BetArea_2
	BetArea_3
	BetArea_4
	BetArea_5
	BetArea_6    //5
	BetArea_Sum4 //6
	BetArea_Sum5
	BetArea_Sum6
	BetArea_Sum7
	BetArea_Sum8
	BetArea_Sum9
	BetArea_Sum10
	BetArea_Sum11
	BetArea_Sum12
	BetArea_Sum13
	BetArea_Sum14
	BetArea_Sum15
	BetArea_Sum16
	BetArea_Sum17
	BetArea_Small   //20
	BetArea_Big     //21
	BetArea_Double1 //22
	BetArea_Double2
	BetArea_Double3
	BetArea_Double4
	BetArea_Double5
	BetArea_Double6
	BetArea_Boom  //28
	BetArea_Boom1 //29
	BetArea_Boom2
	BetArea_Boom3
	BetArea_Boom4
	BetArea_Boom5
	BetArea_Boom6 //34
	BetArea_Max
)

var rate = [BetArea_Max]int32{
	1, 1, 1, 1, 1, 1, //Area1-6
	60, 30, 20, 12, 8, 6, 6, 6, 6, 8, 12, 20, 30, 60, //Sum4-17
	1, 1, //small,big
	10, 10, 10, 10, 10, 10, //Double1-6
	30,                           //Boom
	200, 200, 200, 200, 200, 200, //Boom1-6
}

// 5             6     7    8     9     10
// 9/10/11/12| 8/13| 7/14| 6/15| 5/16| 4/17
var AreaIndex2MaxChipIndex = [BetArea_Max]int32{
	1, 1, 1, 1, 1, 1, //Area1-6
	10, 9, 8, 7, 6, 5, 5, 5, 5, 6, 7, 8, 9, 10, //Sum4-17
	0, 0, //small,big
	2, 2, 2, 2, 2, 2, //Double1-6
	3,                //Boom
	4, 4, 4, 4, 4, 4, //Boom1-6
}
var MaxChipColl = [BetArea_Max][]int32{}
var AllScore = [][BetArea_Max]int32{}
var AllRoll = [][RollNum]int32{}

type BlackBox struct {
	first  *rand.Rand
	second *rand.Rand
	third  *rand.Rand
	count  int
	Point  [RollNum]int32
	Score  [BetArea_Max]int32
}

func CreateBlackBox() *BlackBox {
	return &BlackBox{
		count:  0,
		first:  rand.New(rand.NewSource(rand.Int63n(time.Now().UnixNano()))),
		second: rand.New(rand.NewSource(rand.Int63n(time.Now().UnixNano()))),
		third:  rand.New(rand.NewSource(rand.Int63n(time.Now().UnixNano()))),
		Point:  [RollNum]int32{},
	}
}
func (bb *BlackBox) Roll() {
	if bb.count > 10 {
		bb.first = rand.New(rand.NewSource(rand.Int63n(time.Now().UnixNano())))
		bb.second = rand.New(rand.NewSource(rand.Int63n(time.Now().UnixNano())))
		bb.third = rand.New(rand.NewSource(rand.Int63n(time.Now().UnixNano())))
		bb.count = 0
	}
	bb.count++
	bb.Point[0] = bb.first.Int31n(6)
	bb.Point[1] = bb.second.Int31n(6)
	bb.Point[2] = bb.third.Int31n(6)
	bb.Score = CalcPoint(bb.Point)
}
func CalcPoint(point [RollNum]int32) [BetArea_Max]int32 {
	score := [BetArea_Max]int32{}
	if point[0] == -1 || point[1] == -1 || point[2] == -1 {
		point[0] = rand.Int31n(6)
		point[1] = rand.Int31n(6)
		point[2] = rand.Int31n(6)
	}
	sum := int32(0)
	//Area1-6
	for _, value := range point {
		score[value] += 1
		sum += (value + 1)
	}
	//Sum4-17
	if sum >= 4 && sum <= 17 {
		index := (sum - 4) + BetArea_Sum4
		score[index] = rate[index]
	}
	//small
	if sum >= 4 && sum <= 10 {
		score[BetArea_Small] = rate[BetArea_Small]
	}
	//big
	if sum >= 11 && sum <= 17 {
		score[BetArea_Big] = rate[BetArea_Big]
	}
	//Double1-6
	if point[0] == point[1] {
		index := point[0] + BetArea_Double1
		score[index] = rate[index]
	}
	if point[0] == point[2] {
		index := point[0] + BetArea_Double1
		score[index] = rate[index]
	}
	if point[1] == point[2] {
		index := point[1] + BetArea_Double1
		score[index] = rate[index]
	}
	//Boom Boom1-6
	if point[0] == point[1] && point[1] == point[2] {
		//Boom
		score[BetArea_Boom] += rate[BetArea_Boom]
		//Boom1-6
		index := point[0] + BetArea_Boom1
		score[index] = rate[index]
		score[BetArea_Small] = 0
		score[BetArea_Big] = 0
	}
	return score
}
func init() {
	for i := int32(0); i < 6; i++ {
		for j := int32(0); j < 6; j++ {
			for k := int32(0); k < 6; k++ {
				AllScore = append(AllScore, CalcPoint([RollNum]int32{i, j, k}))
				AllRoll = append(AllRoll, [RollNum]int32{i, j, k})
			}
		}
	}
	for i := int32(0); i < BetArea_Max; i++ {
		arr := []int32{}
		for k := int32(0); k < BetArea_Max; k++ {
			if AreaIndex2MaxChipIndex[k] == AreaIndex2MaxChipIndex[i] {
				arr = append(arr, k)
			}
		}
		MaxChipColl[i] = arr
	}
}
