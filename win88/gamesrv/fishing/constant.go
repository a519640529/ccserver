package fishing

import "time"

const BULLETLIMIT = 2048
const WINDOW_SIZE = 10

// 玩家的炮台类型
const (
	NormalPowerType = iota // 普通炮台
	FreePowerType          // 免费炮台
	BitPowerType           // 钻头贝炮台
)

//玩家数据索引
const (
	GDATAS_HPFISHING_PRANA     int = iota //聚能炮能量
	GDATAS_HPFISHING_ALLBET               //总下注
	GDATAS_HPFISHING_ALLBET64             //总下注 高字节
	GDATAS_HPFISHING_CHANGEBET            //每一局 金币变动总额
	GDATAS_FISHING_SELVIP                 //VIP 炮等级
	GDATAS_HPFISHING_MAX
)
const (
	FishDrop_Rate = 10000
	CountSaveNums = 300
)

const (
	Policy_Mode_Normal PolicyMode = 1
	Policy_Mode_Tide   PolicyMode = 2
)
const (
	FishingSceneAniTimeout = time.Second * 3
)
const (
	FishingSceneStateStart int = iota
	FishingSceneStateClear
	FishingSceneStateMax
)
const (
	FishingPlayerOpFire int = iota
	FishingPlayerOpHitFish
	FishingPlayerOpSetPower
	FishingPlayerOpSelVip
	FishingPlayerOpRobotFire
	FishingPlayerOpRobotHitFish
	FishingPlayerOpLeave
	FishingPlayerOpEnter
	FishingPlayerOpAuto
	FishingPlayerOpSelTarget
	FishingPlayerOpFireRate
	FishingRobotOpAuto
	FishingRobotOpSetPower
	FishingPlayerHangup
	FishingRobotWantLeave
)

const (
	FishingRobotBehaviorCode_StopFire int32 = iota //停止开炮
)
