package main

import (
	"encoding/json"
	"games.yol.com/win88/common"
	"games.yol.com/win88/model"
	"games.yol.com/win88/webapi"
	"github.com/idealeak/goserver/core/logger"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

var LogicLevelMgrSington = &LogicLevelMgr{
	config: make(map[string]*LogicLevelConfig),
	client: &http.Client{Timeout: 30 * time.Second},
}

type LogicLevelMgr struct {
	config map[string]*LogicLevelConfig
	client *http.Client
}
type LogicLevelConfig struct {
	Platform       string
	LogicLevelInfo map[int32]*LogicLevelInfo
}
type LogicLevelInfo struct {
	Id          int32    //分层id
	ClusterName string   //分层名称
	StartAct    int32    //分层开关 1开启 0关闭
	CheckActIds []int32  //分层包含的活动id
	CheckPay    []string //分层包含的充值类型
}

func (this *LogicLevelMgr) GetConfig(platform string) *LogicLevelConfig {
	return this.config[platform]
}

func (this *LogicLevelMgr) UpdateConfig(cfg *LogicLevelConfig) {
	logger.Logger.Trace("++++++++++++++UpdateConfig++++++++++++++")
	this.config[cfg.Platform] = cfg

	if playersOL, ok := PlayerMgrSington.playerOfPlatform[cfg.Platform]; ok {
		for _, player := range playersOL {
			if player != nil && !player.IsRob {
				player.layered = make(map[int]bool)
				for _, v := range player.layerlevels {
					if td, ok := cfg.LogicLevelInfo[int32(v)]; ok {
						if td.StartAct == 1 {
							for _, id := range td.CheckActIds {
								player.layered[int(id)] = true
							}
						}
					}
				}
				//player.ActStateSend2Client()
			}
		}
	}
}

type NewMsg struct {
	Platform string
	SnId     int
	Levels   []int
}

func (this *LogicLevelMgr) SendPostBySnIds(platform string, snids []int32) []NewMsg {
	client := this.client
	form := make(url.Values)
	form.Set("Platform", platform)
	str := ""
	for k, snid := range snids {
		str += strconv.Itoa(int(snid))
		if k+1 < len(snids) {
			str += ","
		}
	}
	form.Set("SnIds", str)
	form.Set("PltName", common.CustomConfig.GetString("PltName"))
	logicLevelUrl := common.CustomConfig.GetString("LogicLevelUrl")
	resp, err := client.PostForm(logicLevelUrl+"/QueryDataBySnIds", form)
	if resp != nil && resp.Status == "200 OK" && err == nil {
		defer resp.Body.Close()
		body, _ := ioutil.ReadAll(resp.Body)
		logger.Logger.Trace(string(body))
		var data []NewMsg
		json.Unmarshal(body, &data)
		return data
	}
	return nil
}
func (this *LogicLevelMgr) LoadConfig() {
	logger.Logger.Trace("++++++++++++++LoadConfig++++++++++++++")
	type LogicLevelConfigData struct {
		Tag int
		Msg []*LogicLevelConfig
	}
	if !model.GameParamData.UseEtcd {
		logger.Logger.Trace("API_GetGradeShopConfigData")
		buff, err := webapi.API_GetLogicLevelConfigData(common.GetAppId())
		if err == nil {
			var data LogicLevelConfigData
			err = json.Unmarshal(buff, &data)
			if err == nil && data.Tag == 0 {
				for _, cfg := range data.Msg {
					this.UpdateConfig(cfg)
				}
			} else {
				logger.Logger.Error("Unmarshal LogicLevelConfigData config data error:", err, string(buff))
			}
		} else {
			logger.Logger.Error("Get LogicLevelConfigData config data error:", err)
		}
	} else {
		EtcdMgrSington.InitLogicLevelConfig()
	}
}
