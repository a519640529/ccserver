package iceage

import (
	"fmt"
	"math/rand"
	"sort"
	"time"
)

type LineData struct {
	Index   int
	Element int
	Count   int
	Score   int
}

var SpinID int64 = 100000 // todo

//图案*3+连线数量=赔率索引，返回元素值和数量
func IsLine(data []int) (e int, count int) {
	e = data[0]
	count = 1
	for i := 1; i < len(data); i++ {
		if data[i] == e {
			count++
		} else {
			break
		}
	}
	if (e != Element_NUT && count < 3) || (e == Element_NUT && count < 4) {
		return -1, -1
	}
	return
}
func CalcLine(data []int, betLines []int64) (lines []LineData) {
	//fmt.Println("************************************************")
	//fmt.Println("data:",data)
	for _, lineNum := range betLines {
		index := int(lineNum)
		lineTemplate := AllLineArray[index-1]
		edata := []int{}
		normalData := []int{}
		for _, pos := range lineTemplate {
			edata = append(edata, data[pos])
			normalData = append(normalData, data[pos])
		}
		//fmt.Println("edata:",edata)
		//fmt.Println("normalData:",normalData)
		if len(edata) == len(normalData) {
			head, count := IsLine(edata)
			if head >= 0 {
				lines = append(lines, LineData{index, head, count, LineScore[head][count-1]})
			}
		} else {
			//fmt.Printf("Line:%v-%v\n", index, lineTemplate)
			//fmt.Println("edata:", edata)
			//fmt.Println("normalData:", normalData)
			normalData = DelSliceRepEle(normalData)
			//fmt.Println("normalData:", normalData)
			for _, value := range normalData {
				_ = value
				replaceData := []int{}
				for i := 0; i < len(edata); i++ {
					replaceData = append(replaceData, edata[i])
				}
				head, count := IsLine(replaceData)
				if head >= 0 {
					lines = append(lines, LineData{index, head, count, 0})
					//fmt.Printf("*******************%v****************************\n", lines)
					break
				}
			}
		}
	}
	//fmt.Print("lines:",lines)
	return
}

//去除切片在的重复的元素
func DelSliceRepEle(data []int) []int {
	if len(data) == 1 {
		return data
	}
	sort.Ints(data)
	for i := 0; i < len(data)-1; {
		if data[i] == data[i+1] {
			data = append(data[:i], data[i+1:]...)
		} else {
			i++
		}
	}
	return data
}

var gSeedV = time.Now().UnixNano()

//生成消消乐之后的数据，distributionData为将要注入的数据源
func MakePlan(cards []int, distributionData []int32, betLines []int64) (newCards []int) {
	var c = make([]int, len(cards))
	for i, card := range cards {
		c[i] = card
	}
	var lines = CalcLine(cards, betLines)
	//fmt.Printf("input card:%v \n",cards)
	//PrintHuman(cards)
	//fmt.Printf("calc result:%v \n",lines)
	if len(lines) == 0 {
		return nil
	}
	for _, line := range lines {
		//if line.Score > 0 {
		var l = AllLineArray[line.Index-1]
		for pos := 0; pos < line.Count; pos++ {
			c[l[pos]] = -1
		}
		//}
	}
	//fmt.Printf("after trim card:%v \n",c)
	//PrintHuman(c)
	for i := 10; i < 15; i++ {
		if c[i] == -1 && c[i-5] != -1 {
			c[i], c[i-5] = c[i-5], c[i]
		}
	}
	for i := 5; i < 10; i++ {
		if c[i] == -1 && c[i-5] != -1 {
			c[i], c[i-5] = c[i-5], c[i]
		}
	}
	for i := 10; i < 15; i++ {
		if c[i] == -1 && c[i-5] != -1 {
			c[i], c[i-5] = c[i-5], c[i]
		}
	}
	//fmt.Printf("after plan card:%v \n",c)
	//PrintHuman(c)
	var dD = make([]int, 0)
	for _, e := range distributionData {
		if int(e) == Element_BONUS || int(e) == Element_FREESPIN || int(e) == Element_JACKPOT {
			continue
		}
		dD = append(dD, int(e))
	}
	gSeedV++
	rand.Seed(gSeedV)
	for i, card := range c {
		if card == -1 {
			c[i] = dD[rand.Intn(len(dD))]
		}
	}
	//fmt.Printf("after fill card:%v \n",c)
	//PrintHuman(c)
	return c
}

type CaclResult struct {
	WinLines      []LineData //中奖的线
	AllScore      int        //总中奖倍率
	BonusScore    int        //小游戏奖励
	SpinFreeTimes int        //免费旋转次数
}

func CaclScore(cards []int, betLines []int64) (int32, int32, int32, int32, bool) {
	curscore := 0
	alllinenum := 0 //中奖线路总数
	allscore := 0
	spinFree := 0
	isJackpot := false
	lines := CalcLine(cards, betLines)
	for _, line := range lines {
		if line.Element == Element_FREESPIN {
			spinFree += FreeSpinTimesRate[line.Count-1]
			continue
		}
		if line.Element == Element_JACKPOT && line.Count == LINE_CELL {
			isJackpot = true
		}
		curscore = LineScore[line.Element][line.Count-1]
		allscore = allscore + curscore
		alllinenum++
	}
	bounsCount := 0
	for _, card := range cards {
		if card == Element_BONUS {
			bounsCount++
		}
	}
	rand.Seed(time.Now().UnixNano())
	//bounsScore:= SmallGameByBouns[bounsCount][0]+rand.Intn(SmallGameByBouns[bounsCount][1])
	return int32(alllinenum), int32(allscore), int32(0), int32(spinFree), isJackpot
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
