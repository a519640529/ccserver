package srvdata

import "games.yol.com/win88/protocol/server"

var PlayerTypeMgrSington = &PlayerTypeMgr{
	types: make(map[int32][]*server.DB_PlayerType),
}

type PlayerTypeMgr struct {
	types map[int32][]*server.DB_PlayerType
}

func (this *PlayerTypeMgr) updateData() {
	types := make(map[int32][]*server.DB_PlayerType)
	for _, data := range PBDB_PlayerTypeMgr.Datas.Arr {
		types[data.GetGameFreeId()] = append(types[data.GetGameFreeId()], data)
	}
	this.types = types
}

func (this *PlayerTypeMgr) GetPlayerType(id int32) []*server.DB_PlayerType {
	if data, exist := this.types[id]; exist {
		return data
	}
	return nil
}
