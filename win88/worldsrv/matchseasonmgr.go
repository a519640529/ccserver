package main

import (
	"games.yol.com/win88/model"
	"games.yol.com/win88/proto"
	"games.yol.com/win88/protocol/tournament"
	"games.yol.com/win88/srvdata"
	"github.com/globalsign/mgo/bson"
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/module"
	"github.com/idealeak/goserver/core/task"
	"sort"
	"strconv"
	"time"
)

var MatchSeasonMgrSington = &MatchSeasonMgr{
	MatchSeasonList: make(map[int32]*MatchSeason),
	MatchSeasonId:   make(map[string]*MatchSeasonId),
}

type MatchSeasonMgr struct {
	MatchSeasonList map[int32]*MatchSeason // snid
	MatchSeasonId   map[string]*MatchSeasonId
}

type MatchSeason struct {
	Id       bson.ObjectId `bson:"_id"`
	Platform string
	SnId     int32
	Name     string
	SeasonId int32 //赛季id
	Lv       int32 //段位
	LastLv   int32 //上赛季段位
	IsAward  bool  //上赛季是否领奖
	AwardTs  int64 //领奖时间
	UpdateTs int64
	dirty    bool
}

type MatchSeasonId struct {
	Id         bson.ObjectId `bson:"_id"`
	Platform   string
	SeasonId   int32 //赛季id
	StartStamp int64 //开始时间戳
	EndStamp   int64 //结束时间戳
	UpdateTs   int64 //更新时间戳
}

func (this *MatchSeasonMgr) exchangeModel2Cache(mms *model.MatchSeason) *MatchSeason {
	if mms == nil {
		return nil
	}
	ms := &MatchSeason{
		Id:       mms.Id,
		Platform: mms.Platform,
		SnId:     mms.SnId,
		Name:     mms.Name,
		Lv:       mms.Lv,
		LastLv:   mms.LastLv,
		IsAward:  mms.IsAward,
		AwardTs:  mms.AwardTs,
		SeasonId: mms.SeasonId,
		UpdateTs: mms.UpdateTs,
	}
	return ms
}

func (this *MatchSeasonMgr) GetMatchSeason(snid int32) *MatchSeason {
	return this.MatchSeasonList[snid]
}

func (this *MatchSeasonMgr) GetAllMatchSeason() map[int32]*MatchSeason {
	return this.MatchSeasonList
}

func (this *MatchSeasonMgr) SetMatchSeason(ms *MatchSeason) {
	if ms == nil {
		return
	}
	this.MatchSeasonList[ms.SnId] = ms
}

func (this *MatchSeasonMgr) DelMatchSeasonCache(snid int32) {
	if this.MatchSeasonList[snid] == nil {
		return
	}
	delete(this.MatchSeasonList, snid)
}

func (this *MatchSeasonMgr) UpdateMatchSeasonLv(p *Player, addlv int32) {
	logger.Logger.Trace("(this *MatchSeasonMgr) UpdateMatchSeasonLv: SnId: ", p.SnId, " addlv: ", addlv)
	if p == nil || p.IsRob {
		return
	}
	platform := p.Platform
	if platform == Default_Platform {
		return
	}
	ms := this.GetMatchSeason(p.SnId)
	if ms != nil {
		ms.Lv = ms.Lv + addlv
		ms.dirty = true
		ms.UpdateTs = time.Now().Unix()
		msid := this.GetMatchSeasonId(platform)
		if msid != nil {
			if addlv != 0 { //段位有变化
				//通知客户端段位更新
				pack := &tournament.SCTMSeasonInfo{
					Id:              msid.SeasonId,
					SeasonTimeStamp: []int64{msid.StartStamp, msid.EndStamp},
					Lv:              ms.Lv,
					LastLv:          ms.LastLv,
					IsAward:         ms.IsAward,
				}
				proto.SetDefaults(pack)
				ok := p.SendToClient(int(tournament.TOURNAMENTID_PACKET_TM_SCTMSeasonInfo), pack)
				logger.Logger.Trace("SCTMSeasonInfoHandler: ok: ", ok, pack)
			}
			//更新排行榜
			logger.Logger.Trace("更新排行榜！！！")
			msrs := MatchSeasonRankMgrSington.GetMatchSeasonRank(platform)
			if msrs == nil { //排行榜没有数据 去缓存中取
				ams := MatchSeasonMgrSington.GetAllMatchSeason()
				mss := []*model.MatchSeason{}
				if ams != nil {
					for _, season := range ams {
						if season.Platform == platform {
							mms := &model.MatchSeason{
								Id:       season.Id,
								Platform: season.Platform,
								SnId:     season.SnId,
								Name:     season.Name,
								SeasonId: season.SeasonId,
								Lv:       season.Lv,
								LastLv:   season.LastLv,
								IsAward:  season.IsAward,
								AwardTs:  season.AwardTs,
							}
							mss = append(mss, mms)
						}
					}
				}
				if mss != nil && len(mss) > 0 {
					cmsrs := []*MatchSeasonRank{}
					sort.Slice(mss, func(i, j int) bool {
						return mss[i].Lv > mss[j].Lv
					})
					if len(mss) > model.GameParamData.MatchSeasonRankMaxNum {
						mss = append(mss[:model.GameParamData.MatchSeasonRankMaxNum])
					}
					for i := 0; i < len(mss); i++ {
						season := mss[i]
						msr := &MatchSeasonRank{
							Id:       season.Id,
							Platform: season.Platform,
							SnId:     season.SnId,
							Name:     season.Name,
							Lv:       season.Lv,
							UpdateTs: season.UpdateTs,
						}
						cmsrs = append(cmsrs, msr)
					}
					MatchSeasonRankMgrSington.SetMatchSeasonRank(platform, cmsrs)
				}
			}
			MatchSeasonRankMgrSington.UpdateMatchSeasonRank(p, ms.Lv)
		}
	}
}

// 段位继承
func (this *MatchSeasonMgr) MatchSeasonInherit(lv int32) int32 {
	logger.Logger.Trace("(this *MatchSeasonMgr) MatchSeasonInherit: lv: ", lv)
	destLv := int32(1)
	for _, v := range srvdata.PBDB_GamMatchLVMgr.Datas.GetArr() {
		if v.Star != nil && len(v.Star) > 1 {
			startStar := v.Star[0]
			endStar := v.Star[1]
			if lv >= startStar && lv <= endStar { //匹配段位
				destLv = v.Star2 //继承后段位
			}
		}
	}
	return destLv
}

func (this *MatchSeasonMgr) UpdateMatchSeasonAward(snid int32) {
	logger.Logger.Trace("(this *MatchSeasonMgr) UpdateMatchSeasonAward ", snid)
	ms := this.GetMatchSeason(snid)
	if ms != nil {
		ms.IsAward = true
		ms.AwardTs = time.Now().Unix()
		ms.UpdateTs = time.Now().Unix()
		ms.dirty = true
	}
}

func (this *MatchSeasonMgr) SaveMatchSeasonData(snid int32, logout bool) {
	logger.Logger.Trace("(this *MatchSeasonMgr) SaveMatchSeasonData ", snid)
	ms := this.MatchSeasonList[snid]
	if ms != nil && ms.dirty {
		ms.dirty = false
		mms := &model.MatchSeason{
			Id:       ms.Id,
			Platform: ms.Platform,
			SnId:     ms.SnId,
			Name:     ms.Name,
			Lv:       ms.Lv,
			LastLv:   ms.LastLv,
			IsAward:  ms.IsAward,
			AwardTs:  ms.AwardTs,
			SeasonId: ms.SeasonId,
			UpdateTs: ms.UpdateTs,
		}
		task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
			return model.UpsertMatchSeason(mms)
		}), task.CompleteNotifyWrapper(func(data interface{}, tt task.Task) {
			logger.Logger.Info("SaveMatchSeasonData!!!")
			if logout {
				this.DelMatchSeasonCache(snid)
			}
		})).StartByFixExecutor("SnId:" + strconv.Itoa(int(snid)))
	}
}

func (this *MatchSeasonMgr) SaveAllMatchSeasonData() {
	for _, msl := range this.MatchSeasonList {
		this.SaveMatchSeasonData(msl.SnId, false)
	}
}

func (this *MatchSeasonMgr) UpdateMatchSeasonId(platform string) {
	logger.Logger.Info("(this *MatchSeasonMgr) UpdateMatchSeasonId")
	if platform == Default_Platform {
		return
	}
	task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
		ret, err := model.QueryMatchSeasonId(platform)
		if err != nil {
			return nil
		}
		return ret
	}), task.CompleteNotifyWrapper(func(data interface{}, tt task.Task) {
		var ret *model.MatchSeasonId
		if data == nil || data.(*model.MatchSeasonId) == nil {
			sstamp, estamp := this.getNowMonthStartAndEnd()
			ret = model.NewMatchSeasonId(platform, int32(1), sstamp, estamp) //初始化赛季
			task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
				return model.UpsertMatchSeasonId(ret)
			}), nil).StartByFixExecutor("UpsertMatchSeasonId")
		} else {
			ret = data.(*model.MatchSeasonId)
		}
		logger.Logger.Info("UpdateMatchSeasonId!!!", ret)
		if ret != nil {
			nowStamp := time.Now().Unix()
			if nowStamp < ret.StartStamp {
				logger.Logger.Error("赛季开始时间错误!!!")
			}
			if nowStamp >= ret.EndStamp { //新赛季
				logger.Logger.Info("新赛季!!!", ret)
				sstamp, estamp := this.getNowMonthStartAndEnd()
				ret.SeasonId++
				ret.StartStamp = sstamp
				ret.EndStamp = estamp
				task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
					return model.UpsertMatchSeasonId(ret)
				}), nil).StartByFixExecutor("UpsertMatchSeasonId")
				//排行榜内的段位继承
				MatchSeasonRankMgrSington.MatchSeasonRankInherit(platform)
				//通知平台玩家继承后的段位数据
				players := PlayerMgrSington.playerOfPlatform[platform]
				for _, p := range players {
					if p != nil && p.IsOnLine() && !p.IsRob {
						ms := MatchSeasonMgrSington.GetMatchSeason(p.SnId)
						if ms != nil {
							if ms.SeasonId < ret.SeasonId { //不同赛季段位继承
								num := ret.SeasonId - ms.SeasonId
								finalLv := ms.Lv
								for i := 0; i < int(num); i++ { //继承几次
									if i == int(num)-1 { //上个赛季
										ms.LastLv = finalLv
									}
									finalLv = MatchSeasonMgrSington.MatchSeasonInherit(finalLv)
								}
								ms.Lv = finalLv
								ms.SeasonId = ret.SeasonId
								ms.IsAward = false
								ms.UpdateTs = time.Now().Unix()
								ms.dirty = true
								MatchSeasonMgrSington.SetMatchSeason(ms) //更新缓存
								pack := &tournament.SCTMSeasonInfo{
									Id:              ret.SeasonId,
									SeasonTimeStamp: []int64{ret.StartStamp, ret.EndStamp},
									Lv:              ms.Lv,
									LastLv:          ms.LastLv,
									IsAward:         ms.IsAward,
								}
								proto.SetDefaults(pack)
								logger.Logger.Trace("SCTMSeasonInfo:", p.SnId, " pack: ", pack)
								p.SendToClient(int(tournament.TOURNAMENTID_PACKET_TM_SCTMSeasonInfo), pack)
							}
						}
					}
				}
			}
			this.MatchSeasonId[platform] = &MatchSeasonId{
				Id:         ret.Id,
				Platform:   ret.Platform,
				SeasonId:   ret.SeasonId,
				StartStamp: ret.StartStamp,
				EndStamp:   ret.EndStamp,
				UpdateTs:   ret.UpdateTs,
			}
		}
	})).StartByFixExecutor("platform: " + platform)
}

func (this *MatchSeasonMgr) GetMatchSeasonId(platform string) *MatchSeasonId {
	logger.Logger.Info("(this *MatchSeasonMgr) GetMatchSeasonId", platform)
	return this.MatchSeasonId[platform]
}

// 获取当月初和月末时间戳
func (this *MatchSeasonMgr) getNowMonthStartAndEnd() (int64, int64) {
	now := time.Now()
	first := now.Format("2006-01") + "-01"
	start, _ := time.ParseInLocation("2006-01-02", first, time.Local)
	last := start.AddDate(0, 1, 0).Format("2006-01-02")
	end, _ := time.ParseInLocation("2006-01-02", last, time.Local)
	return start.Unix(), end.Unix() - 1
}

func (this *MatchSeasonMgr) ModuleName() string {
	return "MatchSeasonMgr"
}

func (this *MatchSeasonMgr) Init() {
	for platform, _ := range PlatformMgrSington.Platforms {
		if platform == Default_Platform {
			continue
		}
		this.UpdateMatchSeasonId(platform)
	}
}

func (this *MatchSeasonMgr) Update() {
	this.SaveAllMatchSeasonData()
}

func (this *MatchSeasonMgr) Shutdown() {
	this.SaveAllMatchSeasonData()
	module.UnregisteModule(this)
}

func (this *MatchSeasonMgr) InterestClockEvent() int {
	//TODO implement me
	//panic("implement me")
	return 1 << CLOCK_EVENT_MONTH
}

func (this *MatchSeasonMgr) OnSecTimer() {
	//TODO implement me
	//panic("implement me")
}

func (this *MatchSeasonMgr) OnMiniTimer() {
	//TODO implement me
	//panic("implement me")
}

func (this *MatchSeasonMgr) OnHourTimer() {
	//TODO implement me
	//panic("implement me")
}

func (this *MatchSeasonMgr) OnDayTimer() {
	//TODO implement me
	//panic("implement me")
}

func (this *MatchSeasonMgr) OnWeekTimer() {
	//TODO implement me
	//panic("implement me")
}

func (this *MatchSeasonMgr) OnMonthTimer() {
	logger.Logger.Info("(this *MatchSeasonMgr) OnMonthTimer")
	for platform, _ := range PlatformMgrSington.Platforms {
		if platform == Default_Platform {
			continue
		}
		this.UpdateMatchSeasonId(platform)
	}
}

func (this *MatchSeasonMgr) OnShutdown() {
	//TODO implement me
	//panic("implement me")
}

func init() {
	module.RegisteModule(MatchSeasonMgrSington, time.Minute*1, 0)
	ClockMgrSington.RegisteSinker(MatchSeasonMgrSington)
}
