package baccarat

import (
	"fmt"
	proto_baccarat "games.yol.com/win88/protocol/baccarat"
	proto_player "games.yol.com/win88/protocol/player"
	proto_server "games.yol.com/win88/protocol/server"
	"math/rand"
	"time"

	"games.yol.com/win88/proto"

	"games.yol.com/win88/common"
	"games.yol.com/win88/gamerule/baccarat"

	"games.yol.com/win88/robot/base"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/core/timer"
)

//const (
//	DVST_ZONE_DRAW int = iota
//	DVST_ZONE_DRAGON
//	DVST_ZONE_TIGER
//	DVST_ZONE_MAX
//)

type BaccaratScene struct {
	base.BaseScene
	*proto_baccarat.SCBaccaratRoomInfo
	players         map[int32]*BaccaratPlayer
	totalBet        map[int64]int64
	dbGameFree      *proto_server.DB_GameFree //百人场静态数据
	upBankerListNum int32                     //上庄列表数量
	bankerList      map[int32]int64
	isRandNum       map[int32]int32
}

func NewBaccaratScene(info *proto_baccarat.SCBaccaratRoomInfo) *BaccaratScene {
	s := &BaccaratScene{
		SCBaccaratRoomInfo: info,
		players:            make(map[int32]*BaccaratPlayer),
	}
	s.Init()
	return s
}

func (s *BaccaratScene) Init() {
	s.players = make(map[int32]*BaccaratPlayer)
	for _, mpd := range s.GetPlayers() {
		p := NewBaccaratPlayer(mpd)
		if p != nil {
			s.AddPlayer(p)
		}
	}
	s.totalBet = make(map[int64]int64)
	for zone, chips := range s.GetTotalChips() {
		s.totalBet[int64(zone)] += chips
	}
	s.isRandNum = make(map[int32]int32)
	s.bankerList = make(map[int32]int64)
	s.upBankerListNum = 0
}

func (s *BaccaratScene) Clear() {
	for _, player := range s.players {
		player.Clear()
	}
	s.totalBet = make(map[int64]int64)
	s.isRandNum[s.dbGameFree.GetId()] = rand.Int31n(3) + 1
}

func (s *BaccaratScene) AddPlayer(p base.Player) {
	if mp, ok := p.(*BaccaratPlayer); ok {
		s.players[p.GetSnId()] = mp
	}
}

func (s *BaccaratScene) DelPlayer(snid int32) {
	if p, exist := s.players[snid]; exist && p != nil {
		delete(s.players, snid)
		if _, ok := s.bankerList[snid]; ok {
			delete(s.bankerList, snid)
		}
	}
}

func (s *BaccaratScene) GetPlayerByPos(pos int32) base.Player {
	return nil
}

func (s *BaccaratScene) GetPlayerBySnid(snid int32) base.Player {
	if p, exist := s.players[snid]; exist {
		return p
	}
	return nil
}

func (this *BaccaratScene) GetMe(s *netlib.Session) base.Player {
	if user, ok := s.GetAttribute(base.SessionAttributeUser).(*proto_player.SCPlayerData); ok {
		player := this.GetPlayerBySnid(user.GetData().GetSnId())
		return player
	}
	return nil
}

func (s *BaccaratScene) IsFull() bool {
	return false
}

func (s *BaccaratScene) IsMatchScene() bool {
	return false
}

func (s *BaccaratScene) IsCoinScene() bool {
	return false
}

var BaccaratChipWeight = []int64{60, 20, 10, 8, 2}
var BaccaratAreaWeight = []int64{8000 / 2, 8000 / 2, 600, 600, 800}

func (scene *BaccaratScene) Action(s *netlib.Session, player *BaccaratPlayer) {
	logger.Logger.Trace("(scene *BaccaratScene) Action ", player.GetSnId())

	//pool := []int{DVST_ZONE_DRAGON, DVST_ZONE_TIGER, DVST_ZONE_DRAGON, DVST_ZONE_TIGER, DVST_ZONE_DRAGON, DVST_ZONE_TIGER, DVST_ZONE_DRAW, DVST_ZONE_DRAGON, DVST_ZONE_TIGER, DVST_ZONE_DRAGON, DVST_ZONE_TIGER, DVST_ZONE_DRAGON, DVST_ZONE_TIGER, DVST_ZONE_DRAGON, DVST_ZONE_TIGER, DVST_ZONE_DRAGON, DVST_ZONE_TIGER, DVST_ZONE_DRAW}
	ok := base.StartSessionGameTimer(s, timer.TimerActionWrapper(func(h timer.TimerHandle, ud interface{}) bool {
		if scene.GetState() != int32(baccarat.BaccaratSceneStateStake) {
			return true
		}
		if time.Now().Before(player.tNextBet) {
			return true
		}

		var idx int32 = 1
		idx = idx << (1 + uint32(rand.Intn(2)))
		//if player.choose == -1 {
		//	idx = (rand.Int31n(2) + 1) * 2
		//	player.choose = idx
		//} else {
		//	idx = player.choose
		//	if rand.Int31n(200) <= 5 {
		//		idx = int32(baccarat.BACCARAT_ZONE_BANKER_DOUBLE)
		//	} else if rand.Int31n(200) <= 5 {
		//		idx = int32(baccarat.BACCARAT_ZONE_PLAYER_DOUBLE)
		//	} else if rand.Int31n(100) <= 5 {
		//		idx = int32(baccarat.BACCARAT_ZONE_TIE)
		//	}
		//}
		rn := common.RandSliceIndexByWight(BaccaratAreaWeight)
		switch rn {
		case 0:
			idx = int32(baccarat.BACCARAT_ZONE_BANKER)
		case 1:
			idx = int32(baccarat.BACCARAT_ZONE_PLAYER)
		case 2:
			idx = int32(baccarat.BACCARAT_ZONE_BANKER_DOUBLE)
		case 3:
			idx = int32(baccarat.BACCARAT_ZONE_PLAYER_DOUBLE)
		case 4:
			idx = int32(baccarat.BACCARAT_ZONE_TIE)
		}
		if rn == 0 || rn == 1 {
			if player.choose == -1 {
				player.choose = idx
			} else {
				idx = player.choose
			}
		}
		chip := int32(0)
		params := scene.GetParamsEx()
		if len(params) != 0 {
			dbGameFree := base.SceneMgrSington.GetSceneDBGameFree(scene.GetRoomId(), params[0])
			if dbGameFree != nil {
				otherParams := dbGameFree.GetOtherIntParams()
				o := common.RandSliceIndexByWight(BaccaratChipWeight)
				chip = otherParams[o]

				//金币不够
				if player.GetCoin() < int64(chip) {
					n := len(otherParams)
					coin := otherParams[n-1]
					base.ExePMCmd(s, fmt.Sprintf("%v%v%v", common.PMCmd_AddCoin, common.PMCmd_SplitToken, coin))
					player.tNextBet = time.Now().Add(time.Second * 10)
					player.Coin = proto.Int64(player.GetCoin() + int64(coin))
					return true
				}
			}
		}

		pack := &proto_baccarat.CSBaccaratOp{
			OpCode: proto.Int32(int32(baccarat.BaccaratPlayerOpBet)),
			Params: []int64{int64(idx), int64(chip)},
		}
		proto.SetDefaults(pack)
		s.Send(int(proto_baccarat.BaccaratPacketID_PACKET_CS_BACCARAT_PLAYEROP), pack)
		player.tNextBet = time.Now().Add(time.Duration(rand.Int31n(1000)+500) * time.Millisecond)
		return true
	}), nil, time.Millisecond*200, -1)
	if ok {
		player.tNextBet = time.Now().Add(time.Duration(rand.Int31n(1000)+500) * time.Millisecond)
	}
}

func (this *BaccaratScene) Update(ts int64) {
	//随机上庄
	if len(this.bankerList) > 0 {
		if this.upBankerListNum < this.isRandNum[this.dbGameFree.GetId()] {
			for snid, ts := range this.bankerList {
				if time.Now().Unix() >= ts {
					player := this.players[snid]
					if player != nil {
						player.Banker()
					}
					delete(this.bankerList, snid)
					break
				}
			}
		}
	}
}
func (this *BaccaratScene) UpBanker(p *BaccaratPlayer) {
	if this.dbGameFree.GetBanker() == 0 {
		return
	}
	if p == nil {
		return
	}
	if p.GetCoin() < int64(this.dbGameFree.GetBanker()) {
		return
	}
	n := this.upBankerListNum
	if n > this.isRandNum[this.dbGameFree.GetId()] {
		return
	}
	bankRate := 100 / (this.dbGameFree.GetSceneType())                   //1,2,3,4->100,50,33,25
	bankRate = int32(float32(bankRate) * float32(float32(bankRate)/100)) ///100,25,10,8
	if rand.Int31n(100) > bankRate {
		return
	}
	if _, ok := this.bankerList[p.GetSnId()]; !ok && len(this.bankerList) < 3 {
		this.bankerList[p.GetSnId()] = time.Now().Add(time.Duration(rand.Intn(20)) * time.Second).Unix()
	}
}
