package base

import (
	"games.yol.com/win88/common"
	"github.com/idealeak/goserver/core/logger"
	b3 "github.com/magicsea/behavior3go"
	b3config "github.com/magicsea/behavior3go/config"
	b3core "github.com/magicsea/behavior3go/core"
	"math/rand"
	"time"
)

type RandWait struct {
	b3core.Action
	minTime int64
	maxTime int64
}

func (this *RandWait) Initialize(setting *b3config.BTNodeCfg) {
	this.Action.Initialize(setting)
	this.minTime = setting.GetPropertyAsInt64("minTime")
	this.maxTime = setting.GetPropertyAsInt64("maxTime")
}

func (this *RandWait) OnOpen(tick *b3core.Tick) {
	var startTime int64 = time.Now().UnixNano() / 1000000
	tick.Blackboard.Set("startTime", startTime, tick.GetTree().GetID(), this.GetID())
	end := this.minTime + rand.Int63n(this.maxTime-this.minTime)
	tick.Blackboard.Set("endTime", startTime+end, tick.GetTree().GetID(), this.GetID())
}

func (this *RandWait) OnTick(tick *b3core.Tick) b3.Status {
	var currTime int64 = time.Now().UnixNano() / 1000000
	var endTime = tick.Blackboard.GetInt64("endTime", tick.GetTree().GetID(), this.GetID())

	if currTime > endTime {
		return b3.SUCCESS
	}

	return b3.RUNNING
}

//RandIntAction
type RandIntAction struct {
	b3core.Action
	index string
	min   int
	max   int
}

func (this *RandIntAction) Initialize(setting *b3config.BTNodeCfg) {
	this.Action.Initialize(setting)
	this.index = setting.GetPropertyAsString("index")
	this.min = setting.GetPropertyAsInt("min")
	this.max = setting.GetPropertyAsInt("max")
}

func (this *RandIntAction) OnTick(tick *b3core.Tick) b3.Status {
	val := common.RandInt(this.min, this.max)
	tick.Blackboard.Set(this.index, val, "", "")
	return b3.SUCCESS
}

//log
type LogAction struct {
	b3core.Action
	level int
	info  string
}

func (this *LogAction) Initialize(setting *b3config.BTNodeCfg) {
	this.Action.Initialize(setting)
	this.info = setting.GetPropertyAsString("info")
	this.level = setting.GetPropertyAsInt("level")
}

func (this *LogAction) OnTick(tick *b3core.Tick) b3.Status {

	logger.Logger.Info(this.info)
	return b3.SUCCESS
}

//游戏离场
type LeaveGame struct {
	b3core.Action
}

func (this *LeaveGame) Initialize(setting *b3config.BTNodeCfg) {
	this.Action.Initialize(setting)
}

func (this *LeaveGame) OnTick(tick *b3core.Tick) b3.Status {
	player := tick.GetTarget().(Player)

	player.LeaveGameMsg(player.GetSnId())
	return b3.SUCCESS
}

//保存游戏出场限制到黑板
type GetOutLimitCoin struct {
	b3core.Action
	gINKey string
}

func (this *GetOutLimitCoin) Initialize(setting *b3config.BTNodeCfg) {
	this.Action.Initialize(setting)
	this.gINKey = setting.GetPropertyAsString("gINKey")
}

func (this *GetOutLimitCoin) OnTick(tick *b3core.Tick) b3.Status {
	player := tick.GetTarget().(Player)
	coin := player.GetOutLimitCoin()
	if coin == -1 {
		return b3.FAILURE
	}

	tick.Blackboard.Set(this.gINKey, coin, "", "")
	return b3.SUCCESS
}

//SetInt
type SetIntAction struct {
	b3core.Action
	gINKey   string
	gIPValue string
}

func (this *SetIntAction) Initialize(setting *b3config.BTNodeCfg) {
	this.Action.Initialize(setting)
	this.gINKey = setting.GetPropertyAsString("gINKey")
	this.gIPValue = setting.GetPropertyAsString("gIPValue")

}

func (this *SetIntAction) OnTick(tick *b3core.Tick) b3.Status {
	val := getBoardIntByPreStr(tick.Blackboard, this.gIPValue, "", "")

	tick.Blackboard.Set(this.gINKey, val, "", "")
	return b3.SUCCESS
}

//SetIntMulti
type SetIntMulti struct {
	b3core.Action
	gINKey    string
	gIPValue1 string
	gIPValue2 string
}

func (this *SetIntMulti) Initialize(setting *b3config.BTNodeCfg) {
	this.Action.Initialize(setting)
	this.gINKey = setting.GetPropertyAsString("gINKey")
	this.gIPValue1 = setting.GetPropertyAsString("gIPValue1")
	this.gIPValue2 = setting.GetPropertyAsString("gIPValue2")

}

func (this *SetIntMulti) OnTick(tick *b3core.Tick) b3.Status {
	val1 := getBoardIntByPreStr(tick.Blackboard, this.gIPValue1, "", "")
	val2 := getBoardIntByPreStr(tick.Blackboard, this.gIPValue2, "", "")

	tick.Blackboard.Set(this.gINKey, val1*val2, "", "")
	return b3.SUCCESS
}

//GetPlayerCoin
type GetPlayerCoin struct {
	b3core.Action
	gINKey string
}

func (this *GetPlayerCoin) Initialize(setting *b3config.BTNodeCfg) {
	this.Action.Initialize(setting)
	this.gINKey = setting.GetPropertyAsString("gINKey")
}

func (this *GetPlayerCoin) OnTick(tick *b3core.Tick) b3.Status {
	player := tick.GetTarget().(Player)
	coin := player.GetCoin()

	tick.Blackboard.Set(this.gINKey, int(coin), "", "")
	return b3.SUCCESS
}

//SetIntDiv
type SetIntDiv struct {
	b3core.Action
	gINKey    string
	gIPValue1 string
	gIPValue2 string
}

func (this *SetIntDiv) Initialize(setting *b3config.BTNodeCfg) {
	this.Action.Initialize(setting)
	this.gINKey = setting.GetPropertyAsString("gINKey")
	this.gIPValue1 = setting.GetPropertyAsString("gIPValue1")
	this.gIPValue2 = setting.GetPropertyAsString("gIPValue2")

}

func (this *SetIntDiv) OnTick(tick *b3core.Tick) b3.Status {
	val1 := getBoardIntByPreStr(tick.Blackboard, this.gIPValue1, "", "")
	val2 := getBoardIntByPreStr(tick.Blackboard, this.gIPValue2, "", "")
	if val2 == 0 {
		return b3.FAILURE
	}

	tick.Blackboard.Set(this.gINKey, val1/val2, "", "")
	return b3.SUCCESS
}

//GetPlayerTakeCoin
type GetPlayerTakeCoin struct {
	b3core.Action
	gINKey string
}

func (this *GetPlayerTakeCoin) Initialize(setting *b3config.BTNodeCfg) {
	this.Action.Initialize(setting)
	this.gINKey = setting.GetPropertyAsString("gINKey")
}

func (this *GetPlayerTakeCoin) OnTick(tick *b3core.Tick) b3.Status {
	player := tick.GetTarget().(Player)
	coin := player.GetTakeCoin()

	tick.Blackboard.Set(this.gINKey, int(coin), "", "")
	return b3.SUCCESS
}
