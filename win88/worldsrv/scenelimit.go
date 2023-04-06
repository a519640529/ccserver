package main

import (
	"fmt"
	"games.yol.com/win88/common"
)

var sceneLimitMgr = SceneLimitManager{
	SameRoomData: make(map[string][][]int32),
}

type SceneLimitManager struct {
	SameRoomData map[string][][]int32 //用户的同房数据，保存了用户一局游戏中，房间人员的SNID
}

func (this *SceneLimitManager) Key(gameid, snid int32) string {
	return fmt.Sprintf("%v-%v", snid, gameid)
}

// 和同一个人5局之内不能同房间游戏
func (this *SceneLimitManager) LimitSamePlace(player *Player, s *Scene) bool {
	if player.IsRob {
		return false
	}
	key := this.Key(int32(s.gameId), player.SnId)
	var logArr = this.SameRoomData[key]
	var samePlaceLimit = int(s.dbGameFree.GetSamePlaceLimit())
	if len(logArr) > samePlaceLimit {
		logArr = logArr[len(logArr)-samePlaceLimit:]
		this.SameRoomData[key] = logArr
	}
	//新用户的防伙牌数据中，有没有场景中用户的检查
	for _, value := range s.players {
		if value.IsRob {
			continue
		}
		limitCount := 0
		for _, log := range logArr {
			if common.InSliceInt32(log, value.SnId) {
				limitCount++
			}
		}
		if limitCount >= samePlaceLimit {
			return true
		}
	}
	//场景中用户的防伙牌数据中，有没有新用户的检查
	for _, value := range s.players {
		if value.IsRob {
			continue
		}
		key := this.Key(int32(s.gameId), value.SnId)
		var logArr = this.SameRoomData[key]
		limitCount := 0
		for _, log := range logArr {
			if common.InSliceInt32(log, player.SnId) {
				limitCount++
			}
		}
		if limitCount >= samePlaceLimit {
			return true
		}
	}
	return false
}
func (this *SceneLimitManager) LimitSamePlaceBySnid(member []*Player, player *Player, gameId, limit int32) bool {
	if player.IsRob {
		return false
	}
	key := this.Key(gameId, player.SnId)
	var logArr = this.SameRoomData[key]
	var samePlaceLimit = int(limit)
	if len(logArr) > samePlaceLimit {
		logArr = logArr[len(logArr)-samePlaceLimit:]
		this.SameRoomData[key] = logArr
	}
	//新用户的防伙牌数据中，有没有场景中用户的检查
	for _, value := range member {
		if value.IsRob {
			continue
		}
		limitCount := 0
		for _, log := range logArr {
			if common.InSliceInt32(log, value.SnId) {
				limitCount++
			}
		}
		if limitCount >= samePlaceLimit {
			return true
		}
	}
	//场景中用户的防伙牌数据中，有没有新用户的检查
	for _, value := range member {
		if value.IsRob {
			continue
		}
		key := this.Key(gameId, value.SnId)
		var logArr = this.SameRoomData[key]
		limitCount := 0
		for _, log := range logArr {
			if common.InSliceInt32(log, player.SnId) {
				limitCount++
			}
		}
		if limitCount >= samePlaceLimit {
			return true
		}
	}
	return false
}
func (this *SceneLimitManager) ReciveData(gameid int32, data []int32) {
	for _, value := range data {
		key := this.Key(gameid, value)
		this.SameRoomData[key] = append(this.SameRoomData[key], data)
	}
}

// 人数的平均分配问题
func (this *SceneLimitManager) LimitAvgPlayer(s *Scene, totlePlayer int) bool {
	scenePlayerCount := s.GetTruePlayerCnt()
	switch {
	case totlePlayer > 1 && totlePlayer < 15:
		//4、如果游戏场的同时在线人数2-14人时，系统每2个人分一个桌，如果有剩余的没有分桌，随机找一个桌子进行分配；
		// 如果有机器人，系统加入机器人，并且保证同桌机器人数量1-3个
		if scenePlayerCount > 2 {
			return true
		}
	case totlePlayer >= 15 && totlePlayer < 22:
		//5、如果游戏场的同时在线人数15-21人，系统每3个人分一个桌，如果有剩余的没有分桌，随机找一个桌子进行分配；
		// 如果有机器人，系统加入机器人，并且保证同桌机器人数量1-3个；
		if scenePlayerCount > 3 {
			return true
		}
	case totlePlayer >= 22 && totlePlayer < 29:
		//6、如果游戏场的同时在线人数22-28人，系统每4个人分一个桌，如果有剩余的没有分桌，随机找一个桌子进行分配；
		// 如果有机器人，系统加入机器人，并且保证同桌机器人数量1-3个；
		if scenePlayerCount > 4 {
			return true
		}
	case totlePlayer >= 29 && totlePlayer < 35:
		//7、如果游戏场的同时在线人数29-35人，系统每5个人分一个桌，如果有剩余的没有分桌，随机找一个桌子进行分配；
		if scenePlayerCount > 5 {
			return true
		}
	case totlePlayer >= 35 && totlePlayer < 43:
		//8、如果游戏场的同时在线人数36-42人，系统每6个人分一个桌，如果有剩余的没有分桌，随机找一个桌子进行分配；
		if scenePlayerCount > 6 {
			return true
		}
	case totlePlayer >= 35 && totlePlayer < 43:
		//9、如果游戏场的同时在线人数43人以上时，系统每7个人分一个桌，如果有剩余的没有分桌，随机找一个桌子进行分配；
		if scenePlayerCount > 5 {
			return true
		}
	case totlePlayer >= 43:
		if scenePlayerCount > 5 {
			return true
		}
	}
	return false
}
