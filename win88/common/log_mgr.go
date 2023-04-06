package common

/////////////////////////////////////////////////////////////
//使用方法
//1、定义自己日志模块名字
//const(
//		BaccaratLogger = "BaccaratLogger"
// )
//2、通过自己定义的日志名获得日志实例，调用方式参考如下：
//-------getLoggerInstanceByName(BaccaratLogger).Trace(time.Now().String())
//-------getLoggerInstanceByName(BaccaratLogger).Warn(time.Now().String())
//-------getLoggerInstanceByName(BaccaratLogger).Error(time.Now().String())
//-------getLoggerInstanceByName(BaccaratLogger).Debug(time.Now().String())
//-------getLoggerInstanceByName(BaccaratLogger).Info(time.Now().String())
//-------getLoggerInstanceByName(BaccaratLogger).Flush()
//3、如果自己定义的日志名在配置文件中没有找到，则使用默认的全局日志
//4、可动态添加自己的日志配置，添加后即可生效
//5、注意确保自己定义的日志模块名与配置日志的文件名一样
//6、同时确保自己配置日志的文件中输出日志的文件名，参考如下：
//	..........
//	..........												↓↓↓这个日志输出的文件名
//		 <rollingfile formatid="all" type="size" filename="./all.log"
//	..........
//	..........
//////////////////////////////////////////////////////////////

import (
	"github.com/cihub/seelog"
	"github.com/howeyc/fsnotify"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/utils"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"sync"
)

var g_LogSet sync.Map

// 获得日志实例
func GetLoggerInstanceByName(name string) (log seelog.LoggerInterface) {
	if v, ok := g_LogSet.Load(name); ok {
		return v.(seelog.LoggerInterface)
	}
	return logger.Logger
}
func init() {
	go utils.CatchPanic(watchFile)
}
func watchFile() {
	watch, err := fsnotify.NewWatcher()
	if err != nil {
		logger.Error(err)
		return
	}
	defer watch.Close()
	var workingDir string
	workDir, workingDirError := os.Getwd()
	if workingDirError != nil {
		workingDir = string(os.PathSeparator)
		logger.Error(workingDirError)
		return
	}
	workingDir = workDir + string(os.PathSeparator)
	err = watch.Watch(workingDir)
	if err != nil {
		logger.Error(err)
		return
	}

	dir_list, e := ioutil.ReadDir(workingDir)
	if e != nil {
		logger.Error("read dir error")
		return
	}
	for _, v := range dir_list {
		if strings.Contains(v.Name(), ".xml") && v.Name() != "logger.xml" {
			log, err := seelog.LoggerFromConfigAsFile(v.Name())
			if err != nil {
				logger.Error(err)
				break
			}
			g_LogSet.Store(strings.TrimRight(v.Name(), ".xml"), log)
		}
	}
	for {
		select {
		case ev := <-watch.Event:
			if path.Ext(ev.Name) != ".xml" {
				break
			}
			fileName := getFileName(ev.Name)
			//logger.log过滤掉
			if fileName == "logger" {
				break
			}
			{
				if ev.IsCreate() {
					log, err := seelog.LoggerFromConfigAsFile(ev.Name)
					if err != nil {
						logger.Error(err)
						break
					}
					g_LogSet.Store(fileName, log)
				}
				if ev.IsModify() {
					log, err := seelog.LoggerFromConfigAsFile(ev.Name)
					if err != nil {
						logger.Error(err)
						break
					}
					g_LogSet.Store(fileName, log)
				}
				if ev.IsDelete() {
					g_LogSet.Delete(fileName)
				}
				if ev.IsRename() {
					g_LogSet.Delete(fileName)
				}
			}
		case err := <-watch.Error:
			{
				logger.Error("error : ", err)
				return
			}
		}
	}
}
func getFileName(fullPath string) string {
	p := path.Base(strings.Replace(fullPath, "\\", "/", -1))
	return strings.TrimSuffix(p, ".xml")
}
