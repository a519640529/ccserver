package luckydice

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func GetBotBetValue(baseScore int64) int64 {
	botWealth := rand.Intn(4999)
	if botWealth < 2000 { // poor
		botWealth = 0
	} else if botWealth < 4000 { // normal
		botWealth = 1
	} else { // rich
		botWealth = 2
	}
	betValues := BetValues[botWealth]
	betValue := betValues[rand.Intn(len(betValues))]
	if betValue == 0 {
		betValue = betValues[0] + rand.Int63n(betValues[len(betValues)-1]-betValues[0])
	}
	return betValue * baseScore
}

func GetBotBetSide() int32 {
	return rand.Int31n(2)
}

func GetDiceResult() (dr *DiceResult) {
	dr = &DiceResult{}
	dr.Update()
	return
}

type DiceResult struct {
	Dice1 int32
	Dice2 int32
	Dice3 int32
}

func (dr *DiceResult) BigSmall() int32 {
	if dr.Dice1+dr.Dice2+dr.Dice3 < 11 {
		return Small
	} else {
		return Big
	}
}

func (dr *DiceResult) Update() *DiceResult {
	dr.Dice1 = 1 + rand.Int31n(6)
	dr.Dice2 = 1 + rand.Int31n(6)
	dr.Dice3 = 1 + rand.Int31n(6)
	return dr
}
