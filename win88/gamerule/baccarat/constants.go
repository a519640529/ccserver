package baccarat

import "time"

////////////////////////////////////////////////////////////////////////////////
//场景状态
////////////////////////////////////////////////////////////////////////////////
const (
	BaccaratSceneStateStakeAnt    int = iota //准备押注状态
	BaccaratSceneStateStake                  //押注状态
	BaccaratSceneStateOpenCardAnt            //准备开牌状态
	BaccaratSceneStateOpenCard               //开牌状态
	BaccaratSceneStateBilled                 //结算状态
	BaccaratSceneStateMax
)

////////////////////////////////////////////////////////////////////////////////
//百家乐超时设置
////////////////////////////////////////////////////////////////////////////////
const (
	BaccaratStakeAntTimeout     = time.Second * 2  //准备押注
	BaccaratStakeTimeout        = time.Second * 11 //押注
	BaccaratOpenCardAntTimeout  = time.Second * 1  //准备
	BaccaratOpenCardTimeout     = time.Second * 9  //开牌
	BaccaratBilledTimeout       = time.Second * 5  //结算
	BaccaratRecordTime          = 5                //回收金币记录时间
	BaccaratBatchSendBetTimeout = time.Second * 1  //发送下注数据时间间隔
)

////////////////////////////////////////////////////////////////////////////////
//玩家操作，玩家也就两个操作
////////////////////////////////////////////////////////////////////////////////
const (
	BaccaratPlayerOpBet           int = iota //下注
	BaccaratPlayerOpGetOLList                //获取在线列表
	BaccaratPlayerOpUpBanker                 //上庄
	BaccaratPlayerOpNowDwonBanker            //在庄的下庄
	BaccaratPlayerOpUpList                   //上庄列表
	BaccaratPlayerOpDwonBanker               //下庄
)

const (
	GameId_Baccarat = 35
)

////////////////////////////////////////////////////////////////////////////////
//百家乐下注区域，注意是位标记
////////////////////////////////////////////////////////////////////////////////
const (
	BACCARAT_ZONE_TIE           int = 1 << iota //和1
	BACCARAT_ZONE_BANKER                        //庄家2
	BACCARAT_ZONE_PLAYER                        //闲家4
	BACCARAT_ZONE_BANKER_DOUBLE                 //庄家对8
	BACCARAT_ZONE_PLAYER_DOUBLE                 //闲家对16
	BACCARAT_ZONE_MAX
)
