package easterisland

import (
	"encoding/json"
	"math/rand"
	"time"

	rule "games.yol.com/win88/gamerule/easterisland"
	"games.yol.com/win88/gamesrv/base"
	"games.yol.com/win88/model"
	"games.yol.com/win88/protocol/easterisland"
	"github.com/idealeak/goserver/core/timer"
)

type EasterIslandPlayerData struct {
	*base.Player
	score     int32   //单线押注数
	spinID    int64   //当前旋转ID
	freeTimes int32   //免费转动次数
	cards     []int32 //15张牌
	//RollGameType     *model.EasterIslandType                 //记录信息
	RollGameType     *GameResultLog                          //记录信息
	enterGameCoin    int64                                   //玩家进入老虎机初始金币
	taxCoin          int64                                   //本局税收
	winCoin          int64                                   //本局收税前赢的钱
	linesWinCoin     int64                                   //本局中奖线赢得钱
	jackpotWinCoin   int64                                   //本局奖池赢的钱
	smallGameWinCoin int64                                   //本局小游戏赢的钱
	betLines         []int64                                 //下注的选线
	totalPriceBonus  int64                                   //小游戏得分
	bonusTimerHandle timer.TimerHandle                       //托管handle
	currentLogId     string                                  //爆奖玩家logid
	leavetime        int32                                   //用户离开时间
	bonusStage       int32                                   //小游戏所处阶段 1选金额 2选倍率
	bonusStartTime   int64                                   //小游戏阶段开始时间
	bonusOpRecord    []int32                                 //小游戏操作记录
	bonusGame        *easterisland.EasterIslandBonusGameInfo //小游戏
	bonusX           []int32                                 //小游戏倍率选项
	billedData       *easterisland.GameBilledData
	DebugGame        bool //测试
	TestNum          int
}

//玩家初始化
func (this *EasterIslandPlayerData) init(s *base.Scene) {
	this.Clean()
	this.score = 0
	this.freeTimes = 0
	//this.RollGameType = &model.EasterIslandType{}
	this.RollGameType = &GameResultLog{}
	this.RollGameType.BaseResult = &model.SlotBaseResultType{}
	this.enterGameCoin = this.Coin
	this.currentLogId = ""
	this.billedData = &easterisland.GameBilledData{}
	this.DebugGame = true
	this.TestNum = 0

	// 加载玩家游戏数据
	if this.GDatas == nil {
		this.GDatas = make(map[string]*model.PlayerGameInfo)
	}

	if d, exist := this.GDatas[s.KeyGamefreeId]; exist {
		gLen := len(d.Data)
		if gLen < EasterIslandIndexMax {
			for i := gLen; i < EasterIslandIndexMax; i++ {
				d.Data = append(d.Data, 0)
			}
		}
	} else {
		pgd := &model.PlayerGameInfo{
			Data: make([]int64, EasterIslandIndexMax, EasterIslandIndexMax),
		}
		this.GDatas[s.KeyGamefreeId] = pgd
	}
	this.LoadPlayerGameData(s.KeyGamefreeId)
	//线条全选
	if len(this.betLines) == 0 {
		this.betLines = rule.AllBetLines
	}
}

//玩家清理数据
func (this *EasterIslandPlayerData) Clean() {
	for i := 0; i < len(this.cards); i++ {
		this.cards[i] = -1
	}
	this.winCoin = 0
	this.taxCoin = 0
	this.linesWinCoin = 0
	this.jackpotWinCoin = 0
	this.smallGameWinCoin = 0
	this.CleanBonus()
}

//清理小游戏数据
func (this *EasterIslandPlayerData) CleanBonus() {
	this.totalPriceBonus = 0
	this.bonusStage = 0
	this.bonusTimerHandle = timer.TimerHandle(0)
	this.bonusStartTime = 0
	this.bonusOpRecord = make([]int32, 0)
	this.bonusGame = nil
	this.bonusX = nil
}

//加载玩家游戏数据
func (this *EasterIslandPlayerData) LoadPlayerGameData(gameFreeId string) {
	if d, exist := this.GDatas[gameFreeId]; exist {
		this.freeTimes = int32(d.Data[EasterIslandFreeTimes])
		if this.freeTimes > 0 && len(d.DataEx) != 0 {
			json.Unmarshal(d.DataEx, &this.betLines)
		}
	}
}

//存储玩家游戏数据
func (this *EasterIslandPlayerData) SavePlayerGameData(gameFreeId string) {
	if d, exist := this.GDatas[gameFreeId]; exist {
		d.Data[EasterIslandFreeTimes] = int64(this.freeTimes)
		d.DataEx, _ = json.Marshal(this.betLines)
	}
}

//黑白名单的限制是否生效
func (this *EasterIslandPlayerData) CheckBlackWriteList(isWin bool) bool {
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
