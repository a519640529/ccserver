package blackjack

import (
	"math/rand"
	"time"
)

var Cards = []*Card{
	{0}, {1}, {2}, {3}, {4}, {5}, {6}, {7}, {8}, {9}, {10}, {11}, {12},
	{13}, {14}, {15}, {16}, {17}, {18}, {19}, {20}, {21}, {22}, {23}, {24}, {25},
	{26}, {27}, {28}, {29}, {30}, {31}, {32}, {33}, {34}, {35}, {36}, {37}, {38},
	{39}, {40}, {41}, {42}, {43}, {44}, {45}, {46}, {47}, {48}, {49}, {50}, {51},
}

func RandomShuffle(cards *[]*Card) {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	*cards = (*cards)[:0]
	for i := 0; i < PokerNum; i++ {
		*cards = append(*cards, Cards...)
	}
	r.Shuffle(len(*cards), func(i, j int) {
		(*cards)[i], (*cards)[j] = (*cards)[j], (*cards)[i]
	})
}

type Card struct {
	value int
}

func (c *Card) Value() int {
	return c.value
}

func (c *Card) Point() int {
	v := c.value%13 + 1
	switch {
	case v >= 2 && v <= 10:
		return v
	case v > 10 && v <= 13:
		return 10
	case v == 1:
		return 1
	}
	return 0
}

func NewCardDefault() *Card {
	return &Card{100}
}

func NewCard(n int32) *Card {
	return &Card{
		value: int(n),
	}
}
