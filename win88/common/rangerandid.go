package common

import (
	"math/rand"
	"time"
)

const (
	MAX_TRY_RAND_CNT = 10
	RANDID_INVALID   = -1
)

type BitPool struct {
	flag []uint64
	cnt  int
}

func (bp *BitPool) Init(cnt int) {
	bp.cnt = cnt/64 + 1
	bp.flag = make([]uint64, bp.cnt, bp.cnt)
}

func (bp *BitPool) Mark(n int) bool {
	idx := n / 64
	off := n % 64
	if idx < 0 || idx >= bp.cnt {
		return false
	}

	bp.flag[idx] |= 1 << uint(off)
	return true
}

func (bp *BitPool) Unmark(n int) bool {
	idx := n / 64
	off := n % 64
	if idx < 0 || idx >= bp.cnt {
		return false
	}

	bp.flag[idx] &= ^(1 << uint(off))
	return true
}

func (bp *BitPool) IsMark(n int) bool {
	idx := n / 64
	off := n % 64
	if idx < 0 || idx >= bp.cnt {
		return false
	}

	return (bp.flag[idx] & (1 << uint(off))) != 0
}

type RandDistinctId struct {
	rand *rand.Rand
	bp   BitPool
	min  int
	max  int
	cap  int
}

func NewRandDistinctId(min, max int) *RandDistinctId {
	var id RandDistinctId
	id.Init(min, max)
	return &id
}

func (rid *RandDistinctId) Init(min, max int) {
	rid.min = min
	rid.max = max
	rid.cap = max - min
	rid.bp.Init(rid.cap)
	rid.rand = rand.New(rand.NewSource(time.Now().UnixNano()))
}

func (rid *RandDistinctId) Alloc(id int) bool {
	if !rid.bp.IsMark(id) {
		rid.bp.Mark(id)
		return true
	}
	return false
}

func (rid *RandDistinctId) Free(id int) bool {
	if rid.bp.IsMark(id) {
		rid.bp.Unmark(id)
		return true
	}
	return false
}

func (rid *RandDistinctId) RandOne() int {
	for i := 0; i < MAX_TRY_RAND_CNT; i++ {
		pos := rid.rand.Intn(rid.cap)
		if !rid.bp.IsMark(pos) {
			rid.bp.Mark(pos)
			return pos + rid.min
		}
	}

	//try one by one
	rpos := rid.rand.Intn(rid.cap)
	if rpos < rid.cap/2 {
		for i := rpos; i >= 0; i-- {
			if !rid.bp.IsMark(i) {
				rid.bp.Mark(i)
				return i + rid.min
			}
		}
	} else {
		for i := rpos; i < rid.cap; i++ {
			if !rid.bp.IsMark(i) {
				rid.bp.Mark(i)
				return i + rid.min
			}
		}
	}
	return RANDID_INVALID
}

//func memConsumed() uint64 {
//	runtime.GC() //GC，排除对象影响
//	var memStat runtime.MemStats
//	runtime.ReadMemStats(&memStat)
//	return memStat.Sys
//}

//func main() {
//	var idPool RandDistinctId
//	before := memConsumed()
//	idPool.Init(MIN, MAX)
//	after := memConsumed()
//
//	idMap := make(map[int]bool)
//	for i := 0; i < 1000000; i++ {
//		id := idPool.RandId()
//		if id != -1 {
//			idMap[id] = true
//		}
//	}
//
//	start := time.Now()
//	for i := 0; i < 10000; i++ {
//		id := idPool.RandId()
//		if id != -1 {
//			fmt.Println(i, "->", id)
//			idMap[id] = true
//		}
//
//	}
//
//	fmt.Println("total take:", time.Now().Sub(start), " len:", len(idMap))
//	fmt.Println(fmt.Sprintf("new %v pool consume %.3f MB", MAX-MIN, float64(after-before)/1024/1024))
//}
