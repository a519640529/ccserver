package redvsblack

import (
	"sort"
)

const (
	CardsKind_Single             int = iota //散牌
	CardsKind_Double                        //小对子2~8的对
	CardsKind_BigDouble                     //大对子9~A的对
	CardsKind_StraightA23                   //顺子地龙
	CardsKind_Straight                      //顺子普通
	CardsKind_RoyalStraight                 //顺子天龙
	CardsKind_Flush                         //同花
	CardsKind_FlushStraightA23              //同花顺地龙
	CardsKind_FlushStraight                 //同花顺普通
	CardsKind_RoyalFlushStraight            //同花顺天龙
	CardsKind_ThreeSame                     //豹子
	CardsKind_Max
)

type CardKindCheckInteface func(cardsInfo *handCardsInfo) *KindOfCard

var cardKindCheckList [CardsKind_Max]CardKindCheckInteface

func registeCardCheckFunc(kind int, kindFunc CardKindCheckInteface) {
	cardKindCheckList[kind] = kindFunc
}
func getCardCheckFunc(kind int) CardKindCheckInteface {
	if kind > CardsKind_Single && kind < CardsKind_Max {
		return cardKindCheckList[kind]
	}
	return nil
}

func init() {
	//1.CardsKind_Double
	registeCardCheckFunc(CardsKind_Double, func(cardsInfo *handCardsInfo) *KindOfCard {
		if cardsInfo.numValue != 2 {
			return nil
		}
		//对子放到前头
		if cardsInfo.cards2[0] != cardsInfo.cards2[1] {
			cardsInfo.cards2[0], cardsInfo.cards2[1], cardsInfo.cards2[2] = cardsInfo.cards2[2], cardsInfo.cards2[1], cardsInfo.cards2[0]
		}
		if CardValueMap[cardsInfo.cards2[0]] < 8 {
			return &KindOfCard{
				Kind:     CardsKind_Double,
				maxValue: cardsInfo.maxValue,
				maxColor: cardsInfo.maxColor,
			}
		}
		return nil
	})
	//1.CardsKind_BigDouble
	registeCardCheckFunc(CardsKind_BigDouble, func(cardsInfo *handCardsInfo) *KindOfCard {
		if cardsInfo.numValue != 2 {
			return nil
		}
		//对子放到前头
		if cardsInfo.cards2[0] != cardsInfo.cards2[1] {
			cardsInfo.cards2[0], cardsInfo.cards2[1], cardsInfo.cards2[2] = cardsInfo.cards2[2], cardsInfo.cards2[1], cardsInfo.cards2[0]
		}
		if CardValueMap[cardsInfo.cards2[0]] >= 8 {
			return &KindOfCard{
				Kind:     CardsKind_BigDouble,
				maxValue: cardsInfo.maxValue,
				maxColor: cardsInfo.maxColor,
			}
		}
		return nil
	})
	//2.顺子地龙
	registeCardCheckFunc(CardsKind_StraightA23, func(cardsInfo *handCardsInfo) *KindOfCard {
		if cardsInfo.numValue != 3 || cardsInfo.numColor < 2 {
			return nil
		}
		if cardsInfo.cards2[0] == 0 && cardsInfo.cards2[1] == 1 && cardsInfo.cards2[2] == 2 {
			return &KindOfCard{
				Kind:     CardsKind_StraightA23,
				maxValue: cardsInfo.maxValue,
				maxColor: cardsInfo.maxColor,
			}
		} else {
			return nil
		}
	})
	//3.顺子
	registeCardCheckFunc(CardsKind_Straight, func(cardsInfo *handCardsInfo) *KindOfCard {
		if cardsInfo.numValue != 3 || cardsInfo.numColor < 2 {
			return nil
		}
		if IsRoundSlice(cardsInfo.cards2) && cardsInfo.cards2[0] != 0 {
			return &KindOfCard{
				Kind:     CardsKind_Straight,
				maxValue: cardsInfo.maxValue,
				maxColor: cardsInfo.maxColor,
			}
		}
		return nil
	})
	//4.皇家顺子
	registeCardCheckFunc(CardsKind_RoyalStraight, func(cardsInfo *handCardsInfo) *KindOfCard {
		if cardsInfo.numValue != 3 || cardsInfo.numColor < 2 {
			return nil
		}
		if cardsInfo.cards2[0] == 0 && cardsInfo.cards2[1] == 11 && cardsInfo.cards2[2] == 12 {
			return &KindOfCard{
				Kind:     CardsKind_RoyalStraight,
				maxValue: cardsInfo.maxValue,
				maxColor: cardsInfo.maxColor,
			}
		}
		return nil
	})
	//5.同花
	registeCardCheckFunc(CardsKind_Flush, func(cardsInfo *handCardsInfo) *KindOfCard {
		if cardsInfo.numColor != 1 {
			return nil
		}
		return &KindOfCard{
			Kind:     CardsKind_Flush,
			maxValue: cardsInfo.maxValue,
			maxColor: cardsInfo.maxColor,
		}
	})
	//6.同花顺地龙
	registeCardCheckFunc(CardsKind_FlushStraightA23, func(cardsInfo *handCardsInfo) *KindOfCard {
		if cardsInfo.numColor != 1 {
			return nil
		}
		if cardsInfo.numValue != 3 {
			return nil
		}
		if cardsInfo.cards2[0] == 0 && cardsInfo.cards2[1] == 1 && cardsInfo.cards2[2] == 2 {
			return &KindOfCard{
				Kind:     CardsKind_FlushStraightA23,
				maxValue: cardsInfo.cards2[2],
				maxColor: cardsInfo.maxColor,
			}
		}
		return nil
	})
	//7.同花顺
	registeCardCheckFunc(CardsKind_FlushStraight, func(cardsInfo *handCardsInfo) *KindOfCard {
		if cardsInfo.numColor != 1 {
			return nil
		}
		if cardsInfo.numValue != 3 {
			return nil
		}
		if IsRoundSlice(cardsInfo.cards2) && cardsInfo.cards2[0] != 0 {
			return &KindOfCard{
				Kind:     CardsKind_FlushStraight,
				maxValue: cardsInfo.maxValue,
				maxColor: cardsInfo.maxColor,
			}
		}
		return nil
	})

	//8.同花顺 QKA
	registeCardCheckFunc(CardsKind_RoyalFlushStraight, func(cardsInfo *handCardsInfo) *KindOfCard {
		if cardsInfo.numColor != 1 {
			return nil
		}
		if cardsInfo.numValue != 3 {
			return nil
		}
		if cardsInfo.cards2[0] == 0 && cardsInfo.cards2[1] == 11 && cardsInfo.cards2[2] == 12 {
			return &KindOfCard{
				Kind:     CardsKind_RoyalFlushStraight,
				maxValue: cardsInfo.maxValue,
				maxColor: cardsInfo.maxColor,
			}
		} else {
			return nil
		}
	})
	//9.豹子
	registeCardCheckFunc(CardsKind_ThreeSame, func(cardsInfo *handCardsInfo) *KindOfCard {
		if cardsInfo.numValue != 1 {
			return nil
		}
		return &KindOfCard{
			Kind:     CardsKind_ThreeSame,
			maxValue: cardsInfo.maxValue,
			maxColor: cardsInfo.maxColor,
		}
	})
}

/*
 * 查看一个切片是否是递增数组
 */
func IsRoundSlice(sl []int) bool {
	sort.Ints(sl)
	if len(sl) != 3 {
		return false
	}
	for i := 0; i < len(sl)-1; i++ {
		if sl[i]+1 != sl[i+1] {
			return false
		}
	}
	return true
}

/*
 * 在一个切片中查找给定的数值是否存在
 */
func InSlice(value int, sl []int) bool {
	for _, v := range sl {
		if v == value {
			return true
		}
	}
	return false
}
