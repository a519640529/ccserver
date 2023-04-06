package tienlen

import (
	"fmt"
	tienlenApi "games.yol.com/win88/api3th/smart/tienlen"
	"games.yol.com/win88/common"
	rule "games.yol.com/win88/gamerule/tienlen"
	"games.yol.com/win88/gamesrv/base"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/tienlen"
	"github.com/cihub/seelog"
	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/logger"
	"sort"
	"strings"
	"time"
)

var tienlenlogger seelog.LoggerInterface

func init() {
	core.RegisteHook(core.HOOK_BEFORE_START, func() error {
		tienlenlogger = common.GetLoggerInstanceByName("TienLenLogger")
		return nil
	})
}

// 房间上的额外数据
type TienLenSceneData struct {
	*base.Scene                                             //场景
	players                    map[int32]*TienLenPlayerData //玩家信息
	seats                      [4]*TienLenPlayerData        //玩家
	poker                      *rule.Poker                  //扑克牌对象
	curGamingPlayerNum         int                          //当前在玩玩家数量
	lastGamingPlayerNum        int                          //上局参与游戏的人数
	lastOpPos                  int32                        //前一个出牌的玩家位置
	currOpPos                  int32                        //当前等待出牌的玩家位置
	lastBombPos                int32                        //上一个出炸弹的玩家位置
	curBombPos                 int32                        //当前出炸弹的玩家位置
	roundScore                 int32                        //小轮分
	startOpPos                 int32                        //开始操作的玩家
	curMinCard                 int32                        //当局最小牌值
	tianHuSnids                []int32                      //天胡玩家id
	winSnids                   []int32                      //赢家id
	lastWinSnid                int32                        //上局赢家id
	masterSnid                 int32                        //房主
	delCards                   [][]int32                    //已出牌
	delOrders                  []int32                      //出牌顺序
	isKongBomb                 bool                         //是否空放
	bombToEnd                  int                          //空放炸弹致底分翻倍次数
	card_play_action_seq       []string                     //出牌历史记录
	card_play_action_seq_int32 [][]int32                    //出牌历史记录
	lastPos                    int32                        //前一个玩家位置
	isAllRob                   bool                         //是否是纯AI场
}

func NewTienLenSceneData(s *base.Scene) *TienLenSceneData {
	sceneEx := &TienLenSceneData{
		Scene:   s,
		poker:   rule.NewPoker(),
		players: make(map[int32]*TienLenPlayerData),
	}
	sceneEx.Clear()
	return sceneEx
}

func (this *TienLenSceneData) init() bool {
	this.tianHuSnids = []int32{}
	this.winSnids = []int32{}
	this.roundScore = 0
	this.currOpPos = rule.InvalidePos
	this.lastOpPos = rule.InvalidePos
	this.curBombPos = rule.InvalidePos
	this.lastBombPos = rule.InvalidePos
	this.lastPos = rule.InvalidePos
	this.UnmarkPass()
	this.curGamingPlayerNum = 0
	this.curMinCard = 0
	this.delCards = [][]int32{}
	this.delOrders = []int32{}
	this.card_play_action_seq = []string{}
	this.card_play_action_seq_int32 = [][]int32{}
	this.bombToEnd = 0
	if this.GetPlayerNum() == 0 {
		this.SetPlayerNum(rule.MaxNumOfPlayer)
	}
	return true
}

func (this *TienLenSceneData) Clear() {
	this.tianHuSnids = []int32{}
	this.winSnids = []int32{}
	this.roundScore = 0
	this.currOpPos = rule.InvalidePos
	this.lastOpPos = rule.InvalidePos
	this.curBombPos = rule.InvalidePos
	this.lastBombPos = rule.InvalidePos
	this.lastPos = rule.InvalidePos
	this.curGamingPlayerNum = 0
	this.curMinCard = 0
	this.UnmarkPass()
	this.delCards = [][]int32{}
	this.delOrders = []int32{}
	this.card_play_action_seq = []string{}
	this.card_play_action_seq_int32 = [][]int32{}
	this.bombToEnd = 0
	this.isAllRob = false
	for _, player := range this.players {
		if player != nil {
			player.UnmarkFlag(base.PlayerState_WaitNext)
		}
	}
}

func (this *TienLenSceneData) CanStart() bool {
	if this.GetGaming() == true {
		return false
	}

	nPlayerCount := this.GetPlayerCnt()
	nRobotCount := this.GetRobotCnt()
	if this.IsMatchScene() {
		if nRobotCount == 4 {
			this.isAllRob = true
		}
		if nPlayerCount == 4 {
			return true
		} else {
			return false
		}
	}
	if nPlayerCount >= 2 && (nRobotCount < nPlayerCount || this.IsPreCreateScene()) { //人数>=2开始
		return true
	}
	return false
}

func (this *TienLenSceneData) GetPlayerCnt() int {
	var cnt int
	for i := 0; i < this.GetPlayerNum(); i++ {
		playerEx := this.seats[i]
		if playerEx == nil {
			continue
		}
		cnt++
	}
	return cnt
}
func (this *TienLenSceneData) GetGameingPlayerCnt() int {
	var cnt int
	for i := 0; i < this.GetPlayerNum(); i++ {
		playerEx := this.seats[i]
		if playerEx == nil {
			continue
		}
		if playerEx.IsGameing() {
			cnt++
		}
	}
	return cnt
}

func (this *TienLenSceneData) GetRobotCnt() int {
	var cnt int
	for i := 0; i < this.GetPlayerNum(); i++ {
		playerEx := this.seats[i]
		if playerEx == nil {
			continue
		}
		if playerEx.IsRob {
			cnt++
		}
	}
	return cnt
}

func (this *TienLenSceneData) GetGameingRobotCnt() int {
	var cnt int
	for i := 0; i < this.GetPlayerNum(); i++ {
		playerEx := this.seats[i]
		if playerEx == nil {
			continue
		}
		if playerEx.IsRob && playerEx.IsGameing() {
			cnt++
		}
	}
	return cnt
}

func (this *TienLenSceneData) GetSeatPlayerCnt() int {
	var cnt int
	for i := 0; i < this.GetPlayerNum(); i++ {
		playerEx := this.seats[i]
		if playerEx == nil {
			continue
		}
		cnt++
	}
	return cnt
}

func (this *TienLenSceneData) BroadcastPlayerLeave(p *base.Player, reason int) {
	scLeavePack := &tienlen.SCTienLenPlayerLeave{
		Pos: proto.Int(p.GetPos()),
	}
	proto.SetDefaults(scLeavePack)
	this.Broadcast(int(tienlen.TienLenPacketID_PACKET_SCTienLenPlayerLeave), scLeavePack, p.GetSid())
}

func (this *TienLenSceneData) BroadcastAudienceNum(p *base.Player) {
	pack := &tienlen.SCTienLenUpdateAudienceNum{
		AudienceNum: proto.Int(this.GetAudiencesNum()),
	}
	proto.SetDefaults(pack)
	this.Broadcast(int(tienlen.TienLenPacketID_PACKET_SCTienLenUpdateAudienceNum), pack, p.GetSid())
}

func (this *TienLenSceneData) delPlayer(p *base.Player) {
	if p, exist := this.players[p.SnId]; exist {
		this.seats[p.GetPos()] = nil
		delete(this.players, p.SnId)
	}
}
func (this *TienLenSceneData) OnPlayerLeave(p *base.Player, reason int) {
	this.delPlayer(p)
	this.BroadcastPlayerLeave(p, reason)
}

func (this *TienLenSceneData) SceneDestroy(force bool) {
	//销毁房间
	this.Scene.Destroy(force)
}

// 广播房主更换
func (this *TienLenSceneData) BroadcastUpdateMasterSnid(changeState bool) {
	pack := &tienlen.SCTienLenUpdateMasterSnid{
		MasterSnid: proto.Int32(this.masterSnid),
	}
	proto.SetDefaults(pack)
	this.Broadcast(int(tienlen.TienLenPacketID_PACKET_SCTienLenUpdateMasterSnid), pack, 0)
	logger.Logger.Trace("广播房主更换 ", pack)

	//重置开始倒计时
	if changeState && this.CanStart() {
		this.Scene.ChangeSceneState(rule.TienLenSceneStateWaitStart)
	}
}

// 广播当前操作的玩家位置
func (this *TienLenSceneData) BroadcastOpPos() {

	if this.currOpPos != rule.InvalidePos && this.seats[this.currOpPos] != nil && this.seats[this.currOpPos].IsRobot() {
		//计算其它人手牌，所有牌-出过的牌-自己手牌
		othercards := []int32{}
		handcards := []int32{}
		for i := int32(0); i < 52; i++ {
			othercards = append(othercards, i)
		}
		card_play_action_seqs := []string{}
		for _, cards := range this.card_play_action_seq_int32 {
			//转棋牌到AI中的牌
			aicards := []int32{}
			for _, card := range cards {
				aicards = append(aicards, tienlenApi.CardToAiCard[card])
			}

			card_play_action_seqs = append(card_play_action_seqs, common.Int32SliceToString(aicards, ","))
			othercards = common.DelSliceIn32s(othercards, aicards)
		}
		//删除手牌
		for _, v := range this.seats[this.currOpPos].cards {
			if v != -1 {
				aicard := tienlenApi.CardToAiCard[v]
				othercards = common.DelSliceInt32(othercards, aicard)
				handcards = append(handcards, aicard)
			}
		}
		//logger.Logger.Infof("%v出牌历史记录：%v",this.currOpPos,this.card_play_action_seq)
		//logger.Logger.Infof("%v出牌历史记录：%v",this.currOpPos,common.StringSliceToString(card_play_action_seqs,"|"))
		//logger.Logger.Infof("%v手牌：%v 数量：%v",this.currOpPos,handcards,len(handcards))
		//logger.Logger.Infof("%v其它人牌：%v 数量：%v",this.currOpPos,othercards,len(othercards))

		lastmove := make(map[int][]int32)
		numCardsLeft := make(map[int]int)
		playedCards := make(map[int][]int32)
		for pos, v := range this.seats {
			if v != nil {
				//出过的牌
				delcards := []int32{}
				for _, cards := range v.delCards {
					for _, card := range cards {
						aicard := tienlenApi.CardToAiCard[card]
						delcards = append(delcards, aicard)
					}
				}
				//logger.Logger.Infof("%v出过的牌：%v",pos,delcards)
				//logger.Logger.Infof("%v剩余的牌：%v",pos,13-len(delcards))
				playedCards[pos] = delcards
				numCardsLeft[pos] = 13 - len(delcards)
				last := len(v.delCards)
				if last > 0 {
					aicards := []int32{}
					for _, card := range v.delCards[last-1] {
						aicards = append(aicards, tienlenApi.CardToAiCard[card])
					}
					//logger.Logger.Infof("%v最后一次出牌：%v",pos,aicards)
					lastmove[pos] = aicards
				} else {
					//logger.Logger.Infof("%v最后一次出牌：%v",pos,"")
					lastmove[pos] = []int32{}
				}

			} else {
				numCardsLeft[pos] = 13
			}
		}
		pack := &tienlen.SCTienLenAIData{
			BombNum:           0, //炸弹数量
			CardPlayActionSeq: proto.String(strings.Replace(common.StringSliceToString(card_play_action_seqs, "|"), "-1", "", -1)),
			LastMove_0:        proto.String(common.Int32SliceToString(lastmove[0], ",")),
			LastMove_1:        proto.String(common.Int32SliceToString(lastmove[1], ",")),
			LastMove_2:        proto.String(common.Int32SliceToString(lastmove[2], ",")),
			LastMove_3:        proto.String(common.Int32SliceToString(lastmove[3], ",")),
			NumCardsLeft_0:    proto.Int32(int32(numCardsLeft[0])),
			NumCardsLeft_1:    proto.Int32(int32(numCardsLeft[1])),
			NumCardsLeft_2:    proto.Int32(int32(numCardsLeft[2])),
			NumCardsLeft_3:    proto.Int32(int32(numCardsLeft[3])),
			OtherHandCards:    proto.String(common.Int32SliceToString(othercards, ",")),
			PlayedCards_0:     proto.String(common.Int32SliceToString(playedCards[0], ",")),
			PlayedCards_1:     proto.String(common.Int32SliceToString(playedCards[1], ",")),
			PlayedCards_2:     proto.String(common.Int32SliceToString(playedCards[2], ",")),
			PlayedCards_3:     proto.String(common.Int32SliceToString(playedCards[3], ",")),
			PlayerHandCards:   proto.String(common.Int32SliceToString(handcards, ",")),
			PlayerPosition:    proto.Int32(this.currOpPos),
		}
		proto.SetDefaults(pack)
		this.seats[this.currOpPos].SendToClient(int(tienlen.TienLenPacketID_PACKET_SCTienLenAI), pack)
		//logger.Logger.Infof("Send Robot AI Data:%v", pack)
	}

	pack := &tienlen.SCTienLenCurOpPos{
		Pos:   proto.Int32(this.currOpPos),
		IsNew: proto.Bool(this.currOpPos == this.lastOpPos || this.lastOpPos == rule.InvalidePos),
	}
	lastOpPlayer := this.GetLastOpPlayer()
	if lastOpPlayer != nil {
		if len(lastOpPlayer.delCards) > 0 {
			lastDelCards := lastOpPlayer.delCards[len(lastOpPlayer.delCards)-1]
			for _, card := range lastDelCards {
				if card != rule.InvalideCard {
					pack.Cards = append(pack.Cards, card)
				}
			}
		}
	}

	if this.lastPos != rule.InvalidePos && this.lastOpPos == this.lastPos {
		// 特殊牌型才去添加额外延迟时间
		if rule.NeedExDelay(pack.Cards) {
			pack.ExDelay = proto.Int32(int32(rule.DelayCanOp))
		}
	}

	proto.SetDefaults(pack)
	this.Broadcast(int(tienlen.TienLenPacketID_PACKET_SCTienLenCurOpPos), pack, 0)
	logger.Logger.Trace("(this *TienLenSceneData) BroadcastOpPos TienLenPacketID_PACKET_SCTienLenCurOpPos", this.GetSceneId(), ";pack:", pack)
}

// 出牌
func (this *TienLenSceneData) DelCards(playerEx *TienLenPlayerData, cards []int32) bool {
	if playerEx != nil && len(cards) != 0 {
		for _, delCard := range cards {
			for i, hcard := range playerEx.cards {
				if delCard == hcard && hcard != rule.InvalideCard {
					playerEx.cards[i] = rule.InvalideCard
					continue
				}
			}
		}
		sort.Slice(cards, func(i, j int) bool {
			v_i := rule.Value(cards[i])
			v_j := rule.Value(cards[j])
			c_i := rule.Color(cards[i])
			c_j := rule.Color(cards[j])
			if v_i > v_j {
				return false
			} else if v_i == v_j {
				return c_i < c_j
			}
			return true
		})
		playerEx.delCards = append(playerEx.delCards, cards)
		this.delCards = append(this.delCards, cards)
		this.delOrders = append(this.delOrders, playerEx.SnId)
		this.card_play_action_seq = append(this.card_play_action_seq, fmt.Sprintf("%v-%v", playerEx.GetPos(), common.Int32SliceToString(cards, ",")))
		this.card_play_action_seq_int32 = append(this.card_play_action_seq_int32, cards)
		return true
	}
	return false
}

func (this *TienLenSceneData) PlayerCanOp(pos int32) bool {
	if pos < 0 || pos >= int32(this.GetPlayerNum()) {
		return false
	}
	if this.seats[pos] != nil && this.seats[pos].CanOp() {
		return true
	}
	this.card_play_action_seq = append(this.card_play_action_seq, fmt.Sprintf("%v-过", pos))
	this.card_play_action_seq_int32 = append(this.card_play_action_seq_int32, []int32{-1})
	return false
}

func (this *TienLenSceneData) AllPlayerEnterGame() {

	for i := 0; i < this.GetPlayerNum(); i++ {
		if this.seats[i] == nil {
			continue
		}

		if this.seats[i].IsMarkFlag(base.PlayerState_GameBreak) == false {
			this.seats[i].UnmarkFlag(base.PlayerState_WaitNext)
			this.seats[i].SyncFlag()
		}
	}
}

// 娱乐版
func (this *TienLenSceneData) IsTienLenYule() bool {
	return this.GetDBGameFree().GetGameId() == common.GameId_TienLen_yl || this.GetDBGameFree().GetGameId() == common.GameId_TienLen_yl_toend
}

// 打到底
func (this *TienLenSceneData) IsTienLenToEnd() bool {
	return this.GetDBGameFree().GetGameId() == common.GameId_TienLen_toend ||
		this.GetDBGameFree().GetGameId() == common.GameId_TienLen_yl_toend ||
		this.GetDBGameFree().GetGameId() == common.GameId_TienLen_yl_toend_m
}
func (this *TienLenSceneData) GetFreeGameSceneType() int32 {
	return int32(this.SceneType)
}

func (this *TienLenSceneData) SendHandCard() {
	this.poker.Shuffle()
	buf := this.poker.GetPokerBuf()

	//牌序- 2, A, K, Q, J, 10, 9, 8, 7, 6, 5, 4, 3
	//红桃- 51,50,49,48,47,46,45,44,43,42,41,40,39
	//方片- 38,37,36,35,34,33,32,31,30,29,28,27,26
	//梅花- 25,24,23,22,21,20,19,18,17,16,15,14,13
	//黑桃- 12,11,10, 9, 8, 7, 6, 5, 4, 3, 2, 1, 0
	test1 := []int32{40, 27, 14, 1}
	test2 := []int32{47, 34, 33, 20, 19, 6}
	test3 := []int32{5, 4, 3, 2, 18, 17, 16, 15}
	test4 := []int32{46, 45, 44, 43, 42, 7, 32, 31, 30, 29}

	need1 := rule.HandCardNum - int32(len(test1))
	need2 := rule.HandCardNum - int32(len(test2))
	need3 := rule.HandCardNum - int32(len(test3))
	need4 := rule.HandCardNum - int32(len(test4))

	tmpBuf := []int32{}

	if rule.TestOpen {
		for i, card := range buf {
			for _, card1 := range test1 {
				if int32(card) == card1 {
					buf[i] = rule.Card(rule.InvalideCard)
				}
			}
			for _, card1 := range test2 {
				if int32(card) == card1 {
					buf[i] = rule.Card(rule.InvalideCard)
				}
			}
			for _, card1 := range test3 {
				if int32(card) == card1 {
					buf[i] = rule.Card(rule.InvalideCard)
				}
			}
			for _, card1 := range test4 {
				if int32(card) == card1 {
					buf[i] = rule.Card(rule.InvalideCard)
				}
			}
		}

		for _, card := range buf {
			if int32(card) != rule.InvalideCard {
				tmpBuf = append(tmpBuf, int32(card))
			}
		}
	}

	var n int
	minCard := int32(999)
	for _, seatPlayerEx := range this.seats {
		if seatPlayerEx != nil {
			bb := []int32{}
			for i := n * rule.Hand_CardNum; i < (n+1)*rule.Hand_CardNum; i++ {
				bb = append(bb, int32(buf[i]))
			}
			if rule.TestOpen {
				bb = []int32{}
				switch seatPlayerEx.GetPos() {
				case 0:
					for _, card := range test1 {
						bb = append(bb, card)
					}
					for i := int32(0); i < need1; i++ {
						bb = append(bb, tmpBuf[i])
					}
					tmpBuf = append(tmpBuf[need1:])
				case 1:
					for _, card := range test2 {
						bb = append(bb, card)
					}
					for i := int32(0); i < need2; i++ {
						bb = append(bb, tmpBuf[i])
					}
					tmpBuf = append(tmpBuf[need2:])
				case 2:
					for _, card := range test3 {
						bb = append(bb, card)
					}
					for i := int32(0); i < need3; i++ {
						bb = append(bb, tmpBuf[i])
					}
					tmpBuf = append(tmpBuf[need3:])
				case 3:
					for _, card := range test4 {
						bb = append(bb, card)
					}
					for i := int32(0); i < need4; i++ {
						bb = append(bb, tmpBuf[i])
					}
					tmpBuf = append(tmpBuf[need4:])
				}
			}

			//排下序，正常应该客户端排序
			sort.Slice(bb, func(i, j int) bool {
				v_i := rule.Value(int32(bb[i]))
				v_j := rule.Value(int32(bb[j]))
				c_i := rule.Color(int32(bb[i]))
				c_j := rule.Color(int32(bb[j]))
				if v_i > v_j {
					return false
				} else if v_i == v_j {
					return c_i < c_j
				}
				return true
			})
			for idx, card := range bb {
				seatPlayerEx.cards[idx] = int32(card)
				if rule.Value(int32(card)) < rule.Value(minCard) {
					this.startOpPos = int32(seatPlayerEx.GetPos())
					minCard = int32(card)
					this.curMinCard = minCard
				} else if rule.Value(int32(card)) == rule.Value(minCard) {
					if rule.Color(int32(card)) < rule.Color(minCard) {
						this.startOpPos = int32(seatPlayerEx.GetPos())
						minCard = int32(card)
						this.curMinCard = minCard
					}
				}
			}

			pack := &tienlen.SCTienLenCard{}
			for j := int32(0); j < rule.HandCardNum; j++ {
				pack.Cards = append(pack.Cards, int32(seatPlayerEx.cards[j]))
			}
			proto.SetDefaults(pack)
			seatPlayerEx.SendToClient(int(tienlen.TienLenPacketID_PACKET_SCTienLenCard), pack)
			logger.Logger.Trace("player_id", seatPlayerEx.SnId, ";SCTienLenCard", pack.Cards)
			n++
		}
	}
}

func (this *TienLenSceneData) SetCurOpPos(pos int32) {
	this.currOpPos = pos
}
func (this *TienLenSceneData) GetCurOpPos() int32 {
	return this.currOpPos
}
func (this *TienLenSceneData) GetCurOpPlayer() *TienLenPlayerData {
	if this.currOpPos < 0 || this.currOpPos >= int32(this.GetPlayerNum()) {
		return nil
	}
	return this.seats[this.currOpPos]
}
func (this *TienLenSceneData) SetLastOpPos(pos int32) {
	this.lastOpPos = pos
}
func (this *TienLenSceneData) GetLastOpPos() int32 {
	return this.lastOpPos
}
func (this *TienLenSceneData) GetLastOpPlayer() *TienLenPlayerData {
	if this.lastOpPos < 0 || this.lastOpPos >= int32(this.GetPlayerNum()) {
		return nil
	}
	return this.seats[this.lastOpPos]
}

func (this *TienLenSceneData) GetLastBombPlayer() *TienLenPlayerData {
	if this.lastBombPos < 0 || this.lastBombPos >= int32(this.GetPlayerNum()) {
		return nil
	}
	return this.seats[this.lastBombPos]
}
func (this *TienLenSceneData) GetCurBombPlayer() *TienLenPlayerData {
	if this.curBombPos < 0 || this.curBombPos >= int32(this.GetPlayerNum()) {
		return nil
	}
	return this.seats[this.curBombPos]
}

// 逆时针找一个空位
func (this *TienLenSceneData) FindOnePos() int {
	for i := 0; i < this.GetPlayerNum(); i++ {
		if this.seats[i] == nil {
			return i
		}
	}
	return int(rule.InvalidePos)
}

func (this *TienLenSceneData) DoNext(pos int32) int32 {
	nextPos := this.GetNextOpPos(pos)
	if nextPos != rule.InvalidePos {
		this.SetCurOpPos(nextPos)
		this.StateStartTime = time.Now()
	}
	this.lastPos = pos
	return nextPos
}

func (this *TienLenSceneData) GetNextOpPos(pos int32) int32 {
	if pos == rule.InvalidePos {
		return rule.InvalidePos
	}
	if pos < 0 || pos >= int32(this.GetPlayerNum()) {
		return rule.InvalidePos
	}

	for i := pos + 1; i < int32(this.GetPlayerNum()); i++ {
		if this.PlayerCanOp(i) {
			return i
		}
	}
	for i := int32(0); i < pos; i++ {
		if this.PlayerCanOp(i) {
			return i
		}
	}

	return rule.InvalidePos
}

func (this *TienLenSceneData) UnmarkPass() {
	for i := 0; i < this.GetPlayerNum(); i++ {
		if this.seats[i] != nil {
			this.seats[i].isPass = false
		}
	}
}

func (this *TienLenSceneData) FindWinPos() int {
	winPos := -1
	if this.lastGamingPlayerNum != 0 && this.curGamingPlayerNum != 0 && this.lastWinSnid != 0 {
		haveLastWinPos := -1
		for i := 0; i < this.GetPlayerNum(); i++ {
			if this.seats[i] != nil {
				if this.seats[i].SnId == this.lastWinSnid {
					haveLastWinPos = i
					break
				}
			}
		}
		if haveLastWinPos != -1 {
			if this.lastGamingPlayerNum > 2 {
				winPos = haveLastWinPos
			} else if this.lastGamingPlayerNum == 2 {
				if this.curGamingPlayerNum == 2 {
					winPos = haveLastWinPos
				}
			}
		}
	}
	return winPos
}

func (this *TienLenSceneData) TrySmallGameBilled() {
	// todo 看是不是炸弹，是炸弹结算分
	if this.isKongBomb {
		this.bombToEnd++
		this.isKongBomb = false
	}
	if this.roundScore > 0 && this.curBombPos != rule.InvalidePos && this.lastBombPos != rule.InvalidePos {
		winPlayer := this.GetCurBombPlayer()
		losePlayer := this.GetLastBombPlayer()
		baseScore := this.BaseScore
		score := int64(this.roundScore) * int64(baseScore)
		if this.IsTienLenToEnd() {
			score = int64(this.roundScore) * int64(baseScore) / 100 //百分比
		}
		losePlayerCoin := losePlayer.GetCoin()
		if losePlayerCoin < score { //输完
			score = losePlayerCoin
		}
		if score != 0 {
			taxRate := this.DbGameFree.GetTaxRate()                               //万分比
			gainScore := int64(float64(score) * float64(10000-taxRate) / 10000.0) //税后
			bombTaxScore := score - gainScore
			// win
			winPlayer.AddCoin(gainScore, common.GainWay_CoinSceneWin, 0, "system", this.GetSceneName())
			winPlayer.winCoin += gainScore
			winPlayer.bombScore += gainScore
			winPlayer.bombTaxScore += bombTaxScore
			//lose
			losePlayer.AddCoin(-score, common.GainWay_CoinSceneLost, 0, "system", this.GetSceneName())
			losePlayer.winCoin -= score
			losePlayer.bombScore -= score

			pack := &tienlen.SCTienLenSmallGameBilled{
				WinPos:      proto.Int(winPlayer.GetPos()),
				WinPosCoin:  proto.Int64(winPlayer.GetCoin()),
				WinCoin:     proto.Int64(gainScore),
				LosePos:     proto.Int(losePlayer.GetPos()),
				LosePosCoin: proto.Int64(losePlayer.GetCoin()),
				LoseCoin:    proto.Int64(score),
			}
			proto.SetDefaults(pack)
			this.Broadcast(int(tienlen.TienLenPacketID_PACKET_SCTienLenSmallGameBilled), pack, 0)
			logger.Logger.Trace("SCTienLenSmallGameBilled: ", pack)
		}
	}
	this.curBombPos = rule.InvalidePos
	this.lastBombPos = rule.InvalidePos
	this.roundScore = 0
	this.UnmarkPass()
}

func (this *TienLenSceneData) IsTianhuPlayer(snid int32) bool {
	for _, tianhusnid := range this.tianHuSnids {
		if snid == tianhusnid {
			return true
		}
	}
	return false
}

func (this *TienLenSceneData) IsWinPlayer(snid int32) bool {
	for _, winSnid := range this.winSnids {
		if snid == winSnid {
			return true
		}
	}
	return false
}

func (this *TienLenSceneData) SystemCoinOut() int64 {
	systemGain := int64(0)
	for i := 0; i < this.GetPlayerNum(); i++ {
		playerData := this.seats[i]

		if playerData != nil && playerData.IsGameing() && playerData.IsRob {
			systemGain += playerData.winCoin
		}
	}
	return systemGain
}
