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
	mq.RegisteSubscriber(model.LoginLogCollName, func(e broker.Event) (err error) {
		msg := e.Message()
		if msg != nil {
			defer func() {
				if err != nil {
					mq.BackUp(e, err)
				}

				e.Ack()

				recover()
			}()

			var log model.LoginLog
			err = json.Unmarshal(msg.Body, &log)
			if err != nil {
				return
			}

			c := svc.LoginLogsCollection(log.Platform)
			if c != nil {
				err = c.Insert(log)
			}
			return
		}
		return nil
	}, broker.Queue(model.LoginLogCollName), broker.DisableAutoAck(), rabbitmq.DurableQueue())

	//for test
	//RegisteSubscriber(model.LoginLogCollName, func(e broker.Event) (err error) {
	//	msg := e.Message()
	//	if msg != nil {
	//		var log model.LoginLog
	//		err = json.Unmarshal(msg.Body, &log)
	//		if err != nil {
	//			return
	//		}
	//
	//		logger.Logger.Trace(log)
	//		return
	//	}
	//	return nil
	//}, broker.Queue(model.LoginLogCollName+"_echo"), rabbitmq.DurableQueue())
	//for test
}
