package luckydice

import (
	"container/list"
	"time"

	"games.yol.com/win88/gamerule/luckydice"
	"games.yol.com/win88/minigame/base"
	"games.yol.com/win88/proto"
	proto_luckydice "games.yol.com/win88/protocol/luckydice"
	"github.com/idealeak/goserver/core/timer"
)

type LuckyDiceSceneData struct {
	*base.Scene                                    //房间信息
	players         map[int32]*LuckyDicePlayerData //玩家信息
	betBigPlayers   map[int32]*LuckyDicePlayerData //押大的玩家信息
	betSmlPlayers   map[int32]*LuckyDicePlayerData //押小的玩家信息
	roundID         int32                          //局数编号
	dices           *luckydice.DiceResult          //骰子
	totalBetBig     int64                          //押大总投注
	totalBetSmall   int64                          //押小总投注
	bigBets         *list.List
	smlBets         *list.List
	allBets         *list.List
	dicesHistory    []*luckydice.DiceResult
	roundHistory    []*proto_luckydice.LuckyDiceRoundSimpleInfo
	roundBetHistory map[int32]*proto_luckydice.SCLuckyDiceRoundBetHistory
	betChanged      bool
	bIntervention   bool //本局是否控制骰子值
}

type LuckyDiceBetInfo struct {
	AccountID int32
	BetSide   int32
	Bet       int64
	IsBotBet  bool
	BetTime   int64
}

func NewLuckyDiceSceneData(s *base.Scene) *LuckyDiceSceneData {
	return &LuckyDiceSceneData{
		Scene:   s,
		players: make(map[int32]*LuckyDicePlayerData),
	}
}

func (this *LuckyDiceSceneData) SaveData(force bool) {
}

func (this *LuckyDiceSceneData) OnPlayerLeave(p *base.Player, reason int) {
	if p, exist := this.players[p.SnId]; exist {
		if p.bet > 0 {
			p.isLeave = true //todo
		} else {
			delete(this.players, p.SnId)
		}
	}
}

func (this *LuckyDiceSceneData) SceneDestroy(force bool) {
	//销毁房间
	this.Scene.Destroy(force)
}

func (this *LuckyDiceSceneData) init() bool {
	if this.DbGameFree == nil {
		return false
	}
	this.roundID = 100000 //todo
	this.dicesHistory = make([]*luckydice.DiceResult, 0)
	this.roundHistory = make([]*proto_luckydice.LuckyDiceRoundSimpleInfo, 0)
	this.roundBetHistory = make(map[int32]*proto_luckydice.SCLuckyDiceRoundBetHistory)
	//this.BroadcastBetChange()
	return true
}

func (this *LuckyDiceSceneData) RefreshSceneData() {
	for _, player := range this.betBigPlayers {
		player.Clean()
	}
	for _, player := range this.betSmlPlayers {
		player.Clean()
	}
	this.betBigPlayers = make(map[int32]*LuckyDicePlayerData)
	this.betSmlPlayers = make(map[int32]*LuckyDicePlayerData)
	this.dices = nil
	this.totalBetBig = 0
	this.totalBetSmall = 0
	this.bigBets = list.New()
	this.smlBets = list.New()
	this.allBets = list.New()
	this.betChanged = true
	this.roundID++
}

func (this *LuckyDiceSceneData) HandleHistory() {
	if this.dices == nil {
		return
	}

	this.dicesHistory = append(this.dicesHistory, this.dices)
	if len(this.dicesHistory) > 100 {
		this.dicesHistory = this.dicesHistory[1:]
	}

	this.roundHistory = append(this.roundHistory, &proto_luckydice.LuckyDiceRoundSimpleInfo{
		WinSide: proto.Int32(this.dices.BigSmall()),
		RoundId: proto.Int32(this.roundID),
	})
	if len(this.roundHistory) > 20 {
		delete(this.roundBetHistory, this.roundHistory[0].GetRoundId())
		this.roundHistory = this.roundHistory[1:]
	}
}

func (this *LuckyDiceSceneData) BroadcastBetChange() {
	timer.StartTimer(timer.TimerActionWrapper(func(h timer.TimerHandle, ud interface{}) bool {
		if this.betChanged {
			this.betChanged = false
			pack := &proto_luckydice.SCLuckyDiceBetChange{
				TotalBet:    []int64{this.totalBetBig, this.totalBetSmall},
				TotalPlayer: []int32{int32(len(this.betBigPlayers)), int32(len(this.betSmlPlayers))},
			}
			proto.SetDefaults(pack)
			this.Broadcast(int(proto_luckydice.LuckyDicePacketID_PACKET_SC_LUCKYDICE_BETCHANGE), pack, -1)
		}
		this.BroadcastBetChange()
		return true
	}), nil, time.Second, 1)
}
