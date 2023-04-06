package api

import (
	"reflect"

	"games.yol.com/win88/common"
	"games.yol.com/win88/model"
	"games.yol.com/win88/mq"
	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/broker/rabbitmq"
)

const (
	LOGCHANEL_BLACKHOLE = "_null_"
)

var RabbitMQPublisher *mq.RabbitMQPublisher

var LogChannelSington = &LogChannel{
	cName: make(map[reflect.Type]string),
}

type LogChannel struct {
	cName map[reflect.Type]string
}

func (c *LogChannel) RegisteLogCName(cname string, log interface{}) {
	t := c.getLogType(log)
	c.cName[t] = cname
}

func (c *LogChannel) getLogType(log interface{}) reflect.Type {
	return reflect.Indirect(reflect.ValueOf(log)).Type()
}

func (c *LogChannel) getLogCName(log interface{}) string {
	t := c.getLogType(log)
	if name, exist := c.cName[t]; exist {
		return name
	}
	return ""
}

func (c *LogChannel) WriteLog(log interface{}) {
	cname := c.getLogCName(log)
	if cname == "" {
		cname = LOGCHANEL_BLACKHOLE
	}
	RabbitMQPublisher.Send(cname, log)
}

func (c *LogChannel) WriteMQData(data *model.RabbitMQData) {
	RabbitMQPublisher.Send(data.MQName, data.Data)
}

func init() {
	LogChannelSington.RegisteLogCName(model.APILogCollName, &model.APILog{})

	//首先加载游戏配置
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
