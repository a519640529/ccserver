package common

import "sort"

type SortObjectSlice []SortObject

type SortObject struct {
	SortValue   int
	ObjectValue int
	RoomId      int
}

func (so SortObjectSlice) Len() int {
	return len(so)
}
func (so SortObjectSlice) Less(i, j int) bool {
	return so[i].SortValue > so[j].SortValue
}
func (so SortObjectSlice) Swap(i, j int) {
	so[i], so[j] = so[j], so[i]
}

func Sort(data SortObjectSlice) {
	sort.Sort(data)
}
