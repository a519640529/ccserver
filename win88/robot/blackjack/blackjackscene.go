package blackjack

import (
	"games.yol.com/win88/common"
	rule "games.yol.com/win88/gamerule/blackjack"
	"games.yol.com/win88/protocol/blackjack"
	"games.yol.com/win88/protocol/player"
	"games.yol.com/win88/robot/base"
	"github.com/idealeak/goserver/core/netlib"
)

type BlackJackScene struct {
	base.BaseScene
	*blackjack.SCBlackJackRoomInfo
	players map[int32]*BlackJackPlayer
}

func NewBlackJackScene(info *blackjack.SCBlackJackRoomInfo) *BlackJackScene {
	s := &BlackJackScene{
		SCBlackJackRoomInfo: info,
		players:             make(map[int32]*BlackJackPlayer),
	}
	return s
}

func (this *BlackJackScene) GetPlayerByPos(pos int32) base.Player {
	if pos >= 0 && pos <= rule.MaxPlayer {
		for _, v := range this.players {
			if v.GetSeat() == pos {
				return v
			}
		}
	}
	return nil
}

func (this *BlackJackScene) GetPlayerBySnid(snid int32) base.Player {
	if p, ok := this.players[snid]; ok {
		return p
	}
	return nil
}

func (this *BlackJackScene) GetMe(s *netlib.Session) base.Player {
	if user, ok := s.GetAttribute(base.SessionAttributeUser).(*player.SCPlayerData); ok {
		return this.GetPlayerBySnid(user.GetData().GetSnId())
	}
	return nil
}

func (this *BlackJackScene) AddPlayer(p base.Player) {
	if mp, ok := p.(*BlackJackPlayer); ok {
		this.players[p.GetSnId()] = mp
	}
}

func (this *BlackJackScene) DelPlayer(snid int32) {
	if _, exist := this.players[snid]; exist {
		delete(this.players, snid)
	}
}

func (this *BlackJackScene) IsFull() bool {
	return len(this.players) > rule.MaxPlayer
}

func (this *BlackJackScene) IsMatchScene() bool {
	return this.GetRoomId() >= common.MatchSceneStartId && this.GetRoomId() <= common.MatchSceneMaxId
}

func (this *BlackJackScene) Update(ts int64) {

}
