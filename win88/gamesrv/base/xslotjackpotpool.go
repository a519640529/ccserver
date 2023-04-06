package base

type XSlotJackpotPool struct {
	//PrizeFund   int64
	JackpotFund int64
}

func (this *XSlotJackpotPool) GetJackPotFund() int64 {
	return this.JackpotFund
}
