package model

//排行榜数据
import (
	"github.com/globalsign/mgo/bson"
	"github.com/idealeak/goserver/core/logger"
	"sort"
	"time"
)

type RankPlayerData struct {
	SnId int32
	Name string
	Val  int64
}

type RankData struct {
	Id   bson.ObjectId
	Key  string
	Data []*RankPlayerData //排行榜内已排好顺序的玩家信息
}

// 适用于小榜,比如总榜数量<=100,大榜适合用跳表实现
func (rd *RankData) Insert(data *RankPlayerData, add bool) {
	cnt := len(rd.Data)
	idx := sort.Search(cnt, func(i int) bool {
		return rd.Data[i].Val <= data.Val
	})
	if add {
		rd.Data = append(rd.Data, nil)
	}
	if idx != cnt {
		copy(rd.Data[idx+1:], rd.Data[idx:cnt])
	}
	cnt = len(rd.Data)
	//索引确保
	if idx >= cnt {
		idx = cnt - 1
	}
	//当前数据放到idx位置
	rd.Data[idx] = data
}

func (rd *RankData) Sort() {
	less := func(i, j int) bool { //倒序
		return rd.Data[i].Val > rd.Data[j].Val
	}
	if !sort.SliceIsSorted(rd.Data, less) {
		sort.Slice(rd.Data, less)
	}
}

func (rd *RankData) Clone() *RankData {
	nrd := &RankData{
		Id:   rd.Id,
		Key:  rd.Key,
		Data: make([]*RankPlayerData, len(rd.Data)),
	}
	copy(nrd.Data, rd.Data)
	return nrd
}

func InitRankData(plt string) []*RankData {
	if rpcCli == nil {
		return nil
	}
	var data []*RankData
	err := rpcCli.CallWithTimeout("RankDataSvc.InitRankData", plt, &data, time.Second*30)
	if err != nil {
		logger.Logger.Errorf("InitRankDBData err:%v", err)
		return nil
	}

	return data
}

type SaveRankDataArgs struct {
	Plt  string
	Key  string
	Data *RankData
}

func SaveRankData(plt, key string, data *RankData) error {
	if rpcCli == nil {
		return ErrRPClientNoConn
	}

	args := &SaveRankDataArgs{
		Plt:  plt,
		Key:  key,
		Data: data,
	}
	var ret bool
	return rpcCli.CallWithTimeout("RankDataSvc.SaveRankData", args, &ret, time.Second*30)
}
