package hundredyxx

import (
	"fmt"
	"games.yol.com/win88/common"
	"games.yol.com/win88/gamerule/hundredyxx"
	"games.yol.com/win88/gamesrv/base"
	"games.yol.com/win88/model"
	"github.com/idealeak/goserver/core/timer"
	"math/rand"
	"sort"
)

/*
每局都走智能化运营，但是在下注完成后检查是否需要单控
因为是否单控在下注完成后才能确定，如果需要走单控，本局智能化运营算未启用
*/

//const (
//	HundredYXX_ZONE_BANKER int = iota
//	HundredYXX_ZONE_HEI
//	HundredYXX_ZONE_HONG
//	HundredYXX_ZONE_MEI
//	HundredYXX_ZONE_FANG
//	HundredYXX_ZONE_MAX
//)

type RollPointSort struct {
	systemCoinOut int64
	point         [3]int32
}

type HundredYXXSceneData struct {
	*base.Scene //房间信息
	//poker              *hundredyxx.Poker                               //牌
	blackBox           *hundredyxx.DiceBox                              //骰子
	betFiledInfo       [hundredyxx.BetField_MAX]*HundredYXXBetFieldData //记录庄、黑、红、梅、方，每个位置的牌的信息
	players            map[int32]*HundredYXXPlayerData                  //玩家信息
	upplayerlist       []*HundredYXXPlayerData                          //上庄玩家列表
	seats              []*HundredYXXPlayerData                          //座位信息
	winTop1            *HundredYXXPlayerData                            //神算子
	betTop5            [HYXX_RICHTOP5]*HundredYXXPlayerData             //押注最多的5位玩家
	bankerSnId         int32                                            //当前庄家id
	betInfo            [hundredyxx.BetField_MAX]int64                   //记录下注位置及筹码
	betInfoRob         [hundredyxx.BetField_MAX]int64                   //记录机器人下注位置及筹码
	trends             []int32                                          //最近6局走势
	trend              int32                                            //本局输赢
	by                 func(p, q *HundredYXXPlayerData) bool            //排序函数
	constSeatKey       string                                           //固定座位的唯一码
	hRunRecord         timer.TimerHandle                                //每十分钟记录一次数据
	bankerWinCoin      int64                                            //四个闲家的总金额
	bankerbetCoin      int64                                            //四个闲家的下注
	hRunBetSend        timer.TimerHandle                                //每0.5秒发送一次下注消息
	fakePlayerNum      int                                              //虚假的玩家数量信息
	killAll            int32
	upplayerCount      int32 //当前上庄玩家已在庄次数
	testing            bool
	singleAdjustSnId   int32
	singleAdjustConfig *model.PlayerSingleAdjust
	logId              string
}

func (this *HundredYXXSceneData) LoadData() {
}

func (this *HundredYXXSceneData) SaveData(force bool) {
}

//Len()
func (s *HundredYXXSceneData) Len() int {
	return len(s.seats)
}

//Less():输赢记录将有高到底排序
func (s *HundredYXXSceneData) Less(i, j int) bool {
	return s.by(s.seats[i], s.seats[j])
}

//Swap()
func (s *HundredYXXSceneData) Swap(i, j int) {
	s.seats[i], s.seats[j] = s.seats[j], s.seats[i]
}

func NewHundredYXXSceneData(s *base.Scene) *HundredYXXSceneData {
	return &HundredYXXSceneData{
		Scene:   s,
		players: make(map[int32]*HundredYXXPlayerData),
	}
}

func (this *HundredYXXSceneData) RebindPlayerSnId(oldSnId, newSnId int32) {
	if p, exist := this.players[oldSnId]; exist {
		delete(this.players, oldSnId)
		this.players[newSnId] = p
	}
}

func (this *HundredYXXSceneData) init() bool {
	this.bankerSnId = -1
	this.blackBox = hundredyxx.CreateDiceBox()
	this.logId, _ = model.AutoIncGameLogId()
	for i := 0; i < hundredyxx.BetField_MAX; i++ {
		this.betFiledInfo[i] = &HundredYXXBetFieldData{
			FieldName: hundredyxx.WinKindStr[i],
		}
	}
	this.LoadData()
	return true
}

func (this *HundredYXXSceneData) Clean() {
	this.bankerWinCoin = 0
	for _, p := range this.players {
		p.Clean()
	}

	for i := 0; i < hundredyxx.BetField_MAX; i++ {
		this.betInfo[i] = 0
		this.betInfoRob[i] = 0
		this.betFiledInfo[i].Clean()
	}
	//重置水池调控标记
	this.SetCpControlled(false)
}

func (this *HundredYXXSceneData) CanStart() bool {
	return true
}

func (this *HundredYXXSceneData) delPlayer(p *base.Player) {
	if p, exist := this.players[p.SnId]; exist {
		delete(this.players, p.SnId)

		if len(this.seats) > 0 {
			for i := 0; i < len(this.seats); i++ {
				seat := this.seats[i]
				if seat != nil {
					if seat.SnId == p.SnId {
						this.seats = append(this.seats[:i], this.seats[i+1:]...)
						i--
					}
				}
			}
		}
		if len(this.upplayerlist) > 0 {
			index := 0
			for _, pp := range this.upplayerlist {
				if pp.SnId == p.SnId {
					this.upplayerlist = append(this.upplayerlist[:index], this.upplayerlist[index+1:]...)
					break
				}
				index++
			}
		}
	}
}

func (this *HundredYXXSceneData) OnPlayerLeave(p *base.Player, reason int) {
	this.delPlayer(p)
	this.BroadcastPlayerLeave(p, reason)
}

func (this *HundredYXXSceneData) BroadcastPlayerLeave(p *base.Player, reason int) {
}

func (this *HundredYXXSceneData) SceneDestroy(force bool) {
	//销毁房间
	this.Scene.Destroy(force)
}

func (this *HundredYXXSceneData) IsNeedCare(p *HundredYXXPlayerData) bool {
	if p.IsNeedCare() {
		return true
	}
	return false
}

func (this *HundredYXXSceneData) IsNeedWeaken(p *HundredYXXPlayerData) bool {
	if p.IsNeedWeaken() {
		return true
	}
	return false
}

//最大赢取倒序
func HYXX_OrderByLatest20Win(l, r *HundredYXXPlayerData) bool {
	return l.cGetWin20 > r.cGetWin20
}

//最大押注倒序
func HYXX_OrderByLatest20Bet(l, r *HundredYXXPlayerData) bool {
	return l.cGetBetGig20 > r.cGetBetGig20
}

func (this *HundredYXXSceneData) Resort() string {
	constSeatKey := fmt.Sprintf("%v", this.bankerSnId)
	best := int64(0)
	count := len(this.seats)
	this.winTop1 = nil
	for i := 0; i < count; i++ {
		playerEx := this.seats[i]
		if playerEx != nil && playerEx.SnId != this.bankerSnId {
			playerEx.SetPos(HYXX_OLPOS)
			if playerEx.cGetWin20 > best {
				best = playerEx.cGetWin20
				this.winTop1 = playerEx
			}
		}
	}
	if this.winTop1 != nil {
		this.winTop1.SetPos(HYXX_BESTWINPOS)
		constSeatKey += fmt.Sprintf("%d_%d", this.winTop1.SnId, this.winTop1.GetPos())
	}

	this.by = HYXX_OrderByLatest20Bet
	sort.Sort(this)
	cnt := 0
	for i := 0; i < HYXX_RICHTOP5; i++ {
		this.betTop5[i] = nil
	}
	for i := 0; i < count && cnt < HYXX_RICHTOP5; i++ {
		playerEx := this.seats[i]
		if playerEx != this.winTop1 && playerEx.SnId != this.bankerSnId {
			playerEx.SetPos(cnt + 1)
			this.betTop5[cnt] = playerEx
			cnt++
			constSeatKey += fmt.Sprintf("%d_%d", playerEx.SnId, playerEx.GetPos())
		}
	}

	return constSeatKey
}

func (this *HundredYXXSceneData) CalcuFakePlayerNum() int {
	truePlayerCount := int32(len(this.players))
	this.fakePlayerNum = int(truePlayerCount)
	return this.fakePlayerNum
}
func (this *HundredYXXSceneData) CalcuRate() int {
	//func (this *HundredYXXSceneData) CalcuRate(data *HundredYXXBetFieldData) int {
	//	return hundredyxx.CardKindMultiple[data.cardsO.Kind]
	return 0
}

func (this *HundredYXXSceneData) CalcWin() {
	point := this.blackBox.Point
	this.trend = point[0] | point[1]<<8 | point[2]<<16
	//先统计单图案输赢
	for i := 0; i <= hundredyxx.BetField_Single_6; i++ {
		this.betFiledInfo[i].Clean()
		winPoint := int32(hundredyxx.BetFieldWinPoint[i][0])
		for j := 0; j < hundredyxx.DICE_NUM; j++ {
			if winPoint == this.blackBox.Point[j] {
				this.betFiledInfo[i].winFlag = 1
				this.betFiledInfo[i].gainRate++ //三个骰子中，有一个相应图案中奖倍率为1，有两个中奖倍率为2，有三个中奖倍率为3
			}
		}
	}
	//再统计双图案输赢
	for i := hundredyxx.BetField_Double_7; i < hundredyxx.BetField_MAX; i++ {
		this.betFiledInfo[i].Clean()
		winPoint1 := int32(hundredyxx.BetFieldWinPoint[i][0])
		winPoint2 := int32(hundredyxx.BetFieldWinPoint[i][1])
		if common.InSliceInt32(this.blackBox.Point[:], winPoint1) && common.InSliceInt32(this.blackBox.Point[:], winPoint2) {
			this.betFiledInfo[i].winFlag = 1
			this.betFiledInfo[i].gainRate = 5
		}
	}
}

func (this *HundredYXXSceneData) IsRobotBanker() bool {
	banker := this.players[this.bankerSnId]
	if banker == nil || banker.IsRob {
		return true
	}
	return false
}

func (this *HundredYXXSceneData) RealPlayerBet() int64 {
	var totle int64
	for _, value := range this.players {
		if !value.IsRob {
			totle += value.betTotal
		}
	}
	return totle
}

func (this *HundredYXXSceneData) AdjustCard() {
	if this.testing || this.IsPrivateScene() {
		return
	}
	//if this.RealPlayerBet() == 0 {
	//	//无真人下注，让路单相对比较漂亮点
	//	for i := 0; i < model.GameParamData.HYXXTryAdjustCardTimes; i++ {
	//		adjust := this.IsAllKill()
	//		if adjust {
	//			logger.Logger.Tracef("scene:%v 水池正常且无真人下注，出现[通杀]，尝试调整一次结果", this.sceneId)
	//		} else {
	//			adjust = this.IsFailKill()
	//			if adjust {
	//				logger.Logger.Tracef("scene:%v 水池正常且无真人下注，出现[通赔]，尝试调整一次结果", this.sceneId)
	//			}
	//		}
	//		if !adjust {
	//			break
	//		}
	//		////通杀或者通赔的情况下，再随机一次
	//		//for i := 0; i < hundredyxx.SEND_CARD_NUM; i++ {
	//		//	this.poker.Restore(this.betFiledInfo[i].cards[2:])
	//		//	this.betFiledInfo[i].cardsO = nil
	//		//}
	//		//for i := 0; i < hundredyxx.SEND_CARD_NUM; i++ {
	//		//	for j := 2; j < hundredyxx.HAND_CARD_NUM; j++ { //重新发后三张
	//		//		this.betFiledInfo[i].cards[j] = int(this.poker.Next())
	//		//	}
	//		//}
	//
	//		this.cpControlled = true
	//
	//		allKill := this.IsAllKill()
	//		allFail := this.IsFailKill()
	//		logger.Logger.Tracef("scene:%v 水池正常且无真人下注,调整后，通杀：%v 通赔：%v", this.sceneId, allKill, allFail)
	//	}
	//	return
	//}
	systemCoinOut := this.GetSystemOut()
	status, _ := base.CoinPoolMgr.GetCoinPoolStatus2(this.GetPlatform(), this.GetGameFreeId(), this.GetGroupId(), systemCoinOut)
	switch status {
	case base.CoinPoolStatus_Normal: //正常模式
		//adjust := this.IsAllKill()
		//if adjust {
		//	logger.Logger.Tracef("scene:%v 水池正常，出现[通杀]，尝试调整一次结果", this.sceneId)
		//} else {
		//	adjust = this.IsFailKill()
		//	if adjust {
		//		logger.Logger.Tracef("scene:%v 水池正常，出现[通赔]，尝试调整一次结果", this.sceneId)
		//	}
		//}
		//if adjust { //通杀或者通赔的情况下，再随机一次
		//	//for i := 0; i < hundredyxx.SEND_CARD_NUM; i++ {
		//	//	this.poker.Restore(this.betFiledInfo[i].cards[2:])
		//	//	this.betFiledInfo[i].cardsO = nil
		//	//}
		//	//for i := 0; i < hundredyxx.SEND_CARD_NUM; i++ {
		//	//	for j := 2; j < hundredyxx.HAND_CARD_NUM; j++ { //重新发后三张
		//	//		this.betFiledInfo[i].cards[j] = int(this.poker.Next())
		//	//	}
		//	//}
		//	this.cpControlled = true
		//	systemCoinOut = this.GetSystemOut()
		//	logger.Logger.Tracef("scene:%v 调整后结果,系统产出：%v", this.sceneId, systemCoinOut)
		//}
		//if !coinPoolMgr.IsMaxOutHaveEnough(this.platform, this.gamefreeId, this.groupId, systemCoinOut) {
		//	for loop := 0; loop < model.GameParamData.HYXXTryAdjustCardTimes; loop++ {
		//		logger.Logger.Tracef("scene:%v 普通水位，开奖结果造成水位异常，尝试调牌 %d", this.sceneId, loop+1)
		//		//for i := 0; i < hundredyxx.SEND_CARD_NUM; i++ {
		//		//	this.poker.Restore(this.betFiledInfo[i].cards[2:])
		//		//	this.betFiledInfo[i].cardsO = nil
		//		//}
		//		//for i := 0; i < hundredyxx.SEND_CARD_NUM; i++ {
		//		//	for j := 2; j < hundredyxx.HAND_CARD_NUM; j++ { //重新发后三张
		//		//		this.betFiledInfo[i].cards[j] = int(this.poker.Next())
		//		//	}
		//		//}
		//		this.cpControlled = true
		//		systemCoinOut = this.GetSystemOut()
		//		if coinPoolMgr.IsMaxOutHaveEnough(this.platform, this.gamefreeId, this.groupId, systemCoinOut) {
		//			return
		//		}
		//	}
		//}
	case base.CoinPoolStatus_Low: //库存值 < 库存下限
		//系统先随机发放用牌，如果庄家不是最大牌，那么系统再根据换好牌概率计算随机是否给庄家调换成最大牌
		if systemCoinOut < 0 {
			//_, isRobBanker := this.GetBanker()
			//logger.Logger.Tracef("scene:%v 低水位，尝试调牌，isRobBanker=%v", this.sceneId, isRobBanker)
			this.SetCpControlled(true)
			this.blackBox.Point = this.GetMinPlayerBetPoint()
			//if isRobBanker {
			//	this.TryEnsureBankerWinByTimes(model.GameParamData.HYXXTryAdjustCardTimes, nil)
			//} else {
			//	this.TryEnsureBankerLoseByTimes(model.GameParamData.HYXXTryAdjustCardTimes, nil)
			//}
		}
	case base.CoinPoolStatus_High, base.CoinPoolStatus_TooHigh: //库存值 > 库存上限
		//adjust := this.IsAllKill()
		//if adjust {
		//	logger.Logger.Tracef("scene:%v 水池盈利，出现[通杀]，尝试调整一次结果", this.sceneId)
		//}
		//if adjust { //通杀的情况下，再随机一次
		//	//for i := 0; i < hundredyxx.SEND_CARD_NUM; i++ {
		//	//	this.poker.Restore(this.betFiledInfo[i].cards[2:])
		//	//	this.betFiledInfo[i].cardsO = nil
		//	//}
		//	//for i := 0; i < hundredyxx.SEND_CARD_NUM; i++ {
		//	//	for j := 2; j < hundredyxx.HAND_CARD_NUM; j++ { //重新发后三张
		//	//		this.betFiledInfo[i].cards[j] = int(this.poker.Next())
		//	//	}
		//	//}
		//	this.cpControlled = true
		//	systemCoinOut = this.GetSystemOut()
		//	logger.Logger.Tracef("scene:%v 调整后结果,系统产出：%v", this.sceneId, systemCoinOut)
		//}
	}
}

func (this *HundredYXXSceneData) TryEnsureBankerWinByTimes(n int, f func() bool) {
	//maxKind := this.betFiledInfo[0].cardsO
	//maxPos := 0
	//for i := 1; i < hundredyxx.SEND_CARD_NUM; i++ {
	//	if this.betFiledInfo[i].cardsO.Value > maxKind.Value {
	//		maxKind = this.betFiledInfo[i].cardsO
	//		maxPos = i
	//	}
	//}
	//
	////先尝试给庄调成相对较大的牌
	//if maxPos != 0 {
	//	old := make([]int, 3, 3)
	//	oldVal := this.betFiledInfo[0].cardsO.Value
	//	for loop := 0; loop < n; loop++ {
	//		copy(old, this.betFiledInfo[0].cards[2:])
	//		this.poker.Restore(this.betFiledInfo[0].cards[2:])
	//		this.betFiledInfo[0].cardsO = nil
	//		for j := 2; j < hundredyxx.HAND_CARD_NUM; j++ { //重新发后三张
	//			this.betFiledInfo[0].cards[j] = int(this.poker.Next())
	//		}
	//
	//		this.betFiledInfo[0].cardsO = hundredyxx.CardsKindFigureUpSington.FigureUpBestCard(this.betFiledInfo[0].cards[:])
	//		if this.betFiledInfo[0].cardsO.Value < oldVal { //比上手还小，还原掉
	//			this.poker.Restore(this.betFiledInfo[0].cards[2:])
	//			this.poker.Kick(old)
	//			copy(this.betFiledInfo[0].cards[2:], old)
	//			this.betFiledInfo[0].cardsO = hundredyxx.CardsKindFigureUpSington.FigureUpBestCard(this.betFiledInfo[0].cards[:])
	//			continue
	//		}
	//		this.cpControlled = true
	//		// 是否满足调控结果
	//		if f != nil && f() {
	//			return
	//		}
	//		if this.betFiledInfo[0].cardsO.Value > maxKind.Value {
	//			return
	//		}
	//	}
	//}
	//
	//type ZoneBetData struct {
	//	pos int
	//	bet int64
	//}
	//var datas [4]*ZoneBetData
	//for i := 1; i < hundredyxx.SEND_CARD_NUM; i++ {
	//	datas[i-1] = &ZoneBetData{
	//		pos: i,
	//		bet: this.betInfo[i] - this.betInfoRob[i],
	//	}
	//}
	////从大到小排序,只统计真实玩家的下注额
	//sort.Slice(datas[:], func(i, j int) bool {
	//	return datas[i].bet > datas[j].bet
	//})
	//
	////至少保证两门
	//adjustCnt := 2 + rand.Intn(2)
	////还不行,尝试换掉其他闲家的牌
	//for i := 0; i < adjustCnt; i++ {
	//	idx := datas[i].pos
	//	old := make([]int, 3, 3)
	//	oldVal := this.betFiledInfo[idx].cardsO.Value
	//	for loop := 0; loop < n; loop++ {
	//		copy(old, this.betFiledInfo[idx].cards[2:])
	//		this.poker.Restore(this.betFiledInfo[idx].cards[2:])
	//		this.betFiledInfo[idx].cardsO = nil
	//		for j := 2; j < hundredyxx.HAND_CARD_NUM; j++ { //重新发后三张
	//			this.betFiledInfo[idx].cards[j] = int(this.poker.Next())
	//		}
	//
	//		this.betFiledInfo[idx].cardsO = hundredyxx.CardsKindFigureUpSington.FigureUpBestCard(this.betFiledInfo[idx].cards[:])
	//		if this.betFiledInfo[idx].cardsO.Value > oldVal { //比上手还大，还原掉
	//			this.poker.Restore(this.betFiledInfo[idx].cards[2:])
	//			this.poker.Kick(old)
	//			copy(this.betFiledInfo[idx].cards[2:], old)
	//			this.betFiledInfo[idx].cardsO = hundredyxx.CardsKindFigureUpSington.FigureUpBestCard(this.betFiledInfo[idx].cards[:])
	//			continue
	//		}
	//		this.cpControlled = true
	//		// 是否满足调控结果
	//		if f != nil && f() {
	//			break
	//		}
	//		if this.betFiledInfo[idx].cardsO.Value < this.betFiledInfo[0].cardsO.Value { //当前比庄家牌小
	//			break
	//		}
	//	}
	//}
}

func (this *HundredYXXSceneData) TryEnsureBankerLoseByTimes(n int, f func() bool) {
	//maxKind := this.betFiledInfo[0].cardsO
	//maxPos := 0
	//for i := 1; i < hundredyxx.SEND_CARD_NUM; i++ {
	//	if this.betFiledInfo[i].cardsO.Value > maxKind.Value {
	//		maxKind = this.betFiledInfo[i].cardsO
	//		maxPos = i
	//	}
	//}
	//
	////先尝试给庄调成相对较小的牌
	//if maxPos != 0 {
	//	old := make([]int, 3, 3)
	//	oldVal := this.betFiledInfo[0].cardsO.Value
	//	for loop := 0; loop < n; loop++ {
	//		copy(old, this.betFiledInfo[0].cards[2:])
	//		this.poker.Restore(this.betFiledInfo[0].cards[2:])
	//		this.betFiledInfo[0].cardsO = nil
	//		for j := 2; j < hundredyxx.HAND_CARD_NUM; j++ { //重新发后三张
	//			this.betFiledInfo[0].cards[j] = int(this.poker.Next())
	//		}
	//
	//		this.betFiledInfo[0].cardsO = hundredyxx.CardsKindFigureUpSington.FigureUpBestCard(this.betFiledInfo[0].cards[:])
	//		if this.betFiledInfo[0].cardsO.Value > oldVal { //比上手还小，还原掉
	//			this.poker.Restore(this.betFiledInfo[0].cards[2:])
	//			this.poker.Kick(old)
	//			copy(this.betFiledInfo[0].cards[2:], old)
	//			this.betFiledInfo[0].cardsO = hundredyxx.CardsKindFigureUpSington.FigureUpBestCard(this.betFiledInfo[0].cards[:])
	//			continue
	//		}
	//		this.cpControlled = true
	//		// 是否满足调控结果
	//		if f != nil && f() {
	//			return
	//		}
	//	}
	//}
	//
	//type ZoneBetData struct {
	//	pos int
	//	bet int64
	//}
	//var datas [4]*ZoneBetData
	//for i := 1; i < hundredyxx.SEND_CARD_NUM; i++ {
	//	datas[i-1] = &ZoneBetData{
	//		pos: i,
	//		bet: this.betInfoRob[i],
	//	}
	//}
	////从大到小排序,只统计机器人的下注额
	//sort.Slice(datas[:], func(i, j int) bool {
	//	return datas[i].bet > datas[j].bet
	//})
	//
	////至少保证两门
	//adjustCnt := 2 + rand.Intn(2)
	////尝试给闲家调较大的牌
	//for i := 0; i < adjustCnt; i++ {
	//	idx := datas[i].pos
	//	old := make([]int, 3, 3)
	//	oldVal := this.betFiledInfo[idx].cardsO.Value
	//	for loop := 0; loop < n; loop++ {
	//		copy(old, this.betFiledInfo[idx].cards[2:])
	//		this.poker.Restore(this.betFiledInfo[idx].cards[2:])
	//		this.betFiledInfo[idx].cardsO = nil
	//		for j := 2; j < hundredyxx.HAND_CARD_NUM; j++ { //重新发后三张
	//			this.betFiledInfo[idx].cards[j] = int(this.poker.Next())
	//		}
	//
	//		this.betFiledInfo[idx].cardsO = hundredyxx.CardsKindFigureUpSington.FigureUpBestCard(this.betFiledInfo[idx].cards[:])
	//		if this.betFiledInfo[idx].cardsO.Value < oldVal { //比上手还小，还原掉
	//			this.poker.Restore(this.betFiledInfo[idx].cards[2:])
	//			this.poker.Kick(old)
	//			copy(this.betFiledInfo[idx].cards[2:], old)
	//			this.betFiledInfo[idx].cardsO = hundredyxx.CardsKindFigureUpSington.FigureUpBestCard(this.betFiledInfo[idx].cards[:])
	//			continue
	//		}
	//		this.cpControlled = true
	//		// 是否满足调控结果
	//		if f != nil && f() {
	//			break
	//		}
	//		if this.betFiledInfo[idx].cardsO.Value > this.betFiledInfo[0].cardsO.Value { //当前比庄家牌大
	//			break
	//		}
	//	}
	//}
}

func (this *HundredYXXSceneData) GetSystemOut() int64 {
	this.CalcWin()
	countTruePlayer := false //默认真人庄
	banker := this.players[this.bankerSnId]
	if this.bankerSnId != -1 && banker != nil {
		if banker.IsRob {
			countTruePlayer = true
		}
	} else {
		countTruePlayer = true
	}
	bankerWinCoin := int64(0)
	CountCoin := [hundredyxx.BetField_MAX]int64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	for _, p := range this.seats {
		p.returnCoin = 0
		if this.bankerSnId == p.SnId {
			continue
		}
		if countTruePlayer {
			if p.IsRob {
				continue
			}
		}
		if !countTruePlayer {
			if !p.IsRob {
				continue
			}
		}
		winCoinCount := int64(0)
		winCoin := [hundredyxx.BetField_MAX]int64{}
		for i := 1; i < hundredyxx.BetField_MAX; i++ {
			if p.betInfo[i] == 0 {
				continue
			}
			if this.betFiledInfo[i].winFlag > 0 {
				winCoin[i] = int64(this.betFiledInfo[i].gainRate) * p.betInfo[i]
			} else {
				winCoin[i] = 0
			}
			CountCoin[i] += winCoin[i]
			winCoinCount += winCoin[i]
		}
		bankerWinCoin += winCoinCount
	}
	if countTruePlayer {
		bankerWinCoin = -bankerWinCoin
	}
	return bankerWinCoin
}

func (this *HundredYXXSceneData) TestCalcWin(points []int32) []HundredYXXBetFieldData {
	testBetZone := [hundredyxx.BetField_MAX]HundredYXXBetFieldData{}
	//先统计单图案输赢
	for i := 0; i <= hundredyxx.BetField_Single_6; i++ {
		winPoint := int32(hundredyxx.BetFieldWinPoint[i][0])
		for j := 0; j < hundredyxx.DICE_NUM; j++ {
			if winPoint == points[j] {
				testBetZone[i].winFlag = 1
				testBetZone[i].gainRate++ //三个骰子中，有一个相应图案中奖倍率为1，有两个中奖倍率为2，有三个中奖倍率为3
			}
		}
	}
	//再统计双图案输赢
	for i := hundredyxx.BetField_Double_7; i < hundredyxx.BetField_MAX; i++ {
		winPoint1 := int32(hundredyxx.BetFieldWinPoint[i][0])
		winPoint2 := int32(hundredyxx.BetFieldWinPoint[i][1])
		if common.InSliceInt32(points[:], winPoint1) && common.InSliceInt32(points[:], winPoint2) {
			testBetZone[i].winFlag = 1
			testBetZone[i].gainRate = 5
		}
	}
	return testBetZone[:]
}

func (this *HundredYXXSceneData) GetTestSystemOut(points []int32) int64 {
	testWinZone := this.TestCalcWin(points[:])

	countTruePlayer := false //默认真人庄
	banker := this.players[this.bankerSnId]
	if this.bankerSnId != -1 && banker != nil {
		if banker.IsRob {
			countTruePlayer = true
		}
	} else {
		countTruePlayer = true
	}
	bankerWinCoin := int64(0)
	CountCoin := [hundredyxx.BetField_MAX]int64{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	for _, p := range this.seats {
		p.returnCoin = 0
		if this.bankerSnId == p.SnId {
			continue
		}
		if countTruePlayer {
			if p.IsRob {
				continue
			}
		}
		if !countTruePlayer {
			if !p.IsRob {
				continue
			}
		}
		winCoinCount := int64(0)
		winCoin := [hundredyxx.BetField_MAX]int64{}
		for i := 1; i < hundredyxx.BetField_MAX; i++ {
			if p.betInfo[i] == 0 {
				continue
			}
			if testWinZone[i].winFlag > 0 {
				winCoin[i] = int64(testWinZone[i].gainRate) * p.betInfo[i]
			} else {
				winCoin[i] = -p.betInfo[i]
			}
			CountCoin[i] += winCoin[i]
			winCoinCount += winCoin[i]
		}
		bankerWinCoin += winCoinCount
	}
	if countTruePlayer {
		bankerWinCoin = -bankerWinCoin
	}
	return bankerWinCoin
}

//取玩家最少赢钱投点
func (this *HundredYXXSceneData) GetMinPlayerBetPoint() [hundredyxx.DICE_NUM]int32 {
	sortArr := []RollPointSort{}
	indexArr := rand.Perm(len(hundredyxx.AllRoll))
	for _, key := range indexArr {
		points := hundredyxx.AllRoll[key]
		systemOut := this.GetTestSystemOut(points[:])
		sortArr = append(sortArr, RollPointSort{
			systemCoinOut: systemOut,
			point:         hundredyxx.AllRoll[key],
		})
	}
	sort.Slice(sortArr, func(i, j int) bool {
		return sortArr[i].systemCoinOut < sortArr[j].systemCoinOut
	})
	if len(sortArr) == 0 {
		return this.blackBox.Point
	}
	return sortArr[rand.Intn(len(sortArr)/3)].point
}

func (this *HundredYXXSceneData) IsAllKill() bool {
	//for i := 0; i < hundredyxx.SEND_CARD_NUM; i++ {
	//	this.betFiledInfo[i].cardsO = hundredyxx.CardsKindFigureUpSington.FigureUpBestCard(this.betFiledInfo[i].cards[:])
	//}
	//banker := this.betFiledInfo[0]
	//for i := 1; i < hundredyxx.SEND_CARD_NUM; i++ {
	//	if this.betFiledInfo[i].cardsO.Value > banker.cardsO.Value {
	//		return false
	//	}
	//}
	return true
}

func (this *HundredYXXSceneData) IsFailKill() bool {
	//for i := 0; i < hundredyxx.SEND_CARD_NUM; i++ {
	//	this.betFiledInfo[i].cardsO = hundredyxx.CardsKindFigureUpSington.FigureUpBestCard(this.betFiledInfo[i].cards[:])
	//}
	//banker := this.betFiledInfo[0]
	//for i := 1; i < hundredyxx.SEND_CARD_NUM; i++ {
	//	if this.betFiledInfo[i].cardsO.Value < banker.cardsO.Value {
	//		return false
	//	}
	//}
	return true
}

func (this *HundredYXXSceneData) ClearUpplayerlist() {
	//清除掉所有金币不足的上庄列表
	if len(this.upplayerlist) > 0 {
		filter := []*HundredYXXPlayerData{}
		for _, pp := range this.upplayerlist {
			if (pp.GetCoin()) < int64(this.GetDBGameFree().GetBanker()) {
				continue
			}
			if this.players[pp.SnId] == nil {
				continue
			}
			filter = append(filter, pp)
		}
		this.upplayerlist = filter
	}
}

func (this *HundredYXXSceneData) InUpplayerlist(snid int32) bool {
	for _, pp := range this.upplayerlist {
		if pp.SnId == snid {
			return true
		}
	}
	return false
}

func (this *HundredYXXSceneData) GetBanker() (*HundredYXXPlayerData, bool) {
	var isRobotBanker bool
	var banker *HundredYXXPlayerData
	if this.bankerSnId != -1 {
		banker = this.players[this.bankerSnId]
		if banker != nil && banker.IsRob {
			isRobotBanker = true
		}
	} else {
		banker = nil
		isRobotBanker = true
	}
	return banker, isRobotBanker
}

func (this *HundredYXXSceneData) TryChangeBanker() {
	if this.upplayerCount >= HYXX_BANKERNUMBERS || this.bankerSnId == -1 {
		this.ClearUpplayerlist()
		if len(this.upplayerlist) > 0 {
			this.upplayerCount = 0
			this.bankerSnId = this.upplayerlist[0].SnId
			this.upplayerlist = this.upplayerlist[1:]
			banker := this.players[this.bankerSnId]
			if banker != nil {
				banker.SetPos(HYXX_BANKERPOS)
				banker.winRecord = []int64{}
				banker.betBigRecord = []int64{}
			} else {
				this.bankerSnId = -1
			}
		} else {
			this.bankerSnId = -1
		}
	}
	//让当前钱不够的庄家下庄
	if this.bankerSnId != -1 {
		banker := this.players[this.bankerSnId]
		if banker != nil {
			if banker.GetCoin() < int64(this.GetDBGameFree().GetBanker()) {
				this.bankerSnId = -1
				banker.SetPos(HYXX_OLPOS)
				banker.winRecord = []int64{}
				banker.betBigRecord = []int64{}
			}
		}
	}
}

// FindSingleAdjustSnId 查询单控玩家id
// 在下注阶段完成后调用
func (this *HundredYXXSceneData) FindSingleAdjustSnId() int32 {
	this.SingleAdjustReset()
	// 单控生效判断
	if this.testing || this.IsPrivateScene() {
		return -1
	}
	// 场里只能有一个真实玩家
	if this.GetRealPlayerCnt() > 1 {
		return -1
	}
	// 只有一个真实玩家下注，或玩家为庄，其余为机器人
	snId := int32(-1)
	realBet := this.RealPlayerBet()
	if realBet <= 0 {
		banker, systemBanker := this.GetBanker()
		if systemBanker || banker == nil {
			return -1
		}
		// 玩家坐庄单控
		snId = banker.GetSnId()
	} else {
		_, systemBanker := this.GetBanker()
		if !systemBanker {
			// 有真实玩家下注，又有玩家坐庄
			return -1
		}
		for _, v := range this.players {
			if v.betTotal > 0 && !v.IsRob {
				if snId > 0 {
					return -1
				} else {
					snId = v.SnId
				}
			}
		}
	}
	if snId == -1 {
		return snId
	}
	// 查询玩家单控配置
	PlayerSingleAdjust := this.players[snId]
	cfg, ok := PlayerSingleAdjust.IsSingleAdjustPlayer()
	if cfg == nil || !ok {
		return -1
	}
	if snId != this.bankerSnId {
		// 配置校验
		if cfg.BetMin < 0 || cfg.BetMax < 0 || cfg.BetMin > cfg.BetMax ||
			cfg.TotalTime <= 0 || cfg.TotalTime <= cfg.CurTime {
			return -1
		}
		// 闲家单控，是否满足下注范围
		bet := this.players[snId].betTotal
		if bet < cfg.BetMin || bet > cfg.BetMax {
			return -1
		}
	} else {
		if cfg.BankerWinMin < 0 || cfg.BankerLoseMin < 0 {
			return -1
		}
	}
	this.singleAdjustSnId = snId
	this.singleAdjustConfig = cfg
	return snId
}

// SingleAdjustReset 重置单控查询结果
func (this *HundredYXXSceneData) SingleAdjustReset() {
	this.singleAdjustSnId = -1
	this.singleAdjustConfig = nil
}

// SingleAdjust 单控调牌
func (this *HundredYXXSceneData) SingleAdjust() {
	systemCoinOut := this.GetSystemOut()
	// todo 庄家单控
	if this.singleAdjustSnId == this.bankerSnId {
		switch this.singleAdjustConfig.Mode {
		case common.SingleAdjustModeWin:

		case common.SingleAdjustModeLose:

		}
		return
	}
	// 闲家单控
	switch this.singleAdjustConfig.Mode {
	case common.SingleAdjustModeWin:
		if systemCoinOut > 0 {
			// 系统庄输
			this.TryEnsureBankerLoseByTimes(model.GameParamData.HYXXTrySingleAdjustCardTimes,
				func() bool {
					if this.GetSystemOut() > 0 {
						return false
					}
					return true
				})
		}
	case common.SingleAdjustModeLose:
		if systemCoinOut < 0 {
			// 系统庄赢
			this.TryEnsureBankerWinByTimes(model.GameParamData.HYXXTrySingleAdjustCardTimes,
				func() bool {
					if this.GetSystemOut() < 0 {
						return false
					}
					return true
				})
		}
	}
}

// SingleAdjustResult 检查单控是否成功
func (this *HundredYXXSceneData) SingleAdjustResult() int32 {
	if this.singleAdjustConfig == nil {
		return 0
	}
	player := this.players[this.singleAdjustSnId]
	if this.bankerSnId == this.singleAdjustSnId {
		// 庄家单控，是否满足单控条件
		switch this.singleAdjustConfig.Mode {
		case common.SingleAdjustModeWin:
			if player.changeScore <= 0 || player.changeScore < this.singleAdjustConfig.BankerWinMin {
				return 0
			}
		case common.SingleAdjustModeLose:
			if player.changeScore >= 0 || -player.changeScore < this.singleAdjustConfig.BankerLoseMin {
				return 0
			}
		default:
			return 0
		}
	} else {
		// 闲家单控，是否满足单控条件
		switch this.singleAdjustConfig.Mode {
		case common.SingleAdjustModeWin:
			if player.changeScore <= 0 {
				return 0
			}
		case common.SingleAdjustModeLose:
			if player.changeScore >= 0 {
				return 0
			}
		default:
			return 0
		}
	}
	player.AddAdjustCount(this.GetGameFreeId())
	return this.singleAdjustConfig.Mode
}

func (this *HundredYXXSceneData) IsSmartOperation() bool {
	if this.testing || this.IsPrivateScene() {
		return false
	}
	return true
}

func (this *HundredYXXSceneData) SmartDeal(state *SceneSendCardStateHundredYXX) {
}
