package common

import (
	"bytes"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"io/ioutil"
	"os"
)

const (
	MDUMP_MAGIC_CODE      uint32 = 0xdeadfee1
	MDUMP_FILEHEAD_LEN           = 16
	MDUMP_VERSION_LASTEST        = 1
)

var (
	errMdumpEmpty      = errors.New(".mdump empty")
	errMdumpFormat     = errors.New(".mdump format error")
	errMdumpCheckSum   = errors.New(".mdump checksum error")
	errMdumpTruncation = errors.New(".mdump truncation error")
)

// 对象序列化接口
type Serializable interface {
	//序列化
	Marshal() ([]byte, error)
	//反序列化
	Unmarshal([]byte, interface{}) error
}

// 内存dump文件头
type MDumpFileHead struct {
	MagicCode uint32 //mdump文件标记
	Version   uint32 //版本号
	DataLen   uint32 //数据区长度
	CheckSum  uint32 //校验和
}

func ReadMDumpFile(fileName string) (head *MDumpFileHead, data []byte, err error) {
	buf, err := ioutil.ReadFile(fileName)
	if err != nil {
		return
	}
	os.Remove(fileName)
	//解析文件头
	head = &MDumpFileHead{}
	err = binary.Read(bytes.NewReader(buf), binary.BigEndian, head)
	if err != nil {
		return
	}

	if head.MagicCode != MDUMP_MAGIC_CODE {
		err = errMdumpFormat
		return
	}

	data = buf[MDUMP_FILEHEAD_LEN:]
	if uint32(len(data)) != head.DataLen {
		err = errMdumpTruncation
		return
	}

	checkSum := crc32.ChecksumIEEE(data)
	if checkSum != head.CheckSum {
		err = errMdumpCheckSum
		return
	}

	return
}

func WriteMDumpFile(fileName string, data []byte) error {
	if len(data) == 0 {
		return errMdumpEmpty
	}

	checkSum := crc32.ChecksumIEEE(data)
	head := MDumpFileHead{
		MagicCode: MDUMP_MAGIC_CODE,
		Version:   MDUMP_VERSION_LASTEST,
		DataLen:   uint32(len(data)),
		CheckSum:  checkSum,
	}

	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()
	//写文件头
	err = binary.Write(file, binary.BigEndian, &head)
	if err != nil {
		return err
	}
	//写数据
	_, err = file.WriteAt(data, int64(MDUMP_FILEHEAD_LEN))
	if err != nil {
		return err
	}

	return err
}
