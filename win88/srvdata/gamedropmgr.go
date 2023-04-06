package srvdata

import (
	"strconv"
)

var GameDropMgrSington = &GameDropMgr{
	GameDropData: make(map[string][]*GameDropData),
}

type GameDropMgr struct {
	GameDropData map[string][]*GameDropData
}

type GameDropData struct {
	ItemId    int32
	Rate      int32
	MinAmount int32
	MaxAmount int32
}

func (this *GameDropMgr) ModuleName() string {
	return "GameDropMgr"
}

func (this *GameDropMgr) GetKey(gameid, basescore int32) string {
	return strconv.FormatInt(int64(gameid), 10) + "_" + strconv.FormatInt(int64(basescore), 10)
}

func (this *GameDropMgr) Init() {
	gdArr := PBDB_Game_DropMgr.Datas.Arr
	if gdArr != nil {
		for _, drop := range gdArr {
			key := this.GetKey(drop.GameId, drop.Bet)
			//道具1
			if drop.Amount1 == nil || len(drop.Amount1) != 2 {
				continue
			}
			gdd1 := &GameDropData{
				ItemId:    drop.ItemId1,
				Rate:      drop.Rate1,
				MinAmount: drop.Amount1[0],
				MaxAmount: drop.Amount1[1],
			}
			this.GameDropData[key] = append(this.GameDropData[key], gdd1)
			//道具2
			//if drop.Amount2 == nil || len(drop.Amount2) != 2 {
			//	continue
			//}
			//gdd2 := &GameDropData{
			//	ItemId:    drop.ItemId2,
			//	Rate:      drop.Rate2,
			//	MinAmount: drop.Amount2[0],
			//	MaxAmount: drop.Amount2[1],
			//}
			//this.GameDropData[key] = append(this.GameDropData[key], gdd2)
		}
	}
}

func (this *GameDropMgr) GetDropInfoByBaseScore(gameid, basescore int32) []*GameDropData {
	key := this.GetKey(gameid, basescore)
	if gdds, exist := this.GameDropData[key]; exist {
		return gdds
	}
	return nil
}
