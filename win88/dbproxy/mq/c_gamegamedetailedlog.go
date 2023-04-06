package mq

import (
	"encoding/json"
	"games.yol.com/win88/dbproxy/svc"
	"games.yol.com/win88/model"
	"games.yol.com/win88/mq"
	"github.com/idealeak/goserver/core/broker"
	"github.com/idealeak/goserver/core/broker/rabbitmq"
	"github.com/idealeak/goserver/core/logger"
)

func init() {
	mq.RegisteSubscriber(model.GameDetailedLogCollName, func(e broker.Event) (err error) {
		msg := e.Message()
		if msg != nil {
			defer func() {
				if err != nil {
					mq.BackUp(e, err)
				}

				e.Ack()

				recover()
			}()

			var log model.GameDetailedLog
			err = json.Unmarshal(msg.Body, &log)
			if err != nil {
				return
			}
			logger.Logger.Tracef("mq receive GameDetailedLog:%v", log)
			c := svc.GameDetailedLogsCollection(log.Platform)
			if c != nil {
				err = c.Insert(log)
				if err != nil {
					logger.Logger.Tracef("c.Insert(log) err:%v", err.Error())
				}
			}
			return
		}
		return nil
	}, broker.Queue(model.GameDetailedLogCollName), broker.DisableAutoAck(), rabbitmq.DurableQueue())
}
