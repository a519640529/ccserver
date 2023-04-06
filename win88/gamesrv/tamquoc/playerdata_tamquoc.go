package tamquoc

import (
	"encoding/json"
	"math/rand"
	"time"

	rule "games.yol.com/win88/gamerule/tamquoc"
	"games.yol.com/win88/gamesrv/base"
	"games.yol.com/win88/model"
	"games.yol.com/win88/protocol/tamquoc"
	"github.com/idealeak/goserver/core/timer"
)

type TamQuocPlayerData struct {
	*base.Player
	spinID           int64             //当前旋转ID
	score            int32             //单线押注数
	freeTimes        int32             //免费转动次数
	cards            []int32           //15张牌
	totalPriceBonus  int64             //小游戏奖金
	bonusTimerHandle timer.TimerHandle //托管handle
	//RollGameType        *model.TamQuocType //记录信息
	RollGameType        *GameResultLog //记录信息
	enterGameCoin       int64          //玩家进入初始金币
	betLines            []int64        //下注的选线
	taxCoin             int64          //本局税收
	winCoin             int64          //本局收税前赢的钱
	linesWinCoin        int64          //本局中奖线赢得钱
	jackpotWinCoin      int64          //本局奖池赢的钱
	smallGameWinCoin    int64          //本局小游戏赢的钱
	currentLogId        string         //爆奖玩家logid
	leavetime           int32          //用户离开时间
	bonusGameTime       int64          //上一次小游戏的时间
	bonusGameStartTime  time.Time      // 小游戏开始时间
	bonusGamePickPos    []int32        // 小游戏位置
	bonusGameCanPickNum int            // 小游戏点击次数
	bonusGame           tamquoc.TamQuocBonusGameInfo
	billedData          *tamquoc.GameBilledData
	DebugGame           bool //测试
	TestNum             int
}

//玩家初始化
func (this *TamQuocPlayerData) init(s *base.Scene) {
	this.Clean()
	this.score = 0
	this.freeTimes = 0
	//this.RollGameType = &model.TamQuocType{}
	this.RollGameType = &GameResultLog{}
	this.RollGameType.BaseResult = &model.SlotBaseResultType{}
	this.enterGameCoin = this.Coin
	this.currentLogId = ""
	this.billedData = &tamquoc.GameBilledData{}
	this.DebugGame = true
	this.TestNum = 0

	initCards := rule.GenerateSlotsData_v2(rule.SYMBOL2)
	for _, card := range initCards {
		this.cards = append(this.cards, int32(card))
	}

	// 加载玩家游戏数据
	if this.GDatas == nil {
		this.GDatas = make(map[string]*model.PlayerGameInfo)
	}

	if d, exist := this.GDatas[s.KeyGamefreeId]; exist {
		gLen := len(d.Data)
		if gLen < TamQuocIndexMax {
			for i := gLen; i < TamQuocIndexMax; i++ {
				d.Data = append(d.Data, 0)
			}
		}
	} else {
		pgd := &model.PlayerGameInfo{
			Data: make([]int64, TamQuocIndexMax, TamQuocIndexMax),
		}
		this.GDatas[s.KeyGamefreeId] = pgd
	}
	this.LoadPlayerGameData(s.KeyGamefreeId)
	//线条全选
	if len(this.betLines) == 0 {
		this.betLines = rule.AllBetLines
	}
	this.bonusGamePickPos = make([]int32, 2)
}

//玩家清理数据
func (this *TamQuocPlayerData) Clean() {
	for i := 0; i < len(this.cards); i++ {
		this.cards[i] = -1
	}
	this.winCoin = 0
	this.taxCoin = 0
	this.linesWinCoin = 0
	this.jackpotWinCoin = 0
	this.smallGameWinCoin = 0
}

//加载玩家游戏数据
func (this *TamQuocPlayerData) LoadPlayerGameData(gameFreeId string) {
	if d, exist := this.GDatas[gameFreeId]; exist {
		this.freeTimes = int32(d.Data[TamQuocFreeTimes])
		this.bonusGameTime = d.Data[TamQuocBonusTime]
		if this.freeTimes > 0 && len(d.DataEx) != 0 {
			json.Unmarshal(d.DataEx, &this.betLines)
		}
	}
}

//存储玩家游戏数据
func (this *TamQuocPlayerData) SavePlayerGameData(gameFreeId string) {
	if d, exist := this.GDatas[gameFreeId]; exist {
		d.Data[TamQuocFreeTimes] = int64(this.freeTimes)
		d.Data[TamQuocBonusTime] = this.bonusGameTime
		d.DataEx, _ = json.Marshal(this.betLines)
	}
}

//黑白名单的限制是否生效
func (this *TamQuocPlayerData) CheckBlackWriteList(isWin bool) bool {
	if isWin && this.BlackLevel > 0 && this.BlackLevel <= 10 {
		rand.Seed(time.Now().UnixNano())
		if rand.Int31n(100) < this.BlackLevel*10 {
			return true
		}
	} else if !isWin && this.WhiteLevel > 0 && this.WhiteLevel <= 10 {
		rand.Seed(time.Now().UnixNano())
		if rand.Int31n(100) < this.WhiteLevel*10 {
			return true
		}
	}
	return false
}
