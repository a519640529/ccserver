package main

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"games.yol.com/win88/common"
	"games.yol.com/win88/gamerule/crash"
	"games.yol.com/win88/mq"
	"github.com/idealeak/goserver/core/broker/rabbitmq"

	"sync"

	"games.yol.com/win88/model"
	"games.yol.com/win88/webapi"
	"github.com/astaxie/beego/cache"
	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/utils"
)

type ParalleFunc func() error

var CacheMemory cache.Cache
var wgParalleLoad = &sync.WaitGroup{}
var ParalleLoadModules []*ParalleLoadModule
var RabbitMQPublisher *mq.RabbitMQPublisher
var RabbitMqConsumer *mq.RabbitMQConsumer

type ParalleLoadModule struct {
	name string
	f    ParalleFunc
}

func RegisteParallelLoadFunc(name string, f ParalleFunc) {
	ParalleLoadModules = append(ParalleLoadModules, &ParalleLoadModule{name: name, f: f})
}

func init() {
	rand.Seed(time.Now().UnixNano())

	//首先加载游戏配置
	core.RegisteHook(core.HOOK_BEFORE_START, func() error {
		model.StartupRPClient(common.CustomConfig.GetString("MgoRpcCliNet"), common.CustomConfig.GetString("MgoRpcCliAddr"), time.Duration(common.CustomConfig.GetInt("MgoRpcCliReconnInterV"))*time.Second)
		model.InitGameParam()
		model.InitNormalParam()

		hs := model.GetCrashHash(0)
		if hs == nil {
			for i := 0; i < crash.POKER_CART_CNT; i++ {
				model.InsertCrashHash(i, crash.Sha256(fmt.Sprintf("%v%v", i, time.Now().UnixNano())))
			}
			hs = model.GetCrashHash(0)
		}
		model.GameParamData.InitGameHash = []string{}
		for _, v := range hs {
			model.GameParamData.InitGameHash = append(model.GameParamData.InitGameHash, v.Hash)
		}
		hsatom := model.GetCrashHashAtom(0)
		if hsatom == nil {
			for i := 0; i < crash.POKER_CART_CNT; i++ {
				model.InsertCrashHashAtom(i, crash.Sha256(fmt.Sprintf("%v%v", i, time.Now().UnixNano())))
			}
			hsatom = model.GetCrashHashAtom(0)
		}
		model.GameParamData.AtomGameHash = []string{}
		for _, v := range hsatom {
			model.GameParamData.AtomGameHash = append(model.GameParamData.AtomGameHash, v.Hash)
		}
		if len(model.GameParamData.AtomGameHash) < crash.POKER_CART_CNT ||
			len(model.GameParamData.InitGameHash) < crash.POKER_CART_CNT {
			panic(errors.New("hash is read error"))
		}
		return nil
	})

	//RegisteParallelLoadFunc("平台红包数据", func() error {
	//	actRandCoinMgr.LoadPlatformData()
	//	return nil
	//})

	RegisteParallelLoadFunc("GMAC", func() error {
		model.InitGMAC()
		return nil
	})

	RegisteParallelLoadFunc("三方游戏配置", func() error {
		model.InitGameConfig()
		return nil
	})

	RegisteParallelLoadFunc("GameKVData", func() error {
		model.InitGameKVData()
		return nil
	})

	RegisteParallelLoadFunc("水池配置", func() error {
		return model.GetAllCoinPoolSettingData()
	})

	RegisteParallelLoadFunc("三方平台热载数据设置", func() error {
		f := func() {
			webapi.ReqCgAddr = model.GameParamData.CgAddr
			if plt, ok := webapi.ThridPlatformMgrSington.ThridPlatformMap.Load("XHJ平台"); ok {
				plt.(*webapi.XHJThridPlatform).IsNeedCheckQuota = model.GameParamData.FGCheckPlatformQuota
				plt.(*webapi.XHJThridPlatform).ReqTimeOut = model.GameParamData.ThirdPltReqTimeout
			}
		}
		f()
		model.GameParamData.Observers = append(model.GameParamData.Observers, f)
		return nil
	})

	core.RegisteHook(core.HOOK_BEFORE_START, func() error {

		//for _, v := range data {
		//	PlatformMgrSington.UpsertPlatform(v.Name, v.Isolated, v.GameStatesData)
		//}

		//ps := []model.GamePlatformState{model.GamePlatformState{LogicId:130000001,Param:"",State:1},model.GamePlatformState{LogicId:150000001,Param:"",State:1}}

		//model.InsertPlatformGameConfig("360",true,ps)

		var err error
		CacheMemory, err = cache.NewCache("memory", `{"interval":60}`)
		if err != nil {
			return err
		}

		//etcd打开连接
		EtcdMgrSington.Init()

		//go func() {
		//	for {
		//		time.Sleep(time.Minute)
		//		EtcdMgrSington.Reset()
		//	}
		//}()

		//rabbitmq打开链接
		RabbitMQPublisher = mq.NewRabbitMQPublisher(common.CustomConfig.GetString("RabbitMQURL"), rabbitmq.Exchange{Name: common.CustomConfig.GetString("RMQExchange"), Durable: true}, common.CustomConfig.GetInt("RMQPublishBacklog"))
		if RabbitMQPublisher != nil {
			err = RabbitMQPublisher.Start()
			if err != nil {
				panic(err)
			}
		}

		RabbitMqConsumer = mq.NewRabbitMQConsumer(common.CustomConfig.GetString("RabbitMQURL"), rabbitmq.Exchange{Name: common.CustomConfig.GetString("RMQExchange"), Durable: true})
		if RabbitMqConsumer != nil {
			RabbitMqConsumer.Start()
		}

		//初始化本地机器人id
		LocalRobotIdMgrSington.Init()

		//开始并行加载数据
		//改为串行加载,后台并发有点扛不住
		paralleCnt := len(ParalleLoadModules)
		if paralleCnt != 0 {
			tStart := time.Now()
			logger.Logger.Infof("===[开始串行加载]===")
			//wgParalleLoad.Add(paralleCnt)
			for _, m := range ParalleLoadModules {
				/*go*/ func(plm *ParalleLoadModule) {
					ts := time.Now()
					defer func() {
						utils.DumpStackIfPanic(plm.name)
						//wgParalleLoad.Done()
						logger.Logger.Infof("[串行加载结束][%v] 花费[%v]", plm.name, time.Now().Sub(ts))
					}()
					logger.Logger.Infof("[开始串行加载][%v] ", plm.name)

					err := plm.f()
					if err != nil {
						logger.Logger.Warnf("[串行加载][%v][error:%v]", plm.name, err)
					}
				}(m)
			}
			//wgParalleLoad.Wait()
			logger.Logger.Infof("===[串行加载结束,耗时:%v]===", time.Now().Sub(tStart))
		}
		return nil
	})

	core.RegisteHook(core.HOOK_AFTER_STOP, func() error {
		//etcd关闭连接
		EtcdMgrSington.Shutdown()
		//关闭rabbitmq连接
		if RabbitMQPublisher != nil {
			RabbitMQPublisher.Stop()
		}

		if RabbitMqConsumer != nil {
			RabbitMqConsumer.Stop()
		}

		//model.ShutdownRPClient()
		return nil
	})
	core.RegisteHook(core.HOOK_BEFORE_START, func() error {
		ThirdPltGameMappingConfig.Init()
		return nil
	})
	//RegisteParallelLoadFunc("分层配置数据", func() error {
	//	//加载分层配置
	//	LogicLevelMgrSington.LoadConfig()
	//	return nil
	//})
}
