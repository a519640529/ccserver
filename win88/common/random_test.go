package common

import (
	"testing"
)

func TestRandom(t *testing.T) {
	r := &RandomGenerator{}
	r.RandomSeed(1000)
	for i := 0; i < 100; i++ {
		t.Log(r.Rand32(100))
	}
}
func TestRandMaxMiddle(t *testing.T) {
	testData := [][]int{
		{0},
		{0, 1},
		{0, 1, 2},
		{0, 1, 2, 3},
	}
	for _, value := range testData {
		index := RandMaxMiddle(len(value))
		if index >= len(value) || index < 0 {
			t.Error("Data", value)
			t.Fatal("Error check in TestRandMaxMiddle")
		}
	}
}
func TestRandLastMiddle(t *testing.T) {
	testData := [][]int{
		{0},
		{0, 1},
		{0, 1, 2},
		{0, 1, 2, 3},
	}
	for _, value := range testData {
		index := RandMaxMiddle(len(value))
		if index >= len(value) || index < 0 {
			t.Error("Data", value)
			t.Fatal("Error check in TestRandLastMiddle")
		}
	}
}
