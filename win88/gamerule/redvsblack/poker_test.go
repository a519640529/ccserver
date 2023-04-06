package redvsblack

import (
	"testing"
)

func TestPoker(t *testing.T) {
	p := NewPoker()
	dup := make(map[Card]bool)
	for i := 0; i < POKER_CART_CNT; i++ {
		card := p.Next()
		if _, exist := dup[card]; exist {
			t.Fatal("found dup card", card)
		}
		dup[card] = true
		if card == -1 {
			t.Fatal("not expect -1")
		} else {
			t.Log(card)
		}
	}
}

//70% failed to try appoint card
func TestTryCard(t *testing.T) {
	p := NewPoker()
	failedCount := 0
	for i := 0; i < 10000; i++ {
		bCard, rCard := p.TryCard(0, 1)
		if p.Count() > 16 {
			bKind := CardsKindFigureUpSington.FigureUpByCard(bCard)
			rKind := CardsKindFigureUpSington.FigureUpByCard(rCard)
			if CompareCards(bKind, rKind) != 1 {
				t.Log(bCard, rCard)
				t.Log(bKind, rKind)
				t.Fatal()
			}
		} else {
			failedCount++
		}
		bCard, rCard = p.TryCard(1, 1)
		if p.Count() > 16 {
			bKind := CardsKindFigureUpSington.FigureUpByCard(bCard)
			rKind := CardsKindFigureUpSington.FigureUpByCard(rCard)
			if CompareCards(bKind, rKind) != -1 {
				t.Log(bCard, rCard)
				t.Log(bKind, rKind)
				t.Fatal()
			}
		} else {
			//failedCount++
		}
	}
	//fmt.Println(failedCount)
}
