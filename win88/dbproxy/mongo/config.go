package mongo

import (
	"github.com/howeyc/fsnotify"
	"github.com/idealeak/goserver/core"
	"github.com/idealeak/goserver/core/logger"
)

type Configuration struct {
	CfgFile string
	watcher *fsnotify.Watcher
}

func (c *Configuration) Name() string {
	return "cmgo"
}

func (c *Configuration) Init() error {
	MgoSessionMgrSington.LoadConfig(c.CfgFile)

	var err error
	c.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		logger.Logger.Warnf("%s NewWatcher err:%v", c.CfgFile, err)
		return err
	}
	err = c.watcher.Watch(c.CfgFile)
	if err != nil {
		logger.Logger.Warnf("%s Watch err:%v", c.CfgFile, err)
		return err
	}

	go func() {
		defer func() {
			if err := recover(); err != nil {
				logger.Logger.Warnf("%s watch modify goroutine err:%v", c.CfgFile, err)
			}
		}()
		for {
			select {
			case ev, ok := <-c.watcher.Event:
				if ok && ev != nil {
					if ev.IsModify() {
						MgoSessionMgrSington.LoadConfig(c.CfgFile)
					}
				} else {
					return
				}
			case err := <-c.watcher.Error:
				logger.Logger.Info("fsnotify error:", err)
			}
		}
		logger.Logger.Warnf("%s watcher quit!", c.CfgFile)
	}()
	return nil
}

func (c *Configuration) Close() error {
	c.watcher.Close()
	MgoSessionMgrSington.Close()
	return nil
}

func init() {
	core.RegistePackage(&Configuration{})
}
