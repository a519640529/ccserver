package base

import (
	"encoding/json"
	"fmt"
	"time"

	"games.yol.com/win88/model"
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/task"
)

const (
	Warning_LoseCoinLimit = 1 //1. 单次输分超过设定值，默认值100000分
	Warning_CoinPoolZero  = 2 //2. 水位线低于0
	Warning_CoinPoolLow   = 3 //3. 水位线低于水位下限
	Warning_BlackPlayer   = 4 //4. 黑名单玩家加入游戏
	Warning_BetCoinMax    = 5 //5. 单个玩家下注超过5000元，可配置，默认值5000
	Warning_WinRate       = 6 //6. 实时赔率 = 玩家产出+1/玩家投入+1 ≥ 5 的玩家加入游戏，可配置，默认值5
	Warning_WinCoin       = 7 //7. 玩家获利 = 玩家产出 - 玩家投入 ≥ 5000元 加入游戏，可配置，默认值5000
)

type GameWarningParam struct {
	WarningType  int
	WarningGame  int
	WarningScene int
	WarningSnid  int
	WarningBet   int
	WarningCoin  int
	WarningRate  int
	WinCoin      int64
	LoseCoin     int64
}

var EmailTitle = "Game server warning!"

func NewGameWarning(param string) {
	timeNow := time.Now()
	task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
		gwParam := GameWarningParam{}
		err := json.Unmarshal([]byte(param), &gwParam)
		if err != nil {
			logger.Logger.Error("Unmarshal game warning param error:", err)
			return err
		}
		var emailContent string
		switch gwParam.WarningType {
		case Warning_LoseCoinLimit:
		case Warning_CoinPoolZero:
			emailContent = fmt.Sprintf("Game %v coin pool drop to zero level.", gwParam.WarningGame)
		case Warning_CoinPoolLow:
			emailContent = fmt.Sprintf("Game %v coin pool drop to low level.", gwParam.WarningGame)
		case Warning_BlackPlayer:
			emailContent = fmt.Sprintf("Player %v where in black list enter %v game.", gwParam.WarningSnid, gwParam.WarningGame)
		case Warning_BetCoinMax:
			emailContent = fmt.Sprintf("Player %v where in %v scene,bet big coin in %v game.",
				gwParam.WarningSnid, gwParam.WarningScene, gwParam.WarningGame)
		case Warning_WinRate:
		case Warning_WinCoin:
		default:
			return nil
		}
		emailContent = emailContent + "\n" + fmt.Sprintf("Time:%v", timeNow)
		println(EmailTitle, emailContent)
		//TODO send email
		//email.InitGmailFrom()
		return nil
	}), nil, "NewGameWarning").Start()
}
func WarningLoseCoin(gameFreeId int32, snid int32, loseCoin int64) {
	if model.GameParamData.WarningLoseLimit == 0 {
		return
	}
	if loseCoin < model.GameParamData.WarningLoseLimit {
		return
	}
	NewGameWarning(fmt.Sprintf(`{"WarningType":%v,"WarningGame":%v,"WarningSnid":%v,"LoseCoin":%v}`,
		Warning_LoseCoinLimit, gameFreeId, snid, loseCoin))
}
func WarningBetCoinCheck(sceneId, gameFreeId int32, snid int32, betCoin int64) {
	if model.GameParamData.WarningBetMax == 0 {
		return
	}
	if betCoin > model.GameParamData.WarningBetMax {
		NewGameWarning(fmt.Sprintf(`{"WarningType":%v,"WarningSnid":%v,"WarningGame":%v,"WarningScene":%v}`,
			Warning_BetCoinMax, snid, gameFreeId, sceneId))
	}
}
func WarningCoinPool(warnType int, gameFreeId int32) {
	NewGameWarning(fmt.Sprintf(`{"WarningType":%v,"WarningGame":%v}`,
		warnType, gameFreeId))
}
func WarningBlackPlayer(snid, gameFreeId int32) {
	NewGameWarning(fmt.Sprintf(`{"WarningType":%v,"WarningSnid":%v,"WarningGame":%v}`,
		Warning_BlackPlayer, snid, gameFreeId))
}
func WarningWinnerRate(snid int32, winCoin, loseCoin int64) {
	if model.GameParamData.WarningWinRate == 0 {
		return
	}
	if (winCoin+1)/(loseCoin+1) < model.GameParamData.WarningWinRate {
		return
	}
	NewGameWarning(fmt.Sprintf(`{"WarningType":%v,"WarningSnid":%v,"WarningRate":%v,"WinCoin":%v,"LoseCoin":%v}`,
		Warning_WinRate, snid, (winCoin+1)/(loseCoin+1), winCoin, loseCoin))
}
func WarningWinnerCoin(snid int32, winCoin, loseCoin int64) {
	if model.GameParamData.WarningWinMoney == 0 {
		return
	}
	if (winCoin - loseCoin) < model.GameParamData.WarningWinMoney {
		return
	}
	NewGameWarning(fmt.Sprintf(`{"WarningType":%v,"WarningSnid":%v,"WarningCoin":%v,"WinCoin":%v,"LoseCoin":%v}`,
		Warning_WinCoin, snid, (winCoin - loseCoin), winCoin, loseCoin))
}
