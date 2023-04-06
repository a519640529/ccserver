package model

import (
	"encoding/json"
	"io/ioutil"

	"github.com/idealeak/goserver.v3/core/logger"
)

type FishPath struct {
	Id     int32
	Stay   int32
	Length int32
}
type fishPathFile struct {
	Count int32
	Pools []FishPath
}

var FishPathPath = "../data/fishpath/path.json"

func GetFishPath() []FishPath {
	buf, err := ioutil.ReadFile(FishPathPath)
	if err != nil {
		logger.Logger.Warn("GetFishPath ioutil.ReadFile error ->", err)
	}
	var fileData = &fishPathFile{}
	err = json.Unmarshal(buf, fileData)
	if err != nil {
		logger.Logger.Warn("GetFishPath json.Unmarshal error ->", err)
	}
	return fileData.Pools
}
