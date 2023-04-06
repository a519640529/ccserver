package dezhoupoker

import (
	"fmt"
	"github.com/idealeak/algorithm"
	"math/rand"
)

var cardKindValue = make(map[int32]int32)

/*
散牌分值=散牌里单牌分值最大的分数
一对子分值=对子分值
二对分值=对子1分值+对子2分值+单张分值
三条分值=三张分值+两个单张分值
顺子分值=对应顺子分值
同花分值=对应同花分值
葫芦分值=（三张分值+对子分值）*2
四条分值=四张分值+单张分值
同花顺分值=同花顺分值
皇家同花顺=皇家同花顺分值
*/
func CalCardsKindScore(cardsInfo *CardsInfo) int32 {
	switch cardsInfo.Kind {
	case KindOfCard_RoyalFlush:
		return CalScoreByKindAndCard(KindOfCard_RoyalFlush, cardsInfo.KindCards[4])
	case KindOfCard_StraightFlush:
		return CalScoreByKindAndCard(KindOfCard_StraightFlush, cardsInfo.KindCards[4])
	case KindOfCard_FourKind:
		return CalScoreByKindAndCard(KindOfCard_FourKind, cardsInfo.KindCards[0]) + CalScoreByKindAndCard(KindOfCard_HighCard, cardsInfo.KindCards[4])
	case KindOfCard_Fullhouse:
		return (CalScoreByKindAndCard(KindOfCard_ThreeKind, cardsInfo.KindCards[0]) + CalScoreByKindAndCard(KindOfCard_TwoPair, cardsInfo.KindCards[3])) * 2
	case KindOfCard_Flush:
		return CalScoreByKindAndCard(KindOfCard_Flush, cardsInfo.KindCards[0])
	case KindOfCard_Straight:
		return CalScoreByKindAndCard(KindOfCard_Straight, cardsInfo.KindCards[4])
	case KindOfCard_ThreeKind:
		return CalScoreByKindAndCard(KindOfCard_ThreeKind, cardsInfo.KindCards[0]) + CalScoreByKindAndCard(KindOfCard_HighCard, cardsInfo.KindCards[3]) + CalScoreByKindAndCard(KindOfCard_HighCard, cardsInfo.KindCards[4])
	case KindOfCard_TwoPair:
		return CalScoreByKindAndCard(KindOfCard_OnePair, cardsInfo.KindCards[0]) + CalScoreByKindAndCard(KindOfCard_OnePair, cardsInfo.KindCards[2]) + CalScoreByKindAndCard(KindOfCard_HighCard, cardsInfo.KindCards[4])
	case KindOfCard_OnePair:
		return CalScoreByKindAndCard(KindOfCard_OnePair, cardsInfo.KindCards[0])
	case KindOfCard_HighCard:
		return CalScoreByKindAndCard(KindOfCard_HighCard, cardsInfo.KindCards[0])
	default:
	}
	return 0
}

func MakeCardKindKey(kind int32, card int32) int32 {
	return kind*10000 + card
}

func SetCardKindValue(key, value int32) {
	cardKindValue[key] = value
}

func CalScoreByKindAndCard(kind int32, card int32) int32 {
	if v, exist := cardKindValue[MakeCardKindKey(kind, card)]; exist {
		return v
	}
	return 0
}

type PlayerCard struct {
	HandCard           []int32
	UserData           interface{}
	WinningProbability int32
	AllCardKind        [KindOfCard_Max]int32
	WinCardKind        [KindOfCard_Max]int32
	winTimes           int32
	CI                 *CardsInfo
}

type GameCtx struct {
	PlayerCards   []*PlayerCard
	CommonCard    []int32
	RestCard      []int32
	CommonCardCnt int
	Possibilities int
}

func (gctx *GameCtx) GetMaxKindOfCard() int32 {
	maxKind := KindOfCard_HighCard
	for _, pc := range gctx.PlayerCards {
		if pc != nil && pc.CI != nil && pc.CI.Kind > maxKind {
			maxKind = pc.CI.Kind
		}
	}
	return maxKind
}

func (gctx *GameCtx) GetMaxCardInfo() *CardsInfo {
	var max *CardsInfo
	for _, pc := range gctx.PlayerCards {
		if pc != nil && pc.CI != nil {
			if max == nil {
				max = pc.CI
			} else if pc.CI.Value > max.Value {
				max = pc.CI
			}
		}
	}
	return max
}

func CalWinningProbability(ctx *GameCtx) {
	if ctx == nil {
		return
	}

	possibilities := uint64(0)
	max := int64(0)
	maxIdx := make([]int, 0, len(ctx.PlayerCards))
	oldCC := len(ctx.CommonCard)
	m := 5 - oldCC
	if m == 0 { //手牌和公牌全部确定
		for k := 0; k < len(ctx.PlayerCards); k++ {
			ctx.PlayerCards[k].CI = KindOfCardFigureUpSington.figureUp(ctx.PlayerCards[k].HandCard, ctx.CommonCard)
			ctx.PlayerCards[k].CI.CalValue()
			if ctx.PlayerCards[k].CI.Value > max {
				max = ctx.PlayerCards[k].CI.Value
				maxIdx = maxIdx[0:0]
				maxIdx = append(maxIdx, k)
			} else if ctx.PlayerCards[k].CI.Value == max {
				maxIdx = append(maxIdx, k)
			}
			ctx.PlayerCards[k].AllCardKind[ctx.PlayerCards[k].CI.Kind]++
		}
		if len(maxIdx) != 0 {
			for _, idx := range maxIdx {
				ctx.PlayerCards[idx].winTimes++
				ctx.PlayerCards[idx].WinCardKind[ctx.PlayerCards[idx].CI.Kind]++
			}
		}

		possibilities = uint64(len(maxIdx))
	} else { //计算各种组合的可能
		//先填充够5张公牌的位置
		for i := oldCC; i < 5; i++ {
			ctx.CommonCard = append(ctx.CommonCard, -1)
		}

		n := len(ctx.RestCard)
		atable := make([]int32, n)
		for i := 0; i < n; i++ {
			atable[i] = int32(i)
		}

		possibilities = algorithm.CombNumber(uint64(n), uint64(m))
		if possibilities > 10000 {
			possibilities = 10000
		}

		c := algorithm.CombinerSelectUseRecursion(atable, m)
		if len(c) > int(possibilities) {
			for i := 0; i < int(possibilities); i++ {
				r := rand.Intn(len(c))
				//设置公牌
				for j := 0; j < len(c[r]); j++ {
					ctx.CommonCard[oldCC+j] = ctx.RestCard[c[r][j]]
				}

				max = 0
				maxIdx = maxIdx[0:0]
				for k := 0; k < len(ctx.PlayerCards); k++ {
					ctx.PlayerCards[k].CI = KindOfCardFigureUpSington.figureUp(ctx.PlayerCards[k].HandCard, ctx.CommonCard)
					ctx.PlayerCards[k].CI.CalValue()
					if ctx.PlayerCards[k].CI.Value > max {
						max = ctx.PlayerCards[k].CI.Value
						maxIdx = maxIdx[0:0]
						maxIdx = append(maxIdx, k)
					} else if ctx.PlayerCards[k].CI.Value == max {
						maxIdx = append(maxIdx, k)
					}
					ctx.PlayerCards[k].AllCardKind[ctx.PlayerCards[k].CI.Kind]++
				}
				if len(maxIdx) != 0 {
					for _, idx := range maxIdx {
						ctx.PlayerCards[idx].winTimes++
						ctx.PlayerCards[idx].WinCardKind[ctx.PlayerCards[idx].CI.Kind]++
					}
				}
			}
		} else {
			for i := 0; i < len(c); i++ {
				//设置公牌
				for j := 0; j < len(c[i]); j++ {
					ctx.CommonCard[oldCC+j] = ctx.RestCard[c[i][j]]
				}

				max = 0
				maxIdx = maxIdx[0:0]
				for k := 0; k < len(ctx.PlayerCards); k++ {
					ctx.PlayerCards[k].CI = KindOfCardFigureUpSington.figureUp(ctx.PlayerCards[k].HandCard, ctx.CommonCard)
					ctx.PlayerCards[k].CI.CalValue()
					if ctx.PlayerCards[k].CI.Value > max {
						max = ctx.PlayerCards[k].CI.Value
						maxIdx = maxIdx[0:0]
						maxIdx = append(maxIdx, k)
					} else if ctx.PlayerCards[k].CI.Value == max {
						maxIdx = append(maxIdx, k)
					}
					ctx.PlayerCards[k].AllCardKind[ctx.PlayerCards[k].CI.Kind]++
				}
				if len(maxIdx) != 0 {
					for _, idx := range maxIdx {
						ctx.PlayerCards[idx].winTimes++
						ctx.PlayerCards[idx].WinCardKind[ctx.PlayerCards[idx].CI.Kind]++
					}
				}
			}
		}
	}

	fmt.Printf("Common Card:")
	if oldCC > 0 {
		for i := 0; i < oldCC; i++ {
			fmt.Printf("%s", Card(ctx.CommonCard[i]))
		}
		ci := KindOfCardFigureUpExSington.FigureUpByCard(ctx.CommonCard[:oldCC])
		if ci != nil {
			fmt.Printf(":%s", ci.KindStr())
		}
	}
	fmt.Println()

	ctx.Possibilities = int(possibilities)
	for i := 0; i < len(ctx.PlayerCards); i++ {
		ctx.PlayerCards[i].WinningProbability = int32(int64(ctx.PlayerCards[i].winTimes) * 10000 / int64(possibilities))
		fmt.Printf("UserData<%#v> %s%s WinningProbability=%.2f%% \n", ctx.PlayerCards[i].UserData, Card(ctx.PlayerCards[i].HandCard[0]), Card(ctx.PlayerCards[i].HandCard[1]), float32(ctx.PlayerCards[i].WinningProbability)/100)
	}
}

func PossibleKindOfCards(ctx *GameCtx) []int32 {
	kinds := make([]int32, KindOfCard_Max)
	if ctx == nil {
		return kinds
	}

	if ctx.CommonCardCnt < 5 {
		return kinds
	}

	n := len(ctx.RestCard)
	atable := make([]int32, n)
	for i := 0; i < n; i++ {
		atable[i] = int32(i)
	}

	var handCard [2]int32
	c := algorithm.CombinerSelectUseRecursion(atable, 2)
	for i := 0; i < len(c); i++ {
		//设置公牌
		for j := 0; j < len(c[i]); j++ {
			handCard[j] = ctx.RestCard[c[i][j]]
		}

		ci := KindOfCardFigureUpSington.figureUp(handCard[:], ctx.CommonCard)
		if ci != nil {
			kinds[ci.Kind]++
		}
	}
	return kinds
}

func CaculKindOfCardCount(kinds []int32, expectKind []int32) (int32, int32) {
	var cnt int32
	for _, k := range expectKind {
		cnt += kinds[k]
	}
	var total int32
	for _, v := range kinds {
		total += v
	}
	return total, cnt
}

func CaculGreaterKindOfCardCount(kinds []int32, cmpKind int32) (int32, int32) {
	var cnt int32
	var total int32
	for k, v := range kinds {
		total += v
		if k >= int(cmpKind) {
			cnt++
		}
	}
	return total, cnt
}

func PossibleGreaterKindOfCards(ctx *GameCtx, cmpCI *CardsInfo) (int, int) {
	if ctx == nil {
		return 0, 1
	}

	if ctx.CommonCardCnt < 5 {
		return 0, 1
	}

	if cmpCI == nil {
		return 0, 1
	}

	n := len(ctx.RestCard)
	atable := make([]int32, n)
	for i := 0; i < n; i++ {
		atable[i] = int32(i)
	}

	var cnt int
	var handCard [2]int32
	c := algorithm.CombinerSelectUseRecursion(atable, 2)
	for i := 0; i < len(c); i++ {
		//设置公牌
		for j := 0; j < len(c[i]); j++ {
			handCard[j] = ctx.RestCard[c[i][j]]
		}

		ci := KindOfCardFigureUpSington.figureUp(handCard[:], ctx.CommonCard)
		if ci != nil {
			ci.CalValue()
		}

		if ci.Value > cmpCI.Value {
			cnt++
		}
	}
	return cnt, len(c)
}
