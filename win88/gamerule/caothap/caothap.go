package caothap

import (
	"math/rand"
	"time"
)

func CardInit() []int32 {
	return cardDataShuffle(cardData[:])
}

func cardDataShuffle(cardData []int32) []int32 {
	rand.Seed(time.Now().UnixNano())
	num := len(cardData)
	for i := 0; i < num; i++ {
		n := rand.Intn(num - i)
		cardData[i], cardData[n] = cardData[n], cardData[i]
	}
	return cardData
}

func IsCardA(card int32) bool {
	return getCardValue(card) == 12
}
func getCardValue(card int32) int {
	return int(card) % 13
}

// 提供外部接口
func GetCardValue(card int32) int {
	return getCardValue(card)
}
