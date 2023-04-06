package main

import (
	_ "games.yol.com/win88"
	"games.yol.com/win88/common"
	_ "games.yol.com/win88/gamesrv/action"
	_ "games.yol.com/win88/gamesrv/base"
	_ "games.yol.com/win88/gamesrv/fishing"
	_ "games.yol.com/win88/gamesrv/fruits"
	_ "games.yol.com/win88/gamesrv/richblessed"
	_ "games.yol.com/win88/gamesrv/tienlen"
	_ "games.yol.com/win88/gamesrv/transact"
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
