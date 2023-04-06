package fortunezhishen

import (
	"math/rand"
)

func (wls *WinResult) Init(gameState int) {
	wls.Cards = make([]int32, 15)
	wls.WinLine = nil
	wls.MakeAFortuneNum = 0
	wls.GemstoneNum = 0
	wls.NewAddGemstone = 0
	if wls.GemstoneRate == nil || (gameState != StopAndRotate && gameState != StopAndRotate2) {
		wls.MidIcon = -1
		wls.GemstoneRate = make([]int64, 15)
		wls.LastRes = nil
	}
}

func (wls *WinResult) CreateLine(gameState int, eleLineAppearRate []int32) {
	wls.Init(gameState)
	//生成元素
	wls.result(gameState, eleLineAppearRate)

}
func (wls *WinResult) Win(betIdx int, gameState int) {
	if gameState == StopAndRotate || gameState == StopAndRotate2 {
		wls.copyRes()
	}
	if gameState != StopAndRotate && gameState != StopAndRotate2 {
		//生成输赢结果
		wls.getWinLine()
	}
	//查询宝石
	wls.initData(betIdx)
}

func (wls *WinResult) copyRes() {
	if len(wls.LastRes) > 0 {
		for k, v := range wls.LastRes {
			if v == Gemstone {
				wls.Cards[k] = v
			}
		}
	}
}
func (wls *WinResult) initData(betIdx int) {
	var mid = []int{1, 2, 3, 6, 7, 8, 11, 12, 13}
	if wls.MidIcon == Gemstone {
		rate := wls.GetGemstoneRate(betIdx, true)
		for _, id := range mid {
			if wls.GemstoneRate[id] == 0 {
				wls.GemstoneRate[id] = rate
			} else {
				break
			}
		}
	}
	for k, v := range wls.Cards {
		if v == Gemstone {
			wls.GemstoneNum++
			if wls.GemstoneRate[k] == 0 {
				wls.NewAddGemstone++
				wls.GemstoneRate[k] = wls.GetGemstoneRate(betIdx, false)
			}
		} else if v == MakeAFortune {
			wls.MakeAFortuneNum++
		}
	}
}

//根据概率 获取宝石倍数
func (wls *WinResult) GetGemstoneRate(betIdx int, big bool) int64 {
	grandPrize, bigPrize := true, true
	if betIdx == 0 || betIdx == 1 || betIdx == 3 {
		bigPrize = false
	}
	if betIdx == 0 || betIdx == 1 || betIdx == 2 || betIdx == 3 {
		grandPrize = false
	}

	nowGemstoneRatePrize := make([]int, len(SmallGemstoneRatePrize))
	if !big {
		copy(nowGemstoneRatePrize, SmallGemstoneRatePrize)
	} else {
		copy(nowGemstoneRatePrize, BigGemstoneRatePrize)
		bigPrize = true
	}
	if !grandPrize {
		nowGemstoneRatePrize[0] = 0
	}
	if !bigPrize {
		nowGemstoneRatePrize[1] = 0
	}

	r := randSliceIndexByWightN(nowGemstoneRatePrize)
	rate := rand.Int63n(4)
	switch r {
	case Bet1_4:
		rate += 1
	case Bet5_8:
		rate += 5
	case Bet9_12:
		rate += 9
	case Bet13_16:
		rate += 13
	case Bet17_19:
		rate = rand.Int63n(3) + 17
	case BetGrandPrize:
		rate = GrandPrize
	case BetBigPrize:
		rate = BigPrize
	case BetMidPrize:
		rate = MidPrize
	case BetSmallPrize:
		rate = SmallPrize
	}
	return rate
}

//获取元素值
func (wls *WinResult) result(gameState int, eleLineAppearRate []int32) {
	var LineRate []int32 //正常元素概率(小宝石) 或者 正常元素概率2(大宝石)
	var noWild []int32   //财神元素概率为0(第一列使用)

	eleLen := len(eleLineAppearRate)
	if gameState == Normal || gameState == StopAndRotate {
		LineRate = make([]int32, eleLen-1)
		copy(LineRate, eleLineAppearRate[:eleLen-1])
	} else {
		LineRate = make([]int32, eleLen-2)
		copy(LineRate, eleLineAppearRate[:eleLen-2])
		LineRate = append(LineRate, eleLineAppearRate[eleLen-1])
	}
	//财神元素概率为0
	noWild = make([]int32, eleLen-1)
	copy(noWild, LineRate)
	noWild[Wild] = 0
	//没有发财
	noMakeAFortune := make([]int32, len(LineRate))
	copy(noMakeAFortune, LineRate)
	noMakeAFortune[MakeAFortune] = 0

	//fmt.Println("元素概率			     ", LineRate)
	//fmt.Println("财神元素概率为0		 ", noWild)

	n := 0
	var haveMakeAFortune = make([]int, 5)
	for i := 0; i < Column; i++ {
		for j := 0; j < Row; j++ {
			if haveMakeAFortune[j] == 0 {
				if j == 0 {
					wls.Cards[n] = RandSliceInt32IndexByWightN(noWild)
				} else {
					wls.Cards[n] = RandSliceInt32IndexByWightN(LineRate)
				}
				if wls.Cards[n] == MakeAFortune {
					haveMakeAFortune[j] = 1
				}
			} else {
				if j == 0 {
					wls.Cards[n] = RandSliceInt32IndexByWightN(noWild)
				} else {
					wls.Cards[n] = RandSliceInt32IndexByWightN(noMakeAFortune)
				}
			}
			n++
		}
	}
	if gameState == FreeGame {
		r := RandSliceInt32IndexByWightN(LineRate)
		wls.MidIcon = r
		var mid = []int{1, 2, 3, 6, 7, 8, 11, 12, 13}
		for _, key := range mid {
			wls.Cards[key] = wls.MidIcon
		}
	}
}

//获取结果
func (wls *WinResult) getWinLine() {
	n := 0
	var flag int32 = -1
	for k, cols := range LineWinNum {
		flag = wls.Cards[cols[0]]
		//宝石 发财不参与线数
		if flag == Gemstone || flag == MakeAFortune {
			continue
		}
		var line []int32
		var pos []int32
		for _, key := range cols {
			if flag == wls.Cards[key] || (Wild == wls.Cards[key] && k != 0) {
				n++
				line = append(line, wls.Cards[key])
				pos = append(pos, int32(key))
			} else {
				if n >= 3 || (flag == Firecrackers && n >= 2) {
					wls.WinLine = append(wls.WinLine, WinLine{
						Lines:  line,
						Poss:   pos,
						LineId: k + 1,
						Rate:   GetRate(flag, len(line)),
					})
				}
				n = 0
				pos = nil
				line = nil
				break
			}
			if n == 5 {
				wls.WinLine = append(wls.WinLine, WinLine{
					Lines:  line,
					Poss:   pos,
					LineId: k + 1,
					Rate:   GetRate(flag, len(line)),
				})
				n = 0
				pos = nil
				line = nil
			}
		}
	}
	//test code
	//if len(wls.WinLine) > 0 {
	//	for k, v := range wls.WinLine {
	//		fmt.Println("=============")
	//		fmt.Print(k, "  ")
	//		PrintWin(v.Lines)
	//		fmt.Println(k, "位置 ", v.Poss, "  中奖线号:", v.LineId)
	//	}
	//	fmt.Println("=============")
	//}
}
func GetRate(ele int32, num int) int64 {
	if data, ok := EleNumRate[ele]; ok {
		if r, ok2 := data[num]; ok2 {
			return r
		}
	}
	return 0
}
