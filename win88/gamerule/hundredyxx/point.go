package hundredyxx

import (
	"math/rand"
	"time"
)

type DiceBox struct {
	first  *rand.Rand
	second *rand.Rand
	third  *rand.Rand
	count  int
	Point  [DICE_NUM]int32
}

var AllRoll = [][DICE_NUM]int32{}

func CreateDiceBox() *DiceBox {
	return &DiceBox{
		count:  0,
		first:  rand.New(rand.NewSource(rand.Int63n(time.Now().UnixNano()))),
		second: rand.New(rand.NewSource(rand.Int63n(time.Now().UnixNano()))),
		third:  rand.New(rand.NewSource(rand.Int63n(time.Now().UnixNano()))),
		Point:  [DICE_NUM]int32{},
	}
}
func (bb *DiceBox) Roll() {
	if bb.count > 10 {
		bb.first = rand.New(rand.NewSource(rand.Int63n(time.Now().UnixNano())))
		bb.second = rand.New(rand.NewSource(rand.Int63n(time.Now().UnixNano())))
		bb.third = rand.New(rand.NewSource(rand.Int63n(time.Now().UnixNano())))
		bb.count = 0
	}
	bb.count++
	bb.Point[0] = bb.first.Int31n(6)
	bb.Point[1] = bb.second.Int31n(6)
	bb.Point[2] = bb.third.Int31n(6)
}

func init() {
	for i := int32(0); i < 6; i++ {
		for j := int32(0); j < 6; j++ {
			for k := int32(0); k < 6; k++ {
				AllRoll = append(AllRoll, [DICE_NUM]int32{i, j, k})
			}
		}
	}
}