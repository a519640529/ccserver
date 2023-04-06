package redvsblack

import "time"

const (
	RedVsBlackStakeAntTimeout     = time.Second * 2         //准备押注
	RedVsBlackStakeTimeout        = time.Second * 10        //押注
	RedVsBlackOpenCardAntTimeout  = time.Second * 1         //准备开牌
	RedVsBlackOpenCardTimeout     = time.Second * 6         //开牌
	RedVsBlackBilledTimeout       = time.Millisecond * 3500 //结算
	RedVsBlackRecordTime          = 5                       //回收金币记录时间
	RedVsBlackBatchSendBetTimeout = time.Second * 1         //发送下注数据时间间隔
)

//场景状态
const (
	RedVsBlackSceneStateStakeAnt    int = iota //准备押注
	RedVsBlackSceneStateStake                  //押注
	RedVsBlackSceneStateOpenCardAnt            //准备开牌
	RedVsBlackSceneStateOpenCard               //开牌
	RedVsBlackSceneStateBilled                 //结算
	RedVsBlackSceneStateMax
)

//玩家操作
const (
	RedVsBlackPlayerOpBet       int = iota //下注
	RedVsBlackPlayerOpGetOLList            //获取在线列表
)
