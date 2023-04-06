package main

import (
	"time"

	"games.yol.com/win88/model"

	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/module"
	"github.com/idealeak/goserver/core/task"
)

var FishJackpotCoinMgr = &FishJackpotCoin{
	Jackpot: make(map[string]int64),
}

type FishJackpotCoin struct {
	Jackpot map[string]int64 //捕鱼奖池
}

type FishJackListManager struct {
	FishJackList map[string][]*FishJackInfo
}

type FishJackInfo struct {
	Ts          int64
	Name        string
	JackpotCoin int64
	JackpotType int32
}

var FishJackListMgr = &FishJackListManager{
	FishJackList: make(map[string][]*FishJackInfo),
}

func (this *FishJackListManager) ModuleName() string {
	return "FishJackListManager"
}

func (this *FishJackListManager) Init() {
	//data := model.
}

func (this *FishJackListManager) Update() {
}

func (this *FishJackListManager) Shutdown() {
	module.UnregisteModule(this)
}

func (this *FishJackListManager) InitJackInfo(platform string) {

	datas, err := model.GetFishJackpotLogByPlatform(platform) // 必须要先得到 不能走协程
	if err == nil {
		for i, v := range datas {
			if i == model.FishJackpotLogMaxLimitPerQuery {
				break
			}
			data := &FishJackInfo{
				Ts:          v.Ts,
				Name:        v.Name,
				JackpotCoin: v.Coin,
				JackpotType: v.JackType,
			}
			this.FishJackList[platform] = append(this.FishJackList[platform], data)
		}
	}
}

func (this *FishJackListManager) GetJackInfo(platform string) []*FishJackInfo {
	if _, exist := this.FishJackList[platform]; !exist {
		this.InitJackInfo(platform)
	}
	return this.FishJackList[platform]
}

func (this *FishJackListManager) Insert(coin int64, snid, roomid, jackType, inGame int32, platform, channel, name string) {

	log := model.NewFishJackpotLogEx(snid, coin, roomid, jackType, inGame, platform, channel, name)
	this.InsertLog(log)
	if _, exist := this.FishJackList[platform]; !exist /*|| len(datas) == 0 */ {
		this.InitJackInfo(platform)
	}
	data := &FishJackInfo{
		Ts:          log.Ts,
		Name:        log.Name,
		JackpotCoin: log.Coin,
		JackpotType: log.JackType,
	}
	d1 := append([]*FishJackInfo{}, this.FishJackList[platform][0:]...)
	this.FishJackList[platform] = append(this.FishJackList[platform][:0], data)
	this.FishJackList[platform] = append(this.FishJackList[platform], d1...)
	if len(this.FishJackList[platform]) > model.FishJackpotLogMaxLimitPerQuery {
		this.FishJackList[platform] = this.FishJackList[platform][:model.FishJackpotLogMaxLimitPerQuery]
	}
	//logger.Logger.Info("FishJackListManager Insert ", d1, this.FishJackList[platform])
}

func (this *FishJackListManager) InsertLog(log *model.FishJackpotLog) {
	task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
		err := model.InsertSignleFishJackpotLog(log)
		if err != nil {
			logger.Logger.Error("FishJackListManager Insert ", err)
		}
		return err
	}), nil, "InsertFishJack").Start()
}

func init() {
	module.RegisteModule(FishJackListMgr, time.Hour, 0)
}
