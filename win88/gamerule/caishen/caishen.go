package caishen

import (
	"fmt"
	"math/rand"
	"time"
)

var SpinID int64 = 100000 // todo

type LineData struct {
	Index    int
	Element  int
	Count    int
	Score    int
	Position []int32
}

//图案*3+连线数量=赔率索引，返回元素值和数量
func isLine(data []int) (e int, count int) {
	e = data[0]
	count = 1
	for i := 1; i < len(data); i++ {
		if data[i] == e {
			count++
		} else {
			break
		}
	}
	if count < 3 {
		return -1, -1
	}
	return
}
func CalcLine(data []int, betLines []int64) (lines []LineData) {
	var cards = make([]int, len(data))
	for i, v := range data {
		cards[i] = v
	}
	for i := 10; i < 15; i++ {
		if cards[i] == Element_WILD || cards[i-5] == Element_WILD || cards[i-10] == Element_WILD {
			cards[i] = Element_WILD
			cards[i-5] = Element_WILD
			cards[i-10] = Element_WILD
		}
	}
	for _, lineNum := range betLines {
		index := int(lineNum)
		lineTemplate := AllLineArray[index-1]
		edata := []int{}
		normalData := []int{}
		realData := []int{}
		epos := []int32{}
		if cards[lineTemplate[0]] == Element_SCATTER || cards[lineTemplate[0]] == Element_BONUS {
			continue
		}
		for _, pos := range lineTemplate {
			edata = append(edata, cards[pos])
			if cards[pos] != Element_WILD {
				normalData = append(normalData, cards[pos])
			}
			realData = append(realData, data[pos])
			epos = append(epos, int32(pos+1))
		}

		if len(edata) == len(normalData) {
			head, count := isLine(edata)
			if head >= 0 {
				lines = append(lines, LineData{index, head, count, LineScore[head][count-1], epos[:count]})
			}
		} else {
			normalData = DelSliceRepEle(normalData)
			for _, value := range normalData {
				replaceData := []int{}
				for i := 0; i < len(edata); i++ {
					if realData[0] == Element_JACKPOT { // WILD 不能替换 JACKPOT
						replaceData = append(replaceData, realData[i])
					} else {
						if edata[i] == Element_WILD {
							replaceData = append(replaceData, value)
						} else {
							replaceData = append(replaceData, edata[i])
						}
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
	for i := 0; i < len(data)-1; i++ {
		if eleFlag[data[i]] {
			continue
		}
		eleFlag[data[i]] = true
		res = append(res, data[i])
	}
	return res
}

type CaclResult struct {
	WinLines      [][]LineData //多次中奖的线
	IsJackpot     bool         //是否爆奖
	AllScore      int          //总中奖倍率
	BonusScore    int          //小游戏奖励
	SpinFreeTimes int          //免费旋转次数
}

func CaclScore(cards []int, betLines []int64) (int32, int32, int32, int32, bool) {
	curscore := 0
	alllinenum := 0 //中奖线路总数
	allscore := 0
	spinFree := 0
	isJackpot := false
	lines := CalcLine(cards, betLines)
	for _, line := range lines {
		if line.Element == Element_SCATTER {
			spinFree += FreeSpinTimesRate[line.Count]
			continue
		}
		if line.Element == Element_JACKPOT && line.Count == LINE_CELL {
			isJackpot = true
		}
		curscore = LineScore[line.Element][line.Count-1]
		line.Score = curscore
		allscore = allscore + curscore
		alllinenum++
	}
	bounsCount := 0
	for _, card := range cards {
		if card == Element_BONUS {
			bounsCount++
		}
		if card == Element_JACKPOT {
			spinFree++
		}
	}
	rand.Seed(time.Now().UnixNano())
	//bounsScore:= SmallGameByBouns[bounsCount][0]+rand.Intn(SmallGameByBouns[bounsCount][1])
	spinFreeTimes := FreeSpinTimesRate[spinFree]
	return int32(alllinenum), int32(allscore), int32(0), int32(spinFreeTimes), isJackpot
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

//prizeFund > limit * roomId走这个检查
func CheckBigWin(slotsData []int) bool {
	if slotsData[0] == Element_WILD || slotsData[5] == Element_WILD || slotsData[10] == Element_WILD || slotsData[4] == Element_WILD || slotsData[9] == Element_WILD || slotsData[14] == Element_WILD {
		return true
	}
	return false
}

//prizeFund < limit * roomId走这个更严格的检查
func CheckSuperWin(slotsData []int) bool {
	if slotsData[0] == Element_WILD || slotsData[5] == Element_WILD || slotsData[10] == Element_WILD || slotsData[4] == Element_WILD || slotsData[9] == Element_WILD || slotsData[14] == Element_WILD {
		return true
	}
	if (slotsData[1] == Element_WILD || slotsData[6] == Element_WILD || slotsData[11] == Element_WILD) && (slotsData[2] == Element_WILD || slotsData[7] == Element_WILD || slotsData[12] == Element_WILD) {
		return true
	}
	if (slotsData[2] == Element_WILD || slotsData[7] == Element_WILD || slotsData[12] == Element_WILD) && (slotsData[3] == Element_WILD || slotsData[8] == Element_WILD || slotsData[13] == Element_WILD) {
		return true
	}
	return false
}

type Symbol int

const (
	SYMBOL1 Symbol = iota + 1
	SYMBOL2
)

func GenerateSlotsData_v2(s Symbol) ([]int, int) {
	var tryCount = 0
Next:
	gSeedV++
	rand.Seed(gSeedV)
	var slotsData = make([]int, 0, ELEMENT_TOTAL)
	for i := 0; i < ELEMENT_TOTAL; i++ {
		if s == SYMBOL1 {
			if i == 0 || i == 5 || i == 10 {
				slotsData = append(slotsData, symbol1[rand.Intn(len(symbol1)-3)+3])
			} else {
				slotsData = append(slotsData, symbol1[rand.Intn(len(symbol1))])
			}
		} else if s == SYMBOL2 {
			if i == 0 || i == 5 || i == 10 {
				slotsData = append(slotsData, symbol2[rand.Intn(len(symbol2)-5)+5])
			} else {
				slotsData = append(slotsData, symbol2[rand.Intn(len(symbol2))])
			}
		}
	}
	tryCount++
	if (s == SYMBOL1 && CheckSuperWin(slotsData)) || (s == SYMBOL2 && CheckBigWin(slotsData)) {
		goto Next
	}
	return slotsData, tryCount
}

type BonusGameResult struct {
	BonusData       []int64 //每次点击的显示奖金，有几个就点击几次，最后一次点击是0
	DataMultiplier  int64   //第一级界面小游戏奖金总和
	Mutiplier       int     //最终小游戏的倍率
	TotalPrizeValue int64   //最终小游戏的奖励
}

var gSeedV = time.Now().UnixNano()

func GenerateBonusGame(totalBet int, startBonus int) BonusGameResult {
	if totalBet <= 0 || startBonus <= 0 {
		return BonusGameResult{
			BonusData:       nil,
			DataMultiplier:  0,
			Mutiplier:       0,
			TotalPrizeValue: 0,
		}
	}
	var bg = BonusGameResult{
		BonusData:       make([]int64, 0),
		DataMultiplier:  0,
		Mutiplier:       0,
		TotalPrizeValue: int64(0),
	}
	gSeedV++
	rand.Seed(gSeedV)
	for _, e := range BonusStepArr {
		rnd := rand.Intn(len(e))
		prizeValue := int64(e[rnd]*float64(totalBet) + 0.00001)
		bg.TotalPrizeValue += prizeValue
		bg.DataMultiplier += prizeValue
		bg.BonusData = append(bg.BonusData, prizeValue)
		if prizeValue == 0 {
			break
		}
	}
	bg.Mutiplier = startBonus + rand.Intn(3)
	bg.TotalPrizeValue *= int64(bg.Mutiplier)
	return bg
}
