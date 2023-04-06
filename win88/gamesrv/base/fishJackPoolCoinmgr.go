package base

import (
	"time"

	"github.com/idealeak/goserver/core/module"
)

var FishJackpotCoinMgr = &FishJackpotCoin{
	Jackpot: make(map[string]*SlotJackpotPool),
}

type FishJackpotCoin struct {
	Jackpot map[string]*SlotJackpotPool //捕鱼奖池
}

func (this *FishJackpotCoin) ModuleName() string {
	return "FishJackpotCoinMgr"
}

func (this *FishJackpotCoin) Init() {
	/*if this.Jackpot == nil {
		this.Jackpot = &SlotJackpotPool{}
		str := SlotsPoolMgr.GetPool(int32(common.GameId_TFishing), this.platform) // 三个场次公用一个
		if str != "" {
			jackpot := &SlotJackpotPool{}
			err := json.Unmarshal([]byte(str), jackpot)
			if err == nil {
				this.Jackpot = jackpot
			}
		}
		if this.Jackpot != nil {
			if this.Jackpot.GetTotalBig() < 1 { // 初始值
				this.Jackpot.AddToBig(true, 888888)
			}
			SlotsPoolMgr.SetPool(int32(common.GameId_TFishing), this.platform, Jackpot)
		}
		logger.Logger.Info("FishJackpotCoin Init ", str, this.Jackpot)
	}*/
}

func (this *FishJackpotCoin) Update() {

}

func (this *FishJackpotCoin) Shutdown() {
	module.UnregisteModule(this)
}

func init() {
	module.RegisteModule(FishJackpotCoinMgr, time.Minute*5, 0)
}
