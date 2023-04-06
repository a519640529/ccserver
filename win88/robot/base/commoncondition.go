package base

import (
	b3 "github.com/magicsea/behavior3go"
	b3config "github.com/magicsea/behavior3go/config"
	b3core "github.com/magicsea/behavior3go/core"
)

//CheckBool
type CheckBool struct {
	b3core.Condition
	keyName string
}

func (this *CheckBool) Initialize(setting *b3config.BTNodeCfg) {
	this.Condition.Initialize(setting)
	this.keyName = setting.GetPropertyAsString("keyName")
}

func (this *CheckBool) OnTick(tick *b3core.Tick) b3.Status {
	var b = tick.Blackboard.GetBool(this.keyName, "", "")
	if b {
		return b3.SUCCESS
	}
	return b3.FAILURE
}

//CheckInt
type CheckInt struct {
	b3core.Condition
	keyName string
	value   int
	cmp     int
}

func (this *CheckInt) Initialize(setting *b3config.BTNodeCfg) {
	this.Condition.Initialize(setting)
	this.keyName = setting.GetPropertyAsString("keyName")
	this.value = setting.GetPropertyAsInt("value")
	this.cmp = setting.GetPropertyAsInt("cmp")
}

func (this *CheckInt) OnTick(tick *b3core.Tick) b3.Status {
	var b = tick.Blackboard.GetInt(this.keyName, "", "")
	if CmpInt(b, this.value, this.cmp) {
		return b3.SUCCESS
	}
	return b3.FAILURE
}

//player coin
type CheckPlayerCoin struct {
	b3core.Condition
	gIPCoin string
	cmp     int
}

func (this *CheckPlayerCoin) Initialize(setting *b3config.BTNodeCfg) {
	this.Condition.Initialize(setting)
	this.gIPCoin = setting.GetPropertyAsString("gIPCoin")
	this.cmp = setting.GetPropertyAsInt("cmp")

}

func (this *CheckPlayerCoin) OnTick(tick *b3core.Tick) b3.Status {
	f := tick.GetTarget().(Player)

	coin := getBoardIntByPreStr(tick.Blackboard, this.gIPCoin, "", "")
	if CmpInt(int(f.GetCoin()), coin, this.cmp) {
		return b3.SUCCESS
	}
	return b3.FAILURE
}

//player gameNum
type CheckPlayerGameNum struct {
	b3core.Condition
	gIPGameNum string
	cmp        int
}

func (this *CheckPlayerGameNum) Initialize(setting *b3config.BTNodeCfg) {
	this.Condition.Initialize(setting)
	this.gIPGameNum = setting.GetPropertyAsString("gIPGameNum")
	this.cmp = setting.GetPropertyAsInt("cmp")

}

func (this *CheckPlayerGameNum) OnTick(tick *b3core.Tick) b3.Status {
	f := tick.GetTarget().(Player)

	coin := getBoardIntByPreStr(tick.Blackboard, this.gIPGameNum, "", "")
	if CmpInt(int(f.GetGameCount()), coin, this.cmp) {
		return b3.SUCCESS
	}
	return b3.FAILURE
}

//player checkLastWinLost
type CheckPlayerLastWinOrLost struct {
	b3core.Condition
	gIPLResult string
	cmp        int
}

func (this *CheckPlayerLastWinOrLost) Initialize(setting *b3config.BTNodeCfg) {
	this.Condition.Initialize(setting)
	this.gIPLResult = setting.GetPropertyAsString("gIPLResult")
	this.cmp = setting.GetPropertyAsInt("cmp")

}

func (this *CheckPlayerLastWinOrLost) OnTick(tick *b3core.Tick) b3.Status {
	f := tick.GetTarget().(Player)

	lr := getBoardIntByPreStr(tick.Blackboard, this.gIPLResult, "", "")
	if CmpInt(int(f.GetLastWinOrLoss()), lr, this.cmp) {
		return b3.SUCCESS
	}
	return b3.FAILURE
}
