package common

import (
	"github.com/idealeak/goserver/core/logger"
)

const (
	A_USER_BLACK = 1 //增加黑名单
)

type Action struct {
	ActionID          int       //执行id
	ActionParamInt64  []int     //整形参数
	ActionParamFloat  []float64 //浮点参数
	ActionParamString []string  //字符串参数
}

var ActionMgrSington = &ActionMgr{
	pool: make(map[int]ActionBase),
}

func init() {

}

type ActionMgr struct {
	pool map[int]ActionBase
}

func (this *ActionMgr) ActionGroup(need interface{}, action []*Action) bool {
	for i := 0; i < len(action); i++ {
		this.action(need, action[i])
	}

	return true
}

func (this *ActionMgr) action(need interface{}, action *Action) bool {
	a, ok := this.pool[action.ActionID]
	if !ok {
		logger.Logger.Warnf("no this action %v", action.ActionID)
		return false
	}

	return a.Action(need, action)
}

func (this *ActionMgr) Register(cid int, c ActionBase) {
	this.pool[cid] = c
}

type ActionBase interface {
	Action(need interface{}, action *Action) bool
}
