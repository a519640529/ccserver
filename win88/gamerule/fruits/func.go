package fruits

import (
	"fmt"
	"math/rand"
	"strconv"
)

func GetLineEleVal(gameState int, needRate int64, eleLineAppearRate [][]int32, isLow bool) (WinResult, []int, [][]int32) {
	var preInt [][]int32
	for i := 0; i < 1000; i++ {
		var wls WinResult
		wls.CreateLine(eleLineAppearRate)
		wls.Win()
		var rate int64
		var bonusNum int
		var wildNum int
		for _, v := range wls.WinLine {
			if len(v.Lines) == 0 {
				continue
			}
			rate += v.Rate
			if v.Lines[0] == Bonus {
				bonusNum += len(v.Lines)
			} else if v.Lines[0] == Wild {
				wildNum += len(v.Lines)
			}
			NowWildNum := 0
			for _, l := range v.Lines {
				if l != Wild && NowWildNum > 0 {
					if NowWildNum < 3 {
						NowWildNum = 0
					}
				} else if l == Wild {
					NowWildNum++
				}
			}
			if NowWildNum >= 3 {
				wildNum += NowWildNum
			}
		}

		//fmt.Printf("%v  || rate  %v", wls.EleValue, rate)
		//fmt.Println()
		var n int64 = 5
		if gameState == FreeGame {
			n = 50
		}
		if wildNum >= 3 || bonusNum >= 3 {
			continue
		}
		if isLow && wls.JackPotNum >= 3 {
			continue
		}
		if rate >= needRate-n && rate <= needRate+n {
			var poss []int32
			for _, v := range wls.WinLine {
				poss = append(poss, v.Poss...)
			}
			var noPoss []int
			for k := range wls.EleValue {
				isF := false
				for _, pn := range poss {
					if k == int(pn) {
						isF = true
						break
					}
				}
				if !isF {
					noPoss = append(noPoss, k)
				}
			}
			//fmt.Println("...........find rate: ", rate, " 第 ", i+1, " 次.")
			return wls, noPoss, nil
		}
		if rate != 0 && rate < 50 && len(preInt) < 10 {
			preInt = append(preInt, wls.EleValue)
		}
	}
	return WinResult{}, nil, preInt
}
func GetLinePos(lineId int) []int {
	if lineId <= 9 || lineId >= 1 {
		return LineWinNum[lineId-1]
	}
	return nil
}
func GetRate(ele int32, num int) int64 {
	if data, ok := EleNumRate[ele]; ok {
		if r, ok2 := data[num]; ok2 {
			return r
		}
	}
	return 0
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
func PrintFruit(idx int32) (str string) {
	switch idx {
	case Wild:
		str += "Wild"
	case Bonus:
		str += "Bonus"
	case Scatter:
		str += "SCATTER"
	case Bar:
		str += "Bar"
	case Cherry:
		str += "樱桃"
	case Bell:
		str += "铃铛"
	case Pineapple:
		str += "菠萝"
	case Grape:
		str += "葡萄"
	case Lemon:
		str += "柠檬"
	case Watermelon:
		str += "西瓜"
	case Banana:
		str += "香蕉"
	case Apple:
		str += "苹果"
	case Bomb:
		str += "炸弹"
	}
	return str
}

func Print(res []int32) {
	fmt.Println(res, len(res))
	str := ""
	for k, ele := range res {
		switch ele {
		case Wild:
			str += "Wild,"
		case Bonus:
			str += "Bonus,"
		case Scatter:
			str += "Scatter,"
		case Bar:
			str += "Bar,"
		case Cherry:
			str += "樱桃,"
		case Bell:
			str += "铃铛,"
		case Pineapple:
			str += "菠萝,"
		case Grape:
			str += "葡萄,"
		case Lemon:
			str += "柠檬,"
		case Watermelon:
			str += "西瓜,"
		case Apple:
			str += "苹果,"
		case Banana:
			str += "香蕉,"
		}
		if (k+1)%5 == 0 {
			fmt.Println("第", strconv.Itoa((k+1)/5), "行     ", str)
			str = ""
		}
	}
}
func PrintWin(lines []int32) {
	str := ""
	for _, ele := range lines {
		switch ele {
		case Wild:
			str += "Wild,"
		case Bonus:
			str += "Bonus,"
		case Scatter:
			str += "Scatter,"
		case Bar:
			str += "Bar,"
		case Cherry:
			str += "樱桃,"
		case Bell:
			str += "铃铛,"
		case Pineapple:
			str += "菠萝,"
		case Grape:
			str += "葡萄,"
		case Lemon:
			str += "柠檬,"
		case Watermelon:
			str += "西瓜,"
		case Banana:
			str += "香蕉,"
		case Apple:
			str += "苹果,"
		case Bomb:
			str += "炸弹,"
		}
	}
	fmt.Println(str)
}
