package common

import (
	"math/rand"
)

// /////////////////////////////////////a2b
func MapToint32(a map[int]bool) (b []int32) {
	b = make([]int32, len(a))
	i := 0
	for k, _ := range a {
		b[i] = int32(k)
		i++
	}
	return
}
func Int32Toint(a []int32) (b []int) {
	b = make([]int, len(a))
	for k, v := range a {
		b[k] = int(v)
	}
	return
}
func ThreeTonullArray(a [3]int) (b []int32) {
	b = make([]int32, len(a), len(a))
	for i := 0; i < len(a); i++ {
		b[i] = int32(a[i])
	}
	return
}
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
func IntToInt64(a []int) []int64 {
	c := make([]int64, len(a), len(a))
	for k, v := range a {
		c[k] = int64(v)
	}
	return c
}
func Int64Toint(a []int64) []int {
	c := make([]int, len(a), len(a))
	for k, v := range a {
		c[k] = int(v)
	}
	return c
}
func Int64ToInt32(a []int64) []int32 {
	c := make([]int32, len(a), len(a))
	for k, v := range a {
		c[k] = int32(v)
	}
	return c
}

// 将数组类型转化为int切片类型的值
func Int32SliceToInt(arr []int32) []int {
	s := make([]int, 0)
	for _, v := range arr {
		s = append(s, int(v))
	}
	return s
}

// 将数组类型转化为int切片类型的值
func IntSliceToInt32(arr []int) []int32 {
	s := make([]int32, 0)
	for _, v := range arr {
		s = append(s, int32(v))
	}
	return s
}
