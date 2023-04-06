
// Code generated by xlsx2proto.
// DO NOT EDIT!

package srvdata

import (
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/server"
)

var PBDB_AnimalColorMgr = &DB_AnimalColorMgr{pool: make(map[int32]*server.DB_AnimalColor), Datas: &server.DB_AnimalColorArray{}}

type DB_AnimalColorMgr struct {
	Datas *server.DB_AnimalColorArray
	pool  map[int32]*server.DB_AnimalColor
}

func (this *DB_AnimalColorMgr) unmarshal(data []byte) error {
	err := proto.Unmarshal(data, this.Datas)
	if err == nil {
		this.arrangeData()
	}
	return err
}

func (this *DB_AnimalColorMgr) reunmarshal(data []byte) error {
	newDatas := &server.DB_AnimalColorArray{}
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

func (this *DB_AnimalColorMgr) arrangeData() {
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

func (this *DB_AnimalColorMgr) GetData(id int32) *server.DB_AnimalColor {
	if data, ok := this.pool[id]; ok {
		return data
	}
	return nil
}

func init() {
	DataMgr.RegisteLoader("DB_AnimalColor.dat", &ProtobufDataLoader{dh: PBDB_AnimalColorMgr})
}
