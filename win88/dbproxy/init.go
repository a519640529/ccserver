package main

import (
	"games.yol.com/win88/common"
	"games.yol.com/win88/dbproxy/svc"
	"games.yol.com/win88/model"
	"games.yol.com/win88/mq"
	"github.com/globalsign/mgo"
	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/broker/rabbitmq"
	"math/rand"
	"time"
)

var rabbitMqConsumer *mq.RabbitMQConsumer

type MgoLogger struct {
}

func (log *MgoLogger) Output(calldepth int, s string) error {
	slog := common.GetLoggerInstanceByName("mongo_logger")
	if slog != nil {
		slog.SetAdditionalStackDepth(2)
		slog.Debug(s)
	}

	return nil
}

func init() {
	mgo.SetLogger(&MgoLogger{})

	rand.Seed(time.Now().Unix())
	core.RegisteHook(core.HOOK_BEFORE_START, func() error {
		//etcd打开连接
		EtcdMgrSington.Init()
		EtcdMgrSington.InitPlatformDBCfg()
		model.InitGameParam()

		rabbitMqConsumer = mq.NewRabbitMQConsumer(common.CustomConfig.GetString("RabbitMQURL"), rabbitmq.Exchange{Name: common.CustomConfig.GetString("RMQExchange"), Durable: true})
		if rabbitMqConsumer != nil {
			rabbitMqConsumer.Start()
		}

		//尝试初始化
		svc.GetOnePlayerIdFromBucket()
		return nil
	})

	core.RegisteHook(core.HOOK_AFTER_STOP, func() error {
		//etcd关闭连接
		EtcdMgrSington.Shutdown()

		if rabbitMqConsumer != nil {
			rabbitMqConsumer.Stop()
		}
		model.ShutdownRPClient()
		return nil
	})
}
