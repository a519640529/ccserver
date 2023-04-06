package fortunezhishen

import (
	"fmt"
	"math/rand"
	"strconv"
)

func GetLineEleVal(betIdx, gameState int, needRate int64, eleLineAppearRate []int32) (WinResult, []int, [][]int32) {
	var preInt [][]int32
	for i := 0; i < 1000; i++ {
		var wls WinResult
		wls.CreateLine(gameState, eleLineAppearRate)
		wls.Win(betIdx, gameState)
		var rate int64
		for _, v := range wls.WinLine {
			rate += v.Rate
		}
		//fmt.Printf("%v  || rate  %v", wls.Cards, rate)
		//fmt.Println()
		var n int64 = 30
		if gameState == FreeGame {
			n = 50
		}
		if rate >= needRate-10 && rate <= needRate+n && wls.GemstoneNum < 6 && wls.MakeAFortuneNum < 3 {
			var poss []int32
			for _, v := range wls.WinLine {
				poss = append(poss, v.Poss...)
			}
			var noPoss []int
			for k, v := range wls.Cards {
				isF := false
				for _, pn := range poss {
					if k == int(pn) {
						isF = true
						break
					}
				}
				if !isF && v != MakeAFortune && v != Gemstone {
					noPoss = append(noPoss, k)
				}
			}
			//fmt.Println("...........find rate: ", rate, " 第 ", i+1, " 次.")
			return wls, noPoss, nil
		}
		if rate < 50 && len(preInt) < 10 && wls.GemstoneNum < 6 && wls.MakeAFortuneNum < 3 {
			preInt = append(preInt, wls.Cards)
		}
	}
	return WinResult{}, nil, preInt
}
func FindInxInArray(a int, b []int) bool {
	for _, v := range b {
		if a == v {
			return true
		}
	}
	return false
}
func RandSliceInt32IndexByWightN(s1 []int32) int32 {
	total := 0
	for _, v := range s1 {
		total += int(v)
	}
	if total <= 0 {
		return 0
	}
	random := rand.Intn(total)
	total = 0
	for i, v := range s1 {
		total += int(v)
		if random < total {
			return int32(i)
		}
	}
	return 0
}
func randSliceIndexByWightN(s1 []int) int {
	total := 0
	for _, v := range s1 {
		total += v
	}
	if total <= 0 {
		return 0
	}
	random := rand.Intn(total)
	total = 0
	for i, v := range s1 {
		total += v
		if random < total {
			return i
		}
	}
	return 0
}
func PrintWin(lines []int32) {
	str := ""
	for _, ele := range lines {
		switch ele {
		case MakeAFortune:
			str += "发财,"
		case Wild:
			str += "财神,"
		case Firecrackers:
			str += "鞭炮,"
		case Drum:
			str += "鼓  ,"
		case Jade:
			str += "玉牌,"
		case Copper:
			str += "铜币,"
		case A:
			str += "A  ,"
		case K:
			str += "K  ,"
		case Q:
			str += "Q  ,"
		case J:
			str += "J  ,"
		case T:
			str += "T  ,"
		case Gemstone:
			str += "宝石,"
		}
	}
	fmt.Println(str)
}
func Print(res []int32) {
	fmt.Println(res)
	str := ""
	for k, ele := range res {
		switch ele {
		case MakeAFortune:
			str += "发财,"
		case Wild:
			str += "财神,"
		case Firecrackers:
			str += "鞭炮,"
		case Drum:
			str += "鼓  ,"
		case Jade:
			str += "玉牌,"
		case Copper:
			str += "铜币,"
		case A:
			str += "A  ,"
		case K:
			str += "K  ,"
		case Q:
			str += "Q  ,"
		case J:
			str += "J  ,"
		case T:
			str += "T  ,"
		case Gemstone:
			str += "宝石,"
		}
		if (k+1)%5 == 0 {
			fmt.Println("第", strconv.Itoa((k+1)/5), "行     ", str)
			str = ""
		}
	}
}
