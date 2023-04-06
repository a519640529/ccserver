package base

import (
	"games.yol.com/win88/common"
	"games.yol.com/win88/srvdata"
	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/logger"
)

var SrvDataMgrEx = &SrvDataManagerEx{
	DataReloader: make(map[string]SrvDataReloadInterface),
}

type SrvDataManagerEx struct {
	DataReloader map[string]SrvDataReloadInterface
}

func RegisterDataReloader(fileName string, sdri SrvDataReloadInterface) {
	SrvDataMgrEx.DataReloader[fileName] = sdri
}

type SrvDataReloadInterface interface {
	Reload()
}

//奔驰宝马表格参数
var SystemChanceMgrEx = &PBDB_SystemChanceMgrEx{}

type PBDB_SystemChanceMgrEx struct {
	RollCoinRate   []int32 //赔率数组
	RollCoinChance []int64 //概率数组
	RollCoinIds    []int32
}

func (this *PBDB_SystemChanceMgrEx) Reload() {
	this.RollCoinIds = nil
	this.RollCoinRate = nil
	this.RollCoinChance = nil
	for _, value := range srvdata.PBDB_SystemChanceMgr.Datas.Arr {
		this.RollCoinIds = append(this.RollCoinIds, value.GetId())
		this.RollCoinRate = append(this.RollCoinRate, value.GetCoin())
		this.RollCoinChance = append(this.RollCoinChance, int64(value.GetId()), int64(value.GetRate()))
	}
	logger.Logger.Trace(this)
}
func (this *PBDB_SystemChanceMgrEx) GetRollCoinIds() []int32 {
	var rollCoinIds = []int32{}
	for _, value := range SystemChanceMgrEx.RollCoinIds {
		rollCoinIds = append(rollCoinIds, value)
	}
	for i := 0; i < len(rollCoinIds); i++ {
		rollIndex := common.RandInt(len(rollCoinIds))
		rollCoinIds[i], rollCoinIds[rollIndex] = rollCoinIds[rollIndex], rollCoinIds[i]
	}
	return rollCoinIds
}

////飞禽走兽表格参数
//var AnimalsChanceMgrEx = &PBDB_AnimalsChanceMgrEx{}
//
//type PBDB_AnimalsChanceMgrEx struct {
//	RollAnimalsIds []int64
//	//RollAnimalsOdds       	[]int64 //赔率数组
//	RollAnimalsOdds map[int64]int64
//	RollAnimalsRate []int64 //概率数组 (id,rate)
//}
//
//func (this *PBDB_AnimalsChanceMgrEx) Reload() {
//	this.RollAnimalsIds = nil
//	this.RollAnimalsOdds = make(map[int64]int64)
//	this.RollAnimalsRate = nil
//	for _, value := range srvdata.PBDB_AnimalsChanceMgr.Datas.Arr {
//		this.RollAnimalsIds = append(this.RollAnimalsIds, int64(value.GetId()))
//		this.RollAnimalsOdds[int64(value.GetId())] = int64(value.GetCoin()[0])
//		this.RollAnimalsRate = append(this.RollAnimalsRate, int64(value.GetId()), int64(value.GetRateA()))
//		if int64(value.GetId()) == int64(rollanimals.RollAnimals_Shark) {
//			this.RollAnimalsOdds[int64(rollanimals.RollAnimals_Big_Shark)] = int64(value.GetCoin()[1])
//		}
//	}
//	this.RollAnimalsOdds[int64(rollanimals.RollAnimals_Bird)] = int64(rollanimals.RollAnimals_BirdOdds)
//	this.RollAnimalsOdds[int64(rollanimals.RollAnimals_Beast)] = int64(rollanimals.RollAnimals_BeastOdds)
//
//	logger.Logger.Trace(this)
//}

//深林舞会表格参数
var AnimalColorMgrEx = &DB_AnimalColorMgrEx{make(map[int32][]int32)}

type DB_AnimalColorMgrEx struct {
	RollColorRate map[int32][]int32 //赔率数组
}

func (this *DB_AnimalColorMgrEx) Reload() {
	for _, value := range srvdata.PBDB_AnimalColorMgr.Datas.Arr {
		this.RollColorRate[value.GetId()] = value.GetColorChance()
	}
	for key, value := range this.RollColorRate {
		if len(value) != 13 {
			logger.Logger.Error("Animal color data reload error.")
			this.RollColorRate[key] = []int32{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
		}
	}
}

var FishTemplateEx = &DB_FishMgrEx{FishPool: make(map[int32]*FishTemplate)}

//初始化鱼模板
type FishTemplate struct {
	ID         int32   //鱼的模板ID
	Name       string  //鱼的昵称
	Exp        int32   //鱼的基础经验
	DropCoin   []int32 //鱼掉落的基础金币数
	Boss       int32   //BOSS
	BaseRateA  []int   //鱼掉落概率
	BaseRateB  []int   //鱼掉落概率
	BaseRateC  []int   //鱼掉落概率
	Speed      int32   //速度
	HP         []int32 //血量
	Ratio      int32   //底分系数
	Ratio1     []int32 //激光炮掉血比例
	FishType   int32   //鱼本身的分类
	RandomCoin string  //掉落金币的倍数权重
	Rate       int32   //概率 大闹天宫尝试
	Jackpot    int32   //奖金鱼概率(天天捕鱼)
	DealRate   int32   //致命一击概率(天天捕鱼)
}
type DB_FishMgrEx struct {
	FishPool map[int32]*FishTemplate
}

func (this *DB_FishMgrEx) Reload() {
	for _, value := range srvdata.PBDB_FishMgr.Datas.Arr {
		ft := &FishTemplate{}
		ft.ID = value.GetId()
		ft.Name = value.GetName()
		ft.Exp = value.GetExp()
		ft.Boss = value.GetIsBoss()
		ft.Speed = value.GetSpeed()
		ft.FishType = value.GetFishType()
		ft.RandomCoin = value.GetRandomCoin()
		this.FishPool[value.GetId()] = ft
	}
	for _, value := range srvdata.PBDB_FishRateMgr.Datas.Arr {
		if fish, ok := this.FishPool[value.GetId()]; ok {
			fish.DropCoin = value.GetGold()
			if len(fish.DropCoin) < 2 {
				fish.DropCoin = append(fish.DropCoin, 0, 0)
				logger.Logger.Errorf("Fish %v drop coin data error.", fish.ID)
			}
			if len(value.GetRateA()) != 2 {
				logger.Logger.Errorf("%v rate error.", value.GetName())
				fish.BaseRateA = []int{1000, 1000}
			} else {
				fish.BaseRateA = nil
				fish.BaseRateA = append(fish.BaseRateA, int(value.GetRateA()[0]), int(value.GetRateA()[1]))
			}
			if len(value.GetRateB()) != 2 {
				logger.Logger.Errorf("%v rate error.", value.GetName())
				fish.BaseRateB = []int{1000, 1000}
			} else {
				fish.BaseRateB = nil
				fish.BaseRateB = append(fish.BaseRateB, int(value.GetRateB()[0]), int(value.GetRateB()[1]))
			}
			if len(value.GetRateC()) != 2 {
				logger.Logger.Errorf("%v rate error.", value.GetName())
				fish.BaseRateC = []int{1000, 1000}
			} else {
				fish.BaseRateC = nil
				fish.BaseRateC = append(fish.BaseRateC, int(value.GetRateC()[0]), int(value.GetRateC()[1]))
			}
		} else {
			logger.Logger.Errorf("%v-%v not find in fish pool.", value.GetName(), value.GetId())
		}
	}
	for _, value := range srvdata.PBDB_FishHPMgr.Datas.Arr {
		if fish, ok := this.FishPool[value.GetId()]; ok {
			fish.HP = value.GetGold() // 血量就是金币
			fish.Jackpot = value.GetRate()
			fish.DealRate = value.GetRate()
			fish.Ratio = value.GetRatio()
			fish.Ratio1 = value.GetRatio1()
		}
	}
	/*for _, value := range srvdata.PBDB_FishLRMgr.Datas.Arr {
		if fish, ok := this.FishPool[value.GetId()]; ok { // 概率
			fish.Rate = value.GetLimitRate()
		}
	}*/
}

/*
var FishPoolEx = &DB_FishPoolEx{BigPool: new(sync.Map), SmallPool: new(sync.Map)}

type DB_FishPoolEx struct {
	SmallPool     *sync.Map
	BigPool       *sync.Map
	SmallPoolUper [FishSceneTypeMax]float64
	BigPoolUper   [FishSceneTypeMax]float64
}
type FishPoolData struct {
	MinValue int64
	MaxValue int64
	Rate     int32
}

func (this *DB_FishPoolEx) Reload() {
	this.SmallPool.Range(func(key, value interface{}) bool {
		this.SmallPool.Delete(key)
		return true
	})
	this.BigPool.Range(func(key, value interface{}) bool {
		this.SmallPool.Delete(key)
		return true
	})
	for _, value := range srvdata.PBDB_FishPoolMgr.Datas.Arr {
		data := FishPoolData{
			MinValue: int64(value.GetPoolMin()),
			MaxValue: int64(value.GetPoolMax()),
			Rate:     value.GetRateAdd(),
		}
		switch value.GetPoolType() {
		case FishSmallCoinPool:
			if arr, ok := this.SmallPool.Load(value.GetSceneType()); ok {
				this.SmallPool.Store(value.GetSceneType(), append(arr.([]FishPoolData), data))
			} else {
				this.SmallPool.Store(value.GetSceneType(), []FishPoolData{data})
			}
			if float64(data.MaxValue) > this.SmallPoolUper[value.GetSceneType()] {
				this.SmallPoolUper[value.GetSceneType()] = float64(data.MaxValue)
			}
		case FishBigCoinPool:
			if arr, ok := this.BigPool.Load(value.GetSceneType()); ok {
				this.BigPool.Store(value.GetSceneType(), append(arr.([]FishPoolData), data))
			} else {
				this.BigPool.Store(value.GetSceneType(), []FishPoolData{data})
			}
			if float64(data.MaxValue) > this.BigPoolUper[value.GetSceneType()] {
				this.BigPoolUper[value.GetSceneType()] = float64(data.MaxValue)
			}
		default:
			logger.Logger.Error("Fish pool error data:", value)
		}
	}
}
func (this *DB_FishPoolEx) CalcPoolRate(sceneType, poolType int32, curPool int64) int {
	switch poolType {
	case FishSmallCoinPool:
		if value, ok := this.SmallPool.Load(sceneType); ok {
			poolData := value.([]FishPoolData)
			for _, value := range poolData {
				if value.MinValue <= curPool && curPool <= value.MaxValue {
					return int(value.Rate)
				}
			}
		}
	case FishBigCoinPool:
		if value, ok := this.BigPool.Load(sceneType); ok {
			poolData := value.([]FishPoolData)
			for _, value := range poolData {
				if value.MinValue <= curPool && curPool <= value.MaxValue {
					return int(value.Rate)
				}
			}
		}
	default:
		logger.Logger.Errorf("Can't get %v-%v pool rate.", sceneType, poolType)
	}
	return 0
}
*/

//初始化在线奖励系统
func init() {
	core.RegisteHook(core.HOOK_BEFORE_START, func() error {
		logger.Logger.Info("初始化牌库[S]")
		defer logger.Logger.Info("初始化牌库[E]")
		return nil
	})
}
func init() {
	srvdata.SrvDataModifyCB = func(fileName string, fullName string) {
		if dr, ok := SrvDataMgrEx.DataReloader[fileName]; ok {
			dr.Reload()
		}
	}
	RegisterDataReloader("DB_SystemChance.dat", SystemChanceMgrEx)
	RegisterDataReloader("DB_AnimalColor.dat", AnimalColorMgrEx)
	RegisterDataReloader("DB_Fish.dat", FishTemplateEx)
	RegisterDataReloader("DB_FishRate.dat", FishTemplateEx)
	RegisterDataReloader("DB_FishHP.dat", FishTemplateEx)
	//RegisterDataReloader("DB_FishLR.dat", FishTemplateEx)
	//RegisterDataReloader("DB_FishPool.dat", FishPoolEx)
}
