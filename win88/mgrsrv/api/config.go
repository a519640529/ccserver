package api

import (
	"github.com/idealeak/goserver/core"
)

var Config = Configuration{}

type Configuration struct {
	StartScript string
	IsDevMode   bool
}

func (this *Configuration) Name() string {
	return "srvctrl"
}

func (this *Configuration) Init() error {
	return nil
}

func (this *Configuration) Close() error {
	return nil
}

func init() {
	core.RegistePackage(&Config)
}
