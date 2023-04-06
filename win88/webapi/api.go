package webapi

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"google.golang.org/protobuf/proto"
)

const DEFAULT_TIMEOUT = time.Duration(time.Second * 30)

var WebApiStats = new(sync.Map)

type ApiStats struct {
	RunTimes        int64 //执行次数
	TotalRuningTime int64 //总执行时间
	MaxRuningTime   int64 //最长执行时间
	TimeoutTimes    int64 //执行超时次数
}

func DeviceOs(os string) string {
	switch os {
	case "ios":
		return "2"
	case "android":
		return "1"
	default:
		return "0"
	}
}

// API调用
// action 为需要调用的api, 例如 /api/Send/SendCode
// params 为需发送的参数
// protocol = https || http ||Post
func API_OP(url string, params url.Values) (map[string]interface{}, error) {
	//res, err := http.PostForm("http://192.168.1.160:9090/api/Sms/SendCaptcha", params)
	res, err := http.PostForm(Config.GameApiURL+url, params)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	//println(url, " ret:", string(body[:]))

	result := make(map[string]interface{})
	json.Unmarshal(body, &result)

	return result, nil
}

// ///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// GET方式请求
func getRequest(appId, action string, params map[string]string, body proto.Message, protocol string, dura time.Duration) ([]byte, error) {
	var client *http.Client
	if strings.ToUpper(protocol) == "HTTPS" {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client = &http.Client{Transport: tr}
	} else {
		client = &http.Client{}
	}
	data, err := proto.Marshal(body)
	if err != nil {
		return nil, err
	}

	callUrl := MakeURL(appId, action, params, data)
	//println("callurl=", callUrl)
	req, err := http.NewRequest("GET", callUrl, nil)
	if err != nil {
		return nil, err
	}

	var stats *ApiStats
	if v, exist := WebApiStats.Load(action); exist {
		stats = v.(*ApiStats)
	} else {
		stats = &ApiStats{}
		WebApiStats.Store(action, stats)
	}

	var isTimeout bool
	start := time.Now()
	defer func() {
		ps := int64(time.Now().Sub(start) / time.Millisecond)
		if stats != nil {
			if isTimeout {
				atomic.AddInt64(&stats.TimeoutTimes, 1)
			}
			atomic.AddInt64(&stats.RunTimes, 1)
			atomic.AddInt64(&stats.TotalRuningTime, ps)
			if atomic.LoadInt64(&stats.MaxRuningTime) < ps {
				atomic.StoreInt64(&stats.MaxRuningTime, ps)
			}
		}
	}()

	//设置超时
	client.Timeout = dura
	req.Close = true
	resp, err := client.Do(req)
	if err != nil {
		if uerr, ok := err.(net.Error); ok {
			isTimeout = uerr.Timeout()
		}
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return body, err
	}
	println("callurl=", callUrl, "code=", resp.StatusCode)
	return nil, fmt.Errorf("StatusCode:%d", resp.StatusCode)
}

// ///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// POST方式请求
func postRequest(appId, action string, params map[string]string, body proto.Message, protocol string, dura time.Duration) ([]byte, error) {
	var client *http.Client
	if strings.ToUpper(protocol) == "HTTPS" {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client = &http.Client{Transport: tr}
	} else {
		client = &http.Client{}
	}

	data, err := proto.Marshal(body)
	if err != nil {
		return nil, err
	}

	callUrl := MakeURL(appId, action, params, data)
	//println("callurl=", callUrl)
	req, err := http.NewRequest("POST", callUrl, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	var stats *ApiStats
	if v, exist := WebApiStats.Load(action); exist {
		stats = v.(*ApiStats)
	} else {
		stats = &ApiStats{}
		WebApiStats.Store(action, stats)
	}

	var isTimeout bool
	start := time.Now()
	defer func() {
		ps := int64(time.Now().Sub(start) / time.Millisecond)
		if stats != nil {
			if isTimeout {
				atomic.AddInt64(&stats.TimeoutTimes, 1)
			}
			atomic.AddInt64(&stats.RunTimes, 1)
			atomic.AddInt64(&stats.TotalRuningTime, ps)
			if atomic.LoadInt64(&stats.MaxRuningTime) < ps {
				atomic.StoreInt64(&stats.MaxRuningTime, ps)
			}
		}
	}()

	//设置超时
	client.Timeout = dura
	req.Close = true
	resp, err := client.Do(req)
	if err != nil {
		if uerr, ok := err.(net.Error); ok {
			isTimeout = uerr.Timeout()
		}
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return body, err
	}
	println("callurl=", callUrl, "code=", resp.StatusCode)
	return nil, fmt.Errorf("StatusCode:%d", resp.StatusCode)
}

func Stats() map[string]ApiStats {
	stats := make(map[string]ApiStats)
	WebApiStats.Range(func(k, v interface{}) bool {
		if s, ok := v.(*ApiStats); ok {
			ss := *s //计数可能不精准
			stats[k.(string)] = ss
		}
		return true
	})
	return stats
}
