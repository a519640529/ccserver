package api3th

import "bytes"

// 牌数据转换
// 1 2 3 4 5 6 7 8 9 10 11 12 13 52 53 =>
// A 2 3 4 5 6 7 8 9 T  J  Q  K  B  C
func CardValueToShowCard(cards []int32) string {
	ret := bytes.NewBuffer([]byte{})
	for _, v := range cards {
		ret.WriteString(AIRecordMap[GetPoint(v)])
	}
	return ret.String()
}

// 扑克牌点数
func GetPoint(v int32) int32 {
	if v >= 52 {
		return v
	}
	return v%13 + 1
}

var AIRecordMap = map[int32]string{
	53: "C",
	52: "B",
	1:  "A",
	2:  "2",
	3:  "3",
	4:  "4",
	5:  "5",
	6:  "6",
	7:  "7",
	8:  "8",
	9:  "9",
	10: "T",
	11: "J",
	12: "Q",
	13: "K",
}

var AIPointMap = map[string]int32{
	"D": 53,
	"X": 52,
	"C": 53,
	"B": 52,
	"A": 1,
	"2": 2,
	"3": 3,
	"4": 4,
	"5": 5,
	"6": 6,
	"7": 7,
	"8": 8,
	"9": 9,
	"T": 10,
	"J": 11,
	"Q": 12,
	"K": 13,
}

// 德州扑克接入的机器人接口需要的牌定义
// SHDC 黑红梅方
// 0:2S 1:3S ... 11:13S 12:1S
// 13:2H 14:3H ... 24:13H 25:1H
// 26:2D 27:3D ... 37:13D 38:1D
// 39:2C 40:3C ... 50:13C 51:1C
func DZPCardToAICard(card int32) int32 {
	var ret int32
	// 点数
	if card%13 == 0 {
		ret += 12
	} else {
		ret = card%13 - 1
	}
	// 花色
	ret += (3 - card/13) * 13
	return ret
}
