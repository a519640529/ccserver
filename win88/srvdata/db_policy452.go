
// Code generated by xlsx2proto.
// DO NOT EDIT!

package srvdata

import (
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/server"
)

var PBDB_Policy452Mgr = &DB_Policy452Mgr{pool: make(map[int32]*server.DB_Policy452), Datas: &server.DB_Policy452Array{}}

type DB_Policy452Mgr struct {
	Datas *server.DB_Policy452Array
	pool  map[int32]*server.DB_Policy452
}

func (this *DB_Policy452Mgr) unmarshal(data []byte) error {
	err := proto.Unmarshal(data, this.Datas)
	if err == nil {
		this.arrangeData()
	}
	return err
}

func (this *DB_Policy452Mgr) reunmarshal(data []byte) error {
	newDatas := &server.DB_Policy452Array{}
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

func (this *DB_Policy452Mgr) arrangeData() {
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

func (this *DB_Policy452Mgr) GetData(id int32) *server.DB_Policy452 {
	if data, ok := this.pool[id]; ok {
		return data
	}
	return nil
}

func init() {
	DataMgr.RegisteLoader("DB_Policy452.dat", &ProtobufDataLoader{dh: PBDB_Policy452Mgr})
}
