package main

import (
	_ "games.yol.com/win88"
	_ "games.yol.com/win88/mgrsrv/api"
	_ "games.yol.com/win88/srvdata"

	"games.yol.com/win88/common"
	"games.yol.com/win88/model"
	"games.yol.com/win88/srvdata"
	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/module"
)

func main() {
	core.RegisterConfigEncryptor(common.ConfigFE)
	defer core.ClosePackages()
	core.LoadPackages("config.json")

	model.InitGameParam()
	logger.Logger.Warnf("log data %v", srvdata.Config.RootPath)
	waiter := module.Start()
	waiter.Wait("main()")
}

func init() {
	//首先加载游戏配置
	//core.RegisteHook(core.HOOK_BEFORE_START, func() error {
	//	model.StartupRPClient(common.CustomConfig.GetString("MgoRpcCliNet"), common.CustomConfig.GetString("MgoRpcCliAddr"), time.Duration(common.CustomConfig.GetInt("MgoRpcCliReconnInterV"))*time.Second)
	//	return nil
	//})
	//
	//core.RegisteHook(core.HOOK_AFTER_STOP, func() error {
	//	model.ShutdownRPClient()
	//	return nil
	//})
}
