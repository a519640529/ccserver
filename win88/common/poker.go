package common

import (
	"fmt"
	"strings"
)

//牌序- K, Q, J,10, 9, 8, 7, 6, 5, 4, 3, 2, A
//     52  53
//黑桃-51,50,49,48,47,46,45,44,43,42,41,40,39
//红桃-38,37,36,35,34,33,32,31,30,29,28,27,26
//梅花-25,24,23,22,21,20,19,18,17,16,15,14,13
//方片-12,11,10, 9, 8, 7, 6, 5, 4, 3, 2, 1, 0

// cards 将癞子牌转点数牌时使用
var cards = [][]int32{
	{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
	{13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25},
	{26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38},
	{39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51},
	{52, 53},
}

// GetPoint 获取牌点数
func GetPoint(val int32) int32 {
	switch {
	case val == 52:
		return 16
	case val == 53:
		return 17
	default:
		return val%13 + 1
	}
}

// PointValue: 13  12  11  10  9  8  7  6  5  4  3  2  1
// LogicValue: 13  12  11  10  9  8  7  6  5  4  3  15 14
var pointLogic = [18]int32{
	1: 14, 2: 15,
	3: 3, 4: 4, 5: 5, 6: 6, 7: 7, 8: 8, 9: 9, 10: 10, 11: 11, 12: 12, 13: 13,
	14: 1, 15: 2,
	16: 16, 17: 17,
}

// GetLogic 获取逻辑值
func GetLogic(val int32) int32 {
	return pointLogic[GetPoint(val)]
}

// PointToLogic 根据点数获取逻辑值
func PointToLogic(point int32) int32 {
	return pointLogic[point]
}

// LogicToPoint 根据逻辑值获取点数
func LogicToPoint(logic int32) int32 {
	return pointLogic[logic]
}

// GetColor 获取花色
func GetColor(val int32) int32 {
	switch {
	case val >= 0 && val <= 12:
		return 0
	case val >= 13 && val <= 25:
		return 1
	case val >= 26 && val <= 38:
		return 2
	case val >= 39 && val <= 51:
		return 3
	default:
		return val
	}
}

// CreatCard 根据点数和花色生成牌
func CreatCard(point int32, color int32, defaultColor int32) int32 {
	if point == 16 {
		return 52
	}
	if point == 17 {
		return 53
	}
	if color < 0 || color > 3 {
		color = defaultColor
	}
	if color < 0 || color > 3 {
		color = 0
	}
	return cards[int(color)][point-1]
}

var color = [4]string{"♦", "♣", "♥", "♠"}

func String(val int32) string {
	switch {
	case val == -1:
		return ""
	case val == 52:
		return "小王"
	case val == 53:
		return "大王"
	default:
		return fmt.Sprint(GetPoint(val), color[GetColor(val)])
	}
}

func StringCards(cards []int32) string {
	s := strings.Builder{}
	for _, v := range cards {
		s.WriteString(String(v))
		s.WriteString(" ")
	}
	return s.String()
}
func StringCardsInt(cards []int) string {
	s := strings.Builder{}
	for _, v := range cards {
		s.WriteString(String(int32(v)))
		s.WriteString(" ")
	}
	return s.String()
}
