package common

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
	"github.com/idealeak/goserver.v3/core/logger"
	"regexp"
	"strconv"
)

var key = []byte("kjh-vgjhhionoommmkokmokoo$%JH")

// 加密
func EnCrypt(orig []byte) (str string) {
	defer func() {
		err := recover()
		if err != nil {
			logger.Logger.Errorf("EnCrypt %v Error %v", string(orig), err)
			str = string(orig)
		}
	}()
	//将秘钥中的每个字节累加,通过sum实现orig的加密工作
	sum := 0
	for i := 0; i < len(key); i++ {
		sum += int(key[0])
	}

	//给明文补码
	var pkcs_code = PKCS5Padding(orig, 8)

	//通过秘钥，对补码后的明文进行加密
	for j := 0; j < len(pkcs_code); j++ {
		pkcs_code[j] += byte(sum)
	}
	//base64.URLEncoding.EncodeToString()
	return base64.URLEncoding.EncodeToString(pkcs_code)
}

// 补码
func PKCS5Padding(orig []byte, size int) []byte {
	//计算明文的长度
	length := len(orig)
	padding := size - length%size
	//向byte类型的数组中重复添加padding
	repeats := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(orig, repeats...)
}

// 解密
func DeCrypt(text string) (str string) {
	defer func() {
		err := recover()
		if err != nil {
			logger.Logger.Errorf("DeCrypt %v Error %v", text, err)
			str = text
		}
	}()
	//orig, err := base64.StdEncoding.DecodeString(text)
	orig, err := base64.URLEncoding.DecodeString(text)
	if err != nil {
		return "密文类型错误"
	}
	sum := 0
	for i := 0; i < len(key); i++ {
		sum += int(key[0])
	}

	//解密
	for j := 0; j < len(orig); j++ {
		orig[j] -= byte(sum)
	}

	//去码
	var pkcs_unCode = PKCS5UnPadding(orig)
	return string(pkcs_unCode)
}

// 去码
func PKCS5UnPadding(orig []byte) []byte {
	//获取最后一位补码的数字
	var tail = int(orig[len(orig)-1])
	return orig[:(len(orig) - tail)]
}

var aesRule, _ = regexp.Compile(`^[0-9]+$`)

const (
	aeskey = "DoNotEditThisKeyDoNotEditThisKey" // 加密的密钥,绝不可以更改
)

// 下面的字符串,也绝不可以更改
var defaultLetters = []rune("idjGfiRogsFnkdKgokdfgdow07u6978uxcvvLiPiDfjafOd2fuFJYYGBJuykbvfk")

func AesEncrypt(origDataStr string) (str string) {
	defer func() {
		err := recover()
		if err != nil {
			logger.Logger.Errorf("AesEncrypt %v Error %v", origDataStr, err)
			str = origDataStr
		}
	}()
	strlen := len(origDataStr)
	b := aesRule.MatchString(origDataStr)
	//不全是数字，或长度为零，不加密
	if !b || strlen == 0 {
		return origDataStr
	}
	phonenum, errint := strconv.Atoi(origDataStr)
	if errint != nil {
		return origDataStr
	}

	text := []byte(origDataStr)
	//指定加密、解密算法为AES，返回一个AES的Block接口对象
	block, err := aes.NewCipher([]byte(aeskey))
	if err != nil {
		panic(err)
	}
	//指定计数器,长度必须等于block的块尺寸
	iv := string(defaultLetters[phonenum%(len(defaultLetters))])
	count := []byte(fmt.Sprintf("%016v", iv))
	//指定分组模式
	blockMode := cipher.NewCTR(block, count)
	//执行加密、解密操作
	message := make([]byte, len(text))
	blockMode.XORKeyStream(message, text)
	//返回明文或密文
	return fmt.Sprintf("%v%v", iv, base64.StdEncoding.EncodeToString(message))
	//return base64.StdEncoding.EncodeToString(message)
}

func AesDecrypt(cryptedstr string) (str string) {
	defer func() {
		err := recover()
		if err != nil {
			logger.Logger.Errorf("AesDecrypt %v Error %v", cryptedstr, err)
			str = cryptedstr
		}
	}()
	strlen := len(cryptedstr)
	b := aesRule.MatchString(cryptedstr)
	//全是数字，或长度为零，不解密
	if b || strlen == 0 {
		return cryptedstr
	}

	iv := cryptedstr[:1]
	str = cryptedstr[1:]
	text, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		logger.Logger.Errorf("AesDecrypt %v  Err:%v", cryptedstr, err)
		return cryptedstr
	}

	//指定加密、解密算法为AES，返回一个AES的Block接口对象
	block, err := aes.NewCipher([]byte(aeskey))
	if err != nil {
		panic(err)
	}
	//指定计数器,长度必须等于block的块尺寸
	count := []byte(fmt.Sprintf("%016v", iv))
	//指定分组模式
	blockMode := cipher.NewCTR(block, count)
	//执行加密、解密操作
	message := make([]byte, len(text))
	blockMode.XORKeyStream(message, text)
	//返回明文或密文
	return string(message)
}
