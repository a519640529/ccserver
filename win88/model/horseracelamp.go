package model

import (
	"errors"
	"github.com/globalsign/mgo/bson"
	"time"
)

// 跑马灯
var (
	HorseRaceLampDBName   = "user"
	HorseRaceLampCollName = "user_horseracelamp"
)

type RemoveHorseRaceLampArg struct {
	Platform string
	Key      string
}

type RemoveHorseRaceLampRet struct {
	Data *interface{}
	Tag  int
}
type HorseRaceLamp struct {
	Id         bson.ObjectId `bson:"_id"`
	Channel    string        //渠道
	Title      string        //标题
	Content    string        //公告内容
	Footer     string
	StartTime  int64  //开始播放的时间
	Interval   int32  //播放间隔
	Count      int32  //播放次数
	CreateTime int64  //创建公告的时间
	Priority   int32  //播放优先级
	MsgType    int32  //公告类型
	Platform   string //公告播放的平台
	State      int32  //状态 0.启用；1.关闭
	Target     []int32
	StandSec   int32
}

func NewHorseRaceLamp(channel, platform, title, content, footer string, startTime int64, interval int32, count int32, priority, state int32,
	msgType int32, target []int32, standSec int32) *HorseRaceLamp {
	horseracelamp := &HorseRaceLamp{
		Id:         bson.NewObjectId(),
		Channel:    channel,
		Title:      title,
		Content:    content,
		Footer:     footer,
		StartTime:  startTime,
		Interval:   interval,
		Count:      count,
		Priority:   priority,
		CreateTime: time.Now().Unix(),
		MsgType:    msgType,
		Platform:   platform,
		State:      state,
		Target:     target,
		StandSec:   standSec,
	}
	return horseracelamp
}

type InsertHorseRaceLampArgs struct {
	HorseRaceLamps []*HorseRaceLamp
	Platform       string
}

func InsertHorseRaceLamp(platform string, notices ...*HorseRaceLamp) (err error) {
	if len(notices) == 0 {
		return errors.New("no data")
	}
	if rpcCli == nil {
		return ErrRPClientNoConn
	}
	args := InsertHorseRaceLampArgs{
		HorseRaceLamps: notices,
		Platform:       platform,
	}
	var ret bool
	err = rpcCli.CallWithTimeout("HorseRaceLampSvc.InsertHorseRaceLamp", args, &ret, time.Second*30)
	if err != nil {
		return err
	}
	return
}

type GetHorseRaceLampArgs struct {
	Platform string
	Id       bson.ObjectId
}

func GetHorseRaceLamp(platform string, Id bson.ObjectId) (ret *HorseRaceLamp, err error) {
	args := GetHorseRaceLampArgs{
		Platform: platform,
		Id:       Id,
	}
	err = rpcCli.CallWithTimeout("HorseRaceLampSvc.GetHorseRaceLamp", args, &ret, time.Second*30)
	if err != nil {
		return nil, err
	}
	return
}

func RemoveHorseRaceLamp(plt, key string) (err error) {
	if rpcCli == nil {
		return ErrRPClientNoConn
	}

	args := &RemoveHorseRaceLampArg{Key: key, Platform: plt}
	var ret bool
	err = rpcCli.CallWithTimeout("HorseRaceLampSvc.RemoveHorseRaceLamp", args, &ret, time.Second*30)
	if err != nil {
		return err
	}
	return nil
}

/*
func StopHorseRaceLamp(key string) (err error) {
	Id := bson.ObjectIdHex(key)
	err = HorseRaceLampCollection().Update(bson.M{"_id": Id}, bson.D{{"$set", bson.D{{"state", NOTICESTATE_STOP}}}})
	return err
}*/

func GetAllHorseRaceLamp(plt string) (notice []HorseRaceLamp, err error) {
	if rpcCli == nil {
		return nil, ErrRPClientNoConn
	}
	err = rpcCli.CallWithTimeout("HorseRaceLampSvc.GetAllHorseRaceLamp", plt, &notice, time.Second*30)
	if err != nil {
		return nil, err
	}
	return notice, nil
}

type EditHorseRaceLampArg struct {
	Key       string
	Channel   string
	Platform  string
	Title     string
	Content   string
	Footer    string
	StartTime int64
	Interval  int32
	Count     int32
	Priority  int32
	MsgType   int32
	State     int32
	Target    []int32
	StandSec  int32
}
type EditHorseRaceLampRet struct {
	Tag int
}

func EditHorseRaceLamp(hrl *HorseRaceLamp) (err error) {
	if rpcCli == nil {
		return ErrRPClientNoConn
	}
	args := &EditHorseRaceLampArg{
		Key:       hrl.Id.Hex(),
		Channel:   hrl.Channel,
		Platform:  hrl.Platform,
		Title:     hrl.Title,
		Content:   hrl.Content,
		Footer:    hrl.Footer,
		StartTime: hrl.StartTime,
		Interval:  hrl.Interval,
		Count:     hrl.Count,
		Priority:  hrl.Priority,
		MsgType:   hrl.MsgType,
		State:     hrl.State,
		Target:    hrl.Target,
		StandSec:  hrl.StandSec,
	}
	var ret bool
	err = rpcCli.CallWithTimeout("HorseRaceLampSvc.EditHorseRaceLamp", args, &ret, time.Second*30)
	return err
}

type QueryHorseRaceLampArg struct {
	Platform     string
	MsgType      int32
	State        int32
	ToIndex      int32
	FromIndex    int
	LimitDataNum int
}
type QueryHorseRaceLampRet struct {
	Data  []HorseRaceLamp
	Count int
}

func GetHorseRaceLampInRangeTsLimitByRange(platform string, msgType, state, fromIndex,
	toIndex int32) (ret *QueryHorseRaceLampRet, err error) {
	if rpcCli == nil {
		return nil, ErrRPClientNoConn
	}
	limitDataNum := toIndex - fromIndex
	if limitDataNum < 0 {
		limitDataNum = 0
	}
	args := &QueryHorseRaceLampArg{
		Platform:     platform,
		MsgType:      msgType,
		State:        state,
		FromIndex:    int(fromIndex),
		ToIndex:      toIndex,
		LimitDataNum: int(limitDataNum),
	}
	err = rpcCli.CallWithTimeout("HorseRaceLampSvc.GetHorseRaceLampInRangeTsLimitByRange", args, &ret, time.Second*30)
	return
}
