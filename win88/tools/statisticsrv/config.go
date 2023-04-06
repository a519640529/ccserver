package main

import (
	"github.com/go-ini/ini"
)

const (
	DEFAULT_CONFIGFILE_NAME = "my.ini"
	DEFAULT_SERVER_PORT     = 8080
	DEFAULT_MGO_MAXDONE     = 1024
	DEFAULT_MGO_MAXINTERVAL = 5
)

var AppCfg Config

type ServerConfig struct {
	Port int
}

type MongoConfig struct {
	Host        string
	Database    string
	UserName    string
	Password    string
	MaxDone     int
	MaxInterval int
}

type Config struct {
	GinMode string
	SC      ServerConfig
	MC      MongoConfig
}

func LoadConfig() error {
	cfg, err := ini.Load(DEFAULT_CONFIGFILE_NAME)
	if err != nil {
		return err
	}

	AppCfg.GinMode = cfg.Section("").Key("app_mode").String()
	srvsec := cfg.Section("server")
	if srvsec != nil {
		port, err := srvsec.Key("http_port").Int()
		if err == nil {
			AppCfg.SC.Port = port
		} else {
			AppCfg.SC.Port = DEFAULT_SERVER_PORT
		}
	} else {
		AppCfg.SC.Port = DEFAULT_SERVER_PORT
	}

	mgosec := cfg.Section("mongo")
	if mgosec != nil {
		AppCfg.MC.Database = mgosec.Key("database").String()
		AppCfg.MC.Host = mgosec.Key("host").String()
		AppCfg.MC.UserName = mgosec.Key("username").String()
		AppCfg.MC.Password = mgosec.Key("password").String()
		maxdone, err := mgosec.Key("maxdone").Int()
		if err == nil {
			AppCfg.MC.MaxDone = maxdone
		} else {
			AppCfg.MC.MaxDone = DEFAULT_MGO_MAXDONE
		}
		maxinterval, err := mgosec.Key("maxinterval").Int()
		if err == nil {
			AppCfg.MC.MaxInterval = maxinterval
		} else {
			AppCfg.MC.MaxInterval = DEFAULT_MGO_MAXINTERVAL
		}
	}
	return nil
}
