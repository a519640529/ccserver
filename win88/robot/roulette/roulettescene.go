package roulette

import (
	"games.yol.com/win88/common"
	rule "games.yol.com/win88/gamerule/roulette"
	"games.yol.com/win88/proto"
	proto_player "games.yol.com/win88/protocol/player"
	proto_roulette "games.yol.com/win88/protocol/roulette"
	"games.yol.com/win88/robot/base"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/core/timer"
	"math/rand"
	"time"
)

type RouletteScene struct {
	base.BaseScene
	*proto_roulette.SCRouletteRoomInfo
	players map[int32]*RoulettePlayer
	point   *rule.PointType
}

func NewRouletteScene(info *proto_roulette.SCRouletteRoomInfo) *RouletteScene {
	s := &RouletteScene{
		SCRouletteRoomInfo: info,
		players:            make(map[int32]*RoulettePlayer),
	}
	s.Init()
	return s
}

func (s *RouletteScene) Init() {
	s.players = make(map[int32]*RoulettePlayer)
	for _, mpd := range s.GetPlayers() {
		p := NewRoulettePlayer(mpd)
		if p != nil {
			s.AddPlayer(p)
		}
	}
	s.point = new(rule.PointType)
	s.point.Init()
}

func (s *RouletteScene) Clear() {
	for _, player := range s.players {
		player.Clear()
	}
}

func (s *RouletteScene) AddPlayer(p base.Player) {
	if mp, ok := p.(*RoulettePlayer); ok {
		s.players[p.GetSnId()] = mp
	}
}

func (s *RouletteScene) DelPlayer(snid int32) {
	if p, exist := s.players[snid]; exist && p != nil {
		delete(s.players, snid)
	}
}

func (s *RouletteScene) GetPlayerByPos(pos int32) base.Player {
	return nil
}

func (s *RouletteScene) GetPlayerBySnid(snid int32) base.Player {
	if p, exist := s.players[snid]; exist {
		return p
	}
	return nil
}

func (this *RouletteScene) GetMe(s *netlib.Session) base.Player {
	if user, ok := s.GetAttribute(base.SessionAttributeUser).(*proto_player.SCPlayerData); ok {
		player := this.GetPlayerBySnid(user.GetData().GetSnId())
		return player
	}
	return nil
}

func (s *RouletteScene) IsFull() bool {
	return false
}

func (s *RouletteScene) IsMatchScene() bool {
	return false
}

func (s *RouletteScene) IsCoinScene() bool {
	return false
}

var RouletteChipWeight = []int64{7000, 2000, 800, 200}

//下注区域概率 直接注 分注 街注/三个号码 角注/四个号码 线注 三面 双面
var RouletteBetAreaWeight = []int64{500, 1000, 1500, 1500, 1500, 1500, 2500}

//双面下注区域概率
var RouletteTwoSideWeight = []int64{3333, 3333, 3333}

func (s *RouletteScene) GetRandBetArea(player *RoulettePlayer) int {
	idx := common.RandSliceIndexByWight(RouletteBetAreaWeight)
	switch idx {
	case rule.BetTypeTwoSide:
		idx2 := common.RandSliceIndexByWight(RouletteTwoSideWeight)
		twoSide := []int{1, 2, 3}
		switch twoSide[idx2] {
		case 1:
			if player.singleDouble == 0 {
				//单双
				if rand.Intn(100) > 50 {
					player.singleDouble = rule.PointSingle
				} else {
					player.singleDouble = rule.PointDouble
				}
			}
		case 2:
			if player.redBlack == 0 {
				//红黑
				if rand.Intn(100) > 50 {
					player.redBlack = rule.PointRed
				} else {
					player.redBlack = rule.PointBlack
				}
			}
		case 3:
			if player.lowHi == 0 {
				//小大
				if rand.Intn(100) > 50 {
					player.lowHi = rule.PointLow
				} else {
					player.lowHi = rule.PointHi
				}
			}
		}
		return player.singleDouble
	}
	points := s.point.PointTypeMap[idx]
	if len(points) <= 0 {
		return rand.Intn(157)
	}
	pos := []int{}
	for k, _ := range points {
		pos = append(pos, k)
	}
	return pos[rand.Intn(len(pos))]
}

func (s *RouletteScene) GetCanBetChip(player *RoulettePlayer, otherParams []int32, isTopPlayer bool) []int64 {
	maxIdx := -1
	for k, v := range otherParams {
		if player.GetCoin()-player.GetBetCoin() >= int64(v) {
			maxIdx = k + 1
		}
	}
	toprcw := []int64{1000, 2000, 4000, 3000}
	if maxIdx == -1 {
		maxIdx = len(toprcw)
	}
	if isTopPlayer {
		return toprcw[:maxIdx]
	} else {
		return RouletteChipWeight[:maxIdx]
	}
}

func (s *RouletteScene) Action(ss *netlib.Session, player *RoulettePlayer, initTime time.Time, isTopPlayer bool) {
	//玩家下注
	randTime := time.Duration(0)
	betInterval := time.Duration(0)
	if isTopPlayer {
		//上座玩家时间
		if rand.Intn(100) > 50 {
			randTime = time.Duration(rand.Intn(4800) + 200) //确定下注开始时间
		} else {
			randTime = time.Duration(rand.Intn(3000) + 10000) //确定下注开始时间
		}
		betInterval = time.Duration(rand.Int31n(300) + 200)
	} else {
		randTime = time.Duration(200) //确定下注开始时间
		betInterval = time.Duration(rand.Int31n(500) + 800)
	}
	ok := base.StartSessionGameTimer(ss, timer.TimerActionWrapper(func(h timer.TimerHandle, ud interface{}) bool {
		if s.GetState() != int32(rule.RouletteSceneStateBet) {
			return true
		}
		if time.Now().Before(player.tNextBet) {
			return true
		}
		newT := time.Now().Sub(player.tNextBet) / time.Millisecond
		if player.betTime-newT < 0 {
			return true
		}
		player.betTime -= newT

		chipIdx := 0
		params := s.GetParamsEx()
		if len(params) != 0 {
			dbGameFree := base.SceneMgrSington.GetSceneDBGameFree(s.GetRoomId(), params[0])
			if dbGameFree != nil {
				otherParams := dbGameFree.GetOtherIntParams()
				rcw := s.GetCanBetChip(player, otherParams, isTopPlayer)
				if len(rcw) == 0 {
					return true
				}
				chipIdx = common.RandSliceIndexByWight(rcw)
				chip := otherParams[chipIdx]
				//金币不够
				if player.GetCoin()-player.GetBetCoin() < int64(chip) {
					//	n := len(otherParams)
					//	coin := otherParams[n-1]
					//	ExePMCmd(ss, fmt.Sprintf("%v%v%v", common.PMCmd_AddCoin, common.PMCmd_SplitToken, coin))
					//	player.tNextBet = time.Now().Add(time.Second * 10)
					//	player.Coin = proto.Int64(player.GetCoin() + int64(coin))
					return true
				}
			}
		}
		idx := s.GetRandBetArea(player)
		player.choose = int32(idx)
		pack := &proto_roulette.CSRoulettePlayerOp{
			OpCode:  proto.Int32(int32(rule.RoulettePlayerOpBet)),
			OpParam: []int64{int64(idx), int64(chipIdx)},
		}
		proto.SetDefaults(pack)
		ss.Send(int(proto_roulette.RouletteMmoPacketID_PACKET_CS_Roulette_PlayerOp), pack)
		player.tNextBet = time.Now().Add(betInterval * time.Millisecond)
		return true
	}), nil, time.Millisecond*200, -1)
	if ok {
		player.tNextBet = time.Now().Add(randTime * time.Millisecond)
	}

}

func (s *RouletteScene) Update(ts int64) {}
