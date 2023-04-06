package model

import (
	"time"
)

var (
	BankBindDBName   = "log"
	BankBindCollName = "log_bankbind"
)

const (
	BankBindLogType_Bank int32 = iota
	BankBindLogType_Ali
)

type BankBindLog struct {
	Snid     int32
	Platform string
	LogType  int32
	Name     string
	Card     string
	Modify   int32
	Time     time.Time
}

func NewBankBindLog(snid int32, platform string, logtype int32, name, card string, modify int32) error {
	if rpcCli == nil {
		return ErrRPClientNoConn
	}
	log := &BankBindLog{
		Snid:     snid,
		Platform: platform,
		LogType:  logtype,
		Name:     name,
		Card:     card,
		Modify:   modify,
		Time:     time.Now(),
	}
	var ret bool
	return rpcCli.CallWithTimeout("BankBindLogSvc.InsertBankBindLog", log, &ret, time.Second*30)
}
