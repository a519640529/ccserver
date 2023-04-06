package main

import (
	"games.yol.com/win88/common"
	//"games.yol.com/win88/gamerule/blackjack"
	//"games.yol.com/win88/gamerule/dezhoupoker"
	//"games.yol.com/win88/gamerule/fivecardstud"
	//"games.yol.com/win88/gamerule/omahapoker"
)

var minPlayGameNum = map[int]int{
	common.GameId_TenHalf:      2,
	common.GameId_DezhouPoker:  2,
	common.GameId_FiveCardStud: 2,
	common.GameId_BlackJack:    1,
	//common.GameId_OmahaPoker:   omahapoker.MinNumOfPlayer,
}

var maxPlayGameNum = map[int]int{
	//common.GameId_DezhouPoker:  int(dezhoupoker.MaxNumOfPlayer),
	//common.GameId_FiveCardStud: int(fivecardstud.MaxNumOfPlayer),
	//common.GameId_BlackJack:    blackjack.MaxPlayer,
	//common.GameId_OmahaPoker:   omahapoker.MaxNumOfPlayer,
}

func GetGameStartMinNum(gameid int) int {
	return minPlayGameNum[gameid]
}
func GetGameSuiableNum(gameid int, flag int32) int {
	minNum, maxNum := minPlayGameNum[gameid], maxPlayGameNum[gameid]
	if flag == MatchTrueMan_Forbid {
		if minNum == maxNum {
			return minNum
		} else {
			return maxNum - 1
		}
	} else {
		if minNum == maxNum {
			return minNum
		} else {
			return maxNum - 2
		}
	}
}
func IsRegularNum(gameid int) bool {
	return minPlayGameNum[gameid] == maxPlayGameNum[gameid]
}
