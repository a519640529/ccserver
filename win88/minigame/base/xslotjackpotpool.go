package base

type XSlotJackpotPool struct {
	//PrizeFund   int64
	JackpotFund []int64
}

func (this *XSlotJackpotPool) InitJackPot(n int) {
	this.JackpotFund = make([]int64, 0)
	for i := 0; i < n; i++ {
		this.JackpotFund = append(this.JackpotFund, 0)
	}
}
func (this *XSlotJackpotPool) GetJackPotFund() []int64 {
	return this.JackpotFund
}
