package model

import (
	"encoding/json"
	"io/ioutil"

	"github.com/idealeak/goserver/core/logger"
)

type ThirdConfigData struct {
	Platform  string
	AgentName string
	AgentKey  string
}
type GameConfig struct {
	DgConfig  []ThirdConfigData //Dg的配置信息
	HboConfig []ThirdConfigData //Hbo的配置信息
}

var GameConfigPath = "../data/thrconfig.json"
var GameConfigData = &GameConfig{}

func InitGameConfig() {
	buf, err := ioutil.ReadFile(GameConfigPath)
	if err != nil {
		logger.Logger.Error("InitGameParam ioutil.ReadFile error ->", err)
	}

	err = json.Unmarshal(buf, GameConfigData)
	if err != nil {
		logger.Logger.Error("InitGameParam json.Unmarshal error ->", err)
		return
	}
}

func GetDgConfigByPlatform(platform string) (string, string, string) {

	for _, value := range GameConfigData.DgConfig {
		if value.Platform == platform {
			return value.AgentName, value.AgentKey, "DG"
		}
	}
	for _, value := range GameConfigData.HboConfig {
		if value.Platform == platform {
			return value.AgentName, value.AgentKey, "HBO"
		}
	}
	return "", "", ""
}

func OnlyGetDgConfigByPlatform(platform string) (string, string, string) {

	for _, value := range GameConfigData.DgConfig {
		if value.Platform == platform {
			return value.AgentName, value.AgentKey, "DG"
		}
	}
	return "", "", ""
}

func OnlyGetHboConfigByPlatform(platform string) (string, string, string) {
	for _, value := range GameConfigData.HboConfig {
		if value.Platform == platform {
			return value.AgentName, value.AgentKey, "HBO"
		}
	}
	return "", "", ""
}

func GetAllDgAgent() []string {
	agentNameArr := []string{}
	for _, value := range GameConfigData.DgConfig {
		agentNameArr = append(agentNameArr, value.AgentName)
	}
	return agentNameArr
}

func GetAllHboAgent() []string {
	agentNameArr := []string{}
	for _, value := range GameConfigData.HboConfig {
		agentNameArr = append(agentNameArr, value.AgentName)
	}
	return agentNameArr
}
