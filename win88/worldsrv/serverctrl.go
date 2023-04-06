package main

import (
	"games.yol.com/win88/common"
	"github.com/idealeak/goserver/core/mongo"
)

var SrvIsMaintaining = true

func init() {
	common.RegisteServerCtrlCallback(func(code int32) {
		switch code {
		case common.SrvCtrlStateSwitchCode:
			SrvIsMaintaining = !SrvIsMaintaining
		case common.SrvCtrlResetMgoSession:
			mongo.ResetAllSession()
		}
	})
}
