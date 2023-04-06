package base

import (
	"games.yol.com/win88/common"
	"github.com/idealeak/goserver/core"
)

var ServerStateMgr = &ServerStateManager{
	State: common.GAME_SESS_STATE_ON,
}

type ServerStateManager struct {
	State common.GameSessState
}

func (this *ServerStateManager) Init() {
	this.State = common.GAME_SESS_STATE_ON
}

func (this *ServerStateManager) SetState(state common.GameSessState) {
	this.State = state
}

func (this *ServerStateManager) GetState() common.GameSessState {
	return this.State
}

func init() {
	core.RegisteHook(core.HOOK_BEFORE_START, func() error {
		ServerStateMgr.Init()
		return nil
	})
}
