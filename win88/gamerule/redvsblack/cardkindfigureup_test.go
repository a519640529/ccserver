package redvsblack

import (
	"testing"
)

//牌序- K, Q, J,10, 9, 8, 7, 6, 5, 4, 3, 2, 1
//黑桃-51,50,49,48,47,46,45,44,43,42,41,40,39
//红桃-38,37,36,35,34,33,32,31,30,29,28,27,26
//梅花-25,24,23,22,21,20,19,18,17,16,15,14,13
//方片-12,11,10, 9, 8, 7, 6, 5, 4, 3, 2, 1, 0

type TestCaseData struct {
	c1           []int
	c2           []int
	expectResult int
}

func TestCompareCards(t *testing.T) {
	compareFunc := func(data []*TestCaseData) {
		for i := 0; i < len(data); i++ {
			tc := data[i]
			cardsInfoO1 := CardsKindFigureUpSington.FigureUpByCard(tc.c1)
			cardsInfoO2 := CardsKindFigureUpSington.FigureUpByCard(tc.c2)
			result := CompareCards(cardsInfoO1, cardsInfoO2)
			if result != tc.expectResult {
				t.Log(tc)
				t.Fatal(cardsInfoO1, "compare", cardsInfoO2, "expect result=", tc.expectResult, " but result=", result)
			} else {
				//t.Log(cardsInfoO1, "compare", cardsInfoO2, "result=", tc.expectResult)
			}
		}
	}
	testCases := []*TestCaseData{
		&TestCaseData{ //♦2♦3♦4 vs ♦6♦4♦3 牌型不同 比牌型 顺子>散牌
			c1: []int{1, 3, 2}, c2: []int{5, 3, 2}, expectResult: 1,
		},
		&TestCaseData{ //♦3♦5♣2 vs ♦6♣6♥6 牌型不同 比牌型 顺子>散牌
			c1: []int{14, 4, 2}, c2: []int{5, 18, 31}, expectResult: -1,
		},
		&TestCaseData{ //♦3♦5♣2 vs ♦6♣8♦10 牌型相同，比大小
			c1: []int{14, 4, 2}, c2: []int{5, 20, 9}, expectResult: -1,
		},
		&TestCaseData{ //♦3♦5♣2 vs ♥2♦5♦3 牌型相同，比大小，大小相同比花色
			c1: []int{14, 4, 2}, c2: []int{27, 17, 2}, expectResult: -1,
		},
		&TestCaseData{ //♦A♦Q♦K vs ♣A♣Q♣K 牌型相同，比大小，大小相同比花色
			c1: []int{0, 11, 12}, c2: []int{13, 24, 25}, expectResult: -1,
		},
		&TestCaseData{ //♣A♣Q♣K vs ♦A♦Q♦K 牌型相同，比花色
			c1: []int{13, 24, 25}, c2: []int{0, 11, 12}, expectResult: 1,
		},
		&TestCaseData{ //♣A♣Q♣K vs ♦A♦2♦3 牌型相同，比花色
			c1: []int{13, 24, 25}, c2: []int{0, 1, 2}, expectResult: 1,
		},
		&TestCaseData{ //♣J♣Q♣K vs ♦A♦2♦3 牌型相同，比花色
			c1: []int{23, 24, 25}, c2: []int{0, 1, 2}, expectResult: 1,
		},
		&TestCaseData{ //♦A♣A♥A vs ♦2♣2♥2 牌型相同，比大小
			c1: []int{0, 13, 26}, c2: []int{1, 14, 27}, expectResult: 1,
		},
		&TestCaseData{ //♥K♠K♦A vs ♣A♥A♦J 牌型相同，比大小
			c1: []int{38, 51, 0}, c2: []int{13, 26, 10}, expectResult: -1,
		},
		&TestCaseData{ //♦2♦6♣8 vs ♦3♦4♣A 牌型相同，比大小
			c1: []int{1, 5, 20}, c2: []int{2, 3, 13}, expectResult: -1,
		},
		&TestCaseData{ //♦2♦6♥A vs ♦3♦4♣A 牌型相同，比大小，大小相同比花色
			c1: []int{1, 5, 26}, c2: []int{2, 3, 13}, expectResult: 1,
		},
		&TestCaseData{ //♦2♦6♣2 vs ♥2♦4♠2 牌型相同，比单牌大小
			c1: []int{1, 5, 14}, c2: []int{27, 3, 40}, expectResult: 1,
		},
		&TestCaseData{ //♦2♦6♣2 vs ♥2♣7♠2 牌型相同，比单牌大小，大小相同比花色
			c1: []int{1, 5, 14}, c2: []int{27, 19, 40}, expectResult: -1,
		},
		&TestCaseData{ //♦2♦3♦4-♣2♣3♣4 牌型相同，大小相同，比花色
			c1: []int{1, 2, 3}, c2: []int{14, 15, 16}, expectResult: -1,
		},
		&TestCaseData{ //♦2♦3♦4-♦6♦7♦8 牌型相同，比大小
			c1: []int{1, 2, 3}, c2: []int{5, 6, 7}, expectResult: -1,
		},
		&TestCaseData{ //♥6♥7♥8-♦6♦7♦8 牌型相同，比花色
			c1: []int{31, 32, 33}, c2: []int{5, 6, 7}, expectResult: 1,
		},
		&TestCaseData{ //♦2♦3♦4-♣2♣3♣4 牌型相同，大小相同，比花色♣>♦
			c1: []int{1, 2, 3}, c2: []int{14, 15, 16}, expectResult: -1,
		},
		&TestCaseData{ //♦2♦3♦4-♣2♣3♣4 牌型相同，大小相同，比花色
			c1: []int{1, 2, 3}, c2: []int{14, 15, 16}, expectResult: -1,
		},
		&TestCaseData{ //♦A♦2♦3 vs ♦J♦Q♦K 牌型不同，♦A♦2♦3为最小同花顺
			c1: []int{0, 1, 2}, c2: []int{10, 11, 12}, expectResult: -1,
		},
		&TestCaseData{ //♦3♣3♥5 vs ♥3♠3♣5 对子相同，单牌相同，比花色
			c1: []int{2, 15, 30}, c2: []int{28, 41, 17}, expectResult: 1,
		},
		&TestCaseData{ //♦3♣3♥5 vs ♥3♠3♣5 对子相同，单牌相同，全比
			c1: []int{2, 15, 30}, c2: []int{28, 41, 17}, expectResult: 1,
		},
		//对子相同，单牌相同，对子大
		&TestCaseData{ //♦3♣3♥5 vs ♥3♠3♣5 对子相同，单牌相同，全比，比5的花色
			c1: []int{2, 15, 30}, c2: []int{28, 41, 17}, expectResult: 1,
		},
		&TestCaseData{ //♦6♣6♥5 vs ♥6♠6♣5 对子相同，单牌相同，全比，比6的花色
			c1: []int{5, 18, 30}, c2: []int{31, 44, 17}, expectResult: -1,
		},
		&TestCaseData{ //♦3♣3♥5 vs ♥3♠3♣6 对子相同，单牌不同
			c1: []int{2, 15, 30}, c2: []int{28, 41, 18}, expectResult: -1,
		},
		//对子相同，单牌不同，对子大
		&TestCaseData{ //♦10♣10♥5 vs ♥10♠10♣6 对子相同，单牌不同
			c1: []int{9, 22, 30}, c2: []int{35, 48, 18}, expectResult: -1,
		},
		//对子不同，单牌相同，单牌大
		&TestCaseData{ //♦10♣10♥J vs ♥3♠3♠J 对子不同，单牌相同
			c1: []int{9, 22, 36}, c2: []int{28, 41, 49}, expectResult: 1,
		},
		//对子不同，单牌相同，对子大
		&TestCaseData{ //♦10♣10♦8 vs ♥3♠3♥8 对子不同，单牌相同
			c1: []int{9, 22, 7}, c2: []int{28, 41, 33}, expectResult: 1,
		},
		//对子不同，单牌不同，单牌大
		&TestCaseData{ //♦10♣10♥J vs ♥3♠3♥8 对子不同，单牌不同
			c1: []int{9, 22, 36}, c2: []int{28, 41, 33}, expectResult: 1,
		},
		//对子不同，单牌不同，对子大
		&TestCaseData{ //♦10♣10♦7 vs ♥3♠3♥8 对子不同，单牌不同
			c1: []int{9, 22, 6}, c2: []int{28, 41, 33}, expectResult: 1,
		},
		&TestCaseData{ //♣A♣Q♣K vs ♦A♦2♦3 天龙 vs 地龙
			c1: []int{13, 24, 25}, c2: []int{0, 1, 2}, expectResult: 1,
		},
		&TestCaseData{ //♥6♥7♥8 vs ♦A♦2♦3 牌型不同
			c1: []int{31, 32, 33}, c2: []int{0, 1, 2}, expectResult: 1,
		},
		&TestCaseData{ //♥A♥2♥3 vs ♦A♦2♦3 牌型不同，比花色
			c1: []int{26, 27, 28}, c2: []int{0, 1, 2}, expectResult: 1,
		},
		&TestCaseData{ //♥A♥2♥3 vs ♦4♦Q♥J 牌型不同
			c1: []int{26, 27, 28}, c2: []int{3, 11, 36}, expectResult: 1,
		},
		&TestCaseData{ //♦A♦2♦3 vs ♣K♣A♣2 牌型不同
			c1: []int{0, 1, 2}, c2: []int{25, 13, 14}, expectResult: 1,
		},
	}

	compareFunc(testCases)
}

func TestFigureUp(t *testing.T) {
	testFunc := func(testData []TestCaseData) {
		for i := 0; i < len(testData); i++ {
			data := testData[i]
			cardsKind := CardsKindFigureUpSington.FigureUpByCard(data.c1)
			if cardsKind.Kind != data.expectResult {
				t.Log(data)
				t.Log(cardsKind)
				t.Fatal("is not expect")
			} else {
				cardsKind.TidyCards()
				t.Log(cardsKind)
			}
		}
	}
	testData := []TestCaseData{
		//散牌
		{c1: []int{1, 15, 19}, expectResult: CardsKind_Single},
		{c1: []int{1, 4, 15}, expectResult: CardsKind_Single},
		{c1: []int{12, 13, 14}, expectResult: CardsKind_Single},
		//对子
		{c1: []int{1, 14, 15}, expectResult: CardsKind_Double},
		{c1: []int{43, 4, 22}, expectResult: CardsKind_Double}, //♠5♦5♣10
		{c1: []int{8, 21, 15}, expectResult: CardsKind_BigDouble},
		{c1: []int{0, 13, 15}, expectResult: CardsKind_BigDouble},
		//顺子
		{c1: []int{19, 5, 33}, expectResult: CardsKind_Straight},
		{c1: []int{33, 32, 8}, expectResult: CardsKind_Straight},
		{c1: []int{13, 1, 2}, expectResult: CardsKind_StraightA23}, //地龙
		//同花
		{c1: []int{13, 20, 23}, expectResult: CardsKind_Flush},
		//同花顺
		{c1: []int{1, 3, 2}, expectResult: CardsKind_FlushStraight},
		{c1: []int{0, 1, 2}, expectResult: CardsKind_FlushStraightA23},     //地龙
		{c1: []int{0, 11, 12}, expectResult: CardsKind_RoyalFlushStraight}, //天龙
		//豹子 ♦3♣3♥3
		{c1: []int{2, 15, 28}, expectResult: CardsKind_ThreeSame},
	}
	testFunc(testData)

}
