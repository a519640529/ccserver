package rollpoint

import (
	"games.yol.com/win88/protocol/player"
	"games.yol.com/win88/protocol/server"
	"games.yol.com/win88/robot/base"
	"math/rand"
	"time"

	"fmt"

	"games.yol.com/win88/common"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/rollpoint"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/netlib"
	"github.com/idealeak/goserver/core/timer"
)

//场景状态
const (
	RollPointSceneStateWait   int32 = iota //等待状态
	RollPointSceneStateStart               //开始
	RollPointSceneStateBet                 //押注
	RollPointSceneStateBilled              //结算
	RollPointSceneStateMax
)

var RollPointDelayTimeMax = 10

var RollPointSelRate = []int64{0, 5000, 1, 3000, 2, 1000, 3, 800, 4, 200}
var RollPointRate = []int32{0, 2000, 1, 500, 2, 500, 3, 500, 4, 500, 5, 600, 6, 700, 7, 800, 8, 900, 9, 1000}
var RollPointRateArea = [][]int32{{0, 1, 2, 3, 4, 5}, {22, 23, 24, 25, 26, 27}, {28}, {29, 30, 31, 32, 33, 34}, {6, 19}, {7, 18}, {8, 17}, {9, 16}, {10, 15}, {11, 12, 13, 14}}

type RollPointScene struct {
	base.BaseScene
	*rollpoint.SCRollPointRoomInfo
	Players    map[int32]base.Player
	dbGameFree *server.DB_GameFree
	BankerId   int32
}

func (this *RollPointScene) GetRoomId() int32                       { return this.GetSceneId() }
func (this *RollPointScene) GetRoomMode() int32                     { return this.SCRollPointRoomInfo.GetRoomMode() }
func (this *RollPointScene) GetGameId() int32                       { return this.SCRollPointRoomInfo.GetGameId() }
func (this *RollPointScene) GetPlayerByPos(pos int32) base.Player   { return nil }
func (this *RollPointScene) GetPlayerBySnid(snid int32) base.Player { return this.Players[snid] }
func (this *RollPointScene) DelPlayer(snid int32)                   { delete(this.Players, snid) }
func (this *RollPointScene) IsFull() bool                           { return false }
func (this *RollPointScene) IsMatchScene() bool                     { return false }
func (this *RollPointScene) AddPlayer(p base.Player) {
	p.Clear()
	this.Players[p.GetSnId()] = p
}
func (this *RollPointScene) Clear() {
	for _, player := range this.Players {
		player.Clear()
	}
}
func (this *RollPointScene) GetMe(s *netlib.Session) base.Player {
	if user, ok := s.GetAttribute(base.SessionAttributeUser).(*player.SCPlayerData); ok {
		player := this.GetPlayerBySnid(user.GetData().GetSnId())
		return player
	}
	return nil
}
func (this *RollPointScene) RollPoint(s *netlib.Session, me *RollPointPlayerData) {
	if base.StartSessionGameTimer(s, timer.TimerActionWrapper(func(h timer.TimerHandle, ud interface{}) bool {
		if this.GetState() != RollPointSceneStateBet {
			return true
		}
		if time.Now().Before(me.tNextBet) {
			return true
		}
		if time.Now().After(me.tEndBet) {
			return true
		}
		var coin int32
		params := this.dbGameFree.GetOtherIntParams()
		selCoinArr := []int64{}
		for i := 0; i < len(RollPointSelRate); i = i + 2 {
			if int64(params[i/2]) <= me.GetCoin() {
				selCoinArr = append(selCoinArr, RollPointSelRate[i], RollPointSelRate[i+1])
			}
		}
		if len(selCoinArr) == 0 {
			coinadd := params[len(params)-1]
			base.ExePMCmd(s, fmt.Sprintf("%v%v%v", common.PMCmd_AddCoin, common.PMCmd_SplitToken, coinadd))
			me.tNextBet = time.Now().Add(time.Second * 15)
			me.Coin = proto.Int64(me.GetCoin() + int64(coinadd))
			return true
		}
		coin = params[int(common.RandItemByWight(selCoinArr))]
		if rand.Int31n(100) < 30 {
			idx := me.selIndex[rand.Intn(len(me.selIndex))]
			logger.Trace(me.selIndex)
			maxBetCoin := this.dbGameFree.GetMaxBetCoin()
			if ok, _ := me.MaxChipCheck(int64(idx), int64(coin), maxBetCoin); ok {
				pack := &rollpoint.CSRollPointOp{
					OpCode: proto.Int32(0),
					Params: []int64{int64(idx), int64(coin)},
				}
				proto.SetDefaults(pack)
				s.Send(int(rollpoint.RPPACKETID_ROLLPOINT_CS_OP), pack)
				me.PushCoin += coin
				me.Coin = proto.Int64(me.GetCoin() - int64(coin))
			}
		}
		me.tNextBet = time.Now().Add(time.Duration(common.RandInt(300, 500)) * time.Millisecond)
		return true
	}), nil, time.Millisecond*100, -1) {
		me.tNextBet = time.Now().Add(time.Duration(200) * time.Millisecond)
		me.tEndBet = time.Now().Add(time.Duration(common.RandInt(5, 14)) * time.Second)
	}
}

func (this *RollPointScene) Update(ts int64) {}
