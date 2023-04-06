package mq

import (
	"games.yol.com/win88/common"
	"games.yol.com/win88/mq"
	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/broker/rabbitmq"
)

var RabbitMQPublisher *mq.RabbitMQPublisher

func init() {
	////首先加载游戏配置
	core.RegisteHook(core.HOOK_BEFORE_START, func() error {
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
		//关闭rabbitmq连接
		if RabbitMQPublisher != nil {
			RabbitMQPublisher.Stop()
		}

		return nil
	})
}
