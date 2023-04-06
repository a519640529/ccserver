package tamquoc

import (
	"fmt"
	"math/rand"
	"time"
)

var SpinID int64 = 100000 // todo

// 小游戏点击次数(起始次数)
var bonusGameNum = 10

type LineData struct {
	Index           int
	Element         int
	Count           int
	Score           int
	Position        []int32
	FreeElementCnt  int // 免费元素个数
	BonusElementCnt int // Bonus元素个数
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
			if dataMap[v] == 3 && v == Element_VUKHI { // 跨服3线不计算
				continue
			}
			return v, dataMap[v]
		}
	}
	return -1, -1
}

func CalcLine(data []int, betLines []int64) (lines []LineData) {
	for _, lineNum := range betLines {
		index := int(lineNum)
		lineTemplate := AllLineArray[index-1]
		edata := []int{}
		epos := []int32{}
		for _, pos := range lineTemplate {
			edata = append(edata, data[pos])
		}
		if len(edata) == LINE_CELL {
			head, count := isLine(edata)
			for _, pos := range lineTemplate {
				if data[pos] == head {
					epos = append(epos, int32(pos+1))
				}
			}
			if head >= 0 && count >= 3 && len(epos) == count {
				var freeElementCnt, bonusElementCnt int // 免费元素个数/Bonus元素个数
				if head == Element_FREESPIN {
					freeElementCnt = count
				}
				if head == Element_BONUS && count >= 3 {
					bonusElementCnt = count
				}
				lines = append(lines, LineData{index, head, count, LineScore[head][count-1], epos, freeElementCnt, bonusElementCnt})
			}
		}
	}
	return
}

type CaclResult struct {
	WinLines      [][]LineData //多次中奖的线
	IsJackpot     bool         //是否爆奖
	AllScore      int          //总中奖倍率
	BonusScore    int          //小游戏奖励
	SpinFreeTimes int          //免费旋转次数
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

type Symbol int

const (
	SYMBOL1 Symbol = iota + 1
	SYMBOL2
)
const (
	DEFALUTROOMMODEL int = iota
	ROOMMODE1
	ROOMMODE2
	ROOMMODE3
)

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
func GenerateSlotsData_v3(roomMode int) []int {
	gSeedV++
	rand.Seed(gSeedV)
	var slotsData = make([]int, 0)
	switch roomMode {
	case ROOMMODE1:
		slotsData = roomMode1[rand.Intn(len(roomMode1))]
	case ROOMMODE2:
		slotsData = roomMode2[rand.Intn(len(roomMode2))]
	case ROOMMODE3:
		slotsData = roomMode3[rand.Intn(len(roomMode3))]
	case DEFALUTROOMMODEL:
		slotsData = defaultRoomModel[rand.Intn(len(defaultRoomModel))]
	}
	return slotsData
}

type BonusGameResult struct {
	BonusData       []int64 //每次点击的显示奖金，有几个就点击几次，最后一次点击是0
	DataMultiplier  int64   //第一级界面小游戏总和
	Mutiplier       int     //最终小游戏的倍率
	TotalPrizeValue int64   //最终小游戏的奖励
}

var gSeedV = time.Now().UnixNano()

func GenerateBonusGame(betValue int, startBonus int) BonusGameResult {
	if betValue <= 0 || startBonus <= 0 {
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
	for i := 0; i < bonusGameNum; i++ {
		var rate int
		rate = bonusNormalRate[rand.Intn(len(bonusNormalRate))]
		if rate == -1 {
			i--
		}
		switch rate {
		case -1: // 未中奖(客户端显示图片)
			rate = 0
		case 1: // 中高倍率
			rate = bonusHighRate[rand.Intn(len(bonusHighRate))]
		default: // 中低倍率
		}
		prizeValue := int64(rate) * int64(betValue)
		bg.TotalPrizeValue += prizeValue
		bg.DataMultiplier += prizeValue
		bg.BonusData = append(bg.BonusData, prizeValue)
		if bonusGameNum >= 24 {
			break
		}
	}
	bg.Mutiplier = startBonus
	bg.TotalPrizeValue *= int64(bg.Mutiplier)
	return bg
}
