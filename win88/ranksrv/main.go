package main

import (
	_ "games.yol.com/win88"
	"games.yol.com/win88/common"
	_ "games.yol.com/win88/ranksrv/mq"
	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/module"
	"github.com/idealeak/goserver/core/schedule"
)

var (
	gDbType  string
	gConnStr string
	gAPIAddr string
	//BoardStoreSinglton *BoardStore = NewBoardStore()
	Dev bool
)

func main() {
	core.RegisterConfigEncryptor(common.ConfigFE)
	core.LoadPackages("config.json")
	defer core.ClosePackages()

	schedule.StartTask()

	waitor := module.Start()
	waitor.Wait("ranksrv")
}

func init() {
	core.RegisteHook(core.HOOK_BEFORE_START, func() error {
		return nil
	})

	core.RegisteHook(core.HOOK_AFTER_STOP, func() error {
		//BoardStoreSinglton.Close()
		return nil
	})
}
