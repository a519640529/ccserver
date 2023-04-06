package tienlen

import (
	"math/rand"
	"time"
)

//牌序- 2, A, K, Q, J, 10, 9, 8, 7, 6, 5, 4, 3
//红桃- 51,50,49,48,47,46,45,44,43,42,41,40,39
//方片- 38,37,36,35,34,33,32,31,30,29,28,27,26
//梅花- 25,24,23,22,21,20,19,18,17,16,15,14,13
//黑桃- 12,11,10, 9, 8, 7, 6, 5, 4, 3, 2, 1, 0

const (
	POKER_CNT          = 52
	PER_CARD_COLOR_MAX = 13
	Hand_CardNum       = 13 //手牌
	Card_Value_2       = 12 //特殊牌值2
	Player_Num_Max     = 4  //桌上最大人數
	HongTao2           = 51
	FangPian2          = 38
	MeiHua2            = 25
	HeiTao2            = 12
)

type Card int

type Poker struct {
	buf [POKER_CNT]Card
}

func (this *Poker) GetPokerBuf() [POKER_CNT]Card {
	return this.buf
}

func NewPoker() *Poker {
	p := &Poker{}
	p.init()
	return p
}

func (this *Poker) init() {
	for i := int32(0); i < POKER_CNT; i++ {
		this.buf[i] = Card(i)
	}
	rand.Seed(time.Now().UnixNano())
	this.Shuffle()
}

func (this *Poker) Shuffle() {
	for i := int32(0); i < POKER_CNT; i++ {
		j := rand.Intn(int(i) + 1)
		this.buf[i], this.buf[j] = this.buf[j], this.buf[i]
	}
}
