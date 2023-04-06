// 生成签名用
package webapi

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"io"
	"net/url"
	"sort"
	"strconv"
	"time"
)

// 生成签名
func MakeSig(appId, action string, params map[string]string, data []byte, ts int) string {
	h := md5.New()
	io.WriteString(h, action)
	io.WriteString(h, strconv.Itoa(ts))
	io.WriteString(h, appId)
	if len(params) > 0 {
		vals := []string{}
		for _, v := range params {
			vals = append(vals, v)
		}
		//sort
		sort.Strings(vals)

		cnt := len(vals)
		for i := 0; i < cnt; i++ {
			io.WriteString(h, vals[i])
		}
	}
	if len(data) > 0 {
		h.Write(data)
	}
	//println()
	return hex.EncodeToString(h.Sum(nil))
}

// 生成请求串
func MakeURL(appId, action string, params map[string]string, data []byte) string {
	var buf bytes.Buffer
	buf.WriteString(Config.GameApiURL)
	buf.WriteString(action)
	buf.WriteString("?")
	if len(params) > 0 {
		for k, v := range params {
			buf.WriteString(k)
			buf.WriteString("=")
			buf.WriteString(url.QueryEscape(v))
			buf.WriteString("&")
		}
	}
	ts := time.Now().Nanosecond()
	buf.WriteString("nano=" + strconv.Itoa(ts) + "&")
	buf.WriteString("sign=")
	buf.WriteString(url.QueryEscape(MakeSig(appId, action, params, data, ts)))
	return buf.String()
}
func MakeMd5String(strs ...string) string {
	buff := md5.New()
	for _, value := range strs {
		io.WriteString(buff, value)
	}
	return hex.EncodeToString(buff.Sum(nil))
}
