// config
package webapi

import (
	"fmt"
	"github.com/idealeak/goserver/core"
)

var (
	Config = Configuration{}
)

type Configuration struct {
	GameApiURL string
}

func (this *Configuration) Name() string {
	return "webapi"
}

func (this *Configuration) Init() (err error) {
	fmt.Println(*this)
	return nil
}

func (this *Configuration) Close() error {
	return nil
}

func init() {
	core.RegistePackage(&Config)
}
