package dragonvstiger

import (
	"math/rand"
	"time"
)

const (
	POKER_CART_CNT     = 52
	PER_CARD_COLOR_MAX = 13
	CardColor_Max      = 4
	Hand_CardNum       = 3
	A_CARD             = 0
	K_CARD             = 12
	T_CARD             = 9
)

var cardSeed = time.Now().UnixNano()

type Card int

var CardValueMap = [PER_CARD_COLOR_MAX]int{13, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}

type Poker struct {
	buf []Card
	//todo test 指定的牌放在最前面
	ctrlPokers []int64
}
type PokerMemData struct {
	Buf []Card
}

func NewPoker() *Poker {
	p := &Poker{}
	p.Shuffle()
	return p
}

func (this *Poker) Shuffle() {
	this.buf = this.buf[:0]
	for i := 0; i < POKER_CART_CNT; i++ {
		this.buf = append(this.buf, Card(i))
	}
	cardSeed++
	rand.Seed(cardSeed)
	for i := 0; i < len(this.buf); i++ {
		j := rand.Intn(i + 1)
		this.buf[i], this.buf[j] = this.buf[j], this.buf[i]
	}
}
func (this *Poker) Next() Card {
	if len(this.buf) <= 0 {
		return -1
	}
	c := this.buf[0]
	this.buf = this.buf[1:]
	return c
}

func (this *Poker) Count() int {
	return len(this.buf)
}

func (this *Poker) TryDTDraw(flag int32) (int32, int32) {
	dCard := this.Next()
	tCard := this.Next()
	d := dCard % PER_CARD_COLOR_MAX
	t := tCard % PER_CARD_COLOR_MAX
	switch flag {
	case 1: // 龙
		if d > t {
			break
		}
		if d < t {
			dCard, tCard = tCard, dCard
			break
		}
		// 龙给一张大于1的牌
		for i := 0; i < len(this.buf); i++ {
			if this.buf[i]%PER_CARD_COLOR_MAX > 0 {
				dCard, this.buf[i] = this.buf[i], dCard
				break
			}
		}
		d = dCard % PER_CARD_COLOR_MAX
		for i := 0; i < len(this.buf); i++ {
			if d > this.buf[i]%PER_CARD_COLOR_MAX {
				tCard, this.buf[i] = this.buf[i], tCard
				break
			}
		}
	case 2: // 虎
		if d < t {
			break
		}
		if d > t {
			dCard, tCard = tCard, dCard
			break
		}
		// 虎给一张大于1的牌
		for i := 0; i < len(this.buf); i++ {
			if this.buf[i]%PER_CARD_COLOR_MAX > 0 {
				tCard, this.buf[i] = this.buf[i], tCard
				break
			}
		}
		t = tCard % PER_CARD_COLOR_MAX
		for i := 0; i < len(this.buf); i++ {
			if t > this.buf[i]%PER_CARD_COLOR_MAX {
				dCard, this.buf[i] = this.buf[i], dCard
				break
			}
		}
	default:
		// 和
		if d == t {
			break
		}
		for i := 0; i < len(this.buf); i++ {
			if this.buf[i]%PER_CARD_COLOR_MAX == d {
				tCard, this.buf[i] = this.buf[i], tCard
				break
			}
		}
	}
	return int32(dCard), int32(tCard)
}
