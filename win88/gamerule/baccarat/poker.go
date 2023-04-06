package baccarat

import (
	"math/rand"
	"time"
)

const (
	POKER_CART_CNT     int = 52
	PER_CARD_COLOR_MAX     = 13
)

//随机数种子，在连续随机时，1纳秒内就要计算多次
//如果使用UnixNano()则会造成不是真正的随机数，在这里需要采用自增的形式
var BaccaratRandSeed = time.Now().UnixNano()

type Poker struct {
	buf []int32
	pos int
}

func NewPoker() *Poker {
	p := &Poker{}
	//BaccaratRandSeed = time.Now().UnixNano()
	//rand.Seed(BaccaratRandSeed)
	BaccaratRandSeed++
	rand.Seed(BaccaratRandSeed)
	p.Shuffle()
	return p
}

//将n副牌洗在一起
func (this *Poker) Shuffle() {
	//6-8随机
	n := (rand.Intn(3) + 6) * POKER_CART_CNT
	this.buf = nil
	for i := 0; i < n; i++ {
		this.buf = append(this.buf, int32(i))
	}
	for i := 0; i < n; i++ {
		j := rand.Intn(i + 1)
		this.buf[i], this.buf[j] = this.buf[j], this.buf[i]
	}
}

//这个地方的洗牌只会对未发的牌操作，已发的牌是不会变的
func (this *Poker) ShuffleNumCard(n int) {
	//BaccaratRandSeed = atomic.AddInt64(&BaccaratRandSeed, 1)
	//rand.Seed(BaccaratRandSeed)
	l := len(this.buf)
	for i := this.pos; i < this.pos+n; i++ {
		j := this.pos + rand.Intn(l-this.pos-1) + 1
		this.buf[i], this.buf[j] = this.buf[j], this.buf[i]
	}
}

//随机拿牌
func (this *Poker) Next() (int32, bool) {
	flag := false
	if len(this.buf) < 6 {
		flag = true
		this.Shuffle()
	}
	n := rand.Intn(len(this.buf))
	c := this.buf[n] % int32(POKER_CART_CNT)
	this.buf = append(this.buf[:n], this.buf[n+1:]...)
	return c, flag
}
func (this *Poker) FindCard(c int32) (nc int32) {
	nc = -1
	for k, v := range this.buf {
		num := v%13 + 1
		if c == num || (num > 10 && c == 0) {
			nc = v % int32(POKER_CART_CNT)
			this.buf = append(this.buf[:k], this.buf[k+1:]...)
			break
		}
	}
	return
}

//返还牌
func (this *Poker) PutIn(c []int32) {
	for _, v := range c {
		if v != -1 {
			this.buf = append(this.buf, v)
		}
	}
}

//拿出一组牌
func (this *Poker) TakeOut(c []int32) {
	if len(c) == 0 {
		return
	}
	for _, v := range c {
		for m, n := range this.buf {
			if v == n {
				this.buf = append(this.buf[:m], this.buf[m+1:]...)
				break
			}
		}
	}
}

func (this *Poker) TryNextN(n int) int32 {
	if this.pos+n >= len(this.buf) {
		return -1
	}
	return this.buf[this.pos+n]
}

func (this *Poker) ChangeNextN(n int, c int32) bool {
	if this.pos+n >= len(this.buf) {
		return false
	}
	this.buf[this.pos+n] = c
	return true
}

func (this *Poker) Count() int {
	return len(this.buf)
}

//card为-1则不在统计
func GetPointNum(cards []int32, pos ...int) int32 {
	value := int32(0)
	for _, e := range pos {
		temp := cards[e]%13 + 1
		if temp > 0 && temp < 10 {
			value += temp
			if value >= 10 {
				value -= 10
			}
		}
	}
	return int32(value)
}

//分析是否满足闲家补一张的条件
func (this *Poker) IsNeedPlayerAndOne(cards []int32) bool {
	player_point := GetPointNum(cards, 0, 1)
	if player_point < 6 {
		banker_point := GetPointNum(cards, 3, 4)
		if banker_point == 8 || banker_point == 9 {
			return false
		}
		return true
	}
	return false
}

//分析是否满足庄家补一张的条件
func (this *Poker) IsNeedBankerAndOne(cards []int32) bool {
	player_point := GetPointNum(cards, 0, 1)
	if player_point == 8 || player_point == 9 {
		return false
	}
	//闲没补牌 player_card_value结果为0 所以更改10的值
	player_card_value := int32(-1)
	if cards[2]%13 != -1 {
		player_card_value = cards[2]%13 + 1
	}
	if player_card_value >= 10 {
		player_card_value = 0
	}
	banker_point := GetPointNum(cards, 3, 4)
	switch banker_point {
	case 3:
		if player_card_value == 8 {
			return false
		}
	case 4:
		if player_card_value == 8 || player_card_value == 9 ||
			player_card_value == 0 {
			return false
		}
	case 5:
		if player_card_value == 1 || player_card_value == 2 ||
			player_card_value == 3 || player_card_value == 8 ||
			player_card_value == 9 || player_card_value == 0 {
			return false
		}
	case 6:
		if player_card_value != 6 && player_card_value != 7 {
			return false
		}
	case 7:
		return false
	case 8, 9:
		//只要有一个是天王就不用补牌
		return false
	}
	return true
}
