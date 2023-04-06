package redvsblack

import (
	"math/rand"
	"time"
)

const (
	POKER_CART_CNT     int = 52
	PER_CARD_COLOR_MAX     = 13
	Hand_CardNum           = 3
	A_CARD                 = 0
)

var cardSeed = time.Now().UnixNano()

type Card int

var CardValueMap = [PER_CARD_COLOR_MAX]int{13, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
var CardFaceMap = [PER_CARD_COLOR_MAX]int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}

type Poker struct {
	buf [POKER_CART_CNT]Card
	pos int
}

func NewPoker() *Poker {
	p := &Poker{}
	p.init()
	return p
}

func (this *Poker) init() {
	for i := 0; i < POKER_CART_CNT; i++ {
		this.buf[i] = Card(i)
	}
	this.Shuffle()
}

func (this *Poker) Shuffle() {
	cardSeed++
	rand.Seed(cardSeed)
	for i := 0; i < POKER_CART_CNT; i++ {
		j := rand.Intn(i + 1)
		this.buf[i], this.buf[j] = this.buf[j], this.buf[i]
	}
	this.pos = 0
}

func (this *Poker) Next() Card {
	if this.pos >= len(this.buf) {
		return -1
	}
	c := this.buf[this.pos]
	this.pos++
	return c
}

func (this *Poker) TryNextN(n int) Card {
	if this.pos+n >= len(this.buf) {
		return -1
	}
	return this.buf[this.pos+n]
}

func (this *Poker) ChangeNextN(n int, c Card) bool {
	if this.pos+n >= len(this.buf) {
		return false
	}
	this.buf[this.pos+n] = c
	return true
}

func (this *Poker) Count() int {
	if len(this.buf) >= this.pos {
		cnt := len(this.buf) - this.pos
		if cnt < 0 {
			cnt = 0
		}
		return cnt
	}
	return 0
}
func (this *Poker) TryCard(flag, kind int) ([]int, []int) {
	this.Shuffle()
	rCards := []int{}
	bCards := []int{}
	for i := 0; i < 6; i++ {
		cards := [2][]int{}
		for j := 0; j < Hand_CardNum; j++ {
			cards[0] = append(cards[0], int(this.Next()))
			cards[1] = append(cards[1], int(this.Next()))
		}
		cardsKind := [2]*KindOfCard{}
		cardsKind[0] = CardsKindFigureUpSington.FigureUpByCard(cards[0])
		cardsKind[1] = CardsKindFigureUpSington.FigureUpByCard(cards[1])
		if flag == 0 {
			if CompareCards(cardsKind[0], cardsKind[1]) == 1 && cardsKind[0].Kind == kind {
				return cards[0], cards[1]
			}
			if CompareCards(cardsKind[0], cardsKind[1]) == 0 && cardsKind[1].Kind == kind {
				return cards[1], cards[0]
			}
		}
		if flag == 1 {
			if CompareCards(cardsKind[0], cardsKind[1]) == 0 && cardsKind[1].Kind == kind {
				return cards[0], cards[1]
			}
			if CompareCards(cardsKind[0], cardsKind[1]) == 1 && cardsKind[0].Kind == kind {
				return cards[1], cards[0]
			}
		}
		bCards = cards[0]
		rCards = cards[1]
	}
	return bCards, rCards
}
