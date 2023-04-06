package caishen

import (
	"encoding/json"
	"math/rand"
	"time"

	rule "games.yol.com/win88/gamerule/caishen"
	"games.yol.com/win88/gamesrv/base"
	"games.yol.com/win88/model"
	"games.yol.com/win88/protocol/caishen"
	"github.com/idealeak/goserver/core/timer"
)

type CaiShenPlayerData struct {
	*base.Player
	spinID    int64   //当前旋转ID
	score     int32   //单线押注数
	freeTimes int32   //免费转动次数
	cards     []int32 //15张牌
	//RollGameType      *model.CaiShenType            //记录信息
	RollGameType      *GameResultLog                //记录信息
	enterGameCoin     int64                         //玩家进入初始金币
	taxCoin           int64                         //本局税收
	winCoin           int64                         //本局收税前赢的钱
	linesWinCoin      int64                         //本局中奖线赢得钱
	jackpotWinCoin    int64                         //本局奖池赢的钱
	smallGameWinCoin  int64                         //本局小游戏赢的钱 //新版本已废弃
	betLines          []int64                       //下注的选线
	totalPriceBonus   int64                         //小游戏得分
	bonusTimerHandle  timer.TimerHandle             //托管handle
	currentLogId      string                        //爆奖玩家logid
	leavetime         int32                         //用户离开时间
	bonusStage        int32                         //小游戏所处阶段 1选金额 2选倍率
	bonusStartTime    int64                         //小游戏阶段开始时间
	bonusOpRecord     []int32                       //小游戏操作记录
	bonusGame         *caishen.CaiShenBonusGameInfo //小游戏
	bonusX            []int32                       //小游戏倍率选项
	bonusGameIconData []int32                       //小游戏选项图标，用于客户端展示
	billedData        *caishen.GameBilledData
	DebugGame         bool //测试
	TestNum           int
}

//玩家初始化
func (this *CaiShenPlayerData) init(s *base.Scene) {
	this.Clean()
	this.score = 0
	this.freeTimes = 0
	//this.RollGameType = &model.CaiShenType{}
	this.RollGameType = &GameResultLog{}
	this.RollGameType.BaseResult = &model.SlotBaseResultType{}
	this.enterGameCoin = this.Coin
	this.currentLogId = ""
	this.billedData = &caishen.GameBilledData{}
	this.DebugGame = true
	this.TestNum = 0

	// 加载玩家游戏数据
	if this.GDatas == nil {
		this.GDatas = make(map[string]*model.PlayerGameInfo)
	}
	if d, exist := this.GDatas[s.KeyGamefreeId]; exist {
		gLen := len(d.Data)
		if gLen < CaiShenIndexMax {
			for i := gLen; i < CaiShenIndexMax; i++ {
				d.Data = append(d.Data, 0)
			}
		}
	} else {
		pgd := &model.PlayerGameInfo{
			Data: make([]int64, CaiShenIndexMax, CaiShenIndexMax),
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
func (this *CaiShenPlayerData) Clean() {
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
func (this *CaiShenPlayerData) CleanBonus() {
	this.totalPriceBonus = 0
	this.bonusStage = 0
	this.bonusTimerHandle = timer.TimerHandle(0)
	this.bonusStartTime = 0
	this.bonusOpRecord = make([]int32, 0)
	this.bonusGame = nil
	this.bonusX = nil
	this.bonusGameIconData = make([]int32, 0)
}

//加载玩家游戏数据
func (this *CaiShenPlayerData) LoadPlayerGameData(gameFreeId string) {
	if d, exist := this.GDatas[gameFreeId]; exist {
		this.freeTimes = int32(d.Data[CaiShenFreeTimes])
		if this.freeTimes > 0 && len(d.DataEx) != 0 {
			json.Unmarshal(d.DataEx, &this.betLines)
		}
	}
}

//存储玩家游戏数据
func (this *CaiShenPlayerData) SavePlayerGameData(gameFreeId string) {
	if d, exist := this.GDatas[gameFreeId]; exist {
		d.Data[CaiShenFreeTimes] = int64(this.freeTimes)
		d.DataEx, _ = json.Marshal(this.betLines)
	}
}

//黑白名单的限制是否生效
func (this *CaiShenPlayerData) CheckBlackWriteList(isWin bool) bool {
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
