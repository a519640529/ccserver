package base

import (
	"games.yol.com/win88/model"
	"reflect"
)

const (
	LOGCHANEL_BLACKHOLE = "_null_"
)

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
	LogChannelSington.RegisteLogCName(model.LoginLogCollName, &model.LoginLog{})
	LogChannelSington.RegisteLogCName(model.CoinGiveLogCollName, &model.CoinGiveLog{})
	LogChannelSington.RegisteLogCName(model.CoinLogCollName, &model.CoinLog{})
	LogChannelSington.RegisteLogCName(model.SceneCoinLogCollName, &model.SceneCoinLog{})
	LogChannelSington.RegisteLogCName(model.GameDetailedLogCollName, &model.GameDetailedLog{})
	LogChannelSington.RegisteLogCName(model.GamePlayerListLogCollName, &model.GamePlayerListLog{})
	LogChannelSington.RegisteLogCName(model.FriendRecordLogCollName, &model.FriendRecord{})
	LogChannelSington.RegisteLogCName(model.ItemLogCollName, &model.ItemLog{})
}
