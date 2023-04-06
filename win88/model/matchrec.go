package model

import (
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

type MatchRec struct {
	Id         bson.ObjectId `bson:"_id"`
	Platform   string
	SnId       int32
	MatchId    int32
	MatchName  string
	GameFreeId int32
	Rank       int32
	SignupType int32
	SignupCnt  int32
	Prizes     int32
	Grade      int32
	Ts         time.Time
}

var (
	MatchRecDBName   = "log"
	MatchRecCollName = "log_matchrec"
)

func NewMatchRec() *MatchRec {
	return &MatchRec{Id: bson.NewObjectId()}
}

func InsertMatchRecs(logs ...*MatchRec) (err error) {
	if rpcCli == nil {
		return ErrRPClientNoConn
	}
	var ret bool
	return rpcCli.CallWithTimeout("MatchRecSvc.InsertMatchRecs", logs, &ret, time.Second*30)
}

type FetchMatchRecsArgs struct {
	Plt  string
	SnId int32
}

func FetchMatchRecs(plt string, snid int32) (recs []MatchRec, err error) {
	if rpcCli == nil {
		return nil, ErrRPClientNoConn
	}
	args := &FetchMatchRecsArgs{
		Plt:  plt,
		SnId: snid,
	}
	err = rpcCli.CallWithTimeout("MatchRecSvc.FetchMatchRecs", args, &recs, time.Second*30)
	return
}

type RemoveMatchRecsArgs struct {
	Plt string
	Ts  time.Time
}

func RemoveMatchRecs(plt string, ts time.Time) (ret *mgo.ChangeInfo, err error) {
	if rpcCli == nil {
		return nil, ErrRPClientNoConn
	}
	args := &RemoveMatchRecsArgs{
		Plt: plt,
		Ts:  ts,
	}
	rpcCli.CallWithTimeout("MatchRecSvc.RemoveMatchRecs", args, &ret, time.Second*30)
	return
}
