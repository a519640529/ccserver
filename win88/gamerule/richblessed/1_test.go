package richblessed

import (
	"fmt"
	"testing"
)

func TestName(t *testing.T) {
	//eleLineAppearRate := []int32{926, 926, 556, 556, 741, 741, 741, 926, 1111, 1296, 1481, 5000, 5000}
	// 0万能元素 福字 Gongs 1 铜锣 GoldenPhoenix 2 金凤凰 Sailboat 3 //帆船GoldenTortoise  4金龟
	// GoldIngot 5 金元宝Copper 6 金钱币 A 7  A  K  8 K  Q 9  Q  J 10 J  Ten 11 10 Nine 12 9
	//eleLineAppearRate := []int32{100, 50, 60, 50, 50, 60, 50, 0, 0, 0, 0, 0, 0}

	//fileName := fmt.Sprintf("classic888-%v-%d.csv", 0, 0)
	//file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm)
	//defer file.Close()
	//if err != nil {
	//	file, err = os.Create(fileName)
	//	if err != nil {
	//		return
	//	}
	//}
	//file.WriteString("随机倍率\n")

	for i := 0; i < 20; i++ {
		var wls WinResult
		//	wls.CreateLine(eleLineAppearRate)
		wls.Win(10, 10)
		var rate int64
		for _, v := range wls.WinLine {
			rate += v.Rate
			fmt.Println(v.LineId, v.Lines, v.Poss)
		}
		fmt.Println(wls.WinLine, wls.AllRate, wls.FreeNum)
		if len(wls.WinLine) > 0 {
			break
		}
		//str := fmt.Sprintf("%v\r\n", rate)
		//file.WriteString(str)
	}

}

func TestWin(t *testing.T) {
	//eleLineAppearRate := []int32{926, 926, 556, 556, 741, 741, 741, 926, 1111, 1296, 1481, 5000, 5000}
	//eleLineAppearRate := []int32{100, 50, 60, 50, 50, 60, 50, 0, 0, 0, 0, 0, 0}

	var wls WinResult
	//wls.CreateLine(eleLineAppearRate)
	wls.EleValue = []int32{1, 1, 1, 0, 5, 6, 2, 7, 7, 9, 4, 7, 4, 8, 7}
	for i := 0; i != 1000; i++ {
		if RandJACKPOT(1000, 1000) {

			fmt.Println("i:", i)
			break
		}
	}
	wls.Win(10, 10)
	var rate int64
	for _, v := range wls.WinLine {
		rate += v.Rate
		fmt.Println(v.LineId, v.Lines, v.Poss)
	}
	fmt.Println(wls.WinLine, rate, wls.FreeNum)

}

func TestJackWin(t *testing.T) {
	//eleLineAppearRate := []int32{926, 926, 556, 556, 741, 741, 741, 926, 1111, 1296, 1481, 5000, 5000}
	//eleLineAppearRate := []int32{100, 50, 60, 50, 50, 60, 50, 0, 0, 0, 0, 0, 0}

	var wls WinResult
	//wls.CreateLine(eleLineAppearRate)
	wls.EleValue = []int32{1, 0, 1, 0, 5, 6, 2, 7, 7, 9, 4, 7, 4, 8, 7}
	JACKPOTElementsParams := []int32{20, 30, 100, 150}
	ret := wls.CanJACKPOT(1000, 1000)
	if ret {
		ele := wls.CreateJACKPOT(JACKPOTElementsParams)
		fmt.Println(ret, ele)
	}
	wls.Win(10, 10)
	var rate int64
	for _, v := range wls.WinLine {
		rate += v.Rate
		fmt.Println(v.LineId, v.Lines, v.Poss)
	}
	fmt.Println(wls.WinLine, rate, wls.FreeNum)

}
