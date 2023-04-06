package main

import (
	"games.yol.com/win88/model"
	"games.yol.com/win88/srvdata"
	"github.com/globalsign/mgo/bson"
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/module"
	"github.com/idealeak/goserver/core/task"
	"math/rand"
	"sort"
	"time"
)

var MatchSeasonRankMgrSington = &MatchSeasonRankMgr{
	MatchSeasonRank:          make(map[string][]*MatchSeasonRank),
	MatchSeasonRankDirty:     make(map[string]bool),
	RobotMatchSeasonRankInit: make(map[string]bool),
	RobotMatchSeasonRank:     make(map[string][]*MatchSeasonRank),
}

type MatchSeasonRankMgr struct {
	MatchSeasonRank          map[string][]*MatchSeasonRank //平台
	MatchSeasonRankDirty     map[string]bool
	RobotMatchSeasonRankInit map[string]bool
	RobotMatchSeasonRank     map[string][]*MatchSeasonRank //平台
}

type MatchSeasonRank struct {
	Id       bson.ObjectId `bson:"_id"`
	Platform string
	SnId     int32
	Name     string
	Lv       int32 //段位
	UpdateTs int64
}

func (this *MatchSeasonRankMgr) UpdateMatchSeasonRank(p *Player, lv int32) {
	logger.Logger.Trace("(this *MatchSeasonRankMgr) UpdateMatchSeasonRank: SnId: ", p.SnId, " lv: ", lv)
	platform := p.Platform
	msrs := this.GetMatchSeasonRank(platform)
	if msrs == nil {
		msrs = []*MatchSeasonRank{}
	}
	have := false
	for _, msr := range msrs {
		if msr.SnId == p.SnId {
			msr.Lv = lv
			msr.UpdateTs = time.Now().Unix()
			have = true
			break
		}
	}
	if !have {
		msr := &MatchSeasonRank{
			Id:       bson.NewObjectId(),
			Platform: platform,
			SnId:     p.SnId,
			Name:     p.Name,
			Lv:       lv,
			UpdateTs: time.Now().Unix(),
		}
		msrs = append(msrs, msr)
	}

	sort.Slice(msrs, func(i, j int) bool {
		return msrs[i].Lv > msrs[j].Lv
	})
	if len(msrs) > model.GameParamData.MatchSeasonRankMaxNum {
		if msrs[len(msrs)-1].SnId != p.SnId { //上榜玩家有变化
			this.MatchSeasonRankDirty[platform] = true
		}
		msrs = append(msrs[:model.GameParamData.MatchSeasonRankMaxNum])
	} else {
		this.MatchSeasonRankDirty[platform] = true
	}
	this.MatchSeasonRank[platform] = msrs
}

func (this *MatchSeasonRankMgr) GetMatchSeasonRank(platform string) []*MatchSeasonRank {
	logger.Logger.Trace("(this *MatchSeasonRankMgr) GetMatchSeasonRank: platform = ", platform)
	return this.MatchSeasonRank[platform]
}

func (this *MatchSeasonRankMgr) SetMatchSeasonRank(platform string, mss []*MatchSeasonRank) {
	logger.Logger.Trace("(this *MatchSeasonRankMgr) SetMatchSeasonRank: mss = ", mss)
	this.MatchSeasonRank[platform] = mss
	this.MatchSeasonRankDirty[platform] = true
}

func (this *MatchSeasonRankMgr) MatchSeasonRankInherit(platform string) {
	msr := this.GetMatchSeasonRank(platform)
	logger.Logger.Trace("(this *MatchSeasonRankMgr) MatchSeasonRankInherit: msr = ", msr)
	if msr == nil {
		return
	}
	for _, rank := range msr {
		rank.Lv = MatchSeasonMgrSington.MatchSeasonInherit(rank.Lv)
	}
	this.SetMatchSeasonRank(platform, msr)
}

func (this *MatchSeasonRankMgr) InitMatchSeasonRank(platform string) {
	logger.Logger.Trace("(this *MatchSeasonRankMgr) InitMatchSeasonRank: ", platform)
	if platform == Default_Platform {
		return
	}
	if this.MatchSeasonRank[platform] != nil {
		return
	}
	if this.MatchSeasonRank[platform] == nil {
		logger.Logger.Trace("(this *MatchSeasonRankMgr) InitMatchSeasonRank: ", this.MatchSeasonRank[platform])
		task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
			ret, err := model.QueryMatchSeasonRank(platform)
			logger.Logger.Trace("(this *MatchSeasonRankMgr) 1 QueryMatchSeasonRank: ", ret)
			if err != nil {
				return nil
			}
			return ret
		}), task.CompleteNotifyWrapper(func(data interface{}, tt task.Task) {
			var ret []*model.MatchSeasonRank
			if data == nil || data.([]*model.MatchSeasonRank) == nil { //初始数据去log_matchseason里面取段位前n名
				task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
					retRank, err := model.QueryMatchSeason(platform)
					logger.Logger.Trace("(this *MatchSeasonRankMgr) 1 QueryMatchSeason: ", ret)
					if err != nil {
						return nil
					}
					return retRank
				}), task.CompleteNotifyWrapper(func(data interface{}, tt task.Task) {
					var retRank []*model.MatchSeason
					logger.Logger.Trace("(this *MatchSeasonRankMgr) 2 QueryMatchSeason: ", ret)
					if data == nil || data.([]*model.MatchSeason) == nil {
						ams := MatchSeasonMgrSington.GetAllMatchSeason()
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
									retRank = append(retRank, mms)
								}
							}
						}
					} else {
						retRank = data.([]*model.MatchSeason)
					}
					if retRank != nil {
						this.MatchSeasonRank[platform] = []*MatchSeasonRank{}
						sort.Slice(retRank, func(i, j int) bool {
							return retRank[i].Lv > retRank[j].Lv
						})
						if len(retRank) > model.GameParamData.MatchSeasonRankMaxNum {
							retRank = append(retRank[:model.GameParamData.MatchSeasonRankMaxNum])
						}
						for i := 0; i < len(retRank); i++ {
							season := retRank[i]
							msr := &MatchSeasonRank{
								Id:       season.Id,
								Platform: season.Platform,
								SnId:     season.SnId,
								Name:     season.Name,
								Lv:       season.Lv,
								UpdateTs: season.UpdateTs,
							}
							this.MatchSeasonRank[platform] = append(this.MatchSeasonRank[platform], msr)
							this.MatchSeasonRankDirty[platform] = true
						}
						logger.Logger.Trace("(this *MatchSeasonRankMgr) 3 QueryMatchSeason: ", this.MatchSeasonRank[platform])
					}
				})).StartByFixExecutor("platform:" + platform)
			} else {
				ret = data.([]*model.MatchSeasonRank)
				this.MatchSeasonRank[platform] = []*MatchSeasonRank{}
				for _, rank := range ret {
					msr := &MatchSeasonRank{
						Id:       rank.Id,
						Platform: rank.Platform,
						SnId:     rank.SnId,
						Name:     rank.Name,
						Lv:       rank.Lv,
						UpdateTs: rank.UpdateTs,
					}
					this.MatchSeasonRank[platform] = append(this.MatchSeasonRank[platform], msr)
				}
				logger.Logger.Trace("(this *MatchSeasonRankMgr) 3 QueryMatchSeasonRank: ", this.MatchSeasonRank[platform])
			}
		})).StartByFixExecutor("platform:" + platform)
	}
}

func (this *MatchSeasonRankMgr) SaveMatchSeasonRank(platform string) {
	logger.Logger.Trace("(this *MatchSeasonRankMgr) SaveMatchSeasonRank: ", platform)
	msrp := this.MatchSeasonRank[platform]
	if msrp != nil && this.MatchSeasonRankDirty[platform] {
		this.MatchSeasonRankDirty[platform] = false
		dirtyMsrs := []*model.MatchSeasonRank{}
		for _, rank := range msrp {
			msr := &model.MatchSeasonRank{
				Id:       rank.Id,
				Platform: rank.Platform,
				SnId:     rank.SnId,
				Name:     rank.Name,
				Lv:       rank.Lv,
				UpdateTs: rank.UpdateTs,
			}
			dirtyMsrs = append(dirtyMsrs, msr)
		}
		if dirtyMsrs != nil && len(dirtyMsrs) > 0 {
			task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
				model.DropMatchSeasonRank(platform)
				model.UpsertMatchSeasonRank(platform, dirtyMsrs)
				return nil
			}), task.CompleteNotifyWrapper(func(data interface{}, tt task.Task) {
			})).StartByFixExecutor("platform:" + platform)
		}
	}
}

func (this *MatchSeasonRankMgr) SaveAllMatchSeasonRank() {
	for platform, _ := range PlatformMgrSington.Platforms {
		if platform == Default_Platform {
			continue
		}
		this.SaveMatchSeasonRank(platform)
	}
}

func (this *MatchSeasonRankMgr) CreateRobotLv() int32 {
	Lv := int32(1)
	now := time.Now()
	first := now.Format("2006-01") + "-01"
	start, _ := time.ParseInLocation("2006-01-02", first, time.Local)
	diffUnix := now.Unix() - start.Unix()
	diffDay := diffUnix/int64(24*60*60) + 1
	data := srvdata.PBDB_MatchRankMgr.GetData(int32(diffDay))
	if data != nil && data.RankStar != nil && len(data.RankStar) > 0 {
		diff := data.RankStar[1] - data.RankStar[0]
		min := data.RankStar[0]
		if data.RankStar[0] > data.RankStar[1] {
			diff = data.RankStar[0] - data.RankStar[1]
			min = data.RankStar[1]
		}
		Lv = rand.Int31n(diff) + min
	}
	return Lv
}

func (this *MatchSeasonRankMgr) CreateRobotMatchSeasonRank(platform string) {
	if this.RobotMatchSeasonRankInit[platform] {
		return
	}
	this.RobotMatchSeasonRank = make(map[string][]*MatchSeasonRank)
	for _, player := range PlayerMgrSington.playerSnMap {
		if player != nil && player.IsRob {
			msr := &MatchSeasonRank{
				Platform: platform,
				SnId:     player.SnId,
				Name:     player.Name,
				Lv:       this.CreateRobotLv(),
			}
			this.RobotMatchSeasonRank[platform] = append(this.RobotMatchSeasonRank[platform], msr)
		}
		if len(this.RobotMatchSeasonRank[platform]) >= model.GameParamData.MatchSeasonRankMaxNum {
			this.RobotMatchSeasonRankInit[platform] = true
			break
		}
	}
}

func (this *MatchSeasonRankMgr) GetRobotMatchSeasonRank(platform string) []*MatchSeasonRank {
	if !this.RobotMatchSeasonRankInit[platform] {
		this.CreateRobotMatchSeasonRank(platform)
	}
	if this.RobotMatchSeasonRank == nil || this.RobotMatchSeasonRank[platform] == nil || len(this.RobotMatchSeasonRank[platform]) < model.GameParamData.MatchSeasonRankMaxNum {
		this.CreateRobotMatchSeasonRank(platform)
	}
	return this.RobotMatchSeasonRank[platform]
}

func (this *MatchSeasonRankMgr) UpdateRobotMatchSeasonRank(platform string) {
	rmsr := this.GetRobotMatchSeasonRank(platform)
	logger.Logger.Trace("(this *MatchSeasonRankMgr) 1 UpdateRobotMatchSeasonRank：", rmsr)
	if rmsr != nil {
		for _, rank := range rmsr {
			diff := rand.Int31n(7) - 3
			rank.Lv += diff
		}
		logger.Logger.Trace("(this *MatchSeasonRankMgr) 2 UpdateRobotMatchSeasonRank：", rmsr)
	}
}

func (this *MatchSeasonRankMgr) ModuleName() string {
	return "MatchSeasonRankMgr"
}

func (this *MatchSeasonRankMgr) Init() {
	for platform, _ := range PlatformMgrSington.Platforms {
		if platform == Default_Platform {
			continue
		}
		this.InitMatchSeasonRank(platform)
		this.RobotMatchSeasonRankInit[platform] = false
	}
}

func (this *MatchSeasonRankMgr) Update() {
	this.SaveAllMatchSeasonRank()
}

func (this *MatchSeasonRankMgr) Shutdown() {
	this.SaveAllMatchSeasonRank()
	module.UnregisteModule(this)
}

func (this *MatchSeasonRankMgr) InterestClockEvent() int {
	//TODO implement me
	//panic("implement me")
	return 1<<CLOCK_EVENT_HOUR | 1<<CLOCK_EVENT_DAY
}

func (this *MatchSeasonRankMgr) OnSecTimer() {
	//TODO implement me
	//panic("implement me")
}

func (this *MatchSeasonRankMgr) OnMiniTimer() {
	//TODO implement me
	//panic("implement me")
}

func (this *MatchSeasonRankMgr) OnHourTimer() {
	//TODO implement me
	//panic("implement me")
	logger.Logger.Trace("(this *MatchSeasonRankMgr) OnHourTimer()")
	for platform, _ := range PlatformMgrSington.Platforms {
		if platform == Default_Platform {
			continue
		}
		this.UpdateRobotMatchSeasonRank(platform)
	}
}

func (this *MatchSeasonRankMgr) OnDayTimer() {
	//TODO implement me
	//panic("implement me")
	for platform, _ := range PlatformMgrSington.Platforms {
		if platform == Default_Platform {
			continue
		}
		this.RobotMatchSeasonRankInit[platform] = false
	}
}

func (this *MatchSeasonRankMgr) OnWeekTimer() {
	//TODO implement me
	//panic("implement me")
}

func (this *MatchSeasonRankMgr) OnMonthTimer() {
	//TODO implement me
	//panic("implement me")
}

func (this *MatchSeasonRankMgr) OnShutdown() {
	//TODO implement me
	//panic("implement me")
}

func init() {
	module.RegisteModule(MatchSeasonRankMgrSington, time.Minute*1, 0)
	ClockMgrSington.RegisteSinker(MatchSeasonRankMgrSington)
}
