package base

import (
	"github.com/idealeak/goserver/core/module"
)

type GameDataMgr struct {
}

func (this *GameDataMgr) ModuleName() string {
	return "GameDataMgr"
}

func (this *GameDataMgr) Init() {
	//model.InitGameData()
}

func (this *GameDataMgr) Update() {
	//model.SaveGameData()
}

func (this *GameDataMgr) Shutdown() {
	//model.SaveGameData()
	module.UnregisteModule(this)
}

func init() {
	//module.RegisteModule(&GameDataMgr{}, time.Minute, 100)
}
