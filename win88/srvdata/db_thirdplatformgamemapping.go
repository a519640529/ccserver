
// Code generated by xlsx2proto.
// DO NOT EDIT!

package srvdata

import (
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/server"
)

var PBDB_ThirdPlatformGameMappingMgr = &DB_ThirdPlatformGameMappingMgr{pool: make(map[int32]*server.DB_ThirdPlatformGameMapping), Datas: &server.DB_ThirdPlatformGameMappingArray{}}

type DB_ThirdPlatformGameMappingMgr struct {
	Datas *server.DB_ThirdPlatformGameMappingArray
	pool  map[int32]*server.DB_ThirdPlatformGameMapping
}

func (this *DB_ThirdPlatformGameMappingMgr) unmarshal(data []byte) error {
	err := proto.Unmarshal(data, this.Datas)
	if err == nil {
		this.arrangeData()
	}
	return err
}

func (this *DB_ThirdPlatformGameMappingMgr) reunmarshal(data []byte) error {
	newDatas := &server.DB_ThirdPlatformGameMappingArray{}
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

func (this *DB_ThirdPlatformGameMappingMgr) arrangeData() {
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

func (this *DB_ThirdPlatformGameMappingMgr) GetData(id int32) *server.DB_ThirdPlatformGameMapping {
	if data, ok := this.pool[id]; ok {
		return data
	}
	return nil
}

func init() {
	DataMgr.RegisteLoader("DB_ThirdPlatformGameMapping.dat", &ProtobufDataLoader{dh: PBDB_ThirdPlatformGameMappingMgr})
}
