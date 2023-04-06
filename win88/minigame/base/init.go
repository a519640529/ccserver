package base

import (
	"games.yol.com/win88/common"
	"games.yol.com/win88/model"
	"games.yol.com/win88/mq"
	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/broker/rabbitmq"
	"math/rand"
	"time"
)

var RabbitMQPublisher *mq.RabbitMQPublisher

func init() {
	model.InitGameParam()

	rand.Seed(time.Now().Unix())
	core.RegisteHook(core.HOOK_BEFORE_START, func() error {
		model.StartupRPClient(common.CustomConfig.GetString("MgoRpcCliNet"), common.CustomConfig.GetString("MgoRpcCliAddr"), time.Duration(common.CustomConfig.GetInt("MgoRpcCliReconnInterV"))*time.Second)
		model.InitGameKVData()

		//rabbitmq打开链接
		RabbitMQPublisher = mq.NewRabbitMQPublisher(common.CustomConfig.GetString("RabbitMQURL"), rabbitmq.Exchange{Name: common.CustomConfig.GetString("RMQExchange"), Durable: true}, common.CustomConfig.GetInt("RMQPublishBacklog"))
		if RabbitMQPublisher != nil {
			err := RabbitMQPublisher.Start()
			if err != nil {
				panic(err)
			}
		}

		return nil
	})

	core.RegisteHook(core.HOOK_AFTER_STOP, func() error {
		model.ShutdownRPClient()

		//关闭rabbitmq连接
		if RabbitMQPublisher != nil {
			RabbitMQPublisher.Stop()
		}
		return nil
	})
}
