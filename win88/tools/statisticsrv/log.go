package main

import (
	"github.com/cihub/seelog"
)

var (
	Log seelog.LoggerInterface
)

func init() {
	Log, _ = seelog.LoggerFromConfigAsFile("logger.xml")
	seelog.ReplaceLogger(Log)
}

func Reload(fileName string) error {
	newLogger, err := seelog.LoggerFromConfigAsFile(fileName)
	if err != nil {
		return err
	}
	if newLogger != nil {
		Log = newLogger
		seelog.ReplaceLogger(Log)
	}
	return nil
}
