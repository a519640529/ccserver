package dezhoupoker

import (
	"fmt"
	"testing"
)

//牌序- K, Q, J,10, 9, 8, 7, 6, 5, 4, 3, 2, 1
//黑桃-51,50,49,48,47,46,45,44,43,42,41,40,39
//红桃-38,37,36,35,34,33,32,31,30,29,28,27,26
//梅花-25,24,23,22,21,20,19,18,17,16,15,14,13
//方片-12,11,10, 9, 8, 7, 6, 5, 4, 3, 2, 1, 0

func TestCards(t *testing.T) {

	cardInfo1 := &CardsInfo{Kind: KindOfCard_Straight, KindCards: []int32{26, 22, 23, 24, 25}}
	cardInfo1.CalValue()

	cardInfo2 := &CardsInfo{Kind: KindOfCard_Straight, KindCards: []int32{39, 35, 23, 24, 25}}
	cardInfo2.CalValue()

	fmt.Println(cardInfo1.ValueScore, cardInfo2.ValueScore)

	fmt.Println("ts")
}

func TestKindOfCardFigureUpEx_FigureUpByCard(t *testing.T) {
	testcases := []struct {
		Cards []int32
		Kind  int32
	}{
		{[]int32{2, 3}, KindOfCard_HighCard},
		{[]int32{2, 3, 7, 9, 19}, KindOfCard_HighCard},
		{[]int32{0, 13}, KindOfCard_OnePair},
		{[]int32{2, 15, 7, 16, 17}, KindOfCard_OnePair},
		{[]int32{2, 15, 3, 16, 17}, KindOfCard_TwoPair},
		{[]int32{2, 15, 3, 16, 17, 18}, KindOfCard_TwoPair},
		{[]int32{2, 15, 3, 16, 17, 18, 7}, KindOfCard_TwoPair},
		{[]int32{2, 15, 28, 16, 17, 18}, KindOfCard_ThreeKind},
		{[]int32{2, 15, 28, 41, 17, 18}, KindOfCard_FourKind},
		{[]int32{2, 16, 17, 18, 19, 9}, KindOfCard_Straight},
		{[]int32{2, 16, 17, 18, 19}, KindOfCard_Straight},
		{[]int32{21, 16, 17, 18, 19}, KindOfCard_Flush},
		{[]int32{21, 16, 17, 18, 19, 1}, KindOfCard_Flush},
		{[]int32{3, 16, 4, 17, 30, 1}, KindOfCard_Fullhouse},
		{[]int32{3, 16, 4, 17, 30}, KindOfCard_Fullhouse},
		{[]int32{15, 16, 17, 18, 19}, KindOfCard_StraightFlush},
		{[]int32{15, 16, 17, 18, 19, 1}, KindOfCard_StraightFlush},
		{[]int32{0, 9, 10, 11, 12, 1}, KindOfCard_RoyalFlush},
	}

	for _, c := range testcases {
		ci := KindOfCardFigureUpExSington.FigureUpByCard(c.Cards)
		if ci == nil || ci.Kind != c.Kind {
			t.Errorf("KindOfCardFigureUpEx_FigureUpByCard test cards%v, expect=%v but ci=%#v", c.Cards, c.Kind, ci)
		}
	}
}

func TestKindOfCardFigureUpEx_IsTing(t *testing.T) {
	testcases := []struct {
		Cards []int32
		Kind  int32
	}{
		{[]int32{2, 3, 8, 18, 19, 9}, KindOfCard_Straight},
		{[]int32{2, 3, 17, 18, 10, 9}, KindOfCard_Straight},
		{[]int32{1, 16, 17, 18, 19}, KindOfCard_Flush},
		{[]int32{1, 3, 4, 7, 19, 40}, KindOfCard_Flush},
		{[]int32{15, 16, 0, 18, 19}, KindOfCard_StraightFlush},
		{[]int32{15, 16, 17, 18, 7}, KindOfCard_StraightFlush},
		{[]int32{0, 9, 25, 11, 12, 1}, KindOfCard_RoyalFlush},
		{[]int32{30, 9, 10, 11, 12, 1}, KindOfCard_RoyalFlush},
		{[]int32{30, 47, 22, 33, 50, 31}, KindOfCard_Straight},
	}
	for _, c := range testcases {
		if !KindOfCardFigureUpExSington.IsTing(c.Cards, c.Kind) {
			t.Errorf("KindOfCardFigureUpEx_FigureUpByCard test cards%v, expectTing=%v but not", c.Cards, c.Kind)
		}
	}
}

func TestHandCardShowStr(t *testing.T) {
	testcases := []struct {
		Cards   []int32
		ShowStr string
	}{
		{[]int32{0, 10}, "JAs"},
		{[]int32{2, 3}, "34s"},
		{[]int32{13, 10}, "JA"},
		{[]int32{9, 22}, "TT"},
		{[]int32{10, 23}, "JJ"},
		{[]int32{2, 0}, "3As"},
	}
	for _, c := range testcases {
		str := HandCardShowStr(c.Cards)
		if str != c.ShowStr {
			t.Errorf("HandCardShowStr test cards%v, expect=%v but %v", c.Cards, c.ShowStr, str)
		}
	}
}

func TestKindOfCardIsBetter(t *testing.T) {
	testcases := []struct {
		HandCard   []int32
		CommonCard []int32
	}{
		{[]int32{2, 3}, []int32{16, 7, 8}},
		{[]int32{2, 3}, []int32{16, 7, 8, 4}},
		{[]int32{2, 3}, []int32{16, 7, 8, 15, 9}},
		{[]int32{2, 3}, []int32{16, 7, 8, 15, 28}},
	}
	for _, c := range testcases {
		if !KindOfCardIsBetter(c.HandCard, c.CommonCard) {
			t.Errorf("KindOfCardIsBetter test handcards%v, commoncard%v expect better", c.HandCard, c.CommonCard)
		}
	}
}
