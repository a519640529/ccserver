package dragonvstiger

import (
	"time"
)

const (
	CardsKind_Normal        int = iota //0.高牌
	CardsKind_Double                   //1.对子
	CardsKind_A23                      //2.顺子(地龙)
	CardsKind_ThreeSort                //3.顺子
	CardsKind_SameColor                //4.金花(同花)
	CardsKind_ColorA23                 //5.同花顺(地龙)
	CardsKind_SameColorSort            //6.同花顺
	CardsKind_ThreeSame                //7.豹子
	CardsKind_235Double
	CardsKind_Boom //9.AAA
	CardsKind_235Boom
	CardsKind_Max
)
const (
	DragonVsTigerStakeAntTimeout     = time.Second * 2         //准备押注
	DragonVsTigerStakeTimeout        = time.Second * 10        //押注
	DragonVsTigerOpenCardAntTimeout  = time.Second * 1         //准备开牌
	DragonVsTigerOpenCardTimeout     = time.Second * 5         //开牌
	DragonVsTigerBilledTimeout       = time.Millisecond * 3500 //结算
	DragonVsTigerRecordTime          = 5                       //回收金币记录时间
	DragonVsTigerBatchSendBetTimeout = time.Second * 1         //发送下注数据时间间隔
)

//场景状态
const (
	DragonVsTigerSceneStateStakeAnt    int = iota //准备押注
	DragonVsTigerSceneStateStake                  //押注
	DragonVsTigerSceneStateOpenCardAnt            //准备开牌
	DragonVsTigerSceneStateOpenCard               //开牌
	DragonVsTigerSceneStateBilled                 //结算
	DragonVsTigerSceneStateMax
)

//玩家操作
const (
	DragonVsTigerPlayerOpBet           int = iota //下注
	DragonVsTigerPlayerOpGetOLList                //获取在线列表
	DragonVsTigerPlayerOpUpBanker                 //上庄
	DragonVsTigerPlayerOpDwonBanker               //下庄
	DragonVsTigerPlayerOpUpList                   //上庄列表
	DragonVsTigerPlayerOpNowDwonBanker            //在庄的下庄
)
const MaxBankerNum = 10
const (
	DVST_ZONE_DRAW int = iota
	DVST_ZONE_DRAGON
	DVST_ZONE_TIGER
	DVST_ZONE_MAX
)
const (
	ROBOT_TYPE_DVTRANDOM int = iota
	ROBOT_TYPE_DVTFWIN
	ROBOT_TYPE_DVTIWIN
)
