package tienlen

import (
	"games.yol.com/win88/protocol/player"
	proto_tienlen "games.yol.com/win88/protocol/tienlen"
	"games.yol.com/win88/robot/base"
	"github.com/idealeak/goserver/core/netlib"
)

type TienLenScene struct {
	base.BaseScene
	*proto_tienlen.SCTienLenRoomInfo
	players map[int32]*TienLenPlayer
}

func NewTienLenScene(info *proto_tienlen.SCTienLenRoomInfo) *TienLenScene {
	s := &TienLenScene{
		SCTienLenRoomInfo: info,
		players:           make(map[int32]*TienLenPlayer),
	}
	s.Init()
	return s
}

func (s *TienLenScene) Init() {
	s.players = make(map[int32]*TienLenPlayer)
	for _, mpd := range s.GetPlayers() {
		p := NewTienLenPlayer(mpd)
		if p != nil {
			s.AddPlayer(p)
		}
	}
}
func (s *TienLenScene) GetIsAllAi() bool {
	var i int
	for _, p := range s.players {
		if p.IsRobot() {
			i++
		}
	}
	return i == 4
}
func (s *TienLenScene) Clear() {
	for _, p := range s.players {
		p.Clear()
	}
}

func (s *TienLenScene) AddPlayer(p base.Player) {
	if mp, ok := p.(*TienLenPlayer); ok {
		s.players[p.GetSnId()] = mp
	}
}

func (s *TienLenScene) DelPlayer(snid int32) {
	if p, exist := s.players[snid]; exist && p != nil {
		delete(s.players, snid)
	}
}

func (s *TienLenScene) GetPlayerByPos(pos int32) base.Player {
	return nil
}

func (s *TienLenScene) GetPlayerBySnid(snid int32) base.Player {
	if p, exist := s.players[snid]; exist {
		return p
	}
	return nil
}

func (this *TienLenScene) GetMe(s *netlib.Session) base.Player {
	if user, ok := s.GetAttribute(base.SessionAttributeUser).(*player.SCPlayerData); ok {
		p := this.GetPlayerBySnid(user.GetData().GetSnId())
		return p
	}
	return nil
}

func (s *TienLenScene) IsFull() bool {
	return len(s.players) >= int(s.MaxPlayerNum)
}

func (s *TienLenScene) IsMatchScene() bool {
	return s.IsMatch == 1 || s.IsMatch == 2 //锦标赛和冠军赛
}

func (s *TienLenScene) IsCoinScene() bool {
	return true
}
func (this *TienLenScene) Update(ts int64) {
}
