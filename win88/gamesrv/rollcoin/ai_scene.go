package rollcoin

import (
	"games.yol.com/win88/gamerule/rollcoin"
	"games.yol.com/win88/gamesrv/base"
	"math/rand"
	"time"
)

var RollCoinSelRate = []int64{0, 51, 1, 26, 2, 18, 3, 9, 4, 1}
var RollCoinSelFlag = []int64{0, 100, 1, 100, 2, 100, 3, 100, 4, 100, 5, 100, 6, 100, 7, 100}

type RollCoinSceneAIMgr struct {
	base.HundredSceneAIMgr
	bankerList map[int32]int64
}

//房间状态变化
func (this *RollCoinSceneAIMgr) OnChangeState(s *base.Scene, oldstate, newstate int) {
	this.HundredSceneAIMgr.OnChangeState(s, oldstate, newstate)
	switch newstate {
	case rollcoin.RollCoinSceneStateBilled:
		this.UpBanker(s)
	}
}

//房间心跳
func (this *RollCoinSceneAIMgr) OnTick(s *base.Scene) {
	this.HundredSceneAIMgr.OnTick(s)
	if len(this.bankerList) > 0 {
		if sceneEx, ok := s.ExtraData.(*RollCoinSceneData); ok {
			for snid, ts := range this.bankerList {
				if ts >= time.Now().Unix() {
					p := sceneEx.players[snid]
					if p != nil {
						s.GetScenePolicy().OnPlayerOp(s, p.Player, rollcoin.RollCoinPlayerOpBanker, nil)
						delete(this.bankerList, snid)
						break
					}
				}
			}
		}
	}
}
func (this *RollCoinSceneAIMgr) UpBanker(s *base.Scene) {
	if sceneEx, ok := s.ExtraData.(*RollCoinSceneData); ok {
		if sceneEx.DbGameFree.GetBanker() == 0 {
			return
		}
		if this.bankerList == nil {
			this.bankerList = make(map[int32]int64)
		}
		if len(sceneEx.BankerList) >= 5 {
			return
		}
		for _, p := range sceneEx.players {
			if !p.IsRob {
				continue
			}
			if p.Coin < int64(sceneEx.BankCoinLimit) {
				continue
			}
			n := rand.Intn(3) + 1
			if len(this.bankerList) > n {
				break
			}

			bankRate := 100 / (sceneEx.DbGameFree.GetSceneType())                //1,2,3,4->100,50,33,25
			bankRate = int32(float32(bankRate) * float32(float32(bankRate)/100)) ///100,25,10,8
			if rand.Int31n(100) > bankRate {
				return
			}
			if _, ok := this.bankerList[p.SnId]; !ok && len(this.bankerList) < 3 {
				this.bankerList[p.GetSnId()] = time.Now().Add(time.Duration(rand.Intn(20)) * time.Second).Unix()
			}
		}
	}
}
