
// Code generated by xlsx2proto.
// DO NOT EDIT!

package srvdata

import (
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/server"
)

var PBDB_Legend_TurnRateMgr = &DB_Legend_TurnRateMgr{pool: make(map[int32]*server.DB_Legend_TurnRate), Datas: &server.DB_Legend_TurnRateArray{}}

type DB_Legend_TurnRateMgr struct {
	Datas *server.DB_Legend_TurnRateArray
	pool  map[int32]*server.DB_Legend_TurnRate
}

func (this *DB_Legend_TurnRateMgr) unmarshal(data []byte) error {
	err := proto.Unmarshal(data, this.Datas)
	if err == nil {
		this.arrangeData()
	}
	return err
}

func (this *DB_Legend_TurnRateMgr) reunmarshal(data []byte) error {
	newDatas := &server.DB_Legend_TurnRateArray{}
	err := proto.Unmarshal(data, newDatas)
	if err == nil {
		for _, item := range newDatas.Arr {
			existItem := this.GetData(item.GetId())
			if existItem == nil {
				this.pool[item.GetId()] = item
				this.Datas.Arr = append(this.Datas.Arr, item)
			} else {
				*existItem = *item
			}
		}
	}
	return err
}

func (this *DB_Legend_TurnRateMgr) arrangeData() {
	if this.Datas == nil {
		return
	}

	dataArr := this.Datas.GetArr()
	if dataArr == nil {
		return
	}

	for _, data := range dataArr {
		this.pool[data.GetId()] = data
	}
}

func (this *DB_Legend_TurnRateMgr) GetData(id int32) *server.DB_Legend_TurnRate {
	if data, ok := this.pool[id]; ok {
		return data
	}
	return nil
}

func init() {
	DataMgr.RegisteLoader("DB_Legend_TurnRate.dat", &ProtobufDataLoader{dh: PBDB_Legend_TurnRateMgr})
}
