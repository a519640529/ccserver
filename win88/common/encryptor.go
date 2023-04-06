package common

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
)

type Encryptor struct {
	buf1 [256]uint8
	buf2 [256]uint8
}

func (e *Encryptor) Init(appId, key string, ts int32) {
	h1 := md5.New()
	io.WriteString(h1, fmt.Sprintf("%v;%v;%v", appId, key, ts))
	raw1 := hex.EncodeToString(h1.Sum(nil))
	n1 := len(raw1)
	for i := 0; i < 256; i++ {
		e.buf1[i] = uint8((raw1[i%n1] ^ uint8(i)) & 0xff)
	}

	h2 := md5.New()
	io.WriteString(h2, key)
	raw2 := hex.EncodeToString(h2.Sum(nil))
	n2 := len(raw2)
	for i := 0; i < 256; i++ {
		e.buf2[i] = uint8((raw2[i%n2] ^ uint8(i)) & 0xff)
	}
}

func (e *Encryptor) Encrypt(buf []byte, size int) {
	var pos1, pos2 int
	for i := 0; i < size; i++ {
		buf[i] ^= e.buf1[pos1]
		buf[i] ^= e.buf2[pos2]
		pos1++
		if pos1 == 255 {
			pos1 = 0
			pos2++
			if pos2 == 255 {
				pos2 = 0
			}
		}
	}
}
