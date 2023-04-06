package webapi

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"games.yol.com/win88/common"
	webapi_proto "games.yol.com/win88/protocol/webapi"
	"github.com/idealeak/goserver.v3/core/logger"
	"google.golang.org/protobuf/proto"
)

const (
	XHJ_StartGameUrl = "rocket/game_srv_third/register_or_login"
	XHJ_ExitGameUrl  = "rocket/game_srv_third/login_out"
)

type XHJThridPlatform struct {
	ThirdPlatformBase
}

func (this *XHJThridPlatform) InitMappingRelation(db map[int32]*WebAPI_ThirdPlatformGameMapping) {
	this.GamefreeIdMappingInfo = make(map[int32]*WebAPI_ThirdPlatformGameMapping)
	this.ThirdgameIdMappingInfo = make(map[string]*WebAPI_ThirdPlatformGameMapping)
	for _, v := range db {
		if v.ThirdPlatformName == this.Name {
			this.GamefreeIdMappingInfo[v.GameFreeID] = v
			this.ThirdgameIdMappingInfo[v.ThirdGameID] = v
		}
	}
}
func (this *XHJThridPlatform) GamefreeId2ThirdGameInfo(gamefreeid int32) (thirdInfo *WebAPI_ThirdPlatformGameMapping) {
	if v, exist := this.GamefreeIdMappingInfo[gamefreeid]; exist {
		return v
	}
	return
}
func (this *XHJThridPlatform) ThirdGameInfo2GamefreeId(thirdInfo *WebAPI_ThirdPlatformGameMapping) (gamefreeid int32) {
	if thirdInfo != nil {
		for _, v := range this.GamefreeIdMappingInfo {
			if thirdInfo.Desc == v.Desc {
				return v.GameFreeID
			}
		}
	}
	return 0
}

func (this *XHJThridPlatform) GetPlatformBase() ThirdPlatformBase {
	return this.ThirdPlatformBase
}
func (this *XHJThridPlatform) ReqCreateAccount(Snid int32, Platform, Channel, ip string) error {

	return nil
}
func (this *XHJThridPlatform) ReqUserBalance(Snid int32, Platform, Channel, ip string) (err error, balance int64) {

	return nil, 0
}
func (this *XHJThridPlatform) ReqIsAllowTransfer(Snid int32, Platform, Channel string) bool {

	return true
}
func (this *XHJThridPlatform) ReqCheckTransferIsSuccess(Snid int32, TransferId string, Platform, Channel string) error {

	return nil
}
func (this *XHJThridPlatform) ReqTransfer(Snid int32, Amount int64, TransferId string, Platform, Channel, ip string) (e error, timeout bool) {

	return nil, false
}

func (this *XHJThridPlatform) ReqEnterGame(Snid int32, gameid string, clientIP string, Platform, Channel string, amount int64) (err error, ret_url string) {
	pack := &webapi_proto.SARocketLogin{
		Snid:     int64(Snid),
		Amount:   amount,
		Platform: Platform,
	}

	buff, err := this.postRequest(common.GetAppId(), XHJ_StartGameUrl, nil, pack, "http", DEFAULT_TIMEOUT)
	logger.Trace("XHJReqEnterGame Return:", string(buff))
	if err != nil {
		return err, ""
	}

	ar := webapi_proto.ASRocketLogin{}
	err = proto.Unmarshal(buff, &ar)
	if err == nil && ar.Url == "" {
		return errors.New(ar.Msg), ""
	}

	return err, ar.Url
}
func (this *XHJThridPlatform) ReqLeaveGame(Snid int32, gameid string, clientIP string, Platform, Channel string) (err error, amount int64) {
	pack := &webapi_proto.SARocketLoginOut{
		Snid:     int64(Snid),
		Platform: Platform,
	}

	buff, err := this.postRequest(common.GetAppId(), XHJ_ExitGameUrl, nil, pack, "http", DEFAULT_TIMEOUT)
	logger.Trace("XHJReqLeaveGame Return:", string(buff))
	if err != nil {
		return err, 0
	}

	ar := webapi_proto.ASRocketLoginOut{}
	err = proto.Unmarshal(buff, &ar)
	if err != nil {
		return errors.New(ar.Msg), 0
	}

	return err, int64(ar.Amount)
}
func (this *XHJThridPlatform) MappingGameName(snid int32) string {

	return strconv.Itoa(int(snid)) + "_XHJ"
}

func (this *XHJThridPlatform) postRequest(appId, action string, params map[string]string, body proto.Message, protocol string, dura time.Duration) ([]byte, error) {
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

	callUrl := makeURL(appId, action, params, data)
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

// 生成请求串
func makeURL(appId, action string, params map[string]string, data []byte) string {
	var buf bytes.Buffer
	buf.WriteString(ReqCgAddr)
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
