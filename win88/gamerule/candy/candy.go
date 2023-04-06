package candy

import (
	"fmt"
	"math/rand"
	"time"
)

// var SpinID int64 = 100000 // todo

type LineData struct {
	Index    int
	Element  int
	Count    int
	Score    int64
	Position []int32
}

//图案*3+连线数量=赔率索引，返回元素值和数量
func isLine(data []int) (e int, count int) {
	tempData := make([]int, LINE_CELL)
	copy(tempData, data)
	dataMap := make(map[int]int, LINE_CELL)

	for _, v := range tempData {
		if _, ok := dataMap[v]; !ok {
			dataMap[v] = 1
		} else {
			dataMap[v]++
		}
	}
	for _, v := range tempData {
		if dataMap[v] >= 3 {
			return v, dataMap[v]
		} else if dataMap[v] >= 2 && (v == Element_QUEKEO || v == Element_RANHMI) {
			return v, dataMap[v]
		}
	}
	return -1, -1
}

// 是否特殊线
func isSpecialLine(data []int) (flag bool, m int) {
	tempData := make([]int, LINE_CELL)
	copy(tempData, data)
	dataMap := make(map[int]int, LINE_CELL)

	for _, v := range tempData {
		if _, ok := dataMap[v]; !ok {
			dataMap[v] = 1
		} else {
			dataMap[v]++
		}
	}
	//if dataMap[Element_VUANGMIEN] == 1 && dataMap[Element_RONG] == 1 && dataMap[Element_NGOC] == 1 {
	//	flag = true
	//	m = LINE_BIGWILD
	//}
	if dataMap[Element_VUANGMIEN] == 1 && dataMap[Element_QUEKEO] == 1 && dataMap[Element_RANHMI] == 1 {
		flag = true
		m = LINE_SMALLWILD
	}
	return
}

func CalcLine(data []int, betLines []int64) (lines []LineData) {
	var isBigWild bool
	var cards = make([]int, len(data))
	copy(cards, data)
	for _, lineNum := range betLines {
		index := int(lineNum)
		lineTemplate := AllLineArray[index-1]
		edata := []int{}
		normalData := []int{}
		realData := []int{}
		epos := []int32{}
		for _, pos := range lineTemplate {
			edata = append(edata, cards[pos])
			if cards[pos] != Element_VUANGMIEN {
				normalData = append(normalData, cards[pos])
			}
			realData = append(realData, data[pos])
			epos = append(epos, int32(pos+1))
		}

		if len(edata) == len(normalData) || len(normalData) == 0 {
			head, count := isLine(edata)
			if head >= 0 {
				lines = append(lines, LineData{index, head, count, LineScore[head][count-1], epos[:count]})
			}
		} else {
			normalData = DelSliceRepEle(normalData)
			if len(normalData) == LINE_CELL-1 {
				if f, m := isSpecialLine(edata); f {
					if m == LINE_SMALLWILD {
						lines = append(lines, LineData{index, Element_Min, LINE_CELL, LineScore[Element_Min][LINE_CELL-1], epos[:LINE_CELL]})
						break
					}
					// 特殊奖励（高倍率 15,20,30,50 多个中线仅计算一次分数）
					if m == LINE_BIGWILD {
						var specialScore int64
						if !isBigWild {
							isBigWild = true
							specialScore = luckyDataRate[rand.Intn(len(luckyDataRate))] * 10 // (特殊倍率 * 10倍，返回时候 / 10 还原，底注配置10，暂不会出现问题)
						}
						lines = append(lines, LineData{index, Element_Min, LINE_CELL, specialScore, epos[:LINE_CELL]})
						continue
					}
				}
			}
			for _, value := range normalData {
				replaceData := []int{}
				for i := 0; i < len(edata); i++ {
					if edata[i] == Element_VUANGMIEN {
						replaceData = append(replaceData, value)
					} else {
						replaceData = append(replaceData, edata[i])
					}
				}
				head, count := isLine(replaceData)
				if head >= 0 {
					lines = append(lines, LineData{index, head, count, LineScore[head][count-1], epos[:count]})
					break
				}
			}
		}
	}
	return
}

//去除切片在的重复的元素
func DelSliceRepEle(data []int) (res []int) {
	if len(data) == 1 {
		return data
	}
	eleFlag := make(map[int]bool)
	for i := 0; i < len(data); i++ {
		if eleFlag[data[i]] {
			continue
		}
		eleFlag[data[i]] = true
		res = append(res, data[i])
	}
	return res
}

func PrintHuman(data []int) {
	var l = len(data)
	if l != ELEMENT_TOTAL {
		return
	}
	for r := 0; r < LINE_ROW; r++ {
		for c := 0; c < LINE_CELL; c++ {
			fmt.Printf("%5s", Element_NAME_MAP[data[r*LINE_CELL+c]])
		}
		fmt.Println()
	}
}

var gSeedV = time.Now().UnixNano()

func GenerateSlotsData_v2(s Symbol) []int {
	gSeedV++
	rand.Seed(gSeedV)
	var slotsData = make([]int, 0, ELEMENT_TOTAL)
	for i := 0; i < ELEMENT_TOTAL; i++ {
		if s == SYMBOL1 {
			slotsData = append(slotsData, symbol1[rand.Intn(len(symbol1))])
		} else if s == SYMBOL2 {
			slotsData = append(slotsData, symbol2[rand.Intn(len(symbol2))])
		}
	}
	return slotsData
}

func GenerateSlotsData_v3() []int {
	return defalutData[rand.Intn(len(defalutData))]
}
