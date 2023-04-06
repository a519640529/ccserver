package model

import (
	"time"

	"github.com/globalsign/mgo/bson"
)

var (
	SysProfitCoinDBName   = "user"
	SysProfitCoinCollName = "user_sysprofitcoin"
)

type SysProfitCoin struct {
	LogId      bson.ObjectId `bson:"_id"` //记录ID
	Key        string
	ProfitCoin map[string]*SysCoin //系统收益
}

type SysCoin struct {
	PlaysBet    int64 // 玩家总下注
	SysPushCoin int64 // 系统总产出
	CommonPool  int64 // 公共产出池 部分游戏用到
	Version     int32 // 数据版本号
}

var sysProfitCoin = &SysProfitCoin{
	ProfitCoin: make(map[string]*SysCoin),
}

func InitSysProfitCoinData(key string) *SysProfitCoin {
	if rpcCli != nil {
		var data *SysProfitCoin
		err := rpcCli.CallWithTimeout("SysProfitCoinSvc.InitSysProfitCoinData", key, &data, time.Second*30)
		if err != nil {
			return sysProfitCoin
		}
		sysProfitCoin = data
		if sysProfitCoin.ProfitCoin == nil {
			sysProfitCoin.ProfitCoin = make(map[string]*SysCoin)
		}
	}
	return sysProfitCoin
}

func UptCoinSysProfitCoinData(key string, data *SysCoin) {
	sysProfitCoin.ProfitCoin[key] = data
}

// 保存
func SaveSysProfitCoin(data *SysProfitCoin) error {
	if rpcCli == nil {
		return ErrRPClientNoConn
	}
	sysProfitCoin = data
	var ret bool
	return rpcCli.CallWithTimeout("SysProfitCoinSvc.SaveSysProfitCoin", sysProfitCoin, &ret, time.Second*30)
}
