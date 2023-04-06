package main

import (
	_ "games.yol.com/win88"
	"games.yol.com/win88/common"
	_ "games.yol.com/win88/common"
	_ "games.yol.com/win88/robot/base"
	_ "games.yol.com/win88/robot/fishing"
	_ "games.yol.com/win88/robot/tienlen"
	_ "games.yol.com/win88/srvdata"
	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/module"
)

func main() {
	core.RegisterConfigEncryptor(common.ConfigFE)
	defer core.ClosePackages()
	core.LoadPackages("config.json")

	waiter := module.Start()
	waiter.Wait("main()")
}
