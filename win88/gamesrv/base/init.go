package base

import (
	"errors"
	"games.yol.com/win88/common"
	"games.yol.com/win88/gamerule/crash"
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
	model.InitFishingParam()
	model.InitNormalParam()
	model.InitGMAC()

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

		hs := model.GetCrashHash(0)

		model.GameParamData.InitGameHash = []string{}
		for _, v := range hs {
			model.GameParamData.InitGameHash = append(model.GameParamData.InitGameHash, v.Hash)
		}
		hsatom := model.GetCrashHashAtom(0)

		model.GameParamData.AtomGameHash = []string{}
		for _, v := range hsatom {
			model.GameParamData.AtomGameHash = append(model.GameParamData.AtomGameHash, v.Hash)
		}

		if len(model.GameParamData.AtomGameHash) < crash.POKER_CART_CNT ||
			len(model.GameParamData.InitGameHash) < crash.POKER_CART_CNT {
			panic(errors.New("hash is read error"))
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
