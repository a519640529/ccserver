package roulette

import (
	"encoding/json"
	"games.yol.com/win88/gamerule/roulette"
	rule "games.yol.com/win88/gamerule/roulette"
	"games.yol.com/win88/gamesrv/base"
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	proto_roulette "games.yol.com/win88/protocol/roulette"
	proto_server "games.yol.com/win88/protocol/server"
	"github.com/idealeak/goserver/core/logger"
	"math/rand"
	"sort"
	"time"
)

//Roulette
type RouletteSceneData struct {
	*base.Scene                                //房间信息
	players      map[int32]*RoulettePlayerData //玩家信息
	seats        []*RoulettePlayerData         //座位信息(玩家列表 前20名和神算子)
	winDigit     int                           //中奖数字
	winRecord    []int32                       //历史记录
	richTop      []*RoulettePlayerData         //富豪no.1 神算子 富豪no.2
	point        *roulette.PointType           //数字位置集合
	betCnt       map[int][]int                 //所有人的下注筹码除了座上玩家 发送筹码用
	winCoin      int64                         //玩家赢的钱 不算座上玩家
	tempTime     time.Time                     //上次发送筹码的时间
	realBetCoin  map[int]int64                 //每个区域下注额 --统计使用
	realWinCoin  map[int]int64                 //每个区域赢的金额 --统计使用
	pointMapNums map[int]bool                  //赢的区域编号 0-156
	bankerWin    int64                         //系统总输赢
}

func NewRouletteSceneData(s *base.Scene) *RouletteSceneData {
	return &RouletteSceneData{
		Scene:   s,
		players: make(map[int32]*RoulettePlayerData),
	}
}
func (this *RouletteSceneData) init() bool {
	this.point = new(roulette.PointType)
	this.point.Init()
	return true
}
func (this *RouletteSceneData) SceneDestroy(force bool) {
	//销毁房间
	this.Scene.Destroy(force)
}
func (this *RouletteSceneData) CanStart() bool {
	if len(this.players) >= 0 {
		return true
	}
	return false
}
func (this *RouletteSceneData) HasRealPlayer() bool {
	for _, p := range this.players {
		if p != nil && !p.IsRob {
			return true
		}
	}
	return false
}
func (this *RouletteSceneData) Clean() {
	for _, playerEx := range this.players {
		playerEx.Clean()
	}
	this.winDigit = -1
	this.winCoin = 0
	this.bankerWin = 0
	this.betCnt = make(map[int][]int)
	this.realBetCoin = make(map[int]int64)
	this.realWinCoin = make(map[int]int64)
	this.pointMapNums = make(map[int]bool)

	if len(this.winRecord) > 20 {
		this.winRecord = this.winRecord[1:]
	}
	//重置水池调控标记
	this.CpControlled = false
}
func (this *RouletteSceneData) delPlayer(p *base.Player) {
	if p, exist := this.players[p.SnId]; exist {
		delete(this.players, p.SnId)
		count := len(this.seats)
		idx := -1
		for i := 0; i < count; i++ {
			if this.seats[i].SnId == p.SnId {
				idx = i
				break
			}
		}
		if idx != -1 {
			temp := this.seats[:idx]
			temp = append(temp, this.seats[idx+1:]...)
			this.seats = temp
		}
	}
}
func (this *RouletteSceneData) SendPlayerList(p *base.Player) {
	player := make([]*proto_roulette.RoulettePlayerData, 0)
	topPlayer := make([]*RoulettePlayerData, 0)
	if len(this.seats) > 21 {
		topPlayer = this.seats[:21]
	} else {
		topPlayer = this.seats
	}
	for _, v := range topPlayer {
		player = append(player, &proto_roulette.RoulettePlayerData{
			Name:            proto.String(v.Name),
			Coin:            proto.Int64(v.Coin),
			Head:            proto.Int32(v.Head),
			HeadOutLine:     proto.Int32(v.HeadOutLine),
			BetCoin:         proto.Int64(v.totalBetCoin),
			Lately20BetCoin: proto.Int64(v.lately20Bet),
			Lately20Win:     proto.Int64(int64(v.lately20Win)),
			NiceId:          proto.Int32(v.NiceId),
		})
	}
	pack := &proto_roulette.SCRoulettePlayerOn{
		Players: player,
	}
	proto.SetDefaults(pack)
	logger.Logger.Trace("SCRoulettePlayerOn: ", pack)
	p.SendToClient(int(proto_roulette.RouletteMmoPacketID_PACKET_SC_Roulette_PLayerOn), pack)
}
func (this *RouletteSceneData) ShowRichTop() {
	this.richTop = make([]*RoulettePlayerData, 0)
	sort.Slice(this.seats, func(i, j int) bool {
		if this.seats[i].lately20Bet > this.seats[j].lately20Bet {
			return true
		}
		return false
	})
	winBig := 0
	winBigSnId := int32(0)
	for _, v := range this.seats {
		if v != nil && v.lately20Win > winBig {
			winBig = v.lately20Win
			winBigSnId = v.SnId
		}
	}
	if winBigSnId != 0 {
		for k, v := range this.seats {
			if v.SnId == winBigSnId {
				this.seats = append(this.seats[:k], this.seats[k+1:]...)
				break
			}
		}
		temp := make([]*RoulettePlayerData, len(this.seats))
		copy(temp, this.seats)
		winBigPlayer := this.players[winBigSnId]
		winBigPlayer.Pos = rule.Roulette_BESTWINPOS
		this.richTop = append(this.richTop, winBigPlayer)
		this.seats = []*RoulettePlayerData{winBigPlayer}
		this.seats = append(this.seats[:], temp...)
	}
	if len(this.seats) > 1 || (len(this.seats) > 1 && winBigSnId != 0) {
		idx := 0
		if winBigSnId != 0 {
			idx++
		}
		rich1 := this.seats[idx]
		rich1.Pos = rule.Roulette_RICHTOP1
		this.richTop = append(this.richTop, rich1)
		if len(this.seats) > idx+1 {
			rich2 := this.seats[idx+1]
			rich2.Pos = rule.Roulette_RICHTOP2
			this.richTop = append(this.richTop, rich2)
		}
	} else if len(this.seats) == 1 && winBigSnId == 0 {
		rich1 := this.seats[0]
		rich1.Pos = rule.Roulette_RICHTOP1
		this.richTop = append(this.richTop, rich1)
	}
	this.SendTopThree()
}

//计算所有结果 找出需要的结果并随机取一个值
func (this *RouletteSceneData) GetNeedComputeRandScore(bankerWinTag int) int {
	allBankerWinCoin := make(map[int]int64)
	for i := 0; i <= 36; i++ {
		bankerWinCoin := int64(0)
		pointMapNums := this.point.PointMapNums[i]
		for _, p := range this.players {
			if p.IsGameing() && !p.IsRob {
				for betPos, coin := range p.betCoin {
					bankerWinCoin += coin
					if _, ok := pointMapNums[betPos]; ok {
						betType := this.point.GetBetType(betPos)
						rate := this.point.RateMap[betType]
						bankerWinCoin -= coin * int64(rate+1)
						pointMapNums[betPos] = false
					}
				}
			}
		}
		allBankerWinCoin[i] = bankerWinCoin
	}
	winPos := []int{}
	minCoin := int64(999999999)
	for key, bankerWin := range allBankerWinCoin {
		switch bankerWinTag {
		case 1:
			//庄家赢
			if bankerWin > 0 {
				winPos = append(winPos, key)
			}
		case 2:
			//庄家输
			if bankerWin < 0 {
				if base.CoinPoolMgr.IsMaxOutHaveEnough(this.Platform, this.GetGameFreeId(), this.GroupId, bankerWin) {
					winPos = append(winPos, key)
				}
			}
		case 3:
			//庄家要输的最惨
			if bankerWin < 0 {
				if base.CoinPoolMgr.IsMaxOutHaveEnough(this.Platform, this.GetGameFreeId(), this.GroupId, bankerWin) {
					if bankerWin < minCoin {
						winPos = []int{key}
						minCoin = bankerWin
					}
				}
			}
		}
	}
	if len(winPos) == 0 {
		return rand.Intn(37)
	}
	return winPos[rand.Intn(len(winPos))]
}

//随机一种结果
func (this *RouletteSceneData) ComputeScore() (bankerWinCoin int64) {
	this.winDigit = rand.Intn(37)
	pointMapNums := this.point.PointMapNums[this.winDigit]
	for _, p := range this.players {
		if p.IsGameing() && !p.IsRob {
			for betPos, coin := range p.betCoin {
				bankerWinCoin += coin
				if _, ok := pointMapNums[betPos]; ok {
					betType := this.point.GetBetType(betPos)
					rate := this.point.RateMap[betType]
					bankerWinCoin -= coin * int64(rate+1)
					pointMapNums[betPos] = false
				}
			}
		}
	}
	return
}

//根据水池状态判定当局结果
func (this *RouletteSceneData) ComputePoolState() {
	//先随机一种结果来计算
	bankerWinCoin := this.ComputeScore()
	state, _ := base.CoinPoolMgr.GetCoinPoolStatus2(this.Platform, this.GetGameFreeId(), this.GroupId, bankerWinCoin)
	switch state {
	case base.CoinPoolStatus_Normal:
	case base.CoinPoolStatus_Low:
		//亏钱状态
		if bankerWinCoin < 0 {
			//庄家要赢
			this.winDigit = this.GetNeedComputeRandScore(1)
			this.CpControlled = true
		}
	case base.CoinPoolStatus_High:
		//赢钱模式
		if bankerWinCoin > 0 {
			//庄家要输 吐分
			this.winDigit = this.GetNeedComputeRandScore(2)
			this.CpControlled = true
		}
	case base.CoinPoolStatus_TooHigh:
		//大赢钱状态
		if bankerWinCoin > 0 {
			//庄家要输 吐分 还要输最惨
			this.winDigit = this.GetNeedComputeRandScore(3)
			this.CpControlled = true
		}
	}
}

// 直接结算
func (this *RouletteSceneData) BilledScore() {
	//中奖的位置号码集合 0-157
	this.pointMapNums = this.point.PointMapNums[this.winDigit]
	bankerTax := int64(0)
	for _, p := range this.players {
		if p.IsGameing() {
			win := -1 //玩家是否押中
			for betPos, coin := range p.betCoin {
				if coin > 0 {
					p.gainCoin -= coin
					if !p.IsRob {
						this.realWinCoin[betPos] -= coin
					}
					p.winCoinRecord[betPos] -= coin
					if _, ok := this.pointMapNums[betPos]; ok {
						win = 1 //押中一次就算
						betType := this.point.GetBetType(betPos)
						rate := this.point.RateMap[betType]
						tempCoin := coin * int64(rate)
						tax := tempCoin * int64(this.DbGameFree.GetTaxRate()) / 10000
						p.winCoin += tempCoin + coin
						p.gainCoin += tempCoin - tax + coin
						p.taxCoin += tax
						this.pointMapNums[betPos] = false
						if p.Pos == rule.Roulette_OLPOS {
							this.winCoin += tempCoin + coin
						}
						if !p.IsRob {
							this.realWinCoin[betPos] += tempCoin + coin - tax
						}
						p.winCoinRecord[betPos] += tempCoin + coin
					}
				}
			}
			p.totalWinRecord = append(p.totalWinRecord, win)
			p.totalBetRecord = append(p.totalBetRecord, p.totalBetCoin)
			if !p.IsRob {
				this.bankerWin -= p.gainCoin + p.taxCoin
				bankerTax += p.taxCoin
				//统计税收
				p.AddServiceFee(p.taxCoin)
			}
		} else if p.gameRouletteTimes > 0 {
			p.totalWinRecord = append(p.totalWinRecord, -1)
			p.totalBetRecord = append(p.totalBetRecord, 0)
		}
	}
	//水池变动
	if this.bankerWin > 0 {
		base.CoinPoolMgr.PushCoin(this.GetGameFreeId(), this.GroupId, this.Platform, this.bankerWin)
	} else if this.bankerWin < 0 {
		base.CoinPoolMgr.PopCoin(this.GetGameFreeId(), this.GroupId, this.Platform, -this.bankerWin)
	}
	//减去税收做统计用
	this.bankerWin += bankerTax
	//开奖记录变动
	this.winRecord = append(this.winRecord, int32(this.winDigit))
}

func (this *RouletteSceneData) SendTopThree() {
	pack := &proto_roulette.SCRouletteTopPlayer{}
	for _, p := range this.richTop {
		if p != nil {
			pack.Players = append(pack.Players, &proto_roulette.RoulettePlayerData{
				SnId:            proto.Int32(p.SnId),
				Name:            proto.String(p.Name),
				Coin:            proto.Int64(p.Coin),
				Head:            proto.Int32(p.Head),
				HeadOutLine:     proto.Int32(p.HeadOutLine),
				BetCoin:         proto.Int64(p.totalBetCoin),
				Lately20BetCoin: proto.Int64(p.lately20Bet),
				Lately20Win:     proto.Int64(int64(p.lately20Win)),
				Pos:             proto.Int32(int32(p.Pos)),
				NiceId:          proto.Int32(p.NiceId),
			})
		}
	}
	this.Broadcast(int(proto_roulette.RouletteMmoPacketID_PACKET_SC_Roulette_TopPlayer), pack, 0)
}

//上座玩家下注通知所有人  自己下注只通知自己
func (this *RouletteSceneData) SendToClickBetInfo(pack *proto_roulette.SCRoulettePlayerOp, player *RoulettePlayerData) {
	//下注失败 只通知自己
	if pack.GetOpRCode() != rule.RoulettePlayerOpSuccess {
		pack.Pos = proto.Int32(int32(rule.Roulette_SELFPOS))
		proto.SetDefaults(pack)
		player.SendToClient(int(proto_roulette.RouletteMmoPacketID_PACKET_SC_Roulette_PlayerOp), pack)
		return
	}
	//上座玩家下注 通知所有人包括自己
	if player.Pos <= rule.Roulette_BESTWINPOS && (pack.GetOpCode() == int32(rule.RoulettePlayerOpBet) ||
		pack.GetOpCode() == int32(rule.RoulettePlayerOpProceedBet)) {
		this.Broadcast(int(proto_roulette.RouletteMmoPacketID_PACKET_SC_Roulette_PlayerOp), pack, player.GetSid())

		newPack := &proto_roulette.SCRoulettePlayerOp{
			OpRCode:        pack.OpRCode,
			OpCode:         pack.OpCode,
			Pos:            proto.Int32(int32(rule.Roulette_SELFPOS)),
			ProceedBetCoin: pack.ProceedBetCoin,
		}
		player.SendToClient(int(proto_roulette.RouletteMmoPacketID_PACKET_SC_Roulette_PlayerOp), newPack)
	} else {
		//自己下注只通知自己
		pack.Pos = proto.Int32(int32(rule.Roulette_SELFPOS))
		player.SendToClient(int(proto_roulette.RouletteMmoPacketID_PACKET_SC_Roulette_PlayerOp), pack)
	}

}
func (this *RouletteSceneData) SaveLog(logid string) {
	this.PlayerWinScoreStatics()

	betCoin := int64(0)
	rouletteType := new(model.RouletteType)
	for betKey, coin := range this.realBetCoin {
		if coin > 0 {
			isWin := -1
			betCoin += coin
			if _, ok := this.pointMapNums[betKey]; ok {
				isWin = 1
			}
			rr := model.RouletteRegion{
				Id:      betKey,
				IsWin:   isWin,
				BetCoin: coin,
				WinCoin: this.realWinCoin[betKey],
				Player:  []model.RoulettePlayer{},
			}
			rouletteType.RouletteRegion = append(rouletteType.RouletteRegion, rr)
		}
	}

	rouletteType.BankerInfo = model.RouletteBanker{
		Point:        this.winDigit,
		TotalBetCoin: betCoin,
		TotalWinCoin: this.bankerWin,
	}

	for _, p := range this.players {
		if !p.IsRob && p.totalBetCoin > 0 {
			rouletteType.Person = append(rouletteType.Person, model.RoulettePerson{
				UserId:       p.SnId,
				BeforeCoin:   p.Coin - p.gainCoin,
				AfterCoin:    p.Coin,
				UserBetTotal: p.totalBetCoin,
				UserWinCoin:  p.gainCoin,
				BetCoin:      p.betCoin,
				IsRob:        p.IsRob,
				WBLevel:      p.WBLevel,
			})
			for k, rr := range rouletteType.RouletteRegion {
				if coin, ok := p.betCoin[rr.Id]; ok {
					rouletteType.RouletteRegion[k].Player = append(rouletteType.RouletteRegion[k].Player, model.RoulettePlayer{
						UserId:  p.SnId,
						BetCoin: coin})
				}
			}
		}
	}
	info, err := model.MarshalGameNoteByHUNDRED(&rouletteType)
	if err == nil {
		winRecord, _ := json.Marshal(this.winRecord)
		this.SaveGameDetailedLog(logid, info, &base.GameDetailedParam{
			Trend20Lately: string(winRecord),
		})
	}

}

//结算之后统计所有人
func (this *RouletteSceneData) PlayerWinScoreStatics() {
	if this.Testing {
		return
	}
	pws := make([]*proto_server.PlayerWinScore, 0)
	realGain := int64(0)
	realNum := 0
	for _, p := range this.players {
		if !p.IsRob && p.totalBetCoin > 0 {
			pws = append(pws, &proto_server.PlayerWinScore{
				SnId:     proto.Int32(p.SnId),
				WinScore: proto.Int64(p.gainCoin),             //玩家输赢额,税后
				Gain:     proto.Int64(p.gainCoin + p.taxCoin), //输赢额,税前
				Tax:      proto.Int64(p.taxCoin),
			})
			realNum++
			realGain += p.gainCoin + p.taxCoin
		}
	}
	if realNum <= 0 {
		return
	}
	pack := &proto_server.GWPlayerWinScore{
		GameFreeId:      proto.Int32(this.GetGameFreeId()),
		PlayerWinScores: pws,
		RobotGain:       proto.Int64(-realGain),
	}
	proto.SetDefaults(pack)
	this.SendToWorld(int(proto_server.SSPacketID_PACKET_GW_PLAYERWINSOCORE), pack)
}
