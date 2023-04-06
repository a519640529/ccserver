package hundredyxx

import (
	"games.yol.com/win88/protocol/hundredyxx"
	"games.yol.com/win88/protocol/player"
	"games.yol.com/win88/protocol/server"
	"games.yol.com/win88/robot/base"
	"github.com/idealeak/goserver/core/netlib"
	"math/rand"
	"time"
)

type HundredYXXScene struct {
	base.BaseScene
	*hundredyxx.SCHundredYXXRoomInfo
	players             map[int32]*HundredYXXPlayer
	dbGameFree          *server.DB_GameFree //百人场静态数据
	numOfGame           int32
	waitingBankerNum    int32
	maxRate             int32
	roomtype            int32
	billed              bool
	nextCheckUpBankTime time.Time //下次检测上庄时间
	tryUpBanker         map[int32]*HundredYXXPlayer
	loadingBankList     bool
}

func NewHundredYXXScene(info *hundredyxx.SCHundredYXXRoomInfo) *HundredYXXScene {
	s := &HundredYXXScene{
		SCHundredYXXRoomInfo: info,
		players:              make(map[int32]*HundredYXXPlayer),
		tryUpBanker:          make(map[int32]*HundredYXXPlayer),
	}
	s.Init()
	return s
}

func (this *HundredYXXScene) UpdateInfo(info *hundredyxx.SCHundredYXXRoomInfo) {
	this.SCHundredYXXRoomInfo = info
}

func (this *HundredYXXScene) ResetCheckTime() {
	this.nextCheckUpBankTime = time.Now().Add(time.Second * time.Duration(rand.Int31n(100)))
}

func (this *HundredYXXScene) IsCheckTimeout() bool {
	return time.Now().After(this.nextCheckUpBankTime)
}

func (s *HundredYXXScene) Init() {
	s.numOfGame = s.GetNumOfGames()
	for _, mpd := range s.GetPlayers() {
		p := NewHundredYXXPlayer(mpd)
		if p != nil {
			s.AddPlayer(p)
		}
	}
}

func (s *HundredYXXScene) Clear(numOfGame int32) {
	if numOfGame == s.numOfGame {
		for _, player := range s.players {
			player.Clear()
		}
		s.maxRate = 0
		s.loadingBankList = false
		s.numOfGame++
	}
}

func (s *HundredYXXScene) AddPlayer(p base.Player) {
	if mp, ok := p.(*HundredYXXPlayer); ok {
		s.players[p.GetSnId()] = mp
	}
}

func (s *HundredYXXScene) DelPlayer(snid int32) {
	if p, exist := s.players[snid]; exist && p != nil {
		delete(s.players, snid)
	}
}

func (s *HundredYXXScene) GetPlayerByPos(pos int32) base.Player {
	return nil
}

func (s *HundredYXXScene) GetPlayerBySnid(snid int32) base.Player {
	if p, exist := s.players[snid]; exist {
		return p
	}
	return nil
}

func (this *HundredYXXScene) GetMe(s *netlib.Session) base.Player {
	if user, ok := s.GetAttribute(base.SessionAttributeUser).(*player.SCPlayerData); ok {
		player := this.GetPlayerBySnid(user.GetData().GetSnId())
		return player
	}
	return nil
}

func (s *HundredYXXScene) IsFull() bool {
	return false
}

func (s *HundredYXXScene) IsMatchScene() bool {
	return false
}

//评估要下的倍率
func (s *HundredYXXScene) EvalChoose(player *HundredYXXPlayer) int {
	if player == nil {
		return 0
	}
	return 0
}

func (s *HundredYXXScene) GetBankerRatesLimit() int {
	return 0
}

func (s *HundredYXXScene) Update(ts int64) {}
