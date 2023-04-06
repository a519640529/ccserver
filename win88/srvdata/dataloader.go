package srvdata

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"io/ioutil"
)

var fileSignMap = make(map[string]string)

type DataLoader interface {
	load(fileFullPath string) error
	reload(fileFullPath string) error
}

type DataHolder interface {
	unmarshal([]byte) error
	reunmarshal([]byte) error
}

type JsonDataLoader struct {
	dh DataHolder
}

func (this *JsonDataLoader) load(fileFullPath string) error {
	buf, err := ioutil.ReadFile(fileFullPath)
	if err != nil {
		return err
	}

	h := md5.New()
	_, err = h.Write(buf)
	if err != nil {
		return err
	}
	fileSign := hex.EncodeToString(h.Sum(nil))
	fileSignMap[fileFullPath] = fileSign

	if this.dh != nil {
		err = this.dh.unmarshal(buf)
	}

	return err
}

func (this *JsonDataLoader) reload(fileFullPath string) error {
	buf, err := ioutil.ReadFile(fileFullPath)
	if err != nil {
		return err
	}

	h := md5.New()
	_, err = h.Write(buf)
	if err != nil {
		return err
	}
	fileSign := hex.EncodeToString(h.Sum(nil))
	if preSign, exist := fileSignMap[fileFullPath]; exist {
		if preSign == fileSign {
			return nil
		}
	}
	fileSignMap[fileFullPath] = fileSign

	if this.dh != nil {
		err = this.dh.reunmarshal(buf)
	}

	return err
}

type ProtobufDataLoader struct {
	dh DataHolder
}

func (this *ProtobufDataLoader) load(fileFullPath string) error {
	buf, err := ioutil.ReadFile(fileFullPath)
	if err != nil {
		return err
	}

	h := md5.New()
	_, err = h.Write(buf)
	if err != nil {
		return err
	}
	fileSign := hex.EncodeToString(h.Sum(nil))
	fileSignMap[fileFullPath] = fileSign

	if this.dh != nil {
		err = this.dh.unmarshal(buf)
	}

	return err
}

func (this *ProtobufDataLoader) reload(fileFullPath string) error {
	buf, err := ioutil.ReadFile(fileFullPath)
	if err != nil {
		return err
	}

	h := md5.New()
	_, err = h.Write(buf)
	if err != nil {
		return err
	}
	fileSign := hex.EncodeToString(h.Sum(nil))
	if preSign, exist := fileSignMap[fileFullPath]; exist {
		if preSign == fileSign {
			return errors.New("content no modify")
		}
	}
	fileSignMap[fileFullPath] = fileSign

	if this.dh != nil {
		err = this.dh.reunmarshal(buf)
	}

	return err
}
