package base

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"games.yol.com/win88/model"

	"github.com/globalsign/mgo/bson"
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/module"
	"github.com/idealeak/goserver/core/task"
)

// HundredJackListManager 排行榜 key: platform+gamefreeId
type HundredJackListManager struct {
	HundredJackTsList   map[string][]*HundredJackInfo
	HundredJackSortList map[string][]*HundredJackInfo
}

// HundredJackListMgr 实例化
var HundredJackListMgr = &HundredJackListManager{
	HundredJackTsList:   make(map[string][]*HundredJackInfo),
	HundredJackSortList: make(map[string][]*HundredJackInfo),
}

// HundredJackInfo 数据结构
type HundredJackInfo struct {
	model.HundredjackpotLog
	linkSnids []int32 //点赞人数
}

// ModuleName .
func (hm *HundredJackListManager) ModuleName() string {
	return "HundredJackListManager"
}

// Init .
func (hm *HundredJackListManager) Init() {
	//data := model.
}

// Update .
func (hm *HundredJackListManager) Update() {
}

// Shutdown .
func (hm *HundredJackListManager) Shutdown() {
	module.UnregisteModule(hm)
}

// ISInitJackInfo 仅初始化一次
var ISInitJackInfo bool

// InitTsJackInfo 初始化TsJackInfo
func (hm *HundredJackListManager) InitTsJackInfo(platform string, freeID int32) {
	key := fmt.Sprintf("%v-%v", platform, freeID)
	task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
		datas, err := model.GetHundredjackpotLogTsByPlatformAndGameFreeID(platform, freeID)
		if err != nil {
			logger.Logger.Error("HundredJackListManager DelOneJackInfo ", err)
			return nil
		}
		return datas
	}), task.CompleteNotifyWrapper(func(data interface{}, tt task.Task) {
		datas := data.([]model.HundredjackpotLog)
		if data != nil && datas != nil {
			for i := range datas {
				if i == model.HundredjackpotLogMaxLimitPerQuery {
					break
				}
				data := &HundredJackInfo{
					HundredjackpotLog: datas[i],
				}
				strlikeSnids := strings.Split(datas[i].LinkeSnids, "|")
				for _, v := range strlikeSnids {
					if v == "" {
						break
					}
					snid, err := strconv.Atoi(v)
					if err == nil {
						data.linkSnids = append(data.linkSnids, int32(snid))
					}
				}
				hm.HundredJackTsList[key] = append(hm.HundredJackTsList[key], data)
			}
			// logger.Logger.Warnf("InitTsJackInfo  data:%v", datas)
		} else {
			hm.HundredJackTsList[key] = []*HundredJackInfo{}
		}
		return
	}), "InitTsJackInfo").Start()
}

// InitSortJackInfo 初始化SortJackInfo
func (hm *HundredJackListManager) InitSortJackInfo(platform string, freeID int32) {

	key := fmt.Sprintf("%v-%v", platform, freeID)
	task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
		datas, err := model.GetHundredjackpotLogCoinByPlatformAndGameFreeID(platform, freeID)
		if err != nil {
			logger.Logger.Error("HundredJackListManager DelOneJackInfo ", err)
			return nil
		}
		return datas
	}), task.CompleteNotifyWrapper(func(data interface{}, tt task.Task) {
		datas := data.([]model.HundredjackpotLog)
		if data != nil && datas != nil {
			for i := range datas {
				if i == model.HundredjackpotLogMaxLimitPerQuery {
					break
				}
				data := &HundredJackInfo{
					HundredjackpotLog: datas[i],
				}
				strlikeSnids := strings.Split(datas[i].LinkeSnids, "|")
				for _, v := range strlikeSnids {
					snid, err := strconv.Atoi(v)
					if err == nil {
						data.linkSnids = append(data.linkSnids, int32(snid))
					}
				}
				hm.HundredJackSortList[key] = append(hm.HundredJackSortList[key], data)
			}
			// logger.Logger.Warnf("InitSortJackInfo  data:%v", datas)
		} else {
			hm.HundredJackSortList[key] = []*HundredJackInfo{}
		}
		return
	}), "InitSortJackInfo").Start()
}

// InitHundredJackListInfo 初始化 HundredJackListInfo
func (hm *HundredJackListManager) InitHundredJackListInfo(platform string, freeID int32) {
	if ISInitJackInfo {
		return
	}
	key := fmt.Sprintf("%v-%v", platform, freeID)
	if _, exist := hm.HundredJackTsList[key]; !exist {
		hm.InitTsJackInfo(platform, freeID)
	}
	if _, exist := hm.HundredJackSortList[key]; !exist {
		hm.InitSortJackInfo(platform, freeID)
	}
	ISInitJackInfo = true
	return
}

// GetJackTsInfo 返回TsInfo
func (hm *HundredJackListManager) GetJackTsInfo(platform string, freeID int32) []*HundredJackInfo {
	key := fmt.Sprintf("%v-%v", platform, freeID)
	if _, exist := hm.HundredJackTsList[key]; !exist { // 玩家进入scene 已经初始化
		hm.InitTsJackInfo(platform, freeID)
	}
	return hm.HundredJackTsList[key]
}

// GetJackSortInfo 返回SortInfo
func (hm *HundredJackListManager) GetJackSortInfo(platform string, freeID int32) []*HundredJackInfo {
	key := fmt.Sprintf("%v-%v", platform, freeID)
	if _, exist := hm.HundredJackSortList[key]; !exist {
		hm.InitSortJackInfo(platform, freeID)
	}
	return hm.HundredJackSortList[key]
}

// Insert 插入
func (hm *HundredJackListManager) Insert(coin, turncoin int64, snid, roomid, jackType, inGame, vip int32, platform, channel, name string, gamedata []string) {
	key := fmt.Sprintf("%v-%v", platform, roomid)
	log := model.NewHundredjackpotLogEx(snid, coin, turncoin, roomid, jackType, inGame, vip, platform, channel, name, gamedata)
	///////////////////实际不走这里
	if _, exist := hm.HundredJackTsList[key]; !exist {
		hm.InitTsJackInfo(platform, roomid)
	}
	if _, exist := hm.HundredJackSortList[key]; !exist {
		hm.InitSortJackInfo(platform, roomid)
	}
	/////////////////////
	hm.InsertLog(log)
	data := &HundredJackInfo{
		HundredjackpotLog: *log,
	}
	/*logger.Logger.Trace("HundredJackListManager log 1 ", log.SnID, log.LogID, data.GameData)
	for _, v := range hm.HundredJackTsList[key] {
		logger.Logger.Trace("HundredJackListManager log 2 ", v.SnID, v.LogID, v.GameData)
	}*/
	for i, v := range hm.HundredJackSortList[key] { // 插入
		if v.Coin < log.Coin {
			d1 := append([]*HundredJackInfo{}, hm.HundredJackSortList[key][i:]...)
			hm.HundredJackSortList[key] = append(hm.HundredJackSortList[key][:i], data)
			hm.HundredJackSortList[key] = append(hm.HundredJackSortList[key], d1...)
			goto Exit
		}
	}
	if len(hm.HundredJackSortList[key]) < model.HundredjackpotLogMaxLimitPerQuery {
		hm.HundredJackSortList[key] = append(hm.HundredJackSortList[key], data)
	}
Exit:
	d1 := append([]*HundredJackInfo{}, hm.HundredJackTsList[key][0:]...)
	hm.HundredJackTsList[key] = append(hm.HundredJackTsList[key][:0], data)
	hm.HundredJackTsList[key] = append(hm.HundredJackTsList[key], d1...)
	var delList []*HundredJackInfo
	if len(hm.HundredJackTsList[key]) > model.HundredjackpotLogMaxLimitPerQuery {
		delList = append(delList, hm.HundredJackTsList[key][model.HundredjackpotLogMaxLimitPerQuery:]...)
		hm.HundredJackTsList[key] = hm.HundredJackTsList[key][:model.HundredjackpotLogMaxLimitPerQuery]
	}
	if len(hm.HundredJackSortList[key]) > model.HundredjackpotLogMaxLimitPerQuery {
		delList = append(delList, hm.HundredJackSortList[key][model.HundredjackpotLogMaxLimitPerQuery:]...)
		hm.HundredJackSortList[key] = hm.HundredJackSortList[key][:model.HundredjackpotLogMaxLimitPerQuery]
	}
	/*for _, v := range hm.HundredJackTsList[key] {
		logger.Logger.Trace("HundredJackListManager log 3 ", v.SnID, v.LogID, v.GameData)
	}*/
	for _, v := range delList {
		if hm.IsCanDel(v, hm.HundredJackTsList[key], hm.HundredJackSortList[key]) { // 两个排行帮都不包含
			logger.Logger.Info("HundredJackListManager DelOneJackInfo ", v.LogID)
			hm.DelOneJackInfo(v.Platform, v.LogID)
		}
	}
}

// IsCanDel 能否删除
func (hm *HundredJackListManager) IsCanDel(deldata *HundredJackInfo, tsList, sortList []*HundredJackInfo) bool {
	for _, v := range tsList {
		if v.LogID == deldata.LogID {
			return false
		}
	}
	for _, v := range sortList {
		if v.LogID == deldata.LogID {
			return false
		}
	}
	return true
}

// InsertLog insert db
func (hm *HundredJackListManager) InsertLog(log *model.HundredjackpotLog) {
	task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
		err := model.InsertHundredjackpotLog(log)
		if err != nil {
			logger.Logger.Error("HundredJackListManager Insert ", err)
		}
		return err
	}), nil, "InsertHundredJack").Start()
}

// UpdateLikeNum updata likenum
func (hm *HundredJackListManager) UpdateLikeNum(plt string, gid bson.ObjectId, like int32, likesnids string) {
	task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
		err := model.UpdateLikeNum(plt, gid, like, likesnids)
		if err != nil {
			logger.Logger.Error("HundredJackListManager UpdateHundredLikeNum ", err)
		}
		return err
	}), nil, "UpdateHundredLikeNum").Start()
}

// UpdatePlayBlackNum updata playblacknum
func (hm *HundredJackListManager) UpdatePlayBlackNum(plt string, gid bson.ObjectId, playblack int32) []string {
	var ret []string
	task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
		data, err := model.UpdatePlayBlackNum(plt, gid, playblack)
		if err != nil {
			logger.Logger.Error("HundredJackListManager DelOneJackInfo ", err)
			return nil
		}
		return data
	}), task.CompleteNotifyWrapper(func(data interface{}, tt task.Task) {
		if data != nil {
			ret = data.([]string)
			logger.Logger.Warnf("UpdatePlayBlackNum  data:%v", ret)
		}
		return
	}), "UpdatePlayBlackNum").Start()

	logger.Logger.Error("HundredJackListManager UpdatePlayBlackNum ", ret)
	if len(ret) == 0 {
		return ret
	}
	return nil
}

// DelOneJackInfo del
func (hm *HundredJackListManager) DelOneJackInfo(plt string, gid bson.ObjectId) {
	task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
		err := model.RemoveHundredjackpotLogOne(plt, gid)
		if err != nil {
			logger.Logger.Error("HundredJackListManager DelOneJackInfo ", err)
		}
		return err
	}), nil, "DelOneJackInfo").Start()
}

func init() {
	module.RegisteModule(HundredJackListMgr, time.Hour, 0)
}
