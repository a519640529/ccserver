package model

import (
	"games.yol.com/win88/rpc"
	"time"
)

var (
	rpcCli *rpc.RPClient
)

func StartupRPClient(net, addr string, d time.Duration) {
	cli := rpcCli
	if cli != nil {
		cli.Stop()
	}
	rpcCli = rpc.NewRPClient(net, addr, d)
	if rpcCli != nil {
		rpcCli.Start()
	}
}

func ShutdownRPClient() {
	if rpcCli != nil {
		cli := rpcCli
		rpcCli = nil
		cli.Stop()
	}
}
