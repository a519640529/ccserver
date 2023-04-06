package main

import (
	"games.yol.com/win88/common"
	"games.yol.com/win88/model"
	"games.yol.com/win88/mq"
	"games.yol.com/win88/ranksrv/base"
	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/broker/rabbitmq"
	"time"
)

var RabbitMQPublisher *mq.RabbitMQPublisher
var RabbitMqConsumer *mq.RabbitMQConsumer

func init() {
	//首先加载游戏配置
	core.RegisteHook(core.HOOK_BEFORE_START, func() error {
		//初始化rpc
		model.StartupRPClient(common.CustomConfig.GetString("MgoRpcCliNet"), common.CustomConfig.GetString("MgoRpcCliAddr"), time.Duration(common.CustomConfig.GetInt("MgoRpcCliReconnInterV"))*time.Second)
		//etcd初始化
		base.EtcdMgrSington.Init()
		base.EtcdMgrSington.InitPlatform() //拉取到所有的平台id

		//rabbitmq打开链接
		RabbitMqConsumer = mq.NewRabbitMQConsumer(common.CustomConfig.GetString("RabbitMQURL"), rabbitmq.Exchange{Name: common.CustomConfig.GetString("RMQExchange"), Durable: true})
		if RabbitMqConsumer != nil {
			RabbitMqConsumer.Start()
		}


		return nil
	})

	core.RegisteHook(core.HOOK_AFTER_STOP, func() error {
		//etcd关闭连接
		base.EtcdMgrSington.Shutdown()

		if RabbitMqConsumer != nil {
			RabbitMqConsumer.Stop()
		}
		model.ShutdownRPClient()
		return nil
	})
}
