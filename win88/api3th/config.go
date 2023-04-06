package api3th

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"games.yol.com/win88/common"
	"github.com/cihub/seelog"
	"github.com/google/go-querystring/query"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	// 智能机器人接口
	APITypeAI = 0
	// 智能化运营接口
	APITypeSmart = 1
)

type BaseConfig struct {
	sync.Once
	methods         map[string]string
	urlNames        map[string]string
	seqNo           uint64 // 日志序号
	IPAddr          string
	AuthKey         string        // 接口认证秘钥
	TimeoutDuration time.Duration // 请求超时时长
	Name            string        // 游戏名称
	ApiType         int           // 接口类型
}

func (c *BaseConfig) LogName() string {
	return fmt.Sprint(c.Name, "Logger")
}

func (c *BaseConfig) Log() seelog.LoggerInterface {
	return common.GetLoggerInstanceByName(c.LogName())
}

func (c *BaseConfig) OrderNum() uint64 {
	return atomic.AddUint64(&c.seqNo, 1)
}

func (c *BaseConfig) Register(method, name, urlName string) {
	if method == "" {
		method = "POST"
	}
	if _, ok := c.urlNames[name]; ok {
		panic(fmt.Sprintf("api3th registered name:%s urlName:%s", name, urlName))
	}
	c.urlNames[name] = urlName
	c.methods[name] = method
}

func (c *BaseConfig) Do(name string, req interface{}, args ...time.Duration) (resp []byte, err error) {
	// 加载配置
	c.Once.Do(func() {
		c.IPAddr = common.CustomConfig.GetString(fmt.Sprintf("%sApi3thAddr", c.Name))
		c.AuthKey = common.CustomConfig.GetString(fmt.Sprintf("%sApiKey", c.Name))
		c.TimeoutDuration = time.Duration(common.CustomConfig.GetInt(fmt.Sprintf("%sApi3thTimeout", c.Name))) * time.Second
	})

	timeout := time.Duration(-1)
	if len(args) > 0 {
		timeout = args[0]
	}

	url, ok := c.urlNames[name]
	if !ok {
		return nil, errors.New(fmt.Sprintf("api3th no register %s", name))
	}

	var code int
	method := strings.ToUpper(c.methods[name])
	switch method {
	case "POST":
		code, resp, err = c.Post(url, req, timeout)
	case "POSTFORM":
		code, resp, err = c.PostForm(url, req)
	case "POSTFORMTIMEOUT":
		code, resp, err = c.PostFormTimeOut(url, req, timeout)
	default:
		err = errors.New(fmt.Sprintf("api3th method error %s", c.methods[name]))
	}
	if code != http.StatusOK {
		err = errors.New(fmt.Sprint("error code ", code))
	}
	return
}

func (c *BaseConfig) Post(url string, req interface{}, timeout time.Duration) (int, []byte, error) {
	logger := c.Log()
	data, err := json.Marshal(req)
	if err != nil {
		logger.Error("json.Marshal() error ", err)
		return 0, nil, err
	}
	seqNo := c.OrderNum()
	logger.Tracef("%s PostRequest[%d] Url %s Param: %s", c.Name, seqNo, fmt.Sprint(c.IPAddr, url), string(data))
	r, err := http.NewRequest("POST", fmt.Sprint(c.IPAddr, url), bytes.NewBuffer(data))
	if err != nil {
		logger.Errorf("%s PostRequest[%d] http.NewRequest() error: %v", c.Name, seqNo, err)
		return 0, nil, err
	}
	r.Header.Set("Content-Type", "application/json")
	if c.AuthKey != "" {
		r.Header.Set("Ocp-Apim-Subscription-Key", c.AuthKey)
	}
	cli := http.Client{}
	if timeout >= 0 {
		cli.Timeout = timeout
	} else {
		cli.Timeout = c.TimeoutDuration
	}
	resp, err := cli.Do(r)
	if err != nil {
		logger.Errorf("%s PostRequest[%d] httpClient.Do() error: %v", c.Name, seqNo, err)
		return 0, nil, err
	}
	defer resp.Body.Close()
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Errorf("%s PostRequest[%d] ioutil.ReadAll() error: %v", c.Name, seqNo, err)
		return 0, nil, err
	}

	logger.Tracef("%s PostRequest[%d] StatusCode=%d Body=%s", c.Name, seqNo, resp.StatusCode, string(respBytes))
	return resp.StatusCode, respBytes, nil
}

func (c *BaseConfig) PostFormTimeOut(url string, req interface{}, timeout time.Duration) (int, []byte, error) {
	logger := c.Log()
	data, err := json.Marshal(req)
	if err != nil {
		logger.Error("json.Marshal() error ", err)
		return 0, nil, err
	}
	seqNo := c.OrderNum()
	logger.Tracef("%s PostRequest[%d] Url %s Param: %s", c.Name, seqNo, fmt.Sprint(c.IPAddr, url), string(data))

	params, query_err := query.Values(req)
	if query_err != nil {
		logger.Errorf("(c *BaseConfig) PostForm query.Values %v", query_err)
		return 0, nil, query_err
	}
	rp := strings.NewReader(params.Encode())

	r, err := http.NewRequest("POST", fmt.Sprint(c.IPAddr, url), rp)
	if err != nil {
		logger.Errorf("%s PostRequest[%d] http.NewRequest() error: %v", c.Name, seqNo, err)
		return 0, nil, err
	}
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if c.AuthKey != "" {
		r.Header.Set("Ocp-Apim-Subscription-Key", c.AuthKey)
	}
	cli := http.Client{}
	if timeout >= 0 {
		cli.Timeout = timeout
	} else {
		cli.Timeout = c.TimeoutDuration
	}
	resp, err := cli.Do(r)
	if err != nil {
		//logger.Errorf("%s PostRequest[%d] httpClient.Do() error: %v", c.Name, seqNo, err)
		return 0, nil, err
	}
	defer resp.Body.Close()
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Errorf("%s PostRequest[%d] ioutil.ReadAll() error: %v", c.Name, seqNo, err)
		return 0, nil, err
	}

	logger.Tracef("%s PostRequest[%d] StatusCode=%d Body=%s", c.Name, seqNo, resp.StatusCode, string(respBytes))
	return resp.StatusCode, respBytes, nil
}

func (c *BaseConfig) PostForm(url1 string, req interface{}) (int, []byte, error) {
	logger := c.Log()

	data, err := json.Marshal(req)
	if err != nil {
		logger.Error("json.Marshal() error ", err)
		return 0, nil, err
	}
	c.Log().Info(fmt.Sprint(c.IPAddr, url1), string(data))

	params, query_err := query.Values(req)
	if query_err != nil {
		logger.Errorf("(c *BaseConfig) PostForm query.Values %v", query_err)
		return 0, nil, query_err
	}
	//c.Log().Info(fmt.Sprint(c.IPAddr, url1),params.Encode())
	//fmt.Println(vals.Encode())
	//
	//params := req.(url.Values)
	res, err := http.PostForm(fmt.Sprint(c.IPAddr, url1), params)
	if err != nil {
		logger.Errorf("(c *BaseConfig) PostForm http.PostForm %v", err)
		return 0, nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logger.Errorf("(c *BaseConfig) PostForm ioutil.ReadAll %v", err)
		return 0, nil, err
	}

	//println(url, " ret:", string(body[:]))

	//result := make(map[string]interface{})
	//json.Unmarshal(body, &result)

	return http.StatusOK, body, err
}

func (c *BaseConfig) Switch() bool {
	switch c.ApiType {
	case 1: // 智能化运营
		return common.CustomConfig.GetBool(fmt.Sprintf("Use%sSmartApi3th", c.Name))
	default: // ai
		return common.CustomConfig.GetBool(fmt.Sprintf("Use%sRobotApi3th", c.Name))
	}
}

func NewBaseConfig(name string, apiType int) *BaseConfig {
	ret := new(BaseConfig)
	ret.methods = make(map[string]string)
	ret.urlNames = make(map[string]string)
	ret.Name = name
	ret.ApiType = apiType
	return ret
}
