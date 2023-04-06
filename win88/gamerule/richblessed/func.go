package richblessed

import (
	"fmt"
	"math/rand"
	"strconv"
)

func GetRate(ele int32, num int) int64 {
	if data, ok := EleNumRate[ele]; ok {
		if r, ok2 := data[num]; ok2 {
			return r
		}
	}
	return 0
}

func RandJACKPOT(bet int64, big int64) bool { // bet  用户在房间的单次下注/房间内最大下注  big 放大倍率
	ret := false
	if rand.Intn(1000) < 3 {
		if float64(rand.Int63n(big)) < float64(bet)/float64(big) {
			ret = true
		}
	}
	return ret
}

func RandSliceInt32IndexByWightN(s1 []int32) int32 {
	total := 0
	for _, v := range s1 {
		total += int(v)
	}
	if total <= 0 {
		return 0
	}
	random := rand.Intn(total)
	total = 0
	for i, v := range s1 {
		total += int(v)
		if random < total {
			return int32(i)
		}
	}
	return 0
}
func Print(res []int32) {
	fmt.Println(res)
	str := ""
	for k, ele := range res {
		switch ele {
		case Scatter:
			str += "福字  ,"
		case Gongs:
			str += "铜锣   ,"
		case GoldenPhoenix:
			str += "金凤凰  ,"
		case Sailboat:
			str += "帆船   ,"
		case GoldenTortoise:
			str += "金龟   ,"
		case GoldIngot:
			str += "金元宝  ,"
		case Copper:
			str += "金钱币  ,"
		case A:
			str += "A	,"
		case K:
			str += "K	,"
		case Q:
			str += "Q	,"
		case J:
			str += "J	,"
		case Ten:
			str += "10	,"
		case Nine:
			str += "9  ,"
		}
		if (k+1)%5 == 0 {
			fmt.Println("第", strconv.Itoa((k+1)/5), "行     ", str)
			str = ""
		}
	}
}
