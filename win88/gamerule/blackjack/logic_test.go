package blackjack

import (
	"fmt"
	"testing"
)

func TestCompareCards(t *testing.T) {
	cs := []*Card{{4}, {18}, {51}}
	bs := []*Card{}
	n, err := CompareCards(bs, cs)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(n)
}
