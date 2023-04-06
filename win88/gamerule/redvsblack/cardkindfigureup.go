package redvsblack

import (
	"games.yol.com/win88/common"
	"sort"
)

var CardsKindFigureUpSington = &CardsKindFigureUp{}

type handCardsInfo struct {
	cards    []int //考虑花色,从小到大排序,注意A的位置
	cards2   []int //不考虑花色,从小到大排序,注意A的位置
	numColor int   //花形数量
	numValue int   //牌型数量
	maxValue int   //最大单牌
	maxColor int   //最大单牌花色
}

type KindOfCard struct {
	Kind       int   //牌型
	maxValue   int   //最大单牌
	maxColor   int   //最大单牌花色
	orgCards   []int //原始牌
	nocCards   []int //去掉花色的牌
	OrderCards []int //排序后的牌
}

func (this *KindOfCard) TidyCards() {
	//	var order []int
	//	if this.Kind == CardsKind_RoyalStraight || this.Kind == CardsKind_RoyalFlushStraight {
	//		order = []int{this.nocCards[1], this.nocCards[2], this.nocCards[0]} //把A放后面
	//	} else {
	//		order = this.nocCards
	//	}
	//	orgCards := make([]int, Hand_CardNum, Hand_CardNum)
	//	copy(orgCards, this.orgCards)
	//	for _, c := range order {
	//		for i, oc := range orgCards {
	//			if c == oc%PER_CARD_COLOR_MAX {
	//				n := len(orgCards)
	//				this.OrderCards = append(this.OrderCards, oc)
	//				rest := orgCards[:i]
	//				if i < n-1 {
	//					rest = append(rest, orgCards[i+1:]...)
	//				}
	//				orgCards = rest
	//				break
	//			}
	//		}
	//	}
	this.OrderCards = this.orgCards
}

func CreateCardKind() *KindOfCard {
	return &KindOfCard{
		Kind:     CardsKind_Single,
		maxValue: 0,
		maxColor: 0,
	}
}

type CardsKindFigureUp struct {
}

func createHandCardsInfo(cards []int) *handCardsInfo {
	cardsInfo := &handCardsInfo{
		cards: make([]int, len(cards)),
	}
	copy(cardsInfo.cards, cards)
	//先整理下牌
	sort.Ints(cardsInfo.cards)
	colorMap := make(map[int]bool)
	kindMap := make(map[int]bool)
	maxValue := 1
	for i := 0; i < Hand_CardNum; i++ {
		c := cardsInfo.cards[i] / PER_CARD_COLOR_MAX
		k := cardsInfo.cards[i] % PER_CARD_COLOR_MAX
		colorMap[c] = true
		kindMap[k] = true
		cardsInfo.cards2 = append(cardsInfo.cards2, k)
		if CardValueMap[k] > CardValueMap[maxValue] {
			maxValue = k
			cardsInfo.maxValue = k
			cardsInfo.maxColor = c
		} else if CardValueMap[k] == CardValueMap[maxValue] {
			if c > cardsInfo.maxColor {
				cardsInfo.maxColor = c
			}
		}
	}
	sort.Ints(cardsInfo.cards2)
	cardsInfo.numColor = len(colorMap)
	cardsInfo.numValue = len(kindMap)
	return cardsInfo
}

func (this *CardsKindFigureUp) FigureUpByCard(cards []int) *KindOfCard {
	cardsInfo := createHandCardsInfo(cards)
	//处理检查的牌型
	cardKind := this.checkCardKind(cardsInfo)
	if cardKind != nil {
		cardKind.nocCards = cardsInfo.cards2
	} else {
		cardKind = &KindOfCard{
			Kind:     CardsKind_Single,
			maxValue: cardsInfo.maxValue,
			maxColor: cardsInfo.maxColor,
			nocCards: cardsInfo.cards2,
		}
	}
	cardKind.orgCards = cards

	return cardKind
}

func (this *CardsKindFigureUp) checkCardKind(cardsInfo *handCardsInfo) *KindOfCard {
	for i := CardsKind_Max - 1; i > 0; i-- {
		kindFunc := getCardCheckFunc(i)
		if kindFunc == nil {
			continue
		}
		koc := kindFunc(cardsInfo)
		if koc != nil {
			return koc
		}
	}
	return nil
}

func CompareCards(l, r *KindOfCard) int {
	if l.Kind == r.Kind { //牌型相同
		ret := compareCardsSameKindBySize(l, r)
		if ret == 0 {
			if l.maxColor > r.maxColor {
				return 1
			} else {
				return -1
			}
		} else {
			return ret
		}
	}
	if l.Kind > r.Kind {
		return 1
	} else {
		return -1
	}
}

//比大小：同牌型最大单牌大小相同的情况下，依次比较剩余牌的大小
func compareCardsSameKindBySize(l, r *KindOfCard) int {
	if l.nocCards[0] == A_CARD && r.nocCards[0] != A_CARD {
		return 1
	}
	if l.nocCards[0] != A_CARD && r.nocCards[0] == A_CARD {
		return -1
	}
	//对子比较(优先比对)
	if l.Kind == CardsKind_Double || l.Kind == CardsKind_BigDouble {
		if CardValueMap[l.nocCards[0]] > CardValueMap[r.nocCards[0]] {
			return 1
		} else if CardValueMap[l.nocCards[0]] < CardValueMap[r.nocCards[0]] {
			return -1
		}
	}
	for i := 2; i >= 0; i-- {
		if CardValueMap[l.nocCards[i]] > CardValueMap[r.nocCards[i]] {
			return 1
		} else if CardValueMap[l.nocCards[i]] < CardValueMap[r.nocCards[i]] {
			return -1
		}
	}
	//大小相等
	return 0
}

//比花色：同牌型最大单牌大小相同的情况下，最大单牌比较花色，黑桃>红桃>梅花>方片
func CompareCardsSameKindByColor(l, r *KindOfCard) int {
	if l.maxColor > r.maxColor {
		return 1
	} else {
		return -1
	}
}

//根据牌型创建相应的牌
func CreateCardByKind(cards [2][3]int, k int) []int {
	//先将两幅手牌放入一个数组中
	//在两幅手牌中找到花色相同的牌
	temp := make([]int, 0)
	for i := 0; i < 2; i++ {
		for j := 0; j < 3; j++ {
			temp = append(temp, cards[i][j])
		}
	}

	//根据不同的类型调用对应的生成函数
	if k == CardsKind_BigDouble {
		return CreateBigDouble(temp)
	} else if k == CardsKind_Straight {
		return CreateStraight(temp)
	} else if k == CardsKind_Flush {
		return CreateFlush(temp)
	} else if k == CardsKind_ThreeSame {
		return CreateThreeSame(temp)
	}
	return temp
}
func CreateBigDouble(temp []int) []int {
	//先找出牌面值大于8的任一一张牌
	sameCard := -1
	for idx, v := range temp {
		c := v % PER_CARD_COLOR_MAX
		if CardValueMap[c] > 8 {
			sameCard = v
			temp[0], temp[idx] = temp[idx], temp[0]
			break
		}
	}

	//现有手牌里查找是否有相同的
	for i := 1; i < len(temp); i++ {
		if temp[i]%PER_CARD_COLOR_MAX == sameCard%PER_CARD_COLOR_MAX {
			temp[1], temp[i] = temp[i], temp[1]
			return temp
		}
	}
	//没有相同的，创造一张相同的牌
	temp[1] = sameCard%PER_CARD_COLOR_MAX + common.RandInt(1, 4)*PER_CARD_COLOR_MAX
	if temp[1] > POKER_CART_CNT {
		temp[1] = temp[1] - POKER_CART_CNT
	}
	return temp
}
func CreateStraight(temp []int) []int {
	firstCard := 0
	//找出第一张不是K的牌
	for k, v := range temp {
		c := v % PER_CARD_COLOR_MAX
		if CardValueMap[c] < 12 {
			firstCard = c
			temp[0], temp[k] = temp[k], temp[0]
			break
		}
	}

	//从第二张开始查找，现有的牌能否凑成顺子，不能则创造一张
	n := 1
	for i := 1; i < len(temp); i++ {
		c := temp[i] % PER_CARD_COLOR_MAX
		if CardValueMap[c] == firstCard+1 {
			temp[n], temp[i] = temp[i], temp[n]
			firstCard++
			n++
		}
	}

	if n < 3 {
		for i := n; i < 3; i++ {
			temp[i] = (firstCard + 1) + common.RandInt(4)*PER_CARD_COLOR_MAX
			firstCard++
		}
	}

	return temp
}
func CreateFlush(temp []int) []int {
	colorMap := make(map[int]int)
	maxColor := 0
	for _, v := range temp {
		c := v / PER_CARD_COLOR_MAX
		colorMap[c] = colorMap[c] + 1
		if colorMap[c] > colorMap[maxColor] {
			maxColor = c
		}
	}

	//将maxColor花色的牌放最前边
	n := 0
	for i, v := range temp {
		c := v / PER_CARD_COLOR_MAX
		if c == maxColor {
			if i != n {
				temp[n], temp[i] = temp[i], temp[n]
			}
			n++
			if n == colorMap[maxColor] || n == 3 {
				break
			}
		}
	}

	//6张牌中至少有2张牌的花色一样，所以最多更换一张牌的花色
	if colorMap[maxColor] < 3 {
		for i := colorMap[maxColor]; i < len(temp); i++ {
			v := temp[i]
			isNext := false

			for j := 0; j < len(temp); j++ {
				if j != i && v%PER_CARD_COLOR_MAX == temp[j]%PER_CARD_COLOR_MAX {
					isNext = true
					break
				}
			}

			if !isNext {
				temp[i] = temp[i]%PER_CARD_COLOR_MAX + maxColor*PER_CARD_COLOR_MAX
				temp[colorMap[maxColor]], temp[i] = temp[i], temp[colorMap[maxColor]]
				break
			}
		}
	}
	return temp
}
func CreateThreeSame(temp []int) []int {
	valueMap := make(map[int]int)
	valueArray := make([]int, 0)
	sameCard := -1
	for _, v := range temp {
		c := v % PER_CARD_COLOR_MAX
		valueMap[c] = valueMap[c] + 1
		if valueMap[c] > valueMap[sameCard] {
			sameCard = c
		}
	}

	if valueMap[sameCard] > 1 {
		//将相同的牌放最前边
		n := 0
		for i, v := range temp {
			c := v % PER_CARD_COLOR_MAX
			if c == sameCard {
				valueArray = append(valueArray, v)
				if i != n {
					temp[n], temp[i] = temp[i], temp[n]
				}
				n++
				if n == valueMap[sameCard] || n == 3 {
					break
				}
			}
		}
	}

	if valueMap[sameCard] < 3 {
		for i := valueMap[sameCard]; i < 3; i++ { //创造豹子
			for j := 0; j < 4; j++ {
				card := j*PER_CARD_COLOR_MAX + sameCard
				if !common.InSliceInt(valueArray, card) {
					temp[i] = card
					valueArray = append(valueArray, card)
					break
				}
			}
		}
	}

	return temp
}
