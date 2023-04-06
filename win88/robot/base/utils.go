package base

import "math/rand"

func RandInSliceIndex(pool []int) int {
	var total int
	for _, v := range pool {
		total += v
	}
	val := int(rand.Int31n(int32(total)))
	total = 0
	for index, v := range pool {
		total += v
		if total >= val {
			return index
		}
	}

	return 0
}

func RandInSliceInt32Index(pool []int32) int {
	var total int32
	for _, v := range pool {
		total += v
	}
	val := rand.Int31n(total)
	total = 0
	for index, v := range pool {
		total += v
		if total >= val {
			return index
		}
	}

	return 0
}
