
// Code generated by xlsx2proto.
// DO NOT EDIT!

package srvdata

import (
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/server"
)

var PBDB_ShopMgr = &DB_ShopMgr{pool: make(map[int32]*server.DB_Shop), Datas: &server.DB_ShopArray{}}

type DB_ShopMgr struct {
	Datas *server.DB_ShopArray
	pool  map[int32]*server.DB_Shop
}

func (this *DB_ShopMgr) unmarshal(data []byte) error {
	err := proto.Unmarshal(data, this.Datas)
	if err == nil {
		this.arrangeData()
	}
	return err
}

func (this *DB_ShopMgr) reunmarshal(data []byte) error {
	newDatas := &server.DB_ShopArray{}
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

func (this *DB_ShopMgr) arrangeData() {
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

func (this *DB_ShopMgr) GetData(id int32) *server.DB_Shop {
	if data, ok := this.pool[id]; ok {
		return data
	}
	return nil
}

func init() {
	DataMgr.RegisteLoader("DB_Shop.dat", &ProtobufDataLoader{dh: PBDB_ShopMgr})
}
