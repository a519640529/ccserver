package fortunezhishen

import (
	gamerule "games.yol.com/win88/gamerule/fortunezhishen"
	"games.yol.com/win88/gamesrv/base"
	"games.yol.com/win88/model"
	"github.com/idealeak/goserver/core/logger"
	"math/rand"
)

type FortuneZhiShenPlayerData struct {
	*base.Player
	roomid              int32 //房间ID
	result              *gamerule.WinResult
	firstFreeTimes      int                           //免费转动次数
	secondFreeTimes     int                           //免费转动次数
	grandPrizeNum       int64                         //中巨奖数量
	bigPrizeNum         int64                         //中大奖数量
	midPrizeNum         int64                         //中中数量
	smallPrizeNum       int64                         //中小奖数量
	rewardGrandPrizeNum int64                         //额外奖励
	betIdx              int                           //当前下注索引
	winTotalRate        int64                         //当前赢的倍率
	gameState           int                           //子状态
	gemstoneRateCoin    []int64                       //宝石数额
	makeAFortuneRate    int64                         //发财倍率
	winCoin             int64                         //赢的钱
	firstWinCoin        int64                         //第一个小游戏赢的钱
	secondWinCoin       int64                         //第二个小游戏赢的钱
	taxCoin             int64                         //税收
	betCoin             int64                         //玩家总下注
	oneBetCoin          int64                         //单注
	leaveTime           int32                         //用户离开时间
	nowGetCoin          int64                         //本局拿到的金额(税后)
	nowFirstTimes       int                           //当前第几轮
	nowSecondTimes      int                           //当前第几轮
	hitPrizePool        []int64                       //奖池
	gemstoneWinCoin     int64                         //宝石派彩
	wl                  []model.FortuneZhiShenWinLine //奖励免费次数记录
	cpCtx               model.CoinPoolCtx             //水池环境
	noWinTimes          int                           //没有中奖次数
	preEleVal           []int32                       //预生成的结果
	preWinRate          int64                         //本局预输赢
	pre70WinRate        int64                         //剩余70%没有给
	normalCards         []int32                       //最后一次正常牌
	newAddGemstone      int                           //新增宝石数量
	preNeedRate         int64
}

func (p *FortuneZhiShenPlayerData) init() {
	p.roomid = 0
	p.result = &gamerule.WinResult{}
}
func (p *FortuneZhiShenPlayerData) CreateResult(eleLineAppearRate []int32) {
	p.result.CreateLine(p.gameState, eleLineAppearRate)
	if len(p.preEleVal) > 0 {
		copy(p.result.Cards, p.preEleVal)
		mid := []int32{2, 3, 6, 7, 8, 11, 12, 13}
		var same = p.preEleVal[1]
		p.result.MidIcon = same
		for _, v := range mid {
			if same != p.preEleVal[v] {
				p.result.MidIcon = -1
			}
		}
		p.preEleVal = nil
	}
	if p.newAddGemstone > 0 {
		p.result.NewAddGemstone = p.newAddGemstone
		p.newAddGemstone = 0
	}
}
func (p *FortuneZhiShenPlayerData) Clear() {
	p.rewardGrandPrizeNum = 0
	p.grandPrizeNum = 0
	p.bigPrizeNum = 0
	p.midPrizeNum = 0
	p.smallPrizeNum = 0
	p.winTotalRate = 0
	p.gemstoneRateCoin = make([]int64, 15)
	p.gameState = gamerule.Normal
	p.winCoin = 0
	//p.SetStartCoin(p.GetCoin())
	p.hitPrizePool = make([]int64, 4)
	p.makeAFortuneRate = 0
	p.gemstoneWinCoin = 0
	p.taxCoin = 0
	//if p.firstFreeTimes == 0 {
	//	p.firstWinCoin = 0
	//	p.nowFirstTimes = 0
	//}
	//if p.secondFreeTimes == 0 {
	//	p.secondWinCoin = 0
	//	p.nowSecondTimes = 0
	//}
	if !p.IsHaveFreeTimes() {
		p.firstWinCoin = 0
		p.nowFirstTimes = 0

		p.secondWinCoin = 0
		p.nowSecondTimes = 0

		p.result.Init(gamerule.Normal)
	}
	p.wl = nil
}
func (p *FortuneZhiShenPlayerData) IsHaveFreeTimes() bool {
	return !(p.firstFreeTimes == 0 && p.secondFreeTimes == 0)
}
func (p *FortuneZhiShenPlayerData) Win() {
	result := p.result
	for k, rgs := range result.GemstoneRate {
		if p.result.MidIcon != -1 {
			var mid = []int{2, 3, 6, 7, 8, 11, 12, 13}
			if gamerule.FindInxInArray(k, mid) {
				continue
			}
		}
		switch rgs {
		case gamerule.GrandPrize:
			p.grandPrizeNum++
		case gamerule.BigPrize:
			p.bigPrizeNum++
		case gamerule.MidPrize:
			p.midPrizeNum++
		case gamerule.SmallPrize:
			p.smallPrizeNum++
		}
	}

	rs := p.result

	for k, v := range rs.GemstoneRate {
		if v <= 20 && v >= 1 {
			p.gemstoneRateCoin[k] = p.betCoin * v
		} else if v > 0 {
			p.gemstoneRateCoin[k] = v * 1000000
		}
	}

	var rate int64
	//单线元素倍率计算
	for _, l := range rs.WinLine {
		//r := p.GetRate(l.Lines[0], len(l.Lines))
		//rs.WinLine[k].Rate = r
		rate += l.Rate
	}
	p.winTotalRate = rate
	if p.gameState == gamerule.Normal {
		if rs.GemstoneNum >= 6 {
			//普通游戏
			//触发 旋转特别奖
			p.secondFreeTimes = 3
			p.wl = append(p.wl, model.FortuneZhiShenWinLine{
				Id:          -1,
				EleValue:    gamerule.Gemstone,
				Num:         rs.GemstoneNum,
				Rate:        -1,
				WinCoin:     p.gemstoneWinCoin,
				WinFreeGame: 2,
			})
		}
		if rs.MakeAFortuneNum >= 3 {
			//触发免费游戏
			p.firstFreeTimes = 6
			p.makeAFortuneRate = gamerule.GetRate(gamerule.MakeAFortune, rs.MakeAFortuneNum)
			p.wl = append(p.wl, model.FortuneZhiShenWinLine{
				Id:          -1,
				EleValue:    gamerule.MakeAFortune,
				Num:         rs.MakeAFortuneNum,
				Rate:        p.makeAFortuneRate,
				WinCoin:     p.makeAFortuneRate * p.oneBetCoin,
				WinFreeGame: 1,
			})
		}
	} else if p.gameState == gamerule.FreeGame {
		if rs.MidIcon == gamerule.MakeAFortune {
			p.firstFreeTimes += 3
			p.makeAFortuneRate = gamerule.GetRate(gamerule.MakeAFortune, rs.MakeAFortuneNum)
			p.wl = append(p.wl, model.FortuneZhiShenWinLine{
				Id:          -1,
				EleValue:    gamerule.MakeAFortune,
				Num:         rs.MakeAFortuneNum,
				Rate:        p.makeAFortuneRate,
				WinCoin:     p.makeAFortuneRate * p.oneBetCoin,
				WinFreeGame: 2,
			})
		} else if rs.GemstoneNum >= 6 {
			//大宝石或者 宝石数量大于6
			//触发 旋转特别奖
			p.secondFreeTimes = 3
			p.wl = append(p.wl, model.FortuneZhiShenWinLine{
				Id:          -1,
				EleValue:    gamerule.Gemstone,
				Num:         rs.GemstoneNum,
				Rate:        -1,
				WinCoin:     p.gemstoneWinCoin,
				WinFreeGame: 0,
			})
		}
	} else if p.gameState == gamerule.StopAndRotate || p.gameState == gamerule.StopAndRotate2 {
		if rs.NewAddGemstone > 0 {
			p.secondFreeTimes = 3
			p.wl = append(p.wl, model.FortuneZhiShenWinLine{
				Id:          -1,
				EleValue:    gamerule.Gemstone,
				Num:         rs.GemstoneNum,
				Rate:        -1,
				WinCoin:     p.gemstoneWinCoin,
				WinFreeGame: 0,
			})
		}
	}

	if p.result.GemstoneNum == 15 {
		p.rewardGrandPrizeNum++
	}
	//发财元素倍率
	p.winTotalRate += p.makeAFortuneRate
	if p.secondFreeTimes > 0 || (p.gameState == gamerule.StopAndRotate ||
		p.gameState == gamerule.StopAndRotate2) {
		p.result.LastRes = rs.Cards
	}
}
func (p *FortuneZhiShenPlayerData) ShowGameState(needRate int64, eleLineAppearRate []int32) {
	var isCanFree bool
	var isCanStop bool
	//免费触发免费或者旋转
	var isCanFree2 bool
	var isCanStop2 bool
	var makeAFortuneNum int
	if p.gameState == gamerule.Normal {
		if needRate >= 400 && needRate < 800 {
			//触发免费
			isCanFree = true
			needRate = rand.Int63n(100)
		} else if needRate >= 800 {
			if needRate <= 1000 {
				makeAFortuneNum = 3
			} else {
				if needRate > 1000 && needRate < 2000 {
					makeAFortuneNum = 4
				} else {
					makeAFortuneNum = 5
				}
				if rand.Intn(100) > 50 {
					//触发免费
					isCanFree = true
					needRate = rand.Int63n(100)
				} else {
					isCanStop = true
					needRate = 0
				}
			}
		}
	} else if p.gameState == gamerule.FreeGame {
		if p.nowFirstTimes == 6 {
			if p.preNeedRate > 3000 && p.preWinRate >= p.pre70WinRate {
				if rand.Intn(100) > 50 {
					isCanFree2 = true
				} else {
					isCanStop2 = true
				}
				needRate = 0
			}
		} else if p.nowFirstTimes > 6 && p.firstFreeTimes == 0 {
			needRate = p.preWinRate
		}
	}
	if needRate < 0 {
		needRate = 0
	}
	logger.Logger.Trace("LineEleVal.................needRate: ", needRate)
	wl, noPoss, preInt := gamerule.GetLineEleVal(p.betIdx, p.gameState, needRate, eleLineAppearRate)
	if noPoss == nil && len(preInt) > 0 {
		wl.Init(p.gameState)
		wl.Cards = preInt[rand.Intn(len(preInt))]
		wl.Win(p.betIdx, p.gameState)
		var poss []int32
		for _, v := range wl.WinLine {
			poss = append(poss, v.Poss...)
		}
		for k := range wl.Cards {
			isBreak := false
			for _, n := range poss {
				if k == int(n) {
					isBreak = true
					break
				}
			}
			if !isBreak {
				noPoss = append(noPoss, k)
			}
		}
	}
	p.preEleVal = wl.Cards
	for _, v := range wl.WinLine {
		p.preWinRate -= v.Rate
	}
	if isCanFree {
		if len(noPoss) > 0 && wl.MakeAFortuneNum < makeAFortuneNum {
			n := 0
			var fortuneCol = make([]int, 5)
			for k, v := range p.preEleVal {
				if v == gamerule.MakeAFortune {
					fortuneCol[k%5] = 1
				}
			}
			for i := 0; i < makeAFortuneNum-wl.MakeAFortuneNum; i++ {
				mpn := len(noPoss)
				if mpn > 0 {
					for m := 0; m < mpn; m++ {
						for k, v := range noPoss {
							if fortuneCol[v%5] == 1 {
								noPoss = append(noPoss[:k], noPoss[k+1:]...)
								break
							}
						}
					}
				} else {
					break
				}
				if len(noPoss) == 0 {
					break
				}
				key := rand.Intn(len(noPoss))
				pos := noPoss[key]
				p.preEleVal[pos] = gamerule.MakeAFortune
				fortuneCol[pos%5] = 1
				n++
			}
			if n == 3 {
				p.preWinRate -= 100
			} else if n == 4 {
				p.preWinRate -= 750
			} else if n == 5 {
				p.preWinRate -= 5000
			}
		}
	} else if isCanStop {
		if len(noPoss) > 0 && wl.GemstoneNum < 6 {
			for i := 0; i < 6-wl.GemstoneNum; i++ {
				key := rand.Intn(len(noPoss))
				pos := noPoss[key]
				p.preEleVal[pos] = gamerule.Gemstone
				noPoss = append(noPoss[:key], noPoss[key+1:]...)
			}
		}
	} else if isCanFree2 {
		for k := range p.preEleVal {
			if k == 1 || k == 2 || k == 3 ||
				k == 6 || k == 7 || k == 8 ||
				k == 11 || k == 12 || k == 13 {
				p.preEleVal[k] = gamerule.MakeAFortune
			}
		}
	} else if isCanStop2 {
		for k := range p.preEleVal {
			if k == 1 || k == 2 || k == 3 ||
				k == 6 || k == 7 || k == 8 ||
				k == 11 || k == 12 || k == 13 {
				p.preEleVal[k] = gamerule.Gemstone
			}
		}
	}
	return
}
