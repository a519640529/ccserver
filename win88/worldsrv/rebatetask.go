package main

import (
	"encoding/json"
	"games.yol.com/win88/common"
	"games.yol.com/win88/model"
	"games.yol.com/win88/srvdata"
	"games.yol.com/win88/webapi"
	"github.com/idealeak/goserver/core/logger"
	"strconv"
)

//返利任务

type RebateTask struct {
	Platform           string                         //平台名称
	RebateSwitch       bool                           //返利开关
	RebateManState     int                            //返利是开启个人返利  0 关闭  1 开启
	ReceiveMode        int                            //领取方式  0实时领取  1次日领取
	NotGiveOverdue     int                            //0不过期   1过期不给  2过期邮件给
	RebateGameCfg      map[string]*RebateGameCfg      //key为GameDif
	RebateGameThirdCfg map[string]*RebateGameThirdCfg //第三方key
	Version            int                            //活动版本 后台控制
}
type RebateGameCfg struct {
	BaseCoin      [3]int32 //返利基准
	RebateRate    [3]int32 //返利比率 万分比
	GameId        int32    //游戏id
	GameMode      int32    //游戏类型
	MaxRebateCoin int64    //最高返利
}
type RebateGameThirdCfg struct {
	BaseCoin      [3]int32 //返利基准
	RebateRate    [3]int32 //返利比率
	ThirdName     string   //第三方key
	ThirdShowName string   //前端显示name
	MaxRebateCoin int64    //最高返利
	ThirdId       string   //三方游戏id
}
type RebateInfo struct {
	rebateTask map[string]*RebateTask
}

var RebateInfoMgrSington = &RebateInfo{
	rebateTask: make(map[string]*RebateTask),
}

func (this *RebateInfo) Init() {
}
func (this *RebateInfo) Update() {
}
func (this *RebateInfo) Shutdown() {

}
func (this *RebateInfo) ModuleName() string {
	return "RebateTask"
}
func (this *RebateInfo) UpdateRebateDataByApi(platform string, RebateInfo RebateInfos, isThirdName bool) {
	rebateTask := this.rebateTask[platform]
	if rebateTask == nil {
		this.rebateTask[platform] = &RebateTask{}
		rebateTask = this.rebateTask[platform]
	}
	rebateTask.Platform = RebateInfo.Platform
	rebateTask.RebateSwitch = RebateInfo.RebateSwitch
	rebateTask.Version = RebateInfo.Version
	rebateTask.NotGiveOverdue = RebateInfo.NotGiveOverdue
	rebateTask.ReceiveMode = RebateInfo.ReceiveMode
	rebateTask.RebateManState = RebateInfo.RebateManState
	if !isThirdName {
		rebateTask.RebateGameCfg = make(map[string]*RebateGameCfg)
		for _, v := range RebateInfo.RebateGameCfg {
			for _, dfm := range srvdata.PBDB_GameFreeMgr.Datas.Arr {
				if dfm.GetGameId() == v.GameId && dfm.GetGameMode() == v.GameMode {
					rebateTask.RebateGameCfg[dfm.GetGameDif()] = v
					break
				}
			}
		}
	} else {
		rebateTask.RebateGameThirdCfg = make(map[string]*RebateGameThirdCfg)
		for _, v := range RebateInfo.RebateGameThirdCfg {
			thirdId := strconv.Itoa(int(ThirdPltGameMappingConfig.FindThirdIdByThird(v.ThirdName)))
			rebateTask.RebateGameThirdCfg[thirdId] = v
		}
	}
}

// 返利数据初始化
func (this *RebateInfo) LoadRebateData() {
	//构建默认的平台数据
	//this.CreateDefaultPlatform()
	//获取平台返利数据  getGameRebateConfig
	type ApiResult struct {
		Tag int
		Msg []RebateInfos
	}
	if !model.GameParamData.UseEtcd {
		rebateBuff, err := webapi.API_GetGameRebateConfig(common.GetAppId())
		if err == nil {
			logger.Logger.Trace("API_GetGameRebateConfig:", string(rebateBuff))
			ar := ApiResult{}
			err = json.Unmarshal(rebateBuff, &ar)
			if err == nil && ar.Tag == 0 {
				for _, v := range ar.Msg {
					rebateTask := this.rebateTask[v.Platform]
					if rebateTask == nil {
						this.rebateTask[v.Platform] = &RebateTask{
							Platform:       v.Platform,
							RebateSwitch:   v.RebateSwitch,
							ReceiveMode:    v.ReceiveMode,
							RebateManState: v.RebateManState,
							NotGiveOverdue: v.NotGiveOverdue,
							Version:        v.Version,
						}
						rebateTask = this.rebateTask[v.Platform]
					}
					rebateTask.RebateGameCfg = make(map[string]*RebateGameCfg)
					rebateTask.RebateGameThirdCfg = make(map[string]*RebateGameThirdCfg)
					for _, v := range v.RebateGameCfg {
						for _, dfm := range srvdata.PBDB_GameFreeMgr.Datas.Arr {
							if dfm.GetGameId() == v.GameId && dfm.GetGameMode() == v.GameMode {
								rebateTask.RebateGameCfg[dfm.GetGameDif()] = v
								break
							}
						}
					}
					for _, v := range v.RebateGameThirdCfg {
						thirdId := strconv.Itoa(int(ThirdPltGameMappingConfig.FindThirdIdByThird(v.ThirdName)))
						v.ThirdId = thirdId
						rebateTask.RebateGameThirdCfg[thirdId] = v
					}
				}
			} else {
				logger.Logger.Error("Unmarshal RebateTask data error:", err)
			}
		} else {
			logger.Logger.Error("Get RebateTask data error:", err)
		}
	} else {
		EtcdMgrSington.InitRebateConfig()
	}
}

// ///////////////////////////////////////////////////////////////////////
// 对应客户端
const (
	//棋牌类
	KaiYuanChessGame = 6  //开元棋牌
	SelfChessGame    = 7  //博乐棋牌
	FGChessGame      = 11 //FG棋牌
	WWGChessGame     = 12 //WWG棋牌
	//电子类
	FGElectronicGame  = 4 //FG电子
	WWGElectronicGame = 5 //WWG电子
)

// //////////////////////////////////
// /个人信息
type TotalInfo struct {
	GameTotalNum       int32  //游戏总局数
	GameMostPartake    string //参与最多游戏名字
	GameMostProfit     string //单局最多盈利游戏名字
	CreateRoomNum      int32  //创建房间总数
	CreateRoomMost     string //创建房间最多游戏名字
	CreateClubNum      int32  //创建俱乐部数量
	CreateClubRoomMost string //创建包间最多游戏名字
	TeamNum            int32  //团队人数
	AchievementTotal   int32  //推广总业绩
	RewardTotal        int32  //推广总奖励
}

// /手动洗码
type CodeCoinRecord struct {
	GameName    string //游戏名称
	GameBetCoin int64  //游戏洗码量
	Rate        int32  //比例
	Coin        int32  //洗码金额
}
type CodeCoinTotal struct {
	TotalCoin      int64 //总计游戏投注
	CodeCoinRecord []*CodeCoinRecord
	PageNo         int //当前页
	PageSize       int //每页数量
	PageNum        int //总页数
}

// /投注记录
type BetCoinRecord struct {
	Ts           int64  //时间戳
	GameName     string //游戏名称
	RecordId     string //注单号
	BetCoin      int64  //投注金额
	ReceivedCoin int64  //已派奖
}
type BetCoinTotal struct {
	BetCoinRecord []*BetCoinRecord
	PageNo        int //当前页
	PageSize      int //每页数量
	PageNum       int //总页数
}

// /账户明细
type CoinDetailed struct {
	Ts       int64 //时间戳
	CoinType int64 //交易类型
	Income   int64 //收入
	Disburse int64 //支出
	Coin     int64 //金额
}
type CoinDetailedTotal struct {
	RechargeCoin int64 //充值
	ExchangeCoin int64 //兑换
	ClubAddCoin  int64 //俱乐部加币
	RebateCoin   int64 //返利
	Activity     int64 //活动获取
	CoinDetailed []*CoinDetailed
	PageNo       int //当前页
	PageSize     int //每页数量
	PageNum      int //总页数
}

// /个人报表
type ReportForm struct {
	ShowType   int   //标签 棋牌游戏等等
	ProfitCoin int64 //盈利总额
	BetCoin    int64 //有效投注总额
	FlowCoin   int64 //流水总额
}

// //////////////////////////////////
type HallGameType struct {
	GameId   int32
	GameMode int32
}
type GameConfig struct {
	Platform    string
	IsKnowType  bool
	BigGameType map[int32]map[int32][]*HallGameType
}

//func (this *PersonInfo) FormatConfig(platform string, btl []*protocol.BigTagList) {
//	if this.PlatformType[platform] == nil || (this.PlatformType[platform] != nil && !this.PlatformType[platform].IsKnowType) {
//		gcf := &GameConfig{
//			Platform:   platform,
//			IsKnowType: true,
//		}
//		for _, bigTag := range btl {
//			gcf.BigGameType[bigTag.GetBigTagId()] = make(map[int32][]*HallGameType)
//			for _, smallTag := range bigTag.SmallTagList {
//				gcf.BigGameType[bigTag.GetBigTagId()][smallTag.GetSmallTagId()] = []*HallGameType{}
//				bs := gcf.BigGameType[bigTag.GetBigTagId()][smallTag.GetSmallTagId()]
//				for _, v := range smallTag.GameType {
//					bs = append(bs, &HallGameType{GameId: v.GetGameId(), GameMode: v.GetGameMode()})
//				}
//			}
//		}
//		this.PlatformType[platform] = gcf
//	}
//}

type PersonInfo struct {
	PlatformType map[string]*GameConfig
}

var PersonInfoMgrSingTon = &PersonInfo{
	PlatformType: make(map[string]*GameConfig),
}

func init() {
	RegisteParallelLoadFunc("平台返利数据", func() error {
		RebateInfoMgrSington.LoadRebateData()
		return nil
	})

}
