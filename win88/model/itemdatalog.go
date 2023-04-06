package model

import (
	"github.com/globalsign/mgo/bson"
	"time"
)

var (
	ItemLogDBName        = "log"
	ItemLogCollName      = "log_itemlog"
	TopicProbeItemLogAck = "ack_itemlog"
)

type ItemLog struct {
	LogId    bson.ObjectId `bson:"_id"`
	Platform string        //平台
	SnId     int32         //玩家id
	LogType  int32         //记录类型 0.获取 1.消耗
	ItemId   int32         //道具id
	ItemName string        //道具名称
	Count    int32         //个数
	CreateTs int64         //记录时间
	Remark   string        //备注
}

func NewItemLog() *ItemLog {
	log := &ItemLog{LogId: bson.NewObjectId()}
	return log
}
func NewItemLogEx(platform string, snId, logType, itemId int32, itemName string, count int32, remark string) *ItemLog {
	itemLog := NewItemLog()
	itemLog.Platform = platform
	itemLog.SnId = snId
	itemLog.LogType = logType
	itemLog.ItemId = itemId
	itemLog.ItemName = itemName
	itemLog.Count = count
	itemLog.CreateTs = time.Now().Unix()
	itemLog.Remark = remark
	return itemLog
}
