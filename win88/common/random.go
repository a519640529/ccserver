package common

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

const (
	RAND32_M int32 = 2147483647
	RAND32_A       = 48271
	RAND32_Q       = RAND32_M / RAND32_A
	RAND32_R       = RAND32_M % RAND32_A
)

type RandomGenerator struct {
	rand32_state int32
}

func (this *RandomGenerator) RandomSeed(seed int32) {
	this.rand32_state = seed
}

func (this *RandomGenerator) Random() int32 {
	hi := this.rand32_state / RAND32_Q
	lo := this.rand32_state % RAND32_Q
	test := RAND32_A*lo - RAND32_R*hi
	if test > 0 {
		this.rand32_state = test
	} else {
		this.rand32_state = test + RAND32_M
	}
	return this.rand32_state - 1
}

func (this *RandomGenerator) Rand32(max int32) int32 {
	return this.Random() % max
}

func (this *RandomGenerator) GetRandomSeed() int32 {
	return this.rand32_state
}

// [l..u)
func RandInt(args ...int) int {
	switch len(args) {
	case 0:
		return rand.Int()
	case 1:
		if args[0] > 0 {
			return rand.Intn(args[0])
		} else {
			return 0
		}
	default:
		l := args[0]
		u := args[1]
		switch {
		case l == u:
			{
				return l
			}
		case l > u:
			{
				return u + rand.Intn(l-u)
			}
		default:
			{
				return l + rand.Intn(u-l)
			}
		}
	}
}

func RandItemByAvg(s1 []int64) int64 {
	if len(s1)%2 != 0 {
		return 0
	}
	rates := []int64{}
	for i := 0; i < len(s1); i = i + 2 {
		rates = append(rates, s1[i+1])
	}
	index := RandInt(0, len(rates))
	return s1[index*2]
}

func RandItemByWight(s1 []int64) int64 {
	if len(s1)%2 != 0 {
		return 0
	}
	rates := []int64{}
	for i := 0; i < len(s1); i = i + 2 {
		rates = append(rates, s1[i+1])
	}
	index := RandSliceIndexByWight(rates)
	return s1[index*2]
}

func RandSliceIndexByWight(s1 []int64) int {
	total := int64(0)
	for _, v := range s1 {
		total += v
	}
	if total <= 0 {
		return 0
	}
	random := rand.Int63n(total)
	total = 0
	for i, v := range s1 {
		total += v
		if random < total {
			return i
		}
	}
	return 0
}

func RandSliceIndexByWight31N(s1 []int32) int {
	total := int32(0)
	for _, v := range s1 {
		total += v
	}
	if total <= 0 {
		return 0
	}
	random := rand.Int31n(total)
	total = 0
	for i, v := range s1 {
		total += v
		if random < total {
			return i
		}
	}
	return 0
}

func RandSliceIndexByWightN(s1 []int) int {
	total := 0
	for _, v := range s1 {
		total += v
	}
	if total <= 0 {
		return 0
	}
	random := rand.Intn(total)
	total = 0
	for i, v := range s1 {
		total += v
		if random < total {
			return i
		}
	}
	return 0
}

func RandNFromSlice(source []int, n int) []int {
	if len(source) == 0 {
		return source
	}
	if n > len(source) {
		cycle := n / len(source)
		rem := n % len(source)
		for i := 0; i < cycle; i++ {
			source = append(source, source...)
		}
		source = append(source, source[:rem]...)
	}
	idxs := rand.Perm(len(source))
	ret := make([]int, len(source))
	for i := 0; i < len(source); i++ {
		ret[i] = source[idxs[i]]
	}
	return ret[:n]
}
func RandInt32Slice(source []int32) int32 {
	if len(source) == 0 {
		return 0
	}
	return source[rand.Intn(len(source))]
}
func RandFromRange(minValue, maxValue int32) int32 {
	if minValue < 0 || maxValue < 0 {
		return 0
	}
	if minValue >= maxValue {
		return minValue
	}
	return rand.Int31n(maxValue-minValue+1) + minValue
}

func RandFromRangeInt64(minValue, maxValue int64) int64 {
	if minValue < 0 || maxValue < 0 {
		return 0
	}
	if minValue >= maxValue {
		return minValue
	}
	return rand.Int63n(maxValue-minValue+1) + minValue
}

func RandSmsCode() string {
	//seed := rand.Int()
	//code := seed % 999999
	//return strconv.Itoa(code)
	return fmt.Sprintf("%06v", rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(1000000))
}
func RandSlice(n int) []int {
	return rand.Perm(n)
}
func RandValueByRang(value int, l, u int) int {
	r := RandInt(l, u)
	return int(float64(value) * float64(r) / 100)
}

// 随机一个索引，最大的是 中间索引
func RandMaxMiddle(arrLen int) int {
	if arrLen <= 0 {
		return 0
	}
	if arrLen%2 == 0 {
		return rand.Intn(arrLen / 2)
	} else {
		return rand.Intn(arrLen/2 + 1)
	}
}

// 随机一个索引，最小的是 中间索引
func RandLastMiddle(arrLen int) int {
	if arrLen <= 0 {
		return 0
	}
	if arrLen%2 == 0 {
		return rand.Intn(arrLen/2) + arrLen/2
	} else {
		return rand.Intn(arrLen/2+1) + arrLen/2
	}
}

// 获得一个线性的随机概率是否满足
func RandLineInt64(curvalue, minValue, maxValue int64) bool {
	ret := RandFromRangeInt64(minValue, maxValue)
	if curvalue > ret {
		return true
	}
	return false
}

// 获得一个正态分布的随机概率是否满足 cur max 都需要传入正值,用法不太对
func RandNormInt64(cur int64, max int64) bool {
	t := math.Abs(rand.NormFloat64()) / 3
	if t > 1 {
		t = 1
	}
	if cur >= max {
		return true
	}

	return cur > int64(t*float64(max))
}

// 获得最小公倍数
func GetMinCommonRate(array []int64) int64 {
	ret := nlcm(array, len(array))
	return ret
}

func gcd(a int64, b int64) int64 {
	if a < b {
		a, b = b, a
	}

	if b == 0 {
		return a
	} else {
		return gcd(b, a%b)
	}
}

func ngcd(array []int64, n int) int64 {
	if n == 1 {
		return array[0]
	}

	return gcd(array[n-1], ngcd(array, n-1))
}

func lcm(a int64, b int64) int64 {
	return a * b / gcd(a, b)
}

func nlcm(array []int64, n int) int64 {
	if n == 1 {
		return array[0]

	} else {
		return lcm(array[n-1], nlcm(array, n-1))
	}
}

// 随机生成遍历的数组列表
func GetRandomList(num int) []int {
	var ret []int
	for i := 0; i < num; i++ {
		ret = append(ret, i)
		r := RandInt(len(ret))
		ret[i], ret[r] = ret[r], ret[i]
	}

	return ret
}

// 从A中随机出B个数字
func GetBNumFromA(A []int32, B int) []int32 {
	if B >= len(A) {
		return A
	}
	var ret []int32

	var rate []int
	for {
		index := rand.Intn(len(A))
		if !InSliceInt(rate, index) {
			rate = append(rate, index)
			ret = append(ret, A[index])
			if len(ret) >= B {
				return ret
			}
		}
	}
}

func GetBElementFromA(A []interface{}, B int) []interface{} {
	if B >= len(A) {
		return A
	}
	var ret []interface{}

	var rate []int
	for {
		index := rand.Intn(len(A))
		if !InSliceInt(rate, index) {
			rate = append(rate, index)
			ret = append(ret, A[index])
			if len(ret) >= B {
				return ret
			}
		}
	}
}
