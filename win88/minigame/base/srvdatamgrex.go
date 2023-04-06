package base

import (
	"games.yol.com/win88/srvdata"
	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/logger"
)

var SrvDataMgrEx = &SrvDataManagerEx{
	DataReloader: make(map[string]SrvDataReloadInterface),
}

type SrvDataManagerEx struct {
	DataReloader map[string]SrvDataReloadInterface
}

func RegisterDataReloader(fileName string, sdri SrvDataReloadInterface) {
	SrvDataMgrEx.DataReloader[fileName] = sdri
}

type SrvDataReloadInterface interface {
	Reload()
}

//初始化在线奖励系统
func init() {
	core.RegisteHook(core.HOOK_BEFORE_START, func() error {
		logger.Logger.Info("初始化牌库[S]")
		defer logger.Logger.Info("初始化牌库[E]")
		return nil
	})
}
func init() {
	srvdata.SrvDataModifyCB = func(fileName string, fullName string) {
		if dr, ok := SrvDataMgrEx.DataReloader[fileName]; ok {
			dr.Reload()
		}
	}
}
