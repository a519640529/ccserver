package crash

import "time"

const (
	CrashStakeAntTimeout     = time.Second * 1         //准备押注
	CrashStakeTimeout        = time.Second * 6         //押注
	CrashOpenCardAntTimeout  = time.Second * 1         //准备开始
	CrashOpenCardTimeout     = time.Minute * 30        //开始
	CrashBilledTimeout       = time.Millisecond * 3500 //结算
	CrashBatchSendBetTimeout = time.Second * 1         //发送下注数据时间间隔
)

//场景状态
const (
	CrashSceneStateStakeAnt    int = iota //准备押注
	CrashSceneStateStake                  //押注
	CrashSceneStateOpenCardAnt            //准备开始
	CrashSceneStateOpenCard               //开始
	CrashSceneStateBilled                 //结算
	CrashSceneStateMax
)

//玩家操作
const (
	CrashPlayerOpBet       int = iota //下注
	CrashPlayerOpGetOLList            //获取在线列表
	CrashPlayerOpParachute            //跳伞
)

//倍率
const (
	MinMultiple = 100   //最小倍率
	MaxMultiple = 10000 //最大倍率
)
