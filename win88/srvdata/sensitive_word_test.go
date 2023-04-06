package srvdata

import (
	"path/filepath"
	"testing"
)

func initSensitiveWordEnv(t *testing.T) {
	fileName := "DB_Sensitive_Words.dat"
	loader := DataMgr.GetLoader(fileName)
	if loader != nil {
		fullPath := filepath.Join("../data", fileName)
		err := loader.load(fullPath)
		if err != nil {
			t.Error(fileName, " loader err:", err)
		}
	}
	err := initSensitiveWordTree()
	if err != nil {
		t.Error(fileName, " init tree err:", err)
	}
}

func TestSensitiveWord(t *testing.T) {
	initSensitiveWordEnv(t)

	testWords := "18岁以下勿看00000000"
	if !HasSensitiveWord([]rune(testWords)) {
		t.Error("HasSensitiveWord")
	}

	ret := ReplaceSensitiveWord([]rune(testWords))
	t.Log(string(ret[:]))
	ret = ReplaceSensitiveWord([]rune("go-vern-menthjdhfjdfhdj"))
	t.Log(string(ret[:]))
	ret = ReplaceSensitiveWord([]rune("【③肖中特】中国爱神之传奇00爱液"))
	t.Log(string(ret[:]))
}
