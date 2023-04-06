package rollcoin

import "time"

const (
	RollCoinWaiteTimeout    = time.Second * 0  //等待
	RollCoinStartTimeout    = time.Second * 1  //开始
	RollCoinRollTimeout     = time.Second * 18 //押注
	RollCoinCoinOverTimeout = time.Second * 1  //结束
	RollCoinGameTimeOut     = time.Second * 14 //比赛

	RollCoinBilledTimeout = time.Second * 5 //结算
)

//场景状态
const (
	RollCoinSceneStateWait     int = iota //等待状态
	RollCoinSceneStateStart               //开始
	RollCoinSceneStateRoll                //押注
	RollCoinSceneStateCoinOver            //押注结束
	RollCoinSceneStateGame                //比赛
	RollCoinSceneStateBilled              //结算
	RollCoinSceneStateMax
)

//玩家操作
const (
	RollCoinPlayerOpPushCoin      int = iota //押注
	RollCoinPlayerOpBanker                   //坐庄
	RollCoinPlayerOpBankerList               //坐庄列表
	RollCoinPlayerList                       //玩家列表
	RollCoinPlayerOpDownBanker               //下庄
	RollCoinPlayerOpCurDownBanker            //当前庄下局下庄
)
