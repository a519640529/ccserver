package srvdata

import (
	"games.yol.com/win88/protocol/server"
)

var CreateRoomMgrSington = &CreateRoomMgr{
	GameIdDatas: make(map[int32][]*server.DB_Createroom),
	Datas:       make(map[int32][]*server.DB_Createroom),
}

type CreateRoomMgr struct {
	GameIdDatas map[int32][]*server.DB_Createroom
	Datas       map[int32][]*server.DB_Createroom
}

func (this *CreateRoomMgr) Init() {
	this.Datas = make(map[int32][]*server.DB_Createroom)
	this.GameIdDatas = make(map[int32][]*server.DB_Createroom)
	for _, v := range PBDB_CreateroomMgr.Datas.GetArr() {
		gameid := v.GameId
		key := gameid*10000 + v.GameSite
		this.Datas[key] = append(this.Datas[key], v)
		this.GameIdDatas[gameid] = append(this.GameIdDatas[gameid], v)
	}
}

func (this *CreateRoomMgr) GetGameSiteByGameId(gameid int32, takeCoin int64) int32 {
	datas := this.GameIdDatas[gameid]
	if datas != nil && len(datas) != 0 {
		for i := len(datas) - 1; i >= 0; i-- {
			goldRange := datas[i].GoldRange
			if len(goldRange) != 0 && goldRange[0] != 0 {
				if takeCoin >= int64(goldRange[0]) {
					return datas[i].GetGameSite()
				}
			}
		}
	}
	return 0
}

func (this *CreateRoomMgr) GetDataByGameIdWithSite(gameid, gamesite int32) []*server.DB_Createroom {
	return this.Datas[gameid*10000+gamesite]
}

func (this *CreateRoomMgr) GetLimitCoinByBaseScore(gameid, gamesite, baseScore int32) int64 {
	limitCoin := int64(0)
	datas := this.GetDataByGameIdWithSite(gameid, gamesite)
	if datas != nil && len(datas) != 0 {
		tmpIds := []int32{}
		for i := 0; i < len(datas); i++ {
			data := datas[i]
			betRange := data.GetBetRange()
			if len(betRange) == 0 {
				continue
			}
			for j := 0; j < len(betRange); j++ {
				if betRange[j] == baseScore && len(data.GetGoldRange()) > 0 && data.GetGoldRange()[0] != 0 {
					tmpIds = append(tmpIds, data.GetId())
					break
				}
			}
		}
		if len(tmpIds) > 0 {
			goldRange := PBDB_CreateroomMgr.GetData(tmpIds[0]).GetGoldRange()
			if len(goldRange) != 0 && goldRange[0] != 0 {
				limitCoin = int64(goldRange[0])
			}
			if limitCoin != 0 {
				for _, id := range tmpIds {
					tmp := PBDB_CreateroomMgr.GetData(id).GetGoldRange()
					if int64(tmp[0]) < limitCoin && tmp[0] != 0 {
						limitCoin = int64(tmp[0])
					}
				}
			}
		}
	}
	return limitCoin
}
