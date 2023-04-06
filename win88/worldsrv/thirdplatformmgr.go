package main

import (
	"encoding/json"
	"games.yol.com/win88/model"
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/module"
	"github.com/idealeak/goserver/core/task"
	"strings"
)

var ThirdPlatformMgrSington = &ThirdPlatformMgr{
	Platforms: make(map[string]*PlatformOfThirdPlatform),
}

type PlatformOfThirdPlatform struct {
	*model.PlatformOfThirdPlatform
	dirty bool
}

func (this *PlatformOfThirdPlatform) AddCoin(platform string, coin int64) {
	if this.ThdPlatform == nil {
		this.ThdPlatform = make(map[string]*model.ThirdPlatform)
	}
	if this.ThdPlatform != nil {
		if tpp, exist := this.ThdPlatform[strings.ToLower(platform)]; exist {
			tpp.Coin += coin

			this.dirty = true
		} else {
			this.ThdPlatform[strings.ToLower(platform)] = &model.ThirdPlatform{
				Coin: coin,
			}
			this.dirty = true
		}
	}
}

func (this *PlatformOfThirdPlatform) AddNextCoin(platform string, coin int64) {
	if this.ThdPlatform == nil {
		this.ThdPlatform = make(map[string]*model.ThirdPlatform)
	}
	if this.ThdPlatform != nil {
		if tpp, exist := this.ThdPlatform[strings.ToLower(platform)]; exist {
			tpp.NextCoin += coin

			this.dirty = true
		} else {
			this.ThdPlatform[strings.ToLower(platform)] = &model.ThirdPlatform{
				NextCoin: coin,
			}
			this.dirty = true
		}
	}
}

func (this *PlatformOfThirdPlatform) SetCoin(platform string, coin int64) {
	if this.ThdPlatform == nil {
		this.ThdPlatform = make(map[string]*model.ThirdPlatform)
	}
	if this.ThdPlatform != nil {
		if tpp, exist := this.ThdPlatform[strings.ToLower(platform)]; exist {
			tpp.Coin = coin
			this.dirty = true
		} else {
			this.ThdPlatform[strings.ToLower(platform)] = &model.ThirdPlatform{
				Coin: coin,
			}
			this.dirty = true
		}
	}
}

func (this *PlatformOfThirdPlatform) SetNextCoin(platform string, coin int64) {
	if this.ThdPlatform == nil {
		this.ThdPlatform = make(map[string]*model.ThirdPlatform)
	}
	if this.ThdPlatform != nil {
		if tpp, exist := this.ThdPlatform[strings.ToLower(platform)]; exist {
			tpp.NextCoin = coin
			this.dirty = true
		} else {
			this.ThdPlatform[strings.ToLower(platform)] = &model.ThirdPlatform{
				NextCoin: coin,
			}
			this.dirty = true
		}
	}
}

func (this *PlatformOfThirdPlatform) GetCoin(platform string) int64 {
	if this.ThdPlatform != nil {
		if tpp, exist := this.ThdPlatform[strings.ToLower(platform)]; exist {
			return tpp.Coin
		}
	}
	return 0
}

func (this *PlatformOfThirdPlatform) GetNextCoin(platform string) int64 {
	if this.ThdPlatform != nil {
		if tpp, exist := this.ThdPlatform[strings.ToLower(platform)]; exist {
			return tpp.NextCoin
		}
	}
	return 0
}

func (this *PlatformOfThirdPlatform) Clone() *PlatformOfThirdPlatform {
	var ptp PlatformOfThirdPlatform
	data, err := json.Marshal(this)
	if err == nil {
		err = json.Unmarshal(data, &ptp)
		if err == nil {
			return &ptp
		}
	}
	return nil
}

type ThirdPlatformMgr struct {
	BaseClockSinker
	Platforms map[string]*PlatformOfThirdPlatform
}

func (this *ThirdPlatformMgr) InitData() {
	platformList, err := model.GetAllThirdPlatform()
	if err != nil {
		logger.Logger.Error("InitData count failed:", err)
	}

	for i := 0; i < len(platformList); i++ {
		p := &platformList[i]
		if p != nil {
			this.Platforms[p.Platform] = &PlatformOfThirdPlatform{PlatformOfThirdPlatform: p}
		}
	}
}

func (this *ThirdPlatformMgr) AddPlatform(platform string) *PlatformOfThirdPlatform {
	ptp := &PlatformOfThirdPlatform{
		PlatformOfThirdPlatform: model.NewThirdPlatform(platform),
	}
	this.Platforms[platform] = ptp
	return ptp
}

func (this *ThirdPlatformMgr) InsertPlatform(platform *PlatformOfThirdPlatform) {
	if platform != nil {
		pCopy := platform.Clone()
		if pCopy != nil {
			platform.dirty = false
			task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
				return model.InsertThirdPlatform(pCopy.PlatformOfThirdPlatform)
			}), nil, "UpdateThirdPlatform").StartByFixExecutor("ThirdPlatform")
		}
	}
}

func (this *ThirdPlatformMgr) GetThirdPlatform(platform string) *PlatformOfThirdPlatform {
	if p, exist := this.Platforms[platform]; exist && p != nil {
		return p
	}
	return nil
}

func (this *ThirdPlatformMgr) GetThirdPlatformCoin(platform, thirdPlatform string) int64 {
	p := this.GetThirdPlatform(platform)
	if p != nil {
		return p.GetCoin(thirdPlatform)
	}
	return 0
}

func (this *ThirdPlatformMgr) AddThirdPlatformCoin(platform, thirdPlatform string, coin int64) bool {
	p := this.GetThirdPlatform(platform)
	if p != nil {
		p.AddCoin(thirdPlatform, coin)
		return true
	}
	return false
}

func (this *ThirdPlatformMgr) ModuleName() string {
	return "ThirdPlatformMgr"
}

func (this *ThirdPlatformMgr) Init() {
	this.InitData()
}

func (this *ThirdPlatformMgr) Update() {
	this.SaveAll(false)
}

func (this *ThirdPlatformMgr) Shutdown() {
	this.SaveAll(true)
	module.UnregisteModule(this)
}

// 感兴趣所有clock event
func (this *ThirdPlatformMgr) InterestClockEvent() int {
	return 1 << CLOCK_EVENT_MONTH
}

func (this *ThirdPlatformMgr) OnMonthTimer() {
	for _, p := range this.Platforms {
		if p != nil {
			p.dirty = true
			for _, thr := range p.ThdPlatform {
				if thr != nil {
					if thr.Coin > thr.NextCoin {
						thr.Coin = thr.NextCoin
					} else {
						thr.NextCoin = thr.Coin
					}
				}
			}
		}
	}

	this.SaveAll(false)
}

func (this *ThirdPlatformMgr) SaveAll(bImm bool) {
	for _, p := range this.Platforms {
		if p != nil && p.dirty {
			pCopy := p.Clone()
			if pCopy != nil {
				if bImm {
					err := model.UpdateThirdPlatform(pCopy.PlatformOfThirdPlatform)
					if err != nil {
						logger.Logger.Warnf("UpdateThirdPlatform err:%v", err)
					} else {
						p.dirty = false
					}
				} else {
					p.dirty = false
					task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
						return model.UpdateThirdPlatform(pCopy.PlatformOfThirdPlatform)
					}), nil, "UpdateThirdPlatform").StartByFixExecutor("ThirdPlatform")
				}
			}
		}
	}
}

func init() {
	//module.RegisteModule(ThirdPlatformMgrSington, time.Minute*5, 0)
	//ClockMgrSington.RegisteSinker(ThirdPlatformMgrSington)
}
