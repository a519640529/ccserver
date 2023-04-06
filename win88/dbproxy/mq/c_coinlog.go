package mq

import (
	"encoding/json"
	"games.yol.com/win88/dbproxy/svc"
	"games.yol.com/win88/model"
	"games.yol.com/win88/mq"
	"github.com/idealeak/goserver/core/broker"
	"github.com/idealeak/goserver/core/broker/rabbitmq"
)

func init() {
	mq.RegisteSubscriber(model.CoinLogCollName, func(e broker.Event) (err error) {
		msg := e.Message()
		if msg != nil {
			defer func() {
				if err != nil {
					mq.BackUp(e, err)
				}

				e.Ack()

				recover()
			}()

			var log model.CoinLog
			err = json.Unmarshal(msg.Body, &log)
			if err != nil {
				return
			}

			if log.Count == 0 { //玩家冲账探针
				RabbitMQPublisher.Send(model.TopicProbeCoinLogAck, log)
			} else {
				c := svc.CoinLogsCollection(log.Platform)
				if c != nil {
					err = c.Insert(log)
					if err == nil {
						err = svc.InsertCoinWAL(log.Platform, model.NewCoinWAL(log.SnId, log.Count, log.LogType, log.InGame, log.CoinType, log.RoomId, log.Time.UnixNano()))
						if err != nil {
							return
						}
					}
				}
			}
			return
		}
		return nil
	}, broker.Queue(model.CoinLogCollName), broker.DisableAutoAck(), rabbitmq.DurableQueue())
}
