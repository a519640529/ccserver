package main

func (this *Player) GetVIPLevel(CoinPayTotal int64) int32 {

	platform := this.GetPlatform()
	if platform != nil {
		if this.ForceVip > 0 {
			if this.ForceVip > int32(len(platform.VipRange)) {
				this.ForceVip = int32(len(platform.VipRange))
			}
			return this.ForceVip
		}

		viplevel := int32(0)
		for i, dbvip := range platform.VipRange {
			if CoinPayTotal < int64(dbvip) {
				break
			}
			viplevel = int32(i + 1)
		}
		return viplevel
	}

	return 0
}

func (this *Player) GetMaxVIPLevel() int32 {

	platform := this.GetPlatform()
	if platform != nil {
		return int32(len(platform.VipRange))
	}

	return 0
}
