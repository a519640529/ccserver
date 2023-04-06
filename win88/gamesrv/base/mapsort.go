package base

import "sort"

type Pair struct {
	Key   int32
	Value int64
}

//map对Value排序-升序
type PairList []Pair

func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p PairList) Len() int           { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].Value < p[j].Value }
func SortMapByValue(m map[int32]int64) PairList {
	p := make(PairList, len(m))
	for k, v := range m {
		p[k] = Pair{k, v}
	}
	sort.Sort(p)
	return p
}

//map对Value排序-降序
type UnPairList []Pair

func (p UnPairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p UnPairList) Len() int           { return len(p) }
func (p UnPairList) Less(i, j int) bool { return p[i].Value > p[j].Value }
func sortUnMapByValue(m map[int32]int64) UnPairList {
	p := make(UnPairList, len(m))
	for k, v := range m {
		p[k] = Pair{k, v}
	}
	sort.Sort(p)
	return p
}
