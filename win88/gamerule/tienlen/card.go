package tienlen

import (
	"fmt"
	"sort"
)

const (
	Tienlen_Pass    int = iota //nil
	Single                     //1 单张
	Twin                       //2 对子
	Straight                   //3 顺子
	ColorStraight              //4 同花顺
	Straight_Twin              //5 连对（双顺
	Triple                     //6 三张
	Four_Bomb                  //7 炸弹
	Straight_Triple            //8三顺
	Plane_Single               //9三带单
	Plane_Twin                 //10三带双
)

// 不同牌分值
const (
	Score1  int32 = 1
	Score5  int32 = 5
	Score10 int32 = 10
	Score15 int32 = 15
	Score20 int32 = 20
	Score25 int32 = 25
)

// 打到底不同牌分值百分比
const (
	Score2End5  int32 = 50
	Score2End10 int32 = 100
	Score2End15 int32 = 150
	Score2End20 int32 = 200
	Score2End25 int32 = 250
)

// Color 黑桃0 梅花1 方片2 紅桃3
func Color(c int32) int {
	return int(c) / PER_CARD_COLOR_MAX
}

func Value(c int32) int {
	return int(c) % PER_CARD_COLOR_MAX
}

func ValueStr(c int32) int {
	if int(c+3)%PER_CARD_COLOR_MAX == 0 {
		return PER_CARD_COLOR_MAX
	}
	return int(c+3) % PER_CARD_COLOR_MAX
}

func IsSingle(cards []int32) bool { //单张
	if len(cards) == 1 {
		return true
	}
	return false
}

func IsTwin(cards []int32) bool { //对子
	if len(cards) == 2 {
		if Value(cards[0]) == Value(cards[1]) {
			return true
		}
	}
	return false
}

func IsStraight(cards []int32) bool { //顺子
	if len(cards) < 3 || len(cards) > 12 {
		return false
	}
	tmpCards := []int{}
	for _, card := range cards {
		if Value(card) == Card_Value_2 {
			return false
		}
		tmpCards = append(tmpCards, Value(card))
	}
	sort.Ints(tmpCards)
	for i := 0; i < len(tmpCards)-1; i++ {
		card := tmpCards[i]
		nextCard := tmpCards[i+1]
		if nextCard-card != 1 { //不相邻
			return false
		}
	}
	//fmt.Println("顺子排序：", tmpCards)
	return true
}

func IsColorStraight(cards []int32) bool { //同花顺子
	// 先是顺子
	if !IsStraight(cards) {
		return false
	}
	c := Color(cards[0])
	for _, card := range cards {
		if c != Color(card) {
			return false
		}
	}
	return true
}

func IsStraightTwin(cards []int32) bool { //连对(双顺)
	if len(cards) < 4 || len(cards)%2 != 0 {
		return false
	}
	tmpCards := []int{}
	for _, card := range cards {
		if Value(card) == Card_Value_2 {
			return false
		}
		tmpCards = append(tmpCards, Value(card))
	}
	sort.Ints(tmpCards)
	for i := 0; i < len(tmpCards); i += 2 {
		card := tmpCards[i]
		nextCard := tmpCards[i+1]
		if card != nextCard {
			return false
		}
		if i < len(cards)-2 {
			if tmpCards[i+2]-card != 1 {
				return false
			}
			if card == Card_Value_2 || nextCard == Card_Value_2 { //2不能加入顺子中
				return false
			}
		}
	}
	return true
}

func IsTriple(cards []int32) bool { //三张
	if len(cards)%3 != 0 {
		return false
	}
	if Value(cards[0]) != Value(cards[1]) {
		return false
	}
	if Value(cards[1]) != Value(cards[2]) {
		return false
	}
	return true
}

func IsFourBomb(cards []int32) bool { //炸弹
	if len(cards) != 4 {
		return false
	}
	if Value(cards[0]) != Value(cards[1]) {
		return false
	}
	if Value(cards[1]) != Value(cards[2]) {
		return false
	}
	if Value(cards[2]) != Value(cards[3]) {
		return false
	}
	return true
}

func IsStraightTriple(cards []int32) bool { //三顺
	if len(cards) == 0 {
		return false
	}
	if len(cards)%3 != 0 {
		return false
	}
	mapSTriple := make(map[int32]int32)
	for _, card := range cards {
		mapSTriple[int32(Value(card))]++
	}
	if len(mapSTriple) > 0 {
		valueSTriple := []int32{}
		for card, num := range mapSTriple {
			if num == 3 {
				valueSTriple = append(valueSTriple, card)
			}
		}
		if len(valueSTriple) != len(cards)/3 {
			return false
		}
		if len(valueSTriple) > 0 {
			sort.Slice(valueSTriple, func(i, j int) bool {
				if valueSTriple[i] > valueSTriple[j] {
					return false
				}
				return true
			})
			tmpPairs := FindOneStraightWithWidth(len(valueSTriple), valueSTriple)
			if len(tmpPairs) > 0 {
				return true
			}
		}
	}

	return false
}

func IsPlaneSingle(cards []int32) bool { //飞机带单包括三带一(只能最后一手出牌)
	if len(cards) < 4 {
		return false
	}
	mapSTriple := make(map[int32]int32)
	for _, card := range cards {
		mapSTriple[int32(Value(card))]++
	}
	if len(mapSTriple) > 0 {
		valueSTriple := []int32{}
		for card, num := range mapSTriple {
			if num == 3 {
				valueSTriple = append(valueSTriple, card)
			}
		}
		if len(valueSTriple)*5 <= len(cards) {
			return false
		}
		if len(valueSTriple) > 0 {
			sort.Slice(valueSTriple, func(i, j int) bool {
				if valueSTriple[i] > valueSTriple[j] {
					return false
				}
				return true
			})
			tmpPairs := FindOneStraightWithWidth(len(valueSTriple), valueSTriple)
			if len(tmpPairs) > 0 {
				return true
			}
		}
	}
	return false
}

func IsPlaneTwin(cards []int32) bool { //飞机带双包括三带二
	if len(cards) == 0 {
		return false
	}
	if len(cards)%5 != 0 {
		return false
	}
	mapSTriple := make(map[int32]int32)
	for _, card := range cards {
		mapSTriple[int32(Value(card))]++
	}
	if len(mapSTriple) > 0 {
		valueSTriple := []int32{}
		for card, num := range mapSTriple {
			if num == 3 {
				valueSTriple = append(valueSTriple, card)
			}
		}
		if len(valueSTriple) != len(cards)/5 {
			return false
		}
		if len(valueSTriple) > 0 {
			sort.Slice(valueSTriple, func(i, j int) bool {
				if valueSTriple[i] > valueSTriple[j] {
					return false
				}
				return true
			})
			tmpPairs := FindOneStraightWithWidth(len(valueSTriple), valueSTriple)
			if len(tmpPairs) > 0 {
				return true
			}
		}
	}
	return false
}

// 返回value_card
func FindPlaneTwinHead(cards []int32) int32 { //找飞机头
	if len(cards) == 0 {
		return InvalideCard
	}
	if len(cards)%5 != 0 {
		return InvalideCard
	}
	mapSTriple := make(map[int32]int32)
	for _, card := range cards {
		mapSTriple[int32(Value(card))]++
	}
	if len(mapSTriple) > 0 {
		valueSTriple := []int32{}
		for card, num := range mapSTriple {
			if num == 3 {
				valueSTriple = append(valueSTriple, card)
			}
		}
		if len(valueSTriple) != len(cards)/5 {
			return InvalideCard
		}
		if len(valueSTriple) > 0 {
			sort.Slice(valueSTriple, func(i, j int) bool {
				if valueSTriple[i] > valueSTriple[j] {
					return false
				}
				return true
			})
			tmpPairs := FindOneStraightWithWidth(len(valueSTriple), valueSTriple)
			if len(tmpPairs) > 0 {
				return tmpPairs[len(tmpPairs)-1]
			}
		}
	}
	return InvalideCard
}

func RulePopEnable(cards []int32) (bool, int) {
	isRule := false
	ruleType := Tienlen_Pass
	switch len(cards) {
	case 1:
		isRule = true
		ruleType = Single
		break
	case 2:
		if IsTwin(cards) {
			isRule = true
			ruleType = Twin
		}
		break
	case 3:
		if IsColorStraight(cards) { //同花顺
			isRule = true
			ruleType = ColorStraight
		} else if IsStraight(cards) { //顺子
			isRule = true
			ruleType = Straight
		} else if IsTriple(cards) { //三张
			isRule = true
			ruleType = Triple
		}
		break
	case 4:
		if IsColorStraight(cards) { //同花顺
			isRule = true
			ruleType = ColorStraight
		} else if IsStraight(cards) { //顺子
			isRule = true
			ruleType = Straight
		} else if IsFourBomb(cards) { //炸弹
			isRule = true
			ruleType = Four_Bomb
		} else if IsStraightTwin(cards) { //连对（双顺）
			isRule = true
			ruleType = Straight_Twin
		}
		break
	default:
		if IsColorStraight(cards) { //同花顺
			isRule = true
			ruleType = ColorStraight
		} else if IsStraight(cards) { //顺子
			isRule = true
			ruleType = Straight
		} else if IsStraightTwin(cards) { //连对（双顺）
			isRule = true
			ruleType = Straight_Twin
		}
		break
	}
	return isRule, ruleType
}

func RulePopEnable_yl(cards []int32) (bool, int) { //娱乐场
	isRule := false
	ruleType := Tienlen_Pass
	switch len(cards) {
	case 1:
		isRule = true
		ruleType = Single
		break
	case 2:
		if IsTwin(cards) {
			isRule = true
			ruleType = Twin
		}
		break
	case 3:
		if IsColorStraight(cards) { //同花顺
			isRule = true
			ruleType = ColorStraight
		} else if IsStraight(cards) { //顺子
			isRule = true
			ruleType = Straight
		} else if IsTriple(cards) { //三张
			isRule = true
			ruleType = Triple
		}
		break
	case 4:
		if IsColorStraight(cards) { //同花顺
			isRule = true
			ruleType = ColorStraight
		} else if IsStraight(cards) { //顺子
			isRule = true
			ruleType = Straight
		} else if IsFourBomb(cards) { //炸弹
			isRule = true
			ruleType = Four_Bomb
		} else if IsStraightTwin(cards) { //连对（双顺）
			isRule = true
			ruleType = Straight_Twin
		} else if IsPlaneSingle(cards) { //飞机带单
			isRule = true
			ruleType = Plane_Single
		}
		break
	default:
		if IsColorStraight(cards) { //同花顺
			isRule = true
			ruleType = ColorStraight
		} else if IsStraight(cards) { //顺子
			isRule = true
			ruleType = Straight
		} else if IsStraightTwin(cards) { //连对（双顺）
			isRule = true
			ruleType = Straight_Twin
		} else if IsStraightTriple(cards) { //三顺
			isRule = true
			ruleType = Straight_Triple
		} else if IsPlaneTwin(cards) { //飞机带双
			isRule = true
			ruleType = Plane_Twin
		} else if IsPlaneSingle(cards) { //飞机带单
			isRule = true
			ruleType = Plane_Single
		}
		break
	}
	return isRule, ruleType
}

// return :可压住,是否炸，炸多少分
func CanDel(lastCards, cards []int32, toEnd bool) (bool, bool, int32) {
	isBomb := false
	bombScore := int32(0)
	if len(lastCards) == 0 || len(cards) == 0 {
		return false, isBomb, bombScore
	}
	for _, card := range lastCards {
		if card == InvalideCard {
			return false, isBomb, bombScore
		}
	}
	for _, card := range cards {
		if card == InvalideCard {
			return false, isBomb, bombScore
		}
	}
	sort.Slice(lastCards, func(i, j int) bool {
		if Value(lastCards[i]) > Value(lastCards[j]) {
			return false
		} else if Value(lastCards[i]) == Value(lastCards[j]) {
			return Color(lastCards[i]) < Color(lastCards[j])
		}
		return true
	})
	sort.Slice(cards, func(i, j int) bool {
		if Value(cards[i]) > Value(cards[j]) {
			return false
		} else if Value(cards[i]) == Value(cards[j]) {
			return Color(cards[i]) < Color(cards[j])
		}
		return true
	})
	lastIsRule, lastRuleType := RulePopEnable(lastCards)
	isRule, ruleType := RulePopEnable(cards)
	//fmt.Println("isRule：", isRule, " ruleType:", ruleType)
	if isRule && lastIsRule {
		switch ruleType {
		case Single: //单张只能压单张
			if lastRuleType == Single {
				lastCard := lastCards[0]
				card := cards[0]
				if Value(card) > Value(lastCard) {
					return true, isBomb, bombScore
				} else if Value(card) == Value(lastCard) {
					return Color(card) > Color(lastCard), isBomb, bombScore
				}
			}
		case Twin: //对子只能压对子
			if lastRuleType == Twin {
				lastCard := lastCards[len(lastCards)-1]
				card := cards[len(cards)-1]
				if Value(card) > Value(lastCard) {
					return true, isBomb, bombScore
				} else if Value(card) == Value(lastCard) {
					return Color(card) > Color(lastCard), isBomb, bombScore
				}
			}
		case Straight: //非同花顺子只能压非同花顺子
			if len(cards) == len(lastCards) && lastRuleType == Straight {
				lastCard := lastCards[len(lastCards)-1]
				card := cards[len(cards)-1]
				if Value(card) > Value(lastCard) {
					return true, isBomb, bombScore
				} else if Value(card) == Value(lastCard) {
					return Color(card) > Color(lastCard), isBomb, bombScore
				}
			}
		case ColorStraight: //同花顺子可压： 同花顺、非同花顺子
			if len(cards) == len(lastCards) && (lastRuleType == ColorStraight || lastRuleType == Straight) {
				lastCard := lastCards[len(lastCards)-1]
				card := cards[len(cards)-1]
				if Value(card) > Value(lastCard) {
					return true, isBomb, bombScore
				} else if Value(card) == Value(lastCard) {
					return Color(card) > Color(lastCard), isBomb, bombScore
				}
			}
		case Straight_Twin: //连对特殊处理下！！！
			twinNum := len(cards)
			switch twinNum {
			case 4: //二连对只能压二连对
				if len(cards) == len(lastCards) && lastRuleType == Straight_Twin {
					lastCard := lastCards[len(lastCards)-1]
					card := cards[len(cards)-1]
					if Value(card) > Value(lastCard) {
						return true, isBomb, bombScore
					} else if Value(card) == Value(lastCard) {
						return Color(card) > Color(lastCard), isBomb, bombScore
					}
				}
			case 6: //三连对可压：三连对、单2
				lastCard := lastCards[len(lastCards)-1]
				if lastRuleType == Straight_Twin {
					if len(lastCards) == 6 { //三连对
						score := Score15
						if toEnd {
							score = Score2End15
						}
						card := cards[len(cards)-1]
						if Value(card) > Value(lastCard) {
							return true, true, score
						} else if Value(card) == Value(lastCard) {
							return Color(card) > Color(lastCard), true, score
						}
					}
				} else if lastRuleType == Single && Value(lastCard) == Card_Value_2 { //单2
					score := Score10
					if lastCard == 51 || lastCard == 38 { //红2 10分
						if toEnd {
							score = Score2End10
						}
					} else if lastCard == 25 || lastCard == 12 { //黑2  5分
						score = Score5
						if toEnd {
							score = Score2End5
						}
					}
					return true, true, score
				}
			case 8: //四连对可压：四连对、三连对、对子2、单2
				lastCard := lastCards[len(lastCards)-1]
				if lastRuleType == Straight_Twin {
					if len(lastCards) == 8 { //四连对
						score := Score20
						if toEnd {
							score = Score2End20
						}
						card := cards[len(cards)-1]
						if Value(card) > Value(lastCard) {
							return true, true, score
						} else if Value(card) == Value(lastCard) {
							return Color(card) > Color(lastCard), true, score
						}
					} else if len(lastCards) == 6 { //三连对
						score := Score15
						if toEnd {
							score = Score2End15
						}
						return true, true, score
					}
				} else if lastRuleType == Twin && Value(lastCard) == Card_Value_2 { //对子2
					tmpScore := int32(0)
					for _, card := range lastCards {
						if card == 51 || card == 38 { //红2 10分
							if toEnd {
								tmpScore += Score2End10
							} else {
								tmpScore += Score10
							}
						} else if card == 25 || card == 12 { //黑2  5分
							if toEnd {
								tmpScore += Score2End5
							} else {
								tmpScore += Score5
							}
						}
					}
					return true, true, tmpScore
				} else if lastRuleType == Single && Value(lastCard) == Card_Value_2 { //单2
					score := Score10
					if lastCard == 51 || lastCard == 38 { //红2 10分
						if toEnd {
							score = Score2End10
						}
						return true, true, score
					} else if lastCard == 25 || lastCard == 12 { //黑2  5分
						score = Score5
						if toEnd {
							score = Score2End5
						}
						return true, true, score
					}
				}
			case 10: //五连对可压：五连对、四连对、三连对、对子2、单2
				lastCard := lastCards[len(lastCards)-1]
				if lastRuleType == Straight_Twin {
					if len(lastCards) == 10 { //五连对
						score := Score25
						if toEnd {
							score = Score2End25
						}
						card := cards[len(cards)-1]
						if Value(card) > Value(lastCard) {
							return true, true, score
						} else if Value(card) == Value(lastCard) {
							return Color(card) > Color(lastCard), true, score
						}
					} else if len(lastCards) == 8 { //四连对
						score := Score20
						if toEnd {
							score = Score2End20
						}
						return true, true, score
					} else if len(lastCards) == 6 { //三连对
						score := Score15
						if toEnd {
							score = Score2End15
						}
						return true, true, score
					}
				} else if lastRuleType == Twin && Value(lastCard) == Card_Value_2 { //对子2
					tmpScore := int32(0)
					for _, card := range lastCards {
						if card == 51 || card == 38 { //红2 10分
							if toEnd {
								tmpScore += Score2End10
							} else {
								tmpScore += Score10
							}
						} else if card == 25 || card == 12 { //黑2  5分
							if toEnd {
								tmpScore += Score2End5
							} else {
								tmpScore += Score5
							}
						}
					}
					return true, true, tmpScore
				} else if lastRuleType == Single && Value(lastCard) == Card_Value_2 { //单2
					score := Score10
					if lastCard == 51 || lastCard == 38 { //红2 10分
						if toEnd {
							score = Score2End10
						}
						return true, true, score
					} else if lastCard == 25 || lastCard == 12 { //黑2  5分
						score = Score5
						if toEnd {
							score = Score2End5
						}
						return true, true, score
					}
				}
			}
		case Triple: //三张只能压三张
			if lastRuleType == Triple {
				lastCard := lastCards[0]
				card := cards[0]
				if Value(card) > Value(lastCard) {
					return true, false, 0
				} else if Value(card) == Value(lastCard) {
					return Color(card) > Color(lastCard), false, 0
				}
			}
		case Four_Bomb: //炸弹可压：炸弹、三连对、对子2、单2
			score := Score2End20
			if lastRuleType == Four_Bomb { //炸弹
				if !toEnd {
					score = Score20
				}
				lastCard := lastCards[0]
				card := cards[0]
				if Value(card) > Value(lastCard) {
					return true, true, score
				}
			} else if lastRuleType == Straight_Twin {
				if len(lastCards) == 6 { //三连对
					if !toEnd {
						score = Score15
					}
					return true, true, score
				}
			} else if lastRuleType == Twin && Value(lastCards[0]) == Card_Value_2 {
				if !toEnd {
					tmpScore := int32(0)
					for _, card := range lastCards {
						if card == 51 || card == 38 { //红2 10分
							tmpScore += Score10
						} else if card == 25 || card == 12 { //黑2  5分
							tmpScore += Score5
						}
					}
					return true, true, tmpScore
				}
				return true, true, score
			} else if lastRuleType == Single && Value(lastCards[0]) == Card_Value_2 {
				if lastCards[0] == 51 || lastCards[0] == 38 { //红2 10分
					if !toEnd {
						score = Score10
					}
					return true, true, score
				} else if lastCards[0] == 25 || lastCards[0] == 12 { //黑2  5分
					if !toEnd {
						score = Score5
					}
					return true, true, score
				}
			}
		}
	}
	return false, false, 0
}

// return :可压住,是否炸，炸多少分
func CanDel_yl(lastCards, cards []int32, toEnd bool) (bool, bool, int32) {
	isBomb := false
	bombScore := int32(0)
	if len(lastCards) == 0 || len(cards) == 0 {
		return false, isBomb, bombScore
	}
	for _, card := range lastCards {
		if card == InvalideCard {
			return false, isBomb, bombScore
		}
	}
	for _, card := range cards {
		if card == InvalideCard {
			return false, isBomb, bombScore
		}
	}
	sort.Slice(lastCards, func(i, j int) bool {
		if Value(lastCards[i]) > Value(lastCards[j]) {
			return false
		} else if Value(lastCards[i]) == Value(lastCards[j]) {
			return Color(lastCards[i]) < Color(lastCards[j])
		}
		return true
	})
	sort.Slice(cards, func(i, j int) bool {
		if Value(cards[i]) > Value(cards[j]) {
			return false
		} else if Value(cards[i]) == Value(cards[j]) {
			return Color(cards[i]) < Color(cards[j])
		}
		return true
	})
	lastIsRule, lastRuleType := RulePopEnable_yl(lastCards)
	isRule, ruleType := RulePopEnable_yl(cards)
	fmt.Println("CanDel_yl: isRule：", isRule, " ruleType:", ruleType)
	if isRule && lastIsRule {
		switch ruleType {
		case Single: //单张只能压单张
			if lastRuleType == Single {
				lastCard := lastCards[0]
				card := cards[0]
				if Value(card) > Value(lastCard) {
					return true, isBomb, bombScore
				} else if Value(card) == Value(lastCard) {
					return Color(card) > Color(lastCard), isBomb, bombScore
				}
			}
		case Twin: //对子只能压对子
			if lastRuleType == Twin {
				lastCard := lastCards[len(lastCards)-1]
				card := cards[len(cards)-1]
				if Value(card) > Value(lastCard) {
					return true, isBomb, bombScore
				} else if Value(card) == Value(lastCard) {
					return Color(card) > Color(lastCard), isBomb, bombScore
				}
			}
		case Straight: //非同花顺子只能压非同花顺子
			if len(cards) == len(lastCards) && lastRuleType == Straight {
				lastCard := lastCards[len(lastCards)-1]
				card := cards[len(cards)-1]
				if Value(card) > Value(lastCard) {
					return true, isBomb, bombScore
				} else if Value(card) == Value(lastCard) {
					return Color(card) > Color(lastCard), isBomb, bombScore
				}
			}
		case ColorStraight: //同花顺子可压： 同花顺、非同花顺子
			if len(cards) == len(lastCards) && (lastRuleType == ColorStraight || lastRuleType == Straight) {
				lastCard := lastCards[len(lastCards)-1]
				card := cards[len(cards)-1]
				if Value(card) > Value(lastCard) {
					return true, isBomb, bombScore
				} else if Value(card) == Value(lastCard) {
					return Color(card) > Color(lastCard), isBomb, bombScore
				}
			}
		case Straight_Twin: //连对特殊处理下！！！
			twinNum := len(cards)
			switch twinNum {
			case 4: //二连对只能压二连对
				if len(cards) == len(lastCards) && lastRuleType == Straight_Twin {
					lastCard := lastCards[len(lastCards)-1]
					card := cards[len(cards)-1]
					if Value(card) > Value(lastCard) {
						return true, isBomb, bombScore
					} else if Value(card) == Value(lastCard) {
						return Color(card) > Color(lastCard), isBomb, bombScore
					}
				}
			case 6: //三连对可压：三连对、单2
				lastCard := lastCards[len(lastCards)-1]
				if lastRuleType == Straight_Twin {
					if len(lastCards) == 6 { //三连对
						score := Score15
						if toEnd {
							score = Score2End15
						}
						card := cards[len(cards)-1]
						if Value(card) > Value(lastCard) {
							return true, true, score
						} else if Value(card) == Value(lastCard) {
							return Color(card) > Color(lastCard), true, score
						}
					}
				} else if lastRuleType == Single && Value(lastCard) == Card_Value_2 { //单2
					score := Score10
					if lastCard == 51 || lastCard == 38 { //红2 10分
						if toEnd {
							score = Score2End10
						}
						return true, true, score
					} else if lastCard == 25 || lastCard == 12 { //黑2  5分
						score = Score5
						if toEnd {
							score = Score2End5
						}
						return true, true, score
					}
				}
			case 8: //四连对可压：四连对、三连对、对子2、单2
				lastCard := lastCards[len(lastCards)-1]
				if lastRuleType == Straight_Twin {
					if len(lastCards) == 8 { //四连对
						score := Score20
						if toEnd {
							score = Score2End20
						}
						card := cards[len(cards)-1]
						if Value(card) > Value(lastCard) {
							return true, true, score
						} else if Value(card) == Value(lastCard) {
							return Color(card) > Color(lastCard), true, score
						}
					} else if len(lastCards) == 6 { //三连对
						score := Score15
						if toEnd {
							score = Score2End15
						}
						return true, true, score
					}
				} else if lastRuleType == Twin && Value(lastCard) == Card_Value_2 { //对子2
					tmpScore := int32(0)
					for _, card := range lastCards {
						if card == 51 || card == 38 { //红2 10分
							if toEnd {
								tmpScore += Score2End10
							} else {
								tmpScore += Score10
							}
						} else if card == 25 || card == 12 { //黑2  5分
							if toEnd {
								tmpScore += Score2End5
							} else {
								tmpScore += Score5
							}
						}
					}
					return true, true, tmpScore
				} else if lastRuleType == Single && Value(lastCard) == Card_Value_2 { //单2
					score := Score10
					if lastCard == 51 || lastCard == 38 { //红2 10分
						if toEnd {
							score = Score2End10
						}
						return true, true, score
					} else if lastCard == 25 || lastCard == 12 { //黑2  5分
						score = Score5
						if toEnd {
							score = Score2End5
						}
						return true, true, score
					}
				}
			case 10: //五连对可压：五连对、四连对、三连对、对子2、单2
				lastCard := lastCards[len(lastCards)-1]
				if lastRuleType == Straight_Twin {
					if len(lastCards) == 10 { //五连对
						score := Score25
						if toEnd {
							score = Score2End25
						}
						card := cards[len(cards)-1]
						if Value(card) > Value(lastCard) {
							return true, true, score
						} else if Value(card) == Value(lastCard) {
							return Color(card) > Color(lastCard), true, score
						}
					} else if len(lastCards) == 8 { //四连对
						score := Score20
						if toEnd {
							score = Score2End20
						}
						return true, true, score
					} else if len(lastCards) == 6 { //三连对
						score := Score15
						if toEnd {
							score = Score2End15
						}
						return true, true, score
					}
				} else if lastRuleType == Twin && Value(lastCard) == Card_Value_2 { //对子2
					tmpScore := int32(0)
					for _, card := range lastCards {
						if card == 51 || card == 38 { //红2 10分
							if toEnd {
								tmpScore += Score2End10
							} else {
								tmpScore += Score10
							}
						} else if card == 25 || card == 12 { //黑2  5分
							if toEnd {
								tmpScore += Score2End5
							} else {
								tmpScore += Score5
							}
						}
					}
					return true, true, tmpScore
				} else if lastRuleType == Single && Value(lastCard) == Card_Value_2 { //单2
					score := Score10
					if lastCard == 51 || lastCard == 38 { //红2 10分
						if toEnd {
							score = Score2End10
						}
						return true, true, score
					} else if lastCard == 25 || lastCard == 12 { //黑2  5分
						score = Score5
						if toEnd {
							score = Score2End5
						}
						return true, true, score
					}
				}
			}
		case Triple: //三张只能压三张
			if lastRuleType == Triple {
				lastCard := lastCards[0]
				card := cards[0]
				if Value(card) > Value(lastCard) {
					return true, false, 0
				} else if Value(card) == Value(lastCard) {
					return Color(card) > Color(lastCard), false, 0
				}
			}
		case Straight_Triple: //三顺只能压三顺
			if lastRuleType == Straight_Triple {
				lastCard := lastCards[0]
				card := cards[0]
				if Value(card) > Value(lastCard) {
					return true, false, 0
				} else if Value(card) == Value(lastCard) {
					return Color(card) > Color(lastCard), false, 0
				}
			}
		case Plane_Twin: //飞机带双只能压飞机带双
			if lastRuleType == Plane_Twin {
				lastCardValue := FindPlaneTwinHead(lastCards)
				cardValue := FindPlaneTwinHead(cards)
				if cardValue > lastCardValue {
					return true, false, 0
				} else {
					return false, false, 0
				}
			}
		case Four_Bomb: //炸弹可压：炸弹、飞机带双、三顺、三张、顺子、三连对以下的对子、单张
			score := Score20
			if toEnd {
				score = Score2End20
			}
			switch lastRuleType {
			case Four_Bomb:
				lastCard := lastCards[0]
				card := cards[0]
				if Value(card) > Value(lastCard) {
					return true, true, score
				}
			case Plane_Twin, Straight_Triple, Triple, Straight, ColorStraight, Twin, Single:
				return true, true, score
			case Straight_Twin:
				if len(lastCards) == 6 || len(lastCards) == 4 { // 三连对、二连对
					return true, true, score
				}
			}
		}
	}
	return false, false, 0
}

func DelSliceInt32(sl []int32, v int32) []int32 {
	index := -1
	for key, value := range sl {
		if value == v {
			index = key
			break
		}
	}
	if index != -1 {
		sl = append(sl[:index], sl[index+1:]...)
	}
	return sl
}

// 计算输家输分数
func GetLoseScore(cards [HandCardNum]int32, toEnd bool) int32 {
	loseScore := int32(0)
	cpCards := []int32{}
	for _, card := range cards {
		if card != InvalideCard {
			cpCards = append(cpCards, card)
		}
	}

	//找单2
	find2Num := 0
	for i := 0; i < len(cpCards); i++ {
		card := cpCards[i]
		switch card {
		case HongTao2, FangPian2:
			if toEnd {
				loseScore += Score2End10
			} else {
				loseScore += Score10
			}
			cpCards = append(cpCards[:i], cpCards[i+1:]...)
			i--
			find2Num++
		case MeiHua2, HeiTao2:
			if toEnd {
				loseScore += Score2End5
			} else {
				loseScore += Score5
			}
			cpCards = append(cpCards[:i], cpCards[i+1:]...)
			i--
			find2Num++
		}
	}
	if find2Num >= 4 { //找到4个2？？
		return 0
	}

	// 找炸弹
	mapBomb := make(map[int32]int)
	for _, card := range cpCards {
		mapBomb[int32(Value(card))]++
	}
	bombValue := []int32{}
	for card, num := range mapBomb {
		if num == 4 {
			bombValue = append(bombValue, card)
		}
	}
	if len(bombValue) > 0 {
		score := Score20
		if toEnd {
			score = Score2End20
		}
		for _, card := range bombValue {
			for i := int32(0); i < 4; i++ {
				cpCards = DelSliceInt32(cpCards, card+(i*PER_CARD_COLOR_MAX))
			}
		}
		loseScore += int32(len(bombValue)) * score
	}

	//找五连
	if len(cpCards) >= 10 {
		sort.Slice(cpCards, func(i, j int) bool {
			if cpCards[i] > cpCards[j] {
				return false
			}
			return true
		})
		map5STwin := make(map[int32]int32)
		for _, card := range cpCards {
			map5STwin[int32(Value(card))]++
		}
		Value5STwin := []int32{}
		for card, num := range map5STwin {
			if num >= 2 {
				Value5STwin = append(Value5STwin, card)
			}
		}
		sort.Slice(Value5STwin, func(i, j int) bool {
			if Value5STwin[i] > Value5STwin[j] {
				return false
			}
			return true
		})
		if len(Value5STwin) > 0 {
			tmpPairs := FindOneStraightWithWidth(5, Value5STwin)
			if len(tmpPairs) > 0 {
				for _, card := range tmpPairs {
					delC := 0
					for i := 0; i < len(cpCards); i++ {
						if cpCards[i] == InvalideCard {
							continue
						}
						if int32(Value(cpCards[i])) == card && (delC < 2) { //删除2次
							delC++
							cpCards = append(cpCards[:i], cpCards[i+1:]...)
							i--
							continue
						}
					}
				}
				score := Score25
				if toEnd {
					score = Score2End25
				}
				loseScore += score
			}
		}
	}

	//找四连
	if len(cpCards) >= 8 {
		sort.Slice(cpCards, func(i, j int) bool {
			if cpCards[i] > cpCards[j] {
				return false
			}
			return true
		})
		map4STwin := make(map[int32]int32)
		for _, card := range cpCards {
			map4STwin[int32(Value(card))]++
		}
		Value4STwin := []int32{}
		for card, num := range map4STwin {
			if num >= 2 {
				Value4STwin = append(Value4STwin, card)
			}
		}
		sort.Slice(Value4STwin, func(i, j int) bool {
			if Value4STwin[i] > Value4STwin[j] {
				return false
			}
			return true
		})
		if len(Value4STwin) > 0 {
			tmpPairs := FindOneStraightWithWidth(4, Value4STwin)
			if len(tmpPairs) > 0 {
				for _, card := range tmpPairs {
					delC := 0
					for i := 0; i < len(cpCards); i++ {
						if cpCards[i] != InvalideCard {
							if int32(Value(cpCards[i])) == card && (delC < 2) { //删除2次
								delC++
								cpCards = append(cpCards[:i], cpCards[i+1:]...)
								i--
								continue
							}
						}
					}
				}
				score := Score20
				if toEnd {
					score = Score2End20
				}
				loseScore += score
			}
		}
	}

	//找三连
	if len(cpCards) >= 6 {
		sort.Slice(cpCards, func(i, j int) bool {
			if cpCards[i] > cpCards[j] {
				return false
			}
			return true
		})
		map3STwin := make(map[int32]int32)
		for _, card := range cpCards {
			map3STwin[int32(Value(card))]++
		}
		Value3STwin := []int32{}
		for card, num := range map3STwin {
			if num >= 2 {
				Value3STwin = append(Value3STwin, card)
			}
		}
		sort.Slice(Value3STwin, func(i, j int) bool {
			if Value3STwin[i] > Value3STwin[j] {
				return false
			}
			return true
		})
		if len(Value3STwin) > 0 {
			tmpPairs := FindOneStraightWithWidth(3, Value3STwin)
			if len(tmpPairs) > 0 {
				for _, card := range tmpPairs {
					delC := 0
					for i := 0; i < len(cpCards); i++ {
						if cpCards[i] == InvalideCard {
							continue
						}
						if int32(Value(cpCards[i])) == card && (delC < 2) { //删除2次
							delC++
							cpCards = append(cpCards[:i], cpCards[i+1:]...)
							i--
							continue
						}
					}
				}
				score := Score15
				if toEnd {
					score = Score2End15
				}
				loseScore += score
			}
		}
	}

	//再找三连
	if len(cpCards) >= 6 {
		sort.Slice(cpCards, func(i, j int) bool {
			if cpCards[i] > cpCards[j] {
				return false
			}
			return true
		})
		map3STwin := make(map[int32]int32)
		for _, card := range cpCards {
			map3STwin[int32(Value(card))]++
		}
		Value3STwin := []int32{}
		for card, num := range map3STwin {
			if num >= 2 {
				Value3STwin = append(Value3STwin, card)
			}
		}
		sort.Slice(Value3STwin, func(i, j int) bool {
			if Value3STwin[i] > Value3STwin[j] {
				return false
			}
			return true
		})
		if len(Value3STwin) > 0 {
			tmpPairs := FindOneStraightWithWidth(3, Value3STwin)
			if len(tmpPairs) > 0 {
				for _, card := range tmpPairs {
					delC := 0
					for i := 0; i < len(cpCards); i++ {
						if cpCards[i] == InvalideCard {
							continue
						}
						if int32(Value(cpCards[i])) == card && (delC < 2) { //删除2次
							delC++
							cpCards = append(cpCards[:i], cpCards[i+1:]...)
							i--
							continue
						}
					}
				}
				score := Score15
				if toEnd {
					score = Score2End15
				}
				loseScore += score
			}
		}
	}
	if !toEnd {
		//找单张
		for _, card := range cpCards {
			if card != InvalideCard {
				loseScore++
			}
		}
	}
	return loseScore
}

// 找一个固定长度的顺子 (2,3,4,5,6) 3-> [2,3,4]
func FindOneStraightWithWidth(n int, pairs []int32) []int32 {
	if len(pairs) == 0 {
		return nil
	}
	lastKey := pairs[0]
	tempPair := []int32{lastKey}
	if n == 1 {
		return tempPair
	}
	for i := 1; i < len(pairs); i++ {
		if pairs[i]-lastKey == 1 {
			tempPair = append(tempPair, pairs[i])
		} else {
			tempPair = []int32{pairs[i]}
		}
		if len(tempPair) == n {
			break
		}
		lastKey = pairs[i]
	}
	if len(tempPair) == n {
		return tempPair
	}
	return nil
}

// 找多个固定长度的顺子(2,3,4,5,6) 3-> [[2,3,4][3,4,5][4,5,6]]
func FindStraightWithWidth(n int, pairs []int32) [][]int32 {
	if len(pairs) == 0 || n < 2 {
		return nil
	}
	var tempPairs [][]int32
	lastKey := pairs[0]
	tempPair := []int32{lastKey}
	for i := 1; i < len(pairs); i++ {
		if pairs[i]-lastKey == 1 {
			tempPair = append(tempPair, pairs[i])
		} else {
			tempPair = []int32{pairs[i]}
		}
		if len(tempPair) == n {
			tempPairs = append(tempPairs, tempPair)
			tempPair = []int32{}
			for j := n - 1; j > 0; j-- {
				tempPair = append(tempPair, pairs[i-j+1])
			}
		}
		lastKey = pairs[i]
	}
	return tempPairs
}

// ///////////////////////////(天胡牌型)///////////////////////
// 判断13张牌里面是否有4个2炸弹
func Have2FourBomb(cards []int32) bool {
	if len(cards) == Hand_CardNum {
		for i := 0; i < len(cards); i++ {
			if Value(cards[i]) == Card_Value_2 {
				tmpInt := 0
				for j := 0; j < len(cards); j++ {
					if Value(cards[i]) == Value(cards[j]) {
						tmpInt++
					}
				}
				if tmpInt == 4 {
					return true
				}
			}
		}
	}
	return false
}

// 判断13张牌里面是否有6连对（单牌2不能带到连牌里）
func Have6StraightTwin(cards []int32) bool {
	cpCards := []int32{}
	for _, card := range cards {
		if int32(Value(card)) != InvalideCard {
			cpCards = append(cpCards, card)
		}
	}
	if len(cpCards) == Hand_CardNum {
		sort.Slice(cpCards, func(i, j int) bool {
			if cpCards[i] > cpCards[j] {
				return false
			}
			return true
		})
		map6STwin := make(map[int32]int32)
		for _, card := range cpCards {
			if Value(card) != Card_Value_2 {
				map6STwin[int32(Value(card))]++
			}
		}
		if len(map6STwin) > 0 {
			Value6STwin := []int32{}
			for card, num := range map6STwin {
				if num >= 2 {
					Value6STwin = append(Value6STwin, card)
				}
			}
			if len(Value6STwin) > 0 {
				sort.Slice(Value6STwin, func(i, j int) bool {
					if Value6STwin[i] > Value6STwin[j] {
						return false
					}
					return true
				})
				tmpPairs := FindOneStraightWithWidth(6, Value6STwin)
				if len(tmpPairs) == 6 {
					return true
				}
			}
		}
	}
	return false
}

// 判断13张牌里面是否有12顺（单牌2不能带到连牌里）
func Have12Straight(cards []int32) bool {
	if len(cards) == Hand_CardNum {
		mapCard := make(map[int32]int)
		for _, card := range cards {
			if Value(card) != Card_Value_2 {
				mapCard[int32(Value(card))]++
			}
		}
		if len(mapCard) == 12 {
			return true
		}
	}
	return false
}

// 根据最小牌值去推荐出牌
func RecommendCardsWithMinCard(cards []int32) []int32 {
	recmCards := []int32{}
	//排序
	sort.Slice(cards, func(i, j int) bool {
		v_i := Value(cards[i])
		v_j := Value(cards[j])
		c_i := Color(cards[i])
		c_j := Color(cards[j])
		if v_i > v_j {
			return false
		} else if v_i == v_j {
			return c_i < c_j
		}
		return true
	})
	//取手牌
	cpCards := []int32{}
	for _, card := range cards {
		if card != InvalideCard {
			cpCards = append(cpCards, card)
		}
	}
	//取最小牌值
	minCard := cpCards[0]
	for _, card := range cpCards {
		minCard = card
		break
	}
	//找顺子
	recmCards = []int32{}
	findNum := 0
	for _, card := range cpCards {
		if Value(card) == Value(minCard)+findNum && Value(card) != Card_Value_2 {
			recmCards = append(recmCards, card)
			findNum++
		} else {
			break
		}
	}
	if len(recmCards) < 3 {
		//找二连对
		recmCards = []int32{}
		find1 := 0
		find2 := 0
		for _, card := range cpCards {
			if Value(card) == Value(minCard) && Value(card) != Card_Value_2 && find1 < 2 {
				recmCards = append(recmCards, card)
				find1++
			}
			if Value(card) == Value(minCard)+1 && Value(card) != Card_Value_2 && find2 < 2 {
				recmCards = append(recmCards, card)
				find2++
			}
		}
		if len(recmCards) != 4 {
			//找三张
			recmCards = []int32{}
			find3 := 0
			for _, card := range cpCards {
				if Value(card) == Value(minCard) && Value(card) != Card_Value_2 && find3 < 3 {
					recmCards = append(recmCards, card)
					find3++
				}
			}
			if len(recmCards) != 3 {
				//找对子
				recmCards = []int32{}
				find4 := 0
				for _, card := range cpCards {
					if Value(card) == Value(minCard) && Value(card) != Card_Value_2 && find4 < 2 {
						recmCards = append(recmCards, card)
						find4++
					}
				}
				if len(recmCards) != 2 {
					// 找单张
					recmCards = []int32{}
					recmCards = append(recmCards, minCard)
				}
			}
		}
	}
	if len(recmCards) == 0 {
		recmCards = append(recmCards, minCard)
	}
	return recmCards
}

// 根据牌型牌数量最多去推荐出牌
func RecommendCardsWithCards(cards []int32) []int32 {
	recmCards := []int32{}
	//排序
	sort.Slice(cards, func(i, j int) bool {
		v_i := Value(cards[i])
		v_j := Value(cards[j])
		c_i := Color(cards[i])
		c_j := Color(cards[j])
		if v_i > v_j {
			return false
		} else if v_i == v_j {
			return c_i < c_j
		}
		return true
	})
	//取手牌
	cpCards := []int32{}
	for _, card := range cards {
		if card != InvalideCard {
			cpCards = append(cpCards, card)
		}
	}
	//取最小牌值
	minCard := cpCards[0]
	for _, card := range cpCards {
		minCard = card
		break
	}
	if IsStraight(cpCards) {
		//5顺直接返回
		for _, card := range cpCards {
			recmCards = append(recmCards, card)
		}
	} else {
		if len(recmCards) != 4 {
			//找4顺
			map4Straight := make(map[int32]int32)
			for _, card := range cpCards {
				if Value(card) == Card_Value_2 {
					continue
				}
				map4Straight[int32(Value(card))]++
			}
			if len(map4Straight) > 0 {
				Value4Straight := []int32{}
				for cardV, num := range map4Straight {
					if num >= 1 {
						Value4Straight = append(Value4Straight, cardV)
					}
				}
				sort.Slice(Value4Straight, func(i, j int) bool {
					if Value4Straight[i] > Value4Straight[j] {
						return false
					}
					return true
				})
				if len(Value4Straight) > 0 {
					tmpPairs := FindOneStraightWithWidth(4, Value4Straight)
					if len(tmpPairs) > 0 {
						for _, cardV := range tmpPairs {
							delC := 0
							for i := 0; i < len(cpCards); i++ {
								if int32(Value(cpCards[i])) == cardV && (delC < 1) { //找1次
									delC++
									recmCards = append(recmCards, cpCards[i])
									continue
								}
							}
						}
					}
				}
			}
		}
		if len(recmCards) != 4 {
			//找炸弹
			mapBomb := make(map[int32]int)
			for _, card := range cpCards {
				mapBomb[int32(Value(card))]++
			}
			if len(mapBomb) > 0 {
				bombValue := []int32{}
				for card, num := range mapBomb {
					if num == 4 {
						bombValue = append(bombValue, card)
						break //找一个炸弹就行
					}
				}
				if len(bombValue) > 0 {
					for _, card := range bombValue {
						for _, cpCard := range cpCards {
							if int32(Value(cpCard)) == card {
								recmCards = append(recmCards, cpCard)
							}
						}
					}
				}
			}
		}
		if len(recmCards) != 4 {
			//找二连对
			map2STwin := make(map[int32]int32)
			for _, card := range cpCards {
				if Value(card) == Card_Value_2 {
					continue
				}
				map2STwin[int32(Value(card))]++
			}
			if len(map2STwin) > 0 {
				Value2STwin := []int32{}
				for card, num := range map2STwin {
					if num >= 2 {
						Value2STwin = append(Value2STwin, card)
					}
				}
				if len(Value2STwin) > 0 {
					sort.Slice(Value2STwin, func(i, j int) bool {
						if Value2STwin[i] > Value2STwin[j] {
							return false
						}
						return true
					})
					tmpPairs := FindOneStraightWithWidth(2, Value2STwin)
					if len(tmpPairs) > 0 {
						for _, card := range tmpPairs {
							delC := 0
							for i := 0; i < len(cpCards); i++ {
								if int32(Value(cpCards[i])) == card && (delC < 2) { //找2次
									delC++
									recmCards = append(recmCards, cpCards[i])
									continue
								}
							}
						}
					}
				}
			}
		}
		if len(recmCards) != 4 {
			//找3顺
			map3Straight := make(map[int32]int32)
			for _, card := range cpCards {
				if Value(card) == Card_Value_2 {
					continue
				}
				map3Straight[int32(Value(card))]++
			}
			if len(map3Straight) > 0 {
				Value3Straight := []int32{}
				for cardV, num := range map3Straight {
					if num >= 1 {
						Value3Straight = append(Value3Straight, cardV)
					}
				}
				sort.Slice(Value3Straight, func(i, j int) bool {
					if Value3Straight[i] > Value3Straight[j] {
						return false
					}
					return true
				})
				if len(Value3Straight) > 0 {
					tmpPairs := FindOneStraightWithWidth(3, Value3Straight)
					if len(tmpPairs) > 0 {
						for _, cardV := range tmpPairs {
							delC := 0
							for i := 0; i < len(cpCards); i++ {
								if int32(Value(cpCards[i])) == cardV && (delC < 1) { //找1次
									delC++
									recmCards = append(recmCards, cpCards[i])
									continue
								}
							}
						}
					}
				}
			}
		}
		if len(recmCards) != 3 {
			//找3张
			mapTriple := make(map[int32]int32)
			for _, card := range cpCards {
				mapTriple[int32(Value(card))]++
			}
			if len(mapTriple) > 0 {
				valueTriple := []int32{}
				for cardV, num := range mapTriple {
					if num >= 3 {
						valueTriple = append(valueTriple, cardV)
					}
				}
				if len(valueTriple) > 0 {
					tripleValue := valueTriple[0]
					delC := 0
					for i := 0; i < len(cpCards); i++ {
						if int32(Value(cpCards[i])) == tripleValue && (delC < 3) { //找3次
							delC++
							recmCards = append(recmCards, cpCards[i])
							continue
						}
					}
				}
			}
		}
		if len(recmCards) != 3 {
			//找对子
			mapTwin := make(map[int32]int32)
			for _, card := range cpCards {
				if Value(card) == Card_Value_2 {
					continue
				}
				mapTwin[int32(Value(card))]++
			}
			if len(mapTwin) > 0 {
				ValueTwin := []int32{}
				for card, num := range mapTwin {
					if num >= 2 {
						ValueTwin = append(ValueTwin, card)
					}
				}
				if len(ValueTwin) > 0 {
					sort.Slice(ValueTwin, func(i, j int) bool {
						if ValueTwin[i] > ValueTwin[j] {
							return false
						}
						return true
					})
					vTwin := ValueTwin[0]
					recmCount := 0
					for _, card := range cpCards {
						if vTwin == int32(Value(card)) && recmCount < 2 {
							recmCards = append(recmCards, card)
							recmCount++
						}
					}
				}
			}
		}
	}
	if len(recmCards) == 0 {
		recmCards = append(recmCards, minCard)
	}
	return recmCards
}

// 根据上家牌去推荐出牌
func RecommendCardsWithLastCards(lastCards, cards []int32) []int32 {
	recmCards := []int32{}
	isHave := false
	isRule, ruleType := RulePopEnable(lastCards)
	cpCards := []int32{}
	for _, card := range cards {
		if card != InvalideCard {
			cpCards = append(cpCards, card)
		}
	}
	if isRule {
		switch ruleType {
		case Single:
			isHave, recmCards = needSingle(lastCards, cpCards)
			if !isHave {

			}
		case Twin:
			isHave, recmCards = needTwin(lastCards, cpCards)
			if !isHave {

			}
		case Straight:
			isHave, recmCards = needStraight(lastCards, cpCards)
			if !isHave {

			}
		case ColorStraight:
			isHave, recmCards = needColorStraight(lastCards, cpCards)
			if !isHave {

			}
		case Straight_Twin:
			isHave, recmCards = needStraightTwin(lastCards, cpCards)
			if !isHave {

			}
		case Triple:
			isHave, recmCards = needTriple(lastCards, cpCards)
			if !isHave {

			}
		case Four_Bomb:
			isHave, recmCards = needFourBomb(lastCards, cpCards)
			if !isHave {

			}
		}
	}
	return recmCards
}

// 需要单张
func needSingle(lastCards, cards []int32) (bool, []int32) {
	haveNeed := false
	needCards := []int32{}
	if len(cards) != 0 && IsSingle(lastCards) {
		sort.Slice(cards, func(i, j int) bool {
			if Value(cards[i]) > Value(cards[j]) {
				return false
			}
			return true
		})
		for _, card := range cards {
			if card != InvalideCard {
				if Value(card) < Value(lastCards[0]) {
					continue
				} else {
					if Value(card) == Value(lastCards[0]) {
						if Color(card) > Color(lastCards[0]) {
							haveNeed = true
							needCards = append(needCards, card)
							break
						}
					} else if Value(card) > Value(lastCards[0]) {
						haveNeed = true
						needCards = append(needCards, card)
						break
					}
				}
			}
		}
	}
	return haveNeed, needCards
}

// 需要对子
func needTwin(lastCards, cards []int32) (bool, []int32) {
	haveNeed := false
	needCards := []int32{}
	if len(cards) != 0 && IsTwin(lastCards) {
		mapTwin := make(map[int32]int32)
		for _, card := range cards {
			if card != InvalideCard {
				mapTwin[int32(Value(card))]++
			}
		}
		if len(mapTwin) > 0 {
			ValueTwin := []int32{}
			for card, num := range mapTwin {
				if num >= 2 {
					ValueTwin = append(ValueTwin, card)
				}
			}
			if len(ValueTwin) > 0 {
				sort.Slice(ValueTwin, func(i, j int) bool {
					if ValueTwin[i] > ValueTwin[j] {
						return false
					}
					return true
				})
				sort.Slice(lastCards, func(i, j int) bool {
					if lastCards[i] > lastCards[j] {
						return false
					}
					return true
				})
				needValue := -1
				for _, card := range ValueTwin {
					if int(card) < Value(lastCards[1]) {
						continue
					} else {
						if int(card) == Value(lastCards[1]) {
							if Color(card) > Color(lastCards[1]) {
								haveNeed = true
								needValue = int(card)
								break
							}
						} else if int(card) > Value(lastCards[1]) {
							haveNeed = true
							needValue = int(card)
							break
						}
					}
				}
				if needValue != -1 {
					sort.Slice(cards, func(i, j int) bool {
						if cards[i] > cards[j] {
							return false
						}
						return true
					})
					needNum := 0
					for i := len(cards) - 1; i >= 0; i-- {
						if Value(cards[i]) == needValue && needNum < 2 {
							needNum++
							needCards = append(needCards, cards[i])
						}
					}
				}
			}
		}
	}
	return haveNeed, needCards
}

// 需要顺子
func needStraight(lastCards, cards []int32) (bool, []int32) {
	haveNeed := false
	needCards := []int32{}
	if len(cards) != 0 && IsStraight(lastCards) {
		sort.Slice(lastCards, func(i, j int) bool {
			if lastCards[i] > lastCards[j] {
				return false
			}
			return true
		})
		sort.Slice(cards, func(i, j int) bool {
			if Value(cards[i]) > Value(cards[j]) {
				return false
			}
			return true
		})
		mapS := make(map[int]int32)
		for _, card := range cards {
			if card != InvalideCard {
				mapS[Value(card)]++
			}
		}
		if len(mapS) > 0 {
			ValueS := []int32{}
			for card, num := range mapS {
				if num >= 1 {
					ValueS = append(ValueS, int32(card))
				}
			}
			idx := len(lastCards)
			if len(ValueS) > 0 {
				sort.Slice(ValueS, func(i, j int) bool {
					if Value(ValueS[i]) > Value(ValueS[j]) {
						return false
					}
					return true
				})
				tmpSMap := FindStraightWithWidth(idx, ValueS)
				if len(tmpSMap) != 0 {
					needCard := int32(-1)
					for _, sCards := range tmpSMap {
						maxSCard := sCards[len(sCards)-1]
						for i := len(cards) - 1; i >= 0; i-- {
							if int32(Value(cards[i])) == maxSCard {
								card := cards[i]
								lastCard := lastCards[len(lastCards)-1]
								if Value(card) < Value(lastCard) {
									continue
								} else {
									if Value(card) == Value(lastCard) {
										if Color(card) > Color(lastCard) {
											haveNeed = true
											needCard = cards[i]
										}
									} else if Value(card) > Value(lastCards[1]) {
										haveNeed = true
										needCard = cards[i]
									}
								}
							}
							if haveNeed && needCard != -1 {
								break
							}
						}
						if haveNeed && needCard != -1 {
							break
						}
					}
					if haveNeed && needCard != -1 {
						needValue := Value(needCard)
						need := 0
						for j := len(cards) - 1; j >= 0; j-- {
							if needValue >= 0 && Value(cards[j]) == needValue && need < idx {
								needCards = append(needCards, cards[j])
								needValue--
								need++
							}
						}
					}
				}
			}
		}
	}
	return haveNeed, needCards
}

func needColorStraight(lastCards, cards []int32) (bool, []int32) {
	haveNeed := false
	needCards := []int32{}
	if len(cards) != 0 && IsColorStraight(lastCards) {
		sliceCS := [4][]int32{} //color-card
		for _, card := range cards {
			if card != InvalideCard {
				sliceCS[Color(card)] = append(sliceCS[Color(card)], card)
			}
		}
		for _, colorCards := range sliceCS {
			if len(colorCards) > 0 && len(colorCards) >= len(lastCards) {
				sort.Slice(colorCards, func(i, j int) bool {
					if Value(colorCards[i]) > Value(colorCards[j]) {
						return false
					}
					return true
				})
				tmpS := FindOneStraightWithWidth(len(lastCards), colorCards)
				if len(tmpS) > 0 {
					card := tmpS[len(tmpS)-1]
					lastCard := lastCards[len(lastCards)-1]
					if Value(card) > Value(lastCard) {
						haveNeed = true
						copy(needCards, tmpS)
					} else if Value(card) == Value(lastCard) {
						if Color(card) > Color(lastCard) {
							haveNeed = true
							copy(needCards, tmpS)
						}
					}
				}
			}
		}
	}
	return haveNeed, needCards
}

func needStraightTwin(lastCards, cards []int32) (bool, []int32) {
	haveNeed := false
	needCards := []int32{}
	if len(cards) != 0 && IsStraightTwin(lastCards) {
		sort.Slice(cards, func(i, j int) bool {
			if cards[i] > cards[j] {
				return false
			}
			return true
		})
		mapSTwin := make(map[int32]int32)
		for _, card := range cards {
			if card != InvalideCard {
				mapSTwin[int32(Value(card))]++
			}
		}
		if len(mapSTwin) > 0 {
			valueSTwin := []int32{}
			for card, num := range mapSTwin {
				if num >= 2 {
					valueSTwin = append(valueSTwin, card)
				}
			}
			if len(valueSTwin) > 0 {
				sort.Slice(valueSTwin, func(i, j int) bool {
					if valueSTwin[i] > valueSTwin[j] {
						return false
					}
					return true
				})
				tmpPairs := FindOneStraightWithWidth(len(lastCards)/2, valueSTwin)
				if len(tmpPairs) > 0 {
					for _, card := range tmpPairs {
						delC := 0
						for i := 0; i < len(cards); i++ {
							if int32(Value(cards[i])) == card && (delC < 2) { //删除2次
								delC++
								haveNeed = true
								needCards = append(needCards, cards[i])
								continue
							}
						}
					}
				}
			}
		}
	}
	return haveNeed, needCards
}

func needTriple(lastCards, cards []int32) (bool, []int32) {
	haveNeed := false
	needCards := []int32{}
	if len(cards) != 0 && IsTriple(lastCards) {
		mapTriple := make(map[int32]int32)
		for _, card := range cards {
			if card != InvalideCard {
				mapTriple[int32(Value(card))]++
			}
		}
		if len(mapTriple) > 0 {
			ValueTriple := []int32{}
			for card, num := range mapTriple {
				if num >= 3 {
					ValueTriple = append(ValueTriple, card)
				}
			}
			if len(ValueTriple) > 0 {
				lastCard := lastCards[0]
				for _, triple := range ValueTriple {
					if triple > lastCard {
						for _, card := range cards {
							if int32(Value(card)) == triple {
								haveNeed = true
								needCards = append(needCards, card)
							}
						}
					}
				}
			}
		}
	}
	return haveNeed, needCards
}
func needFourBomb(lastCards, cards []int32) (bool, []int32) {
	haveNeed := false
	needCards := []int32{}
	if len(cards) != 0 && IsFourBomb(lastCards) {
		mapBomb := make(map[int32]int32)
		for _, card := range cards {
			if card != InvalideCard {
				mapBomb[int32(Value(card))]++
			}
		}
		if len(mapBomb) > 0 {
			ValueBomb := []int32{}
			for card, num := range mapBomb {
				if num >= 4 {
					ValueBomb = append(ValueBomb, card)
				}
			}
			if len(ValueBomb) > 0 {
				lastCard := lastCards[0]
				for _, bomb := range ValueBomb {
					if bomb > lastCard {
						for _, card := range cards {
							if int32(Value(card)) == bomb {
								haveNeed = true
								needCards = append(needCards, card)
							}
						}
					}
				}
			}
		}
	}
	return haveNeed, needCards
}

// 根据上家牌型是否需要额外延迟出牌时间(s)
func NeedExDelay(lastCards []int32) bool {
	if IsStraightTwin(lastCards) && len(lastCards) >= 6 { //三连对、四连对、五连对
		return true
	}
	if IsFourBomb(lastCards) { //炸弹
		return true
	}
	//if IsStraight(lastCards) { //5张以上的顺子
	//	return true
	//}
	if IsPlaneTwin(lastCards) && len(lastCards) >= 10 { //飞机
		return true
	}
	return false
}
