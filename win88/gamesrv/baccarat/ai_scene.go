package baccarat

import (
	"games.yol.com/win88/gamerule/baccarat"
	"games.yol.com/win88/gamesrv/base"
)

type BaccaratSceneAIMgr struct {
	base.HundredSceneAIMgr
}

//房间启动
func (this *BaccaratSceneAIMgr) OnStart(s *base.Scene) {
	this.HundredSceneAIMgr.OnStart(s)
}

//房间心跳
func (this *BaccaratSceneAIMgr) OnTick(s *base.Scene) {
	this.HundredSceneAIMgr.OnTick(s)
}

//房间停止
func (this *BaccaratSceneAIMgr) OnStop(s *base.Scene) {
	this.HundredSceneAIMgr.OnStop(s)
}

//房间状态变化
func (this *BaccaratSceneAIMgr) OnChangeState(s *base.Scene, oldstate, newstate int) {
	this.HundredSceneAIMgr.OnChangeState(s, oldstate, newstate)
	switch newstate {
	case baccarat.BaccaratSceneStateStake:

	case baccarat.BaccaratSceneStateBilled:

	}
}
