package base

import (
	"encoding/json"
	"fmt"
	"games.yol.com/win88/protocol/webapi"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	"games.yol.com/win88/common"
	"games.yol.com/win88/model"
	"games.yol.com/win88/protocol/server"
	"games.yol.com/win88/srvdata"
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/module"
	"github.com/idealeak/goserver/core/task"
)

const (
	CoinPoolStatus_Normal  = iota //库存下限 < 库存值 < 库存上限
	CoinPoolStatus_Low            //库存值 < 库存下限
	CoinPoolStatus_High           //库存上限 < 库存值 < 库存上限+偏移量
	CoinPoolStatus_TooHigh        //库存>库存上限+偏移量
)

var CoinPoolMgr = &CoinPoolManager{
	CoinPool:        new(sync.Map),
	ProfitPool:      new(sync.Map),
	CoinPoolSetting: make(map[string]*webapi.CoinPoolSetting),
	curRunTime:      make(map[string]int32),
	lastTime:        time.Now(),
	PoolCache:       make(map[string]int64),
}

func GetCoinPoolMgr() *CoinPoolManager {
	return CoinPoolMgr
}

const COINPOOLKEYTOKEN = "-"
const COINPOOLKEYTOKEN_GROUP = "+"
const MAXCOINPOOLLOOPCOUNT = 300

type CoinPoolListener interface {
	OnSettingUpdate(oldSetting, newSetting *webapi.CoinPoolSetting)
}

type CoinPoolManager struct {
	CoinPool        *sync.Map //调控池
	ProfitPool      *sync.Map //收益池
	CoinPoolDBKey   string
	ProfitPoolDBKey string
	dirty           bool
	CoinPoolSetting map[string]*webapi.CoinPoolSetting
	lastTime        time.Time
	LastDayDtCount  []int
	tax             float64
	loopCount       int
	curRunTime      map[string]int32
	PoolCache       map[string]int64 //杀率 水池缓存值
	listener        []CoinPoolListener
}

func (this *CoinPoolManager) RegisteListener(l CoinPoolListener) {
	this.listener = append(this.listener, l)
}

func (this *CoinPoolManager) GenKey(gameFreeId int32, platform string, groupId int32) string {
	var key string
	if groupId != 0 {
		key = fmt.Sprintf("%v%v%v", gameFreeId, COINPOOLKEYTOKEN_GROUP, groupId)
	} else {
		key = fmt.Sprintf("%v%v%v", gameFreeId, COINPOOLKEYTOKEN, platform)
	}
	return key
}

func (this *CoinPoolManager) SplitKey(key string) (gameFreeId, groupId int32, platform string, ok bool) {
	if strings.Contains(key, COINPOOLKEYTOKEN) {
		datas := strings.Split(key, COINPOOLKEYTOKEN)
		if len(datas) >= 2 {
			id, err := strconv.Atoi(datas[0])
			if err == nil {
				gameFreeId = int32(id)
			}
			platform = datas[1]
			ok = true
			return
		}
	} else {
		datas := strings.Split(key, COINPOOLKEYTOKEN_GROUP)
		if len(datas) >= 2 {
			id, err := strconv.Atoi(datas[0])
			if err == nil {
				gameFreeId = int32(id)
			}
			id, err = strconv.Atoi(datas[1])
			if err == nil {
				groupId = int32(id)
			}
			ok = true
			return
		}
	}
	return
}

func (this *CoinPoolManager) AddProfitPool(key string, coin int64) bool {
	poolValue, ok := this.ProfitPool.Load(key)
	if !ok {
		this.ProfitPool.Store(key, coin)
		this.dirty = true
		return true
	}
	poolCoin, ok := poolValue.(int64)
	if !ok {
		return false
	}

	this.ProfitPool.Store(key, poolCoin+coin)
	this.dirty = true
	return true
}

func (this *CoinPoolManager) LoadCoin(gameFreeId int32, platform string, groupId int32) int64 {
	key := this.GenKey(gameFreeId, platform, groupId)
	return this._loadCoin(key)
}

func (this *CoinPoolManager) _loadCoin(key string) int64 {
	poolValue, ok := this.CoinPool.Load(key)
	if !ok {
		return 0
	}
	poolCoin, ok := poolValue.(int64)
	if !ok {
		logger.Logger.Errorf("Error type convert in coinpoolmanager LoadCoin:%v-%v", key, poolValue)
		return 0
	}
	return poolCoin
}

func (this *CoinPoolManager) GetProfitPoolCoin(gameFreeId int32, platform string, groupId int32) int64 {
	key := this.GenKey(gameFreeId, platform, groupId)
	poolValue, ok := this.ProfitPool.Load(key)
	if !ok {
		return 0
	}
	poolCoin, ok := poolValue.(int64)
	if !ok {
		logger.Logger.Errorf("Error type convert in coinpoolmanager GetProfitPoolCoin:%v-%v", key, poolValue)
		return 0
	}
	return poolCoin
}

//获得当前的返奖比例
func (this *CoinPoolManager) GetRTP(platform string, gameFreeId int32, groupId int32) float64 {
	ret := float64(0.60)
	curCoin := this.LoadCoin(gameFreeId, platform, groupId)
	setting := this.GetCoinPoolSetting(platform, gameFreeId, groupId)
	if setting == nil {
		return ret
	}

	if curCoin <= int64(setting.GetLowerLimit()) {
		return ret
	}
	if setting.GetLowerLimit() <= 0 {
		logger.Logger.Error("game lowlimit ==0  ", gameFreeId, platform, groupId)
		return ret
	}

	maxRTP := model.GameParamData.MaxRTP

	rate := (curCoin - int64(setting.GetLowerLimit())) * 10000 / int64(setting.GetLowerLimit())
	ret = 1.0/(1.0+math.Pow(2, float64(-rate)/10000)) + model.GameParamData.AddRTP
	if ret > maxRTP {
		ret = maxRTP
	}
	return ret
}

//水池额度>水池下限 触发其他玩家调控
func (this *CoinPoolManager) IsHaveEnough(gameFreeId int32, platform string, groupId int32) bool {
	curCoin := this.LoadCoin(gameFreeId, platform, groupId)
	setting := this.GetCoinPoolSetting(platform, gameFreeId, groupId)
	if setting == nil {
		return false
	}
	if curCoin > int64(setting.GetLowerLimit()) {
		return true
	}
	return false
}

//判定是否满足最大出分,out 为负值是出分，正值为收分
func (this *CoinPoolManager) IsMaxOutHaveEnough(platform string, gameFreeId int32, groupId int32, out int64) bool {
	curCoin := this.LoadCoin(gameFreeId, platform, groupId)
	setting := this.GetCoinPoolSetting(platform, gameFreeId, groupId)
	if setting == nil {
		logger.Logger.Info("game no find setting", platform, gameFreeId, groupId)
		return false
	}
	if out >= 0 {
		return true
	}

	if setting.GetMaxOutValue() == 0 {
		return true
	}
	if int64(math.Abs(float64(out))) < (int64(setting.GetMaxOutValue()*int32(model.NormalParamData.PoolMaxOutRate)/100) +
		(int64(curCoin)-int64(setting.GetLowerLimit()))*int64(common.RandFromRange(model.GameParamData.CoinPoolMinOutRate,
			model.GameParamData.CoinPoolMaxOutRate))/100) {
		return true
	}
	logger.Logger.Info("game maxout no enough ", int64(math.Abs(float64(out))), platform, gameFreeId, groupId)
	return false
}

func (this *CoinPoolManager) IsCoinEnough(gameFreeId, groupId int32, platform string, coin int64) bool {
	key := this.GenKey(gameFreeId, platform, groupId)
	if data := srvdata.PBDB_GameFreeMgr.GetData(gameFreeId); data != nil {
		if !model.IsUseCoinPoolControlGame(strconv.Itoa(int(data.GetGameId())), gameFreeId) {
			return true
		}
	}
	poolValue, ok := this.CoinPool.Load(key)
	if !ok {
		return false
	}
	poolCoin, ok := poolValue.(int64)
	if !ok {
		logger.Logger.Errorf("Error type convert in coinpoolmanager coincheck:%v-%v", key, poolValue)
		return false
	}
	if poolCoin < coin {
		return false
	}
	return true
}
func (this *CoinPoolManager) PopCoin(gameFreeId, groupId int32, platform string, coin int64) bool {
	if coin == 0 {
		return true
	}

	key := this.GenKey(gameFreeId, platform, groupId)
	poolVal := coin
	if coin < 0 { //处理收益池
		setting := this._getCoinPoolSetting(key)
		if setting != nil {
			rate := this.GetProfitRate(setting)
			if rate != 0 { //负值也生效???
				profit := coin * int64(rate) / 100
				this.AddProfitPool(key, -profit)
				poolVal = coin - profit
			}
		}
	}
	ok, curPoolVal := this._popCoin(key, poolVal)
	if ok {
		dbGameFree := srvdata.PBDB_GameFreeMgr.GetData(gameFreeId)
		if dbGameFree != nil && dbGameFree.GetSceneType() != -1 { //非试玩场

			//保存数据
			this.ReportCoinPoolRecEvent(gameFreeId, groupId, platform, -coin, curPoolVal, false)
			//setting := this.GetCoinPoolSetting(platform, gameFreeId, groupId)
			//if setting != nil {
			//	poolValue, ok := this.CoinPool.Load(key)
			//	if ok {
			//		poolCoin, ok := poolValue.(int64)
			//		if ok {
			//			if int64(setting.GetLowerLimit()) > poolCoin {
			//				WarningCoinPool(Warning_CoinPoolLow, gameFreeId)
			//			}
			//
			//			if poolCoin < 0 {
			//				WarningCoinPool(Warning_CoinPoolZero, gameFreeId)
			//			}
			//		}
			//	}
			//}
		}
		return true
	}
	return false
}

func (this *CoinPoolManager) _popCoin(key string, coin int64) (bool, int64) {
	poolValue, ok := this.CoinPool.Load(key)
	if !ok {
		return false, 0
	}
	poolCoin, ok := poolValue.(int64)
	if !ok {
		logger.Logger.Errorf("Error type convert in coinpoolmanager popcoin:%v-%v", key, poolValue)
		return false, 0
	}

	//if poolCoin < coin {
	//	return false
	//}

	curValue := poolCoin - coin
	if curValue < -99999999 {
		curValue = -99999999
	}

	this.CoinPool.Store(key, curValue)
	this.dirty = true
	logger.Logger.Infof("$$$$$$$$金币池 %v 取出 %v 金币，现有金币 %v.$$$$$$$$", key, coin, curValue)
	return true, curValue
}

func (this *CoinPoolManager) GetProfitRate(setting *webapi.CoinPoolSetting) int32 {
	if setting == nil {
		return 0
	}

	rate := setting.GetProfitRate()
	if rate != 0 {
		return rate
	}

	rate = setting.GetProfitManualRate()
	if setting.GetProfitUseManual() {
		return rate
	}

	rate = setting.GetProfitAutoRate()
	if rate != 0 {
		return rate
	}

	return 0
}
func (this *CoinPoolManager) PushCoin(gameFreeId, groupId int32, platform string, coin int64) {
	if coin == 0 {
		return
	}
	key := this.GenKey(gameFreeId, platform, groupId)
	poolVal := coin
	if coin > 0 { //处理收益池
		setting := this._getCoinPoolSetting(key)
		if setting != nil {
			rate := this.GetProfitRate(setting)
			if rate != 0 { //负值也生效???
				profit := coin * int64(rate) / 100
				this.AddProfitPool(key, profit)
				poolVal = coin - profit
			}
		}
	}

	curPoolVal := this._pushCoin(key, poolVal)
	dbGameFree := srvdata.PBDB_GameFreeMgr.GetData(gameFreeId)
	if dbGameFree != nil && dbGameFree.GetSceneType() != -1 { //非试玩场
		//保存数据
		this.ReportCoinPoolRecEvent(gameFreeId, groupId, platform, coin, curPoolVal, false)
		//收益池
	}
	return
}

func (this *CoinPoolManager) _pushCoin(key string, coin int64) int64 {
	poolValue, ok := this.CoinPool.Load(key)
	if !ok {
		this.CoinPool.Store(key, coin)
		logger.Logger.Infof("$$$$$$$$金币池 %v 放入 %v 金币，现有金币 %v.$$$$$$$$", key, coin, coin)
		return 0
	}
	poolCoin, ok := poolValue.(int64)
	if !ok {
		logger.Logger.Errorf("Coin pool push error type value:%v-%v", key, poolValue)
		return 0
	}
	poolCoin += coin
	this.CoinPool.Store(key, poolCoin)
	this.dirty = true
	logger.Logger.Infof("$$$$$$$$金币池 %v 放入 %v 金币，现有金币 %v.$$$$$$$$", key, coin, poolCoin)
	return poolCoin
}

func (this *CoinPoolManager) GetCoinPoolCtx(platform string, gamefreeid, groupId int32) model.CoinPoolCtx {
	setting := this.GetCoinPoolSetting(platform, gamefreeid, groupId)
	if setting == nil {
		return model.CoinPoolCtx{}
	}
	curCoin := this.LoadCoin(gamefreeid, platform, groupId)
	state := CoinPoolStatus_Normal
	switch {
	case curCoin < int64(setting.GetLowerLimit()):
		state = CoinPoolStatus_Low
	case curCoin > int64(setting.GetUpperLimit()) && curCoin < int64(setting.GetUpperLimit()+setting.GetUpperOffsetLimit()):
		state = CoinPoolStatus_High
	case curCoin > int64(setting.GetUpperLimit()+setting.GetUpperOffsetLimit()):
		state = CoinPoolStatus_TooHigh
	}
	ctx := model.CoinPoolCtx{CurrValue: int32(curCoin), CurrMode: int32(state)}
	ctx.LowerLimit = setting.GetLowerLimit()
	ctx.UpperLimit = setting.GetUpperLimit()
	return ctx
}

func (this *CoinPoolManager) GetCoinPoolStatus(platform string, gamefreeid, groupId int32) (int, int) {
	return this._getCoinPoolStatus(platform, gamefreeid, groupId, 0)
}
func (this *CoinPoolManager) GetCoinPoolStatus2(platform string, gamefreeid, groupId int32, value int64) (int, int) {
	return this._getCoinPoolStatus(platform, gamefreeid, groupId, value)
}
func (this *CoinPoolManager) _getCoinPoolStatus(platform string, gamefreeid, groupId int32, value int64) (int, int) {
	curCoin := this.LoadCoin(gamefreeid, platform, groupId)
	if value < 0 {
		curCoin += value
	}
	setting := this.GetCoinPoolSetting(platform, gamefreeid, groupId)
	if setting == nil {
		return CoinPoolStatus_Normal, 10000
	}
	switch {

	case curCoin < int64(setting.GetLowerLimit()):
		lowValue := int64(setting.GetUpperLimit()-setting.GetLowerLimit())/10 + int64(setting.GetLowerLimit()) - curCoin
		return CoinPoolStatus_Low, int((float64(lowValue) / float64(setting.GetLowerLimit()+1)) * 10000)
	case curCoin > int64(setting.GetUpperLimit()) && curCoin < int64(setting.GetUpperLimit()+setting.GetUpperOffsetLimit()):
		highValue := curCoin - int64(setting.GetUpperLimit())
		return CoinPoolStatus_High, int((float64(highValue) / float64(setting.GetUpperLimit()+1)) * 10000) //+1避免除0
	case curCoin > int64(setting.GetUpperLimit()+setting.GetUpperOffsetLimit()):
		highValue := curCoin - int64(setting.GetUpperLimit()+setting.GetUpperOffsetLimit())
		highValue = highValue * 2
		return CoinPoolStatus_TooHigh, int((float64(highValue) / float64(setting.GetUpperLimit()+setting.GetUpperOffsetLimit()+1)) * 10000) //+1避免除0
	default:
		return CoinPoolStatus_Normal, 10000
	}
}

func (this *CoinPoolManager) GetFishCoinPoolStatus(platform string, gamefreeid, groupId int32) int {
	curCoin := this.LoadCoin(gamefreeid, platform, groupId)
	setting := this.GetCoinPoolSetting(platform, gamefreeid, groupId)
	if setting == nil {
		return CoinPoolStatus_Normal
	}
	num := SceneMgrSington.GetPlayerNumByGameFree(platform, gamefreeid, groupId)
	switch {
	case curCoin < int64(setting.GetLowerLimit()):
		return CoinPoolStatus_Low
	case curCoin > int64(setting.GetUpperLimit()) && curCoin < int64(setting.GetUpperLimit()+setting.GetUpperOffsetLimit()):
		if num >= setting.GetMinOutPlayerNum() {
			return CoinPoolStatus_High
		} else {
			return CoinPoolStatus_Normal
		}
	case curCoin > int64(setting.GetUpperLimit()+setting.GetUpperOffsetLimit()):
		if num >= setting.GetMinOutPlayerNum() {
			return CoinPoolStatus_TooHigh
		} else {
			return CoinPoolStatus_Normal
		}
	default:
		return CoinPoolStatus_Normal
	}
}

func (this *CoinPoolManager) UpdateCoinPoolSetting(setting *webapi.CoinPoolSetting) {
	if setting != nil {
		key := this.GenKey(setting.GetGameFreeId(), setting.GetPlatform(), setting.GetGroupId())
		//每次更新，都重置这个数值
		if setting.GetResetTime() != 0 {
			runTime := int32(time.Now().Unix()) + setting.GetResetTime()
			this.curRunTime[key] = runTime
		}
		old, _ := this.CoinPoolSetting[key]
		this.CoinPoolSetting[key] = setting
		for _, v := range this.listener {
			v.OnSettingUpdate(old, setting)
		}
		initValue := setting.GetInitValue()
		if initValue != 0 { //初始化水池
			_, ok := this.CoinPool.Load(key)
			if !ok {
				this._pushCoin(key, int64(initValue))
			}
		}
	}
}

func (this *CoinPoolManager) CompareCoinPoolSetting(old, new *webapi.CoinPoolSetting) bool {
	GameFree := srvdata.PBDB_GameFreeMgr.GetData(old.GetGameFreeId())
	if GameFree != nil {
		if old.GetCtroRate() != new.GetCtroRate() {
			return true
		}
	}
	return false
}

func (this *CoinPoolManager) ReportCoinPoolRecEvent(gameFreeId, groupId int32, platform string, coin int64,
	curCoin int64, ignore bool) {
	if !model.GameParamData.OpenPoolRec {
		return
	}
	key := this.GenKey(gameFreeId, platform, groupId)
	setting := this._getCoinPoolSetting(key)
	if setting != nil {
		//d, e := model.MarshalGameCoinPoolEvent(2, platform, gameFreeId, groupId, coin, curCoin,
		//	int64(setting.GetUpperLimit()), int64(setting.GetLowerLimit()))
		//if e == nil {
		//	rmd := model.NewInfluxDBData("hj.coinpool_record", d)
		//	if rmd != nil {
		//		InfluxDBDataChannelSington.Write(rmd)
		//	}
		//}
	}
}

func (this *CoinPoolManager) ResetCoinPool(wgRcp *server.WGResetCoinPool) {
	if wgRcp != nil {
		key := this.GenKey(wgRcp.GetGameFreeId(), wgRcp.GetPlatform(), wgRcp.GetGroupId())
		if setting, exist := this.CoinPoolSetting[key]; exist {
			switch wgRcp.GetPoolType() {
			case 1: //水池
				value := int64(wgRcp.GetValue())
				if value == -1 {
					initValue := setting.GetInitValue()
					if initValue != 0 { //初始化水池
						value = int64(initValue)
					}
				}
				this.CoinPool.Store(key, value)
				logger.Logger.Infof("$$$$$$$$金币池 %v 重置金币 %v.$$$$$$$$", key, value)
			case 2: //营收池
				value := int64(wgRcp.GetValue())
				if value == -1 {
					value = 0
				}
				this.ProfitPool.Store(key, value)
				logger.Logger.Infof("$$$$$$$$营收池 %v 重置金币 %v.$$$$$$$$", key, value)
			case 3: //水池&营收池
				value := int64(wgRcp.GetValue())
				if value == -1 {
					initValue := setting.GetInitValue()
					if initValue != 0 { //初始化水池
						value = int64(initValue)
					}
				}
				this.CoinPool.Store(key, value)
				logger.Logger.Infof("$$$$$$$$金币池 %v 重置金币 %v.$$$$$$$$", key, value)

				value = int64(wgRcp.GetValue())
				if value == -1 {
					value = 0
				}
				this.ProfitPool.Store(key, value)
				logger.Logger.Infof("$$$$$$$$营收池 %v 重置金币 %v.$$$$$$$$", key, value)
			}

		}
	}
}

func (this *CoinPoolManager) EffectCoinPool(chgs *server.WGProfitControlCorrect) {
	if chgs != nil {
		for _, plt := range chgs.Cfg {
			gcfgs := plt.GetGameCfg()
			for _, cfg := range gcfgs {
				key := this.GenKey(cfg.GetGameFreeId(), plt.GetPlatform(), 0)
				if setting, exist := this.CoinPoolSetting[key]; exist {
					setting.ProfitManualRate = cfg.ManualCorrectRate
					setting.ProfitAutoRate = cfg.AutoCorrectRate
					setting.ProfitUseManual = cfg.UseManual
					if cfg.GetDownPool() {
						curCoin := this._loadCoin(key)
						if curCoin > int64(setting.GetLowerLimit()) {
							dif := curCoin - int64(setting.GetLowerLimit())
							this.PoolCache[key] += dif
							this._popCoin(key, dif)
							logger.Logger.Infof("$$$$$$$$降低水池 %v 降低金额 %v.$$$$$$$$", key, dif)
						}
					} else {
						if cache, ok := this.PoolCache[key]; ok {
							this._pushCoin(key, cache)
							logger.Logger.Infof("$$$$$$$$回正水池 %v 回正金额 %v.$$$$$$$$", key, cache)
							delete(this.PoolCache, key)
						}
					}
				}
			}
		}
	}
}

func (this *CoinPoolManager) GetCoinPoolSetting(platform string, gamefreeid, groupId int32) *webapi.CoinPoolSetting {
	key := this.GenKey(gamefreeid, platform, groupId)
	if setting, exist := this.CoinPoolSetting[key]; exist {
		return setting
	}
	return nil
}

func (this *CoinPoolManager) _getCoinPoolSetting(key string) *webapi.CoinPoolSetting {
	if setting, exist := this.CoinPoolSetting[key]; exist {
		return setting
	}
	return nil
}

func (this *CoinPoolManager) GetCoinPoolStatesByPlatform(platform string, games []*common.GamesIndex) (info map[int32]*common.CoinPoolStatesInfo) {
	info = make(map[int32]*common.CoinPoolStatesInfo)
	for _, g := range games {
		gameId := srvdata.PBDB_GameFreeMgr.GetData(g.GameFreeId).GetGameId()
		key := this.GenKey(g.GameFreeId, platform, g.GroupId)
		pool := this.CoinPoolSetting[key]
		if pool != nil {
			//在当前水池找到了
			state, _ := this.GetCoinPoolStatus(platform, g.GameFreeId, pool.GetGroupId())
			nowsate := int32(1)
			switch state {
			case CoinPoolStatus_Normal:
				nowsate = 1
			case CoinPoolStatus_Low:
				nowsate = 2
			case CoinPoolStatus_High:
				nowsate = 3
			case CoinPoolStatus_TooHigh:
				nowsate = 4
			}
			//当前水池的值
			nowpool := &common.CoinPoolStatesInfo{}
			nowpool.GameId = gameId
			nowpool.GameFreeId = g.GameFreeId
			nowpool.LowerLimit = pool.GetLowerLimit()
			nowpool.UpperLimit = pool.GetUpperLimit()
			nowpool.CoinValue = int32(this.LoadCoin(g.GameFreeId, platform, g.GroupId))
			nowpool.States = nowsate
			info[g.GameFreeId] = nowpool
		} else {
			//在当前水池没有找到并且已经开启的游戏 给默认值
			def := srvdata.PBDB_GameCoinPoolMgr.GetData(g.GameFreeId)
			nowpool := &common.CoinPoolStatesInfo{}
			nowpool.GameId = gameId
			nowpool.GameFreeId = g.GameFreeId
			nowpool.LowerLimit = def.GetLowerLimit()
			nowpool.UpperLimit = def.GetUpperLimit()
			nowpool.CoinValue = def.GetInitValue()
			nowpool.States = 1
			info[g.GameFreeId] = nowpool
		}
	}
	return
}
func (this *CoinPoolManager) GetCoinPoolSettingByGame(platform string, gameId, gameMode, groupId int32) (settings []*webapi.CoinPoolSetting) {
	ids, _ := srvdata.DataMgr.GetGameFreeIds(gameId, gameMode)
	nums := SceneMgrSington.GetPlayerNumByGame(platform, gameId, gameMode, groupId)
	for _, id := range ids {
		setting := this.GetCoinPoolSetting(platform, id, groupId)
		if setting != nil {
			s := &webapi.CoinPoolSetting{
				Platform:         setting.GetPlatform(),
				GameFreeId:       setting.GetGameFreeId(),
				ServerId:         setting.GetServerId(),
				GroupId:          setting.GetGroupId(),
				InitValue:        setting.GetInitValue(),
				LowerLimit:       setting.GetLowerLimit(),
				UpperLimit:       setting.GetUpperLimit(),
				UpperOffsetLimit: setting.GetUpperOffsetLimit(),
				MaxOutValue:      setting.GetMaxOutValue(),
				ChangeRate:       setting.GetChangeRate(),
				MinOutPlayerNum:  setting.GetMinOutPlayerNum(),
				CoinValue:        this.LoadCoin(id, platform, groupId),
				PlayerNum:        nums[id],
				BaseRate:         setting.GetBaseRate(),
				CtroRate:         setting.GetCtroRate(),
				HardTimeMin:      setting.GetHardTimeMin(),
				HardTimeMax:      setting.GetHardTimeMax(),
				NormalTimeMin:    setting.GetNormalTimeMin(),
				NormalTimeMax:    setting.GetNormalTimeMax(),
				EasyTimeMin:      setting.GetEasyTimeMin(),
				EasyTimeMax:      setting.GetEasyTimeMax(),
				EasrierTimeMin:   setting.GetEasrierTimeMin(),
				EasrierTimeMax:   setting.GetEasrierTimeMax(),
				CpCangeType:      setting.GetCpCangeType(),
				CpChangeInterval: setting.GetCpChangeInterval(),
				CpChangeTotle:    setting.GetCpChangeTotle(),
				CpChangeLower:    setting.GetCpChangeLower(),
				CpChangeUpper:    setting.GetCpChangeUpper(),
				CoinPoolMode:     setting.GetCoinPoolMode(),
				ProfitRate:       setting.GetProfitRate(),
				ProfitPool:       this.GetProfitPoolCoin(id, platform, groupId),
				ResetTime:        setting.GetResetTime(),
				ProfitAutoRate:   setting.GetProfitAutoRate(),
				ProfitManualRate: setting.GetProfitManualRate(),
				ProfitUseManual:  setting.GetProfitUseManual(),
			}
			settings = append(settings, s)
		}
	}
	return
}

func (this *CoinPoolManager) GetCoinPoolAIModel(platform string, gamefreeid, groupId int32, value int64) int32 {
	poolsetting := this.GetCoinPoolSetting(platform, gamefreeid, groupId)

	poolCoin := this.LoadCoin(gamefreeid, platform, groupId)
	if value < 0 {
		poolCoin += value
	}

	//num := SceneMgrSington.GetPlayerNumByGameFree(platform, gamefreeid, groupId)
	coinPoolModel := common.CoinPoolAIModel_Normal
	if poolCoin >= int64(poolsetting.GetLowerLimit()) && poolCoin <= int64(poolsetting.GetUpperLimit()) {
		coinPoolModel = common.CoinPoolAIModel_Normal
	} else if poolCoin < int64(poolsetting.GetLowerLimit()) {
		coinPoolModel = common.CoinPoolAIModel_ShouFen
	} else if poolCoin >= int64(poolsetting.GetUpperLimit()) && poolCoin < int64(poolsetting.GetUpperLimit()+poolsetting.GetUpperOffsetLimit()) {
		coinPoolModel = common.CoinPoolAIModel_ZheZhong
	} else if poolCoin >= int64(poolsetting.GetUpperLimit()+poolsetting.GetUpperOffsetLimit()) {
		coinPoolModel = common.CoinPoolAIModel_TuFen
	}
	return coinPoolModel

}

//新的计算模式，根据当前水位中线，上下浮动，来决定当前模式,仅用于对战类测试
func (this *CoinPoolManager) GetCoinPoolNewAIModel(platform string, gamefreeid, groupId int32, value int64) int32 {
	poolsetting := this.GetCoinPoolSetting(platform, gamefreeid, groupId)
	poolCoin := this.LoadCoin(gamefreeid, platform, groupId)
	poolCoin += value

	middleValue := int64((poolsetting.GetLowerLimit() + poolsetting.GetUpperLimit()) / 2)

	coinPoolModel := common.CoinPoolAIModel_Normal
	//先按照线性比例计算

	addValue := int64(0)
	if poolCoin > middleValue {
		addValue = (poolCoin - middleValue) * 100 / (int64(poolsetting.GetUpperLimit()) - middleValue)
		addValue = int64(common.GetSoftMaxNum(float64(addValue), 100))
		if rand.Int63n(100) < addValue {
			coinPoolModel = common.CoinPoolAIModel_TuFen
		}
	} else {
		addValue = (middleValue - poolCoin) * 100 / (middleValue - int64(poolsetting.GetLowerLimit()))
		addValue = int64(common.GetSoftMaxNum(float64(addValue), 100))

		if rand.Int63n(100) < addValue {
			coinPoolModel = common.CoinPoolAIModel_ShouFen
		}
	}

	/*
		//正态分布测试
		if poolCoin > middleValue {
			if common.RandNormInt64(poolCoin - middleValue,int64(poolsetting.GetUpperLimit()) - middleValue) {
				coinPoolModel = common.CoinPoolAIModel_TuFen
			}
		} else {
			if common.RandNormInt64(middleValue - poolCoin,middleValue - int64(poolsetting.GetLowerLimit())) {
				coinPoolModel = common.CoinPoolAIModel_ShouFen
			}
		}
	*/
	/*
		num := SceneMgrSington.GetPlayerNumByGameFree(platform, gamefreeid, groupId)
		coinPoolModel := common.CoinPoolAIModel_Normal
		if poolCoin >= int64(poolsetting.GetLowerLimit()) && poolCoin <= int64(poolsetting.GetUpperLimit()) {
			coinPoolModel = common.CoinPoolAIModel_Normal
		} else if poolCoin < int64(poolsetting.GetLowerLimit()) {
			coinPoolModel = common.CoinPoolAIModel_ShouFen
		} else if poolCoin >= int64(poolsetting.GetUpperLimit()) && poolCoin < int64(poolsetting.GetUpperLimit()+poolsetting.GetUpperOffsetLimit()) {
			coinPoolModel = common.CoinPoolAIModel_ZheZhong
		} else if poolCoin >= int64(poolsetting.GetUpperLimit()+poolsetting.GetUpperOffsetLimit()) && poolsetting.GetMinOutPlayerNum() <= num {
			coinPoolModel = common.CoinPoolAIModel_TuFen
		}
	*/
	return coinPoolModel
}

//N线拉霸开奖几率检查（目前有老虎机，世界杯）
func (this *CoinPoolManager) LineRateCheck(playerEx *Player, allscore int32, gameFreeId int32, platform, gameDiff string, realWinRate, groupId int32) bool {
	//个人实时赔率=（个人总产出+1）/（个人总投入+1）
	winCoin, failCoin := playerEx.GetStaticsData(gameDiff)
	playerRate := int32((float64(winCoin+10000) / float64(failCoin+10000)) * 10000)
	curCoinPoolStatus, _ := CoinPoolMgr._getCoinPoolStatus(platform, gameFreeId, groupId, -int64(allscore))
	curValue := CoinPoolMgr.LoadCoin(gameFreeId, platform, groupId)
	setting := CoinPoolMgr.GetCoinPoolSetting(platform, gameFreeId, groupId)
	if setting == nil {
		return false
	}
	if realWinRate != -1 {
		playerRate = realWinRate
	}
	var trueBaseRate int32 //调节赔率
	changeRate := setting.GetCtroRate()
	switch curCoinPoolStatus {
	case CoinPoolStatus_Low: //库存值 < 库存下限
		if playerRate < changeRate {
			//实际计算赔率 = 基础赔率 -（1 - 库存/库存下线）*10000*Rand（50%-100%）+
			// Rand（10%-50%）*（调节赔率-实时个人赔率）+白名单等级*1000-黑名单等级*1000
			baseRate := setting.GetBaseRate()
			baseRate = baseRate - int32(common.RandValueByRang(int((float64(1)-float64(curValue)/float64(setting.GetLowerLimit()+1))*10000), 50, 100))
			baseRate = baseRate + int32(common.RandValueByRang(int(changeRate-playerRate), 10, 50))

			trueBaseRate = baseRate
		} else {
			//实际计算赔率 = 基础赔率 -（1 - 库存/库存下线）*10000 +
			// Rand（50%-100%）*（调节赔率-实时个人赔率）+
			// 白名单等级*1000-黑名单等级*1000
			baseRate := setting.GetBaseRate()
			baseRate = baseRate - int32((float64(1)-float64(curValue)/float64(setting.GetLowerLimit()+1))*10000)
			baseRate = baseRate + int32(common.RandValueByRang(int(changeRate-playerRate), 50, 100))

			trueBaseRate = baseRate
		}
	case CoinPoolStatus_Normal: //库存下限 < 库存值 < 库存上限
		if playerRate < changeRate {
			//实际计算赔率 = 基础赔率 +（调节赔率 - 实时个人赔率）*Rand(50%-100%)+白名单等级*1000-黑名单等级*1000
			baseRate := setting.GetBaseRate()
			baseRate = baseRate + int32(common.RandValueByRang(int((changeRate-playerRate)), 50, 100))

			trueBaseRate = baseRate
		} else {
			//实际计算赔率 = 基础赔率 +（调节赔率 - 实时个人赔率）*Rand(40%-80%)+白名单等级*1000-黑名单等级*1000
			baseRate := setting.GetBaseRate()
			baseRate = baseRate + int32(common.RandValueByRang(int((changeRate-playerRate)), 40, 80))

			trueBaseRate = baseRate
		}
	case CoinPoolStatus_High: //库存上限 < 库存值 < 库存上限+偏移量
		if playerRate < changeRate {
			//实际计算赔率=基础赔率+(（库存/库存上限-1）*10000+（调节赔率-实时个人赔率）)*Rand(50%-100%)+白名单等级*1000-黑名单等级*1000
			baseRate := setting.GetBaseRate()
			temp := (float64(curValue)/float64(setting.GetUpperLimit())-1)*10000 + float64(changeRate-playerRate)
			baseRate = baseRate + int32(common.RandValueByRang(int(temp), 50, 100))

			trueBaseRate = baseRate
		} else {
			//实际计算赔率=基础赔率+(（库存/库存上限-1）*10000+（调节赔率-实时个人赔率）)*Rand(30%-60%)+白名单等级*1000-黑名单等级*1000
			baseRate := setting.GetBaseRate()
			temp := (float64(curValue)/float64(setting.GetUpperLimit())-1)*10000 + float64(changeRate-playerRate)
			baseRate = baseRate + int32(common.RandValueByRang(int(temp), 30, 60))

			trueBaseRate = baseRate
		}
	case CoinPoolStatus_TooHigh: //库存>库存上限+偏移量
		//吐分量=库存-库存上限
		//吐分概率增益=（库存-库存上限）/库存上限
		richRate := float64(curValue-int64(setting.GetUpperLimit())) / float64(setting.GetUpperLimit())
		if playerRate < changeRate {
			//实际计算赔率=基础赔率+(吐分概率增益*10000+（调节赔率-实时个人赔率）)*Rand(50%-100%)+白名单等级*1000-黑名单等级*1000
			baseRate := setting.GetBaseRate()
			baseRate = baseRate + int32(richRate)*10000
			baseRate = baseRate + int32(common.RandValueByRang(int(changeRate-playerRate), 50, 100))

			trueBaseRate = baseRate
		} else {
			//实际计算赔率=基础赔率+(吐分概率增益*10000+（调节赔率-实时个人赔率）)*Rand(25%-50%)+白名单等级*1000-黑名单等级*1000
			baseRate := setting.GetBaseRate()
			baseRate = baseRate + int32(richRate)*10000
			baseRate = baseRate + int32(common.RandValueByRang(int(changeRate-playerRate), 25, 50))

			trueBaseRate = baseRate
		}
	}

	//每组元素开奖概率=实际计算赔率/开奖元素组合总奖励倍数
	if trueBaseRate < 0 {
		trueBaseRate = 0 //负值时默认实际计算赔率=0
	}

	//黑白名单在此处调控，总有客户说黑白名单无法控制,减少调控导致的影响
	if playerEx.WhiteLevel+playerEx.WhiteFlag > 0 {
		trueBaseRate = trueBaseRate + int32(10000*3*(playerEx.WhiteLevel+playerEx.WhiteFlag)/10.0)
	}

	if playerEx.BlackLevel > 0 {
		trueBaseRate = trueBaseRate - int32(float64(trueBaseRate)*(float64(playerEx.BlackLevel)/10.0)*0.9)
	}

	rate := int(float64(trueBaseRate) / float64(allscore))

	if rate > common.RandInt(10000) {
		return true
	} else {
		//从不中奖元素中随机选择一组出来作为开奖结果
		return false
	}
}

////////////////////////////////////////////////////////////////////
/// Module Implement [beg]
////////////////////////////////////////////////////////////////////
func (this *CoinPoolManager) ModuleName() string {
	return "CoinPoolManager"
}

func (this *CoinPoolManager) Init() {
	this.CoinPoolDBKey = fmt.Sprintf("CoinPoolManager_Srv%v", common.GetSelfSrvId())
	data := model.GetStrKVGameData(this.CoinPoolDBKey)
	coinPool := make(map[string]int64)
	err := json.Unmarshal([]byte(data), &coinPool)
	if err == nil {
		for key, value := range coinPool {
			this.CoinPool.Store(key, value)

		}
	} else {
		logger.Logger.Error("Unmarshal coin pool error:", err)
	}

	this.ProfitPoolDBKey = fmt.Sprintf("ProfitPoolManager_Srv%v", common.GetSelfSrvId())
	data = model.GetStrKVGameData(this.ProfitPoolDBKey)
	profitPool := make(map[string]int64)
	err = json.Unmarshal([]byte(data), &profitPool)
	if err == nil {
		for key, value := range profitPool {
			this.ProfitPool.Store(key, value)
		}
	} else {
		logger.Logger.Error("Unmarshal profit pool error:", err)
	}
}

func (this *CoinPoolManager) Update() {
	if this.dirty {
		this.dirty = false
		this.Save(true)
	}

	//var logs []*model.CoinPoolPump
	tNow := time.Now()
	for key, setting := range this.CoinPoolSetting {
		var changeRate int32
		if setting.GetCpCangeType() == 0 {
			changeRate = setting.GetChangeRate()
		} else {
			weightData := []int{}
			if len(this.LastDayDtCount) > 0 {
				weightData = this.LastDayDtCount
			} else {
				weightData = model.GameParamData.CoinPoolChangeWeight
			}
			weight := common.SliceValueWeight(weightData, time.Now().Hour())
			totleChangeCoin := setting.GetCpChangeTotle()
			changeRate = int32((float64(totleChangeCoin) * weight) / 60)
		}
		hasCoin := this._loadCoin(key)
		if setting.GetCpChangeLower() != 0 && changeRate < 0 && hasCoin < int64(setting.GetCpChangeLower()) {
			changeRate = 0
		} else if setting.GetCpChangeUpper() != 0 && changeRate > 0 && hasCoin > int64(setting.GetCpChangeUpper()) {
			changeRate = 0
		}
		//var changeVal int64
		if changeRate < 0 {
			//if hasCoin >= int64(-changeRate) {
			//	this._popCoin(key, int64(-changeRate))
			//	changeVal = int64(changeRate)
			//} else if hasCoin > 0 {
			//	this._popCoin(key, hasCoin)
			//	changeVal = -hasCoin
			//}
			//<0也要允许往下减,有这个需求
			this._popCoin(key, int64(-changeRate))
			//changeVal = int64(changeRate)
		} else if changeRate > 0 {
			this._pushCoin(key, int64(changeRate))
			//changeVal = -int64(changeRate)
		}

		if setting.GetResetTime() != 0 {
			var runTime int32
			if t, ok := this.curRunTime[key]; ok {
				runTime = t
			} else {
				runTime = int32(tNow.Unix()) + setting.GetResetTime()
				this.curRunTime[key] = runTime
			}

			if runTime < int32(tNow.Unix()) {
				runTime += setting.GetResetTime()
				this.curRunTime[key] = runTime
				initValue := setting.GetInitValue()
				var value int64
				if initValue != 0 { //初始化水池
					value = int64(initValue)
				}
				this.CoinPool.Store(key, value)
			}
		}
	}

	this.lastTime = tNow
}

func (this *CoinPoolManager) Shutdown() {
	this.Save(false)
	module.UnregisteModule(this)
}

func (this *CoinPoolManager) Save(asyn bool) {
	coinPool := make(map[string]int64)
	this.CoinPool.Range(func(key, value interface{}) bool {
		var cache int64
		if !asyn {
			cache = this.PoolCache[key.(string)]
			delete(this.PoolCache, key.(string))
		}
		coinPool[key.(string)] = value.(int64) + cache
		return true
	})
	buff, err := json.Marshal(coinPool)
	if err == nil {
		if asyn {
			task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
				return model.UptStrKVGameData(this.CoinPoolDBKey, string(buff))
			}), nil, "UptStrKVGameData").Start()
		} else {
			model.UptStrKVGameData(this.CoinPoolDBKey, string(buff))
		}
	} else {
		logger.Logger.Error("Marshal coin pool error:", err)
	}

	profitPool := make(map[string]int64)
	this.ProfitPool.Range(func(key, value interface{}) bool {
		profitPool[key.(string)] = value.(int64)
		return true
	})
	buffProfit, err := json.Marshal(profitPool)
	if err == nil {
		if asyn {
			task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
				return model.UptStrKVGameData(this.ProfitPoolDBKey, string(buffProfit))
			}), nil, "UptStrKVGameData").Start()
		} else {
			model.UptStrKVGameData(this.ProfitPoolDBKey, string(buffProfit))
		}
	} else {
		logger.Logger.Error("Marshal profit pool error:", err)
	}
}

func (this *CoinPoolManager) OnMiniTimer() {

}

func (this *CoinPoolManager) OnHourTimer() {

}

func (this *CoinPoolManager) OnDayTimer() {
	logger.Logger.Info("(this *CoinPoolManager) OnDayTimer")
}

func (this *CoinPoolManager) OnWeekTimer() {
	logger.Logger.Info("(this *CoinPoolManager) OnWeekTimer")

}

func (this *CoinPoolManager) OnMonthTimer() {
	logger.Logger.Info("(this *CoinPoolManager) OnMonthTimer")
	// 每月1号清空营收池
	this.ClearProfitPool()
}

func (this *CoinPoolManager) ClearProfitPool() {
	var logs []*model.MonthlyProfitPool
	this.ProfitPool.Range(func(key, value interface{}) bool {
		gamefreeid, groupid, platform, ok := this.SplitKey(key.(string))
		if ok {
			logs = append(logs, model.NewMonthlyProfitPoolEx(int32(common.GetSelfSrvId()), gamefreeid, groupid, platform, value.(int64)))
		}
		return true
	})
	this.ProfitPool = new(sync.Map)
	profitPool := make(map[string]int64)
	buffProfit, err := json.Marshal(profitPool)
	task.New(nil, task.CallableWrapper(func(o *basic.Object) interface{} {
		err = model.UptStrKVGameData(this.ProfitPoolDBKey, string(buffProfit))
		if err != nil {
			logger.Logger.Errorf("UptStrKVGameData err:%v, val:%v", err, string(buffProfit))
		}
		err = model.InsertMonthlyProfitPool(logs...)
		if err != nil {
			logger.Logger.Errorf("InsertMonthlyProfitPool err:%v logs:%v", err, logs)
		}
		return err
	}), nil, "UptStrKVGameData").Start()
}

func (this *CoinPoolManager) GetTax() float64 {
	return this.tax
}
func (this *CoinPoolManager) SetTax(tax float64) {
	this.tax = tax
}

//////////////
var _CoinPoolFishListener = &CoinPoolFishListener{}

type CoinPoolFishListener struct {
}

func (f CoinPoolFishListener) OnSettingUpdate(oldSetting, newSetting *webapi.CoinPoolSetting) {
	if oldSetting == nil || oldSetting.GetCtroRate() != newSetting.GetCtroRate() {
		keyPf := fmt.Sprintf("%v_%v", newSetting.GetPlatform(), newSetting.GetGameFreeId())
		SysProfitCoinMgr.Del(keyPf)
	}
}

func init() {
	module.RegisteModule(CoinPoolMgr, time.Minute, 0)
	CoinPoolMgr.RegisteListener(_CoinPoolFishListener)
	//RegisteDayTimeChangeListener(CoinPoolMgr)
}
