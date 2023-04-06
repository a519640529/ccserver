package common

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/howeyc/fsnotify"
	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/basic"
	"github.com/idealeak/goserver/core/logger"
)

var LastModifyConfig int64

func init() {
	core.RegisteHook(core.HOOK_BEFORE_START, func() error {
		var err error
		workDir, err := os.Getwd()
		if err != nil {
			return err
		}
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			return err
		}

		// Process events
		go func() {
			defer func() {
				if err := recover(); err != nil {
					logger.Logger.Warn("watch logger.xml modify goroutine err:", err)
				}
			}()
			for {
				select {
				case ev, ok := <-watcher.Event:
					if ok && ev != nil {
						if ev.IsModify() {
							obj := core.CoreObject()
							if filepath.Base(ev.Name) == "logger.xml" {
								if obj != nil {
									obj.SendCommand(&loggerParamModifiedCommand{fileName: ev.Name}, false)
								}
							}
						}
					} else {
						return
					}
				case err := <-watcher.Error:
					logger.Logger.Warn("fsnotify error:", err)
				}
			}
			logger.Logger.Warn("logger.xml watcher quit!")
		}()
		watcher.Watch(workDir)
		return nil
	})
}

type loggerParamModifiedCommand struct {
	fileName string
}

func (lmc *loggerParamModifiedCommand) Done(o *basic.Object) error {
	logger.Logger.Info("===reload ", lmc.fileName)
	data, err := ioutil.ReadFile(lmc.fileName)
	if err != nil {
		return err
	}
	if len(data) != 0 {
		err = logger.Reload(lmc.fileName)
		if err != nil {
			logger.Logger.Warn("===reload ", lmc.fileName, err)
		}
	}
	return err
}
