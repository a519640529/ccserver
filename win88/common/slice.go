package common

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

func CopySliceInt32(s []int32) []int32 {
	n := len(s)
	if n != 0 {
		temp := make([]int32, n, n)
		copy(temp, s)
		return temp
	}
	return nil
}

func CopySliceInt64(s []int64) []int64 {
	n := len(s)
	if n != 0 {
		temp := make([]int64, n, n)
		copy(temp, s)
		return temp
	}
	return nil
}

func CopySliceIntToInt32(s []int) []int32 {
	n := len(s)
	if n != 0 {
		temp := make([]int32, n, n)
		for i := 0; i < n; i++ {
			temp[i] = int32(s[i])
		}
		return temp
	}
	return nil
}

func CopySliceInt32ToInt(s []int32) []int {
	n := len(s)
	if n != 0 {
		temp := make([]int, n, n)
		for i := 0; i < n; i++ {
			temp[i] = int(s[i])
		}
		return temp
	}
	return nil
}

func CopySliceInt32ToInt64(s []int32) []int64 {
	n := len(s)
	if n != 0 {
		temp := make([]int64, n, n)
		for i := 0; i < n; i++ {
			temp[i] = int64(s[i])
		}
		return temp
	}
	return nil
}

func InSliceInt32(sl []int32, v int32) bool {
	for _, vv := range sl {
		if vv == v {
			return true
		}
	}
	return false
}

func InSliceInt32Slice(sl []int32, sub []int32) bool {
	for _, vv := range sub {
		if !InSliceInt32(sl, vv) {
			return false
		}
	}
	return true
}

func DelSliceInt32(sl []int32, v int32) []int32 {
	index := -1
	for key, value := range sl {
		if value == v {
			index = key
			break
		}
	}
	if index != -1 {
		sl = append(sl[:index], sl[index+1:]...)
	}
	return sl
}
func DelSliceInt64(sl []int64, v int64) []int64 {
	index := -1
	for key, value := range sl {
		if value == v {
			index = key
		}
	}
	if index != -1 {
		sl = append(sl[:index], sl[index+1:]...)
	}
	return sl
}
func DelSliceInt64s(cards []int64, sl []int64) []int64 {
	c := make([]int64, len(cards))
	s := make([]int64, len(sl))
	copy(c, cards)
	copy(s, sl)
	for i := 0; i < len(sl); i++ {
		for k, v := range c {
			isF := false
			for m, n := range s {
				if v == n {
					c = append(c[:k], c[k+1:]...)
					s = append(s[:m], s[m+1:]...)
					isF = true
					break
				}
			}
			if isF {
				break
			}
		}
	}
	return c
}
func DelSliceIn32s(cards []int32, sl []int32) []int32 {
	c := make([]int32, len(cards))
	s := make([]int32, len(sl))
	copy(c, cards)
	copy(s, sl)
	for i := 0; i < len(sl); i++ {
		for k, v := range c {
			isF := false
			for m, n := range s {
				if v == n {
					c = append(c[:k], c[k+1:]...)
					s = append(s[:m], s[m+1:]...)
					isF = true
					break
				}
			}
			if isF {
				break
			}
		}
	}
	return c
}
func DelSliceInt(sl []int, v int) []int {
	index := -1
	for key, value := range sl {
		if value == v {
			index = key
			break
		}
	}
	if index != -1 {
		sl = append(sl[:index], sl[index+1:]...)
	}
	return sl
}

func DelSliceString(sl []string, v string) ([]string, bool) {
	index := -1
	for key, value := range sl {
		if value == v {
			index = key
			break
		}
	}
	if index != -1 {
		sl = append(sl[:index], sl[index+1:]...)
	}
	return sl, index != -1
}

func InSliceInt32Index(sl []int32, v int32) int {
	for idx, vv := range sl {
		if vv == v {
			return idx
		}
	}
	return -1
}
func InSliceInt64Index(sl []int64, v int64) int {
	for idx, vv := range sl {
		if vv == v {
			return idx
		}
	}
	return -1
}
func IntSliceEqual(left []int, right []int) bool {
	if len(left) != len(right) {
		return false
	}
	for i := 0; i < len(left); i++ {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}
func Int32SliceEqual(left []int32, right []int32) bool {
	if len(left) != len(right) {
		return false
	}
	for i := 0; i < len(left); i++ {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func InSliceInt64(sl []int64, v int64) bool {
	for _, vv := range sl {
		if vv == v {
			return true
		}
	}
	return false
}

func InSliceInt(sl []int, v int) bool {
	for _, vv := range sl {
		if vv == v {
			return true
		}
	}
	return false
}

func InSliceString(sl []string, v string) bool {
	for _, vv := range sl {
		if vv == v {
			return true
		}
	}
	return false
}

func InSliceInterface(sl []interface{}, v interface{}) bool {
	for _, vv := range sl {
		if vv == v {
			return true
		}
	}
	return false
}

func InsertValueToSlice(index int, value int, arr []int) []int {
	arr[index] = value
	playerRandData := arr[index]
	arr = append(arr[:index], arr[index+1:]...)
	var coinRandDataSort = make([]int, 0, len(arr))
	isEnd := true
	for i := 0; i < len(arr); i++ {
		if arr[i] < playerRandData {
			coinRandDataSort = append(coinRandDataSort, arr[:i]...)
			coinRandDataSort = append(coinRandDataSort, playerRandData)
			coinRandDataSort = append(coinRandDataSort, arr[i:]...)
			isEnd = false
			break
		}
	}
	if isEnd {
		coinRandDataSort = append(arr, playerRandData)
	}
	return coinRandDataSort
}
func IsLadderSlice(sl []int) bool {
	sort.Ints(sl)
	switch len(sl) {
	case 0:
		return false
	case 1:
		return true
	default:
		for i := 0; i < len(sl)-1; i++ {
			if sl[i] < sl[i+1] {
				return false
			}
		}
		return true
	}
}

func IsSameSliceStr(dst []string, src []string) bool {
	if len(dst) != len(src) {
		return false
	}
	for i := 0; i < len(dst); i++ {
		if dst[i] != src[i] {
			return false
		}
	}
	return true
}
func StrSliceInt(str string, sep string) ([]int, error) {
	var err error
	t := strings.Split(str, sep)
	var ret []int
	for i := 0; i < len(t); i++ {
		if n, ok := strconv.Atoi(t[i]); ok == nil {
			ret = append(ret, n)
		} else {
			ret = append(ret, 0)
			err = ok
		}
	}
	return ret, err
}

func SliceIntEqual(dst []int, src []int) bool {
	if len(dst) != len(src) {
		return false
	}
	for i := 0; i < len(dst); i++ {
		if dst[i] != src[i] {
			return false
		}
	}
	return true
}
func SliceInt32Equal(dst []int32, src []int32) bool {
	if len(dst) != len(src) {
		return false
	}
	for i := 0; i < len(dst); i++ {
		if dst[i] != src[i] {
			return false
		}
	}
	return true
}
func SliceNoRepeate(data []int) []int {
	var newData []int
	for _, value := range data {
		repeate := false
		for _, n := range newData {
			if n == value {
				repeate = true
			}
		}
		if !repeate {
			newData = append(newData, value)
		}
	}
	return newData
}

func SliceInterfaceToInt32(data []interface{}) []int32 {
	cnt := len(data)
	if cnt > 0 {
		val := make([]int32, 0, cnt)
		for _, f := range data {
			val = append(val, int32(int32(f.(float64))))
		}
		return val
	}
	return nil
}
func SliceMaxValue(sl []int) int {
	maxValue := math.MinInt32
	for _, value := range sl {
		if value > maxValue {
			maxValue = value
		}
	}
	return maxValue
}
func SliceMinValue(sl []int) int {
	minValue := math.MaxInt32
	for _, value := range sl {
		if value < minValue {
			minValue = value
		}
	}
	return minValue
}
func SliceValueCount(sl []int, value int) int {
	var count int
	for _, v := range sl {
		if v == value {
			count++
		}
	}
	return count
}
func SliceValueWeight(sl []int, index int) float64 {
	if index < 0 || index > len(sl) {
		return 0
	}
	value := sl[index]
	totle := 0
	for _, v := range sl {
		totle += v
	}
	return float64(value) / float64(totle)
}

type Int32Slice []int32

func (p Int32Slice) Len() int           { return len(p) }
func (p Int32Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p Int32Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p Int32Slice) Sort()              { sort.Sort(p) }

type Int64Slice []int64

func (p Int64Slice) Len() int           { return len(p) }
func (p Int64Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p Int64Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p Int64Slice) Sort()              { sort.Sort(p) }

type Float64Slice []float64

func (p Float64Slice) Len() int           { return len(p) }
func (p Float64Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p Float64Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p Float64Slice) Sort()              { sort.Sort(p) }

func Int32SliceToString(a []int32, delim string) string {
	return strings.Trim(strings.Join(strings.Fields(fmt.Sprint(a)), delim), "[]")
}

func StringSliceToString(a []string, delim string) string {
	return strings.Trim(strings.Join(strings.Fields(fmt.Sprint(a)), delim), "[]")
}
