
// Code generated by xlsx2proto.
// DO NOT EDIT!

package srvdata

import (
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/server"
)

var PBDB_NameMgr = &DB_NameMgr{pool: make(map[int32]*server.DB_Name), Datas: &server.DB_NameArray{}}

type DB_NameMgr struct {
	Datas *server.DB_NameArray
	pool  map[int32]*server.DB_Name
}

func (this *DB_NameMgr) unmarshal(data []byte) error {
	err := proto.Unmarshal(data, this.Datas)
	if err == nil {
		this.arrangeData()
	}
	return err
}

func (this *DB_NameMgr) reunmarshal(data []byte) error {
	newDatas := &server.DB_NameArray{}
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

func (this *DB_NameMgr) arrangeData() {
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

func (this *DB_NameMgr) GetData(id int32) *server.DB_Name {
	if data, ok := this.pool[id]; ok {
		return data
	}
	return nil
}

func init() {
	DataMgr.RegisteLoader("DB_Name.dat", &ProtobufDataLoader{dh: PBDB_NameMgr})
}
