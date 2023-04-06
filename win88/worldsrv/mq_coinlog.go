package main

import (
	"encoding/json"
	"games.yol.com/win88/model"
	"games.yol.com/win88/mq"
	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/broker"
	"github.com/idealeak/goserver/core/broker/rabbitmq"
)

func init() {
	mq.RegisteSubscriber(model.TopicProbeCoinLogAck, func(e broker.Event) (err error) {
		msg := e.Message()
		if msg != nil {
			defer func() {
				e.Ack()
				recover()
			}()

			var log model.CoinLog
			err = json.Unmarshal(msg.Body, &log)
			if err != nil {
				return
			}

			//通知主线程执行后续操作
			core.CoreObject().SendCommand(basic.CommandWrapper(func(o *basic.Object) error {
				player := PlayerMgrSington.GetPlayerBySnId(log.SnId)
				if player != nil {
					player.Coin += log.RestCount
					player.SyncGameCoin(int(log.RoomId), log.SeqNo)
				}
				return nil
			}), true)
			return
		}
		return nil
	}, broker.Queue(model.TopicProbeCoinLogAck), rabbitmq.DurableQueue())
}
