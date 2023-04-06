package fruits

import "fmt"

func (w *WinResult) Init() {
	w.EleValue = make([]int32, 15)
	w.JackPotNum = 0
	w.WinLine = nil
}

// 玛丽游戏
func (w *WinResult) InitMary() {
	w.MaryOutSide = -1
	w.MaryMidArray = nil
	w.MaryOutRate = 0
	w.MaryMidRate = 0
	w.MaryLianXu = 0
}

// 正常游戏 免费游戏
func (w *WinResult) CreateLine(ele [][]int32) {
	w.Init()
	w.result(ele)
	Print(w.EleValue)
}
func (w *WinResult) Win() {
	w.getWinLine()
}

// 玛丽游戏
func (w *WinResult) CreateMary(maryGame [][]int32) {
	w.InitMary()
	w.MaryOutSide = RandSliceInt32IndexByWightN(maryGame[0])
	for i := 0; i < 4; i++ {
		ele := RandSliceInt32IndexByWightN(maryGame[1])
		w.MaryMidArray = append(w.MaryMidArray, ele)
	}
	fmt.Println("外圈元素", w.MaryOutSide)
	fmt.Println("内圈元素", w.MaryMidArray)
	w.MaryWin()
}
func (w *WinResult) MaryWin() {
	var outRate int64
	switch w.MaryOutSide {
	case Watermelon:
		outRate += 200
	case Grape:
		outRate += 100
	case Lemon:
		outRate += 70
	case Cherry:
		outRate += 50
	case Banana:
		outRate += 20
	case Apple:
		outRate += 10
	case Pineapple:
		outRate += 5
	}
	var flag = w.MaryMidArray[0]
	var n int32
	for _, v := range w.MaryMidArray {
		if flag != v {
			break
		}
		n++
	}
	for _, v := range w.MaryMidArray {
		if w.MaryOutSide == v {
			w.MaryOutRate = outRate
			break
		}
	}
	if n >= 3 {
		if n == 3 {
			w.MaryMidRate = 20
		} else if n == 4 {
			w.MaryMidRate = 500
		}
		w.MaryLianXu = n
	}
	//fmt.Println("外圈倍率:", w.MaryOutRate)
	//fmt.Println("内圈倍率", w.MaryMidRate)
}

func (w *WinResult) result(ele [][]int32) {
	n := 0
	for i := 0; i < Column; i++ {
		for j := 0; j < Row; j++ {
			w.EleValue[n] = RandSliceInt32IndexByWightN(ele[j])
			n++
		}
	}
}

func (w *WinResult) getWinLine() {
	n := 0
	var flag int32 = -1
	for k, cols := range LineWinNum {
		flag = w.EleValue[cols[0]]
		//Bonus Scatter 不参与线数 Bonus下班单独计算
		if flag == Bonus || flag == Scatter {
			continue
		}
		var line []int32
		var pos []int32
		for _, key := range cols {
			//不计算 Bonus
			if (flag == w.EleValue[key] || Wild == w.EleValue[key] || flag == Wild) && w.EleValue[key] != Bonus &&
				w.EleValue[key] != Scatter {
				if Wild != w.EleValue[key] {
					flag = w.EleValue[key]
				}
				n++
				line = append(line, w.EleValue[key])
				pos = append(pos, int32(key))
			} else {
				if n >= 3 || (flag == Banana && n >= 2) {
					w.WinLine = append(w.WinLine, WinLine{
						Lines:  line,
						Poss:   pos,
						LineId: k + 1,
						Rate:   GetRate(flag, n),
					})
				}
				n = 0
				pos = nil
				line = nil
				break
			}
			if n == 5 {
				w.WinLine = append(w.WinLine, WinLine{
					Lines:  line,
					Poss:   pos,
					LineId: k + 1,
					Rate:   GetRate(flag, n),
				})
				n = 0
				pos = nil
				line = nil
			}
		}
	}
	//只计算Bonus
	for k, cols := range LineWinNum {
		flag = w.EleValue[cols[0]]
		if flag != Bonus && flag != Scatter {
			continue
		}
		var line []int32
		var pos []int32
		for _, key := range cols {
			if flag == w.EleValue[key] {
				n++
				line = append(line, w.EleValue[key])
				pos = append(pos, int32(key))
			} else {
				if n >= 3 {
					if flag == Scatter {
						w.JackPotNum += n
					}
					w.WinLine = append(w.WinLine, WinLine{
						Lines:  line,
						Poss:   pos,
						LineId: k + 1,
						Rate:   GetRate(flag, n),
					})
				}
				n = 0
				pos = nil
				line = nil
				break
			}
			if n == 5 {
				w.WinLine = append(w.WinLine, WinLine{
					Lines:  line,
					Poss:   pos,
					LineId: k + 1,
					Rate:   GetRate(flag, n),
				})
				n = 0
				pos = nil
				line = nil
			}
		}

	}
	//test code
	if len(w.WinLine) > 0 {
		fmt.Println("====== 赢的总线数 =======", len(w.WinLine))
		for k, v := range w.WinLine {
			fmt.Print(k+1, "  ")
			PrintWin(v.Lines)
			fmt.Println(k+1, "位置 ", v.Poss, "  中奖线号:", v.LineId, " 线元素:", v.Lines, " 倍率:", v.Rate)
		}
	}
}
