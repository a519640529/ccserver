package model

import (
	"github.com/globalsign/mgo/bson"
	"github.com/idealeak/goserver/core/logger"
	"regexp"
	"sync"
	"time"
)

var gameKVDatas = sync.Map{}

type GameKVData struct {
	DataId   bson.ObjectId `bson:"_id"`
	Key      string
	IntVal   int64
	FloatVal float64
	StrVal   string
	IntArr   []int64
}

func InitGameKVData() error {
	if rpcCli == nil {
		logger.Logger.Errorf("model.InitGameKVData rpcCli == nil")
		return nil
	}

	var args struct{}
	var datas []*GameKVData
	err := rpcCli.CallWithTimeout("GameKVDataSvc.GetAllGameKVData", args, &datas, time.Second*30)
	if err != nil {
		return err
	}

	for i := 0; i < len(datas); i++ {
		gameKVDatas.Store(datas[i].Key, datas[i])
	}

	//@test code
	//for i := 0; i < 100; i++ {
	//	UptIntKVGameData(strconv.Itoa(i), int64(i))
	//}
	//@test code
	return nil
}

func GetIntKVGameData(key string) int64 {
	if val, exist := gameKVDatas.Load(key); exist {
		if data, ok := val.(*GameKVData); ok {
			return data.IntVal
		}
	}
	return 0
}

func GetStrKVGameData(key string) string {
	if val, exist := gameKVDatas.Load(key); exist {
		if data, ok := val.(*GameKVData); ok {
			return data.StrVal
		}
	}
	return ""
}

func GetFloatKVGameData(key string) float64 {
	if val, exist := gameKVDatas.Load(key); exist {
		if data, ok := val.(*GameKVData); ok {
			return data.FloatVal
		}
	}
	return 0
}

func GetIntArrKVGameData(key string) []int64 {
	if val, exist := gameKVDatas.Load(key); exist {
		if data, ok := val.(*GameKVData); ok {
			return data.IntArr
		}
	}
	return []int64{}
}

func MatchStrKVGameData(expr string) map[string]*GameKVData {
	regExpr, err := regexp.Compile(expr)
	if err == nil {
		ret := make(map[string]*GameKVData)
		gameKVDatas.Range(func(k, v interface{}) bool {
			if regExpr.MatchString(k.(string)) {
				ret[k.(string)] = v.(*GameKVData)
			}
			return true
		})
		return ret
	}

	return nil
}

func UptIntKVGameData(key string, val int64) error {
	if rpcCli == nil {
		logger.Logger.Error("model.UptIntKVGameData rpcCli == nil")
		return nil
	}
	if value, exist := gameKVDatas.Load(key); exist {
		if data, ok := value.(*GameKVData); ok {
			data.IntVal = val
			var ok bool
			return rpcCli.CallWithTimeout("GameKVDataSvc.UptGameKVData", data, &ok, time.Second*30)
		}
	} else {
		data := &GameKVData{
			DataId: bson.NewObjectId(),
			Key:    key,
			IntVal: val,
		}
		gameKVDatas.Store(key, data)
		var ok bool
		return rpcCli.CallWithTimeout("GameKVDataSvc.UptGameKVData", data, &ok, time.Second*30)
	}
	return nil
}

func UptFloatKVGameData(key string, val float64) error {
	if value, exist := gameKVDatas.Load(key); exist {
		if data, ok := value.(*GameKVData); ok {
			data.FloatVal = val
			var ok bool
			return rpcCli.CallWithTimeout("GameKVDataSvc.UptGameKVData", data, &ok, time.Second*30)
		}
	} else {
		data := &GameKVData{
			DataId:   bson.NewObjectId(),
			Key:      key,
			FloatVal: val,
		}
		gameKVDatas.Store(key, data)
		var ok bool
		return rpcCli.CallWithTimeout("GameKVDataSvc.UptGameKVData", data, &ok, time.Second*30)
	}
	return nil
}

func UptStrKVGameData(key string, val string) error {
	if rpcCli == nil {
		logger.Logger.Error("model.UptStrKVGameData rpcCli == nil")
		return nil
	}
	if value, exist := gameKVDatas.Load(key); exist {
		if data, ok := value.(*GameKVData); ok {
			data.StrVal = val
			var ok bool
			return rpcCli.CallWithTimeout("GameKVDataSvc.UptGameKVData", data, &ok, time.Second*30)
		}
	} else {
		data := &GameKVData{
			DataId: bson.NewObjectId(),
			Key:    key,
			StrVal: val,
		}
		gameKVDatas.Store(key, data)
		var ok bool
		return rpcCli.CallWithTimeout("GameKVDataSvc.UptGameKVData", data, &ok, time.Second*30)
	}
	return nil
}

func UptIntArrKVGameData(key string, arr []int64) error {
	if value, exist := gameKVDatas.Load(key); exist {
		if data, ok := value.(*GameKVData); ok {
			data.IntArr = arr
			var ok bool
			return rpcCli.CallWithTimeout("GameKVDataSvc.UptGameKVData", data, &ok, time.Second*30)
		}
	} else {
		data := &GameKVData{
			DataId: bson.NewObjectId(),
			Key:    key,
			IntArr: arr,
		}
		gameKVDatas.Store(key, data)
		var ok bool
		return rpcCli.CallWithTimeout("GameKVDataSvc.UptGameKVData", data, &ok, time.Second*30)
	}
	return nil
}
