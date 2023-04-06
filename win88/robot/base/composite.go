package base

import (
	"games.yol.com/win88/common"
	"github.com/idealeak/goserver/core/logger"
	b3 "github.com/magicsea/behavior3go"
	b3config "github.com/magicsea/behavior3go/config"
	b3core "github.com/magicsea/behavior3go/core"
	"math/rand"
)

//随机
type RandomComposite struct {
	b3core.Composite
}

func (this *RandomComposite) OnOpen(tick *b3core.Tick) {
	tick.Blackboard.Set("runningChild", -1, tick.GetTree().GetID(), this.GetID())
}

func (this *RandomComposite) OnTick(tick *b3core.Tick) b3.Status {
	var child = tick.Blackboard.GetInt("runningChild", tick.GetTree().GetID(), this.GetID())
	if -1 == child {
		child = int(rand.Uint32()) % this.GetChildCount()
	}

	var status = this.GetChild(child).Execute(tick)
	if status == b3.RUNNING {
		tick.Blackboard.Set("runningChild", child, tick.GetTree().GetID(), this.GetID())
	} else {
		tick.Blackboard.Set("runningChild", -1, tick.GetTree().GetID(), this.GetID())
	}
	return status
}

//随机weight
type RandomWeightComposite struct {
	b3core.Composite
	weight string //子节点权重 | 分割
}

func (this *RandomWeightComposite) Initialize(setting *b3config.BTNodeCfg) {
	this.Composite.Initialize(setting)
	this.weight = setting.GetPropertyAsString("weight")
}

func (this *RandomWeightComposite) OnOpen(tick *b3core.Tick) {
	tick.Blackboard.Set("runningChild", -1, tick.GetTree().GetID(), this.GetID())
}

func (this *RandomWeightComposite) OnTick(tick *b3core.Tick) b3.Status {
	var child = tick.Blackboard.GetInt("runningChild", tick.GetTree().GetID(), this.GetID())
	if -1 == child {
		iArray, _ := common.StrSliceInt(this.weight, "|")
		child = common.RandSliceIndexByWightN(iArray)
	}

	var status = this.GetChild(child).Execute(tick)
	if status == b3.RUNNING {
		tick.Blackboard.Set("runningChild", child, tick.GetTree().GetID(), this.GetID())
	} else {
		tick.Blackboard.Set("runningChild", -1, tick.GetTree().GetID(), this.GetID())
	}
	return status
}

//Parallel
type ParallelComposite struct {
	b3core.Composite
	failCond int //1有一个失败就失败 0全失败才失败
	succCond int //1有一个成功就成功 0全成功才成功
	//如果不能确定状态 那就有running返回running，不然失败
}

func (this *ParallelComposite) Initialize(setting *b3config.BTNodeCfg) {
	this.Composite.Initialize(setting)
	this.failCond = setting.GetPropertyAsInt("fail_cond")
	this.succCond = setting.GetPropertyAsInt("succ_cond")
}

func (this *ParallelComposite) OnTick(tick *b3core.Tick) b3.Status {
	var failCount int
	var succCount int
	var hasRunning bool
	for i := 0; i < this.GetChildCount(); i++ {
		var status = this.GetChild(i).Execute(tick)
		if status == b3.FAILURE {
			failCount++
		} else if status == b3.SUCCESS {
			succCount++
		} else {
			hasRunning = true
		}
	}
	if (this.failCond == 0 && failCount == this.GetChildCount()) || (this.failCond == 1 && failCount > 0) {
		return b3.FAILURE
	}
	if (this.succCond == 0 && succCount == this.GetChildCount()) || (this.succCond == 1 && succCount > 0) {
		return b3.FAILURE
	}
	if hasRunning {
		return b3.RUNNING
	}
	return b3.FAILURE
}

//SubTree
type SubTreeNode struct {
	b3core.Action
	sTree    *b3core.BehaviorTree
	treeName string
}

func (this *SubTreeNode) Initialize(setting *b3config.BTNodeCfg) {
	this.Action.Initialize(setting)
	this.treeName = setting.GetPropertyAsString("treeName")
	this.sTree = CreateBevTree(this.treeName)
	if nil == this.sTree {
		logger.Logger.Errorf("SubTreeNode Get SubTree Failed, treeName:%v ", this.treeName)
	}
	logger.Logger.Info("SubTreeNode::Initialize ", this, " treeName ", this.treeName)
}

func (this *SubTreeNode) OnTick(tick *b3core.Tick) b3.Status {
	if nil == this.sTree {
		return b3.ERROR
	}
	if tick.GetTarget() == nil {
		panic("unknow error!")
	}
	tar := tick.GetTarget()
	//	glog.Info("subtree: ", this.treeName, " id ", player.id)
	return this.sTree.Tick(tar, tick.Blackboard)
}
