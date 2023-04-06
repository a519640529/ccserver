package roulette

import (
	"fmt"
	"testing"
	"time"
)

func Test1(t *testing.T) {
	point := new(PointType)
	point.Init()
	max := 0
	maxi := 0
	min := 500
	mini := 0
	tim := time.Now()
	t.Log("===============================11", tim)
	for i := 0; i <= 36; i++ {
		pointMapNums := point.PointMapNums[i]
		betTypes := []int{}
		for k, _ := range pointMapNums {
			betTypes = append(betTypes, point.GetBetType(k))
		}
		fmt.Println(i, "betTypes: ", betTypes)
		rate := 0
		for _, v := range betTypes {
			rate += point.RateMap[v]
		}
		if rate > max {
			max = rate
			maxi = i
		}
		if rate < min {
			min = rate
			mini = i
		}
		fmt.Println("==============rate: ", rate)
	}
	t.Log("===============================22", time.Now().Sub(tim))
	fmt.Println(max, maxi)
	fmt.Println(min, mini)

	a := []int{10, 20, 30}
	b := 1
	a = append(a[0:], b)
	fmt.Println("Hello, World!", a)
}
