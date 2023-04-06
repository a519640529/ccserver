package api3th

import (
	"fmt"
	"github.com/cihub/seelog"
)

type TestLog struct {
	seelog.LoggerInterface
}

func (this *TestLog) Tracef(format string, params ...interface{}) {
	fmt.Printf(format+"\n", params...)
}
func (this *TestLog) Debugf(format string, params ...interface{}) {
	fmt.Printf(format+"\n", params...)
}
func (this *TestLog) Infof(format string, params ...interface{}) {
	fmt.Printf(format+"\n", params...)
}
func (this *TestLog) Warnf(format string, params ...interface{}) error {
	fmt.Printf(format+"\n", params...)
	return nil
}
func (this *TestLog) Errorf(format string, params ...interface{}) error {
	fmt.Printf(format+"\n", params...)
	return nil
}
func (this *TestLog) Criticalf(format string, params ...interface{}) error {
	fmt.Printf(format+"\n", params...)
	return nil
}
func (this *TestLog) Trace(v ...interface{}) {
	fmt.Println(v...)
}
func (this *TestLog) Debug(v ...interface{}) {
	fmt.Println(v...)
}
func (this *TestLog) Info(v ...interface{}) {
	fmt.Println(v...)
}
func (this *TestLog) Warn(v ...interface{}) error {
	fmt.Println(v...)
	return nil
}
func (this *TestLog) Error(v ...interface{}) error {
	fmt.Println(v...)
	return nil
}
func (this *TestLog) Critical(v ...interface{}) error {
	fmt.Println(v...)
	return nil
}
func (this *TestLog) Close() {
}
func (this *TestLog) Flush() {

}
func (this *TestLog) Closed() bool {
	return true
}
func (this *TestLog) SetAdditionalStackDepth(depth int) error {
	return nil
}
