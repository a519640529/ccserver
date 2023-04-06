package richblessed

import (
	"fmt"
)

func (w *WinResult) Init() {
	w.EleValue = []int32{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
		-1, -1, -1, -1, -1} // make([]int32, 15)
	// w.JackPotNum = 0
	w.WinLine = nil
	w.IsHaveScatter = false
}

// 正常游戏 免费游戏
func (w *WinResult) CreateLine(ele [][]int32) {
	w.Init()
	w.result(ele)
	Print(w.EleValue)
}

// JACKPOT游戏
func (w *WinResult) InitJACKPOT() {
	w.JackpotEle = -1
}

// JACKPOT游戏
func (w *WinResult) CreateJACKPOT(ele []int32) int32 {
	w.InitJACKPOT()
	w.JackpotEle = RandSliceInt32IndexByWightN(ele)
	return w.JackpotEle
}

func (w *WinResult) CanJACKPOT(bet int64, big int64) (ret bool) {
	if w.IsHaveScatter {
		ret = RandJACKPOT(bet, big)
	}
	return
}

// 奖池中奖 (当前下注  最大下注) 返回中的奖池索引

func (w *WinResult) WinJackPot(bet, maxBet int64) int {
	if w.CanJACKPOT(bet, maxBet) {
		idx := RandSliceInt32IndexByWightN(JKWeight)
		return int(idx)
	}
	return -1
}
func (w *WinResult) Win(bet, maxBet int64) {
	w.getWinLineAndFree()

	w.WinJackPot(bet, maxBet)
}

func (w *WinResult) JACKPOTWin() {
	w.JackpotRate = JkEleNumRate[int((w.JackpotEle))]
}

func (w *WinResult) resultele(eles [][]int32) {

	for n, ele := range eles { // 每行元素有自己的概率

		w.EleValue[n] = RandSliceInt32IndexByWightN(ele)

	}
}

func (w *WinResult) result(ele [][]int32) {
	n := 0
	for j := 0; j < Row; j++ {
		gongnum := 0
		for i := 0; i < Column; i++ {
			w.EleValue[n] = RandSliceInt32IndexByWightN(ele[j])
			if w.EleValue[n] == Scatter && (j == 0 || j == Row-1) { // 最左或者最右不能为万能牌

				for ra := 0; ra != 100; ra++ {
					w.EleValue[n] = RandSliceInt32IndexByWightN(ele[j])
					if w.EleValue[n] != Scatter {
						break
					}
				}
			}
			if w.EleValue[n] == Scatter || w.EleValue[n] == Gongs {
				gongnum++
				if gongnum > 1 { // 一列只能有一个铜锣或万能牌
					for ra := 0; ra != 100; ra++ { // 100次换不掉就是给的概率有问题
						w.EleValue[n] = RandSliceInt32IndexByWightN(ele[j])
						if w.EleValue[n] != Scatter && w.EleValue[n] != Gongs {
							break
						}
					}
				}
			}
			if w.EleValue[n] == Scatter {
				w.IsHaveScatter = true
			}
			n++
		}
	}
}

// 0  1  2  3  4
// 5  6  7  8  9
// 10 11 12 13 14
func (w *WinResult) getWinLine() {
	var ele [][]int
	var count int
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			for m := 0; m < 3; m++ {
				for n := 0; n < 3; n++ {
					for x := 0; x < 3; x++ {
						a1, a2, a3, a4, a5 := i*5, j*5+1, m*5+2, n*5+3, x*5+4
						el := []int{a1, a2, a3, a4, a5}
						//fmt.Println(el)
						ele = append(ele, el)
						count++

						var line []int32
						var pos []int32
						var num int
						var flag = w.EleValue[a1]
						for _, key := range el {
							if flag == w.EleValue[key] {
								line = append(line, w.EleValue[key])
								pos = append(pos, int32(key))
								num++
							} else {
								if num >= 3 {
									w.WinLine = append(w.WinLine, WinLine{
										Lines:  line,
										Poss:   pos,
										LineId: count,
										Rate:   GetRate(flag, num),
									})
								}
								break
							}
						}
					}
				}
			}
		}
	}
	fmt.Println("lel:", len(ele))
}

/*
func (w *WinResult) getWinEleAndFree() {
	var winret [15]int32
	var winnum [15]int32
	for i := 0; i != Column; i++ { //横
		winnum[w.EleValue[i*Row]]++
	}
	for j := 1; j != Row; j++ {
		ele1, ele2, ele3 := int32(0), int32(0), int32(0)
		ele1 = w.EleValue[j]
		ele2 = w.EleValue[1*Row+j]
		ele3 = w.EleValue[2*Row+j]
		for i := 0; i != Column; i++ { //横
			winnum[w.EleValue[i*Row]]++
		}

	}
}*/

func (w *WinResult) getWinLineAndFree() {
	// var weles []WinLine
	Print(w.EleValue)
	for i := 0; i != Column; i++ { //横
		var wele []WinLine
		for k := 0; k != Column; k++ {
			//wel
			wel := WinLine{
				Lines: []int32{w.EleValue[i*Row]}, // 元素
				Poss:  []int32{int32(i * Row)},    // 位置
			}
			if w.EleValue[i*Row] == w.EleValue[k*Row+1] || w.EleValue[k*Row+1] == Scatter {

				wel.Lines = append(wel.Lines, w.EleValue[k*Row+1])
				wel.Poss = append(wel.Poss, int32(k*Row+1))
				wele = append(wele, wel)
			}

		}
		if len(wele) == 0 {
			continue
		}
		for j := 2; j != Row; j++ {

			del := append([]WinLine{}, wele...)
			wele = wele[:0] // 长线代替短线
			for k := 0; k != Column; k++ {
				if w.EleValue[i*Row] == w.EleValue[k*Row+j] || w.EleValue[k*Row+j] == Scatter {

					for _, wel := range del {
						var newwel WinLine
						newwel.Lines = append(newwel.Lines, wel.Lines...)
						newwel.Poss = append(newwel.Poss, wel.Poss...)
						newwel.Lines = append(newwel.Lines, w.EleValue[k*Row+j])
						newwel.Poss = append(newwel.Poss, int32(k*Row+j))
						newwel.Rate = GetRate(newwel.Lines[0], len(newwel.Lines))
						wele = append(wele, newwel)
						// fmt.Println("index: ", i*Row, w.EleValue[i*Row], w.EleValue[k*Row+j], wel, newwel, k*Row+j, k)
					}

				}
			}

			if len(wele) == 0 {
				if j > 2 {
					w.WinLine = append(w.WinLine, del...)
				}
				break
			} else if j == Row-1 {
				w.WinLine = append(w.WinLine, wele...)
			}

		}
		// weles = append(weles, wele...)
	}
	for _, v := range w.WinLine {
		w.AllRate += v.Rate
	}
	fs := 0
	for j := 0; j < Row; j++ { // 前3列就可判断
		for i := 0; i < Column; i++ {
			if w.EleValue[i*Row+j] == Gongs || w.EleValue[i*Row+j] == Scatter {
				fs++
				// fmt.Println("位置: ", j, i, fs, i*Row+j, w.EleValue[i*Row+j])
				break
			}
		}
		if fs != j+1 {
			break
		}
		if fs == 3 {
			w.FreeNum += 10 //10
			break
		}
	}

	//fmt.Println("fs: ", fs)
}
