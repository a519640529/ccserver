package iceage

import (
	"encoding/json"
	"math/rand"
	"strconv"
	"time"

	rule "games.yol.com/win88/gamerule/iceage"
	base "games.yol.com/win88/gamesrv/base"
	"games.yol.com/win88/model"
	"games.yol.com/win88/protocol/iceage"
	"github.com/idealeak/goserver/core/timer"
	//"games.yol.com/win88/common"
)

type IceAgePlayerData struct {
	*base.Player
	spinID           int64                 //当前旋转ID
	score            int32                 //单线押注数
	freeTimes        int32                 //免费转动次数
	cards            []*iceage.IceAgeCards //15张牌
	totalPriceBonus  int64                 //小游戏奖金
	bonusTimerHandle timer.TimerHandle     //托管handle
	//RollGameType     *model.IceAgeType     //记录信息
	RollGameType     *GameResultLog         //记录信息
	enterGameCoin    int64                  //玩家进入初始金币
	betLines         []int64                //下注的选线
	taxCoin          int64                  //本局税收
	winCoin          int64                  //本局收税前赢的钱
	linesWinCoin     int64                  //本局中奖线赢得钱
	jackpotWinCoin   int64                  //本局奖池赢的钱
	smallGameWinCoin int64                  //本局小游戏赢的钱
	currentLogId     string                 //爆奖玩家logid
	leavetime        int32                  //用户离开时间
	billedData       *iceage.GameBilledData //上一局结算信息
	BonusLineIdx     int64                  //bonus中奖线索引
	DebugBonus       bool                   //测试进入bonus
	TestNum          int
}

//玩家初始化
func (this *IceAgePlayerData) init(s *base.Scene) {
	this.Clean()
	this.score = 0
	this.freeTimes = 0
	//this.RollGameType = &model.IceAgeType{}
	this.RollGameType = &GameResultLog{}
	this.RollGameType.BaseResult = &model.SlotBaseResultType{}
	this.enterGameCoin = this.GetCoin()
	this.currentLogId = ""
	this.DebugBonus = true
	this.TestNum = 0
	this.billedData = &iceage.GameBilledData{}

	initCards := getSlotsDataByGroupName(DefaultData_v1)
	cards := make([]int32, 0)
	for _, card := range initCards {
		cards = append(cards, int32(card))
	}
	this.cards = []*iceage.IceAgeCards{
		{
			Card: cards,
		},
	}

	// 加载玩家游戏数据
	if this.GDatas == nil {
		this.GDatas = make(map[string]*model.PlayerGameInfo)
	}
	gameFreeID := strconv.Itoa(int(s.GetGameFreeId()))
	if d, exist := this.GDatas[gameFreeID]; exist {
		gLen := len(d.Data)
		if gLen < IAIndexMax {
			for i := gLen; i < IAIndexMax; i++ {
				d.Data = append(d.Data, 0)
			}
		}
	} else {
		pgd := &model.PlayerGameInfo{
			Data: make([]int64, IAIndexMax, IAIndexMax),
		}
		this.GDatas[gameFreeID] = pgd
	}
	this.LoadPlayerGameData(gameFreeID)

	if len(this.betLines) == 0 {
		this.betLines = make([]int64, rule.LINENUM)
		for i := range this.betLines {
			this.betLines[i] = int64(i + 1)
		}
	}
}

//玩家清理数据
func (this *IceAgePlayerData) Clean() {
	for i := 0; i < len(this.cards); i++ {
		this.cards[i] = nil
	}
	this.winCoin = 0
	this.taxCoin = 0
	this.linesWinCoin = 0
	this.jackpotWinCoin = 0
	this.smallGameWinCoin = 0
}

//加载玩家游戏数据
func (this *IceAgePlayerData) LoadPlayerGameData(gameFreeId string) {
	if d, exist := this.GDatas[gameFreeId]; exist {
		this.freeTimes = int32(d.Data[IAFreeTimes])
		if this.freeTimes > 0 && len(d.DataEx) != 0 {
			json.Unmarshal(d.DataEx, &this.betLines)
		}
	}
}

//存储玩家游戏数据
func (this *IceAgePlayerData) SavePlayerGameData(gameFreeId string) {
	if d, exist := this.GDatas[gameFreeId]; exist {
		d.Data[IAFreeTimes] = int64(this.freeTimes)
		d.DataEx, _ = json.Marshal(this.betLines)
	}
}

//黑白名单的限制是否生效
func (this *IceAgePlayerData) CheckBlackWriteList(isWin bool) bool {
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
