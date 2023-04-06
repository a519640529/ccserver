package model

import (
	"github.com/idealeak/goserver/core/logger"
	"time"
)

type ActMonitor struct {
	SeqNo       int64
	SnId        int32
	Platform    string //平台
	MonitorType int32  //二进制 1.登录 2.兑换 3.游戏
	CreateTime  int64  //创建时间
	Creator     string //创建者
	ReMark      string //备注
}

func NewActMonitor() *ActMonitor {
	log := &ActMonitor{}
	return log
}

func NewActMonitorEx(snid int32, platform string, MonitorType int32, Creator, ReMark string) *ActMonitor {
	cl := NewActMonitor()
	cl.SnId = snid
	cl.Platform = platform
	cl.MonitorType = MonitorType
	cl.CreateTime = time.Now().Unix()
	cl.Creator = Creator
	cl.ReMark = ReMark
	return cl
}

type ActMonitoRet struct {
	Err  error
	Data []*ActMonitor
}

func GetAllActMonitorData() (data []*ActMonitor) {
	if rpcCli == nil {
		return
	}
	ret := &ActMonitoRet{}
	err := rpcCli.CallWithTimeout("ActMonitorSvc.GetAllActMonitorData", 0, ret, time.Second*30)
	if err != nil {
		logger.Logger.Error("Get ActMonitor data eror.", err)
	}
	return
}
func UpsertSignleActMonitor(am *ActMonitor) (err error) {
	if rpcCli == nil {
		return
	}
	ret := &ActMonitoRet{}
	err = rpcCli.CallWithTimeout("ActMonitorSvc.UpsertSignleActMonitor", am, ret, time.Second*30)
	if err != nil {
		logger.Logger.Error("model.UpsertSignleActMonitor eror.", err)
	}
	return
}

func UpdateSignleActMonitor(am *ActMonitor) (err error) {
	if rpcCli == nil {
		return
	}
	ret := &ActMonitoRet{}
	err = rpcCli.CallWithTimeout("ActMonitorSvc.UpdateSignleActMonitor", am, ret, time.Second*30)
	if err != nil {
		logger.Logger.Error("model.UpdateSignleActMonitor eror.", err)
	}
	return
}

func RemoveActMonitorOne(seqno int) error {
	if rpcCli == nil {
		return nil
	}
	ret := &ActMonitoRet{}
	err := rpcCli.CallWithTimeout("ActMonitorSvc.RemoveActMonitorOne", seqno, ret, time.Second*30)
	if err != nil {
		logger.Logger.Error("model.RemoveActMonitorOne eror.", err)
		return err
	}
	return nil
}
