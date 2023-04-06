package model

import (
	"time"

	"github.com/globalsign/mgo/bson"
	"github.com/idealeak/goserver/core/logger"
)

var (
	CoinWALDBName   = "log"
	CoinWALCollName = "log_coinwal"
)

type CoinWAL struct {
	Id       bson.ObjectId `bson:"_id"`
	SnId     int32         //玩家id
	Count    int64         //帐变数量
	InGame   int32         //0：其他 1~N：具体游戏id
	SceneId  int32         //房间id
	CoinType int32         //金币类型 0:钱包 1:保险箱 2:俱乐部账户
	LogType  int32         //log类型
	CurTs    int64         //加个插入时间戳
	Ts       int64         //时间戳
}

func NewCoinWAL(snid int32, count int64, logType, inGame int32, cointype int32, roomid int32, ts int64) *CoinWAL {
	cl := &CoinWAL{Id: bson.NewObjectId()}
	cl.SnId = snid
	cl.Count = count
	cl.InGame = inGame
	cl.CoinType = cointype
	cl.LogType = logType
	cl.SceneId = roomid
	cl.Ts = ts
	cl.CurTs = time.Now().Unix()
	return cl
}

type CoinWALWithSnid_InGame_GreaterTsArgs struct {
	Plt    string
	SnId   int32
	RoomId int32
	Ts     int64
}

func GetCoinWALBySnidAndInGameAndGreaterTs(plt string, id int32, roomid int32, ts int64) (ret []CoinWAL, err error) {
	if rpcCli == nil {
		logger.Logger.Error("model.GetCoinWALBySnidAndInGameAndGreaterTs rpcCli == nil")
		return
	}
	args := &CoinWALWithSnid_InGame_GreaterTsArgs{
		Plt:    plt,
		SnId:   id,
		RoomId: roomid,
		Ts:     ts,
	}
	err = rpcCli.CallWithTimeout("CoinWALSvc.GetCoinWALBySnidAndInGameAndGreaterTs", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Warn("GetCoinWALBySnidAndInGameAndGreaterTs error:", err)
	}
	return
}

func RemoveCoinWALBySnidAndInGameAndGreaterTs(plt string, id int32, roomid int32, ts int64) (err error) {
	if rpcCli == nil {
		logger.Logger.Error("model.RemoveCoinWALBySnidAndInGameAndGreaterTs rpcCli == nil")
		return
	}
	args := &CoinWALWithSnid_InGame_GreaterTsArgs{
		Plt:    plt,
		SnId:   id,
		RoomId: roomid,
		Ts:     ts,
	}
	var ret bool
	err = rpcCli.CallWithTimeout("CoinWALSvc.RemoveCoinWALBySnidAndInGameAndGreaterTs", args, &ret, time.Second*30)
	if err != nil {
		logger.Logger.Warn("RemoveCoinWALBySnidAndInGameAndGreaterTs error:", err)
	}
	return
}
