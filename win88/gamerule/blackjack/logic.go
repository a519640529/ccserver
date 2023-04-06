package blackjack

import (
	"errors"
)

// 牌型
const (
	CardTypeInvalid = 0
	CardTypeA10     = 1 // 黑杰克：一张A 一张10
	CardTypeFive    = 2 // 五小龙：五张牌且没有爆牌
	CardTypeOther   = 3 // 其它点数：点数小于等于21
	CardTypeBoom    = 4 // 爆牌：点数大于21
)

var CardTypeSort = map[int32]int{
	CardTypeBoom:  1,
	CardTypeOther: 2,
	CardTypeFive:  3,
	CardTypeA10:   4,
}

// 获取牌型和点数
func GetCardsType(cards []*Card) (int32, []int32) {
	l := len(cards)
	if l <= 0 || l > MaxCardNum {
		return 0, []int32{}
	}
	// 黑杰克
	if l == 2 {
		if (cards[0].Point() == 1 && cards[1].Point() == 10) || (cards[0].Point() == 10 && cards[1].Point() == 1) {
			return CardTypeA10, []int32{21}
		}
	}
	// 所有点数
	var point int32
	for _, v := range cards {
		point += int32(v.Point())
	}
	points := []int32{point}
	for _, v := range cards {
		if v.Point() == 1 {
			if point+10 > 21 {
				break
			}
			point += 10
			points = append(points, point)
		}
	}
	i := -1 // 最大点数且不爆的点数下标
	for k, v := range points {
		if v <= 21 {
			i = k
		} else {
			break
		}
	}
	// 五小龙
	if l == 5 {
		if i > -1 {
			if points[i] == 21 {
				return CardTypeFive, []int32{21}
			} else {
				return CardTypeFive, points[:i+1]
			}
		}
	}
	// 爆牌
	if points[0] > 21 {
		return CardTypeBoom, []int32{points[0]}
	}
	// 其他点数
	if points[i] == 21 {
		return CardTypeOther, []int32{21}
	}
	return CardTypeOther, points[:i+1]
}

// 手牌比大小
// bankCards 庄家手牌
// playerCards 闲家手牌
// 返回值
// -1 c1 < c2
// 0  c1 == c2
// 1  c1 > c2
func CompareCards(bankCards, playerCards []*Card) (int, error) {
	t1, p1 := GetCardsType(bankCards)
	t2, p2 := GetCardsType(playerCards)

	if t1 == CardTypeInvalid || t2 == CardTypeInvalid {
		return 0, errors.New("Invalid CardType ")
	}
	// 牌型比较
	if CardTypeSort[t1] > CardTypeSort[t2] {
		return 1, nil
	}
	if CardTypeSort[t1] < CardTypeSort[t2] {
		return -1, nil
	}
	// 庄闲都爆牌，则庄赢
	if t1 == CardTypeBoom {
		return 1, nil
	}
	// 黑杰克，五小龙 平局
	if t1 == CardTypeA10 || t1 == CardTypeFive {
		return 0, nil
	}
	// 比点数
	if p1[len(p1)-1] > p2[len(p2)-1] {
		return 1, nil
	}
	if p1[len(p1)-1] == p2[len(p2)-1] {
		return 0, nil
	}
	return -1, nil
}
