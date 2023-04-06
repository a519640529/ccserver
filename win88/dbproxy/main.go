package main

import (
	"games.yol.com/win88/common"
	_ "games.yol.com/win88/dbproxy/mq"
	_ "games.yol.com/win88/dbproxy/svc"
	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/module"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/rpc"
	"time"
)

func main() {
	rand.Seed(time.Now().Unix())
	defer core.ClosePackages()
	core.LoadPackages("config.json")

	rpc.HandleHTTP() // 采用http协议作为rpc载体
	lis, err := net.Listen(common.CustomConfig.GetString("MgoRpcCliNet"), common.CustomConfig.GetString("MgoRpcCliAddr"))
	if err != nil {
		log.Fatalln("fatal error: ", err)
	}
	go http.Serve(lis, nil)

	waitor := module.Start()
	waitor.Wait("main()")
}
