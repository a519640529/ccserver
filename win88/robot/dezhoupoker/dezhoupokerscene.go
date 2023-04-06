package dezhoupoker

import (
	"games.yol.com/win88/common"
	rule "games.yol.com/win88/gamerule/dezhoupoker"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/dezhoupoker"
	"games.yol.com/win88/protocol/player"
	"games.yol.com/win88/robot/base"
	"github.com/idealeak/goserver/core/netlib"
)

type DezhouPokerScene struct {
	base.BaseScene
	*dezhoupoker.SCDezhouPokerRoomInfo
	*dezhoupoker.SCDezhouPokerBankerAndBlindPos
	*dezhoupoker.SCDezhouPokerOperNotify
	seats              [rule.MaxNumOfPlayer]*DezhouPokerPlayer
	players            map[int32]*DezhouPokerPlayer
	waitOpTickCnt      int64 //等待玩家操作时长
	winRobotSnid       []int32
	bigWinSnid         []int32
	bigWinRobotSnid    []int32
	loseRobotSnid      []int32
	notBigWinRobotSnid []int32
	stateFlag          int32
}

func NewDezhouPokerScene(info *dezhoupoker.SCDezhouPokerRoomInfo) *DezhouPokerScene {
	s := &DezhouPokerScene{
		SCDezhouPokerRoomInfo:          info,
		SCDezhouPokerBankerAndBlindPos: nil,
		SCDezhouPokerOperNotify:        nil,
		players:                        make(map[int32]*DezhouPokerPlayer),
	}
	s.Init()
	return s
}

func (this *DezhouPokerScene) Init() {
	for _, mpd := range this.GetPlayers() {
		p := NewDezhouPokerPlayer(mpd)
		if p != nil {
			this.AddPlayer(p)
		}
	}
}

func (this *DezhouPokerScene) Clear() {
	for _, player := range this.players {
		player.Clear()
	}
	this.stateFlag = 0
}

func (this *DezhouPokerScene) AddPlayer(p base.Player) {
	if mp, ok := p.(*DezhouPokerPlayer); ok {
		this.players[p.GetSnId()] = mp
		this.seats[p.GetPos()] = mp
	}
}

func (this *DezhouPokerScene) DelPlayer(snid int32) {
	if p, exist := this.players[snid]; exist && p != nil {
		delete(this.players, snid)
		this.seats[p.GetPos()] = nil
	}
}

func (this *DezhouPokerScene) GetPlayerByPos(pos int32) base.Player {
	if pos >= 0 && pos < rule.MaxNumOfPlayer {
		return this.seats[pos]
	}
	return nil
}

func (this *DezhouPokerScene) GetPlayerBySnid(snid int32) base.Player {
	if p, exist := this.players[snid]; exist {
		return p
	}
	return nil
}

func (this *DezhouPokerScene) GetMe(s *netlib.Session) base.Player {
	if user, ok := s.GetAttribute(base.SessionAttributeUser).(*player.SCPlayerData); ok {
		player := this.GetPlayerBySnid(user.GetData().GetSnId())
		return player
	}
	return nil
}

func (this *DezhouPokerScene) GetPlayerCount() int32 {
	return int32(len(this.players))
}

func (this *DezhouPokerScene) IsFull() bool {
	return len(this.players) >= int(rule.MaxNumOfPlayer)
}

func (this *DezhouPokerScene) IsMatchScene() bool {
	return this.GetRoomId() >= common.MatchSceneStartId && this.GetRoomId() <= common.MatchSceneMaxId
}

func (this *DezhouPokerScene) IsCoinScene() bool {
	return this.GetRoomId() >= common.CoinSceneStartId && this.GetRoomId() <= common.CoinSceneMaxId
}

func (this *DezhouPokerScene) GetBankerAndBlindPos() *dezhoupoker.SCDezhouPokerBankerAndBlindPos {
	return this.SCDezhouPokerBankerAndBlindPos
}

func (this *DezhouPokerScene) SetBankerAndBlindPos(info *dezhoupoker.SCDezhouPokerBankerAndBlindPos) {
	this.SCDezhouPokerBankerAndBlindPos = info
}

func (this *DezhouPokerScene) AddNewCard(cardType int32, card []int32) {
	cardNum := len(this.Cards)
	for i := int32(cardNum); i < rule.CommunityCardNum; i++ {
		this.Cards = append(this.Cards, rule.INVALIDE_CARD)
	}

	if cardType == rule.CardType_FlopCard {
		this.Cards[0] = card[0]
		this.Cards[1] = card[1]
		this.Cards[2] = card[2]
	} else if cardType == rule.CardType_TrunCard {
		this.Cards[rule.TurnCardPos-1] = card[0]
	} else if cardType == rule.CardType_RiverCard {
		this.Cards[rule.RiverCardPos-1] = card[0]
	}
}

func (this *DezhouPokerScene) GetDezhouPokerOperNotify() *dezhoupoker.SCDezhouPokerOperNotify {
	return this.SCDezhouPokerOperNotify
}

func (this *DezhouPokerScene) SetDezhouPokerOperNotify(info *dezhoupoker.SCDezhouPokerOperNotify) {
	this.SCDezhouPokerOperNotify = info
}

func (this *DezhouPokerScene) OnNewGame() {
	this.Clear()
}

func (this *DezhouPokerScene) Update(ts int64) {
}

func (this *DezhouPokerScene) GetRobotCount() int32 {
	robotCount := int32(0)
	for _, player := range this.players {
		if player.IsRobot() {
			robotCount++
		}
	}
	return robotCount
}

/////////////////////////////////////////////////////////////////////////////////////////////////////
func (this *DezhouPokerScene) OnDezhouPokerNormalBilled(info *dezhoupoker.SCDezhouPokerGameBilled) {
	var winnerSnid []int32
	for i := 0; i < len(info.Datas); i++ {
		winnerPos := info.Datas[i].GetPos()
		curPlayer := this.seats[winnerPos]
		if curPlayer != nil {
			winnerSnid = append(winnerSnid, curPlayer.GetSnId())
			if curPlayer.IsRobot() {
				this.winRobotSnid = append(this.winRobotSnid, curPlayer.GetSnId())
			}

			//50倍大盲
			if info.Datas[i].GetWinCoin() >= this.SCDezhouPokerRoomInfo.GetSmallBlind()*2*50 {
				this.bigWinSnid = append(this.bigWinSnid, curPlayer.GetSnId())

				if curPlayer.IsRobot() {
					this.bigWinRobotSnid = append(this.bigWinRobotSnid, curPlayer.GetSnId())
				}
			}

		}
	}

	for _, player := range this.players {
		if player.IsRobot() && player.IsGameing() {
			if common.InSliceInt32(winnerSnid, player.GetSnId()) == false {
				this.loseRobotSnid = append(this.loseRobotSnid, player.GetSnId())
			}
		}
	}
}
func (this *DezhouPokerScene) OnDezhouPokerMiddleBilled(info *dezhoupoker.SCDezhouPokerGameBilledMiddle) {
	var winnerSnid []int32
	winnerPos := info.GetPos()
	curPlayer := this.seats[winnerPos]
	if curPlayer != nil {
		winnerSnid = append(winnerSnid, curPlayer.GetSnId())
		if curPlayer.IsRobot() {
			this.winRobotSnid = append(this.winRobotSnid, curPlayer.GetSnId())
		}

		if info.GetWinCoin() >= this.SCDezhouPokerRoomInfo.GetSmallBlind()*2*50 {
			this.bigWinSnid = append(this.bigWinSnid, curPlayer.GetSnId())
			if curPlayer.IsRobot() {
				this.bigWinRobotSnid = append(this.bigWinRobotSnid, curPlayer.GetSnId())
			}
		}
	}

	for _, player := range this.players {
		if player.IsRobot() && player.IsGameing() {
			if common.InSliceInt32(winnerSnid, player.GetSnId()) == false {
				this.loseRobotSnid = append(this.loseRobotSnid, player.GetSnId())
			}
		}
	}

	for _, player := range this.players {
		if player.IsRobot() && player.IsGameing() {
			if common.InSliceInt32(this.bigWinSnid, player.GetSnId()) == false {
				this.notBigWinRobotSnid = append(this.notBigWinRobotSnid, player.GetSnId())
			}
		}
	}

}

func (this *DezhouPokerScene) CreateGameCtx() *rule.GameCtx {
	commonCardCnt := 0
	state := int(this.GetState())
	if state >= rule.DezhouPokerSceneStateRiver {
		commonCardCnt = 5
	} else if state >= rule.DezhouPokerSceneStateTurn {
		commonCardCnt = 4
	} else if state >= rule.DezhouPokerSceneStateFlop {
		commonCardCnt = 3
	}
	poker := rule.NewPoker()
	ret := rule.GameCtx{
		CommonCard:    this.Cards[0:commonCardCnt],
		PlayerCards:   make([]*rule.PlayerCard, 0, len(this.players)),
		CommonCardCnt: commonCardCnt,
	}
	poker.DelCards(ret.CommonCard)

	for _, p := range this.players {
		if len(p.Cards) > 0 {
			if p.CanOp() {
				pc := &rule.PlayerCard{
					UserData: p.GetSnId(),
					HandCard: p.Cards,
				}
				ret.PlayerCards = append(ret.PlayerCards, pc)
			}
			poker.DelCards(p.Cards)
		}
	}

	ret.RestCard = poker.GetRestCard()

	return &ret
}

func (this *DezhouPokerScene) LeftAllRobot() bool {
	for _, p := range this.players {
		if len(p.Cards) > 0 && p.CanOp() && !p.IsRobot() {
			return false
		}
	}
	return true
}
func (this *DezhouPokerScene) SelectCard(s *netlib.Session) {
	if this.GetMe(s) != nil {
		pack := &dezhoupoker.CSDezhouPokerPlayerOp{
			OpCode: proto.Int32(rule.DezhouPokerPlayerOpSelectCard),
		}
		proto.SetDefaults(pack)
		base.DelaySend(s, int(dezhoupoker.DZPKPacketID_PACKET_CS_DEZHOUPOKER_OP), pack, 3, 9)
	}
}
