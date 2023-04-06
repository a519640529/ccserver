package main

import (
	"games.yol.com/win88/common"
	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/module"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().Unix())
	core.RegisterConfigEncryptor(common.ConfigFE)
	defer core.ClosePackages()
	core.LoadPackages("config.json")

	waiter := module.Start()
	waiter.Wait("main()")
}
