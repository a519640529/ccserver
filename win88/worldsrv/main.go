package main

import (
	"math/rand"
	"time"

	_ "games.yol.com/win88"
	"games.yol.com/win88/common"
	"github.com/idealeak/goserver/core"
	_ "github.com/idealeak/goserver/core/i18n"
	"github.com/idealeak/goserver/core/module"
	"github.com/idealeak/goserver/core/schedule"
)

func main() {
	rand.Seed(time.Now().Unix())
	core.RegisterConfigEncryptor(common.ConfigFE)
	defer core.ClosePackages()
	core.LoadPackages("config.json")

	//启动定时任务
	schedule.StartTask()

	//启动业务模块
	waiter := module.Start()
	waiter.Wait("main()")
}
