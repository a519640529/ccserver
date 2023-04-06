package common

import (
	"testing"
	"time"
)

const (
	BaccaratLogger = "BaccaratLogger"
)

func TestLogMgr(t *testing.T) {
	go func() {
		for {
			time.Sleep(time.Second)
			GetLoggerInstanceByName(BaccaratLogger).Info(time.Now().String())
			GetLoggerInstanceByName(BaccaratLogger).Error(time.Now().String())
			GetLoggerInstanceByName(BaccaratLogger).Debug(time.Now().String())
			GetLoggerInstanceByName(BaccaratLogger).Trace(time.Now().String())
			GetLoggerInstanceByName(BaccaratLogger).Warn(time.Now().String())
		}
	}()
}
