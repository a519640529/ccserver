package dezhoupoker

import (
	"fmt"
	"sort"
)

const (
	KindOfCard_HighCard      int32 = iota //0 高牌：先比最大的牌，如相同则依次比剩余的单张
	KindOfCard_OnePair                    //1 一对：先比对子，对子相同则依次比单张。
	KindOfCard_TwoPair                    //2 两对：先比大对，再比小对，都相同则比单张。
	KindOfCard_ThreeKind                  //3 三条：先比三条，三条相同比单张
	KindOfCard_Straight                   //4 顺子：比顺子的大小。A2345是最小的顺子。
	KindOfCard_Flush                      //5 同花：比最大的单张，如相同则依次比剩余的单张。
	KindOfCard_Fullhouse                  //6 葫芦：先比三条，三条相同比对子
	KindOfCard_FourKind                   //7 四条：先比四条，四条相同比单张。
	KindOfCard_StraightFlush              //8 同花顺
	KindOfCard_RoyalFlush                 //9 皇家同花顺

	KindOfCard_Invalide //无效边池对应的牌型
	KindOfCard_Max
)

//对子延伸牌型
const (
	KindOfCardEx_OverPair   int32 = iota //高对（overpair）是一个玩家自己手上的对子，它比公共牌上任何一张牌可能组成的对子更大
	KindOfCardEx_TopPair                 //由玩家手里的一张牌和最大的一张公共牌组成的对子叫做顶对（top pair）
	KindOfCardEx_MiddlePair              //玩家手上的一张牌和牌面上的一张中等牌组成的对子
	KindOfCardEx_UnderPair               //比所有公共牌数字都小的对子。因此，任何与牌面组成的对子将打败低对
)

var KindOfCardStr = []string{
	"高牌",
	"一对",
	"两对",
	"三条",
	"顺子",
	"同花",
	"葫芦",
	"四条",
	"同花顺",
	"皇家同花顺",
	"无效",
}

type CardData struct {
	Color int32
	Value int32
	Card  int32
}

func (this *CardData) Init(card int32) {
	this.Card = card
	this.Color = this.Card / PER_CARD_COLOR_MAX
	this.Value = this.Card % PER_CARD_COLOR_MAX
}

//----------------------------------------------------------------------------------------------------------------------------------------------------------------
type CardDataManager struct {
	CardDataPool  []CardData //带花色排序
	CardData2Pool []CardData //不带花色排序

	CardValueCount map[int32]int32
	CardColorCount map[int32]int32
}

func (this *CardDataManager) Init() {
	this.CardValueCount = make(map[int32]int32)
	this.CardColorCount = make(map[int32]int32)
}
func (this *CardDataManager) AddCard(card int32) {
	var cardData CardData
	cardData.Init(card)

	this.CardDataPool = append(this.CardDataPool, cardData)

	if _, ok := this.CardValueCount[cardData.Value]; ok {
		this.CardValueCount[cardData.Value]++
	} else {
		this.CardValueCount[cardData.Value] = 1
	}

	if _, ok := this.CardColorCount[cardData.Color]; ok {
		this.CardColorCount[cardData.Color]++
	} else {
		this.CardColorCount[cardData.Color] = 1
	}
}

func (this *CardDataManager) ReCal() {
	this.CardData2Pool = append(this.CardData2Pool, this.CardDataPool...)

	//升序
	len := len(this.CardData2Pool)
	for i := 0; i < len; i++ {
		for j := i + 1; j < len; j++ {
			if this.CardData2Pool[i].Value > this.CardData2Pool[j].Value {
				this.CardData2Pool[i], this.CardData2Pool[j] = this.CardData2Pool[j], this.CardData2Pool[i]
			}
		}
	}
}

//----------------------------------------------------------------------------------------------------------------------------------------------------------------
type CardsInfo struct {
	Kind       int32 //牌型
	KindCards  []int32
	Value      int64 //牌力大小。KK_KK_KK_VV_VV_VV_VV_VV
	ValueScore int32 //计算得分,在数值控制里面用
}

func (this *CardsInfo) KindStr() string {
	return KindOfCardStr[this.Kind]
}

func (this *CardsInfo) MakeValue(kindValue, kind1Value, kind2Value, poker1Value, poker2Value, poker3Value, poker4Value, poker5Value int32) int64 {
	kind_value := kindValue*10000 + kind1Value*100 + kind2Value
	poker_value := poker1Value*100000000 + poker2Value*1000000 + poker3Value*10000 + poker4Value*100 + poker5Value
	result_Value := int64(kind_value)*10000000000 + int64(poker_value)
	return result_Value
}

func (this *CardsInfo) CalValue() {
	switch this.Kind {
	case KindOfCard_RoyalFlush:
		this.Value = this.MakeValue(this.Kind, 0, 0, 0, 0, 0, 0, 0)
	case KindOfCard_StraightFlush:
		this.Value = this.MakeValue(this.Kind, 0, 0, this.ValueToWeight(this.KindCards[0]), 0, 0, 0, 0)
	case KindOfCard_FourKind:
		this.Value = this.MakeValue(this.Kind, this.ValueToWeight(this.KindCards[0]), 0, this.ValueToWeight(this.KindCards[4]), 0, 0, 0, 0)
	case KindOfCard_Fullhouse:
		this.Value = this.MakeValue(this.Kind, this.ValueToWeight(this.KindCards[0]), this.ValueToWeight(this.KindCards[3]), 0, 0, 0, 0, 0)
	case KindOfCard_Flush:
		this.Value = this.MakeValue(this.Kind, 0, 0, this.ValueToWeight(this.KindCards[0]), this.ValueToWeight(this.KindCards[1]), this.ValueToWeight(this.KindCards[2]), this.ValueToWeight(this.KindCards[3]), this.ValueToWeight(this.KindCards[4]))
	case KindOfCard_Straight:
		this.Value = this.MakeValue(this.Kind, 0, 0, this.ValueToWeight(this.KindCards[0]), 0, 0, 0, 0)
	case KindOfCard_ThreeKind:
		this.Value = this.MakeValue(this.Kind, this.ValueToWeight(this.KindCards[0]), 0, this.ValueToWeight(this.KindCards[3]), this.ValueToWeight(this.KindCards[4]), 0, 0, 0)
	case KindOfCard_TwoPair:
		this.Value = this.MakeValue(this.Kind, this.ValueToWeight(this.KindCards[0]), this.ValueToWeight(this.KindCards[2]), this.ValueToWeight(this.KindCards[4]), 0, 0, 0, 0)
	case KindOfCard_OnePair:
		this.Value = this.MakeValue(this.Kind, this.ValueToWeight(this.KindCards[0]), 0, this.ValueToWeight(this.KindCards[2]), this.ValueToWeight(this.KindCards[3]), this.ValueToWeight(this.KindCards[4]), 0, 0)
	case KindOfCard_HighCard:
		this.Value = this.MakeValue(this.Kind, 0, 0, this.ValueToWeight(this.KindCards[0]), this.ValueToWeight(this.KindCards[1]), this.ValueToWeight(this.KindCards[2]), this.ValueToWeight(this.KindCards[3]), this.ValueToWeight(this.KindCards[4]))
	default:
		this.Value = this.MakeValue(this.Kind, 0, 0, 0, 0, 0, 0, 0)
	}

}

func (this *CardsInfo) ValueToWeight(pokerCard int32) int32 {
	cardValue := pokerCard % PER_CARD_COLOR_MAX
	if cardValue == POKER_A {
		return POKER_A_Weight
	} else {
		return cardValue
	}
}

//----------------------------------------------------------------------------------------------------------------------------------------------------------------
type KindOfCardFigureUp struct {
}

var KindOfCardFigureUpSington = &KindOfCardFigureUp{}

func (this *KindOfCardFigureUp) FigureUpByCard(handcards [HandCardNum]int32, communityCards [CommunityCardNum]int32) *CardsInfo {
	return this.figureUp(handcards[:], communityCards[:])
}

func (this *KindOfCardFigureUp) figureUp(handcards []int32, communityCards []int32) *CardsInfo {
	var tempCard []int
	for i := int32(0); i < HandCardNum; i++ {
		tempCard = append(tempCard, int(handcards[i]))
	}
	for i := int32(0); i < CommunityCardNum; i++ {
		tempCard = append(tempCard, int(communityCards[i]))
	}
	//按照升序排序
	sort.Ints(tempCard)

	var cardDataManager CardDataManager
	cardDataManager.Init()
	for i := int32(0); i < TotalCardNum; i++ {
		cardDataManager.AddCard(int32(tempCard[i]))
	}
	cardDataManager.ReCal()

	//根据value 排序
	cardsInfo := this.CalCardKind(&cardDataManager)
	if cardsInfo != nil {
		cardsInfo.CalValue()
	}
	return cardsInfo
}

func (this *KindOfCardFigureUp) CalCardKind(cardDataManager *CardDataManager) *CardsInfo {

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

	fmt.Println("出错了, 永远不应该走到这里 : ", cardDataManager.CardDataPool)
	return nil
}

//皇家同花顺  KindOfCard_RoyalFlush
func (this *KindOfCardFigureUp) IsRoyalFlush(cardDataManager *CardDataManager) *CardsInfo {

	var cur_poker CardData
	cur_poker.Init(cardDataManager.CardDataPool[TotalCardNum-1].Card)
	var rst_kind_card []int32
	rst_kind_card = append(rst_kind_card, cur_poker.Card)

	for i := TotalCardNum - 2; i >= 0; i-- {
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
	for i := int32(0); i < TotalCardNum; i++ {
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
func (this *KindOfCardFigureUp) IsStraightFlush(cardDataManager *CardDataManager) *CardsInfo {

	var cur_poker CardData
	cur_poker.Init(cardDataManager.CardDataPool[TotalCardNum-1].Card)
	var rst_kind_card []int32
	rst_kind_card = append(rst_kind_card, cur_poker.Card)

	for i := TotalCardNum - 2; i >= 0; i-- {
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
func (this *KindOfCardFigureUp) IsFourKind(cardDataManager *CardDataManager) *CardsInfo {

	for cardValue, cardCount := range cardDataManager.CardValueCount {
		if cardCount == 4 {
			cardInfo := &CardsInfo{
				Kind: KindOfCard_FourKind,
			}

			//4张 牌型牌
			for i := TotalCardNum - 1; i >= 0; i-- {
				curCard := cardDataManager.CardData2Pool[i]

				if curCard.Value == cardValue {
					//牌型牌
					cardInfo.KindCards = append(cardInfo.KindCards, curCard.Card)
				}
			}

			//第5张牌型牌
			fithCard := int32(0)
			for i := TotalCardNum - 1; i >= 0; i-- {
				curCard := cardDataManager.CardData2Pool[i]

				if curCard.Card != cardInfo.KindCards[0] {
					if curCard.Value == POKER_A {
						fithCard = curCard.Card
						break
					} else {
						if cardValue != curCard.Value && fithCard < curCard.Value {
							fithCard = curCard.Card
						}
					}
				}
			}
			cardInfo.KindCards = append(cardInfo.KindCards, fithCard)
			return cardInfo
		}
	}
	return nil
}

//葫芦：先比三条，三条相同比对子 KindOfCard_Fullhouse
func (this *KindOfCardFigureUp) IsFullhouse(cardDataManager *CardDataManager) *CardsInfo {

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
		for i := int32(0); i < TotalCardNum; i++ {
			curCard := cardDataManager.CardDataPool[i]

			if curCard.Value == card3Value {
				//牌型牌
				cardInfo.KindCards = append(cardInfo.KindCards, curCard.Card)
			}
		}

		//2对
		for i := int32(0); i < TotalCardNum; i++ {
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
func (this *KindOfCardFigureUp) IsFlush(cardDataManager *CardDataManager) *CardsInfo {
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
	for i := TotalCardNum - 1; i >= 0; i-- {
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
func (this *KindOfCardFigureUp) IsStraight(cardDataManager *CardDataManager) *CardsInfo {
	//需要先判断是最大的顺子，才能再判定普通顺子
	card_info := this.IsStraightMax(cardDataManager)
	if card_info == nil {
		card_info = this.IsStraightNormal(cardDataManager)
	}
	return card_info
}
func (this *KindOfCardFigureUp) IsStraightNormal(cardDataManager *CardDataManager) *CardsInfo {

	var cur_poker CardData
	cur_poker.Init(cardDataManager.CardData2Pool[TotalCardNum-1].Card)
	var rst_kind_card []int32
	rst_kind_card = append(rst_kind_card, cur_poker.Card)

	for i := TotalCardNum - 2; i >= 0; i-- {
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
func (this *KindOfCardFigureUp) IsStraightMax(cardDataManager *CardDataManager) *CardsInfo {
	var cur_poker CardData
	cur_poker.Init(cardDataManager.CardData2Pool[TotalCardNum-1].Card)
	var rst_kind_card []int32
	rst_kind_card = append(rst_kind_card, cur_poker.Card)

	for i := TotalCardNum - 2; i >= 0; i-- {
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
func (this *KindOfCardFigureUp) IsThreeKind(cardDataManager *CardDataManager) *CardsInfo {
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
	for i := TotalCardNum - 1; i >= 0; i-- {
		curCard := cardDataManager.CardDataPool[i]

		if curCard.Value == santiaoValue {
			//牌型牌
			cardInfo.KindCards = append(cardInfo.KindCards, curCard.Card)
		}
	}

	if cardDataManager.CardData2Pool[0].Value == POKER_A && santiaoValue != POKER_A {
		//第4张牌
		cardInfo.KindCards = append(cardInfo.KindCards, cardDataManager.CardData2Pool[0].Card)

		//第5张牌
		for i := TotalCardNum - 1; i >= 0; i-- {
			curCard := cardDataManager.CardData2Pool[i]

			if curCard.Value != santiaoValue && curCard.Value != POKER_A {
				//牌型牌
				cardInfo.KindCards = append(cardInfo.KindCards, curCard.Card)
				break
			}
		}
	} else {
		//第4张 第 5 张牌
		for i := TotalCardNum - 1; i >= 0; i-- {
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
func (this *KindOfCardFigureUp) IsTwoPair(cardDataManager *CardDataManager) *CardsInfo {

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
		for i := int32(0); i < TotalCardNum; i++ {
			curCard := cardDataManager.CardDataPool[i]

			if curCard.Value == int32(card2Value[j]) {
				//牌型牌
				cardInfo.KindCards = append(cardInfo.KindCards, curCard.Card)
			}
		}
	}

	//第5张牌
	if cardDataManager.CardData2Pool[0].Value == POKER_A && int32(card2Value[1]) != POKER_A {
		//牌型牌
		cardInfo.KindCards = append(cardInfo.KindCards, cardDataManager.CardData2Pool[0].Card)
	} else {
		//第5张牌
		for i := TotalCardNum - 1; i >= 0; i-- {
			curCard := cardDataManager.CardData2Pool[i]

			if curCard.Value != int32(card2Value[0]) && curCard.Value != int32(card2Value[1]) {
				//牌型牌
				cardInfo.KindCards = append(cardInfo.KindCards, curCard.Card)
				break
			}
		}
	}

	return cardInfo
}

//一对：先比对子，对子相同则依次比单张。 KindOfCard_OnePair
func (this *KindOfCardFigureUp) IsOnePair(cardDataManager *CardDataManager) *CardsInfo {

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
	for i := int32(0); i < TotalCardNum; i++ {
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
		for i := TotalCardNum - 1; i >= 0; i-- {
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
		for i := TotalCardNum - 1; i >= 0; i-- {
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
func (this *KindOfCardFigureUp) IsHighCard(cardDataManager *CardDataManager) *CardsInfo {
	cardInfo := &CardsInfo{
		Kind: KindOfCard_HighCard,
	}

	if cardDataManager.CardData2Pool[0].Value == POKER_A {
		//牌型牌
		//第 1 张牌
		cardInfo.KindCards = append(cardInfo.KindCards, cardDataManager.CardData2Pool[0].Card)

		//第2 ~ 5张牌
		for i := TotalCardNum - 1; i >= 0; i-- {
			curCard := cardDataManager.CardData2Pool[i]

			cardInfo.KindCards = append(cardInfo.KindCards, curCard.Card)
			if len(cardInfo.KindCards) >= 5 {
				break
			}
		}
	} else {
		//第1 ~ 5张牌
		for i := TotalCardNum - 1; i >= 0; i-- {
			curCard := cardDataManager.CardData2Pool[i]

			cardInfo.KindCards = append(cardInfo.KindCards, curCard.Card)
			if len(cardInfo.KindCards) >= 5 {
				break
			}
		}
	}

	return cardInfo
}
