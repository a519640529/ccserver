package base

import (
	"github.com/idealeak/goserver/core"
)

var (
	ForwardConfig = ForwardConfiguration{}
)

type ForwardConfiguration struct {
	ForwardMakeCards bool
}

func (this *ForwardConfiguration) Name() string {
	return "forward"
}

func (this *ForwardConfiguration) Init() error {
	return nil
}

func (this *ForwardConfiguration) Close() error {
	return nil
}

func init() {
	core.RegistePackage(&ForwardConfig)
}
