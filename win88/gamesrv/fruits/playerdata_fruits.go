package fruits

import (
	"games.yol.com/win88/gamerule/fruits"
	"games.yol.com/win88/gamesrv/base"
	"games.yol.com/win88/model"
)

var ElementsParams = [][]int32{}

type FruitsPlayerData struct {
	*base.Player
	roomid        int32 //房间ID
	result        *fruits.WinResult
	betIdx        int               //下注索引
	betCoin       int64             //下注金额
	oneBetCoin    int64             //单注
	cpCtx         model.CoinPoolCtx //水池环境
	leaveTime     int32             //离开时间
	winCoin       int64             //本局输赢
	freeTimes     int               //免费游戏
	maryFreeTimes int               //玛丽游戏
	gameState     int               //当前游戏模式
	///当局数值
	winLineRate       int64 //线倍率
	JackPotRate       int64 //获取奖池的百分比 5% 10% 15%
	JackPot7Rate      int64 //元素jackpot的倍率
	freeWinTotalRate  int64 //免费游戏赢的倍率
	maryWinTotalRate  int64 //玛丽游戏赢的倍率
	winNowJackPotCoin int64 //当局奖池爆奖
	winNowAllRate     int64 //当局赢得倍率
	//////////////////////////
	//统计数值
	winLineCoin    int64 //线赢金额
	winJackPotCoin int64 //爆奖
	winEle777Coin  int64 //元素数量
	winFreeCoin    int64 //免费游戏
	winMaryCoin    int64 //玛丽游戏
	winFreeTimes   int   //当局获得的免费次数
	//log
	nowFreeTimes int //当前第几轮免费游戏
	nowMaryTimes int //当前第几轮小玛丽游戏
	wl           []model.Classic777WinLine
	maryInNum    int //玛丽游戏连续数量
	taxCoin      int64
	//调控
	preWinRate   int64
	preNeedRate  int64
	pre30WinRate int64
	preEleVal    []int32
	//mary
	preMaryOutSide  int32
	preMaryMidArray []int32
	noWinTimes      int
	//
	startCoin int64
	testNum   int
}

func (p *FruitsPlayerData) init() {
	p.roomid = 0
	p.result = new(fruits.WinResult)
}
func (p *FruitsPlayerData) Clear() {
	p.gameState = fruits.Normal
	p.startCoin = p.Coin
	p.winCoin = 0
	p.JackPotRate = 0
	p.JackPot7Rate = 0
	p.winLineRate = 0
	p.maryWinTotalRate = 0
	p.freeWinTotalRate = 0
	p.winNowAllRate = 0
	p.winNowJackPotCoin = 0
	p.winFreeTimes = 0
	p.maryInNum = 0
	p.taxCoin = 0
	if p.freeTimes == 0 {
		p.winLineCoin = 0
		p.winFreeCoin = 0
		p.winJackPotCoin = 0
		p.winEle777Coin = 0
		p.nowFreeTimes = 0
	}
	if p.maryFreeTimes == 0 {
		p.winMaryCoin = 0
		p.nowMaryTimes = 0
	}
}

// 正常游戏 免费游戏
func (p *FruitsPlayerData) CreateResult(eleLineAppearRate [][]int32) {
	p.result.CreateLine(eleLineAppearRate)
	//if len(p.preEleVal) > 0 {
	//	p.result.EleValue = make([]int32, 15)
	//	copy(p.result.EleValue, p.preEleVal)
	//	p.preEleVal = nil
	//}

	//test mary
	//logger.Logger.Trace("=====11=======", p.testNum, p.testNum%5)
	//logger.Logger.Trace("=====22=======", p.gameState, fruits.FreeGame)
	//if p.testNum%5 == 0 && p.gameState != fruits.FreeGame {
	//	for i := 0; i < 4; i++ {
	//		p.result.EleValue[i] = fruits.Wild
	//	}
	//}
	//p.testNum++
	//if p.testNum >= 100000 {
	//	p.testNum = 0
	//}

	p.result.Win()
	jackPotNum := p.result.JackPotNum
	//JackPot 按数量给倍率 按数量给奖池
	if jackPotNum >= 3 {
		if jackPotNum == 3 {
			p.JackPotRate = 5
			p.JackPot7Rate = 100
		} else if jackPotNum == 4 {
			p.JackPotRate = 10
			p.JackPot7Rate = 200
		} else if jackPotNum == 5 {
			p.JackPotRate = 15
			p.JackPot7Rate = 1750
		}
		p.wl = append(p.wl, model.Classic777WinLine{
			Id:          -1,
			EleValue:    fruits.Scatter,
			Num:         jackPotNum,
			Rate:        p.JackPot7Rate,
			WinCoin:     p.oneBetCoin * p.JackPot7Rate,
			WinFreeGame: -1,
		})
	}
	//线倍率
	var rate int64
	for _, wl := range p.result.WinLine {
		if wl.LineId == 10 {
			continue
		}
		//Wild 元素 按线的连续给玛丽游戏
		NowWildNum := 0
		var flag = wl.Lines[0]
		for _, l := range wl.Lines {
			if l != fruits.Wild && NowWildNum > 0 {
				flag = l
				if NowWildNum < 3 {
					NowWildNum = 0
				}
			} else if l == fruits.Wild {
				NowWildNum++
			}
		}
		if NowWildNum >= 3 {
			newM := model.Classic777WinLine{
				Id:       -1,
				EleValue: fruits.Wild,
				Num:      NowWildNum,
				Rate:     0,
				WinCoin:  0,
			}
			if NowWildNum == 3 {
				p.maryFreeTimes += 1
				newM.WinFreeGame = 0
			} else if NowWildNum == 4 {
				p.maryFreeTimes += 2
				newM.WinFreeGame = 1
			} else if NowWildNum >= 5 {
				p.maryFreeTimes += 3
				newM.WinFreeGame = 2
			}
			p.wl = append(p.wl, newM)
		}

		//统计线倍率
		rate += wl.Rate

		//Bonus 元素 按线给倍率 按线给免费
		if flag == fruits.Bonus {
			n := len(wl.Lines)
			newL := model.Classic777WinLine{
				Id:       wl.LineId,
				EleValue: fruits.Bonus,
				Num:      n,
				Rate:     wl.Rate,
				WinCoin:  p.oneBetCoin * wl.Rate,
			}
			if n == 3 {
				p.freeTimes += 5
				p.winFreeTimes += 5
				newL.WinFreeGame = 3
			} else if n == 4 {
				p.freeTimes += 8
				p.winFreeTimes += 8
				newL.WinFreeGame = 4
			} else if n == 5 {
				p.freeTimes += 10
				p.winFreeTimes += 10
				newL.WinFreeGame = 5
			}
			p.wl = append(p.wl, newL)
		}

	}
	//赢的线钱
	p.winLineCoin += p.oneBetCoin * rate
	p.winLineRate = rate
	//元素输赢
	p.winEle777Coin += p.oneBetCoin * p.JackPot7Rate
}

// 玛丽游戏
func (p *FruitsPlayerData) CreateMary() {
	if len(p.preMaryMidArray) > 0 {
		p.result.MaryMidArray = make([]int32, 4)
		copy(p.result.MaryMidArray, p.preMaryMidArray)
		p.result.MaryOutSide = p.preMaryOutSide
		p.preMaryMidArray = nil
	}

	idx := p.result.MaryOutSide
	maryOutSide := fruits.MaryEleArray[idx]
	midArr := p.result.MaryMidArray
	//判断是否连续
	for _, ele := range midArr {
		if ele != maryOutSide {
			break
		}
		p.maryInNum++
	}
	if p.maryInNum == 0 {
		//不连续 判断是否含有
		for _, ele := range midArr {
			if ele == maryOutSide {
				p.maryInNum++
				break
			}
		}
	}
	var rate int64
	//至少中1个
	if p.maryInNum > 0 {
		switch maryOutSide {
		case fruits.Cherry:
			rate += 50
		case fruits.Pineapple:
			rate += 5
		case fruits.Grape:
			rate += 100
		case fruits.Lemon:
			rate += 70
		case fruits.Watermelon:
			rate += 200
		case fruits.Banana:
			rate += 20
		case fruits.Bonus:
			rate += 10
		}
		if p.maryInNum == 3 {
			rate += 20
		} else if p.maryInNum == 4 {
			rate += 500
		}
	}
	if maryOutSide == fruits.Bomb {
		p.maryFreeTimes = 0
		rate = 0
	}
	p.maryWinTotalRate = rate
	p.winMaryCoin += p.oneBetCoin * p.maryWinTotalRate
}
