package base

import (
	"time"

	"math/rand"

	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/module"
	"github.com/idealeak/goserver/core/netlib"
)

var (
	Config           = Configuration{}
	BenchMarkModule  = &BenchMark{}
	WaitReconnectCfg = []*netlib.SessionConfig{}
	NewReconnectCfg  = [24][]*netlib.SessionConfig{}
)

const (
	ROBOT_SESSEION_START_ID = 100000000
)

type Configuration struct {
	Count         int
	AppId         string
	AccountHost   string
	AccountPrefix string
	GameId        int
	GameMode      int
	RoomId        int
	IgnoreMatchId []int32
	HeadUrl       string
	Doudizhu      bool
	Connects      netlib.SessionConfig
}

func NewSessionConfiguration() *netlib.SessionConfig {
	BenchMarkModule.idx++
	cfg := Config.Connects
	cfg.Id = BenchMarkModule.idx
	cfg.Init()
	return &cfg
}
func (this *Configuration) Name() string {
	return "benchmark"
}

func (this *Configuration) Init() error {
	return nil
}

func (this *Configuration) Close() error {
	return nil
}

type BenchMark struct {
	idx int
}

func (this BenchMark) ModuleName() string {
	return "benchmark-module"
}

func (this *BenchMark) Init() {
	rand.Seed(time.Now().Unix())
	count := Config.Count
	this.idx = ROBOT_SESSEION_START_ID
	logger.Logger.Infof("Startup [%d] connect", Config.Count)
	for count > 0 {
		cfg := NewSessionConfiguration()
		logger.Logger.Info("Start sno ", cfg.Id, " Client Connect.")
		WaitReconnectCfg = append(WaitReconnectCfg, cfg)
		count--
	}
}

func (this *BenchMark) Update() {
	cnt := len(WaitReconnectCfg)
	if cnt > 0 {
		cfg := WaitReconnectCfg[cnt-1]
		WaitReconnectCfg = WaitReconnectCfg[:cnt-1]
		netlib.Connect(cfg)
	}
	return
}

func (this *BenchMark) Shutdown() {
	module.UnregisteModule(this)
}

func init() {
	core.RegistePackage(&Config)
	module.RegisteModule(BenchMarkModule, time.Millisecond, 1)
}
