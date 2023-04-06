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
	mq.RegisteSubscriber(model.ItemLogCollName, func(e broker.Event) (err error) {
		msg := e.Message()
		if msg != nil {
			defer func() {
				if err != nil {
					mq.BackUp(e, err)
				}

				e.Ack()

				recover()
			}()

			var log model.ItemLog
			err = json.Unmarshal(msg.Body, &log)
			if err != nil {
				return
			}

			c := svc.ItemLogsCollection(log.Platform)
			if c != nil {
				err = c.Insert(log)
			}
			return
		}
		return nil
	}, broker.Queue(model.ItemLogCollName), broker.DisableAutoAck(), rabbitmq.DurableQueue())
}
