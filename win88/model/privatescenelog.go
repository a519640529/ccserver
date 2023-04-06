package model

import (
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"time"
)

var (
	PrivateSceneLogDBName   = "log"
	PrivateSceneLogCollName = "log_privatescene"
)

type PrivateSceneLog struct {
	LogId       bson.ObjectId `bson:"_id"`
	SnId        int32         // 玩家id
	Platform    string        // 平台名称
	Channel     string        // 渠道名称
	Promoter    string        // 推广员
	PackageTag  string        // 推广包标识
	GameFreeId  int32         // 游戏id
	SceneId     int32         // 房间id
	CreateFee   int32         // 房费
	CreateTime  time.Time     // 创建时间
	DestroyTime time.Time     // 销毁时间
}

func NewPrivateSceneLog() *PrivateSceneLog {
	return &PrivateSceneLog{
		LogId: bson.NewObjectId(),
	}
}

func InsertPrivateSceneLog(log *PrivateSceneLog) error {
	if rpcCli == nil {
		return ErrRPClientNoConn
	}

	var ret bool
	return rpcCli.CallWithTimeout("PrivateSceneLogSvc.InsertPrivateSceneLog", log, &ret, time.Second*30)
}

type GetPrivateSceneLogBySnIdArgs struct {
	Plt   string
	SnId  int32
	Limit int
}

func GetPrivateSceneLogBySnId(plt string, snid int32, limit int) ([]*PrivateSceneLog, error) {
	if rpcCli == nil {
		return nil, ErrRPClientNoConn
	}
	args := &GetPrivateSceneLogBySnIdArgs{
		Plt:   plt,
		SnId:  snid,
		Limit: limit,
	}
	var logs []*PrivateSceneLog
	err := rpcCli.CallWithTimeout("PrivateSceneLogSvc.GetPrivateSceneLogBySnId", args, &logs, time.Second*30)
	return logs, err
}

type RemovePrivateSceneLogsArgs struct {
	Plt string
	Ts  time.Time
}

func RemovePrivateSceneLogs(plt string, ts time.Time) (*mgo.ChangeInfo, error) {
	if rpcCli == nil {
		return nil, ErrRPClientNoConn
	}

	args := &RemovePrivateSceneLogsArgs{
		Plt: plt,
		Ts:  ts,
	}
	var ret *mgo.ChangeInfo
	err := rpcCli.CallWithTimeout("PrivateSceneLogSvc.InsertPrivateSceneLog", args, &ret, time.Second*30)
	return ret, err
}
