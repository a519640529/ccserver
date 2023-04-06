package fortunezhishen

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"
)

func TestResult(t *testing.T) {

	//a :=,{]int{
	//	A, 			A, Gemstone, Gemstone, 	K,
	//	Gemstone, Jade, 	A, 		K,		 A,
	//	Jade, 		Q, 		K, 		A, 		Jade,
	//}
	eleLineAppearRate := []int32{1, 568, 568, 568, 568, 568, 909, 1136, 1136, 1136, 1136, 1, 568}
	fmt.Println(eleLineAppearRate)

	rand.Seed(time.Now().UnixNano())
	slotRateWeight := []int32{17000, 17000, 3000, 1500, 1000, 800, 700, 600, 500, 400, 200, 100, 80, 60, 40, 35, 30, 25, 20, 18, 16, 14, 12, 10, 9, 8, 7, 6, 5, 4, 3, 2}
	levelRate := [][]int32{{0, 0}, {1, 50}, {51, 80}, {81, 120}, {121, 150}, {151, 180}, {181, 210},
		{211, 240}, {241, 270}, {271, 300}, {301, 350}, {351, 400}, {401, 450}, {451, 500}, {501, 600}, {601, 700}, {701, 800},
		{801, 900}, {901, 1000}, {1001, 1500}, {1501, 2000}, {2001, 2500}, {2501, 3000}, {3001, 3500}, {3501, 4000},
		{4001, 4500}, {4501, 5000}, {5001, 6000}, {6001, 7000}, {7001, 8000}, {8001, 9000}, {9001, 10000}}

	//GetLineEleVal(0, 1, 500000, eleLineAppearRate)

	fileName := fmt.Sprintf("fortunezhishen-%v-%d.csv", 0, 0)
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm)
	defer file.Close()
	if err != nil {
		file, err = os.Create(fileName)
		if err != nil {
			return
		}
	}
	for i := 0; i < 100000; i++ {
		rIdx := RandSliceInt32IndexByWightN(slotRateWeight)
		parsec := levelRate[rIdx]
		var needRate int32
		if parsec[0] != 0 && parsec[1] != 0 {
			needRate = rand.Int31n(parsec[1]-parsec[0]) + parsec[0]
		}
		str := fmt.Sprintf("%v\r\n", needRate)
		file.WriteString(str)
	}

}
