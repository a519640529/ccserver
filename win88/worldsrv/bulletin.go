package main

import (
	"encoding/json"
	"games.yol.com/win88/common"
	"games.yol.com/win88/model"
	"games.yol.com/win88/webapi"
	"github.com/idealeak/goserver/core/logger"
	"github.com/idealeak/goserver/core/module"
	"time"
)

//公告模块

var BulletMgrSington = &BulletMgr{
	BulletMsgList: make(map[int32]*Bullet),
}

type BulletMgr struct {
	BulletMsgList map[int32]*Bullet
}

type Bullet struct {
	Id            int32
	Sort          int32 //排序
	Platform      string
	NoticeTitle   string
	NoticeContent string
	UpdateTime    string
	State         int //0 关闭  1开启
}
type ApiBulletResult struct {
	Tag int
	Msg []Bullet
}

func (this *BulletMgr) query() {
	//不使用etcd的情况下走api获取
	if model.GameParamData.UseEtcd {
		EtcdMgrSington.InitPlatformBulletin()
	} else {
		buff, err := webapi.API_GetBulletData(common.GetAppId())
		//logger.Logger.Warn("bulletin buff: ", string(buff))
		if err == nil {
			info := ApiBulletResult{}
			err = json.Unmarshal([]byte(buff), &info)
			if err == nil {
				for i := 0; i < len(info.Msg); i++ {
					BulletMgrSington.BulletMsgList[info.Msg[i].Id] = &info.Msg[i]
				}
			} else {
				logger.Logger.Error("Unmarshal Bullet data error:", err, string(buff))
			}
		} else {
			logger.Logger.Error("Get Bullet data error:", err)
		}
	}
}

func (this *BulletMgr) clearPlatformBullet(Platform string) {
	for k, v := range this.BulletMsgList {
		if v.Platform == Platform {
			delete(this.BulletMsgList, k)
		}
	}
}

func (this *BulletMgr) updateBullet(id int32, info string) (map[int32]*Bullet, string) {
	platform := ""
	if info == "" {
		delete(this.BulletMsgList, id)
		platform = "delete"
	} else {
		bt := this.Unmarshal(info)
		if bt != nil {
			this.BulletMsgList[id] = bt
			platform = bt.Platform
		}
	}
	return this.BulletMsgList, platform
}
func (this *BulletMgr) Unmarshal(info string) (bt *Bullet) {
	err := json.Unmarshal([]byte(info), &bt)
	if err != nil {
		logger.Logger.Trace("Unmarshal Bullet is error :", err)
		return nil
	}
	return
}
func (this *BulletMgr) ModuleName() string {
	return "BulletMgr"
}
func (this *BulletMgr) Init() {
}
func (this *BulletMgr) Update() {
}
func (this *BulletMgr) Shutdown() {
	module.UnregisteModule(this)
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////
/////////招商列表
///////////////////////////////////////////////////////////////////////////////////////////////////////////

var CustomerMgrSington = &CustomerMgr{
	CustomerMsgList: make(map[int32]*Customer),
}

type CustomerMgr struct {
	CustomerMsgList map[int32]*Customer
}
type Customer struct {
	Id             int32
	Platform       string
	Weixin_account string
	Qq_account     string
	Headurl        string
	Nickname       string
	Status         int
	Ext            string
}
type ApiCustomerResult struct {
	Tag int
	Msg []Customer
}

func (this *CustomerMgr) ModuleName() string {
	return "CustomerMgr"
}
func (this *CustomerMgr) Init() {
}
func (this *CustomerMgr) Update() {
}
func (this *CustomerMgr) Shutdown() {
	module.UnregisteModule(this)
}
func (this *CustomerMgr) query() {
	//不使用etcd的情况下走api获取
	if model.GameParamData.UseEtcd {
		EtcdMgrSington.InitPlatformAgent()
	} else {
		buff, err := webapi.API_GetCustomerData(common.GetAppId())
		logger.Logger.Trace("customer buff:", string(buff))
		if err == nil {
			c_info := ApiCustomerResult{}
			err = json.Unmarshal([]byte(buff), &c_info)
			if err == nil {
				for i := 0; i < len(c_info.Msg); i++ {
					CustomerMgrSington.CustomerMsgList[c_info.Msg[i].Id] = &c_info.Msg[i]
				}
			} else {
				logger.Logger.Trace("CustomerMgr is Unmarshal error.", err)
			}
		} else {
			logger.Logger.Trace("API_GetCustomerData is error. ", err)
		}
	}
}
func (this *CustomerMgr) updateCustomer(id int32, info string) (map[int32]*Customer, string) {
	platform := ""
	if info == "" {
		delete(this.CustomerMsgList, id)
		platform = "delete"
	} else {
		bt := this.Unmarshal(info)
		if bt != nil {
			this.CustomerMsgList[id] = bt
			platform = bt.Platform
		}
	}
	return this.CustomerMsgList, platform
}
func (this *CustomerMgr) Unmarshal(info string) (bt *Customer) {
	err := json.Unmarshal([]byte(info), &bt)
	if err != nil {
		logger.Logger.Trace("Unmarshal Customer is error :", err)
		return nil
	}
	return
}

func init() {
	module.RegisteModule(BulletMgrSington, time.Second*2, 0)
	module.RegisteModule(CustomerMgrSington, time.Second*2, 0)

	RegisteParallelLoadFunc("平台公告", func() error {
		BulletMgrSington.query()
		return nil
	})

	////不使用并发加载，因为并发太快，用到GameParam里面的数据还未来的及加载，下同
	//core.RegisteHook(core.HOOK_BEFORE_START, func() error {
	//	BulletMgrSington.query()
	//	return nil
	//})

	RegisteParallelLoadFunc("平台代理", func() error {
		CustomerMgrSington.query()
		return nil
	})

	//core.RegisteHook(core.HOOK_BEFORE_START, func() error {
	//	CustomerMgrSington.query()
	//	return nil
	//})
}
