package dezhoupoker

import (
	"fmt"
	"math/rand"
	"sort"
	"time"
)

//
//牌序- K, Q, J,10, 9, 8, 7, 6, 5, 4, 3, 2, 1
//黑桃-51,50,49,48,47,46,45,44,43,42,41,40,39
//红桃-38,37,36,35,34,33,32,31,30,29,28,27,26
//梅花-25,24,23,22,21,20,19,18,17,16,15,14,13
//方片-12,11,10, 9, 8, 7, 6, 5, 4, 3, 2, 1, 0

const (
	POKER_A  int32 = 0
	POKER_2  int32 = 1
	POKER_3  int32 = 2
	POKER_4  int32 = 3
	POKER_5  int32 = 4
	POKER_6  int32 = 5
	POKER_7  int32 = 6
	POKER_8  int32 = 7
	POKER_9  int32 = 8
	POKER_10 int32 = 9
	POKER_J  int32 = 10
	POKER_Q  int32 = 11
	POKER_K  int32 = 12

	POKER_CNT int32 = 52

	PER_CARD_COLOR_MAX       = 13
	POKER_A_Weight     int32 = 13
)

const (
	CardColor_Diamond int32 = iota //0,方块
	CardColor_Spade                //1,梅花
	CardColor_Heart                //2,红桃
	CardColor_Club                 //3,黑桃
	CardColor_Joker                //4,王
)

var CardColor = []string{"♦", "♣", "♥", "♠"}
var PokerValue = []string{"A", "2", "3", "4", "5", "6", "7", "8", "9", "T", "J", "Q", "K"}
var CardValueMap = [PER_CARD_COLOR_MAX]int{13, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}

type Card int

func (c Card) String() string {
	if c >= 0 && c < 53 {
		switch c {
		case 52:
			return "[大王]"
		case 53:
			return "[小王]"
		default:
			cc := c / PER_CARD_COLOR_MAX
			cv := c % PER_CARD_COLOR_MAX
			return fmt.Sprintf("[%s%s]", CardColor[cc], PokerValue[cv])
		}
	}
	return "[-]"
}

func (c Card) Color() int {
	return int(c) / PER_CARD_COLOR_MAX
}

func (c Card) Value() int {
	return CardValueMap[int(c)%PER_CARD_COLOR_MAX]
}

func HandCardShowStr(cards []int32) string {
	if len(cards) != 2 {
		return ""
	}

	temp := []int{CardValueMap[cards[0]%PER_CARD_COLOR_MAX], CardValueMap[cards[1]%PER_CARD_COLOR_MAX]}
	sort.Ints(temp)

	str := PokerValue[temp[0]%PER_CARD_COLOR_MAX] + PokerValue[temp[1]%PER_CARD_COLOR_MAX]
	if temp[0] != temp[1] && cards[0]/PER_CARD_COLOR_MAX == cards[1]/PER_CARD_COLOR_MAX {
		str += "s"
	}
	return str
}

type Poker struct {
	buf [POKER_CNT]Card
	pos int
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
	this.pos = 0
}

func (this *Poker) NextByHands(n int) Card {
	interval := make(map[int][]Card)
	interval[0] = []Card{1, 2, 3, 4, 5}
	interval[1] = []Card{6, 7, 8, 9}
	interval[2] = []Card{10, 11, 12, 0}
	var rate []int
	if n == 0 {
		rate = []int{30, 37, 33}
	} else if n == 1 {
		rate = []int{30, 35, 35}
	} else if n == 2 {
		rate = []int{30, 36, 34}
	} else {
		return this.Next()
	}
	m := randInSliceIndex(rate)
	cards := interval[m]
	if len(cards) != 0 {
		r := rand.Intn(len(cards))
		cardr := cards[r]
		this.FindCardByR(cardr)
	}
	return this.Next()
}

func (this *Poker) FindCardByR(n Card) {
	a := []Card{}
	for k, v := range this.buf {
		if v%13 == n && k >= this.pos {
			a = append(a, v)
		}
	}
	if len(a) != 0 {
		r := rand.Intn(len(a))
		for k, v := range this.buf {
			if v == a[r] {
				if this.pos <= len(this.buf) {
					this.buf[k], this.buf[this.pos] = this.buf[this.pos], this.buf[k]
				}
				break
			}
		}
	}
}

func (this *Poker) Next() Card {
	if this.pos >= len(this.buf) {
		return -1
	}
	c := this.buf[this.pos]
	this.pos++
	return c
}

func (this *Poker) MakeCard(cardValue, cardColor int32) int32 {
	return cardColor*PER_CARD_COLOR_MAX + cardValue
}

func (this *Poker) Count() int {
	if len(this.buf) >= this.pos {
		return len(this.buf) - this.pos
	}
	return 0
}

func (this *Poker) GetRestCard() []int32 {
	cnt := this.Count()
	if cnt <= 0 {
		return []int32{}
	}
	ret := make([]int32, cnt)
	for i := 0; i < cnt; i++ {
		ret[i] = int32(this.buf[this.pos+i])
	}
	return ret
}

func (this *Poker) DelCard(c int32) {
	if this.Count() <= 0 {
		return
	}

	for i := this.pos; i < len(this.buf); i++ {
		if int32(this.buf[i]) == c {
			this.buf[i], this.buf[this.pos] = this.buf[this.pos], this.buf[i]
			this.pos++
			return
		}
	}
}

func (this *Poker) DelCards(cards []int32) {
	for _, c := range cards {
		this.DelCard(c)
	}
}

func randInSliceIndex(pool []int) int {
	var total int
	for _, v := range pool {
		total += v
	}
	val := int(rand.Int31n(int32(total)))
	total = 0
	for index, v := range pool {
		total += v
		if total >= val {
			return index
		}
	}

	return 0
}
