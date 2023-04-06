package minipoker

import (
	"math/rand"
	"sort"
	"time"
)

type Int32Slice []int32

func (a Int32Slice) Len() int           { return len(a) }
func (a Int32Slice) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a Int32Slice) Less(i, j int) bool { return a[i] < a[j] }

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

func GetCards() []int32 {
	cardsData := cardDataShuffle(cardData[:]) // 洗牌
	// 随机开始区牌索引
	pos := rand.Intn(CARDDATANUM - CARDNUM)
	return cardsData[pos : pos+CARDNUM]
}
func GetCardsName(cards []int32) string {
	var s string
	for _, v := range cards {
		s += cardName[int(v)]
	}
	return s
}

func GetWinRate(cardsType int) int32 {
	return int32(cardsTypeRate[cardsType-1])
}
func CheckCardsType(cardsType int) bool {
	if cardsType < CARDTYPE_MIN || cardsType > CARDTYPE_MAX {
		return false
	}
	if cardsType != CARDTYPE_STRAIGHT_FLUSH &&
		cardsType != CARDTYPE_FOURCARD && cardsType != CARDTYPE_THREETAKEPAIR && cardsType != CARDTYPE_FLUSH {
		return true
	}
	return false
}

// 计算牌型分
func CalcCardsTypeScore(betValue int64, cardsType int) int64 {
	return betValue * int64(cardsTypeRate[cardsType-1]) / 10
}

// 计算牌型
func CalcCardsType(cards []int32) int {
	if len(cards) != CARDNUM {
		return -1
	}
	cardsTemp := make([]int32, len(cards))
	copy(cardsTemp, cards)
	// 同花顺
	if isStraightFlush(cardsTemp) {
		if isStraightFlushJ(cardsTemp) {
			return CARDTYPE_STRAIGHT_FLUSH_J
		}
		return CARDTYPE_STRAIGHT_FLUSH
	}
	// 四张
	if isFourCard(cardsTemp) {
		return CARDTYPE_FOURCARD
	}
	// 三带对
	if isThreeCard(cardsTemp) {
		cardsDataTemp := make([]int32, len(cards))
		copy(cardsDataTemp, cards)
		rmdCard := rmThreeCard(cardsDataTemp)
		if isHavePair(rmdCard) {
			return CARDTYPE_THREETAKEPAIR
		}
	}
	// 同花
	if isFlush(cardsTemp) {
		return CARDTYPE_FLUSH
	}
	// 顺子
	if isStraight(cardsTemp) {
		return CARDTYPE_STRAIGHT
	}
	// 三带单
	if isThreeCard(cardsTemp) {
		cardsDataTemp := make([]int32, len(cards))
		copy(cardsDataTemp, cards)
		rmdCard := rmThreeCard(cardsDataTemp)
		if !isHavePair(rmdCard) {
			return CARDTYPE_THREETAKESINGLE
		}
	}
	// 两对
	if isHavePair(cardsTemp) {
		cardsDataTemp := make([]int32, len(cards))
		copy(cardsDataTemp, cards)
		rmdCard := rmPair(cardsDataTemp)
		if isHavePair(rmdCard) {
			return CARDTYPE_TWOPAIR
		}
	}
	// 一对
	if isHavePair(cardsTemp) {
		if isPairJ(cardsTemp) {
			return CARDTYPE_ONEPAIR_J
		}
		return CARDTYPE_ONEPAIR
	}
	// 散牌
	if isHighCard(cardsTemp) {
		return CARDTYPE_HIGHCARD
	}
	return -1
}

// 同花顺 J
func isStraightFlushJ(cards []int32) bool {
	if len(cards) != CARDNUM {
		return false
	}
	cardsValue := getCardsValue(cards)
	sort.Sort(Int32Slice(cardsValue))
	if cardsValue[1] > 6 {
		return true
	}
	return false
}

// 同花顺
func isStraightFlush(cards []int32) bool {
	if len(cards) != CARDNUM {
		return false
	}
	if isStraight(cards) && isFlush(cards) {
		return true
	}
	return false
}

// 顺子
func isStraight(cards []int32) bool {
	if len(cards) != CARDNUM {
		return false
	}
	cardsValue := getCardsValue(cards)
	sort.Sort(Int32Slice(cardsValue))
	cv := cardsValue[0]

	for i := 1; i < len(cardsValue); i++ {
		//// A2345,A10JQK
		//		//if i == len(cardsValue)-1 && cardsValue[i] == 12 && (cardsValue[0] == 0 || cardsValue[len(cardsValue)-1] == 11) {
		//		//	return true
		//		//}
		//		//if cardsValue[i] != cv+1 {
		//		//	break
		//		//}
		//		//if i == len(cardsValue)-1 {
		//		//	return true
		//		//}
		//		//cv = cardsValue[i]

		if cardsValue[i] != cv+1 {
			if cardsValue[0] == 0 && i == 1 && cardsValue[1] == 9 { //A10JQK
				continue
			}
			return false
		}

		cv = cardsValue[i]
		return true
	}
	return false
}

// 同花
func isFlush(cards []int32) bool {
	if len(cards) != CARDNUM {
		return false
	}
	cardsType := getCardsType(cards)
	ct := cardsType[0]
	for i := 1; i < len(cardsType); i++ {
		if ct != cardsType[i] {
			return false
		}
	}
	return true
}

// 四张
func isFourCard(cards []int32) bool {
	if len(cards) != CARDNUM {
		return false
	}
	cardsValue := getCardsValue(cards)
	cvMap := make(map[int32]int, len(cardsValue))
	for _, v := range cardsValue {
		if _, ok := cvMap[v]; !ok {
			cvMap[v] = 0
		}
		cvMap[v]++
		if cvMap[v] >= 4 {
			return true
		}
	}
	return false
}

// 三张
func isThreeCard(cards []int32) bool {
	if len(cards) != CARDNUM {
		return false
	}
	cardsValue := getCardsValue(cards)
	cvMap := make(map[int32]int, len(cardsValue))
	for _, v := range cardsValue {
		if _, ok := cvMap[v]; !ok {
			cvMap[v] = 0
		}
		cvMap[v]++
		if cvMap[v] >= 3 {
			return true
		}
	}
	return false
}

// 去除三张
func rmThreeCard(cards []int32) []int32 {
	if len(cards) != CARDNUM {
		return nil
	}
	cardsTemp := make([]int32, len(cards))
	copy(cardsTemp, cards)

	cardsValue := getCardsValue(cardsTemp)
	cvMap := make(map[int32]int, len(cardsValue))
	for _, v := range cardsValue {
		if _, ok := cvMap[v]; !ok {
			cvMap[v] = 0
		}
		cvMap[v]++
	}

	var rmCardValue int32
	for c, n := range cvMap {
		if n == 3 {
			rmCardValue = c
		}
	}
	for i := 0; i < len(cardsTemp); i++ {
		if int32(getCardValue(cardsTemp[i])) == rmCardValue {
			cardsTemp = append(cardsTemp[:i], cardsTemp[i+1:]...)
			if i > 0 {
				i--
			}
		}
	}
	return cardsTemp
}

// 去除对子
func rmPair(cards []int32) []int32 {
	if len(cards) != CARDNUM {
		return nil
	}
	cardsTemp := make([]int32, len(cards))
	copy(cardsTemp, cards)

	cardsValue := getCardsValue(cardsTemp)
	cvMap := make(map[int32]int, len(cardsValue))
	for _, v := range cardsValue {
		if _, ok := cvMap[v]; !ok {
			cvMap[v] = 0
		}
		cvMap[v]++
	}

	var rmCardValue int32
	for c, n := range cvMap {
		if n == 2 {
			rmCardValue = c
		}
	}
	for i := 0; i < len(cardsTemp); i++ {
		if int32(getCardValue(cardsTemp[i])) == rmCardValue {
			cardsTemp = append(cardsTemp[:i], cardsTemp[i+1:]...)
			if i > 0 {
				i--
			}
		}
	}
	return cardsTemp
}

// 对子
func isHavePair(cards []int32) bool {
	cardsValue := getCardsValue(cards)
	cvMap := make(map[int32]int, len(cardsValue))
	for _, v := range cardsValue {
		if _, ok := cvMap[v]; !ok {
			cvMap[v] = 0
		}
		cvMap[v]++
		if cvMap[v] >= 2 {
			return true
		}
	}
	return false
}

// 对子（J、Q、K、A）
func isPairJ(cards []int32) bool {
	cardsValue := getCardsValue(cards)
	sort.Sort(Int32Slice(cardsValue))
	for i := 1; i < len(cardsValue); i++ {
		if cardsValue[i-1] == cardsValue[i] && (cardsValue[i] > 9 || cardsValue[i] == 0) {
			return true
		}
	}
	return false
}

// 散牌
func isHighCard(cards []int32) bool {
	if !isFlush(cards) && !isStraight(cards) && !isFourCard(cards) && !isThreeCard(cards) && !isHavePair(cards) {
		return true
	}
	return false
}

// 获取一对的值和数量 (测试跑数据使用,日后可删除)
func GetCardsPair(cards []int32, cardsMap map[int32]int32) {
	cardsValue := getCardsValue(cards)
	cvMap := make(map[int32]int, len(cardsValue))
	for _, v := range cardsValue {
		if _, ok := cvMap[v]; !ok {
			cvMap[v] = 0
		}
		cvMap[v]++
		if cvMap[v] == 2 {
			if _, ok := cardsMap[v]; !ok {
				cardsMap[v] = 0
				continue
			}
			cardsMap[v]++
		}
	}
	return
}

func getCardsValue(cards []int32) []int32 {
	cardsTemp := make([]int32, len(cards))
	copy(cardsTemp, cards)
	for i := 0; i < len(cardsTemp); i++ {
		cardsTemp[i] = int32(getCardValue(cardsTemp[i]))
	}
	return cardsTemp
}

func getCardsType(cards []int32) []int32 {
	cardsTemp := make([]int32, len(cards))
	copy(cardsTemp, cards)
	for i := 0; i < len(cardsTemp); i++ {
		cardsTemp[i] = int32(getCardType(cardsTemp[i]))
	}
	return cardsTemp
}

func getCardType(card int32) int {
	return int(card) / 13
}
func getCardValue(card int32) int {
	return int(card) % 13
}
