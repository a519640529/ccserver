package main

import (
	_ "games.yol.com/win88"
	"games.yol.com/win88/common"
	"games.yol.com/win88/model"

	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/module"
)

func main() {
	core.RegisterConfigEncryptor(common.ConfigFE)
	defer core.ClosePackages()
	core.LoadPackages("config.json")

	model.InitGameParam()

	waiter := module.Start()
	waiter.Wait("main()")
}

/*
   提交测试
*/
