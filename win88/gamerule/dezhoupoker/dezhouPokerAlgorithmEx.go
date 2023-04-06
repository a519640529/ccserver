package dezhoupoker

import (
	"sort"
)

//----------------------------------------------------------------------------------------------------------------------------------------------------------------
type KindOfCardFigureUpEx struct {
}

var KindOfCardFigureUpExSington = &KindOfCardFigureUpEx{}

func (this *KindOfCardFigureUpEx) FigureUpByCard(cards []int32) *CardsInfo {
	if len(cards) == 0 {
		return nil
	}
	var tempCard []int
	for _, v := range cards {
		tempCard = append(tempCard, int(v))
	}

	//按照升序排序
	sort.Ints(tempCard)

	var cardDataManager CardDataManager
	cardDataManager.Init()
	for i := 0; i < len(tempCard); i++ {
		cardDataManager.AddCard(int32(tempCard[i]))
	}
	cardDataManager.ReCal()

	//根据value 排序
	cardsInfo := this.CalCardKind(&cardDataManager)
	if cardsInfo != nil && len(cards) > 5 {
		cardsInfo.CalValue()
	}
	return cardsInfo
}

func (this *KindOfCardFigureUpEx) CalCardKind(cardDataManager *CardDataManager) *CardsInfo {

	card_info := this.IsRoyalFlush(cardDataManager)
	if card_info != nil {
		return card_info
	}
	card_info = this.IsStraightFlush(cardDataManager)
	if card_info != nil {
		return card_info
	}

	card_info = this.IsFourKind(cardDataManager)
	if card_info != nil {
		return card_info
	}

	card_info = this.IsFullhouse(cardDataManager)
	if card_info != nil {
		return card_info
	}

	card_info = this.IsFlush(cardDataManager)
	if card_info != nil {
		return card_info
	}

	card_info = this.IsStraight(cardDataManager)
	if card_info != nil {
		return card_info
	}

	card_info = this.IsThreeKind(cardDataManager)
	if card_info != nil {
		return card_info
	}

	card_info = this.IsTwoPair(cardDataManager)
	if card_info != nil {
		return card_info
	}

	card_info = this.IsOnePair(cardDataManager)
	if card_info != nil {
		return card_info
	}

	card_info = this.IsHighCard(cardDataManager)
	if card_info != nil {
		return card_info
	}

	return nil
}

//皇家同花顺  KindOfCard_RoyalFlush
func (this *KindOfCardFigureUpEx) IsRoyalFlush(cardDataManager *CardDataManager) *CardsInfo {

	cardCount := len(cardDataManager.CardDataPool)
	if cardCount == 0 {
		return nil
	}

	var cur_poker CardData
	cur_poker.Init(cardDataManager.CardDataPool[cardCount-1].Card)
	var rst_kind_card []int32
	rst_kind_card = append(rst_kind_card, cur_poker.Card)

	for i := cardCount - 2; i >= 0; i-- {
		if cardDataManager.CardDataPool[i].Card != cur_poker.Card-1 ||
			cardDataManager.CardDataPool[i].Color != cur_poker.Color {
			rst_kind_card = nil

			cur_poker.Init(cardDataManager.CardDataPool[i].Card)
			rst_kind_card = append(rst_kind_card, cur_poker.Card)
			continue
		}

		cur_poker.Init(cardDataManager.CardDataPool[i].Card)
		rst_kind_card = append(rst_kind_card, cur_poker.Card)

		if len(rst_kind_card) >= 4 {
			break
		}
	}

	if len(rst_kind_card) < 4 {
		return nil
	}

	if cur_poker.Value != POKER_10 {
		return nil
	}
	//已经找到10，J,Q,K, 找同色 A
	for i := 0; i < cardCount; i++ {
		if cardDataManager.CardDataPool[i].Value == POKER_A && cardDataManager.CardDataPool[i].Color == cur_poker.Color {
			cardInfo := &CardsInfo{
				Kind: KindOfCard_RoyalFlush,
			}
			//牌型对应的牌
			//A
			cardInfo.KindCards = append(cardInfo.KindCards, cardDataManager.CardDataPool[i].Card)
			//KQJ 10
			for i := 0; i < len(rst_kind_card); i++ {
				cardInfo.KindCards = append(cardInfo.KindCards, rst_kind_card[i])
			}
			return cardInfo
		}
	}

	return nil
}

//同花顺  KindOfCard_StraightFlush
func (this *KindOfCardFigureUpEx) IsStraightFlush(cardDataManager *CardDataManager) *CardsInfo {
	cardCount := len(cardDataManager.CardDataPool)
	if cardCount == 0 {
		return nil
	}

	var cur_poker CardData
	cur_poker.Init(cardDataManager.CardDataPool[cardCount-1].Card)
	var rst_kind_card []int32
	rst_kind_card = append(rst_kind_card, cur_poker.Card)

	for i := cardCount - 2; i >= 0; i-- {
		if cardDataManager.CardDataPool[i].Card != cur_poker.Card-1 || cur_poker.Color != cardDataManager.CardDataPool[i].Card/PER_CARD_COLOR_MAX {
			rst_kind_card = nil

			cur_poker.Init(cardDataManager.CardDataPool[i].Card)
			rst_kind_card = append(rst_kind_card, cur_poker.Card)
			continue
		}

		cur_poker.Init(cardDataManager.CardDataPool[i].Card)

		rst_kind_card = append(rst_kind_card, cur_poker.Card)

		if len(rst_kind_card) >= 5 {
			break
		}
	}

	if len(rst_kind_card) == 5 {
		cardInfo := &CardsInfo{
			Kind: KindOfCard_StraightFlush,
		}

		//牌型对应的牌
		for i := 0; i < len(rst_kind_card); i++ {
			cardInfo.KindCards = append(cardInfo.KindCards, rst_kind_card[i])
		}
		return cardInfo
	} else {
		return nil
	}
}

//四条：先比四条，四条相同比单张。 KindOfCard_FourKind
func (this *KindOfCardFigureUpEx) IsFourKind(cardDataManager *CardDataManager) *CardsInfo {
	cardCount := len(cardDataManager.CardDataPool)
	if cardCount < 4 {
		return nil
	}

	for typeCardValue, typeCardCount := range cardDataManager.CardValueCount {
		if typeCardCount == 4 {
			cardInfo := &CardsInfo{
				Kind: KindOfCard_FourKind,
			}

			//4张 牌型牌
			for i := cardCount - 1; i >= 0; i-- {
				curCard := cardDataManager.CardData2Pool[i]

				if curCard.Value == typeCardValue {
					//牌型牌
					cardInfo.KindCards = append(cardInfo.KindCards, curCard.Card)
				}
			}

			if cardCount >= 5 {
				//第5张牌型牌
				fithCard := int32(0)
				for i := cardCount - 1; i >= 0; i-- {
					curCard := cardDataManager.CardData2Pool[i]

					if curCard.Card != cardInfo.KindCards[0] {
						if curCard.Value == POKER_A {
							fithCard = curCard.Card
							break
						} else {
							if fithCard < curCard.Value {
								fithCard = curCard.Card
							}
						}
					}
				}
				cardInfo.KindCards = append(cardInfo.KindCards, fithCard)
			}
			return cardInfo
		}
	}
	return nil
}

//葫芦：先比三条，三条相同比对子 KindOfCard_Fullhouse
func (this *KindOfCardFigureUpEx) IsFullhouse(cardDataManager *CardDataManager) *CardsInfo {
	cardCount := len(cardDataManager.CardDataPool)
	if cardCount < 5 {
		return nil
	}

	card2Value := INVALIDE_CARD
	card3Value := INVALIDE_CARD
	card3Count := int32(0)

	for cardValue, cardCount := range cardDataManager.CardValueCount {
		if cardCount == 3 {
			card3Count++
			if card3Value == POKER_A {
				continue
			}

			if card3Value < cardValue || cardValue == POKER_A {
				card3Value = cardValue
			}
		} else if cardCount == 2 {
			if card2Value == POKER_A {
				continue
			}

			if card2Value < cardValue || cardValue == POKER_A {
				card2Value = cardValue
			}
		}
	}

	//总共7张，如果有两个3条，就必然不会有2对.如果有两个3条，把较小的哪个当做两对
	if card3Count >= 2 {
		for cardValue, cardCount := range cardDataManager.CardValueCount {
			if cardCount == 3 {
				if card3Value != cardValue {
					card2Value = cardValue
					break
				}
			}
		}
	}

	if card2Value != INVALIDE_CARD && card3Value != INVALIDE_CARD {
		cardInfo := &CardsInfo{
			Kind: KindOfCard_Fullhouse,
		}

		//3对
		for i := 0; i < cardCount; i++ {
			curCard := cardDataManager.CardDataPool[i]

			if curCard.Value == card3Value {
				//牌型牌
				cardInfo.KindCards = append(cardInfo.KindCards, curCard.Card)
			}
		}

		//2对
		for i := 0; i < cardCount; i++ {
			curCard := cardDataManager.CardDataPool[i]

			if curCard.Value == card2Value {
				//牌型牌
				cardInfo.KindCards = append(cardInfo.KindCards, curCard.Card)
			}
		}

		return cardInfo
	}

	return nil
}

//同花：比最大的单张，如相同则依次比剩余的单张。 KindOfCard_Flush
func (this *KindOfCardFigureUpEx) IsFlush(cardDataManager *CardDataManager) *CardsInfo {
	cardCount := len(cardDataManager.CardDataPool)
	if cardCount < 5 {
		return nil
	}

	sameColorCount := int32(0)
	sameColorValue := int32(0)

	for colorValue, colorCount := range cardDataManager.CardColorCount {
		if colorCount > sameColorCount {
			sameColorCount = colorCount
			sameColorValue = colorValue
		}
	}

	if sameColorCount < 5 {
		return nil
	}

	cardInfo := &CardsInfo{
		Kind: KindOfCard_Flush,
	}

	bHasPoker_A := false
	var temp_rst []int
	for i := cardCount - 1; i >= 0; i-- {
		curCard := cardDataManager.CardDataPool[i]

		if curCard.Color == sameColorValue {
			temp_rst = append(temp_rst, int(curCard.Card))

			if curCard.Value == POKER_A {
				bHasPoker_A = true
			}
		}
	}
	sort.Ints(temp_rst)

	if bHasPoker_A {
		cardInfo.KindCards = append(cardInfo.KindCards, int32(temp_rst[0]))

		data_len := len(temp_rst)
		for i := data_len - 1; i >= 0; i-- {
			cardInfo.KindCards = append(cardInfo.KindCards, int32(temp_rst[i]))
			if len(cardInfo.KindCards) >= 5 {
				break
			}
		}
	} else {
		data_len := len(temp_rst)
		for i := data_len - 1; i >= 0; i-- {
			cardInfo.KindCards = append(cardInfo.KindCards, int32(temp_rst[i]))
			if len(cardInfo.KindCards) >= 5 {
				break
			}
		}
	}

	return cardInfo
}

//顺子：比顺子的大小。A2345是最小的顺子,AKQJ10是最大的顺子。 KindOfCard_Straight
func (this *KindOfCardFigureUpEx) IsStraight(cardDataManager *CardDataManager) *CardsInfo {
	card_info := this.IsStraightNormal(cardDataManager)
	if card_info == nil {
		card_info = this.IsStraightMax(cardDataManager)
	}
	return card_info
}
func (this *KindOfCardFigureUpEx) IsStraightNormal(cardDataManager *CardDataManager) *CardsInfo {
	cardCount := len(cardDataManager.CardDataPool)
	if cardCount < 5 {
		return nil
	}

	var cur_poker CardData
	cur_poker.Init(cardDataManager.CardData2Pool[cardCount-1].Card)
	var rst_kind_card []int32
	rst_kind_card = append(rst_kind_card, cur_poker.Card)

	for i := cardCount - 2; i >= 0; i-- {
		if cardDataManager.CardData2Pool[i].Value == cur_poker.Value {
			continue
		}
		if cardDataManager.CardData2Pool[i].Value != cur_poker.Value-1 {
			rst_kind_card = nil

			cur_poker.Init(cardDataManager.CardData2Pool[i].Card)
			rst_kind_card = append(rst_kind_card, cur_poker.Card)
			continue
		}

		cur_poker.Init(cardDataManager.CardData2Pool[i].Card)

		rst_kind_card = append(rst_kind_card, cur_poker.Card)

		if len(rst_kind_card) >= 5 {
			break
		}
	}

	if len(rst_kind_card) == 5 {
		cardInfo := &CardsInfo{
			Kind: KindOfCard_Straight,
		}
		cardInfo.KindCards = rst_kind_card
		return cardInfo
	} else {
		return nil
	}
}
func (this *KindOfCardFigureUpEx) IsStraightMax(cardDataManager *CardDataManager) *CardsInfo {
	cardCount := len(cardDataManager.CardDataPool)
	if cardCount == 0 {
		return nil
	}

	var cur_poker CardData
	cur_poker.Init(cardDataManager.CardData2Pool[cardCount-1].Card)
	var rst_kind_card []int32
	rst_kind_card = append(rst_kind_card, cur_poker.Card)

	for i := cardCount - 2; i >= 0; i-- {
		if cardDataManager.CardData2Pool[i].Value == cur_poker.Value {
			continue
		}
		if cardDataManager.CardData2Pool[i].Value != cur_poker.Value-1 {
			rst_kind_card = nil

			cur_poker.Init(cardDataManager.CardData2Pool[i].Card)
			rst_kind_card = append(rst_kind_card, cur_poker.Card)
			continue
		}

		cur_poker.Init(cardDataManager.CardData2Pool[i].Card)

		rst_kind_card = append(rst_kind_card, cur_poker.Card)

		if len(rst_kind_card) >= 4 {
			break
		}
	}
	if len(rst_kind_card) < 4 {
		return nil
	}

	if cur_poker.Value == POKER_10 && cardDataManager.CardData2Pool[0].Value == POKER_A {
		cardInfo := &CardsInfo{
			Kind: KindOfCard_Straight,
		}
		cardInfo.KindCards = append(cardInfo.KindCards, cardDataManager.CardData2Pool[0].Card)
		cardInfo.KindCards = append(cardInfo.KindCards, rst_kind_card...)
		return cardInfo
	} else {
		return nil
	}
}

//三条：先比三条，三条相同比单张 KindOfCard_ThreeKind
func (this *KindOfCardFigureUpEx) IsThreeKind(cardDataManager *CardDataManager) *CardsInfo {
	cardCount := len(cardDataManager.CardDataPool)
	if cardCount < 3 {
		return nil
	}

	santiaoValue := int32(INVALIDE_CARD)
	for cardValue, cardCount := range cardDataManager.CardValueCount {
		if cardCount == 3 {
			santiaoValue = cardValue
		}
	}
	if santiaoValue == INVALIDE_CARD {
		return nil
	}
	cardInfo := &CardsInfo{
		Kind: KindOfCard_ThreeKind,
	}

	//前3张
	for i := cardCount - 1; i >= 0; i-- {
		curCard := cardDataManager.CardDataPool[i]

		if curCard.Value == santiaoValue {
			//牌型牌
			cardInfo.KindCards = append(cardInfo.KindCards, curCard.Card)
		}
	}

	if cardDataManager.CardData2Pool[0].Value == POKER_A && santiaoValue != POKER_A {
		if cardCount >= 4 {
			//第4张牌
			cardInfo.KindCards = append(cardInfo.KindCards, cardDataManager.CardData2Pool[0].Card)
		}

		if cardCount >= 5 {
			//第5张牌
			for i := cardCount - 1; i >= 0; i-- {
				curCard := cardDataManager.CardData2Pool[i]

				if curCard.Value != santiaoValue && curCard.Value != POKER_A {
					//牌型牌
					cardInfo.KindCards = append(cardInfo.KindCards, curCard.Card)
					break
				}
			}
		}
	} else {
		//第4张 第 5 张牌
		for i := cardCount - 1; i >= 0; i-- {
			curCard := cardDataManager.CardData2Pool[i]

			if curCard.Value != santiaoValue {
				//牌型牌
				cardInfo.KindCards = append(cardInfo.KindCards, curCard.Card)
				if len(cardInfo.KindCards) >= 5 {
					break
				}
			}
		}
	}
	return cardInfo
}

//两对：先比大对，再比小对，都相同则比单张。 KindOfCard_TwoPair
func (this *KindOfCardFigureUpEx) IsTwoPair(cardDataManager *CardDataManager) *CardsInfo {
	cardCount := len(cardDataManager.CardDataPool)
	if cardCount < 4 {
		return nil
	}

	var card2Value []int
	for cardValue, cardCount := range cardDataManager.CardValueCount {
		if cardCount == 2 {
			card2Value = append(card2Value, int(cardValue))
		}
	}
	if len(card2Value) < 2 {
		return nil
	}
	sort.Ints(card2Value)

	cardInfo := &CardsInfo{
		Kind: KindOfCard_TwoPair,
	}

	if len(card2Value) == 3 {
		if card2Value[0] == int(POKER_A) {
			//删除中间
			card2Value = append(card2Value[:1], card2Value[2:]...)
		} else {
			//删除第一个
			card2Value = append(card2Value[:0], card2Value[1:]...)
		}
	}
	if card2Value[0] == int(POKER_A) {
		card2Value[0], card2Value[1] = card2Value[1], card2Value[0]
	}
	for j := 1; j >= 0; j-- {
		//2对
		for i := 0; i < cardCount; i++ {
			curCard := cardDataManager.CardDataPool[i]

			if curCard.Value == int32(card2Value[j]) {
				//牌型牌
				cardInfo.KindCards = append(cardInfo.KindCards, curCard.Card)
			}
		}
	}

	if cardCount >= 5 {
		//第5张牌
		if cardDataManager.CardData2Pool[0].Value == POKER_A && int32(card2Value[1]) != POKER_A {
			//牌型牌
			cardInfo.KindCards = append(cardInfo.KindCards, cardDataManager.CardData2Pool[0].Card)
		} else {
			//第5张牌
			for i := cardCount - 1; i >= 0; i-- {
				curCard := cardDataManager.CardData2Pool[i]

				if curCard.Value != int32(card2Value[0]) && curCard.Value != int32(card2Value[1]) {
					//牌型牌
					cardInfo.KindCards = append(cardInfo.KindCards, curCard.Card)
					break
				}
			}
		}
	}

	return cardInfo
}

//一对：先比对子，对子相同则依次比单张。 KindOfCard_OnePair
func (this *KindOfCardFigureUpEx) IsOnePair(cardDataManager *CardDataManager) *CardsInfo {
	cardCount := len(cardDataManager.CardDataPool)
	if cardCount < 2 {
		return nil
	}

	card1Value := INVALIDE_CARD
	for cardValue, cardCount := range cardDataManager.CardValueCount {
		if cardCount == 2 {
			card1Value = cardValue
			break
		}
	}
	if card1Value == INVALIDE_CARD {
		return nil
	}

	cardInfo := &CardsInfo{
		Kind: KindOfCard_OnePair,
	}

	//1对
	for i := 0; i < cardCount; i++ {
		curCard := cardDataManager.CardDataPool[i]

		if curCard.Value == card1Value {
			//牌型牌
			cardInfo.KindCards = append(cardInfo.KindCards, curCard.Card)
		}
	}

	//第3 ~ 5张牌
	if cardDataManager.CardData2Pool[0].Value == POKER_A && card1Value != POKER_A {
		//牌型牌
		//第 3 张
		cardInfo.KindCards = append(cardInfo.KindCards, cardDataManager.CardData2Pool[0].Card)

		//第4 ~ 5张牌
		for i := cardCount - 1; i >= 0; i-- {
			curCard := cardDataManager.CardData2Pool[i]

			if curCard.Value != card1Value && curCard.Value != POKER_A {
				//牌型牌
				cardInfo.KindCards = append(cardInfo.KindCards, curCard.Card)
				if len(cardInfo.KindCards) >= 5 {
					break
				}
			}
		}
	} else {
		//第3 ~ 5张牌
		for i := cardCount - 1; i >= 0; i-- {
			curCard := cardDataManager.CardData2Pool[i]

			if curCard.Value != card1Value {
				//牌型牌
				cardInfo.KindCards = append(cardInfo.KindCards, curCard.Card)
				if len(cardInfo.KindCards) >= 5 {
					break
				}
			}
		}
	}

	return cardInfo
}

//高牌：先比最大的牌，如相同则依次比剩余的单张 KindOfCard_HighCard
func (this *KindOfCardFigureUpEx) IsHighCard(cardDataManager *CardDataManager) *CardsInfo {
	cardCount := len(cardDataManager.CardDataPool)
	if cardCount == 0 {
		return nil
	}

	cardInfo := &CardsInfo{
		Kind: KindOfCard_HighCard,
	}

	if cardDataManager.CardData2Pool[0].Value == POKER_A {
		//牌型牌
		//第 1 张牌
		cardInfo.KindCards = append(cardInfo.KindCards, cardDataManager.CardData2Pool[0].Card)

		//第2 ~ 5张牌
		for i := cardCount - 1; i >= 0; i-- {
			curCard := cardDataManager.CardData2Pool[i]

			cardInfo.KindCards = append(cardInfo.KindCards, curCard.Card)
			if len(cardInfo.KindCards) >= 5 {
				break
			}
		}
	} else {
		//第1 ~ 5张牌
		for i := cardCount - 1; i >= 0; i-- {
			curCard := cardDataManager.CardData2Pool[i]

			cardInfo.KindCards = append(cardInfo.KindCards, curCard.Card)
			if len(cardInfo.KindCards) >= 5 {
				break
			}
		}
	}

	return cardInfo
}

func (this *KindOfCardFigureUpEx) IsTing(cards []int32, kind int32) bool {
	for i := int32(0); i < POKER_CNT; i++ {
		if hasCard(cards, i) {
			continue
		}

		var tempCard []int
		for _, v := range cards {
			tempCard = append(tempCard, int(v))
		}
		tempCard = append(tempCard, int(i))
		//按照升序排序
		sort.Ints(tempCard)

		var cardDataManager CardDataManager
		cardDataManager.Init()
		for i := 0; i < len(tempCard); i++ {
			cardDataManager.AddCard(int32(tempCard[i]))
		}
		cardDataManager.ReCal()

		ci := this.CalCardKind(&cardDataManager)
		if ci != nil && ci.Kind == kind {
			return true
		}
	}

	return false
}

func (this *KindOfCardFigureUpEx) IsTingKinds(cards []int32, kinds []int32) bool {
	for _, k := range kinds {
		if this.IsTing(cards, k) {
			return true
		}
	}

	return false
}

func (this *KindOfCardFigureUpEx) TingCount(cards, exclude []int32, kind int32) int {
	cnt := 0
	for i := int32(0); i < POKER_CNT; i++ {
		if hasCard(exclude, i) {
			continue
		}
		if hasCard(cards, i) {
			continue
		}

		var tempCard []int
		for _, v := range cards {
			tempCard = append(tempCard, int(v))
		}
		tempCard = append(tempCard, int(i))
		//按照升序排序
		sort.Ints(tempCard)

		var cardDataManager CardDataManager
		cardDataManager.Init()
		for i := 0; i < len(tempCard); i++ {
			cardDataManager.AddCard(int32(tempCard[i]))
		}
		cardDataManager.ReCal()

		ci := this.CalCardKind(&cardDataManager)
		if ci != nil && ci.Kind == kind {
			cnt++
		}
	}

	return cnt
}

func (this *KindOfCardFigureUpEx) TingKindsCount(cards, exclude, kinds []int32) int {
	cnt := 0
	for _, k := range kinds {
		cnt += this.TingCount(cards, exclude, k)
	}
	return cnt
}

func hasCard(cards []int32, c int32) bool {
	for _, _c := range cards {
		if _c == c {
			return true
		}
	}

	return false
}

func IsOverPair(ci *CardsInfo, handCard []int32, commonCard []int32) bool {
	if ci == nil || ci.Kind != KindOfCard_OnePair {
		return false
	}
	if len(handCard) != 2 {
		return false
	}

	if handCard[0]%PER_CARD_COLOR_MAX != handCard[1]%PER_CARD_COLOR_MAX {
		return false
	}

	max := CardValueMap[handCard[0]%PER_CARD_COLOR_MAX]
	for _, c := range commonCard {
		v := CardValueMap[c%PER_CARD_COLOR_MAX]
		if v > max {
			return false
		}
	}

	return true
}

func IsUnderPair(ci *CardsInfo, handCard []int32, commonCard []int32) bool {
	if ci == nil || ci.Kind != KindOfCard_OnePair {
		return false
	}
	if len(handCard) != 2 {
		return false
	}

	if handCard[0]%PER_CARD_COLOR_MAX != handCard[1]%PER_CARD_COLOR_MAX {
		return false
	}

	pv := CardValueMap[handCard[0]%PER_CARD_COLOR_MAX]
	for _, c := range commonCard {
		v := CardValueMap[c%PER_CARD_COLOR_MAX]
		if v < pv {
			return false
		}
	}

	return true
}

func IsTopPair(ci *CardsInfo, handCard []int32, commonCard []int32) bool {
	if ci == nil || ci.Kind != KindOfCard_OnePair {
		return false
	}
	if len(handCard) != 2 {
		return false
	}

	pv := CardValueMap[ci.KindCards[0]%PER_CARD_COLOR_MAX]
	max := 0
	for _, c := range commonCard {
		v := CardValueMap[c%PER_CARD_COLOR_MAX]
		if v > max {
			max = v
		}
	}

	if pv != max {
		return false
	}

	for _, c := range handCard {
		v := CardValueMap[c%PER_CARD_COLOR_MAX]
		if v == max {
			return true
		}
	}

	return false
}

func IsMiddlePair(ci *CardsInfo, handCard []int32, commonCard []int32) bool {
	if ci == nil || ci.Kind != KindOfCard_OnePair {
		return false
	}
	if len(handCard) != 2 {
		return false
	}

	pv := CardValueMap[ci.KindCards[0]%PER_CARD_COLOR_MAX]
	max := 0
	min := 13
	for _, c := range commonCard {
		v := CardValueMap[c%PER_CARD_COLOR_MAX]
		if v > max {
			max = v
		}
		if v < min {
			min = v
		}
	}

	if pv <= min || pv >= max {
		return false
	}

	for _, c := range handCard {
		v := CardValueMap[c%PER_CARD_COLOR_MAX]
		if v == pv {
			return true
		}
	}

	return false
}

func IsButtomPair(ci *CardsInfo, handCard []int32, commonCard []int32) bool {
	if ci == nil || ci.Kind != KindOfCard_OnePair {
		return false
	}
	if len(handCard) != 2 {
		return false
	}

	pv := CardValueMap[ci.KindCards[0]%PER_CARD_COLOR_MAX]
	min := 13
	for _, c := range commonCard {
		v := CardValueMap[c%PER_CARD_COLOR_MAX]
		if v < min {
			min = v
		}
	}

	if pv != min {
		return false
	}

	for _, c := range handCard {
		v := CardValueMap[c%PER_CARD_COLOR_MAX]
		if v == min {
			return true
		}
	}

	return false
}
func KindOfCardIsBetter(handCard []int32, commonCard []int32) bool {
	ok, _, _ := KindOfCardIsBetterEx(handCard, commonCard)
	return ok
}

//牌型是否较上个阶段有所提升
func KindOfCardIsBetterEx(handCard []int32, commonCard []int32) (bool, *CardsInfo, *CardsInfo) {
	var pre, cur *CardsInfo
	switch len(commonCard) {
	case 3:
		pre = KindOfCardFigureUpExSington.FigureUpByCard(commonCard)
		cur = KindOfCardFigureUpExSington.FigureUpByCard(append(handCard, commonCard...))
	case 4:
		pre = KindOfCardFigureUpExSington.FigureUpByCard(commonCard)
		cur = KindOfCardFigureUpExSington.FigureUpByCard(append(handCard, commonCard...))
	case 5:
		pre = KindOfCardFigureUpExSington.FigureUpByCard(commonCard)
		cur = KindOfCardFigureUpExSington.FigureUpByCard(append(handCard, commonCard...))
	}
	if pre != nil && cur != nil {
		if cur.Kind > pre.Kind {
			return true, pre, cur
		}
	}
	return false, pre, cur
}
