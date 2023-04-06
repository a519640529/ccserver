package common

import (
	"bytes"
	"crypto/rc4"
	"encoding/base32"
	"encoding/binary"
	"math/rand"
)

const encodeHex = "23456789ABCDEFGHJKLMNPQRSTUVWXYZ"

var rc4Key = []byte{0x12, 0x57, 0xb8, 0xd8, 0x60, 0xae, 0x4c, 0xbd}

var HexEncoding = base32.NewEncoding(encodeHex)
var Rc4Cipher, _ = rc4.NewCipher(rc4Key)

// 生成兑换码
func GeneExchangeCode(seq int32) string {
	buf := bytes.NewBuffer(nil)
	binary.Write(buf, binary.BigEndian, seq)
	buf.WriteByte(byte(rand.Intn(0xff)))
	var dst [5]byte
	Rc4Cipher.XORKeyStream(dst[:], buf.Bytes())
	return HexEncoding.EncodeToString(dst[:])
}
