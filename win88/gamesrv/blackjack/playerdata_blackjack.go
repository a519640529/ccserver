package blackjack

import (
	"games.yol.com/win88/gamerule/blackjack"
	"games.yol.com/win88/gamesrv/base"
	"games.yol.com/win88/model"
	"strconv"
)

type BlackJackHand struct {
	handCards []*blackjack.Card // 玩家手牌
	mulCards  *blackjack.Card   // 双倍手牌
	tp        int32             // 类型
	point     []int32           // 点数
	state     int32             // 已操作 0 默认 1 已停牌
	betCoin   int64             // 下注额
	betChange int64             // 输赢分(税后)
	revenue   int64             // 本局税收
	index     int               // 智能化运营取牌位置
}

type BlackJackPlayerData struct {
	*base.Player
	hands     [2]BlackJackHand
	isBet     bool  // 已下注
	isBuy     bool  // 已买保险
	baoCoin   int64 // 保险金
	baoChange int64 // 保险金输赢分（税后）
	revenue   int64 // 保险金税收

	isBanker bool // 是否庄家
	seat     int  // 座位号
	betTime  int  // 计算下注时长

	loseStatus    bool  // 分段调控亏钱状态
	winStatus     bool  // 分段调控大赢家状态
	lossStatus    bool  // 赔率调控亏钱状态
	leaveNum      int   // 机器人离场
	noOptionTimes int   // 连续自动下注次数
	opCodeFlag    int32 // 操作标记
	opRightFlag   int32 // 区分分派之后的左右牌 0默认 1右
}

func (this *BlackJackPlayerData) Init() {
	for i := 0; i < len(this.hands); i++ {
		this.hands[i].mulCards = blackjack.NewCardDefault()
	}
}

func (this *BlackJackPlayerData) Hands() []model.BlackJackCardInfo {
	var res []model.BlackJackCardInfo
	for _, v := range this.hands {
		if len(v.handCards) > 0 {
			info := model.BlackJackCardInfo{
				Cards:         blackjack.CardsToInt32(v.handCards),
				CardType:      v.tp,
				CardPoint:     v.point,
				BetCoin:       v.betCoin,
				GainCoinNoTax: v.betChange,
			}
			if v.betChange > 0 {
				info.IsWin = 1
			} else if v.betChange < 0 {
				info.IsWin = -1
			}
			res = append(res, info)
		}
	}
	return res
}

// 下注总额
func (this *BlackJackPlayerData) BetCoin() int64 {
	return this.hands[0].betCoin + this.hands[1].betCoin
}

func (this *BlackJackPlayerData) LeftBetCoin() int64 {
	return this.hands[0].betCoin
}

func (this *BlackJackPlayerData) RightBetCoin() int64 {
	return this.hands[1].betCoin
}

func (this *BlackJackPlayerData) BetChange() int64 {
	return this.hands[0].betChange + this.hands[1].betChange
}

func (this *BlackJackPlayerData) LeftBetChange() int64 {
	return this.hands[0].betChange
}

func (this *BlackJackPlayerData) RightBetChange() int64 {
	return this.hands[1].betChange
}

func (this *BlackJackPlayerData) Release() {
	this.isBet = false
	this.isBuy = false
	this.baoCoin = 0
	this.baoChange = 0
	this.revenue = 0
	this.loseStatus = false
	this.winStatus = false
	this.lossStatus = false
	this.opCodeFlag = 999
	this.opRightFlag = 0
	for k := range this.hands {
		this.hands[k].handCards = this.hands[k].handCards[:0]
		this.hands[k].mulCards = blackjack.NewCardDefault()
		this.hands[k].tp = 0
		this.hands[k].point = []int32{}
		this.hands[k].state = 0
		this.hands[k].betCoin = 0
		this.hands[k].betChange = 0
		this.hands[k].revenue = 0
		this.hands[k].index = -1
	}
	// 清除代玩座位的玩家数据
	if this.Player != nil && this.IsGameing() && this.GetPos() != this.seat {
		this.Player = nil
	}
}

func (this *BlackJackPlayerData) GetDaliyGameData(id int) (*model.PlayerGameStatics, *model.PlayerGameStatics) {
	gameId := strconv.Itoa(id)
	if this.TodayGameData == nil {
		this.TodayGameData = model.NewPlayerGameCtrlData()
	}
	if this.TodayGameData.CtrlData == nil {
		this.TodayGameData.CtrlData = make(map[string]*model.PlayerGameStatics)
	}
	if _, ok := this.TodayGameData.CtrlData[gameId]; !ok {
		this.TodayGameData.CtrlData[gameId] = &model.PlayerGameStatics{}
	}
	if this.YesterdayGameData == nil {
		this.YesterdayGameData = model.NewPlayerGameCtrlData()
	}
	if this.YesterdayGameData.CtrlData == nil {
		this.YesterdayGameData.CtrlData = make(map[string]*model.PlayerGameStatics)
	}
	if _, ok := this.YesterdayGameData.CtrlData[gameId]; !ok {
		this.YesterdayGameData.CtrlData[gameId] = &model.PlayerGameStatics{}
	}
	return this.TodayGameData.CtrlData[gameId], this.YesterdayGameData.CtrlData[gameId]
}

func (this *BlackJackPlayerData) SetNowTodayGameData(totalIn, totalOut int64, keyGameId string) {
	if this.GDatas == nil {
		this.GDatas = make(map[string]*model.PlayerGameInfo)
	}
	if this.TodayGameData.CtrlData == nil {
		this.TodayGameData.CtrlData = make(map[string]*model.PlayerGameStatics)
	}
	//总值数据累加
	if data, ok := this.GDatas[keyGameId]; ok {
		data.Statics.TotalIn += totalIn
		data.Statics.TotalOut += totalOut
		data.Statics.GameTimes++
	} else {
		gs := &model.PlayerGameInfo{}
		gs.Statics.TotalIn = totalIn
		gs.Statics.TotalOut = totalOut
		gs.Statics.GameTimes++
		this.GDatas[keyGameId] = gs
	}
	//当天数据累加
	if data, ok := this.TodayGameData.CtrlData[keyGameId]; ok {
		data.TotalIn += totalIn
		data.TotalOut += totalOut
		data.GameTimes++
	} else {
		gs := &model.PlayerGameStatics{}
		gs.TotalIn = totalIn
		gs.TotalOut = totalOut
		gs.GameTimes++
		this.TodayGameData.CtrlData[keyGameId] = gs
	}
}

func (p *BlackJackPlayerData) IsStopAction() bool {
	return p.hands[0].state == 1 && (len(p.hands[1].handCards) == 0 || p.hands[1].state == 1)
}

func NewBlackJackPlayerData(p *base.Player) *BlackJackPlayerData {
	ret := &BlackJackPlayerData{
		Player: p,
	}
	ret.Init()
	return ret
}
