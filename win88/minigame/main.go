package main

import (
	_ "games.yol.com/win88"
	"games.yol.com/win88/common"
	_ "games.yol.com/win88/minigame/action"
	_ "games.yol.com/win88/minigame/base"
	_ "games.yol.com/win88/minigame/game/candy"
	_ "games.yol.com/win88/minigame/game/caothap"
	_ "games.yol.com/win88/minigame/game/luckydice"
	_ "games.yol.com/win88/minigame/game/minipoker"
	_ "games.yol.com/win88/minigame/transact"
	_ "games.yol.com/win88/srvdata"
	"github.com/idealeak/goserver/core"
	_ "github.com/idealeak/goserver/core/i18n"
	"github.com/idealeak/goserver/core/module"
)

func main() {
	core.RegisterConfigEncryptor(common.ConfigFE)
	defer core.ClosePackages()
	core.LoadPackages("config.json")

	waiter := module.Start()
	waiter.Wait("main()")
}
