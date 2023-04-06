package base

import (
	"github.com/idealeak/goserver/core/basic"
)

type SceneEventHandler interface {
	Handler(*Scene, interface{})
}

type SceneEventHandlerWrapper func(*Scene, interface{})

func (sehw SceneEventHandlerWrapper) Handler(s *Scene, d interface{}) {
	sehw(s, d)
}

type SceneEvent struct {
	d interface{}
	h SceneEventHandler
}

type SceneEventCommand struct {
	*Scene
	*SceneEvent
}

func (this *SceneEventCommand) Done(o *basic.Object) error {
	defer o.ProcessSeqnum()
	this.SceneEvent.h.Handler(this.Scene, this.SceneEvent.d)
	return nil
}

type SceneTimerHandlerWrapper func(*Scene, interface{})
type SceneTimer struct {
	d interface{}
	c int32
}
