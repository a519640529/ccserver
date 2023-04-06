package common

import (
	"github.com/idealeak/goserver/core/transact"
)

const (
	TransType_Login            transact.TransType = 1000
	TransType_Logout                              = 1001
	TransType_WebTrascate                         = 1002
	TransType_AddCoin                             = 1003
	TransType_ViewData                            = 1004
	TransType_DayTimeChange                       = 1005
	TransType_CoinSceneChange                     = 1006
	TransType_WebApi                              = 1007
	TransType_WebApi_ForRank                      = 1101
	TransType_GameSrvWebApi                       = 1008
	TransType_QueryCoinPool                       = 1009
	TransType_StopServer                          = 1010
	TransType_QueryAllCoinPool                    = 1011
	TransType_ActThrSrvWebApi                     = 1012
	TransType_MatchSceneChange                    = 1013
	TransType_MiniGameAddCoin                     = 1014
	TransType_ServerCtrl                          = 1015
)

type M2GWebTrascate struct {
	Tag   int
	Param string
}

type M2GWebApiRequest struct {
	Path     string
	RawQuery string
	ReqIp    string
	Body     []byte
}

type M2GWebApiResponse struct {
	Body []byte
}

type W2GQueryCoinPool struct {
	GameId   int32
	GameMode int32
	Platform string
	GroupId  int32
}
type PlatformStates struct {
	Platform string
	GamesVal map[int32]*CoinPoolStatesInfo
}
type GamesIndex struct {
	GameFreeId int32
	GroupId    int32
}
type QueryGames struct {
	Index map[string][]*GamesIndex
}

// 单个平台各游戏水池信息概况
type CoinPoolStatesInfo struct {
	GameId     int32 //当前游戏id
	GameFreeId int32 //游戏id
	LowerLimit int32 //库存下限
	UpperLimit int32 //库存上限
	CoinValue  int32 //当前库存值
	States     int32 //水池状态
}

type CoinPoolSetting struct {
	Platform         string //平台id
	GameFreeId       int32  //游戏id
	ServerId         int32  //服务器id
	GroupId          int32  //组id
	InitValue        int32  //初始库存值
	LowerLimit       int32  //库存下限
	UpperLimit       int32  //库存上限
	UpperOffsetLimit int32  //上限偏移值
	MaxOutValue      int32  //最大吐钱数
	ChangeRate       int32  //库存变化速度
	MinOutPlayerNum  int32  //最少吐钱人数
	UpperLimitOfOdds int32  //赔率上限(万分比)
	CoinValue        int64  //当前库存值
	PlayerNum        int32  //当前在线人数
	BaseRate         int32  //基础赔率
	CtroRate         int32  //调节赔率
	HardTimeMin      int32  //收分调节频率下限
	HardTimeMax      int32  //收分调节频率上限
	NormalTimeMin    int32  //正常调节频率下限
	NormalTimeMax    int32  //正常调节频率上限
	EasyTimeMin      int32  //放分调节频率下限
	EasyTimeMax      int32  //放分调节频率上限
	EasrierTimeMin   int32  //吐分调节频率下限
	EasrierTimeMax   int32  //吐分分调节频率上限
	CpCangeType      int32
	CpChangeInterval int32
	CpChangeTotle    int32
	CpChangeLower    int32
	CpChangeUpper    int32
	ProfitRate       int32 //收益比例
	ProfitPool       int64 //当前收益池
	CoinPoolMode     int32 //当前池模式
	ResetTime        int32
	ProfitAutoRate   int32
	ProfitManualRate int32
	ProfitUseManual  bool
}
