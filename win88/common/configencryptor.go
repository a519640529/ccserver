package common

const (
	aa           uint = 0x7E
	bb                = 0x33
	cc                = 0xA1
	ENCRYPT_KEY1      = 0xa61fce5e // A = 0x20, B = 0xFD, C = 0x07, first = 0x1F, key = a61fce5e
	ENCRYPT_KEY2      = 0x443ffc04 // A = 0x7A, B = 0xCF, C = 0xE5, first = 0x3F, key = 443ffc04
	ENCRYPT_KEY3      = 0x12345678
)

var ConfigFE = &ConfigFileEncryptor{}

type ConfigFileEncryptor struct {
	m_nPos1          int
	m_nPos2          int
	m_nPos3          int
	m_cGlobalEncrypt EncryptCode
}
type EncryptCode struct {
	m_bufEncrypt1 [256]uint8
	m_bufEncrypt2 [256]uint8
	m_bufEncrypt3 [256]uint8
}

func (this *EncryptCode) init(key1, key2, key3 uint) {
	var a1, b1, c1, fst1 uint
	a1 = ((key1 >> 0) & 0xFF) ^ aa
	b1 = ((key1 >> 8) & 0xFF) ^ bb
	c1 = ((key1 >> 24) & 0xFF) ^ cc
	fst1 = (key1 >> 16) & 0xFF

	var a2, b2, c2, fst2 uint
	a2 = ((key2 >> 0) & 0xFF) ^ aa
	b2 = ((key2 >> 8) & 0xFF) ^ bb
	c2 = ((key2 >> 24) & 0xFF) ^ cc
	fst2 = (key2 >> 16) & 0xFF

	i := 0
	nCode := uint8(fst1)
	for i = 0; i < 256; i++ {
		this.m_bufEncrypt1[i] = nCode
		nCode = (uint8(a1)*nCode*nCode + uint8(b1)*nCode + uint8(c1)) & 0xFF
	}

	nCode = uint8(fst2)
	for i = 0; i < 256; i++ {
		this.m_bufEncrypt2[i] = nCode
		nCode = (uint8(a2)*nCode*nCode + uint8(b2)*nCode + uint8(c2)) & 0xFF
	}

	for i = 0; i < 256; i++ {
		this.m_bufEncrypt3[i] = uint8(((key3 >> uint(i%4)) ^ uint(i)) & 0xff)
	}
}
func (this *ConfigFileEncryptor) init(key1, key2, key3 uint) {
	this.m_cGlobalEncrypt.init(key1, key2, key3)
}
func (this *ConfigFileEncryptor) IsCipherText(buf []byte) bool {
	size := len(buf)
	if size < 4 {
		return false
	}
	//0x1b454e43
	if buf[size-1] == 0x43 && buf[size-2] == 0x4e && buf[size-3] == 0x45 && buf[size-4] == 0x1b {
		return true
	}
	return false
}
func (this *ConfigFileEncryptor) Encrypt(buf []byte) []byte {
	size := len(buf)
	oldPos1, oldPos2, oldPos3 := this.m_nPos1, this.m_nPos2, this.m_nPos3
	for i := 0; i < size; i++ {
		buf[i] ^= this.m_cGlobalEncrypt.m_bufEncrypt1[this.m_nPos1]
		buf[i] ^= this.m_cGlobalEncrypt.m_bufEncrypt2[this.m_nPos2]
		buf[i] ^= this.m_cGlobalEncrypt.m_bufEncrypt3[this.m_nPos3]
		this.m_nPos1++
		if this.m_nPos1 >= 256 {
			this.m_nPos1 = 0
			this.m_nPos2++
			if this.m_nPos2 >= 256 {
				this.m_nPos2 = 0
			}
		}
		this.m_nPos3++
		if this.m_nPos3 >= 256 {
			this.m_nPos3 = 0
		}
	}
	this.m_nPos1, this.m_nPos2, this.m_nPos3 = oldPos1, oldPos2, oldPos3
	buf = append(buf, 0x1b, 0x45, 0x4e, 0x43)
	return buf
}
func (this *ConfigFileEncryptor) Decrtypt(buf []byte) []byte {
	size := len(buf) - 4
	oldPos1, oldPos2, oldPos3 := this.m_nPos1, this.m_nPos2, this.m_nPos3
	for i := 0; i < size; i++ {
		buf[i] ^= this.m_cGlobalEncrypt.m_bufEncrypt1[this.m_nPos1]
		buf[i] ^= this.m_cGlobalEncrypt.m_bufEncrypt2[this.m_nPos2]
		buf[i] ^= this.m_cGlobalEncrypt.m_bufEncrypt3[this.m_nPos3]
		this.m_nPos1++
		if this.m_nPos1 >= 256 {
			this.m_nPos1 = 0
			this.m_nPos2++
			if this.m_nPos2 >= 256 {
				this.m_nPos2 = 0
			}
		}
		this.m_nPos3++
		if this.m_nPos3 >= 256 {
			this.m_nPos3 = 0
		}
	}
	this.m_nPos1, this.m_nPos2, this.m_nPos3 = oldPos1, oldPos2, oldPos3
	return buf[:size]
}
func init() {
	ConfigFE.init(ENCRYPT_KEY1, ENCRYPT_KEY2, ENCRYPT_KEY3)
}
