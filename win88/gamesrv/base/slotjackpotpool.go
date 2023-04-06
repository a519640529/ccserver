package base

type SlotJackpotPool struct {
	Grand  int64 //玩家巨奖池
	Big    int64 //玩家大奖池
	Middle int64 //玩家中奖池
	Small  int64 //玩家小奖池

	GrandRob     int64   //机器人巨奖池
	BigRob       int64   //机器人大奖池
	MiddleRob    int64   //机器人中奖池
	SmallRob     int64   //机器人小奖池
	Normal       []int64 //玩家普通奖池
	NormalRob    []int64 //机器人普通奖池
	SysAllOutput int64
	BetTotal     int64
	//金鼓齐鸣 等游戏奖池根据玩家下注金额而变化
	SmallPools   []int64 //小奖固定额度
	MiddlePools  []int64
	BigPools     []int64
	HugePools    []int64
	BigRobPools  []int64
	HugeRobPools []int64
}

func (this *SlotJackpotPool) GetTotalBig() int64 {
	return this.Big + this.BigRob
}

func (this *SlotJackpotPool) GetTotalGrand() int64 {
	return this.Grand + this.GrandRob
}

func (this *SlotJackpotPool) GetTotalMiddle() int64 {
	return this.Middle + this.MiddleRob
}

func (this *SlotJackpotPool) GetTotalSmall() int64 {
	return this.Small + this.SmallRob
}
func (this *SlotJackpotPool) GetNormal(idx int) int64 {
	return this.Normal[idx]
}

func (this *SlotJackpotPool) AddToGrand(isRob bool, coin int64) {
	if isRob {
		this.GrandRob += coin
	} else {
		this.Grand += coin
	}
}

func (this *SlotJackpotPool) AddToBig(isRob bool, coin int64) {
	if isRob {
		this.BigRob += coin
	} else {
		this.Big += coin
	}
}

func (this *SlotJackpotPool) AddToMiddle(isRob bool, coin int64) {
	if isRob {
		this.MiddleRob += coin
	} else {
		this.Middle += coin
	}
}

func (this *SlotJackpotPool) AddToSmall(isRob bool, coin int64) {
	if isRob {
		this.SmallRob += coin
	} else {
		this.Small += coin
	}
}

func (this *SlotJackpotPool) NormalIsEnough(isRob bool, idx int, coin int64) bool {
	if isRob {
		if idx >= 0 && idx < len(this.NormalRob) {
			return this.NormalRob[idx] >= coin
		}
	} else {
		if idx >= 0 && idx < len(this.Normal) {
			return this.Normal[idx] >= coin
		}
	}
	return false
}

func (this *SlotJackpotPool) AddToNormal(isRob bool, idx int, coin int64) {
	if isRob {
		if idx >= 0 && idx < len(this.NormalRob) {
			this.NormalRob[idx] += coin
		}
	} else {
		if idx >= 0 && idx < len(this.Normal) {
			this.Normal[idx] += coin
		}
	}
}

func (this *SlotJackpotPool) AddToBetTotal(isRob bool, coin int64) {
	if !isRob {
		this.BetTotal += coin
	}
}

func (this *SlotJackpotPool) AddToSystemOut(isRob bool, coin int64) {
	if !isRob {
		this.SysAllOutput += coin
	}
}

func (this *SlotJackpotPool) AddToBigPoolByIdx(isRob bool, idx int32, coin int64) {
	//logger.Logger.Tracef("向大奖池放入金币 ===>%d", coin)
	if isRob {
		this.BigRobPools[idx] += coin
	} else {
		this.BigPools[idx] += coin
	}
}

func (this *SlotJackpotPool) AddToHugePoolByIdx(isRob bool, idx int32, coin int64) {
	//logger.Logger.Tracef("向巨奖池放入金币 ===>%d", coin)
	if isRob {
		this.HugeRobPools[idx] += coin
	} else {
		this.HugePools[idx] += coin
	}
}

func (this *SlotJackpotPool) AddToPoolsByIdx(isRob bool, idx int32, coin int64) {
	this.AddToBigPoolByIdx(isRob, idx, coin)
	this.AddToHugePoolByIdx(isRob, idx, coin)
}

func (this *SlotJackpotPool) GetSmallPools(idx int) int64 {
	return this.SmallPools[idx]
}
func (this *SlotJackpotPool) GetMiddlePools(idx int) int64 {
	return this.MiddlePools[idx]
}

func (this *SlotJackpotPool) GetBigPools(isRob bool, idx int) int64 {
	if isRob {
		return this.BigRobPools[idx]
	} else {
		return this.BigPools[idx]
	}
}
func (this *SlotJackpotPool) GetHugePools(isRob bool, idx int) int64 {
	if isRob {
		return this.HugeRobPools[idx]
	} else {
		return this.HugePools[idx]
	}
}
