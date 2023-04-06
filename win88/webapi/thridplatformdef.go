package webapi

import (
	"errors"

	"games.yol.com/win88/common"
)

var ReqCgAddr string

// 没有单独创建账号的接口
var ErrNoCreated = errors.New("no create account")

type ThirdError struct {
	err     error
	isColse bool
}

func (e ThirdError) Error() string {
	return e.err.Error()
}

func (e ThirdError) IsClose() bool {
	return e.isColse
}

const (
	ThridAccountStatus_OK        int32 = 200   //三方账户状态正常
	ThridAccountStatus_Exception int32 = 403   //三方账户状态异常
	CgReqThirdApiTimeOutCode     int   = 12345 //cg工程请求三方接口超时返回Code约定
)
const (
	Req_MD5_Salt = "me13ekihf9FGipFMnd56wqqRtfgd2DrqsG"
)

type WebAPI_ThirdPlatformGameMapping struct {
	GameFreeID            int32
	ThirdPlatformName     string
	ThirdGameID           string
	Desc                  string
	ScreenOrientationType int32
	ThirdID               int32
}

// 第三方平台基类
type ThirdPlatformBase struct {
	Id                     int32
	Name                   string
	Sort                   int32 //排序，在客户端要变现的顺序，暂时没有用到这个
	BaseURL                string
	Tag                    string //获取平台标记(没有中文，纯小写字母)
	SceneId                int    //场景id,与SceneMgr中的sceneID对应
	VultGameID             int32  //虚拟游戏ID
	BaseGameID             int    //场景ID
	IsNeedCheckQuota       bool   //是否需要检查平台配额
	CurrentPltQuota        int64  //当前配额剩余的金额
	ReqTimeOut             int32  //请求超时设置
	GamefreeIdMappingInfo  map[int32]*WebAPI_ThirdPlatformGameMapping
	ThirdgameIdMappingInfo map[string]*WebAPI_ThirdPlatformGameMapping
	TransferInteger        bool //转账现金必须为整数金额
}

// 定义第三方平台接口,后续新加的平台则实现如下接口就可以
type IThirdPlatform interface {
	GetPlatformBase() ThirdPlatformBase
	MappingGameName(snId int32) string                                                                                           //映射游戏名
	ReqCreateAccount(snId int32, platform, channel, ip string) error                                                             //创建账户
	ReqUserBalance(snId int32, platform, channel, ip string) (err error, balance int64)                                          //用户余额
	ReqIsAllowTransfer(snId int32, platform, channel string) bool                                                                //是否允许转账
	ReqTransfer(snId int32, amount int64, transferId string, platform, channel, ip string) (e error, timeout bool)               //转账
	ReqEnterGame(snId int32, gameId string, clientIP string, platform, channel string, amount int64) (err error, ret_url string) //进入游戏，返回进入游戏的URL
	ReqLeaveGame(snId int32, gameId string, clientIP string, platform, channel string) (err error, amount int64)                 //退出游戏，返回金币
	InitMappingRelation(db map[int32]*WebAPI_ThirdPlatformGameMapping)                                                           //初始化映射关系
	GamefreeId2ThirdGameInfo(gamefreeid int32) (thirdInfo *WebAPI_ThirdPlatformGameMapping)                                      //我方gamefreeid映射到三方
	ThirdGameInfo2GamefreeId(thirdInfo *WebAPI_ThirdPlatformGameMapping) (gamefreeid int32)                                      //三方信息映射到我方gamefreeid
}

// 将定义的三方平台注册到管理器中，在这里统一管理，避免在其它地方做多分支判断
func init() {
	ThridPlatformMgrSington.register(&XHJThridPlatform{
		ThirdPlatformBase{
			Id:               1,
			Name:             "XHJ平台",
			BaseURL:          "",
			Tag:              "xhj",
			IsNeedCheckQuota: false,
			BaseGameID:       common.GameId_Thr_XHJ,
			SceneId:          901,
			VultGameID:       9010001,
		},
	})
}
