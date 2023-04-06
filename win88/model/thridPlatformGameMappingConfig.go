package model

import (
	"encoding/json"
	"github.com/idealeak/goserver/core/logger"
	"io/ioutil"
)

var (
	ThirdPltGameMappingConfig = ThirdPlatformGameMappingConfiguration{}
)

var GameMappingPath = "../data/DB_ThirdPlatformGameMapping.json"

type ThirdPlatformGameMappingConfiguration struct {
	Arr []mappingItem
}
type mappingItem struct {
	SystemGameID      int32
	ThirdPlatformName string
	ThirdGameID       string
	Desc              string
}

func InitGameMappingConfig() {
	buf, err := ioutil.ReadFile(GameMappingPath)
	if err != nil {
		logger.Logger.Error("InitGameMappingConfig ioutil.ReadFile error ->", err)
	}

	err = json.Unmarshal(buf, &ThirdPltGameMappingConfig)
	if err != nil {
		logger.Logger.Error("InitGameMappingConfig json.Unmarshal error ->", err)
		return
	}
}

func (this *ThirdPlatformGameMappingConfiguration) FindByGameID(systemGameID int32) (ok bool, item *mappingItem) {
	for k, v := range this.Arr {
		if v.SystemGameID == systemGameID {
			return true, &this.Arr[k]
		}
	}
	return false, nil
}
