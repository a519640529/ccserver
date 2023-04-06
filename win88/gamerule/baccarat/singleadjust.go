package baccarat

import (
	"math/rand"
)

//开和
func (this *Poker) GetTIE() (cards []int32) {
	for k := 0; k < 20; k++ {
		banker := [3]int32{-1, -1, -1}
		xian := [3]int32{-1, -1, -1}
		for i := 0; i < 2; i++ {
			c, _ := this.Next()
			banker[i] = c
		}
		if isPair(banker) {
			this.PutIn(banker[:])
			continue
		}
		xc, _ := this.Next()
		xian[0] = xc
		bankNum := num(banker)
		c := int(xian[0]%13 + 1)
		if c >= 10 {
			c = 0
		}
		nc := this.FindCard(int32(bankNum - c))
		if nc == -1 {
			this.PutIn(banker[:])
			this.PutIn(xian[:])
			continue
		}
		xian[1] = nc
		if isPair(xian) {
			this.PutIn(banker[:])
			this.PutIn(xian[:])
		} else {
			cards = append(cards, xian[:]...)
			cards = append(cards, banker[:]...)
			return
		}
	}
	return nil
}

//开庄赢 没有对子
func (this *Poker) GetBankerWin() (cards []int32) {
	banker := [3]int32{-1, -1, -1}
	xian := [3]int32{-1, -1, -1}
	for k := 0; k < 20; k++ {
		for i := 0; i < 2; i++ {
			c, _ := this.Next()
			banker[i] = c
		}
		if isPair(banker) {
			this.PutIn(banker[:])
			continue
		}
		for i := 0; i < 2; i++ {
			c, _ := this.Next()
			xian[i] = c
		}
		if isPair(xian) {
			this.PutIn(xian[:])
			continue
		}
		bankerNum := num(banker)
		xianNum := num(xian)
		if bankerNum > xianNum {
			cards = append(cards, xian[:]...)
			cards = append(cards, banker[:]...)
			return
		} else if bankerNum < xianNum {
			cards = append(cards, banker[:]...)
			cards = append(cards, xian[:]...)
			return
		} else {
			this.PutIn(banker[:])
			this.PutIn(xian[:])
			banker = [3]int32{-1, -1, -1}
			xian = [3]int32{-1, -1, -1}
		}
	}
	return nil
}

//开闲赢 没有对子
func (this *Poker) GetXianWin() (cards []int32) {
	banker := [3]int32{-1, -1, -1}
	xian := [3]int32{-1, -1, -1}
	for k := 0; k < 20; k++ {
		for i := 0; i < 2; i++ {
			c, _ := this.Next()
			banker[i] = c
		}
		if isPair(banker) {
			this.PutIn(banker[:])
			continue
		}
		for i := 0; i < 2; i++ {
			c, _ := this.Next()
			xian[i] = c
		}
		if isPair(xian) {
			this.PutIn(xian[:])
			continue
		}
		bankerNum := num(banker)
		xianNum := num(xian)
		if bankerNum > xianNum {
			cards = append(cards, banker[:]...)
			cards = append(cards, xian[:]...)
			return
		} else if bankerNum < xianNum {
			cards = append(cards, xian[:]...)
			cards = append(cards, banker[:]...)
			return
		} else {
			this.PutIn(banker[:])
			this.PutIn(xian[:])
			banker = [3]int32{-1, -1, -1}
			xian = [3]int32{-1, -1, -1}
		}
	}
	return nil
}

//开庄赢 带庄对
func (this *Poker) GetBankerAndBankerPair() (cards []int32) {
	banker := [3]int32{-1, -1, -1}
	irand := this.GetRandPair()
	if irand == -1 {
		return nil
	}
	buf, npair := delSclice(this.buf, []int32{irand, irand})
	if len(npair) != 2 {
		return nil
	}
	this.buf = buf
	banker = [3]int32{npair[0], npair[1], -1}
	bankerNum := num(banker)
	for i := 0; i < 20; i++ {
		xian := [3]int32{-1, -1, -1}
		for k := 0; k < 2; k++ {
			c, _ := this.Next()
			xian[k] = c
		}
		if isPair(xian) {
			this.PutIn(xian[:])
			continue
		}
		xianNum := num(xian)
		if bankerNum > xianNum {
			cards = append(cards, xian[:]...)
			cards = append(cards, banker[:]...)
			return
		} else {
			this.PutIn(xian[:])
		}
	}
	this.PutIn(banker[:])
	return nil
}

//开庄赢 带闲对
func (this *Poker) GetBankerAndXianPair() (cards []int32) {
	xian := [3]int32{-1, -1, -1}
	irand := this.GetRandPair()
	if irand == -1 {
		return nil
	}
	buf, npair := delSclice(this.buf, []int32{irand, irand})
	if len(npair) != 2 {
		return nil
	}
	this.buf = buf
	xian = [3]int32{npair[0], npair[1], -1}

	xianNum := num(xian)
	for i := 0; i < 20; i++ {
		banker := [3]int32{-1, -1, -1}
		for k := 0; k < 2; k++ {
			c, _ := this.Next()
			banker[k] = c
		}
		if isPair(banker) {
			this.PutIn(banker[:])
			continue
		}
		bankerNum := num(banker)
		if bankerNum > xianNum {
			cards = append(cards, xian[:]...)
			cards = append(cards, banker[:]...)
			return
		} else {
			this.PutIn(banker[:])
		}
	}
	this.PutIn(xian[:])
	return nil
}

//开闲赢 带闲对
func (this *Poker) GetXianAndXianPair() (cards []int32) {
	xian := [3]int32{-1, -1, -1}
	imax := this.GetRandPair()
	if imax == -1 {
		return nil
	}
	buf, npair := delSclice(this.buf, []int32{imax, imax})
	if len(npair) != 2 {
		return nil
	}
	this.buf = buf
	xian = [3]int32{npair[0], npair[1], -1}
	xianNum := num(xian)
	for i := 0; i < 20; i++ {
		banker := [3]int32{-1, -1, -1}
		for k := 0; k < 2; k++ {
			c, _ := this.Next()
			banker[k] = c
		}
		if isPair(banker) {
			this.PutIn(banker[:])
			continue
		}
		bankerNum := num(banker)
		if bankerNum < xianNum {
			cards = append(cards, xian[:]...)
			cards = append(cards, banker[:]...)
			return
		} else {
			this.PutIn(banker[:])
		}
	}
	this.PutIn(xian[:])
	return nil
}

//开闲赢 带庄对
func (this *Poker) GetXianAndBankerPair() (cards []int32) {
	banker := [3]int32{-1, -1, -1}
	irand := this.GetRandPair()
	if irand == -1 {
		return nil
	}
	buf, npair := delSclice(this.buf, []int32{irand, irand})
	if len(npair) != 2 {
		return nil
	}
	this.buf = buf
	banker = [3]int32{npair[0], npair[1], -1}
	bankerNum := num(banker)
	for i := 0; i < 20; i++ {
		xian := [3]int32{-1, -1, -1}
		for k := 0; k < 2; k++ {
			c, _ := this.Next()
			xian[k] = c
		}
		if isPair(xian) {
			this.PutIn(xian[:])
			continue
		}
		xianNum := num(xian)
		if bankerNum < xianNum {
			cards = append(cards, xian[:]...)
			cards = append(cards, banker[:]...)
			return
		} else {
			this.PutIn(xian[:])
		}
	}
	this.PutIn(banker[:])
	return nil
}

//开和赢 带庄对
func (this *Poker) GetTieAndBankerPair() (cards []int32) {
	banker := [3]int32{-1, -1, -1}
	irand := this.GetRandPair()
	if irand == -1 {
		return nil
	}
	buf, npair := delSclice(this.buf, []int32{irand, irand})
	if len(npair) != 2 {
		return nil
	}
	this.buf = buf
	banker = [3]int32{npair[0], npair[1], -1}
	bankerNum := num(banker)
	for i := 0; i < 20; i++ {
		xian := [3]int32{-1, -1, -1}
		//for k := 0; k < 2; k++ {
		//	c, _ := this.Next()
		//	xian[k] = c
		//}

		c, _ := this.Next()
		xian[0] = c
		bn := bankerNum
		xn := c%13 + 1
		if xn >= 10 {
			xn = 0
		}
		if bankerNum < int(xn) {
			bn += 10
		}
		another := this.FindCard(int32(bn) - xn)
		if another == -1 {
			this.PutIn(xian[:])
			continue
		}
		xian[1] = another

		if isPair(xian) {
			this.PutIn(xian[:])
			continue
		}
		xianNum := num(xian)
		if bankerNum == xianNum {
			cards = append(cards, xian[:]...)
			cards = append(cards, banker[:]...)
			return
		} else {
			this.PutIn(xian[:])
		}
	}
	this.PutIn(banker[:])
	return nil
}

//开和赢 带闲对
func (this *Poker) GetTieAndXianPair() (cards []int32) {
	xian := [3]int32{-1, -1, -1}
	imax := this.GetRandPair()
	if imax == -1 {
		return nil
	}
	buf, npair := delSclice(this.buf, []int32{imax, imax})
	if len(npair) != 2 {
		return nil
	}
	this.buf = buf
	xian = [3]int32{npair[0], npair[1], -1}
	xianNum := num(xian)
	for i := 0; i < 20; i++ {
		banker := [3]int32{-1, -1, -1}
		c, _ := this.Next()
		banker[0] = c
		xn := xianNum
		bn := c%13 + 1
		if bn >= 10 {
			bn = 0
		}
		if xianNum < int(bn) {
			xn += 10
		}
		another := this.FindCard(int32(xn) - bn)
		if another == -1 {
			this.PutIn(banker[:])
			continue
		}
		banker[1] = another
		if isPair(banker) {
			this.PutIn(banker[:])
			continue
		}
		bankerNum := num(banker)
		if bankerNum == xianNum {
			cards = append(cards, xian[:]...)
			cards = append(cards, banker[:]...)
			return
		} else {
			this.PutIn(banker[:])
		}
	}
	this.PutIn(xian[:])
	return nil
}

//开庄赢 开庄对 闲对
func (this *Poker) GetBankerAndBankerXianPair() (cards []int32) {
	irand := this.GetMinRandPair()
	if irand == -1 {
		return nil
	}
	imax := this.GetMaxPair(irand)
	if imax == -1 {
		return nil
	}

	buf, minpair := delSclice(this.buf, []int32{irand, irand})
	if len(minpair) != 2 {
		return nil
	}
	this.buf = buf
	buf2, maxpair := delSclice(this.buf, []int32{imax, imax})
	if len(maxpair) != 2 {
		return nil
	}
	this.buf = buf2
	banker := [3]int32{maxpair[0], maxpair[1], -1}
	xian := [3]int32{minpair[0], minpair[1], -1}

	bankerNum := num(banker)
	xianNum := num(xian)
	if bankerNum > xianNum {
		cards = append(cards, xian[:]...)
		cards = append(cards, banker[:]...)
		return
	} else if bankerNum < xianNum {
		cards = append(cards, banker[:]...)
		cards = append(cards, xian[:]...)
		return
	}
	this.PutIn(banker[:])
	this.PutIn(xian[:])
	return nil
}

//开闲赢 开庄对 闲对
func (this *Poker) GetXianAndBankerXianPair() (cards []int32) {
	irand := this.GetMinRandPair()
	if irand == 100 {
		return nil
	}
	imax := this.GetMaxPair(irand)
	if imax == -1 {
		return nil
	}

	buf, minpair := delSclice(this.buf, []int32{irand, irand})
	if len(minpair) != 2 {
		return nil
	}
	this.buf = buf
	buf2, maxpair := delSclice(this.buf, []int32{imax, imax})
	if len(maxpair) != 2 {
		return nil
	}
	this.buf = buf2
	xian := [3]int32{maxpair[0], maxpair[1], -1}
	banker := [3]int32{minpair[0], minpair[1], -1}

	bankerNum := num(banker)
	xianNum := num(xian)
	if bankerNum < xianNum {
		cards = append(cards, xian[:]...)
		cards = append(cards, banker[:]...)
		return
	} else if bankerNum > xianNum {
		cards = append(cards, banker[:]...)
		cards = append(cards, xian[:]...)
		return
	}
	this.PutIn(banker[:])
	this.PutIn(xian[:])
	return nil
}

//开和 开庄对 闲对
func (this *Poker) GetTieAndBankerXianPair() (cards []int32) {
	ifour := this.GetFourCards()
	if ifour == -1 {
		return nil
	}
	buf, ifours := delSclice(this.buf, []int32{ifour, ifour, ifour, ifour})
	if len(ifours) != 4 {
		return nil
	}
	this.buf = buf
	return []int32{ifours[0], ifours[1], -1, ifours[2], ifours[3], -1}
}
func (this *Poker) GetFourCards() int32 {
	ipairs := make(map[int32]int)
	for _, v := range this.buf {
		ipairs[v%13+1]++
	}
	isclice := make([]int32, 0)
	for k, v := range ipairs {
		if v >= 4 {
			isclice = append(isclice, k)
		}
	}
	if len(isclice) == 0 {
		return -1
	}
	return isclice[rand.Intn(len(isclice))]
}
func (this *Poker) GetRandPair() int32 {
	ipairs := make(map[int32]int)
	randpair := make([]int32, 0)
	for _, v := range this.buf {
		i := v%13 + 1
		ipairs[i]++
		if ipairs[i] >= 2 {
			isHave := false
			for _, n := range randpair {
				if n == i {
					isHave = true
					break
				}
			}
			if !isHave {
				randpair = append(randpair, i)
			}
		}
	}
	if len(randpair) == 0 {
		return -1
	}
	optimal := make([]int32, 0)
	for _, v := range randpair {
		if v < 10 {
			optimal = append(optimal, v)
		}
	}
	if len(optimal) > 0 {
		return optimal[rand.Intn(len(optimal))]
	}
	return randpair[rand.Intn(len(randpair))]
}
func (this *Poker) GetMinRandPair() int32 {
	ipairs := make(map[int32]int)
	randpair := make([]int32, 0)
	for _, v := range this.buf {
		i := v%13 + 1
		if i == 1 || i == 2 || i == 6 || i == 7 {
			ipairs[i]++
		}
		if ipairs[i] >= 2 {
			randpair = append(randpair, i)
		}
	}
	if len(randpair) == 0 {
		return -1
	}
	return randpair[rand.Intn(len(randpair))]
}
func (this *Poker) GetMaxPair(irand int32) int32 {
	ipairs := make(map[int32]int)
	imaxpair := make([]int32, 0)
	for _, v := range this.buf {
		i := v%13 + 1
		ipairs[i]++
		if i < 10 && i > irand && ipairs[i] >= 2 {
			imaxpair = append(imaxpair, i)
		}
	}
	if len(imaxpair) == 0 {
		return -1
	}
	return imaxpair[rand.Intn(len(imaxpair))]
}
func delSclice(y, d []int32) (ny, np []int32) {
	ny = make([]int32, len(y))
	np = make([]int32, 0)
	copy(ny, y)
	cmap := make(map[int32]int)
	for i := 0; i < len(d); i++ {
		for k, v := range ny {
			c := v%13 + 1
			color := rand.Int31n(4)
			if cmap[color] >= 1 {
				color++
				color = color % 4
			}
			if c == d[i] && (v%52+1)/13 == color {
				cmap[color]++
				np = append(np, v%52)
				ny = append(ny[:k], ny[k+1:]...)
				break
			}
		}
	}
	return ny, np
}

func num(cards [3]int32) (num int) {
	for _, c := range cards {
		temp := int(c%13 + 1)
		if temp > 0 && temp < 10 {
			num += temp
			if num >= 10 {
				num -= 10
			}
		}
	}
	return
}

func isPair(cards [3]int32) bool {
	a := cards[0]%13 + 1
	b := cards[1]%13 + 1
	c := cards[2]%13 + 1
	if a != b && a != c && b != c {
		return false
	}
	return true
}

//单控补牌
func (this *Poker) SingleRepairCard(cards []int32) (bool, []int32) {
	xianPoint := GetPointNum(cards[:], 0, 1)   //闲家点数
	bankerPoint := GetPointNum(cards[:], 3, 4) //庄家点数
	if xianPoint == 8 || xianPoint == 9 ||
		bankerPoint == 8 || bankerPoint == 9 {
		return true, cards
	}

	xianRepair := make(map[int32]bool)
	for i := int32(0); i <= 5; i++ {
		xianRepair[i] = true
	}
	xianNotRepair := make([]int32, 0)

	if bankerPoint == 3 {
		xianNotRepair = append(xianNotRepair, 8)
	} else if bankerPoint == 4 {
		xianNotRepair = append(xianNotRepair, 8, 9, 0)
	} else if bankerPoint == 5 {
		xianNotRepair = append(xianNotRepair, 1, 2, 3, 8, 9, 0)
	} else if bankerPoint == 6 {
		xianNotRepair = append(xianNotRepair, 1, 2, 3, 4, 5, 8, 9, 0)
	} else if bankerPoint == 7 {
		if _, ok := xianRepair[xianPoint]; ok {
			//闲必须补0
			for i := 0; i < this.Count(); i++ {
				c, _ := this.Next()
				if (c%13 + 1) >= 10 {
					cards[2] = c % 52
					return true, cards
				} else {
					this.PutIn([]int32{c})
				}
			}
		}
		return true, cards
	}
	//庄补牌
	bankerNeedCard := func() bool {
		xianrepair := cards[2]
		if xianrepair == -1 && bankerPoint == 6 {
			return true
		}
		for i := 0; i < this.Count(); i++ {
			c, _ := this.Next()
			repair := c%13 + 1
			bankerpt := bankerPoint
			if repair < 10 {
				bankerpt += repair
				if bankerpt >= 10 {
					bankerpt -= 10
				}
			}
			xianpt := GetPointNum(cards[:], 0, 1, 2)
			if (bankerPoint > xianPoint && bankerpt > xianpt) ||
				(bankerPoint < xianPoint && bankerpt < xianpt) ||
				(bankerPoint == xianPoint && bankerpt == xianpt) {
				cards[5] = c % 52
				return true
			}
			this.PutIn([]int32{c})
		}
		return false
	}
	//闲补牌
	if _, ok := xianRepair[xianPoint]; ok {
		for i := 0; i < this.Count(); i++ {
			c, _ := this.Next()
			repair := c%13 + 1
			isHave := false
			for _, v := range xianNotRepair {
				rnew := repair
				if rnew >= 10 {
					rnew = 0
				}
				if rnew == v {
					isHave = true
					break
				}
			}
			//三张牌一样
			isSame := cards[0]%13+1 == cards[1]%13+1 && cards[0]%13+1 == repair && cards[1]%13+1 == repair
			if !isHave {
				//20%的概率 三张相同
				r := rand.Intn(100)
				if isSame && r < 80 {
					this.PutIn([]int32{c})
					continue
				}
				cards[2] = c % 52
				if bankerNeedCard() {
					return true, cards
				} else {
					cards[2] = -1
				}
			}
			this.PutIn([]int32{c})
		}
	}
	if bankerNeedCard() {
		return true, cards
	}
	return false, cards
}
