package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/howeyc/fsnotify"
	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/logger"
)

var fileSignMap = make(map[string]string)
var Config = Configuration{}

type Configuration struct {
	RootPath string
	watcher  *fsnotify.Watcher
}

func (this *Configuration) Name() string {
	return "gameconfig"
}

func (this *Configuration) Init() error {
	workDir, err := os.Getwd()
	if err != nil {
		return err
	}
	this.RootPath = filepath.Join(workDir, this.RootPath)
	this.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		logger.Logger.Info(" fsnotify.NewWatcher err:", err)
		return err
	}

	// Process events
	go func() {
		defer func() {
			if err := recover(); err != nil {
				logger.Logger.Warn("watch gameconfig director modify goroutine err:", err)
			}
		}()
		for {
			select {
			case ev, ok := <-this.watcher.Event:
				if ok && ev != nil {
					if ev.IsModify() {
						obj := core.CoreObject()
						if filepath.Ext(ev.Name) == ".json" {
							if obj != nil {
								obj.SendCommand(&fileModifiedCommand{fileName: ev.Name}, false)
							}
							logger.Logger.Info("fsnotify event:", ev)
						}
					}
				} else {
					return
				}
			case err := <-this.watcher.Error:
				logger.Logger.Info("fsnotify error:", err)
			}
		}
	}()
	this.watcher.Watch(this.RootPath)

	//遍历所有json
	filepath.Walk(this.RootPath, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			if filepath.Ext(info.Name()) == ".json" {
				UpdateGameConfigPolicy(path)
			}
		}
		return nil
	})
	return nil
}

func (this *Configuration) Close() error {
	this.watcher.Close()
	return nil
}

type fileModifiedCommand struct {
	fileName string
}

func (fmc *fileModifiedCommand) Done(o *basic.Object) error {
	fn := filepath.Base(fmc.fileName)
	logger.Logger.Info("modified file name ======", fn)
	return UpdateGameConfigPolicy(fmc.fileName)
}

func UpdateGameConfigPolicy(fullPath string) error {
	//logger.Logger.Info("UpdateGameConfigPolicy file: ", fullPath)
	buf, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return err
	}

	if len(buf) == 0 {
		return nil
	}

	h := md5.New()
	_, err = h.Write(buf)
	if err != nil {
		return err
	}
	fileSign := hex.EncodeToString(h.Sum(nil))
	if preSign, exist := fileSignMap[fullPath]; exist {
		if preSign == fileSign {
			return nil
		}
	}
	fileSignMap[fullPath] = fileSign
	spd := &ScenePolicyData{}
	err = json.Unmarshal(buf, spd)
	if err != nil {
		logger.Logger.Info("UpdateGameConfigPolicy json.Unmarshal error", err)
	}
	if err == nil && spd.Init() {
		for _, m := range spd.GameMode {
			//logger.Logger.Info("New game config ver:", spd.ConfigVer)
			if !CheckGameConfigVer(spd.ConfigVer, spd.GameId, m) {
				//TeaHouseMgr.UpdateGameConfigVer(spd.ConfigVer, spd.GameId, m)
			}
			RegisteScenePolicy(int(spd.GameId), int(m), spd)
		}
	}
	return err
}

func init() {
	core.RegistePackage(&Config)
}
