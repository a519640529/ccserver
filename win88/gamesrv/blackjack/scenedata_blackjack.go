package blackjack

import (
	"bytes"
	"games.yol.com/win88/common"
	rule "games.yol.com/win88/gamerule/blackjack"
	"games.yol.com/win88/gamesrv/base"
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/blackjack"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/timer"

	"math"
	"math/rand"
	"sort"
	"strconv"
	"time"
)

type BlackJackSceneData struct {
	*base.Scene
	players  map[int32]*BlackJackPlayerData           // 所有玩家信息
	chairIDs [rule.MaxPlayer + 1]*BlackJackPlayerData // 座位信息
	pokers   []*rule.Card
	pos      int // 当前操作的玩家位置
	lastTime time.Time
	buy      bool // 保险调控
	robotNum int  // 机器人数量
	gameNum  int  // 游戏中的玩家数量

	n             int
	maxHand       []*rule.Card
	backup        [rule.MaxPlayer]*BlackJackPlayerData
	logicId       string
	playerData    []*model.BlackJackPlayer
	gamingNum     int
	endType       int
	timeOutDeal   time.Duration
	hTimerOpDelay timer.TimerHandle //上一个玩家停牌后，下一个玩家延迟操作
	winResult     int               //最终结果
	lastOperaPos  int               //上一个玩家操作位置
}

func NewBlackJackSceneData(s *base.Scene) *BlackJackSceneData {
	return &BlackJackSceneData{
		Scene:   s,
		players: make(map[int32]*BlackJackPlayerData),
	}
}

func (this *BlackJackSceneData) Init() bool {
	banker := NewBlackJackPlayerData(nil)
	banker.isBanker = true
	banker.seat = 0
	this.chairIDs[0] = banker

	for k := range this.chairIDs[1:] {
		player := NewBlackJackPlayerData(nil)
		player.seat = k + 1
		this.chairIDs[k+1] = player
	}

	this.GameStartTime = time.Now()
	this.SetPlayerNum(rule.MaxPlayer)
	this.gameNum = 0
	this.logicId, _ = model.AutoIncGameLogId()
	this.endType = 0 // 0 一个玩家结算一次 1 玩家有几手牌结算几次
	this.timeOutDeal = time.Second * 1
	return true
}

func (this *BlackJackSceneData) Release() {
	this.robotNum = 0
	this.gameNum = 0
	this.logicId, _ = model.AutoIncGameLogId()
	this.playerData = this.playerData[:0]
	this.gamingNum = this.GetGameingPlayerCnt()
	this.SetCpControlled(false)
	this.SetSystemCoinOut(0)
}

// InitRobotNumGameNum 机器人玩家数量统计
func (this *BlackJackSceneData) InitRobotNumGameNum() {
	for _, v := range this.chairIDs[1:] {
		if v.Player != nil && v.IsGameing() {
			this.gameNum++
			if v.IsRobot() {
				this.robotNum++
			}
		}
	}
}

func (this *BlackJackSceneData) OnPlayerLeave(p *base.Player) {
	if p == nil {
		return
	}
	if v, ok := this.players[p.SnId]; ok {
		delete(this.players, v.SnId)
	} else {
		return
	}
	pos := []int32{}
	for k, v := range this.chairIDs[1:] {
		if v.Player != nil && v.SnId == p.SnId {
			if p.GetPos() == v.seat {
				pos = append([]int32{int32(v.seat)}, pos...)
			} else {
				pos = append(pos, int32(v.seat))
			}
			// 备份游戏记录
			if v.isBet && v.Billed {
				player := NewBlackJackPlayerData(nil)
				player.seat = v.seat
				this.chairIDs[k+1] = player
				this.backup[k] = v
			}
			v.Player = nil
		}
	}
	pack := &blackjack.SCBlackJackPlayerLeave{
		Pos: pos,
	}
	proto.SetDefaults(pack)
	this.Broadcast(int(blackjack.BlackJackPacketID_SC_PLAYER_LEAVE), pack, p.GetSid())
	logger.Logger.Tracef("--> Broadcast PlayerLeave snid=%d pos=%d %s", p.SnId, p.GetPos(), pack.String())
}

func (this *BlackJackSceneData) SceneDestroy(b bool) {
	this.Destroy(b)
}

func (this *BlackJackSceneData) JudgeStart() bool {
	// 都是机器人不开始
	//if this.GetRealPlayerCnt() == 0 {
	//	return false
	//}
	n := 0
	for _, v := range this.chairIDs[1:] {
		if v.Player != nil && v.IsReady() {
			n++
		}
	}
	// 有一个人准备就开始游戏
	return n > 0
}

func (this *BlackJackSceneData) AutoBet(coin int64) {
	for _, v := range this.chairIDs {
		if v.Player != nil && v.IsGameing() && !v.isBet {
			this.Bet(v.seat, coin)
			v.noOptionTimes++
		}
	}
}

func (this *BlackJackSceneData) JudgeBetEnd() bool {
	for _, v := range this.chairIDs {
		if v.Player != nil && v.IsGameing() && !v.isBet {
			return false
		}
	}
	return true
}

func (this *BlackJackSceneData) PlayerCoin(p *BlackJackPlayerData) int64 {
	if p == nil {
		return 0
	}
	switch this.SceneState.GetState() {
	case rule.StatusBuyEnd, rule.StatusPlayer, rule.StatusBanker:
		if p.isBuy {
			if p.baoChange > 0 {
				return p.GetCoin() + p.baoChange - p.BetCoin()
			} else if p.baoChange < 0 {
				return p.GetCoin() - p.baoCoin - p.BetCoin()
			}
		}
	}
	return p.GetCoin() - p.BetCoin() - p.baoCoin
}

func (this *BlackJackSceneData) Revenue(p *BlackJackPlayerData) int64 {
	return p.revenue + p.hands[0].revenue + p.hands[1].revenue
}

// 下注
func (this *BlackJackSceneData) Bet(seat int, coin int64) *blackjack.SCBlackJackPlayerBet {
	player := BlackJackGetPlayerData(this.Scene, seat)
	if player.Player == nil || !player.IsGameing() || player.isBet {
		return nil
	}
	player.isBet = true
	player.hands[0].betCoin = coin
	// 广播下注结果
	pack := &blackjack.SCBlackJackPlayerBet{
		Code:   blackjack.SCBlackJackPlayerBet_Success,
		Pos1:   proto.Int32(int32(player.GetPos())),
		Pos2:   proto.Int32(int32(seat)),
		Coin:   proto.Int64(coin),
		ReCoin: proto.Int64(this.PlayerCoin(player)),
	}
	this.Broadcast(int(blackjack.BlackJackPacketID_SC_PLAYER_BET), pack, 0)
	logger.Logger.Trace("--> Bet SCBlackJackPlayerBet_BetCoin ", pack.String())
	// 记录下注花费时长
	if player.GetPos() == player.seat {
		player.betTime = this.SceneState.GetTimeout(this.Scene)
	}
	return pack
}

// 确定谁坐一号座位
func (this *BlackJackSceneData) SortSeat() {
	var players []*BlackJackPlayerData
	for _, v := range this.chairIDs {
		if v.Player != nil && v.IsGameing() && v.GetPos() == v.seat && v.isBet {
			players = append(players, v)
		}
	}
	if len(players) < 1 {
		return
	}
	// 找出1号位玩家
	seat := -1
	for {
		if len(players) == 1 {
			seat = players[0].seat
			break
		}
		// 时间
		t := int(rule.TimeoutBet/time.Second) + 10
		var arr = make([]int, 0)
		for _, v := range players {
			if v.betTime > t {
				continue
			}
			if v.betTime < t {
				arr = arr[:0]
				t = v.betTime
			}
			arr = append(arr, v.seat)
		}
		if len(arr) < 1 {
			break
		}
		if len(arr) == 1 {
			seat = arr[0]
			break
		}
		// 下注额
		players = players[:0]
		for _, v := range arr {
			players = append(players, BlackJackGetPlayerData(this.Scene, v))
		}
		arr = arr[:0]
		b := int64(-1)
		for _, v := range players {
			coin := v.BetCoin()
			if coin < b {
				continue
			}
			if coin > b {
				arr = arr[:0]
				b = coin
			}
			arr = append(arr, v.seat)
		}
		if len(arr) < 1 {
			break
		}
		if len(arr) == 1 {
			seat = arr[0]
			break
		}
		// SnID
		players = players[:0]
		for _, v := range arr {
			players = append(players, BlackJackGetPlayerData(this.Scene, v))
		}
		sort.SliceStable(players, func(i, j int) bool {
			return players[i].SnId < players[j].SnId
		})
		seat = players[0].seat
		break
	}
	if seat < 1 {
		return
	}
	logger.Logger.Trace("--> One seat ", seat)
	// 换座位
	pd := BlackJackGetPlayerData(this.Scene, seat)
	for k, v := range this.chairIDs {
		if v == pd {
			for i := k; i > 1; i-- {
				this.chairIDs[i] = this.chairIDs[i-1]
			}
			break
		}
	}
	this.chairIDs[1] = pd
	// 广播换位消息
	pack := &blackjack.SCBlackJackPos{}
	for _, v := range this.chairIDs[1:] {
		pack.Seats = append(pack.Seats, int32(v.seat))
	}
	proto.SetDefaults(pack)
	this.Broadcast(int(blackjack.BlackJackPacketID_SC_PLAYER_Pos), pack, 0)
	logger.Logger.Trace("--> Broadcast SCBlackJackPos ", pack.String())
}

// 发牌
func (this *BlackJackSceneData) FaPai() {
	num := 0
	for i := 0; i < 2; i++ {
		for _, v := range this.chairIDs {
			if v.isBet || v.isBanker {
				num++
				v.hands[0].handCards = append(v.hands[0].handCards, this.GetCard(1)...)
				//这里不要删除便于给客户端做测试
				//if v.Player != nil && !v.IsRob {
				//	v.hands[0].handCards = []*rule.Card{rule.NewCardDefaultA(), rule.NewCardDefaultJ()}
				//}
			}
		}
	}
	n := time.Duration((num-1)*300 + 850)
	this.timeOutDeal = time.Second * n

	// test
	//庄家第一张为A
	//玩家 22 或 56
	//cs := [][]*rule.Card{
	//	//{rule.Cards[1], rule.Cards[14]},  // 22
	//	//{rule.Cards[4], rule.Cards[18]}, // 56
	//	//{rule.Cards[0], rule.Cards[10]},  // A11
	//	//{rule.Cards[9], rule.Cards[10]}, // 10J
	//	//{rule.Cards[10], rule.Cards[9]},  // J10
	//	//{rule.Cards[10], rule.Cards[0]},  // JA
	//	{rule.Cards[10], rule.Cards[10]}, // JJ
	//	//{rule.Cards[0], rule.Cards[0]}, // AA
	//}
	//bcs := [][]*rule.Card{
	//	//{rule.Cards[9], rule.Cards[1]},   // 10 2
	//	//{rule.Cards[10], rule.Cards[0]},  // J A
	//	//{rule.Cards[0], rule.Cards[10]}, // A J
	//	{rule.Cards[10], rule.Cards[10]}, // J J
	//	//{rule.Cards[0], rule.Cards[13]},  // A A
	//	//{rule.Cards[0], rule.Cards[5]}, // A 6
	//}
	//this.chairIDs[0].hands[0].handCards = bcs[common.RandInt(len(bcs))]
	//for _, v := range this.chairIDs[1:] {
	//	if v.isBet {
	//		v.hands[0].handCards = make([]*rule.Card, 2)
	//		copy(v.hands[0].handCards, cs[common.RandInt(len(cs))])
	//	}
	//}
	// test

	// 类型和点数
	for _, v := range this.chairIDs {
		handUpdate(v)
	}

	if !this.Testing {
		// 黑白名单发牌调控
		this.blackWhile()
		// 赔率调控
		if model.IsCheckPlayerRateControl(this.GetKeyGameId(), this.GetGameFreeId()) {
			this.RateControl()
		}
	}
}

func (this *BlackJackSceneData) blackWhile() {
	var has bool
	var hands []BlackJackHand
	var seats []*BlackJackPlayerData
	for i := 0; i <= rule.MaxPlayer; i++ {
		if this.chairIDs[i].isBet || this.chairIDs[i].isBanker {
			hands = append(hands, this.chairIDs[i].hands[0])
			seats = append(seats, this.chairIDs[i])
			if this.chairIDs[i].Player != nil && (this.chairIDs[i].BlackLevel > 0 || this.chairIDs[i].WhiteLevel+this.chairIDs[i].WhiteFlag > 0) {
				has = true
			}
		}
	}
	if !has {
		return
	}
	sort.SliceStable(hands, func(i, j int) bool {
		res, err := rule.CompareCards(hands[i].handCards, hands[j].handCards)
		if err != nil {
			logger.Error("blackWhile error ", err)
		}
		return res > 0
	})
	sort.SliceStable(seats, func(i, j int) bool {
		if seats[i].isBanker {
			if seats[j].isBanker {
				return false
			} else {
				return 0 > seats[j].WhiteLevel+seats[j].WhiteFlag
			}
		} else {
			if seats[j].isBanker {
				return seats[i].WhiteLevel+seats[i].WhiteFlag > 0
			} else {
				return seats[i].WhiteLevel+seats[i].WhiteFlag > seats[j].WhiteLevel+seats[j].WhiteFlag
			}
		}
	})
	sort.SliceStable(seats, func(i, j int) bool {
		if seats[i].isBanker {
			if seats[j].isBanker {
				return false
			} else {
				return 0 < seats[j].BlackLevel
			}
		} else {
			if seats[j].isBanker {
				return seats[i].BlackLevel < 0
			} else {
				return seats[i].BlackLevel < seats[j].BlackLevel
			}
		}
	})
	for k, v := range seats {
		v.hands[0].handCards = hands[k].handCards
		v.hands[0].point = hands[k].point
		v.hands[0].tp = hands[k].tp
		v.hands[0].state = hands[k].state
	}
}

func (this *BlackJackSceneData) RateControl() {
	// 库存值低于库存下限值，不做细化调控
	if status, _ := base.CoinPoolMgr.GetCoinPoolStatus(this.GetPlatform(), this.GetGameFreeId(), this.GetGroupId()); status == base.CoinPoolStatus_Low {
		return
	}

	players := make(map[int]*BlackJackPlayerData)
	for k, v := range this.chairIDs[1:] {
		if v.Player != nil && !v.IsRobot() && v.IsGameing() && v.BlackLevel == 0 && v.WhiteLevel+v.WhiteFlag == 0 {
			players[k+1] = v
		}
	}

	banker := this.chairIDs[0]

	for _, v := range players {
		// 是否已经存在分段调控
		if data, ok := v.TodayGameData.CtrlData[this.GetKeyGameId()]; ok {
			if data.WinGameTimes > 0 || data.LoseGameTimes > 0 {
				var rate float64
				if ydata, ok := v.YesterdayGameData.CtrlData[this.GetKeyGameId()]; ok {
					rate = float64(ydata.TotalOut+1) / float64(ydata.TotalIn+1)
				}
				if rate > 1 {
					if data.LoseGameTimes > 0 {
						data.LoseGameTimes--
						if this.RandInt(100) < 40 {
							// 标记该玩家亏钱状态
							v.loseStatus = true
						}
					} else if data.WinGameTimes > 0 {
						data.WinGameTimes--
						if this.RandInt(100) < 40 {
							// 标记该玩家大赢钱状态
							v.winStatus = true
						}
					}
				} else {
					// 输赢调控已经存在
					if data.LoseGameTimes > 0 {
						data.LoseGameTimes--
						if this.RandInt(100) < 40 {
							// 拿不到最大牌(比庄家牌小)
							if v.hands[0].point[len(v.hands[0].point)-1] > banker.hands[0].point[len(banker.hands[0].point)-1] {
								handsSwap(v, banker)
							}
							// 标记该玩家亏钱状态
							v.loseStatus = true
						}
					} else if data.WinGameTimes > 0 {
						data.WinGameTimes--
						if this.RandInt(100) < 40 {
							// 拿到最大牌(比庄家牌大)
							if v.hands[0].point[len(v.hands[0].point)-1] < banker.hands[0].point[len(banker.hands[0].point)-1] {
								handsSwap(v, banker)
							}
							// 标记该玩家大赢钱状态
							v.winStatus = true
						}
					}
				}
				continue
			}
		}
		// 赔率调控
		if data, ok := v.GDatas[this.GetKeyGameId()]; ok {
			//投入产出比
			totalOut := data.Statics.TotalOut
			totalIn := data.Statics.TotalIn
			CoinPayTotal := v.CoinPayTotal
			rate := float64(totalOut-totalIn) / float64(CoinPayTotal+1)
			n := 0
			if rate > 1 && rate <= 2 {
				n = 20
			} else if rate > 2 && rate <= 3 {
				n = 40
			} else if rate > 3 && rate <= 5 {
				n = 60
			} else if rate > 5 && rate <= 10 {
				n = 80
			} else if rate > 10 {
				n = 100
			}
			if n != 0 {
				if this.RandInt(100) < n {
					// 标记玩家亏钱状态
					v.lossStatus = true
				}
				continue
			}
		}
		// 检查是否需要分段调控
		if ydata, ok := v.YesterdayGameData.CtrlData[this.GetKeyGameId()]; ok {
			rate := float64(ydata.TotalOut+1) / float64(ydata.TotalIn+1)
			if tdata, ok := v.TodayGameData.CtrlData[this.GetKeyGameId()]; ok {
				if tdata.TotalIn > 500 && rate > 1 {
					if (tdata.TotalIn - tdata.TotalOut) < (ydata.TotalOut-ydata.TotalIn)*(rand.Int63n(30)+50)/100 {
						lose := float64(ydata.TotalOut-ydata.TotalIn+v.GetCoin()) / float64(v.TodayGameData.RechargeCoin+10000)
						r := float64(rand.Intn(10)) / 10
						if lose > 0 && lose < r {
							tdata.LoseGameTimes = 10
							tdata.LoseGameTimes--
							if this.RandInt(100) < 40 {
								// 标记该玩家亏钱状态
								v.loseStatus = true
							}
						}
					} else if (tdata.TotalIn - tdata.TotalOut) > (ydata.TotalOut-ydata.TotalIn)*(rand.Int63n(70)+80)/100 {
						win := float64(v.TodayGameData.RechargeCoin) / float64(ydata.TotalOut-ydata.TotalIn)
						r := float64(rand.Intn(10)) / 10
						if win > 0 && win < r {
							tdata.WinGameTimes = 10
							tdata.WinGameTimes--
							if this.RandInt(100) < 40 {
								// 标记该玩家大赢钱状态
								v.winStatus = true
							}
						}
					}
				} else if tdata.TotalIn > 500 && rate <= 1 {
					if (tdata.TotalOut - tdata.TotalIn) < (ydata.TotalIn-ydata.TotalOut)*int64(rand.Intn(20)+30)+1000+
						v.TodayGameData.RechargeCoin*(int64(this.RandInt(30)+20)) {
						win := (float64(ydata.TotalOut)*0.8 + float64(v.TodayGameData.RechargeCoin)) / float64(ydata.TotalOut+10000)
						r := float64(this.RandInt(10)) / 10
						if win > 0 && win < r {
							tdata.WinGameTimes = 10
							tdata.WinGameTimes--
							if this.RandInt(100) < 40 {
								// 拿到最大牌
								if v.hands[0].point[len(v.hands[0].point)-1] < banker.hands[0].point[len(banker.hands[0].point)-1] {
									handsSwap(v, banker)
								}
								// 标记该玩家大赢钱状态
								v.winStatus = true
							}
						}
					} else if (tdata.TotalOut - tdata.TotalIn) >= (ydata.TotalIn-ydata.TotalOut)*(rand.Int63n(20)+50)+
						v.TodayGameData.RechargeCoin*(int64(this.RandInt(70)+80)) {
						lose := float64(tdata.TotalOut-tdata.TotalIn) / float64(v.TodayGameData.RechargeCoin+10000)
						r := float64(this.RandInt(10)) / 10
						if lose > 0 && lose < r {
							tdata.LoseGameTimes = 10
							tdata.LoseGameTimes--
							if this.RandInt(100) < 40 {
								// 拿不到最大牌
								if v.hands[0].point[len(v.hands[0].point)-1] > banker.hands[0].point[len(banker.hands[0].point)-1] {
									handsSwap(v, banker)
								}
								// 标记该玩家亏钱状态
								v.loseStatus = true
							}
						}
					}
				}
			}
		}
	}
}

func handsSwap(v, banker *BlackJackPlayerData) {
	v.hands[0].handCards, banker.hands[0].handCards = banker.hands[0].handCards, v.hands[0].handCards
	v.hands[0].tp, banker.hands[0].tp = banker.hands[0].tp, v.hands[0].tp
	v.hands[0].point, banker.hands[0].point = banker.hands[0].point, v.hands[0].point
	v.hands[0].state, banker.hands[0].state = banker.hands[0].state, v.hands[0].state
}

func handUpdate(p *BlackJackPlayerData) {
	for i := 0; i < len(p.hands); i++ {
		if len(p.hands[i].handCards) == 0 {
			continue
		}
		p.hands[i].tp, p.hands[i].point = rule.GetCardsType(p.hands[i].handCards)
		// 黑杰克强制停牌
		if p.hands[i].tp == rule.CardTypeA10 {
			p.hands[i].state = 1
		}
	}
}

func (this *BlackJackSceneData) eq(point int) *rule.Card {
	return nil
}

func (this *BlackJackSceneData) ne(point int) *rule.Card {
	return nil
}

func (this *BlackJackSceneData) gt(point int) bool {
	for k, v := range this.pokers {
		if v.Point() > point {
			this.pokers[0], this.pokers[k] = this.pokers[k], this.pokers[0]
			return true
		}
	}
	return false
}

func (this *BlackJackSceneData) ge(point int) bool {
	for k, v := range this.pokers {
		if v.Point() >= point {
			this.pokers[0], this.pokers[k] = this.pokers[k], this.pokers[0]
			return true
		}
	}
	return false
}

func (this *BlackJackSceneData) lt(point int) bool {
	for k, v := range this.pokers {
		x := v.Point()
		if x == 1 {
			x = 11
		}
		if x < point {
			this.pokers[0], this.pokers[k] = this.pokers[k], this.pokers[0]
			return true
		}
	}
	return false
}

func (this *BlackJackSceneData) le(point int) bool {
	for k, v := range this.pokers {
		x := v.Point()
		if x == 1 {
			x = 11
		}
		if x <= point {
			this.pokers[0], this.pokers[k] = this.pokers[k], this.pokers[0]
			return true
		}
	}
	return false
}

func (this *BlackJackSceneData) max(hand *BlackJackHand) int {
	if len(hand.point) == 1 {
		return int(hand.point[0])
	}
	for i := len(hand.point) - 1; i >= 0; i-- {
		if hand.point[i] <= 21 {
			return int(hand.point[i])
		}
	}
	return int(hand.point[0])
}

// 分牌
func (this *BlackJackSceneData) FenPai(p *BlackJackPlayerData) (isSuccess bool) {
	// 是否可分
	if len(p.hands[0].handCards) > 0 && len(p.hands[1].handCards) > 0 {
		return
	}
	if p.hands[0].state == 1 || len(p.hands[0].handCards) != 2 ||
		p.hands[0].handCards[0].Point() != p.hands[0].handCards[1].Point() {
		return
	}
	// 玩家金币 < 玩家下注额 + 当前下注额	(金币不足)
	if this.PlayerCoin(p) < p.BetCoin() {
		pack := &blackjack.SCBlackJackPlayerOperate{
			Code:    blackjack.SCBlackJackPlayerOperate_ErrCoin,
			Operate: proto.Int32(rule.SubFenPai),
			Pos:     proto.Int32(int32(p.seat)),
		}
		proto.SetDefaults(pack)
		p.SendToClient(int(blackjack.BlackJackPacketID_SC_PLAYER_OPERATE), pack)
		return
	}

	// 分牌
	var cards []*rule.Card
	// 分牌调控
	// 1.黑白名单调控 或 水池调控 （黑白名单调控优先）
	// 2.赔率调控 (和上面的调控互斥)
	if !this.Testing && !p.IsRobot() {
		if p.BlackLevel > 0 { // 黑名单
			for i := 0; i < 2; i++ {
				if this.RandInt(100) < int(p.BlackLevel)*10 {
					x := p.hands[0].handCards[i].Point()
					if x == 1 {
						x = 11
					}
					this.lt(17 - x)
					cards = append(cards, this.GetCard(1)[0])
				}
			}
		}
		if p.WhiteLevel+p.WhiteFlag > 0 { // 白名单
			for i := 0; i < 2; i++ {
				if this.RandInt(100) < int(p.WhiteLevel+p.WhiteFlag)*10 {
					x := p.hands[0].handCards[i].Point()
					if x == 1 {
						x = 11
					}
					this.ge(17 - x)
					cards = append(cards, this.GetCard(1)[0])
				}
			}
		}
		// 水池调控
		if len(cards) == 0 && model.IsUseCoinPoolControlGame(this.GetKeyGameId(), this.GetGameFreeId()) {
			cards = this.CoinPoolFenPai(&p.hands[0])
		}
		// 赔率调控
		if len(cards) == 0 && model.IsCheckPlayerRateControl(this.GetKeyGameId(), this.GetGameFreeId()) {
			if p.lossStatus || p.loseStatus {
				cards = this.FenPaiLow(&p.hands[0])
			} else if p.winStatus {
				cards = this.FenPaiBigWin(&p.hands[0])
			}
		}
	}
	for len(cards) < 2 {
		cards = append(cards, this.GetCard(1)[0])
	}
	// test
	//cards = []*rule.Card{rule.Cards[0], rule.Cards[1]} // A 2
	//cards = []*rule.Card{rule.Cards[1], rule.Cards[0]} // 2 A
	//cards = []*rule.Card{rule.Cards[5], rule.Cards[4]} // 6 5
	//cards = []*rule.Card{rule.Cards[8], rule.Cards[8]} // 9 9
	//cards = []*rule.Card{rule.Cards[9], rule.Cards[9]} // 10 10
	//cards = []*rule.Card{rule.Cards[9], rule.Cards[1]} // 10 2
	//cards = []*rule.Card{rule.Cards[9], rule.Cards[5]} // 10 6
	//cards = []*rule.Card{rule.Cards[0], rule.Cards[0]} // A A
	// tes

	p.hands[1].handCards = []*rule.Card{p.hands[0].handCards[1], cards[1]}
	p.hands[0].handCards[1] = cards[0]
	p.hands[1].betCoin = p.hands[0].betCoin
	for i := 0; i < 2; i++ {
		p.hands[i].tp, p.hands[i].point = rule.GetCardsType(p.hands[i].handCards)
		// 黑杰克强制停牌
		if p.hands[i].tp == rule.CardTypeA10 {
			p.hands[i].state = 1
		}
	}

	// 广播分牌
	pack := &blackjack.SCBlackJackPlayerOperate{
		Code:    blackjack.SCBlackJackPlayerOperate_Success,
		Operate: proto.Int32(rule.SubFenPai),
		Pos:     proto.Int32(int32(p.seat)),
		ReCoin:  proto.Int64(this.PlayerCoin(p)),
		Num:     proto.Int32(int32(len(this.pokers))),
		BetCoin: proto.Int64(p.hands[1].betCoin),
	}
	for i := 0; i < 2; i++ {
		hand := &blackjack.BlackJackCards{
			Cards:   []int32{int32(cards[i].Value())},
			DCards:  proto.Int32(int32(p.hands[i].mulCards.Value())),
			Type:    proto.Int32(p.hands[i].tp),
			Point:   p.hands[i].point,
			State:   proto.Int32(p.hands[i].state),
			Id:      proto.Int32(int32(i)),
			Seat:    proto.Int32(int32(p.seat)),
			BetCoin: proto.Int64(p.hands[i].betCoin),
		}
		pack.Cards = append(pack.Cards, hand)
	}
	this.playerOperate(p, pack)
	logger.Logger.Trace("--> FenPai SCBlackJackPlayerOperate ", pack.String())
	pExtra, _ := p.GetExtraData().(*BlackJackPlayerData)
	pExtra.opCodeFlag = rule.SubFenPai
	pExtra.opRightFlag = rule.OpDefault
	if p.hands[0].state == 1 {
		this.NotifySkip(p, rule.SubSkipLeft)
		pExtra.opRightFlag = rule.OpRight
	}
	return true
}

// 双倍
func (this *BlackJackSceneData) Double(p *BlackJackPlayerData) (isSuccess bool) {
	// 是否可以双倍
	i := 0 // 双倍操作的手牌
	if p.hands[0].state == 0 {
		if len(p.hands[0].handCards) != 2 {
			return
		}
	} else if p.hands[1].state == 0 {
		if len(p.hands[1].handCards) != 2 {
			return
		}
		i = 1
	} else {
		return
	}
	// 点数为11才能双倍
	//p1 := p.hands[i].handCards[0].Point()
	//p2 := p.hands[i].handCards[1].Point()
	//if p1 == 1 || p2 == 1 {
	//	if p1 == 1 && p2 == 1 {
	//		return
	//	}
	//	if p1 == 1 && p2 != 10 {
	//		return
	//	}
	//	if p2 == 1 && p1 != 10 {
	//		return
	//	}
	//} else {
	//	if p1+p2 != 11 {
	//		return
	//	}
	//}
	// 玩家金币 < 当前下注金额	(金币不足)
	if this.PlayerCoin(p) < p.hands[i].betCoin {
		pack := &blackjack.SCBlackJackPlayerOperate{
			Code:    blackjack.SCBlackJackPlayerOperate_ErrCoin,
			Operate: proto.Int32(rule.SubFenPai),
			Pos:     proto.Int32(int32(p.seat)),
		}
		proto.SetDefaults(pack)
		p.SendToClient(int(blackjack.BlackJackPacketID_SC_PLAYER_OPERATE), pack)
		return
	}
	// 双倍
	var card *rule.Card
	// 双倍调控
	// 1.黑白名单调控 或 水池调控 （黑白名单调控优先）
	// 2.赔率调控 (和上面的调控互斥)
	if !this.GetTesting() && !p.IsRobot() {
		if p.BlackLevel > 0 && this.RandInt(100) < int(p.BlackLevel)*10 {
			xMin, xMax := int(p.hands[i].point[0]), int(p.hands[i].point[len(p.hands[i].point)-1])
			// 寻找牌c,使牌c+xMax <= 17 或 c+xMin > 21
			if this.le(17-xMax) || this.gt(21-xMin) {
				card = this.GetCard(1)[0]
			}
		}
		if p.WhiteLevel+p.WhiteFlag > 0 && this.RandInt(100) < int(p.WhiteLevel+p.WhiteFlag)*10 {
			x := int(p.hands[i].point[len(p.hands[i].point)-1]) // 最大值
			if x <= 16 {
				if this.ge(17 - x) {
					card = this.GetCard(1)[0]
				}
			} else {
				x = int(p.hands[i].point[0]) // 最小值
				if this.le(21 - x) {
					card = this.GetCard(1)[0]
				}
			}
		}
		// 水池调控
		if card == nil && model.IsUseCoinPoolControlGame(this.GetKeyGameId(), this.GetGameFreeId()) {
			card = this.CoinPoolDouble(&p.hands[i])
		}
		// 赔率调控
		if card == nil && model.IsCheckPlayerRateControl(this.GetKeyGameId(), this.GetGameFreeId()) {
			if p.lossStatus || p.loseStatus {
				card = this.DoubleLow(&p.hands[i])
			} else if p.winStatus {
				card = this.DoubleBigWin(&p.hands[i])
			}
		}
	}
	if card == nil {
		card = this.GetCard(1)[0]
		//card = rule.Cards[0]
	}

	// test
	//card = rule.Cards[0]
	// test

	p.hands[i].state = 1
	p.hands[i].handCards = append(p.hands[i].handCards, card)
	p.hands[i].mulCards = card
	p.hands[i].tp, p.hands[i].point = rule.GetCardsType(p.hands[i].handCards)
	p.hands[i].betCoin *= 2

	// 广播双倍
	pack := &blackjack.SCBlackJackPlayerOperate{
		Code:    blackjack.SCBlackJackPlayerOperate_Success,
		Operate: proto.Int32(rule.SubDouble),
		Pos:     proto.Int32(int32(p.seat)),
		ReCoin:  proto.Int64(this.PlayerCoin(p)),
		Num:     proto.Int32(int32(len(this.pokers))),
	}
	if p.hands[i].tp == rule.CardTypeBoom {
		pack.BetCoin = proto.Int64(-p.hands[i].betCoin)
	} else {
		pack.BetCoin = proto.Int64(p.hands[i].betCoin / 2)
	}
	pack.Cards = []*blackjack.BlackJackCards{
		{
			Cards:   []int32{int32(card.Value())},
			DCards:  proto.Int32(int32(card.Value())),
			Type:    proto.Int32(p.hands[i].tp),
			Point:   p.hands[i].point,
			State:   proto.Int32(p.hands[i].state),
			Id:      proto.Int32(int32(i)),
			BetCoin: proto.Int64(p.hands[i].betCoin),
		},
	}
	this.playerOperate(p, pack)
	logger.Logger.Trace("--> Double SCBlackJackPlayerOperate ", pack.String())
	// 爆牌停牌
	pExtra, _ := p.GetExtraData().(*BlackJackPlayerData)
	pExtra.opRightFlag = rule.OpDefault
	if p.hands[0].state == 1 {
		pExtra.opRightFlag = rule.OpRight
	}
	pExtra.opCodeFlag = int32(rule.SubDouble)
	if p.hands[i].tp == rule.CardTypeBoom {
		this.NotifySkip(p, rule.SubSkipBomb)
	} else if i == 0 {
		this.NotifySkip(p, rule.SubSkipLeft)
	} else {
		this.NotifySkip(p)
	}
	return true
}

// 要牌
func (this *BlackJackSceneData) Outs(p *BlackJackPlayerData) (isSuccess bool) {
	// 是否可以要牌
	i := 0
	if p.hands[0].state == 1 {
		if len(p.hands[1].handCards) == 0 {
			return
		}
		if p.hands[1].state == 1 {
			return
		}
		i = 1
	}

	// 要牌
	var card *rule.Card
	// 要牌调控
	// 1.黑白名单调控 或 水池调控 （黑白名单调控优先）
	// 2.赔率调控 (和上面的调控互斥)
	if !this.GetTesting() && !p.IsRobot() {
		if p.BlackLevel > 0 && this.RandInt(100) < int(p.BlackLevel)*10 {
			xMin, xMax := int(p.hands[i].point[0]), int(p.hands[i].point[len(p.hands[i].point)-1])
			// 寻找牌c,使牌c+xMax <= 17 或 c+xMin > 21
			if this.le(17-xMax) || this.gt(21-xMin) {
				card = this.GetCard(1)[0]
			}
		}
		if p.WhiteLevel+p.WhiteFlag > 0 && this.RandInt(100) < int(p.WhiteLevel+p.WhiteFlag)*10 {
			x := int(p.hands[i].point[len(p.hands[i].point)-1]) // 最大值
			if x <= 9 {
				if this.ge(17 - x) {
					card = this.GetCard(1)[0]
				}
			} else {
				x = int(p.hands[i].point[0]) // 最小值
				if this.le(21 - x) {
					card = this.GetCard(1)[0]
				}
			}
		}
		// 水池调控
		if card == nil && model.IsUseCoinPoolControlGame(this.GetKeyGameId(), this.GetGameFreeId()) {
			card = this.CoinPoolOuts(&p.hands[i])
		}
		// 赔率调控
		if card == nil && model.IsCheckPlayerRateControl(this.GetKeyGameId(), this.GetGameFreeId()) {
			if p.lossStatus || p.loseStatus {
				card = this.OutsLow(&p.hands[i])
			} else if p.winStatus {
				card = this.OutsBigWin(&p.hands[i])
			}
		}
	}
	if card == nil {
		card = this.GetCard(1)[0]
		//card = rule.Cards[0]
	}

	// test
	//card = rule.Cards[9]
	// test

	p.hands[i].handCards = append(p.hands[i].handCards, card)
	p.hands[i].tp, p.hands[i].point = rule.GetCardsType(p.hands[i].handCards)
	// 判断停牌
	if p.hands[i].tp == rule.CardTypeBoom ||
		len(p.hands[i].handCards) >= rule.MaxCardNum ||
		p.hands[i].point[len(p.hands[i].point)-1] == 21 {
		p.hands[i].state = 1
	}

	// 广播要牌
	pack := &blackjack.SCBlackJackPlayerOperate{
		Code:    blackjack.SCBlackJackPlayerOperate_Success,
		Operate: proto.Int32(rule.SubOuts),
		Pos:     proto.Int32(int32(p.seat)),
		Num:     proto.Int32(int32(len(this.pokers))),
		ReCoin:  proto.Int64(this.PlayerCoin(p)),
	}
	pack.Cards = []*blackjack.BlackJackCards{
		{
			Cards:   []int32{int32(card.Value())},
			DCards:  proto.Int32(int32(p.hands[i].mulCards.Value())),
			Type:    proto.Int32(p.hands[i].tp),
			Point:   p.hands[i].point,
			State:   proto.Int32(p.hands[i].state),
			Id:      proto.Int32(int32(i)),
			BetCoin: proto.Int64(p.hands[i].betCoin),
		},
	}
	if p.hands[i].tp == rule.CardTypeBoom {
		pack.BetCoin = proto.Int64(-p.hands[i].betCoin)
	}
	this.playerOperate(p, pack)
	logger.Logger.Trace("--> Outs SCBlackJackPlayerOperate ", pack.String())
	// 爆牌停牌
	pExtra, _ := p.GetExtraData().(*BlackJackPlayerData)
	pExtra.opRightFlag = rule.OpDefault
	if p.hands[0].state == 1 {
		pExtra.opRightFlag = rule.OpRight
	}
	pExtra.opCodeFlag = int32(rule.SubOuts)
	if p.hands[i].tp == rule.CardTypeBoom {
		this.NotifySkip(p, rule.SubSkipBomb)
	} else {
		if i == 0 && p.hands[i].state == 1 {
			this.NotifySkip(p, rule.SubSkipLeft)
		} else if i == 1 && p.hands[i].state == 1 {
			this.NotifySkip(p)
		}
	}
	return true
}

func (this *BlackJackSceneData) CoinPoolFenPai(hand *BlackJackHand) []*rule.Card {
	status, _ := base.CoinPoolMgr.GetCoinPoolStatus(this.GetPlatform(), this.GetGameFreeId(), this.GetGroupId())
	switch status {
	case base.CoinPoolStatus_Normal: // 正常状态

	case base.CoinPoolStatus_Low: // 亏钱状态
		return this.FenPaiLow(hand)

	case base.CoinPoolStatus_High: // 赢钱状态
		var ret []*rule.Card
		for i := 0; i < 2; i++ {
			x := hand.handCards[i].Point()
			if x == 1 {
				x = 11
			}
			if this.RandInt(100) < 80 {
				this.ge(17 - x)
			}
			ret = append(ret, this.GetCard(1)[0])
		}
		return ret

	case base.CoinPoolStatus_TooHigh: // 大赢钱状态
		return this.FenPaiBigWin(hand)
	}
	return []*rule.Card{}
}

func (this *BlackJackSceneData) CoinPoolOuts(hand *BlackJackHand) *rule.Card {
	status, _ := base.CoinPoolMgr.GetCoinPoolStatus(this.GetPlatform(), this.GetGameFreeId(), this.GetGroupId())
	switch status {
	case base.CoinPoolStatus_Normal:

	case base.CoinPoolStatus_Low:
		return this.OutsLow(hand)

	case base.CoinPoolStatus_High:
		x := int(hand.point[len(hand.point)-1])
		if x <= 9 {
			if this.RandInt(100) < 80 {
				this.ge(17 - x)
			}
		} else {
			x = int(hand.point[0])
			if this.RandInt(100) < 80 {
				this.le(21 - x)
			}
		}
		return this.GetCard(1)[0]

	case base.CoinPoolStatus_TooHigh:
		return this.OutsBigWin(hand)
	}
	return nil
}

func (this *BlackJackSceneData) CoinPoolDouble(hand *BlackJackHand) *rule.Card {
	status, _ := base.CoinPoolMgr.GetCoinPoolStatus(this.GetPlatform(), this.GetGameFreeId(), this.GetGroupId())
	switch status {
	case base.CoinPoolStatus_Normal:

	case base.CoinPoolStatus_Low:
		return this.DoubleLow(hand)

	case base.CoinPoolStatus_High:
		x := int(hand.point[0])
		if x >= 10 {
			if this.RandInt(100) < 80 {
				this.le(21 - x)
			}
		}
		return this.GetCard(1)[0]

	case base.CoinPoolStatus_TooHigh:
		return this.DoubleBigWin(hand)
	}
	return nil
}

//=================
// 亏钱状态
//=================
func (this *BlackJackSceneData) FenPaiLow(hand *BlackJackHand) []*rule.Card {
	var ret []*rule.Card
	for i := 0; i < 2; i++ {
		x := hand.handCards[i].Point()
		if x == 1 {
			x = 11
		}
		this.lt(17 - x)
		ret = append(ret, this.GetCard(1)[0])
	}
	return ret
}

func (this *BlackJackSceneData) OutsLow(hand *BlackJackHand) *rule.Card {
	xMin, xMax := int(hand.point[0]), int(hand.point[len(hand.point)-1])
	// 寻找c,使c+xMax<=17 或 c+xMin > 21
	if this.lt(17-xMax) || this.gt(21-xMin) {
		return this.GetCard(1)[0]
	}
	return nil
}

func (this *BlackJackSceneData) DoubleLow(hand *BlackJackHand) *rule.Card {
	return this.OutsLow(hand)
}

//=================
// 大赢钱状态
//=================
func (this *BlackJackSceneData) FenPaiBigWin(hand *BlackJackHand) []*rule.Card {
	var ret []*rule.Card
	for i := 0; i < 2; i++ {
		x := hand.handCards[i].Point()
		if x == 1 {
			x = 11
		}
		this.ge(17 - x)
		ret = append(ret, this.GetCard(1)[0])
	}
	return ret
}

func (this *BlackJackSceneData) OutsBigWin(hand *BlackJackHand) *rule.Card {
	x := int(hand.point[len(hand.point)-1])
	if x <= 9 {
		this.ge(17 - x)
	} else {
		x = int(hand.point[0])
		this.le(21 - x)
	}
	return this.GetCard(1)[0]
}

func (this *BlackJackSceneData) DoubleBigWin(hand *BlackJackHand) *rule.Card {
	x := int(hand.point[0])
	if x >= 10 {
		this.le(21 - x)
	}
	return this.GetCard(1)[0]
}

func (this *BlackJackSceneData) AllSkip(p *BlackJackPlayerData) (isSuccess bool) {
	for i := 0; i < 2; i++ {
		if p.hands[i].state != 1 && len(p.hands[i].handCards) > 0 {
			p.hands[i].state = 1
			break
		}
	}
	if p.hands[0].state == 1 && len(p.hands[1].handCards) > 0 && p.hands[1].state != 1 {
		pExtra, _ := p.GetExtraData().(*BlackJackPlayerData)
		pExtra.opCodeFlag = rule.SubSkipLeft
		pExtra.opRightFlag = rule.OpRight
		this.NotifySkip(p, rule.SubSkipLeft)
	} else {
		this.NotifySkip(p)
	}
	this.StateStartTime = time.Now()
	return true
}

// 停牌
func (this *BlackJackSceneData) Skip(p *BlackJackPlayerData) (isSuccess bool) {
	pExtra, _ := p.GetExtraData().(*BlackJackPlayerData)
	for i := 0; i < 2; i++ {
		if p.hands[i].state != 1 && len(p.hands[i].handCards) > 0 {
			p.hands[i].state = 1
			if i == 0 {
				// 广播停牌
				pack := &blackjack.SCBlackJackPlayerOperate{
					Code:    blackjack.SCBlackJackPlayerOperate_Success,
					Operate: proto.Int32(rule.SubSkipLeft),
					Pos:     proto.Int32(int32(p.seat)),
					Num:     proto.Int32(int32(len(this.pokers))),
				}
				this.playerOperate(p, pack)
				pExtra.opCodeFlag = rule.SubSkipLeft
				pExtra.opRightFlag = rule.OpRight
				logger.Logger.Trace("--> Skip SCBlackJackPlayerOperate ", pack.String())
			} else {
				this.NotifySkip(p)
			}
			break
		}
	}
	return true
}

func (this *BlackJackSceneData) GetCard(n int) (cards []*rule.Card) {
	//return []*rule.Card{rule.Cards[1], rule.Cards[1]}
	cards = make([]*rule.Card, n)
	copy(cards, this.pokers[:n])
	this.pokers = this.pokers[n:]
	return cards
}

// 放回到牌堆
func (this *BlackJackSceneData) Push(c *rule.Card) {
	if c == nil {
		return
	}
	this.pokers = append(this.pokers, c)
	i := this.RandInt(len(this.pokers))
	this.pokers[len(this.pokers)-1], this.pokers[i] = this.pokers[i], this.pokers[len(this.pokers)-1]
}

// 从牌堆中取出
func (this *BlackJackSceneData) Remove(c *rule.Card) bool {
	if c == nil {
		return false
	}
	for k, v := range this.pokers {
		if v.Value() == c.Value() {
			this.pokers = append(this.pokers[:k], this.pokers[k+1:]...)
			return true
		}
	}
	return false
}

// 买保险
func (this *BlackJackSceneData) Buy() {
	seat := this.chairIDs[this.pos]
	seat.isBuy = true
	seat.baoCoin = seat.BetCoin() / 2
	pack := &blackjack.SCBlackJackBuy{
		Code:   blackjack.SCBlackJackBuy_Success,
		Pos:    proto.Int32(int32(seat.seat)),
		Coin:   proto.Int64(seat.baoCoin),
		ReCoin: proto.Int64(this.PlayerCoin(seat)),
	}
	this.Broadcast(int(blackjack.BlackJackPacketID_SC_PLAYER_BUY), pack, 0)
	logger.Logger.Trace("--> Broadcast SCBlackJackBuy ", pack.String())
	if !this.GetTesting() && !this.buy && model.IsUseCoinPoolControlGame(this.GetKeyGameId(), this.GetGameFreeId()) {
		this.buy = true
		// 根据水池状态进行庄家换牌操作
		status, _ := base.CoinPoolMgr.GetCoinPoolStatus(this.GetPlatform(), this.GetGameFreeId(), this.GetGroupId())
		switch status {
		case base.CoinPoolStatus_Low:
			if this.chairIDs[0].hands[0].tp == rule.CardTypeA10 {
				// 调为非黑杰克
				for k, v := range this.pokers {
					if v.Point() != 10 {
						this.bankerSwap(k)
						break
					}
				}
			}
		case base.CoinPoolStatus_High:
			if this.chairIDs[0].hands[0].tp != rule.CardTypeA10 {
				// 50%调为黑杰克
				if this.RandInt(100) < 50 {
					for k, v := range this.pokers {
						if v.Point() == 10 {
							this.bankerSwap(k)
							break
						}
					}
				}
			}
		case base.CoinPoolStatus_TooHigh:
			if this.chairIDs[0].hands[0].tp != rule.CardTypeA10 {
				// 100%调为黑杰克
				for k, v := range this.pokers {
					if v.Point() == 10 {
						this.bankerSwap(k)
						break
					}
				}
			}
		}
	}
	//this.NotifyCards(seat)
	logger.Logger.Trace("--> Send Player NotifyCards", seat.hands[0])
}

func (this *BlackJackSceneData) bankerSwap(k int) {
	hand := this.chairIDs[0].hands[0]
	hand.handCards[1], this.pokers[k] = this.pokers[k], hand.handCards[1]
	handUpdate(this.chairIDs[0])
}

// 通知庄家牌型信息
// p == nil 广播, 否则发给执行玩家
func (this *BlackJackSceneData) NotifyCards(p *BlackJackPlayerData) {
	hand := this.chairIDs[0].hands[0]
	banker := &blackjack.SCBlackJackNotifyCards{
		Cards: &blackjack.BlackJackCards{
			Id:     proto.Int32(0),
			Cards:  rule.CardsToInt32(hand.handCards),
			DCards: proto.Int32(int32(hand.mulCards.Value())),
			Type:   proto.Int32(hand.tp),
			Point:  hand.point,
			State:  proto.Int32(hand.state),
			Seat:   proto.Int32(0),
		},
		Num: proto.Int32(int32(len(this.pokers))),
	}
	proto.SetDefaults(banker)
	if p != nil {
		p.SendToClient(int(blackjack.BlackJackPacketID_SC_NOTIFY_CARDS), banker)
		return
	}
	this.Broadcast(int(blackjack.BlackJackPacketID_SC_NOTIFY_CARDS), banker, 0)
	logger.Logger.Trace("--> banker SCBlackJackNotifyCards", banker.String())
}

// 通知买保险
// 返回值，false 没有需要买保险的玩家了， true 通知成功
func (this *BlackJackSceneData) NotifyBuy() bool {
	for {
		this.pos++
		if this.pos > rule.MaxPlayer {
			return false
		}
		seat := this.chairIDs[this.pos]
		// 跳过空位
		if !seat.isBet {
			continue
		}
		// 通知
		pack := &blackjack.SCBlackJackNotifyBuy{
			Pos: proto.Int32(int32(seat.seat)),
		}
		this.Broadcast(int(blackjack.BlackJackPacketID_SC_NOTIFY_BUY), pack, 0)
		logger.Logger.Trace("--> Broadcast SCBlackJackNotifyBuy ", pack.String())
		this.StateStartTime = time.Now()
		return true
	}
}

func (this *BlackJackSceneData) HasBuy() bool {
	for _, v := range this.chairIDs[1:] {
		if v.Player == nil || !v.IsGameing() {
			continue
		}
		if v.isBuy {
			return true
		}
	}
	return false
}

func (this *BlackJackSceneData) NotifyPlayer() bool {
	//this.n = 0
	//this.pos = 1
	//isDelay := false
	for {
		//this.pos++
		if this.pos > rule.MaxPlayer {
			return false
		}
		seat := this.chairIDs[this.pos]
		// 跳过空位
		if !seat.isBet {
			this.pos++
			continue
		}
		// 是否已停牌
		if seat.IsStopAction() {
			//if seat.hands[0].state == 1 {
			// 广播停牌
			//this.NotifySkip(seat)	//因为不论是爆牌还是主动停牌，操作结束后都会广播停牌，所以通知下一个玩家操作时不需要再广播停牌
			this.pos++
			//isDelay = true
			continue
		}
		//if isDelay {
		if this.hTimerOpDelay != timer.TimerHandle(0) {
			timer.StopTimer(this.hTimerOpDelay)
			this.hTimerOpDelay = timer.TimerHandle(0)
		}
		hTimerOpDelay, ok := common.DelayInvake(func() {
			// 通知玩家操作
			pack := &blackjack.SCBlackJackNotifyOperate{
				Pos:     proto.Int32(int32(seat.seat)),
				LastPos: proto.Int32(int32(this.lastOperaPos)),
			}
			this.notifyOperate(seat, pack)
			logger.Logger.Trace("-->Delay NotifyOperate SCBlackJackNotifyOperate ", pack.String())
			this.lastOperaPos = seat.seat
			this.StateStartTime = time.Now()
		}, nil, time.Duration(rule.TimeoutDelayOp), int(1))
		if ok {
			this.hTimerOpDelay = hTimerOpDelay
			this.StateStartTime = time.Now() //因为玩家延迟操作可能导致这个阶段超时，所以先重置一下时间
		}
		//} else {
		//	// 通知玩家操作
		//	pack := &blackjack.SCBlackJackNotifyOperate{
		//		Pos: proto.Int32(int32(seat.seat)),
		//	}
		//	this.notifyOperate(seat, pack)
		//	logger.Logger.Trace("--> NotifyOperate SCBlackJackNotifyOperate ", pack.String())
		//
		//	this.stateStartTime = time.Now()
		//}
		return true
	}
}

func (this *BlackJackSceneData) GetAICardsAndSeat(seat *BlackJackPlayerData) (string, int32) {
	buf := bytes.NewBuffer([]byte{})
	for _, v := range this.chairIDs[1:] {
		if v.isBet {
			for _, hand := range v.hands {
				if len(hand.handCards) > 0 {
					buf.WriteString(";")
				}
			}
		}
	}
	// ai座位号
	var seatId int = 1
	for _, v := range this.chairIDs[1:] {
		if !v.isBet {
			continue
		}
		if v.GetSnId() == seat.GetSnId() {
			break
		}
		seatId++
		if len(v.hands[1].handCards) > 0 {
			seatId++
		}
	}
	return buf.String(), int32(seatId)
}

// 广播停牌
// n 取值如下:
// 40 左侧停牌
// 41 爆牌停牌
func (this *BlackJackSceneData) NotifySkip(p *BlackJackPlayerData, n ...int64) {
	// 广播停牌
	pack := &blackjack.SCBlackJackPlayerOperate{
		Code:    blackjack.SCBlackJackPlayerOperate_Success,
		Operate: proto.Int32(rule.SubSkip),
		Pos:     proto.Int32(int32(this.chairIDs[this.pos].seat)),
		Num:     proto.Int32(int32(len(this.pokers))),
	}

	pExtra, _ := p.GetExtraData().(*BlackJackPlayerData)
	pExtra.opCodeFlag = int32(rule.SubSkip)
	if len(n) > 0 && n[0] > 0 {
		pExtra.opCodeFlag = int32(n[0])
		pack.Operate = proto.Int32(int32(n[0]))
	}
	this.Broadcast(int(blackjack.BlackJackPacketID_SC_PLAYER_OPERATE), pack, 0)
	logger.Logger.Trace("--> Skip SCBlackJackPlayerOperate ", pack.String())
}

func (this *BlackJackSceneData) calculate(banker, v *BlackJackPlayerData) {
	if v == nil || banker == nil || v.Player == nil || v.Billed {
		return
	}
	if !v.isBet {
		return
	}
	times := 0
	for k := 0; k < len(v.hands); k++ {
		if len(v.hands[k].handCards) == 0 {
			continue
		}
		times++
		n, err := rule.CompareCards(banker.hands[0].handCards, v.hands[k].handCards)
		if err != nil {
			logger.Logger.Errorf("BlackJack CompareCards Error banker:%+v player:%+v %s", *banker, *v, err.Error())
			return
		}
		if n > 0 || v.hands[k].tp == rule.CardTypeBoom {
			// 输钱
			v.hands[k].betChange = -v.hands[k].betCoin
			// 去掉已经结算的
			if v.hands[k].tp == rule.CardTypeBoom {
				times--
			} else {
				if this.winResult != rule.ResultWin {
					this.winResult = rule.ResultLost
				} else {
					this.winResult = rule.ResultWinAndLost
				}
			}
		} else if n < 0 {
			// 赢钱
			var gain float64
			var winScore int64
			if v.hands[k].mulCards.Value() == 100 {
				// 没有双倍
				winScore = v.hands[k].betCoin
			} else {
				// 双倍
				winScore = v.hands[k].betCoin * 2
			}
			switch v.hands[k].tp {
			case rule.CardTypeA10:
				gain = float64(winScore) * rule.A10Rate
			case rule.CardTypeFive:
				gain = float64(winScore) * rule.FiveRate
			case rule.CardTypeOther:
				gain = float64(winScore) * rule.OtherRate
			}
			v.hands[k].revenue = int64(math.Ceil(gain * float64(this.DbGameFree.GetTaxRate()) / 10000))
			v.hands[k].betChange = int64(gain) - v.hands[k].revenue

			if this.winResult != rule.ResultLost {
				this.winResult = rule.ResultWin
			} else {
				this.winResult = rule.ResultWinAndLost
			}
		}
	}
	if this.endType == 0 {
		if times > 1 {
			times = 1
		}
	}
	this.n += times
}

func (this *BlackJackSceneData) result(v *BlackJackPlayerData, k int) {
	if v == nil || v.Player == nil || !v.isBet || v.Billed {
		return
	}
	v.GameTimes++
	// 总输赢分
	all := v.BetChange() + v.baoChange
	// 总税收
	allRevenue := this.Revenue(v)
	total := all + allRevenue
	curIsWin := 0
	if all > 0 {
		v.WinTimes++
		curIsWin = 1
		v.AddServiceFee(allRevenue)
		v.AddCoin(all, common.GainWay_CoinSceneWin, 0, "system", this.GetSceneName())
	} else if all < 0 {
		v.SetLostTimes(v.GetLostTimes() + 1)
		curIsWin = -1
		v.AddCoin(all, common.GainWay_CoinSceneLost, 0, "system", this.GetSceneName())
	} else {
		curIsWin = 0
	}
	// 都是机器人不记录牌局记录
	if this.gameNum == this.robotNum {
		return
	}

	// 水池变动
	if !this.GetTesting() && !v.IsRobot() {
		//水池变动
		if all < 0 {
			base.CoinPoolMgr.PushCoin(int32(this.SceneType), this.GetGroupId(), this.GetPlatform(), -total)
		} else if all > 0 {
			base.CoinPoolMgr.PopCoin(int32(this.SceneType), this.GetGroupId(), this.GetPlatform(), total)
		}
		this.SetSystemCoinOut(-total)
	}
	// 统一按税前算
	v.Statics(this.GetKeyGameId(), this.KeyGamefreeId, total, true)

	if v.GetCurrentCoin() == 0 {
		v.SetCurrentCoin(v.GetTakeCoin())
	}
	v.SaveSceneCoinLog(v.GetTakeCoin(), all, v.GetCoin(), 0, allRevenue, total, 0, 0)
	v.SetCurrentCoin(v.GetCoin())

	if !v.IsRobot() {
		var totalIn, totalOut int64
		if curIsWin > 0 {
			totalOut = total
		} else {
			totalIn -= all
		}
		validFlow := totalIn + totalOut
		validBet := common.AbsI64(totalIn - totalOut)
		this.SaveGamePlayerListLog(v.GetSnId(),
			base.GetSaveGamePlayerListLogParam(v.Platform, v.Channel, v.BeUnderAgentCode, v.PackageID, this.logicId, v.InviterId, totalIn, totalOut,
				allRevenue, 0, v.BetCoin()+v.baoCoin, all, validBet, validFlow, this.IsPlayerFirst(this.GetPlayer(v.GetSnId())), false))
	}
	player := &model.BlackJackPlayer{
		UserId:        v.SnId,
		UserIcon:      v.Head,
		GainCoinNoTax: all,
		IsWin:         int32(curIsWin),
		IsRob:         v.IsRobot(),
		Flag:          v.GetFlag(),
		Platform:      v.Platform,
		Channel:       v.Channel,
		Promoter:      strconv.Itoa(int(v.PromoterTree)),
		PackageTag:    v.PackageID,
		InviterId:     v.InviterId,
		WBLevel:       v.WBLevel,
		IsFirst:       this.IsPlayerFirst(this.GetPlayer(v.SnId)),
		Tax:           this.Revenue(v),
		Hands:         v.Hands(),
		BaoCoin:       v.baoCoin,
		BaoChange:     v.baoChange,
		BetCoin:       v.BetCoin(),
		BetChange:     v.BetChange(),
		Seat:          k,
	}
	this.playerData = append(this.playerData, player)
}

// CalculateResult 下注结算
func (this *BlackJackSceneData) CalculateResult() {
	/*
		banker的Player是nil, banker是系统庄家, banker = this.chairIDs[0]
	*/
	//水池上下文环境
	this.CpCtx = base.CoinPoolMgr.GetCoinPoolCtx(this.GetPlatform(), this.GetGameFreeId(), this.GetGroupId())
	banker := this.chairIDs[0]
	this.n = 0
	for k, v := range this.chairIDs[1:] {
		this.calculate(banker, v)
		this.result(v, k+1)
	}

	// 都是机器人不记录牌局记录
	if this.gameNum == this.robotNum {
		return
	}

	// 统计牌局详细记录
	if this.gameNum > this.robotNum || this.IsRobFightGame() { //排除掉纯机器人的牌局
		var totalBetCoin, totalBetChange int64
		bjType := model.BlackJackType{
			PlayerCount:     this.gamingNum,
			RoomId:          int32(this.GetSceneId()),
			RoomType:        int32(this.SceneType),
			NumOfGames:      this.NumOfGames,
			BankerCards:     rule.CardsToInt32(banker.hands[0].handCards),
			BankerCardType:  banker.hands[0].tp,
			BankerCardPoint: banker.hands[0].point,
			PlayerData:      this.playerData,
		}
		for k, v := range this.chairIDs[1:] {
			if !v.isBet && this.backup[k] != nil && this.backup[k].isBet {
				v = this.backup[k]
			}
			if !v.isBet {
				continue
			}
			totalBetCoin += v.BetCoin() + v.baoCoin
			totalBetChange += v.BetChange() + v.baoChange + this.Revenue(v)
		}
		bjType.BetCoin = totalBetCoin
		bjType.GainCoinTax = totalBetChange
		if info, err := model.MarshalGameNoteByFIGHT(&bjType); err == nil {
			this.SaveGameDetailedLog(this.logicId, info, &base.GameDetailedParam{})
		}
	}
}

func (this *BlackJackSceneData) End() *blackjack.SCBlackJackEnd {
	if this.endType != 0 {
		return this.End2()
	}
	pack := &blackjack.SCBlackJackEnd{}
	for _, v := range this.chairIDs {
		if v.isBanker || !v.isBet || !v.IsGameing() {
			continue
		}
		seat := &blackjack.BlackJackPlayerEnd{
			Pos:       proto.Int32(int32(v.seat)),
			Gain:      proto.Int64(v.BetChange()),
			LeftGain:  proto.Int64(v.LeftBetChange()),
			RightGain: proto.Int64(v.RightBetChange()),
			Coin:      proto.Int64(v.GetCurrentCoin()),
		}
		pack.Players = append(pack.Players, seat)
	}
	proto.SetDefaults(pack)
	logger.Logger.Trace("--> End SCBlackJackEnd ", pack.String())
	return pack
}

func (this *BlackJackSceneData) End2() *blackjack.SCBlackJackEnd {
	pack := &blackjack.SCBlackJackEnd{}
	for _, v := range this.chairIDs {
		if v.Player == nil || !v.isBet || !v.IsGameing() {
			continue
		}
		for k, hand := range v.hands {
			if len(hand.handCards) > 0 && hand.tp != rule.CardTypeBoom {
				seat := &blackjack.BlackJackPlayerEnd{
					Pos:       proto.Int32(int32(v.seat)),
					Coin:      proto.Int64(v.GetCurrentCoin()),
					Gain:      proto.Int64(hand.betChange),
					LeftGain:  proto.Int64(v.LeftBetChange()),
					RightGain: proto.Int64(v.RightBetChange()),
					IsDouble:  proto.Bool(true),
				}
				if hand.mulCards != nil && hand.mulCards.Value() == rule.NewCardDefault().Value() {
					seat.IsDouble = proto.Bool(false)
				}
				if k == 0 {
					if v.hands[1].betChange > 0 {
						seat.Coin = proto.Int64(v.GetCurrentCoin() - v.hands[1].betChange)
					} else if v.hands[1].betChange == 0 {
						seat.Coin = proto.Int64(v.GetCurrentCoin() - v.hands[1].betCoin)
					}
				}
				pack.Players = append(pack.Players, seat)
			}
		}
	}
	proto.SetDefaults(pack)
	logger.Logger.Trace("--> End SCBlackJackEnd ", pack.String())
	return pack
}

// RobotLeave 机器人离开策略
func (t *BlackJackSceneData) RobotLeave(s *base.Scene) bool {
	hasLeave := false
	if !t.GetTesting() {
		// 金币数量在后台设定区间内，机器人离开
		for _, v := range t.chairIDs[1:] {
			if v.Player != nil && v.IsRobot() {
				if len(t.DbGameFree.GetRobotLimitCoin()) > 1 {
					if v.GetCoin() > int64(t.DbGameFree.GetRobotLimitCoin()[0]) &&
						v.GetCoin() < int64(t.DbGameFree.GetRobotLimitCoin()[1]) {
						s.PlayerLeave(v.Player, common.PlayerLeaveReason_DropLine, false)
						hasLeave = true
					}
				}
			}
		}
	}
	// 检查机器人离开标记
	for _, v := range t.chairIDs[1:] {
		if v.Player != nil && v.IsRobot() && v.leaveNum > 0 {
			v.leaveNum--
			if v.leaveNum == 0 {
				if t.GetRealPlayerCnt() >= 3 {
					s.PlayerLeave(v.Player, common.PlayerLeaveReason_DropLine, false)
					hasLeave = true
				} else {
					var n int
					for _, v := range t.chairIDs {
						if v.Player != nil && v.IsRobot() {
							n++
						}
					}
					if n > 1 {
						s.PlayerLeave(v.Player, common.PlayerLeaveReason_DropLine, false)
						hasLeave = true
					}
				}
			}
		}
	}
	// 有3个真人，标记所有机器人2-4局后离开
	if t.GetRealPlayerCnt() >= 3 {
		for _, v := range t.chairIDs[1:] {
			if v.Player != nil && v.IsRobot() && v.leaveNum == 0 {
				v.leaveNum = t.RandInt(2, 5)
			}
		}
	}
	// 有新增玩家，标记一个机器人离开
	realNum := t.gameNum - t.robotNum
	if realNum >= 1 {
		if n := t.GetRealPlayerCnt() - realNum; n > 0 {
			for i := 0; i < n; i++ {
				for _, v := range t.chairIDs[1:] {
					if v.Player != nil && v.IsRobot() && v.leaveNum == 0 {
						v.leaveNum = t.RandInt(2, 5)
						break
					}
				}
			}
		}
	}
	return hasLeave
}

func (this *BlackJackSceneData) BetMinMax() []int64 {
	return []int64{int64(this.DbGameFree.GetBaseScore()), int64(this.DbGameFree.GetMaxChip())}
}

// 庄家牌型已经最大则停牌
func (this *BlackJackSceneData) MaxHand() {
	var hands []*rule.Card
	for _, v := range this.chairIDs[1:] {
		if !v.isBet {
			continue
		}
		for _, hand := range v.hands {
			if len(hand.handCards) == 0 {
				continue
			}
			if len(hands) == 0 {
				hands = hand.handCards
			} else if n, _ := rule.CompareCards(hand.handCards, hands); n > 0 {
				hands = hand.handCards
			}
		}
	}
	this.maxHand = hands
}

func (this *BlackJackSceneData) AllBilled() bool {
	flag := true
	for _, v := range this.chairIDs[1:] {
		if v.Player != nil && v.isBet && !v.Billed {
			flag = false
			break
		}
	}
	return flag
}

func (this *BlackJackSceneData) BankerStop() bool {
	if len(this.maxHand) == 0 {
		return false
	}
	n, _ := rule.CompareCards(this.chairIDs[0].hands[0].handCards, this.maxHand)
	return n > 0
}

func (this *BlackJackSceneData) BetState() {
	this.GameNowTime = time.Now()
	this.NumOfGames++
	this.NotifySceneRoundStart(this.NumOfGames)
	// 重置场景数据
	this.Release()
	// 统计玩家数量和机器人数量
	this.InitRobotNumGameNum()
	// 服务费
	if fee := this.GetCoinSceneServiceFee(); fee > 0 && !this.GetTesting() {
		for _, v := range this.chairIDs[1:] {
			if v.Player != nil && v.IsGameing() {
				v.AddServiceFee(int64(fee))
				v.AddCoin(-int64(fee), common.GainWay_ServiceFee, base.SyncFlag_ToClient, "system", this.GetSceneName())
			}
		}
	}
	//同步防伙牌数据
	this.SyncScenePlayer()
}

func (this *BlackJackSceneData) notifyOperate(p *BlackJackPlayerData, pack *blackjack.SCBlackJackNotifyOperate) {
	if p.IsRobot() {
		str, seatId := this.GetAICardsAndSeat(p)
		pack.Cards = proto.String(str)
		pack.Seat = proto.Int32(seatId)
	}
	p.SendToClient(int(blackjack.BlackJackPacketID_SC_NOTIFY_OPERATE), pack)
	this.Broadcast(int(blackjack.BlackJackPacketID_SC_NOTIFY_OPERATE), pack, p.GetSid())
}

func (this *BlackJackSceneData) playerOperate(p *BlackJackPlayerData, pack *blackjack.SCBlackJackPlayerOperate) {
	if p.IsRobot() {
		str, seatId := this.GetAICardsAndSeat(p)
		pack.CardsStr = proto.String(str)
		pack.Seat = proto.Int32(seatId)
	}
	p.SendToClient(int(blackjack.BlackJackPacketID_SC_PLAYER_OPERATE), pack)
	this.Broadcast(int(blackjack.BlackJackPacketID_SC_PLAYER_OPERATE), pack, p.GetSid())
}
