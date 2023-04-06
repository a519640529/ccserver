package fishing

import "math"

func PowerValide(sceneType int32, power int32) bool {
	low := math.Pow10(int(sceneType))
	high := math.Pow10(int(sceneType) + 1)
	if power < int32(low) || power > int32(high) {
		return false
	} else {
		return true
	}
}

/*
	获取最小值
*/
func Min(nums ...int) int {
	var min int
	for _, val := range nums {
		if min == 0 || val <= min {
			min = val
		}
	}
	return min
}
