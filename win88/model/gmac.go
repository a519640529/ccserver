package model

import (
	"encoding/json"
	"io/ioutil"

	"github.com/idealeak/goserver/core/logger"
)

//Game Manager Access Control

type GMAC struct {
	InviteRobot      int32
	InviteRobotAgent int32
	AgentCreateRoom  int32
	SpecialCareOf    int32
	SpecailCareOfIds []int32
	LuYiTingLevel    int32
	ChangeCardLevel  int32
	WhiteList        []string
}

var GMACPath = "../data/gmac.json"
var GMACData = &GMAC{}

func InitGMAC() {
	buf, err := ioutil.ReadFile(GMACPath)
	if err != nil {
		logger.Logger.Warn("InitGMAC ioutil.ReadFile error ->", err)
	}

	err = json.Unmarshal(buf, GMACData)
	if err != nil {
		logger.Logger.Warn("InitGMAC json.Unmarshal error ->", err)
	}

	logger.Logger.Info("InitGMAC param=", GMACData)
}
