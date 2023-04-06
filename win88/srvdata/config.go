package srvdata

import (
	"os"
	"path/filepath"
	"strings"

	"games.yol.com/win88/model"

	"github.com/howeyc/fsnotify"
	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/logger"
)

var Config = Configuration{}

var SrvDataModifyCB func(string, string)

type Configuration struct {
	RootPath string   //文件目录
	Files    []string //废弃，现在需要装载的文件过多，不再一个一个添加到这个文件数组了
	LoadLib  bool     //是否装载牌库文件
	watcher  *fsnotify.Watcher
}

func (this *Configuration) Name() string {
	return "data"
}

func (this *Configuration) Init() error {
	workDir, workingDirError := os.Getwd()
	if workingDirError != nil {
		return workingDirError
	}
	this.RootPath = filepath.Join(workDir, this.RootPath)
	var err error
	this.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		logger.Logger.Info(" fsnotify.NewWatcher err:", err)
	}

	// Process events
	go func() {
		defer func() {
			if err := recover(); err != nil {
				logger.Logger.Warn("watch data director modify goroutine err:", err)
			}
		}()
		for {
			select {
			case ev, ok := <-this.watcher.Event:
				if ok && ev != nil {
					if ev.IsModify() || ev.IsRename() {
						obj := core.CoreObject()
						if filepath.Ext(ev.Name) == ".dat" {
							if obj != nil {
								obj.SendCommand(&fileModifiedCommand{fileName: ev.Name}, false)
							}
							logger.Logger.Info("fsnotify event:", ev)
						} else if filepath.Ext(ev.Name) == ".json" {
							if obj != nil {
								obj.SendCommand(&gameParamModifiedCommand{fileName: ev.Name}, false)
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
		logger.Logger.Warn(this.RootPath, " watcher quit!")
	}()
	loaderFiles := []string{}
	this.watcher.Watch(this.RootPath)
	filepath.Walk(this.RootPath, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			this.watcher.Watch(path)
		}
		name := info.Name()
		if filepath.Ext(name) != ".dat" {
			return nil
		}

		loaderFiles = append(loaderFiles, name)
		return nil
	})
	for _, fileName := range loaderFiles {
		if IsLibFile(fileName) && this.LoadLib == false { //是牌库文件，但是不安排装载的话，就不加载库文件了
			continue
		}
		loader := DataMgr.GetLoader(fileName)
		if loader != nil {
			fullPath := filepath.Join(this.RootPath, fileName)
			err := loader.load(fullPath)
			if err != nil {
				logger.Logger.Info(fileName, " loader err:", err)
			}
			if SrvDataModifyCB != nil {
				SrvDataModifyCB(filepath.Base(fileName), fileName)
			}
		} else {
			logger.Logger.Warn(fileName, " no loader")
		}
	}
	//配桌管理
	PlayerTypeMgrSington.updateData()

	err = initSensitiveWordTree()
	if err != nil {
		return err
	}
	err = updateClientVers()
	if err != nil {
		return err
	}
	//DB_Createroom
	CreateRoomMgrSington.Init()
	GameDropMgrSington.Init()
	return nil
}
func IsLibFile(name string) bool {
	strArr := strings.Split(name, "_")
	if len(strArr) < 1 {
		return false
	}
	if strArr[0] == "LB" {
		return true
	} else {
		return false
	}
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
	loader := DataMgr.GetLoader(fn)
	if loader != nil {
		err := loader.reload(fmc.fileName)
		if err != nil {
			logger.Logger.Info(fn, " loader err:", err)
		}
	}

	switch fn {
	case "DB_Sensitive_Words.dat":
		sensitiveWordsTree = make(map[rune]*Word)
		err := initSensitiveWordTree()
		if err != nil {
			return err
		}
	case "DB_ClientVer.dat":
		err := updateClientVers()
		if err != nil {
			return err
		}
	case "DB_PlayerType.dat":
		PlayerTypeMgrSington.updateData()
	case "DB_Createroom.dat":
		CreateRoomMgrSington.Init()
	case "DB_Game_Drop.dat":
		GameDropMgrSington.Init()
	}

	//数据变动回调
	if SrvDataModifyCB != nil {
		SrvDataModifyCB(fn, fmc.fileName)
	}
	return nil
}

type gameParamModifiedCommand struct {
	fileName string
}

func (gmc *gameParamModifiedCommand) Done(o *basic.Object) error {
	fn := filepath.Base(gmc.fileName)
	switch fn {
	case "gameparam.json":
		model.InitGameParam()
	case "gmac.json":
		model.InitGMAC()
	case "thrconfig.json":
		model.InitGameConfig()
	case "fishingparam.json":
		model.InitFishingParam()
	//case "bullfightparam.json":
	//	model.InitBullFightParam()
	//case "winthreeparam.json":
	//	model.InitWinThreeParam()
	//case "mahjongparam.json":
	//	model.InitMahJongParam()
	case "normalparam.json":
		model.InitNormalParam()
	}

	//数据变动回调
	if SrvDataModifyCB != nil {
		SrvDataModifyCB(fn, gmc.fileName)
	}
	return nil
}

func init() {
	core.RegistePackage(&Config)
}
